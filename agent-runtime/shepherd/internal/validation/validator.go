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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/authority"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/gsd"
)

const maxValidatorOutputBytes = 2 * 1024 * 1024

type ArtifactHash struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
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
	RequestID      string    `json:"request_id"`
	Nonce          string    `json:"nonce"`
	Repository     string    `json:"repository"`
	PullRequest    int       `json:"pull_request"`
	BaseBranch     string    `json:"base_branch"`
	BaseHead       string    `json:"base_head"`
	CandidateHead  string    `json:"candidate_head"`
	ObservedHead   string    `json:"observed_head"`
	StateVersion   int64     `json:"state_version"`
	ContractHash   string    `json:"contract_hash"`
	EvidenceHash   string    `json:"evidence_hash"`
	Verdict        string    `json:"verdict"`
	LocalGates     bool      `json:"local_gates"`
	UAT            bool      `json:"uat"`
	MilestoneValid bool      `json:"milestone_valid"`
	IssuedAt       time.Time `json:"issued_at"`
	ExpiresAt      time.Time `json:"expires_at"`
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
	if err := probePiValidator(ctx, v.Command, v.Timeout, v.Environment); err != nil {
		return Result{}, err
	}
	if inside, err := pathInside(request.WorkDir, request.StateDir); err != nil || inside {
		if err != nil {
			return Result{}, err
		}
		return Result{}, errors.New("validator state directory must be outside the candidate worktree")
	}
	observedHead, err := gitHead(ctx, request.WorkDir)
	if err != nil {
		return Result{}, err
	}
	if observedHead != request.CandidateHead {
		return Result{}, errors.New("candidate head moved before validation")
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
	if err := os.MkdirAll(sessionsDir, 0o700); err != nil {
		return Result{}, err
	}
	previousSession := latestSessionID(sessionsDir, request.WorkDir)
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
	proof, err := readValidationProof(resultPath)
	if err != nil {
		return Result{}, err
	}
	if err := validateProof(proof, protected); err != nil {
		return Result{}, err
	}
	sessionID, model, thinking, err := readNewSession(sessionsDir, request.WorkDir, previousSession, started)
	if err != nil {
		return Result{}, err
	}
	if model != authority.RequiredValidator {
		return Result{}, fmt.Errorf("validator model was %s", model)
	}
	if thinking != "high" {
		return Result{}, fmt.Errorf("validator thinking was %s", thinking)
	}
	return Result{
		ObservedModel: model, Thinking: thinking, SessionID: sessionID,
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

func probePiValidator(ctx context.Context, command []string, _ time.Duration, env []string) error {
	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	args := append([]string{}, command[1:]...)
	args = append(args, "--help")
	cmd := exec.CommandContext(probeCtx, command[0], args...)
	cmd.Env = append(os.Environ(), env...)
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
		"--tools", "read,bash,grep,find,ls",
		"--no-extensions", "--no-skills", "--no-prompt-templates", "--no-themes", "--no-context-files",
		"--no-approve",
		"--session-dir", sessionsDir,
		"--name", "shepherd-validator-"+protected.RequestID,
		"--print", prompt,
	)
	cmd := exec.CommandContext(ctx, v.Command[0], args...)
	cmd.Dir = request.WorkDir
	cmd.Env = sanitizedValidatorEnvironment(append(os.Environ(), v.Environment...))
	cmd.Env = append(cmd.Env,
		"PI_CODING_AGENT_DIR="+filepath.Join(request.GSDHome, "agent"),
		"GSD_PROJECT_ROOT="+request.WorkDir,
		"GSD_HOME="+request.GSDHome,
		"GSD_STATE_DIR="+request.StateDir,
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=",
	)
	output, err := limitedCombinedOutput(cmd, maxValidatorOutputBytes)
	if err != nil {
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
	return "You are the read-only Shepherd independent validator. Use only read, bash, grep, find, and ls. " +
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
	var text strings.Builder
	assistantMessages := 0
	for scanner.Scan() {
		var event struct {
			Type    string `json:"type"`
			Message struct {
				Role    string `json:"role"`
				Content any    `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(scanner.Bytes(), &event) != nil || event.Type != "message_end" || event.Message.Role != "assistant" {
			continue
		}
		assistantMessages++
		collectText(event.Message.Content, &text)
	}
	if err := scanner.Err(); err != nil {
		return proofFile{}, fmt.Errorf("read validator event stream: %w", err)
	}
	proof, err := decodeProofFromText(text.String())
	if err != nil {
		return proofFile{}, fmt.Errorf("validator structured evidence missing: assistant_messages=%d text_bytes=%d", assistantMessages, text.Len())
	}
	return proof, nil
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
	for start := 0; start < len(value); start++ {
		if value[start] != '{' {
			continue
		}
		depth := 0
		quoted := false
		escaped := false
		for end := start; end < len(value); end++ {
			char := value[end]
			if quoted {
				if escaped {
					escaped = false
					continue
				}
				switch char {
				case '\\':
					escaped = true
				case '"':
					quoted = false
				}
				continue
			}
			switch char {
			case '"':
				quoted = true
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					var proof proofFile
					if json.Unmarshal([]byte(value[start:end+1]), &proof) == nil && proof.RequestID != "" && proof.Nonce != "" {
						return proof, nil
					}
					end = len(value)
				}
			}
		}
	}
	return proofFile{}, errors.New("validator did not return structured JSON evidence")
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
	}
	return nil
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
	environment := make([]string, 0, len(source)+2)
	for _, entry := range source {
		name, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(strings.ToUpper(name), "GIT_") {
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

func readNewSession(root, workDir, previous string, started time.Time) (string, string, string, error) {
	id, err := gsd.LatestSessionID(root, workDir)
	if err != nil {
		return "", "", "", err
	}
	if id == previous {
		return "", "", "", errors.New("validator did not create a new session")
	}
	model, thinking, err := gsd.ReadSessionIdentity(root, workDir)
	if err != nil {
		return "", "", "", err
	}
	path, err := latestSessionPath(root, workDir, id)
	if err != nil {
		return "", "", "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", "", "", err
	}
	if info.ModTime().Before(started) {
		return "", "", "", errors.New("validator session predates invocation")
	}
	return id, model, thinking, nil
}

func latestSessionID(root, workDir string) string {
	id, err := gsd.LatestSessionID(root, workDir)
	if err != nil {
		return ""
	}
	return id
}

func latestSessionPath(root, workDir, id string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") || found != "" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		first, _, _ := strings.Cut(string(raw), "\n")
		if strings.Contains(first, `"id":"`+id+`"`) && strings.Contains(first, `"cwd":"`+workDir+`"`) {
			found = path
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", errors.New("validator session path not found")
	}
	return found, nil
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
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &limitedWriter{buffer: &stdout, max: maxBytes}
	cmd.Stderr = &limitedWriter{buffer: &stderr, max: maxBytes}
	err := cmd.Run()
	return append(stdout.Bytes(), stderr.Bytes()...), err
}

type limitedWriter struct {
	buffer *bytes.Buffer
	max    int64
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if w.max <= 0 {
		return len(p), nil
	}
	if int64(len(p)) >= w.max {
		w.buffer.Reset()
		_, _ = w.buffer.Write(p[int64(len(p))-w.max:])
		return len(p), nil
	}
	overflow := int64(w.buffer.Len()+len(p)) - w.max
	if overflow > 0 {
		retained := append([]byte(nil), w.buffer.Bytes()[overflow:]...)
		w.buffer.Reset()
		_, _ = w.buffer.Write(retained)
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
