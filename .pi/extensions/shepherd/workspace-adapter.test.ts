import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { mkdir, readFile, symlink, writeFile } from "node:fs/promises";
import { basename, join } from "node:path";
import test from "node:test";

import {
	GitAdapter,
	canonicalIssueBranch,
	type GitCommandExecutor,
} from "./git-adapter.ts";
import { createLocalGitFixture, git, write } from "./issue-476-git-fixture.ts";
import {
	WorkspaceAdapter,
	type WorkspaceAdapterOptions,
	type WorkspaceClaimRequest,
} from "./workspace-adapter.ts";

function adapterWithLeaseOptions(leaseOptions: NonNullable<WorkspaceAdapterOptions["leaseOptions"]>): WorkspaceAdapter {
	return new WorkspaceAdapter(new GitAdapter(), { leaseOptions });
}

function executorGatedOnAdd(started: () => void, waitUntilReleased: Promise<void>): GitCommandExecutor {
	return async (request) => {
		if (request.args[0] === "add") {
			started();
			await waitUntilReleased;
		}
		return new Promise((resolve, reject) => {
			execFile("git", request.args, {
				cwd: request.cwd,
				encoding: "buffer",
				env: request.env,
				maxBuffer: request.maxOutputBytes,
				timeout: request.timeoutMs,
				killSignal: "SIGTERM",
			}, (error, stdout) => error ? reject(error) : resolve(stdout));
		});
	};
}

async function requestFor(
	fixture: Awaited<ReturnType<typeof createLocalGitFixture>>,
	overrides: Partial<WorkspaceClaimRequest> = {},
): Promise<WorkspaceClaimRequest> {
	const gitAdapter = new GitAdapter();
	return {
		coordinator: await gitAdapter.inspect(fixture.coordinator),
		trustedWorktreeRoot: fixture.worktreeRoot,
		issue: 476,
		slug: "shepherd-worktree-git-adapter",
		parentIssue: 471,
		parentSlug: "pi-agent-session-shepherd",
		parentHead: fixture.parentHead,
		ownershipId: "run-471-generation-1-lane-476",
		allowedScopes: [".pi/extensions/shepherd", ".planning/phases/476-shepherd-worktree-git-adapter"],
		...overrides,
	};
}

test("creates one canonical isolated issue worktree from the exact parent base", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new WorkspaceAdapter(new GitAdapter());
	const workspace = await adapter.claim(await requestFor(fixture));
	assert.equal(workspace.branch, canonicalIssueBranch(476, "shepherd-worktree-git-adapter"));
	assert.equal(workspace.prBase, fixture.parentBranch);
	assert.equal(workspace.baseHead, fixture.parentHead);
	assert.equal(workspace.head, fixture.parentHead);
	assert.equal(workspace.reused, false);
	assert.equal(workspace.status.clean, true);
	assert.equal(workspace.repositoryIdentity, (await requestFor(fixture)).coordinator.repositoryIdentity);
	assert.equal(basename(workspace.cwd), "issue-476-shepherd-worktree-git-adapter");
	assert.notEqual(workspace.cwd, fixture.coordinator);
	assert.equal((await git(workspace.cwd, "branch", "--show-current")).trim(), workspace.branch);
	await workspace.release();
});

test("reconciles an exact crash retry without creating a duplicate branch or worktree", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new WorkspaceAdapter(new GitAdapter());
	const request = await requestFor(fixture);
	const first = await adapter.claim(request);
	await first.release();
	const second = await new WorkspaceAdapter(new GitAdapter()).claim(request);
	assert.equal(second.reused, true);
	assert.equal(second.cwd, first.cwd);
	assert.equal(second.worktreeIdentity, first.worktreeIdentity);
	const inventory = await new GitAdapter().listWorktrees(request.coordinator);
	assert.equal(inventory.filter((entry) => entry.branch === first.branch).length, 1);
	await second.release();
});

test("fails closed for a second owner and preserves the first owner's workspace", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new WorkspaceAdapter(new GitAdapter());
	const request = await requestFor(fixture);
	const first = await adapter.claim(request);
	await assert.rejects(adapter.claim({ ...request, ownershipId: "different-owner" }), /already owned/);
	assert.equal((await git(first.cwd, "branch", "--show-current")).trim(), first.branch);
	await first.release();
});

test("two concurrent distinct owners cannot become active mutators", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const request = await requestFor(fixture);
	const results = await Promise.allSettled([
		new WorkspaceAdapter(new GitAdapter()).claim({ ...request, ownershipId: "owner-a" }),
		new WorkspaceAdapter(new GitAdapter()).claim({ ...request, ownershipId: "owner-b" }),
	]);
	assert.equal(results.filter((result) => result.status === "fulfilled").length, 1);
	assert.equal(results.filter((result) => result.status === "rejected").length, 1);
	for (const result of results) if (result.status === "fulfilled") await result.value.release();
	const inventory = await new GitAdapter().listWorktrees(request.coordinator);
	assert.equal(inventory.filter((entry) => entry.branch === canonicalIssueBranch(476, request.slug)).length, 1);
});

