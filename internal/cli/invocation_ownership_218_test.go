package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/coordination"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type invocationCoordinator218 struct {
	closed        atomic.Int32
	resolved      atomic.Int32
	admitted      atomic.Int32
	beforeResolve func()
	closeErr      error
}

func (c *invocationCoordinator218) Close() error { c.closed.Add(1); return c.closeErr }
func (c *invocationCoordinator218) ResolveRateLimit(ctx context.Context, _ string, _ string, _ connectors.RateLimitScopeKey, _ []connsdk.RateLimitBudget) (connsdk.RateLimitAdmission, connsdk.RateLimitObserver, error) {
	c.resolved.Add(1)
	if c.beforeResolve != nil {
		c.beforeResolve()
	}
	if c.closed.Load() != 0 {
		return nil, nil, errors.New("another invocation closed this client")
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	return c, c, nil
}
func (c *invocationCoordinator218) Admit(ctx context.Context, _ connsdk.RateLimitRequest) error {
	c.admitted.Add(1)
	return ctx.Err()
}
func (c *invocationCoordinator218) Observe(context.Context, connsdk.RateLimitObservation) {}
func TestCLIInvocationFailureOwnership218(t *testing.T) {
	for _, tc := range []struct {
		name         string
		args         []string
		project      bool
		registryFail bool
		want         int
	}{{"success", []string{"version"}, false, false, 0}, {"registry_failure", []string{"version"}, false, true, 1}, {"project_failure", []string{"credentials", "list"}, false, false, 1}, {"credential_failure", []string{"credentials", "test", "absent"}, true, false, 1}, {"command_failure", []string{"unknown218"}, true, false, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if tc.project {
				if err := app.InitProject(root); err != nil {
					t.Fatal(err)
				}
			}
			c := &invocationCoordinator218{closeErr: errors.New("private endpoint must not leak")}
			o := defaultAppOpeners()
			created := 0
			o.newSharedScope = func(string) coordination.OwnedSharedRateLimitClient { created++; return c }
			if tc.registryFail {
				o.newRegistry = func() (*connectors.Registry, error) { return nil, errors.New("late registry refusal") }
			}
			var out, diag bytes.Buffer
			code := run(append([]string{"--root", root}, tc.args...), &out, &diag, o)
			if (code == 0) != (tc.want == 0) || created != 1 || c.closed.Load() != 1 {
				t.Fatalf("result/ownership code=%d created=%d closed=%d", code, created, c.closed.Load())
			}
			if !strings.Contains(diag.String(), "warning: invocation resource cleanup failed") || strings.Contains(diag.String(), "private endpoint") {
				t.Fatal("cleanup error lost or unsanitized")
			}
			if tc.want == 0 && out.Len() == 0 {
				t.Fatal("cleanup changed completed command result")
			}
		})
	}
}

