// Package googleforms implements the native pm Google Forms connector. It is a
// declarative-HTTP per-system connector that copies the stripe template: a thin
// package composing the connsdk toolkit (Requester + extraction helpers + cursor
// state) with Google Forms-specific stream definitions, endpoints, and an
// OAuth2 refresh-token authenticator.
//
// The Google Forms API (https://forms.googleapis.com/v1) authenticates with a
// short-lived OAuth2 access token. This connector is configured with a long-lived
// refresh token (plus client id/secret) and exchanges it for an access token at
// Google's token endpoint, then attaches a Bearer header to each API request.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package googleforms

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
	googleFormsDefaultBaseURL  = "https://forms.googleapis.com/v1"
	googleFormsDefaultTokenURL = "https://oauth2.googleapis.com/token"
	googleFormsUserAgent       = "polymetrics-go-cli"
	googleFormsDefaultPageSize = 5000
	googleFormsMaxPageSize     = 5000
	// googleFormsFixtureTime is the deterministic timestamp used by fixture-mode
	// records.
	googleFormsFixtureTime = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("google-forms", New)
}

// New returns the Google Forms connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Google Forms connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth2 token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "google-forms" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "google-forms",
		DisplayName:     "Google Forms",
		IntegrationType: "api",
		Description:     "Reads Google Forms metadata, form items, and submitted responses through the Google Forms REST API using an OAuth2 refresh token.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to the Forms
// API. In fixture mode it short-circuits without a network call. Otherwise it
// validates config, that the OAuth2 secrets are present, and performs a bounded
// metadata read of the first configured form.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := googleFormsBaseURL(cfg); err != nil {
		return err
	}
	if err := requireSecrets(cfg); err != nil {
		return err
	}
	formIDs, err := formIDs(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded metadata read of the first form confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "forms/"+url.PathEscape(formIDs[0]), nil, nil, nil); err != nil {
		return fmt.Errorf("check google-forms: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: googleFormsStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a stream starts with an empty
// incremental cursor (full sync).
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
		stream = streamForms
	}
	switch stream {
	case streamForms, streamResponses, streamFormItems:
	default:
		return fmt.Errorf("google-forms stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	// Validate base URL early (SSRF guard) before any token exchange.
	if _, err := googleFormsBaseURL(req.Config); err != nil {
		return err
	}
	formIDs, err := formIDs(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	switch stream {
	case streamForms:
		return c.readForms(ctx, r, formIDs, false, emit)
	case streamFormItems:
		return c.readForms(ctx, r, formIDs, true, emit)
	default: // responses
		pageSize, err := googleFormsPageSize(req.Config)
		if err != nil {
			return err
		}
		filter := responseFilter(req)
		for _, id := range formIDs {
			if err := c.readResponses(ctx, r, id, pageSize, filter, emit); err != nil {
				return err
			}
		}
		return nil
	}
}

// readForms fetches each configured form's metadata. When items is true it emits
// one record per form item (the form_items stream); otherwise one record per form
// (the forms stream).
func (c Connector) readForms(ctx context.Context, r *connsdk.Requester, formIDs []string, items bool, emit func(connectors.Record) error) error {
	for _, id := range formIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, "forms/"+url.PathEscape(id), nil, nil)
		if err != nil {
			return fmt.Errorf("read google-forms form %s: %w", id, err)
		}
		form, err := connsdk.RecordsAt(resp.Body, ".")
		if err != nil {
			return fmt.Errorf("decode google-forms form %s: %w", id, err)
		}
		if len(form) == 0 {
			continue
		}
		obj := map[string]any(form[0])
		if items {
			for _, rec := range mapFormItemRecords(id, obj) {
				if err := emit(rec); err != nil {
					return err
				}
			}
			continue
		}
		if err := emit(mapFormRecord(obj)); err != nil {
			return err
		}
	}
	return nil
}

// readResponses drives the Forms API responses pagination. List responses return
// {responses:[...], nextPageToken:string}; the next page is requested with
// pageToken=<token>. The loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) readResponses(ctx context.Context, r *connsdk.Requester, formID string, pageSize int, filter string, emit func(connectors.Record) error) error {
	path := "forms/" + url.PathEscape(formID) + "/responses"
	pageToken := ""
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("pageSize", strconv.Itoa(pageSize))
		if filter != "" {
			query.Set("filter", filter)
		}
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read google-forms responses %s: %w", formID, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "responses")
		if err != nil {
			return fmt.Errorf("decode google-forms responses %s: %w", formID, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapResponseRecord(formID, item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "nextPageToken")
		if err != nil {
			return fmt.Errorf("decode google-forms nextPageToken %s: %w", formID, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		pageToken = next
	}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	formID := "fixture_form_1"
	if ids := strings.TrimSpace(req.Config.Config["form_id"]); ids != "" {
		if parts := splitFormIDs(ids); len(parts) > 0 {
			formID = parts[0]
		}
	}
	switch stream {
	case streamForms:
		for i := 1; i <= 2; i++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			rec := connectors.Record{
				"form_id":        fmt.Sprintf("fixture_form_%d", i),
				"revision_id":    fmt.Sprintf("0000000%d", i),
				"responder_uri":  fmt.Sprintf("https://docs.google.com/forms/d/e/fixture_%d/viewform", i),
				"title":          fmt.Sprintf("Fixture Form %d", i),
				"document_title": fmt.Sprintf("Fixture Form %d Doc", i),
				"description":    "Deterministic fixture form.",
				"item_count":     2,
				"connector":      "google-forms",
				"fixture":        true,
			}
			if err := emit(rec); err != nil {
				return err
			}
		}
	case streamFormItems:
		for i := 1; i <= 2; i++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			rec := connectors.Record{
				"form_id":     formID,
				"item_id":     fmt.Sprintf("item_%d", i),
				"title":       fmt.Sprintf("Question %d", i),
				"description": "Deterministic fixture item.",
				"question_id": fmt.Sprintf("q_%d", i),
				"connector":   "google-forms",
				"fixture":     true,
			}
			if err := emit(rec); err != nil {
				return err
			}
		}
	default: // responses
		for i := 1; i <= 2; i++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			rec := connectors.Record{
				"response_id":         fmt.Sprintf("response_%d", i),
				"form_id":             formID,
				"create_time":         googleFormsFixtureTime,
				"last_submitted_time": googleFormsFixtureTime,
				"respondent_email":    fmt.Sprintf("fixture+%d@example.com", i),
				"total_score":         i,
				"answers":             map[string]any{"q_1": map[string]any{"textAnswers": map[string]any{"answers": []any{map[string]any{"value": fmt.Sprintf("answer %d", i)}}}}},
				"connector":           "google-forms",
				"fixture":             true,
			}
			if cursor := req.State["cursor"]; cursor != "" {
				rec["previous_cursor"] = cursor
			}
			if err := emit(rec); err != nil {
				return err
			}
		}
	}
	return nil
}

