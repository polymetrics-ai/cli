// Package salesloft implements the native pm Salesloft connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package composing the connsdk toolkit (Requester + Bearer /
// OAuth2 refresh auth + RecordsAt extraction + cursor state) with Salesloft
// stream definitions, endpoints, and page-number pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Salesloft REST API v2 specifics:
//   - Base URL: https://api.salesloft.com/v2
//   - Auth: Authorization: Bearer <api_key> (API key) or an OAuth2 access token
//     refreshed from a refresh_token grant.
//   - List responses: {"data":[...],"metadata":{"paging":{"next_page":N|null}}}.
//   - Pagination: request the next page with ?page=N&per_page=100 until
//     metadata.paging.next_page is null.
package salesloft

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
	salesloftDefaultBaseURL  = "https://api.salesloft.com/v2"
	salesloftDefaultTokenURL = "https://accounts.salesloft.com/oauth/token"
	salesloftDefaultPageSize = 100
	salesloftMaxPageSize     = 100
	salesloftUserAgent       = "polymetrics-go-cli"
	// salesloftFixtureUpdated is the deterministic updated_at used by fixture
	// records (2026-01-01T00:00:00Z).
	salesloftFixtureUpdated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("salesloft", New)
}

// New returns the Salesloft connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Salesloft connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "salesloft" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "salesloft",
		DisplayName:     "Salesloft",
		IntegrationType: "api",
		Description:     "Reads Salesloft people, accounts, cadences, users, and emails through the Salesloft REST API v2.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Salesloft. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := salesloftBaseURL(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the users list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "users", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check salesloft: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: salesloftStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Salesloft stream starts
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
		stream = "people"
	}
	endpoint, ok := salesloftStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("salesloft stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := salesloftPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := salesloftMaxPages(req.Config)
	if err != nil {
		return err
	}
	updatedSince := incrementalLowerBound(req)
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, updatedSince, emit)
}

// harvest drives Salesloft's metadata.paging.next_page page-number pagination.
// Salesloft list responses are {data:[...], metadata:{paging:{next_page:N|null}}};
// the next page is requested with page=<next_page>. The shape (token lives in the
// body, not a Link header, and is a page number) is specific enough that the loop
// lives here, built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, updatedSince string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))
	// Salesloft supports server-side incremental filtering via updated_at[gte].
	if updatedSince != "" {
		base.Set("updated_at[gte]", updatedSince)
		base.Set("sort_by", "updated_at")
		base.Set("sort_direction", "ASC")
	}

	page := "1"
	for pageNum := 0; maxPages == 0 || pageNum < maxPages; pageNum++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", page)
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read salesloft %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode salesloft %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "metadata.paging.next_page")
		if err != nil {
			return fmt.Errorf("decode salesloft %s paging: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" || next == "0" {
			return nil
		}
		page = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise salesloft credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                  i,
			"email_address":       fmt.Sprintf("fixture+%d@example.com", i),
			"email":               fmt.Sprintf("fixture+%d@example.com", i),
			"first_name":          "Fixture",
			"last_name":           fmt.Sprintf("Person%d", i),
			"display_name":        fmt.Sprintf("Fixture Person%d", i),
			"name":                fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"title":               "Engineer",
			"phone":               "+15555550100",
			"person_company_name": "Fixture Co",
			"domain":              "fixture.example.com",
			"website":             "https://fixture.example.com",
			"industry":            "Software",
			"company_type":        "B2B",
			"city":                "Atlanta",
			"country":             "US",
			"team_cadence":        false,
			"shared":              true,
			"active":              true,
			"guid":                fmt.Sprintf("guid-%d", i),
			"time_zone":           "America/New_York",
			"subject":             fmt.Sprintf("Fixture email %d", i),
			"status":              "sent",
			"bounced":             false,
			"view_tracking":       true,
			"click_tracking":      true,
			"do_not_contact":      false,
			"account":             map[string]any{"id": 100 + i},
			"owner":               map[string]any{"id": 200 + i},
			"sent_at":             salesloftFixtureUpdated,
			"created_at":          salesloftFixtureUpdated,
			"updated_at":          salesloftFixtureUpdated,
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "salesloft"
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

// requester builds a connsdk.Requester wired with the resolved auth and base URL.
// Secrets only ever flow into the authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := salesloftBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	auth, err := c.authenticator(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: salesloftUserAgent,
	}, nil
}

// authenticator selects between API-key Bearer auth and OAuth2 refresh-token
// auth based on which secrets are present. API key takes precedence when set.
func (c Connector) authenticator(cfg connectors.RuntimeConfig) (connsdk.Authenticator, error) {
	if apiKey := salesloftSecret(cfg, "api_key"); strings.TrimSpace(apiKey) != "" {
		return connsdk.Bearer(apiKey), nil
	}
	// A standalone access_token (no refresh) is also accepted as a Bearer token.
	accessToken := salesloftSecret(cfg, "access_token")
	refreshToken := salesloftSecret(cfg, "refresh_token")
	clientID := salesloftSecret(cfg, "client_id")
	clientSecret := salesloftSecret(cfg, "client_secret")

	if strings.TrimSpace(refreshToken) != "" && strings.TrimSpace(clientID) != "" && strings.TrimSpace(clientSecret) != "" {
		return &oauthRefreshAuth{
			tokenURL:     salesloftTokenURL(cfg),
			clientID:     clientID,
			clientSecret: clientSecret,
			refreshToken: refreshToken,
			seedToken:    strings.TrimSpace(accessToken),
			client:       c.Client,
		}, nil
	}
	if strings.TrimSpace(accessToken) != "" {
		return connsdk.Bearer(accessToken), nil
	}
	return nil, errors.New("salesloft connector requires secret api_key, or access_token, or client_id+client_secret+refresh_token")
}

// incrementalLowerBound returns the RFC3339 lower bound for updated_at[gte],
// derived from the incremental cursor (if any) or else the start_date config. An
// empty result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

// salesloftSecret resolves a secret. It accepts both the flat key (e.g.
// "api_key") and the dotted catalog form (e.g. "credentials.api_key") since the
// catalog declares secret fields under the credentials object.
func salesloftSecret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	if v, ok := cfg.Secrets[key]; ok {
		return v
	}
	return cfg.Secrets["credentials."+key]
}

// salesloftBaseURL resolves and validates the base URL. The default is
// api.salesloft.com/v2; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func salesloftBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return salesloftDefaultBaseURL, nil
	}
	if err := validateURL(base, "base_url"); err != nil {
		return "", err
	}
	return strings.TrimRight(base, "/"), nil
}

// salesloftTokenURL resolves the OAuth token endpoint, validating any override.
func salesloftTokenURL(cfg connectors.RuntimeConfig) string {
	tokenURL := strings.TrimSpace(cfg.Config["token_url"])
	if tokenURL == "" {
		return salesloftDefaultTokenURL
	}
	if err := validateURL(tokenURL, "token_url"); err != nil {
		// Fall back to the default rather than returning a partial URL; the
		// token request will surface a clear error if the default is wrong.
		return salesloftDefaultTokenURL
	}
	return tokenURL
}

func validateURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("salesloft config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("salesloft config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("salesloft config %s must include a host", field)
	}
	return nil
}

func salesloftPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return salesloftDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("salesloft config page_size must be an integer: %w", err)
	}
	if value < 1 || value > salesloftMaxPageSize {
		return 0, fmt.Errorf("salesloft config page_size must be between 1 and %d", salesloftMaxPageSize)
	}
	return value, nil
}

func salesloftMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("salesloft config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("salesloft config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// Write is unsupported: Salesloft is read-only in this connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
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
