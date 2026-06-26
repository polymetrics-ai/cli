package microsoftdataverse_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	microsoftdataverse "polymetrics.ai/internal/connectors/microsoft-dataverse"
)

func TestReadPaginatesAuthenticatesAndMapsAccounts(t *testing.T) {
	var sawGrant string
	var sawScope string
	var sawAuth string
	var page2 bool
	var nextBase string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tenant/oauth2/v2.0/token":
			_ = r.ParseForm()
			sawGrant = r.Form.Get("grant_type")
			sawScope = r.Form.Get("scope")
			_, _ = w.Write([]byte(`{"access_token":"dataverse_token","expires_in":3600}`))
		case "/api/data/v9.2/accounts":
			sawAuth = r.Header.Get("Authorization")
			if r.URL.Query().Get("$skiptoken") == "NEXT" {
				page2 = true
				_, _ = w.Write([]byte(`{"value":[{"accountid":"a3","name":"Cleo Co","emailaddress1":"cleo@example.com"}]}`))
				return
			}
			payload := map[string]any{
				"@odata.nextLink": nextBase + "/api/data/v9.2/accounts?$skiptoken=NEXT",
				"value": []map[string]any{
					{"accountid": "a1", "name": "Ada Co", "emailaddress1": "ada@example.com"},
					{"accountid": "a2", "name": "Grace Co", "emailaddress1": "grace@example.com"},
				},
			}
			b, _ := json.Marshal(payload)
			_, _ = w.Write(b)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	nextBase = srv.URL

	c := microsoftdataverse.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/api/data/v9.2", "token_url": srv.URL + "/tenant/oauth2/v2.0/token", "scope": "https://org.crm.dynamics.com/.default", "page_size": "2"},
		Secrets: map[string]string{"client_id": "client", "client_secret": "secret", "tenant_id": "tenant"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawGrant != "client_credentials" || sawScope != "https://org.crm.dynamics.com/.default" || sawAuth != "Bearer dataverse_token" {
		t.Fatalf("grant/scope/auth = %q/%q/%q", sawGrant, sawScope, sawAuth)
	}
	if !page2 || len(got) != 3 || got[0]["id"] != "a1" || got[0]["name"] != "Ada Co" {
		t.Fatalf("records not mapped: page2=%v got=%+v", page2, got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := microsoftdataverse.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	var fixture []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		fixture = append(fixture, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(fixture) == 0 || fixture[0]["fixture"] != true {
		t.Fatalf("fixture records = %+v", fixture)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "microsoft-dataverse" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v err=%v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write err = %v", err)
	}
	if got, ok := connectors.NewRegistry().Get("microsoft-dataverse"); !ok || got.Metadata().Capabilities.Write {
		t.Fatalf("registry/capabilities failed: ok=%v got=%T", ok, got)
	}
}
