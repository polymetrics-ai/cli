package buildkite_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/buildkite"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Buildkite
// connector: Bearer auth, RFC 5988 Link-header pagination across two pages, and
// record mapping for the pipelines stream. Red until
// internal/connectors/buildkite exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/organizations/acme/pipelines" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// rel="next" pointing at page 2 — connector must follow it.
			next := fmt.Sprintf("%s/organizations/acme/pipelines?page=2&per_page=100", srv.URL)
			w.Header().Set("Link", fmt.Sprintf("<%s>; rel=\"next\"", next))
			_, _ = w.Write([]byte(`[{"id":"p1","slug":"web","name":"Web","created_at":"2026-01-01T00:00:00Z"},{"id":"p2","slug":"api","name":"API","created_at":"2026-01-02T00:00:00Z"}]`))
		case "2":
			// No Link header => last page.
			_, _ = w.Write([]byte(`[{"id":"p3","slug":"infra","name":"Infra","created_at":"2026-01-03T00:00:00Z"}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := buildkite.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "organization": "acme"},
		Secrets: map[string]string{"api_key": "bkua_secret_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pipelines", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer bkua_secret_123" {
		t.Fatalf("Authorization = %q, want Bearer bkua_secret_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["slug"] == nil {
			t.Fatalf("record missing id/slug: %+v", rec)
		}
	}
}

// TestReadOrganizationsStreamNoOrgRequired confirms the top-level organizations
// stream reads from /organizations and does not need an organization slug.
func TestReadOrganizationsStreamNoOrgRequired(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/organizations" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"o1","slug":"acme","name":"Acme","created_at":"2026-01-01T00:00:00Z"}]`))
	}))
	defer srv.Close()

	c := buildkite.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "bkua_secret_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organizations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read organizations: %v", err)
	}
	if len(got) != 1 || got[0]["slug"] != "acme" {
		t.Fatalf("organizations = %+v, want one acme record", got)
	}
}

// TestReadFixtureMode exercises the credential-free deterministic path used by
// the conformance harness.
func TestReadFixtureMode(t *testing.T) {
	c := buildkite.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pipelines", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

func TestCheckFixtureModeNoNetwork(t *testing.T) {
	c := buildkite.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = buildkite.New() // ensure init ran
	c := buildkite.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("buildkite"); !ok {
		t.Fatal("registry did not resolve buildkite (self-registration)")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := buildkite.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"organizations": false, "pipelines": false, "builds": false, "agents": false}
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
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}
