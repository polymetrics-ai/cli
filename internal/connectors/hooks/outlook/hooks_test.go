// Package outlook_test exercises the outlook hook set's AuthHook (OAuth2
// refresh-token grant, gmail pattern per this task's mandate) and StreamHook
// (@odata.nextLink pagination). Written test-first per TDD convention: every
// test below fails to compile before hooks.go exists.
package outlook

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

// tokenServer stands in for Microsoft's identity platform token endpoint.
// It MUST be TLS: the hook requires token_url to be https (THREAT-MODEL.md
// Delta 2). respond is invoked per request with the decoded form body.
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
		Config: map[string]string{"token_url": tokenURL, "scope": "Mail.Read"},
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
		Hook:         "outlook",
		TokenURL:     "{{ config.token_url }}",
		ClientID:     "{{ secrets.client_id }}",
		ClientSecret: "{{ secrets.client_secret }}",
		Token:        "{{ secrets.refresh_token }}",
		Scopes:       "{{ config.scope }}",
	}
}

func newTestHooks(now func() time.Time, client *http.Client) *Hooks {
	h := New().(*Hooks)
	h.Now = now
	h.Client = client
	return h
}

func newClientHooks(client *http.Client) *Hooks {
	h := New().(*Hooks)
	h.Client = client
	return h
}

func doAuthenticatedRequest(t *testing.T, auth connsdk.Authenticator) *http.Request {
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

// --- registration -------------------------------------------------------

func TestHooksRegisteredUnderOutlook(t *testing.T) {
	h := engine.HooksFor("outlook")
	if h == nil {
		t.Fatal("engine.HooksFor(\"outlook\") = nil, want registered hooks (hooks/outlook's init() must call engine.RegisterHooks)")
	}
	if h.ConnectorName() != "outlook" {
		t.Fatalf("ConnectorName() = %q, want %q", h.ConnectorName(), "outlook")
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("registered outlook hooks does not implement engine.AuthHook")
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("registered outlook hooks does not implement engine.StreamHook")
	}
}

// --- refresh-grant form shape -------------------------------------------

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
	if got := gotForm.Get("scope"); got != "Mail.Read" {
		t.Fatalf("scope = %q, want the configured scope", got)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer tok_abc" {
		t.Fatalf("Authorization header = %q, want %q", got, "Bearer tok_abc")
	}
}

func TestAuthenticator_ScopeOmittedWhenUnset(t *testing.T) {
	var gotForm url.Values
	srv, client, _ := tokenServer(t, func(form url.Values) (int, map[string]any) {
		gotForm = form
		return http.StatusOK, map[string]any{"access_token": "tok_abc", "expires_in": 3600}
	})

	cfg := baseCfg(srv.URL)
	delete(cfg.Config, "scope")

	h := newClientHooks(client)
	auth, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	_ = doAuthenticatedRequest(t, auth)

	if _, ok := gotForm["scope"]; ok {
		t.Fatalf("form has scope key = %v, want omitted entirely when unset", gotForm["scope"])
	}
}

// --- caching / expiry ----------------------------------------------------

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

	req1 := doAuthenticatedRequest(t, auth)
	req2 := doAuthenticatedRequest(t, auth)

	if *hits != 1 {
		t.Fatalf("token endpoint hits = %d, want 1 (second Apply should reuse the cached token)", *hits)
	}
	if req1.Header.Get("Authorization") != req2.Header.Get("Authorization") {
		t.Fatalf("Authorization headers differ across cached requests: %q vs %q", req1.Header.Get("Authorization"), req2.Header.Get("Authorization"))
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

	_ = doAuthenticatedRequest(t, auth) // primes the cache, expires at now+3600s

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

// --- error paths ----------------------------------------------------------

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
	err = auth.Apply(context.Background(), req)
	if err == nil {
		t.Fatal("Apply() error = nil, want an error for a non-2xx token endpoint response")
	}
	if req.Header.Get("Authorization") != "" {
		t.Fatalf("Authorization header set = %q after a failed token exchange, want empty (no silent unauthenticated fallback)", req.Header.Get("Authorization"))
	}
}

