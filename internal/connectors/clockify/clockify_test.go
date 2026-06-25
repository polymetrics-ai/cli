package clockify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/clockify"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Clockify
// connector: X-Api-Key auth, page-number pagination over a top-level JSON array,
// the workspace-scoped path, and record mapping. Red until
// internal/connectors/clockify exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	wantPath := "/v1/workspaces/ws_123/clients"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("X-Api-Key")
		if r.URL.Path != wantPath {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		switch page {
		case "1":
			// Full page (pageSize=2) -> connector must request page 2.
			_, _ = w.Write([]byte(`[{"id":"cl_1","name":"Acme","workspaceId":"ws_123"},{"id":"cl_2","name":"Beta","workspaceId":"ws_123"}]`))
		case "2":
			// Short page -> stop.
			_, _ = w.Write([]byte(`[{"id":"cl_3","name":"Gamma","workspaceId":"ws_123"}]`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := clockify.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":     srv.URL,
			"workspace_id": "ws_123",
			"page_size":    "2",
		},
		Secrets: map[string]string{"api_key": "key_test_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "key_test_abc" {
		t.Fatalf("X-Api-Key = %q, want key_test_abc", sawAPIKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestReadWorkspacesUnscoped verifies the workspaces stream hits the top-level
// /v1/workspaces endpoint (not scoped under a workspace id).
func TestReadWorkspacesUnscoped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/workspaces" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("page") == "1" {
			_, _ = w.Write([]byte(`[{"id":"ws_123","name":"My Workspace"}]`))
			return
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := clockify.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "workspace_id": "ws_123", "page_size": "50"},
		Secrets: map[string]string{"api_key": "key_test_abc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "workspaces", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read workspaces: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "ws_123" {
		t.Fatalf("workspaces = %+v, want one record id=ws_123", got)
	}
}

// TestFixtureModeNoNetwork verifies fixture mode yields deterministic records
// without any network or credentials, so conformance passes credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := clockify.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"workspaces", "clients", "projects", "tags", "users"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture Read(%s) = %d records, want 2", stream, len(got))
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
	// Check must also short-circuit in fixture mode (no creds).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams verifies the published catalog has the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := clockify.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "clockify" {
		t.Fatalf("Catalog.Connector = %q, want clockify", cat.Connector)
	}
	want := map[string]bool{"workspaces": false, "clients": false, "projects": false, "tags": false, "users": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegistryResolution verifies self-registration via init() and resolution
// through the shared registry.
func TestRegistryResolution(t *testing.T) {
	_ = clockify.New() // ensure package init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("clockify")
	if !ok {
		t.Fatal("registry did not resolve clockify (self-registration)")
	}
	if got.Name() != "clockify" {
		t.Fatalf("resolved connector Name() = %q, want clockify", got.Name())
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
	if caps.Write {
		t.Fatalf("clockify is read-only; Write capability should be false")
	}
}

// TestMissingAPIKey verifies a non-fixture read without an api_key errors before
// any network call.
func TestMissingAPIKey(t *testing.T) {
	c := clockify.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"workspace_id": "ws_123"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: cfg}, func(connectors.Record) error {
		return nil
	})
	if err == nil {
		t.Fatal("Read without api_key should error")
	}
}
