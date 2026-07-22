import assert from "node:assert/strict";
import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import {
	DurableGhParentReadiness,
	ProductionPiEntrypointController,
	createProductionPiHostController,
} from "./production-pi-host.ts";
import { GitAdapter, type GitBinding } from "./git-adapter.ts";

function productionState(issue: number, status = "waiting_human") {
	return {
		schemaVersion: 1 as const,
		kind: "production_autonomous" as const,
		parentIssue: issue,
		repository: "acme/widgets",
		planId: "production-plan",
		planDigest: "d".repeat(64),
		parentBranch: "feat/471-parent",
		parentBaseBranch: "main",
		runId: "run-production",
		resourceGeneration: 1,
		generation: 1,
		revision: 1,
		maxConcurrency: 2,
		timeoutMs: 30_000,
		status,
		stage: status === "waiting_human" ? "human_decision" : "schedule",
		createdAt: "2026-07-22T08:00:00.000Z",
		updatedAt: "2026-07-22T08:01:00.000Z",
		children: [],
	};
}

function delegate(issue: number, calls: string[]) {
	let persisted: ReturnType<typeof productionState> | undefined;
	return {
		async status() { calls.push("status"); return persisted; },
		async start() { calls.push("start"); persisted = productionState(issue); return persisted; },
		async resume() { calls.push("resume"); persisted = productionState(issue); return persisted; },
		async stop() { calls.push("stop"); persisted = productionState(issue, "stopped"); return persisted; },
		async shutdown() { calls.push("shutdown"); },
	};
}

async function writeProductionPlan(
	root: string,
	options: { parentBranch?: string; parentBaseBranch?: string } = {},
): Promise<void> {
	const directory = join(root, ".planning", "shepherd");
	await mkdir(directory, { recursive: true });
	await writeFile(join(directory, "issue-479.json"), JSON.stringify({
		schemaVersion: 2,
		planId: "production-plan",
		parentIssue: 479,
		repository: "acme/widgets",
		title: "Production plan",
		objective: "Prove the production launch authority boundary.",
		parentBranch: options.parentBranch ?? "feat/471-parent",
		parentBaseBranch: options.parentBaseBranch ?? "main",
		actorAllowlist: ["maintainer"],
		decisionExpiresAt: "2099-07-22T00:00:00.000Z",
		children: [{
			id: "child-a",
			issue: 480,
			title: "Child A",
			task: "Implement child A.",
			slug: "child-a",
			dependsOn: [],
			access: "mutating",
			writeScopes: [".pi/extensions/shepherd"],
			requiredSkills: ["javascript-testing-patterns"],
			verification: [{
				id: "focused",
				executable: "node",
				args: ["--test"],
				cwd: ".",
				timeoutMs: 30_000,
				maxOutputBytes: 65_536,
			}],
			humanGates: [],
			maxAttempts: 2,
			maxCorrections: 1,
		}],
	}), { mode: 0o600 });
}

test("entrypoint ensures one exact marker-bound parent draft before starting production work", async () => {
	const calls: string[] = [];
	const controller = new ProductionPiEntrypointController({
		issue: 479,
		delegate: delegate(479, calls),
		async ensurePlan() { calls.push("ensure-plan"); },
		async validateAuthority() { calls.push("validate-authority"); },
		async ensureParentDraft(issue, signal) {
			assert.equal(issue, 479);
			assert.equal(signal.aborted, false);
			calls.push("ensure-parent-draft");
		},
		resources: [],
	});

	const result = await controller.start({
		action: "start",
		issue: 479,
		backend: "sdk-inproc",
		maxConcurrency: 2,
		timeoutMs: 30_000,
	});
	assert.equal(result.kind, "production_autonomous");
	assert.deepEqual(calls, ["status", "ensure-plan", "validate-authority", "ensure-parent-draft", "start"]);

	await controller.resume({
		action: "resume",
		issue: 479,
		backend: "sdk-inproc",
		maxConcurrency: 2,
		timeoutMs: 30_000,
	});
	assert.deepEqual(calls, [
		"status", "ensure-plan", "validate-authority", "ensure-parent-draft", "start",
		"validate-authority", "resume",
	]);
	await controller.shutdown();
});

