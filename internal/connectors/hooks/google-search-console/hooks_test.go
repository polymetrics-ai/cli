package googlesearchconsole

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func newRuntime(baseURL string) *engine.Runtime {
	return &engine.Runtime{Requester: &connsdk.Requester{BaseURL: baseURL}}
}

// --- registration ---

func TestInit_RegistersHooks(t *testing.T) {
	h := engine.HooksFor("google-search-console")
	if h == nil {
		t.Fatal(`engine.HooksFor("google-search-console") = nil, want a registered hook set (init() must call engine.RegisterHooks)`)
	}
	if h.ConnectorName() != "google-search-console" {
		t.Fatalf("ConnectorName() = %q, want google-search-console", h.ConnectorName())
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered hooks do not implement StreamHook")
	}
}

// --- ReadStream dispatch ---

func TestReadStream_UnknownStreamFallsBackToDeclarative(t *testing.T) {
	h := &Hooks{}
	for _, name := range []string{"sites", "sitemaps", "not_a_real_stream", ""} {
		handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: name}, connectors.ReadRequest{}, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
		if err != nil {
			t.Fatalf("ReadStream(%q): %v", name, err)
		}
		if handled {
			t.Fatalf("ReadStream(%q) handled = true, want false (declarative fallback)", name)
		}
	}
}

// TestReadStream_PaginatesAndAuthenticates is the red-first test: it asserts
// the search-analytics stream pages through two responses via startRow
// offset carried inside the POST body, and that rows are flattened into
// records keyed by the requested dimension. Auth is applied by rt.Requester
// (built by the engine, exercised here directly), not by the hook itself.
func TestReadStream_PaginatesAndAuthenticates(t *testing.T) {
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

	rt := &engine.Runtime{Requester: &connsdk.Requester{BaseURL: srv.URL, Auth: connsdk.Bearer("ya29.test_token")}}
	req := connectors.ReadRequest{
		Stream: "search_analytics_by_date",
		Config: connectors.RuntimeConfig{
			Config: map[string]string{
				"site_urls":  "https://example.com/",
				"start_date": "2026-06-01",
				"end_date":   "2026-06-30",
				"page_size":  "2",
			},
		},
	}

	h := &Hooks{}
	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "search_analytics_by_date"}, req, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true for search_analytics_by_date")
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

// TestReadStream_MultipleSitesFanOut verifies the hook loops over every
// configured site_urls entry.
func TestReadStream_MultipleSitesFanOut(t *testing.T) {
	var sawPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPaths = append(sawPaths, r.URL.Path)
		_, _ = w.Write([]byte(`{"rows":[]}`))
	}))
	defer srv.Close()

	rt := &engine.Runtime{Requester: &connsdk.Requester{BaseURL: srv.URL}}
	req := connectors.ReadRequest{
		Config: connectors.RuntimeConfig{
			Config: map[string]string{"site_urls": "https://a.example.com/,https://b.example.com/"},
		},
	}
	h := &Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "search_analytics_by_query"}, req, rt, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if len(sawPaths) != 2 {
		t.Fatalf("requests = %d, want 2 (one per site)", len(sawPaths))
	}
}

// TestGscMaxPages mirrors legacy's gscMaxPages parse-tolerance matrix.
func TestGscMaxPages(t *testing.T) {
	cases := []struct {
		raw     string
		want    int
		wantErr bool
	}{
		{"", 0, false},
		{"all", 0, false},
		{"UNLIMITED", 0, false},
		{"3", 3, false},
		{"-1", 0, true},
		{"not-a-number", 0, true},
	}
	for _, c := range cases {
		got, err := gscMaxPages(connectors.RuntimeConfig{Config: map[string]string{"max_pages": c.raw}})
		if c.wantErr {
			if err == nil {
				t.Errorf("gscMaxPages(%q): want error, got nil", c.raw)
			}
			continue
		}
		if err != nil {
			t.Errorf("gscMaxPages(%q): unexpected error: %v", c.raw, err)
		}
		if got != c.want {
			t.Errorf("gscMaxPages(%q) = %d, want %d", c.raw, got, c.want)
		}
	}
}

// TestAnalyticsDateRangeDefaults verifies default start/end date resolution.
func TestAnalyticsDateRangeDefaults(t *testing.T) {
	start, end, err := analyticsDateRange(connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: map[string]string{}}})
	if err != nil {
		t.Fatalf("analyticsDateRange: %v", err)
	}
	if start != "2021-01-01" {
		t.Fatalf("start = %q, want 2021-01-01", start)
	}
	if end == "" {
		t.Fatal("end date should default to today (UTC), got empty")
	}
}

// TestAnalyticsDateRangeCursorOverridesStartDate verifies the incremental
// cursor (a previously-synced date) takes priority over start_date config.
func TestAnalyticsDateRangeCursorOverridesStartDate(t *testing.T) {
	req := connectors.ReadRequest{
		Config: connectors.RuntimeConfig{Config: map[string]string{"start_date": "2020-01-01"}},
		State:  map[string]string{"cursor": "2026-05-01"},
	}
	start, _, err := analyticsDateRange(req)
	if err != nil {
		t.Fatalf("analyticsDateRange: %v", err)
	}
	if start != "2026-05-01" {
		t.Fatalf("start = %q, want 2026-05-01 (cursor should override start_date)", start)
	}
}
