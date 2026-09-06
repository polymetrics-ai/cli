package syncplan

import (
	"encoding/json"
	"strings"
	"testing"

	"polymetrics.ai/internal/synccontract"
)

func testDigest(marker string) string {
	return "sha256:" + strings.Repeat(marker, 64)
}

func validPlan() Plan {
	axes, ok := synccontract.ModeAxes(synccontract.ModeIncrementalUpsert)
	if !ok {
		panic("incremental_upsert mode is missing")
	}
	return Plan{
		ContractVersion:  ContractVersion,
		Source:           BindingRef{Kind: synccontract.BindingKindStream, ID: "stream.orders"},
		Target:           BindingRef{Kind: synccontract.BindingKindAction, ID: "action.upsert_orders"},
		Mode:             synccontract.ModeIncrementalUpsert,
		Axes:             axes,
		GenerationDigest: testDigest("1"),
		ArtifactDigest:   testDigest("2"),
		Executors: []ExecutorRef{
			{Role: ExecutorRoleDestination, ID: "declarative_api.destination.v1", Digest: testDigest("3")},
			{Role: ExecutorRoleSource, ID: "declarative_api.source.v1", Digest: testDigest("4")},
		},
		Foundation:     FoundationRef{ID: "runtime.declarative.v1", Digest: testDigest("5"), Available: true, Reference: "atlas/runtime.declarative.v1"},
		EvidenceDigest: testDigest("6"),
	}
}

func TestResultRequiresExactlyOneClosedOutcome(t *testing.T) {
	cases := []Result{
		{Kind: ResultKindExecutable, Plan: ptr(validPlan())},
		{Kind: ResultKindIncompatible, Incompatibility: &Incompatibility{Axis: "key", Code: "stable_key_required"}},
		{Kind: ResultKindFoundationGap, FoundationGap: &FoundationGap{FoundationID: "runtime.example.v1", Reference: "atlas/runtime.example.v1"}},
	}
	for _, result := range cases {
		if err := result.Validate(); err != nil {
			t.Fatalf("valid result %#v: %v", result, err)
		}
	}
	if err := (Result{Kind: ResultKindExecutable}).Validate(); err == nil {
		t.Fatal("executable result without plan accepted")
	}
	if err := (Result{Kind: ResultKind("unknown")}).Validate(); err == nil {
		t.Fatal("unknown result kind accepted")
	}
}

func TestPlanAndResultCanonicalRoundTripPreservesIdentityDigests(t *testing.T) {
	original := Result{Kind: ResultKindExecutable, Plan: ptr(validPlan())}
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Result
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if err := decoded.Validate(); err != nil {
		t.Fatal(err)
	}
	reencoded, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != string(reencoded) {
		t.Fatalf("canonical result bytes changed:\nfirst=%s\nsecond=%s", encoded, reencoded)
	}
	if decoded.Plan.GenerationDigest != original.Plan.GenerationDigest || decoded.Plan.ArtifactDigest != original.Plan.ArtifactDigest || decoded.Plan.Executors[0].Digest != original.Plan.Executors[0].Digest || decoded.Plan.Foundation.Digest != original.Plan.Foundation.Digest || decoded.Plan.EvidenceDigest != original.Plan.EvidenceDigest {
		t.Fatalf("digest bindings changed: %#v", decoded.Plan)
	}
}

func TestPlanRejectsUnsortedOrMalformedIdentity(t *testing.T) {
	plan := validPlan()
	plan.Executors = []ExecutorRef{
		{Role: ExecutorRoleSource, ID: "z.executor", Digest: testDigest("3")},
		{Role: ExecutorRoleDestination, ID: "a.executor", Digest: testDigest("4")},
	}
	if err := plan.Validate(); err == nil {
		t.Fatal("unsorted executor selection accepted")
	}
	plan = validPlan()
	plan.GenerationDigest = "not-a-digest"
	if err := plan.Validate(); err == nil {
		t.Fatal("malformed generation digest accepted")
	}
}

func ptr(plan Plan) *Plan { return &plan }
