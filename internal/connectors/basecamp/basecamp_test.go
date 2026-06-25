package basecamp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/basecamp"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Basecamp
// connector: OAuth2 refresh-token exchange, Bearer auth on API requests, RFC5988
// Link-header pagination across two pages, and record mapping. The token endpoint
// and the API are both served by the same httptest server (the connector resolves
// token_url and base_url from config), so no live credentials are needed.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIAuth string
	var sawTokenForm string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/authorization/token":
			_ = r.ParseForm()
			sawTokenForm = r.Form.Encode()
			_, _ = w.Write([]byte(`{"access_token":"access_abc","expires_in":1209600}`))
			return
		case r.URL.Path == "/9999999/projects.json":
			sawAPIAuth = r.Header.Get("Authorization")
			if r.URL.Query().Get("page") == "2" {
				// Final page: no Link header signals the end of pagination.
				_, _ = w.Write([]byte(`[{"id":3,"name":"Gamma","status":"archived"}]`))
				return
			}
			// First page links to page 2 via the Link header (absolute URL on
			// this same server, which the connector must follow verbatim).
			next := "<" + baseURLOf(r) + "/9999999/projects.json?page=2>; rel=\"next\""
			w.Header().Set("Link", next)
			_, _ = w.Write([]byte(`[{"id":1,"name":"Alpha","status":"active"},{"id":2,"name":"Beta","status":"active"}]`))
			return
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := basecamp.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"token_url":  srv.URL + "/authorization/token",
			"account_id": "9999999",
		},
		Secrets: map[string]string{
			"client_id":              "cid",
			"client_secret":          "csecret",
			"client_refresh_token_2": "rtok",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIAuth != "Bearer access_abc" {
		t.Fatalf("API Authorization = %q, want Bearer access_abc", sawAPIAuth)
	}
	// The token request must carry the Basecamp refresh-grant form.
	if !strings.Contains(sawTokenForm, "type=refresh") || !strings.Contains(sawTokenForm, "refresh_token=rtok") {
		t.Fatalf("token form = %q, want type=refresh & refresh_token=rtok", sawTokenForm)
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

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records with
// no network access (conformance runs credential-free).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := basecamp.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"projects", "people", "events"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) returned no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture record missing id: %+v", rec)
			}
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestRegisteredReadOnly asserts registry self-registration and read-only caps.
func TestRegisteredReadOnly(t *testing.T) {
	_ = basecamp.New() // ensure init ran
	c := basecamp.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("basecamp"); !ok {
		t.Fatal("registry did not resolve basecamp (self-registration)")
	}
}

// TestCatalogStreams asserts the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := basecamp.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"projects": false, "people": false, "events": false}
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

// baseURLOf reconstructs the absolute base URL of the test server from a request.
func baseURLOf(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
