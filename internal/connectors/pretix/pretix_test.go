package pretix_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/pretix"
)

func TestReadEventsAuthenticatesPaginatesAndMaps(t *testing.T) {
	var sawAuth string
	var pages []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/organizers/acme/events/" {
			http.NotFound(w, r)
			return
		}
		sawAuth = r.Header.Get("Authorization")
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		pages = append(pages, page)
		switch page {
		case "1":
			next := fmt.Sprintf("%s/api/v1/organizers/acme/events/?page=2", srv.URL)
			_, _ = w.Write([]byte(`{"results":[{"slug":"conf","name":{"en":"Conference"},"date_from":"2026-03-01T00:00:00Z"}],"next":"` + next + `"}`))
		case "2":
			_, _ = w.Write([]byte(`{"results":[{"slug":"workshop","name":{"en":"Workshop"},"date_from":"2026-04-01T00:00:00Z"}],"next":null}`))
		default:
			t.Fatalf("unexpected page %q", page)
		}
	}))
	defer srv.Close()

	c := pretix.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/api/v1", "organizer": "acme"}, Secrets: map[string]string{"api_token": "pretix-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Token pretix-token" {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if len(pages) != 2 || pages[0] != "1" || pages[1] != "2" {
		t.Fatalf("pages = %v", pages)
	}
	if len(got) != 2 || got[0]["id"] != "conf" || got[1]["name"] == nil {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := pretix.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "pretix" || len(cat.Streams) < 3 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	for _, stream := range cat.Streams {
		if len(stream.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", stream.Name)
		}
	}
	if _, ok := connectors.NewRegistry().Get("pretix"); !ok {
		t.Fatal("registry did not resolve pretix")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
