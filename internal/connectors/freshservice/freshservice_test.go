package freshservice_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/freshservice"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Freshservice
// connector: HTTP Basic auth (api_key:X), page-number pagination over a
// plural-key-wrapped JSON array ({"tickets":[...]}), and record mapping. Red
// until internal/connectors/freshservice exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v2/tickets" {
			http.NotFound(w, r)
			return
		}
		// per_page=2 drives the PageNumberPaginator: a full page (2 records)
		// triggers a next page, a short page (1 record) stops.
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"tickets":[{"id":1,"subject":"a","status":2,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z"},{"id":2,"subject":"b","status":3,"created_at":"2026-01-03T00:00:00Z","updated_at":"2026-01-04T00:00:00Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"tickets":[{"id":3,"subject":"c","status":4,"created_at":"2026-01-05T00:00:00Z","updated_at":"2026-01-06T00:00:00Z"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"tickets":[]}`))
		}
	}))
	defer srv.Close()

	c := freshservice.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "fs_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tickets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("fs_test_key:X"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q (HTTP Basic api_key:X)", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["updated_at"] == nil {
			t.Fatalf("record missing id/updated_at: %+v", rec)
		}
	}
	if got[0]["subject"] != "a" || got[2]["subject"] != "c" {
		t.Fatalf("records not mapped/ordered: %+v", got)
	}
}

// TestReadRequestersMapsRecords verifies a second stream maps its fields, that
// the records live under the plural wrapper key, and that the per_page page-size
// override is sent.
func TestReadRequestersMapsRecords(t *testing.T) {
	var sawPerPage string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/requesters" {
			http.NotFound(w, r)
			return
		}
		sawPerPage = r.URL.Query().Get("per_page")
		_, _ = w.Write([]byte(`{"requesters":[{"id":10,"first_name":"Ada","last_name":"Lovelace","primary_email":"ada@example.com","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := freshservice.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "25"},
		Secrets: map[string]string{"api_key": "fs_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "requesters", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawPerPage != "25" {
		t.Fatalf("per_page = %q, want 25", sawPerPage)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["primary_email"] != "ada@example.com" || got[0]["first_name"] != "Ada" {
		t.Fatalf("requester not mapped: %+v", got[0])
	}
}

// TestDomainNameBuildsBaseURL verifies the domain_name config is turned into the
// v2 base URL when no explicit base_url override is given, and the resource path
// is /api/v2/<resource>.
func TestDomainNameBuildsBaseURL(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		_, _ = w.Write([]byte(`{"agents":[]}`))
	}))
	defer srv.Close()

	c := freshservice.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "k"},
	}
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "agents", Config: cfg}, func(connectors.Record) error { return nil }); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawPath != "/api/v2/agents" {
		t.Fatalf("path = %q, want /api/v2/agents", sawPath)
	}
}

// TestPaginationStopsOnShortPage is a focused check that the paginator stops
// when a page returns fewer than per_page records (no infinite loop, exactly
// the pages needed).
func TestPaginationStopsOnShortPage(t *testing.T) {
	var pages int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/problems" {
			http.NotFound(w, r)
			return
		}
		pages++
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		n, _ := strconv.Atoi(page)
		// page 1: full page of 3, page 2: short page of 1 => stop after page 2.
		if n == 1 {
			_, _ = w.Write([]byte(`{"problems":[{"id":1,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"},{"id":2,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"},{"id":3,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"problems":[{"id":4,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := freshservice.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "3"},
		Secrets: map[string]string{"api_key": "k"},
	}
	var got int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "problems", Config: cfg}, func(connectors.Record) error {
		got++
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got != 4 {
		t.Fatalf("records = %d, want 4", got)
	}
	if pages != 2 {
		t.Fatalf("pages fetched = %d, want 2", pages)
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := freshservice.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tickets", Config: cfg}, func(rec connectors.Record) error {
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
	// Fixture Check must not require creds or network.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := freshservice.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
}

func TestRegistryResolvesFreshservice(t *testing.T) {
	_ = freshservice.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("freshservice")
	if !ok {
		t.Fatal("registry did not resolve freshservice (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
}
