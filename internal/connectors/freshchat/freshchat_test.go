package freshchat_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/freshchat"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Freshchat
// connector: Bearer auth (api_key), Freshchat page/items_per_page pagination
// over the "agents" wrapper array across two pages, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/agents" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		switch page {
		case "1":
			// A full page (== items_per_page) so the loop requests page 2.
			_, _ = w.Write([]byte(`{"agents":[` +
				`{"id":"a1","email":"a1@example.com","first_name":"Ada"},` +
				`{"id":"a2","email":"a2@example.com","first_name":"Bert"}` +
				`],"pagination":{"current_page":1,"total_pages":2}}`))
		case "2":
			// A short page (< items_per_page) so the loop stops.
			_, _ = w.Write([]byte(`{"agents":[` +
				`{"id":"a3","email":"a3@example.com","first_name":"Cleo"}` +
				`],"pagination":{"current_page":2,"total_pages":2}}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"agents":[]}`))
		}
	}))
	defer srv.Close()

	c := freshchat.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":     srv.URL,
			"page_size":    "2",
			"account_name": "freshfoods",
		},
		Secrets: map[string]string{"api_key": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "agents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
	if got[0]["email"] != "a1@example.com" || got[0]["first_name"] != "Ada" {
		t.Fatalf("mapping wrong for first record: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// with no network access, so conformance passes without live creds.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := freshchat.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var n int
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		n++
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if n == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	// Check must also short-circuit without creds in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams asserts the connector publishes the core read-only streams.
func TestCatalogStreams(t *testing.T) {
	c := freshchat.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"agents": false, "users": false, "groups": false, "channels": false, "roles": false}
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

// TestRegisteredReadOnly confirms self-registration via the connector registry
// and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = freshchat.New() // ensure init ran
	caps := freshchat.New().Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("freshchat"); !ok {
		t.Fatal("registry did not resolve freshchat (self-registration)")
	}
}

// TestBaseURLRejectsBadScheme guards SSRF: a non-http(s) base_url is rejected.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := freshchat.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "agents", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http base_url scheme")
	}
}
