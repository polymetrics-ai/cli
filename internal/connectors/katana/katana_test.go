package katana_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/katana"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Katana
// connector: Bearer auth, page-number pagination over the top-level data[]
// array (stopping on a short page), and record mapping. Red until
// internal/connectors/katana exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/products" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		sawPages = append(sawPages, page)
		// limit=2: page 1 is a full page (keep going), page 2 is short (stop).
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{"data":[{"id":1,"name":"Widget","is_sellable":true,"created_at":"2026-01-01T00:00:00.000Z","updated_at":"2026-01-02T00:00:00.000Z"},{"id":2,"name":"Gadget","is_sellable":false,"created_at":"2026-01-03T00:00:00.000Z","updated_at":"2026-01-04T00:00:00.000Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":3,"name":"Gizmo","is_sellable":true,"created_at":"2026-01-05T00:00:00.000Z","updated_at":"2026-01-06T00:00:00.000Z"}]}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := katana.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "kat_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer kat_test_123" {
		t.Fatalf("Authorization = %q, want Bearer kat_test_123", sawAuth)
	}
	if len(sawPages) != 2 {
		t.Fatalf("requested pages = %v, want 2 pages", sawPages)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
	// Spot-check record mapping fidelity.
	if got[0]["name"] != "Widget" || got[2]["name"] != "Gizmo" {
		t.Fatalf("unexpected mapped names: %v / %v", got[0]["name"], got[2]["name"])
	}
}

// TestReadFixtureModeNoNetwork ensures fixture mode emits deterministic records
// with no network access (required for credential-free conformance).
func TestReadFixtureModeNoNetwork(t *testing.T) {
	c := katana.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sales_orders", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits without creds in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := katana.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog has the core streams with
// primary keys and cursor fields.
func TestCatalogStreams(t *testing.T) {
	c := katana.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "katana" {
		t.Fatalf("Catalog connector = %q, want katana", cat.Connector)
	}
	want := map[string]bool{"products": false, "materials": false, "variants": false, "sales_orders": false, "customers": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegistryResolution confirms self-registration via NewRegistry().Get.
func TestRegistryResolution(t *testing.T) {
	_ = katana.New() // ensure package init() ran
	r := connectors.NewRegistry()
	c, ok := r.Get("katana")
	if !ok {
		t.Fatal("registry did not resolve katana (self-registration)")
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, katana is read-only (Write should be false)", caps)
	}
}
