package googledirectory_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	googledirectory "polymetrics.ai/internal/connectors/google-directory"
)

func TestReadPaginatesAuthenticatesAndMapsUsers(t *testing.T) {
	var sawAuth string
	var sawTokens []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/admin/directory/v1/users" {
			http.NotFound(w, r)
			return
		}
		sawTokens = append(sawTokens, r.URL.Query().Get("pageToken"))
		switch r.URL.Query().Get("pageToken") {
		case "":
			_, _ = w.Write([]byte(`{"nextPageToken":"NEXT","users":[{"id":"u1","primaryEmail":"ada@example.com","name":{"fullName":"Ada"}},{"id":"u2","primaryEmail":"grace@example.com","name":{"fullName":"Grace"}}]}`))
		case "NEXT":
			_, _ = w.Write([]byte(`{"users":[{"id":"u3","primaryEmail":"cleo@example.com","name":{"fullName":"Cleo"}}]}`))
		default:
			t.Fatalf("unexpected pageToken %q", r.URL.Query().Get("pageToken"))
		}
	}))
	defer srv.Close()

	c := googledirectory.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/admin/directory/v1", "customer_id": "my_customer", "page_size": "2"},
		Secrets: map[string]string{"authorization.access_token": "google_token"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer google_token" {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if len(sawTokens) != 2 || sawTokens[0] != "" || sawTokens[1] != "NEXT" {
		t.Fatalf("page tokens = %v", sawTokens)
	}
	if len(got) != 3 || got[0]["primary_email"] != "ada@example.com" || got[0]["name"] != "Ada" {
		t.Fatalf("records not mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := googledirectory.New()
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
	if err != nil || cat.Connector != "google-directory" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v err=%v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write err = %v", err)
	}
	if got, ok := connectors.NewRegistry().Get("google-directory"); !ok || got.Metadata().Capabilities.Write {
		t.Fatalf("registry/capabilities failed: ok=%v got=%T", ok, got)
	}
}
