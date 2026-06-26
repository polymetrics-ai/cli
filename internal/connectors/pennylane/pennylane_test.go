package pennylane_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/pennylane"
)

func TestReadCustomersPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		paths = append(paths, r.URL.Path+"?"+r.URL.RawQuery)
		if r.URL.Path != "/customers" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"items":[{"id":1,"name":"Acme","updated_at":"2026-01-01T00:00:00Z"}],"has_more":true,"next_cursor":"next-1"}`))
		case "next-1":
			_, _ = w.Write([]byte(`{"items":[{"id":2,"name":"Beta","updated_at":"2026-01-02T00:00:00Z"}],"has_more":false,"next_cursor":null}`))
		default:
			t.Fatalf("unexpected cursor %q", r.URL.Query().Get("cursor"))
		}
	}))
	defer srv.Close()

	c := pennylane.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1"}, Secrets: map[string]string{"api_key": "pl_test"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer pl_test" {
		t.Fatalf("Authorization = %q, want Bearer pl_test", sawAuth)
	}
	if len(got) != 2 || got[0]["id"] == nil || got[1]["name"] != "Beta" {
		t.Fatalf("records = %+v, want two mapped customers", got)
	}
	if len(paths) != 2 {
		t.Fatalf("requested %d pages, want 2: %v", len(paths), paths)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := pennylane.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "pennylane" || len(cat.Streams) < 2 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("pennylane"); !ok {
		t.Fatal("registry did not resolve pennylane")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
