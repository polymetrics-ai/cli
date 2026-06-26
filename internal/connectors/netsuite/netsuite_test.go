package netsuite_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/netsuite"
)

func TestReadPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var offsets []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/customer" {
			http.NotFound(w, r)
			return
		}
		offsets = append(offsets, r.URL.Query().Get("offset"))
		switch r.URL.Query().Get("offset") {
		case "":
			_, _ = w.Write([]byte(`{"items":[{"id":"1","entityId":"Acme","lastModifiedDate":"2026-01-01T00:00:00Z"}],"hasMore":true}`))
		case "1":
			_, _ = w.Write([]byte(`{"items":[{"id":"2","entityId":"Globex","lastModifiedDate":"2026-01-02T00:00:00Z"}],"hasMore":false}`))
		default:
			t.Fatalf("unexpected offset %q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	c := netsuite.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": srv.URL, "realm": "123456", "page_size": "1"},
		Secrets: map[string]string{
			"consumer_key":    "ck",
			"consumer_secret": "cs",
			"token_key":       "tk",
			"token_secret":    "ts",
		},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !strings.HasPrefix(sawAuth, "OAuth ") || !strings.Contains(sawAuth, `realm="123456"`) {
		t.Fatalf("Authorization = %q, want OAuth header with realm", sawAuth)
	}
	if strings.Join(offsets, ",") != ",1" {
		t.Fatalf("offsets = %v, want first page then offset 1", offsets)
	}
	if len(got) != 2 || got[0]["id"] != "1" || got[0]["entity_id"] != "Acme" || got[1]["entity_id"] != "Globex" {
		t.Fatalf("mapped records wrong: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := netsuite.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records missing id: %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) == 0 || cat.Streams[0].PrimaryKey[0] != "id" {
		t.Fatalf("catalog streams invalid: %+v", cat.Streams)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("netsuite"); !ok {
		t.Fatal("registry did not resolve netsuite")
	}
}