test("same-owner overlapping retries receive one writable lease and release permits an exact retry", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const request = await requestFor(fixture);
	const results = await Promise.allSettled([
		new WorkspaceAdapter(new GitAdapter()).claim(request),
		new WorkspaceAdapter(new GitAdapter()).claim(request),
	]);
	const fulfilled = results.filter((result) => result.status === "fulfilled");
	assert.equal(fulfilled.length, 1);
	assert.equal(results.filter((result) => result.status === "rejected").length, 1);
	const first = fulfilled[0].value;
	await first.assertOwned();
	await first.release();
	await first.release();
	await assert.rejects(first.assertOwned(), /released|ownership.*lost/i);
	const retry = await new WorkspaceAdapter(new GitAdapter()).claim(request);
	assert.equal(retry.reused, true);
	assert.equal(retry.worktreeIdentity, first.worktreeIdentity);
	await retry.release();
});

test("a released workspace cannot commit after an exact replacement lease becomes active", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const gitAdapter = new GitAdapter();
	const adapter = new WorkspaceAdapter(gitAdapter);
	const request = await requestFor(fixture);
	const released = await adapter.claim(request);
	await released.release();
	const replacement = await adapter.claim(request);
	await write(replacement.cwd, ".pi/extensions/shepherd/post-release.ts", "export const fenced = true;\n");
	const before = (await git(replacement.cwd, "rev-parse", "HEAD")).trim();
	const [mutation] = await Promise.allSettled([
		adapter.commitIssueChanges(released, {
			issue: released.issue,
			slug: released.slug,
			branch: released.branch,
			expectedHead: before,
			message: "test(shepherd): reject released mutator",
			scopes: [".pi/extensions/shepherd/post-release.ts"],
		}),
	]);
	const after = (await git(replacement.cwd, "rev-parse", "HEAD")).trim();
	const replacementMutation = await adapter.commitIssueChanges(replacement, {
		issue: replacement.issue,
		slug: replacement.slug,
		branch: replacement.branch,
		expectedHead: before,
		message: "test(shepherd): replacement owns mutation",
		scopes: [".pi/extensions/shepherd/post-release.ts"],
	});
	const replacementHead = (await git(replacement.cwd, "rev-parse", "HEAD")).trim();
	await replacement.release();
	assert.equal(mutation.status, "rejected", "the released claim retained a usable Git mutation authority");
	assert.equal(after, before, "a released claim advanced the replacement owner's branch");
	assert.equal(replacementMutation.committed, true);
	assert.equal(replacementHead, replacementMutation.head);
	assert.notEqual(replacementHead, before);
});

test("release waits for an already accepted Git mutation to finish", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	let signalAddStarted!: () => void;
	const addStarted = new Promise<void>((resolve) => signalAddStarted = resolve);
	let unblockAdd!: () => void;
	const addGate = new Promise<void>((resolve) => unblockAdd = resolve);
	const gitAdapter = new GitAdapter({ execute: executorGatedOnAdd(signalAddStarted, addGate) });
	const adapter = new WorkspaceAdapter(gitAdapter);
	const workspace = await adapter.claim(await requestFor(fixture));
	t.after(() => workspace.release().catch(() => undefined));
	await write(workspace.cwd, ".pi/extensions/shepherd/in-flight.ts", "export const inFlight = true;\n");
	const commitIssueChanges = Reflect.get(adapter, "commitIssueChanges");
	assert.equal(typeof commitIssueChanges, "function", "workspace mutations require an adapter-minted active claim capability");
	const mutation = Reflect.apply(commitIssueChanges, adapter, [workspace, {
		issue: workspace.issue,
		slug: workspace.slug,
		branch: workspace.branch,
		expectedHead: workspace.head,
		message: "test(shepherd): serialize release",
		scopes: [".pi/extensions/shepherd/in-flight.ts"],
	}]) as Promise<unknown>;
	await addStarted;
	let releaseSettled = false;
	const release = workspace.release().finally(() => releaseSettled = true);
	await new Promise<void>((resolve) => setImmediate(resolve));
	const settledWhileMutationBlocked = releaseSettled;
	unblockAdd();
	const [mutationResult, releaseResult] = await Promise.allSettled([mutation, release]);
	assert.equal(settledWhileMutationBlocked, false, "release completed while an accepted mutation was in flight");
	assert.equal(mutationResult.status, "fulfilled");
	assert.equal(releaseResult.status, "fulfilled");
});

