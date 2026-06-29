// Package dingconnect implements the native pm DingConnect connector. It is a
// declarative-HTTP per-system connector following the stripe template: a thin
// package that composes the connsdk toolkit (Requester + api_key header auth +
// RecordsAt extraction) with DingConnect-specific stream definitions, endpoints,
// and Skip-offset pagination.
//
// The DingConnect API authenticates with an `api_key` request header (no
// prefix), exposes read-only reference/catalog endpoints under
// https://api.dingconnect.com/api/V1 that each return a JSON envelope
// {"Items":[...]}, and paginates by injecting a numeric `Skip` offset (page size
// 100) on every request. The DingConnect source is read-only (full-refresh); it
// exposes no reverse-ETL writes, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package dingconnect

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
	dingDefaultBaseURL  = "https://api.dingconnect.com"
	dingAPIPrefix       = "api/V1"
	dingDefaultPageSize = 100
	dingMaxPageSize     = 100
	dingUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("ding-connect", New)
}

// New returns the DingConnect connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm DingConnect connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "ding-connect" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "ding-connect",
		DisplayName:     "Ding Connect",
		IntegrationType: "api",
		Description:     "Reads DingConnect reference catalogs (countries, currencies, regions, providers, and products) through the DingConnect REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to DingConnect.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := dingBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(dingSecret(cfg)) == "" {
		return errors.New("ding-connect connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of GetCurrencies confirms auth and connectivity without
	// mutating anything (currencies is a small, stable list).
	query := url.Values{"Skip": []string{"0"}}
	if err := r.DoJSON(ctx, http.MethodGet, dingAPIPrefix+"/GetCurrencies", query, nil, nil); err != nil {
		return fmt.Errorf("check ding-connect: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: dingStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "countries"
	}
	endpoint, ok := dingStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("ding-connect stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := dingPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := dingMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives DingConnect's Skip-offset pagination. Each list endpoint returns
// {"Items":[...]}; the Skip query param is injected on every request (starting at
// 0) and advanced by the page size until a short page (fewer than pageSize items)
// is returned. There is no body token paginator for this exact shape, so the loop
// lives here, built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := dingAPIPrefix + "/" + endpoint.resource
	skip := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("Skip", strconv.Itoa(skip))
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read ding-connect %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "Items")
		if err != nil {
			return fmt.Errorf("decode ding-connect %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (or an empty one) signals the end of the resource.
		if len(records) < pageSize {
			return nil
		}
		skip += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise ding-connect credential-free (mirrors the
// stripe fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		idx := strconv.Itoa(i)
		item := map[string]any{
			"CountryIso":          "C" + idx,
			"CountryName":         "Fixture Country " + idx,
			"CurrencyIso":         "X" + idx + "X",
			"CurrencyName":        "Fixture Currency " + idx,
			"RegionCode":          "R" + idx,
			"RegionName":          "Fixture Region " + idx,
			"ProviderCode":        "PROV" + idx,
			"Name":                "Fixture Provider " + idx,
			"SkuCode":             "SKU" + idx,
			"DefaultDisplayText":  "Fixture Product " + idx,
			"RedemptionMechanism": "Immediate",
			"ProcessingMode":      "Immediate",
			"connector":           "ding-connect",
			"fixture":             true,
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with api_key header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := dingBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := dingSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("ding-connect connector requires secret api_key")
	}
	headers := map[string]string{}
	if corr := strings.TrimSpace(cfg.Config["x_correlation_id"]); corr != "" {
		headers["X-Correlation-Id"] = corr
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           connsdk.APIKeyHeader("api_key", secret, ""),
		UserAgent:      dingUserAgent,
		DefaultHeaders: headers,
	}, nil
}

func dingSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// dingBaseURL resolves and validates the base URL. The default is
// api.dingconnect.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func dingBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return dingDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("ding-connect config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("ding-connect config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("ding-connect config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func dingPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return dingDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("ding-connect config page_size must be an integer: %w", err)
	}
	if value < 1 || value > dingMaxPageSize {
		return 0, fmt.Errorf("ding-connect config page_size must be between 1 and %d", dingMaxPageSize)
	}
	return value, nil
}

func dingMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("ding-connect config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("ding-connect config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. The DingConnect source is
// read-only; reverse-ETL writes are intentionally unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// dingUUID builds a stable synthetic primary key from the given item fields,
// mirroring the synthetic "uuid" the upstream source assigns to these keyless
// reference resources. Empty/missing components are skipped.
func dingUUID(item map[string]any, keys ...string) string {
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		if v := stringField(item, k); v != "" {
			parts = append(parts, v)
		}
	}
	return strings.Join(parts, ":")
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
