package breezyhr_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	breezyhr "polymetrics/internal/connectors/breezy-hr"
)

// TestReadPositionsPaginatesAndAuthenticates is the red-first test for the
// Breezy HR connector: raw-API-key Authorization header, page-based pagination
// over the top-level positions array, and record mapping (_id -> position_id).
// Red until internal/connectors/breezy-hr is implemented.
func TestReadPositionsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/company/co_123/positions" {
			http.NotFound(w, r)
			return
		}
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		switch page {
		case "1":
			// Full page (limit=2) signals more pages may follow.
			_, _ = w.Write([]byte(`[{"_id":"pos_1","name":"Engineer","state":"published","type":{"name":"Full-Time"}},{"_id":"pos_2","name":"Designer","state":"draft","type":{"name":"Contract"}}]`))
		case "2":
			// Short page (1 < limit) stops pagination.
			_, _ = w.Write([]byte(`[{"_id":"pos_3","name":"PM","state":"published","type":{"name":"Full-Time"}}]`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := breezyhr.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"company_id": "co_123",
			"page_size":  "2",
		},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "positions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "key_abc" {
		t.Fatalf("Authorization = %q, want raw api_key key_abc", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (across 2 pages)", len(got))
	}
	first := got[0]
	if first["position_id"] != "pos_1" {
		t.Fatalf("position_id = %v, want pos_1 (mapped from _id)", first["position_id"])
	}
	if first["type"] != "Full-Time" {
		t.Fatalf("type = %v, want Full-Time (flattened from type.name)", first["type"])
	}
	if first["name"] != "Engineer" {
		t.Fatalf("name = %v, want Engineer", first["name"])
	}
}

// TestReadCandidatesSubstream confirms candidates are read per-position and the
// position_id is propagated onto each candidate record.
func TestReadCandidatesSubstream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/company/co_123/positions":
			_, _ = w.Write([]byte(`[{"_id":"pos_1","name":"Engineer","type":{"name":"Full-Time"}}]`))
		case "/company/co_123/position/pos_1/candidates":
			_, _ = w.Write([]byte(`[{"_id":"cand_1","name":"Ada","email_address":"ada@example.com","stage":{"name":"Applied"}}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := breezyhr.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "company_id": "co_123"},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "candidates", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read candidates: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("candidate records = %d, want 1", len(got))
	}
	rec := got[0]
	if rec["id"] != "cand_1" {
		t.Fatalf("id = %v, want cand_1", rec["id"])
	}
	if rec["position_id"] != "pos_1" {
		t.Fatalf("position_id = %v, want pos_1 (propagated from parent)", rec["position_id"])
	}
	if rec["stage"] != "Applied" {
		t.Fatalf("stage = %v, want Applied (flattened from stage.name)", rec["stage"])
	}
}

// TestReadPipelines confirms the pipelines stream maps _id -> id.
func TestReadPipelines(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/company/co_123/pipelines" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"_id":"pipe_1","name":"Default"},{"_id":"pipe_2","name":"Engineering"}]`))
	}))
	defer srv.Close()

	c := breezyhr.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "company_id": "co_123"},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pipelines", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read pipelines: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("pipeline records = %d, want 2", len(got))
	}
	if got[0]["id"] != "pipe_1" || got[1]["name"] != "Engineering" {
		t.Fatalf("unexpected pipeline records: %+v", got)
	}
}

// TestFixtureModeReadsWithoutNetwork confirms credential-free conformance: with
// mode=fixture the connector emits deterministic records and makes no request.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := breezyhr.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"positions", "candidates", "pipelines"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read %s: %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
	}
	// Check must also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams
// with primary keys.
func TestCatalogStreams(t *testing.T) {
	c := breezyhr.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"positions": false, "candidates": false, "pipelines": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
			if len(s.PrimaryKey) == 0 {
				t.Fatalf("stream %s missing primary key", s.Name)
			}
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = breezyhr.New() // ensure init ran
	c := breezyhr.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only connector)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("breezy-hr"); !ok {
		t.Fatal("registry did not resolve breezy-hr (self-registration)")
	}
}
