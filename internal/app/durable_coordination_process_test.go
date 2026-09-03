package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/coordination"
	"polymetrics.ai/internal/synccontract"
)

const (
	productionParkingHelperEnv     = "POLYMETRICS_PRODUCTION_PARKING_HELPER"
	productionParkingConnectorName = "github"
	productionParkingResetDelay    = 30 * time.Second
)

func productionParkingRegistry() (*connectors.Registry, error) {
	registry := connectors.NewEmptyRegistry()
	if err := registry.Register(&productionParkingSource{}); err != nil {
		return nil, err
	}
	if err := registry.Register(connectors.Warehouse{}); err != nil {
		return nil, err
	}
	return registry, nil
}

// TestProductionParkingCompositionSurvivesProcessKill drives app.Open,
// credential/connection resolution, ETL dispatch, durable admission, and
// automatic resume in three distinct processes.
func TestProductionParkingCompositionSurvivesProcessKill(t *testing.T) {
	mode := os.Getenv(productionParkingHelperEnv)
	if mode != "" {
		runProductionParkingHelper(t, mode)
		return
	}
	root := t.TempDir()
	setup := productionParkingCommand(root, "setup", "")
	stdout, err := setup.StdoutPipe()
	if err != nil {
		t.Fatalf("setup stdout: %v", err)
	}
	setup.Stderr = os.Stderr
	if err := setup.Start(); err != nil {
		t.Fatalf("start setup process: %v", err)
	}
	ready, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil || !strings.HasPrefix(ready, "parked:") {
		_ = setup.Process.Kill()
		t.Fatalf("setup readiness = %q, %v", ready, err)
	}
	runID := strings.TrimSpace(strings.TrimPrefix(ready, "parked:"))
	if err := setup.Process.Kill(); err != nil {
		t.Fatalf("kill setup process: %v", err)
	}
	if err := setup.Wait(); err == nil {
		t.Fatal("killed setup process unexpectedly succeeded")
	}

	counter := filepath.Join(root, "provider-sends")
	if got := productionParkingSendCount(t, counter); got != 2 {
		t.Fatalf("provider sends before restart = %d, want initial success plus terminal 429", got)
	}
	parkingStore, err := coordination.OpenFileRateParkingStore(filepath.Join(root, ".polymetrics", "state", "rate-parking.json"))
	if err != nil {
		t.Fatalf("open production parking store after killed writer: %v", err)
	}
	parkedRuns, err := parkingStore.List()
	if err != nil || len(parkedRuns) != 1 || parkedRuns[0].RunID != runID {
		t.Fatalf("production parking records after killed writer = %#v, %v; want run %q", parkedRuns, err, runID)
	}
	resetAt := parkedRuns[0].ResetAt
	// Keep the post-crash admission proof honest under a slow runner: the
	// restarted process must still be fenced after startup work consumes more
	// than the old two-second reset window.
	time.Sleep(3 * time.Second)
	if !time.Now().Before(resetAt) {
		t.Fatalf("fixture reset elapsed before restarted admission: reset_at=%s", resetAt)
	}
	admit := productionParkingCommand(root, "admission", runID)
	if output, err := admit.CombinedOutput(); err != nil {
		t.Fatalf("admission process: %v\n%s", err, output)
	}
	if got := productionParkingSendCount(t, counter); got != 2 {
		t.Fatalf("parked-scope refusal provider sends = %d, want 2 (zero new sends)", got)
	}

	if wait := time.Until(resetAt); wait > 0 {
		time.Sleep(wait + 100*time.Millisecond)
	}
	resume := productionParkingCommand(root, "resume", runID)
	if output, err := resume.CombinedOutput(); err != nil {
		t.Fatalf("resume process: %v\n%s", err, output)
	}
	if got := productionParkingSendCount(t, counter); got != 3 {
		t.Fatalf("provider sends after one resume = %d, want 3", got)
	}
	if runs, err := parkingStore.List(); err != nil || len(runs) != 0 {
		t.Fatalf("production parking records after resume = %#v, %v; want empty", runs, err)
	}
}

func productionParkingCommand(root, mode, runID string) *exec.Cmd {
	command := exec.Command(os.Args[0], "-test.run=^TestProductionParkingCompositionSurvivesProcessKill$")
	command.Env = append(os.Environ(),
		productionParkingHelperEnv+"="+mode,
		"POLYMETRICS_PRODUCTION_PARKING_ROOT="+root,
		"POLYMETRICS_PRODUCTION_PARKING_RUN="+runID,
	)
	return command
}

