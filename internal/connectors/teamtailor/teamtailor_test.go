package teamtailor_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/teamtailor"
)

func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawVersion = r.Header.Get("X-Api-Version")
		if r.URL.Path != "/jobs" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page[number]") {
		case "", "1":
			_, _ = w.Write([]byte(`{"data":[{"id":"job_1","type":"jobs","attributes":{"title":"Engineer","created-at":"2026-01-01T00:00:00Z"}},{"id":"job_2","type":"jobs","attributes":{"title":"Designer","created-at":"2026-01-02T00:00:00Z"}}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"job_3","type":"jobs","attributes":{"title":"Manager","created-at":"2026-01-03T00:00:00Z"}}]}`))
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page[number]"))
		}
	}))
	defer srv.Close()

	c := teamtailor.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "x_api_version": "20240404", "page_size": "2", "api": "test_api_key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "jobs", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Token token=test_api_key" {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if sawVersion != "20240404" {
		t.Fatalf("X-Api-Version = %q", sawVersion)
	}
	if len(got) != 3 || got[0]["id"] != "job_1" || got[2]["title"] != "Manager" {
		t.Fatalf("records not paginated/mapped: %+v", got)
	}
}

func TestFixtureModeNoCredentials(t *testing.T) {
	c := teamtailor.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var count int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "jobs", Config: cfg}, func(rec connectors.Record) error {
		count++
		if rec["id"] == nil {
			t.Fatalf("fixture missing id: %+v", rec)
		}
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if count == 0 {
		t.Fatal("fixture emitted no records")
	}
}

func TestCatalogRegistrationAndReadOnly(t *testing.T) {
	c := teamtailor.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "teamtailor" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v", cat)
	}
	if caps := c.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v", caps)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
	if _, ok := connectors.NewRegistry().Get("teamtailor"); !ok {
		t.Fatal("registry did not resolve teamtailor")
	}
}
