package jotform_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/jotform"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the
// Jotform APIKEY header is sent, that the connector follows resultSet
// offset/limit pagination across two pages reading the top-level content[]
// array, and that records are mapped with id/created_at present.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("APIKEY")
		if r.URL.Path != "/user/forms" {
			http.NotFound(w, r)
			return
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit == 0 {
			t.Errorf("expected a limit query param, got none")
		}
		switch offset {
		case 0:
			_, _ = w.Write([]byte(`{"responseCode":200,"content":[` +
				`{"id":"f_1","title":"Contact","created_at":"2026-01-01 00:00:00","status":"ENABLED"},` +
				`{"id":"f_2","title":"Survey","created_at":"2026-01-02 00:00:00","status":"ENABLED"}` +
				`],"resultSet":{"offset":0,"limit":2,"count":2}}`))
		case 2:
			_, _ = w.Write([]byte(`{"responseCode":200,"content":[` +
				`{"id":"f_3","title":"RSVP","created_at":"2026-01-03 00:00:00","status":"DISABLED"}` +
				`],"resultSet":{"offset":2,"limit":2,"count":1}}`))
		default:
			t.Errorf("unexpected offset=%d", offset)
			_, _ = w.Write([]byte(`{"responseCode":200,"content":[],"resultSet":{"offset":0,"limit":2,"count":0}}`))
		}
	}))
	defer srv.Close()

	c := jotform.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "forms", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "key_test_123" {
		t.Fatalf("APIKEY = %q, want key_test_123", sawAPIKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["created_at"] == nil {
			t.Fatalf("record missing id/created_at: %+v", rec)
		}
	}
}

// TestReadSubmissionsMapsFormID covers a second stream and the form-scoped
// endpoint, confirming the per-stream routing table and record mapper.
func TestReadSubmissionsMapsFormID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("APIKEY") == "" {
			t.Errorf("missing APIKEY header on submissions request")
		}
		if r.URL.Path != "/user/submissions" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"responseCode":200,"content":[` +
			`{"id":"s_1","form_id":"f_1","created_at":"2026-02-01 10:00:00","status":"ACTIVE","new":"1"}` +
			`],"resultSet":{"offset":0,"limit":100,"count":1}}`))
	}))
	defer srv.Close()

	c := jotform.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "submissions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read submissions: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["form_id"] != "f_1" {
		t.Fatalf("form_id = %v, want f_1", got[0]["form_id"])
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// with no network access so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := jotform.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "forms", Config: cfg}, func(rec connectors.Record) error {
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
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode (no creds).
func TestCheckFixtureMode(t *testing.T) {
	c := jotform.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams ensures the published catalog has the core streams.
func TestCatalogStreams(t *testing.T) {
	c := jotform.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"forms": false, "submissions": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = jotform.New() // ensure init ran
	c := jotform.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("jotform"); !ok {
		t.Fatal("registry did not resolve jotform (self-registration)")
	}
}
