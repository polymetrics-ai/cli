// Package onesignal implements the native pm OneSignal connector. It is a
// declarative-HTTP per-system connector built on the stripe template: a thin
// package that composes the connsdk toolkit (Requester + Authorization: Basic
// API-key auth + RecordsAt extraction) with OneSignal-specific stream
// definitions, endpoints, and offset pagination.
//
// OneSignal splits credentials by scope: account-level endpoints (apps) use the
// user / organization auth key, while app-scoped endpoints (players,
// notifications, outcomes) use the per-app REST API key. Both are sent as
// "Authorization: Basic <key>" (the key is used directly, not base64 user:pass),
// so connsdk.APIKeyHeader with a "Basic " prefix is the matching authenticator.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package onesignal

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
	onesignalDefaultBaseURL  = "https://onesignal.com/api/v1"
	onesignalDefaultPageSize = 300
	onesignalMaxPageSize     = 300
	onesignalUserAgent       = "polymetrics-go-cli"
	// onesignalFixtureCreated is the deterministic created_at unix timestamp used
	// by fixture-mode records (2026-01-01T00:00:00Z in unix seconds).
	onesignalFixtureCreated int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("onesignal", New)
}

// New returns the OneSignal connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm OneSignal connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "onesignal" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "onesignal",
		DisplayName:     "OneSignal",
		IntegrationType: "api",
		Description:     "Reads OneSignal apps, devices, notifications, and outcomes through the OneSignal REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

// Check verifies the connector is configured well enough to talk to OneSignal. In
// fixture mode it short-circuits without a network call. Otherwise it does a
// bounded read of the account-level apps endpoint to confirm auth/connectivity.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := onesignalBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(userAuthKey(cfg)) == "" {
		return errors.New("onesignal connector requires secret user_auth_key")
	}
	r, err := c.requester(cfg, scopeUser)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "apps", nil, nil, nil); err != nil {
		return fmt.Errorf("check onesignal: %w", err)
	}
	return nil
}

// Write is unsupported: OneSignal is a read-only source connector. The method
// exists only to satisfy the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: onesignalStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "apps"
	}
	endpoint, ok := onesignalStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("onesignal stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config, endpoint.scope)
	if err != nil {
		return err
	}

	appID, err := onesignalAppID(req.Config)
	if err != nil {
		return err
	}
	if (endpoint.needsAppID || strings.Contains(endpoint.resource, "{app_id}")) && appID == "" {
		return fmt.Errorf("onesignal stream %q requires config app_id", stream)
	}
	resource := strings.ReplaceAll(endpoint.resource, "{app_id}", url.PathEscape(appID))

	base := url.Values{}
	if endpoint.needsAppID && appID != "" {
		base.Set("app_id", appID)
	}

	if !endpoint.paginated {
		return c.readSingle(ctx, r, resource, endpoint, base, emit)
	}

	pageSize, err := onesignalPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := onesignalMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, resource, endpoint, base, pageSize, maxPages, emit)
}

// readSingle reads a non-paginated endpoint (apps, outcomes) in one request.
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, resource string, endpoint streamEndpoint, base url.Values, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, resource, base, nil)
	if err != nil {
		return fmt.Errorf("read onesignal %s: %w", resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode onesignal %s: %w", resource, err)
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

// harvest drives OneSignal's offset/limit pagination. List endpoints return
// {records:[...], total_count, offset, limit}; the next page advances offset by
// the page size until offset >= total_count or a short page is returned. connsdk
// has an OffsetPaginator, but OneSignal's per-stream records path plus the
// total_count stop condition is cleanly expressed with this in-package loop built
// on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, resource string, endpoint streamEndpoint, base url.Values, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, resource, query, nil)
		if err != nil {
			return fmt.Errorf("read onesignal %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode onesignal %s page: %w", resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		// Stop on a short/empty page; otherwise advance by the records returned
		// and respect total_count when present.
		if len(records) == 0 || len(records) < pageSize {
			return nil
		}
		offset += len(records)
		if total, ok := parseIntField(resp.Body, "total_count"); ok && offset >= total {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise onesignal credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                  fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":                fmt.Sprintf("Fixture %s %d", stream, i),
			"identifier":          fmt.Sprintf("token_%d", i),
			"app_id":              "app_fixture",
			"players":             int64(100 * i),
			"messageable_players": int64(90 * i),
			"organization_id":     "org_fixture",
			"created_at":          onesignalFixtureCreated + int64(i),
			"updated_at":          onesignalFixtureCreated + int64(i),
			"device_type":         int64(1),
			"device_os":           "16.0",
			"language":            "en",
			"session_count":       int64(i),
			"external_user_id":    fmt.Sprintf("ext_%d", i),
			"invalid_identifier":  false,
			"last_active":         onesignalFixtureCreated + int64(i),
			"successful":          int64(10 * i),
			"failed":              int64(0),
			"errored":             int64(0),
			"converted":           int64(i),
			"remaining":           int64(0),
			"queued_at":           onesignalFixtureCreated + int64(i),
			"send_after":          onesignalFixtureCreated + int64(i),
			"completed_at":        onesignalFixtureCreated + int64(i) + 60,
			"canceled":            false,
			"value":               float64(i),
			"aggregation":         "count",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the scope-appropriate
// Authorization: Basic API-key auth and the resolved base URL. The secret only
// ever flows into connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig, scope authScope) (*connsdk.Requester, error) {
	base, err := onesignalBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	var key, label string
	switch scope {
	case scopeUser:
		key, label = userAuthKey(cfg), "user_auth_key"
	default:
		key, label = appAPIKey(cfg), "app_api_key"
	}
	if strings.TrimSpace(key) == "" {
		return nil, fmt.Errorf("onesignal connector requires secret %s", label)
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", key, "Basic "),
		UserAgent: onesignalUserAgent,
	}, nil
}

func appAPIKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	if v := cfg.Secrets["app_api_key"]; v != "" {
		return v
	}
	return cfg.Secrets["rest_api_key"]
}

func userAuthKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["user_auth_key"]
}

func onesignalAppID(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return "", nil
	}
	return strings.TrimSpace(cfg.Config["app_id"]), nil
}

// onesignalBaseURL resolves and validates the base URL. The default is the
// OneSignal legacy API host; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func onesignalBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return onesignalDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("onesignal config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("onesignal config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("onesignal config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func onesignalPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return onesignalDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("onesignal config page_size must be an integer: %w", err)
	}
	if value < 1 || value > onesignalMaxPageSize {
		return 0, fmt.Errorf("onesignal config page_size must be between 1 and %d", onesignalMaxPageSize)
	}
	return value, nil
}

func onesignalMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("onesignal config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("onesignal config max_pages must be 0 for unlimited or a positive integer")
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

// parseIntField reads an integer field out of a JSON body at a dotted path.
func parseIntField(body []byte, path string) (int, bool) {
	s, err := connsdk.StringAt(body, path)
	if err != nil || strings.TrimSpace(s) == "" {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, false
	}
	return n, true
}
