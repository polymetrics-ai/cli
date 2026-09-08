package coordination

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime/pprof"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

type scopeClient218 struct {
	closes  atomic.Int32
	calls   atomic.Int32
	err     error
	resolve func(context.Context) error
}

func (c *scopeClient218) Close() error { c.closes.Add(1); return c.err }
func (c *scopeClient218) ResolveRateLimit(ctx context.Context, connector, policy string, scope connectors.RateLimitScopeKey, budgets []connsdk.RateLimitBudget) (connsdk.RateLimitAdmission, connsdk.RateLimitObserver, error) {
	c.calls.Add(1)
	if c.resolve != nil {
		if err := c.resolve(ctx); err != nil {
			return nil, nil, err
		}
	}
	if c.closes.Load() != 0 {
		return nil, nil, errors.New("closed fixture client")
	}
	return nil, nil, nil
}
func TestSharedRateLimitScopeOwnership218(t *testing.T) {
	t.Run("unused", func(t *testing.T) {
		opened := 0
		s := NewSharedRateLimitScope("snapshot", SharedRateLimitScopeOptions{Open: func(string) (OwnedSharedRateLimitClient, error) { opened++; return &scopeClient218{}, nil }})
		for range 2 {
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
		}
		if opened != 0 {
			t.Fatal("unused close constructed a client")
		}
		_, _, err := s.ResolveRateLimit(context.Background(), "connector", "policy", "opaque-scope", nil)
		var unavailable *SharedRateLimitUnavailableError
		if !errors.As(err, &unavailable) || unavailable.Reason != SharedRateLimitCoordinatorClosed {
			t.Fatal("closed owner admitted initialization")
		}
	})
	t.Run("single_allocation_and_close", func(t *testing.T) {
		var opens atomic.Int32
		c := &scopeClient218{}
		s := NewSharedRateLimitScope("snapshot-a", SharedRateLimitScopeOptions{Open: func(addr string) (OwnedSharedRateLimitClient, error) {
			if addr != "snapshot-a" {
				t.Error("configuration changed")
			}
			opens.Add(1)
			return c, nil
		}})
		var wg sync.WaitGroup
		for range 32 {
			wg.Go(func() {
				client, err := s.acquire()
				if err != nil || client != c {
					t.Error("wrong owned identity")
				}
			})
		}
		wg.Wait()
		for range 32 {
			wg.Go(func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			})
		}
		wg.Wait()
		if opens.Load() != 1 || c.closes.Load() != 1 {
			t.Fatal("duplicate allocation/closure")
		}
	})
	t.Run("partial_constructor_and_close_errors", func(t *testing.T) {
		created := errors.New("constructor completion")
		closed := errors.New("close completion")
		c := &scopeClient218{err: closed}
		s := NewSharedRateLimitScope("private", SharedRateLimitScopeOptions{Open: func(string) (OwnedSharedRateLimitClient, error) { return c, created }})
		if _, err := s.acquire(); err == nil {
			t.Fatal("constructor error accepted")
		}
		for range 2 {
			err := s.Close()
			if !errors.Is(err, created) || !errors.Is(err, closed) {
				t.Fatal("compound completion causes lost")
			}
		}
		if c.closes.Load() != 1 {
			t.Fatal("partial client not closed exactly once")
		}
	})
	t.Run("first_use_close_race", func(t *testing.T) {
		var opens atomic.Int32
		c := &scopeClient218{}
		s := NewSharedRateLimitScope("a", SharedRateLimitScopeOptions{Open: func(string) (OwnedSharedRateLimitClient, error) { opens.Add(1); return c, nil }})
		var wg sync.WaitGroup
		for range 16 {
			wg.Go(func() { _, _ = s.acquire() })
			wg.Go(func() { _ = s.Close() })
		}
		wg.Wait()
		if opens.Load() > 1 || c.closes.Load() != opens.Load() {
			t.Fatal("allocation escaped owner close")
		}
	})
	t.Run("caller_context_each_resolution", func(t *testing.T) {
		type contextKey218 struct{}
		key := contextKey218{}
		calls := 0
		c := &scopeClient218{resolve: func(ctx context.Context) error {
			calls++
			if ctx.Value(key) != calls {
				t.Error("stale availability context")
			}
			return ctx.Err()
		}}
		s := NewSharedRateLimitScope("a", SharedRateLimitScopeOptions{Open: func(string) (OwnedSharedRateLimitClient, error) { return c, nil }})
		t.Cleanup(func() {
			if err := s.Close(); err != nil {
				t.Error(err)
			}
		})
		for i := 1; i <= 2; i++ {
			_, _, err := s.ResolveRateLimit(context.WithValue(context.Background(), key, i), "c", "p", "scope", nil)
			if err != nil {
				t.Fatal(err)
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, _, err := s.ResolveRateLimit(ctx, "c", "p", "scope", nil)
		if !errors.Is(err, context.Canceled) || calls != 2 {
			t.Fatal("canceled use reached client")
		}
	})
	t.Run("concurrent_scopes_isolated", func(t *testing.T) {
		for _, same := range []bool{false, true} {
			a, b := &scopeClient218{}, &scopeClient218{}
			entered, release := make(chan struct{}), make(chan struct{})
			b.resolve = func(context.Context) error { close(entered); <-release; return nil }
			addrB := "b"
			if same {
				addrB = "a"
			}
			sa := NewSharedRateLimitScope("a", SharedRateLimitScopeOptions{Open: func(string) (OwnedSharedRateLimitClient, error) { return a, nil }})
			sb := NewSharedRateLimitScope(addrB, SharedRateLimitScopeOptions{Open: func(string) (OwnedSharedRateLimitClient, error) { return b, nil }})
			_, _, err := sa.ResolveRateLimit(context.Background(), "c", "p", "scope-a", nil)
			if err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { _, _, err := sb.ResolveRateLimit(context.Background(), "c", "p", "scope-b", nil); done <- err }()
			<-entered
			if err := sa.Close(); err != nil {
				t.Fatal(err)
			}
			if b.closes.Load() != 0 {
				t.Fatal("A closed B")
			}
			close(release)
			if err := <-done; err != nil {
				t.Fatal(err)
			}
			_ = sb.Close()
			if a.closes.Load() != 1 || b.closes.Load() != 1 {
				t.Fatal("wrong close identities")
			}
		}
	})
}

type pingHook218 struct{ calls *atomic.Int32 }

func (h pingHook218) DialHook(redis.DialHook) redis.DialHook {
	return func(context.Context, string, string) (net.Conn, error) { return nil, errors.New("unexpected dial") }
}
func (h pingHook218) ProcessHook(redis.ProcessHook) redis.ProcessHook {
	return func(_ context.Context, c redis.Cmder) error {
		h.calls.Add(1)
		if c.Name() != "ping" {
			return errors.New("unexpected Redis command")
		}
		c.(*redis.StatusCmd).SetVal("PONG")
		return nil
	}
}
func (h pingHook218) ProcessPipelineHook(redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(context.Context, []redis.Cmder) error { return errors.New("unexpected pipeline") }
}
func TestSharedRateLimitBorrowedRegistryIdentity218(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "unused.invalid:6379"})
	var calls atomic.Int32
	client.AddHook(pingHook218{&calls})
	r := NewSharedRateLimitRegistry(&Dragonfly{client: client})
	t.Cleanup(func() {
		if err := r.Close(); err != nil {
			t.Error(err)
		}
	})
	budgets := []connsdk.RateLimitBudget{fixedRequestBudget(7, 60)}
	for range 2 {
		a, o, err := r.ResolveRateLimit(context.Background(), "literal-connector", "literal-policy", "literal-opaque-scope", budgets)
		if err != nil {
			t.Fatal(err)
		}
		limiter, ok := a.(*SharedRateLimiter)
		if !ok || o != limiter || limiter.registry != r || limiter.key != (RateLimitKey{Connector: "literal-connector", PolicyID: "literal-policy", Scope: "literal-opaque-scope"}) {
			t.Fatal("borrowed capability changed actual shared limiter identity")
		}
		if *limiter.budgets[0].Limit != 7 {
			t.Fatal("budget changed")
		}
	}
	if calls.Load() != 2 {
		t.Fatal("availability result was cached")
	}
}

