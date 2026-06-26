// Package kisi implements the native pm Kisi connector. Kisi is a cloud-based
// physical access-control platform; this connector reads members, locks, groups,
// users, and logins from the Kisi REST API.
//
// It is a declarative-HTTP per-system connector built on the connsdk toolkit
// (Requester + APIKeyHeader auth + offset/limit pagination over top-level JSON
// arrays), modeled on the stripe reference connector. It self-registers with the
// connectors registry via RegisterFactory in init().
//
// Kisi authenticates with an "Authorization: KISI-LOGIN <api_key>" header, paginates
// list endpoints with limit/offset query parameters, and returns each list as a
// top-level JSON array. The API is full-refresh only (no incremental cursor) and
// exposes no safe reverse-ETL writes, so the connector is read-only.
package kisi

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
	kisiDefaultBaseURL  = "https://api.kisi.io"
	kisiDefaultPageSize = 100
	kisiMaxPageSize     = 100
	kisiUserAgent       = "polymetrics-go-cli"
	kisiAuthScheme      = "KISI-LOGIN "
)

func init() {
	connectors.RegisterFactory("kisi", New)
}

// New returns the Kisi connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Kisi connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "kisi" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "kisi",
		DisplayName:     "Kisi",
		IntegrationType: "api",
		Description:     "Reads Kisi physical access-control data: members, locks, groups, users, and logins via the Kisi REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Kisi. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := kisiBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(kisiSecret(cfg)) == "" {
		return errors.New("kisi connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the members list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "members", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check kisi: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: kisiStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "members"
	}
	endpoint, ok := kisiStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("kisi stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := kisiPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := kisiMaxPages(req.Config)
	if err != nil {
		return err
	}

	paginator := &connsdk.OffsetPaginator{
		LimitParam:  "limit",
		OffsetParam: "offset",
		PageSize:    pageSize,
	}
	// Kisi list endpoints return a top-level JSON array (recordsPath ""). Each
	// element is mapped through the stream's mapper before it is emitted.
	mapEmit := func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, nil, paginator, "", maxPages, mapEmit); err != nil {
		return fmt.Errorf("read kisi %s: %w", endpoint.resource, err)
	}
	return nil
}

// Write is unsupported: Kisi is read-only for pm (no safe reverse-ETL writes for
// a physical access-control system). It satisfies the connectors.Connector
// interface by reporting the unsupported-operation sentinel.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise kisi credential-free (mirrors the stripe
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                           int64(i),
			"name":                         fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"email":                        fmt.Sprintf("fixture+%d@example.com", i),
			"description":                  "fixture record",
			"type":                         "fixture",
			"role_id":                      int64(1),
			"user_id":                      int64(i),
			"place_id":                     int64(1),
			"login_count":                  int64(i),
			"confirmed":                    true,
			"access_enabled":               true,
			"online":                       true,
			"geofence_restriction_enabled": false,
			"created_at":                   "2026-01-01T00:00:00Z",
			"updated_at":                   "2026-01-02T00:00:00Z",
			"last_used_at":                 "2026-01-03T00:00:00Z",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with KISI-LOGIN header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := kisiBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := kisiSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("kisi connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, kisiAuthScheme),
		UserAgent: kisiUserAgent,
	}, nil
}

func kisiSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// kisiBaseURL resolves and validates the base URL. The default is api.kisi.io;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func kisiBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return kisiDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("kisi config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("kisi config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("kisi config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func kisiPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return kisiDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("kisi config page_size must be an integer: %w", err)
	}
	if value < 1 || value > kisiMaxPageSize {
		return 0, fmt.Errorf("kisi config page_size must be between 1 and %d", kisiMaxPageSize)
	}
	return value, nil
}

func kisiMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("kisi config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("kisi config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
