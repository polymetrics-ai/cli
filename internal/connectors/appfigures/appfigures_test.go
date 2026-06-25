package appfigures_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/appfigures"
)

// TestReadReviewsPaginatesAndAuthenticates is the red-first test: Bearer auth on
// the Authorization header, Appfigures page-number pagination over the reviews[]
// array across two pages, and record mapping. Red until the appfigures package
// exists.
func TestReadReviewsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/reviews" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// total=3, pages=2 -> drives a second page request.
			_, _ = w.Write([]byte(`{"total":3,"pages":2,"this_page":1,"reviews":[{"id":"r1","stars":5,"product":111},{"id":"r2","stars":4,"product":111}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"total":3,"pages":2,"this_page":2,"reviews":[{"id":"r3","stars":3,"product":111}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"total":3,"pages":2,"this_page":3,"reviews":[]}`))
		}
	}))
	defer srv.Close()

	c := appfigures.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "pat_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reviews", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer pat_test_123" {
		t.Fatalf("Authorization = %q, want Bearer pat_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("review record missing id: %+v", rec)
		}
	}
}

// TestReadProductsFlattensKeyedObject verifies that the products stream, whose
// upstream payload is a JSON object keyed by product id, is flattened into one
// record per product (not a single root record).
func TestReadProductsFlattensKeyedObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/products/mine" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"111":{"id":111,"name":"App One","store":"apple"},"222":{"id":222,"name":"App Two","store":"google_play"}}`))
	}))
	defer srv.Close()

	c := appfigures.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "pat_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 products", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("product record missing id/name: %+v", rec)
		}
	}
}

// TestFixtureModeNeedsNoNetwork confirms the credential-free fixture path emits
// deterministic records so conformance can run without live creds or a server.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := appfigures.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reviews", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Fixture Check must also succeed without a network call.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := appfigures.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("appfigures is read-only; Write should be false, got %+v", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
}

func TestRegistryResolves(t *testing.T) {
	_ = appfigures.New() // ensure init() ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("appfigures"); !ok {
		t.Fatal("registry did not resolve appfigures (self-registration)")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := appfigures.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "pat_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reviews", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http(s) base_url scheme")
	}
}
