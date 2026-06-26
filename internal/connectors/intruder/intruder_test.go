package intruder_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/intruder"
)

// TestReadPaginatesAndAuthenticates is the red-first test: Bearer auth, Intruder
// offset/limit pagination over results[], and record mapping for the targets
// stream. Red until internal/connectors/intruder exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/targets" {
			http.NotFound(w, r)
			return
		}
		// Offset-based pagination: page one returns a full page (limit records)
		// so the loop requests page two; page two is short and stops.
		offset := r.URL.Query().Get("offset")
		switch offset {
		case "", "0":
			_, _ = w.Write([]byte(`{"results":[` +
				targetJSON(1) + `,` + targetJSON(2) + `],"count":3}`))
		case "2":
			_, _ = w.Write([]byte(`{"results":[` + targetJSON(3) + `],"count":3}`))
		default:
			t.Errorf("unexpected offset=%q", offset)
			_, _ = w.Write([]byte(`{"results":[]}`))
		}
	}))
	defer srv.Close()

	c := intruder.New()
	cfg := connectors.RuntimeConfig{
		// page_size 2 forces a two-page walk against the test server.
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"access_token": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "targets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_test_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["address"] == nil {
			t.Fatalf("record missing id/address: %+v", rec)
		}
	}
}

// TestReadOccurrencesSubstream verifies the parent-slice substream: the
// connector first lists issues, then reads /issues/{id}/occurrences per issue.
func TestReadOccurrencesSubstream(t *testing.T) {
	var issuesHit, occHits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/issues":
			issuesHit++
			_, _ = w.Write([]byte(`{"results":[{"id":10,"title":"a"},{"id":11,"title":"b"}]}`))
		case "/issues/10/occurrences":
			occHits++
			_, _ = w.Write([]byte(`{"results":[{"id":100,"target":"host-a","port":443}]}`))
		case "/issues/11/occurrences":
			occHits++
			_, _ = w.Write([]byte(`{"results":[{"id":110,"target":"host-b","port":80}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := intruder.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"access_token": "tok"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "occurrences", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read occurrences: %v", err)
	}
	if issuesHit != 1 {
		t.Fatalf("issues endpoint hit %d times, want 1", issuesHit)
	}
	if occHits != 2 {
		t.Fatalf("occurrences endpoint hit %d times, want 2 (one per issue)", occHits)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["issue_id"] == nil {
			t.Fatalf("occurrence record missing id/issue_id: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access so credential-free conformance works.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := intruder.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"issues", "scans", "targets", "occurrences"} {
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
	// Check short-circuits in fixture mode (no creds, no network).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestRegisteredReadOnly verifies self-registration and capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = intruder.New() // ensure init ran
	c := intruder.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("intruder"); !ok {
		t.Fatal("registry did not resolve intruder (self-registration)")
	}
}

// TestCatalogStreams checks the published streams.
func TestCatalogStreams(t *testing.T) {
	c := intruder.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"issues": true, "scans": true, "targets": true, "occurrences": true}
	for _, s := range cat.Streams {
		delete(want, s.Name)
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s has no primary key", s.Name)
		}
	}
	if len(want) != 0 {
		t.Fatalf("catalog missing streams: %v", want)
	}
}

func targetJSON(id int) string {
	s := strconv.Itoa(id)
	return `{"id":` + s + `,"address":"host-` + s + `.example.com","tags":["web"]}`
}
