package reddit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	reddithooks "polymetrics.ai/internal/connectors/hooks/reddit"
)

func TestAuthenticator_ExchangesRefreshTokenAndSetsBearer(t *testing.T) {
	const wantAccessToken = "fresh_access_token_fixture"

	var gotMethod, gotPath, gotAuthHeader, gotUserAgent, gotContentType string
	var gotForm url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuthHeader = r.Header.Get("Authorization")
		gotUserAgent = r.Header.Get("User-Agent")
		gotContentType = r.Header.Get("Content-Type")
		_ = r.ParseForm()
		gotForm = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": wantAccessToken,
			"token_type":   "bearer",
			"expires_in":   3600,
			"scope":        "read",
		})
	}))
	defer srv.Close()

	h := reddithooks.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"token_url":       srv.URL + "/api/v1/access_token",
			"reddit_username": "my_bot_account",
		},
		Secrets: map[string]string{
			"refresh_token": "rt_fixture",
			"client_id":     "cid_fixture",
			"client_secret": "csecret_fixture",
		},
	}

	authenticator, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{Mode: "custom", Hook: "reddit"})
	if err != nil {
		t.Fatalf("Authenticator() error = %v", err)
	}
	if authenticator == nil {
		t.Fatal("Authenticator() = nil, want a non-nil connsdk.Authenticator")
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("token request method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/v1/access_token" {
		t.Fatalf("token request path = %q, want /api/v1/access_token", gotPath)
	}
	if gotContentType != "application/x-www-form-urlencoded" {
		t.Fatalf("token request Content-Type = %q, want application/x-www-form-urlencoded", gotContentType)
	}
	if !strings.HasPrefix(gotAuthHeader, "Basic ") {
		t.Fatalf("token request Authorization = %q, want HTTP Basic", gotAuthHeader)
	}
	if gotForm.Get("grant_type") != "refresh_token" {
		t.Fatalf("form grant_type = %q, want refresh_token", gotForm.Get("grant_type"))
	}
	if gotForm.Get("refresh_token") != "rt_fixture" {
		t.Fatalf("form refresh_token = %q, want rt_fixture", gotForm.Get("refresh_token"))
	}
	if !strings.Contains(gotUserAgent, "by /u/my_bot_account") {
		t.Fatalf("token request User-Agent = %q, want a conforming Reddit User-Agent", gotUserAgent)
	}

	req, err := http.NewRequest(http.MethodGet, "https://oauth.reddit.com/r/golang/new", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := authenticator.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer "+wantAccessToken {
		t.Fatalf("outbound Authorization = %q, want Bearer %s", got, wantAccessToken)
	}
}

func TestAuthenticator_MissingRefreshTokenErrors(t *testing.T) {
	h := reddithooks.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{}, Secrets: map[string]string{"client_id": "cid"}}
	if _, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{Mode: "custom", Hook: "reddit"}); err == nil {
		t.Fatal("Authenticator() error = nil, want error for missing refresh_token")
	}
}

func TestAuthenticator_MissingClientIDErrors(t *testing.T) {
	h := reddithooks.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{}, Secrets: map[string]string{"refresh_token": "rt"}}
	if _, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{Mode: "custom", Hook: "reddit"}); err == nil {
		t.Fatal("Authenticator() error = nil, want error for missing client_id")
	}
}

func TestAuthenticator_ServerErrorPropagates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer srv.Close()

	h := reddithooks.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"token_url": srv.URL},
		Secrets: map[string]string{"refresh_token": "rt", "client_id": "cid"},
	}
	if _, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{Mode: "custom", Hook: "reddit"}); err == nil {
		t.Fatal("Authenticator() error = nil, want error for non-2xx response")
	}
}
