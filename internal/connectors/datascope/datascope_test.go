package datascope_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/datascope"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the DataScope
// connector: it asserts the Authorization header carries the raw api_key,
// offset/limit pagination walks two pages of the top-level JSON array, and the
// records are mapped through (id present).
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPaths = append(sawPaths, r.URL.Path)
		if r.URL.Path != "/locations" {
			http.NotFound(w, r)
			return
		}
		// page_size is 200 in the connector; emulate a full first page so the
		// paginator advances, then a short second page so it stops.
		offset := r.URL.Query().Get("offset")
		switch offset {
		case "", "0":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(buildLocations(0, 200)))
		case "200":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(buildLocations(200, 1)))
		default:
			t.Errorf("unexpected offset=%q", offset)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := datascope.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "200"},
		Secrets: map[string]string{"api_key": "dsk_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "locations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "dsk_test_123" {
		t.Fatalf("Authorization = %q, want raw api_key dsk_test_123", sawAuth)
	}
	if len(got) != 201 {
		t.Fatalf("records = %d, want 201 (2 pages: 200 + 1)", len(got))
	}
	if len(sawPaths) != 2 {
		t.Fatalf("requests = %d, want 2 (pagination)", len(sawPaths))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

// buildLocations returns a JSON array of n location objects starting at the
// given base id, exercising the root-array record selector.
func buildLocations(base, n int) string {
	out := "["
	for i := 0; i < n; i++ {
		if i > 0 {
			out += ","
		}
		id := strconv.Itoa(base + i + 1)
		out += `{"id":` + id + `,"name":"Site ` + id + `","city":"Lisbon"}`
	}
	return out + "]"
}

// TestAnswersAppliesDateWindow asserts the answers stream sends the start/end
// request parameters (DataScope's date filter) and reads /v2/answers.
func TestAnswersAppliesDateWindow(t *testing.T) {
	var sawStart, sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		sawStart = r.URL.Query().Get("start")
		_, _ = w.Write([]byte(`[{"form_answer_id":42,"form_name":"Inspection","created_at":"01/02/2026 10:00"}]`))
	}))
	defer srv.Close()

	c := datascope.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "start_date": "01/01/2026 00:00"},
		Secrets: map[string]string{"api_key": "dsk_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "answers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read answers: %v", err)
	}
	if sawPath != "/v2/answers" {
		t.Fatalf("path = %q, want /v2/answers", sawPath)
	}
	if sawStart != "01/01/2026 00:00" {
		t.Fatalf("start = %q, want the configured start_date", sawStart)
	}
	if len(got) != 1 || got[0]["form_answer_id"] == nil {
		t.Fatalf("answers records = %+v, want one with form_answer_id", got)
	}
}

// TestFixtureModeNeedsNoNetwork confirms fixture mode emits deterministic
// records without any credentials or network, so conformance can run.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := datascope.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "locations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check + Catalog must also work credential-free in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
}

// TestRegistryResolves asserts the connector self-registers and resolves via the
// shared registry, with read-only capabilities (no reverse-ETL writes).
func TestRegistryResolves(t *testing.T) {
	_ = datascope.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("datascope")
	if !ok {
		t.Fatal("registry did not resolve datascope (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
}
