package openfda_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/openfda"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the openFDA
// connector: the optional api_key flows through as a query parameter, offset
// pagination (skip/limit) walks two pages over results[], the record is mapped,
// and meta.results.total bounds the walk. Red until internal/connectors/openfda
// exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.URL.Query().Get("api_key")
		if r.URL.Path != "/drug/event.json" {
			http.NotFound(w, r)
			return
		}
		// Page size 2, total 3 records: page 1 (skip 0) returns 2, page 2
		// (skip 2) returns the last 1, then the walk stops.
		skip, _ := strconv.Atoi(r.URL.Query().Get("skip"))
		switch skip {
		case 0:
			_, _ = w.Write([]byte(`{"meta":{"results":{"skip":0,"limit":2,"total":3}},"results":[{"safetyreportid":"r1","receivedate":"20240101"},{"safetyreportid":"r2","receivedate":"20240102"}]}`))
		case 2:
			_, _ = w.Write([]byte(`{"meta":{"results":{"skip":2,"limit":2,"total":3}},"results":[{"safetyreportid":"r3","receivedate":"20240103"}]}`))
		default:
			t.Errorf("unexpected skip=%d", skip)
			_, _ = w.Write([]byte(`{"meta":{"results":{"skip":0,"limit":2,"total":3}},"results":[]}`))
		}
	}))
	defer srv.Close()

	c := openfda.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "k_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "drug_event", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "k_test_123" {
		t.Fatalf("api_key = %q, want k_test_123", sawAPIKey)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["safetyreportid"] == nil {
			t.Fatalf("record missing safetyreportid: %+v", rec)
		}
	}
}

// TestReadWithoutAPIKeyOmitsParam verifies that when no api_key secret is set,
// no api_key query parameter is attached (openFDA allows anonymous calls).
func TestReadWithoutAPIKeyOmitsParam(t *testing.T) {
	var hadAPIKey bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hadAPIKey = r.URL.Query()["api_key"]
		_, _ = w.Write([]byte(`{"meta":{"results":{"skip":0,"limit":100,"total":1}},"results":[{"safetyreportid":"r1"}]}`))
	}))
	defer srv.Close()

	c := openfda.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "drug_event", Config: cfg}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if hadAPIKey {
		t.Fatal("api_key query param should be absent when no secret is configured")
	}
}

func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := openfda.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "drug_event", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := openfda.New()
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
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q missing fields", s.Name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = openfda.New() // ensure init ran
	c := openfda.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only public API)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("openfda"); !ok {
		t.Fatal("registry did not resolve openfda (self-registration)")
	}
}
