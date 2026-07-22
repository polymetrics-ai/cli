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

function passedVerification(summary = "verified"): ProductionStageCheckpoint {
	return checkpoint(summary, {
		verification: {
			status: "passed",
			resultDigest: "9".repeat(64),
			commands: [{ id: "tests", status: "passed" }],
		},
	});
}

function command(action: "start" | "resume") {
	return { action, issue: 479, backend: "sdk-inproc" as const, maxConcurrency: 2, timeoutMs: 30_000 };
}

async function eventually(assertion: () => void): Promise<void> {
	for (let attempt = 0; attempt < 200; attempt += 1) {
		try { assertion(); return; } catch { await new Promise((resolve) => setTimeout(resolve, 2)); }
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
		async verify(context) { calls.push(`verify:${context.child.id}`); return passedVerification(); },
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
		async observeIntervention() { return { status: "pending" }; },
		async acknowledge(effectKey) { calls.push(`ack:${effectKey}`); },
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
			async observe() { return { status: "pending" as const }; },
			async close() {},
		},
		newRunId: () => "run-1",
	});
	const running = controller.start(command("start"));
	let startFailure: unknown;
	void running.catch((error) => { startFailure = error; });
	await eventually(() => {
		if (startFailure !== undefined) throw startFailure;
		assert.deepEqual(starts, ["alpha", "beta"]);
	});
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
		parentGate: { async request() { return { requestId: "merge" }; }, async observe() { return { status: "pending" as const }; }, async close() {} },
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

test("failed pre-publication verification runs one bounded correction before publish", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-pre-pr-correction-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const manifest = plan([spec("alpha", 501, [], "owned/alpha")]);
	let verificationRuns = 0;
	const pipeline = greenPipeline({
		async verify(context) {
			pipeline.calls.push(`verify:${context.child.id}`);
			verificationRuns += 1;
			return checkpoint(verificationRuns === 1 ? "verification failed" : "verification passed", {
				verification: verificationRuns === 1 ? {
					status: "failed", resultDigest: "8".repeat(64),
					commands: [{ id: "tests", status: "failed", failureKind: "exit" }],
				} : {
					status: "passed", resultDigest: "9".repeat(64), commands: [{ id: "tests", status: "passed" }],
				},
			});
		},
	});
	const controller = new ProductionShepherdController({
		stateStore: new ProductionFileStateStore(root),
		intake: { async load() { return { plan: manifest, digest: "unused", path: "fixture" }; } },
		recovery: { async open() { return { reconciled: 0 }; } },
		pipeline,
		finalizer: { async finalize() { return { pullRequest: 438, head: HEAD_C, summary: "ready" }; }, async close() {} },
		parentGate: { async request() { return { requestId: "merge" }; }, async observe() { return { status: "pending" as const }; }, async close() {} },
		newRunId: () => "run-pre-pr-correction",
	});
	const state = await controller.start(command("start"));
	assert.equal(state.status, "waiting_human");
	assert.deepEqual(pipeline.calls.filter((entry) => /^(?:verify|correct|publish):alpha$/.test(entry)), [
		"verify:alpha", "correct:alpha", "verify:alpha", "publish:alpha",
	]);
});

test("exhausted retries wait for an exact child decision and authorized resume continues the failed stage once", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-retry-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const child = spec("alpha", 501, [], "owned/alpha");
	child.maxAttempts = 1;
	const manifest = plan([child]);
	let verificationRuns = 0;
	const base = greenPipeline();
	const pipeline = greenPipeline({
		async verify(context) {
			pipeline.calls.push(`verify:${context.child.id}`);
			verificationRuns += 1;
			if (verificationRuns === 1) throw new ProductionLifecycleError("retryable", "transient verification");
			return passedVerification();
		},
		async observeIntervention() { return { status: "authorized" }; },
	});
	const store = new ProductionFileStateStore(root);
	let nextRun = 0;
	const options = {
		stateStore: store,
		intake: { async load() { return { plan: manifest, digest: "unused", path: "fixture" }; } },
		recovery: { async open() { return { reconciled: 0 }; } },
		pipeline,
		finalizer: { async finalize() { return { pullRequest: 472, head: HEAD_C, summary: "ready" }; }, async close() {} },
		parentGate: { async request() { return { requestId: "merge" }; }, async observe() { return { status: "pending" as const }; }, async close() {} },
		newRunId: () => `run-retry-${++nextRun}`,
	};
	const first = await new ProductionShepherdController(options).start(command("start"));
	assert.equal(first.status, "waiting_human");
	assert.equal(first.childGate?.reason, "retry_budget_exhausted");
	assert.equal(first.children[0].attempts, 1);
	assert.equal(first.children[0].stage, "verification");

	const resumed = await new ProductionShepherdController(options).resume(command("resume"));
	assert.equal(resumed.status, "waiting_human");
	assert.equal(resumed.humanGate?.requestId, "merge");
	assert.equal(resumed.generation, 2);
	assert.equal(resumed.resourceGeneration, 1, "resume must retain exact PR/review/integration resource identity");
	assert.equal(resumed.children[0].attempts, 2);
	assert.equal(resumed.children[0].authorizedAttempts, 1);
	assert.equal(pipeline.calls.filter((entry) => entry === "workspace:alpha").length, 1);
	assert.equal(pipeline.calls.filter((entry) => entry === "implement:alpha").length, 1);
	assert.equal(pipeline.calls.filter((entry) => entry === "verify:alpha").length, 2);
	assert.equal(pipeline.calls.filter((entry) => entry === "publish:alpha").length, 1);
	assert.equal(pipeline.calls.filter((entry) => entry === "review:alpha").length, 1);
	void base;
});