func TestAuthenticator_MissingRefreshTokenIsError(t *testing.T) {
	cfg := baseCfg("https://login.microsoftonline.com/common/oauth2/v2.0/token")
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

func TestAuthenticator_MissingClientIDIsError(t *testing.T) {
	cfg := baseCfg("https://login.microsoftonline.com/common/oauth2/v2.0/token")
	delete(cfg.Secrets, "client_id")

	h := New().(*Hooks)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing client_id")
	}
	if !strings.Contains(err.Error(), "client_id") {
		t.Fatalf("error = %q, want it to name client_id", err.Error())
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

func TestAuthenticator_TokenURLUnparseableIsError(t *testing.T) {
	cfg := baseCfg("://not-a-url")

	h := New().(*Hooks)
	if _, err := h.Authenticator(context.Background(), cfg, baseSpec()); err == nil {
		t.Fatal("Authenticator() error = nil, want an error for an unparseable token_url")
	}
}

// --- ctx cancellation -----------------------------------------------------

func TestAuthenticator_HonorsContextCancellation(t *testing.T) {
	srv, client, _ := tokenServer(t, func(form url.Values) (int, map[string]any) {
		return http.StatusOK, map[string]any{"access_token": "tok_abc", "expires_in": 3600}
	})

	h := newClientHooks(client)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err := auth.Apply(ctx, req); err == nil {
		t.Fatal("Apply(cancelled ctx) error = nil, want a cancellation error")
	}
}

// --- secret redaction ------------------------------------------------------

func TestAuthenticator_ErrorsNeverContainSecretText(t *testing.T) {
	const (
		secretMarkerClientSecret = "client-secret-fixture"
		secretMarkerRefreshToken = "refresh-token-fixture"
	)

	srv, client, _ := tokenServer(t, func(form url.Values) (int, map[string]any) {
		return http.StatusUnauthorized, map[string]any{"error": "invalid_grant"}
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
	for _, marker := range []string{secretMarkerClientSecret, secretMarkerRefreshToken} {
		if strings.Contains(msg, marker) {
			t.Fatalf("error text contains secret marker %q: %s", marker, msg)
		}
	}
}

// --- StreamHook: @odata.nextLink pagination -------------------------------

func graphServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	var srv *httptest.Server

	mux.HandleFunc("/me/messages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("$skiptoken") {
		case "":
			if got := r.URL.Query().Get("$top"); got != "100" {
				t.Fatalf("first page $top = %q, want 100", got)
			}
			_, _ = w.Write([]byte(`{"value":[{"id":"msg_1","subject":"Hello","receivedDateTime":"2026-01-01T00:00:00Z","lastModifiedDateTime":"2026-01-01T00:00:00Z","webLink":"https://example.com/1"}],"@odata.nextLink":"` + srv.URL + `/me/messages?$skiptoken=page2"}`))
		case "page2":
			if got := r.URL.Query().Get("$top"); got != "" {
				t.Fatalf("second page must carry no extra query beyond nextLink's own, got $top=%q", got)
			}
			_, _ = w.Write([]byte(`{"value":[{"id":"msg_2","subject":"World","receivedDateTime":"2026-01-02T00:00:00Z","lastModifiedDateTime":"2026-01-02T00:00:00Z","webLink":"https://example.com/2"}]}`))
		default:
			t.Fatalf("unexpected $skiptoken %q", r.URL.Query().Get("$skiptoken"))
		}
	})

	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newRuntime(t *testing.T, baseURL string) *engine.Runtime {
	t.Helper()
	return &engine.Runtime{
		Requester: &connsdk.Requester{BaseURL: baseURL},
		Bundle: &engine.Bundle{
			Name:    "outlook",
			Schemas: map[string]*engine.StreamSchema{"messages": schemaWithProperties(t, []string{"id", "subject", "received_date_time", "last_modified_date_time", "web_link"})},
		},
	}
}

// schemaWithProperties builds a minimal *engine.StreamSchema projecting
// exactly props, mirroring hooks/microsoft-teams/hooks_test.go's identical
// helper -- enough for ReadStream to run without a full loaded bundle.
func schemaWithProperties(t *testing.T, props []string) *engine.StreamSchema {
	t.Helper()
	doc := map[string]any{"$schema": "http://json-schema.org/draft-07/schema#", "type": "object", "properties": map[string]any{}}
	propsMap := doc["properties"].(map[string]any)
	for _, p := range props {
		propsMap[p] = map[string]any{"type": []string{"string", "null"}}
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}
	sch, err := engine.CompileSchema(raw)
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}
	return &engine.StreamSchema{Schema: sch}
}

func TestReadStream_FollowsODataNextLink(t *testing.T) {
	srv := graphServer(t)
	rt := newRuntime(t, srv.URL)

	h := New().(*Hooks)
	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "messages"}, connectors.ReadRequest{Config: connectors.RuntimeConfig{}}, rt, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream handled = false, want true for a recognized stream")
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (both pages followed via @odata.nextLink)", len(got))
	}
	if got[0]["subject"] != "Hello" || got[1]["received_date_time"] != "2026-01-02T00:00:00Z" {
		t.Fatalf("mapped records wrong: %+v", got)
	}
}

func TestReadStream_UnrecognizedStreamFallsBack(t *testing.T) {
	srv := graphServer(t)
	rt := newRuntime(t, srv.URL)

	h := New().(*Hooks)
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "not_a_stream"}, connectors.ReadRequest{Config: connectors.RuntimeConfig{}}, rt, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("ReadStream handled = true for an unrecognized stream, want false (declarative fallback)")
	}
}
