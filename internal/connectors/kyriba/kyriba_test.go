package kyriba_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/kyriba"
)

func TestReadPaginatesAuthenticatesAndMapsBankAccounts(t *testing.T) {
	var sawGrant string
	var sawAuth string
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			_ = r.ParseForm()
			sawGrant = r.Form.Get("grant_type")
			_, _ = w.Write([]byte(`{"access_token":"kyriba_token","expires_in":3600}`))
		case "/api/v1/bank-accounts":
			sawAuth = r.Header.Get("Authorization")
			pages = append(pages, r.URL.Query().Get("page"))
			switch r.URL.Query().Get("page") {
			case "1", "":
				_, _ = w.Write([]byte(`{"data":[{"id":"ba1","accountNumber":"001","currency":"USD"},{"id":"ba2","accountNumber":"002","currency":"EUR"}]}`))
			case "2":
				_, _ = w.Write([]byte(`{"data":[{"id":"ba3","accountNumber":"003","currency":"GBP"}]}`))
			default:
				t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := kyriba.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/api/v1", "token_url": srv.URL + "/oauth/token", "page_size": "2"},
		Secrets: map[string]string{"client_id": "client", "client_secret": "secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bank_accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawGrant != "client_credentials" || sawAuth != "Bearer kyriba_token" {
		t.Fatalf("grant/auth = %q/%q", sawGrant, sawAuth)
	}
	if len(pages) != 2 || len(got) != 3 || got[0]["account_number"] != "001" {
		t.Fatalf("pages=%v records=%+v", pages, got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := kyriba.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	var fixture []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bank_accounts", Config: cfg}, func(rec connectors.Record) error {
		fixture = append(fixture, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(fixture) == 0 || fixture[0]["fixture"] != true {
		t.Fatalf("fixture records = %+v", fixture)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "kyriba" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v err=%v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write err = %v", err)
	}
	if got, ok := connectors.NewRegistry().Get("kyriba"); !ok || got.Metadata().Capabilities.Write {
		t.Fatalf("registry/capabilities failed: ok=%v got=%T", ok, got)
	}
}
