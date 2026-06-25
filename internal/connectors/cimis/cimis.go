// Package cimis implements the native pm CIMIS (California Irrigation Management
// Information System) connector. It is a declarative-HTTP per-system connector
// modeled on the stripe reference: a thin package that composes the connsdk
// toolkit (Requester + APIKeyQuery auth + JSON extraction) with CIMIS-specific
// stream definitions and the /api/data + /api/station endpoints.
//
// The CIMIS Web API (https://et.water.ca.gov/api/data) is a single-request,
// date-range-bounded weather/ET API: there is no pagination, the requested
// date range bounds the result set. Records arrive nested under
// Data.Providers[].Records[] and are flattened across providers in code.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package cimis

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	cimisDefaultBaseURL = "https://et.water.ca.gov"
	cimisUserAgent      = "polymetrics-go-cli"
)

// defaultDailyDataItems mirrors the CIMIS API default when no daily items are
// configured (the 14 standard daily items, comma-joined).
const defaultDailyDataItems = "day-air-tmp-avg,day-air-tmp-max,day-air-tmp-min,day-dew-pnt,day-eto,day-asce-eto,day-precip,day-rel-hum-avg,day-rel-hum-max,day-rel-hum-min,day-soil-tmp-avg,day-sol-rad-avg,day-vap-pres-avg,day-wind-spd-avg"

func init() {
	connectors.RegisterFactory("cimis", New)
}

// New returns the CIMIS connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm CIMIS connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "cimis" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "cimis",
		DisplayName:     "CIMIS",
		IntegrationType: "api",
		Description:     "Reads California Irrigation Management Information System (CIMIS) daily and hourly weather/ET observations and station metadata through the CIMIS Web API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

// Check verifies the connector is configured well enough to talk to CIMIS. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := cimisBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cimisSecret(cfg)) == "" {
		return errors.New("cimis connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// The station endpoint is the lightest read that confirms connectivity; it
	// does not require the appKey but exercises the same base URL and client.
	if err := r.DoJSON(ctx, http.MethodGet, "api/station", nil, nil, nil); err != nil {
		return fmt.Errorf("check cimis: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: cimisStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a CIMIS stream starts with an
// empty incremental cursor (full sync over the configured date range).
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
		stream = "daily"
	}
	def, ok := cimisStreamDefs[stream]
	if !ok {
		return fmt.Errorf("cimis stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, def, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	query, err := readQuery(stream, def, req.Config)
	if err != nil {
		return err
	}

	resp, err := r.Do(ctx, http.MethodGet, def.resource, query, nil)
	if err != nil {
		return fmt.Errorf("read cimis %s: %w", stream, err)
	}

	var records []map[string]any
	if def.providersPath {
		records, err = flattenProviderRecords(resp.Body)
	} else {
		records, err = connsdk.RecordsAt(resp.Body, def.recordsPath)
	}
	if err != nil {
		return fmt.Errorf("decode cimis %s: %w", stream, err)
	}

	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(def.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// readQuery builds the /api/data query (or empty for /api/station). It validates
// the required targets/date-range config for data streams.
func readQuery(stream string, def streamDef, cfg connectors.RuntimeConfig) (url.Values, error) {
	if !def.providersPath {
		// Station endpoint takes no query parameters.
		return url.Values{}, nil
	}

	targets := strings.TrimSpace(cfg.Config["targets"])
	if targets == "" {
		return nil, errors.New("cimis config targets is required for daily/hourly streams")
	}
	start := strings.TrimSpace(cfg.Config["start_date"])
	end := strings.TrimSpace(cfg.Config["end_date"])
	if start == "" || end == "" {
		return nil, errors.New("cimis config start_date and end_date are required for daily/hourly streams")
	}

	q := url.Values{}
	q.Set("targets", targets)
	q.Set("startDate", normalizeDate(start))
	q.Set("endDate", normalizeDate(end))
	q.Set("dataItems", dataItemsFor(def.scope, cfg))
	if unit := strings.TrimSpace(cfg.Config["unit_of_measure"]); unit != "" {
		q.Set("unitOfMeasure", unit)
	}
	if scs := strings.TrimSpace(cfg.Config["prioritize_scs"]); scs != "" {
		q.Set("prioritizeSCS", scs)
	}
	return q, nil
}

// dataItemsFor resolves the comma-delimited dataItems for a scope from config,
// falling back to the daily defaults when unset for the daily stream.
func dataItemsFor(s scope, cfg connectors.RuntimeConfig) string {
	switch s {
	case scopeHourly:
		if items := strings.TrimSpace(cfg.Config["hourly_data_items"]); items != "" {
			return normalizeList(items)
		}
		return "hly-eto,hly-air-tmp,hly-rel-hum,hly-sol-rad,hly-wind-spd"
	default:
		if items := strings.TrimSpace(cfg.Config["daily_data_items"]); items != "" {
			return normalizeList(items)
		}
		return defaultDailyDataItems
	}
}

// normalizeList trims whitespace around comma-separated items (configs may arrive
// as JSON arrays serialized loosely, e.g. "day-eto, day-precip").
func normalizeList(in string) string {
	parts := strings.Split(in, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, ",")
}

// normalizeDate accepts an RFC3339 timestamp (the catalog config format) or a
// bare yyyy-mm-dd date and returns the yyyy-mm-dd form the CIMIS API expects.
func normalizeDate(in string) string {
	in = strings.TrimSpace(in)
	if idx := strings.IndexByte(in, 'T'); idx > 0 {
		return in[:idx]
	}
	return in
}

// flattenProviderRecords walks Data.Providers[].Records[] and concatenates the
// records across every provider. connsdk.RecordsAt only descends one array
// level, so this small loop handles the nested CIMIS shape.
func flattenProviderRecords(body []byte) ([]map[string]any, error) {
	providers, err := connsdk.RecordsAt(body, "Data.Providers")
	if err != nil {
		return nil, err
	}
	var out []map[string]any
	for _, provider := range providers {
		raw, ok := provider["Records"]
		if !ok {
			continue
		}
		items, ok := raw.([]any)
		if !ok {
			continue
		}
		for _, item := range items {
			if obj, ok := item.(map[string]any); ok {
				out = append(out, obj)
			}
		}
	}
	return out, nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise cimis credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, def streamDef, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var item map[string]any
		if def.providersPath {
			item = map[string]any{
				"Date":     fmt.Sprintf("2026-01-0%d", i),
				"Julian":   fmt.Sprintf("%d", i),
				"Station":  "2",
				"Standard": "english",
				"ZipCodes": []any{"95823"},
				"Scope":    string(def.scope),
				"DayAirTmpAvg": map[string]any{
					"Value": fmt.Sprintf("5%d.0", i),
					"Qc":    "",
					"Unit":  "(F)",
				},
			}
			if def.scope == scopeHourly {
				item["Hour"] = fmt.Sprintf("%02d00", i)
			}
		} else {
			item = map[string]any{
				"StationNbr": fmt.Sprintf("%d", i),
				"Name":       fmt.Sprintf("Fixture Station %d", i),
				"City":       "Sacramento",
				"County":     "Sacramento",
				"IsActive":   "True",
			}
		}
		record := def.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// Write is unsupported: CIMIS is a read-only public data API.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// requester builds a connsdk.Requester wired with appKey query-param auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := cimisBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := cimisSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("cimis connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("appKey", secret),
		UserAgent: cimisUserAgent,
	}, nil
}

func cimisSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// cimisBaseURL resolves and validates the base URL. The default is
// et.water.ca.gov; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func cimisBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return cimisDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("cimis config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("cimis config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("cimis config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
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
