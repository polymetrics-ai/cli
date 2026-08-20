package engine

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/transportpolicy"
	"polymetrics.ai/internal/safety"
)

// httpErrorStatus extracts the HTTP status from err when it wraps a
// *connsdk.HTTPError, for delete's missing_ok_status matching.
func httpErrorStatus(err error) (int, bool) {
	var httpErr *connsdk.HTTPError
	if !errors.As(err, &httpErr) {
		return 0, false
	}
	return httpErr.Status, true
}

// findWriteAction resolves name against b.Writes.
func findWriteAction(b Bundle, name string) (WriteAction, error) {
	for _, a := range b.Writes {
		if a.Name == name {
			return a, nil
		}
	}
	return WriteAction{}, fmt.Errorf("engine: write action %q not found in bundle %q", name, b.Name)
}

// compiledRecordSchema lazily compiles action.RecordSchema. A write action
// with no record_schema declared skips validation entirely (e.g. actions
// whose body shape is fully hook-driven).
func compiledRecordSchema(action WriteAction) (*Schema, error) {
	if len(action.RecordSchema) == 0 {
		return nil, nil
	}
	sch, err := CompileSchema(action.RecordSchema)
	if err != nil {
		return nil, fmt.Errorf("engine: write action %q: compile record_schema: %w", action.Name, err)
	}
	return sch, nil
}

// ValidateWrite compiles action.record_schema once and validates every
// record against it. A structural error names the (0-indexed) record index,
// matching current per-connector behavior (e.g. stripe/write.go's "record
// %d" convention, 1-indexed there; engine reports 0-indexed since that is
// the natural Go slice index every caller already has).
func ValidateWrite(ctx context.Context, b Bundle, req connectors.WriteRequest, records []connectors.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	action, err := findWriteAction(b, req.Action)
	if err != nil {
		return err
	}
	sch, err := compiledRecordSchema(action)
	if err != nil {
		return err
	}
	for i, rec := range records {
		if sch != nil {
			if err := sch.Validate(map[string]any(rec)); err != nil {
				return &Error{Connector: b.Name, Action: action.Name, Page: -1, RecordIndex: i, Err: err}
			}
		}
		if err := validateWriteBody(action, rec); err != nil {
			return &Error{Connector: b.Name, Action: action.Name, Page: -1, RecordIndex: i, Err: err}
		}
	}
	return nil
}

func validateWriteBody(action WriteAction, rec connectors.Record) error {
	if bodyTypeOf(action) != "json_array" || len(action.BodySchema) == 0 {
		return nil
	}
	_, err := buildJSONArrayPayload(action, rec)
	return err
}

// DryRunWrite validates and prepares every record without a network call. Its
// warnings show the first fully resolved method/path that execution will send,
// while its digest binds the complete prepared request set and execution
// identity used by the approval gate.
func DryRunWrite(ctx context.Context, b Bundle, req connectors.WriteRequest, records []connectors.Record, h Hooks) (connectors.WritePreview, error) {
	prepared, err := prepareDeclarativeWrite(ctx, b, req, records, h)
	if err != nil {
		return connectors.WritePreview{}, err
	}
	return PreviewPreparedWrite(prepared)
}

// resolveWriteRequestLine interpolates the action's base URL and path against
// the same complete runtime configuration and record that execution uses.
func resolveWriteRequestLine(b Bundle, action WriteAction, rec connectors.Record, cfg connectors.RuntimeConfig) (method, path string, err error) {
	vars := Vars{Config: cfg.Config, Secrets: cfg.Secrets, Record: map[string]any(rec)}

	baseURL, err := Interpolate(b.HTTP.URL, vars)
	if err != nil {
		return "", "", fmt.Errorf("engine: resolve write base url: %w", err)
	}
	relPath, err := InterpolatePath(action.Path, vars)
	if err != nil {
		return "", "", fmt.Errorf("engine: write action %q: resolve path: %w", action.Name, err)
	}
	return methodOrDefault(action.Method), joinURL(baseURL, relPath), nil
}

func copyRecordMap(src map[string]any) map[string]any {
	out := make(map[string]any, len(src))
	for k, v := range src {
		switch typed := v.(type) {
		case map[string]any:
			out[k] = copyRecordMap(typed)
		case connectors.Record:
			out[k] = copyRecordMap(map[string]any(typed))
		default:
			out[k] = v
		}
	}
	return out
}

type writeActionRedactedError struct {
	err    error
	values []string
}

func (e *writeActionRedactedError) Error() string {
	message := e.err.Error()
	var httpErr *connsdk.HTTPError
	if errors.Is(e.err, transportpolicy.ErrRedirectRefused) {
		message = transportpolicy.ErrRedirectRefused.Error()
	} else if errors.As(e.err, &httpErr) {
		message = fmt.Sprintf("provider returned HTTP status %d", httpErr.Status)
	}
	if len(e.values) != 0 {
		message += ": protected request values redacted"
	}
	return safety.RedactErrorText(redactWriteLiterals(message, e.values))
}

func (e *writeActionRedactedError) Unwrap() error {
	return e.err
}

func redactWriteActionError(err error, action WriteAction, rec connectors.Record) error {
	if err == nil {
		return err
	}
	values := writeActionRedactionValues(action, rec)
	var httpErr *connsdk.HTTPError
	if len(values) == 0 && !errors.As(err, &httpErr) {
		return err
	}
	return &writeActionRedactedError{err: err, values: values}
}

