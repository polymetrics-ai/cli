// Package oncehub implements the native pm OnceHub connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference connector: a thin package that composes the connsdk toolkit
// (Requester + APIKeyHeader auth + RecordsAt extraction + cursor state) with
// OnceHub-specific stream definitions, endpoints, and pagination.
//
// OnceHub exposes a REST API at https://api.oncehub.com. Auth is an API-Key
// header. List endpoints return {"data":[...]} and paginate with an `after`
// cursor (set to the last record's id) plus a `limit` page size; the presence of
// a rel="next" Link header signals there is another page.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package oncehub

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
	oncehubDefaultBaseURL  = "https://api.oncehub.com"
	oncehubDefaultPageSize = 100
	oncehubMaxPageSize     = 100
	oncehubUserAgent       = "polymetrics-go-cli"
	// oncehubFixtureUpdated is the deterministic last_updated_time base used by
	// the fixture-mode records.
	oncehubFixtureUpdated = "2026-01-01T00:00:00.000Z"
	oncehubFixtureCreated = "2026-01-01T00:00:00.000Z"
)

func init() {
	connectors.RegisterFactory("oncehub", New)
}

// New returns the OnceHub connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm OnceHub connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "oncehub" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "oncehub",
		DisplayName:     "OnceHub",
		IntegrationType: "api",
		Description:     "Reads OnceHub bookings, contacts, booking pages, users, and event types through the OnceHub REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to OnceHub. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := oncehubBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(oncehubSecret(cfg)) == "" {
		return errors.New("oncehub connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the users list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "/v2/users", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check oncehub: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: oncehubStreams()}, nil
}

// Write satisfies the connectors.Connector interface. OnceHub is read-only in pm
// (no safe reverse-ETL action set), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: an OnceHub stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time.
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
		stream = "bookings"
	}
	endpoint, ok := oncehubStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("oncehub stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := oncehubPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := oncehubMaxPages(req.Config)
	if err != nil {
		return err
	}
	lower := ""
	if endpoint.incremental {
		lower = incrementalLowerBound(req)
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, lower, emit)
}

// harvest drives OnceHub's cursor pagination. List endpoints return
// {"data":[...]} and the next page is requested with after=<last record id>; the
// presence of a rel="next" Link header signals another page. There is no body
// token paginator in connsdk for this exact shape, so the loop lives here, built
// on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, updatedGT string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	if updatedGT != "" {
		base.Set("last_updated_time.gt", updatedGT)
	}

	after := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if after != "" {
			query.Set("after", after)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read oncehub %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode oncehub %s page: %w", endpoint.resource, err)
		}
		lastID := ""
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			lastID = stringField(item, "id")
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Stop unless the Link header advertises a next page (mirrors the
		// upstream manifest stop_condition) and we have a cursor to advance with.
		if !hasNextLink(resp.Header.Get("Link")) || lastID == "" {
			return nil
		}
		after = lastID
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise oncehub credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                   fmt.Sprintf("%s_fixture_%d", stream, i),
			"object":               strings.TrimSuffix(stream, "s"),
			"subject":              fmt.Sprintf("Fixture booking %d", i),
			"status":               "scheduled",
			"booking_page":         "page_fixture_1",
			"event_type":           "evt_fixture_1",
			"contact":              "contact_fixture_1",
			"owner":                "user_fixture_1",
			"starting_time":        oncehubFixtureCreated,
			"duration_minutes":     30,
			"customer_timezone":    "UTC",
			"location_description": "Online",
			"tracking_id":          fmt.Sprintf("trk_%d", i),
			"in_trash":             false,
			"first_name":           fmt.Sprintf("Fixture %d", i),
			"last_name":            "Example",
			"email":                fmt.Sprintf("fixture+%d@example.com", i),
			"mobile_phone":         "+10000000000",
			"timezone":             "UTC",
			"role_name":            "member",
			"name":                 fmt.Sprintf("Fixture %d", i),
			"label":                fmt.Sprintf("fixture-%d", i),
			"url":                  fmt.Sprintf("https://oncehub.example/%d", i),
			"active":               true,
			"creation_time":        oncehubFixtureCreated,
			"last_updated_time":    oncehubFixtureUpdated,
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

// requester builds a connsdk.Requester wired with API-Key header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := oncehubBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := oncehubSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("oncehub connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("API-Key", secret, ""),
		UserAgent: oncehubUserAgent,
	}, nil
}

// incrementalLowerBound returns the last_updated_time.gt lower bound, derived
// from the incremental cursor (if any) or else the start_date config. An empty
// result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func oncehubSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// oncehubBaseURL resolves and validates the base URL. The default is
// api.oncehub.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func oncehubBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return oncehubDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("oncehub config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("oncehub config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("oncehub config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func oncehubPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return oncehubDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("oncehub config page_size must be an integer: %w", err)
	}
	if value < 1 || value > oncehubMaxPageSize {
		return 0, fmt.Errorf("oncehub config page_size must be between 1 and %d", oncehubMaxPageSize)
	}
	return value, nil
}

func oncehubMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("oncehub config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("oncehub config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// hasNextLink reports whether an RFC 5988 Link header value advertises a
// rel="next" page. OnceHub uses this header as its pagination stop signal.
func hasNextLink(header string) bool {
	if header == "" {
		return false
	}
	for _, part := range strings.Split(header, ",") {
		segs := strings.Split(part, ";")
		if len(segs) < 2 {
			continue
		}
		for _, attr := range segs[1:] {
			attr = strings.TrimSpace(attr)
			if attr == `rel="next"` || attr == "rel=next" {
				return true
			}
		}
	}
	return false
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

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
