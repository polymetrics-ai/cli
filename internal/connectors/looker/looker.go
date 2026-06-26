// Package looker implements a read-only Looker API connector. It supports either
// a direct access_token secret or Looker's client_id/client_secret login flow.
package looker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultPageSize = 100
	maxPageSize     = 500
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("looker", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "looker" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "looker", DisplayName: "Looker", IntegrationType: "api", Description: "Reads Looker users, groups, folders, looks, and dashboards through the Looker API 4.0.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if err := requireCredentials(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "users", url.Values{"limit": []string{"1"}, "offset": []string{"0"}}, nil, nil); err != nil {
		return fmt.Errorf("check looker: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "users"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("looker stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		query := url.Values{"limit": []string{strconv.Itoa(pageSize)}, "offset": []string{strconv.Itoa(offset)}}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read looker %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode looker %s: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		offset += len(records)
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%d", i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "display_name": fmt.Sprintf("Fixture User %d", i), "email": fmt.Sprintf("fixture+%d@example.com", i), "title": fmt.Sprintf("Fixture %d", i)}
		rec := endpoint.mapRecord(item)
		rec["connector"] = "looker"
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireCredentials(cfg); err != nil {
		return nil, err
	}
	var auth connsdk.Authenticator
	if token := strings.TrimSpace(accessToken(cfg)); token != "" {
		auth = connsdk.Bearer(token)
	} else {
		auth = &loginAuth{tokenURL: tokenURL(cfg, base), clientID: cfg.Secrets["client_id"], clientSecret: cfg.Secrets["client_secret"], client: c.Client}
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: auth, UserAgent: userAgent}, nil
}

type loginAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	client       *http.Client
	mu           sync.Mutex
	token        string
	expires      time.Time
}

func (a *loginAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *loginAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" && time.Now().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	form := url.Values{"client_id": []string{a.clientID}, "client_secret": []string{a.clientSecret}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("looker login: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("looker login: request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("looker login: token endpoint returned %s", resp.Status)
	}
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("looker login: decode response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("looker login: token response missing access_token")
	}
	a.token = out.AccessToken
	ttl := time.Duration(out.ExpiresIn) * time.Second
	if ttl <= 0 {
		ttl = time.Hour
	}
	a.expires = time.Now().Add(ttl)
	return a.token, nil
}

func requireCredentials(cfg connectors.RuntimeConfig) error {
	if strings.TrimSpace(accessToken(cfg)) != "" {
		return nil
	}
	if cfg.Secrets == nil || strings.TrimSpace(cfg.Secrets["client_id"]) == "" || strings.TrimSpace(cfg.Secrets["client_secret"]) == "" {
		return errors.New("looker connector requires secret access_token or client_id and client_secret")
	}
	return nil
}
func accessToken(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["access_token"]
}
func tokenURL(cfg connectors.RuntimeConfig, base string) string {
	if v := strings.TrimSpace(cfg.Config["token_url"]); v != "" {
		return v
	}
	return strings.TrimRight(base, "/") + "/login"
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return "", errors.New("looker connector requires config base_url (for example https://company.looker.com/api/4.0)")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("looker config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("looker config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("looker config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}
func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "looker config page_size")
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("looker config max_pages must be a non-negative integer: %w", err)
	}
	return value, nil
}
func boundedInt(raw string, def, max int, name string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	if value < 1 || value > max {
		return 0, fmt.Errorf("%s must be between 1 and %d", name, max)
	}
	return value, nil
}
func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
