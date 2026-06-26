package opuswatch_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/opuswatch"
)

func TestReadPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawKey string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("X-API-Key")
		if r.URL.Path != "/monitors" {
			http.NotFound(w, r)
			return
		}
		pages = append(pages, r.URL.Query().Get("page"))
		switch r.URL.Query().Get("page") {
		case "1", "":
			_, _ = w.Write([]byte(`{"data":[{"id":"m1","name":"Checkout","status":"up","updated_at":"2026-01-01T00:00:00Z"}],"next_page":"2"}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"m2","name":"API","status":"down","updated_at":"2026-01-02T00:00:00Z"}]}`))
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := opuswatch.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}, Secrets: map[string]string{"api_key": "op_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "monitors", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "op_key" {
		t.Fatalf("X-API-Key = %q, want op_key", sawKey)
	}
	if len(got) != 2 || got[0]["name"] != "Checkout" || got[1]["status"] != "down" {
		t.Fatalf("records wrong: pages=%v records=%+v", pages, got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := opuswatch.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "incidents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records missing id: %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) < 2 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("opuswatch"); !ok {
		t.Fatal("registry did not resolve opuswatch")
	}
}
