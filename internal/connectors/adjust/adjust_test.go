package adjust_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/adjust"
)

func TestReadReportsPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/reports-service/report" {
			http.NotFound(w, r)
			return
		}
		pages = append(pages, r.URL.Query().Get("page"))
		if r.URL.Query().Get("dimensions") != "country" || r.URL.Query().Get("metrics") != "installs" {
			t.Fatalf("query dimensions=%q metrics=%q", r.URL.Query().Get("dimensions"), r.URL.Query().Get("metrics"))
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"rows":[{"dimensions":{"country":"US"},"metrics":{"installs":10}}],"next_page":2}`))
		case "2":
			_, _ = w.Write([]byte(`{"rows":[{"dimensions":{"country":"DE"},"metrics":{"installs":3}}]}`))
		default:
			t.Fatalf("unexpected page=%q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := adjust.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "dimensions": "country", "metrics": "installs", "ingest_start": "2026-01-01", "end_date": "2026-01-02"}, Secrets: map[string]string{"api_token": "test-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reports", Config: cfg}, func(rec connectors.Record) error {
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
	if got[0]["country"] != "US" || got[1]["installs"] != float64(3) {
		t.Fatalf("records not mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistryAndWrite(t *testing.T) {
	c := adjust.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var n int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reports", Config: cfg}, func(connectors.Record) error { n++; return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if n == 0 {
		t.Fatal("fixture Read emitted no records")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) == 0 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write err = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("adjust"); !ok {
		t.Fatal("registry did not resolve adjust")
	}
}
