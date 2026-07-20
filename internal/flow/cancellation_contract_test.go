package flow

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/events"
	"polymetrics.ai/internal/telemetry"
)

func TestEngineCancellationPreservesEventsTelemetryCheckpointLedgerAndLease(t *testing.T) {
	root := t.TempDir()
	telemetryDir := filepath.Join(root, ".polymetrics", "telemetry")
	ctx, handle := telemetry.Init(context.Background(), telemetry.Config{
		Exporter:    telemetry.ExporterFile,
		ProjectRoot: root,
		Directory:   filepath.Join(".polymetrics", "telemetry"),
		RunID:       "flow-cancel-contract",
	}, func(string) {})
	collector := events.NewCollector()
	ctx = events.WithEmitter(ctx, collector)
	ctx, cancel := context.WithCancel(ctx)

	app := &cancelFlowApp{started: make(chan struct{})}
	ledger := &stubLedger{}
	checkpoint := &FileCheckpointStore{Dir: t.TempDir()}
	lockDir := t.TempDir()
	engine := &Engine{
		Manifest: FlowManifest{
			Version: 1,
			Name:    "cancel-flow",
			Steps: []FlowStep{
				{ID: "blocked", Kind: KindSync, Connection: "local", Streams: []string{"records"}, Out: []string{"records"}},
				{ID: "must-not-run", Kind: KindQuery, SQL: "SELECT * FROM records", In: []string{"records"}},
			},
		},
		App:        app,
		Ledger:     ledger,
		Checkpoint: checkpoint,
		LockDir:    lockDir,
	}

	type outcome struct {
		result RunResult
		err    error
	}
	done := make(chan outcome, 1)
	go func() {
		result, err := engine.Run(ctx, RunOptions{})
		done <- outcome{result: result, err: err}
	}()

	select {
	case <-app.started:
		cancel()
	case <-time.After(5 * time.Second):
		t.Fatal("flow step did not start")
	}

	var got outcome
	select {
	case got = <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("cancelled flow did not return")
	}
	if got.err == nil || !strings.Contains(got.err.Error(), context.Canceled.Error()) {
		t.Fatalf("Run error = %v, want cancellation", got.err)
	}
	if got.result.Status != "failed" || len(got.result.Steps) != 1 || got.result.Steps[0].Status != "failed" {
		t.Fatalf("cancel result = %#v", got.result)
	}
	if app.queryCalls != 0 {
		t.Fatalf("later query calls = %d, want 0", app.queryCalls)
	}
	if status, err := checkpoint.Get("cancel-flow", "blocked"); err != nil || status != "" {
		t.Fatalf("cancelled step checkpoint = %q, err=%v", status, err)
	}
	if _, err := os.Stat(filepath.Join(lockDir, "flow-cancel-flow.lock")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("flow lease remains after cancellation: %v", err)
	}

	wantEvents := []string{
		"flow::started:running",
		"flow:blocked:started:running",
		"flow:blocked:failed:failed",
		"flow::failed:failed",
	}
	assertStringSlice(t, eventSequence(collector.Events()), wantEvents)

	records := ledger.all()
	wantLedger := []LedgerRecord{
		{Mode: "flow", Operation: "cancel-flow/blocked", Status: "running"},
		{Mode: "flow", Operation: "cancel-flow/blocked", Status: "failed", Error: context.Canceled.Error()},
	}
	if len(records) < 3 {
		t.Fatalf("ledger records = %#v, want step running/failed and flow failed", records)
	}
	for i, want := range wantLedger {
		if records[i] != want {
			t.Fatalf("ledger[%d] = %#v, want %#v", i, records[i], want)
		}
	}
	last := records[len(records)-1]
	if last.Mode != "flow" || last.Operation != "cancel-flow" || last.Status != "failed" || !strings.Contains(last.Error, context.Canceled.Error()) {
		t.Fatalf("final ledger record = %#v", last)
	}

	telemetry.Shutdown(context.Background(), handle, func(string) {})
	data := readFlowTelemetry(t, telemetryDir)
	for _, needle := range []string{"pm.flow.run", "pm.flow.step", "context_canceled", "pm.stage.duration", "sync"} {
		if !bytes.Contains(data, []byte(needle)) {
			t.Fatalf("telemetry missing %q:\n%s", needle, data)
		}
	}
}

type cancelFlowApp struct {
	started    chan struct{}
	queryCalls int
}

func (a *cancelFlowApp) ETLRun(ctx context.Context, _ string, _ []string) (ETLResult, error) {
	close(a.started)
	<-ctx.Done()
	return ETLResult{}, ctx.Err()
}

func (a *cancelFlowApp) QuerySQL(context.Context, string, int) ([]map[string]any, error) {
	a.queryCalls++
	return nil, nil
}

func (a *cancelFlowApp) RLMRun(context.Context, RLMRunRequest) (RLMResult, error) {
	return RLMResult{}, nil
}
