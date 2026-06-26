package trackpms_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	trackpms "polymetrics.ai/internal/connectors/track-pms"
)

func TestReadReservationsAuthenticatesAndMaps(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/reservations" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer fixture-token" {
			t.Fatal("authorization header was not applied")
		}
		_, _ = w.Write([]byte(`{"reservations":[{"id":"res_1","confirmation_number":"CN-1","status":"booked","arrival_date":"2026-01-02"}]}`))
	}))
	defer srv.Close()

	c := trackpms.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1"}, Secrets: map[string]string{"access_token": "fixture-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reservations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "res_1" || got[0]["status"] != "booked" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := trackpms.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "track-pms" || len(cat.Streams) < 4 {
		t.Fatalf("catalog = %+v", cat)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "guests", Config: cfg}, func(rec connectors.Record) error {
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
	if _, ok := connectors.NewRegistry().Get("track-pms"); !ok {
		t.Fatal("track-pms was not self-registered")
	}
}
