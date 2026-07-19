package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/config"
)

func TestWorkerServeJSONStartupFailureEmitsOnlyError(t *testing.T) {
	cfg := workerTestConfig(t.TempDir(), true, "127.0.0.1:7233")
	runtime := newFakeWorkerRuntime()
	runtime.serveErr = errors.New("startup failed")

	stdout, _, code := runWorkerCommand(ctxForWorkerTest(), cfg, runtime, "worker", "serve")
	if code == 0 {
		t.Fatal("worker serve startup failure exit = 0, want non-zero")
	}
	if strings.Contains(stdout, "WorkerServe") {
		t.Fatalf("worker serve emitted start envelope before readiness")
	}
	assertSingleWorkerJSONEnvelopeKind(t, stdout, "Error")
}

func TestWorkerServeJSONStartsOnlyAfterReady(t *testing.T) {
	cfg := workerTestConfig(t.TempDir(), true, "127.0.0.1:7233")
	runtime := newFakeWorkerRuntime()

	stdout, stderr, code := runWorkerCommand(ctxForWorkerTest(), cfg, runtime, "worker", "serve")
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}
	if runtime.readyCalls != 1 {
		t.Fatalf("ready calls = %d, want 1", runtime.readyCalls)
	}
	assertSingleWorkerJSONEnvelopeKind(t, stdout, "WorkerServe")
}

func TestWorkerServePlainStartsOnlyAfterReady(t *testing.T) {
	cfg := workerTestConfig(t.TempDir(), false, "127.0.0.1:7233")
	runtime := newFakeWorkerRuntime()

	stdout, stderr, code := runWorkerCommand(ctxForWorkerTest(), cfg, runtime, "worker", "serve")
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}
	if runtime.readyCalls != 1 {
		t.Fatalf("ready calls = %d, want 1", runtime.readyCalls)
	}
	if !strings.Contains(stdout, "pm worker serving") {
		t.Fatalf("plain worker serve missing ready output")
	}
}

func ctxForWorkerTest() context.Context {
	return context.Background()
}

func runWorkerCommand(ctx context.Context, cfg config.Config, runtime *fakeWorkerRuntime, args ...string) (string, string, int) {
	var stdout, stderr bytes.Buffer
	cmd := newRootCmdWithWorkerRuntime(ctx, cfg, &stdout, &stderr, runtime.runtime())
	if err := executeRootCmd(cmd, args); err != nil {
		code := writeError(ctx, &stdout, &stderr, mapCobraErr(err), cfg.JSON)
		return stdout.String(), stderr.String(), code
	}
	return stdout.String(), stderr.String(), 0
}

func assertSingleWorkerJSONEnvelopeKind(t *testing.T, stdout, wantKind string) {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(stdout))
	var env map[string]any
	if err := dec.Decode(&env); err != nil {
		t.Fatalf("decode worker serve envelope: %v", err)
	}
	var extra map[string]any
	if err := dec.Decode(&extra); err == nil {
		t.Fatalf("worker serve emitted more than one JSON envelope")
	}
	if env["kind"] != wantKind {
		t.Fatalf("worker serve envelope kind = %v, want %s", env["kind"], wantKind)
	}
}
