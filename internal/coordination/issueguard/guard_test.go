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

func TestValidatePRAllowsFixDeliveryIssueNumberWithoutHash(t *testing.T) {
	body := "After the investigation, they authorized a focused v0.1.1 fix on issue 596/PR 597 to make Gong calls list date-range and limit behavior correct."
	result := ValidatePR("feat(pmbroker): add broker v1 HTTP client contract", body)
	if !result.OK {
		t.Fatalf("ValidatePR() OK = false, violations = %v", result.Violations)
	}
	if len(result.Issues) != 1 || result.Issues[0].Number != 596 || result.Issues[0].Closing {
		t.Fatalf("ValidatePR() issues = %#v, want non-closing issue 596", result.Issues)
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
		"References #123\n",
		"Ship this. Issue 123 is unrelated.\n",
		"This is a fix. Issue 123 is unrelated.\n",
		"Do not implement issue #123\n",
		"Do not fix issue #123\n",
		"Do not ship issue 123\n",
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
