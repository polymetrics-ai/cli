import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFile } from "node:fs/promises";
import test from "node:test";

import {
	GitHubParentOrchestrator,
	createParentOrchestrationPlan,
	materializeChildRecord,
	selectReadyChildren,
	type ChildIntegrationReceipt,
	type CreateChildIssueRequest,
	type CreatePullRequestRequest,
	type GitHubChildIssue,
	type GitHubOrchestrationTransport,
	type GitHubPullRequestQuery,
	type GitHubRosterQuery,
	type GitHubRosterSnapshot,
	type IntegrateChildRequest,
	type MarkParentReadyRequest,
	type ParentDecisionBroker,
	type ParentDecisionPolicy,
	type ParentOrchestrationPlan,
	type PublishRosterRequest,
	type PullRequestMarkerQuery,
} from "./github-orchestrator.ts";
import type { GitHubPullRequestEvidence } from "./github-evidence.ts";
import { createIndependentReviewWork } from "./review-router.ts";
import {
	createHumanDecisionRecord,
	type HumanDecisionEvidence,
	type HumanDecisionRecord,
} from "./human-decision.ts";
import type {
	GitHubDecisionPollOptions,
	GitHubDecisionPollResult,
	GitHubDecisionRequest,
} from "./github-decision-broker.ts";
import type {
	ClaimedWorkspace,
	WorkspaceHandoffEvidence,
} from "./workspace-adapter.ts";

const objectivePath = ".pi/extensions/shepherd/fixtures/issue-478/parent-objective.json";
const baseSha = "a".repeat(40);
const headSha = "b".repeat(40);

async function plan(): Promise<ParentOrchestrationPlan> {
	return createParentOrchestrationPlan(JSON.parse(await readFile(objectivePath, "utf8")));
}

function issueFrom(request: CreateChildIssueRequest, number = 811): GitHubChildIssue {
	return {
		number,
		marker: request.marker,
		title: request.title,
		body: request.body,
		state: "open",
		parentIssue: request.parentIssue,
	};
}

function cleanPullRequest(
	request: CreatePullRequestRequest,
	overrides: Partial<GitHubPullRequestEvidence> = {},
	number = 812,
): GitHubPullRequestEvidence {
	const review = {
		...createIndependentReviewWork({
			repository: request.repository,
			workItemId: request.workItemId,
			pullRequest: number,
			generation: request.generation,
			baseSha: request.baseSha,
			headSha: request.headSha,
			changedPaths: request.changedPaths,
			allowedScopes: request.allowedScopes,
		}),
		completedAt: "2026-07-21T12:00:00.000Z",
		verdict: "clean" as const,
		findings: [],
	};
	return {
		schemaVersion: 1,
		number,
		marker: request.marker,
		title: request.title,
		body: request.body,
		state: "open",
		draft: request.draft,
		baseBranch: request.baseBranch,
		headBranch: request.headBranch,
		baseSha: request.baseSha,
		headSha: request.headSha,
		changedPaths: [...request.changedPaths],
		mergeState: "clean",
		checksComplete: true,
		checks: [{
			id: "check-verify-1",
			name: "verify",
			producerId: "github-actions:workflow-verify",
			status: "completed",
			conclusion: "success",
			headSha: request.headSha,
			completedAt: "2026-07-21T11:55:00.000Z",
		}],
		requestedChanges: [],
		threads: [],
		reviews: [review],
		reviewsComplete: true,
		dispositions: [],
		observedAt: "2026-07-21T12:00:00.000Z",
		...overrides,
	};
}

function reviewDigest(review: GitHubPullRequestEvidence["reviews"][number]): string {
	return createHash("sha256").update(JSON.stringify({
		idempotencyMarker: review.idempotencyMarker,
		repository: review.repository,
		workItemId: review.workItemId,
		pullRequest: review.pullRequest,
		generation: review.generation,
		baseSha: review.baseSha,
		headSha: review.headSha,
		changedPaths: review.changedPaths,
		allowedScopes: review.allowedScopes,
		completedAt: review.completedAt,
		verdict: review.verdict,
		findings: review.findings,
	})).digest("hex");
}

function attestReview(review: GitHubPullRequestEvidence["reviews"][number]) {
	return {
		schemaVersion: 1,
		authority: "controller",
		sessionId: `session-${review.pullRequest}-${review.workItemId}`,
		runId: `run-${review.pullRequest}-${review.generation}`,
		provider: "openai-codex",
		model: "gpt-5.6-sol",
		reasoningEffort: "xhigh",
		readOnly: true,
		repository: review.repository,
		workItemId: review.workItemId,
		pullRequest: review.pullRequest,
		generation: review.generation,
		baseSha: review.baseSha,
		headSha: review.headSha,
		changedPaths: review.changedPaths,
		allowedScopes: review.allowedScopes,
		reviewMarker: review.idempotencyMarker,
		resultDigest: reviewDigest(review),
		completedAt: review.completedAt,
	};
}

class FakeTransport implements GitHubOrchestrationTransport {
	issues: GitHubChildIssue[] = [];
	pullRequests: GitHubPullRequestEvidence[] = [];
	rosters: GitHubRosterSnapshot[] = [];
	integrations: ChildIntegrationReceipt[] = [];
	createIssueCalls = 0;
	createPullRequestCalls = 0;
	publishRosterCalls = 0;
	integrateCalls = 0;
	markReadyCalls = 0;
	throwAfterIssuePublish = false;
	throwAfterPullRequestPublish = false;
	throwAfterRosterPublish = false;
	throwAfterIntegration = false;
	throwAfterReady = false;
	malformedPullRequestResponse = false;
	malformedIntegrationResponse = false;
	malformedReadyResponse = false;
	incompleteIssueLookup = false;
	incompletePullRequestLookup = false;
	incompleteRosterLookup = false;
	incompleteIntegrationLookup = false;
	ancestry = true;
	onPullRequestRead?: (query: PullRequestMarkerQuery, matches: GitHubPullRequestEvidence[], read: number) => GitHubPullRequestEvidence[];
	#pullRequestReads = 0;

