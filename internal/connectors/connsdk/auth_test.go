package connsdk

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/credential"
)

func newReq(t *testing.T) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "https://example.com/path?a=1", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	return req
}

func TestBearerSetsAuthorization(t *testing.T) {
	req := newReq(t)
	if err := Bearer("tok123").Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer tok123" {
		t.Fatalf("Authorization = %q", got)
	}
}

func TestAPIKeyHeaderWithPrefix(t *testing.T) {
	req := newReq(t)
	if err := APIKeyHeader("X-Api-Key", "abc", "Token ").Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := req.Header.Get("X-Api-Key"); got != "Token abc" {
		t.Fatalf("X-Api-Key = %q", got)
	}
}

func TestBasicAuth(t *testing.T) {
	req := newReq(t)
	if err := Basic("user", "pass").Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	user, pass, ok := req.BasicAuth()
	if !ok || user != "user" || pass != "pass" {
		t.Fatalf("BasicAuth() = %q,%q,%v", user, pass, ok)
	}
}

func TestBasicWithRequirementsAllowsBlankOptionalPassword(t *testing.T) {
	const apiKey = "basic-optional-password-canary"
	req := newReq(t)
	if err := BasicWithRequirements(apiKey, "", true, false).Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	username, password, ok := req.BasicAuth()
	wantHash := sha256.Sum256([]byte(apiKey))
	gotHash := sha256.Sum256([]byte(username))
	if !ok || len(username) != len(apiKey) || gotHash != wantHash || len(password) != 0 {
		t.Fatal("BasicWithRequirements did not preserve the declaration-authorized blank password")
	}
}

func TestAPIKeyQueryAddsParam(t *testing.T) {
	req := newReq(t)
	if err := APIKeyQuery("api_key", "secretval").Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := req.URL.Query().Get("api_key"); got != "secretval" {
		t.Fatalf("api_key = %q", got)
	}
	if got := req.URL.Query().Get("a"); got != "1" {
		t.Fatalf("existing query param dropped: a = %q", got)
	}
}

func TestRequiredAuthenticatorsRejectEmptyCredentialBeforeRequestMutation(t *testing.T) {
	for _, tt := range []struct {
		name string
		auth Authenticator
	}{
		{name: "bearer", auth: Bearer("")},
		{name: "basic", auth: Basic("user", "")},
		{name: "API key header", auth: APIKeyHeader("X-API-Key", "", "Token ")},
		{name: "API key query", auth: APIKeyQuery("api_key", "")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := newReq(t)
			if err := tt.auth.Apply(context.Background(), req); err == nil {
				t.Fatal("Apply() accepted an empty required credential")
			} else {
				var empty *credential.EmptySecretError
				if !errors.As(err, &empty) {
					t.Fatalf("Apply() error is not typed empty-secret classification: %T", err)
				}
			}
			if got := req.Header.Get("Authorization"); got != "" {
				t.Fatalf("Authorization header emitted for empty credential: %q", got)
			}
			if got := req.Header.Get("X-API-Key"); got != "" {
				t.Fatalf("API-key header emitted for empty credential: %q", got)
			}
			if got := req.URL.Query().Get("api_key"); got != "" {
				t.Fatal("API-key query parameter emitted for empty credential")
			}
		})
	}
}

func TestOAuth2ClientCredentialsRejectsEmptyRequiredMaterialBeforeTokenRequest(t *testing.T) {
	for _, tt := range []struct {
		name         string
		clientID     string
		clientSecret string
	}{
		{name: "client ID", clientSecret: "synthetic-client-secret"},
		{name: "client secret", clientID: "synthetic-client-id"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var tokenCalls int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				atomic.AddInt32(&tokenCalls, 1)
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer srv.Close()

			auth := &OAuth2ClientCredentials{
				TokenURL:     srv.URL,
				ClientID:     tt.clientID,
				ClientSecret: tt.clientSecret,
			}
			req := newReq(t)
			err := auth.Apply(context.Background(), req)
			if err == nil {
				t.Fatal("Apply() accepted empty required OAuth2 material")
			}
			var empty *credential.EmptySecretError
			if !errors.As(err, &empty) {
				t.Fatalf("Apply() error is not typed empty-secret classification: %T", err)
			}
			if got := atomic.LoadInt32(&tokenCalls); got != 0 {
				t.Fatalf("token endpoint calls = %d, want 0", got)
			}
			if got := req.Header.Get("Authorization"); got != "" {
				t.Fatalf("Authorization header emitted for empty OAuth2 material: %q", got)
			}
		})
	}
}

func TestOAuth2ClientCredentialsFetchesAndCaches(t *testing.T) {
	var tokenCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&tokenCalls, 1)
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		if r.Form.Get("grant_type") != "client_credentials" {
			t.Errorf("grant_type = %q", r.Form.Get("grant_type"))
		}
		_, _ = w.Write([]byte(`{"access_token":"AT","token_type":"Bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	now := time.Unix(1_700_000_000, 0)
	auth := &OAuth2ClientCredentials{
		TokenURL:     srv.URL,
		ClientID:     "id",
		ClientSecret: "sec",
		Now:          func() time.Time { return now },
	}

	for i := 0; i < 3; i++ {
		req := newReq(t)
		if err := auth.Apply(context.Background(), req); err != nil {
			t.Fatalf("Apply: %v", err)
		}
		if got := req.Header.Get("Authorization"); got != "Bearer AT" {
			t.Fatalf("Authorization = %q", got)
		}
	}
	if got := atomic.LoadInt32(&tokenCalls); got != 1 {
		t.Fatalf("token endpoint calls = %d, want 1 (cached)", got)
	}
}
