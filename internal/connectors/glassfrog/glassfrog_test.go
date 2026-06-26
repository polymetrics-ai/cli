package glassfrog_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/glassfrog"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the GlassFrog
// connector: X-Auth-Token API-key header auth, page-number pagination over the
// resource-named array (circles[]), and record mapping. Red until
// internal/connectors/glassfrog exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("X-Auth-Token")
		sawAccept = r.Header.Get("Accept")
		if r.URL.Path != "/circles" {
			http.NotFound(w, r)
			return
		}
		// per_page=2 page-number pagination: full page -> there is a next page;
		// short page (1 record) -> stop.
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"circles":[{"id":1,"name":"General Company Circle","short_name":"GCC","organization_id":7},{"id":2,"name":"Engineering","short_name":"Eng","organization_id":7}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"circles":[{"id":3,"name":"Marketing","short_name":"Mkt","organization_id":7}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"circles":[]}`))
		}
	}))
	defer srv.Close()

	c := glassfrog.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "gf_test_token"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "circles", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "gf_test_token" {
		t.Fatalf("X-Auth-Token = %q, want gf_test_token", sawAuth)
	}
	if sawAccept != "application/json" {
		t.Fatalf("Accept = %q, want application/json", sawAccept)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["id"] == nil || got[0]["name"] == nil || got[0]["short_name"] == nil {
		t.Fatalf("record missing mapped fields: %+v", got[0])
	}
	if got[2]["name"] != "Marketing" {
		t.Fatalf("last record name = %v, want Marketing", got[2]["name"])
	}
}

// TestReadFixtureMode confirms credential-free fixture reads work for every
// stream so the conformance harness can run without live creds.
func TestReadFixtureMode(t *testing.T) {
	c := glassfrog.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"assignments", "circles", "people", "projects", "roles"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s): no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode (no network).
func TestCheckFixtureMode(t *testing.T) {
	c := glassfrog.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := glassfrog.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"assignments": false, "circles": false, "people": false, "projects": false, "roles": false}
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

// TestBadBaseURLRejected confirms SSRF-guard validation on base_url override.
func TestBadBaseURLRejected(t *testing.T) {
	c := glassfrog.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil"},
		Secrets: map[string]string{"api_key": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "circles", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with non-http base_url should be rejected")
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	c := glassfrog.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (GlassFrog API is read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("glassfrog"); !ok {
		t.Fatal("registry did not resolve glassfrog (self-registration)")
	}
}
