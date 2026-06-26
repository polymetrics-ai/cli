package perk_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/perk"
)

func TestReadTripsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth, sawVersion string
	var offsets []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawVersion = r.Header.Get("Api-Version")
		offsets = append(offsets, r.URL.Query().Get("offset"))
		if r.URL.Path != "/trips" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "0":
			_, _ = w.Write([]byte(`{"trips":[{"id":"t1","trip_name":"Launch","modified":"2026-01-01T00:00:00Z"}],"limit":1,"offset":0,"total":2}`))
		case "1":
			_, _ = w.Write([]byte(`{"trips":[{"id":"t2","trip_name":"Return","modified":"2026-01-02T00:00:00Z"}],"limit":1,"offset":1,"total":2}`))
		case "2":
			_, _ = w.Write([]byte(`{"trips":[],"limit":1,"offset":2,"total":2}`))
		default:
			t.Fatalf("unexpected offset %q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	c := perk.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1"}, Secrets: map[string]string{"api_key": "perk_key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "trips", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "ApiKey perk_key" || sawVersion != "1" {
		t.Fatalf("auth headers = %q / %q, want ApiKey and Api-Version", sawAuth, sawVersion)
	}
	if len(got) != 2 || got[1]["trip_name"] != "Return" {
		t.Fatalf("records = %+v, want two mapped trips", got)
	}
	if len(offsets) != 3 {
		t.Fatalf("offsets = %v, want three requests including empty stop page", offsets)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := perk.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "trips", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "perk" || len(cat.Streams) < 2 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("perk"); !ok {
		t.Fatal("registry did not resolve perk")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
