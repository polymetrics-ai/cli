package engine

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/coordination"
	"sync/atomic"
	"testing"
	"time"
)

type borrowedRateCoordinator218 struct {
	missingHandles   bool
	missingAdmission bool
	missingObserver  bool
	calls            atomic.Int32
	admissions       atomic.Int32
	err              error
}

func (c *borrowedRateCoordinator218) ResolveRateLimit(ctx context.Context, _ string, _ string, _ connectors.RateLimitScopeKey, _ []connsdk.RateLimitBudget) (connsdk.RateLimitAdmission, connsdk.RateLimitObserver, error) {
	c.calls.Add(1)
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if c.err != nil {
		return nil, nil, c.err
	}
	if c.missingHandles {
		return nil, nil, nil
	}
	if c.missingAdmission {
		return nil, c, nil
	}
	if c.missingObserver {
		return c, nil, nil
	}
	return c, c, nil
}
func (c *borrowedRateCoordinator218) Admit(ctx context.Context, _ connsdk.RateLimitRequest) error {
	c.admissions.Add(1)
	return ctx.Err()
}
func (c *borrowedRateCoordinator218) Observe(context.Context, connsdk.RateLimitObservation) {}
func TestBorrowedSharedRateConsumers218(t *testing.T) {
	for _, consumer := range []string{"check", "saved_read", "direct_read", "typed_write"} {
		for _, mode := range []string{"shared_healthy", "local", "nil", "unavailable", "canceled", "deadline", "closed", "missing_handles", "missing_admission", "missing_observer"} {
			t.Run(consumer+"/"+mode, func(t *testing.T) {
				var sends atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					sends.Add(1)
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"data":[{"id":"literal-record218","name":"retained"}],"ok":true}`))
				}))
				defer server.Close()
				c := &borrowedRateCoordinator218{}
				cfg := rateLimitTestConfig(t)
				cfg.SharedRateLimits = c
				bundle := withAllRateLimit(newTestBundle(t, server, StreamSpec{}))
				bundle.HTTP.Check = &RequestSpec{Method: http.MethodGet, Path: "/check"}
				if consumer == "direct_read" {
					bundle = withAllRateLimit(directReadBundle(server.URL, http.MethodGet, "/items"))
				}
				bundle.RateLimits.Policies[0].Coordination = connsdk.RateLimitCoordinationRequireShared
				ctx := context.Background()
				var cancel context.CancelFunc
				switch mode {
				case "local":
					bundle.RateLimits.Policies[0].Coordination = ""
				case "nil":
					cfg.SharedRateLimits = nil
				case "missing_admission":
					c.missingAdmission = true
				case "missing_observer":
					c.missingObserver = true
				case "missing_handles":
					c.missingHandles = true
				case "unavailable":
					c.err = &coordination.SharedRateLimitUnavailableError{Component: "dragonfly", Reason: coordination.SharedRateLimitCoordinatorUnreachable}
				case "canceled":
					ctx, cancel = context.WithCancel(ctx)
					cancel()
				case "deadline":
					ctx, cancel = context.WithDeadline(ctx, time.Now().Add(-time.Second))
					defer cancel()
				case "closed":
					s := coordination.NewSharedRateLimitScope("", coordination.SharedRateLimitScopeOptions{})
					_ = s.Close()
					cfg.SharedRateLimits = s
				}
				var err error
				rows := 0
				switch consumer {
				case "check":
					err = Check(ctx, bundle, cfg, nil)
				case "saved_read":
					err = Read(ctx, bundle, connectors.ReadRequest{Stream: "widgets", Config: cfg}, nil, func(r connectors.Record) error {
						rows++
						if r["id"] != "literal-record218" {
							t.Error("wrong record")
						}
						return nil
					})
				case "direct_read":
					var result connectors.DirectReadResult
					result, err = DirectRead(ctx, bundle, connectors.DirectReadRequest{Method: http.MethodGet, Path: "/items", Config: cfg, OutputPolicy: "json_redacted"}, nil)
					if err == nil && result.Body == nil {
						t.Error("direct result missing")
					}
				case "typed_write":
					var runtime *Runtime
					runtime, err = newRuntime(ctx, bundle, cfg, nil)
					if err == nil {
						err = executeWriteRecord(ctx, bundle, WriteAction{Name: "submit", Method: http.MethodPost, Path: "/form", BodyType: "form"}, connectors.Record{"name": "literal218"}, 0, cfg, runtime)
					}
				}
				healthy := mode == "shared_healthy" || mode == "local"
				if healthy {
					if err != nil || sends.Load() != 1 {
						t.Fatalf("healthy consumer err=%v sends=%d", err, sends.Load())
					}
					if consumer == "saved_read" && rows != 1 {
						t.Fatal("saved read lost retained record")
					}
					if mode == "local" && c.calls.Load() != 0 {
						t.Fatal("local policy consulted shared capability")
					}
					if mode == "shared_healthy" && (c.calls.Load() == 0 || c.admissions.Load() != 1) {
						t.Fatal("shared consumer bypassed admission")
					}
					return
				}
				if err == nil || sends.Load() != 0 || rows != 0 {
					t.Fatalf("refusal err=%v sends=%d rows=%d", err, sends.Load(), rows)
				}
				switch mode {
				case "canceled":
					if !errors.Is(err, context.Canceled) {
						t.Fatal("cancellation identity lost")
					}
				case "deadline":
					if !errors.Is(err, context.DeadlineExceeded) {
						t.Fatal("deadline identity lost")
					}
				default:
					var typed *coordination.SharedRateLimitUnavailableError
					var refusal *connsdk.RateBudgetRefusalError
					if !errors.As(err, &typed) || !errors.As(err, &refusal) {
						t.Fatalf("shared refusal types lost: %T", err)
					}
				}
			})
		}
	}
}
