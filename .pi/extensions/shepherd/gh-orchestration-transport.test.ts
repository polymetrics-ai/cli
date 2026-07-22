import assert from "node:assert/strict";
import test from "node:test";

import {
	GhCliOrchestrationTransport,
	GhRequiredCheckPolicySource,
	createProductionGitHubOrchestrationFacade,
	type GhExecutionOptions,
	type GhOrchestrationExecutor,
} from "./gh-orchestration-transport.ts";
import type {
	CreateChildIssueRequest,
	ExternalCallContext,
	IntegrateChildRequest,
} from "./github-orchestrator.ts";

function context(signal = new AbortController().signal): ExternalCallContext {
	return {
		signal,
		deadlineAt: new Date(Date.now() + 60_000).toISOString(),
		acknowledgeAbort() {},
	};
}

const mutation = {
	schemaVersion: 1 as const,
	operation: "child_issue" as const,
	idempotencyKey: "shepherd-mutation:v1:child_issue:abc",
	intentDigest: "d".repeat(64),
	expectedResourceRevision: null,
};

function childRequest(): CreateChildIssueRequest {
	const marker = "<!-- shepherd-child-issue:v1:471:child-a:abc -->";
	return {
		repository: "acme/widgets",
		parentIssue: 471,
		marker,
		title: "Child A",
		body: `Implement A\n\nParent: #471\n\n${marker}`,
		mutation,
	};
}

test("reconciles a child issue timeout after GitHub published the exact marker", async () => {
	const issues: unknown[] = [];
	const calls: Array<{ file: string; args: readonly string[]; options: GhExecutionOptions }> = [];
	const execute: GhOrchestrationExecutor = async (file, args, options) => {
		calls.push({ file, args, options });
		const endpoint = args[3] ?? args[2] ?? "";
		if (args.includes("POST") && String(endpoint).endsWith("/issues")) {
			issues.push({ number: 17, title: "Child A", body: childRequest().body, state: "open" });
			throw Object.assign(new Error("socket closed after publication"), { code: "ECONNRESET" });
		}
		if (String(endpoint).includes("/issues?")) return JSON.stringify(issues);
		throw new Error(`unexpected gh call ${args.join(" ")}`);
	};
	const transport = new GhCliOrchestrationTransport({ execute, maxPages: 2 });
	const result = await transport.createChildIssue(childRequest(), context());
	assert.equal(result.value.number, 17);
	assert.equal(result.applied, false);
	assert.ok(calls.every((call) => call.file === "gh" && call.options.maxOutputBytes <= 2 * 1024 * 1024));
	assert.ok(calls.every((call) => !call.args.some((arg) => /token|authorization/i.test(arg))));
});

test("bounded cancellation reaches the injected gh executor and acknowledges abort", async () => {
	const controller = new AbortController();
	let acknowledged = 0;
	const execute: GhOrchestrationExecutor = async (_file, _args, options) => {
		queueMicrotask(() => controller.abort());
		await new Promise<void>((_resolve, reject) => options.signal.addEventListener("abort", () => reject(new Error("aborted")), { once: true }));
		return "";
	};
	const transport = new GhCliOrchestrationTransport({ execute });
	await assert.rejects(transport.findChildIssues({ repository: "acme/widgets", marker: childRequest().marker }, {
		...context(controller.signal),
		acknowledgeAbort() { acknowledged += 1; },
	}), /cancel|abort/i);
	assert.equal(acknowledged, 1);
});

test("production facade exposes the real transport and explicitly refuses to fake parent-ready durable authority", () => {
	const facade = createProductionGitHubOrchestrationFacade({ execute: async () => "[]" });
	assert.ok(facade.transport instanceof GhCliOrchestrationTransport);
	assert.equal(facade.parentReadyAuthority, null);
	assert.equal(facade.parentReadyAuthorityDependency, "required-external-durable-authority");
});

test("reads stable authoritative required-check policies for parent and base branches", async () => {
	const execute: GhOrchestrationExecutor = async (_file, args) => {
		assert.match(args[3], /required_status_checks$/);
		return JSON.stringify({
			strict: true,
			contexts: ["legacy-ci"],
			checks: [{ context: "build", app_id: 15368 }],
		});
	};
	const source = new GhRequiredCheckPolicySource({
		execute,
		now: () => new Date("2026-07-22T00:00:00.000Z"),
	});
	const first = await source.findRequiredCheckPolicies({ repository: "acme/widgets", baseBranch: "main" }, context());
	const second = await source.findRequiredCheckPolicies({ repository: "acme/widgets", baseBranch: "main" }, context());
	assert.deepEqual(second, first);
	const bundle = await source.findParentOrchestrationPolicyBundle!({
		repository: "acme/widgets",
		parentIssue: 471,
		generation: 7,
		parentBranch: "feat/parent",
		parentBaseBranch: "main",
	}, context());
	assert.equal(bundle.items[0].policyBundle.requiredCheckPolicies.length, 2);
	assert.deepEqual(bundle.items[0].policyBundle.requiredCheckPolicies[0].requiredChecks, [
		{ name: "build", producerId: "15368" },
		{ name: "legacy-ci", producerId: "legacy" },
	]);
});

