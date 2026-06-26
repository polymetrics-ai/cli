package workable_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/workable"
)

func TestReadJobsUsesBearerAndMapsJobs(t *testing.T) {
	var sawAuth bool
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		if r.URL.Path != "/jobs" {
			http.NotFound(w, r)
			return
		}
		sawAuth = r.Header.Get("Authorization") != ""
		_, _ = w.Write([]byte(`{"jobs":[{"id":"job_1","shortcode":"ENG","title":"Engineer"}]}`))
	}))
	defer srv.Close()

	c := workable.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "account_subdomain": "example", "start_date": "20260101T000000Z"},
		Secrets: map[string]string{"api_key": "x"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "jobs", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawPath != "/jobs" || !sawAuth {
		t.Fatalf("path/auth = %q/%v", sawPath, sawAuth)
	}
	if len(got) != 1 || got[0]["id"] != "job_1" || got[0]["title"] != "Engineer" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := workable.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "jobs", Config: cfg}, func(rec connectors.Record) error {
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
	if cat.Connector != "workable" || len(cat.Streams) == 0 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	if _, ok := connectors.NewRegistry().Get("workable"); !ok {
		t.Fatal("registry did not resolve workable")
	}
	if c.Metadata().Capabilities.Write {
		t.Fatal("connector should be read-only")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
