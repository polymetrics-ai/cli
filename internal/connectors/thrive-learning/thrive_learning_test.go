package thrivelearning_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	thrivelearning "polymetrics.ai/internal/connectors/thrive-learning"
)

func TestReadUsersPaginatesAndAuthenticates(t *testing.T) {
	var sawUser, sawPassword string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawUser, sawPassword, _ = r.BasicAuth()
		if r.URL.Path != "/users" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("updated_since") != "2026-01-01T00:00:00Z" {
			t.Fatalf("updated_since = %q", r.URL.Query().Get("updated_since"))
		}
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`{"items":[{"id":"u1","email":"a@example.com","name":"Ada"},{"id":"u2","email":"b@example.com","name":"Ben"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"items":[{"id":"u3","email":"c@example.com","name":"Cy"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"items":[]}`))
		}
	}))
	defer srv.Close()

	c := thrivelearning.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "tenant-1", "password": "test-password", "start_date": "2026-01-01T00:00:00Z", "page_size": "2"},
		Secrets: map[string]string{"password": "test-password"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawUser != "tenant-1" || sawPassword != "test-password" {
		t.Fatalf("basic auth = %q/%q", sawUser, sawPassword)
	}
	if len(got) != 3 || got[0]["email"] == nil {
		t.Fatalf("records = %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := thrivelearning.New()
	assertConnector(t, c, "thrive-learning")
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
