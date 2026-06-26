package exchangerates_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"polymetrics.ai/internal/connectors"
	exchangerates "polymetrics.ai/internal/connectors/exchange-rates"
)

// TestReadExchangeRatesIteratesDatesAndAuthenticates is the red-first test: the
// access_key query param auth, date-by-date iteration (this API's form of
// pagination) over a start_date..end_date window, and record mapping that
// flattens the nested rates object. Red until the package exists.
func TestReadExchangeRatesIteratesDatesAndAuthenticates(t *testing.T) {
	var sawKey string
	var datePaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("access_key")
		switch r.URL.Path {
		case "/2026-01-01":
			datePaths = append(datePaths, r.URL.Path)
			_, _ = w.Write([]byte(`{"success":true,"historical":true,"timestamp":1767225600,"base":"EUR","date":"2026-01-01","rates":{"USD":1.05,"GBP":0.85}}`))
		case "/2026-01-02":
			datePaths = append(datePaths, r.URL.Path)
			_, _ = w.Write([]byte(`{"success":true,"historical":true,"timestamp":1767312000,"base":"EUR","date":"2026-01-02","rates":{"USD":1.06,"GBP":0.86}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := exchangerates.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"start_date": "2026-01-01",
			"end_date":   "2026-01-02",
			"base":       "EUR",
		},
		Secrets: map[string]string{"access_key": "test_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "exchange_rates", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "test_key_123" {
		t.Fatalf("access_key = %q, want test_key_123", sawKey)
	}
	if len(datePaths) != 2 {
		t.Fatalf("requested %d date pages, want 2: %v", len(datePaths), datePaths)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (two dates)", len(got))
	}
	// Records must carry the date (primary key / cursor) and flattened rate fields.
	dates := []string{}
	for _, rec := range got {
		d, _ := rec["date"].(string)
		dates = append(dates, d)
		if rec["base"] != "EUR" {
			t.Fatalf("record base = %v, want EUR", rec["base"])
		}
		rates, ok := rec["rates"].(map[string]any)
		if !ok {
			t.Fatalf("record rates not a map: %#v", rec["rates"])
		}
		if _, ok := rates["USD"]; !ok {
			t.Fatalf("record rates missing USD: %#v", rates)
		}
	}
	sort.Strings(dates)
	if dates[0] != "2026-01-01" || dates[1] != "2026-01-02" {
		t.Fatalf("dates = %v, want [2026-01-01 2026-01-02]", dates)
	}
}

// TestReadSymbols verifies the symbols stream maps each currency code into its
// own record using the same access_key query auth.
func TestReadSymbols(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/symbols" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("access_key") != "k" {
			t.Errorf("missing access_key on symbols request")
		}
		_, _ = w.Write([]byte(`{"success":true,"symbols":{"USD":"United States Dollar","EUR":"Euro"}}`))
	}))
	defer srv.Close()

	c := exchangerates.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "start_date": "2026-01-01"},
		Secrets: map[string]string{"access_key": "k"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "symbols", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read symbols: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("symbols records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["code"] == nil || rec["name"] == nil {
			t.Fatalf("symbol record missing code/name: %#v", rec)
		}
	}
}

// TestFixtureModeNeedsNoNetwork ensures conformance can run credential-free.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := exchangerates.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "start_date": "2026-01-01"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	for _, stream := range []string{"exchange_rates", "latest", "symbols"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read fixture %s: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s emitted no records", stream)
		}
	}
}

func TestRegistryResolvesExchangeRates(t *testing.T) {
	_ = exchangerates.New() // ensure init ran
	c := exchangerates.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("exchange-rates is read-only; Write should be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("exchange-rates"); !ok {
		t.Fatal("registry did not resolve exchange-rates (self-registration)")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := exchangerates.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"exchange_rates": false, "latest": false, "symbols": false}
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
