package mailerlite_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/mailerlite"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the MailerLite
// connector: Bearer auth on every request, cursor pagination across two pages
// (data[] + meta.next_cursor), and record mapping. Red until the package exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawAccept = r.Header.Get("Accept")
		if r.URL.Path != "/subscribers" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"sub_1","email":"a@example.com","status":"active"},{"id":"sub_2","email":"b@example.com","status":"active"}],"meta":{"next_cursor":"CURSOR_2"}}`))
		case "CURSOR_2":
			_, _ = w.Write([]byte(`{"data":[{"id":"sub_3","email":"c@example.com","status":"unsubscribed"}],"meta":{"next_cursor":null}}`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"data":[],"meta":{"next_cursor":null}}`))
		}
	}))
	defer srv.Close()

	c := mailerlite.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "subscribers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_test_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_test_123", sawAuth)
	}
	if sawAccept != "application/json" {
		t.Fatalf("Accept = %q, want application/json", sawAccept)
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

// TestReadCampaignsMapping checks a second stream maps its core fields.
func TestReadCampaignsMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/campaigns" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"cmp_1","name":"Newsletter","type":"regular","status":"sent","created_at":"2026-01-01T00:00:00Z"}],"meta":{"next_cursor":null}}`))
	}))
	defer srv.Close()

	c := mailerlite.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_token": "tok_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	rec := got[0]
	if rec["id"] != "cmp_1" || rec["name"] != "Newsletter" || rec["status"] != "sent" {
		t.Fatalf("campaign record mapped wrong: %+v", rec)
	}
}

// TestFixtureMode confirms the credential-free fixture path emits records.
func TestFixtureMode(t *testing.T) {
	c := mailerlite.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "subscribers", Config: cfg}, func(rec connectors.Record) error {
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
}

// TestCheckFixtureNoNetwork ensures Check passes in fixture mode without creds.
func TestCheckFixtureNoNetwork(t *testing.T) {
	c := mailerlite.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams checks the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := mailerlite.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"subscribers": false, "campaigns": false, "groups": false, "segments": false, "automations": false}
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
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegistryResolves confirms self-registration via the connectors registry.
func TestRegistryResolves(t *testing.T) {
	_ = mailerlite.New() // ensure init() ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("mailerlite"); !ok {
		t.Fatal("registry did not resolve mailerlite (self-registration)")
	}
}

// TestBaseURLValidation rejects SSRF-risky base URLs.
func TestBaseURLValidation(t *testing.T) {
	c := mailerlite.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil"},
		Secrets: map[string]string{"api_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "subscribers", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http base_url scheme")
	}
}
