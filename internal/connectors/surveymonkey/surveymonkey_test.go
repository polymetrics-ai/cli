package surveymonkey_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/surveymonkey"
)

func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/surveys" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = fmt.Fprintf(w, `{"data":[{"id":"s1","title":"NPS","date_created":"2026-01-01T00:00:00Z"}],"links":{"next":"%s/surveys?page=2"}}`, srv.URL)
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"s2","title":"CSAT","date_created":"2026-01-02T00:00:00Z"}],"links":{}}`))
		default:
			t.Fatalf("unexpected page=%q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := surveymonkey.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1"}, Secrets: map[string]string{"access_token": "test-access-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "surveys", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer test-access-token" {
		t.Fatalf("Authorization = %q, want bearer test token", sawAuth)
	}
	if len(got) != 2 || got[0]["id"] != "s1" || got[0]["title"] != "NPS" {
		t.Fatalf("records not paginated/mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := surveymonkey.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "surveymonkey" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v, want surveymonkey streams", cat)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "surveys", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	if _, ok := connectors.NewRegistry().Get("surveymonkey"); !ok {
		t.Fatal("registry did not resolve surveymonkey")
	}
	_, err = c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, []connectors.Record{{"id": "1"}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
