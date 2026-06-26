package primetric_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/primetric"
)

func TestReadEmployeesFetchesTokenAuthenticatesPaginatesAndMaps(t *testing.T) {
	var tokenRequests int
	var sawBearer string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			tokenRequests++
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm: %v", err)
			}
			if r.Form.Get("client_id") != "client-id" || r.Form.Get("client_secret") != "client-secret" {
				t.Fatalf("unexpected token form: %s", r.Form.Encode())
			}
			_, _ = w.Write([]byte(`{"access_token":"access-token","expires_in":3600}`))
		case "/api/v1/employees":
			sawBearer = r.Header.Get("Authorization")
			pages = append(pages, r.URL.Query().Get("page"))
			switch r.URL.Query().Get("page") {
			case "", "1":
				_, _ = w.Write([]byte(`{"data":[{"id":11,"first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}],"meta":{"current_page":1,"total_pages":2}}`))
			case "2":
				_, _ = w.Write([]byte(`{"data":[{"id":12,"first_name":"Grace","last_name":"Hopper","email":"grace@example.com"}],"meta":{"current_page":2,"total_pages":2}}`))
			default:
				t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := primetric.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/api/v1", "token_url": srv.URL + "/oauth/token"}, Secrets: map[string]string{"client_id": "client-id", "client_secret": "client-secret"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "employees", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if tokenRequests != 1 {
		t.Fatalf("token requests = %d, want 1", tokenRequests)
	}
	if sawBearer != "Bearer access-token" {
		t.Fatalf("Authorization = %q", sawBearer)
	}
	if strings.Join(pages, ",") != "1,2" {
		t.Fatalf("pages = %v", pages)
	}
	if len(got) != 2 || got[0]["id"] == nil || got[1]["email"] != "grace@example.com" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := primetric.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "employees", Config: cfg}, func(rec connectors.Record) error {
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
	if cat.Connector != "primetric" || len(cat.Streams) < 3 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	for _, stream := range cat.Streams {
		if len(stream.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", stream.Name)
		}
	}
	if _, ok := connectors.NewRegistry().Get("primetric"); !ok {
		t.Fatal("registry did not resolve primetric")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
