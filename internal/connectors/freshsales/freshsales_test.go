package freshsales_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/freshsales"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Freshsales
// connector: Token auth header, Freshsales page-number pagination across the
// meta.total_pages count over the contacts[] array, and record mapping. Red
// until internal/connectors/freshsales exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/crm/sales/api/contacts/view/0" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"contacts":[{"id":1,"first_name":"Ada","last_name":"Lovelace","email":"ada@example.com","updated_at":"2026-01-01T00:00:00Z"},{"id":2,"first_name":"Grace","last_name":"Hopper","email":"grace@example.com","updated_at":"2026-01-02T00:00:00Z"}],"meta":{"total_pages":2,"total":3}}`))
		case "2":
			_, _ = w.Write([]byte(`{"contacts":[{"id":3,"first_name":"Katherine","last_name":"Johnson","email":"kj@example.com","updated_at":"2026-01-03T00:00:00Z"}],"meta":{"total_pages":2,"total":3}}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"contacts":[],"meta":{"total_pages":2,"total":3}}`))
		}
	}))
	defer srv.Close()

	c := freshsales.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Token token=key_test_123" {
		t.Fatalf("Authorization = %q, want Token token=key_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
	// Record mapping: first contact should map first_name through.
	if got[0]["first_name"] != "Ada" {
		t.Fatalf("first record first_name = %v, want Ada", got[0]["first_name"])
	}
}

// TestFixtureModeReadsWithoutNetwork ensures credential-free conformance works:
// mode=fixture emits deterministic records with no HTTP server.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := freshsales.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	for _, stream := range []string{"contacts", "sales_accounts", "deals", "leads"} {
		got = got[:0]
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture Read(%s) records = %d, want 2", stream, len(got))
		}
		if got[0]["id"] == nil {
			t.Fatalf("fixture Read(%s) record missing id: %+v", stream, got[0])
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := freshsales.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestRegisteredReadOnly asserts self-registration and the read-only capability
// shape, and that the catalog publishes streams.
func TestRegisteredReadOnly(t *testing.T) {
	c := freshsales.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write false", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("freshsales"); !ok {
		t.Fatal("registry did not resolve freshsales (self-registration)")
	}

	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
}
