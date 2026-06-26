package sigmacomputing

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadWorkbooksPaginatesAndAuthenticates(t *testing.T) {
	authFailures := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			authFailures++
		}
		if r.URL.Path != "/v2/workbooks" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("limit") != "2" {
			t.Fatalf("limit query = %q, want 2", r.URL.Query().Get("limit"))
		}
		switch r.URL.Query().Get("page") {
		case "":
			_, _ = w.Write([]byte(`{"entries":[{"id":"wb_1","name":"Revenue","updatedAt":"2026-01-01T00:00:00Z"},{"id":"wb_2","name":"Sales","updatedAt":"2026-01-02T00:00:00Z"}],"nextPage":"cursor-2"}`))
		case "cursor-2":
			_, _ = w.Write([]byte(`{"entries":[{"id":"wb_3","name":"Ops","updatedAt":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Fatalf("unexpected page query %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "2"}, Secrets: map[string]string{"access_token": "test-token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "workbooks", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if authFailures != 0 {
		t.Fatal("bearer token was not applied to every request")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
	if got[0]["id"] != "wb_1" || got[0]["name"] != "Revenue" {
		t.Fatalf("first record = %+v", got[0])
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	for _, stream := range []string{"workbooks", "datasets", "teams", "members"} {
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
	if cat.Connector != "sigma-computing" || len(cat.Streams) < 4 {
		t.Fatalf("catalog = %+v", cat)
	}
	got, ok := connectors.NewRegistry().Get("sigma-computing")
	if !ok {
		t.Fatal("registry did not resolve sigma-computing")
	}
	if caps := got.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v", caps)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want unsupported", err)
	}
}
