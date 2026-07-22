import assert from "node:assert/strict";
import { createHash } from "node:crypto";
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
import { createAgentSessionAttestation, createIndependentReviewWork } from "./review-router.ts";
import { productionWorkspaceOwnershipId } from "./production-workspace-lifecycle.ts";
import {
	composeProductionShepherdController,
	createExactHeadReviewRoleRequestFactory,
	createProductionRecoveryProbeTable,
	createProductionShepherdController,
	type ProductionControllerCompositionOptions,
	type ProductionGitObjectReadRequest,
	type ProductionRuntimeRecoveryProbeOptions,
} from "./production-runtime.ts";

const SHA_A = "a".repeat(40);
const SHA_B = "b".repeat(40);
const SHA_C = "c".repeat(40);
const DIGEST_A = "a".repeat(64);
const DIGEST_B = "b".repeat(64);
const EFFECT_KINDS = [
	"workspace_claim", "agent_implementation", "agent_correction", "shell_verification", "git_commit", "git_push",
	"child_pull_request", "independent_review", "child_integration", "parent_refresh", "child_head_reconciliation",
	"human_request", "human_consume", "parent_merge_observation",
] as const;

function stableDigest(value: unknown): string {
	const canonical = (item: unknown): unknown => {
		if (Array.isArray(item)) return item.map(canonical);
		if (item === null || typeof item !== "object") return item;
		return Object.fromEntries(Object.keys(item as Record<string, unknown>).sort().flatMap((key) => {
			const child = (item as Record<string, unknown>)[key];
			return child === undefined ? [] : [[key, canonical(child)]];
		}));
	};
	return createHash("sha256").update(JSON.stringify(canonical(value))).digest("hex");
}

