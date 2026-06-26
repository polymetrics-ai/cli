package solarwindsservicedesk_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	solarwindsservicedesk "polymetrics.ai/internal/connectors/solarwinds-service-desk"
)

func TestReadIncidentsUsesBearerAndVendorAccept(t *testing.T) {
	var sawAuth string
	var sawAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawAccept = r.Header.Get("Accept")
		if r.URL.Path != "/incidents.json" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":10,"name":"Laptop issue"}]`))
	}))
	defer srv.Close()

	c := solarwindsservicedesk.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}, Secrets: map[string]string{"api_key_2": "fixture-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "incidents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer fixture-token" || sawAccept != "application/vnd.samanage.v1.1+json" {
		t.Fatalf("auth/accept headers not set as expected")
	}
	if len(got) != 1 || got[0]["id"] == nil {
		t.Fatalf("records = %+v, want incident record", got)
	}
}

func TestFixtureRegistryCatalogAndWrite(t *testing.T) {
	c := solarwindsservicedesk.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "solarwinds-service-desk" || len(cat.Streams) == 0 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if _, ok := connectors.NewRegistry().Get("solarwinds-service-desk"); !ok {
		t.Fatal("registry did not resolve solarwinds-service-desk")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
