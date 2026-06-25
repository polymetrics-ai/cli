package jira_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/jira"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Jira
// connector: HTTP Basic auth (email:api_token), Jira startAt/maxResults/total
// offset pagination over issues[], and record mapping. Two pages of issues are
// served; the connector must stop when startAt+len >= total.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/rest/api/3/search" {
			http.NotFound(w, r)
			return
		}
		startAt, _ := strconv.Atoi(r.URL.Query().Get("startAt"))
		switch startAt {
		case 0:
			_, _ = w.Write([]byte(`{"startAt":0,"maxResults":2,"total":3,"issues":[
				{"id":"10001","key":"PROJ-1","self":"https://x/10001","fields":{"summary":"first","status":{"name":"Open"},"updated":"2026-01-01T00:00:00.000+0000","created":"2026-01-01T00:00:00.000+0000"}},
				{"id":"10002","key":"PROJ-2","self":"https://x/10002","fields":{"summary":"second","status":{"name":"Done"},"updated":"2026-01-02T00:00:00.000+0000","created":"2026-01-02T00:00:00.000+0000"}}
			]}`))
		case 2:
			_, _ = w.Write([]byte(`{"startAt":2,"maxResults":2,"total":3,"issues":[
				{"id":"10003","key":"PROJ-3","self":"https://x/10003","fields":{"summary":"third","status":{"name":"Open"},"updated":"2026-01-03T00:00:00.000+0000","created":"2026-01-03T00:00:00.000+0000"}}
			]}`))
		default:
			t.Errorf("unexpected startAt=%d", startAt)
			_, _ = w.Write([]byte(`{"startAt":0,"maxResults":2,"total":0,"issues":[]}`))
		}
	}))
	defer srv.Close()

	c := jira.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"domain":    "example.atlassian.net",
			"email":     "user@example.com",
			"page_size": "2",
		},
		Secrets: map[string]string{"credentials.api_token": "tok_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "issues", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user@example.com:tok_123"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	// Record mapping: top-level id/key plus flattened summary from fields.
	if got[0]["id"] != "10001" || got[0]["key"] != "PROJ-1" {
		t.Fatalf("first record id/key wrong: %+v", got[0])
	}
	if got[0]["summary"] != "first" {
		t.Fatalf("first record summary not flattened from fields: %+v", got[0])
	}
	if got[0]["status"] != "Open" {
		t.Fatalf("first record status not flattened: %+v", got[0])
	}
	if got[2]["key"] != "PROJ-3" {
		t.Fatalf("third record key wrong: %+v", got[2])
	}
}

// TestReadProjectsPaginatesValues exercises the values[]-shaped project/search
// endpoint with the same offset pagination.
func TestReadProjectsPaginatesValues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/project/search" {
			http.NotFound(w, r)
			return
		}
		startAt, _ := strconv.Atoi(r.URL.Query().Get("startAt"))
		switch startAt {
		case 0:
			_, _ = w.Write([]byte(`{"startAt":0,"maxResults":1,"total":2,"values":[
				{"id":"1","key":"PROJ","name":"Project One","projectTypeKey":"software"}
			]}`))
		case 1:
			_, _ = w.Write([]byte(`{"startAt":1,"maxResults":1,"total":2,"values":[
				{"id":"2","key":"OTHER","name":"Project Two","projectTypeKey":"business"}
			]}`))
		default:
			t.Errorf("unexpected startAt=%d", startAt)
		}
	}))
	defer srv.Close()

	c := jira.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"email":     "user@example.com",
			"page_size": "1",
		},
		Secrets: map[string]string{"credentials.api_token": "tok_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read projects: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("projects = %d, want 2", len(got))
	}
	if got[0]["key"] != "PROJ" || got[1]["key"] != "OTHER" {
		t.Fatalf("project keys wrong: %+v", got)
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := jira.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"issues", "projects", "users"} {
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
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := jira.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("jira is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"issues": false, "projects": false, "users": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

func TestRegistryResolvesJira(t *testing.T) {
	_ = jira.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("jira"); !ok {
		t.Fatal("registry did not resolve jira (self-registration)")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := jira.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com", "email": "u@example.com"},
		Secrets: map[string]string{"credentials.api_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "issues", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http(s) base_url scheme")
	}
}
