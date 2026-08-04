package googlecalendar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestHooksRegistered(t *testing.T) {
	h := engine.HooksFor("google-calendar")
	if h == nil {
		t.Fatal("registered hooks = nil")
	}
	if h.ConnectorName() != "google-calendar" {
		t.Fatalf("ConnectorName() = %q", h.ConnectorName())
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("hooks do not implement AuthHook")
	}
	if _, ok := h.(engine.CheckHook); ok {
		t.Fatal("hooks should not override declarative Check")
	}
	if _, ok := h.(engine.StreamHook); ok {
		t.Fatal("hooks should not override declarative ReadStream")
	}
}

func TestAuthenticatorConformanceFixtureNoOps(t *testing.T) {
	auth, err := Hooks{}.Authenticator(context.Background(), connectors.RuntimeConfig{ProjectDir: "__polymetrics_conformance_fixture__"}, engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "https://example.test/calendar", nil)
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "" {
		t.Fatalf("Authorization = %q, want empty", got)
	}
}

func TestAuthenticatorRefreshTokenExchange(t *testing.T) {
	var sawForm bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if r.Form.Get("grant_type") != "refresh_token" {
			t.Fatalf("grant_type = %q", r.Form.Get("grant_type"))
		}
		if r.Form.Get("client_id") != "calendar-client" || r.Form.Get("client_secret") != "calendar-client-secret" || r.Form.Get("refresh_token") != "calendar-refresh" {
			t.Fatalf("unexpected OAuth form keys")
		}
		sawForm = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "calendar-access", "expires_in": 3600})
	}))
	defer srv.Close()

	auth, err := Hooks{}.Authenticator(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "live"}}, engine.AuthSpec{TokenURL: srv.URL, ClientID: "calendar-client", ClientSecret: "calendar-client-secret", Token: "calendar-refresh"})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "https://example.test/calendar", nil)
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !sawForm {
		t.Fatal("OAuth refresh endpoint was not called")
	}
	if got := req.Header.Get("Authorization"); got != "Bearer calendar-access" {
		t.Fatalf("Authorization = %q", got)
	}
}

func TestAuthenticatorLiveSyntheticValuesStillRefreshes(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "calendar-access", "expires_in": 3600})
	}))
	defer srv.Close()

	auth, err := Hooks{}.Authenticator(context.Background(), connectors.RuntimeConfig{}, engine.AuthSpec{TokenURL: srv.URL, ClientID: "synthetic-conformance-secret", ClientSecret: "synthetic-conformance-secret", Token: "synthetic-conformance-secret"})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "https://example.test/calendar", nil)
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !called {
		t.Fatal("expected live auth to call refresh endpoint")
	}
	if got := req.Header.Get("Authorization"); got != "Bearer calendar-access" {
		t.Fatalf("Authorization = %q", got)
	}
}

func TestAuthenticatorMissingRefreshTokenErrorIsRedacted(t *testing.T) {
	_, err := Hooks{}.Authenticator(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "live"}}, engine.AuthSpec{TokenURL: "https://oauth2.googleapis.com/token", ClientID: "calendar-client", ClientSecret: "calendar-client-secret"})
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "calendar-client-secret") {
		t.Fatalf("error leaked secret: %v", err)
	}
	if !strings.Contains(err.Error(), "refresh token is required") {
		t.Fatalf("error = %v", err)
	}
}

func TestFreeBusyOperationDirectReadRejectsInvalidTimeBounds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("server should not be called for invalid typed bounds")
	}))
	defer srv.Close()

	b, err := engine.Load(defs.FS, "google-calendar")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	b.HTTP.URL = srv.URL + "/calendar/v3"
	conn := engine.New(b, Hooks{})
	_, err = conn.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
		Operation: "google-calendar.freebusy.query",
		Config:    connectors.RuntimeConfig{ProjectDir: "__polymetrics_conformance_fixture__"},
		Body: map[string]any{
			"timeMin": "not-a-date-time",
			"timeMax": "2020-01-02T00:00:00Z",
			"items":   []any{map[string]any{"id": "primary"}},
		},
		MaxBytes: 4096,
	})
	if err == nil {
		t.Fatal("OperationDirectRead accepted invalid timeMin")
	}
}

func TestFreeBusyOperationDirectReadFixtureOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/calendar/v3/freeBusy" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["timeMin"] != "2020-01-01T00:00:00Z" || body["timeMax"] != "2020-01-02T00:00:00Z" {
			t.Fatalf("unexpected time bounds: %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"kind": "calendar#freeBusy", "calendars": map[string]any{"primary": map[string]any{"busy": []any{}}}})
	}))
	defer srv.Close()

	b, err := engine.Load(defs.FS, "google-calendar")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	b.HTTP.URL = srv.URL + "/calendar/v3"
	conn := engine.New(b, Hooks{})
	direct, err := conn.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
		Operation: "google-calendar.freebusy.query",
		Config:    connectors.RuntimeConfig{ProjectDir: "__polymetrics_conformance_fixture__"},
		Body: map[string]any{
			"timeMin": "2020-01-01T00:00:00Z",
			"timeMax": "2020-01-02T00:00:00Z",
			"items":   []any{map[string]any{"id": "primary"}},
		},
		MaxBytes: 4096,
	})
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	body, err := json.Marshal(direct.Body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	if !strings.Contains(string(body), "calendar#freeBusy") {
		t.Fatalf("body = %s", body)
	}
}
