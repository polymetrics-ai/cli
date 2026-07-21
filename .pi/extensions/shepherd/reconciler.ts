import { types as nodeTypes } from "node:util";

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
	| { kind: "await_human_decision"; blocker: "not_spawned_human_gate"; reason: "hard_human_gate" | "retry_budget_exhausted" | "pending_authenticated_decision" }
	| { kind: "retry"; failure: Exclude<FailureClass, "hard_human_gate">; nextRetryAttempts: number; remainingRetries: number }
	| { kind: "correct"; failure: Exclude<FailureClass, "hard_human_gate">; nextCorrectionRounds: number; remainingCorrections: number }
	| { kind: "at_capacity" }
	| { kind: "await_stage_evidence"; stage: ParentLifecycleStage }
	| { kind: "invalid_snapshot"; reason: string }
	| { kind: "aborted"; reason: "human_rejected" }
	| { kind: "blocked"; reason: "terminal_blocked" }
	| { kind: "complete" };

function noSpawn(blocker: RepositoryBlocker, reason: string): Extract<ReconcileDecision, { kind: "no_spawn" }> {
	return { kind: "no_spawn", blocker, reason };
}

function awaitHuman(reason: Extract<ReconcileDecision, { kind: "await_human_decision" }>["reason"]): ReconcileDecision {
	return { kind: "await_human_decision", blocker: "not_spawned_human_gate", reason };
}

function invalidGraphDecision(error: DependencyGraphError): Extract<ReconcileDecision, { kind: "no_spawn" }> {
	const blocker: RepositoryBlocker = error.code === "ambiguous_scope" || error.code === "conflict_component_too_large"
		? "not_spawned_write_scope_collision"
		: "not_spawned_dependency_blocked";
	return noSpawn(blocker, `invalid dependency graph: ${error.code}`);
}

function qualityConstraintDecision(input: ReconcileInput): ReconcileDecision | undefined {
	const constraints = input.canonical.constraints;
	if (constraints.hardHumanGate) {
		return awaitHuman("hard_human_gate");
	}
	if (constraints.verificationBlocked) {
		return noSpawn("not_spawned_verification_blocked", "verification is blocked");
	}
	if (constraints.reviewBlocked) return noSpawn("not_spawned_review_blocked", "review is blocked");
	return undefined;
}

function spawnCapabilityDecision(input: ReconcileInput, selectedIds: readonly string[]): ReconcileDecision | undefined {
	const constraints = input.canonical.constraints;
	if (!constraints.runtimeCapabilityAvailable) {
		return noSpawn("not_spawned_runtime_capability_missing", "required runtime capability is unavailable");
	}
	const selectedMutator = input.canonical.workItems.some((candidate) =>
		selectedIds.includes(candidate.id) && candidate.access === "mutating",
	);
	if (!constraints.isolationAvailable && selectedMutator) {
		return noSpawn("not_spawned_isolation_missing", "mutating work lacks an isolated worktree");
	}
	return undefined;
}

type OwnDataDescriptor = PropertyDescriptor & { value: unknown };
type OwnDataDescriptorMap = Record<string, OwnDataDescriptor>;

function isEnumerableDataDescriptor(descriptor: PropertyDescriptor | undefined): descriptor is OwnDataDescriptor {
	return descriptor !== undefined && descriptor.enumerable === true && Object.hasOwn(descriptor, "value");
}

function exactOwnDataDescriptors(
	value: unknown,
	requiredKeys: readonly string[],
	optionalKeys: readonly string[] = [],
): OwnDataDescriptorMap | undefined {
	if (typeof value !== "object" || value === null || nodeTypes.isProxy(value) || Array.isArray(value)) return undefined;
	const prototype = Object.getPrototypeOf(value);
	if (prototype !== Object.prototype && prototype !== null) return undefined;
	const descriptors = Object.getOwnPropertyDescriptors(value);
	const keys = Reflect.ownKeys(descriptors);
	const allowed = new Set([...requiredKeys, ...optionalKeys]);
	if (!requiredKeys.every((key) => Object.hasOwn(descriptors, key))) return undefined;
	for (const key of keys) {
		if (typeof key !== "string" || !allowed.has(key) || !isEnumerableDataDescriptor(descriptors[key])) {
			return undefined;
		}
	}
	return descriptors as OwnDataDescriptorMap;
}

