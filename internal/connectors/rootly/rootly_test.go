package rootly_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/rootly"
)

func TestReadIncidentsFollowsLinksAuthenticatesAndFlattensAttributes(t *testing.T) {
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/incidents" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatal("missing bearer authorization")
		}
		page := r.URL.Query().Get("page[number]")
		if page == "" {
			page = "1"
		}
		pages = append(pages, page)
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{"data":[{"id":"inc_1","attributes":{"title":"API outage","status":"resolved"}},{"id":"inc_2","attributes":{"title":"Queue lag","status":"started"}}],"links":{"next":"/v1/incidents?page[number]=2&page[size]=2"}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"inc_3","attributes":{"title":"Cache miss","status":"mitigated"}}],"links":{"next":null}}`))
		default:
			t.Fatalf("unexpected page %s", page)
		}
	}))
	defer srv.Close()

	c := rootly.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "2"}, Secrets: map[string]string{"api_key": "test-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "incidents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(pages) != 2 || pages[0] != "1" || pages[1] != "2" {
		t.Fatalf("pages = %v, want [1 2]", pages)
	}
	if len(got) != 3 || got[0]["id"] != "inc_1" || got[0]["title"] != "API outage" {
		t.Fatalf("records not flattened: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := rootly.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "rootly" || len(cat.Streams) == 0 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	for _, stream := range cat.Streams {
		count := 0
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream.Name, Config: cfg}, func(connectors.Record) error { count++; return nil }); err != nil {
			t.Fatalf("fixture Read(%s): %v", stream.Name, err)
		}
		if count == 0 || len(stream.PrimaryKey) == 0 {
			t.Fatalf("stream %q not fixture-ready: count=%d pk=%v", stream.Name, count, stream.PrimaryKey)
		}
	}
	if _, ok := connectors.NewRegistry().Get("rootly"); !ok {
		t.Fatal("registry did not resolve rootly")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
