package microsoftlists

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

func newTestRuntime(srv *httptest.Server, streamName string, props []string, cfg connectors.RuntimeConfig) *engine.Runtime {
	sch := schemaWithProperties(props)
	b := &engine.Bundle{
		Name:    "microsoft-lists",
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
		propsMap[p] = map[string]any{"type": []string{"string", "boolean", "object", "null"}}
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
			_, _ = fmt.Fprintf(w, `{"value":[{"id":"1","name":"L1"},{"id":"2","name":"L2"}],"@odata.nextLink":"%s/sites/site1/lists?$skiptoken=page2"}`, srv.URL)
		case "page2":
			_, _ = w.Write([]byte(`{"value":[{"id":"3","name":"L3"}]}`))
		default:
			t.Errorf("unexpected 3rd request: %s", r.URL.String())
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, "lists", []string{"id", "name"}, connectors.RuntimeConfig{Config: map[string]string{"site_id": "site1"}})
	stream := engine.StreamSpec{Name: "lists"}

	var got []connectors.Record
	h := New()
	handled, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "lists", Config: rt.Config}, rt, func(r connectors.Record) error {
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
}

func TestReadStream_ListItemsSendsExpandFieldsAndRequiresListID(t *testing.T) {
	var gotExpand string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotExpand = r.URL.Query().Get("$expand")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":[{"id":"i1","fields":{"Title":"Item 1"}}]}`))
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, "list_items", []string{"id", "fields"}, connectors.RuntimeConfig{Config: map[string]string{"site_id": "site1", "list_id": "list1"}})
	stream := engine.StreamSpec{Name: "list_items"}

	var got []connectors.Record
	h := New()
	if _, err := h.ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "list_items", Config: rt.Config}, rt, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	}); err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if gotExpand != "fields" {
		t.Fatalf("$expand = %q, want fields", gotExpand)
	}
	if len(got) != 1 || got[0]["id"] != "i1" {
		t.Fatalf("records = %+v, want one i1 item", got)
	}
}

func TestReadStream_MissingListIDErrorsForScopedStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("no request should be sent when list_id is missing for a list-scoped stream")
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, "columns", []string{"id"}, connectors.RuntimeConfig{Config: map[string]string{"site_id": "site1"}})
	stream := engine.StreamSpec{Name: "columns"}

	handled, err := New().ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "columns", Config: rt.Config}, rt, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream did not error on missing list_id")
	}
	if !handled {
		t.Fatal("ReadStream handled = false, want true (an error is still a handled dispatch)")
	}
}

func TestReadStream_MissingSiteIDErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("no request should be sent when site_id is missing")
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, "lists", []string{"id"}, connectors.RuntimeConfig{})
	stream := engine.StreamSpec{Name: "lists"}

	handled, err := New().ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "lists", Config: rt.Config}, rt, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream did not error on missing site_id")
	}
	if !handled {
		t.Fatal("ReadStream handled = false, want true")
	}
}

func TestReadStream_ListsFlattenNestedListTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":[{"id":"l1","name":"L1","list":{"template":"genericList"}}]}`))
	}))
	defer srv.Close()

	rt := newTestRuntime(srv, "lists", []string{"id", "name", "list_template"}, connectors.RuntimeConfig{Config: map[string]string{"site_id": "site1"}})
	stream := engine.StreamSpec{Name: "lists"}

	var got []connectors.Record
	if _, err := New().ReadStream(context.Background(), stream, connectors.ReadRequest{Stream: "lists", Config: rt.Config}, rt, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	}); err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(got) != 1 || got[0]["list_template"] != "genericList" {
		t.Fatalf("records = %+v, want list_template=genericList", got)
	}
}

func TestReadStream_UnknownStreamFallsBackToDeclarative(t *testing.T) {
	rt := &engine.Runtime{Bundle: &engine.Bundle{Name: "microsoft-lists", Schemas: map[string]*engine.StreamSchema{}}}
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
	if New().ConnectorName() != "microsoft-lists" {
		t.Fatalf("ConnectorName() = %q, want %q", New().ConnectorName(), "microsoft-lists")
	}
	hooks := engine.HooksFor("microsoft-lists")
	if hooks == nil {
		t.Fatal("engine.HooksFor(\"microsoft-lists\") = nil, want a registered hook set")
	}
	if _, ok := hooks.(engine.StreamHook); !ok {
		t.Fatal("registered microsoft-lists hooks do not implement engine.StreamHook")
	}
}