// Write is unsupported: the Google Forms connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// requester builds a connsdk.Requester wired with the OAuth2 refresh-token
// authenticator and the resolved base URL. Secrets only ever flow into the
// authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := googleFormsBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireSecrets(cfg); err != nil {
		return nil, err
	}
	tokenURL, err := googleFormsTokenURL(cfg)
	if err != nil {
		return nil, err
	}
	auth := &oauthRefreshAuth{
		tokenURL:     tokenURL,
		clientID:     cfg.Secrets["client_id"],
		clientSecret: cfg.Secrets["client_secret"],
		refreshToken: cfg.Secrets["client_refresh_token"],
		client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: googleFormsUserAgent,
	}, nil
}

// oauthRefreshAuth implements connsdk.Authenticator using the OAuth2 refresh-token
// grant. It exchanges the long-lived refresh token for a short-lived access token
// at Google's token endpoint and caches it until shortly before expiry. The
// refresh token, client secret, and access token are never logged.
type oauthRefreshAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	refreshToken string
	client       *http.Client

	mu      sync.Mutex
	token   string
	expires time.Time
	now     func() time.Time
}

func (a *oauthRefreshAuth) nowFn() time.Time {
	if a.now != nil {
		return a.now()
	}
	return time.Now()
}

func (a *oauthRefreshAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *oauthRefreshAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Refresh 60s before expiry to avoid edge races.
	if a.token != "" && a.nowFn().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("google-forms: token URL is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)
	form.Set("refresh_token", a.refreshToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("google-forms: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("google-forms: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("google-forms: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("google-forms: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("google-forms: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.nowFn().Add(ttl)
	return a.token, nil
}

// responseFilter derives the Forms API responses filter from the incremental
// cursor (if any) or else the start_date config, both interpreted as a lower
// bound on submission timestamp. An empty result means no filter (full sync).
func responseFilter(req connectors.ReadRequest) string {
	bound := connsdk.Cursor(req.State)
	if bound == "" {
		bound = strings.TrimSpace(req.Config.Config["start_date"])
	}
	if bound == "" {
		return ""
	}
	return "timestamp >= " + bound
}

func requireSecrets(cfg connectors.RuntimeConfig) error {
	if cfg.Secrets == nil {
		return errors.New("google-forms connector requires secrets client_id, client_secret, client_refresh_token")
	}
	for _, name := range []string{"client_id", "client_secret", "client_refresh_token"} {
		if strings.TrimSpace(cfg.Secrets[name]) == "" {
			return fmt.Errorf("google-forms connector requires secret %s", name)
		}
	}
	return nil
}

// formIDs returns the configured form IDs (comma/space/newline separated).
func formIDs(cfg connectors.RuntimeConfig) ([]string, error) {
	ids := splitFormIDs(cfg.Config["form_id"])
	if len(ids) == 0 {
		return nil, errors.New("google-forms connector requires config form_id (one or more form IDs)")
	}
	return ids, nil
}

func splitFormIDs(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == ' ' || r == '\t' || r == '\r'
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	return out
}

// googleFormsBaseURL resolves and validates the base URL. The default is
// forms.googleapis.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func googleFormsBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	return resolveHTTPURL(cfg.Config["base_url"], googleFormsDefaultBaseURL, "base_url")
}

// googleFormsTokenURL resolves and validates the OAuth2 token endpoint, with the
// same SSRF validation as the base URL.
func googleFormsTokenURL(cfg connectors.RuntimeConfig) (string, error) {
	return resolveHTTPURL(cfg.Config["token_url"], googleFormsDefaultTokenURL, "token_url")
}

func resolveHTTPURL(raw, def, label string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("google-forms config %s is invalid: %w", label, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("google-forms config %s must use http or https, got %q", label, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("google-forms config %s must include a host", label)
	}
	return strings.TrimRight(raw, "/"), nil
}

func googleFormsPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return googleFormsDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("google-forms config page_size must be an integer: %w", err)
	}
	if value < 1 || value > googleFormsMaxPageSize {
		return 0, fmt.Errorf("google-forms config page_size must be between 1 and %d", googleFormsMaxPageSize)
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
