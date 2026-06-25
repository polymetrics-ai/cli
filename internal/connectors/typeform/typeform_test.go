package typeform_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/typeform"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Typeform
// connector: Bearer auth (personal access token), Typeform page/page_size
// pagination over items[], and record mapping for the forms stream.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/forms" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{"total_items":3,"page_count":2,"items":[{"id":"form_1","title":"Onboarding","last_updated_at":"2026-01-01T00:00:00Z"},{"id":"form_2","title":"NPS","last_updated_at":"2026-01-02T00:00:00Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"total_items":3,"page_count":2,"items":[{"id":"form_3","title":"Survey","last_updated_at":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`{"total_items":3,"page_count":2,"items":[]}`))
		}
	}))
	defer srv.Close()

	c := typeform.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"access_token": "tfp_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "forms", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tfp_abc123" {
		t.Fatalf("Authorization = %q, want Bearer tfp_abc123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["title"] == nil {
			t.Fatalf("record missing id/title: %+v", rec)
		}
	}
}

// TestReadResponsesPerForm exercises the per-form responses stream, which reads
// GET /forms/{form_id}/responses paginated by page over items[].
func TestReadResponsesPerForm(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/forms/u6nXL7/responses" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		switch page {
		case "1":
			_, _ = w.Write([]byte(`{"total_items":2,"page_count":2,"items":[{"response_id":"r_1","submitted_at":"2026-01-01T00:00:00Z","landed_at":"2026-01-01T00:00:00Z"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"total_items":2,"page_count":2,"items":[{"response_id":"r_2","submitted_at":"2026-01-02T00:00:00Z","landed_at":"2026-01-02T00:00:00Z"}]}`))
		default:
			_, _ = w.Write([]byte(`{"total_items":2,"page_count":2,"items":[]}`))
		}
	}))
	defer srv.Close()

	c := typeform.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "1", "form_ids": "u6nXL7"},
		Secrets: map[string]string{"access_token": "tfp_abc123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "responses", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read responses: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("responses = %d, want 2 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["response_id"] == nil {
			t.Fatalf("response record missing response_id: %+v", rec)
		}
		if rec["form_id"] != "u6nXL7" {
			t.Fatalf("response form_id = %v, want u6nXL7", rec["form_id"])
		}
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// with no network call and no credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := typeform.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"forms", "responses", "workspaces", "themes"} {
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
	// Check must also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := typeform.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "typeform" {
		t.Fatalf("catalog connector = %q, want typeform", cat.Connector)
	}
	want := map[string]bool{"forms": false, "responses": false, "workspaces": false, "themes": false, "images": false}
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
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestRegistryResolves(t *testing.T) {
	_ = typeform.New() // ensure init ran
	c := typeform.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("typeform is read-only, want Write=false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("typeform"); !ok {
		t.Fatal("registry did not resolve typeform (self-registration)")
	}
}
