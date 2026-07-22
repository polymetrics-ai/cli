import assert from "node:assert/strict";
import test from "node:test";

import { createHumanDecisionRecord, type HumanDecisionRecord } from "./human-decision.ts";
import type {
	ExternalCallContext,
	ParentDecisionBroker,
} from "./github-orchestrator.ts";
import {
	ProductionHumanParentMergeGate,
	buildProductionChildInterventionDecisionRequest,
	type AuthoritativeParentMergeState,
	type ParentPullRequestMergeLookup,
	type ProductionParentMergeRequest,
} from "./production-human-gate.ts";

const HEAD = "a".repeat(40);
const MERGE = "b".repeat(40);

function context(): ExternalCallContext {
	return {
		signal: new AbortController().signal,
		deadlineAt: new Date(Date.now() + 60_000).toISOString(),
		acknowledgeAbort() {},
	};
}

function request(): ProductionParentMergeRequest {
	return {
		requestId: "parent-merge-7",
		repository: "acme/widgets",
		parentIssue: 471,
		pullRequest: 472,
		generation: 7,
		headSha: HEAD,
		actorAllowlist: ["maintainer"],
		expiresAt: "2026-07-23T00:00:00.000Z",
		question: "Merge the exact verified parent head?",
	};
}

function decisionRecord(status: "pending" | "decided" | "consumed", option?: "approve-merge" | "reject"): HumanDecisionRecord {
	const value = createHumanDecisionRecord({
		requestId: "parent-merge-7",
		gate: "parent_merge",
		binding: {
			repository: "acme/widgets",
			target: { kind: "pull_request", number: 472 },
			generation: 7,
			headSha: HEAD,
		},
		allowedOptions: ["approve-merge", "reject"],
		actorAllowlist: ["maintainer"],
		expiresAt: "2026-07-23T00:00:00.000Z",
		question: "Merge the exact verified parent head?",
	}, new Date("2026-07-22T00:00:00.000Z"));
	value.status = status;
	value.requestComment = {
		id: 1,
		url: "https://github.com/acme/widgets/pull/472#issuecomment-1",
		actor: "shepherd-host",
		createdAt: "2026-07-22T00:00:00.000Z",
	};
	if (option !== undefined) {
		value.decision = {
			option,
			actor: "maintainer",
			sourceUrl: "https://github.com/acme/widgets/pull/472#issuecomment-2",
			decidedAt: "2026-07-22T00:00:01.000Z",
		};
	}
	if (status === "consumed") {
		value.consumedAt = "2026-07-22T00:00:02.000Z";
		value.updatedAt = value.consumedAt;
	} else if (status === "decided") value.updatedAt = "2026-07-22T00:00:01.000Z";
	return value;
}

class Broker implements ParentDecisionBroker {
	record = decisionRecord("pending");
	requests: unknown[] = [];
	consumeCount = 0;

	async request(value: unknown): Promise<HumanDecisionRecord> {
		this.requests.push(value);
		return this.record;
	}

	async poll(): Promise<HumanDecisionRecord> { return this.record; }

	async consume(): Promise<HumanDecisionRecord> {
		this.consumeCount += 1;
		this.record = decisionRecord("consumed", "approve-merge");
		return this.record;
	}
}

class Lookup implements ParentPullRequestMergeLookup {
	state: AuthoritativeParentMergeState = {
		repository: "acme/widgets",
		pullRequest: 472,
		headSha: HEAD,
		state: "open",
		mergedAt: null,
		mergeCommitSha: null,
		revision: 1,
		observedAt: "2026-07-22T00:00:03.000Z",
	};
	async observeExactPullRequest(): Promise<AuthoritativeParentMergeState> { return this.state; }
}

test("requests the distinct exact repository/PR/generation/head parent merge gate", async () => {
	const broker = new Broker();
	const gate = new ProductionHumanParentMergeGate(broker, new Lookup());
	await gate.request(request(), context());
	assert.deepEqual(broker.requests, [{
		...request(),
		gate: "parent_merge",
		allowedOptions: ["approve-merge", "reject"],
	}]);
});

test("approval is not completion until GitHub authoritatively reports the exact head merged", async () => {
	const broker = new Broker();
	broker.record = decisionRecord("decided", "approve-merge");
	const lookup = new Lookup();
	const gate = new ProductionHumanParentMergeGate(broker, lookup);
	assert.equal((await gate.observe(request(), context())).status, "approved_waiting_for_merge");
	lookup.state = {
		...lookup.state,
		state: "merged",
		mergedAt: "2026-07-22T00:00:04.000Z",
		mergeCommitSha: MERGE,
		revision: 2,
	};
	const observed = await gate.observe(request(), context());
	assert.equal(observed.status, "merged");
	if (observed.status === "merged") {
		assert.equal(observed.headSha, HEAD);
		assert.equal(observed.mergeCommitSha, MERGE);
	}
});

test("rejects stale head, unauthorized broker shape, and ambiguous authoritative observations", async () => {
	const broker = new Broker();
	broker.record = decisionRecord("decided", "approve-merge");
	const lookup = new Lookup();
	const gate = new ProductionHumanParentMergeGate(broker, lookup);
	lookup.state = { ...lookup.state, headSha: "c".repeat(40) };
	await assert.rejects(gate.observe(request(), context()), /head/i);

	broker.record = { ...decisionRecord("decided", "approve-merge"), binding: {
		...decisionRecord("decided", "approve-merge").binding,
		generation: 8,
	} };
	await assert.rejects(new ProductionHumanParentMergeGate(broker, new Lookup()).observe(request(), context()), /binding|generation|marker/i);
});

test("builds an issue-bound child intervention that also works before PR publication", () => {
	const built = buildProductionChildInterventionDecisionRequest({
		requestId: "child-a-retry-3",
		repository: "acme/widgets",
		childIssue: 42,
		generation: 3,
		reason: "retry_budget_exhausted",
		actorAllowlist: ["maintainer"],
		expiresAt: "2026-07-23T00:00:00.000Z",
		question: "Authorize exactly one additional attempt?",
	});
	assert.equal(built.gate, "scope");
	assert.equal(built.parentIssue, 42);
	assert.equal(built.pullRequest, 42);
	assert.equal(built.headSha, undefined);
	assert.deepEqual(built.allowedOptions, ["authorize-one-retry", "abort-child"]);
	assert.throws(() => buildProductionChildInterventionDecisionRequest({
		...({} as Parameters<typeof buildProductionChildInterventionDecisionRequest>[0]),
		reason: "operator_preference" as never,
	}), /exhausted|invalid/i);
});
