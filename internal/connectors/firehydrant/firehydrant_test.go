package firehydrant_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/firehydrant"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the FireHydrant
// connector: Bearer auth, FireHydrant {data:[...], pagination:{next}} page-number
// pagination over data[], and record mapping. Red until
// internal/connectors/firehydrant exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/incidents" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"data":[{"id":"inc_1","name":"DB down","created_at":"2026-01-01T00:00:00Z"},{"id":"inc_2","name":"API slow","created_at":"2026-01-02T00:00:00Z"}],"pagination":{"page":1,"next":2,"prev":null}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"inc_3","name":"Cache miss","created_at":"2026-01-03T00:00:00Z"}],"pagination":{"page":2,"next":null,"prev":1}}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"data":[],"pagination":{"next":null}}`))
		}
	}))
	defer srv.Close()

	c := firehydrant.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "fhb_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "incidents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer fhb_test_123" {
		t.Fatalf("Authorization = %q, want Bearer fhb_test_123", sawAuth)
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

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access so conformance runs credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := firehydrant.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "services", Config: cfg}, func(rec connectors.Record) error {
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
	// Check must also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams asserts the published streams cover the core set and each
// stream has a primary key.
func TestCatalogStreams(t *testing.T) {
	c := firehydrant.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"incidents": false, "services": false, "teams": false, "environments": false, "functionalities": false}
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

// TestRegisteredReadOnly asserts self-registration via the registry and the
// read-only capability set.
func TestRegisteredReadOnly(t *testing.T) {
	_ = firehydrant.New() // ensure init ran
	c := firehydrant.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write false (read-only connector)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("firehydrant"); !ok {
		t.Fatal("registry did not resolve firehydrant (self-registration)")
	}
}

// TestBaseURLValidation rejects an invalid base_url scheme (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := firehydrant.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_token": "fhb_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "incidents", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected")
	}
}
