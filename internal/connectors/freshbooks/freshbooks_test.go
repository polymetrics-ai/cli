package freshbooks_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/freshbooks"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the FreshBooks
// connector: Bearer auth with the oauth_access_token secret, FreshBooks
// page/pages pagination over response.result.clients, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPaths = append(sawPaths, r.URL.Path)
		if r.URL.Path != "/accounting/account/ACC123/users/clients" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"response":{"result":{"clients":[{"id":11,"organization":"Acme","email":"a@acme.test","updated":"2026-01-01 00:00:00"},{"id":12,"organization":"Beta","email":"b@beta.test","updated":"2026-01-02 00:00:00"}],"page":1,"pages":2,"per_page":2,"total":3}}}`))
		case "2":
			_, _ = w.Write([]byte(`{"response":{"result":{"clients":[{"id":13,"organization":"Gamma","email":"c@gamma.test","updated":"2026-01-03 00:00:00"}],"page":2,"pages":2,"per_page":2,"total":3}}}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"response":{"result":{"clients":[],"page":3,"pages":2,"per_page":2,"total":3}}}`))
		}
	}))
	defer srv.Close()

	c := freshbooks.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "account_id": "ACC123", "page_size": "2"},
		Secrets: map[string]string{"oauth_access_token": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages); paths=%v", len(got), sawPaths)
	}
	if len(sawPaths) != 2 {
		t.Fatalf("requests = %d, want 2 (pagination must stop after pages=2)", len(sawPaths))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["organization"] == nil {
			t.Fatalf("record missing id/organization: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork confirms credential-free fixture reads emit
// deterministic records without any HTTP call.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := freshbooks.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"clients", "invoices", "expenses", "payments", "items"} {
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
		if got[0]["id"] == nil {
			t.Fatalf("fixture Read(%s) record missing id: %+v", stream, got[0])
		}
	}

	// Check in fixture mode requires no creds and must succeed.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCheckRequiresSecret(t *testing.T) {
	c := freshbooks.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"account_id": "ACC123"}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check without oauth_access_token should fail")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := freshbooks.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com", "account_id": "ACC123"},
		Secrets: map[string]string{"oauth_access_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with ftp base_url should reject scheme, got %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := freshbooks.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"clients": false, "invoices": false, "expenses": false, "payments": false, "items": false}
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
			t.Fatalf("catalog missing expected stream %q", name)
		}
	}
}

func TestRegistryResolves(t *testing.T) {
	_ = freshbooks.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("freshbooks")
	if !ok {
		t.Fatal("registry did not resolve freshbooks (self-registration)")
	}
	if got.Name() != "freshbooks" {
		t.Fatalf("resolved connector name = %q, want freshbooks", got.Name())
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
