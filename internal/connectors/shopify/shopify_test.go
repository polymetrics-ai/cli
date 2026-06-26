package shopify_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/shopify"
)

func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawToken string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.Header.Get("X-Shopify-Access-Token")
		if r.URL.Path != "/admin/api/2026-01/customers.json" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page_info") {
		case "":
			w.Header().Set("Link", fmt.Sprintf(`<%s/admin/api/2026-01/customers.json?page_info=page-2&limit=1>; rel="next"`, srv.URL))
			_, _ = w.Write([]byte(`{"customers":[{"id":1,"email":"one@example.com","created_at":"2026-01-01T00:00:00Z"}]}`))
		case "page-2":
			_, _ = w.Write([]byte(`{"customers":[{"id":2,"email":"two@example.com","created_at":"2026-01-02T00:00:00Z"}]}`))
		default:
			t.Fatalf("unexpected page_info=%q", r.URL.Query().Get("page_info"))
		}
	}))
	defer srv.Close()

	c := shopify.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/admin/api/2026-01", "page_size": "1"}, Secrets: map[string]string{"access_token": "test-access-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "test-access-token" {
		t.Fatalf("X-Shopify-Access-Token = %q, want test token", sawToken)
	}
	if len(got) != 2 || got[0]["id"] == nil || got[0]["email"] != "one@example.com" {
		t.Fatalf("records not paginated/mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := shopify.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "shopify" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v, want shopify streams", cat)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	if _, ok := connectors.NewRegistry().Get("shopify"); !ok {
		t.Fatal("registry did not resolve shopify")
	}
	_, err = c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, []connectors.Record{{"id": "1"}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
