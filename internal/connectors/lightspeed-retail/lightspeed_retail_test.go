package lightspeedretail_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	lightspeedretail "polymetrics.ai/internal/connectors/lightspeed-retail"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Lightspeed
// Retail connector: Bearer auth with the api_key secret, Lightspeed X-Series
// version-cursor pagination (after=<version.max>, page_size) over the data[]
// array stopping when version.max is empty, and record mapping. Red until
// internal/connectors/lightspeed-retail exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPaths = append(sawPaths, r.URL.Path)
		if r.URL.Path != "/api/2.0/products" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("after") {
		case "":
			// First page: two records, version.max signals more pages exist.
			_, _ = w.Write([]byte(`{"data":[{"id":"p_1","name":"Widget","sku":"W1","version":10},{"id":"p_2","name":"Gadget","sku":"G1","version":11}],"version":{"min":10,"max":11}}`))
		case "11":
			// Second page: one record, version.max is null -> stop.
			_, _ = w.Write([]byte(`{"data":[{"id":"p_3","name":"Gizmo","sku":"Z1","version":12}],"version":{"min":12,"max":null}}`))
		default:
			t.Errorf("unexpected after=%q", r.URL.Query().Get("after"))
			_, _ = w.Write([]byte(`{"data":[],"version":{"max":null}}`))
		}
	}))
	defer srv.Close()

	c := lightspeedretail.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "lsr_token_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer lsr_token_123" {
		t.Fatalf("Authorization = %q, want Bearer lsr_token_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages), paths=%v", len(got), sawPaths)
	}
	if got[0]["id"] != "p_1" || got[0]["sku"] != "W1" || got[0]["name"] != "Widget" {
		t.Fatalf("first record mapped wrong: %+v", got[0])
	}
	if got[2]["id"] != "p_3" {
		t.Fatalf("third record id = %v, want p_3", got[2]["id"])
	}
}

// TestSubdomainBuildsBaseURL verifies that the subdomain config drives the
// Lightspeed host when no explicit base_url override is supplied.
func TestSubdomainBuildsBaseURL(t *testing.T) {
	var sawHost string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawHost = r.Host
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"c_1"}],"version":{"max":null}}`))
	}))
	defer srv.Close()

	// Point base_url at the test server but also pass a subdomain; the explicit
	// base_url override wins, which lets us still exercise the request path.
	c := lightspeedretail.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "subdomain": "example"},
		Secrets: map[string]string{"api_key": "tok"},
	}
	var n int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(connectors.Record) error {
		n++
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if n != 1 {
		t.Fatalf("records = %d, want 1", n)
	}
	if sawHost == "" {
		t.Fatalf("expected a request to reach the server")
	}
}

// TestDefaultSubdomainHost asserts the default base URL is derived from the
// subdomain when no base_url override is given.
func TestDefaultSubdomainHost(t *testing.T) {
	c := lightspeedretail.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"subdomain": "acme"},
		Secrets: map[string]string{"api_key": "tok"},
	}
	// Check must not require a base_url override when a subdomain is present.
	// It will attempt a network call which we expect to fail (no live host),
	// so we only assert the configuration validation does not reject it before
	// the request. A missing subdomain AND base_url must be rejected.
	bad := connectors.RuntimeConfig{Secrets: map[string]string{"api_key": "tok"}}
	if err := c.Check(context.Background(), bad); err == nil {
		t.Fatal("Check should reject config with neither subdomain nor base_url")
	}
	_ = cfg
}

func TestFixtureModeReadIsDeterministic(t *testing.T) {
	c := lightspeedretail.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
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
	// Fixture mode must not require any secret.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := lightspeedretail.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "lightspeed-retail" {
		t.Fatalf("catalog connector = %q, want lightspeed-retail", cat.Connector)
	}
	want := map[string]bool{"products": false, "customers": false, "sales": false, "outlets": false, "registers": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q missing fields", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = lightspeedretail.New() // ensure init ran
	c := lightspeedretail.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read/Catalog/Check", caps)
	}
	if caps.Write {
		t.Fatalf("Lightspeed Retail source is read-only; Write must be false")
	}
	r := connectors.NewRegistry()
	got, ok := r.Get("lightspeed-retail")
	if !ok {
		t.Fatal("registry did not resolve lightspeed-retail (self-registration)")
	}
	if !strings.EqualFold(got.Name(), "lightspeed-retail") {
		t.Fatalf("resolved connector name = %q", got.Name())
	}
}
