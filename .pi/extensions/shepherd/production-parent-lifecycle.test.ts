import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import test from "node:test";

import type { ProductionParentPlanDocument } from "./autonomous-production-contract.ts";
import {
	createProductionAutonomousState,
	validateProductionAutonomousState,
	type ProductionAutonomousState,
} from "./autonomous-production-state.ts";
import {
	createRequiredGitHubCheckPolicy,
	type GitHubChangedPathEvidence,
	type GitHubPullRequestEvidence,
	type RequiredGitHubCheckPolicy,
} from "./github-evidence.ts";
import {
	createCanonicalPullRequestSnapshot,
	validateChildIntegrationReceipt,
	type AuthoritativeLookup,
	type ChildIntegrationQuery,
	type ChildIntegrationReceipt,
	type ExternalCallContext,
	type GitAncestryProof,
	type GitAncestryQuery,
	type ParentDecisionBroker,
	type PullRequestMarkerQuery,
} from "./github-orchestrator.ts";
import { createHumanDecisionRecord, type HumanDecisionRecord } from "./human-decision.ts";
import { createProductionOrchestrationPlan } from "./production-orchestration-plan.ts";
import type { ProductionReviewArtifact } from "./production-review-adapter.ts";
import {
	createAgentSessionAttestation,
	createIndependentReviewWork,
	validateIndependentReviewRecord,
	type IndependentReviewRecord,
	type IndependentReviewTarget,
} from "./review-router.ts";
import type { GitHubDecisionRequest } from "./github-decision-broker.ts";
import type {
	AuthoritativeParentMergeState,
	ParentPullRequestMergeLookup,
} from "./production-human-gate.ts";
import {
	ProductionParentFinalizer,
	ProductionParentGateAdapter,
	type ProductionParentCheckPolicyAuthority,
	type ProductionParentFinalizationTransport,
	type ProductionParentReadyTransitionPort,
	type ProductionParentReadyTransitionReceipt,
	type ProductionParentReviewAuthority,
} from "./production-parent-lifecycle.ts";

const PARENT_BASE = "1".repeat(40);
const CHILD_HEAD = "2".repeat(40);
const PARENT_HEAD = "3".repeat(40);
const MOVED_HEAD = "4".repeat(40);
const MERGE_HEAD = "5".repeat(40);
const OBSERVED = "2026-07-20T00:00:05.000Z";
const REVIEWED = "2026-07-20T00:00:06.000Z";
const REOBSERVED = "2026-07-20T00:00:07.000Z";

function deferred<T>() {
	let resolve!: (value: T | PromiseLike<T>) => void;
	let reject!: (reason?: unknown) => void;
	const promise = new Promise<T>((accept, decline) => { resolve = accept; reject = decline; });
	return { promise, resolve, reject };
}

function plan(): ProductionParentPlanDocument {
	return {
		schemaVersion: 2,
		planId: "plan-479-parent-lifecycle",
		parentIssue: 479,
		repository: "acme/widgets",
		title: "Production parent lifecycle",
		objective: "Finalize the exact parent and wait for a human-owned merge.",
		parentBranch: "feat/471-shepherd",
		parentBaseBranch: "main",
		actorAllowlist: ["maintainer"],
		decisionExpiresAt: "2026-08-01T00:00:00.000Z",
		children: [{
			id: "lane",
			issue: 501,
			title: "Production lane",
			task: "Implement the production lane.",
			slug: "production-lane",
			dependsOn: [],
			access: "mutating",
			writeScopes: ["owned/lane"],
			requiredSkills: ["javascript-testing-patterns"],
			verification: [{
				id: "tests",
				executable: "node",
				args: ["--test", "owned/lane.test.ts"],
				cwd: ".",
				timeoutMs: 30_000,
				maxOutputBytes: 1_000_000,
			}],
			humanGates: ["review"],
			maxAttempts: 2,
			maxCorrections: 1,
		}],
	};
}

