package finage_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/finage"
)

// TestReadMostActivesAuthenticates is the red-first test for the Finage
// connector: apikey query-parameter auth, top-level array record extraction, and
// record mapping for the most_active_us_stocks stream.
func TestReadMostActivesAuthenticates(t *testing.T) {
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("apikey")
		if r.URL.Path != "/fnd/market-information/us/most-actives" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"symbol":"AAPL","company_name":"Apple Inc.","change":1.5,"change_percentage":"0.8%","price":"190.0"},{"symbol":"TSLA","company_name":"Tesla","change":-2.0,"change_percentage":"-1.0%","price":"250.0"}]`))
	}))
	defer srv.Close()

	c := finage.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "fin_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "most_active_us_stocks", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "fin_test_123" {
		t.Fatalf("apikey = %q, want fin_test_123", sawKey)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["symbol"] != "AAPL" || got[0]["company_name"] != "Apple Inc." {
		t.Fatalf("record[0] mapped wrong: %+v", got[0])
	}
}

// TestReadMarketNewsPaginatesOverSymbols asserts that the symbol-partitioned
// market_news stream issues one request per configured symbol (two here, the
// "two pages" requirement) and extracts the nested news[] array.
func TestReadMarketNewsPaginatesOverSymbols(t *testing.T) {
	var paths []string
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("apikey")
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/news/market/AAPL":
			_, _ = w.Write([]byte(`{"symbol":"AAPL","news":[{"title":"Apple A","url":"https://x/a","source":"src","date":"2026-01-01"}]}`))
		case "/news/market/TSLA":
			_, _ = w.Write([]byte(`{"symbol":"TSLA","news":[{"title":"Tesla B","url":"https://x/b","source":"src","date":"2026-01-02"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := finage.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "symbols": "AAPL,TSLA"},
		Secrets: map[string]string{"api_key": "fin_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "market_news", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "fin_test_123" {
		t.Fatalf("apikey = %q, want fin_test_123", sawKey)
	}
	if len(paths) != 2 || paths[0] != "/news/market/AAPL" || paths[1] != "/news/market/TSLA" {
		t.Fatalf("paths = %v, want one request per symbol", paths)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (one news item per symbol)", len(got))
	}
	if got[0]["title"] != "Apple A" || got[0]["symbol"] != "AAPL" {
		t.Fatalf("record[0] mapped wrong: %+v", got[0])
	}
}

// TestFixtureModeNeedsNoNetwork confirms fixture mode emits deterministic records
// without any credentials or network, so conformance runs credential-free.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := finage.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "most_gainers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["symbol"] == nil {
		t.Fatalf("fixture record missing symbol: %+v", got[0])
	}
}

func TestCheckFixtureNoNetwork(t *testing.T) {
	c := finage.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := finage.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "finage" {
		t.Fatalf("catalog connector = %q, want finage", cat.Connector)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	want := map[string]bool{"most_active_us_stocks": false, "market_news": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing expected stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = finage.New() // ensure init ran
	caps := finage.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("finage is read-only; Write should be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("finage"); !ok {
		t.Fatal("registry did not resolve finage (self-registration)")
	}
}
