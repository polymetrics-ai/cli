package posthog_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/posthog"
)

func TestReadEventsPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/projects/42/events/" {
			http.NotFound(w, r)
			return
		}
		pages = append(pages, r.URL.Query().Get("page"))
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"results":[{"id":"evt_1","event":"signup","timestamp":"2026-01-01T00:00:00Z","distinct_id":"user_1"}],"next":"` + srvURL(r) + `/api/projects/42/events/?page=2"}`))
		case "2":
			_, _ = w.Write([]byte(`{"results":[{"id":"evt_2","event":"purchase","timestamp":"2026-01-02T00:00:00Z","distinct_id":"user_2"}],"next":null}`))
		default:
			t.Fatalf("unexpected page=%q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := posthog.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "project_id": "42", "start_date": "2026-01-01T00:00:00Z"}, Secrets: map[string]string{"api_key": "test-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer test-token" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if len(pages) != 2 || len(got) != 2 {
		t.Fatalf("pages=%v records=%d, want 2 pages/records", pages, len(got))
	}
	if got[0]["event"] != "signup" || got[1]["distinct_id"] != "user_2" {
		t.Fatalf("records not mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistryAndWrite(t *testing.T) {
	c := posthog.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var n int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(connectors.Record) error { n++; return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if n == 0 {
		t.Fatal("fixture Read emitted no records")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) < 2 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write err = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("posthog"); !ok {
		t.Fatal("registry did not resolve posthog")
	}
}

func srvURL(r *http.Request) string { return "http://" + r.Host }