func writeActionRedactionValues(action WriteAction, rec connectors.Record) []string {
	record := copyRecordMap(map[string]any(rec))
	seen := map[string]bool{}
	for _, field := range action.RedactFields {
		field = strings.TrimPrefix(strings.TrimSpace(field), "record.")
		if field == "" {
			continue
		}
		value, err := resolveRecordPathValue(record, strings.Split(field, "."))
		if err != nil {
			continue
		}
		collectWriteRedactionValues(value, seen)
	}
	values := make([]string, 0, len(seen))
	for value := range seen {
		values = append(values, value)
	}
	sortWriteRedactionLiterals(values)
	return values
}

func collectWriteRedactionValues(value any, out map[string]bool) {
	switch typed := value.(type) {
	case nil:
		return
	case string:
		addWriteRedactionValue(typed, out)
	case []string:
		for _, item := range typed {
			addWriteRedactionValue(item, out)
		}
	case []any:
		for _, item := range typed {
			collectWriteRedactionValues(item, out)
		}
	case map[string]any:
		for _, item := range typed {
			collectWriteRedactionValues(item, out)
		}
	case connectors.Record:
		for _, item := range typed {
			collectWriteRedactionValues(item, out)
		}
	default:
		addWriteRedactionValue(stringify(typed), out)
	}
}

func addWriteRedactionValue(value string, out map[string]bool) {
	value = strings.TrimSpace(value)
	if value == "" || value == "***" || value == "redacted" {
		return
	}
	out[value] = true
}

func redactWriteLiterals(text string, values []string) string {
	for _, literal := range writeRedactionLiterals(values) {
		text = strings.ReplaceAll(text, literal, "redacted")
	}
	return text
}

func sortWriteRedactionLiterals(values []string) {
	sort.Slice(values, func(i, j int) bool {
		if len(values[i]) != len(values[j]) {
			return len(values[i]) > len(values[j])
		}
		return values[i] < values[j]
	})
}

func writeRedactionLiterals(values []string) []string {
	seen := map[string]bool{}
	literals := make([]string, 0, len(values)*4)
	for _, value := range values {
		for _, literal := range writeRedactionLiteralForms(value) {
			if seen[literal] {
				continue
			}
			seen[literal] = true
			literals = append(literals, literal)
		}
	}
	sortWriteRedactionLiterals(literals)
	return literals
}

func writeRedactionLiteralForms(value string) []string {
	forms := []string{value, urlencodeSegment(value), url.QueryEscape(value), url.PathEscape(value)}
	seen := map[string]bool{}
	out := make([]string, 0, len(forms))
	for _, form := range forms {
		form = strings.TrimSpace(form)
		if form == "" || seen[form] {
			continue
		}
		seen[form] = true
		out = append(out, form)
	}
	return out
}

func joinURL(base, path string) string {
	if path == "" {
		return base
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	trimmedBase := strings.TrimRight(base, "/")
	return trimmedBase + "/" + strings.TrimLeft(path, "/")
}

// Write executes action per record, one HTTP request per record (design
// §B.5: batch semantics stay one-request-per-record in wave0). A WriteHook
// that returns handled=true for a record bypasses the declarative body/
// request construction entirely for that record. Accounting matches legacy
// fail-fast semantics (stripe/write.go:66): on the first failure (validation
// or per-record request error, including ctx cancellation), RecordsWritten
// reflects completed successes and RecordsFailed = len(records) -
// RecordsWritten; the loop stops immediately rather than continuing best-
// effort.
func Write(ctx context.Context, b Bundle, req connectors.WriteRequest, records []connectors.Record, h Hooks) (connectors.WriteResult, error) {
	prepared, err := prepareDeclarativeWrite(ctx, b, req, records, h)
	if err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	preview, err := PreviewPreparedWrite(prepared)
	if err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	action, err := findWriteAction(b, req.Action)
	if err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}

	var result connectors.WriteResult
	err = ExecutePreparedWrite(ctx, prepared, req.Approval, preview.Digest, func(executeCtx context.Context) error {
		var executeErr error
		result, executeErr = executeApprovedWrite(executeCtx, b, action, req, records, h)
		return executeErr
	})
	if err != nil && result.RecordsWritten == 0 && result.RecordsFailed == 0 {
		result.RecordsFailed = len(records)
	}
	return result, err
}

// applyWriteRecordHook returns the records the declarative body is built from.
// prepareDeclarativeWrite and executeApprovedWrite both call it, so the request
// an operator approves in the preview is the request that runs.
func applyWriteRecordHook(h Hooks, action WriteAction, records []connectors.Record) ([]connectors.Record, error) {
	mapper, ok := h.(WriteRecordHook)
	if !ok {
		return records, nil
	}
	mapped := make([]connectors.Record, 0, len(records))
	for _, rec := range records {
		pinned, handled, err := mapper.MapWriteRecord(action, rec)
		if err != nil {
			return nil, err
		}
		if !handled {
			pinned = rec
		}
		mapped = append(mapped, pinned)
	}
	return mapped, nil
}

