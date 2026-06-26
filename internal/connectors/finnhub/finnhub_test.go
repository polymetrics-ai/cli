package finnhub_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/finnhub"
)

// TestCompanyNewsAuthAndSymbolIteration is the red-first test for the Finnhub
// connector: it asserts the X-Finnhub-Token auth header, iteration across two
// configured symbols (Finnhub has no list pagination; per-symbol streams loop
// over the symbols config), and record mapping of the company_news array body.
func TestCompanyNewsAuthAndSymbolIteration(t *testing.T) {
	var sawToken string
	seen := map[string]bool{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.Header.Get("X-Finnhub-Token")
		if r.URL.Path != "/company-news" {
			http.NotFound(w, r)
			return
		}
		sym := r.URL.Query().Get("symbol")
		seen[sym] = true
		switch sym {
		case "AAPL":
			_, _ = w.Write([]byte(`[{"id":1,"category":"company","datetime":1700000000,"headline":"Apple A","source":"Reuters","related":"AAPL","url":"https://x/1"}]`))
		case "MSFT":
			_, _ = w.Write([]byte(`[{"id":2,"category":"company","datetime":1700000100,"headline":"MSFT B","source":"Reuters","related":"MSFT","url":"https://x/2"}]`))
		default:
			t.Errorf("unexpected symbol=%q", sym)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := finnhub.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"symbols":    "AAPL,MSFT",
			"start_date": "2023-01-01T00:00:00Z",
		},
		Secrets: map[string]string{"api_key": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "company_news", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "tok_123" {
		t.Fatalf("X-Finnhub-Token = %q, want tok_123", sawToken)
	}
	if !seen["AAPL"] || !seen["MSFT"] {
		t.Fatalf("expected both symbols queried, saw %+v", seen)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (one per symbol)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["headline"] == nil || rec["symbol"] == nil {
			t.Fatalf("record missing id/headline/symbol: %+v", rec)
		}
	}
}

// TestStockSymbolsExchange asserts the exchange-scoped stream maps the array body
// and forwards the exchange config as a query param.
func TestStockSymbolsExchange(t *testing.T) {
	var sawExchange string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/stock/symbol" {
			http.NotFound(w, r)
			return
		}
		sawExchange = r.URL.Query().Get("exchange")
		_, _ = w.Write([]byte(`[{"symbol":"AAPL","description":"APPLE INC","displaySymbol":"AAPL","type":"Common Stock","currency":"USD","mic":"XNAS"},{"symbol":"MSFT","description":"MICROSOFT","displaySymbol":"MSFT","type":"Common Stock","currency":"USD","mic":"XNAS"}]`))
	}))
	defer srv.Close()

	c := finnhub.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "exchange": "US"},
		Secrets: map[string]string{"api_key": "tok_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "stock_symbols", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawExchange != "US" {
		t.Fatalf("exchange = %q, want US", sawExchange)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["symbol"] != "AAPL" || got[0]["mic"] != "XNAS" {
		t.Fatalf("unexpected first record: %+v", got[0])
	}
}

// TestFixtureMode confirms credential-free fixture reads produce deterministic
// records for conformance.
func TestFixtureMode(t *testing.T) {
	c := finnhub.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "company_news", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode produced no records")
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogStreams asserts the published catalog covers the core streams.
func TestCatalogStreams(t *testing.T) {
	c := finnhub.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{
		"stock_symbols":         false,
		"company_news":          false,
		"market_news":           false,
		"company_profile":       false,
		"stock_recommendations": false,
	}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegistryResolves asserts self-registration via the connectors registry.
func TestRegistryResolves(t *testing.T) {
	_ = finnhub.New() // ensure init() ran
	caps := finnhub.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("finnhub"); !ok {
		t.Fatal("registry did not resolve finnhub (self-registration)")
	}
}
