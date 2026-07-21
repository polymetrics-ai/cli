export const PARENT_LIFECYCLE_STAGES = [
	"INTAKE",
	"RESEARCH",
	"PARENT_PLAN",
	"ISSUE_CREATE",
	"PARENT_SETUP",
	"SCHEDULE",
	"TASK_PLAN",
	"SUB_BRANCH",
	"EXECUTE",
	"SUB_PR_OPEN",
	"VERIFY",
	"REVIEW",
	"CORRECT",
	"INTEGRATE",
	"FINAL_VERIFY",
	"HUMAN_DECISION",
	"MERGE",
	"COMPLETE",
	"ABORTED",
	"BLOCKED",
] as const;

export type ParentLifecycleStage = typeof PARENT_LIFECYCLE_STAGES[number];

export const REPOSITORY_BLOCKERS = [
	"not_spawned_dependency_blocked",
	"not_spawned_write_scope_collision",
	"not_spawned_human_gate",
	"not_spawned_isolation_missing",
	"not_spawned_runtime_capability_missing",
	"not_spawned_review_blocked",
	"not_spawned_verification_blocked",
] as const;

export type RepositoryBlocker = typeof REPOSITORY_BLOCKERS[number];

export interface LifecycleFacts {
	researchRequired?: boolean;
	researchComplete?: boolean;
	parentPlanComplete?: boolean;
	issuesCreated?: boolean;
	parentSetupComplete?: boolean;
	readyWorkAvailable?: boolean;
	allTasksIntegrated?: boolean;
	taskPlanComplete?: boolean;
	isolatedBranchReady?: boolean;
	executionCheckpointed?: boolean;
	subPrOpen?: boolean;
	verificationPassed?: boolean;
	reviewClean?: boolean;
	correctionRequired?: boolean;
	correctionCheckpointed?: boolean;
	integrationConfirmed?: boolean;
	finalVerificationPassed?: boolean;
	humanDecision?: "pending" | "approve_merge" | "reject";
	humanDecisionAuthenticated?: boolean;
	exactHeadRevalidated?: boolean;
	mergeConfirmed?: boolean;
	hardHumanGate?: boolean;
}

export interface LifecycleTransitionRequest {
	from: ParentLifecycleStage;
	to: ParentLifecycleStage;
	facts: LifecycleFacts | Record<string, unknown>;
}

export interface LifecycleTransitionDecision {
	allowed: boolean;
	reason: string;
}

interface TransitionRule {
	to: ParentLifecycleStage;
	guard: (facts: LifecycleFacts) => boolean;
	failureReason: string;
}

const transitionRules: Readonly<Record<Exclude<ParentLifecycleStage, "COMPLETE" | "ABORTED" | "BLOCKED">, readonly TransitionRule[]>> = {
	INTAKE: [
		{ to: "RESEARCH", guard: (facts) => facts.researchRequired === true, failureReason: "research is not required" },
		{ to: "PARENT_PLAN", guard: (facts) => facts.researchRequired === false, failureReason: "required research cannot be skipped" },
	],
	RESEARCH: [
		{ to: "PARENT_PLAN", guard: (facts) => facts.researchComplete === true, failureReason: "research evidence is incomplete" },
	],
	PARENT_PLAN: [
		{ to: "ISSUE_CREATE", guard: (facts) => facts.parentPlanComplete === true, failureReason: "parent plan evidence is missing" },
	],
	ISSUE_CREATE: [
		{ to: "PARENT_SETUP", guard: (facts) => facts.issuesCreated === true, failureReason: "issue creation evidence is missing" },
	],
	PARENT_SETUP: [
		{ to: "SCHEDULE", guard: (facts) => facts.parentSetupComplete === true, failureReason: "parent setup evidence is missing" },
	],
	SCHEDULE: [
		{ to: "TASK_PLAN", guard: (facts) => facts.readyWorkAvailable === true, failureReason: "ready work evidence is missing" },
		{ to: "FINAL_VERIFY", guard: (facts) => facts.allTasksIntegrated === true, failureReason: "task integration evidence is incomplete" },
	],
	TASK_PLAN: [
		{ to: "SUB_BRANCH", guard: (facts) => facts.taskPlanComplete === true, failureReason: "task plan evidence is missing" },
	],
	SUB_BRANCH: [
		{ to: "EXECUTE", guard: (facts) => facts.isolatedBranchReady === true, failureReason: "isolated branch evidence is missing" },
	],
	EXECUTE: [
		{ to: "SUB_PR_OPEN", guard: (facts) => facts.executionCheckpointed === true, failureReason: "execution checkpoint evidence is missing" },
	],
	SUB_PR_OPEN: [
		{ to: "VERIFY", guard: (facts) => facts.subPrOpen === true, failureReason: "sub-PR evidence is missing" },
	],
	VERIFY: [
		{
			to: "REVIEW",
			guard: (facts) => facts.verificationPassed === true && facts.correctionRequired !== true,
			failureReason: "verification evidence is not passing",
		},
		{ to: "CORRECT", guard: (facts) => facts.correctionRequired === true, failureReason: "correction evidence is missing" },
	],
	REVIEW: [
		{
			to: "INTEGRATE",
			guard: (facts) => facts.reviewClean === true && facts.correctionRequired !== true,
			failureReason: "review evidence is not clean",
		},
		{ to: "CORRECT", guard: (facts) => facts.correctionRequired === true, failureReason: "correction evidence is missing" },
	],
	CORRECT: [
		{ to: "VERIFY", guard: (facts) => facts.correctionCheckpointed === true, failureReason: "correction checkpoint evidence is missing" },
	],
	INTEGRATE: [
		{ to: "SCHEDULE", guard: (facts) => facts.integrationConfirmed === true, failureReason: "integration evidence is missing" },
	],
	FINAL_VERIFY: [
		{ to: "HUMAN_DECISION", guard: (facts) => facts.finalVerificationPassed === true, failureReason: "final verification evidence is not passing" },
	],
	HUMAN_DECISION: [
		{
			to: "MERGE",
			guard: (facts) => facts.humanDecision === "approve_merge"
				&& facts.humanDecisionAuthenticated === true
				&& facts.exactHeadRevalidated === true,
			failureReason: "an authenticated approve-merge decision and exact-head revalidation are required",
		},
		{
			to: "ABORTED",
			guard: (facts) => facts.humanDecision === "reject" && facts.humanDecisionAuthenticated === true,
			failureReason: "an authenticated rejection decision is required",
		},
	],
	MERGE: [
		{ to: "COMPLETE", guard: (facts) => facts.mergeConfirmed === true, failureReason: "merge confirmation evidence is missing" },
	],
};

