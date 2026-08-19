package certify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestDirectReadCheckpointResumesOnlyMatchingCandidates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress", "github-direct-read.json")
	candidates := []directReadCandidate{{
		StageName: "direct_read_sweep_readme",
		Command:   "readme view",
		Args:      []string{"github", "readme", "view", "--json"},
		OutputAssertions: []engine.CertificationOutputAssertion{{
			JSONPointer: "/response/name",
			Equals:      "README.md",
		}},
	}}
	config := map[string]string{"owner": "Polymetrics-Cert", "repo": "fixture"}

	checkpoint, err := newDirectReadCheckpoint("github", candidates, config)
	if err != nil {
		t.Fatalf("newDirectReadCheckpoint() error = %v", err)
	}
	checkpoint.Completed[candidates[0].StageName] = true
	if err := saveDirectReadCheckpoint(path, checkpoint); err != nil {
		t.Fatalf("saveDirectReadCheckpoint() error = %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(checkpoint) error = %v", err)
	}
	for _, forbidden := range []string{"Polymetrics-Cert", "fixture", "readme view"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("checkpoint persisted raw candidate/configuration text %q", forbidden)
		}
	}

	loaded, err := loadDirectReadCheckpoint(path, "github", candidates, config)
	if err != nil {
		t.Fatalf("loadDirectReadCheckpoint() error = %v", err)
	}
	if !loaded.Completed[candidates[0].StageName] {
		t.Fatal("loaded checkpoint did not retain the completed direct-read stage")
	}

	changed := append([]directReadCandidate(nil), candidates...)
	changed[0].OutputAssertions[0].Equals = "changed-after-schema-compile"
	if _, err := loadDirectReadCheckpoint(path, "github", changed, config); err == nil {
		t.Fatal("loadDirectReadCheckpoint() accepted a checkpoint for changed assertions")
	}
}
