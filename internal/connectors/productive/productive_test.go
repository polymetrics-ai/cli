package productive_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/productive"
)

func TestReadProjectsAuthenticatesPaginatesAndMaps(t *testing.T) {
	var sawAuth, sawOrg string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/projects" {
			http.NotFound(w, r)
			return
		}
		sawAuth = r.Header.Get("X-Auth-Token")
		sawOrg = r.Header.Get("X-Organization-Id")
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		pages = append(pages, page)
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{"data":[{"id":"101","type":"projects","attributes":{"name":"Migration","updated_at":"2026-01-01T00:00:00Z"}}],"meta":{"current_page":1,"total_pages":2}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"102","type":"projects","attributes":{"name":"Support","updated_at":"2026-01-02T00:00:00Z"}}],"meta":{"current_page":2,"total_pages":2}}`))
		default:
			t.Fatalf("unexpected page %q", page)
		}
	}))
	defer srv.Close()

	c := productive.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/api/v2", "organization_id": "org-1"}, Secrets: map[string]string{"api_key": "productive-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "productive-token" || sawOrg != "org-1" {
		t.Fatalf("headers auth=%q org=%q", sawAuth, sawOrg)
	}
	if len(pages) != 2 || pages[0] != "1" || pages[1] != "2" {
		t.Fatalf("pages = %v", pages)
	}
	if len(got) != 2 || got[0]["id"] != "101" || got[1]["name"] != "Support" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := productive.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
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
	if cat.Connector != "productive" || len(cat.Streams) < 3 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	for _, stream := range cat.Streams {
		if len(stream.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", stream.Name)
		}
	}
	if _, ok := connectors.NewRegistry().Get("productive"); !ok {
		t.Fatal("registry did not resolve productive")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
