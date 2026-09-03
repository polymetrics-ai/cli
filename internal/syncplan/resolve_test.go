package syncplan

import (
	"testing"

	"polymetrics.ai/internal/synccontract"
)

func TestResolveSealsOrClassifiesBeforeIO(t *testing.T) {
	plan := validPlan()
	plan.Axes.Budget = synccontract.Budget{MaxRecords: 100}
	result := Resolve(plan, synccontract.Budget{MaxRecords: 10})
	if err := result.Validate(); err != nil || result.Kind != ResultKindExecutable || result.Plan.Axes.Budget.MaxRecords != 10 {
		t.Fatalf("reduced result = %#v, err=%v", result, err)
	}
	if got := Resolve(plan, synccontract.Budget{MaxRecords: 101}); got.Kind != ResultKindIncompatible || got.Incompatibility.Axis != "budget" {
		t.Fatalf("widened budget = %#v", got)
	}
}

func TestResolveRejectsModeApplyContradiction(t *testing.T) {
	plan := validPlan()
	plan.Mode = synccontract.ModeIncrementalUpsert
	plan.Axes.Apply = synccontract.DestinationApplyAppend

	result := Resolve(plan, synccontract.Budget{})
	if result.Kind != ResultKindIncompatible || result.Incompatibility == nil || result.Incompatibility.Axis != "apply" {
		t.Fatalf("incremental_upsert + append result = %#v, want apply C3 incompatibility", result)
	}
}

func TestResolveClassifiesRealInvalidAxis(t *testing.T) {
	cases := []struct {
		name string
		edit func(*Plan)
		axis string
	}{
		{"contract version", func(plan *Plan) { plan.ContractVersion++ }, "contract_version"},
		{"mode", func(plan *Plan) { plan.Mode = "unknown" }, "mode"},
		{"binding", func(plan *Plan) { plan.Source.Kind = "route" }, "binding"},
		{"identity", func(plan *Plan) { plan.EvidenceDigest = "bad" }, "identity"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			plan := validPlan()
			test.edit(&plan)
			result := Resolve(plan, synccontract.Budget{})
			if result.Kind != ResultKindIncompatible || result.Incompatibility == nil || result.Incompatibility.Axis != test.axis {
				t.Fatalf("result = %#v, want %s C3", result, test.axis)
			}
		})
	}
}

func TestResolveClassifiesFoundationGapAndAmbiguousExecutor(t *testing.T) {
	plan := validPlan()
	plan.Foundation.Available = false
	result := Resolve(plan, synccontract.Budget{})
	if result.Kind != ResultKindFoundationGap || result.FoundationGap == nil || result.FoundationGap.FoundationID != plan.Foundation.ID {
		t.Fatalf("unavailable foundation result = %#v", result)
	}

	plan = validPlan()
	plan.Executors = append(plan.Executors, ExecutorRef{Role: ExecutorRoleSource, ID: "second.executor", Digest: testDigest("7")})
	result = Resolve(plan, synccontract.Budget{})
	if result.Kind != ResultKindIncompatible || result.Incompatibility == nil || result.Incompatibility.Axis != "executor" || result.Incompatibility.Code != "ambiguous_executor_selection" {
		t.Fatalf("ambiguous executor result = %#v", result)
	}
}

func TestResolveAcceptsCanonicalModeMatrix(t *testing.T) {
	for _, mode := range synccontract.AllModes() {
		t.Run(string(mode), func(t *testing.T) {
			plan := validPlan()
			axes, ok := synccontract.ModeAxes(mode)
			if !ok {
				t.Fatalf("ModeAxes(%q) missing", mode)
			}
			plan.Mode = mode
			plan.Axes = axes
			result := Resolve(plan, synccontract.Budget{})
			if err := result.Validate(); err != nil || result.Kind != ResultKindExecutable {
				t.Fatalf("canonical %q result = %#v, err=%v", mode, result, err)
			}
		})
	}
}
