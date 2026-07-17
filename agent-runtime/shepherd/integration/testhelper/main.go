//go:build integration

package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	implementationModel = "openai-codex/gpt-5.5"
	validatorModel      = "openai-codex/gpt-5.6-sol"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "integration helper:", err)
		os.Exit(2)
	}
}

func run(args []string) error {
	role := processRole()
	switch {
	case len(args) > 0 && args[0] == "api":
		if role != "github" {
			return errors.New("non-GitHub fake received GitHub argv")
		}
		return runGitHub(args)
	case validHelpArgs(args):
		if role != "pi" {
			return errors.New("non-Pi fake received help argv")
		}
		fmt.Println("--mode --print --session-dir --tools --model --thinking --no-tools")
		return nil
	case value(args, "--mode") == "json" && contains(args, "--no-tools"):
		if role != "pi" {
			return errors.New("non-Pi fake received recovery argv")
		}
		if err := validateRecoveryArgs(args); err != nil {
			return err
		}
		return runRecoveryPlanner(args)
	case value(args, "--mode") == "json":
		if role != "pi" {
			return errors.New("non-Pi fake received validator argv")
		}
		if err := validateValidatorArgs(args); err != nil {
			return err
		}
		return runValidator(args)
	case equalStrings(args, []string{"headless", "query"}):
		if role != "gsd" {
			return errors.New("non-GSD fake received query argv")
		}
		return runQuery()
	case len(args) > 0 && args[0] == "headless":
		if role != "gsd" {
			return errors.New("non-GSD fake received unit argv")
		}
		if err := validateUnitArgs(args); err != nil {
			return err
		}
		return runUnit(args)
	default:
		return fmt.Errorf("unsupported argv shape %q", args)
	}
}

func runQuery() error {
	if scenarioName() == "unknown-unit" {
		fmt.Print(`{"state":{"activeMilestone":{"id":"M001"},"phase":"execution","nextAction":"dispatch","blockers":[]},"next":{"action":"dispatch","unitType":"invented-unit","unitId":"M001/S01/T01"}}`)
		return nil
	}
	projectRoot := os.Getenv("GSD_PROJECT_ROOT")
	if !filepath.IsAbs(projectRoot) {
		return errors.New("query project root must be absolute")
	}
	state, err := os.ReadFile(filepath.Join(projectRoot, ".gsd", "STATE.md"))
	if err == nil && strings.Contains(string(state), "process-boundary completed state") {
		fmt.Print(`{"state":{"activeMilestone":{"id":"M001"},"phase":"complete","nextAction":"stop","blockers":[]},"next":{"action":"stop"}}`)
		return nil
	}
	if scenarioName() == "no-candidate-change" || scenarioName() == "missing-candidate-artifact" {
		if _, markerErr := os.Stat(candidateCompletionMarker(projectRoot)); markerErr == nil {
			fmt.Print(`{"state":{"activeMilestone":{"id":"M001"},"phase":"complete","nextAction":"stop","blockers":[]},"next":{"action":"stop"}}`)
			return nil
		} else if !errors.Is(markerErr, os.ErrNotExist) {
			return markerErr
		}
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if scenarioName() == "planning-route" {
		fmt.Print(`{"state":{"activeMilestone":{"id":"M001"},"phase":"planning","nextAction":"dispatch","blockers":[]},"next":{"action":"dispatch","unitType":"plan-milestone","unitId":"M001"}}`)
		return nil
	}
	fmt.Print(`{"state":{"activeMilestone":{"id":"M001"},"activeSlice":{"id":"S01"},"activeTask":{"id":"T01"},"phase":"execution","nextAction":"dispatch","blockers":[]},"next":{"action":"dispatch","unitType":"execute-task","unitId":"M001/S01/T01"}}`)
	return nil
}

func resumeAuthorized() bool {
	stateDir := os.Getenv("GSD_STATE_DIR")
	if stateDir == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(stateDir, "integration-decision-resume"))
	return err == nil
}