test("entrypoint fails closed before production state when parent draft preparation fails", async () => {
	const calls: string[] = [];
	const controller = new ProductionPiEntrypointController({
		issue: 479,
		delegate: delegate(479, calls),
		async ensurePlan() {},
		async validateAuthority() {},
		async ensureParentDraft() { throw new Error("parent draft evidence is ambiguous"); },
		resources: [],
	});
	await assert.rejects(controller.start({
		action: "start",
		issue: 479,
		backend: "sdk-inproc",
		maxConcurrency: 2,
		timeoutMs: 30_000,
	}), /parent draft evidence is ambiguous/);
	assert.deepEqual(calls, ["status"]);
	assert.equal(await controller.status(479), undefined);
});

test("entrypoint plan bootstrap failure creates no state and performs no authority or draft mutation", async () => {
	const calls: string[] = [];
	const controller = new ProductionPiEntrypointController({
		issue: 479,
		delegate: delegate(479, calls),
		async ensurePlan() {
			calls.push("ensure-plan");
			throw new Error("planning issue facts are ambiguous");
		},
		async validateAuthority() { calls.push("validate-authority"); },
		async ensureParentDraft() { calls.push("ensure-parent-draft"); },
		resources: [],
	});
	await assert.rejects(controller.start({
		action: "start",
		issue: 479,
		backend: "sdk-inproc",
		maxConcurrency: 2,
		timeoutMs: 30_000,
	}), /planning issue facts are ambiguous/);
	assert.deepEqual(calls, ["status", "ensure-plan"]);
	assert.equal(await controller.status(479), undefined);
});

test("production host revalidates the live remote default on resume before delegate mutation", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "production-pi-host-default-move-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	await writeProductionPlan(root);
	const calls: string[] = [];
	let inspections = 0;
	const binding = (defaultBranch: string): GitBinding => ({
		cwd: root,
		repositoryIdentity: "1".repeat(64),
		worktreeIdentity: "2".repeat(64),
		remoteName: "origin",
		remoteIdentity: "3".repeat(64),
		fetchEndpointIdentity: "4".repeat(64),
		pushEndpointIdentity: "4".repeat(64),
		defaultBranch,
	});
	const controller = await createProductionPiHostController({
		issue: 479,
		repositoryRoot: root,
		stateRoot: join(root, "state"),
		trustedWorktreeRoot: join(root, "worktrees"),
		runtimeSdk: {} as never,
		dependencies: {
			git: new GitAdapter({ execute: async () => Buffer.from("") }),
			async inspectCoordinator() {
				inspections += 1;
				return binding(inspections < 3 ? "main" : "trunk");
			},
			createAgentRuntime: () => ({ async run() { throw new Error("not dispatched"); }, async abort() {}, async close() {} }) as never,
			createReviewSession: () => ({ async run() { throw new Error("not dispatched"); } }),
			createController: () => delegate(479, calls) as never,
			createParentReadyAuthority: () => ({
				async readParentReadyState() { return null; },
				async beginParentReady() { throw new Error("not dispatched"); },
				async compareConsumeAndMarkParentReady() { throw new Error("not dispatched"); },
				async settleParentReady() { throw new Error("not dispatched"); },
				async quarantineAndRollbackParentReady() { throw new Error("not dispatched"); },
			}),
			createParentReadiness: () => ({ async markExistingDraftReady() { throw new Error("not dispatched"); } }),
			createParentDraftEnsurer: () => async () => { calls.push("ensure-parent-draft"); },
		},
	});
	await controller.start({ action: "start", issue: 479, backend: "sdk-inproc", maxConcurrency: 2, timeoutMs: 30_000 });
	await assert.rejects(
		controller.resume({ action: "resume", issue: 479, backend: "sdk-inproc", maxConcurrency: 2, timeoutMs: 30_000 }),
		/default branch/i,
	);
	assert.equal(calls.filter((entry) => entry === "resume").length, 0);
	assert.equal(inspections, 3);
	await controller.shutdown();
});

