// Package quickbooks: hooks_test.go covers both hook interfaces this bundle
// registers: AuthHook (OAuth2 refresh-token grant, ported from
// hooks/gmail/hooks_test.go's shape) and StreamHook (the embedded
// STARTPOSITION/MAXRESULTS harvest loop, parity-tested against legacy
// quickbooks.go's TestReadCustomersAuthenticatesPaginatesAndMapsRecords).
package quickbooks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

// --- test helpers -----------------------------------------------------

func tokenServer(t *testing.T, respond func(form url.Values) (int, map[string]any)) (*httptest.Server, *http.Client, *int32) {
	t.Helper()
	var hits int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if err := r.ParseForm(); err != nil {
			t.Fatalf("token server: parse form: %v", err)
		}
		status, body := respond(r.PostForm)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body == nil {
			body = map[string]any{"error": "server_error"}
		}
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	return srv, srv.Client(), &hits
}

func baseCfg(tokenURL string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{"token_url": tokenURL, "realm_id": "123"},
		Secrets: map[string]string{
			"client_id":     "client-id-fixture",
			"client_secret": "client-secret-fixture",
			"refresh_token": "refresh-token-fixture",
		},
	}
}

func baseSpec() engine.AuthSpec {
	return engine.AuthSpec{
		Mode:         "custom",
		Hook:         "quickbooks",
		TokenURL:     "{{ config.token_url }}",
		ClientID:     "{{ secrets.client_id }}",
		ClientSecret: "{{ secrets.client_secret }}",
		Token:        "{{ secrets.refresh_token }}",
	}
}

func newClientHooks(client *http.Client) *Hooks {
	h := New().(*Hooks)
	h.Client = client
	return h
}

func newTestHooks(now func() time.Time, client *http.Client) *Hooks {
	h := New().(*Hooks)
	h.Now = now
	h.Client = client
	return h
}

func doAuthenticatedRequest(t *testing.T, auth interface {
	Apply(ctx context.Context, req *http.Request) error
}) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return req
}

// --- registration ---------------------------------------------------------

func TestHooksRegisteredUnderQuickbooks(t *testing.T) {
	h := engine.HooksFor("quickbooks")
	if h == nil {
		t.Fatal("engine.HooksFor(\"quickbooks\") = nil, want registered hooks")
	}
	if h.ConnectorName() != "quickbooks" {
		t.Fatalf("ConnectorName() = %q, want %q", h.ConnectorName(), "quickbooks")
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("registered quickbooks hooks does not implement engine.AuthHook")
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered quickbooks hooks does not implement engine.StreamHook")
	}
}

// --- AuthHook: refresh-grant form shape -------------------------------------

func TestAuthenticator_RefreshGrantFormShape(t *testing.T) {
	var gotForm url.Values
	srv, client, hits := tokenServer(t, func(form url.Values) (int, map[string]any) {
		gotForm = form
		return http.StatusOK, map[string]any{"access_token": "tok_abc", "token_type": "Bearer", "expires_in": 3600}
	})

	h := newClientHooks(client)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)

	if *hits != 1 {
		t.Fatalf("token endpoint hits = %d, want 1", *hits)
	}
	if got := gotForm.Get("grant_type"); got != "refresh_token" {
		t.Fatalf("grant_type = %q, want %q", got, "refresh_token")
	}
	if got := gotForm.Get("refresh_token"); got != "refresh-token-fixture" {
		t.Fatalf("refresh_token = %q, want %q", got, "refresh-token-fixture")
	}
	if got := gotForm.Get("client_id"); got != "client-id-fixture" {
		t.Fatalf("client_id = %q, want %q", got, "client-id-fixture")
	}
	if got := gotForm.Get("client_secret"); got != "client-secret-fixture" {
		t.Fatalf("client_secret = %q, want %q", got, "client-secret-fixture")
	}
	if got := req.Header.Get("Authorization"); got != "Bearer tok_abc" {
		t.Fatalf("Authorization header = %q, want %q", got, "Bearer tok_abc")
	}
}

