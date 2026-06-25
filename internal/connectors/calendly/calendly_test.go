package calendly_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/calendly"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Calendly
// connector: Bearer auth, Calendly's pagination.next_page cursor over the
// collection[] array across two pages, and record mapping. The organization
// query param is resolved from /users/me and threaded into each request.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawOrg string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		switch r.URL.Path {
		case "/users/me":
			_, _ = w.Write([]byte(`{"resource":{"uri":"https://api.calendly.com/users/U1","current_organization":"https://api.calendly.com/organizations/ORG1","name":"Ada"}}`))
		case "/scheduled_events":
			sawOrg = r.URL.Query().Get("organization")
			if r.URL.Query().Get("page_token") == "" {
				// Page 1: hand back a next_page that points back at this server.
				next := "http://" + r.Host + "/scheduled_events?organization=" + r.URL.Query().Get("organization") + "&page_token=TOK2"
				_, _ = w.Write([]byte(`{"collection":[{"uri":"https://api.calendly.com/scheduled_events/E1","name":"Intro","status":"active","start_time":"2026-01-01T10:00:00Z"},{"uri":"https://api.calendly.com/scheduled_events/E2","name":"Demo","status":"active","start_time":"2026-01-02T10:00:00Z"}],"pagination":{"count":2,"next_page":"` + next + `","next_page_token":"TOK2"}}`))
				return
			}
			_, _ = w.Write([]byte(`{"collection":[{"uri":"https://api.calendly.com/scheduled_events/E3","name":"Review","status":"active","start_time":"2026-01-03T10:00:00Z"}],"pagination":{"count":1,"next_page":null,"next_page_token":null}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := calendly.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "cal_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "scheduled_events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer cal_test_123" {
		t.Fatalf("Authorization = %q, want Bearer cal_test_123", sawAuth)
	}
	if sawOrg != "https://api.calendly.com/organizations/ORG1" {
		t.Fatalf("organization param = %q, want resolved org from /users/me", sawOrg)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["start_time"] == nil {
			t.Fatalf("record missing id/start_time: %+v", rec)
		}
	}
	// The uri must be reduced to a trailing id for the primary key.
	if got[0]["id"] != "E1" {
		t.Fatalf("first record id = %v, want E1", got[0]["id"])
	}
}

// TestFixtureMode confirms the credential-free path emits deterministic records
// so conformance runs without live creds.
func TestFixtureMode(t *testing.T) {
	c := calendly.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"event_types", "scheduled_events", "organization_memberships", "groups", "users"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
	// Check must short-circuit in fixture mode with no creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCheckResolvesUser exercises a live-style Check against /users/me.
func TestCheckResolvesUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/me" {
			http.NotFound(w, r)
			return
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"resource":{"uri":"https://api.calendly.com/users/U1","current_organization":"https://api.calendly.com/organizations/ORG1"}}`))
	}))
	defer srv.Close()

	c := calendly.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "cal_test_123"},
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check: %v", err)
	}
}

// TestCatalogStreams asserts the published catalog and metadata shape.
func TestCatalogStreams(t *testing.T) {
	c := calendly.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"event_types": false, "scheduled_events": false, "organization_memberships": false, "groups": false, "users": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing expected stream %q", name)
		}
	}
	if !c.Metadata().Capabilities.Read {
		t.Fatal("metadata should advertise Read")
	}
}

// TestRegistryResolves confirms self-registration via init().
func TestRegistryResolves(t *testing.T) {
	_ = calendly.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("calendly"); !ok {
		t.Fatal("registry did not resolve calendly (self-registration)")
	}
}
