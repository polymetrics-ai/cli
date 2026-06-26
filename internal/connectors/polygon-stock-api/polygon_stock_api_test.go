package polygonstockapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadPaginatesAuthenticatesAndMapsRecords(t *testing.T) {
	var sawAuth string
	requests := 0
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v3/reference/tickers" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("limit") != "2" {
			t.Fatalf("limit = %q", r.URL.Query().Get("limit"))
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"results":[{"ticker":"AAPL","name":"Apple Inc.","market":"stocks","locale":"us"},{"ticker":"MSFT","name":"Microsoft","market":"stocks","locale":"us"}],"next_url":"` + srv.URL + `/v3/reference/tickers?cursor=next&limit=2"}`))
		case "next":
			_, _ = w.Write([]byte(`{"results":[{"ticker":"GOOG","name":"Alphabet","market":"stocks","locale":"us"}]}`))
		default:
			t.Fatalf("unexpected cursor %q", r.URL.Query().Get("cursor"))
		}
	}))
	defer srv.Close()

	c := New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "unit-token"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tickers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer unit-token" {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if requests != 2 || len(got) != 3 || got[0]["ticker"] != "AAPL" || got[0]["market"] != "stocks" {
		t.Fatalf("requests=%d records=%+v", requests, got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "dividends", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["ticker"] == nil {
		t.Fatalf("fixture records = %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) < 3 {
		t.Fatalf("Catalog = %+v, err=%v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("polygon-stock-api"); !ok {
		t.Fatal("registry did not resolve polygon-stock-api")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write error = %v", err)
	}
}
