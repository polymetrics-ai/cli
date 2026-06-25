package appcues_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/appcues"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Appcues
// connector. Appcues uses HTTP Basic auth (username + password) and returns a
// top-level JSON array per resource under accounts/{account_id}/{resource}. The
// connector advances pages with the `page` query param and stops on a short page.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPaths = append(sawPaths, r.URL.Path)
		if r.URL.Path != "/accounts/acct_123/flows" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// Full page (page size 2) => connector should request page 2.
			_, _ = w.Write([]byte(`[{"id":"flow_1","name":"Welcome","state":"PUBLISHED"},{"id":"flow_2","name":"Onboarding","state":"DRAFT"}]`))
		case "2":
			// Short page => terminal.
			_, _ = w.Write([]byte(`[{"id":"flow_3","name":"Survey","state":"PUBLISHED"}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := appcues.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"account_id": "acct_123",
			"username":   "key_abc",
			"page_size":  "2",
		},
		Secrets: map[string]string{"password": "secret_xyz"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "flows", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("key_abc:secret_xyz"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages); paths=%v", len(got), sawPaths)
	}
	if got[0]["id"] != "flow_1" || got[0]["name"] != "Welcome" {
		t.Fatalf("record[0] mapping wrong: %+v", got[0])
	}
	if got[2]["id"] != "flow_3" {
		t.Fatalf("record[2] id = %v, want flow_3", got[2]["id"])
	}
	if len(sawPaths) != 2 {
		t.Fatalf("expected 2 page requests, got %d: %v", len(sawPaths), sawPaths)
	}
}

// TestFixtureModeNoNetwork verifies the credential-free fixture path used by the
// conformance harness emits deterministic records without any network call.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := appcues.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "segments", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for i, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record %d missing id: %+v", i, rec)
		}
	}
	// Fixture Check must succeed without credentials or network.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams asserts the published catalog exposes the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := appcues.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "appcues" {
		t.Fatalf("catalog connector = %q, want appcues", cat.Connector)
	}
	want := map[string]bool{"flows": false, "segments": false, "tags": false, "checklists": false, "banners": false}
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

// TestBaseURLSSRFValidation ensures an override with a bad scheme is rejected.
func TestBaseURLSSRFValidation(t *testing.T) {
	c := appcues.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   "file:///etc/passwd",
			"account_id": "acct_123",
			"username":   "u",
		},
		Secrets: map[string]string{"password": "p"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "flows", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for file:// base_url, got nil")
	}
}

// TestRegisteredReadOnly verifies self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	c := appcues.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	got, ok := r.Get("appcues")
	if !ok {
		t.Fatal("registry did not resolve appcues (self-registration)")
	}
	if got.Name() != "appcues" {
		t.Fatalf("resolved connector name = %q, want appcues", got.Name())
	}
}
