package contract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeliverySchemasAreValidJSONAndDoNotGrantMerge(t *testing.T) {
	t.Parallel()

	directory := filepath.Join("..", "..", "..", "..", ".agents", "agentic-delivery", "schemas")
	names := []string{"issue-milestone-context", "gsd-unit-dispatch", "gsd-worker-handoff", "external-effect-intent", "local-review-verdict", "human-decision", "canary-record"}
	paths := make([]string, 0, len(names))
	for _, name := range names {
		paths = append(paths, filepath.Join(directory, name+".schema.json"))
	}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var schema map[string]any
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if strings.Contains(strings.ToLower(string(raw)), `"merge"`) {
			t.Fatalf("%s exposes merge as a schema value", path)
		}
	}
}
