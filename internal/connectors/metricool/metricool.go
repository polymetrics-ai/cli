// Package metricool implements the native pm Metricool connector. It follows the
// declarative-HTTP template established by the stripe connector: a thin package
// that composes the connsdk toolkit (Requester + APIKeyHeader auth + RecordsAt
// extraction) with Metricool-specific stream definitions and endpoints.
//
// Metricool's analytics API is account-scoped (userId) and partitioned by brand
// (blogId), and is itself not paginated. The connector therefore fans out one
// request per configured blog_id, stamping blogId onto each emitted record so
// the multi-brand rows stay attributable.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package metricool

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	metricoolDefaultBaseURL = "https://app.metricool.com/api"
	metricoolAuthHeader     = "X-Mc-Auth"
	metricoolUserAgent      = "polymetrics-go-cli"
	// metricoolDefaultLookbackDays mirrors the upstream connector default of
	// 60 days back when no start_date is configured.
	metricoolDefaultLookbackDays = 60
	// metricoolFixtureDate is the deterministic publish date used by fixture
	// records (2026-01-01).
	metricoolFixtureDate = "2026-01-01T00:00:00"
)

func init() {
	connectors.RegisterFactory("metricool", New)
}

// New returns the Metricool connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Metricool connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "metricool" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "metricool",
		DisplayName:     "Metricool",
		IntegrationType: "api",
		Description:     "Reads Metricool brand profiles and per-brand Instagram, Facebook, LinkedIn, and TikTok post analytics through the Metricool REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Metricool. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := metricoolBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(metricoolSecret(cfg)) == "" {
		return errors.New("metricool connector requires secret user_token")
	}
	userID := strings.TrimSpace(cfg.Config["user_id"])
	if userID == "" {
		return errors.New("metricool connector requires config user_id")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// Listing brand profiles confirms auth and connectivity without scoping to a
	// specific blog or mutating anything.
	q := url.Values{"userId": []string{userID}}
	if err := r.DoJSON(ctx, http.MethodGet, "admin/simpleProfiles", q, nil, nil); err != nil {
		return fmt.Errorf("check metricool: %w", err)
	}
	return nil
}

// Write is unsupported: Metricool is a read-only analytics source. The method
// exists only to satisfy the connectors.Connector interface; Capabilities.Write
// is false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: metricoolStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "brands"
	}
	endpoint, ok := metricoolStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("metricool stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	userID := strings.TrimSpace(req.Config.Config["user_id"])
	if userID == "" {
		return errors.New("metricool connector requires config user_id")
	}
	from, to, err := metricoolDateRange(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	base.Set("userId", userID)
	switch endpoint.dates {
	case dateLegacy:
		base.Set("start", from.Format("20060102"))
		base.Set("end", to.Format("20060102"))
	case dateV2:
		base.Set("from", from.Format("2006-01-02T15:04:05"))
		base.Set("to", to.Format("2006-01-02T15:04:05"))
	}

	// Account-wide streams (brands) read once; per-brand streams fan out across
	// every configured blog_id.
	if !endpoint.perBlog {
		return c.readOne(ctx, r, endpoint, base, "", emit)
	}
	blogs := metricoolBlogIDs(req.Config)
	if len(blogs) == 0 {
		return errors.New("metricool connector requires config blog_ids for per-brand streams")
	}
	for _, blog := range blogs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := c.readOne(ctx, r, endpoint, base, blog, emit); err != nil {
			return err
		}
	}
	return nil
}

// readOne issues a single request for one endpoint+blog and emits its records.
// When blog is non-empty it is sent as blogId and stamped onto each record so
// rows from different brands remain distinguishable.
func (c Connector) readOne(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, blog string, emit func(connectors.Record) error) error {
	query := cloneValues(base)
	if blog != "" {
		query.Set("blogId", blog)
	}
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
	if err != nil {
		return fmt.Errorf("read metricool %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode metricool %s: %w", endpoint.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		record := endpoint.mapRecord(item)
		if blog != "" {
			record["blogId"] = blog
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise metricool credential-free.
func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           int64(i),
			"label":        fmt.Sprintf("Fixture Brand %d", i),
			"title":        fmt.Sprintf("Fixture Brand %d", i),
			"userId":       int64(9999),
			"url":          "https://example.com",
			"timezone":     "UTC",
			"postId":       fmt.Sprintf("post_fixture_%d", i),
			"videoId":      fmt.Sprintf("video_fixture_%d", i),
			"type":         "image",
			"publishDate":  metricoolFixtureDate,
			"text":         fmt.Sprintf("Fixture post %d", i),
			"interactions": int64(10 * i),
			"impressions":  int64(100 * i),
			"reach":        int64(90 * i),
			"likes":        int64(5 * i),
			"comments":     int64(i),
			"shares":       int64(i),
			"saved":        int64(i),
			"clicks":       int64(2 * i),
			"views":        int64(200 * i),
			"engagement":   int64(3 * i),
		}
		record := endpoint.mapRecord(item)
		if endpoint.perBlog {
			record["blogId"] = "fixture_blog_1"
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the X-Mc-Auth API-key header
// auth and the resolved base URL. The secret only ever flows into
// connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := metricoolBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := metricoolSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("metricool connector requires secret user_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(metricoolAuthHeader, secret, ""),
		UserAgent: metricoolUserAgent,
	}, nil
}

func metricoolSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["user_token"]
}

// metricoolBlogIDs parses the comma-separated blog_ids config into a clean slice.
func metricoolBlogIDs(cfg connectors.RuntimeConfig) []string {
	raw := strings.TrimSpace(cfg.Config["blog_ids"])
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// metricoolDateRange resolves the [from, to] window from config, defaulting to
// the upstream 60-days-back lookback when start_date is unset and to "now" when
// end_date is unset.
func metricoolDateRange(cfg connectors.RuntimeConfig) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	to := now
	if raw := strings.TrimSpace(cfg.Config["end_date"]); raw != "" {
		t, err := parseMetricoolTime(raw)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("metricool config end_date is invalid: %w", err)
		}
		to = t
	}
	from := to.AddDate(0, 0, -metricoolDefaultLookbackDays)
	if raw := strings.TrimSpace(cfg.Config["start_date"]); raw != "" {
		t, err := parseMetricoolTime(raw)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("metricool config start_date is invalid: %w", err)
		}
		from = t
	}
	if from.After(to) {
		from = to.AddDate(0, 0, -1)
	}
	return from, to, nil
}

func parseMetricoolTime(raw string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05Z", "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("expected RFC3339 or YYYY-MM-DD, got %q", raw)
}

// metricoolBaseURL resolves and validates the base URL. The default is
// app.metricool.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func metricoolBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return metricoolDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("metricool config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("metricool config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("metricool config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
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
