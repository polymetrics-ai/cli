// Package invoiced implements the native pm Invoiced source connector. It is a
// declarative-HTTP per-system connector built on the stripe template: a thin
// package that composes the connsdk toolkit (Requester + HTTP Basic auth +
// root-array record extraction + page-number pagination) with Invoiced-specific
// stream definitions and endpoints.
//
// The Invoiced REST API (https://api.invoiced.com) authenticates with HTTP Basic
// where the API key is the username and the password is blank. List endpoints
// return a top-level JSON array and paginate with page/per_page query params.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package invoiced

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
	invoicedDefaultBaseURL  = "https://api.invoiced.com"
	invoicedDefaultPageSize = 100
	invoicedMaxPageSize     = 100
	invoicedUserAgent       = "polymetrics-go-cli"
	// invoicedFixtureUpdated is the deterministic `updated_at` timestamp used by
	// fixture-mode records (2026-01-01T00:00:00Z in unix seconds).
	invoicedFixtureUpdated int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("invoiced", New)
}

// New returns the Invoiced connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Invoiced source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "invoiced" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "invoiced",
		DisplayName:     "Invoiced",
		IntegrationType: "api",
		Description:     "Reads Invoiced customers, invoices, payments, subscriptions, and estimates through the Invoiced REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Invoiced. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := invoicedBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(invoicedSecret(cfg)) == "" {
		return errors.New("invoiced connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the customers list confirms auth and connectivity
	// without mutating anything.
	q := url.Values{"per_page": []string{"1"}, "page": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "customers", q, nil, nil); err != nil {
		return fmt.Errorf("check invoiced: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: invoicedStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: an Invoiced stream starts
// with an empty cursor (full refresh), the only mode the upstream API exposes.
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
		stream = "customers"
	}
	endpoint, ok := invoicedStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("invoiced stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := invoicedPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := invoicedMaxPages(req.Config)
	if err != nil {
		return err
	}

	// Invoiced lists return a top-level JSON array and paginate with
	// page/per_page (PageIncrement). connsdk's PageNumberPaginator + Harvest
	// drive this exactly: the mapper flattens each root-array element.
	paginator := &connsdk.PageNumberPaginator{
		PageParam: "page",
		SizeParam: "per_page",
		StartPage: 1,
		PageSize:  pageSize,
	}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, nil, paginator, "", maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

// Write satisfies the connectors.Connector interface. Invoiced is a read-only
// source connector (Capabilities.Write=false), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise invoiced credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":              int64(1000 + i),
			"object":          strings.TrimSuffix(stream, "s"),
			"name":            fmt.Sprintf("Fixture %d", i),
			"number":          fmt.Sprintf("%s-%04d", strings.ToUpper(stream[:3]), i),
			"email":           fmt.Sprintf("fixture+%d@example.com", i),
			"type":            "company",
			"currency":        "usd",
			"balance":         float64(100 * i),
			"phone":           "555-0100",
			"country":         "US",
			"customer":        int64(1001),
			"invoice":         int64(2001),
			"status":          "paid",
			"total":           float64(1000 * i),
			"amount":          float64(1000 * i),
			"method":          "credit_card",
			"paid":            true,
			"closed":          false,
			"approved":        true,
			"plan":            "monthly",
			"quantity":        int64(1),
			"date":            invoicedFixtureUpdated,
			"due_date":        invoicedFixtureUpdated + 86400,
			"expiration_date": invoicedFixtureUpdated + 86400,
			"start_date":      invoicedFixtureUpdated,
			"period_start":    invoicedFixtureUpdated,
			"period_end":      invoicedFixtureUpdated + 2592000,
			"created_at":      invoicedFixtureUpdated,
			"updated_at":      invoicedFixtureUpdated + int64(i),
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
	base, err := invoicedBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := invoicedSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("invoiced connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, ""),
		UserAgent: invoicedUserAgent,
	}, nil
}

func invoicedSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// invoicedBaseURL resolves and validates the base URL. The default is
// api.invoiced.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func invoicedBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return invoicedDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invoiced config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("invoiced config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("invoiced config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func invoicedPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return invoicedDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invoiced config page_size must be an integer: %w", err)
	}
	if value < 1 || value > invoicedMaxPageSize {
		return 0, fmt.Errorf("invoiced config page_size must be between 1 and %d", invoicedMaxPageSize)
	}
	return value, nil
}

func invoicedMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invoiced config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("invoiced config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
