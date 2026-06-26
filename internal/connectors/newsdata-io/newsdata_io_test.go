package newsdataio_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	newsdataio "polymetrics.ai/internal/connectors/newsdata-io"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the NewsData.io
// connector: apikey query-param auth, nextPage/page body-token pagination over
// results[], and record mapping across two pages.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.URL.Query().Get("apikey")
		if r.URL.Path != "/latest" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "":
			_, _ = w.Write([]byte(`{"status":"success","totalResults":3,"results":[{"article_id":"a1","title":"First","pubDate":"2026-06-01 10:00:00"},{"article_id":"a2","title":"Second","pubDate":"2026-06-01 11:00:00"}],"nextPage":"PAGE2TOKEN"}`))
		case "PAGE2TOKEN":
			_, _ = w.Write([]byte(`{"status":"success","totalResults":3,"results":[{"article_id":"a3","title":"Third","pubDate":"2026-06-01 12:00:00"}],"nextPage":null}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"status":"success","results":[],"nextPage":null}`))
		}
	}))
	defer srv.Close()

	c := newsdataio.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "pub_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "latest", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "pub_test_123" {
		t.Fatalf("apikey query = %q, want pub_test_123", sawAPIKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["article_id"] == nil || rec["title"] == nil {
			t.Fatalf("record missing article_id/title: %+v", rec)
		}
	}
}

// TestFixtureMode confirms the connector emits deterministic records with no
// network and no credentials (conformance harness path).
func TestFixtureMode(t *testing.T) {
	c := newsdataio.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "latest", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["article_id"] == nil {
		t.Fatalf("fixture record missing article_id: %+v", got[0])
	}
	// Check should not need network in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := newsdataio.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "newsdata-io" {
		t.Fatalf("catalog connector = %q, want newsdata-io", cat.Connector)
	}
	want := map[string]bool{"latest": false, "crypto": false, "archive": false, "sources": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing expected stream %q", name)
		}
	}
}

func TestRegistryResolves(t *testing.T) {
	_ = newsdataio.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("newsdata-io"); !ok {
		t.Fatal("registry did not resolve newsdata-io (self-registration)")
	}
	c := newsdataio.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only news feed)", caps)
	}
}

func TestMissingAPIKeyRejected(t *testing.T) {
	c := newsdataio.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check should fail without api_key in non-fixture mode")
	}
}
