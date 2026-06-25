package applesearchads_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	applesearchads "polymetrics/internal/connectors/apple-search-ads"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Apple Search
// Ads connector. It asserts the OAuth2 client-credentials token exchange, the
// Authorization bearer header on the API call, the required X-AP-Context org
// header, Apple's offset/limit pagination across two pages of a {data,
// pagination} envelope, and record mapping for the campaigns stream.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		sawTokenGrant  string
		sawTokenClient string
		tokenRequests  int
		sawAuth        string
		sawOrgContext  string
		campaignPages  int
	)

	// Apple ID OAuth2 token endpoint (client_credentials grant).
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenRequests++
		_ = r.ParseForm()
		sawTokenGrant = r.Form.Get("grant_type")
		sawTokenClient = r.Form.Get("client_id")
		if r.Form.Get("client_secret") == "" {
			t.Errorf("token request missing client_secret")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"asa_live_token","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	// Apple Search Ads campaign management API endpoint with offset/limit
	// pagination returning the {data, pagination} envelope.
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawOrgContext = r.Header.Get("X-AP-Context")
		if r.URL.Path != "/campaigns" {
			http.NotFound(w, r)
			return
		}
		campaignPages++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"data":[{"id":111,"name":"Camp A","status":"ENABLED","modificationTime":"2026-01-01T00:00:00.000Z"},{"id":222,"name":"Camp B","status":"PAUSED","modificationTime":"2026-01-02T00:00:00.000Z"}],"pagination":{"totalResults":3,"startIndex":0,"itemsPerPage":2}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":333,"name":"Camp C","status":"ENABLED","modificationTime":"2026-01-03T00:00:00.000Z"}],"pagination":{"totalResults":3,"startIndex":2,"itemsPerPage":2}}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"data":[],"pagination":{"totalResults":3,"startIndex":4,"itemsPerPage":2}}`))
		}
	}))
	defer apiSrv.Close()

	c := applesearchads.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":               apiSrv.URL,
			"token_refresh_endpoint": tokenSrv.URL,
			"org_id":                 "123456",
			"page_size":              "2",
		},
		Secrets: map[string]string{
			"client_id":     "SEARCHADS.client.abc",
			"client_secret": "supersecretjwt",
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
		t.Fatal("expected at least one OAuth2 token exchange")
	}
	if sawTokenGrant != "client_credentials" {
		t.Fatalf("token grant_type = %q, want client_credentials", sawTokenGrant)
	}
	if sawTokenClient != "SEARCHADS.client.abc" {
		t.Fatalf("token client_id = %q, want the configured client id", sawTokenClient)
	}
	if sawAuth != "Bearer asa_live_token" {
		t.Fatalf("Authorization = %q, want Bearer asa_live_token", sawAuth)
	}
	if sawOrgContext != "orgId=123456" {
		t.Fatalf("X-AP-Context = %q, want orgId=123456", sawOrgContext)
	}
	if campaignPages != 2 {
		t.Fatalf("campaign requests = %d, want 2 (pagination across 2 pages)", campaignPages)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["id"] == nil || got[0]["name"] == nil || got[0]["modification_time"] == nil {
		t.Fatalf("record missing mapped id/name/modification_time: %+v", got[0])
	}
}

// TestFindStreamPaginates verifies the POST .../find streams (adgroups) send a
// pagination selector body, carry the org context + bearer auth, and paginate
// across two pages using pagination.totalResults.
func TestFindStreamPaginates(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"asa_tok","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	var (
		findPages    int
		sawMethod    string
		sawOffsets   []json.Number
		sawOrgHeader string
	)
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/adgroups/find" {
			http.NotFound(w, r)
			return
		}
		findPages++
		sawMethod = r.Method
		sawOrgHeader = r.Header.Get("X-AP-Context")
		var body struct {
			Pagination struct {
				Offset json.Number `json:"offset"`
				Limit  json.Number `json:"limit"`
			} `json:"pagination"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		sawOffsets = append(sawOffsets, body.Pagination.Offset)
		w.Header().Set("Content-Type", "application/json")
		switch body.Pagination.Offset.String() {
		case "", "0":
			_, _ = w.Write([]byte(`{"data":[{"id":11,"campaignId":111,"name":"AG A","status":"ENABLED","modificationTime":"2026-01-01T00:00:00.000Z"},{"id":12,"campaignId":111,"name":"AG B","status":"ENABLED","modificationTime":"2026-01-01T00:00:00.000Z"}],"pagination":{"totalResults":3,"startIndex":0,"itemsPerPage":2}}`))
		case "2":
			_, _ = w.Write([]byte(`{"data":[{"id":13,"campaignId":111,"name":"AG C","status":"PAUSED","modificationTime":"2026-01-02T00:00:00.000Z"}],"pagination":{"totalResults":3,"startIndex":2,"itemsPerPage":2}}`))
		default:
			t.Errorf("unexpected offset=%q", body.Pagination.Offset.String())
		}
	}))
	defer apiSrv.Close()

	c := applesearchads.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":               apiSrv.URL,
			"token_refresh_endpoint": tokenSrv.URL,
			"org_id":                 "777",
			"page_size":              "2",
		},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "csec"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "adgroups", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read adgroups: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Fatalf("find method = %q, want POST", sawMethod)
	}
	if sawOrgHeader != "orgId=777" {
		t.Fatalf("X-AP-Context = %q, want orgId=777", sawOrgHeader)
	}
	if findPages != 2 {
		t.Fatalf("find requests = %d, want 2 (pagination across 2 pages)", findPages)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[2]["campaign_id"] == nil || got[2]["name"] == nil {
		t.Fatalf("adgroup record missing mapped campaign_id/name: %+v", got[2])
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records for
// every stream without credentials or network access (conformance without live
// creds), and that Check + Catalog work credential-free in fixture mode.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := applesearchads.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"campaigns", "adgroups", "keywords", "ads"} {
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
		if got[0]["id"] == nil {
			t.Fatalf("fixture Read(%s) record missing id: %+v", stream, got[0])
		}
	}
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

// TestRegistryResolvesAppleSearchAds confirms self-registration and the read-only
// capability profile.
func TestRegistryResolvesAppleSearchAds(t *testing.T) {
	_ = applesearchads.New() // ensure init ran
	r := connectors.NewRegistry()
	conn, ok := r.Get("apple-search-ads")
	if !ok {
		t.Fatal("registry did not resolve apple-search-ads (self-registration failed)")
	}
	if conn.Name() != "apple-search-ads" {
		t.Fatalf("Name() = %q, want apple-search-ads", conn.Name())
	}
	caps := conn.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
	if caps.Write {
		t.Fatalf("apple-search-ads should be read-only, got Write=true")
	}
}

// TestBaseURLSSRFValidation ensures a non-http(s) base_url override is rejected.
func TestBaseURLSSRFValidation(t *testing.T) {
	c := applesearchads.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "org_id": "1"},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "csec"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme validation error, got %v", err)
	}
}

// TestMissingOrgID ensures the connector rejects a non-fixture read without org_id.
func TestMissingOrgID(t *testing.T) {
	c := applesearchads.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "https://example.com", "token_refresh_endpoint": "https://example.com/token"},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "csec"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "org_id") {
		t.Fatalf("expected org_id validation error, got %v", err)
	}
}
