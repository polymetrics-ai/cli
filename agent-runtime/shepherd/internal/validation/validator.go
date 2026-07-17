package validation

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/authority"
	shepherdgit "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/git"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/gsd"
)

const maxValidatorOutputBytes = 2 * 1024 * 1024
const DeletionSentinelHash = shepherdgit.DeletionSentinelHash

type ArtifactHash struct {
	Path    string `json:"path"`
	Hash    string `json:"hash"`
	Deleted bool   `json:"deleted,omitempty"`
}

type Request struct {
	RequestID      string
	Repository     string
	PullRequest    int
	BaseBranch     string
	DeliveryID     string
	Generation     int64
	UnitID         string
	UnitType       string
	Attempt        int64
	StateVersion   int64
	WorkDir        string
	GSDHome        string
	StateDir       string
	BaseHead       string
	CandidateHead  string
	ContractHash   string
	EvidenceHash   string
	ArtifactHashes []ArtifactHash
	RequireGates   GateRequirements
}

type GateRequirements struct {
	LocalGates     bool `json:"local_gates"`
	UAT            bool `json:"uat"`
	MilestoneValid bool `json:"milestone_valid"`
}

type Result struct {
	ObservedModel  string
	Thinking       string
	SessionID      string
	Verdict        string
	ObservedHead   string
	EvidenceHash   string
	LocalGates     bool
	UAT            bool
	MilestoneValid bool
	IssuedAt       time.Time
	ExpiresAt      time.Time
}

type Validator interface {
	Validate(context.Context, Request) (Result, error)
}

type GSDValidator struct {
	Command     []string
	GSDHome     string
	StateDir    string
	SessionsDir string
	Timeout     time.Duration
	Now         func() time.Time
	Nonce       func() (string, error)
	Environment []string
}

type protectedRequest struct {
	RequestID      string           `json:"request_id"`
	Nonce          string           `json:"nonce"`
	Repository     string           `json:"repository"`
	PullRequest    int              `json:"pull_request"`
	BaseBranch     string           `json:"base_branch"`
	DeliveryID     string           `json:"delivery_id"`
	Generation     int64            `json:"generation"`
	UnitID         string           `json:"unit_id"`
	UnitType       string           `json:"unit_type"`
	Attempt        int64            `json:"attempt"`
	StateVersion   int64            `json:"state_version"`
	WorkDir        string           `json:"work_dir"`
	BaseHead       string           `json:"base_head"`
	CandidateHead  string           `json:"candidate_head"`
	ContractHash   string           `json:"contract_hash"`
	EvidenceHash   string           `json:"evidence_hash"`
	ArtifactHashes []ArtifactHash   `json:"artifact_hashes"`
	RequiredGates  GateRequirements `json:"required_gates"`
	IssuedAt       time.Time        `json:"issued_at"`
	ExpiresAt      time.Time        `json:"expires_at"`
}

type proofFile struct {
	StreamSessionID           string    `json:"stream_session_id,omitempty"`
	StreamProofHash           string    `json:"stream_proof_hash,omitempty"`
	StreamSuccessfulToolCount int       `json:"stream_successful_tool_count,omitempty"`
	RequestID                 string    `json:"request_id"`
	Nonce                     string    `json:"nonce"`
	Repository                string    `json:"repository"`
	PullRequest               int       `json:"pull_request"`
	BaseBranch                string    `json:"base_branch"`
	BaseHead                  string    `json:"base_head"`
	CandidateHead             string    `json:"candidate_head"`
	ObservedHead              string    `json:"observed_head"`
	StateVersion              int64     `json:"state_version"`
	ContractHash              string    `json:"contract_hash"`
	EvidenceHash              string    `json:"evidence_hash"`
	Verdict                   string    `json:"verdict"`
	LocalGates                bool      `json:"local_gates"`
	UAT                       bool      `json:"uat"`
	MilestoneValid            bool      `json:"milestone_valid"`
	IssuedAt                  time.Time `json:"issued_at"`
	ExpiresAt                 time.Time `json:"expires_at"`
}

