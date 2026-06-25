package perf

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/coordination"
	"polymetrics.ai/internal/ledger"
	"polymetrics.ai/internal/runtimecheck"
)

type CompareRequest struct {
	Iterations int  `json:"iterations"`
	Runtime    bool `json:"runtime"`
}

type Result struct {
	Mode          string        `json:"mode"`
	Iterations    int           `json:"iterations"`
	Records       int           `json:"records"`
	Duration      time.Duration `json:"duration"`
	Average       time.Duration `json:"average"`
	RecordsPerSec float64       `json:"records_per_sec"`
	Error         string        `json:"error,omitempty"`
}

type Comparison struct {
	DependencyFree Result               `json:"dependency_free"`
	RuntimeBacked  *Result              `json:"runtime_backed,omitempty"`
	RuntimeReport  *runtimecheck.Report `json:"runtime_report,omitempty"`
	Explanation    map[string]string    `json:"explanation"`
}

type SyncModeBenchmarkRequest struct {
	Records int `json:"records"`
}

type SyncModeBenchmarkResult struct {
	Mode          string        `json:"mode"`
	Records       int           `json:"records"`
	Duration      time.Duration `json:"duration"`
	RecordsPerSec float64       `json:"records_per_sec"`
	Error         string        `json:"error,omitempty"`
}

type SyncModeBenchmark struct {
	Records     int                       `json:"records"`
	Results     []SyncModeBenchmarkResult `json:"results"`
	Explanation string                    `json:"explanation"`
}

func Compare(ctx context.Context, req CompareRequest) (Comparison, error) {
	if req.Iterations <= 0 {
		req.Iterations = 25
	}
	free, err := runDependencyFree(ctx, req.Iterations)
	if err != nil {
		return Comparison{}, err
	}
	comparison := Comparison{
		DependencyFree: free,
		Explanation: map[string]string{
			"dependency_free": "Uses local JSON state, AES-GCM file vault, JSONL warehouse/outbox, and in-process ETL/reverse ETL. It has no database, cache, or workflow server requirement.",
			"runtime_backed":  "Uses the same local ETL path plus external runtime checks, Dragonfly lease coordination, and PostgreSQL run-ledger writes. Temporal is health-checked as the durable workflow target.",
		},
	}
	if !req.Runtime {
		return comparison, nil
	}
	cfg := runtimecheck.FromEnv()
	report := runtimecheck.Doctor(ctx, cfg)
	comparison.RuntimeReport = &report
	if !runtimecheck.Healthy(report) {
		result := Result{Mode: "runtime-backed", Iterations: req.Iterations, Error: "runtime services are not all healthy"}
		comparison.RuntimeBacked = &result
		return comparison, nil
	}
	runtimeResult, err := runRuntimeBacked(ctx, req.Iterations, cfg)
	if err != nil {
		runtimeResult.Error = err.Error()
		comparison.RuntimeBacked = &runtimeResult
		return comparison, nil
	}
	comparison.RuntimeBacked = &runtimeResult
	return comparison, nil
}

func CompareSyncModes(ctx context.Context, req SyncModeBenchmarkRequest) (SyncModeBenchmark, error) {
	if req.Records <= 0 {
		req.Records = 1000
	}
	modes := []string{
		"full_refresh_append",
		"full_refresh_overwrite",
		"full_refresh_overwrite_deduped",
		"incremental_append",
		"incremental_append_deduped",
	}
	results := make([]SyncModeBenchmarkResult, 0, len(modes))
	for _, mode := range modes {
		result := runSyncModeBenchmark(ctx, mode, req.Records)
		results = append(results, result)
	}
	return SyncModeBenchmark{
		Records: req.Records,
		Results: results,
		Explanation: "Synthetic dependency-free benchmark using local JSONL file source and local JSONL warehouse destination. " +
			"Deduped modes include raw history and final materialization cost.",
	}, nil
}

func runSyncModeBenchmark(ctx context.Context, mode string, records int) SyncModeBenchmarkResult {
	root := filepath.Join(os.TempDir(), fmt.Sprintf("pm-sync-mode-%s-%d", mode, time.Now().UnixNano()))
	start := time.Now()
	loaded, err := runFileToWarehouse(ctx, root, mode, records)
	duration := time.Since(start)
	result := SyncModeBenchmarkResult{Mode: mode, Records: loaded, Duration: duration}
	if duration > 0 {
		result.RecordsPerSec = float64(loaded) / duration.Seconds()
	}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}

func runFileToWarehouse(ctx context.Context, root, mode string, records int) (int, error) {
	if err := app.InitProject(root); err != nil {
		return 0, err
	}
	sourcePath := filepath.Join(root, "source.jsonl")
	if err := writeSyntheticSource(sourcePath, records); err != nil {
		return 0, err
	}
	a, err := app.Open(root)
	if err != nil {
		return 0, err
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "file-local",
		Connector: "file",
		Config:    map[string]string{"path": sourcePath, "stream": "records"},
	}); err != nil {
		return 0, err
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "warehouse-local",
		Connector: "warehouse",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		return 0, err
	}
	if _, err := a.CreateConnection(ctx, app.CreateConnectionRequest{
		Name:        "file_to_warehouse",
		Source:      app.EndpointConfig{Connector: "file", Credential: "file-local"},
		Destination: app.EndpointConfig{Connector: "warehouse", Credential: "warehouse-local"},
		Streams: map[string]app.StreamConfig{
			"records": {SyncMode: mode, CursorField: "updated_at", PrimaryKey: []string{"id"}, DestinationTable: "records"},
		},
	}); err != nil {
		return 0, err
	}
	run, err := a.RunETL(ctx, app.RunETLRequest{Connection: "file_to_warehouse", Stream: "records", BatchSize: 500})
	if err != nil {
		return run.RecordsLoaded, err
	}
	return run.RecordsLoaded, nil
}

