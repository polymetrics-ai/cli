package publicapis_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	publicapis "polymetrics.ai/internal/connectors/public-apis"
)

func TestReadEntriesPaginatesWithOffsetAndMaps(t *testing.T) {
	var offsets []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/entries" {
			http.NotFound(w, r)
			return
		}
		offset := r.URL.Query().Get("offset")
		if offset == "" {
			offset = "0"
		}
		offsets = append(offsets, offset)
		switch offset {
		case "0":
			_, _ = w.Write([]byte(`{"count":3,"entries":[{"API":"Cat Facts","Description":"Daily cat facts","Category":"Animals"},{"API":"Dog Facts","Description":"Daily dog facts","Category":"Animals"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"count":3,"entries":[{"API":"Numbers","Description":"Number facts","Category":"Math"}]}`))
		default:
			t.Fatalf("unexpected offset %q", offset)
		}
	}))
	defer srv.Close()

	c := publicapis.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "2"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "entries", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(offsets) != 2 || offsets[0] != "0" || offsets[1] != "2" {
		t.Fatalf("offsets = %v", offsets)
	}
	if len(got) != 3 || got[0]["id"] != "Cat Facts" || got[2]["category"] != "Math" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := publicapis.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "entries", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "public-apis" || len(cat.Streams) < 2 {
		t.Fatalf("unexpected catalog: %+v", cat)
	}
	for _, stream := range cat.Streams {
		if len(stream.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", stream.Name)
		}
	}
	if _, ok := connectors.NewRegistry().Get("public-apis"); !ok {
		t.Fatal("registry did not resolve public-apis")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
