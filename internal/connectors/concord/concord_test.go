package concord_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/concord"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Concord
// connector: X-API-KEY auth, page-increment pagination over an org-scoped
// agreements endpoint (root-array record selector), and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("X-API-KEY")
		sawPaths = append(sawPaths, r.URL.Path)
		if r.URL.Path != "/organizations/42/agreements" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "0":
			// Full page (size 2) -> there is a next page.
			_, _ = w.Write([]byte(`[{"uid":"a1","title":"Agreement One"},{"uid":"a2","title":"Agreement Two"}]`))
		case "1":
			// Short page -> last page.
			_, _ = w.Write([]byte(`[{"uid":"a3","title":"Agreement Three"}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := concord.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":        srv.URL,
			"organization_id": "42",
			"page_size":       "2",
		},
		Secrets: map[string]string{"api_key": "test-key-abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "agreements", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "test-key-abc" {
		t.Fatalf("X-API-KEY = %q, want test-key-abc", sawAPIKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages; paths=%v", len(got), sawPaths)
	}
	for _, rec := range got {
		if rec["uid"] == nil {
			t.Fatalf("record missing uid: %+v", rec)
		}
	}
}

// TestReadUserOrganizations exercises a non-org-scoped stream whose records live
// under an "organizations" field path (not the root array).
func TestReadUserOrganizations(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/me/organizations" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"organizations":[{"id":1,"name":"Org A"},{"id":2,"name":"Org B"}]}`))
	}))
	defer srv.Close()

	c := concord.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "k"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "user_organizations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil || got[0]["name"] == nil {
		t.Fatalf("record missing id/name: %+v", got[0])
	}
}

// TestFixtureMode confirms the credential-free deterministic path emits records
// without any network access.
func TestFixtureMode(t *testing.T) {
	c := concord.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"agreements", "user_organizations", "folders", "reports", "tags"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read %s: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s emitted 0 records", stream)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams asserts the published catalog has the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := concord.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"agreements": false, "user_organizations": false, "folders": false, "reports": false, "tags": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
			if len(s.PrimaryKey) == 0 {
				t.Fatalf("stream %s missing primary key", s.Name)
			}
			if len(s.Fields) == 0 {
				t.Fatalf("stream %s missing fields", s.Name)
			}
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = concord.New() // ensure init ran
	c := concord.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("concord"); !ok {
		t.Fatal("registry did not resolve concord (self-registration)")
	}
}

// TestBaseURLValidation rejects non-http(s) base_url overrides (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := concord.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tags", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http base_url, got nil")
	}
}
