import assert from "node:assert/strict";
import { mkdtemp, stat } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import {
	createProductionAutonomousState,
	evolveProductionState,
	ProductionFileStateStore,
	type ProductionStateFence,
} from "./autonomous-production-state.ts";
import type {
	ProductionEffectRecord,
	ProductionParentPlanDocument,
} from "./autonomous-production-contract.ts";
import type { ProductionEffectJournalPort } from "./autonomous-effect-journal.ts";
import type { ProductionEffectRecoveryPort } from "./autonomous-recovery.ts";
import { GitAdapter, type GitBinding } from "./git-adapter.ts";
import type { ParentReadyDurableAuthorityBoundary } from "./github-orchestrator.ts";
import { ProductionShepherdController } from "./production-controller.ts";
import type { ProductionParentReadyTransitionPort } from "./production-parent-lifecycle.ts";
import {
	composeProductionShepherdController,
	createExactHeadReviewRoleRequestFactory,
	createProductionShepherdController,
	type ProductionControllerCompositionOptions,
	type ProductionGitObjectReadRequest,
} from "./production-runtime.ts";

const SHA_A = "a".repeat(40);
const SHA_B = "b".repeat(40);
const DIGEST_A = "a".repeat(64);
const DIGEST_B = "b".repeat(64);

function plan(): ProductionParentPlanDocument {
	return {
		schemaVersion: 2,
		planId: "production-479",
		parentIssue: 479,
		repository: "acme/widgets",
		title: "Production Shepherd",
		objective: "Exercise the production composition only.",
		parentBranch: "feat/471-parent",
		parentBaseBranch: "main",
		actorAllowlist: ["maintainer"],
		decisionExpiresAt: "2026-08-01T00:00:00.000Z",
		children: [{
			id: "runtime",
			issue: 479,
			title: "Runtime composition",
			task: "Compose the production runtime.",
			slug: "runtime-composition",
			dependsOn: [],
			access: "mutating",
			writeScopes: [".pi/extensions/shepherd/production-runtime.ts"],
			requiredSkills: ["architecture-patterns"],
			verification: [{
				id: "focused-test",
				executable: "node",
				args: ["--test", ".pi/extensions/shepherd/production-runtime.test.ts"],
				cwd: ".",
				timeoutMs: 30_000,
				maxOutputBytes: 1024 * 1024,
			}],
			humanGates: [],
			maxAttempts: 2,
			maxCorrections: 1,
		}],
	};
}

function coordinator(cwd = "/tmp/production-runtime"): GitBinding {
	return {
		cwd,
		repositoryIdentity: "1".repeat(64),
		worktreeIdentity: "2".repeat(64),
		remoteName: "origin",
		remoteIdentity: "3".repeat(64),
		fetchEndpointIdentity: "4".repeat(64),
		pushEndpointIdentity: "4".repeat(64),
		defaultBranch: "main",
	};
}

function fence(state: ReturnType<typeof createProductionAutonomousState>): ProductionStateFence {
	return {
		issue: state.parentIssue,
		revision: state.revision,
		generation: state.generation,
		runId: state.runId,
	};
}

function unexpected(name: string): never {
	throw new Error(`unexpected ${name}`);
}

test("production factory synchronously selects durable production adapters and rejects a missing recovery authority", async () => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-runtime-"));
	const git = new GitAdapter({ execute: async () => Buffer.from("") });
	const effectRecovery: ProductionEffectRecoveryPort = {
		async observe() { return { resultDigest: DIGEST_A }; },
		async apply() {},
	};
	const parentReadyAuthority = {
		async readParentReadyState() { return null; },
		async beginParentReady() { return unexpected("begin parent ready"); },
		async compareConsumeAndMarkParentReady() { return unexpected("mark parent ready"); },
		async settleParentReady() { return unexpected("settle parent ready"); },
		async quarantineAndRollbackParentReady() { return unexpected("rollback parent ready"); },
	} as unknown as ParentReadyDurableAuthorityBoundary;
	const parentReadiness: ProductionParentReadyTransitionPort = {
		async markExistingDraftReady() { return unexpected("draft-to-ready transition"); },
	};
	const options = {
		repositoryRoot: root,
		stateRoot: join(root, "state"),
		trustedWorktreeRoot: join(root, "worktrees"),
		coordinator: coordinator(root),
		git,
		agentSession: {
			async run() { return unexpected("implementation AgentSession"); },
			async abort() {},
			async close() {},
		},
		reviewSession: {
			async run() { return unexpected("review AgentSession"); },
		},
		effectRecovery,
		parentReadyAuthority,
		parentReadiness,
		dispositionActor: "shepherd-controller",
	};

	const controller = createProductionShepherdController(options);
	assert.ok(controller instanceof ProductionShepherdController);
	assert.equal(await controller.status(479), undefined);
	assert.equal((await stat(options.stateRoot)).isDirectory(), true);
	await controller.shutdown();

	assert.throws(
		() => createProductionShepherdController({ ...options, effectRecovery: undefined as never }),
		/recovery authority is required/,
	);
});