func setupInvocationEngine218(t *testing.T, url string) (string, *connectors.Registry) {
	t.Helper()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatal(err)
	}
	limit, window := 10, 60
	b := engine.Bundle{Name: "fixture218", Metadata: engine.Metadata{Name: "fixture218", DisplayName: "Fixture218", IntegrationType: "api"}, HTTP: engine.HTTPBase{URL: url, Check: &engine.RequestSpec{Method: http.MethodGet, Path: "/check"}}, RateLimits: &connsdk.RateLimits{SchemaVersion: 1, State: connsdk.RateLimitStateDeclared, Policies: []connsdk.RateLimitPolicy{{ID: "literal-policy218", Coordination: connsdk.RateLimitCoordinationRequireShared, Selector: connsdk.RateLimitSelector{All: true}, Scope: connsdk.RateLimitScope{SubjectKind: connsdk.RateLimitScopeAccount, SubjectConfig: "account_id"}, Budgets: []connsdk.RateLimitBudget{{Model: connsdk.RateLimitBudgetFixedWindow, Dimension: connsdk.RateLimitBudgetSustained, Unit: connsdk.RateLimitBudgetRequests, Limit: &limit, WindowSeconds: &window}}}}}}
	registry := connectors.NewEmptyRegistry()
	if err := registry.Register(engine.New(b, nil)); err != nil {
		t.Fatal(err)
	}
	a, err := app.OpenWithRegistry(root, registry)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := a.Close(); err != nil {
			t.Error(err)
		}
	})
	if _, err := a.AddCredential(context.Background(), app.AddCredentialRequest{Name: "literal-credential218", Connector: "fixture218", Config: map[string]string{"account_id": "literal-account218"}}); err != nil {
		t.Fatal(err)
	}
	return root, registry
}
func TestCLIConcurrentInvocationOwnership218(t *testing.T) {
	var sends atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/check" {
			t.Error("wrong Check path")
		}
		sends.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	rootA, registryA := setupInvocationEngine218(t, server.URL)
	rootB, registryB := setupInvocationEngine218(t, server.URL)
	a, b := &invocationCoordinator218{}, &invocationCoordinator218{}
	entered, release := make(chan struct{}), make(chan struct{})
	var enteredOnce, releaseOnce sync.Once
	releaseB := func() { releaseOnce.Do(func() { close(release) }) }
	defer releaseB()
	b.beforeResolve = func() { enteredOnce.Do(func() { close(entered) }); <-release }
	invoke := func(root string, registry *connectors.Registry, c *invocationCoordinator218) int {
		o := defaultAppOpeners()
		o.registry = registry
		o.newSharedScope = func(string) coordination.OwnedSharedRateLimitClient { return c }
		var out, diag bytes.Buffer
		code := run([]string{"--root", root, "--json", "credentials", "test", "literal-credential218"}, &out, &diag, o)
		if code == 0 && !strings.Contains(out.String(), `"status": "ok"`) && !strings.Contains(out.String(), `"status":"ok"`) {
			t.Error("actual credential result lost")
		}
		return code
	}
	done := make(chan int, 1)
	go func() { done <- invoke(rootB, registryB, b) }()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("B did not reach actual shared admission")
	}
	if code := invoke(rootA, registryA, a); code != 0 {
		t.Fatalf("A exit=%d", code)
	}
	if a.closed.Load() != 1 || b.closed.Load() != 0 || sends.Load() != 1 {
		t.Fatal("A crossed B ownership or send boundary")
	}
	releaseB()
	if code := <-done; code != 0 {
		t.Fatalf("B exit=%d", code)
	}
	if b.closed.Load() != 1 || b.admitted.Load() != 1 || sends.Load() != 2 {
		t.Fatal("B did not complete with its own live capability")
	}
}
func TestCLIBorrowsSuppliedAppAndCoordinator218(t *testing.T) {
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatal(err)
	}
	c := &invocationCoordinator218{}
	a, err := app.OpenWithRuntime(root, app.RuntimeOptions{SharedRateLimits: c})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := a.Close(); err != nil {
			t.Error(err)
		}
	})
	o := defaultAppOpeners()
	o.runtimeOptions.SharedRateLimits = c
	if code := run([]string{"--root", root, "version"}, io.Discard, io.Discard, o); code != 0 {
		t.Fatal("borrowed capability Run failed")
	}
	if code := runWithAppOpeners([]string{"--root", root, "credentials", "list"}, io.Discard, io.Discard, func(string) (*app.App, error) { return a, nil }, func(string) (*app.App, error) { return a, nil }); code != 0 {
		t.Fatal("borrowed App Run failed")
	}
	if c.closed.Load() != 0 {
		t.Fatal("CLI closed borrowed capability")
	}
	// The custom opener's App remains usable for its owner after CLI returns.
	if _, err := a.AddCredential(context.Background(), app.AddCredentialRequest{Name: "local218", Connector: "file", Config: map[string]string{"path": "fixture.jsonl"}}); err != nil {
		t.Fatal(err)
	}
}
