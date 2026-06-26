package ezofficeinventory_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/ezofficeinventory"
)

// TestReadPaginatesAndAuthenticates is the red-first test: EZOfficeInventory
// authenticates with a `token` header, paginates with a `page` query param
// (PageIncrement starting at 1, page_size 25), and returns records under the
// `assets` field. It asserts the auth header, pagination across two pages, and
// record mapping. Red until internal/connectors/ezofficeinventory exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.Header.Get("token")
		if r.URL.Path != "/assets.api" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "1":
			// A full page (page_size 25 worth of identifiers would be 25, but the
			// connector advances while total_pages > page), so signal more pages.
			_, _ = w.Write([]byte(`{"assets":[{"identifier":1,"name":"Laptop"},{"identifier":2,"name":"Monitor"}],"total_pages":2,"page":1}`))
		case "2":
			_, _ = w.Write([]byte(`{"assets":[{"identifier":3,"name":"Dock"}],"total_pages":2,"page":2}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"assets":[],"total_pages":2,"page":3}`))
		}
	}))
	defer srv.Close()

	c := ezofficeinventory.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "tok_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "assets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "tok_abc123" {
		t.Fatalf("token header = %q, want tok_abc123", sawToken)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["identifier"] == nil || rec["name"] == nil {
			t.Fatalf("record missing identifier/name: %+v", rec)
		}
	}
}

// TestReadFixtureMode confirms the credential-free fixture path emits
// deterministic records so the conformance harness runs without live creds.
func TestReadFixtureMode(t *testing.T) {
	c := ezofficeinventory.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "members", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	if got[0]["id"] == nil {
		t.Fatalf("fixture member missing id: %+v", got[0])
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode (no network).
func TestCheckFixtureMode(t *testing.T) {
	c := ezofficeinventory.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams asserts the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := ezofficeinventory.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"assets": false, "inventories": false, "members": false, "locations": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Errorf("stream %q has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("catalog missing core stream %q", name)
		}
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF validation on base_url override.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := ezofficeinventory.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "assets", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected")
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = ezofficeinventory.New() // ensure init ran
	caps := ezofficeinventory.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("ezofficeinventory"); !ok {
		t.Fatal("registry did not resolve ezofficeinventory (self-registration)")
	}
}
