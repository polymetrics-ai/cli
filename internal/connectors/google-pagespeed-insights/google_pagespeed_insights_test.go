package googlepagespeedinsights_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	googlepagespeedinsights "polymetrics.ai/internal/connectors/google-pagespeed-insights"
)

// TestReadIteratesURLsAndAuthenticates is the red-first test for the PageSpeed
// Insights connector. The PageSpeed API has no pagination and no record array:
// each runPagespeed request returns one report for one (url, strategy) pair.
// The connector "paginates" by iterating the configured urls list, issuing one
// request per url and flattening each report into one record. Auth is the
// `key` query parameter. This test configures two urls so the connector makes
// two sequential requests (the natural multi-page analog) and asserts auth, the
// strategy/category/url query params, and record mapping.
func TestReadIteratesURLsAndAuthenticates(t *testing.T) {
	var sawKey string
	var requestedURLs []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/runPagespeed" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		sawKey = q.Get("key")
		requestedURLs = append(requestedURLs, q.Get("url"))
		if q.Get("strategy") != "desktop" {
			t.Errorf("strategy = %q, want desktop", q.Get("strategy"))
		}
		if q.Get("category") != "performance" {
			t.Errorf("category = %q, want performance", q.Get("category"))
		}
		analyzed := q.Get("url")
		_, _ = w.Write([]byte(`{
			"kind":"pagespeedonline#result",
			"id":"` + analyzed + `",
			"analysisUTCTimestamp":"2026-01-01T00:00:00.000Z",
			"lighthouseResult":{
				"requestedUrl":"` + analyzed + `",
				"finalUrl":"` + analyzed + `",
				"lighthouseVersion":"11.0.0",
				"fetchTime":"2026-01-01T00:00:00.000Z",
				"categories":{"performance":{"id":"performance","title":"Performance","score":0.95}}
			},
			"loadingExperience":{"overall_category":"FAST"}
		}`))
	}))
	defer srv.Close()

	c := googlepagespeedinsights.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"urls":       "https://a.example.com,https://b.example.com",
			"strategies": "desktop",
			"categories": "performance",
		},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pagespeed_reports", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "key_test_123" {
		t.Fatalf("key query param = %q, want key_test_123", sawKey)
	}
	if len(requestedURLs) != 2 {
		t.Fatalf("requests = %d, want 2 (one per configured url)", len(requestedURLs))
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["url"] == nil || rec["strategy"] == nil {
			t.Fatalf("record missing url/strategy: %+v", rec)
		}
		if rec["performance_score"] == nil {
			t.Fatalf("record missing performance_score: %+v", rec)
		}
		if rec["analysis_utc_timestamp"] == nil {
			t.Fatalf("record missing analysis_utc_timestamp: %+v", rec)
		}
	}
}

// TestReadFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any network access so conformance can run credential-free.
func TestReadFixtureModeNoNetwork(t *testing.T) {
	c := googlepagespeedinsights.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pagespeed_reports", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode emitted no records")
	}
	for _, rec := range got {
		if rec["url"] == nil || rec["strategy"] == nil {
			t.Fatalf("fixture record missing url/strategy: %+v", rec)
		}
	}
}

// TestCheckFixtureMode verifies Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := googlepagespeedinsights.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams verifies the catalog exposes the report stream with a
// primary key and fields.
func TestCatalogStreams(t *testing.T) {
	c := googlepagespeedinsights.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) == 0 {
		t.Fatal("catalog has no streams")
	}
	var found bool
	for _, s := range cat.Streams {
		if s.Name == "pagespeed_reports" {
			found = true
			if len(s.PrimaryKey) == 0 {
				t.Fatal("pagespeed_reports stream missing primary key")
			}
			if len(s.Fields) == 0 {
				t.Fatal("pagespeed_reports stream missing fields")
			}
		}
	}
	if !found {
		t.Fatal("catalog missing pagespeed_reports stream")
	}
}

// TestRegisteredReadOnly verifies self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	_ = googlepagespeedinsights.New() // ensure init ran
	c := googlepagespeedinsights.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("google-pagespeed-insights"); !ok {
		t.Fatal("registry did not resolve google-pagespeed-insights (self-registration)")
	}
}
