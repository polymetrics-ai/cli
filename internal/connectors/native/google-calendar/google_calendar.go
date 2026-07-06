// Package googlecalendar implements the native pm Google Calendar connector. It
// is a declarative-HTTP per-system connector built on the stripe template: a
// thin package that composes the connsdk toolkit (Requester + an in-package
// OAuth2 refresh-token authenticator + RecordsAt extraction + nextPageToken
// cursor pagination) with Google Calendar v3 stream definitions and endpoints.
//
// The directory and registry key are the hyphenated bare system name
// "google-calendar"; the Go package is "googlecalendar" (Go identifiers cannot
package googlecalendar

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	registryName   = "google-calendar"
	defaultBaseURL = "https://www.googleapis.com/calendar/v3"
	// defaultTokenURL is Google's OAuth2 token endpoint used for the
	// refresh-token grant.
	defaultTokenURL   = "https://oauth2.googleapis.com/token"
	defaultPageSize   = 250
	maxPageSize       = 2500
	userAgent         = "polymetrics-go-cli"
	defaultCalendarID = "primary"
	itemsPath         = "items"
	nextTokenPath     = "nextPageToken"
	pageTokenParam    = "pageToken"
	maxResultsParam   = "maxResults"
)

// New returns the Google Calendar connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Google Calendar connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Google Calendar",
		IntegrationType: "api",
		Description:     "Reads Google Calendar calendar lists, events, settings, and access control rules through the Calendar API v3 using an OAuth2 refresh token.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Google
// Calendar. In fixture mode it short-circuits without a network call.
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
	if err := requireSecrets(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the user's calendar list confirms auth and connectivity
	// without mutating anything.
	q := url.Values{maxResultsParam: []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "users/me/calendarList", q, nil, nil); err != nil {
		return fmt.Errorf("check google-calendar: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: googleCalendarStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a stream starts with an
// empty incremental cursor (full sync). Google Calendar's supported sync mode is
// full_refresh; the events stream can still carry an updated cursor.
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
		stream = "calendar_list"
	}
	endpoint, ok := googleCalendarStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("google-calendar stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	if err := requireSecrets(req.Config); err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	size, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	pages, err := maxPages(req.Config)
	if err != nil {
		return err
	}

	path := endpoint.path(calendarID(req.Config))
	base := url.Values{maxResultsParam: []string{strconv.Itoa(size)}}
	paginator := &connsdk.CursorPaginator{
		CursorParam: pageTokenParam,
		TokenPath:   nextTokenPath,
	}
	// connsdk.Harvest's emit hook takes connsdk.Record, an alias for
	// map[string]any, which is exactly what mapRecord consumes.
	wrapped := func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, path, base, paginator, itemsPath, pages, wrapped); err != nil {
		return fmt.Errorf("read google-calendar %s: %w", stream, err)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise google-calendar credential-free (mirrors the
// stripe fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":          fmt.Sprintf("%s_fixture_%d", stream, i),
			"kind":        "calendar#" + strings.TrimSuffix(stream, "s"),
			"etag":        fmt.Sprintf("\"etag_%d\"", i),
			"summary":     fmt.Sprintf("Fixture Calendar %d", i),
			"description": "fixture",
			"timeZone":    "UTC",
			"accessRole":  "owner",
			"primary":     i == 1,
			"selected":    true,
			"status":      "confirmed",
			"iCalUID":     fmt.Sprintf("fixture-%d@example.com", i),
			"location":    "Remote",
			"htmlLink":    "https://calendar.google.com/event?eid=fixture",
			"created":     "2026-01-01T00:00:00Z",
			"updated":     fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"start":       map[string]any{"dateTime": "2026-01-01T09:00:00Z"},
			"end":         map[string]any{"dateTime": "2026-01-01T09:30:00Z"},
			"role":        "owner",
			"scope":       map[string]any{"type": "user", "value": "fixture@example.com"},
			"value":       "UTC",
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

// requester builds a connsdk.Requester wired with the OAuth2 refresh-token
// authenticator and the resolved base URL. Secrets only ever flow into the
// authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireSecrets(cfg); err != nil {
		return nil, err
	}
	auth := &oauthRefreshAuth{
		tokenURL:     tokenURL(cfg),
		clientID:     secret(cfg, "client_id"),
		clientSecret: secret(cfg, "client_secret"),
		refreshToken: secret(cfg, "client_refresh_token_2"),
		client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: userAgent,
	}, nil
}

// requireSecrets ensures the three OAuth secrets are present.
func requireSecrets(cfg connectors.RuntimeConfig) error {
	for _, name := range []string{"client_id", "client_secret", "client_refresh_token_2"} {
		if strings.TrimSpace(secret(cfg, name)) == "" {
			return fmt.Errorf("google-calendar connector requires secret %s", name)
		}
	}
	return nil
}

func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}

// calendarID resolves the configured calendar id, defaulting to "primary".
func calendarID(cfg connectors.RuntimeConfig) string {
	if id := strings.TrimSpace(cfg.Config["calendarid"]); id != "" {
		return id
	}
	return defaultCalendarID
}

// baseURL resolves and validates the base URL. The default is
// www.googleapis.com/calendar/v3; any override must be an absolute https (or
// http for local test servers) URL with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	if err := validateURL(base, "base_url"); err != nil {
		return "", err
	}
	return strings.TrimRight(base, "/"), nil
}

// tokenURL resolves the OAuth2 token endpoint, allowing a config override (used
// by tests) that is validated for scheme+host.
func tokenURL(cfg connectors.RuntimeConfig) string {
	if override := strings.TrimSpace(cfg.Config["token_url"]); override != "" {
		if err := validateURL(override, "token_url"); err == nil {
			return strings.TrimRight(override, "/")
		}
	}
	return defaultTokenURL
}

func validateURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("google-calendar config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("google-calendar config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("google-calendar config %s must include a host", field)
	}
	return nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("google-calendar config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("google-calendar config page_size must be between 1 and %d", maxPageSize)
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
		return 0, fmt.Errorf("google-calendar config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("google-calendar config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Google Calendar is exposed as a read-only source (there
// are no safe, generic reverse-ETL write actions for calendars/events here), so
// Capabilities.Write is false and Write returns ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
