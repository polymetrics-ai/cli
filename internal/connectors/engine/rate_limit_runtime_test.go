package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/coordination"
)

type engineRateLimitClock struct {
	now   time.Time
	waits []time.Duration
}

func (c *engineRateLimitClock) Now() time.Time { return c.now }

func (c *engineRateLimitClock) Sleep(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.waits = append(c.waits, d)
	c.now = c.now.Add(d)
	return nil
}

func rateLimitTestConfig(t *testing.T) connectors.RuntimeConfig {
	t.Helper()
	return rateLimitTestConfigForBinding(t, "binding-test-001", "")
}

func rateLimitTestConfigForBinding(t *testing.T, bindingID, revision string) connectors.RuntimeConfig {
	t.Helper()
	identity, err := connectors.NewCoordinationIdentity([]byte("rate-limit-test-salt"), connectors.CredentialBinding{
		BindingID:      bindingID,
		ProviderFamily: "test-provider",
		AuthProfile:    "test-profile",
	})
	if err != nil {
		t.Fatalf("NewCoordinationIdentity: %v", err)
	}
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   "https://example.test",
			"account_id": "account-test-001",
			"tier":       "pro",
			"auth_type":  "oauth",
		},
		CredentialRevision:   revision,
		CoordinationIdentity: identity,
	}
}

func loadRateLimitFixture(t *testing.T, name string) Bundle {
	t.Helper()
	fixtureFS := os.DirFS("testdata/rate-limit-enforcement")
	bundle, err := Load(fixtureFS, name)
	if err != nil {
		t.Fatalf("Load(%s): %v", name, err)
	}
	return bundle
}

func TestDeclaredRateLimitFixtureSelectsEndpointTierAndAuth(t *testing.T) {
	bundle := loadRateLimitFixture(t, "paced")
	clock := &engineRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
	t.Cleanup(restore)

	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	matched, err := runtime.RequesterFor(http.MethodGet, "/widgets")
	if err != nil {
		t.Fatalf("RequesterFor widgets: %v", err)
	}
	if matched.Admission != nil || matched.Observer != nil || matched.RouteRateLimits == nil {
		t.Fatalf("matched requester rate limit = %+v, want path-aware admission/observation", matched)
	}
	costHeader, err := matched.RouteRateLimits.AdmitRoute(context.Background(), connsdk.RateLimitRoute{Method: http.MethodGet, Path: "/widgets", Attempt: 1})
	if err != nil {
		t.Fatalf("matched route admission: %v", err)
	}
	if got, want := costHeader, "X-Actual-Cost"; got != want {
		t.Fatalf("matched route cost header = %q, want %q", got, want)
	}
	unmatched, err := runtime.RequesterFor(http.MethodGet, "/check")
	if err != nil {
		t.Fatalf("RequesterFor check: %v", err)
	}
	if unmatched.Admission != nil || unmatched.Observer != nil || unmatched.RouteRateLimits == nil {
		t.Fatal("endpoint-mismatched check acquired a rate-limit policy")
	}
	costHeader, err = unmatched.RouteRateLimits.AdmitRoute(context.Background(), connsdk.RateLimitRoute{Method: http.MethodGet, Path: "/check", Attempt: 1})
	if err != nil {
		t.Fatalf("unmatched route admission: %v", err)
	}
	if costHeader != "" {
		t.Fatalf("unmatched route cost header = %q, want empty", costHeader)
	}

	for key, value := range map[string]string{"tier": "free", "auth_type": "api_key"} {
		cfg := rateLimitTestConfig(t)
		cfg.Config[key] = value
		rt, err := newRuntime(context.Background(), bundle, cfg, nil)
		if err != nil {
			t.Fatalf("newRuntime %s: %v", key, err)
		}
		requester, err := rt.RequesterFor(http.MethodGet, "/widgets")
		if err != nil {
			t.Fatalf("RequesterFor %s: %v", key, err)
		}
		if requester.Admission != nil {
			t.Fatalf("%s mismatch acquired an admission", key)
		}
		costHeader, err := requester.RouteRateLimits.AdmitRoute(context.Background(), connsdk.RateLimitRoute{Method: http.MethodGet, Path: "/widgets", Attempt: 1})
		if err != nil {
			t.Fatalf("%s mismatched route admission: %v", key, err)
		}
		if costHeader != "" {
			t.Fatalf("%s mismatch acquired route cost header %q", key, costHeader)
		}
	}
}

