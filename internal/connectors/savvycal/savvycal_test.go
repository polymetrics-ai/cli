package savvycal_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/savvycal"
)

func TestReadEventsFollowsLinksAuthenticatesAndMaps(t *testing.T) {
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/events" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatal("missing bearer authorization")
		}
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		pages = append(pages, page)
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{"data":[{"id":"ev_1","name":"Intro"},{"id":"ev_2","name":"Demo"}],"links":{"next":"/v1/events?page=2"}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"ev_3","name":"Review"}],"links":{"next":null}}`))
		default:
			t.Fatalf("unexpected page %s", page)
		}
	}))
	defer srv.Close()

	c := savvycal.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "2"}, Secrets: map[string]string{"api_key": "test-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(pages) != 2 || len(got) != 3 || got[0]["id"] != "ev_1" {
		t.Fatalf("unexpected pages/records pages=%v records=%+v", pages, got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := savvycal.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "savvycal" || len(cat.Streams) == 0 {
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
	if _, ok := connectors.NewRegistry().Get("savvycal"); !ok {
		t.Fatal("registry did not resolve savvycal")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
