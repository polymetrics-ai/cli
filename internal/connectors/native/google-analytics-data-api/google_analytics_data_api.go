// Package googleanalyticsdataapi implements the native pm Google Analytics 4
// (GA4) Data API connector. It is a declarative-HTTP per-system connector built
// on the same shape as the stripe reference: a thin package that composes the
// connsdk toolkit (Requester + Bearer/OAuth2 access-token auth + JSON extraction)
// with GA4-specific report definitions, the runReport endpoint, and fixed
// metadata/audience-export direct reads.
//
// GA4 stream reads are POST runReport calls that take a dimension x metric query
// and return rows. Each published "stream" is a canned report spec (see
// streams.go); a row is flattened to a record by projecting
// dimensionHeaders/metricHeaders onto the row's dimensionValues/metricValues.
// Bounded GET direct reads cover property metadata and audience-export metadata
// without exposing arbitrary method, URL, or body input.
//
// The connector is read-only: the GA4 Data API surface exposed here has no safe
// reverse-ETL writes, so Capabilities.Write is false.
package googleanalyticsdataapi

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
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	connectorName        = "google-analytics-data-api"
	gaDefaultBaseURL     = "https://analyticsdata.googleapis.com"
	gaAPIVersion         = "v1beta"
	gaDefaultPageSize    = 10000
	gaMaxPageSize        = 250000
	gaUserAgent          = "polymetrics-go-cli"
	gaDefaultStartDate   = "30daysAgo"
	gaDefaultEndDate     = "today"
	gaFixturePropertyID  = "000000000"
	gaFixtureRecordCount = 2
)

// New returns the GA4 Data API connector.
func New() *Connector {
	bundle := mustLoadBundle()
	return &Connector{Base: engine.NewBase(bundle), bundle: bundle}
}

