package workdayrest_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	workdayrest "polymetrics.ai/internal/connectors/workday-rest"
)

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := workdayrest.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "workday-rest" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want workday-rest streams", cat)
	}
	var rows []connectors.Record
	err = c.Read(context.Background(), connectors.ReadRequest{Stream: "workers", Config: fixture}, func(rec connectors.Record) error {
		rows = append(rows, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(rows) == 0 || rows[0]["fixture"] != true || rows[0]["id"] == nil {
		t.Fatalf("fixture rows = %+v, want fixture records with id", rows)
	}
	if _, ok := connectors.NewRegistry().Get("workday-rest"); !ok {
		t.Fatal("registry did not resolve workday-rest")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}

func TestReadLiveUsesBearerAuthAndTenantPath(t *testing.T) {
	var sawBearer bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ccx/api/hcm/v1/tenant/workers" {
			http.NotFound(w, r)
			return
		}
		sawBearer = strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ")
		_, _ = w.Write([]byte(`{"data":[{"id":"worker-1","descriptor":"Ada","updated":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := workdayrest.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "tenant": "tenant", "page_size": "1", "max_pages": "1"},
		Secrets: map[string]string{"access_token": "test-token"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "workers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawBearer {
		t.Fatal("authorization header was not bearer")
	}
	if len(got) != 1 || got[0]["id"] != "worker-1" {
		t.Fatalf("records = %+v, want worker-1", got)
	}
}
