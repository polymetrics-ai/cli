import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import { promisify } from "node:util";

import { captureTargetEvidence } from "./target-evidence.ts";

const execFileAsync = promisify(execFile);
const head = "a".repeat(40);
const repositoryIdentity = "b".repeat(64);
const worktreeIdentity = "c".repeat(64);
const statusCommand = "git -c core.fsmonitor=false -c core.untrackedCache=false -c status.showUntrackedFiles=all -c diff.ignoreSubmodules=none status --porcelain=v1 --untracked-files=all --ignore-submodules=none";
const repositoryRemoteCommand = "git config --local --no-includes --get remote.origin.url";
const repositoryRemote = "git@github.com:polymetrics-ai/cli.git\n";
const pullRequestCommand = "gh pr view 438 --repo github.com/polymetrics-ai/cli --json number,state,isDraft,baseRefName,headRefName,headRefOid,url,mergeStateStatus,reviewDecision,statusCheckRollup";

function fakeExec(outputs) {
	const calls = [];
	return {
		calls,
		async exec(file, args, cwd) {
			calls.push({ file, args, cwd });
			const key = `${file} ${args.join(" ")}`;
			if (!(key in outputs)) throw new Error(`unexpected command: ${key}`);
			return outputs[key];
		},
	};
}

test("binds a clean local checkout to the exact open PR head", async () => {
	const harness = fakeExec({
		"git rev-parse HEAD": `${head}\n`,
		"git branch --show-current": "feat/cli-architecture-v2\n",
		[statusCommand]: "",
		[repositoryRemoteCommand]: repositoryRemote,
		[pullRequestCommand]: JSON.stringify({
			number: 438,
			state: "OPEN",
			isDraft: true,
			baseRefName: "main",
			headRefName: "feat/cli-architecture-v2",
			headRefOid: head,
			url: "https://github.com/polymetrics-ai/cli/pull/438",
			mergeStateStatus: "CLEAN",
			reviewDecision: "",
			statusCheckRollup: [
				{ __typename: "CheckRun", name: "verify", status: "COMPLETED", conclusion: "SUCCESS" },
				{ __typename: "StatusContext", context: "security/snyk", state: "SUCCESS" },
			],
		}),
	});
	const result = await captureTargetEvidence({ cwd: "/repo", issue: 397, pr: 438, repositoryIdentity, worktreeIdentity }, harness.exec);
	assert.deepEqual(result, {
		cwd: "/repo",
		repositoryIdentity,
		worktreeIdentity,
		branch: "feat/cli-architecture-v2",
		candidateHead: head,
		clean: true,
		pr: 438,
		prUrl: "https://github.com/polymetrics-ai/cli/pull/438",
		baseBranch: "main",
		draft: true,
		prState: "OPEN",
		mergeStateStatus: "CLEAN",
		reviewDecision: "",
		statusChecks: [
			{ name: "security/snyk", status: "SUCCESS" },
			{ name: "verify", status: "COMPLETED", conclusion: "SUCCESS" },
		],
	});
	assert.ok(harness.calls.every((call) => Array.isArray(call.args)), "commands must use argv arrays");
});

test("captures a clean local exact head without invoking GitHub when no PR is supplied", async () => {
	const harness = fakeExec({
		"git rev-parse HEAD": `${head}\n`,
		"git branch --show-current": "feat/local-only\n",
		[statusCommand]: "",
	});
	const result = await captureTargetEvidence({ cwd: "/repo", issue: 471, repositoryIdentity, worktreeIdentity }, harness.exec);
	assert.deepEqual(result, {
		cwd: "/repo",
		repositoryIdentity,
		worktreeIdentity,
		branch: "feat/local-only",
		candidateHead: head,
		clean: true,
	});
	assert.equal(harness.calls.some((call) => call.file === "gh"), false);
});

test("fails closed on a dirty tree, closed PR, branch mismatch, or stale local head", async () => {
	const base = {
		"git rev-parse HEAD": `${head}\n`,
		"git branch --show-current": "feat/cli-architecture-v2\n",
		[statusCommand]: "",
		[repositoryRemoteCommand]: repositoryRemote,
		[pullRequestCommand]: JSON.stringify({
			number: 438,
			state: "OPEN",
			isDraft: true,
			baseRefName: "main",
			headRefName: "feat/cli-architecture-v2",
			headRefOid: head,
			url: "https://github.com/polymetrics-ai/cli/pull/438",
			mergeStateStatus: "CLEAN",
			reviewDecision: "",
			statusCheckRollup: [],
		}),
	};
	for (const [name, overrides, message] of [
		["dirty", { [statusCommand]: " M file\n" }, /must be clean/],
		["closed", { [pullRequestCommand]: JSON.stringify({ number: 438, state: "CLOSED", isDraft: false, baseRefName: "main", headRefName: "feat/cli-architecture-v2", headRefOid: head, url: "https://github.com/polymetrics-ai/cli/pull/438", mergeStateStatus: "CLEAN", reviewDecision: "", statusCheckRollup: [] }) }, /must be open/],
		["branch", { "git branch --show-current": "main\n" }, /branch does not match/],
		["head", { "git rev-parse HEAD": `${"b".repeat(40)}\n` }, /head does not match/],
	]) {
		const harness = fakeExec({ ...base, ...overrides });
		await assert.rejects(
			captureTargetEvidence({ cwd: "/repo", issue: 397, pr: 438, repositoryIdentity, worktreeIdentity }, harness.exec),
			message,
			name,
		);
	}
});

