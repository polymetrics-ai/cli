package microsoftentraid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

// newTestRuntime builds a minimal *engine.Runtime pointed at srv with a
// Bundle whose Schemas[streamName] projects exactly props — enough for
// ReadStream to run without needing a full loaded bundle (mirrors sentry's
// hooks_test.go pattern).
func newTestRuntime(srv *httptest.Server, streamName string, props []string, cfg connectors.RuntimeConfig) *engine.Runtime {
	sch := schemaWithProperties(props)
	b := &engine.Bundle{
		Name:    "microsoft-entra-id",
		Schemas: map[string]*engine.StreamSchema{streamName: sch},
	}
	return &engine.Runtime{
		Requester: &connsdk.Requester{BaseURL: srv.URL},
		Bundle:    b,
		Config:    cfg,
	}
}

func schemaWithProperties(props []string) *engine.StreamSchema {
	doc := map[string]any{
		"$schema":    "http://json-schema.org/draft-07/schema#",
		"type":       "object",
		"properties": map[string]any{},
	}
	propsMap := doc["properties"].(map[string]any)
	for _, p := range props {
		propsMap[p] = map[string]any{"type": []string{"string", "boolean", "null"}}
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

func TestReadStream_NextLinkPaginationFollowsAbsoluteURL(t *testing.T) {
	var paths []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.String())
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("$skiptoken") {
		case "":
			_, _ = fmt.Fprintf(w, `{"value":[{"id":"1","displayName":"A"},{"id":"2","displayName":"B"}],"@odata.nextLink":"%s/users?$skiptoken=page2"}`, srv.URL)
		case "page2":
			_, _ = w.Write([]byte(`{"value":[{"id":"3","displayName":"C"}]}`))
		default:
			t.Errorf("unexpected 3rd request: %s", r.URL.String())
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, "users", []string{"id", "display_name"}, connectors.RuntimeConfig{})
	stream := engine.StreamSpec{Name: "users"}

	var got []connectors.Record
	h := New()
	handled, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "users", Config: rt.Config}, rt, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream handled = false, want true")
	}
	if len(paths) != 2 {
		t.Fatalf("requests = %d, want exactly 2; paths=%v", len(paths), paths)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3; got=%+v", len(got), got)
	}
	if got[0]["display_name"] != "A" || got[2]["display_name"] != "C" {
		t.Fatalf("record mapping wrong: %+v", got)
	}
}

func TestReadStream_NoNextLinkStopsAfterOnePage(t *testing.T) {
	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":[{"id":"10","displayName":"Backend"}]}`))
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, "groups", []string{"id", "display_name"}, connectors.RuntimeConfig{})
	stream := engine.StreamSpec{Name: "groups"}

	var got []connectors.Record
	h := New()
	handled, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "groups", Config: rt.Config}, rt, func(r connectors.Record) error {
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
		t.Fatalf("requests = %d, want exactly 1 (no @odata.nextLink -> single page)", requests)
	}
	if len(got) != 1 || got[0]["display_name"] != "Backend" {
		t.Fatalf("records = %+v, want one Backend group", got)
	}
}

func TestReadStream_ProjectionKeepsOnlySchemaProperties(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":[{"id":"10","displayName":"Backend","mail":"internal-secret-should-not-leak@example.com"}]}`))
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, "groups", []string{"id", "display_name"}, connectors.RuntimeConfig{})
	stream := engine.StreamSpec{Name: "groups"}

	var got []connectors.Record
	h := New()
	if _, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "groups", Config: rt.Config}, rt, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	}); err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if _, ok := got[0]["mail"]; ok {
		t.Fatalf("record = %+v, want mail dropped (not a declared schema property)", got[0])
	}
}

func TestReadStream_PageSizeSentOnFirstRequestOnly(t *testing.T) {
	var tops []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tops = append(tops, r.URL.Query().Get("$top"))
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("cursor2") == "" && len(tops) == 1 {
			_, _ = fmt.Fprintf(w, `{"value":[{"id":"1"}],"@odata.nextLink":"%s/users?cursor2=1"}`, srv.URL)
			return
		}
		_, _ = w.Write([]byte(`{"value":[{"id":"2"}]}`))
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, "users", []string{"id"}, connectors.RuntimeConfig{Config: map[string]string{"page_size": "50"}})
	stream := engine.StreamSpec{Name: "users"}

	h := New()
	if _, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "users", Config: rt.Config}, rt, func(connectors.Record) error { return nil }); err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(tops) != 2 {
		t.Fatalf("requests = %d, want 2", len(tops))
	}
	if tops[0] != "50" {
		t.Fatalf("first request $top = %q, want 50", tops[0])
	}
	if tops[1] != "" {
		t.Fatalf("second request $top = %q, want empty (nextLink carries its own params)", tops[1])
	}
}

func TestReadStream_InvalidPageSizeErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("no request should be sent when page_size is invalid")
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, "users", []string{"id"}, connectors.RuntimeConfig{Config: map[string]string{"page_size": "not-a-number"}})
	stream := engine.StreamSpec{Name: "users"}

	handled, err := New().ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "users", Config: rt.Config}, rt, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream did not error on invalid page_size")
	}
	if !handled {
		t.Fatal("ReadStream handled = false, want true (an error is still a handled dispatch)")
	}
}

func TestReadStream_UnknownStreamFallsBackToDeclarative(t *testing.T) {
	rt := &engine.Runtime{Bundle: &engine.Bundle{Name: "microsoft-entra-id", Schemas: map[string]*engine.StreamSchema{}}}
	stream := engine.StreamSpec{Name: "unknown_stream"}

	handled, err := New().ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "unknown_stream"}, rt, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("ReadStream handled = true for an unrecognized stream, want false (declarative fallback)")
	}
}

func TestConnectorNameAndRegistration(t *testing.T) {
	if New().ConnectorName() != "microsoft-entra-id" {
		t.Fatalf("ConnectorName() = %q, want %q", New().ConnectorName(), "microsoft-entra-id")
	}
	hooks := engine.HooksFor("microsoft-entra-id")
	if hooks == nil {
		t.Fatal("engine.HooksFor(\"microsoft-entra-id\") = nil, want a registered hook set")
	}
	if _, ok := hooks.(engine.StreamHook); !ok {
		t.Fatal("registered microsoft-entra-id hooks do not implement engine.StreamHook")
	}
}
