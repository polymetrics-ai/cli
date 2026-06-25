// Package codefresh implements the native pm Codefresh connector. It follows the
// declarative-HTTP template established by the stripe package: a thin package
// that composes the connsdk toolkit (Requester + API-key header auth + RecordsAt
// extraction) with Codefresh-specific stream definitions, endpoints, and
// page-based pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Codefresh is read-only here: the connector surfaces projects, pipelines,
// agents, and contexts. There are no obvious safe reverse-ETL writes, so
// Capabilities.Write is false.
package codefresh

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
	codefreshDefaultBaseURL  = "https://g.codefresh.io/api"
	codefreshDefaultPageSize = 50
	codefreshMaxPageSize     = 100
	codefreshUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("codefresh", New)
}

// New returns the Codefresh connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Codefresh connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "codefresh" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "codefresh",
		DisplayName:     "Codefresh",
		IntegrationType: "api",
		Description:     "Reads Codefresh projects, pipelines, runner agents, and shared contexts through the Codefresh REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Codefresh. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := codefreshBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(codefreshAPIKey(cfg)) == "" {
		return errors.New("codefresh connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the projects list confirms auth and connectivity
	// without mutating anything.
	q := url.Values{"limit": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "projects", q, nil, nil); err != nil {
		return fmt.Errorf("check codefresh: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: codefreshStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "projects"
	}
	endpoint, ok := codefreshStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("codefresh stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := codefreshPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := codefreshMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Codefresh page-based pagination. Codefresh list endpoints take
// a 1-based "page" param plus a "limit" page size; a page shorter than the
// requested size signals the end. Responses come in a few shapes (bare arrays
// for /agents and /contexts, {projects:[...]} for /projects, {docs:[...]} for
// /pipelines), all handled via the endpoint's recordsPath and connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("page", strconv.Itoa(page))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read codefresh %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode codefresh %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short (or empty) page means the listing is exhausted.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise codefresh credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		id := fmt.Sprintf("%s_fixture_%d", endpoint.resource, i)
		item := map[string]any{
			"id":              id,
			"_id":             id,
			"projectName":     fmt.Sprintf("project-%d", i),
			"favorite":        false,
			"pipelinesNumber": int64(i),
			"updatedAt":       "2026-01-01T00:00:00Z",
			"name":            fmt.Sprintf("agent-%d", i),
			"version":         "1.0.0",
			"status":          "online",
			"created_at":      "2026-01-01T00:00:00Z",
			"type":            "config",
			"owner":           "account",
			"metadata": map[string]any{
				"name":       fmt.Sprintf("%s-%d", stream, i),
				"project":    "fixture",
				"isPublic":   false,
				"created_at": "2026-01-01T00:00:00Z",
				"updated_at": "2026-01-01T00:00:00Z",
				"owner":      "account",
			},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Codefresh API-key header auth
// (the token goes into the Authorization header verbatim, no Bearer prefix) and
// the optional account scoping header. The secret only ever flows into the
// authenticator/headers; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := codefreshBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := codefreshAPIKey(cfg)
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("codefresh connector requires secret api_key")
	}
	headers := map[string]string{}
	if account := strings.TrimSpace(codefreshAccountID(cfg)); account != "" {
		// Codefresh scopes some requests by account; pass it through as an
		// access-token header alongside the API key.
		headers["X-Access-Token"] = account
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           connsdk.APIKeyHeader("Authorization", key, ""),
		UserAgent:      codefreshUserAgent,
		DefaultHeaders: headers,
	}, nil
}

func codefreshAPIKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

func codefreshAccountID(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["account_id"]
}

// codefreshBaseURL resolves and validates the base URL. The default is
// g.codefresh.io/api; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func codefreshBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return codefreshDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("codefresh config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("codefresh config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("codefresh config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func codefreshPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return codefreshDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("codefresh config page_size must be an integer: %w", err)
	}
	if value < 1 || value > codefreshMaxPageSize {
		return 0, fmt.Errorf("codefresh config page_size must be between 1 and %d", codefreshMaxPageSize)
	}
	return value, nil
}

func codefreshMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("codefresh config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("codefresh config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// Write satisfies the connectors.Connector interface. Codefresh is read-only in
// this connector, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
