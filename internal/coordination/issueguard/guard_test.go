package issueguard

import (
	"strings"
	"testing"
)

func TestValidatePRAcceptsClosingKeyword(t *testing.T) {
	result := ValidatePR("feat(github): add cli surface metadata", "Implements the first slice.\n\nCloses #123\n")
	if !result.OK {
		t.Fatalf("ValidatePR() OK = false, violations = %v", result.Violations)
	}
	if len(result.Issues) != 1 || result.Issues[0].Number != 123 || !result.Issues[0].Closing {
		t.Fatalf("ValidatePR() issues = %#v", result.Issues)
	}
}

func TestValidatePRRejectsMissingIssue(t *testing.T) {
	result := ValidatePR("feat(github): add cli surface metadata", "No linked issue yet.")
	if result.OK {
		t.Fatal("ValidatePR() OK = true, want false")
	}
	if !containsViolation(result.Violations, "PR body must reference an issue") {
		t.Fatalf("ValidatePR() violations = %v", result.Violations)
	}
}

func TestValidatePRRejectsTitleOnlyReference(t *testing.T) {
	result := ValidatePR("feat(github): add cli surface metadata fixes #123", "Body explains work but omits linkage.")
	if result.OK {
		t.Fatal("ValidatePR() OK = true, want false")
	}
	if !containsViolation(result.Violations, "PR body must reference an issue") {
		t.Fatalf("ValidatePR() violations = %v", result.Violations)
	}
}

func TestValidatePRAllowsNonClosingReferenceForStackedIncrement(t *testing.T) {
	result := ValidatePR("test(github): add guard coverage", "Part of a stacked implementation.\n\nRefs #456\n")
	if !result.OK {
		t.Fatalf("ValidatePR() OK = false, violations = %v", result.Violations)
	}
	if len(result.Issues) != 1 || result.Issues[0].Number != 456 || result.Issues[0].Closing {
		t.Fatalf("ValidatePR() issues = %#v", result.Issues)
	}
}

func TestValidatePRAllowsNarrativeNonClosingReferences(t *testing.T) {
	body := "Implement the promoted provider-inert Windows signing foundation for PM issues #554 and #555; reference #550/#554/#555 without closing them; never merge."
	result := ValidatePR("feat(packaging): add provider-inert Windows signing foundation", body)
	if !result.OK {
		t.Fatalf("ValidatePR() OK = false, violations = %v", result.Violations)
	}
	if len(result.Issues) != 3 {
		t.Fatalf("ValidatePR() issues = %#v, want 3 issues", result.Issues)
	}
	for i, want := range []int{550, 554, 555} {
		if result.Issues[i].Number != want || result.Issues[i].Closing {
			t.Fatalf("ValidatePR() issues = %#v, want non-closing issue %d at index %d", result.Issues, want, i)
		}
	}
}

func TestValidatePRAllowsDeliveryIssueNumberWithoutHash(t *testing.T) {
	body := "Ship the focused PM v0.1.1 Gong calls list correction for issue 596: add bounded --from/--to filters."
	result := ValidatePR("fix(connectors): add Gong calls list date filters", body)
	if !result.OK {
		t.Fatalf("ValidatePR() OK = false, violations = %v", result.Violations)
	}
	if len(result.Issues) != 1 || result.Issues[0].Number != 596 || result.Issues[0].Closing {
		t.Fatalf("ValidatePR() issues = %#v, want non-closing issue 596", result.Issues)
	}
}

func TestValidatePRAllowsNoMistakesDeliveryRecord(t *testing.T) {
	body := noMistakesDeliveryBody()
	result := ValidatePR("ci: add dry-run Homebrew tap notification", body)
	if !result.OK {
		t.Fatalf("ValidatePR() OK = false, violations = %v", result.Violations)
	}
	if len(result.Issues) != 0 {
		t.Fatalf("ValidatePR() issues = %#v, want none", result.Issues)
	}
	if !result.DeliveryRecord {
		t.Fatal("ValidatePR() DeliveryRecord = false, want true")
	}
}

