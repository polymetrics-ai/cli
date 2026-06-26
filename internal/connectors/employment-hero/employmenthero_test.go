package employmenthero_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	employmenthero "polymetrics.ai/internal/connectors/employment-hero"
)

// TestReadOrganisationsPaginatesAndAuthenticates is the red-first test: it
// asserts the Bearer auth header, Employment Hero page_index/items_per_page
// pagination across two pages of the data.items envelope, and record mapping.
// Red until internal/connectors/employmenthero exists.
func TestReadOrganisationsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPaths = append(sawPaths, r.URL.Path+"?"+r.URL.RawQuery)
		if r.URL.Path != "/organisations" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page_index") {
		case "", "1":
			_, _ = w.Write([]byte(`{"data":{"items":[{"id":"org_1","name":"Acme"},{"id":"org_2","name":"Globex"}],"page_index":1,"item_per_page":2,"total_items":3,"total_pages":2}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":{"items":[{"id":"org_3","name":"Initech"}],"page_index":2,"item_per_page":2,"total_items":3,"total_pages":2}}`))
		default:
			t.Errorf("unexpected page_index=%q", r.URL.Query().Get("page_index"))
			_, _ = w.Write([]byte(`{"data":{"items":[],"total_pages":2}}`))
		}
	}))
	defer srv.Close()

	c := employmenthero.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "items_per_page": "2"},
		Secrets: map[string]string{"api_key": "eh_secret_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organisations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer eh_secret_123" {
		t.Fatalf("Authorization = %q, want Bearer eh_secret_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages; paths=%v", len(got), sawPaths)
	}
	if got[0]["id"] != "org_1" || got[2]["id"] != "org_3" {
		t.Fatalf("unexpected record ids: %v", got)
	}
	if len(sawPaths) != 2 {
		t.Fatalf("requested %d pages, want 2: %v", len(sawPaths), sawPaths)
	}
}

// TestReadEmployeesSubstreamUsesOrganizationID verifies substreams resolve the
// org id from config and hit the org-scoped path.
func TestReadEmployeesSubstreamUsesOrganizationID(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"items":[{"id":"emp_1","first_name":"Ada","last_name":"Lovelace"}],"total_pages":1}}`))
	}))
	defer srv.Close()

	c := employmenthero.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "organization_id": "org_42"},
		Secrets: map[string]string{"api_key": "eh_secret_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "employees", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawPath != "/organisations/org_42/employees" {
		t.Fatalf("path = %q, want /organisations/org_42/employees", sawPath)
	}
	if len(got) != 1 || got[0]["id"] != "emp_1" {
		t.Fatalf("unexpected records: %v", got)
	}
}

// TestEmployeesSubstreamRequiresOrgID asserts a missing org id is a clear error.
func TestEmployeesSubstreamRequiresOrgID(t *testing.T) {
	c := employmenthero.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "https://api.employmenthero.com/api/v1"},
		Secrets: map[string]string{"api_key": "eh_secret_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "employees", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error when employees substream lacks an organization id")
	}
}

// TestFixtureModeIsCredentialFree confirms conformance can run without creds.
func TestFixtureModeIsCredentialFree(t *testing.T) {
	c := employmenthero.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	for _, stream := range []string{"organisations", "employees", "leave_requests", "teams"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(fixture, %s) = %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s produced no records", stream)
		}
		if got[0]["id"] == nil {
			t.Fatalf("fixture record for %s missing id: %v", stream, got[0])
		}
	}
}

// TestCatalogStreams checks the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := employmenthero.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"organisations": false, "employees": false, "leave_requests": false, "teams": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly verifies self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	c := employmenthero.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatal("employment-hero is read-only; Write must be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("employment-hero"); !ok {
		t.Fatal("registry did not resolve employment-hero (self-registration)")
	}
}

// TestBaseURLRejectsBadScheme guards SSRF on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := employmenthero.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organisations", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http base_url scheme")
	}
}
