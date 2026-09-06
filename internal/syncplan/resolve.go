package syncplan

import (
	"errors"

	"polymetrics.ai/internal/synccontract"
)

// Resolve is a pure deterministic resolver over immutable plan facts. It
// returns exactly one executable plan, C3 incompatibility, or C4 foundation
// gap without constructing an executor or crossing an I/O boundary.
func Resolve(plan Plan, caller synccontract.Budget) Result {
	if err := plan.Validate(); err != nil {
		return incompatibleResult(axisFor(err), codeFor(err, "invalid_contract"))
	}
	if err := caller.Validate(); err != nil {
		return incompatibleResult(axisFor(err), codeFor(err, "invalid_budget"))
	}
	if widens(caller, plan.Axes.Budget) {
		return incompatibleResult("budget", "caller_budget_widens_declaration")
	}
	if !plan.Foundation.Available {
		return Result{Kind: ResultKindFoundationGap, FoundationGap: &FoundationGap{FoundationID: plan.Foundation.ID, Reference: plan.Foundation.Reference}}
	}
	if !hasExactlyOneExecutorPerRole(plan.Executors) {
		return incompatibleResult("executor", "ambiguous_executor_selection")
	}
	if err := plan.ValidateExecutable(); err != nil {
		return incompatibleResult(axisFor(err), codeFor(err, "incompatible_axes"))
	}
	sealed := plan
	sealed.Axes.Budget = reduce(plan.Axes.Budget, caller)
	return Result{Kind: ResultKindExecutable, Plan: &sealed}
}

func incompatibleResult(axis, code string) Result {
	return Result{Kind: ResultKindIncompatible, Incompatibility: &Incompatibility{Axis: axis, Code: code}}
}

func axisFor(err error) string {
	var axis interface{ SyncAxis() string }
	if errors.As(err, &axis) && axis.SyncAxis() != "" {
		return axis.SyncAxis()
	}
	return "identity"
}

func codeFor(err error, fallback string) string {
	var validation *ValidationError
	if errors.As(err, &validation) && validation.Code != "" {
		return validation.Code
	}
	var axis *synccontract.AxisError
	if errors.As(err, &axis) && axis.Code != "" {
		return axis.Code
	}
	return fallback
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