func (v GSDValidator) Validate(ctx context.Context, request Request) (Result, error) {
	request = v.withDefaults(request)
	if err := validateRequest(request); err != nil {
		return Result{}, err
	}
	canonicalWorkDir, err := filepath.EvalSymlinks(filepath.Clean(request.WorkDir))
	if err != nil {
		return Result{}, fmt.Errorf("resolve candidate worktree: %w", err)
	}
	request.WorkDir = canonicalWorkDir
	if len(v.Command) == 0 {
		return Result{}, errors.New("validator command is required")
	}
	if inside, err := pathInside(request.WorkDir, request.StateDir); err != nil || inside {
		if err != nil {
			return Result{}, err
		}
		return Result{}, errors.New("validator state directory must be outside the candidate worktree")
	}
	if err := probePiValidator(ctx, v.Command, v.Timeout, v.Environment, request.WorkDir,
		request.StateDir, request.GSDHome); err != nil {
		return Result{}, err
	}
	observedHead, err := gitHead(ctx, request.WorkDir)
	if err != nil {
		return Result{}, err
	}
	if observedHead != request.CandidateHead {
		return Result{}, errors.New("candidate head moved before validation")
	}
	if err := verifyArtifactHashes(request.WorkDir, request.ArtifactHashes); err != nil {
		return Result{}, err
	}
	now := time.Now().UTC()
	if v.Now != nil {
		now = v.Now().UTC()
	}
	nonceFunc := v.Nonce
	if nonceFunc == nil {
		nonceFunc = newNonce
	}
	nonce, err := nonceFunc()
	if err != nil {
		return Result{}, err
	}
	root := filepath.Join(request.StateDir, "validation", safePathPart(request.RequestID), nonce)
	if err := os.MkdirAll(root, 0o700); err != nil {
		return Result{}, err
	}
	requestPath := filepath.Join(root, "request.json")
	resultPath := filepath.Join(root, "result.json")
	if _, err := os.Lstat(resultPath); err == nil {
		return Result{}, errors.New("stale validation result exists before validator start")
	} else if err != nil && !os.IsNotExist(err) {
		return Result{}, err
	}
	protected := protectedRequest{
		RequestID: request.RequestID, Nonce: nonce, Repository: request.Repository, PullRequest: request.PullRequest,
		BaseBranch: request.BaseBranch, DeliveryID: request.DeliveryID, Generation: request.Generation,
		UnitID: request.UnitID, UnitType: request.UnitType, Attempt: request.Attempt, StateVersion: request.StateVersion,
		WorkDir: request.WorkDir, BaseHead: request.BaseHead, CandidateHead: request.CandidateHead,
		ContractHash: request.ContractHash, EvidenceHash: request.EvidenceHash, ArtifactHashes: request.ArtifactHashes,
		RequiredGates: request.RequireGates, IssuedAt: now, ExpiresAt: now.Add(30 * time.Minute),
	}
	if err := writeExclusiveJSON(requestPath, protected); err != nil {
		return Result{}, err
	}
	sessionsDir := v.SessionsDir
	if sessionsDir == "" {
		sessionsDir = filepath.Join(request.StateDir, "validation", safePathPart(request.RequestID), nonce, "sessions")
	}
	if err := os.MkdirAll(filepath.Dir(sessionsDir), 0o700); err != nil {
		return Result{}, err
	}
	if err := os.Mkdir(sessionsDir, 0o700); err != nil {
		if errors.Is(err, os.ErrExist) {
			return Result{}, errors.New("validator session directory must be fresh for this invocation")
		}
		return Result{}, err
	}
	sessionBaseline, err := gsd.CaptureSessionIdentityBaseline(sessionsDir, request.WorkDir)
	if err != nil {
		return Result{}, fmt.Errorf("capture validator session baseline: %w", err)
	}
	started := time.Now().UTC()
	if err := v.runDedicatedValidator(ctx, request, protected, requestPath, resultPath, sessionsDir); err != nil {
		return Result{}, err
	}
	if observedHead, err = gitHead(ctx, request.WorkDir); err != nil {
		return Result{}, err
	}
	if observedHead != request.CandidateHead {
		return Result{}, errors.New("candidate head moved during validation")
	}
	if err := verifyArtifactHashes(request.WorkDir, request.ArtifactHashes); err != nil {
		return Result{}, fmt.Errorf("revalidate candidate artifacts after validation: %w", err)
	}
	proof, err := readValidationProof(resultPath)
	if err != nil {
		return Result{}, err
	}
	if err := validateProof(proof, protected); err != nil {
		return Result{}, err
	}
	sessionEvidence, err := gsd.ReadSessionIdentityEvidenceForRun(sessionsDir, request.WorkDir, sessionBaseline, started, authority.RequiredValidator, "high")
	if err != nil {
		return Result{}, fmt.Errorf("bind validator session evidence: %w", err)
	}
	if proof.StreamSessionID == "" || proof.StreamSessionID != sessionEvidence.SessionID {
		return Result{}, errors.New("validator event stream session does not match durable session evidence")
	}
	if err := gsd.ValidateSessionAssistantProof(sessionsDir, sessionEvidence.SessionID,
		proof.StreamProofHash); err != nil {
		return Result{}, fmt.Errorf("bind validator proof to durable session: %w", err)
	}
	return Result{
		ObservedModel: sessionEvidence.Model, Thinking: sessionEvidence.Thinking, SessionID: sessionEvidence.SessionID,
		Verdict: proof.Verdict, ObservedHead: proof.ObservedHead, EvidenceHash: proof.EvidenceHash,
		LocalGates: proof.LocalGates, UAT: proof.UAT, MilestoneValid: proof.MilestoneValid,
		IssuedAt: proof.IssuedAt, ExpiresAt: proof.ExpiresAt,
	}, nil
}

