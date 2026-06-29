// Package microsoftteams implements the native pm Microsoft Teams connector. It
// is a declarative-HTTP per-system connector built on the connsdk toolkit
// (OAuth2 client-credentials auth against Microsoft Entra ID + Microsoft Graph
// value[]/@odata.nextLink pagination) plus Teams-specific stream definitions.
//
// The directory is internal/connectors/microsoft-teams (hyphenated to match the
// bare system name); the Go package identifier is sanitized to microsoftteams.
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package microsoftteams

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	// graphDefaultBaseURL is the production-stable Microsoft Graph REST API.
	graphDefaultBaseURL = "https://graph.microsoft.com/v1.0"
	// graphDefaultScope requests the static application-permission scope for the
	// client-credentials flow.
	graphDefaultScope = "https://graph.microsoft.com/.default"
	graphUserAgent    = "polymetrics-go-cli"
	// graphFixtureDate is the deterministic date used by fixture-mode records.
	graphFixtureDate = "2026-01-01"
)

func init() {
	connectors.RegisterFactory("microsoft-teams", New)
}

// New returns the Microsoft Teams connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Microsoft Teams connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth2 token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "microsoft-teams" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "microsoft-teams",
		DisplayName:     "Microsoft Teams",
		IntegrationType: "api",
		Description:     "Reads Microsoft Teams users, groups, channels, and device-usage reports through the Microsoft Graph REST API using an OAuth2 client-credentials grant.",
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
	if _, err := graphBaseURL(cfg); err != nil {
		return err
	}
	if err := requireCredentials(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the organization collection confirms the token exchange
	// and connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "organization", url.Values{"$top": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check microsoft-teams: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: graphStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader. Microsoft Graph collections
// used here are full-refresh, so an empty cursor is the starting state.
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
	endpoint, ok := graphStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("microsoft-teams stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	// Validate config (including base_url scheme) before any network use.
	if _, err := graphBaseURL(req.Config); err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	query := url.Values{}
	if stream == "team_device_usage_report" {
		// The Teams device usage report requires the aggregation period.
		query.Set("period", devicePeriod(req.Config))
		query.Set("$format", "application/json")
	}
	maxPages := graphMaxPages(req.Config)
	return c.harvest(ctx, r, endpoint, query, maxPages, emit)
}

// harvest drives Microsoft Graph's @odata.nextLink pagination. Graph collections
// return {value:[...], "@odata.nextLink":"<absolute url>"}; the next page is the
// nextLink URL verbatim. connsdk.Requester treats an absolute http(s) path as-is,
// so the nextLink is passed straight back in. This shape is not covered by a
// connsdk paginator, so the loop lives here, built on Requester + RecordsAt +
// StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, firstQuery url.Values, maxPages int, emit func(connectors.Record) error) error {
	path := endpoint.resource
	query := firstQuery
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read microsoft-teams %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "value")
		if err != nil {
			return fmt.Errorf("decode microsoft-teams %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// The pagination key "@odata.nextLink" contains a literal dot, so it
		// cannot be read via connsdk.StringAt's dotted-path traversal; decode the
		// literal key directly.
		next, err := nextLink(resp.Body)
		if err != nil {
			return fmt.Errorf("decode microsoft-teams %s nextLink: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		// nextLink is an absolute URL carrying its own query (skiptoken etc.);
		// pass it as the path with no extra query so it is used verbatim.
		path = next
		query = nil
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise microsoft-teams credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                fmt.Sprintf("%s_fixture_%d", strings.ReplaceAll(stream, "/", "_"), i),
			"displayName":       fmt.Sprintf("Fixture %s %d", stream, i),
			"description":       "deterministic fixture record",
			"userPrincipalName": fmt.Sprintf("fixture+%d@example.com", i),
			"mail":              fmt.Sprintf("fixture+%d@example.com", i),
			"mailNickname":      fmt.Sprintf("fixture%d", i),
			"jobTitle":          "Engineer",
			"officeLocation":    "Remote",
			"mobilePhone":       "",
			"accountEnabled":    true,
			"visibility":        "Private",
			"createdDateTime":   graphFixtureDate + "T00:00:00Z",
			"securityEnabled":   true,
			"mailEnabled":       true,
			"email":             fmt.Sprintf("fixture+%d@example.com", i),
			"membershipType":    "standard",
			"webUrl":            "https://teams.microsoft.com/fixture",
			"lastActivityDate":  graphFixtureDate,
			"isDeleted":         false,
			"usedWeb":           true,
			"usedWindowsPhone":  false,
			"usedAndroidPhone":  false,
			"usedIOS":           true,
			"usedMac":           false,
			"reportPeriod":      devicePeriod(req.Config),
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

// Write is unsupported: microsoft-teams is a read-only source connector.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// requester builds a connsdk.Requester wired with OAuth2 client-credentials auth
// against Microsoft Entra ID and the resolved Graph base URL. The secret only
// ever flows into the authenticator; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := graphBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireCredentials(cfg); err != nil {
		return nil, err
	}
	auth := &connsdk.OAuth2ClientCredentials{
		TokenURL:     graphTokenURL(cfg),
		ClientID:     graphClientID(cfg),
		ClientSecret: graphClientSecret(cfg),
		Scopes:       []string{graphScope(cfg)},
		Client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: graphUserAgent,
	}, nil
}

// requireCredentials enforces that the secrets needed for the client-credentials
// grant are present. The refresh_token is part of the upstream OAuth config but
// the native port uses the application-permission client-credentials flow.
func requireCredentials(cfg connectors.RuntimeConfig) error {
	if strings.TrimSpace(graphClientSecret(cfg)) == "" {
		return errors.New("microsoft-teams connector requires secret client_secret")
	}
	if strings.TrimSpace(graphTenantID(cfg)) == "" {
		return errors.New("microsoft-teams connector requires secret tenant_id")
	}
	if strings.TrimSpace(graphClientID(cfg)) == "" {
		return errors.New("microsoft-teams connector requires config client_id")
	}
	return nil
}

func graphClientSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["client_secret"]
}

func graphTenantID(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets != nil {
		if v := strings.TrimSpace(cfg.Secrets["tenant_id"]); v != "" {
			return v
		}
	}
	return strings.TrimSpace(cfg.Config["tenant_id"])
}

func graphClientID(cfg connectors.RuntimeConfig) string {
	if cfg.Config != nil {
		if v := strings.TrimSpace(cfg.Config["client_id"]); v != "" {
			return v
		}
	}
	if cfg.Secrets != nil {
		return strings.TrimSpace(cfg.Secrets["client_id"])
	}
	return ""
}

func graphScope(cfg connectors.RuntimeConfig) string {
	if cfg.Config != nil {
		if v := strings.TrimSpace(cfg.Config["scope"]); v != "" {
			return v
		}
	}
	return graphDefaultScope
}

// graphTokenURL resolves the OAuth2 token endpoint. A token_url config override
// (used by tests) wins; otherwise the per-tenant Entra ID v2 token endpoint is
// derived from the tenant id.
func graphTokenURL(cfg connectors.RuntimeConfig) string {
	if cfg.Config != nil {
		if v := strings.TrimSpace(cfg.Config["token_url"]); v != "" {
			return v
		}
	}
	tenant := graphTenantID(cfg)
	if tenant == "" {
		tenant = "common"
	}
	return "https://login.microsoftonline.com/" + url.PathEscape(tenant) + "/oauth2/v2.0/token"
}

// graphBaseURL resolves and validates the base URL. The default is
// graph.microsoft.com/v1.0; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func graphBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return graphDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("microsoft-teams config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("microsoft-teams config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("microsoft-teams config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// devicePeriod resolves the Teams device usage report aggregation period. Graph
// accepts D7, D30, D90, D180; default D7.
func devicePeriod(cfg connectors.RuntimeConfig) string {
	if cfg.Config != nil {
		switch strings.ToUpper(strings.TrimSpace(cfg.Config["period"])) {
		case "D7", "D30", "D90", "D180":
			return strings.ToUpper(strings.TrimSpace(cfg.Config["period"]))
		}
	}
	return "D7"
}

func graphMaxPages(cfg connectors.RuntimeConfig) int {
	raw := strings.ToLower(strings.TrimSpace(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0
	}
	var n int
	if _, err := fmt.Sscanf(raw, "%d", &n); err != nil || n < 0 {
		return 0
	}
	return n
}

// nextLink reads the literal "@odata.nextLink" key from a Graph response body.
// It is an absolute URL (carrying its own $skiptoken) or empty on the last page.
func nextLink(body []byte) (string, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var envelope struct {
		NextLink string `json:"@odata.nextLink"`
	}
	if err := dec.Decode(&envelope); err != nil {
		return "", fmt.Errorf("decode graph envelope: %w", err)
	}
	return envelope.NextLink, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
