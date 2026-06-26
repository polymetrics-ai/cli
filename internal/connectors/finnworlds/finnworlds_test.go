package finnworlds_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/finnworlds"
)

// TestReadPartitionsAndAuthenticates is the red-first test for the Finnworlds
// connector. Finnworlds authenticates with an API key query param (`key`) and has
// no page-token pagination; instead a read fans out across the configured
// `tickers` list, one request per ticker. The test asserts: the key query param
// is sent, the connector fans out over 2 tickers (2 "pages"), and records are
// extracted from the nested result.output.dividends path with the partition value
// stitched in. Red until internal/connectors/finnworlds exists.
func TestReadPartitionsAndAuthenticates(t *testing.T) {
	var sawKey string
	seenTickers := map[string]bool{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("key")
		ticker := r.URL.Query().Get("ticker")
		seenTickers[ticker] = true
		if !strings.HasSuffix(r.URL.Path, "/dividends") {
			http.NotFound(w, r)
			return
		}
		switch ticker {
		case "AAPL":
			_, _ = w.Write([]byte(`{"result":{"output":{"dividends":[{"date":"2026-01-01","dividend_rate":"0.24"},{"date":"2026-04-01","dividend_rate":"0.25"}]}}}`))
		case "MSFT":
			_, _ = w.Write([]byte(`{"result":{"output":{"dividends":[{"date":"2026-02-01","dividend_rate":"0.75"}]}}}`))
		default:
			t.Errorf("unexpected ticker=%q", ticker)
			_, _ = w.Write([]byte(`{"result":{"output":{"dividends":[]}}}`))
		}
	}))
	defer srv.Close()

	c := finnworlds.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "tickers": "AAPL,MSFT"},
		Secrets: map[string]string{"key": "fw_secret_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "dividends", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "fw_secret_123" {
		t.Fatalf("key query = %q, want fw_secret_123", sawKey)
	}
	if !seenTickers["AAPL"] || !seenTickers["MSFT"] {
		t.Fatalf("expected both tickers requested, saw %v", seenTickers)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 partitions)", len(got))
	}
	tickerCount := 0
	for _, rec := range got {
		if rec["date"] == nil {
			t.Fatalf("record missing date: %+v", rec)
		}
		// AddFields-style stitch: the partition ticker is attached to each record.
		if tk, _ := rec["ticker"].(string); tk == "AAPL" || tk == "MSFT" {
			tickerCount++
		}
	}
	if tickerCount != 3 {
		t.Fatalf("expected ticker stitched onto all 3 records, got %d", tickerCount)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = finnworlds.New() // ensure init ran
	c := finnworlds.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("finnworlds is read-only, Write should be false: %+v", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("finnworlds"); !ok {
		t.Fatal("registry did not resolve finnworlds (self-registration)")
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := finnworlds.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "commodities", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode should emit deterministic records")
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}