func runProductionParkingHelper(t *testing.T, mode string) {
	root := os.Getenv("POLYMETRICS_PRODUCTION_PARKING_ROOT")
	runID := os.Getenv("POLYMETRICS_PRODUCTION_PARKING_RUN")
	switch mode {
	case "setup":
		if err := InitProject(root); err != nil {
			t.Fatal(err)
		}
		a, err := openWithRegistry(root, false, productionParkingRegistry)
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: productionParkingConnectorName}); err != nil {
			t.Fatal(err)
		}
		if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "warehouse", Connector: "warehouse", Config: map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")}}); err != nil {
			t.Fatal(err)
		}
		if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
			Name: "durable_parking", Source: EndpointConfig{Connector: productionParkingConnectorName, Credential: "source"},
			Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
			Streams:     map[string]StreamConfig{"customers": {SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, DestinationTable: "durable_customers"}},
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := a.RunETL(ctx, RunETLRequest{Connection: "durable_parking", Stream: "customers"}); err != nil {
			t.Fatalf("initial checkpoint run: %v", err)
		}
		productionParkingReads.Store(1)
		parked, err := a.RunETL(ctx, RunETLRequest{Connection: "durable_parking", Stream: "customers"})
		if !errors.Is(err, coordination.ErrRateLimitParked) || parked.Status != string(coordination.RateParkingOutcomeParkedRateLimit) {
			t.Fatalf("parking run = %+v, %v", parked, err)
		}
		fmt.Printf("parked:%s\n", parked.ID)
		select {}
	case "admission":
		a, err := openWithRegistry(root, false, productionParkingRegistry)
		if err != nil {
			t.Fatal(err)
		}
		before := a.state.StreamStates[streamStateKey("durable_parking", "customers")]
		_, err = a.RunETL(context.Background(), RunETLRequest{Connection: "durable_parking", Stream: "customers"})
		if !errors.Is(err, coordination.ErrRateLimitParked) {
			t.Fatalf("same-scope admission error = %v, want ErrRateLimitParked", err)
		}
		after := a.state.StreamStates[streamStateKey("durable_parking", "customers")]
		if !transportStreamStateEqual(before, after) {
			t.Fatal("parked-scope refusal advanced the checkpoint")
		}
	case "resume":
		a, err := openWithRegistry(root, false, productionParkingRegistry)
		if err != nil {
			t.Fatal(err)
		}
		run, err := a.GetRun(runID)
		if err != nil || run.Status != "resumed" {
			t.Fatalf("resumed original run = %+v, %v", run, err)
		}
	default:
		t.Fatalf("unknown production parking helper mode %q", mode)
	}
}

var productionParkingReads atomic.Int64

type productionParkingSource struct{}

func (*productionParkingSource) Name() string { return productionParkingConnectorName }
func (*productionParkingSource) Metadata() connectors.Metadata {
	metadata := (connectors.Sample{}).Metadata()
	metadata.Name = productionParkingConnectorName
	return metadata
}
func (*productionParkingSource) Check(context.Context, connectors.RuntimeConfig) error { return nil }
func (*productionParkingSource) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	catalog, err := (connectors.Sample{}).Catalog(ctx, cfg)
	catalog.Connector = productionParkingConnectorName
	return catalog, err
}
func (*productionParkingSource) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	scope := connectors.RateLimitScopeKey("production-parking-scope")
	if req.Config.RateParkingAdmission != nil {
		if err := req.Config.RateParkingAdmission.Admit(scope); err != nil {
			return err
		}
	}
	file, err := os.OpenFile(filepath.Join(filepath.Dir(req.Config.ProjectDir), "provider-sends"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.Write([]byte{'x'}); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if os.Getenv(productionParkingHelperEnv) == "setup" && productionParkingReads.Add(1) == 2 {
		return &connsdk.RateLimitError{
			HTTPError: &connsdk.HTTPError{Status: 429, URL: "https://fixture.invalid/customers"},
			Source:    connsdk.RateLimitObservationSourceRetryAfter, HasReset: true, ResetAt: time.Now().UTC().Add(productionParkingResetDelay),
		}
	}
	return emit(connectors.Record{"id": "1", "name": "durable", "updated_at": "2026-08-15T00:00:00Z"})
}
func (*productionParkingSource) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
func (*productionParkingSource) RateLimitParkingScope(context.Context, connectors.RuntimeConfig, string, error) (connectors.RateLimitScopeKey, error) {
	return connectors.RateLimitScopeKey("production-parking-scope"), nil
}

func productionParkingSendCount(t *testing.T, path string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read provider send count: %v", err)
	}
	return len(data)
}

