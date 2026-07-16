package validation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testNonce = "11111111111111111111111111111111"

func TestProductionValidatorRejectsUnsupportedInventedGSDInterface(t *testing.T) {
	t.Parallel()
	gsdPath, err := exec.LookPath("gsd")
	if err != nil {
		t.Skip("pinned GSD executable is not installed")
	}
	repo := initializedValidationRepo(t)
	request := validationRequestForRepo(t, repo)
	validator := GSDValidator{Command: []string{gsdPath}, GSDHome: request.GSDHome, StateDir: request.StateDir, Timeout: 5 * time.Second}
	if _, err := validator.Validate(context.Background(), request); err == nil || !strings.Contains(err.Error(), "does not advertise") {
		t.Fatalf("former gsd headless shepherd-validate interface was not rejected quickly: %v", err)
	}
}

func TestGSDValidatorRejectsInvalidProductionPiEvidence(t *testing.T) {
	if os.Getenv("GO_WANT_VALIDATOR_HELPER") != "" {
		return
	}
	for _, test := range []struct {
		name    string
		mode    string
		stale   bool
		wantErr string
	}{
		{name: "no validation-result producer exists", mode: "missing-result", wantErr: "agent_end"},
		{name: "stale pre-existing result", mode: "success", stale: true, wantErr: "stale validation result exists"},
		{name: "no new validator session", mode: "no-session", wantErr: "did not create or update"},
		{name: "validator session model is GPT-5.5", mode: "gpt55", wantErr: "unexpected current-session model"},
		{name: "thinking is not high", mode: "low-thinking", wantErr: "unexpected current-session thinking"},
		{name: "result head mismatch", mode: "head-mismatch", wantErr: "head or base"},
		{name: "evidence mismatch", mode: "evidence-mismatch", wantErr: "governance or evidence"},
		{name: "request nonce mismatch", mode: "nonce-mismatch", wantErr: "request nonce"},
		{name: "repository mismatch", mode: "repository-mismatch", wantErr: "repository or PR"},
		{name: "pull request mismatch", mode: "pr-mismatch", wantErr: "repository or PR"},
		{name: "candidate moves during validation", mode: "move-candidate", wantErr: "candidate head moved during validation"},
		{name: "artifact changes during validation", mode: "change-artifact", wantErr: "revalidate candidate artifacts"},
		{name: "base branch stale", mode: "base-mismatch", wantErr: "head or base"},
		{name: "governance state version stale", mode: "state-version-mismatch", wantErr: "governance or evidence"},
		{name: "missing required gates", mode: "missing-gates", wantErr: "required local gates"},
		{name: "malformed result", mode: "malformed", wantErr: "structured evidence missing"},
		{name: "duplicate proof field", mode: "duplicate-proof-field", wantErr: "structured evidence missing"},
		{name: "unknown proof field", mode: "unknown-proof-field", wantErr: "structured evidence missing"},
		{name: "ambiguous proof objects", mode: "ambiguous-proof", wantErr: "structured evidence missing"},
		{name: "missing turn lifecycle", mode: "missing-turn", wantErr: "message_start is out of order"},
		{name: "duplicate turn lifecycle", mode: "duplicate-turn", wantErr: "turn_start is out of order"},
		{name: "out of order turn lifecycle", mode: "out-of-order-turn", wantErr: "turn_end is out of order"},
		{name: "stream session mismatch", mode: "stream-session-mismatch", wantErr: "session does not match"},
		{name: "durable proof mismatch", mode: "durable-proof-mismatch", wantErr: "does not match"},
		{name: "errored proof message", mode: "stream-error-proof", wantErr: "not a successful terminal"},
		{name: "length stopped proof message", mode: "stream-length-proof", wantErr: "not a successful terminal"},
		{name: "retrying agent end", mode: "stream-agent-retry", wantErr: "agent_end"},
		{name: "errored agent end", mode: "stream-agent-error", wantErr: "agent_end"},
		{name: "empty turn", mode: "stream-empty-turn", wantErr: "turn_end is out of order"},
		{name: "tool before message", mode: "stream-tool-before-message", wantErr: "forbidden tool"},
		{name: "duplicate lifecycle type", mode: "stream-duplicate-type", wantErr: "duplicate JSON fields"},
		{name: "duplicate agent retry", mode: "stream-duplicate-agent-retry", wantErr: "duplicate JSON fields"},
		{name: "duplicate stop reason", mode: "stream-duplicate-stop", wantErr: "duplicate JSON fields"},
		{name: "case alias lifecycle type", mode: "stream-case-type", wantErr: "duplicate JSON fields"},
		{name: "case alias agent retry", mode: "stream-case-agent-retry", wantErr: "duplicate JSON fields"},
		{name: "case alias stop reason", mode: "stream-case-stop", wantErr: "duplicate JSON fields"},
		{name: "zero tool proceed", mode: "stream-zero-tool", wantErr: "lacks successful evidence"},
		{name: "missing tool outcome", mode: "stream-missing-tool-outcome", wantErr: "explicit success"},
		{name: "failed tool outcome", mode: "stream-failed-tool", wantErr: "explicit success"},
		{name: "nonzero exit", mode: "nonzero", wantErr: "dedicated validator failed"},
		{name: "timeout", mode: "timeout", wantErr: "context deadline exceeded"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := initializedValidationRepo(t)
			request := validationRequestForRepo(t, repo)
			validator := helperValidator(test.mode, request.StateDir, request.GSDHome)
			if test.mode == "timeout" {
				validator.Timeout = 50 * time.Millisecond
			}
			if test.stale {
				root := filepath.Join(request.StateDir, "validation", safePathPart(derivedRequestID(request)), testNonce)
				if err := os.MkdirAll(root, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, "result.json"), []byte(`{"stale":true}`), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			beforeHead := gitForValidationTest(t, repo, "rev-parse", "HEAD")
			beforeGSD := readValidationFile(t, filepath.Join(repo, ".gsd", "STATE.md"))
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

func TestGSDValidatorReturnsBoundRetryAndHaltForRatification(t *testing.T) {
	if os.Getenv("GO_WANT_VALIDATOR_HELPER") != "" {
		return
	}
	for _, verdict := range []string{"retry", "halt"} {
		t.Run(verdict, func(t *testing.T) {
			repo := initializedValidationRepo(t)
			request := validationRequestForRepo(t, repo)
			result, err := helperValidator(verdict, request.StateDir, request.GSDHome).Validate(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if result.Verdict != strings.ToUpper(verdict) || result.ObservedHead != request.CandidateHead || result.EvidenceHash != request.EvidenceHash {
				t.Fatalf("result=%+v", result)
			}
		})
	}
}

func TestValidatorEventStreamRejectsUnboundToolUpdates(t *testing.T) {
	for _, event := range []string{
		`{"type":"tool_execution_update"}`,
		`{"type":"tool_execution_end"}`,
	} {
		stream := strings.Join([]string{`{"type":"session","id":"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"}`,
			`{"type":"agent_start"}`, `{"type":"turn_start"}`, event,
			`{"type":"turn_end"}`, `{"type":"agent_end"}`, `{"type":"agent_settled"}`}, "\n")
		if _, err := proofFromPiJSON([]byte(stream)); err == nil || !strings.Contains(err.Error(), "unbound") {
			t.Fatalf("event=%s err=%v", event, err)
		}
	}
}

func TestGSDValidatorUsesExactReadOnlyPiProcessBoundary(t *testing.T) {
	if os.Getenv("GO_WANT_VALIDATOR_HELPER") != "" {
		return
	}
	repo := initializedValidationRepo(t)
	request := validationRequestForRepo(t, repo)
	validator := helperValidator("success", request.StateDir, request.GSDHome)
	result, err := validator.Validate(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.SessionID == "" || result.ObservedModel != "openai-codex/gpt-5.6-sol" || result.Thinking != "high" || result.Verdict != "PROCEED" || result.ObservedHead != request.CandidateHead || result.EvidenceHash != request.EvidenceHash {
		t.Fatalf("result=%+v request=%+v", result, request)
	}
}

func TestValidatorRejectsPreexistingSessionDirectory(t *testing.T) {
	if os.Getenv("GO_WANT_VALIDATOR_HELPER") != "" {
		return
	}
	repo := initializedValidationRepo(t)
	request := validationRequestForRepo(t, repo)
	preexisting := filepath.Join(t.TempDir(), "reused-session")
	if err := os.MkdirAll(preexisting, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(preexisting, "stale.jsonl"), []byte("stale\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	validator := helperValidator("success", request.StateDir, request.GSDHome)
	validator.SessionsDir = preexisting
	if _, err := validator.Validate(context.Background(), request); err == nil || !strings.Contains(err.Error(), "must be fresh") {
		t.Fatalf("preexisting validator session directory error=%v", err)
	}
}

func TestValidatorCapabilityProbeCannotMutateCandidate(t *testing.T) {
	if os.Getenv("GO_WANT_VALIDATOR_HELPER") != "" {
		return
	}
	repo := initializedValidationRepo(t)
	request := validationRequestForRepo(t, repo)
	validator := helperValidator("probe-side-effect", request.StateDir, request.GSDHome)
	if _, err := validator.Validate(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(repo, "probe-side-effect.txt")); !os.IsNotExist(err) {
		t.Fatalf("probe side effect reached candidate: %v", err)
	}
}

func TestGSDValidatorRetryUsesFreshNonceDirectory(t *testing.T) {
	if os.Getenv("GO_WANT_VALIDATOR_HELPER") != "" {
		return
	}
	repo := initializedValidationRepo(t)
	request := validationRequestForRepo(t, repo)
	nonces := []string{"11111111111111111111111111111111", "22222222222222222222222222222222"}
	validator := helperValidator("success", request.StateDir, request.GSDHome)
	validator.Nonce = func() (string, error) {
		nonce := nonces[0]
		nonces = nonces[1:]
		return nonce, nil
	}
	if _, err := validator.Validate(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if _, err := validator.Validate(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(request.StateDir, "validation", safePathPart(derivedRequestID(request)))
	for _, nonce := range []string{"11111111111111111111111111111111", "22222222222222222222222222222222"} {
		if _, err := os.Stat(filepath.Join(root, nonce, "result.json")); err != nil {
			t.Fatalf("fresh retry result %s missing: %v", nonce, err)
		}
	}
}

func TestLivePiValidatorSmoke(t *testing.T) {
	if os.Getenv("POLYMETRICS_SHEPHERD_LIVE_VALIDATOR") != "1" {
		t.Skip("set POLYMETRICS_SHEPHERD_LIVE_VALIDATOR=1 for the opt-in live Pi smoke")
	}
	piPath, err := exec.LookPath("pi")
	if err != nil {
		t.Fatal(err)
	}
	repo := initializedValidationRepo(t)
	request := validationRequestForRepo(t, repo)
	if configured := os.Getenv("PI_CODING_AGENT_DIR"); configured != "" {
		request.GSDHome = filepath.Dir(configured)
	} else if home, homeErr := os.UserHomeDir(); homeErr == nil {
		request.GSDHome = filepath.Join(home, ".pi")
	}
	request.RequireGates = GateRequirements{}
	validator := GSDValidator{Command: []string{piPath}, GSDHome: request.GSDHome, StateDir: request.StateDir, Timeout: 2 * time.Minute}
	result, err := validator.Validate(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.ObservedModel != "openai-codex/gpt-5.6-sol" || result.Thinking != "high" || result.SessionID == "" || result.ObservedHead != request.CandidateHead || result.EvidenceHash != request.EvidenceHash {
		t.Fatalf("live validator evidence=%+v", result)
	}
	t.Logf("live validator model=%s thinking=%s session=%s verdict=%s head=%s evidence=%s", result.ObservedModel, result.Thinking, result.SessionID, result.Verdict, result.ObservedHead, result.EvidenceHash)
}

func TestGSDValidatorHelperProcess(t *testing.T) {
	mode := os.Getenv("GO_WANT_VALIDATOR_HELPER")
	if mode == "" {
		return
	}
	defer func() { os.Exit(0) }()
	args := os.Args
	if containsArg(args, "--help") {
		if mode == "probe-side-effect" {
			if err := os.WriteFile("probe-side-effect.txt", []byte("probe\n"), 0o600); err != nil {
				os.Exit(2)
			}
		}
		for _, required := range []string{"--no-tools", "--no-extensions", "--no-skills",
			"--no-prompt-templates", "--no-themes", "--no-context-files", "--no-approve"} {
			if !containsArg(args, required) {
				fmt.Fprintf(os.Stderr, "probe missing %s\n", required)
				os.Exit(2)
			}
		}
		if os.Getenv("HOME") == "" || os.Getenv("GSD_PROJECT_ROOT") == "" ||
			os.Getenv("GH_TOKEN") != "" || os.Getenv("SSH_AUTH_SOCK") != "" {
			fmt.Fprintln(os.Stderr, "probe environment is not isolated")
			os.Exit(2)
		}
		fmt.Println("--mode --print --session-dir --tools --model --thinking")
		return
	}
	assertValidatorPiArguments(args)
	if mode == "nonzero" {
		os.Exit(7)
	}
	if mode == "timeout" {
		time.Sleep(time.Second)
		return
	}
	prompt := args[len(args)-1]
	requestJSON := strings.TrimPrefix(prompt[strings.LastIndex(prompt, "Request JSON: "):], "Request JSON: ")
	var request protectedRequest
	if err := json.Unmarshal([]byte(requestJSON), &request); err != nil {
		fmt.Fprintln(os.Stderr, "invalid bounded validator request")
		os.Exit(2)
	}
	worktree, err := os.Getwd()
	if err != nil || worktree != request.WorkDir || os.Getenv("GSD_PROJECT_ROOT") != request.WorkDir {
		fmt.Fprintln(os.Stderr, "candidate worktree binding mismatch")
		os.Exit(2)
	}
	sessionDir := flagValue(args, "--session-dir")
	if sessionDir == "" || !strings.HasPrefix(sessionDir, requestPathStateRoot(request)+string(os.PathSeparator)) {
		fmt.Fprintln(os.Stderr, "dedicated session directory is not protected")
		os.Exit(2)
	}
	model := "openai-codex/gpt-5.6-sol"
	thinking := "high"
	if mode != "no-session" {
		if mode == "gpt55" {
			model = "openai-codex/gpt-5.5"
		}
		if mode == "low-thinking" {
			thinking = "medium"
		}
		if err := writeValidationSession(sessionDir, worktree, model, thinking, request.Nonce); err != nil {
			fmt.Fprintln(os.Stderr, "cannot create validator session")
			os.Exit(2)
		}
	}
	if mode == "move-candidate" {
		path := filepath.Join(worktree, "validator-moved.txt")
		if err := os.WriteFile(path, []byte("moved\n"), 0o600); err != nil {
			os.Exit(2)
		}
		gitHelper(worktree, "add", "validator-moved.txt")
		gitHelper(worktree, "commit", "-qm", "validator moved candidate")
	}
	if mode == "change-artifact" {
		if err := os.Remove(filepath.Join(worktree, "candidate.txt")); err != nil {
			os.Exit(2)
		}
	}
	if mode == "missing-result" {
		fmt.Println(`{"type":"agent_end","messages":[]}`)
		return
	}
	streamSessionID := request.Nonce[:8] + "-2222-3333-4444-555555555555"
	if mode == "malformed" {
		emitPiTextEvent("{not-json", streamSessionID, mode)
		return
	}
	proof := proofFile{
		RequestID: request.RequestID, Nonce: request.Nonce, Repository: request.Repository, PullRequest: request.PullRequest,
		BaseBranch: request.BaseBranch, BaseHead: request.BaseHead, CandidateHead: request.CandidateHead,
		ObservedHead: request.CandidateHead, StateVersion: request.StateVersion,
		ContractHash: request.ContractHash, EvidenceHash: request.EvidenceHash, Verdict: "PROCEED",
		LocalGates: true, UAT: true, MilestoneValid: true, IssuedAt: request.IssuedAt, ExpiresAt: request.ExpiresAt,
	}
	switch mode {
	case "head-mismatch":
		proof.ObservedHead = strings.Repeat("9", 40)
	case "evidence-mismatch":
		proof.EvidenceHash = "sha256:" + strings.Repeat("9", 64)
	case "nonce-mismatch":
		proof.Nonce = "wrong"
	case "repository-mismatch":
		proof.Repository = "other/repo"
	case "pr-mismatch":
		proof.PullRequest++
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
		os.Exit(2)
	}
	switch mode {
	case "duplicate-proof-field":
		encoded = []byte(strings.Replace(string(encoded), `"verdict":"PROCEED"`,
			`"verdict":"PROCEED","verdict":"HALT"`, 1))
	case "unknown-proof-field":
		encoded = []byte(strings.TrimSuffix(string(encoded), "}") + `,"unexpected":true}`)
	case "ambiguous-proof":
		encoded = append(append([]byte(nil), encoded...), encoded...)
	case "missing-turn", "duplicate-turn", "out-of-order-turn":
		emitInvalidPiLifecycle(mode, string(encoded), streamSessionID)
		return
	case "stream-session-mismatch":
		streamSessionID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	}
	if mode != "no-session" {
		durableSessionID := request.Nonce[:8] + "-2222-3333-4444-555555555555"
		durableProof := string(encoded)
		if mode == "durable-proof-mismatch" {
			durableProof = `{"verdict":"HALT"}`
		}
		if err := appendValidationSessionProof(flagValue(args, "--session-dir"), durableSessionID,
			model, durableProof); err != nil {
			os.Exit(2)
		}
	}
	emitPiTextEvent(string(encoded), streamSessionID, mode)
}

func assertValidatorPiArguments(args []string) {
	expected := map[string]string{
		"--mode": "json", "--model": "openai-codex/gpt-5.6-sol", "--thinking": "high",
		"--tools": "read,grep,find,ls",
	}
	for flag, want := range expected {
		if got := flagValue(args, flag); got != want {
			fmt.Fprintf(os.Stderr, "%s=%s want %s\n", flag, got, want)
			os.Exit(2)
		}
	}
	for _, required := range []string{"--print", "--session-dir", "--system-prompt", "--no-extensions", "--no-skills", "--no-prompt-templates", "--no-themes", "--no-context-files", "--no-approve"} {
		if !containsArg(args, required) {
			fmt.Fprintf(os.Stderr, "missing %s\n", required)
			os.Exit(2)
		}
	}
	if containsArg(args, "headless") || containsArg(args, "shepherd-validate") || containsArg(args, "edit") || containsArg(args, "write") || containsArg(args, "subagent") {
		fmt.Fprintln(os.Stderr, "unsafe or invented validator argument")
		os.Exit(2)
	}
}

func TestSanitizedValidatorEnvironmentDropsCredentialsAndHome(t *testing.T) {
	t.Parallel()
	environment := sanitizedValidatorEnvironment([]string{
		"PATH=/usr/bin", "HOME=/secret-home", "GH_TOKEN=secret", "GITHUB_TOKEN=secret",
		"SSH_AUTH_SOCK=/tmp/agent", "GIT_CONFIG_GLOBAL=/secret", "OPENAI_API_KEY=secret",
		"GO_WANT_VALIDATOR_HELPER=valid",
	})
	joined := strings.Join(environment, "\n")
	for _, forbidden := range []string{"HOME=", "GH_TOKEN=", "GITHUB_TOKEN=", "SSH_AUTH_SOCK=",
		"GIT_CONFIG_GLOBAL=", "OPENAI_API_KEY="} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("validator environment retained %s: %q", forbidden, joined)
		}
	}
	for _, required := range []string{"PATH=/usr/bin", "GO_WANT_VALIDATOR_HELPER=valid",
		"GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS="} {
		if !strings.Contains(joined, required) {
			t.Fatalf("validator environment omitted %s: %q", required, joined)
		}
	}
}

func emitInvalidPiLifecycle(mode, text, sessionID string) {
	fmt.Printf("{\"type\":\"session\",\"id\":%q}\n", sessionID)
	fmt.Println(`{"type":"agent_start"}`)
	switch mode {
	case "missing-turn":
		fmt.Println(`{"type":"message_start","message":{"role":"assistant","content":[]}}`)
	case "duplicate-turn":
		fmt.Println(`{"type":"turn_start"}`)
		fmt.Println(`{"type":"turn_start"}`)
	case "out-of-order-turn":
		fmt.Println(`{"type":"turn_end"}`)
	}
	_ = text
	fmt.Println(`{"type":"agent_end"}`)
	fmt.Println(`{"type":"agent_settled"}`)
}

func emitPiTextEvent(text, sessionID, mode string) {
	if mode == "stream-duplicate-type" {
		fmt.Printf("{\"type\":\"session\",\"type\":\"agent_start\",\"id\":%q}\n", sessionID)
		return
	}
	if mode == "stream-case-type" {
		fmt.Printf("{\"type\":\"session\",\"Type\":\"agent_start\",\"id\":%q}\n", sessionID)
		return
	}
	fmt.Printf("{\"type\":\"session\",\"id\":%q}\n", sessionID)
	fmt.Println(`{"type":"agent_start"}`)
	if mode == "stream-empty-turn" {
		fmt.Println(`{"type":"turn_start"}`)
		fmt.Println(`{"type":"turn_end"}`)
	}
	if mode == "stream-tool-before-message" {
		fmt.Println(`{"type":"turn_start"}`)
		fmt.Println(`{"type":"tool_execution_start","toolName":"read","toolCallId":"early"}`)
	} else if mode != "stream-zero-tool" {
		fmt.Println(`{"type":"turn_start"}`)
		fmt.Println(`{"type":"message_start","message":{"role":"assistant","content":[]}}`)
		fmt.Println(`{"type":"message_end","message":{"role":"assistant","stopReason":"toolUse","content":[]}}`)
		fmt.Println(`{"type":"tool_execution_start","toolName":"read","toolCallId":"evidence"}`)
		switch mode {
		case "stream-missing-tool-outcome":
			fmt.Println(`{"type":"tool_execution_end","toolName":"read","toolCallId":"evidence"}`)
		case "stream-failed-tool":
			fmt.Println(`{"type":"tool_execution_end","toolName":"read","toolCallId":"evidence","isError":true}`)
		default:
			fmt.Println(`{"type":"tool_execution_end","toolName":"read","toolCallId":"evidence","isError":false}`)
		}
		fmt.Println(`{"type":"turn_end"}`)
	}
	fmt.Println(`{"type":"turn_start"}`)
	fmt.Println(`{"type":"message_start","message":{"role":"assistant","content":[]}}`)
	stopReason := "stop"
	switch mode {
	case "stream-error-proof":
		stopReason = "error"
	case "stream-length-proof":
		stopReason = "length"
	}
	event := map[string]any{"type": "message_end", "message": map[string]any{"role": "assistant", "stopReason": stopReason, "content": []any{map[string]any{"type": "text", "text": text}}}}
	raw, _ := json.Marshal(event)
	switch mode {
	case "stream-duplicate-stop":
		raw = []byte(strings.Replace(string(raw), `"stopReason":"stop"`, `"stopReason":"error","stopReason":"stop"`, 1))
	case "stream-case-stop":
		raw = []byte(strings.Replace(string(raw), `"stopReason":"stop"`, `"stopReason":"error","StopReason":"stop"`, 1))
	}
	fmt.Println(string(raw))
	fmt.Println(`{"type":"turn_end"}`)
	switch mode {
	case "stream-agent-retry":
		fmt.Println(`{"type":"agent_end","willRetry":true}`)
	case "stream-agent-error":
		fmt.Println(`{"type":"agent_end","status":"error"}`)
	case "stream-duplicate-agent-retry":
		fmt.Println(`{"type":"agent_end","willRetry":true,"willRetry":false}`)
	case "stream-case-agent-retry":
		fmt.Println(`{"type":"agent_end","willRetry":true,"WillRetry":false}`)
	default:
		fmt.Println(`{"type":"agent_end"}`)
	}
	fmt.Println(`{"type":"agent_settled"}`)
}

func helperValidator(mode, stateDir, gsdHome string) GSDValidator {
	return GSDValidator{
		Command: []string{os.Args[0], "-test.run=TestGSDValidatorHelperProcess", "--"},
		GSDHome: gsdHome, StateDir: stateDir, Timeout: 5 * time.Second,
		Now: func() time.Time { return time.Now().UTC() }, Nonce: func() (string, error) { return testNonce, nil },
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
	if err := os.MkdirAll(filepath.Join(gsdHome, "agent"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatal(err)
	}
	artifactRaw, err := os.ReadFile(filepath.Join(repo, "candidate.txt"))
	if err != nil {
		t.Fatal(err)
	}
	artifactSum := sha256.Sum256(artifactRaw)
	candidateHead := gitForValidationTest(t, repo, "rev-parse", "HEAD")
	return Request{
		Repository: "polymetrics-ai/cli", PullRequest: 391, BaseBranch: "feat/372-gsd-pi-go-shepherd",
		DeliveryID: "issue-389", Generation: 2, UnitID: "M001/S01/T01", UnitType: "execute-task", Attempt: 1,
		StateVersion: 44, WorkDir: repo, GSDHome: gsdHome, StateDir: stateDir,
		BaseHead: candidateHead, CandidateHead: candidateHead,
		ContractHash: "sha256:" + strings.Repeat("c", 64), EvidenceHash: "sha256:" + strings.Repeat("e", 64),
		ArtifactHashes: []ArtifactHash{{Path: "candidate.txt", Hash: "sha256:" + hex.EncodeToString(artifactSum[:])}},
		RequireGates:   GateRequirements{LocalGates: true, UAT: false, MilestoneValid: false},
	}
}

func appendValidationSessionProof(sessionDir, sessionID, model, proof string) error {
	provider, modelID, ok := strings.Cut(model, "/")
	if !ok {
		return errors.New("invalid validation session model")
	}
	event := map[string]any{"type": "message", "message": map[string]any{
		"role": "assistant", "provider": provider, "model": modelID, "stopReason": "stop",
		"content": []any{map[string]any{"type": "text", "text": proof}},
	}}
	raw, err := json.Marshal(event)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(sessionDir, sessionID+".jsonl"), os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(append(raw, '\n'))
	return errors.Join(writeErr, file.Close())
}

func writeValidationSession(sessionDir, worktree, model, thinking, nonce string) error {
	if err := os.MkdirAll(sessionDir, 0o700); err != nil {
		return err
	}
	id := nonce[:8] + "-2222-3333-4444-555555555555"
	path := filepath.Join(sessionDir, id+".jsonl")
	provider, modelID, ok := strings.Cut(model, "/")
	if !ok {
		return fmt.Errorf("bad model %s", model)
	}
	content := fmt.Sprintf("{\"type\":\"session\",\"version\":3,\"timestamp\":%q,\"id\":%q,\"cwd\":%q}\n{\"type\":\"model_change\",\"provider\":%q,\"modelId\":%q}\n{\"type\":\"thinking_level_change\",\"thinkingLevel\":%q}\n{\"type\":\"message\",\"message\":{\"role\":\"assistant\",\"provider\":%q,\"model\":%q}}\n", time.Now().UTC().Format(time.RFC3339Nano), id, worktree, provider, modelID, thinking, provider, modelID)
	return os.WriteFile(path, []byte(content), 0o600)
}

func requestPathStateRoot(request protectedRequest) string {
	return os.Getenv("GSD_STATE_DIR")
}

func flagValue(args []string, name string) string {
	for i, arg := range args {
		if arg == name && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func containsArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
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