function policies(value = plan()): RequiredGitHubCheckPolicy[] {
	return [
		createRequiredGitHubCheckPolicy({
			schemaVersion: 1,
			repository: value.repository,
			baseBranch: value.parentBranch,
			revision: 11,
			requiredChecks: [{ name: "child-ci", producerId: "github-actions" }],
		}),
		createRequiredGitHubCheckPolicy({
			schemaVersion: 1,
			repository: value.repository,
			baseBranch: value.parentBaseBranch,
			revision: 12,
			requiredChecks: [{ name: "parent-ci", producerId: "github-actions" }],
		}),
	];
}

function pullRequest(input: {
	workItemId: string;
	number: number;
	marker: string;
	baseBranch: string;
	headBranch: string;
	baseSha: string;
	headSha: string;
	policy: RequiredGitHubCheckPolicy;
	allowedScopes: string[];
	changedPaths: string[];
	draft?: boolean;
	state?: "open" | "closed" | "merged";
	reviews?: IndependentReviewRecord[];
	revision?: number;
	observedAt?: string;
}): GitHubPullRequestEvidence {
	const observedAt = input.observedAt ?? OBSERVED;
	return {
		schemaVersion: 2,
		repository: "acme/widgets",
		workItemId: input.workItemId,
		generation: 1,
		number: input.number,
		marker: input.marker,
		title: `${input.workItemId} pull request`,
		body: `${input.marker}\nBounded production pull request.`,
		state: input.state ?? "open",
		draft: input.draft ?? false,
		baseBranch: input.baseBranch,
		headBranch: input.headBranch,
		baseSha: input.baseSha,
		headSha: input.headSha,
		changedPathsComplete: true,
		changedPaths: [...input.changedPaths],
		allowedScopes: [...input.allowedScopes],
		mergeState: "clean",
		policyDigest: input.policy.digest,
		checksComplete: true,
		checks: [{
			id: `check-${input.number}`,
			name: input.policy.requiredChecks[0].name,
			producerId: input.policy.requiredChecks[0].producerId,
			sequence: 1,
			status: "completed",
			conclusion: "success",
			headSha: input.headSha,
			updatedAt: observedAt,
			completedAt: observedAt,
		}],
		requestedChangesComplete: true,
		requestedChanges: [],
		threadsComplete: true,
		threads: [],
		reviews: input.reviews ?? [],
		reviewsComplete: true,
		dispositionsComplete: true,
		dispositions: [],
		revision: input.revision ?? 10,
		observedAt,
	};
}

function receipt(value = plan(), policyValues = policies(value)): ChildIntegrationReceipt {
	const orchestration = createProductionOrchestrationPlan(value, 1, policyValues);
	const child = orchestration.children[0];
	const childPullRequest = pullRequest({
		workItemId: child.id,
		number: 501,
		marker: child.markers.pullRequest,
		baseBranch: value.parentBranch,
		headBranch: "feat/501-production-lane",
		baseSha: PARENT_BASE,
		headSha: CHILD_HEAD,
		policy: policyValues.find((policy) => policy.baseBranch === value.parentBranch)!,
		allowedScopes: [...child.writeScopes],
		changedPaths: ["owned/lane/file.ts"],
	});
	return validateChildIntegrationReceipt({
		childId: child.id,
		pullRequest: childPullRequest.number,
		generation: 1,
		marker: child.markers.pullRequest,
		baseSha: PARENT_BASE,
		headSha: CHILD_HEAD,
		parentBranch: value.parentBranch,
		pullRequestSnapshot: createCanonicalPullRequestSnapshot(childPullRequest),
		observation: {
			revision: childPullRequest.revision,
			observedAt: childPullRequest.observedAt,
			state: "open",
		},
		controllerProvenance: {
			authority: "controller",
			planDigest: orchestration.canonical.digest,
			policyDigest: childPullRequest.policyDigest,
			policyRevision: 11,
			policyObservedAt: OBSERVED,
			changedPathDigest: "6".repeat(64),
			reviewAuthorizationDigest: "7".repeat(64),
			reviewResultDigest: "8".repeat(64),
			reviewCompletedAt: OBSERVED,
			evidenceRevision: childPullRequest.revision,
			observedAt: OBSERVED,
		},
		transportProvenance: {
			authority: "transport",
			idempotencyKey: "integrate-lane-1",
			intentDigest: "9".repeat(64),
			revision: 13,
		},
		integratedAt: "2026-07-20T00:00:05.001Z",
	});
}

