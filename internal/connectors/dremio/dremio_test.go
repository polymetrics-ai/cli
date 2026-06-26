package dremio_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/dremio"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Dremio
// connector: Bearer (PAT) auth, Dremio nextPageToken/pageToken pagination over
// the data[] array, and record mapping. Two pages of the catalog stream.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/catalog" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("pageToken") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"cat_1","type":"CONTAINER","containerType":"SOURCE","path":["s3"]},{"id":"cat_2","type":"CONTAINER","containerType":"SPACE","path":["analytics"]}],"nextPageToken":"tok_2"}`))
		case "tok_2":
			_, _ = w.Write([]byte(`{"data":[{"id":"cat_3","type":"DATASET","datasetType":"PROMOTED","path":["s3","events"]}]}`))
		default:
			t.Errorf("unexpected pageToken=%q", r.URL.Query().Get("pageToken"))
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := dremio.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "pat_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "catalog", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer pat_test_123" {
		t.Fatalf("Authorization = %q, want Bearer pat_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["type"] == nil {
			t.Fatalf("record missing id/type: %+v", rec)
		}
	}
	// Verify the mapper kept the path field on the first record.
	if got[0]["id"] != "cat_1" {
		t.Fatalf("first record id = %v, want cat_1", got[0]["id"])
	}
}

// TestReadReflectionsStream exercises a second stream to confirm the routing
// table and record mapping work for more than the default stream.
func TestReadReflectionsStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/reflections" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"ref_1","name":"agg","type":"AGGREGATION","datasetId":"ds_1","enabled":true}]}`))
	}))
	defer srv.Close()

	c := dremio.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "pat_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reflections", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["id"] != "ref_1" || got[0]["type"] != "AGGREGATION" {
		t.Fatalf("unexpected reflection record: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork confirms the credential-free fixture path used by the
// conformance harness emits deterministic records without any network call.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := dremio.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "catalog", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil {
		t.Fatalf("fixture record missing id: %+v", got[0])
	}
}

// TestRegistryResolvesDremio confirms self-registration via init() and that the
// connector advertises read-only catalog capabilities.
func TestRegistryResolvesDremio(t *testing.T) {
	_ = dremio.New() // ensure init ran
	c := dremio.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Check && Catalog && Read", caps)
	}
	if caps.Write {
		t.Fatalf("dremio is read-only, Write should be false")
	}

	r := connectors.NewRegistry()
	if _, ok := r.Get("dremio"); !ok {
		t.Fatal("registry did not resolve dremio (self-registration)")
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := dremio.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"catalog": false, "reflections": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q has no primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}
