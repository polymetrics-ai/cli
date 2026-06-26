package eventbrite_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/eventbrite"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Eventbrite
// connector: Bearer auth, continuation-token pagination over the events[] array,
// and record mapping. Red until internal/connectors/eventbrite exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/organizations/123/events/" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("continuation") {
		case "":
			_, _ = w.Write([]byte(`{"pagination":{"has_more_items":true,"continuation":"tok2"},"events":[{"id":"ev_1","name":{"text":"Launch"},"changed":"2026-01-01T00:00:00Z","status":"live"},{"id":"ev_2","name":{"text":"Party"},"changed":"2026-01-02T00:00:00Z","status":"live"}]}`))
		case "tok2":
			_, _ = w.Write([]byte(`{"pagination":{"has_more_items":false,"continuation":"tok2"},"events":[{"id":"ev_3","name":{"text":"Demo"},"changed":"2026-01-03T00:00:00Z","status":"draft"}]}`))
		default:
			t.Errorf("unexpected continuation=%q", r.URL.Query().Get("continuation"))
			_, _ = w.Write([]byte(`{"pagination":{"has_more_items":false},"events":[]}`))
		}
	}))
	defer srv.Close()

	c := eventbrite.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "organization_id": "123"},
		Secrets: map[string]string{"private_token": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["changed"] == nil {
			t.Fatalf("record missing id/changed: %+v", rec)
		}
	}
	// name is a nested object on the wire ({"text": "..."}); the mapper should
	// flatten it to a string.
	if got[0]["name"] != "Launch" {
		t.Fatalf("name = %v, want flattened string \"Launch\"", got[0]["name"])
	}
}

// TestReadOrganizationsNoOrgID exercises the entry-point stream that needs no
// org/event id and reads from /users/me/organizations/.
func TestReadOrganizationsNoOrgID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/me/organizations/" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"pagination":{"has_more_items":false},"organizations":[{"id":"org_1","name":"Acme"}]}`))
	}))
	defer srv.Close()

	c := eventbrite.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"private_token": "tok_abc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organizations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read organizations: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "org_1" {
		t.Fatalf("organizations records = %+v, want one org_1", got)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network call, so conformance passes credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := eventbrite.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"organizations", "events", "attendees", "orders", "ticket_classes"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) returned no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := eventbrite.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := eventbrite.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "eventbrite" {
		t.Fatalf("catalog connector = %q, want eventbrite", cat.Connector)
	}
	want := map[string]bool{"organizations": false, "events": false, "attendees": false, "orders": false, "ticket_classes": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s has no primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing expected stream %q", name)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := eventbrite.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"private_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organizations", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected (SSRF guard)")
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = eventbrite.New() // ensure init ran
	c := eventbrite.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only source)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("eventbrite"); !ok {
		t.Fatal("registry did not resolve eventbrite (self-registration)")
	}
}