test("reads checks from the authoritative live head and fails closed on bounded pagination", async () => {
	const marker = "<!-- shepherd-pr:v1:471:child-a:abc -->";
	const metadata = {
		workItemId: "child-a",
		generation: 7,
		marker,
		baseSha: "1".repeat(40),
		headSha: "2".repeat(40),
		changedPaths: ["old.ts"],
		allowedScopes: ["src/**"],
		policyDigest: "d".repeat(64),
	};
	const metadataBody = `<!-- shepherd-transport-pr-meta:v1:${Buffer.from(JSON.stringify(metadata)).toString("base64url")} -->`;
	const liveHead = "3".repeat(40);
	const execute: GhOrchestrationExecutor = async (_file, args) => {
		const endpoint = args[3] ?? "";
		if (endpoint.includes("/pulls?")) return JSON.stringify([{ number: 9, body: `body\n${marker}` }]);
		if (endpoint.endsWith("/pulls/9")) return JSON.stringify({
			number: 9,
			title: "Child A",
			body: `body\n${marker}`,
			state: "open",
			draft: false,
			merged: false,
			merged_at: null,
			mergeable_state: "clean",
			updated_at: "2026-07-22T00:00:05.000Z",
			base: { ref: "feat/parent", sha: "4".repeat(40) },
			head: { ref: "feat/child", sha: liveHead },
		});
		if (endpoint.includes("/issues/9/comments?")) return JSON.stringify([{ id: 81, body: metadataBody }]);
		if (endpoint.includes("/pulls/9/files?")) return JSON.stringify([{ filename: "src/live.ts" }]);
		if (endpoint.includes(`/commits/${liveHead}/check-runs?`)) return JSON.stringify({
			check_runs: Array.from({ length: 100 }, (_, index) => ({
				id: index + 1,
				name: `build-${index}`,
				status: "completed",
				conclusion: "success",
				app: { id: 15368 },
				updated_at: "2026-07-22T00:00:01.000Z",
				completed_at: "2026-07-22T00:00:01.000Z",
			})),
		});
		if (endpoint.includes(`/commits/${liveHead}/statuses?`)) return "[]";
		if (endpoint.includes("/pulls/9/reviews?")) return "[]";
		if (args[1] === "graphql") return JSON.stringify({
			data: { repository: { pullRequest: { reviewThreads: {
				nodes: [],
				pageInfo: { hasNextPage: false, endCursor: null },
			} } } },
		});
		throw new Error(`unexpected gh call ${args.join(" ")}`);
	};
	const transport = new GhCliOrchestrationTransport({ execute, maxPages: 1 });
	const result = await transport.findPullRequests({ repository: "acme/widgets", marker }, context());
	assert.equal(result.items.length, 1);
	assert.equal(result.items[0].headSha, liveHead);
	assert.deepEqual(result.items[0].changedPaths, ["src/live.ts"]);
	assert.equal(result.items[0].checks.length, 100);
	assert.equal(result.items[0].checks[0].producerId, "15368");
	assert.equal(result.items[0].checks[0].headSha, liveHead);
	assert.equal(result.items[0].checksComplete, false);
});

test("refuses child integration into a default branch before any GitHub mutation", async () => {
	let calls = 0;
	const transport = new GhCliOrchestrationTransport({ execute: async () => {
		calls += 1;
		return "{}";
	} });
	const request = {
		repository: "acme/widgets",
		childId: "child-a",
		pullRequest: 9,
		generation: 7,
		marker: "<!-- shepherd-pr:v1:471:child-a:abc -->",
		baseSha: "1".repeat(40),
		headSha: "2".repeat(40),
		parentBranch: "release",
		parentBaseBranch: "release",
		mutation,
	} as unknown as IntegrateChildRequest;
	await assert.rejects(transport.integrateChild(request, context()), /refuses default-branch/i);
	assert.equal(calls, 0);
});

