package paypaltransaction

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

// --- registration -----------------------------------------------------

func TestHooksRegisteredUnderPaypalTransaction(t *testing.T) {
	h := engine.HooksFor("paypal-transaction")
	if h == nil {
		t.Fatal(`engine.HooksFor("paypal-transaction") = nil, want registered hooks (hooks/paypal-transaction's init() must call engine.RegisterHooks)`)
	}
	if h.ConnectorName() != "paypal-transaction" {
		t.Fatalf("ConnectorName() = %q, want %q", h.ConnectorName(), "paypal-transaction")
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("registered paypal-transaction hooks does not implement engine.AuthHook")
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered paypal-transaction hooks does not implement engine.StreamHook")
	}
}

// --- AuthHook: Basic-credential-in / Bearer-out token exchange ---------

// TestAuthenticator_ExchangesBasicForBearerAndCaches mirrors legacy's
// TestReadPaginatesAndAuthenticates auth assertions: the token endpoint sees
// HTTP Basic client_id:client_secret and grant_type=client_credentials in the
// form body, the returned access_token is used as Authorization: Bearer on
// subsequent requests, and a second Apply call within the cached TTL does not
// re-hit the token endpoint.
func TestAuthenticator_ExchangesBasicForBearerAndCaches(t *testing.T) {
	var sawMethod, sawBasicAuth, sawGrantType, sawContentType string
	tokenHits := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/oauth2/token" {
			http.NotFound(w, r)
			return
		}
		tokenHits++
		sawMethod = r.Method
		sawBasicAuth = r.Header.Get("Authorization")
		sawContentType = r.Header.Get("Content-Type")
		_ = r.ParseForm()
		sawGrantType = r.Form.Get("grant_type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"A123","token_type":"Bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	h := New().(*Hooks)
	h.Client = srv.Client()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "csecret"},
	}

	authr, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	req1, _ := http.NewRequest(http.MethodGet, "https://example.invalid/v1/reporting/balances", nil)
	if err := authr.Apply(context.Background(), req1); err != nil {
		t.Fatalf("Apply (1st): %v", err)
	}
	if tokenHits != 1 {
		t.Fatalf("token hits after 1st Apply = %d, want 1", tokenHits)
	}
	if sawMethod != http.MethodPost {
		t.Fatalf("token request method = %q, want POST", sawMethod)
	}
	wantBasic := "Basic " + base64.StdEncoding.EncodeToString([]byte("cid:csecret"))
	if sawBasicAuth != wantBasic {
		t.Fatalf("token Authorization = %q, want %q", sawBasicAuth, wantBasic)
	}
	if sawGrantType != "client_credentials" {
		t.Fatalf("token grant_type = %q, want client_credentials", sawGrantType)
	}
	if sawContentType != "application/x-www-form-urlencoded" {
		t.Fatalf("token Content-Type = %q, want application/x-www-form-urlencoded", sawContentType)
	}
	if got := req1.Header.Get("Authorization"); got != "Bearer A123" {
		t.Fatalf("data request Authorization = %q, want Bearer A123", got)
	}

	// A second Apply within the cached TTL must not re-hit the token endpoint.
	req2, _ := http.NewRequest(http.MethodGet, "https://example.invalid/v1/reporting/balances", nil)
	if err := authr.Apply(context.Background(), req2); err != nil {
		t.Fatalf("Apply (2nd): %v", err)
	}
	if tokenHits != 1 {
		t.Fatalf("token hits after 2nd Apply = %d, want 1 (cached)", tokenHits)
	}
	if got := req2.Header.Get("Authorization"); got != "Bearer A123" {
		t.Fatalf("2nd data request Authorization = %q, want Bearer A123 (cached)", got)
	}
}

// TestAuthenticator_RefetchesAfterExpiry confirms the cache expires and a
// second token exchange happens once the cached token is within 60s of its
// declared expiry.
func TestAuthenticator_RefetchesAfterExpiry(t *testing.T) {
	tokenHits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenHits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok-` + strconv.Itoa(tokenHits) + `","token_type":"Bearer","expires_in":100}`))
	}))
	defer srv.Close()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	h := New().(*Hooks)
	h.Client = srv.Client()
	h.Now = func() time.Time { return now }
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "csecret"},
	}
	authr, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	req1, _ := http.NewRequest(http.MethodGet, "https://example.invalid/x", nil)
	if err := authr.Apply(context.Background(), req1); err != nil {
		t.Fatalf("Apply (1st): %v", err)
	}
	if tokenHits != 1 {
		t.Fatalf("token hits = %d, want 1", tokenHits)
	}

	// Advance past expiry (100s TTL, 60s early-refresh margin).
	now = now.Add(45 * time.Second)
	req2, _ := http.NewRequest(http.MethodGet, "https://example.invalid/x", nil)
	if err := authr.Apply(context.Background(), req2); err != nil {
		t.Fatalf("Apply (2nd): %v", err)
	}
	if tokenHits != 2 {
		t.Fatalf("token hits after expiry = %d, want 2 (re-fetched)", tokenHits)
	}
	if got := req2.Header.Get("Authorization"); got != "Bearer tok-2" {
		t.Fatalf("2nd data request Authorization = %q, want Bearer tok-2", got)
	}
}

// TestAuthenticator_RequiresBaseURLAndCredentials covers Authenticator's
// required-field validation.
func TestAuthenticator_RequiresBaseURLAndCredentials(t *testing.T) {
	h := New().(*Hooks)
	cases := []struct {
		name string
		cfg  connectors.RuntimeConfig
	}{
		{"missing base_url", connectors.RuntimeConfig{Secrets: map[string]string{"client_id": "a", "client_secret": "b"}}},
		{"missing client_id", connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://x"}, Secrets: map[string]string{"client_secret": "b"}}},
		{"missing client_secret", connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://x"}, Secrets: map[string]string{"client_id": "a"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := h.Authenticator(context.Background(), tc.cfg, engine.AuthSpec{}); err == nil {
				t.Fatalf("Authenticator(%s) = nil error, want an error", tc.name)
			}
		})
	}
}

