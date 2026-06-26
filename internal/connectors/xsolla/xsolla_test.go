package xsolla_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/xsolla"
)

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := xsolla.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "xsolla" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want xsolla streams", cat)
	}
	var rows []connectors.Record
	err = c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: fixture}, func(rec connectors.Record) error {
		rows = append(rows, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(rows) == 0 || rows[0]["fixture"] != true || rows[0]["id"] == nil {
		t.Fatalf("fixture rows = %+v, want fixture records with id", rows)
	}
	if _, ok := connectors.NewRegistry().Get("xsolla"); !ok {
		t.Fatal("registry did not resolve xsolla")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}

func TestReadLiveUsesBasicAuthAndMapsProjects(t *testing.T) {
	var sawBasicAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects" {
			http.NotFound(w, r)
			return
		}
		sawBasicAuth = strings.HasPrefix(r.Header.Get("Authorization"), "Basic ")
		_, _ = w.Write([]byte(`{"items":[{"id":"project-1","name":"Launch","updated_at":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := xsolla.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1", "max_pages": "1"}, Secrets: map[string]string{"merchant_id": "test-merchant", "api_key": "test-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawBasicAuth {
		t.Fatal("authorization header was not HTTP Basic")
	}
	if len(got) != 1 || got[0]["id"] != "project-1" {
		t.Fatalf("records = %+v, want project-1", got)
	}
}