function receiptDigest(value: ChildIntegrationReceipt): string {
	return createHash("sha256").update(JSON.stringify(validateChildIntegrationReceipt(value))).digest("hex");
}

function succeededState(value = plan(), exactReceipt = receipt(value)): ProductionAutonomousState {
	const state = createProductionAutonomousState(value, {
		runId: "run-parent-lifecycle",
		now: new Date("2026-07-20T00:00:00.000Z"),
		timeoutMs: 30_000,
	});
	state.stage = "child_lifecycle";
	state.children[0].status = "succeeded";
	state.children[0].stage = "succeeded";
	state.children[0].checkpoint = {
		summary: "Child integrated authoritatively.",
		pullRequest: exactReceipt.pullRequest,
		integrationReceiptDigest: receiptDigest(exactReceipt),
		parentHead: PARENT_HEAD,
	};
	return validateProductionAutonomousState(state);
}

class PolicyAuthority implements ProductionParentCheckPolicyAuthority {
	items: RequiredGitHubCheckPolicy[];
	complete = true;
	block = false;

	constructor(items = policies()) { this.items = items; }

	async findRequiredCheckPolicies(
		_query: Parameters<ProductionParentCheckPolicyAuthority["findRequiredCheckPolicies"]>[0],
		context: ExternalCallContext,
	): Promise<AuthoritativeLookup<RequiredGitHubCheckPolicy>> {
		if (this.block) {
			return new Promise((_, reject) => context.signal.addEventListener("abort", () => reject(new Error("policy call aborted")), { once: true }));
		}
		return { items: structuredClone(this.items), complete: this.complete };
	}
}

class FinalizationTransport implements ProductionParentFinalizationTransport {
	parent: GitHubPullRequestEvidence;
	exactReceipt: ChildIntegrationReceipt;
	pullLookupComplete = true;
	receiptLookupComplete = true;
	ancestry = true;
	pullLookups = 0;

	constructor(value = plan(), policyValues = policies(value), exactReceipt = receipt(value, policyValues)) {
		const orchestration = createProductionOrchestrationPlan(value, 1, policyValues);
		this.parent = pullRequest({
			workItemId: `parent-${value.parentIssue}`,
			number: 472,
			marker: orchestration.markers.parentPullRequest,
			baseBranch: value.parentBaseBranch,
			headBranch: value.parentBranch,
			baseSha: PARENT_BASE,
			headSha: PARENT_HEAD,
			policy: policyValues.find((policy) => policy.baseBranch === value.parentBaseBranch)!,
			allowedScopes: ["owned/lane"],
			changedPaths: ["owned/lane/file.ts"],
		});
		this.exactReceipt = exactReceipt;
	}

	async findPullRequests(_query: PullRequestMarkerQuery, _context: ExternalCallContext) {
		this.pullLookups += 1;
		return { items: [structuredClone(this.parent)], complete: this.pullLookupComplete };
	}

	async findChildIntegration(_query: ChildIntegrationQuery, _context: ExternalCallContext) {
		return { items: [structuredClone(this.exactReceipt)], complete: this.receiptLookupComplete };
	}

	async proveAncestry(query: GitAncestryQuery, _context: ExternalCallContext): Promise<GitAncestryProof> {
		return {
			schemaVersion: 1,
			authority: "transport",
			repository: query.repository,
			ancestorSha: query.ancestorSha,
			descendantSha: query.descendantSha,
			result: this.ancestry,
			revision: 14,
			observedAt: REOBSERVED,
		};
	}
}

class ReviewAuthority implements ProductionParentReviewAuthority {
	readonly transport: FinalizationTransport;
	verdict: "clean" | "findings" = "clean";
	moveHead = false;
	lastTarget: IndependentReviewTarget | undefined;

	constructor(transport: FinalizationTransport) { this.transport = transport; }

