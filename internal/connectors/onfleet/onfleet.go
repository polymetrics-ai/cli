// Package onfleet implements the native pm Onfleet connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference connector: a thin package that composes the connsdk toolkit
// (Requester + Basic auth + RecordsAt extraction + lastId pagination) with
// Onfleet-specific stream definitions and endpoints.
//
// Onfleet authenticates with HTTP Basic auth where the API key is the username
// and the password is empty (the catalog carries a placeholder `password`
// secret that Onfleet expects to be blank). The connector is read-only: the
// Onfleet write surface (creating tasks, etc.) is operationally sensitive and
// out of scope for a safe reverse-ETL allow-list, so Capabilities.Write=false.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package onfleet

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
	onfleetDefaultBaseURL = "https://onfleet.com/api/v2"
	onfleetUserAgent      = "polymetrics-go-cli"
	// onfleetFixtureCreated is the deterministic timeCreated value used by the
	// fixture-mode records (unix milliseconds, 2026-01-01T00:00:00Z).
	onfleetFixtureCreated int64 = 1767225600000
)

func init() {
	connectors.RegisterFactory("onfleet", New)
}

// New returns the Onfleet connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Onfleet connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "onfleet" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "onfleet",
		DisplayName:     "Onfleet",
		IntegrationType: "api",
		Description:     "Reads Onfleet tasks, workers, teams, hubs, and administrators through the Onfleet REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Onfleet. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := onfleetBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(onfleetAPIKey(cfg)) == "" {
		return errors.New("onfleet connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// The auth-test endpoint confirms credentials without reading any data.
	if err := r.DoJSON(ctx, http.MethodGet, "auth/test", nil, nil, nil); err != nil {
		return fmt.Errorf("check onfleet: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: onfleetStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: an Onfleet stream starts
// with an empty incremental cursor (full sync). Onfleet's supported sync mode is
// full_refresh, so the cursor is informational only.
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
		stream = "tasks"
	}
	endpoint, ok := onfleetStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("onfleet stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := onfleetMaxPages(req.Config)
	if err != nil {
		return err
	}
	if endpoint.paginated {
		return c.harvestPaginated(ctx, r, endpoint, maxPages, emit)
	}
	return c.harvestArray(ctx, r, endpoint, emit)
}

// harvestArray reads a non-paginated Onfleet list endpoint that returns a
// top-level JSON array (workers, teams, hubs, admins).
func (c Connector) harvestArray(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read onfleet %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.arrayPath)
	if err != nil {
		return fmt.Errorf("decode onfleet %s: %w", endpoint.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// harvestPaginated drives Onfleet's lastId pagination. The tasks/all endpoint
// returns {lastId:"...", tasks:[...]}; the next page is requested with
// lastId=<lastId>. An empty/absent lastId ends pagination. This exact shape has
// no body-token paginator in connsdk, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvestPaginated(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, maxPages int, emit func(connectors.Record) error) error {
	lastID := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if lastID != "" {
			query.Set("lastId", lastID)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read onfleet %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.arrayPath)
		if err != nil {
			return fmt.Errorf("decode onfleet %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "lastId")
		if err != nil {
			return fmt.Errorf("decode onfleet %s lastId: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return nil
		}
		lastID = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise onfleet credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               fmt.Sprintf("%s_fixture_%d", stream, i),
			"shortId":          fmt.Sprintf("%02x", i),
			"state":            int64(i),
			"completed":        false,
			"trackingURL":      fmt.Sprintf("https://onf.lt/fixture%d", i),
			"worker":           "worker_fixture_1",
			"merchant":         "org_fixture_1",
			"executor":         "org_fixture_1",
			"creator":          "admin_fixture_1",
			"name":             fmt.Sprintf("Fixture %d", i),
			"phone":            "+15555550100",
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"type":             "super",
			"isActive":         true,
			"onDuty":           false,
			"activeTask":       "",
			"hub":              "hub_fixture_1",
			"address":          "1 Fixture St",
			"timeLastSeen":     onfleetFixtureCreated + int64(i),
			"timeCreated":      onfleetFixtureCreated + int64(i),
			"timeLastModified": onfleetFixtureCreated + int64(i),
			"connector":        "onfleet",
			"fixture":          true,
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

// requester builds a connsdk.Requester wired with Basic auth (api_key as
// username, blank password), the resolved base URL, and Onfleet's JSON headers.
// The secret only ever flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := onfleetBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	apiKey := onfleetAPIKey(cfg)
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("onfleet connector requires secret api_key")
	}
	// Onfleet uses HTTP Basic auth with the API key as the username and an empty
	// password. The catalog `password` secret is a documented placeholder that
	// Onfleet expects to be blank, so it is intentionally not used as the
	// password.
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(apiKey, ""),
		UserAgent: onfleetUserAgent,
	}, nil
}

func onfleetAPIKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// onfleetBaseURL resolves and validates the base URL. The default is
// onfleet.com/api/v2; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func onfleetBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return onfleetDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("onfleet config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("onfleet config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("onfleet config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func onfleetMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("onfleet config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("onfleet config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Onfleet is exposed read-only. Implementing the method
// satisfies the connectors.Connector interface without advertising the
// capability (Metadata reports Write=false).
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsFailed: len(records)}, connectors.ErrUnsupportedOperation
}
