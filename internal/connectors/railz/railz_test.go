package railz_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/railz"
)

func TestReadBusinessesAuthenticatesPaginatesAndMapsRecords(t *testing.T) {
	var sawAuth string
	var offsets []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v1/businesses" {
			http.NotFound(w, r)
			return
		}
		offsets = append(offsets, r.URL.Query().Get("offset"))
		switch r.URL.Query().Get("offset") {
		case "0":
			_, _ = w.Write([]byte(`{"data":[{"businessName":"Northwind","businessUuid":"biz_1","createdAt":"2026-01-01T00:00:00Z"},{"businessName":"Contoso","businessUuid":"biz_2","createdAt":"2026-01-02T00:00:00Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"businessName":"Fabrikam","businessUuid":"biz_3","createdAt":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Fatalf("unexpected offset %q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	c := railz.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/v1", "page_size": "2"}, Secrets: map[string]string{"access_token": "railz_token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "businesses", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer railz_token" {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if len(offsets) != 2 || offsets[0] != "0" || offsets[1] != "2" {
		t.Fatalf("offsets = %v", offsets)
	}
	if len(got) != 3 || got[0]["id"] != "biz_1" || got[0]["name"] != "Northwind" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := railz.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "railz" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v", cat)
	}
	for _, stream := range cat.Streams {
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream.Name, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		}); err != nil {
			t.Fatalf("Read fixture %s: %v", stream.Name, err)
		}
		if len(got) == 0 || got[0]["id"] == nil {
			t.Fatalf("fixture %s records = %+v", stream.Name, got)
		}
	}
	if _, ok := connectors.NewRegistry().Get("railz"); !ok {
		t.Fatal("registry did not resolve railz")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
