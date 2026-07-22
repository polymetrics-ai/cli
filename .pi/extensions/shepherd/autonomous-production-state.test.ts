import assert from "node:assert/strict";
import { mkdtemp, readFile, readdir, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import type { ProductionParentPlanDocument } from "./autonomous-production-contract.ts";
import {
	ProductionFileStateStore,
	advanceProductionGeneration,
	createProductionAutonomousState,
	evolveProductionState,
	refreshProductionChildOwnership,
	waitForProductionChildIntervention,
	type ProductionStateFence,
} from "./autonomous-production-state.ts";

function plan(): ProductionParentPlanDocument {
	return {
		schemaVersion: 2,
		planId: "plan-479",
		parentIssue: 479,
		repository: "polymetrics/polymetrics",
		title: "Production Shepherd",
		objective: "Exercise the complete production path",
		parentBranch: "feat/471-pi-agent-session-shepherd",
		parentBaseBranch: "main",
		actorAllowlist: ["maintainer"],
		decisionExpiresAt: "2026-08-01T00:00:00.000Z",
		children: [{
			id: "state",
			issue: 501,
			title: "Durable state",
			task: "Implement durable state without persisting this task text",
			slug: "durable-state",
			dependsOn: [],
			access: "mutating",
			writeScopes: [".pi/extensions/shepherd"],
			requiredSkills: ["javascript-testing-patterns"],
			verification: [{
				id: "state-tests",
				executable: "node",
				args: ["--test", ".pi/extensions/shepherd/autonomous-production-state.test.ts"],
				cwd: ".",
				timeoutMs: 30_000,
				maxOutputBytes: 1_048_576,
			}],
			humanGates: [],
			maxAttempts: 2,
			maxCorrections: 1,
		}],
	};
}

function fence(state: ReturnType<typeof createProductionAutonomousState>): ProductionStateFence {
	return { issue: state.parentIssue, revision: state.revision, generation: state.generation, runId: state.runId };
}

test("persists a canonical plan binding, budgets, checkpoint truth, and no task/prompt material", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-state-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const store = new ProductionFileStateStore(root);
	const initial = createProductionAutonomousState(plan(), {
		runId: "run-479-1",
		now: new Date("2026-07-22T10:00:00.000Z"),
	});
	await store.create(initial);

	const loaded = await store.load(479);
	assert.deepEqual(loaded, initial);
	assert.equal(loaded?.children[0].attempts, 0);
	assert.equal(loaded?.children[0].corrections, 0);
	assert.equal(loaded?.children[0].maxAttempts, 2);
	assert.equal(loaded?.children[0].maxCorrections, 1);
	assert.equal("task" in (loaded?.children[0] ?? {}), false);
	const serialized = await readFile(join(root, "production-issue-479.json"), "utf8");
	assert.equal(serialized.includes("Implement durable state"), false);
	assert.equal(serialized.includes("objective"), false);

	const executing = evolveProductionState(initial, fence(initial), (draft) => {
		const child = draft.children[0];
		child.status = "running";
		child.stage = "implementation";
		child.attempts = 1;
		child.checkpoint = { summary: "implementation accepted", effectKey: "effect-implementation" };
	}, new Date("2026-07-22T10:01:00.000Z"));
	await store.compareAndSwap(fence(initial), executing);
	assert.equal((await store.load(479))?.children[0].checkpoint?.summary, "implementation accepted");
	assert.deepEqual(await readdir(root), ["production-issue-479.json"]);
});

test("CAS rejects concurrent writers, stale generations, binding drift, counter regression, and exhausted budgets", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-cas-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const storeA = new ProductionFileStateStore(root);
	const storeB = new ProductionFileStateStore(root);
	const initial = createProductionAutonomousState(plan(), { runId: "run-1", now: new Date("2026-07-22T10:00:00Z") });
	await storeA.create(initial);
	const nextA = evolveProductionState(initial, fence(initial), (draft) => { draft.idleReason = "capacity"; });
	const nextB = evolveProductionState(initial, fence(initial), (draft) => { draft.idleReason = "dependencies"; });
	const settlements = await Promise.allSettled([
		storeA.compareAndSwap(fence(initial), nextA),
		storeB.compareAndSwap(fence(initial), nextB),
	]);
	assert.equal(settlements.filter((result) => result.status === "fulfilled").length, 1);
	assert.equal(settlements.filter((result) => result.status === "rejected").length, 1);

	const current = (await storeA.load(479))!;
	const drift = structuredClone(current);
	drift.revision += 1;
	drift.children[0].writeScopes = ["other/path"];
	await assert.rejects(storeA.compareAndSwap(fence(current), drift), /binding|immutable/i);

	const generation2 = advanceProductionGeneration(current, fence(current), "run-2", new Date("2026-07-22T10:02:00Z"));
	await storeA.compareAndSwap(fence(current), generation2);
	await assert.rejects(storeA.compareAndSwap(fence(current), generation2), /stale|CAS|fence/i);

	const overBudget = structuredClone(generation2);
	overBudget.revision += 1;
	overBudget.children[0].attempts = 3;
	await assert.rejects(storeA.compareAndSwap(fence(generation2), overBudget), /attempt|budget/i);

	const regressed = structuredClone(generation2);
	regressed.revision += 1;
	regressed.children[0].attempts = -1;
	await assert.rejects(storeA.compareAndSwap(fence(generation2), regressed), /attempt/i);
});

