// Package customerly implements the native pm Customerly connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit (Requester
// + Bearer auth + RecordsAt extraction + page-increment pagination) with
// Customerly-specific stream definitions and endpoints, following the shape of
// the reference stripe connector.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Customerly's v1 API is read-only for the streams we expose (users, leads), so
// this connector does not declare Write capability.
package customerly

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
	customerlyDefaultBaseURL  = "https://api.customerly.io/v1"
	customerlyDefaultPageSize = 50
	customerlyMaxPageSize     = 100
	customerlyUserAgent       = "polymetrics-go-cli"
	// customerlyFixtureUpdate is the deterministic last_update used by fixture
	// records.
	customerlyFixtureUpdate = "2026-01-01 00:00:00"
)

func init() {
	connectors.RegisterFactory("customerly", New)
}

// New returns the Customerly connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Customerly connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "customerly" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "customerly",
		DisplayName:     "Customerly",
		IntegrationType: "api",
		Description:     "Reads Customerly users and leads through the Customerly v1 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Customerly.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := customerlyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(customerlySecret(cfg)) == "" {
		return errors.New("customerly connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the users list confirms auth and connectivity without
	// mutating anything.
	q := url.Values{"page": []string{"0"}, "per_page": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "users/list", q, nil, nil); err != nil {
		return fmt.Errorf("check customerly: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: customerlyStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Customerly stream starts
// with an empty incremental cursor (full sync).
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
		stream = "users"
	}
	endpoint, ok := customerlyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("customerly stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := customerlyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := customerlyMaxPages(req.Config)
	if err != nil {
		return err
	}

	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// Write satisfies the connectors.Connector interface. Customerly is exposed
// read-only here (no allow-listed reverse-ETL actions), so writes are rejected.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Customerly's page-increment pagination. Pages start at 0 (the
// API's start_from_page), carry per_page as the size, and a page shorter than
// per_page signals the end. The connsdk PageNumberPaginator coerces a zero
// StartPage to 1, so this small loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{
			"sort":           []string{"last_update"},
			"sort_direction": []string{"desc"},
			"page":           []string{strconv.Itoa(page)},
			"per_page":       []string{strconv.Itoa(pageSize)},
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read customerly %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode customerly %s page %d: %w", endpoint.resource, page, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short (or empty) page means we have reached the end.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise customerly credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"user_id":         int64(i),
			"crmhero_user_id": fmt.Sprintf("%s_fixture_%d", stream, i),
			"email":           fmt.Sprintf("fixture+%d@example.com", i),
			"name":            fmt.Sprintf("Fixture %d", i),
			"username":        fmt.Sprintf("fixture_%d", i),
			"role":            "user",
			"country":         "US",
			"city":            "San Francisco",
			"timezone":        "America/Los_Angeles",
			"sub_active":      true,
			"sub_status":      "active",
			"create_date":     customerlyFixtureUpdate,
			"last_update":     customerlyFixtureUpdate,
			"first_seen_at":   customerlyFixtureUpdate,
			"last_activity":   customerlyFixtureUpdate,
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
	base, err := customerlyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := customerlySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("customerly connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: customerlyUserAgent,
	}, nil
}

func customerlySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// customerlyBaseURL resolves and validates the base URL. The default is
// api.customerly.io/v1; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func customerlyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return customerlyDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("customerly config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("customerly config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("customerly config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func customerlyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return customerlyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("customerly config page_size must be an integer: %w", err)
	}
	if value < 1 || value > customerlyMaxPageSize {
		return 0, fmt.Errorf("customerly config page_size must be between 1 and %d", customerlyMaxPageSize)
	}
	return value, nil
}

func customerlyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("customerly config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("customerly config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