type observedRealClient218 struct {
	*SharedRateLimitRegistry
	id     string
	closed *atomic.Int32
}

func (c *observedRealClient218) Close() error {
	c.closed.Add(1)
	return c.SharedRateLimitRegistry.Close()
}
func TestSharedRateLimitRealResourceChild218(t *testing.T) {
	if os.Getenv("PM_TEST_SHARED_REAL_CHILD_218") != "1" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSharedRateLimitRealResourceChild218$", "-test.v", "-test.count=1")
		cmd.Env = append(os.Environ(), "PM_TEST_SHARED_REAL_CHILD_218=1")
		out, err := cmd.CombinedOutput()
		t.Logf("real child:\n%s", out)
		if err != nil {
			t.Fatalf("child failed: %v", err)
		}
		return
	}
	loops := func() int {
		var b bytes.Buffer
		if err := pprof.Lookup("goroutine").WriteTo(&b, 2); err != nil {
			t.Fatal(err)
		}
		return strings.Count(b.String(), ".(*CircuitBreakerManager).cleanupLoop(")
	}
	before := loops()
	var closed atomic.Int32
	ids := map[string]bool{}
	scopes := []*SharedRateLimitScope{}
	for range 4 {
		s := NewSharedRateLimitScope("unused.invalid:6379", SharedRateLimitScopeOptions{Open: func(addr string) (OwnedSharedRateLimitClient, error) {
			r := OpenSharedRateLimitRegistry(addr)
			id := fmt.Sprintf("%p", r.dragonfly.client)
			if ids[id] {
				t.Fatal("real client identity reused while live")
			}
			ids[id] = true
			t.Logf("constructed actual redis.Client id=%s", id)
			return &observedRealClient218{r, id, &closed}, nil
		}})
		c, err := s.acquire()
		if err != nil {
			t.Fatal(err)
		}
		if again, err := s.acquire(); err != nil || again != c {
			t.Fatal("real owner identity changed")
		}
		scopes = append(scopes, s)
	}
	deadline := time.Now().Add(2 * time.Second)
	for loops() != before+4 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	during := loops()
	if during != before+4 {
		t.Fatalf("real constructors=%d maintenance loops before=%d during=%d", len(ids), before, during)
	}
	for _, s := range scopes {
		client := s.client.(*observedRealClient218)
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		_ = s.Close()
		t.Logf("closed actual redis.Client id=%s", client.id)
	}
	deadline = time.Now().Add(2 * time.Second)
	for loops() != before && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	after := loops()
	t.Logf("actual_clients=%d actual_closes=%d maintenance_loops before=%d during=%d after=%d", len(ids), closed.Load(), before, during, after)
	if closed.Load() != 4 || after != before {
		t.Fatal("real owned maintenance did not quiesce")
	}
}
