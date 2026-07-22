import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { mkdtemp, rm } from "node:fs/promises";
import { join } from "node:path";
import { tmpdir } from "node:os";
import test from "node:test";

import type { AgentSessionHandoff, RoleRunRequest } from "./agent-session-runtime.ts";
import type { ProductionChildSpec } from "./autonomous-production-contract.ts";
import { GitAdapter, type GitCommandExecutor } from "./git-adapter.ts";
import { createLocalGitFixture, git, write } from "./issue-476-git-fixture.ts";
import {
	ProductionWorkspaceLifecycle,
	productionWorkspaceOwnershipId,
	type ProductionAgentSessionPort,
	type ProductionParentRefreshEvidence,
	type ProductionWorkspaceAdapterPort,
} from "./production-workspace-lifecycle.ts";
import type { ClaimedWorkspace, WorkspaceClaimRequest } from "./workspace-adapter.ts";
import { WorkspaceAdapter } from "./workspace-adapter.ts";

const SHA_A = "a".repeat(40);
const SHA_B = "b".repeat(40);
const SHA_C = "c".repeat(40);
const ID_A = "1".repeat(64);
const ID_B = "2".repeat(64);

function child(id = "lane-b", issue = 502): ProductionChildSpec {
	return {
		id,
		issue,
		title: "Workspace lifecycle",
		task: "Implement the bounded workspace lifecycle",
		slug: `${id}-workspace`,
		dependsOn: [],
		access: "mutating",
		writeScopes: [`.pi/extensions/shepherd/${id}`],
		requiredSkills: ["javascript-testing-patterns"],
		verification: [{
			id: "focused",
			executable: "node",
			args: ["--test", `${id}.test.ts`],
			cwd: ".",
			timeoutMs: 30_000,
			maxOutputBytes: 1_048_576,
		}],
		humanGates: [],
		maxAttempts: 2,
		maxCorrections: 1,
	};
}

function handoff(request: RoleRunRequest): AgentSessionHandoff {
	return {
		schemaVersion: 1,
		role: request.role,
		status: "completed",
		summary: "implemented",
		observedMutation: true,
		changedPaths: [...request.authority.writePrefixes],
		verification: [],
		findings: [],
		...request.binding,
	};
}

class FakeAgent implements ProductionAgentSessionPort {
	requests: RoleRunRequest[] = [];
	aborted: string[] = [];
	closed = 0;
	gate: Promise<void> | undefined;
	abortFailure: Error | undefined;

	async run(request: RoleRunRequest): Promise<AgentSessionHandoff> {
		this.requests.push(request);
		await this.gate;
		return handoff(request);
	}
	async abort(runId: string): Promise<void> {
		this.aborted.push(runId);
		if (this.abortFailure) throw this.abortFailure;
	}
	async close(): Promise<void> { this.closed += 1; }
}

class FakeWorkspace implements ProductionWorkspaceAdapterPort {
	claims: WorkspaceClaimRequest[] = [];
	released: number[] = [];
	commits = 0;
	pushes = 0;
	handoffs = 0;
	refreshes = 0;

	async claim(request: WorkspaceClaimRequest): Promise<ClaimedWorkspace> {
		this.claims.push(request);
		const index = this.claims.length;
		let released = false;
		return {
			...request.coordinator,
			cwd: join(request.trustedWorktreeRoot, `issue-${request.issue}-${request.slug}`),
			worktreeIdentity: `${index}`.repeat(64).slice(0, 64),
			issue: request.issue,
			slug: request.slug,
			branch: `feat/${request.issue}-${request.slug}`,
			prBase: request.parentBranch,
			baseHead: request.parentHead,
			head: request.parentHead,
			trustedWorktreeRoot: request.trustedWorktreeRoot,
			allowedScopes: [...request.allowedScopes],
			claimId: `${index + 5}`.repeat(64).slice(0, 64),
			reused: request.leaseMode === "resume",
			status: { clean: true, entries: [] },
			changedScope: [],
			assertOwned: async () => { if (released) throw new Error("released"); },
			release: async () => { if (!released) this.released.push(request.issue); released = true; },
		};
	}

