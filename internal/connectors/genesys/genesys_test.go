package genesys_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/genesys"
)

func TestReadPaginatesAuthenticatesAndMapsUsers(t *testing.T) {
	var sawGrant string
	var sawAuth string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			_ = r.ParseForm()
			sawGrant = r.Form.Get("grant_type")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"genesys_token","expires_in":3600}`))
		case "/api/v2/users":
			sawAuth = r.Header.Get("Authorization")
			pages = append(pages, r.URL.Query().Get("pageNumber"))
			switch r.URL.Query().Get("pageNumber") {
			case "1", "":
				_, _ = w.Write([]byte(`{"entities":[{"id":"u1","name":"Ada","email":"ada@example.com"},{"id":"u2","name":"Grace","email":"grace@example.com"}]}`))
			case "2":
				_, _ = w.Write([]byte(`{"entities":[{"id":"u3","name":"Cleo","email":"cleo@example.com"}]}`))
			default:
				t.Fatalf("unexpected pageNumber %q", r.URL.Query().Get("pageNumber"))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := genesys.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/api/v2", "token_url": srv.URL + "/oauth/token", "page_size": "2"},
		Secrets: map[string]string{"client_id": "client", "client_secret": "secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawGrant != "client_credentials" || sawAuth != "Bearer genesys_token" {
		t.Fatalf("grant/auth = %q/%q", sawGrant, sawAuth)
	}
	if strings.Join(pages, ",") != "1,2" {
		t.Fatalf("pages = %v", pages)
	}
	if len(got) != 3 || got[0]["display_name"] != "Ada" || got[0]["email"] != "ada@example.com" {
		t.Fatalf("records not mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := genesys.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	var fixture []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		fixture = append(fixture, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(fixture) == 0 || fixture[0]["fixture"] != true {
		t.Fatalf("fixture records = %+v", fixture)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "genesys" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v err=%v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write err = %v", err)
	}
	if got, ok := connectors.NewRegistry().Get("genesys"); !ok || got.Metadata().Capabilities.Write {
		t.Fatalf("registry/capabilities failed: ok=%v got=%T", ok, got)
	}
}
