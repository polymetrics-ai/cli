// Package netsuite tests the netsuite AuthHook (OAuth 1.0a HMAC-SHA256
// request signing, ported from legacy netsuite.go's oauth1).
package netsuite

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func baseCfg() connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{"realm": "123456"},
		Secrets: map[string]string{
			"consumer_key":    "ck",
			"consumer_secret": "cs",
			"token_key":       "tk",
			"token_secret":    "ts",
		},
	}
}

func baseSpec() engine.AuthSpec { return engine.AuthSpec{Mode: "custom", Hook: "netsuite"} }

func doRequest(t *testing.T, auth interface {
	Apply(ctx context.Context, req *http.Request) error
}, method, target string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, target, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return req
}

// --- registration ---------------------------------------------------------

func TestHooksRegisteredUnderNetsuite(t *testing.T) {
	h := engine.HooksFor("netsuite")
	if h == nil {
		t.Fatal("engine.HooksFor(\"netsuite\") = nil, want registered hooks (hooks/netsuite's init() must call engine.RegisterHooks)")
	}
	if h.ConnectorName() != "netsuite" {
		t.Fatalf("ConnectorName() = %q, want %q", h.ConnectorName(), "netsuite")
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("registered netsuite hooks does not implement engine.AuthHook")
	}
}

// --- header shape -----------------------------------------------------------

