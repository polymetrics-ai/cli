package clockodo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/clockodo"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Clockodo
// connector: custom-header auth (X-ClockodoApiUser / X-ClockodoApiKey /
// X-Clockodo-External-Application), page-based pagination across two pages via
// the `paging` object, and record mapping out of the `customers` wrapper key.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawUser, sawKey, sawApp string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawUser = r.Header.Get("X-ClockodoApiUser")
		sawKey = r.Header.Get("X-ClockodoApiKey")
		sawApp = r.Header.Get("X-Clockodo-External-Application")
		if r.URL.Path != "/v2/customers" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{"paging":{"items_per_page":50,"current_page":1,"count_pages":2,"count_items":3},"customers":[{"id":1,"name":"Acme","active":true},{"id":2,"name":"Beta","active":true}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"paging":{"items_per_page":50,"current_page":2,"count_pages":2,"count_items":3},"customers":[{"id":3,"name":"Gamma","active":false}]}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"paging":{"current_page":3,"count_pages":2},"customers":[]}`))
		}
	}))
	defer srv.Close()

	c := clockodo.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":             srv.URL,
			"email_address":        "me@example.com",
			"external_application": "polymetrics;ops@example.com",
		},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawUser != "me@example.com" {
		t.Fatalf("X-ClockodoApiUser = %q, want me@example.com", sawUser)
	}
	if sawKey != "key_abc" {
		t.Fatalf("X-ClockodoApiKey = %q, want key_abc", sawKey)
	}
	if sawApp != "polymetrics;ops@example.com" {
		t.Fatalf("X-Clockodo-External-Application = %q, want polymetrics;ops@example.com", sawApp)
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

// TestFixtureModeNeedsNoNetwork confirms credential-free conformance: fixture
// mode emits deterministic records without any HTTP call.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := clockodo.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"customers", "projects", "services", "users"} {
		var got []connectors.Record
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
	// Check in fixture mode must not require creds or network.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := clockodo.New()
	md := c.Metadata()
	if !md.Capabilities.Read || !md.Capabilities.Catalog || md.Capabilities.Write {
		t.Fatalf("capabilities = %+v, want Read && Catalog && !Write", md.Capabilities)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
}

func TestRegisteredViaRegistry(t *testing.T) {
	_ = clockodo.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("clockodo"); !ok {
		t.Fatal("registry did not resolve clockodo (self-registration)")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := clockodo.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":             "ftp://evil.example.com",
			"email_address":        "me@example.com",
			"external_application": "app;ops@example.com",
		},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should fail SSRF validation")
	}
}
