package engine

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestDeclaredSessionAuthRejectsDynamicTokenRoute(t *testing.T) {
	_, err := buildDeclaredSessionAuthenticator(AuthSpec{Mode: "declared_session", TokenURL: "{{ config.session_url }}", Username: "{{ config.username }}", Password: "{{ secrets.password }}"}, Vars{Config: map[string]string{"username": "person@example.test"}, Secrets: map[string]string{"password": "test-password"}}, nil)
	if err == nil {
		t.Fatal("dynamic session route was accepted")
	}
}

type declaredSessionRoundTripFunc func(*http.Request) (*http.Response, error)

func (f declaredSessionRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestDeclaredSessionAuthUsesResolvedTenantOrigin(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	requests := 0
	http.DefaultTransport = declaredSessionRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		switch request.URL.String() {
		case "http://127.0.0.1:8080/api/session":
			if request.Method != http.MethodPost {
				t.Fatalf("session method = %s, want POST", request.Method)
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"id":"issued-session"}`)), Request: request}, nil
		case "http://127.0.0.1:8080/api/card":
			if request.Header.Get("X-Metabase-Session") != "issued-session" {
				t.Fatal("card request omitted declared session header")
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`[]`)), Request: request}, nil
		default:
			t.Fatalf("session auth request = %s, want declared tenant route", request.URL)
			return nil, nil
		}
	})

	runtime, err := newRuntime(context.Background(), Bundle{
		Name: "metabase",
		HTTP: HTTPBase{
			URL:          "https://unused.invalid",
			TenantOrigin: &TenantOriginSpec{ConfigKey: "instance_api_url", AppendPath: "/api", AllowLoopbackHTTP: true},
			Auth: []AuthSpec{{
				Mode:     "declared_session",
				TokenURL: "https://unused.invalid/session",
				Username: "{{ config.username }}",
				Password: "{{ secrets.password }}",
			}},
		},
	}, connectors.RuntimeConfig{
		Config:  map[string]string{"instance_api_url": "http://127.0.0.1:8080", "username": "person@example.test"},
		Secrets: map[string]string{"password": "test-password"},
	}, nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	requester, err := runtime.RequesterFor(http.MethodGet, "/card")
	if err != nil {
		t.Fatalf("RequesterFor: %v", err)
	}
	if _, err := requester.Do(context.Background(), http.MethodGet, "/card", nil, nil); err != nil {
		t.Fatalf("declared session request: %v", err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want one session exchange plus card request", requests)
	}
}