	async findChangedPathEvidence(
		query: Omit<IndependentReviewTarget, "changedPaths" | "allowedScopes">,
		_context: ExternalCallContext,
	): Promise<AuthoritativeLookup<GitHubChangedPathEvidence>> {
		return {
			complete: true,
			items: [{
				schemaVersion: 1,
				authority: "controller",
				repository: query.repository,
				workItemId: query.workItemId,
				pullRequest: query.pullRequest,
				generation: query.generation,
				baseSha: query.baseSha,
				headSha: query.headSha,
				paths: ["owned/lane/file.ts"],
				complete: true,
				revision: this.transport.parent.revision,
				observedAt: this.transport.parent.observedAt,
			}],
		};
	}

	async review(target: IndependentReviewTarget, _context: ExternalCallContext): Promise<ProductionReviewArtifact> {
		this.lastTarget = structuredClone(target);
		const work = createIndependentReviewWork(target);
		const review = validateIndependentReviewRecord({
			...work,
			completedAt: REVIEWED,
			verdict: this.verdict,
			findings: this.verdict === "clean" ? [] : [{ id: "parent-finding", severity: "blocking", summary: "Parent is unsafe." }],
		});
		const attestation = createAgentSessionAttestation({ sessionId: "review-session", runId: "review-run", review });
		if (this.moveHead) {
			this.transport.parent.headSha = MOVED_HEAD;
			this.transport.parent.checks[0].headSha = MOVED_HEAD;
		} else {
			this.transport.parent.reviews = [review];
		}
		this.transport.parent.revision += 1;
		this.transport.parent.observedAt = REOBSERVED;
		this.transport.parent.checks[0].updatedAt = REOBSERVED;
		this.transport.parent.checks[0].completedAt = REOBSERVED;
		return {
			schemaVersion: 1,
			review,
			attestation,
			dispositions: [],
			revision: 77,
			publishedAt: REVIEWED,
		};
	}
}

function finalizationHarness() {
	const value = plan();
	const policyValues = policies(value);
	const exactReceipt = receipt(value, policyValues);
	const transport = new FinalizationTransport(value, policyValues, exactReceipt);
	const policyAuthority = new PolicyAuthority(policyValues);
	const reviews = new ReviewAuthority(transport);
	const state = succeededState(value, exactReceipt);
	return { value, policyValues, exactReceipt, transport, policyAuthority, reviews, state };
}

test("finalizes only the stable exact parent head with authoritative receipts, trusted green CI, and clean xhigh read-only review", async () => {
	const harness = finalizationHarness();
	const finalizer = new ProductionParentFinalizer({
		transport: harness.transport,
		policies: harness.policyAuthority,
		reviews: harness.reviews,
	});
	const result = await finalizer.finalize(harness.value, harness.state, new AbortController().signal);
	assert.deepEqual(result, {
		pullRequest: 472,
		head: PARENT_HEAD,
		summary: `Parent PR #472 at ${PARENT_HEAD} finalized with 1 exact child receipt, trusted green CI, and a clean independent xhigh read-only review.`,
	});
	assert.equal(harness.transport.pullLookups >= 2, true);
	assert.equal(harness.reviews.lastTarget?.headSha, PARENT_HEAD);
	assert.equal(harness.reviews.lastTarget?.baseBranch, "main");
});