func TestUnknownRateLimitFixturePreservesRequesterBehavior(t *testing.T) {
	for _, state := range []connsdk.RateLimitState{connsdk.RateLimitStateUnknown, connsdk.RateLimitStateNotApplicable} {
		t.Run(string(state), func(t *testing.T) {
			bundle := loadRateLimitFixture(t, "unknown")
			bundle.RateLimits.State = state
			runtime, err := newRuntime(context.Background(), bundle, connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}}, nil)
			if err != nil {
				t.Fatalf("newRuntime: %v", err)
			}
			requester, err := runtime.RequesterFor(http.MethodGet, "/widgets")
			if err != nil {
				t.Fatalf("RequesterFor: %v", err)
			}
			if requester != runtime.Requester || requester.Admission != nil || requester.Observer != nil {
				t.Fatalf("%s declaration changed the requester", state)
			}
		})
	}
}

func TestAbsentRateLimitDeclarationPreservesRequesterBehavior(t *testing.T) {
	bundle := Bundle{Name: "absent", HTTP: HTTPBase{URL: "https://example.test"}}
	runtime, err := newRuntime(context.Background(), bundle, connectors.RuntimeConfig{}, nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	requester, err := runtime.RequesterFor(http.MethodGet, "/widgets")
	if err != nil {
		t.Fatalf("RequesterFor: %v", err)
	}
	if requester != runtime.Requester || requester.Admission != nil || requester.Observer != nil {
		t.Fatal("absent declaration changed the requester")
	}
}

func TestRateLimitResolverRefusesUnsupportedScopeKind(t *testing.T) {
	bundle := loadRateLimitFixture(t, "paced")
	bundle.RateLimits.Policies[0].Scope.SubjectKind = connsdk.RateLimitScopeSubjectKind("outside-vocabulary")
	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	requester, err := runtime.RequesterFor(http.MethodGet, "/widgets")
	if err != nil {
		t.Fatalf("RequesterFor: %v", err)
	}
	if _, err := requester.RouteRateLimits.AdmitRoute(context.Background(), connsdk.RateLimitRoute{Method: http.MethodGet, Path: "/widgets", Attempt: 1}); err == nil {
		t.Fatal("route admission accepted an unsupported scope kind")
	}
}

func TestRateLimitResolverRefusesMultipleActualCostHeadersPerPolicy(t *testing.T) {
	bundle := loadRateLimitFixture(t, "paced")
	bundle.RateLimits.Policies[0].Budgets[1].Cost.ResponseHeader = "X-Other-Cost"
	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	requester, err := runtime.RequesterFor(http.MethodGet, "/widgets")
	if err != nil {
		t.Fatalf("RequesterFor: %v", err)
	}
	if _, err := requester.RouteRateLimits.AdmitRoute(context.Background(), connsdk.RateLimitRoute{Method: http.MethodGet, Path: "/widgets", Attempt: 1}); err == nil {
		t.Fatal("route admission accepted multiple actual-cost headers for one policy")
	}
}

func TestRateLimitResolverConsumesDeclaredActualCostHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Actual-Cost", "2")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	t.Cleanup(server.Close)
	bundle := loadRateLimitFixture(t, "paced")
	bundle.HTTP.URL = server.URL
	clock := &engineRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
	t.Cleanup(restore)
	config := rateLimitTestConfig(t)
	config.Config["base_url"] = server.URL
	for attempt := 0; attempt < 2; attempt++ {
		runtime, err := newRuntime(context.Background(), bundle, config, nil)
		if err != nil {
			t.Fatalf("newRuntime: %v", err)
		}
		requester, err := runtime.RequesterFor(http.MethodGet, "/widgets")
		if err != nil {
			t.Fatalf("RequesterFor: %v", err)
		}
		if _, err := requester.Do(context.Background(), http.MethodGet, "/widgets", nil, nil); err != nil {
			t.Fatalf("request %d: %v", attempt, err)
		}
	}
	if len(clock.waits) != 1 || clock.waits[0] != time.Minute {
		t.Fatalf("actual cost did not tighten the declared fixed budget, waits = %v", clock.waits)
	}
}

