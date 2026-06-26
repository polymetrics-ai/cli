package smartsheets_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/smartsheets"
)

func TestReadRowsPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/2.0/sheets/900" {
			http.NotFound(w, r)
			return
		}
		pages = append(pages, r.URL.Query().Get("page"))
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"id":900,"name":"Plan","pageNumber":1,"totalPages":2,"columns":[{"id":11,"title":"Name"}],"rows":[{"id":1,"rowNumber":1,"modifiedAt":"2026-01-01T00:00:00Z","cells":[{"columnId":11,"value":"Alpha"}]}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"id":900,"name":"Plan","pageNumber":2,"totalPages":2,"columns":[{"id":11,"title":"Name"}],"rows":[{"id":2,"rowNumber":2,"modifiedAt":"2026-01-02T00:00:00Z","cells":[{"columnId":11,"value":"Beta"}]}]}`))
		default:
			t.Fatalf("unexpected page=%q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := smartsheets.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/2.0", "spreadsheet_id": "900", "page_size": "1"}, Secrets: map[string]string{"access_token": "test-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sheet_rows", Config: cfg}, func(rec connectors.Record) error {
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
		t.Fatalf("pages=%v records=%d, want 2 pages/records", pages, len(got))
	}
	if got[0]["row_id"] != float64(1) || got[0]["Name"] != "Alpha" || got[1]["Name"] != "Beta" {
		t.Fatalf("rows not mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistryAndWrite(t *testing.T) {
	c := smartsheets.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var n int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sheet_rows", Config: cfg}, func(connectors.Record) error { n++; return nil }); err != nil {
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
	if _, ok := connectors.NewRegistry().Get("smartsheets"); !ok {
		t.Fatal("registry did not resolve smartsheets")
	}
}
