package ticketmaster_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/ticketmaster"
)

func TestReadEventsPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("apikey")
		if r.URL.Path != "/events.json" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "0":
			_, _ = w.Write([]byte(`{"_embedded":{"events":[{"id":"e1","name":"Night One"},{"id":"e2","name":"Night Two"}]}}`))
		case "1":
			_, _ = w.Write([]byte(`{"_embedded":{"events":[{"id":"e3","name":"Finale"}]}}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"_embedded":{"events":[]}}`))
		}
	}))
	defer srv.Close()

	c := ticketmaster.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "2"}, Secrets: map[string]string{"api_key": "tm-key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "tm-key" {
		t.Fatalf("apikey query = %q", sawKey)
	}
	if len(got) != 3 || got[0]["name"] == nil {
		t.Fatalf("records = %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := ticketmaster.New()
	assertConnector(t, c, "ticketmaster")
}

func assertConnector(t *testing.T, c connectors.Connector, name string) {
	t.Helper()
	if c.Name() != name {
		t.Fatalf("Name = %q, want %q", c.Name(), name)
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want check/catalog/read and no write", caps)
	}
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), fixture)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != name || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v", cat)
	}
	for _, stream := range cat.Streams {
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream.Name, Config: fixture}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		}); err != nil {
			t.Fatalf("fixture Read(%s): %v", stream.Name, err)
		}
		if len(got) == 0 || got[0]["id"] == nil {
			t.Fatalf("fixture %s records = %+v", stream.Name, got)
		}
	}
	result, err := c.Write(context.Background(), connectors.WriteRequest{Config: fixture}, []connectors.Record{{"id": "1"}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) || result.RecordsFailed != 1 {
		t.Fatalf("Write = (%+v, %v), want unsupported with one failed record", result, err)
	}
	if _, ok := connectors.NewRegistry().Get(name); !ok {
		t.Fatalf("registry did not resolve %s", name)
	}
}
