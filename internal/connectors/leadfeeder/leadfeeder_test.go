package leadfeeder_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/leadfeeder"
)

// TestReadAccountsPaginatesAndAuthenticates is the red-first test: it asserts the
// Leadfeeder "Token token=" auth header, JSON:API links.next pagination across two
// pages of data[], and record mapping.
func TestReadAccountsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/accounts" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page[number]") {
		case "", "1":
			// First page: include a links.next that points back at page 2.
			_, _ = w.Write([]byte(`{
				"data":[
					{"id":"acc_1","type":"accounts","attributes":{"name":"Acme","industry":"Software","status":"active"}},
					{"id":"acc_2","type":"accounts","attributes":{"name":"Beta","industry":"Retail","status":"active"}}
				],
				"links":{"self":"` + srv.URL + `/accounts?page[number]=1","next":"` + srv.URL + `/accounts?page[number]=2"}
			}`))
		case "2":
			_, _ = w.Write([]byte(`{
				"data":[
					{"id":"acc_3","type":"accounts","attributes":{"name":"Gamma","industry":"Finance","status":"active"}}
				],
				"links":{"self":"` + srv.URL + `/accounts?page[number]=2","next":null}
			}`))
		default:
			t.Errorf("unexpected page[number]=%q", r.URL.Query().Get("page[number]"))
			_, _ = w.Write([]byte(`{"data":[],"links":{"next":null}}`))
		}
	}))
	defer srv.Close()

	c := leadfeeder.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Token token=tok_abc" {
		t.Fatalf("Authorization = %q, want Token token=tok_abc", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (across 2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
		if rec["name"] == nil {
			t.Fatalf("record missing flattened attribute name: %+v", rec)
		}
	}
	if got[0]["name"] != "Acme" {
		t.Fatalf("first record name = %v, want Acme", got[0]["name"])
	}
}

// TestReadLeadsUsesAccountPath asserts the nested leads endpoint uses account_id
// and forwards start_date/end_date as query params.
func TestReadLeadsUsesAccountPath(t *testing.T) {
	var sawPath, sawStart string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		sawStart = r.URL.Query().Get("start_date")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"lead_1","type":"leads","attributes":{"name":"Visitor Co","quality":5}}],"links":{"next":null}}`))
	}))
	defer srv.Close()

	c := leadfeeder.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"account_id": "acc_42",
			"start_date": "2026-01-01T00:00:00Z",
		},
		Secrets: map[string]string{"api_token": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "leads", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read leads: %v", err)
	}
	if sawPath != "/accounts/acc_42/leads" {
		t.Fatalf("leads path = %q, want /accounts/acc_42/leads", sawPath)
	}
	if sawStart != "2026-01-01" {
		t.Fatalf("start_date query = %q, want 2026-01-01", sawStart)
	}
	if len(got) != 1 || got[0]["id"] != "lead_1" {
		t.Fatalf("leads = %+v, want one lead_1", got)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records with
// no network access so conformance runs without credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := leadfeeder.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"accounts", "leads", "visits", "custom_feeds"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read %s: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read %s emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}

	// Check should also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := leadfeeder.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "leadfeeder" {
		t.Fatalf("catalog connector = %q, want leadfeeder", cat.Connector)
	}
	want := map[string]bool{"accounts": false, "leads": false, "visits": false, "custom_feeds": false}
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
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestRegistryResolvesLeadfeeder(t *testing.T) {
	_ = leadfeeder.New() // ensure init() ran
	r := connectors.NewRegistry()
	got, ok := r.Get("leadfeeder")
	if !ok {
		t.Fatal("registry did not resolve leadfeeder (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("leadfeeder is read-only; Write should be false, got %+v", caps)
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := leadfeeder.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_token": "tok_abc"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with bad base_url scheme should fail with base_url error, got %v", err)
	}
}
