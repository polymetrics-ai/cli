// Package bigmailer implements the native pm BigMailer source connector. It is a
// declarative-HTTP per-system connector modeled on the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + X-API-Key auth +
// RecordsAt extraction + cursor pagination) with BigMailer-specific stream
// definitions and endpoints.
//
// BigMailer is read-only here (no safe reverse-ETL writes are exposed), so the
// connector advertises Write=false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package bigmailer

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
	bigmailerDefaultBaseURL  = "https://api.bigmailer.io/v1"
	bigmailerDefaultPageSize = 100
	bigmailerMaxPageSize     = 100
	bigmailerUserAgent       = "polymetrics-go-cli"
	bigmailerAuthHeader      = "X-API-Key"
	// bigmailerFixtureCreated is the deterministic `created` timestamp used by
	// the fixture-mode records (2026-01-01T00:00:00Z in unix seconds).
	bigmailerFixtureCreated int64 = 1767225600
	// bigmailerMaxBrands bounds how many brands a substream read will fan out
	// over, so a misbehaving account cannot cause an unbounded brand crawl.
	bigmailerMaxBrands = 1000
)

func init() {
	connectors.RegisterFactory("bigmailer", New)
}

// New returns the BigMailer connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm BigMailer connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "bigmailer" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "bigmailer",
		DisplayName:     "BigMailer",
		IntegrationType: "api",
		Description:     "Reads BigMailer brands, users, contacts, lists, and custom fields through the BigMailer REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to BigMailer. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := bigmailerBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(bigmailerSecret(cfg)) == "" {
		return errors.New("bigmailer connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the brands list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "brands", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check bigmailer: %w", err)
	}
	return nil
}

// Write is unsupported: BigMailer is read-only in this connector (no safe
// reverse-ETL action set is exposed). It satisfies the connectors.Connector
// interface but always rejects writes.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: bigmailerStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "brands"
	}
	endpoint, ok := bigmailerStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("bigmailer stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := bigmailerPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := bigmailerMaxPages(req.Config)
	if err != nil {
		return err
	}

	if !endpoint.brandSub {
		return c.harvest(ctx, r, endpoint.path, nil, pageSize, maxPages, endpoint.mapRecord, emit)
	}
	return c.harvestSubstream(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives BigMailer's cursor pagination for a single endpoint. List
// responses are {data:[...], has_more:bool, cursor:string}; the next page is
// requested with cursor=<value> until has_more is false. stamp, when non-nil, is
// applied to every record before mapping (used to add brand_id on substreams).
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, stamp map[string]any, pageSize, maxPages int, mapRecord func(map[string]any) connectors.Record, emit func(connectors.Record) error) error {
	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read bigmailer %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode bigmailer %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			for k, v := range stamp {
				item[k] = v
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		hasMore, err := connsdk.StringAt(resp.Body, "has_more")
		if err != nil {
			return fmt.Errorf("decode bigmailer %s has_more: %w", path, err)
		}
		next, err := connsdk.StringAt(resp.Body, "cursor")
		if err != nil {
			return fmt.Errorf("decode bigmailer %s cursor: %w", path, err)
		}
		if hasMore != "true" || strings.TrimSpace(next) == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

// harvestSubstream reads a brand-scoped stream by first listing brand ids, then
// paginating /brands/{brand_id}/<resource> for each, stamping brand_id onto every
// emitted record.
func (c Connector) harvestSubstream(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	brandIDs, err := c.listBrandIDs(ctx, r, pageSize)
	if err != nil {
		return err
	}
	for _, brandID := range brandIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := "brands/" + url.PathEscape(brandID) + "/" + endpoint.resource
		stamp := map[string]any{"brand_id": brandID}
		if err := c.harvest(ctx, r, path, stamp, pageSize, maxPages, endpoint.mapRecord, emit); err != nil {
			return err
		}
	}
	return nil
}

// listBrandIDs collects every brand id, bounded by bigmailerMaxBrands.
func (c Connector) listBrandIDs(ctx context.Context, r *connsdk.Requester, pageSize int) ([]string, error) {
	var ids []string
	cursor := ""
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, "brands", query, nil)
		if err != nil {
			return nil, fmt.Errorf("list bigmailer brands: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return nil, fmt.Errorf("decode bigmailer brands page: %w", err)
		}
		for _, item := range records {
			id := stringField(item, "id")
			if id == "" {
				continue
			}
			ids = append(ids, id)
			if len(ids) >= bigmailerMaxBrands {
				return ids, nil
			}
		}
		hasMore, err := connsdk.StringAt(resp.Body, "has_more")
		if err != nil {
			return nil, fmt.Errorf("decode bigmailer brands has_more: %w", err)
		}
		next, err := connsdk.StringAt(resp.Body, "cursor")
		if err != nil {
			return nil, fmt.Errorf("decode bigmailer brands cursor: %w", err)
		}
		if hasMore != "true" || strings.TrimSpace(next) == "" {
			return ids, nil
		}
		cursor = next
	}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise bigmailer credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	endpoint := bigmailerStreamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               fmt.Sprintf("%s_fixture_%d", stream, i),
			"brand_id":         "brand_fixture_1",
			"name":             fmt.Sprintf("Fixture %d", i),
			"from_name":        fmt.Sprintf("Fixture %d", i),
			"from_email":       fmt.Sprintf("fixture+%d@example.com", i),
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"connection_id":    "conn_fixture_1",
			"num_contacts":     int64(10 * i),
			"contact_limit":    int64(1000),
			"role":             "admin",
			"tag":              fmt.Sprintf("field_%d", i),
			"type":             "text",
			"unsubscribe_all":  false,
			"num_soft_bounces": int64(0),
			"num_hard_bounces": int64(0),
			"num_complaints":   int64(0),
			"created":          bigmailerFixtureCreated + int64(i),
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with X-API-Key auth and the resolved
// base URL. The secret only ever flows into connsdk.APIKeyHeader; it is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := bigmailerBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := bigmailerSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("bigmailer connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(bigmailerAuthHeader, secret, ""),
		UserAgent: bigmailerUserAgent,
	}, nil
}

func bigmailerSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// bigmailerBaseURL resolves and validates the base URL. The default is
// api.bigmailer.io; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func bigmailerBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return bigmailerDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("bigmailer config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("bigmailer config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("bigmailer config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func bigmailerPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return bigmailerDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("bigmailer config page_size must be an integer: %w", err)
	}
	if value < 1 || value > bigmailerMaxPageSize {
		return 0, fmt.Errorf("bigmailer config page_size must be between 1 and %d", bigmailerMaxPageSize)
	}
	return value, nil
}

func bigmailerMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("bigmailer config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("bigmailer config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
