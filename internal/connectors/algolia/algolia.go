// Package algolia implements the native pm Algolia source connector. It follows
// the declarative-HTTP template established by the stripe package: a thin package
// that composes the connsdk toolkit (Requester + Algolia api-key/app-id header
// auth + RecordsAt extraction) with Algolia-specific stream definitions and
// endpoints.
//
// Algolia's Search REST API is a configuration/management API: this connector
// exposes the application's indices, API keys, and per-index settings. It is
// read-only (full-refresh) and self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package algolia

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
	algoliaDefaultBaseURL = "https://%s.algolia.net"
	algoliaUserAgent      = "polymetrics-go-cli"
	// algoliaMaxPages bounds the indices pagination loop defensively; "all"
	// disables the bound. The default is generous for a management API.
	algoliaDefaultMaxPages = 100
)

func init() {
	connectors.RegisterFactory("algolia", New)
}

// New returns the Algolia connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Algolia source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "algolia" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "algolia",
		DisplayName:     "Algolia",
		IntegrationType: "api",
		Description:     "Reads Algolia indices, API keys, and index settings through the Algolia Search REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Algolia. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := algoliaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(algoliaSecret(cfg)) == "" {
		return errors.New("algolia connector requires secret api_key")
	}
	if strings.TrimSpace(cfg.Config["application_id"]) == "" {
		return errors.New("algolia connector requires config application_id")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the indices list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "1/indexes", url.Values{"page": []string{"0"}}, nil, nil); err != nil {
		return fmt.Errorf("check algolia: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: algoliaStreams()}, nil
}

// Write satisfies the connectors.Connector interface. The Algolia source
// connector is read-only (full-refresh management API); reverse-ETL writes are
// not exposed, so this always reports the operation as unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "indices"
	}
	endpoint, ok := algoliaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("algolia stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	switch stream {
	case "indices":
		maxPages, err := algoliaMaxPages(req.Config)
		if err != nil {
			return err
		}
		return c.harvestIndices(ctx, r, endpoint, maxPages, emit)
	case "index_settings":
		return c.readSettings(ctx, r, endpoint, req.Config, emit)
	default:
		return c.readSingle(ctx, r, endpoint, emit)
	}
}

// harvestIndices drives Algolia's page-number pagination for the indices list.
// The response is {"items":[...],"nbPages":N}; pages run 0..N-1 via the `page`
// query param. connsdk has no body-driven nbPages paginator, so the small loop
// lives here, built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvestIndices(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, maxPages int, emit func(connectors.Record) error) error {
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{"page": []string{strconv.Itoa(page)}}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read algolia %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode algolia %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		nbPages, err := connsdk.StringAt(resp.Body, "nbPages")
		if err != nil {
			return fmt.Errorf("decode algolia %s nbPages: %w", endpoint.resource, err)
		}
		total, perr := strconv.Atoi(strings.TrimSpace(nbPages))
		// Stop when we have consumed the last page, when the API reports no
		// further pages, or when a page returns nothing.
		if perr != nil || total <= 0 || page+1 >= total || len(records) == 0 {
			return nil
		}
	}
	return nil
}

// readSingle reads an endpoint whose response is a single object or a top-level
// array (e.g. the api_keys list under "keys"). It is unpaginated.
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read algolia %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode algolia %s: %w", endpoint.resource, err)
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

// readSettings reads the settings object for a single index named by the
// index_name config. The settings response is a flat object with no envelope, so
// the index name is injected into the record before mapping.
func (c Connector) readSettings(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	indexName := strings.TrimSpace(cfg.Config["index_name"])
	if indexName == "" {
		return errors.New("algolia index_settings stream requires config index_name")
	}
	resource := fmt.Sprintf(endpoint.resource, url.PathEscape(indexName))
	resp, err := r.Do(ctx, http.MethodGet, resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read algolia %s: %w", resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode algolia %s: %w", resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		item["index_name"] = indexName
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise algolia credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	endpoint := algoliaStreamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"name":                 fmt.Sprintf("fixture_index_%d", i),
			"entries":              int64(10 * i),
			"dataSize":             int64(1024 * i),
			"fileSize":             int64(2048 * i),
			"createdAt":            "2026-01-01T00:00:00Z",
			"updatedAt":            "2026-01-02T00:00:00Z",
			"pendingTask":          false,
			"value":                fmt.Sprintf("fixture_key_%d", i),
			"description":          fmt.Sprintf("Fixture key %d", i),
			"acl":                  []any{"search"},
			"indexes":              []any{fmt.Sprintf("fixture_index_%d", i)},
			"index_name":           fmt.Sprintf("fixture_index_%d", i),
			"searchableAttributes": []any{"title", "description"},
			"ranking":              []any{"typo", "geo", "words"},
			"hitsPerPage":          int64(20),
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
		// index_settings is conceptually a single object per index; one record
		// is enough but two keeps the fixture shape uniform across streams.
	}
	return nil
}

// requester builds a connsdk.Requester wired with Algolia's app-id + api-key
// header auth and the resolved base URL. The secret only ever flows into the
// X-Algolia-API-Key header; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := algoliaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := algoliaSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("algolia connector requires secret api_key")
	}
	appID := strings.TrimSpace(cfg.Config["application_id"])
	if appID == "" {
		return nil, errors.New("algolia connector requires config application_id")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("X-Algolia-API-Key", secret, ""),
		UserAgent: algoliaUserAgent,
		DefaultHeaders: map[string]string{
			"X-Algolia-Application-Id": appID,
		},
	}, nil
}

func algoliaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// algoliaBaseURL resolves and validates the base URL. The default is derived
// from the application_id (https://<app>.algolia.net); any override must be an
// absolute https (or http for local test servers) URL with a host to bound SSRF
// risk.
func algoliaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		appID := strings.TrimSpace(cfg.Config["application_id"])
		if appID == "" {
			return "", errors.New("algolia connector requires config application_id (or base_url)")
		}
		return fmt.Sprintf(algoliaDefaultBaseURL, url.PathEscape(appID)), nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("algolia config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("algolia config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("algolia config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func algoliaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" {
		return algoliaDefaultMaxPages, nil
	}
	if raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("algolia config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("algolia config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
