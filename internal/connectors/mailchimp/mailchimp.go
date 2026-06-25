// Package mailchimp implements the native pm Mailchimp Marketing API connector.
// It follows the declarative-HTTP per-system connector shape established by the
// stripe package: a thin package that composes the connsdk toolkit (Requester +
// Basic/Bearer auth + RecordsAt extraction + offset pagination) with
// Mailchimp-specific stream definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Mailchimp's Marketing API is datacenter-scoped: the host is
// https://<dc>.api.mailchimp.com/3.0 where <dc> is the suffix of the API key
// (e.g. "us6" in "abc123-us6") or the data_center config field. Auth is either
// HTTP Basic (username "anystring", password = API key) or a Bearer OAuth access
// token. Pagination is offset/count, with total_items in each response.
package mailchimp

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
	mailchimpDefaultPageSize = 100
	mailchimpMaxPageSize     = 1000
	mailchimpUserAgent       = "polymetrics-go-cli"
	mailchimpBasicUsername   = "anystring"
)

func init() {
	connectors.RegisterFactory("mailchimp", New)
}

// New returns the Mailchimp connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Mailchimp connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "mailchimp" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "mailchimp",
		DisplayName:     "Mailchimp",
		IntegrationType: "api",
		Description:     "Reads Mailchimp Marketing API audiences (lists), campaigns, reports, and automations through the datacenter-scoped REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Mailchimp.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := mailchimpBaseURL(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the lists endpoint confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "lists", url.Values{"count": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check mailchimp: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: mailchimpStreams()}, nil
}

// Write satisfies the connectors.Connector interface. The Mailchimp connector is
// read-only (no allow-listed reverse-ETL actions), so writes are unsupported.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a Mailchimp stream starts
// with an empty incremental cursor (full sync); the start_date config can raise
// the lower bound at read time.
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
		stream = "lists"
	}
	endpoint, ok := mailchimpStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("mailchimp stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := mailchimpPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := mailchimpMaxPages(req.Config)
	if err != nil {
		return err
	}
	since := incrementalLowerBound(req)
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, since, emit)
}

// harvest drives Mailchimp's offset/count pagination. Mailchimp list endpoints
// return {<recordsKey>:[...], total_items:N}; the next page is requested by
// advancing offset by the page size until offset >= total_items or a short page
// is returned. The loop lives here (rather than connsdk.Harvest) because the
// stop condition reads total_items out of the body.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, since string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("count", strconv.Itoa(pageSize))
	if since != "" {
		// Mailchimp uses since_* filters scoped per endpoint; lists/campaigns
		// support since_date_created. Apply when present to bound the sync.
		base.Set(sinceParam(endpoint.resource), since)
	}

	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read mailchimp %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode mailchimp %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		count := len(records)
		offset += count
		total := mailchimpTotalItems(resp.Body)
		// Stop on a short page (fewer than requested) or once we've consumed
		// every item the server reports.
		if count == 0 || count < pageSize || (total > 0 && offset >= total) {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise mailchimp credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":             fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"web_id":         i,
			"name":           fmt.Sprintf("Fixture %d", i),
			"date_created":   "2026-01-01T00:00:00+00:00",
			"create_time":    "2026-01-01T00:00:00+00:00",
			"start_time":     "2026-01-01T00:00:00+00:00",
			"send_time":      "2026-01-02T00:00:00+00:00",
			"type":           "regular",
			"status":         "sent",
			"emails_sent":    100 * i,
			"campaign_title": fmt.Sprintf("Fixture Campaign %d", i),
			"list_id":        "lists_fixture_1",
			"abuse_reports":  0,
			"unsubscribed":   i,
			"connector":      "mailchimp",
			"fixture":        true,
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

// requester builds a connsdk.Requester wired with the resolved datacenter base
// URL and the credential-appropriate authenticator. The secret only ever flows
// into connsdk.Basic/connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := mailchimpBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	auth, err := mailchimpAuth(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: mailchimpUserAgent,
	}, nil
}

