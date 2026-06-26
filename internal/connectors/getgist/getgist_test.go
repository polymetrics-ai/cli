package getgist_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/getgist"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Gist
// connector: Bearer auth, page-number pagination over the resource-named array
// (e.g. "contacts") with a pages.next link, and record mapping across 2 pages.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/contacts" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page") {
		case "", "1":
			// Full page (per_page default) signals there may be more; include a
			// pages.next link so the connector requests page 2.
			_, _ = w.Write([]byte(`{"contacts":[{"id":1,"email":"a@example.com","type":"user","updated_at":1700000000},{"id":2,"email":"b@example.com","type":"lead","updated_at":1700000100}],"pages":{"next":"https://api.getgist.com/contacts?page=2"}}`))
		case "2":
			_, _ = w.Write([]byte(`{"contacts":[{"id":3,"email":"c@example.com","type":"user","updated_at":1700000200}],"pages":{"next":null}}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"contacts":[],"pages":{"next":null}}`))
		}
	}))
	defer srv.Close()

	c := getgist.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "gist_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer gist_test_123" {
		t.Fatalf("Authorization = %q, want Bearer gist_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["email"] == nil {
			t.Fatalf("record missing id/email: %+v", rec)
		}
	}
}

// TestReadStopsOnShortPage confirms the connector stops when a page returns
// fewer than page_size records even if no pages.next is consulted, guarding
// against infinite loops on APIs that omit the link.
func TestReadStopsOnShortPage(t *testing.T) {
	var pages int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tags" {
			http.NotFound(w, r)
			return
		}
		pages++
		// One short page (1 record < page_size 5) and no next link.
		_, _ = w.Write([]byte(`{"tags":[{"id":10,"name":"vip"}]}`))
	}))
	defer srv.Close()

	c := getgist.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "5"},
		Secrets: map[string]string{"api_key": "gist_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tags", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if pages != 1 {
		t.Fatalf("requested %d pages, want 1 (short page should stop)", pages)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
}

// TestFixtureMode confirms credential-free deterministic reads with no network.
func TestFixtureMode(t *testing.T) {
	c := getgist.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil {
		t.Fatalf("fixture record missing id: %+v", got[0])
	}
	// Check must also short-circuit without creds in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := getgist.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "getgist" {
		t.Fatalf("catalog connector = %q, want getgist", cat.Connector)
	}
	want := map[string]bool{"contacts": false, "tags": false, "segments": false, "campaigns": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := getgist.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with ftp base_url err = %v, want base_url scheme rejection", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = getgist.New() // ensure init ran
	c := getgist.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only API)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("getgist"); !ok {
		t.Fatal("registry did not resolve getgist (self-registration)")
	}
}
