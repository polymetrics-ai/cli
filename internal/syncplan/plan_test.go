package syncplan

import (
	"encoding/json"
	"testing"

	"polymetrics.ai/internal/synccontract"
)

func validPlan() Plan {
	return Plan{
		ContractVersion: ContractVersion,
		SourceBinding:   "stream.orders",
		TargetBinding:   "action.upsert_orders",
		Mode:            synccontract.ModeIncrementalUpsert,
		Axes: synccontract.ExecutionAxes{
			Progression: synccontract.SourceProgressionCursor,
			Apply:       synccontract.DestinationApplyAppend,
			Object:      synccontract.ObjectClassObject,
			Binding:     synccontract.BindingKindStream,
			Key:         synccontract.KeyShapeStable,
			Delete:      synccontract.DeletePolicyNone,
		},
	}
}

func TestResultRequiresExactlyOneClosedOutcome(t *testing.T) {
	cases := []Result{
		{Kind: ResultKindExecutable, Plan: ptr(validPlan())},
		{Kind: ResultKindIncompatible, Incompatibility: &Incompatibility{Axis: "key", Code: "stable_key_required"}},
		{Kind: ResultKindFoundationGap, FoundationGap: &FoundationGap{FoundationID: "runtime.example.v1", Reference: "provider.invalid/docs"}},
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

func TestPlanAndResultRoundTripWithoutExecutorState(t *testing.T) {
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
	if string(encoded) != `{"kind":"executable","plan":{"contract_version":1,"source_binding":"stream.orders","target_binding":"action.upsert_orders","mode":"incremental_upsert","axes":{"progression":"cursor","apply":"append","object":"object","binding":"stream","key":"stable","delete":"none","budget":{}}}}` {
		t.Fatalf("canonical result bytes changed: %s", encoded)
	}
}

func ptr(plan Plan) *Plan { return &plan }