export function isParentLifecycleStage(value: unknown): value is ParentLifecycleStage {
	return typeof value === "string" && (PARENT_LIFECYCLE_STAGES as readonly string[]).includes(value);
}

export function evaluateLifecycleTransition(request: LifecycleTransitionRequest): LifecycleTransitionDecision {
	const { from, to } = request;
	if (!isParentLifecycleStage(from) || !isParentLifecycleStage(to)
		|| typeof request.facts !== "object" || request.facts === null || Array.isArray(request.facts)) {
		return { allowed: false, reason: "invalid lifecycle stage" };
	}
	const facts = request.facts as LifecycleFacts;
	if (from === "COMPLETE" || from === "ABORTED" || from === "BLOCKED") {
		return { allowed: false, reason: `${from} is a terminal lifecycle stage` };
	}
	const rule = transitionRules[from].find((candidate) => candidate.to === to);
	if (rule === undefined) return { allowed: false, reason: `unsafe lifecycle transition ${from} -> ${to}` };
	return rule.guard(facts)
		? { allowed: true, reason: "transition_allowed" }
		: { allowed: false, reason: rule.failureReason };
}

export type FailureClass = "transient_verification" | "transient_review" | "hard_human_gate";
export type FailurePolicyAction = "retry" | "correct" | "wait_for_human";

export interface FailurePolicyRequest {
	failure: FailureClass;
	retryAttempts: number;
	maxRetries: number;
	correctionRounds: number;
	maxCorrectionRounds: number;
}

export interface FailurePolicyDecision {
	action: FailurePolicyAction;
	remainingRetries: number;
	remainingCorrections: number;
}

function assertNonNegativeSafeInteger(value: number, description: string): void {
	if (!Number.isSafeInteger(value) || value < 0) {
		throw new RangeError(`${description} must be a non-negative safe integer`);
	}
}

export function decideFailurePolicy(request: FailurePolicyRequest): FailurePolicyDecision {
	if (request.failure !== "transient_verification"
		&& request.failure !== "transient_review"
		&& request.failure !== "hard_human_gate") {
		throw new TypeError("invalid failure class");
	}
	assertNonNegativeSafeInteger(request.retryAttempts, "retry attempts");
	assertNonNegativeSafeInteger(request.maxRetries, "maximum retries");
	assertNonNegativeSafeInteger(request.correctionRounds, "correction rounds");
	assertNonNegativeSafeInteger(request.maxCorrectionRounds, "maximum correction rounds");
	const remainingRetries = Math.max(0, request.maxRetries - request.retryAttempts);
	const remainingCorrections = Math.max(0, request.maxCorrectionRounds - request.correctionRounds);

	if (request.failure === "hard_human_gate") {
		return { action: "wait_for_human", remainingRetries, remainingCorrections };
	}
	if (remainingRetries > 0) {
		return { action: "retry", remainingRetries: remainingRetries - 1, remainingCorrections };
	}
	if (remainingCorrections > 0) {
		return { action: "correct", remainingRetries: 0, remainingCorrections: remainingCorrections - 1 };
	}
	return { action: "wait_for_human", remainingRetries: 0, remainingCorrections: 0 };
}
