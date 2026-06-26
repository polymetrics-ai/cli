package insightful_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/insightful"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Insightful
// connector. Insightful list endpoints (e.g. /employee) return a top-level JSON
// array; pagination is driven by an optional `next` cursor token echoed back as
// a `next` request parameter. This test asserts Bearer auth, cursor pagination
// across two pages, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/employee" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("next") {
		case "":
			// First page: two records plus a forward cursor.
			_, _ = w.Write([]byte(`{"data":[{"id":"emp_1","name":"Ada","email":"ada@example.com","updatedAt":1700000000},{"id":"emp_2","name":"Grace","email":"grace@example.com","updatedAt":1700000100}],"next":"cursor_2"}`))
		case "cursor_2":
			// Last page: one record, no next token (stop condition).
			_, _ = w.Write([]byte(`{"data":[{"id":"emp_3","name":"Katherine","email":"k@example.com","updatedAt":1700000200}]}`))
		default:
			t.Errorf("unexpected next=%q", r.URL.Query().Get("next"))
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := insightful.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "employee", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_test_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["updatedAt"] == nil {
			t.Fatalf("record missing id/updatedAt: %+v", rec)
		}
	}
}

// TestReadTopLevelArray confirms the connector also handles endpoints that
// return a bare top-level JSON array (no envelope), which is the common
// Insightful response shape for list resources.
func TestReadTopLevelArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/team" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"team_1","name":"Engineering"},{"id":"team_2","name":"Design"}]`))
	}))
	defer srv.Close()

	c := insightful.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "tok"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "team", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] != "team_1" {
		t.Fatalf("first team id = %v, want team_1", got[0]["id"])
	}
}

// TestFixtureMode confirms the credential-free fixture path emits deterministic
// records without any network access, so conformance passes without live creds.
func TestFixtureMode(t *testing.T) {
	c := insightful.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "employee", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check must also short-circuit in fixture mode (no creds).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestRegisteredReadOnly asserts self-registration via the process-global
// factory registry and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	c := insightful.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only source)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("insightful"); !ok {
		t.Fatal("registry did not resolve insightful (self-registration)")
	}
}

// TestCatalogStreams asserts the published catalog exposes the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := insightful.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "insightful" {
		t.Fatalf("catalog connector = %q, want insightful", cat.Connector)
	}
	want := map[string]bool{"employee": false, "team": false, "projects": false, "directory": false}
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
