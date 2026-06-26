package eventzilla_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/eventzilla"
)

// TestReadEventsPaginatesAndAuthenticates is the red-first test for the
// Eventzilla connector: x-api-key header auth, offset/limit pagination over two
// pages of the events stream (records under the "events" field path), and record
// mapping. Red until internal/connectors/eventzilla exists.
func TestReadEventsPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("x-api-key")
		if r.URL.Path != "/events" {
			http.NotFound(w, r)
			return
		}
		// limit=2 forces two pages: a full first page, then a short second page.
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"events":[{"id":1,"title":"Launch","status":"live"},{"id":2,"title":"Expo","status":"live"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"events":[{"id":3,"title":"Closing","status":"draft"}]}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"events":[]}`))
		}
	}))
	defer srv.Close()

	c := eventzilla.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"x-api-key": "ez_live_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "ez_live_key" {
		t.Fatalf("x-api-key = %q, want ez_live_key", sawAPIKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["title"] == nil {
			t.Fatalf("record missing id/title: %+v", rec)
		}
	}
}

// TestReadAttendeesSubstream verifies the substream path: the connector lists
// events to discover event ids, then reads /events/{id}/attendees for each,
// paginating the child endpoint and stamping the parent event_id.
func TestReadAttendeesSubstream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/events":
			if r.URL.Query().Get("offset") == "" || r.URL.Query().Get("offset") == "0" {
				_, _ = w.Write([]byte(`{"events":[{"id":7,"title":"Launch"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"events":[]}`))
		case "/events/7/attendees":
			if r.URL.Query().Get("offset") == "" || r.URL.Query().Get("offset") == "0" {
				_, _ = w.Write([]byte(`{"attendees":[{"id":11,"email":"a@x.com"},{"id":12,"email":"b@x.com"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"attendees":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := eventzilla.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "100"},
		Secrets: map[string]string{"x-api-key": "ez_live_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "attendees", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("attendees = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["event_id"] == nil {
			t.Fatalf("attendee missing parent event_id: %+v", rec)
		}
		if got, want := stringify(rec["event_id"]), "7"; got != want {
			t.Fatalf("event_id = %q, want %q", got, want)
		}
	}
}

func TestFixtureModeRead(t *testing.T) {
	c := eventzilla.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"events", "categories", "users", "attendees", "tickets"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) returned no records", stream)
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := eventzilla.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := eventzilla.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("eventzilla is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 5 {
		t.Fatalf("streams = %d, want >= 5", len(cat.Streams))
	}
}

func TestRegistryResolves(t *testing.T) {
	_ = eventzilla.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("eventzilla"); !ok {
		t.Fatal("registry did not resolve eventzilla (self-registration)")
	}
}

func stringify(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		return fmt.Sprintf("%v", t)
	}
}
