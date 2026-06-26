package factorial_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/factorial"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Factorial
// connector: X-API-KEY header auth, page-increment pagination over data[] (two
// pages), and record mapping. Red until internal/connectors/factorial exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	var sawAuthHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("X-API-KEY")
		sawAuthHeader = r.Header.Get("Authorization")
		if r.URL.Path != "/employees/employees" {
			http.NotFound(w, r)
			return
		}
		// PageIncrement starts at page 1; the server keys pages off ?page=.
		switch r.URL.Query().Get("page") {
		case "1", "":
			// Full page (limit=2) -> the paginator must request the next page.
			_, _ = w.Write([]byte(`{"data":[{"id":1,"first_name":"Ada","last_name":"Lovelace","email":"ada@example.com","updated_at":"2026-01-02T00:00:00.000Z"},{"id":2,"first_name":"Grace","last_name":"Hopper","email":"grace@example.com","updated_at":"2026-01-03T00:00:00.000Z"}]}`))
		case "2":
			// Short page (1 < limit) -> pagination stops here.
			_, _ = w.Write([]byte(`{"data":[{"id":3,"first_name":"Katherine","last_name":"Johnson","email":"kj@example.com","updated_at":"2026-01-04T00:00:00.000Z"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := factorial.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "limit": "2"},
		Secrets: map[string]string{"api_key": "fac_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "employees", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "fac_test_123" {
		t.Fatalf("X-API-KEY = %q, want fac_test_123", sawKey)
	}
	if sawAuthHeader != "" {
		t.Fatalf("Authorization header = %q, want empty (api-key header auth only)", sawAuthHeader)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
	if got[0]["full_name"] != "Ada Lovelace" {
		t.Fatalf("full_name = %v, want \"Ada Lovelace\"", got[0]["full_name"])
	}
}

// TestCheckHitsCheckStream verifies Check performs a bounded live read against
// the check stream (public_credentials) with the api key applied.
func TestCheckHitsCheckStream(t *testing.T) {
	var sawKey, sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("X-API-KEY")
		sawPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	c := factorial.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "fac_test_123"},
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if sawKey != "fac_test_123" {
		t.Fatalf("X-API-KEY = %q, want fac_test_123", sawKey)
	}
	if sawPath != "/api_public/credentials" {
		t.Fatalf("check path = %q, want /api_public/credentials", sawPath)
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access, so credential-free conformance works.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := factorial.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	// Check must short-circuit with no creds and no base_url.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}

	for _, stream := range []string{"employees", "teams", "leaves", "leave_types", "locations"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
		if got[0]["id"] == nil {
			t.Fatalf("fixture %s record missing id: %+v", stream, got[0])
		}
	}
}

// TestCatalogStreams verifies the published catalog and that every stream has a
// primary key.
func TestCatalogStreams(t *testing.T) {
	c := factorial.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "factorial" {
		t.Fatalf("connector = %q, want factorial", cat.Connector)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q missing fields", s.Name)
		}
	}
}

// TestBaseURLValidation rejects non-http(s) schemes to bound SSRF.
func TestBaseURLValidation(t *testing.T) {
	c := factorial.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "employees", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for invalid base_url scheme")
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = factorial.New() // ensure init ran
	c := factorial.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write == false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("factorial"); !ok {
		t.Fatal("registry did not resolve factorial (self-registration)")
	}
}