test("resume recovers an exact stale same-request lease without reviving the fenced owner", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const request = await requestFor(fixture);
	const crashed = adapterWithLeaseOptions({
		processId: 10_001,
		processIdentity: "process-10001-start-1",
		isProcessAlive: () => false,
		tokenFactory: () => "workspace-crashed-owner",
	});
	const first = await crashed.claim(request);
	const startOnly = adapterWithLeaseOptions({
		processId: 10_002,
		processIdentity: "process-10002-start-1",
		isProcessAlive: (pid: number) => pid === 10_002,
		tokenFactory: () => "workspace-start-only-owner",
	});
	await assert.rejects(startOnly.claim(request), /stale.*resume/i);
	const recovering = adapterWithLeaseOptions({
		processId: 10_003,
		processIdentity: "process-10003-start-1",
		isProcessAlive: (pid: number) => pid === 10_003,
		tokenFactory: () => "workspace-recovered-owner",
	});
	const recovered = await recovering.claim({ ...request, leaseMode: "resume" });
	await recovered.assertOwned();
	await assert.rejects(first.assertOwned(), /ownership.*lost|token mismatch/i);
	assert.equal(recovered.reused, true);
	assert.equal(recovered.worktreeIdentity, first.worktreeIdentity);
	await recovered.release();
});

test("rejects branch aliases and path collisions while preserving unique state", async (t) => {
	const aliasFixture = await createLocalGitFixture();
	t.after(aliasFixture.cleanup);
	await git(aliasFixture.coordinator, "branch", "feat/476-alias", aliasFixture.parentHead);
	await assert.rejects(
		new WorkspaceAdapter(new GitAdapter()).claim(await requestFor(aliasFixture)),
		/aliased branch ownership.*feat\/476-alias/,
	);
	assert.equal((await git(aliasFixture.coordinator, "rev-parse", "feat/476-alias")).trim(), aliasFixture.parentHead);

	const pathFixture = await createLocalGitFixture();
	t.after(pathFixture.cleanup);
	const collision = join(pathFixture.worktreeRoot, "issue-476-shepherd-worktree-git-adapter");
	await mkdir(collision);
	await write(collision, "unique.txt", "preserve collision\n");
	await assert.rejects(
		new WorkspaceAdapter(new GitAdapter()).claim(await requestFor(pathFixture)),
		/worktree path collision/,
	);
	assert.equal(await readFile(join(collision, "unique.txt"), "utf8"), "preserve collision\n");
});

test("rejects stale parent heads, repository aliases, and symlinked trusted roots", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new WorkspaceAdapter(new GitAdapter());
	await assert.rejects(adapter.claim(await requestFor(fixture, { parentHead: "a".repeat(40) })), /parent head mismatch|not present/);
	const mismatched = await requestFor(fixture);
	mismatched.coordinator = { ...mismatched.coordinator, repositoryIdentity: "f".repeat(64) };
	await assert.rejects(adapter.claim(mismatched), /repository identity mismatch/);
	const link = join(fixture.root, "workers-link");
	await symlink(fixture.worktreeRoot, link, "dir");
	await assert.rejects(adapter.claim(await requestFor(fixture, { trustedWorktreeRoot: link })), /trusted worktree root.*symlink/);
});

test("reports and preserves dirty unique state on retry and emits exact handoff evidence", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new WorkspaceAdapter(new GitAdapter());
	const request = await requestFor(fixture, {
		allowedScopes: ["parent.txt", "unique.txt"],
	});
	const first = await adapter.claim(request);
	await write(first.cwd, "parent.txt", "modified but preserved\n");
	await write(first.cwd, "unique.txt", "unique and preserved\n");
	await first.release();
	const retried = await adapter.claim(request);
	assert.equal(retried.reused, true);
	assert.equal(retried.status.clean, false);
	assert.deepEqual(retried.changedScope, ["parent.txt", "unique.txt"]);
	const handoff = await adapter.captureHandoff(retried, "passed");
	assert.equal(handoff.baseHead, fixture.parentHead);
	assert.equal(handoff.head, fixture.parentHead);
	assert.equal(handoff.prBase, fixture.parentBranch);
	assert.equal(handoff.verificationState, "passed");
	assert.deepEqual(handoff.changedScope, ["parent.txt", "unique.txt"]);
	assert.equal(await readFile(join(first.cwd, "parent.txt"), "utf8"), "modified but preserved\n");
	assert.equal(await readFile(join(first.cwd, "unique.txt"), "utf8"), "unique and preserved\n");
	await retried.release();
});