test("stop cancels and joins accepted work at every external child lifecycle stage", async (t) => {
	for (const stage of ["workspace", "implementation", "verification", "publication", "review", "integration"] as const) {
		await t.test(stage, async () => {
			const root = await mkdtemp(join(tmpdir(), `shepherd-production-stop-${stage}-`));
			t.after(() => rm(root, { recursive: true, force: true }));
			const manifest = plan([spec("alpha", 501, [], "owned/alpha")]);
			const entered = deferred<void>();
			const release = deferred<ProductionStageCheckpoint>();
			const pipeline = greenPipeline();
			const block = async (context: ProductionChildPipelineContext) => {
				pipeline.calls.push(`${stage}:${context.child.id}`);
				entered.resolve();
				return release.promise;
			};
			switch (stage) {
				case "workspace": pipeline.workspace = block; break;
				case "implementation": pipeline.implement = block; break;
				case "verification": pipeline.verify = block; break;
				case "publication": pipeline.publish = block; break;
				case "review": pipeline.review = block; break;
				case "integration": pipeline.integrate = block; break;
			}
			pipeline.abort = async (runId) => {
				pipeline.calls.push(`abort:${runId}`);
				release.resolve(checkpoint("cancelled stage settled"));
			};
			const controller = new ProductionShepherdController({
				stateStore: new ProductionFileStateStore(root),
				intake: { async load() { return { plan: manifest, digest: "unused", path: "fixture" }; } },
				recovery: { async open() { return { reconciled: 0 }; } },
				pipeline,
				finalizer: { async finalize() { throw new Error("stopped run must not finalize"); }, async close() {} },
				parentGate: { async request() { throw new Error("stopped run must not request parent gate"); }, async observe() { return { status: "pending" as const }; }, async close() {} },
				newRunId: () => `run-stop-${stage}`,
			});
			const running = controller.start(command("start"));
			await entered.promise;
			const stopped = await controller.stop(479);
			assert.deepEqual(await running, stopped);
			assert.equal(stopped.status, "stopped");
			assert.equal(stopped.children[0].status, "cancelled");
			assert.equal(stopped.children[0].resumeStage, stage);
			const abortIndex = pipeline.calls.indexOf(`abort:run-stop-${stage}`);
			const joinIndex = pipeline.calls.indexOf(`join:run-stop-${stage}`);
			assert.ok(abortIndex >= 0 && joinIndex > abortIndex, "accepted work must join after abort");
		});
	}
});

