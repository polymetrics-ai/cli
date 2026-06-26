package theguardianapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	theguardianapi "polymetrics.ai/internal/connectors/the-guardian-api"
)

func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	var sawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("api-key")
		sawQuery = r.URL.Query().Get("q")
		if r.URL.Path != "/search" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"response":{"status":"ok","results":[{"id":"world/1","webTitle":"First","webPublicationDate":"2026-01-01T00:00:00Z"},{"id":"world/2","webTitle":"Second","webPublicationDate":"2026-01-02T00:00:00Z"}]}}`))
		case "2":
			_, _ = w.Write([]byte(`{"response":{"status":"ok","results":[{"id":"world/3","webTitle":"Third","webPublicationDate":"2026-01-03T00:00:00Z"}]}}`))
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := theguardianapi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "2", "query": "climate"}, Secrets: map[string]string{"api_key": "test_api_key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "search", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "test_api_key" {
		t.Fatalf("api-key = %q", sawKey)
	}
	if sawQuery != "climate" {
		t.Fatalf("q = %q", sawQuery)
	}
	if len(got) != 3 || got[0]["id"] != "world/1" || got[2]["title"] != "Third" {
		t.Fatalf("records not paginated/mapped: %+v", got)
	}
}

func TestFixtureModeNoCredentials(t *testing.T) {
	c := theguardianapi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var count int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "search", Config: cfg}, func(rec connectors.Record) error {
		count++
		if rec["id"] == nil {
			t.Fatalf("fixture missing id: %+v", rec)
		}
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if count == 0 {
		t.Fatal("fixture emitted no records")
	}
}

func TestCatalogRegistrationAndReadOnly(t *testing.T) {
	c := theguardianapi.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "the-guardian-api" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v", cat)
	}
	if caps := c.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v", caps)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
	if _, ok := connectors.NewRegistry().Get("the-guardian-api"); !ok {
		t.Fatal("registry did not resolve the-guardian-api")
	}
}