func TestWholeConnectorPolicyAlsoPacesHookRuntimeRequester(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)
	clock := &engineRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
	t.Cleanup(restore)
	bundle := withAllRateLimit(Bundle{Name: "hooked", HTTP: HTTPBase{URL: server.URL}})
	for attempt := 0; attempt < 2; attempt++ {
		runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
		if err != nil {
			t.Fatalf("newRuntime: %v", err)
		}
		if runtime.Requester.Admission == nil {
			t.Fatal("whole-connector policy was not attached to hook runtime requester")
		}
		if _, err := runtime.Requester.Do(context.Background(), http.MethodGet, "/hook", nil, nil); err != nil {
			t.Fatalf("hook requester call %d: %v", attempt, err)
		}
	}
	requireRateLimitWait(t, clock)
}

func TestRateLimitScopeRegistrySharesLinkedBindingAndIgnoresCredentialRevision(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)
	clock := &engineRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
	t.Cleanup(restore)
	bundle := withAllRateLimit(Bundle{Name: "scoped", HTTP: HTTPBase{URL: server.URL}})

	for _, cfg := range []connectors.RuntimeConfig{
		rateLimitTestConfigForBinding(t, "linked-binding-001", "credential-revision-a"),
		rateLimitTestConfigForBinding(t, "linked-binding-001", "credential-revision-b"),
		rateLimitTestConfigForBinding(t, "unlinked-binding-002", "credential-revision-a"),
	} {
		runtime, err := newRuntime(context.Background(), bundle, cfg, nil)
		if err != nil {
			t.Fatalf("newRuntime: %v", err)
		}
		if _, err := runtime.Requester.Do(context.Background(), http.MethodGet, "/scope", nil, nil); err != nil {
			t.Fatalf("scoped request: %v", err)
		}
	}
	if got, want := clock.waits, []time.Duration{time.Second}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("scope waits = %v, want %v: linked revision rotation must share; unlinked binding must isolate", got, want)
	}
}

func allRateLimitPolicy() *connsdk.RateLimits {
	limit, seconds := 1, 1
	return &connsdk.RateLimits{
		SchemaVersion: 1,
		State:         connsdk.RateLimitStateDeclared,
		Policies: []connsdk.RateLimitPolicy{{
			ID:       "all-http",
			Selector: connsdk.RateLimitSelector{All: true},
			Scope:    connsdk.RateLimitScope{SubjectKind: connsdk.RateLimitScopeAccount, SubjectConfig: "account_id"},
			Budgets: []connsdk.RateLimitBudget{{
				Model:         connsdk.RateLimitBudgetFixedWindow,
				Dimension:     connsdk.RateLimitBudgetSustained,
				Unit:          connsdk.RateLimitBudgetRequests,
				Limit:         &limit,
				WindowSeconds: &seconds,
			}},
		}},
	}
}

func withAllRateLimit(bundle Bundle) Bundle {
	bundle.RateLimits = allRateLimitPolicy()
	return bundle
}

func TestRuntimeAppliesProjectRateLimitAdmissionTimeout(t *testing.T) {
	cfg := rateLimitTestConfig(t)
	cfg.ProjectDir = t.TempDir()
	restore := ConfigureRateLimitAdmissionTimeout(cfg.ProjectDir, 17*time.Millisecond)
	t.Cleanup(restore)

	runtime, err := newRuntime(context.Background(), withAllRateLimit(Bundle{
		Name: "deadline-project",
		HTTP: HTTPBase{URL: "https://example.test"},
	}), cfg, nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	if got, want := runtime.Requester.RateLimitAdmissionTimeout, 17*time.Millisecond; got != want {
		t.Fatalf("requester admission timeout = %s, want %s", got, want)
	}
}

func requireRateLimitWait(t *testing.T, clock *engineRateLimitClock) {
	t.Helper()
	if len(clock.waits) != 1 || clock.waits[0] != time.Second {
		t.Fatalf("rate-limit waits = %v, want [1s]", clock.waits)
	}
}

func TestRateLimitAdmissionCoversCheckAndPaginatedRead(t *testing.T) {
	for _, tt := range []struct {
		name string
		run  func(context.Context, Bundle, connectors.RuntimeConfig) error
	}{
		{
			name: "check",
			run: func(ctx context.Context, bundle Bundle, cfg connectors.RuntimeConfig) error {
				return Check(ctx, bundle, cfg, nil)
			},
		},
		{
			name: "read",
			run: func(ctx context.Context, bundle Bundle, cfg connectors.RuntimeConfig) error {
				return Read(ctx, bundle, connectors.ReadRequest{Stream: "widgets", Config: cfg}, nil, func(connectors.Record) error { return nil })
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":[]}`))
			}))
			t.Cleanup(server.Close)
			bundle := newTestBundle(t, server, StreamSpec{})
			bundle.HTTP.Check = &RequestSpec{Method: http.MethodGet, Path: "/check"}
			bundle = withAllRateLimit(bundle)
			clock := &engineRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
			restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
			t.Cleanup(restore)
			cfg := rateLimitTestConfig(t)
			if err := tt.run(context.Background(), bundle, cfg); err != nil {
				t.Fatalf("first run: %v", err)
			}
			if err := tt.run(context.Background(), bundle, cfg); err != nil {
				t.Fatalf("second run: %v", err)
			}
			requireRateLimitWait(t, clock)
		})
	}
}

