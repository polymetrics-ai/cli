// Package quickbooks implements the quickbooks bundle's two Tier-2 hooks
// (conventions.md §1's 2-interface cap):
//
//  1. AuthHook — an OAuth 2.0 refresh-token-grant connsdk.Authenticator,
//     porting internal/connectors/hooks/gmail/hooks.go's oauthRefreshAuth
//     almost verbatim. This is the REAL, catalog-documented QuickBooks auth
//     scheme (website/.enrich/enr/source-quickbooks.json:
//     client_id/client_secret/refresh_token, hourly access-token expiry) —
//     legacy internal/connectors/quickbooks/quickbooks.go's simplified
//     reference implementation uses a pre-issued static access_token secret
//     instead, which this bundle deliberately does not carry forward (see
//     defs/quickbooks/docs.md's Known limits).
//  2. StreamHook — ports legacy quickbooks.go's harvest loop verbatim: every
//     stream (customers/invoices/payments/accounts/vendors) shares the same
//     v3/company/{realmId}/query endpoint, and QuickBooks' STARTPOSITION/
//     MAXRESULTS pagination state is embedded inside the single "query"
//     query-string VALUE rather than sent as independent parameters — the
//     engine's 6 declarative pagination types cannot express this
//     (docs/migration/quarantine.json's original ENGINE_GAP blocker).
//
// Secret values (client_secret, the refresh token, cached access tokens)
// flow ONLY into the outgoing token-request form or the Authorization
// header; they are never logged and never appear in an error string
// (THREAT-MODEL.md Delta 2).
//
// Line-count self-report (conventions.md §1's ~300-line soft target): this
// file runs to ~380 lines, over the soft target but under the 400-line hard
// ceiling, with exactly 2 hook interfaces (AuthHook + StreamHook, at the
// cap). The size is accounted for by two genuinely independent, both-
// mandated shapes: (1) the full OAuth2 refresh-grant Authenticator
// (identical in shape/size to gmail's ~130-line oauthRefreshAuth), and (2)
// the 5-entity harvest/mapRecord table ported from legacy quickbooks.go
// (qbEndpoints + 5 small mapRecord functions + realm_id/page_size/max_pages
// validation helpers). Neither shape alone would need a hook; QuickBooks
// needs both simultaneously (real OAuth2 auth AND in-query-string
// pagination), which is why this bundle carries both interfaces in one
// package rather than splitting further (no 3rd interface is used, so no
// Tier-3 escalation is triggered).
package quickbooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	defaultPageSize = 1000
	safetyMaxPages  = 10000
)

func init() {
	engine.RegisterHooks("quickbooks", func() engine.Hooks { return New() })
}

// Hooks is the quickbooks hook set: AuthHook + StreamHook.
type Hooks struct {
	// Now is injectable for tests; nil uses time.Now (mirrors gmail's Hooks.Now).
	Now func() time.Time
	// Client overrides the HTTP client used for the token exchange; nil uses
	// a default client with a 30s timeout.
	Client *http.Client
}

// New returns a fresh quickbooks Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "quickbooks" }

// --- AuthHook -------------------------------------------------------------

// Authenticator resolves the OAuth2 refresh-token-grant connsdk.Authenticator
// for spec (mode "custom", hook "quickbooks"). See hooks/gmail/hooks.go for
// the identical shape this ports from.
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, spec engine.AuthSpec) (connsdk.Authenticator, error) {
	tokenURL, err := interpolateRequired(spec.TokenURL, "token_url", cfg)
	if err != nil {
		return nil, err
	}
	if err := validateHTTPSURL(tokenURL, "token_url"); err != nil {
		return nil, err
	}

	clientID, err := interpolateRequired(spec.ClientID, "client_id", cfg)
	if err != nil {
		return nil, err
	}
	clientSecret, err := interpolateRequired(spec.ClientSecret, "client_secret", cfg)
	if err != nil {
		return nil, err
	}
	refreshToken, err := interpolateRequired(spec.Token, "refresh_token", cfg)
	if err != nil {
		return nil, err
	}

	return &oauthRefreshAuth{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		client:       h.Client,
		now:          h.Now,
	}, nil
}

func interpolateRequired(tmpl, field string, cfg connectors.RuntimeConfig) (string, error) {
	val, err := engine.Interpolate(tmpl, engine.Vars{Config: cfg.Config, Secrets: cfg.Secrets})
	if err != nil {
		return "", fmt.Errorf("quickbooks oauth: resolve %s: %w", field, err)
	}
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("quickbooks oauth: %s is required", field)
	}
	return val, nil
}

// validateHTTPSURL fails closed on anything but a well-formed https:// URL
// with a host (THREAT-MODEL.md Delta 2, mirrors gmail's identical guard).
func validateHTTPSURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("quickbooks oauth: %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("quickbooks oauth: %s must use https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("quickbooks oauth: %s must include a host", field)
	}
	return nil
}

// oauthRefreshAuth implements connsdk.Authenticator for the QuickBooks OAuth
// 2.0 refresh-token grant: exchange the refresh token for a short-lived
// access token at tokenURL, cache it until 60s before its declared expiry,
// then set Authorization: Bearer <token> on each request. client_id and
// client_secret are both always sent (QuickBooks' token endpoint requires
// both, unlike gmail's optional client_secret).
type oauthRefreshAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	refreshToken string
	client       *http.Client

	now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *oauthRefreshAuth) timeNow() time.Time {
	if a.now != nil {
		return a.now()
	}
	return time.Now()
}

