package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	decisionlog "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/decision"
	shepherdgithub "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/github"
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
	exact := question
	exact.ID = "depth_verification_M001_confirm"
	response, approved = approveDepthQuestion(exact)
	if !approved || response.Value != exact.Options[0] || response.Cancelled {
		t.Fatalf("exact response=%+v approved=%t", response, approved)
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

func TestJoinTerminalFailurePreservesPrimaryCause(t *testing.T) {
	t.Parallel()
	primary := errors.New("canonical unit did not advance")
	secondary := errors.New("worktree must be clean")
	joined := joinTerminalFailure(primary, secondary)
	if !errors.Is(joined, primary) || !errors.Is(joined, secondary) {
		t.Fatalf("joined error=%v", joined)
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

func TestLoadConfigRequiresCompleteDecisionPRBinding(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.json")
	raw := `{"gsd_command":["gsd"],"work_dir":"/tmp/work","gsd_home":"/tmp/home","state_dir":"/tmp/state","repository":"polymetrics-ai/cli"}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(path); err == nil {
		t.Fatal("partial PR binding accepted")
	}
}

func TestLoadConfigDefaultsToBoundedNestedAgentEnvelopeSize(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.json")
	raw := `{"gsd_command":["gsd"],"work_dir":"/tmp/work","gsd_home":"/tmp/home","state_dir":"/tmp/state","repository":"polymetrics-ai/cli","pull_request":388}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if config.MaxEventBytes < 4_000_000 || config.MaxEventBytes > 8*1024*1024 {
		t.Fatalf("max event bytes=%d does not admit measured nested-agent envelopes within 8 MiB", config.MaxEventBytes)
	}
}

type recordingDecisionPublisher struct {
	summaries []string
	err       error
}

func (p *recordingDecisionPublisher) SyncDecisionComment(_ context.Context, _ shepherdgithub.Target, summary string) error {
	p.summaries = append(p.summaries, summary)
	return p.err
}

func TestAppendAndPublishDecisionRetainsDurableRecordWhenPublicationFails(t *testing.T) {
	t.Parallel()

	store, err := decisionlog.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	publisher := &recordingDecisionPublisher{err: errors.New("offline")}
	question := gsd.Question{ID: "scope", Title: "What should ship?"}
	response := gsd.UIResponse{Value: "Full safe parity"}
	err = appendAndPublishDecision(context.Background(), store, publisher,
		shepherdgithub.Target{Repository: "polymetrics-ai/cli", PullRequest: 388, DeliveryID: "issue-380"},
		"issue-380", "execution-1", "discuss-milestone/M001", question, response, "shepherd", "approved issue context")
	if err == nil {
		t.Fatal("publication failure did not fail the answered gate")
	}
	records, readErr := store.Records()
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(records) != 1 || records[0].Actor != decisionlog.ActorShepherd || records[0].At.Before(time.Now().Add(-time.Minute)) {
		t.Fatalf("durable records=%+v", records)
	}
	if len(publisher.summaries) != 1 {
		t.Fatalf("publication attempts=%d", len(publisher.summaries))
	}
}

func TestRecordOperationalDecisionPublishesAttributedLedgerEntry(t *testing.T) {
	t.Parallel()

	store, err := decisionlog.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	publisher := &recordingDecisionPublisher{}
	err = recordOperationalDecision(context.Background(), store, publisher,
		shepherdgithub.Target{Repository: "polymetrics-ai/cli", PullRequest: 388, DeliveryID: "issue-380"},
		"issue-380", "reconcile/M001/S01/T04", "Amend the issue write scope?", "Add only internal/connectors/defs/defs_test.go",
		"shepherd", "approved task plan requires the production embed assertion")
	if err != nil {
		t.Fatal(err)
	}
	records, err := store.Records()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].Actor != decisionlog.ActorShepherd || records[0].Question != "Amend the issue write scope?" {
		t.Fatalf("records=%+v", records)
	}
	if len(publisher.summaries) != 1 || !strings.Contains(publisher.summaries[0], "Add only internal/connectors/defs/defs_test.go") {
		t.Fatalf("published summaries=%v", publisher.summaries)
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

func TestProtectedIssueContextIsImmutableAndHashBound(t *testing.T) {
	t.Parallel()
	stateDir := t.TempDir()
	raw := []byte(`{"issue":380,"parent_issue":380,"objective":"Asana parity","scope":["connector"],"non_goals":[],"acceptance_criteria":["green"],"dependencies":[],"write_scope":["internal/connectors/engine/**"],"required_reading":["AGENTS.md"],"required_skills":["golang-how-to"],"tdd":{"red":"fail","green":"pass","refactor":"clean"},"verification":["go test ./..."],"safety":["no secrets"],"human_gates":["main merge"],"branch":"feat/380-asana-cli-parity","pr_base":"main","review_route":"local","sources":["https://github.com/polymetrics-ai/cli/issues/380"]}`)
	if err := materializeProtectedIssueContext(stateDir, 380, raw); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(raw)
	context, err := loadProtectedIssueContext(stateDir, 380, "sha256:"+hex.EncodeToString(hash[:]))
	if err != nil {
		t.Fatal(err)
	}
	if len(context.WriteScope) != 1 || context.WriteScope[0] != "internal/connectors/engine/**" {
		t.Fatalf("context=%+v", context)
	}
	if err := materializeProtectedIssueContext(stateDir, 380, append(raw, ' ')); err == nil {
		t.Fatal("protected context overwrite accepted")
	}
	if _, err := loadProtectedIssueContext(stateDir, 380, "sha256:"+strings.Repeat("0", 64)); err == nil {
		t.Fatal("context hash mismatch accepted")
	}
}
