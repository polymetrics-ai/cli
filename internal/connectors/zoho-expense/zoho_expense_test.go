package zohoexpense_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	zohoexpense "polymetrics.ai/internal/connectors/zoho-expense"
)

func TestReadExpensesAuthenticatesAndExtractsStreamEnvelope(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/expense/expenses" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"expenses":[{"expense_id":"exp_1","merchant_name":"Cafe","amount":42.5,"modified_time":"2026-01-02T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := zohoexpense.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/expense"}, Secrets: map[string]string{"access_token": "test_access_token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "expenses", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Zoho-oauthtoken test_access_token" {
		t.Fatalf("Authorization = %q, want Zoho-oauthtoken", sawAuth)
	}
	if len(got) != 1 || got[0]["id"] != "exp_1" || got[0]["merchant_name"] != "Cafe" {
		t.Fatalf("records = %+v, want mapped expense", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := zohoexpense.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reports", Config: fixture}, func(rec connectors.Record) error {
		rows = append(rows, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(rows) == 0 || rows[0]["id"] == nil {
		t.Fatalf("fixture rows = %+v, want records with id", rows)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "zoho-expense" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want zoho-expense streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("zoho-expense"); !ok {
		t.Fatal("registry did not resolve zoho-expense")
	}
	if c.Metadata().Capabilities.Write {
		t.Fatal("zoho-expense should be read-only")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
