import {
	decideFailurePolicy,
	evaluateLifecycleTransition,
	isParentLifecycleStage,
	type FailureClass,
	type LifecycleFacts,
	type ParentLifecycleStage,
	type RepositoryBlocker,
} from "./autonomy-policy.ts";
import {
	DependencyGraphError,
	selectReadyWork,
	validateDependencyGraph,
	type DependencyWorkItem,
	type ReadyQueueSelection,
} from "./dependency-graph.ts";

export interface PersistedAutonomyState {
	stage: ParentLifecycleStage;
	retryAttempts: number;
	correctionRounds: number;
}

export interface SpawnConstraints {
	runtimeCapabilityAvailable: boolean;
	isolationAvailable: boolean;
	hardHumanGate: boolean;
	verificationBlocked: boolean;
	reviewBlocked: boolean;
}

export interface CanonicalAutonomySnapshot {
	observedStage: ParentLifecycleStage;
	proposedStage?: ParentLifecycleStage;
	transitionFacts?: LifecycleFacts;
	workItems: DependencyWorkItem[];
	maxConcurrency: number;
	constraints: SpawnConstraints;
}

export interface RetryBudget {
	maxRetries: number;
	maxCorrectionRounds: number;
}

export interface ReconcileInput {
	persisted: PersistedAutonomyState;
	canonical: CanonicalAutonomySnapshot;
	budget: RetryBudget;
	failure?: FailureClass;
}

export type ReconcileDecision =
	| { kind: "reconcile_stage"; stage: ParentLifecycleStage; reason: "canonical_stage_differs" }
	| { kind: "transition"; from: ParentLifecycleStage; to: ParentLifecycleStage; reason: string }
	| { kind: "spawn"; itemIds: string[] }
	| { kind: "no_spawn"; blocker: RepositoryBlocker; reason: string }
	| { kind: "retry"; failure: Exclude<FailureClass, "hard_human_gate">; nextRetryAttempts: number; remainingRetries: number }
	| { kind: "correct"; failure: Exclude<FailureClass, "hard_human_gate">; nextCorrectionRounds: number; remainingCorrections: number }
	| { kind: "at_capacity" }
	| { kind: "await_stage_evidence"; stage: ParentLifecycleStage }
	| { kind: "complete" };

function noSpawn(blocker: RepositoryBlocker, reason: string): ReconcileDecision {
	return { kind: "no_spawn", blocker, reason };
}

function blockerForUnsafeTransition(from: ParentLifecycleStage, to: ParentLifecycleStage): RepositoryBlocker {
	if (from === "VERIFY" && to === "REVIEW") return "not_spawned_verification_blocked";
	if (from === "REVIEW" && to === "INTEGRATE") return "not_spawned_review_blocked";
	return "not_spawned_human_gate";
}

function invalidGraphDecision(error: DependencyGraphError): ReconcileDecision {
	const blocker: RepositoryBlocker = error.code === "ambiguous_scope"
		? "not_spawned_write_scope_collision"
		: "not_spawned_dependency_blocked";
	return noSpawn(blocker, `invalid dependency graph: ${error.code}`);
}

function qualityConstraintDecision(input: ReconcileInput): ReconcileDecision | undefined {
	const constraints = input.canonical.constraints;
	if (constraints.hardHumanGate) {
		return noSpawn("not_spawned_human_gate", "hard human gate requires an authenticated decision");
	}
	if (constraints.verificationBlocked) {
		return noSpawn("not_spawned_verification_blocked", "verification is blocked");
	}
	if (constraints.reviewBlocked) return noSpawn("not_spawned_review_blocked", "review is blocked");
	return undefined;
}

function spawnCapabilityDecision(input: ReconcileInput): ReconcileDecision | undefined {
	const constraints = input.canonical.constraints;
	if (!constraints.runtimeCapabilityAvailable) {
		return noSpawn("not_spawned_runtime_capability_missing", "required runtime capability is unavailable");
	}
	const mutatingWorkRemains = input.canonical.workItems.some((candidate) =>
		candidate.status === "pending" && candidate.access === "mutating",
	);
	if (!constraints.isolationAvailable && mutatingWorkRemains) {
		return noSpawn("not_spawned_isolation_missing", "mutating work lacks an isolated worktree");
	}
	return undefined;
}

