package connsdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
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