func (v GSDValidator) withDefaults(request Request) Request {
	if request.GSDHome == "" {
		request.GSDHome = v.GSDHome
	}
	if request.StateDir == "" {
		request.StateDir = v.StateDir
	}
	if request.RequestID == "" {
		request.RequestID = derivedRequestID(request)
	}
	return request
}

func probePiValidator(ctx context.Context, command []string, _ time.Duration, env []string,
	_ string, stateDir string, _ string,
) error {
	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		return err
	}
	probeHome, err := os.MkdirTemp(stateDir, "validator-probe-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(probeHome) }()
	probeProject := filepath.Join(probeHome, "project")
	probeAgent := filepath.Join(probeHome, "agent")
	probeGSDHome := filepath.Join(probeHome, "gsd-home")
	probeState := filepath.Join(probeHome, "state")
	for _, directory := range []string{probeProject, probeAgent, probeGSDHome, probeState} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return err
		}
	}
	args := append([]string{}, command[1:]...)
	args = append(args, "--no-tools", "--no-extensions", "--no-skills", "--no-prompt-templates",
		"--no-themes", "--no-context-files", "--no-approve", "--help")
	cmd := exec.CommandContext(probeCtx, command[0], args...)
	cmd.Dir = probeProject
	cmd.Env = sanitizedValidatorEnvironment(append(os.Environ(), env...))
	cmd.Env = append(cmd.Env, "HOME="+probeHome, "PI_CODING_AGENT_DIR="+probeAgent,
		"GSD_PROJECT_ROOT="+probeProject, "GSD_HOME="+probeGSDHome, "GSD_STATE_DIR="+probeState,
		"GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=")
	output, err := limitedCombinedOutput(cmd, 128*1024)
	if err != nil {
		return fmt.Errorf("validator Pi capability probe failed: %w", err)
	}
	help := string(output)
	for _, required := range []string{"--mode", "--print", "--session-dir", "--tools", "--model", "--thinking"} {
		if !strings.Contains(help, required) {
			return fmt.Errorf("validator Pi command does not advertise %s", required)
		}
	}
	return nil
}

func (v GSDValidator) runDedicatedValidator(ctx context.Context, request Request, protected protectedRequest, requestPath, resultPath, sessionsDir string) error {
	if v.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v.Timeout)
		defer cancel()
	}
	prompt, err := validationPrompt(protected, requestPath)
	if err != nil {
		return err
	}
	args := append([]string{}, v.Command[1:]...)
	args = append(args,
		"--mode", "json",
		"--system-prompt", "You are a deterministic read-only validation process. Follow the supplied validation contract and finish with exactly one compact JSON object, with no markdown or commentary.",
		"--model", authority.RequiredValidator,
		"--thinking", "high",
		"--tools", "read,grep,find,ls",
		"--no-extensions", "--no-skills", "--no-prompt-templates", "--no-themes", "--no-context-files",
		"--no-approve",
		"--session-dir", sessionsDir,
		"--name", "shepherd-validator-"+protected.RequestID,
		"--print", prompt,
	)
	cmd := exec.CommandContext(ctx, v.Command[0], args...)
	cmd.Dir = request.WorkDir
	validatorHome := filepath.Join(sessionsDir, "home")
	if err := os.MkdirAll(validatorHome, 0o700); err != nil {
		return fmt.Errorf("create isolated validator home: %w", err)
	}
	cmd.Env = sanitizedValidatorEnvironment(append(os.Environ(), v.Environment...))
	cmd.Env = append(cmd.Env,
		"HOME="+validatorHome,
		"PI_CODING_AGENT_DIR="+filepath.Join(request.GSDHome, "agent"),
		"GSD_PROJECT_ROOT="+request.WorkDir,
		"GSD_HOME="+request.GSDHome,
		"GSD_STATE_DIR="+request.StateDir,
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=",
	)
	output, err := limitedCombinedOutput(cmd, maxValidatorOutputBytes)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return fmt.Errorf("dedicated validator failed: %w", err)
	}
	proof, err := proofFromPiJSON(output)
	if err != nil {
		return err
	}
	return writeExclusiveJSON(resultPath, proof)
}

