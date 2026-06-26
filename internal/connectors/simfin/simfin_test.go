package simfin

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadCompaniesPaginatesAndAuthenticates(t *testing.T) {
	authFailures := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("api-key") != "test-token" {
			authFailures++
		}
		if r.URL.Path != "/api/v3/companies/list" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("limit") != "2" {
			t.Fatalf("limit query = %q, want 2", r.URL.Query().Get("limit"))
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"data":[{"id":"sim_1","name":"Acme Corp","ticker":"ACME"},{"id":"sim_2","name":"Beta Corp","ticker":"BETA"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"sim_3","name":"Gamma Corp","ticker":"GAM"}]}`))
		default:
			t.Fatalf("unexpected page query %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "2"}, Secrets: map[string]string{"api_key": "test-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "companies", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if authFailures != 0 {
		t.Fatal("api-key query auth was not applied to every request")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
	if got[0]["id"] != "sim_1" || got[0]["ticker"] != "ACME" {
		t.Fatalf("first record = %+v", got[0])
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	for _, stream := range []string{"companies", "statements", "markets"} {
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
	if cat.Connector != "simfin" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v", cat)
	}
	got, ok := connectors.NewRegistry().Get("simfin")
	if !ok {
		t.Fatal("registry did not resolve simfin")
	}
	if caps := got.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v", caps)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want unsupported", err)
	}
}
