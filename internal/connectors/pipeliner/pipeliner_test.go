package pipeliner

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadPaginatesAuthenticatesAndMapsRecords(t *testing.T) {
	var sawAuth, sawLimit, sawOffset string
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v100/rest/spaces/test-space/entities/Accounts" {
			http.NotFound(w, r)
			return
		}
		sawLimit = r.URL.Query().Get("limit")
		sawOffset = r.URL.Query().Get("offset")
		switch sawOffset {
		case "0":
			_, _ = w.Write([]byte(`{"data":[{"id":"a1","name":"Acme","modified":"2026-01-01T00:00:00Z"},{"id":"a2","name":"Beta","modified":"2026-01-02T00:00:00Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"a3","name":"Core","modified":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Fatalf("unexpected offset %q", sawOffset)
		}
	}))
	defer srv.Close()

	c := New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/api/v100/rest", "space_id": "test-space", "page_size": "2"},
		Secrets: map[string]string{"username": "unit-user", "password": "unit-pass"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Basic "+base64.StdEncoding.EncodeToString([]byte("unit-user:unit-pass")) {
		t.Fatalf("Authorization header not basic auth, got %q", sawAuth)
	}
	if sawLimit != "2" || requests != 2 {
		t.Fatalf("pagination limit=%q requests=%d, want limit 2 and 2 requests", sawLimit, requests)
	}
	if len(got) != 3 || got[0]["id"] != "a1" || got[0]["updated_at"] != "2026-01-01T00:00:00Z" {
		t.Fatalf("mapped records = %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 || cat.Streams[0].PrimaryKey[0] != "id" {
		t.Fatalf("catalog = %+v", cat)
	}
	if _, ok := connectors.NewRegistry().Get("pipeliner"); !ok {
		t.Fatal("registry did not resolve pipeliner")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
	if strings.TrimSpace(c.Metadata().DisplayName) == "" || c.Metadata().Capabilities.Write {
		t.Fatalf("metadata = %+v", c.Metadata())
	}
}
