// Package chift implements the native pm Chift connector. It follows the stripe
// declarative-HTTP template: a thin package that composes the connsdk toolkit
// (Requester + extraction) with Chift-specific stream definitions, endpoints,
// and a session-token authenticator.
//
// Chift authenticates with a session-token exchange rather than a standard
// form-encoded OAuth2 client-credentials grant: a POST to /token with a JSON
// body of {clientId, clientSecret, accountId} returns {access_token}, which is
// then carried as a Bearer token on data requests. connsdk's
// OAuth2ClientCredentials posts a form with different keys, so a small
// in-package authenticator (sessionTokenAuth) implements connsdk.Authenticator.
//
// Like github and stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package chift

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
	chiftDefaultBaseURL  = "https://api.chift.eu"
	chiftDefaultPageSize = 100
	chiftMaxPageSize     = 1000
	chiftUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("chift", New)
}

// New returns the Chift connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Chift connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the session-token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "chift" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "chift",
		DisplayName:     "Chift",
		IntegrationType: "api",
		Description:     "Reads Chift consumers, connections, and syncs through the Chift REST API using a session-token (client credentials) exchange.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Chift. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := chiftBaseURL(cfg); err != nil {
		return err
	}
	creds := resolveCredentials(cfg)
	if err := creds.validate(); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the consumers list confirms the token exchange, auth,
	// and connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "consumers", url.Values{"size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check chift: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: chiftStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "consumers"
	}
	endpoint, ok := chiftStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("chift stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := chiftPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := chiftMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// Write is unsupported: Chift is a read-only source connector. The method
// exists to satisfy connectors.Connector; Capabilities.Write is false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Chift's offset/limit pagination. Chift list endpoints return a
// top-level JSON array; the next page is requested by advancing the offset by
// the page size until a short (less-than-full) page is returned. Built on
// connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("size", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read chift %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode chift %s page: %w", endpoint.resource, err)
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
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise chift credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		id := fmt.Sprintf("%s_fixture_%d", endpoint.resource, i)
		item := map[string]any{
			"consumerid":   id,
			"connectionid": id,
			"syncid":       id,
			"name":         fmt.Sprintf("Fixture %d", i),
			"email":        fmt.Sprintf("fixture+%d@example.com", i),
			"phone":        "",
			"redirect_url": "https://example.com/redirect",
			"api":          "accounting",
			"status":       "active",
			"active":       true,
			"created_on":   "2026-01-01T00:00:00Z",
			"updated_on":   "2026-01-01T00:00:00Z",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the session-token
// authenticator and the resolved base URL. Credentials only ever flow into the
// authenticator's token exchange; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := chiftBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	creds := resolveCredentials(cfg)
	if err := creds.validate(); err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      newSessionTokenAuth(base, creds, c.Client),
		UserAgent: chiftUserAgent,
	}, nil
}

// chiftCredentials carries the three required secrets resolved from cfg.Secrets.
type chiftCredentials struct {
	clientID     string
	clientSecret string
	accountID    string
}

func (cr chiftCredentials) validate() error {
	var missing []string
	if cr.clientID == "" {
		missing = append(missing, "client_id")
	}
	if cr.clientSecret == "" {
		missing = append(missing, "client_secret")
	}
	if cr.accountID == "" {
		missing = append(missing, "account_id")
	}
	if len(missing) > 0 {
		return fmt.Errorf("chift connector requires secret(s): %s", strings.Join(missing, ", "))
	}
	return nil
}

func resolveCredentials(cfg connectors.RuntimeConfig) chiftCredentials {
	if cfg.Secrets == nil {
		return chiftCredentials{}
	}
	return chiftCredentials{
		clientID:     strings.TrimSpace(cfg.Secrets["client_id"]),
		clientSecret: strings.TrimSpace(cfg.Secrets["client_secret"]),
		accountID:    strings.TrimSpace(cfg.Secrets["account_id"]),
	}
}

// chiftBaseURL resolves and validates the base URL. The default is
// api.chift.eu; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func chiftBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return chiftDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("chift config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("chift config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("chift config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func chiftPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return chiftDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("chift config page_size must be an integer: %w", err)
	}
	if value < 1 || value > chiftMaxPageSize {
		return 0, fmt.Errorf("chift config page_size must be between 1 and %d", chiftMaxPageSize)
	}
	return value, nil
}

func chiftMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("chift config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("chift config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
