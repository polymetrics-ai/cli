// Package drip implements the native pm Drip connector. It is a declarative-HTTP
// per-system connector modeled on the stripe reference package: a thin package
// that composes the connsdk toolkit (Requester + Basic auth + RecordsAt
// extraction + page/meta.total_pages pagination) with Drip-specific stream
// definitions and endpoints.
//
// Drip's API is account-scoped (https://api.getdrip.com/v2/{account_id}/...)
// and authenticates with HTTP Basic, using the API key as the username and a
// blank password. The /v2/accounts endpoint is the single account-agnostic
// resource. It is a read-only source connector.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package drip

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
	dripDefaultBaseURL  = "https://api.getdrip.com/v2"
	dripDefaultPageSize = 100
	dripMaxPageSize     = 1000
	dripUserAgent       = "polymetrics-go-cli"
	// dripFixtureCreated is the deterministic created_at timestamp used by the
	// fixture-mode records.
	dripFixtureCreated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("drip", New)
}

// New returns the Drip connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Drip connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "drip" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "drip",
		DisplayName:     "Drip",
		IntegrationType: "api",
		Description:     "Reads Drip subscribers, campaigns, broadcasts, and accounts through the Drip REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Drip. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := dripBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(dripSecret(cfg)) == "" {
		return errors.New("drip connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the account-agnostic accounts list confirms auth and
	// connectivity without requiring account_id or mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "accounts", nil, nil, nil); err != nil {
		return fmt.Errorf("check drip: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: dripStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Drip stream starts with an
// empty incremental cursor (full sync). Drip only supports full_refresh, but the
// cursor scaffolding mirrors the stripe template.
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
		stream = "subscribers"
	}
	endpoint, ok := dripStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("drip stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := endpointPath(req.Config, endpoint)
	if err != nil {
		return err
	}
	pageSize, err := dripPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := dripMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, path, endpoint, pageSize, maxPages, emit)
}

// harvest drives Drip's page/meta pagination. Drip lists return
// {meta:{page,total_pages,...}, <recordsKey>:[...]}. The next page is requested
// with page=<n+1> until the current page reaches meta.total_pages (or an empty
// page is returned). connsdk has no paginator for this exact shape, so the loop
// lives here, built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read drip %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode drip %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		// Stop on an empty page (covers endpoints with no meta, e.g. accounts).
		if len(records) == 0 {
			return nil
		}
		totalPages := pageInt(resp.Body, "meta.total_pages")
		if totalPages == 0 {
			// No pagination metadata: a single page of results.
			return nil
		}
		if page >= totalPages {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise drip credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"status":           "active",
			"created_at":       dripFixtureCreated,
			"name":             fmt.Sprintf("Fixture %d", i),
			"subject":          fmt.Sprintf("Fixture subject %d", i),
			"from_name":        "Fixture Sender",
			"from_email":       "sender@example.com",
			"time_zone":        "Etc/UTC",
			"utc_offset":       int64(0),
			"lifetime_value":   int64(1000 * i),
			"subscriber_count": int64(10 * i),
			"email_count":      int64(i),
			"send_at":          dripFixtureCreated,
			"tags":             []any{"vip"},
			"custom_fields":    map[string]any{"plan": "pro"},
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

// requester builds a connsdk.Requester wired with HTTP Basic auth (api_key as
// username, blank password) and the resolved base URL. The secret only ever
// flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := dripBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := dripSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("drip connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, ""),
		UserAgent: dripUserAgent,
	}, nil
}

// endpointPath builds the request path for a stream. Account-scoped streams are
// prefixed with the account_id segment; the accounts stream is global.
func endpointPath(cfg connectors.RuntimeConfig, endpoint streamEndpoint) (string, error) {
	if !endpoint.accountScoped {
		return endpoint.resource, nil
	}
	account := strings.TrimSpace(cfg.Config["account_id"])
	if account == "" {
		return "", fmt.Errorf("drip stream %q requires config account_id", endpoint.resource)
	}
	return url.PathEscape(account) + "/" + endpoint.resource, nil
}

func dripSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// dripBaseURL resolves and validates the base URL. The default is
// api.getdrip.com/v2; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func dripBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return dripDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("drip config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("drip config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("drip config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func dripPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return dripDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("drip config page_size must be an integer: %w", err)
	}
	if value < 1 || value > dripMaxPageSize {
		return 0, fmt.Errorf("drip config page_size must be between 1 and %d", dripMaxPageSize)
	}
	return value, nil
}

func dripMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("drip config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("drip config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// pageInt reads an integer value at a dotted JSON path, returning 0 when missing
// or unparseable.
func pageInt(body []byte, path string) int {
	s, err := connsdk.StringAt(body, path)
	if err != nil || s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

// Write is not supported: Drip is a read-only source connector here.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
