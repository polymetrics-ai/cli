package orb_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/orb"
)

func TestReadPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var cursors []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/customers" {
			http.NotFound(w, r)
			return
		}
		cursors = append(cursors, r.URL.Query().Get("cursor"))
		switch r.URL.Query().Get("cursor") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"cus_1","name":"Acme","created_at":"2026-01-01T00:00:00Z"}],"pagination_metadata":{"has_more":true,"next_cursor":"cur_2"}}`))
		case "cur_2":
			_, _ = w.Write([]byte(`{"data":[{"id":"cus_2","name":"Globex","created_at":"2026-01-02T00:00:00Z"}],"pagination_metadata":{"has_more":false}}`))
		default:
			t.Fatalf("unexpected cursor %q", r.URL.Query().Get("cursor"))
		}
	}))
	defer srv.Close()

	c := orb.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}, Secrets: map[string]string{"api_key": "orb_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer orb_key" {
		t.Fatalf("Authorization = %q, want Bearer", sawAuth)
	}
	if len(cursors) != 2 || cursors[1] != "cur_2" || len(got) != 2 || got[1]["name"] != "Globex" {
		t.Fatalf("pagination/mapping wrong: cursors=%v records=%+v", cursors, got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := orb.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "subscriptions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records missing id: %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) < 3 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("orb"); !ok {
		t.Fatal("registry did not resolve orb")
	}
}
