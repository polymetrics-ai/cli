package avni_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/avni"
)

func TestReadSubjectsPaginatesAuthenticatesAndMapsRecords(t *testing.T) {
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "avni_user" || pass != "test_password" {
			t.Fatalf("unexpected basic auth user=%q ok=%v", user, ok)
		}
		if r.URL.Path != "/api/subjects" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		pages = append(pages, page)
		switch page {
		case "", "1":
			_, _ = w.Write([]byte(`{"items":[{"id":"sub_1","name":"Ada","updated_at":"2026-01-01T00:00:00Z"},{"id":"sub_2","name":"Grace","updated_at":"2026-01-02T00:00:00Z"}],"next_page":"2"}`))
		case "2":
			_, _ = w.Write([]byte(`{"items":[{"id":"sub_3","name":"Katherine","updated_at":"2026-01-03T00:00:00Z"}],"next_page":""}`))
		default:
			t.Fatalf("unexpected page %q", page)
		}
	}))
	defer srv.Close()

	c := avni.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "avni_user", "start_date": "2026-01-01T00:00:00Z", "page_size": "2"},
		Secrets: map[string]string{"password": "test_password"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "subjects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("requests = %d, want 2 pages", len(pages))
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3", len(got))
	}
	if got[0]["id"] != "sub_1" || got[0]["name"] != "Ada" {
		t.Fatalf("first record mapped wrong: %+v", got[0])
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := avni.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"subjects", "encounters"} {
		count := 0
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(connectors.Record) error {
			count++
			return nil
		}); err != nil {
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
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "avni" || len(cat.Streams) != 2 {
		t.Fatalf("catalog = %+v, want avni with 2 streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("avni"); !ok {
		t.Fatal("registry did not resolve avni")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
