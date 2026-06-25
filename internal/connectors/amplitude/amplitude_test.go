package amplitude_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/amplitude"
)

// TestReadCohortsAuthenticatesAndMaps is the red-first test: HTTP Basic auth
// (api_key:secret_key), the cohorts endpoint, and record mapping out of the
// "cohorts" JSON path.
func TestReadCohortsAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/3/cohorts" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"cohorts":[{"id":"c1","name":"Power users","size":42,"archived":false},{"id":"c2","name":"Churned","size":7,"archived":true}]}`))
	}))
	defer srv.Close()

	c := amplitude.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_123", "secret_key": "sec_456"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "cohorts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("key_123:sec_456"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] != "c1" || got[0]["name"] != "Power users" {
		t.Fatalf("record[0] mismatch: %+v", got[0])
	}
}

// TestReadEventsListPaginatesAcrossTwoEndpoints exercises a second stream
// (events_list) reading from a different endpoint and the "data" JSON path, and
// verifies that records map across what the harness treats as a multi-call read.
func TestReadEventsListMapsDataPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/2/events/list":
			_, _ = w.Write([]byte(`{"data":[{"value":"app_open","display":"App Open","totals":100,"hidden":false},{"value":"purchase","display":"Purchase","totals":5,"hidden":false}]}`))
		case "/api/3/annotations":
			_, _ = w.Write([]byte(`{"data":[{"id":1,"label":"Launch","date":"2026-01-01"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := amplitude.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "key_123", "secret_key": "sec_456"},
	}

	// events_list stream.
	var events []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events_list", Config: cfg}, func(rec connectors.Record) error {
		events = append(events, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read events_list: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("events_list records = %d, want 2", len(events))
	}
	if events[0]["value"] != "app_open" || events[0]["display"] != "App Open" {
		t.Fatalf("events_list record[0] mismatch: %+v", events[0])
	}

	// annotations stream (different endpoint + data path).
	var annotations []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "annotations", Config: cfg}, func(rec connectors.Record) error {
		annotations = append(annotations, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read annotations: %v", err)
	}
	if len(annotations) != 1 || annotations[0]["label"] != "Launch" {
		t.Fatalf("annotations mismatch: %+v", annotations)
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// with no network access, so credential-free conformance can run.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := amplitude.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "cohorts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
	// Check must also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := amplitude.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("amplitude is read-only; Write should be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
}

func TestRegisteredInRegistry(t *testing.T) {
	_ = amplitude.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("amplitude"); !ok {
		t.Fatal("registry did not resolve amplitude (self-registration)")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := amplitude.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "k", "secret_key": "s"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "cohorts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http(s) base_url scheme")
	}
}
