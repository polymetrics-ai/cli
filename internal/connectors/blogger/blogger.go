// Package blogger implements the native pm Blogger (Google Blogger API v3)
// connector. It follows the declarative-HTTP shape of the stripe reference
// connector: a thin package composing the connsdk toolkit (Requester + retry +
// RecordsAt/StringAt extraction) with Blogger-specific stream definitions,
// endpoints, and a refresh-token OAuth2 authenticator.
//
// Blogger uses Google's OAuth2 web-server flow: the connector exchanges the
// supplied long-lived refresh token (plus client_id/client_secret) for a
// short-lived bearer access token at the Google token endpoint, then reads the
// blog-scoped REST resources (blogs, posts, pages, comments) with
// pageToken/nextPageToken pagination.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package blogger

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
	bloggerDefaultBaseURL  = "https://www.googleapis.com/blogger/v3"
	bloggerDefaultTokenURL = "https://oauth2.googleapis.com/token"
	bloggerDefaultPageSize = 100
	bloggerMaxPageSize     = 500
	bloggerUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("blogger", New)
}

// New returns the Blogger connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Blogger connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "blogger" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "blogger",
		DisplayName:     "Blogger",
		IntegrationType: "api",
		Description:     "Reads Blogger (Google Blogger API v3) blogs, posts, pages, and comments using an OAuth2 refresh token. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Blogger. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := bloggerBaseURL(cfg); err != nil {
		return err
	}
	blogID, err := bloggerBlogID(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the blog resource confirms token exchange, auth, and
	// connectivity without mutating anything.
	endpoint := bloggerStreamEndpoints["blogs"]
	if err := r.DoJSON(ctx, http.MethodGet, endpoint.path(blogID), nil, nil, nil); err != nil {
		return fmt.Errorf("check blogger: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: bloggerStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Blogger stream starts with
// an empty incremental cursor (full sync). Blogger list endpoints do not accept
// an arbitrary updated-since filter, so the cursor is informational here.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "posts"
	}
	endpoint, ok := bloggerStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("blogger stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	blogID, err := bloggerBlogID(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := bloggerPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := bloggerMaxPages(req.Config)
	if err != nil {
		return err
	}

	path := endpoint.path(blogID)
	if endpoint.single {
		return c.readSingle(ctx, r, path, endpoint, emit)
	}
	return c.harvest(ctx, r, path, endpoint, pageSize, maxPages, emit)
}

// readSingle reads an endpoint that returns one resource object (the blog).
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return fmt.Errorf("read blogger %s: %w", path, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode blogger %s: %w", path, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// harvest drives Blogger's pageToken/nextPageToken pagination. List responses
// are {items:[...], nextPageToken?:string}; the next page is requested with
// pageToken=<nextPageToken>. connsdk has a CursorPaginator for body-token
// pagination, but the loop is kept in-package for clarity and to bound maxPages.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("maxResults", strconv.Itoa(pageSize))

	pageToken := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read blogger %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "items")
		if err != nil {
			return fmt.Errorf("decode blogger %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "nextPageToken")
		if err != nil {
			return fmt.Errorf("decode blogger %s nextPageToken: %w", path, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		pageToken = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise blogger credential-free (mirrors stripe).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	count := 2
	if endpoint.single {
		count = 1
	}
	for i := 1; i <= count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		published := fmt.Sprintf("2026-01-0%dT00:00:00Z", i)
		updated := fmt.Sprintf("2026-01-1%dT00:00:00Z", i)
		item := map[string]any{
			"id":          fmt.Sprintf("%s_fixture_%d", stream, i),
			"kind":        "blogger#" + strings.TrimSuffix(stream, "s"),
			"name":        fmt.Sprintf("Fixture Blog %d", i),
			"title":       fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"content":     fmt.Sprintf("Fixture content %d", i),
			"description": fmt.Sprintf("Fixture description %d", i),
			"url":         fmt.Sprintf("https://example.blogspot.com/%s/%d", stream, i),
			"published":   published,
			"updated":     updated,
			"status":      "LIVE",
			"author":      map[string]any{"id": fmt.Sprintf("author_%d", i), "displayName": fmt.Sprintf("Fixture Author %d", i)},
			"blog":        map[string]any{"id": "fixture_blog_1"},
			"post":        map[string]any{"id": "fixture_post_1"},
			"replies":     map[string]any{"totalItems": "0"},
			"posts":       map[string]any{"totalItems": "2"},
			"pages":       map[string]any{"totalItems": "2"},
		}
		record := endpoint.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the refresh-token OAuth2
// authenticator and the resolved base URL. Secrets only ever flow into the
// authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := bloggerBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	auth, err := c.authenticator(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: bloggerUserAgent,
	}, nil
}

// authenticator builds the refresh-token OAuth2 authenticator from the resolved
// secrets and token URL.
func (c Connector) authenticator(cfg connectors.RuntimeConfig) (connsdk.Authenticator, error) {
	clientID := bloggerSecret(cfg, "client_id")
	clientSecret := bloggerSecret(cfg, "client_secret")
	refreshToken := bloggerSecret(cfg, "client_refresh_token")
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" || strings.TrimSpace(refreshToken) == "" {
		return nil, errors.New("blogger connector requires secrets client_id, client_secret, and client_refresh_token")
	}
	return &refreshTokenAuth{
		TokenURL:     bloggerTokenURL(cfg),
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
		Client:       c.Client,
	}, nil
}

// refreshTokenAuth fetches and caches a bearer token using the OAuth2
// refresh-token grant, refreshing automatically before expiry. connsdk's
// OAuth2ClientCredentials covers only the client-credentials grant, so this
// small in-package authenticator handles Google's refresh_token grant.
type refreshTokenAuth struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	RefreshToken string
	Client       *http.Client
	// Now is injectable for tests. Defaults to time.Now.
	Now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *refreshTokenAuth) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
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
	// Reuse a cached token until 60s before expiry to avoid edge races.
	if a.token != "" && a.now().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", a.ClientID)
	form.Set("client_secret", a.ClientSecret)
	form.Set("refresh_token", a.RefreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("blogger oauth2: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := a.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("blogger oauth2: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("blogger oauth2: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("blogger oauth2: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("blogger oauth2: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.now().Add(ttl)
	return a.token, nil
}

func bloggerSecret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

// bloggerBlogID resolves the blog_id config field, required for live reads.
func bloggerBlogID(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return "", errors.New("blogger connector requires config blog_id")
	}
	blogID := strings.TrimSpace(cfg.Config["blog_id"])
	if blogID == "" {
		return "", errors.New("blogger connector requires config blog_id")
	}
	return url.PathEscape(blogID), nil
}

// bloggerBaseURL resolves and validates the base URL. The default is
// googleapis.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func bloggerBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return bloggerDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("blogger config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("blogger config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("blogger config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// bloggerTokenURL resolves the OAuth2 token endpoint, defaulting to Google's.
// A test or self-hosted override is honored but validated lightly by the
// authenticator on use.
func bloggerTokenURL(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return bloggerDefaultTokenURL
	}
	if v := strings.TrimSpace(cfg.Config["token_url"]); v != "" {
		return v
	}
	return bloggerDefaultTokenURL
}

func bloggerPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return bloggerDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("blogger config page_size must be an integer: %w", err)
	}
	if value < 1 || value > bloggerMaxPageSize {
		return 0, fmt.Errorf("blogger config page_size must be between 1 and %d", bloggerMaxPageSize)
	}
	return value, nil
}

func bloggerMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("blogger config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("blogger config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}

// Write is unsupported: the Blogger connector is read-only. The method exists to
// satisfy the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
