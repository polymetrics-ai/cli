package zendesksupport_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	zendesksupport "polymetrics.ai/internal/connectors/zendesk-support"
)

// TestReadPaginatesAndAuthenticatesAPIToken is the red-first test for the
// Zendesk Support connector. It asserts:
//   - API-token auth is sent as HTTP Basic (<email>/token : <api_token>),
//   - cursor pagination follows meta.after_cursor / page[after] across 2 pages
//     and stops when meta.has_more is false,
//   - records are extracted from the named array (tickets) and mapped.
func TestReadPaginatesAndAuthenticatesAPIToken(t *testing.T) {
	var sawAuth string
	var sawPageSize string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v2/tickets" {
			http.NotFound(w, r)
			return
		}
		sawPageSize = r.URL.Query().Get("page[size]")
		switch r.URL.Query().Get("page[after]") {
		case "":
			_, _ = w.Write([]byte(`{"tickets":[{"id":1,"subject":"a","status":"open","updated_at":"2026-01-01T00:00:00Z"},{"id":2,"subject":"b","status":"open","updated_at":"2026-01-02T00:00:00Z"}],"meta":{"has_more":true,"after_cursor":"CURSOR2"},"links":{"next":"x"}}`))
		case "CURSOR2":
			_, _ = w.Write([]byte(`{"tickets":[{"id":3,"subject":"c","status":"solved","updated_at":"2026-01-03T00:00:00Z"}],"meta":{"has_more":false,"after_cursor":null},"links":{"next":null}}`))
		default:
			t.Errorf("unexpected page[after]=%q", r.URL.Query().Get("page[after]"))
			_, _ = w.Write([]byte(`{"tickets":[],"meta":{"has_more":false}}`))
		}
	}))
	defer srv.Close()

	c := zendesksupport.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "subdomain": "acme"},
		Secrets: map[string]string{"credentials.api_token": "tok_123", "credentials.email": "agent@example.com"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tickets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("agent@example.com/token:tok_123"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if sawPageSize == "" {
		t.Fatalf("page[size] was not sent")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["subject"] == nil {
			t.Fatalf("record missing id/subject: %+v", rec)
		}
	}
}

// TestReadOAuthBearer asserts the OAuth2 access-token credential path sends a
// Bearer header instead of Basic.
func TestReadOAuthBearer(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"users":[{"id":7,"name":"Ada","email":"ada@example.com","role":"admin","updated_at":"2026-01-01T00:00:00Z"}],"meta":{"has_more":false}}`))
	}))
	defer srv.Close()

	c := zendesksupport.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "subdomain": "acme"},
		Secrets: map[string]string{"credentials.access_token": "oauth_abc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer oauth_abc" {
		t.Fatalf("Authorization = %q, want Bearer oauth_abc", sawAuth)
	}
	if len(got) != 1 || got[0]["email"] != "ada@example.com" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

// TestFixtureModeReadsWithoutNetwork ensures fixture mode emits deterministic
// records with no credentials and no network, so conformance works credential-free.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := zendesksupport.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"tickets", "users", "organizations", "groups", "satisfaction_ratings"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := zendesksupport.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := zendesksupport.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"tickets": false, "users": false, "organizations": false, "groups": false, "satisfaction_ratings": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := zendesksupport.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"credentials.access_token": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tickets", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme rejection, got %v", err)
	}
}

func TestRegisteredAndResolves(t *testing.T) {
	_ = zendesksupport.New() // ensure init ran
	c := zendesksupport.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("zendesk-support"); !ok {
		t.Fatal("registry did not resolve zendesk-support (self-registration)")
	}
}
