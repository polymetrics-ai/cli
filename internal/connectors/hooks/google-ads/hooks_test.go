// Package googleads's hooks_test.go proves hooks.go's StreamHook ports
// legacy internal/connectors/google-ads/google_ads.go's
// readAccessibleCustomers/search/mapRecord functions byte-for-byte, and that
// auth (Bearer + developer-token + optional login-customer-id) resolves
// declaratively via the engine's real auth/header pipeline.
package googleads

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
	h := engine.HooksFor("google-ads")
	if h == nil {
		t.Fatal(`engine.HooksFor("google-ads") = nil, want a registered hook set (init() must call engine.RegisterHooks)`)
	}
	if h.ConnectorName() != "google-ads" {
		t.Fatalf("ConnectorName() = %q, want google-ads", h.ConnectorName())
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered hooks do not implement StreamHook")
	}
	if _, ok := h.(engine.AuthHook); ok {
		t.Fatal("registered hooks implement AuthHook, want none (auth is fully declarative via bearer mode + base.headers)")
	}
}

// --- ReadStream dispatch ---

func TestReadStream_UnknownStreamFallsBackToDeclarative(t *testing.T) {
	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "not_a_real_stream"}, connectors.ReadRequest{}, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("handled = true for an unrecognized stream name, want false (declarative fallback)")
	}
}

func TestReadStream_EmptyStreamNameDefaultsToAccessibleCustomers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/customers:listAccessibleCustomers" {
			t.Errorf("request = %s %s, want GET /customers:listAccessibleCustomers", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"resourceNames":[]}`))
	}))
	defer srv.Close()

	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: ""}, connectors.ReadRequest{}, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false for empty stream name, want true (defaults to accessible_customers)")
	}
}

// --- accessible_customers ---

func TestReadStream_AccessibleCustomersDerivesCustomerIDFromResourceName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/customers:listAccessibleCustomers" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"resourceNames":["customers/1111111111","customers/2222222222"]}`))
	}))
	defer srv.Close()

	var got []connectors.Record
	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "accessible_customers"}, connectors.ReadRequest{}, newRuntime(srv.URL), func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["customer_id"] != "1111111111" || got[0]["resource_name"] != "customers/1111111111" {
		t.Fatalf("record 0 = %+v", got[0])
	}
	if got[1]["customer_id"] != "2222222222" || got[1]["resource_name"] != "customers/2222222222" {
		t.Fatalf("record 1 = %+v", got[1])
	}
}

