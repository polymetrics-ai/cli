package device

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestDeviceAuthRedactsFailureBody(t *testing.T) {
	const marker = "redaction-sentinel"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(marker))
	}))
	defer server.Close()

	flow := &Flow{cfg: Config{DeviceAuthURL: server.URL, ClientID: "client", HTTPClient: server.Client()}}
	_, err := flow.requestDeviceAuth(context.Background())
	if err == nil {
		t.Fatal("requestDeviceAuth() error = nil, want status error")
	}
	if strings.Contains(err.Error(), marker) {
		t.Fatalf("device authorization error exposed response body: %v", err)
	}
	if !strings.Contains(err.Error(), "status 400") {
		t.Fatalf("device authorization error = %v, want status", err)
	}
}

func TestPollTokenRedactsProviderErrorDetails(t *testing.T) {
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
	_, err := flow.pollToken(context.Background(), "device-code")
	if err == nil {
		t.Fatal("pollToken() error = nil, want provider error")
	}
	if strings.Contains(err.Error(), marker) {
		t.Fatalf("token poll error exposed provider detail: %v", err)
	}
	if !strings.Contains(err.Error(), "provider_error") {
		t.Fatalf("token poll error = %v, want generic provider_error code", err)
	}
}