test("handoff rejects mutable PR-base, base-head, and scope evidence not bound by the persisted claim", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new WorkspaceAdapter(new GitAdapter());
	const workspace = await adapter.claim(await requestFor(fixture));
	const ancestor = (await git(workspace.cwd, "rev-parse", `${fixture.parentHead}^`)).trim();
	const mutable = workspace as unknown as {
		prBase: string;
		baseHead: string;
		allowedScopes: readonly string[];
	};
	mutable.prBase = "feat/999-forged-parent";
	await assert.rejects(adapter.captureHandoff(workspace, "passed"), /immutable persisted original claim/i);
	mutable.prBase = fixture.parentBranch;
	mutable.baseHead = ancestor;
	await assert.rejects(adapter.captureHandoff(workspace, "passed"), /immutable persisted original claim/i);
	mutable.baseHead = fixture.parentHead;
	mutable.allowedScopes = ["parent.txt"];
	await assert.rejects(adapter.captureHandoff(workspace, "passed"), /immutable persisted original claim/i);
	await workspace.release();
});

test("handoff fails closed if the immutable persisted claim is modified after acquisition", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new WorkspaceAdapter(new GitAdapter());
	const workspace = await adapter.claim(await requestFor(fixture));
	const claimPath = join(fixture.worktreeRoot, ".shepherd-workspace-claims", "issue-476.json");
	const persisted = JSON.parse(await readFile(claimPath, "utf8")) as Record<string, unknown>;
	assert.deepEqual(persisted.allowedScopes, [...workspace.allowedScopes]);
	persisted.prBase = "feat/999-forged-parent";
	await writeFile(claimPath, `${JSON.stringify(persisted)}\n`, { mode: 0o600 });
	await assert.rejects(adapter.captureHandoff(workspace, "passed"), /persisted|claim|immutable/i);
	await workspace.release();
});

test("handoff audits the complete committed path set before applying immutable scopes", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new WorkspaceAdapter(new GitAdapter());
	const workspace = await adapter.claim(await requestFor(fixture, {
		allowedScopes: [".pi/extensions/shepherd"],
	}));
	await write(workspace.cwd, ".pi/extensions/shepherd/allowed.ts", "export const allowed = true;\n");
	await write(workspace.cwd, "outside.txt", "must not be omitted\n");
	await git(workspace.cwd, "add", "--", ".pi/extensions/shepherd/allowed.ts", "outside.txt");
	await git(workspace.cwd, "commit", "-m", "test(shepherd): mixed handoff scope");
	const [handoff] = await Promise.allSettled([adapter.captureHandoff(workspace, "passed")]);
	await workspace.release();
	assert.equal(handoff.status, "rejected", "handoff passed after omitting a committed path outside the immutable scopes");
	if (handoff.status === "rejected") assert.match(String(handoff.reason), /outside.*scope|scope.*outside/i);
});

test("handoff rejects a dirty literal backslash path instead of aliasing it into an allowed scope", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new WorkspaceAdapter(new GitAdapter());
	const workspace = await adapter.claim(await requestFor(fixture, {
		allowedScopes: [".pi/extensions/shepherd"],
	}));
	const escaped = String.raw`.pi\extensions\shepherd\dirty.ts`;
	await write(workspace.cwd, escaped, "must remain a root-level filename\n");
	const [handoff] = await Promise.allSettled([adapter.captureHandoff(workspace, "passed")]);
	await workspace.release();
	assert.equal(handoff.status, "rejected");
	if (handoff.status === "rejected") assert.match(String(handoff.reason), /scope|backslash/i);
});

test("fails handoff exact-head verification when the parent base is not an ancestor", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new WorkspaceAdapter(new GitAdapter());
	const workspace = await adapter.claim(await requestFor(fixture, { allowedScopes: ["new.txt"] }));
	await git(workspace.cwd, "config", "user.name", "Shepherd Test");
	await git(workspace.cwd, "config", "user.email", "shepherd@example.invalid");
	await write(workspace.cwd, "new.txt", "new\n");
	await git(workspace.cwd, "add", "--", "new.txt");
	await git(workspace.cwd, "commit", "-m", "test: new head");
	const handoff = await adapter.captureHandoff(workspace, "pending");
	assert.notEqual(handoff.head, workspace.head);
	assert.equal(handoff.baseHead, fixture.parentHead);
	assert.deepEqual(handoff.changedScope, ["new.txt"]);
	await git(workspace.cwd, "checkout", "--orphan", "unrelated");
	await git(workspace.cwd, "rm", "-rf", ".");
	await write(workspace.cwd, "unrelated.txt", "unrelated\n");
	await git(workspace.cwd, "add", "--", "unrelated.txt");
	await git(workspace.cwd, "commit", "-m", "test: unrelated head");
	await assert.rejects(
		adapter.captureHandoff({ ...workspace, branch: "unrelated" }, "failed"),
		/canonical branch|base is not an ancestor|active immutable claim/,
	);
	await workspace.release();
});
