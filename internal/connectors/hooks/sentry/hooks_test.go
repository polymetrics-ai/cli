package sentry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

// newTestRuntime builds a minimal *engine.Runtime pointed at srv with a
// Bundle whose Schemas[streamName] projects exactly props — enough for
// ReadStream to run without needing a full loaded bundle.
func newTestRuntime(srv *httptest.Server, streamName string, props []string, cfg connectors.RuntimeConfig) *engine.Runtime {
	sch := schemaWithProperties(props)
	b := &engine.Bundle{
		Name:    "sentry",
		Schemas: map[string]*engine.StreamSchema{streamName: sch},
	}
	return &engine.Runtime{
		Requester: &connsdk.Requester{BaseURL: srv.URL},
		Bundle:    b,
		Config:    cfg,
	}
}

// schemaWithProperties builds a *engine.StreamSchema whose Properties()
// returns exactly props, using the same compiled-schema path production
// bundles use (engine.CompileSchema over a minimal draft-07 document) so
// this test exercises the real Properties() accessor, not a hand-rolled
// stand-in.
func schemaWithProperties(props []string) *engine.StreamSchema {
	doc := map[string]any{
		"$schema":    "http://json-schema.org/draft-07/schema#",
		"type":       "object",
		"properties": map[string]any{},
	}
	propsMap := doc["properties"].(map[string]any)
	for _, p := range props {
		propsMap[p] = map[string]any{"type": []string{"string", "integer", "boolean", "null"}}
	}
	sch, err := engine.CompileSchema(mustMarshal(doc))
	if err != nil {
		panic(err)
	}
	return &engine.StreamSchema{Schema: sch}
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func TestReadStream_LinkHeaderResultsFalseStopsAfterTwoPages(t *testing.T) {
	var paths []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path+"?"+r.URL.RawQuery)
		switch r.URL.Query().Get("cursor") {
		case "":
			w.Header().Set("Link", `<`+srv.URL+`/x?cursor=0:100:0>; rel="next"; results="true"`)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":"1"},{"id":"2"}]`))
		case "0:100:0":
			// Sentry ALWAYS emits rel="next", even on the truly-last page —
			// the twist this hook exists to handle. results="false" must
			// stop pagination without a 3rd request.
			w.Header().Set("Link", `<`+srv.URL+`/x?cursor=0:200:0>; rel="next"; results="false"`)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":"3"}]`))
		default:
			t.Errorf("unexpected 3rd request with cursor=%q", r.URL.Query().Get("cursor"))
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, "issues", []string{"id"}, connectors.RuntimeConfig{
		Config: map[string]string{"organization": "acme", "project": "backend"},
	})
	stream := engine.StreamSpec{Name: "issues"}

	var got []connectors.Record
	h := New()
	handled, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "issues", Config: rt.Config}, rt, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream handled = false, want true (StreamHook must handle every declared stream)")
	}
	if len(paths) != 2 {
		t.Fatalf("requests = %d, want exactly 2; paths=%v", len(paths), paths)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3; got=%+v", len(got), got)
	}
	gotIDs := []string{}
	for _, r := range got {
		id, _ := r["id"].(string)
		gotIDs = append(gotIDs, id)
	}
	if !reflect.DeepEqual(gotIDs, []string{"1", "2", "3"}) {
		t.Fatalf("record id sequence = %v, want [1 2 3]", gotIDs)
	}
}

func TestReadStream_NoLinkHeaderStopsAfterOnePage(t *testing.T) {
	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"10","slug":"backend"}]`))
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, "projects", []string{"id", "slug"}, connectors.RuntimeConfig{})
	stream := engine.StreamSpec{Name: "projects"}

	var got []connectors.Record
	h := New()
	handled, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "projects", Config: rt.Config}, rt, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream handled = false, want true")
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want exactly 1 (no Link header at all -> single page)", requests)
	}
	if len(got) != 1 || got[0]["slug"] != "backend" {
		t.Fatalf("records = %+v, want one backend project", got)
	}
}

func TestReadStream_ProjectionKeepsOnlySchemaProperties(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"10","slug":"backend","internal_secret_field":"should-not-leak"}]`))
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, "projects", []string{"id", "slug"}, connectors.RuntimeConfig{})
	stream := engine.StreamSpec{Name: "projects"}

	var got []connectors.Record
	h := New()
	if _, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "projects", Config: rt.Config}, rt, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	}); err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if _, ok := got[0]["internal_secret_field"]; ok {
		t.Fatalf("record = %+v, want internal_secret_field dropped (not a declared schema property)", got[0])
	}
	if got[0]["id"] != "10" || got[0]["slug"] != "backend" {
		t.Fatalf("record = %+v, want id=10 slug=backend", got[0])
	}
}

func TestReadStream_MissingOrganizationErrorsForScopedStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("no request should be sent when a required config value is missing")
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, "issues", []string{"id"}, connectors.RuntimeConfig{Config: map[string]string{"project": "backend"}})
	stream := engine.StreamSpec{Name: "issues"}

	handled, err := h().ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "issues", Config: rt.Config}, rt, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream did not error on missing organization config")
	}
	if !handled {
		t.Fatal("ReadStream handled = false, want true (an error is still a handled dispatch, not a declarative fallback)")
	}
}

func h() *Hooks { return New() }

func TestReadStream_UnknownStreamFallsBackToDeclarative(t *testing.T) {
	rt := &engine.Runtime{Bundle: &engine.Bundle{Name: "sentry", Schemas: map[string]*engine.StreamSchema{}}}
	stream := engine.StreamSpec{Name: "unknown_stream"}

	handled, err := New().ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "unknown_stream"}, rt, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("ReadStream handled = true for a stream with no known schema, want false (declarative fallback)")
	}
}

func TestConnectorNameAndRegistration(t *testing.T) {
	if New().ConnectorName() != "sentry" {
		t.Fatalf("ConnectorName() = %q, want %q", New().ConnectorName(), "sentry")
	}
	hooks := engine.HooksFor("sentry")
	if hooks == nil {
		t.Fatal("engine.HooksFor(\"sentry\") = nil, want a registered hook set (init() must call engine.RegisterHooks)")
	}
	if _, ok := hooks.(engine.StreamHook); !ok {
		t.Fatal("registered sentry hooks do not implement engine.StreamHook")
	}
}
