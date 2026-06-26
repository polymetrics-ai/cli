package qualaroo_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/qualaroo"
)

func TestReadNudgesAuthenticatesPaginatesAndMapsRecords(t *testing.T) {
	var sawAuth string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v1/nudges" {
			http.NotFound(w, r)
			return
		}
		pages = append(pages, r.URL.Query().Get("page"))
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"nudges":[{"id":101,"name":"Homepage survey","status":"active"},{"id":102,"name":"Pricing survey","status":"paused"}],"pagination":{"next_page":2}}`))
		case "2":
			_, _ = w.Write([]byte(`{"nudges":[{"id":103,"name":"Checkout survey","status":"active"}],"pagination":{}}`))
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := qualaroo.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}, Secrets: map[string]string{"api_key": "qualaroo_key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "nudges", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != `Token token="qualaroo_key"` {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if len(pages) != 2 || pages[0] != "1" || pages[1] != "2" {
		t.Fatalf("pages = %v, want [1 2]", pages)
	}
	if len(got) != 3 || got[0]["id"] == nil || got[0]["name"] != "Homepage survey" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := qualaroo.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "qualaroo" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v", cat)
	}
	for _, stream := range cat.Streams {
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream.Name, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		}); err != nil {
			t.Fatalf("Read fixture %s: %v", stream.Name, err)
		}
		if len(got) == 0 || got[0]["id"] == nil {
			t.Fatalf("fixture %s records = %+v", stream.Name, got)
		}
	}
	if _, ok := connectors.NewRegistry().Get("qualaroo"); !ok {
		t.Fatal("registry did not resolve qualaroo")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
