package paypaltransaction_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	paypaltransaction "polymetrics/internal/connectors/paypal-transaction"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the PayPal
// Transaction connector: it exercises the OAuth2 client-credentials token
// exchange, the Bearer auth applied to data requests, page-increment pagination
// over the transactions report (total_pages), and record mapping. Red until
// internal/connectors/paypal-transaction exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		tokenHits    int
		sawTokenAuth string
		sawTokenForm string
		sawDataAuth  string
		pagesSeen    []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/oauth2/token":
			tokenHits++
			sawTokenAuth = r.Header.Get("Authorization")
			_ = r.ParseForm()
			sawTokenForm = r.Form.Get("grant_type")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"A123","token_type":"Bearer","expires_in":3600}`))
		case "/v1/reporting/transactions":
			sawDataAuth = r.Header.Get("Authorization")
			page := r.URL.Query().Get("page")
			if page == "" {
				page = "1"
			}
			pagesSeen = append(pagesSeen, page)
			switch page {
			case "1":
				_, _ = w.Write([]byte(`{"transaction_details":[{"transaction_info":{"transaction_id":"T1","transaction_amount":{"currency_code":"USD","value":"10.00"},"transaction_status":"S","transaction_initiation_date":"2026-01-01T00:00:00Z"}}],"page":1,"total_pages":2}`))
			case "2":
				_, _ = w.Write([]byte(`{"transaction_details":[{"transaction_info":{"transaction_id":"T2","transaction_amount":{"currency_code":"USD","value":"20.00"},"transaction_status":"S","transaction_initiation_date":"2026-01-02T00:00:00Z"}}],"page":2,"total_pages":2}`))
			default:
				t.Errorf("unexpected page=%q", page)
				_, _ = w.Write([]byte(`{"transaction_details":[],"page":99,"total_pages":2}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := paypaltransaction.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"token_url":  srv.URL + "/v1/oauth2/token",
			"start_date": "2026-01-01T00:00:00Z",
			"end_date":   "2026-01-03T00:00:00Z",
		},
		Secrets: map[string]string{
			"client_id":     "cid",
			"client_secret": "csecret",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "transactions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if tokenHits == 0 {
		t.Fatal("expected OAuth2 token endpoint to be called")
	}
	if !strings.HasPrefix(sawTokenAuth, "Basic ") {
		t.Fatalf("token request Authorization = %q, want Basic ... (client credentials)", sawTokenAuth)
	}
	if sawTokenForm != "client_credentials" {
		t.Fatalf("token grant_type = %q, want client_credentials", sawTokenForm)
	}
	if sawDataAuth != "Bearer A123" {
		t.Fatalf("data request Authorization = %q, want Bearer A123", sawDataAuth)
	}
	if len(pagesSeen) != 2 {
		t.Fatalf("pages requested = %v, want 2 pages", pagesSeen)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (across 2 pages)", len(got))
	}
	if got[0]["transaction_id"] != "T1" || got[1]["transaction_id"] != "T2" {
		t.Fatalf("record mapping wrong: %+v", got)
	}
	if got[0]["currency_code"] != "USD" || got[0]["amount"] != "10.00" {
		t.Fatalf("nested mapping wrong: %+v", got[0])
	}
}

// TestBalancesSinglePage covers a stream with no pagination and a different
// records path, confirming the data-driven routing table.
func TestBalancesSinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"B","token_type":"Bearer","expires_in":3600}`))
		case "/v1/reporting/balances":
			_, _ = w.Write([]byte(`{"balances":[{"currency":"USD","total_balance":{"currency_code":"USD","value":"100.00"}},{"currency":"EUR","total_balance":{"currency_code":"EUR","value":"50.00"}}],"account_id":"ACC1"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := paypaltransaction.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"token_url":  srv.URL + "/v1/oauth2/token",
			"start_date": "2026-01-01T00:00:00Z",
		},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "csecret"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "balances", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read balances: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("balances = %d, want 2", len(got))
	}
	if got[0]["currency"] != "USD" {
		t.Fatalf("balance mapping wrong: %+v", got[0])
	}
}

// TestFixtureMode confirms a credential-free deterministic read so conformance
// passes without live PayPal creds.
func TestFixtureMode(t *testing.T) {
	c := paypaltransaction.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"transactions", "balances", "products", "disputes"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read %s: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture %s produced no records", stream)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogAndMetadata asserts the published catalog and read-only caps.
func TestCatalogAndMetadata(t *testing.T) {
	c := paypaltransaction.New()
	md := c.Metadata()
	if !md.Capabilities.Read || !md.Capabilities.Catalog || !md.Capabilities.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", md.Capabilities)
	}
	if md.Capabilities.Write {
		t.Fatalf("paypal-transaction is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("expected >=3 streams, got %d", len(cat.Streams))
	}
	want := map[string]bool{"transactions": false, "balances": false}
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
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegistryResolution confirms self-registration under the bare system name.
func TestRegistryResolution(t *testing.T) {
	_ = paypaltransaction.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("paypal-transaction"); !ok {
		t.Fatal("registry did not resolve paypal-transaction (self-registration)")
	}
}
