package loopback

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCallbackHandlerRedactsProviderErrorDetails(t *testing.T) {
	const marker = "redaction-sentinel"
	results := make(chan result, 1)
	flow := &Flow{}
	handler := flow.callbackHandler("/callback", "expected-state", results)

	query := url.Values{}
	query.Set("state", "expected-state")
	query.Set("error", "unexpected-"+marker)
	query.Set("error_description", marker)
	req := httptest.NewRequest(http.MethodGet, "/callback?"+query.Encode(), nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	got := <-results
	if got.err == nil {
		t.Fatal("callback result error = nil, want provider error")
	}
	if strings.Contains(got.err.Error(), marker) {
		t.Fatalf("callback error exposed provider detail: %v", got.err)
	}
	if !strings.Contains(got.err.Error(), "provider_error") {
		t.Fatalf("callback error = %v, want generic provider_error code", got.err)
	}
}

func TestExchangeCodeRedactsProviderErrorDetails(t *testing.T) {
	const marker = "redaction-sentinel"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":             "unexpected-" + marker,
			"error_description": marker,
		})
	}))
	defer server.Close()

	flow := &Flow{cfg: Config{TokenURL: server.URL, ClientID: "client", HTTPClient: server.Client()}}
	_, err := flow.exchangeCode(context.Background(), "code", "verifier", "http://127.0.0.1/callback")
	if err == nil {
		t.Fatal("exchangeCode() error = nil, want provider error")
	}
	if strings.Contains(err.Error(), marker) {
		t.Fatalf("token exchange error exposed provider detail: %v", err)
	}
	if !strings.Contains(err.Error(), "provider_error") {
		t.Fatalf("token exchange error = %v, want generic provider_error code", err)
	}
}
