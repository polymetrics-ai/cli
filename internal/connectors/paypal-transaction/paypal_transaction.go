// Package paypaltransaction implements the native pm PayPal Transaction
// connector. It is a declarative-HTTP per-system connector built on the same
// shape as the stripe reference: a thin package that composes the connsdk
// toolkit (Requester + OAuth2 access-token auth + JSON extraction) with
// PayPal-specific stream definitions, endpoints, and pagination.
//
// PayPal authenticates with the OAuth 2.0 client-credentials grant: the
// client_id/client_secret are exchanged at /v1/oauth2/token (HTTP Basic) for a
// short-lived bearer access token, which is then sent on every data request.
// The reporting endpoints (transactions, balances) take start_date/end_date
// query params; the catalog and dispute listings do not.
//
// Like github/stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// The connector is read-only: PayPal exposes no safe generic reverse-ETL writes
// for this data set, so Capabilities.Write is false.
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
)

const (
	paypalProdBaseURL     = "https://api-m.paypal.com"
	paypalSandboxBaseURL  = "https://api-m.sandbox.paypal.com"
	paypalTokenPath       = "/v1/oauth2/token"
	paypalUserAgent       = "polymetrics-go-cli"
	paypalDefaultPageSize = 100
	// paypalFixtureDate is the deterministic ISO timestamp used by fixture-mode
	// records.
	paypalFixtureDate = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("paypal-transaction", New)
}

// New returns the PayPal Transaction connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm PayPal Transaction connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth2 token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "paypal-transaction" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "paypal-transaction",
		DisplayName:     "PayPal Transaction",
		IntegrationType: "api",
		Description:     "Reads PayPal transactions, balances, catalog products, and customer disputes through the PayPal REST API using OAuth 2.0 client-credentials auth.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to PayPal. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := paypalBaseURL(cfg); err != nil {
		return err
	}
	id, secret := paypalCreds(cfg)
	if strings.TrimSpace(id) == "" || strings.TrimSpace(secret) == "" {
		return errors.New("paypal-transaction connector requires secrets client_id and client_secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the balances endpoint confirms the OAuth2 token exchange
	// and connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "v1/reporting/balances", nil, nil, nil); err != nil {
		return fmt.Errorf("check paypal-transaction: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: paypalStreams()}, nil
}

// Write is unsupported: PayPal exposes no safe generic reverse-ETL writes for
// this data set, so Capabilities.Write is false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a stream starts with an
// empty incremental cursor (full sync), which the start_date config raises at
// read time.
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
		stream = "transactions"
	}
	endpoint, ok := paypalStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("paypal-transaction stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := paypalMaxPages(req.Config)
	if err != nil {
		return err
	}
	base, err := c.baseQuery(req, endpoint)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, base, maxPages, emit)
}

// baseQuery builds the query params common to all pages of a stream. Reporting
// endpoints require an RFC3339 start_date/end_date window.
func (c Connector) baseQuery(req connectors.ReadRequest, endpoint streamEndpoint) (url.Values, error) {
	q := url.Values{}
	if !endpoint.dateRange {
		return q, nil
	}
	start := lowerBound(req)
	if start == "" {
		return nil, errors.New("paypal-transaction config start_date is required for reporting streams")
	}
	if _, err := time.Parse(time.RFC3339, start); err != nil {
		return nil, fmt.Errorf("paypal-transaction config start_date must be RFC3339: %w", err)
	}
	q.Set("start_date", start)

	end := strings.TrimSpace(req.Config.Config["end_date"])
	if end == "" {
		end = time.Now().UTC().Format(time.RFC3339)
	} else if _, err := time.Parse(time.RFC3339, end); err != nil {
		return nil, fmt.Errorf("paypal-transaction config end_date must be RFC3339: %w", err)
	}
	q.Set("end_date", end)
	if endpoint.resource == "v1/reporting/transactions" {
		q.Set("fields", "all")
	}
	return q, nil
}

