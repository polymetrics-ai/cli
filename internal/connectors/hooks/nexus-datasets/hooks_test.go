// Package nexusdatasets tests the nexus-datasets AuthHook (HMAC-SHA256
// request signing, ported from legacy nexus_datasets.go's hmacAuth).
package nexusdatasets

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func baseCfg() connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{},
		Secrets: map[string]string{
			"access_key_id": "AKID123",
			"user_id":       "user-1",
			"secret_key":    "shhh-secret",
			"api_key":       "data-api-key",
		},
	}
}

func baseSpec() engine.AuthSpec {
	return engine.AuthSpec{Mode: "custom", Hook: "nexus-datasets"}
}

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

// --- registration -------------------------------------------------------

func TestHooksRegisteredUnderNexusDatasets(t *testing.T) {
	h := engine.HooksFor("nexus-datasets")
	if h == nil {
		t.Fatal("engine.HooksFor(\"nexus-datasets\") = nil, want registered hooks (hooks/nexus-datasets's init() must call engine.RegisterHooks)")
	}
	if h.ConnectorName() != "nexus-datasets" {
		t.Fatalf("ConnectorName() = %q, want %q", h.ConnectorName(), "nexus-datasets")
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("registered nexus-datasets hooks does not implement engine.AuthHook")
	}
}

// --- header shape ---------------------------------------------------------

func TestAuthenticator_SetsIdentityHeaders(t *testing.T) {
	h := New().(*Hooks)
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doRequest(t, auth, http.MethodGet, "http://example.invalid/datasets/orders?limit=100&offset=0")

	if got := req.Header.Get("X-Infor-AccessKeyId"); got != "AKID123" {
		t.Fatalf("X-Infor-AccessKeyId = %q, want AKID123", got)
	}
	if got := req.Header.Get("X-Infor-UserId"); got != "user-1" {
		t.Fatalf("X-Infor-UserId = %q, want user-1", got)
	}
	if got := req.Header.Get("X-Infor-ApiKey"); got != "data-api-key" {
		t.Fatalf("X-Infor-ApiKey = %q, want data-api-key", got)
	}
	if got := req.Header.Get("X-Infor-Timestamp"); strings.TrimSpace(got) == "" {
		t.Fatal("X-Infor-Timestamp not set")
	}
	if got := req.Header.Get("Authorization"); !strings.HasPrefix(got, "InforNexus AKID123:") {
		t.Fatalf("Authorization = %q, want prefix %q", got, "InforNexus AKID123:")
	}
}

// TestAuthenticator_SignatureMatchesCanonicalScheme pins the exact
// HMAC-SHA256(secret_key, method + "\n" + path + "\n" + timestamp) scheme
// ported from legacy's hmacAuth.Apply, using an injected clock for a
// deterministic timestamp.
func TestAuthenticator_SignatureMatchesCanonicalScheme(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	h := &Hooks{Now: func() time.Time { return fixed }}
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doRequest(t, auth, http.MethodGet, "http://example.invalid/datasets/orders")

	ts := strconv.FormatInt(fixed.UTC().Unix(), 10)
	canonical := strings.Join([]string{http.MethodGet, "/datasets/orders", ts}, "\n")
	mac := hmac.New(sha256.New, []byte("shhh-secret"))
	mac.Write([]byte(canonical))
	wantSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if got := req.Header.Get("X-Infor-Timestamp"); got != ts {
		t.Fatalf("X-Infor-Timestamp = %q, want %q", got, ts)
	}
	wantAuth := "InforNexus AKID123:" + wantSig
	if got := req.Header.Get("Authorization"); got != wantAuth {
		t.Fatalf("Authorization = %q, want %q", got, wantAuth)
	}
}

// TestAuthenticator_SignatureChangesWithPath proves the path is part of the
// signed canonical string (a signature valid for one path must not validate
// for another).
func TestAuthenticator_SignatureChangesWithPath(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	h := &Hooks{Now: func() time.Time { return fixed }}
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	reqA := doRequest(t, auth, http.MethodGet, "http://example.invalid/datasets/orders")
	reqB := doRequest(t, auth, http.MethodGet, "http://example.invalid/datasets/invoices")

	if reqA.Header.Get("Authorization") == reqB.Header.Get("Authorization") {
		t.Fatal("signature identical across different request paths, want distinct signatures")
	}
}