func TestAuthenticator_SetsOAuthHeader(t *testing.T) {
	h := New().(*Hooks)
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doRequest(t, auth, http.MethodGet, "http://example.invalid/customer?limit=1")

	got := req.Header.Get("Authorization")
	if !strings.HasPrefix(got, "OAuth ") {
		t.Fatalf("Authorization = %q, want prefix %q", got, "OAuth ")
	}
	if !strings.Contains(got, `realm="123456"`) {
		t.Fatalf("Authorization = %q, want realm=%q", got, "123456")
	}
	for _, want := range []string{
		`oauth_consumer_key="ck"`,
		`oauth_token="tk"`,
		`oauth_signature_method="HMAC-SHA256"`,
		`oauth_version="1.0"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("Authorization = %q, want it to contain %q", got, want)
		}
	}
	if !strings.Contains(got, "oauth_signature=") {
		t.Fatal("Authorization missing oauth_signature")
	}
	if !strings.Contains(got, "oauth_timestamp=") {
		t.Fatal("Authorization missing oauth_timestamp")
	}
	if !strings.Contains(got, "oauth_nonce=") {
		t.Fatal("Authorization missing oauth_nonce")
	}
}

// TestAuthenticator_SignatureMatchesCanonicalScheme pins the exact
// signature scheme ported from legacy's oauthSignature, using an injected
// clock for a deterministic timestamp/nonce.
func TestAuthenticator_SignatureMatchesCanonicalScheme(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	h := &Hooks{Now: func() time.Time { return fixed }}
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doRequest(t, auth, http.MethodGet, "http://example.invalid/customer?limit=1")

	ts := strconv.FormatInt(fixed.Unix(), 10)
	nonce := strconv.FormatInt(fixed.UnixNano(), 36)

	params := map[string]string{
		"oauth_consumer_key":     "ck",
		"oauth_token":            "tk",
		"oauth_signature_method": "HMAC-SHA256",
		"oauth_timestamp":        ts,
		"oauth_nonce":            nonce,
		"oauth_version":          "1.0",
		"limit":                  "1",
	}
	u, _ := url.Parse("http://example.invalid/customer?limit=1")
	wantSig := oauthSignature(http.MethodGet, u, params, "cs", "ts")

	got := req.Header.Get("Authorization")
	wantFrag := fmt.Sprintf("oauth_signature=%q", percentEncode(wantSig))
	if !strings.Contains(got, wantFrag) {
		t.Fatalf("Authorization = %q, want it to contain %q", got, wantFrag)
	}
}

// TestAuthenticator_SignatureChangesWithQuery proves the query parameters
// are part of the signed canonical string (a signature valid for one query
// must not validate for another).
func TestAuthenticator_SignatureChangesWithQuery(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	h := &Hooks{Now: func() time.Time { return fixed }}
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	reqA := doRequest(t, auth, http.MethodGet, "http://example.invalid/customer?limit=1&offset=0")
	reqB := doRequest(t, auth, http.MethodGet, "http://example.invalid/customer?limit=1&offset=100")

	if reqA.Header.Get("Authorization") == reqB.Header.Get("Authorization") {
		t.Fatal("signature identical across different query strings, want distinct signatures")
	}
}

// TestAuthenticator_SignsFreshPerRequest: legacy computes a new timestamp
// and nonce on every call, never caching — two Apply calls at different
// times must produce different signatures.
func TestAuthenticator_SignsFreshPerRequest(t *testing.T) {
	current := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	h := &Hooks{Now: func() time.Time { return current }}
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req1 := doRequest(t, auth, http.MethodGet, "http://example.invalid/customer")
	current = current.Add(1 * time.Second)
	req2 := doRequest(t, auth, http.MethodGet, "http://example.invalid/customer")

	if req1.Header.Get("Authorization") == req2.Header.Get("Authorization") {
		t.Fatal("Authorization identical across requests at different timestamps, want a fresh signature per request")
	}
}

// --- credential resolution: config vs secrets fallback --------------------

// TestAuthenticator_CredentialsFromConfigFallback mirrors legacy's
// configOrSecret: every one of the five values may come from either
// cfg.Config or cfg.Secrets.
func TestAuthenticator_CredentialsFromConfigFallback(t *testing.T) {
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"realm":           "999",
			"consumer_key":    "ck-cfg",
			"consumer_secret": "cs-cfg",
			"token_key":       "tk-cfg",
			"token_secret":    "ts-cfg",
		},
	}
	h := New().(*Hooks)
	auth, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doRequest(t, auth, http.MethodGet, "http://example.invalid/customer")
	got := req.Header.Get("Authorization")
	if !strings.Contains(got, `realm="999"`) || !strings.Contains(got, `oauth_consumer_key="ck-cfg"`) {
		t.Fatalf("Authorization = %q, want values sourced from config", got)
	}
}

// --- error paths ------------------------------------------------------------

func TestAuthenticator_MissingRealmIsError(t *testing.T) {
	cfg := baseCfg()
	delete(cfg.Config, "realm")
	h := New().(*Hooks)
	if _, err := h.Authenticator(context.Background(), cfg, baseSpec()); err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing realm")
	} else if !strings.Contains(err.Error(), "realm") {
		t.Fatalf("error = %q, want it to mention realm", err.Error())
	}
}

func TestAuthenticator_MissingConsumerKeyIsError(t *testing.T) {
	cfg := baseCfg()
	delete(cfg.Secrets, "consumer_key")
	h := New().(*Hooks)
	if _, err := h.Authenticator(context.Background(), cfg, baseSpec()); err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing consumer_key")
	} else if !strings.Contains(err.Error(), "consumer_key") {
		t.Fatalf("error = %q, want it to mention consumer_key", err.Error())
	}
}

func TestAuthenticator_MissingConsumerSecretIsError(t *testing.T) {
	cfg := baseCfg()
	delete(cfg.Secrets, "consumer_secret")
	h := New().(*Hooks)
	if _, err := h.Authenticator(context.Background(), cfg, baseSpec()); err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing consumer_secret")
	} else if !strings.Contains(err.Error(), "consumer_secret") {
		t.Fatalf("error = %q, want it to mention consumer_secret", err.Error())
	}
}

func TestAuthenticator_MissingTokenKeyIsError(t *testing.T) {
	cfg := baseCfg()
	delete(cfg.Secrets, "token_key")
	h := New().(*Hooks)
	if _, err := h.Authenticator(context.Background(), cfg, baseSpec()); err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing token_key")
	} else if !strings.Contains(err.Error(), "token_key") {
		t.Fatalf("error = %q, want it to mention token_key", err.Error())
	}
}

func TestAuthenticator_MissingTokenSecretIsError(t *testing.T) {
	cfg := baseCfg()
	delete(cfg.Secrets, "token_secret")
	h := New().(*Hooks)
	if _, err := h.Authenticator(context.Background(), cfg, baseSpec()); err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing token_secret")
	} else if !strings.Contains(err.Error(), "token_secret") {
		t.Fatalf("error = %q, want it to mention token_secret", err.Error())
	}
}

// --- secret redaction --------------------------------------------------------

func TestAuthenticator_ErrorsNeverContainSecretText(t *testing.T) {
	cfg := baseCfg()
	cfg.Secrets["consumer_key"] = "consumer-key-marker-xyz"
	cfg.Secrets["token_key"] = "token-key-marker-xyz"
	cfg.Secrets["token_secret"] = "token-secret-marker-xyz"
	delete(cfg.Secrets, "consumer_secret")
	h := New().(*Hooks)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("expected an error for missing consumer_secret")
	}
	for _, marker := range []string{"consumer-key-marker-xyz", "token-key-marker-xyz", "token-secret-marker-xyz"} {
		if strings.Contains(err.Error(), marker) {
			t.Fatalf("error text contains secret marker %q: %s", marker, err.Error())
		}
	}
}

// --- BaseURLFromRealm helper -------------------------------------------------

func TestBaseURLFromRealm(t *testing.T) {
	got := BaseURLFromRealm("123456_SB1")
	want := "https://123456-sb1.suitetalk.api.netsuite.com/services/rest/record/v1"
	if got != want {
		t.Fatalf("BaseURLFromRealm(%q) = %q, want %q", "123456_SB1", got, want)
	}
}
