package authority

import (
	"testing"
	"time"
)

func validRequest(now time.Time) RatificationRequest {
	return RatificationRequest{
		Repository: "polymetrics-ai/cli", PR: 380, BaseBranch: "main",
		BaseSHA:       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		CandidateHead: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		ObservedHead:  "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		RunID:         "run-1", Generation: 2, UnitID: "M001/S01/U01", Attempt: 1, StateVersion: 7,
		ContractHash: "sha256:contract", EvidenceHash: "sha256:evidence",
		Validator: RequiredValidator, Thinking: "high", Verdict: "PROCEED",
		LocalGates: true, UAT: true, MilestoneValid: true,
		IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(10 * time.Minute),
	}
}

func TestRatificationBindsExactEvidence(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0).UTC()
	request := validRequest(now)
	attestation, err := Ratify(request, now)
	if err != nil {
		t.Fatalf("valid ratification failed: %v", err)
	}
	if err := attestation.Recheck(request.Repository, request.PR, request.CandidateHead, request.Generation, request.StateVersion, now); err != nil {
		t.Fatalf("valid recheck failed: %v", err)
	}
	if err := attestation.Recheck(request.Repository, request.PR, "cccccccccccccccccccccccccccccccccccccccc", request.Generation, request.StateVersion, now); err == nil {
		t.Fatal("expected delivery-time moved head to fail")
	}
}

func TestRatificationRejectsStaleHeadModelAndExpiry(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0).UTC()
	request := validRequest(now)
	request.ObservedHead = "cccccccccccccccccccccccccccccccccccccccc"
	if _, err := Ratify(request, now); err == nil {
		t.Fatal("expected stale head to fail")
	}
	request = validRequest(now)
	request.Validator = "openai-codex/gpt-5.5"
	if _, err := Ratify(request, now); err == nil {
		t.Fatal("expected validator downgrade to fail")
	}
	request = validRequest(now)
	request.ExpiresAt = now
	if _, err := Ratify(request, now); err == nil {
		t.Fatal("expected expired ratification to fail")
	}
}
