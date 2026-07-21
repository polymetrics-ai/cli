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

test("ordinary missing evidence waits, skipped transitions are invalid, and corrections block advancement", () => {
	const missing = input({
		persisted: { stage: "PARENT_PLAN", retryAttempts: 0, correctionRounds: 0 },
		canonical: {
			...input().canonical,
			observedStage: "PARENT_PLAN",
			proposedStage: "ISSUE_CREATE",
			transitionFacts: { parentPlanComplete: false },
		},
	});
	assert.deepEqual(reconcileAutonomy(missing), { kind: "await_stage_evidence", stage: "PARENT_PLAN" });

	const skipped = structuredClone(missing);
	skipped.canonical.proposedStage = "MERGE";
	assert.deepEqual(reconcileAutonomy(skipped), {
		kind: "invalid_snapshot",
		reason: "unsafe lifecycle transition PARENT_PLAN -> MERGE",
	});

	const correction = input({
		persisted: { stage: "VERIFY", retryAttempts: 0, correctionRounds: 0 },
		canonical: {
			...input().canonical,
			observedStage: "VERIFY",
			proposedStage: "REVIEW",
			transitionFacts: { verificationPassed: true, correctionRequired: true },
		},
	});
	assert.deepEqual(reconcileAutonomy(correction), { kind: "await_stage_evidence", stage: "VERIFY" });
});

test("lifecycle advancement cannot contradict the authoritative schedule queue", () => {
	const prematureCompletion = input({ canonical: {
		...input().canonical,
		proposedStage: "FINAL_VERIFY",
		transitionFacts: { allTasksIntegrated: true },
		workItems: [item({ id: "pending" })],
	} });
	assert.deepEqual(reconcileAutonomy(prematureCompletion), {
		kind: "invalid_snapshot",
		reason: "lifecycle facts conflict with authoritative work queue",
	});
});

test("dependency blockers override claimed ready-work lifecycle facts", () => {
	for (const status of ["failed", "blocked"] as const) {
		const dependencyBlocked = input({ canonical: {
			...input().canonical,
			proposedStage: "TASK_PLAN",
			transitionFacts: { readyWorkAvailable: true },
			workItems: [
				item({ id: "prerequisite", status }),
				item({ id: "waiting", dependsOn: ["prerequisite"] }),
			],
		} });
		assert.deepEqual(reconcileAutonomy(dependencyBlocked), {
			kind: "no_spawn",
			blocker: "not_spawned_dependency_blocked",
			reason: "not_spawned_dependency_blocked",
		}, status);
	}
});

test("transient failures retry, then correct, while real decision points enter resumable human wait", () => {
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
		kind: "await_human_decision",
		blocker: "not_spawned_human_gate",
		reason: "hard_human_gate",
	});

	const exhausted = input({
		persisted: { stage: "VERIFY", retryAttempts: 2, correctionRounds: 1 },
		canonical: { ...input().canonical, observedStage: "VERIFY" },
		failure: "transient_verification",
	});
	assert.deepEqual(reconcileAutonomy(exhausted), {
		kind: "await_human_decision",
		blocker: "not_spawned_human_gate",
		reason: "retry_budget_exhausted",
	});
});