test("composition reconciles an already-checkpointed crash effect once before scheduling without replaying it", async () => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-recovery-"));
	const store = new ProductionFileStateStore(join(root, "state"));
	const initial = createProductionAutonomousState(plan(), {
		runId: "run-before-crash",
		now: new Date("2026-07-22T10:00:00.000Z"),
		maxConcurrency: 1,
		timeoutMs: 30_000,
	});
	await store.create(initial);
	const checkpointed = evolveProductionState(initial, fence(initial), (draft) => {
		draft.status = "stopped";
		draft.stage = "schedule";
		const child = draft.children[0];
		child.status = "succeeded";
		child.stage = "succeeded";
		child.checkpoint = {
			summary: "effect was projected before the process crashed",
			effectKey: DIGEST_A,
			effectKeys: [DIGEST_A],
			pullRequest: 47,
			integrationReceiptDigest: DIGEST_B,
			parentHead: SHA_B,
		};
	}, new Date("2026-07-22T10:01:00.000Z"));
	await store.compareAndSwap(fence(initial), checkpointed);

	const pending: ProductionEffectRecord = {
		schemaVersion: 1,
		key: DIGEST_A,
		kind: "child_pull_request",
		phase: "observed",
		runId: checkpointed.runId,
		generation: checkpointed.generation,
		childId: "runtime",
		intentDigest: DIGEST_B,
		preparedAt: "2026-07-22T10:00:30.000Z",
		observedAt: "2026-07-22T10:00:31.000Z",
		resultDigest: DIGEST_B,
	};
	let journalApplyCount = 0;
	let recoveryApplyCount = 0;
	let recoveryObserveCount = 0;
	let externalReplayCount = 0;
	let journalRecord: ProductionEffectRecord | undefined = structuredClone(pending);
	const effects: ProductionEffectJournalPort = {
		async prepare() { return unexpected("effect preparation"); },
		async load(key) { return key === journalRecord?.key ? structuredClone(journalRecord) : undefined; },
		async listNonApplied() { return journalRecord?.phase === "applied" ? [] : [structuredClone(journalRecord!)]; },
		async observe() { return unexpected("observed effect replay"); },
		async apply(key) {
			assert.equal(key, DIGEST_A);
			journalApplyCount += 1;
			journalRecord = { ...journalRecord!, phase: "applied", appliedAt: "2026-07-22T10:02:00.000Z" };
			return structuredClone(journalRecord);
		},
	};
	const effectRecovery: ProductionEffectRecoveryPort = {
		async observe() {
			recoveryObserveCount += 1;
			externalReplayCount += 1;
			return { resultDigest: DIGEST_B };
		},
		async apply(record) {
			recoveryApplyCount += 1;
			assert.equal(record.key, DIGEST_A);
			assert.equal(record.phase, "observed");
			const durable = await store.load(479);
			assert.equal(durable?.children[0].checkpoint?.effectKeys?.includes(record.key), true);
		},
	};

	const inert = async () => unexpected("child pipeline stage");
	const composition: ProductionControllerCompositionOptions = {
		stateStore: store,
		intake: { async load() { return { plan: plan(), digest: DIGEST_A, path: "/repo/.planning/shepherd/issue-479.json" }; } },
		effects,
		effectRecovery,
		workspaceLifecycle: { claim: inert, async abort() {}, async close() {} },
		github: {
			createPlan: inert,
			ensureChildIssue: inert,
			ensureChildPullRequest: inert,
			integrateChild: inert,
			async stop() { return { kind: "joined" as const, active: 0, unacknowledged: 0 }; },
		},
		reviewer: { review: inert },
		reviewRepository: { find: inert, publish: inert, recordDispositions: inert },
		decisionBroker: { request: inert, poll: inert, consume: inert },
		parentHeads: { observe: inert },
		coordinator: coordinator(root),
		trustedWorktreeRoot: join(root, "worktrees"),
		dispositionActor: "shepherd-controller",
		finalizer: {
			async finalize() { return { pullRequest: 472, head: SHA_B, summary: "exact parent finalization" }; },
			async close() {},
		},
		parentGate: {
			async request() { return { requestId: "parent-gate-479" }; },
			async observe() { return { status: "pending" as const }; },
			async close() {},
		},
		newRunId: () => "run-after-crash",
		now: () => new Date("2026-07-22T10:02:00.000Z"),
	};
	const controller = composeProductionShepherdController(composition);
	const result = await controller.resume({
		action: "resume",
		issue: 479,
		backend: "sdk-inproc",
		maxConcurrency: 1,
		timeoutMs: 30_000,
	});

	assert.equal(result.status, "waiting_human");
	assert.equal(result.resourceGeneration, 1);
	assert.equal(result.generation, 2);
	assert.equal(recoveryObserveCount, 0);
	assert.equal(recoveryApplyCount, 1);
	assert.equal(journalApplyCount, 1);
	assert.equal(externalReplayCount, 0);
	await controller.shutdown();
});

