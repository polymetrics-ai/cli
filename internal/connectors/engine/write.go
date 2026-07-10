package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	switch bodyTypeOf(action) {
	case "form":
		form := buildForm(rec, action.PathFields)
		_, err := rt.Requester.DoForm(ctx, method, path, nil, form)
		return err
	case "graphql":
		payload, err := buildGraphQLPayload(action.GraphQL, vars)
		if err != nil {
			return err
		}
		resp, err := rt.Requester.Do(ctx, method, path, nil, payload)
		if err != nil {
			return err
		}
		return graphQLErrors(resp.Body)
	case "none":
		body := buildBodyFieldsPayload(rec, action.BodyFields)
		if len(body) == 0 {
			_, err := rt.Requester.Do(ctx, method, path, nil, nil)
			return err
		}
		_, err := rt.Requester.Do(ctx, method, path, nil, body)
		return err
	case "json_array":
		payload, err := buildJSONArrayPayload(action, rec)
		if err != nil {
			return err
		}
		_, err = rt.Requester.Do(ctx, method, path, nil, payload)
		return err
	case "multipart":
		form, err := buildMultipartPayload(action, rec, cfg)
		if err != nil {
			return err
		}
		_, err = rt.Requester.DoMultipart(ctx, method, path, nil, form)
		return err
	default: // "json" (default)
		var body map[string]any
		if len(action.BodyFields) > 0 {
			body = buildBodyFieldsPayload(rec, action.BodyFields)
		} else {
			body = buildJSONBody(rec, action.PathFields)
		}
		var payload any
		if len(body) > 0 {
			payload = body
		}
		_, err := rt.Requester.Do(ctx, method, path, nil, payload)
		return err
	}
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

func buildMultipartPayload(action WriteAction, rec connectors.Record, cfg connectors.RuntimeConfig) (connsdk.MultipartForm, error) {
	if action.Multipart == nil {
		return connsdk.MultipartForm{}, fmt.Errorf("engine: write action %q: multipart spec is required", action.Name)
	}
	form := connsdk.MultipartForm{Fields: map[string]string{}}
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
			resolved, size, err := resolveMultipartFilePath(cfg.ProjectDir, path, part.MaxBytes)
			if err != nil {
				return connsdk.MultipartForm{}, fmt.Errorf("engine: write action %q: multipart file part %q: %w", action.Name, part.Name, err)
			}
			total += size
			if action.Multipart.MaxBytes > 0 && total > action.Multipart.MaxBytes {
				return connsdk.MultipartForm{}, fmt.Errorf("engine: write action %q: multipart payload too large: %d bytes exceeds limit %d", action.Name, total, action.Multipart.MaxBytes)
			}
			form.Files = append(form.Files, connsdk.MultipartFile{
				FieldName:   part.Name,
				Path:        resolved,
				ContentType: part.ContentType,
				MaxBytes:    part.MaxBytes,
			})
		default:
			return connsdk.MultipartForm{}, fmt.Errorf("engine: write action %q: multipart part %q has unsupported type %q", action.Name, part.Name, part.Type)
		}
	}
	return form, nil
}

func resolveMultipartFilePath(projectDir, raw string, maxBytes int64) (string, int64, error) {
	if strings.TrimSpace(projectDir) == "" {
		projectDir = "."
	}
	if err := safetyRejectLocalFilePath(projectDir, raw); err != nil {
		return "", 0, err
	}
	rootAbs, err := filepath.Abs(projectDir)
	if err != nil {
		return "", 0, fmt.Errorf("resolve project root: %w", err)
	}
	if resolvedRoot, err := filepath.EvalSymlinks(rootAbs); err == nil {
		rootAbs = resolvedRoot
	}
	var candidate string
	if filepath.IsAbs(raw) {
		candidate = filepath.Clean(raw)
	} else {
		candidate = filepath.Join(rootAbs, filepath.Clean(raw))
	}
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", 0, err
	}
	if err := requireInsideRoot(rootAbs, resolved); err != nil {
		return "", 0, err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", 0, err
	}
	if !info.Mode().IsRegular() {
		return "", 0, fmt.Errorf("file must be a regular file")
	}
	if maxBytes > 0 && info.Size() > maxBytes {
		return "", 0, fmt.Errorf("file too large: %d bytes exceeds limit %d", info.Size(), maxBytes)
	}
	return resolved, info.Size(), nil
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
