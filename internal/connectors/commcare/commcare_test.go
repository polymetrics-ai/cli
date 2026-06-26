package commcare_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commcare"
)

func TestReadFormsPaginatesAuthenticatesAndMapsRecords(t *testing.T) {
	var sawAuth string
	var offsets []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/a/demo/api/v0.5/form/" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("app_id") != "app_1" {
			t.Fatalf("app_id query = %q", r.URL.Query().Get("app_id"))
		}
		offset := r.URL.Query().Get("offset")
		offsets = append(offsets, offset)
		switch offset {
		case "", "0":
			_, _ = w.Write([]byte(`{"objects":[{"id":"form_1","received_on":"2026-01-01T00:00:00Z"},{"id":"form_2","received_on":"2026-01-02T00:00:00Z"}],"meta":{"next":"/a/demo/api/v0.5/form/?offset=2&limit=2&app_id=app_1"}}`))
		case "2":
			_, _ = w.Write([]byte(`{"objects":[{"id":"form_3","received_on":"2026-01-03T00:00:00Z"}],"meta":{"next":null}}`))
		default:
			t.Fatalf("unexpected offset %q", offset)
		}
	}))
	defer srv.Close()

	c := commcare.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "project_space": "demo", "app_id": "app_1", "page_size": "2"}, Secrets: map[string]string{"api_key": "cc_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "forms", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "ApiKey cc_key" || len(offsets) != 2 {
		t.Fatalf("auth/pages wrong auth=%q offsets=%v", sawAuth, offsets)
	}
	if len(got) != 3 || got[0]["id"] != "form_1" || got[0]["received_on"] == nil {
		t.Fatalf("records mapped wrong: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := commcare.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"forms", "cases"} {
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
	if err != nil || cat.Connector != "commcare" || len(cat.Streams) != 2 {
		t.Fatalf("Catalog = %+v err=%v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("commcare"); !ok {
		t.Fatal("registry did not resolve commcare")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
