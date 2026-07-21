import assert from "node:assert/strict";
import test from "node:test";

import { REPOSITORY_BLOCKERS, type RepositoryBlocker } from "./autonomy-policy.ts";
import type { DependencyWorkItem } from "./dependency-graph.ts";
import { reconcileAutonomy, type ReconcileInput } from "./reconciler.ts";

function item(overrides: Partial<DependencyWorkItem> & Pick<DependencyWorkItem, "id">): DependencyWorkItem {
	return {
		id: overrides.id,
		dependsOn: [],
		status: "pending",
		access: "mutating",
		writeScopes: [`src/${overrides.id}`],
		...overrides,
	};
}

function input(overrides: Partial<ReconcileInput> = {}): ReconcileInput {
	return {
		persisted: {
			stage: "SCHEDULE",
			retryAttempts: 0,
			correctionRounds: 0,
		},
		canonical: {
			observedStage: "SCHEDULE",
			workItems: [item({ id: "worker" })],
			maxConcurrency: 2,
			constraints: {
				runtimeCapabilityAvailable: true,
				isolationAvailable: true,
				hardHumanGate: false,
				verificationBlocked: false,
				reviewBlocked: false,
			},
		},
		budget: { maxRetries: 2, maxCorrectionRounds: 1 },
		...overrides,
	};
}

test("canonical stage drift is reconciled before any action or spawn", () => {
	const candidate = input({
		persisted: { stage: "REVIEW", retryAttempts: 0, correctionRounds: 0 },
		canonical: { ...input().canonical, observedStage: "VERIFY" },
	});
	assert.deepEqual(reconcileAutonomy(candidate), {
		kind: "reconcile_stage",
		stage: "VERIFY",
		reason: "canonical_stage_differs",
	});
});

test("safe proposed transitions advance and unsafe transitions fail closed", () => {
	const safe = input({
		persisted: { stage: "VERIFY", retryAttempts: 0, correctionRounds: 0 },
		canonical: {
			...input().canonical,
			observedStage: "VERIFY",
			proposedStage: "REVIEW",
			transitionFacts: { verificationPassed: true },
		},
	});
	assert.deepEqual(reconcileAutonomy(safe), {
		kind: "transition",
		from: "VERIFY",
		to: "REVIEW",
		reason: "transition_allowed",
	});

	const unsafe = structuredClone(safe);
	unsafe.canonical.transitionFacts = { verificationPassed: false };
	assert.deepEqual(reconcileAutonomy(unsafe), {
		kind: "no_spawn",
		blocker: "not_spawned_verification_blocked",
		reason: "verification evidence is not passing",
	});
});

test("transient failures retry, then correct, while hard gates wait for a human", () => {
	const retry = input({ failure: "transient_verification" });
	assert.deepEqual(reconcileAutonomy(retry), {
		kind: "retry",
		failure: "transient_verification",
		nextRetryAttempts: 1,
		remainingRetries: 1,
	});

	const correction = input({
		persisted: { stage: "VERIFY", retryAttempts: 2, correctionRounds: 0 },
		canonical: { ...input().canonical, observedStage: "VERIFY" },
		failure: "transient_verification",
	});
	assert.deepEqual(reconcileAutonomy(correction), {
		kind: "correct",
		failure: "transient_verification",
		nextCorrectionRounds: 1,
		remainingCorrections: 0,
	});

	assert.deepEqual(reconcileAutonomy(input({ failure: "hard_human_gate" })), {
		kind: "no_spawn",
		blocker: "not_spawned_human_gate",
		reason: "hard human gate requires an authenticated decision",
	});
});

test("dependency-ready work produces one deterministic bounded spawn decision", () => {
	const candidate = input({
		canonical: {
			...input().canonical,
			maxConcurrency: 2,
			workItems: [
				item({ id: "central", writeScopes: ["src"] }),
				item({ id: "leaf-a", writeScopes: ["src/a"] }),
				item({ id: "leaf-b", writeScopes: ["src/b"] }),
			],
		},
	});
	assert.deepEqual(reconcileAutonomy(candidate), {
		kind: "spawn",
		itemIds: ["leaf-a", "leaf-b"],
	});
});

