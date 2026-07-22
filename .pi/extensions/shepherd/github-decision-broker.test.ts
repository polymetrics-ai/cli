import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

import {
	FileHumanDecisionRepository,
	createHumanDecisionRecord,
	type HumanDecisionBinding,
	type HumanDecisionRecord,
	type HumanDecisionRepository,
	type HumanDecisionTransaction,
} from "./human-decision.ts";
import {
	GhCliDecisionTransport,
	GitHubDecisionBroker,
	renderDecisionRequestComment,
	type GitHubComment,
	type GitHubDecisionTransport,
} from "./github-decision-broker.ts";

const fixture = JSON.parse(readFileSync(
	".pi/extensions/shepherd/fixtures/issue-477/github-comments.json",
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

type Cycle6ReadRecord = GitHubDecisionBroker["readRecord"];

class MemoryRepository implements HumanDecisionRepository {
	readonly states = new Map<string, HumanDecisionRecord>();
	private queue = Promise.resolve();

	async transact<T>(requestId: string, operation: (state: HumanDecisionRecord | null) => Promise<HumanDecisionTransaction<T>> | HumanDecisionTransaction<T>): Promise<T> {
		let resolveResult!: (value: T) => void;
		let rejectResult!: (error: unknown) => void;
		const result = new Promise<T>((resolve, reject) => { resolveResult = resolve; rejectResult = reject; });
		this.queue = this.queue.then(async () => {
			try {
				const update = await operation(structuredClone(this.states.get(requestId) ?? null));
				if (update.state === null) this.states.delete(requestId);
				else this.states.set(requestId, structuredClone(update.state));
				resolveResult(update.value);
			} catch (error) {
				rejectResult(error);
			}
		});
		await result;
		return result;
	}

	async load(requestId: string): Promise<HumanDecisionRecord | null> {
		return structuredClone(this.states.get(requestId) ?? null);
	}
}

class FakeTransport implements GitHubDecisionTransport {
	readonly created: Array<{ binding: HumanDecisionBinding; body: string }> = [];
	readonly listed: HumanDecisionBinding[] = [];
	comments: GitHubComment[] = [];
	authenticatedActor = "shepherd-host";
	createFailure?: unknown;
	createFailureAfterPublish?: unknown;
	listFailures: unknown[] = [];
	requestCommentId = 1001;
	requestCommentTimestamp = "2026-07-21T10:00:00.000Z";

	async getAuthenticatedActor(): Promise<string> { return this.authenticatedActor; }
	async listComments(binding: HumanDecisionBinding): Promise<GitHubComment[]> {
		this.listed.push(structuredClone(binding));
		if (this.listFailures.length > 0) throw this.listFailures.shift();
		return structuredClone(this.comments);
	}
	async createDecisionRequestComment(record: HumanDecisionRecord): Promise<GitHubComment> {
		if (this.createFailure) throw this.createFailure;
		const binding = record.binding;
		const body = renderDecisionRequestComment(record);
		this.created.push({ binding: structuredClone(binding), body });
		const comment: GitHubComment = {
			id: this.requestCommentId,
			url: `https://github.com/${binding.repository}/${binding.target.kind === "issue" ? "issues" : "pull"}/${binding.target.number}#issuecomment-${this.requestCommentId}`,
			body,
			actor: { login: this.authenticatedActor, type: "User" },
			createdAt: this.requestCommentTimestamp,
			updatedAt: this.requestCommentTimestamp,
		};
		this.comments.push(comment);
		if (this.createFailureAfterPublish) {
			const failure = this.createFailureAfterPublish;
			this.createFailureAfterPublish = undefined;
			throw failure;
		}
		return structuredClone(comment);
	}
}

function brokerHarness(
	binding = issueBinding,
	overrides: {
		polling?: { maxAttempts: number; initialDelayMs: number; maxDelayMs: number };
		transportRetry?: { maxAttempts: number; initialDelayMs: number; maxDelayMs: number };
	} = {},
) {
	const repository = new MemoryRepository();
	const transport = new FakeTransport();
	let now = new Date("2026-07-21T10:00:00.000Z");
	const sleeps: number[] = [];
	const broker = new GitHubDecisionBroker(repository, transport, {
		now: () => now,
		sleep: async (delayMs) => { sleeps.push(delayMs); now = new Date(now.valueOf() + delayMs); },
		polling: overrides.polling ?? { maxAttempts: 3, initialDelayMs: 10, maxDelayMs: 15 },
		transportRetry: overrides.transportRetry,
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

function classifiedTransportFailure(retryable: boolean, marker: string): Error & { retryable: boolean } {
	return Object.assign(new Error(`raw transport failure ${marker}`), { retryable });
}

function consumedRecordAt(
	request: ReturnType<typeof brokerHarness>["request"],
	mode: "creation" | "request_comment" | "decision" | "consumption" | "update" | "all",
): HumanDecisionRecord {
	const record = createHumanDecisionRecord({
		requestId: request.requestId,
		gate: request.gate,
		binding: issueBinding,
		allowedOptions: request.allowedOptions,
		actorAllowlist: request.actorAllowlist,
		expiresAt: request.expiresAt,
		question: request.question,
	}, new Date("2026-07-21T10:00:00.000Z"));
	const future = "2026-07-21T10:05:02.000Z";
	const requestComment = {
		id: 1001,
		url: "https://github.com/polymetrics-ai/cli/issues/471#issuecomment-1001",
		actor: "shepherd-host",
		createdAt: mode === "creation" || mode === "request_comment" || mode === "all"
			? future
			: "2026-07-21T10:00:10.000Z",
	};
	const decision = {
		option: "approve",
		actor: "maintainer-one",
		sourceUrl: "https://github.com/polymetrics-ai/cli/issues/471#issuecomment-1002",
		decidedAt: ["creation", "request_comment", "decision", "all"].includes(mode)
			? "2026-07-21T10:05:03.000Z"
			: "2026-07-21T10:01:00.000Z",
	};
	const consumedAt = ["creation", "request_comment", "decision", "consumption", "all"].includes(mode)
		? "2026-07-21T10:05:04.000Z"
		: "2026-07-21T10:02:00.000Z";
	return {
		...record,
		createdAt: mode === "creation" || mode === "all" ? future : record.createdAt,
		requestComment,
		status: "consumed",
		decision,
		consumedAt,
		updatedAt: mode === "update" ? future : consumedAt,
	};
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

test("cycle 8 real broker resumes an exact consumed request after expiry without reviving new decisions", async (t) => {
	const harness = brokerHarness();
	await harness.broker.request(harness.request);
	harness.setNow("2026-07-21T10:02:00.000Z");
	harness.transport.comments.push(fixture.allowlistedHuman);
	const polled = await harness.broker.poll(harness.request.requestId, issueBinding);
	assert.equal(polled.status, "decided");
	const consumedEvidence = await harness.broker.consume(harness.request.requestId, issueBinding);
	const consumedRecord = await harness.broker.readRecord(harness.request.requestId, issueBinding);
	assert.equal(consumedRecord.status, "consumed");
	harness.setNow("2026-07-22T10:00:01.000Z");
	const restarted = new GitHubDecisionBroker(harness.repository, harness.transport, {
		now: () => new Date("2026-07-22T10:00:01.000Z"),
	});

	await t.test("an exact durable replay returns the already prepared consumed state", async () => {
		const replayed = await restarted.request(harness.request);
		assert.deepEqual(replayed, consumedRecord);
		assert.deepEqual(replayed.decision, consumedEvidence);
		assert.equal(replayed.requestComment?.id, harness.transport.requestCommentId);
		assert.equal(harness.transport.created.length, 1, "restart must not publish a second marker");
	});

	await t.test("the consumed decision remains one-shot after replay", async () => {
		await assert.rejects(
			restarted.consume(harness.request.requestId, issueBinding),
			/consum/i,
		);
		assert.deepEqual(await restarted.readRecord(harness.request.requestId, issueBinding), consumedRecord);
	});

	await t.test("a changed retry cannot hide behind the expired lifetime", async () => {
		await assert.rejects(
			restarted.request({
				...harness.request,
				question: "Approve a different exact gate after expiry?",
			}),
			/conflict|differs/i,
		);
		assert.equal(harness.transport.created.length, 1);
		assert.deepEqual(await restarted.readRecord(harness.request.requestId, issueBinding), consumedRecord);
	});

	await t.test("a genuinely new expired request is rejected before persistence or publication", async () => {
		const fresh = brokerHarness();
		fresh.setNow("2026-07-22T10:00:01.000Z");
		const requestId = "req-477-cycle8-new-expired";
		await assert.rejects(
			fresh.broker.request({ ...fresh.request, requestId }),
			/expired/i,
		);
		assert.equal(await fresh.repository.load(requestId), null);
		assert.equal(fresh.transport.created.length, 0);
	});

	await t.test("a response first observed at expiry cannot create a new decision", async () => {
		const fresh = brokerHarness();
		await fresh.broker.request(fresh.request);
		fresh.setNow("2026-07-22T10:00:00.000Z");
		fresh.transport.comments.push({
			...fixture.allowlistedHuman,
			createdAt: "2026-07-22T10:00:00.000Z",
			updatedAt: "2026-07-22T10:00:00.000Z",
		});
		const result = await fresh.broker.poll(fresh.request.requestId, issueBinding);
		assert.equal(result.status, "expired");
		assert.equal((await fresh.broker.readRecord(fresh.request.requestId, issueBinding)).status, "expired");
		await assert.rejects(fresh.broker.consume(fresh.request.requestId, issueBinding), /expired|decid|ready/i);
	});
});

test("accepts a real safe-integer comment ID and second-resolution GitHub timestamp", async () => {
	const harness = brokerHarness();
	harness.transport.requestCommentId = 5_034_006_493;
	harness.transport.requestCommentTimestamp = "2026-07-21T10:00:00Z";
	harness.setNow("2026-07-21T10:00:00.750Z");
	const record = await harness.broker.request({
		...harness.request,
		expiresAt: "2026-07-22T10:00:00Z",
	});
	assert.equal(record.requestComment?.id, 5_034_006_493);
	assert.equal(record.requestComment?.createdAt, "2026-07-21T10:00:00.000Z");
});

test("serializes concurrent request creation to one external comment", async () => {
	const harness = brokerHarness();
	const results = await Promise.all([
		harness.broker.request(harness.request),
		harness.broker.request(harness.request),
		new GitHubDecisionBroker(harness.repository, harness.transport).request(harness.request),
	]);
	assert.equal(harness.transport.created.length, 1);
	assert.equal(new Set(results.map((result) => result.requestComment?.id)).size, 1);
});

test("recovers an exact owned marker after a crash before local comment persistence", async () => {
	const harness = brokerHarness();
	const original = await harness.broker.request(harness.request);
	const state = harness.repository.states.get("req-477")!;
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
		const state = harness.repository.states.get("req-477")!;
		delete state.requestComment;
		harness.repository.states.set("req-477", state);
		await assert.rejects(new GitHubDecisionBroker(harness.repository, harness.transport).request(harness.request), /marker|owner|collision|duplicate/i, variant);
	}
});

test("accepts only one exact unedited allowlisted human command and persists minimal evidence", async () => {
	const harness = brokerHarness();
	await harness.broker.request(harness.request);
	harness.setNow("2026-07-21T10:02:00.000Z");
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
		{ ...fixture.bot, id: 2020, actor: { login: "github-actions[bot]", type: "Bot" } },
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

test("allows a bot-authenticated host to own the marker while never treating it as a human response", async () => {
	const harness = brokerHarness();
	harness.transport.authenticatedActor = "shepherd-app[bot]";
	const record = await harness.broker.request(harness.request);
	assert.equal(record.requestComment?.actor, "shepherd-app[bot]");
	const result = await harness.broker.poll("req-477", issueBinding);
	assert.equal(result.status, "pending");
});

test("parent merge requires the exact approve-merge affirmative command", async () => {
	const invalid = brokerHarness(prBinding);
	await assert.rejects(invalid.broker.request({
		...invalid.request,
		gate: "parent_merge",
		allowedOptions: ["approve", "reject"],
	} as unknown as Parameters<GitHubDecisionBroker["request"]>[0]), /approve-merge/i);

	const harness = brokerHarness(prBinding);
	await harness.broker.request({
		...harness.request,
		gate: "parent_merge",
		allowedOptions: ["approve-merge", "reject"],
	});
	harness.setNow("2026-07-21T10:02:00.000Z");
	harness.transport.comments.push({
		...fixture.allowlistedHuman,
		url: "https://github.com/polymetrics-ai/cli/pull/477#issuecomment-2001",
		body: "/shepherd decide req-477 approve-merge",
	});
	const result = await harness.broker.poll("req-477", prBinding);
	assert.equal(result.status, "decided");
	assert.equal(result.decision?.option, "approve-merge");
});

test("escapes untrusted Markdown structure and safely mentions configured humans", () => {
	const record = createHumanDecisionRecord({
		requestId: "req-display",
		gate: "requirements",
		binding: issueBinding,
		allowedOptions: ["approve", "reject"],
		actorAllowlist: ["maintainer-one", "maintainer2"],
		expiresAt: "2026-07-22T10:00:00.000Z",
		question: "### Fake approval\n- merged\n<details>spoof</details>",
	}, new Date("2026-07-21T10:00:00.000Z"));
	const body = renderDecisionRequestComment(record);
	assert.doesNotMatch(body, /\n### Fake approval|\n- merged\n|<details>/);
	assert.match(body, /> \\#\\#\\# Fake approval/);
	assert.match(body, /Allowed humans: @maintainer-one, @maintainer2/);
});

test("duplicate or conflicting valid commands are ambiguous and fail closed", async () => {
	for (const secondOption of ["approve", "reject"] as const) {
		const harness = brokerHarness();
		await harness.broker.request(harness.request);
		harness.setNow("2026-07-21T10:02:00.000Z");
		harness.transport.comments.push(
			fixture.allowlistedHuman,
			{ ...fixture.allowlistedHuman, id: 2099, url: `${fixture.allowlistedHuman.url}-duplicate`, body: `/shepherd decide req-477 ${secondOption}` },
		);
		await assert.rejects(harness.broker.poll("req-477", issueBinding), /ambiguous|multiple/i);
	}
});

test("a duplicate marker introduced after request creation makes polling fail closed", async () => {
	const harness = brokerHarness();
	await harness.broker.request(harness.request);
	const marker = harness.transport.comments[0];
	harness.transport.comments.push({ ...marker, id: marker.id + 1, url: `${marker.url}-duplicate` });
	await assert.rejects(harness.broker.poll("req-477", issueBinding), /marker|duplicat|authoritative/i);
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

test("does not accept an allowlisted command whose authoritative timestamp is still in the future", async () => {
	const harness = brokerHarness();
	await harness.broker.request(harness.request);
	harness.transport.comments.push({
		...fixture.allowlistedHuman,
		createdAt: "2026-07-21T11:00:00.000Z",
		updatedAt: "2026-07-21T11:00:00.000Z",
	});
	const result = await harness.broker.poll("req-477", issueBinding);
	assert.equal(result.status, "pending");
});

test("retries transient transport failures with bounded backoff", async () => {
	const harness = brokerHarness(issueBinding, {
		polling: { maxAttempts: 1, initialDelayMs: 10, maxDelayMs: 10 },
		transportRetry: { maxAttempts: 3, initialDelayMs: 2, maxDelayMs: 4 },
	});
	await harness.broker.request(harness.request);
	harness.transport.listFailures.push(classifiedTransportFailure(true, "transient-marker"));
	const result = await harness.broker.poll("req-477", issueBinding);
	assert.equal(result.status, "pending");
	assert.deepEqual(harness.sleeps, [2]);
});

test("recovers a transient create response without publishing a duplicate marker", async () => {
	const harness = brokerHarness(issueBinding, {
		transportRetry: { maxAttempts: 3, initialDelayMs: 2, maxDelayMs: 4 },
	});
	harness.transport.createFailureAfterPublish = classifiedTransportFailure(true, "post-create-marker");
	const record = await harness.broker.request(harness.request);
	assert.equal(harness.transport.created.length, 1);
	assert.equal(harness.transport.comments.length, 1);
	assert.equal(record.requestComment?.id, harness.transport.requestCommentId);
	assert.deepEqual(harness.sleeps, [2]);
});

test("caps transient transport retries and backoff", async () => {
	const harness = brokerHarness(issueBinding, {
		polling: { maxAttempts: 1, initialDelayMs: 10, maxDelayMs: 10 },
		transportRetry: { maxAttempts: 3, initialDelayMs: 2, maxDelayMs: 4 },
	});
	await harness.broker.request(harness.request);
	harness.transport.listFailures.push(
		classifiedTransportFailure(true, "transient-one"),
		classifiedTransportFailure(true, "transient-two"),
		classifiedTransportFailure(true, "transient-three"),
	);
	await assert.rejects(harness.broker.poll("req-477", issueBinding), (error: unknown) => {
		assert.ok(error instanceof Error);
		assert.equal((error as Error & { retryable?: boolean }).retryable, true);
		assert.equal(error.message, "GitHub transport transient failure");
		return true;
	});
	assert.deepEqual(harness.sleeps, [2, 4]);
});

test("fails permanent transport errors immediately with a redacted classification", async () => {
	const harness = brokerHarness(issueBinding, {
		polling: { maxAttempts: 1, initialDelayMs: 10, maxDelayMs: 10 },
		transportRetry: { maxAttempts: 3, initialDelayMs: 2, maxDelayMs: 4 },
	});
	await harness.broker.request(harness.request);
	harness.transport.listFailures.push(classifiedTransportFailure(false, "permanent-marker"));
	await assert.rejects(harness.broker.poll("req-477", issueBinding), (error: unknown) => {
		assert.ok(error instanceof Error);
		assert.equal(error.message, "GitHub transport permanent failure");
		assert.equal((error as Error & { retryable?: boolean }).retryable, false);
		assert.doesNotMatch(error.message, /permanent-marker/);
		return true;
	});
	assert.deepEqual(harness.sleeps, []);
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
		id: 5_034_006_493,
		html_url: "https://github.com/polymetrics-ai/cli/issues/471#issuecomment-5034006493",
		body: "safe",
		user: { login: "maintainer2", type: "User" },
		created_at: "2026-07-21T10:00:00Z",
		updated_at: "2026-07-21T10:00:00Z",
	}]);
	const transport = new GhCliDecisionTransport(async (file, args) => {
		calls.push({ file, args });
		if (args.includes("/user")) return JSON.stringify({ login: "shepherd-host" });
		if (args.includes("--method") && args.includes("POST")) return page.slice(1, -1);
		return page;
	});
	assert.equal(await transport.getAuthenticatedActor(), "shepherd-host");
	assert.equal((await transport.listComments(issueBinding)).length, 1);
	await transport.createDecisionRequestComment(createHumanDecisionRecord({
		requestId: "req-transport",
		gate: "requirements",
		binding: issueBinding,
		allowedOptions: ["approve", "reject"],
		actorAllowlist: ["maintainer-one"],
		expiresAt: "2026-07-22T10:00:00.000Z",
		question: "Approve the transport request?",
	}, new Date("2026-07-21T10:00:00.000Z")));
	assert.ok(calls.every((call) => call.file === "gh"));
	assert.ok(calls.every((call) => !call.args.some((arg) => /token|authorization|bearer/i.test(arg))));
	assert.ok(calls.some((call) => call.args.includes("repos/polymetrics-ai/cli/issues/471/comments?per_page=100&page=1")));
});

test("typed gh adapter classifies transient and permanent executor failures without leaking raw text", async () => {
	for (const [status, retryable] of [[503, true], [401, false]] as const) {
		const marker = `raw-adapter-marker-${status}`;
		const transport = new GhCliDecisionTransport(async () => {
			throw Object.assign(new Error(`executor failed ${marker}`), {
				code: 1,
				stderr: `HTTP ${status}: ${marker}`,
			});
		});
		await assert.rejects(transport.listComments(issueBinding), (error: unknown) => {
			assert.ok(error instanceof Error);
			assert.equal((error as Error & { retryable?: boolean }).retryable, retryable);
			assert.equal(error.message, `GitHub transport ${retryable ? "transient" : "permanent"} failure`);
			assert.doesNotMatch(error.message, new RegExp(marker));
			return true;
		});
	}
});

test("typed gh adapter fails closed when pagination exceeds its fixed window or binding is hostile", async () => {
	const raw = {
		id: 7,
		html_url: "https://github.com/polymetrics-ai/cli/issues/471#issuecomment-7",
		body: "safe",
		user: { login: "maintainer-one", type: "User" },
		created_at: "2026-07-21T10:00:00.000Z",
		updated_at: "2026-07-21T10:00:00.000Z",
	};
	let calls = 0;
	const transport = new GhCliDecisionTransport(async () => {
		calls += 1;
		return JSON.stringify(Array.from({ length: 100 }, (_, index) => ({ ...raw, id: raw.id + index })));
	});
	await assert.rejects(transport.listComments(issueBinding), /bounded pagination/i);
	assert.equal(calls, 10);
	await assert.rejects(
		transport.listComments({ ...issueBinding, repository: "owner/repo;--method=POST" }),
		/repository/i,
	);
	assert.equal(calls, 10);
});

test("live GitHub comment test is skipped without an explicitly designated sandbox", { skip: !process.env.PM_SHEPHERD_GITHUB_SANDBOX }, () => {
	assert.fail("Live sandbox mutation is intentionally not implemented in the default test run");
});

test("file repository type remains usable by the broker without a live transport", () => {
	assert.equal(typeof FileHumanDecisionRepository, "function");
});

test("cycle 6 rereads the broker-owned canonical record across compact poll and evidence consume", async () => {
	const harness = brokerHarness();
	const readRecord: Cycle6ReadRecord = harness.broker.readRecord.bind(harness.broker);
	const requested = await harness.broker.request(harness.request);
	assert.ok(requested.requestComment);

	const pending = await readRecord(harness.request.requestId, issueBinding);
	assert.equal(pending.status, "pending");
	pending.question = "mutated caller copy";
	assert.equal((await readRecord(harness.request.requestId, issueBinding)).question, harness.request.question);

	harness.setNow("2026-07-21T10:02:00.000Z");
	harness.transport.comments.push(fixture.allowlistedHuman);
	const polled = await harness.broker.poll(harness.request.requestId, issueBinding);
	assert.equal(polled.status, "decided");
	const decided = await readRecord(harness.request.requestId, issueBinding);
	assert.equal(decided.status, "decided");
	if (polled.status === "decided") assert.deepEqual(decided.decision, polled.decision);

	const evidence = await harness.broker.consume(harness.request.requestId, issueBinding);
	const consumed = await readRecord(harness.request.requestId, issueBinding);
	assert.equal(consumed.status, "consumed");
	assert.deepEqual(consumed.decision, evidence);
	assert.ok(consumed.requestComment);
	assert.ok(consumed.consumedAt);
	await assert.rejects(
		readRecord(harness.request.requestId, { ...issueBinding, generation: issueBinding.generation + 1 }),
		/binding|stale|generation/i,
	);
});

test("cycle 7 real broker rejects every future durable chronology coordinate on owned reread", async (t) => {
	for (const mode of ["creation", "request_comment", "decision", "consumption", "update", "all"] as const) {
		await t.test(mode, async () => {
			const harness = brokerHarness();
			harness.setNow("2026-07-21T10:05:00.000Z");
			harness.repository.states.set(harness.request.requestId, consumedRecordAt(harness.request, mode));
			await assert.rejects(
				harness.broker.readRecord(harness.request.requestId, issueBinding),
				/future|observation|clock|chronology|timestamp/i,
			);
		});
	}
});

test("cycle 7 broker rejects finite schema credentials before persistence or comment publication", async (t) => {
	const samples = [
		"client-key-data: SYNTHETIC_KUBERNETES_KEY_DATA",
		"token: SYNTHETIC_KUBERNETES_TOKEN",
		'{"auth":"SYNTHETIC_DOCKER_AUTH"}',
		'{"identitytoken":"SYNTHETIC_DOCKER_IDENTITY_TOKEN"}',
		"aws_access_key_id = SYNTHETIC_AWS_ACCESS_KEY_ID",
		"aws_secret_access_key = SYNTHETIC_AWS_SECRET_ACCESS_KEY",
		"aws_session_token = SYNTHETIC_AWS_SESSION_TOKEN",
		"ASIAABCDEFGHIJKLMNOP",
	];
	for (const [index, question] of samples.entries()) {
		await t.test(`schema form ${index + 1}`, async () => {
			const harness = brokerHarness();
			let rejection: unknown;
			try { await harness.broker.request({ ...harness.request, question }); } catch (error) { rejection = error; }
			assert.ok(rejection instanceof Error);
			assert.match(rejection.message, /credential|secret|sensitive/i);
			assert.doesNotMatch(rejection.message, /SYNTHETIC_/u);
			assert.equal(harness.repository.states.size, 0);
			assert.equal(harness.transport.created.length, 0);
		});
	}
});

test("cycle 8 provider-neutral credential suffixes close the real broker request boundary", async (t) => {
	const suffixAssignments = [
		"UNLISTED_ALPHA_AUTHORIZATION=SYNTHETIC_CYCLE8_AUTHORIZATION_MARKER",
		"UNLISTED_BRAVO_TOKEN=SYNTHETIC_CYCLE8_TOKEN_MARKER",
		"UNLISTED_CHARLIE_ACCESS_TOKEN=SYNTHETIC_CYCLE8_ACCESS_TOKEN_MARKER",
		"UNLISTED_DELTA_REFRESH_TOKEN=SYNTHETIC_CYCLE8_REFRESH_TOKEN_MARKER",
		"UNLISTED_ECHO_API_KEY=SYNTHETIC_CYCLE8_API_KEY_MARKER",
		"UNLISTED_FOXTROT_PASSWORD=SYNTHETIC_CYCLE8_PASSWORD_MARKER",
		"UNLISTED_GOLF_SECRET=SYNTHETIC_CYCLE8_SECRET_MARKER",
		"UNLISTED_HOTEL_CLIENT_SECRET=SYNTHETIC_CYCLE8_CLIENT_SECRET_MARKER",
		"UNLISTED_INDIA_PRIVATE_KEY=SYNTHETIC_CYCLE8_PRIVATE_KEY_MARKER",
		"UNLISTED_JULIET_DATABASE_URL=SYNTHETIC_CYCLE8_DATABASE_URL_MARKER",
		"UNLISTED_KILO_CREDENTIAL=SYNTHETIC_CYCLE8_CREDENTIAL_MARKER",
		"UNLISTED_LIMA_CREDENTIALS=SYNTHETIC_CYCLE8_CREDENTIALS_MARKER",
		"UNLISTED_MIKE_COOKIE=SYNTHETIC_CYCLE8_COOKIE_MARKER",
		"UNLISTED_NOVEMBER_COOKIES=SYNTHETIC_CYCLE8_COOKIES_MARKER",
		"UNLISTED_OSCAR_SET_COOKIE=SYNTHETIC_CYCLE8_SET_COOKIE_MARKER",
		"UNLISTED_PAPA_SESSION=SYNTHETIC_CYCLE8_SESSION_MARKER",
		"UNLISTED_QUEBEC_SESSION_ID=SYNTHETIC_CYCLE8_SESSION_ID_MARKER",
		"UNLISTED_ROMEO_SESSION_TOKEN=SYNTHETIC_CYCLE8_SESSION_TOKEN_MARKER",
		"UNLISTED_SIERRA_SESSION_COOKIE=SYNTHETIC_CYCLE8_SESSION_COOKIE_MARKER",
		"UNLISTED_TANGO_CSRF_TOKEN=SYNTHETIC_CYCLE8_CSRF_TOKEN_MARKER",
	];
	const finiteSchemaAssignments = [
		"client-key-data: SYNTHETIC_CYCLE8_KUBERNETES_KEY_DATA",
		"token: SYNTHETIC_CYCLE8_KUBERNETES_TOKEN",
		'{"auth":"SYNTHETIC_CYCLE8_DOCKER_AUTH"}',
		'{"identitytoken":"SYNTHETIC_CYCLE8_DOCKER_IDENTITY_TOKEN"}',
		"aws_access_key_id = SYNTHETIC_CYCLE8_AWS_ACCESS_KEY_ID",
		"aws_secret_access_key = SYNTHETIC_CYCLE8_AWS_SECRET_ACCESS_KEY",
		"aws_session_token = SYNTHETIC_CYCLE8_AWS_SESSION_TOKEN",
		"ASIAABCDEFGHIJKLMNOP",
	];

	await t.test("classifies every recognized suffix with an unknown provider prefix", async () => {
		for (const question of suffixAssignments) {
			const harness = brokerHarness();
			await assert.rejects(
				harness.broker.request({ ...harness.request, question }),
				/credential|secret|sensitive/i,
				question,
			);
			assert.equal(harness.repository.states.size, 0, question);
			assert.equal(harness.transport.created.length, 0, question);
		}
	});

	await t.test("rejects without reflecting the classified synthetic value", async () => {
		for (const question of suffixAssignments) {
			const harness = brokerHarness();
			const marker = question.slice(question.indexOf("=") + 1);
			let rejection: unknown;
			try {
				await harness.broker.request({ ...harness.request, question });
			} catch (error) {
				rejection = error;
			}
			assert.ok(rejection instanceof Error, question);
			assert.match(rejection.message, /credential|secret|sensitive/i, question);
			assert.doesNotMatch(rejection.message, new RegExp(marker), question);
			assert.equal(harness.repository.states.size, 0, question);
			assert.equal(harness.transport.created.length, 0, question);
		}
	});

	await t.test("allows only the exact documented public FEATURE_TOKEN field", async () => {
		const publicHarness = brokerHarness();
		const publicQuestion = "FEATURE_TOKEN=non-sensitive-build-label";
		const created = await publicHarness.broker.request({ ...publicHarness.request, question: publicQuestion });
		assert.equal(created.question, publicQuestion);
		assert.equal(publicHarness.repository.states.size, 1);
		assert.equal(publicHarness.transport.created.length, 1);

		const nearbyHarness = brokerHarness();
		await assert.rejects(
			nearbyHarness.broker.request({
				...nearbyHarness.request,
				question: "UNLISTED_FEATURE_TOKEN=SYNTHETIC_CYCLE8_NEARBY_MARKER",
			}),
			/credential|secret|sensitive/i,
		);
		assert.equal(nearbyHarness.repository.states.size, 0);
		assert.equal(nearbyHarness.transport.created.length, 0);
	});

	await t.test("retains the finite Kubernetes Docker and AWS forms", async () => {
		for (const question of finiteSchemaAssignments) {
			const harness = brokerHarness();
			await assert.rejects(
				harness.broker.request({ ...harness.request, question }),
				/credential|secret|sensitive/i,
				question,
			);
			assert.equal(harness.repository.states.size, 0, question);
			assert.equal(harness.transport.created.length, 0, question);
		}
	});
});
