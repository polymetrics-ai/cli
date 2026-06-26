package mixmax_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/mixmax"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Mixmax
// connector: X-API-Token auth, the Mixmax {results:[...],next,hasNext} cursor
// pagination over two pages, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawToken string
	var sawAuthHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.Header.Get("X-API-Token")
		sawAuthHeader = r.Header.Get("Authorization")
		if r.URL.Path != "/codesnippets" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("next") {
		case "":
			_, _ = w.Write([]byte(`{"results":[{"_id":"cs_1","title":"One","createdAt":"2026-01-01T00:00:00Z"},{"_id":"cs_2","title":"Two","createdAt":"2026-01-02T00:00:00Z"}],"next":"page2","hasNext":true}`))
		case "page2":
			_, _ = w.Write([]byte(`{"results":[{"_id":"cs_3","title":"Three","createdAt":"2026-01-03T00:00:00Z"}],"next":null,"hasNext":false}`))
		default:
			t.Errorf("unexpected next=%q", r.URL.Query().Get("next"))
			_, _ = w.Write([]byte(`{"results":[],"hasNext":false}`))
		}
	}))
	defer srv.Close()

	c := mixmax.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "codesnippets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "tok_test_123" {
		t.Fatalf("X-API-Token = %q, want tok_test_123", sawToken)
	}
	if sawAuthHeader != "" {
		t.Fatalf("Authorization header should be empty, got %q", sawAuthHeader)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["_id"] == nil {
			t.Fatalf("record missing _id: %+v", rec)
		}
	}
	if got[0]["title"] != "One" {
		t.Fatalf("first record title = %v, want One", got[0]["title"])
	}
}

// TestReadStopsWithoutNextToken ensures the loop terminates when the API omits a
// next token even if hasNext is missing.
func TestReadStopsWithoutNextToken(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"_id":"m_1","subject":"hi"}],"next":"","hasNext":false}`))
	}))
	defer srv.Close()

	c := mixmax.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "tok"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "messages", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (no next token must stop)", calls)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
}

// TestFixtureMode verifies credential-free conformance: fixture mode emits
// deterministic records with no network.
func TestFixtureMode(t *testing.T) {
	c := mixmax.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sequences", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["_id"] == nil {
			t.Fatalf("fixture record missing _id: %+v", rec)
		}
	}
	// Check also short-circuits in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCheckRequiresSecret(t *testing.T) {
	c := mixmax.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check with no api_key should fail")
	}
}

func TestBadBaseURLRejected(t *testing.T) {
	c := mixmax.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "codesnippets", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with non-http base_url should be rejected (SSRF guard)")
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := mixmax.New()
	md := c.Metadata()
	if md.Name != "mixmax" {
		t.Fatalf("metadata name = %q, want mixmax", md.Name)
	}
	if !md.Capabilities.Read || !md.Capabilities.Catalog || !md.Capabilities.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", md.Capabilities)
	}
	if md.Capabilities.Write {
		t.Fatalf("mixmax is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "mixmax" {
		t.Fatalf("catalog connector = %q, want mixmax", cat.Connector)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
}

func TestUnknownStreamRejected(t *testing.T) {
	c := mixmax.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"mode": "fixture"},
		Secrets: map[string]string{"api_key": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "does_not_exist", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("unknown stream should be rejected")
	}
}

func TestRegisteredInRegistry(t *testing.T) {
	_ = mixmax.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("mixmax"); !ok {
		t.Fatal("registry did not resolve mixmax (self-registration)")
	}
}
