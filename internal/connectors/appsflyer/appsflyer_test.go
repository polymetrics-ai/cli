package appsflyer_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/appsflyer"
)

func TestReadInstallsAuthenticatesAndMapsCSV(t *testing.T) {
	var sawAuth string
	var sawFrom string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/raw-data/export/app/com.example/installs_report/v5" {
			http.NotFound(w, r)
			return
		}
		sawFrom = r.URL.Query().Get("from")
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("AppsFlyer ID,Event Time,Media Source,Campaign\naf_1,2026-01-01 00:00:00,network_a,winter\naf_2,2026-01-02 00:00:00,network_b,spring\n"))
	}))
	defer srv.Close()

	c := appsflyer.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "app_id": "com.example", "start_date": "2026-01-01", "end_date": "2026-01-02"}, Secrets: map[string]string{"api_token": "test-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "installs_report", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer test-token" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if sawFrom != "2026-01-01" {
		t.Fatalf("from = %q, want 2026-01-01", sawFrom)
	}
	if len(got) != 2 || got[0]["appsflyer_id"] != "af_1" || got[1]["campaign"] != "spring" {
		t.Fatalf("csv records not mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistryAndWrite(t *testing.T) {
	c := appsflyer.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var n int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "installs_report", Config: cfg}, func(connectors.Record) error { n++; return nil }); err != nil {
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
	if _, ok := connectors.NewRegistry().Get("appsflyer"); !ok {
		t.Fatal("registry did not resolve appsflyer")
	}
}
