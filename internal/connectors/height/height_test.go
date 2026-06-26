package height_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/height"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Height
// connector: api-key auth header, nextPageToken pagination over the {"list":[...]}
// envelope, and record mapping. Red until internal/connectors/height exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/tasks" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("after") {
		case "":
			_, _ = w.Write([]byte(`{"list":[{"id":"task_1","name":"First","createdAt":"2026-01-01T00:00:00.000Z","status":"backlog"},{"id":"task_2","name":"Second","createdAt":"2026-01-02T00:00:00.000Z","status":"inProgress"}],"nextPageToken":"task_2"}`))
		case "task_2":
			_, _ = w.Write([]byte(`{"list":[{"id":"task_3","name":"Third","createdAt":"2026-01-03T00:00:00.000Z","status":"done"}],"nextPageToken":null}`))
		default:
			t.Errorf("unexpected after=%q", r.URL.Query().Get("after"))
			_, _ = w.Write([]byte(`{"list":[],"nextPageToken":null}`))
		}
	}))
	defer srv.Close()

	c := height.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "secret_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tasks", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "api-key secret_abc" {
		t.Fatalf("Authorization = %q, want %q", sawAuth, "api-key secret_abc")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["createdAt"] == nil {
			t.Fatalf("record missing id/createdAt: %+v", rec)
		}
	}
	if got[0]["name"] != "First" || got[2]["status"] != "done" {
		t.Fatalf("record mapping wrong: %+v", got)
	}
}

// TestReadWorkspaceSingleObject verifies the workspace stream, whose response is a
// single object at the root rather than a {"list":[...]} envelope.
func TestReadWorkspaceSingleObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspace" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"id":"ws_1","name":"Acme","createdAt":"2026-01-01T00:00:00.000Z","model":"workspace"}`))
	}))
	defer srv.Close()

	c := height.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "secret_abc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "workspace", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["id"] != "ws_1" || got[0]["name"] != "Acme" {
		t.Fatalf("workspace mapping wrong: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records with
// no network access so credential-free conformance passes.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := height.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"tasks", "lists", "field_templates", "users", "workspace"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("Read(%s) fixture emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture record for %s missing id: %+v", stream, rec)
			}
		}
	}
	// Check also short-circuits in fixture mode without creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCheckRequiresSecret(t *testing.T) {
	c := height.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check without api_key should fail")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := height.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "secret_abc"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tasks", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with non-http base_url should fail SSRF validation")
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := height.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("height is read-only, Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("expected at least 3 streams, got %d", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
}

func TestRegisteredViaRegistry(t *testing.T) {
	_ = height.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("height"); !ok {
		t.Fatal("registry did not resolve height (self-registration)")
	}
}
