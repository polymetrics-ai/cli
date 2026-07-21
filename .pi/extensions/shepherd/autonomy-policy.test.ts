import assert from "node:assert/strict";
import test from "node:test";

import {
	PARENT_LIFECYCLE_STAGES,
	REPOSITORY_BLOCKERS,
	decideFailurePolicy,
	evaluateLifecycleTransition,
	type ParentLifecycleStage,
} from "./autonomy-policy.ts";

test("parent lifecycle vocabulary covers intake through complete including correction and human wait", () => {
	assert.deepEqual(PARENT_LIFECYCLE_STAGES, [
		"INTAKE", "RESEARCH", "PARENT_PLAN", "ISSUE_CREATE", "PARENT_SETUP", "SCHEDULE",
		"TASK_PLAN", "SUB_BRANCH", "EXECUTE", "SUB_PR_OPEN", "VERIFY", "REVIEW", "CORRECT",
		"INTEGRATE", "FINAL_VERIFY", "HUMAN_DECISION", "MERGE", "COMPLETE", "BLOCKED",
	]);
	assert.equal(new Set(PARENT_LIFECYCLE_STAGES).size, PARENT_LIFECYCLE_STAGES.length);
	assert.equal(REPOSITORY_BLOCKERS.length, 7);
});

test("allows every safe lifecycle edge with its required canonical fact", () => {
	const cases: Array<[ParentLifecycleStage, ParentLifecycleStage, Record<string, unknown>]> = [
		["INTAKE", "RESEARCH", { researchRequired: true }],
		["INTAKE", "PARENT_PLAN", { researchRequired: false }],
		["RESEARCH", "PARENT_PLAN", { researchComplete: true }],
		["PARENT_PLAN", "ISSUE_CREATE", { parentPlanComplete: true }],
		["ISSUE_CREATE", "PARENT_SETUP", { issuesCreated: true }],
		["PARENT_SETUP", "SCHEDULE", { parentSetupComplete: true }],
		["SCHEDULE", "TASK_PLAN", { readyWorkAvailable: true }],
		["SCHEDULE", "FINAL_VERIFY", { allTasksIntegrated: true }],
		["TASK_PLAN", "SUB_BRANCH", { taskPlanComplete: true }],
		["SUB_BRANCH", "EXECUTE", { isolatedBranchReady: true }],
		["EXECUTE", "SUB_PR_OPEN", { executionCheckpointed: true }],
		["SUB_PR_OPEN", "VERIFY", { subPrOpen: true }],
		["VERIFY", "REVIEW", { verificationPassed: true }],
		["VERIFY", "CORRECT", { correctionRequired: true }],
		["REVIEW", "INTEGRATE", { reviewClean: true }],
		["REVIEW", "CORRECT", { correctionRequired: true }],
		["CORRECT", "VERIFY", { correctionCheckpointed: true }],
		["INTEGRATE", "SCHEDULE", { integrationConfirmed: true }],
		["FINAL_VERIFY", "HUMAN_DECISION", { finalVerificationPassed: true }],
		["HUMAN_DECISION", "MERGE", { humanDecision: "approve_merge", exactHeadRevalidated: true }],
		["MERGE", "COMPLETE", { mergeConfirmed: true }],
	];

	for (const [from, to, facts] of cases) {
		assert.deepEqual(evaluateLifecycleTransition({ from, to, facts }), {
			allowed: true,
			reason: "transition_allowed",
		}, `${from} -> ${to}`);
	}
});

test("unsafe, skipped, stale-head, and under-evidenced transitions fail closed", () => {
	const cases: Array<[ParentLifecycleStage, ParentLifecycleStage, Record<string, unknown>]> = [
		["INTAKE", "PARENT_PLAN", { researchRequired: true }],
		["RESEARCH", "PARENT_PLAN", { researchComplete: false }],
		["VERIFY", "REVIEW", { verificationPassed: false }],
		["REVIEW", "INTEGRATE", { reviewClean: false }],
		["SCHEDULE", "FINAL_VERIFY", { allTasksIntegrated: false }],
		["FINAL_VERIFY", "HUMAN_DECISION", { finalVerificationPassed: false }],
		["HUMAN_DECISION", "MERGE", { humanDecision: "approve_merge", exactHeadRevalidated: false }],
		["HUMAN_DECISION", "MERGE", { humanDecision: "pending", exactHeadRevalidated: true }],
		["MERGE", "COMPLETE", { mergeConfirmed: false }],
		["INTAKE", "COMPLETE", {}],
	];

	for (const [from, to, facts] of cases) {
		const decision = evaluateLifecycleTransition({ from, to, facts });
		assert.equal(decision.allowed, false, `${from} -> ${to}`);
		assert.match(decision.reason, /unsafe|missing|research|verification|review|integrat|decision|head|merge/i);
	}
});

test("only an explicit hard gate can enter BLOCKED and BLOCKED is terminal", () => {
	assert.equal(evaluateLifecycleTransition({ from: "EXECUTE", to: "BLOCKED", facts: {} }).allowed, false);
	assert.deepEqual(
		evaluateLifecycleTransition({ from: "EXECUTE", to: "BLOCKED", facts: { hardHumanGate: true } }),
		{ allowed: true, reason: "transition_allowed" },
	);
	assert.equal(
		evaluateLifecycleTransition({ from: "BLOCKED", to: "SCHEDULE", facts: { hardHumanGate: false } }).allowed,
		false,
	);
});

test("transient verification and review failures consume bounded retry then correction budgets", () => {
	for (const failure of ["transient_verification", "transient_review"] as const) {
		assert.deepEqual(decideFailurePolicy({
			failure,
			retryAttempts: 0,
			maxRetries: 2,
			correctionRounds: 0,
			maxCorrectionRounds: 1,
		}), { action: "retry", remainingRetries: 1, remainingCorrections: 1 });

		assert.deepEqual(decideFailurePolicy({
			failure,
			retryAttempts: 2,
			maxRetries: 2,
			correctionRounds: 0,
			maxCorrectionRounds: 1,
		}), { action: "correct", remainingRetries: 0, remainingCorrections: 0 });

		assert.deepEqual(decideFailurePolicy({
			failure,
			retryAttempts: 2,
			maxRetries: 2,
			correctionRounds: 1,
			maxCorrectionRounds: 1,
		}), { action: "wait_for_human", remainingRetries: 0, remainingCorrections: 0 });
	}
});

test("hard human gates never consume or bypass retry and correction budgets", () => {
	assert.deepEqual(decideFailurePolicy({
		failure: "hard_human_gate",
		retryAttempts: 0,
		maxRetries: 4,
		correctionRounds: 0,
		maxCorrectionRounds: 4,
	}), { action: "wait_for_human", remainingRetries: 4, remainingCorrections: 4 });
	assert.throws(() => decideFailurePolicy({
		failure: "transient_review",
		retryAttempts: -1,
		maxRetries: 1,
		correctionRounds: 0,
		maxCorrectionRounds: 1,
	}), /non-negative safe integer/);
});
