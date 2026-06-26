package gnews_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/gnews"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the GNews
// connector: apikey query-param auth, page-number pagination over articles[],
// and record mapping (including the flattened nested source object). Red until
// internal/connectors/gnews exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	var sawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("apikey")
		sawQuery = r.URL.Query().Get("q")
		if r.URL.Path != "/search" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"totalArticles":3,"articles":[
				{"id":"a1","title":"First","url":"https://ex.com/1","publishedAt":"2026-01-01T00:00:00Z","source":{"name":"Src One","url":"https://src.one"}},
				{"id":"a2","title":"Second","url":"https://ex.com/2","publishedAt":"2026-01-02T00:00:00Z","source":{"name":"Src Two","url":"https://src.two"}}
			]}`))
		case "2":
			_, _ = w.Write([]byte(`{"totalArticles":3,"articles":[
				{"id":"a3","title":"Third","url":"https://ex.com/3","publishedAt":"2026-01-03T00:00:00Z","source":{"name":"Src Three","url":"https://src.three"}}
			]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"totalArticles":3,"articles":[]}`))
		}
	}))
	defer srv.Close()

	c := gnews.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"query":     "Apple",
			"page_size": "2", // force a short final page -> stop after page 2
		},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "search", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "key_test_123" {
		t.Fatalf("apikey = %q, want key_test_123", sawKey)
	}
	if sawQuery != "Apple" {
		t.Fatalf("q = %q, want Apple", sawQuery)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	// Record mapping: id, url, publishedAt, and the flattened source fields.
	first := got[0]
	if first["id"] != "a1" || first["url"] != "https://ex.com/1" {
		t.Fatalf("record mapping wrong: %+v", first)
	}
	if first["source_name"] != "Src One" || first["source_url"] != "https://src.one" {
		t.Fatalf("nested source not flattened: %+v", first)
	}
	if first["published_at"] != "2026-01-01T00:00:00Z" {
		t.Fatalf("published_at mapping wrong: %+v", first)
	}
}

func TestTopHeadlinesStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/top-headlines" {
			http.NotFound(w, r)
			return
		}
		if topic := r.URL.Query().Get("topic"); topic != "technology" {
			t.Errorf("topic = %q, want technology", topic)
		}
		_, _ = w.Write([]byte(`{"totalArticles":1,"articles":[
			{"id":"h1","title":"Headline","url":"https://ex.com/h1","publishedAt":"2026-02-01T00:00:00Z","source":{"name":"HSrc"}}
		]}`))
	}))
	defer srv.Close()

	c := gnews.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":            srv.URL,
			"top_headlines_topic": "technology",
		},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "top_headlines", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "h1" {
		t.Fatalf("top_headlines records = %+v", got)
	}
}

func TestFixtureMode(t *testing.T) {
	c := gnews.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "search", Config: cfg}, func(rec connectors.Record) error {
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
		if rec["id"] == nil || rec["url"] == nil {
			t.Fatalf("fixture record missing id/url: %+v", rec)
		}
	}
}

func TestCatalogAndRegistry(t *testing.T) {
	_ = gnews.New() // ensure init ran
	c := gnews.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read/Catalog/Check", caps)
	}
	if caps.Write {
		t.Fatalf("gnews is read-only; Write should be false, got %+v", caps)
	}

	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 2 {
		t.Fatalf("expected at least 2 streams, got %d", len(cat.Streams))
	}

	r := connectors.NewRegistry()
	if _, ok := r.Get("gnews"); !ok {
		t.Fatal("registry did not resolve gnews (self-registration)")
	}
}
