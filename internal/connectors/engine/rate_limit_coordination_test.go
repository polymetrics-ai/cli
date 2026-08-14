package engine

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
		t.Fatal("endpoint require_shared hook request did not refuse")
	} else {
		var unavailable *coordination.SharedRateLimitUnavailableError
		if !errors.As(err, &unavailable) {
			t.Fatalf("endpoint require_shared hook error = %T %v, want typed shared-coordinator refusal", err, err)
		}
	}
	if requests != 0 {
		t.Fatalf("endpoint require_shared hook request reached the provider %d times", requests)
	}
	if _, err := runtime.Requester.Do(context.Background(), http.MethodGet, "/unmatched", nil, nil); err != nil {
		t.Fatalf("unmatched hook request: %v", err)
	}
	if requests != 1 {
		t.Fatalf("unmatched hook request count = %d, want 1", requests)
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
			w.Header().Set("Location", "/api/repos/a%2Fb")
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
	bundle.RateLimits.Policies[0].Coordination = connsdk.RateLimitCoordinationRequireShared
	bundle.RateLimits.Policies[0].Selector = connsdk.RateLimitSelector{Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodGet, Path: "/repos/{id}"}}}
	restore := replaceSharedRateLimitRegistryForTest(nil)
	t.Cleanup(restore)

	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	for _, path := range []string{"/repos/a%2Fb", "/start"} {
		_, err := runtime.Requester.Do(context.Background(), http.MethodGet, path, nil, nil)
		var unavailable *coordination.SharedRateLimitUnavailableError
		if !errors.As(err, &unavailable) {
			t.Fatalf("escaped endpoint request %q error = %T %v, want typed shared-coordinator refusal", path, err, err)
		}
	}
	if startRequests != 1 {
		t.Fatalf("redirect source request count = %d, want 1", startRequests)
	}
	if providerRequests != 0 {
		t.Fatalf("escaped endpoint request reached the provider %d times", providerRequests)
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
				return Bundle{Name: "acme", HTTP: HTTPBase{URL: baseURL}, Operations: []OperationSpec{{ID: "acme.lookup", Kind: "rest_read", Summary: "lookup", Risk: "low", Approval: "none", OutputPolicy: "json_redacted", REST: &RESTOperationSpec{Method: http.MethodGet, Path: "/lookup", MaxBytes: 1024}}}, Surface: &APISurface{Endpoints: []SurfaceEndpoint{{Method: http.MethodGet, Path: "/lookup", Operation: &SurfaceOperation{Model: "direct_read"}}}}}
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
				_, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{Operation: "acme.widgets.query", Config: cfg, Body: map[string]any{"id": "widget-1"}}, nil)
				return err
			},
		},
		{
			name:   "binary download",
			method: http.MethodGet,
			path:   "/file",
			bundle: func(baseURL string) Bundle {
				return Bundle{Name: "acme", HTTP: HTTPBase{URL: baseURL}, Operations: []OperationSpec{{ID: "acme.file", Kind: "binary_download", Summary: "file", Risk: "low", Approval: "none", Binary: &BinaryOperationSpec{Method: http.MethodGet, Path: "/file", MaxBytes: 1024}}}, Surface: &APISurface{Endpoints: []SurfaceEndpoint{{Method: http.MethodGet, Path: "/file", Operation: &SurfaceOperation{}}}}}
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
	if _, err := runtime.Requester.Do(context.Background(), http.MethodGet, "/items", nil, nil); err != nil {
		t.Fatalf("process-local request: %v", err)
	}
	if got, want := requests["/items"], 1; got != want {
		t.Fatalf("process-local request count = %d, want %d", got, want)
	}
	_, err = runtime.Requester.Do(context.Background(), http.MethodGet, "/admin", nil, nil)
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

func TestUndeclaredRateLimitPolicyHasNoCoordinationProvenance(t *testing.T) {
	connector := New(Bundle{Name: "undeclared", HTTP: HTTPBase{URL: "https://example.test"}}, nil)
	if _, ok := connectors.RateLimitCoordinationOf(connector); ok {
		t.Fatal("undeclared rate limits exposed process-local coordination provenance")
	}
}
