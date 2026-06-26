package fastly_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/fastly"
)

// TestReadServicesPaginatesAndAuthenticates is the red-first test: it asserts the
// Fastly-Key auth header, page/per_page pagination across two pages of the
// top-level services array, and record mapping. Red until
// internal/connectors/fastly exists.
func TestReadServicesPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var pagesSeen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Fastly-Key")
		if r.URL.Path != "/service" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		pagesSeen = append(pagesSeen, page)
		switch page {
		case "1":
			// Full page (per_page=2) signals there may be more.
			_, _ = w.Write([]byte(`[{"id":"svc_1","name":"alpha","updated_at":"2026-01-01T00:00:00Z"},{"id":"svc_2","name":"beta","updated_at":"2026-01-02T00:00:00Z"}]`))
		case "2":
			// Short page (< per_page) signals the end.
			_, _ = w.Write([]byte(`[{"id":"svc_3","name":"gamma","updated_at":"2026-01-03T00:00:00Z"}]`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := fastly.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"fastly_api_token": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "services", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "tok_test_123" {
		t.Fatalf("Fastly-Key = %q, want tok_test_123", sawAuth)
	}
	if len(pagesSeen) != 2 {
		t.Fatalf("pages requested = %v, want 2 pages", pagesSeen)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestReadCurrentUserSingleObject verifies a single-object stream (no pagination)
// maps the root object into one record and sends the auth header.
func TestReadCurrentUserSingleObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/current_user" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Fastly-Key") == "" {
			t.Error("missing Fastly-Key header")
		}
		_, _ = w.Write([]byte(`{"id":"usr_1","login":"a@example.com","name":"Ada","role":"engineer","customer_id":"cus_1"}`))
	}))
	defer srv.Close()

	c := fastly.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"fastly_api_token": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "current_user", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["id"] != "usr_1" || got[0]["login"] != "a@example.com" {
		t.Fatalf("unexpected record: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records with
// no network access, so conformance runs without live credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := fastly.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"services", "current_user", "current_customer", "datacenters"} {
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
		for _, rec := range got {
			if len(rec) == 0 {
				t.Fatalf("fixture Read(%s) emitted empty record", stream)
			}
		}
	}
	// Check + Catalog must succeed in fixture mode.
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

// TestBaseURLValidation rejects non-http(s) and hostless overrides (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := fastly.New()
	bad := []string{"ftp://example.com", "://nohost", "file:///etc/passwd"}
	for _, b := range bad {
		cfg := connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": b},
			Secrets: map[string]string{"fastly_api_token": "tok"},
		}
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: "current_user", Config: cfg}, func(connectors.Record) error { return nil })
		if err == nil {
			t.Fatalf("Read with base_url=%q should fail validation", b)
		}
	}
}

// TestMissingSecret rejects a non-fixture read with no token.
func TestMissingSecret(t *testing.T) {
	c := fastly.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "current_user", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with no fastly_api_token should fail")
	}
}

// TestUnknownStream is rejected.
func TestUnknownStream(t *testing.T) {
	c := fastly.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "does_not_exist", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read of unknown stream should fail")
	}
}

// TestRegisteredReadOnly verifies self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = fastly.New() // ensure init ran
	c := fastly.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("fastly"); !ok {
		t.Fatal("registry did not resolve fastly (self-registration)")
	}
}
