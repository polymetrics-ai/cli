package stockdata_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/stockdata"
)

func TestReadEODPricesAuthenticatesWithQueryAndMapsData(t *testing.T) {
	var sawToken bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.URL.Query().Get("api_token") == "test-token"
		if r.URL.Path != "/data/eod" {
			http.NotFound(w, r)
			return
		}
		if got := r.URL.Query().Get("symbols"); got != "AAPL" {
			t.Fatalf("symbols = %q, want AAPL", got)
		}
		_, _ = w.Write([]byte(`{"data":[{"ticker":"AAPL","date":"2026-01-01","close":123.45}]}`))
	}))
	defer srv.Close()

	c := stockdata.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "symbols": "AAPL"}, Secrets: map[string]string{"api_token": "test-token"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "eod_prices", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawToken {
		t.Fatal("request did not include api_token query auth")
	}
	if len(got) != 1 || got[0]["ticker"] != "AAPL" || got[0]["date"] != "2026-01-01" {
		t.Fatalf("records = %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := stockdata.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil || cat.Connector != "stockdata" || len(cat.Streams) < 3 {
		t.Fatalf("Catalog = %+v err=%v", cat, err)
	}
	var fixture []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Config: cfg}, func(rec connectors.Record) error {
		fixture = append(fixture, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(fixture) == 0 || fixture[0]["fixture"] != true {
		t.Fatalf("fixture records = %+v", fixture)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write err = %v", err)
	}
	if got, ok := connectors.NewRegistry().Get("stockdata"); !ok || got.Metadata().Capabilities.Write {
		t.Fatalf("registry/capabilities failed: ok=%v got=%T", ok, got)
	}
}
