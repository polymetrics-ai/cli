package ebayfinance_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	ebayfinance "polymetrics.ai/internal/connectors/ebay-finance"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts Bearer
// auth, eBay Finances limit/offset pagination across two pages of the
// transaction endpoint, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		paths = append(paths, r.URL.Path)
		if r.URL.Path != "/sell/finances/v1/transaction" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{
				"total": 3,
				"limit": 2,
				"offset": 0,
				"transactions": [
					{"transactionId":"t1","transactionType":"SALE","amount":{"value":"10.00","currency":"USD"},"transactionDate":"2026-01-01T00:00:00.000Z"},
					{"transactionId":"t2","transactionType":"REFUND","amount":{"value":"5.00","currency":"USD"},"transactionDate":"2026-01-02T00:00:00.000Z"}
				]
			}`))
		case "2":
			_, _ = w.Write([]byte(`{
				"total": 3,
				"limit": 2,
				"offset": 2,
				"transactions": [
					{"transactionId":"t3","transactionType":"SALE","amount":{"value":"7.00","currency":"USD"},"transactionDate":"2026-01-03T00:00:00.000Z"}
				]
			}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"transactions":[]}`))
		}
	}))
	defer srv.Close()

	c := ebayfinance.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/sell/finances/v1", "page_size": "2"},
		Secrets: map[string]string{"client_access_token": "tok_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "transactions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc123" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if len(paths) != 2 {
		t.Fatalf("requests = %d, want 2 pages", len(paths))
	}
	for _, rec := range got {
		if rec["transactionId"] == nil {
			t.Fatalf("record missing transactionId: %+v", rec)
		}
	}
	if got[0]["transactionId"] != "t1" || got[2]["transactionId"] != "t3" {
		t.Fatalf("unexpected record ordering: %v", got)
	}
	// amount.value should be flattened to a top-level field.
	if got[0]["amount_value"] != "10.00" {
		t.Fatalf("amount_value = %v, want 10.00", got[0]["amount_value"])
	}
}

// TestPayoutsStream confirms a second core stream maps records from its own
// response array field.
func TestPayoutsStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sell/finances/v1/payout" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{
			"total":1,"limit":20,"offset":0,
			"payouts":[{"payoutId":"p1","payoutStatus":"SUCCEEDED","amount":{"value":"99.00","currency":"USD"},"payoutDate":"2026-01-05T00:00:00.000Z"}]
		}`))
	}))
	defer srv.Close()

	c := ebayfinance.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/sell/finances/v1"},
		Secrets: map[string]string{"client_access_token": "tok"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "payouts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read payouts: %v", err)
	}
	if len(got) != 1 || got[0]["payoutId"] != "p1" {
		t.Fatalf("payouts = %v, want one payout p1", got)
	}
	if got[0]["payoutStatus"] != "SUCCEEDED" {
		t.Fatalf("payoutStatus = %v", got[0]["payoutStatus"])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records
// without any network call so conformance runs without creds.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := ebayfinance.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"transactions", "payouts", "transfers", "seller_funds_summary"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}
	// Check should short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := ebayfinance.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"client_access_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "transactions", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme rejection, got %v", err)
	}
}

// TestRegistryResolution confirms self-registration and capabilities.
func TestRegistryResolution(t *testing.T) {
	_ = ebayfinance.New() // ensure init ran
	c := ebayfinance.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("ebay-finance is read-only; Write should be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("ebay-finance"); !ok {
		t.Fatal("registry did not resolve ebay-finance (self-registration)")
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := ebayfinance.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"transactions": false, "payouts": false, "transfers": false, "seller_funds_summary": false}
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
