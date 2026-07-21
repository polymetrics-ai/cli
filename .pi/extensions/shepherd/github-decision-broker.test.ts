import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import {
	FileHumanDecisionRepository,
	type HumanDecisionBinding,
} from "./human-decision.ts";
import {
	GhCliDecisionTransport,
	GitHubDecisionBroker,
	renderDecisionRequestComment,
	type GitHubComment,
	type GitHubDecisionTransport,
} from "./github-decision-broker.ts";

const fixture = JSON.parse(await readFile(
	new URL("./fixtures/issue-477/github-comments.json", import.meta.url),
	"utf8",
)) as Record<string, GitHubComment>;
const head = "a".repeat(40);
const issueBinding: HumanDecisionBinding = {
	repository: "polymetrics-ai/cli",
	target: { kind: "issue", number: 471 },
	generation: 3,
};
const prBinding: HumanDecisionBinding = {
	repository: "polymetrics-ai/cli",
	target: { kind: "pull_request", number: 477 },
	generation: 3,
	headSha: head,
};

class MemoryRepository {
	readonly states = new Map<string, unknown>();
	private queue = Promise.resolve();

	async transact<T>(requestId: string, operation: (state: unknown) => Promise<{ state: unknown; value: T }> | { state: unknown; value: T }): Promise<T> {
		let resolveResult!: (value: T) => void;
		let rejectResult!: (error: unknown) => void;
		const result = new Promise<T>((resolve, reject) => { resolveResult = resolve; rejectResult = reject; });
		this.queue = this.queue.then(async () => {
			try {
				const update = await operation(structuredClone(this.states.get(requestId) ?? null));
				this.states.set(requestId, structuredClone(update.state));
				resolveResult(update.value);
			} catch (error) {
				rejectResult(error);
			}
		});
		await result;
		return result;
	}

	async load(requestId: string): Promise<unknown> {
		return structuredClone(this.states.get(requestId) ?? null);
	}
}

class FakeTransport implements GitHubDecisionTransport {
	readonly created: Array<{ binding: HumanDecisionBinding; body: string }> = [];
	readonly listed: HumanDecisionBinding[] = [];
	comments: GitHubComment[] = [];
	authenticatedActor = "shepherd-host";
	createFailure?: Error;

	async getAuthenticatedActor(): Promise<string> { return this.authenticatedActor; }
	async listComments(binding: HumanDecisionBinding): Promise<GitHubComment[]> {
		this.listed.push(structuredClone(binding));
		return structuredClone(this.comments);
	}
	async createComment(binding: HumanDecisionBinding, body: string): Promise<GitHubComment> {
		if (this.createFailure) throw this.createFailure;
		this.created.push({ binding: structuredClone(binding), body });
		const comment: GitHubComment = {
			id: 1000 + this.created.length,
			url: `https://github.com/${binding.repository}/issues/${binding.target.number}#issuecomment-${1000 + this.created.length}`,
			body,
			actor: { login: this.authenticatedActor, type: "User" },
			createdAt: "2026-07-21T10:00:00.000Z",
			updatedAt: "2026-07-21T10:00:00.000Z",
		};
		this.comments.push(comment);
		return structuredClone(comment);
	}
}

function brokerHarness(binding = issueBinding) {
	const repository = new MemoryRepository();
	const transport = new FakeTransport();
	let now = new Date("2026-07-21T10:00:00.000Z");
	const sleeps: number[] = [];
	const broker = new GitHubDecisionBroker(repository, transport, {
		now: () => now,
		sleep: async (delayMs) => { sleeps.push(delayMs); now = new Date(now.valueOf() + delayMs); },
		polling: { maxAttempts: 3, initialDelayMs: 10, maxDelayMs: 15 },
	});
	const request = {
		requestId: "req-477",
		gate: binding.target.kind === "issue" ? "requirements" as const : "review" as const,
		repository: binding.repository,
		parentIssue: 471,
		pullRequest: 477,
		generation: binding.generation,
		...(binding.headSha ? { headSha: binding.headSha } : {}),
		allowedOptions: ["approve", "reject"],
		actorAllowlist: ["maintainer-one", "maintainer-two"],
		expiresAt: "2026-07-22T10:00:00.000Z",
		question: "Approve this exact gate?",
	};
	return { repository, transport, broker, request, sleeps, setNow: (value: string) => { now = new Date(value); } };
}

