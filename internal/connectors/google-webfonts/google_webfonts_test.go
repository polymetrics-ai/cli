package googlewebfonts_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	googlewebfonts "polymetrics.ai/internal/connectors/google-webfonts"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Google
// Web Fonts connector: API-key query auth (?key=...), pagination across two
// pages via the optional pageToken/nextPageToken loop, and record mapping of
// font items.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("key")
		if r.URL.Path != "/webfonts" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("pageToken") {
		case "":
			_, _ = w.Write([]byte(`{"kind":"webfonts#webfontList","items":[` +
				`{"family":"Roboto","category":"sans-serif","version":"v30","lastModified":"2023-01-01","variants":["regular","700"],"subsets":["latin"]},` +
				`{"family":"Open Sans","category":"sans-serif","version":"v40","lastModified":"2023-02-02","variants":["regular"],"subsets":["latin","greek"]}` +
				`],"nextPageToken":"page2"}`))
		case "page2":
			_, _ = w.Write([]byte(`{"kind":"webfonts#webfontList","items":[` +
				`{"family":"Lato","category":"sans-serif","version":"v24","lastModified":"2023-03-03","variants":["regular"],"subsets":["latin"]}` +
				`]}`))
		default:
			t.Errorf("unexpected pageToken=%q", r.URL.Query().Get("pageToken"))
			_, _ = w.Write([]byte(`{"items":[]}`))
		}
	}))
	defer srv.Close()

	c := googlewebfonts.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "AIza_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "webfonts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "AIza_test_123" {
		t.Fatalf("key query param = %q, want AIza_test_123", sawKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["family"] == nil {
			t.Fatalf("record missing family: %+v", rec)
		}
		if rec["variant_count"] == nil {
			t.Fatalf("record missing derived variant_count: %+v", rec)
		}
	}
	if got[0]["family"] != "Roboto" {
		t.Fatalf("first record family = %v, want Roboto", got[0]["family"])
	}
}

// TestSortStreamSetsQueryParam verifies the sort-specialized streams pass the
// expected sort query parameter to the API.
func TestSortStreamSetsQueryParam(t *testing.T) {
	var sawSort string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawSort = r.URL.Query().Get("sort")
		_, _ = w.Write([]byte(`{"items":[{"family":"Roboto","variants":["regular"]}]}`))
	}))
	defer srv.Close()

	c := googlewebfonts.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "AIza_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "popular_fonts", Config: cfg}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Read popular_fonts: %v", err)
	}
	if sawSort != "popularity" {
		t.Fatalf("sort query param = %q, want popularity", sawSort)
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network access, so conformance passes without live credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := googlewebfonts.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "webfonts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["family"] == nil {
			t.Fatalf("fixture record missing family: %+v", rec)
		}
	}

	// Check must also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := googlewebfonts.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	if cat.Connector != "google-webfonts" {
		t.Fatalf("catalog connector = %q, want google-webfonts", cat.Connector)
	}
}

func TestBaseURLValidation(t *testing.T) {
	c := googlewebfonts.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "x"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "webfonts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected base_url scheme validation error")
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = googlewebfonts.New() // ensure init ran
	c := googlewebfonts.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only API)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("google-webfonts"); !ok {
		t.Fatal("registry did not resolve google-webfonts (self-registration)")
	}
}
