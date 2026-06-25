package instagram_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/instagram"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Instagram
// connector: Bearer access_token auth, Facebook Graph API paging.next
// pagination over data[], and record mapping. Two pages are served; the next
// page is requested via the absolute paging.next URL the first page returns.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if !strings.HasSuffix(r.URL.Path, "/media") {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("after") {
		case "":
			// First page returns a paging.next absolute URL pointing back here.
			next := srv.URL + r.URL.Path + "?after=CURSOR2"
			_, _ = w.Write([]byte(`{"data":[` +
				`{"id":"media_1","caption":"hello","media_type":"IMAGE","timestamp":"2026-01-01T00:00:00+0000"},` +
				`{"id":"media_2","caption":"world","media_type":"VIDEO","timestamp":"2026-01-02T00:00:00+0000"}` +
				`],"paging":{"cursors":{"after":"CURSOR2"},"next":"` + next + `"}}`))
		case "CURSOR2":
			_, _ = w.Write([]byte(`{"data":[` +
				`{"id":"media_3","caption":"third","media_type":"IMAGE","timestamp":"2026-01-03T00:00:00+0000"}` +
				`],"paging":{"cursors":{"after":"CURSOR3"}}}`))
		default:
			t.Errorf("unexpected after=%q", r.URL.Query().Get("after"))
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := instagram.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "ig_user_id": "17841400000000000"},
		Secrets: map[string]string{"access_token": "IGQVJtoken123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "media", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer IGQVJtoken123" {
		t.Fatalf("Authorization = %q, want Bearer IGQVJtoken123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["media_type"] == nil {
			t.Fatalf("record missing id/media_type: %+v", rec)
		}
	}
	if got[0]["caption"] != "hello" {
		t.Fatalf("first record caption = %v, want hello", got[0]["caption"])
	}
}

// TestUsersStreamSingleObject verifies the users stream, which the Graph API
// returns as a single object (not an array), maps to one record.
func TestUsersStreamSingleObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/17841400000000000") {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"id":"17841400000000000","username":"acme","followers_count":1200,"media_count":42}`))
	}))
	defer srv.Close()

	c := instagram.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "ig_user_id": "17841400000000000"},
		Secrets: map[string]string{"access_token": "tok"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read users: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("users records = %d, want 1", len(got))
	}
	if got[0]["username"] != "acme" {
		t.Fatalf("username = %v, want acme", got[0]["username"])
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network access so conformance runs credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := instagram.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"users", "media", "stories", "user_insights"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted 0 records", stream)
		}
		if got[0]["id"] == nil {
			t.Fatalf("fixture %s record missing id: %+v", stream, got[0])
		}
	}
	// Check + Catalog in fixture mode must not error.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
}

// TestBaseURLSSRFValidation rejects non-http(s) base_url overrides.
func TestBaseURLSSRFValidation(t *testing.T) {
	c := instagram.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"access_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "media", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}

// TestRegistryResolution confirms self-registration and metadata.
func TestRegistryResolution(t *testing.T) {
	_ = instagram.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("instagram")
	if !ok {
		t.Fatal("registry did not resolve instagram (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("instagram is read-only; Write should be false, got %+v", caps)
	}
}
