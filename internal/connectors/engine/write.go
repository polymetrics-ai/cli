package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
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

// DryRunWrite validates every record and returns a staged-count preview
// whose Warnings include the fully-resolved method/path for the FIRST
// record (representative preview; every record shares the same action). Any
// secret value is redacted (never interpolated in cleartext into the
// preview) — DryRunWrite performs no network call.
func DryRunWrite(ctx context.Context, b Bundle, req connectors.WriteRequest, records []connectors.Record, h Hooks) (connectors.WritePreview, error) {
	if err := ValidateWrite(ctx, b, req, records); err != nil {
		return connectors.WritePreview{}, err
	}
	action, err := findWriteAction(b, req.Action)
	if err != nil {
		return connectors.WritePreview{}, err
	}

	cfg := materializeConfigDefaults(b, req.Config)

	warnings := []string{fmt.Sprintf("%s executes a live mutation only after approval; dry run performs no external call", action.Name)}
	if len(records) > 0 {
		method, path, err := resolveWriteRequestLine(b, action, records[0], cfg)
		if err != nil {
			return connectors.WritePreview{}, err
		}
		warnings = append(warnings, fmt.Sprintf("resolved request: %s %s", method, path))
	}

	return connectors.WritePreview{
		RecordsStaged: len(records),
		Action:        action.Name,
		Warnings:      warnings,
	}, nil
}

