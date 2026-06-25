package googlesearchconsole_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	googlesearchconsole "polymetrics.ai/internal/connectors/google-search-console"
)

// TestSearchAnalyticsPaginatesAndAuthenticates is the red-first test: it asserts
// the OAuth Bearer header is sent, that the search-analytics stream pages through
// two responses via startRow offset, and that the rows are flattened into records
// keyed by the requested dimension.
func TestSearchAnalyticsPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawDimensions []string
	var sawStartRows []float64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		const wantPath = "/sites/https:%2F%2Fexample.com%2F/searchAnalytics/query"
		if r.URL.EscapedPath() != wantPath {
			t.Errorf("path = %q, want %q", r.URL.EscapedPath(), wantPath)
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		var body struct {
			StartDate  string   `json:"startDate"`
			EndDate    string   `json:"endDate"`
			Dimensions []string `json:"dimensions"`
			RowLimit   int      `json:"rowLimit"`
			StartRow   float64  `json:"startRow"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		sawDimensions = body.Dimensions
		sawStartRows = append(sawStartRows, body.StartRow)
		if body.RowLimit != 2 {
			t.Errorf("rowLimit = %d, want 2", body.RowLimit)
		}
		switch body.StartRow {
		case 0:
			_, _ = w.Write([]byte(`{"rows":[
				{"keys":["2026-06-01"],"clicks":10,"impressions":100,"ctr":0.1,"position":3.2},
				{"keys":["2026-06-02"],"clicks":5,"impressions":50,"ctr":0.1,"position":4.5}
			]}`))
		case 2:
			_, _ = w.Write([]byte(`{"rows":[
				{"keys":["2026-06-03"],"clicks":1,"impressions":20,"ctr":0.05,"position":7.0}
			]}`))
		default:
			t.Errorf("unexpected startRow=%v", body.StartRow)
			_, _ = w.Write([]byte(`{"rows":[]}`))
		}
	}))
	defer srv.Close()

	c := googlesearchconsole.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"site_urls":  "https://example.com/",
			"start_date": "2026-06-01",
			"end_date":   "2026-06-30",
			"page_size":  "2",
		},
		Secrets: map[string]string{"authorization.access_token": "ya29.test_token"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "search_analytics_by_date", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer ya29.test_token" {
		t.Fatalf("Authorization = %q, want Bearer ya29.test_token", sawAuth)
	}
	if len(sawDimensions) != 1 || sawDimensions[0] != "date" {
		t.Fatalf("dimensions = %v, want [date]", sawDimensions)
	}
	if len(sawStartRows) != 2 || sawStartRows[0] != 0 || sawStartRows[1] != 2 {
		t.Fatalf("startRows = %v, want [0 2]", sawStartRows)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["date"] == nil || rec["clicks"] == nil {
			t.Fatalf("record missing date/clicks: %+v", rec)
		}
		if rec["site_url"] != "https://example.com/" {
			t.Fatalf("record site_url = %v, want https://example.com/", rec["site_url"])
		}
		if rec["search_type"] != "web" {
			t.Fatalf("record search_type = %v, want web", rec["search_type"])
		}
	}
}

// TestSitesListAuthAndMapping covers the GET-based sites stream: auth header, the
// siteEntry array extraction, and record mapping.
func TestSitesListAuthAndMapping(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/sites" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"siteEntry":[
			{"siteUrl":"https://example.com/","permissionLevel":"siteOwner"},
			{"siteUrl":"sc-domain:example.org","permissionLevel":"siteFullUser"}
		]}`))
	}))
	defer srv.Close()

	c := googlesearchconsole.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"authorization.access_token": "ya29.test_token"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sites", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer ya29.test_token" {
		t.Fatalf("Authorization = %q, want Bearer ya29.test_token", sawAuth)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["site_url"] != "https://example.com/" || got[0]["permission_level"] != "siteOwner" {
		t.Fatalf("first record mapped wrong: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no network access, which conformance relies on to run without credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := googlesearchconsole.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"sites", "sitemaps", "search_analytics_by_date", "search_analytics_by_query"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}

	// Check in fixture mode must not require credentials and must not hit network.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCheckRequiresToken ensures non-fixture Check rejects a missing access token
// before any network call.
func TestCheckRequiresToken(t *testing.T) {
	c := googlesearchconsole.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check with no access_token should fail")
	}
}

// TestCatalogStreams ensures the catalog exposes the core streams with primary
// keys defined.
func TestCatalogStreams(t *testing.T) {
	c := googlesearchconsole.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{
		"sites":                       false,
		"sitemaps":                    false,
		"search_analytics_by_date":    false,
		"search_analytics_by_query":   false,
		"search_analytics_by_page":    false,
		"search_analytics_by_country": false,
	}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Errorf("stream %q has no primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("catalog missing expected stream %q", name)
		}
	}
}

// TestBaseURLValidation rejects malformed base_url overrides to bound SSRF risk.
func TestBaseURLValidation(t *testing.T) {
	c := googlesearchconsole.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil"},
		Secrets: map[string]string{"authorization.access_token": "ya29.test_token"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sites", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with bad base_url err = %v, want base_url scheme error", err)
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	c := googlesearchconsole.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("google-search-console"); !ok {
		t.Fatal("registry did not resolve google-search-console (self-registration)")
	}
}
