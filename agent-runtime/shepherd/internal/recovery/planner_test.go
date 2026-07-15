package recovery

import (
	"context"
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

const plannerTestNonce = "0123456789abcdef0123456789abcdef"

func TestPiPlannerRequiresBoundSolHighEvidence(t *testing.T) {
	if os.Getenv("GO_WANT_RECOVERY_PLANNER_HELPER") != "" {
		return
	}
	planner := testPlanner(t, "success")
	request := testPlannerRequest()
	result, err := planner.Plan(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.SchemaVersion != SchemaVersion || result.RequestNonce != plannerTestNonce || result.DeliveryID != request.DeliveryID ||
		result.Generation != request.Generation || result.UnitID != request.UnitID || result.Attempt != request.Attempt ||
		result.HeadSHA != request.HeadSHA || result.FailureClass != request.Failure.Class || result.EvidenceHash != request.EvidenceHash ||
		result.AuthorityScopeHash != request.AuthorityScopeHash || result.ObservedModel != RequiredModel || result.Thinking != RequiredThinking ||
		result.SessionID == "" || result.SessionFingerprint == "" || result.Action != ActionRetryAfterBackoff ||
		len(result.BoundedPlanSteps) != 1 || result.BoundedPlanSteps[0].Primitive != PrimitiveRetryFreshAttempt {
		t.Fatalf("planner result=%+v", result)
	}
}

func TestPiPlannerRejectsStaticMalformedOversizedAndMismatchedOutput(t *testing.T) {
	if os.Getenv("GO_WANT_RECOVERY_PLANNER_HELPER") != "" {
		return
	}
	for _, mode := range []string{
		"static", "malformed", "oversized", "duplicate", "case-duplicate", "unknown-field", "trailing", "wrong-nonce",
		"wrong-head", "wrong-class", "wrong-evidence", "wrong-authority", "unknown-action", "forbidden-action",
		"executable-step", "wrong-backoff", "expired", "gpt55", "low-thinking", "stale-session", "no-session", "tool-attempt",
		"unknown-event", "missing-agent-end", "missing-message-start", "unmatched-message-update", "stream-session-mismatch", "hidden-session-tool", "hidden-tool-result-role", "hidden-tool-metadata", "additional-session",
	} {
		t.Run(mode, func(t *testing.T) {
			planner := testPlanner(t, mode)
			if _, err := planner.Plan(context.Background(), testPlannerRequest()); err == nil {
				t.Fatalf("planner mode %s was accepted", mode)
			}
		})
	}
}

func TestPiPlannerRejectsSymlinkedInvocationParent(t *testing.T) {
	if os.Getenv("GO_WANT_RECOVERY_PLANNER_HELPER") != "" {
		return
	}
	planner := testPlanner(t, "success")
	request := testPlannerRequest()
	recoveryRoot := filepath.Join(planner.config.StateDir, "recovery")
	if err := os.Mkdir(recoveryRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(recoveryRoot, requestPathID(request))); err != nil {
		t.Fatal(err)
	}
	if _, err := planner.Plan(context.Background(), request); err == nil {
		t.Fatal("symlinked recovery invocation parent was accepted")
	}
	if entries, err := os.ReadDir(outside); err != nil || len(entries) != 0 {
		t.Fatalf("symlink target was mutated: entries=%v err=%v", entries, err)
	}
}

func TestPiPlannerCapabilityProbeRunsOnlyInProtectedInvocationRoot(t *testing.T) {
	if os.Getenv("GO_WANT_RECOVERY_PLANNER_HELPER") != "" {
		return
	}
	planner := testPlanner(t, "probe-write")
	if _, err := planner.Plan(context.Background(), testPlannerRequest()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat("probe-side-effect"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("probe touched caller worktree: %v", err)
	}
	matches, err := filepath.Glob(filepath.Join(planner.config.StateDir, "recovery", "*", plannerTestNonce, "probe-side-effect"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("protected probe evidence=%v err=%v", matches, err)
	}
}

func TestPiPlannerRejectsNonceReplay(t *testing.T) {
	if os.Getenv("GO_WANT_RECOVERY_PLANNER_HELPER") != "" {
		return
	}
	planner := testPlanner(t, "success")
	if _, err := planner.Plan(context.Background(), testPlannerRequest()); err != nil {
		t.Fatal(err)
	}
	if _, err := planner.Plan(context.Background(), testPlannerRequest()); err == nil {
		t.Fatal("replayed planner nonce was accepted")
	}
}

func TestRecoveryPlannerHelperProcess(t *testing.T) {
	mode := os.Getenv("GO_WANT_RECOVERY_PLANNER_HELPER")
	if mode == "" {
		return
	}
	if hasPlannerArg(os.Args, "--help") {
		if mode == "probe-write" {
			_ = os.WriteFile("probe-side-effect", []byte("bounded"), 0o600)
		}
		fmt.Println("--mode --print --session-dir --no-tools --model --thinking")
		return
	}
	requestRaw := strings.TrimPrefix(plannerArg(os.Args, "--print"), recoveryPromptPrefix)
	var request plannerRequest
	if err := json.Unmarshal([]byte(requestRaw), &request); err != nil {
		os.Exit(2)
	}
	if plannerArg(os.Args, "--model") != RequiredModel || plannerArg(os.Args, "--thinking") != RequiredThinking || !hasPlannerArg(os.Args, "--no-tools") {
		os.Exit(3)
	}
	if mode == "no-session" {
		emitPlannerText(`{"schema_version":1}`)
		return
	}
	model, thinking := RequiredModel, RequiredThinking
	if mode == "gpt55" {
		model = "openai-codex/gpt-5.5"
	}
	if mode == "low-thinking" {
		thinking = "medium"
	}
	sessionPath, err := writePlannerSession(plannerArg(os.Args, "--session-dir"), request.SessionCWD, model, thinking)
	if err != nil {
		os.Exit(4)
	}
	if mode == "stale-session" {
		old := time.Now().UTC().Add(-time.Hour)
		_ = os.Chtimes(sessionPath, old, old)
	}
	if mode == "additional-session" {
		childID := "019f5d4a-9fb4-7852-b640-d6fdf71bd3d8"
		child := fmt.Sprintf("{\"type\":\"session\",\"version\":3,\"timestamp\":%q,\"id\":%q,\"cwd\":%q,\"parentSession\":%q}\n",
			time.Now().UTC().Format(time.RFC3339Nano), childID, request.SessionCWD, sessionPath)
		_ = os.WriteFile(filepath.Join(plannerArg(os.Args, "--session-dir"), childID+".jsonl"), []byte(child), 0o600)
	}
	var hiddenToolEvidence string
	switch mode {
	case "hidden-session-tool":
		hiddenToolEvidence = "{\"type\":\"message\",\"message\":{\"role\":\"assistant\",\"content\":[{\"type\":\"toolCall\"}]}}\n"
	case "hidden-tool-result-role":
		hiddenToolEvidence = "{\"type\":\"message\",\"message\":{\"role\":\"toolResult\",\"content\":[{\"type\":\"text\"}]}}\n"
	case "hidden-tool-metadata":
		hiddenToolEvidence = "{\"type\":\"message\",\"metadata\":{\"toolName\":\"bash\"}}\n"
	}
	if hiddenToolEvidence != "" {
		file, _ := os.OpenFile(sessionPath, os.O_APPEND|os.O_WRONLY, 0)
		_, _ = file.WriteString(hiddenToolEvidence)
		_ = file.Close()
	}
	if mode == "static" {
		emitPlannerText("Sol/high recovery planning required before next retry")
		return
	}
	result := recommendation{
		SchemaVersion:      SchemaVersion,
		RequestNonce:       request.RequestNonce,
		Issue:              request.Issue,
		DeliveryID:         request.DeliveryID,
		Generation:         request.Generation,
		UnitID:             request.UnitID,
		Attempt:            request.Attempt,
		HeadSHA:            request.HeadSHA,
		FailureClass:       request.FailureClass,
		EvidenceHash:       request.EvidenceHash,
		AuthorityScopeHash: request.AuthorityScopeHash,
		Action:             ActionRetryAfterBackoff,
		BoundedPlanSteps:   []PlanStep{{Primitive: PrimitiveRetryFreshAttempt}},
		BackoffMS:          request.ControllerBackoffMS,
		IssuedAt:           request.IssuedAt,
		ExpiresAt:          request.ExpiresAt,
	}
	switch mode {
	case "wrong-nonce":
		result.RequestNonce = strings.Repeat("f", 32)
	case "wrong-head":
		result.HeadSHA = strings.Repeat("f", 40)
	case "wrong-class":
		result.FailureClass = FailureDeadWorker
	case "wrong-evidence":
		result.EvidenceHash = "sha256:" + strings.Repeat("f", 64)
	case "wrong-authority":
		result.AuthorityScopeHash = "sha256:" + strings.Repeat("f", 64)
	case "unknown-action":
		result.Action = Action("execute_shell")
	case "forbidden-action":
		result.Action = ActionRetrySameUnit
	case "executable-step":
		result.Action = ActionRunRecoveryPlan
		result.BoundedPlanSteps = []PlanStep{{Primitive: Primitive("shell")}}
	case "wrong-backoff":
		result.BackoffMS++
	case "expired":
		result.ExpiresAt = request.IssuedAt.Add(-time.Second)
	}
	raw, _ := json.Marshal(result)
	if mode == "tool-attempt" {
		fmt.Println(`{"type":"tool_execution_start","toolName":"read"}`)
	}
	switch mode {
	case "malformed":
		emitPlannerText("{not-json")
	case "oversized":
		emitPlannerText(strings.Repeat("x", maxPlannerOutputBytes+1))
	case "duplicate":
		emitPlannerText(strings.Replace(string(raw), `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1))
	case "case-duplicate":
		emitPlannerText(strings.TrimSuffix(string(raw), "}") + `,"Action":"block"}`)
	case "unknown-field":
		emitPlannerText(strings.TrimSuffix(string(raw), "}") + `,"command":"rm -rf"}`)
	case "trailing":
		emitPlannerText(string(raw) + " trailing")
	default:
		emitPlannerText(string(raw))
	}
}

func testPlanner(t *testing.T, mode string) *PiPlanner {
	t.Helper()
	stateDir := t.TempDir()
	gsdHome := t.TempDir()
	planner, err := NewPiPlanner(PiPlannerConfig{
		Command:     []string{os.Args[0], "-test.run=TestRecoveryPlannerHelperProcess", "--"},
		GSDHome:     gsdHome,
		StateDir:    stateDir,
		Timeout:     3 * time.Second,
		Now:         func() time.Time { return time.Unix(1_800_000_000, 0).UTC() },
		Nonce:       func() (string, error) { return plannerTestNonce, nil },
		Environment: []string{"GO_WANT_RECOVERY_PLANNER_HELPER=" + mode},
	})
	if err != nil {
		t.Fatal(err)
	}
	return planner
}

func testPlannerRequest() Request {
	return Request{
		Issue:              389,
		DeliveryID:         "issue-389",
		Generation:         4,
		UnitID:             "execute-task/M001/S01/T01",
		Attempt:            2,
		HeadSHA:            strings.Repeat("a", 40),
		Failure:            Failure{Class: FailureSilentTool, Reversible: true},
		EvidenceHash:       "sha256:" + strings.Repeat("b", 64),
		AuthorityScopeHash: "sha256:" + strings.Repeat("c", 64),
		ControllerBackoff:  2 * time.Second,
	}
}

func writePlannerSession(root, workDir, model, thinking string) (string, error) {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return "", err
	}
	id := "019f5d4a-9fb4-7852-b640-d6fdf71bd3d9"
	provider, modelID, _ := strings.Cut(model, "/")
	path := filepath.Join(root, id+".jsonl")
	content := fmt.Sprintf("{\"type\":\"session\",\"version\":3,\"timestamp\":%q,\"id\":%q,\"cwd\":%q}\n{\"type\":\"model_change\",\"provider\":%q,\"modelId\":%q}\n{\"type\":\"thinking_level_change\",\"thinkingLevel\":%q}\n{\"type\":\"message\",\"message\":{\"role\":\"assistant\",\"provider\":%q,\"model\":%q}}\n", time.Now().UTC().Format(time.RFC3339Nano), id, workDir, provider, modelID, thinking, provider, modelID)
	return path, os.WriteFile(path, []byte(content), 0o600)
}

func emitPlannerText(text string) {
	mode := os.Getenv("GO_WANT_RECOVERY_PLANNER_HELPER")
	sessionID := "019f5d4a-9fb4-7852-b640-d6fdf71bd3d9"
	if mode == "stream-session-mismatch" {
		sessionID = "019f5d4a-9fb4-7852-b640-d6fdf71bd3d8"
	}
	fmt.Printf("{\"type\":\"session\",\"id\":%q}\n", sessionID)
	fmt.Println(`{"type":"agent_start"}`)
	fmt.Println(`{"type":"turn_start"}`)
	if mode == "unmatched-message-update" {
		fmt.Println(`{"type":"message_update"}`)
	}
	if mode == "unknown-event" {
		fmt.Println(`{"type":"unqualified_event"}`)
	}
	if mode != "missing-message-start" {
		fmt.Println(`{"type":"message_start","message":{"role":"assistant","content":[]}}`)
	}
	event := map[string]any{"type": "message_end", "message": map[string]any{"role": "assistant", "content": []map[string]string{{"type": "text", "text": text}}}}
	raw, _ := json.Marshal(event)
	fmt.Println(string(raw))
	fmt.Println(`{"type":"turn_end"}`)
	if mode != "missing-agent-end" {
		fmt.Println(`{"type":"agent_end"}`)
		fmt.Println(`{"type":"agent_settled"}`)
	}
	os.Exit(0)
}

func plannerArg(args []string, name string) string {
	for index, arg := range args {
		if arg == name && index+1 < len(args) {
			return args[index+1]
		}
	}
	return ""
}

func hasPlannerArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

func TestLivePiRecoveryPlannerSmoke(t *testing.T) {
	if os.Getenv("POLYMETRICS_SHEPHERD_LIVE_RECOVERY") != "1" {
		t.Skip("set POLYMETRICS_SHEPHERD_LIVE_RECOVERY=1 for the opt-in live Pi smoke")
	}
	piPath, err := exec.LookPath("pi")
	if err != nil {
		t.Fatal(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	planner, err := NewPiPlanner(PiPlannerConfig{
		Command:  []string{piPath},
		GSDHome:  filepath.Join(home, ".pi"),
		StateDir: t.TempDir(),
		Timeout:  2 * time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := planner.Plan(context.Background(), testPlannerRequest())
	if err != nil {
		t.Fatal(err)
	}
	if result.ObservedModel != RequiredModel || result.Thinking != RequiredThinking || result.SessionID == "" || result.SessionFingerprint == "" {
		t.Fatalf("live recovery evidence=%+v", result)
	}
	t.Logf("live recovery planner model=%s thinking=%s session=%s action=%s evidence=%s", result.ObservedModel, result.Thinking, result.SessionID, result.Action, result.PlannerEvidenceHash)
}
