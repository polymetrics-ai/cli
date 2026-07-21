import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { access, mkdir, readFile, realpath } from "node:fs/promises";
import { join } from "node:path";
import test from "node:test";

import {
	GitAdapter,
	canonicalIssueBranch,
	type GitCommandExecutor,
	type GitCommandRequest,
} from "./git-adapter.ts";
import { createLocalGitFixture, git, write, type LocalGitFixture } from "./issue-476-git-fixture.ts";
import { resolveCanonicalGitWorktree } from "./target-evidence.ts";
import { WorkspaceAdapter, type ClaimedWorkspace } from "./workspace-adapter.ts";

function recordingExecutor(requests: GitCommandRequest[]): GitCommandExecutor {
	return (request) => new Promise((resolve, reject) => {
		requests.push({ ...request, args: [...request.args] });
		execFile("git", request.args, {
			cwd: request.cwd,
			encoding: "buffer",
			env: request.env,
			maxBuffer: request.maxOutputBytes,
			timeout: request.timeoutMs,
			killSignal: "SIGTERM",
		}, (error, stdout) => error ? reject(error) : resolve(stdout));
	});
}

async function claimedMutationWorkspace(
	adapter: GitAdapter,
	fixture: LocalGitFixture,
	allowedScopes: readonly string[],
): Promise<{ workspaceAdapter: WorkspaceAdapter; workspace: ClaimedWorkspace }> {
	const workspaceAdapter = new WorkspaceAdapter(adapter);
	const workspace = await workspaceAdapter.claim({
		coordinator: await adapter.inspect(fixture.coordinator),
		trustedWorktreeRoot: fixture.worktreeRoot,
		issue: 476,
		slug: "shepherd-worktree-git-adapter",
		parentIssue: 471,
		parentSlug: "pi-agent-session-shepherd",
		parentHead: fixture.parentHead,
		ownershipId: "issue-476-git-adapter-test",
		allowedScopes,
	});
	return { workspaceAdapter, workspace };
}

test("derives one canonical safe issue branch", () => {
	assert.equal(canonicalIssueBranch(476, "shepherd-worktree-git-adapter"), "feat/476-shepherd-worktree-git-adapter");
	assert.throws(() => canonicalIssueBranch(0, "valid"), /issue/);
	for (const slug of ["../escape", "UPPER", "two--dashes", "-leading", "trailing-", "main", "bad value", "bad\u0000value"]) {
		assert.throws(() => canonicalIssueBranch(476, slug), /slug/);
	}
});

test("binds canonical repository identity across linked worktrees", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const sibling = join(fixture.root, "identity-worktree");
	await git(fixture.coordinator, "worktree", "add", "-b", "feat/900-identity", "--", sibling, fixture.parentHead);
	const adapter = new GitAdapter();
	const coordinator = await adapter.inspect(fixture.coordinator);
	const linked = await adapter.inspect(sibling);
	assert.equal(coordinator.repositoryIdentity, linked.repositoryIdentity);
	assert.equal(coordinator.remoteIdentity, linked.remoteIdentity);
	assert.equal(coordinator.fetchEndpointIdentity, linked.fetchEndpointIdentity);
	assert.equal(coordinator.pushEndpointIdentity, linked.pushEndpointIdentity);
	assert.notEqual(coordinator.worktreeIdentity, linked.worktreeIdentity);
	assert.equal(coordinator.remoteName, "origin");
});

test("uses the canonical identities already persisted by Shepherd target evidence", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const sibling = join(fixture.root, "state-identity-worktree");
	await git(fixture.coordinator, "worktree", "add", "-b", "feat/901-state-identity", "--", sibling, fixture.parentHead);
	const adapter = new GitAdapter();
	for (const cwd of [fixture.coordinator, sibling]) {
		const [adapterBinding, stateBinding] = await Promise.all([
			adapter.inspect(cwd),
			resolveCanonicalGitWorktree(cwd),
		]);
		assert.equal(adapterBinding.cwd, stateBinding.cwd);
		assert.equal(adapterBinding.repositoryIdentity, stateBinding.repositoryIdentity);
		assert.equal(adapterBinding.worktreeIdentity, stateBinding.worktreeIdentity);
	}
	for (const remote of [
		"https://Example.invalid/Owner/Repo.git",
		"ssh://git@Example.invalid/Owner/Repo.git",
		"git@Example.invalid:Owner/Repo.git",
		`file://${fixture.remote}`,
	]) {
		await git(fixture.coordinator, "remote", "set-url", "origin", remote);
		const [adapterBinding, stateBinding] = await Promise.all([
			adapter.inspect(fixture.coordinator),
			resolveCanonicalGitWorktree(fixture.coordinator),
		]);
		assert.equal(adapterBinding.repositoryIdentity, stateBinding.repositoryIdentity);
		assert.equal(adapterBinding.worktreeIdentity, stateBinding.worktreeIdentity);
	}
	await git(fixture.coordinator, "remote", "remove", "origin");
	const [adapterWithoutRemote, stateWithoutRemote] = await Promise.all([
		adapter.inspect(fixture.coordinator),
		resolveCanonicalGitWorktree(fixture.coordinator),
	]);
	assert.equal(adapterWithoutRemote.repositoryIdentity, stateWithoutRemote.repositoryIdentity);
	assert.equal(adapterWithoutRemote.worktreeIdentity, stateWithoutRemote.worktreeIdentity);
});