func executeApprovedWrite(ctx context.Context, b Bundle, action WriteAction, req connectors.WriteRequest, records []connectors.Record, h Hooks) (connectors.WriteResult, error) {
	cfg := materializeConfigDefaults(b, req.Config)

	rt, err := newRuntime(ctx, b, cfg, h)
	if err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	result := connectors.WriteResult{}
	for i, rec := range records {
		if err := ctx.Err(); err != nil {
			result.RecordsFailed = len(records) - result.RecordsWritten - result.RecordsUnchanged
			return result, err
		}

		if wh, ok := h.(WriteHook); ok {
			handled, err := wh.ExecuteWrite(ctx, action, rec, rt)
			if err != nil {
				result.RecordsFailed = len(records) - result.RecordsWritten - result.RecordsUnchanged
				return result, &Error{Connector: b.Name, Action: action.Name, Page: -1, RecordIndex: i, Err: redactWriteActionError(err, action, rec)}
			}
			if handled {
				result.RecordsWritten++
				continue
			}
		}

		pinned, err := applyWriteRecordHook(h, action, []connectors.Record{rec})
		if err != nil {
			result.RecordsFailed = len(records) - result.RecordsWritten - result.RecordsUnchanged
			return result, &Error{Connector: b.Name, Action: action.Name, Page: -1, RecordIndex: i, Err: redactWriteActionError(err, action, rec)}
		}
		response, err := executeWriteRecordWithResponse(ctx, b, action, pinned[0], i, cfg, rt)
		var responseErr error
		if response != nil {
			providerResponse, providerResponseErr := writeProviderResponse(response, i)
			result.ProviderResponses = append(result.ProviderResponses, providerResponse)
			responseErr = providerResponseErr
		}
		if err != nil {
			if isMissingOkDelete(action, err) {
				result.RecordsUnchanged++
				continue
			}
			result.RecordsFailed = len(records) - result.RecordsWritten - result.RecordsUnchanged
			class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
			return result, &Error{Connector: b.Name, Action: action.Name, Page: -1, RecordIndex: i, Class: class, Hint: hint, Err: redactWriteActionError(err, action, rec)}
		}
		if responseErr != nil {
			result.RecordsFailed = len(records) - result.RecordsWritten - result.RecordsUnchanged
			return result, &Error{Connector: b.Name, Action: action.Name, Page: -1, RecordIndex: i, Err: redactWriteActionError(responseErr, action, rec)}
		}
		result.RecordsWritten++
	}
	return result, nil
}

// executeWriteRecord performs the single HTTP request for one record: builds
// the path from path_fields, the body per body_type, and issues Do/DoForm.
func executeWriteRecord(ctx context.Context, b Bundle, action WriteAction, rec connectors.Record, recordIndex int, cfg connectors.RuntimeConfig, rt *Runtime) error {
	_, err := executeWriteRecordWithResponse(ctx, b, action, rec, recordIndex, cfg, rt)
	return err
}

// executeWriteRecordWithResponse is the private result-preserving form used
// by the named-action executor. The exported connector surface remains the
// closed WriteAction contract; no caller can provide a route, verb, or body.
func executeWriteRecordWithResponse(ctx context.Context, b Bundle, action WriteAction, rec connectors.Record, recordIndex int, cfg connectors.RuntimeConfig, rt *Runtime) (*connsdk.Response, error) {
	vars := Vars{Config: cfg.Config, Secrets: cfg.Secrets, Record: map[string]any(rec)}

	path, err := InterpolatePath(action.Path, vars)
	if err != nil {
		return nil, fmt.Errorf("engine: write action %q: resolve path: %w", action.Name, err)
	}
	method := methodOrDefault(action.Method)

	// Resolved ONCE here and threaded through every body_type branch, so a
	// declared query can never silently apply to some body types and not
	// others. buildWriteQuery returns a nil url.Values when the action
	// declares no query, which is byte-identical to the literal nil every
	// branch passed before write-action query support existed.
	query, err := buildWriteQuery(action, vars)
	if err != nil {
		return nil, err
	}
	requesterForAction, err := rt.requesterFor(method, action.Path)
	if err != nil {
		return nil, err
	}
	requester, err := writeRequester(requesterForAction, action)
	if err != nil {
		return nil, err
	}

	switch bodyTypeOf(action) {
	case "form":
		form := buildForm(rec, action.PathFields)
		return requester.DoForm(ctx, method, path, query, form)
	case "graphql":
		payload, err := buildGraphQLPayload(action.GraphQL, vars)
		if err != nil {
			return nil, err
		}
		resp, err := requester.Do(ctx, method, path, query, payload)
		if err != nil {
			return resp, err
		}
		if err := graphQLErrors(resp.Body); err != nil {
			return resp, errors.New("provider GraphQL response contains errors")
		}
		return resp, nil
	case "none":
		body := buildBodyFieldsPayload(rec, action.BodyFields)
		if len(body) == 0 {
			return requester.Do(ctx, method, path, query, nil)
		}
		return requester.Do(ctx, method, path, query, body)
	case "json_array":
		payload, err := buildJSONArrayPayload(action, rec)
		if err != nil {
			return nil, err
		}
		return requester.Do(ctx, method, path, query, payload)
	case "multipart":
		root, err := openMultipartRoot(cfg.ProjectDir)
		if err != nil {
			return nil, fmt.Errorf("engine: write action %q: %w", action.Name, err)
		}
		defer func() { _ = root.Close() }()
		form, err := buildMultipartPayload(action, rec, recordIndex, cfg, root)
		if err != nil {
			return nil, err
		}
		return requester.DoMultipart(ctx, method, path, query, form)
	case "base64_upload":
		payload, err := buildBase64UploadPayload(action, rec, recordIndex, cfg)
		if err != nil {
			return nil, err
		}
		return requester.Do(ctx, method, path, query, payload)
	default: // "json" (default)
		var body map[string]any
		if len(action.BodyFields) > 0 {
			body = buildBodyFieldsPayload(rec, action.BodyFields)
		} else {
			body = buildJSONBody(rec, action.PathFields)
		}
		body, err = applyDynamicFields(action, rec, body)
		if err != nil {
			return nil, err
		}
		var payload any
		if len(body) > 0 || action.BodyRequired {
			if body == nil {
				body = map[string]any{}
			}
			payload = body
		}
		return requester.Do(ctx, method, path, query, payload)
	}
}

