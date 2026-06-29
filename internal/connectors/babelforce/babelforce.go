// Package babelforce implements the native pm Babelforce connector. It is a
// declarative-HTTP per-system connector built on the stripe template: a thin
// package that composes the connsdk toolkit (Requester + dual-header auth +
// RecordsAt extraction at "items" + page-cursor state) with Babelforce-specific
// stream definitions, endpoints, and record mappers.
//
// Babelforce's reporting/list endpoints return the envelope
// {"items":[...],"pagination":{"current":N,"max":M}} and authenticate with two
// headers, X-Auth-Access-ID and X-Auth-Access-Token. It is read-only.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package babelforce

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
	babelforceDefaultRegion   = "services"
	babelforceDefaultPageSize = 102
	babelforceMaxPageSize     = 102
	babelforceUserAgent       = "polymetrics-go-cli"
	babelforceAccessIDHeader  = "X-Auth-Access-ID"
	babelforceTokenHeader     = "X-Auth-Access-Token"
	// babelforceFixtureCreated is the deterministic dateCreated (unix seconds) used
	// by fixture-mode records (2026-01-01T00:00:00Z).
	babelforceFixtureCreated int64 = 1767225600
)

// babelforceRegions is the allow-list of region subdomains the default base URL
// may use, matching the upstream upstream spec enum.
var babelforceRegions = map[string]bool{
	"services":     true,
	"us-east":      true,
	"ap-southeast": true,
}

func init() {
	connectors.RegisterFactory("babelforce", New)
}

// New returns the Babelforce connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Babelforce connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "babelforce" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "babelforce",
		DisplayName:     "Babelforce",
		IntegrationType: "api",
		Description:     "Reads Babelforce call reporting, recordings, numbers, and users through the Babelforce v2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Babelforce.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := babelforceBaseURL(cfg); err != nil {
		return err
	}
	id, token := babelforceSecrets(cfg)
	if strings.TrimSpace(id) == "" || strings.TrimSpace(token) == "" {
		return errors.New("babelforce connector requires secrets access_key_id and access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the calls report confirms auth and connectivity without
	// mutating anything.
	q := url.Values{"max": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "calls/reporting/simple", q, nil, nil); err != nil {
		return fmt.Errorf("check babelforce: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: babelforceStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Babelforce is read-only
// (a call/telephony reporting source), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a Babelforce stream starts
// with an empty incremental cursor (full sync), which the date_created_from
// config can raise at read time.
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
		stream = "calls"
	}
	endpoint, ok := babelforceStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("babelforce stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := babelforcePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := babelforceMaxPages(req.Config)
	if err != nil {
		return err
	}
	lower, upper, err := dateFilters(req)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, lower, upper, emit)
}

// harvest drives Babelforce's page-cursor pagination. Reporting/list responses
// return {items:[...], pagination:{current:N, max:M}}; the next page is requested
// with page=<pagination.current + 1>, and reading stops once the response no
// longer carries pagination.current. The loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, fromUnix, toUnix string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("max", strconv.Itoa(pageSize))
	if fromUnix != "" {
		base.Set("filters.dateCreated.from", fromUnix)
	}
	if toUnix != "" {
		base.Set("filters.dateCreated.to", toUnix)
	}

	page := ""
	for pageNum := 0; maxPages == 0 || pageNum < maxPages; pageNum++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if page != "" {
			query.Set("page", page)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read babelforce %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "items")
		if err != nil {
			return fmt.Errorf("decode babelforce %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Stop condition mirrors the upstream manifest: halt once the response no
		// longer carries pagination.current. Otherwise next page = current + 1.
		current, err := connsdk.StringAt(resp.Body, "pagination.current")
		if err != nil {
			return fmt.Errorf("decode babelforce %s pagination: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(current) == "" {
			return nil
		}
		curNum, err := strconv.Atoi(strings.TrimSpace(current))
		if err != nil {
			// Non-numeric current: cannot advance, stop to avoid a loop.
			return nil
		}
		page = strconv.Itoa(curNum + 1)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise babelforce credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		created := strconv.FormatInt(babelforceFixtureCreated+int64(i), 10)
		item := map[string]any{
			"id":              fmt.Sprintf("%s_fixture_%d", stream, i),
			"type":            "inbound",
			"state":           "finished",
			"source":          "fixture",
			"domain":          "example.babelforce.com",
			"from":            "+10000000000",
			"to":              "+19999999999",
			"anonymous":       false,
			"duration":        int64(30 * i),
			"finishReason":    "completed",
			"conversationId":  "conv_fixture_1",
			"sessionId":       "sess_fixture_1",
			"parentId":        nil,
			"dateCreated":     created,
			"dateEstablished": created,
			"dateFinished":    created,
			"lastUpdated":     created,
			"url":             "https://example.babelforce.com/recordings/fixture",
			"number":          "+19999999999",
			"name":            fmt.Sprintf("Fixture %d", i),
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

// requester builds a connsdk.Requester wired with dual-header auth and the
// resolved base URL. The secrets only ever flow into request headers; they are
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := babelforceBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	id, token := babelforceSecrets(cfg)
	if strings.TrimSpace(id) == "" || strings.TrimSpace(token) == "" {
		return nil, errors.New("babelforce connector requires secrets access_key_id and access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		UserAgent: babelforceUserAgent,
		DefaultHeaders: map[string]string{
			babelforceAccessIDHeader: strings.TrimSpace(id),
			babelforceTokenHeader:    strings.TrimSpace(token),
		},
	}, nil
}

// dateFilters returns the unix-seconds [from, to] bounds for the dateCreated
// filter, derived from the incremental cursor (if any, used as the lower bound)
// or else the date_created_from config, plus the optional date_created_to.
func dateFilters(req connectors.ReadRequest) (from, to string, err error) {
	from = strings.TrimSpace(req.Config.Config["date_created_from"])
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		from = cursor
	}
	if from != "" {
		if _, e := strconv.ParseInt(from, 10, 64); e != nil {
			return "", "", fmt.Errorf("babelforce date_created_from must be a unix timestamp: %w", e)
		}
	}
	to = strings.TrimSpace(req.Config.Config["date_created_to"])
	if to != "" {
		if _, e := strconv.ParseInt(to, 10, 64); e != nil {
			return "", "", fmt.Errorf("babelforce date_created_to must be a unix timestamp: %w", e)
		}
	}
	return from, to, nil
}

func babelforceSecrets(cfg connectors.RuntimeConfig) (id, token string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return cfg.Secrets["access_key_id"], cfg.Secrets["access_token"]
}

// babelforceBaseURL resolves and validates the base URL. The default is derived
// from the region config (region.babelforce.com/api/v2). Any base_url override
// must be an absolute https (or http for local test servers) URL with a host to
// bound SSRF risk.
func babelforceBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		region := strings.TrimSpace(cfg.Config["region"])
		if region == "" {
			region = babelforceDefaultRegion
		}
		if !babelforceRegions[region] {
			return "", fmt.Errorf("babelforce config region %q is not one of services, us-east, ap-southeast", region)
		}
		return fmt.Sprintf("https://%s.babelforce.com/api/v2", region), nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("babelforce config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("babelforce config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("babelforce config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func babelforcePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return babelforceDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("babelforce config page_size must be an integer: %w", err)
	}
	if value < 1 || value > babelforceMaxPageSize {
		return 0, fmt.Errorf("babelforce config page_size must be between 1 and %d", babelforceMaxPageSize)
	}
	return value, nil
}

func babelforceMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("babelforce config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("babelforce config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
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