test("rejects credential-bearing or mismatched remote identity without echoing credentials", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new GitAdapter();
	const original = await adapter.inspect(fixture.coordinator);
	await git(fixture.coordinator, "remote", "set-url", "origin", "https://user:not-a-secret@example.invalid/repo.git");
	await assert.rejects(adapter.inspect(fixture.coordinator), (error: Error) => {
		assert.match(error.message, /credential|remote/i);
		assert.doesNotMatch(error.message, /not-a-secret/);
		return true;
	});
	await git(fixture.coordinator, "remote", "set-url", "origin", fixture.remote);
	assert.deepEqual(await adapter.assertBinding(original), original);
	await assert.rejects(adapter.assertBinding({ ...original, repositoryIdentity: "f".repeat(64) }), /repository identity mismatch/);
});

test("reports dirty tracked and untracked state without changing either file", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new GitAdapter();
	const binding = await adapter.inspect(fixture.coordinator);
	await write(fixture.coordinator, "parent.txt", "dirty parent\n");
	await write(fixture.coordinator, "untracked.txt", "unique\n");
	const status = await adapter.status(binding);
	assert.equal(status.clean, false);
	assert.deepEqual(status.entries.map((entry) => [entry.code, entry.path]), [
		[" M", "parent.txt"],
		["??", "untracked.txt"],
	]);
	assert.equal(await readFile(join(fixture.coordinator, "parent.txt"), "utf8"), "dirty parent\n");
	assert.equal(await readFile(join(fixture.coordinator, "untracked.txt"), "utf8"), "unique\n");
});

test("creates a bounded commit, returns exact head evidence, and makes no-op retry idempotent", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new GitAdapter();
	const branch = canonicalIssueBranch(476, "shepherd-worktree-git-adapter");
	const { workspaceAdapter, workspace } = await claimedMutationWorkspace(
		adapter, fixture, [".pi/extensions/shepherd/example.ts"],
	);
	t.after(() => workspace.release().catch(() => undefined));
	await write(workspace.cwd, ".pi/extensions/shepherd/example.ts", "export const value = 1;\n");
	const committed = await workspaceAdapter.commitIssueChanges(workspace, {
		issue: 476,
		slug: "shepherd-worktree-git-adapter",
		branch,
		expectedHead: fixture.parentHead,
		message: "test(shepherd): add bounded fixture",
		scopes: [".pi/extensions/shepherd/example.ts"],
	});
	assert.equal(committed.committed, true);
	assert.match(committed.head, /^[0-9a-f]{40}$/);
	assert.notEqual(committed.head, fixture.parentHead);
	const retried = await workspaceAdapter.commitIssueChanges(workspace, {
		issue: 476,
		slug: "shepherd-worktree-git-adapter",
		branch,
		expectedHead: committed.head,
		message: "test(shepherd): add bounded fixture",
		scopes: [".pi/extensions/shepherd/example.ts"],
	});
	assert.deepEqual(retried, { committed: false, previousHead: committed.head, head: committed.head });
});

test("refuses a scoped commit when dirty or staged state exists outside its allowlist", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new GitAdapter();
	const branch = canonicalIssueBranch(476, "shepherd-worktree-git-adapter");
	const { workspaceAdapter, workspace } = await claimedMutationWorkspace(adapter, fixture, ["allowed.txt"]);
	t.after(() => workspace.release().catch(() => undefined));
	await write(workspace.cwd, "allowed.txt", "allowed\n");
	await write(workspace.cwd, "unique.txt", "preserve\n");
	await assert.rejects(workspaceAdapter.commitIssueChanges(workspace, {
		issue: 476,
		slug: "shepherd-worktree-git-adapter",
		branch,
		expectedHead: fixture.parentHead,
		message: "test: bounded",
		scopes: ["allowed.txt"],
	}), /outside declared scopes.*unique\.txt/);
	assert.equal((await git(workspace.cwd, "rev-parse", "HEAD")).trim(), fixture.parentHead);
	assert.equal(await readFile(join(workspace.cwd, "unique.txt"), "utf8"), "preserve\n");
	await access(join(workspace.cwd, "allowed.txt"));
});

