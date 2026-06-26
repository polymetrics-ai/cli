package inflowinventory_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/inflowinventory"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the inFlow
// Inventory connector: raw Authorization API-key header (no Bearer prefix), the
// versioned Accept header, companyid embedded in the path, cursor pagination via
// count/after over a top-level JSON array, and record mapping. Red until
// internal/connectors/inflowinventory exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth, sawAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawAccept = r.Header.Get("Accept")
		if r.URL.Path != "/acme-co/products" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("after") {
		case "":
			_, _ = w.Write([]byte(`[{"productId":"p1","name":"Widget","sku":"W-1"},{"productId":"p2","name":"Gadget","sku":"G-1"}]`))
		case "p2":
			_, _ = w.Write([]byte(`[{"productId":"p3","name":"Gizmo","sku":"Z-1"}]`))
		default:
			t.Errorf("unexpected after=%q", r.URL.Query().Get("after"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := inflowinventory.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "companyid": "acme-co", "page_size": "2"},
		Secrets: map[string]string{"api_key": "ik_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "ik_test_123" {
		t.Fatalf("Authorization = %q, want raw api key ik_test_123 (no Bearer prefix)", sawAuth)
	}
	if !strings.Contains(sawAccept, "version=2024-03-12") {
		t.Fatalf("Accept = %q, want versioned media type", sawAccept)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["productId"] != "p1" || got[0]["name"] != "Widget" || got[0]["sku"] != "W-1" {
		t.Fatalf("first record mapped wrong: %+v", got[0])
	}
	if got[2]["productId"] != "p3" {
		t.Fatalf("last record mapped wrong: %+v", got[2])
	}
}

// TestFixtureModeNeedsNoNetwork confirms the conformance fixture path emits
// deterministic records with no credentials and no network access.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := inflowinventory.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	for _, stream := range []string{"products", "customers", "vendors", "sales_orders", "categories"} {
		got = got[:0]
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

// TestCheckFixtureNoCreds ensures Check short-circuits in fixture mode.
func TestCheckFixtureNoCreds(t *testing.T) {
	c := inflowinventory.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check = %v, want nil", err)
	}
}

// TestCatalogStreams asserts the published catalog covers the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := inflowinventory.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"products": false, "customers": false, "vendors": false, "sales_orders": false, "categories": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
			if len(s.PrimaryKey) == 0 {
				t.Fatalf("stream %s has no primary key", s.Name)
			}
			if len(s.Fields) == 0 {
				t.Fatalf("stream %s has no fields", s.Name)
			}
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestRegisteredReadOnly verifies self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = inflowinventory.New() // ensure init ran
	c := inflowinventory.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("inflowinventory"); !ok {
		t.Fatal("registry did not resolve inflowinventory (self-registration)")
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs to bound SSRF.
func TestBaseURLValidation(t *testing.T) {
	c := inflowinventory.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "companyid": "acme-co"},
		Secrets: map[string]string{"api_key": "ik_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}
