package dixa_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/dixa"
)

// TestReadWindowsAndAuthenticates is the red-first test for the Dixa connector.
// Dixa's conversation_export endpoint returns a top-level JSON array and is
// "paginated" by advancing a [updated_after, updated_before) date window by
// batch_size days. This test asserts Bearer auth, that the connector walks more
// than one date window (pagination across 2 windows), and that records map.
func TestReadWindowsAndAuthenticates(t *testing.T) {
	var sawAuth string
	var windows int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/conversation_export" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		if q.Get("updated_after") == "" || q.Get("updated_before") == "" {
			t.Errorf("missing window params: %v", q)
		}
		windows++
		// First window returns two conversations, later windows return empty.
		if windows == 1 {
			_, _ = w.Write([]byte(`[{"id":1,"updated_at":1700000000000,"created_at":1699990000000,"status":"Open","initial_channel":"Email"},{"id":2,"updated_at":1700000100000,"created_at":1699990100000,"status":"Closed","initial_channel":"Chat"}]`))
			return
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := dixa.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"start_date": "2023-11-01",
			"end_date":   "2023-11-30",
			"batch_size": "15",
		},
		Secrets: map[string]string{"api_token": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "conversations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_test_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_test_123", sawAuth)
	}
	if windows < 2 {
		t.Fatalf("date windows requested = %d, want >= 2 (pagination)", windows)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["updated_at"] == nil {
			t.Fatalf("record missing id/updated_at: %+v", rec)
		}
		if rec["status"] == nil {
			t.Fatalf("record missing status: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork confirms the credential-free fixture path emits
// deterministic records without any HTTP server.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := dixa.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "conversations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil || got[0]["updated_at"] == nil {
		t.Fatalf("fixture record missing id/updated_at: %+v", got[0])
	}
}

// TestCheckFixtureNoNetwork ensures Check short-circuits in fixture mode.
func TestCheckFixtureNoNetwork(t *testing.T) {
	c := dixa.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestRegisteredReadOnly verifies capabilities and self-registration.
func TestRegisteredReadOnly(t *testing.T) {
	c := dixa.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("dixa"); !ok {
		t.Fatal("registry did not resolve dixa (self-registration)")
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := dixa.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) == 0 {
		t.Fatal("catalog has no streams")
	}
	var hasConversations bool
	for _, s := range cat.Streams {
		if s.Name == "conversations" {
			hasConversations = true
			if len(s.PrimaryKey) == 0 || s.PrimaryKey[0] != "id" {
				t.Fatalf("conversations primary key = %v, want [id]", s.PrimaryKey)
			}
			if len(s.CursorFields) == 0 || s.CursorFields[0] != "updated_at" {
				t.Fatalf("conversations cursor = %v, want [updated_at]", s.CursorFields)
			}
		}
	}
	if !hasConversations {
		t.Fatal("catalog missing conversations stream")
	}
}
