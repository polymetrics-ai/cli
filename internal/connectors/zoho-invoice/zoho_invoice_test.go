package zohoinvoice_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	zohoinvoice "polymetrics.ai/internal/connectors/zoho-invoice"
)

func TestReadInvoicesAuthenticatesAndAddsOrganization(t *testing.T) {
	var sawAuth string
	var sawOrg string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawOrg = r.URL.Query().Get("organization_id")
		if r.URL.Path != "/invoice/invoices" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"invoices":[{"invoice_id":"inv_1","customer_id":"cust_1","status":"sent","last_modified_time":"2026-01-02T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := zohoinvoice.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/invoice", "organization_id": "org_123"}, Secrets: map[string]string{"access_token": "test_access_token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "invoices", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Zoho-oauthtoken test_access_token" {
		t.Fatalf("Authorization = %q, want Zoho-oauthtoken", sawAuth)
	}
	if sawOrg != "org_123" {
		t.Fatalf("organization_id query = %q, want org_123", sawOrg)
	}
	if len(got) != 1 || got[0]["id"] != "inv_1" || got[0]["status"] != "sent" {
		t.Fatalf("records = %+v, want mapped invoice", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := zohoinvoice.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: fixture}, func(rec connectors.Record) error {
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
	if cat.Connector != "zoho-invoice" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want zoho-invoice streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("zoho-invoice"); !ok {
		t.Fatal("registry did not resolve zoho-invoice")
	}
	if c.Metadata().Capabilities.Write {
		t.Fatal("zoho-invoice should be read-only")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
