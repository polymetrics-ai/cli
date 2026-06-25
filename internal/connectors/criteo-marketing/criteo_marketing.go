// Package criteomarketing implements the native pm Criteo Marketing connector.
// It follows the declarative-HTTP per-system template established by the stripe
// connector: a thin package composing the connsdk toolkit (Requester + OAuth2
// client-credentials auth + RecordsAt extraction) with Criteo-specific stream
// definitions and endpoints.
//
// Criteo Marketing Solutions is a read-only advertising data source for this
// connector. It authenticates with the OAuth2 client-credentials grant
// (client_id + client_secret exchanged for a short-lived bearer token at
// https://api.criteo.com/oauth2/token) and serves JSONAPI-shaped payloads from
// https://api.criteo.com. List endpoints page by offset/limit and return records
// under "data" (resource streams) or "Rows" (statistics report).
//
// The package self-registers with the connectors registry via RegisterFactory in
// init() under the key "criteo-marketing"; the registryset package blank-imports
// this package in the production binary to run that side effect.
package criteomarketing

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
	registryName = "criteo-marketing"

	criteoDefaultBaseURL = "https://api.criteo.com"
	// criteoTokenPath is appended to the base URL to form the OAuth2 token
	// endpoint (https://api.criteo.com/oauth2/token).
	criteoTokenPath       = "/oauth2/token"
	criteoDefaultPageSize = 100
	criteoMaxPageSize     = 1000
	criteoUserAgent       = "polymetrics-go-cli"
	// criteoFixtureDay is the deterministic report day used by fixture records.
	criteoFixtureDay = "2026-01-01"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Criteo Marketing connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Criteo Marketing connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth2 token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Criteo Marketing",
		IntegrationType: "api",
		Description:     "Reads Criteo Marketing Solutions ad sets, advertisers, campaigns, audiences, and ad spend statistics through the Criteo REST API using OAuth2 client-credentials auth.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Criteo. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := criteoBaseURL(cfg); err != nil {
		return err
	}
	if err := requireCredentials(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the advertisers list confirms the token exchange and
	// connectivity without mutating anything.
	q := url.Values{"limit": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, criteoStreamEndpoints["advertisers"].resource, q, nil, nil); err != nil {
		return fmt.Errorf("check criteo-marketing: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: criteoStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Criteo Marketing is a
// read-only source for this connector, so writes are unsupported.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "ad_sets"
	}
	endpoint, ok := criteoStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("criteo-marketing stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	if _, err := criteoBaseURL(req.Config); err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := criteoPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := criteoMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, req.Config, emit)
}

// harvest drives Criteo's offset/limit pagination. List endpoints return
// {data:[...]} (resource streams) or {Rows:[...]} (statistics) with no explicit
// "has more" flag, so the loop advances by offset and stops on a short page. The
// statistics report additionally requires startDate/endDate/currency filters.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	if reportFilters := statisticsFilters(endpoint, cfg); reportFilters != nil {
		for k, vs := range reportFilters {
			for _, v := range vs {
				base.Add(k, v)
			}
		}
	}

	method := endpoint.method
	if method == "" {
		method = http.MethodGet
	}

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("offset", strconv.Itoa(page*pageSize))
		resp, err := r.Do(ctx, method, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read criteo-marketing %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode criteo-marketing %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// No next-page token in the payload; a short page means we are done.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// statisticsFilters returns the report filters (start/end date, currency) that
// the statistics endpoint requires, or nil for resource streams.
func statisticsFilters(endpoint streamEndpoint, cfg connectors.RuntimeConfig) url.Values {
	if !strings.Contains(endpoint.resource, "statistics") {
		return nil
	}
	out := url.Values{}
	if start := strings.TrimSpace(cfg.Config["start_date"]); start != "" {
		out.Set("startDate", start)
	}
	if end := strings.TrimSpace(cfg.Config["end_date"]); end != "" {
		out.Set("endDate", end)
	}
	if currency := strings.TrimSpace(cfg.Config["currency"]); currency != "" {
		out.Set("currency", currency)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise criteo-marketing credential-free (mirrors the
// stripe fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	endpoint := criteoStreamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var item map[string]any
		if stream == "statistics" {
			item = map[string]any{
				"AdvertiserId": "adv_fixture_1",
				"CampaignId":   fmt.Sprintf("cmp_fixture_%d", i),
				"Day":          criteoFixtureDay,
				"Clicks":       int64(10 * i),
				"Displays":     int64(1000 * i),
				"Spend":        float64(i) * 12.5,
				"Currency":     fixtureCurrency(req.Config),
			}
		} else {
			item = map[string]any{
				"id":   fmt.Sprintf("%s_fixture_%d", stream, i),
				"type": strings.TrimSuffix(stream, "s"),
				"attributes": map[string]any{
					"name":          fmt.Sprintf("Fixture %d", i),
					"advertiserId":  "adv_fixture_1",
					"campaignId":    "cmp_fixture_1",
					"datasetId":     "ds_fixture_1",
					"status":        "Active",
					"mediaType":     "Display",
					"objective":     "Conversion",
					"country":       "US",
					"currency":      fixtureCurrency(req.Config),
					"timezone":      "UTC",
					"goal":          "Sales",
					"description":   fmt.Sprintf("Fixture audience %d", i),
					"nbActiveUsers": int64(500 * i),
				},
			}
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

// requester builds a connsdk.Requester wired with OAuth2 client-credentials auth
// and the resolved base URL. The client_id/client_secret only ever flow into the
// connsdk OAuth2ClientCredentials authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := criteoBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireCredentials(cfg); err != nil {
		return nil, err
	}
	auth := &connsdk.OAuth2ClientCredentials{
		TokenURL:     base + criteoTokenPath,
		ClientID:     criteoClientID(cfg),
		ClientSecret: criteoClientSecret(cfg),
		Client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: criteoUserAgent,
	}, nil
}

func requireCredentials(cfg connectors.RuntimeConfig) error {
	if strings.TrimSpace(criteoClientID(cfg)) == "" {
		return errors.New("criteo-marketing connector requires secret client_id")
	}
	if strings.TrimSpace(criteoClientSecret(cfg)) == "" {
		return errors.New("criteo-marketing connector requires secret client_secret")
	}
	return nil
}

func criteoClientID(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["client_id"]
}

func criteoClientSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["client_secret"]
}

// criteoBaseURL resolves and validates the base URL. The default is
// api.criteo.com; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func criteoBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return criteoDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("criteo-marketing config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("criteo-marketing config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("criteo-marketing config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func criteoPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return criteoDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("criteo-marketing config page_size must be an integer: %w", err)
	}
	if value < 1 || value > criteoMaxPageSize {
		return 0, fmt.Errorf("criteo-marketing config page_size must be between 1 and %d", criteoMaxPageSize)
	}
	return value, nil
}

func criteoMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("criteo-marketing config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("criteo-marketing config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureCurrency(cfg connectors.RuntimeConfig) string {
	if cfg.Config != nil {
		if c := strings.TrimSpace(cfg.Config["currency"]); c != "" {
			return c
		}
	}
	return "USD"
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
