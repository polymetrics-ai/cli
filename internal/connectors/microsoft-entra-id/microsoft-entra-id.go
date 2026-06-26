// Package microsoftentraid implements the native pm Microsoft Entra ID
// connector. It is a declarative-HTTP per-system connector built on the connsdk
// toolkit: an OAuth2 client-credentials authenticator against Microsoft's
// identity platform, Microsoft Graph directory stream definitions, and the
// Graph @odata.nextLink pagination shape.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// The directory name is "microsoft-entra-id" (hyphenated) but the Go package
// identifier is sanitized to "microsoftentraid"; the registry key and Name()
// remain the bare hyphenated "microsoft-entra-id".
package microsoftentraid

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
	connectorName       = "microsoft-entra-id"
	defaultBaseURL      = "https://graph.microsoft.com/v1.0"
	defaultLoginBaseURL = "https://login.microsoftonline.com"
	graphScope          = "https://graph.microsoft.com/.default"
	userAgent           = "polymetrics-go-cli"
	defaultPageSize     = 100
	maxPageSize         = 999
)

func init() {
	connectors.RegisterFactory(connectorName, New)
}

// New returns the Microsoft Entra ID connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Microsoft Entra ID connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// (and the OAuth2 token exchange). Left nil in production; injectable for
	// tests.
	Client *http.Client
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Microsoft Entra ID",
		IntegrationType: "api",
		Description:     "Reads Microsoft Entra ID (Azure AD) directory objects — users, groups, applications, service principals, and directory roles — from the Microsoft Graph API using OAuth2 client credentials.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Microsoft
// Graph. In fixture mode it short-circuits without a network call.
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
	if err := requireSecrets(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the users collection confirms the token exchange, auth,
	// and connectivity without mutating anything.
	q := url.Values{"$top": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "users", q, nil, nil); err != nil {
		return fmt.Errorf("check microsoft-entra-id: %w", err)
	}
	return nil
}

// Write is unsupported: the Microsoft Entra ID connector is read-only. Mutating
// directory objects (users, groups, applications) is out of scope for reverse
// ETL and intentionally not exposed.
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
		stream = "users"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("microsoft-entra-id stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
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

// harvest drives Microsoft Graph's @odata.nextLink pagination. Graph collection
// responses are {value:[...], "@odata.nextLink": "<absolute url>"}; the next
// page is requested by GETting that absolute URL (which already carries the
// $skiptoken cursor). connsdk.Requester treats an http(s) path as absolute, so
// the loop simply re-feeds the nextLink. There is no body-token paginator in
// connsdk for this exact (absolute-URL) shape, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := endpoint.resource
	query := url.Values{"$top": []string{strconv.Itoa(pageSize)}}

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read microsoft-entra-id %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "value")
		if err != nil {
			return fmt.Errorf("decode microsoft-entra-id %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := nextLink(resp.Body)
		if err != nil {
			return fmt.Errorf("decode microsoft-entra-id %s nextLink: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		// The nextLink is an absolute URL that already carries the cursor and any
		// page-size hints; subsequent pages must not re-merge $top.
		path = next
		query = nil
	}
	return nil
}

// nextLink extracts the Microsoft Graph "@odata.nextLink" absolute URL from a
// collection response body. The key contains a literal dot, so the connsdk
// dotted-path helpers cannot select it; we decode the top-level object directly.
func nextLink(body []byte) (string, error) {
	var envelope struct {
		NextLink string `json:"@odata.nextLink"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return "", fmt.Errorf("decode graph envelope: %w", err)
	}
	return envelope.NextLink, nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free (mirrors the
// stripe fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	endpoint := streamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                     fmt.Sprintf("%s-fixture-%d", stream, i),
			"displayName":            fmt.Sprintf("Fixture %s %d", stream, i),
			"userPrincipalName":      fmt.Sprintf("fixture+%d@example.com", i),
			"givenName":              fmt.Sprintf("Fixture%d", i),
			"surname":                "Example",
			"mail":                   fmt.Sprintf("fixture+%d@example.com", i),
			"jobTitle":               "Engineer",
			"department":             "Platform",
			"officeLocation":         "Remote",
			"mobilePhone":            "+15555550100",
			"accountEnabled":         true,
			"description":            fmt.Sprintf("Fixture %s record %d", stream, i),
			"mailNickname":           fmt.Sprintf("fixture%d", i),
			"mailEnabled":            false,
			"securityEnabled":        true,
			"visibility":             "Private",
			"createdDateTime":        "2026-01-01T00:00:00Z",
			"appId":                  fmt.Sprintf("app-%d", i),
			"signInAudience":         "AzureADMyOrg",
			"publisherDomain":        "example.com",
			"servicePrincipalType":   "Application",
			"appOwnerOrganizationId": "org-fixture",
			"roleTemplateId":         fmt.Sprintf("role-template-%d", i),
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with OAuth2 client-credentials auth
// against the tenant token endpoint and the resolved Graph base URL. Secrets
// only ever flow into connsdk.OAuth2ClientCredentials; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireSecrets(cfg); err != nil {
		return nil, err
	}
	token, err := tokenURL(cfg)
	if err != nil {
		return nil, err
	}
	auth := &connsdk.OAuth2ClientCredentials{
		TokenURL:     token,
		ClientID:     secret(cfg, "client_id"),
		ClientSecret: secret(cfg, "client_secret"),
		Scopes:       []string{graphScope},
		Client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: userAgent,
	}, nil
}

func requireSecrets(cfg connectors.RuntimeConfig) error {
	for _, key := range []string{"client_id", "client_secret", "tenant_id"} {
		if strings.TrimSpace(secret(cfg, key)) == "" {
			return fmt.Errorf("microsoft-entra-id connector requires secret %s", key)
		}
	}
	return nil
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

// baseURL resolves and validates the Graph base URL. The default is
// graph.microsoft.com/v1.0; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	return resolveURL(cfg.Config["base_url"], defaultBaseURL)
}

// tokenURL resolves the OAuth2 token endpoint. By default it is derived from the
// tenant_id secret against login.microsoftonline.com; an explicit token_url
// config override (validated for scheme+host) takes precedence for local test
// servers.
func tokenURL(cfg connectors.RuntimeConfig) (string, error) {
	if override := strings.TrimSpace(cfg.Config["token_url"]); override != "" {
		return resolveURL(override, "")
	}
	tenant := strings.TrimSpace(secret(cfg, "tenant_id"))
	if tenant == "" {
		return "", errors.New("microsoft-entra-id connector requires secret tenant_id")
	}
	loginBase, err := resolveURL(cfg.Config["login_base_url"], defaultLoginBaseURL)
	if err != nil {
		return "", err
	}
	return loginBase + "/" + url.PathEscape(tenant) + "/oauth2/v2.0/token", nil
}

// resolveURL validates an absolute http/https URL, falling back to def when raw
// is empty. An empty def with an empty raw is an error.
func resolveURL(raw, def string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if def == "" {
			return "", errors.New("microsoft-entra-id connector requires a URL")
		}
		return def, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("microsoft-entra-id config url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("microsoft-entra-id config url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("microsoft-entra-id config url must include a host")
	}
	return strings.TrimRight(raw, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("microsoft-entra-id config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("microsoft-entra-id config page_size must be between 1 and %d", maxPageSize)
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
		return 0, fmt.Errorf("microsoft-entra-id config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("microsoft-entra-id config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