test("parent finalization fails closed for incomplete child, forged receipt, incomplete or untrusted CI, changed head, and non-clean review", async (t) => {
	await t.test("incomplete child", async () => {
		const harness = finalizationHarness();
		harness.state.children[0].status = "blocked";
		harness.state.children[0].stage = "integration";
		await assert.rejects(new ProductionParentFinalizer({
			transport: harness.transport, policies: harness.policyAuthority, reviews: harness.reviews,
		}).finalize(harness.value, harness.state, new AbortController().signal), /every child|succeeded|receipt/i);
	});

	await t.test("forged receipt digest", async () => {
		const harness = finalizationHarness();
		harness.state.children[0].checkpoint!.integrationReceiptDigest = "f".repeat(64);
		await assert.rejects(new ProductionParentFinalizer({
			transport: harness.transport, policies: harness.policyAuthority, reviews: harness.reviews,
		}).finalize(harness.value, harness.state, new AbortController().signal), /receipt.*digest|exact receipt/i);
	});

	await t.test("incomplete CI evidence", async () => {
		const harness = finalizationHarness();
		harness.transport.parent.checksComplete = false;
		await assert.rejects(new ProductionParentFinalizer({
			transport: harness.transport, policies: harness.policyAuthority, reviews: harness.reviews,
		}).finalize(harness.value, harness.state, new AbortController().signal), /evidence_incomplete|ci_not_green|complete/i);
	});

	await t.test("untrusted check producer", async () => {
		const harness = finalizationHarness();
		harness.transport.parent.checks[0].producerId = "untrusted-app";
		await assert.rejects(new ProductionParentFinalizer({
			transport: harness.transport, policies: harness.policyAuthority, reviews: harness.reviews,
		}).finalize(harness.value, harness.state, new AbortController().signal), /ci_not_green|trusted.*ci/i);
	});

	await t.test("parent head changed during review", async () => {
		const harness = finalizationHarness();
		harness.reviews.moveHead = true;
		await assert.rejects(new ProductionParentFinalizer({
			transport: harness.transport, policies: harness.policyAuthority, reviews: harness.reviews,
		}).finalize(harness.value, harness.state, new AbortController().signal), /head.*moved|changed.*head/i);
	});

	await t.test("review returned findings", async () => {
		const harness = finalizationHarness();
		harness.reviews.verdict = "findings";
		await assert.rejects(new ProductionParentFinalizer({
			transport: harness.transport, policies: harness.policyAuthority, reviews: harness.reviews,
		}).finalize(harness.value, harness.state, new AbortController().signal), /clean.*review|review.*clean/i);
	});

	await t.test("authoritative receipt lookup is incomplete", async () => {
		const harness = finalizationHarness();
		harness.transport.receiptLookupComplete = false;
		await assert.rejects(new ProductionParentFinalizer({
			transport: harness.transport, policies: harness.policyAuthority, reviews: harness.reviews,
		}).finalize(harness.value, harness.state, new AbortController().signal), /receipt.*incomplete|authoritative/i);
	});

	await t.test("receipt is not in the exact parent ancestry", async () => {
		const harness = finalizationHarness();
		harness.transport.ancestry = false;
		await assert.rejects(new ProductionParentFinalizer({
			transport: harness.transport, policies: harness.policyAuthority, reviews: harness.reviews,
		}).finalize(harness.value, harness.state, new AbortController().signal), /ancestry|integrated/i);
	});
});

class ReadyTransition implements ProductionParentReadyTransitionPort {
	readonly transport: FinalizationTransport;
	calls: Parameters<ProductionParentReadyTransitionPort["markExistingDraftReady"]>[0][] = [];

	constructor(transport: FinalizationTransport) { this.transport = transport; }

	async markExistingDraftReady(
		request: Parameters<ProductionParentReadyTransitionPort["markExistingDraftReady"]>[0],
		_context: ExternalCallContext,
	): Promise<ProductionParentReadyTransitionReceipt> {
		this.calls.push(structuredClone(request));
		this.transport.parent.draft = false;
		this.transport.parent.revision += 1;
		return {
			...request,
			schemaVersion: 1,
			authority: "transport",
			operation: "existing_draft_to_ready",
			appliedRevision: this.transport.parent.revision,
			observedAt: REOBSERVED,
		};
	}
}

test("a parent draft requires the explicit bounded existing-draft-to-ready port and is authoritatively re-observed", async () => {
	const blocked = finalizationHarness();
	blocked.transport.parent.draft = true;
	await assert.rejects(new ProductionParentFinalizer({
		transport: blocked.transport, policies: blocked.policyAuthority, reviews: blocked.reviews,
	}).finalize(blocked.value, blocked.state, new AbortController().signal), /draft.*transition|ready.*port/i);

	const harness = finalizationHarness();
	harness.transport.parent.draft = true;
	const readiness = new ReadyTransition(harness.transport);
	const result = await new ProductionParentFinalizer({
		transport: harness.transport,
		policies: harness.policyAuthority,
		reviews: harness.reviews,
		readiness,
	}).finalize(harness.value, harness.state, new AbortController().signal);
	assert.equal(result.head, PARENT_HEAD);
	assert.deepEqual(readiness.calls, [{
		repository: "acme/widgets",
		parentIssue: 479,
		pullRequest: 472,
		generation: 1,
		branch: "feat/471-shepherd",
		headSha: PARENT_HEAD,
		expectedRevision: 11,
	}]);
	assert.equal(harness.transport.parent.draft, false);
});

