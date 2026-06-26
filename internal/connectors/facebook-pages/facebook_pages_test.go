package facebookpages_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	facebookpages "polymetrics.ai/internal/connectors/facebook-pages"
)

func TestReadPostsPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v19.0/page_123/posts" {
			http.NotFound(w, r)
			return
		}
		pages = append(pages, r.URL.Query().Get("after"))
		switch r.URL.Query().Get("after") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"post_1","message":"hello","created_time":"2026-01-01T00:00:00+0000"}],"paging":{"next":"` + srvURL(r) + `/v19.0/page_123/posts?after=next"}}`))
		case "next":
			_, _ = w.Write([]byte(`{"data":[{"id":"post_2","permalink_url":"https://example.com/post/2","updated_time":"2026-01-02T00:00:00+0000"}]}`))
		default:
			t.Fatalf("unexpected after=%q", r.URL.Query().Get("after"))
		}
	}))
	defer srv.Close()

	c := facebookpages.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/v19.0", "page_id": "page_123", "page_size": "1"},
		Secrets: map[string]string{"access_token": "test-token"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "posts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer test-token" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if len(pages) != 2 {
		t.Fatalf("pages = %d, want 2", len(pages))
	}
	if len(got) != 2 || got[0]["id"] != "post_1" || got[1]["permalink_url"] != "https://example.com/post/2" {
		t.Fatalf("records not mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistryAndWrite(t *testing.T) {
	c := facebookpages.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var n int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "posts", Config: cfg}, func(connectors.Record) error { n++; return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if n == 0 {
		t.Fatal("fixture Read emitted no records")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) < 2 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write err = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("facebook-pages"); !ok {
		t.Fatal("registry did not resolve facebook-pages")
	}
}

func srvURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
