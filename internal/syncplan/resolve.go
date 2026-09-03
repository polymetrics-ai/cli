package syncplan

import (
	"fmt"

	"polymetrics.ai/internal/synccontract"
)

// Resolve is a pure CP03 resolver. It classifies only immutable contract facts;
// executor and foundation lookup remain later checkpoints.
func Resolve(plan Plan, caller synccontract.Budget) Result {
	if err := plan.Validate(); err != nil {
		return Result{Kind: ResultKindIncompatible, Incompatibility: &Incompatibility{Axis: axisFor(err), Code: "invalid_contract"}}
	}
	if err := caller.Validate(); err != nil || widens(caller, plan.Axes.Budget) {
		return Result{Kind: ResultKindIncompatible, Incompatibility: &Incompatibility{Axis: "budget", Code: "caller_budget_widens_declaration"}}
	}
	if code := incompatible(plan.Axes); code != "" {
		return Result{Kind: ResultKindIncompatible, Incompatibility: &Incompatibility{Axis: code, Code: "incompatible_axes"}}
	}
	sealed := plan
	sealed.Axes.Budget = reduce(plan.Axes.Budget, caller)
	return Result{Kind: ResultKindExecutable, Plan: &sealed}
}

func axisFor(err error) string {
	var axis *synccontract.UnknownAxisValueError
	if fmt.Errorf("%w", err) != nil && errorAs(err, &axis) {
		return axis.Axis
	}
	return "budget"
}

func errorAs(err error, target **synccontract.UnknownAxisValueError) bool {
	if typed, ok := err.(*synccontract.UnknownAxisValueError); ok {
		*target = typed
		return true
	}
	return false
}

func incompatible(a synccontract.ExecutionAxes) string {
	if a.Progression == synccontract.SourceProgressionSnapshot && a.Apply != synccontract.DestinationApplyReplace {
		return "apply"
	}
	if a.Progression == synccontract.SourceProgressionChangeCapture && (a.Apply != synccontract.DestinationApplyChangeApply || a.Delete != synccontract.DeletePolicyTombstone) {
		return "delete"
	}
	return ""
}

func widens(caller, declared synccontract.Budget) bool {
	return (declared.MaxRecords > 0 && caller.MaxRecords > declared.MaxRecords) || (declared.MaxBytes > 0 && caller.MaxBytes > declared.MaxBytes) || (declared.MaxBatches > 0 && caller.MaxBatches > declared.MaxBatches)
}
func reduce(declared, caller synccontract.Budget) synccontract.Budget {
	if caller.MaxRecords > 0 && (declared.MaxRecords == 0 || caller.MaxRecords < declared.MaxRecords) {
		declared.MaxRecords = caller.MaxRecords
	}
	if caller.MaxBytes > 0 && (declared.MaxBytes == 0 || caller.MaxBytes < declared.MaxBytes) {
		declared.MaxBytes = caller.MaxBytes
	}
	if caller.MaxBatches > 0 && (declared.MaxBatches == 0 || caller.MaxBatches < declared.MaxBatches) {
		declared.MaxBatches = caller.MaxBatches
	}
	return declared
}
