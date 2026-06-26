package wufoo_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/wufoo"
)

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := wufoo.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "wufoo" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want wufoo streams", cat)
	}
	var rows []connectors.Record
	err = c.Read(context.Background(), connectors.ReadRequest{Stream: "forms", Config: fixture}, func(rec connectors.Record) error {
		rows = append(rows, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(rows) == 0 || rows[0]["fixture"] != true || rows[0]["id"] == nil {
		t.Fatalf("fixture rows = %+v, want fixture records with id", rows)
	}
	if _, ok := connectors.NewRegistry().Get("wufoo"); !ok {
		t.Fatal("registry did not resolve wufoo")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}

func TestReadLiveUsesBasicAuthAndMapsForms(t *testing.T) {
	var sawBasicAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/forms.json" {
			http.NotFound(w, r)
			return
		}
		sawBasicAuth = strings.HasPrefix(r.Header.Get("Authorization"), "Basic ")
		_, _ = w.Write([]byte(`{"Forms":[{"Hash":"form-1","Name":"Intake","DateUpdated":"2026-01-01 00:00:00"}]}`))
	}))
	defer srv.Close()

	c := wufoo.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1", "max_pages": "1"}, Secrets: map[string]string{"api_key": "test-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "forms", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawBasicAuth {
		t.Fatal("authorization header was not HTTP Basic")
	}
	if len(got) != 1 || got[0]["Hash"] != "form-1" {
		t.Fatalf("records = %+v, want form-1", got)
	}
}
