package elasticsearch_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/elasticsearch"
)

func TestReadDocumentsPaginatesAuthenticatesAndMapsRecords(t *testing.T) {
	var sawUser string
	var froms []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "elastic" || pass != "password" {
			t.Fatalf("unexpected basic auth user=%q ok=%v", user, ok)
		}
		sawUser = user
		if r.URL.Path != "/orders/_search" {
			http.NotFound(w, r)
			return
		}
		from := r.URL.Query().Get("from")
		froms = append(froms, from)
		switch from {
		case "", "0":
			_, _ = w.Write([]byte(`{"hits":{"hits":[{"_id":"doc_1","_source":{"order_number":"1001"}},{"_id":"doc_2","_source":{"order_number":"1002"}}]}}`))
		case "2":
			_, _ = w.Write([]byte(`{"hits":{"hits":[{"_id":"doc_3","_source":{"order_number":"1003"}}]}}`))
		default:
			t.Fatalf("unexpected from %q", from)
		}
	}))
	defer srv.Close()

	c := elasticsearch.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"endpoint": srv.URL, "index": "orders", "username": "elastic", "page_size": "2"}, Secrets: map[string]string{"password": "password"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "documents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawUser != "elastic" || len(froms) != 2 {
		t.Fatalf("auth/pages wrong user=%q froms=%v", sawUser, froms)
	}
	if len(got) != 3 || got[0]["id"] != "doc_1" || got[0]["order_number"] != "1001" {
		t.Fatalf("records mapped wrong: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := elasticsearch.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"indices", "documents"} {
		count := 0
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(connectors.Record) error { count++; return nil }); err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if count == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "elasticsearch" || len(cat.Streams) != 2 {
		t.Fatalf("Catalog = %+v err=%v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("elasticsearch"); !ok {
		t.Fatal("registry did not resolve elasticsearch")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
