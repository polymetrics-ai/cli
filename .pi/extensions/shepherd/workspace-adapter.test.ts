import assert from "node:assert/strict";
import { mkdir, readFile, symlink, writeFile } from "node:fs/promises";
import { basename, join } from "node:path";
import test from "node:test";

import { GitAdapter, canonicalIssueBranch } from "./git-adapter.ts";
import { createLocalGitFixture, git, write } from "./issue-476-git-fixture.ts";
import { WorkspaceAdapter, type WorkspaceClaimRequest } from "./workspace-adapter.ts";

interface WorkspaceLeaseCapability {
	assertOwned(): Promise<void>;
	release(): Promise<void>;
}

function withLease(workspace: object): WorkspaceLeaseCapability {
	return workspace as WorkspaceLeaseCapability;
}

async function releaseIfPresent(workspace: object): Promise<void> {
	const candidate = workspace as Partial<WorkspaceLeaseCapability>;
	if (typeof candidate.release === "function") await candidate.release();
}

function adapterWithLeaseOptions(leaseOptions: Record<string, unknown>): WorkspaceAdapter {
	return Reflect.construct(WorkspaceAdapter, [new GitAdapter(), { leaseOptions }]) as WorkspaceAdapter;
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
});

test("reconciles an exact crash retry without creating a duplicate branch or worktree", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new WorkspaceAdapter(new GitAdapter());
	const request = await requestFor(fixture);
	const first = await adapter.claim(request);
	await releaseIfPresent(first);
	const second = await new WorkspaceAdapter(new GitAdapter()).claim(request);
	t.after(() => releaseIfPresent(second));
	assert.equal(second.reused, true);
	assert.equal(second.cwd, first.cwd);
	assert.equal(second.worktreeIdentity, first.worktreeIdentity);
	const inventory = await new GitAdapter().listWorktrees(request.coordinator);
	assert.equal(inventory.filter((entry) => entry.branch === first.branch).length, 1);
});

test("fails closed for a second owner and preserves the first owner's workspace", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new WorkspaceAdapter(new GitAdapter());
	const request = await requestFor(fixture);
	const first = await adapter.claim(request);
	await assert.rejects(adapter.claim({ ...request, ownershipId: "different-owner" }), /already owned/);
	assert.equal((await git(first.cwd, "branch", "--show-current")).trim(), first.branch);
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
	for (const result of results) if (result.status === "fulfilled") await releaseIfPresent(result.value);
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
	await withLease(first).assertOwned();
	await withLease(first).release();
	await withLease(first).release();
	await assert.rejects(withLease(first).assertOwned(), /released|ownership.*lost/i);
	const retry = await new WorkspaceAdapter(new GitAdapter()).claim(request);
	t.after(() => releaseIfPresent(retry));
	assert.equal(retry.reused, true);
	assert.equal(retry.worktreeIdentity, first.worktreeIdentity);
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
	const recovering = adapterWithLeaseOptions({
		processId: 10_002,
		processIdentity: "process-10002-start-1",
		isProcessAlive: (pid: number) => pid === 10_002,
		tokenFactory: () => "workspace-recovered-owner",
	});
	const recovered = await recovering.claim({ ...request, leaseMode: "resume" } as WorkspaceClaimRequest);
	t.after(() => releaseIfPresent(recovered));
	await withLease(recovered).assertOwned();
	await assert.rejects(withLease(first).assertOwned(), /ownership.*lost|token mismatch/i);
	assert.equal(recovered.reused, true);
	assert.equal(recovered.worktreeIdentity, first.worktreeIdentity);
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
	const request = await requestFor(fixture);
	const first = await adapter.claim(request);
	await write(first.cwd, "parent.txt", "modified but preserved\n");
	await write(first.cwd, "unique.txt", "unique and preserved\n");
	await releaseIfPresent(first);
	const retried = await adapter.claim(request);
	t.after(() => releaseIfPresent(retried));
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
});

test("handoff rejects mutable PR-base, base-head, and scope evidence not bound by the persisted claim", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new WorkspaceAdapter(new GitAdapter());
	const workspace = await adapter.claim(await requestFor(fixture));
	t.after(() => releaseIfPresent(workspace));
	const ancestor = (await git(workspace.cwd, "rev-parse", `${fixture.parentHead}^`)).trim();
	const attempts = await Promise.allSettled([
		adapter.captureHandoff({ ...workspace, prBase: "feat/999-forged-parent" }, "passed"),
		adapter.captureHandoff({ ...workspace, baseHead: ancestor }, "passed"),
		adapter.captureHandoff({ ...workspace, allowedScopes: ["parent.txt"] }, "passed"),
	]);
	assert.deepEqual(attempts.map((result) => result.status), ["rejected", "rejected", "rejected"]);
});

test("handoff fails closed if the immutable persisted claim is modified after acquisition", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new WorkspaceAdapter(new GitAdapter());
	const workspace = await adapter.claim(await requestFor(fixture));
	t.after(() => releaseIfPresent(workspace));
	const claimPath = join(fixture.worktreeRoot, ".shepherd-workspace-claims", "issue-476.json");
	const persisted = JSON.parse(await readFile(claimPath, "utf8")) as Record<string, unknown>;
	assert.deepEqual(persisted.allowedScopes, [...workspace.allowedScopes]);
	persisted.prBase = "feat/999-forged-parent";
	await writeFile(claimPath, `${JSON.stringify(persisted)}\n`, { mode: 0o600 });
	await assert.rejects(adapter.captureHandoff(workspace, "passed"), /persisted|claim|immutable/i);
});

test("fails handoff exact-head verification when the parent base is not an ancestor", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new WorkspaceAdapter(new GitAdapter());
	const workspace = await adapter.claim(await requestFor(fixture));
	await git(workspace.cwd, "config", "user.name", "Shepherd Test");
	await git(workspace.cwd, "config", "user.email", "shepherd@example.invalid");
	await write(workspace.cwd, "new.txt", "new\n");
	await git(workspace.cwd, "add", "--", "new.txt");
	await git(workspace.cwd, "commit", "-m", "test: new head");
	const handoff = await adapter.captureHandoff(workspace, "pending");
	assert.notEqual(handoff.head, workspace.head);
	assert.equal(handoff.baseHead, fixture.parentHead);
	await git(workspace.cwd, "checkout", "--orphan", "unrelated");
	await git(workspace.cwd, "rm", "-rf", ".");
	await write(workspace.cwd, "unrelated.txt", "unrelated\n");
	await git(workspace.cwd, "add", "--", "unrelated.txt");
	await git(workspace.cwd, "commit", "-m", "test: unrelated head");
	await assert.rejects(adapter.captureHandoff({ ...workspace, branch: "unrelated" }, "failed"), /canonical branch|base is not an ancestor/);
});
