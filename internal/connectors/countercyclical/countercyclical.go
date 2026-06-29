// Package countercyclical implements the native pm Countercyclical source
// connector. Countercyclical is a financial-intelligence platform for
// investment teams; this connector reads its investments, valuations, and memos
// streams over the REST API.
//
// It follows the declarative-HTTP template established by the stripe connector:
// a thin package that composes the connsdk toolkit (Requester + API-key query
// auth + root-array extraction) with Countercyclical-specific stream
// definitions and endpoints. The upstream API authenticates with an `apiKey`
// query parameter (upstream ApiKeyAuthenticator, inject_into request_parameter),
// returns each stream as a root-level JSON array, and exposes no pagination.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package countercyclical

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
	defaultBaseURL  = "https://api.countercyclical.io/v1"
	defaultPageSize = 100
	maxPageSize     = 1000
	userAgent       = "polymetrics-go-cli"
	// apiKeyParam is the query parameter the Countercyclical API uses to carry
	// the API key (ApiKeyAuthenticator with inject_into request_parameter).
	apiKeyParam = "apiKey"
)

func init() {
	connectors.RegisterFactory("countercyclical", New)
}

// New returns the Countercyclical connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Countercyclical source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "countercyclical" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "countercyclical",
		DisplayName:     "Countercyclical",
		IntegrationType: "api",
		Description:     "Reads Countercyclical investments, valuations, and research memos through the Countercyclical REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to
// Countercyclical. In fixture mode it short-circuits without a network call.
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
	if strings.TrimSpace(secret(cfg)) == "" {
		return errors.New("countercyclical connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the investments list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "investments", nil, nil, nil); err != nil {
		return fmt.Errorf("check countercyclical: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streamCatalog()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "investments"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("countercyclical stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
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
	return c.harvest(ctx, r, endpoint, size, pages, emit)
}

// Write is unsupported: Countercyclical is a read-only source. The method exists
// to satisfy the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest reads a stream. The Countercyclical API returns each stream as a
// root-level JSON array with no documented pagination, so the default path
// performs a single request. As a defensive measure against a server that does
// honor limit/offset, the loop continues while a full page (== pageSize) comes
// back; a short page (or a server that ignores the params and returns everything
// at once) terminates the loop. Unknown query params are ignored by such APIs,
// so this stays safe for the no-pagination case.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, size, pages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; pages == 0 || page < pages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(size))
		if offset > 0 {
			query.Set("offset", strconv.Itoa(offset))
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read countercyclical %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode countercyclical %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Stop when the page is short (or empty): either the API has no
		// pagination and returned everything, or this was the final page.
		if len(records) < size {
			return nil
		}
		offset += size
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free (mirrors
// stripe's fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := fixtureItem(stream, endpoint.resource, i)
		record := endpoint.mapRecord(item)
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// fixtureItem builds a deterministic raw object for a stream/index pair. It
// carries a superset of every stream's fields so each mapper has values to
// project.
func fixtureItem(stream, resource string, i int) map[string]any {
	const fixtureCreated = "2026-01-01T00:00:00Z"
	const fixtureUpdated = "2026-01-02T00:00:00Z"
	return map[string]any{
		"id":              fmt.Sprintf("%s_fixture_%d", resource, i),
		"name":            fmt.Sprintf("Fixture %s %d", stream, i),
		"editedName":      fmt.Sprintf("Fixture %s %d", stream, i),
		"title":           fmt.Sprintf("Fixture memo %d", i),
		"body":            "Fixture body.",
		"description":     "Fixture record for conformance.",
		"tickerSymbol":    "FIX",
		"exchange":        "NASDAQ",
		"country":         "US",
		"sector":          "Technology",
		"industry":        "Software",
		"marketType":      "public",
		"financingType":   "equity",
		"employees":       int64(100 * i),
		"website":         "https://example.com",
		"visibility":      "private",
		"status":          "draft",
		"delineation":     "annual",
		"discountRate":    0.1,
		"growthMetric":    "revenue",
		"growthRate":      0.2,
		"terminalRate":    0.03,
		"terminalPeriod":  "perpetuity",
		"startingQuarter": int64(1),
		"startingYear":    int64(2026),
		"endingQuarter":   int64(4),
		"endingYear":      int64(2031),
		"shareToken":      fmt.Sprintf("tok_%d", i),
		"documentType":    "memo",
		"emoji":           "📈",
		"backgroundColor": "#ffffff",
		"foregroundColor": "#000000",
		"views":           int64(10 * i),
		"isArchived":      false,
		"isFavorite":      true,
		"isLocked":        false,
		"archived":        false,
		"favorited":       true,
		"locked":          false,
		"publiclyVisible": false,
		"sourcesVisible":  true,
		"tocVisible":      true,
		"bannerVisible":   false,
		"createdAt":       fixtureCreated,
		"updatedAt":       fixtureUpdated,
	}
}

// requester builds a connsdk.Requester wired with API-key query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := secret(cfg)
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("countercyclical connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery(apiKeyParam, key),
		UserAgent: userAgent,
	}, nil
}

func secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// baseURL resolves and validates the base URL. The default is
// api.countercyclical.io; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("countercyclical config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("countercyclical config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("countercyclical config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("countercyclical config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("countercyclical config page_size must be between 1 and %d", maxPageSize)
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
		return 0, fmt.Errorf("countercyclical config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("countercyclical config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
