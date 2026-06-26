package fullstory_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/fullstory"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the FullStory
// connector: it asserts the Authorization header (Basic <api_key>), cursor
// pagination across two pages via next_page_token/pageToken, and record mapping
// from the results[] array. Red until internal/connectors/fullstory exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/segments/v2" {
			http.NotFound(w, r)
			return
		}
		calls++
		switch r.URL.Query().Get("pageToken") {
		case "":
			_, _ = w.Write([]byte(`{"results":[{"id":"seg_1","name":"All Sessions"},{"id":"seg_2","name":"Mobile"}],"next_page_token":"PAGE2"}`))
		case "PAGE2":
			_, _ = w.Write([]byte(`{"results":[{"id":"seg_3","name":"Web"}],"next_page_token":""}`))
		default:
			t.Errorf("unexpected pageToken=%q", r.URL.Query().Get("pageToken"))
			_, _ = w.Write([]byte(`{"results":[]}`))
		}
	}))
	defer srv.Close()

	c := fullstory.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "fs_api_secret", "uid": "uid_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "segments", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Basic fs_api_secret" {
		t.Fatalf("Authorization = %q, want Basic fs_api_secret", sawAuth)
	}
	if calls != 2 {
		t.Fatalf("server calls = %d, want 2 (paginated)", calls)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network call, so conformance can run without live credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := fullstory.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "segments", Config: cfg}, func(rec connectors.Record) error {
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
	// Check must also succeed in fixture mode without creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := fullstory.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "fullstory" {
		t.Fatalf("catalog connector = %q, want fullstory", cat.Connector)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if s.Name == "" || len(s.PrimaryKey) == 0 || len(s.Fields) == 0 {
			t.Fatalf("stream %+v missing name/pk/fields", s)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := fullstory.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "x", "uid": "y"},
	}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check should reject non-http(s) base_url")
	}
}

func TestRegisteredAndMetadata(t *testing.T) {
	_ = fullstory.New() // ensure init ran
	caps := fullstory.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("fullstory is read-only; Write should be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("fullstory"); !ok {
		t.Fatal("registry did not resolve fullstory (self-registration)")
	}
}
