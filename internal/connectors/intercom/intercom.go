// Package intercom implements the native pm Intercom source connector. It is a
// declarative-HTTP per-system connector built in the shape of the stripe
// reference: a thin package that composes the connsdk toolkit (Requester +
// Bearer auth + RecordsAt extraction + cursor state) with Intercom-specific
// stream definitions, endpoints, and pagination.
//
// Intercom list endpoints return {"type":"list","data":[...],"pages":{"next":
// {"starting_after":"..."}}}; the next page is requested with
// starting_after=<cursor>. There is no body-token paginator in connsdk for this
// exact shape, so the loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
//
// Like stripe/github, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package intercom

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
	intercomDefaultBaseURL  = "https://api.intercom.io"
	intercomDefaultPageSize = 50
	intercomMaxPageSize     = 150
	intercomUserAgent       = "polymetrics-go-cli"
	// intercomFixtureCreated is the deterministic created_at timestamp used by
	// the fixture-mode records (2026-01-01T00:00:00Z in unix seconds).
	intercomFixtureCreated int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("intercom", New)
}

// New returns the Intercom connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Intercom source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "intercom" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "intercom",
		DisplayName:     "Intercom",
		IntegrationType: "api",
		Description:     "Reads Intercom contacts, companies, conversations, admins, and tags through the Intercom REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Intercom. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := intercomBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(intercomSecret(cfg)) == "" {
		return errors.New("intercom connector requires secret access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the admins list confirms auth and connectivity without
	// mutating anything; the /me endpoint also works but admins is a plain list.
	if err := r.DoJSON(ctx, http.MethodGet, "admins", nil, nil, nil); err != nil {
		return fmt.Errorf("check intercom: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: intercomStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: an Intercom stream starts
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
		stream = "contacts"
	}
	endpoint, ok := intercomStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("intercom stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := intercomPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := intercomMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Intercom's cursor pagination. List endpoints return
// {data:[...], pages:{next:{starting_after:<cursor>}}}; the next page is
// requested with starting_after=<cursor>. When the response has no
// pages.next.starting_after the harvest stops. Endpoints that return a single
// page without a pages object (e.g. admins, tags) terminate after one request.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))

	startingAfter := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if startingAfter != "" {
			query.Set("starting_after", startingAfter)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read intercom %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode intercom %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "pages.next.starting_after")
		if err != nil {
			return fmt.Errorf("decode intercom %s pages: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return nil
		}
		startingAfter = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise intercom credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                       fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"type":                     strings.TrimSuffix(stream, "s"),
			"created_at":               intercomFixtureCreated + int64(i),
			"updated_at":               intercomFixtureCreated + int64(i) + 100,
			"email":                    fmt.Sprintf("fixture+%d@example.com", i),
			"name":                     fmt.Sprintf("Fixture %d", i),
			"company_id":               "comp_fixture_1",
			"external_id":              fmt.Sprintf("ext_%d", i),
			"state":                    "open",
			"open":                     true,
			"read":                     false,
			"job_title":                "Support",
			"away_mode_enabled":        false,
			"unsubscribed_from_emails": false,
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

// requester builds a connsdk.Requester wired with Bearer auth, the resolved base
// URL, and the Intercom-Version header. The secret only ever flows into
// connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := intercomBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := intercomSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("intercom connector requires secret access_token")
	}
	headers := map[string]string{}
	if version := strings.TrimSpace(cfg.Config["api_version"]); version != "" {
		headers["Intercom-Version"] = version
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           connsdk.Bearer(secret),
		UserAgent:      intercomUserAgent,
		DefaultHeaders: headers,
	}, nil
}

func intercomSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["access_token"]
}

// intercomBaseURL resolves and validates the base URL. The default is
// api.intercom.io; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func intercomBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return intercomDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("intercom config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("intercom config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("intercom config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func intercomPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return intercomDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("intercom config page_size must be an integer: %w", err)
	}
	if value < 1 || value > intercomMaxPageSize {
		return 0, fmt.Errorf("intercom config page_size must be between 1 and %d", intercomMaxPageSize)
	}
	return value, nil
}

func intercomMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("intercom config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("intercom config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}

// Write satisfies the connectors.Connector interface. Intercom is exposed as a
// read-only source here, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
