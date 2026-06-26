package scryfall_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/scryfall"
)

func TestReadCardsPaginatesPublicAPIAndMaps(t *testing.T) {
	var pages []string
	var sawAuth bool
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cards/search" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "" {
			sawAuth = true
		}
		if r.URL.Query().Get("q") != "type:legendary" {
			t.Fatalf("q = %q, want type:legendary", r.URL.Query().Get("q"))
		}
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		pages = append(pages, page)
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{"object":"list","has_more":true,"next_page":"` + srv.URL + `/cards/search?q=type%3Alegendary&page=2","data":[{"id":"card_1","name":"Atraxa","set":"one"},{"id":"card_2","name":"Jodah","set":"dmu"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"object":"list","has_more":false,"data":[{"id":"card_3","name":"Krenko","set":"war"}]}`))
		default:
			t.Fatalf("unexpected page %s", page)
		}
	}))
	defer srv.Close()

	c := scryfall.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "q": "type:legendary"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "cards", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth {
		t.Fatal("scryfall connector sent credentials")
	}
	if len(pages) != 2 || len(got) != 3 || got[0]["name"] != "Atraxa" {
		t.Fatalf("unexpected pages/records pages=%v records=%+v", pages, got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := scryfall.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "scryfall" || len(cat.Streams) == 0 {
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
	if _, ok := connectors.NewRegistry().Get("scryfall"); !ok {
		t.Fatal("registry did not resolve scryfall")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
