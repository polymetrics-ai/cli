package mercadoads_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	mercadoads "polymetrics.ai/internal/connectors/mercado-ads"
)

// TestReadAuthenticatesAndMapsRecords is the red-first test: it asserts the
// OAuth2 refresh-token exchange happens, the resulting bearer token is sent on
// the data request, the product_id filter and Api-Version header are applied,
// records are extracted from the "advertisers" array, and mapping populates the
// stream's fields.
func TestReadAuthenticatesAndMapsRecords(t *testing.T) {
	var (
		sawTokenForm   string
		sawDataAuth    string
		sawProductID   string
		sawAPIVersion  string
		tokenRequested bool
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			tokenRequested = true
			_ = r.ParseForm()
			sawTokenForm = r.Form.Encode()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "tok_abc123",
				"token_type":   "bearer",
				"expires_in":   21600,
			})
		case "/advertising/advertisers":
			sawDataAuth = r.Header.Get("Authorization")
			sawProductID = r.URL.Query().Get("product_id")
			sawAPIVersion = r.Header.Get("Api-Version")
			_, _ = w.Write([]byte(`{"advertisers":[{"advertiser_id":111,"advertiser_name":"Acme","account_name":"acme-co","site_id":"MLB"},{"advertiser_id":222,"advertiser_name":"Beta","account_name":"beta-co","site_id":"MLA"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := mercadoads.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":      srv.URL,
			"token_url":     srv.URL + "/oauth/token",
			"lookback_days": "7",
		},
		Secrets: map[string]string{
			"client_id":            "cid",
			"client_secret":        "csecret",
			"client_refresh_token": "rtoken",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "brand_advertisers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !tokenRequested {
		t.Fatal("token endpoint was never called")
	}
	if !strings.Contains(sawTokenForm, "grant_type=refresh_token") {
		t.Fatalf("token form = %q, want grant_type=refresh_token", sawTokenForm)
	}
	if !strings.Contains(sawTokenForm, "refresh_token=rtoken") {
		t.Fatalf("token form = %q, want refresh_token=rtoken", sawTokenForm)
	}
	if sawDataAuth != "Bearer tok_abc123" {
		t.Fatalf("data Authorization = %q, want Bearer tok_abc123", sawDataAuth)
	}
	if sawProductID != "BADS" {
		t.Fatalf("product_id = %q, want BADS", sawProductID)
	}
	if sawAPIVersion != "1" {
		t.Fatalf("Api-Version = %q, want 1", sawAPIVersion)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["advertiser_id"] == nil || got[0]["advertiser_name"] != "Acme" {
		t.Fatalf("first record not mapped: %+v", got[0])
	}
}

// TestReadPaginatesMetrics asserts the offset/limit paginator walks two pages of
// the metrics endpoint and stops on a short page.
func TestReadPaginatesMetrics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/oauth/token":
			_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
		case strings.HasSuffix(r.URL.Path, "/metrics"):
			off := r.URL.Query().Get("offset")
			switch off {
			case "", "0":
				// full page of 100 -> paginator continues
				var items []string
				for i := 0; i < 100; i++ {
					items = append(items, `{"date":"2026-01-01","prints":5,"clicks":1,"cost":2.5}`)
				}
				_, _ = w.Write([]byte(`{"metrics":[` + strings.Join(items, ",") + `]}`))
			case "100":
				// short page -> stop
				_, _ = w.Write([]byte(`{"metrics":[{"date":"2026-01-02","prints":3,"clicks":0,"cost":1.0}]}`))
			default:
				t.Errorf("unexpected offset=%q", off)
				_, _ = w.Write([]byte(`{"metrics":[]}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := mercadoads.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":      srv.URL,
			"token_url":     srv.URL + "/oauth/token",
			"lookback_days": "7",
		},
		Secrets: map[string]string{
			"client_id":            "cid",
			"client_secret":        "csecret",
			"client_refresh_token": "rtoken",
		},
	}

	count := 0
	err := c.Read(context.Background(), connectors.ReadRequest{
		Stream: "brand_campaigns_metrics",
		Config: cfg,
		State:  map[string]string{"advertiser_id": "111", "campaign_id": "999"},
	}, func(rec connectors.Record) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if count != 101 {
		t.Fatalf("records = %d, want 101 (2 pages)", count)
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any credentials or network access.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := mercadoads.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "brand_advertisers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if got[0]["advertiser_id"] == nil {
		t.Fatalf("fixture record missing advertiser_id: %+v", got[0])
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := mercadoads.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	names := map[string]bool{}
	for _, s := range cat.Streams {
		names[s.Name] = true
	}
	for _, want := range []string{"brand_advertisers", "display_advertisers", "product_advertisers"} {
		if !names[want] {
			t.Fatalf("catalog missing stream %q", want)
		}
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	c := mercadoads.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("mercado-ads"); !ok {
		t.Fatal("registry did not resolve mercado-ads (self-registration)")
	}
}
