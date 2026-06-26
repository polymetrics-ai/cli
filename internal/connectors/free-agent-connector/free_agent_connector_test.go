package freeagentconnector_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	freeagentconnector "polymetrics.ai/internal/connectors/free-agent-connector"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it exercises the
// FreeAgent OAuth2 refresh-token exchange (Basic-auth token endpoint),
// Bearer auth on the data endpoints, page/per_page pagination across two pages,
// and contact record mapping (url primary key).
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawTokenAuth string
	var sawTokenForm string
	var sawDataAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/token_endpoint":
			sawTokenAuth = r.Header.Get("Authorization")
			_ = r.ParseForm()
			sawTokenForm = r.Form.Get("grant_type") + "/" + r.Form.Get("refresh_token")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"acc_tok_abc","token_type":"bearer","expires_in":3600}`))
		case "/v2/contacts":
			sawDataAuth = r.Header.Get("Authorization")
			switch r.URL.Query().Get("page") {
			case "", "1":
				_, _ = w.Write([]byte(`{"contacts":[{"url":"https://api/contacts/1","organisation_name":"Acme","updated_at":"2026-01-01T00:00:00Z"},{"url":"https://api/contacts/2","organisation_name":"Globex","updated_at":"2026-01-02T00:00:00Z"}]}`))
			case "2":
				_, _ = w.Write([]byte(`{"contacts":[{"url":"https://api/contacts/3","organisation_name":"Initech","updated_at":"2026-01-03T00:00:00Z"}]}`))
			default:
				_, _ = w.Write([]byte(`{"contacts":[]}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := freeagentconnector.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": srv.URL + "/v2", "page_size": "2"},
		Secrets: map[string]string{
			"client_id":              "cid",
			"client_secret":          "csecret",
			"client_refresh_token_2": "rtok",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawTokenAuth == "" || sawTokenAuth[:6] != "Basic " {
		t.Fatalf("token endpoint Authorization = %q, want Basic auth", sawTokenAuth)
	}
	if sawTokenForm != "refresh_token/rtok" {
		t.Fatalf("token form = %q, want refresh_token/rtok", sawTokenForm)
	}
	if sawDataAuth != "Bearer acc_tok_abc" {
		t.Fatalf("data Authorization = %q, want Bearer acc_tok_abc", sawDataAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["url"] == nil {
			t.Fatalf("record missing url primary key: %+v", rec)
		}
	}
}

// TestFixtureMode confirms credential-free deterministic reads (conformance).
func TestFixtureMode(t *testing.T) {
	c := freeagentconnector.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "invoices", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["url"] == nil {
		t.Fatalf("fixture record missing url: %+v", got[0])
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams asserts the published stream set and their keys.
func TestCatalogStreams(t *testing.T) {
	c := freeagentconnector.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"contacts": true, "invoices": true, "bills": true, "projects": true, "tasks": true}
	if len(cat.Streams) != len(want) {
		t.Fatalf("streams = %d, want %d", len(cat.Streams), len(want))
	}
	for _, s := range cat.Streams {
		if !want[s.Name] {
			t.Fatalf("unexpected stream %q", s.Name)
		}
		if len(s.PrimaryKey) != 1 || s.PrimaryKey[0] != "url" {
			t.Fatalf("stream %q primary key = %v, want [url]", s.Name, s.PrimaryKey)
		}
	}
}

// TestRegistryResolution asserts self-registration under the bare system name.
func TestRegistryResolution(t *testing.T) {
	_ = freeagentconnector.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("free-agent-connector"); !ok {
		t.Fatal("registry did not resolve free-agent-connector (self-registration)")
	}
}

// TestBaseURLValidation rejects non-http(s) overrides (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := freeagentconnector.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"client_id": "a", "client_secret": "b", "client_refresh_token_2": "c"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected base_url scheme validation error")
	}
}
