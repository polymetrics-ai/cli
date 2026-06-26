package openexchangerates_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"polymetrics.ai/internal/connectors"
	openexchangerates "polymetrics.ai/internal/connectors/open-exchange-rates"
)

// TestReadLatestAuthenticatesAndMaps is the red-first test for the latest
// stream: the app_id must arrive as a query parameter and the rates object must
// be flattened into one record per currency.
func TestReadLatestAuthenticatesAndMaps(t *testing.T) {
	var sawAppID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAppID = r.URL.Query().Get("app_id")
		if r.URL.Path != "/latest.json" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"timestamp":1700000000,"base":"USD","rates":{"EUR":0.92,"GBP":0.79,"JPY":149.5}}`))
	}))
	defer srv.Close()

	c := openexchangerates.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"app_id": "oer_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "latest", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAppID != "oer_test_123" {
		t.Fatalf("app_id query = %q, want oer_test_123", sawAppID)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (one per currency)", len(got))
	}
	for _, rec := range got {
		if rec["currency"] == nil || rec["rate"] == nil {
			t.Fatalf("record missing currency/rate: %+v", rec)
		}
		if rec["base"] != "USD" {
			t.Fatalf("record base = %v, want USD", rec["base"])
		}
	}
}

// TestReadHistoricalPaginatesAcrossDates drives the multi-page path: the
// historical stream walks one request per date from start_date up to today,
// hitting historical/{date}.json each time.
func TestReadHistoricalPaginatesAcrossDates(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/historical/2026-06-23.json":
			_, _ = w.Write([]byte(`{"timestamp":1750636800,"base":"USD","rates":{"EUR":0.90}}`))
		case "/historical/2026-06-24.json":
			_, _ = w.Write([]byte(`{"timestamp":1750723200,"base":"USD","rates":{"EUR":0.91}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := openexchangerates.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"start_date": "2026-06-23",
			"end_date":   "2026-06-24",
		},
		Secrets: map[string]string{"app_id": "oer_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "historical", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	sort.Strings(paths)
	if len(paths) != 2 || paths[0] != "/historical/2026-06-23.json" || paths[1] != "/historical/2026-06-24.json" {
		t.Fatalf("paths = %v, want two historical date requests", paths)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (one per date)", len(got))
	}
	for _, rec := range got {
		if rec["date"] == nil || rec["currency"] != "EUR" || rec["rate"] == nil {
			t.Fatalf("record missing date/currency/rate: %+v", rec)
		}
	}
}

// TestFixtureModeNeedsNoNetwork ensures conformance can run without creds.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := openexchangerates.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "latest", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["currency"] == nil || rec["rate"] == nil {
			t.Fatalf("fixture record missing currency/rate: %+v", rec)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := openexchangerates.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write false (read-only FX API)", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want at least 3", len(cat.Streams))
	}
}

func TestRegisteredWithRegistry(t *testing.T) {
	_ = openexchangerates.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("open-exchange-rates"); !ok {
		t.Fatal("registry did not resolve open-exchange-rates (self-registration)")
	}
}
