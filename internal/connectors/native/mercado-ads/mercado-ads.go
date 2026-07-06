// Package mercadoads implements the native pm Mercado Ads connector. It follows
// the declarative-HTTP template established by the stripe connector: a thin
// package composing the connsdk toolkit (Requester + record extraction +
// offset pagination) with Mercado-Ads-specific stream definitions and an OAuth2
// refresh-token authenticator.
//
// The directory is internal/connectors/mercado-ads (hyphenated, the bare system
// package blank-imports this package in the production binary to run that side
// effect.
//
// Data source: the Mercado Libre Advertising API
// (https://api.mercadolibre.com/advertising/...). Auth is OAuth2 with the
// refresh_token grant against https://api.mercadolibre.com/oauth/token. This
// connector is read-only.
package mercadoads

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
)

const (
	defaultBaseURL  = "https://api.mercadolibre.com"
	defaultTokenURL = "https://api.mercadolibre.com/oauth/token"
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
	apiVersion      = "1"
)

// New returns the Mercado Ads connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Mercado Ads connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "mercado-ads" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "mercado-ads",
		DisplayName:     "Mercado Ads",
		IntegrationType: "api",
		Description:     "Reads Mercado Ads brand, display, and product advertisers and daily campaign metrics from the Mercado Libre Advertising API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Mercado Ads.
// In fixture mode it short-circuits without a network call.
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
	// A bounded read of the brand advertisers list confirms the OAuth refresh and
	// connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "advertising/advertisers", url.Values{"product_id": []string{"BADS"}}, nil, nil); err != nil {
		return fmt.Errorf("check mercado-ads: %w", err)
	}
	return nil
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
		stream = "brand_advertisers"
	}
	def, ok := streamDefs[stream]
	if !ok {
		return fmt.Errorf("mercado-ads stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, def, req, emit)
	}

	if err := requireSecrets(req.Config); err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPagesConfig(req.Config)
	if err != nil {
		return err
	}

	switch def.kind {
	case advertisersKind:
		return c.readAdvertisers(ctx, r, def, emit)
	case metricsKind:
		return c.readMetrics(ctx, r, stream, def, req, maxPages, emit)
	default:
		return fmt.Errorf("mercado-ads stream %q has unknown kind", stream)
	}
}