// mailchimpAuth resolves the authenticator from the configured secrets. An OAuth
// access_token uses Bearer; an API key uses HTTP Basic with the conventional
// "anystring" username.
func mailchimpAuth(cfg connectors.RuntimeConfig) (connsdk.Authenticator, error) {
	if token := mailchimpAccessToken(cfg); token != "" {
		return connsdk.Bearer(token), nil
	}
	if apiKey := mailchimpAPIKey(cfg); apiKey != "" {
		return connsdk.Basic(mailchimpBasicUsername, apiKey), nil
	}
	return nil, errors.New("mailchimp connector requires secret credentials.access_token or credentials.apikey")
}

func mailchimpAccessToken(cfg connectors.RuntimeConfig) string {
	return strings.TrimSpace(secret(cfg, "credentials.access_token"))
}

func mailchimpAPIKey(cfg connectors.RuntimeConfig) string {
	return strings.TrimSpace(secret(cfg, "credentials.apikey"))
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

// mailchimpBaseURL resolves and validates the datacenter-scoped base URL. The
// base_url config overrides everything (used by tests/proxies); otherwise the
// host is derived from the data_center config or the API key suffix
// (e.g. "abc-us6" -> dc "us6"). Any override must be an absolute http/https URL
// with a host to bound SSRF risk.
func mailchimpBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
		parsed, err := url.Parse(base)
		if err != nil {
			return "", fmt.Errorf("mailchimp config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("mailchimp config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("mailchimp config base_url must include a host")
		}
		return strings.TrimRight(base, "/"), nil
	}

	dc := mailchimpDataCenter(cfg)
	if dc == "" {
		return "", errors.New("mailchimp connector requires a data_center config or an API key with a datacenter suffix (e.g. key-us6)")
	}
	if !validDataCenter(dc) {
		return "", fmt.Errorf("mailchimp data center %q is invalid", dc)
	}
	return fmt.Sprintf("https://%s.api.mailchimp.com/3.0", dc), nil
}

// mailchimpDataCenter returns the datacenter token, preferring the explicit
// data_center config and falling back to the API key suffix after the last "-".
func mailchimpDataCenter(cfg connectors.RuntimeConfig) string {
	if dc := strings.TrimSpace(cfg.Config["data_center"]); dc != "" {
		return strings.ToLower(dc)
	}
	apiKey := mailchimpAPIKey(cfg)
	if idx := strings.LastIndex(apiKey, "-"); idx >= 0 && idx < len(apiKey)-1 {
		return strings.ToLower(apiKey[idx+1:])
	}
	return ""
}

// validDataCenter bounds the datacenter token to a safe host label (letters and
// digits only) so it cannot inject characters into the URL host.
func validDataCenter(dc string) bool {
	if dc == "" || len(dc) > 16 {
		return false
	}
	for _, r := range dc {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

// incrementalLowerBound returns the lower-bound timestamp for the since_* filter,
// derived from the incremental cursor (if any) or else the start_date config.
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

// sinceParam returns the endpoint-specific since_* filter param. Mailchimp scopes
// these per resource (since_date_created for lists, since_create_time for
// campaigns). Endpoints without a known since filter return an empty resource
// suffix, which is harmless because callers only use it when since != "".
func sinceParam(resource string) string {
	switch resource {
	case "lists":
		return "since_date_created"
	case "campaigns", "automations":
		return "since_create_time"
	case "reports":
		return "since_send_time"
	default:
		return "since_date_created"
	}
}

func mailchimpPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return mailchimpDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mailchimp config page_size must be an integer: %w", err)
	}
	if value < 1 || value > mailchimpMaxPageSize {
		return 0, fmt.Errorf("mailchimp config page_size must be between 1 and %d", mailchimpMaxPageSize)
	}
	return value, nil
}

func mailchimpMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mailchimp config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("mailchimp config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// mailchimpTotalItems reads the total_items count from a response body, returning
// 0 when absent or unparsable.
func mailchimpTotalItems(body []byte) int {
	raw, err := connsdk.StringAt(body, "total_items")
	if err != nil || raw == "" {
		return 0
	}
	total, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return total
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
