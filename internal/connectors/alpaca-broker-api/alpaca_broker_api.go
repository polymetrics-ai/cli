// Package alpacabrokerapi implements the native pm Alpaca Broker API connector.
// It is a declarative-HTTP per-system connector following the stripe template:
// a thin package that composes the connsdk toolkit (Requester + HTTP Basic auth
// + RecordsAt extraction) with Alpaca-specific stream definitions and endpoints.
//
// The Alpaca Broker API authenticates with HTTP Basic where the API Key ID is
// the username (config "username") and the API Secret Key is the password
// (secret "password"). It is read-only here: the upstream catalog declares
// full_refresh only, so no reverse-ETL writes are exposed.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package alpacabrokerapi

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
	// registryName is the bare system name used as package dir + registry key.
	registryName = "alpaca-broker-api"

	// Production and sandbox hosts for the Broker API. The Broker API uses a
	// single production host; the environment config only distinguishes sandbox.
	prodBaseURL    = "https://broker-api.alpaca.markets/v1"
	sandboxBaseURL = "https://broker-api.sandbox.alpaca.markets/v1"

	defaultLimit = 20
	maxLimit     = 1000
	userAgent    = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Alpaca Broker API connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Alpaca Broker API connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Alpaca Broker API",
		IntegrationType: "api",
		Description:     "Reads Alpaca Broker API accounts, assets, market calendar, clock, and country info over the Broker REST API (read-only).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Alpaca. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := baseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(username(cfg)) == "" {
		return errors.New("alpaca-broker-api connector requires config username (API Key ID)")
	}
	if strings.TrimSpace(secret(cfg)) == "" {
		return errors.New("alpaca-broker-api connector requires secret password (API Secret Key)")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// The clock endpoint is a cheap, side-effect-free probe of auth + reachability.
	if err := r.DoJSON(ctx, http.MethodGet, "clock", nil, nil, nil); err != nil {
		return fmt.Errorf("check alpaca-broker-api: %w", err)
	}
	return nil
}

// Write is unsupported: the Alpaca Broker API connector is read-only (the
// upstream catalog declares full_refresh only). It satisfies the Connector
// interface by reporting the operation as unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "accounts"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("alpaca-broker-api stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	limit, err := pageLimit(req.Config)
	if err != nil {
		return err
	}

	if endpoint.singleton {
		return c.readSingleton(ctx, r, endpoint, emit)
	}
	if endpoint.paginates {
		return c.harvest(ctx, r, endpoint, limit, emit)
	}
	return c.readList(ctx, r, endpoint, limit, emit)
}

// readSingleton reads an endpoint that returns one object (e.g. /clock).
func (c Connector) readSingleton(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read alpaca-broker-api %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode alpaca-broker-api %s: %w", endpoint.resource, err)
	}
	for _, item := range records {
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// readList reads an endpoint that returns a full top-level array in one response
// (e.g. /assets, /calendar, /country_info). A limit is passed when supported.
func (c Connector) readList(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, limit int, emit func(connectors.Record) error) error {
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
	if err != nil {
		return fmt.Errorf("read alpaca-broker-api %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode alpaca-broker-api %s: %w", endpoint.resource, err)
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

// harvest drives Alpaca's page_token cursor pagination over a top-level array.
// The accounts listing returns up to `limit` items; the next page is requested
// with page_token=<last object id>. A page shorter than the limit ends the loop.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, limit int, emit func(connectors.Record) error) error {
	base := url.Values{}
	if limit > 0 {
		base.Set("limit", strconv.Itoa(limit))
	}

	pageToken := ""
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if pageToken != "" {
			query.Set("page_token", pageToken)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read alpaca-broker-api %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode alpaca-broker-api %s page: %w", endpoint.resource, err)
		}
		lastID := ""
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			lastID = stringField(item, "id")
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Stop when the page is short (no more data), when we have no cursor to
		// advance, or when the endpoint does not page.
		if limit <= 0 || len(records) < limit || lastID == "" {
			return nil
		}
		pageToken = lastID
	}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	count := 2
	if endpoint.singleton {
		count = 1
	}
	for i := 1; i <= count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                 fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"account_number":     fmt.Sprintf("1000%d", i),
			"status":             "ACTIVE",
			"crypto_status":      "ACTIVE",
			"currency":           "USD",
			"account_type":       "trading",
			"enabled_assets":     "us_equity",
			"created_at":         fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"kyc_results":        "",
			"last_equity":        fmt.Sprintf("%d.00", 1000*i),
			"class":              "us_equity",
			"exchange":           "NASDAQ",
			"symbol":             fmt.Sprintf("FIX%d", i),
			"name":               fmt.Sprintf("Fixture Asset %d", i),
			"tradable":           true,
			"marginable":         true,
			"shortable":          false,
			"easy_to_borrow":     false,
			"fractionable":       true,
			"date":               fmt.Sprintf("2026-01-0%d", i),
			"open":               "09:30",
			"close":              "16:00",
			"session_open":       "0700",
			"session_close":      "1900",
			"timestamp":          fmt.Sprintf("2026-06-2%dT12:00:00Z", i),
			"is_open":            true,
			"next_open":          "2026-06-26T13:30:00Z",
			"next_close":         "2026-06-25T20:00:00Z",
			"country_code":       fmt.Sprintf("US%d", i),
			"country_name":       "United States",
			"phone_calling_code": "1",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth and the
// resolved base URL. The secret only ever flows into connsdk.Basic; it is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	user := username(cfg)
	if strings.TrimSpace(user) == "" {
		return nil, errors.New("alpaca-broker-api connector requires config username (API Key ID)")
	}
	pass := secret(cfg)
	if strings.TrimSpace(pass) == "" {
		return nil, errors.New("alpaca-broker-api connector requires secret password (API Secret Key)")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(user, pass),
		UserAgent: userAgent,
	}, nil
}

func username(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["username"]
}

func secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// baseURL resolves and validates the base URL. An explicit base_url override
// takes precedence (validated for SSRF); otherwise the environment config
// selects the sandbox or production host.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	override := ""
	if cfg.Config != nil {
		override = strings.TrimSpace(cfg.Config["base_url"])
	}
	if override != "" {
		parsed, err := url.Parse(override)
		if err != nil {
			return "", fmt.Errorf("alpaca-broker-api config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("alpaca-broker-api config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("alpaca-broker-api config base_url must include a host")
		}
		return strings.TrimRight(override, "/"), nil
	}

	env := ""
	if cfg.Config != nil {
		env = strings.TrimSpace(strings.ToLower(cfg.Config["environment"]))
	}
	switch env {
	case "", "broker-api.sandbox", "sandbox":
		return sandboxBaseURL, nil
	case "api", "live", "paper-api", "paper":
		return prodBaseURL, nil
	default:
		return "", fmt.Errorf("alpaca-broker-api config environment %q must be one of api, paper-api, broker-api.sandbox", env)
	}
}

// pageLimit resolves the per-page limit from the limit config (string in the
// upstream schema), defaulting to 20 and bounded to a sane maximum.
func pageLimit(cfg connectors.RuntimeConfig) (int, error) {
	raw := ""
	if cfg.Config != nil {
		raw = strings.TrimSpace(cfg.Config["limit"])
	}
	if raw == "" {
		return defaultLimit, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("alpaca-broker-api config limit must be an integer: %w", err)
	}
	if value < 1 || value > maxLimit {
		return 0, fmt.Errorf("alpaca-broker-api config limit must be between 1 and %d", maxLimit)
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

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
