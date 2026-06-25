package auth0_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/auth0"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Auth0
// connector: Bearer auth on the Management API, page/per_page offset pagination
// over the resource-named array (users), include_totals envelope, and record
// mapping. Red until internal/connectors/auth0 exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawIncludeTotals string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawIncludeTotals = r.URL.Query().Get("include_totals")
		if r.URL.Path != "/api/v2/users" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "0":
			// per_page=2, total=3 -> a full page signals more.
			_, _ = w.Write([]byte(`{"start":0,"limit":2,"length":2,"total":3,"users":[{"user_id":"auth0|1","email":"a@example.com","created_at":"2026-01-01T00:00:00.000Z"},{"user_id":"auth0|2","email":"b@example.com","created_at":"2026-01-02T00:00:00.000Z"}]}`))
		case "1":
			_, _ = w.Write([]byte(`{"start":2,"limit":2,"length":1,"total":3,"users":[{"user_id":"auth0|3","email":"c@example.com","created_at":"2026-01-03T00:00:00.000Z"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"start":0,"limit":2,"total":3,"users":[]}`))
		}
	}))
	defer srv.Close()

	c := auth0.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"access_token": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", sawAuth)
	}
	if sawIncludeTotals != "true" {
		t.Fatalf("include_totals = %q, want true", sawIncludeTotals)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["user_id"] == nil || rec["email"] == nil {
			t.Fatalf("record missing user_id/email: %+v", rec)
		}
	}
}

// TestReadClientCredentialsAuth verifies that when only M2M client credentials
// are provided, the connector fetches a token from /oauth/token and uses it as a
// bearer on the resource request.
func TestReadClientCredentialsAuth(t *testing.T) {
	var sawResourceAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)
			if !strings.Contains(string(body), "grant_type=client_credentials") {
				t.Errorf("token request missing client_credentials grant: %s", body)
			}
			_, _ = w.Write([]byte(`{"access_token":"m2m_token","token_type":"Bearer","expires_in":86400}`))
		case "/api/v2/clients":
			sawResourceAuth = r.Header.Get("Authorization")
			_, _ = w.Write([]byte(`{"start":0,"limit":50,"length":1,"total":1,"clients":[{"client_id":"cid_1","name":"App"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := auth0.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url": srv.URL,
			"audience": srv.URL + "/api/v2/",
		},
		Secrets: map[string]string{
			"client_id":     "the_client_id",
			"client_secret": "the_client_secret",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawResourceAuth != "Bearer m2m_token" {
		t.Fatalf("resource Authorization = %q, want Bearer m2m_token", sawResourceAuth)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
}

// TestFixtureModeNeedsNoNetwork confirms credential-free conformance: fixture
// mode emits deterministic records without any HTTP call.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := auth0.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"users", "clients", "connections", "roles", "organizations"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) produced no records", stream)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := auth0.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "auth0" {
		t.Fatalf("catalog connector = %q, want auth0", cat.Connector)
	}
	want := map[string]bool{"users": true, "clients": true, "connections": true, "roles": true, "organizations": true}
	for _, s := range cat.Streams {
		delete(want, s.Name)
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	if len(want) != 0 {
		t.Fatalf("catalog missing streams: %v", want)
	}
}

func TestBaseURLValidationRejectsBadScheme(t *testing.T) {
	c := auth0.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"access_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected base_url scheme validation to reject file://")
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = auth0.New() // ensure init ran
	c := auth0.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("auth0 is a read-only source; Write should be false, got %+v", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("auth0"); !ok {
		t.Fatal("registry did not resolve auth0 (self-registration)")
	}
}
