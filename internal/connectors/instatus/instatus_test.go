package instatus_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/instatus"
)

// TestReadPagesPaginatesAndAuthenticates is the red-first test for the Instatus
// connector: Bearer auth, page/per_page pagination over a top-level JSON array
// (stop on a short page), and record mapping for the top-level `pages` stream.
func TestReadPagesPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v2/pages" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// Full page of 2 == per_page, so the harvester must request page 2.
			_, _ = w.Write([]byte(`[{"id":"pg_1","subdomain":"a","name":"A","status":"UP"},{"id":"pg_2","subdomain":"b","name":"B","status":"UP"}]`))
		case "2":
			// Short page => stop.
			_, _ = w.Write([]byte(`[{"id":"pg_3","subdomain":"c","name":"C","status":"UP"}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := instatus.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "ik_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pages", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer ik_test_123" {
		t.Fatalf("Authorization = %q, want Bearer ik_test_123", sawAuth)
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

// TestReadComponentsUsesPageID exercises a parent-scoped stream: components are
// read from /v2/{page_id}/components and require a page_id config value.
func TestReadComponentsUsesPageID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/pg_42/components" {
			t.Errorf("unexpected path %q", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"cmp_1","name":"API","status":"OPERATIONAL"}]`))
	}))
	defer srv.Close()

	c := instatus.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_id": "pg_42"},
		Secrets: map[string]string{"api_key": "ik_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "components", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "cmp_1" {
		t.Fatalf("components = %+v, want one cmp_1 record", got)
	}
}

// TestComponentsRequirePageID ensures a parent-scoped stream errors clearly when
// page_id is missing rather than building a malformed URL.
func TestComponentsRequirePageID(t *testing.T) {
	c := instatus.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "https://api.instatus.com"},
		Secrets: map[string]string{"api_key": "ik_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "components", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read(components) without page_id should error")
	}
}

// TestFixtureMode confirms the credential-free deterministic path works without
// any network access so conformance can run without live creds.
func TestFixtureMode(t *testing.T) {
	c := instatus.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"pages", "components", "incidents", "maintenances"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("Read(%s) fixture emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogAndRegistry verifies the published catalog and self-registration.
func TestCatalogAndRegistry(t *testing.T) {
	c := instatus.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}

	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("instatus is read-only; Write should be false")
	}

	r := connectors.NewRegistry()
	if _, ok := r.Get("instatus"); !ok {
		t.Fatal("registry did not resolve instatus (self-registration)")
	}
}
