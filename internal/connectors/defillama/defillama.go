// Package defillama implements the native pm DefiLlama connector. DefiLlama is a
// public, unauthenticated DeFi analytics API split across a few hosts
// (api.llama.fi, stablecoins.llama.fi). This package follows the stripe
// declarative-HTTP template: a thin package composing the connsdk Requester +
// RecordsAt extraction with DefiLlama-specific stream definitions and endpoints.
//
// The API requires no credentials and supports full-refresh reads only, so the
// connector is read-only (Capabilities.Write = false). Like stripe it
// self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package defillama

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defillamaMainBaseURL        = "https://api.llama.fi"
	defillamaStablecoinsBaseURL = "https://stablecoins.llama.fi"
	defillamaDefaultPageSize    = 1000
	defillamaMaxPageSize        = 5000
	defillamaUserAgent          = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("defillama", New)
}

// New returns the DefiLlama connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm DefiLlama connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "defillama" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "defillama",
		DisplayName:     "DefiLlama",
		IntegrationType: "api",
		Description:     "Reads DefiLlama DeFi analytics: protocols, chains, stablecoins, DEX volumes, and fees/revenue from the public DefiLlama REST API. Read-only; no authentication required.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector can reach DefiLlama. The API needs no credentials,
// so in fixture mode this short-circuits, and otherwise it performs a bounded
// read of the protocols endpoint to confirm connectivity.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	endpoint := defillamaStreamEndpoints["protocols"]
	r, err := c.requester(cfg, endpoint.host)
	if err != nil {
		return err
	}
	query := url.Values{"limit": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, endpoint.resource, query, nil, nil); err != nil {
		return fmt.Errorf("check defillama: %w", err)
	}
	return nil
}

// Write is unsupported: DefiLlama is a read-only public analytics API. The method
// exists only to satisfy the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: defillamaStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "protocols"
	}
	endpoint, ok := defillamaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("defillama stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config, endpoint.host)
	if err != nil {
		return err
	}
	pageSize, err := defillamaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := defillamaMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest reads an endpoint. DefiLlama returns full lists with no server-side
// pagination, but for the large list endpoints (protocols, chains) the connector
// pages the response client-side with limit/offset to keep payloads bounded and
// to drive a real multi-page loop. Endpoints that are not paginated are read in a
// single request. Records are extracted at endpoint.recordsPath via
// connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	for k, v := range endpoint.query {
		base.Set(k, v)
	}

	if !endpoint.paginated {
		records, err := c.fetchPage(ctx, r, endpoint, base)
		if err != nil {
			return err
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
		records, err := c.fetchPage(ctx, r, endpoint, query)
		if err != nil {
			return err
		}
		if err := emitAll(ctx, records, endpoint, emit); err != nil {
			return err
		}
		// A short page means we have reached the end of the list.
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

func (c Connector) fetchPage(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, query url.Values) ([]map[string]any, error) {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
	if err != nil {
		return nil, fmt.Errorf("read defillama %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return nil, fmt.Errorf("decode defillama %s: %w", endpoint.resource, err)
	}
	return records, nil
}

func emitAll(ctx context.Context, records []map[string]any, endpoint streamEndpoint, emit func(connectors.Record) error) error {
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
// conformance harness can exercise defillama credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           fmt.Sprintf("%s_fixture_%d", stream, i),
			"defillamaId":  fmt.Sprintf("%d", i),
			"name":         fmt.Sprintf("Fixture %s %d", stream, i),
			"displayName":  fmt.Sprintf("Fixture %s %d", stream, i),
			"slug":         fmt.Sprintf("fixture-%d", i),
			"symbol":       "FIX",
			"category":     "Dexes",
			"chain":        "Ethereum",
			"chains":       []any{"Ethereum"},
			"tvl":          float64(1000 * i),
			"mcap":         float64(2000 * i),
			"change_1d":    float64(i),
			"change_7d":    float64(i * 2),
			"url":          "https://example.com",
			"gecko_id":     fmt.Sprintf("fixture-%d", i),
			"tokenSymbol":  "FIX",
			"cmcId":        fmt.Sprintf("%d", i),
			"chainId":      float64(i),
			"pegType":      "peggedUSD",
			"pegMechanism": "fiat-backed",
			"price":        1.0,
			"circulating":  map[string]any{"peggedUSD": float64(1000 * i)},
			"total24h":     float64(100 * i),
			"total7d":      float64(700 * i),
			"total30d":     float64(3000 * i),
			"totalAllTime": float64(99999 * i),
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester for the given host. DefiLlama is
// unauthenticated, so no Auth is set. The base URL is resolved from the host kind
// plus any config override and validated to bound SSRF risk.
func (c Connector) requester(cfg connectors.RuntimeConfig, host hostKind) (*connsdk.Requester, error) {
	base, err := defillamaBaseURL(cfg, host)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		UserAgent: defillamaUserAgent,
	}, nil
}

// defillamaBaseURL resolves and validates the base URL for a host. Defaults are
// the public DefiLlama hosts. A generic "base_url" override applies to the main
// host; a "stablecoins_base_url" override applies to the stablecoins host. Any
// override must be an absolute http(s) URL with a host to bound SSRF risk.
func defillamaBaseURL(cfg connectors.RuntimeConfig, host hostKind) (string, error) {
	defaultBase := defillamaMainBaseURL
	overrideKey := "base_url"
	if host == hostStablecoins {
		defaultBase = defillamaStablecoinsBaseURL
		overrideKey = "stablecoins_base_url"
	}
	override := ""
	if cfg.Config != nil {
		override = strings.TrimSpace(cfg.Config[overrideKey])
		// A single base_url override (test servers) applies to all hosts when a
		// host-specific override is not set.
		if override == "" && host == hostStablecoins {
			override = strings.TrimSpace(cfg.Config["base_url"])
		}
	}
	if override == "" {
		return defaultBase, nil
	}
	parsed, err := url.Parse(override)
	if err != nil {
		return "", fmt.Errorf("defillama config %s is invalid: %w", overrideKey, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("defillama config %s must use http or https, got %q", overrideKey, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("defillama config %s must include a host", overrideKey)
	}
	return strings.TrimRight(override, "/"), nil
}

func defillamaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := ""
	if cfg.Config != nil {
		raw = strings.TrimSpace(cfg.Config["page_size"])
	}
	if raw == "" {
		return defillamaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("defillama config page_size must be an integer: %w", err)
	}
	if value < 1 || value > defillamaMaxPageSize {
		return 0, fmt.Errorf("defillama config page_size must be between 1 and %d", defillamaMaxPageSize)
	}
	return value, nil
}

func defillamaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := ""
	if cfg.Config != nil {
		raw = strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	}
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("defillama config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, fmt.Errorf("defillama config max_pages must be 0 for unlimited or a positive integer")
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
