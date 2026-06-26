// Package microsoftlists implements the native pm Microsoft Lists connector. It
// is a declarative-HTTP per-system connector built on the connsdk toolkit
// (Requester + OAuth2 client-credentials auth + RecordsAt extraction) reading
// SharePoint/Microsoft Lists resources through the Microsoft Graph API.
//
// The connector self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the
// production binary to run that side effect.
//
// The package directory is "microsoft-lists" (the bare system name, hyphen
// included) and the registry key is likewise "microsoft-lists"; only the Go
// package identifier is sanitized to "microsoftlists".
package microsoftlists

import (
	"context"
	"encoding/json"
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
	// graphDefaultBaseURL is the Microsoft Graph v1.0 root.
	graphDefaultBaseURL = "https://graph.microsoft.com/v1.0"
	// graphDefaultScope is the client-credentials scope for app-only Graph access.
	graphDefaultScope = "https://graph.microsoft.com/.default"
	// loginBaseURL is the Microsoft identity platform token host.
	loginBaseURL = "https://login.microsoftonline.com"

	graphDefaultPageSize = 100
	graphMaxPageSize     = 200
	graphUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("microsoft-lists", New)
}

// New returns the Microsoft Lists connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Microsoft Lists connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// (and the OAuth2 token fetch). Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "microsoft-lists" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "microsoft-lists",
		DisplayName:     "Microsoft Lists",
		IntegrationType: "api",
		Description:     "Reads SharePoint/Microsoft Lists, list items, columns, and content types from a site through the Microsoft Graph API using app-only OAuth2 client credentials.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Graph. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := graphBaseURL(cfg); err != nil {
		return err
	}
	siteID, err := siteID(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the site's lists confirms token acquisition, auth, and
	// connectivity without mutating anything.
	path := "sites/" + url.PathEscape(siteID) + "/lists"
	if err := r.DoJSON(ctx, http.MethodGet, path, url.Values{"$top": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check microsoft-lists: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

// InitialState satisfies connectors.StatefulReader. Graph full_refresh streams
// start with an empty cursor.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

// Read streams records for the requested stream. The lists stream is
// site-scoped; list_items/columns/content_types require a configured list_id.
func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "lists"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("microsoft-lists stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	site, err := siteID(req.Config)
	if err != nil {
		return err
	}
	listID := ""
	if endpoint.needsListID {
		listID = strings.TrimSpace(req.Config.Config["list_id"])
		if listID == "" {
			return fmt.Errorf("microsoft-lists stream %q requires config list_id", stream)
		}
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := graphPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := graphMaxPages(req.Config)
	if err != nil {
		return err
	}
	path := "sites/" + url.PathEscape(site) + "/" + endpoint.resourceFor(url.PathEscape(listID))
	return c.harvest(ctx, r, path, endpoint, pageSize, maxPages, emit)
}

// harvest drives Microsoft Graph's @odata.nextLink pagination. Graph list
// responses return {"value":[...], "@odata.nextLink":"<absolute url>"}; the
// next page is fetched by following that absolute URL verbatim. connsdk has no
// paginator for the body-embedded absolute-next-link shape, so the loop lives
// here, built on connsdk.Requester + connsdk.RecordsAt plus the local nextLink
// helper (the @odata.nextLink key name contains a dot, so the dotted-path
// helpers cannot address it). The Requester treats an http(s) path as absolute,
// so nextLink is passed straight through.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	query := url.Values{}
	query.Set("$top", strconv.Itoa(pageSize))
	for k, v := range endpoint.query {
		query.Set(k, v)
	}

	next := path
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		// First request carries the query params; subsequent requests follow
		// the absolute nextLink which already encodes its own params.
		var reqQuery url.Values
		if page == 0 {
			reqQuery = query
		}
		resp, err := r.Do(ctx, http.MethodGet, next, reqQuery, nil)
		if err != nil {
			return fmt.Errorf("read microsoft-lists %s: %w", endpoint.resourceFor(""), err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "value")
		if err != nil {
			return fmt.Errorf("decode microsoft-lists page: %w", err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		link, err := nextLink(resp.Body)
		if err != nil {
			return fmt.Errorf("decode microsoft-lists nextLink: %w", err)
		}
		if link == "" {
			return nil
		}
		next = link
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise microsoft-lists credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		ts := fmt.Sprintf("2026-01-0%dT00:00:00Z", i)
		item := map[string]any{
			"id":                   fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":                 fmt.Sprintf("%s-%d", stream, i),
			"displayName":          fmt.Sprintf("Fixture %s %d", stream, i),
			"description":          "polymetrics fixture record",
			"webUrl":               fmt.Sprintf("https://contoso.sharepoint.com/%s/%d", stream, i),
			"eTag":                 fmt.Sprintf("etag-%d", i),
			"createdDateTime":      ts,
			"lastModifiedDateTime": ts,
			"list":                 map[string]any{"template": "genericList"},
			"contentType":          map[string]any{"id": "0x0100FIXTURE"},
			"fields":               map[string]any{"Title": fmt.Sprintf("Item %d", i)},
			"columnGroup":          "Custom Columns",
			"required":             false,
			"readOnly":             false,
			"hidden":               false,
			"indexed":              false,
			"group":                "Custom Content Types",
			"sealed":               false,
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

// requester builds a connsdk.Requester wired with OAuth2 client-credentials auth
// and the resolved Graph base URL. The client_secret only ever flows into the
// authenticator; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := graphBaseURL(cfg)
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
		UserAgent: graphUserAgent,
	}, nil
}

// authenticator builds the OAuth2 client-credentials authenticator that mints
// app-only Graph tokens. tenant_id/client_id/client_secret come from secrets;
// the token URL and scope can be overridden via config for test servers and
// sovereign clouds.
func (c Connector) authenticator(cfg connectors.RuntimeConfig) (*connsdk.OAuth2ClientCredentials, error) {
	clientID := secret(cfg, "client_id")
	clientSecret := secret(cfg, "client_secret")
	tenantID := secret(cfg, "tenant_id")
	if clientID == "" || clientSecret == "" || tenantID == "" {
		return nil, errors.New("microsoft-lists connector requires secrets client_id, client_secret, and tenant_id")
	}
	tokenURL, err := tokenURL(cfg, tenantID)
	if err != nil {
		return nil, err
	}
	scope := strings.TrimSpace(cfg.Config["scope"])
	if scope == "" {
		scope = graphDefaultScope
	}
	return &connsdk.OAuth2ClientCredentials{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{scope},
		Client:       c.Client,
	}, nil
}

// tokenURL resolves the OAuth2 token endpoint. The default is the public
// Microsoft identity platform v2.0 token endpoint for the tenant; an override
// (token_url config) must be an absolute http(s) URL with a host to bound SSRF
// risk (also used to point at a local test server).
func tokenURL(cfg connectors.RuntimeConfig, tenantID string) (string, error) {
	override := strings.TrimSpace(cfg.Config["token_url"])
	if override == "" {
		return loginBaseURL + "/" + url.PathEscape(tenantID) + "/oauth2/v2.0/token", nil
	}
	if err := validateAbsoluteURL("token_url", override); err != nil {
		return "", err
	}
	return override, nil
}

// graphBaseURL resolves and validates the Graph base URL. The default is
// graph.microsoft.com/v1.0; any override must be an absolute http(s) URL with a
// host to bound SSRF risk.
func graphBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return graphDefaultBaseURL, nil
	}
	if err := validateAbsoluteURL("base_url", base); err != nil {
		return "", err
	}
	return strings.TrimRight(base, "/"), nil
}

func validateAbsoluteURL(field, raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("microsoft-lists config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("microsoft-lists config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("microsoft-lists config %s must include a host", field)
	}
	return nil
}

func siteID(cfg connectors.RuntimeConfig) (string, error) {
	id := strings.TrimSpace(cfg.Config["site_id"])
	if id == "" {
		return "", errors.New("microsoft-lists connector requires config site_id")
	}
	return id, nil
}

func graphPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return graphDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("microsoft-lists config page_size must be an integer: %w", err)
	}
	if value < 1 || value > graphMaxPageSize {
		return 0, fmt.Errorf("microsoft-lists config page_size must be between 1 and %d", graphMaxPageSize)
	}
	return value, nil
}

func graphMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("microsoft-lists config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("microsoft-lists config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// nextLink extracts the OData @odata.nextLink absolute URL from a Graph
// collection response. The key name itself contains a dot, so the dotted-path
// helpers in connsdk cannot address it; a direct decode is used instead.
func nextLink(body []byte) (string, error) {
	var envelope struct {
		Next string `json:"@odata.nextLink"`
	}
	if len(body) == 0 {
		return "", nil
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return "", err
	}
	return strings.TrimSpace(envelope.Next), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// secret resolves a secret value by key, returning "" when absent.
func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Secrets[key])
}

// Write is unsupported: microsoft-lists is a read-only connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsFailed: len(records)}, connectors.ErrUnsupportedOperation
}
