package zoom_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/zoom"
)

func TestReadUsersAuthenticatesAndFollowsNextPageToken(t *testing.T) {
	var sawAuth string
	var sawToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v2/users" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("next_page_token") {
		case "":
			_, _ = w.Write([]byte(`{"users":[{"id":"user_1","email":"ada@example.com","first_name":"Ada","last_name":"Lovelace"}],"next_page_token":"next_1"}`))
		case "next_1":
			sawToken = "next_1"
			_, _ = w.Write([]byte(`{"users":[{"id":"user_2","email":"grace@example.com","first_name":"Grace","last_name":"Hopper"}],"next_page_token":""}`))
		default:
			t.Fatalf("unexpected token %q", r.URL.Query().Get("next_page_token"))
		}
	}))
	defer srv.Close()

	c := zoom.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/v2", "page_size": "1"}, Secrets: map[string]string{"access_token": "test_access_token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer test_access_token" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if sawToken != "next_1" {
		t.Fatalf("next_page_token = %q, want next_1", sawToken)
	}
	if len(got) != 2 || got[0]["id"] != "user_1" || got[1]["email"] != "grace@example.com" {
		t.Fatalf("records = %+v, want two mapped users", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := zoom.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "meetings", Config: fixture}, func(rec connectors.Record) error {
		rows = append(rows, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(rows) == 0 || rows[0]["id"] == nil {
		t.Fatalf("fixture rows = %+v, want records with id", rows)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "zoom" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want zoom streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("zoom"); !ok {
		t.Fatal("registry did not resolve zoom")
	}
	if c.Metadata().Capabilities.Write {
		t.Fatal("zoom should be read-only")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
