package printify_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/printify"
)

func TestReadProductsAuthenticatesPaginatesAndMaps(t *testing.T) {
	var sawAuth string
	var pages []string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/shops/42/products.json" {
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
			_, _ = w.Write([]byte(`{"data":[{"id":"p1","title":"Shirt","visible":true}],"current_page":1,"last_page":2,"next_page_url":"` + fmt.Sprintf("%s/v1/shops/42/products.json?page=2", srv.URL) + `"}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"p2","title":"Mug","visible":false}],"current_page":2,"last_page":2,"next_page_url":null}`))
		default:
			t.Fatalf("unexpected page %q", page)
		}
	}))
	defer srv.Close()

	c := printify.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/v1", "shop_id": "42"}, Secrets: map[string]string{"api_token": "printify-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer printify-token" {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if len(pages) != 2 || pages[0] != "1" || pages[1] != "2" {
		t.Fatalf("pages = %v", pages)
	}
	if len(got) != 2 || got[0]["id"] != "p1" || got[1]["title"] != "Mug" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestReadShopsMapsTopLevelArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/shops.json" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":42,"title":"API Shop","sales_channel":"api"}]`))
	}))
	defer srv.Close()

	c := printify.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/v1"}, Secrets: map[string]string{"api_token": "printify-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "shops", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["id"] == nil || got[0]["title"] != "API Shop" {
		t.Fatalf("unexpected shop records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := printify.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error {
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
	if cat.Connector != "printify" || len(cat.Streams) < 4 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	for _, stream := range cat.Streams {
		if len(stream.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", stream.Name)
		}
	}
	if _, ok := connectors.NewRegistry().Get("printify"); !ok {
		t.Fatal("registry did not resolve printify")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