func TestAuthenticator_CachesTokenAcrossRequests(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	srv, client, hits := tokenServer(t, func(form url.Values) (int, map[string]any) {
		return http.StatusOK, map[string]any{"access_token": "tok_1", "expires_in": 3600}
	})

	h := newTestHooks(func() time.Time { return now }, client)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	_ = doAuthenticatedRequest(t, auth)
	_ = doAuthenticatedRequest(t, auth)

	if *hits != 1 {
		t.Fatalf("token endpoint hits = %d, want 1 (second Apply should reuse the cached token)", *hits)
	}
}

func TestAuthenticator_RefreshesWithin60sOfExpiry(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	call := 0
	srv, client, hits := tokenServer(t, func(form url.Values) (int, map[string]any) {
		call++
		return http.StatusOK, map[string]any{"access_token": "tok_" + itoa(call), "expires_in": 3600}
	})

	current := now
	h := newTestHooks(func() time.Time { return current }, client)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	_ = doAuthenticatedRequest(t, auth)

	current = now.Add(3539 * time.Second)
	_ = doAuthenticatedRequest(t, auth)
	if *hits != 1 {
		t.Fatalf("token endpoint hits = %d after t+3539s, want 1 (still within cache window)", *hits)
	}

	current = now.Add(3541 * time.Second)
	_ = doAuthenticatedRequest(t, auth)
	if *hits != 2 {
		t.Fatalf("token endpoint hits = %d after t+3541s, want 2 (60s-early refresh must trigger)", *hits)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// --- AuthHook: error paths --------------------------------------------------

func TestAuthenticator_NonSuccessTokenResponseIsError(t *testing.T) {
	srv, client, _ := tokenServer(t, func(form url.Values) (int, map[string]any) {
		return http.StatusUnauthorized, map[string]any{"error": "invalid_grant"}
	})

	h := newClientHooks(client)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err := auth.Apply(context.Background(), req); err == nil {
		t.Fatal("Apply() error = nil, want an error for a non-2xx token endpoint response")
	}
	if req.Header.Get("Authorization") != "" {
		t.Fatalf("Authorization header set after a failed token exchange, want empty")
	}
}

func TestAuthenticator_MissingRefreshTokenIsError(t *testing.T) {
	cfg := baseCfg("https://oauth2.bearer.token.intuit.com/oauth2/v1/tokens/bearer")
	delete(cfg.Secrets, "refresh_token")

	h := New().(*Hooks)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing refresh token")
	}
	if !strings.Contains(err.Error(), "refresh") {
		t.Fatalf("error = %q, want it to name the missing refresh_token field", err.Error())
	}
}

func TestAuthenticator_MissingClientSecretIsError(t *testing.T) {
	cfg := baseCfg("https://oauth2.bearer.token.intuit.com/oauth2/v1/tokens/bearer")
	delete(cfg.Secrets, "client_secret")

	h := New().(*Hooks)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing client_secret (QuickBooks requires it, unlike gmail)")
	}
	if !strings.Contains(err.Error(), "client_secret") {
		t.Fatalf("error = %q, want it to name client_secret", err.Error())
	}
}

func TestAuthenticator_TokenURLMustBeHTTPS(t *testing.T) {
	cfg := baseCfg("http://insecure.example.invalid/token")

	h := newClientHooks(nil)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("Authenticator() error = nil, want a fail-closed error for a non-https token_url")
	}
	if !strings.Contains(err.Error(), "https") {
		t.Fatalf("error = %q, want it to mention the https requirement", err.Error())
	}
}

func TestAuthenticator_ErrorsNeverContainSecretText(t *testing.T) {
	srv, client, _ := tokenServer(t, func(form url.Values) (int, map[string]any) {
		return http.StatusUnauthorized, map[string]any{"error": "invalid_grant", "access_token": "tok_super_secret_access_value"}
	})

	h := newClientHooks(client)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	err = auth.Apply(context.Background(), req)
	if err == nil {
		t.Fatal("expected an error from the 401 token response")
	}
	msg := err.Error()
	for _, marker := range []string{"client-secret-fixture", "refresh-token-fixture"} {
		if strings.Contains(msg, marker) {
			t.Fatalf("error text contains secret marker %q: %s", marker, msg)
		}
	}
}

// --- StreamHook: harvest loop (parity with legacy quickbooks.go) -----------

func newRuntime(baseURL string) *engine.Runtime {
	return &engine.Runtime{Requester: &connsdk.Requester{BaseURL: baseURL, Auth: connsdk.Bearer("qb_token")}}
}

