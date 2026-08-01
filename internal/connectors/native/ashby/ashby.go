// Package ashby implements the native pm Ashby connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference connector: a thin package that composes the connsdk toolkit
// (Requester + Basic auth + RecordsAt extraction) with Ashby-specific stream
// definitions, endpoints, and its POST cursor-in-body pagination.
//
// Ashby is an applicant-tracking system. Its REST API lives at
// https://api.ashbyhq.com; list-style reads are POST requests authenticated
// with HTTP Basic auth (the API key is the username, the password is blank) and
// commonly return {success, results:[...], moreDataAvailable, nextCursor}. The
// native package owns Ashby's POST cursor-in-body reads while the generated
// bundle owns fixed direct-read and reverse-ETL write definitions.
package ashby

import (
	"context"
	"encoding/json"
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
	ashbyDefaultBaseURL  = "https://api.ashbyhq.com"
	ashbyDefaultPageSize = 100
	ashbyMaxPageSize     = 100
	ashbyUserAgent       = "polymetrics-go-cli"
	// ashbyAPIVersion is sent in the Accept header per Ashby's docs
	// (Accept: application/json; version=1).
	ashbyAccept = "application/json; version=1"
)

// New returns the Ashby connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Ashby connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "ashby" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "ashby",
		DisplayName:     "Ashby",
		IntegrationType: "api",
		Description:     "Reads Ashby applicant-tracking data and exposes typed, gated Ashby reverse-ETL writes through the documented REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: true},
	}
}

// Check verifies the connector is configured well enough to talk to Ashby. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := ashbyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(ashbySecret(cfg)) == "" {
		return errors.New("ashby connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded apiKey.info request confirms auth and connectivity without
	// mutating anything.
	if _, err := r.Do(ctx, http.MethodPost, "apiKey.info", nil, map[string]any{}); err != nil {
		return fmt.Errorf("check ashby: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: ashbyStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "candidates"
	}
	endpoint, ok := ashbyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("ashby stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := ashbyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := ashbyMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, req, pageSize, maxPages, emit)
}

// harvest drives Ashby's POST cursor-in-body pagination. Each stream has a
// fixed path and an allow-list of documented request body fields generated from
// Ashby's OpenAPI. Caller input can populate only those fields; cursor, limit,
// page count, and client-side incremental filtering stay bounded here.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, req connectors.ReadRequest, pageSize, maxPages int, emit func(connectors.Record) error) error {
	cursor := strings.TrimSpace(req.State["cursor"])
	lowerBound := cursor
	if lowerBound == "" {
		lowerBound = strings.TrimSpace(req.Config.Config["start_date"])
	}
	baseBody, err := ashbyStreamBody(endpoint, req.Config, req.Query, pageSize)
	if err != nil {
		return err
	}
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		body := cloneMap(baseBody)
		if cursor != "" {
			body["cursor"] = cursor
		}
		resp, err := r.Do(ctx, http.MethodPost, endpoint.path, nil, body)
		if err != nil {
			return fmt.Errorf("read ashby %s: %w", endpoint.path, err)
		}
		pageBody, records, err := ashbyResultRecords(resp.Body)
		if err != nil {
			return fmt.Errorf("decode ashby %s page: %w", endpoint.path, err)
		}
		for i, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			record := ashbyProjectRecord(endpoint, item, page, i)
			if !ashbyPassesCursor(record, endpoint.cursorField, lowerBound) {
				continue
			}
			if err := emit(record); err != nil {
				return err
			}
		}
		next := strings.TrimSpace(stringValue(pageBody["nextCursor"]))
		if !boolValue(pageBody["moreDataAvailable"]) || next == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise Ashby credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		record := ashbyProjectRecord(endpoint, ashbyFixtureItem(stream, endpoint, i), 0, i-1)
		record["connector"] = "ashby"
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

func ashbyStreamBody(endpoint streamEndpoint, cfg connectors.RuntimeConfig, query map[string]string, pageSize int) (map[string]any, error) {
	body := map[string]any{"limit": pageSize}
	for field, kind := range endpoint.requestFields {
		if field == "limit" || field == "cursor" {
			continue
		}
		raw := strings.TrimSpace(query[field])
		if raw == "" {
			raw = strings.TrimSpace(cfg.Config[field])
		}
		if raw == "" {
			continue
		}
		value, err := ashbyCoerceRequestValue(field, kind, raw)
		if err != nil {
			return nil, err
		}
		body[field] = value
	}
	for _, field := range endpoint.requiredFields {
		if _, ok := body[field]; !ok {
			return nil, fmt.Errorf("ashby stream %s requires documented request field %q", endpoint.path, field)
		}
	}
	return body, nil
}

func ashbyCoerceRequestValue(field, kind, raw string) (any, error) {
	switch kind {
	case "integer":
		value, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("ashby request field %s must be an integer: %w", field, err)
		}
		return value, nil
	case "number":
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("ashby request field %s must be a number: %w", field, err)
		}
		return value, nil
	case "boolean":
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("ashby request field %s must be a boolean: %w", field, err)
		}
		return value, nil
	case "array":
		if strings.HasPrefix(raw, "[") {
			var out []any
			if err := json.Unmarshal([]byte(raw), &out); err != nil {
				return nil, fmt.Errorf("ashby request field %s must be a JSON array: %w", field, err)
			}
			return out, nil
		}
		items := strings.Split(raw, ",")
		out := make([]string, 0, len(items))
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item != "" {
				out = append(out, item)
			}
		}
		return out, nil
	case "object":
		var out map[string]any
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			return nil, fmt.Errorf("ashby request field %s must be a JSON object: %w", field, err)
		}
		return out, nil
	default:
		return raw, nil
	}
}

