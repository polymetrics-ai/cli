// Package nasa implements the native pm NASA connector. It follows the
// declarative-HTTP template established by the stripe connector: a thin package
// that composes the connsdk toolkit (Requester + api_key query auth + RecordsAt
// extraction) with NASA-specific stream definitions and endpoints across the
// NASA Open APIs (api.nasa.gov) — APOD, NeoWs asteroids, EPIC, and Mars rover
// photos.
//
// It is read-only: the NASA Open APIs expose no reverse-ETL writes. Like the
// other per-system connectors it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package nasa

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
	nasaDefaultBaseURL = "https://api.nasa.gov"
	nasaUserAgent      = "polymetrics-go-cli"
	// nasaMaxBrowsePages bounds the neo_browse pagination by default so a sync
	// cannot run unbounded against the (large) NeoWs dataset.
	nasaDefaultMaxPages = 5
)

func init() {
	connectors.RegisterFactory("nasa", New)
}

// New returns the NASA connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm NASA connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "nasa" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "nasa",
		DisplayName:     "NASA",
		IntegrationType: "api",
		Description:     "Reads NASA Open API data: Astronomy Picture of the Day, Near-Earth Objects (NeoWs feed and browse), EPIC Earth imagery, and Mars rover photos. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to NASA. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := nasaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(nasaSecret(cfg)) == "" {
		return errors.New("nasa connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded APOD read confirms the api_key and connectivity without mutating
	// anything (NASA APIs are read-only).
	if err := r.DoJSON(ctx, http.MethodGet, "planetary/apod", nil, nil, nil); err != nil {
		return fmt.Errorf("check nasa: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: nasaStreams()}, nil
}

// Write satisfies the connectors.Connector interface. The NASA Open APIs are
// read-only, so reverse-ETL writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "apod"
	}
	spec, ok := nasaStreamSpecs[stream]
	if !ok {
		return fmt.Errorf("nasa stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, spec, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	query := streamQuery(stream, req.Config)

	switch spec.paginate {
	case paginatePage:
		maxPages, err := nasaMaxPages(req.Config)
		if err != nil {
			return err
		}
		return c.harvestPaged(ctx, r, spec, query, maxPages, emit)
	default:
		return c.harvestSingle(ctx, r, spec, query, emit)
	}
}

// harvestSingle fetches one response and emits every record under the stream's
// arrayPath (which may be a single top-level object, e.g. APOD).
func (c Connector) harvestSingle(ctx context.Context, r *connsdk.Requester, spec streamSpec, query url.Values, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, spec.resource, query, nil)
	if err != nil {
		return fmt.Errorf("read nasa %s: %w", spec.resource, err)
	}
	return emitRecords(ctx, resp.Body, spec, emit)
}

// harvestPaged walks NeoWs page-based pagination: each response carries a
// page.{number,total_pages} object; the loop advances ?page=N until total_pages
// is reached or the bound maxPages is hit.
func (c Connector) harvestPaged(ctx context.Context, r *connsdk.Requester, spec streamSpec, base url.Values, maxPages int, emit func(connectors.Record) error) error {
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, spec.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read nasa %s page %d: %w", spec.resource, page, err)
		}
		if err := emitRecords(ctx, resp.Body, spec, emit); err != nil {
			return err
		}
		totalPages, err := connsdk.StringAt(resp.Body, "page.total_pages")
		if err != nil {
			return fmt.Errorf("decode nasa %s page meta: %w", spec.resource, err)
		}
		total, convErr := strconv.Atoi(strings.TrimSpace(totalPages))
		if convErr != nil || total <= 0 {
			// No usable pagination metadata: treat the response as the only page.
			return nil
		}
		if page+1 >= total {
			return nil
		}
	}
	return nil
}

func emitRecords(ctx context.Context, body []byte, spec streamSpec, emit func(connectors.Record) error) error {
	records, err := connsdk.RecordsAt(body, spec.arrayPath)
	if err != nil {
		return fmt.Errorf("decode nasa %s: %w", spec.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(spec.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// streamQuery builds the per-stream request query from config. NASA endpoints
// take a handful of optional date/sol filters; the api_key is added separately by
// the authenticator so it never appears here.
func streamQuery(stream string, cfg connectors.RuntimeConfig) url.Values {
	query := url.Values{}
	conf := cfg.Config
	switch stream {
	case "apod":
		if v := strings.TrimSpace(conf["start_date"]); v != "" {
			query.Set("start_date", v)
		}
		if v := strings.TrimSpace(conf["end_date"]); v != "" {
			query.Set("end_date", v)
		}
		if v := strings.TrimSpace(conf["count"]); v != "" {
			query.Set("count", v)
		}
		if boolConfig(conf["thumbs"]) {
			query.Set("thumbs", "true")
		}
	case "neo_feed":
		if v := strings.TrimSpace(conf["start_date"]); v != "" {
			query.Set("start_date", v)
		}
		if v := strings.TrimSpace(conf["end_date"]); v != "" {
			query.Set("end_date", v)
		}
	case "mars_photos":
		sol := strings.TrimSpace(conf["sol"])
		if sol == "" {
			sol = "1000"
		}
		query.Set("sol", sol)
	}
	return query
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise nasa credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, spec streamSpec, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := fixtureItem(stream, i)
		if err := emit(spec.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func fixtureItem(stream string, i int) map[string]any {
	switch stream {
	case "apod":
		return map[string]any{
			"date":            fmt.Sprintf("2026-01-0%d", i),
			"title":           fmt.Sprintf("Fixture Sky %d", i),
			"explanation":     "Deterministic fixture record; no network call.",
			"media_type":      "image",
			"url":             fmt.Sprintf("https://apod.example/%d.jpg", i),
			"service_version": "v1",
		}
	case "epic":
		return map[string]any{
			"identifier": fmt.Sprintf("2026010%d000000", i),
			"caption":    "Fixture Earth view",
			"image":      fmt.Sprintf("epic_1b_2026010%d", i),
			"version":    "03",
			"date":       fmt.Sprintf("2026-01-0%d 00:00:00", i),
		}
	case "mars_photos":
		return map[string]any{
			"id":         fmt.Sprintf("mars_fixture_%d", i),
			"sol":        int64(1000 + i),
			"img_src":    fmt.Sprintf("https://mars.example/%d.jpg", i),
			"earth_date": fmt.Sprintf("2026-01-0%d", i),
			"camera":     map[string]any{"name": "FHAZ", "full_name": "Front Hazard Avoidance Camera"},
			"rover":      map[string]any{"name": "Curiosity"},
		}
	default: // neo_feed, neo_browse
		return map[string]any{
			"id":                                fmt.Sprintf("200043%d", i),
			"neo_reference_id":                  fmt.Sprintf("200043%d", i),
			"name":                              fmt.Sprintf("%d Fixteros", i),
			"nasa_jpl_url":                      "https://ssd.jpl.nasa.gov/fixture",
			"absolute_magnitude_h":              10.0 + float64(i),
			"is_potentially_hazardous_asteroid": false,
			"is_sentry_object":                  false,
		}
	}
}

// requester builds a connsdk.Requester wired with api_key query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := nasaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := nasaSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("nasa connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("api_key", secret),
		UserAgent: nasaUserAgent,
	}, nil
}

func nasaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// nasaBaseURL resolves and validates the base URL. The default is api.nasa.gov;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func nasaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return nasaDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("nasa config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("nasa config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("nasa config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func nasaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" {
		return nasaDefaultMaxPages, nil
	}
	if raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("nasa config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("nasa config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func boolConfig(raw string) bool {
	return strings.EqualFold(strings.TrimSpace(raw), "true")
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
