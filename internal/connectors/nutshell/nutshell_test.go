package nutshell_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/nutshell"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Nutshell
// connector: HTTP Basic auth (username + API token), Nutshell page[page]
// pagination over a top-level {"contacts":[...]} envelope, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/contacts" {
			http.NotFound(w, r)
			return
		}
		// Nutshell paginates with page[page] starting at 0; page[limit] is the size.
		switch r.URL.Query().Get("page[page]") {
		case "0":
			_, _ = w.Write([]byte(`{"contacts":[{"id":1,"name":"Ada"},{"id":2,"name":"Grace"}]}`))
		case "1":
			_, _ = w.Write([]byte(`{"contacts":[{"id":3,"name":"Katherine"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"contacts":[]}`))
		default:
			t.Errorf("unexpected page[page]=%q", r.URL.Query().Get("page[page]"))
			_, _ = w.Write([]byte(`{"contacts":[]}`))
		}
	}))
	defer srv.Close()

	c := nutshell.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "user@example.com", "page_size": "2"},
		Secrets: map[string]string{"password": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user@example.com:tok_123"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (pages 0 and 1)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestReadFixtureMode confirms credential-free fixture reads work for conformance.
func TestReadFixtureMode(t *testing.T) {
	c := nutshell.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "leads", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil {
		t.Fatalf("fixture record missing id: %+v", got[0])
	}

	// Check + Catalog must also short-circuit without network in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 5 {
		t.Fatalf("catalog streams = %d, want >= 5", len(cat.Streams))
	}
}

// TestBaseURLValidation rejects non-http(s) base_url overrides (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := nutshell.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "username": "u"},
		Secrets: map[string]string{"password": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}

// TestRegisteredReadOnly verifies self-registration and capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = nutshell.New() // ensure init ran
	caps := nutshell.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("nutshell"); !ok {
		t.Fatal("registry did not resolve nutshell (self-registration)")
	}
}
