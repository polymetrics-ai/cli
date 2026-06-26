package vismaeconomic_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	vismaeconomic "polymetrics.ai/internal/connectors/visma-economic"
)

func TestReadCustomersSendsEconomicHeadersAndMapsCollection(t *testing.T) {
	var sawApp, sawGrant string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawApp = r.Header.Get("X-AppSecretToken")
		sawGrant = r.Header.Get("X-AgreementGrantToken")
		if r.URL.Path != "/customers" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"collection":[{"customerNumber":1,"name":"Acme Co","currency":"DKK"}]}`))
	}))
	defer srv.Close()

	c := vismaeconomic.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}, Secrets: map[string]string{"app_secret_token": "dummy-app", "agreement_grant_token": "dummy-grant"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawApp != "dummy-app" || sawGrant != "dummy-grant" {
		t.Fatalf("auth headers = %q/%q", sawApp, sawGrant)
	}
	if len(got) != 1 || got[0]["id"] != "1" || got[0]["name"] != "Acme Co" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := vismaeconomic.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "visma-economic" || len(cat.Streams) == 0 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	if _, ok := connectors.NewRegistry().Get("visma-economic"); !ok {
		t.Fatal("registry did not resolve visma-economic")
	}
	if caps := c.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want read-only Check/Catalog/Read", caps)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