test("one child failure aborts and joins its running sibling before durable failure", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-sibling-abort-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const manifest = plan([
		spec("alpha", 501, [], "owned/alpha"),
		spec("beta", 502, [], "owned/beta"),
	]);
	const betaEntered = deferred<void>();
	const betaRelease = deferred<ProductionStageCheckpoint>();
	const pipeline = greenPipeline({
		async implement(context) {
			pipeline.calls.push(`implement:${context.child.id}`);
			if (context.child.id === "alpha") {
				await betaEntered.promise;
				throw new ProductionLifecycleError("terminal", "bounded terminal failure");
			}
			betaEntered.resolve();
			return betaRelease.promise;
		},
		async abort(runId) {
			pipeline.calls.push(`abort:${runId}`);
			betaRelease.resolve(checkpoint("sibling abort settled"));
		},
	});
	const controller = new ProductionShepherdController({
		stateStore: new ProductionFileStateStore(root),
		intake: { async load() { return { plan: manifest, digest: "unused", path: "fixture" }; } },
		recovery: { async open() { return { reconciled: 0 }; } },
		pipeline,
		finalizer: { async finalize() { throw new Error("failed siblings must not finalize"); }, async close() {} },
		parentGate: { async request() { throw new Error("failed siblings must not request a gate"); }, async observe() { return { status: "pending" as const }; }, async close() {} },
		newRunId: () => "run-sibling-abort",
	});
	const state = await controller.start(command("start"));
	assert.equal(state.status, "failed");
	assert.equal(state.terminalBlocker, "production child lifecycle failed closed");
	assert.ok(pipeline.calls.includes("abort:run-sibling-abort"));
	assert.ok(pipeline.calls.indexOf("join:run-sibling-abort") > pipeline.calls.indexOf("abort:run-sibling-abort"));
	assert.notEqual(state.children.find((child) => child.id === "beta")?.status, "running");
});

test("resume rejects changed plan identity and run policy before recovery or mutation", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-resume-binding-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const original = plan([spec("alpha", 501, [], "owned/alpha")]);
	let currentPlan = original;
	const entered = deferred<void>();
	const release = deferred<ProductionStageCheckpoint>();
	const pipeline = greenPipeline({
		async workspace(context) {
			pipeline.calls.push(`workspace:${context.child.id}`);
			entered.resolve();
			return release.promise;
		},
		async abort(runId) {
			pipeline.calls.push(`abort:${runId}`);
			release.resolve(checkpoint("cancelled"));
		},
	});
	let recoveries = 0;
	let runNumber = 0;
	const options = {
		stateStore: new ProductionFileStateStore(root),
		intake: { async load() { return { plan: currentPlan, digest: "unused", path: "fixture" }; } },
		recovery: { async open() { recoveries += 1; return { reconciled: 0 }; } },
		pipeline,
		finalizer: { async finalize() { throw new Error("stopped run must not finalize"); }, async close() {} },
		parentGate: { async request() { throw new Error("stopped run must not request a gate"); }, async observe() { return { status: "pending" as const }; }, async close() {} },
		newRunId: () => `run-binding-${++runNumber}`,
	};
	const firstController = new ProductionShepherdController(options);
	const running = firstController.start(command("start"));
	await entered.promise;
	await firstController.stop(479);
	await running;
	assert.equal(recoveries, 1);

	const changed = structuredClone(original);
	changed.children[0].writeScopes = ["owned/escaped"];
	currentPlan = changed;
	await assert.rejects(
		new ProductionShepherdController(options).resume(command("resume")),
		/plan binding changed/i,
	);
	assert.equal(recoveries, 1);
	assert.equal(pipeline.calls.filter((entry) => entry.startsWith("workspace:")).length, 1);

	currentPlan = original;
	await assert.rejects(
		new ProductionShepherdController(options).resume({ ...command("resume"), timeoutMs: 60_000 }),
		/resume concurrency\/timeout differs/i,
	);
	assert.equal(pipeline.calls.filter((entry) => entry.startsWith("workspace:")).length, 1);
});

