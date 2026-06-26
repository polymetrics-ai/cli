package googleclassroom_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	googleclassroom "polymetrics.ai/internal/connectors/google-classroom"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Google
// Classroom connector: it exchanges the OAuth refresh token for an access token,
// sends it as a Bearer header, follows pageToken/nextPageToken pagination over
// the courses[] array across two pages, and maps records. Red until
// internal/connectors/google-classroom exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var tokenExchanges int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/token":
			// OAuth2 refresh-token grant.
			tokenExchanges++
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if got := r.PostFormValue("grant_type"); got != "refresh_token" {
				t.Errorf("grant_type = %q, want refresh_token", got)
			}
			if got := r.PostFormValue("refresh_token"); got != "rt_123" {
				t.Errorf("refresh_token = %q, want rt_123", got)
			}
			if got := r.PostFormValue("client_id"); got != "cid_123" {
				t.Errorf("client_id = %q, want cid_123", got)
			}
			if got := r.PostFormValue("client_secret"); got != "csecret_123" {
				t.Errorf("client_secret = %q, want csecret_123", got)
			}
			_, _ = w.Write([]byte(`{"access_token":"at_abc","token_type":"Bearer","expires_in":3600}`))
		case r.URL.Path == "/v1/courses":
			sawAuth = r.Header.Get("Authorization")
			switch r.URL.Query().Get("pageToken") {
			case "":
				_, _ = w.Write([]byte(`{"courses":[{"id":"c1","name":"Math","updateTime":"2026-01-01T00:00:00Z"},{"id":"c2","name":"Science","updateTime":"2026-01-02T00:00:00Z"}],"nextPageToken":"PAGE2"}`))
			case "PAGE2":
				_, _ = w.Write([]byte(`{"courses":[{"id":"c3","name":"History","updateTime":"2026-01-03T00:00:00Z"}]}`))
			default:
				t.Errorf("unexpected pageToken=%q", r.URL.Query().Get("pageToken"))
				_, _ = w.Write([]byte(`{"courses":[]}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := googleclassroom.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"token_url": srv.URL + "/token",
		},
		Secrets: map[string]string{
			"client_id":            "cid_123",
			"client_secret":        "csecret_123",
			"client_refresh_token": "rt_123",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "courses", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer at_abc" {
		t.Fatalf("Authorization = %q, want Bearer at_abc", sawAuth)
	}
	if tokenExchanges == 0 {
		t.Fatal("expected at least one token exchange")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["id"] != "c1" || got[0]["name"] != "Math" {
		t.Fatalf("record[0] = %+v, want id=c1 name=Math", got[0])
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["updateTime"] == nil {
			t.Fatalf("record missing id/updateTime: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork verifies the credential-free fixture path used by the
// conformance harness emits deterministic records without any network call.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := googleclassroom.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"courses", "teachers", "students", "courseWork", "announcements"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("Read(%s) fixture emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("Read(%s) fixture record missing id: %+v", stream, rec)
			}
		}
	}
}

// TestCheckFixtureMode confirms Check succeeds without creds in fixture mode and
// fails when refresh-token secrets are missing in live mode.
func TestCheckFixtureMode(t *testing.T) {
	c := googleclassroom.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	err := c.Check(context.Background(), connectors.RuntimeConfig{})
	if err == nil {
		t.Fatal("Check without secrets should fail")
	}
}

// TestBaseURLSSRFGuard rejects non-http(s) base_url overrides.
func TestBaseURLSSRFGuard(t *testing.T) {
	c := googleclassroom.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"client_id": "a", "client_secret": "b", "client_refresh_token": "c"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "courses", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme error, got %v", err)
	}
}

// TestCatalogStreams verifies the published catalog exposes the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := googleclassroom.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"courses": false, "teachers": false, "students": false, "courseWork": false, "announcements": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
			if len(s.PrimaryKey) == 0 {
				t.Fatalf("stream %q missing primary key", s.Name)
			}
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = googleclassroom.New() // ensure init ran
	c := googleclassroom.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only source)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("google-classroom"); !ok {
		t.Fatal("registry did not resolve google-classroom (self-registration)")
	}
}
