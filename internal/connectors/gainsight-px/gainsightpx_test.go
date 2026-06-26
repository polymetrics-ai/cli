package gainsightpx_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	gainsightpx "polymetrics.ai/internal/connectors/gainsight-px"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Gainsight PX
// connector: X-APTRINSIC-API-KEY header auth, scrollId pagination over two pages
// of {"accounts":[...],"scrollId":...}, and record mapping. Red until the
// internal/connectors/gainsight-px package exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("X-APTRINSIC-API-KEY")
		if r.URL.Path != "/accounts" {
			http.NotFound(w, r)
			return
		}
		calls++
		switch r.URL.Query().Get("scrollId") {
		case "":
			_, _ = w.Write([]byte(`{"accounts":[{"id":"acc_1","name":"Acme"},{"id":"acc_2","name":"Globex"}],"scrollId":"page2"}`))
		case "page2":
			_, _ = w.Write([]byte(`{"accounts":[{"id":"acc_3","name":"Initech"}],"scrollId":""}`))
		default:
			t.Errorf("unexpected scrollId=%q", r.URL.Query().Get("scrollId"))
			_, _ = w.Write([]byte(`{"accounts":[],"scrollId":""}`))
		}
	}))
	defer srv.Close()

	c := gainsightpx.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "px_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "px_test_123" {
		t.Fatalf("X-APTRINSIC-API-KEY = %q, want px_test_123", sawKey)
	}
	if calls != 2 {
		t.Fatalf("server calls = %d, want 2 (two pages)", calls)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (two pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestFixtureModeNeedsNoNetwork verifies fixture mode emits deterministic records
// without any network access so conformance can run without live credentials.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := gainsightpx.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"accounts", "users", "feature", "segments"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) returned no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestUnknownStreamRejected ensures unknown streams are an error, not a silent
// empty read.
func TestUnknownStreamRejected(t *testing.T) {
	c := gainsightpx.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "does_not_exist", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read(unknown stream) should return an error")
	}
}

// TestCatalogAndRegistry asserts the published catalog and self-registration.
func TestCatalogAndRegistry(t *testing.T) {
	c := gainsightpx.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("gainsight-px is read-only; Write should be false")
	}

	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}

	r := connectors.NewRegistry()
	if _, ok := r.Get("gainsight-px"); !ok {
		t.Fatal("registry did not resolve gainsight-px (self-registration)")
	}
}
