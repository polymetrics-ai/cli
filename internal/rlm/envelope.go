package rlm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"polymetrics.ai/internal/connectors"
)

// readEnvelopedRecords reads a local warehouse NDJSON file and returns each
// record with its envelope _polymetrics_raw_id injected into the record map.
//
// The warehouse envelope is {"_polymetrics_raw_id": "...", "record": {...}}.
// Unlike readNDJSON (which discards the id), this preserves _polymetrics_raw_id
// so downstream scoring/sorting and reference-integrity checks have a stable
// per-record identity. Lines without an envelope wrapper (a bare record object)
// are passed through unchanged.
func readEnvelopedRecords(path string) ([]connectors.Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var records []connectors.Record
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var env struct {
			RawID  string            `json:"_polymetrics_raw_id"`
			Record connectors.Record `json:"record"`
		}
		if err := json.Unmarshal(line, &env); err != nil {
			return nil, fmt.Errorf("parse line: %w", err)
		}
		rec := env.Record
		if rec == nil {
			// Not an envelope — treat the whole line as a bare record.
			if err := json.Unmarshal(line, &rec); err != nil {
				return nil, fmt.Errorf("parse line: %w", err)
			}
		}
		if env.RawID != "" {
			if _, exists := rec["_polymetrics_raw_id"]; !exists {
				rec["_polymetrics_raw_id"] = env.RawID
			}
		}
		records = append(records, rec)
	}
	return records, sc.Err()
}
