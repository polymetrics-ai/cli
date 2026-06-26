package thinkificcourses_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	thinkificcourses "polymetrics.ai/internal/connectors/thinkific-courses"
)

func TestReadCoursesPaginatesAndAuthenticates(t *testing.T) {
	var sawKey, sawSubdomain string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("X-Auth-API-Key")
		sawSubdomain = r.Header.Get("X-Auth-Subdomain")
		if r.URL.Path != "/courses" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`{"items":[{"id":1,"name":"Intro","slug":"intro"},{"id":2,"name":"Deep Dive","slug":"deep-dive"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"items":[{"id":3,"name":"Wrap Up","slug":"wrap-up"}]}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"items":[]}`))
		}
	}))
	defer srv.Close()

	c := thinkificcourses.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "X-Auth-Subdomain": "academy", "page_size": "2"},
		Secrets: map[string]string{"api_key": "test-api-key"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "courses", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "test-api-key" || sawSubdomain != "academy" {
		t.Fatalf("auth headers = key %q subdomain %q", sawKey, sawSubdomain)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
	if got[0]["id"] == nil || got[0]["name"] == nil || got[0]["slug"] == nil {
		t.Fatalf("record missing course fields: %+v", got[0])
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := thinkificcourses.New()
	assertConnector(t, c, "thinkific-courses")
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
