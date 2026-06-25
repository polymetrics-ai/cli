package clickupapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	clickupapi "polymetrics.ai/internal/connectors/clickup-api"
)

// TestReadTasksPaginatesAndAuthenticates is the red-first test for the ClickUp
// connector: personal-token auth (Authorization header, no Bearer prefix),
// page-based pagination over the tasks endpoint signalled by last_page, and
// record mapping. ClickUp returns {"tasks":[...],"last_page":bool} and the next
// page is requested with page=N (0-indexed).
func TestReadTasksPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/team/team_1/task" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		sawPages = append(sawPages, page)
		switch page {
		case "", "0":
			_, _ = w.Write([]byte(`{"tasks":[{"id":"t1","name":"Alpha","date_updated":"1700000000000"},{"id":"t2","name":"Beta","date_updated":"1700000001000"}],"last_page":false}`))
		case "1":
			_, _ = w.Write([]byte(`{"tasks":[{"id":"t3","name":"Gamma","date_updated":"1700000002000"}],"last_page":true}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"tasks":[],"last_page":true}`))
		}
	}))
	defer srv.Close()

	c := clickupapi.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "team_id": "team_1"},
		Secrets: map[string]string{"api_token": "pk_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tasks", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	// ClickUp personal tokens are sent raw, with no "Bearer " prefix.
	if sawAuth != "pk_test_123" {
		t.Fatalf("Authorization = %q, want pk_test_123 (no Bearer prefix)", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if len(sawPages) != 2 {
		t.Fatalf("requested %d pages (%v), want 2", len(sawPages), sawPages)
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestReadTeamsMapsRecords verifies the top-level teams stream extracts the
// {"teams":[...]} envelope without requiring any config scoping.
func TestReadTeamsMapsRecords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/team" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"teams":[{"id":"team_1","name":"Acme","color":"#fff"},{"id":"team_2","name":"Beta Co","color":"#000"}]}`))
	}))
	defer srv.Close()

	c := clickupapi.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "pk_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "teams", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] != "team_1" || got[0]["name"] != "Acme" {
		t.Fatalf("first record mapped wrong: %+v", got[0])
	}
}

// TestFixtureMode confirms the credential-free fixture path emits deterministic
// records for every core stream with no network access.
func TestFixtureMode(t *testing.T) {
	c := clickupapi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"teams", "spaces", "folders", "lists", "tasks"} {
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
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := clickupapi.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestRegistryResolves confirms self-registration via init() and that the
// connector is read-only (no write capability).
func TestRegistryResolves(t *testing.T) {
	_ = clickupapi.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("clickup-api")
	if !ok {
		t.Fatal("registry did not resolve clickup-api (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
}

// TestCatalogStreams confirms the published catalog lists the core streams.
func TestCatalogStreams(t *testing.T) {
	c := clickupapi.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"teams": false, "spaces": false, "folders": false, "lists": false, "tasks": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}
