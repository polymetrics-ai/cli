package paystack_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/paystack"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Paystack
// connector: Bearer auth with the secret_key, Paystack meta.next page-number
// pagination over data[], and record mapping across two pages.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/customer" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"status":true,"message":"ok","data":[{"id":1,"customer_code":"CUS_1","email":"a@example.com","createdAt":"2026-01-01T00:00:00.000Z"},{"id":2,"customer_code":"CUS_2","email":"b@example.com","createdAt":"2026-01-02T00:00:00.000Z"}],"meta":{"total":3,"perPage":2,"page":1,"pageCount":2,"next":2,"previous":null}}`))
		case "2":
			_, _ = w.Write([]byte(`{"status":true,"message":"ok","data":[{"id":3,"customer_code":"CUS_3","email":"c@example.com","createdAt":"2026-01-03T00:00:00.000Z"}],"meta":{"total":3,"perPage":2,"page":2,"pageCount":2,"next":null,"previous":1}}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"status":true,"data":[],"meta":{"next":null}}`))
		}
	}))
	defer srv.Close()

	c := paystack.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"secret_key": "sk_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer sk_test_123" {
		t.Fatalf("Authorization = %q, want Bearer sk_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["customer_code"] == nil {
			t.Fatalf("record missing id/customer_code: %+v", rec)
		}
	}
}

// TestReadTransactionsMapping checks a second stream maps its core fields.
func TestReadTransactionsMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transaction" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"status":true,"message":"ok","data":[{"id":99,"reference":"ref_99","amount":50000,"currency":"NGN","status":"success","createdAt":"2026-02-01T00:00:00.000Z"}],"meta":{"next":null}}`))
	}))
	defer srv.Close()

	c := paystack.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"secret_key": "sk_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "transactions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	rec := got[0]
	if rec["reference"] != "ref_99" || rec["status"] != "success" {
		t.Fatalf("unexpected mapping: %+v", rec)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network call (credential-free conformance).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := paystack.New()
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
	// Check fixture mode also short-circuits without creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams verifies the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := paystack.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"customers": false, "transactions": false, "subscriptions": false, "invoices": false, "disputes": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegistryResolves confirms self-registration via the registry.
func TestRegistryResolves(t *testing.T) {
	_ = paystack.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("paystack"); !ok {
		t.Fatal("registry did not resolve paystack (self-registration)")
	}
	c := paystack.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