test("child integration transport never merges a PR and publishes nothing until Git evidence is authoritatively merged", async () => {
	const marker = "<!-- shepherd-pr:v1:471:child-a:abc -->";
	const methods: string[] = [];
	const execute: GhOrchestrationExecutor = async (_file, args) => {
		const method = args[2] ?? "";
		const endpoint = args[3] ?? "";
		methods.push(method);
		if (endpoint.includes("/pulls?")) return JSON.stringify([{ number: 9, body: marker }]);
		if (endpoint.includes("/issues/9/comments?")) return "[]";
		if (endpoint === "/repos/acme/widgets") return JSON.stringify({ default_branch: "main" });
		if (endpoint === "/repos/acme/widgets/git/ref/heads/feat%2Fparent") return JSON.stringify({
			ref: "refs/heads/feat/parent",
			object: { type: "commit", sha: "3".repeat(40) },
		});
		if (endpoint.endsWith("/pulls/9")) return JSON.stringify({
			number: 9,
			merged: false,
			merged_at: null,
			base: { ref: "feat/parent", sha: "1".repeat(40) },
			head: { ref: "feat/child", sha: "2".repeat(40) },
		});
		throw new Error(`unexpected gh call ${args.join(" ")}`);
	};
	const transport = new GhCliOrchestrationTransport({ execute });
	const request = {
		repository: "acme/widgets",
		childId: "child-a",
		pullRequest: 9,
		generation: 7,
		marker,
		baseSha: "1".repeat(40),
		headSha: "2".repeat(40),
		parentBranch: "feat/parent",
		parentBaseBranch: "main",
		integration: {
			schemaVersion: 1,
			authority: "git",
			parentBranch: "feat/parent",
			baseSha: "1".repeat(40),
			headSha: "2".repeat(40),
			mergeCommitSha: "3".repeat(40),
			parentHead: "3".repeat(40),
			reused: false,
		},
		pullRequestSnapshot: {},
		observation: {},
		controllerProvenance: {},
		mutation: {
			schemaVersion: 1,
			operation: "child_integration",
			idempotencyKey: "shepherd-mutation:v1:child_integration:abc",
			intentDigest: "d".repeat(64),
			expectedResourceRevision: null,
		},
	} as unknown as IntegrateChildRequest;
	await assert.rejects(
		transport.integrateChild(request, context()),
		/merged|integration.*evidence|uncertain publication/i,
	);
	assert.deepEqual([...new Set(methods)], ["GET"]);
});

test("child integration transport publishes one receipt only after exact merged PR and parent-ref reconciliation", async () => {
	const marker = "<!-- shepherd-pr:v1:471:child-a:abc -->";
	const calls: string[][] = [];
	const execute: GhOrchestrationExecutor = async (_file, args) => {
		calls.push([...args]);
		const method = args[2] ?? "";
		const endpoint = args[3] ?? "";
		if (endpoint.includes("/pulls?")) return JSON.stringify([{ number: 9, body: marker }]);
		if (endpoint.includes("/issues/9/comments?")) return "[]";
		if (endpoint === "/repos/acme/widgets") return JSON.stringify({ default_branch: "main" });
		if (endpoint === "/repos/acme/widgets/git/ref/heads/feat%2Fparent") return JSON.stringify({
			ref: "refs/heads/feat/parent",
			object: { type: "commit", sha: "3".repeat(40) },
		});
		if (endpoint.endsWith("/pulls/9")) return JSON.stringify({
			number: 9,
			merged: true,
			merged_at: "2026-07-22T00:00:00.000Z",
			merge_commit_sha: "3".repeat(40),
			base: { ref: "feat/parent", sha: "3".repeat(40) },
			head: { ref: "feat/child", sha: "2".repeat(40) },
		});
		if (method === "POST" && endpoint.endsWith("/issues/9/comments")) {
			return JSON.stringify({ id: 88, body: args.at(-1) });
		}
		if (method === "PATCH" && endpoint.endsWith("/issues/comments/88")) return "{}";
		throw new Error(`unexpected gh call ${args.join(" ")}`);
	};
	const now = new Date("2026-07-22T00:00:01.000Z");
	const transport = new GhCliOrchestrationTransport({ execute, now: () => now });
	const request = {
		repository: "acme/widgets",
		childId: "child-a",
		pullRequest: 9,
		generation: 7,
		marker,
		baseSha: "1".repeat(40),
		headSha: "2".repeat(40),
		parentBranch: "feat/parent",
		parentBaseBranch: "main",
		integration: {
			schemaVersion: 1,
			authority: "git",
			parentBranch: "feat/parent",
			baseSha: "1".repeat(40),
			headSha: "2".repeat(40),
			mergeCommitSha: "3".repeat(40),
			parentHead: "3".repeat(40),
			reused: false,
		},
		pullRequestSnapshot: {},
		observation: {},
		controllerProvenance: {},
		mutation: {
			schemaVersion: 1,
			operation: "child_integration",
			idempotencyKey: "shepherd-mutation:v1:child_integration:abc",
			intentDigest: "d".repeat(64),
			expectedResourceRevision: null,
		},
	} as unknown as IntegrateChildRequest;
	const result = await transport.integrateChild(request, context());
	assert.equal(result.value.childId, "child-a");
	assert.equal(result.value.headSha, "2".repeat(40));
	assert.equal(result.applied, true);
	assert.equal(calls.some((args) => args.some((arg) => /\/merge$/u.test(arg))), false);
});
