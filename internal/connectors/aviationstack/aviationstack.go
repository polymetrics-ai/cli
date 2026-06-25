// Package aviationstack implements the native pm aviationstack connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit: a
// connsdk.Requester wired with access_key query auth, limit/offset pagination
// over aviationstack's {pagination, data[]} envelope, and aviationstack-specific
// stream definitions and record mappers.
//
// aviationstack exposes read-only reference and flight data, so this connector is
// read-only (no write actions). It mirrors the stripe template's shape: fixture
// mode for credential-free conformance, base_url override with SSRF validation,
// and self-registration via RegisterFactory in init().
package aviationstack

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
	defaultBaseURL  = "https://api.aviationstack.com/v1"
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("aviationstack", New)
}

// New returns the aviationstack connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm aviationstack connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "aviationstack" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "aviationstack",
		DisplayName:     "Aviationstack",
		IntegrationType: "api",
		Description:     "Reads aviationstack flights and aviation reference data (airlines, airports, airplanes, countries) through the aviationstack REST API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to
// aviationstack. In fixture mode it short-circuits without a network call.
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
	if strings.TrimSpace(accessKey(cfg)) == "" {
		return errors.New("aviationstack connector requires secret access_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the countries reference list confirms auth and
	// connectivity without mutating anything (countries is a small static set).
	q := url.Values{"limit": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "countries", q, nil, nil); err != nil {
		return fmt.Errorf("check aviationstack: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: aviationStreams()}, nil
}

// Write is unsupported: aviationstack is a read-only data source with no
// reverse-ETL write endpoints. It satisfies the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "airlines"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("aviationstack stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	// Validate base_url even before building the requester so an invalid override
	// (SSRF guard) surfaces a clear error.
	if _, err := baseURL(req.Config); err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := pageSizeOf(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPagesOf(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives aviationstack's limit/offset pagination. List responses are
// shaped {pagination:{limit,offset,count,total}, data:[...]}; the loop advances
// offset by the page size until a short (or empty) page is returned, or until the
// reported total is reached. There is no envelope-aware paginator in connsdk for
// this exact shape, so the loop lives here on connsdk.Requester + RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read aviationstack %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode aviationstack %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Stop on a short page: fewer records than requested means the last page.
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
		// Defensively stop once we have reached the reported total, if present.
		if total, ok := paginationTotal(resp.Body); ok && offset >= total {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                  fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"flight_date":         "2026-01-01",
			"flight_status":       "scheduled",
			"airline_name":        fmt.Sprintf("Fixture Air %d", i),
			"iata_code":           fmt.Sprintf("F%d", i),
			"icao_code":           fmt.Sprintf("FX%d", i),
			"country_name":        "Fixtureland",
			"country_iso2":        "FX",
			"country_iso3":        "FXL",
			"airport_name":        fmt.Sprintf("Fixture Airport %d", i),
			"latitude":            "0.0",
			"longitude":           "0.0",
			"timezone":            "UTC",
			"registration_number": fmt.Sprintf("FX-%d", i),
			"model_name":          "Fixture 100",
			"plane_status":        "active",
			"capital":             "Fixture City",
			"continent":           "FX",
			"population":          "1000",
			"airline": map[string]any{
				"name": fmt.Sprintf("Fixture Air %d", i),
				"iata": fmt.Sprintf("F%d", i),
			},
			"flight": map[string]any{
				"iata":   fmt.Sprintf("F%d100", i),
				"icao":   fmt.Sprintf("FX%d100", i),
				"number": "100",
			},
			"departure": map[string]any{
				"airport":   "Fixture Departure",
				"iata":      "FXD",
				"scheduled": "2026-01-01T08:00:00+00:00",
			},
			"arrival": map[string]any{
				"airport":   "Fixture Arrival",
				"iata":      "FXA",
				"scheduled": "2026-01-01T10:00:00+00:00",
			},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with access_key query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := accessKey(cfg)
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("aviationstack connector requires secret access_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("access_key", key),
		UserAgent: userAgent,
	}, nil
}

func accessKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["access_key"]
}

// baseURL resolves and validates the base URL. The default is
// api.aviationstack.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("aviationstack config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("aviationstack config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("aviationstack config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSizeOf(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("aviationstack config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("aviationstack config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPagesOf(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("aviationstack config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("aviationstack config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// paginationTotal reads pagination.total from a list response body. It returns
// (0, false) when the field is absent or not an integer.
func paginationTotal(body []byte) (int, bool) {
	raw, err := connsdk.StringAt(body, "pagination.total")
	if err != nil || strings.TrimSpace(raw) == "" {
		return 0, false
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return value, true
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
