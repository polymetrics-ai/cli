package gsd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/contract"
)

func TestBootstrapIssueProjectIsAtomicAndIssueScoped(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	context := testIssueContext(389)
	identity := IssueProjectIdentity{
		DeliveryID: "issue-389", Issue: 389, ParentIssue: 372,
		Branch: "feat/389-autonomous-shepherd", BaseBranch: "feat/372-gsd-pi-go-shepherd",
		ProjectRoot: root, InitialHead: strings.Repeat("a", 40), ContextHash: "sha256:" + strings.Repeat("b", 64),
		GSDVersion: "1.11.0",
	}
	if err := BootstrapIssueProject(root, identity, context); err != nil {
		t.Fatal(err)
	}
	if err := BootstrapIssueProject(root, identity, context); err != nil {
		t.Fatalf("exact restart failed: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(root, ".gsd", "ISSUE.json"))
	if err != nil {
		t.Fatal(err)
	}
	var observed IssueProjectIdentity
	if err := json.Unmarshal(raw, &observed); err != nil {
		t.Fatal(err)
	}
	if observed != identity {
		t.Fatalf("identity=%+v want %+v", observed, identity)
	}
	preferences, err := os.ReadFile(filepath.Join(root, ".gsd", "PREFERENCES.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"planning:", "gpt-5.6-sol", "execution:", "gpt-5.5", "thinking: high"} {
		if !strings.Contains(string(preferences), required) {
			t.Fatalf("preferences missing %q: %s", required, preferences)
		}
	}

	drifted := identity
	drifted.Issue = 390
	if err := BootstrapIssueProject(root, drifted, testIssueContext(390)); err == nil {
		t.Fatal("project was rebound to a different issue")
	}
}

func testIssueContext(issue int) contract.IssueContext {
	return contract.IssueContext{
		Issue: issue, ParentIssue: 372, Objective: "Build an autonomous Shepherd",
		Scope: []string{"agent-runtime/shepherd"}, NonGoals: []string{"merge main"},
		AcceptanceCriteria: []string{"one command reaches human gate"},
		WriteScope:         []string{"agent-runtime/shepherd/**"}, RequiredReading: []string{"AGENTS.md"},
		RequiredSkills: []string{"golang-how-to"}, Verification: []string{"go test ./..."},
		Safety: []string{"no secrets"}, HumanGates: []string{"main merge"},
		Branch: "feat/389-autonomous-shepherd", PRBase: "feat/372-gsd-pi-go-shepherd",
		ReviewRoute: "claude", Sources: []string{"https://github.com/polymetrics-ai/cli/issues/389"},
	}
}
