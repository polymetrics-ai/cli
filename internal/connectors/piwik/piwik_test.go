package piwik

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadPaginatesAuthenticatesAndMapsRecords(t *testing.T) {
	var sawToken, sawMethod string
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/index.php" {
			http.NotFound(w, r)
			return
		}
		sawToken = r.URL.Query().Get("token_auth")
		sawMethod = r.URL.Query().Get("method")
		if r.URL.Query().Get("filter_limit") != "2" {
			t.Fatalf("filter_limit = %q", r.URL.Query().Get("filter_limit"))
		}
		switch r.URL.Query().Get("filter_offset") {
		case "0":
			_, _ = w.Write([]byte(`[{"idVisit":"v1","visitorId":"abc","lastActionDateTime":"2026-01-01 00:00:00"},{"idVisit":"v2","visitorId":"def","lastActionDateTime":"2026-01-02 00:00:00"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"idVisit":"v3","visitorId":"ghi","lastActionDateTime":"2026-01-03 00:00:00"}]`))
		default:
			t.Fatalf("unexpected offset %q", r.URL.Query().Get("filter_offset"))
		}
	}))
	defer srv.Close()

	c := New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "site_id": "7", "page_size": "2"},
		Secrets: map[string]string{"token_auth": "unit-token"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "visits", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "unit-token" || sawMethod != "Live.getLastVisitsDetails" {
		t.Fatalf("token=%q method=%q", sawToken, sawMethod)
	}
	if requests != 2 || len(got) != 3 || got[0]["visit_id"] != "v1" || got[0]["last_action_at"] == nil {
		t.Fatalf("requests=%d records=%+v", requests, got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sites", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["site_id"] == nil {
		t.Fatalf("fixture records = %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) < 2 {
		t.Fatalf("Catalog = %+v, err=%v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("piwik"); !ok {
		t.Fatal("registry did not resolve piwik")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write error = %v", err)
	}
}
