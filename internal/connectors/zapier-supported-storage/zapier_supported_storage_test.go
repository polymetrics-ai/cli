package zapiersupportedstorage_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	zapierstorage "polymetrics.ai/internal/connectors/zapier-supported-storage"
)

func TestReadRecordsAuthenticatesAndMaps(t *testing.T) {
	var sawSecret string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawSecret = r.URL.Query().Get("secret")
		if r.URL.Path != "/api/records" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"records":[{"id":"key_1","key":"key_1","value":"stored value","updated_at":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := zapierstorage.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}, Secrets: map[string]string{"secret": "zapier_secret"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "records", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawSecret != "zapier_secret" {
		t.Fatalf("secret query = %q, want configured secret", sawSecret)
	}
	if len(got) != 1 || got[0]["id"] != "key_1" || got[0]["value"] != "stored value" {
		t.Fatalf("records = %+v, want mapped storage record", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := zapierstorage.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "records", Config: fixture}, func(rec connectors.Record) error {
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
	if cat.Connector != "zapier-supported-storage" || len(cat.Streams) != 1 {
		t.Fatalf("catalog = %+v, want storage records stream", cat)
	}
	if _, ok := connectors.NewRegistry().Get("zapier-supported-storage"); !ok {
		t.Fatal("registry did not resolve zapier-supported-storage")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
