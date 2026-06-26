package uscensus_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	uscensus "polymetrics.ai/internal/connectors/us-census"
)

func TestReadQueryAddsAPIKeyAndMapsHeaderRows(t *testing.T) {
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("key")
		if r.URL.Path != "/data/2019/cbp" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("get") != "NAME,ESTAB" || r.URL.Query().Get("for") != "us:*" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`[["NAME","ESTAB"],["United States","1"]]`))
	}))
	defer srv.Close()

	c := uscensus.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "query_path": "data/2019/cbp", "query_params": "get=NAME,ESTAB&for=us:*"}, Secrets: map[string]string{"api_key": "dummy-key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "query", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "dummy-key" {
		t.Fatalf("key query = %q", sawKey)
	}
	if len(got) != 1 || got[0]["name"] != "United States" || got[0]["estab"] != "1" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := uscensus.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "query", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["name"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "us-census" || len(cat.Streams) == 0 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	if _, ok := connectors.NewRegistry().Get("us-census"); !ok {
		t.Fatal("registry did not resolve us-census")
	}
	if caps := c.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want read-only Check/Catalog/Read", caps)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
