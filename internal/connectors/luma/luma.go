// Package luma implements the native pm Luma (lu.ma) connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference connector: a thin package that composes the connsdk toolkit
// (Requester + APIKeyHeader auth + RecordsAt extraction) with Luma-specific
// stream definitions, endpoints, and cursor pagination.
//
// Luma's public API is read-only event management (events, guests, hosts); there
// is no safe reverse-ETL surface, so the connector advertises Write=false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package luma

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	lumaDefaultBaseURL = "https://api.lu.ma/public/v1"
	lumaAPIKeyHeader   = "x-luma-api-key"
	lumaUserAgent      = "polymetrics-go-cli"
	// lumaMaxPages bounds an unbounded cursor walk as a safety valve; overridable
	// via the max_pages config.
	lumaDefaultMaxPages = 0 // 0 = unlimited
)

func init() {
	connectors.RegisterFactory("luma", New)
}

// New returns the Luma connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Luma connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "luma" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "luma",
		DisplayName:     "lu.ma",
		IntegrationType: "api",
		Description:     "Reads lu.ma events, event guests, and event hosts through the Luma public REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Luma. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := lumaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(lumaSecret(cfg)) == "" {
		return errors.New("luma connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the events list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "calendar/list-events", nil, nil, nil); err != nil {
		return fmt.Errorf("check luma: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: lumaStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Luma's public API has no
// safe reverse-ETL surface, so the connector is read-only and Write is
// unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "events"
	}
	endpoint, ok := lumaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("luma stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	if endpoint.requiresEventID {
		eventID := strings.TrimSpace(req.Config.Config["event_api_id"])
		if eventID == "" {
			return fmt.Errorf("luma stream %q requires config event_api_id", stream)
		}
		base.Set("event_api_id", eventID)
	}
	maxPages := lumaMaxPages(req.Config)
	return c.harvest(ctx, r, endpoint, base, maxPages, emit)
}

// harvest drives Luma's cursor pagination. Luma list responses look like
// {entries:[{<entryKey>:{...}}, ...], has_more:bool, next_cursor:string}; the
// next page is requested with pagination_cursor=<next_cursor>. The wrapper shape
// (records nested under entries[].<entryKey>) has no direct connsdk paginator, so
// the loop lives here, built on connsdk.Requester + connsdk.RecordsAt +
// connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, maxPages int, emit func(connectors.Record) error) error {
	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if cursor != "" {
			query.Set("pagination_cursor", cursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read luma %s: %w", endpoint.resource, err)
		}
		entries, err := connsdk.RecordsAt(resp.Body, "entries")
		if err != nil {
			return fmt.Errorf("decode luma %s page: %w", endpoint.resource, err)
		}
		for _, entry := range entries {
			if err := ctx.Err(); err != nil {
				return err
			}
			item := unwrapEntry(entry, endpoint.entryKey)
			if item == nil {
				continue
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		hasMore, err := connsdk.StringAt(resp.Body, "has_more")
		if err != nil {
			return fmt.Errorf("decode luma %s has_more: %w", endpoint.resource, err)
		}
		next, err := connsdk.StringAt(resp.Body, "next_cursor")
		if err != nil {
			return fmt.Errorf("decode luma %s next_cursor: %w", endpoint.resource, err)
		}
		if hasMore != "true" || strings.TrimSpace(next) == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

// unwrapEntry pulls the actual record out of a Luma entries[] wrapper. When
// entryKey is set, the record lives at entry[entryKey]; otherwise the entry is
// itself the record.
func unwrapEntry(entry map[string]any, entryKey string) map[string]any {
	if entryKey == "" {
		return entry
	}
	inner, ok := entry[entryKey].(map[string]any)
	if !ok {
		return nil
	}
	return inner
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise luma credential-free (mirrors stripe's fixture
// intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"api_id":          fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":            fmt.Sprintf("Fixture %s %d", stream, i),
			"description":     "deterministic fixture record",
			"start_at":        fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"end_at":          fmt.Sprintf("2026-01-0%dT01:00:00Z", i),
			"timezone":        "America/New_York",
			"url":             fmt.Sprintf("https://lu.ma/fixture-%d", i),
			"cover_url":       "",
			"visibility":      "public",
			"created_at":      "2026-01-01T00:00:00Z",
			"calendar_api_id": "cal_fixture_1",
			"event_api_id":    "evt_fixture_1",
			"email":           fmt.Sprintf("fixture+%d@example.com", i),
			"approval_status": "approved",
			"registered_at":   "2026-01-01T00:00:00Z",
			"checked_in_at":   "",
			"user_api_id":     fmt.Sprintf("usr_fixture_%d", i),
			"user_name":       fmt.Sprintf("Fixture User %d", i),
			"avatar_url":      "",
			"access_level":    "admin",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the x-luma-api-key header and
// the resolved base URL. The secret only ever flows into the auth header; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := lumaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := lumaSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("luma connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(lumaAPIKeyHeader, secret, ""),
		UserAgent: lumaUserAgent,
	}, nil
}

func lumaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// lumaBaseURL resolves and validates the base URL. The default is api.lu.ma; any
// override must be an absolute https (or http for local test servers) URL with a
// host to bound SSRF risk.
func lumaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return lumaDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("luma config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("luma config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("luma config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func lumaMaxPages(cfg connectors.RuntimeConfig) int {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return lumaDefaultMaxPages
	}
	value := 0
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return lumaDefaultMaxPages
		}
		value = value*10 + int(ch-'0')
	}
	return value
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