func writeProviderResponse(response *connsdk.Response, recordIndex int) (connectors.WriteProviderResponse, error) {
	result := connectors.WriteProviderResponse{
		RecordIndex: recordIndex,
		Status:      response.Status,
		Headers:     writeProviderHeaders(response.Header),
	}
	if !writeProviderResponseDeclaresJSON(response.Header) {
		result.Body, result.BodyEncoding = writeProviderRawBody(response.Body)
		return result, nil
	}
	body, err := decodeDirectReadBody(response.Body, -1)
	if err != nil {
		result.Body, result.BodyEncoding = writeProviderRawBody(response.Body)
		return result, errors.New("provider response is not valid JSON")
	}
	result.Body = body
	return result, nil
}

func writeProviderRawBody(body []byte) (any, string) {
	if utf8.Valid(body) {
		return string(body), "text"
	}
	return base64.StdEncoding.EncodeToString(body), "base64"
}

func writeProviderResponseDeclaresJSON(headers map[string][]string) bool {
	_, declared := writeProviderResponseContentType(headers)
	return declared
}

func writeProviderResponseContentType(headers map[string][]string) (present, json bool) {
	for name, values := range headers {
		if !strings.EqualFold(name, "Content-Type") {
			continue
		}
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			present = true
			mediaType, _, err := mime.ParseMediaType(value)
			if err != nil {
				return present, false
			}
			mediaType = strings.ToLower(mediaType)
			if mediaType != "application/json" && !strings.HasSuffix(mediaType, "+json") {
				return present, false
			}
			json = true
		}
	}
	return present, json
}

func writeProviderHeaders(headers map[string][]string) map[string]connectors.WriteProviderHeader {
	if len(headers) == 0 {
		return nil
	}
	result := make(map[string]connectors.WriteProviderHeader, len(headers))
	for name, values := range headers {
		result[name] = connectors.WriteProviderHeader{Values: append([]string(nil), values...)}
	}
	return result
}

// writeRequester clones the shared requester and permits mutation replay only
// when the action carries provider-scoped idempotency evidence.
func writeRequester(base *connsdk.Requester, action WriteAction) (*connsdk.Requester, error) {
	if base == nil {
		return nil, fmt.Errorf("engine: write action %q: requester is nil", action.Name)
	}
	header := strings.TrimSpace(action.IdempotencyKeyHeader)
	requester := *base
	requester.DefaultHeaders = make(map[string]string, len(base.DefaultHeaders)+1)
	for name, value := range base.DefaultHeaders {
		if header != "" && strings.EqualFold(name, header) {
			continue
		}
		requester.DefaultHeaders[name] = value
	}
	if header == "" {
		if action.Kind != "delete" || action.Delete == nil || !action.Delete.Idempotent {
			requester.DisableRetries = true
		}
		return &requester, nil
	}
	keyBytes := make([]byte, 16)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("engine: write action %q: create idempotency key: %w", action.Name, err)
	}
	requester.DefaultHeaders[header] = hex.EncodeToString(keyBytes)
	return &requester, nil
}

const (
	defaultDynamicFieldsMaxKeys       = 100
	defaultDynamicFieldsMaxValueBytes = 4096
)