function exactOwnArrayValues(value: unknown): unknown[] | undefined {
	if (typeof value !== "object" || value === null || nodeTypes.isProxy(value) || !Array.isArray(value)) {
		return undefined;
	}
	if (Object.getPrototypeOf(value) !== Array.prototype) return undefined;
	const descriptors = Object.getOwnPropertyDescriptors(value) as Record<string, PropertyDescriptor>;
	const keys = Reflect.ownKeys(descriptors);
	if (keys.some((key) => typeof key !== "string")) return undefined;
	const lengthDescriptor = descriptors.length;
	if (lengthDescriptor === undefined
		|| !Object.hasOwn(lengthDescriptor, "value")
		|| lengthDescriptor.enumerable !== false) return undefined;
	const length = lengthDescriptor.value;
	if (!isNonNegativeSafeInteger(length) || keys.length !== length + 1) return undefined;

	const values: unknown[] = [];
	for (let index = 0; index < length; index += 1) {
		const descriptor = descriptors[String(index)];
		if (!isEnumerableDataDescriptor(descriptor)) return undefined;
		values.push(descriptor.value);
	}
	return values;
}

function isNonNegativeSafeInteger(value: unknown): value is number {
	return typeof value === "number" && Number.isSafeInteger(value) && value >= 0;
}

const lifecycleBooleanFacts = new Set([
	"researchRequired", "researchComplete", "parentPlanComplete", "issuesCreated",
	"parentSetupComplete", "readyWorkAvailable", "allTasksIntegrated", "taskPlanComplete",
	"isolatedBranchReady", "executionCheckpointed", "subPrOpen", "verificationPassed",
	"reviewClean", "correctionRequired", "correctionCheckpointed", "integrationConfirmed",
	"finalVerificationPassed", "humanDecisionAuthenticated", "exactHeadRevalidated", "mergeConfirmed",
	"hardHumanGate",
]);

function isLifecycleFacts(value: unknown): value is LifecycleFacts {
	const descriptors = exactOwnDataDescriptors(value, [], [...lifecycleBooleanFacts, "humanDecision"]);
	if (descriptors === undefined) return false;
	return Object.entries(descriptors).every(([key, descriptor]) => key === "humanDecision"
		? descriptor.value === "pending" || descriptor.value === "approve_merge" || descriptor.value === "reject"
		: lifecycleBooleanFacts.has(key) && typeof descriptor.value === "boolean");
}

function isWorkItemShape(value: unknown): value is DependencyWorkItem {
	const descriptors = exactOwnDataDescriptors(value, ["id", "dependsOn", "status", "access", "writeScopes"]);
	if (descriptors === undefined) return false;
	const dependencies = exactOwnArrayValues(descriptors.dependsOn.value);
	const writeScopes = exactOwnArrayValues(descriptors.writeScopes.value);
	return typeof descriptors.id.value === "string"
		&& dependencies !== undefined && dependencies.every((dependency) => typeof dependency === "string")
		&& (descriptors.status.value === "pending" || descriptors.status.value === "running"
			|| descriptors.status.value === "succeeded" || descriptors.status.value === "failed"
			|| descriptors.status.value === "blocked")
		&& (descriptors.access.value === "read_only" || descriptors.access.value === "mutating")
		&& writeScopes !== undefined && writeScopes.every((scope) => typeof scope === "string");
}

