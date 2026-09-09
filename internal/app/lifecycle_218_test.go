package app

import (
	"context"
	"errors"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/coordination"
	"polymetrics.ai/internal/synccontract"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type borrowedCoordinator218 struct{ closes atomic.Int32 }

func (c *borrowedCoordinator218) Close() error { c.closes.Add(1); return nil }
func (c *borrowedCoordinator218) ResolveRateLimit(context.Context, string, string, connectors.RateLimitScopeKey, []connsdk.RateLimitBudget) (connsdk.RateLimitAdmission, connsdk.RateLimitObserver, error) {
	return nil, nil, nil
}
func TestAppBorrowedRuntime218(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	registry := connectors.NewEmptyRegistry()
	if err := registry.Register(&appTransportConnector{meta: connectors.Metadata{Name: "fixture218"}}); err != nil {
		t.Fatal(err)
	}
	borrowed := &borrowedCoordinator218{}
	for _, reverse := range []bool{false, true} {
		options := RuntimeOptions{SharedRateLimits: borrowed}
		var a *App
		var err error
		if reverse {
			a, err = OpenForReverseExecutionWithRegistryAndRuntime(root, registry, options)
		} else {
			a, err = OpenWithRegistryAndRuntime(root, registry, options)
		}
		if err != nil {
			t.Fatal(err)
		}
		if a.Registry() != registry || a.ExecutableRuntime(connectors.RuntimeConfig{}).SharedRateLimits != borrowed {
			t.Fatal("supplied runtime/registry lost")
		}
		if !reverse {
			if _, err := a.AddCredential(context.Background(), AddCredentialRequest{Name: "fixture218", Connector: "fixture218"}); err != nil {
				t.Fatal(err)
			}
		}
		_, runtime, err := a.ResolveConnectorCredential(context.Background(), "fixture218", "fixture218", nil)
		if err != nil || runtime.SharedRateLimits != borrowed {
			t.Fatal("credential runtime lost borrowed capability")
		}
		if err := a.Close(); err != nil {
			t.Fatal(err)
		}
		_ = a.Close()
		if borrowed.closes.Load() != 0 {
			t.Fatal("App closed borrowed capability")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if a, err := OpenWithRegistryAndRuntime(root, registry, RuntimeOptions{Context: ctx, SharedRateLimits: borrowed}); a != nil || !errors.Is(err, context.Canceled) {
		t.Fatal("canceled constructor did not cleanly fail")
	}
	if borrowed.closes.Load() != 0 {
		t.Fatal("failed constructor closed borrowed owner")
	}
}

func TestAppParkingLifecycle218(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(map[bool]string{false: "ordinary", true: "reverse"}[reverse], func(t *testing.T) {
			root := t.TempDir()
			if err := InitProject(root); err != nil {
				t.Fatal(err)
			}
			at := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
			committed := at.Add(time.Second)
			observed := true
			checkpoint := synccontract.CheckpointEnvelope{StateVersion: synccontract.StateVersion, Source: synccontract.SourceIdentity{Engine: "fixture218", AccountOrCluster: "fixture-account", ObjectScope: "records"}, Mechanism: "fixture-parking", SchemaVersion: "v1", ProtocolVersion: "v1", SourceGeneration: synccontract.OpaqueToken("generation"), SnapshotBarrier: &synccontract.SnapshotBarrier{Kind: "fixture", Token: synccontract.OpaqueToken("barrier")}, Position: synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("literal-position")}, PositionObserved: &observed, Partitions: []synccontract.PartitionState{}, Dedupe: synccontract.DedupeIdentity{Kind: "fixture", Value: synccontract.OpaqueToken("identity")}, DedupeWindow: synccontract.DedupeWindow{Kind: "fixture", Start: synccontract.OpaqueToken("start"), End: synccontract.OpaqueToken("end")}, ObservedAt: at, CommittedAt: &committed}
			if err := checkpoint.Validate(); err != nil {
				t.Fatalf("invalid literal fixture before constructor: %v", err)
			}
			borrowed := &borrowedCoordinator218{}
			entered, canceled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
			var releaseOnce sync.Once
			releaseCallback := func() { releaseOnce.Do(func() { close(release) }) }
			defer releaseCallback()
			created := make(chan *App, 1)
			var store coordination.RateParkingStore
			options := RuntimeOptions{SharedRateLimits: borrowed, newParking: func(a *App, s coordination.RateParkingStore) *coordination.RateParkingCoordinator {
				store = s
				seed := coordination.NewRateParkingCoordinator(coordination.RateParkingCoordinatorOptions{Store: s, Scheduler: &appRateParkingTestScheduler{}, Now: func() time.Time { return at }, Resume: func(context.Context, coordination.ParkedRateLimitRun) error { return nil }})
				if err := seed.Start(context.Background()); err != nil {
					t.Error(err)
					return nil
				}
				if _, err := seed.Park(context.Background(), coordination.RateParkingRequest{RunID: "literal-run218", Scope: "literal-scope218", Checkpoint: checkpoint, ResetAt: at.Add(time.Hour), Reason: connsdk.RateLimitObservationSourceRetryAfter}); err != nil {
					t.Error(err)
					seed.Close()
					return nil
				}
				seed.Close()
				c := coordination.NewRateParkingCoordinator(coordination.RateParkingCoordinatorOptions{Store: s, Scheduler: &appRateParkingTestScheduler{}, Now: func() time.Time { return at.Add(2 * time.Hour) }, Resume: func(ctx context.Context, run coordination.ParkedRateLimitRun) error {
					if a.ExecutableRuntime(connectors.RuntimeConfig{}).SharedRateLimits != borrowed || run.RunID != "literal-run218" || !reflect.DeepEqual(run.Checkpoint, checkpoint) {
						t.Error("due resume lost injected capability or retained checkpoint")
					}
					close(entered)
					<-ctx.Done()
					close(canceled)
					<-release
					return ctx.Err()
				}})
				created <- a
				return c
			}}
			done := make(chan error, 1)
			go func() {
				var a *App
				var err error
				if reverse {
					a, err = OpenForReverseExecutionWithRegistryAndRuntime(root, connectors.NewRegistry(), options)
				} else {
					a, err = OpenWithRegistryAndRuntime(root, connectors.NewRegistry(), options)
				}
				if a != nil {
					_ = a.Close()
				}
				done <- err
			}()
			var a *App
			select {
			case a = <-created:
			case err := <-done:
				t.Fatalf("parking setup did not complete: %v", err)
			case <-time.After(5 * time.Second):
				t.Fatal("constructor did not reach parking creation")
			}
			select {
			case <-entered:
			case <-time.After(5 * time.Second):
				t.Fatal("due resume not reached during construction")
			}
			closed := make(chan struct{})
			go func() { _ = a.Close(); close(closed) }()
			select {
			case <-canceled:
			case <-time.After(5 * time.Second):
				t.Fatal("App Close did not cancel active callback")
			}
			select {
			case <-closed:
				t.Fatal("App Close returned before callback joined")
			default:
			}
			releaseCallback()
			select {
			case <-closed:
			case <-time.After(5 * time.Second):
				t.Fatal("App Close deadlocked with callback")
			}
			if err := <-done; !errors.Is(err, context.Canceled) {
				t.Fatalf("interrupted constructor error=%v", err)
			}
			runs, err := store.List()
			if err != nil || len(runs) != 1 || runs[0].RunID != "literal-run218" || !reflect.DeepEqual(runs[0].Checkpoint, checkpoint) {
				t.Fatal("teardown deleted or changed durable acknowledged work")
			}
			if borrowed.closes.Load() != 0 {
				t.Fatal("parking teardown closed borrowed capability")
			}
		})
	}
}
