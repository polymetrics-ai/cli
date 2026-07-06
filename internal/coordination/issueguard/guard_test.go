package issueguard

import (
	"strings"
	"testing"
)

func TestValidatePRBodyAcceptsClosingKeyword(t *testing.T) {
	result := ValidatePRBody("feat(github): add cli surface metadata", "Implements the first slice.\n\nCloses #123\n")
	if !result.OK {
		t.Fatalf("ValidatePRBody() OK = false, violations = %v", result.Violations)
	}
	if len(result.Issues) != 1 || result.Issues[0].Number != 123 || !result.Issues[0].Closing {
		t.Fatalf("ValidatePRBody() issues = %#v", result.Issues)
	}
}

func TestValidatePRBodyRejectsMissingIssue(t *testing.T) {
	result := ValidatePRBody("feat(github): add cli surface metadata", "No linked issue yet.")
	if result.OK {
		t.Fatal("ValidatePRBody() OK = true, want false")
	}
	if !containsViolation(result.Violations, "PR body must reference an issue") {
		t.Fatalf("ValidatePRBody() violations = %v", result.Violations)
	}
}

func TestValidatePRBodyRejectsTitleOnlyReference(t *testing.T) {
	result := ValidatePRBody("feat(github): add cli surface metadata fixes #123", "Body explains work but omits linkage.")
	if result.OK {
		t.Fatal("ValidatePRBody() OK = true, want false")
	}
	if !containsViolation(result.Violations, "PR body must reference an issue") {
		t.Fatalf("ValidatePRBody() violations = %v", result.Violations)
	}
}

func TestValidatePRBodyAllowsNonClosingReferenceForStackedIncrement(t *testing.T) {
	result := ValidatePRBody("test(github): add guard coverage", "Part of a stacked implementation.\n\nRefs #456\n")
	if !result.OK {
		t.Fatalf("ValidatePRBody() OK = false, violations = %v", result.Violations)
	}
	if len(result.Issues) != 1 || result.Issues[0].Number != 456 || result.Issues[0].Closing {
		t.Fatalf("ValidatePRBody() issues = %#v", result.Issues)
	}
}

func TestValidatePRBodyRejectsNonConventionalTitle(t *testing.T) {
	result := ValidatePRBody("add cli surface metadata", "Closes #123\n")
	if result.OK {
		t.Fatal("ValidatePRBody() OK = true, want false")
	}
	if !containsViolation(result.Violations, "PR title must use Conventional Commits") {
		t.Fatalf("ValidatePRBody() violations = %v", result.Violations)
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
