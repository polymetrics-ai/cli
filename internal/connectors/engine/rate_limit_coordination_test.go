package engine

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	var refusal *connsdk.RateBudgetRefusalError
	if !errors.As(err, &refusal) || refusal.Code != connsdk.RateBudgetRefusalSharedCoordinatorUnavailable {
		t.Fatalf("RequesterFor error = %T %v, want RateBudgetRefusalError/shared_coordinator_unavailable", err, err)
	}
	var unavailable *coordination.SharedRateLimitUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("RequesterFor error = %T %v, want typed shared-coordinator refusal", err, err)
	}
	if got, want := unavailable.Reason, coordination.SharedRateLimitCoordinatorNotConfigured; got != want {
		t.Fatalf("shared refusal reason = %q, want %q", got, want)
	}
}

func TestRequireSharedRateLimitPolicyPreservesCanceledContext(t *testing.T) {
	bundle := withAllRateLimit(Bundle{Name: "shared-required-canceled", HTTP: HTTPBase{URL: "https://example.test"}})
	bundle.RateLimits.Policies[0].Coordination = connsdk.RateLimitCoordinationRequireShared
	restore := replaceSharedRateLimitRegistryForTest(nil)
	t.Cleanup(restore)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := newRuntime(ctx, bundle, rateLimitTestConfig(t), nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled require_shared runtime error = %v, want context.Canceled", err)
	}
	var unavailable *coordination.SharedRateLimitUnavailableError
	if errors.As(err, &unavailable) {
		t.Fatalf("cancelled require_shared runtime returned unavailable reason %q", unavailable.Reason)
	}
}

func TestEndpointRequireSharedPolicyGatesHookRequesterAtSend(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	bundle := withAllRateLimit(Bundle{Name: "endpoint-shared-required", HTTP: HTTPBase{URL: server.URL}})
	bundle.RateLimits.Policies[0].Coordination = connsdk.RateLimitCoordinationRequireShared
	bundle.RateLimits.Policies[0].Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodGet, Path: "/hook/{id}"}}}
	restore := replaceSharedRateLimitRegistryForTest(nil)
	t.Cleanup(restore)

	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	if _, err := runtime.Requester.Do(context.Background(), http.MethodGet, "/hook/42", nil, nil); err == nil {
		t.Fatal("default endpoint-policy requester did not refuse an undeclared hook route")
	}
	if requests != 0 {
		t.Fatalf("default endpoint-policy requester reached the provider %d times", requests)
	}
	requester, err := runtime.RequesterFor(http.MethodGet, "/hook/{id}")
	if err != nil {
		t.Fatalf("RequesterFor hook route: %v", err)
	}
	_, err = requester.Do(context.Background(), http.MethodGet, "/hook/42", nil, nil)
	var unavailable *coordination.SharedRateLimitUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("declared hook route error = %T %v, want typed shared-coordinator refusal", err, err)
	}
	if requests != 0 {
		t.Fatalf("declared endpoint require_shared requester reached the provider %d times", requests)
	}
	if _, err := runtime.Requester.Do(context.Background(), http.MethodGet, "/unmatched", nil, nil); err == nil {
		t.Fatal("default endpoint-policy requester did not refuse an unresolved route")
	}
	if requests != 0 {
		t.Fatalf("unresolved default hook request reached the provider %d times", requests)
	}
}

func TestDefaultRequesterRefusesEndpointPolicyWithoutDeclaredRoute(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	bundle := withAllRateLimit(Bundle{Name: "endpoint-default-requester", HTTP: HTTPBase{URL: server.URL}})
	bundle.RateLimits.Policies[0].Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodPost, Path: "/widgets/{id}"}}}
	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}

	_, err = runtime.Requester.Do(context.Background(), http.MethodPost, "/widgets/42", nil, map[string]any{"name": "fixture"})
	if err == nil {
		t.Fatal("default endpoint-policy requester sent without a declared route")
	}
	if requests != 0 {
		t.Fatalf("default endpoint-policy requester sends = %d, want 0", requests)
	}
}