// harvest drives the per-stream pagination style and emits each mapped record.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, maxPages int, emit func(connectors.Record) error) error {
	switch endpoint.pagination {
	case paginationPageToken:
		return c.harvestPageToken(ctx, r, endpoint, base, maxPages, emit)
	case paginationPageIncrement:
		return c.harvestPageIncrement(ctx, r, endpoint, base, maxPages, emit)
	default:
		return c.harvestSingle(ctx, r, endpoint, base, emit)
	}
}

// harvestSingle reads one page (no pagination).
func (c Connector) harvestSingle(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, base, nil)
	if err != nil {
		return fmt.Errorf("read paypal-transaction %s: %w", endpoint.resource, err)
	}
	return emitRecords(ctx, resp.Body, endpoint, emit)
}

// harvestPageIncrement walks page=1..total_pages. PayPal reporting and catalog
// endpoints return {..., page:N, total_pages:M}.
func (c Connector) harvestPageIncrement(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, maxPages int, emit func(connectors.Record) error) error {
	pageSize := paypalDefaultPageSize
	if endpoint.resource == "v1/catalogs/products" {
		pageSize = 20 // products endpoint caps page_size at 20
	}
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))
		query.Set("page_size", strconv.Itoa(pageSize))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read paypal-transaction %s page %d: %w", endpoint.resource, page, err)
		}
		if err := emitRecords(ctx, resp.Body, endpoint, emit); err != nil {
			return err
		}
		total, terr := connsdk.StringAt(resp.Body, "total_pages")
		if terr != nil {
			return fmt.Errorf("decode paypal-transaction %s total_pages: %w", endpoint.resource, terr)
		}
		totalPages, perr := strconv.Atoi(strings.TrimSpace(total))
		if perr != nil || totalPages <= 0 {
			// No total_pages signal: stop after the first page.
			return nil
		}
		if page >= totalPages {
			return nil
		}
	}
	return nil
}

// harvestPageToken follows the HATEOAS links[rel=next] href that PayPal disputes
// return.
func (c Connector) harvestPageToken(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, maxPages int, emit func(connectors.Record) error) error {
	path := endpoint.resource
	query := cloneValues(base)
	query.Set("page_size", "50")
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read paypal-transaction %s: %w", endpoint.resource, err)
		}
		if err := emitRecords(ctx, resp.Body, endpoint, emit); err != nil {
			return err
		}
		next := nextLink(resp.Body)
		if next == "" {
			return nil
		}
		// Subsequent requests use the absolute next href; clear merged query so it
		// is not duplicated onto the already-complete URL.
		path = next
		query = url.Values{}
	}
	return nil
}

