package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// tenantTokenAuth fetches and caches a Feishu/Lark tenant_access_token using the
// internal-app exchange (app_id + app_secret -> token), refreshing automatically
// before expiry. It implements connsdk.Authenticator so the Requester can apply
// a Bearer header to every Bitable call. Secret values are never logged.
type tenantTokenAuth struct {
	client    *http.Client
	baseURL   string
	appID     string
	appSecret string

	// now is injectable for tests; defaults to time.Now.
	now func() time.Time

	mu      sync.Mutex
	cached  string
	expires time.Time
}

// Apply satisfies connsdk.Authenticator: ensure a fresh token and set the
// Authorization header.
func (a *tenantTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	tok, err := a.token(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	return nil
}

func (a *tenantTokenAuth) clock() time.Time {
	if a.now != nil {
		return a.now()
	}
	return time.Now()
}

// token returns a valid tenant_access_token, fetching a new one if the cache is
// empty or within 60s of expiry.
func (a *tenantTokenAuth) token(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cached != "" && a.clock().Add(60*time.Second).Before(a.expires) {
		return a.cached, nil
	}
	if strings.TrimSpace(a.appID) == "" || strings.TrimSpace(a.appSecret) == "" {
		return "", errors.New("feishu: app_id and app_secret are required for token exchange")
	}

	body, err := json.Marshal(map[string]string{
		"app_id":     a.appID,
		"app_secret": a.appSecret,
	})
	if err != nil {
		return "", fmt.Errorf("feishu: encode token request: %w", err)
	}
	tokenURL := strings.TrimRight(a.baseURL, "/") + "/" + feishuTokenPath

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("feishu: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("feishu: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("feishu: token endpoint returned %s", resp.Status)
	}

	var out struct {
		Code              json.Number `json:"code"`
		Msg               string      `json:"msg"`
		TenantAccessToken string      `json:"tenant_access_token"`
		Expire            json.Number `json:"expire"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("feishu: decode token response: %w", err)
	}
	if code := out.Code.String(); code != "" && code != "0" {
		return "", fmt.Errorf("feishu: token exchange failed: code %s: %s", code, out.Msg)
	}
	if strings.TrimSpace(out.TenantAccessToken) == "" {
		return "", errors.New("feishu: token response missing tenant_access_token")
	}

	ttl := 2 * time.Hour
	if secs, err := out.Expire.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.cached = out.TenantAccessToken
	a.expires = a.clock().Add(ttl)
	return a.cached, nil
}
