package humanitix_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/humanitix"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Humanitix
// connector: x-api-key header auth, page-increment pagination over the events[]
// field, and record mapping. The events list returns a short final page to stop
// pagination. Red until internal/connectors/humanitix exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	var pagesSeen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("x-api-key")
		if r.URL.Path != "/events" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		pagesSeen = append(pagesSeen, page)
		switch page {
		case "1":
			// A full page (pageSize=2) signals there may be more.
			_, _ = w.Write([]byte(`{"total":3,"page":1,"pageSize":2,"events":[{"_id":"ev_1","name":"Launch","updatedAt":"2026-01-01T00:00:00Z"},{"_id":"ev_2","name":"Meetup","updatedAt":"2026-01-02T00:00:00Z"}]}`))
		case "2":
			// A short page (1 < pageSize) ends pagination.
			_, _ = w.Write([]byte(`{"total":3,"page":2,"pageSize":2,"events":[{"_id":"ev_3","name":"Workshop","updatedAt":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"events":[]}`))
		}
	}))
	defer srv.Close()

	c := humanitix.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "hx_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "hx_test_123" {
		t.Fatalf("x-api-key = %q, want hx_test_123", sawKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages); pages seen=%v", len(got), pagesSeen)
	}
	if len(pagesSeen) != 2 || pagesSeen[0] != "1" || pagesSeen[1] != "2" {
		t.Fatalf("pages requested = %v, want [1 2]", pagesSeen)
	}
	for _, rec := range got {
		if rec["_id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing _id/name: %+v", rec)
		}
	}
}

// TestReadFixtureMode asserts the credential-free fixture path emits
// deterministic records without any network call.
func TestReadFixtureMode(t *testing.T) {
	c := humanitix.New()
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
	if got[0]["_id"] == nil {
		t.Fatalf("fixture record missing _id: %+v", got[0])
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := humanitix.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams
// with _id primary keys.
func TestCatalogStreams(t *testing.T) {
	c := humanitix.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"events": false, "orders": false, "tickets": false, "tags": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
			if len(s.PrimaryKey) != 1 || s.PrimaryKey[0] != "_id" {
				t.Fatalf("stream %q primary key = %v, want [_id]", s.Name, s.PrimaryKey)
			}
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration via the global registry and
// that the connector advertises read (not write) capability.
func TestRegisteredReadOnly(t *testing.T) {
	_ = humanitix.New() // ensure init ran
	caps := humanitix.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("humanitix"); !ok {
		t.Fatal("registry did not resolve humanitix (self-registration)")
	}
}

// TestSubstreamRequiresEventID confirms orders/tickets require an event_id config
// when not in fixture mode.
func TestSubstreamRequiresEventID(t *testing.T) {
	c := humanitix.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "https://example.invalid"},
		Secrets: map[string]string{"api_key": "hx_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "orders", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("orders read without event_id should error")
	}
}
