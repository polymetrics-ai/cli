// Package gmail implements the native pm Gmail connector. It follows the
// declarative-HTTP shape established by the stripe connector: a thin package that
// composes the connsdk toolkit (Requester + cursor pagination + RecordsAt
// extraction) with Gmail-API-specific stream definitions and endpoints.
//
// The upstream source-gmail connector reads the Gmail REST API
// (https://gmail.googleapis.com/gmail/v1), authenticated with a Google OAuth 2.0
// refresh-token grant. This connector exposes the four core list resources of
// that API as streams — messages, threads, drafts, and labels — which are the
// id-bearing surfaces the upstream source enumerates. It is read-only: the
// upstream source supports only full_refresh and Gmail has no safe reverse-ETL
// write surface here.
//
// Auth: client_id, client_secret, and client_refresh_token secrets are exchanged
// for a short-lived access token at the Google token endpoint (see auth.go); only
// the resolved access token is sent (as a Bearer header) to the Gmail API. The
// secret values are never logged.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package gmail

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	registryName      = "gmail"
	defaultBaseURL    = "https://gmail.googleapis.com/gmail/v1"
	defaultTokenURL   = "https://oauth2.googleapis.com/token"
	defaultScope      = "https://www.googleapis.com/auth/gmail.readonly"
	defaultUserID     = "me"
	defaultPageSize   = 100
	maxPageSize       = 500
	userAgent         = "polymetrics-go-cli"
	pageTokenParam    = "pageToken"
	nextPageTokenPath = "nextPageToken"
	maxResultsParam   = "maxResults"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Gmail connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Gmail connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Gmail",
		IntegrationType: "api",
		Description:     "Reads Gmail messages, threads, drafts, and labels via the Google OAuth 2.0 refresh-token grant.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

// Check verifies the connector is configured well enough to talk to the Gmail
// API. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := baseURL(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A read of the labels list confirms auth and connectivity without mutating
	// anything; the labels endpoint is small and unpaginated.
	path := fmt.Sprintf("users/%s/labels", userIDPathSegment(cfg))
	if err := r.DoJSON(ctx, http.MethodGet, path, nil, nil, nil); err != nil {
		return fmt.Errorf("check gmail: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: gmailStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: each stream starts with an
// empty incremental cursor (full sync). The connector advertises only
// full_refresh upstream, but the cursor scaffolding is kept so an incremental
// mode can be layered on later without an interface change.
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
		stream = "messages"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("gmail stream %q not found", stream)
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
	resourcePath := fmt.Sprintf(endpoint.resource, userIDPathSegment(req.Config))

	base := url.Values{}
	applyListFilters(base, req.Config)

	if !endpoint.paginated {
		// Single-page resources (labels) are read directly; pagination params do
		// not apply.
		resp, err := r.Do(ctx, http.MethodGet, resourcePath, base, nil)
		if err != nil {
			return fmt.Errorf("read gmail %s: %w", stream, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode gmail %s: %w", stream, err)
		}
		for _, rec := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(rec)); err != nil {
				return err
			}
		}
		return nil
	}

	base.Set(maxResultsParam, strconv.Itoa(pageSize))
	paginator := &connsdk.CursorPaginator{
		CursorParam: pageTokenParam,
		TokenPath:   nextPageTokenPath,
		FirstQuery:  base,
	}
	return connsdk.Harvest(ctx, r, http.MethodGet, resourcePath, nil, paginator, endpoint.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

// Write reports that this connector is read-only. The Gmail source has no safe
// reverse-ETL surface in this connector, so Capabilities.Write is false and Write
// always returns ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free (mirrors the
// stripe fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                    fmt.Sprintf("%s_fixture_%d", stream, i),
			"threadId":              fmt.Sprintf("thread_fixture_%d", i),
			"snippet":               fmt.Sprintf("Fixture %s snippet %d", stream, i),
			"historyId":             strconv.Itoa(1000 + i),
			"name":                  fmt.Sprintf("Fixture %s %d", stream, i),
			"type":                  "user",
			"messageListVisibility": "show",
			"labelListVisibility":   "labelShow",
			"messagesTotal":         int64(i),
			"messagesUnread":        int64(0),
			"threadsTotal":          int64(i),
			"threadsUnread":         int64(0),
			"message": map[string]any{
				"id":       fmt.Sprintf("msg_fixture_%d", i),
				"threadId": fmt.Sprintf("thread_fixture_%d", i),
			},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the Google OAuth refresh-token
// authenticator and the resolved base URL. Secrets only ever flow into the
// authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	tokenURL, err := tokenURL(cfg)
	if err != nil {
		return nil, err
	}
	clientID := secret(cfg, "client_id")
	clientSecret := secret(cfg, "client_secret")
	refreshToken := refreshTokenSecret(cfg)
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(refreshToken) == "" {
		return nil, errors.New("gmail connector requires secrets credentials.client_id and credentials.client_refresh_token")
	}
	auth := &oauthRefreshAuth{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		scope:        defaultScope,
		client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: userAgent,
	}, nil
}

// applyListFilters adds the optional includeSpamTrash flag and a Gmail search
// query (q) derived from start_date to the list request.
func applyListFilters(query url.Values, cfg connectors.RuntimeConfig) {
	if includeSpamTrash(cfg) {
		query.Set("includeSpamTrash", "true")
	}
	if q := startDateQuery(cfg); q != "" {
		query.Set("q", q)
	}
}

// includeSpamTrash reports whether SPAM and TRASH should be included, from the
// include_spam_and_trash config (default false).
func includeSpamTrash(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["include_spam_and_trash"]), "true")
}

