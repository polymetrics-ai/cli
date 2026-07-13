package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/gsd"
)

func TestExplicitDepthApprovalIsNarrowlyScoped(t *testing.T) {
	t.Parallel()

	question := gsd.Question{ID: "depth_verification_M001-n6ms9v_confirm", Method: "select", Title: "Confirm the printed depth verification is sufficient", Options: []string{"Confirm (Recommended)", "Reject depth"}}
	response, approved := approveDepthQuestion(question)
	if !approved || response.Value != question.Options[0] || response.Cancelled {
		t.Fatalf("response=%+v approved=%t", response, approved)
	}
	variant := question
	variant.Options = []string{"Depth verified (Recommended)", "Needs revision", "None of the above"}
	response, approved = approveDepthQuestion(variant)
	if !approved || response.Value != variant.Options[0] || response.Cancelled {
		t.Fatalf("variant response=%+v approved=%t", response, approved)
	}
	for _, other := range []gsd.Question{
		{Method: "confirm", Title: question.Title, Options: []string{"Yes", "No"}},
		{ID: "approve_dependency", Method: "select", Title: "Approve dependency addition", Options: question.Options},
		{ID: "depth_verification_M001-n6ms9v_review", Method: "select", Title: question.Title, Options: question.Options},
		{ID: "depth_verification_M001-n6ms9v_confirm", Method: "select", Title: question.Title, Options: []string{"Proceed", "Stop"}},
	} {
		if response, approved := approveDepthQuestion(other); approved || !response.Cancelled {
			t.Fatalf("unrelated gate was approved: question=%+v response=%+v", other, response)
		}
	}
}

func TestGovernedPathRejectsEscape(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	inside := filepath.Join(root, "context.md")
	if err := os.WriteFile(inside, []byte("context"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := governedPath(root, inside); err != nil {
		t.Fatalf("inside path rejected: %v", err)
	}
	outside := filepath.Join(t.TempDir(), "context.md")
	if err := os.WriteFile(outside, []byte("context"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := governedPath(root, outside); err == nil {
		t.Fatal("expected path escape to fail")
	}
	symlink := filepath.Join(root, "link.json")
	if err := os.Symlink(outside, symlink); err != nil {
		t.Fatal(err)
	}
	if _, err := governedPath(root, symlink); err == nil {
		t.Fatal("expected symlink escape to fail")
	}
}

func TestDeliveryIDIsStablePerIssue(t *testing.T) {
	t.Parallel()
	if got := deliveryID(372); got != "issue-372" {
		t.Fatalf("delivery id=%q", got)
	}
}

func TestTerminalDiagnosticIsSingleLineAndBounded(t *testing.T) {
	t.Parallel()
	got := terminalDiagnostic("first\n[headless] Error: session switch cancelled\x00\n")
	if got != "[headless] Error: session switch cancelled" {
		t.Fatalf("diagnostic=%q", got)
	}
	if len(terminalDiagnostic(string(make([]byte, 2000)))) > 512 {
		t.Fatal("diagnostic was not bounded")
	}
}

func TestLoadConfigRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.json")
	raw := `{"gsd_command":["gsd"],"work_dir":"/tmp/work","gsd_home":"/tmp/home","state_dir":"/tmp/state","unexpected":true}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(path); err == nil {
		t.Fatal("expected unknown config field to fail")
	}
}

func TestMaterializeContainerContextCopiesPlanningFileIntoProtectedOverlay(t *testing.T) {
	t.Parallel()
	workDir := t.TempDir()
	contextPath := filepath.Join(workDir, ".planning", "phases", "issue-380", "ISSUE-CONTEXT.json")
	stateDir := t.TempDir()
	raw := []byte(`{"issue":380}`)
	config := fileConfig{Runtime: "podman", WorkDir: workDir, StateDir: stateDir}
	if err := materializeContainerContext(config, contextPath, raw); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(stateDir, "runtime", "planning", "phases", "issue-380", "ISSUE-CONTEXT.json")
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(raw) {
		t.Fatalf("protected context=%q, want %q", got, raw)
	}
}
