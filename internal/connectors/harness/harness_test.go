package harness_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/harness"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Harness
// connector: x-api-key auth, NextGen page-index pagination over data.content,
// and nested-record mapping. The Harness NextGen list envelope is
// {"data":{"content":[{"organization":{...}}],"pageIndex":N,"totalPages":N}}.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	var sawAccount string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("x-api-key")
		sawAccount = r.URL.Query().Get("accountIdentifier")
		if r.URL.Path != "/ng/api/organizations" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("pageIndex") {
		case "", "0":
			_, _ = w.Write([]byte(`{"status":"SUCCESS","data":{"totalPages":2,"totalItems":2,"pageItemCount":1,"pageSize":1,"pageIndex":0,"empty":false,"content":[{"organization":{"identifier":"org_1","name":"Org One","accountIdentifier":"acct_123","description":"first"}}]}}`))
		case "1":
			_, _ = w.Write([]byte(`{"status":"SUCCESS","data":{"totalPages":2,"totalItems":2,"pageItemCount":1,"pageSize":1,"pageIndex":1,"empty":false,"content":[{"organization":{"identifier":"org_2","name":"Org Two","accountIdentifier":"acct_123","description":"second"}}]}}`))
		default:
			t.Errorf("unexpected pageIndex=%q", r.URL.Query().Get("pageIndex"))
			_, _ = w.Write([]byte(`{"status":"SUCCESS","data":{"totalPages":2,"totalItems":2,"pageItemCount":0,"pageSize":1,"pageIndex":2,"empty":true,"content":[]}}`))
		}
	}))
	defer srv.Close()

	c := harness.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "account_id": "acct_123", "page_size": "1"},
		Secrets: map[string]string{"api_key": "pat.abc.123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organizations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "pat.abc.123" {
		t.Fatalf("x-api-key = %q, want pat.abc.123", sawKey)
	}
	if sawAccount != "acct_123" {
		t.Fatalf("accountIdentifier = %q, want acct_123", sawAccount)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (2 pages)", len(got))
	}
	if got[0]["identifier"] != "org_1" || got[1]["identifier"] != "org_2" {
		t.Fatalf("unexpected identifiers: %+v", got)
	}
	for _, rec := range got {
		if rec["identifier"] == nil || rec["name"] == nil {
			t.Fatalf("record missing identifier/name: %+v", rec)
		}
	}
}

// TestReadProjectsNestedMapping verifies the projects stream unwraps the
// {"project":{...}} envelope and carries orgIdentifier through.
func TestReadProjectsNestedMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ng/api/projects" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"status":"SUCCESS","data":{"totalPages":1,"totalItems":1,"pageItemCount":1,"pageSize":50,"pageIndex":0,"empty":false,"content":[{"project":{"identifier":"proj_1","orgIdentifier":"org_1","name":"Proj One","color":"#0063F7","modules":["CD"]}}]}}`))
	}))
	defer srv.Close()

	c := harness.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "account_id": "acct_123"},
		Secrets: map[string]string{"api_key": "pat.abc.123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["identifier"] != "proj_1" || got[0]["org_identifier"] != "org_1" {
		t.Fatalf("unexpected project mapping: %+v", got[0])
	}
}

// TestFixtureModeNeedsNoNetwork confirms fixture mode emits deterministic records
// without any HTTP credentials or server, which is what conformance relies on.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := harness.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"organizations", "projects", "services"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
		for _, rec := range got {
			if rec["identifier"] == nil {
				t.Fatalf("fixture %s record missing identifier: %+v", stream, rec)
			}
		}
	}
}

// TestCheckFixtureSkipsNetwork confirms Check short-circuits in fixture mode.
func TestCheckFixtureSkipsNetwork(t *testing.T) {
	c := harness.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF check on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := harness.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "account_id": "acct_123"},
		Secrets: map[string]string{"api_key": "pat.abc.123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organizations", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	c := harness.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("harness"); !ok {
		t.Fatal("registry did not resolve harness (self-registration)")
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := harness.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"organizations": false, "projects": false, "services": false}
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
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}
