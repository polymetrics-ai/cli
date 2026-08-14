package engine

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/coordination"
)

func TestRateLimitCoordinationSchemaAllowsOnlyExplicitRequireShared(t *testing.T) {
	base := map[string]any{
		"schema_version": 1,
		"state":          "declared",
		"policies": []any{map[string]any{
			"id":       "paced",
			"source":   map[string]any{"url": "https://docs.example.test/rate-limits", "retrieved_at": "2026-08-14"},
			"selector": map[string]any{"all": true},
			"scope":    map[string]any{"subject_kind": "account", "subject_config": "account_id"},
			"budgets":  []any{map[string]any{"model": "fixed_window", "dimension": "sustained", "unit": "requests", "limit": 1, "window_seconds": 60}},
		}},
	}
	for _, tt := range []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "absent defaults process local", value: "", wantErr: false},
		{name: "explicit require shared", value: "require_shared", wantErr: false},
		{name: "process local cannot be declared as shared selector", value: "process_local", wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			policy := base["policies"].([]any)[0].(map[string]any)
			if tt.value == "" {
				delete(policy, "coordination")
			} else {
				policy["coordination"] = tt.value
			}
			raw, err := json.Marshal(base)
			if err != nil {
				t.Fatalf("marshal test declaration: %v", err)
			}
			err = metaSchemas.rateLimits.Validate(mustDecodeAny(raw))
			if (err != nil) != tt.wantErr {
				t.Fatalf("schema error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}

func TestRequireSharedRateLimitPolicyRefusesWithoutCoordinator(t *testing.T) {
	bundle := withAllRateLimit(Bundle{Name: "shared-required", HTTP: HTTPBase{URL: "https://example.test"}})
	bundle.RateLimits.Policies[0].Coordination = connsdk.RateLimitCoordinationRequireShared
	restore := replaceSharedRateLimitRegistryForTest(nil)
	t.Cleanup(restore)

	_, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
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
	status, ok := connectors.RateLimitCoordinationOf(New(bundle, nil))
	if !ok {
		t.Fatal("engine connector did not expose rate-limit coordination status")
	}
	if got, want := status.Mode, connectors.RateLimitCoordinationProcessLocal; got != want {
		t.Fatalf("local policy inspect mode = %q, want %q", got, want)
	}
}
