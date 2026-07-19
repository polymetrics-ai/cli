package rlm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/safety"
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
	defer f.Close()
	return readEnvelopedRecordsFrom(f)
}

func readWarehouseRecords(dir, table string) ([]connectors.Record, error) {
	fs, f, err := openWarehouseTable(dir, table)
	if err != nil {
		return nil, err
	}
	defer fs.Close()
	defer f.Close()
	return readEnvelopedRecordsFrom(f)
}

func readWarehouseTable(dir, table string) ([]byte, []connectors.Record, error) {
	fs, f, err := openWarehouseTable(dir, table)
	if err != nil {
		return nil, nil, err
	}
	defer fs.Close()
	data, err := io.ReadAll(f)
	closeErr := f.Close()
	if err != nil {
		return nil, nil, err
	}
	if closeErr != nil {
		return nil, nil, closeErr
	}
	records, err := readEnvelopedRecordsFrom(bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	return data, records, nil
}

func openWarehouseTable(dir, table string) (*safety.LocalWriteFS, *os.File, error) {
	if err := validateInTable(table); err != nil {
		return nil, nil, err
	}
	fs, err := safety.OpenLocalWriteFS(dir, false)
	if err != nil {
		return nil, nil, err
	}
	f, err := fs.Open(table + ".ndjson")
	if err != nil {
		_ = fs.Close()
		return nil, nil, err
	}
	return fs, f, nil
}

func readEnvelopedRecordsFrom(r io.Reader) ([]connectors.Record, error) {
	var records []connectors.Record
	sc := bufio.NewScanner(r)
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
