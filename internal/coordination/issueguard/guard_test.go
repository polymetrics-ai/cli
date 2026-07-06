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

func TestValidatePRRejectsAmbiguousIssueRelationship(t *testing.T) {
	tests := []string{
		"Related to #123\n",
		"Issue #123\n",
		"References #123\n",
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
