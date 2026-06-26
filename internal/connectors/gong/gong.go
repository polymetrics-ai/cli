// Package gong implements the native pm Gong connector. It is a declarative-HTTP
// per-system connector built on the connsdk toolkit, mirroring the stripe
// reference connector: a thin package that composes Requester + Basic auth +
// RecordsAt extraction + Gong's cursor pagination with Gong-specific stream
// definitions and endpoints.
//
// Gong exposes a read-only REST API (https://api.gong.io/v2). Auth is HTTP Basic
// using a generated access key and access key secret (the connector's simplest
// supported scheme; OAuth2 is also offered by Gong but not required here).
// Listing endpoints paginate with a cursor token returned at records.cursor in
// the body and supplied back as the "cursor" query parameter; the absence of the
// "records" object signals the last page.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package gong

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	gongDefaultBaseURL  = "https://api.gong.io/v2"
	gongDefaultPageSize = 100
	gongMaxPageSize     = 100
	gongUserAgent       = "polymetrics-go-cli"
	// gongFixtureStart is the deterministic ISO-8601 timestamp used by the
	// fixture-mode records.
	gongFixtureStart = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("gong", New)
}

// New returns the Gong connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Gong connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "gong" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "gong",
		DisplayName:     "Gong",
		IntegrationType: "api",
		Description:     "Reads Gong users, calls, and scorecards through the Gong REST API (read-only).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Gong. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := gongBaseURL(cfg); err != nil {
		return err
	}
	key, secret := gongSecret(cfg)
	if strings.TrimSpace(key) == "" || strings.TrimSpace(secret) == "" {
		return errors.New("gong connector requires secrets credentials.access_key and credentials.access_key_secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the users list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "users", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check gong: %w", err)
	}
	return nil
}

// Write is unsupported: Gong is a read-only source connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: gongStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Gong stream starts with an
// empty incremental cursor (full sync), which the start_date config can raise at
// read time.
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
	endpoint, ok := gongStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("gong stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	lower, err := incrementalLowerBound(req)
	if err != nil {
		return err
	}
	pageSize, err := gongPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := gongMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, lower, emit)
}

// harvest drives Gong's cursor pagination. Listing responses look like
// {records:{cursor:"..."}, <recordsPath>:[...]}; the next page is requested with
// cursor=<records.cursor>, and the absence of the "records" object (or an empty
// cursor) ends the walk. There is no connsdk paginator for this exact shape, so
// the loop lives here, built on Requester + RecordsAt + StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, fromDateTime string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	if fromDateTime != "" && endpoint.startParam != "" {
		base.Set(endpoint.startParam, fromDateTime)
	}

	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read gong %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode gong %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "records.cursor")
		if err != nil {
			return fmt.Errorf("decode gong %s cursor: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise gong credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	endpoint := gongStreamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            fmt.Sprintf("%s_fixture_%d", stream, i),
			"emailAddress":  fmt.Sprintf("fixture+%d@example.com", i),
			"firstName":     fmt.Sprintf("Fixture%d", i),
			"lastName":      "User",
			"title":         "Account Executive",
			"active":        true,
			"phoneNumber":   "",
			"managerId":     "mgr_fixture_1",
			"created":       gongFixtureStart,
			"started":       gongFixtureStart,
			"scheduled":     gongFixtureStart,
			"duration":      int64(600 * i),
			"direction":     "Inbound",
			"system":        "Zoom",
			"scope":         "Internal",
			"media":         "Video",
			"language":      "eng",
			"url":           fmt.Sprintf("https://app.gong.io/call?id=%d", i),
			"isPrivate":     false,
			"scorecardId":   fmt.Sprintf("sc_fixture_%d", i),
			"scorecardName": fmt.Sprintf("Scorecard %d", i),
			"workspaceId":   "ws_fixture_1",
			"enabled":       true,
			"updated":       gongFixtureStart,
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

// requester builds a connsdk.Requester wired with Basic auth and the resolved
// base URL. The credentials only ever flow into connsdk.Basic; they are never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := gongBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	key, secret := gongSecret(cfg)
	if strings.TrimSpace(key) == "" || strings.TrimSpace(secret) == "" {
		return nil, errors.New("gong connector requires secrets credentials.access_key and credentials.access_key_secret")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(key, secret),
		UserAgent: gongUserAgent,
	}, nil
}

// incrementalLowerBound returns the ISO-8601 lower bound for the fromDateTime
// filter, derived from the incremental cursor (if any) or else the start_date
// config. An empty result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor, nil
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return "", nil
	}
	if _, err := time.Parse(time.RFC3339, startDate); err != nil {
		return "", fmt.Errorf("gong config start_date must be RFC3339: %w", err)
	}
	return startDate, nil
}

// gongSecret resolves the access key and secret from the Secrets map, accepting
// both the dotted catalog form (credentials.access_key) and the bare leaf form
// (access_key) for robustness across config shapes.
func gongSecret(cfg connectors.RuntimeConfig) (key, secret string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	key = firstNonEmpty(cfg.Secrets["credentials.access_key"], cfg.Secrets["access_key"])
	secret = firstNonEmpty(cfg.Secrets["credentials.access_key_secret"], cfg.Secrets["access_key_secret"])
	return key, secret
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// gongBaseURL resolves and validates the base URL. The default is api.gong.io;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func gongBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return gongDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("gong config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("gong config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("gong config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func gongPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return gongDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gong config page_size must be an integer: %w", err)
	}
	if value < 1 || value > gongMaxPageSize {
		return 0, fmt.Errorf("gong config page_size must be between 1 and %d", gongMaxPageSize)
	}
	return value, nil
}

func gongMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gong config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("gong config max_pages must be 0 for unlimited or a positive integer")
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