// TestAuthenticator_SignsFreshPerRequest: legacy computes a new timestamp
// (and therefore signature) on every call, never caching — two Apply calls
// at different times must produce different signatures.
func TestAuthenticator_SignsFreshPerRequest(t *testing.T) {
	current := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	h := &Hooks{Now: func() time.Time { return current }}
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req1 := doRequest(t, auth, http.MethodGet, "http://example.invalid/datasets/orders")
	current = current.Add(1 * time.Second)
	req2 := doRequest(t, auth, http.MethodGet, "http://example.invalid/datasets/orders")

	if req1.Header.Get("Authorization") == req2.Header.Get("Authorization") {
		t.Fatal("Authorization identical across requests at different timestamps, want a fresh signature per request")
	}
	if req1.Header.Get("X-Infor-Timestamp") == req2.Header.Get("X-Infor-Timestamp") {
		t.Fatal("X-Infor-Timestamp identical across requests at different timestamps")
	}
}

// --- credential resolution: config fallback ------------------------------

// TestAuthenticator_AccessKeyIDAndUserIDFromConfig mirrors legacy's
// requester(): access_key_id/user_id are read from cfg.Config in legacy
// (not secrets), so the hook must accept them there too.
func TestAuthenticator_AccessKeyIDAndUserIDFromConfig(t *testing.T) {
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"access_key_id": "AKID-cfg", "user_id": "user-cfg"},
		Secrets: map[string]string{
			"secret_key": "shhh-secret",
			"api_key":    "data-api-key",
		},
	}
	h := New().(*Hooks)
	auth, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doRequest(t, auth, http.MethodGet, "http://example.invalid/datasets/orders")
	if got := req.Header.Get("X-Infor-AccessKeyId"); got != "AKID-cfg" {
		t.Fatalf("X-Infor-AccessKeyId = %q, want AKID-cfg", got)
	}
	if got := req.Header.Get("X-Infor-UserId"); got != "user-cfg" {
		t.Fatalf("X-Infor-UserId = %q, want user-cfg", got)
	}
}

// --- error paths ------------------------------------------------------

func TestAuthenticator_MissingAccessKeyIDIsError(t *testing.T) {
	cfg := baseCfg()
	delete(cfg.Secrets, "access_key_id")
	h := New().(*Hooks)
	if _, err := h.Authenticator(context.Background(), cfg, baseSpec()); err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing access_key_id")
	} else if !strings.Contains(err.Error(), "access_key_id") {
		t.Fatalf("error = %q, want it to mention access_key_id", err.Error())
	}
}

func TestAuthenticator_MissingUserIDIsError(t *testing.T) {
	cfg := baseCfg()
	delete(cfg.Secrets, "user_id")
	h := New().(*Hooks)
	if _, err := h.Authenticator(context.Background(), cfg, baseSpec()); err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing user_id")
	} else if !strings.Contains(err.Error(), "user_id") {
		t.Fatalf("error = %q, want it to mention user_id", err.Error())
	}
}

func TestAuthenticator_MissingSecretKeyIsError(t *testing.T) {
	cfg := baseCfg()
	delete(cfg.Secrets, "secret_key")
	h := New().(*Hooks)
	if _, err := h.Authenticator(context.Background(), cfg, baseSpec()); err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing secret_key")
	} else if !strings.Contains(err.Error(), "secret_key") {
		t.Fatalf("error = %q, want it to mention secret_key", err.Error())
	}
}

func TestAuthenticator_MissingAPIKeyIsError(t *testing.T) {
	cfg := baseCfg()
	delete(cfg.Secrets, "api_key")
	h := New().(*Hooks)
	if _, err := h.Authenticator(context.Background(), cfg, baseSpec()); err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing api_key")
	} else if !strings.Contains(err.Error(), "api_key") {
		t.Fatalf("error = %q, want it to mention api_key", err.Error())
	}
}

// --- secret redaction ------------------------------------------------------

func TestAuthenticator_ErrorsNeverContainSecretText(t *testing.T) {
	cfg := baseCfg()
	delete(cfg.Secrets, "secret_key")
	h := New().(*Hooks)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("expected an error for missing secret_key")
	}
	for _, marker := range []string{"AKID123", "data-api-key"} {
		if strings.Contains(err.Error(), marker) {
			t.Fatalf("error text contains secret marker %q: %s", marker, err.Error())
		}
	}
}
