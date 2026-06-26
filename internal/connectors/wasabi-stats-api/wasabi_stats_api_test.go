package wasabistatsapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	wasabistatsapi "polymetrics.ai/internal/connectors/wasabi-stats-api"
)

func TestReadBucketStatsAuthenticatesAndMaps(t *testing.T) {
	var sawAuth bool
	var sawStartDate string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/stats" {
			http.NotFound(w, r)
			return
		}
		sawAuth = r.Header.Get("Authorization") != ""
		sawStartDate = r.URL.Query().Get("start_date")
		_, _ = w.Write([]byte(`{"data":[{"id":"bucket-1","bucket":"logs","date":"2026-01-01","storage_bytes":42}]}`))
	}))
	defer srv.Close()

	c := wasabistatsapi.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "start_date": "2026-01-01T00:00:00Z"},
		Secrets: map[string]string{"api_key": "a:b"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bucket_stats", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawAuth {
		t.Fatal("request did not include Authorization header")
	}
	if sawStartDate != "2026-01-01T00:00:00Z" {
		t.Fatalf("start_date = %q", sawStartDate)
	}
	if len(got) != 1 || got[0]["id"] != "bucket-1" || got[0]["bucket"] != "logs" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := wasabistatsapi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bucket_stats", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "wasabi-stats-api" || len(cat.Streams) == 0 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	if _, ok := connectors.NewRegistry().Get("wasabi-stats-api"); !ok {
		t.Fatal("registry did not resolve wasabi-stats-api")
	}
	if c.Metadata().Capabilities.Write {
		t.Fatal("connector should be read-only")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