func runUnit(args []string) error {
	runtimeModel := implementationModel
	if scenarioName() == "planning-route" {
		runtimeModel = validatorModel
	}
	if value(args, "--model") != runtimeModel {
		return fmt.Errorf("runtime model=%q want %q", value(args, "--model"), runtimeModel)
	}
	workDir := os.Getenv("GSD_PROJECT_ROOT")
	if !filepath.IsAbs(workDir) {
		return errors.New("GSD_PROJECT_ROOT must be absolute")
	}
	if scenarioName() == "running-termination" {
		return blockRunningUnit()
	}
	if scenarioName() == "always-fail" || (scenarioName() == "decision-resume" && !resumeAuthorized()) {
		if err := writeGovernedSession(workDir, runtimeModel); err != nil {
			return err
		}
		fmt.Println(`{"type":"agent_start"}`)
		fmt.Println(`{"type":"model_select","model":{"provider":"openai-codex","id":"gpt-5.5"}}`)
		fmt.Println(`{"type":"thinking_level_select","level":"high"}`)
		return errors.New("bounded exhausted process failure")
	}
	if scenarioName() == "recoverable-failure" {
		firstAttempt := filepath.Join(os.Getenv("GSD_STATE_DIR"), "integration-first-attempt-failed")
		if _, err := os.Stat(firstAttempt); errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(os.Getenv("GSD_STATE_DIR"), 0o700); err != nil {
				return err
			}
			if err := os.WriteFile(firstAttempt, []byte("failed\n"), 0o600); err != nil {
				return err
			}
			if err := writeGovernedSession(workDir, runtimeModel); err != nil {
				return err
			}
			fmt.Println(`{"type":"agent_start"}`)
			fmt.Println(`{"type":"model_select","model":{"provider":"openai-codex","id":"gpt-5.5"}}`)
			fmt.Println(`{"type":"thinking_level_select","level":"high"}`)
			return errors.New("bounded first-attempt process failure")
		} else if err != nil {
			return err
		}
	}
	if scenarioName() == "no-candidate-change" {
		return finishImplementation(workDir, runtimeModel)
	}
	artifact := filepath.Join(workDir, "agent-runtime", "shepherd", "integration-artifact.txt")
	if strings.HasPrefix(scenarioName(), "tracked-deletion") {
		if err := os.Remove(artifact); err != nil {
			return err
		}
	} else if scenarioName() != "missing-candidate-artifact" && scenarioName() != "gsd-state-only" {
		if err := os.MkdirAll(filepath.Dir(artifact), 0o700); err != nil {
			return err
		}
		artifactBody := "process-boundary artifact\n"
		if scenarioName() == "recoverable-validation-failure" {
			artifactBody = "process-boundary artifact " + filepath.Base(workDir) + "\n"
		}
		if err := os.WriteFile(artifact, []byte(artifactBody), 0o600); err != nil {
			return err
		}
	}
	gsdState := filepath.Join(workDir, ".gsd", "STATE.md")
	if scenarioName() == "missing-candidate-artifact" {
		if err := os.Remove(gsdState); err != nil {
			return err
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(gsdState), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(gsdState, []byte("process-boundary completed state\n"), 0o600); err != nil {
			return err
		}
	}
	return finishImplementation(workDir, runtimeModel)
}

