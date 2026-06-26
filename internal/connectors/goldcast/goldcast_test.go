package goldcast_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/goldcast"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the
// Authorization: Token <access_key> header, DRF-style next-link pagination
// across two pages, and id-based record mapping for the events stream.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPaths []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPaths = append(sawPaths, r.URL.RequestURI())
		if r.URL.Path != "/event/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("offset") {
		case "":
			// First page: DRF envelope with an absolute next link.
			_, _ = w.Write([]byte(`{"count":3,"next":"` + srv.URL + `/event/?offset=2","results":[` +
				`{"id":"ev_1","title":"Launch","organization":"org_1"},` +
				`{"id":"ev_2","title":"Demo","organization":"org_1"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"count":3,"next":null,"results":[` +
				`{"id":"ev_3","title":"Recap","organization":"org_1"}]}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"count":0,"next":null,"results":[]}`))
		}
	}))
	defer srv.Close()

	c := goldcast.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"access_key": "ak_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Token ak_test_123" {
		t.Fatalf("Authorization = %q, want Token ak_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across two pages (paths=%v)", len(got), sawPaths)
	}
	if got[0]["id"] != "ev_1" || got[2]["id"] != "ev_3" {
		t.Fatalf("unexpected record ids: %v", got)
	}
	if got[0]["title"] != "Launch" {
		t.Fatalf("record mapping dropped title: %+v", got[0])
	}
}

// TestReadTopLevelArray confirms the connector also handles the documented
// raw top-level JSON array response shape (no DRF envelope), single page.
func TestReadTopLevelArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/core/organization/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"org_1","name":"Acme"},{"id":"org_2","name":"Globex"}]`))
	}))
	defer srv.Close()

	c := goldcast.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"access_key": "ak_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "organizations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 || got[0]["id"] != "org_1" || got[1]["name"] != "Globex" {
		t.Fatalf("top-level array read = %v, want 2 orgs", got)
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access (used by the conformance harness without live creds).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := goldcast.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check must also short-circuit in fixture mode (no creds, no network).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCheckRequiresSecret(t *testing.T) {
	c := goldcast.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://customapi.goldcast.io"}}
	err := c.Check(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "access_key") {
		t.Fatalf("Check without secret = %v, want access_key error", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := goldcast.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "goldcast" {
		t.Fatalf("catalog connector = %q, want goldcast", cat.Connector)
	}
	want := map[string]bool{"organizations": false, "events": false, "agenda_items": false, "discussion_groups": false, "tracks": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 || s.PrimaryKey[0] != "id" {
			t.Fatalf("stream %q primary key = %v, want [id]", s.Name, s.PrimaryKey)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	c := goldcast.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
	if caps.Write {
		t.Fatalf("goldcast is read-only, Write should be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("goldcast"); !ok {
		t.Fatal("registry did not resolve goldcast (self-registration)")
	}
}
