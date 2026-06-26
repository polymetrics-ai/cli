package poplar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadPaginatesAuthenticatesAndMapsRecords(t *testing.T) {
	var sawAuth, sawPerPage string
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v1/campaigns" {
			http.NotFound(w, r)
			return
		}
		sawPerPage = r.URL.Query().Get("per_page")
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`{"data":[{"id":"camp_1","name":"One","status":"active"},{"id":"camp_2","name":"Two","status":"draft"}],"meta":{"next_page":2}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":"camp_3","name":"Three","status":"archived"}],"meta":{"next_page":null}}`))
		default:
			t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
		}
	}))
	defer srv.Close()

	c := New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/v1", "page_size": "2"},
		Secrets: map[string]string{"api_token": "unit-token"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer unit-token" || sawPerPage != "2" {
		t.Fatalf("auth=%q per_page=%q", sawAuth, sawPerPage)
	}
	if requests != 2 || len(got) != 3 || got[0]["id"] != "camp_1" || got[0]["status"] != "active" {
		t.Fatalf("requests=%d records=%+v", requests, got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "orders", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) < 2 {
		t.Fatalf("Catalog = %+v, err=%v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("poplar"); !ok {
		t.Fatal("registry did not resolve poplar")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write error = %v", err)
	}
}
