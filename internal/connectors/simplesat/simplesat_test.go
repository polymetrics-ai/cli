package simplesat_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/simplesat"
)

func TestReadAnswersUsesTokenHeader(t *testing.T) {
	var sawToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.Header.Get("X-Simplesat-Token")
		if r.URL.Path != "/answers/" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("page_size") != "2" {
			t.Fatalf("page_size = %q, want 2", r.URL.Query().Get("page_size"))
		}
		_, _ = w.Write([]byte(`{"results":[{"id":101,"rating":5},{"id":102,"rating":4}]}`))
	}))
	defer srv.Close()

	c := simplesat.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "2"}, Secrets: map[string]string{"api_key": "fixture-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "answers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "fixture-token" {
		t.Fatalf("X-Simplesat-Token was not set from api_key")
	}
	if len(got) != 2 || got[0]["id"] == nil {
		t.Fatalf("records = %+v, want two answer records", got)
	}
}

func TestFixtureRegistryCatalogAndWrite(t *testing.T) {
	c := simplesat.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "simplesat" || len(cat.Streams) == 0 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if _, ok := connectors.NewRegistry().Get("simplesat"); !ok {
		t.Fatal("registry did not resolve simplesat")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