func writeSyntheticSource(path string, records int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < records; i++ {
		record := map[string]any{
			"id":         fmt.Sprintf("rec_%06d", i),
			"name":       fmt.Sprintf("Record %06d", i),
			"updated_at": base.Add(time.Duration(i) * time.Second).Format(time.RFC3339Nano),
		}
		if err := encoder.Encode(record); err != nil {
			return err
		}
	}
	return nil
}

func runDependencyFree(ctx context.Context, iterations int) (Result, error) {
	root := filepath.Join(os.TempDir(), fmt.Sprintf("pm-perf-free-%d", time.Now().UnixNano()))
	start := time.Now()
	records, err := runLocalETLLoop(ctx, root, iterations)
	if err != nil {
		return Result{}, err
	}
	duration := time.Since(start)
	return summarize("dependency-free", iterations, records, duration), nil
}

func runRuntimeBacked(ctx context.Context, iterations int, cfg runtimecheck.Config) (Result, error) {
	root := filepath.Join(os.TempDir(), fmt.Sprintf("pm-perf-runtime-%d", time.Now().UnixNano()))
	start := time.Now()
	dragonfly := coordination.OpenDragonfly(cfg.DragonflyAddr)
	defer dragonfly.Close()
	if err := dragonfly.Ping(ctx); err != nil {
		return Result{Mode: "runtime-backed", Iterations: iterations}, err
	}
	leaseKey := fmt.Sprintf("polymetrics:perf:%d", time.Now().UnixNano())
	acquired, err := dragonfly.AcquireLease(ctx, leaseKey, "running", 30*time.Second)
	if err != nil {
		return Result{Mode: "runtime-backed", Iterations: iterations}, err
	}
	if !acquired {
		return Result{Mode: "runtime-backed", Iterations: iterations}, fmt.Errorf("runtime performance lease was not acquired")
	}
	defer dragonfly.ReleaseLease(ctx, leaseKey)

	pg, err := ledger.OpenPostgres(ctx, cfg.PostgresURL)
	if err != nil {
		return Result{Mode: "runtime-backed", Iterations: iterations}, err
	}
	defer pg.Close()
	if err := pg.Migrate(ctx); err != nil {
		return Result{Mode: "runtime-backed", Iterations: iterations}, err
	}
	records, err := runLocalETLLoop(ctx, root, iterations)
	if err != nil {
		return Result{Mode: "runtime-backed", Iterations: iterations}, err
	}
	duration := time.Since(start)
	result := summarize("runtime-backed", iterations, records, duration)
	if err := pg.Append(ctx, ledger.RunRecord{
		ID:             fmt.Sprintf("perf_%d", time.Now().UnixNano()),
		Mode:           "runtime-backed",
		Operation:      "perf_compare",
		Status:         "completed",
		RecordsRead:    records,
		RecordsWritten: records,
		Duration:       int64(duration),
		CreatedAt:      time.Now().UTC(),
	}); err != nil {
		return result, err
	}
	return result, nil
}

func runLocalETLLoop(ctx context.Context, root string, iterations int) (int, error) {
	if err := app.InitProject(root); err != nil {
		return 0, err
	}
	a, err := app.Open(root)
	if err != nil {
		return 0, err
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{Name: "sample-local", Connector: "sample", Secrets: map[string]string{"token": "local"}}); err != nil {
		return 0, err
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{Name: "warehouse-local", Connector: "warehouse", Config: map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")}}); err != nil {
		return 0, err
	}
	if _, err := a.CreateConnection(ctx, app.CreateConnectionRequest{
		Name:        "sample_to_warehouse",
		Source:      app.EndpointConfig{Connector: "sample", Credential: "sample-local"},
		Destination: app.EndpointConfig{Connector: "warehouse", Credential: "warehouse-local"},
		Streams: map[string]app.StreamConfig{
			"customers": {SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, DestinationTable: "sample_customers"},
		},
	}); err != nil {
		return 0, err
	}
	records := 0
	for i := 0; i < iterations; i++ {
		run, err := a.RunETL(ctx, app.RunETLRequest{Connection: "sample_to_warehouse", Stream: "customers"})
		if err != nil {
			return records, err
		}
		records += run.RecordsLoaded
	}
	return records, nil
}

func summarize(mode string, iterations, records int, duration time.Duration) Result {
	average := time.Duration(0)
	if iterations > 0 {
		average = duration / time.Duration(iterations)
	}
	rps := 0.0
	if duration > 0 {
		rps = float64(records) / duration.Seconds()
	}
	return Result{Mode: mode, Iterations: iterations, Records: records, Duration: duration, Average: average, RecordsPerSec: rps}
}
