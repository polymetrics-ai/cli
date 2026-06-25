package uptick_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/uptick"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Uptick
// connector. It exercises:
//   - the OAuth2 password-grant token exchange against {base_url}/api/oauth2/token/,
//   - the resulting Bearer access token applied to data requests,
//   - links.next full-URL cursor pagination across two pages,
//   - record mapping for the clients stream (id + updated present).
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var tokenForm url.Values
	var sawAuth string
	var dataRequests int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/oauth2/token/":
			_ = r.ParseForm()
			tokenForm = r.PostForm
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"tok_abc","token_type":"Bearer","expires_in":3600}`))
		case r.URL.Path == "/api/v2.14/clients/":
			sawAuth = r.Header.Get("Authorization")
			dataRequests++
			// First page returns a links.next pointing at page 2; second page ends.
			if r.URL.Query().Get("page") == "2" {
				_, _ = w.Write([]byte(`{"data":[{"id":3,"updated":"2026-01-03T00:00:00.000000Z","name":"Gamma"}],"links":{"next":null}}`))
				return
			}
			next := "http://" + r.Host + "/api/v2.14/clients/?page=2"
			resp := `{"data":[{"id":1,"updated":"2026-01-01T00:00:00.000000Z","name":"Alpha"},{"id":2,"updated":"2026-01-02T00:00:00.000000Z","name":"Beta"}],"links":{"next":"` + next + `"}}`
			_, _ = w.Write([]byte(resp))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := uptick.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": srv.URL, "username": "ops@example.com"},
		Secrets: map[string]string{
			"client_id":     "cid",
			"client_secret": "csecret",
			"password":      "pw",
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

	if sawAuth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", sawAuth)
	}
	if dataRequests != 2 {
		t.Fatalf("data requests = %d, want 2 (links.next pagination)", dataRequests)
	}
	if tokenForm.Get("grant_type") != "password" {
		t.Fatalf("token grant_type = %q, want password", tokenForm.Get("grant_type"))
	}
	if tokenForm.Get("username") != "ops@example.com" || tokenForm.Get("password") != "pw" {
		t.Fatalf("token form missing username/password creds: %v", tokenForm)
	}
	if tokenForm.Get("client_id") != "cid" || tokenForm.Get("client_secret") != "csecret" {
		t.Fatalf("token form missing client creds: %v", tokenForm)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["updated"] == nil {
			t.Fatalf("record missing id/updated: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork verifies that mode=fixture emits deterministic records
// without any network access so conformance can run without live credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := uptick.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "properties", Config: cfg}, func(rec connectors.Record) error {
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
		if rec["id"] == nil || rec["updated"] == nil {
			t.Fatalf("fixture record missing id/updated: %+v", rec)
		}
	}

	// Fixture Check must succeed without a network call too.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams with
// id primary key and updated cursor.
func TestCatalogStreams(t *testing.T) {
	c := uptick.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"tasks": false, "clients": false, "properties": false, "invoices": false, "assets": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
			if len(s.PrimaryKey) == 0 || s.PrimaryKey[0] != "id" {
				t.Fatalf("stream %s primary key = %v, want [id]", s.Name, s.PrimaryKey)
			}
			if len(s.CursorFields) == 0 || s.CursorFields[0] != "updated" {
				t.Fatalf("stream %s cursor = %v, want [updated]", s.Name, s.CursorFields)
			}
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs to bound SSRF.
func TestBaseURLValidation(t *testing.T) {
	c := uptick.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "username": "u"},
		Secrets: map[string]string{"client_id": "c", "client_secret": "s", "password": "p"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with bad base_url err = %v, want base_url scheme error", err)
	}
}

// TestRegisteredReadOnly checks self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = uptick.New() // ensure init ran
	c := uptick.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("uptick is read-only, Write should be false: %+v", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("uptick"); !ok {
		t.Fatal("registry did not resolve uptick (self-registration)")
	}
}

// guard against accidental JSON tag drift in case mappers are changed.
var _ = json.Marshal