// readAdvertisers performs the single GET /advertising/advertisers request,
// filtered by product_id, with the Api-Version header.
func (c Connector) readAdvertisers(ctx context.Context, r *connsdk.Requester, def streamDef, emit func(connectors.Record) error) error {
	query := url.Values{"product_id": []string{def.productID}}
	resp, err := r.Do(ctx, http.MethodGet, "advertising/advertisers", query, nil)
	if err != nil {
		return fmt.Errorf("read mercado-ads advertisers: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, def.recordsPath)
	if err != nil {
		return fmt.Errorf("decode mercado-ads advertisers: %w", err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(def.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// readMetrics walks the offset/limit paginated per-campaign metrics endpoint.
// The advertiser_id and campaign_id come from req.State (the substream partition
// keys a real sync would have resolved from the parent campaigns stream).
func (c Connector) readMetrics(ctx context.Context, r *connsdk.Requester, stream string, def streamDef, req connectors.ReadRequest, maxPages int, emit func(connectors.Record) error) error {
	advertiserID := strings.TrimSpace(req.State["advertiser_id"])
	campaignID := strings.TrimSpace(req.State["campaign_id"])
	if advertiserID == "" || campaignID == "" {
		return errors.New("mercado-ads metrics streams require advertiser_id and campaign_id in state (substream partition keys)")
	}
	tmpl, ok := metricsEndpointTemplate[stream]
	if !ok {
		return fmt.Errorf("mercado-ads stream %q has no metrics endpoint", stream)
	}
	path := fmt.Sprintf(tmpl, url.PathEscape(advertiserID), url.PathEscape(campaignID))

	base := url.Values{}
	if dateFrom := strings.TrimSpace(req.Config.Config["start_date"]); dateFrom != "" {
		base.Set("date_from", dateFrom)
	}
	if dateTo := strings.TrimSpace(req.Config.Config["end_date"]); dateTo != "" {
		base.Set("date_to", dateTo)
	}

	pager := &connsdk.OffsetPaginator{
		LimitParam:  "limit",
		OffsetParam: "offset",
		PageSize:    defaultPageSize,
	}
	return connsdk.Harvest(ctx, r, http.MethodGet, path, base, pager, def.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(def.mapRecord(rec))
	})
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise mercado-ads credential-free (mirrors the
// stripe fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, def streamDef, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var item map[string]any
		if def.kind == metricsKind {
			item = map[string]any{
				"date":          fmt.Sprintf("2026-01-0%d", i),
				"advertiser_id": "111",
				"campaign_id":   "999",
				"prints":        json.Number(strconv.Itoa(100 * i)),
				"clicks":        json.Number(strconv.Itoa(5 * i)),
				"cost":          json.Number(strconv.Itoa(2 * i)),
				"ctr":           json.Number("0.05"),
				"cpc":           json.Number("0.4"),
			}
		} else {
			item = map[string]any{
				"advertiser_id":   json.Number(strconv.Itoa(100 + i)),
				"advertiser_name": fmt.Sprintf("Fixture Advertiser %d", i),
				"account_name":    fmt.Sprintf("fixture-account-%d", i),
				"site_id":         "MLB",
			}
		}
		record := def.mapRecord(item)
		record["_stream"] = stream
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the OAuth2 refresh-token
// authenticator, the resolved base URL, and the Api-Version header. The secrets
// only ever flow into the authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	tokenURL, err := tokenURLConfig(cfg)
	if err != nil {
		return nil, err
	}
	auth := &refreshTokenAuth{
		tokenURL:     tokenURL,
		clientID:     cfg.Secrets["client_id"],
		clientSecret: cfg.Secrets["client_secret"],
		refreshToken: cfg.Secrets["client_refresh_token"],
		client:       c.Client,
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           auth,
		UserAgent:      userAgent,
		DefaultHeaders: map[string]string{"Api-Version": apiVersion},
	}, nil
}

// refreshTokenAuth fetches and caches a bearer token using the OAuth2
// refresh_token grant against the Mercado Libre token endpoint. connsdk only
// ships client-credentials, so this small in-package authenticator handles the
// refresh-token flow. It refreshes 60s before expiry and never logs secrets.
type refreshTokenAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	refreshToken string
	client       *http.Client

	mu      sync.Mutex
	token   string
	expires time.Time
	now     func() time.Time
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

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)
	form.Set("refresh_token", a.refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("mercado-ads oauth: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("mercado-ads oauth: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("mercado-ads oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("mercado-ads oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("mercado-ads oauth: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.timeNow().Add(ttl)
	return a.token, nil
}

func requireSecrets(cfg connectors.RuntimeConfig) error {
	if cfg.Secrets == nil {
		return errors.New("mercado-ads connector requires secrets client_id, client_secret, client_refresh_token")
	}
	for _, key := range []string{"client_id", "client_secret", "client_refresh_token"} {
		if strings.TrimSpace(cfg.Secrets[key]) == "" {
			return fmt.Errorf("mercado-ads connector requires secret %s", key)
		}
	}
	return nil
}

// baseURL resolves and validates the base URL. The default is
// api.mercadolibre.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validateURL(cfg.Config["base_url"], defaultBaseURL, "base_url")
}

// tokenURLConfig resolves and validates the OAuth token endpoint URL.
func tokenURLConfig(cfg connectors.RuntimeConfig) (string, error) {
	return validateURL(cfg.Config["token_url"], defaultTokenURL, "token_url")
}

func validateURL(raw, def, field string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("mercado-ads config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("mercado-ads config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("mercado-ads config %s must include a host", field)
	}
	return strings.TrimRight(raw, "/"), nil
}

func maxPagesConfig(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mercado-ads config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("mercado-ads config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Mercado Ads is read-only
// for reverse ETL, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