func validationPrompt(request protectedRequest, requestPath string) (string, error) {
	encoded, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	return "You are the read-only Shepherd independent validator. Use only read, grep, find, and ls; no shell or network-capable tool is available. " +
		"Do not edit files, write files, mutate GitHub, promote, merge, or reveal secrets. " +
		"Copy every identity, nonce, head, hash, repository, PR, base, and state-version field exactly from the request; do not omit any field. " +
		"Do not run broad tests or unlisted checks. If all required_gates are false, call no tools and return PROCEED after copying the binding fields; do not invent extra requirements. Otherwise use at most five bounded read-only tool calls. " +
		"Validate the exact candidate worktree and return exactly one compact JSON object with keys " +
		"request_id, nonce, repository, pull_request, base_branch, base_head, candidate_head, observed_head, state_version, contract_hash, evidence_hash, verdict, local_gates, uat, milestone_valid, issued_at, expires_at. " +
		"Use PROCEED only if evidence and required gates pass; otherwise RETRY or HALT. " +
		"Request file: " + requestPath + ". Request JSON: " + string(encoded), nil
}

func proofFromPiJSON(output []byte) (proofFile, error) {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 4096), maxValidatorOutputBytes)
	var proofTexts []string
	started, ended, settled, sessionSeen := false, false, false, false
	turnActive, messageActive := false, false
	activeMessageRole, streamSessionID := "", ""
	turns, messagesInTurn := 0, 0
	proofObserved := false
	activeTools := make(map[string]string)
	toolCalls, successfulToolCalls := 0, 0
	lines := 0
	for scanner.Scan() {
		lines++
		if settled {
			return proofFile{}, errors.New("validator event stream continued after agent_settled")
		}
		var event struct {
			Type       string `json:"type"`
			ID         string `json:"id"`
			ToolCallID string `json:"toolCallId"`
			ToolName   string `json:"toolName"`
			WillRetry  bool   `json:"willRetry"`
			IsError    *bool  `json:"isError"`
			Status     string `json:"status"`
			Message    struct {
				Role       string `json:"role"`
				Content    any    `json:"content"`
				StopReason string `json:"stopReason"`
			} `json:"message"`
		}
		if duplicateErr := gsd.RejectDuplicateJSONFields(scanner.Bytes()); duplicateErr != nil {
			return proofFile{}, errors.New("validator event stream contains duplicate JSON fields")
		}
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil || event.Type == "" {
			return proofFile{}, errors.New("validator event stream contains malformed JSON")
		}
		if ended && event.Type != "agent_settled" {
			return proofFile{}, errors.New("validator event stream continued after agent_end")
		}
		switch event.Type {
		case "session":
			if lines != 1 || sessionSeen || started || event.ID == "" {
				return proofFile{}, errors.New("validator session event is incomplete or out of order")
			}
			sessionSeen = true
			streamSessionID = event.ID
		case "agent_start":
			if !sessionSeen || started {
				return proofFile{}, errors.New("validator agent_start is out of order")
			}
			started = true
		case "agent_end":
			if !started || ended || event.WillRetry || (event.Status != "" && event.Status != "success") ||
				turnActive || messageActive || turns == 0 || len(activeTools) != 0 {
				return proofFile{}, errors.New("validator agent_end has active or missing lifecycle state")
			}
			ended = true
		case "agent_settled":
			if !ended {
				return proofFile{}, errors.New("validator agent_settled preceded agent_end")
			}
			settled = true
		case "turn_start":
			if !started || ended || proofObserved || turnActive || messageActive || len(activeTools) != 0 {
				return proofFile{}, errors.New("validator turn_start is out of order")
			}
			turnActive = true
			turns++
			messagesInTurn = 0
		case "message_start":
			if !turnActive || messageActive || proofObserved || event.Message.Role == "" {
				return proofFile{}, errors.New("validator message_start is out of order")
			}
			messageActive = true
			activeMessageRole = event.Message.Role
		case "message_update":
			if !turnActive || !messageActive || event.Message.Role != activeMessageRole {
				return proofFile{}, errors.New("validator message_update is out of order")
			}
		case "message_end":
			if !turnActive || !messageActive || event.Message.Role != activeMessageRole {
				return proofFile{}, errors.New("validator message_end is out of order")
			}
			if event.Message.Role == "assistant" {
				var text strings.Builder
				collectText(event.Message.Content, &text)
				if value := strings.TrimSpace(text.String()); value != "" {
					if event.Message.StopReason != "stop" || proofObserved || len(activeTools) != 0 {
						return proofFile{}, errors.New("validator proof message is not a successful terminal response")
					}
					proofTexts = append(proofTexts, value)
					proofObserved = true
				}
			}
			messagesInTurn++
			messageActive = false
			activeMessageRole = ""
		case "tool_execution_start":
			if !turnActive || messageActive || messagesInTurn == 0 || proofObserved || event.ToolCallID == "" || !allowedValidatorTool(event.ToolName) ||
				activeTools[event.ToolCallID] != "" || toolCalls >= 5 {
				return proofFile{}, errors.New("validator used an unbound or forbidden tool")
			}
			activeTools[event.ToolCallID] = event.ToolName
			toolCalls++
		case "tool_execution_update":
			if !turnActive || messageActive {
				return proofFile{}, errors.New("validator tool update is outside its turn")
			}
			toolName, active := activeTools[event.ToolCallID]
			if event.ToolCallID == "" || event.ToolName == "" || !active || toolName != event.ToolName {
				return proofFile{}, errors.New("validator tool update is unbound")
			}
		case "tool_execution_end":
			if !turnActive || messageActive {
				return proofFile{}, errors.New("validator tool completion is outside its turn")
			}
			toolName, active := activeTools[event.ToolCallID]
			if event.ToolCallID == "" || event.ToolName == "" || !active || toolName != event.ToolName {
				return proofFile{}, errors.New("validator tool completion is unbound")
			}
			if event.IsError == nil || *event.IsError || (event.Status != "" && event.Status != "success") {
				return proofFile{}, errors.New("validator tool completion did not report explicit success")
			}
			delete(activeTools, event.ToolCallID)
			successfulToolCalls++
		case "turn_end":
			if !turnActive || messageActive || messagesInTurn == 0 || len(activeTools) != 0 {
				return proofFile{}, errors.New("validator turn_end is out of order")
			}
			turnActive = false
		default:
			return proofFile{}, fmt.Errorf("validator emitted unsupported event %q", event.Type)
		}
	}
	if err := scanner.Err(); err != nil {
		return proofFile{}, fmt.Errorf("read validator event stream: %w", err)
	}
	if !sessionSeen || !started || !ended || !settled || len(proofTexts) != 1 {
		return proofFile{}, fmt.Errorf("validator structured evidence missing: proof_messages=%d", len(proofTexts))
	}
	proof, err := decodeProofFromText(proofTexts[0])
	if err != nil {
		return proofFile{}, fmt.Errorf("validator structured evidence missing: proof_messages=%d", len(proofTexts))
	}
	proof.StreamSessionID = streamSessionID
	proof.StreamSuccessfulToolCount = successfulToolCalls
	proofDigest := sha256.Sum256([]byte(proofTexts[0]))
	proof.StreamProofHash = "sha256:" + hex.EncodeToString(proofDigest[:])
	return proof, nil
}

