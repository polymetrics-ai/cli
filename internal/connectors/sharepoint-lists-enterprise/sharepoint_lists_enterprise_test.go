package sharepointlistsenterprise_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	sharepointlistsenterprise "polymetrics.ai/internal/connectors/sharepoint-lists-enterprise"
)

func TestContractFixtureAndWrite(t *testing.T) {
	c := sharepointlistsenterprise.New()
	if c.Name() != "sharepoint-lists-enterprise" {
		t.Fatalf("Name() = %q, want sharepoint-lists-enterprise", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want read-only Check/Catalog/Read", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) == 0 || cat.Streams[0].Name != "lists" {
		t.Fatalf("catalog streams = %+v, want lists first", cat.Streams)
	}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("sharepoint-lists-enterprise"); !ok {
		t.Fatal("registry did not resolve sharepoint-lists-enterprise")
	}
}

func TestReadListsUsesClientCredentialsBearer(t *testing.T) {
	var sawAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tenant-1/oauth2/v2.0/token":
			_, _ = w.Write([]byte(`{"access_token":"issued-token","expires_in":3600}`))
		case "/sites/site-1/lists":
			sawAuth = r.Header.Get("Authorization") == "Bearer issued-token"
			_, _ = w.Write([]byte(`{"value":[{"id":"list_1","name":"Accounts","displayName":"Accounts"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "login_base_url": srv.URL, "tenant_id": "tenant-1", "site_id": "site-1", "max_pages": "1"}, Secrets: map[string]string{"client_id": "test-client", "client_secret": "test-secret"}}
	var got []connectors.Record
	if err := sharepointlistsenterprise.New().Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawAuth {
		t.Fatal("oauth bearer auth was not applied")
	}
	if len(got) != 1 || got[0]["displayName"] != "Accounts" {
		t.Fatalf("records = %+v, want list displayName", got)
	}
}
