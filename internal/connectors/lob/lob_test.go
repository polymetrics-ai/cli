package lob_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/lob"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Lob connector:
// HTTP Basic auth (API key as username, blank password), Lob next_url/after
// cursor pagination over the data[] array, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/postcards" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("after") {
		case "":
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"psc_1","date_created":"2026-01-01T00:00:00Z"},{"id":"psc_2","date_created":"2026-01-02T00:00:00Z"}],"count":2,"next_url":"/postcards?limit=2&after=psc_2"}`))
		case "psc_2":
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"psc_3","date_created":"2026-01-03T00:00:00Z"}],"count":1,"next_url":null}`))
		default:
			t.Errorf("unexpected after=%q", r.URL.Query().Get("after"))
			_, _ = w.Write([]byte(`{"object":"list","data":[],"next_url":null}`))
		}
	}))
	defer srv.Close()

	c := lob.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "test_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "postcards", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("test_key_123:"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["date_created"] == nil {
			t.Fatalf("record missing id/date_created: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records with
// no network access, so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := lob.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "addresses", Config: cfg}, func(rec connectors.Record) error {
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

// TestCheckFixtureMode ensures Check short-circuits in fixture mode (no creds).
func TestCheckFixtureMode(t *testing.T) {
	c := lob.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog has the core streams with
// primary keys and cursor fields.
func TestCatalogStreams(t *testing.T) {
	c := lob.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
		if len(s.CursorFields) == 0 {
			t.Fatalf("stream %q missing cursor fields", s.Name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration via NewRegistry and that the
// connector advertises read (not write) capability.
func TestRegisteredReadOnly(t *testing.T) {
	_ = lob.New() // ensure init ran
	c := lob.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("lob"); !ok {
		t.Fatal("registry did not resolve lob (self-registration)")
	}
}
