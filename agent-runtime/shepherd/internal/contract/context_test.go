package contract

import (
	"strings"
	"testing"
)

func TestDecodeIssueContextBindsIssueAndRejectsUnsafeScope(t *testing.T) {
	t.Parallel()

	valid := `{"issue":372,"parent_issue":372,"objective":"Build governed Shepherd","scope":["runtime"],"non_goals":[],"acceptance_criteria":["green"],"dependencies":[],"write_scope":["agent-runtime/shepherd/**"],"required_reading":["AGENTS.md"],"required_skills":["golang-how-to"],"tdd":{"red":"tests fail","green":"tests pass","refactor":"simplify"},"verification":["go test ./..."],"safety":["no secrets"],"human_gates":["main merge"],"branch":"feat/372-gsd-pi-go-shepherd","pr_base":"main","review_route":"local","sources":["https://github.com/polymetrics-ai/cli/issues/372"]}`
	context, raw, err := DecodeIssueContext(strings.NewReader(valid), 372)
	if err != nil {
		t.Fatalf("valid context: %v", err)
	}
	if context.Issue != 372 || len(raw) == 0 {
		t.Fatalf("context=%+v raw=%d", context, len(raw))
	}
	if _, _, err := DecodeIssueContext(strings.NewReader(strings.Replace(valid, `"issue":372`, `"issue":373`, 1)), 372); err == nil {
		t.Fatal("expected issue mismatch")
	}
	if _, _, err := DecodeIssueContext(strings.NewReader(strings.Replace(valid, `agent-runtime/shepherd/**`, `../outside`, 1)), 372); err == nil {
		t.Fatal("expected unsafe scope")
	}
}
