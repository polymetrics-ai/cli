package square_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/square"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Square
// connector: Bearer auth + Square-Version header, cursor pagination (query
// "cursor" in, body field "cursor" out) over the "payments" array, and record
// mapping. Red until internal/connectors/square exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth, sawVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawVersion = r.Header.Get("Square-Version")
		if r.URL.Path != "/v2/payments" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"payments":[{"id":"pay_1","created_at":"2026-01-01T00:00:00Z","status":"COMPLETED"},{"id":"pay_2","created_at":"2026-01-02T00:00:00Z","status":"COMPLETED"}],"cursor":"CURSOR2"}`))
		case "CURSOR2":
			_, _ = w.Write([]byte(`{"payments":[{"id":"pay_3","created_at":"2026-01-03T00:00:00Z","status":"COMPLETED"}]}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"payments":[]}`))
		}
	}))
	defer srv.Close()

	c := square.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/v2"},
		Secrets: map[string]string{"credentials.api_key": "EAAA_test_token"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "payments", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer EAAA_test_token" {
		t.Fatalf("Authorization = %q, want Bearer EAAA_test_token", sawAuth)
	}
	if sawVersion == "" {
		t.Fatalf("Square-Version header was not set")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["created_at"] == nil {
			t.Fatalf("record missing id/created_at: %+v", rec)
		}
	}
}

// TestReadCustomersMapsFields confirms a second stream routes to its own
// endpoint and maps the documented fields.
func TestReadCustomersMapsFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/customers" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"customers":[{"id":"cust_1","created_at":"2026-01-01T00:00:00Z","given_name":"Ada","email_address":"ada@example.com"}]}`))
	}))
	defer srv.Close()

	c := square.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/v2"},
		Secrets: map[string]string{"credentials.api_key": "tok"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read customers: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["given_name"] != "Ada" || got[0]["email_address"] != "ada@example.com" {
		t.Fatalf("customer fields not mapped: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// with no network access and no credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := square.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "payments", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil {
		t.Fatalf("fixture record missing id: %+v", got[0])
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestRegisteredReadOnly confirms self-registration via the global registry and
// the read-only capability profile.
func TestRegisteredReadOnly(t *testing.T) {
	_ = square.New() // ensure init ran
	c := square.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("square should be read-only, got Write=true")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("square"); !ok {
		t.Fatal("registry did not resolve square (self-registration)")
	}
}
