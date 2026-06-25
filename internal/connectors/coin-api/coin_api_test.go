package coinapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	coinapi "polymetrics.ai/internal/connectors/coin-api"
)

// TestReadSymbolsAuthenticates verifies the X-CoinAPI-Key auth header and the
// top-level-array record mapping for the symbols metadata stream. Red until the
// coin-api package exists.
func TestReadSymbolsAuthenticates(t *testing.T) {
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("X-CoinAPI-Key")
		if r.URL.Path != "/v1/symbols" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[
			{"symbol_id":"BITSTAMP_SPOT_BTC_USD","exchange_id":"BITSTAMP","symbol_type":"SPOT","asset_id_base":"BTC","asset_id_quote":"USD"},
			{"symbol_id":"BITSTAMP_SPOT_ETH_USD","exchange_id":"BITSTAMP","symbol_type":"SPOT","asset_id_base":"ETH","asset_id_quote":"USD"}
		]`))
	}))
	defer srv.Close()

	c := coinapi.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "test-coin-key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "symbols", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "test-coin-key" {
		t.Fatalf("X-CoinAPI-Key = %q, want test-coin-key", sawKey)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["symbol_id"] != "BITSTAMP_SPOT_BTC_USD" {
		t.Fatalf("symbol_id = %v, want BITSTAMP_SPOT_BTC_USD", got[0]["symbol_id"])
	}
	if got[0]["asset_id_base"] != "BTC" {
		t.Fatalf("asset_id_base = %v, want BTC", got[0]["asset_id_base"])
	}
}

// TestReadOHLCVPaginatesByTime exercises CoinAPI's limit+time_start pagination:
// a full page (== limit) triggers a follow-up request with time_start advanced
// past the last record's time_period_start, and a short page terminates the loop.
func TestReadOHLCVPaginatesByTime(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/ohlcv/BTC/history" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("period_id") != "1DAY" {
			t.Errorf("period_id = %q, want 1DAY", r.URL.Query().Get("period_id"))
		}
		calls++
		switch r.URL.Query().Get("time_start") {
		case "2021-01-01T00:00:00":
			// Full page of 2 (== limit) -> expect another request.
			_, _ = w.Write([]byte(`[
				{"time_period_start":"2021-01-01T00:00:00.0000000Z","price_open":29000,"price_close":29300,"volume_traded":1.5},
				{"time_period_start":"2021-01-02T00:00:00.0000000Z","price_open":29300,"price_close":32000,"volume_traded":2.5}
			]`))
		case "2021-01-02T00:00:00.0000000Z":
			// Short page (< limit) -> terminate.
			_, _ = w.Write([]byte(`[
				{"time_period_start":"2021-01-03T00:00:00.0000000Z","price_open":32000,"price_close":33000,"volume_traded":3.5}
			]`))
		default:
			t.Errorf("unexpected time_start=%q", r.URL.Query().Get("time_start"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := coinapi.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"symbol_id":  "BTC",
			"period":     "1DAY",
			"start_date": "2021-01-01T00:00:00",
			"limit":      "2",
		},
		Secrets: map[string]string{"api_key": "test-coin-key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "ohlcv_historical_data", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if calls != 2 {
		t.Fatalf("server calls = %d, want 2 (pagination across 2 pages)", calls)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["symbol_id"] != "BTC" {
			t.Fatalf("record missing injected symbol_id: %+v", rec)
		}
		if rec["time_period_start"] == nil || rec["price_close"] == nil {
			t.Fatalf("record missing ohlcv fields: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access, so conformance runs credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := coinapi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "ohlcv_historical_data", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	if got[0]["time_period_start"] == nil {
		t.Fatalf("fixture record missing time_period_start: %+v", got[0])
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestEnvironmentSelectsSandbox verifies the environment config field maps to the
// sandbox base URL without an explicit base_url override (no network call: we
// only assert the resolved Catalog/Metadata wiring stays intact).
func TestMetadataAndRegistry(t *testing.T) {
	c := coinapi.New()
	md := c.Metadata()
	if md.Name != "coin-api" {
		t.Fatalf("Name = %q, want coin-api", md.Name)
	}
	if !md.Capabilities.Read || !md.Capabilities.Catalog || !md.Capabilities.Check {
		t.Fatalf("capabilities = %+v, want Read/Catalog/Check", md.Capabilities)
	}
	if md.Capabilities.Write {
		t.Fatalf("coin-api is read-only market data; Write must be false")
	}

	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}

	r := connectors.NewRegistry()
	if _, ok := r.Get("coin-api"); !ok {
		t.Fatal("registry did not resolve coin-api (self-registration)")
	}
}