func TestValidatePRRejectsIncompleteNoMistakesDeliveryRecord(t *testing.T) {
	body := "## Intent\n\nImplement a CI update.\n\n## Pipeline\n\nUpdates from [git push no-mistakes](https://github.com/kunchenguid/no-mistakes)\n"
	result := ValidatePR("ci: add dry-run Homebrew tap notification", body)
	if result.OK {
		t.Fatal("ValidatePR() OK = true, want false")
	}
	if !containsViolation(result.Violations, "PR body must reference an issue") {
		t.Fatalf("ValidatePR() violations = %v", result.Violations)
	}
}

func TestValidatePRAcceptsUnvalidatedCheckpointCanonicalIssueLinks(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "LF", body: unvalidatedCheckpointBody()},
		{name: "CRLF", body: strings.ReplaceAll(unvalidatedCheckpointBody(), "\n", "\r\n")},
		{name: "GitHub-indented body", body: unvalidatedCheckpointBodyWithGitHubIndentation()},
		{name: "GitHub-indented CRLF body", body: strings.ReplaceAll(unvalidatedCheckpointBodyWithGitHubIndentation(), "\n", "\r\n")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePR("chore(airtable): stage unvalidated parity checkpoint", tt.body)
			if !result.OK {
				t.Fatalf("ValidatePR() OK = false, violations = %v", result.Violations)
			}

			want := []int{3070, 3071, 3072, 3073, 3074, 3075, 3076, 3077}
			if len(result.Issues) != len(want) {
				t.Fatalf("ValidatePR() issues = %#v, want %v", result.Issues, want)
			}
			for i, number := range want {
				if result.Issues[i].Number != number || result.Issues[i].Closing {
					t.Fatalf("ValidatePR() issues = %#v, want non-closing issue %d at index %d", result.Issues, number, i)
				}
			}
		})
	}
}

func TestValidatePRRejectsCanonicalIssueSectionWithoutCompletedTaskWording(t *testing.T) {
	body := strings.Join([]string{
		"## Unvalidated cloud checkpoint — do not merge yet",
		"",
		"This draft pull request records a task that is still in progress.",
		"",
		"## Canonical issue links preserved from the task record",
		"",
		"- https://github.com/polymetrics-ai/cli/issues/3070",
		"",
	}, "\n")
	result := ValidatePR("chore(airtable): stage unvalidated parity checkpoint", body)
	if result.OK {
		t.Fatal("ValidatePR() OK = true, want false")
	}
	if !containsViolation(result.Violations, "PR body must reference an issue") {
		t.Fatalf("ValidatePR() violations = %v", result.Violations)
	}
}

func TestValidatePRRejectsCanonicalIssueSectionWithoutCheckpointHeading(t *testing.T) {
	body := strings.Join([]string{
		"## Staged parity branch record",
		"",
		"This draft pull request is an unvalidated cloud checkpoint for the completed `cli-airtable-parity-wave03-r1` connector parity task.",
		"",
		"## Canonical issue links preserved from the task record",
		"",
		"- https://github.com/polymetrics-ai/cli/issues/3070",
		"",
	}, "\n")
	result := ValidatePR("chore(airtable): stage unvalidated parity checkpoint", body)
	if result.OK {
		t.Fatal("ValidatePR() OK = true, want false")
	}
	if !containsViolation(result.Violations, "PR body must reference an issue") {
		t.Fatalf("ValidatePR() violations = %v", result.Violations)
	}
}

