package granola_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/granola"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Granola
// connector: Bearer auth with the grn_ key, cursor-based pagination over the
// notes[] array (hasMore + cursor), and record mapping. Red until
// internal/connectors/granola exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/notes" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"notes":[{"id":"not_1","title":"Standup","created_at":"2026-01-01T00:00:00Z"},{"id":"not_2","title":"Planning","created_at":"2026-01-02T00:00:00Z"}],"hasMore":true,"cursor":"page2"}`))
		case "page2":
			_, _ = w.Write([]byte(`{"notes":[{"id":"not_3","title":"Retro","created_at":"2026-01-03T00:00:00Z"}],"hasMore":false,"cursor":""}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"notes":[],"hasMore":false}`))
		}
	}))
	defer srv.Close()

	c := granola.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "grn_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "notes", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer grn_test_123" {
		t.Fatalf("Authorization = %q, want Bearer grn_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["title"] == nil {
			t.Fatalf("record missing id/title: %+v", rec)
		}
	}
	if got[0]["id"] != "not_1" || got[2]["id"] != "not_3" {
		t.Fatalf("unexpected record order/ids: %+v", got)
	}
}

// TestReadDetailedNotes confirms the detailed_notes stream fans out from the
// notes list to per-note GET /notes/{id} detail fetches.
func TestReadDetailedNotes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/notes":
			if r.URL.Query().Get("cursor") == "" {
				_, _ = w.Write([]byte(`{"notes":[{"id":"not_1"},{"id":"not_2"}],"hasMore":false,"cursor":""}`))
				return
			}
			_, _ = w.Write([]byte(`{"notes":[],"hasMore":false}`))
		case "/notes/not_1":
			_, _ = w.Write([]byte(`{"id":"not_1","title":"Standup","summary":"a summary","created_at":"2026-01-01T00:00:00Z"}`))
		case "/notes/not_2":
			_, _ = w.Write([]byte(`{"id":"not_2","title":"Planning","summary":"b summary","created_at":"2026-01-02T00:00:00Z"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := granola.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "grn_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "detailed_notes", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read detailed_notes: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("detailed records = %d, want 2", len(got))
	}
	if got[0]["summary"] == nil || got[1]["summary"] == nil {
		t.Fatalf("detail record missing summary: %+v", got)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// with no network access and no credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := granola.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "notes", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := granola.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"notes": false, "detailed_notes": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = granola.New() // ensure init ran
	c := granola.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("granola"); !ok {
		t.Fatal("registry did not resolve granola (self-registration)")
	}
}
