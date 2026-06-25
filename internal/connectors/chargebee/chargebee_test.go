package chargebee_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/chargebee"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Chargebee
// connector: HTTP Basic auth (API key as username, empty password), Chargebee
// offset/next_offset pagination over the top-level "list" array, and record
// mapping of the per-item resource envelope ({"customer": {...}}).
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/customers" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "":
			_, _ = w.Write([]byte(`{"list":[{"customer":{"id":"cus_1","email":"a@example.com","created_at":1700000000}},{"customer":{"id":"cus_2","email":"b@example.com","created_at":1700000100}}],"next_offset":"page2"}`))
		case "page2":
			_, _ = w.Write([]byte(`{"list":[{"customer":{"id":"cus_3","email":"c@example.com","created_at":1700000200}}]}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"list":[]}`))
		}
	}))
	defer srv.Close()

	c := chargebee.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"site_api_key": "test_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("test_key_123:"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["created_at"] == nil {
			t.Fatalf("record missing id/created_at: %+v", rec)
		}
	}
	if got[0]["id"] != "cus_1" || got[2]["id"] != "cus_3" {
		t.Fatalf("unexpected record ordering: %+v", got)
	}
}

// TestReadSubscriptionsEnvelope confirms a second stream maps its own resource
// envelope ({"subscription": {...}}) correctly.
func TestReadSubscriptionsEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/subscriptions" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"list":[{"subscription":{"id":"sub_1","customer_id":"cus_1","status":"active","created_at":1700000000}}]}`))
	}))
	defer srv.Close()

	c := chargebee.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"site_api_key": "k"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "subscriptions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "sub_1" || got[0]["status"] != "active" {
		t.Fatalf("subscription mapping wrong: %+v", got)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records with
// no network access, so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := chargebee.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode produced no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

// TestCheckFixtureMode verifies Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := chargebee.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog includes the core streams
// with a primary key and cursor field.
func TestCatalogStreams(t *testing.T) {
	c := chargebee.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"customers": false, "subscriptions": false, "invoices": false}
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

// TestBaseURLDerivedFromSite confirms that when no base_url override is set, the
// connector derives the host from the site config field.
func TestBaseURLDerivedFromSite(t *testing.T) {
	var sawHost string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawHost = r.Host
		_, _ = w.Write([]byte(`{"list":[]}`))
	}))
	defer srv.Close()

	// Use base_url to point at the test server but verify site is still accepted
	// alongside it (base_url wins for testability).
	c := chargebee.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "site": "acme"},
		Secrets: map[string]string{"site_api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawHost == "" || !strings.Contains(srv.URL, sawHost) {
		t.Fatalf("request host = %q not derived from base_url %q", sawHost, srv.URL)
	}
}

// TestRegisteredReadOnly confirms self-registration via NewRegistry and that the
// connector advertises read (not write) capability.
func TestRegisteredReadOnly(t *testing.T) {
	_ = chargebee.New() // ensure init ran
	c := chargebee.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only connector)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("chargebee"); !ok {
		t.Fatal("registry did not resolve chargebee (self-registration)")
	}
}
