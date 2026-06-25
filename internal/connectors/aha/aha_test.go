package aha_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/aha"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Aha! connector:
// Bearer auth, Aha! page-number pagination (page/per_page with a pagination object
// carrying total_pages/current_page), the resource-keyed envelope, and record
// mapping. Aha! lists features at GET <base>/api/v1/features and returns
// {"features":[...],"pagination":{"total_records":..,"total_pages":..,"current_page":..}}.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var pagesSeen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v1/features" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		pagesSeen = append(pagesSeen, page)
		switch page {
		case "", "1":
			_, _ = w.Write([]byte(`{"features":[` +
				`{"id":"1","reference_num":"PROJ-1","name":"Feature one","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z"},` +
				`{"id":"2","reference_num":"PROJ-2","name":"Feature two","created_at":"2026-01-03T00:00:00Z","updated_at":"2026-01-04T00:00:00Z"}` +
				`],"pagination":{"total_records":3,"total_pages":2,"current_page":1}}`))
		case "2":
			_, _ = w.Write([]byte(`{"features":[` +
				`{"id":"3","reference_num":"PROJ-3","name":"Feature three","created_at":"2026-01-05T00:00:00Z","updated_at":"2026-01-06T00:00:00Z"}` +
				`],"pagination":{"total_records":3,"total_pages":2,"current_page":2}}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"features":[],"pagination":{"total_records":3,"total_pages":2,"current_page":3}}`))
		}
	}))
	defer srv.Close()

	c := aha.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "per_page": "2"},
		Secrets: map[string]string{"api_key": "aha_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "features", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer aha_test_key" {
		t.Fatalf("Authorization = %q, want Bearer aha_test_key", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages); pages seen = %v", len(got), pagesSeen)
	}
	// Ensure we actually requested a second page (pagination drove the loop).
	requestedPage2 := false
	for _, p := range pagesSeen {
		if p == "2" {
			requestedPage2 = true
		}
	}
	if !requestedPage2 {
		t.Fatalf("expected a second page request, pages seen = %v", pagesSeen)
	}
	// Record mapping: every record carries its id and reference_num.
	for _, rec := range got {
		if rec["id"] == nil || rec["reference_num"] == nil {
			t.Fatalf("record missing id/reference_num: %+v", rec)
		}
	}
	if got[0]["name"] != "Feature one" {
		t.Fatalf("first record name = %v, want Feature one", got[0]["name"])
	}
}

// TestFixtureModeNoNetwork confirms the credential-free fixture path emits
// deterministic records without any network call, so conformance passes without
// live creds.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := aha.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "features", Config: cfg}, func(rec connectors.Record) error {
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
	// Check must also short-circuit in fixture mode (no creds required).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams with
// primary keys and cursor fields.
func TestCatalogStreams(t *testing.T) {
	c := aha.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	byName := map[string]connectors.Stream{}
	for _, s := range cat.Streams {
		byName[s.Name] = s
	}
	for _, name := range []string{"features", "products", "ideas"} {
		s, ok := byName[name]
		if !ok {
			t.Fatalf("catalog missing stream %q", name)
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", name)
		}
	}
}

// TestRegistryResolvesAha confirms self-registration via init() and that the
// connector is read-only (Write disabled).
func TestRegistryResolvesAha(t *testing.T) {
	_ = aha.New() // ensure init ran
	c := aha.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("aha is read-only; Write should be false, got %+v", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("aha"); !ok {
		t.Fatal("registry did not resolve aha (self-registration)")
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF validation on base_url overrides.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := aha.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "features", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme rejection, got %v", err)
	}
}
