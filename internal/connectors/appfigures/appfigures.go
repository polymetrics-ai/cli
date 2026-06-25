// Package appfigures implements the native pm Appfigures connector. It follows
// the declarative-HTTP template established by the stripe connector: a thin
// package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt/StringAt extraction) with Appfigures-specific stream definitions,
// endpoints, and response shapes.
//
// Appfigures (https://appfigures.com) is an app-store analytics platform; its
// v2 REST API (https://api.appfigures.com/v2/) authenticates a Personal Access
// Token as an OAuth2-style Bearer token. The connector is read-only.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package appfigures

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	appfiguresDefaultBaseURL  = "https://api.appfigures.com/v2"
	appfiguresDefaultPageSize = 100
	appfiguresMaxPageSize     = 500
	appfiguresUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("appfigures", New)
}

// New returns the Appfigures connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Appfigures connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "appfigures" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "appfigures",
		DisplayName:     "Appfigures",
		IntegrationType: "api",
		Description:     "Reads Appfigures app-store analytics: reviews, products, sales and ratings reports, and store categories via the Appfigures v2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Appfigures.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := appfiguresBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(appfiguresSecret(cfg)) == "" {
		return errors.New("appfigures connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of products/mine confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "products/mine", nil, nil, nil); err != nil {
		return fmt.Errorf("check appfigures: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: appfiguresStreams()}, nil
}

// Write is unsupported: Appfigures is a read-only analytics source with no safe
// reverse-ETL surface, so the connector advertises Write=false and rejects
// writes.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "reviews"
	}
	endpoint, ok := appfiguresStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("appfigures stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	switch endpoint.shape {
	case shapeKeyedObject:
		return c.readKeyedObject(ctx, r, endpoint, req, emit)
	default:
		pageSize, err := appfiguresPageSize(req.Config)
		if err != nil {
			return err
		}
		maxPages, err := appfiguresMaxPages(req.Config)
		if err != nil {
			return err
		}
		return c.readPaged(ctx, r, endpoint, req, pageSize, maxPages, emit)
	}
}

// readPaged drives Appfigures' page-number pagination. List endpoints such as
// reviews return {"total":N,"pages":P,"this_page":I,"<records>":[...]}; the next
// page is requested with page=<n>. The loop stops when this_page reaches pages
// or a short/empty page is returned. There is no connsdk paginator for this
// exact body shape, so the loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) readPaged(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, req connectors.ReadRequest, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := appfiguresQuery(req.Config)
	base.Set("count", strconv.Itoa(pageSize))

	page := 1
	for pageNum := 0; maxPages == 0 || pageNum < maxPages; pageNum++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read appfigures %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode appfigures %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) == 0 {
			return nil
		}
		totalPages := pagesFromBody(resp.Body)
		if totalPages > 0 && page >= totalPages {
			return nil
		}
		// Defensive stop: if the API does not report total pages, halt on a
		// short page rather than looping forever.
		if totalPages == 0 && len(records) < pageSize {
			return nil
		}
		page++
	}
	return nil
}

// readKeyedObject reads an endpoint whose body is a JSON object keyed by id
// (products/mine, reports/*). Each value object becomes one record. Keys are
// sorted for deterministic emission order.
func (c Connector) readKeyedObject(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	query := appfiguresQuery(req.Config)
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
	if err != nil {
		return fmt.Errorf("read appfigures %s: %w", endpoint.resource, err)
	}
	objects, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode appfigures %s: %w", endpoint.resource, err)
	}
	for _, obj := range objects {
		for _, item := range flattenKeyedObject(obj) {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
	}
	return nil
}

// flattenKeyedObject turns a top-level object keyed by id into a slice of its
// value objects, sorted by key for determinism. If a value is not itself an
// object (e.g. the body was already a flat record), the original object is
// returned as a single record.
func flattenKeyedObject(obj map[string]any) []map[string]any {
	keys := make([]string, 0, len(obj))
	allObjects := len(obj) > 0
	for k, v := range obj {
		keys = append(keys, k)
		if _, ok := v.(map[string]any); !ok {
			allObjects = false
		}
	}
	if !allObjects {
		return []map[string]any{obj}
	}
	sort.Strings(keys)
	out := make([]map[string]any, 0, len(keys))
	for _, k := range keys {
		out = append(out, obj[k].(map[string]any))
	}
	return out
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise appfigures credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                fmt.Sprintf("%s_fixture_%d", stream, i),
			"product":           int64(100 + i),
			"name":              fmt.Sprintf("Fixture %s %d", stream, i),
			"title":             fmt.Sprintf("Fixture title %d", i),
			"review":            "Great app, fixture review.",
			"author":            fmt.Sprintf("fixture_author_%d", i),
			"version":           "1.0.0",
			"date":              fmt.Sprintf("2026-01-0%dT00:00:00", i),
			"stars":             json.Number(strconv.Itoa(5 - (i % 2))),
			"iso":               "US",
			"has_response":      false,
			"weight":            int64(i),
			"developer":         "Fixture Dev",
			"vendor_identifier": fmt.Sprintf("com.fixture.app%d", i),
			"store":             "apple",
			"store_id":          int64(1),
			"added":             "2026-01-01T00:00:00",
			"updated":           "2026-01-02T00:00:00",
			"downloads":         int64(1000 * i),
			"revenue":           fmt.Sprintf("%d.00", 50*i),
			"breakdown":         "5:10,4:2,3:1,2:0,1:0",
			"average":           json.Number("4.5"),
			"device":            "iphone",
			"subtype":           "free",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := appfiguresBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := appfiguresSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("appfigures connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: appfiguresUserAgent,
	}, nil
}

// appfiguresQuery builds the per-request query params common to a stream:
// start_date filters, store, and group_by from config when present.
func appfiguresQuery(cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	if store := strings.TrimSpace(cfg.Config["search_store"]); store != "" {
		q.Set("store", store)
	}
	if groupBy := strings.TrimSpace(cfg.Config["group_by"]); groupBy != "" {
		q.Set("group_by", groupBy)
	}
	if start := strings.TrimSpace(cfg.Config["start_date"]); start != "" {
		q.Set("start", start)
	}
	if end := strings.TrimSpace(cfg.Config["end_date"]); end != "" {
		q.Set("end", end)
	}
	return q
}

func appfiguresSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// appfiguresBaseURL resolves and validates the base URL. The default is
// api.appfigures.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func appfiguresBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return appfiguresDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("appfigures config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("appfigures config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("appfigures config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func appfiguresPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return appfiguresDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("appfigures config page_size must be an integer: %w", err)
	}
	if value < 1 || value > appfiguresMaxPageSize {
		return 0, fmt.Errorf("appfigures config page_size must be between 1 and %d", appfiguresMaxPageSize)
	}
	return value, nil
}

func appfiguresMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("appfigures config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("appfigures config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// pagesFromBody reads the "pages" field from a paginated response body, falling
// back to 0 (unknown) when absent or unparseable.
func pagesFromBody(body []byte) int {
	raw, err := connsdk.StringAt(body, "pages")
	if err != nil || strings.TrimSpace(raw) == "" {
		return 0
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value < 0 {
		return 0
	}
	return value
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
