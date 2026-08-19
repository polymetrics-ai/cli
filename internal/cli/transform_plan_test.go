package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/connectors/database"
)

func TestReadTransformPlanFileReturnsNormalizedClosedPlan(t *testing.T) {
	path := filepath.Join(t.TempDir(), "transform.json")
	if err := os.WriteFile(path, []byte(`{"select":[{"source":"id","target":"event_id","type":"int64"}],"version":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	plan, err := readTransformPlanFile(path)
	if err != nil {
		t.Fatalf("readTransformPlanFile() error = %v", err)
	}
	if got, want := string(plan.NormalizedJSON()), `{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"}]}`; got != want {
		t.Fatalf("normalized plan = %s, want %s", got, want)
	}
}

func TestReadTransformPlanFileRefusesNonRegularPathBeforeRead(t *testing.T) {
	_, err := readTransformPlanFile(t.TempDir())
	if !errors.Is(err, database.ErrTransformPlanInvalid) {
		t.Fatalf("readTransformPlanFile(directory) error = %T %v, want ErrTransformPlanInvalid", err, err)
	}
}

func TestReadTransformPlanFileRefusesEmptyAndOversizedInput(t *testing.T) {
	for name, raw := range map[string][]byte{
		"empty":     nil,
		"oversized": make([]byte, maxTransformPlanFileBytes+1),
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "transform.json")
			if err := os.WriteFile(path, raw, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := readTransformPlanFile(path); !errors.Is(err, database.ErrTransformPlanInvalid) {
				t.Fatalf("readTransformPlanFile(%s) error = %T %v, want ErrTransformPlanInvalid", name, err, err)
			}
		})
	}
}
