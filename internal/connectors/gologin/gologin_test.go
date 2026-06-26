package gologin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/gologin"
)

// TestReadProfilesPaginatesAndAuthenticates is the red-first test: Bearer auth on
// the api_key secret, GoLogin page-number pagination over the profiles[] envelope
// across two pages, and record mapping.
func TestReadProfilesPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/browser/v2" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// A full page (page size 2 for the test via config) signals more pages.
			_, _ = w.Write([]byte(`{"profiles":[{"id":"p1","name":"Alpha","updatedAt":"2026-01-01T00:00:00Z"},{"id":"p2","name":"Beta","updatedAt":"2026-01-02T00:00:00Z"}],"allProfilesCount":3}`))
		case "2":
			_, _ = w.Write([]byte(`{"profiles":[{"id":"p3","name":"Gamma","updatedAt":"2026-01-03T00:00:00Z"}],"allProfilesCount":3}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"profiles":[]}`))
		}
	}))
	defer srv.Close()

	c := gologin.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "gl_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "profiles", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer gl_test_123" {
		t.Fatalf("Authorization = %q, want Bearer gl_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (across 2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestReadFoldersRootArray exercises a root-array stream (folders) with a single
// page, confirming the record selector path "" works.
func TestReadFoldersRootArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/folders" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"f1","name":"Work"},{"id":"f2","name":"Personal"}]`))
	}))
	defer srv.Close()

	c := gologin.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "gl_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "folders", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read folders: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("folders = %d, want 2", len(got))
	}
}

// TestFixtureModeReadsWithoutNetwork confirms the credential-free fixture path
// emits deterministic records so conformance runs without live creds.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := gologin.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"profiles", "folders", "user", "tags"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read %s: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read %s emitted no records", stream)
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode (no network).
func TestCheckFixtureMode(t *testing.T) {
	c := gologin.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog includes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := gologin.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"profiles": false, "folders": false, "user": false, "tags": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegistryResolvesGologin confirms self-registration via init().
func TestRegistryResolvesGologin(t *testing.T) {
	_ = gologin.New() // ensure the package init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("gologin")
	if !ok {
		t.Fatal("registry did not resolve gologin (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
