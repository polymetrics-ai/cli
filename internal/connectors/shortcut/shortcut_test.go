package shortcut

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadStoriesPaginatesAndAuthenticates(t *testing.T) {
	authFailures := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Shortcut-Token") != "test-token" {
			authFailures++
		}
		if r.URL.Path != "/api/v3/stories" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("limit") != "2" {
			t.Fatalf("limit query = %q, want 2", r.URL.Query().Get("limit"))
		}
		switch r.URL.Query().Get("next") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"story_1","name":"Import","updated_at":"2026-01-01T00:00:00Z"},{"id":"story_2","name":"Map","updated_at":"2026-01-02T00:00:00Z"}],"next":"cursor-2"}`))
		case "cursor-2":
			_, _ = w.Write([]byte(`{"data":[{"id":"story_3","name":"Export","updated_at":"2026-01-03T00:00:00Z"}],"next":null}`))
		default:
			t.Fatalf("unexpected next query %q", r.URL.Query().Get("next"))
		}
	}))
	defer srv.Close()

	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "2"}, Secrets: map[string]string{"api_token": "test-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "stories", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if authFailures != 0 {
		t.Fatal("shortcut token header was not applied to every request")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
	if got[0]["id"] != "story_1" || got[0]["name"] != "Import" {
		t.Fatalf("first record = %+v", got[0])
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	for _, stream := range []string{"stories", "epics", "projects", "iterations"} {
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		}); err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 || got[0]["id"] == nil {
			t.Fatalf("fixture records for %s = %+v", stream, got)
		}
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "shortcut" || len(cat.Streams) < 4 {
		t.Fatalf("catalog = %+v", cat)
	}
	got, ok := connectors.NewRegistry().Get("shortcut")
	if !ok {
		t.Fatal("registry did not resolve shortcut")
	}
	if caps := got.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v", caps)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want unsupported", err)
	}
}
