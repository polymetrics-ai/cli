package device

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestDeviceFlowRedactsTransportErrors(t *testing.T) {
	const marker = "redaction-sentinel"
	flow := &Flow{cfg: Config{
		DeviceAuthURL: "https://provider.invalid/device",
		TokenURL:      "https://provider.invalid/token",
		ClientID:      "client",
		HTTPClient: &http.Client{Transport: errorRoundTripper{err: &url.Error{
			Op:  "Post",
			URL: "https://provider.invalid/token?device_code=" + marker,
			Err: errors.New("transport failure"),
		}}},
	}}

	for name, call := range map[string]func() error{
		"device authorization": func() error {
			_, err := flow.requestDeviceAuth(context.Background())
			return err
		},
		"token poll": func() error {
			_, err := flow.pollToken(context.Background(), "device-code")
			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := call()
			if err == nil {
				t.Fatal("flow error = nil, want transport error")
			}
			if strings.Contains(err.Error(), marker) {
				t.Fatalf("flow error exposed transport detail: %v", err)
			}
			if !strings.Contains(err.Error(), "oauth transport request failed") {
				t.Fatalf("flow error = %v, want generic transport error", err)
			}
		})
	}
}

func TestPollTokenRejectsNonOKSuccessPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "token"})
	}))
	defer server.Close()

	flow := &Flow{cfg: Config{TokenURL: server.URL, ClientID: "client", HTTPClient: server.Client()}}
	credential, err := flow.pollToken(context.Background(), "device-code")
	if err == nil {
		t.Fatal("pollToken() error = nil, want status error")
	}
	if credential != nil {
		t.Fatalf("pollToken() credential = %#v, want nil", credential)
	}
	if !strings.Contains(err.Error(), "status 400") {
		t.Fatalf("pollToken() error = %v, want status", err)
	}
}

type errorRoundTripper struct {
	err error
}

func (r errorRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, r.err
}
