package northpasslms_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	northpasslms "polymetrics.ai/internal/connectors/northpass-lms"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Northpass LMS
// connector: X-Api-Key auth, JSON:API data[] extraction, links.next pagination
// across two pages, and record mapping (flattening attributes). Red until the
// package exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("X-Api-Key")
		if r.URL.Path != "/v2/courses" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page") {
		case "", "1":
			// First page advertises a next link to page 2.
			_, _ = w.Write([]byte(`{
				"data":[
					{"id":"c1","type":"courses","attributes":{"name":"Intro","slug":"intro","status":"published"}},
					{"id":"c2","type":"courses","attributes":{"name":"Advanced","slug":"adv","status":"draft"}}
				],
				"links":{"self":"` + srvURL + `/v2/courses?page=1","next":"` + srvURL + `/v2/courses?page=2"}
			}`))
		case "2":
			// Final page has no next link.
			_, _ = w.Write([]byte(`{
				"data":[
					{"id":"c3","type":"courses","attributes":{"name":"Expert","slug":"exp","status":"published"}}
				],
				"links":{"self":"` + srvURL + `/v2/courses?page=2"}
			}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"data":[],"links":{}}`))
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := northpasslms.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/v2"},
		Secrets: map[string]string{"api_key": "np_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "courses", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "np_test_123" {
		t.Fatalf("X-Api-Key = %q, want np_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	// Record mapping: id must be present and attributes flattened to top level.
	first := got[0]
	if first["id"] != "c1" {
		t.Fatalf("first record id = %v, want c1", first["id"])
	}
	if first["name"] != "Intro" {
		t.Fatalf("first record name = %v, want Intro (attributes must be flattened)", first["name"])
	}
	if first["type"] != "courses" {
		t.Fatalf("first record type = %v, want courses", first["type"])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network call, so conformance runs without live credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := northpasslms.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"people", "courses", "course_enrollments", "groups"} {
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
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

// TestCheckFixtureAndMissingSecret covers the credential-free check path and the
// missing-secret error.
func TestCheckFixtureAndMissingSecret(t *testing.T) {
	c := northpasslms.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	err := c.Check(context.Background(), connectors.RuntimeConfig{})
	if err == nil {
		t.Fatal("Check without api_key should error")
	}
}

// TestCatalogAndMetadata asserts the published catalog and read-only capability.
func TestCatalogAndMetadata(t *testing.T) {
	c := northpasslms.New()
	meta := c.Metadata()
	if !meta.Capabilities.Read || meta.Capabilities.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", meta.Capabilities)
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
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
}

// TestRegistryResolves confirms self-registration via the global registry.
func TestRegistryResolves(t *testing.T) {
	_ = northpasslms.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("northpass-lms"); !ok {
		t.Fatal("registry did not resolve northpass-lms (self-registration)")
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF validation on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := northpasslms.New()
	err := c.Check(context.Background(), connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "x"},
	})
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Check with ftp base_url err = %v, want base_url scheme error", err)
	}
}
