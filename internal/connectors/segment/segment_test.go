package segment_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/segment"
)

func TestContractFixtureAndWrite(t *testing.T) {
	c := segment.New()
	if c.Name() != "segment" {
		t.Fatalf("Name() = %q, want segment", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want read-only Check/Catalog/Read", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) == 0 || cat.Streams[0].Name != "workspaces" {
		t.Fatalf("catalog streams = %+v, want workspaces first", cat.Streams)
	}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "workspaces", Config: cfg}, func(rec connectors.Record) error {
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
	if _, ok := connectors.NewRegistry().Get("segment"); !ok {
		t.Fatal("registry did not resolve segment")
	}
}

func TestReadWorkspacesUsesBearerAndRecordsKey(t *testing.T) {
	var sawAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces" {
			http.NotFound(w, r)
			return
		}
		sawAuth = r.Header.Get("Authorization") == "Bearer test-token"
		_, _ = w.Write([]byte(`{"workspaces":[{"id":"ws_1","name":"Warehouse","slug":"warehouse"}]}`))
	}))
	defer srv.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "max_pages": "1"}, Secrets: map[string]string{"api_token": "test-token"}}
	var got []connectors.Record
	if err := segment.New().Read(context.Background(), connectors.ReadRequest{Stream: "workspaces", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawAuth {
		t.Fatal("bearer auth header was not applied")
	}
	if len(got) != 1 || got[0]["slug"] != "warehouse" {
		t.Fatalf("records = %+v, want workspace slug", got)
	}
}
