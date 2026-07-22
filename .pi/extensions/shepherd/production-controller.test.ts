import assert from "node:assert/strict";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import {
	ProductionLifecycleError,
	type ProductionChildSpec,
	type ProductionParentPlanDocument,
	type ProductionStageCheckpoint,
	type ProductionWorkspaceBinding,
} from "./autonomous-production-contract.ts";
import { ProductionFileStateStore } from "./autonomous-production-state.ts";
import {
	ProductionShepherdController,
	type ProductionChildPipelineContext,
	type ProductionChildPipelinePort,
} from "./production-controller.ts";

const HEAD_A = "a".repeat(40);
const HEAD_B = "b".repeat(40);
const HEAD_C = "c".repeat(40);

function deferred<T>() {
	let resolve!: (value: T) => void;
	let reject!: (reason?: unknown) => void;
	const promise = new Promise<T>((accept, decline) => { resolve = accept; reject = decline; });
	return { promise, resolve, reject };
}

function spec(id: string, issue: number, dependsOn: string[], scope: string): ProductionChildSpec {
	return {
		id,
		issue,
		title: id,
		task: `implement ${id}`,
		slug: id,
		dependsOn,
		access: "mutating",
		writeScopes: [scope],
		requiredSkills: ["javascript-testing-patterns"],
		verification: [{ id: `${id}-tests`, executable: "node", args: ["--test", `${id}.test.ts`], cwd: ".", timeoutMs: 30_000, maxOutputBytes: 1_000_000 }],
		humanGates: [],
		maxAttempts: 2,
		maxCorrections: 1,
	};
}

function plan(children: ProductionChildSpec[]): ProductionParentPlanDocument {
	return {
		schemaVersion: 2,
		planId: "production-479",
		parentIssue: 479,
		repository: "owner/repo",
		title: "Production Shepherd",
		objective: "Complete every autonomous stage",
		parentBranch: "feat/parent",
		parentBaseBranch: "main",
		actorAllowlist: ["maintainer"],
		decisionExpiresAt: "2026-08-01T00:00:00.000Z",
		children,
	};
}

function binding(child: ProductionChildSpec, baseHead = HEAD_A, head = HEAD_B, suffix = "1"): ProductionWorkspaceBinding {
	return {
		claimId: `claim-${child.id}-${suffix}`,
		ownershipId: `owner-${child.id}`,
		repositoryIdentity: "d".repeat(64),
		worktreeIdentity: `worktree-${child.id}-${suffix}`,
		cwd: `/tmp/${child.id}-${suffix}`,
		branch: `feat/${child.issue}-${child.slug}`,
		baseBranch: "feat/parent",
		baseHead,
		head,
		writeScopes: [...child.writeScopes],
	};
}

function checkpoint(summary: string, extra: Partial<ProductionStageCheckpoint> = {}): ProductionStageCheckpoint {
	return { summary, ...extra };
}

function command(action: "start" | "resume") {
	return { action, issue: 479, backend: "sdk-inproc" as const, maxConcurrency: 2, timeoutMs: 30_000 };
}

async function eventually(assertion: () => void): Promise<void> {
	for (let attempt = 0; attempt < 200; attempt += 1) {
		try { assertion(); return; } catch { await new Promise((resolve) => setImmediate(resolve)); }
	}
	assertion();
}

function greenPipeline(overrides: Partial<ProductionChildPipelinePort> = {}): ProductionChildPipelinePort & { calls: string[] } {
	const calls: string[] = [];
	return {
		calls,
		async workspace(context) {
			calls.push(`workspace:${context.child.id}`);
			return checkpoint("workspace", { workspace: binding(context.child) });
		},
		async implement(context) { calls.push(`implement:${context.child.id}`); return checkpoint("implemented"); },
		async verify(context) { calls.push(`verify:${context.child.id}`); return checkpoint("verified"); },
		async publish(context) { calls.push(`publish:${context.child.id}`); return checkpoint("published", { pullRequest: context.child.issue + 1000 }); },
		async review(context) {
			calls.push(`review:${context.child.id}`);
			return checkpoint("reviewed", { review: {
				status: "clean", baseHead: HEAD_A, head: HEAD_B, resultDigest: "e".repeat(64),
				authorizationDigest: "f".repeat(64), completedAt: "2026-07-22T10:00:00.000Z", findings: [],
			} });
		},
		async correct(context) { calls.push(`correct:${context.child.id}`); return checkpoint("corrected"); },
		async refresh(context) {
			calls.push(`refresh:${context.child.id}`);
			return checkpoint("refreshed", { effectKey: "refresh-key", workspace: binding(context.child, HEAD_C, HEAD_B, "2") });
		},
		async integrate(context) { calls.push(`integrate:${context.child.id}`); return checkpoint("integrated", { integrationReceiptDigest: "1".repeat(64), parentHead: HEAD_C }); },
		async requestIntervention(context) {
			calls.push(`intervention:${context.child.id}`);
			return { requestId: `intervention-${context.child.id}`, pullRequest: context.child.issue + 1000, head: HEAD_B };
		},
		async observeIntervention() { return "pending"; },
		async abort(runId) { calls.push(`abort:${runId}`); },
		async join(runId) { calls.push(`join:${runId}`); },
		async close() { calls.push("close"); },
		...overrides,
	};
}

