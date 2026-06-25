package greenhouse_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/greenhouse"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Greenhouse
// connector: HTTP Basic auth (API token as username, blank password),
// RFC-5988 Link-header pagination over a top-level JSON array, and record
// mapping. Red until internal/connectors/greenhouse exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/candidates" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			w.Header().Set("Link", "<"+srv.URL+"/candidates?page=2&per_page=2>; rel=\"next\"")
			_, _ = w.Write([]byte(`[{"id":1,"first_name":"Ada","last_name":"Lovelace"},{"id":2,"first_name":"Grace","last_name":"Hopper"}]`))
		case "2":
			// No Link header -> last page.
			_, _ = w.Write([]byte(`[{"id":3,"first_name":"Katherine","last_name":"Johnson"}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := greenhouse.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "gh_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "candidates", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	// gh_test_key with blank password, base64("gh_test_key:") = Z2hfdGVzdF9rZXk6
	if sawAuth != "Basic Z2hfdGVzdF9rZXk6" {
		t.Fatalf("Authorization = %q, want Basic Z2hfdGVzdF9rZXk6", sawAuth)
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

// TestReadFixtureNoNetwork verifies fixture mode emits deterministic records
// without any network access so credential-free conformance can run.
func TestReadFixtureNoNetwork(t *testing.T) {
	c := greenhouse.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "jobs", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil {
		t.Fatalf("fixture record missing id: %+v", got[0])
	}
}

// TestCheckFixtureModeNoNetwork verifies Check short-circuits in fixture mode.
func TestCheckFixtureModeNoNetwork(t *testing.T) {
	c := greenhouse.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams
// with primary keys.
func TestCatalogStreams(t *testing.T) {
	c := greenhouse.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"candidates": true, "applications": true, "jobs": true, "offers": true, "users": true}
	seen := map[string]bool{}
	for _, s := range cat.Streams {
		seen[s.Name] = true
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration and capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = greenhouse.New() // ensure init ran
	c := greenhouse.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("greenhouse is read-only, but Write capability is set")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("greenhouse"); !ok {
		t.Fatal("registry did not resolve greenhouse (self-registration)")
	}
}
