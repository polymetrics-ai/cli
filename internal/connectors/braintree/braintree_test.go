package braintree_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/braintree"
)

func TestReadTransactionsPaginatesAuthenticatesAndMapsRecords(t *testing.T) {
	var sawUser string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "pub_key" || pass != "priv_key" {
			t.Fatalf("unexpected basic auth user=%q ok=%v", user, ok)
		}
		sawUser = user
		if r.URL.Path != "/merchants/merchant_1/transactions" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		pages = append(pages, page)
		switch page {
		case "", "1":
			_, _ = w.Write([]byte(`{"transactions":[{"id":"txn_1","amount":"10.00","status":"settled"},{"id":"txn_2","amount":"11.00","status":"submitted"}],"pagination":{"next_page":"2"}}`))
		case "2":
			_, _ = w.Write([]byte(`{"transactions":[{"id":"txn_3","amount":"12.00","status":"settled"}],"pagination":{}}`))
		default:
			t.Fatalf("unexpected page %q", page)
		}
	}))
	defer srv.Close()

	c := braintree.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "merchant_id": "merchant_1", "public_key": "pub_key", "page_size": "2"}, Secrets: map[string]string{"private_key": "priv_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "transactions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawUser != "pub_key" || len(pages) != 2 {
		t.Fatalf("auth/pages wrong user=%q pages=%v", sawUser, pages)
	}
	if len(got) != 3 || got[0]["id"] != "txn_1" || got[0]["amount"] != "10.00" {
		t.Fatalf("records mapped wrong: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := braintree.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"transactions", "customers", "subscriptions"} {
		count := 0
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(connectors.Record) error { count++; return nil }); err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if count == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "braintree" || len(cat.Streams) != 3 {
		t.Fatalf("Catalog = %+v err=%v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("braintree"); !ok {
		t.Fatal("registry did not resolve braintree")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
