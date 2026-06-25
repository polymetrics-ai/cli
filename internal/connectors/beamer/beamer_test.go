package beamer_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/beamer"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Beamer
// connector: Bearer auth, Beamer page-increment pagination (page=0,1,...) over a
// root-level JSON array, page size carried via maxResults, and record mapping.
// Beamer returns a bare array per page and stops when a short page arrives.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawMaxResults string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawMaxResults = r.URL.Query().Get("maxResults")
		if r.URL.Path != "/v0/nps" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "0":
			// A full page (== page size 2) so the loop requests page 1.
			_, _ = w.Write([]byte(`[{"id":"n1","date":"2026-01-01T00:00:00Z","score":9},{"id":"n2","date":"2026-01-02T00:00:00Z","score":8}]`))
		case "1":
			// A short page (< page size) ends pagination.
			_, _ = w.Write([]byte(`[{"id":"n3","date":"2026-01-03T00:00:00Z","score":7}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := beamer.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "beamer_secret_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "nps", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer beamer_secret_123" {
		t.Fatalf("Authorization = %q, want Bearer beamer_secret_123", sawAuth)
	}
	if sawMaxResults != "2" {
		t.Fatalf("maxResults = %q, want 2", sawMaxResults)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["date"] == nil {
			t.Fatalf("record missing id/date: %+v", rec)
		}
	}
}

// TestReadIncrementalUsesStartDate asserts the incremental cursor / start_date
// flows into the Beamer dateFrom request parameter.
func TestReadIncrementalUsesStartDate(t *testing.T) {
	var sawDateFrom string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawDateFrom = r.URL.Query().Get("dateFrom")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := beamer.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "start_date": "2026-05-01T00:00:00Z"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "nps", Config: cfg}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawDateFrom != "2026-05-01T00:00:00Z" {
		t.Fatalf("dateFrom = %q, want 2026-05-01T00:00:00Z", sawDateFrom)
	}
}

// TestFixtureModeNeedsNoNetwork confirms the credential-free fixture path used by
// the conformance harness.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := beamer.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "nps", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits without a network call in
// fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := beamer.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published streams include the verified nps
// stream with its primary key and cursor field.
func TestCatalogStreams(t *testing.T) {
	c := beamer.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	var nps *connectors.Stream
	for i := range cat.Streams {
		if cat.Streams[i].Name == "nps" {
			nps = &cat.Streams[i]
		}
	}
	if nps == nil {
		t.Fatal("nps stream missing from catalog")
	}
	if len(nps.PrimaryKey) != 1 || nps.PrimaryKey[0] != "id" {
		t.Fatalf("nps primary key = %v, want [id]", nps.PrimaryKey)
	}
	if len(nps.CursorFields) != 1 || nps.CursorFields[0] != "date" {
		t.Fatalf("nps cursor fields = %v, want [date]", nps.CursorFields)
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = beamer.New() // ensure init ran
	c := beamer.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("beamer"); !ok {
		t.Fatal("registry did not resolve beamer (self-registration)")
	}
}
