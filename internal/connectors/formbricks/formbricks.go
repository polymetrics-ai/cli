// Package formbricks implements the native pm Formbricks connector. It follows
// the declarative-HTTP template established by the stripe connector: a thin
// package that composes the connsdk toolkit (Requester + X-API-Key auth +
// RecordsAt extraction over data[]) with Formbricks-specific stream definitions
// and endpoints.
//
// Formbricks is a read-only analytics/survey source for this connector. It
// self-registers with the connectors registry via RegisterFactory in init(); the
// registryset package blank-imports this package in the production binary to run
// that side effect.
package formbricks

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
	formbricksDefaultBaseURL  = "https://app.formbricks.com/api/v1"
	formbricksDefaultPageSize = 50
	formbricksMaxPageSize     = 100
	formbricksUserAgent       = "polymetrics-go-cli"
	formbricksAPIKeyHeader    = "X-API-Key"
)

func init() {
	connectors.RegisterFactory("formbricks", New)
}

// New returns the Formbricks connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Formbricks connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "formbricks" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "formbricks",
		DisplayName:     "Formbricks",
		IntegrationType: "api",
		Description:     "Reads Formbricks surveys, responses, action classes, attribute classes, and webhooks through the Formbricks management API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Formbricks.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := formbricksBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(formbricksSecret(cfg)) == "" {
		return errors.New("formbricks connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the surveys list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "management/surveys", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check formbricks: %w", err)
	}
	return nil
}

// Write is unsupported: Formbricks is a read-only source for this connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: formbricksStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Formbricks stream starts
// with an empty incremental cursor (full sync).
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
		stream = "surveys"
	}
	endpoint, ok := formbricksStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("formbricks stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := formbricksPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := formbricksMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Formbricks pagination. List responses wrap their records under
// {data:[...]}. The responses endpoint supports offset pagination via limit/skip;
// other endpoints return their full collection in one page. The loop stops when a
// short page (fewer than pageSize records) is returned, maxPages is reached, or
// (for non-paginated endpoints) after the first page.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	skip := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if endpoint.paginated {
			query.Set("limit", strconv.Itoa(pageSize))
			query.Set("skip", strconv.Itoa(skip))
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read formbricks %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode formbricks %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Non-paginated endpoints return everything in one page; a short page
		// (or any page) for paginated endpoints with fewer than pageSize records
		// signals the end of the collection.
		if !endpoint.paginated || len(records) < pageSize {
			return nil
		}
		skip += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise formbricks credential-free (mirrors stripe).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":          fmt.Sprintf("Fixture %d", i),
			"type":          "link",
			"status":        "inProgress",
			"description":   fmt.Sprintf("fixture %s %d", stream, i),
			"environmentId": "env_fixture_1",
			"surveyId":      "srv_fixture_1",
			"contactId":     "ctc_fixture_1",
			"finished":      i%2 == 0,
			"archived":      false,
			"url":           "https://example.com/webhook",
			"source":        "user",
			"surveyIds":     []any{"srv_fixture_1"},
			"triggers":      []any{"responseCreated"},
			"data":          map[string]any{"q1": "answer"},
			"meta":          map[string]any{"source": "fixture"},
			"createdAt":     "2026-01-01T00:00:00.000Z",
			"updatedAt":     fmt.Sprintf("2026-01-0%dT00:00:00.000Z", i),
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "formbricks"
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

// requester builds a connsdk.Requester wired with X-API-Key auth and the resolved
// base URL. The secret only ever flows into connsdk.APIKeyHeader; it is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := formbricksBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := formbricksSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("formbricks connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(formbricksAPIKeyHeader, secret, ""),
		UserAgent: formbricksUserAgent,
	}, nil
}

func formbricksSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// formbricksBaseURL resolves and validates the base URL. The default is
// app.formbricks.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func formbricksBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return formbricksDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("formbricks config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("formbricks config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("formbricks config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func formbricksPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return formbricksDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("formbricks config page_size must be an integer: %w", err)
	}
	if value < 1 || value > formbricksMaxPageSize {
		return 0, fmt.Errorf("formbricks config page_size must be between 1 and %d", formbricksMaxPageSize)
	}
	return value, nil
}

func formbricksMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("formbricks config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("formbricks config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
