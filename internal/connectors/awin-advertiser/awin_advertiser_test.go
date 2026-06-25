package awinadvertiser_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	awinadvertiser "polymetrics/internal/connectors/awin-advertiser"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the Awin
// Bearer auth header, page-number pagination across two pages over a top-level
// JSON array, and record mapping for the transactions stream.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/advertisers/42/transactions/" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// Full page (page size 2) -> connector requests page 2.
			_, _ = w.Write([]byte(`[{"id":1001,"transactionDate":"2026-01-01T00:00:00","commissionAmount":{"amount":12.5,"currency":"GBP"},"transactionStatus":"approved"},{"id":1002,"transactionDate":"2026-01-02T00:00:00","commissionAmount":{"amount":7.0,"currency":"GBP"},"transactionStatus":"pending"}]`))
		case "2":
			// Short page -> stop.
			_, _ = w.Write([]byte(`[{"id":1003,"transactionDate":"2026-01-03T00:00:00","commissionAmount":{"amount":3.0,"currency":"GBP"},"transactionStatus":"approved"}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := awinadvertiser.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":     srv.URL,
			"advertiserId": "42",
			"page_size":    "2",
			"start_date":   "2026-01-01",
		},
		Secrets: map[string]string{"api_key": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "transactions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_test_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["transactionStatus"] == nil {
			t.Fatalf("record missing id/transactionStatus: %+v", rec)
		}
	}
}

// TestReportStreamMapping verifies the campaign_performance report stream maps
// its aggregated rows including the composite primary-key fields.
func TestReportStreamMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/advertisers/42/reports/aggregated/publisher" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"publisherId":7,"publisherName":"Acme","impressions":100,"clicks":10,"totalSaleAmount":{"amount":500.0,"currency":"GBP"},"totalComm":50.0}]`))
	}))
	defer srv.Close()

	c := awinadvertiser.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":     srv.URL,
			"advertiserId": "42",
			"start_date":   "2026-01-01",
		},
		Secrets: map[string]string{"api_key": "tok_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaign_performance", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["publisherId"] == nil || got[0]["clicks"] == nil {
		t.Fatalf("report record missing fields: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records
// without any network call so conformance runs without live credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := awinadvertiser.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	var n int
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "transactions", Config: cfg}, func(rec connectors.Record) error {
		n++
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture) = %v", err)
	}
	if n == 0 {
		t.Fatal("fixture mode emitted no records")
	}
}

// TestRegistryResolves confirms self-registration and metadata capabilities.
func TestRegistryResolves(t *testing.T) {
	_ = awinadvertiser.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("awin-advertiser"); !ok {
		t.Fatal("registry did not resolve awin-advertiser (self-registration)")
	}
	caps := awinadvertiser.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
