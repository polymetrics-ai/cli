package linkrunner_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/linkrunner"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the
// linkrunner-key API-key header, page/limit pagination across two pages, the
// data.campaigns record path, and primary-key mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("linkrunner-key")
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v1/campaigns" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"data":{"campaigns":[{"display_id":"camp_1","name":"Spring"},{"display_id":"camp_2","name":"Summer"}]}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":{"campaigns":[{"display_id":"camp_3","name":"Fall"}]}}`))
		case "3":
			_, _ = w.Write([]byte(`{"data":{"campaigns":[]}}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"data":{"campaigns":[]}}`))
		}
	}))
	defer srv.Close()

	c := linkrunner.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/api/v1", "page_size": "2"},
		Secrets: map[string]string{"linkrunner-key": "lr_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "lr_test_123" {
		t.Fatalf("linkrunner-key header = %q, want lr_test_123", sawKey)
	}
	if sawAuth != "" {
		t.Fatalf("Authorization header should be empty, got %q", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["display_id"] == nil {
			t.Fatalf("record missing display_id: %+v", rec)
		}
	}
}

// TestReadAttributedUsers verifies the attributed_users stream reads from
// data.users and forwards the required display_id request parameter.
func TestReadAttributedUsers(t *testing.T) {
	var sawDisplayID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/attributed-users" {
			http.NotFound(w, r)
			return
		}
		sawDisplayID = r.URL.Query().Get("display_id")
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"data":{"users":[{"campaign_display_id":"camp_1","attributed_at":"2026-01-01T00:00:00Z"}]}}`))
		default:
			_, _ = w.Write([]byte(`{"data":{"users":[]}}`))
		}
	}))
	defer srv.Close()

	c := linkrunner.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL + "/api/v1", "display_id": "camp_1"},
		Secrets: map[string]string{"linkrunner-key": "lr_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "attributed_users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawDisplayID != "camp_1" {
		t.Fatalf("display_id param = %q, want camp_1", sawDisplayID)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
}

// TestFixtureMode confirms the credential-free fixture path emits records.
func TestFixtureMode(t *testing.T) {
	c := linkrunner.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	if got[0]["display_id"] == nil {
		t.Fatalf("fixture record missing display_id: %+v", got[0])
	}
}

// TestCatalogAndMetadata verifies the published catalog and read-only caps.
func TestCatalogAndMetadata(t *testing.T) {
	c := linkrunner.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && Catalog && !Write", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 2 {
		t.Fatalf("streams = %d, want >= 2", len(cat.Streams))
	}
}

// TestRegistryResolution confirms self-registration via NewRegistry().
func TestRegistryResolution(t *testing.T) {
	_ = linkrunner.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("linkrunner"); !ok {
		t.Fatal("registry did not resolve linkrunner (self-registration)")
	}
}
