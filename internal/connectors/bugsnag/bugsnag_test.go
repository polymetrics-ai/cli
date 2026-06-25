package bugsnag_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bugsnag"
)

// TestReadOrganizationsPaginatesAndAuthenticates is the red-first test: it
// asserts the Bugsnag personal-token auth header (Authorization: token <t>),
// the mandatory X-Version: 2 header, Link-header rel="next" pagination across
// two pages of the top-level /user/organizations endpoint, and record mapping.
func TestReadOrganizationsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth, sawVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawVersion = r.Header.Get("X-Version")
		if r.URL.Path != "/user/organizations" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			w.Header().Set("Link", fmt.Sprintf(`<%s/user/organizations?page=2>; rel="next"`, baseOf(r)))
			_, _ = w.Write([]byte(`[{"id":"org_1","name":"Acme","slug":"acme"},{"id":"org_2","name":"Globex","slug":"globex"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"id":"org_3","name":"Initech","slug":"initech"}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := bugsnag.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"auth_token": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organizations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "token tok_abc" {
		t.Fatalf("Authorization = %q, want %q", sawAuth, "token tok_abc")
	}
	if sawVersion != "2" {
		t.Fatalf("X-Version = %q, want 2", sawVersion)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	if got[0]["id"] != "org_1" || got[0]["name"] != "Acme" {
		t.Fatalf("first record mapping wrong: %+v", got[0])
	}
}

// TestReadErrorsDiscoversParentProject verifies a project-scoped stream
// (errors) resolves its project_id from config and hits the right endpoint.
func TestReadErrorsWithProjectID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj_9/errors" {
			t.Errorf("unexpected path %q", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"err_1","error_class":"NPE","message":"boom","project_id":"proj_9"}]`))
	}))
	defer srv.Close()

	c := bugsnag.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "project_id": "proj_9"},
		Secrets: map[string]string{"auth_token": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "errors", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "err_1" || got[0]["error_class"] != "NPE" {
		t.Fatalf("errors mapping wrong: %+v", got)
	}
}

// TestReadProjectsDiscoversOrganization verifies a child stream (projects)
// auto-discovers its parent organization via /user/organizations when no
// organization_id is configured.
func TestReadProjectsDiscoversOrganization(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user/organizations":
			_, _ = w.Write([]byte(`[{"id":"org_x","name":"X","slug":"x"}]`))
		case "/organizations/org_x/projects":
			_, _ = w.Write([]byte(`[{"id":"proj_1","name":"Web","organization_id":"org_x"}]`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := bugsnag.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"auth_token": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "proj_1" {
		t.Fatalf("projects mapping wrong: %+v", got)
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := bugsnag.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"organizations", "projects", "errors", "events", "collaborators", "releases"} {
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
		if got[0]["id"] == nil {
			t.Fatalf("fixture Read(%s): record missing id: %+v", stream, got[0])
		}
	}
	// Check must also short-circuit in fixture mode without creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := bugsnag.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "bugsnag" {
		t.Fatalf("catalog connector = %q, want bugsnag", cat.Connector)
	}
	want := map[string]bool{"organizations": false, "projects": false, "errors": false, "events": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = bugsnag.New() // ensure init() ran
	caps := bugsnag.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("bugsnag is read-only; Write capability should be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("bugsnag"); !ok {
		t.Fatal("registry did not resolve bugsnag (self-registration)")
	}
}

// baseOf reconstructs the absolute base URL the test server is listening on so
// the handler can emit absolute Link-header next URLs (as Bugsnag does).
func baseOf(r *http.Request) string {
	return "http://" + r.Host
}