func TestValidatePRIgnoresIssueLinksAfterCanonicalCheckpointSection(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "LF", body: unvalidatedCheckpointBodyWithTrailingSection()},
		{name: "CRLF", body: strings.ReplaceAll(unvalidatedCheckpointBodyWithTrailingSection(), "\n", "\r\n")},
		{name: "GitHub-indented trailing heading", body: strings.Replace(unvalidatedCheckpointBodyWithTrailingSection(), "\n## Later notes", "\n  ## Later notes", 1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePR("chore(airtable): stage unvalidated parity checkpoint", tt.body)
			if !result.OK {
				t.Fatalf("ValidatePR() OK = false, violations = %v", result.Violations)
			}

			want := []int{3070, 3071, 3072, 3073, 3074, 3075, 3076, 3077}
			if len(result.Issues) != len(want) {
				t.Fatalf("ValidatePR() issues = %#v, want %v", result.Issues, want)
			}
			for i, number := range want {
				if result.Issues[i].Number != number {
					t.Fatalf("ValidatePR() issues = %#v, want issue %d at index %d", result.Issues, number, i)
				}
			}
			for _, ref := range result.Issues {
				if ref.Number == 9001 {
					t.Fatalf("ValidatePR() harvested issue 9001 from the section after the canonical issue links, issues = %#v", result.Issues)
				}
			}
		})
	}
}

func TestValidatePRAllowsLetteredDeliveryIssueMigrationIntent(t *testing.T) {
	body := "Implement the focused connector-boundary Issue B migration on branch refactor/connector-engine-policy-migration: remove GitHub-specific shared runtime policy names."
	result := ValidatePR("feat(connectors): genericize repository read policies", body)
	if !result.OK {
		t.Fatalf("ValidatePR() OK = false, violations = %v", result.Violations)
	}
	if !result.ExplicitIssueWording {
		t.Fatal("ValidatePR() ExplicitIssueWording = false, want true")
	}
	if len(result.Issues) != 0 {
		t.Fatalf("ValidatePR() issues = %#v, want no numeric issue refs", result.Issues)
	}
}

func TestValidatePRAcceptsCrossRepositoryClosingReference(t *testing.T) {
	result := ValidatePR("feat: add release provenance and linux packages", "Closes polymetrics-ai/cli#551\n")
	if !result.OK {
		t.Fatalf("ValidatePR() OK = false, violations = %v", result.Violations)
	}
	if len(result.Issues) != 1 || result.Issues[0].Number != 551 || !result.Issues[0].Closing {
		t.Fatalf("ValidatePR() issues = %#v", result.Issues)
	}
}

func TestValidatePRAcceptsIssueFirstDeliveryIntent(t *testing.T) {
	body := "Implement the first shippable cross-platform PM release-trust slice under parent polymetrics-ai/cli#550 by fully delivering issues #551 and #552 only."
	result := ValidatePR("feat: add release provenance and linux packages", body)
	if !result.OK {
		t.Fatalf("ValidatePR() OK = false, violations = %v", result.Violations)
	}

	want := []int{550, 551, 552}
	if len(result.Issues) != len(want) {
		t.Fatalf("ValidatePR() issues = %#v", result.Issues)
	}
	for i, number := range want {
		if result.Issues[i].Number != number || result.Issues[i].Closing {
			t.Fatalf("ValidatePR() issues = %#v", result.Issues)
		}
	}
}

func TestValidatePRRejectsAmbiguousIssueRelationship(t *testing.T) {
	tests := []string{
		"Related to #123\n",
		"Mentions #123\n",
		"Issue #123\n",
		"Issue 123\n",
		"Issue B\n",
		"Implement Issue B\n",
		"Implement issue a migration\n",
		"References #123\n",
		"Ship this. Issue 123 is unrelated.\n",
		"Do not implement issue #123\n",
		"Do not ship issue 123\n",
		"Do not implement Issue B migration\n",
	}
	for _, body := range tests {
		result := ValidatePR("feat(github): add cli surface metadata", body)
		if result.OK {
			t.Fatalf("ValidatePR(%q) OK = true, want false", body)
		}
		if !containsViolation(result.Violations, "PR body must reference an issue") {
			t.Fatalf("ValidatePR(%q) violations = %v", body, result.Violations)
		}
	}
}

