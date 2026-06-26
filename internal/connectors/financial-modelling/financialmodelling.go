// Package financialmodelling implements the native pm Financial Modeling Prep
// (FMP) connector. It follows the stripe declarative-HTTP template: a thin
// package that composes the connsdk toolkit (Requester + apikey query auth +
// top-level array extraction) with FMP-specific stream definitions and
// endpoints.
//
// The bare system name (registry key, directory) is "financial-modelling"; the
// Go package identifier cannot contain a hyphen, so it is "financialmodelling".
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// FMP authenticates with an apikey query parameter and most endpoints return a
// top-level JSON array. The list endpoints are not paginated; the stock screener
// honours limit/offset, which the read loop drives generically.
package financialmodelling

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
	fmDefaultBaseURL  = "https://financialmodelingprep.com/api/v3"
	fmDefaultPageSize = 1000
	fmMaxPageSize     = 10000
	fmUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("financial-modelling", New)
}

// New returns the Financial Modeling Prep connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Financial Modeling Prep connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "financial-modelling" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "financial-modelling",
		DisplayName:     "Financial Modelling",
		IntegrationType: "api",
		Description:     "Reads stock and ETF symbol lists, the stock screener, and delisted companies from the Financial Modeling Prep REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to FMP. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := fmBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(fmSecret(cfg)) == "" {
		return errors.New("financial-modelling connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the stock screener confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "stock-screener", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check financial-modelling: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: fmStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "stocks"
	}
	endpoint, ok := fmStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("financial-modelling stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := fmPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := fmMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, req.Config, pageSize, maxPages, emit)
}

// harvest reads an FMP endpoint. Non-paginated list endpoints return their whole
// array in one response, so they make a single request. The stock screener
// honours limit/offset: a full page (len == pageSize) implies another page, a
// short page ends the loop. The loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt with a root field path.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, cfg connectors.RuntimeConfig, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := fmScreenerParams(endpoint, cfg)

	if !endpoint.paginated {
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, base, nil)
		if err != nil {
			return fmt.Errorf("read financial-modelling %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode financial-modelling %s: %w", endpoint.resource, err)
		}
		return emitAll(ctx, records, endpoint, emit)
	}

	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read financial-modelling %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode financial-modelling %s page: %w", endpoint.resource, err)
		}
		if err := emitAll(ctx, records, endpoint, emit); err != nil {
			return err
		}
		// A short (or empty) page means there is no further page.
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

func emitAll(ctx context.Context, records []connsdk.Record, endpoint streamEndpoint, emit func(connectors.Record) error) error {
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

// fmScreenerParams builds the optional screener filter params from config. They
// only apply to the stock-screener endpoint.
func fmScreenerParams(endpoint streamEndpoint, cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	if endpoint.resource != "stock-screener" {
		return q
	}
	if v := strings.TrimSpace(cfg.Config["exchange"]); v != "" {
		q.Set("exchange", v)
	}
	if v := strings.TrimSpace(cfg.Config["marketcapmorethan"]); v != "" {
		q.Set("marketCapMoreThan", v)
	}
	if v := strings.TrimSpace(cfg.Config["marketcaplowerthan"]); v != "" {
		q.Set("marketCapLowerThan", v)
	}
	return q
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	endpoint := fmStreamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"symbol":            fmt.Sprintf("FIX%d", i),
			"name":              fmt.Sprintf("Fixture Company %d", i),
			"companyName":       fmt.Sprintf("Fixture Company %d", i),
			"price":             100.0 + float64(i),
			"exchange":          "NASDAQ Global Select",
			"exchangeShortName": "NASDAQ",
			"type":              "stock",
			"marketCap":         int64(1000000000 * i),
			"sector":            "Technology",
			"industry":          "Software",
			"beta":              1.1,
			"volume":            int64(1000000 * i),
			"country":           "US",
			"isEtf":             false,
			"isActivelyTrading": true,
			"ipoDate":           "2001-01-01",
			"delistedDate":      "2020-01-01",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with apikey query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := fmBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := fmSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("financial-modelling connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("apikey", secret),
		UserAgent: fmUserAgent,
	}, nil
}

func fmSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// fmBaseURL resolves and validates the base URL. The default is
// financialmodelingprep.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func fmBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return fmDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("financial-modelling config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("financial-modelling config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("financial-modelling config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fmPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return fmDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("financial-modelling config page_size must be an integer: %w", err)
	}
	if value < 1 || value > fmMaxPageSize {
		return 0, fmt.Errorf("financial-modelling config page_size must be between 1 and %d", fmMaxPageSize)
	}
	return value, nil
}

func fmMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("financial-modelling config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("financial-modelling config max_pages must be 0 for unlimited or a positive integer")
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

// Write is unsupported: FMP is a read-only market-data API.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
