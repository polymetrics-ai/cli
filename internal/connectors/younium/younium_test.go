package younium_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/younium"
)

func TestReadSubscriptionsAuthenticatesAndMaps(t *testing.T) {
	var sawAuth, sawLegalEntity string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawLegalEntity = r.Header.Get("X-Younium-Legal-Entity")
		if r.URL.Path != "/Subscriptions" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"sub_1","name":"Growth","accountId":"acc_1","updated":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := younium.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "username": "api-user", "legal_entity": "LE-1"}, Secrets: map[string]string{"password": "account_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "subscriptions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !strings.HasPrefix(sawAuth, "Basic ") {
		t.Fatalf("Authorization = %q, want basic auth", sawAuth)
	}
	if sawLegalEntity != "LE-1" {
		t.Fatalf("X-Younium-Legal-Entity = %q, want LE-1", sawLegalEntity)
	}
	if len(got) != 1 || got[0]["id"] != "sub_1" || got[0]["account_id"] != "acc_1" {
		t.Fatalf("records = %+v, want mapped subscription", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := younium.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: fixture}, func(rec connectors.Record) error {
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
	if cat.Connector != "younium" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want younium streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("younium"); !ok {
		t.Fatal("registry did not resolve younium")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
