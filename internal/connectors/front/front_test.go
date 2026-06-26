package front_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/front"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Front
// connector: Bearer auth against the Authorization header, Front body-cursor
// pagination over _results[] driven by _pagination.next (an absolute next-page
// URL), and record mapping. Red until internal/connectors/front exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var hits int
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/contacts" {
			http.NotFound(w, r)
			return
		}
		hits++
		if r.URL.Query().Get("page_token") == "" {
			// Page 1: a single contact plus a next-page URL in the body.
			next := srv.URL + "/contacts?page_token=PAGE2"
			_, _ = w.Write([]byte(`{"_results":[{"id":"crd_1","name":"Ada"}],"_pagination":{"next":"` + next + `"}}`))
			return
		}
		// Page 2: a single contact and no further pages.
		_, _ = w.Write([]byte(`{"_results":[{"id":"crd_2","name":"Grace"}],"_pagination":{"next":null}}`))
	}))
	defer srv.Close()

	c := front.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "fk_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer fk_test_123" {
		t.Fatalf("Authorization = %q, want Bearer fk_test_123", sawAuth)
	}
	if hits != 2 {
		t.Fatalf("server hits = %d, want 2 pages", hits)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (across 2 pages)", len(got))
	}
	if got[0]["id"] != "crd_1" || got[1]["id"] != "crd_2" {
		t.Fatalf("record ids = %v / %v, want crd_1 / crd_2", got[0]["id"], got[1]["id"])
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network call (credential-free conformance).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := front.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "conversations", Config: cfg}, func(rec connectors.Record) error {
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

	// Check must succeed in fixture mode with no secret.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams
// with primary keys.
func TestCatalogStreams(t *testing.T) {
	c := front.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"contacts": false, "conversations": false, "inboxes": false, "tags": false, "teammates": false}
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

// TestRegistryResolution confirms the connector self-registers and resolves via
// the shared registry.
func TestRegistryResolution(t *testing.T) {
	_ = front.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("front"); !ok {
		t.Fatal("registry did not resolve front (self-registration)")
	}
}

// TestBaseURLRejectsBadScheme confirms SSRF guard on base_url override.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := front.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "fk_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with bad base_url scheme = %v, want base_url error", err)
	}
}
