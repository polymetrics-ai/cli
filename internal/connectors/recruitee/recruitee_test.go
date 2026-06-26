package recruitee_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/recruitee"
)

func TestReadOffersAuthenticatesPaginatesAndMapsRecords(t *testing.T) {
	var sawAuth string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/c/42/offers" {
			http.NotFound(w, r)
			return
		}
		pages = append(pages, r.URL.Query().Get("page"))
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`{"offers":[{"id":1,"title":"Backend Engineer","status":"published"},{"id":2,"title":"Data Engineer","status":"published"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"offers":[{"id":3,"title":"Product Analyst","status":"draft"}]}`))
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := recruitee.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "company_id": "42", "page_size": "2"}, Secrets: map[string]string{"api_key": "recruitee_key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "offers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer recruitee_key" {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if len(pages) != 2 || pages[0] != "1" || pages[1] != "2" {
		t.Fatalf("pages = %v", pages)
	}
	if len(got) != 3 || got[0]["id"] == nil || got[0]["title"] != "Backend Engineer" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := recruitee.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "recruitee" || len(cat.Streams) == 0 {
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
	if _, ok := connectors.NewRegistry().Get("recruitee"); !ok {
		t.Fatal("registry did not resolve recruitee")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