test("rejects unsafe scopes, stale heads, branch aliases, and direct default-branch push", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new GitAdapter();
	const { workspaceAdapter, workspace } = await claimedMutationWorkspace(adapter, fixture, ["README.md"]);
	t.after(() => workspace.release().catch(() => undefined));
	const common = {
		issue: 476,
		slug: "shepherd-worktree-git-adapter",
		branch: fixture.parentBranch,
		expectedHead: fixture.parentHead,
		message: "test: rejected",
	};
	for (const scopes of [["."], ["../escape"], ["/absolute"], [".git/config"], ["bad\u0000path"]]) {
		await assert.rejects(workspaceAdapter.commitIssueChanges(workspace, { ...common, scopes }), /branch|scope/);
	}
	await assert.rejects(workspaceAdapter.pushIssueBranch(workspace, {
		issue: 476,
		slug: "shepherd-worktree-git-adapter",
		branch: "main",
		expectedHead: fixture.parentHead,
		defaultBranch: "main",
	}), /canonical|default branch/);
	await assert.rejects(adapter.diff(workspace, {
		baseHead: "a".repeat(40),
		head: fixture.parentHead,
		scopes: ["README.md"],
	}), /base head.*not present|object/i);
});

test("pushes only the canonical branch and verifies the exact remote head", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const requests: GitCommandRequest[] = [];
	const adapter = new GitAdapter({ execute: recordingExecutor(requests) });
	const branch = canonicalIssueBranch(476, "shepherd-worktree-git-adapter");
	const { workspaceAdapter, workspace } = await claimedMutationWorkspace(adapter, fixture, ["README.md"]);
	t.after(() => workspace.release().catch(() => undefined));
	const evidence = await workspaceAdapter.pushIssueBranch(workspace, {
		issue: 476,
		slug: "shepherd-worktree-git-adapter",
		branch,
		expectedHead: fixture.parentHead,
		defaultBranch: "main",
	});
	assert.deepEqual(evidence, { branch, head: fixture.parentHead, remoteName: "origin" });
	assert.equal((await git(fixture.remote, "rev-parse", `refs/heads/${branch}`)).trim(), fixture.parentHead);
	const flattened = requests.flatMap((request) => request.args);
	for (const forbidden of ["--force", "--force-with-lease", "reset", "clean", "prune", "remove"]) {
		assert.equal(flattened.includes(forbidden), false);
	}
	const push = requests.find((request) => request.args[0] === "push");
	assert.deepEqual(push?.args, ["push", "--porcelain", "--", fixture.remote, branch]);
});

test("rejects a changed effective push endpoint before the alternate remote receives objects", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const alternate = join(fixture.root, "alternate-push.git");
	await git(fixture.root, "init", "--bare", "--initial-branch=main", alternate);
	const adapter = new GitAdapter();
	const branch = canonicalIssueBranch(476, "shepherd-worktree-git-adapter");
	const { workspaceAdapter, workspace } = await claimedMutationWorkspace(adapter, fixture, ["README.md"]);
	t.after(() => workspace.release().catch(() => undefined));
	await git(workspace.cwd, "config", "remote.origin.pushurl", alternate);

	const [push] = await Promise.allSettled([workspaceAdapter.pushIssueBranch(workspace, {
		issue: 476,
		slug: "shepherd-worktree-git-adapter",
		branch,
		expectedHead: fixture.parentHead,
		defaultBranch: "main",
	})]);
	const [alternateRef, alternateObject] = await Promise.allSettled([
		git(alternate, "show-ref", "--verify", `refs/heads/${branch}`),
		git(alternate, "cat-file", "-e", `${fixture.parentHead}^{commit}`),
	]);
	assert.equal(alternateRef.status, "rejected", "the changed pushurl received the issue branch before validation");
	assert.equal(alternateObject.status, "rejected", "the changed pushurl received Git objects before validation");
	assert.equal(push.status, "rejected");
	if (push.status === "rejected") assert.match(String(push.reason), /push endpoint|remote.*mismatch/i);
});

