package pandadoc_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/pandadoc"
)

func TestReadPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var srv *httptest.Server
	var sawPage string
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/documents" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"results":[{"id":"doc_1","name":"Proposal","status":"document.draft"},{"id":"doc_2","name":"Order","status":"document.sent"}],"next":"` + srv.URL + `/documents?page=2&count=2"}`))
		case "2":
			sawPage = "2"
			_, _ = w.Write([]byte(`{"results":[{"id":"doc_3","name":"Invoice","status":"document.completed"}],"next":null}`))
		default:
			t.Fatalf("unexpected page query %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := pandadoc.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "count": "2"}, Secrets: map[string]string{"api_key": "panda_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "documents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "API-Key panda_key" {
		t.Fatalf("Authorization = %q, want PandaDoc API-Key auth", sawAuth)
	}
	if sawPage != "2" {
		t.Fatalf("second page = %q, want 2", sawPage)
	}
	if len(got) != 3 || got[0]["id"] != "doc_1" || got[2]["status"] != "document.completed" {
		t.Fatalf("mapped records = %+v, want three documents", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := pandadoc.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "templates", Config: fixture}, func(rec connectors.Record) error {
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
	if cat.Connector != "pandadoc" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want pandadoc streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("pandadoc"); !ok {
		t.Fatal("registry did not resolve pandadoc")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
