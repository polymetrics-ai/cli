package wikipediapageviews_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	wikipediapageviews "polymetrics.ai/internal/connectors/wikipedia-pageviews"
)

func TestReadPageviewsBuildsPerArticlePath(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		if r.URL.Path != "/api/rest_v1/metrics/pageviews/per-article/en.wikipedia.org/all-access/user/Ada_Lovelace/daily/20260101/20260102" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"items":[{"project":"en.wikipedia.org","article":"Ada_Lovelace","timestamp":"2026010100","views":10}]}`))
	}))
	defer srv.Close()

	c := wikipediapageviews.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{
		"base_url": srv.URL,
		"project":  "en.wikipedia.org",
		"access":   "all-access",
		"agent":    "user",
		"article":  "Ada_Lovelace",
		"start":    "20260101",
		"end":      "20260102",
		"country":  "US",
	}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pageviews", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawPath == "" {
		t.Fatal("server did not receive request")
	}
	if len(got) != 1 || got[0]["article"] != "Ada_Lovelace" || got[0]["views"] == nil {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := wikipediapageviews.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pageviews", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "wikipedia-pageviews" || len(cat.Streams) == 0 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	if _, ok := connectors.NewRegistry().Get("wikipedia-pageviews"); !ok {
		t.Fatal("registry did not resolve wikipedia-pageviews")
	}
	if c.Metadata().Capabilities.Write {
		t.Fatal("connector should be read-only")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
