// Package instagram implements the native pm Instagram connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package that composes the connsdk toolkit (Requester +
// Bearer auth + RecordsAt extraction + cursor state) with Instagram-specific
// stream definitions, Facebook Graph API endpoints, and record mappers.
//
// The Instagram source reads from the Facebook Graph API
// (https://graph.facebook.com/<version>) using a long-lived access token as a
// Bearer credential. Edges return {data:[...], paging:{next:"<absolute url>",
// cursors:{after:"..."}}}; pagination follows the absolute paging.next URL,
// which connsdk.Requester treats as an absolute request.
//
// Like stripe and github, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package instagram

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
	instagramDefaultBaseURL  = "https://graph.facebook.com/v23.0"
	instagramDefaultPageSize = 100
	instagramMaxPageSize     = 100
	instagramMaxPagesDefault = 0 // unlimited
	instagramUserAgent       = "polymetrics-go-cli"
	// instagramFixtureTimestamp is the deterministic timestamp used by
	// fixture-mode records.
	instagramFixtureTimestamp = "2026-01-01T00:00:00+0000"
	// defaultUserInsightMetrics/Period are applied to the user_insights edge,
	// which requires metric+period query params.
	defaultUserInsightMetrics = "reach,follower_count"
	defaultUserInsightPeriod  = "day"
)

func init() {
	connectors.RegisterFactory("instagram", New)
}

// New returns the Instagram connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Instagram connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk
	// Requester. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "instagram" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "instagram",
		DisplayName:     "Instagram",
		IntegrationType: "api",
		Description:     "Reads Instagram Business/Creator account profile, media, stories, and account insights through the Facebook Graph API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to the Graph
// API. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := instagramBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(instagramSecret(cfg)) == "" {
		return errors.New("instagram connector requires secret access_token")
	}
	userID, err := instagramUserID(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the account profile confirms auth and connectivity
	// without mutating anything.
	query := url.Values{"fields": []string{"id,username"}}
	if err := r.DoJSON(ctx, http.MethodGet, userID, query, nil, nil); err != nil {
		return fmt.Errorf("check instagram: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: instagramStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Instagram is a read-only
// source for reverse ETL purposes here, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: an Instagram stream starts
// with an empty incremental cursor (full sync) which the start_date config can
// raise at read time.
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
		stream = "users"
	}
	endpoint, ok := instagramStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("instagram stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	userID, err := instagramUserID(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	path := userID
	if endpoint.edge != "" {
		path = userID + "/" + endpoint.edge
	}
	query := url.Values{}
	if endpoint.fields != "" {
		query.Set("fields", endpoint.fields)
	}
	if stream == "user_insights" {
		applyInsightParams(query, req.Config)
	}

	if endpoint.single {
		return c.readSingle(ctx, r, path, query, endpoint, emit)
	}

	pageSize, err := instagramPageSize(req.Config)
	if err != nil {
		return err
	}
	if pageSize > 0 {
		query.Set("limit", strconv.Itoa(pageSize))
	}
	maxPages, err := instagramMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, path, query, endpoint, maxPages, emit)
}

// readSingle reads an endpoint that returns a single Graph API node (the users
// profile) and emits it as one record.
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, path string, query url.Values, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return fmt.Errorf("read instagram %s: %w", path, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode instagram %s: %w", path, err)
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

// harvest drives the Graph API's cursor pagination. Edges return
// {data:[...], paging:{next:"<absolute url>", cursors:{after:"..."}}}; the next
// page is the absolute paging.next URL, which connsdk.Requester fetches as-is.
// There is no body-token paginator in connsdk for the absolute-next shape, so
// the loop lives here, built on connsdk.Requester + connsdk.RecordsAt +
// connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, query url.Values, endpoint streamEndpoint, maxPages int, emit func(connectors.Record) error) error {
	nextURL := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		reqPath := path
		reqQuery := query
		if nextURL != "" {
			// paging.next is an absolute URL with all params embedded; do not
			// merge our base query (it would duplicate fields/limit).
			reqPath = nextURL
			reqQuery = nil
		}
		resp, err := r.Do(ctx, http.MethodGet, reqPath, reqQuery, nil)
		if err != nil {
			return fmt.Errorf("read instagram %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode instagram %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "paging.next")
		if err != nil {
			return fmt.Errorf("decode instagram %s paging: %w", path, err)
		}
		if strings.TrimSpace(next) == "" || len(records) == 0 {
			return nil
		}
		nextURL = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise instagram credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	endpoint := instagramStreamEndpoints[stream]
	count := 2
	if endpoint.single {
		count = 1
	}
	for i := 1; i <= count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                 fmt.Sprintf("%s_fixture_%d", stream, i),
			"username":           "fixture_account",
			"name":               "Fixture Account",
			"biography":          "deterministic fixture",
			"caption":            fmt.Sprintf("Fixture post %d", i),
			"media_type":         "IMAGE",
			"media_product_type": "FEED",
			"permalink":          fmt.Sprintf("https://instagram.com/p/fixture_%d", i),
			"timestamp":          instagramFixtureTimestamp,
			"like_count":         int64(10 * i),
			"comments_count":     int64(i),
			"followers_count":    int64(1200),
			"follows_count":      int64(300),
			"media_count":        int64(42),
			"name_insight":       "reach",
			"period":             "day",
			"values": []any{
				map[string]any{"value": int64(100 * i), "end_time": instagramFixtureTimestamp},
			},
		}
		// user_insights mapper keys off "name"; supply it for that stream.
		if stream == "user_insights" {
			item["name"] = "reach"
		}
		record := endpoint.mapRecord(item)
		if cursor := connsdk.Cursor(req.State); cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := instagramBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := instagramSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("instagram connector requires secret access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: instagramUserAgent,
	}, nil
}

// applyInsightParams sets the metric/period params required by the insights
// edge, honoring config overrides.
func applyInsightParams(query url.Values, cfg connectors.RuntimeConfig) {
	metrics := strings.TrimSpace(cfg.Config["insight_metrics"])
	if metrics == "" {
		metrics = defaultUserInsightMetrics
	}
	period := strings.TrimSpace(cfg.Config["insight_period"])
	if period == "" {
		period = defaultUserInsightPeriod
	}
	query.Set("metric", metrics)
	query.Set("period", period)
}

func instagramSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["access_token"]
}

// instagramUserID resolves the Instagram Business Account node id. It is
// required for live reads (it is the path root for every edge).
func instagramUserID(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return "", errors.New("instagram connector requires config ig_user_id")
	}
	id := strings.TrimSpace(cfg.Config["ig_user_id"])
	if id == "" {
		id = strings.TrimSpace(cfg.Config["instagram_business_account_id"])
	}
	if id == "" {
		return "", errors.New("instagram connector requires config ig_user_id")
	}
	// Guard against path traversal / injection into the request path.
	if strings.ContainsAny(id, "/?#") {
		return "", fmt.Errorf("instagram config ig_user_id contains invalid characters")
	}
	return id, nil
}

// instagramBaseURL resolves and validates the base URL. The default is
// graph.facebook.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func instagramBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return instagramDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("instagram config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("instagram config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("instagram config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func instagramPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return instagramDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("instagram config page_size must be an integer: %w", err)
	}
	if value < 1 || value > instagramMaxPageSize {
		return 0, fmt.Errorf("instagram config page_size must be between 1 and %d", instagramMaxPageSize)
	}
	return value, nil
}

func instagramMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return instagramMaxPagesDefault, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("instagram config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("instagram config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