func TestDefaultRequesterDoesNotPartiallyAdmitMixedEndpointPolicies(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	clock := &engineRateLimitClock{now: time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)}
	restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
	t.Cleanup(restore)
	bundle := withAllRateLimit(Bundle{Name: "mixed-default-requester", HTTP: HTTPBase{URL: server.URL}})
	endpointPolicy := bundle.RateLimits.Policies[0]
	endpointPolicy.ID = "widgets-endpoint"
	endpointPolicy.Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodPost, Path: "/widgets/{id}"}}}
	bundle.RateLimits.Policies = append(bundle.RateLimits.Policies, endpointPolicy)
	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}

	if _, err := runtime.Requester.Do(context.Background(), http.MethodPost, "/widgets/42", nil, map[string]any{"name": "fixture"}); err == nil {
		t.Fatal("default requester sent a mixed-policy endpoint without a declared route")
	}
	if requests != 0 {
		t.Fatalf("mixed-policy default requester sends = %d, want 0", requests)
	}

	requester, err := runtime.RequesterFor(http.MethodGet, "/read")
	if err != nil {
		t.Fatalf("RequesterFor read route: %v", err)
	}
	if _, err := requester.Do(context.Background(), http.MethodGet, "/read", nil, nil); err != nil {
		t.Fatalf("declared default-policy request: %v", err)
	}
	if requests != 1 {
		t.Fatalf("declared default-policy sends = %d, want 1", requests)
	}
	if len(clock.waits) != 0 {
		t.Fatalf("default requester consumed a policy before refusal; waits = %v, want none", clock.waits)
	}
}

func TestEndpointRequireSharedPolicyGatesInterpolatedRequesterPathAtSend(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	bundle := withAllRateLimit(Bundle{Name: "interpolated-endpoint-shared-required", HTTP: HTTPBase{URL: server.URL}})
	bundle.RateLimits.Policies[0].Coordination = connsdk.RateLimitCoordinationRequireShared
	bundle.RateLimits.Policies[0].Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodGet, Path: "/widgets/special"}}}
	restore := replaceSharedRateLimitRegistryForTest(nil)
	t.Cleanup(restore)

	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	requester, err := runtime.RequesterFor(http.MethodGet, "/widgets/{id}")
	if err != nil {
		t.Fatalf("RequesterFor: %v", err)
	}
	if requester.RouteRateLimits == nil {
		t.Fatal("RequesterFor did not retain endpoint rate-limit resolution")
	}
	_, err = requester.Do(context.Background(), http.MethodGet, "/widgets/special", nil, nil)
	var unavailable *coordination.SharedRateLimitUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("interpolated endpoint request error = %T %v, want typed shared-coordinator refusal", err, err)
	}
	if got, want := unavailable.Reason, coordination.SharedRateLimitCoordinatorNotConfigured; got != want {
		t.Fatalf("interpolated endpoint shared refusal reason = %q, want %q", got, want)
	}
	if requests != 0 {
		t.Fatalf("interpolated endpoint request reached the provider %d times", requests)
	}
}

