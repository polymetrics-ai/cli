package clarifai_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	clarifai "polymetrics.ai/internal/connectors/clarif-ai"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Clarif-ai
// connector: Clarifai "Authorization: Key <pat>" auth, page/per_page pagination
// across two pages over the apps[] array, and record mapping. Red until
// internal/connectors/clarif-ai exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/users/me-user/apps" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"status":{"code":10000},"apps":[{"id":"app_1","name":"App One"},{"id":"app_2","name":"App Two"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"status":{"code":10000},"apps":[{"id":"app_3","name":"App Three"}]}`))
		default:
			_, _ = w.Write([]byte(`{"status":{"code":10000},"apps":[]}`))
		}
	}))
	defer srv.Close()

	c := clarifai.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "user_id": "me-user", "page_size": "2"},
		Secrets: map[string]string{"api_key": "pat_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "applications", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Key pat_test_123" {
		t.Fatalf("Authorization = %q, want Key pat_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
	if got[0]["name"] != "App One" {
		t.Fatalf("record name = %v, want App One", got[0]["name"])
	}
}

// TestUserIDPath verifies the user_id config is interpolated into the request
// path for a user-scoped stream (models).
func TestUserIDPath(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		_, _ = w.Write([]byte(`{"status":{"code":10000},"models":[{"id":"m_1"}]}`))
	}))
	defer srv.Close()

	c := clarifai.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "user_id": "acme"},
		Secrets: map[string]string{"api_key": "pat_test_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "models", Config: cfg}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawPath != "/users/acme/models" {
		t.Fatalf("path = %q, want /users/acme/models", sawPath)
	}
}

// TestFixtureMode confirms the credential-free fixture path emits deterministic
// records with no network access (mandatory for conformance).
func TestFixtureMode(t *testing.T) {
	c := clarifai.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"applications", "datasets", "models", "model_versions", "workflows"} {
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
				t.Fatalf("fixture record missing id: %+v", rec)
			}
		}
	}

	// Check + Catalog must also work credential-free in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = clarifai.New() // ensure init ran
	c := clarifai.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only API)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("clarif-ai"); !ok {
		t.Fatal("registry did not resolve clarif-ai (self-registration)")
	}
}
