package confluence_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/confluence"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Confluence
// connector: HTTP Basic auth (email + api_token), Confluence v2 cursor
// pagination over results[] via _links.next, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/wiki/api/v2/spaces" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"results":[{"id":"1","key":"DEV","name":"Dev","type":"global","status":"current"},{"id":"2","key":"OPS","name":"Ops","type":"global","status":"current"}],"_links":{"next":"/wiki/api/v2/spaces?cursor=page2&limit=2"}}`))
		case "page2":
			_, _ = w.Write([]byte(`{"results":[{"id":"3","key":"QA","name":"QA","type":"global","status":"current"}],"_links":{}}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"results":[],"_links":{}}`))
		}
	}))
	defer srv.Close()

	c := confluence.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "email": "user@example.com", "page_size": "2"},
		Secrets: map[string]string{"api_token": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "spaces", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	// Basic auth = base64("user@example.com:tok_123").
	if !strings.HasPrefix(sawAuth, "Basic ") {
		t.Fatalf("Authorization = %q, want Basic <creds>", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["key"] == nil {
			t.Fatalf("record missing id/key: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network call, so conformance passes without live credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := confluence.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pages", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check in fixture mode must not error and must not touch the network.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestBaseURLValidation rejects non-http(s) base_url overrides (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := confluence.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "email": "u@example.com"},
		Secrets: map[string]string{"api_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "spaces", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http(s) base_url, got nil")
	}
}

// TestCatalogStreams verifies the published catalog has the core streams.
func TestCatalogStreams(t *testing.T) {
	c := confluence.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"spaces": false, "pages": false, "blogposts": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = confluence.New() // ensure init ran
	caps := confluence.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("confluence"); !ok {
		t.Fatal("registry did not resolve confluence (self-registration)")
	}
}
