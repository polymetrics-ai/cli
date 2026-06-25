package sentry_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/sentry"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts Bearer
// auth, Sentry Link-header cursor pagination over two pages (rel="next";
// results="true" then results="false"), and record mapping for the issues
// stream. Red until internal/connectors/sentry exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPaths = append(sawPaths, r.URL.Path)
		if r.URL.Path != "/api/0/projects/acme/backend/issues/" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			// First page: signal a real next page.
			w.Header().Set("Link", "<"+baseOf(r)+"/api/0/projects/acme/backend/issues/?cursor=0:100:0>; rel=\"next\"; results=\"true\"")
			_, _ = w.Write([]byte(`[{"id":"1","title":"NPE","status":"unresolved"},{"id":"2","title":"Timeout","status":"resolved"}]`))
		case "0:100:0":
			// Second page: next link present but results="false" -> stop.
			w.Header().Set("Link", "<"+baseOf(r)+"/api/0/projects/acme/backend/issues/?cursor=0:200:0>; rel=\"next\"; results=\"false\"")
			_, _ = w.Write([]byte(`[{"id":"3","title":"OOM","status":"unresolved"}]`))
		default:
			t.Errorf("unexpected cursor=%q", r.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := sentry.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "organization": "acme", "project": "backend"},
		Secrets: map[string]string{"auth_token": "tok_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "issues", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages); paths=%v", len(got), sawPaths)
	}
	if len(sawPaths) != 2 {
		t.Fatalf("requests = %d, want exactly 2 pages; paths=%v", len(sawPaths), sawPaths)
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["title"] == nil {
			t.Fatalf("record missing id/title: %+v", rec)
		}
	}
}

// TestReadProjectsStream exercises the org/project-independent projects endpoint
// and its record mapping.
func TestReadProjectsStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/0/projects/" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"10","slug":"backend","name":"Backend","platform":"python"}]`))
	}))
	defer srv.Close()

	c := sentry.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "organization": "acme", "project": "backend"},
		Secrets: map[string]string{"auth_token": "tok_abc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read projects: %v", err)
	}
	if len(got) != 1 || got[0]["slug"] != "backend" {
		t.Fatalf("projects records = %+v, want one backend project", got)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// with no live credentials and no network access (conformance path).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := sentry.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	// Each stream's identity field: releases are keyed by version, the rest by id.
	idField := map[string]string{"projects": "id", "issues": "id", "events": "id", "releases": "version"}
	for stream, key := range idField {
		var n int
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			if rec[key] == nil {
				t.Fatalf("fixture %s record missing %s: %+v", stream, key, rec)
			}
			n++
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read %s: %v", stream, err)
		}
		if n == 0 {
			t.Fatalf("fixture %s emitted no records", stream)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCheckRejectsBadBaseURL ensures the SSRF guard rejects non-http(s) schemes.
func TestCheckRejectsBadBaseURL(t *testing.T) {
	c := sentry.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "organization": "acme", "project": "backend"},
		Secrets: map[string]string{"auth_token": "tok_abc"},
	}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check should reject non-http(s) base_url")
	}
}

// TestRegisteredReadOnly verifies self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = sentry.New() // ensure init ran
	caps := sentry.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("sentry"); !ok {
		t.Fatal("registry did not resolve sentry (self-registration)")
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	cat, err := sentry.New().Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"projects": false, "issues": false, "events": false, "releases": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// baseOf reconstructs the test server base URL for embedding in Link headers.
func baseOf(r *http.Request) string {
	host := r.Host
	if host == "" {
		host = "127.0.0.1"
	}
	return strings.TrimSuffix("http://"+host, "/")
}
