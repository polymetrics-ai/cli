package persona_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/persona"
)

func TestReadInquiriesPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var requests int
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		requests++
		if r.URL.Path != "/inquiries" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page[after]") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"inq_1","type":"inquiry","attributes":{"status":"pending","created-at":"2026-01-01T00:00:00Z"}}],"links":{"next":"` + srv.URL + `/inquiries?page%5Bafter%5D=inq_1","prev":null}}`))
		case "inq_1":
			_, _ = w.Write([]byte(`{"data":[{"id":"inq_2","type":"inquiry","attributes":{"status":"approved","created-at":"2026-01-02T00:00:00Z"}}],"links":{"next":null,"prev":null}}`))
		default:
			t.Fatalf("unexpected cursor %q", r.URL.Query().Get("page[after]"))
		}
	}))
	defer srv.Close()

	c := persona.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1"}, Secrets: map[string]string{"api_key": "persona_key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "inquiries", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer persona_key" {
		t.Fatalf("Authorization = %q, want Bearer persona_key", sawAuth)
	}
	if len(got) != 2 || got[0]["id"] != "inq_1" || got[1]["id"] != "inq_2" {
		t.Fatalf("records = %+v, want two JSON:API records", got)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := persona.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "inquiries", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "persona" || len(cat.Streams) < 2 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("persona"); !ok {
		t.Fatal("registry did not resolve persona")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
