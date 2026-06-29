// Package dwolla implements the native pm Dwolla source connector. It follows
// the declarative-HTTP template established by the stripe package: a thin
// package composing the connsdk toolkit (Requester + OAuth2 client-credentials
// auth + HAL _embedded extraction + cursor state) with Dwolla-specific stream
// definitions and endpoints.
//
// Dwolla is a HAL+JSON API. List endpoints nest their records under
// _embedded.<key> and advertise the next page as an absolute URL at
// _links.next.href, so pagination is a small in-package loop (mirroring stripe's
// harvest) rather than one of the generic connsdk paginators.
//
// Authentication is the OAuth2 client-credentials grant against
// https://<environment>.dwolla.com/token, where environment is "api" (production)
// or "api-sandbox". connsdk.OAuth2ClientCredentials handles minting, caching, and
// refreshing the bearer token.
//
// The connector is read-only: Dwolla is an upstream source with no reverse-ETL
// surface, so Write returns ErrUnsupportedOperation and Capabilities.Write=false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package dwolla

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
	dwollaProdBaseURL     = "https://api.dwolla.com"
	dwollaSandboxBaseURL  = "https://api-sandbox.dwolla.com"
	dwollaDefaultPageSize = 25
	dwollaMaxPageSize     = 200
	dwollaUserAgent       = "polymetrics-go-cli"
	dwollaHALAccept       = "application/vnd.dwolla.v1.hal+json"
	// dwollaFixtureCreated is the deterministic `created` timestamp used by the
	// fixture-mode records.
	dwollaFixtureCreated = "2026-01-01T00:00:00.000Z"
)

func init() {
	connectors.RegisterFactory("dwolla", New)
}

// New returns the Dwolla connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Dwolla source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth2 token fetcher. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "dwolla" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "dwolla",
		DisplayName:     "Dwolla",
		IntegrationType: "api",
		Description:     "Reads Dwolla customers, events, exchange partners, and business classifications via the Dwolla HAL+JSON REST API using OAuth2 client-credentials.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Dwolla. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := dwollaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(dwollaClientID(cfg)) == "" || strings.TrimSpace(dwollaClientSecret(cfg)) == "" {
		return errors.New("dwolla connector requires secrets client_id and client_secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the customers list confirms auth (token mint) and
	// connectivity without mutating anything.
	q := url.Values{"limit": []string{"1"}}
	if _, err := r.Do(ctx, http.MethodGet, "customers", q, nil); err != nil {
		return fmt.Errorf("check dwolla: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: dwollaStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Dwolla stream starts with
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
		stream = "customers"
	}
	endpoint, ok := dwollaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("dwolla stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := dwollaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := dwollaMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Dwolla's HAL pagination. List responses nest records under
// _embedded.<embedKey> and advertise the next page as an absolute URL at
// _links.next.href. There is no body-URL paginator in connsdk for this exact
// shape, so the loop lives here, built on connsdk.Requester + connsdk.RecordsAt +
// connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	// recordsPath uses dotted notation; embed keys may contain hyphens (valid in
	// a JSON object key, and selectPath splits only on ".").
	recordsPath := "_embedded." + endpoint.embedKey

	// The first request is the relative resource path with a page-size limit; on
	// subsequent pages Dwolla hands back a fully-qualified next URL which the
	// Requester treats as absolute.
	nextURL := ""
	firstQuery := url.Values{"limit": []string{strconv.Itoa(pageSize)}}

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var (
			resp *connsdk.Response
			err  error
		)
		if nextURL == "" {
			resp, err = r.Do(ctx, http.MethodGet, endpoint.resource, firstQuery, nil)
		} else {
			resp, err = r.Do(ctx, http.MethodGet, nextURL, nil, nil)
		}
		if err != nil {
			return fmt.Errorf("read dwolla %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, recordsPath)
		if err != nil {
			return fmt.Errorf("decode dwolla %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "_links.next.href")
		if err != nil {
			return fmt.Errorf("decode dwolla %s next link: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return nil
		}
		nextURL = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise dwolla credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"firstName":    fmt.Sprintf("Fixture%d", i),
			"lastName":     "Example",
			"email":        fmt.Sprintf("fixture+%d@example.com", i),
			"type":         "personal",
			"status":       "verified",
			"businessName": "",
			"name":         fmt.Sprintf("Fixture %s %d", stream, i),
			"topic":        "customer_created",
			"resourceId":   fmt.Sprintf("res_%d", i),
			"created":      dwollaFixtureCreated,
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

// Write is unsupported: Dwolla is a read-only source connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// requester builds a connsdk.Requester wired with OAuth2 client-credentials auth,
// the resolved base URL, and the Dwolla HAL Accept header. The secret only ever
// flows into the authenticator; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := dwollaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	clientID := strings.TrimSpace(dwollaClientID(cfg))
	clientSecret := strings.TrimSpace(dwollaClientSecret(cfg))
	if clientID == "" || clientSecret == "" {
		return nil, errors.New("dwolla connector requires secrets client_id and client_secret")
	}
	tokenURL, err := dwollaTokenURL(base)
	if err != nil {
		return nil, err
	}
	auth := &connsdk.OAuth2ClientCredentials{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: dwollaUserAgent,
		Accept:    dwollaHALAccept,
	}, nil
}

func dwollaClientID(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["client_id"]
}

func dwollaClientSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["client_secret"]
}

// dwollaBaseURL resolves and validates the base URL. An explicit base_url config
// override wins (validated for scheme+host to bound SSRF risk); otherwise the
// environment config selects production ("api") or sandbox ("api-sandbox").
func dwollaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
		parsed, err := url.Parse(base)
		if err != nil {
			return "", fmt.Errorf("dwolla config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("dwolla config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("dwolla config base_url must include a host")
		}
		return strings.TrimRight(base, "/"), nil
	}

	switch strings.TrimSpace(strings.ToLower(cfg.Config["environment"])) {
	case "", "api", "production", "prod":
		return dwollaProdBaseURL, nil
	case "api-sandbox", "sandbox":
		return dwollaSandboxBaseURL, nil
	default:
		return "", fmt.Errorf("dwolla config environment must be 'api' or 'api-sandbox', got %q", cfg.Config["environment"])
	}
}

// dwollaTokenURL derives the OAuth2 token endpoint from the resolved base URL.
// Dwolla hosts the token endpoint at <scheme>://<host>/token on the same host as
// the API, so test servers (which share one host for token + data) work too.
func dwollaTokenURL(base string) (string, error) {
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("dwolla base url is invalid: %w", err)
	}
	parsed.Path = "/token"
	parsed.RawQuery = ""
	return parsed.String(), nil
}

func dwollaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return dwollaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("dwolla config page_size must be an integer: %w", err)
	}
	if value < 1 || value > dwollaMaxPageSize {
		return 0, fmt.Errorf("dwolla config page_size must be between 1 and %d", dwollaMaxPageSize)
	}
	return value, nil
}

func dwollaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("dwolla config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("dwolla config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
