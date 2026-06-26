package deputy_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/deputy"
)

// TestReadAuthenticatesAndPaginates is the red-first test for the Deputy
// connector: Bearer auth on the Authorization header, offset pagination over a
// top-level JSON array (Deputy's resource endpoints return a bare array), and
// record mapping. Red until internal/connectors/deputy exists.
func TestReadAuthenticatesAndPaginates(t *testing.T) {
	var sawAuth string
	var sawAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawAccept = r.Header.Get("Accept")
		if r.URL.Path != "/api/v1/resource/Company" {
			http.NotFound(w, r)
			return
		}
		// Deputy offset pagination: ?start=N over a bare top-level array. The
		// connector reads the page size (here 2), then advances start by 2 until
		// a short page signals the end.
		start := r.URL.Query().Get("start")
		switch start {
		case "", "0":
			_, _ = w.Write([]byte(`[{"Id":1,"CompanyName":"HQ"},{"Id":2,"CompanyName":"Branch"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"Id":3,"CompanyName":"Remote"}]`))
		default:
			t.Errorf("unexpected start=%q", start)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := deputy.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "tok_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "locations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc123" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc123", sawAuth)
	}
	if sawAccept != "application/json" {
		t.Fatalf("Accept = %q, want application/json", sawAccept)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	// record mapping: source Id flattened to id, CompanyName -> company_name.
	first := got[0]
	if first["id"] == nil {
		t.Fatalf("record missing mapped id: %+v", first)
	}
	if got := stringify(first["id"]); got != "1" {
		t.Fatalf("mapped id = %q, want 1 (from source Id): %+v", got, first)
	}
	if first["company_name"] != "HQ" {
		t.Fatalf("company_name = %v, want HQ: %+v", first["company_name"], first)
	}
}

// stringify renders a decoded JSON scalar (json.Number, string, etc.) as a
// string for comparison in assertions.
func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprintf("%v", t)
	}
}

// TestEmployeesEndpoint confirms a non-resource stream (employees) hits its
// distinct path and maps records.
func TestEmployeesEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/supervise/employee" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"Id":10,"DisplayName":"Ada"},{"Id":11,"DisplayName":"Grace"}]`))
	}))
	defer srv.Close()

	c := deputy.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "tok"},
	}
	var n int
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "employees", Config: cfg}, func(rec connectors.Record) error {
		n++
		if rec["id"] == nil {
			t.Errorf("employee record missing id: %+v", rec)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Read employees: %v", err)
	}
	if n != 2 {
		t.Fatalf("employees = %d, want 2", n)
	}
}

// TestFixtureMode confirms credential-free fixture reads work for conformance.
func TestFixtureMode(t *testing.T) {
	c := deputy.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	var n int
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "locations", Config: cfg}, func(rec connectors.Record) error {
		n++
		if rec["id"] == nil {
			t.Errorf("fixture record missing id: %+v", rec)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if n == 0 {
		t.Fatal("fixture mode emitted no records")
	}
}

// TestCatalogAndMetadata confirms the published catalog and read-only metadata.
func TestCatalogAndMetadata(t *testing.T) {
	c := deputy.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
	if caps.Write {
		t.Fatal("deputy should be read-only (Write=false)")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Errorf("stream %q missing primary key", s.Name)
		}
	}
}

// TestRegistryResolution confirms self-registration via init() resolves through
// the central registry.
func TestRegistryResolution(t *testing.T) {
	_ = deputy.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("deputy"); !ok {
		t.Fatal("registry did not resolve deputy (self-registration)")
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs (SSRF guard) and requires
// a host.
func TestBaseURLValidation(t *testing.T) {
	c := deputy.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "locations", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http base_url")
	}
}
