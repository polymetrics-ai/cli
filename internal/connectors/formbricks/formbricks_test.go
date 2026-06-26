package formbricks_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/formbricks"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Formbricks
// connector: X-API-Key auth, offset (skip) pagination over data[] for the
// responses stream, and record mapping. Red until internal/connectors/formbricks
// exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("X-API-Key")
		sawPaths = append(sawPaths, r.URL.Path)
		if r.URL.Path != "/management/responses" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("skip") {
		case "", "0":
			// First page: a full page (2 records) signals more pages exist.
			_, _ = w.Write([]byte(`{"data":[{"id":"resp_1","surveyId":"srv_1","finished":true},{"id":"resp_2","surveyId":"srv_1","finished":false}]}`))
		case "2":
			// Second page: a short page (1 record) ends pagination.
			_, _ = w.Write([]byte(`{"data":[{"id":"resp_3","surveyId":"srv_2","finished":true}]}`))
		default:
			t.Errorf("unexpected skip=%q", r.URL.Query().Get("skip"))
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := formbricks.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "fb_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "responses", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "fb_test_key" {
		t.Fatalf("X-API-Key = %q, want fb_test_key", sawAPIKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages), paths=%v", len(got), sawPaths)
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["survey_id"] == nil {
			t.Fatalf("record missing id/survey_id: %+v", rec)
		}
	}
	if len(sawPaths) != 2 {
		t.Fatalf("expected 2 page requests, got %d: %v", len(sawPaths), sawPaths)
	}
}

// TestReadSurveysSinglePage exercises a non-paginated stream that returns all
// records under data[] in one response.
func TestReadSurveysSinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "fb_key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/management/surveys" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"srv_1","name":"NPS","type":"link","status":"inProgress","environmentId":"env_1"}]}`))
	}))
	defer srv.Close()

	c := formbricks.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "fb_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "surveys", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["id"] != "srv_1" || got[0]["name"] != "NPS" || got[0]["environment_id"] != "env_1" {
		t.Fatalf("unexpected survey record: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records with
// no network access, so conformance can run without live credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := formbricks.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"surveys", "responses", "action_classes", "attribute_classes", "webhooks"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s, fixture): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture mode for %s emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}

	// Check must not require a network call in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestUnknownStream rejects an unknown stream name.
func TestUnknownStream(t *testing.T) {
	c := formbricks.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "nope", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with unknown stream should error")
	}
}

// TestBaseURLValidation rejects non-http(s) base URLs (SSRF guard).
func TestBaseURLValidation(t *testing.T) {
	c := formbricks.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "surveys", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should be rejected")
	}
}

// TestCatalogAndMetadata checks the published streams and read-only capabilities.
func TestCatalogAndMetadata(t *testing.T) {
	c := formbricks.New()
	meta := c.Metadata()
	if !meta.Capabilities.Read || !meta.Capabilities.Catalog || !meta.Capabilities.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", meta.Capabilities)
	}
	if meta.Capabilities.Write {
		t.Fatalf("formbricks is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 5 {
		t.Fatalf("catalog streams = %d, want >= 5", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
}

// TestRegisteredWithRegistry confirms self-registration resolves via NewRegistry.
func TestRegisteredWithRegistry(t *testing.T) {
	_ = formbricks.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("formbricks"); !ok {
		t.Fatal("registry did not resolve formbricks (self-registration)")
	}
}