func allowedValidatorTool(name string) bool {
	return name == "read" || name == "grep" || name == "find" || name == "ls"
}

func collectText(value any, builder *strings.Builder) {
	switch typed := value.(type) {
	case map[string]any:
		if text, ok := typed["text"].(string); ok {
			builder.WriteString(text)
			builder.WriteByte('\n')
		}
		if delta, ok := typed["delta"].(string); ok {
			builder.WriteString(delta)
		}
		for _, nested := range typed {
			collectText(nested, builder)
		}
	case []any:
		for _, nested := range typed {
			collectText(nested, builder)
		}
	}
}

func decodeProofFromText(value string) (proofFile, error) {
	raw := []byte(strings.TrimSpace(value))
	fields, err := strictProofFields(raw)
	if err != nil {
		return proofFile{}, errors.New("validator did not return one strict JSON proof")
	}
	allowed := map[string]struct{}{
		"request_id": {}, "nonce": {}, "repository": {}, "pull_request": {}, "base_branch": {},
		"base_head": {}, "candidate_head": {}, "observed_head": {}, "state_version": {},
		"contract_hash": {}, "evidence_hash": {}, "verdict": {}, "local_gates": {}, "uat": {},
		"milestone_valid": {}, "issued_at": {}, "expires_at": {},
	}
	if len(fields) != len(allowed) {
		return proofFile{}, errors.New("validator proof has missing or extra fields")
	}
	for name := range fields {
		if _, ok := allowed[name]; !ok {
			return proofFile{}, errors.New("validator proof has an unknown field")
		}
	}
	var proof proofFile
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&proof); err != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		proof.RequestID == "" || proof.Nonce == "" {
		return proofFile{}, errors.New("validator did not return structured JSON evidence")
	}
	return proof, nil
}

