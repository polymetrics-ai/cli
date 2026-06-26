package zohodesk_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	zohodesk "polymetrics.ai/internal/connectors/zoho-desk"
)

func TestReadTicketsAuthenticatesAndSetsOrgHeader(t *testing.T) {
	var sawAuth string
	var sawOrg string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawOrg = r.Header.Get("orgId")
		if r.URL.Path != "/desk/tickets" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"ticket_1","ticketNumber":"101","subject":"Help","modifiedTime":"2026-01-02T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := zohodesk.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/desk", "org_id": "desk_org"}, Secrets: map[string]string{"access_token": "test_access_token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tickets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Zoho-oauthtoken test_access_token" {
		t.Fatalf("Authorization = %q, want Zoho-oauthtoken", sawAuth)
	}
	if sawOrg != "desk_org" {
		t.Fatalf("orgId header = %q, want desk_org", sawOrg)
	}
	if len(got) != 1 || got[0]["id"] != "ticket_1" || got[0]["subject"] != "Help" {
		t.Fatalf("records = %+v, want mapped ticket", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := zohodesk.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: fixture}, func(rec connectors.Record) error {
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
	if cat.Connector != "zoho-desk" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want zoho-desk streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("zoho-desk"); !ok {
		t.Fatal("registry did not resolve zoho-desk")
	}
	if c.Metadata().Capabilities.Write {
		t.Fatal("zoho-desk should be read-only")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
