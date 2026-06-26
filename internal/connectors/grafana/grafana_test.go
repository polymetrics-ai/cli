package grafana_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/grafana"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Grafana
// connector: Bearer auth on the api_key secret, page/limit pagination over the
// top-level array returned by /api/search, and record mapping. Red until
// internal/connectors/grafana exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/search" {
			http.NotFound(w, r)
			return
		}
		sawType = r.URL.Query().Get("type")
		page := r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case "1":
			// Full page (limit defaults to 2 in the test config below) -> a
			// next page must be requested.
			_, _ = w.Write([]byte(`[{"id":1,"uid":"abc","title":"Ops","type":"dash-db"},{"id":2,"uid":"def","title":"Sales","type":"dash-db"}]`))
		case "2":
			// Short page -> pagination stops here.
			_, _ = w.Write([]byte(`[{"id":3,"uid":"ghi","title":"Infra","type":"dash-db"}]`))
		default:
			t.Errorf("unexpected page=%q", page)
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	c := grafana.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "glsa_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "dashboards", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer glsa_test_123" {
		t.Fatalf("Authorization = %q, want Bearer glsa_test_123", sawAuth)
	}
	if sawType != "dash-db" {
		t.Fatalf("type = %q, want dash-db", sawType)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["uid"] == nil {
			t.Fatalf("record missing id/uid: %+v", rec)
		}
	}
}

// TestReadDatasourcesSinglePage exercises a non-paginated top-level-array stream.
func TestReadDatasourcesSinglePage(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/datasources" {
			http.NotFound(w, r)
			return
		}
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"uid":"ds1","name":"Prometheus","type":"prometheus","isDefault":true},{"id":2,"uid":"ds2","name":"Loki","type":"loki","isDefault":false}]`))
	}))
	defer srv.Close()

	c := grafana.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "glsa_test_123"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "datasources", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if calls != 1 {
		t.Fatalf("datasources requested %d times, want 1 (no pagination)", calls)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["name"] != "Prometheus" {
		t.Fatalf("record name = %v, want Prometheus", got[0]["name"])
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network access (mode=fixture), so conformance works without creds.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := grafana.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"dashboards", "folders", "datasources", "org_users", "alert_rules"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s emitted no records", stream)
		}
	}
	// Check short-circuits without creds in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := grafana.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	for _, s := range cat.Streams {
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %s missing primary key", s.Name)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	c := grafana.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write == false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("grafana"); !ok {
		t.Fatal("registry did not resolve grafana (self-registration)")
	}
}
