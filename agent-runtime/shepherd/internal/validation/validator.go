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
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/authority"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/gsd"
)

type ArtifactHash struct {
	Path string
	Hash string
}

type Request struct {
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
	BaseHead       string
	CandidateHead  string
	ContractHash   string
	EvidenceHash   string
	ArtifactHashes []ArtifactHash
	RequireGates   GateRequirements
}

type GateRequirements struct {
	LocalGates     bool
	UAT            bool
	MilestoneValid bool
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
	Runner      *gsd.Runner
	SessionsDir string
	Now         func() time.Time
}

func (v GSDValidator) Validate(ctx context.Context, request Request) (Result, error) {
	if err := validateRequest(request); err != nil {
		return Result{}, err
	}
	if v.Runner == nil {
		return Result{}, errors.New("validator runner is required")
	}
	observedHead, err := gitHead(ctx, request.WorkDir)
	if err != nil {
		return Result{}, err
	}
	if observedHead != request.CandidateHead {
		return Result{}, errors.New("candidate head moved before validation")
	}
	runner, err := v.Runner.WithWorkDir(request.WorkDir)
	if err != nil {
		return Result{}, err
	}
	runner, err = runner.WithModel(authority.RequiredValidator)
	if err != nil {
		return Result{}, err
	}
	var observedModel, observedThinking string
	result := runner.Run(ctx, "validate-milestone", nil, gsd.Observer{Event: func(event gsd.Event) {
		if event.Kind == gsd.EventModelSelect && event.Model != "" && observedModel == "" {
			observedModel = event.Model
		}
		if event.Thinking != "" && observedThinking == "" {
			observedThinking = event.Thinking
		}
	}})
	if result.Terminal != gsd.TerminalSuccess {
		if result.Err != nil {
			return Result{}, fmt.Errorf("independent validation failed: %w", result.Err)
		}
		return Result{}, fmt.Errorf("independent validation failed: %s", result.Terminal)
	}
	sessionID := ""
	if v.SessionsDir != "" {
		if id, sessionErr := gsd.LatestSessionID(v.SessionsDir, request.WorkDir); sessionErr == nil {
			sessionID = id
		}
		if observedModel == "" || observedThinking == "" {
			model, thinking, identityErr := gsd.ReadSessionIdentity(v.SessionsDir, request.WorkDir)
			if identityErr == nil {
				if observedModel == "" {
					observedModel = model
				}
				if observedThinking == "" {
					observedThinking = thinking
				}
			}
		}
	}
	if sessionID == "" {
		sessionID = derivedSessionID(request)
	}
	proof, err := readValidationProof(request.WorkDir)
	if err != nil {
		return Result{}, err
	}
	return Result{
		ObservedModel: observedModel, Thinking: observedThinking, SessionID: sessionID,
		Verdict: proof.Verdict, ObservedHead: observedHead, EvidenceHash: proof.EvidenceHash,
		LocalGates: proof.LocalGates, UAT: proof.UAT, MilestoneValid: proof.MilestoneValid,
		IssuedAt: proof.IssuedAt, ExpiresAt: proof.ExpiresAt,
	}, nil
}

type proofFile struct {
	Verdict        string    `json:"verdict"`
	EvidenceHash   string    `json:"evidence_hash"`
	LocalGates     bool      `json:"local_gates"`
	UAT            bool      `json:"uat"`
	MilestoneValid bool      `json:"milestone_valid"`
	IssuedAt       time.Time `json:"issued_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

func readValidationProof(workDir string) (proofFile, error) {
	path := filepath.Join(workDir, ".gsd", "shepherd-validation.json")
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
	if request.Repository == "" || request.PullRequest <= 0 || request.BaseBranch == "" || request.DeliveryID == "" || request.UnitID == "" || request.UnitType == "" || request.WorkDir == "" || request.GSDHome == "" {
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

func derivedSessionID(request Request) string {
	hash := sha256.Sum256([]byte(request.DeliveryID + ":" + request.UnitID + ":" + request.CandidateHead + ":" + request.EvidenceHash))
	encoded := hex.EncodeToString(hash[:])
	return encoded[:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32]
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
