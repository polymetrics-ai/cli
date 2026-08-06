package engine

import (
	"context"
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
