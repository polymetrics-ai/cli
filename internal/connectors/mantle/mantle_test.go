package mantle_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/mantle"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Mantle
// connector: Bearer auth on the Authorization header, Mantle's cursor/hasNextPage
// pagination over the customers[] selector, and record mapping. Red until
// internal/connectors/mantle is implemented.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawTake string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawTake = r.URL.Query().Get("take")
		if r.URL.Path != "/v1/customers" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"customers":[{"id":"cust_1","name":"Acme","updatedAt":"2026-01-01T00:00:00+00:00"},{"id":"cust_2","name":"Globex","updatedAt":"2026-01-02T00:00:00+00:00"}],"cursor":"cust_2","hasNextPage":true}`))
		case "cust_2":
			_, _ = w.Write([]byte(`{"customers":[{"id":"cust_3","name":"Initech","updatedAt":"2026-01-03T00:00:00+00:00"}],"cursor":"cust_3","hasNextPage":false}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"customers":[],"cursor":"","hasNextPage":false}`))
		}
	}))
	defer srv.Close()

	c := mantle.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "mantle_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer mantle_key_123" {
		t.Fatalf("Authorization = %q, want Bearer mantle_key_123", sawAuth)
	}
	if sawTake == "" {
		t.Fatalf("expected take page-size query param to be set")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["id"] != "cust_1" || got[0]["name"] != "Acme" {
		t.Fatalf("record[0] mapping wrong: %+v", got[0])
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["updatedAt"] == nil {
			t.Fatalf("record missing id/updatedAt: %+v", rec)
		}
	}
}

// TestReadSubscriptions exercises the second stream's selector (subscriptions[])
// and its createdAt cursor field mapping.
func TestReadSubscriptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/subscriptions" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"subscriptions":[{"id":"sub_1","active":true,"total":99,"createdAt":"2026-01-01T00:00:00.000Z"}],"cursor":"sub_1","hasNextPage":false}`))
	}))
	defer srv.Close()

	c := mantle.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "k"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "subscriptions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["id"] != "sub_1" || got[0]["createdAt"] == nil {
		t.Fatalf("subscription mapping wrong: %+v", got[0])
	}
}

// TestFixtureModeReadsWithoutNetwork confirms the credential-free conformance
// path emits deterministic records with no HTTP server in sight.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := mantle.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"customers", "subscriptions"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
		if got[0]["id"] == nil {
			t.Fatalf("fixture %s record missing id: %+v", stream, got[0])
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogAndMetadata checks the published catalog and read-only capabilities.
func TestCatalogAndMetadata(t *testing.T) {
	c := mantle.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("mantle is read-only; Write capability should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) != 2 {
		t.Fatalf("streams = %d, want 2", len(cat.Streams))
	}
}

// TestRegistryResolves confirms self-registration via init() resolves through the
// shared registry.
func TestRegistryResolves(t *testing.T) {
	_ = mantle.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("mantle"); !ok {
		t.Fatal("registry did not resolve mantle (self-registration)")
	}
}

// TestBadBaseURLRejected guards the SSRF validation on base_url overrides.
func TestBadBaseURLRejected(t *testing.T) {
	c := mantle.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil/"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected base_url scheme validation error")
	}
}
