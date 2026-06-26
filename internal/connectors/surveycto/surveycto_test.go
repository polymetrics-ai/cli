package surveycto_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/surveycto"
)

func TestReadFormsAuthenticatesWithBasicAndMapsRecords(t *testing.T) {
	var sawUser string
	var sawBasic bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		sawUser = user
		sawBasic = ok && pass == "test-password"
		if r.URL.Path != "/forms" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"forms":[{"id":"customer","title":"Customer Survey","version":"1"}]}`))
	}))
	defer srv.Close()

	c := surveycto.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}, Secrets: map[string]string{"username": "test-user", "password": "test-password"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "forms", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawBasic || sawUser != "test-user" {
		t.Fatalf("basic auth missing or wrong user: ok=%v user=%q", sawBasic, sawUser)
	}
	if len(got) != 1 || got[0]["id"] != "customer" || got[0]["title"] != "Customer Survey" {
		t.Fatalf("records = %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := surveycto.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil || cat.Connector != "surveycto" || len(cat.Streams) < 3 {
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
	if got, ok := connectors.NewRegistry().Get("surveycto"); !ok || got.Metadata().Capabilities.Write {
		t.Fatalf("registry/capabilities failed: ok=%v got=%T", ok, got)
	}
}
