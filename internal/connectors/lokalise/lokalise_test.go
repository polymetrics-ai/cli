package lokalise_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/lokalise"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Lokalise
// connector: X-Api-Token header auth, project-scoped endpoint path, offset
// pagination over two pages driven by the X-Pagination-Page-Count header, and
// record mapping of the keys stream.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawToken string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.Header.Get("X-Api-Token")
		sawPaths = append(sawPaths, r.URL.Path)
		if r.URL.Path != "/projects/proj_123/keys" {
			http.NotFound(w, r)
			return
		}
		// Two pages: page 1 of 2, then page 2 of 2.
		w.Header().Set("X-Pagination-Page-Count", "2")
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		w.Header().Set("X-Pagination-Page", page)
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{"project_id":"proj_123","keys":[{"key_id":1,"created_at":"2026-01-01T00:00:00Z"},{"key_id":2,"created_at":"2026-01-02T00:00:00Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"project_id":"proj_123","keys":[{"key_id":3,"created_at":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"keys":[]}`))
		}
	}))
	defer srv.Close()

	c := lokalise.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "project_id": "proj_123"},
		Secrets: map[string]string{"api_key": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "keys", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "tok_abc" {
		t.Fatalf("X-Api-Token = %q, want tok_abc", sawToken)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if len(sawPaths) != 2 {
		t.Fatalf("requests = %d, want 2 (pagination must stop at page count)", len(sawPaths))
	}
	for _, rec := range got {
		if rec["key_id"] == nil {
			t.Fatalf("record missing key_id: %+v", rec)
		}
	}
}

// TestReadLanguagesMapsFields confirms a second stream maps its primary key and
// records under the languages array.
func TestReadLanguagesMapsFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj_123/languages" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Pagination-Page-Count", "1")
		_, _ = w.Write([]byte(`{"languages":[{"lang_id":640,"lang_iso":"en","lang_name":"English"},{"lang_id":597,"lang_iso":"ru","lang_name":"Russian"}]}`))
	}))
	defer srv.Close()

	c := lokalise.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "project_id": "proj_123"},
		Secrets: map[string]string{"api_key": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "languages", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["lang_iso"] != "en" || got[0]["lang_name"] != "English" {
		t.Fatalf("first language = %+v, want en/English", got[0])
	}
}

// TestFixtureModeNeedsNoNetwork confirms fixture mode emits deterministic records
// without any HTTP call (credential-free conformance).
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := lokalise.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"keys", "languages", "translations", "contributors", "comments"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := lokalise.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := lokalise.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"keys": false, "languages": false, "translations": false, "contributors": false, "comments": false}
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

// TestBaseURLRejectsBadScheme guards the SSRF validation on base_url override.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := lokalise.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "project_id": "p"},
		Secrets: map[string]string{"api_key": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "keys", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}

// TestRegistryResolution confirms self-registration via NewRegistry().Get.
func TestRegistryResolution(t *testing.T) {
	_ = lokalise.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("lokalise"); !ok {
		t.Fatal("registry did not resolve lokalise (self-registration)")
	}
}

// TestMetadataReadOnly confirms capabilities are read-only (no write).
func TestMetadataReadOnly(t *testing.T) {
	c := lokalise.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only connector)", caps)
	}
}
