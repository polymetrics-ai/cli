// Package ebayfulfillment implements the ebay-fulfillment bundle's Tier-2
// hooks (conventions.md §1): an AuthHook porting legacy auth.go's
// refreshTokenAuth (OAuth 2.0 refresh-token grant, HTTP-Basic client auth)
// almost verbatim, and a StreamHook covering 2 streams that re-project GET
// /sell/fulfillment/v1/order beyond schema/computed_fields: order_line_items
// explodes lineItems[] into one record per line item (no declarative "1
// record becomes N"), and shipping_fulfillments projects
// fulfillmentStartInstructions[0] (an array-INDEX access computed_fields
// cannot express). Both are legitimate Tier-2 StreamHook triggers (§1:
// "sub-resource fan-out reads"); orders/payment_disputes need neither and
// stay declarative (docs.md). Two hook interfaces, at the Tier-2 cap.
// Secrets flow ONLY into the token-request form/Authorization header.
package ebayfulfillment

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
	orderResource   = "/sell/fulfillment/v1/order"
	orderRecordsKey = "orders"
	defaultLimit    = 50
)

func init() {
	engine.RegisterHooks("ebay-fulfillment", func() engine.Hooks { return New() })
}

// Hooks is the ebay-fulfillment hook set: AuthHook + StreamHook.
type Hooks struct {
	Now    func() time.Time // injectable for tests; nil uses time.Now
	Client *http.Client     // overrides the token-exchange client; nil uses a 30s-timeout default
}

// New returns a fresh ebay-fulfillment Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "ebay-fulfillment" }

// --- AuthHook ---

// Authenticator resolves the refresh-token-grant Authenticator for spec
// (mode "custom"). spec.Token is the refresh token (mirrors gmail's
// identical AuthSpec-field mapping).
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, spec engine.AuthSpec) (connsdk.Authenticator, error) {
	tokenURL, err := interpolateRequired(spec.TokenURL, "token_url", cfg)
	if err != nil {
		return nil, err
	}
	if err := validateHTTPSURL(tokenURL, "token_url"); err != nil {
		return nil, err
	}
	refreshToken, err := interpolateRequired(spec.Token, "refresh_token", cfg)
	if err != nil {
		return nil, err
	}
	// client_id/client_secret/scope are optional (legacy sends Basic only if
	// either is set, scope only if set); resolved best-effort since
	// engine.Interpolate hard-errors on an absent key outside "when" (§3).
	clientID := interpolateOptional(spec.ClientID, cfg)
	clientSecret := interpolateOptional(spec.ClientSecret, cfg)
	scope := interpolateOptional(spec.Scopes, cfg)

	return &refreshTokenAuth{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		scopes:       scope,
		client:       h.Client,
		now:          h.Now,
	}, nil
}

func interpolateRequired(tmpl, field string, cfg connectors.RuntimeConfig) (string, error) {
	vars := engine.Vars{Config: cfg.Config, Secrets: cfg.Secrets}
	val, err := engine.Interpolate(tmpl, vars)
	if err != nil {
		return "", fmt.Errorf("ebay-fulfillment oauth: resolve %s: %w", field, err)
	}
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("ebay-fulfillment oauth: %s is required", field)
	}
	return val, nil
}

// interpolateOptional resolves tmpl best-effort: any error (absent key —
// the intended case — or a CRLF/unknown-filter failure) resolves to "".
// Verified benign at its call sites: client_id/client_secret/scope, all
// optional-when-empty OAuth token-request form values.
func interpolateOptional(tmpl string, cfg connectors.RuntimeConfig) string {
	if strings.TrimSpace(tmpl) == "" {
		return ""
	}
	vars := engine.Vars{Config: cfg.Config, Secrets: cfg.Secrets}
	val, err := engine.Interpolate(tmpl, vars)
	if err != nil {
		return ""
	}
	return val
}

// validateHTTPSURL fails closed on anything but a well-formed https:// URL
// with a host (SSRF hardening; stricter than legacy's http-tolerant
// validateURL — documented, never-limiting deviation, docs.md).
func validateHTTPSURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("ebay-fulfillment oauth: %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("ebay-fulfillment oauth: %s must use https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("ebay-fulfillment oauth: %s must include a host", field)
	}
	return nil
}

// refreshTokenAuth ports legacy's refreshTokenAuth field-for-field: exchange
// the refresh token for an access token at tokenURL (HTTP Basic
// clientID:clientSecret), cache until 60s before expiry, set Authorization:
// Bearer <token> on each request.
type refreshTokenAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	refreshToken string
	scopes       string
	client       *http.Client
	now          func() time.Time
	mu           sync.Mutex
	token        string
	expires      time.Time
}

func (a *refreshTokenAuth) timeNow() time.Time {
	if a.now != nil {
		return a.now()
	}
	return time.Now()
}

