// Package paypaltransaction implements the paypal-transaction bundle's two
// Tier-2 hooks (conventions.md §1's 2-interface cap):
//
//  1. AuthHook — PayPal's OAuth 2.0 client-credentials token exchange
//     authenticates the CLIENT with HTTP Basic (client_id:client_secret in
//     the Authorization header), not the engine's declarative
//     "oauth2_client_credentials" mode, which always POSTs a
//     grant_type/client_id/client_secret/scope FORM body
//     (engine/auth.go's buildOAuth2ClientCredentials /
//     connsdk.OAuth2ClientCredentials.accessToken) — PayPal's real wire
//     request sends client_id/client_secret via HTTP Basic and the form body
//     carries only grant_type=client_credentials. Neither built-in mode fits
//     (mirrors hooks/jamf-pro/hooks.go's identical reasoning for its own
//     Basic-credential token exchange), so a custom AuthHook is the correct,
//     minimal escape hatch: a direct port of legacy's basicTokenAuth
//     (internal/connectors/paypal-transaction/paypal_transaction.go).
//  2. StreamHook — the disputes stream's real wire pagination is a HATEOAS
//     links:[{rel,href}] ARRAY (matched by rel=="next"), not a bare string
//     path. The engine's only body-path next-page-URL pagination type
//     (pagination.type: next_url) reads a single dotted string path via
//     connsdk.StringAt/selectPath (interpolate.go/extract.go), which has no
//     array-element-matching-by-sibling-field grammar — a fixed numeric
//     array index would only ever read one array position, never "whichever
//     element has rel==next". This ports legacy's harvestPageToken loop
//     verbatim. transactions/balances/products remain fully declarative
//     (page_number/none pagination, streams.json) — only disputes needs
//     this hook.
//
// Secret values (client_id, client_secret, the cached access token) flow
// ONLY into the outgoing token-request Basic header or the Authorization
// header; they are never logged and never appear in an error string
// (THREAT-MODEL.md Delta 2).
//
// Line-count self-report (conventions.md §1's ~300-line soft target): this
// file is under the soft target. The two hook interfaces are independent,
// both-mandated shapes (a Basic-credential-in/Bearer-out token exchange,
// and a HATEOAS links-array pagination loop); neither alone would need a
// hook, but PayPal genuinely needs both simultaneously, which is why they
// share one package rather than splitting further (no 3rd interface is
// used, so no Tier-3 escalation is triggered).
package paypaltransaction

import (
	"context"
	"encoding/base64"
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
	tokenPath           = "/v1/oauth2/token"
	defaultDisputesSize = 50
)

func init() {
	engine.RegisterHooks("paypal-transaction", func() engine.Hooks { return New() })
}

// Hooks is the paypal-transaction hook set: AuthHook + StreamHook. Now/Client
// are test-injection overrides; every method is otherwise a pure function of
// its arguments (the returned Authenticator carries its own token cache,
// mirroring legacy's basicTokenAuth reuse across pages).
type Hooks struct {
	// Now is injectable for tests; nil uses time.Now.
	Now func() time.Time
	// Client overrides the HTTP client used for the token exchange and the
	// disputes StreamHook requests go through rt.Requester instead. Nil uses
	// a default client with a 30s timeout for the token exchange.
	Client *http.Client
}

// New returns a fresh paypal-transaction Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "paypal-transaction" }

var (
	_ engine.Hooks      = (*Hooks)(nil)
	_ engine.AuthHook   = (*Hooks)(nil)
	_ engine.StreamHook = (*Hooks)(nil)
)

// --- AuthHook ---------------------------------------------------------------

// Authenticator resolves the token-caching Bearer connsdk.Authenticator for
// spec (mode "custom", hook "paypal-transaction"), reading client_id/
// client_secret directly from cfg.Secrets and base_url from cfg.Config
// (mirrors jamf-pro's AuthHook: the custom-auth AuthSpec carries no
// templated fields of its own to resolve, so cfg is read directly rather
// than via engine.Interpolate).
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, _ engine.AuthSpec) (connsdk.Authenticator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.Config["base_url"]), "/")
	if baseURL == "" {
		return nil, errors.New("paypal-transaction connector requires config base_url")
	}
	clientID := ""
	clientSecret := ""
	if cfg.Secrets != nil {
		clientID = strings.TrimSpace(cfg.Secrets["client_id"])
		clientSecret = strings.TrimSpace(cfg.Secrets["client_secret"])
	}
	if clientID == "" || clientSecret == "" {
		return nil, errors.New("paypal-transaction connector requires secrets client_id and client_secret")
	}

	return &basicTokenAuth{
		TokenURL:     baseURL + tokenPath,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Client:       h.Client,
		Now:          h.Now,
	}, nil
}

