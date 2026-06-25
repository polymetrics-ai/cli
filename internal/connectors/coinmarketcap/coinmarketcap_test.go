package coinmarketcap_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/coinmarketcap"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the
// CoinMarketCap X-CMC_PRO_API_KEY auth header, start/limit pagination across two
// pages of the listings/latest endpoint, and record mapping. CoinMarketCap
// returns {status:{...}, data:[...]} and paginates with a 1-based `start`
// cursor; a short page (fewer than `limit` records) ends the read.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("X-CMC_PRO_API_KEY")
		if r.URL.Path != "/v1/cryptocurrency/listings/latest" {
			http.NotFound(w, r)
			return
		}
		// limit=2 per page; page 1 starts at 1, page 2 starts at 3.
		switch r.URL.Query().Get("start") {
		case "", "1":
			_, _ = w.Write([]byte(`{"status":{"error_code":0},"data":[` +
				`{"id":1,"name":"Bitcoin","symbol":"BTC","cmc_rank":1},` +
				`{"id":1027,"name":"Ethereum","symbol":"ETH","cmc_rank":2}]}`))
		case "3":
			_, _ = w.Write([]byte(`{"status":{"error_code":0},"data":[` +
				`{"id":825,"name":"Tether","symbol":"USDT","cmc_rank":3}]}`))
		default:
			t.Errorf("unexpected start=%q", r.URL.Query().Get("start"))
			_, _ = w.Write([]byte(`{"status":{"error_code":0},"data":[]}`))
		}
	}))
	defer srv.Close()

	c := coinmarketcap.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "cmc_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "listings_latest", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "cmc_test_123" {
		t.Fatalf("X-CMC_PRO_API_KEY = %q, want cmc_test_123", sawKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["symbol"] != "BTC" || got[0]["name"] != "Bitcoin" {
		t.Fatalf("record[0] = %+v, want Bitcoin/BTC", got[0])
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

// TestReadSingleObjectStream covers global_metrics, whose `data` is a single
// object (not an array) and is therefore returned as one record.
func TestReadSingleObjectStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/global-metrics/quotes/latest" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"status":{"error_code":0},"data":{"active_cryptocurrencies":9000,"total_cryptocurrencies":12000}}`))
	}))
	defer srv.Close()

	c := coinmarketcap.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "cmc_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "global_metrics", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["active_cryptocurrencies"] == nil {
		t.Fatalf("record missing active_cryptocurrencies: %+v", got[0])
	}
}

// TestFixtureMode confirms a credential-free deterministic read for conformance.
func TestFixtureMode(t *testing.T) {
	c := coinmarketcap.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"map", "listings_latest", "categories", "fiat", "global_metrics"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %q produced no records", stream)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := coinmarketcap.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
}

func TestReadOnlyMetadataAndRegistry(t *testing.T) {
	_ = coinmarketcap.New() // ensure init ran
	c := coinmarketcap.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("coinmarketcap is read-only; Write should be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("coinmarketcap"); !ok {
		t.Fatal("registry did not resolve coinmarketcap (self-registration)")
	}
}

// TestBaseURLValidation rejects non-http(s) base_url overrides (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := coinmarketcap.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "map", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for file:// base_url")
	}
}
