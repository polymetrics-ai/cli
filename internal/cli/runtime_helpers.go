package cli

import (
	"context"
	"fmt"
	"time"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
	pmruntime "polymetrics.ai/internal/runtime"
	"polymetrics.ai/internal/runtimecheck"
)

func runtimeETLLeaseRequest(run app.Run) pmruntime.LeaseRequest {
	return pmruntime.LeaseRequest{Key: "polymetrics:etl:" + run.ID, Value: "recording", TTL: 30 * time.Second}
}

func runtimeETLRunRecord(run app.Run) pmruntime.RunRecord {
	return pmruntime.RunRecord{
		ID:             run.ID,
		Mode:           "runtime-backed",
		Operation:      "etl",
		Status:         run.Status,
		RecordsRead:    run.RecordsRead,
		RecordsWritten: run.RecordsLoaded,
		Duration:       run.CompletedAt.Sub(run.StartedAt).Nanoseconds(),
		CreatedAt:      run.StartedAt,
	}
}

func recordRuntimeETL(ctx context.Context, run app.Run, cfg config.Config) error {
	runtimeCfg := runtimecheck.FromConfig(cfg)
	report := runtimecheck.Doctor(ctx, runtimeCfg)
	if !runtimecheck.Healthy(report) {
		return fmt.Errorf("runtime dependencies are not healthy; run `pm runtime doctor --json` for details")
	}
	dragonfly := pmruntime.OpenDragonflyLeaseStore(runtimeCfg.DragonflyAddr)
	defer dragonfly.Close()
	pg, err := pmruntime.OpenPostgresRunLedger(ctx, runtimeCfg.PostgresURL)
	if err != nil {
		return err
	}
	defer pg.Close()
	if err := pg.Migrate(ctx); err != nil {
		return err
	}
	module := pmruntime.Module{Leases: dragonfly, Ledger: pg}
	return module.RecordRunWithLease(ctx, runtimeETLLeaseRequest(run), runtimeETLRunRecord(run))
}
