package ubidots_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/ubidots"
)

func TestReadDevicesAuthenticatesAndMaps(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2.0/devices/" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("X-Auth-Token") != "fixture-token" {
			t.Fatal("token header was not applied")
		}
		_, _ = w.Write([]byte(`{"results":[{"id":"dev_1","label":"plant-1","name":"Plant 1","created_at":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := ubidots.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}, Secrets: map[string]string{"token": "fixture-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "devices", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "dev_1" || got[0]["label"] != "plant-1" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := ubidots.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "ubidots" || len(cat.Streams) < 4 {
		t.Fatalf("catalog = %+v", cat)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "variables", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["fixture"] != true || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v", got)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, []connectors.Record{{"id": "x"}}); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
	if caps := c.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v", caps)
	}
	if _, ok := connectors.NewRegistry().Get("ubidots"); !ok {
		t.Fatal("ubidots was not self-registered")
	}
}
