package engine

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

// newWriteTestBundle builds a minimal Bundle wired against srv with a single
// write action, defaults filled in by the caller via action.
func newWriteTestBundle(srv *httptest.Server, action WriteAction) Bundle {
	if action.Name == "" {
		action.Name = "update_widget"
	}
	if action.Method == "" {
		action.Method = http.MethodPost
	}
	return Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Writes: []WriteAction{
			action,
		},
	}
}

func captureServer(t *testing.T, status int, body string) (*httptest.Server, *capturedRequest) {
	t.Helper()
	cap := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cap.method = r.Method
		cap.path = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		cap.body = raw
		cap.contentType = r.Header.Get("Content-Type")
		if status != 0 {
			w.WriteHeader(status)
		}
		if body != "" {
			_, _ = w.Write([]byte(body))
		}
	}))
	t.Cleanup(srv.Close)
	return srv, cap
}

type capturedRequest struct {
	method      string
	path        string
	body        []byte
	contentType string
}

func (c *capturedRequest) form() url.Values {
	v, _ := url.ParseQuery(string(c.body))
	return v
}

func (c *capturedRequest) json() map[string]any {
	var m map[string]any
	_ = json.Unmarshal(c.body, &m)
	return m
}

// --- body construction: json default (record minus path_fields) ---

func TestWriteJSONBodyDefaultExcludesPathFields(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{"ok":true}`)
	b := newWriteTestBundle(srv, WriteAction{
		Kind:       "update",
		Method:     http.MethodPatch,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
	})

	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "update_widget", Config: connectors.RuntimeConfig{}}, []connectors.Record{
		{"id": "42", "name": "new-name"},
	}, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("result = %+v", result)
	}
	if cap.method != http.MethodPatch {
		t.Fatalf("method = %q, want PATCH", cap.method)
	}
	if cap.path != "/widgets/42" {
		t.Fatalf("path = %q, want /widgets/42", cap.path)
	}
	got := cap.json()
	if _, ok := got["id"]; ok {
		t.Fatalf("body = %+v, id (a path_field) must not be in the body", got)
	}
	if got["name"] != "new-name" {
		t.Fatalf("body = %+v, want name=new-name", got)
	}
}

// --- body construction: form (stripe-shape) ---

func TestWriteFormBodyStripeShape(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{"id":"cus_1"}`)
	b := newWriteTestBundle(srv, WriteAction{
		Name:     "create_customer",
		Kind:     "create",
		Method:   http.MethodPost,
		Path:     "/customers",
		BodyType: "form",
		RecordSchema: json.RawMessage(`{
			"type": "object", "minProperties": 1,
			"properties": {
				"email": {"type": "string"}, "name": {"type": "string"},
				"description": {"type": "string"}, "phone": {"type": "string"}
			}
		}`),
	})

	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "create_customer"}, []connectors.Record{
		{"email": "ada@example.com", "name": "Ada Lovelace"},
	}, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if cap.contentType != "application/x-www-form-urlencoded" {
		t.Fatalf("content-type = %q, want form-urlencoded", cap.contentType)
	}
	form := cap.form()
	if form.Get("email") != "ada@example.com" || form.Get("name") != "Ada Lovelace" {
		t.Fatalf("form = %v", form)
	}
}

// --- body construction: none + body_fields (delete-with-body) ---

func TestWriteNoneBodyTypeSendsNoBody(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, "")
	b := newWriteTestBundle(srv, WriteAction{
		Kind:       "delete",
		Method:     http.MethodDelete,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
		BodyType:   "none",
	})

	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "update_widget"}, []connectors.Record{
		{"id": "42"},
	}, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if len(cap.body) != 0 {
		t.Fatalf("body = %q, want empty for body_type none", string(cap.body))
	}
}

func TestWriteBodyFieldsAllowListForDeleteWithBody(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, "")
	b := newWriteTestBundle(srv, WriteAction{
		Name:       "delete_file",
		Kind:       "delete",
		Method:     http.MethodDelete,
		Path:       "/files/{{ record.path }}",
		PathFields: []string{"path"},
		BodyFields: []string{"message", "sha", "branch"},
	})

	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "delete_file"}, []connectors.Record{
		{"path": "a.txt", "message": "remove file", "sha": "abc123", "branch": "main", "extra_untouched": "x"},
	}, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	got := cap.json()
	if got["message"] != "remove file" || got["sha"] != "abc123" || got["branch"] != "main" {
		t.Fatalf("body = %+v, want body_fields present", got)
	}
	if _, ok := got["extra_untouched"]; ok {
		t.Fatalf("body = %+v, want only body_fields present", got)
	}
	if _, ok := got["path"]; ok {
		t.Fatalf("body = %+v, path is a path_field and must not appear in body", got)
	}
}

