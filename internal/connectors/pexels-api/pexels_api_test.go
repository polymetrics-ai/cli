package pexelsapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	pexelsapi "polymetrics.ai/internal/connectors/pexels-api"
)

func TestReadPhotosPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var pages []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		pages = append(pages, r.URL.Query().Get("page"))
		if r.URL.Path != "/v1/search" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`{"page":1,"per_page":1,"photos":[{"id":101,"url":"https://pexels.example/101","photographer":"Ada","src":{"original":"https://img.example/101.jpg"},"alt":"A photo"}],"next_page":"` + srv.URL + `/v1/search?page=2&per_page=1&query=nature"}`))
		case "2":
			_, _ = w.Write([]byte(`{"page":2,"per_page":1,"photos":[{"id":102,"url":"https://pexels.example/102","photographer":"Grace","src":{"original":"https://img.example/102.jpg"},"alt":"Another photo"}]}`))
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := pexelsapi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "query": "nature", "page_size": "1"}, Secrets: map[string]string{"api_key": "pexels_key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "photos", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "pexels_key" {
		t.Fatalf("Authorization = %q, want pexels_key", sawAuth)
	}
	if len(got) != 2 || got[0]["id"] == nil || got[1]["photographer"] != "Grace" {
		t.Fatalf("records = %+v, want mapped photos", got)
	}
	if len(pages) != 2 {
		t.Fatalf("pages = %v, want two requests", pages)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := pexelsapi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "photos", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "pexels-api" || len(cat.Streams) < 2 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("pexels-api"); !ok {
		t.Fatal("registry did not resolve pexels-api")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