func TestEndpointRequireSharedPolicyGatesEscapedAndBasePrefixedPathsAtSend(t *testing.T) {
	startRequests := 0
	providerRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case "/api/start":
			startRequests++
			w.Header().Set("Location", "/api/unbound")
			w.WriteHeader(http.StatusFound)
		case "/api/repos/a%2Fb":
			providerRequests++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	bundle := withAllRateLimit(Bundle{Name: "escaped-endpoint-shared-required", HTTP: HTTPBase{URL: server.URL + "/api"}})
	shared := bundle.RateLimits.Policies[0]
	shared.Coordination = connsdk.RateLimitCoordinationRequireShared
	shared.Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodGet, Path: "/repos/{id}"}}}
	local := shared
	local.ID = "start-local"
	local.Coordination = ""
	local.Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodGet, Path: "/start"}}}
	bundle.RateLimits.Policies = []connsdk.RateLimitPolicy{local, shared}
	restore := replaceSharedRateLimitRegistryForTest(nil)
	t.Cleanup(restore)

	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	for _, tt := range []struct {
		name         string
		declaredPath string
		path         string
		reason       coordination.SharedRateLimitUnavailableReason
	}{
		{name: "escaped", declaredPath: "/repos/{id}", path: "/repos/a%2Fb", reason: coordination.SharedRateLimitCoordinatorNotConfigured},
		{name: "base prefixed", declaredPath: "/unbound", path: "/api/unbound", reason: coordination.SharedRateLimitRouteUnresolved},
	} {
		t.Run(tt.name, func(t *testing.T) {
			requester, err := runtime.RequesterFor(http.MethodGet, tt.declaredPath)
			if err != nil {
				t.Fatalf("RequesterFor endpoint route: %v", err)
			}
			_, err = requester.Do(context.Background(), http.MethodGet, tt.path, nil, nil)
			var unavailable *coordination.SharedRateLimitUnavailableError
			if !errors.As(err, &unavailable) {
				t.Fatalf("endpoint request %q error = %T %v, want typed shared-coordinator refusal", tt.path, err, err)
			}
			if got, want := unavailable.Reason, tt.reason; got != want {
				t.Fatalf("endpoint request %q refusal reason = %q, want %q", tt.path, got, want)
			}
		})
	}
	requester, err := runtime.RequesterFor(http.MethodGet, "/start")
	if err != nil {
		t.Fatalf("RequesterFor redirect source route: %v", err)
	}
	_, err = requester.Do(context.Background(), http.MethodGet, "/start", nil, nil)
	var unavailable *coordination.SharedRateLimitUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("redirected endpoint request error = %T %v, want typed shared-coordinator refusal", err, err)
	}
	if got, want := unavailable.Reason, coordination.SharedRateLimitRouteUnresolved; got != want {
		t.Fatalf("redirected endpoint refusal reason = %q, want %q", got, want)
	}
	if startRequests != 1 {
		t.Fatalf("redirect source request count = %d, want 1", startRequests)
	}
	if providerRequests != 0 {
		t.Fatalf("escaped endpoint request reached the provider %d times", providerRequests)
	}
}

func TestEndpointSharedRateLimitAdmissionUsesRedirectDestination(t *testing.T) {
	var startRequests, destinationRequests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/start":
			startRequests++
			http.Redirect(w, r, "/repos/widget", http.StatusFound)
		case "/repos/widget":
			destinationRequests++
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	bundle := withAllRateLimit(Bundle{Name: "redirect-destination-shared", HTTP: HTTPBase{URL: server.URL}})
	local := bundle.RateLimits.Policies[0]
	local.ID = "start-local"
	local.Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodGet, Path: "/start"}}}
	shared := local
	shared.ID = "repos-shared"
	shared.Coordination = connsdk.RateLimitCoordinationRequireShared
	shared.Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodGet, Path: "/repos/{id}"}}}
	bundle.RateLimits.Policies = []connsdk.RateLimitPolicy{local, shared}
	restore := replaceSharedRateLimitRegistryForTest(nil)
	t.Cleanup(restore)

	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	requester, err := runtime.RequesterFor(http.MethodGet, "/start")
	if err != nil {
		t.Fatalf("RequesterFor redirect source: %v", err)
	}
	_, err = requester.Do(context.Background(), http.MethodGet, "/start", nil, nil)
	var unavailable *coordination.SharedRateLimitUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("redirect request error = %T %v, want destination shared-coordinator refusal", err, err)
	}
	if got, want := unavailable.Reason, coordination.SharedRateLimitCoordinatorNotConfigured; got != want {
		t.Fatalf("redirect destination refusal reason = %q, want %q", got, want)
	}
	if got, want := startRequests, 1; got != want {
		t.Fatalf("redirect source requests = %d, want %d", got, want)
	}
	if got, want := destinationRequests, 0; got != want {
		t.Fatalf("shared redirect destination requests = %d, want %d", got, want)
	}
}