	async commitIssueChanges(workspace: ClaimedWorkspace) {
		this.commits += 1;
		const previousHead = workspace.head;
		workspace.head = SHA_B;
		return { committed: true, previousHead, head: workspace.head };
	}
	async pushIssueBranch(workspace: ClaimedWorkspace) {
		this.pushes += 1;
		return { branch: workspace.branch, head: workspace.head, remoteName: "origin" as const };
	}
	async captureHandoff(workspace: ClaimedWorkspace, verificationState: "pending" | "passed" | "failed") {
		this.handoffs += 1;
		return {
			issue: workspace.issue, branch: workspace.branch, prBase: workspace.prBase,
			baseHead: workspace.baseHead, head: workspace.head, changedScope: [...workspace.changedScope],
			verificationState, repositoryIdentity: workspace.repositoryIdentity,
			worktreeIdentity: workspace.worktreeIdentity, dirty: false,
		};
	}
	async refreshParent(workspace: ClaimedWorkspace): Promise<ProductionParentRefreshEvidence> {
		this.refreshes += 1;
		const previousBaseHead = workspace.baseHead;
		const previousHead = workspace.head;
		workspace.baseHead = SHA_C;
		workspace.head = SHA_C;
		return {
			outcome: "reclaimed", previousBaseHead, baseHead: workspace.baseHead,
			previousHead, head: workspace.head, verificationInvalidated: true, reviewInvalidated: true,
		};
	}
}

async function fixture(t: test.TestContext) {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-production-workspace-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const workspaceAdapter = new FakeWorkspace();
	const agentSession = new FakeAgent();
	const verificationCalls: Array<{ cwd: string; signal?: AbortSignal }> = [];
	let verificationPassed = true;
	const lifecycle = new ProductionWorkspaceLifecycle({
		workspaceAdapter,
		agentSession,
		verification: {
			async runAll(cwd, commands, signal) {
				verificationCalls.push({ cwd, ...(signal ? { signal } : {}) });
				return commands.map((entry) => ({
					id: entry.id,
					status: verificationPassed ? "passed" as const : "failed" as const,
					exitCode: verificationPassed ? 0 : 1,
					signal: null,
					stdout: "",
					stderr: "",
					durationMs: 1,
					...(verificationPassed ? {} : { failureKind: "exit" as const }),
				}));
			},
		},
	});
	const coordinator = {
		cwd: root,
		repositoryIdentity: ID_A,
		worktreeIdentity: ID_B,
		remoteName: "origin" as const,
		remoteIdentity: "3".repeat(64),
		fetchEndpointIdentity: "4".repeat(64),
		pushEndpointIdentity: "4".repeat(64),
		defaultBranch: "main",
	};
	return {
		root, workspaceAdapter, agentSession, lifecycle, verificationCalls,
		setVerificationPassed(value: boolean) { verificationPassed = value; },
		claim: (lane = child(), generation = 1, mode: "start" | "resume" = "start", runId = "run-479") => lifecycle.claim({
			runId, generation, coordinator, trustedWorktreeRoot: root,
			parentIssue: 471, parentBranch: "feat/cli-architecture-v2", parentHead: SHA_A,
			child: lane, mode,
		}),
	};
}

