package certify

import (
	"context"
	"strings"
	"testing"
)

func TestFullParityClaimRefusesNonLiveWriteAction(t *testing.T) {
	rc := &runContext{ctx: context.Background(), opts: Options{Full: true, Write: true, RequireFullParity: true}}
	rep := Report{Capabilities: Capabilities{WriteActions: map[string]WriteActionResult{
		"create_issue": {Result: "pass"},
		"update_issue": {Result: "not_live", Path: "/repos/{owner}/{repo}/issues/{issue_number}", Risk: "requires run-owned fixture"},
	}}}
	if err := stageFullParityClaim(rc, &rep); err != nil {
		t.Fatalf("stageFullParityClaim() = %v", err)
	}
	if rep.FullParityVerified() {
		t.Fatal("FullParityVerified() = true with a non-live action")
	}
	if len(rep.Stages) != 1 || rep.Stages[0].Passed || !strings.Contains(rep.Stages[0].Error, "update_issue") {
		t.Fatalf("full parity stage = %+v, want named refusal", rep.Stages)
	}
}

func TestFullParityClaimAcceptsOnlyProviderVerifiedActions(t *testing.T) {
	rc := &runContext{ctx: context.Background(), opts: Options{Full: true, Write: true, RequireFullParity: true}}
	rep := Report{Capabilities: Capabilities{WriteActions: map[string]WriteActionResult{
		"create_issue": {Result: "pass"},
		"update_issue": {Result: "pass"},
	}}}
	if err := stageFullParityClaim(rc, &rep); err != nil {
		t.Fatalf("stageFullParityClaim() = %v", err)
	}
	if !rep.FullParityVerified() {
		t.Fatalf("full parity stage = %+v, want verified", rep.Stages)
	}
}
