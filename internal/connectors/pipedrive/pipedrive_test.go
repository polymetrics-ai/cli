package pipedrive_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/pipedrive"
)

func TestReadDealsPaginatesAndAuthenticates(t *testing.T) {
	var sawToken string
	var starts []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.URL.Query().Get("api_token")
		starts = append(starts, r.URL.Query().Get("start"))
		if r.URL.Path != "/deals" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("start") {
		case "0":
			_, _ = w.Write([]byte(`{"success":true,"data":[{"id":1,"title":"First deal","update_time":"2026-01-01 00:00:00"}],"additional_data":{"pagination":{"more_items_in_collection":true,"next_start":1}}}`))
		case "1":
			_, _ = w.Write([]byte(`{"success":true,"data":[{"id":2,"title":"Second deal","update_time":"2026-01-02 00:00:00"}],"additional_data":{"pagination":{"more_items_in_collection":false,"next_start":null}}}`))
		default:
			t.Fatalf("unexpected start %q", r.URL.Query().Get("start"))
		}
	}))
	defer srv.Close()

	c := pipedrive.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1"}, Secrets: map[string]string{"api_token": "pipe_key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "deals", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "pipe_key" {
		t.Fatalf("api_token = %q, want pipe_key", sawToken)
	}
	if len(got) != 2 || got[0]["id"] == nil || got[1]["title"] != "Second deal" {
		t.Fatalf("records = %+v, want mapped deals", got)
	}
	if len(starts) != 2 {
		t.Fatalf("starts = %v, want two requests", starts)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := pipedrive.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "deals", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "pipedrive" || len(cat.Streams) < 3 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("pipedrive"); !ok {
		t.Fatal("registry did not resolve pipedrive")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
