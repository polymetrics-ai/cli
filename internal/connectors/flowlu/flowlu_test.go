package flowlu_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/flowlu"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Flowlu
// connector: api_key query-param auth, Flowlu page-number pagination over
// response.items (stopping on an empty page), and record mapping. Red until
// internal/connectors/flowlu exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("api_key")
		if r.URL.Path != "/crm/account/list" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"response":{"total_result":3,"total":2,"count":2,"page":1,"items":[{"id":1,"name":"Acme","type":1,"active":1},{"id":2,"name":"Globex","type":2,"active":1}]}}`))
		case "2":
			_, _ = w.Write([]byte(`{"response":{"total_result":3,"total":2,"count":2,"page":2,"items":[{"id":3,"name":"Initech","type":1,"active":0}]}}`))
		case "3":
			_, _ = w.Write([]byte(`{"response":{"total_result":3,"total":2,"count":2,"page":3,"items":[]}}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"response":{"items":[]}}`))
		}
	}))
	defer srv.Close()

	c := flowlu.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"page_size": "2",
		},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "key_test_123" {
		t.Fatalf("api_key query param = %q, want key_test_123", sawKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (across 2 non-empty pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["name"] == nil {
			t.Fatalf("record missing id/name: %+v", rec)
		}
	}
}

// TestFixtureMode verifies the deterministic, network-free fixture path used by
// the conformance harness.
func TestFixtureMode(t *testing.T) {
	c := flowlu.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tasks", Config: cfg}, func(rec connectors.Record) error {
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
	// Fixture mode must not require any secret.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := flowlu.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := flowlu.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "k"},
	}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check should reject non-http(s) base_url")
	}
}

func TestReadOnlyMetadataAndRegistry(t *testing.T) {
	_ = flowlu.New() // ensure init ran
	c := flowlu.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("flowlu"); !ok {
		t.Fatal("registry did not resolve flowlu (self-registration)")
	}
}
