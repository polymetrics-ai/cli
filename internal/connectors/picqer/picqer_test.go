package picqer_test

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/picqer"
)

func TestReadProductsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth, sawUA string
	var offsets []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawUA = r.Header.Get("User-Agent")
		offsets = append(offsets, r.URL.Query().Get("offset"))
		if r.URL.Path != "/products" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "0":
			_, _ = w.Write([]byte(`[{"idproduct":11,"name":"Box","productcode":"BOX"}]`))
		case "1":
			_, _ = w.Write([]byte(`[]`))
		default:
			t.Fatalf("unexpected offset %q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	c := picqer.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1"}, Secrets: map[string]string{"api_key": "picqer_key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("Read: %v", err)
	}
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("picqer_key:"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if sawUA == "" {
		t.Fatal("missing required User-Agent")
	}
	if len(got) != 1 || got[0]["id"] == nil || got[0]["name"] != "Box" {
		t.Fatalf("records = %+v, want mapped product", got)
	}
	if len(offsets) != 2 {
		t.Fatalf("offsets = %v, want two requests", offsets)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := picqer.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "picqer" || len(cat.Streams) < 3 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("picqer"); !ok {
		t.Fatal("registry did not resolve picqer")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