test("rejects chained URL rewrites before the twice-expanded endpoint receives objects", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const alternateDirectory = join(fixture.root, "alternate");
	const alternate = join(alternateDirectory, "origin.git");
	await mkdir(alternateDirectory);
	await git(fixture.root, "init", "--bare", "--initial-branch=main", alternate);
	await git(fixture.coordinator, "remote", "set-url", "origin", "safe:origin.git");
	await git(fixture.coordinator, "config", `url.file://${fixture.root}/.insteadOf`, "safe:");
	await git(
		fixture.coordinator,
		"config",
		`url.file://${alternateDirectory}/.insteadOf`,
		`file://${fixture.root}/`,
	);
	const adapter = new GitAdapter();
	const branch = canonicalIssueBranch(476, "shepherd-worktree-git-adapter");
	const { workspaceAdapter, workspace } = await claimedMutationWorkspace(adapter, fixture, ["README.md"]);
	t.after(() => workspace.release().catch(() => undefined));
	const [push] = await Promise.allSettled([workspaceAdapter.pushIssueBranch(workspace, {
		issue: 476,
		slug: "shepherd-worktree-git-adapter",
		branch,
		expectedHead: fixture.parentHead,
		defaultBranch: "main",
	})]);
	const [alternateRef, alternateObject] = await Promise.allSettled([
		git(alternate, "show-ref", "--verify", `refs/heads/${branch}`),
		git(alternate, "cat-file", "-e", `${fixture.parentHead}^{commit}`),
	]);
	assert.equal(push.status, "rejected");
	if (push.status === "rejected") assert.match(String(push.reason), /rewrite|endpoint/i);
	assert.equal(alternateRef.status, "rejected", "the twice-expanded endpoint received the issue branch");
	assert.equal(alternateObject.status, "rejected", "the twice-expanded endpoint received Git objects");
});

test("rejects literal backslash paths instead of canonicalizing them into an allowed scope", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new GitAdapter();
	const binding = await adapter.inspect(fixture.coordinator);
	const escaped = String.raw`.pi\extensions\shepherd\escape.ts`;
	await write(fixture.coordinator, escaped, "must remain a root-level filename\n");
	await git(fixture.coordinator, "add", "--", escaped);
	await git(fixture.coordinator, "commit", "-m", "test(shepherd): literal backslash path");
	const head = (await git(fixture.coordinator, "rev-parse", "HEAD")).trim();
	await assert.rejects(adapter.diff(binding, {
		baseHead: fixture.parentHead,
		head,
		scopes: [".pi/extensions/shepherd"],
	}), /scope|backslash/i);
});

test("rejects forged mutation leases and alternate-root lease issuance during an active claim", async (t) => {
	const fixture = await createLocalGitFixture();
	t.after(fixture.cleanup);
	const adapter = new GitAdapter();
	const branch = canonicalIssueBranch(476, "shepherd-worktree-git-adapter");
	const { workspace } = await claimedMutationWorkspace(adapter, fixture, ["README.md"]);
	t.after(() => workspace.release().catch(() => undefined));
	const alternateStateRoot = join(fixture.root, "alternate-state");
	await mkdir(alternateStateRoot, { mode: 0o700 });
	await assert.rejects(adapter.acquireMutationLease({}, workspace, {
		issue: 476,
		slug: "shepherd-worktree-git-adapter",
		branch,
		baseHead: fixture.parentHead,
		targetCwd: workspace.cwd,
		allowedScopes: ["README.md"],
		stateRoot: await realpath(alternateStateRoot),
		runId: "forged-alternate-owner",
		mode: "start",
	}), /only be issued by the owning workspace adapter/i);
	await assert.rejects(access(join(alternateStateRoot, "leases")));

	const forgedLease = {
		assertOwned: async () => undefined,
		release: async () => undefined,
	};
	const mutations = [
		adapter.fetchBranch(forgedLease, workspace, branch),
		adapter.addIssueWorktree(forgedLease, workspace, {
			trustedRoot: fixture.worktreeRoot,
			path: workspace.cwd,
			issue: 476,
			slug: "shepherd-worktree-git-adapter",
			branch,
			baseHead: fixture.parentHead,
		}),
		adapter.commitIssueChanges(forgedLease, workspace, {
			issue: 476,
			slug: "shepherd-worktree-git-adapter",
			branch,
			expectedHead: fixture.parentHead,
			message: "test(shepherd): forged",
			scopes: ["README.md"],
		}),
		adapter.pushIssueBranch(forgedLease, workspace, {
			issue: 476,
			slug: "shepherd-worktree-git-adapter",
			branch,
			expectedHead: fixture.parentHead,
			defaultBranch: "main",
		}),
	];
	for (const mutation of mutations) await assert.rejects(mutation, /not issued by this adapter/i);
});
