package invoiced_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/invoiced"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Invoiced
// connector: HTTP Basic auth (api_key as username, blank password), page-number
// pagination over a root-level JSON array, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/customers" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "1":
			// Full page (per_page=2) signals there is a next page.
			_, _ = w.Write([]byte(`[{"id":101,"name":"Acme","number":"CUST-01"},{"id":102,"name":"Globex","number":"CUST-02"}]`))
		case "2":
			// Short page ends pagination.
			_, _ = w.Write([]byte(`[{"id":103,"name":"Initech","number":"CUST-03"}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := invoiced.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "test_api_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("test_api_key:"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
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

// TestReadFixtureMode confirms the credential-free fixture path emits
// deterministic records without any network access.
func TestReadFixtureMode(t *testing.T) {
	c := invoiced.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "invoices", Config: cfg}, func(rec connectors.Record) error {
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

	// Check should short-circuit in fixture mode with no creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestBaseURLValidation rejects non-http(s) overrides to bound SSRF.
func TestBaseURLValidation(t *testing.T) {
	c := invoiced.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with file:// base_url err = %v, want base_url validation error", err)
	}
}

// TestCatalogAndMetadata checks the published catalog and read-only capabilities.
func TestCatalogAndMetadata(t *testing.T) {
	c := invoiced.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only connector)", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"customers": true, "invoices": true, "payments": true, "subscriptions": true, "estimates": true}
	got := map[string]bool{}
	for _, s := range cat.Streams {
		got[s.Name] = true
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name := range want {
		if !got[name] {
			t.Fatalf("catalog missing stream %q (got %v)", name, got)
		}
	}
}

// TestRegistryResolution confirms self-registration via init() resolves through
// the production registry.
func TestRegistryResolution(t *testing.T) {
	_ = invoiced.New() // ensure the package init() has run
	r := connectors.NewRegistry()
	if _, ok := r.Get("invoiced"); !ok {
		t.Fatal("registry did not resolve invoiced (self-registration)")
	}
}
