package safety_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSmokeNoBuildReversePlanPreviewRunOrdering(t *testing.T) {
	makefile := readRepoMakefile(t)
	recipe := extractMakeTargetRecipe(t, makefile, "smoke-no-build")

	markers := []struct {
		name string
		text string
	}{
		{name: "reverse plan", text: `PLAN_OUTPUT=$$(./pm reverse plan customers_to_outbox`},
		{name: "plan id extraction", text: `PLAN_ID=$$(printf '%s\n' "$$PLAN_OUTPUT" | awk '/Created reverse plan/ {print $$4}');`},
		{name: "reverse preview", text: `./pm reverse preview "$$PLAN_ID" --root "$$SMOKE_DIR" --json >/dev/null;`},
		{name: "approval extraction", text: `APPROVAL=$$(printf '%s\n' "$$PLAN_OUTPUT" | awk '/Approval token:/ {print $$3}');`},
		{name: "reverse run", text: `./pm reverse run "$$PLAN_ID" --approve "$$APPROVAL" --root "$$SMOKE_DIR" --json >/dev/null;`},
	}

	indexes := make([]int, len(markers))
	for i, marker := range markers {
		indexes[i] = strings.Index(recipe, marker.text)
		if indexes[i] < 0 {
			t.Fatalf("smoke-no-build recipe missing %s step %q", marker.name, marker.text)
		}
	}
	for i := 1; i < len(indexes); i++ {
		if indexes[i-1] >= indexes[i] {
			t.Fatalf("smoke-no-build reverse ordering invalid: %s must come before %s", markers[i-1].name, markers[i].name)
		}
	}
}

func readRepoMakefile(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test file")
	}
	path := filepath.Join(filepath.Dir(file), "..", "..", "Makefile")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}
	return string(data)
}

func extractMakeTargetRecipe(t *testing.T, makefile, target string) string {
	t.Helper()
	lines := strings.Split(makefile, "\n")
	prefix := target + ":"
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			var recipe strings.Builder
			for _, recipeLine := range lines[i+1:] {
				if strings.TrimSpace(recipeLine) == "" {
					recipe.WriteByte('\n')
					continue
				}
				if !strings.HasPrefix(recipeLine, "\t") {
					break
				}
				recipe.WriteString(recipeLine)
				recipe.WriteByte('\n')
			}
			return recipe.String()
		}
	}
	t.Fatalf("Makefile target %q not found", target)
	return ""
}