func strictProofFields(raw []byte) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return nil, errors.New("validator proof must be an object")
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyToken.(string)
		if !ok {
			return nil, errors.New("validator proof key must be a string")
		}
		if _, duplicate := fields[key]; duplicate {
			return nil, errors.New("validator proof has duplicate fields")
		}
		var rawValue json.RawMessage
		if err := decoder.Decode(&rawValue); err != nil {
			return nil, err
		}
		fields[key] = rawValue
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') || decoder.Decode(&struct{}{}) != io.EOF {
		return nil, errors.New("validator proof has trailing data")
	}
	return fields, nil
}

func validateProof(proof proofFile, request protectedRequest) error {
	if proof.RequestID != request.RequestID || proof.Nonce != request.Nonce {
		return errors.New("validation result does not match request nonce")
	}
	if proof.Repository != request.Repository || proof.PullRequest != request.PullRequest {
		return errors.New("validation result repository or PR does not match request")
	}
	if proof.BaseBranch != request.BaseBranch || proof.BaseHead != request.BaseHead || proof.CandidateHead != request.CandidateHead || proof.ObservedHead != request.CandidateHead {
		return errors.New("validation result head or base does not match request")
	}
	if proof.StateVersion != request.StateVersion || proof.ContractHash != request.ContractHash || proof.EvidenceHash != request.EvidenceHash {
		return errors.New("validation result governance or evidence does not match request")
	}
	if proof.Verdict != "PROCEED" && proof.Verdict != "RETRY" && proof.Verdict != "HALT" {
		return errors.New("validator returned an unsupported verdict")
	}
	if proof.Verdict == "PROCEED" &&
		(request.RequiredGates.LocalGates || request.RequiredGates.UAT || request.RequiredGates.MilestoneValid) &&
		proof.StreamSuccessfulToolCount == 0 {
		return errors.New("validator PROCEED lacks successful evidence-gathering tool use")
	}
	if proof.Verdict == "PROCEED" && request.RequiredGates.LocalGates && !proof.LocalGates {
		return errors.New("validator did not pass required local gates")
	}
	if proof.Verdict == "PROCEED" && request.RequiredGates.UAT && !proof.UAT {
		return errors.New("validator did not pass required UAT gate")
	}
	if proof.Verdict == "PROCEED" && request.RequiredGates.MilestoneValid && !proof.MilestoneValid {
		return errors.New("validator did not pass required milestone gate")
	}
	if proof.IssuedAt.IsZero() || proof.ExpiresAt.IsZero() || !proof.ExpiresAt.After(proof.IssuedAt) || !proof.IssuedAt.Equal(request.IssuedAt) || !proof.ExpiresAt.Equal(request.ExpiresAt) {
		return errors.New("validation result validity window does not match request")
	}
	return nil
}

func readValidationProof(path string) (proofFile, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return proofFile{}, fmt.Errorf("read independent validation proof: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Size() > 64*1024 {
		return proofFile{}, errors.New("independent validation proof must be a bounded regular file")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return proofFile{}, err
	}
	var proof proofFile
	if err := json.Unmarshal(raw, &proof); err != nil {
		return proofFile{}, fmt.Errorf("decode independent validation proof: %w", err)
	}
	if proof.StreamSessionID == "" {
		return proofFile{}, errors.New("independent validation proof is missing stream session identity")
	}
	if !validHash(proof.StreamProofHash) {
		return proofFile{}, errors.New("independent validation proof is missing stream proof hash")
	}
	if proof.Verdict == "" {
		return proofFile{}, errors.New("independent validation proof is missing verdict")
	}
	if !validHash(proof.EvidenceHash) {
		return proofFile{}, errors.New("independent validation proof has invalid evidence hash")
	}
	if proof.IssuedAt.IsZero() || proof.ExpiresAt.IsZero() {
		return proofFile{}, errors.New("independent validation proof is missing validity timestamps")
	}
	return proof, nil
}

func validateRequest(request Request) error {
	if request.RequestID == "" || request.Repository == "" || request.PullRequest <= 0 || request.BaseBranch == "" || request.DeliveryID == "" || request.UnitID == "" || request.UnitType == "" || request.WorkDir == "" || request.GSDHome == "" || request.StateDir == "" {
		return errors.New("complete validation request identity is required")
	}
	if request.Generation <= 0 || request.Attempt <= 0 || request.StateVersion <= 0 || !validSHA(request.BaseHead) || !validSHA(request.CandidateHead) || !validHash(request.ContractHash) || !validHash(request.EvidenceHash) {
		return errors.New("complete validation heads and hashes are required")
	}
	if len(request.ArtifactHashes) == 0 {
		return errors.New("validation evidence requires artifact hashes")
	}
	for _, artifact := range request.ArtifactHashes {
		if strings.TrimSpace(artifact.Path) == "" || !validHash(artifact.Hash) {
			return errors.New("validation artifact hashes must be complete")
		}
		if artifact.Deleted != (artifact.Hash == DeletionSentinelHash) {
			return errors.New("validation deletion artifact identity is inconsistent")
		}
	}
	return nil
}

