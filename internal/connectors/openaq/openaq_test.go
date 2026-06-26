package openaq_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/openaq"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the OpenAQ
// connector: it asserts the X-API-Key header is sent, that page-number
// pagination walks across two pages (results[] + meta.found), and that records
// are mapped through the stream's mapper. Red until internal/connectors/openaq
// exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	var sawPages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("X-API-Key")
		if r.URL.Path != "/countries" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		sawPages = append(sawPages, page)
		// limit=2, found=3 -> page 1 returns 2, page 2 returns 1, then stop.
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{"meta":{"page":1,"limit":2,"found":3},"results":[{"id":1,"code":"US","name":"United States"},{"id":2,"code":"CA","name":"Canada"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"meta":{"page":2,"limit":2,"found":3},"results":[{"id":3,"code":"MX","name":"Mexico"}]}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"meta":{"page":3,"limit":2,"found":3},"results":[]}`))
		}
	}))
	defer srv.Close()

	c := openaq.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "countries", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "key_test_123" {
		t.Fatalf("X-API-Key = %q, want key_test_123", sawKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if len(sawPages) != 2 {
		t.Fatalf("requested %d pages (%v), want exactly 2", len(sawPages), sawPages)
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["code"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/code/name: %+v", rec)
		}
	}
}

// TestReadFixtureMode confirms credential-free fixture reads work for every
// declared stream (conformance runs without live keys).
func TestReadFixtureMode(t *testing.T) {
	c := openaq.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"countries", "parameters", "locations", "instruments", "manufacturers"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture Read(%s) returned %d records, want 2", stream, len(got))
		}
		if got[0]["id"] == nil {
			t.Fatalf("fixture Read(%s) record missing id: %+v", stream, got[0])
		}
	}
}

// TestCatalogStreams asserts the published catalog has the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := openaq.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"countries": false, "parameters": false, "locations": false, "instruments": false, "manufacturers": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestCheckRequiresAPIKey ensures Check fails fast without credentials and
// short-circuits in fixture mode.
func TestCheckRequiresAPIKey(t *testing.T) {
	c := openaq.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{}); err == nil {
		t.Fatal("Check without api_key should fail")
	}
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check in fixture mode should pass, got %v", err)
	}
}

// TestRejectsBadBaseURL guards against SSRF via a non-http(s) base_url override.
func TestRejectsBadBaseURL(t *testing.T) {
	c := openaq.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "countries", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}

// TestRegistryResolves confirms the connector self-registers and is read-only.
func TestRegistryResolves(t *testing.T) {
	_ = openaq.New() // ensure init ran
	caps := openaq.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("openaq is read-only; Write should be false, got %+v", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("openaq"); !ok {
		t.Fatal("registry did not resolve openaq (self-registration)")
	}
}
