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
	plan.Axes.Progression = synccontract.SourceProgressionSnapshot
	plan.Axes.Apply = synccontract.DestinationApplyAppend
	if got := Resolve(plan, synccontract.Budget{}); got.Kind != ResultKindIncompatible || got.Incompatibility.Axis != "apply" {
		t.Fatalf("incompatible axes = %#v", got)
	}
}
