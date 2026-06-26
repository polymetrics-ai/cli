package searxng_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/searxng"
)

// TestSearchPaginatesNoAuth is the red-first test: the SearXNG connector must
// query {base}/search with format=json, paginate by pageno over results[], map
// each result to a flat record, and require NO credentials (public instances are
// open). Red until internal/connectors/searxng exists.
func TestSearchPaginatesNoAuth(t *testing.T) {
	var sawFormat, sawQuery string
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			http.NotFound(w, r)
			return
		}
		sawFormat = r.URL.Query().Get("format")
		sawQuery = r.URL.Query().Get("q")
		sawAuth = r.Header.Get("Authorization")
		switch r.URL.Query().Get("pageno") {
		case "", "1":
			_, _ = w.Write([]byte(`{"query":"go etl","number_of_results":3,"results":[
				{"url":"https://reddit.com/r/dataengineering/a","title":"First","content":"c1","engine":"reddit","engines":["reddit"],"score":1.5,"category":"general","publishedDate":"2026-01-01T00:00:00"},
				{"url":"https://reddit.com/r/dataengineering/b","title":"Second","content":"c2","engine":"duckduckgo","engines":["duckduckgo","brave"],"score":1.1,"category":"general"}
			]}`))
		case "2":
			_, _ = w.Write([]byte(`{"query":"go etl","number_of_results":3,"results":[
				{"url":"https://reddit.com/r/dataengineering/c","title":"Third","content":"c3","engine":"reddit","engines":["reddit"],"score":0.9,"category":"general"}
			]}`))
		default:
			t.Errorf("unexpected pageno=%q", r.URL.Query().Get("pageno"))
			_, _ = w.Write([]byte(`{"results":[]}`))
		}
	}))
	defer srv.Close()

	c := searxng.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{
		"base_url":  srv.URL,
		"query":     "go etl",
		"page_size": "2", // short final page (1 < 2) stops pagination
		"max_pages": "all",
	}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "search", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawFormat != "json" {
		t.Fatalf("format = %q, want json", sawFormat)
	}
	if sawQuery != "go etl" {
		t.Fatalf("q = %q, want 'go etl'", sawQuery)
	}
	if sawAuth != "" {
		t.Fatalf("Authorization should be empty without a secret, got %q", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (across 2 pages)", len(got))
	}
	first := got[0]
	if first["url"] != "https://reddit.com/r/dataengineering/a" || first["title"] != "First" {
		t.Fatalf("record mapping wrong: %+v", first)
	}
	if first["engine"] != "reddit" || first["content"] != "c1" {
		t.Fatalf("record fields wrong: %+v", first)
	}
	if first["published_date"] != "2026-01-01T00:00:00" {
		t.Fatalf("published_date mapping wrong: %+v", first)
	}
	// The []engines array is flattened to a comma-joined, warehouse-friendly string.
	if got[1]["engines"] != "duckduckgo,brave" {
		t.Fatalf("engines flattening wrong: %+v", got[1]["engines"])
	}
}

// TestRedditStreamScopesQuery verifies the convenience reddit stream targets
// reddit.com (and a subreddit when configured) so any general SearXNG instance
// returns Reddit results without depending on a reddit engine being installed.
func TestRedditStreamScopesQuery(t *testing.T) {
	var sawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawQuery = r.URL.Query().Get("q")
		_, _ = w.Write([]byte(`{"results":[{"url":"https://www.reddit.com/r/dataengineering/comments/x","title":"ETL pain","content":"...","engine":"google"}]}`))
	}))
	defer srv.Close()

	c := searxng.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{
		"base_url":  srv.URL,
		"query":     "best etl tool",
		"subreddit": "dataengineering",
	}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reddit", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !strings.Contains(sawQuery, "site:reddit.com/r/dataengineering") {
		t.Fatalf("reddit query not scoped to subreddit: %q", sawQuery)
	}
	if !strings.Contains(sawQuery, "best etl tool") {
		t.Fatalf("reddit query missing user terms: %q", sawQuery)
	}
	if len(got) != 1 || got[0]["title"] != "ETL pain" {
		t.Fatalf("reddit records = %+v", got)
	}
}

// TestOptionalBearerAuth verifies that when an optional api_key secret is set
// (for instances behind an auth proxy) it is sent as a Bearer header.
func TestOptionalBearerAuth(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"results":[{"url":"https://x/1","title":"t"}]}`))
	}))
	defer srv.Close()

	c := searxng.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "query": "x"},
		Secrets: map[string]string{"api_key": "tok_123"},
	}
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "search", Config: cfg}, func(connectors.Record) error { return nil }); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_123", sawAuth)
	}
}

func TestFixtureMode(t *testing.T) {
	c := searxng.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reddit", Config: cfg}, func(rec connectors.Record) error {
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
		if rec["url"] == nil || rec["title"] == nil {
			t.Fatalf("fixture record missing url/title: %+v", rec)
		}
	}
}

func TestCatalogAndRegistry(t *testing.T) {
	c := searxng.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read/Catalog/Check", caps)
	}
	if caps.Write {
		t.Fatalf("searxng is read-only; Write should be false, got %+v", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 2 {
		t.Fatalf("expected at least 2 streams, got %d", len(cat.Streams))
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("searxng"); !ok {
		t.Fatal("registry did not resolve searxng (self-registration)")
	}
}

// TestLiveRegistryResolves guards that searxng — a pm-native connector with no
// catalog_data.json entry — is still exposed by the production (live) registry
// that the CLI uses, via RegisterNativeLive. Red until that hook exists.
func TestLiveRegistryResolves(t *testing.T) {
	r := connectors.NewLiveRegistry()
	if _, ok := r.Get("searxng"); !ok {
		t.Fatal("live registry did not resolve searxng (RegisterNativeLive)")
	}
}

func TestRequiresBaseURL(t *testing.T) {
	c := searxng.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"query": "x"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "search", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error when base_url is missing in non-fixture mode")
	}
}