test("claims canonical per-child workspaces with generation-independent stable ownership", async (t) => {
	const f = await fixture(t);
	const first = await f.claim(child("lane-b", 502), 1, "start");
	const second = await f.claim(child("lane-c", 503), 2, "resume");
	assert.equal(f.workspaceAdapter.claims[0].ownershipId, productionWorkspaceOwnershipId(471, 502, "lane-b"));
	assert.equal(f.workspaceAdapter.claims[1].leaseMode, "resume");
	assert.notEqual(first.binding.worktreeIdentity, second.binding.worktreeIdentity);
	assert.deepEqual(first.binding.writeScopes, [".pi/extensions/shepherd/lane-b"]);
	await Promise.all([first.join(), second.join()]);
	assert.deepEqual(f.workspaceAdapter.released.sort(), [502, 503]);
	const resumed = await f.claim(child("lane-b", 502), 9, "resume", "rotated-resume-run");
	assert.equal(f.workspaceAdapter.claims[2].ownershipId, f.workspaceAdapter.claims[0].ownershipId);
	await resumed.join();
});

test("runs implementation in the claimed workspace then verifies, commits, pushes, and captures exact handoff", async (t) => {
	const f = await fixture(t);
	const session = await f.claim();
	const implementation = await session.implement({ timeoutMs: 30_000, context: ["strict RED GREEN refactor"] });
	assert.equal(implementation.status, "completed");
	const request = f.agentSession.requests[0];
	assert.equal(request.role, "implementation");
	assert.equal(request.workspace.cwd, session.binding.cwd);
	assert.equal(request.authority.readOnly, false);
	assert.deepEqual(request.authority.writePrefixes, session.binding.writeScopes);
	assert.match(request.context.join("\n"), /javascript-testing-patterns|RED GREEN/);

	assert.equal((await session.verify()).every((result) => result.status === "passed"), true);
	const commit = await session.commit("feat(shepherd): complete lane B");
	assert.equal(commit.head, SHA_B);
	assert.equal((await session.push()).head, SHA_B);
	const captured = await session.captureHandoff();
	assert.equal(captured.verificationState, "passed");
	assert.equal(captured.head, SHA_B);
	await session.join();
	assert.deepEqual(f.workspaceAdapter.released, [502]);
});

test("verification failure blocks commit and refresh invalidates previously passing verification and review", async (t) => {
	const f = await fixture(t);
	const session = await f.claim();
	f.setVerificationPassed(false);
	const failed = await session.verify();
	assert.equal(failed[0].status, "failed");
	await assert.rejects(session.commit("feat: unsafe"), /verification/i);
	f.setVerificationPassed(true);
	await session.verify();
	const refresh = await session.refreshParent({ previousParentHead: SHA_A, newParentHead: SHA_C, effectKey: "refresh:lane-b:1" });
	assert.equal(refresh.verificationInvalidated, true);
	assert.equal(refresh.reviewInvalidated, true);
	assert.equal(session.binding.baseHead, SHA_C);
	await assert.rejects(session.commit("feat: stale evidence"), /verification/i);
	await session.verify();
	await session.commit("feat: refreshed evidence");
	await session.join();
});

test("correction runs read-write in the same claimed workspace and invalidates prior exact-head evidence", async (t) => {
	const f = await fixture(t);
	const session = await f.claim();
	await session.verify();
	await session.correct({
		timeoutMs: 30_000,
		findings: ["P1: lifecycle must release only its own lease"],
		context: ["Apply the bounded correction only"],
	});
	const request = f.agentSession.requests[0];
	assert.equal(request.role, "correction");
	assert.equal(request.workspace.cwd, session.binding.cwd);
	assert.equal(request.authority.readOnly, false);
	assert.deepEqual(request.authority.writePrefixes, session.binding.writeScopes);
	assert.match(request.task, /P1: lifecycle must release only its own lease/);
	await assert.rejects(session.commit("fix: stale verification"), /verification/i);
	await session.verify();
	await session.commit("fix: verified correction");
	await session.join();
});

