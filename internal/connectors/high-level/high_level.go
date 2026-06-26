// Package highlevel implements the native pm HighLevel (Go HighLevel /
// LeadConnector) connector. It is a declarative-HTTP per-system connector built
// on the connsdk toolkit (Requester + x-api-key auth + RecordsAt extraction)
// wired to the Airbyte source-high-level proxy API. It mirrors the stripe and
// cal-com reference connectors' shape.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// The directory and registry key are "high-level"; the Go package identifier is
// "highlevel" (hyphens are not valid in Go identifiers).
package highlevel

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
	registryName   = "high-level"
	displayName    = "High Level"
	defaultBaseURL = "https://api.leadconnectorpro.co"
	// defaultAPIVersion is the LeadConnector API version header value. The proxy
	// accepts it for forward compatibility and downstream calls require it.
	defaultAPIVersion     = "2021-07-28"
	defaultPageSize       = 100
	maxPageSize           = 100
	userAgent             = "polymetrics-go-cli"
	secretField           = "api_key"
	fixtureRecordsPerPage = 2
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the HighLevel connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm HighLevel connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     displayName,
		IntegrationType: "api",
		Description:     "Reads HighLevel (Go HighLevel / LeadConnector) contacts, opportunities, pipelines, custom fields, and form submissions for a location through the HighLevel REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to HighLevel.
