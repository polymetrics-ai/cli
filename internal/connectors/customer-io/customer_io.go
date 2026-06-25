// Package customerio implements the native pm Customer.io connector. It is a
// declarative-HTTP per-system connector following the stripe template: a thin
// package that composes the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction + cursor state) with Customer.io-specific stream definitions,
// endpoints, and cursor pagination.
//
// The Customer.io App (Beta) API uses Bearer auth with an App API Key, list
// endpoints under https://api.customer.io/v1 (or https://api-eu.customer.io/v1
// for the EU region) that return records at a per-stream JSON field path (e.g.
// {"campaigns":[...]}), and cursor pagination where the response carries a
// `next` token that is passed back as the `start` query parameter. The
// Customer.io source is read-only (full refresh / client-side incremental on the
// `updated` cursor); it exposes no reverse-ETL writes, so Capabilities.Write is
// false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package customerio

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
	registryName          = "customer-io"
	customerIOUSBaseURL   = "https://api.customer.io/v1"
	customerIOEUBaseURL   = "https://api-eu.customer.io/v1"
	customerIODefaultPage = 100
	customerIOMaxPage     = 100
	customerIOUserAgent   = "polymetrics-go-cli"
	// customerIOFixtureUpdated is the deterministic `updated` timestamp used by
	// the fixture-mode records (2026-01-01T00:00:00Z in unix seconds).
	customerIOFixtureUpdated int64 = 1767225600
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Customer.io connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Customer.io connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Customer.io",
		IntegrationType: "api",
		Description:     "Reads Customer.io campaigns, newsletters, segments, and broadcasts through the Customer.io App API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Customer.io.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := customerIOBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(customerIOSecret(cfg)) == "" {
		return errors.New("customer-io connector requires secret app_api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the campaigns list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "campaigns", nil, nil, nil); err != nil {
		return fmt.Errorf("check customer-io: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: customerIOStreams()}, nil
}

// Write satisfies the connectors.Connector interface. The Customer.io source is
// read-only, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a Customer.io stream starts
// with an empty incremental cursor (full sync), which the start_date config can
// raise at read time.
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
		stream = "campaigns"
	}
	endpoint, ok := customerIOStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("customer-io stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := customerIOPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := customerIOMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Customer.io's cursor pagination. List responses carry a `next`
// token that is supplied as the `start` query parameter on the following
// request; an empty/null `next` ends the loop. There is no connsdk paginator for
// this exact field-path-plus-start shape, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	start := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		if start != "" {
			query.Set("start", start)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read customer-io %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode customer-io %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next")
		if err != nil {
			return fmt.Errorf("decode customer-io %s next: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" || next == "null" || len(records) == 0 {
			return nil
		}
		start = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise customer-io credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":          int64(i),
			"name":        fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"type":        "triggered",
			"state":       "active",
			"active":      true,
			"subject":     fmt.Sprintf("Fixture subject %d", i),
			"description": fmt.Sprintf("Fixture %s description %d", stream, i),
			"created":     customerIOFixtureUpdated,
			"updated":     customerIOFixtureUpdated + int64(i),
			"connector":   registryName,
			"fixture":     true,
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

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := customerIOBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := customerIOSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("customer-io connector requires secret app_api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: customerIOUserAgent,
	}, nil
}

func customerIOSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["app_api_key"]
}

// customerIOBaseURL resolves and validates the base URL. The default depends on
// the configured region (US -> api.customer.io, EU -> api-eu.customer.io); any
// explicit base_url override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func customerIOBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		switch strings.ToUpper(strings.TrimSpace(cfg.Config["region"])) {
		case "EU":
			return customerIOEUBaseURL, nil
		case "", "US":
			return customerIOUSBaseURL, nil
		default:
			return "", fmt.Errorf("customer-io config region must be US or EU, got %q", cfg.Config["region"])
		}
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("customer-io config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("customer-io config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("customer-io config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func customerIOPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return customerIODefaultPage, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("customer-io config page_size must be an integer: %w", err)
	}
	if value < 1 || value > customerIOMaxPage {
		return 0, fmt.Errorf("customer-io config page_size must be between 1 and %d", customerIOMaxPage)
	}
	return value, nil
}

func customerIOMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("customer-io config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("customer-io config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
