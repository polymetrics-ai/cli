package gitlab_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/gitlab"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the GitLab
// connector: Bearer auth, GitLab page/per_page + Link-header rel="next"
// pagination over a top-level array, and record mapping. Red until
// internal/connectors/gitlab exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/projects" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// Advertise a next page via the RFC 5988 Link header.
			w.Header().Set("Link", `<`+srvURL+`/projects?page=2&per_page=2>; rel="next"`)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"id":1,"name":"alpha","created_at":"2026-01-01T00:00:00Z"},{"id":2,"name":"beta","created_at":"2026-01-02T00:00:00Z"}]`))
		case "2":
			// No Link header => last page.
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"id":3,"name":"gamma","created_at":"2026-01-03T00:00:00Z"}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := gitlab.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"credentials.access_token": "glpat-secret-123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer glpat-secret-123" {
		t.Fatalf("Authorization = %q, want Bearer glpat-secret-123", sawAuth)
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

// TestFixtureModeReadsWithoutNetwork confirms the credential-free fixture path so
// conformance can run without live creds (mirrors the stripe template intent).
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := gitlab.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "groups", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode produced no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode without a
// network call or credentials.
func TestCheckFixtureMode(t *testing.T) {
	c := gitlab.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := gitlab.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"projects": false, "groups": false, "users": false, "issues": false}
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

// TestBaseURLRejectsBadScheme confirms SSRF guard on base_url override.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := gitlab.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"credentials.access_token": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected")
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = gitlab.New() // ensure init ran
	c := gitlab.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("gitlab"); !ok {
		t.Fatal("registry did not resolve gitlab (self-registration)")
	}
}
