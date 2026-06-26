// Package openaq implements the native pm OpenAQ connector. It is a declarative-
// HTTP per-system connector built on the same shape as the stripe reference: a
// thin package composing the connsdk toolkit (Requester + X-API-Key auth +
// RecordsAt extraction) with OpenAQ-specific stream definitions and endpoints.
//
// OpenAQ v3 (https://api.openaq.org/v3) is a read-only air-quality reference API
// authenticated with an X-API-Key header. Its list endpoints share a common
// envelope: {"meta":{"page":N,"limit":L,"found":F},"results":[...]} with
// page/limit pagination. This connector is therefore read-only (Write=false).
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package openaq

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
	openaqDefaultBaseURL  = "https://api.openaq.org/v3"
	openaqDefaultPageSize = 100
	openaqMaxPageSize     = 1000
	openaqUserAgent       = "polymetrics-go-cli"
	openaqAPIKeyHeader    = "X-API-Key"
)

func init() {
	connectors.RegisterFactory("openaq", New)
}

// New returns the OpenAQ connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm OpenAQ connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "openaq" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "openaq",
		DisplayName:     "OpenAQ",
		IntegrationType: "api",
		Description:     "Reads OpenAQ air quality reference data (countries, parameters, locations, instruments, and manufacturers) from the OpenAQ v3 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to OpenAQ. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := openaqBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(openaqSecret(cfg)) == "" {
		return errors.New("openaq connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the countries list confirms auth and connectivity.
	if err := r.DoJSON(ctx, http.MethodGet, "countries", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check openaq: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: openaqStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "countries"
	}
	endpoint, ok := openaqStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("openaq stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := openaqPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := openaqMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, req.Config, emit)
}

// harvest drives OpenAQ's page/limit pagination. List responses are
// {meta:{found:F,...}, results:[...]}; pages are 1-indexed. The loop stops when a
// page returns fewer than pageSize results, when the accumulated count reaches
// meta.found, or when maxPages is hit. connsdk has no paginator for the
// found-total shape, so the loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	if filter := openaqCountryFilter(cfg); filter != "" {
		base.Set("countries_id", filter)
	}

	seen := 0
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read openaq %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode openaq %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		seen += len(records)

		// Stop on a short page (fewer than a full page worth of records).
		if len(records) < pageSize {
			return nil
		}
		// Stop once we have read everything the meta total advertises.
		if found := parseFound(resp.Body); found > 0 && seen >= found {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise openaq credential-free (mirrors the stripe
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            int64(i),
			"code":          fmt.Sprintf("C%d", i),
			"name":          fmt.Sprintf("%s fixture %d", stream, i),
			"units":         "µg/m³",
			"displayName":   fmt.Sprintf("Fixture %d", i),
			"description":   "fixture record",
			"locality":      fmt.Sprintf("Locality %d", i),
			"timezone":      "UTC",
			"isMobile":      false,
			"isMonitor":     true,
			"connector":     "openaq",
			"fixture":       true,
			"datetimeFirst": "2024-01-01T00:00:00Z",
			"datetimeLast":  "2026-01-01T00:00:00Z",
			"country":       map[string]any{"id": int64(i), "code": fmt.Sprintf("C%d", i), "name": fmt.Sprintf("Country %d", i)},
			"manufacturer":  map[string]any{"id": int64(i), "name": fmt.Sprintf("Maker %d", i)},
		}
		record := endpoint.mapRecord(item)
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with X-API-Key auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := openaqBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := openaqSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("openaq connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(openaqAPIKeyHeader, secret, ""),
		UserAgent: openaqUserAgent,
	}, nil
}

func openaqSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// openaqCountryFilter returns the optional comma-separated countries_id filter
// from config (the catalog's country_ids field), normalized for the query param.
func openaqCountryFilter(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	for _, key := range []string{"countries_id", "country_ids"} {
		if v := strings.TrimSpace(cfg.Config[key]); v != "" {
			return v
		}
	}
	return ""
}

// openaqBaseURL resolves and validates the base URL. The default is
// api.openaq.org/v3; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func openaqBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return openaqDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("openaq config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("openaq config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("openaq config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func openaqPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return openaqDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("openaq config page_size must be an integer: %w", err)
	}
	if value < 1 || value > openaqMaxPageSize {
		return 0, fmt.Errorf("openaq config page_size must be between 1 and %d", openaqMaxPageSize)
	}
	return value, nil
}

func openaqMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("openaq config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("openaq config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// parseFound reads meta.found from a list response, returning -1 when absent so
// the caller falls back to short-page detection.
func parseFound(body []byte) int {
	raw, err := connsdk.StringAt(body, "meta.found")
	if err != nil || strings.TrimSpace(raw) == "" {
		return -1
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return -1
	}
	return value
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

// Write is not supported: OpenAQ is a read-only reference API. It returns
// ErrUnsupportedOperation so the connector satisfies the Connector interface
// while advertising Write=false in Metadata.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
