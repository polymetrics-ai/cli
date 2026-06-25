// Package bluetally implements the native pm BlueTally connector. BlueTally is an
// IT asset management platform; this connector reads its assets, employees,
// licenses, maintenances, and accessories through the BlueTally REST API.
//
// It follows the declarative-HTTP template established by the stripe connector: a
// thin package composing the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction + offset pagination) with BlueTally-specific stream
// definitions and endpoints. It self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// The BlueTally API base is https://app.bluetallyapp.com, list endpoints live
// under /api/v1/<resource>, authentication is an Authorization: Bearer <api_key>
// header, list responses are top-level JSON arrays, and pagination is
// offset-based via limit/offset query parameters. This connector is read-only:
// the upstream source supports full-refresh reads only.
package bluetally

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
	bluetallyDefaultBaseURL  = "https://app.bluetallyapp.com"
	bluetallyAPIPrefix       = "api/v1"
	bluetallyDefaultPageSize = 50
	bluetallyMaxPageSize     = 100
	bluetallyUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("bluetally", New)
}

// New returns the BlueTally connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm BlueTally connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "bluetally" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "bluetally",
		DisplayName:     "BlueTally",
		IntegrationType: "api",
		Description:     "Reads BlueTally IT asset management data (assets, employees, licenses, maintenances, accessories) through the BlueTally REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to BlueTally. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := bluetallyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(bluetallySecret(cfg)) == "" {
		return errors.New("bluetally connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the assets list confirms auth and connectivity without
	// mutating anything.
	query := url.Values{"limit": []string{"1"}}
	if _, err := r.Do(ctx, http.MethodGet, resourcePath("assets"), query, nil); err != nil {
		return fmt.Errorf("check bluetally: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: bluetallyStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a BlueTally stream starts
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
		stream = "assets"
	}
	endpoint, ok := bluetallyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("bluetally stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := bluetallyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := bluetallyMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives BlueTally's offset pagination. List endpoints return a
// top-level JSON array; the next page is requested with offset += limit until a
// short (under-full) page is returned. There is no body token to follow, so the
// loop lives here, built on connsdk.Requester + connsdk.RecordsAt with a
// root-path ("") record selector.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := resourcePath(endpoint.resource)
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(page*pageSize))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read bluetally %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode bluetally %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page means there is no further data.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise bluetally credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                       int64(i),
			"asset_id":                 fmt.Sprintf("BT-%04d", i),
			"asset_name":               fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"asset_serial":             fmt.Sprintf("SN-FIXTURE-%d", i),
			"name":                     fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"email":                    fmt.Sprintf("fixture+%d@example.com", i),
			"title":                    "Engineer",
			"status_id":                int64(1),
			"category_id":              int64(2),
			"category_name":            "Laptops",
			"product_id":               int64(3),
			"product_name":             "Fixture Model",
			"location_id":              int64(4),
			"department_id":            int64(5),
			"supplier_id":              int64(6),
			"manufacturer_id":          int64(7),
			"manager_id":               int64(8),
			"archived":                 false,
			"number_of_assets":         int64(i),
			"number_of_accessories":    int64(0),
			"number_of_consumables":    int64(0),
			"number_of_licenses":       int64(0),
			"licensed_to_name":         fmt.Sprintf("Team %d", i),
			"licensed_to_email":        fmt.Sprintf("team+%d@example.com", i),
			"license_type":             "perpetual",
			"order_number":             fmt.Sprintf("ORD-%d", i),
			"purchase_date":            "2026-01-01",
			"expiration_date":          "2027-01-01",
			"termination_date":         "",
			"warranty_expiration_date": "2027-01-01",
			"purchase_cost":            1000 * i,
			"unit_cost":                10 * i,
			"cost":                     50 * i,
			"currency":                 "USD",
			"minimum_seats":            int64(1),
			"number_of_seats":          int64(10),
			"available":                int64(9),
			"quantity":                 int64(10),
			"model_number":             fmt.Sprintf("MN-%d", i),
			"type":                     "repair",
			"asset_id_ref":             int64(i),
			"start_date":               "2026-01-01",
			"end_date":                 "2026-01-05",
			"notes":                    "fixture record",
			"created_at":               "2026-01-01T00:00:00Z",
			"updated_at":               fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
		}
		// maintenances reference an asset by integer asset_id, distinct from the
		// asset stream's string asset_id; set it appropriately for this stream.
		if stream == "maintenances" {
			item["asset_id"] = int64(i)
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

// Write satisfies the connectors.Connector interface. BlueTally is read-only in
// this connector (the upstream source supports full-refresh reads only), so
// writes are unsupported.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := bluetallyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := bluetallySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("bluetally connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: bluetallyUserAgent,
	}, nil
}

// resourcePath joins the API prefix with a resource segment, e.g.
// "assets" -> "api/v1/assets".
func resourcePath(resource string) string {
	return bluetallyAPIPrefix + "/" + resource
}

func bluetallySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// bluetallyBaseURL resolves and validates the base URL. The default is
// app.bluetallyapp.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func bluetallyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return bluetallyDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("bluetally config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("bluetally config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("bluetally config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func bluetallyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return bluetallyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("bluetally config page_size must be an integer: %w", err)
	}
	if value < 1 || value > bluetallyMaxPageSize {
		return 0, fmt.Errorf("bluetally config page_size must be between 1 and %d", bluetallyMaxPageSize)
	}
	return value, nil
}

func bluetallyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("bluetally config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("bluetally config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
