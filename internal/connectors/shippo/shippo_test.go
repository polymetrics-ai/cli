package shippo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadAddressesPaginatesAndAuthenticates(t *testing.T) {
	authFailures := 0
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "ShippoToken test-token" {
			authFailures++
		}
		if r.URL.Path != "/addresses" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			if r.URL.Query().Get("results") != "2" {
				t.Fatalf("results query = %q, want 2", r.URL.Query().Get("results"))
			}
			_, _ = w.Write([]byte(`{"results":[{"object_id":"addr_1","name":"Alice","updated_at":"2026-01-01T00:00:00Z"},{"object_id":"addr_2","name":"Bob","updated_at":"2026-01-02T00:00:00Z"}],"next":"` + srv.URL + `/addresses?page=2"}`))
		case "2":
			_, _ = w.Write([]byte(`{"results":[{"object_id":"addr_3","name":"Carol","updated_at":"2026-01-03T00:00:00Z"}],"next":null}`))
		default:
			t.Fatalf("unexpected page query %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "2"}, Secrets: map[string]string{"api_token": "test-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "addresses", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if authFailures != 0 {
		t.Fatal("auth header was not applied to every request")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
	if got[0]["id"] != "addr_1" || got[0]["name"] != "Alice" {
		t.Fatalf("first record = %+v", got[0])
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	for _, stream := range []string{"addresses", "parcels", "shipments", "transactions"} {
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		}); err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 || got[0]["id"] == nil {
			t.Fatalf("fixture records for %s = %+v", stream, got)
		}
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "shippo" || len(cat.Streams) < 4 {
		t.Fatalf("catalog = %+v", cat)
	}
	got, ok := connectors.NewRegistry().Get("shippo")
	if !ok {
		t.Fatal("registry did not resolve shippo")
	}
	if caps := got.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v", caps)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want unsupported", err)
	}
}