test("join waits for accepted AgentSession work and releases exactly that session's lease once", async (t) => {
	const f = await fixture(t);
	let unblock!: () => void;
	f.agentSession.gate = new Promise<void>((resolve) => unblock = resolve);
	const first = await f.claim(child("lane-b", 502));
	const second = await f.claim(child("lane-c", 503));
	const implementation = first.implement({ timeoutMs: 30_000 });
	let joined = false;
	const joining = first.join().then(() => joined = true);
	await new Promise<void>((resolve) => setImmediate(resolve));
	assert.equal(joined, false);
	assert.deepEqual(f.workspaceAdapter.released, []);
	unblock();
	await Promise.all([implementation, joining]);
	assert.deepEqual(f.workspaceAdapter.released, [502]);
	await first.join();
	assert.deepEqual(f.workspaceAdapter.released, [502]);
	await second.join();
	assert.deepEqual(f.workspaceAdapter.released, [502, 503]);
});

test("abort joins only sessions for its run before release and close owns the AgentSession port once", async (t) => {
	const f = await fixture(t);
	const session = await f.claim();
	await f.lifecycle.abort("run-479");
	assert.deepEqual(f.agentSession.aborted, ["run-479"]);
	assert.deepEqual(f.workspaceAdapter.released, [502]);
	await session.join();
	await f.lifecycle.close();
	await f.lifecycle.close();
	assert.equal(f.agentSession.closed, 1);
});

test("abort failure cannot strand the run's accepted workspace lease", async (t) => {
	const f = await fixture(t);
	await f.claim();
	f.agentSession.abortFailure = new Error("agent abort failed");
	await assert.rejects(f.lifecycle.abort("run-479"), /abort failed/);
	assert.deepEqual(f.workspaceAdapter.released, [502]);
});

function executeThenTimeoutOn(...commands: string[]): GitCommandExecutor {
	const timedOut = new Set(commands);
	return (request) => new Promise((resolve, reject) => {
		execFile("git", request.args, {
			cwd: request.cwd,
			encoding: "buffer",
			env: request.env,
			maxBuffer: request.maxOutputBytes,
			timeout: request.timeoutMs,
			killSignal: "SIGTERM",
		}, (error, stdout) => {
			if (error) { reject(error); return; }
			if (timedOut.delete(request.args[0])) {
				const timeout = new Error("simulated timeout") as Error & { code: string };
				timeout.code = "ETIMEDOUT";
				reject(timeout);
				return;
			}
			resolve(stdout);
		});
	});
}

