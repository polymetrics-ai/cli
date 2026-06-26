package docuseal_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/docuseal"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the DocuSeal
// connector: X-Auth-Token header auth, cursor pagination over data[] driven by
// pagination.next/after, and record mapping. Red until
// internal/connectors/docuseal exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("X-Auth-Token")
		sawAccept = r.Header.Get("Accept")
		if r.URL.Path != "/templates" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("after") {
		case "":
			// First page: two records, pagination.next points at the last id so
			// the connector requests the next page with after=2.
			_, _ = w.Write([]byte(`{"data":[{"id":1,"name":"NDA","slug":"abc","created_at":"2026-01-01T00:00:00Z"},{"id":2,"name":"MSA","slug":"def","created_at":"2026-01-02T00:00:00Z"}],"pagination":{"count":2,"next":2,"prev":1}}`))
		case "2":
			// Second page: one record, pagination.next is null -> stop.
			_, _ = w.Write([]byte(`{"data":[{"id":3,"name":"SOW","slug":"ghi","created_at":"2026-01-03T00:00:00Z"}],"pagination":{"count":1,"next":null,"prev":3}}`))
		default:
			t.Errorf("unexpected after=%q", r.URL.Query().Get("after"))
			_, _ = w.Write([]byte(`{"data":[],"pagination":{"count":0,"next":null,"prev":null}}`))
		}
	}))
	defer srv.Close()

	c := docuseal.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "templates", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "tok_test_123" {
		t.Fatalf("X-Auth-Token = %q, want tok_test_123", sawAuth)
	}
	if sawAccept != "application/json" {
		t.Fatalf("Accept = %q, want application/json", sawAccept)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestReadSubmissionsMapping confirms the submissions mapper flattens the
// nested template object id and core scalar fields.
func TestReadSubmissionsMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/submissions" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":10,"slug":"sub1","status":"completed","template":{"id":99,"name":"NDA"},"created_at":"2026-02-01T00:00:00Z","completed_at":"2026-02-02T00:00:00Z"}],"pagination":{"count":1,"next":null,"prev":10}}`))
	}))
	defer srv.Close()

	c := docuseal.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "submissions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	rec := got[0]
	if rec["status"] != "completed" {
		t.Fatalf("status = %v, want completed", rec["status"])
	}
	if rec["template_id"] == nil {
		t.Fatalf("template_id missing, want flattened nested template.id: %+v", rec)
	}
}

// TestFixtureMode confirms credential-free fixture reads work for conformance.
func TestFixtureMode(t *testing.T) {
	c := docuseal.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "templates", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode produced no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}

	// Check must short-circuit without network in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
}

// TestCatalogAndRegistry confirms the catalog publishes the core streams and
// the connector self-registers and resolves via the registry.
func TestCatalogAndRegistry(t *testing.T) {
	c := docuseal.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"templates": true, "submissions": true, "submitters": true}
	for _, s := range cat.Streams {
		delete(want, s.Name)
	}
	if len(want) != 0 {
		t.Fatalf("catalog missing streams: %v", want)
	}

	meta := c.Metadata()
	if !meta.Capabilities.Read || !meta.Capabilities.Catalog || !meta.Capabilities.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", meta.Capabilities)
	}

	r := connectors.NewRegistry()
	if _, ok := r.Get("docuseal"); !ok {
		t.Fatal("registry did not resolve docuseal (self-registration)")
	}
}