test("effect acknowledgment happens only after the exact checkpoint is durably persisted", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-effect-ack-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const manifest = plan([spec("alpha", 501, [], "owned/alpha")]);
	const store = new ProductionFileStateStore(root);
	const acknowledgments: Array<{
		effectKey: string;
		stateRevision: number;
		stateEffectKey?: string;
		stateEffectKeys?: string[];
		durableRevision?: number;
		durableEffectKey?: string;
		durableEffectKeys?: string[];
	}> = [];
	const pipeline = greenPipeline({
		async verify(context) {
			pipeline.calls.push(`verify:${context.child.id}`);
			return checkpoint("verified exact head", {
				verification: {
					status: "passed", resultDigest: "9".repeat(64), commands: [{ id: "tests", status: "passed" }],
				},
				effectKey: "verification-effect",
				effectKeys: ["verification-log-effect"],
			});
		},
		async acknowledge(effectKey, state) {
			pipeline.calls.push(`ack:${effectKey}`);
			const child = state.children.find((candidate) => candidate.id === "alpha");
			const durable = await store.load(479);
			acknowledgments.push({
				effectKey,
				stateRevision: state.revision,
				...(child?.checkpoint?.effectKey === undefined ? {} : { stateEffectKey: child.checkpoint.effectKey }),
				...(child?.checkpoint?.effectKeys === undefined ? {} : { stateEffectKeys: child.checkpoint.effectKeys }),
				...(durable?.revision === undefined ? {} : { durableRevision: durable.revision }),
				...(durable?.children[0].checkpoint?.effectKey === undefined
					? {} : { durableEffectKey: durable.children[0].checkpoint.effectKey }),
				...(durable?.children[0].checkpoint?.effectKeys === undefined
					? {} : { durableEffectKeys: durable.children[0].checkpoint.effectKeys }),
			});
		},
	});
	const controller = new ProductionShepherdController({
		stateStore: store,
		intake: { async load() { return { plan: manifest, digest: "unused", path: "fixture" }; } },
		recovery: { async open() { return { reconciled: 0 }; } },
		pipeline,
		finalizer: { async finalize() { return { pullRequest: 472, head: HEAD_C, summary: "ready" }; }, async close() {} },
		parentGate: { async request() { return { requestId: "merge" }; }, async observe() { return { status: "pending" as const }; }, async close() {} },
		newRunId: () => "run-effect-ack",
	});
	const state = await controller.start(command("start"));
	assert.equal(state.status, "waiting_human");
	assert.equal(pipeline.calls.filter((entry) => entry === "ack:verification-effect").length, 1);
	assert.equal(pipeline.calls.filter((entry) => entry === "ack:verification-log-effect").length, 1);
	assert.equal(acknowledgments.length, 2);
	for (const acknowledgment of acknowledgments) {
		assert.equal(acknowledgment.stateEffectKey, "verification-effect");
		assert.deepEqual(acknowledgment.stateEffectKeys, ["verification-effect", "verification-log-effect"]);
		assert.equal(acknowledgment.durableRevision, acknowledgment.stateRevision);
		assert.equal(acknowledgment.durableEffectKey, "verification-effect");
		assert.deepEqual(acknowledgment.durableEffectKeys, ["verification-effect", "verification-log-effect"]);
	}
});

test("resume completes only from an exact-head authoritative merge and persists its evidence", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-parent-merge-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const manifest = plan([spec("alpha", 501, [], "owned/alpha")]);
	const pipeline = greenPipeline();
	let merged = false;
	let run = 0;
	const options = {
		stateStore: new ProductionFileStateStore(root),
		intake: { async load() { return { plan: manifest, digest: "unused", path: "fixture" }; } },
		recovery: { async open() { return { reconciled: 0 }; } },
		pipeline,
		finalizer: { async finalize() { return { pullRequest: 472, head: HEAD_C, summary: "ready" }; }, async close() {} },
		parentGate: {
			async request() { return { requestId: "merge-exact-parent" }; },
			async observe() {
				return merged ? {
					status: "merged" as const,
					repository: "owner/repo",
					pullRequest: 472,
					head: HEAD_C,
					mergedAt: "2026-07-22T12:00:00.000Z",
					mergeCommitSha: "9".repeat(40),
					revision: 27,
					observedAt: "2026-07-22T12:00:01.000Z",
				} : { status: "approved_waiting_for_merge" as const };
			},
			async close() {},
		},
		newRunId: () => `run-parent-merge-${++run}`,
	};
	const waiting = await new ProductionShepherdController(options).start(command("start"));
	assert.equal(waiting.status, "waiting_human");
	assert.equal((await new ProductionShepherdController(options).resume(command("resume"))).status, "waiting_human");
	merged = true;
	const completed = await new ProductionShepherdController(options).resume(command("resume"));
	assert.equal(completed.status, "completed");
	assert.equal(completed.stage, "completed");
	assert.equal(completed.humanGate?.status, "merged");
	assert.deepEqual(completed.humanGate?.mergeEvidence, {
		mergedAt: "2026-07-22T12:00:00.000Z",
		mergeCommitSha: "9".repeat(40),
		revision: 27,
		observedAt: "2026-07-22T12:00:01.000Z",
	});

	const movedOptions = {
		...options,
		parentGate: {
			...options.parentGate,
			async observe() {
				return {
					status: "merged" as const,
					repository: "owner/repo",
					pullRequest: 472,
					head: HEAD_B,
					mergedAt: "2026-07-22T12:00:00.000Z",
					mergeCommitSha: "9".repeat(40),
					revision: 28,
					observedAt: "2026-07-22T12:00:02.000Z",
				};
			},
		},
	};
	await assert.rejects(
		new ProductionShepherdController(movedOptions).resume(command("resume")),
		/exact-head gate/i,
	);
});
