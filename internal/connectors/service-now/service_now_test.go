package servicenow_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	servicenow "polymetrics.ai/internal/connectors/service-now"
)

func TestContractFixtureAndWrite(t *testing.T) {
	c := servicenow.New()
	if c.Name() != "service-now" {
		t.Fatalf("Name() = %q, want service-now", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want read-only Check/Catalog/Read", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) == 0 || cat.Streams[0].Name != "incidents" {
		t.Fatalf("catalog streams = %+v, want incidents first", cat.Streams)
	}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "incidents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 || got[0]["sys_id"] == nil {
		t.Fatalf("fixture records = %+v, want sys_id", got)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("service-now"); !ok {
		t.Fatal("registry did not resolve service-now")
	}
}

func TestReadIncidentsUsesBasicAuthAndResultKey(t *testing.T) {
	var sawAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/now/table/incident" {
			http.NotFound(w, r)
			return
		}
		user, pass, ok := r.BasicAuth()
		sawAuth = ok && user == "test-user" && pass == "test-pass"
		if r.URL.Query().Get("sysparm_limit") != "1" {
			t.Fatal("sysparm_limit query was not forwarded")
		}
		_, _ = w.Write([]byte(`{"result":[{"sys_id":"inc_1","number":"INC001","updated_on":"2026-01-01 00:00:00"}]}`))
	}))
	defer srv.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "username": "test-user", "page_size": "1", "max_pages": "1"}, Secrets: map[string]string{"password": "test-pass"}}
	var got []connectors.Record
	if err := servicenow.New().Read(context.Background(), connectors.ReadRequest{Stream: "incidents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawAuth {
		t.Fatal("basic auth was not applied")
	}
	if len(got) != 1 || got[0]["number"] != "INC001" {
		t.Fatalf("records = %+v, want incident number", got)
	}
}