test("rejects secret-like checkpoint summaries before persistence without echoing the value", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-secret-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const store = new ProductionFileStateStore(root);
	const initial = createProductionAutonomousState(plan(), { runId: "run-1", now: new Date("2026-07-22T09:59:00.000Z") });
	await store.create(initial);
	const secret = "github_pat_SYNTHETIC_DO_NOT_PERSIST_123456789";
	assert.throws(() => evolveProductionState(initial, fence(initial), (draft) => {
		draft.children[0].checkpoint = { summary: `model output ${secret}` };
	}), (error: unknown) => {
		assert.ok(error instanceof Error);
		assert.match(error.message, /secret|credential|sensitive/i);
		assert.doesNotMatch(error.message, /SYNTHETIC/);
		return true;
	});
});

test("only an exact refresh receipt may replace immutable ownership and it clears downstream truth", () => {
	const initial = createProductionAutonomousState(plan(), { runId: "run-1", now: new Date("2026-07-22T09:59:00.000Z") });
	const firstBinding = {
		claimId: "claim-1",
		ownershipId: "owner-1",
		repositoryIdentity: "repo-identity",
		worktreeIdentity: "worktree-1",
		cwd: "/bounded/worktree-1",
		branch: "child/state",
		baseBranch: "feat/471-pi-agent-session-shepherd",
		baseHead: "a".repeat(40),
		head: "b".repeat(40),
		writeScopes: [".pi/extensions/shepherd"],
	};
	const claimed = evolveProductionState(initial, fence(initial), (draft) => {
		draft.children[0].attempts = 1;
		draft.children[0].status = "running";
		draft.children[0].stage = "review";
		draft.children[0].ownership = firstBinding;
		draft.children[0].checkpoint = {
			summary: "reviewed old head",
		workspace: firstBinding,
		pullRequest: 500,
		integrationReceiptDigest: "c".repeat(64),
		review: {
			status: "clean",
			baseHead: "a".repeat(40),
			head: "b".repeat(40),
			resultDigest: "d".repeat(64),
			completedAt: "2026-07-22T10:00:00.000Z",
			findings: [],
		},
	};
	});
	const newBinding = {
		...firstBinding,
		claimId: "claim-2",
		ownershipId: "owner-2",
		worktreeIdentity: "worktree-2",
		cwd: "/bounded/worktree-2",
		baseHead: "e".repeat(40),
		head: "f".repeat(40),
	};
	assert.throws(() => evolveProductionState(claimed, fence(claimed), (draft) => {
		draft.children[0].ownership = newBinding;
	}), /ownership|refresh/i);
	const refreshed = refreshProductionChildOwnership(claimed, fence(claimed), {
		childId: "state",
		previousClaimId: "claim-1",
		previousBaseHead: "a".repeat(40),
		newBinding,
		effectKey: "refresh-effect",
		now: new Date("2026-07-22T10:01:00.000Z"),
	});
	assert.equal(refreshed.children[0].ownership?.claimId, "claim-2");
	assert.equal(refreshed.children[0].stage, "verification");
	assert.equal(refreshed.children[0].checkpoint?.review, undefined);
	assert.equal(refreshed.children[0].checkpoint?.integrationReceiptDigest, undefined);
	assert.throws(() => refreshProductionChildOwnership(claimed, fence(claimed), {
		childId: "state",
		previousClaimId: "wrong-claim",
		previousBaseHead: "a".repeat(40),
		newBinding,
		effectKey: "refresh-effect",
	}), /previous|claim/i);
});

test("budget exhaustion persists an exact child intervention wait instead of terminal failure", () => {
	const initial = createProductionAutonomousState(plan(), { runId: "run-1" });
	const exhausted = evolveProductionState(initial, fence(initial), (draft) => {
		draft.children[0].attempts = draft.children[0].maxAttempts;
		draft.children[0].corrections = draft.children[0].maxCorrections;
		draft.children[0].status = "blocked";
		draft.children[0].lastFailure = {
			kind: "human_required",
			summary: "retry and correction budgets exhausted",
			at: draft.updatedAt,
		};
	});
	const waiting = waitForProductionChildIntervention(exhausted, fence(exhausted), {
		childId: "state",
		requestId: "intervention-501-1",
		reason: "authorize one bounded corrective attempt",
		pullRequest: 500,
		head: "f".repeat(40),
	});
	assert.equal(waiting.status, "waiting_human");
	assert.equal(waiting.childGate?.status, "pending");
	assert.equal(waiting.childGate?.generation, waiting.generation);
	assert.equal(waiting.children[0].status, "blocked");
	assert.equal(waiting.children[0].attempts, 2);
	assert.equal(waiting.children[0].corrections, 1);
	assert.throws(() => evolveProductionState(waiting, fence(waiting), (draft) => {
		draft.childGate = { ...draft.childGate!, head: "e".repeat(40) };
	}), /immutable|gate|binding/i);
});
