// Package awinadvertiser implements the native pm Awin Advertiser connector. It
// is a declarative-HTTP per-system connector that composes the connsdk toolkit
// (Requester + Bearer auth + RecordsAt extraction over top-level arrays + page
// pagination) with Awin-specific stream definitions and endpoints. It follows the
// stripe template shape.
//
// Like the other connectors it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package awinadvertiser

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
	awinDefaultBaseURL  = "https://api.awin.com"
	awinDefaultPageSize = 100
	awinMaxPageSize     = 100
	awinUserAgent       = "polymetrics-go-cli"
	awinDateLayout      = "2006-01-02"
	// awinFixtureDate is the deterministic transactionDate used by fixture records.
	awinFixtureDate = "2026-01-01T00:00:00"
)

func init() {
	connectors.RegisterFactory("awin-advertiser", New)
}

// New returns the Awin Advertiser connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Awin Advertiser connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "awin-advertiser" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "awin-advertiser",
		DisplayName:     "Awin Advertiser",
		IntegrationType: "api",
		Description:     "Reads Awin advertiser transactions, publisher-aggregated performance reports, and publisher relationships through the Awin Advertiser REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Awin. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := awinBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(awinSecret(cfg)) == "" {
		return errors.New("awin-advertiser connector requires secret api_key")
	}
	advertiserID, err := awinAdvertiserID(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the publisher list confirms auth and connectivity
	// without mutating anything.
	path := fmt.Sprintf("advertisers/%s/publishers/", advertiserID)
	if err := r.DoJSON(ctx, http.MethodGet, path, nil, nil, nil); err != nil {
		return fmt.Errorf("check awin-advertiser: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: awinStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: an Awin stream starts with an
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
		stream = "transactions"
	}
	endpoint, ok := awinStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("awin-advertiser stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	advertiserID, err := awinAdvertiserID(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := awinPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := awinMaxPages(req.Config)
	if err != nil {
		return err
	}
	base, err := awinDateWindow(req, endpoint)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("advertisers/%s/%s", advertiserID, endpoint.resource)
	return c.harvest(ctx, r, path, base, pageSize, maxPages, endpoint.mapRecord, emit)
}

// harvest drives page-number pagination over Awin's top-level JSON array
// responses. Awin returns a bare array per request; the connector advances the
// `page` param and stops when a short page (fewer than pageSize records) comes
// back. The loop lives here, built on connsdk.Requester + connsdk.RecordsAt with
// a root ("") path so the bare array is decoded directly.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, base url.Values, pageSize, maxPages int, mapRecord func(map[string]any) connectors.Record, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))
		query.Set("pageSize", strconv.Itoa(pageSize))
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read awin-advertiser %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode awin-advertiser %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	endpoint := awinStreamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                int64(1000 + i),
			"advertiserId":      int64(42),
			"publisherId":       int64(i),
			"publisherName":     fmt.Sprintf("Publisher %d", i),
			"name":              fmt.Sprintf("Publisher %d", i),
			"siteName":          fmt.Sprintf("site-%d.example.com", i),
			"displayUrl":        fmt.Sprintf("https://site-%d.example.com", i),
			"kind":              "standard",
			"status":            "active",
			"transactionDate":   awinFixtureDate,
			"validationDate":    awinFixtureDate,
			"type":              "Commission group transaction",
			"transactionStatus": "approved",
			"region":            "GB",
			"currency":          "GBP",
			"impressions":       int64(100 * i),
			"clicks":            int64(10 * i),
			"totalNo":           int64(i),
			"totalComm":         float64(i) * 12.5,
			"saleAmount":        map[string]any{"amount": float64(i) * 100, "currency": "GBP"},
			"commissionAmount":  map[string]any{"amount": float64(i) * 12.5, "currency": "GBP"},
			"totalSaleAmount":   map[string]any{"amount": float64(i) * 100, "currency": "GBP"},
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

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := awinBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := awinSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("awin-advertiser connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: awinUserAgent,
	}, nil
}

// awinDateWindow builds the startDate/endDate/timezone query window for
// date-windowed resources (e.g. transactions). The lower bound comes from the
// incremental cursor (if any) or the start_date config; the upper bound is now.
// Non-windowed resources (reports/publishers) receive an empty base.
func awinDateWindow(req connectors.ReadRequest, endpoint streamEndpoint) (url.Values, error) {
	base := url.Values{}
	if endpoint.dateType == "" {
		return base, nil
	}
	lower := strings.TrimSpace(connsdk.Cursor(req.State))
	if lower == "" {
		lower = strings.TrimSpace(req.Config.Config["start_date"])
	}
	if lower != "" {
		// Accept either a date or an RFC3339 timestamp; normalize to Awin's
		// "YYYY-MM-DDTHH:MM:SS" form.
		start, err := normalizeAwinDate(lower)
		if err != nil {
			return nil, fmt.Errorf("awin-advertiser config start_date: %w", err)
		}
		base.Set("startDate", start)
	}
	base.Set("endDate", time.Now().UTC().Format("2006-01-02T15:04:05"))
	base.Set("dateType", endpoint.dateType)
	base.Set("timezone", "UTC")
	return base, nil
}

func normalizeAwinDate(value string) (string, error) {
	value = strings.TrimSpace(value)
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC().Format("2006-01-02T15:04:05"), nil
	}
	if t, err := time.Parse(awinDateLayout, value); err == nil {
		return t.Format("2006-01-02T15:04:05"), nil
	}
	return "", fmt.Errorf("must be YYYY-MM-DD or RFC3339, got %q", value)
}

func awinSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

func awinAdvertiserID(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return "", errors.New("awin-advertiser connector requires config advertiserId")
	}
	id := strings.TrimSpace(cfg.Config["advertiserId"])
	if id == "" {
		return "", errors.New("awin-advertiser connector requires config advertiserId")
	}
	for _, r := range id {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("awin-advertiser config advertiserId must be numeric, got %q", id)
		}
	}
	return id, nil
}

// awinBaseURL resolves and validates the base URL. The default is api.awin.com;
// any override must be an absolute https (or http for local test servers) URL with
// a host to bound SSRF risk.
func awinBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return awinDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("awin-advertiser config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("awin-advertiser config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("awin-advertiser config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func awinPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return awinDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("awin-advertiser config page_size must be an integer: %w", err)
	}
	if value < 1 || value > awinMaxPageSize {
		return 0, fmt.Errorf("awin-advertiser config page_size must be between 1 and %d", awinMaxPageSize)
	}
	return value, nil
}

func awinMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("awin-advertiser config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("awin-advertiser config max_pages must be 0 for unlimited or a positive integer")
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

// Write is unsupported: the Awin Advertiser connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