// --- record_schema validation ---

func TestValidateWriteRecordSchemaErrorCarriesRecordIndex(t *testing.T) {
	b := Bundle{
		Name: "acme",
		Writes: []WriteAction{{
			Name:   "create_issue",
			Kind:   "create",
			Method: http.MethodPost,
			Path:   "/issues",
			RecordSchema: json.RawMessage(`{
				"type": "object", "required": ["title"],
				"properties": {"title": {"type": "string"}}
			}`),
		}},
	}

	err := ValidateWrite(context.Background(), b, connectors.WriteRequest{Action: "create_issue"}, []connectors.Record{
		{"title": "ok"},
		{"body": "missing title"},
	})
	if err == nil {
		t.Fatalf("ValidateWrite: want error for record 1 (0-indexed) missing required title")
	}
	if !strings.Contains(err.Error(), "1") {
		t.Fatalf("error = %q, want it to name the record index", err.Error())
	}
}

func TestValidateWriteRecordSchemaValidPasses(t *testing.T) {
	b := Bundle{
		Name: "acme",
		Writes: []WriteAction{{
			Name:   "create_issue",
			Kind:   "create",
			Method: http.MethodPost,
			Path:   "/issues",
			RecordSchema: json.RawMessage(`{
				"type": "object", "required": ["title"],
				"properties": {"title": {"type": "string"}}
			}`),
		}},
	}

	err := ValidateWrite(context.Background(), b, connectors.WriteRequest{Action: "create_issue"}, []connectors.Record{
		{"title": "ok"},
	})
	if err != nil {
		t.Fatalf("ValidateWrite: %v", err)
	}
}

// --- DryRunWrite ---

func TestDryRunWritePreviewResolvedMethodPathSecretsRedacted(t *testing.T) {
	b := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: "https://api.example.com"},
		Writes: []WriteAction{{
			Name:         "update_customer",
			Kind:         "update",
			Method:       http.MethodPost,
			Path:         "/customers/{{ record.id }}",
			PathFields:   []string{"id"},
			RecordSchema: json.RawMessage(`{"type": "object", "properties": {"id": {"type": "string"}, "name": {"type": "string"}}}`),
		}},
	}

	cfg := connectors.RuntimeConfig{Secrets: map[string]string{"client_secret": "sk_live_abc123"}}
	preview, err := DryRunWrite(context.Background(), b, connectors.WriteRequest{Action: "update_customer", Config: cfg}, []connectors.Record{
		{"id": "cus_1", "name": "New Name"},
	}, nil)
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	if preview.RecordsStaged != 1 {
		t.Fatalf("RecordsStaged = %d, want 1", preview.RecordsStaged)
	}
	if preview.Action != "update_customer" {
		t.Fatalf("Action = %q, want update_customer", preview.Action)
	}
	joined := strings.Join(preview.Warnings, " | ")
	if !strings.Contains(joined, "POST") || !strings.Contains(joined, "/customers/cus_1") {
		t.Fatalf("Warnings = %v, want resolved method+path", preview.Warnings)
	}
	if strings.Contains(joined, "sk_live_abc123") {
		t.Fatalf("Warnings = %v, secret leaked into preview", preview.Warnings)
	}
}

// --- delete semantics: missing_ok_status ---

func TestWriteDeleteMissingOkStatusCountsAsWritten(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Name:       "delete_label",
		Kind:       "delete",
		Method:     http.MethodDelete,
		Path:       "/labels/{{ record.name }}",
		PathFields: []string{"name"},
		Delete:     &DeleteSpec{Idempotent: true, MissingOkStatus: []int{404}},
	})

	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "delete_label"}, []connectors.Record{
		{"name": "bug"},
	}, nil)
	if err != nil {
		t.Fatalf("Write: %v (404 on idempotent delete should count as written, not error)", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("result = %+v, want 404 counted as written", result)
	}
}

