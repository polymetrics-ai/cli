package engine

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/coordination"
)

func TestRequireSharedRateLimitPolicyRefusesWithoutCoordinator(t *testing.T) {
	bundle := withAllRateLimit(Bundle{Name: "shared-required", HTTP: HTTPBase{URL: "https://example.test"}})
	bundle.RateLimits.Policies[0].Coordination = connsdk.RateLimitCoordinationRequireShared
	restore := replaceSharedRateLimitRegistryForTest(nil)
	t.Cleanup(restore)

	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	_, err = runtime.RequesterFor(http.MethodGet, "/widgets")
	var unavailable *coordination.SharedRateLimitUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("RequesterFor error = %T %v, want typed shared-coordinator refusal", err, err)
	}
	if got, want := unavailable.Reason, coordination.SharedRateLimitCoordinatorNotConfigured; got != want {
		t.Fatalf("shared refusal reason = %q, want %q", got, want)
	}
}

func TestLocalRateLimitPolicyNeverInheritsSharedRequirement(t *testing.T) {
	bundle := withAllRateLimit(Bundle{Name: "local-default", HTTP: HTTPBase{URL: "https://example.test"}})
	restore := replaceSharedRateLimitRegistryForTest(nil)
	t.Cleanup(restore)

	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	requester, err := runtime.RequesterFor(http.MethodGet, "/widgets")
	if err != nil {
		t.Fatalf("local policy inherited a shared requirement: %v", err)
	}
	if requester.Admission == nil {
		t.Fatal("local policy did not retain local rate-limit admission")
	}
	if got, want := RateLimitCoordinationOf(New(bundle, nil)).Mode, connectors.RateLimitCoordinationProcessLocal; got != want {
		t.Fatalf("local policy inspect mode = %q, want %q", got, want)
	}
}
