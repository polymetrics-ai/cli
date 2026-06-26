package persistiq_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/persistiq"
)

func TestReadLeadsPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("X-API-KEY")
		pages = append(pages, r.URL.Query().Get("page"))
		if r.URL.Path != "/v1/leads" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`{"leads":[{"id":"l1","email":"one@example.com","status":"active"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"leads":[]}`))
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := persistiq.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1"}, Secrets: map[string]string{"api_key": "piq_key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "leads", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "piq_key" {
		t.Fatalf("X-API-KEY = %q, want piq_key", sawKey)
	}
	if len(got) != 1 || got[0]["email"] != "one@example.com" {
		t.Fatalf("records = %+v, want mapped lead", got)
	}
	if len(pages) != 2 {
		t.Fatalf("pages = %v, want two requests", pages)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := persistiq.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "leads", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "persistiq" || len(cat.Streams) < 3 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("persistiq"); !ok {
		t.Fatal("registry did not resolve persistiq")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
