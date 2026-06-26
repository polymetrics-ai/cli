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
// connector: HTTP Basic auth, Klarna pagination.next cursor pagination over the
// payouts[] array, and record mapping. Red until internal/connectors/klarna
// exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var page2Path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/settlements/v1/payouts" {
			http.NotFound(w, r)
			return
		}
		// Page 1 has a pagination.next pointing back at this endpoint with an
		// offset; page 2 has no next, ending the loop.
		if r.URL.Query().Get("offset") == "500" {
			page2Path = r.URL.Path
			_, _ = w.Write([]byte(`{"payouts":[{"payment_reference":"P3","payout_date":"2026-01-03T00:00:00Z","currency_code":"EUR"}],"pagination":{}}`))
			return
		}
		_, _ = w.Write([]byte(`{"payouts":[` +
			`{"payment_reference":"P1","payout_date":"2026-01-01T00:00:00Z","currency_code":"EUR"},` +
			`{"payment_reference":"P2","payout_date":"2026-01-02T00:00:00Z","currency_code":"USD"}` +
			`],"pagination":{"next":"/settlements/v1/payouts?offset=500&size=500"}}`))
	}))
	defer srv.Close()

	c := klarna.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"password": "secret_pw"},
	}
	// username is config (not a secret) per the Klarna spec.
	cfg.Config["username"] = "merchant_user"

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "payouts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("merchant_user:secret_pw"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if page2Path != "/settlements/v1/payouts" {
		t.Fatalf("page 2 was not requested via pagination.next (path=%q)", page2Path)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["payment_reference"] == nil || rec["payout_date"] == nil {
			t.Fatalf("record missing primary key fields: %+v", rec)
		}
	}
	if got[0]["currency_code"] != "EUR" {
		t.Fatalf("record mapping wrong: %+v", got[0])
	}
}

// TestReadTransactions confirms the second core stream reads its own
// field_path (transactions[]) and maps records.
func TestReadTransactions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/settlements/v1/transactions" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"transactions":[{"capture_id":"cap_1","type":"SALE","amount":2000,"currency_code":"EUR","payment_reference":"P1"}],"pagination":{}}`))
	}))
	defer srv.Close()

	c := klarna.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "u"},
		Secrets: map[string]string{"password": "p"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "transactions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read transactions: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["type"] != "SALE" || got[0]["capture_id"] != "cap_1" {
		t.Fatalf("transaction mapping wrong: %+v", got[0])
	}
}

// TestRegionBaseURL checks the region/playground -> base URL derivation matches
// Klarna's documented URL scheme.
func TestRegionBaseURL(t *testing.T) {
	cases := []struct {
		region     string
		playground string
		want       string
	}{
		{"eu", "false", "https://api.klarna.com"},
		{"na", "false", "https://api-na.klarna.com"},
		{"oc", "false", "https://api-oc.klarna.com"},
		{"eu", "true", "https://api.playground.klarna.com"},
		{"na", "true", "https://api-na.playground.klarna.com"},
		{"", "false", "https://api.klarna.com"},
	}
	for _, tc := range cases {
		got, err := klarna.ResolveBaseURL(connectors.RuntimeConfig{Config: map[string]string{
			"region":     tc.region,
			"playground": tc.playground,
		}})
		if err != nil {
			t.Fatalf("ResolveBaseURL(region=%q playground=%q): %v", tc.region, tc.playground, err)
		}
		if got != tc.want {
			t.Fatalf("ResolveBaseURL(region=%q playground=%q) = %q, want %q", tc.region, tc.playground, got, tc.want)
		}
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := klarna.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "payouts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["payment_reference"] == nil {
		t.Fatalf("fixture record missing payment_reference: %+v", got[0])
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := klarna.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com", "username": "u"},
		Secrets: map[string]string{"password": "p"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "payouts", Config: cfg}, func(rec connectors.Record) error {
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme rejection, got %v", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = klarna.New() // ensure init ran
	caps := klarna.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("klarna is read-only; Write should be false, got %+v", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("klarna"); !ok {
		t.Fatal("registry did not resolve klarna (self-registration)")
	}
}
