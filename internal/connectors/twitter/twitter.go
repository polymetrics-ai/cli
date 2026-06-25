// Package twitter implements the native pm Twitter (X) connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit: a
// connsdk.Requester with App-only Bearer auth, connsdk.RecordsAt extraction, and
// a small in-package cursor loop for Twitter v2's meta.next_token pagination.
//
// It reads the Twitter API v2 recent-search endpoint, exposing the Tweets and
// Authors streams that mirror the upstream Airbyte source-twitter connector.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package twitter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	twitterDefaultBaseURL  = "https://api.twitter.com/2"
	twitterDefaultPageSize = 100
	twitterMaxPageSize     = 100
	twitterMinPageSize     = 10
	twitterUserAgent       = "polymetrics-go-cli"
	// twitterTweetFields and twitterUserFields request the expanded field sets so
	// the mapped records carry useful columns. author_id expansion populates
	// includes.users[] which feeds the authors stream.
	twitterTweetFieldList = "id,text,author_id,created_at,conversation_id,lang,source,in_reply_to_user_id,possibly_sensitive,public_metrics"
	twitterUserFieldList  = "id,name,username,created_at,description,location,verified,protected,url,public_metrics"
)

func init() {
	connectors.RegisterFactory("twitter", New)
}

// New returns the Twitter connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Twitter connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "twitter" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "twitter",
		DisplayName:     "Twitter",
		IntegrationType: "api",
		Description:     "Reads tweets and their authors matching a search query from the Twitter (X) API v2 recent search endpoint using an App-only Bearer token.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Twitter. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := twitterBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(twitterSecret(cfg)) == "" {
		return errors.New("twitter connector requires secret api_key (App-only Bearer token)")
	}
	query := twitterQuery(cfg)
	if query == "" {
		return errors.New("twitter connector requires config query")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded recent-search confirms auth, query validity, and connectivity
	// without writing anything.
	q := url.Values{}
	q.Set("query", query)
	q.Set("max_results", strconv.Itoa(twitterMinPageSize))
	if err := r.DoJSON(ctx, http.MethodGet, "tweets/search/recent", q, nil, nil); err != nil {
		return fmt.Errorf("check twitter: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: twitterStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "tweets"
	}
	endpoint, ok := twitterStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("twitter stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	query := twitterQuery(req.Config)
	if query == "" {
		return errors.New("twitter connector requires config query")
	}
	pageSize, err := twitterPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := twitterMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, query, req.Config, pageSize, maxPages, emit)
}

// harvest drives Twitter v2's cursor pagination. Recent-search responses are
// {data:[...], includes:{users:[...]}, meta:{next_token,...}}; the next page is
// requested with next_token=<meta.next_token> until it is absent. The exact
// {meta.next_token} shape is not covered by a connsdk Paginator, so the loop
// lives here, built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
//
// Both streams page over the same endpoint; the only difference is which JSON
// path (endpoint.recordsPath) the records are read from per page.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, query string, cfg connectors.RuntimeConfig, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("query", query)
	base.Set("max_results", strconv.Itoa(pageSize))
	base.Set("tweet.fields", twitterTweetFieldList)
	base.Set("expansions", "author_id")
	base.Set("user.fields", twitterUserFieldList)
	if start := strings.TrimSpace(cfg.Config["start_date"]); start != "" {
		base.Set("start_time", start)
	}
	if end := strings.TrimSpace(cfg.Config["end_date"]); end != "" {
		base.Set("end_time", end)
	}

	nextToken := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		q := cloneValues(base)
		if nextToken != "" {
			q.Set("next_token", nextToken)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, q, nil)
		if err != nil {
			return fmt.Errorf("read twitter %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode twitter %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		token, err := connsdk.StringAt(resp.Body, "meta.next_token")
		if err != nil {
			return fmt.Errorf("decode twitter %s next_token: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(token) == "" {
			return nil
		}
		nextToken = token
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise twitter credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	endpoint := twitterStreamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var item map[string]any
		switch stream {
		case "authors":
			item = map[string]any{
				"id":             fmt.Sprintf("author_fixture_%d", i),
				"name":           fmt.Sprintf("Fixture Author %d", i),
				"username":       fmt.Sprintf("fixture_author_%d", i),
				"created_at":     "2026-01-01T00:00:00.000Z",
				"description":    "Deterministic fixture author.",
				"location":       "Nowhere",
				"verified":       false,
				"protected":      false,
				"url":            "https://example.com",
				"public_metrics": map[string]any{"followers_count": 100 * i},
			}
		default:
			item = map[string]any{
				"id":                  fmt.Sprintf("tweet_fixture_%d", i),
				"text":                fmt.Sprintf("Fixture tweet %d", i),
				"author_id":           fmt.Sprintf("author_fixture_%d", i),
				"created_at":          "2026-01-01T00:00:00.000Z",
				"conversation_id":     fmt.Sprintf("tweet_fixture_%d", i),
				"lang":                "en",
				"source":              "Twitter Web App",
				"in_reply_to_user_id": nil,
				"possibly_sensitive":  false,
				"public_metrics":      map[string]any{"like_count": 10 * i},
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

// requester builds a connsdk.Requester wired with App-only Bearer auth and the
// resolved base URL. The secret only ever flows into connsdk.Bearer; it is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := twitterBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := twitterSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("twitter connector requires secret api_key (App-only Bearer token)")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: twitterUserAgent,
	}, nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func twitterSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

func twitterQuery(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Config["query"])
}

// twitterBaseURL resolves and validates the base URL. The default is
// api.twitter.com; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func twitterBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return twitterDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("twitter config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("twitter config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("twitter config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func twitterPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return twitterDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("twitter config page_size must be an integer: %w", err)
	}
	if value < twitterMinPageSize || value > twitterMaxPageSize {
		return 0, fmt.Errorf("twitter config page_size must be between %d and %d", twitterMinPageSize, twitterMaxPageSize)
	}
	return value, nil
}

func twitterMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("twitter config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("twitter config max_pages must be 0 for unlimited or a positive integer")
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