func TestValidatePRRejectsNonConventionalTitle(t *testing.T) {
	result := ValidatePR("add cli surface metadata", "Closes #123\n")
	if result.OK {
		t.Fatal("ValidatePR() OK = true, want false")
	}
	if !containsViolation(result.Violations, "PR title must use Conventional Commits") {
		t.Fatalf("ValidatePR() violations = %v", result.Violations)
	}
}

func TestValidatePRRejectsTitleAcceptedByOldLooseScopePattern(t *testing.T) {
	result := ValidatePR("feat(/github): add cli surface metadata", "Closes #123\n")
	if result.OK {
		t.Fatal("ValidatePR() OK = true, want false")
	}
	if !containsViolation(result.Violations, "PR title must use Conventional Commits") {
		t.Fatalf("ValidatePR() violations = %v", result.Violations)
	}
}

func containsViolation(violations []string, want string) bool {
	for _, violation := range violations {
		if strings.Contains(violation, want) {
			return true
		}
	}
	return false
}

func noMistakesDeliveryBody() string {
	return strings.Join([]string{
		"## Intent",
		"",
		"Implement the CLI-side least-privilege Homebrew tap notification.",
		"",
		"## What Changed",
		"",
		"- Added a least-privilege dry-run notification.",
		"",
		"## Testing",
		"",
		"Targeted validation passed.",
		"",
		"## Pipeline",
		"",
		"Updates from [git push no-mistakes](https://github.com/kunchenguid/no-mistakes)",
		"",
	}, "\n")
}

func unvalidatedCheckpointBody() string {
	return strings.Join([]string{
		"## Unvalidated cloud checkpoint — do not merge yet",
		"",
		"This draft pull request is an unvalidated cloud checkpoint for the completed `cli-airtable-parity-wave03-r1` connector parity task. It exists only as durable cloud proof of the current local task branch state.",
		"",
		"Review and validation are pending and will run later in a batch of five connectors. This pull request must not be merged yet.",
		"",
		"This staging pass did not run review, tests, lint, documentation generation, validation pipelines, rebase, or merge work. It only publishes the existing clean local task branch HEAD.",
		"",
		"## Recorded task branch",
		"",
		"- Connector/task: `airtable`",
		"- Branch: `fm/cli-airtable-parity-wave03-r1`",
		"- HEAD: `69f87c4976a0dbd225eb74d9768f4e74294e120e`",
		"- Base target: `main`",
		"",
		"## Canonical issue links preserved from the task record",
		"",
		"- https://github.com/polymetrics-ai/cli/issues/3070",
		"- https://github.com/polymetrics-ai/cli/issues/3071",
		"- https://github.com/polymetrics-ai/cli/issues/3077",
		"- https://github.com/polymetrics-ai/cli/issues/3072",
		"- https://github.com/polymetrics-ai/cli/issues/3073",
		"- https://github.com/polymetrics-ai/cli/issues/3074",
		"- https://github.com/polymetrics-ai/cli/issues/3075",
		"- https://github.com/polymetrics-ai/cli/issues/3076",
		"",
	}, "\n")
}

func unvalidatedCheckpointBodyWithTrailingSection() string {
	return unvalidatedCheckpointBody() + strings.Join([]string{
		"## Later notes outside the task record",
		"",
		"- https://github.com/polymetrics-ai/cli/issues/9001",
		"",
	}, "\n")
}

func unvalidatedCheckpointBodyWithGitHubIndentation() string {
	lines := strings.Split(unvalidatedCheckpointBody(), "\n")
	indent := false
	for i, line := range lines {
		if strings.HasPrefix(line, "This draft pull request") {
			indent = true
		}
		if indent && line != "" {
			lines[i] = "    " + line
		}
		if line == "- https://github.com/polymetrics-ai/cli/issues/3070" {
			indent = false
		}
	}
	return strings.Join(lines, "\n")
}