func TestEndpointLocalRateLimitAdmissionAllowsRedirectDestination(t *testing.T) {
	var destinationRequests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/start":
			http.Redirect(w, r, "/repos/widget", http.StatusFound)
		case "/repos/widget":
			destinationRequests++
			_, _ = w.Write([]byte(`{"route":"destination"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	bundle := withAllRateLimit(Bundle{Name: "redirect-destination-local", HTTP: HTTPBase{URL: server.URL}})
	start := bundle.RateLimits.Policies[0]
	start.ID = "start-local"
	start.Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodGet, Path: "/start"}}}
	destination := start
	destination.ID = "repos-local"
	destination.Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodGet, Path: "/repos/{id}"}}}
	bundle.RateLimits.Policies = []connsdk.RateLimitPolicy{start, destination}
	restore := replaceSharedRateLimitRegistryForTest(nil)
	t.Cleanup(restore)

	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	requester, err := runtime.RequesterFor(http.MethodGet, "/start")
	if err != nil {
		t.Fatalf("RequesterFor redirect source: %v", err)
	}
	response, err := requester.Do(context.Background(), http.MethodGet, "/start", nil, nil)
	if err != nil {
		t.Fatalf("local redirect request: %v", err)
	}
	if got, want := string(response.Body), `{"route":"destination"}`; got != want {
		t.Fatalf("local redirect response = %q, want %q", got, want)
	}
	if got, want := destinationRequests, 1; got != want {
		t.Fatalf("local redirect destination requests = %d, want %d", got, want)
	}
}

func TestEndpointSharedRateLimitAdmissionCanonicalizesBasePrefixedRedirectDestination(t *testing.T) {
	var destinationRequests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case "/api/start":
			http.Redirect(w, r, "/api/repos/widget", http.StatusFound)
		case "/api/repos/widget":
			destinationRequests++
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	bundle := withAllRateLimit(Bundle{Name: "base-prefixed-redirect-shared", HTTP: HTTPBase{URL: server.URL + "/api"}})
	local := bundle.RateLimits.Policies[0]
	local.ID = "start-local"
	local.Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodGet, Path: "/start"}}}
	shared := local
	shared.ID = "repos-shared"
	shared.Coordination = connsdk.RateLimitCoordinationRequireShared
	shared.Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodGet, Path: "/repos/{id}"}}}
	bundle.RateLimits.Policies = []connsdk.RateLimitPolicy{local, shared}
	restore := replaceSharedRateLimitRegistryForTest(nil)
	t.Cleanup(restore)

	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	requester, err := runtime.RequesterFor(http.MethodGet, "/start")
	if err != nil {
		t.Fatalf("RequesterFor redirect source: %v", err)
	}
	_, err = requester.Do(context.Background(), http.MethodGet, "/start", nil, nil)
	var unavailable *coordination.SharedRateLimitUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("base-prefixed redirect request error = %T %v, want destination shared-coordinator refusal", err, err)
	}
	if got, want := unavailable.Reason, coordination.SharedRateLimitCoordinatorNotConfigured; got != want {
		t.Fatalf("base-prefixed redirect destination refusal reason = %q, want %q", got, want)
	}
	if got, want := destinationRequests, 0; got != want {
		t.Fatalf("base-prefixed shared redirect destination requests = %d, want %d", got, want)
	}
}

func TestResponseFormatErrorPreservesOnlySafeErrorIdentities(t *testing.T) {
	formatted := formatResponseError("safe formatted response error", &connsdk.HTTPError{Status: http.StatusBadRequest, URL: "https://example.test/graphql", Body: "access_token=must-not-be-exposed"})
	var httpErr *connsdk.HTTPError
	if errors.As(formatted, &httpErr) {
		t.Fatalf("formatted response exposed raw HTTP error body %q", httpErr.Body)
	}
	formatted = formatResponseError("safe formatted response error", &connsdk.HTTPError{Status: http.StatusUnauthorized, URL: "https://example.test/graphql", Body: "access_token=must-not-be-exposed"})
	var credentialRejected *connsdk.CredentialRejectedError
	if !errors.As(formatted, &credentialRejected) {
		t.Fatalf("formatted provider 401 did not preserve safe credential rejection: %v", formatted)
	}
	if credentialRejected.Error() != "provider rejected the credential" {
		t.Fatalf("formatted provider 401 credential error = %q", credentialRejected)
	}

	shared := &coordination.SharedRateLimitUnavailableError{Component: "dragonfly", Reason: coordination.SharedRateLimitCoordinatorNotConfigured}
	formatted = formatResponseError("safe formatted response error", shared)
	var unavailable *coordination.SharedRateLimitUnavailableError
	if !errors.As(formatted, &unavailable) {
		t.Fatalf("formatted response did not preserve typed shared refusal: %v", formatted)
	}
	if unavailable != shared {
		t.Fatalf("formatted response unavailable error = %p, want %p", unavailable, shared)
	}

	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		formatted = formatResponseError("safe formatted response error", cause)
		if !errors.Is(formatted, cause) {
			t.Fatalf("formatted response error = %v, want %v identity", formatted, cause)
		}
	}
}

func TestRateLimitDeclarationRejectsPointDefaultCostAboveCapacity(t *testing.T) {
	for _, tt := range []struct {
		name   string
		budget int
	}{
		{name: "window", budget: 0},
		{name: "bucket", budget: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			bundle := loadRateLimitFixture(t, "paced")
			cost := 3.0
			bundle.RateLimits.Policies[0].Budgets[tt.budget].Cost.DefaultCost = &cost
			if err := validateRateLimits(*bundle.RateLimits, bundle.Spec); err == nil {
				t.Fatal("rate-limit declaration accepted a default point cost above capacity")
			}
		})
	}
}

func TestEndpointRequireSharedErrorSurvivesOperationFormatting(t *testing.T) {
	for _, tt := range []struct {
		name   string
		method string
		path   string
		bundle func(string) Bundle
		run    func(Bundle, connectors.RuntimeConfig) error
	}{
		{
			name:   "direct read",
			method: http.MethodGet,
			path:   "/items",
			bundle: func(baseURL string) Bundle { return directReadBundle(baseURL, http.MethodGet, "/items") },
			run: func(bundle Bundle, cfg connectors.RuntimeConfig) error {
				_, err := DirectRead(context.Background(), bundle, connectors.DirectReadRequest{Method: http.MethodGet, Path: "/items", Config: cfg, OutputPolicy: "json_redacted"}, nil)
				return err
			},
		},
		{
			name:   "operation direct read",
			method: http.MethodGet,
			path:   "/lookup",
			bundle: func(baseURL string) Bundle {
				return Bundle{Name: "acme", HTTP: HTTPBase{URL: baseURL}, Operations: []OperationSpec{{ID: "acme.lookup", Kind: "rest_read", Summary: "lookup", Risk: "low", Approval: "none", OutputPolicy: "json_redacted", REST: &RESTOperationSpec{Method: http.MethodGet, Path: "/lookup", MaxBytes: 1024}}}}
			},
			run: func(bundle Bundle, cfg connectors.RuntimeConfig) error {
				_, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{Operation: "acme.lookup", Config: cfg}, nil)
				return err
			},
		},
		{
			name:   "graphql direct read",
			method: http.MethodPost,
			path:   "/graphql",
			bundle: func(baseURL string) Bundle { return graphQLOperationBundle(baseURL, "graphql_query") },
			run: func(bundle Bundle, cfg connectors.RuntimeConfig) error {
				_, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{Operation: "acme.widgets.query", Config: cfg, Body: map[string]any{"id": "widget-1", "first": 1}}, nil)
				return err
			},
		},
		{
			name:   "binary download",
			method: http.MethodGet,
			path:   "/file",
			bundle: func(baseURL string) Bundle {
				return Bundle{Name: "acme", HTTP: HTTPBase{URL: baseURL}, Operations: []OperationSpec{{ID: "acme.file", Kind: "binary_download", Summary: "file", Risk: "low", Approval: "none", Binary: &BinaryOperationSpec{Method: http.MethodGet, Path: "/file", MaxBytes: 1024, ContentTypes: []string{"application/octet-stream"}, Response: &OperationResponseSpec{SuccessStatuses: []string{"200"}}}}}}
			},
			run: func(bundle Bundle, cfg connectors.RuntimeConfig) error {
				_, err := OperationBinaryDownload(context.Background(), bundle, BinaryDownloadRequest{Operation: "acme.file", Config: cfg, DestRoot: t.TempDir()}, nil)
				return err
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				requests++
			}))
			t.Cleanup(server.Close)
			bundle := withAllRateLimit(tt.bundle(server.URL))
			policy := &bundle.RateLimits.Policies[0]
			policy.Coordination = connsdk.RateLimitCoordinationRequireShared
			policy.Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: tt.method, Path: tt.path}}}
			restore := replaceSharedRateLimitRegistryForTest(nil)
			t.Cleanup(restore)
			cfg := rateLimitTestConfig(t)
			cfg.Config["base_url"] = server.URL
			err := tt.run(bundle, cfg)
			var unavailable *coordination.SharedRateLimitUnavailableError
			if !errors.As(err, &unavailable) {
				t.Fatalf("formatted require_shared error = %T %v, want typed shared-coordinator refusal", err, err)
			}
			if got, want := unavailable.Reason, coordination.SharedRateLimitCoordinatorNotConfigured; got != want {
				t.Fatalf("formatted require_shared reason = %q, want %q", got, want)
			}
			if requests != 0 {
				t.Fatalf("formatted require_shared request reached the provider %d times", requests)
			}
		})
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

func TestMixedRateLimitPoliciesExposePolicyScopedCoordination(t *testing.T) {
	requests := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests[r.URL.Path]++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	bundle := withAllRateLimit(Bundle{Name: "mixed-rate-limit-coordination", HTTP: HTTPBase{URL: server.URL}})
	local := bundle.RateLimits.Policies[0]
	local.ID = "items-local"
	local.Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodGet, Path: "/items"}}}
	shared := local
	shared.ID = "admin-shared"
	shared.Coordination = connsdk.RateLimitCoordinationRequireShared
	shared.Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodGet, Path: "/admin"}}}
	bundle.RateLimits.Policies = []connsdk.RateLimitPolicy{local, shared}
	restore := replaceSharedRateLimitRegistryForTest(nil)
	t.Cleanup(restore)

	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	requester, err := runtime.RequesterFor(http.MethodGet, "/items")
	if err != nil {
		t.Fatalf("RequesterFor local route: %v", err)
	}
	if _, err := requester.Do(context.Background(), http.MethodGet, "/items", nil, nil); err != nil {
		t.Fatalf("process-local request: %v", err)
	}
	if got, want := requests["/items"], 1; got != want {
		t.Fatalf("process-local request count = %d, want %d", got, want)
	}
	requester, err = runtime.RequesterFor(http.MethodGet, "/admin")
	if err != nil {
		t.Fatalf("RequesterFor shared route: %v", err)
	}
	_, err = requester.Do(context.Background(), http.MethodGet, "/admin", nil, nil)
	var unavailable *coordination.SharedRateLimitUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("require_shared request error = %T %v, want typed shared-coordinator refusal", err, err)
	}
	if got := requests["/admin"]; got != 0 {
		t.Fatalf("require_shared request reached the provider %d times", got)
	}
	status, ok := connectors.RateLimitCoordinationOf(New(bundle, nil))
	if !ok {
		t.Fatal("mixed policies did not expose rate-limit coordination status")
	}
	if got, want := status.Mode, connectors.RateLimitCoordinationMixed; got != want {
		t.Fatalf("mixed policy inspect mode = %q, want %q", got, want)
	}
	if got, want := status.Message, "Rate-limit coordination is policy-scoped: process-local policies protect this pm process only and are not shared across processes; require_shared policies refuse before sending when the optional coordinator is unavailable."; got != want {
		t.Fatalf("mixed policy inspect message = %q, want %q", got, want)
	}
}

func TestSelectorTierCannotWeakenRequireSharedInspection(t *testing.T) {
	bundle := withAllRateLimit(Bundle{Name: "tiered-shared", HTTP: HTTPBase{URL: "https://example.test"}})
	shared := bundle.RateLimits.Policies[0]
	shared.ID = "tiered-shared"
	shared.Coordination = connsdk.RateLimitCoordinationRequireShared
	shared.Selector = connsdk.RateLimitSelector{Tiers: []string{"restricted"}}
	bundle.RateLimits.Policies = append(bundle.RateLimits.Policies, shared)

	status, ok := connectors.RateLimitCoordinationOf(New(bundle, nil))
	if !ok {
		t.Fatal("tiered shared policy did not expose rate-limit coordination status")
	}
	if got, want := status.Mode, connectors.RateLimitCoordinationMixed; got != want {
		t.Fatalf("tiered shared policy inspect mode = %q, want %q", got, want)
	}
	if !strings.Contains(status.Message, "require_shared policies refuse before sending") {
		t.Fatalf("tiered shared policy inspect message = %q, want generic require_shared refusal", status.Message)
	}
}

func TestUndeclaredRateLimitPolicyHasNoCoordinationProvenance(t *testing.T) {
	connector := New(Bundle{Name: "undeclared", HTTP: HTTPBase{URL: "https://example.test"}}, nil)
	if _, ok := connectors.RateLimitCoordinationOf(connector); ok {
		t.Fatal("undeclared rate limits exposed process-local coordination provenance")
	}
}
