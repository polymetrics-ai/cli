package zendesksunshine_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	zendesksunshine "polymetrics.ai/internal/connectors/zendesk-sunshine"
)

func TestReadObjectTypesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/sunshine/objects/types" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"key":"asset","schema":{"properties":{"name":{"type":"string"}}},"created_at":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := zendesksunshine.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/api/sunshine", "email": "agent@example.com", "subdomain": "fixture"}, Secrets: map[string]string{"api_token": "zendesk_token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "object_types", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !strings.HasPrefix(sawAuth, "Basic ") {
		t.Fatalf("Authorization = %q, want Zendesk basic token auth", sawAuth)
	}
	if len(got) != 1 || got[0]["id"] != "asset" {
		t.Fatalf("records = %+v, want mapped object type", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := zendesksunshine.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "objects", Config: fixture}, func(rec connectors.Record) error {
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
	if cat.Connector != "zendesk-sunshine" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want sunshine streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("zendesk-sunshine"); !ok {
		t.Fatal("registry did not resolve zendesk-sunshine")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
