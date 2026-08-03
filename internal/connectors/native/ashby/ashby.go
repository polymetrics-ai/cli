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
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	ashbydefs "polymetrics.ai/internal/connectors/defs/ashby"
)

const (
	ashbyDefaultBaseURL  = "https://api.ashbyhq.com"
	ashbyDefaultPageSize = 100
	// ashbyDefaultMaxPages is 0, meaning unbounded: the read follows Ashby's
	// nextCursor until the provider stops reporting more data. It must stay
	// equal to spec.json's max_pages default, which callers see as the
	// documented read bound.
	ashbyDefaultMaxPages = 0
	ashbyMaxPageSize     = 100
	ashbyUserAgent       = "polymetrics-go-cli"
	// ashbyAccept is sent per Ashby's documented API version header
	// shape (Accept: application/json; version=1).
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
	resp, err := r.Do(ctx, http.MethodPost, "apiKey.info", nil, map[string]any{})
	if err != nil {
		return fmt.Errorf("check ashby: %w", err)
	}
	if err := ashbyValidateSuccessEnvelope(resp.Body); err != nil {
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
	if ashbyHasSyncToken(req.Query) || ashbyHasSyncToken(req.State) || ashbyHasSyncToken(req.Config.Config) {
		return errors.New("ashby syncToken reads are blocked pending ashby-sync-token-checkpoint-foundation")
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
// and page count stay bounded here.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, req connectors.ReadRequest, pageSize, maxPages int, emit func(connectors.Record) error) error {
	pageCursor := ""
	seenPageCursors := map[string]struct{}{}
	_, supportsCursor := endpoint.requestFields["cursor"]
	baseBody, err := ashbyStreamBody(endpoint, req.Config, req.Query, pageSize)
	if err != nil {
		return err
	}
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		body := cloneMap(baseBody)
		if pageCursor != "" && supportsCursor {
			body["cursor"] = pageCursor
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
			if err := emit(record); err != nil {
				return err
			}
		}
		next := strings.TrimSpace(stringValue(pageBody["nextCursor"]))
		if !boolValue(pageBody["moreDataAvailable"]) || next == "" || !supportsCursor {
			return nil
		}
		if maxPages > 0 && page+1 >= maxPages {
			return nil
		}
		if _, seen := seenPageCursors[next]; seen {
			return fmt.Errorf("read ashby %s: repeated pagination cursor", endpoint.path)
		}
		seenPageCursors[next] = struct{}{}
		pageCursor = next
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	fixtures, err := ashbyFixtureFS()
	if err != nil {
		return err
	}
	pages, err := ashbyFixtureBodies(fixtures, stream)
	if err != nil {
		return err
	}
	if len(pages) == 0 {
		return fmt.Errorf("ashby stream %q has no replay fixtures", stream)
	}
	for pageIndex, body := range pages {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, records, err := ashbyResultRecords(body)
		if err != nil {
			return fmt.Errorf("decode ashby fixture %s page %d: %w", stream, pageIndex+1, err)
		}
		for recordIndex, item := range records {
			record := ashbyProjectRecord(endpoint, item, pageIndex, recordIndex)
			if err := emit(record); err != nil {
				return err
			}
		}
	}
	return nil
}

type ashbyFixturePage struct {
	Response struct {
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body"`
	} `json:"response"`
}

func ashbyFixtureFS() (fs.FS, error) {
	return ashbydefs.Fixtures()
}

func ashbyFixtureBodies(fixtures fs.FS, stream string) ([][]byte, error) {
	dir := path.Join("streams", stream)
	entries, err := fs.ReadDir(fixtures, dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read ashby fixtures %s: %w", dir, err)
	}
	bodies := make([][]byte, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		raw, err := fs.ReadFile(fixtures, path.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read ashby fixture %s/%s: %w", dir, entry.Name(), err)
		}
		body, err := ashbyFixtureBody(raw)
		if err != nil {
			return nil, fmt.Errorf("parse ashby fixture %s/%s: %w", dir, entry.Name(), err)
		}
		bodies = append(bodies, body)
	}
	return bodies, nil
}

func ashbyFixtureBody(raw []byte) ([]byte, error) {
	var page ashbyFixturePage
	if err := json.Unmarshal(raw, &page); err != nil {
		return nil, err
	}
	if len(page.Response.Body) == 0 {
		return raw, nil
	}
	if page.Response.Status >= 400 {
		return nil, fmt.Errorf("status %d", page.Response.Status)
	}
	return page.Response.Body, nil
}

func ashbyStreamBody(endpoint streamEndpoint, cfg connectors.RuntimeConfig, query map[string]string, pageSize int) (map[string]any, error) {
	if ashbyHasSyncToken(query) || ashbyHasSyncToken(cfg.Config) {
		return nil, errors.New("ashby syncToken reads are blocked pending ashby-sync-token-checkpoint-foundation")
	}
	body := map[string]any{}
	if _, ok := endpoint.requestFields["limit"]; ok {
		body["limit"] = pageSize
	}
	for field, kind := range endpoint.requestFields {
		if field == "limit" || field == "cursor" {
			continue
		}
		raw := strings.TrimSpace(query[field])
		if raw == "" {
			raw = strings.TrimSpace(cfg.Config[field])
		}
		if fixed := strings.TrimSpace(endpoint.fixedRequestFields[field]); fixed != "" {
			if raw != "" && !strings.EqualFold(raw, fixed) {
				if gap := strings.TrimSpace(endpoint.fixedRequestFieldGaps[field]); gap != "" {
					return nil, fmt.Errorf("ashby stream %s supports only %s=%s; non-default values are blocked pending variant-schema foundation %s", endpoint.path, field, fixed, gap)
				}
				return nil, fmt.Errorf("ashby stream %s supports only %s=%s; non-default values are blocked pending variant-schema foundation", endpoint.path, field, fixed)
			}
			raw = fixed
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
	if len(endpoint.requiredAnyFields) > 0 {
		present := false
		for _, field := range endpoint.requiredAnyFields {
			if _, ok := body[field]; ok {
				present = true
				break
			}
		}
		if !present {
			return nil, fmt.Errorf("ashby stream %s requires one of documented request fields %s", endpoint.path, strings.Join(endpoint.requiredAnyFields, ", "))
		}
	}
	return body, nil
}

func ashbyHasSyncToken(values map[string]string) bool {
	return strings.TrimSpace(values["syncToken"]) != "" || strings.TrimSpace(values["sync_token"]) != ""
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
	if err := ashbyValidateSuccessEnvelopeValue(decoded); err != nil {
		return nil, nil, err
	}
	results, ok := decoded["results"]
	if !ok || results == nil {
		return nil, nil, errors.New("ashby response missing results")
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
		return nil, nil, fmt.Errorf("ashby response results field has unsupported type %T", results)
	}
}

func ashbyValidateSuccessEnvelope(body []byte) error {
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return fmt.Errorf("decode ashby response envelope: %w", err)
	}
	return ashbyValidateSuccessEnvelopeValue(decoded)
}

func ashbyValidateSuccessEnvelopeValue(value any) error {
	decoded, ok := value.(map[string]any)
	if !ok {
		return errors.New("ashby response envelope must be an object")
	}
	rawSuccess, ok := decoded["success"]
	if !ok {
		return errors.New("ashby response missing success field")
	}
	success, ok := rawSuccess.(bool)
	if !ok {
		return errors.New("ashby response success field must be boolean")
	}
	if !success {
		return errors.New("ashby response success=false")
	}
	return nil
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
	if raw == "" {
		return ashbyDefaultMaxPages, nil
	}
	if raw == "all" || raw == "unlimited" {
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
