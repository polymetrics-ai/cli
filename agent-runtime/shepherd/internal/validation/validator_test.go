package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGSDValidatorRejectsInvalidProductionEvidence(t *testing.T) {
	if os.Getenv("GO_WANT_VALIDATOR_HELPER") != "" {
		return
	}
	for _, test := range []struct {
		name    string
		mode    string
		stale   bool
		wantErr string
	}{
		{name: "no validation-result producer exists", mode: "no-result", wantErr: "read independent validation proof"},
		{name: "stale pre-existing result", mode: "success", stale: true, wantErr: "stale validation result exists"},
		{name: "no new validator session", mode: "no-session", wantErr: "no session"},
		{name: "validator session model is GPT-5.5", mode: "gpt55", wantErr: "validator model"},
		{name: "thinking is not high", mode: "low-thinking", wantErr: "validator thinking"},
		{name: "result head mismatch", mode: "head-mismatch", wantErr: "head or base"},
		{name: "evidence mismatch", mode: "evidence-mismatch", wantErr: "governance or evidence"},
		{name: "request nonce mismatch", mode: "nonce-mismatch", wantErr: "request nonce"},
		{name: "candidate moves during validation", mode: "move-candidate", wantErr: "candidate head moved during validation"},
		{name: "base branch stale", mode: "base-mismatch", wantErr: "head or base"},
		{name: "governance state version stale", mode: "state-version-mismatch", wantErr: "governance or evidence"},
		{name: "retry verdict", mode: "retry", wantErr: "PROCEED"},
		{name: "halt verdict", mode: "halt", wantErr: "PROCEED"},
		{name: "missing required gates", mode: "missing-gates", wantErr: "required local gates"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := initializedValidationRepo(t)
			request := validationRequestForRepo(t, repo)
			if test.stale {
				root := filepath.Join(request.StateDir, "validation", safePathPart(derivedRequestID(request)))
				if err := os.MkdirAll(root, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, "result.json"), []byte(`{"stale":true}`), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			beforeHead := gitForValidationTest(t, repo, "rev-parse", "HEAD")
			beforeGSD := readValidationFile(t, filepath.Join(repo, ".gsd", "STATE.md"))
			validator := helperValidator(test.mode, request.StateDir, request.GSDHome)
			_, err := validator.Validate(context.Background(), request)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("Validate() err=%v want containing %q", err, test.wantErr)
			}
			if got := gitForValidationTest(t, repo, "rev-parse", "HEAD"); got != beforeHead && test.mode != "move-candidate" {
				t.Fatalf("candidate head changed for %s: got %s want %s", test.mode, got, beforeHead)
			}
			if got := readValidationFile(t, filepath.Join(repo, ".gsd", "STATE.md")); got != beforeGSD {
				t.Fatalf("candidate GSD state changed: got %q want %q", got, beforeGSD)
			}
		})
	}
}

