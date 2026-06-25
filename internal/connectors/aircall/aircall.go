// Package aircall implements the native pm Aircall connector. It follows the
// declarative-HTTP template established by the stripe connector: a thin package
// that composes the connsdk toolkit (Requester + Basic auth + RecordsAt
// extraction + meta.next_page_link pagination) with Aircall-specific stream
// definitions and endpoints.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// Aircall's REST API (https://developer.aircall.io/api-references/) uses HTTP
// Basic auth with api_id as the username and api_token as the password, returns
// list payloads keyed by the resource name (e.g. {"calls":[...]}) alongside a
// "meta" object, and paginates by following meta.next_page_link. This connector
// is read-only: the Aircall API has no obvious safe reverse-ETL surface.
package aircall

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
	aircallDefaultBaseURL  = "https://api.aircall.io/v1"
	aircallDefaultPageSize = 50
	aircallMaxPageSize     = 50
	aircallUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("aircall", New)
}

// New returns the Aircall connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Aircall connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk
	// Requester. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "aircall" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "aircall",
		DisplayName:     "Aircall",
		IntegrationType: "api",
		Description:     "Reads Aircall calls, users, contacts, numbers, and teams through the Aircall REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Aircall. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := aircallBaseURL(cfg); err != nil {
		return err
	}
	id, token := aircallSecrets(cfg)
	if strings.TrimSpace(id) == "" || strings.TrimSpace(token) == "" {
		return errors.New("aircall connector requires secrets api_id and api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the calls list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "calls", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check aircall: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: aircallStreams()}, nil
}

// Write satisfies the connectors.Connector interface. The Aircall connector is
// read-only (no safe reverse-ETL surface), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: an Aircall stream starts
// with an empty incremental cursor (full sync), which the start_date config can
// raise at read time.
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
		stream = "calls"
	}
	endpoint, ok := aircallStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("aircall stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := aircallPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := aircallMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, incrementalLowerBound(req), emit)
}

// harvest drives Aircall's meta.next_page_link pagination. Aircall list
// responses are {"<resource>":[...], "meta":{"next_page_link":<url|null>}}; the
// next page is fetched by following next_page_link verbatim (the Requester
// treats an absolute http(s) path as-is). The first request carries per_page and
// an optional `from` lower bound.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, fromUnix string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))
	if fromUnix != "" && (endpoint.resource == "calls" || endpoint.resource == "contacts") {
		base.Set("from", fromUnix)
	}

	path := endpoint.resource
	query := base
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read aircall %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode aircall %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "meta.next_page_link")
		if err != nil {
			return fmt.Errorf("decode aircall %s next_page_link: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return nil
		}
		// next_page_link is an absolute URL that already carries page/per_page;
		// follow it verbatim and drop the seed query so we do not duplicate it.
		path = next
		query = nil
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise aircall credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                  int64(i),
			"sid":                 fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"direction":           "inbound",
			"status":              "done",
			"started_at":          int64(1767225600 + i),
			"answered_at":         int64(1767225601 + i),
			"ended_at":            int64(1767225700 + i),
			"duration":            int64(60 * i),
			"raw_digits":          "+15551230000",
			"archived":            false,
			"name":                fmt.Sprintf("Fixture %d", i),
			"email":               fmt.Sprintf("fixture+%d@example.com", i),
			"available":           true,
			"availability_status": "available",
			"created_at":          "2026-01-01T00:00:00.000Z",
			"first_name":          fmt.Sprintf("First%d", i),
			"last_name":           fmt.Sprintf("Last%d", i),
			"company_name":        "Fixture Co",
			"is_shared":           false,
			"digits":              "+15551230000",
			"country":             "US",
			"open":                true,
			"connector":           "aircall",
			"fixture":             true,
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

// requester builds a connsdk.Requester wired with Basic auth and the resolved
// base URL. The secrets only ever flow into connsdk.Basic; they are never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := aircallBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	id, token := aircallSecrets(cfg)
	if strings.TrimSpace(id) == "" || strings.TrimSpace(token) == "" {
		return nil, errors.New("aircall connector requires secrets api_id and api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(id, token),
		UserAgent: aircallUserAgent,
	}, nil
}

// incrementalLowerBound returns the unix-seconds lower bound for the `from`
// filter, derived from the incremental cursor (if any) or else the start_date
// config. An empty result means no lower bound (full sync). start_date is an
// ISO-8601 instant per the Aircall catalog schema.
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		if u := toUnixSeconds(cursor); u != "" {
			return u
		}
		return cursor
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return ""
	}
	return toUnixSeconds(startDate)
}

func aircallSecrets(cfg connectors.RuntimeConfig) (string, string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return cfg.Secrets["api_id"], cfg.Secrets["api_token"]
}

// aircallBaseURL resolves and validates the base URL. The default is
// api.aircall.io/v1; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func aircallBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return aircallDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("aircall config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("aircall config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("aircall config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func aircallPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["per_page"])
	if raw == "" {
		return aircallDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("aircall config per_page must be an integer: %w", err)
	}
	if value < 1 || value > aircallMaxPageSize {
		return 0, fmt.Errorf("aircall config per_page must be between 1 and %d", aircallMaxPageSize)
	}
	return value, nil
}

func aircallMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("aircall config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("aircall config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// toUnixSeconds converts an RFC3339 instant (Aircall start_date is ISO-8601) or
// an already-numeric value into a unix-seconds string. It returns "" when the
// value cannot be interpreted.
func toUnixSeconds(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return value
	}
	for _, layout := range []string{"2006-01-02T15:04:05.000Z07:00", time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, value); err == nil {
			return strconv.FormatInt(t.Unix(), 10)
		}
	}
	return ""
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
