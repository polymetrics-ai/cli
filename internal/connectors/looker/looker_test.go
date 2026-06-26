package looker_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/looker"
)

func TestReadPaginatesAuthenticatesAndMapsUsers(t *testing.T) {
	var sawClientID string
	var sawAuth string
	var offsets []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/4.0/login":
			_ = r.ParseForm()
			sawClientID = r.Form.Get("client_id")
			_, _ = w.Write([]byte(`{"access_token":"looker_token","token_type":"Bearer","expires_in":3600}`))
		case "/api/4.0/users":
			sawAuth = r.Header.Get("Authorization")
			offsets = append(offsets, r.URL.Query().Get("offset"))
			switch r.URL.Query().Get("offset") {
			case "0", "":
				_, _ = w.Write([]byte(`[{"id":"1","display_name":"Ada","email":"ada@example.com"},{"id":"2","display_name":"Grace","email":"grace@example.com"}]`))
			case "2":
				_, _ = w.Write([]byte(`[{"id":"3","display_name":"Cleo","email":"cleo@example.com"}]`))
			default:
				t.Fatalf("unexpected offset %q", r.URL.Query().Get("offset"))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := looker.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/api/4.0", "token_url": srv.URL + "/api/4.0/login", "page_size": "2"},
		Secrets: map[string]string{"client_id": "looker_client", "client_secret": "looker_secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawClientID != "looker_client" || sawAuth != "Bearer looker_token" {
		t.Fatalf("client/auth = %q/%q", sawClientID, sawAuth)
	}
	if len(offsets) != 2 || offsets[0] != "0" || offsets[1] != "2" {
		t.Fatalf("offsets = %v", offsets)
	}
	if len(got) != 3 || got[0]["display_name"] != "Ada" || got[0]["email"] != "ada@example.com" {
		t.Fatalf("records not mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := looker.New()
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
	if err != nil || cat.Connector != "looker" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v err=%v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write err = %v", err)
	}
	if got, ok := connectors.NewRegistry().Get("looker"); !ok || got.Metadata().Capabilities.Write {
		t.Fatalf("registry/capabilities failed: ok=%v got=%T", ok, got)
	}
}
