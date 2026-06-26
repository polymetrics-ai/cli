package oncehub_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/oncehub"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the OnceHub
// connector: API-Key header auth, Link-header "next" pagination driven by the
// `after=<last record id>` cursor over data[], and record mapping. Red until
// internal/connectors/oncehub exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("API-Key")
		if r.URL.Path != "/v2/bookings" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("after") {
		case "":
			// First page advertises a next page via the Link header.
			w.Header().Set("Link", fmt.Sprintf(`<%s/v2/bookings?after=bk_2>; rel="next"`, "http://"+r.Host))
			_, _ = w.Write([]byte(`{"data":[{"id":"bk_1","last_updated_time":"2026-01-01T00:00:00.000Z","status":"scheduled"},{"id":"bk_2","last_updated_time":"2026-01-02T00:00:00.000Z","status":"scheduled"}]}`))
		case "bk_2":
			// Last page: no Link header means no next page.
			_, _ = w.Write([]byte(`{"data":[{"id":"bk_3","last_updated_time":"2026-01-03T00:00:00.000Z","status":"completed"}]}`))
		default:
			t.Errorf("unexpected after=%q", r.URL.Query().Get("after"))
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := oncehub.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bookings", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "key_test_123" {
		t.Fatalf("API-Key header = %q, want key_test_123", sawAPIKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["last_updated_time"] == nil {
			t.Fatalf("record missing id/last_updated_time: %+v", rec)
		}
	}
	if got[0]["id"] != "bk_1" || got[2]["id"] != "bk_3" {
		t.Fatalf("unexpected record order: %+v", got)
	}
}

// TestReadIncrementalLowerBound asserts the start_date config is passed as the
// last_updated_time.gt filter on the first request for incremental streams.
func TestReadIncrementalLowerBound(t *testing.T) {
	var sawFilter string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawFilter = r.URL.Query().Get("last_updated_time.gt")
		_, _ = w.Write([]byte(`{"data":[{"id":"bk_1","last_updated_time":"2026-02-01T00:00:00.000Z"}]}`))
	}))
	defer srv.Close()

	c := oncehub.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "start_date": "2026-01-15T00:00:00Z"},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bookings", Config: cfg}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawFilter != "2026-01-15T00:00:00Z" {
		t.Fatalf("last_updated_time.gt = %q, want 2026-01-15T00:00:00Z", sawFilter)
	}
}

func TestFixtureModeReadNoNetwork(t *testing.T) {
	c := oncehub.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
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
}

func TestCheckFixtureMode(t *testing.T) {
	c := oncehub.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := oncehub.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "oncehub" {
		t.Fatalf("catalog connector = %q, want oncehub", cat.Connector)
	}
	want := map[string]bool{"bookings": true, "contacts": true, "booking_pages": true, "users": true, "event_types": true}
	for _, s := range cat.Streams {
		delete(want, s.Name)
		if len(s.PrimaryKey) == 0 || s.PrimaryKey[0] != "id" {
			t.Fatalf("stream %q primary key = %v, want [id]", s.Name, s.PrimaryKey)
		}
	}
	if len(want) != 0 {
		t.Fatalf("catalog missing streams: %v", want)
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := oncehub.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bookings", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected")
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = oncehub.New() // ensure init ran
	c := oncehub.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("oncehub"); !ok {
		t.Fatal("registry did not resolve oncehub (self-registration)")
	}
}
