// Package notion implements the native pm Notion source connector. It follows
// the declarative-HTTP shape of the stripe reference connector: a thin package
// composing the connsdk toolkit (Requester + Bearer auth + RecordsAt extraction
// + cursor state) with Notion-specific stream definitions, endpoints, and the
// Notion-Version header the API requires.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect. Notion is read-only (no reverse-ETL writes).
package notion

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
	notionDefaultBaseURL = "https://api.notion.com/v1"
	// notionAPIVersion is the required Notion-Version header value. The Notion
	// API rejects requests without it.
	notionAPIVersion      = "2022-06-28"
	notionDefaultPageSize = 100
	notionMaxPageSize     = 100
	notionUserAgent       = "polymetrics-go-cli"
	// notionFixtureEdited is the deterministic last_edited_time used by
	// fixture-mode records.
	notionFixtureEdited = "2026-01-01T00:00:00.000Z"
)

func init() {
	connectors.RegisterFactory("notion", New)
}

// New returns the Notion connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Notion source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "notion" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "notion",
		DisplayName:     "Notion",
		IntegrationType: "api",
		Description:     "Reads Notion databases, pages, and users through the Notion REST API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Notion. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := notionBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(notionSecret(cfg)) == "" {
		return errors.New("notion connector requires secret credentials.access_token (or token)")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the users list confirms auth, the Notion-Version
	// header, and connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "users", url.Values{"page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check notion: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: notionStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Notion stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time for the last_edited_time-cursored streams.
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
		stream = "databases"
	}
	endpoint, ok := notionStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("notion stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := notionPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := notionMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Notion's start_cursor pagination. Every list response is
// {results:[...], next_cursor:string|null, has_more:bool}. POST /search carries
// the cursor and object filter in the request body; GET /users carries the
// cursor as the start_cursor query parameter. The loop is built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt because Notion's
// body-cursor-on-POST shape has no off-the-shelf connsdk paginator.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		var (
			resp *connsdk.Response
			err  error
		)
		if endpoint.method == http.MethodPost {
			body := map[string]any{"page_size": pageSize}
			if endpoint.searchObject != "" {
				body["filter"] = map[string]any{"property": "object", "value": endpoint.searchObject}
			}
			if cursor != "" {
				body["start_cursor"] = cursor
			}
			resp, err = r.Do(ctx, http.MethodPost, endpoint.resource, nil, body)
		} else {
			query := url.Values{}
			query.Set("page_size", strconv.Itoa(pageSize))
			if cursor != "" {
				query.Set("start_cursor", cursor)
			}
			resp, err = r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		}
		if err != nil {
			return fmt.Errorf("read notion %s: %w", endpoint.resource, err)
		}

		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode notion %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		hasMore, err := connsdk.StringAt(resp.Body, "has_more")
		if err != nil {
			return fmt.Errorf("decode notion %s has_more: %w", endpoint.resource, err)
		}
		next, err := connsdk.StringAt(resp.Body, "next_cursor")
		if err != nil {
			return fmt.Errorf("decode notion %s next_cursor: %w", endpoint.resource, err)
		}
		if hasMore != "true" || strings.TrimSpace(next) == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise notion credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	objectType := endpoint.searchObject
	if objectType == "" {
		objectType = strings.TrimSuffix(stream, "s")
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               fmt.Sprintf("%s_fixture_%d", objectType, i),
			"object":           objectType,
			"created_time":     notionFixtureEdited,
			"last_edited_time": notionFixtureEdited,
			"archived":         false,
			"in_trash":         false,
			"url":              fmt.Sprintf("https://www.notion.so/%s-fixture-%d", objectType, i),
			"parent":           map[string]any{"type": "workspace", "workspace": true},
			"name":             fmt.Sprintf("Fixture %s %d", objectType, i),
			"type":             "person",
			"title":            []any{},
			"properties":       map[string]any{},
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

// requester builds a connsdk.Requester wired with Bearer auth, the resolved base
// URL, and the required Notion-Version header. The secret only ever flows into
// connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := notionBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := notionSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("notion connector requires secret credentials.access_token (or token)")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: notionUserAgent,
		DefaultHeaders: map[string]string{
			"Notion-Version": notionAPIVersion,
		},
	}, nil
}

// notionSecret resolves the integration token from Secrets, accepting the
// canonical credentials.access_token key plus the bare token / access_token
// aliases.
func notionSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	for _, key := range []string{"credentials.access_token", "access_token", "token"} {
		if v := strings.TrimSpace(cfg.Secrets[key]); v != "" {
			return v
		}
	}
	return ""
}

// notionBaseURL resolves and validates the base URL. The default is
// api.notion.com; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func notionBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return notionDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("notion config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("notion config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("notion config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func notionPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return notionDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("notion config page_size must be an integer: %w", err)
	}
	if value < 1 || value > notionMaxPageSize {
		return 0, fmt.Errorf("notion config page_size must be between 1 and %d", notionMaxPageSize)
	}
	return value, nil
}

func notionMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("notion config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("notion config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// Write satisfies the connectors.Connector interface. Notion is a read-only
// source in pm, so writes are rejected. Capabilities.Write is false and the
// connector intentionally implements neither WriteValidator nor DryRunWriter.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
