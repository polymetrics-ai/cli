package pagerduty_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/pagerduty"
)

func TestReadPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var sawOffset string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/incidents" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"incidents":[{"id":"P1","incident_number":101,"title":"Disk full","status":"triggered"},{"id":"P2","incident_number":102,"title":"CPU high","status":"acknowledged"}],"more":true}`))
		case "2":
			sawOffset = "2"
			_, _ = w.Write([]byte(`{"incidents":[{"id":"P3","incident_number":103,"title":"Recovered","status":"resolved"}],"more":false}`))
		default:
			t.Fatalf("unexpected offset query %q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	c := pagerduty.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "limit": "2"}, Secrets: map[string]string{"api_key": "pager_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "incidents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Token token=pager_key" {
		t.Fatalf("Authorization = %q, want PagerDuty token auth", sawAuth)
	}
	if sawOffset != "2" {
		t.Fatalf("second page offset = %q, want 2", sawOffset)
	}
	if len(got) != 3 || got[0]["id"] != "P1" || got[2]["status"] != "resolved" {
		t.Fatalf("mapped records = %+v, want three incidents", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := pagerduty.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: fixture}, func(rec connectors.Record) error {
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
	if cat.Connector != "pagerduty" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want pagerduty streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("pagerduty"); !ok {
		t.Fatal("registry did not resolve pagerduty")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