// startDateQuery converts the RFC3339 start_date config into a Gmail search
// query of the form "after:<unix-seconds>", or returns "" when unset. An
// unparseable value yields no filter rather than an error so a single bad config
// does not block a full read.
func startDateQuery(cfg connectors.RuntimeConfig) string {
	raw := strings.TrimSpace(cfg.Config["start_date"])
	if raw == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return ""
	}
	return "after:" + strconv.FormatInt(t.Unix(), 10)
}

// userIDPathSegment resolves the {userId} path segment, defaulting to "me", and
// URL-escapes it (an email address may be supplied for delegated access).
func userIDPathSegment(cfg connectors.RuntimeConfig) string {
	user := strings.TrimSpace(cfg.Config["user_id"])
	if user == "" {
		user = defaultUserID
	}
	return url.PathEscape(user)
}

// secret resolves a secret by its flattened last-segment key, falling back to the
// dotted credentials.<key> form the catalog declares.
func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	if v := strings.TrimSpace(cfg.Secrets[key]); v != "" {
		return v
	}
	return strings.TrimSpace(cfg.Secrets["credentials."+key])
}

// refreshTokenSecret resolves the refresh token, which the Gmail catalog names
// client_refresh_token. It also accepts the generic refresh_token key for
// flexibility across runtime secret-flattening schemes.
func refreshTokenSecret(cfg connectors.RuntimeConfig) string {
	if v := secret(cfg, "client_refresh_token"); v != "" {
		return v
	}
	return secret(cfg, "refresh_token")
}

// baseURL resolves and validates the base URL. The default is the Gmail API host;
// any override must be an absolute http/https URL with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validatedURL(cfg.Config["base_url"], defaultBaseURL, "base_url")
}

// tokenURL resolves the OAuth token endpoint. A token_url override wins;
// otherwise the Google default is used.
func tokenURL(cfg connectors.RuntimeConfig) (string, error) {
	return validatedURL(cfg.Config["token_url"], defaultTokenURL, "token_url")
}

func validatedURL(raw, fallback, field string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("gmail config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("gmail config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("gmail config %s must include a host", field)
	}
	return strings.TrimRight(raw, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gmail config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("gmail config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gmail config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("gmail config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