func ashbyResultRecords(body []byte) (map[string]any, []map[string]any, error) {
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, nil, err
	}
	results, ok := decoded["results"]
	if !ok || results == nil {
		return decoded, []map[string]any{decoded}, nil
	}
	switch typed := results.(type) {
	case []any:
		records := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if rec, ok := item.(map[string]any); ok {
				records = append(records, rec)
			} else {
				records = append(records, map[string]any{"value": item})
			}
		}
		return decoded, records, nil
	case map[string]any:
		return decoded, []map[string]any{typed}, nil
	default:
		return decoded, []map[string]any{{"value": typed}}, nil
	}
}

func ashbyProjectRecord(endpoint streamEndpoint, item map[string]any, page, index int) connectors.Record {
	record := connectors.Record{}
	if len(endpoint.fields) == 0 {
		for key, value := range item {
			record[key] = value
		}
	} else {
		for _, field := range endpoint.fields {
			if value, ok := item[field.Name]; ok {
				record[field.Name] = value
			}
		}
	}
	for _, field := range endpoint.syntheticFields {
		if _, ok := record[field]; !ok {
			record[field] = fmt.Sprintf("ashby_%s_%d_%d", strings.Trim(endpoint.path, "/"), page+1, index+1)
		}
	}
	return record
}

func ashbyFixtureItem(stream string, endpoint streamEndpoint, i int) map[string]any {
	item := map[string]any{}
	for _, field := range endpoint.fields {
		item[field.Name] = ashbyFixtureValue(stream, field, i)
	}
	for _, key := range endpoint.primaryKey {
		if _, ok := item[key]; !ok || item[key] == nil || item[key] == "" {
			item[key] = fmt.Sprintf("%s_fixture_%d", snakeish(stream), i)
		}
	}
	if endpoint.cursorField != "" {
		item[endpoint.cursorField] = fmt.Sprintf("2026-01-%02dT00:00:00Z", i)
	}
	return item
}

func ashbyFixtureValue(stream string, field connectors.Field, i int) any {
	lower := strings.ToLower(field.Name)
	switch field.Type {
	case "boolean":
		return i%2 == 0
	case "integer":
		return i
	case "number":
		return float64(i)
	case "array":
		return []any{fmt.Sprintf("%s_fixture_%d", lower, i)}
	case "object":
		return map[string]any{"id": fmt.Sprintf("%s_%s_fixture_%d", snakeish(stream), lower, i)}
	default:
		if lower == "id" || strings.HasSuffix(lower, "id") {
			return fmt.Sprintf("%s_%s_fixture_%d", snakeish(stream), lower, i)
		}
		if strings.Contains(lower, "email") {
			return fmt.Sprintf("fixture+%d@example.invalid", i)
		}
		if strings.Contains(lower, "url") {
			return fmt.Sprintf("https://example.invalid/%s/%d", snakeish(stream), i)
		}
		if strings.Contains(lower, "date") || strings.HasSuffix(lower, "at") || strings.Contains(lower, "time") {
			return fmt.Sprintf("2026-01-%02dT00:00:00Z", i)
		}
		return fmt.Sprintf("%s fixture %d", field.Name, i)
	}
}

func ashbyPassesCursor(record connectors.Record, cursorField, lowerBound string) bool {
	if cursorField == "" || lowerBound == "" {
		return true
	}
	value := strings.TrimSpace(stringValue(record[cursorField]))
	return value == "" || value >= lowerBound
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case nil:
		return ""
	default:
		return fmt.Sprint(typed)
	}
}

func boolValue(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		parsed, _ := strconv.ParseBool(typed)
		return parsed
	default:
		return false
	}
}

func snakeish(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, ".", "_")
	value = strings.ReplaceAll(value, "/", "_")
	if value == "" {
		return "ashby"
	}
	return value
}

// requester builds a connsdk.Requester wired with Basic auth (api key as
// username, blank password), the resolved base URL, and the Ashby Accept header.
// The secret only ever flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := ashbyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := ashbySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("ashby connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, ""),
		UserAgent: ashbyUserAgent,
		Accept:    ashbyAccept,
	}, nil
}

func ashbySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// ashbyBaseURL resolves and validates the base URL. The default is
// api.ashbyhq.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func ashbyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return ashbyDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("ashby config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("ashby config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("ashby config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func ashbyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return ashbyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("ashby config page_size must be an integer: %w", err)
	}
	if value < 1 || value > ashbyMaxPageSize {
		return 0, fmt.Errorf("ashby config page_size must be between 1 and %d", ashbyMaxPageSize)
	}
	return value, nil
}

func ashbyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("ashby config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("ashby config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
