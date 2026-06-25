// Package bingads implements the native pm Bing Ads (Microsoft Advertising)
// connector. It follows the declarative-HTTP template established by the stripe
// connector: a thin package that composes the connsdk toolkit (Requester + a
// Microsoft OAuth refresh-token authenticator + RecordsAt extraction) with
// Bing-Ads-specific stream definitions and endpoints.
//
// The Bing Ads API exposes a REST/JSON surface over the v13 Customer Management
// and Campaign Management services. Every operation is an HTTP POST to a
// "/Query"-style endpoint with a small JSON body; the response wraps the records
// in a named array (e.g. {"AccountsInfo":[...]}, {"Campaigns":[...]}). Auth is
// OAuth2: the configured refresh_token is exchanged at the Microsoft identity
// platform token endpoint for a short-lived access token, which is sent as
// Authorization: Bearer <token> alongside the DeveloperToken header.
//
// Like stripe/github, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect. The Go package is named
// bingads (a valid identifier) while the connector's registry key is the bare
// system name "bing-ads".
package bingads

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	// connectorName is the registry key / catalog alias target. It is the bare
	// system name and intentionally differs from the Go package identifier.
	connectorName = "bing-ads"

	// defaultCustomerBaseURL is the production Customer Management REST root.
	defaultCustomerBaseURL = "https://clientcenter.api.bingads.microsoft.com/CustomerManagement/v13"
	// defaultCampaignBaseURL is the production Campaign Management REST root.
	defaultCampaignBaseURL = "https://campaign.api.bingads.microsoft.com/CampaignManagement/v13"
	// defaultTokenURLTemplate is the Microsoft identity platform token endpoint;
	// %s is the tenant id (default "common").
	defaultTokenURLTemplate = "https://login.microsoftonline.com/%s/oauth2/v2.0/token"

	userAgent       = "polymetrics-go-cli"
	defaultTenantID = "common"
)

func init() {
	connectors.RegisterFactory(connectorName, New)
}

// New returns the Bing Ads connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Bing Ads connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Bing Ads",
		IntegrationType: "api",
		Description:     "Reads Microsoft Advertising (Bing Ads) accounts, users, campaigns, ad groups, and ads through the v13 Customer Management and Campaign Management REST APIs.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Bing Ads. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := customerBaseURL(cfg); err != nil {
		return err
	}
	if err := validateSecrets(cfg); err != nil {
		return err
	}
	r, _, err := c.requester(cfg, serviceCustomer)
	if err != nil {
		return err
	}
	// A bounded AccountsInfo query confirms the OAuth exchange, the developer
	// token, and connectivity without mutating anything.
	ep := streamEndpoints["accounts"]
	if _, err := r.Do(ctx, http.MethodPost, ep.path, nil, ep.body(cfg)); err != nil {
		return fmt.Errorf("check bing-ads: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: bingStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "accounts"
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("bing-ads stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, ep, req, emit)
	}

	r, _, err := c.requester(req.Config, ep.kind)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodPost, ep.path, nil, ep.body(req.Config))
	if err != nil {
		return fmt.Errorf("read bing-ads %s: %w", stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, ep.recordsPath)
	if err != nil {
		return fmt.Errorf("decode bing-ads %s: %w", stream, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(ep.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// Write is not supported: the Bing Ads connector is read-only (Capabilities.Write
// is false). It satisfies the connectors.Connector interface by returning the
// shared unsupported-operation error.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise bing-ads credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, ep streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for _, item := range ep.fixture() {
		if err := ctx.Err(); err != nil {
			return err
		}
		record := ep.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the Microsoft OAuth
// refresh-token authenticator, the resolved base URL for the stream's service,
// the DeveloperToken header, and (for campaign-scoped streams) the CustomerId and
// CustomerAccountId headers. Secrets only ever flow into the authenticator and
// headers; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig, kind serviceKind) (*connsdk.Requester, *oauthRefreshAuth, error) {
	if err := validateSecrets(cfg); err != nil {
		return nil, nil, err
	}

	var base string
	var err error
	switch kind {
	case serviceCampaign:
		base, err = campaignBaseURL(cfg)
	default:
		base, err = customerBaseURL(cfg)
	}
	if err != nil {
		return nil, nil, err
	}

	tokenURL, err := tokenURL(cfg)
	if err != nil {
		return nil, nil, err
	}

	auth := &oauthRefreshAuth{
		tokenURL:     tokenURL,
		clientID:     secret(cfg, "client_id"),
		clientSecret: secret(cfg, "client_secret"),
		refreshToken: secret(cfg, "refresh_token"),
		client:       c.Client,
	}

	headers := map[string]string{
		"DeveloperToken": secret(cfg, "developer_token"),
	}
	if kind == serviceCampaign {
		if v := strings.TrimSpace(cfg.Config["customer_id"]); v != "" {
			headers["CustomerId"] = v
		}
		if v := customerAccountID(cfg); v != "" {
			headers["CustomerAccountId"] = v
		}
	}

	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           auth,
		UserAgent:      userAgent,
		DefaultHeaders: headers,
	}, auth, nil
}

// validateSecrets enforces the required credential set. Values are never logged.
func validateSecrets(cfg connectors.RuntimeConfig) error {
	for _, name := range []string{"developer_token", "client_id", "refresh_token"} {
		if strings.TrimSpace(secret(cfg, name)) == "" {
			return fmt.Errorf("bing-ads connector requires secret %s", name)
		}
	}
	return nil
}

func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}

// customerAccountID resolves the CustomerAccountId, falling back to account_id.
func customerAccountID(cfg connectors.RuntimeConfig) string {
	if v := strings.TrimSpace(cfg.Config["customer_account_id"]); v != "" {
		return v
	}
	return strings.TrimSpace(cfg.Config["account_id"])
}

// customerBaseURL resolves and validates the Customer Management base URL.
func customerBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	return resolveBaseURL(cfg, "base_url", defaultCustomerBaseURL)
}

// campaignBaseURL resolves and validates the Campaign Management base URL.
func campaignBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	return resolveBaseURL(cfg, "campaign_base_url", defaultCampaignBaseURL)
}

// resolveBaseURL reads an override from cfg.Config[key] (or returns def) and
// validates scheme+host to bound SSRF risk.
func resolveBaseURL(cfg connectors.RuntimeConfig, key, def string) (string, error) {
	base := strings.TrimSpace(cfg.Config[key])
	if base == "" {
		return def, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("bing-ads config %s is invalid: %w", key, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("bing-ads config %s must use http or https, got %q", key, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("bing-ads config %s must include a host", key)
	}
	return strings.TrimRight(base, "/"), nil
}

// tokenURL resolves the OAuth token endpoint. A token_url override wins;
// otherwise the Microsoft identity platform endpoint for the configured tenant
// is used. Overrides are scheme/host validated.
func tokenURL(cfg connectors.RuntimeConfig) (string, error) {
	if override := strings.TrimSpace(cfg.Config["token_url"]); override != "" {
		parsed, err := url.Parse(override)
		if err != nil {
			return "", fmt.Errorf("bing-ads config token_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("bing-ads config token_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("bing-ads config token_url must include a host")
		}
		return override, nil
	}
	tenant := strings.TrimSpace(secret(cfg, "tenant_id"))
	if tenant == "" {
		tenant = strings.TrimSpace(cfg.Config["tenant_id"])
	}
	if tenant == "" {
		tenant = defaultTenantID
	}
	return fmt.Sprintf(defaultTokenURLTemplate, url.PathEscape(tenant)), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