test("each repository contract condition emits exactly one no-spawn blocker category", () => {
	const cases: Array<[RepositoryBlocker, ReconcileInput]> = [
		["not_spawned_human_gate", input({ canonical: {
			...input().canonical,
			constraints: { ...input().canonical.constraints, hardHumanGate: true },
		} })],
		["not_spawned_runtime_capability_missing", input({ canonical: {
			...input().canonical,
			constraints: { ...input().canonical.constraints, runtimeCapabilityAvailable: false },
		} })],
		["not_spawned_isolation_missing", input({ canonical: {
			...input().canonical,
			constraints: { ...input().canonical.constraints, isolationAvailable: false },
		} })],
		["not_spawned_verification_blocked", input({ canonical: {
			...input().canonical,
			constraints: { ...input().canonical.constraints, verificationBlocked: true },
		} })],
		["not_spawned_review_blocked", input({ canonical: {
			...input().canonical,
			constraints: { ...input().canonical.constraints, reviewBlocked: true },
		} })],
		["not_spawned_dependency_blocked", input({ canonical: {
			...input().canonical,
			workItems: [item({ id: "failed", status: "failed" }), item({ id: "waiting", dependsOn: ["failed"] })],
		} })],
		["not_spawned_write_scope_collision", input({ canonical: {
			...input().canonical,
			workItems: [
				item({ id: "running", status: "running", writeScopes: ["src/shared"] }),
				item({ id: "waiting", writeScopes: ["src/shared/file.ts"] }),
			],
		} })],
	];

	assert.deepEqual(cases.map(([blocker]) => blocker).sort(), [...REPOSITORY_BLOCKERS].sort());
	for (const [expected, candidate] of cases) {
		const decision = reconcileAutonomy(candidate);
		assert.equal(decision.kind, "no_spawn", expected);
		if (decision.kind !== "no_spawn") continue;
		assert.equal(decision.blocker, expected);
		assert.equal(REPOSITORY_BLOCKERS.filter((blocker) => blocker === decision.blocker).length, 1);
	}
});

test("invalid graph facts fail closed into their single corresponding repository blocker", () => {
	const unknown = input({ canonical: {
		...input().canonical,
		workItems: [item({ id: "worker", dependsOn: ["missing"] })],
	} });
	assert.deepEqual(reconcileAutonomy(unknown), {
		kind: "no_spawn",
		blocker: "not_spawned_dependency_blocked",
		reason: "invalid dependency graph: unknown_dependency",
	});

	const ambiguous = input({ canonical: {
		...input().canonical,
		workItems: [item({ id: "worker", writeScopes: ["."] })],
	} });
	assert.deepEqual(reconcileAutonomy(ambiguous), {
		kind: "no_spawn",
		blocker: "not_spawned_write_scope_collision",
		reason: "invalid dependency graph: ambiguous_scope",
	});
});

test("complete and at-capacity snapshots are distinct from no-spawn blockers", () => {
	assert.deepEqual(reconcileAutonomy(input({ canonical: {
		...input().canonical,
		workItems: [item({ id: "done", status: "succeeded" })],
	} })), {
		kind: "transition",
		from: "SCHEDULE",
		to: "FINAL_VERIFY",
		reason: "all_tasks_integrated",
	});
	assert.deepEqual(reconcileAutonomy(input({ canonical: {
		...input().canonical,
		maxConcurrency: 1,
		workItems: [item({ id: "running", status: "running" })],
	} })), { kind: "at_capacity" });
});

test("reconciliation is pure and idempotent for the same persisted and canonical snapshot", () => {
	const candidate = input();
	const before = structuredClone(candidate);
	const first = reconcileAutonomy(candidate);
	const second = reconcileAutonomy(candidate);
	assert.deepEqual(first, second);
	assert.deepEqual(candidate, before);
});
