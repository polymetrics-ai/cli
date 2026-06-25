package woocommerce_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/woocommerce"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the WooCommerce
// connector: HTTP Basic auth (consumer key/secret), WordPress page-number
// pagination over a top-level JSON array bounded by X-WP-TotalPages, and record
// mapping. Red until internal/connectors/woocommerce exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/wp-json/wc/v3/orders" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-WP-TotalPages", "2")
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`[{"id":1,"status":"processing","total":"10.00","date_created":"2026-01-01T00:00:00"},{"id":2,"status":"completed","total":"20.00","date_created":"2026-01-02T00:00:00"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"id":3,"status":"pending","total":"30.00","date_created":"2026-01-03T00:00:00"}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := woocommerce.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "ck_test", "api_secret": "cs_test"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "orders", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("ck_test:cs_test"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["status"] == nil {
			t.Fatalf("record missing id/status: %+v", rec)
		}
	}
}

// TestReadShopBaseURL confirms the base URL is derived from the shop config when
// no explicit base_url override is given, and that per_page drives page size.
func TestReadShopBaseURL(t *testing.T) {
	var sawPerPage string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wp-json/wc/v3/products" {
			http.NotFound(w, r)
			return
		}
		sawPerPage = r.URL.Query().Get("per_page")
		w.Header().Set("X-WP-TotalPages", "1")
		_, _ = w.Write([]byte(`[{"id":10,"name":"Widget","status":"publish"}]`))
	}))
	defer srv.Close()

	c := woocommerce.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "50"},
		Secrets: map[string]string{"api_key": "ck", "api_secret": "cs"},
	}
	var n int
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(connectors.Record) error {
		n++
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if n != 1 {
		t.Fatalf("records = %d, want 1", n)
	}
	if sawPerPage != "50" {
		t.Fatalf("per_page = %q, want 50", sawPerPage)
	}
}

// TestFixtureMode confirms credential-free fixture reads emit deterministic
// records for conformance.
func TestFixtureMode(t *testing.T) {
	c := woocommerce.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
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

// TestCheckFixture confirms Check short-circuits in fixture mode (no network).
func TestCheckFixture(t *testing.T) {
	c := woocommerce.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := woocommerce.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"orders": false, "products": false, "customers": false, "coupons": false}
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
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

func TestReadOnlyCapabilities(t *testing.T) {
	c := woocommerce.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("woocommerce"); !ok {
		t.Fatal("registry did not resolve woocommerce (self-registration)")
	}
}

// TestBadBaseURL ensures SSRF validation rejects a non-http(s) base_url.
func TestBadBaseURL(t *testing.T) {
	c := woocommerce.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil/x"},
		Secrets: map[string]string{"api_key": "ck", "api_secret": "cs"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "orders", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should fail SSRF validation")
	}
}