test("creates exactly one marker-owned request across retry and broker restart", async () => {
	const harness = brokerHarness();
	const first = await harness.broker.request(harness.request);
	const second = await harness.broker.request(harness.request);
	const restarted = new GitHubDecisionBroker(harness.repository, harness.transport);
	const third = await restarted.request(harness.request);
	assert.equal(harness.transport.created.length, 1);
	assert.equal(first.requestComment?.id, second.requestComment?.id);
	assert.equal(second.requestComment?.id, third.requestComment?.id);
	assert.match(harness.transport.created[0].body, /<!-- shepherd-decision:v1:req-477:/);
	assert.equal(harness.transport.created[0].body, renderDecisionRequestComment(first));
});

test("recovers an exact owned marker after a crash before local comment persistence", async () => {
	const harness = brokerHarness();
	const original = await harness.broker.request(harness.request);
	const state = harness.repository.states.get("req-477") as Record<string, unknown>;
	delete state.requestComment;
	harness.repository.states.set("req-477", state);
	const recovered = await new GitHubDecisionBroker(harness.repository, harness.transport).request(harness.request);
	assert.equal(recovered.requestComment?.id, original.requestComment?.id);
	assert.equal(harness.transport.created.length, 1);
});

test("fails closed on duplicate, foreign-owner, or body-colliding markers", async () => {
	for (const variant of ["duplicate", "foreign", "collision"] as const) {
		const harness = brokerHarness();
		const record = await harness.broker.request(harness.request);
		const owned = harness.transport.comments[0];
		if (variant === "duplicate") harness.transport.comments.push({ ...owned, id: owned.id + 1, url: `${owned.url}-2` });
		if (variant === "foreign") harness.transport.comments[0] = { ...owned, actor: { login: "attacker", type: "User" } };
		if (variant === "collision") harness.transport.comments[0] = { ...owned, body: `${record.idempotencyMarker}\nmalicious replacement` };
		const state = harness.repository.states.get("req-477") as Record<string, unknown>;
		delete state.requestComment;
		harness.repository.states.set("req-477", state);
		await assert.rejects(new GitHubDecisionBroker(harness.repository, harness.transport).request(harness.request), /marker|owner|collision|duplicate/i, variant);
	}
});

test("accepts only one exact unedited allowlisted human command and persists minimal evidence", async () => {
	const harness = brokerHarness();
	await harness.broker.request(harness.request);
	harness.transport.comments.push(fixture.allowlistedHuman);
	const result = await harness.broker.poll("req-477", issueBinding);
	assert.equal(result.status, "decided");
	assert.deepEqual(result.decision, {
		option: "approve",
		actor: "maintainer-one",
		sourceUrl: fixture.allowlistedHuman.url,
		decidedAt: fixture.allowlistedHuman.createdAt,
	});
	assert.equal(JSON.stringify(harness.repository.states.get("req-477")).includes("/shepherd decide"), false);
});

test("ignores bot, edited, disallowed, unknown, hostile multiline, emoji, review, CI, and silence", async () => {
	const harness = brokerHarness();
	await harness.broker.request(harness.request);
	harness.transport.comments.push(
		fixture.bot,
		fixture.edited,
		fixture.hostileMultiline,
		fixture.emoji,
		{ ...fixture.allowlistedHuman, id: 2010, actor: { login: "stranger", type: "User" } },
		{ ...fixture.allowlistedHuman, id: 2011, body: "/shepherd decide req-477 maybe" },
		{ ...fixture.allowlistedHuman, id: 2012, body: "APPROVED reviewDecision=APPROVED statusChecks=SUCCESS" },
	);
	const result = await harness.broker.poll("req-477", issueBinding);
	assert.equal(result.status, "pending");
	assert.equal(result.attempts, 3);
	assert.deepEqual(harness.sleeps, [10, 15]);
});

