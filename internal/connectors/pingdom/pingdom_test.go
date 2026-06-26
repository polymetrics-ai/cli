package pingdom_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/pingdom"
)

func TestReadChecksPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var offsets []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		offsets = append(offsets, r.URL.Query().Get("offset"))
		if r.URL.Path != "/checks" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "0":
			_, _ = w.Write([]byte(`{"checks":[{"id":100,"name":"Homepage","status":"up","hostname":"example.com"}]}`))
		case "1":
			_, _ = w.Write([]byte(`{"checks":[]}`))
		default:
			t.Fatalf("unexpected offset %q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	c := pingdom.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1"}, Secrets: map[string]string{"api_key": "ping_key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "checks", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer ping_key" {
		t.Fatalf("Authorization = %q, want Bearer ping_key", sawAuth)
	}
	if len(got) != 1 || got[0]["id"] == nil || got[0]["name"] != "Homepage" {
		t.Fatalf("records = %+v, want mapped check", got)
	}
	if len(offsets) != 2 {
		t.Fatalf("offsets = %v, want two requests", offsets)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := pingdom.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "checks", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "pingdom" || len(cat.Streams) < 2 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("pingdom"); !ok {
		t.Fatal("registry did not resolve pingdom")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