func finishImplementation(workDir, runtimeModel string) error {
	stateDir := os.Getenv("GSD_STATE_DIR")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		return err
	}
	if scenarioName() == "no-candidate-change" || scenarioName() == "missing-candidate-artifact" {
		if err := os.WriteFile(candidateCompletionMarker(workDir), []byte("complete\n"), 0o600); err != nil {
			return err
		}
	}
	if err := appendObservation(stateDir, map[string]any{
		"role": "governed-unit", "model": runtimeModel, "thinking": "high", "work_dir": workDir,
	}); err != nil {
		return err
	}
	if err := writeGovernedSession(workDir, runtimeModel); err != nil {
		return err
	}
	fmt.Println(`{"type":"agent_start"}`)
	if scenarioName() != "implementation-missing-turn" {
		fmt.Println(`{"type":"turn_start"}`)
	}
	fmt.Printf("{\"type\":\"model_select\",\"model\":{\"provider\":\"openai-codex\",\"id\":%q}}\n",
		strings.TrimPrefix(runtimeModel, "openai-codex/"))
	fmt.Println(`{"type":"thinking_level_select","level":"high"}`)
	if scenarioName() != "implementation-no-workflow-tool" {
		workflowTools := []string{"gsd_task_complete", "gsd_exec", "gsd_exec_search", "gsd_resume", "gsd_capture_thought"}
		if scenarioName() == "planning-route" {
			workflowTools = []string{"gsd_milestone_status", "gsd_plan_milestone", "gsd_plan_slice", "gsd_plan_task"}
		} else if scenarioName() == "implementation-partial-workflow" {
			workflowTools = workflowTools[1:]
		}
		for index, workflowTool := range workflowTools {
			callID := fmt.Sprintf("workflow-transition-%d", index)
			fmt.Printf("{\"type\":\"tool_execution_start\",\"toolName\":\"mcp__gsd-workflow__%s\",\"toolCallId\":%q}\n", workflowTool, callID)
			fmt.Printf("{\"type\":\"tool_execution_end\",\"toolName\":\"mcp__gsd-workflow__%s\",\"toolCallId\":%q,\"isError\":false}\n", workflowTool, callID)
		}
	}
	if scenarioName() != "implementation-missing-turn" {
		stopReason := "stop"
		if scenarioName() == "implementation-error-stop" {
			stopReason = "error"
		}
		fmt.Printf("{\"type\":\"turn_end\",\"message\":{\"role\":\"assistant\",\"stopReason\":%q}}\n", stopReason)
	}
	if scenarioName() == "implementation-retrying-end" {
		fmt.Println(`{"type":"agent_end","status":"success","willRetry":true}`)
	} else if scenarioName() == "implementation-agent-error" {
		fmt.Println(`{"type":"agent_end","status":"error"}`)
	} else {
		fmt.Println(`{"type":"agent_end","status":"success"}`)
	}
	return nil
}

func writeGovernedSession(workDir, runtimeModel string) error {
	sessionModel, sessionThinking := runtimeModel, "high"
	if scenarioName() == "wrong-implementation-session-model" {
		sessionModel = validatorModel
	}
	if scenarioName() == "wrong-implementation-session-thinking" {
		sessionThinking = "medium"
	}
	digest := sha256.Sum256([]byte(workDir))
	sessionID := "019f7000-1111-7111-8111-" + hex.EncodeToString(digest[:6])
	return writeSession(filepath.Join(os.Getenv("HOME"), "agent", "sessions"), workDir,
		sessionID, sessionModel, sessionThinking)
}

