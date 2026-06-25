// Package klaviyo implements the native pm Klaviyo connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference connector: a thin package that composes the connsdk toolkit
// (Requester + APIKeyHeader auth + RecordsAt extraction) with Klaviyo-specific
// stream definitions, endpoints, and JSON:API cursor pagination.
//
// Klaviyo's REST API is a JSON:API: list responses return {data:[...],
// links:{next: "<absolute url>"}}. Each resource object carries a top-level
// string id plus a nested attributes object; the record mappers in streams.go
// flatten a curated subset to the top level.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// Klaviyo is read-only here: there is no obviously-safe reverse-ETL write set,
// so Capabilities.Write is false.
package klaviyo

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
	klaviyoDefaultBaseURL  = "https://a.klaviyo.com/api"
	klaviyoDefaultRevision = "2024-10-15"
	klaviyoDefaultPageSize = 100
	klaviyoMaxPageSize     = 100
	klaviyoUserAgent       = "polymetrics-go-cli"
	// klaviyoFixtureCreated is the deterministic created timestamp used by the
	// fixture-mode records.
	klaviyoFixtureCreated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("klaviyo", New)
}

// New returns the Klaviyo connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Klaviyo connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "klaviyo" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "klaviyo",
		DisplayName:     "Klaviyo",
		IntegrationType: "api",
		Description:     "Reads Klaviyo profiles, events, campaigns, lists, metrics, and segments through the Klaviyo REST (JSON:API) API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Klaviyo. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := klaviyoBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(klaviyoSecret(cfg)) == "" {
		return errors.New("klaviyo connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the accounts endpoint confirms auth and connectivity
	// without mutating anything. (accounts is the canonical "who am I" read.)
	if err := r.DoJSON(ctx, http.MethodGet, "accounts", nil, nil, nil); err != nil {
		return fmt.Errorf("check klaviyo: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: klaviyoStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Klaviyo stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "profiles"
	}
	endpoint, ok := klaviyoStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("klaviyo stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := klaviyoPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := klaviyoMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Klaviyo's JSON:API cursor pagination. List responses return
// {data:[...], links:{next:"<absolute url>"}}; the next page is fetched by
// following links.next verbatim (the connsdk Requester treats an absolute URL
// path as-is). The loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt, because the next-token is a full URL in
// the body rather than a Link header or a bare cursor param.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	// First page: relative resource path + page[size].
	path := endpoint.resource
	query := url.Values{}
	query.Set("page[size]", strconv.Itoa(pageSize))

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read klaviyo %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode klaviyo %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "links.next")
		if err != nil {
			return fmt.Errorf("decode klaviyo %s links.next: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return nil
		}
		// Subsequent pages: follow the absolute next URL verbatim. The cursor is
		// already encoded in it, so no extra query params are merged.
		path = next
		query = nil
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise klaviyo credential-free (mirrors stripe).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":   fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"type": strings.TrimSuffix(stream, "s"),
			"attributes": map[string]any{
				"email":            fmt.Sprintf("fixture+%d@example.com", i),
				"phone_number":     "",
				"external_id":      fmt.Sprintf("ext_%d", i),
				"first_name":       fmt.Sprintf("Fixture %d", i),
				"last_name":        "Klaviyo",
				"organization":     "Example Co",
				"name":             fmt.Sprintf("Fixture %s %d", stream, i),
				"status":           "active",
				"archived":         false,
				"channel":          "email",
				"is_active":        true,
				"is_processing":    false,
				"integration_name": "Klaviyo",
				"timestamp":        int64(1767225600 + i),
				"datetime":         klaviyoFixtureCreated,
				"uuid":             fmt.Sprintf("uuid_%d", i),
				"created":          klaviyoFixtureCreated,
				"updated":          klaviyoFixtureCreated,
				"created_at":       klaviyoFixtureCreated,
				"updated_at":       klaviyoFixtureCreated,
				"scheduled_at":     "",
				"send_time":        "",
			},
		}
		record := endpoint.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Klaviyo-API-Key auth, the
// resolved base URL, and the required revision header. The secret only ever
// flows into connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := klaviyoBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := klaviyoSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("klaviyo connector requires secret api_key")
	}
	headers := map[string]string{
		"revision": klaviyoRevision(cfg),
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           connsdk.APIKeyHeader("Authorization", secret, "Klaviyo-API-Key "),
		UserAgent:      klaviyoUserAgent,
		DefaultHeaders: headers,
	}, nil
}

func klaviyoSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// klaviyoRevision resolves the API revision (date version) header. The default
// is a known-good stable revision; an override may be supplied via config.
func klaviyoRevision(cfg connectors.RuntimeConfig) string {
	if cfg.Config != nil {
		if rev := strings.TrimSpace(cfg.Config["revision"]); rev != "" {
			return rev
		}
	}
	return klaviyoDefaultRevision
}

// klaviyoBaseURL resolves and validates the base URL. The default is
// a.klaviyo.com/api; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func klaviyoBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return klaviyoDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("klaviyo config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("klaviyo config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("klaviyo config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func klaviyoPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return klaviyoDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("klaviyo config page_size must be an integer: %w", err)
	}
	if value < 1 || value > klaviyoMaxPageSize {
		return 0, fmt.Errorf("klaviyo config page_size must be between 1 and %d", klaviyoMaxPageSize)
	}
	return value, nil
}

func klaviyoMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("klaviyo config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("klaviyo config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Klaviyo is read-only in pm:
// there is no allow-listed reverse-ETL write set, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
