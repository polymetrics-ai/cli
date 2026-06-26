package nocrm_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/nocrm"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the noCRM
// connector: X-API-KEY auth, offset/limit pagination over a top-level JSON
// array using X-TOTAL-COUNT to detect the final page, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("X-API-KEY")
		if r.URL.Path != "/api/v2/leads" {
			http.NotFound(w, r)
			return
		}
		// Two pages of two records each; total of 3 so the second page is short.
		w.Header().Set("X-TOTAL-COUNT", "3")
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`[{"id":1,"title":"Lead One","status":"todo"},{"id":2,"title":"Lead Two","status":"won"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"id":3,"title":"Lead Three","status":"lost"}]`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := nocrm.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/api/v2", "page_size": "2"},
		Secrets: map[string]string{"api_key": "secret_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "leads", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "secret_key_123" {
		t.Fatalf("X-API-KEY = %q, want secret_key_123", sawAPIKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["title"] == nil {
			t.Fatalf("record missing id/title: %+v", rec)
		}
	}
}

func TestReadStopsAtPageSizeBoundary(t *testing.T) {
	// When the API returns a full page but X-TOTAL-COUNT is absent, the loop
	// must stop after the first short page rather than spinning forever.
	var pages int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pages++
		if r.URL.Path != "/api/v2/pipelines" {
			http.NotFound(w, r)
			return
		}
		// Single record (< page_size) so this is the last page.
		_, _ = w.Write([]byte(`[{"id":10,"name":"Default Pipeline"}]`))
	}))
	defer srv.Close()

	c := nocrm.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/api/v2", "page_size": "100"},
		Secrets: map[string]string{"api_key": "k"},
	}
	var got int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pipelines", Config: cfg}, func(connectors.Record) error {
		got++
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got != 1 {
		t.Fatalf("records = %d, want 1", got)
	}
	if pages != 1 {
		t.Fatalf("requested %d pages, want 1 (short page should stop)", pages)
	}
}

func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := nocrm.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil {
		t.Fatalf("fixture record missing id: %+v", got[0])
	}
}

func TestCheckFixtureModeNoNetwork(t *testing.T) {
	c := nocrm.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := nocrm.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"leads": false, "pipelines": false, "users": false, "teams": false, "prospecting_lists": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Errorf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("catalog missing stream %q", name)
		}
	}
}

func TestRegistryResolvesNocrm(t *testing.T) {
	_ = nocrm.New() // ensure init() ran
	r := connectors.NewRegistry()
	got, ok := r.Get("nocrm")
	if !ok {
		t.Fatal("registry did not resolve nocrm (self-registration)")
	}
	if got.Name() != "nocrm" {
		t.Fatalf("Name() = %q, want nocrm", got.Name())
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := nocrm.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "leads", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected base_url scheme validation error")
	}
}
