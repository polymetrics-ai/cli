package plaid

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadPaginatesAuthenticatesAndMapsRecords(t *testing.T) {
	var sawClientID, sawSecret string
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/institutions/get" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		sawClientID, _ = body["client_id"].(string)
		sawSecret, _ = body["secret"].(string)
		if body["count"].(float64) != 2 {
			t.Fatalf("count = %v", body["count"])
		}
		switch body["offset"].(float64) {
		case 0:
			_, _ = w.Write([]byte(`{"institutions":[{"institution_id":"ins_1","name":"One Bank","country_codes":["US"]},{"institution_id":"ins_2","name":"Two Bank","country_codes":["US","CA"]}],"total":3}`))
		case 2:
			_, _ = w.Write([]byte(`{"institutions":[{"institution_id":"ins_3","name":"Three Bank","country_codes":["GB"]}],"total":3}`))
		default:
			t.Fatalf("unexpected offset %v", body["offset"])
		}
	}))
	defer srv.Close()

	c := New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2", "country_codes": "US,CA"},
		Secrets: map[string]string{"client_id": "unit-client", "secret": "unit-secret"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "institutions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawClientID != "unit-client" || sawSecret != "unit-secret" {
		t.Fatalf("auth body client_id=%q secret=%q", sawClientID, sawSecret)
	}
	if requests != 2 || len(got) != 3 || got[0]["institution_id"] != "ins_1" || got[0]["country_codes"] == "" {
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
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "institutions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["institution_id"] == nil {
		t.Fatalf("fixture records = %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) < 2 {
		t.Fatalf("Catalog = %+v, err=%v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("plaid"); !ok {
		t.Fatal("registry did not resolve plaid")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write error = %v", err)
	}
}
