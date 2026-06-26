package fulcrum_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/fulcrum"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Fulcrum
// connector: X-ApiToken auth, page-number pagination over the "forms" array
// using total_pages/current_page, and record mapping. Red until
// internal/connectors/fulcrum exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.Header.Get("X-ApiToken")
		if r.URL.Path != "/forms.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"forms":[{"id":"form_1","name":"Inspection"},{"id":"form_2","name":"Survey"}],"total_count":3,"current_page":1,"total_pages":2,"per_page":2}`))
		case "2":
			_, _ = w.Write([]byte(`{"forms":[{"id":"form_3","name":"Audit"}],"total_count":3,"current_page":2,"total_pages":2,"per_page":2}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"forms":[],"total_pages":2,"current_page":99}`))
		}
	}))
	defer srv.Close()

	c := fulcrum.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "forms", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "tok_test_123" {
		t.Fatalf("X-ApiToken = %q, want tok_test_123", sawToken)
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

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records
// without touching the network, so conformance runs without live creds.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := fulcrum.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"forms", "records", "projects", "choice_lists", "classification_sets"} {
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
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog carries the core streams
// with primary keys.
func TestCatalogStreams(t *testing.T) {
	c := fulcrum.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "fulcrum" {
		t.Fatalf("catalog connector = %q, want fulcrum", cat.Connector)
	}
	want := map[string]bool{"forms": false, "records": false, "projects": false}
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
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestRegistryResolves asserts self-registration via NewRegistry().Get.
func TestRegistryResolves(t *testing.T) {
	_ = fulcrum.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("fulcrum"); !ok {
		t.Fatal("registry did not resolve fulcrum (self-registration)")
	}
	caps := fulcrum.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
