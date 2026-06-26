// Package front implements the native pm Front connector. It is a
// declarative-HTTP per-system connector built on the stripe template: a thin
// package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction) with Front-specific stream definitions, endpoints, and
// the Front body-cursor pagination shape ({_results:[...], _pagination:{next}}).
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Front is read-only here: the API key only grants list access and the catalog
// declares full_refresh sync, so no reverse-ETL write surface is exposed.
package front

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
	frontDefaultBaseURL  = "https://api2.frontapp.com"
	frontDefaultPageSize = 50
	frontMaxPageSize     = 100
	frontUserAgent       = "polymetrics-go-cli"
	// frontFixtureCreated is the deterministic created_at value (unix seconds,
	// 2026-01-01T00:00:00Z) used by fixture-mode records.
	frontFixtureCreated int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("front", New)
}

// New returns the Front connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Front connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "front" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "front",
		DisplayName:     "Front",
		IntegrationType: "api",
		Description:     "Reads Front contacts, conversations, inboxes, tags, teammates, and channels through the Front Core REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Front. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := frontBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(frontSecret(cfg)) == "" {
		return errors.New("front connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the inboxes list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "inboxes", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check front: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: frontStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Front stream starts with
// an empty cursor (full sync). Front list endpoints are not range-filterable by
// the connector, so the cursor is informational for downstream incremental
// dedupe rather than a server-side filter.
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
		stream = "contacts"
	}
	endpoint, ok := frontStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("front stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := frontPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := frontMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Front's body-cursor pagination. Front lists return
// {_results:[...], _pagination:{next:<absolute url|null>}}; the first page is
// requested at <resource>?limit=<n> and each subsequent page is requested by
// following the absolute _pagination.next URL. connsdk.Requester.Do treats an
// http(s) path as absolute, so the next URL is passed through directly. There is
// no connsdk paginator for this exact shape, so the loop lives here, built on
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	// First request: relative resource path with the page size.
	path := endpoint.resource
	query := url.Values{"limit": []string{strconv.Itoa(pageSize)}}

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read front %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "_results")
		if err != nil {
			return fmt.Errorf("decode front %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "_pagination.next")
		if err != nil {
			return fmt.Errorf("decode front %s pagination: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" || strings.EqualFold(next, "null") {
			return nil
		}
		// Subsequent pages: follow the absolute next URL verbatim. It already
		// carries the page token (and limit), so no extra query is merged.
		path = next
		query = nil
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise front credential-free (mirrors the stripe
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                               fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":                             fmt.Sprintf("Fixture %d", i),
			"description":                      "front fixture record",
			"subject":                          fmt.Sprintf("Fixture conversation %d", i),
			"status":                           "open",
			"email":                            fmt.Sprintf("fixture+%d@example.com", i),
			"username":                         fmt.Sprintf("fixture%d", i),
			"first_name":                       "Fixture",
			"last_name":                        fmt.Sprintf("%d", i),
			"address":                          fmt.Sprintf("fixture+%d@example.com", i),
			"type":                             "smtp",
			"send_as":                          fmt.Sprintf("fixture+%d@example.com", i),
			"highlight":                        "blue",
			"is_private":                       false,
			"is_public":                        true,
			"is_spammer":                       false,
			"is_admin":                         false,
			"is_available":                     true,
			"is_blocked":                       false,
			"is_valid":                         true,
			"is_visible_in_conversation_lists": true,
			"created_at":                       frontFixtureCreated + int64(i),
			"updated_at":                       frontFixtureCreated + int64(i),
			"last_message_at":                  frontFixtureCreated + int64(i),
			"waiting_since":                    frontFixtureCreated + int64(i),
			"custom_fields":                    map[string]any{},
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
	base, err := frontBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := frontSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("front connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: frontUserAgent,
	}, nil
}

func frontSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// frontBaseURL resolves and validates the base URL. The default is
// api2.frontapp.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func frontBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return frontDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("front config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("front config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("front config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func frontPageSize(cfg connectors.RuntimeConfig) (int, error) {
	// page_limit is the catalog config field; page_size is accepted as an alias.
	raw := strings.TrimSpace(cfg.Config["page_limit"])
	if raw == "" {
		raw = strings.TrimSpace(cfg.Config["page_size"])
	}
	if raw == "" {
		return frontDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("front config page_limit must be an integer: %w", err)
	}
	if value < 1 || value > frontMaxPageSize {
		return 0, fmt.Errorf("front config page_limit must be between 1 and %d", frontMaxPageSize)
	}
	return value, nil
}

func frontMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("front config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("front config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Front is exposed read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