test("duplicate or conflicting valid commands are ambiguous and fail closed", async () => {
	for (const secondOption of ["approve", "reject"] as const) {
		const harness = brokerHarness();
		await harness.broker.request(harness.request);
		harness.transport.comments.push(
			fixture.allowlistedHuman,
			{ ...fixture.allowlistedHuman, id: 2099, url: `${fixture.allowlistedHuman.url}-duplicate`, body: `/shepherd decide req-477 ${secondOption}` },
		);
		await assert.rejects(harness.broker.poll("req-477", issueBinding), /ambiguous|multiple/i);
	}
});

test("stale repository, target, generation, or head is rejected before GitHub polling", async () => {
	const harness = brokerHarness(prBinding);
	await harness.broker.request(harness.request);
	for (const binding of [
		{ ...prBinding, repository: "other/repo" },
		{ ...prBinding, target: { kind: "pull_request" as const, number: 478 } },
		{ ...prBinding, generation: 4 },
		{ ...prBinding, headSha: "b".repeat(40) },
	]) {
		const before = harness.transport.listed.length;
		await assert.rejects(harness.broker.poll("req-477", binding), /binding|stale|target/i);
		assert.equal(harness.transport.listed.length, before);
	}
});

test("expires without inference and never accepts a response at or after expiry", async () => {
	const harness = brokerHarness();
	await harness.broker.request(harness.request);
	harness.setNow("2026-07-22T10:00:00.000Z");
	harness.transport.comments.push(fixture.allowlistedHuman);
	const result = await harness.broker.poll("req-477", issueBinding);
	assert.equal(result.status, "expired");
});

test("issue and PR routes preserve the exact typed repository, target, generation, and head", async () => {
	for (const binding of [issueBinding, prBinding]) {
		const harness = brokerHarness(binding);
		await harness.broker.request(harness.request);
		assert.deepEqual(harness.transport.created[0].binding, binding);
	}
});

test("typed gh adapter uses ambient host auth, bounded argv calls, pagination, and strict payload parsing", async () => {
	const calls: Array<{ file: string; args: string[] }> = [];
	const page = JSON.stringify([{
		id: 7,
		html_url: "https://github.com/polymetrics-ai/cli/issues/471#issuecomment-7",
		body: "safe",
		user: { login: "maintainer-one", type: "User" },
		created_at: "2026-07-21T10:00:00.000Z",
		updated_at: "2026-07-21T10:00:00.000Z",
	}]);
	const transport = new GhCliDecisionTransport(async (file, args) => {
		calls.push({ file, args });
		if (args.includes("/user")) return JSON.stringify({ login: "shepherd-host" });
		if (args.includes("--method") && args.includes("POST")) return page.slice(1, -1);
		return page;
	});
	assert.equal(await transport.getAuthenticatedActor(), "shepherd-host");
	assert.equal((await transport.listComments(issueBinding)).length, 1);
	await transport.createComment(issueBinding, "bounded body");
	assert.ok(calls.every((call) => call.file === "gh"));
	assert.ok(calls.every((call) => !call.args.some((arg) => /token|authorization|bearer/i.test(arg))));
	assert.ok(calls.some((call) => call.args.includes("repos/polymetrics-ai/cli/issues/471/comments?per_page=100&page=1")));
});

test("live GitHub comment test is skipped without an explicitly designated sandbox", { skip: !process.env.PM_SHEPHERD_GITHUB_SANDBOX }, () => {
	assert.fail("Live sandbox mutation is intentionally not implemented in the default test run");
});

test("file repository type remains usable by the broker without a live transport", () => {
	assert.equal(typeof FileHumanDecisionRepository, "function");
});
