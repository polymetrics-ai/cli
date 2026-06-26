package asana_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/asana"
)

func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/projects" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "":
			_, _ = fmt.Fprintf(w, `{"data":[{"gid":"p1","name":"Launch","resource_type":"project"}],"next_page":{"uri":"%s/projects?offset=page-2"}}`, srv.URL)
		case "page-2":
			_, _ = w.Write([]byte(`{"data":[{"gid":"p2","name":"Retention","resource_type":"project"}],"next_page":null}`))
		default:
			t.Fatalf("unexpected offset=%q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	c := asana.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1"}, Secrets: map[string]string{"access_token": "test-access-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer test-access-token" {
		t.Fatalf("Authorization = %q, want bearer test token", sawAuth)
	}
	if len(got) != 2 || got[0]["gid"] != "p1" || got[0]["name"] != "Launch" {
		t.Fatalf("records not paginated/mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := asana.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "asana" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v, want asana streams", cat)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 || got[0]["gid"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	if _, ok := connectors.NewRegistry().Get("asana"); !ok {
		t.Fatal("registry did not resolve asana")
	}
	_, err = c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, []connectors.Record{{"gid": "1"}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
