package coingeckocoins_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	coingeckocoins "polymetrics.ai/internal/connectors/coingecko-coins"
)

// TestReadMarketChartAuthenticatesAndFlattens is the red-first test for the
// market_chart stream: the pro API key rides on the x-cg-pro-api-key header,
// the vs_currency/days query params are set, and the parallel
// prices/market_caps/total_volumes arrays of [timestamp, value] pairs are
// flattened into one record per timestamp carrying coin_id, vs_currency, and the
// three numeric series.
func TestReadMarketChartAuthenticatesAndFlattens(t *testing.T) {
	var sawKey, sawVsCurrency, sawDays, sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("x-cg-pro-api-key")
		sawVsCurrency = r.URL.Query().Get("vs_currency")
		sawDays = r.URL.Query().Get("days")
		sawPath = r.URL.Path
		if !strings.HasSuffix(r.URL.Path, "/coins/bitcoin/market_chart") {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{
			"prices":[[1700000000000,42000.5],[1700003600000,42100.0]],
			"market_caps":[[1700000000000,800000000000],[1700003600000,801000000000]],
			"total_volumes":[[1700000000000,25000000000],[1700003600000,26000000000]]
		}`))
	}))
	defer srv.Close()

	c := coingeckocoins.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":    srv.URL,
			"coin_id":     "bitcoin",
			"vs_currency": "usd",
			"days":        "7",
		},
		Secrets: map[string]string{"api_key": "cg_pro_secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "market_chart", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "cg_pro_secret" {
		t.Fatalf("x-cg-pro-api-key = %q, want cg_pro_secret", sawKey)
	}
	if sawVsCurrency != "usd" {
		t.Fatalf("vs_currency = %q, want usd", sawVsCurrency)
	}
	if sawDays != "7" {
		t.Fatalf("days = %q, want 7", sawDays)
	}
	if !strings.HasSuffix(sawPath, "/coins/bitcoin/market_chart") {
		t.Fatalf("path = %q, want .../coins/bitcoin/market_chart", sawPath)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (one per timestamp)", len(got))
	}
	first := got[0]
	if first["coin_id"] != "bitcoin" || first["vs_currency"] != "usd" {
		t.Fatalf("record missing coin_id/vs_currency: %+v", first)
	}
	if first["timestamp"] == nil || first["price"] == nil || first["market_cap"] == nil || first["total_volume"] == nil {
		t.Fatalf("record missing series fields: %+v", first)
	}
}

// TestReadHistoryPaginatesAcrossDates is the red-first pagination test: the
// history stream walks one /coins/{id}/history request per day from start_date
// to end_date (the date param is the cursor that advances the "pages"), and
// emits one snapshot record per date.
func TestReadHistoryPaginatesAcrossDates(t *testing.T) {
	var dates []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/coins/bitcoin/history") {
			http.NotFound(w, r)
			return
		}
		date := r.URL.Query().Get("date")
		dates = append(dates, date)
		_, _ = w.Write([]byte(`{
			"id":"bitcoin","symbol":"btc","name":"Bitcoin",
			"market_data":{
				"current_price":{"usd":42000.5,"eur":39000.0},
				"market_cap":{"usd":800000000000},
				"total_volume":{"usd":25000000000}
			}
		}`))
	}))
	defer srv.Close()

	c := coingeckocoins.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":    srv.URL,
			"coin_id":     "bitcoin",
			"vs_currency": "usd",
			"start_date":  "01-01-2026",
			"end_date":    "02-01-2026",
		},
		Secrets: map[string]string{"api_key": "cg_pro_secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "history", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(dates) != 2 {
		t.Fatalf("requested dates = %v, want 2 pages (01-01-2026, 02-01-2026)", dates)
	}
	if dates[0] != "01-01-2026" || dates[1] != "02-01-2026" {
		t.Fatalf("date pagination = %v, want [01-01-2026 02-01-2026]", dates)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (one snapshot per date)", len(got))
	}
	if got[0]["coin_id"] != "bitcoin" || got[0]["date"] != "01-01-2026" {
		t.Fatalf("record missing coin_id/date: %+v", got[0])
	}
	if got[0]["current_price"] == nil || got[0]["market_cap"] == nil {
		t.Fatalf("record missing market_data fields: %+v", got[0])
	}
}

// TestFixtureModeNeedsNoNetwork confirms fixture mode emits deterministic
// records with no live credentials or network access, so the conformance harness
// can exercise the connector credential-free.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := coingeckocoins.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"market_chart", "history", "coin"} {
		var n int
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(connectors.Record) error {
			n++
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if n == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogAndCapabilities verifies the published catalog and read-only
// capabilities.
func TestCatalogAndCapabilities(t *testing.T) {
	c := coingeckocoins.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "coingecko-coins" {
		t.Fatalf("catalog connector = %q, want coingecko-coins", cat.Connector)
	}
	want := map[string]bool{"market_chart": false, "history": false, "coin": false}
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

// TestRegistryResolution confirms self-registration: the connector resolves via
// the package-global registry under its bare system name.
func TestRegistryResolution(t *testing.T) {
	_ = coingeckocoins.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("coingecko-coins"); !ok {
		t.Fatal("registry did not resolve coingecko-coins (self-registration)")
	}
}