// TestAuthenticator_TokenEndpointFailure confirms a non-2xx token response
// surfaces as an error rather than a silent empty-token Bearer header.
func TestAuthenticator_TokenEndpointFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	h := New().(*Hooks)
	h.Client = srv.Client()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "bad"},
	}
	authr, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "https://example.invalid/x", nil)
	if err := authr.Apply(context.Background(), req); err == nil {
		t.Fatal("Apply with a failing token endpoint should return an error")
	}
}

// --- StreamHook: disputes HATEOAS links[] pagination -------------------

func newTestRuntime(srv *httptest.Server) *engine.Runtime {
	return &engine.Runtime{
		Requester: &connsdk.Requester{BaseURL: srv.URL},
		Bundle:    &engine.Bundle{Name: "paypal-transaction"},
	}
}

// TestReadStream_DisputesFollowsHATEOASLinksArray is the core ENGINE_GAP
// regression: the disputes stream's next-page signal is an ARRAY of
// {rel,href} objects (links:[...]), not a bare string path — this hook must
// find the rel=="next" entry and follow its absolute href across 2 pages,
// stopping when no such entry is present (an empty links array).
func TestReadStream_DisputesFollowsHATEOASLinksArray(t *testing.T) {
	var sawPageSizes []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/customer/disputes" {
			http.NotFound(w, r)
			return
		}
		sawPageSizes = append(sawPageSizes, r.URL.Query().Get("page_size"))
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.RawQuery, "next_page_token") {
			_, _ = w.Write([]byte(`{"items":[{"dispute_id":"PP-D-2","reason":"UNAUTHORIZED","status":"RESOLVED","dispute_state":"RESOLVED","dispute_amount":{"currency_code":"USD","value":"5.00"},"create_time":"2026-01-02T00:00:00Z","update_time":"2026-01-02T00:05:00Z"}],"links":[]}`))
			return
		}
		nextHref := "http://" + r.Host + "/v1/customer/disputes?page_size=50&next_page_token=abc"
		_, _ = w.Write([]byte(`{"items":[{"dispute_id":"PP-D-1","reason":"MERCHANDISE_OR_SERVICE_NOT_RECEIVED","status":"RESOLVED","dispute_state":"RESOLVED","dispute_amount":{"currency_code":"USD","value":"12.31"},"create_time":"2026-01-01T00:00:00Z","update_time":"2026-01-01T00:05:00Z"}],"links":[{"href":"` + nextHref + `","rel":"next","method":"GET"}]}`))
	}))
	defer srv.Close()

	h := New().(*Hooks)
	rt := newTestRuntime(srv)
	req := connectors.ReadRequest{Stream: "disputes", Config: connectors.RuntimeConfig{Config: map[string]string{}}}
	stream := engine.StreamSpec{Name: "disputes"}

	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), stream, req, rt, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream(disputes) handled = false, want true")
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (across 2 HATEOAS pages)", len(got))
	}
	if got[0]["dispute_id"] != "PP-D-1" || got[1]["dispute_id"] != "PP-D-2" {
		t.Fatalf("record mapping/order wrong: %+v", got)
	}
	if got[0]["dispute_amount_value"] != "12.31" {
		t.Fatalf("nested dispute_amount mapping wrong: %+v", got[0])
	}
	if len(sawPageSizes) != 2 || sawPageSizes[0] != "50" {
		t.Fatalf("page_size query params = %v, want [\"50\", ...]", sawPageSizes)
	}
}

// TestReadStream_UnrecognizedStreamFallsBackToDeclarative confirms the hook
// only claims "disputes" and returns handled=false for every other stream
// name, letting the engine's declarative path (or its own "not found" error)
// take over.
func TestReadStream_UnrecognizedStreamFallsBackToDeclarative(t *testing.T) {
	h := New().(*Hooks)
	for _, name := range []string{"transactions", "balances", "products", ""} {
		stream := engine.StreamSpec{Name: name}
		req := connectors.ReadRequest{Stream: name, Config: connectors.RuntimeConfig{}}
		handled, err := h.ReadStream(context.Background(), stream, req, &engine.Runtime{}, func(connectors.Record) error { return nil })
		if err != nil {
			t.Fatalf("ReadStream(%q): %v", name, err)
		}
		if handled {
			t.Fatalf("ReadStream(%q) handled = true, want false (declarative fallback)", name)
		}
	}
}

// TestReadStream_MaxPagesCapsDisputesPagination confirms max_pages bounds
// the HATEOAS follow loop even when more pages remain.
func TestReadStream_MaxPagesCapsDisputesPagination(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		nextHref := "http://" + r.Host + "/v1/customer/disputes?page_size=50&next_page_token=" + strconv.Itoa(hits)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"dispute_id":"PP-D-` + strconv.Itoa(hits) + `"}],"links":[{"href":"` + nextHref + `","rel":"next"}]}`))
	}))
	defer srv.Close()

	h := New().(*Hooks)
	rt := newTestRuntime(srv)
	req := connectors.ReadRequest{Stream: "disputes", Config: connectors.RuntimeConfig{Config: map[string]string{"max_pages": "2"}}}
	var got []connectors.Record
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "disputes"}, req, rt, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if hits != 2 {
		t.Fatalf("requests made = %d, want 2 (max_pages cap)", hits)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
}