// applyDynamicFields is the execution-time half of the typed dynamic-key
// contract (the declaration-time half is bundle.go's validateDynamicFields).
// It is the ONLY path by which a caller-supplied key reaches a request body,
// and it is deliberately narrow:
//
//   - the region must be a JSON object under the declared field;
//   - every key must match the bundle-declared, both-ends-anchored pattern;
//   - every value must be a SCALAR of a declared type — an object or array is
//     rejected outright, which is what stops caller input from ever becoming
//     request structure;
//   - no key may shadow a structural key already present in the body.
//
// A nil spec returns body untouched, so actions without dynamic_fields behave
// exactly as they did before this capability existed.
func applyDynamicFields(action WriteAction, rec connectors.Record, body map[string]any) (map[string]any, error) {
	spec := action.DynamicFields
	if spec == nil {
		return body, nil
	}
	fail := func(format string, args ...any) error {
		return fmt.Errorf("engine: write action %q: dynamic_fields: "+format, append([]any{action.Name}, args...)...)
	}

	raw, present := rec[spec.Field]
	// The container is always consumed, so it never also serializes as a
	// nested object when the target is inline.
	delete(body, spec.Field)
	if !present || raw == nil {
		return body, nil
	}
	region, ok := raw.(map[string]any)
	if !ok {
		return nil, fail("field %q must be an object, got %T", spec.Field, raw)
	}
	if len(region) == 0 {
		return body, nil
	}

	maxKeys := spec.MaxKeys
	if maxKeys <= 0 {
		maxKeys = defaultDynamicFieldsMaxKeys
	}
	if len(region) > maxKeys {
		return nil, fail("%d keys exceeds max_keys %d", len(region), maxKeys)
	}
	maxValueBytes := spec.MaxValueBytes
	if maxValueBytes <= 0 {
		maxValueBytes = defaultDynamicFieldsMaxValueBytes
	}
	pattern, err := compileDynamicKeyPattern(spec.KeyPattern)
	if err != nil {
		return nil, fail("key_pattern: %w", err)
	}
	allowed := map[string]bool{}
	for _, vt := range spec.ValueTypes {
		allowed[vt] = true
	}
	if len(allowed) == 0 {
		allowed = dynamicFieldsValueTypes
	}

	// Structural keys a dynamic key must never shadow.
	reserved := toSet(action.PathFields)
	for _, bf := range action.BodyFields {
		reserved[bf] = true
	}
	if action.BodyField != "" {
		reserved[action.BodyField] = true
	}
	reserved[spec.Field] = true

	// Sorted for deterministic error messages across runs.
	keys := make([]string, 0, len(region))
	for k := range region {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make(map[string]any, len(region))
	for _, key := range keys {
		if !pattern.MatchString(key) {
			return nil, fail("key %q does not match key_pattern %q", key, spec.KeyPattern)
		}
		if reserved[key] {
			return nil, fail("key %q collides with a declared path/body field", key)
		}
		typeName, err := dynamicScalarType(region[key])
		if err != nil {
			return nil, fail("key %q: %w", key, err)
		}
		if !allowed[typeName] {
			return nil, fail("key %q has value type %q which is not in value_types", key, typeName)
		}
		if n := dynamicValueEncodedLen(region[key]); n > maxValueBytes {
			return nil, fail("key %q value is %d bytes, exceeds max_value_bytes %d", key, n, maxValueBytes)
		}
		out[key] = region[key]
	}

	if strings.TrimSpace(spec.Target) == "nested" {
		if body == nil {
			body = map[string]any{}
		}
		body[spec.Field] = out
		return body, nil
	}
	if body == nil {
		body = make(map[string]any, len(out))
	}
	for key, value := range out {
		// Guards against a dynamic key shadowing a body key that came from
		// the record itself rather than from a declared field list.
		if _, exists := body[key]; exists {
			return nil, fail("key %q collides with an existing body key", key)
		}
		body[key] = value
	}
	return body, nil
}

// dynamicScalarType names the JSON scalar type of v, or errors when v is a
// composite. Rejecting composites here is the load-bearing anti-escape-hatch
// invariant, so this deliberately has no "flatten" or "coerce" branch.
func dynamicScalarType(v any) (string, error) {
	switch v.(type) {
	case nil:
		return "null", nil
	case string:
		return "string", nil
	case bool:
		return "boolean", nil
	case json.Number, float64, float32,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return "number", nil
	case map[string]any, []any:
		return "", fmt.Errorf("value must be a scalar (string, number, boolean, null), got %T", v)
	default:
		return "", fmt.Errorf("value must be a scalar (string, number, boolean, null), got %T", v)
	}
}

func dynamicValueEncodedLen(v any) int {
	if s, ok := v.(string); ok {
		return len(s)
	}
	raw, err := json.Marshal(v)
	if err != nil {
		// Unmarshalable values are already rejected by dynamicScalarType;
		// treat an encoding failure as maximally large rather than as zero.
		return math.MaxInt
	}
	return len(raw)
}

// buildWriteQuery resolves action.Query against vars using the IDENTICAL
// resolveQueryParams semantics stream.Query and check.query use — see that
// function's doc comment in read.go. vars is the same Vars the path was just
// interpolated from, so a query template may reference record fields exactly
// as a path template can.
//
// A nil/empty Query returns a nil url.Values rather than an empty one: an
// empty url.Values would still take resolveURL's "len(query) > 0" branch as
// false, but returning nil keeps the pre-existing call shape literally
// unchanged for every action that declares no query.
func buildWriteQuery(action WriteAction, vars Vars) (url.Values, error) {
	if len(action.Query) == 0 {
		return nil, nil
	}
	q, err := resolveQueryParams(action.Query, vars)
	if err != nil {
		return nil, fmt.Errorf("engine: write action %q: %w", action.Name, err)
	}
	return q, nil
}

func bodyTypeOf(action WriteAction) string {
	if action.BodyType == "" {
		return "json"
	}
	return action.BodyType
}

// buildJSONBody returns every record field not consumed by path_fields
// (design §B.2 default body construction rule).
func buildJSONBody(rec connectors.Record, pathFields []string) map[string]any {
	excluded := toSet(pathFields)
	out := make(map[string]any, len(rec))
	for k, v := range rec {
		if excluded[k] {
			continue
		}
		out[k] = v
	}
	return out
}

// buildBodyFieldsPayload returns only the allow-listed body_fields present
// on rec (used for delete-with-body actions like github's delete_file).
func buildBodyFieldsPayload(rec connectors.Record, bodyFields []string) map[string]any {
	if len(bodyFields) == 0 {
		return nil
	}
	out := make(map[string]any, len(bodyFields))
	for _, f := range bodyFields {
		if v, ok := rec[f]; ok {
			out[f] = v
		}
	}
	return out
}

func buildJSONArrayPayload(action WriteAction, rec connectors.Record) (any, error) {
	value, err := resolveRecordPathValue(map[string]any(rec), strings.Split(action.BodyField, "."))
	if err != nil {
		return nil, fmt.Errorf("engine: write action %q: resolve body_field %q: %w", action.Name, action.BodyField, err)
	}
	if len(action.BodySchema) > 0 {
		sch, err := CompileSchema(action.BodySchema)
		if err != nil {
			return nil, fmt.Errorf("engine: write action %q: compile body_schema: %w", action.Name, err)
		}
		if err := sch.Validate(value); err != nil {
			return nil, fmt.Errorf("engine: write action %q: body_schema: %w", action.Name, err)
		}
	}
	return value, nil
}

// maxBase64UploadDecodedBytes is the hard ceiling a declared max_decoded_bytes
// is clamped against, mirroring maxOperationDirectReadBytes on the inbound
// side. A base64 body is inherently buffered — there is no streaming form of it
// — so the ceiling IS the containment against memory exhaustion.
const maxBase64UploadDecodedBytes = 16 << 20

// buildBase64UploadPayload builds the JSON body for body_type
// "base64_upload": the ordinary declared body, with the source field removed
// and the declared content field set to a validated, canonically encoded
// payload.
//
// Bounds and containment deliberately mirror the bounded binary-download
// direction so the two halves of file transfer obey one set of rules: clamp the
// declared bound against a hard engine ceiling, read one byte past the limit and
// REJECT rather than truncate (a truncated attachment is a silently corrupt
// upload), confine filesystem access with os.Root, and verify the approved
// payload digest before anything is transmitted.
func buildBase64UploadPayload(action WriteAction, rec connectors.Record, recordIndex int, cfg connectors.RuntimeConfig) (map[string]any, error) {
	spec := action.Base64Upload
	if spec == nil {
		return nil, fmt.Errorf("engine: write action %q: base64_upload spec is required", action.Name)
	}

	var body map[string]any
	if len(action.BodyFields) > 0 {
		body = buildBodyFieldsPayload(rec, action.BodyFields)
	} else {
		body = buildJSONBody(rec, action.PathFields)
	}
	if body == nil {
		body = map[string]any{}
	}

	decoded, err := resolveBase64UploadPayload(action, spec, rec, recordIndex, cfg)
	if err != nil {
		return nil, err
	}

	encoded := base64.StdEncoding.EncodeToString(decoded)
	if maxEncoded := base64UploadMaxEncodedBytes(spec); int64(len(encoded)) > maxEncoded {
		return nil, fmt.Errorf("engine: write action %q: base64 upload too large: %d encoded bytes exceeds limit %d", action.Name, len(encoded), maxEncoded)
	}

	// The source field must never reach the wire: in "path" mode it is a local
	// filesystem path, and transmitting it would hand the provider the
	// operator's directory layout.
	delete(body, spec.SourceField)
	body[spec.ContentField] = encoded
	return body, nil
}

// resolveBase64UploadPayload returns the decoded payload bytes for either source
// mode, bounded by the clamped decoded limit.
func resolveBase64UploadPayload(action WriteAction, spec *Base64UploadSpec, rec connectors.Record, recordIndex int, cfg connectors.RuntimeConfig) ([]byte, error) {
	maxDecoded := clampBase64UploadDecodedBytes(spec.MaxDecodedBytes)

	raw, err := resolveRecordPathValue(map[string]any(rec), []string{spec.SourceField})
	if err != nil {
		return nil, fmt.Errorf("engine: write action %q: resolve base64_upload source_field %q: %w", action.Name, spec.SourceField, err)
	}
	value, ok := raw.(string)
	if !ok || strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("engine: write action %q: base64_upload source_field %q requires a non-empty string", action.Name, spec.SourceField)
	}

	if spec.Source == "base64" {
		// "Official base64" in the sense the operation ledgers use: RFC 4648
		// standard alphabet, canonical padding, no embedded line breaks.
		//
		// Two checks, because neither alone is enough. Strict() rejects
		// non-zero trailing padding bits, which the ordinary decoder silently
		// accepts — but Go's decoder skips \r and \n UNCONDITIONALLY, in Strict
		// mode too, so a wrapped MIME-style payload would sail through and be
		// re-encoded into something the operator never wrote. The alphabet check
		// is what actually enforces "one canonical token".
		if err := requireCanonicalBase64Alphabet(value); err != nil {
			return nil, fmt.Errorf("engine: write action %q: base64_upload source_field %q: %w", action.Name, spec.SourceField, err)
		}
		decoded, err := base64.StdEncoding.Strict().DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("engine: write action %q: base64_upload source_field %q is not canonical base64: %w", action.Name, spec.SourceField, err)
		}
		if int64(len(decoded)) > maxDecoded {
			return nil, fmt.Errorf("engine: write action %q: base64 upload too large: %d decoded bytes exceeds limit %d", action.Name, len(decoded), maxDecoded)
		}
		return decoded, nil
	}

	decoded, err := readBoundedProjectFile(cfg.ProjectDir, value, maxDecoded)
	if err != nil {
		return nil, fmt.Errorf("engine: write action %q: base64_upload source_field %q: %w", action.Name, spec.SourceField, err)
	}
	if err := verifyApprovedPayload(action, spec.SourceField, recordIndex, decoded, cfg); err != nil {
		return nil, err
	}
	return decoded, nil
}