func verifyArtifactHashes(workDir string, artifacts []ArtifactHash) error {
	if len(artifacts) == 0 || len(artifacts) > 128 {
		return errors.New("validation artifact set is empty or exceeds the governed limit")
	}
	for _, artifact := range artifacts {
		parent, leaf, parentMissing, err := openStableArtifactParent(workDir, artifact.Path)
		if err != nil {
			return err
		}
		if parent != nil {
			defer func(root *os.Root) { _ = root.Close() }(parent)
		}
		if artifact.Deleted {
			if artifact.Hash != DeletionSentinelHash {
				return errors.New("validation deleted artifact must use the deletion sentinel")
			}
			if parentMissing {
				continue
			}
			if _, err := parent.Lstat(leaf); errors.Is(err, os.ErrNotExist) {
				continue
			} else if err != nil {
				return err
			}
			return errors.New("validation deleted artifact exists in the candidate worktree")
		}
		if artifact.Hash == DeletionSentinelHash {
			return errors.New("validation present artifact must not use the deletion sentinel")
		}
		if parentMissing {
			return errors.New("validation present artifact is missing")
		}
		file, info, err := openStableArtifactFile(parent, leaf)
		if err != nil {
			return err
		}
		digest := sha256.New()
		written, copyErr := io.Copy(digest, io.LimitReader(file, 8*1024*1024+1))
		closeErr := file.Close()
		if copyErr != nil || closeErr != nil {
			return errors.Join(copyErr, closeErr)
		}
		if !info.Mode().IsRegular() {
			return errors.New("validation present artifact must be a regular non-symlink file")
		}
		if written > 8*1024*1024 {
			return errors.New("validation artifact exceeds the governed size limit")
		}
		observed := "sha256:" + hex.EncodeToString(digest.Sum(nil))
		if observed != artifact.Hash {
			return errors.New("validation artifact hash does not match protected evidence")
		}
	}
	return nil
}

func openStableArtifactParent(workDir, relPath string) (*os.Root, string, bool, error) {
	parts, err := validateArtifactRelativePath(relPath)
	if err != nil {
		return nil, "", false, err
	}
	rootPath := filepath.Clean(workDir)
	before, err := os.Lstat(rootPath)
	if err != nil {
		return nil, "", false, err
	}
	if before.Mode()&os.ModeSymlink != 0 || !before.IsDir() {
		return nil, "", false, errors.New("validation worktree root must be a real directory")
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, "", false, err
	}
	rootInfo, statErr := root.Stat(".")
	after, lstatErr := os.Lstat(rootPath)
	if statErr != nil || lstatErr != nil || after.Mode()&os.ModeSymlink != 0 || !os.SameFile(before, rootInfo) || !os.SameFile(before, after) {
		_ = root.Close()
		return nil, "", false, errors.New("validation worktree root changed during open")
	}
	for _, part := range parts[:len(parts)-1] {
		next, missing, err := openStableChildRoot(root, part)
		if err != nil {
			_ = root.Close()
			return nil, "", false, err
		}
		if missing {
			_ = root.Close()
			return nil, parts[len(parts)-1], true, nil
		}
		_ = root.Close()
		root = next
	}
	return root, parts[len(parts)-1], false, nil
}

func validateArtifactRelativePath(relPath string) ([]string, error) {
	if filepath.IsAbs(relPath) || filepath.Clean(relPath) != relPath || relPath == "." || relPath == ".." || strings.HasPrefix(relPath, "../") || strings.ContainsRune(relPath, 0) {
		return nil, errors.New("validation artifact path is not a clean relative path")
	}
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return nil, errors.New("validation artifact path is not a clean relative path")
		}
	}
	return parts, nil
}

func openStableChildRoot(parent *os.Root, name string) (*os.Root, bool, error) {
	before, err := parent.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		return nil, true, nil
	}
	if err != nil {
		return nil, false, err
	}
	if before.Mode()&os.ModeSymlink != 0 || !before.IsDir() {
		return nil, false, errors.New("validation artifact parent is not a stable directory")
	}
	child, err := parent.OpenRoot(name)
	if err != nil {
		return nil, false, err
	}
	childInfo, statErr := child.Stat(".")
	after, lstatErr := parent.Lstat(name)
	if statErr != nil || lstatErr != nil || after.Mode()&os.ModeSymlink != 0 || !os.SameFile(before, childInfo) || !os.SameFile(before, after) {
		_ = child.Close()
		return nil, false, errors.New("validation artifact parent changed during open")
	}
	return child, false, nil
}

