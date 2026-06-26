// Package outlook implements a read-only Microsoft Outlook connector using
// Microsoft Graph mail endpoints. It exchanges a provided refresh token for a
// bearer token, then follows Graph value[] and @odata.nextLink pagination.
package outlook

import (
	"bytes"
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
	defaultBaseURL  = "https://graph.microsoft.com/v1.0"
	defaultPageSize = 100
	maxPageSize     = 999
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("outlook", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "outlook" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "outlook", DisplayName: "Outlook", IntegrationType: "api", Description: "Reads Outlook messages, mail folders, and calendar events through Microsoft Graph.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	return r.DoJSON(ctx, http.MethodGet, "/me/messages", url.Values{"$top": []string{"1"}}, nil, nil)
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: "outlook", Streams: streams()}, nil
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
		stream = "messages"
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("outlook stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, ep, req, emit)
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
	return harvest(ctx, r, ep, size, max, emit)
}

func harvest(ctx context.Context, r *connsdk.Requester, ep streamEndpoint, size, max int, emit func(connectors.Record) error) error {
	path := ep.path
	query := url.Values{"$top": []string{strconv.Itoa(size)}}
	for page := 0; max == 0 || page < max; page++ {
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read outlook %s: %w", ep.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "value")
		if err != nil {
			return err
		}
		for _, item := range records {
			if err := emit(ep.mapRecord(map[string]any(item))); err != nil {
				return err
			}
		}
		next, err := nextLink(resp.Body)
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

func nextLink(body []byte) (string, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var env struct {
		Next string `json:"@odata.nextLink"`
	}
	if err := dec.Decode(&env); err != nil {
		return "", fmt.Errorf("decode graph envelope: %w", err)
	}
	return env.Next, nil
}

func readFixture(ctx context.Context, stream string, ep streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "subject": fmt.Sprintf("Fixture %d", i), "displayName": fmt.Sprintf("Fixture %d", i), "name": fmt.Sprintf("Fixture %d", i), "receivedDateTime": "2026-01-01T00:00:00Z", "lastModifiedDateTime": "2026-01-01T00:00:00Z", "createdDateTime": "2026-01-01T00:00:00Z", "webLink": "https://example.com"}
		rec := ep.mapRecord(item)
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
		return nil, errors.New("outlook connector requires client_id, client_secret, and refresh_token")
	}
	return &refreshTokenAuth{Client: c.Client, TokenURL: tokenURL(cfg), ClientID: clientID, ClientSecret: clientSecret, RefreshToken: refresh, Scope: strings.TrimSpace(cfg.Config["scope"])}, nil
}

type refreshTokenAuth struct {
	Client       *http.Client
	TokenURL     string
	ClientID     string
	ClientSecret string
	RefreshToken string
	Scope        string

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
	if strings.TrimSpace(a.Scope) != "" {
		form.Set("scope", a.Scope)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("outlook oauth: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	client := a.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("outlook oauth: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("outlook oauth: token endpoint returned %s", resp.Status)
	}
	var out struct {
		AccessToken string      `json:"access_token"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("outlook oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("outlook oauth: token response missing access_token")
	}
	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = time.Now().Add(ttl)
	return a.token, nil
}

func tokenURL(cfg connectors.RuntimeConfig) string {
	if v := strings.TrimSpace(cfg.Config["token_url"]); v != "" {
		return v
	}
	tenant := strings.TrimSpace(cfg.Config["tenant_id"])
	if tenant == "" {
		tenant = "common"
	}
	return "https://login.microsoftonline.com/" + url.PathEscape(tenant) + "/oauth2/v2.0/token"
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	return validateBaseURL("outlook", base)
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

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > maxPageSize {
		return 0, fmt.Errorf("outlook config page_size must be between 1 and %d", maxPageSize)
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
		return 0, errors.New("outlook config max_pages must be a non-negative integer")
	}
	return n, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

type streamEndpoint struct {
	path      string
	mapRecord func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{"messages": {path: "/me/messages", mapRecord: messageRecord}, "mail_folders": {path: "/me/mailFolders", mapRecord: folderRecord}, "events": {path: "/me/events", mapRecord: eventRecord}}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "messages", Description: "Outlook mail messages.", PrimaryKey: []string{"id"}, CursorFields: []string{"received_date_time"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "subject", Type: "string"}, {Name: "received_date_time", Type: "timestamp"}, {Name: "last_modified_date_time", Type: "timestamp"}, {Name: "web_link", Type: "string"}}},
		{Name: "mail_folders", Description: "Outlook mail folders.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "display_name", Type: "string"}, {Name: "total_item_count", Type: "integer"}, {Name: "unread_item_count", Type: "integer"}}},
		{Name: "events", Description: "Outlook calendar events.", PrimaryKey: []string{"id"}, CursorFields: []string{"last_modified_date_time"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "subject", Type: "string"}, {Name: "created_date_time", Type: "timestamp"}, {Name: "last_modified_date_time", Type: "timestamp"}, {Name: "web_link", Type: "string"}}},
	}
}

func messageRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "subject": item["subject"], "received_date_time": item["receivedDateTime"], "last_modified_date_time": item["lastModifiedDateTime"], "web_link": item["webLink"]}
}

func folderRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "display_name": item["displayName"], "total_item_count": item["totalItemCount"], "unread_item_count": item["unreadItemCount"]}
}

func eventRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "subject": item["subject"], "created_date_time": item["createdDateTime"], "last_modified_date_time": item["lastModifiedDateTime"], "web_link": item["webLink"]}
}
