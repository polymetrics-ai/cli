package hoorayhr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestHooksRegistered(t *testing.T) {
	h := engine.HooksFor("hoorayhr")
	if h == nil {
		t.Fatal("registered hooks = nil")
	}
	if h.ConnectorName() != "hoorayhr" {
		t.Fatalf("ConnectorName() = %q", h.ConnectorName())
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("hooks do not implement AuthHook")
	}
}

func TestAuthenticatorBuildsFromConfig(t *testing.T) {
	auth, err := (&Hooks{}).Authenticator(context.Background(), connectors.RuntimeConfig{
		Config:  map[string]string{"hoorayhrusername": "user@example.test"},
		Secrets: map[string]string{"hoorayhrpassword": "password"},
	}, engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	if auth == nil {
		t.Fatal("Authenticator returned nil")
	}
}

func TestAuthenticatorRequiresUsernameAndPassword(t *testing.T) {
	if _, err := (&Hooks{}).Authenticator(context.Background(), connectors.RuntimeConfig{}, engine.AuthSpec{}); err == nil {
		t.Fatal("Authenticator without config/secrets: want error, got nil")
	}
}

func TestAuthenticatorReportsLoginRateLimitActivity(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost || req.URL.Path != "/authentication" {
			t.Errorf("login request = %s %s, want POST /authentication", req.Method, req.URL.Path)
		}
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accessToken":"session-token-must-not-escape"}`))
	}))
	t.Cleanup(server.Close)

	report := connectors.NewRateLimitReport()
	report.Declare("hoorayhr", connectors.RateLimitDeclarationDeclared)
	auth, err := (&Hooks{Client: server.Client()}).Authenticator(context.Background(), connectors.RuntimeConfig{
		Config:          map[string]string{"base_url": server.URL, "hoorayhrusername": "user@example.test"},
		Secrets:         map[string]string{"hoorayhrpassword": "password-must-not-escape"},
		RateLimitReport: report,
	}, engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	login, ok := auth.(*sessionTokenAuth)
	if !ok {
		t.Fatalf("authenticator type = %T, want *sessionTokenAuth", auth)
	}
	now := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)
	login.login.MaxRetries = 1
	login.login.Now = func() time.Time { return now }
	login.login.Sleep = func(context.Context, time.Duration) error { return nil }
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://data.example.test/records", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext: %v", err)
	}
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	summary := report.Snapshot()
	if len(summary.Connectors) != 1 {
		t.Fatalf("rate-limit connector count = %d, want 1", len(summary.Connectors))
	}
	connector := summary.Connectors[0]
	if connector.Declaration != connectors.RateLimitDeclarationDeclared || connector.Provider429Observed != 1 || connector.Provider429Honored != 1 || connector.ProviderWaitMS != 1000 || connector.RequestCount != 2 {
		t.Fatalf("login rate-limit activity = %+v", connector)
	}
	encoded, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal rate-limit activity: %v", err)
	}
	for _, forbidden := range []string{"user@example.test", "password-must-not-escape", "session-token-must-not-escape", server.URL} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("rate-limit activity leaked %q", forbidden)
		}
	}
}
