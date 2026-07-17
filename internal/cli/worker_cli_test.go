package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/worker"
)

func TestWorkerServeJSONStartupFailureEmitsOnlyError(t *testing.T) {
	root := writeMigrationConfig(t, `runtime:
  temporal_addr: 127.0.0.1:7233
`)
	origServe := workerServe
	workerServe = func(_ context.Context, _ string, _ *worker.PodmanActivities, _ func()) error {
		return errors.New("startup failed")
	}
	t.Cleanup(func() { workerServe = origServe })

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", root, "--json", "worker", "serve"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("worker serve startup failure exit = 0, want non-zero")
	}
	if strings.Contains(stdout.String(), "WorkerServe") {
		t.Fatalf("worker serve emitted start envelope before readiness")
	}
	assertSingleWorkerJSONEnvelopeKind(t, stdout.String(), "Error")
}

func TestWorkerServeJSONStartsOnlyAfterReady(t *testing.T) {
	root := writeMigrationConfig(t, `runtime:
  temporal_addr: 127.0.0.1:7233
`)
	origServe := workerServe
	var stdout bytes.Buffer
	workerServe = func(ctx context.Context, _ string, _ *worker.PodmanActivities, ready func()) error {
		if ctx == nil {
			t.Fatal("worker serve got nil context")
		}
		if stdout.Len() != 0 {
			t.Fatal("worker serve wrote stdout before readiness callback")
		}
		ready()
		if stdout.Len() == 0 {
			t.Fatal("worker serve did not emit start envelope after readiness callback")
		}
		return nil
	}
	t.Cleanup(func() { workerServe = origServe })

	var stderr bytes.Buffer
	code := Run([]string{"--root", root, "--json", "worker", "serve"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	assertSingleWorkerJSONEnvelopeKind(t, stdout.String(), "WorkerServe")
}

func TestWorkerServePlainStartsOnlyAfterReady(t *testing.T) {
	root := writeMigrationConfig(t, `runtime:
  temporal_addr: 127.0.0.1:7233
`)
	origServe := workerServe
	var stdout bytes.Buffer
	workerServe = func(_ context.Context, _ string, _ *worker.PodmanActivities, ready func()) error {
		if stdout.Len() != 0 {
			t.Fatal("worker serve wrote plain stdout before readiness callback")
		}
		ready()
		return nil
	}
	t.Cleanup(func() { workerServe = origServe })

	var stderr bytes.Buffer
	code := Run([]string{"--root", root, "worker", "serve"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "pm worker serving") {
		t.Fatalf("plain worker serve missing ready output")
	}
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
