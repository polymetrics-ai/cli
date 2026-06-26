// Package fullstory implements the native pm FullStory connector. It is a
// declarative-HTTP per-system connector following the stripe template: a thin
// package that composes the connsdk toolkit (Requester + Basic api-key auth +
// RecordsAt extraction + a small next_page_token cursor loop) with
// FullStory-specific stream definitions and endpoints.
//
// FullStory's analytics API is read-only for our purposes (segments, users,
// events), so this connector exposes Check/Catalog/Read and no Write actions.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package fullstory

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
	fullstoryDefaultBaseURL  = "https://api.fullstory.com"
	fullstoryDefaultPageSize = 200
	fullstoryMaxPageSize     = 1000
	fullstoryUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("fullstory", New)
}

// New returns the FullStory connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm FullStory connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "fullstory" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "fullstory",
		DisplayName:     "Fullstory",
		IntegrationType: "api",
		Description:     "Reads FullStory segments, users, and events through the FullStory REST API (read-only analytics export).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to FullStory. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := fullstoryBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(fullstoryAPIKey(cfg)) == "" {
		return errors.New("fullstory connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the segments list confirms auth and connectivity
	// without mutating anything.
	endpoint := fullstoryStreamEndpoints["segments"]
	if err := r.DoJSON(ctx, http.MethodGet, endpoint.resource, url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check fullstory: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: fullstoryStreams()}, nil
}

// Write satisfies the connectors.Connector interface. FullStory is a read-only
// analytics source, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: FullStory only supports
// full_refresh, so a stream starts with an empty cursor.
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
		stream = "segments"
	}
	endpoint, ok := fullstoryStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("fullstory stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := fullstoryPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := fullstoryMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives FullStory's next_page_token cursor pagination. List endpoints
// return {results:[...], next_page_token:"..."}; the next page is requested with
// pageToken=<token>. The loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))

	pageToken := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read fullstory %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode fullstory %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_page_token")
		if err != nil {
			return fmt.Errorf("decode fullstory %s next_page_token: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		pageToken = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise fullstory credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                 fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":               fmt.Sprintf("Fixture %s %d", stream, i),
			"description":        "deterministic fixture record",
			"creator":            "fixture@example.com",
			"created":            "2026-01-01T00:00:00Z",
			"updated":            "2026-01-01T00:00:00Z",
			"is_public":          false,
			"type":               "fixture",
			"uid":                fmt.Sprintf("uid_%d", i),
			"display_name":       fmt.Sprintf("Fixture User %d", i),
			"email":              fmt.Sprintf("fixture+%d@example.com", i),
			"is_being_processed": false,
			"user_id":            fmt.Sprintf("user_%d", i),
			"session_id":         fmt.Sprintf("sess_%d", i),
			"event_time":         "2026-01-01T00:00:00Z",
			"device_id":          fmt.Sprintf("dev_%d", i),
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

// requester builds a connsdk.Requester wired with Basic api-key auth, the
// resolved base URL, and the optional FullStory uid header. The api_key only ever
// flows into the Authorization header via connsdk; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := fullstoryBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	apiKey := fullstoryAPIKey(cfg)
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("fullstory connector requires secret api_key")
	}
	headers := map[string]string{}
	if uid := strings.TrimSpace(fullstoryUID(cfg)); uid != "" {
		headers["FS-Uid"] = uid
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", apiKey, "Basic "),
		UserAgent: fullstoryUserAgent,
		// FullStory issues legacy v1 (no prefix) and v2 (Basic) keys; the
		// Airbyte connector sends the raw key with a "Basic " prefix, which we
		// mirror via APIKeyHeader above. DefaultHeaders carries the uid.
		DefaultHeaders: headers,
	}, nil
}

func fullstoryAPIKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

func fullstoryUID(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["uid"]
}

// fullstoryBaseURL resolves and validates the base URL. The default is
// api.fullstory.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func fullstoryBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return fullstoryDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("fullstory config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("fullstory config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("fullstory config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fullstoryPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return fullstoryDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("fullstory config page_size must be an integer: %w", err)
	}
	if value < 1 || value > fullstoryMaxPageSize {
		return 0, fmt.Errorf("fullstory config page_size must be between 1 and %d", fullstoryMaxPageSize)
	}
	return value, nil
}

func fullstoryMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("fullstory config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("fullstory config max_pages must be 0 for unlimited or a positive integer")
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
