package reddit_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func roundTripResponse(r *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    r,
	}
}

func TestExecuteWrite_AcquiresBoundS3LeaseWithoutForwardingOAuth(t *testing.T) {
	const pngMIME = "image/png"
	png := []byte{"\x89"[0], 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "emoji.png"), png, 0o600); err != nil {
		t.Fatalf("write fixture image: %v", err)
	}
	digest := sha256.Sum256(png)

	for _, tc := range []struct {
		name      string
		action    string
		leasePath string
	}{
		{name: "emoji", action: "emoji_asset_upload_s3", leasePath: "/api/v1/golang/emoji_asset_upload_s3.json"},
		{name: "widget", action: "widget_image_upload_s3", leasePath: "/r/golang/api/widget_image_upload_s3"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var leaseCalls, uploadCalls int
			requester := &connsdk.Requester{
				BaseURL: "https://oauth.reddit.test",
				Client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					switch r.URL.Host {
					case "oauth.reddit.test":
						leaseCalls++
						if r.Method != http.MethodPost || r.URL.Path != tc.leasePath {
							t.Errorf("lease request = %s %s, want POST %s", r.Method, r.URL.Path, tc.leasePath)
						}
						if got := r.Header.Get("Authorization"); got != "Bearer reddit-fixture-token" {
							t.Errorf("lease Authorization = %q, want Reddit bearer token", got)
						}
						if err := r.ParseForm(); err != nil {
							t.Errorf("parse lease form: %v", err)
						}
						if got := r.PostForm.Get("filepath"); got != "emoji.png" {
							t.Errorf("lease filepath = %q, want basename only", got)
						}
						if got := r.PostForm.Get("mimetype"); got != pngMIME {
							t.Errorf("lease mimetype = %q, want %q", got, pngMIME)
						}
						return roundTripResponse(r, http.StatusOK, `{"s3UploadLease":{"action":"https://s3.amazonaws.com/reddit-fixture","fields":[{"name":"key","value":"golang/emoji.png"},{"name":"policy","value":"lease-policy"}]}}`), nil
					case "s3.amazonaws.com":
						uploadCalls++
						if r.Method != http.MethodPost {
							t.Errorf("S3 upload method = %s, want POST form upload", r.Method)
						}
						if got := r.Header.Get("Authorization"); got != "" {
							t.Errorf("S3 Authorization = %q, must not receive Reddit OAuth credentials", got)
						}
						if err := r.ParseMultipartForm(1 << 20); err != nil {
							t.Errorf("parse S3 multipart form: %v", err)
						}
						if got := r.MultipartForm.Value["key"]; len(got) != 1 || got[0] != "golang/emoji.png" {
							t.Errorf("S3 key field = %#v", got)
						}
						file, _, err := r.FormFile("file")
						if err != nil {
							t.Errorf("S3 file field: %v", err)
						} else {
							got, readErr := io.ReadAll(file)
							_ = file.Close()
							if readErr != nil || !bytes.Equal(got, png) {
								t.Errorf("S3 file = %x, read error = %v", got, readErr)
							}
						}
						return roundTripResponse(r, http.StatusNoContent, ""), nil
					default:
						t.Errorf("unexpected outbound host %q", r.URL.Host)
						return roundTripResponse(r, http.StatusBadGateway, ""), nil
					}
				})},
				Auth:           connsdk.Bearer("reddit-fixture-token"),
				DefaultHeaders: map[string]string{"X-Reddit-Fixture": "present"},
				DisableRetries: true,
			}
			rt := &engine.Runtime{
				Requester: requester,
				Config: connectors.RuntimeConfig{
					ProjectDir: dir,
					Config:     map[string]string{"subreddit": "golang"},
					ApprovedPayloadSHA256: map[string]string{
						connectors.PayloadApprovalKey(0, "file_path"): hex.EncodeToString(digest[:]),
					},
				},
			}
			handled, err := reddithooks.New().ExecuteWrite(context.Background(), engine.WriteAction{
				Name:   tc.action,
				Method: http.MethodPost,
				Path:   tc.leasePath,
			}, connectors.Record{"file_path": "emoji.png", "mimetype": pngMIME}, rt)
			if err != nil {
				t.Fatalf("ExecuteWrite(%s): %v", tc.action, err)
			}
			if !handled {
				t.Fatalf("ExecuteWrite(%s) did not handle the declared lease action", tc.action)
			}
			if leaseCalls != 1 || uploadCalls != 1 {
				t.Fatalf("lease/upload calls = %d/%d, want 1/1", leaseCalls, uploadCalls)
			}
		})
	}
}

func TestExecuteWrite_RejectsUntrustedLeaseHostBeforeUpload(t *testing.T) {
	dir := t.TempDir()
	png := []byte{"\x89"[0], 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0}
	if err := os.WriteFile(filepath.Join(dir, "emoji.png"), png, 0o600); err != nil {
		t.Fatalf("write fixture image: %v", err)
	}
	digest := sha256.Sum256(png)
	var calls int
	rt := &engine.Runtime{
		Requester: &connsdk.Requester{
			BaseURL: "https://oauth.reddit.test",
			Client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Host == "oauth.reddit.test" {
					return roundTripResponse(r, http.StatusOK, `{"s3UploadLease":{"action":"https://not-s3.example/upload","fields":[]}}`), nil
				}
				t.Errorf("untrusted lease caused outbound request to %q", r.URL.Host)
				return roundTripResponse(r, http.StatusBadGateway, ""), nil
			})},
			DisableRetries: true,
		},
		Config: connectors.RuntimeConfig{
			ProjectDir: dir,
			Config:     map[string]string{"subreddit": "golang"},
			ApprovedPayloadSHA256: map[string]string{
				connectors.PayloadApprovalKey(0, "file_path"): hex.EncodeToString(digest[:]),
			},
		},
	}
	_, err := reddithooks.New().ExecuteWrite(context.Background(), engine.WriteAction{
		Name:   "emoji_asset_upload_s3",
		Method: http.MethodPost,
		Path:   "/api/v1/{{ config.subreddit }}/emoji_asset_upload_s3.json",
	}, connectors.Record{"file_path": "emoji.png", "mimetype": "image/png"}, rt)
	if err == nil || !strings.Contains(err.Error(), "S3") {
		t.Fatalf("ExecuteWrite() error = %v, want rejected S3 lease host", err)
	}
	if calls != 1 {
		t.Fatalf("lease/upload calls = %d, want only the lease request", calls)
	}
}