test("production host rejects a conventional default alias as the parent integration target", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "production-pi-host-main-target-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	await writeProductionPlan(root, { parentBaseBranch: "trunk", parentBranch: "main" });
	const calls: string[] = [];
	const coordinator: GitBinding = {
		cwd: root,
		repositoryIdentity: "1".repeat(64),
		worktreeIdentity: "2".repeat(64),
		remoteName: "origin",
		remoteIdentity: "3".repeat(64),
		fetchEndpointIdentity: "4".repeat(64),
		pushEndpointIdentity: "4".repeat(64),
		defaultBranch: "trunk",
	};
	const controller = await createProductionPiHostController({
		issue: 479,
		repositoryRoot: root,
		stateRoot: join(root, "state"),
		trustedWorktreeRoot: join(root, "worktrees"),
		runtimeSdk: {} as never,
		dependencies: {
			git: new GitAdapter({ execute: async () => Buffer.from("") }),
			async inspectCoordinator() { return coordinator; },
			createAgentRuntime: () => ({ async run() { throw new Error("not dispatched"); }, async abort() {}, async close() {} }) as never,
			createReviewSession: () => ({ async run() { throw new Error("not dispatched"); } }),
			createController: () => delegate(479, calls) as never,
			createParentReadyAuthority: () => ({
				async readParentReadyState() { return null; },
				async beginParentReady() { throw new Error("not dispatched"); },
				async compareConsumeAndMarkParentReady() { throw new Error("not dispatched"); },
				async settleParentReady() { throw new Error("not dispatched"); },
				async quarantineAndRollbackParentReady() { throw new Error("not dispatched"); },
			}),
			createParentReadiness: () => ({ async markExistingDraftReady() { throw new Error("not dispatched"); } }),
			createParentDraftEnsurer: () => async () => { calls.push("ensure-parent-draft"); },
		},
	});
	await assert.rejects(
		controller.start({ action: "start", issue: 479, backend: "sdk-inproc", maxConcurrency: 2, timeoutMs: 30_000 }),
		/default branch|parent integration target/i,
	);
	assert.equal(calls.includes("start"), false);
	assert.equal(calls.includes("ensure-parent-draft"), false);
	await controller.shutdown();
});

test("production Pi host composes separate implementation, review, and planning AgentSession runtimes", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "production-pi-host-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	await writeProductionPlan(root);
	const coordinator: GitBinding = {
		cwd: root,
		repositoryIdentity: "1".repeat(64),
		worktreeIdentity: "2".repeat(64),
		remoteName: "origin",
		remoteIdentity: "3".repeat(64),
		fetchEndpointIdentity: "4".repeat(64),
		pushEndpointIdentity: "4".repeat(64),
		defaultBranch: "main",
	};
	const git = new GitAdapter({ execute: async () => Buffer.from("") });
	const runtimeRoles: string[] = [];
	const captured: Record<string, unknown>[] = [];
	const calls: string[] = [];
	const controller = await createProductionPiHostController({
		issue: 479,
		repositoryRoot: root,
		stateRoot: join(root, "state"),
		trustedWorktreeRoot: join(root, "worktrees"),
		runtimeSdk: {} as never,
		dependencies: {
			git,
			async inspectCoordinator() { return coordinator; },
			createAgentRuntime(role) {
				runtimeRoles.push(role);
				return {
					async run() { throw new Error("not dispatched in construction test"); },
					async abort() {},
					async close() { calls.push(`close-${role}`); },
				} as never;
			},
			createReviewSession(_runtime, requestFactory) {
				assert.equal(typeof requestFactory, "function");
				return { async run() { throw new Error("not dispatched in construction test"); } };
			},
			createController(options) {
				captured.push(options as unknown as Record<string, unknown>);
				return delegate(479, calls) as never;
			},
			createParentReadyAuthority() {
				return {
					async readParentReadyState() { return null; },
					async beginParentReady() { throw new Error("not dispatched"); },
					async compareConsumeAndMarkParentReady() { throw new Error("not dispatched"); },
					async settleParentReady() { throw new Error("not dispatched"); },
					async quarantineAndRollbackParentReady() { throw new Error("not dispatched"); },
				};
			},
			createParentReadiness() {
				return { async markExistingDraftReady() { throw new Error("not dispatched"); } };
			},
			createParentDraftEnsurer() {
				return async () => { calls.push("ensure-parent-draft"); };
			},
		},
	});

	assert.deepEqual(runtimeRoles, ["implementation", "review", "planning"]);
	assert.equal(captured.length, 1);
	assert.equal(captured[0].parentIssue, 479);
	assert.equal(captured[0].coordinator, coordinator);
	assert.equal(captured[0].git, git);
	assert.equal(Object.hasOwn(captured[0], "effectRecovery"), false);
	await controller.start({
		action: "start",
		issue: 479,
		backend: "sdk-inproc",
		maxConcurrency: 2,
		timeoutMs: 30_000,
	});
	assert.ok(calls.includes("ensure-parent-draft"));
	await controller.shutdown();
	assert.ok(calls.includes("close-implementation"));
	assert.ok(calls.includes("close-review"));
	assert.ok(calls.includes("close-planning"));
});

