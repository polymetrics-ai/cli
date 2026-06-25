package googleanalyticsdataapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	googleanalyticsdataapi "polymetrics.ai/internal/connectors/google-analytics-data-api"
)

// runReportRequest mirrors the subset of the GA4 runReport request body the
// connector sends, so the test server can assert on offset/limit pagination.
type runReportRequest struct {
	Offset json.Number `json:"offset"`
	Limit  json.Number `json:"limit"`
}

// TestReadPaginatesAndAuthenticates is the red-first test for the GA4 Data API
// connector: Bearer (OAuth2 access_token) auth, offset/limit pagination across 2
// pages of runReport rows, and row->record mapping (dimension/metric headers
// projected onto flat fields). Red until the package exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPath = r.URL.Path
		if r.Method != http.MethodPost {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		var body runReportRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Two-page report: total rowCount=3, limit=2 per page.
		offset := body.Offset.String()
		switch offset {
		case "", "0":
			_, _ = w.Write([]byte(`{
				"dimensionHeaders":[{"name":"date"}],
				"metricHeaders":[{"name":"activeUsers","type":"TYPE_INTEGER"}],
				"rows":[
					{"dimensionValues":[{"value":"20260101"}],"metricValues":[{"value":"10"}]},
					{"dimensionValues":[{"value":"20260102"}],"metricValues":[{"value":"20"}]}
				],
				"rowCount":3
			}`))
		case "2":
			_, _ = w.Write([]byte(`{
				"dimensionHeaders":[{"name":"date"}],
				"metricHeaders":[{"name":"activeUsers","type":"TYPE_INTEGER"}],
				"rows":[
					{"dimensionValues":[{"value":"20260103"}],"metricValues":[{"value":"30"}]}
				],
				"rowCount":3
			}`))
		default:
			t.Errorf("unexpected offset=%q", offset)
			_, _ = w.Write([]byte(`{"rows":[],"rowCount":3}`))
		}
	}))
	defer srv.Close()

	c := googleanalyticsdataapi.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":     srv.URL,
			"property_ids": "123456",
			"page_size":    "2",
		},
		Secrets: map[string]string{"credentials.access_token": "ya29.test-token"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "daily_active_users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer ya29.test-token" {
		t.Fatalf("Authorization = %q, want Bearer ya29.test-token", sawAuth)
	}
	if !strings.Contains(sawPath, "properties/123456:runReport") {
		t.Fatalf("request path = %q, want it to contain properties/123456:runReport", sawPath)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["date"] == nil || rec["activeUsers"] == nil {
			t.Fatalf("record missing mapped dimension/metric: %+v", rec)
		}
		if rec["property_id"] != "123456" {
			t.Fatalf("record property_id = %v, want 123456", rec["property_id"])
		}
	}
}

// TestReadRequiresPropertyID asserts a real (non-fixture) read fails fast when
// property_ids is missing.
func TestReadRequiresPropertyID(t *testing.T) {
	c := googleanalyticsdataapi.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "https://analyticsdata.googleapis.com"},
		Secrets: map[string]string{"credentials.access_token": "tok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "daily_active_users", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read without property_ids should fail")
	}
}

// TestFixtureModeNoNetwork asserts fixture mode emits deterministic records with
// no network access, so conformance works without live creds.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := googleanalyticsdataapi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "daily_active_users", Config: cfg}, func(rec connectors.Record) error {
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
		if rec["date"] == nil {
			t.Fatalf("fixture record missing date: %+v", rec)
		}
	}
	// Check must also short-circuit in fixture mode (no creds).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams asserts the published catalog exposes the core report
// streams with primary keys and cursor fields.
func TestCatalogStreams(t *testing.T) {
	c := googleanalyticsdataapi.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	byName := map[string]connectors.Stream{}
	for _, s := range cat.Streams {
		byName[s.Name] = s
	}
	s, ok := byName["daily_active_users"]
	if !ok {
		t.Fatalf("missing daily_active_users stream; have %v", cat.Streams)
	}
	if len(s.PrimaryKey) == 0 {
		t.Fatal("daily_active_users must declare a primary key")
	}
	if len(s.Fields) == 0 {
		t.Fatal("daily_active_users must declare fields")
	}
}

// TestRegisteredReadOnly asserts self-registration via the registry and that the
// connector advertises read (not write) capability.
func TestRegisteredReadOnly(t *testing.T) {
	_ = googleanalyticsdataapi.New() // ensure init ran
	c := googleanalyticsdataapi.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, GA4 Data API is read-only", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("google-analytics-data-api"); !ok {
		t.Fatal("registry did not resolve google-analytics-data-api (self-registration)")
	}
}
