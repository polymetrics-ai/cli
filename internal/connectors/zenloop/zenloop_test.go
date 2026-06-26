package zenloop_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/zenloop"
)

func TestReadAnswersPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v1/answers" {
			http.NotFound(w, r)
			return
		}
		pages = append(pages, r.URL.Query().Get("page"))
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"data":[{"id":"a1","score":9,"comment":"great","inserted_at":"2026-01-01T00:00:00Z","survey":{"id":"s1"}}],"meta":{"next_page":2}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"a2","score":4,"comment":"ok","inserted_at":"2026-01-02T00:00:00Z","survey":{"id":"s1"}}],"meta":{"next_page":null}}`))
		default:
			t.Fatalf("unexpected page=%q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := zenloop.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/v1", "survey_id": "s1", "page_size": "1"}, Secrets: map[string]string{"api_token": "test-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "answers", Config: cfg}, func(rec connectors.Record) error {
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
		t.Fatalf("pages=%v records=%d, want two pages", pages, len(got))
	}
	if got[0]["id"] != "a1" || got[0]["survey_id"] != "s1" || got[1]["score"] != float64(4) {
		t.Fatalf("records not mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistryAndWrite(t *testing.T) {
	c := zenloop.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var n int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "answers", Config: cfg}, func(connectors.Record) error { n++; return nil }); err != nil {
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
	if _, ok := connectors.NewRegistry().Get("zenloop"); !ok {
		t.Fatal("registry did not resolve zenloop")
	}
}
