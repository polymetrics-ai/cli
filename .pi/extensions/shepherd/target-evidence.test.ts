import assert from "node:assert/strict";
import test from "node:test";

import { captureTargetEvidence } from "./target-evidence.ts";

const head = "a".repeat(40);

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
		"git status --porcelain": "",
		"gh pr view 438 --json number,state,isDraft,baseRefName,headRefName,headRefOid,url,mergeStateStatus,reviewDecision,statusCheckRollup": JSON.stringify({
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
	const result = await captureTargetEvidence({ cwd: "/repo", issue: 397, pr: 438 }, harness.exec);
	assert.deepEqual(result, {
		cwd: "/repo",
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
			{ name: "verify", status: "COMPLETED", conclusion: "SUCCESS" },
			{ name: "security/snyk", status: "SUCCESS" },
		],
	});
	assert.ok(harness.calls.every((call) => Array.isArray(call.args)), "commands must use argv arrays");
});

test("captures a clean local exact head without invoking GitHub when no PR is supplied", async () => {
	const harness = fakeExec({
		"git rev-parse HEAD": `${head}\n`,
		"git branch --show-current": "feat/local-only\n",
		"git status --porcelain": "",
	});
	const result = await captureTargetEvidence({ cwd: "/repo", issue: 471 }, harness.exec);
	assert.deepEqual(result, {
		cwd: "/repo",
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
		"git status --porcelain": "",
		"gh pr view 438 --json number,state,isDraft,baseRefName,headRefName,headRefOid,url,mergeStateStatus,reviewDecision,statusCheckRollup": JSON.stringify({
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
		["dirty", { "git status --porcelain": " M file\n" }, /must be clean/],
		["closed", { [Object.keys(base)[3]]: JSON.stringify({ number: 438, state: "CLOSED", isDraft: false, baseRefName: "main", headRefName: "feat/cli-architecture-v2", headRefOid: head, url: "url", mergeStateStatus: "CLEAN", reviewDecision: "", statusCheckRollup: [] }) }, /must be open/],
		["branch", { "git branch --show-current": "main\n" }, /branch does not match/],
		["head", { "git rev-parse HEAD": `${"b".repeat(40)}\n` }, /head does not match/],
	]) {
		const harness = fakeExec({ ...base, ...overrides });
		await assert.rejects(captureTargetEvidence({ cwd: "/repo", issue: 397, pr: 438 }, harness.exec), message, name);
	}
});