function isReconcileInput(value: unknown): value is ReconcileInput {
	const root = exactOwnDataDescriptors(value, ["persisted", "canonical", "budget"], ["failure"]);
	if (root === undefined) return false;
	const persisted = exactOwnDataDescriptors(root.persisted.value, ["stage", "retryAttempts", "correctionRounds"]);
	if (persisted === undefined
		|| !isParentLifecycleStage(persisted.stage.value)
		|| !isNonNegativeSafeInteger(persisted.retryAttempts.value)
		|| !isNonNegativeSafeInteger(persisted.correctionRounds.value)) return false;
	const budget = exactOwnDataDescriptors(root.budget.value, ["maxRetries", "maxCorrectionRounds"]);
	if (budget === undefined
		|| !isNonNegativeSafeInteger(budget.maxRetries.value)
		|| !isNonNegativeSafeInteger(budget.maxCorrectionRounds.value)) return false;
	const canonical = exactOwnDataDescriptors(
		root.canonical.value,
		["observedStage", "workItems", "maxConcurrency", "constraints"],
		["proposedStage", "transitionFacts"],
	);
	if (canonical === undefined) return false;
	const workItems = exactOwnArrayValues(canonical.workItems.value);
	if (!isParentLifecycleStage(canonical.observedStage.value)
		|| !isNonNegativeSafeInteger(canonical.maxConcurrency.value)
		|| workItems === undefined
		|| !workItems.every(isWorkItemShape)
		|| (Object.hasOwn(canonical, "proposedStage") && !isParentLifecycleStage(canonical.proposedStage.value))
		|| (Object.hasOwn(canonical, "transitionFacts") && !isLifecycleFacts(canonical.transitionFacts.value))) return false;
	const constraints = exactOwnDataDescriptors(canonical.constraints.value, [
		"runtimeCapabilityAvailable", "isolationAvailable", "hardHumanGate",
		"verificationBlocked", "reviewBlocked",
	]);
	if (constraints === undefined
		|| !Object.values(constraints).every((descriptor) => typeof descriptor.value === "boolean")) return false;
	return !Object.hasOwn(root, "failure")
		|| root.failure.value === "transient_verification"
		|| root.failure.value === "transient_review"
		|| root.failure.value === "hard_human_gate";
}

function deepFreeze<T>(value: T): T {
	if (typeof value !== "object" || value === null || Object.isFrozen(value)) return value;
	for (const child of Object.values(value)) deepFreeze(child);
	return Object.freeze(value);
}

function snapshotReconcileInput(candidate: unknown): ReconcileInput | undefined {
	try {
		if (!isReconcileInput(candidate)) return undefined;
		const snapshot: unknown = structuredClone(candidate);
		return isReconcileInput(snapshot) ? deepFreeze(snapshot) : undefined;
	} catch {
		return undefined;
	}
}

type ScheduleSelectionFailure = Extract<ReconcileDecision, { kind: "invalid_snapshot" | "no_spawn" }>;
type ScheduleSelectionResult = ReadyQueueSelection | ScheduleSelectionFailure;

function selectScheduleWork(input: ReconcileInput): ScheduleSelectionResult {
	try {
		return selectReadyWork(input.canonical.workItems, { maxConcurrency: input.canonical.maxConcurrency });
	} catch (error) {
		if (error instanceof DependencyGraphError) return invalidGraphDecision(error);
		if (error instanceof RangeError) {
			return noSpawn("not_spawned_runtime_capability_missing", "invalid concurrency policy");
		}
		return { kind: "invalid_snapshot", reason: "invalid autonomy snapshot" };
	}
}

function scheduleSelectionMatchesStage(
	selection: ReadyQueueSelection,
	proposedStage: ParentLifecycleStage,
): boolean {
	if (proposedStage === "FINAL_VERIFY") return selection.kind === "complete";
	if (proposedStage === "TASK_PLAN") return selection.kind === "selected";
	return true;
}

