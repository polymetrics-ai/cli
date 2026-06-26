package yotpo_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/yotpo"
)

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := yotpo.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "yotpo" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want yotpo streams", cat)
	}
	var rows []connectors.Record
	err = c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: fixture}, func(rec connectors.Record) error {
		rows = append(rows, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(rows) == 0 || rows[0]["fixture"] != true || rows[0]["id"] == nil {
		t.Fatalf("fixture rows = %+v, want fixture records with id", rows)
	}
	if _, ok := connectors.NewRegistry().Get("yotpo"); !ok {
		t.Fatal("registry did not resolve yotpo")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}

func TestReadLiveUsesBearerAuthAndStorePath(t *testing.T) {
	var sawBearer bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/core/v3/stores/store-1/products" {
			http.NotFound(w, r)
			return
		}
		sawBearer = strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ")
		_, _ = w.Write([]byte(`{"products":[{"id":"product-1","name":"Widget","updated_at":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := yotpo.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "store_id": "store-1", "page_size": "1", "max_pages": "1"}, Secrets: map[string]string{"access_token": "test-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawBearer {
		t.Fatal("authorization header was not bearer")
	}
	if len(got) != 1 || got[0]["id"] != "product-1" {
		t.Fatalf("records = %+v, want product-1", got)
	}
}
