package recreation_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/recreation"
)

func TestReadFacilitiesAuthenticatesPaginatesAndMapsRecords(t *testing.T) {
	var sawKey string
	var offsets []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("apikey")
		if r.URL.Path != "/api/v1/facilities" {
			http.NotFound(w, r)
			return
		}
		offsets = append(offsets, r.URL.Query().Get("offset"))
		switch r.URL.Query().Get("offset") {
		case "0":
			_, _ = w.Write([]byte(`{"RECDATA":[{"FacilityID":"1","FacilityName":"Pine Campground","FacilityTypeDescription":"Campground"},{"FacilityID":"2","FacilityName":"Cedar Trail","FacilityTypeDescription":"Trail"}],"METADATA":{"RESULTS":{"TOTAL_COUNT":3,"CURRENT_COUNT":2}}}`))
		case "2":
			_, _ = w.Write([]byte(`{"RECDATA":[{"FacilityID":"3","FacilityName":"Lake Overlook","FacilityTypeDescription":"Day Use"}],"METADATA":{"RESULTS":{"TOTAL_COUNT":3,"CURRENT_COUNT":1}}}`))
		default:
			t.Fatalf("unexpected offset %q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	c := recreation.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/api/v1", "page_size": "2"}, Secrets: map[string]string{"api_key": "ridb_key"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "facilities", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "ridb_key" {
		t.Fatalf("apikey = %q", sawKey)
	}
	if len(offsets) != 2 || offsets[0] != "0" || offsets[1] != "2" {
		t.Fatalf("offsets = %v", offsets)
	}
	if len(got) != 3 || got[0]["id"] != "1" || got[0]["name"] != "Pine Campground" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := recreation.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "recreation" || len(cat.Streams) == 0 {
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
	if _, ok := connectors.NewRegistry().Get("recreation"); !ok {
		t.Fatal("registry did not resolve recreation")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
