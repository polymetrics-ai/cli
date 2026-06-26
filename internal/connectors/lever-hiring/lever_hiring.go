// Package leverhiring implements the native pm Lever Hiring connector. It is a
// declarative-HTTP per-system connector following the stripe/acuity-scheduling
// template: a thin package that composes the connsdk toolkit (Requester + auth +
// RecordsAt extraction + cursor state) with Lever-specific stream definitions,
// endpoints, and the Lever Data API's hasNext/next offset pagination.
//
// The Lever Data API (https://api.lever.co/v1) authenticates either with an API
// key supplied as the HTTP Basic auth username (password blank), or with an
// OAuth bearer access token. List endpoints return a JSON envelope
// {"data":[...],"next":"<cursor>","hasNext":bool}; the next page is requested
// with offset=<next>. The Lever source is read-only (full-refresh); it exposes
// no reverse-ETL writes, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package leverhiring

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
	registryName          = "lever-hiring"
	leverProductionURL    = "https://api.lever.co/v1"
	leverSandboxURL       = "https://api.sandbox.lever.co/v1"
	leverDefaultPageSize  = 100
	leverMaxPageSize      = 100
	leverUserAgent        = "polymetrics-go-cli"
	leverFixtureCreatedMs = int64(1767225600000) // 2026-01-01T00:00:00Z in unix millis
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Lever Hiring connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Lever Hiring connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Lever Hiring",
		IntegrationType: "api",
		Description:     "Reads Lever Hiring opportunities, postings, users, requisitions, and stages through the Lever Data API. Read-only (full-refresh).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Lever. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := leverBaseURL(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the postings list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "postings", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check lever-hiring: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: leverStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Lever Hiring is read-only
// (no reverse-ETL writes), so it always reports the operation as unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a Lever stream starts with an
// empty incremental cursor (full sync).
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
		stream = "opportunities"
	}
	endpoint, ok := leverStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("lever-hiring stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := leverPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := leverMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Lever's hasNext/next offset pagination. List responses return
// {data:[...], next:"<cursor>", hasNext:bool}; the next page is requested with
// offset=<next>. There is no body-token paginator in connsdk that consumes both
// hasNext and an offset param, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))

	offset := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if offset != "" {
			query.Set("offset", offset)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read lever-hiring %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode lever-hiring %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		hasNext, err := connsdk.StringAt(resp.Body, "hasNext")
		if err != nil {
			return fmt.Errorf("decode lever-hiring %s hasNext: %w", endpoint.resource, err)
		}
		next, err := connsdk.StringAt(resp.Body, "next")
		if err != nil {
			return fmt.Errorf("decode lever-hiring %s next: %w", endpoint.resource, err)
		}
		if hasNext != "true" || strings.TrimSpace(next) == "" {
			return nil
		}
		offset = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise lever-hiring credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":              fmt.Sprintf("Fixture %d", i),
			"text":              fmt.Sprintf("Fixture %s %d", stream, i),
			"headline":          fmt.Sprintf("Headline %d", i),
			"stage":             "stage_fixture_1",
			"state":             "published",
			"status":            "open",
			"origin":            "sourced",
			"username":          fmt.Sprintf("fixture%d", i),
			"email":             fmt.Sprintf("fixture+%d@example.com", i),
			"accessRole":        "team member",
			"requisitionCode":   fmt.Sprintf("REQ-%d", i),
			"headcountTotal":    int64(i),
			"headcountHired":    int64(0),
			"owner":             "user_fixture_1",
			"user":              "user_fixture_1",
			"hiringManager":     "user_fixture_1",
			"sources":           []any{"LinkedIn"},
			"tags":              []any{"fixture"},
			"emails":            []any{fmt.Sprintf("fixture+%d@example.com", i)},
			"categories":        map[string]any{"team": "Engineering"},
			"createdAt":         leverFixtureCreatedMs + int64(i),
			"updatedAt":         leverFixtureCreatedMs + int64(i),
			"lastInteractionAt": leverFixtureCreatedMs + int64(i),
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

// requester builds a connsdk.Requester wired with Lever auth and the resolved
// base URL. When an OAuth access token is supplied it uses Bearer auth; otherwise
// it uses HTTP Basic auth with the API key as the username and an empty password.
// The secret only ever flows into connsdk.Bearer/connsdk.Basic; it is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := leverBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	auth, err := leverAuth(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: leverUserAgent,
	}, nil
}

// leverAuth selects the authenticator from the resolved secrets. An access token
// (OAuth) takes precedence over an API key (Basic). The secret keys accept both
// the bare name and the catalog's dotted "credentials." prefix.
func leverAuth(cfg connectors.RuntimeConfig) (connsdk.Authenticator, error) {
	if token := leverSecret(cfg, "access_token", "credentials.access_token"); token != "" {
		return connsdk.Bearer(token), nil
	}
	if apiKey := leverSecret(cfg, "api_key", "credentials.api_key"); apiKey != "" {
		// Lever Basic auth: API key as username, blank password.
		return connsdk.Basic(apiKey, ""), nil
	}
	return nil, errors.New("lever-hiring connector requires secret api_key or access_token")
}

// leverSecret returns the first non-empty secret among the provided keys.
func leverSecret(cfg connectors.RuntimeConfig, keys ...string) string {
	if cfg.Secrets == nil {
		return ""
	}
	for _, k := range keys {
		if v := strings.TrimSpace(cfg.Secrets[k]); v != "" {
			return v
		}
	}
	return ""
}

// leverBaseURL resolves and validates the base URL. The default is
// api.lever.co (or api.sandbox.lever.co when environment=Sandbox); any override
// must be an absolute https (or http for local test servers) URL with a host to
// bound SSRF risk.
func leverBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		if strings.EqualFold(strings.TrimSpace(cfg.Config["environment"]), "sandbox") {
			return leverSandboxURL, nil
		}
		return leverProductionURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("lever-hiring config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("lever-hiring config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("lever-hiring config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func leverPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return leverDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("lever-hiring config page_size must be an integer: %w", err)
	}
	if value < 1 || value > leverMaxPageSize {
		return 0, fmt.Errorf("lever-hiring config page_size must be between 1 and %d", leverMaxPageSize)
	}
	return value, nil
}

func leverMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("lever-hiring config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("lever-hiring config max_pages must be 0 for unlimited or a positive integer")
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
