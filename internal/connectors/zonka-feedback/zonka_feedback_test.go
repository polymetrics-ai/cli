package zonkafeedback_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	zonkafeedback "polymetrics.ai/internal/connectors/zonka-feedback"
)

func TestReadResponsesAuthenticatesAndPaginates(t *testing.T) {
	var sawAuth string
	var sawSecondPage bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/zonka/responses" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"responses":[{"id":"resp_1","survey_id":"survey_1","rating":9,"updated_at":"2026-01-02T00:00:00Z"},{"id":"resp_2","survey_id":"survey_1","rating":8,"updated_at":"2026-01-03T00:00:00Z"}]}`))
		case "2":
			sawSecondPage = true
			_, _ = w.Write([]byte(`{"responses":[{"id":"resp_3","survey_id":"survey_2","rating":7,"updated_at":"2026-01-04T00:00:00Z"}]}`))
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := zonkafeedback.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/zonka", "page_size": "2"}, Secrets: map[string]string{"auth_token": "test_auth_token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "responses", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer test_auth_token" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if !sawSecondPage {
		t.Fatal("expected second page request")
	}
	if len(got) != 3 || got[0]["id"] != "resp_1" || got[2]["rating"] == nil {
		t.Fatalf("records = %+v, want three responses", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := zonkafeedback.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "surveys", Config: fixture}, func(rec connectors.Record) error {
		rows = append(rows, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(rows) == 0 || rows[0]["id"] == nil {
		t.Fatalf("fixture rows = %+v, want records with id", rows)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "zonka-feedback" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want zonka-feedback streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("zonka-feedback"); !ok {
		t.Fatal("registry did not resolve zonka-feedback")
	}
	if c.Metadata().Capabilities.Write {
		t.Fatal("zonka-feedback should be read-only")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
