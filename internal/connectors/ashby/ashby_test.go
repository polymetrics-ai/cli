package ashby_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/ashby"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Ashby
// connector: HTTP Basic auth (api key as username, empty password), Ashby's
// POST cursor-in-body pagination over results[], and record mapping across two
// pages. Red until internal/connectors/ashby exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawAccept = r.Header.Get("Accept")
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/candidate/list" {
			http.NotFound(w, r)
			return
		}
		var body map[string]any
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		cursor, _ := body["cursor"].(string)
		switch cursor {
		case "":
			_, _ = w.Write([]byte(`{"success":true,"results":[{"id":"cand_1","name":"Ada"},{"id":"cand_2","name":"Grace"}],"moreDataAvailable":true,"nextCursor":"PAGE2"}`))
		case "PAGE2":
			_, _ = w.Write([]byte(`{"success":true,"results":[{"id":"cand_3","name":"Katherine"}],"moreDataAvailable":false,"syncToken":"sync_abc"}`))
		default:
			t.Errorf("unexpected cursor=%q", cursor)
			_, _ = w.Write([]byte(`{"success":true,"results":[],"moreDataAvailable":false}`))
		}
	}))
	defer srv.Close()

	c := ashby.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "test_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "candidates", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	// "test_key_123:" base64-encoded => dGVzdF9rZXlfMTIzOg==
	if want := "Basic dGVzdF9rZXlfMTIzOg=="; sawAuth != want {
		t.Fatalf("Authorization = %q, want %q", sawAuth, want)
	}
	if sawAccept == "" {
		t.Fatalf("Accept header was not set")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork asserts the credential-free fixture path emits
// deterministic records without any network call, so conformance passes
// without live creds.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := ashby.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "jobs", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}

	// Check must also short-circuit in fixture mode (no creds).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog has the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := ashby.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q has no fields", s.Name)
		}
	}
}

// TestRegistryResolution asserts the connector self-registers and resolves via
// the shared registry, and is read-only (no write capability).
func TestRegistryResolution(t *testing.T) {
	_ = ashby.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("ashby")
	if !ok {
		t.Fatal("registry did not resolve ashby (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("ashby source must be read-only, got Write=true")
	}
}
