package sendgrid_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/sendgrid"
)

// TestReadListsPaginatesAndAuthenticates is the red-first test for the SendGrid
// connector: Bearer auth, marketing-API {result:[...]} extraction, and
// _metadata.next full-URL cursor pagination across two pages.
func TestReadListsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var calls int
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/marketing/lists" {
			http.NotFound(w, r)
			return
		}
		calls++
		switch r.URL.Query().Get("page_token") {
		case "":
			// First page advertises a full-URL next link.
			next := srv.URL + "/marketing/lists?page_token=tok2&page_size=100"
			_, _ = w.Write([]byte(`{"result":[{"id":"l1","name":"A","contact_count":5},{"id":"l2","name":"B","contact_count":6}],"_metadata":{"next":"` + next + `"}}`))
		case "tok2":
			_, _ = w.Write([]byte(`{"result":[{"id":"l3","name":"C","contact_count":7}],"_metadata":{}}`))
		default:
			t.Errorf("unexpected page_token=%q", r.URL.Query().Get("page_token"))
			_, _ = w.Write([]byte(`{"result":[],"_metadata":{}}`))
		}
	}))
	defer srv.Close()

	c := sendgrid.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "SG.test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer SG.test_123" {
		t.Fatalf("Authorization = %q, want Bearer SG.test_123", sawAuth)
	}
	if calls != 2 {
		t.Fatalf("server calls = %d, want 2 (two pages)", calls)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	if got[0]["id"] != "l1" || got[0]["name"] != "A" {
		t.Fatalf("record mapping wrong: %+v", got[0])
	}
}

// TestReadBouncesTopLevelArray verifies the suppression/bounces stream extracts
// records from a top-level JSON array and stops when a short page comes back.
func TestReadBouncesTopLevelArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/suppression/bounces" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			// Full page of 2 (default page size for the test config) -> more pages.
			_, _ = w.Write([]byte(`[{"email":"a@x.com","created":1700000000,"reason":"550","status":"5.1.1"},{"email":"b@x.com","created":1700000100,"reason":"550","status":"5.1.1"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"email":"c@x.com","created":1700000200,"reason":"550","status":"5.1.1"}]`))
		default:
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := sendgrid.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "SG.test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "suppression_bounces", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read bounces: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("bounce records = %d, want 3 across 2 pages", len(got))
	}
	if got[0]["email"] != "a@x.com" {
		t.Fatalf("bounce mapping wrong: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network access (credential-free conformance).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := sendgrid.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

// TestCheckFixtureAndCatalog verifies Check short-circuits in fixture mode and
// the catalog exposes the core streams.
func TestCheckFixtureAndCatalog(t *testing.T) {
	c := sendgrid.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"lists": true, "segments": true, "contacts": true, "suppression_bounces": true}
	seen := map[string]bool{}
	for _, s := range cat.Streams {
		seen[s.Name] = true
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs to bound SSRF risk.
func TestBaseURLValidation(t *testing.T) {
	c := sendgrid.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "SG.test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with file:// base_url err = %v, want base_url validation error", err)
	}
}

// TestRegistryResolvesSendgrid verifies self-registration via init().
func TestRegistryResolvesSendgrid(t *testing.T) {
	_ = sendgrid.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("sendgrid"); !ok {
		t.Fatal("registry did not resolve sendgrid (self-registration)")
	}
	caps := sendgrid.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
}
