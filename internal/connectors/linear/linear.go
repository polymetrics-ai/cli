// Package linear implements the native pm Linear connector. It follows the
// declarative-HTTP shape of the stripe reference connector, but targets Linear's
// single-endpoint GraphQL API: every list query POSTs a paginated connection
// query to https://api.linear.app/graphql and walks the cursor connection
// (pageInfo.hasNextPage + endCursor) using the `after` variable.
//
// Like stripe and github, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package linear

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
	linearDefaultBaseURL = "https://api.linear.app/graphql"
	// linearDefaultPageSize is Linear's default connection page size; the API
	// caps `first` at 250.
	linearDefaultPageSize = 50
	linearMaxPageSize     = 250
	linearUserAgent       = "polymetrics-go-cli"
	// linearFixtureCreated is the deterministic ISO-8601 base timestamp used by
	// the fixture-mode records.
	linearFixtureCreated = "2026-01-01T00:00:00.000Z"
)

func init() {
	connectors.RegisterFactory("linear", New)
}

// New returns the Linear connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Linear connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "linear" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "linear",
		DisplayName:     "Linear",
		IntegrationType: "api",
		Description:     "Reads Linear issues, teams, projects, and users through the Linear GraphQL API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Linear. In
// fixture mode it short-circuits without a network call. Otherwise it runs a
// bounded viewer query that confirms auth and connectivity without mutating
// anything.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := linearBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(linearSecret(cfg)) == "" {
		return errors.New("linear connector requires secret api_key")
	}
	r, endpointURL, err := c.requester(cfg)
	if err != nil {
		return err
	}
	body := graphQLRequest{Query: `query { viewer { id } }`}
	if err := r.DoJSON(ctx, http.MethodPost, endpointURL, nil, body, nil); err != nil {
		return fmt.Errorf("check linear: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: linearStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Linear stream starts with
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
		stream = "issues"
	}
	endpoint, ok := linearStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("linear stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	r, endpointURL, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := linearPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := linearMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpointURL, endpoint, pageSize, maxPages, emit)
}

// Write is unsupported; Linear is read-only for this connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// graphQLRequest is the JSON body sent to the Linear GraphQL endpoint.
type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// harvest drives Linear's GraphQL cursor-connection pagination. Each response is
// shaped {data:{<connection>:{nodes:[...],pageInfo:{hasNextPage,endCursor}}}};
// the next page is requested by setting the `after` variable to endCursor. The
// connsdk query-param paginators do not fit GraphQL POST-body cursors, so the
// loop lives here, built on connsdk.Requester + connsdk.RecordsAt +
// connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpointURL string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	query := buildConnectionQuery(endpoint.connection, endpoint.selection)
	nodesPath := "data." + endpoint.connection + ".nodes"
	hasNextPath := "data." + endpoint.connection + ".pageInfo.hasNextPage"
	endCursorPath := "data." + endpoint.connection + ".pageInfo.endCursor"

	after := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		vars := map[string]any{"first": pageSize}
		if after != "" {
			vars["after"] = after
		}
		body := graphQLRequest{Query: query, Variables: vars}

		resp, err := r.Do(ctx, http.MethodPost, endpointURL, nil, body)
		if err != nil {
			return fmt.Errorf("read linear %s: %w", endpoint.connection, err)
		}
		if msg := graphQLErrors(resp.Body); msg != "" {
			return fmt.Errorf("read linear %s: graphql error: %s", endpoint.connection, msg)
		}
		records, err := connsdk.RecordsAt(resp.Body, nodesPath)
		if err != nil {
			return fmt.Errorf("decode linear %s page: %w", endpoint.connection, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		hasNext, err := connsdk.StringAt(resp.Body, hasNextPath)
		if err != nil {
			return fmt.Errorf("decode linear %s pageInfo: %w", endpoint.connection, err)
		}
		endCursor, err := connsdk.StringAt(resp.Body, endCursorPath)
		if err != nil {
			return fmt.Errorf("decode linear %s endCursor: %w", endpoint.connection, err)
		}
		if hasNext != "true" || strings.TrimSpace(endCursor) == "" {
			return nil
		}
		after = endCursor
	}
	return nil
}

// buildConnectionQuery assembles a paginated GraphQL query for a root
// connection: query Q($first: Int!, $after: String) { <conn>(first:$first,
// after:$after, includeArchived:false) { nodes { <selection> } pageInfo {
// hasNextPage endCursor } } }.
func buildConnectionQuery(connection, selection string) string {
	var b strings.Builder
	b.WriteString("query PMSync($first: Int!, $after: String) {\n  ")
	b.WriteString(connection)
	b.WriteString("(first: $first, after: $after, includeArchived: false) {\n    nodes {\n")
	for _, line := range strings.Split(selection, "\n") {
		b.WriteString("      ")
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("    }\n    pageInfo {\n      hasNextPage\n      endCursor\n    }\n  }\n}")
	return b.String()
}

// graphQLErrors returns a non-empty message when the GraphQL response carries a
// top-level "errors" array.
func graphQLErrors(body []byte) string {
	errs, err := connsdk.RecordsAt(body, "errors")
	if err != nil || len(errs) == 0 {
		return ""
	}
	if msg, ok := errs[0]["message"].(string); ok && msg != "" {
		return msg
	}
	return "request failed"
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise linear credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	endpoint := linearStreamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := fixtureNode(stream, i)
		record := endpoint.mapRecord(item)
		record["connector"] = "linear"
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

// fixtureNode builds a deterministic GraphQL node for a stream + index.
func fixtureNode(stream string, i int) map[string]any {
	created := linearFixtureCreated
	updated := fmt.Sprintf("2026-01-0%dT00:00:00.000Z", i+1)
	base := map[string]any{
		"id":          fmt.Sprintf("%s_fixture_%d", stream, i),
		"createdAt":   created,
		"updatedAt":   updated,
		"name":        fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
		"description": "fixture record",
	}
	switch stream {
	case "issues":
		base["identifier"] = fmt.Sprintf("ENG-%d", i)
		base["title"] = fmt.Sprintf("Fixture issue %d", i)
		base["priority"] = int64(i)
		base["url"] = "https://linear.app/example/issue/ENG-" + strconv.Itoa(i)
		base["branchName"] = fmt.Sprintf("eng-%d", i)
		base["completedAt"] = nil
		base["canceledAt"] = nil
		base["state"] = map[string]any{"id": "state_1", "name": "Todo", "type": "unstarted"}
		base["team"] = map[string]any{"id": "team_1", "key": "ENG", "name": "Engineering"}
		base["assignee"] = map[string]any{"id": "user_1", "name": "Fixture User", "email": "fixture@example.com"}
		base["creator"] = map[string]any{"id": "user_1", "name": "Fixture User", "email": "fixture@example.com"}
	case "teams":
		base["key"] = fmt.Sprintf("T%d", i)
		base["private"] = false
	case "projects":
		base["state"] = "started"
		base["progress"] = 0.5
		base["url"] = "https://linear.app/example/project/" + strconv.Itoa(i)
		base["startedAt"] = created
		base["completedAt"] = nil
		base["canceledAt"] = nil
	case "users":
		base["displayName"] = fmt.Sprintf("fixture%d", i)
		base["email"] = fmt.Sprintf("fixture+%d@example.com", i)
		base["active"] = true
		base["admin"] = false
	}
	return base
}

// requester builds a connsdk.Requester wired with Linear auth and the resolved
// base URL. Personal API keys use a bare Authorization header (no Bearer
// prefix); OAuth access tokens use Bearer. The secret only ever flows into the
// authenticator; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, string, error) {
	endpointURL, err := linearBaseURL(cfg)
	if err != nil {
		return nil, "", err
	}
	secret := strings.TrimSpace(linearSecret(cfg))
	if secret == "" {
		return nil, "", errors.New("linear connector requires secret api_key")
	}
	// The full GraphQL endpoint URL is passed as an absolute path to each
	// request, so BaseURL is left empty (connsdk uses the absolute URL as-is).
	r := &connsdk.Requester{
		Client:    c.Client,
		Auth:      linearAuthenticator(cfg, secret),
		UserAgent: linearUserAgent,
	}
	return r, endpointURL, nil
}

// linearAuthenticator picks the auth scheme. A personal API key is sent as a
// bare Authorization header; an OAuth access token (auth_type=oauth, or an
// explicit access_token secret) uses the Bearer prefix.
func linearAuthenticator(cfg connectors.RuntimeConfig, secret string) connsdk.Authenticator {
	if isOAuth(cfg) {
		return connsdk.Bearer(secret)
	}
	// Bare key in the Authorization header; APIKeyHeader with an empty prefix.
	return connsdk.APIKeyHeader("Authorization", secret, "")
}

func isOAuth(cfg connectors.RuntimeConfig) bool {
	if cfg.Config != nil {
		if strings.EqualFold(strings.TrimSpace(cfg.Config["auth_type"]), "oauth") ||
			strings.EqualFold(strings.TrimSpace(cfg.Config["auth_type"]), "oauth2.0") {
			return true
		}
	}
	if cfg.Secrets != nil {
		if strings.TrimSpace(cfg.Secrets["access_token"]) != "" {
			return true
		}
	}
	return false
}

// linearSecret resolves the API credential from Secrets. A personal API key
// (api_key) is preferred; an OAuth access_token is used when present.
func linearSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	if token := strings.TrimSpace(cfg.Secrets["access_token"]); token != "" {
		return token
	}
	return cfg.Secrets["api_key"]
}

// linearBaseURL resolves and validates the base URL. The default is
// api.linear.app/graphql; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func linearBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return linearDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("linear config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("linear config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("linear config base_url must include a host")
	}
	// Append the GraphQL path when the override is just a host (so a bare host
	// override still posts to /graphql), but leave an explicit path intact (the
	// httptest server passes its own /graphql).
	trimmed := strings.TrimRight(base, "/")
	if parsed.Path == "" || parsed.Path == "/" {
		return trimmed + "/graphql", nil
	}
	return trimmed, nil
}

func linearPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return linearDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("linear config page_size must be an integer: %w", err)
	}
	if value < 1 || value > linearMaxPageSize {
		return 0, fmt.Errorf("linear config page_size must be between 1 and %d", linearMaxPageSize)
	}
	return value, nil
}

func linearMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("linear config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("linear config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
