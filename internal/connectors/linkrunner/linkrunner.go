// Package linkrunner implements the native pm Linkrunner connector. Linkrunner
// is a Mobile Measurement Partner (MMP); this connector reads campaign and
// attributed-user analytics from its Data API.
//
// It is a declarative-HTTP per-system connector built on the connsdk toolkit
// (Requester + APIKeyHeader auth + RecordsAt extraction), copying the stripe
// reference shape. It self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package linkrunner

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
	linkrunnerDefaultBaseURL  = "https://api.linkrunner.io/api/v1"
	linkrunnerDefaultPageSize = 100
	linkrunnerMaxPageSize     = 100
	linkrunnerAPIKeyHeader    = "linkrunner-key"
	linkrunnerSecretField     = "linkrunner-key"
	linkrunnerUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("linkrunner", New)
}

// New returns the Linkrunner connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Linkrunner connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "linkrunner" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "linkrunner",
		DisplayName:     "Linkrunner",
		IntegrationType: "api",
		Description:     "Reads Linkrunner mobile attribution campaigns and attributed users from the Linkrunner Data API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Linkrunner.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := linkrunnerBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(linkrunnerSecret(cfg)) == "" {
		return fmt.Errorf("linkrunner connector requires secret %s", linkrunnerSecretField)
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the campaigns list confirms auth and connectivity.
	query := url.Values{"page": []string{"1"}, "limit": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "campaigns", query, nil, nil); err != nil {
		return fmt.Errorf("check linkrunner: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: linkrunnerStreams()}, nil
}

// Write is unsupported: the Linkrunner Data API is read-only for reverse ETL
// purposes, so this connector advertises Write=false and rejects writes.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "campaigns"
	}
	endpoint, ok := linkrunnerStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("linkrunner stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := linkrunnerPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := linkrunnerMaxPages(req.Config)
	if err != nil {
		return err
	}
	base, err := streamQuery(stream, req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, base, pageSize, maxPages, emit)
}

// harvest drives Linkrunner's page/limit pagination. Responses are
// {data:{<records>:[...]}}; pages start at 1 and increment until a short or
// empty page is returned. connsdk has a PageIncrement-style paginator, but the
// nested record path plus the empty-page stop condition are handled inline on
// top of connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))
		query.Set("limit", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read linkrunner %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode linkrunner %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Stop on an empty or short page (no further records to fetch).
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise linkrunner credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"display_id":          fmt.Sprintf("camp_fixture_%d", i),
			"name":                fmt.Sprintf("Fixture Campaign %d", i),
			"created_at":          "2026-01-01T00:00:00Z",
			"update_at":           "2026-01-02T00:00:00Z",
			"google":              true,
			"meta":                false,
			"meta_campaign_id":    "",
			"active":              true,
			"default_link":        false,
			"attributed_users":    int64(10 * i),
			"link":                fmt.Sprintf("https://lr.example/c/%d", i),
			"shareable_link":      fmt.Sprintf("https://lr.example/s/%d", i),
			"website":             "https://example.com",
			"domain":              "example.com",
			"campaign_display_id": fmt.Sprintf("camp_fixture_%d", i),
			"campaign_name":       fmt.Sprintf("Fixture Campaign %d", i),
			"attributed_at":       fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"installed_at":        fmt.Sprintf("2026-01-0%dT00:05:00Z", i),
			"store_click_at":      fmt.Sprintf("2026-01-0%dT00:01:00Z", i),
			"ad_set_id":           fmt.Sprintf("adset_%d", i),
			"user_data":           map[string]any{"id": fmt.Sprintf("user_%d", i), "email": fmt.Sprintf("fixture+%d@example.com", i)},
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

// streamQuery returns the stream-specific base query parameters (filters scoped
// to each stream), excluding pagination which harvest adds per page.
func streamQuery(stream string, cfg connectors.RuntimeConfig) (url.Values, error) {
	q := url.Values{}
	switch stream {
	case "campaigns":
		if v := strings.TrimSpace(cfg.Config["filter"]); v != "" {
			q.Set("filter", v)
		}
		if v := strings.TrimSpace(cfg.Config["channel"]); v != "" {
			q.Set("channel", v)
		}
	case "attributed_users":
		displayID := strings.TrimSpace(cfg.Config["display_id"])
		if displayID == "" {
			return nil, errors.New("linkrunner attributed_users stream requires config display_id")
		}
		q.Set("display_id", displayID)
		if v := strings.TrimSpace(cfg.Config["start_timestamp"]); v != "" {
			q.Set("start_timestamp", v)
		}
		if v := strings.TrimSpace(cfg.Config["end_timestamp"]); v != "" {
			q.Set("end_timestamp", v)
		}
		if v := strings.TrimSpace(cfg.Config["timezone"]); v != "" {
			q.Set("timezone", v)
		}
	}
	return q, nil
}

// requester builds a connsdk.Requester wired with the linkrunner-key API-key
// header, the resolved base URL, and the standard user agent. The secret only
// ever flows into connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := linkrunnerBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := linkrunnerSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, fmt.Errorf("linkrunner connector requires secret %s", linkrunnerSecretField)
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(linkrunnerAPIKeyHeader, secret, ""),
		UserAgent: linkrunnerUserAgent,
	}, nil
}

func linkrunnerSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[linkrunnerSecretField]
}

// linkrunnerBaseURL resolves and validates the base URL. The default is
// api.linkrunner.io; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func linkrunnerBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return linkrunnerDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("linkrunner config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("linkrunner config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("linkrunner config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func linkrunnerPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return linkrunnerDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("linkrunner config page_size must be an integer: %w", err)
	}
	if value < 1 || value > linkrunnerMaxPageSize {
		return 0, fmt.Errorf("linkrunner config page_size must be between 1 and %d", linkrunnerMaxPageSize)
	}
	return value, nil
}

func linkrunnerMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("linkrunner config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("linkrunner config max_pages must be 0 for unlimited or a positive integer")
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
