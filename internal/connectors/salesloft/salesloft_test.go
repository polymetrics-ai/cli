package salesloft_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/salesloft"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Salesloft
// connector: Bearer auth, metadata.paging.next_page page-number pagination over
// data[], and record mapping. Red until internal/connectors/salesloft exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/people" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		sawPages = append(sawPages, page)
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case "", "1":
			_, _ = w.Write([]byte(`{"data":[{"id":1,"email_address":"a@example.com"},{"id":2,"email_address":"b@example.com"}],"metadata":{"paging":{"per_page":2,"current_page":1,"next_page":2,"total_pages":2}}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":3,"email_address":"c@example.com"}],"metadata":{"paging":{"per_page":2,"current_page":2,"next_page":null,"total_pages":2}}}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"data":[],"metadata":{"paging":{"next_page":null}}}`))
		}
	}))
	defer srv.Close()

	c := salesloft.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "people", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer key_test_123" {
		t.Fatalf("Authorization = %q, want Bearer key_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages); pages requested = %v", len(got), sawPages)
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["email_address"] == nil {
			t.Fatalf("record missing id/email_address: %+v", rec)
		}
	}
}

// TestReadFixtureMode confirms credential-free fixture reads work for conformance.
func TestReadFixtureMode(t *testing.T) {
	c := salesloft.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
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

// TestCheckFixtureMode confirms Check short-circuits without creds in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := salesloft.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published streams include the core set with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := salesloft.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "salesloft" {
		t.Fatalf("catalog connector = %q, want salesloft", cat.Connector)
	}
	want := map[string]bool{"people": false, "accounts": false, "cadences": false, "users": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Errorf("stream %q has no primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("catalog missing expected stream %q", name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = salesloft.New() // ensure init ran
	c := salesloft.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("salesloft"); !ok {
		t.Fatal("registry did not resolve salesloft (self-registration)")
	}
}

// TestOAuthRefreshToken confirms OAuth credentials trigger a token refresh and the
// resulting access token is used as the Bearer credential.
func TestOAuthRefreshToken(t *testing.T) {
	var tokenHits int
	var apiAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/oauth/token":
			tokenHits++
			_, _ = w.Write([]byte(`{"access_token":"fresh_access_token","token_type":"bearer","expires_in":3600}`))
		case "/users":
			apiAuth = r.Header.Get("Authorization")
			_, _ = w.Write([]byte(`{"data":[{"id":10,"email":"u@example.com"}],"metadata":{"paging":{"next_page":null}}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := salesloft.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"token_url": srv.URL + "/oauth/token",
		},
		Secrets: map[string]string{
			"client_id":     "cid",
			"client_secret": "csecret",
			"refresh_token": "rtoken",
		},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(oauth): %v", err)
	}
	if tokenHits == 0 {
		t.Fatal("expected token endpoint to be called for OAuth refresh")
	}
	if apiAuth != "Bearer fresh_access_token" {
		t.Fatalf("api Authorization = %q, want Bearer fresh_access_token", apiAuth)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
}
