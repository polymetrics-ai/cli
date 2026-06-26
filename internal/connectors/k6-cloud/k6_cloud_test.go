package k6cloud_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	k6cloud "polymetrics.ai/internal/connectors/k6-cloud"
)

// TestReadTestsPaginatesAndAuthenticates is the red-first test for the k6-cloud
// connector: Bearer auth on Authorization, PageIncrement pagination over the
// k6-tests endpoint across two pages, and record mapping from the "k6-tests"
// field. Red until internal/connectors/k6-cloud is implemented.
func TestReadTestsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/loadtests/v2/tests" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// A full page (page_size records) signals there may be more.
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"k6-tests":[` + fullPage(1, 32) + `]}`))
		case "2":
			// A short page ends pagination.
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"k6-tests":[{"id":1001,"name":"checkout","project_id":7}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"k6-tests":[]}`))
		}
	}))
	defer srv.Close()

	c := k6cloud.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "32"},
		Secrets: map[string]string{"api_token": "tok_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "k6-tests", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc123" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc123", sawAuth)
	}
	// 32 records from page 1 + 1 record from page 2 = 33 across two pages.
	if len(got) != 33 {
		t.Fatalf("records = %d, want 33 (2 pages)", len(got))
	}
	last := got[len(got)-1]
	if last["id"] == nil || last["name"] != "checkout" {
		t.Fatalf("last record mapping wrong: %+v", last)
	}
}

// fullPage builds a JSON array body of n k6-test objects (without the surrounding
// brackets) so a page can hit exactly page_size and trigger a next-page fetch.
func fullPage(start, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			out += ","
		}
		id := strconv.Itoa(start*1000 + i)
		out += `{"id":` + id + `,"name":"test-` + id + `","project_id":1}`
	}
	return out
}

// TestReadProjectsTraversesOrganizations checks the projects substream: it first
// reads organizations, then fetches projects per organization id.
func TestReadProjectsTraversesOrganizations(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v3/organizations":
			_, _ = w.Write([]byte(`{"organizations":[{"id":11,"name":"acme"},{"id":12,"name":"globex"}]}`))
		case "/v3/organizations/11/projects":
			_, _ = w.Write([]byte(`{"projects":[{"id":101,"name":"p1","organization_id":11}]}`))
		case "/v3/organizations/12/projects":
			_, _ = w.Write([]byte(`{"projects":[{"id":102,"name":"p2","organization_id":12}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := k6cloud.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "tok"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read projects: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("projects = %d, want 2 (one per org)", len(got))
	}
	for _, r := range got {
		if r["organization_id"] == nil {
			t.Fatalf("project missing organization_id: %+v", r)
		}
		if r["id"] == nil {
			t.Fatalf("project missing id: %+v", r)
		}
	}
}

// TestReadOrganizations checks the top-level organizations stream maps records.
func TestReadOrganizations(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/organizations" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"organizations":[{"id":1,"name":"acme","billing_email":"b@acme.test"}]}`))
	}))
	defer srv.Close()

	c := k6cloud.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "tok"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organizations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read organizations: %v", err)
	}
	if len(got) != 1 || got[0]["name"] != "acme" || got[0]["billing_email"] != "b@acme.test" {
		t.Fatalf("organization mapping wrong: %+v", got)
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access and no credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := k6cloud.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"organizations", "projects", "k6-tests"} {
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		}); err != nil {
			t.Fatalf("fixture Read %s: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s emitted no records", stream)
		}
		for _, r := range got {
			if r["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, r)
			}
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogAndMetadata checks the published catalog and read-only capabilities.
func TestCatalogAndMetadata(t *testing.T) {
	c := k6cloud.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("k6-cloud is read-only; Write must be false: %+v", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"organizations": true, "projects": true, "k6-tests": true}
	for _, s := range cat.Streams {
		delete(want, s.Name)
	}
	if len(want) != 0 {
		t.Fatalf("catalog missing streams: %v", want)
	}
}

// TestRegistryResolvesK6Cloud confirms self-registration resolves via the registry.
func TestRegistryResolvesK6Cloud(t *testing.T) {
	_ = k6cloud.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("k6-cloud"); !ok {
		t.Fatal("registry did not resolve k6-cloud (self-registration)")
	}
}

// TestBaseURLValidation rejects non-http(s) base_url overrides (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := k6cloud.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organizations", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}