function decisionRecord(request: GitHubDecisionRequest, status: "pending" | "decided" | "consumed", option?: "approve-merge" | "reject") {
	const value = createHumanDecisionRecord({
		requestId: request.requestId,
		gate: "parent_merge",
		binding: {
			repository: request.repository,
			target: { kind: "pull_request", number: request.pullRequest },
			generation: request.generation,
			headSha: request.headSha,
		},
		allowedOptions: ["approve-merge", "reject"],
		actorAllowlist: [...request.actorAllowlist],
		expiresAt: request.expiresAt,
		question: request.question,
	}, new Date("2026-07-20T00:00:00.000Z"));
	value.requestComment = {
		id: 1,
		url: `https://github.com/acme/widgets/pull/${request.pullRequest}#issuecomment-1`,
		actor: "shepherd-host",
		createdAt: "2026-07-20T00:00:00.000Z",
	};
	value.status = status;
	if (option !== undefined) {
		value.decision = {
			option,
			actor: "maintainer",
			sourceUrl: `https://github.com/acme/widgets/pull/${request.pullRequest}#issuecomment-2`,
			decidedAt: "2026-07-20T00:00:01.000Z",
		};
		value.updatedAt = "2026-07-20T00:00:01.000Z";
	}
	if (status === "consumed") {
		value.consumedAt = "2026-07-20T00:00:02.000Z";
		value.updatedAt = value.consumedAt;
	}
	return value;
}

class GateBroker implements ParentDecisionBroker {
	requestValue: GitHubDecisionRequest | undefined;
	record: HumanDecisionRecord | undefined;
	blockRequest = false;
	blockPoll = false;

	async request(value: GitHubDecisionRequest, context: ExternalCallContext): Promise<HumanDecisionRecord> {
		this.requestValue = structuredClone(value);
		if (this.blockRequest) {
			return new Promise((_, reject) => context.signal.addEventListener("abort", () => reject(new Error("broker request aborted")), { once: true }));
		}
		this.record = decisionRecord(value, "pending");
		return this.record;
	}

	async poll(
		_requestId: string,
		_binding: Parameters<ParentDecisionBroker["poll"]>[1],
		_options: Parameters<ParentDecisionBroker["poll"]>[2],
		context: ExternalCallContext,
	): Promise<HumanDecisionRecord> {
		if (this.blockPoll) {
			return new Promise((_, reject) => context.signal.addEventListener("abort", () => reject(new Error("broker poll aborted")), { once: true }));
		}
		if (!this.record) throw new Error("decision was not requested");
		return this.record;
	}

	async consume(): Promise<HumanDecisionRecord> {
		if (!this.requestValue || !this.record?.decision) throw new Error("decision is not consumable");
		this.record = decisionRecord(this.requestValue, "consumed", this.record.decision.option as "approve-merge" | "reject");
		return this.record;
	}
}

class MergeLookup implements ParentPullRequestMergeLookup {
	state: AuthoritativeParentMergeState = {
		repository: "acme/widgets",
		pullRequest: 472,
		headSha: PARENT_HEAD,
		state: "open",
		mergedAt: null,
		mergeCommitSha: null,
		revision: 20,
		observedAt: REOBSERVED,
	};

	async observeExactPullRequest(
		_query: Parameters<ParentPullRequestMergeLookup["observeExactPullRequest"]>[0],
		_context: ExternalCallContext,
	): Promise<AuthoritativeParentMergeState> {
		return structuredClone(this.state);
	}
}

function waitingState(state: ProductionAutonomousState, requestId: string): ProductionAutonomousState {
	const value = structuredClone(state);
	value.status = "waiting_human";
	value.stage = "human_decision";
	value.humanGate = {
		repository: "acme/widgets",
		pullRequest: 472,
		generation: 1,
		head: PARENT_HEAD,
		requestId,
		status: "pending",
	};
	return validateProductionAutonomousState(value);
}

