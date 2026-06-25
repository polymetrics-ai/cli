package blogger_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/blogger"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Blogger
// connector: it exchanges a refresh token for an access token at the token
// endpoint, applies the resulting Bearer token to API requests, follows
// pageToken/nextPageToken pagination across two pages, and maps records.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		sawAPIAuth   string
		sawGrant     string
		sawClientID  string
		sawRefresh   string
		tokenCalls   int
		listCalls    int
		lastPageSize string
	)

	mux := http.NewServeMux()
	// Google OAuth2 token endpoint (refresh_token grant).
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		tokenCalls++
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse token form: %v", err)
		}
		sawGrant = r.PostForm.Get("grant_type")
		sawClientID = r.PostForm.Get("client_id")
		sawRefresh = r.PostForm.Get("refresh_token")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "ya29.fresh_access_token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})
	// Blogger posts list endpoint with pageToken pagination.
	mux.HandleFunc("/blogs/42/posts", func(w http.ResponseWriter, r *http.Request) {
		listCalls++
		sawAPIAuth = r.Header.Get("Authorization")
		lastPageSize = r.URL.Query().Get("maxResults")
		switch r.URL.Query().Get("pageToken") {
		case "":
			_, _ = w.Write([]byte(`{"kind":"blogger#postList","nextPageToken":"PAGE2","items":[` +
				`{"id":"p1","title":"First","published":"2026-01-01T00:00:00Z","updated":"2026-01-02T00:00:00Z","url":"https://b/1","author":{"id":"a1","displayName":"Ada"}},` +
				`{"id":"p2","title":"Second","published":"2026-01-03T00:00:00Z","updated":"2026-01-04T00:00:00Z","url":"https://b/2","author":{"id":"a2","displayName":"Grace"}}` +
				`]}`))
		case "PAGE2":
			_, _ = w.Write([]byte(`{"kind":"blogger#postList","items":[` +
				`{"id":"p3","title":"Third","published":"2026-01-05T00:00:00Z","updated":"2026-01-06T00:00:00Z","url":"https://b/3","author":{"id":"a3","displayName":"Katherine"}}` +
				`]}`))
		default:
			t.Errorf("unexpected pageToken=%q", r.URL.Query().Get("pageToken"))
			_, _ = w.Write([]byte(`{"items":[]}`))
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := blogger.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"token_url": srv.URL + "/token",
			"blog_id":   "42",
		},
		Secrets: map[string]string{
			"client_id":            "cid-123",
			"client_secret":        "csecret-456",
			"client_refresh_token": "refresh-789",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "posts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if tokenCalls == 0 {
		t.Fatal("expected at least one token-endpoint call to exchange the refresh token")
	}
	if sawGrant != "refresh_token" {
		t.Fatalf("grant_type = %q, want refresh_token", sawGrant)
	}
	if sawClientID != "cid-123" {
		t.Fatalf("client_id = %q, want cid-123", sawClientID)
	}
	if sawRefresh != "refresh-789" {
		t.Fatalf("refresh_token = %q, want refresh-789", sawRefresh)
	}
	if sawAPIAuth != "Bearer ya29.fresh_access_token" {
		t.Fatalf("API Authorization = %q, want Bearer ya29.fresh_access_token", sawAPIAuth)
	}
	if listCalls != 2 {
		t.Fatalf("list calls = %d, want 2 (pagination across 2 pages)", listCalls)
	}
	if lastPageSize == "" {
		t.Fatal("expected maxResults page size query param to be set")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["title"] == nil {
			t.Fatalf("record missing id/title: %+v", rec)
		}
	}
	// Author flattening: nested author.displayName must surface as author_display_name.
	if got[0]["author_display_name"] != "Ada" {
		t.Fatalf("author_display_name = %v, want Ada", got[0]["author_display_name"])
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network access (conformance runs credential-free).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := blogger.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"blogs", "posts", "pages", "comments"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}

	// Check must succeed in fixture mode with no network/creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := blogger.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "blogger" {
		t.Fatalf("catalog connector = %q, want blogger", cat.Connector)
	}
	want := map[string]bool{"blogs": false, "posts": false, "pages": false, "comments": false}
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
			t.Fatalf("catalog missing expected stream %q", name)
		}
	}
}

func TestReadOnlyCapabilities(t *testing.T) {
	c := blogger.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (Blogger is read-only here)", caps)
	}
}

func TestRegisteredInRegistry(t *testing.T) {
	_ = blogger.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("blogger"); !ok {
		t.Fatal("registry did not resolve blogger (self-registration)")
	}
}

func TestBaseURLValidation(t *testing.T) {
	c := blogger.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example/v3", "blog_id": "1"},
		Secrets: map[string]string{"client_id": "a", "client_secret": "b", "client_refresh_token": "c"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "posts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme rejection, got %v", err)
	}
	_ = url.Values{} // keep net/url imported for parity with template
}
