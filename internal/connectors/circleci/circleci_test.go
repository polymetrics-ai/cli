package circleci_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/circleci"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the CircleCI
// connector: Circle-Token header auth, items[]/next_page_token pagination over
// two pages, and record mapping. Red until internal/connectors/circleci exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.Header.Get("Circle-Token")
		if r.URL.Path != "/project/gh/acme/widgets/pipeline" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page-token") {
		case "":
			_, _ = w.Write([]byte(`{"items":[{"id":"p1","number":1,"state":"created","created_at":"2026-01-01T00:00:00Z"},{"id":"p2","number":2,"state":"created","created_at":"2026-01-02T00:00:00Z"}],"next_page_token":"PAGE2"}`))
		case "PAGE2":
			_, _ = w.Write([]byte(`{"items":[{"id":"p3","number":3,"state":"created","created_at":"2026-01-03T00:00:00Z"}],"next_page_token":null}`))
		default:
			t.Errorf("unexpected page-token=%q", r.URL.Query().Get("page-token"))
			_, _ = w.Write([]byte(`{"items":[],"next_page_token":null}`))
		}
	}))
	defer srv.Close()

	c := circleci.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":     srv.URL,
			"project_slug": "gh/acme/widgets",
		},
		Secrets: map[string]string{"api_key": "circle_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pipelines", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "circle_test_123" {
		t.Fatalf("Circle-Token = %q, want circle_test_123", sawToken)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["created_at"] == nil {
			t.Fatalf("record missing id/created_at: %+v", rec)
		}
	}
}

// TestReadProjectSingleObject verifies the projects stream, which returns a
// single object (not an items[] list), maps to one record.
func TestReadProjectSingleObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/project/gh/acme/widgets" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"id":"proj_1","slug":"gh/acme/widgets","name":"widgets","organization_name":"acme","vcs_info":{"default_branch":"main"}}`))
	}))
	defer srv.Close()

	c := circleci.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":     srv.URL,
			"project_slug": "gh/acme/widgets",
		},
		Secrets: map[string]string{"api_key": "circle_test_123"},
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
	if got[0]["slug"] != "gh/acme/widgets" {
		t.Fatalf("project slug = %v, want gh/acme/widgets", got[0]["slug"])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// with no live credentials and no network access.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := circleci.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"projects", "pipelines", "workflows", "jobs"} {
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
		if got[0]["id"] == nil {
			t.Fatalf("Read(%s) fixture record missing id: %+v", stream, got[0])
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := circleci.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog includes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := circleci.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"projects": false, "pipelines": false, "workflows": false, "jobs": false}
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

// TestRegisteredReadOnly confirms self-registration via the registry and that
// the connector advertises read (not write) capability.
func TestRegisteredReadOnly(t *testing.T) {
	_ = circleci.New() // ensure init ran
	c := circleci.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("circleci"); !ok {
		t.Fatal("registry did not resolve circleci (self-registration)")
	}
}
