package devinai_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	devinai "polymetrics.ai/internal/connectors/devin-ai"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Devin AI
// connector: Bearer auth from api_token, Devin cursor pagination over items[]
// (has_next_page/end_cursor/after), the org_id in the path, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPaths = append(sawPaths, r.URL.Path)
		if r.URL.Path != "/v3/organizations/org-test/sessions" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("after") {
		case "":
			_, _ = w.Write([]byte(`{"items":[{"session_id":"devin-1","status":"finished","created_at":"2026-01-01T00:00:00Z"},{"session_id":"devin-2","status":"running","created_at":"2026-01-02T00:00:00Z"}],"has_next_page":true,"end_cursor":"cur2"}`))
		case "cur2":
			_, _ = w.Write([]byte(`{"items":[{"session_id":"devin-3","status":"blocked","created_at":"2026-01-03T00:00:00Z"}],"has_next_page":false,"end_cursor":null}`))
		default:
			t.Errorf("unexpected after=%q", r.URL.Query().Get("after"))
			_, _ = w.Write([]byte(`{"items":[],"has_next_page":false}`))
		}
	}))
	defer srv.Close()

	c := devinai.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "org_id": "org-test"},
		Secrets: map[string]string{"api_token": "cog_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sessions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer cog_test_123" {
		t.Fatalf("Authorization = %q, want Bearer cog_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages (paths=%v)", len(got), sawPaths)
	}
	for _, rec := range got {
		if rec["session_id"] == nil || rec["created_at"] == nil {
			t.Fatalf("record missing session_id/created_at: %+v", rec)
		}
	}
}

// TestFixtureModeReadsWithoutNetwork ensures fixture mode is deterministic and
// performs no network call, so conformance can run credential-free.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := devinai.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"sessions", "sessions_insights", "session_messages", "playbooks", "secrets"} {
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
}

func TestCheckFixtureMode(t *testing.T) {
	c := devinai.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := devinai.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com", "org_id": "org-test"},
		Secrets: map[string]string{"api_token": "cog_x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sessions", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with ftp base_url = %v, want base_url scheme error", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	c := devinai.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && Catalog && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("devin-ai"); !ok {
		t.Fatal("registry did not resolve devin-ai (self-registration)")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := devinai.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want at least 3", len(cat.Streams))
	}
	want := map[string]bool{"sessions": true, "sessions_insights": true, "session_messages": true}
	for _, s := range cat.Streams {
		delete(want, s.Name)
	}
	if len(want) != 0 {
		t.Fatalf("catalog missing core streams: %v", want)
	}
}
