package simplecast

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadPodcastsPaginatesAndAuthenticates(t *testing.T) {
	authFailures := 0
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			authFailures++
		}
		if r.URL.Path != "/podcasts" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("limit") != "2" {
			t.Fatalf("limit query = %q, want 2", r.URL.Query().Get("limit"))
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"collection":[{"id":"pod_1","title":"Metrics","updated_at":"2026-01-01T00:00:00Z"},{"id":"pod_2","title":"Agents","updated_at":"2026-01-02T00:00:00Z"}],"pages":{"next":{"href":"` + srv.URL + `/podcasts?page=2&limit=2"}}}`))
		case "2":
			_, _ = w.Write([]byte(`{"collection":[{"id":"pod_3","title":"Warehouse","updated_at":"2026-01-03T00:00:00Z"}],"pages":{"next":null}}`))
		default:
			t.Fatalf("unexpected page query %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "2"}, Secrets: map[string]string{"access_token": "test-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "podcasts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if authFailures != 0 {
		t.Fatal("bearer token was not applied to every request")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
	if got[0]["id"] != "pod_1" || got[0]["title"] != "Metrics" {
		t.Fatalf("first record = %+v", got[0])
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	for _, stream := range []string{"podcasts", "episodes"} {
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
	if cat.Connector != "simplecast" || len(cat.Streams) < 2 {
		t.Fatalf("catalog = %+v", cat)
	}
	got, ok := connectors.NewRegistry().Get("simplecast")
	if !ok {
		t.Fatal("registry did not resolve simplecast")
	}
	if caps := got.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v", caps)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want unsupported", err)
	}
}
