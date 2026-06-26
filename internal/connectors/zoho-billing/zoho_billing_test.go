package zohobilling_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	zohobilling "polymetrics.ai/internal/connectors/zoho-billing"
)

func TestReadSubscriptionsAuthenticatesAndPaginates(t *testing.T) {
	var sawAuth string
	var sawSecondPage bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/subscriptions" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"subscriptions":[{"subscription_id":"sub_1","customer_id":"cust_1","status":"live","updated_time":"2026-01-01T00:00:00Z"},{"subscription_id":"sub_2","customer_id":"cust_2","status":"trial","updated_time":"2026-01-02T00:00:00Z"}]}`))
		case "2":
			sawSecondPage = true
			_, _ = w.Write([]byte(`{"subscriptions":[{"subscription_id":"sub_3","customer_id":"cust_3","status":"cancelled","updated_time":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := zohobilling.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/api", "page_size": "2"}, Secrets: map[string]string{"access_token": "test_access_token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "subscriptions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Zoho-oauthtoken test_access_token" {
		t.Fatalf("Authorization = %q, want Zoho-oauthtoken", sawAuth)
	}
	if !sawSecondPage {
		t.Fatal("expected second page request")
	}
	if len(got) != 3 || got[0]["id"] != "sub_1" || got[2]["status"] != "cancelled" {
		t.Fatalf("records = %+v, want three mapped subscriptions", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := zohobilling.New()
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
	if cat.Connector != "zoho-billing" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want zoho-billing streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("zoho-billing"); !ok {
		t.Fatal("registry did not resolve zoho-billing")
	}
	if c.Metadata().Capabilities.Write {
		t.Fatal("zoho-billing should be read-only")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