func TestRateLimitAdmissionCoversFanOutIDRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/projects":
			_, _ = w.Write([]byte(`{"data":[{"id":"project-1"}]}`))
		case "/projects/project-1/widgets":
			_, _ = w.Write([]byte(`{"data":[]}`))
		default:
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)
	bundle := withAllRateLimit(newTestBundle(t, server, StreamSpec{
		Path:    "/projects/{{ fanout.id }}/widgets",
		Records: RecordsSpec{Path: "data"},
		FanOut: &FanOutSpec{
			IDsFrom: FanOutIDsFrom{Request: &FanOutIDsRequest{Path: "/projects", RecordsPath: "data", IDField: "id"}},
			Into:    FanOutInto{PathVar: "project_id"},
		},
	}))
	clock := &engineRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
	t.Cleanup(restore)
	if err := Read(context.Background(), bundle, connectors.ReadRequest{Stream: "widgets", Config: rateLimitTestConfig(t)}, nil, func(connectors.Record) error { return nil }); err != nil {
		t.Fatalf("Read: %v", err)
	}
	requireRateLimitWait(t, clock)
}

func TestRateLimitAdmissionCoversDirectReadsAndBinaryDownload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/file" {
			_, _ = w.Write([]byte("binary fixture"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)
	config := rateLimitTestConfig(t)
	tests := []struct {
		name string
		run  func(Bundle) error
		make func() Bundle
	}{
		{
			name: "direct read",
			make: func() Bundle { return withAllRateLimit(directReadBundle(server.URL, http.MethodGet, "/items")) },
			run: func(bundle Bundle) error {
				_, err := DirectRead(context.Background(), bundle, connectors.DirectReadRequest{Method: http.MethodGet, Path: "/items", Config: config, OutputPolicy: "json_redacted"}, nil)
				return err
			},
		},
		{
			name: "operation direct read",
			make: func() Bundle {
				return withAllRateLimit(Bundle{Name: "acme", HTTP: HTTPBase{URL: server.URL}, Operations: []OperationSpec{{ID: "acme.lookup", Kind: "rest_read", Summary: "lookup", Risk: "low", Approval: "none", OutputPolicy: "json_redacted", REST: &RESTOperationSpec{Method: http.MethodGet, Path: "/lookup", MaxBytes: 1024}}}, Surface: &APISurface{Endpoints: []SurfaceEndpoint{{Method: http.MethodGet, Path: "/lookup", Operation: &SurfaceOperation{Model: "direct_read"}}}}})
			},
			run: func(bundle Bundle) error {
				_, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{Operation: "acme.lookup", Config: config}, nil)
				return err
			},
		},
		{
			name: "binary download",
			make: func() Bundle {
				return withAllRateLimit(Bundle{Name: "acme", HTTP: HTTPBase{URL: server.URL}, Operations: []OperationSpec{{ID: "acme.file", Kind: "binary_download", Summary: "file", Risk: "low", Approval: "none", Binary: &BinaryOperationSpec{Method: http.MethodGet, Path: "/file", MaxBytes: 1024}}}, Surface: &APISurface{Endpoints: []SurfaceEndpoint{{Method: http.MethodGet, Path: "/file", Operation: &SurfaceOperation{}}}}})
			},
			run: func(bundle Bundle) error {
				_, err := OperationBinaryDownload(context.Background(), bundle, BinaryDownloadRequest{Operation: "acme.file", Config: config, DestRoot: t.TempDir()}, nil)
				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := &engineRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
			restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
			t.Cleanup(restore)
			if err := tt.run(tt.make()); err != nil {
				t.Fatalf("first run: %v", err)
			}
			if err := tt.run(tt.make()); err != nil {
				t.Fatalf("second run: %v", err)
			}
			requireRateLimitWait(t, clock)
		})
	}
}

