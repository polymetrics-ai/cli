package newsdata_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/newsdata"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the NewsData
// connector: apikey query auth, nextPage/page pagination over results[], and
// record mapping. Red until internal/connectors/newsdata exists.
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
			_, _ = w.Write([]byte(`{"status":"success","totalResults":3,"results":[{"article_id":"a1","title":"One","link":"https://x/1","pubDate":"2026-01-01 00:00:00"},{"article_id":"a2","title":"Two","link":"https://x/2","pubDate":"2026-01-02 00:00:00"}],"nextPage":"TOKEN2"}`))
		case "TOKEN2":
			_, _ = w.Write([]byte(`{"status":"success","totalResults":3,"results":[{"article_id":"a3","title":"Three","link":"https://x/3","pubDate":"2026-01-03 00:00:00"}],"nextPage":null}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"status":"success","results":[],"nextPage":null}`))
		}
	}))
	defer srv.Close()

	c := newsdata.New()
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
		t.Fatalf("apikey = %q, want pub_test_123", sawAPIKey)
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

// TestSourcesStreamMapping checks the sources stream maps id and name, and that
// a stream-specific endpoint is reached.
func TestSourcesStreamMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sources" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"status":"success","totalResults":1,"results":[{"id":"bbc","name":"BBC News","url":"https://bbc.co.uk","category":["top"],"language":["en"],"country":["united kingdom"]}]}`))
	}))
	defer srv.Close()

	c := newsdata.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "pub_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sources", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(sources): %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["id"] != "bbc" || got[0]["name"] != "BBC News" {
		t.Fatalf("source record mismatch: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records
// without any network access, so conformance works credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := newsdata.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"latest", "crypto", "sources"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := newsdata.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestRegistryResolvesNewsdata(t *testing.T) {
	_ = newsdata.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("newsdata"); !ok {
		t.Fatal("registry did not resolve newsdata (self-registration)")
	}
	c := newsdata.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := newsdata.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"latest": false, "crypto": false, "sources": false}
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

func TestBaseURLSSRFRejected(t *testing.T) {
	c := newsdata.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil"},
		Secrets: map[string]string{"api_key": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "latest", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected")
	}
}
