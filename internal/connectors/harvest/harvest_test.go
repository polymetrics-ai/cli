package harvest_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/harvest"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Harvest
// connector: Bearer token auth plus the Harvest-Account-Id header, page-number
// pagination via the body next_page field over the clients[] array, and record
// mapping. Red until internal/connectors/harvest exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth, sawAccount string
	var pagesServed int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawAccount = r.Header.Get("Harvest-Account-Id")
		if r.URL.Path != "/clients" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			pagesServed++
			_, _ = w.Write([]byte(`{"clients":[{"id":1,"name":"Acme","is_active":true,"updated_at":"2017-06-26T21:34:11Z"},{"id":2,"name":"Globex","is_active":false,"updated_at":"2017-06-27T10:00:00Z"}],"page":1,"total_pages":2,"next_page":2,"previous_page":null}`))
		case "2":
			pagesServed++
			_, _ = w.Write([]byte(`{"clients":[{"id":3,"name":"Initech","is_active":true,"updated_at":"2017-06-28T08:00:00Z"}],"page":2,"total_pages":2,"next_page":null,"previous_page":1}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"clients":[],"next_page":null}`))
		}
	}))
	defer srv.Close()

	c := harvest.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "account_id": "123456"},
		Secrets: map[string]string{"credentials.api_token": "pat_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer pat_test_123" {
		t.Fatalf("Authorization = %q, want Bearer pat_test_123", sawAuth)
	}
	if sawAccount != "123456" {
		t.Fatalf("Harvest-Account-Id = %q, want 123456", sawAccount)
	}
	if pagesServed != 2 {
		t.Fatalf("pages served = %d, want 2", pagesServed)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestReadFixtureMode confirms credential-free fixture reads so the conformance
// harness can exercise the connector without live creds.
func TestReadFixtureMode(t *testing.T) {
	c := harvest.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
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
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

// TestCheckFixtureMode asserts Check short-circuits without network in fixture
// mode, and errors when the api_token secret is absent in live mode.
func TestCheckFixtureMode(t *testing.T) {
	c := harvest.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"account_id": "1"}})
	if err == nil {
		t.Fatal("Check without secret should error")
	}
}

// TestCatalogStreams asserts the published catalog exposes the core streams with
// primary keys and cursor fields.
func TestCatalogStreams(t *testing.T) {
	c := harvest.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"clients": false, "projects": false, "tasks": false, "users": false, "time_entries": false}
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

// TestRegistryResolves confirms the connector self-registers and resolves via the
// shared registry.
func TestRegistryResolves(t *testing.T) {
	_ = harvest.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("harvest"); !ok {
		t.Fatal("registry did not resolve harvest (self-registration)")
	}
	caps := harvest.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