export function reconcileAutonomy(candidate: unknown): ReconcileDecision {
	const input = snapshotReconcileInput(candidate);
	if (input === undefined) return { kind: "invalid_snapshot", reason: "invalid autonomy snapshot" };
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
	if (input.canonical.observedStage === "ABORTED") return { kind: "aborted", reason: "human_rejected" };
	if (input.canonical.observedStage === "BLOCKED") return { kind: "blocked", reason: "terminal_blocked" };

	let scheduleSelection: ReadyQueueSelection | undefined;
	if (input.canonical.observedStage === "SCHEDULE") {
		const selected = selectScheduleWork(input);
		if (selected.kind === "invalid_snapshot" || selected.kind === "no_spawn") return selected;
		scheduleSelection = selected;
		if (scheduleSelection.kind === "blocked") {
			return noSpawn(scheduleSelection.blocker, scheduleSelection.blocker);
		}
	}

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
			return awaitHuman("hard_human_gate");
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
		return awaitHuman("retry_budget_exhausted");
	}

	if (input.canonical.proposedStage !== undefined) {
		const transition = evaluateLifecycleTransition({
			from: input.canonical.observedStage,
			to: input.canonical.proposedStage,
			facts: input.canonical.transitionFacts ?? {},
		});
		if (!transition.allowed) {
			if (transition.reason.startsWith("unsafe lifecycle transition")) {
				return { kind: "invalid_snapshot", reason: transition.reason };
			}
			if (input.canonical.observedStage === "HUMAN_DECISION") {
				return awaitHuman("pending_authenticated_decision");
			}
			if (input.canonical.transitionFacts?.correctionRequired === true) {
				return { kind: "await_stage_evidence", stage: input.canonical.observedStage };
			}
			if (input.canonical.observedStage === "VERIFY" && input.canonical.proposedStage === "REVIEW") {
				return noSpawn("not_spawned_verification_blocked", transition.reason);
			}
			if (input.canonical.observedStage === "REVIEW" && input.canonical.proposedStage === "INTEGRATE") {
				return noSpawn("not_spawned_review_blocked", transition.reason);
			}
			return { kind: "await_stage_evidence", stage: input.canonical.observedStage };
		}
		if (input.canonical.observedStage === "SCHEDULE") {
			if (scheduleSelection === undefined
				|| !scheduleSelectionMatchesStage(scheduleSelection, input.canonical.proposedStage)) {
				return { kind: "invalid_snapshot", reason: "lifecycle facts conflict with authoritative work queue" };
			}
		}
		return {
			kind: "transition",
			from: input.canonical.observedStage,
			to: input.canonical.proposedStage,
			reason: transition.reason,
		};
	}

	if (input.canonical.observedStage === "HUMAN_DECISION") {
		return awaitHuman("pending_authenticated_decision");
	}
	if (input.canonical.observedStage !== "SCHEDULE") {
		return { kind: "await_stage_evidence", stage: input.canonical.observedStage };
	}
	const selection = scheduleSelection ?? selectScheduleWork(input);
	if (selection.kind === "invalid_snapshot" || selection.kind === "no_spawn") return selection;
	if (selection.kind === "complete") {
		return { kind: "transition", from: "SCHEDULE", to: "FINAL_VERIFY", reason: "all_tasks_integrated" };
	}
	if (selection.kind === "at_capacity") return { kind: "at_capacity" };
	if (selection.kind === "blocked") return noSpawn(selection.blocker, selection.blocker);
	if (selection.kind === "selected") {
		let selectedIds = selection.itemIds;
		if (!input.canonical.constraints.isolationAvailable) {
			const readerSelection = selectReadyWork(input.canonical.workItems, {
				maxConcurrency: input.canonical.maxConcurrency,
				allowMutating: false,
			});
			if (readerSelection.kind === "selected") selectedIds = readerSelection.itemIds;
		}
		const spawnConstrained = spawnCapabilityDecision(input, selectedIds);
		if (spawnConstrained !== undefined) return spawnConstrained;
		return { kind: "spawn", itemIds: selectedIds };
	}
	return { kind: "at_capacity" };
}