test("real workspace refresh reclaims an untouched child and rebases unique child commits onto the advanced parent", async (t) => {
	for (const uniqueCommit of [false, true]) {
		const fixture = await createLocalGitFixture();
		t.after(fixture.cleanup);
		const gitAdapter = new GitAdapter();
		const workspaceAdapter = new WorkspaceAdapter(gitAdapter, { leaseOptions: {
			processId: 20_000 + (uniqueCommit ? 1 : 0),
			processIdentity: `production-refresh-test-${uniqueCommit ? "rebase" : "reclaim"}`,
			isProcessAlive: () => true,
			tokenFactory: () => `production-refresh-token-${uniqueCommit ? "rebase" : "reclaim"}`,
		} });
		const workspace = await workspaceAdapter.claim({
			coordinator: await gitAdapter.inspect(fixture.coordinator),
			trustedWorktreeRoot: fixture.worktreeRoot,
			issue: uniqueCommit ? 512 : 511,
			slug: uniqueCommit ? "refresh-rebase" : "refresh-reclaim",
			parentIssue: 471,
			parentBranch: fixture.parentBranch,
			parentHead: fixture.parentHead,
			ownershipId: `production-refresh-${uniqueCommit ? "rebase" : "reclaim"}`,
			allowedScopes: ["child.txt"],
		});
		if (uniqueCommit) {
			await write(workspace.cwd, "child.txt", "unique child change\n");
			const committed = await workspaceAdapter.commitIssueChanges(workspace, {
				issue: workspace.issue, slug: workspace.slug, branch: workspace.branch,
				expectedHead: workspace.head, message: "test: unique child change", scopes: ["child.txt"],
			});
			workspace.head = committed.head;
		}
		await write(fixture.coordinator, "parent-advanced.txt", "advanced parent\n");
		await git(fixture.coordinator, "add", "--", "parent-advanced.txt");
		await git(fixture.coordinator, "commit", "-m", "test: advance parent");
		const advancedParent = (await git(fixture.coordinator, "rev-parse", "HEAD")).trim();
		const evidence = await workspaceAdapter.refreshParent(workspace, {
			previousParentHead: fixture.parentHead,
			newParentHead: advancedParent,
			effectKey: `refresh:${workspace.issue}:1`,
		});
		assert.equal(evidence.outcome, uniqueCommit ? "rebased" : "reclaimed");
		assert.equal(evidence.baseHead, advancedParent);
		assert.equal(evidence.verificationInvalidated, true);
		assert.equal(evidence.reviewInvalidated, true);
		assert.equal(await git(workspace.cwd, "merge-base", "--is-ancestor", advancedParent, evidence.head).then(() => true), true);
		if (uniqueCommit) assert.equal(await import("node:fs/promises").then(({ readFile }) => readFile(join(workspace.cwd, "child.txt"), "utf8")), "unique child change\n");
		await workspace.release();
		const resumedGit = new GitAdapter();
		const resumedAdapter = new WorkspaceAdapter(resumedGit, { leaseOptions: {
			processId: 21_000 + (uniqueCommit ? 1 : 0),
			processIdentity: `production-refresh-resume-${uniqueCommit ? "rebase" : "reclaim"}`,
			isProcessAlive: () => true,
			tokenFactory: () => `production-refresh-resume-token-${uniqueCommit ? "rebase" : "reclaim"}`,
		} });
		const resumed = await resumedAdapter.claim({
			coordinator: await resumedGit.inspect(fixture.coordinator),
			trustedWorktreeRoot: fixture.worktreeRoot,
			issue: workspace.issue,
			slug: workspace.slug,
			parentIssue: 471,
			parentBranch: fixture.parentBranch,
			parentHead: advancedParent,
			ownershipId: `production-refresh-${uniqueCommit ? "rebase" : "reclaim"}`,
			allowedScopes: ["child.txt"],
			leaseMode: "resume",
		});
		assert.equal(resumed.baseHead, advancedParent);
		assert.equal(resumed.head, evidence.head);
		await resumed.release();
	}
});

test("typed commit and push reconcile authoritative exact-head state after post-publication timeouts", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const gitAdapter = new GitAdapter({ execute: executeThenTimeoutOn("commit", "push") });
	const workspaceAdapter = new WorkspaceAdapter(gitAdapter, { leaseOptions: {
		processId: 20_003,
		processIdentity: "production-timeout-test",
		isProcessAlive: () => true,
		tokenFactory: () => "production-timeout-token",
	} });
	const workspace = await workspaceAdapter.claim({
		coordinator: await gitAdapter.inspect(fixture.coordinator),
		trustedWorktreeRoot: fixture.worktreeRoot,
		issue: 513,
		slug: "authoritative-timeout",
		parentIssue: 471,
		parentBranch: fixture.parentBranch,
		parentHead: fixture.parentHead,
		ownershipId: "production-authoritative-timeout",
		allowedScopes: ["timeout.txt"],
	});
	await write(workspace.cwd, "timeout.txt", "published exactly once\n");
	const committed = await workspaceAdapter.commitIssueChanges(workspace, {
		issue: workspace.issue, slug: workspace.slug, branch: workspace.branch,
		expectedHead: workspace.head, message: "test: reconcile commit timeout", scopes: ["timeout.txt"],
	});
	workspace.head = committed.head;
	assert.equal(committed.committed, true);
	const pushed = await workspaceAdapter.pushIssueBranch(workspace, {
		issue: workspace.issue, slug: workspace.slug, branch: workspace.branch,
		expectedHead: committed.head, defaultBranch: "main",
	});
	assert.equal(pushed.head, committed.head);
	assert.equal((await git(fixture.remote, "rev-parse", `refs/heads/${workspace.branch}`)).trim(), committed.head);
	await workspace.release();
});