function plan(): ProductionParentPlanDocument {
	return {
		schemaVersion: 2 as const,
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

function recoveryState() {
	const state = createProductionAutonomousState(plan(), {
		runId: "run-recovery",
		now: new Date("2026-07-22T10:00:00.000Z"),
		maxConcurrency: 1,
		timeoutMs: 30_000,
	});
	state.stage = "child_lifecycle";
	state.children[0].status = "running";
	state.children[0].stage = "verification";
	state.children[0].attempts = 1;
	state.children[0].ownership = {
		claimId: DIGEST_A,
		ownershipId: "production:runtime",
		repositoryIdentity: "1".repeat(64),
		worktreeIdentity: "2".repeat(64),
		cwd: "/tmp/production-runtime/issue-479",
		branch: "feat/479-runtime-composition",
		baseBranch: "feat/471-parent",
		baseHead: SHA_A,
		head: SHA_B,
		writeScopes: [".pi/extensions/shepherd/production-runtime.ts"],
	};
	return state;
}

function recoveryRecord(
	kind: typeof EFFECT_KINDS[number],
	descriptor: unknown,
): ProductionEffectRecord {
	return {
		schemaVersion: 1,
		key: DIGEST_B,
		kind,
		phase: "prepared",
		runId: "run-recovery",
		generation: 1,
		...(kind === "parent_merge_observation" ? {} : { childId: "runtime" }),
		intentDigest: DIGEST_A,
		recoveryDescriptor: descriptor,
		preparedAt: "2026-07-22T10:00:00.500Z",
	};
}

function recoveryProbeOptions(
	overrides: Partial<ProductionRuntimeRecoveryProbeOptions> = {},
): ProductionRuntimeRecoveryProbeOptions {
	const inert = async () => unexpected("recovery evidence port");
	return {
		git: {
			assertBinding: inert,
			inspect: inert,
			currentBranch: inert,
			resolveBranchHead: inert,
			resolveRemoteBranchHead: inert,
			readCommitSubject: inert,
			isAncestor: inert,
			status: inert,
			diff: inert,
		},
		workspace: {
			findClaim: inert,
			findParentRefreshReceipt: inert,
			findChildHeadReceipt: inert,
		},
		agentEffects: { find: inert },
		coordinator: coordinator(),
		trustedWorktreeRoot: "/tmp/production-runtime/worktrees",
		verification: { runAll: inert },
		github: { findPullRequests: inert, findChildIntegration: inert, proveAncestry: inert },
		reviews: { find: inert },
		decisions: { load: inert },
		parentMerges: { observeExactPullRequest: inert },
		parentHeads: { observe: inert },
		dispositionActor: "shepherd-controller",
		now: () => new Date("2026-07-22T10:00:01.000Z"),
		...overrides,
	};
}

test("default workspace recovery accepts only the exact immutable claim evidence", async () => {
	const state = recoveryState();
	delete state.children[0].ownership;
	state.children[0].stage = "workspace";
	const ownershipId = productionWorkspaceOwnershipId(479, 479, "runtime");
	const descriptor = {
		operation: "workspace_claim",
		childId: "runtime",
		childIssue: 479,
		childSlug: "runtime-composition",
		generation: 1,
		parentBaseBranch: "main",
		parentBranch: "feat/471-parent",
		parentIssue: 479,
		planDigest: state.planDigest,
		repository: "acme/widgets",
		parentHead: SHA_A,
		mode: "start",
		attempt: 1,
		coordinator: coordinator(),
		trustedWorktreeRoot: "/tmp/production-runtime/worktrees",
		writeScopes: [".pi/extensions/shepherd/production-runtime.ts"],
	};
	const evidence = {
		claimId: DIGEST_A,
		repositoryIdentity: "1".repeat(64),
		worktreeIdentity: "2".repeat(64),
		cwd: "/tmp/production-runtime/issue-479",
		branch: "feat/479-runtime-composition",
		baseBranch: "feat/471-parent",
		baseHead: SHA_A,
		head: SHA_B,
		writeScopes: [".pi/extensions/shepherd/production-runtime.ts"],
		clean: true,
		changedScope: [],
	};
	const calls: unknown[] = [];
	const probes = createProductionRecoveryProbeTable(recoveryProbeOptions({
		workspace: {
			async findClaim(request) { calls.push(request); return evidence; },
			async findParentRefreshReceipt() { return undefined; },
			async findChildHeadReceipt() { return undefined; },
		},
	} as Partial<ProductionRuntimeRecoveryProbeOptions>));
	const result = await probes.workspace_claim({
		record: recoveryRecord("workspace_claim", descriptor),
		descriptor,
		currentState: state,
		signal: new AbortController().signal,
	});

	assert.equal(result.status, "applied");
	if (result.status !== "applied") return;
	assert.equal(calls.length, 1);
	assert.equal((calls[0] as { ownershipId: string }).ownershipId, ownershipId);
	assert.equal(result.projectedState.children[0].checkpoint?.workspace?.ownershipId, ownershipId);
	assert.equal(result.projectedState.children[0].checkpoint?.effectKey, DIGEST_B);
});

test("agent recovery distinguishes absent, incomplete, and exact completed receipts", async () => {
	const state = recoveryState();
	const workspace = state.children[0].ownership!;
	const descriptor = {
		operation: "agent_implementation",
		childId: "runtime",
		childIssue: 479,
		childSlug: "runtime-composition",
		generation: 1,
		parentBaseBranch: "main",
		parentBranch: "feat/471-parent",
		parentIssue: 479,
		planDigest: state.planDigest,
		repository: "acme/widgets",
		workspace,
		attempt: 1,
		corrections: 0,
		taskDigest: DIGEST_A,
	};
	const request = {
		record: recoveryRecord("agent_implementation", descriptor),
		descriptor,
		currentState: state,
		signal: new AbortController().signal,
	};
	const absent = createProductionRecoveryProbeTable(recoveryProbeOptions({
		agentEffects: { async find() { return undefined; } },
	}));
	assert.deepEqual(await absent.agent_implementation(request), { status: "absent" });

	const start = {
		schemaVersion: 1 as const,
		effectKey: DIGEST_B,
		claimId: workspace.claimId,
		role: "implementation" as const,
		binding: workspace,
	};
	const incomplete = createProductionRecoveryProbeTable(recoveryProbeOptions({
		agentEffects: { async find() { return { start }; } },
	}));
	await assert.rejects(incomplete.agent_implementation(request), /lacks an exact completion|ambiguous/i);

	const value = { workspace };
	const complete = createProductionRecoveryProbeTable(recoveryProbeOptions({
		agentEffects: { async find() { return {
			start,
			completion: { ...start, resultDigest: stableDigest(value), completedBinding: workspace },
		}; } },
	}));
	const applied = await complete.agent_implementation(request);
	assert.equal(applied.status, "applied");
	if (applied.status !== "applied") return;
	assert.equal(applied.resultDigest, stableDigest(value));
	assert.equal(applied.projectedState.children[0].checkpoint?.effectKey, DIGEST_B);
});

test("observed no-op commit projects once while a merely prepared unchanged commit remains absent", async () => {
	const state = recoveryState();
	const workspace = state.children[0].ownership!;
	const descriptor = {
		operation: "git_commit",
		childId: "runtime",
		childIssue: 479,
		childSlug: "runtime-composition",
		generation: 1,
		parentBaseBranch: "main",
		parentBranch: "feat/471-parent",
		parentIssue: 479,
		planDigest: state.planDigest,
		repository: "acme/widgets",
		workspace,
		issue: 479,
		slug: "runtime-composition",
		message: "feat(shepherd): complete #479 runtime-composition",
		attempt: 1,
		corrections: 0,
	};
	const probes = createProductionRecoveryProbeTable(recoveryProbeOptions({
		git: {
			async assertBinding(value) { return value; },
			async inspect() { return coordinator(workspace.cwd); },
			async currentBranch() { return workspace.branch; },
			async resolveBranchHead() { return workspace.head; },
			async resolveRemoteBranchHead() { return undefined; },
			async readCommitSubject() { return unexpected("commit subject"); },
			async isAncestor() { return true; },
			async status() { return { clean: true, entries: [] }; },
			async diff() { return { baseHead: SHA_A, head: SHA_B, changedScope: [] }; },
		},
	}));
	const prepared = recoveryRecord("git_commit", descriptor);
	assert.deepEqual(await probes.git_commit({
		record: prepared,
		descriptor,
		currentState: state,
		signal: new AbortController().signal,
	}), { status: "absent" });
	const value = { committed: false, previousHead: workspace.head, head: workspace.head };
	const observed = { ...prepared, phase: "observed" as const, resultDigest: stableDigest(value), observedAt: "2026-07-22T10:00:00.750Z" };
	const applied = await probes.git_commit({
		record: observed,
		descriptor,
		currentState: state,
		signal: new AbortController().signal,
	});
	assert.equal(applied.status, "applied");
	if (applied.status !== "applied") return;
	assert.equal(applied.resultDigest, stableDigest(value));
	assert.equal(applied.projectedState.children[0].checkpoint?.effectKey, DIGEST_B);
});

test("shuffled publication recovery projects PR, push, and commit without regressing exact head truth", async () => {
	const state = recoveryState();
	const before = state.children[0].ownership!;
	const after = { ...before, head: SHA_C };
	const marker = "<!-- shepherd-child-pr:v1:479:runtime:0123456789abcdef01234567 -->";
	const common = {
		childId: "runtime",
		childIssue: 479,
		childSlug: "runtime-composition",
		generation: 1,
		parentBaseBranch: "main",
		parentBranch: "feat/471-parent",
		parentIssue: 479,
		planDigest: state.planDigest,
		repository: "acme/widgets",
	};
	const pullDescriptor = {
		...common,
		operation: "child_pull_request",
		branch: after.branch,
		baseBranch: after.baseBranch,
		baseHead: after.baseHead,
		head: after.head,
		changedScope: [".pi/extensions/shepherd/production-runtime.ts"],
		marker,
	};
	const pullEvidence = {
		schemaVersion: 2 as const,
		repository: "acme/widgets",
		workItemId: "runtime",
		generation: 1,
		number: 47,
		marker,
		title: "Runtime recovery",
		body: `Runtime recovery\n${marker}`,
		state: "open" as const,
		draft: false,
		baseBranch: after.baseBranch,
		headBranch: after.branch,
		baseSha: after.baseHead,
		headSha: after.head,
		changedPathsComplete: true,
		changedPaths: [".pi/extensions/shepherd/production-runtime.ts"],
		allowedScopes: [".pi/extensions/shepherd/production-runtime.ts"],
		mergeState: "clean" as const,
		policyDigest: DIGEST_A,
		checksComplete: true,
		checks: [],
		requestedChangesComplete: true,
		requestedChanges: [],
		threadsComplete: true,
		threads: [],
		reviews: [],
		reviewsComplete: true,
		dispositionsComplete: true,
		dispositions: [],
		revision: 1,
		observedAt: "2026-07-22T10:00:00.000Z",
	};
	const commitKey = "5".repeat(64);
	const pullRecord = { ...recoveryRecord("child_pull_request", pullDescriptor), key: "3".repeat(64) };
	let wrongLocalHead = true;
	let localHeadReads = 0;
	let remoteHeadReads = 0;
	const probes = createProductionRecoveryProbeTable(recoveryProbeOptions({
		git: {
			async assertBinding(value) { return value; },
			async inspect() { return coordinator(before.cwd); },
			async currentBranch() { return before.branch; },
			async resolveBranchHead() { localHeadReads += 1; return wrongLocalHead ? SHA_B : SHA_C; },
			async resolveRemoteBranchHead() { remoteHeadReads += 1; return SHA_C; },
			async readCommitSubject() {
				return `feat(shepherd): complete #479 runtime-composition [shepherd-effect:${commitKey}]`;
			},
			async isAncestor() { return true; },
			async status() { return { clean: true, entries: [] }; },
			async diff() { return { baseHead: SHA_A, head: SHA_C, changedScope: [".pi/extensions/shepherd/production-runtime.ts"] }; },
		},
		github: {
			async findPullRequests() { return { items: [pullEvidence], complete: true }; },
			async findChildIntegration() { return unexpected("child integration"); },
			async proveAncestry() { return unexpected("ancestry"); },
		},
	}));
	await assert.rejects(probes.child_pull_request({
		record: pullRecord,
		descriptor: pullDescriptor,
		currentState: state,
		signal: new AbortController().signal,
	}), /local or remote branch moved|publication head/i);
	assert.equal(remoteHeadReads, 0, "wrong local head fails before trusting the remote or PR");
	wrongLocalHead = false;
	const pull = await probes.child_pull_request({
		record: pullRecord,
		descriptor: pullDescriptor,
		currentState: state,
		signal: new AbortController().signal,
	});
	assert.equal(pull.status, "applied");
	if (pull.status !== "applied") return;
	assert.ok(localHeadReads >= 2);
	assert.equal(remoteHeadReads, 1);
	assert.equal(pull.projectedState.children[0].ownership?.head, SHA_C);

	const pushDescriptor = { ...common, operation: "git_push", branch: after.branch, head: after.head, workspace: after };
	const pushRecord = { ...recoveryRecord("git_push", pushDescriptor), key: "4".repeat(64) };
	const push = await probes.git_push({
		record: pushRecord,
		descriptor: pushDescriptor,
		currentState: pull.projectedState,
		signal: new AbortController().signal,
	});
	assert.equal(push.status, "applied");
	if (push.status !== "applied") return;

	const commitDescriptor = {
		...common,
		operation: "git_commit",
		workspace: before,
		issue: 479,
		slug: "runtime-composition",
		message: "feat(shepherd): complete #479 runtime-composition",
		attempt: 1,
		corrections: 0,
	};
	const commitRecord = { ...recoveryRecord("git_commit", commitDescriptor), key: commitKey };
	const commit = await probes.git_commit({
		record: commitRecord,
		descriptor: commitDescriptor,
		currentState: push.projectedState,
		signal: new AbortController().signal,
	});
	assert.equal(commit.status, "applied");
	if (commit.status !== "applied") return;
	assert.equal(commit.projectedState.children[0].ownership?.head, SHA_C);
	assert.equal(commit.projectedState.children[0].stage, "review");
	assert.deepEqual(new Set(commit.projectedState.children[0].checkpoint?.effectKeys), new Set([
		pullRecord.key,
		pushRecord.key,
		commitRecord.key,
	]));
});

test("default production recovery probes are exhaustive and fail closed on absent or ambiguous exact review evidence", async () => {
	const state = recoveryState();
	const descriptor = {
		operation: "independent_review",
		childId: "runtime",
		childIssue: 479,
		childSlug: "runtime-composition",
		generation: 1,
		parentBaseBranch: "main",
		parentBranch: "feat/471-parent",
		parentIssue: 479,
		planDigest: state.planDigest,
		repository: "acme/widgets",
		target: {
			repository: "acme/widgets",
			workItemId: "runtime",
			pullRequest: 47,
			generation: 1,
			baseBranch: "feat/471-parent",
			headBranch: "feat/479-runtime-composition",
			baseSha: SHA_A,
			headSha: SHA_B,
			changedPaths: [".pi/extensions/shepherd/production-runtime.ts"],
			allowedScopes: [".pi/extensions/shepherd/production-runtime.ts"],
		},
	};
	const request = {
		record: recoveryRecord("independent_review", descriptor),
		descriptor,
		currentState: state,
		signal: new AbortController().signal,
	};
	const absent = createProductionRecoveryProbeTable(recoveryProbeOptions({
		reviews: { async find() { return { items: [], complete: true }; } },
	}));
	assert.deepEqual(Object.keys(absent).sort(), [...EFFECT_KINDS].sort());
	assert.deepEqual(await absent.independent_review(request), { status: "absent" });

	const work = createIndependentReviewWork(descriptor.target);
	const cleanReview = {
		...work,
		completedAt: "2026-07-22T10:00:00.750Z",
		verdict: "clean" as const,
		findings: [],
	};
	const blockedReview = {
		...work,
		completedAt: cleanReview.completedAt,
		verdict: "findings" as const,
		findings: [{ id: "blocking-1", severity: "blocking" as const, summary: "Conflicting exact result." }],
	};
	const ambiguous = createProductionRecoveryProbeTable(recoveryProbeOptions({
		reviews: { async find() { return { items: [{
			schemaVersion: 1,
			review: cleanReview,
			attestation: createAgentSessionAttestation({ sessionId: "review-clean", runId: "run-clean", review: cleanReview }),
			dispositions: [],
			revision: 1,
			publishedAt: "2026-07-22T10:00:00.800Z",
		}, {
			schemaVersion: 1,
			review: blockedReview,
			attestation: createAgentSessionAttestation({ sessionId: "review-blocked", runId: "run-blocked", review: blockedReview }),
			dispositions: [],
			revision: 2,
			publishedAt: "2026-07-22T10:00:00.900Z",
		}], complete: true }; } },
	}));
	await assert.rejects(ambiguous.independent_review(request), /ambiguous/i);
});

test("default shell recovery reruns only the bounded read-only commands and projects their stable checkpoint", async () => {
	const state = recoveryState();
	const workspace = state.children[0].ownership!;
	const commands = plan().children[0].verification;
	const descriptor = {
		operation: "shell_verification",
		childId: "runtime",
		childIssue: 479,
		childSlug: "runtime-composition",
		generation: 1,
		parentBaseBranch: "main",
		parentBranch: "feat/471-parent",
		parentIssue: 479,
		planDigest: state.planDigest,
		repository: "acme/widgets",
		workspace,
		attempt: 1,
		corrections: 0,
		commands,
	};
	const calls: Array<{ cwd: string; commands: unknown }> = [];
	const probes = createProductionRecoveryProbeTable(recoveryProbeOptions({
		verification: {
			async runAll(cwd, value) {
				calls.push({ cwd, commands: value });
				return [{
					id: "focused-test",
					status: "passed",
					exitCode: 0,
					signal: null,
					stdout: "ignored",
					stderr: "",
					durationMs: 7,
				}];
			},
		},
	}));
	const evidence = await probes.shell_verification({
		record: recoveryRecord("shell_verification", descriptor),
		descriptor,
		currentState: state,
		signal: new AbortController().signal,
	});

	assert.equal(evidence.status, "applied");
	if (evidence.status !== "applied") return;
	assert.deepEqual(calls, [{ cwd: workspace.cwd, commands }]);
	assert.equal(evidence.projectedState.revision, state.revision + 1);
	assert.equal(evidence.projectedState.children[0].checkpoint?.effectKey, DIGEST_B);
	assert.deepEqual(evidence.projectedState.children[0].checkpoint?.verification?.commands, [{
		id: "focused-test",
		status: "passed",
	}]);
	assert.equal(evidence.projectedState.children[0].checkpoint?.verification?.status, "passed");
});

test("production factory synchronously selects durable adapters and constructs genuine recovery by default", async () => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-runtime-"));
	const git = new GitAdapter({ execute: async () => Buffer.from("") });
	const effectRecovery: ProductionEffectRecoveryPort = {
		async observe() { return { status: "applied", resultDigest: DIGEST_A }; },
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
		parentIssue: 479,
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

	const defaultRecovery = createProductionShepherdController({ ...options, effectRecovery: undefined });
	assert.ok(defaultRecovery instanceof ProductionShepherdController);
	await defaultRecovery.shutdown();
	assert.throws(() => createProductionShepherdController({ ...options, parentIssue: 0 }), /parent issue/);
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
		recoveryDescriptor: { kind: "child_pull_request", marker: "runtime-crash-effect" },
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
			return { status: "applied", resultDigest: DIGEST_B };
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
			prepare() { return { requestId: "parent-gate-479" }; },
			async request() { return { requestId: "parent-gate-479" }; },
			async observe() { return { status: "pending" as const }; },
			async close() {},
		},
		parentMergeEffects: {
			async observe() { return unexpected("parent merge effect observation"); },
			async acknowledge() {},
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
