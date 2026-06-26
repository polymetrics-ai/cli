package financialmodelling_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	financialmodelling "polymetrics.ai/internal/connectors/financial-modelling"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the
// Financial Modeling Prep API key is injected as the apikey query parameter,
// that the connector paginates across two pages of the limit/offset-style
// stock screener, and that records are mapped from the top-level JSON array.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	var pages int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.URL.Query().Get("apikey")
		if r.URL.Path != "/stock-screener" {
			http.NotFound(w, r)
			return
		}
		pages++
		switch r.URL.Query().Get("offset") {
		case "", "0":
			// A full page (== limit) signals there is another page.
			_, _ = w.Write([]byte(`[{"symbol":"AAPL","companyName":"Apple Inc.","marketCap":3000000000000},{"symbol":"MSFT","companyName":"Microsoft","marketCap":2500000000000}]`))
		case "2":
			// A short page ends pagination.
			_, _ = w.Write([]byte(`[{"symbol":"GOOG","companyName":"Alphabet","marketCap":1800000000000}]`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := financialmodelling.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "fmp_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "stock_screener", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "fmp_test_key" {
		t.Fatalf("apikey query = %q, want fmp_test_key", sawAPIKey)
	}
	if pages != 2 {
		t.Fatalf("server saw %d pages, want 2", pages)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["symbol"] == nil {
			t.Fatalf("record missing symbol: %+v", rec)
		}
	}
}

// TestReadStocksList exercises a list endpoint that returns a single top-level
// array (no pagination) and confirms record mapping.
func TestReadStocksList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/stock/list" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"symbol":"AAPL","name":"Apple Inc.","price":195.5,"exchange":"NASDAQ","exchangeShortName":"NASDAQ","type":"stock"}]`))
	}))
	defer srv.Close()

	c := financialmodelling.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "fmp_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "stocks", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["symbol"] != "AAPL" || got[0]["exchange_short_name"] != "NASDAQ" {
		t.Fatalf("unexpected mapped record: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access so conformance runs without credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := financialmodelling.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "stocks", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["symbol"] == nil {
			t.Fatalf("fixture record missing symbol: %+v", rec)
		}
	}

	// Check must short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := financialmodelling.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"stocks": false, "etfs": false, "stock_screener": false, "delisted_companies": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestRegistryResolves confirms self-registration via NewRegistry().
func TestRegistryResolves(t *testing.T) {
	_ = financialmodelling.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("financial-modelling"); !ok {
		t.Fatal("registry did not resolve financial-modelling (self-registration)")
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs to bound SSRF.
func TestBaseURLValidation(t *testing.T) {
	c := financialmodelling.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "stocks", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected base_url scheme validation error")
	}
}