type validationRequest struct {
	RequestID     string `json:"request_id"`
	Nonce         string `json:"nonce"`
	Repository    string `json:"repository"`
	PullRequest   int    `json:"pull_request"`
	BaseBranch    string `json:"base_branch"`
	BaseHead      string `json:"base_head"`
	CandidateHead string `json:"candidate_head"`
	StateVersion  int64  `json:"state_version"`
	ContractHash  string `json:"contract_hash"`
	EvidenceHash  string `json:"evidence_hash"`
	RequiredGates struct {
		LocalGates     bool `json:"local_gates"`
		UAT            bool `json:"uat"`
		MilestoneValid bool `json:"milestone_valid"`
	} `json:"required_gates"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func runValidator(args []string) error {
	if value(args, "--model") != validatorModel || value(args, "--thinking") != "high" ||
		value(args, "--tools") != "read,grep,find,ls" {
		return errors.New("validator invocation is not isolated Sol/high")
	}
	scenario := scenarioName()
	if scenario == "recoverable-validation-failure" {
		executable, err := os.Executable()
		if err != nil {
			return err
		}
		marker := filepath.Join(filepath.Dir(executable), "integration-validation-failed-once")
		if _, err := os.Stat(marker); errors.Is(err, os.ErrNotExist) {
			if err := os.WriteFile(marker, []byte("failed\n"), 0o600); err != nil {
				return err
			}
			time.Sleep(time.Minute)
			return errors.New("validation timeout was not enforced")
		} else if err != nil {
			return err
		}
	}
	if scenario == "missing-validator-evidence" {
		fmt.Println(`{"type":"agent_end"}`)
		return nil
	}
	request, err := requestFromPrompt(value(args, "--print"))
	if err != nil {
		return err
	}
	if scenario == "candidate-moved" {
		if err := moveCandidateHead(os.Getenv("GSD_PROJECT_ROOT")); err != nil {
			return err
		}
	}
	if scenario == "artifact-changed" {
		if err := os.Remove(filepath.Join(os.Getenv("GSD_PROJECT_ROOT"), "agent-runtime", "shepherd",
			"integration-artifact.txt")); err != nil {
			return err
		}
	}
	if scenario == "tracked-deletion-recreated" {
		path := filepath.Join(os.Getenv("GSD_PROJECT_ROOT"), "agent-runtime", "shepherd", "integration-artifact.txt")
		if err := os.WriteFile(path, []byte("recreated during validation\n"), 0o600); err != nil {
			return err
		}
	}
	model := validatorModel
	if scenario == "gpt55-validator" {
		model = implementationModel
	}
	thinking := "high"
	if scenario == "non-high-validator" {
		thinking = "medium"
	}
	const validatorSessionID = "019f7000-2222-7222-8222-222222222222"
	if err := writeSession(value(args, "--session-dir"), os.Getenv("GSD_PROJECT_ROOT"),
		validatorSessionID, model, thinking); err != nil {
		return err
	}
	verdict := "PROCEED"
	switch scenario {
	case "validator-retry":
		verdict = "RETRY"
	case "validator-halt":
		verdict = "HALT"
	}
	observedHead := request.CandidateHead
	if scenario == "stale-validator-head" {
		observedHead = request.BaseHead
	}
	localGates := request.RequiredGates.LocalGates
	if scenario == "missing-required-gate" {
		localGates = false
	}
	proof := map[string]any{
		"request_id": request.RequestID, "nonce": request.Nonce,
		"repository": request.Repository, "pull_request": request.PullRequest,
		"base_branch": request.BaseBranch, "base_head": request.BaseHead,
		"candidate_head": request.CandidateHead, "observed_head": observedHead,
		"state_version": request.StateVersion, "contract_hash": request.ContractHash,
		"evidence_hash": request.EvidenceHash, "verdict": verdict,
		"local_gates": localGates, "uat": request.RequiredGates.UAT,
		"milestone_valid": request.RequiredGates.MilestoneValid,
		"issued_at":       request.IssuedAt, "expires_at": request.ExpiresAt,
	}
	raw, err := json.Marshal(proof)
	if err != nil {
		return err
	}
	if err := appendSessionProof(value(args, "--session-dir"), validatorSessionID, model, string(raw)); err != nil {
		return err
	}
	agentEnd := map[string]any{"type": "agent_end", "status": "success"}
	if scenario == "validator-agent-error" {
		agentEnd["status"] = "error"
	}
	events := []map[string]any{
		{"type": "session", "id": validatorSessionID}, {"type": "agent_start"}, {"type": "turn_start"},
		{"type": "message_start", "message": map[string]any{"role": "assistant", "content": []any{}}},
		{"type": "message_end", "message": map[string]any{"role": "assistant", "stopReason": "toolUse", "content": []any{}}},
		{"type": "tool_execution_start", "toolCallId": "validator-read", "toolName": "read"},
		{"type": "tool_execution_end", "toolCallId": "validator-read", "toolName": "read", "isError": false},
		{"type": "turn_end"}, {"type": "turn_start"},
		{"type": "message_start", "message": map[string]any{"role": "assistant", "content": []any{}}},
		{"type": "message_end", "message": map[string]any{
			"role": "assistant", "stopReason": "stop",
			"content": []any{map[string]any{"type": "text", "text": string(raw)}},
		}},
		{"type": "turn_end"}, agentEnd, {"type": "agent_settled"},
	}
	encoder := json.NewEncoder(os.Stdout)
	for _, event := range events {
		if err := encoder.Encode(event); err != nil {
			return err
		}
	}
	return nil
}

type recoveryRequest struct {
	SchemaVersion       int       `json:"schema_version"`
	RequestNonce        string    `json:"request_nonce"`
	Issue               int       `json:"issue"`
	DeliveryID          string    `json:"delivery_id"`
	Generation          int64     `json:"generation"`
	UnitID              string    `json:"unit_id"`
	Attempt             int64     `json:"attempt"`
	HeadSHA             string    `json:"head_sha"`
	FailureClass        string    `json:"failure_class"`
	EvidenceHash        string    `json:"evidence_hash"`
	AuthorityScopeHash  string    `json:"authority_scope_hash"`
	ControllerBackoffMS int64     `json:"controller_backoff_ms"`
	IssuedAt            time.Time `json:"issued_at"`
	ExpiresAt           time.Time `json:"expires_at"`
	SessionCWD          string    `json:"session_cwd"`
}

func runRecoveryPlanner(args []string) error {
	if value(args, "--model") != validatorModel || value(args, "--thinking") != "high" || !contains(args, "--no-tools") {
		return errors.New("recovery planner invocation is not isolated Sol/high")
	}
	request, err := recoveryRequestFromPrompt(value(args, "--print"))
	if err != nil {
		return err
	}
	const sessionID = "019f7000-3333-7333-8333-333333333333"
	if err := writeSession(value(args, "--session-dir"), request.SessionCWD, sessionID,
		validatorModel, "high"); err != nil {
		return err
	}
	recommendation := map[string]any{
		"schema_version": request.SchemaVersion, "request_nonce": request.RequestNonce,
		"issue": request.Issue, "delivery_id": request.DeliveryID,
		"generation": request.Generation, "unit_id": request.UnitID, "attempt": request.Attempt,
		"head_sha": request.HeadSHA, "failure_class": request.FailureClass,
		"evidence_hash": request.EvidenceHash, "authority_scope_hash": request.AuthorityScopeHash,
		"action":             "retry_same_unit",
		"bounded_plan_steps": []map[string]any{{"primitive": "retry_fresh_attempt"}},
		"backoff_ms":         request.ControllerBackoffMS,
		"issued_at":          request.IssuedAt, "expires_at": request.ExpiresAt,
	}
	raw, err := json.Marshal(recommendation)
	if err != nil {
		return err
	}
	events := []map[string]any{
		{"type": "session", "id": sessionID}, {"type": "agent_start"}, {"type": "turn_start"},
		{"type": "message_start", "message": map[string]any{"role": "assistant"}},
		{"type": "message_end", "message": map[string]any{"role": "assistant",
			"content": []map[string]any{{"type": "text", "text": string(raw)}}}},
		{"type": "turn_end"}, {"type": "agent_end"}, {"type": "agent_settled"},
	}
	encoder := json.NewEncoder(os.Stdout)
	for _, event := range events {
		if err := encoder.Encode(event); err != nil {
			return err
		}
	}
	return nil
}

func recoveryRequestFromPrompt(prompt string) (recoveryRequest, error) {
	const marker = "Request JSON: "
	index := strings.LastIndex(prompt, marker)
	if index < 0 {
		return recoveryRequest{}, errors.New("recovery prompt has no request JSON")
	}
	var request recoveryRequest
	if err := json.Unmarshal([]byte(prompt[index+len(marker):]), &request); err != nil {
		return recoveryRequest{}, err
	}
	return request, nil
}

func requestFromPrompt(prompt string) (validationRequest, error) {
	const marker = "Request JSON: "
	index := strings.LastIndex(prompt, marker)
	if index < 0 {
		return validationRequest{}, errors.New("validator prompt has no request JSON")
	}
	var request validationRequest
	if err := json.Unmarshal([]byte(prompt[index+len(marker):]), &request); err != nil {
		return validationRequest{}, err
	}
	return request, nil
}

func appendSessionProof(directory, id, model, proof string) error {
	provider, modelID, ok := strings.Cut(model, "/")
	if !ok {
		return errors.New("session model must be provider-qualified")
	}
	event := map[string]any{"type": "message", "message": map[string]any{
		"role": "assistant", "provider": provider, "model": modelID, "stopReason": "stop",
		"content": []any{map[string]any{"type": "text", "text": proof}},
	}}
	file, err := os.OpenFile(filepath.Join(directory, id+".jsonl"), os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	encodeErr := json.NewEncoder(file).Encode(event)
	return errors.Join(encodeErr, file.Close())
}

func writeSession(directory, workDir, id, model, thinking string) error {
	if !filepath.IsAbs(directory) || !filepath.IsAbs(workDir) {
		return errors.New("session directory and worktree must be absolute")
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	provider, modelID, ok := strings.Cut(model, "/")
	if !ok {
		return errors.New("session model must be provider-qualified")
	}
	rows := []map[string]any{
		{"type": "session", "version": 3, "timestamp": now, "id": id, "cwd": workDir},
		{"type": "model_change", "provider": provider, "modelId": modelID},
		{"type": "thinking_level_change", "thinkingLevel": thinking},
		{"type": "message", "message": map[string]any{"role": "assistant", "provider": provider, "model": modelID, "stopReason": "stop"}},
	}
	path := filepath.Join(directory, id+".jsonl")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	for _, row := range rows {
		if err := json.NewEncoder(writer).Encode(row); err != nil {
			_ = file.Close()
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func appendObservation(directory string, observation map[string]any) error {
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(directory, "integration-observations.jsonl"),
		os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	encodeErr := json.NewEncoder(file).Encode(observation)
	closeErr := file.Close()
	return errors.Join(encodeErr, closeErr)
}

type githubComment struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	User      struct {
		Login string `json:"login"`
		ID    int64  `json:"id"`
		Type  string `json:"type"`
	} `json:"user"`
}

func runGitHub(args []string) error {
	if len(args) < 2 {
		return errors.New("fake GitHub endpoint is required")
	}
	statePath := os.Getenv("SHEPHERD_INTEGRATION_FAKE_GITHUB_STATE")
	if !filepath.IsAbs(statePath) {
		return errors.New("fake GitHub state path must be absolute")
	}
	if equalStrings(args, []string{"api", "user", "--method", "GET"}) {
		if err := recordGitHubInvocation(statePath, args, nil); err != nil {
			return err
		}
		fmt.Print(`{"login":"shepherd-bot","type":"Bot"}`)
		return nil
	}
	const commentsEndpoint = "repos/polymetrics-ai/cli/issues/391/comments"
	method := value(args, "--method")
	isList := equalStrings(args, []string{"api", commentsEndpoint, "--paginate", "--slurp", "--method", "GET"})
	isCreate := equalStrings(args, []string{"api", commentsEndpoint, "--method", "POST", "--input", "-"})
	updatePrefix := "repos/polymetrics-ai/cli/issues/comments/"
	updateID := ""
	if len(args) > 1 {
		updateID = strings.TrimPrefix(args[1], updatePrefix)
	}
	parsedUpdateID, updateIDErr := strconv.ParseInt(updateID, 10, 64)
	isUpdate := len(args) == 6 && args[0] == "api" && strings.HasPrefix(args[1], updatePrefix) &&
		updateIDErr == nil && parsedUpdateID > 0 && args[2] == "--method" && args[3] == "PATCH" &&
		args[4] == "--input" && args[5] == "-"
	if !isList && !isCreate && !isUpdate {
		return fmt.Errorf("fake GitHub rejected non-comment or wrong-target invocation")
	}
	comments, err := readGitHubComments(statePath)
	if err != nil {
		return err
	}
	if isList {
		if err := recordGitHubInvocation(statePath, args, nil); err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode([][]githubComment{comments})
	}
	payload, err := readCommentPayload()
	if err != nil {
		return err
	}
	if err := recordGitHubInvocation(statePath, args, []byte(payload.Body)); err != nil {
		return err
	}
	switch method {
	case "POST":
		comment := newBotComment(101, payload.Body)
		if len(comments) > 0 {
			comment.ID = comments[len(comments)-1].ID + 1
		}
		comments = append(comments, comment)
		if err := writeGitHubComments(statePath, comments); err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(comment)
	case "PATCH":
		index := -1
		for i := range comments {
			if comments[i].ID == parsedUpdateID {
				index = i
				break
			}
		}
		if index < 0 {
			return errors.New("fake GitHub update target does not exist")
		}
		comments[index].Body = payload.Body
		comments[index].UpdatedAt = comments[index].CreatedAt
		if err := writeGitHubComments(statePath, comments); err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(comments[index])
	default:
		return fmt.Errorf("unsupported fake GitHub method %q", method)
	}
}

type commentPayload struct {
	Body string `json:"body"`
}

func readCommentPayload() (commentPayload, error) {
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, 64*1024+1))
	if err != nil || len(raw) == 0 || len(raw) > 64*1024 {
		return commentPayload{}, errors.New("fake GitHub comment payload is empty or oversized")
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return commentPayload{}, errors.New("fake GitHub comment payload must be an object")
	}
	var payload commentPayload
	fields := 0
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil || key != "body" || fields != 0 {
			return commentPayload{}, errors.New("fake GitHub comment payload has duplicate or unknown fields")
		}
		if err := decoder.Decode(&payload.Body); err != nil || payload.Body == "" {
			return commentPayload{}, errors.New("fake GitHub comment body is required")
		}
		fields++
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') || fields != 1 {
		return commentPayload{}, errors.New("fake GitHub comment payload must contain one body")
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return commentPayload{}, errors.New("fake GitHub comment payload has trailing data")
	}
	return payload, nil
}

func recordGitHubInvocation(statePath string, args []string, body []byte) error {
	digest := ""
	if body != nil {
		hash := sha256.Sum256(body)
		digest = "sha256:" + hex.EncodeToString(hash[:])
	}
	record, err := json.Marshal(map[string]any{"args": args, "body_digest": digest})
	if err != nil {
		return err
	}
	file, err := os.OpenFile(statePath+".calls", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(append(record, '\n'))
	closeErr := file.Close()
	return errors.Join(writeErr, closeErr)
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func readGitHubComments(path string) ([]githubComment, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var comments []githubComment
	if err := json.Unmarshal(raw, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func writeGitHubComments(path string, comments []githubComment) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, err := json.Marshal(comments)
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}

func newBotComment(id int64, body string) githubComment {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	comment := githubComment{ID: id, Body: body, CreatedAt: now, UpdatedAt: now}
	comment.User.Login, comment.User.ID, comment.User.Type = "shepherd-bot", 9001, "Bot"
	return comment
}

func blockRunningUnit() error {
	stateDir := os.Getenv("GSD_STATE_DIR")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		return err
	}
	fifo := filepath.Join(stateDir, "integration-running-release")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(stateDir, "integration-running-ready"), []byte("ready\n"), 0o600); err != nil {
		return err
	}
	file, err := os.Open(fifo)
	if err != nil {
		return err
	}
	_, readErr := io.Copy(io.Discard, io.LimitReader(file, 1))
	closeErr := file.Close()
	return errors.Join(errors.New("running integration fixture released"), readErr, closeErr)
}

func moveCandidateHead(workDir string) error {
	path := filepath.Join(workDir, "agent-runtime", "shepherd", "validator-move.txt")
	if err := os.WriteFile(path, []byte("moved during validation\n"), 0o600); err != nil {
		return err
	}
	for _, args := range [][]string{{"add", "agent-runtime/shepherd/validator-move.txt"},
		{"commit", "-qm", "test: move candidate during validation"}} {
		cmd := exec.Command("git", append([]string{"-C", workDir}, args...)...)
		if raw, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git %v: %w: %s", args, err, raw)
		}
	}
	return nil
}

func validHelpArgs(args []string) bool {
	return equalStrings(args, []string{"--help"}) || equalStrings(args, []string{
		"--no-tools", "--no-extensions", "--no-skills", "--no-prompt-templates",
		"--no-themes", "--no-context-files", "--no-approve", "--help",
	})
}

func validateUnitArgs(args []string) error {
	if len(args) != 14 || args[0] != "headless" || args[1] != "--json" || args[2] != "--supervised" ||
		args[3] != "--model" || args[5] != "--response-timeout" || args[7] != "--max-restarts" ||
		args[8] != "0" || args[9] != "--events" ||
		args[10] != "agent_start,agent_end,turn_start,turn_end,tool_execution_start,tool_execution_end,model_select,thinking_level_select,extension_ui_request" ||
		args[11] != "--timeout" || (args[13] != "execute-task" && args[13] != "plan-milestone") {
		return errors.New("fake GSD unit rejected noncanonical argv")
	}
	for _, index := range []int{6, 12} {
		value, err := strconv.ParseInt(args[index], 10, 64)
		if err != nil || value <= 0 {
			return errors.New("fake GSD unit requires positive bounded timeouts")
		}
	}
	return nil
}

func validateValidatorArgs(args []string) error {
	if len(args) != 22 || args[0] != "--mode" || args[1] != "json" ||
		args[2] != "--system-prompt" || args[3] == "" || args[4] != "--model" || args[5] != validatorModel ||
		args[6] != "--thinking" || args[7] != "high" || args[8] != "--tools" ||
		args[9] != "read,grep,find,ls" || args[10] != "--no-extensions" || args[11] != "--no-skills" ||
		args[12] != "--no-prompt-templates" || args[13] != "--no-themes" ||
		args[14] != "--no-context-files" || args[15] != "--no-approve" || args[16] != "--session-dir" ||
		!filepath.IsAbs(args[17]) || args[18] != "--name" || !strings.HasPrefix(args[19], "shepherd-validator-") ||
		args[20] != "--print" || args[21] == "" {
		return errors.New("fake validator rejected noncanonical isolated argv")
	}
	return nil
}

func validateRecoveryArgs(args []string) error {
	if len(args) != 21 || args[0] != "--mode" || args[1] != "json" ||
		args[2] != "--system-prompt" || args[3] == "" || args[4] != "--model" || args[5] != validatorModel ||
		args[6] != "--thinking" || args[7] != "high" || args[8] != "--no-tools" ||
		args[9] != "--no-extensions" || args[10] != "--no-skills" || args[11] != "--no-prompt-templates" ||
		args[12] != "--no-themes" || args[13] != "--no-context-files" || args[14] != "--no-approve" ||
		args[15] != "--session-dir" || !filepath.IsAbs(args[16]) || args[17] != "--name" ||
		!strings.HasPrefix(args[18], "shepherd-recovery-") || args[19] != "--print" || args[20] == "" {
		return errors.New("fake recovery planner rejected noncanonical isolated argv")
	}
	return nil
}

func candidateCompletionMarker(workDir string) string {
	digest := sha256.Sum256([]byte(workDir))
	return filepath.Join(os.Getenv("GSD_STATE_DIR"), "candidate-complete-"+hex.EncodeToString(digest[:8]))
}

func contains(args []string, target string) bool {
	for _, value := range args {
		if value == target {
			return true
		}
	}
	return false
}

func value(args []string, flag string) string {
	for index := 0; index+1 < len(args); index++ {
		if args[index] == flag {
			return args[index+1]
		}
	}
	return ""
}

func processRole() string {
	name := filepath.Base(os.Args[0])
	switch {
	case name == "gh":
		return "github"
	case strings.HasPrefix(name, "fake-gsd-"):
		return "gsd"
	case strings.HasPrefix(name, "fake-pi-"):
		return "pi"
	default:
		return "unknown"
	}
}

func scenarioName() string {
	name := filepath.Base(os.Args[0])
	name = strings.TrimPrefix(name, "fake-gsd-")
	return strings.TrimPrefix(name, "fake-pi-")
}
