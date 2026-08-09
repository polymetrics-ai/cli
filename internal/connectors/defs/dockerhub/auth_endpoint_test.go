package dockerhub_test

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	connectorDefs "polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	dockerhubhooks "polymetrics.ai/internal/connectors/hooks/dockerhub"
)

func TestDockerhubPATExchangeUsesDedicatedAuthURL(t *testing.T) {
	const pat = "fixture-pat-value"
	var foreignAuthHits int32
	var foreignDataHits int32
	foreign := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			atomic.AddInt32(&foreignAuthHits, 1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"token": "foreign-session"})
		case http.MethodGet:
			atomic.AddInt32(&foreignDataHits, 1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"results": []any{}})
		default:
			t.Errorf("foreign base request method = %s, want GET", r.Method)
		}
	}))
	t.Cleanup(foreign.Close)

	var authHits int32
	authServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/users/login" {
			t.Errorf("auth request = %s %s, want POST /v2/users/login", r.Method, r.URL.Path)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode auth request: %v", err)
		}
		if body["password"] != pat {
			t.Error("auth endpoint did not receive the configured Docker Hub PAT")
		}
		atomic.AddInt32(&authHits, 1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "auth-session"})
	}))
	t.Cleanup(authServer.Close)

	roots := x509.NewCertPool()
	roots.AddCert(foreign.Certificate())
	roots.AddCert(authServer.Certificate())
	transport, ok := foreign.Client().Transport.(*http.Transport)
	if !ok {
		t.Fatalf("foreign client transport = %T, want *http.Transport", foreign.Client().Transport)
	}
	transport = transport.Clone()
	if transport.TLSClientConfig == nil {
		t.Fatal("foreign client TLS config is nil")
	}
	tlsConfig := transport.TLSClientConfig.Clone()
	tlsConfig.RootCAs = roots
	transport.TLSClientConfig = tlsConfig
	client := &http.Client{Transport: transport}
	t.Cleanup(transport.CloseIdleConnections)

	bundle, err := engine.Load(connectorDefs.FS, "dockerhub")
	if err != nil {
		t.Fatalf("load Docker Hub bundle: %v", err)
	}
	if got := bundle.Spec.Defaults()["auth_url"]; got != "https://hub.docker.com/v2" {
		t.Fatalf("auth_url default = %q, want Docker Hub authentication API base", got)
	}
	if got := bundle.HTTP.Auth[0].TokenURL; got != "{{ config.auth_url }}/users/login" {
		t.Fatalf("Docker Hub PAT token URL template = %q, want config.auth_url/users/login", got)
	}
	for _, operation := range []string{"dockerhub.create_auth_token", "dockerhub.create_login", "dockerhub.create_2fa_login"} {
		var found *engine.OperationSpec
		for i := range bundle.Operations {
			if bundle.Operations[i].ID == operation {
				found = &bundle.Operations[i]
				break
			}
		}
		if found == nil || found.REST == nil || found.REST.BaseURL != "{{ config.auth_url }}" || !found.ResponseSensitive {
			t.Fatalf("%s must use the dedicated response-sensitive authentication base", operation)
		}
	}

	hooks := dockerhubhooks.New().(*dockerhubhooks.Hooks)
	hooks.Client = client
	runtime, err := engine.NewRuntime(context.Background(), bundle, connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":        foreign.URL + "/v2",
			"auth_url":        authServer.URL + "/v2",
			"docker_username": "fixture-user",
			"namespace":       "fixture",
		},
		Secrets: map[string]string{"docker_pat": pat},
	}, hooks)
	if err != nil {
		t.Fatalf("new Docker Hub runtime: %v", err)
	}
	runtime.Requester.Client = client
	if _, err := runtime.Requester.Do(context.Background(), http.MethodGet, "/namespaces/fixture/repositories", nil, nil); err != nil {
		t.Fatalf("Docker Hub request through foreign data base: %v", err)
	}
	if got := atomic.LoadInt32(&authHits); got != 1 {
		t.Fatalf("dedicated auth endpoint hits = %d, want 1", got)
	}
	if got := atomic.LoadInt32(&foreignAuthHits); got != 0 {
		t.Fatalf("foreign base received %d PAT login request(s), want 0", got)
	}
	if got := atomic.LoadInt32(&foreignDataHits); got != 1 {
		t.Fatalf("foreign base data requests = %d, want 1", got)
	}
}