function preparedState(state: ProductionAutonomousState, requestId: string): ProductionAutonomousState {
	const value = structuredClone(state);
	value.status = "running";
	value.stage = "human_decision";
	value.humanGate = {
		repository: "acme/widgets",
		pullRequest: 472,
		generation: 1,
		head: PARENT_HEAD,
		requestId,
		status: "prepared",
	};
	return validateProductionAutonomousState(value);
}

test("the controller gate binds the exact durable request and approval only waits until an authoritative human merge", async () => {
	const harness = finalizationHarness();
	const broker = new GateBroker();
	const lookup = new MergeLookup();
	const gate = new ProductionParentGateAdapter(broker, lookup);
	const finalization = {
		pullRequest: 472,
		head: PARENT_HEAD,
		summary: "Exact finalization.",
	};
	const prepared = gate.prepare(harness.value, harness.state, finalization);
	const requested = await gate.request(harness.value, preparedState(harness.state, prepared.requestId), finalization, new AbortController().signal);
	assert.match(requested.requestId, /^parent-merge-479-1-[0-9a-f]{24}$/u);
	assert.deepEqual(broker.requestValue, {
		requestId: requested.requestId,
		gate: "parent_merge",
		repository: "acme/widgets",
		parentIssue: 479,
		pullRequest: 472,
		generation: 1,
		headSha: PARENT_HEAD,
		allowedOptions: ["approve-merge", "reject"],
		actorAllowlist: ["maintainer"],
		expiresAt: "2026-08-01T00:00:00.000Z",
		question: `Approve the human-owned merge of parent PR #472 at exact head ${PARENT_HEAD}? Shepherd cannot merge the default branch.`,
	});

	const state = waitingState(harness.state, requested.requestId);
	broker.record = decisionRecord(broker.requestValue!, "decided", "approve-merge");
	assert.deepEqual(await gate.observe(harness.value, state, new AbortController().signal), { status: "approved_waiting_for_merge" });
	lookup.state = {
		...lookup.state,
		state: "merged",
		mergedAt: "2026-07-20T00:00:08.000Z",
		mergeCommitSha: MERGE_HEAD,
		revision: 21,
		observedAt: "2026-07-20T00:00:08.000Z",
	};
	assert.deepEqual(await gate.observe(harness.value, state, new AbortController().signal), {
		status: "merged",
		repository: "acme/widgets",
		pullRequest: 472,
		head: PARENT_HEAD,
		mergedAt: "2026-07-20T00:00:08.000Z",
		mergeCommitSha: MERGE_HEAD,
		revision: 21,
		observedAt: "2026-07-20T00:00:08.000Z",
	});
	assert.equal("merge" in gate, false);
	assert.equal("mergeParent" in gate, false);
});

test("the controller gate returns a typed exact-head invalidation before polling stale approval", async () => {
	const harness = finalizationHarness();
	const broker = new GateBroker();
	const lookup = new MergeLookup();
	const gate = new ProductionParentGateAdapter(broker, lookup);
	const finalization = {
		pullRequest: 472, head: PARENT_HEAD, summary: "Exact finalization.",
	};
	const prepared = gate.prepare(harness.value, harness.state, finalization);
	const request = await gate.request(harness.value, preparedState(harness.state, prepared.requestId), finalization, new AbortController().signal);
	lookup.state.headSha = MOVED_HEAD;
	assert.deepEqual(
		await gate.observe(harness.value, waitingState(harness.state, request.requestId), new AbortController().signal),
		{
			status: "invalidated",
			repository: "acme/widgets",
			pullRequest: 472,
			previousHead: PARENT_HEAD,
			currentHead: MOVED_HEAD,
			revision: 20,
			observedAt: REOBSERVED,
		},
	);
});

