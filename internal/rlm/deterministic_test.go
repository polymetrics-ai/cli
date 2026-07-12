package rlm

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/connectors"
)

// makeSimpleSpec returns a two-feature spec used in scoring unit tests.
// email: weight=0.5, score_if_set=1.0
// company: weight=0.5, score_if_set=1.0
func makeSimpleSpec() *Spec {
	return &Spec{
		Name: "test-spec",
		Features: []Feature{
			{Name: "email", Weight: 0.5, ScoreIfSet: 1.0},
			{Name: "company", Weight: 0.5, ScoreIfSet: 1.0},
		},
	}
}

// --- Determinism ---

func TestDeterminismSameInputSameOutput(t *testing.T) {
	spec := makeSimpleSpec()
	records := []connectors.Record{
		{"_polymetrics_raw_id": "r1", "email": "a@b.com", "company": "Acme"},
		{"_polymetrics_raw_id": "r2", "email": "c@d.com"},
		{"_polymetrics_raw_id": "r3", "company": "Beta"},
	}

	first, err := scoreRecords(spec, records)
	if err != nil {
		t.Fatalf("first run error: %v", err)
	}
	second, err := scoreRecords(spec, records)
	if err != nil {
		t.Fatalf("second run error: %v", err)
	}

	if len(first) != len(second) {
		t.Fatalf("length mismatch: %d vs %d", len(first), len(second))
	}
	for i := range first {
		s1 := first[i]["_rlm_score"]
		s2 := second[i]["_rlm_score"]
		if s1 != s2 {
			t.Errorf("row %d: score mismatch %v vs %v", i, s1, s2)
		}
		id1 := first[i]["_polymetrics_raw_id"]
		id2 := second[i]["_polymetrics_raw_id"]
		if id1 != id2 {
			t.Errorf("row %d: id mismatch %v vs %v", i, id1, id2)
		}
	}
}

// --- Weighted sum ---

func TestScoringWeightedSum(t *testing.T) {
	// email present (score_if_set=1.0, weight=0.5) → contributes 0.5
	// company absent → contributes 0.0
	// normalized score = (0.5*1.0 + 0.5*0.0) / (0.5+0.5) = 0.5
	spec := makeSimpleSpec()
	records := []connectors.Record{
		{"_polymetrics_raw_id": "r1", "email": "a@b.com"},
	}
	out, err := scoreRecords(spec, records)
	if err != nil {
		t.Fatalf("scoreRecords error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 record, got %d", len(out))
	}
	score, ok := out[0]["_rlm_score"].(float64)
	if !ok {
		t.Fatalf("_rlm_score is not float64: %T %v", out[0]["_rlm_score"], out[0]["_rlm_score"])
	}
	const want = 0.5
	if score != want {
		t.Errorf("score = %f, want %f", score, want)
	}
}

// --- ScoreIfSet ---

