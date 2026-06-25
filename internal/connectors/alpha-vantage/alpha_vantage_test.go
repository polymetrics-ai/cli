package alphavantage_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	alphavantage "polymetrics/internal/connectors/alpha-vantage"
)

// TestReadDailyAuthenticatesAndMaps is the red-first test: the apikey query
// parameter carries auth, the function/symbol query params are set, and each
// dated entry under "Time Series (Daily)" is flattened into a record with a
// date primary key and OHLCV fields. Alpha Vantage has no pagination; the
// response is one object keyed by date.
func TestReadDailyAuthenticatesAndMaps(t *testing.T) {
	var sawAPIKey, sawFunction, sawSymbol string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/query" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		sawAPIKey = q.Get("apikey")
		sawFunction = q.Get("function")
		sawSymbol = q.Get("symbol")
		_, _ = w.Write([]byte(`{
			"Meta Data": {"2. Symbol": "IBM"},
			"Time Series (Daily)": {
				"2026-06-25": {"1. open": "100.0", "2. high": "110.0", "3. low": "99.0", "4. close": "105.0", "5. volume": "12345"},
				"2026-06-24": {"1. open": "98.0", "2. high": "101.0", "3. low": "97.0", "4. close": "100.0", "5. volume": "54321"}
			}
		}`))
	}))
	defer srv.Close()

	c := alphavantage.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "symbol": "IBM"},
		Secrets: map[string]string{"api_key": "demo_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "time_series_daily", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "demo_key_123" {
		t.Fatalf("apikey = %q, want demo_key_123", sawAPIKey)
	}
	if sawFunction != "TIME_SERIES_DAILY" {
		t.Fatalf("function = %q, want TIME_SERIES_DAILY", sawFunction)
	}
	if sawSymbol != "IBM" {
		t.Fatalf("symbol = %q, want IBM", sawSymbol)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["date"] == nil || rec["symbol"] == nil {
			t.Fatalf("record missing date/symbol: %+v", rec)
		}
		if rec["open"] == nil || rec["close"] == nil || rec["volume"] == nil {
			t.Fatalf("record missing OHLCV: %+v", rec)
		}
	}
}

// TestReadGlobalQuote covers the single-object Global Quote stream, which is not
// a dated map but a flat object that should map to exactly one record.
func TestReadGlobalQuote(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("function"); got != "GLOBAL_QUOTE" {
			t.Errorf("function = %q, want GLOBAL_QUOTE", got)
		}
		_, _ = w.Write([]byte(`{
			"Global Quote": {
				"01. symbol": "IBM",
				"02. open": "100.0",
				"03. high": "110.0",
				"04. low": "99.0",
				"05. price": "105.0",
				"06. volume": "12345",
				"07. latest trading day": "2026-06-25",
				"08. previous close": "104.0",
				"09. change": "1.0",
				"10. change percent": "0.96%"
			}
		}`))
	}))
	defer srv.Close()

	c := alphavantage.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "symbol": "IBM"},
		Secrets: map[string]string{"api_key": "demo_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "global_quote", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	rec := got[0]
	if rec["symbol"] != "IBM" {
		t.Fatalf("symbol = %v, want IBM", rec["symbol"])
	}
	if rec["price"] != "105.0" {
		t.Fatalf("price = %v, want 105.0", rec["price"])
	}
	if rec["latest_trading_day"] != "2026-06-25" {
		t.Fatalf("latest_trading_day = %v, want 2026-06-25", rec["latest_trading_day"])
	}
}

// TestReadRejectsAPIErrorNote ensures an Alpha Vantage error/rate-limit payload
// (which returns HTTP 200 with a "Note"/"Error Message"/"Information" field) is
// surfaced as an error rather than silently emitting zero records.
func TestReadSurfacesAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Error Message": "Invalid API call. Please retry."}`))
	}))
	defer srv.Close()

	c := alphavantage.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "symbol": "BADSYM"},
		Secrets: map[string]string{"api_key": "demo_key_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "time_series_daily", Config: cfg}, func(connectors.Record) error {
		return nil
	})
	if err == nil {
		t.Fatal("Read should surface Alpha Vantage error payloads")
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access, so conformance runs without live credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := alphavantage.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"mode": "fixture", "symbol": "IBM"},
		// No secrets and no base_url: a network call would fail.
	}
	var count int
	for _, stream := range []string{"time_series_daily", "time_series_weekly", "time_series_monthly", "global_quote"} {
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			count++
			if rec["symbol"] == nil {
				t.Fatalf("fixture record missing symbol: %+v", rec)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
	}
	if count == 0 {
		t.Fatal("fixture mode emitted no records")
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := alphavantage.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestRegisteredReadOnly verifies self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	c := alphavantage.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (market data is read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("alpha-vantage"); !ok {
		t.Fatal("registry did not resolve alpha-vantage (self-registration)")
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := alphavantage.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "alpha-vantage" {
		t.Fatalf("catalog connector = %q, want alpha-vantage", cat.Connector)
	}
	want := map[string]bool{
		"time_series_daily":   false,
		"time_series_weekly":  false,
		"time_series_monthly": false,
		"global_quote":        false,
	}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}
