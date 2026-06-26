package mailgun_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/mailgun"
)

// TestReadDomainsPaginatesAndAuthenticates is the red-first test: Mailgun uses
// HTTP Basic auth (username "api", password = private_key), the v3 domains list
// returns {"items":[...],"total_count":N} and is paged with skip/limit (offset).
// Asserts auth header, offset pagination across 2 pages over items[], and mapping.
func TestReadDomainsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v3/domains" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("skip") {
		case "", "0":
			// First full page (page size 2) -> there is more.
			_, _ = w.Write([]byte(`{"total_count":3,"items":[{"id":"d1","name":"a.example.com","state":"active","created_at":"Wed, 01 Jan 2020 00:00:00 GMT"},{"id":"d2","name":"b.example.com","state":"active","created_at":"Wed, 01 Jan 2020 00:00:00 GMT"}]}`))
		case "2":
			// Short page -> stop.
			_, _ = w.Write([]byte(`{"total_count":3,"items":[{"id":"d3","name":"c.example.com","state":"unverified","created_at":"Wed, 01 Jan 2020 00:00:00 GMT"}]}`))
		default:
			t.Errorf("unexpected skip=%q", r.URL.Query().Get("skip"))
			_, _ = w.Write([]byte(`{"total_count":3,"items":[]}`))
		}
	}))
	defer srv.Close()

	c := mailgun.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"private_key": "key-secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "domains", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("api:key-secret"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
	if got[0]["name"] != "a.example.com" {
		t.Fatalf("first domain name = %v, want a.example.com", got[0]["name"])
	}
}

// TestReadEventsPagingNext exercises the paging.next cursor pagination used by
// the events stream (and other v3 sub-resources). The response carries an
// absolute next URL in paging.next; an empty page or a next that returns no
// items stops the loop.
func TestReadEventsPagingNext(t *testing.T) {
	var base string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/example.com/events" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("page") == "next" {
			// Second (final) page: items present but next points to an empty page.
			_, _ = w.Write([]byte(`{"items":[{"id":"e2","event":"delivered","timestamp":1577836801}],"paging":{"next":"` + base + `/v3/example.com/events?page=empty"}}`))
			return
		}
		if r.URL.Query().Get("page") == "empty" {
			_, _ = w.Write([]byte(`{"items":[],"paging":{"next":"` + base + `/v3/example.com/events?page=empty"}}`))
			return
		}
		// First page.
		_, _ = w.Write([]byte(`{"items":[{"id":"e1","event":"accepted","timestamp":1577836800}],"paging":{"next":"` + base + `/v3/example.com/events?page=next"}}`))
	}))
	defer srv.Close()
	base = srv.URL

	c := mailgun.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "domain_name": "example.com"},
		Secrets: map[string]string{"private_key": "key-secret"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read events: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("event records = %d, want 2 (2 pages then empty)", len(got))
	}
	if got[0]["id"] != "e1" || got[1]["id"] != "e2" {
		t.Fatalf("event ids = %v, %v want e1, e2", got[0]["id"], got[1]["id"])
	}
}

// TestFixtureModeNoNetwork confirms the credential-free fixture path emits
// deterministic records without any network call.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := mailgun.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "domains", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture Read emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check in fixture mode must not error and must not need creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams asserts the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := mailgun.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"domains": false, "events": false, "mailing_lists": false, "tags": false}
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

// TestRegistryResolves confirms self-registration resolves via NewRegistry.
func TestRegistryResolves(t *testing.T) {
	_ = mailgun.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("mailgun")
	if !ok {
		t.Fatal("registry did not resolve mailgun (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
}