// resolveWriteRequestLine interpolates the action's base URL and path
// against rec/cfg, redacting {{ secrets.* }} and action redaction fields so
// previews do not expose values that must stay out of operator-visible URLs.
func resolveWriteRequestLine(b Bundle, action WriteAction, rec connectors.Record, cfg connectors.RuntimeConfig) (method, path string, err error) {
	redactedSecrets := make(map[string]string, len(cfg.Secrets))
	for k := range cfg.Secrets {
		redactedSecrets[k] = "***"
	}
	vars := Vars{Config: cfg.Config, Secrets: redactedSecrets, Record: previewRecordForWriteAction(rec, action.RedactFields)}

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

func previewRecordForWriteAction(rec connectors.Record, redactFields []string) map[string]any {
	out := copyRecordMap(map[string]any(rec))
	for _, field := range redactFields {
		redactPreviewRecordField(out, strings.TrimPrefix(strings.TrimSpace(field), "record."))
	}
	return out
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

func redactPreviewRecordField(record map[string]any, field string) {
	if field == "" {
		return
	}
	parts := strings.Split(field, ".")
	current := record
	for _, part := range parts[:len(parts)-1] {
		next, ok := current[part]
		if !ok {
			return
		}
		switch typed := next.(type) {
		case map[string]any:
			current = typed
		case connectors.Record:
			current = map[string]any(typed)
		default:
			return
		}
	}
	if _, ok := current[parts[len(parts)-1]]; ok {
		current[parts[len(parts)-1]] = "redacted"
	}
}

type writeActionRedactedError struct {
	err    error
	values []string
}

func (e *writeActionRedactedError) Error() string {
	return safety.RedactErrorText(redactWriteLiterals(e.err.Error(), e.values))
}

func (e *writeActionRedactedError) Unwrap() error {
	return e.err
}

func redactWriteActionError(err error, action WriteAction, rec connectors.Record) error {
	if err == nil || len(action.RedactFields) == 0 {
		return err
	}
	values := writeActionRedactionValues(action, rec)
	if len(values) == 0 {
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
	if err := ValidateWrite(ctx, b, req, records); err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	action, err := findWriteAction(b, req.Action)
	if err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}

	cfg := materializeConfigDefaults(b, req.Config)

	rt, err := newRuntime(ctx, b, cfg, h)
	if err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}

	result := connectors.WriteResult{}
	for i, rec := range records {
		if err := ctx.Err(); err != nil {
			result.RecordsFailed = len(records) - result.RecordsWritten
			return result, err
		}

		if wh, ok := h.(WriteHook); ok {
			handled, err := wh.ExecuteWrite(ctx, action, rec, rt)
			if err != nil {
				result.RecordsFailed = len(records) - result.RecordsWritten
				return result, &Error{Connector: b.Name, Action: action.Name, Page: -1, RecordIndex: i, Err: redactWriteActionError(err, action, rec)}
			}
			if handled {
				result.RecordsWritten++
				continue
			}
		}

		if err := executeWriteRecord(ctx, b, action, rec, i, cfg, rt); err != nil {
			if isMissingOkDelete(action, err) {
				result.RecordsWritten++
				continue
			}
			result.RecordsFailed = len(records) - result.RecordsWritten
			class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
			return result, &Error{Connector: b.Name, Action: action.Name, Page: -1, RecordIndex: i, Class: class, Hint: hint, Err: redactWriteActionError(err, action, rec)}
		}
		result.RecordsWritten++
	}
	return result, nil
}

// executeWriteRecord performs the single HTTP request for one record: builds
// the path from path_fields, the body per body_type, and issues Do/DoForm.
func executeWriteRecord(ctx context.Context, b Bundle, action WriteAction, rec connectors.Record, recordIndex int, cfg connectors.RuntimeConfig, rt *Runtime) error {
	vars := Vars{Config: cfg.Config, Secrets: cfg.Secrets, Record: map[string]any(rec)}

	path, err := InterpolatePath(action.Path, vars)
	if err != nil {
		return fmt.Errorf("engine: write action %q: resolve path: %w", action.Name, err)
	}
	method := methodOrDefault(action.Method)

	// Resolved ONCE here and threaded through every body_type branch, so a
	// declared query can never silently apply to some body types and not
	// others. buildWriteQuery returns a nil url.Values when the action
	// declares no query, which is byte-identical to the literal nil every
	// branch passed before write-action query support existed.
	query, err := buildWriteQuery(action, vars)
	if err != nil {
		return err
	}

	switch bodyTypeOf(action) {
	case "form":
		form := buildForm(rec, action.PathFields)
		_, err := rt.Requester.DoForm(ctx, method, path, query, form)
		return err
	case "graphql":
		payload, err := buildGraphQLPayload(action.GraphQL, vars)
		if err != nil {
			return err
		}
		resp, err := rt.Requester.Do(ctx, method, path, query, payload)
		if err != nil {
			return err
		}
		return graphQLErrors(resp.Body)
	case "none":
		body := buildBodyFieldsPayload(rec, action.BodyFields)
		if len(body) == 0 {
			_, err := rt.Requester.Do(ctx, method, path, query, nil)
			return err
		}
		_, err := rt.Requester.Do(ctx, method, path, query, body)
		return err
	case "json_array":
		payload, err := buildJSONArrayPayload(action, rec)
		if err != nil {
			return err
		}
		_, err = rt.Requester.Do(ctx, method, path, query, payload)
		return err
	case "multipart":
		root, err := openMultipartRoot(cfg.ProjectDir)
		if err != nil {
			return fmt.Errorf("engine: write action %q: %w", action.Name, err)
		}
		defer root.Close()
		form, err := buildMultipartPayload(action, rec, recordIndex, cfg, root)
		if err != nil {
			return err
		}
		_, err = rt.Requester.DoMultipart(ctx, method, path, query, form)
		return err
	default: // "json" (default)
		var body map[string]any
		if len(action.BodyFields) > 0 {
			body = buildBodyFieldsPayload(rec, action.BodyFields)
		} else {
			body = buildJSONBody(rec, action.PathFields)
		}
		body, err = applyDynamicFields(action, rec, body)
		if err != nil {
			return err
		}
		var payload any
		if len(body) > 0 {
			payload = body
		}
		_, err := rt.Requester.Do(ctx, method, path, query, payload)
		return err
	}
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

func requireInsideRoot(rootAbs, pathAbs string) error {
	rel, err := filepath.Rel(rootAbs, pathAbs)
	if err != nil {
		return fmt.Errorf("compare multipart file path to project root: %w", err)
	}
	if rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)) {
		return nil
	}
	return fmt.Errorf("multipart file path outside the project root is not allowed")
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
// listed in action.delete.missing_ok_status — an idempotent delete's 404 (or
// whatever status the bundle allow-lists) counts as written, not failed
// (design §B.5).
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
