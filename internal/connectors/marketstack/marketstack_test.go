package marketstack_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/marketstack"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Marketstack
// connector: api-key-in-query auth (access_key), offset/limit pagination over
// two pages of data[], and record mapping. Red until
// internal/connectors/marketstack exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("access_key")
		if r.URL.Path != "/eod" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			// Full page (limit=2) -> there is a next page.
			_, _ = w.Write([]byte(`{"pagination":{"limit":2,"offset":0,"count":2,"total":3},"data":[{"symbol":"AAPL","date":"2026-01-02T00:00:00+0000","close":185.5},{"symbol":"AAPL","date":"2026-01-03T00:00:00+0000","close":186.1}]}`))
		case "2":
			// Short page -> last page.
			_, _ = w.Write([]byte(`{"pagination":{"limit":2,"offset":2,"count":1,"total":3},"data":[{"symbol":"AAPL","date":"2026-01-04T00:00:00+0000","close":187.0}]}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"pagination":{"limit":2,"offset":0,"count":0,"total":3},"data":[]}`))
		}
	}))
	defer srv.Close()

	c := marketstack.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2", "symbols": "AAPL"},
		Secrets: map[string]string{"api_key": "ms_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "eod", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "ms_test_123" {
		t.Fatalf("access_key = %q, want ms_test_123", sawKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["symbol"] == nil || rec["date"] == nil {
			t.Fatalf("record missing symbol/date: %+v", rec)
		}
	}
}

// TestReadExchangesMapsNestedFields verifies the flat exchanges stream maps the
// top-level data[] objects, including nested currency/timezone codes.
func TestReadExchangesMapsNestedFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/exchanges" {
			http.NotFound(w, r)
			return
		}
		// One short page so pagination stops immediately.
		_, _ = w.Write([]byte(`{"pagination":{"limit":1000,"offset":0,"count":1,"total":1},"data":[{"mic":"XNAS","name":"NASDAQ Stock Exchange","acronym":"NASDAQ","country":"USA","country_code":"US","city":"New York","currency":{"code":"USD","symbol":"$","name":"US Dollar"},"timezone":{"timezone":"America/New_York","abbr":"EST"}}]}`))
	}))
	defer srv.Close()

	c := marketstack.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "ms_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "exchanges", Config: cfg}, func(rec connectors.Record) error {
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
	if rec["mic"] != "XNAS" {
		t.Fatalf("mic = %v, want XNAS", rec["mic"])
	}
	if rec["currency_code"] != "USD" {
		t.Fatalf("currency_code = %v, want USD (nested flatten)", rec["currency_code"])
	}
	if rec["timezone"] != "America/New_York" {
		t.Fatalf("timezone = %v, want America/New_York (nested flatten)", rec["timezone"])
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// with no network access so conformance can run without credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := marketstack.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"exchanges", "tickers", "eod", "splits", "dividends"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("Read(%s) fixture records = %d, want 2", stream, len(got))
		}
	}

	// Check in fixture mode must not require credentials or network.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := marketstack.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"exchanges": true, "tickers": true, "eod": true, "splits": true, "dividends": true}
	if len(cat.Streams) != len(want) {
		t.Fatalf("streams = %d, want %d", len(cat.Streams), len(want))
	}
	for _, s := range cat.Streams {
		if !want[s.Name] {
			t.Fatalf("unexpected stream %q", s.Name)
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q missing fields", s.Name)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := marketstack.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "ms_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "exchanges", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected (SSRF guard)")
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = marketstack.New() // ensure init ran
	c := marketstack.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("marketstack"); !ok {
		t.Fatal("registry did not resolve marketstack (self-registration)")
	}
}