test("production controller drives parallel disjoint children, dependencies, integration, and exact parent human wait", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-controller-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const manifest = plan([
		spec("alpha", 501, [], "owned/alpha"),
		spec("beta", 502, [], "owned/beta"),
		spec("gamma", 503, ["alpha"], "owned/gamma"),
	]);
	const gates = new Map<string, ReturnType<typeof deferred<void>>>();
	const starts: string[] = [];
	const base = greenPipeline();
	const pipeline = greenPipeline({
		async workspace(context: ProductionChildPipelineContext) {
			starts.push(context.child.id);
			const gate = deferred<void>();
			gates.set(context.child.id, gate);
			await gate.promise;
			return checkpoint("workspace", { workspace: binding(context.child) });
		},
	});
	let recoveryCalls = 0;
	let parentRequests = 0;
	const controller = new ProductionShepherdController({
		stateStore: new ProductionFileStateStore(root),
		intake: { async load() { return { plan: manifest, digest: "unused", path: "fixture" }; } },
		recovery: { async open() { recoveryCalls += 1; return { reconciled: 0 }; } },
		pipeline,
		finalizer: { async finalize() { return { pullRequest: 472, head: HEAD_C, summary: "parent ready" }; }, async close() {} },
		parentGate: {
			async request() { parentRequests += 1; return { requestId: "parent-merge-1" }; },
			async observe() { return "pending" as const; },
			async close() {},
		},
		newRunId: () => "run-1",
	});
	const running = controller.start(command("start"));
	await eventually(() => assert.deepEqual(starts, ["alpha", "beta"]));
	assert.equal(starts.includes("gamma"), false);
	gates.get("alpha")!.resolve();
	await eventually(() => assert.deepEqual(starts, ["alpha", "beta", "gamma"]));
	gates.get("beta")!.resolve();
	gates.get("gamma")!.resolve();
	const state = await running;
	assert.equal(state.status, "waiting_human");
	assert.equal(state.stage, "human_decision");
	assert.equal(state.humanGate?.pullRequest, 472);
	assert.equal(state.humanGate?.head, HEAD_C);
	assert.equal(recoveryCalls, 1);
	assert.equal(parentRequests, 1);
	for (const child of manifest.children) {
		assert.deepEqual(pipeline.calls.filter((entry) => entry.endsWith(`:${child.id}`)), [
			`implement:${child.id}`, `verify:${child.id}`, `publish:${child.id}`,
			`review:${child.id}`, `integrate:${child.id}`,
		]);
	}
	assert.equal("mergeMain" in controller, false);
	void base;
});

test("correction and stale-parent refresh both force verification and a fresh exact-head review before integration", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-refresh-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const manifest = plan([spec("alpha", 501, [], "owned/alpha")]);
	const base = greenPipeline();
	let reviews = 0;
	let integrations = 0;
	const pipeline = greenPipeline({
		async review(context) {
			pipeline.calls.push(`review:${context.child.id}`);
			reviews += 1;
			if (reviews === 1) return checkpoint("findings", { review: {
				status: "blocked", baseHead: HEAD_A, head: HEAD_B,
				findings: [{ id: "F-1", summary: "blocking issue" }],
			} });
			return base.review(context);
		},
		async integrate(context) {
			pipeline.calls.push(`integrate:${context.child.id}`);
			integrations += 1;
			if (integrations === 1) throw new ProductionLifecycleError("stale_parent", "parent moved", ["head_moved"]);
			return checkpoint("integrated", { integrationReceiptDigest: "1".repeat(64), parentHead: HEAD_C });
		},
	});
	const controller = new ProductionShepherdController({
		stateStore: new ProductionFileStateStore(root),
		intake: { async load() { return { plan: manifest, digest: "unused", path: "fixture" }; } },
		recovery: { async open() { return { reconciled: 0 }; } },
		pipeline,
		finalizer: { async finalize() { return { pullRequest: 472, head: HEAD_C, summary: "ready" }; }, async close() {} },
		parentGate: { async request() { return { requestId: "merge" }; }, async observe() { return "pending" as const; }, async close() {} },
		newRunId: () => "run-refresh",
	});
	const state = await controller.start(command("start"));
	assert.equal(state.status, "waiting_human");
	assert.equal(state.children[0].corrections, 1);
	assert.equal(pipeline.calls.filter((entry) => entry === "correct:alpha").length, 1);
	assert.equal(pipeline.calls.filter((entry) => entry === "refresh:alpha").length, 1);
	assert.equal(pipeline.calls.filter((entry) => entry === "verify:alpha").length, 3);
	assert.equal(pipeline.calls.filter((entry) => entry === "review:alpha").length, 3);
	assert.equal(pipeline.calls.filter((entry) => entry === "integrate:alpha").length, 2);
	assert.equal(state.children[0].ownership?.baseHead, HEAD_C);
	assert.equal(state.children[0].checkpoint?.review?.status, "clean");
});
