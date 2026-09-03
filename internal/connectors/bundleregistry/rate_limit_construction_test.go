package bundleregistry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
)

func TestLazyConstructionSharesGitHubRateAdmission(t *testing.T) {
	var sends atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		sends.Add(1)
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"rate limited"}`))
	}))
	t.Cleanup(server.Close)

	firstRegistry, err := NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	first, ok := firstRegistry.Get("github")
	if !ok {
		t.Fatal("first production registry did not select github")
	}
	secondRegistry, err := NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	second, ok := secondRegistry.Get("github")
	if !ok {
		t.Fatal("second production registry did not select github")
	}
	coordination, ok := connectors.RateLimitCoordinationOf(first)
	if !ok || coordination.Mode != connectors.RateLimitCoordinationProcessLocal {
		t.Fatalf("selected github rate coordination = %#v, %t; want process-local declared coordination", coordination, ok)
	}

	check := func(connector connectors.Connector, ip string) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
		defer cancel()
		_ = connector.Check(ctx, githubRateConstructionConfig(t, server.URL, ip))
	}
	check(first, "198.51.100.10")
	if got := sends.Load(); got != 1 {
		t.Fatalf("first selected GitHub check sent %d requests, want 1", got)
	}
	check(second, "198.51.100.10")
	if got := sends.Load(); got != 1 {
		t.Fatalf("same-scope selected GitHub construction sent %d requests after 429/reset, want 1", got)
	}
	check(second, "198.51.100.11")
	if got := sends.Load(); got != 2 {
		t.Fatalf("different GitHub rate scope sent %d requests, want 2", got)
	}
}

func githubRateConstructionConfig(t *testing.T, baseURL, ip string) connectors.RuntimeConfig {
	t.Helper()
	identity, err := connectors.NewCoordinationIdentity([]byte("cp06-rate-construction-salt"), connectors.CredentialBinding{
		BindingID:      "cp06-rate-construction-binding",
		ProviderFamily: "github",
		AuthProfile:    "public",
	})
	if err != nil {
		t.Fatal(err)
	}
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"auth_type":     "public",
			"public_access": "true",
			"base_url":      baseURL,
			"owner":         "octocat",
			"repo":          "example",
			"rate_limit_ip": ip,
		},
		CoordinationIdentity: identity,
	}
}
