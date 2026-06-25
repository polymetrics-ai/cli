package cin7_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/cin7"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Cin7
// connector: dual-header auth (api-auth-accountid + api-auth-applicationkey),
// PageIncrement pagination over page=1,2 with a short-page stop, and record
// mapping out of the Products record-selector path.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAccount, sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAccount = r.Header.Get("api-auth-accountid")
		sawKey = r.Header.Get("api-auth-applicationkey")
		if r.URL.Path != "/externalapi/v2/product" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// A full page (page size 2 in the test) signals more pages follow.
			_, _ = w.Write([]byte(`{"Total":3,"Page":1,"Products":[{"ID":"p1","Name":"Widget","SKU":"W-1"},{"ID":"p2","Name":"Gadget","SKU":"G-1"}]}`))
		case "2":
			// A short page (fewer than page size) ends pagination.
			_, _ = w.Write([]byte(`{"Total":3,"Page":2,"Products":[{"ID":"p3","Name":"Gizmo","SKU":"Z-1"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"Products":[]}`))
		}
	}))
	defer srv.Close()

	c := cin7.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/externalapi/v2", "accountid": "acct-123", "page_size": "2"},
		Secrets: map[string]string{"api_key": "key-abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAccount != "acct-123" {
		t.Fatalf("api-auth-accountid = %q, want acct-123", sawAccount)
	}
	if sawKey != "key-abc" {
		t.Fatalf("api-auth-applicationkey = %q, want key-abc", sawKey)
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
// deterministic records without any network call (used by conformance).
func TestReadFixtureMode(t *testing.T) {
	c := cin7.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check + Catalog must also work credential-free in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
}

// TestRegisteredReadOnly verifies self-registration and the read-only
// capability set.
func TestRegisteredReadOnly(t *testing.T) {
	_ = cin7.New() // ensure init ran
	c := cin7.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only connector)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("cin7"); !ok {
		t.Fatal("registry did not resolve cin7 (self-registration)")
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF check on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := cin7.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com", "accountid": "a"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected")
	}
}
