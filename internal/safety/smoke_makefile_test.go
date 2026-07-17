package safety_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type smokeReverseStatement struct {
	name      string
	statement string
}

var smokeReverseExpectedStatements = []smokeReverseStatement{
	{name: "reverse plan", statement: `PLAN_OUTPUT=$$(./pm reverse plan customers_to_outbox --source-table sample_customers --destination outbox:outbox-local --map id:external_id --map name:full_name --map email:email --root "$$SMOKE_DIR")`},
	{name: "plan id extraction", statement: `PLAN_ID=$$(printf '%s\n' "$$PLAN_OUTPUT" | awk '/Created reverse plan/ {print $$4}')`},
	{name: "reverse preview", statement: `./pm reverse preview "$$PLAN_ID" --root "$$SMOKE_DIR" --json >/dev/null`},
	{name: "approval extraction", statement: `APPROVAL=$$(printf '%s\n' "$$PLAN_OUTPUT" | awk '/Approval token:/ {print $$3}')`},
	{name: "reverse run", statement: `./pm reverse run "$$PLAN_ID" --approve "$$APPROVAL" --root "$$SMOKE_DIR" --json >/dev/null`},
}

func TestSmokeNoBuildReversePlanPreviewRunOrdering(t *testing.T) {
	makefile := readRepoMakefile(t)
	recipe := extractMakeTargetRecipe(t, makefile, "smoke-no-build")

	if ok, reason := smokeNoBuildReversePlanPreviewRunOrdering(recipe); !ok {
		t.Fatal(reason)
	}
}

func TestSmokeNoBuildReversePlanPreviewRunOrderingRejectsSyntheticFalsePositives(t *testing.T) {
	tests := []struct {
		name     string
		makefile string
	}{
		{
			name:     "commented markers do not count",
			makefile: makefileWithSmokeRecipe(prefixedSmokeReverseStatements("# ")...),
		},
		{
			name:     "echo text does not count",
			makefile: makefileWithSmokeRecipe(quotedSmokeReverseStatements("echo ")...),
		},
		{
			name:     "printf help text does not count",
			makefile: makefileWithSmokeRecipe(quotedSmokeReverseStatements("printf '%s\\n' usage:")...),
		},
		{
			name:     "false-prefixed commands do not count",
			makefile: makefileWithSmokeRecipe(prefixedSmokeReverseStatements("false && ")...),
		},
		{
			name:     "unrelated target does not count",
			makefile: makefileWithSmokeRecipe("true") + makeTargetRecipe("other-target", smokeReverseStatementLines()...),
		},
		{
			name:     "wrong-order markers do not count",
			makefile: makefileWithSmokeRecipe(reverseSmokeReverseStatementLines()...),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe := extractMakeTargetRecipe(t, tt.makefile, "smoke-no-build")
			if ok, _ := smokeNoBuildReversePlanPreviewRunOrdering(recipe); ok {
				t.Fatalf("accepted invalid smoke-no-build reverse recipe:\n%s", recipe)
			}
		})
	}
}

func smokeNoBuildReversePlanPreviewRunOrdering(recipe string) (bool, string) {
	statements := executableMakeRecipeStatements(recipe)
	next := 0
	for _, marker := range smokeReverseExpectedStatements {
		found := false
		for next < len(statements) {
			if statements[next] == marker.statement {
				found = true
				next++
				break
			}
			next++
		}
		if !found {
			return false, fmt.Sprintf("smoke-no-build recipe missing executable %s statement %q in order; executable statements: %q", marker.name, marker.statement, statements)
		}
	}
	return true, ""
}

func executableMakeRecipeStatements(recipe string) []string {
	var logicalLines []string
	var current strings.Builder
	for _, rawLine := range strings.Split(recipe, "\n") {
		line := strings.TrimSpace(strings.TrimPrefix(rawLine, "\t"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "@"))
		line, continued := trimMakeContinuation(line)
		if line != "" {
			if current.Len() > 0 {
				current.WriteByte(' ')
			}
			current.WriteString(line)
		}
		if continued {
			continue
		}
		if current.Len() > 0 {
			logicalLines = append(logicalLines, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		logicalLines = append(logicalLines, current.String())
	}

	var statements []string
	for _, line := range logicalLines {
		for _, statement := range splitShellStatements(line) {
			statement = strings.TrimSpace(statement)
			if statement == "" || strings.HasPrefix(statement, "#") {
				continue
			}
			statements = append(statements, statement)
		}
	}
	return statements
}

func splitShellStatements(line string) []string {
	var statements []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false
	for _, char := range line {
		switch {
		case escaped:
			current.WriteRune(char)
			escaped = false
		case char == '\\' && !inSingleQuote:
			current.WriteRune(char)
			escaped = true
		case char == '\'' && !inDoubleQuote:
			current.WriteRune(char)
			inSingleQuote = !inSingleQuote
		case char == '"' && !inSingleQuote:
			current.WriteRune(char)
			inDoubleQuote = !inDoubleQuote
		case char == ';' && !inSingleQuote && !inDoubleQuote:
			statements = append(statements, current.String())
			current.Reset()
		default:
			current.WriteRune(char)
		}
	}
	statements = append(statements, current.String())
	return statements
}

func trimMakeContinuation(line string) (string, bool) {
	line = strings.TrimRight(line, " \t")
	if !strings.HasSuffix(line, `\`) {
		return strings.TrimSpace(line), false
	}
	return strings.TrimSpace(strings.TrimSuffix(line, `\`)), true
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

func makefileWithSmokeRecipe(lines ...string) string {
	return makeTargetRecipe("smoke-no-build", lines...)
}

func makeTargetRecipe(target string, lines ...string) string {
	var makefile strings.Builder
	makefile.WriteString(".PHONY: ")
	makefile.WriteString(target)
	makefile.WriteString("\n")
	makefile.WriteString(target)
	makefile.WriteString(":\n")
	for _, line := range lines {
		makefile.WriteByte('\t')
		makefile.WriteString(line)
		makefile.WriteByte('\n')
	}
	return makefile.String()
}

func prefixedSmokeReverseStatements(prefix string) []string {
	lines := make([]string, 0, len(smokeReverseExpectedStatements))
	for _, statement := range smokeReverseExpectedStatements {
		lines = append(lines, prefix+statement.statement)
	}
	return lines
}

func quotedSmokeReverseStatements(prefix string) []string {
	lines := make([]string, 0, len(smokeReverseExpectedStatements))
	for _, statement := range smokeReverseExpectedStatements {
		lines = append(lines, prefix+"'"+statement.statement+"'")
	}
	return lines
}

func smokeReverseStatementLines() []string {
	lines := make([]string, 0, len(smokeReverseExpectedStatements))
	for _, statement := range smokeReverseExpectedStatements {
		lines = append(lines, statement.statement)
	}
	return lines
}

func reverseSmokeReverseStatementLines() []string {
	lines := smokeReverseStatementLines()
	for left, right := 0, len(lines)-1; left < right; left, right = left+1, right-1 {
		lines[left], lines[right] = lines[right], lines[left]
	}
	return lines
}
