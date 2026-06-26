package newsapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	newsapi "polymetrics.ai/internal/connectors/news-api"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the
// X-Api-Key header is sent, that page/pageSize pagination walks two pages of the
// `articles` array, and that each article is mapped (url primary key, publishedAt
// cursor) into a record.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("X-Api-Key")
		if r.URL.Path != "/everything" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "1":
			// pageSize=2 full page -> there is a next page.
			_, _ = w.Write([]byte(`{"status":"ok","totalResults":3,"articles":[` +
				`{"source":{"id":"the-verge","name":"The Verge"},"author":"A","title":"T1","url":"https://example.com/1","publishedAt":"2026-01-01T00:00:00Z","content":"c1"},` +
				`{"source":{"id":null,"name":"Wired"},"author":"B","title":"T2","url":"https://example.com/2","publishedAt":"2026-01-02T00:00:00Z","content":"c2"}` +
				`]}`))
		case "2":
			// short page -> stop after this.
			_, _ = w.Write([]byte(`{"status":"ok","totalResults":3,"articles":[` +
				`{"source":{"id":"bbc","name":"BBC"},"author":"C","title":"T3","url":"https://example.com/3","publishedAt":"2026-01-03T00:00:00Z","content":"c3"}` +
				`]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"status":"ok","totalResults":3,"articles":[]}`))
		}
	}))
	defer srv.Close()

	c := newsapi.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2", "search_query": "bitcoin"},
		Secrets: map[string]string{"api_key": "test_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "everything", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "test_key_123" {
		t.Fatalf("X-Api-Key = %q, want test_key_123", sawKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["url"] == nil || rec["published_at"] == nil {
			t.Fatalf("record missing url/published_at: %+v", rec)
		}
		if rec["source_name"] == nil {
			t.Fatalf("record missing flattened source_name: %+v", rec)
		}
	}
	if got[0]["source_id"] != "the-verge" {
		t.Fatalf("source_id = %v, want the-verge", got[0]["source_id"])
	}
}

// TestSourcesStreamReadsSourcesArray verifies the non-paginated sources endpoint
// reads from the `sources` array (different from `articles`) and maps id->id.
func TestSourcesStreamReadsSourcesArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/top-headlines/sources" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"status":"ok","sources":[` +
			`{"id":"bbc-news","name":"BBC News","description":"d","url":"https://bbc.co.uk","category":"general","language":"en","country":"gb"}` +
			`]}`))
	}))
	defer srv.Close()

	c := newsapi.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "k"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sources", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read sources: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "bbc-news" {
		t.Fatalf("sources records = %+v, want one with id bbc-news", got)
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no live credentials and no network.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := newsapi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "everything", Config: cfg}, func(rec connectors.Record) error {
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
		if rec["url"] == nil {
			t.Fatalf("fixture record missing url: %+v", rec)
		}
	}
	// Check must also succeed credential-free in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestRegistryResolvesAndMetadata(t *testing.T) {
	_ = newsapi.New() // ensure init ran
	c := newsapi.New()
	if c.Name() != "news-api" {
		t.Fatalf("Name = %q, want news-api", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
	if caps.Write {
		t.Fatalf("news-api is read-only, want Write=false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("news-api"); !ok {
		t.Fatal("registry did not resolve news-api (self-registration)")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
}
