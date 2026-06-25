// Package auth0 implements the native pm Auth0 connector. It is a
// declarative-HTTP per-system connector: a thin package that composes the
// connsdk toolkit (Requester + Bearer / OAuth2 client-credentials auth +
// RecordsAt extraction) with Auth0 Management API v2 stream definitions,
// endpoints, and page-based pagination. It follows the stripe reference shape.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// Auth0 is a read-only source (the upstream Airbyte connector supports only
// full-refresh/incremental reads), so Capabilities.Write is false and there is
// no write.go.
package auth0

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
	// auth0DefaultPageSize is the Public Cloud per_page maximum.
	auth0DefaultPageSize = 50
	auth0MaxPageSize     = 100
	auth0UserAgent       = "polymetrics-go-cli"
	// auth0FixtureCreated is the deterministic created_at used by fixture records.
	auth0FixtureCreated = "2026-01-01T00:00:00.000Z"
)

func init() {
	connectors.RegisterFactory("auth0", New)
}

// New returns the Auth0 connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Auth0 connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "auth0" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "auth0",
		DisplayName:     "Auth0",
		IntegrationType: "api",
		Description:     "Reads Auth0 users, clients, connections, roles, and organizations from the Auth0 Management API v2.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Auth0. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := auth0BaseURL(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the clients list confirms auth and connectivity without
	// reading PII-heavy user records.
	query := url.Values{"per_page": []string{"1"}, "include_totals": []string{"true"}, "page": []string{"0"}}
	if err := r.DoJSON(ctx, http.MethodGet, "api/v2/clients", query, nil, nil); err != nil {
		return fmt.Errorf("check auth0: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Auth0 is a read-only
// source, so writes are unsupported (Capabilities.Write is false).
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: auth0Streams()}, nil
}

// InitialState satisfies connectors.StatefulReader: an Auth0 stream starts with
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
		stream = "users"
	}
	endpoint, ok := auth0StreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("auth0 stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := auth0PageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := auth0MaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Auth0's zero-indexed page/per_page pagination. List endpoints
// return an include_totals envelope {start,limit,length,total,<resource>:[...]}.
// The loop advances pages until a short page is returned or the total is
// reached. There is no body-token paginator in connsdk for this resource-named
// array shape, so the loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))
		query.Set("include_totals", "true")

		resp, err := r.Do(ctx, http.MethodGet, "api/v2/"+endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read auth0 %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.arrayKey)
		if err != nil {
			return fmt.Errorf("decode auth0 %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer than per_page) means we have reached the end. This
		// is robust whether or not the server honors include_totals.
		if len(records) < pageSize {
			return nil
		}
		// Defensive bound against runaway pagination using the reported total.
		if total, ok := totalRecords(resp.Body); ok && (page+1)*pageSize >= total {
			return nil
		}
	}
	return nil
}

// totalRecords reads the include_totals "total" field, if present and numeric.
func totalRecords(body []byte) (int, bool) {
	raw, err := connsdk.StringAt(body, "total")
	if err != nil || raw == "" {
		return 0, false
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return n, true
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise auth0 credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		idValue := fmt.Sprintf("%s_fixture_%d", endpoint.resource, i)
		item := map[string]any{
			"user_id":              "auth0|" + idValue,
			"client_id":            idValue,
			"id":                   idValue,
			"email":                fmt.Sprintf("fixture+%d@example.com", i),
			"email_verified":       true,
			"username":             fmt.Sprintf("fixture_%d", i),
			"name":                 fmt.Sprintf("Fixture %d", i),
			"nickname":             fmt.Sprintf("fix%d", i),
			"given_name":           "Fixture",
			"family_name":          strconv.Itoa(i),
			"picture":              "https://example.com/avatar.png",
			"created_at":           auth0FixtureCreated,
			"updated_at":           auth0FixtureCreated,
			"last_login":           auth0FixtureCreated,
			"logins_count":         int64(i),
			"blocked":              false,
			"description":          "fixture record",
			"app_type":             "spa",
			"is_first_party":       true,
			"oidc_conformant":      true,
			"global":               false,
			"display_name":         fmt.Sprintf("Fixture %d", i),
			"strategy":             "auth0",
			"is_domain_connection": false,
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

// requester builds a connsdk.Requester wired with the resolved base URL and the
// appropriate authenticator. If an access_token secret is supplied it is used as
// a bearer directly; otherwise client_id/client_secret + audience drive the
// OAuth2 client-credentials grant against the tenant's /oauth/token endpoint.
// Secrets only ever flow into the authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := auth0BaseURL(cfg)
	if err != nil {
		return nil, err
	}
	auth, err := c.authenticator(cfg, base)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: auth0UserAgent,
	}, nil
}

// authenticator selects between the static access-token and the M2M
// client-credentials auth modes based on which secrets are present.
func (c Connector) authenticator(cfg connectors.RuntimeConfig, base string) (connsdk.Authenticator, error) {
	if token := secret(cfg, "access_token"); strings.TrimSpace(token) != "" {
		return connsdk.Bearer(token), nil
	}
	clientID := strings.TrimSpace(secret(cfg, "client_id"))
	if clientID == "" {
		clientID = strings.TrimSpace(cfg.Config["client_id"])
	}
	clientSecret := strings.TrimSpace(secret(cfg, "client_secret"))
	if clientID == "" || clientSecret == "" {
		return nil, errors.New("auth0 connector requires either secret access_token or client_id + client_secret")
	}
	audience := strings.TrimSpace(cfg.Config["audience"])
	if audience == "" {
		// Default to the Management API audience for the tenant.
		audience = strings.TrimRight(base, "/") + "/api/v2/"
	}
	return &connsdk.OAuth2ClientCredentials{
		TokenURL:     strings.TrimRight(base, "/") + "/oauth/token",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Client:       c.Client,
		ExtraParams:  url.Values{"audience": []string{audience}},
	}, nil
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

// auth0BaseURL resolves and validates the base URL. There is no fixed default
// (the base is the tenant domain, e.g. https://dev-org.us.auth0.com); any value
// must be an absolute https (or http for local test servers) URL with a host to
// bound SSRF risk.
func auth0BaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return "", errors.New("auth0 connector requires config base_url (your tenant domain)")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("auth0 config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("auth0 config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("auth0 config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func auth0PageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return auth0DefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("auth0 config page_size must be an integer: %w", err)
	}
	if value < 1 || value > auth0MaxPageSize {
		return 0, fmt.Errorf("auth0 config page_size must be between 1 and %d", auth0MaxPageSize)
	}
	return value, nil
}

func auth0MaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("auth0 config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("auth0 config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
