package partnerstack_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/partnerstack"
)

func TestReadPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var sawCursor string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/partnerships" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"ptn_1","status":"active","email":"a@example.com"},{"id":"ptn_2","status":"pending","email":"b@example.com"}],"pagination":{"next":"cursor_2"}}`))
		case "cursor_2":
			sawCursor = "cursor_2"
			_, _ = w.Write([]byte(`{"data":[{"id":"ptn_3","status":"inactive","email":"c@example.com"}],"pagination":{"next":""}}`))
		default:
			t.Fatalf("unexpected cursor query %q", r.URL.Query().Get("cursor"))
		}
	}))
	defer srv.Close()

	c := partnerstack.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "limit": "2"}, Secrets: map[string]string{"api_key": "partnerstack_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "partnerships", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer partnerstack_key" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if sawCursor != "cursor_2" {
		t.Fatalf("second page cursor = %q, want cursor_2", sawCursor)
	}
	if len(got) != 3 || got[0]["id"] != "ptn_1" || got[2]["status"] != "inactive" {
		t.Fatalf("mapped records = %+v, want three partnerships", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := partnerstack.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: fixture}, func(rec connectors.Record) error {
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
	if cat.Connector != "partnerstack" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want partnerstack streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("partnerstack"); !ok {
		t.Fatal("registry did not resolve partnerstack")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