test("exact-head review request reads only bounded Git objects and exposes no mutation", async () => {
	const binding = coordinator("/repo");
	const reads: ProductionGitObjectReadRequest[] = [];
	const git = {
		async assertBinding(value: GitBinding) { assert.deepEqual(value, binding); return value; },
		async resolveBranchHead(_value: GitBinding, branch: string) {
			assert.equal(branch, "feat/479-runtime-composition");
			return SHA_B;
		},
	};
	const factory = createExactHeadReviewRoleRequestFactory({
		git,
		coordinator: binding,
		parentIssue: 479,
		readObject: async (request) => {
			reads.push(request);
			return Buffer.from("line one\nline two\n");
		},
		now: () => new Date("2026-07-22T10:00:00.000Z"),
	});
	const signal = new AbortController().signal;
	const request = factory({
		schemaVersion: 1,
		kind: "codex_independent",
		provider: "openai-codex",
		model: "gpt-5.6-sol",
		reasoningEffort: "xhigh",
		readOnly: true,
		repository: "acme/widgets",
		workItemId: "runtime",
		pullRequest: 47,
		generation: 3,
		baseBranch: "feat/471-parent",
		headBranch: "feat/479-runtime-composition",
		baseSha: SHA_A,
		headSha: SHA_B,
		changedPaths: ["src/runtime.ts"],
		allowedScopes: ["src"],
		idempotencyMarker: "review-runtime-3",
	}, {
		signal,
		deadlineAt: "2026-07-22T10:00:30.000Z",
		acknowledgeAbort() {},
	});

	assert.equal(request.role, "review");
	assert.equal(request.authority.readOnly, true);
	assert.deepEqual(request.authority.writePrefixes, []);
	assert.deepEqual(request.capabilities, []);
	assert.equal(await request.workspace.readText("src/runtime.ts", { offset: 5, limit: 3, signal }), "one");
	assert.deepEqual(reads.map(({ cwd, headSha, path, maxOutputBytes }) => ({ cwd, headSha, path, maxOutputBytes })), [{
		cwd: "/repo",
		headSha: SHA_B,
		path: "src/runtime.ts",
		maxOutputBytes: 256 * 1024,
	}]);
	await assert.rejects(request.workspace.readText("../secret", { signal }), /traversal/);
	await assert.rejects(request.workspace.writeText("src/runtime.ts", "changed", signal), /read-only/);
	await assert.rejects(request.workspace.editText("src/runtime.ts", "one", "two", signal), /read-only/);
});
