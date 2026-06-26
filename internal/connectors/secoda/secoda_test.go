package secoda_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/secoda"
)

func TestContractFixtureAndWrite(t *testing.T) {
	c := secoda.New()
	if c.Name() != "secoda" {
		t.Fatalf("Name() = %q, want secoda", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want read-only Check/Catalog/Read", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) == 0 || cat.Streams[0].Name != "tables" {
		t.Fatalf("catalog streams = %+v, want tables first", cat.Streams)
	}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tables", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("secoda"); !ok {
		t.Fatal("registry did not resolve secoda")
	}
}

func TestReadTablesUsesBearerAndResults(t *testing.T) {
	var sawAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tables" {
			http.NotFound(w, r)
			return
		}
		sawAuth = r.Header.Get("Authorization") == "Bearer test-token"
		if r.URL.Query().Get("page_size") != "1" {
			t.Fatalf("page_size query was not forwarded")
		}
		_, _ = w.Write([]byte(`{"results":[{"id":"tbl_1","name":"Orders","updated_at":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1", "max_pages": "1"}, Secrets: map[string]string{"api_key": "test-token"}}
	var got []connectors.Record
	if err := secoda.New().Read(context.Background(), connectors.ReadRequest{Stream: "tables", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawAuth {
		t.Fatal("bearer auth header was not applied")
	}
	if len(got) != 1 || got[0]["name"] != "Orders" {
		t.Fatalf("records = %+v, want Orders", got)
	}
}