test("durable parent readiness reconciles timeout-after-apply and replays no mutation after restart", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "production-parent-ready-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const initialTime = "2026-07-22T10:00:00.000Z";
	const appliedTime = "2026-07-22T10:00:02.000Z";
	const expectedRevision = Math.floor(new Date(initialTime).valueOf() / 1_000);
	const calls: string[][] = [];
	let read = 0;
	const execute = async (_file, args) => {
		calls.push([...args]);
		if (args[1] === "graphql") throw new Error("response lost after GraphQL apply");
		read += 1;
		return JSON.stringify({
			number: 438,
			state: "open",
			draft: read === 1,
			node_id: "PR_kwDOProduction",
			updated_at: read === 1 ? initialTime : appliedTime,
			head: { ref: "feat/471-parent", sha: "a".repeat(40) },
		});
	};
	const request = {
		repository: "acme/widgets",
		parentIssue: 479,
		pullRequest: 438,
		generation: 1,
		branch: "feat/471-parent",
		headSha: "a".repeat(40),
		expectedRevision,
	};
	const context = {
		signal: new AbortController().signal,
		deadlineAt: "2099-07-22T10:10:00.000Z",
		acknowledgeAbort() {},
	};
	const first = new DurableGhParentReadiness(root, {
		execute,
		now: () => new Date("2026-07-22T10:00:03.000Z"),
	});
	const receipt = await first.markExistingDraftReady(request, context);
	assert.equal(receipt.appliedRevision, expectedRevision + 2);
	assert.equal(receipt.operation, "existing_draft_to_ready");
	assert.equal(calls.filter((args) => args[1] === "graphql").length, 1);
	assert.equal(calls.some((args) => args.includes("merge")), false);

	const restarted = new DurableGhParentReadiness(root, { execute });
	assert.deepEqual(await restarted.markExistingDraftReady(request, context), receipt);
	assert.equal(calls.length, 3);
});

test("durable parent readiness accepts an exact non-draft observation in the expected revision second", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "production-parent-ready-same-second-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const updatedAt = "2026-07-22T10:00:00.000Z";
	const expectedRevision = Math.floor(new Date(updatedAt).valueOf() / 1_000);
	const calls: string[][] = [];
	const adapter = new DurableGhParentReadiness(root, {
		execute: async (_file, args) => {
			calls.push([...args]);
			assert.notEqual(args[1], "graphql");
			return JSON.stringify({
				number: 438,
				state: "open",
				draft: false,
				node_id: "PR_kwDOSameSecond",
				updated_at: updatedAt,
				head: { ref: "feat/471-parent", sha: "a".repeat(40) },
			});
		},
	});
	const receipt = await adapter.markExistingDraftReady({
		repository: "acme/widgets",
		parentIssue: 479,
		pullRequest: 438,
		generation: 1,
		branch: "feat/471-parent",
		headSha: "a".repeat(40),
		expectedRevision,
	}, {
		signal: new AbortController().signal,
		deadlineAt: "2099-07-22T10:10:00.000Z",
		acknowledgeAbort() {},
	});
	assert.equal(receipt.appliedRevision, expectedRevision);
	assert.equal(calls.length, 1);
});

