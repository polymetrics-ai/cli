package amazonads_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	amazonads "polymetrics.ai/internal/connectors/amazon-ads"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts the
// Login with Amazon refresh_token exchange, the required Amazon Ads headers
// (ClientId + Scope + Bearer access token), startIndex/count offset pagination
// across two pages of a top-level JSON array, and record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		sawTokenGrant    string
		sawTokenClient   string
		sawAuth          string
		sawClientID      string
		sawScope         string
		tokenRequests    int
		campaignRequests int
	)

	// Login with Amazon token endpoint.
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenRequests++
		_ = r.ParseForm()
		sawTokenGrant = r.Form.Get("grant_type")
		sawTokenClient = r.Form.Get("client_id")
		if r.Form.Get("refresh_token") == "" {
			t.Errorf("token request missing refresh_token")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"atok_live_123","token_type":"bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	// Amazon Ads API endpoint with startIndex/count pagination.
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawClientID = r.Header.Get("Amazon-Advertising-API-ClientId")
		sawScope = r.Header.Get("Amazon-Advertising-API-Scope")
		if r.URL.Path != "/v2/sp/campaigns" {
			http.NotFound(w, r)
			return
		}
		campaignRequests++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("startIndex") {
		case "", "0":
			_, _ = w.Write([]byte(`[{"campaignId":111,"name":"Camp A","state":"enabled"},{"campaignId":222,"name":"Camp B","state":"paused"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"campaignId":333,"name":"Camp C","state":"enabled"}]`))
		default:
			t.Errorf("unexpected startIndex=%q", r.URL.Query().Get("startIndex"))
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer apiSrv.Close()

	c := amazonads.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   apiSrv.URL,
			"token_url":  tokenSrv.URL,
			"profile_id": "987654321",
			"page_size":  "2",
		},
		Secrets: map[string]string{
			"client_id":     "amzn1.application-oa2-client.abc",
			"client_secret": "supersecret",
			"refresh_token": "Atzr|refresh123",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if tokenRequests == 0 {
		t.Fatal("expected at least one Login with Amazon token exchange")
	}
	if sawTokenGrant != "refresh_token" {
		t.Fatalf("token grant_type = %q, want refresh_token", sawTokenGrant)
	}
	if sawTokenClient != "amzn1.application-oa2-client.abc" {
		t.Fatalf("token client_id = %q, want the configured client id", sawTokenClient)
	}
	if sawAuth != "Bearer atok_live_123" {
		t.Fatalf("Authorization = %q, want Bearer atok_live_123", sawAuth)
	}
	if sawClientID != "amzn1.application-oa2-client.abc" {
		t.Fatalf("Amazon-Advertising-API-ClientId = %q, want the configured client id", sawClientID)
	}
	if sawScope != "987654321" {
		t.Fatalf("Amazon-Advertising-API-Scope = %q, want 987654321", sawScope)
	}
	if campaignRequests != 2 {
		t.Fatalf("campaign requests = %d, want 2 (pagination across 2 pages)", campaignRequests)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["campaign_id"] == nil || got[0]["name"] == nil {
		t.Fatalf("record missing mapped campaign_id/name: %+v", got[0])
	}
}

// TestProfilesNoScopeHeader verifies the profiles stream does NOT send a scope
// header (profiles enumerate the scopes themselves) and maps the top-level array.
func TestProfilesNoScopeHeader(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"atok","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	var sawScope string
	var scopeHeaderPresent bool
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, scopeHeaderPresent = r.Header["Amazon-Advertising-Api-Scope"]
		sawScope = r.Header.Get("Amazon-Advertising-API-Scope")
		_, _ = w.Write([]byte(`[{"profileId":987654321,"countryCode":"US","currencyCode":"USD","timezone":"America/Los_Angeles","accountInfo":{"marketplaceStringId":"ATVPDKIKX0DER","type":"seller"}}]`))
	}))
	defer apiSrv.Close()

	c := amazonads.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": apiSrv.URL, "token_url": tokenSrv.URL},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "csec", "refresh_token": "rtok"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "profiles", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read profiles: %v", err)
	}
	if scopeHeaderPresent || sawScope != "" {
		t.Fatalf("profiles stream must not send a scope header, got %q", sawScope)
	}
	if len(got) != 1 || got[0]["profile_id"] == nil {
		t.Fatalf("profiles mapping failed: %+v", got)
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// without any credentials or network access (conformance without live creds).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := amazonads.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"profiles", "campaigns", "ad_groups", "portfolios"} {
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		}); err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture Read(%s) = %d records, want 2", stream, len(got))
		}
	}
	// Check + Catalog must also work credential-free in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
}

func TestRegistryResolvesAmazonAds(t *testing.T) {
	_ = amazonads.New() // ensure init ran
	r := connectors.NewRegistry()
	conn, ok := r.Get("amazon-ads")
	if !ok {
		t.Fatal("registry did not resolve amazon-ads (self-registration failed)")
	}
	if conn.Name() != "amazon-ads" {
		t.Fatalf("Name() = %q, want amazon-ads", conn.Name())
	}
	caps := conn.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
}

// TestBaseURLSSRFValidation ensures a non-http(s) base_url override is rejected.
func TestBaseURLSSRFValidation(t *testing.T) {
	c := amazonads.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd"},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "csec", "refresh_token": "rtok"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme validation error, got %v", err)
	}
}

// sanity: token response decodes as JSON (guards against fixture drift).
func TestTokenJSONSanity(t *testing.T) {
	var v map[string]any
	if err := json.Unmarshal([]byte(`{"access_token":"x","expires_in":3600}`), &v); err != nil {
		t.Fatal(err)
	}
}