// Connector is the native pm Google Analytics 4 Data API connector.
type Connector struct {
	engine.Base
	bundle engine.Bundle

	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func mustLoadBundle() engine.Bundle {
	bundle, err := engine.Load(defs.FS, connectorName)
	if err != nil {
		panic("native/" + connectorName + ": failed to load defs/" + connectorName + " bundle: " + err.Error())
	}
	return bundle
}

func (c *Connector) bundleOrLoad() engine.Bundle {
	if c != nil && c.bundle.Name != "" {
		return c.bundle
	}
	return mustLoadBundle()
}

func (c *Connector) Name() string { return connectorName }

func (c *Connector) Metadata() connectors.Metadata {
	if c != nil && c.bundle.Name != "" {
		return c.Base.Metadata()
	}
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Google Analytics 4 (GA4)",
		IntegrationType: "api",
		Description:     "Reads Google Analytics 4 reports and bounded metadata resources from the Analytics Data API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Manifest exposes the native connector's authored bundle shape to generated
// help, docs, and skills while stream reads and operation direct reads remain
// implemented by the native code above. engine.Base intentionally provides
// Definition()/CommandSurface() only, so this method keeps the GA-specific guide
// truthful without changing every other native connector.
func (c *Connector) Manifest() connectors.Manifest {
	return connectors.Manifest{
		Metadata: c.Metadata(),
		ConfigFields: []connectors.ConfigField{
			{Name: "property_ids", Description: "Comma, space, or newline separated GA4 numeric property IDs; native reads use the first property ID per read call.", Required: true},
			{Name: "property_id", Description: "Optional single GA4 numeric property ID for direct metadata/audience-export commands; defaults to the first property_ids value."},
			{Name: "audience_export_id", Description: "Audience export ID used by the get audience export direct command."},
			{Name: "base_url", Description: "Analytics Data API base URL override for local fixture tests only.", Default: gaDefaultBaseURL},
			{Name: "date_ranges_start_date", Description: "GA4 report start date, either YYYY-MM-DD or a GA4 relative token such as 30daysAgo.", Default: gaDefaultStartDate},
			{Name: "date_ranges_end_date", Description: "GA4 report end date, either YYYY-MM-DD or a GA4 relative token such as today or yesterday.", Default: gaDefaultEndDate},
			{Name: "page_size", Description: "Native runReport page size; must be between 1 and 250000.", Default: strconv.Itoa(gaDefaultPageSize)},
			{Name: "max_pages", Description: "Native runReport page cap. Use a positive integer, 0, all, or unlimited for unbounded reads."},
			{Name: "mode", Description: "Set to fixture for credential-free connector-owned tests; do not use for live provider validation."},
			{Name: "keep_empty_rows", Description: "Legacy compatibility flag retained for credential compatibility; native preset reads currently send false."},
			{Name: "convert_conversions_event", Description: "Legacy compatibility flag retained for credential compatibility."},
			{Name: "custom_reports_array", Description: "Legacy custom report JSON string. Custom reports remain outside this documented parity slice."},
			{Name: "lookback_window", Description: "Legacy lookback-window setting retained for credential compatibility."},
			{Name: "subscription_tier", Description: "Informational GA4 property tier for quota planning."},
			{Name: "window_in_days", Description: "Legacy window setting retained for credential compatibility."},
		},
		SecretFields: []connectors.SecretField{
			{Name: "access_token", Description: "OAuth2 bearer access token with Analytics Data API read access; prefer --from-env or --value-stdin."},
			{Name: "credentials", Description: "Legacy flattened bearer token payload; prefer access_token for new credentials."},
		},
		AuthModes: []connectors.AuthModeSpec{
			{
				Name:         "oauth2_bearer",
				Description:  "OAuth2 bearer token with Google Analytics Data API read access.",
				ConfigFields: []string{"property_ids"},
				SecretFields: []string{"access_token", "credentials"},
				Read:         true,
				Write:        false,
			},
		},
		Streams:         gaStreams(),
		SyncModes:       []string{"full_refresh_append", "full_refresh_overwrite", "full_refresh_overwrite_deduped", "incremental_append", "incremental_append_deduped"},
		SourceSyncModes: []string{"full_refresh", "incremental"},
		Pagination: connectors.PaginationSpec{
			Type:           "offset_limit",
			PageSizeField:  "page_size",
			PageLimitField: "max_pages",
			DefaultLimit:   strconv.Itoa(gaDefaultPageSize),
		},
		Risk: connectors.RiskSpec{
			Read:     "external Google Analytics Data API reads for configured properties; direct reads are fixed-target, bounded, and JSON-redacted",
			Write:    "unsupported",
			Mutation: "none",
			Approval: "none for read-only operations; future audience-export creation would require plan, preview, explicit approval, and execute before being advertised",
		},
	}
}

// Check verifies the connector is configured well enough to talk to the GA4 Data
// API. In fixture mode it short-circuits without a network call.
func (c *Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := gaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(gaSecret(cfg)) == "" {
		return errors.New("google-analytics-data-api connector requires an OAuth2 bearer access token (secret access_token or credentials)")
	}
	property, err := gaPropertyID(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded one-row runReport confirms auth, the property id, and
	// connectivity without mutating anything.
	spec := gaReports["daily_active_users"]
	body := buildReportBody(spec, cfg, 0, 1)
	path := reportPath(property)
	if err := r.DoJSON(ctx, http.MethodPost, path, nil, body, nil); err != nil {
		return fmt.Errorf("check google-analytics-data-api: %w", err)
	}
	return nil
}

func (c *Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: gaStreams()}, nil
}

// Write satisfies the connectors.Connector interface. The GA4 Data API is a
// reporting (read) API with no safe reverse-ETL writes, so writes are
// unsupported and Capabilities.Write is false.
func (c *Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a GA4 stream starts with an
// empty cursor; the supported_sync_modes are full_refresh, but date-dimensioned
// reports can carry a "date" cursor the start_date config raises at read time.
func (c *Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

func (c *Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "daily_active_users"
	}
	spec, ok := gaReports[stream]
	if !ok {
		return fmt.Errorf("google-analytics-data-api stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, spec, req, emit)
	}

	property, err := gaPropertyID(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := gaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := gaMaxPages(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, spec, property, req.Config, pageSize, maxPages, emit)
}

var implementedDirectOperations = map[string]bool{
	"google-analytics-data-api.get_metadata":          true,
	"google-analytics-data-api.list_audience_exports": true,
	"google-analytics-data-api.get_audience_export":   true,
}

// OperationDirectRead executes the connector's bounded direct-read operations.
// Fixture mode returns sanitized local responses; live mode delegates to the
// declarative operation executor for the reviewed fixed endpoint definitions.
func (c *Connector) OperationDirectRead(ctx context.Context, req connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.DirectReadResult{}, err
	}
	if !implementedDirectOperations[req.Operation] {
		return connectors.DirectReadResult{}, fmt.Errorf("google-analytics-data-api operation %q is not executable in this connector slice: %w", req.Operation, connectors.ErrUnsupportedOperation)
	}
	var err error
	req, err = normalizeOperationDirectReadRequest(req)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	if fixtureMode(req.Config) {
		return operationFixture(ctx, req)
	}
	return engine.OperationDirectRead(ctx, c.bundleOrLoad(), req, nil)
}

func normalizeOperationDirectReadRequest(req connectors.OperationDirectReadRequest) (connectors.OperationDirectReadRequest, error) {
	req.PathParams = cloneStringMap(req.PathParams)
	req.Config.Config = cloneStringMap(req.Config.Config)
	req.Config.Secrets = cloneStringMap(req.Config.Secrets)
	if req.PathParams == nil {
		req.PathParams = map[string]string{}
	}
	if req.Config.Config == nil {
		req.Config.Config = map[string]string{}
	}
	if token := gaSecret(req.Config); strings.TrimSpace(token) != "" {
		req.Config.Secrets["access_token"] = token
	}
	propertyID, ok, err := gaOperationDirectReadPropertyID(req)
	if err != nil {
		return req, err
	}
	if ok {
		req.PathParams["property_id"] = propertyID
	}
	if req.PathParams["audience_export_id"] == "" {
		if id := firstNonEmpty(req.Config.Config["audience_export_id"], req.Config.Config["audienceExportsId"]); id != "" {
			req.PathParams["audience_export_id"] = strings.TrimPrefix(id, "audienceExports/")
		}
	}
	return req, nil
}

func gaOperationDirectReadPropertyID(req connectors.OperationDirectReadRequest) (string, bool, error) {
	if raw := strings.TrimSpace(req.PathParams["property_id"]); raw != "" {
		property, err := normalizeGAPropertyID(raw, "path property_id")
		if err != nil {
			return "", false, err
		}
		return property, true, nil
	}
	if raw := strings.TrimSpace(req.Config.Config["property_id"]); raw != "" {
		property, err := normalizeGAPropertyID(raw, "config property_id")
		if err != nil {
			return "", false, err
		}
		return property, true, nil
	}
	if raw := strings.TrimSpace(req.Config.Config["property_ids"]); raw != "" {
		property, err := normalizeGAPropertyID(raw, "config property_ids")
		if err != nil {
			return "", false, err
		}
		return property, true, nil
	}
	return "", false, nil
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func operationFixture(ctx context.Context, req connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.DirectReadResult{}, err
	}
	property := firstNonEmpty(req.PathParams["property_id"], gaFixturePropertyID)
	audienceExportID := firstNonEmpty(req.PathParams["audience_export_id"], "audience_export_fixture_1")
	switch req.Operation {
	case "google-analytics-data-api.get_metadata":
		return connectors.DirectReadResult{
			Connector: connectorName,
			Method:    http.MethodGet,
			Path:      fmt.Sprintf("v1beta/properties/%s/metadata", property),
			Status:    http.StatusOK,
			Body: map[string]any{
				"name":       fmt.Sprintf("properties/%s/metadata", property),
				"dimensions": []any{map[string]any{"apiName": "date", "uiName": "Date"}},
				"metrics":    []any{map[string]any{"apiName": "activeUsers", "uiName": "Active users", "type": "TYPE_INTEGER"}},
			},
		}, nil
	case "google-analytics-data-api.list_audience_exports":
		return connectors.DirectReadResult{
			Connector: connectorName,
			Method:    http.MethodGet,
			Path:      fmt.Sprintf("v1beta/properties/%s/audienceExports", property),
			Status:    http.StatusOK,
			Body: map[string]any{
				"audienceExports": []any{map[string]any{
					"name":        fmt.Sprintf("properties/%s/audienceExports/%s", property, audienceExportID),
					"displayName": "Fixture audience export",
					"state":       "ACTIVE",
				}},
			},
		}, nil
	case "google-analytics-data-api.get_audience_export":
		return connectors.DirectReadResult{
			Connector: connectorName,
			Method:    http.MethodGet,
			Path:      fmt.Sprintf("v1beta/properties/%s/audienceExports/%s", property, audienceExportID),
			Status:    http.StatusOK,
			Body: map[string]any{
				"name":        fmt.Sprintf("properties/%s/audienceExports/%s", property, audienceExportID),
				"displayName": "Fixture audience export",
				"state":       "ACTIVE",
				"rowCount":    1,
			},
		}, nil
	default:
		return connectors.DirectReadResult{}, fmt.Errorf("google-analytics-data-api fixture operation %q: %w", req.Operation, connectors.ErrUnsupportedOperation)
	}
}

// harvest drives GA4 offset/limit pagination. runReport returns
// {dimensionHeaders, metricHeaders, rows, rowCount}; each page is requested with
// offset advanced by limit until offset >= rowCount (or rows run out). The loop
// lives here because the offset paginator is body-driven and report-specific.
func (c *Connector) harvest(ctx context.Context, r *connsdk.Requester, spec reportSpec, property string, cfg connectors.RuntimeConfig, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := reportPath(property)
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		body := buildReportBody(spec, cfg, offset, pageSize)
		resp, err := r.Do(ctx, http.MethodPost, path, nil, body)
		if err != nil {
			return fmt.Errorf("read google-analytics-data-api %s: %w", spec.name, err)
		}
		report, err := decodeReport(resp.Body)
		if err != nil {
			return fmt.Errorf("decode google-analytics-data-api %s page: %w", spec.name, err)
		}
		for _, row := range report.Rows {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRow(property, report, row)); err != nil {
				return err
			}
		}
		emitted := offset + len(report.Rows)
		// Stop when we've consumed all reported rows or the page came back
		// short (defensive against a missing/zero rowCount).
		if len(report.Rows) == 0 || (report.RowCount > 0 && emitted >= report.RowCount) || len(report.Rows) < pageSize {
			return nil
		}
		offset = emitted
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free (mirrors
// stripe's fixture intent).
func (c *Connector) readFixture(ctx context.Context, spec reportSpec, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 0; i < gaFixtureRecordCount; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		record := connectors.Record{
			"property_id": gaFixturePropertyID,
			"connector":   connectorName,
			"fixture":     true,
		}
		for di, dim := range spec.dimensions {
			if dim == "date" {
				record[dim] = fmt.Sprintf("2026010%d", i+1)
				continue
			}
			record[dim] = fmt.Sprintf("%s_fixture_%d", dim, i+1)
			_ = di
		}
		for mi, metric := range spec.metrics {
			record[metric] = strconv.Itoa((i + 1) * (mi + 1) * 10)
		}
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth (the GA4 OAuth2
// access token), the resolved base URL, and a JSON content type. The secret only
// ever flows into connsdk.Bearer; it is never logged.
func (c *Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := gaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := gaSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("google-analytics-data-api connector requires an OAuth2 bearer access token (secret access_token or credentials)")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: gaUserAgent,
	}, nil
}

// reportPath builds the runReport endpoint path for a property id, e.g.
// "v1beta/properties/123456:runReport".
func reportPath(property string) string {
	return fmt.Sprintf("%s/properties/%s:runReport", gaAPIVersion, property)
}

// buildReportBody constructs the runReport request body for a report spec at the
// given offset/limit, applying the configured date range.
func buildReportBody(spec reportSpec, cfg connectors.RuntimeConfig, offset, limit int) map[string]any {
	dims := make([]map[string]string, 0, len(spec.dimensions))
	for _, d := range spec.dimensions {
		dims = append(dims, map[string]string{"name": d})
	}
	metrics := make([]map[string]string, 0, len(spec.metrics))
	for _, m := range spec.metrics {
		metrics = append(metrics, map[string]string{"name": m})
	}
	start, end := gaDateRange(cfg)
	return map[string]any{
		"dimensions":    dims,
		"metrics":       metrics,
		"dateRanges":    []map[string]string{{"startDate": start, "endDate": end}},
		"offset":        offset,
		"limit":         limit,
		"keepEmptyRows": false,
	}
}

func gaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	// The catalog declares the secret under nested credentials.* keys; accept the
	// access_token under both the dotted and bare forms (and a couple of common
	// aliases) so callers can flatten the vault however they like.
	for _, key := range []string{
		"credentials.access_token",
		"access_token",
		"credentials",
		"credentials.api_key",
		"api_key",
	} {
		if v := strings.TrimSpace(cfg.Secrets[key]); v != "" {
			return v
		}
	}
	return ""
}

// gaPropertyID resolves the single GA4 property id to read. The catalog's
// property_ids config is a list; this connector reads the first id (callers can
// run one stream-read per property). The id may be given bare ("123456") or
// prefixed ("properties/123456").
func gaPropertyID(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(firstNonEmpty(cfg.Config["property_ids"], cfg.Config["property_id"]))
	if raw == "" {
		return "", errors.New("google-analytics-data-api connector requires config property_ids")
	}
	return normalizeGAPropertyID(raw, "config property_ids")
}

func normalizeGAPropertyID(raw, field string) (string, error) {
	raw = strings.TrimSpace(raw)
	for _, sep := range []string{",", " ", "\n"} {
		if i := strings.IndexAny(raw, sep); i >= 0 {
			raw = raw[:i]
			break
		}
	}
	raw = strings.TrimSpace(strings.TrimPrefix(raw, "properties/"))
	if raw == "" {
		return "", fmt.Errorf("google-analytics-data-api %s is empty", field)
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("google-analytics-data-api property id must be numeric, got %q", raw)
		}
	}
	return raw, nil
}

// gaDateRange resolves the report date range from config, defaulting to the last
// 30 days. GA4 accepts YYYY-MM-DD or relative tokens (NdaysAgo, today,
// yesterday), which are passed through unchanged.
func gaDateRange(cfg connectors.RuntimeConfig) (string, string) {
	start := strings.TrimSpace(firstNonEmpty(cfg.Config["date_ranges_start_date"], cfg.Config["start_date"]))
	if start == "" {
		start = gaDefaultStartDate
	}
	end := strings.TrimSpace(firstNonEmpty(cfg.Config["date_ranges_end_date"], cfg.Config["end_date"]))
	if end == "" {
		end = gaDefaultEndDate
	}
	return start, end
}

// gaBaseURL resolves and validates the base URL. The default is
// analyticsdata.googleapis.com; any override must be an absolute https (or http
// for local test servers) URL with a host to bound SSRF risk.
func gaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return gaDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("google-analytics-data-api config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("google-analytics-data-api config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("google-analytics-data-api config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func gaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return gaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("google-analytics-data-api config page_size must be an integer: %w", err)
	}
	if value < 1 || value > gaMaxPageSize {
		return 0, fmt.Errorf("google-analytics-data-api config page_size must be between 1 and %d", gaMaxPageSize)
	}
	return value, nil
}

func gaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("google-analytics-data-api config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("google-analytics-data-api config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
