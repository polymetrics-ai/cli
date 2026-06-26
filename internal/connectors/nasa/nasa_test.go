package nasa_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/nasa"
)

// TestReadNeoBrowsePaginatesAndAuthenticates is the red-first test for the NASA
// connector. NASA APIs authenticate with an api_key query parameter (not a
// header), and the NeoWs browse endpoint paginates with page.number /
// page.total_pages over a near_earth_objects[] array. This asserts auth,
// pagination across 2 pages, and record mapping. Red until
// internal/connectors/nasa exists.
func TestReadNeoBrowsePaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("api_key")
		if r.URL.Path != "/neo/rest/v1/neo/browse" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "0":
			_, _ = w.Write([]byte(`{
				"near_earth_objects":[
					{"id":"2000433","neo_reference_id":"2000433","name":"433 Eros","absolute_magnitude_h":10.4,"is_potentially_hazardous_asteroid":false},
					{"id":"2000719","neo_reference_id":"2000719","name":"719 Albert","absolute_magnitude_h":15.5,"is_potentially_hazardous_asteroid":false}
				],
				"page":{"number":0,"total_pages":2,"size":2,"total_elements":3}
			}`))
		case "1":
			_, _ = w.Write([]byte(`{
				"near_earth_objects":[
					{"id":"2000887","neo_reference_id":"2000887","name":"887 Alinda","absolute_magnitude_h":13.8,"is_potentially_hazardous_asteroid":false}
				],
				"page":{"number":1,"total_pages":2,"size":2,"total_elements":3}
			}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"near_earth_objects":[],"page":{"number":9,"total_pages":2}}`))
		}
	}))
	defer srv.Close()

	c := nasa.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "test_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "neo_browse", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "test_key_123" {
		t.Fatalf("api_key query = %q, want test_key_123", sawKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestReadApodSingleObject confirms the APOD stream maps a single top-level
// object (no wrapper array) into one record keyed by date.
func TestReadApodSingleObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/planetary/apod" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"date":"2026-06-25","title":"The Cosmos","media_type":"image","url":"https://example.com/a.jpg","explanation":"A nice picture."}`))
	}))
	defer srv.Close()

	c := nasa.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "test_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "apod", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["date"] != "2026-06-25" || got[0]["title"] != "The Cosmos" {
		t.Fatalf("unexpected apod record: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork confirms the credential-free fixture path emits
// deterministic records without any network call, as conformance requires.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := nasa.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"apod", "neo_feed", "neo_browse", "epic", "mars_photos"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("Read(%s) fixture emitted no records", stream)
		}
	}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := nasa.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"apod": false, "neo_feed": false, "neo_browse": false, "epic": false, "mars_photos": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegistryResolution confirms the connector self-registers and resolves via
// the registry, and that it is read-only (no Write capability).
func TestRegistryResolution(t *testing.T) {
	_ = nasa.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("nasa"); !ok {
		t.Fatal("registry did not resolve nasa (self-registration)")
	}
	caps := nasa.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
}
