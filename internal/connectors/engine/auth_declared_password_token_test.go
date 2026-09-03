package engine

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

type passwordTokenRoundTripFunc func(*http.Request) (*http.Response, error)

func (f passwordTokenRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestDeclaredPasswordTokenAuthSendsPasswordOnlyToFixedTokenRoute(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var tokenRequests atomic.Int64
	http.DefaultTransport = passwordTokenRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.String() != "https://api2.myhours.com/api/tokens/login" {
			t.Fatalf("token URL = %s, want fixed declared route", request.URL)
		}
		tokenRequests.Add(1)
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read token body: %v", err)
		}
		if string(body) != `{"email":"person@example.test","password":"test-password"}` {
			t.Fatal("token request omitted the declared email/password JSON binding")
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"accessToken":"issued-token"}`)), Request: request}, nil
	})

	auth, err := buildDeclaredPasswordTokenAuthenticator(AuthSpec{
		Mode:     "declared_password_token",
		TokenURL: "https://api2.myhours.com/api/tokens/login",
		Username: "{{ config.email }}",
		Password: "{{ secrets.password }}",
	}, Vars{Config: map[string]string{"email": "person@example.test"}, Secrets: map[string]string{"password": "test-password"}})
	if err != nil {
		t.Fatalf("buildDeclaredPasswordTokenAuthenticator: %v", err)
	}
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api2.myhours.com/api/Clients", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := auth.Apply(context.Background(), request); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if request.Header.Get("Authorization") != "Bearer issued-token" {
		t.Fatalf("Authorization = %q, want issued bearer token", request.Header.Get("Authorization"))
	}
	if request.Header.Get("password") != "" || request.URL.RawQuery != "" {
		t.Fatal("password leaked beyond the fixed token request")
	}
	if got := tokenRequests.Load(); got != 1 {
		t.Fatalf("token requests = %d, want one", got)
	}
}

func TestDeclaredPasswordTokenAuthRejectsDynamicOrNonHTTPSRoute(t *testing.T) {
	for _, tokenURL := range []string{"{{ config.token_url }}", "http://api2.myhours.com/api/tokens/login", "https://api2.myhours.com/api/tokens/login?next=1"} {
		if _, err := buildDeclaredPasswordTokenAuthenticator(AuthSpec{TokenURL: tokenURL, Username: "{{ config.email }}", Password: "{{ secrets.password }}"}, Vars{Config: map[string]string{"email": "person@example.test"}, Secrets: map[string]string{"password": "test-password"}}); err == nil {
			t.Fatalf("token URL %q was accepted", tokenURL)
		}
	}
}

func TestDeclaredPasswordTokenAuthRefusesRedirectBeforeCredentialReplay(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var originRequests, destinationRequests atomic.Int64
	http.DefaultTransport = passwordTokenRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.String() {
		case "https://api2.myhours.com/api/tokens/login":
			originRequests.Add(1)
			return &http.Response{
				StatusCode: http.StatusTemporaryRedirect,
				Header:     http.Header{"Location": []string{"https://redirect.example/token"}},
				Body:       io.NopCloser(strings.NewReader(`redirect`)),
				Request:    request,
			}, nil
		case "https://redirect.example/token":
			destinationRequests.Add(1)
			t.Fatal("password token redirect destination received a request")
			return nil, nil
		default:
			t.Fatalf("unexpected password token request %s", request.URL)
			return nil, nil
		}
	})

	auth, err := buildDeclaredPasswordTokenAuthenticator(AuthSpec{
		Mode:     "declared_password_token",
		TokenURL: "https://api2.myhours.com/api/tokens/login",
		Username: "{{ config.email }}",
		Password: "{{ secrets.password }}",
	}, Vars{Config: map[string]string{"email": "person@example.test"}, Secrets: map[string]string{"password": "test-password"}})
	if err != nil {
		t.Fatalf("buildDeclaredPasswordTokenAuthenticator: %v", err)
	}
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api2.myhours.com/api/Clients", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := auth.Apply(context.Background(), request); err == nil {
		t.Fatal("password token redirect was accepted")
	}
	if originRequests.Load() != 1 || destinationRequests.Load() != 0 {
		t.Fatalf("password token redirect requests origin/destination = %d/%d, want 1/0", originRequests.Load(), destinationRequests.Load())
	}
}
