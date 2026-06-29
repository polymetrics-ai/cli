package twitter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/twitter"
)

// TestReadTweetsPaginatesAndAuthenticates is the red-first test for the Twitter
// connector: App-only Bearer auth, Twitter v2 meta.next_token pagination over
// data[], and record mapping. Red until internal/connectors/twitter exists.
func TestReadTweetsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/tweets/search/recent" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("query") == "" {
			t.Errorf("missing query param")
		}
		switch r.URL.Query().Get("next_token") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"1","text":"hello","author_id":"a1","created_at":"2026-06-20T10:00:00.000Z"},{"id":"2","text":"world","author_id":"a2","created_at":"2026-06-20T11:00:00.000Z"}],"meta":{"next_token":"PAGE2","result_count":2}}`))
		case "PAGE2":
			_, _ = w.Write([]byte(`{"data":[{"id":"3","text":"again","author_id":"a3","created_at":"2026-06-20T12:00:00.000Z"}],"meta":{"result_count":1}}`))
		default:
			t.Errorf("unexpected next_token=%q", r.URL.Query().Get("next_token"))
			_, _ = w.Write([]byte(`{"data":[],"meta":{"result_count":0}}`))
		}
	}))
	defer srv.Close()

	c := twitter.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "query": "from:upstream"},
		Secrets: map[string]string{"api_key": "BEARER_TEST_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tweets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer BEARER_TEST_123" {
		t.Fatalf("Authorization = %q, want Bearer BEARER_TEST_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["text"] == nil {
			t.Fatalf("record missing id/text: %+v", rec)
		}
	}
}

// TestReadAuthorsFromIncludes verifies the authors stream is harvested from the
// includes.users[] expansion of the same recent-search endpoint.
func TestReadAuthorsFromIncludes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tweets/search/recent" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("next_token") != "" {
			_, _ = w.Write([]byte(`{"data":[],"meta":{"result_count":0}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"1","text":"hi","author_id":"a1"}],"includes":{"users":[{"id":"a1","name":"Ada","username":"ada"}]},"meta":{"result_count":1}}`))
	}))
	defer srv.Close()

	c := twitter.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "query": "from:ada"},
		Secrets: map[string]string{"api_key": "BEARER_TEST_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "authors", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("authors = %d, want 1", len(got))
	}
	if got[0]["username"] != "ada" || got[0]["id"] != "a1" {
		t.Fatalf("author record wrong: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access and no credentials, so conformance passes credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := twitter.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"tweets", "authors"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
		if got[0]["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", got[0])
		}
	}
	// Check must also short-circuit in fixture mode without creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestRegistryResolution confirms self-registration via init() and read-only
// capabilities.
func TestRegistryResolution(t *testing.T) {
	_ = twitter.New() // ensure init ran
	c := twitter.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("twitter is read-only, Write should be false: %+v", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("twitter"); !ok {
		t.Fatal("registry did not resolve twitter (self-registration)")
	}
}

// TestBaseURLValidation rejects non-http(s) base_url overrides (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := twitter.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "query": "x"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tweets", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected base_url scheme validation error")
	}
}
