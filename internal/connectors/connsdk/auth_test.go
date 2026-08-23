package connsdk

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

func assertCredentialBytes(t *testing.T, got, want string) {
	t.Helper()
	if gotLength, wantLength := len(got), len(want); gotLength != wantLength {
		t.Fatalf("credential length = %d, want %d", gotLength, wantLength)
	}
	if gotHash, wantHash := sha256.Sum256([]byte(got)), sha256.Sum256([]byte(want)); gotHash != wantHash {
		t.Fatalf("credential SHA-256 = %x, want %x", gotHash, wantHash)
	}
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

func TestAuthenticatorsPreserveCredentialBytes(t *testing.T) {
	t.Run("bearer header", func(t *testing.T) {
		token := strings.Repeat("header-canary-", 1024) + "\t "
		req := newReq(t)
		if err := Bearer(token).Apply(context.Background(), req); err != nil {
			t.Fatalf("Apply: %v", err)
		}
		assertCredentialBytes(t, req.Header.Get("Authorization"), "Bearer "+token)
	})

	t.Run("API key header", func(t *testing.T) {
		key := "\tapi-header-canary "
		req := newReq(t)
		if err := APIKeyHeader("X-API-Key", key, "Token ").Apply(context.Background(), req); err != nil {
			t.Fatalf("Apply: %v", err)
		}
		assertCredentialBytes(t, req.Header.Get("X-API-Key"), "Token "+key)
	})

	t.Run("Basic", func(t *testing.T) {
		username := "\tbasic-user"
		password := "basic-password "
		req := newReq(t)
		if err := Basic(username, password).Apply(context.Background(), req); err != nil {
			t.Fatalf("Apply: %v", err)
		}
		gotUsername, gotPassword, ok := req.BasicAuth()
		if !ok {
			t.Fatal("BasicAuth() = false")
		}
		assertCredentialBytes(t, gotUsername, username)
		assertCredentialBytes(t, gotPassword, password)
	})

	t.Run("API key query", func(t *testing.T) {
		key := "query-canary\n"
		req := newReq(t)
		if err := APIKeyQuery("api_key", key).Apply(context.Background(), req); err != nil {
			t.Fatalf("Apply: %v", err)
		}
		assertCredentialBytes(t, req.URL.Query().Get("api_key"), key)
	})
}

func TestHeaderAuthenticatorsRejectProhibitedControlCharacters(t *testing.T) {
	for _, tt := range []struct {
		name string
		auth Authenticator
	}{
		{name: "bearer newline", auth: Bearer("bearer-canary\nvalue")},
		{name: "API key header carriage return", auth: APIKeyHeader("X-API-Key", "api-key-canary\rvalue", "")},
		{name: "Basic optional password newline", auth: BasicWithRequirements("basic-user", "basic-password\nvalue", true, false)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := newReq(t)
			err := tt.auth.Apply(context.Background(), req)
			var invalid *credential.InvalidSecretValueError
			if !errors.As(err, &invalid) {
				t.Fatalf("Apply() error is not typed invalid-secret classification: %T", err)
			}
			if got := req.Header.Get("Authorization"); got != "" {
				t.Fatal("Authorization header emitted for prohibited credential bytes")
			}
			if got := req.Header.Get("X-API-Key"); got != "" {
				t.Fatal("API-key header emitted for prohibited credential bytes")
			}
		})
	}
}

func TestRequiredAuthenticatorsRejectEmptyCredentialBeforeRequestMutation(t *testing.T) {
	for _, tt := range []struct {
		name string
		auth Authenticator
	}{
		{name: "bearer", auth: Bearer("")},
		{name: "bearer LF-only", auth: Bearer("\n")},
		{name: "bearer CRLF-only", auth: Bearer("\r\n")},
		{name: "basic", auth: Basic("user", "")},
		{name: "API key header", auth: APIKeyHeader("X-API-Key", "", "Token ")},
		{name: "API key query", auth: APIKeyQuery("api_key", "")},
		{name: "API key query LF-only", auth: APIKeyQuery("api_key", "\n")},
		{name: "API key query CRLF-only", auth: APIKeyQuery("api_key", "\r\n")},
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
		{name: "client ID LF-only", clientID: "\n", clientSecret: "synthetic-client-secret"},
		{name: "client secret CRLF-only", clientID: "synthetic-client-id", clientSecret: "\r\n"},
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

func TestOAuth2ClientCredentialsPreservesFormCredentialBytes(t *testing.T) {
	clientID := "client-id-canary\n"
	clientSecret := strings.Repeat("client-secret-canary-", 1024) + "\n"
	var gotClientID, gotClientSecret string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		gotClientID = r.PostForm.Get("client_id")
		gotClientSecret = r.PostForm.Get("client_secret")
		_, _ = w.Write([]byte(`{"access_token":"access-token","expires_in":3600}`))
	}))
	defer srv.Close()

	auth := &OAuth2ClientCredentials{
		TokenURL:     srv.URL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	if err := auth.Apply(context.Background(), newReq(t)); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	assertCredentialBytes(t, gotClientID, clientID)
	assertCredentialBytes(t, gotClientSecret, clientSecret)
}

func TestOAuth2ClientCredentialsRejectsHeaderControlToken(t *testing.T) {
	auth := &OAuth2ClientCredentials{
		token:   "access-token\ncanary",
		expires: time.Now().Add(time.Hour),
	}
	req := newReq(t)
	err := auth.Apply(context.Background(), req)
	var invalid *credential.InvalidSecretValueError
	if !errors.As(err, &invalid) {
		t.Fatalf("Apply() error is not typed invalid-secret classification: %T", err)
	}
	if got := req.Header.Get("Authorization"); got != "" {
		t.Fatal("Authorization header emitted for prohibited access-token bytes")
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