test("authenticated human approval resumes work and rejection reaches terminal abort", () => {
	const human = input({
		persisted: { stage: "HUMAN_DECISION", retryAttempts: 0, correctionRounds: 0 },
		canonical: { ...input().canonical, observedStage: "HUMAN_DECISION" },
	});
	assert.deepEqual(reconcileAutonomy(human), {
		kind: "await_human_decision",
		blocker: "not_spawned_human_gate",
		reason: "pending_authenticated_decision",
	});

	const approved = structuredClone(human);
	approved.canonical.proposedStage = "MERGE";
	approved.canonical.transitionFacts = {
		humanDecision: "approve_merge",
		humanDecisionAuthenticated: true,
		exactHeadRevalidated: true,
	};
	assert.equal(reconcileAutonomy(approved).kind, "transition");

	const rejected = structuredClone(human);
	rejected.canonical.proposedStage = "ABORTED" as "HUMAN_DECISION";
	rejected.canonical.transitionFacts = { humanDecision: "reject", humanDecisionAuthenticated: true };
	assert.deepEqual(reconcileAutonomy(rejected), {
		kind: "transition",
		from: "HUMAN_DECISION",
		to: "ABORTED",
		reason: "transition_allowed",
	});

	const aborted = input({
		persisted: { stage: "ABORTED" as "BLOCKED", retryAttempts: 0, correctionRounds: 0 },
		canonical: { ...input().canonical, observedStage: "ABORTED" as "BLOCKED" },
	});
	assert.deepEqual(reconcileAutonomy(aborted), { kind: "aborted", reason: "human_rejected" });
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

test("missing isolation removes only selected mutators and preserves selected readers", () => {
	const candidate = input({ canonical: {
		...input().canonical,
		maxConcurrency: 4,
		constraints: { ...input().canonical.constraints, isolationAvailable: false },
		workItems: [
			item({ id: "running", status: "running", writeScopes: ["src/shared"] }),
			item({ id: "waiting-writer", writeScopes: ["src/shared/file.ts"] }),
			item({ id: "safe-reader", access: "read_only", writeScopes: [] }),
		],
	} });
	assert.deepEqual(reconcileAutonomy(candidate), { kind: "spawn", itemIds: ["safe-reader"] });
});

test("dependency and collision blockers take precedence over spawn capability blockers", () => {
	const unavailable = {
		...input().canonical.constraints,
		runtimeCapabilityAvailable: false,
		isolationAvailable: false,
	};
	assert.deepEqual(reconcileAutonomy(input({ canonical: {
		...input().canonical,
		constraints: unavailable,
		workItems: [item({ id: "failed", status: "failed" }), item({ id: "waiting", dependsOn: ["failed"] })],
	} })), {
		kind: "no_spawn",
		blocker: "not_spawned_dependency_blocked",
		reason: "not_spawned_dependency_blocked",
	});
	assert.deepEqual(reconcileAutonomy(input({ canonical: {
		...input().canonical,
		constraints: unavailable,
		workItems: [
			item({ id: "running", status: "running", writeScopes: ["src/shared"] }),
			item({ id: "waiting", writeScopes: ["src/shared/file.ts"] }),
		],
	} })), {
		kind: "no_spawn",
		blocker: "not_spawned_write_scope_collision",
		reason: "not_spawned_write_scope_collision",
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
		if (expected === "not_spawned_human_gate") {
			assert.equal(decision.kind, "await_human_decision", expected);
			if (decision.kind === "await_human_decision") assert.equal(decision.blocker, expected);
			continue;
		}
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

test("completion is not masked by capabilities needed only for future spawns", () => {
	const candidate = input({
		persisted: { stage: "COMPLETE", retryAttempts: 0, correctionRounds: 0 },
		canonical: {
			...input().canonical,
			observedStage: "COMPLETE",
			workItems: [item({ id: "done", status: "succeeded" })],
			constraints: {
				...input().canonical.constraints,
				runtimeCapabilityAvailable: false,
				isolationAvailable: false,
			},
		},
	});
	assert.deepEqual(reconcileAutonomy(candidate), { kind: "complete" });
});

test("canonical BLOCKED state returns a terminal reconciliation result", () => {
	const candidate = input({
		persisted: { stage: "BLOCKED", retryAttempts: 0, correctionRounds: 0 },
		canonical: { ...input().canonical, observedStage: "BLOCKED" },
	});
	assert.deepEqual(reconcileAutonomy(candidate), { kind: "blocked", reason: "terminal_blocked" });
});

test("runtime-invalid stages and concurrency facts fail closed without throwing", () => {
	const invalidStage = input({
		canonical: { ...input().canonical, observedStage: "UNKNOWN" as "SCHEDULE" },
	});
	assert.deepEqual(reconcileAutonomy(invalidStage), {
		kind: "invalid_snapshot",
		reason: "invalid autonomy snapshot",
	});

	const invalidConcurrency = input({
		canonical: { ...input().canonical, maxConcurrency: 0 },
	});
	assert.deepEqual(reconcileAutonomy(invalidConcurrency), {
		kind: "no_spawn",
		blocker: "not_spawned_runtime_capability_missing",
		reason: "invalid concurrency policy",
	});
});

test("complete exact runtime DTO validation is BigInt-safe and returns one typed fail-closed result", () => {
	const expected = { kind: "invalid_snapshot", reason: "invalid autonomy snapshot" } as const;
	const hostile: unknown[] = [
		null,
		{},
		{ ...input(), unexpected: true },
		{ ...input(), persisted: { ...input().persisted, retryAttempts: 1n } },
		{ ...input(), budget: { ...input().budget, maxRetries: 1n } },
		{ ...input(), canonical: { ...input().canonical, maxConcurrency: 1n } },
		{ ...input(), canonical: {
			...input().canonical,
			constraints: { ...input().canonical.constraints, reviewBlocked: "false" },
		} },
		{ ...input(), canonical: {
			...input().canonical,
			workItems: [{ ...item({ id: "worker" }), writeScopes: [1n] }],
		} },
	];
	for (const candidate of hostile) {
		assert.doesNotThrow(() => reconcileAutonomy(candidate as ReconcileInput));
		assert.deepEqual(reconcileAutonomy(candidate as ReconcileInput), expected);
	}
});

test("caller-owned Proxy iterators are rejected before they can mutate or escape validation", () => {
	const expected = { kind: "invalid_snapshot", reason: "invalid autonomy snapshot" } as const;
	const mutableItems = [item({ id: "worker" })];
	const changingItems = new Proxy(mutableItems, {
		get(target, property, receiver) {
			if (property === Symbol.iterator) {
				return function* hostileIterator() {
					target[0].status = "succeeded";
					yield target[0];
				};
			}
			return Reflect.get(target, property, receiver);
		},
	});
	const changing = input({ canonical: {
		...input().canonical,
		workItems: changingItems,
	} });
	assert.deepEqual(reconcileAutonomy(changing), expected);
	assert.equal(mutableItems[0].status, "pending", "caller iterator must never run");

	const throwingItems = new Proxy([item({ id: "worker" })], {
		get(target, property, receiver) {
			if (property === Symbol.iterator) return () => { throw new Error("hostile iterator invoked"); };
			return Reflect.get(target, property, receiver);
		},
	});
	const throwing = input({ canonical: { ...input().canonical, workItems: throwingItems } });
	assert.doesNotThrow(() => reconcileAutonomy(throwing));
	assert.deepEqual(reconcileAutonomy(throwing), expected);
});

test("accessor-bearing DTOs are rejected without executing caller code", () => {
	const expected = { kind: "invalid_snapshot", reason: "invalid autonomy snapshot" } as const;
	let rootGetterCalls = 0;
	const rootAccessor = input();
	const pendingCanonical = rootAccessor.canonical;
	const completeCanonical = {
		...pendingCanonical,
		workItems: [item({ id: "worker", status: "succeeded" })],
	};
	Object.defineProperty(rootAccessor, "canonical", {
		configurable: true,
		enumerable: true,
		get() {
			rootGetterCalls += 1;
			return rootGetterCalls % 2 === 1 ? pendingCanonical : completeCanonical;
		},
	});

	let nestedGetterCalls = 0;
	const nestedAccessor = input();
	Object.defineProperty(nestedAccessor.canonical.workItems[0], "status", {
		configurable: true,
		enumerable: true,
		get() {
			nestedGetterCalls += 1;
			return nestedGetterCalls % 2 === 1 ? "pending" : "succeeded";
		},
	});

	let arrayGetterCalls = 0;
	const arrayAccessor = input();
	const pendingItem = arrayAccessor.canonical.workItems[0];
	const completeItem = item({ id: "worker", status: "succeeded" });
	Object.defineProperty(arrayAccessor.canonical.workItems, "0", {
		configurable: true,
		enumerable: true,
		get() {
			arrayGetterCalls += 1;
			return arrayGetterCalls % 2 === 1 ? pendingItem : completeItem;
		},
	});

	const decisions = [rootAccessor, nestedAccessor, arrayAccessor].map((candidate) => [
		reconcileAutonomy(candidate),
		reconcileAutonomy(candidate),
	]);
	assert.deepEqual(
		{ rootGetterCalls, nestedGetterCalls, arrayGetterCalls },
		{ rootGetterCalls: 0, nestedGetterCalls: 0, arrayGetterCalls: 0 },
		"descriptor validation must not execute caller accessors",
	);
	for (const pair of decisions) assert.deepEqual(pair, [expected, expected]);
});

test("reconciliation is pure and idempotent for the same persisted and canonical snapshot", () => {
	const candidate = input();
	const before = structuredClone(candidate);
	const first = reconcileAutonomy(candidate);
	const second = reconcileAutonomy(candidate);
	assert.deepEqual(first, second);
	assert.deepEqual(candidate, before);
});