func TestProductionParkingReplayAndSchemaDriftRecovery(t *testing.T) {
	for _, test := range []struct {
		name           string
		originalStatus string
		mutate         func(*synccontract.CheckpointEnvelope)
		wantResume     bool
	}{
		{name: "already acknowledged item is not replayed after parking status publish crash", originalStatus: "running", wantResume: true, mutate: func(checkpoint *synccontract.CheckpointEnvelope) {
			checkpoint.Position.Primary = append(checkpoint.Position.Primary, 0x7f)
			checkpoint.ObservedAt = checkpoint.CommittedAt.Add(time.Second)
			committed := checkpoint.ObservedAt.Add(time.Second)
			checkpoint.CommittedAt = &committed
		}},
		{name: "schema drift retains parked work", originalStatus: string(coordination.RateParkingOutcomeParkedRateLimit), mutate: func(checkpoint *synccontract.CheckpointEnvelope) {
			checkpoint.SchemaVersion = "drifted-v2"
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			if err := InitProject(root); err != nil {
				t.Fatal(err)
			}
			a, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "sample", Connector: "sample"}); err != nil {
				t.Fatal(err)
			}
			if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "warehouse", Connector: "warehouse", Config: map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")}}); err != nil {
				t.Fatal(err)
			}
			if _, err := a.CreateConnection(ctx, CreateConnectionRequest{Name: "replay_guard", Source: EndpointConfig{Connector: "sample", Credential: "sample"}, Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"}, Streams: map[string]StreamConfig{"customers": {SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, CursorField: "updated_at", DestinationTable: "replay_customers"}}}); err != nil {
				t.Fatal(err)
			}
			if _, err := a.RunETL(ctx, RunETLRequest{Connection: "replay_guard", Stream: "customers"}); err != nil {
				t.Fatal(err)
			}
			key := streamStateKey("replay_guard", "customers")
			parkedCheckpoint := a.state.StreamStates[key].Checkpoint.Clone()
			current := parkedCheckpoint.Clone()
			test.mutate(&current)
			const parkedRunID = "run_replay_guard"
			if _, err := a.updateState(func(state state) (state, error) {
				stream := state.StreamStates[key]
				stream.Checkpoint = &current
				state.StreamStates[key] = stream
				state.Runs = append(state.Runs, Run{ID: parkedRunID, Type: "etl", Connection: "replay_guard", Stream: "customers", Status: test.originalStatus, StartedAt: time.Now().UTC()})
				return state, nil
			}); err != nil {
				t.Fatal(err)
			}
			a.rateParking.Close()
			store, err := coordination.OpenFileRateParkingStore(filepath.Join(root, ".polymetrics", "state", "rate-parking.json"))
			if err != nil {
				t.Fatal(err)
			}
			if _, created, err := store.Create(coordination.ParkedRateLimitRun{RunID: parkedRunID, Outcome: coordination.RateParkingOutcomeParkedRateLimit, Scope: "replay-scope", Checkpoint: parkedCheckpoint, ResetAt: time.Now().UTC().Add(-time.Second), Reason: connsdk.RateLimitObservationSourceRetryAfter}); err != nil || !created {
				t.Fatalf("create replay fixture = %t, %v", created, err)
			}
			beforeRuns := len(a.state.Runs)
			reopened, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			if len(reopened.state.Runs) != beforeRuns {
				t.Fatalf("recovery replayed an acknowledged item: runs %d -> %d", beforeRuns, len(reopened.state.Runs))
			}
			original, ok := reopened.runByID(parkedRunID)
			if !ok {
				t.Fatal("original parked run disappeared")
			}
			runs, err := store.List()
			if test.wantResume {
				if original.Status != "resumed" || err != nil || len(runs) != 0 {
					t.Fatalf("acknowledged recovery = status %q records %#v err %v", original.Status, runs, err)
				}
			} else if original.Status != string(coordination.RateParkingOutcomeParkedRateLimit) || err != nil || len(runs) != 1 {
				t.Fatalf("schema drift recovery = status %q records %#v err %v", original.Status, runs, err)
			}
		})
	}
}
