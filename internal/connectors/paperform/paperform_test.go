package paperform_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/paperform"
)

func TestReadPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var sawPage string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/forms" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"results":[{"id":"form_1","title":"Lead capture","created_at":"2026-01-01T00:00:00Z"},{"id":"form_2","title":"Survey","created_at":"2026-01-02T00:00:00Z"}],"has_more":true}`))
		case "2":
			sawPage = "2"
			_, _ = w.Write([]byte(`{"results":[{"id":"form_3","title":"Registration","created_at":"2026-01-03T00:00:00Z"}],"has_more":false}`))
		default:
			t.Fatalf("unexpected page query %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := paperform.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "limit": "2"}, Secrets: map[string]string{"api_key": "paperform_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "forms", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer paperform_key" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if sawPage != "2" {
		t.Fatalf("second page = %q, want 2", sawPage)
	}
	if len(got) != 3 || got[0]["id"] != "form_1" || got[2]["title"] != "Registration" {
		t.Fatalf("mapped records = %+v, want three forms", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := paperform.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "submissions", Config: fixture}, func(rec connectors.Record) error {
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
	if cat.Connector != "paperform" || len(cat.Streams) < 2 {
		t.Fatalf("catalog = %+v, want paperform streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("paperform"); !ok {
		t.Fatal("registry did not resolve paperform")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
