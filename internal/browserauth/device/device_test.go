package device_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/browserauth/device"
)

func TestDeviceFlowSuccessAfterPending(t *testing.T) {
	var pollCount int32
	mux := http.NewServeMux()
	mux.HandleFunc("/device_authorization", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"device_code":               "test-device-code",
			"user_code":                 "ABCD-EFGH",
			"verification_uri":          "https://example.invalid/activate",
			"verification_uri_complete": "https://example.invalid/activate?user_code=ABCD-EFGH",
			"expires_in":                60,
			"interval":                  1,
		})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&pollCount, 1)
		w.Header().Set("Content-Type", "application/json")
		if n < 2 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "authorization_pending"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"scope":        "read",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	var gotUserCode, gotURI string
	flow, err := device.New(device.Config{
		DeviceAuthURL: server.URL + "/device_authorization",
		TokenURL:      server.URL + "/token",
		ClientID:      "client-123",
		OnUserCode: func(userCode, verificationURI, _ string) {
			gotUserCode = userCode
			gotURI = verificationURI
		},
		Timeout:    5 * time.Second,
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if flow.Name() != "device_code" {
		t.Fatalf("Name() = %q", flow.Name())
	}

	cred, err := flow.Login(context.Background())
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if gotUserCode != "ABCD-EFGH" {
		t.Fatalf("OnUserCode userCode = %q", gotUserCode)
	}
	if gotURI != "https://example.invalid/activate" {
		t.Fatalf("OnUserCode verificationURI = %q", gotURI)
	}
	if cred.OAuth == nil || cred.OAuth.AccessToken != "test-access-token" {
		t.Fatalf("Login() credential = %+v", cred)
	}
	if cred.Session != nil {
		t.Fatalf("device flow must never set Session, got %+v", cred.Session)
	}
}

func TestDeviceFlowAccessDenied(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/device_authorization", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"device_code":      "test-device-code",
			"user_code":        "ABCD-EFGH",
			"verification_uri": "https://example.invalid/activate",
			"expires_in":       60,
			"interval":         1,
		})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "access_denied"})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	flow, err := device.New(device.Config{
		DeviceAuthURL: server.URL + "/device_authorization",
		TokenURL:      server.URL + "/token",
		ClientID:      "client-123",
		OnUserCode:    func(string, string, string) {},
		Timeout:       5 * time.Second,
		HTTPClient:    server.Client(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = flow.Login(context.Background())
	if err == nil || !strings.Contains(err.Error(), "denied") {
		t.Fatalf("Login() error = %v, want access denied", err)
	}
}

func TestDeviceFlowRejectsMissingVerificationURI(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/device_authorization", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"device_code": "test-device-code",
			"user_code":   "ABCD-EFGH",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	called := false
	flow, err := device.New(device.Config{
		DeviceAuthURL: server.URL + "/device_authorization",
		TokenURL:      server.URL + "/token",
		ClientID:      "client-123",
		OnUserCode: func(string, string, string) {
			called = true
		},
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = flow.Login(context.Background())
	if err == nil || !strings.Contains(err.Error(), "verification_uri") {
		t.Fatalf("Login() error = %v, want verification_uri validation failure", err)
	}
	if called {
		t.Fatal("OnUserCode called for an invalid device authorization response")
	}
}

func TestDeviceFlowRequiresOnUserCode(t *testing.T) {
	_, err := device.New(device.Config{
		DeviceAuthURL: "https://example.invalid/device_authorization",
		TokenURL:      "https://example.invalid/token",
		ClientID:      "client-123",
	})
	if err == nil {
		t.Fatalf("New() without OnUserCode: want error, got nil")
	}
}
