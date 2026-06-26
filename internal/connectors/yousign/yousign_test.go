package yousign_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/yousign"
)

func TestReadSignatureRequestsAuthenticatesAndMaps(t *testing.T) {
	var sawAuth, sawLimit string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawLimit = r.URL.Query().Get("limit")
		if r.URL.Path != "/v3/signature_requests" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"sr_1","name":"Contract","status":"done","created_at":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := yousign.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/v3", "limit": "25"}, Secrets: map[string]string{"api_key": "yousign_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "signature_requests", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer yousign_key" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if sawLimit != "25" {
		t.Fatalf("limit = %q, want 25", sawLimit)
	}
	if len(got) != 1 || got[0]["id"] != "sr_1" || got[0]["status"] != "done" {
		t.Fatalf("records = %+v, want mapped signature request", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := yousign.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: fixture}, func(rec connectors.Record) error {
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
	if cat.Connector != "yousign" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want yousign streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("yousign"); !ok {
		t.Fatal("registry did not resolve yousign")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