func TestGSDValidatorAcceptsFreshBoundResultAndSession(t *testing.T) {
	if os.Getenv("GO_WANT_VALIDATOR_HELPER") != "" {
		return
	}
	repo := initializedValidationRepo(t)
	request := validationRequestForRepo(t, repo)
	result, err := helperValidator("success", request.StateDir, request.GSDHome).Validate(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.SessionID == "" || result.ObservedModel != "openai-codex/gpt-5.6-sol" || result.Thinking != "high" || result.Verdict != "PROCEED" || result.ObservedHead != request.CandidateHead || result.EvidenceHash != request.EvidenceHash {
		t.Fatalf("result=%+v request=%+v", result, request)
	}
}

func TestGSDValidatorHelperProcess(t *testing.T) {
	mode := os.Getenv("GO_WANT_VALIDATOR_HELPER")
	if mode == "" {
		return
	}
	args := os.Args
	requestPath := flagValue(args, "--request")
	resultPath := flagValue(args, "--result")
	worktree := flagValue(args, "--worktree")
	if requestPath == "" || resultPath == "" || worktree == "" {
		fmt.Fprintln(os.Stderr, "missing validator arguments")
		os.Exit(2)
	}
	raw, err := os.ReadFile(requestPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	var request protectedRequest
	if err := json.Unmarshal(raw, &request); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if mode != "no-session" {
		model := "openai-codex/gpt-5.6-sol"
		thinking := "high"
		if mode == "gpt55" {
			model = "openai-codex/gpt-5.5"
		}
		if mode == "low-thinking" {
			thinking = "medium"
		}
		if err := writeValidationSession(os.Getenv("GSD_HOME"), worktree, model, thinking); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
	if mode == "move-candidate" {
		path := filepath.Join(worktree, "validator-moved.txt")
		if err := os.WriteFile(path, []byte("moved\n"), 0o600); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		gitHelper(worktree, "add", "validator-moved.txt")
		gitHelper(worktree, "commit", "-qm", "validator moved candidate")
	}
	if mode == "no-result" {
		return
	}
	now := time.Now().UTC()
	proof := proofFile{
		RequestID: request.RequestID, Nonce: request.Nonce, BaseBranch: request.BaseBranch, BaseHead: request.BaseHead,
		CandidateHead: request.CandidateHead, ObservedHead: request.CandidateHead, StateVersion: request.StateVersion,
		ContractHash: request.ContractHash, EvidenceHash: request.EvidenceHash, Verdict: "PROCEED",
		LocalGates: true, UAT: true, MilestoneValid: true, IssuedAt: now.Add(-time.Second), ExpiresAt: now.Add(time.Hour),
	}
	switch mode {
	case "head-mismatch":
		proof.ObservedHead = strings.Repeat("9", 40)
	case "evidence-mismatch":
		proof.EvidenceHash = "sha256:" + strings.Repeat("9", 64)
	case "nonce-mismatch":
		proof.Nonce = "wrong"
	case "base-mismatch":
		proof.BaseBranch = "stale"
	case "state-version-mismatch":
		proof.StateVersion++
	case "retry":
		proof.Verdict = "RETRY"
	case "halt":
		proof.Verdict = "HALT"
	case "missing-gates":
		proof.LocalGates = false
	}
	encoded, err := json.Marshal(proof)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	file, err := os.OpenFile(resultPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if _, err := file.Write(encoded); err != nil {
		_ = file.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := file.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func helperValidator(mode, stateDir, gsdHome string) GSDValidator {
	return GSDValidator{
		Command:     []string{os.Args[0], "-test.run=TestGSDValidatorHelperProcess", "--"},
		GSDHome:     gsdHome,
		StateDir:    stateDir,
		SessionsDir: filepath.Join(gsdHome, "agent", "sessions"),
		Timeout:     5 * time.Second,
		Now:         func() time.Time { return time.Now().UTC() },
		Environment: []string{"GO_WANT_VALIDATOR_HELPER=" + mode},
	}
}

func initializedValidationRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "test@example.invalid"}, {"config", "user.name", "Test"}} {
		gitHelper(root, args...)
	}
	if err := os.MkdirAll(filepath.Join(root, ".gsd"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".gsd", "STATE.md"), []byte("state\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "candidate.txt"), []byte("candidate\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitHelper(root, "add", "candidate.txt", ".gsd/STATE.md")
	gitHelper(root, "commit", "-qm", "candidate")
	return root
}

func validationRequestForRepo(t *testing.T, repo string) Request {
	t.Helper()
	gsdHome := filepath.Join(t.TempDir(), "home")
	stateDir := filepath.Join(t.TempDir(), "state")
	if err := os.MkdirAll(filepath.Join(gsdHome, "agent", "sessions"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatal(err)
	}
	return Request{
		Repository: "polymetrics-ai/cli", PullRequest: 391, BaseBranch: "feat/372-gsd-pi-go-shepherd",
		DeliveryID: "issue-389", Generation: 2, UnitID: "M001/S01/T01", UnitType: "execute-task", Attempt: 1,
		StateVersion: 44, WorkDir: repo, GSDHome: gsdHome, StateDir: stateDir,
		BaseHead: strings.Repeat("a", 40), CandidateHead: gitForValidationTest(t, repo, "rev-parse", "HEAD"),
		ContractHash: "sha256:" + strings.Repeat("c", 64), EvidenceHash: "sha256:" + strings.Repeat("e", 64),
		ArtifactHashes: []ArtifactHash{{Path: "candidate.txt", Hash: "sha256:" + strings.Repeat("f", 64)}},
		RequireGates:   GateRequirements{LocalGates: true, UAT: true, MilestoneValid: true},
	}
}

func writeValidationSession(gsdHome, worktree, model, thinking string) error {
	sessions := filepath.Join(gsdHome, "agent", "sessions")
	if err := os.MkdirAll(sessions, 0o700); err != nil {
		return err
	}
	id := "11111111-2222-3333-4444-555555555555"
	path := filepath.Join(sessions, id+".jsonl")
	provider, modelID, ok := strings.Cut(model, "/")
	if !ok {
		return fmt.Errorf("bad model %s", model)
	}
	content := fmt.Sprintf("{\"type\":\"session\",\"id\":%q,\"cwd\":%q}\n{\"type\":\"model_change\",\"provider\":%q,\"modelId\":%q}\n{\"type\":\"thinking_level_change\",\"thinkingLevel\":%q}\n", id, worktree, provider, modelID, thinking)
	return os.WriteFile(path, []byte(content), 0o600)
}

func flagValue(args []string, name string) string {
	for i, arg := range args {
		if arg == name && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func gitForValidationTest(t *testing.T, root string, args ...string) string {
	t.Helper()
	return strings.TrimSpace(gitHelper(root, args...))
}

func gitHelper(root string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("git %v: %v: %s", args, err, out))
	}
	return string(out)
}

func readValidationFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