func TestWriteDeleteNonListed404Fails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Name:       "delete_label",
		Kind:       "delete",
		Method:     http.MethodDelete,
		Path:       "/labels/{{ record.name }}",
		PathFields: []string{"name"},
		Delete:     &DeleteSpec{Idempotent: false},
	})

	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "delete_label"}, []connectors.Record{
		{"name": "bug"},
	}, nil)
	if err == nil {
		t.Fatalf("Write: want error for 404 not in missing_ok_status")
	}
	if result.RecordsWritten != 0 || result.RecordsFailed != 1 {
		t.Fatalf("result = %+v, want 0 written / 1 failed", result)
	}
}

func TestWriteNonListedStatusFails(t *testing.T) {
	// A non-retryable client error (400) so the test does not incur real
	// connsdk retry/backoff sleeps (500 would trigger 4 retries by default).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Name:       "delete_label",
		Kind:       "delete",
		Method:     http.MethodDelete,
		Path:       "/labels/{{ record.name }}",
		PathFields: []string{"name"},
		Delete:     &DeleteSpec{Idempotent: true, MissingOkStatus: []int{404}},
	})

	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "delete_label"}, []connectors.Record{
		{"name": "bug"},
	}, nil)
	if err == nil {
		t.Fatalf("Write: want error for 400 (not a missing_ok_status match)")
	}
	if result.RecordsWritten != 0 || result.RecordsFailed != 1 {
		t.Fatalf("result = %+v, want 0 written / 1 failed", result)
	}
}

// --- accounting parity with legacy semantics (stripe/write.go:66) ---

func TestWriteAccountingFailFastRemainderCountsAsFailed(t *testing.T) {
	// A non-retryable client error (400) on the second record so the test
	// does not incur real connsdk retry/backoff sleeps.
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Kind:       "update",
		Method:     http.MethodPost,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
	})

	records := []connectors.Record{
		{"id": "1"}, {"id": "2"}, {"id": "3"}, {"id": "4"},
	}
	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "update_widget"}, records, nil)
	if err == nil {
		t.Fatalf("Write: want error on second record's 400")
	}
	// Matches stripe/write.go:66 fail-fast semantics: RecordsWritten counts
	// successes up to the failure, RecordsFailed = len(records) - RecordsWritten.
	if result.RecordsWritten != 1 {
		t.Fatalf("RecordsWritten = %d, want 1", result.RecordsWritten)
	}
	if result.RecordsFailed != len(records)-result.RecordsWritten {
		t.Fatalf("RecordsFailed = %d, want %d", result.RecordsFailed, len(records)-result.RecordsWritten)
	}
}

func TestWriteValidationFailureReportsAllRecordsFailed(t *testing.T) {
	b := Bundle{
		Name: "acme",
		Writes: []WriteAction{{
			Name:   "create_issue",
			Kind:   "create",
			Method: http.MethodPost,
			Path:   "/issues",
			RecordSchema: json.RawMessage(`{
				"type": "object", "required": ["title"],
				"properties": {"title": {"type": "string"}}
			}`),
		}},
	}
	records := []connectors.Record{{"title": "ok"}, {"body": "no title"}}

	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "create_issue"}, records, nil)
	if err == nil {
		t.Fatalf("Write: want validation error")
	}
	if result.RecordsFailed != len(records) {
		t.Fatalf("RecordsFailed = %d, want %d (validation fails before any network call)", result.RecordsFailed, len(records))
	}
}

// --- ctx cancellation mid-loop ---

func TestWriteCtxCancelMidLoopAccounting(t *testing.T) {
	// A WriteHook lets this test cancel deterministically BETWEEN records
	// (ExecuteWrite runs once per record, before that record's declarative
	// body/request construction) rather than racing an HTTP round trip
	// against ctx cancellation — a real but separately-covered scenario
	// (TestReadCtxCancelMidPage covers the read-side "cancel while a request
	// is in flight" case). The hook always reports handled=false, so record 1
	// completes its real declarative Do() against srv; it then cancels ctx
	// while "handling" record 2 (before record 2's own HTTP call), so record
	// 2's request is itself refused by ctx and record 3 is never attempted.
	ctx, cancel := context.WithCancel(context.Background())
	srv, cap := captureServer(t, http.StatusOK, "")
	b := newWriteTestBundle(srv, WriteAction{
		Kind:       "update",
		Method:     http.MethodPost,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
	})

	h := &writeHookFunc{fn: func(_ context.Context, _ WriteAction, rec connectors.Record, _ *Runtime) (bool, error) {
		if rec["id"] == "2" {
			cancel()
		}
		return false, nil
	}}

	records := []connectors.Record{{"id": "1"}, {"id": "2"}, {"id": "3"}}
	result, err := Write(ctx, b, connectors.WriteRequest{Action: "update_widget"}, records, h)
	if err == nil {
		t.Fatalf("Write: want context.Canceled surfaced")
	}
	if result.RecordsWritten != 1 {
		t.Fatalf("RecordsWritten = %d, want 1 (record 1 completes; ctx is cancelled before record 2's own request is attempted)", result.RecordsWritten)
	}
	if result.RecordsFailed != len(records)-result.RecordsWritten {
		t.Fatalf("RecordsFailed = %d, want %d", result.RecordsFailed, len(records)-result.RecordsWritten)
	}
	if cap.path != "/widgets/1" {
		t.Fatalf("last observed request path = %q, want /widgets/1 (only record 1 ever reached the server)", cap.path)
	}
}

