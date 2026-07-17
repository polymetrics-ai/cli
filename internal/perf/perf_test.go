package perf

import (
	"context"
	"errors"
	"strings"
	"testing"

	pmlogging "polymetrics.ai/internal/logging"
	"polymetrics.ai/internal/runtimecheck"
)

func TestCompareDependencyFree(t *testing.T) {
	comparison, err := Compare(context.Background(), CompareRequest{Iterations: 2})
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if comparison.DependencyFree.Mode != "dependency-free" {
		t.Fatalf("mode = %q", comparison.DependencyFree.Mode)
	}
	if comparison.DependencyFree.Records != 6 {
		t.Fatalf("records = %d, want 6", comparison.DependencyFree.Records)
	}
	if comparison.Explanation["dependency_free"] == "" {
		t.Fatalf("missing dependency-free explanation")
	}
}

func TestCompareRejectsOversizedIterations(t *testing.T) {
	_, err := Compare(context.Background(), CompareRequest{Iterations: MaxCompareIterations + 1})
	if err == nil {
		t.Fatal("Compare() oversized iterations error = nil")
	}
	if !strings.Contains(err.Error(), "iterations") || !strings.Contains(err.Error(), "max") {
		t.Fatalf("Compare() oversized error = %q, want iterations max guard", err.Error())
	}
}

func TestCompareSyncModesRejectsOversizedRecords(t *testing.T) {
	_, err := CompareSyncModes(context.Background(), SyncModeBenchmarkRequest{Records: MaxSyncModeRecords + 1})
	if err == nil {
		t.Fatal("CompareSyncModes() oversized records error = nil")
	}
	if !strings.Contains(err.Error(), "records") || !strings.Contains(err.Error(), "max") {
		t.Fatalf("CompareSyncModes() oversized error = %q, want records max guard", err.Error())
	}
}

func TestCompareRedactsRuntimeBackedError(t *testing.T) {
	oldDependencyFree := dependencyFreeRunner
	oldRuntimeBacked := runtimeBackedRunner
	oldRuntimeDoctor := runtimeDoctor
	t.Cleanup(func() {
		dependencyFreeRunner = oldDependencyFree
		runtimeBackedRunner = oldRuntimeBacked
		runtimeDoctor = oldRuntimeDoctor
	})

	const secret = "pm-test-runtime-redaction-secret-423"
	registry := pmlogging.NewValueRegistry()
	registry.Register(secret)
	ctx := pmlogging.WithRegistry(context.Background(), registry)

	dependencyFreeRunner = func(context.Context, int) (Result, error) {
		return Result{Mode: "dependency-free", Iterations: 1, Records: 3}, nil
	}
	runtimeDoctor = func(context.Context, runtimecheck.Config) runtimecheck.Report {
		return runtimecheck.Report{Mode: "runtime", Checks: []runtimecheck.CheckResult{
			{Name: "postgres", Status: "ok"},
			{Name: "dragonfly", Status: "ok"},
			{Name: "temporal", Status: "ok"},
		}}
	}
	runtimeBackedRunner = func(context.Context, int, runtimecheck.Config) (Result, error) {
		return Result{Mode: "runtime-backed", Iterations: 1}, errors.New("runtime write failed with token " + secret)
	}

	comparison, err := Compare(ctx, CompareRequest{Iterations: 1, Runtime: true})
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if comparison.RuntimeBacked == nil {
		t.Fatal("Compare() RuntimeBacked = nil")
	}
	if strings.Contains(comparison.RuntimeBacked.Error, secret) {
		t.Fatalf("RuntimeBacked.Error leaked registered secret: %q", comparison.RuntimeBacked.Error)
	}
	if !strings.Contains(comparison.RuntimeBacked.Error, "[redacted]") {
		t.Fatalf("RuntimeBacked.Error = %q, want redaction marker", comparison.RuntimeBacked.Error)
	}
}
