package convex_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/convex"
)

func TestReadDocumentsPaginatesAuthenticatesAndMapsRecords(t *testing.T) {
	var sawAuth string
	var cursors []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/tables/messages/documents" {
			http.NotFound(w, r)
			return
		}
		cursor := r.URL.Query().Get("cursor")
		cursors = append(cursors, cursor)
		switch cursor {
		case "":
			_, _ = w.Write([]byte(`{"documents":[{"_id":"doc_1","text":"hello"},{"_id":"doc_2","text":"world"}],"cursor":"next_1"}`))
		case "next_1":
			_, _ = w.Write([]byte(`{"documents":[{"_id":"doc_3","text":"again"}],"cursor":""}`))
		default:
			t.Fatalf("unexpected cursor %q", cursor)
		}
	}))
	defer srv.Close()

	c := convex.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "table": "messages"}, Secrets: map[string]string{"access_key": "convex_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "documents", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer convex_key" || len(cursors) != 2 {
		t.Fatalf("auth/pages wrong auth=%q cursors=%v", sawAuth, cursors)
	}
	if len(got) != 3 || got[0]["id"] != "doc_1" || got[0]["text"] != "hello" {
		t.Fatalf("records mapped wrong: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := convex.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"tables", "documents"} {
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
	if err != nil || cat.Connector != "convex" || len(cat.Streams) != 2 {
		t.Fatalf("Catalog = %+v err=%v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("convex"); !ok {
		t.Fatal("registry did not resolve convex")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
