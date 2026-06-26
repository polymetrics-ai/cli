package pocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadPaginatesAuthenticatesAndMapsRecords(t *testing.T) {
	var sawConsumerKey, sawAccessToken string
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/v3/get" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		sawConsumerKey, _ = body["consumer_key"].(string)
		sawAccessToken, _ = body["access_token"].(string)
		if body["count"].(float64) != 2 {
			t.Fatalf("count = %v", body["count"])
		}
		switch body["offset"].(float64) {
		case 0:
			_, _ = w.Write([]byte(`{"status":1,"list":{"100":{"item_id":"100","resolved_title":"One","given_url":"https://example.com/1","time_updated":"1760000000"},"101":{"item_id":"101","resolved_title":"Two","given_url":"https://example.com/2","time_updated":"1760000001"}}}`))
		case 2:
			_, _ = w.Write([]byte(`{"status":1,"list":{"102":{"item_id":"102","resolved_title":"Three","given_url":"https://example.com/3","time_updated":"1760000002"}}}`))
		default:
			t.Fatalf("unexpected offset %v", body["offset"])
		}
	}))
	defer srv.Close()

	c := New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/v3", "page_size": "2"},
		Secrets: map[string]string{"consumer_key": "unit-consumer", "access_token": "unit-access"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "items", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawConsumerKey != "unit-consumer" || sawAccessToken != "unit-access" {
		t.Fatalf("auth body consumer_key=%q access_token=%q", sawConsumerKey, sawAccessToken)
	}
	if requests != 2 || len(got) != 3 || got[0]["item_id"] == nil || got[0]["title"] == nil {
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
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "items", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["item_id"] == nil {
		t.Fatalf("fixture records = %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) < 1 {
		t.Fatalf("Catalog = %+v, err=%v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("pocket"); !ok {
		t.Fatal("registry did not resolve pocket")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write error = %v", err)
	}
}
