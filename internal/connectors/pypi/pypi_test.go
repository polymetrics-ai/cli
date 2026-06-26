package pypi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/pypi"
)

func TestReadReleasesMapsPyPIJSON(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		if r.URL.Path != "/pypi/sampleproject/json" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"info":{"name":"sampleproject","version":"4.0.0","summary":"Sample"},"releases":{"3.0.0":[{"filename":"sample-3.tar.gz","upload_time_iso_8601":"2025-01-01T00:00:00Z","size":10}],"4.0.0":[{"filename":"sample-4.tar.gz","upload_time_iso_8601":"2026-01-01T00:00:00Z","size":20}]}}`))
	}))
	defer srv.Close()

	c := pypi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "project_name": "sampleproject"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "releases", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawPath != "/pypi/sampleproject/json" {
		t.Fatalf("path = %q", sawPath)
	}
	if len(got) != 2 || got[0]["project_name"] != "sampleproject" || got[0]["version"] == nil || got[0]["filename"] == nil {
		t.Fatalf("unexpected release records: %+v", got)
	}
}

func TestReadProjectVersionMapsInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pypi/sampleproject/4.0.0/json" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"info":{"name":"sampleproject","version":"4.0.0","summary":"Sample"},"urls":[{"filename":"sample-4.tar.gz"}]}`))
	}))
	defer srv.Close()

	c := pypi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "project_name": "sampleproject", "version": "4.0.0"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "project", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["name"] != "sampleproject" || got[0]["version"] != "4.0.0" {
		t.Fatalf("unexpected project record: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := pypi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "releases", Config: cfg}, func(rec connectors.Record) error {
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
	if cat.Connector != "pypi" || len(cat.Streams) < 2 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	for _, stream := range cat.Streams {
		if len(stream.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", stream.Name)
		}
	}
	if _, ok := connectors.NewRegistry().Get("pypi"); !ok {
		t.Fatal("registry did not resolve pypi")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
