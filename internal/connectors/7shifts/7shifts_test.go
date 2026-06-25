package sevenshifts_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	sevenshifts "polymetrics/internal/connectors/7shifts"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the 7shifts
// connector: Bearer auth on the access_token, 7shifts cursor pagination
// (meta.cursor.next -> ?cursor=), record extraction at data[], and field
// mapping. Red until internal/connectors/7shifts is implemented.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawCompany string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v2/company/4242/users" {
			http.NotFound(w, r)
			return
		}
		sawCompany = "4242"
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":1,"first_name":"Ada","modified":"2026-01-01T00:00:00+00:00"},{"id":2,"first_name":"Grace","modified":"2026-01-02T00:00:00+00:00"}],"meta":{"cursor":{"current":"c0","next":"c1","count":2}}}`))
		case "c1":
			_, _ = w.Write([]byte(`{"data":[{"id":3,"first_name":"Katherine","modified":"2026-01-03T00:00:00+00:00"}],"meta":{"cursor":{"current":"c1","next":null,"count":1}}}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"data":[],"meta":{"cursor":{"next":null}}}`))
		}
	}))
	defer srv.Close()

	c := sevenshifts.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "company_id": "4242"},
		Secrets: map[string]string{"access_token": "tok_live_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_live_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_live_123", sawAuth)
	}
	if sawCompany != "4242" {
		t.Fatalf("expected request to company 4242 path")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["first_name"] == nil {
			t.Fatalf("record missing id/first_name: %+v", rec)
		}
	}
}

// TestReadCompaniesTopLevel exercises the one non-partitioned stream
// (/v2/companies) which needs no company_id.
func TestReadCompaniesTopLevel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/companies" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":4242,"name":"Acme Diner","modified":"2026-01-01T00:00:00+00:00"}],"meta":{"cursor":{"next":null}}}`))
	}))
	defer srv.Close()

	c := sevenshifts.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"access_token": "tok_live_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "companies", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read companies: %v", err)
	}
	if len(got) != 1 || got[0]["name"] != "Acme Diner" {
		t.Fatalf("companies = %+v, want one Acme Diner", got)
	}
}

// TestFixtureMode confirms the credential-free deterministic path works so the
// conformance harness can run without live creds or network.
func TestFixtureMode(t *testing.T) {
	c := sevenshifts.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "company_id": "4242"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "shifts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check must not require network in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams asserts the published catalog covers the core streams.
func TestCatalogStreams(t *testing.T) {
	c := sevenshifts.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"companies": false, "locations": false, "departments": false, "users": false, "shifts": false, "time_punches": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegistryResolution asserts self-registration via init() and resolution
// through the shared registry.
func TestRegistryResolution(t *testing.T) {
	_ = sevenshifts.New() // ensure package init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("7shifts")
	if !ok {
		t.Fatal("registry did not resolve 7shifts (self-registration)")
	}
	if got.Name() != "7shifts" {
		t.Fatalf("Name() = %q, want 7shifts", got.Name())
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
}
