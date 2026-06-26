package fillout_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/fillout"
)

// TestReadSubmissionsPaginatesAndAuthenticates is the red-first test for the
// Fillout connector: Bearer auth on /v1/api/forms and the per-form submissions
// endpoint, offset/limit pagination across two pages of the `responses` array,
// and record mapping (submissionId -> id with form_id attached).
func TestReadSubmissionsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		switch r.URL.Path {
		case "/forms":
			_, _ = w.Write([]byte(`[{"formId":"form_a","name":"Survey A"}]`))
		case "/forms/form_a/submissions":
			switch r.URL.Query().Get("offset") {
			case "", "0":
				// A full page of 2 (limit=2) means there is another page.
				_, _ = w.Write([]byte(`{"responses":[{"submissionId":"s1","submissionTime":"2026-01-01T00:00:00Z","lastUpdatedAt":"2026-01-01T00:00:00Z"},{"submissionId":"s2","submissionTime":"2026-01-02T00:00:00Z","lastUpdatedAt":"2026-01-02T00:00:00Z"}],"totalResponses":3,"pageCount":2}`))
			case "2":
				_, _ = w.Write([]byte(`{"responses":[{"submissionId":"s3","submissionTime":"2026-01-03T00:00:00Z","lastUpdatedAt":"2026-01-03T00:00:00Z"}],"totalResponses":3,"pageCount":2}`))
			default:
				t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
				_, _ = w.Write([]byte(`{"responses":[]}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := fillout.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "fk_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "submissions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer fk_test_123" {
		t.Fatalf("Authorization = %q, want Bearer fk_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
		if rec["form_id"] != "form_a" {
			t.Fatalf("record form_id = %v, want form_a: %+v", rec["form_id"], rec)
		}
	}
}

// TestReadFormsAuthenticates exercises the forms stream (top-level array,
// formId -> id mapping) and confirms the api_key flows into the Bearer header.
func TestReadFormsAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/forms" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"formId":"form_a","name":"Survey A"},{"formId":"form_b","name":"Survey B"}]`))
	}))
	defer srv.Close()

	c := fillout.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "fk_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "forms", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer fk_test_123" {
		t.Fatalf("Authorization = %q, want Bearer fk_test_123", sawAuth)
	}
	if len(got) != 2 {
		t.Fatalf("forms = %d, want 2", len(got))
	}
	if got[0]["id"] != "form_a" || got[0]["name"] != "Survey A" {
		t.Fatalf("first form mapped wrong: %+v", got[0])
	}
}

// TestFixtureModeNeedsNoNetwork confirms fixture mode emits deterministic
// records without any credentials or network access (conformance path).
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := fillout.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"forms", "questions", "submissions"} {
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
		if got[0]["id"] == nil {
			t.Fatalf("fixture %s record missing id: %+v", stream, got[0])
		}
	}
}

// TestRegistryResolvesFillout confirms self-registration via init() and that the
// connector is read-only (no Write capability).
func TestRegistryResolvesFillout(t *testing.T) {
	_ = fillout.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("fillout"); !ok {
		t.Fatal("registry did not resolve fillout (self-registration)")
	}
	caps := fillout.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("fillout should be read-only, got Write=true")
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := fillout.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"forms": false, "questions": false, "submissions": false}
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
