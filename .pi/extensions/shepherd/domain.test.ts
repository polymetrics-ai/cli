import assert from "node:assert/strict";
import test from "node:test";

import {
	assessLaneEvidence,
	rateDimensions,
	reconcileInterruptedRun,
	selectReadyLanes,
} from "./domain.ts";

const head = "a".repeat(40);
const binding = {
	runId: "run-1",
	generation: 1,
	laneId: "scout",
	candidateHead: head,
	validationNonce: "nonce-1234567890",
	readOnly: true,
	provider: "openai-codex",
	model: "gpt-5.6-sol",
	thinking: "xhigh",
};
const perfectDimensions = {
	correctStage: 1,
	artifactValid: 1,
	gatesRespected: 1,
	realProgress: 1,
	noHallucination: 1,
	noConflict: 1,
};

function evidence(overrides = {}) {
	return {
		...binding,
		summary: "Exact-head read-only inspection completed.",
		dimensions: perfectDimensions,
		observedMutation: false,
		...overrides,
	};
}

test("rates six bounded dimensions with a geometric mean", () => {
	assert.equal(rateDimensions(perfectDimensions), 1);
	assert.ok(Math.abs(rateDimensions({ ...perfectDimensions, realProgress: 0.64 }) - Math.pow(0.64, 1 / 6)) < 1e-12);
	assert.throws(() => rateDimensions({ ...perfectDimensions, noConflict: 1.1 }), /between 0 and 1/);
});

test("accepts fresh exact-bound evidence and diagnoses low-quality evidence", () => {
	assert.deepEqual(assessLaneEvidence(binding, evidence()), {
		decision: "proceed",
		score: 1,
		hardGates: [],
	});

	const result = assessLaneEvidence(
		binding,
		evidence({ dimensions: { ...perfectDimensions, realProgress: 0.01 } }),
	);
	assert.equal(result.decision, "correct");
	assert.equal(result.hardGates.length, 0);
});

test("hard gates override a perfect model-authored score", () => {
	const cases = [
		[evidence({ runId: "stale" }), "run_identity_mismatch"],
		[evidence({ generation: 2 }), "generation_mismatch"],
		[evidence({ laneId: "validator" }), "lane_identity_mismatch"],
		[evidence({ candidateHead: "b".repeat(40) }), "stale_head"],
		[evidence({ validationNonce: "stale-nonce" }), "stale_nonce"],
		[evidence({ model: "gpt-5.5" }), "model_mismatch"],
		[evidence({ thinking: "high" }), "thinking_mismatch"],
		[evidence({ observedMutation: true }), "read_only_violation"],
	];

	for (const [candidate, gate] of cases) {
		const result = assessLaneEvidence(binding, candidate);
		assert.equal(result.decision, "halt");
		assert.ok(result.hardGates.includes(gate), gate);
	}
});

test("selects dependency-ready lanes with bounded concurrency and one mutator", () => {
	const lanes = [
		{ id: "worker-a", mutating: true, dependsOn: [] },
		{ id: "worker-b", mutating: true, dependsOn: [] },
		{ id: "scout", mutating: false, dependsOn: [] },
		{ id: "reviewer", mutating: false, dependsOn: ["worker-a"] },
	];
	const statuses = new Map(lanes.map((lane) => [lane.id, "pending"]));
	assert.deepEqual(
		selectReadyLanes(lanes, statuses, 2).map((lane) => lane.id),
		["worker-a", "scout"],
	);
	statuses.set("worker-a", "succeeded");
	assert.deepEqual(
		selectReadyLanes(lanes, statuses, 2).map((lane) => lane.id),
		["worker-b", "scout"],
	);
});

test("restart reconciliation never trusts persisted running work", () => {
	const run = {
		schemaVersion: 1,
		issue: 471,
		repositoryIdentity: "b".repeat(64),
		worktreeIdentity: "c".repeat(64),
		runId: "run-1",
		generation: 1,
		status: "running",
		candidateHead: head,
		validationNonce: "nonce-1234567890",
		createdAt: "2026-07-21T08:00:00Z",
		updatedAt: "2026-07-21T08:00:00Z",
		lanes: [
			{ id: "scout", role: "scout", mutating: false, dependsOn: [], status: "running" },
			{ id: "validator", role: "validator", mutating: false, dependsOn: [], status: "succeeded" },
		],
	};
	const reconciled = reconcileInterruptedRun(run, "2026-07-21T09:00:00Z");
	assert.equal(reconciled.status, "interrupted");
	assert.equal(reconciled.lanes[0].status, "interrupted");
	assert.equal(reconciled.lanes[1].status, "succeeded");
	assert.equal(reconciled.updatedAt, "2026-07-21T09:00:00Z");
	assert.equal(run.status, "running", "input must not be mutated");
});

test("restart reconciliation interrupts every unfinished lane in an all-pending running checkpoint", () => {
	const run = {
		schemaVersion: 1,
		issue: 471,
		repositoryIdentity: "b".repeat(64),
		worktreeIdentity: "c".repeat(64),
		runId: "run-before-dispatch",
		generation: 1,
		status: "running",
		candidateHead: head,
		validationNonce: "nonce-1234567890",
		createdAt: "2026-07-21T08:00:00Z",
		updatedAt: "2026-07-21T08:00:00Z",
		lanes: [
			{ id: "scout", role: "scout", mutating: false, dependsOn: [], status: "pending" },
			{ id: "validator", role: "validator", mutating: false, dependsOn: [], status: "pending" },
		],
	};

	const reconciled = reconcileInterruptedRun(run, "2026-07-21T09:00:00Z");

	assert.equal(reconciled.status, "interrupted");
	assert.deepEqual(reconciled.lanes.map((lane) => lane.status), ["interrupted", "interrupted"]);
	assert.deepEqual(run.lanes.map((lane) => lane.status), ["pending", "pending"], "input must not be mutated");
});
