package rollbar_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/rollbar"
)

func TestReadItemsPaginatesAuthenticatesAndMaps(t *testing.T) {
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/1/items/" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("X-Rollbar-Access-Token") == "" {
			t.Fatal("missing Rollbar access token header")
		}
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		pages = append(pages, page)
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{"result":{"page":1,"total_pages":2,"items":[{"id":101,"title":"panic one","environment":"prod"},{"id":102,"title":"panic two","environment":"prod"}]}}`))
		case "2":
			_, _ = w.Write([]byte(`{"result":{"page":2,"total_pages":2,"items":[{"id":103,"title":"panic three","environment":"prod"}]}}`))
		default:
			t.Fatalf("unexpected page %s", page)
		}
	}))
	defer srv.Close()

	c := rollbar.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "2"}, Secrets: map[string]string{"access_token": "test-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "items", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(pages) != 2 || pages[0] != "1" || pages[1] != "2" {
		t.Fatalf("pages = %v, want [1 2]", pages)
	}
	if len(got) != 3 || got[0]["id"] == nil || got[0]["title"] != "panic one" {
		t.Fatalf("records not mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := rollbar.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "rollbar" || len(cat.Streams) == 0 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	for _, stream := range cat.Streams {
		if len(stream.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", stream.Name)
		}
		count := 0
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream.Name, Config: cfg}, func(connectors.Record) error { count++; return nil }); err != nil {
			t.Fatalf("fixture Read(%s): %v", stream.Name, err)
		}
		if count == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream.Name)
		}
	}
	if _, ok := connectors.NewRegistry().Get("rollbar"); !ok {
		t.Fatal("registry did not resolve rollbar")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