func TestReadStream_CustomersAuthenticatesPaginatesAndMapsRecords(t *testing.T) {
	var sawAuth string
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v3/company/123/query" {
			http.NotFound(w, r)
			return
		}
		query := r.URL.Query().Get("query")
		queries = append(queries, query)
		switch {
		case strings.Contains(query, "STARTPOSITION 1"):
			_, _ = w.Write([]byte(`{"QueryResponse":{"Customer":[{"Id":"1","DisplayName":"Ada Lovelace","Active":true},{"Id":"2","DisplayName":"Grace Hopper","Active":true}]}}`))
		case strings.Contains(query, "STARTPOSITION 3"):
			_, _ = w.Write([]byte(`{"QueryResponse":{"Customer":[{"Id":"3","DisplayName":"Katherine Johnson","Active":true}]}}`))
		default:
			t.Fatalf("unexpected query %q", query)
		}
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{
		Stream: "customers",
		Config: connectors.RuntimeConfig{Config: map[string]string{"realm_id": "123", "page_size": "2"}},
	}
	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "customers"}, req, newRuntime(srv.URL), func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true for a known stream")
	}
	if sawAuth != "Bearer qb_token" {
		t.Fatalf("Authorization = %q, want Bearer qb_token (rt.Requester's auth must be used)", sawAuth)
	}
	if len(queries) != 2 {
		t.Fatalf("queries = %v, want two pages", queries)
	}
	if len(got) != 3 || got[0]["id"] != "1" || got[0]["display_name"] != "Ada Lovelace" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

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

func TestReadStream_EmptyStreamNameDefaultsToCustomers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"QueryResponse":{"Customer":[]}}`))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: map[string]string{"realm_id": "123"}}}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: ""}, req, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true (empty stream name should default to customers)")
	}
}

func TestReadStream_MissingRealmIDIsError(t *testing.T) {
	h := Hooks{}
	req := connectors.ReadRequest{Stream: "customers", Config: connectors.RuntimeConfig{Config: map[string]string{}}}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "customers"}, req, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if !handled {
		t.Fatal("handled = false, want true (a known stream with bad config is still a handled error, not a declarative fallback)")
	}
	if err == nil {
		t.Fatal("ReadStream error = nil, want an error for a missing realm_id")
	}
}

func TestReadStream_PathUnsafeRealmIDIsError(t *testing.T) {
	h := Hooks{}
	req := connectors.ReadRequest{Stream: "customers", Config: connectors.RuntimeConfig{Config: map[string]string{"realm_id": "../etc/passwd"}}}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "customers"}, req, newRuntime("http://example.invalid"), func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("ReadStream error = nil, want an error for a path-unsafe realm_id")
	}
}

func TestReadStream_MaxPagesCapsRequestCount(t *testing.T) {
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query().Get("query"))
		_, _ = w.Write([]byte(`{"QueryResponse":{"Customer":[{"Id":"1","DisplayName":"Ada"},{"Id":"2","DisplayName":"Grace"}]}}`))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{
		Stream: "customers",
		Config: connectors.RuntimeConfig{Config: map[string]string{"realm_id": "123", "page_size": "2", "max_pages": "1"}},
	}
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "customers"}, req, newRuntime(srv.URL), func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(queries) != 1 {
		t.Fatalf("queries = %d, want exactly 1 (max_pages=1 cap), even though every page was full", len(queries))
	}
}

func TestReadStream_InvoicesUnwrapsCustomerRefValue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"QueryResponse":{"Invoice":[{"Id":"1","DocNumber":"DOC-1","CustomerRef":{"value":"42","name":"Fixture"},"TotalAmt":100,"Balance":50}]}}`))
	}))
	defer srv.Close()

	h := Hooks{}
	req := connectors.ReadRequest{Stream: "invoices", Config: connectors.RuntimeConfig{Config: map[string]string{"realm_id": "123", "page_size": "10"}}}
	var got []connectors.Record
	_, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "invoices"}, req, newRuntime(srv.URL), func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if len(got) != 1 || got[0]["customer_ref"] != "42" {
		t.Fatalf("unexpected records: %+v", got)
	}
}
