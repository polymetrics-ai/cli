// Package zendesksupport implements the native pm Zendesk Support connector. It
// follows the declarative-HTTP shape established by the stripe connector: a thin
// package that composes the connsdk toolkit (Requester + extraction helpers +
// cursor state) with Zendesk-specific stream definitions, endpoints, and auth.
//
// Zendesk has two notable traits versus a vanilla bearer API:
//   - Auth has two supported modes. The Airbyte Open Source default is an API
//     token sent via HTTP Basic as "<email>/token:<api_token>". The OAuth modes
//     send an access_token via Authorization: Bearer. This connector resolves
//     whichever secret is present, preferring the OAuth access_token.
//   - Most collection endpoints use cursor pagination: page[size] sets the page
//     size and page[after] carries the cursor; the response body reports the next
//     cursor at meta.after_cursor and whether more pages exist at meta.has_more.
//     There is no connsdk paginator for this exact shape, so the loop lives here,
//     built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect. The registry key is the bare system name
// "zendesk-support" even though the Go package identifier is zendesksupport.
package zendesksupport

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
	registryName    = "zendesk-support"
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Zendesk Support connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Zendesk Support connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Zendesk Support",
		IntegrationType: "api",
		Description:     "Reads Zendesk Support tickets, users, organizations, groups, and satisfaction ratings through the Zendesk Support REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

// Check verifies the connector is configured well enough to talk to Zendesk. In
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
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the groups list confirms auth and connectivity without
	// mutating anything (groups is small on every account).
	if err := r.DoJSON(ctx, http.MethodGet, "groups", url.Values{"page[size]": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check zendesk-support: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Zendesk stream starts with
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
		stream = "tickets"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("zendesk-support stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	size, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	pages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, size, pages, emit)
}

// Write is unsupported: the Zendesk Support connector is read-only (no
// allow-listed reverse-ETL actions), so Capabilities.Write is false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Zendesk's cursor pagination. Collection responses look like
// {"<key>":[...], "meta":{"has_more":bool, "after_cursor":"..."}}; the next page
// is requested with page[after]=<after_cursor>. The loop stops when has_more is
// false or no cursor is returned.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("page[size]", strconv.Itoa(pageSize))

	after := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if after != "" {
			query.Set("page[after]", after)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read zendesk-support %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode zendesk-support %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		hasMore, err := connsdk.StringAt(resp.Body, "meta.has_more")
		if err != nil {
			return fmt.Errorf("decode zendesk-support %s has_more: %w", endpoint.resource, err)
		}
		nextCursor, err := connsdk.StringAt(resp.Body, "meta.after_cursor")
		if err != nil {
			return fmt.Errorf("decode zendesk-support %s after_cursor: %w", endpoint.resource, err)
		}
		if hasMore != "true" || strings.TrimSpace(nextCursor) == "" {
			return nil
		}
		after = nextCursor
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise zendesk-support credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":              int64(i),
			"subject":         fmt.Sprintf("Fixture ticket %d", i),
			"description":     fmt.Sprintf("Fixture %s %d", stream, i),
			"status":          "open",
			"priority":        "normal",
			"type":            "question",
			"name":            fmt.Sprintf("Fixture %d", i),
			"email":           fmt.Sprintf("fixture+%d@example.com", i),
			"role":            "end-user",
			"active":          true,
			"verified":        true,
			"details":         "fixture org",
			"notes":           "fixture",
			"default":         false,
			"deleted":         false,
			"score":           "good",
			"comment":         "thanks",
			"reason":          "no_reason",
			"ticket_id":       int64(100 + i),
			"requester_id":    int64(200 + i),
			"assignee_id":     int64(300 + i),
			"organization_id": int64(400 + i),
			"group_id":        int64(500 + i),
			"brand_id":        int64(600 + i),
			"shared_tickets":  false,
			"shared_comments": false,
			"time_zone":       "UTC",
			"phone":           "",
			"created_at":      "2026-01-01T00:00:00Z",
			"updated_at":      fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"connector":       registryName,
			"fixture":         true,
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

// requester builds a connsdk.Requester wired with the resolved auth and base URL.
// Secrets only ever flow into the connsdk authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	auth, err := authenticator(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: userAgent,
	}, nil
}

// authenticator resolves the Zendesk credential. An OAuth access_token is sent as
// Bearer; otherwise an API token is sent via HTTP Basic as
// "<email>/token:<api_token>". The secret is never logged.
func authenticator(cfg connectors.RuntimeConfig) (connsdk.Authenticator, error) {
	if token := secret(cfg, "credentials.access_token", "access_token"); token != "" {
		return connsdk.Bearer(token), nil
	}
	apiToken := secret(cfg, "credentials.api_token", "api_token")
	if apiToken != "" {
		email := strings.TrimSpace(firstNonEmpty(
			secret(cfg, "credentials.email", "email"),
			cfg.Config["email"],
		))
		if email == "" {
			return nil, errors.New("zendesk-support API token auth requires credentials.email")
		}
		return connsdk.Basic(email+"/token", apiToken), nil
	}
	return nil, errors.New("zendesk-support requires a secret: credentials.access_token or credentials.api_token (+ email)")
}

// baseURL resolves and validates the base URL. By default it is derived from the
// subdomain config as https://<subdomain>.zendesk.com/api/v2. Any base_url
// override must be an absolute https (or http for local test servers) URL with a
// host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	override := strings.TrimSpace(cfg.Config["base_url"])
	if override != "" {
		parsed, err := url.Parse(override)
		if err != nil {
			return "", fmt.Errorf("zendesk-support config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("zendesk-support config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("zendesk-support config base_url must include a host")
		}
		return strings.TrimRight(override, "/") + "/api/v2", nil
	}
	subdomain := strings.TrimSpace(cfg.Config["subdomain"])
	if subdomain == "" {
		return "", errors.New("zendesk-support requires config subdomain (or base_url)")
	}
	if !validSubdomain(subdomain) {
		return "", fmt.Errorf("zendesk-support config subdomain %q is invalid", subdomain)
	}
	return "https://" + subdomain + ".zendesk.com/api/v2", nil
}

func validSubdomain(s string) bool {
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-':
		default:
			return false
		}
	}
	return s != ""
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("zendesk-support config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("zendesk-support config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("zendesk-support config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("zendesk-support config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// secret looks up a secret under any of the provided keys (the catalog uses
// dotted keys like "credentials.api_token", but a bare key may also be supplied).
func secret(cfg connectors.RuntimeConfig, keys ...string) string {
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

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
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
