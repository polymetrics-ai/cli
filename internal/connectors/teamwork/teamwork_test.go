package teamwork_test

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/teamwork"
)

func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/projects.json" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"projects":[{"id":"1","name":"Migration","created-on":"2026-01-01T00:00:00Z"},{"id":"2","name":"Launch","created-on":"2026-01-02T00:00:00Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"projects":[{"id":"3","name":"Retention","created-on":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := teamwork.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "username": "user@example.com", "page_size": "2"}, Secrets: map[string]string{"password": "test_password"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user@example.com:test_password"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if len(got) != 3 || got[0]["id"] != "1" || got[2]["name"] != "Retention" {
		t.Fatalf("records not paginated/mapped: %+v", got)
	}
}

func TestFixtureModeNoCredentials(t *testing.T) {
	c := teamwork.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var count int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: cfg}, func(rec connectors.Record) error {
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
	c := teamwork.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "teamwork" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v", cat)
	}
	if caps := c.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v", caps)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
	if _, ok := connectors.NewRegistry().Get("teamwork"); !ok {
		t.Fatal("registry did not resolve teamwork")
	}
}
