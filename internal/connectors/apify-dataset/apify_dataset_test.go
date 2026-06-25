package apifydataset_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	apifydataset "polymetrics/internal/connectors/apify-dataset"
)

// TestReadItemsPaginatesAndAuthenticates is the red-first test for the Apify
// Dataset connector. The item_collection stream reads
// GET /v2/datasets/{datasetId}/items which returns a top-level JSON array, and
// paginates by offset/limit. It must send the Bearer token and stop once a short
// page is returned.
func TestReadItemsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/datasets/ds_123/items" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			// Full page (limit=2) -> there is a next page.
			_, _ = w.Write([]byte(`[{"id":"a","value":1},{"id":"b","value":2}]`))
		case "2":
			// Short page -> last page.
			_, _ = w.Write([]byte(`[{"id":"c","value":3}]`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := apifydataset.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "dataset_id": "ds_123", "page_size": "2"},
		Secrets: map[string]string{"token": "apify_api_secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "item_collection", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer apify_api_secret" {
		t.Fatalf("Authorization = %q, want Bearer apify_api_secret", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	// item_collection wraps each raw item under a "data" key (dynamic schema).
	for _, rec := range got {
		if rec["data"] == nil {
			t.Fatalf("item record missing data wrapper: %+v", rec)
		}
	}
}

// TestReadDatasetCollectionPaginates exercises the management-endpoint shape,
// where datasets are wrapped under data.items and paginate by offset/limit.
func TestReadDatasetCollectionPaginates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/datasets" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"data":{"total":3,"offset":0,"limit":2,"count":2,"items":[{"id":"ds1","name":"one"},{"id":"ds2","name":"two"}]}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":{"total":3,"offset":2,"limit":2,"count":1,"items":[{"id":"ds3","name":"three"}]}}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"data":{"items":[]}}`))
		}
	}))
	defer srv.Close()

	c := apifydataset.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"token": "apify_api_secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "dataset_collection", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["id"] != "ds1" {
		t.Fatalf("first dataset id = %v, want ds1", got[0]["id"])
	}
}

// TestReadDatasetSingle reads the single-dataset metadata stream, which returns
// one object under data.
func TestReadDatasetSingle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/datasets/ds_123" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":{"id":"ds_123","name":"my-dataset","itemCount":42}}`))
	}))
	defer srv.Close()

	c := apifydataset.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "dataset_id": "ds_123"},
		Secrets: map[string]string{"token": "apify_api_secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "dataset", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["id"] != "ds_123" || got[0]["name"] != "my-dataset" {
		t.Fatalf("record = %+v, want id=ds_123 name=my-dataset", got[0])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access so conformance can run without credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := apifydataset.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "dataset_id": "ds_fixture"}}

	for _, stream := range []string{"item_collection", "dataset_collection", "dataset"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := apifydataset.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestRegisteredReadOnly confirms self-registration via the connectors registry
// and that the connector advertises read (not write) capability.
func TestRegisteredReadOnly(t *testing.T) {
	c := apifydataset.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only connector)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("apify-dataset"); !ok {
		t.Fatal("registry did not resolve apify-dataset (self-registration)")
	}
}

// TestCatalogStreams confirms the published streams are present with primary keys.
func TestCatalogStreams(t *testing.T) {
	c := apifydataset.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"item_collection": false, "dataset_collection": false, "dataset": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}
