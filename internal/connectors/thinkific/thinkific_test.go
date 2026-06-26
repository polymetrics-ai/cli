package thinkific_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/thinkific"
)

func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	var sawSubdomain string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("X-Auth-API-Key")
		sawSubdomain = r.Header.Get("X-Auth-Subdomain")
		if r.URL.Path != "/api/public/v1/courses" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"items":[{"id":1,"name":"Foundations","slug":"foundations","created_at":"2026-01-01T00:00:00Z"},{"id":2,"name":"Advanced","slug":"advanced","created_at":"2026-01-02T00:00:00Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"items":[{"id":3,"name":"Teams","slug":"teams","created_at":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := thinkific.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "subdomain": "academy", "page_size": "2"}, Secrets: map[string]string{"api_key": "test_api_key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "courses", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "test_api_key" {
		t.Fatalf("X-Auth-API-Key = %q", sawKey)
	}
	if sawSubdomain != "academy" {
		t.Fatalf("X-Auth-Subdomain = %q", sawSubdomain)
	}
	if len(got) != 3 || got[0]["id"] == nil || got[2]["name"] != "Teams" {
		t.Fatalf("records not paginated/mapped: %+v", got)
	}
}

func TestFixtureModeNoCredentials(t *testing.T) {
	c := thinkific.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var count int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "courses", Config: cfg}, func(rec connectors.Record) error {
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
	c := thinkific.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "thinkific" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v", cat)
	}
	if caps := c.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v", caps)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
	if _, ok := connectors.NewRegistry().Get("thinkific"); !ok {
		t.Fatal("registry did not resolve thinkific")
	}
}
