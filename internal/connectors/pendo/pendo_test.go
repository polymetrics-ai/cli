package pendo_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/pendo"
)

func TestReadPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawKey string
	var sawPage string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("x-pendo-integration-key")
		if r.URL.Path != "/visitor" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"data":[{"id":"vis_1","email":"a@example.com","lastVisit":"2026-01-01T00:00:00Z"},{"id":"vis_2","email":"b@example.com","lastVisit":"2026-01-02T00:00:00Z"}],"next":"2"}`))
		case "2":
			sawPage = "2"
			_, _ = w.Write([]byte(`{"data":[{"id":"vis_3","email":"c@example.com","lastVisit":"2026-01-03T00:00:00Z"}],"next":""}`))
		default:
			t.Fatalf("unexpected page query %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := pendo.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "limit": "2"}, Secrets: map[string]string{"integration_key": "pendo_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "visitors", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "pendo_key" {
		t.Fatalf("x-pendo-integration-key = %q, want configured key", sawKey)
	}
	if sawPage != "2" {
		t.Fatalf("second page = %q, want 2", sawPage)
	}
	if len(got) != 3 || got[0]["id"] != "vis_1" || got[2]["email"] != "c@example.com" {
		t.Fatalf("mapped records = %+v, want three visitors", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := pendo.New()
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
	if cat.Connector != "pendo" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want pendo streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("pendo"); !ok {
		t.Fatal("registry did not resolve pendo")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
