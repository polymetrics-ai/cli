package bingads_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bingads"
)

// TestReadAccountsAuthenticatesAndMaps is the red-first test for the Bing Ads
// connector. It stands up a fake Microsoft OAuth token endpoint plus a fake Bing
// Ads Customer Management REST endpoint and asserts that:
//   - the connector exchanges the refresh_token for an access token,
//   - the API call carries Authorization: Bearer <token> and the DeveloperToken
//     header,
//   - the AccountsInfo array is mapped into records keyed by Id.
func TestReadAccountsAuthenticatesAndMaps(t *testing.T) {
	var (
		sawGrant     string
		sawRefresh   string
		sawAuth      string
		sawDevToken  string
		tokenCalled  int
		accountsHits int
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		tokenCalled++
		_ = r.ParseForm()
		sawGrant = r.Form.Get("grant_type")
		sawRefresh = r.Form.Get("refresh_token")
		_, _ = w.Write([]byte(`{"access_token":"ACCESS_XYZ","token_type":"Bearer","expires_in":3600}`))
	})
	mux.HandleFunc("/CustomerManagement/v13/AccountsInfo/Query", func(w http.ResponseWriter, r *http.Request) {
		accountsHits++
		sawAuth = r.Header.Get("Authorization")
		sawDevToken = r.Header.Get("DeveloperToken")
		_, _ = w.Write([]byte(`{"AccountsInfo":[{"Id":"111","Name":"Acme","Number":"X0001","AccountLifeCycleStatus":"Active"},{"Id":"222","Name":"Globex","Number":"X0002","AccountLifeCycleStatus":"Paused"}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := bingads.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL + "/CustomerManagement/v13",
			"token_url": srv.URL + "/oauth/token",
		},
		Secrets: map[string]string{
			"client_id":       "cid",
			"client_secret":   "csecret",
			"developer_token": "DEVTOKEN123",
			"refresh_token":   "rtok",
			"tenant_id":       "common",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if tokenCalled == 0 {
		t.Fatal("OAuth token endpoint was never called")
	}
	if sawGrant != "refresh_token" {
		t.Fatalf("grant_type = %q, want refresh_token", sawGrant)
	}
	if sawRefresh != "rtok" {
		t.Fatalf("refresh_token = %q, want rtok", sawRefresh)
	}
	if sawAuth != "Bearer ACCESS_XYZ" {
		t.Fatalf("Authorization = %q, want Bearer ACCESS_XYZ", sawAuth)
	}
	if sawDevToken != "DEVTOKEN123" {
		t.Fatalf("DeveloperToken = %q, want DEVTOKEN123", sawDevToken)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["Id"] != "111" || got[0]["Name"] != "Acme" {
		t.Fatalf("record[0] = %+v, want Id=111 Name=Acme", got[0])
	}
}

// TestReadCampaignsSendsAccountHeadersAndBody asserts that campaign-scoped
// streams send the CustomerId / CustomerAccountId headers and the AccountId in
// the POST body, and that the Campaigns array maps correctly. This exercises the
// second stream family (campaign management) and the header/body wiring.
func TestReadCampaignsSendsAccountHeadersAndBody(t *testing.T) {
	var (
		sawCustomerID string
		sawAccountID  string
		sawBody       map[string]any
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"ACCESS_XYZ","token_type":"Bearer","expires_in":3600}`))
	})
	mux.HandleFunc("/CampaignManagement/v13/Campaigns/QueryByAccountId", func(w http.ResponseWriter, r *http.Request) {
		sawCustomerID = r.Header.Get("CustomerId")
		sawAccountID = r.Header.Get("CustomerAccountId")
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		_, _ = w.Write([]byte(`{"Campaigns":[{"Id":"900","Name":"Brand","Status":"Active","CampaignType":"Search"}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := bingads.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"campaign_base_url":   srv.URL + "/CampaignManagement/v13",
			"token_url":           srv.URL + "/oauth/token",
			"customer_id":         "C42",
			"customer_account_id": "777",
		},
		Secrets: map[string]string{
			"client_id":       "cid",
			"developer_token": "DEVTOKEN123",
			"refresh_token":   "rtok",
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
	if sawCustomerID != "C42" {
		t.Fatalf("CustomerId header = %q, want C42", sawCustomerID)
	}
	if sawAccountID != "777" {
		t.Fatalf("CustomerAccountId header = %q, want 777", sawAccountID)
	}
	if sawBody["AccountId"] != "777" {
		t.Fatalf("body AccountId = %v, want 777", sawBody["AccountId"])
	}
	if len(got) != 1 || got[0]["Id"] != "900" || got[0]["Name"] != "Brand" {
		t.Fatalf("campaign records = %+v, want one Id=900 Name=Brand", got)
	}
}

// TestFixtureModeNeedsNoNetwork confirms the credential-free fixture path emits
// deterministic records without any HTTP call, so conformance can run without
// live Microsoft Advertising credentials.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := bingads.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"accounts", "campaigns", "ad_groups", "ads", "users"} {
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
		for _, rec := range got {
			if rec["Id"] == nil {
				t.Fatalf("fixture %s record missing Id: %+v", stream, rec)
			}
		}
	}

	// Check must also short-circuit in fixture mode.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams asserts the published catalog exposes the core streams with
// primary keys.
func TestCatalogStreams(t *testing.T) {
	c := bingads.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"accounts": false, "campaigns": false, "ad_groups": false, "ads": false, "users": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
			if len(s.PrimaryKey) == 0 {
				t.Fatalf("stream %s missing primary key", s.Name)
			}
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestRegistryResolution confirms self-registration under the bing-ads key and
// that the read-only capability set is published.
func TestRegistryResolution(t *testing.T) {
	_ = bingads.New() // ensure init() ran
	if got := bingads.New().Name(); got != "bing-ads" {
		t.Fatalf("Name() = %q, want bing-ads", got)
	}
	r := connectors.NewRegistry()
	conn, ok := r.Get("bing-ads")
	if !ok {
		t.Fatal("registry did not resolve bing-ads (self-registration)")
	}
	caps := conn.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
	if caps.Write {
		t.Fatal("bing-ads is read-only; Write capability must be false")
	}
}

// TestBaseURLValidationRejectsBadScheme guards the SSRF check on base_url
// overrides.
func TestBaseURLValidationRejectsBadScheme(t *testing.T) {
	c := bingads.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"developer_token": "d", "refresh_token": "r", "client_id": "c"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url validation error, got %v", err)
	}
}
