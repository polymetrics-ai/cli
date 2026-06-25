package chameleon_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/chameleon"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Chameleon
// connector: it asserts the X-Account-Secret auth header, limit/offset
// pagination across two pages, and record mapping from the surveys field_path.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawSecret string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawSecret = r.Header.Get("X-Account-Secret")
		if r.URL.Path != "/edit/surveys" {
			http.NotFound(w, r)
			return
		}
		// page size is 2 (set via limit config); the first full page signals more.
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"surveys":[{"id":"s_1","title":"NPS","updated_at":"2026-01-01T00:00:00.000Z"},{"id":"s_2","title":"CSAT","updated_at":"2026-01-02T00:00:00.000Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"surveys":[{"id":"s_3","title":"Onboarding","updated_at":"2026-01-03T00:00:00.000Z"}]}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"surveys":[]}`))
		}
	}))
	defer srv.Close()

	c := chameleon.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "limit": "2"},
		Secrets: map[string]string{"api_key": "acct_secret_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "surveys", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawSecret != "acct_secret_123" {
		t.Fatalf("X-Account-Secret = %q, want acct_secret_123", sawSecret)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["updated_at"] == nil {
			t.Fatalf("record missing id/updated_at: %+v", rec)
		}
	}
	if got[0]["title"] != "NPS" {
		t.Fatalf("first record title = %v, want NPS", got[0]["title"])
	}
}

// TestSegmentsRecordMapping verifies a second stream maps from its own
// field_path (segments) and preserves typed fields.
func TestSegmentsRecordMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/edit/segments" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"segments":[{"id":"seg_1","name":"Power Users","updated_at":"2026-02-01T00:00:00.000Z"}]}`))
	}))
	defer srv.Close()

	c := chameleon.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "acct_secret_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "segments", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "seg_1" || got[0]["name"] != "Power Users" {
		t.Fatalf("segments mapping wrong: %+v", got)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access (credential-free conformance).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := chameleon.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "surveys", Config: cfg}, func(rec connectors.Record) error {
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
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams checks the published catalog exposes the core streams with
// id primary keys and updated_at cursors.
func TestCatalogStreams(t *testing.T) {
	c := chameleon.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"surveys": false, "segments": false, "tours": false, "launchers": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
			if len(s.PrimaryKey) == 0 || s.PrimaryKey[0] != "id" {
				t.Fatalf("stream %s primary key = %v, want [id]", s.Name, s.PrimaryKey)
			}
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestRegisteredReadOnly asserts self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = chameleon.New() // ensure init ran
	c := chameleon.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("chameleon"); !ok {
		t.Fatal("registry did not resolve chameleon (self-registration)")
	}
}
