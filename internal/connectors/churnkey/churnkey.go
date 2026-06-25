// Package churnkey implements the native pm Churnkey connector. It is a
// declarative-HTTP per-system connector modeled on the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + two custom auth headers
// + RecordsAt extraction) with Churnkey-specific stream definitions, endpoints,
// and limit/skip offset pagination.
//
// Churnkey exposes a read-only "Data API" (https://api.churnkey.co/v1/data)
// authenticated with two headers: x-ck-api-key (the secret api_key) and x-ck-app
// (the application id). There are no safe reverse-ETL write actions — the only
// mutating endpoints are GDPR deletes — so this connector is read-only.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package churnkey

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
	churnkeyDefaultBaseURL  = "https://api.churnkey.co/v1/data"
	churnkeyDefaultPageSize = 100
	// churnkeyMaxPageSize matches the Churnkey Data API cap of 10,000 records
	// per request for the sessions endpoint.
	churnkeyMaxPageSize = 10000
	churnkeyUserAgent   = "polymetrics-go-cli"
	churnkeyAPIKeyHdr   = "x-ck-api-key"
	churnkeyAppHdr      = "x-ck-app"
	// churnkeyFixtureCreated is the deterministic createdAt used by fixture-mode
	// records.
	churnkeyFixtureCreated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("churnkey", New)
}

// New returns the Churnkey connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Churnkey connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "churnkey" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "churnkey",
		DisplayName:     "Churnkey",
		IntegrationType: "api",
		Description:     "Reads Churnkey cancel-flow sessions and aggregated session counts through the read-only Churnkey Data API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Churnkey. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := churnkeyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(churnkeySecret(cfg)) == "" {
		return errors.New("churnkey connector requires secret api_key")
	}
	if strings.TrimSpace(churnkeyApp(cfg)) == "" {
		return errors.New("churnkey connector requires config x-ck-app (app id)")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the sessions list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "sessions", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check churnkey: %w", err)
	}
	return nil
}

// Write is unsupported: Churnkey's Data API is read-only for reverse ETL
// purposes (its only mutating endpoints are GDPR deletes), so the connector
// declines all writes.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: churnkeyStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "sessions"
	}
	endpoint, ok := churnkeyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("churnkey stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := churnkeyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := churnkeyMaxPages(req.Config)
	if err != nil {
		return err
	}
	if !endpoint.paginated {
		return c.readSinglePage(ctx, r, endpoint, emit)
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Churnkey's limit/skip offset pagination over a top-level JSON
// array. Churnkey responses carry no envelope or has_more flag, so the loop
// stops when a page returns fewer records than the requested page size. There is
// no connsdk paginator for this exact top-level-array offset shape, so the loop
// lives here, built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	skip := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		if skip > 0 {
			query.Set("skip", strconv.Itoa(skip))
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read churnkey %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode churnkey %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		skip += pageSize
	}
	return nil
}

// readSinglePage reads an unpaginated endpoint (session-aggregation) once and
// emits every record in the returned array.
func (c Connector) readSinglePage(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read churnkey %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode churnkey %s: %w", endpoint.resource, err)
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

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise churnkey credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var item map[string]any
		if stream == "session_aggregation" {
			item = map[string]any{
				"month":           "2026-0" + strconv.Itoa(i),
				"offerType":       "DISCOUNT",
				"billingInterval": "month",
				"canceled":        i%2 == 0,
				"count":           int64(10 * i),
			}
		} else {
			item = map[string]any{
				"_id":       fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
				"org":       "org_fixture",
				"provider":  "stripe",
				"mode":      "live",
				"aborted":   false,
				"canceled":  i%2 == 1,
				"createdAt": churnkeyFixtureCreated,
				"updatedAt": churnkeyFixtureCreated,
				"customer": map[string]any{
					"id":              fmt.Sprintf("cus_fixture_%d", i),
					"email":           fmt.Sprintf("fixture+%d@example.com", i),
					"planId":          "plan_pro",
					"billingInterval": "month",
				},
				"acceptedOffer": map[string]any{
					"offerType": "PAUSE",
				},
			}
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "churnkey"
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

// requester builds a connsdk.Requester wired with the two Churnkey auth headers,
// the resolved base URL, and the x-ck-app header. The secret only ever flows into
// the x-ck-api-key header; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := churnkeyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := churnkeySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("churnkey connector requires secret api_key")
	}
	app := strings.TrimSpace(churnkeyApp(cfg))
	if app == "" {
		return nil, errors.New("churnkey connector requires config x-ck-app (app id)")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(churnkeyAPIKeyHdr, secret, ""),
		UserAgent: churnkeyUserAgent,
		DefaultHeaders: map[string]string{
			churnkeyAppHdr: app,
		},
	}, nil
}

func churnkeySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// churnkeyApp resolves the x-ck-app application id from config. It accepts both
// the catalog-canonical "x-ck-app" key and a friendlier "app_id" alias.
func churnkeyApp(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	if v := strings.TrimSpace(cfg.Config["x-ck-app"]); v != "" {
		return v
	}
	return strings.TrimSpace(cfg.Config["app_id"])
}

// churnkeyBaseURL resolves and validates the base URL. The default is
// api.churnkey.co; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func churnkeyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return churnkeyDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("churnkey config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("churnkey config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("churnkey config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func churnkeyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return churnkeyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("churnkey config page_size must be an integer: %w", err)
	}
	if value < 1 || value > churnkeyMaxPageSize {
		return 0, fmt.Errorf("churnkey config page_size must be between 1 and %d", churnkeyMaxPageSize)
	}
	return value, nil
}

func churnkeyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("churnkey config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("churnkey config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
