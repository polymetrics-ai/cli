package highlevel_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	highlevel "polymetrics.ai/internal/connectors/high-level"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the HighLevel
// connector: x-api-key auth, the required locationId query param, cursor
// pagination that follows the response meta.nextPageUrl across two pages, and
// record mapping. Red until internal/connectors/high-level exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	var sawVersion string
	var sawLocation string
	pages := 0

	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("x-api-key")
		sawVersion = r.Header.Get("Version")
		if loc := r.URL.Query().Get("locationId"); loc != "" {
			sawLocation = loc
		}
		if r.URL.Path != "/upstream/contacts" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("startAfterId") {
		case "":
			pages++
			next := srvURL + "/upstream/contacts?locationId=loc_123&startAfterId=ct_2"
			_, _ = w.Write([]byte(`{"contacts":[{"id":"ct_1","dateUpdated":"2026-01-01T00:00:00Z"},{"id":"ct_2","dateUpdated":"2026-01-02T00:00:00Z"}],"meta":{"nextPageUrl":"` + next + `","startAfterId":"ct_2"}}`))
		case "ct_2":
			pages++
			_, _ = w.Write([]byte(`{"contacts":[{"id":"ct_3","dateUpdated":"2026-01-03T00:00:00Z"}],"meta":{"nextPageUrl":null}}`))
		default:
			t.Errorf("unexpected startAfterId=%q", r.URL.Query().Get("startAfterId"))
			_, _ = w.Write([]byte(`{"contacts":[],"meta":{}}`))
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := highlevel.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "location_id": "loc_123"},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "key_abc" {
		t.Fatalf("x-api-key = %q, want key_abc", sawAPIKey)
	}
	if sawVersion == "" {
		t.Fatalf("Version header = %q, want non-empty", sawVersion)
	}
	if sawLocation != "loc_123" {
		t.Fatalf("locationId = %q, want loc_123", sawLocation)
	}
	if pages != 2 {
		t.Fatalf("server saw %d pages, want 2", pages)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

// TestReadSingleRequestStream verifies a non-paginated stream (pipelines) makes
// exactly one request and maps records from its selector.
func TestReadSingleRequestStream(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/upstream/pipelines" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"pipelines":[{"id":"pl_1","name":"Sales"},{"id":"pl_2","name":"Onboarding"}]}`))
	}))
	defer srv.Close()

	c := highlevel.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "location_id": "loc_123"},
		Secrets: map[string]string{"api_key": "key_abc"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pipelines", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1 (single-request stream)", requests)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records with
// no network access so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := highlevel.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check must also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCheckRequiresLocationAndSecret(t *testing.T) {
	c := highlevel.New()
	// Missing location_id.
	err := c.Check(context.Background(), connectors.RuntimeConfig{
		Config:  map[string]string{},
		Secrets: map[string]string{"api_key": "key_abc"},
	})
	if err == nil {
		t.Fatal("Check should fail when location_id is missing")
	}
	// Missing api_key.
	err = c.Check(context.Background(), connectors.RuntimeConfig{
		Config:  map[string]string{"location_id": "loc_123"},
		Secrets: map[string]string{},
	})
	if err == nil {
		t.Fatal("Check should fail when api_key secret is missing")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := highlevel.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "high-level" {
		t.Fatalf("catalog connector = %q, want high-level", cat.Connector)
	}
	want := map[string]bool{"contacts": false, "opportunities": false, "pipelines": false, "custom_fields": false, "form_submissions": false}
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
			t.Fatalf("catalog missing expected stream %q", name)
		}
	}
}

func TestRegisteredAndResolvable(t *testing.T) {
	_ = highlevel.New() // ensure init ran
	c := highlevel.New()
	if c.Name() != "high-level" {
		t.Fatalf("Name() = %q, want high-level", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("high-level"); !ok {
		t.Fatal("registry did not resolve high-level (self-registration)")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := highlevel.New()
	err := c.Read(context.Background(), connectors.ReadRequest{
		Stream: "contacts",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": "ftp://evil.example.com", "location_id": "loc_123"},
			Secrets: map[string]string{"api_key": "key_abc"},
		},
	}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with bad base_url scheme = %v, want base_url error", err)
	}
}
