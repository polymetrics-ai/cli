// Package rlm implements the Record-Level Model (RLM) scoring pipeline.
// It provides deterministic, fixture, and (stub) model backends for scoring
// warehouse records against a JSON spec.
package rlm

import (
	"context"
	"errors"
	"time"
)

// ErrNotImplemented is returned by backends that have not yet been implemented.
// The model backend always returns this sentinel until Phase 4 is approved.
var ErrNotImplemented = errors.New("rlm: model backend not implemented (requires Phase 4 approval)")

// Analyzer is the strategy interface for all RLM backends.
// Implementations must be safe for concurrent use from a single goroutine per Run call.
type Analyzer interface {
	// Run scores every record in req.InTable and materializes results to req.OutTable.
	// It must be deterministic: identical InTable + Spec always produces identical OutTable.
	Run(ctx context.Context, req RunRequest) (RunResult, error)
	// Mode returns the backend identifier ("deterministic", "fixture", "model").
	Mode() string
}

// RunRequest carries all inputs for a single RLM scoring run.
type RunRequest struct {
	Spec         *Spec           // parsed scoring spec
	InTable      string          // source warehouse table name (no path — resolved via WarehouseDir)
	OutTable     string          // destination warehouse table name
	WarehouseDir string          // resolved from app runtime config
	Warehouse    *WarehouseScope // optional held project-root scope for all table effects
	DryRun       bool            // if true, compute scores but do not write OutTable
}

// RunResult is the machine-readable result of a scoring run.
type RunResult struct {
	Mode          string        `json:"mode"`
	InTable       string        `json:"in_table"`
	OutTable      string        `json:"out_table"`
	RecordsRead   int           `json:"records_read"`
	RecordsScored int           `json:"records_scored"`
	RecordsFailed int           `json:"records_failed"`
	Duration      time.Duration `json:"duration_ns"`
	DryRun        bool          `json:"dry_run"`
}