// requireCanonicalBase64Alphabet rejects any byte outside RFC 4648's standard
// alphabet and its padding character — line breaks and whitespace included, and
// the URL-safe '-'/'_' pair with them.
func requireCanonicalBase64Alphabet(value string) error {
	for i := 0; i < len(value); i++ {
		c := value[i]
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case c == '+', c == '/', c == '=':
		default:
			return fmt.Errorf("value is not canonical base64: unexpected character at offset %d", i)
		}
	}
	return nil
}

// verifyApprovedPayload binds the bytes about to be transmitted to the bytes the
// operator approved, mirroring the multipart file part's contract exactly: when
// the runtime carries an approval map at all, this field must appear in it and
// its digest must match. A nil map means no approval flow is in play (a direct
// engine.Write, e.g. from a test), which is the existing multipart behaviour.
func verifyApprovedPayload(action WriteAction, field string, recordIndex int, payload []byte, cfg connectors.RuntimeConfig) error {
	if cfg.ApprovedPayloadSHA256 == nil {
		return nil
	}
	expected := cfg.ApprovedPayloadSHA256[connectors.PayloadApprovalKey(recordIndex, field)]
	if expected == "" {
		return fmt.Errorf("engine: write action %q: base64_upload source_field %q is missing its approved payload digest", action.Name, field)
	}
	sum := sha256.Sum256(payload)
	if !strings.EqualFold(hex.EncodeToString(sum[:]), expected) {
		return fmt.Errorf("engine: write action %q: base64_upload source_field %q payload does not match its approved digest", action.Name, field)
	}
	return nil
}

