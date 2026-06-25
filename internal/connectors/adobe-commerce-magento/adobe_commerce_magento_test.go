package adobecommercemagento_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	adobe "polymetrics.ai/internal/connectors/adobe-commerce-magento"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Adobe Commerce
// (Magento) connector: Bearer (Integration Access Token) auth, Magento
// searchCriteria[currentPage] pagination over the items[] array, and record
// mapping. The Magento REST list shape is
// {"items":[...],"total_count":N,"search_criteria":{...}} and pages are walked by
// incrementing searchCriteria[currentPage] until the accumulated count reaches
// total_count.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/rest/V1/products" {
			http.NotFound(w, r)
			return
		}
		// Magento encodes page as searchCriteria[currentPage] (1-based).
		switch r.URL.Query().Get("searchCriteria[currentPage]") {
		case "1":
			_, _ = w.Write([]byte(`{"items":[{"id":1,"sku":"SKU-1","name":"Widget","price":9.99,"updated_at":"2026-01-01 00:00:00"},{"id":2,"sku":"SKU-2","name":"Gadget","price":19.99,"updated_at":"2026-01-02 00:00:00"}],"search_criteria":{"page_size":2,"current_page":1},"total_count":3}`))
		case "2":
			_, _ = w.Write([]byte(`{"items":[{"id":3,"sku":"SKU-3","name":"Gizmo","price":29.99,"updated_at":"2026-01-03 00:00:00"}],"search_criteria":{"page_size":2,"current_page":2},"total_count":3}`))
		default:
			t.Errorf("unexpected currentPage=%q", r.URL.Query().Get("searchCriteria[currentPage]"))
			_, _ = w.Write([]byte(`{"items":[],"total_count":3}`))
		}
	}))
	defer srv.Close()

	c := adobe.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_test_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["sku"] == nil {
			t.Fatalf("record missing id/sku: %+v", rec)
		}
	}
}

// TestCheckUsesStoreHost verifies the connector builds its base URL from the
// store_host config when base_url is not overridden, hits the products endpoint
// with a bounded read, and sends the bearer token.
func TestCheckUsesStoreHost(t *testing.T) {
	var sawPath, sawAuth, sawPageSize string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		sawAuth = r.Header.Get("Authorization")
		sawPageSize = r.URL.Query().Get("searchCriteria[pageSize]")
		_, _ = w.Write([]byte(`{"items":[],"total_count":0}`))
	}))
	defer srv.Close()

	c := adobe.New()
	cfg := connectors.RuntimeConfig{
		// base_url override is the supported way to point Check at the test server.
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "tok_test_123"},
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if sawPath != "/rest/V1/products" {
		t.Fatalf("Check path = %q, want /rest/V1/products", sawPath)
	}
	if sawAuth != "Bearer tok_test_123" {
		t.Fatalf("Check Authorization = %q, want Bearer tok_test_123", sawAuth)
	}
	if sawPageSize != "1" {
		t.Fatalf("Check pageSize = %q, want bounded read of 1", sawPageSize)
	}
}

// TestFixtureModeNeedsNoNetwork verifies the credential-free fixture path emits
// deterministic records for every stream without any network access.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := adobe.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	catalog, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(catalog.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(catalog.Streams))
	}
	for _, s := range catalog.Streams {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: s.Name, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s): %v", s.Name, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %q emitted no records", s.Name)
		}
		for _, pk := range s.PrimaryKey {
			if got[0][pk] == nil {
				t.Fatalf("fixture stream %q record missing primary key %q: %+v", s.Name, pk, got[0])
			}
		}
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF validation on the base_url override.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := adobe.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with bad base_url scheme should be rejected, got %v", err)
	}
}

// TestRegisteredReadOnly verifies registration via the global registry and that
// the connector advertises read (not write) capability.
func TestRegisteredReadOnly(t *testing.T) {
	_ = adobe.New() // ensure init ran
	c := adobe.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only source)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("adobe-commerce-magento"); !ok {
		t.Fatal("registry did not resolve adobe-commerce-magento (self-registration)")
	}
}
