package authority

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

const RequiredValidator = "openai-codex/gpt-5.6-sol"

var shaPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
var hashPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type RatificationRequest struct {
	Repository     string
	PR             int
	BaseBranch     string
	BaseSHA        string
	CandidateHead  string
	ObservedHead   string
	RunID          string
	Generation     int64
	UnitID         string
	Attempt        int64
	StateVersion   int64
	ContractHash   string
	EvidenceHash   string
	Validator      string
	Thinking       string
	Verdict        string
	LocalGates     bool
	UAT            bool
	MilestoneValid bool
	IssuedAt       time.Time
	ExpiresAt      time.Time
}

type Attestation struct {
	Repository   string
	PR           int
	BaseBranch   string
	BaseSHA      string
	HeadSHA      string
	RunID        string
	Generation   int64
	UnitID       string
	Attempt      int64
	StateVersion int64
	ContractHash string
	EvidenceHash string
	Validator    string
	Thinking     string
	Verdict      string
	IssuedAt     time.Time
	ExpiresAt    time.Time
}

func Ratify(request RatificationRequest, now time.Time) (Attestation, error) {
	if request.Repository == "" || request.PR <= 0 || request.BaseBranch == "" || request.RunID == "" || request.UnitID == "" {
		return Attestation{}, errors.New("repository, PR, branch, run, and unit identity are required")
	}
	for label, sha := range map[string]string{"base": request.BaseSHA, "candidate": request.CandidateHead, "observed": request.ObservedHead} {
		if !shaPattern.MatchString(sha) {
			return Attestation{}, fmt.Errorf("%s SHA is invalid", label)
		}
	}
	if request.CandidateHead != request.ObservedHead {
		return Attestation{}, errors.New("candidate evidence moved before ratification")
	}
	if request.Generation <= 0 || request.Attempt <= 0 || request.StateVersion <= 0 ||
		!hashPattern.MatchString(request.ContractHash) || !hashPattern.MatchString(request.EvidenceHash) {
		return Attestation{}, errors.New("generation, attempt, state version, and evidence hashes are required")
	}
	if request.Validator != RequiredValidator || request.Thinking != "high" {
		return Attestation{}, fmt.Errorf("validator must be %s with high thinking", RequiredValidator)
	}
	if request.Verdict != "PROCEED" || !request.LocalGates || !request.UAT || !request.MilestoneValid {
		return Attestation{}, errors.New("ratification requires PROCEED and all local, UAT, and milestone gates")
	}
	if request.IssuedAt.IsZero() || request.IssuedAt.After(now) || request.ExpiresAt.IsZero() || !request.ExpiresAt.After(request.IssuedAt) || !now.Before(request.ExpiresAt) {
		return Attestation{}, errors.New("attestation validity window is invalid or expired")
	}
	return Attestation{
		Repository: request.Repository, PR: request.PR, BaseBranch: request.BaseBranch, BaseSHA: request.BaseSHA,
		HeadSHA: request.CandidateHead, RunID: request.RunID, Generation: request.Generation,
		UnitID: request.UnitID, Attempt: request.Attempt, StateVersion: request.StateVersion,
		ContractHash: request.ContractHash, EvidenceHash: request.EvidenceHash,
		Validator: request.Validator, Thinking: request.Thinking, Verdict: request.Verdict,
		IssuedAt: request.IssuedAt.UTC(), ExpiresAt: request.ExpiresAt.UTC(),
	}, nil
}

func (a Attestation) Recheck(repository string, pr int, baseBranch, observedBase, observedHead string, generation, stateVersion int64, now time.Time) error {
	if a.Repository != repository || a.PR != pr || a.BaseBranch != baseBranch || a.BaseSHA != observedBase || a.HeadSHA != observedHead {
		return errors.New("ratified target or head moved")
	}
	if a.Generation != generation || a.StateVersion != stateVersion {
		return errors.New("ratification is stale for the current governance state")
	}
	if !now.Before(a.ExpiresAt) {
		return errors.New("ratification expired")
	}
	return nil
}