// emitRecords extracts the records at the endpoint's path and emits each mapped
// record.
func emitRecords(ctx context.Context, body []byte, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	records, err := connsdk.RecordsAt(body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode paypal-transaction %s: %w", endpoint.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise paypal-transaction credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := fixtureItem(stream, i)
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

// fixtureItem builds a raw, pre-flatten PayPal object for fixture mode shaped
// like the real API responses so the same mappers apply.
func fixtureItem(stream string, i int) map[string]any {
	idx := strconv.Itoa(i)
	money := func(code, val string) map[string]any {
		return map[string]any{"currency_code": code, "value": val}
	}
	switch stream {
	case "balances":
		return map[string]any{
			"currency":          "USD",
			"primary":           i == 1,
			"total_balance":     money("USD", "100.0"+idx),
			"available_balance": money("USD", "90.0"+idx),
			"withheld_balance":  money("USD", "10.0"+idx),
		}
	case "products":
		return map[string]any{
			"id":          "PROD-fixture-" + idx,
			"name":        "Fixture Product " + idx,
			"description": "fixture",
			"type":        "SERVICE",
			"category":    "SOFTWARE",
			"create_time": paypalFixtureDate,
		}
	case "disputes":
		return map[string]any{
			"dispute_id":     "PP-D-fixture-" + idx,
			"reason":         "MERCHANDISE_OR_SERVICE_NOT_RECEIVED",
			"status":         "RESOLVED",
			"dispute_state":  "RESOLVED",
			"dispute_amount": money("USD", "12.3"+idx),
			"create_time":    paypalFixtureDate,
			"update_time":    paypalFixtureDate,
		}
	default: // transactions
		return map[string]any{
			"transaction_info": map[string]any{
				"transaction_id":              "T-fixture-" + idx,
				"transaction_status":          "S",
				"transaction_event_code":      "T0006",
				"transaction_initiation_date": paypalFixtureDate,
				"transaction_updated_date":    paypalFixtureDate,
				"transaction_amount":          money("USD", "10.0"+idx),
				"fee_amount":                  money("USD", "-0.5"+idx),
				"paypal_account_id":           "ACC-fixture",
			},
		}
	}
}

// requester builds a connsdk.Requester wired with OAuth2 client-credentials auth
// and the resolved base URL. Secrets only ever flow into the authenticator; they
// are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := paypalBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	id, secret := paypalCreds(cfg)
	if strings.TrimSpace(id) == "" || strings.TrimSpace(secret) == "" {
		return nil, errors.New("paypal-transaction connector requires secrets client_id and client_secret")
	}
	auth := &basicTokenAuth{
		tokenURL:     paypalTokenURL(cfg, base),
		clientID:     id,
		clientSecret: secret,
		client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: paypalUserAgent,
	}, nil
}

// basicTokenAuth fetches and caches a PayPal OAuth 2.0 access token using the
// client-credentials grant. Unlike connsdk.OAuth2ClientCredentials, PayPal's
// token endpoint authenticates the client with HTTP Basic, so this small
// in-package authenticator sends the credentials in the Authorization header.
type basicTokenAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	client       *http.Client

	mu      sync.Mutex
	token   string
	expires time.Time
	now     func() time.Time
}

func (a *basicTokenAuth) clock() time.Time {
	if a.now != nil {
		return a.now()
	}
	return time.Now()
}

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
	if a.token != "" && a.clock().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("paypal-transaction oauth2: token URL is required")
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("paypal-transaction oauth2: build token request: %w", err)
	}
	creds := base64.StdEncoding.EncodeToString([]byte(a.clientID + ":" + a.clientSecret))
	req.Header.Set("Authorization", "Basic "+creds)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("paypal-transaction oauth2: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("paypal-transaction oauth2: token endpoint returned %s", resp.Status)
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
	a.expires = a.clock().Add(ttl)
	return a.token, nil
}

// paypalCreds resolves the client id/secret from secrets.
func paypalCreds(cfg connectors.RuntimeConfig) (string, string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return cfg.Secrets["client_id"], cfg.Secrets["client_secret"]
}

// paypalBaseURL resolves and validates the base URL. The default is the PayPal
// production host (or sandbox when is_sandbox is set); any base_url override must
// be an absolute https (or http for local test servers) URL with a host to bound
// SSRF risk.
func paypalBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		if isSandbox(cfg) {
			return paypalSandboxBaseURL, nil
		}
		return paypalProdBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("paypal-transaction config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("paypal-transaction config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("paypal-transaction config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// paypalTokenURL resolves the OAuth2 token endpoint: an explicit token_url
// override (used by tests) wins, otherwise it is derived from the base host.
func paypalTokenURL(cfg connectors.RuntimeConfig, base string) string {
	if override := strings.TrimSpace(cfg.Config["token_url"]); override != "" {
		return override
	}
	return strings.TrimRight(base, "/") + paypalTokenPath
}

func isSandbox(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["is_sandbox"]), "true")
}

func paypalMaxPages(cfg connectors.RuntimeConfig) (int, error) {
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

// lowerBound returns the RFC3339 start bound for reporting streams, from the
// incremental cursor (if any) or else the start_date config.
func lowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
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

// nextLink extracts the rel="next" href from a PayPal HATEOAS links array.
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
