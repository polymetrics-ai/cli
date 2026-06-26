package nebiusai_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	nebiusai "polymetrics.ai/internal/connectors/nebius-ai"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Nebius AI
// connector: Bearer auth, OpenAI-compatible has_more/after pagination over
// data[], and record mapping across two pages.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v1/files" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("after") {
		case "":
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"file-1","created_at":1700000000,"filename":"a.jsonl","bytes":10},{"id":"file-2","created_at":1700000100,"filename":"b.jsonl","bytes":20}],"has_more":true}`))
		case "file-2":
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"file-3","created_at":1700000200,"filename":"c.jsonl","bytes":30}],"has_more":false}`))
		default:
			t.Errorf("unexpected after=%q", r.URL.Query().Get("after"))
			_, _ = w.Write([]byte(`{"object":"list","data":[],"has_more":false}`))
		}
	}))
	defer srv.Close()

	c := nebiusai.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "nb_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "files", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer nb_test_123" {
		t.Fatalf("Authorization = %q, want Bearer nb_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
	if got[0]["filename"] != "a.jsonl" {
		t.Fatalf("record mapping wrong, filename = %v, want a.jsonl", got[0]["filename"])
	}
}

func TestReadModelsMapsRecords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"meta-llama/Llama-3.1-8B","object":"model","created":1700000000,"owned_by":"nebius"}]}`))
	}))
	defer srv.Close()

	c := nebiusai.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "nb_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "models", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["id"] != "meta-llama/Llama-3.1-8B" || got[0]["owned_by"] != "nebius" {
		t.Fatalf("model record mapping wrong: %+v", got[0])
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := nebiusai.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"models", "files", "batches"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) produced no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture record missing id: %+v", rec)
			}
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := nebiusai.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "nebius-ai" {
		t.Fatalf("catalog connector = %q, want nebius-ai", cat.Connector)
	}
	want := map[string]bool{"models": false, "files": false, "batches": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = nebiusai.New() // ensure init ran
	c := nebiusai.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities Write = true, want false (read-only connector)")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("nebius-ai"); !ok {
		t.Fatal("registry did not resolve nebius-ai (self-registration)")
	}
}
