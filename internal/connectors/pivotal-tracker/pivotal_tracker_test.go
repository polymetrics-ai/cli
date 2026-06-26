package pivotaltracker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadPaginatesAuthenticatesAndMapsRecords(t *testing.T) {
	var sawToken, sawLimit string
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		sawToken = r.Header.Get("X-TrackerToken")
		if r.URL.Path != "/services/v5/projects/42/stories" {
			http.NotFound(w, r)
			return
		}
		sawLimit = r.URL.Query().Get("limit")
		switch r.URL.Query().Get("offset") {
		case "0":
			_, _ = w.Write([]byte(`[{"id":1,"name":"One","current_state":"started","updated_at":"2026-01-01T00:00:00Z"},{"id":2,"name":"Two","current_state":"finished","updated_at":"2026-01-02T00:00:00Z"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"id":3,"name":"Three","current_state":"accepted","updated_at":"2026-01-03T00:00:00Z"}]`))
		default:
			t.Fatalf("unexpected offset %q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	c := New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/services/v5", "project_id": "42", "page_size": "2"},
		Secrets: map[string]string{"api_token": "unit-token"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "stories", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "unit-token" {
		t.Fatalf("X-TrackerToken = %q", sawToken)
	}
	if sawLimit != "2" || requests != 2 {
		t.Fatalf("limit=%q requests=%d, want limit 2 and 2 requests", sawLimit, requests)
	}
	if len(got) != 3 || got[0]["id"] == nil || got[0]["state"] != "started" {
		t.Fatalf("mapped records = %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) < 3 {
		t.Fatalf("Catalog = %+v, err=%v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("pivotal-tracker"); !ok {
		t.Fatal("registry did not resolve pivotal-tracker")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write error = %v", err)
	}
}
