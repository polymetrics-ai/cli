package luma_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/luma"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Luma
// connector: x-luma-api-key auth, Luma cursor pagination
// (pagination_cursor / next_cursor / has_more), and the nested
// entries[].event record mapping. Red until internal/connectors/luma exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var pages int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("x-luma-api-key")
		if r.URL.Path != "/calendar/list-events" {
			http.NotFound(w, r)
			return
		}
		pages++
		switch r.URL.Query().Get("pagination_cursor") {
		case "":
			_, _ = w.Write([]byte(`{"entries":[{"event":{"api_id":"evt-1","name":"Launch","start_at":"2026-01-01T00:00:00Z"}},{"event":{"api_id":"evt-2","name":"Demo","start_at":"2026-01-02T00:00:00Z"}}],"has_more":true,"next_cursor":"cur-2"}`))
		case "cur-2":
			_, _ = w.Write([]byte(`{"entries":[{"event":{"api_id":"evt-3","name":"Recap","start_at":"2026-01-03T00:00:00Z"}}],"has_more":false,"next_cursor":null}`))
		default:
			t.Errorf("unexpected pagination_cursor=%q", r.URL.Query().Get("pagination_cursor"))
			_, _ = w.Write([]byte(`{"entries":[],"has_more":false}`))
		}
	}))
	defer srv.Close()

	c := luma.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "luma-secret-123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "luma-secret-123" {
		t.Fatalf("x-luma-api-key = %q, want luma-secret-123", sawAuth)
	}
	if pages != 2 {
		t.Fatalf("pages = %d, want 2", pages)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["api_id"] != "evt-1" || got[0]["name"] != "Launch" {
		t.Fatalf("first record mapped wrong: %+v", got[0])
	}
	if got[2]["api_id"] != "evt-3" {
		t.Fatalf("last record mapped wrong: %+v", got[2])
	}
}

// TestReadGuestsRequiresEventID confirms the event_guests substream reads from
// /event/get-guests, threads the configured event_api_id query param, and
// unwraps the nested entries[].guest objects.
func TestReadGuestsRequiresEventID(t *testing.T) {
	var sawEventID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/event/get-guests" {
			http.NotFound(w, r)
			return
		}
		sawEventID = r.URL.Query().Get("event_api_id")
		_, _ = w.Write([]byte(`{"entries":[{"guest":{"api_id":"gst-1","approval_status":"approved","name":"Ada"}}],"has_more":false}`))
	}))
	defer srv.Close()

	c := luma.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "event_api_id": "evt-42"},
		Secrets: map[string]string{"api_key": "luma-secret-123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "event_guests", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawEventID != "evt-42" {
		t.Fatalf("event_api_id = %q, want evt-42", sawEventID)
	}
	if len(got) != 1 || got[0]["api_id"] != "gst-1" || got[0]["approval_status"] != "approved" {
		t.Fatalf("guest record mapped wrong: %+v", got)
	}
}

// TestFixtureMode confirms credential-free fixture reads emit deterministic
// records for the conformance harness.
func TestFixtureMode(t *testing.T) {
	c := luma.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["api_id"] == nil {
			t.Fatalf("fixture record missing api_id: %+v", rec)
		}
	}
	// Check must succeed without creds in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogAndMetadata confirms the published catalog and read-only metadata.
func TestCatalogAndMetadata(t *testing.T) {
	c := luma.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("luma is read-only, Write should be false: %+v", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 2 {
		t.Fatalf("streams = %d, want >= 2", len(cat.Streams))
	}
	if cat.Connector != "luma" {
		t.Fatalf("catalog connector = %q, want luma", cat.Connector)
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs to bound SSRF.
func TestBaseURLValidation(t *testing.T) {
	c := luma.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected base_url scheme validation error")
	}
}

// TestRegistryResolution confirms self-registration via NewRegistry.
func TestRegistryResolution(t *testing.T) {
	_ = luma.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("luma"); !ok {
		t.Fatal("registry did not resolve luma (self-registration)")
	}
}
