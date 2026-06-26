package float_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/float"
)

// TestReadPaginatesAndAuthenticates is the red-first test: Bearer auth, Float
// page-number pagination over a top-level JSON array using the
// X-Pagination-Page-Count header, and record mapping for the people stream.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/people" {
			http.NotFound(w, r)
			return
		}
		// Float standard pagination: pages numbered from 1, total page count in
		// the X-Pagination-Page-Count header. Body is a top-level array.
		w.Header().Set("X-Pagination-Page-Count", "2")
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		switch page {
		case "1":
			_, _ = w.Write([]byte(`[{"people_id":1,"name":"Ada Lovelace","email":"ada@example.com"},{"people_id":2,"name":"Grace Hopper","email":"grace@example.com"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"people_id":3,"name":"Katherine Johnson","email":"katherine@example.com"}]`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := float.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"access_token": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "people", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["people_id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing people_id/name: %+v", rec)
		}
	}
}

func TestReadProjectsMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Pagination-Page-Count", "1")
		_, _ = w.Write([]byte(`[{"project_id":10,"name":"Apollo","client_id":5,"active":1,"budget_total":1000}]`))
	}))
	defer srv.Close()

	c := float.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"access_token": "tok_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["project_id"] == nil || got[0]["name"] != "Apollo" || got[0]["client_id"] == nil {
		t.Fatalf("project record not mapped: %+v", got[0])
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := float.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "people", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["people_id"] == nil {
		t.Fatalf("fixture record missing people_id: %+v", got[0])
	}
	// Check also short-circuits in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCheckRequiresToken(t *testing.T) {
	c := float.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check without access_token should fail")
	}
}

func TestBadBaseURLRejected(t *testing.T) {
	c := float.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"access_token": "tok_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "people", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("non-http base_url should be rejected (SSRF guard)")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := float.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "float" {
		t.Fatalf("catalog connector = %q, want float", cat.Connector)
	}
	want := map[string]bool{"people": false, "projects": false, "clients": false, "tasks": false, "departments": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestRegistryResolution(t *testing.T) {
	_ = float.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("float"); !ok {
		t.Fatal("registry did not resolve float (self-registration)")
	}
}

func TestMetadataReadOnly(t *testing.T) {
	c := float.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
	if caps.Write {
		t.Fatalf("float is read-only; Write should be false, got %+v", caps)
	}
}