export function reconcileAutonomy(input: ReconcileInput): ReconcileDecision {
	if (!isParentLifecycleStage(input.persisted.stage)
		|| !isParentLifecycleStage(input.canonical.observedStage)
		|| (input.canonical.proposedStage !== undefined && !isParentLifecycleStage(input.canonical.proposedStage))) {
		return noSpawn("not_spawned_human_gate", "invalid lifecycle stage");
	}
	if (input.persisted.stage !== input.canonical.observedStage) {
		return { kind: "reconcile_stage", stage: input.canonical.observedStage, reason: "canonical_stage_differs" };
	}

	try {
		validateDependencyGraph(input.canonical.workItems);
	} catch (error) {
		if (error instanceof DependencyGraphError) return invalidGraphDecision(error);
		throw error;
	}

	if (input.canonical.observedStage === "COMPLETE") return { kind: "complete" };

	const constrained = qualityConstraintDecision(input);
	if (constrained !== undefined) return constrained;

	if (input.failure !== undefined) {
		const failureDecision = decideFailurePolicy({
			failure: input.failure,
			retryAttempts: input.persisted.retryAttempts,
			maxRetries: input.budget.maxRetries,
			correctionRounds: input.persisted.correctionRounds,
			maxCorrectionRounds: input.budget.maxCorrectionRounds,
		});
		if (input.failure === "hard_human_gate") {
			return noSpawn("not_spawned_human_gate", "hard human gate requires an authenticated decision");
		}
		if (failureDecision.action === "retry") {
			return {
				kind: "retry",
				failure: input.failure,
				nextRetryAttempts: input.persisted.retryAttempts + 1,
				remainingRetries: failureDecision.remainingRetries,
			};
		}
		if (failureDecision.action === "correct") {
			return {
				kind: "correct",
				failure: input.failure,
				nextCorrectionRounds: input.persisted.correctionRounds + 1,
				remainingCorrections: failureDecision.remainingCorrections,
			};
		}
		return noSpawn("not_spawned_human_gate", "retry and correction budgets are exhausted");
	}

	if (input.canonical.proposedStage !== undefined) {
		const transition = evaluateLifecycleTransition({
			from: input.canonical.observedStage,
			to: input.canonical.proposedStage,
			facts: input.canonical.transitionFacts ?? {},
		});
		if (!transition.allowed) {
			return noSpawn(
				blockerForUnsafeTransition(input.canonical.observedStage, input.canonical.proposedStage),
				transition.reason,
			);
		}
		return {
			kind: "transition",
			from: input.canonical.observedStage,
			to: input.canonical.proposedStage,
			reason: transition.reason,
		};
	}

	if (input.canonical.observedStage !== "SCHEDULE") {
		return { kind: "await_stage_evidence", stage: input.canonical.observedStage };
	}
	let selection: ReadyQueueSelection;
	try {
		selection = selectReadyWork(input.canonical.workItems, { maxConcurrency: input.canonical.maxConcurrency });
	} catch (error) {
		if (error instanceof RangeError) {
			return noSpawn("not_spawned_runtime_capability_missing", "invalid concurrency policy");
		}
		throw error;
	}
	if (selection.kind === "complete") {
		return { kind: "transition", from: "SCHEDULE", to: "FINAL_VERIFY", reason: "all_tasks_integrated" };
	}
	if (selection.kind === "at_capacity") return { kind: "at_capacity" };
	const spawnConstrained = spawnCapabilityDecision(input);
	if (spawnConstrained !== undefined) return spawnConstrained;
	if (selection.kind === "selected") return { kind: "spawn", itemIds: selection.itemIds };
	if (selection.kind === "blocked") return noSpawn(selection.blocker, selection.blocker);
	return { kind: "at_capacity" };
}
