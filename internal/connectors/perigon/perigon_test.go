package perigon_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/perigon"
)

func TestReadArticlesPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("apiKey")
		pages = append(pages, r.URL.Query().Get("page"))
		if r.URL.Path != "/v1/articles/all" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`{"articles":[{"articleId":"a1","title":"One","pubDate":"2026-01-01T00:00:00Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"articles":[]}`))
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := perigon.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1"}, Secrets: map[string]string{"api_key": "perigon_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "articles", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "perigon_key" {
		t.Fatalf("apiKey = %q, want perigon_key", sawKey)
	}
	if len(got) != 1 || got[0]["article_id"] != "a1" || got[0]["title"] != "One" {
		t.Fatalf("records = %+v, want mapped article", got)
	}
	if len(pages) != 2 {
		t.Fatalf("pages = %v, want two requests", pages)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := perigon.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "articles", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["article_id"] == nil {
		t.Fatalf("fixture records = %+v, want article_id", got)
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "perigon" || len(cat.Streams) < 1 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("perigon"); !ok {
		t.Fatal("registry did not resolve perigon")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
