package rlm

import (
	"context"
	"fmt"
	"time"

	"polymetrics.ai/internal/connectors"
)

// DefaultFixtureRows is a hardcoded set of contact records used for
// credential-free CI runs and demos.
var DefaultFixtureRows = []connectors.Record{
	{"_polymetrics_raw_id": "fixture-1", "email": "alice@acme.com", "company": "Acme Corp", "title": "VP of Engineering"},
	{"_polymetrics_raw_id": "fixture-2", "email": "bob@beta.io", "company": "Beta Inc", "title": ""},
	{"_polymetrics_raw_id": "fixture-3", "email": "", "company": "Gamma Ltd", "title": "CTO"},
	{"_polymetrics_raw_id": "fixture-4", "email": "dana@delta.com", "company": "", "title": ""},
	{"_polymetrics_raw_id": "fixture-5", "email": "", "company": "", "title": ""},
}

// FixtureAnalyzer scores DefaultFixtureRows using the same algorithm as
// DeterministicAnalyzer but ignores InTable entirely (no file I/O needed).
type FixtureAnalyzer struct{}

// Mode returns the backend identifier.
func (f *FixtureAnalyzer) Mode() string { return "fixture" }

// Run scores DefaultFixtureRows and materializes results to req.OutTable.
// It does not read req.InTable.
func (f *FixtureAnalyzer) Run(_ context.Context, req RunRequest) (RunResult, error) {
	start := time.Now()
	result := RunResult{
		Mode:        f.Mode(),
		InTable:     req.InTable,
		OutTable:    req.OutTable,
		RecordsRead: len(DefaultFixtureRows),
		DryRun:      req.DryRun,
	}

	scored, err := scoreRecords(req.Spec, DefaultFixtureRows)
	if err != nil {
		return result, fmt.Errorf("rlm: fixture score: %w", err)
	}
	result.RecordsScored = len(scored)

	if !req.DryRun {
		now := time.Now().UTC().Format(time.RFC3339)
		if err := writeOutTable(req.WarehouseDir, req.OutTable, scored, f.Mode(), req.Spec.Name, now); err != nil {
			return result, fmt.Errorf("rlm: fixture write OutTable: %w", err)
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}
