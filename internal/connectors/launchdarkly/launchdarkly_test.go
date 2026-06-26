package launchdarkly_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/launchdarkly"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the LaunchDarkly
// connector: the raw access token is sent in the Authorization header (no Bearer
// prefix), offset/limit pagination walks two pages of items[], and each record is
// mapped with its primary key. Red until internal/connectors/launchdarkly exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/projects" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			// page_size is 2 here, so a full page signals there is another page.
			_, _ = w.Write([]byte(`{"items":[{"_id":"p1","key":"alpha","name":"Alpha"},{"_id":"p2","key":"beta","name":"Beta"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"items":[{"_id":"p3","key":"gamma","name":"Gamma"}]}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"items":[]}`))
		}
	}))
	defer srv.Close()

	c := launchdarkly.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"access_token": "api-12345"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	// LaunchDarkly uses the raw token in Authorization, no "Bearer " prefix.
	if sawAuth != "api-12345" {
		t.Fatalf("Authorization = %q, want api-12345 (no Bearer prefix)", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["_id"] == nil || rec["key"] == nil {
			t.Fatalf("record missing _id/key: %+v", rec)
		}
	}
}

// TestReadProjectScopedStream covers a stream whose path embeds the project_key
// config (flags), confirming the path is templated correctly.
func TestReadProjectScopedStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/flags/my-project" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"items":[{"key":"flag-1","name":"Flag One"}]}`))
	}))
	defer srv.Close()

	c := launchdarkly.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "project_key": "my-project"},
		Secrets: map[string]string{"access_token": "api-12345"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "flags", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read flags: %v", err)
	}
	if len(got) != 1 || got[0]["key"] != "flag-1" {
		t.Fatalf("flags records = %+v, want 1 with key flag-1", got)
	}
}

// TestFixtureMode confirms the credential-free fixture path emits deterministic
// records so the conformance harness can run without live secrets.
func TestFixtureMode(t *testing.T) {
	c := launchdarkly.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"projects", "members", "auditlog", "flags", "environments"} {
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
	}
	// Check must also short-circuit in fixture mode (no creds, no network).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogAndMetadata verifies the published catalog and read-only metadata.
func TestCatalogAndMetadata(t *testing.T) {
	c := launchdarkly.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog has %d streams, want >= 3", len(cat.Streams))
	}
}

// TestRegistryResolution confirms the connector self-registers via init() and is
// resolvable through the shared registry.
func TestRegistryResolution(t *testing.T) {
	_ = launchdarkly.New() // ensure the package init() ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("launchdarkly"); !ok {
		t.Fatal("registry did not resolve launchdarkly (self-registration)")
	}
}
