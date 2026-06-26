package iterable_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/iterable"
)

func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("Api-Key")
		if r.URL.Path != "/lists" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("pageToken") {
		case "":
			_, _ = w.Write([]byte(`{"lists":[{"id":1,"name":"Newsletter","listType":"Standard"}],"nextPageToken":"page-2"}`))
		case "page-2":
			_, _ = w.Write([]byte(`{"lists":[{"id":2,"name":"Buyers","listType":"Standard"}]}`))
		default:
			t.Fatalf("unexpected pageToken=%q", r.URL.Query().Get("pageToken"))
		}
	}))
	defer srv.Close()

	c := iterable.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1"}, Secrets: map[string]string{"api_key": "test-api-key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "test-api-key" {
		t.Fatalf("Api-Key header = %q, want test key", sawKey)
	}
	if len(got) != 2 || got[0]["id"] == nil || got[0]["name"] != "Newsletter" {
		t.Fatalf("records not paginated/mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := iterable.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "iterable" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v, want iterable streams", cat)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	if _, ok := connectors.NewRegistry().Get("iterable"); !ok {
		t.Fatal("registry did not resolve iterable")
	}
	_, err = c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, []connectors.Record{{"id": "1"}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
