// Package buzzsprout implements the native pm Buzzsprout connector. It follows
// the stripe declarative-HTTP template: a thin package that composes the connsdk
// toolkit (Requester + token-header auth + RecordsAt extraction) with
// Buzzsprout-specific stream definitions and endpoints.
//
// Buzzsprout's API (https://github.com/Buzzsprout/buzzsprout-api) is account- and
// podcast-scoped: episode endpoints live under /api/{podcast_id}/... while the
// podcasts index is account-level at /api/podcasts.json. Every path ends in
// .json, auth is an "Authorization: Token token=<api_key>" header, and list
// responses are top-level JSON arrays with no server pagination. This connector
// is read-only.
//
// Like stripe/github, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package buzzsprout

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	buzzsproutDefaultBaseURL  = "https://www.buzzsprout.com"
	buzzsproutDefaultPageSize = 100
	buzzsproutMaxPageSize     = 1000
	buzzsproutMaxPagesGuard   = 1000
	buzzsproutUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("buzzsprout", New)
}

// New returns the Buzzsprout connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Buzzsprout connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "buzzsprout" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "buzzsprout",
		DisplayName:     "Buzzsprout",
		IntegrationType: "api",
		Description:     "Reads Buzzsprout podcasts and episodes (titles, publish dates, durations, play counts) through the Buzzsprout REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

// Check verifies the connector is configured well enough to talk to Buzzsprout.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := buzzsproutBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(buzzsproutSecret(cfg)) == "" {
		return errors.New("buzzsprout connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the account-level podcasts list confirms auth and
	// connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "api/podcasts.json", nil, nil, nil); err != nil {
		return fmt.Errorf("check buzzsprout: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: buzzsproutStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Buzzsprout stream starts
// with an empty incremental cursor (full sync), which the start_date config can
// raise at read time for the episodes stream.
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
		stream = "episodes"
	}
	endpoint, ok := buzzsproutStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("buzzsprout stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	path, err := endpointPath(req.Config, endpoint)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := buzzsproutPageSize(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, path, endpoint, pageSize, emit)
}

// Write is unsupported: Buzzsprout is exposed read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest reads the endpoint. Buzzsprout returns a single top-level JSON array
// with no native pagination, so production reads exactly one page. To stay
// robust against any future/page-style behavior (and to support test servers
// that paginate), the loop requests page=N and stops on a short or empty page.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, pageSize int, emit func(connectors.Record) error) error {
	mapRecord := endpoint.mapRecord

	for page := 1; page <= buzzsproutMaxPagesGuard; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read buzzsprout %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode buzzsprout %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer than the page size) or an empty page means there
		// is no more data. Buzzsprout always returns the full set on page 1, so
		// this terminates after one request in production.
		if pageSize <= 0 || len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise buzzsprout credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":              int64(i),
			"title":           fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"published_at":    fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"duration":        int64(600 * i),
			"episode_number":  int64(i),
			"season_number":   int64(1),
			"total_plays":     int64(100 * i),
			"explicit":        false,
			"private":         false,
			"hq":              true,
			"author":          fmt.Sprintf("Fixture Author %d", i),
			"language":        "en",
			"description":     "Fixture record (no network).",
			"connector":       "buzzsprout",
			"fixture":         true,
			"guid":            fmt.Sprintf("%s_fixture_%d", stream, i),
			"website_address": "https://www.buzzsprout.com",
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

// requester builds a connsdk.Requester wired with token-header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := buzzsproutBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := buzzsproutSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("buzzsprout connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, "Token token="),
		UserAgent: buzzsproutUserAgent,
		DefaultHeaders: map[string]string{
			"Content-Type": "application/json; charset=utf-8",
		},
	}, nil
}

// endpointPath builds the resolved API path for an endpoint, scoping by
// podcast_id for non-account-level resources.
func endpointPath(cfg connectors.RuntimeConfig, endpoint streamEndpoint) (string, error) {
	if endpoint.accountLevel {
		return "api/" + endpoint.resource, nil
	}
	podcastID := strings.TrimSpace(cfg.Config["podcast_id"])
	if podcastID == "" {
		return "", errors.New("buzzsprout connector requires config podcast_id for this stream")
	}
	if !isSafePathSegment(podcastID) {
		return "", fmt.Errorf("buzzsprout config podcast_id %q contains invalid characters", podcastID)
	}
	return "api/" + podcastID + "/" + endpoint.resource, nil
}

func buzzsproutSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// buzzsproutBaseURL resolves and validates the base URL. The default is
// www.buzzsprout.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func buzzsproutBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return buzzsproutDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("buzzsprout config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("buzzsprout config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("buzzsprout config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func buzzsproutPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return buzzsproutDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("buzzsprout config page_size must be an integer: %w", err)
	}
	if value < 1 || value > buzzsproutMaxPageSize {
		return 0, fmt.Errorf("buzzsprout config page_size must be between 1 and %d", buzzsproutMaxPageSize)
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// isSafePathSegment ensures a config-provided value cannot break out of the URL
// path (no slashes, no dot-segments, no control characters).
func isSafePathSegment(s string) bool {
	if s == "" || s == "." || s == ".." {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return false
		}
	}
	return true
}
