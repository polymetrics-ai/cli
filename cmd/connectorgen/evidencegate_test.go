package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFoundationEvidence_AllPassedClaimsHaveMatchingGateAndSHA(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	manifest := readEvidenceFixture(t, filepath.Join(root, "data", "cli-current-foundations-main-integration-r1", "evidence-manifest.json"))
	tdd := readEvidenceFixture(t, filepath.Join(root, ".planning", "phases", "cli-current-foundations-main-integration-r1", "TDD-LEDGER.md"))
	review := readEvidenceFixture(t, filepath.Join(root, ".planning", "phases", "cli-current-foundations-main-integration-r1", "REVIEW.md"))
	if err := validateFoundationEvidence(manifest, tdd, review); err != nil {
		t.Fatalf("checked-in evidence: %v", err)
	}

	var document map[string]any
	if err := json.Unmarshal(manifest, &document); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		edit func(map[string]any)
		want string
	}{
		{
			name: "reviewed SHA mismatch",
			edit: func(value map[string]any) { value["reviewed_sha"] = strings.Repeat("0", 40) },
			want: "does not match review source_sha",
		},
		{
			name: "passed claim paired with pending gate",
			edit: func(value map[string]any) { evidenceGateAt(value, 2)["status"] = "passed" },
			want: "disagrees with TDD ledger",
		},
		{
			name: "cached result promoted to passed gate",
			edit: func(value map[string]any) { evidenceGateAt(value, 0)["command_indexes"] = []any{float64(2)} },
			want: "relies on cached command evidence",
		},
		{
			name: "mode-limited evidence loses scope",
			edit: func(value map[string]any) { evidenceGateAt(value, 0)["modes"] = []any{} },
			want: "no named command, test, or mode evidence",
		},
		{
			name: "provisional evidence loses reason",
			edit: func(value map[string]any) { evidenceGateAt(value, 2)["reason"] = "" },
			want: "provisional gate INT-03 has no reason",
		},
		{
			name: "deferred evidence loses reason",
			edit: func(value map[string]any) { evidenceGateAt(value, 5)["reason"] = "" },
			want: "deferred gate INT-06 has no reason",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			copyRaw, err := json.Marshal(document)
			if err != nil {
				t.Fatal(err)
			}
			var mutated map[string]any
			if err := json.Unmarshal(copyRaw, &mutated); err != nil {
				t.Fatal(err)
			}
			tc.edit(mutated)
			raw, err := json.Marshal(mutated)
			if err != nil {
				t.Fatal(err)
			}
			err = validateFoundationEvidence(raw, tdd, review)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func evidenceGateAt(document map[string]any, index int) map[string]any {
	return document["gate_ledger"].([]any)[index].(map[string]any)
}

func readEvidenceFixture(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
