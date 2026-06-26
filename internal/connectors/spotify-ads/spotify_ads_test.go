package spotifyads_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	spotifyads "polymetrics.ai/internal/connectors/spotify-ads"
)

func TestReadAdAccountsAuthenticatesAndMapsRecords(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/ad_accounts" {
			http.NotFound(w, r)
			return
		}
		if got := r.URL.Query().Get("limit"); got != "2" {
			t.Fatalf("limit = %q, want 2", got)
		}
		_, _ = w.Write([]byte(`{"ad_accounts":[{"id":"acct1","name":"Main Account","country":"US"}]}`))
	}))
	defer srv.Close()

	c := spotifyads.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "2"}, Secrets: map[string]string{"access_token": "test-token"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "ad_accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer test-token" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if len(got) != 1 || got[0]["id"] != "acct1" || got[0]["name"] != "Main Account" {
		t.Fatalf("records = %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := spotifyads.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil || cat.Connector != "spotify-ads" || len(cat.Streams) < 3 {
		t.Fatalf("Catalog = %+v err=%v", cat, err)
	}
	var fixture []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Config: cfg}, func(rec connectors.Record) error {
		fixture = append(fixture, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(fixture) == 0 || fixture[0]["fixture"] != true {
		t.Fatalf("fixture records = %+v", fixture)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write err = %v", err)
	}
	if got, ok := connectors.NewRegistry().Get("spotify-ads"); !ok || got.Metadata().Capabilities.Write {
		t.Fatalf("registry/capabilities failed: ok=%v got=%T", ok, got)
	}
}