// In fixture mode it short-circuits without a network call.
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
	location, err := locationID(cfg)
	if err != nil {
		return err
	}
	if strings.TrimSpace(secret(cfg)) == "" {
		return fmt.Errorf("%s connector requires secret %s", registryName, secretField)
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the pipelines endpoint confirms auth and connectivity
	// without mutating anything; pipelines is small and unpaginated.
	query := url.Values{"locationId": []string{location}}
	if err := r.DoJSON(ctx, http.MethodGet, "airbyte/pipelines", query, nil, nil); err != nil {
		return fmt.Errorf("check %s: %w", registryName, err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: highLevelStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "contacts"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("%s stream %q not found", registryName, stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	location, err := locationID(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := pageSizeFromConfig(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPagesFromConfig(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, location, pageSize, maxPages, emit)
}

// harvest drives HighLevel pagination. Cursor streams follow the absolute
// meta.nextPageUrl returned in each page (falling back to a top-level
// nextPageUrl); when no next URL is present, pagination ends. Single-request
// streams (pipelines, custom_fields) make exactly one call.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, location string, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := "airbyte/" + endpoint.resource
	base := url.Values{}
	base.Set("locationId", location)
	if endpoint.style == styleCursor {
		base.Set("limit", strconv.Itoa(pageSize))
	}

	if endpoint.style == styleNone {
		resp, err := r.Do(ctx, http.MethodGet, path, base, nil)
		if err != nil {
			return fmt.Errorf("read %s %s: %w", registryName, endpoint.resource, err)
		}
		_, err = c.emitResponse(ctx, endpoint, resp.Body, emit)
		return err
	}

	// Cursor pagination: the first request uses path+base; subsequent requests
	// follow the absolute next URL returned by the server. connsdk.Requester
	// treats an http(s) path as absolute and uses it as-is, so we pass the next
	// URL directly as the path with no extra query params.
	reqPath := path
	query := base
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, reqPath, query, nil)
		if err != nil {
			return fmt.Errorf("read %s %s: %w", registryName, endpoint.resource, err)
		}
		count, err := c.emitResponse(ctx, endpoint, resp.Body, emit)
		if err != nil {
			return err
		}
		next := nextPageURL(resp.Body)
		// Stop when there is no next URL, or when a page returns no records to
		// avoid following a self-referential cursor forever.
		if next == "" || count == 0 {
			return nil
		}
		reqPath = next
		query = nil
	}
	return nil
}

// emitResponse extracts records at the endpoint's selector and emits each mapped
// record, returning the count so the pagination loop can detect an empty page.
func (c Connector) emitResponse(ctx context.Context, endpoint streamEndpoint, body []byte, emit func(connectors.Record) error) (int, error) {
	records, err := connsdk.RecordsAt(body, endpoint.recordsPath)
	if err != nil {
		return 0, fmt.Errorf("decode %s %s page: %w", registryName, endpoint.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return 0, err
		}
	}
	return len(records), nil
}

// nextPageURL reads the next-page URL out of a HighLevel response body. The
// canonical location is meta.nextPageUrl; a top-level nextPageUrl is accepted as
// a fallback. An empty string means there is no next page.
func nextPageURL(body []byte) string {
	if v, err := connsdk.StringAt(body, "meta.nextPageUrl"); err == nil {
		if s := strings.TrimSpace(v); s != "" && !strings.EqualFold(s, "null") {
			return s
		}
	}
	if v, err := connsdk.StringAt(body, "nextPageUrl"); err == nil {
		if s := strings.TrimSpace(v); s != "" && !strings.EqualFold(s, "null") {
			return s
		}
	}
	return ""
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= fixtureRecordsPerPage; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":              fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"locationId":      "loc_fixture",
			"contactName":     fmt.Sprintf("Fixture Contact %d", i),
			"firstName":       fmt.Sprintf("Fixture%d", i),
			"lastName":        "Example",
			"name":            fmt.Sprintf("Fixture %s %d", stream, i),
			"email":           fmt.Sprintf("fixture+%d@example.com", i),
			"phone":           "+15555550000",
			"type":            "lead",
			"source":          "fixture",
			"pipelineId":      "pl_fixture",
			"pipelineStageId": "stage_fixture",
			"status":          "open",
			"monetaryValue":   int64(100 * i),
			"contactId":       "contacts_fixture_1",
			"assignedTo":      "user_fixture",
			"formId":          "form_fixture",
			"fieldKey":        fmt.Sprintf("contact.fixture_%d", i),
			"dataType":        "TEXT",
			"model":           "contact",
			"position":        int64(i),
			"stages":          []any{},
			"dateAdded":       "2026-01-01T00:00:00Z",
			"dateUpdated":     "2026-01-01T00:00:00Z",
			"createdAt":       "2026-01-01T00:00:00Z",
			"connector":       registryName,
			"fixture":         true,
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with x-api-key auth, the resolved
// base URL, and the Version header. The secret only ever flows into the
// authenticator header; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("%s connector requires secret %s", registryName, secretField)
	}
	headers := map[string]string{
		"Version": apiVersion(cfg),
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           connsdk.APIKeyHeader("x-api-key", token, ""),
		UserAgent:      userAgent,
		DefaultHeaders: headers,
	}, nil
}

func secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[secretField]
}

// locationID resolves the required location_id config value.
func locationID(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config != nil {
		if v := strings.TrimSpace(cfg.Config["location_id"]); v != "" {
			return v, nil
		}
	}
	return "", fmt.Errorf("%s connector requires config location_id", registryName)
}

// apiVersion resolves the Version header value, allowing a config override for
// forward compatibility.
func apiVersion(cfg connectors.RuntimeConfig) string {
	if cfg.Config != nil {
		if v := strings.TrimSpace(cfg.Config["api_version"]); v != "" {
			return v
		}
	}
	return defaultAPIVersion
}

// baseURL resolves and validates the base URL. The default is the proxy host;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", registryName, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", registryName, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New(registryName + " config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSizeFromConfig(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config page_size must be an integer: %w", registryName, err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("%s config page_size must be between 1 and %d", registryName, maxPageSize)
	}
	return value, nil
}

func maxPagesFromConfig(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config max_pages must be an integer, all, or unlimited: %w", registryName, err)
	}
	if value < 0 {
		return 0, errors.New(registryName + " config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is not supported: HighLevel is exposed read-only. The method exists to
// satisfy the connectors.Connector interface.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
