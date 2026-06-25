package braze_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/braze"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Braze
// connector: Bearer auth on the Authorization header, Braze page-number
// pagination (page=0,1,...) over the per-stream array, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/campaigns/list" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "0":
			// A full page (the connector's page size is small in tests via config).
			_, _ = w.Write([]byte(`{"message":"success","campaigns":[{"id":"c1","name":"Welcome","last_edited":"2026-01-01T00:00:00Z"},{"id":"c2","name":"Winback","last_edited":"2026-01-02T00:00:00Z"}]}`))
		case "1":
			// A short page signals the end of pagination.
			_, _ = w.Write([]byte(`{"message":"success","campaigns":[{"id":"c3","name":"Onboarding","last_edited":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"message":"success","campaigns":[]}`))
		}
	}))
	defer srv.Close()

	c := braze.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "rest_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer rest_key_123" {
		t.Fatalf("Authorization = %q, want Bearer rest_key_123", sawAuth)
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

// TestReadCanvasesStream confirms a second stream maps from its own top-level
// array field ("canvases") and primary key.
func TestReadCanvasesStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/canvas/list" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "0":
			_, _ = w.Write([]byte(`{"message":"success","canvases":[{"id":"cv1","name":"Journey","last_edited":"2026-02-01T00:00:00Z"}]}`))
		default:
			_, _ = w.Write([]byte(`{"message":"success","canvases":[]}`))
		}
	}))
	defer srv.Close()

	c := braze.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "100"},
		Secrets: map[string]string{"api_key": "rest_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "canvases", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read canvases: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "cv1" {
		t.Fatalf("canvases = %+v, want one record id=cv1", got)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records with
// no network access (so conformance runs without live creds).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := braze.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "segments", Config: cfg}, func(rec connectors.Record) error {
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
	// Check should also short-circuit without creds in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog includes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := braze.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"campaigns": false, "canvases": false, "segments": false, "events": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := braze.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "rest_key_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = braze.New() // ensure init ran
	c := braze.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("braze is read-only; Write should be false, got %+v", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("braze"); !ok {
		t.Fatal("registry did not resolve braze (self-registration)")
	}
}
