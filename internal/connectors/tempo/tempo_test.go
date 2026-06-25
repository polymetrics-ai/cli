package tempo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/tempo"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Tempo
// connector: Bearer auth with the api_token, Tempo v4 metadata.next pagination
// over results[], and record mapping. Two pages are served; the second is
// reached only by following the absolute metadata.next URL the first returns.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if !strings.HasPrefix(r.URL.Path, "/worklogs") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("offset") == "2" {
			_, _ = w.Write([]byte(`{"results":[{"tempoWorklogId":3,"timeSpentSeconds":600,"startDate":"2026-01-03"}],"metadata":{"count":1,"offset":2,"limit":2}}`))
			return
		}
		// First page advertises a next link (absolute URL) with offset=2.
		_, _ = w.Write([]byte(`{"results":[{"tempoWorklogId":1,"timeSpentSeconds":3600,"startDate":"2026-01-01"},{"tempoWorklogId":2,"timeSpentSeconds":1800,"startDate":"2026-01-02"}],"metadata":{"count":2,"offset":0,"limit":2,"next":"` + srv.URL + `/worklogs?offset=2&limit=2"}}`))
	}))
	defer srv.Close()

	c := tempo.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_token": "tok_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "worklogs", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer tok_test_123" {
		t.Fatalf("Authorization = %q, want Bearer tok_test_123", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["tempo_worklog_id"] == nil {
			t.Fatalf("record missing tempo_worklog_id: %+v", rec)
		}
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network access, so conformance runs without live credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := tempo.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"accounts", "customers", "worklogs", "workload-schemes"} {
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
	}
}

// TestCheckFixtureMode verifies Check short-circuits in fixture mode and rejects
// a missing api_token in live mode.
func TestCheckFixtureMode(t *testing.T) {
	c := tempo.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	err := c.Check(context.Background(), connectors.RuntimeConfig{})
	if err == nil {
		t.Fatal("Check without api_token should fail")
	}
}

// TestCatalogStreams verifies the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	c := tempo.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"accounts": false, "customers": false, "worklogs": false, "workload-schemes": false}
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

// TestRegisteredReadOnly verifies self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	c := tempo.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("tempo"); !ok {
		t.Fatal("registry did not resolve tempo (self-registration)")
	}
}
