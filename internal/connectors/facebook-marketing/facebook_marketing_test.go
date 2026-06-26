package facebookmarketing_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	facebookmarketing "polymetrics.ai/internal/connectors/facebook-marketing"
)

func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/me/adaccounts" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("after") {
		case "":
			_, _ = fmt.Fprintf(w, `{"data":[{"id":"act_1","account_id":"1","name":"Primary"}],"paging":{"next":"%s/me/adaccounts?after=page-2"}}`, srv.URL)
		case "page-2":
			_, _ = w.Write([]byte(`{"data":[{"id":"act_2","account_id":"2","name":"Second"}],"paging":{}}`))
		default:
			t.Fatalf("unexpected after=%q", r.URL.Query().Get("after"))
		}
	}))
	defer srv.Close()

	c := facebookmarketing.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "1"},
		Secrets: map[string]string{"access_token": "test-access-token"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "ad_accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer test-access-token" {
		t.Fatalf("Authorization = %q, want bearer test token", sawAuth)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["account_id"] != "1" || got[0]["name"] != "Primary" {
		t.Fatalf("first record not mapped: %+v", got[0])
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := facebookmarketing.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "facebook-marketing" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v, want facebook-marketing streams", cat)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "ad_accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	if _, ok := connectors.NewRegistry().Get("facebook-marketing"); !ok {
		t.Fatal("registry did not resolve facebook-marketing")
	}
	_, err = c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, []connectors.Record{{"id": "1"}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
