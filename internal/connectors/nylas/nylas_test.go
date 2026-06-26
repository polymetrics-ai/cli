package nylas_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/nylas"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Nylas
// connector: Bearer auth with the api_key secret, cursor pagination over the
// data[] array (next_cursor in the body -> page_token query param), and record
// mapping. Red until internal/connectors/nylas exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPaths = append(sawPaths, r.URL.Path)
		if r.URL.Path != "/v3/grants/me/calendars" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page_token") {
		case "":
			_, _ = w.Write([]byte(`{"request_id":"r1","data":[{"id":"cal_1","grant_id":"me","name":"Work","timezone":"UTC"},{"id":"cal_2","grant_id":"me","name":"Home","timezone":"UTC"}],"next_cursor":"page2"}`))
		case "page2":
			_, _ = w.Write([]byte(`{"request_id":"r2","data":[{"id":"cal_3","grant_id":"me","name":"Personal","timezone":"UTC"}],"next_cursor":null}`))
		default:
			t.Errorf("unexpected page_token=%q", r.URL.Query().Get("page_token"))
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := nylas.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "nyk_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "calendars", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer nyk_test_123" {
		t.Fatalf("Authorization = %q, want Bearer nyk_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages); paths=%v", len(got), sawPaths)
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
	if len(sawPaths) != 2 {
		t.Fatalf("requests = %d, want 2 pages", len(sawPaths))
	}
}

// TestEventsRequireCalendarID verifies events read forwards calendar_id as a
// query parameter (Nylas requires it for the events endpoint).
func TestEventsRequireCalendarID(t *testing.T) {
	var sawCalendarID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/grants/me/events" {
			http.NotFound(w, r)
			return
		}
		sawCalendarID = r.URL.Query().Get("calendar_id")
		_, _ = w.Write([]byte(`{"request_id":"r1","data":[{"id":"evt_1","grant_id":"me","calendar_id":"primary","title":"Sync"}],"next_cursor":null}`))
	}))
	defer srv.Close()

	c := nylas.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "calendar_id": "primary"},
		Secrets: map[string]string{"api_key": "nyk_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read events: %v", err)
	}
	if sawCalendarID != "primary" {
		t.Fatalf("calendar_id = %q, want primary", sawCalendarID)
	}
	if len(got) != 1 || got[0]["id"] != "evt_1" {
		t.Fatalf("events = %+v, want one evt_1", got)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without a network call, so credential-free conformance passes.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := nylas.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"calendars", "contacts", "messages", "events"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := nylas.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check fixture mode: %v", err)
	}
}

// TestCatalogStreams confirms the catalog exposes the core streams with primary
// keys.
func TestCatalogStreams(t *testing.T) {
	c := nylas.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"calendars": false, "contacts": false, "messages": false, "events": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capability.
func TestRegisteredReadOnly(t *testing.T) {
	_ = nylas.New() // ensure init ran
	c := nylas.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only connector)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("nylas"); !ok {
		t.Fatal("registry did not resolve nylas (self-registration)")
	}
}