async function initRepository(root) {
	await mkdir(root);
	await execFileAsync("git", ["init", "--quiet", root]);
	await execFileAsync("git", [
		"-C", root,
		"-c", "user.name=Shepherd Test",
		"-c", "user.email=shepherd@example.invalid",
		"commit", "--allow-empty", "--quiet", "-m", "seed",
	]);
}

test("real Git status cannot hide untracked files through repository configuration", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-clean-untracked-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const repo = join(root, "repo");
	await initRepository(repo);
	await execFileAsync("git", ["-C", repo, "config", "status.showUntrackedFiles", "no"]);
	await writeFile(join(repo, "untracked.txt"), "must be visible\n");
	await assert.rejects(
		captureTargetEvidence({ cwd: repo, issue: 471, repositoryIdentity, worktreeIdentity }),
		/must be clean/,
	);
});

test("real Git status cannot hide a dirty submodule through repository configuration", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-clean-submodule-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const parent = join(root, "parent");
	const child = join(root, "child");
	await initRepository(parent);
	await initRepository(child);
	await writeFile(join(child, "tracked.txt"), "clean\n");
	await execFileAsync("git", ["-C", child, "add", "tracked.txt"]);
	await execFileAsync("git", ["-C", child, "-c", "user.name=Shepherd Test", "-c", "user.email=shepherd@example.invalid", "commit", "--quiet", "-m", "tracked"]);
	await execFileAsync("git", ["-C", parent, "-c", "protocol.file.allow=always", "submodule", "add", "--quiet", child, "module"]);
	await execFileAsync("git", ["-C", parent, "-c", "user.name=Shepherd Test", "-c", "user.email=shepherd@example.invalid", "commit", "--quiet", "-am", "submodule"]);
	await execFileAsync("git", ["-C", parent, "config", "diff.ignoreSubmodules", "all"]);
	await writeFile(join(parent, "module", "tracked.txt"), "dirty\n");
	await assert.rejects(
		captureTargetEvidence({ cwd: parent, issue: 471, repositoryIdentity, worktreeIdentity }),
		/must be clean/,
	);
});

test("rejects credential-bearing or cross-PR GitHub URLs", async () => {
	const base = {
		"git rev-parse HEAD": `${head}\n`,
		"git branch --show-current": "feat/cli-architecture-v2\n",
		[statusCommand]: "",
		[repositoryRemoteCommand]: repositoryRemote,
	};
	for (const url of [
		"https://user@example.invalid/org/repo/pull/438",
		"https://example.invalid/polymetrics-ai/cli/pull/438",
		"https://github.com/polymetrics-ai/cli/pull/999",
	]) {
		const harness = fakeExec({
			...base,
			[pullRequestCommand]: JSON.stringify({
				number: 438,
				state: "OPEN",
				isDraft: false,
				baseRefName: "main",
				headRefName: "feat/cli-architecture-v2",
				headRefOid: head,
				url,
				mergeStateStatus: "CLEAN",
				reviewDecision: "",
				statusCheckRollup: [],
			}),
		});
		await assert.rejects(
			captureTargetEvidence({ cwd: "/repo", issue: 397, pr: 438, repositoryIdentity, worktreeIdentity }, harness.exec),
			/URL|pull request evidence/i,
		);
	}
});

test("rejects a pull request from a different repository with the same branch and head", async () => {
	const pullRequest = JSON.stringify({
		number: 438,
		state: "OPEN",
		isDraft: false,
		baseRefName: "main",
		headRefName: "feat/cli-architecture-v2",
		headRefOid: head,
		url: "https://github.com/attacker-fork/cli/pull/438",
		mergeStateStatus: "CLEAN",
		reviewDecision: "",
		statusCheckRollup: [],
	});
	const harness = fakeExec({
		"git rev-parse HEAD": `${head}\n`,
		"git branch --show-current": "feat/cli-architecture-v2\n",
		[statusCommand]: "",
		[repositoryRemoteCommand]: repositoryRemote,
		[pullRequestCommand]: pullRequest,
	});

	await assert.rejects(
		captureTargetEvidence({ cwd: "/repo", issue: 397, pr: 438, repositoryIdentity, worktreeIdentity }, harness.exec),
		/repository/i,
	);
	assert.deepEqual(
		harness.calls.find((call) => call.file === "gh")?.args.slice(0, 5),
		["pr", "view", "438", "--repo", "github.com/polymetrics-ai/cli"],
	);
});