func (a *refreshTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *refreshTokenAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" && a.timeNow().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("ebay-fulfillment oauth: token URL is required")
	}
	if strings.TrimSpace(a.refreshToken) == "" {
		return "", errors.New("ebay-fulfillment oauth: refresh_token is required")
	}
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", a.refreshToken)
	if a.scopes != "" {
		form.Set("scope", a.scopes)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("ebay-fulfillment oauth: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")
	if a.clientID != "" || a.clientSecret != "" {
		creds := base64.StdEncoding.EncodeToString([]byte(a.clientID + ":" + a.clientSecret))
		httpReq.Header.Set("Authorization", "Basic "+creds)
	}
	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ebay-fulfillment oauth: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ebay-fulfillment oauth: token endpoint returned %s", resp.Status)
	}
	var out struct {
		AccessToken string      `json:"access_token"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("ebay-fulfillment oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("ebay-fulfillment oauth: token response missing access_token")
	}
	a.token = out.AccessToken
	ttl := 7200 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.timeNow().Add(ttl)
	return a.token, nil
}

// --- StreamHook ---

// orderShapedStreams is the set of stream names this hook owns.
var orderShapedStreams = map[string]bool{
	"order_line_items":      true,
	"shipping_fulfillments": true,
}

// ReadStream implements engine.StreamHook, mirroring legacy's harvest +
// emitProjected: fetch GET /sell/fulfillment/v1/order (following "next", or
// advancing offset on a short-page stop), re-projecting per stream name.
func (h *Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	name := stream.Name
	if !orderShapedStreams[name] {
		return false, nil
	}
	if err := ctx.Err(); err != nil {
		return true, err
	}
	limit := pageSize(req.Config)
	base := url.Values{}
	base.Set("limit", strconv.Itoa(limit))
	if filter := dateFilter(req); filter != "" {
		base.Set("filter", filter)
	}

	path := orderResource
	query := cloneValues(base)
	offset := 0
	for {
		if err := ctx.Err(); err != nil {
			return true, err
		}
		resp, err := rt.Requester.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return true, fmt.Errorf("read ebay-fulfillment %s: %w", name, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, orderRecordsKey)
		if err != nil {
			return true, fmt.Errorf("decode ebay-fulfillment %s page: %w", name, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return true, err
			}
			if err := emitOrderProjection(name, map[string]any(item), emit); err != nil {
				return true, err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next")
		if err != nil {
			return true, fmt.Errorf("decode ebay-fulfillment %s next: %w", name, err)
		}
		// Follow eBay's absolute "next" URL verbatim (legacy's harvest
		// performs no host validation here either).
		if next = strings.TrimSpace(next); next != "" {
			path, query = next, url.Values{}
			continue
		}
		if len(records) < limit {
			return true, nil
		}
		offset += limit
		path, query = orderResource, cloneValues(base)
		query.Set("offset", strconv.Itoa(offset))
	}
}

// emitOrderProjection maps a raw order item into one or more records
// depending on the requested stream, mirroring legacy's emitProjected.
func emitOrderProjection(stream string, item map[string]any, emit func(connectors.Record) error) error {
	if stream == "shipping_fulfillments" {
		return emit(shippingFulfillmentRecord(item))
	}
	lineItems, _ := item["lineItems"].([]any)
	for _, li := range lineItems {
		lineItem, ok := li.(map[string]any)
		if !ok {
			continue
		}
		if err := emit(lineItemRecord(item, lineItem)); err != nil {
			return err
		}
	}
	return nil
}

// lineItemRecord flattens one lineItem into the order_line_items shape,
// carrying the parent order's identity/date — ports legacy verbatim.
func lineItemRecord(order, lineItem map[string]any) connectors.Record {
	rec := connectors.Record{
		"line_item_id":                 lineItem["lineItemId"],
		"order_id":                     order["orderId"],
		"legacy_item_id":               lineItem["legacyItemId"],
		"sku":                          lineItem["sku"],
		"title":                        lineItem["title"],
		"quantity":                     lineItem["quantity"],
		"line_item_fulfillment_status": lineItem["lineItemFulfillmentStatus"],
		"creation_date":                order["creationDate"],
	}
	if total, ok := lineItem["total"].(map[string]any); ok {
		rec["total_value"] = total["value"]
		rec["total_currency"] = total["currency"]
	}
	return rec
}

// shippingFulfillmentRecord projects an order into a shipment-oriented row
// using fulfillmentStartInstructions[0]/shippingStep — ports legacy verbatim.
func shippingFulfillmentRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"order_id":                 item["orderId"],
		"legacy_order_id":          item["legacyOrderId"],
		"order_fulfillment_status": item["orderFulfillmentStatus"],
		"creation_date":            item["creationDate"],
	}
	fsi, _ := item["fulfillmentStartInstructions"].([]any)
	if len(fsi) == 0 {
		return rec
	}
	first, _ := fsi[0].(map[string]any)
	rec["shipping_step"] = first["fulfillmentInstructionsType"]
	ship, _ := first["shippingStep"].(map[string]any)
	to, _ := ship["shipTo"].(map[string]any)
	if to == nil {
		return rec
	}
	if fullName, ok := to["fullName"]; ok {
		rec["ship_to_name"] = fullName
	}
	if addr, ok := to["contactAddress"].(map[string]any); ok {
		rec["ship_to_city"] = addr["city"]
		rec["ship_to_state_or_province"] = addr["stateOrProvince"]
		rec["ship_to_postal_code"] = addr["postalCode"]
		rec["ship_to_country_code"] = addr["countryCode"]
	}
	return rec
}

// dateFilter builds the eBay creationdate filter from the incremental
// cursor or start_date config — ports legacy's dateFilter verbatim.
func dateFilter(req connectors.ReadRequest) string {
	lower := connsdk.Cursor(req.State)
	if lower == "" {
		lower = strings.TrimSpace(req.Config.Config["start_date"])
	}
	if lower == "" {
		return ""
	}
	return "creationdate:[" + lower + "..]"
}

// pageSize resolves config.page_size (1-1000, default 50; legacy's
// resolvePageSize bounds). An invalid value falls back to the default.
func pageSize(cfg connectors.RuntimeConfig) int {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultLimit
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > 1000 {
		return defaultLimit
	}
	return value
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
