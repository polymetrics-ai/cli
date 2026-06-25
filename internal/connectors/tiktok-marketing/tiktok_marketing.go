// Package tiktokmarketing implements the native pm TikTok Marketing connector.
// It is a declarative-HTTP per-system connector built on the connsdk toolkit
// (Requester + a custom Access-Token header authenticator + RecordsAt extraction)
// composed with TikTok Business API stream definitions and endpoints.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect. The registry key is "tiktok-marketing".
//
// The connector is read-only: the TikTok Business API exposes campaign/ad
// management mutations, but there is no obviously safe reverse-ETL write surface
// for a generic sync, so Capabilities.Write is false.
package tiktokmarketing

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
	tiktokDefaultBaseURL  = "https://business-api.tiktok.com/open_api/v1.3"
	tiktokDefaultPageSize = 100
	tiktokMaxPageSize     = 1000
	tiktokUserAgent       = "polymetrics-go-cli"
	// accessTokenHeader is TikTok's custom auth header (not Bearer).
	accessTokenHeader = "Access-Token"
)

func init() {
	connectors.RegisterFactory("tiktok-marketing", New)
}

// New returns the TikTok Marketing connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm TikTok Marketing connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "tiktok-marketing" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "tiktok-marketing",
		DisplayName:     "TikTok Marketing",
		IntegrationType: "api",
		Description:     "Reads TikTok Business advertisers, campaigns, ad groups, and ads through the TikTok Marketing (Business) API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to TikTok. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := tiktokBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(tiktokAccessToken(cfg)) == "" {
		return errors.New("tiktok-marketing connector requires secret credentials.access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the advertiser info endpoint confirms auth and
	// connectivity without mutating anything. advertiser_id, when present, is
	// echoed; otherwise the call still validates the token shape.
	query := url.Values{"page": []string{"1"}, "page_size": []string{"1"}}
	if adv := strings.TrimSpace(cfg.Config["advertiser_id"]); adv != "" {
		query.Set("advertiser_ids", `["`+adv+`"]`)
	}
	resp, err := r.Do(ctx, http.MethodGet, "advertiser/info/", query, nil)
	if err != nil {
		return fmt.Errorf("check tiktok-marketing: %w", err)
	}
	if err := tiktokAPIError(resp.Body); err != nil {
		return fmt.Errorf("check tiktok-marketing: %w", err)
	}
	return nil
}

// Write is required by the connectors.Connector interface but is unsupported:
// this connector is read-only (Capabilities.Write is false).
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: tiktokStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "campaigns"
	}
	endpoint, ok := tiktokStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("tiktok-marketing stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := tiktokPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := tiktokMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, req.Config, endpoint, pageSize, maxPages, emit)
}

// harvest drives TikTok's page-number pagination. List endpoints return
// {code, message, data:{list:[...], page_info:{page, total_page, ...}}}; the next
// page is requested with page=page+1 until page >= total_page. The non-zero code
// envelope is surfaced as an error.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("page_size", strconv.Itoa(pageSize))
	if adv := strings.TrimSpace(cfg.Config["advertiser_id"]); adv != "" {
		base.Set("advertiser_id", adv)
		// advertiser/info/ takes a JSON-array advertiser_ids param instead.
		if endpoint.resource == "advertiser/info/" {
			base.Del("advertiser_id")
			base.Set("advertiser_ids", `["`+adv+`"]`)
		}
	}

	page := 1
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if maxPages > 0 && page > maxPages {
			return nil
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read tiktok-marketing %s: %w", endpoint.resource, err)
		}
		if err := tiktokAPIError(resp.Body); err != nil {
			return fmt.Errorf("read tiktok-marketing %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.listPath)
		if err != nil {
			return fmt.Errorf("decode tiktok-marketing %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		totalPage := tiktokTotalPage(resp.Body)
		// Stop when we've reached the last page, when the API reports no total
		// pages, or when a page came back empty (defensive against bad totals).
		if totalPage <= page || len(records) == 0 {
			return nil
		}
		page++
	}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise tiktok-marketing credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		idx := strconv.Itoa(i)
		item := map[string]any{
			"advertiser_id":    "adv_fixture_1",
			"advertiser_name":  "Fixture Advertiser " + idx,
			"company":          "Fixture Co " + idx,
			"status":           "STATUS_ENABLE",
			"currency":         "USD",
			"timezone":         "Etc/UTC",
			"country":          "US",
			"role":             "ROLE_ADVERTISER",
			"campaign_id":      "campaign_fixture_" + idx,
			"campaign_name":    "Fixture Campaign " + idx,
			"objective_type":   "TRAFFIC",
			"budget":           json100(i),
			"budget_mode":      "BUDGET_MODE_DAY",
			"operation_status": "ENABLE",
			"adgroup_id":       "adgroup_fixture_" + idx,
			"adgroup_name":     "Fixture AdGroup " + idx,
			"placement_type":   "PLACEMENT_TYPE_AUTOMATIC",
			"ad_id":            "ad_fixture_" + idx,
			"ad_name":          "Fixture Ad " + idx,
			"call_to_action":   "SHOP_NOW",
			"create_time":      "2026-01-0" + idx + "T00:00:00Z",
			"modify_time":      "2026-01-0" + idx + "T12:00:00Z",
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "tiktok-marketing"
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

// requester builds a connsdk.Requester wired with the custom Access-Token header
// authenticator and the resolved base URL. The secret only ever flows into the
// authenticator; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := tiktokBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := tiktokAccessToken(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("tiktok-marketing connector requires secret credentials.access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(accessTokenHeader, token, ""),
		UserAgent: tiktokUserAgent,
	}, nil
}

// tiktokAccessToken resolves the access token secret. The catalog secret field is
// dotted ("credentials.access_token"); a bare "access_token" key is also accepted
// as a convenience.
func tiktokAccessToken(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	if v := strings.TrimSpace(cfg.Secrets["credentials.access_token"]); v != "" {
		return v
	}
	return strings.TrimSpace(cfg.Secrets["access_token"])
}

// tiktokBaseURL resolves and validates the base URL. The default is
// business-api.tiktok.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func tiktokBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return tiktokDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("tiktok-marketing config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("tiktok-marketing config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("tiktok-marketing config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func tiktokPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return tiktokDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("tiktok-marketing config page_size must be an integer: %w", err)
	}
	if value < 1 || value > tiktokMaxPageSize {
		return 0, fmt.Errorf("tiktok-marketing config page_size must be between 1 and %d", tiktokMaxPageSize)
	}
	return value, nil
}

func tiktokMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("tiktok-marketing config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("tiktok-marketing config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// tiktokAPIError inspects the TikTok envelope code: a non-zero code is an API
// error and is surfaced with its message (never any secret).
func tiktokAPIError(body []byte) error {
	code, err := connsdk.StringAt(body, "code")
	if err != nil || code == "" || code == "0" {
		return nil
	}
	message, _ := connsdk.StringAt(body, "message")
	if strings.TrimSpace(message) == "" {
		message = "request failed"
	}
	return fmt.Errorf("tiktok api error code %s: %s", code, message)
}

// tiktokTotalPage reads data.page_info.total_page from the envelope, returning 0
// when absent or unparseable (which the caller treats as "stop after this page").
func tiktokTotalPage(body []byte) int {
	raw, err := connsdk.StringAt(body, "data.page_info.total_page")
	if err != nil || strings.TrimSpace(raw) == "" {
		return 0
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}
	return value
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

// json100 returns a small deterministic numeric budget for fixture records.
func json100(i int) int { return 100 * i }
