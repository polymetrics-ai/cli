package watchmode_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/watchmode"
)

func TestReadSearchUsesQueryKeyAndMapsResults(t *testing.T) {
	var sawKey bool
	var sawSearchValue string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/search/" {
			http.NotFound(w, r)
			return
		}
		sawKey = r.URL.Query().Get("apiKey") != ""
		sawSearchValue = r.URL.Query().Get("search_value")
		_, _ = w.Write([]byte(`{"title_results":[{"id":101,"name":"Terminator","type":"movie"}]}`))
	}))
	defer srv.Close()

	c := watchmode.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "search_val": "Terminator"},
		Secrets: map[string]string{"api_key": "x"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "search", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawKey {
		t.Fatal("request did not include apiKey query parameter")
	}
	if sawSearchValue != "Terminator" {
		t.Fatalf("search_value = %q", sawSearchValue)
	}
	if len(got) != 1 || got[0]["id"] == nil || got[0]["name"] != "Terminator" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := watchmode.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "search", Config: cfg}, func(rec connectors.Record) error {
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
	if cat.Connector != "watchmode" || len(cat.Streams) == 0 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	if _, ok := connectors.NewRegistry().Get("watchmode"); !ok {
		t.Fatal("registry did not resolve watchmode")
	}
	if c.Metadata().Capabilities.Write {
		t.Fatal("connector should be read-only")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