func TestReadStream_AccessibleCustomersEmptyResourceNamesEmitsNothing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"resourceNames":[]}`))
	}))
	defer srv.Close()

	var got []connectors.Record
	h := Hooks{}
	if _, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "accessible_customers"}, connectors.ReadRequest{}, newRuntime(srv.URL), func(r connectors.Record) error {
		got = append(got, r)
		return nil
	}); err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("records = %+v, want none", got)
	}
}

// --- campaigns / ad_groups: GAQL search body + pagination ---

func TestReadStream_CampaignsRequiresCustomerID(t *testing.T) {
	h := Hooks{}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "campaigns"}, connectors.ReadRequest{Config: connectors.RuntimeConfig{}}, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream: want error when customer_id is unset, got nil")
	}
}

type gaqlSearchRequest struct {
	Query     string `json:"query"`
	PageSize  int    `json:"pageSize"`
	PageToken string `json:"pageToken"`
}

func TestReadStream_CampaignsPaginatesViaBodyPageToken(t *testing.T) {
	var sawMethod, sawPath string
	var sawTokens []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		var body gaqlSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Query == "" {
			t.Fatal("query is empty, want the fixed allow-listed GAQL campaign query")
		}
		sawTokens = append(sawTokens, body.PageToken)
		switch body.PageToken {
		case "":
			_, _ = w.Write([]byte(`{"results":[{"campaign":{"id":"111","name":"Brand","status":"ENABLED","resourceName":"customers/1234567890/campaigns/111"}}],"nextPageToken":"next-page"}`))
		case "next-page":
			_, _ = w.Write([]byte(`{"results":[{"campaign":{"id":"222","name":"Demand","status":"PAUSED","resourceName":"customers/1234567890/campaigns/222"}}]}`))
		default:
			t.Fatalf("unexpected pageToken=%q", body.PageToken)
		}
	}))
	defer srv.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{"customer_id": "1234567890", "page_size": "1"}}
	var got []connectors.Record
	h := Hooks{}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "campaigns"}, connectors.ReadRequest{Config: cfg}, newRuntime(srv.URL), func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if sawMethod != http.MethodPost || sawPath != "/customers/1234567890/googleAds:search" {
		t.Fatalf("request = %s %s, want POST /customers/1234567890/googleAds:search", sawMethod, sawPath)
	}
	if len(sawTokens) != 2 || sawTokens[0] != "" || sawTokens[1] != "next-page" {
		t.Fatalf("page tokens = %v", sawTokens)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] != "111" || got[0]["name"] != "Brand" || got[0]["status"] != "ENABLED" || got[0]["resource_name"] != "customers/1234567890/campaigns/111" {
		t.Fatalf("record 0 = %+v", got[0])
	}
	if got[1]["id"] != "222" || got[1]["name"] != "Demand" || got[1]["status"] != "PAUSED" {
		t.Fatalf("record 1 = %+v", got[1])
	}
}

func TestReadStream_AdGroupsMapsNestedAdGroupField(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body gaqlSearchRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = w.Write([]byte(`{"results":[{"adGroup":{"id":"999","name":"Fixture Ad Group","status":"ENABLED","resourceName":"customers/1234567890/adGroups/999"}}]}`))
	}))
	defer srv.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{"customer_id": "1234567890"}}
	var got []connectors.Record
	h := Hooks{}
	if _, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "ad_groups"}, connectors.ReadRequest{Config: cfg}, newRuntime(srv.URL), func(r connectors.Record) error {
		got = append(got, r)
		return nil
	}); err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "999" || got[0]["name"] != "Fixture Ad Group" {
		t.Fatalf("records = %+v", got)
	}
}

func TestReadStream_PageSizeOutOfRangeErrors(t *testing.T) {
	cfg := connectors.RuntimeConfig{Config: map[string]string{"customer_id": "1234567890", "page_size": "999999"}}
	h := Hooks{}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "campaigns"}, connectors.ReadRequest{Config: cfg}, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream: want error for out-of-range page_size, got nil")
	}
}

func TestReadStream_MaxPagesBoundsRequestCount(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"results":[{"campaign":{"id":"1","name":"x","status":"ENABLED"}}],"nextPageToken":"always-more"}`))
	}))
	defer srv.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{"customer_id": "1234567890", "max_pages": "2"}}
	h := Hooks{}
	if _, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "campaigns"}, connectors.ReadRequest{Config: cfg}, newRuntime(srv.URL), func(connectors.Record) error { return nil }); err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2 (max_pages cap)", requests)
	}
}

// --- resolvePageSize/resolveMaxPages ---

func TestResolvePageSize(t *testing.T) {
	cases := []struct {
		raw     string
		want    int
		wantErr bool
	}{
		{"", defaultPageSize, false},
		{"500", 500, false},
		{"10000", 10000, false},
		{"0", 0, true},
		{"10001", 0, true},
		{"not-a-number", 0, true},
	}
	for _, c := range cases {
		got, err := resolvePageSize(connectors.RuntimeConfig{Config: map[string]string{"page_size": c.raw}})
		if c.wantErr {
			if err == nil {
				t.Errorf("resolvePageSize(%q): want error, got nil", c.raw)
			}
			continue
		}
		if err != nil {
			t.Errorf("resolvePageSize(%q): %v", c.raw, err)
			continue
		}
		if got != c.want {
			t.Errorf("resolvePageSize(%q) = %d, want %d", c.raw, got, c.want)
		}
	}
}

func TestResolveMaxPages(t *testing.T) {
	cases := []struct {
		raw     string
		want    int
		wantErr bool
	}{
		{"", 0, false},
		{"all", 0, false},
		{"unlimited", 0, false},
		{"5", 5, false},
		{"-1", 0, true},
		{"nope", 0, true},
	}
	for _, c := range cases {
		got, err := resolveMaxPages(connectors.RuntimeConfig{Config: map[string]string{"max_pages": c.raw}})
		if c.wantErr {
			if err == nil {
				t.Errorf("resolveMaxPages(%q): want error, got nil", c.raw)
			}
			continue
		}
		if err != nil {
			t.Errorf("resolveMaxPages(%q): %v", c.raw, err)
			continue
		}
		if got != c.want {
			t.Errorf("resolveMaxPages(%q) = %d, want %d", c.raw, got, c.want)
		}
	}
}
