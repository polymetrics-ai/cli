// Package delighted implements the native pm Delighted connector. It is a
// declarative-HTTP per-system connector copied from the stripe template: a thin
// package composing the connsdk toolkit (Requester + HTTP Basic auth +
// page-number pagination over top-level JSON arrays) with Delighted-specific
// stream definitions and endpoints.
//
// The Delighted API authenticates via HTTP Basic auth using the API key as the
// username with a blank password. List endpoints (survey_responses, people,
// bounces, unsubscribes) return a top-level JSON array and paginate by page
// number with a per_page size (max 100); a short final page signals the end.
// The metrics endpoint returns a single JSON object.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package delighted

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	delightedDefaultBaseURL  = "https://api.delighted.com/v1"
	delightedDefaultPageSize = 100
	delightedMaxPageSize     = 100
	delightedUserAgent       = "polymetrics-go-cli"
	// delightedFixtureCreated is the deterministic created/updated timestamp used
	// by the fixture-mode records (2026-01-01T00:00:00Z in unix seconds).
	delightedFixtureCreated int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("delighted", New)
}

// New returns the Delighted connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Delighted connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "delighted" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "delighted",
		DisplayName:     "Delighted",
		IntegrationType: "api",
		Description:     "Reads Delighted survey responses, people, bounces, unsubscribes, and aggregate metrics through the Delighted REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Delighted. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := delightedBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(delightedSecret(cfg)) == "" {
		return errors.New("delighted connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the metrics endpoint confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "metrics.json", nil, nil, nil); err != nil {
		return fmt.Errorf("check delighted: %w", err)
	}
	return nil
}

// Write is unsupported: Delighted is a read-only source connector. The method
// exists to satisfy connectors.Connector; Capabilities.Write is false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: delightedStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Delighted stream starts
// with an empty incremental cursor (full sync), which the since config can raise
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
		stream = "survey_responses"
	}
	endpoint, ok := delightedStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("delighted stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := delightedPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := delightedMaxPages(req.Config)
	if err != nil {
		return err
	}
	base := url.Values{}
	if endpoint.supportsSince {
		since, err := incrementalLowerBound(req)
		if err != nil {
			return err
		}
		if since != "" {
			base.Set("since", since)
		}
	}

	if endpoint.single {
		return c.readSingle(ctx, r, endpoint, base, emit)
	}
	return c.harvest(ctx, r, endpoint, base, pageSize, maxPages, emit)
}

// harvest drives Delighted's page-number pagination. List endpoints return a
// top-level JSON array; the next page is requested with page=N+1 and per_page=N.
// A page shorter than per_page (or empty) signals the last page. This shape is
// not covered by connsdk.Harvest's recordsPath model cleanly because the cursor
// lives in the response length, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("per_page", strconv.Itoa(pageSize))
		query.Set("page", strconv.Itoa(page))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read delighted %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode delighted %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer records than requested) is the last page.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readSingle reads an endpoint that returns one JSON object (metrics).
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, base, nil)
	if err != nil {
		return fmt.Errorf("read delighted %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode delighted %s: %w", endpoint.resource, err)
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

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise delighted credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	count := 2
	if endpoint.single {
		count = 1
	}
	for i := 1; i <= count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                       fmt.Sprintf("%s_fixture_%d", strings.TrimSuffix(stream, "s"), i),
			"person":                   fmt.Sprintf("person_fixture_%d", i),
			"person_id":                fmt.Sprintf("person_fixture_%d", i),
			"survey_type":              "nps",
			"score":                    int64(7 + i),
			"comment":                  fmt.Sprintf("Fixture comment %d", i),
			"permalink":                fmt.Sprintf("https://app.delighted.com/r/fixture_%d", i),
			"created_at":               delightedFixtureCreated + int64(i),
			"updated_at":               delightedFixtureCreated + int64(i),
			"name":                     fmt.Sprintf("Fixture %d", i),
			"email":                    fmt.Sprintf("fixture+%d@example.com", i),
			"phone_number":             "",
			"last_sent_at":             delightedFixtureCreated,
			"last_responded_at":        delightedFixtureCreated + int64(i),
			"next_survey_scheduled_at": delightedFixtureCreated + 86400,
			"bounced_at":               delightedFixtureCreated + int64(i),
			"unsubscribed_at":          delightedFixtureCreated + int64(i),
			"person_properties":        map[string]any{"connector": "delighted", "fixture": true},
			"notes":                    []any{},
			"tags":                     []any{"fixture"},
			"nps":                      int64(42),
			"promoter_count":           int64(10),
			"promoter_percent":         "50.0",
			"passive_count":            int64(6),
			"passive_percent":          "30.0",
			"detractor_count":          int64(4),
			"detractor_percent":        "20.0",
			"response_count":           int64(20),
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

// requester builds a connsdk.Requester wired with HTTP Basic auth (API key as
// username, blank password) and the resolved base URL. The secret only ever
// flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := delightedBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := delightedSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("delighted connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, ""),
		UserAgent: delightedUserAgent,
	}, nil
}

// incrementalLowerBound returns the unix-seconds lower bound for the since
// filter, derived from the incremental cursor (if any) or else the since config.
// An empty result means no lower bound (full sync). The Delighted since param is
// a unix timestamp.
func incrementalLowerBound(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return normalizeSince(cursor)
	}
	since := strings.TrimSpace(req.Config.Config["since"])
	if since == "" {
		return "", nil
	}
	return normalizeSince(since)
}

// normalizeSince accepts either a unix-seconds string or an RFC3339 timestamp and
// returns unix seconds as a string.
func normalizeSince(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return value, nil
	}
	// Accept both RFC3339 and the "2006-01-02 15:04:05" form Delighted documents.
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
		if t, err := time.Parse(layout, value); err == nil {
			return strconv.FormatInt(t.Unix(), 10), nil
		}
	}
	return "", fmt.Errorf("delighted config since must be a unix timestamp or RFC3339 datetime, got %q", value)
}

func delightedSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// delightedBaseURL resolves and validates the base URL. The default is
// api.delighted.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func delightedBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return delightedDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("delighted config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("delighted config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("delighted config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func delightedPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return delightedDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("delighted config page_size must be an integer: %w", err)
	}
	if value < 1 || value > delightedMaxPageSize {
		return 0, fmt.Errorf("delighted config page_size must be between 1 and %d", delightedMaxPageSize)
	}
	return value, nil
}

func delightedMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("delighted config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("delighted config max_pages must be 0 for unlimited or a positive integer")
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
