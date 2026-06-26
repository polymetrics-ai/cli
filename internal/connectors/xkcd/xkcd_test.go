package xkcd_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/xkcd"
)

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := xkcd.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "xkcd" || len(cat.Streams) < 2 {
		t.Fatalf("catalog = %+v, want xkcd streams", cat)
	}
	var rows []connectors.Record
	err = c.Read(context.Background(), connectors.ReadRequest{Stream: "latest", Config: fixture}, func(rec connectors.Record) error {
		rows = append(rows, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(rows) == 0 || rows[0]["fixture"] != true || rows[0]["num"] == nil {
		t.Fatalf("fixture rows = %+v, want fixture records with num", rows)
	}
	if _, ok := connectors.NewRegistry().Get("xkcd"); !ok {
		t.Fatal("registry did not resolve xkcd")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}

func TestReadLiveLatestUsesPublicJSONEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/info.0.json" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"num":42,"title":"Geography","safe_title":"Geography","year":"2006","month":"1","day":"1"}`))
	}))
	defer srv.Close()

	c := xkcd.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "latest", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["num"] != float64(42) {
		t.Fatalf("records = %+v, want comic 42", got)
	}
}
