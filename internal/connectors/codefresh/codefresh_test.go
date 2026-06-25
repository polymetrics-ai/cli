package codefresh_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/codefresh"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Codefresh
// connector: API-key Authorization header (raw token, no Bearer prefix),
// page-based pagination over the projects list ({projects:[...]}), and record
// mapping. Red until internal/connectors/codefresh exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawAccount string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawAccount = r.Header.Get("X-Access-Token")
		if r.URL.Path != "/projects" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`{"projects":[{"id":"prj_1","projectName":"alpha"},{"id":"prj_2","projectName":"beta"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"projects":[{"id":"prj_3","projectName":"gamma"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"projects":[]}`))
		}
	}))
	defer srv.Close()

	c := codefresh.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "tok_test_123", "account_id": "acct_42"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "tok_test_123" {
		t.Fatalf("Authorization = %q, want raw token tok_test_123", sawAuth)
	}
	if sawAccount != "acct_42" {
		t.Fatalf("X-Access-Token = %q, want acct_42", sawAccount)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

// TestReadRootArrayStream covers a stream whose payload is a bare top-level
// JSON array (agents), confirming the records-path "" extraction and a short
// final page stopping pagination.
func TestReadRootArrayStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents" {
			http.NotFound(w, r)
			return
		}
		// One short page -> stops immediately.
		_, _ = w.Write([]byte(`[{"id":"ag_1","name":"runner-1"}]`))
	}))
	defer srv.Close()

	c := codefresh.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "tok", "account_id": "acct"},
	}
	var n int
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "agents", Config: cfg}, func(rec connectors.Record) error {
		n++
		if rec["id"] == nil {
			t.Fatalf("agent record missing id: %+v", rec)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Read agents: %v", err)
	}
	if n != 1 {
		t.Fatalf("agents = %d, want 1", n)
	}
}

func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := codefresh.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	for _, stream := range []string{"projects", "pipelines", "agents", "contexts"} {
		got = got[:0]
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
		if got[0]["id"] == nil {
			t.Fatalf("fixture %s record missing id: %+v", stream, got[0])
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogAndRegistry(t *testing.T) {
	c := codefresh.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("codefresh is read-only, Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}

	r := connectors.NewRegistry()
	if _, ok := r.Get("codefresh"); !ok {
		t.Fatal("registry did not resolve codefresh (self-registration)")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := codefresh.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "tok", "account_id": "acct"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http base_url scheme")
	}
}
