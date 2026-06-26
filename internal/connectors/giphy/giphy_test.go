package giphy_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/giphy"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Giphy
// connector: api_key query auth, offset/limit pagination over two pages using
// the response pagination object, and record mapping at data[]. Red until
// internal/connectors/giphy exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	var sawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.URL.Query().Get("api_key")
		sawQuery = r.URL.Query().Get("q")
		if r.URL.Path != "/gifs/search" {
			http.NotFound(w, r)
			return
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		switch offset {
		case 0:
			_, _ = w.Write([]byte(`{"data":[` +
				`{"id":"gif_1","type":"gif","title":"one","rating":"g","url":"https://giphy.com/gifs/gif_1"},` +
				`{"id":"gif_2","type":"gif","title":"two","rating":"pg","url":"https://giphy.com/gifs/gif_2"}` +
				`],"pagination":{"total_count":3,"count":2,"offset":0}}`))
		case 2:
			_, _ = w.Write([]byte(`{"data":[` +
				`{"id":"gif_3","type":"gif","title":"three","rating":"g","url":"https://giphy.com/gifs/gif_3"}` +
				`],"pagination":{"total_count":3,"count":1,"offset":2}}`))
		default:
			t.Errorf("unexpected offset=%d", offset)
			_, _ = w.Write([]byte(`{"data":[],"pagination":{"total_count":3,"count":0,"offset":` + strconv.Itoa(offset) + `}}`))
		}
	}))
	defer srv.Close()

	c := giphy.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "query": "cats", "page_size": "2"},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "gif_search", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "key_test_123" {
		t.Fatalf("api_key = %q, want key_test_123", sawAPIKey)
	}
	if sawQuery != "cats" {
		t.Fatalf("q = %q, want cats", sawQuery)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["type"] == nil {
			t.Fatalf("record missing id/type: %+v", rec)
		}
	}
}

// TestTrendingNoQuery verifies the trending stream does not require a query and
// still paginates and maps records.
func TestTrendingNoQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gifs/trending" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"t1","type":"gif","title":"trend"}],"pagination":{"total_count":1,"count":1,"offset":0}}`))
	}))
	defer srv.Close()

	c := giphy.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "25"},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "trending_gifs", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read trending: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("trending records = %d, want 1", len(got))
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network call so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := giphy.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "gif_search", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check is a no-op in fixture mode (no creds required).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = giphy.New() // ensure init ran
	c := giphy.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only API)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("giphy"); !ok {
		t.Fatal("registry did not resolve giphy (self-registration)")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := giphy.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "giphy" {
		t.Fatalf("catalog connector = %q, want giphy", cat.Connector)
	}
	want := map[string]bool{"gif_search": false, "sticker_search": false, "trending_gifs": false, "clip_search": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing expected stream %q", name)
		}
	}
}