	async findChildIssues(query: { repository: string; marker: string }): Promise<any> {
		return { items: this.issues.filter((candidate) => candidate.marker === query.marker), complete: !this.incompleteIssueLookup };
	}

	async createChildIssue(request: CreateChildIssueRequest): Promise<GitHubChildIssue> {
		this.createIssueCalls += 1;
		const created = issueFrom(request, 810 + this.createIssueCalls);
		this.issues.push(created);
		if (this.throwAfterIssuePublish) {
			this.throwAfterIssuePublish = false;
			throw new Error("simulated timeout after publish");
		}
		return created;
	}

	async findPullRequests(query: PullRequestMarkerQuery): Promise<GitHubPullRequestEvidence[]> {
		this.#pullRequestReads += 1;
		const matches = this.pullRequests.filter((candidate) => candidate.marker === query.marker);
		return { items: this.onPullRequestRead?.(query, matches, this.#pullRequestReads) ?? matches, complete: !this.incompletePullRequestLookup } as never;
	}

	async createPullRequest(request: CreatePullRequestRequest): Promise<GitHubPullRequestEvidence> {
		this.createPullRequestCalls += 1;
		const created = cleanPullRequest(request, {
			draft: request.draft,
			checks: [],
			reviews: [],
		}, 900 + this.createPullRequestCalls);
		this.pullRequests.push(created);
		if (this.throwAfterPullRequestPublish) {
			this.throwAfterPullRequestPublish = false;
			throw new Error("simulated timeout after pull request publish");
		}
		if (this.malformedPullRequestResponse) return { malformed: true } as never;
		return created;
	}

	async findParentRosters(query: GitHubRosterQuery): Promise<any> {
		return { items: this.rosters.filter((candidate) => candidate.marker === query.marker), complete: !this.incompleteRosterLookup };
	}

	async publishParentRoster(request: PublishRosterRequest): Promise<GitHubRosterSnapshot> {
		this.publishRosterCalls += 1;
		const existing = this.rosters.find((candidate) => candidate.marker === request.marker);
		const published: GitHubRosterSnapshot = {
			id: existing?.id ?? 1001,
			marker: request.marker,
			parentIssue: request.parentIssue,
			generation: request.generation,
			body: request.body,
			updatedAt: "2026-07-21T12:00:00.000Z",
		};
		if (existing) this.rosters.splice(this.rosters.indexOf(existing), 1, published);
		else this.rosters.push(published);
		if (this.throwAfterRosterPublish) {
			this.throwAfterRosterPublish = false;
			throw new Error("simulated timeout after roster publish");
		}
		return published;
	}

	async findChildIntegration(query: GitHubPullRequestQuery & { childId?: string; marker?: string }): Promise<any> {
		return {
			items: this.integrations.filter((candidate) => query.pullRequest !== undefined
				? candidate.pullRequest === query.pullRequest
				: candidate.childId === query.childId && candidate.marker === query.marker),
			complete: !this.incompleteIntegrationLookup,
		};
	}

	async integrateChild(request: IntegrateChildRequest): Promise<ChildIntegrationReceipt> {
		this.integrateCalls += 1;
		const receipt: ChildIntegrationReceipt = {
			childId: request.childId,
			pullRequest: request.pullRequest,
			generation: request.generation,
			marker: request.marker,
			baseSha: request.baseSha,
			headSha: request.headSha,
			parentBranch: request.parentBranch,
			integratedAt: "2026-07-21T12:01:00.000Z",
		};
		this.integrations.push(receipt);
		if (this.throwAfterIntegration) {
			this.throwAfterIntegration = false;
			throw new Error("simulated timeout after integration");
		}
		if (this.malformedIntegrationResponse) return { malformed: true } as never;
		return receipt;
	}

	async markParentReady(request: MarkParentReadyRequest): Promise<GitHubPullRequestEvidence> {
		this.markReadyCalls += 1;
		const index = this.pullRequests.findIndex((candidate) => candidate.number === request.pullRequest);
		if (index < 0) throw new Error("parent pull request missing");
		const updated = { ...this.pullRequests[index], draft: false, headSha: request.headSha };
		this.pullRequests.splice(index, 1, updated);
		if (this.throwAfterReady) {
			this.throwAfterReady = false;
			throw new Error("simulated timeout after ready transition");
		}
		if (this.malformedReadyResponse) return { malformed: true } as never;
		return updated;
	}

	async isAncestor(_query: { repository: string; ancestorSha: string; descendantSha: string }): Promise<boolean> {
		return this.ancestry;
	}
}

function orchestratorFor(transport: FakeTransport, broker?: ParentDecisionBroker): GitHubParentOrchestrator {
	const sessionSource = {
		async findAttestations(target: { pullRequest: number; workItemId: string }): Promise<unknown> {
			const attestations = transport.pullRequests
				.filter((pullRequest) => pullRequest.number === target.pullRequest)
				.flatMap((pullRequest) => pullRequest.reviews)
				.filter((review) => review.workItemId === target.workItemId)
				.map(attestReview);
			return { items: attestations, complete: true };
		},
	};
	return new GitHubParentOrchestrator(transport, broker, sessionSource as never);
}

const approvedDecision: HumanDecisionEvidence = {
	option: "approve-merge",
	actor: "maintainer",
	sourceUrl: "https://github.com/polymetrics-ai/cli/pull/900#issuecomment-1",
	decidedAt: "2026-07-21T12:00:30.000Z",
};

class FakeDecisionBroker implements ParentDecisionBroker {
	requests: GitHubDecisionRequest[] = [];
	consumes = 0;
	pollResult: GitHubDecisionPollResult = { status: "decided", decision: approvedDecision, attempts: 1 };
	recordStatus: HumanDecisionRecord["status"] = "pending";

	async request(request: GitHubDecisionRequest): Promise<HumanDecisionRecord> {
		this.requests.push(request);
		if (request.gate !== "parent_merge") throw new Error("fake parent broker accepts only parent_merge requests");
		const record = createHumanDecisionRecord({
			requestId: request.requestId,
			gate: request.gate,
			binding: {
				repository: request.repository,
				target: { kind: "pull_request", number: request.pullRequest },
				generation: request.generation,
				...(request.headSha ? { headSha: request.headSha } : {}),
			},
			allowedOptions: request.allowedOptions,
			actorAllowlist: request.actorAllowlist,
			expiresAt: request.expiresAt,
			question: request.question,
		}, new Date("2026-07-21T12:00:00.000Z"));
		return this.recordStatus === "consumed"
			? { ...record, status: "consumed", decision: approvedDecision, consumedAt: "2026-07-21T12:00:40.000Z" }
			: record;
	}

	async poll(_requestId: string, _binding: HumanDecisionRecord["binding"], _options?: GitHubDecisionPollOptions): Promise<GitHubDecisionPollResult> {
		return this.pollResult;
	}

	async consume(_requestId: string, _binding: HumanDecisionRecord["binding"]): Promise<HumanDecisionEvidence> {
		this.consumes += 1;
		return this.pollResult.status === "decided" ? this.pollResult.decision : approvedDecision;
	}
}

function childHandoff(issue: number, branch: string, prBase: string, overrides: Partial<WorkspaceHandoffEvidence> = {}): WorkspaceHandoffEvidence {
	return {
		issue,
		branch,
		prBase,
		baseHead: baseSha,
		head: headSha,
		changedScope: [".pi/extensions/shepherd/github-evidence.ts"],
		verificationState: "passed",
		repositoryIdentity: "1".repeat(64),
		worktreeIdentity: "2".repeat(64),
		dirty: false,
		...overrides,
	};
}

const decisionPolicy: ParentDecisionPolicy = {
	requestId: "parent-471-generation-3",
	actorAllowlist: ["maintainer"],
	expiresAt: "2027-07-21T12:00:00.000Z",
	question: "Approve the exact reviewed parent head for the human merge gate?",
};

test("turns a parent objective into bounded canonical child records and reuses DAG scheduling", async () => {
	const candidate = await plan();
	assert.equal(candidate.children.length, 2);
	assert.deepEqual(candidate.children.map((child) => child.id), ["evidence", "orchestrator"]);
	assert.deepEqual(candidate.children[0].branch, { kind: "canonical_issue_branch", slug: "github-evidence" });
	assert.equal(candidate.children[0].prBase, candidate.parentBranch);
	assert.match(candidate.children[0].markers.issue, /^<!-- shepherd-child-issue:v1:/);
	assert.match(candidate.children[0].markers.pullRequest, /^<!-- shepherd-child-pr:v1:/);
	assert.deepEqual(selectReadyChildren(candidate, { evidence: "pending", orchestrator: "pending" }, 2), {
		kind: "selected",
		itemIds: ["evidence"],
	});
	assert.deepEqual(selectReadyChildren(candidate, { evidence: "succeeded", orchestrator: "pending" }, 2), {
		kind: "selected",
		itemIds: ["orchestrator"],
	});
});

test("rejects extra fields, cycles, unsafe scopes/branches, missing requirements, and oversized rosters", async () => {
	const source = JSON.parse(await readFile(objectivePath, "utf8")) as Record<string, unknown>;
	assert.throws(() => createParentOrchestrationPlan({ ...source, unexpected: true }), /field|shape|parent/i);
	const children = source.children as Array<Record<string, unknown>>;
	assert.throws(() => createParentOrchestrationPlan({ ...source, children: [{ ...children[0], dependsOn: ["evidence"] }] }), /cycle/i);
	assert.throws(() => createParentOrchestrationPlan({ ...source, parentBranch: "../main" }), /branch/i);
	for (const invalidRef of ["bad branch", ".hidden/topic", "topic.lock", "topic/.lock", "@", "topic@{one", "topic."]) {
		assert.throws(() => createParentOrchestrationPlan({ ...source, parentBranch: invalidRef }), /branch|ref/i, invalidRef);
	}
	for (const invalidGeneration of [0, -1]) {
		assert.throws(() => createParentOrchestrationPlan({ ...source, generation: invalidGeneration }), /generation|positive/i);
	}
	assert.throws(() => createParentOrchestrationPlan({ ...source, children: [{ ...children[0], writeScopes: ["../outside"] }] }), /scope/i);
	assert.throws(() => createParentOrchestrationPlan({ ...source, children: [{ ...children[0], requiredSkills: [] }] }), /skills/i);
	assert.throws(() => createParentOrchestrationPlan({ ...source, children: Array.from({ length: 65 }, (_, index) => ({ ...children[0], id: `child-${index}`, slug: `child-${index}`, dependsOn: [] })) }), /64|bounded|children/i);
});

test("serializes concurrent issue ensures, reconciles after create, and fails closed on incomplete lookup", async () => {
	const candidate = await plan();
	const transport = new FakeTransport();
	const orchestrator = orchestratorFor(transport);
	const [first, second] = await Promise.all([
		orchestrator.ensureChildIssue(candidate, "evidence"),
		orchestrator.ensureChildIssue(candidate, "evidence"),
	]);
	assert.deepEqual(first, second);
	assert.equal(transport.createIssueCalls, 1);
	assert.equal(transport.issues.length, 1);

	const incomplete = new FakeTransport();
	incomplete.incompleteIssueLookup = true;
	await assert.rejects(orchestratorFor(incomplete).ensureChildIssue(candidate, "evidence"), /complete|pagination|authoritative/i);
	assert.equal(incomplete.createIssueCalls, 0);
});

test("recovers timeout-after-issue-publish and restart without duplicate mutation", async () => {
	const candidate = await plan();
	const transport = new FakeTransport();
	transport.throwAfterIssuePublish = true;
	const orchestrator = orchestratorFor(transport);
	const first = await orchestrator.ensureChildIssue(candidate, "evidence");
	const restarted = await orchestratorFor(transport).ensureChildIssue(candidate, "evidence");
	assert.deepEqual(restarted, first);
	assert.equal(transport.createIssueCalls, 1);
	assert.equal(transport.issues.length, 1);
});

test("fails closed on marker collision or duplicate matching resources", async () => {
	const candidate = await plan();
	const child = candidate.children[0];
	const transport = new FakeTransport();
	transport.issues.push({
		number: 811,
		marker: child.markers.issue,
		title: "unrelated title",
		body: "hostile collision",
		state: "open",
		parentIssue: candidate.parentIssue,
	});
	await assert.rejects(orchestratorFor(transport).ensureChildIssue(candidate, child.id), /marker|collision|mismatch/i);
	transport.issues.push({ ...transport.issues[0], number: 812 });
	await assert.rejects(orchestratorFor(transport).ensureChildIssue(candidate, child.id), /duplicate|ambiguous|marker/i);
	assert.equal(transport.createIssueCalls, 0);
});

test("creates a draft parent PR and stacked child PR with exclusive Closes versus Refs linkage", async () => {
	const candidate = await plan();
	const transport = new FakeTransport();
	const orchestrator = orchestratorFor(transport);
	const issue = await orchestrator.ensureChildIssue(candidate, "evidence");
	const child = materializeChildRecord(candidate, "evidence", issue);
	const childPr = await orchestrator.ensureChildPullRequest(candidate, child, childHandoff(issue.number, child.branch, child.prBase));
	assert.equal(childPr.baseBranch, candidate.parentBranch);
	assert.match(childPr.body, new RegExp(`Refs #${issue.number}`));
	assert.match(childPr.body, new RegExp(`Refs #${candidate.parentIssue}`));
	assert.doesNotMatch(childPr.body, /\bCloses\b/i);

	const parentPr = await orchestrator.ensureParentDraftPullRequest(candidate, childHandoff(
		candidate.parentIssue,
		candidate.parentBranch,
		candidate.parentBaseBranch,
		{ changedScope: [], head: "c".repeat(40) },
	));
	assert.equal(parentPr.draft, true);
	assert.equal(parentPr.baseBranch, candidate.parentBaseBranch);
	assert.match(parentPr.body, new RegExp(`Closes #${candidate.parentIssue}`));
	assert.doesNotMatch(parentPr.body, /Closes #811/);
	assert.equal(transport.createPullRequestCalls, 2);
});

test("serializes concurrent PR ensures and recovers timeout or malformed mutation responses only by complete re-read", async () => {
	const candidate = await plan();
	const transport = new FakeTransport();
	const orchestrator = orchestratorFor(transport);
	const issue = await orchestrator.ensureChildIssue(candidate, "evidence");
	const child = materializeChildRecord(candidate, "evidence", issue);
	const handoff = childHandoff(issue.number, child.branch, child.prBase);
	const [first, second] = await Promise.all([
		orchestrator.ensureChildPullRequest(candidate, child, handoff),
		orchestrator.ensureChildPullRequest(candidate, child, handoff),
	]);
	assert.deepEqual(first, second);
	assert.equal(transport.createPullRequestCalls, 1);
	assert.equal(transport.pullRequests.length, 1);

	for (const mode of ["timeout", "malformed"] as const) {
		const recovery = new FakeTransport();
		if (mode === "timeout") recovery.throwAfterPullRequestPublish = true;
		else recovery.malformedPullRequestResponse = true;
		const recoveryOrchestrator = orchestratorFor(recovery);
		const recoveryIssue = await recoveryOrchestrator.ensureChildIssue(candidate, "evidence");
		const recoveryChild = materializeChildRecord(candidate, "evidence", recoveryIssue);
		const recovered = await recoveryOrchestrator.ensureChildPullRequest(
			candidate,
			recoveryChild,
			childHandoff(recoveryIssue.number, recoveryChild.branch, recoveryChild.prBase),
		);
		assert.equal(recovered.number, 901);
		assert.equal(recovery.createPullRequestCalls, 1);
	}

	const incomplete = new FakeTransport();
	incomplete.incompletePullRequestLookup = true;
	const incompleteOrchestrator = orchestratorFor(incomplete);
	const incompleteIssue = issueFrom({
		repository: candidate.repository,
		parentIssue: candidate.parentIssue,
		marker: candidate.children[0].markers.issue,
		title: candidate.children[0].title,
		body: candidate.children[0].issueBody,
	});
	const incompleteChild = materializeChildRecord(candidate, "evidence", incompleteIssue);
	await assert.rejects(incompleteOrchestrator.ensureChildPullRequest(
		candidate,
		incompleteChild,
		childHandoff(incompleteIssue.number, incompleteChild.branch, incompleteChild.prBase),
	), /complete|pagination|authoritative/i);
	assert.equal(incomplete.createPullRequestCalls, 0);
});

test("captures upstream workspace handoff and rejects dirty, failed, mismatched, or out-of-scope evidence", async () => {
	const candidate = await plan();
	const issue = issueFrom({
		repository: candidate.repository,
		parentIssue: candidate.parentIssue,
		marker: candidate.children[0].markers.issue,
		title: candidate.children[0].title,
		body: candidate.children[0].issueBody,
	}, 811);
	const child = materializeChildRecord(candidate, "evidence", issue);
	const expected = childHandoff(issue.number, child.branch, child.prBase);
	let requestedState = "";
	const source = {
		async captureHandoff(_workspace: ClaimedWorkspace, verificationState: "passed"): Promise<WorkspaceHandoffEvidence> {
			requestedState = verificationState;
			return expected;
		},
	};
	const orchestrator = orchestratorFor(new FakeTransport());
	assert.deepEqual(await orchestrator.captureChildHandoff(candidate, child, {} as ClaimedWorkspace, source), expected);
	assert.equal(requestedState, "passed");
	for (const invalid of [
		{ ...expected, dirty: true },
		{ ...expected, verificationState: "failed" as const },
		{ ...expected, branch: "feat/unrelated" },
		{ ...expected, changedScope: ["cmd/pm/main.go"] },
	]) {
		await assert.rejects(orchestrator.captureChildHandoff(candidate, child, {} as ClaimedWorkspace, { captureHandoff: async () => invalid }), /handoff|dirty|verification|head|branch|scope/i);
	}
});

test("validates every immutable materialized-child field and canonical issue before any downstream transport", async () => {
	const candidate = await plan();
	const transport = new FakeTransport();
	const issue = issueFrom({
		repository: candidate.repository,
		parentIssue: candidate.parentIssue,
		marker: candidate.children[0].markers.issue,
		title: candidate.children[0].title,
		body: candidate.children[0].issueBody,
	}, 811);
	transport.issues.push(issue);
	const child = materializeChildRecord(candidate, "evidence", issue);
	const mutations: Array<[string, (value: any) => void]> = [
		["issue", (value) => { value.issue = 999; value.branch = "feat/999-github-evidence"; }],
		["branch", (value) => { value.branch = "feat/811-forged"; }],
		["prBase", (value) => { value.prBase = "evil/parent"; }],
		["title", (value) => { value.title = "forged title"; }],
		["objective", (value) => { value.objective = "forged objective"; }],
		["dependsOn", (value) => { value.dependsOn = ["forged"]; }],
		["status", (value) => { value.status = "running"; }],
		["access", (value) => { value.access = "read_only"; }],
		["writeScopes", (value) => { value.writeScopes = ["evil/scope"]; }],
		["requiredSkills", (value) => { value.requiredSkills = ["forged-skill"]; }],
		["verification", (value) => { value.verification = [{ id: "forged", kind: "test", description: "forged" }]; }],
		["humanGates", (value) => { value.humanGates = ["merge"]; }],
		["markers", (value) => { value.markers = { ...value.markers, pullRequest: "<!-- forged -->" }; }],
		["issueBody", (value) => { value.issueBody = "forged body"; }],
	];
	for (const [field, mutate] of mutations) {
		const forged = structuredClone(child);
		mutate(forged);
		let captures = 0;
		await assert.rejects(orchestratorFor(transport).captureChildHandoff(
			candidate,
			forged,
			{} as ClaimedWorkspace,
			{ captureHandoff: async () => { captures += 1; return childHandoff(forged.issue, forged.branch, forged.prBase); } },
		), /materialized|canonical|plan|issue|topology/i, field);
		assert.equal(captures, 0, field);
	}
});

test("captures parent workspace setup through the upstream handoff boundary", async () => {
	const candidate = await plan();
	const expected = childHandoff(candidate.parentIssue, candidate.parentBranch, candidate.parentBaseBranch, {
		changedScope: [".pi/extensions/shepherd/github-orchestrator.ts"],
	});
	let requestedState = "";
	const source = {
		async captureHandoff(_workspace: ClaimedWorkspace, verificationState: "passed"): Promise<WorkspaceHandoffEvidence> {
			requestedState = verificationState;
			return expected;
		},
	};
	const captured = await orchestratorFor(new FakeTransport()).captureParentHandoff(
		candidate,
		{} as ClaimedWorkspace,
		source,
	);
	assert.deepEqual(captured, expected);
	assert.equal(requestedState, "passed");
});

test("roster publication reconciles before update and recovers timeout-after-publish", async () => {
	const candidate = await plan();
	const transport = new FakeTransport();
	transport.throwAfterRosterPublish = true;
	const orchestrator = orchestratorFor(transport);
	const pending = { evidence: "pending", orchestrator: "pending" } as const;
	const first = await orchestrator.reconcileParentRoster(candidate, pending);
	const restarted = await orchestratorFor(transport).reconcileParentRoster(candidate, pending);
	assert.deepEqual(restarted, first);
	assert.equal(transport.publishRosterCalls, 1);
	await orchestrator.reconcileParentRoster(candidate, { evidence: "succeeded", orchestrator: "pending" });
	assert.equal(transport.publishRosterCalls, 2);
	assert.equal(transport.rosters.length, 1);
});

test("integrates only green reviewed exact-head scoped children and rechecks head immediately before mutation", async () => {
	const candidate = await plan();
	const transport = new FakeTransport();
	const orchestrator = orchestratorFor(transport);
	const issue = issueFrom({
		repository: candidate.repository,
		parentIssue: candidate.parentIssue,
		marker: candidate.children[0].markers.issue,
		title: candidate.children[0].title,
		body: candidate.children[0].issueBody,
	}, 811);
	const child = materializeChildRecord(candidate, "evidence", issue);
	const handoff = childHandoff(issue.number, child.branch, child.prBase);
	const request: CreatePullRequestRequest = {
		repository: candidate.repository,
		workItemId: child.id,
		generation: candidate.generation,
		marker: child.markers.pullRequest,
		title: child.title,
		body: `Refs #${issue.number}\nRefs #${candidate.parentIssue}\n\n${child.markers.pullRequest}`,
		draft: false,
		baseBranch: child.prBase,
		headBranch: child.branch,
		baseSha: handoff.baseHead,
		headSha: handoff.head,
		changedPaths: handoff.changedScope,
		allowedScopes: child.writeScopes,
	};
	transport.pullRequests.push(cleanPullRequest(request));
	const integrated = await orchestrator.integrateChild(candidate, child, handoff);
	assert.equal(integrated.kind, "integrated");
	assert.equal(transport.integrateCalls, 1);

	const movedTransport = new FakeTransport();
	movedTransport.pullRequests.push(cleanPullRequest(request));
	movedTransport.onPullRequestRead = (_query, matches, read) => read < 2
		? matches
		: matches.map((candidatePr) => ({ ...candidatePr, headSha: "c".repeat(40) }));
	const moved = await orchestratorFor(movedTransport).integrateChild(candidate, child, handoff);
	assert.equal(moved.kind, "blocked");
	if (moved.kind === "blocked") assert.ok(moved.blockers.includes("head_moved"));
	assert.equal(movedTransport.integrateCalls, 0);

	const outside = await orchestratorFor(transport).integrateChild(candidate, child, { ...handoff, changedScope: ["cmd/pm/main.go"] });
	assert.equal(outside.kind, "blocked");
});

test("integration recovers timeout and malformed mutation responses, but incomplete receipt lookup fails closed", async () => {
	const candidate = await plan();
	const issue = issueFrom({
		repository: candidate.repository,
		parentIssue: candidate.parentIssue,
		marker: candidate.children[0].markers.issue,
		title: candidate.children[0].title,
		body: candidate.children[0].issueBody,
	}, 811);
	const child = materializeChildRecord(candidate, "evidence", issue);
	const handoff = childHandoff(issue.number, child.branch, child.prBase);
	const request: CreatePullRequestRequest = {
		repository: candidate.repository,
		workItemId: child.id,
		generation: candidate.generation,
		marker: child.markers.pullRequest,
		title: child.title,
		body: `Refs #${issue.number}\nRefs #${candidate.parentIssue}\n\n${child.markers.pullRequest}`,
		draft: false,
		baseBranch: child.prBase,
		headBranch: child.branch,
		baseSha: handoff.baseHead,
		headSha: handoff.head,
		changedPaths: handoff.changedScope,
		allowedScopes: child.writeScopes,
	};
	for (const mode of ["timeout", "malformed"] as const) {
		const transport = new FakeTransport();
		transport.issues.push(issue);
		transport.pullRequests.push(cleanPullRequest(request));
		if (mode === "timeout") transport.throwAfterIntegration = true;
		else transport.malformedIntegrationResponse = true;
		const result = await orchestratorFor(transport).integrateChild(candidate, child, handoff);
		assert.equal(result.kind, "integrated", mode);
		if (result.kind === "integrated") assert.equal(result.reused, true, mode);
		assert.equal(transport.integrateCalls, 1, mode);
		assert.equal(transport.integrations.length, 1, mode);
	}

	const incomplete = new FakeTransport();
	incomplete.issues.push(issue);
	incomplete.pullRequests.push(cleanPullRequest(request));
	incomplete.incompleteIntegrationLookup = true;
	await assert.rejects(orchestratorFor(incomplete).integrateChild(candidate, child, handoff), /complete|pagination|authoritative/i);
	assert.equal(incomplete.integrateCalls, 0);
});

test("review coverage must bind the planned repository, child, generation, paths, and scopes", async () => {
	const candidate = await plan();
	const issue = issueFrom({
		repository: candidate.repository,
		parentIssue: candidate.parentIssue,
		marker: candidate.children[0].markers.issue,
		title: candidate.children[0].title,
		body: candidate.children[0].issueBody,
	}, 811);
	const child = materializeChildRecord(candidate, "evidence", issue);
	const handoff = childHandoff(issue.number, child.branch, child.prBase);
	const request: CreatePullRequestRequest = {
		repository: candidate.repository,
		workItemId: child.id,
		generation: candidate.generation,
		marker: child.markers.pullRequest,
		title: child.title,
		body: `Refs #${issue.number}\nRefs #${candidate.parentIssue}\n\n${child.markers.pullRequest}`,
		draft: false,
		baseBranch: child.prBase,
		headBranch: child.branch,
		baseSha: handoff.baseHead,
		headSha: handoff.head,
		changedPaths: handoff.changedScope,
		allowedScopes: child.writeScopes,
	};
	for (const targetChanges of [
		{ repository: "other/cli" },
		{ workItemId: "other-child" },
		{ generation: candidate.generation + 1 },
		{ changedPaths: [] },
		{ allowedScopes: [".pi/extensions/shepherd"] },
	]) {
		const transport = new FakeTransport();
		const pr = cleanPullRequest(request);
		pr.reviews = [{
			...createIndependentReviewWork({
				repository: candidate.repository,
				workItemId: child.id,
				pullRequest: pr.number,
				generation: candidate.generation,
				baseSha: handoff.baseHead,
				headSha: handoff.head,
				changedPaths: handoff.changedScope,
				allowedScopes: child.writeScopes,
				...targetChanges,
			}),
			completedAt: "2026-07-21T12:00:00.000Z",
			verdict: "clean",
			findings: [],
		}];
		transport.pullRequests.push(pr);
		const result = await orchestratorFor(transport).integrateChild(candidate, child, handoff);
		assert.equal(result.kind, "blocked", JSON.stringify(targetChanges));
		assert.equal(transport.integrateCalls, 0);
	}
});

test("restart reuses an exact bound integration receipt after GitHub closes the merged child PR", async () => {
	const candidate = await plan();
	const issue = issueFrom({
		repository: candidate.repository,
		parentIssue: candidate.parentIssue,
		marker: candidate.children[0].markers.issue,
		title: candidate.children[0].title,
		body: candidate.children[0].issueBody,
	}, 811);
	const child = materializeChildRecord(candidate, "evidence", issue);
	const handoff = childHandoff(issue.number, child.branch, child.prBase);
	const request: CreatePullRequestRequest = {
		repository: candidate.repository,
		workItemId: child.id,
		generation: candidate.generation,
		marker: child.markers.pullRequest,
		title: child.title,
		body: `Refs #${issue.number}\nRefs #${candidate.parentIssue}\n\n${child.markers.pullRequest}`,
		draft: false,
		baseBranch: child.prBase,
		headBranch: child.branch,
		baseSha: handoff.baseHead,
		headSha: handoff.head,
		changedPaths: handoff.changedScope,
		allowedScopes: child.writeScopes,
	};
	const transport = new FakeTransport();
	transport.pullRequests.push(cleanPullRequest(request, { state: "merged" }));
	transport.integrations.push({
		childId: child.id,
		pullRequest: 812,
		generation: candidate.generation,
		marker: child.markers.pullRequest,
		baseSha: handoff.baseHead,
		headSha: handoff.head,
		parentBranch: candidate.parentBranch,
		integratedAt: "2026-07-21T12:01:00.000Z",
	});
	const result = await orchestratorFor(transport).integrateChild(candidate, child, handoff);
	assert.equal(result.kind, "integrated");
	if (result.kind === "integrated") assert.equal(result.reused, true);
	assert.equal(transport.integrateCalls, 0);
});

test("keeps the parent draft until all children and an exact-generation/head consumed human decision pass", async () => {
	const candidate = await plan();
	const transport = new FakeTransport();
	const broker = new FakeDecisionBroker();
	const parentHead = "e".repeat(40);
	const parentRequest: CreatePullRequestRequest = {
		repository: candidate.repository,
		workItemId: `parent-${candidate.parentIssue}`,
		generation: candidate.generation,
		marker: candidate.markers.parentPullRequest,
		title: candidate.title,
		body: `Closes #${candidate.parentIssue}\n\n${candidate.markers.parentPullRequest}`,
		draft: true,
		baseBranch: candidate.parentBaseBranch,
		headBranch: candidate.parentBranch,
		baseSha,
		headSha: parentHead,
		changedPaths: [".pi/extensions/shepherd/github-orchestrator.ts"],
		allowedScopes: [
			".pi/extensions/shepherd/github-evidence.ts",
			".pi/extensions/shepherd/github-orchestrator.ts",
		],
	};
	transport.pullRequests.push(cleanPullRequest(parentRequest, { draft: true }, 900));
	const receipts: ChildIntegrationReceipt[] = candidate.children.map((child, index) => ({
		childId: child.id,
		pullRequest: 812 + index,
		generation: candidate.generation,
		marker: child.markers.pullRequest,
		baseSha,
		headSha: String(index + 1).repeat(40),
		parentBranch: candidate.parentBranch,
		integratedAt: "2026-07-21T12:01:00.000Z",
	}));
	transport.integrations.push(...receipts);

	const orchestrator = orchestratorFor(transport, broker);
	const blocked = await orchestrator.reconcileParentReadiness(candidate, receipts.slice(0, 1), decisionPolicy);
	assert.equal(blocked.kind, "blocked");
	assert.equal(broker.requests.length, 0);
	assert.equal(transport.markReadyCalls, 0);

	const ready = await orchestrator.reconcileParentReadiness(candidate, receipts, decisionPolicy);
	assert.equal(ready.kind, "ready");
	assert.equal(transport.markReadyCalls, 1);
	assert.equal(broker.consumes, 1);
	assert.deepEqual(
		{
			gate: broker.requests[0].gate,
			parentIssue: broker.requests[0].parentIssue,
			pullRequest: broker.requests[0].pullRequest,
			generation: broker.requests[0].generation,
			headSha: broker.requests[0].headSha,
			allowedOptions: broker.requests[0].allowedOptions,
		},
		{
			gate: "parent_merge",
			parentIssue: candidate.parentIssue,
			pullRequest: 900,
			generation: candidate.generation,
			headSha: parentHead,
			allowedOptions: ["approve-merge", "reject"],
		},
	);
	assert.equal("mergeParent" in transport, false);
});

test("pending/rejected decisions and parent head movement never mark ready", async () => {
	const candidate = await plan();
	const receipts: ChildIntegrationReceipt[] = candidate.children.map((child, index) => ({
		childId: child.id,
		pullRequest: 812 + index,
		generation: candidate.generation,
		marker: child.markers.pullRequest,
		baseSha,
		headSha: String(index + 1).repeat(40),
		parentBranch: candidate.parentBranch,
		integratedAt: "2026-07-21T12:01:00.000Z",
	}));
	const makeTransport = (): FakeTransport => {
		const transport = new FakeTransport();
		transport.pullRequests.push(cleanPullRequest({
			repository: candidate.repository,
			workItemId: `parent-${candidate.parentIssue}`,
			generation: candidate.generation,
			marker: candidate.markers.parentPullRequest,
			title: candidate.title,
			body: `Closes #${candidate.parentIssue}\n\n${candidate.markers.parentPullRequest}`,
			draft: true,
			baseBranch: candidate.parentBaseBranch,
			headBranch: candidate.parentBranch,
			baseSha,
			headSha: "e".repeat(40),
			changedPaths: [".pi/extensions/shepherd/github-orchestrator.ts"],
			allowedScopes: [
				".pi/extensions/shepherd/github-evidence.ts",
				".pi/extensions/shepherd/github-orchestrator.ts",
			],
		}, { draft: true }, 900));
		transport.integrations.push(...receipts);
		return transport;
	};

	for (const pollResult of [
		{ status: "pending", attempts: 1 } as const,
		{ status: "decided", decision: { ...approvedDecision, option: "reject" }, attempts: 1 } as const,
	]) {
		const transport = makeTransport();
		const broker = new FakeDecisionBroker();
		broker.pollResult = pollResult;
		const result = await orchestratorFor(transport, broker).reconcileParentReadiness(candidate, receipts, decisionPolicy);
		assert.notEqual(result.kind, "ready");
		assert.equal(transport.markReadyCalls, 0);
	}

	const movedTransport = makeTransport();
	movedTransport.onPullRequestRead = (_query, matches, read) => read < 2
		? matches
		: matches.map((candidatePr) => ({ ...candidatePr, headSha: "f".repeat(40) }));
	const moved = await orchestratorFor(movedTransport, new FakeDecisionBroker()).reconcileParentReadiness(candidate, receipts, decisionPolicy);
	assert.equal(moved.kind, "blocked");
	assert.equal(movedTransport.markReadyCalls, 0);
});

test("parent readiness reconciles caller receipts against complete authoritative records and current ancestry", async () => {
	const candidate = await plan();
	const parentHead = "e".repeat(40);
	const receipts: ChildIntegrationReceipt[] = candidate.children.map((child, index) => ({
		childId: child.id,
		pullRequest: 812 + index,
		generation: candidate.generation,
		marker: child.markers.pullRequest,
		baseSha,
		headSha: String(index + 1).repeat(40),
		parentBranch: candidate.parentBranch,
		integratedAt: "2026-07-21T12:01:00.000Z",
	}));
	const setup = (): FakeTransport => {
		const transport = new FakeTransport();
		transport.integrations.push(...receipts);
		transport.pullRequests.push(cleanPullRequest({
			repository: candidate.repository,
			workItemId: `parent-${candidate.parentIssue}`,
			generation: candidate.generation,
			marker: candidate.markers.parentPullRequest,
			title: candidate.title,
			body: `Closes #${candidate.parentIssue}\n\n${candidate.markers.parentPullRequest}`,
			draft: true,
			baseBranch: candidate.parentBaseBranch,
			headBranch: candidate.parentBranch,
			baseSha,
			headSha: parentHead,
			changedPaths: [".pi/extensions/shepherd/github-orchestrator.ts"],
			allowedScopes: candidate.children.flatMap((child) => child.writeScopes),
		}, { draft: true }, 900));
		return transport;
	};

	const forgedTransport = setup();
	const forged = receipts.map((receipt, index) => index === 0 ? { ...receipt, pullRequest: 9_999 } : receipt);
	const forgedResult = await orchestratorFor(forgedTransport, new FakeDecisionBroker())
		.reconcileParentReadiness(candidate, forged, decisionPolicy);
	assert.equal(forgedResult.kind, "blocked");
	assert.equal(forgedTransport.markReadyCalls, 0);

	const forcePushed = setup();
	forcePushed.ancestry = false;
	const stale = await orchestratorFor(forcePushed, new FakeDecisionBroker())
		.reconcileParentReadiness(candidate, receipts, decisionPolicy);
	assert.equal(stale.kind, "blocked");
	assert.equal(forcePushed.markReadyCalls, 0);

	const incomplete = setup();
	incomplete.incompleteIntegrationLookup = true;
	const incompleteResult = await orchestratorFor(incomplete, new FakeDecisionBroker())
		.reconcileParentReadiness(candidate, receipts, decisionPolicy);
	assert.equal(incompleteResult.kind, "blocked");
	assert.equal(incomplete.markReadyCalls, 0);
});

test("parent ready transition recovers timeout and malformed responses only after exact authoritative re-read", async () => {
	const candidate = await plan();
	const receipts: ChildIntegrationReceipt[] = candidate.children.map((child, index) => ({
		childId: child.id,
		pullRequest: 812 + index,
		generation: candidate.generation,
		marker: child.markers.pullRequest,
		baseSha,
		headSha: String(index + 1).repeat(40),
		parentBranch: candidate.parentBranch,
		integratedAt: "2026-07-21T12:01:00.000Z",
	}));
	for (const mode of ["timeout", "malformed"] as const) {
		const transport = new FakeTransport();
		transport.integrations.push(...receipts);
		transport.pullRequests.push(cleanPullRequest({
			repository: candidate.repository,
			workItemId: `parent-${candidate.parentIssue}`,
			generation: candidate.generation,
			marker: candidate.markers.parentPullRequest,
			title: candidate.title,
			body: `Closes #${candidate.parentIssue}\n\n${candidate.markers.parentPullRequest}`,
			draft: true,
			baseBranch: candidate.parentBaseBranch,
			headBranch: candidate.parentBranch,
			baseSha,
			headSha: "e".repeat(40),
			changedPaths: [".pi/extensions/shepherd/github-orchestrator.ts"],
			allowedScopes: candidate.children.flatMap((child) => child.writeScopes),
		}, { draft: true }, 900));
		if (mode === "timeout") transport.throwAfterReady = true;
		else transport.malformedReadyResponse = true;
		const result = await orchestratorFor(transport, new FakeDecisionBroker())
			.reconcileParentReadiness(candidate, receipts, decisionPolicy);
		assert.equal(result.kind, "ready", mode);
		if (result.kind === "ready") assert.equal(result.reused, true, mode);
		assert.equal(transport.markReadyCalls, 1, mode);
	}
});