test("request/poll calls are bounded, caller cancellation propagates, and close aborts and joins active work", async () => {
	const harness = finalizationHarness();
	const lookup = new MergeLookup();
	const timedBroker = new GateBroker();
	timedBroker.blockRequest = true;
	const timed = new ProductionParentGateAdapter(timedBroker, lookup, { requestTimeoutMs: 10, closeTimeoutMs: 100 });
	const finalization = {
		pullRequest: 472, head: PARENT_HEAD, summary: "Exact finalization.",
	};
	const timedPrepared = timed.prepare(harness.value, harness.state, finalization);
	await assert.rejects(timed.request(harness.value, preparedState(harness.state, timedPrepared.requestId), finalization, new AbortController().signal), /request.*timed out|timed out.*request/i);
	await timed.close();

	const cancelled = new ProductionParentGateAdapter(new GateBroker(), lookup);
	const caller = new AbortController();
	caller.abort(new Error("caller cancelled"));
	const cancelledPrepared = cancelled.prepare(harness.value, harness.state, finalization);
	await assert.rejects(cancelled.request(harness.value, preparedState(harness.state, cancelledPrepared.requestId), finalization, caller.signal), /cancelled/i);

	const pollBroker = new GateBroker();
	const polling = new ProductionParentGateAdapter(pollBroker, lookup, { pollTimeoutMs: 10, closeTimeoutMs: 100 });
	const pollingPrepared = polling.prepare(harness.value, harness.state, finalization);
	const requested = await polling.request(harness.value, preparedState(harness.state, pollingPrepared.requestId), finalization, new AbortController().signal);
	pollBroker.blockPoll = true;
	await assert.rejects(polling.observe(
		harness.value,
		waitingState(harness.state, requested.requestId),
		new AbortController().signal,
	), /poll.*timed out|timed out.*poll/i);
	await polling.close();

	const closingBroker = new GateBroker();
	closingBroker.blockRequest = true;
	const closing = new ProductionParentGateAdapter(closingBroker, lookup, { requestTimeoutMs: 10_000, closeTimeoutMs: 100 });
	const closingPrepared = closing.prepare(harness.value, harness.state, finalization);
	const active = closing.request(harness.value, preparedState(harness.state, closingPrepared.requestId), finalization, new AbortController().signal);
	const activeAssertion = assert.rejects(active, /closed|cancelled/i);
	await new Promise((resolve) => setImmediate(resolve));
	await closing.close();
	await activeAssertion;
	await assert.rejects(closing.request(harness.value, preparedState(harness.state, closingPrepared.requestId), finalization, new AbortController().signal), /closed/i);
});

test("finalizer cancellation and close fence authority work", async () => {
	const harness = finalizationHarness();
	const blockingPolicies = new PolicyAuthority(harness.policyValues);
	blockingPolicies.block = true;
	const finalizer = new ProductionParentFinalizer({
		transport: harness.transport,
		policies: blockingPolicies,
		reviews: harness.reviews,
		timeoutMs: 10_000,
		closeTimeoutMs: 100,
	});
	const active = finalizer.finalize(harness.value, harness.state, new AbortController().signal);
	const activeAssertion = assert.rejects(active, /closed|cancelled/i);
	await new Promise((resolve) => setImmediate(resolve));
	await finalizer.close();
	await activeAssertion;
	await assert.rejects(finalizer.finalize(harness.value, harness.state, new AbortController().signal), /closed/i);
});

test("caller cancellation does not settle until the accepted parent operation acknowledges and joins", async () => {
	const harness = finalizationHarness();
	const release = deferred<void>();
	let underlyingSettled = false;
	const policies: ProductionParentCheckPolicyAuthority = {
		async findRequiredCheckPolicies(_query, context) {
			return new Promise((resolve, reject) => context.signal.addEventListener("abort", () => {
				void release.promise.then(() => {
					underlyingSettled = true;
					reject(new Error("accepted policy operation joined after abort"));
				});
			}, { once: true }));
		},
	};
	const finalizer = new ProductionParentFinalizer({
		transport: harness.transport,
		policies,
		reviews: harness.reviews,
		timeoutMs: 10_000,
	});
	const caller = new AbortController();
	let settled = false;
	const active = finalizer.finalize(harness.value, harness.state, caller.signal)
		.finally(() => { settled = true; });
	void active.catch(() => undefined);
	await new Promise((resolve) => setImmediate(resolve));
	caller.abort(new Error("stop requested"));
	await new Promise((resolve) => setImmediate(resolve));
	assert.equal(settled, false, "adapter promise must remain live while accepted work is joining");
	assert.equal(underlyingSettled, false);
	release.resolve();
	await assert.rejects(active, /cancelled|joined|abort/i);
	assert.equal(underlyingSettled, true);
	await finalizer.close();
});
