package tvmazeschedule_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	tvmazeschedule "polymetrics.ai/internal/connectors/tvmaze-schedule"
)

func TestReadSchedulePublicAPIMaps(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/schedule" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("country") != "US" || r.URL.Query().Get("date") != "2026-01-01" {
			t.Fatalf("schedule query = %q", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`[{"id":101,"name":"Pilot","airdate":"2026-01-01","airtime":"20:00","show":{"id":1,"name":"Fixture Show"}}]`))
	}))
	defer srv.Close()

	c := tvmazeschedule.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "country": "US", "date": "2026-01-01"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "schedule", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["id"] == nil || got[0]["show_name"] != "Fixture Show" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := tvmazeschedule.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "tvmaze-schedule" || len(cat.Streams) < 2 {
		t.Fatalf("catalog = %+v", cat)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "web_schedule", Config: cfg}, func(rec connectors.Record) error {
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
	if _, ok := connectors.NewRegistry().Get("tvmaze-schedule"); !ok {
		t.Fatal("tvmaze-schedule was not self-registered")
	}
}
