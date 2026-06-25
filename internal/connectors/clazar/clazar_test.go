package clazar_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/clazar"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Clazar
// connector. It asserts: (1) the OAuth2 client-credentials token is fetched from
// the configured token endpoint and the resulting bearer token is sent on data
// requests; (2) page-increment pagination walks two pages (page=1, page=2)
// stopping on a short final page; (3) records are extracted from the `results`
// array and mapped. Red until internal/connectors/clazar exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var tokenCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/authenticate/":
			tokenCalls++
			if r.Method != http.MethodPost {
				t.Errorf("token request method = %q, want POST", r.Method)
			}
			_ = r.ParseForm()
			if r.PostForm.Get("grant_type") != "client_credentials" {
				t.Errorf("grant_type = %q, want client_credentials", r.PostForm.Get("grant_type"))
			}
			if r.PostForm.Get("client_id") != "cid_123" || r.PostForm.Get("client_secret") != "secret_abc" {
				t.Errorf("client credentials not forwarded: id=%q", r.PostForm.Get("client_id"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"tok_xyz","token_type":"Bearer","expires_in":3600}`))
		case "/buyers":
			sawAuth = r.Header.Get("Authorization")
			switch r.URL.Query().Get("page") {
			case "1":
				_, _ = w.Write([]byte(`{"results":[{"id":"b1","name":"Acme","last_modified_at":"2026-01-01T00:00:00.000000Z"},{"id":"b2","name":"Globex","last_modified_at":"2026-01-02T00:00:00.000000Z"}],"next":"page=2"}`))
			case "2":
				_, _ = w.Write([]byte(`{"results":[{"id":"b3","name":"Initech","last_modified_at":"2026-01-03T00:00:00.000000Z"}],"next":null}`))
			default:
				t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
				_, _ = w.Write([]byte(`{"results":[]}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := clazar.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"page_size": "2",
		},
		Secrets: map[string]string{"client_id": "cid_123", "client_secret": "secret_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "buyers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !strings.HasPrefix(sawAuth, "Bearer ") || !strings.Contains(sawAuth, "tok_xyz") {
		t.Fatalf("Authorization = %q, want Bearer tok_xyz", sawAuth)
	}
	if tokenCalls == 0 {
		t.Fatal("token endpoint was never called")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["last_modified_at"] == nil {
			t.Fatalf("record missing id/last_modified_at: %+v", rec)
		}
		if rec["name"] == nil {
			t.Fatalf("record missing mapped name: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records with
// no network access, so the conformance harness can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := clazar.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "listings", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check must also short-circuit without creds in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams with
// primary keys and cursor fields.
func TestCatalogStreams(t *testing.T) {
	c := clazar.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"buyers": true, "listings": true, "contracts": true, "opportunities": true, "private_offers": true}
	seen := map[string]bool{}
	for _, s := range cat.Streams {
		seen[s.Name] = true
		if len(s.PrimaryKey) == 0 {
			t.Errorf("stream %q has no primary key", s.Name)
		}
		if len(s.CursorFields) == 0 {
			t.Errorf("stream %q has no cursor fields", s.Name)
		}
		if len(s.Fields) == 0 {
			t.Errorf("stream %q has no fields", s.Name)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("catalog missing core stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = clazar.New() // ensure init ran
	caps := clazar.New().Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("clazar is read-only; Write should be false, got %+v", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("clazar"); !ok {
		t.Fatal("registry did not resolve clazar (self-registration)")
	}
}
