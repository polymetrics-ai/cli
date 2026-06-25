// Package closecom implements the native pm Close.com (Close CRM) connector. It
// is a declarative-HTTP per-system connector following the stripe template: a
// thin package that composes the connsdk toolkit (Requester + HTTP Basic auth +
// RecordsAt extraction + cursor state) with Close-specific stream definitions,
// endpoints, and offset pagination.
//
// The Close API uses HTTP Basic auth with the API key as the username and an
// empty password, list endpoints under https://api.close.com/api/v1 that return
// {"data":[...],"has_more":bool}, and _skip/_limit offset pagination. The Close
// source is read-only (full-refresh / incremental on date_updated); it exposes
// no reverse-ETL writes, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package closecom

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	closeDefaultBaseURL  = "https://api.close.com/api/v1"
	closeDefaultPageSize = 100
	closeMaxPageSize     = 100
	closeUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("close-com", New)
}

// New returns the Close.com connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Close.com connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "close-com" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "close-com",
		DisplayName:     "Close.com",
		IntegrationType: "api",
		Description:     "Reads Close CRM leads, contacts, opportunities, activities, and users through the Close REST API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Close. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := closeBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(closeSecret(cfg)) == "" {
		return errors.New("close-com connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the lead list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "lead/", url.Values{"_limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check close-com: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: closeStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Close stream starts with
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
		stream = "leads"
	}
	endpoint, ok := closeStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("close-com stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := closePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := closeMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Close's _skip/_limit offset pagination. Close lists return
// {data:[...], has_more:bool}; the next page is requested with _skip advanced by
// the page size. The loop lives here (rather than connsdk.Harvest) so it can stop
// on has_more=false rather than only on a short page.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	skip := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("_limit", strconv.Itoa(pageSize))
		query.Set("_skip", strconv.Itoa(skip))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read close-com %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode close-com %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		hasMore, err := connsdk.StringAt(resp.Body, "has_more")
		if err != nil {
			return fmt.Errorf("decode close-com %s has_more: %w", endpoint.resource, err)
		}
		if hasMore != "true" || len(records) == 0 {
			return nil
		}
		skip += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise close-com credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	resource := strings.TrimRight(endpoint.resource, "/")
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":              fmt.Sprintf("%s_fixture_%d", resource, i),
			"display_name":    fmt.Sprintf("Fixture %d", i),
			"name":            fmt.Sprintf("Fixture %d", i),
			"first_name":      "Fixture",
			"last_name":       fmt.Sprintf("%d", i),
			"email":           fmt.Sprintf("fixture+%d@example.com", i),
			"title":           "Contact",
			"_type":           "Note",
			"status_label":    "Potential",
			"status_type":     "active",
			"direction":       "outbound",
			"value":           int64(1000 * i),
			"value_currency":  "USD",
			"confidence":      int64(50),
			"organization_id": "orga_fixture",
			"lead_id":         "lead_fixture_1",
			"user_id":         "user_fixture_1",
			"user_name":       "Fixture User",
			"created_by":      "user_fixture_1",
			"date_created":    fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"date_updated":    fmt.Sprintf("2026-01-0%dT00:00:00Z", i+1),
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "close-com"
		record["stream"] = stream
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth (API key as
// username, empty password) and the resolved base URL. The secret only ever
// flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := closeBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := closeSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("close-com connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, ""),
		UserAgent: closeUserAgent,
	}, nil
}

func closeSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// closeBaseURL resolves and validates the base URL. The default is
// api.close.com; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func closeBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return closeDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("close-com config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("close-com config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("close-com config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func closePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return closeDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("close-com config page_size must be an integer: %w", err)
	}
	if value < 1 || value > closeMaxPageSize {
		return 0, fmt.Errorf("close-com config page_size must be between 1 and %d", closeMaxPageSize)
	}
	return value, nil
}

func closeMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("close-com config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("close-com config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// Write satisfies the connectors.Connector interface. The Close source is
// read-only (no approved reverse-ETL actions), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
