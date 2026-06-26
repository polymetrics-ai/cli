package toggl_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/toggl"
)

func TestReadTimeEntriesAuthenticatesAndFiltersDates(t *testing.T) {
	var sawUser, sawPassword string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawUser, sawPassword, _ = r.BasicAuth()
		if r.URL.Path != "/me/time_entries" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("start_date") != "2026-01-01" || r.URL.Query().Get("end_date") != "2026-01-31" {
			t.Fatalf("date filters = %q/%q", r.URL.Query().Get("start_date"), r.URL.Query().Get("end_date"))
		}
		_, _ = w.Write([]byte(`[{"id":1,"description":"Build","workspace_id":9},{"id":2,"description":"Test","workspace_id":9}]`))
	}))
	defer srv.Close()

	c := toggl.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "workspace_id": "9", "organization_id": "4", "start_date": "2026-01-01", "end_date": "2026-01-31"}, Secrets: map[string]string{"api_token": "toggl-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "time_entries", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawUser != "toggl-token" || sawPassword != "api_token" {
		t.Fatalf("basic auth = %q/%q", sawUser, sawPassword)
	}
	if len(got) != 2 || got[0]["description"] == nil {
		t.Fatalf("records = %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := toggl.New()
	assertConnector(t, c, "toggl")
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
