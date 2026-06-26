package waiteraid_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/waiteraid"
)

func TestReadReservationsSendsAuthHeadersAndMaps(t *testing.T) {
	var sawAuth, sawRestaurant string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("X-Auth-Hash")
		sawRestaurant = r.Header.Get("X-Restaurant-ID")
		if r.URL.Path != "/reservations" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("start_date") != "2026-01-01" {
			t.Fatalf("start_date query = %q", r.URL.Query().Get("start_date"))
		}
		_, _ = w.Write([]byte(`{"reservations":[{"id":"res_1","guest_name":"Ada","date":"2026-01-02","status":"confirmed"}]}`))
	}))
	defer srv.Close()

	c := waiteraid.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "start_date": "2026-01-01"}, Secrets: map[string]string{"auth_hash": "dummy-auth", "restid": "dummy-restaurant"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reservations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "dummy-auth" || sawRestaurant != "dummy-restaurant" {
		t.Fatalf("auth headers = %q/%q", sawAuth, sawRestaurant)
	}
	if len(got) != 1 || got[0]["id"] != "res_1" || got[0]["guest_name"] != "Ada" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := waiteraid.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reservations", Config: cfg}, func(rec connectors.Record) error {
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
	if cat.Connector != "waiteraid" || len(cat.Streams) == 0 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	if _, ok := connectors.NewRegistry().Get("waiteraid"); !ok {
		t.Fatal("registry did not resolve waiteraid")
	}
	if caps := c.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want read-only Check/Catalog/Read", caps)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
