package zapsign_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/zapsign"
)

func TestReadDocumentsAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v1/docs/" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"results":[{"token":"doc_1","name":"Agreement","status":"signed","created_at":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := zapsign.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/api/v1"}, Secrets: map[string]string{"api_token": "zapsign_token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "documents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Token zapsign_token" {
		t.Fatalf("Authorization = %q, want ZapSign token auth", sawAuth)
	}
	if len(got) != 1 || got[0]["id"] != "doc_1" || got[0]["status"] != "signed" {
		t.Fatalf("records = %+v, want mapped document", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := zapsign.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "signers", Config: fixture}, func(rec connectors.Record) error {
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
	if cat.Connector != "zapsign" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want zapsign streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("zapsign"); !ok {
		t.Fatal("registry did not resolve zapsign")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
