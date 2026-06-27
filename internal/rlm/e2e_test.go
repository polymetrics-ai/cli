package rlm

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLikelyCustomersFlowOffline(t *testing.T) {
	specData, err := os.ReadFile("testdata/likely_customers.json")
	if err != nil {
		t.Fatalf("read testdata/likely_customers.json: %v", err)
	}
	spec, err := ParseSpec(specData)
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}

	dir := t.TempDir()

	f := &FixtureAnalyzer{}
	result, err := f.Run(context.Background(), RunRequest{
		Spec:         spec,
		InTable:      "contacts",
		OutTable:     "lead_scores",
		WarehouseDir: dir,
	})
	if err != nil {
		t.Fatalf("FixtureAnalyzer.Run: %v", err)
	}
	if result.RecordsScored < 5 {
		t.Errorf("RecordsScored = %d, want >= 5", result.RecordsScored)
	}

	// Read OutTable and check top score >= 0.5
	outPath := filepath.Join(dir, "lead_scores.ndjson")
	f2, err := os.Open(outPath)
	if err != nil {
		t.Fatalf("open OutTable: %v", err)
	}
	defer f2.Close()

	sc := bufio.NewScanner(f2)
	sc.Scan()
	var firstRow map[string]any
	if err := json.Unmarshal(sc.Bytes(), &firstRow); err != nil {
		t.Fatalf("unmarshal first row: %v", err)
	}
	topScore, ok := firstRow["_rlm_score"].(float64)
	if !ok {
		t.Fatalf("_rlm_score missing or not float64 in first row: %v", firstRow["_rlm_score"])
	}
	if topScore < 0.5 {
		t.Errorf("top _rlm_score = %f, want >= 0.5", topScore)
	}

	// Re-run: determinism check
	dir2 := t.TempDir()
	f3 := &FixtureAnalyzer{}
	result2, err := f3.Run(context.Background(), RunRequest{
		Spec:         spec,
		InTable:      "contacts",
		OutTable:     "lead_scores",
		WarehouseDir: dir2,
	})
	if err != nil {
		t.Fatalf("second FixtureAnalyzer.Run: %v", err)
	}
	if result2.RecordsScored != result.RecordsScored {
		t.Errorf("second run RecordsScored = %d, want %d", result2.RecordsScored, result.RecordsScored)
	}
}
