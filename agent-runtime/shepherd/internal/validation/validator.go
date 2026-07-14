package validation

import (
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
}

type proofFile struct {
	RequestID      string    `json:"request_id"`
	Nonce          string    `json:"nonce"`
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
	if len(v.Command) == 0 {
		return Result{}, errors.New("validator command is required")
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
	nonce, err := newNonce()
	if err != nil {
		return Result{}, err
	}
	root := filepath.Join(request.StateDir, "validation", safePathPart(request.RequestID))
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
		RequiredGates: request.RequireGates, IssuedAt: now,
	}
	if err := writeExclusiveJSON(requestPath, protected); err != nil {
		return Result{}, err
	}
	sessionsDir := v.SessionsDir
	if sessionsDir == "" {
		sessionsDir = filepath.Join(request.GSDHome, "agent", "sessions")
	}
	previousSession := latestSessionID(sessionsDir, request.WorkDir)
	started := time.Now().UTC()
	if err := v.runDedicatedValidator(ctx, request, requestPath, resultPath); err != nil {
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

func (v GSDValidator) runDedicatedValidator(ctx context.Context, request Request, requestPath, resultPath string) error {
	if v.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v.Timeout)
		defer cancel()
	}
	args := append([]string{}, v.Command[1:]...)
	args = append(args,
		"headless", "shepherd-validate",
		"--request", requestPath,
		"--result", resultPath,
		"--worktree", request.WorkDir,
		"--model", authority.RequiredValidator,
		"--thinking", "high",
	)
	cmd := exec.CommandContext(ctx, v.Command[0], args...)
	cmd.Dir = request.WorkDir
	cmd.Env = append(os.Environ(), v.Environment...)
	cmd.Env = append(cmd.Env,
		"GSD_PROJECT_ROOT="+request.WorkDir,
		"GSD_HOME="+request.GSDHome,
		"GSD_STATE_DIR="+request.StateDir,
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("dedicated validator failed: %w: %s", err, boundedOutput(output))
	}
	return nil
}

func validateProof(proof proofFile, request protectedRequest) error {
	if proof.RequestID != request.RequestID || proof.Nonce != request.Nonce {
		return errors.New("validation result does not match request nonce")
	}
	if proof.BaseBranch != request.BaseBranch || proof.BaseHead != request.BaseHead || proof.CandidateHead != request.CandidateHead || proof.ObservedHead != request.CandidateHead {
		return errors.New("validation result head or base does not match request")
	}
	if proof.StateVersion != request.StateVersion || proof.ContractHash != request.ContractHash || proof.EvidenceHash != request.EvidenceHash {
		return errors.New("validation result governance or evidence does not match request")
	}
	if proof.Verdict != "PROCEED" {
		return errors.New("validator did not return PROCEED")
	}
	if request.RequiredGates.LocalGates && !proof.LocalGates {
		return errors.New("validator did not pass required local gates")
	}
	if request.RequiredGates.UAT && !proof.UAT {
		return errors.New("validator did not pass required UAT gate")
	}
	if request.RequiredGates.MilestoneValid && !proof.MilestoneValid {
		return errors.New("validator did not pass required milestone gate")
	}
	if proof.IssuedAt.IsZero() || proof.ExpiresAt.IsZero() || !proof.ExpiresAt.After(proof.IssuedAt) {
		return errors.New("validation result validity window is invalid")
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
	if proof.Verdict == "" || !validHash(proof.EvidenceHash) || proof.IssuedAt.IsZero() || proof.ExpiresAt.IsZero() {
		return proofFile{}, errors.New("independent validation proof is incomplete")
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
	raw, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("inspect candidate head: %w", err)
	}
	return strings.TrimSpace(string(raw)), nil
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
	input := request.DeliveryID + ":" + request.UnitID + ":" + request.CandidateHead + ":" + request.EvidenceHash
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

func boundedOutput(raw []byte) string {
	value := strings.TrimSpace(string(raw))
	if len(value) > 2048 {
		return value[:2048]
	}
	return value
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
