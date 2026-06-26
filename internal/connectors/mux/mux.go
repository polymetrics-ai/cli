// Package mux implements the native pm Mux connector. It is a declarative-HTTP
// per-system connector built on the same shape as the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + Basic auth + RecordsAt
// extraction) with Mux-specific stream definitions and endpoints.
//
// Mux exposes its Video and System resources behind HTTP Basic auth, where the
// username is the API access token id and the password is its secret key. List
// endpoints return {"data":[...]} and paginate by page number (page + limit).
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package mux

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
	muxDefaultBaseURL  = "https://api.mux.com"
	muxDefaultPageSize = 25
	muxMaxPageSize     = 100
	muxUserAgent       = "polymetrics-go-cli"
	// muxFixtureCreated is the deterministic created_at (unix seconds) used by
	// fixture-mode records (2026-01-01T00:00:00Z).
	muxFixtureCreated int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("mux", New)
}

// New returns the Mux connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Mux connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "mux" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "mux",
		DisplayName:     "Mux",
		IntegrationType: "api",
		Description:     "Reads Mux Video assets, live streams, direct uploads, and system signing keys through the Mux REST API using HTTP Basic authentication.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

// Check verifies the connector is configured well enough to talk to Mux. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := muxBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(muxUsername(cfg)) == "" {
		return errors.New("mux connector requires config username (Mux token id)")
	}
	if strings.TrimSpace(muxSecret(cfg)) == "" {
		return errors.New("mux connector requires secret password (Mux token secret)")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the assets list confirms auth and connectivity without
	// mutating anything.
	query := url.Values{"limit": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "video/v1/assets", query, nil, nil); err != nil {
		return fmt.Errorf("check mux: %w", err)
	}
	return nil
}

// Write is unsupported: Mux is a read-only source connector (Capabilities.Write
// is false). It satisfies the connectors.Connector interface by returning
// ErrUnsupportedOperation, mirroring the built-in read-only connectors.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: muxStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "assets"
	}
	endpoint, ok := muxStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("mux stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := muxPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := muxMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Mux's page-number pagination. Mux list endpoints return
// {"data":[...]} and accept page + limit query params; the loop advances the
// page number until a page returns fewer than `limit` records (a short page),
// mirroring connsdk.PageNumberPaginator semantics. It is implemented here so the
// record id (used for nothing here) and the short-page stop condition stay
// explicit alongside connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))

	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read mux %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode mux %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer records than the requested limit) means there is no
		// further page to fetch.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise mux credential-free (mirrors stripe's fixture
// intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                      fmt.Sprintf("%s_fixture_%d", stream, i),
			"status":                  "ready",
			"created_at":              strconv.FormatInt(muxFixtureCreated+int64(i), 10),
			"duration":                float64(60 * i),
			"max_resolution_tier":     "1080p",
			"encoding_tier":           "smart",
			"mp4_support":             "none",
			"master_access":           "none",
			"test":                    true,
			"stream_key":              fmt.Sprintf("sk_fixture_%d", i),
			"latency_mode":            "low",
			"reconnect_window":        float64(60),
			"max_continuous_duration": float64(43200),
			"asset_id":                fmt.Sprintf("asset_fixture_%d", i),
			"url":                     fmt.Sprintf("https://storage.example.com/upload/%d", i),
			"timeout":                 float64(3600),
			"cors_origin":             "*",
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

// requester builds a connsdk.Requester wired with HTTP Basic auth and the
// resolved base URL. The secret only ever flows into connsdk.Basic; it is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := muxBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	username := muxUsername(cfg)
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("mux connector requires config username (Mux token id)")
	}
	secret := muxSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("mux connector requires secret password (Mux token secret)")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, secret),
		UserAgent: muxUserAgent,
	}, nil
}

func muxUsername(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["username"]
}

func muxSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// muxBaseURL resolves and validates the base URL. The default is api.mux.com;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func muxBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return muxDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("mux config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("mux config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("mux config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func muxPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return muxDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mux config page_size must be an integer: %w", err)
	}
	if value < 1 || value > muxMaxPageSize {
		return 0, fmt.Errorf("mux config page_size must be between 1 and %d", muxMaxPageSize)
	}
	return value, nil
}

func muxMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mux config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("mux config max_pages must be 0 for unlimited or a positive integer")
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