// readBoundedProjectFile reads at most maxBytes from a regular file confined to
// projectDir.
//
// Containment is os.Root, not a lexical prefix check: os.Root refuses a path
// that escapes the root through "..", through an absolute component, or through
// a symlink, and it does so on the OPEN rather than on a separate stat, which
// closes the check-then-open race that resolveMultipartFilePath's
// EvalSymlinks-then-Stat sequence leaves open. The lexical
// safety.ValidateLocalWritePath check still runs first: it is cheap, it rejects
// control characters, and it produces the error message operators already know.
func readBoundedProjectFile(projectDir, raw string, maxBytes int64) ([]byte, error) {
	if strings.TrimSpace(projectDir) == "" {
		projectDir = "."
	}
	if err := safety.ValidateLocalWritePath(projectDir, raw, "base64 upload source path", false); err != nil {
		return nil, err
	}
	rootAbs, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}

	rel := filepath.Clean(filepath.FromSlash(raw))
	if filepath.IsAbs(rel) {
		rel, err = filepath.Rel(rootAbs, rel)
		if err != nil {
			return nil, fmt.Errorf("compare source path to project root: %w", err)
		}
	}
	if !filepath.IsLocal(rel) {
		return nil, fmt.Errorf("source path outside the project root is not allowed")
	}

	root, err := os.OpenRoot(rootAbs)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()

	file, err := root.Open(rel)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("source must be a regular file")
	}
	if info.Size() > maxBytes {
		return nil, fmt.Errorf("base64 upload too large: %d bytes exceeds limit %d", info.Size(), maxBytes)
	}

	// Read one byte past the bound so a file that grew between Stat and Read is
	// rejected rather than silently truncated into a corrupt upload.
	payload, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(payload)) > maxBytes {
		return nil, fmt.Errorf("base64 upload too large: exceeds limit %d", maxBytes)
	}
	return payload, nil
}

// clampBase64UploadDecodedBytes clamps a declared bound against the engine
// ceiling, matching clampOperationDirectReadMaxBytes's request → spec → ceiling
// shape on the inbound side. Bundle load already requires a positive bound; the
// non-positive fallback here keeps a hand-built spec (a test, a future caller)
// bounded rather than unbounded.
func clampBase64UploadDecodedBytes(declared int64) int64 {
	if declared <= 0 || declared > maxBase64UploadDecodedBytes {
		return maxBase64UploadDecodedBytes
	}
	return declared
}

// base64UploadMaxEncodedBytes returns the encoded-size bound, defaulting to the
// base64 length of the clamped decoded bound when none is declared.
func base64UploadMaxEncodedBytes(spec *Base64UploadSpec) int64 {
	if spec.MaxEncodedBytes > 0 {
		return spec.MaxEncodedBytes
	}
	return int64(base64.StdEncoding.EncodedLen(int(clampBase64UploadDecodedBytes(spec.MaxDecodedBytes))))
}

func buildMultipartPayload(action WriteAction, rec connectors.Record, recordIndex int, cfg connectors.RuntimeConfig, root *os.Root) (connsdk.MultipartForm, error) {
	if action.Multipart == nil {
		return connsdk.MultipartForm{}, fmt.Errorf("engine: write action %q: multipart spec is required", action.Name)
	}
	form := connsdk.MultipartForm{Fields: map[string]string{}, MaxBytes: action.Multipart.MaxBytes}
	var total int64
	for _, part := range action.Multipart.Parts {
		value, err := resolveRecordPathValue(map[string]any(rec), strings.Split(part.Field, "."))
		if err != nil {
			if part.Required {
				return connsdk.MultipartForm{}, fmt.Errorf("engine: write action %q: multipart part %q: %w", action.Name, part.Name, err)
			}
			continue
		}
		if value == nil {
			if part.Required {
				return connsdk.MultipartForm{}, fmt.Errorf("engine: write action %q: multipart part %q is required", action.Name, part.Name)
			}
			continue
		}
		switch part.Type {
		case "field":
			form.Fields[part.Name] = stringifyAny(value)
		case "file":
			path, ok := value.(string)
			if !ok || strings.TrimSpace(path) == "" {
				return connsdk.MultipartForm{}, fmt.Errorf("engine: write action %q: multipart file part %q requires a file path string", action.Name, part.Name)
			}
			expectedSHA256 := cfg.ApprovedPayloadSHA256[connectors.PayloadApprovalKey(recordIndex, part.Field)]
			if cfg.ApprovedPayloadSHA256 != nil && expectedSHA256 == "" {
				return connsdk.MultipartForm{}, fmt.Errorf("engine: write action %q: multipart file part %q is missing its approved payload digest", action.Name, part.Name)
			}
			relPath, size, err := resolveMultipartFilePath(root, cfg.ProjectDir, path, part.MaxBytes)
			if err != nil {
				return connsdk.MultipartForm{}, fmt.Errorf("engine: write action %q: multipart file part %q: %w", action.Name, part.Name, err)
			}
			total += size
			if action.Multipart.MaxBytes > 0 && total > action.Multipart.MaxBytes {
				return connsdk.MultipartForm{}, fmt.Errorf("engine: write action %q: multipart payload too large: %d bytes exceeds limit %d", action.Name, total, action.Multipart.MaxBytes)
			}
			form.Files = append(form.Files, connsdk.MultipartFile{
				FieldName: part.Name,
				// Root and RelPath, not an absolute path: every later Stat and
				// Open re-checks containment instead of trusting this one.
				Root:              root,
				RelPath:           relPath,
				Path:              relPath,
				ContentType:       part.ContentType,
				AllowedMediaTypes: part.AllowedMediaTypes,
				MaxBytes:          part.MaxBytes,
				ExpectedSHA256:    expectedSHA256,
			})
		default:
			return connsdk.MultipartForm{}, fmt.Errorf("engine: write action %q: multipart part %q has unsupported type %q", action.Name, part.Name, part.Type)
		}
	}
	return form, nil
}

