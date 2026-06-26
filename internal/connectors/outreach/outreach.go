// Package outreach implements a read-only connector for the Outreach API v2. It
// uses the documented OAuth refresh-token flow and follows JSON:API links.next
// pagination for a small set of high-value collections.
package outreach

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
	defaultBaseURL  = "https://api.outreach.io/api/v2"
	defaultTokenURL = "https://api.outreach.io/oauth/token"
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("outreach", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "outreach" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "outreach", DisplayName: "Outreach", IntegrationType: "api", Description: "Reads Outreach prospects, accounts, sequences, and mailings.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	return r.DoJSON(ctx, http.MethodGet, "/prospects", url.Values{"page[size]": []string{"1"}}, nil, nil)
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: "outreach", Streams: streams()}, nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "prospects"
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("outreach stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, req, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	size, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	max, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	return harvest(ctx, r, ep, req, size, max, emit)
}

func harvest(ctx context.Context, r *connsdk.Requester, ep streamEndpoint, req connectors.ReadRequest, size, max int, emit func(connectors.Record) error) error {
	path := ep.path
	query := url.Values{"page[size]": []string{strconv.Itoa(size)}}
	if lower := lowerBound(req); lower != "" {
		query.Set("filter[updatedAt]", lower)
	}
	for page := 0; max == 0 || page < max; page++ {
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read outreach %s: %w", ep.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return err
		}
		for _, item := range records {
			if err := emit(jsonAPIRecord(map[string]any(item))); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "links.next")
		if err != nil {
			return err
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		path = next
		query = nil
	}
	return nil
}

func readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "type": strings.TrimSuffix(stream, "s"), "attributes": map[string]any{"email": fmt.Sprintf("fixture+%d@example.com", i), "name": fmt.Sprintf("Fixture %d", i), "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z"}}
		rec := jsonAPIRecord(item)
		if cursor := req.State[connsdk.CursorStateKey]; cursor != "" {
			rec["previous_cursor"] = cursor
		}
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
	auth, err := c.authenticator(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: auth, UserAgent: userAgent}, nil
}

func (c Connector) authenticator(cfg connectors.RuntimeConfig) (connsdk.Authenticator, error) {
	clientID := configOrSecret(cfg, "client_id")
	clientSecret := configOrSecret(cfg, "client_secret")
	refresh := configOrSecret(cfg, "refresh_token")
	if clientID == "" || clientSecret == "" || refresh == "" {
		return nil, errors.New("outreach connector requires client_id, client_secret, and refresh_token")
	}
	return &refreshTokenAuth{Client: c.Client, TokenURL: tokenURL(cfg), ClientID: clientID, ClientSecret: clientSecret, RefreshToken: refresh, RedirectURI: strings.TrimSpace(cfg.Config["redirect_uri"])}, nil
}

type refreshTokenAuth struct {
	Client       *http.Client
	TokenURL     string
	ClientID     string
	ClientSecret string
	RefreshToken string
	RedirectURI  string

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *refreshTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *refreshTokenAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" && time.Now().Add(time.Minute).Before(a.expires) {
		return a.token, nil
	}
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", a.ClientID)
	form.Set("client_secret", a.ClientSecret)
	form.Set("refresh_token", a.RefreshToken)
	if a.RedirectURI != "" {
		form.Set("redirect_uri", a.RedirectURI)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("outreach oauth: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	client := a.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("outreach oauth: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("outreach oauth: token endpoint returned %s", resp.Status)
	}
	var out struct {
		AccessToken string      `json:"access_token"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("outreach oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("outreach oauth: token response missing access_token")
	}
	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = time.Now().Add(ttl)
	return a.token, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	return validateBaseURL("outreach", base)
}

func tokenURL(cfg connectors.RuntimeConfig) string {
	if v := strings.TrimSpace(cfg.Config["token_url"]); v != "" {
		return v
	}
	return defaultTokenURL
}

func validateBaseURL(name, raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", name, err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", name, u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", name)
	}
	return strings.TrimRight(raw, "/"), nil
}

func configOrSecret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Config != nil {
		if v := strings.TrimSpace(cfg.Config[key]); v != "" {
			return v
		}
	}
	if cfg.Secrets != nil {
		return strings.TrimSpace(cfg.Secrets[key])
	}
	return ""
}

func lowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > maxPageSize {
		return 0, fmt.Errorf("outreach config page_size must be between 1 and %d", maxPageSize)
	}
	return n, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.ToLower(strings.TrimSpace(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, errors.New("outreach config max_pages must be a non-negative integer")
	}
	return n, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

type streamEndpoint struct{ path string }

var streamEndpoints = map[string]streamEndpoint{"prospects": {path: "/prospects"}, "accounts": {path: "/accounts"}, "sequences": {path: "/sequences"}, "mailings": {path: "/mailings"}}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "type", Type: "string"}, {Name: "email", Type: "string"}, {Name: "name", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}}
	return []connectors.Stream{{Name: "prospects", Description: "Outreach prospects.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields}, {Name: "accounts", Description: "Outreach accounts.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields}, {Name: "sequences", Description: "Outreach sequences.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields}, {Name: "mailings", Description: "Outreach mailings.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields}}
}

func jsonAPIRecord(item map[string]any) connectors.Record {
	attrs := map[string]any{}
	if raw, ok := item["attributes"].(map[string]any); ok {
		attrs = raw
	}
	return connectors.Record{"id": item["id"], "type": item["type"], "email": attrs["email"], "name": first(attrs, "name", "displayName"), "created_at": attrs["createdAt"], "updated_at": attrs["updatedAt"]}
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}
