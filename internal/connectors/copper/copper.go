// Package copper implements the native pm Copper (ProsperWorks) CRM connector.
// It is a declarative-HTTP per-system connector built on the connsdk toolkit:
// a connsdk.Requester carries the three Copper auth headers (X-PW-AccessToken,
// X-PW-Application, X-PW-UserEmail), and the read path drives Copper's
// POST /<resource>/search page_number pagination over top-level JSON arrays.
//
// It is read-only: the Copper API does support writes, but reverse-ETL into a
// live CRM is out of scope for this port, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package copper

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
	copperDefaultBaseURL  = "https://api.copper.com/developer_api/v1"
	copperApplication     = "developer_api"
	copperDefaultPageSize = 100
	copperMaxPageSize     = 200
	copperUserAgent       = "polymetrics-go-cli"
	// copperFixtureModified is the deterministic `date_modified` timestamp used
	// by the fixture-mode records (2026-01-01T00:00:00Z in unix seconds).
	copperFixtureModified int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("copper", New)
}

// New returns the Copper connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Copper connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "copper" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "copper",
		DisplayName:     "Copper",
		IntegrationType: "api",
		Description:     "Reads Copper CRM people, companies, opportunities, leads, and tasks through the Copper REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Copper. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := copperBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(copperSecret(cfg)) == "" {
		return errors.New("copper connector requires secret api_key")
	}
	if strings.TrimSpace(cfg.Config["user_email"]) == "" {
		return errors.New("copper connector requires config user_email")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded single-record search confirms auth and connectivity without
	// mutating anything.
	body := map[string]any{"page_number": 1, "page_size": 1}
	if _, err := r.Do(ctx, http.MethodPost, "people/search", nil, body); err != nil {
		return fmt.Errorf("check copper: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: copperStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Copper stream starts with
// an empty incremental cursor (full refresh — Copper only supports full_refresh
// upstream).
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
		stream = "people"
	}
	endpoint, ok := copperStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("copper stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := copperPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := copperMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Copper's page_number pagination. Copper search endpoints are
// POST /<resource>/search returning a top-level JSON array; the next page is
// requested by incrementing page_number, and the walk stops when a page returns
// fewer than page_size records. There is no connsdk Paginator for body-driven
// page numbers over a top-level array, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := endpoint.resource + "/search"
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		body := map[string]any{
			"page_number": page,
			"page_size":   pageSize,
			"sort_by":     "date_modified",
		}
		resp, err := r.Do(ctx, http.MethodPost, path, nil, body)
		if err != nil {
			return fmt.Errorf("read copper %s: %w", endpoint.resource, err)
		}
		// Copper returns a top-level array; "" selects the root.
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode copper %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise copper credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               int64(i),
			"name":             fmt.Sprintf("%s fixture %d", strings.TrimSuffix(stream, "s"), i),
			"first_name":       fmt.Sprintf("Fixture%d", i),
			"last_name":        "Example",
			"company_name":     "Fixture Co",
			"company_id":       int64(1000),
			"assignee_id":      int64(42),
			"status":           "Open",
			"monetary_value":   float64(1000 * i),
			"emails":           []any{map[string]any{"email": fmt.Sprintf("fixture+%d@example.com", i), "category": "work"}},
			"phone_numbers":    []any{map[string]any{"number": "+15555550100", "category": "work"}},
			"date_created":     copperFixtureModified,
			"date_modified":    copperFixtureModified + int64(i),
			"related_resource": map[string]any{"id": int64(7), "type": "person"},
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "copper"
		record["fixture"] = true
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the three Copper auth headers,
// the resolved base URL, and the user-agent. The api_key only ever flows into a
// header value; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := copperBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := copperSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("copper connector requires secret api_key")
	}
	email := strings.TrimSpace(cfg.Config["user_email"])
	if email == "" {
		return nil, errors.New("copper connector requires config user_email")
	}
	headers := map[string]string{
		"X-PW-AccessToken": strings.TrimSpace(secret),
		"X-PW-Application": copperApplication,
		"X-PW-UserEmail":   email,
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		UserAgent:      copperUserAgent,
		DefaultHeaders: headers,
	}, nil
}

func copperSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// copperBaseURL resolves and validates the base URL. The default is
// api.copper.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func copperBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return copperDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("copper config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("copper config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("copper config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func copperPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return copperDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("copper config page_size must be an integer: %w", err)
	}
	if value < 1 || value > copperMaxPageSize {
		return 0, fmt.Errorf("copper config page_size must be between 1 and %d", copperMaxPageSize)
	}
	return value, nil
}

func copperMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("copper config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("copper config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Copper is a read-only source in this port.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