// openMultipartRoot opens the containment root for a multipart upload. Every
// access to a multipart source file goes through this root, so an escaping path
// is refused at each open rather than by a single check performed beforehand.
func openMultipartRoot(projectDir string) (*os.Root, error) {
	if strings.TrimSpace(projectDir) == "" {
		projectDir = "."
	}
	root, err := os.OpenRoot(projectDir)
	if err != nil {
		return nil, fmt.Errorf("open project root: %w", err)
	}
	return root, nil
}

// resolveMultipartFilePath converts a record-supplied path into a path relative
// to root and pre-checks its type and size.
//
// The returned path is deliberately root-relative rather than absolute: it is
// re-resolved through os.Root on every subsequent Stat and Open, which is what
// closes the check-then-open race the previous EvalSymlinks-then-compare
// approach left open. The lexical safety check is kept ahead of it as a cheap
// first filter with a clearer message, but it is no longer load-bearing for
// containment.
func resolveMultipartFilePath(root *os.Root, projectDir, raw string, maxBytes int64) (string, int64, error) {
	if strings.TrimSpace(projectDir) == "" {
		projectDir = "."
	}
	if err := safetyRejectLocalFilePath(projectDir, raw); err != nil {
		return "", 0, err
	}
	rel, err := multipartRootRelativePath(projectDir, raw)
	if err != nil {
		return "", 0, err
	}
	info, err := root.Stat(rel)
	if err != nil {
		return "", 0, err
	}
	if !info.Mode().IsRegular() {
		return "", 0, fmt.Errorf("file must be a regular file")
	}
	if maxBytes > 0 && info.Size() > maxBytes {
		return "", 0, fmt.Errorf("file too large: %d bytes exceeds limit %d", info.Size(), maxBytes)
	}
	return rel, info.Size(), nil
}

// multipartRootRelativePath expresses raw relative to projectDir. Absolute paths
// inside the project directory stay accepted, as they were before, but they are
// converted to root-relative form so os.Root can confine them. The project
// directory is compared both as given and with symlinks resolved, because on
// macOS a configured "/var/..." root and an absolute "/private/var/..." record
// value denote the same directory.
func multipartRootRelativePath(projectDir, raw string) (string, error) {
	clean := filepath.Clean(raw)
	if !filepath.IsAbs(clean) {
		if !filepath.IsLocal(clean) {
			return "", fmt.Errorf("multipart file path must stay inside the project root")
		}
		return clean, nil
	}
	rootAbs, err := filepath.Abs(projectDir)
	if err != nil {
		return "", fmt.Errorf("resolve project root: %w", err)
	}
	candidates := []string{rootAbs}
	if resolvedRoot, err := filepath.EvalSymlinks(rootAbs); err == nil && resolvedRoot != rootAbs {
		candidates = append(candidates, resolvedRoot)
	}
	for _, base := range candidates {
		rel, err := filepath.Rel(base, clean)
		if err != nil {
			continue
		}
		if rel == "." || filepath.IsLocal(rel) {
			return rel, nil
		}
	}
	return "", fmt.Errorf("multipart file path outside the project root is not allowed")
}

func safetyRejectLocalFilePath(projectDir, raw string) error {
	return safety.ValidateLocalWritePath(projectDir, raw, "multipart file path", false)
}

// buildForm builds a url.Values form body from every record field not
// consumed by path_fields, stringifying each value (matches
// stripe/write.go's customerForm shape/intent, generalized to any record).
func buildForm(rec connectors.Record, pathFields []string) url.Values {
	excluded := toSet(pathFields)
	keys := make([]string, 0, len(rec))
	for k := range rec {
		if excluded[k] {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys) // deterministic encoding order
	form := url.Values{}
	for _, k := range keys {
		v := rec[k]
		if v == nil {
			continue
		}
		if s, ok := v.(string); ok {
			if s == "" {
				continue
			}
			form.Set(k, s)
			continue
		}
		form.Set(k, stringifyAny(v))
	}
	return form
}

func stringifyAny(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

func toSet(items []string) map[string]bool {
	out := make(map[string]bool, len(items))
	for _, it := range items {
		out[it] = true
	}
	return out
}

// isMissingOkDelete reports whether err is an HTTP error whose status is
// listed in action.delete.missing_ok_status. The response is an expected
// idempotent no-op, but it is never a write: callers receive it through
// RecordsUnchanged so a command cannot claim a provider mutation occurred.
func isMissingOkDelete(action WriteAction, err error) bool {
	if action.Kind != "delete" || action.Delete == nil || len(action.Delete.MissingOkStatus) == 0 {
		return false
	}
	status, ok := httpErrorStatus(err)
	if !ok {
		return false
	}
	for _, s := range action.Delete.MissingOkStatus {
		if s == status {
			return true
		}
	}
	return false
}
