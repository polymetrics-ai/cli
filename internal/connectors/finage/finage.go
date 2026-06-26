// Package finage implements the native pm Finage connector. It follows the
// declarative-HTTP per-system connector shape established by the stripe package:
// a thin package that composes the connsdk toolkit (Requester + apikey query
// auth + RecordsAt extraction) with Finage-specific stream definitions and
// endpoints.
//
// Finage (https://finage.co.uk/docs/api) is a real-time market data API. This
// connector reads a core set of US market-information streams (most active
// stocks, top gainers/losers, sector performance, delisted companies) plus
// per-symbol market news. The API authenticates with an `apikey` query
// parameter and returns full JSON arrays (no offset/page/cursor pagination), so
// the read path performs a single GET per resource — and one GET per configured
// symbol for the symbol-partitioned news stream.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package finage

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
	finageDefaultBaseURL = "https://api.finage.co.uk"
	finageUserAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("finage", New)
}

// New returns the Finage connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Finage connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "finage" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "finage",
		DisplayName:     "Finage",
		IntegrationType: "api",
		Description:     "Reads Finage US market data: most active stocks, top gainers and losers, sector performance, delisted companies, and per-symbol market news via the Finage REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

// Check verifies the connector is configured well enough to talk to Finage. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := finageBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(finageSecret(cfg)) == "" {
		return errors.New("finage connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of sector-performance confirms auth and connectivity
	// without mutating anything; it needs no symbol partition.
	endpoint := finageStreamEndpoints["sector_performance"]
	if err := r.DoJSON(ctx, http.MethodGet, endpoint.resource, staticQuery(endpoint), nil, nil); err != nil {
		return fmt.Errorf("check finage: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: finageStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "most_active_us_stocks"
	}
	endpoint, ok := finageStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("finage stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	if endpoint.symbolPath {
		symbols := configSymbols(req.Config)
		if len(symbols) == 0 {
			return errors.New("finage stream " + stream + " requires config symbols")
		}
		for _, symbol := range symbols {
			if err := ctx.Err(); err != nil {
				return err
			}
			path := fmt.Sprintf(endpoint.resource, url.PathEscape(symbol))
			if err := c.fetchPage(ctx, r, path, endpoint, symbol, emit); err != nil {
				return err
			}
		}
		return nil
	}

	return c.fetchPage(ctx, r, endpoint.resource, endpoint, "", emit)
}

// fetchPage performs a single GET against path, extracts records at the
// endpoint's recordsPath, and emits each mapped record. When symbol is non-empty
// (news partition) it is stamped onto records missing one so per-symbol news is
// attributable.
func (c Connector) fetchPage(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, symbol string, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, path, staticQuery(endpoint), nil)
	if err != nil {
		return fmt.Errorf("read finage %s: %w", path, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode finage %s: %w", path, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if symbol != "" {
			if _, ok := item["symbol"]; !ok {
				item["symbol"] = symbol
			}
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise finage credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"symbol":            fmt.Sprintf("FIX%d", i),
			"company_name":      fmt.Sprintf("Fixture Co %d", i),
			"change":            float64(i),
			"change_percentage": fmt.Sprintf("%d.0%%", i),
			"price":             fmt.Sprintf("%d.00", 100*i),
			"sector":            fmt.Sprintf("Sector %d", i),
			"exchange":          "NASDAQ",
			"ipo_date":          "2020-01-01",
			"delisted_date":     "2026-01-01",
			"title":             fmt.Sprintf("Fixture headline %d", i),
			"url":               fmt.Sprintf("https://example.com/news/%d", i),
			"source":            "fixture",
			"description":       fmt.Sprintf("Fixture description %d", i),
			"date":              "2026-01-01",
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
	base, err := finageBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := finageSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("finage connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("apikey", secret),
		UserAgent: finageUserAgent,
	}, nil
}

// staticQuery returns a copy of the endpoint's static query parameters as
// url.Values, or nil when the endpoint declares none.
func staticQuery(endpoint streamEndpoint) url.Values {
	if len(endpoint.extraParams) == 0 {
		return nil
	}
	q := url.Values{}
	for k, v := range endpoint.extraParams {
		q.Set(k, v)
	}
	return q
}

// configSymbols parses the comma- or whitespace-separated symbols config into a
// trimmed, non-empty list.
func configSymbols(cfg connectors.RuntimeConfig) []string {
	raw := strings.TrimSpace(cfg.Config["symbols"])
	if raw == "" {
		return nil
	}
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	return out
}

func finageSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// finageBaseURL resolves and validates the base URL. The default is
// api.finage.co.uk; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func finageBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return finageDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("finage config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("finage config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("finage config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Finage is a read-only market-data source.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
