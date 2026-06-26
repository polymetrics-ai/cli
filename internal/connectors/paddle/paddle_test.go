package paddle_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/paddle"
)

func TestReadPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var sawAfter string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/transactions" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("after") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"txn_1","status":"completed","created_at":"2026-01-01T00:00:00Z"},{"id":"txn_2","status":"paid","created_at":"2026-01-02T00:00:00Z"}],"meta":{"pagination":{"next":"txn_2"}}}`))
		case "txn_2":
			sawAfter = "txn_2"
			_, _ = w.Write([]byte(`{"data":[{"id":"txn_3","status":"refunded","created_at":"2026-01-03T00:00:00Z"}],"meta":{"pagination":{"next":null}}}`))
		default:
			t.Fatalf("unexpected after query %q", r.URL.Query().Get("after"))
		}
	}))
	defer srv.Close()

	c := paddle.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "2"}, Secrets: map[string]string{"api_key": "paddle_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "transactions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer paddle_key" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if sawAfter != "txn_2" {
		t.Fatalf("second page cursor = %q, want txn_2", sawAfter)
	}
	if len(got) != 3 || got[0]["id"] != "txn_1" || got[2]["status"] != "refunded" {
		t.Fatalf("mapped records = %+v, want three transactions", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := paddle.New()
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
	if cat.Connector != "paddle" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want paddle streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("paddle"); !ok {
		t.Fatal("registry did not resolve paddle")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
