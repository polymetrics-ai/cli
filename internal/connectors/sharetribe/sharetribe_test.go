package sharetribe_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/sharetribe"
)

func TestContractFixtureAndWrite(t *testing.T) {
	c := sharetribe.New()
	if c.Name() != "sharetribe" {
		t.Fatalf("Name() = %q, want sharetribe", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want read-only Check/Catalog/Read", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) == 0 || cat.Streams[0].Name != "listings" {
		t.Fatalf("catalog streams = %+v, want listings first", cat.Streams)
	}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "listings", Config: cfg}, func(rec connectors.Record) error {
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
	if _, ok := connectors.NewRegistry().Get("sharetribe"); !ok {
		t.Fatal("registry did not resolve sharetribe")
	}
}

func TestReadListingsUsesBearerAndDataKey(t *testing.T) {
	var sawAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/integration_api/listings/query" {
			http.NotFound(w, r)
			return
		}
		sawAuth = r.Header.Get("Authorization") == "Bearer test-token"
		_, _ = w.Write([]byte(`{"data":[{"id":"listing_1","type":"listing","attributes":{"title":"Bike"}}]}`))
	}))
	defer srv.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/v1", "max_pages": "1"}, Secrets: map[string]string{"oauth_access_token": "test-token"}}
	var got []connectors.Record
	if err := sharetribe.New().Read(context.Background(), connectors.ReadRequest{Stream: "listings", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawAuth {
		t.Fatal("bearer auth header was not applied")
	}
	if len(got) != 1 || got[0]["id"] != "listing_1" {
		t.Fatalf("records = %+v, want listing id", got)
	}
}
