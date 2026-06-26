package gitbook_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/gitbook"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the GitBook
// connector: Bearer auth on Authorization, GitBook cursor pagination over the
// items[] array with the next page cursor at next.page passed back as ?page=,
// and record mapping. Red until internal/connectors/gitbook exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/orgs" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "":
			_, _ = w.Write([]byte(`{"items":[{"id":"org_1","title":"Alpha"},{"id":"org_2","title":"Beta"}],"next":{"page":"cursor_2"}}`))
		case "cursor_2":
			_, _ = w.Write([]byte(`{"items":[{"id":"org_3","title":"Gamma"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"items":[]}`))
		}
	}))
	defer srv.Close()

	c := gitbook.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"access_token": "gb_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organizations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer gb_test_123" {
		t.Fatalf("Authorization = %q, want Bearer gb_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
	if got[0]["title"] != "Alpha" {
		t.Fatalf("first record title = %v, want Alpha", got[0]["title"])
	}
}

// TestReadUserSingleObject confirms the /user endpoint (a single object, not a
// paginated list) is mapped to exactly one record.
func TestReadUserSingleObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"id":"user_1","displayName":"Ada","email":"ada@example.com"}`))
	}))
	defer srv.Close()

	c := gitbook.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"access_token": "gb_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read users: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["id"] != "user_1" || got[0]["display_name"] != "Ada" {
		t.Fatalf("user record = %+v, want id=user_1 display_name=Ada", got[0])
	}
}

// TestFixtureMode confirms credential-free fixture reads emit deterministic
// records without any network call.
func TestFixtureMode(t *testing.T) {
	c := gitbook.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"users", "organizations", "org_members", "content"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read %s: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read %s emitted 0 records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

// TestCheckFixture confirms Check short-circuits in fixture mode (no network,
// no creds required).
func TestCheckFixture(t *testing.T) {
	c := gitbook.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams
// with primary keys.
func TestCatalogStreams(t *testing.T) {
	c := gitbook.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"users": false, "organizations": false, "org_members": false, "content": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
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
	_ = gitbook.New() // ensure init ran
	c := gitbook.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("gitbook"); !ok {
		t.Fatal("registry did not resolve gitbook (self-registration)")
	}
}
