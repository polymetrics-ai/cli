package uscensus

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func newRuntime(baseURL string) *engine.Runtime {
	return &engine.Runtime{Requester: &connsdk.Requester{BaseURL: baseURL}}
}

// --- registration ---

func TestInit_RegistersHooks(t *testing.T) {
	h := engine.HooksFor("us-census")
	if h == nil {
		t.Fatal("engine.HooksFor(\"us-census\") = nil, want a registered hook set (init() must call engine.RegisterHooks)")
	}
	if h.ConnectorName() != "us-census" {
		t.Fatalf("ConnectorName() = %q, want us-census", h.ConnectorName())
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered hooks do not implement StreamHook")
	}
}

// --- ReadStream: query path/params, header-row mapping ---

func TestReadStream_AddsAPIKeyAndMapsHeaderRows(t *testing.T) {
	var sawKey, sawPath, sawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("key")
		sawPath = r.URL.Path
		sawQuery = r.URL.RawQuery
		if r.URL.Path != "/data/2019/cbp" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[["NAME","ESTAB"],["United States","1"]]`))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{
		Stream: "query",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"query_path": "data/2019/cbp", "query_params": "get=NAME,ESTAB&for=us:*"},
			Secrets: map[string]string{"api_key": "dummy-key"},
		},
	}
	rt := &engine.Runtime{Requester: &connsdk.Requester{BaseURL: srv.URL, Auth: connsdk.APIKeyQuery("key", "dummy-key")}}

	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "query"}, req, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true (StreamHook always handles the query stream)")
	}
	if sawKey != "dummy-key" {
		t.Fatalf("key query = %q, want dummy-key", sawKey)
	}
	if sawPath != "/data/2019/cbp" {
		t.Fatalf("path = %q", sawPath)
	}
	if sawQuery == "" {
		t.Fatal("request carried no query string at all")
	}
	if len(got) != 1 || got[0]["name"] != "United States" || got[0]["estab"] != "1" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestReadStream_HeaderOnlyResponseEmitsNoRecords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[["NAME","ESTAB"]]`))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: map[string]string{"query_path": "data/2019/cbp", "query_params": "get=NAME,ESTAB"}}}
	rt := newRuntime(srv.URL)

	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "query"}, req, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if len(got) != 0 {
		t.Fatalf("got %d records, want 0 for a header-only response", len(got))
	}
}

func TestReadStream_EmptyHeaderCellIsSkippedPerColumn(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[["NAME","",  "ESTAB"],["United States","ignored","1"]]`))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: map[string]string{"query_path": "data/2019/cbp", "query_params": "get=NAME,,ESTAB"}}}
	rt := newRuntime(srv.URL)

	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "query"}, req, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if len(got) != 1 {
		t.Fatalf("got %d records, want 1", len(got))
	}
	if _, ok := got[0][""]; ok {
		t.Fatalf("record has an empty-string key, want the empty-header column skipped: %+v", got[0])
	}
	if got[0]["name"] != "United States" || got[0]["estab"] != "1" {
		t.Fatalf("unexpected record: %+v", got[0])
	}
}

func TestReadStream_ShortRowSkipsMissingColumns(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[["NAME","ESTAB"],["United States"]]`))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: map[string]string{"query_path": "data/2019/cbp", "query_params": "get=NAME,ESTAB"}}}
	rt := newRuntime(srv.URL)

	var got []connectors.Record
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "query"}, req, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(got) != 1 || got[0]["name"] != "United States" {
		t.Fatalf("unexpected records: %+v", got)
	}
	if _, ok := got[0]["estab"]; ok {
		t.Fatalf("record has estab set from a short row, want it absent: %+v", got[0])
	}
}

func TestReadStream_NumericCellsStringify(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[["NAME","ESTAB"],["United States",42]]`))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: map[string]string{"query_path": "data/2019/cbp", "query_params": "get=NAME,ESTAB"}}}
	rt := newRuntime(srv.URL)

	var got []connectors.Record
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "query"}, req, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(got) != 1 || got[0]["estab"] != "42" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

// --- error paths ---

func TestReadStream_MissingQueryPathIsError(t *testing.T) {
	h := Hooks{}
	req := connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: map[string]string{"query_params": "get=NAME"}}}
	rt := newRuntime("http://example.invalid")

	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "query"}, req, rt, func(connectors.Record) error { return nil })
	if !handled {
		t.Fatal("handled = false, want true (this is a real config error, not a fallback signal)")
	}
	if err == nil {
		t.Fatal("ReadStream error = nil, want an error naming the missing query_path")
	}
}

func TestReadStream_InvalidQueryParamsIsError(t *testing.T) {
	h := Hooks{}
	req := connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: map[string]string{"query_path": "data/2019/cbp", "query_params": "%zz"}}}
	rt := newRuntime("http://example.invalid")

	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "query"}, req, rt, func(connectors.Record) error { return nil })
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if err == nil {
		t.Fatal("ReadStream error = nil, want an error for invalid query_params")
	}
}

func TestReadStream_NonArrayResponseIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"not":"an array"}`))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: map[string]string{"query_path": "data/2019/cbp", "query_params": "get=NAME"}}}
	rt := newRuntime(srv.URL)

	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "query"}, req, rt, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream error = nil, want a decode error for a non-array response body")
	}
}

func TestReadStream_HonorsContextCancellation(t *testing.T) {
	h := Hooks{}
	req := connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: map[string]string{"query_path": "data/2019/cbp", "query_params": "get=NAME"}}}
	rt := newRuntime("http://example.invalid")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := h.ReadStream(ctx, engine.StreamSpec{Name: "query"}, req, rt, func(connectors.Record) error { return nil })
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("ReadStream(cancelled ctx) error = %v, want context.Canceled", err)
	}
}

func TestReadStream_EmitErrorPropagates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[["NAME","ESTAB"],["United States","1"],["Puerto Rico","2"]]`))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: map[string]string{"query_path": "data/2019/cbp", "query_params": "get=NAME,ESTAB"}}}
	rt := newRuntime(srv.URL)

	wantErr := errors.New("emit boom")
	calls := 0
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "query"}, req, rt, func(connectors.Record) error {
		calls++
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("ReadStream error = %v, want wantErr", err)
	}
	if calls != 1 {
		t.Fatalf("emit called %d times, want exactly 1 (stop on first error)", calls)
	}
}
