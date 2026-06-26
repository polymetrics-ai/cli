package payfit_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/payfit"
)

func TestReadPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var sawOffset string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/employees" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"data":[{"id":"emp_1","first_name":"Ada","last_name":"Lovelace"},{"id":"emp_2","first_name":"Grace","last_name":"Hopper"}],"meta":{"next_offset":"2"}}`))
		case "2":
			sawOffset = "2"
			_, _ = w.Write([]byte(`{"data":[{"id":"emp_3","first_name":"Katherine","last_name":"Johnson"}],"meta":{"next_offset":""}}`))
		default:
			t.Fatalf("unexpected offset query %q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	c := payfit.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "limit": "2"}, Secrets: map[string]string{"api_key": "payfit_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "employees", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer payfit_key" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if sawOffset != "2" {
		t.Fatalf("second page offset = %q, want 2", sawOffset)
	}
	if len(got) != 3 || got[0]["id"] != "emp_1" || got[2]["last_name"] != "Johnson" {
		t.Fatalf("mapped records = %+v, want three employees", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := payfit.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contracts", Config: fixture}, func(rec connectors.Record) error {
		rows = append(rows, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(rows) == 0 || rows[0]["id"] == nil {
		t.Fatalf("fixture rows = %+v, want records with id", rows)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "payfit" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want payfit streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("payfit"); !ok {
		t.Fatal("registry did not resolve payfit")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
