package klarna_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/klarna"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Klarna
// connector: HTTP Basic auth, Settlements API offset/size pagination over the
// "payouts" array, and record mapping. Red until internal/connectors/klarna
// exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/settlements/v1/payouts" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			// Full first page (size=2) -> there is a next page.
			_, _ = w.Write([]byte(`{"pagination":{"count":2,"total":3,"offset":0},"payouts":[{"payout_reference":"po_1","currency_code":"EUR","totals":{"settlement_amount":1000}},{"payout_reference":"po_2","currency_code":"EUR","totals":{"settlement_amount":2000}}]}`))
		case "2":
			// Short final page -> stop.
			_, _ = w.Write([]byte(`{"pagination":{"count":1,"total":3,"offset":2},"payouts":[{"payout_reference":"po_3","currency_code":"EUR","totals":{"settlement_amount":3000}}]}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"pagination":{"count":0,"total":3},"payouts":[]}`))
		}
	}))
	defer srv.Close()

	c := klarna.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"password": "sharedsecret", "username": "merchant_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "payouts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("merchant_abc:sharedsecret"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["payout_reference"] == nil {
			t.Fatalf("record missing payout_reference: %+v", rec)
		}
	}
	if got[0]["payout_reference"] != "po_1" || got[2]["payout_reference"] != "po_3" {
		t.Fatalf("unexpected ordering: %+v", got)
	}
}

// TestReadTransactionsPaginates exercises the second core stream and confirms
// the "transactions" JSON path + offset pagination.
func TestReadTransactionsPaginates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/settlements/v1/transactions" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"pagination":{"count":2,"total":3},"transactions":[{"transaction_id":"tx_1","amount":100,"currency_code":"EUR"},{"transaction_id":"tx_2","amount":200,"currency_code":"EUR"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"pagination":{"count":1,"total":3},"transactions":[{"transaction_id":"tx_3","amount":300,"currency_code":"EUR"}]}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	c := klarna.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"password": "secret", "username": "merchant_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "transactions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
}

// TestFixtureMode confirms credential-free deterministic reads for conformance.
func TestFixtureMode(t *testing.T) {
	c := klarna.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"payouts", "transactions", "payout_summary"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) returned no records", stream)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestBaseURLFromRegion confirms region+playground resolution and SSRF guarding
// of any explicit base_url override.
func TestBaseURLFromRegion(t *testing.T) {
	c := klarna.New()
	// Production NA region should require credentials but reach the right host;
	// we cannot make a live call, so only assert Check rejects missing secret.
	cfg := connectors.RuntimeConfig{Config: map[string]string{"region": "na", "playground": "false"}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check should fail without a password secret")
	}

	// An invalid base_url override must be rejected (SSRF guard).
	bad := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"password": "s", "username": "u"},
	}
	if err := c.Check(context.Background(), bad); err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Check should reject non-http base_url, got %v", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := klarna.New()
	md := c.Metadata()
	if !md.Capabilities.Read || md.Capabilities.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", md.Capabilities)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("want >=3 streams, got %d", len(cat.Streams))
	}
}

func TestRegisteredWithRegistry(t *testing.T) {
	_ = klarna.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("klarna"); !ok {
		t.Fatal("registry did not resolve klarna (self-registration)")
	}
}
