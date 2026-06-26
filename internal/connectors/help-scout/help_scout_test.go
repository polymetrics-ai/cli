package helpscout_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	helpscout "polymetrics.ai/internal/connectors/help-scout"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Help Scout
// connector. It asserts:
//   - the OAuth2 client-credentials token exchange happens against the token
//     endpoint using the supplied client_id/client_secret,
//   - the resulting bearer token is sent on data requests,
//   - HAL+JSON page pagination walks all pages (page.totalPages),
//   - records are extracted from _embedded.<resource> and mapped.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawGrant, sawClientID, sawClientSecret string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_ = r.ParseForm()
			sawGrant = r.Form.Get("grant_type")
			sawClientID = r.Form.Get("client_id")
			sawClientSecret = r.Form.Get("client_secret")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"hs_token_abc","token_type":"bearer","expires_in":172800}`))
		case "/conversations":
			sawAuth = r.Header.Get("Authorization")
			switch r.URL.Query().Get("page") {
			case "", "1":
				_, _ = w.Write([]byte(`{"_embedded":{"conversations":[{"id":1,"number":101,"subject":"a","status":"active"},{"id":2,"number":102,"subject":"b","status":"closed"}]},"page":{"number":1,"size":25,"totalElements":3,"totalPages":2}}`))
			case "2":
				_, _ = w.Write([]byte(`{"_embedded":{"conversations":[{"id":3,"number":103,"subject":"c","status":"open"}]},"page":{"number":2,"size":25,"totalElements":3,"totalPages":2}}`))
			default:
				t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
				_, _ = w.Write([]byte(`{"_embedded":{"conversations":[]},"page":{"number":99,"size":25,"totalElements":3,"totalPages":2}}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := helpscout.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"token_url": srv.URL + "/oauth2/token",
		},
		Secrets: map[string]string{
			"client_id":     "app_id_123",
			"client_secret": "app_secret_456",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "conversations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if sawGrant != "client_credentials" {
		t.Fatalf("grant_type = %q, want client_credentials", sawGrant)
	}
	if sawClientID != "app_id_123" || sawClientSecret != "app_secret_456" {
		t.Fatalf("token exchange creds = %q/%q, want app_id_123/app_secret_456", sawClientID, sawClientSecret)
	}
	if sawAuth != "Bearer hs_token_abc" {
		t.Fatalf("Authorization = %q, want Bearer hs_token_abc", sawAuth)
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

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access and no credentials, so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := helpscout.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"conversations", "customers", "mailboxes", "users"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
		if got[0]["id"] == nil {
			t.Fatalf("fixture %s record missing id: %+v", stream, got[0])
		}
	}

	// Check must also short-circuit without creds in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams asserts the published catalog exposes the core streams with
// primary keys and cursor fields.
func TestCatalogStreams(t *testing.T) {
	c := helpscout.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"conversations": false, "customers": false, "mailboxes": false, "users": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs to bound SSRF.
func TestBaseURLValidation(t *testing.T) {
	c := helpscout.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"client_id": "x", "client_secret": "y"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "conversations", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for invalid base_url scheme")
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = helpscout.New() // ensure init ran
	c := helpscout.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("help-scout"); !ok {
		t.Fatal("registry did not resolve help-scout (self-registration)")
	}
}
