package everhour_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/everhour"
)

// TestReadProjectsAuthenticates is the red-first test for the Everhour
// connector: it asserts the X-Api-Key header is sent and that a top-level JSON
// array of projects is mapped into records.
func TestReadProjectsAuthenticates(t *testing.T) {
	var sawKey string
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("X-Api-Key")
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/projects" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"as:123","name":"Website","status":"open","type":"board"},{"id":"as:456","name":"Mobile","status":"open","type":"board"}]`))
	}))
	defer srv.Close()

	c := everhour.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "tok_secret_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "tok_secret_123" {
		t.Fatalf("X-Api-Key = %q, want tok_secret_123", sawKey)
	}
	if sawAuth != "" {
		t.Fatalf("Authorization header should not be set, got %q", sawAuth)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] != "as:123" || got[0]["name"] != "Website" {
		t.Fatalf("first record mapped wrong: %+v", got[0])
	}
}

// TestReadTasksSubstreamPaginates exercises the parent/child read: tasks are
// fetched per project, so reading tasks across two projects must hit two
// per-project endpoints and concatenate the results (pagination across pages).
func TestReadTasksSubstreamPaginates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") == "" {
			http.Error(w, "missing key", http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/projects":
			_, _ = w.Write([]byte(`[{"id":"as:1","name":"P1"},{"id":"as:2","name":"P2"}]`))
		case "/projects/as:1/tasks":
			_, _ = w.Write([]byte(`[{"id":"t1","name":"Task 1","status":"open"}]`))
		case "/projects/as:2/tasks":
			_, _ = w.Write([]byte(`[{"id":"t2","name":"Task 2","status":"open"},{"id":"t3","name":"Task 3","status":"completed"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := everhour.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "tok_secret_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tasks", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read tasks: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("task records = %d, want 3 (across 2 projects)", len(got))
	}
	// Each task should carry the parent project id for downstream joins.
	for _, rec := range got {
		if rec["project_id"] == nil {
			t.Fatalf("task record missing project_id: %+v", rec)
		}
	}
}

func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := everhour.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"projects", "clients", "users", "tasks"} {
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
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}

	// Check must not require network in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := everhour.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
	if caps.Write {
		t.Fatalf("everhour is read-only; Write should be false")
	}
}

func TestRegistryResolvesEverhour(t *testing.T) {
	_ = everhour.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("everhour"); !ok {
		t.Fatal("registry did not resolve everhour (self-registration)")
	}
}

func TestBaseURLValidatesScheme(t *testing.T) {
	c := everhour.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with non-http base_url should fail SSRF validation")
	}
}
