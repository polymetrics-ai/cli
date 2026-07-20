package rlm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"polymetrics.ai/internal/connectors"
)

// DeterministicAnalyzer scores records using pure Go weighted-feature scoring.
// It is fully offline and reproducible: identical InTable + Spec always produces
// identical OutTable.
type DeterministicAnalyzer struct{}

// Mode returns the backend identifier.
func (d *DeterministicAnalyzer) Mode() string { return "deterministic" }

// Run scores every record in req.InTable and materializes results to req.OutTable.
func (d *DeterministicAnalyzer) Run(ctx context.Context, req RunRequest) (RunResult, error) {
	start := time.Now()
	result := RunResult{
		Mode:     d.Mode(),
		InTable:  req.InTable,
		OutTable: req.OutTable,
		DryRun:   req.DryRun,
	}

	warehouse, closeWarehouse, err := req.warehouseScope()
	if err != nil {
		return result, fmt.Errorf("rlm: open warehouse: %w", err)
	}
	if closeWarehouse {
		defer warehouse.Close()
	}

	// Read InTable beneath the held warehouse root, preserving the envelope's
	// _polymetrics_raw_id for stable sorting and reference checks.
	records, err := readWarehouseRecordsInScope(warehouse, req.InTable)
	if err != nil {
		return result, fmt.Errorf("rlm: read InTable %q: %w", req.InTable, err)
	}
	result.RecordsRead = len(records)

	scored, err := scoreRecords(req.Spec, records)
	if err != nil {
		return result, fmt.Errorf("rlm: score: %w", err)
	}
	result.RecordsScored = len(scored)

	if !req.DryRun {
		if err := validateOutTable(req.OutTable); err != nil {
			return result, err
		}
		now := time.Now().UTC().Format(time.RFC3339)
		if err := writeOutTableInScope(warehouse, req.OutTable, scored, d.Mode(), req.Spec.Name, now); err != nil {
			return result, fmt.Errorf("rlm: write OutTable: %w", err)
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// writeOutTable atomically writes scored records to OutTable NDJSON.
func writeOutTable(dir, table string, records []connectors.Record, mode, specName, scoredAt string) error {
	warehouse, err := openWarehouseScope(dir, dir)
	if err != nil {
		return err
	}
	defer warehouse.Close()
	return writeOutTableInScope(warehouse, table, records, mode, specName, scoredAt)
}

func writeOutTableInScope(warehouse *WarehouseScope, table string, records []connectors.Record, mode, specName, scoredAt string) error {
	f, tmpPath, outPath, err := warehouse.openOutputTemp(table)
	if err != nil {
		return err
	}
	removeTemp := func() error {
		err := warehouse.remove(tmpPath)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	enc := json.NewEncoder(f)
	for _, rec := range records {
		out := make(map[string]any, len(rec)+4)
		for k, v := range rec {
			out[k] = v
		}
		out["_rlm_score"] = rec["_rlm_score"]
		out["_rlm_mode"] = mode
		out["_rlm_spec"] = specName
		out["_rlm_scored_at"] = scoredAt
		if err := enc.Encode(out); err != nil {
			return errors.Join(err, f.Close(), removeTemp())
		}
	}

	if err := f.Close(); err != nil {
		return errors.Join(err, removeTemp())
	}
	if err := warehouse.commitOutput(tmpPath, outPath); err != nil {
		return errors.Join(err, removeTemp())
	}
	return nil
}

// scoreRecords applies spec feature scoring to a slice of records and returns
// the scored records with _rlm_score appended, sorted by score desc then
// _polymetrics_raw_id asc.
func scoreRecords(spec *Spec, records []connectors.Record) ([]connectors.Record, error) {
	if len(records) == 0 {
		return []connectors.Record{}, nil
	}

	// Compute total weight for normalization
	totalWeight := 0.0
	for _, f := range spec.Features {
		totalWeight += f.Weight
	}

	out := make([]connectors.Record, 0, len(records))
	for _, rec := range records {
		rawScore := 0.0
		for _, feat := range spec.Features {
			fs := featureScore(feat, rec)
			rawScore += feat.Weight * fs
		}

		normalized := 0.0
		if totalWeight > 0 {
			normalized = rawScore / totalWeight
		}

		scored := make(connectors.Record, len(rec)+1)
		for k, v := range rec {
			scored[k] = v
		}
		scored["_rlm_score"] = normalized
		out = append(out, scored)
	}

	// Sort: score desc, then _polymetrics_raw_id asc (shared helper).
	sortScored(out)

	return out, nil
}

// featureScore computes the raw score (before weight) for one feature on one record.
func featureScore(feat Feature, rec connectors.Record) float64 {
	val, exists := rec[feat.Name]

	// ScoreIfSet: field exists and is non-empty string or non-zero numeric
	if feat.ScoreIfSet != 0 {
		if exists && isNonEmpty(val) {
			return feat.ScoreIfSet
		}
		return feat.Default
	}

	// ScoreIfGT: numeric field > threshold
	if feat.ScoreIfGT != nil && feat.Threshold != nil {
		if exists {
			n := toFloat(val)
			if n > *feat.Threshold {
				return *feat.ScoreIfGT
			}
		}
		return feat.Default
	}

	return feat.Default
}

func isNonEmpty(v any) bool {
	if v == nil {
		return false
	}
	switch s := v.(type) {
	case string:
		return s != ""
	case float64:
		return s != 0
	case bool:
		return s
	default:
		return true
	}
}

func toFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}