func TestScoringScoreIfSet_Present(t *testing.T) {
	spec := &Spec{
		Name: "test",
		Features: []Feature{
			{Name: "email", Weight: 1.0, ScoreIfSet: 0.9},
		},
	}
	out, err := scoreRecords(spec, []connectors.Record{
		{"_polymetrics_raw_id": "r1", "email": "x@y.com"},
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	score := out[0]["_rlm_score"].(float64)
	if score != 0.9 {
		t.Errorf("score = %f, want 0.9", score)
	}
}

func TestScoringScoreIfSet_Absent(t *testing.T) {
	spec := &Spec{
		Name: "test",
		Features: []Feature{
			{Name: "email", Weight: 1.0, ScoreIfSet: 0.9, Default: 0.1},
		},
	}
	out, err := scoreRecords(spec, []connectors.Record{
		{"_polymetrics_raw_id": "r1", "email": ""},
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	score := out[0]["_rlm_score"].(float64)
	if score != 0.1 {
		t.Errorf("score = %f, want 0.1", score)
	}
}

// --- ScoreIfGT ---

func TestScoringScoreIfGT_Above(t *testing.T) {
	threshold := 5.0
	scoreVal := 1.0
	spec := &Spec{
		Name: "test",
		Features: []Feature{
			{Name: "amount", Weight: 1.0, ScoreIfGT: &scoreVal, Threshold: &threshold},
		},
	}
	out, err := scoreRecords(spec, []connectors.Record{
		{"_polymetrics_raw_id": "r1", "amount": 10.0},
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	score := out[0]["_rlm_score"].(float64)
	if score != 1.0 {
		t.Errorf("score = %f, want 1.0", score)
	}
}

func TestScoringScoreIfGT_Below(t *testing.T) {
	threshold := 5.0
	scoreVal := 1.0
	spec := &Spec{
		Name: "test",
		Features: []Feature{
			{Name: "amount", Weight: 1.0, ScoreIfGT: &scoreVal, Threshold: &threshold, Default: 0.0},
		},
	}
	out, err := scoreRecords(spec, []connectors.Record{
		{"_polymetrics_raw_id": "r1", "amount": 2.0},
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	score := out[0]["_rlm_score"].(float64)
	if score != 0.0 {
		t.Errorf("score = %f, want 0.0", score)
	}
}

// --- Edge cases ---

func TestScoringAllZeroWeights(t *testing.T) {
	spec := &Spec{
		Name: "test",
		Features: []Feature{
			{Name: "email", Weight: 0.0, ScoreIfSet: 1.0},
		},
	}
	out, err := scoreRecords(spec, []connectors.Record{
		{"_polymetrics_raw_id": "r1", "email": "a@b.com"},
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	score := out[0]["_rlm_score"].(float64)
	if score != 0.0 {
		t.Errorf("score = %f, want 0.0", score)
	}
}

func TestScoringEmptyRecordSet(t *testing.T) {
	spec := makeSimpleSpec()
	out, err := scoreRecords(spec, []connectors.Record{})
	if err != nil {
		t.Fatalf("error on empty input: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected 0 records, got %d", len(out))
	}
}

// --- Sorting ---

func TestSortingByScoreDesc(t *testing.T) {
	spec := makeSimpleSpec()
	records := []connectors.Record{
		{"_polymetrics_raw_id": "r1"},                                     // score=0
		{"_polymetrics_raw_id": "r2", "email": "a@b.com"},                 // score=0.5
		{"_polymetrics_raw_id": "r3", "email": "a@b.com", "company": "X"}, // score=1.0
	}
	out, err := scoreRecords(spec, records)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if out[0]["_polymetrics_raw_id"] != "r3" {
		t.Errorf("row 0 should be r3 (score=1.0), got %v", out[0]["_polymetrics_raw_id"])
	}
	if out[1]["_polymetrics_raw_id"] != "r2" {
		t.Errorf("row 1 should be r2 (score=0.5), got %v", out[1]["_polymetrics_raw_id"])
	}
	if out[2]["_polymetrics_raw_id"] != "r1" {
		t.Errorf("row 2 should be r1 (score=0.0), got %v", out[2]["_polymetrics_raw_id"])
	}
}

func TestSortingTiebreakerByRawID(t *testing.T) {
	// Both have same score; tie broken by _polymetrics_raw_id asc
	spec := makeSimpleSpec()
	records := []connectors.Record{
		{"_polymetrics_raw_id": "r2", "email": "a@b.com"},
		{"_polymetrics_raw_id": "r1", "email": "x@y.com"},
	}
	out, err := scoreRecords(spec, records)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if out[0]["_polymetrics_raw_id"] != "r1" {
		t.Errorf("tie should be broken by raw id asc: row 0 = %v", out[0]["_polymetrics_raw_id"])
	}
}

// --- Materialization ---

func writeInTableNDJSON(t *testing.T, dir, table string, rows []connectors.Record) string {
	t.Helper()
	path := filepath.Join(dir, table+".ndjson")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create InTable: %v", err)
	}
	defer func() { _ = f.Close() }()
	enc := json.NewEncoder(f)
	for _, row := range rows {
		// Wrap in localRawRecord format as written by internal/app
		wrapped := map[string]any{
			"_polymetrics_raw_id": row["_polymetrics_raw_id"],
			"record":              row,
		}
		if err := enc.Encode(wrapped); err != nil {
			t.Fatalf("encode row: %v", err)
		}
	}
	return path
}

func TestMaterializationWritesNDJSON(t *testing.T) {
	dir := t.TempDir()
	rows := []connectors.Record{
		{"_polymetrics_raw_id": "r1", "email": "a@b.com", "company": "Acme"},
		{"_polymetrics_raw_id": "r2", "email": "b@c.com"},
		{"_polymetrics_raw_id": "r3", "company": "Beta"},
	}
	writeInTableNDJSON(t, dir, "contacts", rows)

	spec := makeSimpleSpec()
	req := RunRequest{
		Spec:         spec,
		InTable:      "contacts",
		OutTable:     "lead_scores",
		WarehouseDir: dir,
	}

	a := &DeterministicAnalyzer{}
	result, err := a.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if result.RecordsScored != 3 {
		t.Errorf("RecordsScored = %d, want 3", result.RecordsScored)
	}

	outPath := filepath.Join(dir, "lead_scores.ndjson")
	f, err := os.Open(outPath)
	if err != nil {
		t.Fatalf("OutTable not written: %v", err)
	}
	defer func() { _ = f.Close() }()

	var outRows []map[string]any
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var row map[string]any
		if err := json.Unmarshal(sc.Bytes(), &row); err != nil {
			t.Fatalf("unmarshal row: %v", err)
		}
		outRows = append(outRows, row)
	}

	if len(outRows) != 3 {
		t.Fatalf("OutTable row count = %d, want 3", len(outRows))
	}

	requiredFields := []string{"_rlm_score", "_rlm_mode", "_rlm_spec", "_rlm_scored_at"}
	for i, row := range outRows {
		for _, field := range requiredFields {
			if _, ok := row[field]; !ok {
				t.Errorf("row %d missing field %q", i, field)
			}
		}
	}
}

func TestMaterializationPreservesSourceFields(t *testing.T) {
	dir := t.TempDir()
	rows := []connectors.Record{
		{"_polymetrics_raw_id": "r1", "email": "a@b.com", "company": "Acme"},
	}
	writeInTableNDJSON(t, dir, "contacts", rows)

	a := &DeterministicAnalyzer{}
	_, err := a.Run(context.Background(), RunRequest{
		Spec:         makeSimpleSpec(),
		InTable:      "contacts",
		OutTable:     "lead_scores",
		WarehouseDir: dir,
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}

	outPath := filepath.Join(dir, "lead_scores.ndjson")
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read OutTable: %v", err)
	}
	var row map[string]any
	if err := json.Unmarshal(data, &row); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := row["email"]; !ok {
		t.Error("OutTable row missing 'email' source field")
	}
}

func TestMaterializationAtomic(t *testing.T) {
	dir := t.TempDir()
	rows := []connectors.Record{
		{"_polymetrics_raw_id": "r1", "email": "a@b.com"},
	}
	writeInTableNDJSON(t, dir, "contacts", rows)

	a := &DeterministicAnalyzer{}
	_, err := a.Run(context.Background(), RunRequest{
		Spec:         makeSimpleSpec(),
		InTable:      "contacts",
		OutTable:     "lead_scores",
		WarehouseDir: dir,
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}

	// No temp files should remain after a successful run
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if e.Name() != "contacts.ndjson" && e.Name() != "lead_scores.ndjson" {
			t.Errorf("unexpected temp file left behind: %s", e.Name())
		}
	}
}

func TestDryRunDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	rows := []connectors.Record{
		{"_polymetrics_raw_id": "r1", "email": "a@b.com"},
	}
	writeInTableNDJSON(t, dir, "contacts", rows)

	a := &DeterministicAnalyzer{}
	_, err := a.Run(context.Background(), RunRequest{
		Spec:         makeSimpleSpec(),
		InTable:      "contacts",
		OutTable:     "lead_scores",
		WarehouseDir: dir,
		DryRun:       true,
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}

	outPath := filepath.Join(dir, "lead_scores.ndjson")
	if _, err := os.Stat(outPath); err == nil {
		t.Error("OutTable should not be written when DryRun=true")
	}
}
