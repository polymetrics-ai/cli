package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
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
	if sch == nil {
		return nil
	}
	for i, rec := range records {
		if err := sch.Validate(map[string]any(rec)); err != nil {
			return &Error{Connector: b.Name, Action: action.Name, Page: -1, RecordIndex: i, Err: err}
		}
	}
	return nil
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

	warnings := []string{fmt.Sprintf("%s executes a live mutation only after approval; dry run performs no external call", action.Name)}
	if len(records) > 0 {
		method, path, err := resolveWriteRequestLine(b, action, records[0], req.Config)
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
// against rec/cfg, redacting {{ secrets.* }} references so a preview can
// never leak a secret value even though the method/path is otherwise fully
// resolved for operator review (THREAT-MODEL §1).
func resolveWriteRequestLine(b Bundle, action WriteAction, rec connectors.Record, cfg connectors.RuntimeConfig) (method, path string, err error) {
	redactedSecrets := make(map[string]string, len(cfg.Secrets))
	for k := range cfg.Secrets {
		redactedSecrets[k] = "***"
	}
	vars := Vars{Config: cfg.Config, Secrets: redactedSecrets, Record: map[string]any(rec)}

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

	rt, err := newRuntime(ctx, b, req.Config, h)
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
				return result, &Error{Connector: b.Name, Action: action.Name, Page: -1, RecordIndex: i, Err: err}
			}
			if handled {
				result.RecordsWritten++
				continue
			}
		}

		if err := executeWriteRecord(ctx, b, action, rec, req.Config, rt); err != nil {
			if isMissingOkDelete(action, err) {
				result.RecordsWritten++
				continue
			}
			result.RecordsFailed = len(records) - result.RecordsWritten
			class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
			return result, &Error{Connector: b.Name, Action: action.Name, Page: -1, RecordIndex: i, Class: class, Hint: hint, Err: err}
		}
		result.RecordsWritten++
	}
	return result, nil
}

// executeWriteRecord performs the single HTTP request for one record: builds
// the path from path_fields, the body per body_type, and issues Do/DoForm.
func executeWriteRecord(ctx context.Context, b Bundle, action WriteAction, rec connectors.Record, cfg connectors.RuntimeConfig, rt *Runtime) error {
	vars := Vars{Config: cfg.Config, Secrets: cfg.Secrets, Record: map[string]any(rec)}

	path, err := InterpolatePath(action.Path, vars)
	if err != nil {
		return fmt.Errorf("engine: write action %q: resolve path: %w", action.Name, err)
	}
	method := methodOrDefault(action.Method)

	query, err := buildWriteQuery(action.Query, vars)
	if err != nil {
		return fmt.Errorf("engine: write action %q: resolve query: %w", action.Name, err)
	}

	switch bodyTypeOf(action) {
	case "form":
		form := buildForm(rec, writeBodyExcludedFields(action))
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
	default: // "json" (default)
		var body map[string]any
		if len(action.BodyFields) > 0 {
			body = buildBodyFieldsPayload(rec, action.BodyFields)
		} else {
			body = buildJSONBody(rec, writeBodyExcludedFields(action))
		}
		var payload any
		if len(body) > 0 {
			payload = body
		}
		_, err := rt.Requester.Do(ctx, method, path, query, payload)
		return err
	}
}

func writeBodyExcludedFields(action WriteAction) []string {
	out := append([]string(nil), action.PathFields...)
	for field := range action.Query {
		out = append(out, field)
	}
	return out
}

func buildWriteQuery(templates map[string]string, vars Vars) (url.Values, error) {
	values := url.Values{}
	for name, tmpl := range templates {
		resolved, err := Interpolate(tmpl, vars)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if strings.TrimSpace(resolved) == "" {
			continue
		}
		values.Set(name, resolved)
	}
	return values, nil
}

func bodyTypeOf(action WriteAction) string {
	if action.BodyType == "" {
		return "json"
	}
	return action.BodyType
}

// buildJSONBody returns every record field not consumed by path or query
// fields (design §B.2 default body construction rule).
func buildJSONBody(rec connectors.Record, excludedFields []string) map[string]any {
	excluded := toSet(excludedFields)
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

// buildForm builds a url.Values form body from every record field not
// consumed by path/query fields, stringifying each value (matches
// stripe/write.go's customerForm shape/intent, generalized to any record).
func buildForm(rec connectors.Record, excludedFields []string) url.Values {
	excluded := toSet(excludedFields)
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