// basicTokenAuth implements connsdk.Authenticator using PayPal's OAuth 2.0
// client-credentials grant with HTTP Basic client authentication — a direct
// port of legacy's basicTokenAuth
// (internal/connectors/paypal-transaction/paypal_transaction.go). The
// fetched bearer token is cached until 60s before its declared expiry;
// client_id/client_secret and the token itself are never logged.
type basicTokenAuth struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	// Client is used for the token request. Defaults to a 30s client.
	Client *http.Client
	// Now is injectable for tests. Defaults to time.Now.
	Now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *basicTokenAuth) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

// Apply ensures a fresh token and sets the Authorization header.
func (a *basicTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *basicTokenAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" && a.now().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.TokenURL) == "" {
		return "", errors.New("paypal-transaction oauth2: token URL is required")
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("paypal-transaction oauth2: build token request: %w", err)
	}
	creds := base64.StdEncoding.EncodeToString([]byte(a.ClientID + ":" + a.ClientSecret))
	req.Header.Set("Authorization", "Basic "+creds)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := a.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("paypal-transaction oauth2: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("paypal-transaction oauth2: token endpoint returned status %d", resp.StatusCode)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("paypal-transaction oauth2: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("paypal-transaction oauth2: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.now().Add(ttl)
	return a.token, nil
}

// --- StreamHook: disputes HATEOAS links[] pagination ------------------------

// ReadStream implements engine.StreamHook, handling only "disputes"
// (handled=true); every other stream name returns handled=false so the
// engine falls back to its own fully-declarative read path for
// transactions/balances/products.
func (h *Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if stream.Name != "disputes" {
		return false, nil
	}
	if err := ctx.Err(); err != nil {
		return true, err
	}

	maxPages, err := maxPagesFrom(req.Config)
	if err != nil {
		return true, err
	}
	return true, h.harvestDisputes(ctx, rt.Requester, maxPages, emit)
}

// harvestDisputes ports legacy's harvestPageToken loop verbatim for the
// disputes stream: the first request carries page_size=50 against
// /v1/customer/disputes; each subsequent page follows the HATEOAS
// links:[{rel,href}] entry whose rel=="next" (an absolute href) until no
// such entry is found, or maxPages (if set) is reached.
func (h *Hooks) harvestDisputes(ctx context.Context, r *connsdk.Requester, maxPages int, emit func(connectors.Record) error) error {
	path := "/v1/customer/disputes"
	query := url.Values{"page_size": []string{strconv.Itoa(defaultDisputesSize)}}

	for page := 0; maxPages <= 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("paypal-transaction: read disputes: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "items")
		if err != nil {
			return fmt.Errorf("paypal-transaction: decode disputes page: %w", err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(disputeRecord(item)); err != nil {
				return err
			}
		}

		next := nextLink(resp.Body)
		if next == "" {
			return nil
		}
		// Subsequent requests use the absolute next href verbatim; clear the
		// merged query so it is not duplicated onto the already-complete URL
		// (mirrors legacy's harvestPageToken: "path = next; query = url.Values{}").
		path = next
		query = url.Values{}
	}
	return nil
}

// nextLink extracts the rel="next" href from a PayPal HATEOAS links array
// (byte-for-byte port of legacy's nextLink,
// internal/connectors/paypal-transaction/paypal_transaction.go).
func nextLink(body []byte) string {
	var parsed struct {
		Links []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return ""
	}
	for _, l := range parsed.Links {
		if strings.EqualFold(l.Rel, "next") && strings.TrimSpace(l.Href) != "" {
			return l.Href
		}
	}
	return ""
}

// disputeRecord flattens a raw dispute item (byte-for-byte port of legacy's
// disputeRecord mapper).
func disputeRecord(item map[string]any) connectors.Record {
	amount := nestedObject(item, "dispute_amount")
	return connectors.Record{
		"dispute_id":                   item["dispute_id"],
		"reason":                       item["reason"],
		"status":                       item["status"],
		"dispute_state":                item["dispute_state"],
		"dispute_amount_currency_code": amount["currency_code"],
		"dispute_amount_value":         amount["value"],
		"create_time":                  item["create_time"],
		"update_time":                  item["update_time"],
	}
}

// nestedObject returns item[key] as a map, or an empty (non-nil) map when
// the key is absent or not an object, so field lookups stay nil-safe.
func nestedObject(item map[string]any, key string) map[string]any {
	if item == nil {
		return map[string]any{}
	}
	if obj, ok := item[key].(map[string]any); ok {
		return obj
	}
	return map[string]any{}
}

// maxPagesFrom mirrors legacy's paypalMaxPages
// (internal/connectors/paypal-transaction/paypal_transaction.go).
func maxPagesFrom(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("paypal-transaction config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("paypal-transaction config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}