func TestRateLimitAdmissionCoversDeclarativeFormAndMultipartWrites(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)
	for _, tt := range []struct {
		name   string
		action WriteAction
		record connectors.Record
		config func(*testing.T) connectors.RuntimeConfig
	}{
		{
			name:   "form write",
			action: WriteAction{Name: "submit", Method: http.MethodPost, Path: "/form", BodyType: "form"},
			record: connectors.Record{"name": "widget"},
			config: func(t *testing.T) connectors.RuntimeConfig { return rateLimitTestConfig(t) },
		},
		{
			name:   "multipart write",
			action: WriteAction{Name: "upload", Method: http.MethodPost, Path: "/upload", BodyType: "multipart", Multipart: &MultipartSpec{Parts: []MultipartPartSpec{{Name: "file", Type: "file", Field: "file", Required: true, MaxBytes: 1024}}}},
			record: connectors.Record{"file": "payload.txt"},
			config: func(t *testing.T) connectors.RuntimeConfig {
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "payload.txt"), []byte("payload"), 0o600); err != nil {
					t.Fatalf("write multipart fixture: %v", err)
				}
				cfg := rateLimitTestConfig(t)
				cfg.ProjectDir = dir
				return cfg
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			clock := &engineRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
			restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
			t.Cleanup(restore)
			bundle := withAllRateLimit(Bundle{Name: "acme", HTTP: HTTPBase{URL: server.URL}})
			cfg := tt.config(t)
			for run := 0; run < 2; run++ {
				runtime, err := newRuntime(context.Background(), bundle, cfg, nil)
				if err != nil {
					t.Fatalf("newRuntime: %v", err)
				}
				if err := executeWriteRecord(context.Background(), bundle, tt.action, tt.record, run, cfg, runtime); err != nil {
					t.Fatalf("write %d: %v", run, err)
				}
			}
			requireRateLimitWait(t, clock)
		})
	}
}

func TestRateLimitAdmissionCoversOperationDirectWrite(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)
	clock := &engineRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
	t.Cleanup(restore)
	batchable := true
	bundle := withAllRateLimit(Bundle{
		Name: "acme", HTTP: HTTPBase{URL: server.URL},
		Operations: []OperationSpec{{ID: "acme.update", Kind: "rest_write", Summary: "update", Risk: "low", Approval: "none", OutputPolicy: "json", MutationClass: "ordinary", Batchable: &batchable, REST: &RESTOperationSpec{Method: http.MethodPost, Path: "/update", ContentType: "application/json", MaxBytes: 1024, BodySchema: json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}}}`)}}},
		Surface:    &APISurface{Endpoints: []SurfaceEndpoint{{Method: http.MethodPost, Path: "/update", Operation: &SurfaceOperation{}}}},
	})
	for run := 0; run < 2; run++ {
		req := connectors.OperationDirectWriteRequest{Operation: "acme.update", Config: rateLimitTestConfig(t), Body: map[string]any{"name": "widget"}}
		preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
		if err != nil {
			t.Fatalf("preview %d: %v", run, err)
		}
		req.PreviewDigest = preview.Digest
		if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err != nil {
			t.Fatalf("operation direct write %d: %v", run, err)
		}
	}
	requireRateLimitWait(t, clock)
}

func TestRateLimitAdmissionCoversOperationDirectMultipartWrite(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)
	clock := &engineRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
	t.Cleanup(restore)
	dir := t.TempDir()
	payload := []byte("multipart admission fixture")
	path := writeMultipartOperationSource(t, dir, "payload.txt", payload)
	bundle := withAllRateLimit(multipartOperationBundle(t, server.URL))
	for run := 0; run < 2; run++ {
		req := multipartOperationRequest(dir, path, payload)
		cfg := rateLimitTestConfig(t)
		cfg.ProjectDir = dir
		cfg.CredentialRevision = req.Config.CredentialRevision
		cfg.ConfigurationDigest = req.Config.ConfigurationDigest
		cfg.WriteApprovalScope = req.Config.WriteApprovalScope
		cfg.ApprovedPayloadSHA256 = req.Config.ApprovedPayloadSHA256
		req.Config = cfg
		preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
		if err != nil {
			t.Fatalf("preview %d: %v", run, err)
		}
		req.PreviewDigest = preview.Digest
		req.Approval = approvedEvidenceForPreview(t, preview)
		if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err != nil {
			t.Fatalf("operation direct multipart write %d: %v", run, err)
		}
	}
	requireRateLimitWait(t, clock)
}