// --- WriteHook ---

func TestWriteHookHandledBypassesDeclarative(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("declarative HTTP request should not happen when WriteHook handles the write")
	}))
	t.Cleanup(srv.Close)
	b := newWriteTestBundle(srv, WriteAction{
		Name:   "merge_pull_request",
		Kind:   "custom",
		Method: http.MethodPut,
		Path:   "/pulls/{{ record.pull_number }}/merge",
	})

	calls := 0
	h := &writeHookFunc{fn: func(ctx context.Context, action WriteAction, rec connectors.Record, rt *Runtime) (bool, error) {
		calls++
		return true, nil
	}}

	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "merge_pull_request"}, []connectors.Record{
		{"pull_number": 7},
	}, h)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if calls != 1 {
		t.Fatalf("WriteHook called %d times, want 1", calls)
	}
	if result.RecordsWritten != 1 {
		t.Fatalf("result = %+v, want 1 written via hook", result)
	}
}

func TestWriteHookNotHandledFallsBackToDeclarative(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, "")
	b := newWriteTestBundle(srv, WriteAction{
		Kind:       "update",
		Method:     http.MethodPost,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
	})

	h := &writeHookFunc{fn: func(ctx context.Context, action WriteAction, rec connectors.Record, rt *Runtime) (bool, error) {
		return false, nil
	}}

	result, err := Write(context.Background(), b, connectors.WriteRequest{Action: "update_widget"}, []connectors.Record{
		{"id": "42", "name": "x"},
	}, h)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if result.RecordsWritten != 1 {
		t.Fatalf("result = %+v, want 1", result)
	}
	if cap.path != "/widgets/42" {
		t.Fatalf("path = %q, want declarative fallback to run the real request", cap.path)
	}
}

// --- form body with non-string field values ---

func TestWriteFormBodyStringifiesNonStringValues(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, "")
	b := newWriteTestBundle(srv, WriteAction{
		Name:       "update_widget_form",
		Kind:       "update",
		Method:     http.MethodPost,
		Path:       "/widgets/{{ record.id }}",
		PathFields: []string{"id"},
		BodyType:   "form",
	})

	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "update_widget_form"}, []connectors.Record{
		{"id": "42", "quantity": 3, "active": true},
	}, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	form := cap.form()
	if form.Get("quantity") != "3" {
		t.Fatalf("form.quantity = %q, want 3", form.Get("quantity"))
	}
	if form.Get("active") != "true" {
		t.Fatalf("form.active = %q, want true", form.Get("active"))
	}
}

// --- unknown action ---

func TestWriteUnknownActionErrors(t *testing.T) {
	b := Bundle{Name: "acme", Writes: []WriteAction{{Name: "known_action", Method: http.MethodPost, Path: "/x"}}}
	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: "does-not-exist"}, []connectors.Record{{}}, nil)
	if err == nil {
		t.Fatalf("Write: want error for unknown action")
	}
}

// --- test-only hook adapter ---

type writeHookFunc struct {
	fn func(ctx context.Context, action WriteAction, rec connectors.Record, rt *Runtime) (bool, error)
}

func (w *writeHookFunc) ConnectorName() string { return "write-hook-func-test" }
func (w *writeHookFunc) ExecuteWrite(ctx context.Context, action WriteAction, rec connectors.Record, rt *Runtime) (bool, error) {
	return w.fn(ctx, action, rec, rt)
}