func openStableArtifactFile(parent *os.Root, leaf string) (*os.File, os.FileInfo, error) {
	before, err := parent.Lstat(leaf)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil, errors.New("validation present artifact is missing")
	}
	if err != nil {
		return nil, nil, err
	}
	if before.Mode()&os.ModeSymlink != 0 || !before.Mode().IsRegular() {
		return nil, nil, errors.New("validation present artifact must be a regular non-symlink file")
	}
	file, err := parent.Open(leaf)
	if err != nil {
		return nil, nil, fmt.Errorf("open validation artifact: %w", err)
	}
	opened, statErr := file.Stat()
	after, lstatErr := parent.Lstat(leaf)
	if statErr != nil || lstatErr != nil || after.Mode()&os.ModeSymlink != 0 || !after.Mode().IsRegular() || !os.SameFile(before, opened) || !os.SameFile(before, after) {
		_ = file.Close()
		return nil, nil, errors.New("validation artifact changed during open")
	}
	return file, opened, nil
}

func gitHead(ctx context.Context, workDir string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", workDir, "rev-parse", "HEAD")
	cmd.Env = sanitizedValidatorEnvironment(os.Environ())
	raw, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("inspect candidate head: %w", err)
	}
	return strings.TrimSpace(string(raw)), nil
}

func sanitizedValidatorEnvironment(source []string) []string {
	allowed := map[string]struct{}{
		"PATH": {}, "TMPDIR": {}, "TMP": {}, "TEMP": {}, "LANG": {}, "LC_ALL": {},
		"LC_CTYPE": {}, "TZ": {}, "SSL_CERT_FILE": {}, "SSL_CERT_DIR": {},
	}
	environment := make([]string, 0, len(allowed)+4)
	for _, entry := range source {
		name, _, found := strings.Cut(entry, "=")
		upper := strings.ToUpper(name)
		_, safeRuntimeVariable := allowed[upper]
		if !found || !safeRuntimeVariable && !strings.HasPrefix(upper, "GO_WANT_VALIDATOR_HELPER") {
			continue
		}
		environment = append(environment, entry)
	}
	return append(environment, "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=")
}

func writeExclusiveJSON(path string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(raw)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func newNonce() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

func derivedRequestID(request Request) string {
	input := fmt.Sprintf("%s:%d:%s:%d:%d:%s:%s", request.DeliveryID, request.Generation, request.UnitID, request.Attempt, request.StateVersion, request.CandidateHead, request.EvidenceHash)
	sum := sha256.Sum256([]byte(input))
	return "validation-" + hex.EncodeToString(sum[:8])
}

func pathInside(root, child string) (bool, error) {
	rootEval, err := filepath.EvalSymlinks(filepath.Clean(root))
	if err != nil {
		return false, err
	}
	childClean := filepath.Clean(child)
	if _, err := os.Stat(childClean); err == nil {
		childClean, err = filepath.EvalSymlinks(childClean)
		if err != nil {
			return false, err
		}
	}
	rel, err := filepath.Rel(rootEval, childClean)
	if err != nil {
		return false, err
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))), nil
}

func safePathPart(value string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			return r
		}
		return '-'
	}, value)
}

func limitedCombinedOutput(cmd *exec.Cmd, maxBytes int64) ([]byte, error) {
	configureValidationProcessTree(cmd)
	var stdout, stderr bytes.Buffer
	stdoutWriter := &limitedWriter{buffer: &stdout, max: maxBytes}
	stderrWriter := &limitedWriter{buffer: &stderr, max: maxBytes}
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter
	err := cmd.Run()
	cleanupErr := cleanupValidationProcessTree(cmd)
	if stdoutWriter.exceeded || stderrWriter.exceeded {
		return nil, errors.Join(errors.New("validator process output exceeded its bound"), err, cleanupErr)
	}
	return stdout.Bytes(), errors.Join(err, cleanupErr)
}

type limitedWriter struct {
	buffer   *bytes.Buffer
	max      int64
	exceeded bool
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if w.max <= 0 {
		w.exceeded = true
		return len(p), nil
	}
	remaining := w.max - int64(w.buffer.Len())
	if remaining <= 0 {
		w.exceeded = true
		return len(p), nil
	}
	if int64(len(p)) > remaining {
		w.exceeded = true
		_, _ = w.buffer.Write(p[:remaining])
		return len(p), nil
	}
	_, _ = w.buffer.Write(p)
	return len(p), nil
}

func validSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

func validHash(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	for _, char := range strings.TrimPrefix(value, "sha256:") {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}