test("durable parent readiness never accepts an equal revision while the exact parent stays draft", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "production-parent-ready-still-draft-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const updatedAt = "2026-07-22T10:00:00.000Z";
	const expectedRevision = Math.floor(new Date(updatedAt).valueOf() / 1_000);
	let reads = 0;
	let mutations = 0;
	const adapter = new DurableGhParentReadiness(root, {
		execute: async (_file, args) => {
			if (args[1] === "graphql") {
				mutations += 1;
				return "{}";
			}
			reads += 1;
			return JSON.stringify({
				number: 438,
				state: "open",
				draft: true,
				node_id: "PR_kwDOStillDraft",
				updated_at: updatedAt,
				head: { ref: "feat/471-parent", sha: "a".repeat(40) },
			});
		},
	});
	await assert.rejects(adapter.markExistingDraftReady({
		repository: "acme/widgets",
		parentIssue: 479,
		pullRequest: 438,
		generation: 1,
		branch: "feat/471-parent",
		headSha: "a".repeat(40),
		expectedRevision,
	}, {
		signal: new AbortController().signal,
		deadlineAt: "2099-07-22T10:10:00.000Z",
		acknowledgeAbort() {},
	}), /not authoritatively marked ready/);
	assert.equal(reads, 2);
	assert.equal(mutations, 1);
});

test("durable parent readiness reconciles a prepared same-second restart and duplicate without another mutation", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "production-parent-ready-prepared-restart-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const updatedAt = "2026-07-22T10:00:00.000Z";
	const expectedRevision = Math.floor(new Date(updatedAt).valueOf() / 1_000);
	const request = {
		repository: "acme/widgets",
		parentIssue: 479,
		pullRequest: 438,
		generation: 1,
		branch: "feat/471-parent",
		headSha: "a".repeat(40),
		expectedRevision,
	};
	const context = {
		signal: new AbortController().signal,
		deadlineAt: "2099-07-22T10:10:00.000Z",
		acknowledgeAbort() {},
	};
	let firstReads = 0;
	let mutations = 0;
	const first = new DurableGhParentReadiness(root, {
		execute: async (_file, args) => {
			if (args[1] === "graphql") {
				mutations += 1;
				throw new Error("response lost while parent still appeared draft");
			}
			firstReads += 1;
			return JSON.stringify({
				number: 438,
				state: "open",
				draft: true,
				node_id: "PR_kwDOPreparedRestart",
				updated_at: updatedAt,
				head: { ref: "feat/471-parent", sha: "a".repeat(40) },
			});
		},
	});
	await assert.rejects(first.markExistingDraftReady(request, context), /response lost/);
	assert.equal(firstReads, 2);
	assert.equal(mutations, 1);

	let restartReads = 0;
	const restarted = new DurableGhParentReadiness(root, {
		execute: async (_file, args) => {
			assert.notEqual(args[1], "graphql");
			restartReads += 1;
			return JSON.stringify({
				number: 438,
				state: "open",
				draft: false,
				node_id: "PR_kwDOPreparedRestart",
				updated_at: updatedAt,
				head: { ref: "feat/471-parent", sha: "a".repeat(40) },
			});
		},
	});
	const reconciled = await restarted.markExistingDraftReady(request, context);
	assert.equal(reconciled.appliedRevision, expectedRevision);
	assert.deepEqual(await restarted.markExistingDraftReady(request, context), reconciled);
	assert.equal(restartReads, 1);
	assert.equal(mutations, 1);
});

test("durable parent readiness rejects a moved exact head before mutation", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "production-parent-ready-moved-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	let calls = 0;
	const adapter = new DurableGhParentReadiness(root, {
		execute: async () => {
			calls += 1;
			return JSON.stringify({
				number: 438,
				state: "open",
				draft: true,
				node_id: "PR_kwDOMoved",
				updated_at: "2026-07-22T10:00:00.000Z",
				head: { ref: "feat/471-parent", sha: "b".repeat(40) },
			});
		},
	});
	await assert.rejects(adapter.markExistingDraftReady({
		repository: "acme/widgets",
		parentIssue: 479,
		pullRequest: 438,
		generation: 1,
		branch: "feat/471-parent",
		headSha: "a".repeat(40),
		expectedRevision: Math.floor(new Date("2026-07-22T10:00:00.000Z").valueOf() / 1_000),
	}, {
		signal: new AbortController().signal,
		deadlineAt: "2099-07-22T10:10:00.000Z",
		acknowledgeAbort() {},
	}), /moved from the exact/);
	assert.equal(calls, 1);
});
