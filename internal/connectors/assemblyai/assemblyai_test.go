package assemblyai_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/assemblyai"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the AssemblyAI
// connector: raw API-key Authorization header, page_details.next_url pagination
// over transcripts[], and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v2/transcript" {
			http.NotFound(w, r)
			return
		}
		// Second page is requested via the absolute next_url with after_id=t2.
		if r.URL.Query().Get("after_id") == "t2" {
			_, _ = w.Write([]byte(`{"transcripts":[{"id":"t3","status":"completed","created":"2026-01-03T00:00:00Z"}],"page_details":{"next_url":null}}`))
			return
		}
		// First page returns 2 records and a next_url pointing back at this server.
		next := srv.URL + "/v2/transcript?after_id=t2&limit=2"
		_, _ = w.Write([]byte(`{"transcripts":[{"id":"t1","status":"completed","created":"2026-01-01T00:00:00Z"},{"id":"t2","status":"completed","created":"2026-01-02T00:00:00Z"}],"page_details":{"next_url":"` + next + `"}}`))
	}))
	defer srv.Close()

	c := assemblyai.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "transcript", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	// AssemblyAI uses the raw API key in the Authorization header (no Bearer prefix).
	if sawAuth != "key_123" {
		t.Fatalf("Authorization = %q, want raw key_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["status"] == nil {
			t.Fatalf("record missing id/status: %+v", rec)
		}
	}
	if got[0]["id"] != "t1" || got[2]["id"] != "t3" {
		t.Fatalf("unexpected record ordering: %+v", got)
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network call, so conformance works without credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := assemblyai.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "transcript", Config: cfg}, func(rec connectors.Record) error {
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
	// Check must short-circuit in fixture mode (no creds).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestBaseURLSSRFValidation rejects non-http(s) override schemes.
func TestBaseURLSSRFValidation(t *testing.T) {
	c := assemblyai.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "key_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "transcript", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme rejection, got %v", err)
	}
}

// TestCatalogStreams verifies the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := assemblyai.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "assemblyai" {
		t.Fatalf("catalog connector = %q, want assemblyai", cat.Connector)
	}
	want := map[string]bool{"transcript": false}
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

// TestRegisteredReadOnly verifies self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	c := assemblyai.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("assemblyai"); !ok {
		t.Fatal("registry did not resolve assemblyai (self-registration)")
	}
}
