// Package kissmetrics implements the native pm Kissmetrics connector. It is a
// declarative-HTTP per-system connector built on the stripe template: a thin
// package that composes the connsdk toolkit (Requester + Basic auth + RecordsAt
// extraction + offset pagination) with Kissmetrics-specific stream definitions
// and endpoints.
//
// Kissmetrics exposes a read-only query API. The top-level `products` collection
// lists the products visible to the authenticated user; reports, events, and
// properties are nested under a product partition
// (products/{product_id}/<resource>) and therefore require a product_id config.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package kissmetrics

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
	kissmetricsDefaultBaseURL  = "https://query.kissmetrics.io/v3"
	kissmetricsDefaultPageSize = 50
	kissmetricsMaxPageSize     = 200
	kissmetricsUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("kissmetrics", New)
}

// New returns the Kissmetrics connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Kissmetrics connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "kissmetrics" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "kissmetrics",
		DisplayName:     "Kissmetrics",
		IntegrationType: "api",
		Description:     "Reads Kissmetrics products, reports, events, and properties through the Kissmetrics query API using HTTP Basic authentication.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Kissmetrics.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := kissmetricsBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(kissmetricsUsername(cfg)) == "" {
		return errors.New("kissmetrics connector requires config username")
	}
	if strings.TrimSpace(kissmetricsPassword(cfg)) == "" {
		return errors.New("kissmetrics connector requires secret password")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the products list confirms auth and connectivity without
	// mutating anything.
	q := url.Values{"limit": []string{"1"}, "offset": []string{"0"}}
	if err := r.DoJSON(ctx, http.MethodGet, "products", q, nil, nil); err != nil {
		return fmt.Errorf("check kissmetrics: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: kissmetricsStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "products"
	}
	endpoint, ok := kissmetricsStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("kissmetrics stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	path, err := streamPath(endpoint, req.Config)
	if err != nil {
		return err
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := kissmetricsPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := kissmetricsMaxPages(req.Config)
	if err != nil {
		return err
	}

	paginator := &connsdk.OffsetPaginator{
		LimitParam:  "limit",
		OffsetParam: "offset",
		PageSize:    pageSize,
	}
	return connsdk.Harvest(ctx, r, http.MethodGet, path, url.Values{}, paginator, "data", maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

// Write is unsupported: Kissmetrics is a read-only query source. It satisfies the
// connectors.Connector interface while signalling no reverse-ETL capability.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// streamPath resolves the request path for a stream. Top-level streams read
// their resource directly; nested streams are scoped to a product partition and
// require a product_id config.
func streamPath(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !endpoint.nested {
		return endpoint.resource, nil
	}
	productID := strings.TrimSpace(cfg.Config["product_id"])
	if productID == "" {
		return "", fmt.Errorf("kissmetrics stream %q requires config product_id", endpoint.resource)
	}
	return "products/" + url.PathEscape(productID) + "/" + endpoint.resource, nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise kissmetrics credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"product_id":   "prod_fixture_1",
			"name":         fmt.Sprintf("%s fixture %d", strings.TrimSuffix(stream, "s"), i),
			"display_name": fmt.Sprintf("Fixture %d", i),
			"type":         "fixture",
			"created_at":   "2026-01-01T00:00:00Z",
			"updated_at":   "2026-01-01T00:00:00Z",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Basic auth and the resolved
// base URL. The password only ever flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := kissmetricsBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	username := kissmetricsUsername(cfg)
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("kissmetrics connector requires config username")
	}
	password := kissmetricsPassword(cfg)
	if strings.TrimSpace(password) == "" {
		return nil, errors.New("kissmetrics connector requires secret password")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, password),
		UserAgent: kissmetricsUserAgent,
	}, nil
}

func kissmetricsUsername(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["username"]
}

func kissmetricsPassword(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// kissmetricsBaseURL resolves and validates the base URL. The default is
// query.kissmetrics.io; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func kissmetricsBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return kissmetricsDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("kissmetrics config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("kissmetrics config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("kissmetrics config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func kissmetricsPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return kissmetricsDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("kissmetrics config page_size must be an integer: %w", err)
	}
	if value < 1 || value > kissmetricsMaxPageSize {
		return 0, fmt.Errorf("kissmetrics config page_size must be between 1 and %d", kissmetricsMaxPageSize)
	}
	return value, nil
}

func kissmetricsMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("kissmetrics config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("kissmetrics config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
