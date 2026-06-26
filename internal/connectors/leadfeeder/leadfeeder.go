// Package leadfeeder implements the native pm Leadfeeder source connector. It is a
// declarative-HTTP per-system connector modeled on the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + "Token token=" API-key
// auth + RecordsAt extraction + JSON:API links.next pagination) with
// Leadfeeder-specific stream definitions and endpoints.
//
// Leadfeeder exposes accounts and, nested under each account, leads, visits, and
// custom feeds. The connector is read-only: the Leadfeeder API offers no
// reverse-ETL writes that make sense here, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package leadfeeder

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
	leadfeederDefaultBaseURL  = "https://api.leadfeeder.com"
	leadfeederDefaultPageSize = 100
	leadfeederMaxPageSize     = 100
	leadfeederUserAgent       = "polymetrics-go-cli"
	// leadfeederAuthPrefix is the literal prefix Leadfeeder requires on the
	// Authorization header value: "Token token=<api_token>".
	leadfeederAuthPrefix = "Token token="
)

func init() {
	connectors.RegisterFactory("leadfeeder", New)
}

// New returns the Leadfeeder connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Leadfeeder source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "leadfeeder" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "leadfeeder",
		DisplayName:     "Leadfeeder",
		IntegrationType: "api",
		Description:     "Reads Leadfeeder accounts and their leads, visits, and custom feeds through the Leadfeeder JSON:API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Leadfeeder. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := leadfeederBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(leadfeederSecret(cfg)) == "" {
		return errors.New("leadfeeder connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the accounts list confirms auth and connectivity.
	q := url.Values{}
	q.Set("page[size]", "1")
	if err := r.DoJSON(ctx, http.MethodGet, "accounts", q, nil, nil); err != nil {
		return fmt.Errorf("check leadfeeder: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: leadfeederStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Leadfeeder is read-only:
// the API offers no reverse-ETL writes that make sense here, so Write always
// reports the operation as unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a Leadfeeder stream starts
// with an empty incremental cursor (full sync), which start_date can raise at read
// time.
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
		stream = "accounts"
	}
	endpoint, ok := leadfeederStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("leadfeeder stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := endpointPath(endpoint, req.Config)
	if err != nil {
		return err
	}
	pageSize, err := leadfeederPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := leadfeederMaxPages(req.Config)
	if err != nil {
		return err
	}
	base := url.Values{}
	base.Set("page[size]", strconv.Itoa(pageSize))
	if endpoint.dateScoped {
		start, end, err := dateWindow(req)
		if err != nil {
			return err
		}
		if start != "" {
			base.Set("start_date", start)
		}
		if end != "" {
			base.Set("end_date", end)
		}
	}
	return c.harvest(ctx, r, path, base, maxPages, endpoint.mapRecord, emit)
}

// harvest drives Leadfeeder's JSON:API pagination. List responses are
// {data:[...], links:{next:<url|null>}}; the next page is the absolute URL at
// links.next, followed until it is null/absent. connsdk.Requester treats an
// absolute http(s) path as-is, so following links.next is a direct Do call.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, base url.Values, maxPages int, mapRecord func(map[string]any) connectors.Record, emit func(connectors.Record) error) error {
	nextURL := ""
	query := base
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		reqPath := path
		reqQuery := query
		if nextURL != "" {
			// links.next is an absolute URL carrying its own page params; do not
			// merge the base query into it (it already encodes page[size] etc.).
			reqPath = nextURL
			reqQuery = nil
		}
		resp, err := r.Do(ctx, http.MethodGet, reqPath, reqQuery, nil)
		if err != nil {
			return fmt.Errorf("read leadfeeder %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode leadfeeder %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "links.next")
		if err != nil {
			return fmt.Errorf("decode leadfeeder %s links.next: %w", path, err)
		}
		next = strings.TrimSpace(next)
		if next == "" || strings.EqualFold(next, "null") {
			return nil
		}
		nextURL = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise leadfeeder credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":   fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"type": stream,
			"attributes": map[string]any{
				"name":             fmt.Sprintf("Fixture %s %d", stream, i),
				"industry":         "Software",
				"status":           "active",
				"time_zone":        "UTC",
				"currency":         "USD",
				"website":          fmt.Sprintf("https://example-%d.test", i),
				"city":             "Helsinki",
				"country":          "FI",
				"employee_count":   "11-50",
				"quality":          int64(i),
				"visits":           int64(10 * i),
				"first_visit_date": "2026-01-01",
				"last_visit_date":  "2026-01-02",
				"visit_length":     int64(120 * i),
				"pageviews":        int64(3 * i),
				"source":           "organic",
				"referring_url":    "https://google.com",
				"hostname":         "example.test",
				"visit_date":       "2026-01-02",
				"started_at":       "2026-01-02T10:00:00Z",
				"ended_at":         "2026-01-02T10:05:00Z",
			},
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

// requester builds a connsdk.Requester wired with "Token token=" API-key auth and
// the resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it
// is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := leadfeederBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := leadfeederSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("leadfeeder connector requires secret api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, leadfeederAuthPrefix),
		UserAgent: leadfeederUserAgent,
	}, nil
}

// endpointPath builds the request path for a stream, validating that
// account-scoped streams have an account_id config.
func endpointPath(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !endpoint.accountScoped {
		return endpoint.resource, nil
	}
	accountID := strings.TrimSpace(cfg.Config["account_id"])
	if accountID == "" {
		return "", fmt.Errorf("leadfeeder stream %q requires config account_id", endpoint.resource)
	}
	return "accounts/" + url.PathEscape(accountID) + "/" + endpoint.resource, nil
}

// dateWindow resolves the start_date/end_date query params (yyyy-mm-dd) for
// date-scoped streams. The lower bound comes from the incremental cursor (if any)
// or else the start_date config; the upper bound from end_date config or today.
func dateWindow(req connectors.ReadRequest) (string, string, error) {
	start := ""
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		start = cursor
	} else if sd := strings.TrimSpace(req.Config.Config["start_date"]); sd != "" {
		start = sd
	}
	startDay, err := toDateOnly(start)
	if err != nil {
		return "", "", fmt.Errorf("leadfeeder config start_date: %w", err)
	}
	end := strings.TrimSpace(req.Config.Config["end_date"])
	endDay, err := toDateOnly(end)
	if err != nil {
		return "", "", fmt.Errorf("leadfeeder config end_date: %w", err)
	}
	if startDay != "" && endDay == "" {
		// Leadfeeder requires both start_date and end_date when start is given;
		// default the upper bound to today (UTC).
		endDay = time.Now().UTC().Format("2006-01-02")
	}
	return startDay, endDay, nil
}

// toDateOnly normalizes an RFC3339 timestamp or yyyy-mm-dd string to yyyy-mm-dd.
// An empty input returns "".
func toDateOnly(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC().Format("2006-01-02"), nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t.Format("2006-01-02"), nil
	}
	return "", fmt.Errorf("%q must be RFC3339 or yyyy-mm-dd", value)
}

func leadfeederSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

// leadfeederBaseURL resolves and validates the base URL. The default is
// api.leadfeeder.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func leadfeederBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return leadfeederDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("leadfeeder config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("leadfeeder config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("leadfeeder config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func leadfeederPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return leadfeederDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("leadfeeder config page_size must be an integer: %w", err)
	}
	if value < 1 || value > leadfeederMaxPageSize {
		return 0, fmt.Errorf("leadfeeder config page_size must be between 1 and %d", leadfeederMaxPageSize)
	}
	return value, nil
}

func leadfeederMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("leadfeeder config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("leadfeeder config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
