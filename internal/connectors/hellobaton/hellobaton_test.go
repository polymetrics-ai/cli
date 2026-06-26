package hellobaton_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/hellobaton"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts that the
// connector authenticates by injecting api_key as a query parameter, follows
// Hellobaton's DRF-style `next` URL pagination across two pages, extracts the
// `results` array, and maps each record. Red until the package exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	var pathHits []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.URL.Query().Get("api_key")
		pathHits = append(pathHits, r.URL.Path)
		if r.URL.Path != "/api/projects" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// next points at an absolute URL for page 2 (RequestPath style).
			_, _ = w.Write([]byte(`{"count":3,"next":"` + srv.URL + `/api/projects?page=2","previous":null,"results":[{"id":1,"name":"Alpha","created":"2026-01-01T00:00:00Z"},{"id":2,"name":"Beta","created":"2026-01-02T00:00:00Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"count":3,"next":null,"previous":null,"results":[{"id":3,"name":"Gamma","created":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"count":0,"next":null,"results":[]}`))
		}
	}))
	defer srv.Close()

	c := hellobaton.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/api"},
		Secrets: map[string]string{"api_key": "secret_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "secret_key_123" {
		t.Fatalf("api_key query = %q, want secret_key_123", sawAPIKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (across 2 pages); hits=%v", len(got), pathHits)
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
	if len(pathHits) < 2 {
		t.Fatalf("expected at least 2 page requests, got %v", pathHits)
	}
}

// TestFixtureModeNoNetwork verifies the deterministic fixture path emits records
// without any network call, so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := hellobaton.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "milestones", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := hellobaton.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams verifies the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := hellobaton.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"projects": false, "milestones": false, "tasks": false, "phases": false, "companies": false, "users": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Errorf("stream %q has no primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("catalog missing core stream %q", name)
		}
	}
}

// TestBaseURLValidation rejects SSRF-unsafe overrides.
func TestBaseURLValidation(t *testing.T) {
	c := hellobaton.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme error, got %v", err)
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = hellobaton.New() // ensure init ran
	c := hellobaton.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("hellobaton is read-only; Write should be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("hellobaton"); !ok {
		t.Fatal("registry did not resolve hellobaton (self-registration)")
	}
}
