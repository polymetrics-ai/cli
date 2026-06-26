package spacexapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	spacexapi "polymetrics.ai/internal/connectors/spacex-api"
)

func TestReadLaunchesUsesPublicAPI(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		if r.URL.Path != "/launches" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"l_1","name":"CRS-1"}]`))
	}))
	defer srv.Close()

	c := spacexapi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "launches", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawPath != "/launches" {
		t.Fatalf("path = %q, want /launches", sawPath)
	}
	if len(got) != 1 || got[0]["id"] == nil {
		t.Fatalf("records = %+v, want launch record", got)
	}
}

func TestFixtureRegistryCatalogAndWrite(t *testing.T) {
	c := spacexapi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "spacex-api" || len(cat.Streams) == 0 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if _, ok := connectors.NewRegistry().Get("spacex-api"); !ok {
		t.Fatal("registry did not resolve spacex-api")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
