package engine

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
)

// queryCaptureServer records the raw query string of every request so write
// query-parameter tests can assert on what actually reached the wire, which
// captureServer (write_test.go) deliberately does not capture.
func queryCaptureServer(t *testing.T) (*httptest.Server, *[]url.Values) {
	t.Helper()
	seen := &[]url.Values{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*seen = append(*seen, r.URL.Query())
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	return srv, seen
}

func writeOneRecord(t *testing.T, b Bundle, action string, cfg connectors.RuntimeConfig, rec connectors.Record) error {
	t.Helper()
	_, err := Write(context.Background(), b, connectors.WriteRequest{Action: action, Config: cfg}, []connectors.Record{rec}, nil)
	return err
}

// TestWriteActionQueryPlainString: a plain-string query entry resolves against
// config the same way stream.Query does, and an unresolved key is a hard error
// (the zero-migration-risk dialect, not a silent drop).
func TestWriteActionQueryPlainString(t *testing.T) {
	srv, seen := queryCaptureServer(t)
	b := newWriteTestBundle(srv, WriteAction{
		Name:   "update_widget",
		Kind:   "update",
		Method: http.MethodPost,
		Path:   "/widgets",
		Query:  map[string]QueryParam{"tenant": {Template: "{{ config.tenant }}"}},
	})
	cfg := connectors.RuntimeConfig{Config: map[string]string{"tenant": "acme-1"}}
	if err := writeOneRecord(t, b, "update_widget", cfg, connectors.Record{"id": "w1"}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if len(*seen) != 1 {
		t.Fatalf("want 1 request, got %d", len(*seen))
	}
	if got := (*seen)[0].Get("tenant"); got != "acme-1" {
		t.Fatalf("tenant query = %q, want %q", got, "acme-1")
	}

	// Unresolved config key must hard-error, matching the stream dialect.
	if err := writeOneRecord(t, b, "update_widget", connectors.RuntimeConfig{}, connectors.Record{"id": "w1"}); err == nil {
		t.Fatal("want hard error for unresolved config key in plain-string query entry, got nil")
	}
}

// TestWriteActionQueryOmitWhenAbsent: the object form drops the param instead
// of erroring when its key is absent.
func TestWriteActionQueryOmitWhenAbsent(t *testing.T) {
	srv, seen := queryCaptureServer(t)
	b := newWriteTestBundle(srv, WriteAction{
		Name:   "update_widget",
		Kind:   "update",
		Method: http.MethodPost,
		Path:   "/widgets",
		Query: map[string]QueryParam{
			"status": {Template: "{{ config.status }}", OmitWhenAbsent: true},
		},
	})
	if err := writeOneRecord(t, b, "update_widget", connectors.RuntimeConfig{}, connectors.Record{"id": "w1"}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, ok := (*seen)[0]["status"]; ok {
		t.Fatalf("status must be omitted entirely, got %v", (*seen)[0])
	}
}

// TestWriteActionQueryDefault: the object form sends the declared literal
// instead of hard-erroring when its key is absent.
func TestWriteActionQueryDefault(t *testing.T) {
	srv, seen := queryCaptureServer(t)
	b := newWriteTestBundle(srv, WriteAction{
		Name:   "update_widget",
		Kind:   "update",
		Method: http.MethodPost,
		Path:   "/widgets",
		Query: map[string]QueryParam{
			"page_size": {Template: "{{ config.page_size }}", Default: "50"},
		},
	})
	if err := writeOneRecord(t, b, "update_widget", connectors.RuntimeConfig{}, connectors.Record{"id": "w1"}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got := (*seen)[0].Get("page_size"); got != "50" {
		t.Fatalf("page_size = %q, want %q", got, "50")
	}
}

// TestWriteActionQueryFromRecordField: query templates resolve against the
// same Vars the path already interpolates from, so record fields are usable.
func TestWriteActionQueryFromRecordField(t *testing.T) {
	srv, seen := queryCaptureServer(t)
	b := newWriteTestBundle(srv, WriteAction{
		Name:   "update_widget",
		Kind:   "update",
		Method: http.MethodPost,
		Path:   "/widgets",
		Query:  map[string]QueryParam{"id": {Template: "{{ record.id }}"}},
	})
	if err := writeOneRecord(t, b, "update_widget", connectors.RuntimeConfig{}, connectors.Record{"id": "w1"}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got := (*seen)[0].Get("id"); got != "w1" {
		t.Fatalf("id = %q, want %q", got, "w1")
	}
}

func TestBuildWriteQueryOmitWhenAbsentScopesMissingRecordValuesToTheirDeclaredQuery(t *testing.T) {
	for _, tc := range []struct {
		name    string
		param   QueryParam
		vars    Vars
		want    string
		wantErr string
	}{
		{
			name:  "missing optional record value is omitted",
			param: QueryParam{Template: "{{ record.optional }}", OmitWhenAbsent: true},
			vars:  Vars{Config: map[string]string{"optional": "wrong-source"}, Record: map[string]any{"other": "present"}},
		},
		{
			name:    "missing required record value fails",
			param:   QueryParam{Template: "{{ record.required }}"},
			vars:    Vars{Config: map[string]string{"required": "wrong-source"}},
			wantErr: "unresolved key",
		},
		{
			name:  "explicit record value is preserved",
			param: QueryParam{Template: "{{ record.optional }}", OmitWhenAbsent: true},
			vars:  Vars{Record: map[string]any{"optional": "record-value"}},
			want:  "record-value",
		},
		{
			name:  "config omission remains unchanged",
			param: QueryParam{Template: "{{ config.optional }}", OmitWhenAbsent: true},
			vars:  Vars{},
		},
		{
			name:  "secret omission remains unchanged",
			param: QueryParam{Template: "{{ secrets.optional }}", OmitWhenAbsent: true},
			vars:  Vars{},
		},
		{
			name:  "incremental omission remains unchanged",
			param: QueryParam{Template: "{{ incremental.lower_bound }}", OmitWhenAbsent: true},
			vars:  Vars{},
		},
		{
			name:    "wrong source remains a failure",
			param:   QueryParam{Template: "{{ query.optional }}", OmitWhenAbsent: true},
			vars:    Vars{},
			wantErr: "does not permit query references",
		},
		{
			name:    "malformed explicit value remains a failure",
			param:   QueryParam{Template: "{{ record.optional }}", OmitWhenAbsent: true},
			vars:    Vars{Record: map[string]any{"optional": "bad\r\nvalue"}},
			wantErr: "contains CR/LF",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query, err := buildWriteQuery(WriteAction{
				Name:  "update_widget",
				Query: map[string]QueryParam{"optional": tc.param},
			}, tc.vars)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("buildWriteQuery error = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("buildWriteQuery: %v", err)
			}
			if got := query.Get("optional"); got != tc.want {
				t.Fatalf("optional query = %q, want %q", got, tc.want)
			}
			if tc.want == "" {
				if _, present := query["optional"]; present {
					t.Fatalf("optional query = %v, want parameter omitted", query)
				}
			}
		})
	}
}

func TestBuildWriteQueryPreflightsEveryExpressionBeforeRecordOmission(t *testing.T) {
	for _, tc := range []struct {
		name      string
		template  string
		vars      Vars
		want      string
		wantOmit  bool
		wantError string
	}{
		{
			name:      "later query reference",
			template:  "{{ record.optional }}{{ query.forbidden }}",
			vars:      Vars{Record: map[string]any{"id": "w1"}},
			wantError: "does not permit query references",
		},
		{
			name:      "reversed source order",
			template:  "{{ query.forbidden }}{{ record.optional }}",
			vars:      Vars{Record: map[string]any{"id": "w1"}},
			wantError: "does not permit query references",
		},
		{
			name:      "later invalid config expression",
			template:  "{{ record.optional }}{{ config.updated_at | unix_seconds }}",
			vars:      Vars{Config: map[string]string{"updated_at": "not-a-time"}, Record: map[string]any{"id": "w1"}},
			wantError: "invalid RFC3339 value",
		},
		{
			name:      "invalid filter after absent record",
			template:  "{{ record.optional | not-a-filter }}",
			vars:      Vars{Record: map[string]any{"id": "w1"}},
			wantError: `unknown filter "not-a-filter"`,
		},
		{
			name:      "malformed reference after absent record",
			template:  "{{ record.optional }}{{ config. }}",
			vars:      Vars{Record: map[string]any{"id": "w1"}},
			wantError: `malformed reference "config."`,
		},
		{
			name:      "unclosed expression after absent record",
			template:  "{{ record.optional }}{{ config.scope",
			vars:      Vars{Record: map[string]any{"id": "w1"}},
			wantError: "malformed template delimiter",
		},
		{
			name:     "mixed expression omits absent record",
			template: "prefix={{ record.optional }}&scope={{ config.scope }}",
			vars: Vars{
				Config: map[string]string{"scope": "workspace-1"},
				Record: map[string]any{"id": "w1"},
			},
			wantOmit: true,
		},
		{
			name:     "mixed expression materializes present record",
			template: "prefix={{ record.optional }}&scope={{ config.scope }}",
			vars: Vars{
				Config: map[string]string{"scope": "workspace-1"},
				Record: map[string]any{"optional": "record-value"},
			},
			want: "prefix=record-value&scope=workspace-1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query, err := buildWriteQuery(WriteAction{
				Name:  "update_widget",
				Query: map[string]QueryParam{"optional": {Template: tc.template, OmitWhenAbsent: true}},
			}, tc.vars)
			if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("buildWriteQuery error = %v, want %q", err, tc.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("buildWriteQuery: %v", err)
			}
			if tc.wantOmit {
				if _, present := query["optional"]; present {
					t.Fatalf("optional query = %#v, want parameter omitted", query)
				}
				return
			}
			if got := query.Get("optional"); got != tc.want {
				t.Fatalf("optional query = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildWriteQueryRejectsMalformedTemplatesAndReferenceTails(t *testing.T) {
	for _, tc := range []struct {
		name     string
		template string
		wantErr  string
	}{
		{
			name:     "nested delimiter cannot hide wrong source",
			template: "{{ record.optional {{ query.forbidden }}",
			wantErr:  "nested opening delimiter",
		},
		{
			name:     "stray closing delimiter",
			template: "{{ record.optional }} }}",
			wantErr:  "stray closing delimiter",
		},
		{
			name:     "unbalanced opening delimiter",
			template: "{{ record.optional",
			wantErr:  "unbalanced opening delimiter",
		},
		{
			name:     "empty expression",
			template: "{{ }}",
			wantErr:  "empty expression",
		},
		{
			name:     "empty filter stage",
			template: "{{ record.optional | }}",
			wantErr:  "malformed filter chain",
		},
		{
			name:     "config tail",
			template: "{{ config.mode.unexpected }}",
			wantErr:  "malformed reference",
		},
		{
			name:     "secret tail",
			template: "{{ secrets.token.unexpected }}",
			wantErr:  "malformed reference",
		},
		{
			name:     "query tail",
			template: "{{ query.value.unexpected }}",
			wantErr:  "malformed reference",
		},
		{
			name:     "incremental tail",
			template: "{{ incremental.lower_bound.unexpected }}",
			wantErr:  "malformed reference",
		},
		{
			name:     "fanout tail",
			template: "{{ fanout.id.unexpected }}",
			wantErr:  "malformed reference",
		},
		{
			name:     "cursor tail",
			template: "{{ cursor.unexpected }}",
			wantErr:  "malformed reference",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := buildWriteQuery(WriteAction{
				Name:  "update_widget",
				Query: map[string]QueryParam{"optional": {Template: tc.template, OmitWhenAbsent: true}},
			}, Vars{Record: map[string]any{"id": "w1"}})
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("buildWriteQuery error = %v, want %q", err, tc.wantErr)
			}
		})
	}

	query, err := buildWriteQuery(WriteAction{
		Name:  "update_widget",
		Query: map[string]QueryParam{"optional": {Template: "{{ record.settings.primary.id }}", OmitWhenAbsent: true}},
	}, Vars{Record: map[string]any{"settings": map[string]any{"primary": map[string]any{"id": "nested-id"}}}})
	if err != nil {
		t.Fatalf("buildWriteQuery dotted record path: %v", err)
	}
	if got := query.Get("optional"); got != "nested-id" {
		t.Fatalf("dotted record query = %q, want %q", got, "nested-id")
	}
}

func TestBuildWriteQuerySecretFilterErrorsDoNotExposeValues(t *testing.T) {
	for _, tc := range []struct {
		name      string
		template  string
		secret    string
		forbidden []string
	}{
		{
			name:      "unix seconds",
			template:  "{{ record.optional }}{{ secrets.token | unix_seconds }}",
			secret:    "secret-canary-unix",
			forbidden: []string{"secret-canary-unix"},
		},
		{
			name:      "base64 conversion",
			template:  "{{ record.optional }}{{ secrets.token | base64 | unix_seconds }}",
			secret:    "secret-canary-base64",
			forbidden: []string{"secret-canary-base64", base64.StdEncoding.EncodeToString([]byte("secret-canary-base64"))},
		},
		{
			name:      "urlencode conversion",
			template:  "{{ record.optional }}{{ secrets.token | urlencode | unix_seconds }}",
			secret:    "secret canary urlencode",
			forbidden: []string{"secret canary urlencode", urlencodeSegment("secret canary urlencode")},
		},
		{
			name:      "path segment conversion",
			template:  "{{ record.optional }}{{ secrets.token | last_path_segment | unix_seconds }}",
			secret:    "prefix/secret-canary-segment",
			forbidden: []string{"prefix/secret-canary-segment", "secret-canary-segment"},
		},
		{
			name:      "join type error",
			template:  "{{ record.optional }}{{ secrets.token | join:, }}",
			secret:    "secret-canary-join",
			forbidden: []string{"secret-canary-join"},
		},
		{
			name:      "raw value error",
			template:  "{{ record.optional }}{{ secrets.token }}",
			secret:    "secret-canary-crlf\r\n",
			forbidden: []string{"secret-canary-crlf"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := buildWriteQuery(WriteAction{
				Name:  "update_widget",
				Query: map[string]QueryParam{"optional": {Template: tc.template, OmitWhenAbsent: true}},
			}, Vars{
				Record:  map[string]any{"id": "w1"},
				Secrets: map[string]string{"token": tc.secret},
			})
			if err == nil {
				t.Fatal("buildWriteQuery error = nil, want secret-safe failure")
			}
			for _, value := range tc.forbidden {
				if strings.Contains(err.Error(), value) {
					t.Fatalf("buildWriteQuery error exposed secret value %q: %v", value, err)
				}
			}
		})
	}
}

func TestWriteActionRecordQueryRejectionsHappenBeforeProviderIO(t *testing.T) {
	for _, tc := range []struct {
		name    string
		param   QueryParam
		cfg     connectors.RuntimeConfig
		record  connectors.Record
		wantErr string
	}{
		{
			name:    "required record value",
			param:   QueryParam{Template: "{{ record.required }}"},
			record:  connectors.Record{"id": "w1"},
			wantErr: "unresolved key",
		},
		{
			name:    "undeclared record value cannot cross-bind",
			param:   QueryParam{Template: "{{ record.declared }}"},
			record:  connectors.Record{"id": "w1", "undeclared": "attempted-value"},
			wantErr: "unresolved key",
		},
		{
			name:    "wrong source",
			param:   QueryParam{Template: "{{ query.optional }}", OmitWhenAbsent: true},
			record:  connectors.Record{"id": "w1"},
			wantErr: "does not permit query references",
		},
		{
			name:    "missing record cannot hide later wrong source",
			param:   QueryParam{Template: "{{ record.optional }}{{ query.forbidden }}", OmitWhenAbsent: true},
			record:  connectors.Record{"id": "w1"},
			wantErr: "does not permit query references",
		},
		{
			name:    "nested delimiter cannot hide later wrong source",
			param:   QueryParam{Template: "{{ record.optional {{ query.forbidden }}", OmitWhenAbsent: true},
			record:  connectors.Record{"id": "w1"},
			wantErr: "nested opening delimiter",
		},
		{
			name:    "malformed explicit value",
			param:   QueryParam{Template: "{{ record.optional }}", OmitWhenAbsent: true},
			record:  connectors.Record{"id": "w1", "optional": "bad\r\nvalue"},
			wantErr: "contains CR/LF",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, seen := queryCaptureServer(t)
			bundle := newWriteTestBundle(srv, WriteAction{
				Name: "update_widget", Kind: "update", Method: http.MethodPost, Path: "/widgets",
				Query: map[string]QueryParam{"optional": tc.param},
			})
			err := writeOneRecord(t, bundle, "update_widget", tc.cfg, tc.record)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Write error = %v, want %q", err, tc.wantErr)
			}
			if len(*seen) != 0 {
				t.Fatalf("rejected query reached provider; requests = %d", len(*seen))
			}
		})
	}
}

func TestWriteActionSecretQueryErrorsDoNotReachProviderOrExposeValues(t *testing.T) {
	const secret = "secret-canary-provider"
	srv, seen := queryCaptureServer(t)
	bundle := newWriteTestBundle(srv, WriteAction{
		Name: "update_widget", Kind: "update", Method: http.MethodPost, Path: "/widgets",
		Query: map[string]QueryParam{
			"optional": {Template: "{{ record.optional }}{{ secrets.token | unix_seconds }}", OmitWhenAbsent: true},
		},
	})
	err := writeOneRecord(t, bundle, "update_widget", connectors.RuntimeConfig{Secrets: map[string]string{"token": secret}}, connectors.Record{"id": "w1"})
	if err == nil {
		t.Fatal("Write error = nil, want secret-safe failure")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("Write error exposed secret value: %v", err)
	}
	if len(*seen) != 0 {
		t.Fatalf("rejected query reached provider; requests = %d", len(*seen))
	}
}

func TestWriteActionOptionalRecordQueryIsOmittedOrPreservedAtProvider(t *testing.T) {
	for _, tc := range []struct {
		name   string
		record connectors.Record
		want   string
	}{
		{name: "missing optional record value is omitted", record: connectors.Record{"id": "w1"}},
		{name: "explicit record value is preserved", record: connectors.Record{"id": "w1", "optional": "record-value"}, want: "record-value"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, seen := queryCaptureServer(t)
			bundle := newWriteTestBundle(srv, WriteAction{
				Name: "update_widget", Kind: "update", Method: http.MethodPost, Path: "/widgets",
				Query: map[string]QueryParam{"optional": {Template: "{{ record.optional }}", OmitWhenAbsent: true}},
			})
			if err := writeOneRecord(t, bundle, "update_widget", connectors.RuntimeConfig{}, tc.record); err != nil {
				t.Fatalf("Write: %v", err)
			}
			if len(*seen) != 1 {
				t.Fatalf("provider requests = %d, want 1", len(*seen))
			}
			if got := (*seen)[0].Get("optional"); got != tc.want {
				t.Fatalf("optional query = %q, want %q", got, tc.want)
			}
			if tc.want == "" {
				if _, present := (*seen)[0]["optional"]; present {
					t.Fatalf("optional query = %v, want parameter omitted", (*seen)[0])
				}
			}
		})
	}
}

// TestWriteActionQueryAllBodyTypes: the resolved query must reach the wire for
// every body_type branch, not just the default JSON one. executeWriteRecord
// passed nil in all six branches before this capability existed.
func TestWriteActionQueryAllBodyTypes(t *testing.T) {
	for _, tc := range []struct {
		bodyType string
		action   WriteAction
		record   connectors.Record
	}{
		{"json", WriteAction{}, connectors.Record{"id": "w1"}},
		{"form", WriteAction{BodyType: "form"}, connectors.Record{"id": "w1"}},
		{"none", WriteAction{BodyType: "none"}, connectors.Record{"id": "w1"}},
		{
			"json_array",
			WriteAction{BodyType: "json_array", BodyField: "items"},
			connectors.Record{"id": "w1", "items": []any{map[string]any{"id": "w1"}}},
		},
		{"graphql", WriteAction{BodyType: "graphql", GraphQL: &GraphQLRequestSpec{
			Document: "mutation M { noop }", OperationName: "M",
		}}, connectors.Record{"id": "w1"}},
	} {
		t.Run(tc.bodyType, func(t *testing.T) {
			srv, seen := queryCaptureServer(t)
			action := tc.action
			action.Name = "update_widget"
			action.Kind = "update"
			action.Method = http.MethodPost
			action.Path = "/widgets"
			action.Query = map[string]QueryParam{"tenant": {Template: "{{ config.tenant }}"}}
			b := newWriteTestBundle(srv, action)
			cfg := connectors.RuntimeConfig{Config: map[string]string{"tenant": "acme-1"}}
			if err := writeOneRecord(t, b, "update_widget", cfg, tc.record); err != nil {
				t.Fatalf("Write: %v", err)
			}
			if len(*seen) != 1 {
				t.Fatalf("want 1 request, got %d", len(*seen))
			}
			if got := (*seen)[0].Get("tenant"); got != "acme-1" {
				t.Fatalf("body_type %s: tenant query = %q, want %q", tc.bodyType, got, "acme-1")
			}
		})
	}
}

// TestWriteActionQueryMultipartBodyType covers the sixth body_type branch,
// which needs a real file on disk and so cannot join the table above.
func TestWriteActionQueryMultipartBodyType(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/media.txt", []byte("hello media"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	var seen url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Query()
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			t.Errorf("drain multipart request body: %v", err)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	b := newWriteTestBundle(srv, WriteAction{
		Name:     "upload_media",
		Kind:     "update",
		Method:   http.MethodPut,
		Path:     "/media",
		BodyType: "multipart",
		Query:    map[string]QueryParam{"tenant": {Template: "{{ config.tenant }}"}},
		Multipart: &MultipartSpec{MaxBytes: 1024, Parts: []MultipartPartSpec{
			{Name: "mediaFile", Type: "file", Field: "media_file_path", ContentType: "text/plain", Required: true, MaxBytes: 1024},
		}},
	})
	cfg := connectors.RuntimeConfig{ProjectDir: dir, Config: map[string]string{"tenant": "acme-1"}}
	if err := writeOneRecord(t, b, "upload_media", cfg, connectors.Record{"media_file_path": "media.txt"}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got := seen.Get("tenant"); got != "acme-1" {
		t.Fatalf("multipart tenant query = %q, want %q", got, "acme-1")
	}
}

// TestWriteActionNoQueryUnchanged is the regression guard for the
// additive-and-opt-in invariant: an action declaring no query must send no
// query string at all, exactly as before this capability existed.
func TestWriteActionNoQueryUnchanged(t *testing.T) {
	srv, seen := queryCaptureServer(t)
	b := newWriteTestBundle(srv, WriteAction{
		Name:   "update_widget",
		Kind:   "update",
		Method: http.MethodPost,
		Path:   "/widgets",
	})
	cfg := connectors.RuntimeConfig{Config: map[string]string{"tenant": "acme-1"}}
	if err := writeOneRecord(t, b, "update_widget", cfg, connectors.Record{"id": "w1"}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if len((*seen)[0]) != 0 {
		t.Fatalf("want no query params, got %v", (*seen)[0])
	}
}

// TestWriteActionQueryParsesBothDialects: writes.json must accept the same two
// forms streams.json already accepts — bare string and object — via the
// existing QueryParam.UnmarshalJSON, with no second parser.
func TestWriteActionQueryParsesBothDialects(t *testing.T) {
	var action WriteAction
	raw := `{
		"name": "update_widget",
		"kind": "update",
		"method": "POST",
		"path": "/widgets",
		"record_schema": {"type": "object"},
		"risk": "low",
		"query": {
			"plain": "{{ config.a }}",
			"obj": {"template": "{{ config.b }}", "omit_when_absent": true, "default": "d"}
		}
	}`
	if err := json.Unmarshal([]byte(raw), &action); err != nil {
		t.Fatalf("unmarshal WriteAction: %v", err)
	}
	if got := action.Query["plain"]; got.Template != "{{ config.a }}" || got.OmitWhenAbsent || got.Default != "" {
		t.Fatalf("plain-string entry = %+v, want Template only", got)
	}
	obj := action.Query["obj"]
	if obj.Template != "{{ config.b }}" || !obj.OmitWhenAbsent || obj.Default != "d" {
		t.Fatalf("object entry = %+v, want all three fields", obj)
	}
}

// TestWritesSchemaAcceptsQuery pins the writes.json schema to accept the new
// optional query object, matching how streams.schema.json types the same
// construct.
func TestWritesSchemaAcceptsQuery(t *testing.T) {
	sch, err := CompileSchema(writesSchemaBytes(t))
	if err != nil {
		t.Fatalf("compile writes schema: %v", err)
	}
	doc := map[string]any{"actions": []any{map[string]any{
		"name":          "update_widget",
		"kind":          "update",
		"method":        "POST",
		"path":          "/widgets",
		"record_schema": map[string]any{"type": "object"},
		"risk":          "low",
		"query": map[string]any{
			"plain": "{{ config.a }}",
			"obj":   map[string]any{"template": "{{ config.b }}", "omit_when_absent": true},
		},
	}}}
	if err := sch.Validate(doc); err != nil {
		t.Fatalf("writes.schema.json must accept query: %v", err)
	}
}

func writesSchemaBytes(t *testing.T) json.RawMessage {
	t.Helper()
	if !strings.Contains(writesSchemaJSON, "\"query\"") {
		t.Fatalf("writes.schema.json does not declare a query property")
	}
	return json.RawMessage(writesSchemaJSON)
}

// TestBundleLoadWiresWriteQueryAndDynamicFields proves both new write-side
// capabilities survive the REAL bundle loader — meta-schema validation,
// decoding, and semantic validation — rather than only being reachable by
// constructing a WriteAction in Go. A schema field that the loader rejects, or
// silently drops, would be worse than no field at all.
func TestBundleLoadWiresWriteQueryAndDynamicFields(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/writes.json"] = &fstest.MapFile{Data: []byte(`{
		"actions": [{
			"name": "sync_member",
			"kind": "upsert",
			"method": "POST",
			"path": "/members",
			"risk": "low",
			"record_schema": {
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"id": {"type": "string"},
					"custom_fields": {"type": "object"}
				}
			},
			"query": {
				"plain": "{{ config.tenant }}",
				"obj": {"template": "{{ config.missing }}", "omit_when_absent": true}
			},
			"dynamic_fields": {
				"field": "custom_fields",
				"key_pattern": "^[A-Za-z][A-Za-z0-9_]*$",
				"max_keys": 25,
				"value_types": ["string", "number"],
				"max_value_bytes": 256,
				"target": "inline"
			}
		}]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load with query + dynamic_fields must succeed: %v", err)
	}
	if len(b.Writes) != 1 {
		t.Fatalf("Writes = %+v", b.Writes)
	}
	action := b.Writes[0]

	if got := action.Query["plain"]; got.Template != "{{ config.tenant }}" || got.OmitWhenAbsent {
		t.Fatalf("plain query entry lost in load: %+v", got)
	}
	if got := action.Query["obj"]; !got.OmitWhenAbsent {
		t.Fatalf("object query entry lost in load: %+v", got)
	}
	if action.DynamicFields == nil {
		t.Fatal("dynamic_fields dropped by the loader")
	}
	if action.DynamicFields.Field != "custom_fields" ||
		action.DynamicFields.MaxKeys != 25 ||
		action.DynamicFields.MaxValueBytes != 256 ||
		action.DynamicFields.Target != "inline" ||
		len(action.DynamicFields.ValueTypes) != 2 {
		t.Fatalf("dynamic_fields not fully decoded: %+v", action.DynamicFields)
	}
}

// TestBundleLoadRejectsInvalidDynamicFields proves the loader ENFORCES the
// declaration-time contract rather than merely accepting the field.
func TestBundleLoadRejectsInvalidDynamicFields(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/writes.json"] = &fstest.MapFile{Data: []byte(`{
		"actions": [{
			"name": "sync_member",
			"kind": "upsert",
			"method": "POST",
			"path": "/members",
			"risk": "low",
			"record_schema": {"type": "object"},
			"body_type": "form",
			"dynamic_fields": {"field": "custom_fields", "key_pattern": "^[A-Za-z]+$"}
		}]
	}`)}
	if _, err := Load(fsys, "acme"); err == nil {
		t.Fatal("dynamic_fields on an unsupported body_type must be rejected at load")
	}
}