func (a *oauthRefreshAuth) httpClient() *http.Client {
	if a.client != nil {
		return a.client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// Apply ensures a fresh access token and sets the Authorization header.
func (a *oauthRefreshAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *oauthRefreshAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" && a.timeNow().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("quickbooks oauth: token URL is required")
	}
	if strings.TrimSpace(a.refreshToken) == "" {
		return "", errors.New("quickbooks oauth: refresh_token is required")
	}
	if strings.TrimSpace(a.clientID) == "" {
		return "", errors.New("quickbooks oauth: client_id is required")
	}
	if strings.TrimSpace(a.clientSecret) == "" {
		return "", errors.New("quickbooks oauth: client_secret is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", a.refreshToken)
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("quickbooks oauth: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.httpClient().Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("quickbooks oauth: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("quickbooks oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("quickbooks oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("quickbooks oauth: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.timeNow().Add(ttl)
	return a.token, nil
}

// --- StreamHook -------------------------------------------------------------

// qbEndpoint maps a stream name to its QuickBooks entity name and record
// mapper (ported verbatim from legacy quickbooks.go's qbEndpoints table).
type qbEndpoint struct {
	entity    string
	mapRecord func(map[string]any) connectors.Record
}

var qbEndpoints = map[string]qbEndpoint{
	"customers": {entity: "Customer", mapRecord: qbCustomer},
	"invoices":  {entity: "Invoice", mapRecord: qbInvoice},
	"payments":  {entity: "Payment", mapRecord: qbPayment},
	"accounts":  {entity: "Account", mapRecord: qbAccount},
	"vendors":   {entity: "Vendor", mapRecord: qbVendor},
}

// ReadStream implements engine.StreamHook, handling every declared stream
// with handled=true; an unknown stream name returns handled=false so the
// engine can surface its own "stream not found" error.
func (h *Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	name := stream.Name
	if name == "" {
		name = "customers"
	}
	endpoint, ok := qbEndpoints[name]
	if !ok {
		return false, nil
	}

	realmID, err := realmIDFrom(req.Config)
	if err != nil {
		return true, err
	}
	pageSize, err := pageSizeFrom(req.Config)
	if err != nil {
		return true, err
	}
	maxPages, err := maxPagesFrom(req.Config)
	if err != nil {
		return true, err
	}

	return true, h.harvest(ctx, rt.Requester, realmID, endpoint, pageSize, maxPages, emit)
}

// harvest ports legacy quickbooks.go's harvest loop verbatim: a page's
// STARTPOSITION/MAXRESULTS are baked into the single "query" query-string
// value; the loop stops when a page returns fewer than pageSize records, or
// maxPages (if set) is reached.
func (h *Hooks) harvest(ctx context.Context, r *connsdk.Requester, realmID string, endpoint qbEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	limit := maxPages
	if limit <= 0 {
		limit = safetyMaxPages
	}
	start := 1
	for page := 0; page < limit; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("query", fmt.Sprintf("SELECT * FROM %s STARTPOSITION %d MAXRESULTS %d", endpoint.entity, start, pageSize))
		resp, err := r.Do(ctx, http.MethodGet, "v3/company/"+realmID+"/query", query, nil)
		if err != nil {
			return fmt.Errorf("read quickbooks %s: %w", endpoint.entity, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "QueryResponse."+endpoint.entity)
		if err != nil {
			return fmt.Errorf("decode quickbooks %s: %w", endpoint.entity, err)
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
		start += pageSize
	}
	return nil
}

// realmIDFrom resolves and path-safety-validates config.realm_id (mirrors
// legacy's cleanSegment guard: rejects any value containing "/", "?", "#",
// or "..", since it is interpolated directly into the request path).
func realmIDFrom(cfg connectors.RuntimeConfig) (string, error) {
	realmID := strings.TrimSpace(cfg.Config["realm_id"])
	if realmID == "" || strings.ContainsAny(realmID, "/?#") || strings.Contains(realmID, "..") {
		return "", errors.New("quickbooks connector requires a path-safe config realm_id")
	}
	return realmID, nil
}

func pageSizeFrom(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > 1000 {
		return 0, fmt.Errorf("quickbooks config page_size must be between 1 and 1000")
	}
	return value, nil
}

func maxPagesFrom(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, errors.New("quickbooks config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}

// --- record mappers (ported verbatim from legacy quickbooks.go) -----------

func qbCustomer(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["Id"], "display_name": item["DisplayName"], "active": item["Active"], "balance": item["Balance"]}
}
func qbInvoice(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["Id"], "doc_number": item["DocNumber"], "customer_ref": refValue(item["CustomerRef"]), "total_amt": item["TotalAmt"], "balance": item["Balance"]}
}
func qbPayment(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["Id"], "customer_ref": refValue(item["CustomerRef"]), "total_amt": item["TotalAmt"], "txn_date": item["TxnDate"]}
}
func qbAccount(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["Id"], "name": item["Name"], "classification": item["Classification"], "account_type": item["AccountType"]}
}
func qbVendor(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["Id"], "display_name": item["DisplayName"], "active": item["Active"], "balance": item["Balance"]}
}

func refValue(v any) any {
	if m, ok := v.(map[string]any); ok {
		return m["value"]
	}
	return v
}
