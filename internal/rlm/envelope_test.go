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

// readPlainNDJSON reads a non-enveloped OutTable NDJSON file (one JSON object per line).
func readPlainNDJSON(t *testing.T, path string) []connectors.Record {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %q: %v", path, err)
	}
	defer f.Close()
	var out []connectors.Record
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var rec connectors.Record
		if err := json.Unmarshal(line, &rec); err != nil {
			t.Fatalf("parse out line: %v", err)
		}
		out = append(out, rec)
	}
	return out
}

// TestDeterministicRun_FilePath_CarriesRawID is the missing characterization test
// (Slice 0a). It exercises the real FILE path (not in-memory scoreRecords) and
// asserts the warehouse envelope's _polymetrics_raw_id survives into the OutTable.
// Before the readEnvelopedRecords fix this fails: readNDJSON drops the id.
func TestDeterministicRun_FilePath_CarriesRawID(t *testing.T) {
	whDir := t.TempDir()
	inPath := filepath.Join(whDir, "contacts.ndjson")
	rows := `{"_polymetrics_raw_id":"r1","record":{"email":"a@b.com","company":"Acme"}}
{"_polymetrics_raw_id":"r2","record":{"email":"c@d.com"}}
`
	if err := os.WriteFile(inPath, []byte(rows), 0o644); err != nil {
		t.Fatalf("write InTable: %v", err)
	}

	spec := &Spec{Name: "s", Features: []Feature{{Name: "email", Weight: 1, ScoreIfSet: 1}}}
	a := &DeterministicAnalyzer{}
	if _, err := a.Run(context.Background(), RunRequest{
		Spec: spec, InTable: "contacts", OutTable: "scored", WarehouseDir: whDir,
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	out := readPlainNDJSON(t, filepath.Join(whDir, "scored.ndjson"))
	if len(out) != 2 {
		t.Fatalf("want 2 output rows, got %d", len(out))
	}
	seen := map[string]bool{}
	for i, row := range out {
		id, ok := row["_polymetrics_raw_id"].(string)
		if !ok || id == "" {
			t.Fatalf("output row %d missing _polymetrics_raw_id: %v", i, row)
		}
		seen[id] = true
	}
	if !seen["r1"] || !seen["r2"] {
		t.Fatalf("output did not carry both raw ids: %v", seen)
	}
}

// TestValidateOutTable rejects path-escaping table names (Slice 0a security helper).
func TestValidateOutTable(t *testing.T) {
	bad := []string{"../etc", "a/b", "..", "/abs", "x/../y"}
	for _, name := range bad {
		if err := validateOutTable(name); err == nil {
			t.Errorf("validateOutTable(%q) = nil, want error", name)
		}
	}
	good := []string{"lead_scores", "contacts_scored", "t1"}
	for _, name := range good {
		if err := validateOutTable(name); err != nil {
			t.Errorf("validateOutTable(%q) = %v, want nil", name, err)
		}
	}
}
