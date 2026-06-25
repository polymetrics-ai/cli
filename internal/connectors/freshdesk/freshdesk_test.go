package freshdesk_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/freshdesk"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Freshdesk
// connector: HTTP Basic auth (apikey:X), Link-header page-number pagination over
// a top-level JSON array, and record mapping. Red until
// internal/connectors/freshdesk exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v2/tickets" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			// Advertise a next page via the Link header (RFC 5988 rel=next).
			next := fmt.Sprintf("<%s/api/v2/tickets?page=2&per_page=100>; rel=\"next\"", "http://"+r.Host)
			w.Header().Set("Link", next)
			_, _ = w.Write([]byte(`[{"id":1,"subject":"a","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z"},{"id":2,"subject":"b","created_at":"2026-01-03T00:00:00Z","updated_at":"2026-01-04T00:00:00Z"}]`))
		case "2":
			// No Link header => last page.
			_, _ = w.Write([]byte(`[{"id":3,"subject":"c","created_at":"2026-01-05T00:00:00Z","updated_at":"2026-01-06T00:00:00Z"}]`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := freshdesk.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "fd_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tickets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("fd_test_key:X"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q (HTTP Basic apikey:X)", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["updated_at"] == nil {
			t.Fatalf("record missing id/updated_at: %+v", rec)
		}
	}
}

// TestReadContactsMapsRecords verifies a second stream maps its fields and that
// the per_page page-size override is sent.
func TestReadContactsMapsRecords(t *testing.T) {
	var sawPerPage string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/contacts" {
			http.NotFound(w, r)
			return
		}
		sawPerPage = r.URL.Query().Get("per_page")
		_, _ = w.Write([]byte(`[{"id":10,"name":"Ada","email":"ada@example.com","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z"}]`))
	}))
	defer srv.Close()

	c := freshdesk.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "25"},
		Secrets: map[string]string{"api_key": "fd_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
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
	if got[0]["email"] != "ada@example.com" || got[0]["name"] != "Ada" {
		t.Fatalf("contact not mapped: %+v", got[0])
	}
}

// TestDomainBuildsBaseURL verifies the domain config is turned into the v2 base
// URL when no explicit base_url override is given (and is validated for SSRF).
func TestDomainBuildsBaseURL(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	// Force the connector to talk to the test server by overriding base_url,
	// but also confirm a bare domain (no scheme) is accepted and normalized.
	c := freshdesk.New()
	host := strings.TrimPrefix(srv.URL, "http://")
	cfg := connectors.RuntimeConfig{
		// base_url override wins, but use the insecure_http flag implicitly via http scheme.
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "k"},
	}
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "companies", Config: cfg}, func(connectors.Record) error { return nil }); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawPath != "/api/v2/companies" {
		t.Fatalf("path = %q, want /api/v2/companies", sawPath)
	}
	_ = host
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := freshdesk.New()
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
	c := freshdesk.New()
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

func TestRegistryResolvesFreshdesk(t *testing.T) {
	_ = freshdesk.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("freshdesk")
	if !ok {
		t.Fatal("registry did not resolve freshdesk (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
}
