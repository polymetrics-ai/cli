import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFile } from "node:fs/promises";
import test from "node:test";

import {
	GitHubParentOrchestrator,
	createCanonicalPullRequestSnapshot,
	createDurableMutationIntent,
	createParentOrchestrationPlan,
	materializeChildRecord,
	selectReadyChildren,
	type ChildIntegrationReceipt,
	type CreateChildIssueRequest,
	type CreatePullRequestRequest,
	type AgentSessionAttestationSource,
	type AuthoritativeLookup,
	type ChildIntegrationQuery,
	type ChildIssueMarkerQuery,
	type ExternalCallContext,
	type DurableMutationResult,
	type GitHubChildIssue,
	type GitHubOrchestrationTransport,
	type GitAncestryProof,
	type GitAncestryQuery,
	type GitHubPullRequestQuery,
	type GitHubRosterQuery,
	type GitHubRosterSnapshot,
	type IntegrateChildRequest,
	type MarkParentReadyRequest,
	type ParentDecisionBroker,
	type ParentDecisionPolicy,
	type ParentReadyAuthorization,
	type ParentReadyAuthorityCoordinate,
	type ParentReadyCompareEffectResult,
	type ParentReadyDurableAuthorityBoundary,
	type ParentReadyFreshnessEnvelope,
	type ParentReadyJournalQuery,
	type ParentReadyOperationJournal,
	type ParentReadySettlementRecord,
	type PreparedParentReadyOperation,
	type ParentOrchestrationPlan,
	type PublishRosterRequest,
	type PullRequestMarkerQuery,
	type RequiredCheckPolicySource,
	type ParentReadyRecoveryFence,
	type RollbackParentReadyRequest,
} from "./github-orchestrator.ts";
import * as githubOrchestratorApi from "./github-orchestrator.ts";
import * as githubEvidenceApi from "./github-evidence.ts";
import type {
	GitHubChangedPathEvidence,
	GitHubPullRequestEvidence,
	RequiredGitHubCheckPolicyObservation,
} from "./github-evidence.ts";
import {
	createAgentSessionAttestation,
	createIndependentReviewWork,
	independentReviewAuthorizationDigest,
	independentReviewResultDigest,
} from "./review-router.ts";
import {
	createHumanDecisionRecord,
	consumeHumanDecision,
	recordHumanDecision,
	recordHumanDecisionRequestComment,
	type HumanDecisionEvidence,
	type HumanDecisionRecord,
	type HumanDecisionRepository,
} from "./human-decision.ts";
import {
	GitHubDecisionBroker,
	renderDecisionRequestComment,
	type GitHubComment,
	type GitHubDecisionPollOptions,
	type GitHubDecisionPollResult,
	type GitHubDecisionRequest,
	type GitHubDecisionTransport,
} from "./github-decision-broker.ts";
import type {
	ClaimedWorkspace,
	WorkspaceHandoffEvidence,
} from "./workspace-adapter.ts";

const objectivePath = ".pi/extensions/shepherd/fixtures/issue-478/parent-objective.json";
const baseSha = "a".repeat(40);
const headSha = "b".repeat(40);
type PullRequestFixtureRequest = Omit<CreatePullRequestRequest, "mutation" | "policyDigest">
	& Partial<Pick<CreatePullRequestRequest, "policyDigest">>;
type TestCallContext = { signal?: AbortSignal };

async function plan(): Promise<ParentOrchestrationPlan> {
	return cycle3Plan();
}

function createPlanFromSource(source: Record<string, unknown>): ParentOrchestrationPlan {
	return createParentOrchestrationPlan(source, {
		schemaVersion: 1,
		requiredCheckPolicies: [
			cycle3CheckPolicy(String(source.parentBranch)),
			cycle3CheckPolicy(String(source.parentBaseBranch)),
		] as never,
	});
}

function issueFrom(request: Omit<CreateChildIssueRequest, "mutation">, number = 811): GitHubChildIssue {
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
	request: PullRequestFixtureRequest,
	overrides: Partial<GitHubPullRequestEvidence> = {},
	number = 812,
): GitHubPullRequestEvidence {
	const review = {
		...createIndependentReviewWork({
			repository: request.repository,
			workItemId: request.workItemId,
			pullRequest: number,
			generation: request.generation,
			baseBranch: request.baseBranch,
			headBranch: request.headBranch,
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
		schemaVersion: 2,
		repository: request.repository,
		workItemId: request.workItemId,
		generation: request.generation,
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
		changedPathsComplete: true,
		changedPaths: [...request.changedPaths],
		allowedScopes: [...request.allowedScopes],
		mergeState: "clean",
		policyDigest: request.policyDigest ?? String(cycle3CheckPolicy(request.baseBranch).digest),
		checksComplete: true,
		checks: [{
			id: "check-verify-1",
			name: "verify",
			producerId: "github-actions:workflow-verify",
			sequence: 1,
			status: "completed",
			conclusion: "success",
			headSha: request.headSha,
			updatedAt: "2026-07-21T11:55:00.000Z",
			completedAt: "2026-07-21T11:55:00.000Z",
		}],
		requestedChangesComplete: true,
		requestedChanges: [],
		threadsComplete: true,
		threads: [],
		reviews: [review],
		reviewsComplete: true,
		dispositionsComplete: true,
		dispositions: [],
		revision: 42,
		observedAt: "2026-07-21T12:05:00.000Z",
		...overrides,
	};
}

function attestReview(review: GitHubPullRequestEvidence["reviews"][number]) {
	const attempt = review.completedAt.replace(/[^0-9]/gu, "");
	return createAgentSessionAttestation({
		sessionId: `session-${review.pullRequest}-${review.workItemId}-${attempt}`,
		runId: `run-${review.pullRequest}-${review.generation}-${attempt}`,
		review,
	});
}

class FakeTransport implements GitHubOrchestrationTransport, ParentReadyDurableAuthorityBoundary {
	issues: GitHubChildIssue[] = [];
	pullRequests: GitHubPullRequestEvidence[] = [];
	rosters: GitHubRosterSnapshot[] = [];
	integrations: ChildIntegrationReceipt[] = [];
	createIssueCalls = 0;
	createPullRequestCalls = 0;
	publishRosterCalls = 0;
	integrateCalls = 0;
	markReadyCalls = 0;
	rollbackReadyCalls = 0;
	throwAfterIssuePublish = false;
	throwAfterPullRequestPublish = false;
	throwAfterRosterPublish = false;
	throwAfterIntegration = false;
	throwAfterReady = false;
	malformedPullRequestResponse = false;
	malformedIntegrationResponse = false;
	malformedReadyResponse = false;
	issueVisibilityLag = 0;
	pullRequestVisibilityLag = 0;
	rosterVisibilityLag = 0;
	integrationVisibilityLag = 0;
	parentReadyVisibilityLag = 0;
	rosterRevisionDelta = 1;
	readyRevisionDelta = 1;
	callContexts: Array<{ operation: string; signal: boolean }> = [];
	incompleteIssueLookup = false;
	incompletePullRequestLookup = false;
	incompleteRosterLookup = false;
	incompleteIntegrationLookup = false;
	ancestry = true;
	ancestryProof?: (query: GitAncestryQuery) => Promise<unknown>;
	onPullRequestRead?: (query: PullRequestMarkerQuery, matches: GitHubPullRequestEvidence[], read: number) => GitHubPullRequestEvidence[];
	#pullRequestReads = 0;
	#issueInvisibleReads = 0;
	#pullRequestInvisibleReads = 0;
	#rosterInvisibleReads = 0;
	#integrationInvisibleReads = 0;
	#parentReadyInvisibleReads = 0;
	#legacyMutation = 0;
	#mutationRevision = 0;
	#mutations = new Map<string, { digest: string; revision: number; value: unknown }>();

	trackContext(operation: string, context?: TestCallContext): void {
		this.callContexts.push({ operation, signal: context?.signal instanceof AbortSignal });
	}

	#applyMutation(request: Record<string, unknown>, operation: string, effect: (revision: number) => unknown): any {
		const candidate = request.mutation as Record<string, unknown> | undefined;
		const key = typeof candidate?.idempotencyKey === "string"
			? candidate.idempotencyKey
			: `${operation}:legacy:${++this.#legacyMutation}`;
		const digest = typeof candidate?.intentDigest === "string" ? candidate.intentDigest : "legacy";
		const existing = this.#mutations.get(key);
		if (existing !== undefined) {
			if (existing.digest !== digest) throw new Error("durable mutation key reused with a different intent digest");
			return {
				schemaVersion: 1,
				idempotencyKey: key,
				intentDigest: digest,
				revision: existing.revision,
				applied: false,
				value: existing.value,
			};
		}
		const revision = this.#mutationRevision + 1;
		const value = effect(revision);
		this.#mutationRevision = revision;
		this.#mutations.set(key, { digest, revision, value });
		return {
			schemaVersion: 1,
			idempotencyKey: key,
			intentDigest: digest,
			revision,
			applied: true,
			value,
		};
	}

	async findChildIssues(query: { repository: string; marker: string }, context?: TestCallContext): Promise<any> {
		this.trackContext("findChildIssues", context);
		const hidden = this.#issueInvisibleReads > 0;
		if (hidden) this.#issueInvisibleReads -= 1;
		return {
			items: hidden ? [] : this.issues.filter((candidate) => candidate.marker === query.marker),
			complete: !this.incompleteIssueLookup,
		};
	}

	async createChildIssue(request: CreateChildIssueRequest, context?: TestCallContext): Promise<any> {
		this.trackContext("createChildIssue", context);
		const result = this.#applyMutation(request as unknown as Record<string, unknown>, "issue", () => {
			this.createIssueCalls += 1;
			const created = issueFrom(request, 810 + this.createIssueCalls);
			this.issues.push(created);
			this.#issueInvisibleReads = this.issueVisibilityLag;
			return created;
		});
		if (this.throwAfterIssuePublish) {
			this.throwAfterIssuePublish = false;
			throw new Error("simulated timeout after publish");
		}
		return result;
	}

	async findPullRequests(query: PullRequestMarkerQuery, context?: TestCallContext): Promise<any> {
		this.trackContext("findPullRequests", context);
		this.#pullRequestReads += 1;
		const matches = this.pullRequests.filter((candidate) => candidate.marker === query.marker);
		const hiddenAfterCreate = this.#pullRequestInvisibleReads > 0;
		const hiddenAfterReady = this.#parentReadyInvisibleReads > 0 && matches.some((candidate) => !candidate.draft);
		if (hiddenAfterCreate) this.#pullRequestInvisibleReads -= 1;
		if (hiddenAfterReady) this.#parentReadyInvisibleReads -= 1;
		const visible = hiddenAfterCreate || hiddenAfterReady ? [] : matches;
		return {
			items: this.onPullRequestRead?.(query, visible, this.#pullRequestReads) ?? visible,
			complete: !this.incompletePullRequestLookup,
		} as never;
	}

	async createPullRequest(request: CreatePullRequestRequest, context?: TestCallContext): Promise<any> {
		this.trackContext("createPullRequest", context);
		const result = this.#applyMutation(request as unknown as Record<string, unknown>, "pull-request", () => {
			this.createPullRequestCalls += 1;
			const created = cleanPullRequest(request, {
				draft: request.draft,
				checks: [],
				reviews: [],
			}, 900 + this.createPullRequestCalls);
			this.pullRequests.push(created);
			this.#pullRequestInvisibleReads = this.pullRequestVisibilityLag;
			return created;
		});
		if (this.throwAfterPullRequestPublish) {
			this.throwAfterPullRequestPublish = false;
			throw new Error("simulated timeout after pull request publish");
		}
		if (this.malformedPullRequestResponse) return { malformed: true } as never;
		return result;
	}

	async findParentRosters(query: GitHubRosterQuery, context?: TestCallContext): Promise<any> {
		this.trackContext("findParentRosters", context);
		const hidden = this.#rosterInvisibleReads > 0;
		if (hidden) this.#rosterInvisibleReads -= 1;
		return {
			items: hidden ? [] : this.rosters.filter((candidate) => candidate.marker === query.marker),
			complete: !this.incompleteRosterLookup,
		};
	}

	async publishParentRoster(request: PublishRosterRequest, context?: TestCallContext): Promise<any> {
		this.trackContext("publishParentRoster", context);
		const result = this.#applyMutation(request as unknown as Record<string, unknown>, "roster", () => {
			const existing = this.rosters.find((candidate) => candidate.marker === request.marker);
			if (request.mutation.expectedResourceRevision !== (existing?.revision ?? null)) {
				throw new Error("simulated roster conditional revision conflict");
			}
			this.publishRosterCalls += 1;
			const published: GitHubRosterSnapshot = {
				id: existing?.id ?? 1001,
				marker: request.marker,
				parentIssue: request.parentIssue,
				generation: request.generation,
				body: request.body,
				statuses: { ...request.statuses },
				statusEpoch: request.statusEpoch,
				revision: (existing?.revision ?? 0) + this.rosterRevisionDelta,
				updatedAt: "2026-07-21T12:00:00.000Z",
			};
			if (existing) this.rosters.splice(this.rosters.indexOf(existing), 1, published);
			else this.rosters.push(published);
			this.#rosterInvisibleReads = this.rosterVisibilityLag;
			return published;
		});
		if (this.throwAfterRosterPublish) {
			this.throwAfterRosterPublish = false;
			throw new Error("simulated timeout after roster publish");
		}
		return result;
	}

	async findChildIntegration(query: any, context?: TestCallContext): Promise<any> {
		this.trackContext("findChildIntegration", context);
		const hidden = this.#integrationInvisibleReads > 0;
		if (hidden) this.#integrationInvisibleReads -= 1;
		return {
			items: hidden ? [] : this.integrations.filter((candidate) => query.pullRequest !== undefined
				? candidate.pullRequest === query.pullRequest
				: candidate.childId === query.childId && candidate.marker === query.marker),
			complete: !this.incompleteIntegrationLookup,
		};
	}

	async integrateChild(request: IntegrateChildRequest, context?: TestCallContext): Promise<any> {
		this.trackContext("integrateChild", context);
		const result = this.#applyMutation(request as unknown as Record<string, unknown>, "integration", (revision) => {
			this.integrateCalls += 1;
			const receipt: ChildIntegrationReceipt = {
				childId: request.childId,
				pullRequest: request.pullRequest,
				generation: request.generation,
				marker: request.marker,
				baseSha: request.baseSha,
				headSha: request.headSha,
				parentBranch: request.parentBranch,
				pullRequestSnapshot: request.pullRequestSnapshot,
				observation: request.observation,
				controllerProvenance: request.controllerProvenance,
				transportProvenance: {
					authority: "transport",
					idempotencyKey: request.mutation.idempotencyKey,
					intentDigest: request.mutation.intentDigest,
					revision,
				},
				integratedAt: "2026-07-21T12:10:00.000Z",
			};
			this.integrations.push(receipt);
			this.#integrationInvisibleReads = this.integrationVisibilityLag;
			return receipt;
		});
		if (this.throwAfterIntegration) {
			this.throwAfterIntegration = false;
			throw new Error("simulated timeout after integration");
		}
		if (this.malformedIntegrationResponse) return { malformed: true } as never;
		return result;
	}

	async markParentReady(request: MarkParentReadyRequest, context?: TestCallContext): Promise<any> {
		this.trackContext("markParentReady", context);
		const authorization = request.authorization;
		const { digest, ...authorizationPayload } = authorization;
		const digests = [
			digest,
			authorization.decisionDigest,
			authorization.childRosterDigest,
			authorization.policySetDigest,
			authorization.parentReviewDigest,
			authorization.parentPathDigest,
			authorization.planDigest,
		];
		if (
			authorization.schemaVersion !== 1
			|| authorization.repository !== request.repository
			|| authorization.generation !== request.generation
			|| authorization.pullRequest !== request.pullRequest
			|| authorization.decisionRequestId !== request.decisionRequestId
			|| authorization.headSha !== request.headSha
			|| authorization.pullRequestRevision !== request.mutation.expectedResourceRevision
			|| createHash("sha256").update(JSON.stringify(authorizationPayload)).digest("hex") !== digest
			|| digests.some((digest) => !/^[0-9a-f]{64}$/u.test(digest))
		) {
			throw new Error("simulated parent ready authorization conflict");
		}
		const result = this.#applyMutation(request as unknown as Record<string, unknown>, "parent-ready", () => {
			const index = this.pullRequests.findIndex((candidate) => candidate.number === request.pullRequest);
			if (index < 0) throw new Error("parent pull request missing");
			if (request.mutation.expectedResourceRevision !== this.pullRequests[index].revision) {
				throw new Error("simulated parent ready conditional revision conflict");
			}
			this.markReadyCalls += 1;
			const updated = {
				...this.pullRequests[index],
				draft: false,
				headSha: request.headSha,
				revision: this.pullRequests[index].revision + this.readyRevisionDelta,
			};
			this.pullRequests.splice(index, 1, updated);
			this.#parentReadyInvisibleReads = this.parentReadyVisibilityLag;
			return updated;
		});
		if (this.throwAfterReady) {
			this.throwAfterReady = false;
			throw new Error("simulated timeout after ready transition");
		}
		if (this.malformedReadyResponse) return { malformed: true } as never;
		return result;
	}

	async rollbackParentReady(request: RollbackParentReadyRequest, context?: TestCallContext): Promise<any> {
		this.trackContext("rollbackParentReady", context);
		return this.#applyMutation(request as unknown as Record<string, unknown>, "parent-ready-rollback", () => {
			const index = this.pullRequests.findIndex((candidate) => candidate.number === request.pullRequest);
			if (index < 0) throw new Error("parent pull request missing");
			if (request.mutation.expectedResourceRevision !== this.pullRequests[index].revision) {
				throw new Error("simulated parent ready rollback conditional revision conflict");
			}
			this.rollbackReadyCalls += 1;
			const updated = {
				...this.pullRequests[index],
				draft: true,
				headSha: request.headSha,
				revision: this.pullRequests[index].revision + 1,
			};
			this.pullRequests.splice(index, 1, updated);
			return updated;
		});
	}

	async compareConsumeAndMarkParentReady(request: MarkParentReadyRequest, context?: TestCallContext): Promise<any> {
		return {
			schemaVersion: 1,
			kind: "applied",
			mutation: await this.markParentReady(request, context),
		};
	}

	async quarantineAndRollbackParentReady(request: RollbackParentReadyRequest, context?: TestCallContext): Promise<any> {
		return this.rollbackParentReady(request, context);
	}

	async isAncestor(_query: { repository: string; ancestorSha: string; descendantSha: string }): Promise<boolean> {
		return this.ancestry;
	}

	async proveAncestry(query: GitAncestryQuery, context?: TestCallContext): Promise<any> {
		this.trackContext("proveAncestry", context);
		if (this.ancestryProof !== undefined) return this.ancestryProof(query);
		return {
			schemaVersion: 1,
			authority: "transport",
			repository: query.repository,
			ancestorSha: query.ancestorSha,
			descendantSha: query.descendantSha,
			result: this.ancestry,
			revision: 1,
			observedAt: "2026-07-21T12:05:00.000Z",
		};
	}
}

function defaultPolicySource(transport: FakeTransport) {
	return {
		async findRequiredCheckPolicies(
			query: { repository: string; baseBranch: string },
			context?: TestCallContext,
		): Promise<unknown> {
			transport.trackContext("findRequiredCheckPolicies", context);
			const policy = cycle3CheckPolicy(query.baseBranch);
			return {
				items: [{
					schemaVersion: 1,
					authority: "controller",
					repository: query.repository,
					baseBranch: query.baseBranch,
					revision: policy.revision,
					digest: policy.digest,
					observedAt: "2026-07-21T12:06:00.000Z",
				}],
				complete: true,
			};
		},
	};
}

function orchestratorFor(
	transport: FakeTransport,
	broker?: ParentDecisionBroker,
	policySource: unknown = defaultPolicySource(transport),
	timeoutMs = 25,
	parentReadyAuthority: ParentReadyDurableAuthorityBoundary = transport,
	now: () => Date = () => new Date("2026-07-22T00:00:00.000Z"),
): GitHubParentOrchestrator {
	const sessionSource = {
		async findAttestations(target: { pullRequest: number; workItemId: string }, context?: TestCallContext): Promise<unknown> {
			transport.trackContext("findAttestations", context);
			const attestations = transport.pullRequests
				.filter((pullRequest) => pullRequest.number === target.pullRequest)
				.flatMap((pullRequest) => pullRequest.reviews)
				.filter((review) => review.workItemId === target.workItemId)
				.map(attestReview);
			return { items: attestations, complete: true };
		},
		async findChangedPathEvidence(target: {
			repository: string;
			workItemId: string;
			pullRequest: number;
			generation: number;
			baseSha: string;
			headSha: string;
		}, context?: TestCallContext): Promise<unknown> {
			transport.trackContext("findChangedPathEvidence", context);
			const matches = transport.pullRequests
				.filter((pullRequest) => pullRequest.number === target.pullRequest
					&& pullRequest.repository === target.repository
					&& pullRequest.workItemId === target.workItemId
					&& pullRequest.generation === target.generation
					&& pullRequest.baseSha === target.baseSha
					&& pullRequest.headSha === target.headSha)
				.map((pullRequest) => ({
					schemaVersion: 1,
					authority: "controller",
					repository: target.repository,
					workItemId: target.workItemId,
					pullRequest: target.pullRequest,
					generation: target.generation,
					baseSha: target.baseSha,
					headSha: target.headSha,
					paths: [...pullRequest.changedPaths],
					complete: true,
					revision: Math.max(1, pullRequest.revision - 1),
					observedAt: "2026-07-21T11:58:00.000Z",
				}));
			return { items: matches, complete: true };
		},
	};
	return new (GitHubParentOrchestrator as any)(
		transport,
		broker,
		sessionSource,
		policySource,
		{
			externalCallTimeoutMs: timeoutMs,
			parentReadyAuthority,
			now,
		},
	) as GitHubParentOrchestrator;
}

const approvedDecision: HumanDecisionEvidence = {
	option: "approve-merge",
	actor: "maintainer",
	sourceUrl: "https://github.com/polymetrics-ai/cli/pull/900#issuecomment-2",
	decidedAt: "2026-07-21T12:00:30.000Z",
};

class FakeDecisionBroker implements ParentDecisionBroker {
	requests: GitHubDecisionRequest[] = [];
	consumes = 0;
	callContexts: Array<{ operation: string; signal: boolean }> = [];
	pollResult: GitHubDecisionPollResult = { status: "decided", decision: approvedDecision, attempts: 1 };
	recordStatus: HumanDecisionRecord["status"] = "pending";

	private recordFor(request: GitHubDecisionRequest): HumanDecisionRecord {
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
		} as Parameters<typeof createHumanDecisionRecord>[0], new Date("2026-07-21T12:00:00.000Z"));
		return {
			...record,
			requestComment: {
				id: 1,
				url: `https://github.com/${request.repository}/pull/${request.pullRequest}#issuecomment-1`,
				actor: "shepherd-host",
				createdAt: "2026-07-21T12:00:00.000Z",
			},
			updatedAt: "2026-07-21T12:00:00.000Z",
		};
	}

	async request(request: GitHubDecisionRequest, context?: TestCallContext): Promise<HumanDecisionRecord> {
		this.callContexts.push({ operation: "broker.request", signal: context?.signal instanceof AbortSignal });
		this.requests.push(request);
		if (request.gate !== "parent_merge") throw new Error("fake parent broker accepts only parent_merge requests");
		const record = this.recordFor(request);
		return this.recordStatus === "consumed"
			? {
				...record,
				status: "consumed",
				decision: this.pollResult.status === "decided" ? this.pollResult.decision : approvedDecision,
				consumedAt: "2026-07-21T12:00:40.000Z",
				updatedAt: "2026-07-21T12:00:40.000Z",
			}
			: record;
	}

	async poll(
		_requestId: string,
		_binding: HumanDecisionRecord["binding"],
		_options?: GitHubDecisionPollOptions,
		context?: TestCallContext,
	): Promise<HumanDecisionRecord> {
		this.callContexts.push({ operation: "broker.poll", signal: context?.signal instanceof AbortSignal });
		const record = this.recordFor(this.requests.at(-1)!);
		if (this.pollResult.status === "pending") return record;
		if (this.pollResult.status === "expired") {
			return { ...record, status: "expired", updatedAt: record.expiresAt };
		}
		return {
			...record,
			status: "decided",
			decision: this.pollResult.decision,
			updatedAt: this.pollResult.decision.decidedAt,
		};
	}

	async consume(_requestId: string, _binding: HumanDecisionRecord["binding"], context?: TestCallContext): Promise<HumanDecisionRecord> {
		this.callContexts.push({ operation: "broker.consume", signal: context?.signal instanceof AbortSignal });
		this.consumes += 1;
		const decision = this.pollResult.status === "decided" ? this.pollResult.decision : approvedDecision;
		this.recordStatus = "consumed";
		return {
			...this.recordFor(this.requests.at(-1)!),
			status: "consumed",
			decision,
			consumedAt: "2026-07-21T12:00:40.000Z",
			updatedAt: "2026-07-21T12:00:40.000Z",
		};
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

function childPullRequestRequest(
	candidate: ParentOrchestrationPlan,
	child: ReturnType<typeof materializeChildRecord>,
	handoff: WorkspaceHandoffEvidence,
): PullRequestFixtureRequest {
	return {
		repository: candidate.repository,
		workItemId: child.id,
		generation: candidate.generation,
		marker: child.markers.pullRequest,
		title: child.title,
		body: `Refs #${child.issue}\nRefs #${candidate.parentIssue}\n\n${child.markers.pullRequest}`,
		draft: false,
		baseBranch: child.prBase,
		headBranch: child.branch,
		baseSha: handoff.baseHead,
		headSha: handoff.head,
		changedPaths: handoff.changedScope,
		allowedScopes: child.writeScopes,
	};
}

const decisionPolicy: ParentDecisionPolicy = {
	requestId: "parent-471-generation-3",
	actorAllowlist: ["maintainer"],
	expiresAt: "2027-07-21T12:00:00.000Z",
	question: "Approve the exact reviewed parent head for the human merge gate?",
};

type Cycle6BrokerAdapterFactory = typeof githubOrchestratorApi.adaptGitHubDecisionBroker;
type Cycle6ReviewAuthorizationDigest = ChildIntegrationReceipt["controllerProvenance"]["reviewAuthorizationDigest"];
type Cycle6ReviewCompletedAt = ChildIntegrationReceipt["controllerProvenance"]["reviewCompletedAt"];

class MemoryHumanDecisionRepository implements HumanDecisionRepository {
	readonly records = new Map<string, HumanDecisionRecord>();

	async load(requestId: string): Promise<HumanDecisionRecord | null> {
		const record = this.records.get(requestId);
		return record === undefined ? null : structuredClone(record);
	}

	async transact<T>(
		requestId: string,
		operation: (state: HumanDecisionRecord | null) => Promise<{ state: HumanDecisionRecord | null; value: T }>
			| { state: HumanDecisionRecord | null; value: T },
	): Promise<T> {
		const current = await this.load(requestId);
		const result = await operation(current);
		if (result.state === null) this.records.delete(requestId);
		else this.records.set(requestId, structuredClone(result.state));
		return result.value;
	}
}

function controllerProvenanceFor(
	candidate: ParentOrchestrationPlan,
	pullRequest: GitHubPullRequestEvidence,
) {
	const changedPathEvidence = {
		schemaVersion: 1 as const,
		authority: "controller" as const,
		repository: pullRequest.repository,
		workItemId: pullRequest.workItemId,
		pullRequest: pullRequest.number,
		generation: pullRequest.generation,
		baseSha: pullRequest.baseSha,
		headSha: pullRequest.headSha,
		paths: [...pullRequest.changedPaths],
		complete: true as const,
		revision: Math.max(1, pullRequest.revision - 1),
		observedAt: "2026-07-21T11:58:00.000Z",
	};
	const { revision: _revision, observedAt: _observedAt, ...stableChangedPathEvidence } = changedPathEvidence;
	const review = pullRequest.reviews[0];
	assert.ok(review);
	const policy = cycle3CheckPolicy(pullRequest.baseBranch);
	return {
		authority: "controller" as const,
		planDigest: candidate.canonical.digest,
		policyDigest: String(policy.digest),
		policyRevision: Number(policy.revision),
		policyObservedAt: "2026-07-21T12:06:00.000Z",
		changedPathDigest: createHash("sha256").update(JSON.stringify(stableChangedPathEvidence)).digest("hex"),
		reviewAuthorizationDigest: independentReviewAuthorizationDigest(review),
		reviewResultDigest: independentReviewResultDigest(review),
		reviewCompletedAt: review.completedAt,
		evidenceRevision: changedPathEvidence.revision,
		observedAt: changedPathEvidence.observedAt,
	};
}

function integrationMutationProjection(value: {
	repository: string;
	childId: string;
	pullRequest: number;
	generation: number;
	marker: string;
	baseSha: string;
	headSha: string;
	parentBranch: string;
	pullRequestSnapshot: ReturnType<typeof createCanonicalPullRequestSnapshot>;
	controllerProvenance: ReturnType<typeof controllerProvenanceFor>;
}) {
	return {
		repository: value.repository,
		childId: value.childId,
		pullRequest: value.pullRequest,
		generation: value.generation,
		marker: value.marker,
		baseSha: value.baseSha,
		headSha: value.headSha,
		parentBranch: value.parentBranch,
		pullRequestIdentityDigest: value.pullRequestSnapshot.digest,
		planDigest: value.controllerProvenance.planDigest,
		policyDigest: value.controllerProvenance.policyDigest,
		policyRevision: value.controllerProvenance.policyRevision,
		changedPathDigest: value.controllerProvenance.changedPathDigest,
		reviewAuthorizationDigest: value.controllerProvenance.reviewAuthorizationDigest,
	};
}

function seedIntegrationRoster(
	candidate: ParentOrchestrationPlan,
	transport: FakeTransport,
): ChildIntegrationReceipt[] {
	const receipts = candidate.children.map((child, index): ChildIntegrationReceipt => {
		const pullRequestNumber = 812 + index;
		const issueNumber = 811 + index;
		const childHead = String(index + 1).repeat(40);
		transport.issues.push(issueFrom({
			repository: candidate.repository,
			parentIssue: candidate.parentIssue,
			marker: child.markers.issue,
			title: child.title,
			body: child.issueBody,
		}, issueNumber));
		const childPullRequest = cleanPullRequest({
			repository: candidate.repository,
			workItemId: child.id,
			generation: candidate.generation,
			marker: child.markers.pullRequest,
			title: child.title,
			body: `Refs #${issueNumber}\nRefs #${candidate.parentIssue}\n\n${child.markers.pullRequest}`,
			draft: false,
			baseBranch: candidate.parentBranch,
			headBranch: `feat/${issueNumber}-${child.branch.slug}`,
			baseSha,
			headSha: childHead,
			changedPaths: [child.writeScopes[0]],
			allowedScopes: child.writeScopes,
		}, { state: "merged" }, pullRequestNumber);
		transport.pullRequests.push(childPullRequest);
		const snapshot = createCanonicalPullRequestSnapshot(childPullRequest);
		const observation = {
			revision: childPullRequest.revision,
			observedAt: childPullRequest.observedAt,
			state: childPullRequest.state,
		};
		const controllerProvenance = controllerProvenanceFor(candidate, childPullRequest);
		const mutation = createDurableMutationIntent(
			"child_integration",
			[candidate.repository, child.markers.pullRequest],
			integrationMutationProjection({
				repository: candidate.repository,
				childId: child.id,
				pullRequest: pullRequestNumber,
				generation: candidate.generation,
				marker: child.markers.pullRequest,
				baseSha,
				headSha: childHead,
				parentBranch: candidate.parentBranch,
				pullRequestSnapshot: snapshot,
				controllerProvenance,
			}),
			null,
		);
		return {
			childId: child.id,
			pullRequest: pullRequestNumber,
			generation: candidate.generation,
			marker: child.markers.pullRequest,
			baseSha,
			headSha: childHead,
			parentBranch: candidate.parentBranch,
			pullRequestSnapshot: snapshot,
			observation,
			controllerProvenance,
			transportProvenance: {
				authority: "transport",
				idempotencyKey: mutation.idempotencyKey,
				intentDigest: mutation.intentDigest,
				revision: index + 1,
			},
			integratedAt: "2026-07-21T12:10:00.000Z",
		};
	});
	transport.integrations.push(...receipts);
	return receipts;
}

function addParentPullRequest(
	candidate: ParentOrchestrationPlan,
	transport: FakeTransport,
	changedPaths = [candidate.children[0].writeScopes[0]],
): GitHubPullRequestEvidence {
	const parent = cleanPullRequest({
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
		changedPaths,
		allowedScopes: candidate.children.flatMap((child) => child.writeScopes),
	}, { draft: true }, 900);
	transport.pullRequests.push(parent);
	return parent;
}

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
	assert.throws(() => createPlanFromSource({ ...source, unexpected: true }), /field|shape|parent/i);
	const children = source.children as Array<Record<string, unknown>>;
	assert.throws(() => createPlanFromSource({ ...source, children: [{ ...children[0], dependsOn: ["evidence"] }] }), /cycle/i);
	assert.throws(() => createPlanFromSource({ ...source, parentBranch: "../main" }), /branch/i);
	for (const invalidRef of [
		"bad branch", ".hidden/topic", "topic.lock", "topic/.lock", "@", "HEAD", "refs/heads/HEAD",
		"FETCH_HEAD", "ORIG_HEAD", "MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD", "REBASE_HEAD",
		"BISECT_HEAD", "AUTO_MERGE", "topic/FETCH_HEAD", "refs/heads/topic", "topic@{one", "topic.",
	]) {
		assert.throws(() => createPlanFromSource({ ...source, parentBranch: invalidRef }), /branch|ref/i, invalidRef);
	}
	for (const invalidGeneration of [0, -1]) {
		assert.throws(() => createPlanFromSource({ ...source, generation: invalidGeneration }), /generation|positive/i);
	}
	assert.throws(() => createPlanFromSource({ ...source, children: [{ ...children[0], writeScopes: ["../outside"] }] }), /scope/i);
	assert.throws(() => createPlanFromSource({ ...source, children: [{ ...children[0], requiredSkills: [] }] }), /skills/i);
	assert.throws(() => createPlanFromSource({ ...source, children: Array.from({ length: 65 }, (_, index) => ({ ...children[0], id: `child-${index}`, slug: `child-${index}`, dependsOn: [] })) }), /64|bounded|children/i);
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
	const transport = new FakeTransport();
	transport.issues.push(issue);
	const orchestrator = orchestratorFor(transport);
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
	const first = await orchestrator.reconcileParentRoster(candidate, pending, 1);
	const restarted = await orchestratorFor(transport).reconcileParentRoster(candidate, pending, 1);
	assert.deepEqual(restarted, first);
	assert.equal(transport.publishRosterCalls, 1);
	await orchestrator.reconcileParentRoster(candidate, { evidence: "succeeded", orchestrator: "pending" }, 2);
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
	transport.issues.push(issue);
	const handoff = childHandoff(issue.number, child.branch, child.prBase);
	const request: PullRequestFixtureRequest = {
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
	movedTransport.issues.push(issue);
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
	const request: PullRequestFixtureRequest = {
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
	const request: PullRequestFixtureRequest = {
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
		transport.issues.push(issue);
		const pr = cleanPullRequest(request);
		pr.reviews = [{
			...createIndependentReviewWork({
				repository: candidate.repository,
				workItemId: child.id,
				pullRequest: pr.number,
				generation: candidate.generation,
				baseBranch: child.prBase,
				headBranch: child.branch,
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

test("restart reuses stable integration identity after a later merged-PR observation", async () => {
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
	const request: PullRequestFixtureRequest = {
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
	transport.issues.push(issue);
	const mergedPullRequest = cleanPullRequest(request, { state: "merged" });
	transport.pullRequests.push(mergedPullRequest);
	const snapshot = createCanonicalPullRequestSnapshot(mergedPullRequest);
	const observation = {
		revision: mergedPullRequest.revision,
		observedAt: mergedPullRequest.observedAt,
		state: mergedPullRequest.state,
	};
	const controllerProvenance = controllerProvenanceFor(candidate, mergedPullRequest);
	const existingMutation = createDurableMutationIntent(
		"child_integration",
		[candidate.repository, child.markers.pullRequest],
		integrationMutationProjection({
			repository: candidate.repository,
			childId: child.id,
			pullRequest: 812,
			generation: candidate.generation,
			marker: child.markers.pullRequest,
			baseSha: handoff.baseHead,
			headSha: handoff.head,
			parentBranch: candidate.parentBranch,
			pullRequestSnapshot: snapshot,
			controllerProvenance,
		}),
		null,
	);
	transport.integrations.push({
		childId: child.id,
		pullRequest: 812,
		generation: candidate.generation,
		marker: child.markers.pullRequest,
		baseSha: handoff.baseHead,
		headSha: handoff.head,
		parentBranch: candidate.parentBranch,
		pullRequestSnapshot: snapshot,
		observation,
		controllerProvenance,
		transportProvenance: {
			authority: "transport",
			idempotencyKey: existingMutation.idempotencyKey,
			intentDigest: existingMutation.intentDigest,
			revision: 1,
		},
		integratedAt: "2026-07-21T12:10:00.000Z",
	});
	transport.pullRequests[0] = {
		...mergedPullRequest,
		revision: mergedPullRequest.revision + 1,
		observedAt: "2026-07-21T12:06:00.000Z",
	};
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
	const parentRequest: PullRequestFixtureRequest = {
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
	const receipts = seedIntegrationRoster(candidate, transport);

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
	const seed = new FakeTransport();
	const receipts = seedIntegrationRoster(candidate, seed);
	const makeTransport = (): FakeTransport => {
		const transport = new FakeTransport();
		transport.pullRequests.push(...structuredClone(seed.pullRequests));
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
	const seed = new FakeTransport();
	const receipts = seedIntegrationRoster(candidate, seed);
	const setup = (): FakeTransport => {
		const transport = new FakeTransport();
		transport.pullRequests.push(...structuredClone(seed.pullRequests));
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

	for (const forge of [
		(receipt: any) => { receipt.controllerProvenance.planDigest = "f".repeat(64); },
		(receipt: any) => { receipt.transportProvenance.intentDigest = "f".repeat(64); },
		(receipt: any) => { receipt.pullRequestSnapshot.number = 9_999; },
	]) {
		const transport = setup();
		const authoritativeForgery = structuredClone(receipts);
		forge(authoritativeForgery[0]);
		transport.integrations.splice(0, transport.integrations.length, ...authoritativeForgery);
		const result = await orchestratorFor(transport, new FakeDecisionBroker())
			.reconcileParentReadiness(candidate, authoritativeForgery, decisionPolicy);
		assert.equal(result.kind, "blocked");
		assert.equal(transport.markReadyCalls, 0);
	}

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
	for (const mode of ["timeout", "malformed"] as const) {
		const transport = new FakeTransport();
		const receipts = seedIntegrationRoster(candidate, transport);
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

function cycle3CheckPolicy(baseBranch: string, overrides: Record<string, unknown> = {}): Record<string, unknown> {
	const input = {
		schemaVersion: 1,
		repository: "polymetrics-ai/cli",
		baseBranch,
		revision: 7,
		requiredChecks: [
			{ name: "verify", producerId: "github-actions:workflow-verify" },
		],
		...overrides,
	};
	const create = (githubEvidenceApi as Record<string, unknown>).createRequiredGitHubCheckPolicy;
	return typeof create === "function"
		? (create as (value: unknown) => Record<string, unknown>)(input)
		: { ...input, digest: "0".repeat(64) };
}

async function cycle3Plan(singleChild = false, policyOverrides: Record<string, unknown> = {}): Promise<ParentOrchestrationPlan> {
	const source = JSON.parse(await readFile(objectivePath, "utf8")) as Record<string, unknown>;
	if (singleChild) source.children = [(source.children as unknown[])[0]];
	const bundle = {
		schemaVersion: 1,
		requiredCheckPolicies: [
			cycle3CheckPolicy(String(source.parentBranch), policyOverrides),
			cycle3CheckPolicy(String(source.parentBaseBranch)),
		],
	};
	return (createParentOrchestrationPlan as unknown as (
		value: unknown,
		policies: unknown,
	) => ParentOrchestrationPlan)(source, bundle);
}

test("cycle 3 persists an opaque canonical plan and revalidates every cloned public boundary", async () => {
	const candidate = await cycle3Plan();
	const canonical = (candidate as unknown as { canonical: Record<string, unknown> }).canonical;
	assert.equal(canonical.schemaVersion, 1);
	assert.equal(typeof canonical.serialized, "string");
	assert.match(String(canonical.digest), /^[0-9a-f]{64}$/u);
	const persisted = JSON.parse(JSON.stringify(candidate)) as ParentOrchestrationPlan;
	assert.deepEqual(selectReadyChildren(persisted, { evidence: "pending", orchestrator: "pending" }, 2), {
		kind: "selected",
		itemIds: ["evidence"],
	});

	const tamperedValues = [
		{ ...persisted, repository: "other/cli" },
		{ ...persisted, generation: persisted.generation + 1 },
		{ ...persisted, parentBranch: "feat/tampered-parent" },
		{ ...persisted, markers: { ...persisted.markers, roster: "<!-- forged -->" } },
		{ ...persisted, children: persisted.children.map((child, index) => index === 0
			? { ...child, title: "tampered title" }
			: child) },
		{ ...persisted, canonical: { ...canonical, digest: "f".repeat(64) } },
		{ ...persisted, canonical: { ...canonical, serialized: `${String(canonical.serialized)} ` } },
	];
	for (const tampered of tamperedValues) {
		assert.throws(
			() => selectReadyChildren(tampered as ParentOrchestrationPlan, { evidence: "pending", orchestrator: "pending" }, 1),
			/canonical|digest|plan|provenance|tamper/i,
	);
}

	const transport = new FakeTransport();
	await assert.rejects(
		orchestratorFor(transport).ensureChildIssue(tamperedValues[0] as ParentOrchestrationPlan, "evidence"),
		/canonical|digest|plan|provenance|tamper/i,
	);
	assert.equal(transport.createIssueCalls, 0);
});

test("cycle 3 rejects plan proxies, accessors, cycles, unknown/oversized canonical state, and secrets before effects", async () => {
	const candidate = await cycle3Plan();
	let trapInvoked = false;
	const proxied = new Proxy(candidate, {
		get() {
			trapInvoked = true;
			throw new Error("SECRET_PLAN_TOKEN_SHOULD_NOT_ESCAPE");
		},
	});
	assert.throws(
		() => selectReadyChildren(proxied, { evidence: "pending", orchestrator: "pending" }, 1),
		/plan|proxy|canonical|shape/i,
	);
	assert.equal(trapInvoked, false);

	const accessor = { ...candidate } as Record<string, unknown>;
	Object.defineProperty(accessor, "repository", {
		enumerable: true,
		get() {
			trapInvoked = true;
			throw new Error("SECRET_PLAN_TOKEN_SHOULD_NOT_ESCAPE");
		},
	});
	assert.throws(
		() => selectReadyChildren(accessor as unknown as ParentOrchestrationPlan, { evidence: "pending", orchestrator: "pending" }, 1),
		/plan|accessor|canonical|shape/i,
	);
	assert.equal(trapInvoked, false);

	const cycle = JSON.parse(JSON.stringify(candidate)) as ParentOrchestrationPlan & { loop?: unknown };
	cycle.loop = cycle;
	assert.throws(() => selectReadyChildren(cycle, { evidence: "pending", orchestrator: "pending" }, 1), /cycle|field|plan|canonical/i);
	const canonical = (candidate as unknown as { canonical: Record<string, unknown> }).canonical;
	assert.throws(() => selectReadyChildren({
		...candidate,
		canonical: { ...canonical, serialized: "x".repeat(1_000_001) },
	} as unknown as ParentOrchestrationPlan, { evidence: "pending", orchestrator: "pending" }, 1), /bounded|canonical|serialized|plan/i);
	try {
		selectReadyChildren({ ...candidate, unexpected: "SECRET_PLAN_TOKEN_SHOULD_NOT_ESCAPE" } as ParentOrchestrationPlan,
			{ evidence: "pending", orchestrator: "pending" }, 1);
		assert.fail("unknown plan field must fail");
	} catch (error) {
		assert.doesNotMatch(String(error), /SECRET_PLAN_TOKEN_SHOULD_NOT_ESCAPE/u);
	}
});

test("cycle 3 rejects top-level read-only children and binds versioned CI policy into markers and digest", async () => {
	const source = JSON.parse(await readFile(objectivePath, "utf8")) as Record<string, unknown>;
	const children = source.children as Array<Record<string, unknown>>;
	await assert.rejects(async () => (createParentOrchestrationPlan as any)({
		...source,
		children: [{ ...children[0], access: "read_only", writeScopes: [] }],
	}, {
		schemaVersion: 1,
		requiredCheckPolicies: [cycle3CheckPolicy(String(source.parentBranch)), cycle3CheckPolicy(String(source.parentBaseBranch))],
	}), /read.only|mutating|scope|child/i);
	const current = await cycle3Plan();
	const moved = await cycle3Plan(false, { revision: 8 });
	assert.notEqual(
		(current as unknown as { canonical: { digest: string } }).canonical.digest,
		(moved as unknown as { canonical: { digest: string } }).canonical.digest,
	);
	assert.notEqual(current.markers.parentPullRequest, moved.markers.parentPullRequest);
});

test("cycle 3 durable mutation metadata deduplicates issue, PR, roster, integration, and ready effects across orchestrators", async () => {
	const candidate = await cycle3Plan(true);
	const transport = new FakeTransport();
	const firstOrchestrator = orchestratorFor(transport, new FakeDecisionBroker());
	const secondOrchestrator = orchestratorFor(transport, new FakeDecisionBroker());
	const [firstIssue, secondIssue] = await Promise.all([
		firstOrchestrator.ensureChildIssue(candidate, "evidence"),
		secondOrchestrator.ensureChildIssue(JSON.parse(JSON.stringify(candidate)), "evidence"),
	]);
	assert.deepEqual(firstIssue, secondIssue);
	assert.equal(transport.createIssueCalls, 1);
	const child = materializeChildRecord(candidate, "evidence", firstIssue);
	const handoff = childHandoff(child.issue, child.branch, child.prBase);
	const [firstPr, secondPr] = await Promise.all([
		firstOrchestrator.ensureChildPullRequest(candidate, child, handoff),
		secondOrchestrator.ensureChildPullRequest(JSON.parse(JSON.stringify(candidate)), child, handoff),
	]);
	assert.deepEqual(firstPr, secondPr);
	assert.equal(transport.createPullRequestCalls, 1);

	const statuses = { evidence: "running" } as const;
	await Promise.all([
		(firstOrchestrator.reconcileParentRoster as any)(candidate, statuses, 1),
		(secondOrchestrator.reconcileParentRoster as any)(JSON.parse(JSON.stringify(candidate)), statuses, 1),
	]);
	assert.equal(transport.publishRosterCalls, 1);

	const request: PullRequestFixtureRequest = {
		repository: candidate.repository,
		workItemId: child.id,
		generation: candidate.generation,
		marker: child.markers.pullRequest,
		title: child.title,
		body: `Refs #${child.issue}\nRefs #${candidate.parentIssue}\n\n${child.markers.pullRequest}`,
		draft: false,
		baseBranch: child.prBase,
		headBranch: child.branch,
		baseSha: handoff.baseHead,
		headSha: handoff.head,
		changedPaths: handoff.changedScope,
		allowedScopes: child.writeScopes,
	};
	transport.pullRequests[0] = cleanPullRequest(request, {}, firstPr.number);
	const [firstIntegration, secondIntegration] = await Promise.all([
		firstOrchestrator.integrateChild(candidate, child, handoff),
		secondOrchestrator.integrateChild(JSON.parse(JSON.stringify(candidate)), child, handoff),
	]);
	assert.equal(firstIntegration.kind, "integrated");
	assert.equal(secondIntegration.kind, "integrated");
	assert.equal(transport.integrateCalls, 1);
	const receipt = firstIntegration.kind === "integrated" ? firstIntegration.receipt : undefined;
	assert.ok(receipt);
	assert.deepEqual(
		{
			repository: (receipt as any).pullRequestSnapshot.repository,
			number: (receipt as any).pullRequestSnapshot.number,
			marker: (receipt as any).pullRequestSnapshot.marker,
			baseBranch: (receipt as any).pullRequestSnapshot.baseBranch,
			headBranch: (receipt as any).pullRequestSnapshot.headBranch,
			baseSha: (receipt as any).pullRequestSnapshot.baseSha,
			headSha: (receipt as any).pullRequestSnapshot.headSha,
			changedPaths: (receipt as any).pullRequestSnapshot.changedPaths,
			generation: (receipt as any).pullRequestSnapshot.generation,
		},
		{
			repository: candidate.repository,
			number: firstPr.number,
			marker: child.markers.pullRequest,
			baseBranch: child.prBase,
			headBranch: child.branch,
			baseSha: handoff.baseHead,
			headSha: handoff.head,
			changedPaths: handoff.changedScope,
			generation: candidate.generation,
		},
	);
	assert.equal((receipt as any).controllerProvenance.planDigest,
		(candidate as unknown as { canonical: { digest: string } }).canonical.digest);
	assert.match((receipt as any).transportProvenance.intentDigest, /^[0-9a-f]{64}$/u);

	const parentHead = "e".repeat(40);
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
		changedPaths: handoff.changedScope,
		allowedScopes: child.writeScopes,
	}, { draft: true }, 900));
	await Promise.all([
		firstOrchestrator.reconcileParentReadiness(candidate, [receipt!], decisionPolicy),
		secondOrchestrator.reconcileParentReadiness(JSON.parse(JSON.stringify(candidate)), [receipt!], decisionPolicy),
	]);
	assert.equal(transport.markReadyCalls, 1);
});

test("cycle 3 retries authoritative visibility after every timeout-after-effect mutation", async () => {
	const candidate = await cycle3Plan(true);
	const transport = new FakeTransport();
	const broker = new FakeDecisionBroker();
	const orchestrator = orchestratorFor(transport, broker);

	transport.issueVisibilityLag = 2;
	transport.throwAfterIssuePublish = true;
	const issue = await orchestrator.ensureChildIssue(candidate, "evidence");
	assert.equal(transport.createIssueCalls, 1);
	const child = materializeChildRecord(candidate, "evidence", issue);
	const handoff = childHandoff(child.issue, child.branch, child.prBase);

	transport.pullRequestVisibilityLag = 2;
	transport.throwAfterPullRequestPublish = true;
	const pullRequest = await orchestrator.ensureChildPullRequest(candidate, child, handoff);
	assert.equal(transport.createPullRequestCalls, 1);

	transport.rosterVisibilityLag = 2;
	transport.throwAfterRosterPublish = true;
	const roster = await (orchestrator.reconcileParentRoster as any)(candidate, { evidence: "running" }, 1);
	assert.equal((roster as GitHubRosterSnapshot).statusEpoch, 1);
	assert.equal(transport.publishRosterCalls, 1);

	const request: PullRequestFixtureRequest = {
		repository: candidate.repository,
		workItemId: child.id,
		generation: candidate.generation,
		marker: child.markers.pullRequest,
		title: child.title,
		body: `Refs #${child.issue}\nRefs #${candidate.parentIssue}\n\n${child.markers.pullRequest}`,
		draft: false,
		baseBranch: child.prBase,
		headBranch: child.branch,
		baseSha: handoff.baseHead,
		headSha: handoff.head,
		changedPaths: handoff.changedScope,
		allowedScopes: child.writeScopes,
	};
	const pullRequestIndex = transport.pullRequests.findIndex((candidatePr) => candidatePr.number === pullRequest.number);
	transport.pullRequests[pullRequestIndex] = cleanPullRequest(request, {}, pullRequest.number);
	transport.integrationVisibilityLag = 2;
	transport.throwAfterIntegration = true;
	const integrated = await orchestrator.integrateChild(candidate, child, handoff);
	assert.equal(integrated.kind, "integrated");
	assert.equal(transport.integrateCalls, 1);
	if (integrated.kind !== "integrated") throw new Error("integration visibility recovery failed");

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
		changedPaths: handoff.changedScope,
		allowedScopes: child.writeScopes,
	}, { draft: true }, 900));
	transport.parentReadyVisibilityLag = 2;
	transport.throwAfterReady = true;
	const ready = await orchestrator.reconcileParentReadiness(candidate, [integrated.receipt], decisionPolicy);
	assert.equal(ready.kind, "ready");
	assert.equal(transport.markReadyCalls, 1);
	for (const operation of [
		"findChildIssues", "createChildIssue", "findPullRequests", "createPullRequest",
		"findParentRosters", "publishParentRoster", "findChildIntegration", "integrateChild",
		"findChangedPathEvidence", "findAttestations", "findRequiredCheckPolicies", "proveAncestry",
		"markParentReady",
	]) {
		assert.ok(transport.callContexts.some((entry) => entry.operation === operation && entry.signal), operation);
	}
	for (const operation of ["broker.request", "broker.poll", "broker.consume"]) {
		assert.ok(broker.callContexts.some((entry) => entry.operation === operation && entry.signal), operation);
	}
});

test("cycle 3 roster revisions and status epochs reject regression and drain rejected FIFO work", async () => {
	const candidate = await cycle3Plan(true);
	const transport = new FakeTransport();
	const orchestrator = orchestratorFor(transport);
	const succeeded = await (orchestrator.reconcileParentRoster as any)(candidate, { evidence: "succeeded" }, 2);
	assert.equal((succeeded as any).statusEpoch, 2);
	await assert.rejects(
		(orchestratorFor(transport).reconcileParentRoster as any)(JSON.parse(JSON.stringify(candidate)), { evidence: "running" }, 1),
		/epoch|stale|regress|revision|conditional/i,
	);
	assert.equal(transport.publishRosterCalls, 1);
	const afterRejection = await (orchestrator.reconcileParentRoster as any)(candidate, { evidence: "succeeded" }, 3);
	assert.equal((afterRejection as any).statusEpoch, 3);
	assert.equal(transport.publishRosterCalls, 2);
});

test("cycle 3 requires an exact literal-true ancestry proof and rejects truthy or mismatched coordinates", async () => {
	const candidate = await cycle3Plan(true);
	const setup = async (proof: (query: GitAncestryQuery) => Promise<unknown>) => {
		const transport = new FakeTransport();
		transport.ancestryProof = proof;
		const orchestrator = orchestratorFor(transport, new FakeDecisionBroker());
		const issue = await orchestrator.ensureChildIssue(candidate, "evidence");
		const child = materializeChildRecord(candidate, "evidence", issue);
		const handoff = childHandoff(child.issue, child.branch, child.prBase);
		const childRequest: PullRequestFixtureRequest = {
			repository: candidate.repository,
			workItemId: child.id,
			generation: candidate.generation,
			marker: child.markers.pullRequest,
			title: child.title,
			body: `Refs #${child.issue}\nRefs #${candidate.parentIssue}\n\n${child.markers.pullRequest}`,
			draft: false,
			baseBranch: child.prBase,
			headBranch: child.branch,
			baseSha: handoff.baseHead,
			headSha: handoff.head,
			changedPaths: handoff.changedScope,
			allowedScopes: child.writeScopes,
		};
		transport.pullRequests.push(cleanPullRequest(childRequest));
		const integrated = await orchestrator.integrateChild(candidate, child, handoff);
		assert.equal(integrated.kind, "integrated");
		if (integrated.kind !== "integrated") throw new Error("setup integration failed");
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
			changedPaths: handoff.changedScope,
			allowedScopes: child.writeScopes,
		}, { draft: true }, 900));
		return orchestrator.reconcileParentReadiness(candidate, [integrated.receipt], decisionPolicy);
	};

	for (const proof of [
		async (query: GitAncestryQuery) => ({ ...query, schemaVersion: 1, authority: "transport", result: "true", revision: 1, observedAt: "2026-07-21T12:05:00.000Z" }),
		async (query: GitAncestryQuery) => ({ ...query, repository: "other/cli", schemaVersion: 1, authority: "transport", result: true, revision: 1, observedAt: "2026-07-21T12:05:00.000Z" }),
		async (query: GitAncestryQuery) => ({ ...query, descendantSha: "f".repeat(40), schemaVersion: 1, authority: "transport", result: true, revision: 1, observedAt: "2026-07-21T12:05:00.000Z" }),
	]) {
		const decision = await setup(proof);
		assert.equal(decision.kind, "blocked");
	}
});

test("cycle 4 separates stable integrated PR identity from volatile observation evidence", async () => {
	const candidate = await cycle3Plan(true);
	const issue = issueFrom({
		repository: candidate.repository,
		parentIssue: candidate.parentIssue,
		marker: candidate.children[0].markers.issue,
		title: candidate.children[0].title,
		body: candidate.children[0].issueBody,
	}, 811);
	const child = materializeChildRecord(candidate, "evidence", issue);
	const handoff = childHandoff(issue.number, child.branch, child.prBase);
	const initial = cleanPullRequest(childPullRequestRequest(candidate, child, handoff), { state: "merged" });
	const refreshed = { ...initial, revision: initial.revision + 9, observedAt: "2026-07-21T12:20:00.000Z" };
	const initialSnapshot = createCanonicalPullRequestSnapshot(initial);
	const refreshedSnapshot = createCanonicalPullRequestSnapshot(refreshed);
	assert.equal(initialSnapshot.digest, refreshedSnapshot.digest);
	assert.notEqual(initialSnapshot.revision, refreshedSnapshot.revision);
	assert.notEqual(initialSnapshot.observedAt, refreshedSnapshot.observedAt);
});

test("cycle 4 rejects coherently re-digested wrong child topology before parent readiness", async () => {
	for (const [name, changes] of [
		["head branch", { headBranch: "attacker/topic" }],
		["base branch", { baseBranch: "main" }],
		["marker", { marker: "<!-- shepherd-child-pr:v1:471:evidence:ffffffffffffffffffffffff -->" }],
		["out-of-scope path", { changedPaths: ["outside/owned-scope.ts"] }],
	] as const) {
		const candidate = await cycle3Plan(true);
		const transport = new FakeTransport();
		const receipts = seedIntegrationRoster(candidate, transport);
		const child = candidate.children[0];
		const index = transport.pullRequests.findIndex((pullRequest) => pullRequest.marker === child.markers.pullRequest);
		const current = { ...transport.pullRequests[index], ...changes } as GitHubPullRequestEvidence;
		if ("marker" in changes) current.body = current.body.replace(child.markers.pullRequest, current.marker);
		transport.pullRequests[index] = current;
		const snapshot = createCanonicalPullRequestSnapshot(current);
		const receipt = { ...receipts[0], pullRequestSnapshot: snapshot };
		const mutation = createDurableMutationIntent("child_integration", [candidate.repository, child.markers.pullRequest], {
			repository: candidate.repository,
			childId: child.id,
			pullRequest: receipt.pullRequest,
			generation: candidate.generation,
			marker: child.markers.pullRequest,
			baseSha: snapshot.baseSha,
			headSha: snapshot.headSha,
			parentBranch: candidate.parentBranch,
			pullRequestSnapshot: snapshot,
			observation: receipt.observation,
			controllerProvenance: receipt.controllerProvenance,
		}, null);
		receipt.transportProvenance = {
			authority: "transport",
			idempotencyKey: mutation.idempotencyKey,
			intentDigest: mutation.intentDigest,
			revision: receipt.transportProvenance.revision,
		};
		transport.integrations.splice(0, transport.integrations.length, receipt);
		addParentPullRequest(candidate, transport);
		const broker = new FakeDecisionBroker();
		const result = await orchestratorFor(transport, broker).reconcileParentReadiness(candidate, [receipt], decisionPolicy);
		assert.equal(result.kind, "blocked", name);
		assert.equal(broker.requests.length, 0, name);
		assert.equal(transport.markReadyCalls, 0, name);
	}
});

test("cycle 4 re-reads one complete current required-check policy before integration and readiness", async () => {
	const candidate = await cycle3Plan(true);
	const variants: Array<[string, (query: { repository: string; baseBranch: string }) => unknown]> = [
		["moved", (query) => {
			const policy = cycle3CheckPolicy(query.baseBranch, { revision: 8 });
			return { items: [{ schemaVersion: 1, authority: "controller", repository: query.repository, baseBranch: query.baseBranch, revision: 8, digest: policy.digest, observedAt: "2026-07-21T12:06:00.000Z" }], complete: true };
		}],
		["incomplete", (query) => {
			const policy = cycle3CheckPolicy(query.baseBranch);
			return { items: [{ schemaVersion: 1, authority: "controller", repository: query.repository, baseBranch: query.baseBranch, revision: 7, digest: policy.digest, observedAt: "2026-07-21T12:06:00.000Z" }], complete: false };
		}],
		["wrong repository", (query) => {
			const policy = cycle3CheckPolicy(query.baseBranch);
			return { items: [{ schemaVersion: 1, authority: "controller", repository: "other/cli", baseBranch: query.baseBranch, revision: 7, digest: policy.digest, observedAt: "2026-07-21T12:06:00.000Z" }], complete: true };
		}],
		["stale", (query) => {
			const policy = cycle3CheckPolicy(query.baseBranch);
			return { items: [{ schemaVersion: 1, authority: "controller", repository: query.repository, baseBranch: query.baseBranch, revision: 7, digest: policy.digest, observedAt: "2026-07-21T11:00:00.000Z" }], complete: true };
		}],
	];
	for (const [name, response] of variants) {
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
		const handoff = childHandoff(issue.number, child.branch, child.prBase);
		transport.pullRequests.push(cleanPullRequest(childPullRequestRequest(candidate, child, handoff)));
		const source = {
			async findRequiredCheckPolicies(query: { repository: string; baseBranch: string }, context?: TestCallContext) {
				transport.trackContext("findRequiredCheckPolicies", context);
				return response(query);
			},
		};
		const result = await orchestratorFor(transport, undefined, source).integrateChild(candidate, child, handoff);
		assert.equal(result.kind, "blocked", name);
		assert.equal(transport.integrateCalls, 0, name);
	}

	const readinessTransport = new FakeTransport();
	const receipts = seedIntegrationRoster(candidate, readinessTransport);
	addParentPullRequest(candidate, readinessTransport);
	const broker = new FakeDecisionBroker();
	const movedSource = {
		async findRequiredCheckPolicies(query: { repository: string; baseBranch: string }, context?: TestCallContext) {
			readinessTransport.trackContext("findRequiredCheckPolicies", context);
			const policy = cycle3CheckPolicy(query.baseBranch, { revision: 8 });
			return { items: [{ schemaVersion: 1, authority: "controller", repository: query.repository, baseBranch: query.baseBranch, revision: 8, digest: policy.digest, observedAt: "2026-07-21T12:06:00.000Z" }], complete: true };
		},
	};
	const readiness = await orchestratorFor(readinessTransport, broker, movedSource)
		.reconcileParentReadiness(candidate, receipts, decisionPolicy);
	assert.equal(readiness.kind, "blocked");
	assert.equal(broker.requests.length, 0);
});

function hangUntilAbort(context?: TestCallContext): Promise<never> {
	return new Promise((_, reject) => {
		if (context?.signal === undefined) return;
		if (context.signal.aborted) {
			reject(new Error("external operation aborted"));
			return;
		}
		context.signal.addEventListener("abort", () => reject(new Error("external operation aborted")), { once: true });
	});
}

async function settleWithin<T>(promise: Promise<T>, milliseconds = 100): Promise<T | "hung"> {
	return Promise.race([
		promise,
		new Promise<"hung">((resolve) => setTimeout(() => resolve("hung"), milliseconds)),
	]);
}

test("cycle 4 bounds and cancels never-settling calls, drains keyed work, and reconciles late effects", async () => {
	const candidate = await cycle3Plan(true);
	const transport = new FakeTransport();
	const normalLookup = transport.findChildIssues.bind(transport);
	transport.findChildIssues = (async (_query: unknown, context?: TestCallContext) => hangUntilAbort(context)) as never;
	const orchestrator = orchestratorFor(transport, undefined, defaultPolicySource(transport), 15);
	const first = orchestrator.ensureChildIssue(candidate, "evidence").then(
		() => ({ kind: "success" as const }),
		(error) => ({ kind: "error" as const, error }),
	);
	const firstOutcome = await settleWithin(first);
	transport.findChildIssues = normalLookup as never;
	const secondOutcome = await settleWithin(orchestrator.ensureChildIssue(candidate, "evidence"));
	assert.notEqual(firstOutcome, "hung");
	assert.notEqual(secondOutcome, "hung");
	if (firstOutcome !== "hung" && firstOutcome.kind === "error") {
		assert.ok(firstOutcome.error instanceof Error);
		assert.equal((firstOutcome.error as any).code, "external_timeout");
		assert.equal((firstOutcome.error as any).uncertain, true);
	}

	const lateTransport = new FakeTransport();
	lateTransport.createChildIssue = (async (request: CreateChildIssueRequest, context?: TestCallContext) => {
		lateTransport.trackContext("createChildIssue", context);
		return new Promise((_, reject) => {
			context?.signal?.addEventListener("abort", () => {
				lateTransport.createIssueCalls += 1;
				lateTransport.issues.push(issueFrom(request, 811));
				reject(new Error("Authorization: Bearer LATE_EFFECT_MARKER"));
			}, { once: true });
		});
	}) as never;
	const recovered = await settleWithin(orchestratorFor(lateTransport, undefined, defaultPolicySource(lateTransport), 15)
		.ensureChildIssue(candidate, "evidence"));
	assert.notEqual(recovered, "hung");
	assert.equal(lateTransport.createIssueCalls, 1);
});

test("cycle 4 rejects sensitive valid-field text and normalizes every external rejection shape", async () => {
	const source = JSON.parse(await readFile(objectivePath, "utf8")) as Record<string, unknown>;
	const marker = "CYCLE4_SENSITIVE_MARKER";
	const children = source.children as Array<Record<string, unknown>>;
	const variants = [
		{ ...source, title: `Authorization: Bearer ${marker}` },
		{ ...source, objective: `token=${marker}` },
		{ ...source, children: [{ ...children[0], title: `api_key=${marker}` }] },
		{ ...source, children: [{ ...children[0], objective: `https://user:${marker}@example.invalid/path` }] },
		{ ...source, children: [{ ...children[0], verification: [{ id: "focused", kind: "test", description: `password=${marker}` }] }] },
	];
	for (const variant of variants) {
		let rejection: unknown;
		try {
			createPlanFromSource(variant);
		} catch (error) {
			rejection = error;
		}
		assert.ok(rejection instanceof Error);
		assert.doesNotMatch(String(rejection), new RegExp(marker, "u"));
		assert.match(String(rejection), /credential|secret|sensitive|plan/i);
	}

	for (const reason of [new Error(`token=${marker}`), `token=${marker}`, { message: `token=${marker}` }, undefined]) {
		const transport = new FakeTransport();
		transport.findChildIssues = (async () => Promise.reject(reason)) as never;
		let rejection: unknown;
		try {
			await orchestratorFor(transport).ensureChildIssue(await cycle3Plan(true), "evidence");
		} catch (error) {
			rejection = error;
		}
		assert.ok(rejection instanceof Error);
		assert.equal((rejection as any).code, "external_port_failed");
		assert.doesNotMatch(String(rejection), new RegExp(marker, "u"));
		assert.ok(String(rejection).length <= 256);
	}
});

test("cycle 4 rejects sensitive decision questions before broker publication", async () => {
	const candidate = await cycle3Plan(true);
	const transport = new FakeTransport();
	const receipts = seedIntegrationRoster(candidate, transport);
	addParentPullRequest(candidate, transport);
	const broker = new FakeDecisionBroker();
	let result: Awaited<ReturnType<GitHubParentOrchestrator["reconcileParentReadiness"]>> | undefined;
	let rejection: unknown;
	try {
		result = await orchestratorFor(transport, broker).reconcileParentReadiness(candidate, receipts, {
			...decisionPolicy,
			question: "Authorization: Bearer CYCLE4_DECISION_MARKER",
		});
	} catch (error) {
		rejection = error;
	}
	if (result !== undefined) assert.notEqual(result.kind, "ready");
	assert.equal(broker.requests.length, 0);
	if (rejection !== undefined) assert.doesNotMatch(String(rejection), /CYCLE4_DECISION_MARKER/u);
});

test("cycle 4 requires authoritative CAS resource revisions to advance", async () => {
	const candidate = await cycle3Plan(true);
	const rosterTransport = new FakeTransport();
	const orchestrator = orchestratorFor(rosterTransport);
	await (orchestrator.reconcileParentRoster as any)(candidate, { evidence: "running" }, 1);
	rosterTransport.rosterRevisionDelta = 0;
	await assert.rejects(
		(orchestrator.reconcileParentRoster as any)(candidate, { evidence: "succeeded" }, 2),
		/CAS|revision|advance|stale|conditional/i,
	);

	const readyTransport = new FakeTransport();
	const receipts = seedIntegrationRoster(candidate, readyTransport);
	addParentPullRequest(candidate, readyTransport);
	readyTransport.readyRevisionDelta = 0;
	const ready = await orchestratorFor(readyTransport, new FakeDecisionBroker())
		.reconcileParentReadiness(candidate, receipts, decisionPolicy);
	assert.notEqual(ready.kind, "ready");
});

test("cycle 4 compares exact dense plan array lengths before generic descriptor traversal", async () => {
	const candidate = await cycle3Plan();
	const sparse = JSON.parse(JSON.stringify(candidate)) as ParentOrchestrationPlan;
	sparse.children.length = 1_000_000;
	assert.throws(
		() => selectReadyChildren(sparse, { evidence: "pending", orchestrator: "pending" }, 1),
		/bounded|dense|array|canonical|plan/i,
	);

	const source = JSON.parse(await readFile(objectivePath, "utf8")) as Record<string, unknown>;
	const dense = Array.from({ length: 65 }, (_, index) => ({ ...(source.children as Array<Record<string, unknown>>)[0], id: `child-${index}` }));
	const original = Object.getOwnPropertyDescriptors;
	let traversed = false;
	let rejection: unknown;
	Object.getOwnPropertyDescriptors = ((value: object) => {
		if (value === dense) {
			traversed = true;
			throw new Error("descriptor traversal must not occur");
		}
		return original(value);
	}) as typeof Object.getOwnPropertyDescriptors;
	try {
		createPlanFromSource({ ...source, children: dense });
	} catch (error) {
		rejection = error;
	} finally {
		Object.getOwnPropertyDescriptors = original;
	}
	assert.equal(traversed, false);
	assert.match(String(rejection), /bounded|children|64/i);
});

test("cycle 4 passes AbortSignal to workspace handoff and redacts workspace failures", async () => {
	const candidate = await cycle3Plan(true);
	const issue = issueFrom({
		repository: candidate.repository,
		parentIssue: candidate.parentIssue,
		marker: candidate.children[0].markers.issue,
		title: candidate.children[0].title,
		body: candidate.children[0].issueBody,
	}, 811);
	const child = materializeChildRecord(candidate, "evidence", issue);
	const transport = new FakeTransport();
	transport.issues.push(issue);
	let signal = false;
	let rejection: unknown;
	try {
		await orchestratorFor(transport).captureChildHandoff(candidate, child, {} as ClaimedWorkspace, {
			async captureHandoff(_workspace: ClaimedWorkspace, _state: "passed", context?: TestCallContext): Promise<WorkspaceHandoffEvidence> {
				signal = context?.signal instanceof AbortSignal;
				throw new Error("token=CYCLE4_WORKSPACE_MARKER");
			},
		} as never);
	} catch (error) {
		rejection = error;
	}
	assert.equal(signal, true);
	assert.ok(rejection instanceof Error);
	assert.equal((rejection as any).code, "external_port_failed");
	assert.doesNotMatch(String(rejection), /CYCLE4_WORKSPACE_MARKER/u);
});

async function cycle5ReadinessScenario() {
	const candidate = await cycle3Plan(true);
	const transport = new FakeTransport();
	const receipts = seedIntegrationRoster(candidate, transport);
	addParentPullRequest(candidate, transport);
	return { candidate, transport, receipts };
}

function handoffForReceipt(
	candidate: ParentOrchestrationPlan,
	transport: FakeTransport,
	receipt: ChildIntegrationReceipt,
): { child: ReturnType<typeof materializeChildRecord>; handoff: WorkspaceHandoffEvidence } {
	const issue = transport.issues.find((entry) => entry.number === receipt.pullRequest - 1);
	assert.ok(issue);
	const child = materializeChildRecord(candidate, receipt.childId, issue);
	return {
		child,
		handoff: {
			issue: issue.number,
			branch: child.branch,
			prBase: child.prBase,
			baseHead: receipt.baseSha,
			head: receipt.headSha,
			changedScope: [...receipt.pullRequestSnapshot.changedPaths],
			verificationState: "passed",
			repositoryIdentity: "1".repeat(64),
			worktreeIdentity: "2".repeat(64),
			dirty: false,
		},
	};
}

async function assertReadinessDoesNotMutate(
	transport: FakeTransport,
	operation: () => Promise<unknown>,
): Promise<void> {
	let result: unknown;
	try {
		result = await operation();
	} catch {
		// A typed/normalized rejection is an acceptable fail-closed result at an untrusted boundary.
	}
	assert.notEqual((result as { kind?: string } | undefined)?.kind, "ready");
	assert.equal(transport.markReadyCalls, 0);
}

test("cycle 5 runtime-validates exact broker request, poll, and consume records before ready", async (t) => {
	const recordMutations: Array<[string, (record: Record<string, unknown>) => void]> = [
		["extra field", (record) => { record.extra = true; }],
		["wrong request ID", (record) => { record.requestId = "other-request"; }],
		["wrong gate", (record) => { record.gate = "merge"; }],
		["forged marker", (record) => { record.idempotencyMarker = "<!-- forged -->"; }],
		["wrong options", (record) => { record.allowedOptions = ["reject", "approve-merge"]; }],
		["wrong allowlist", (record) => { record.actorAllowlist = ["other-maintainer"]; }],
		["wrong expiry", (record) => { record.expiresAt = "2028-07-21T12:00:00.000Z"; }],
		["wrong question", (record) => { record.question = "Different approval question"; }],
		["incoherent consumed status", (record) => { record.status = "consumed"; }],
	];
	for (const [name, mutate] of recordMutations) {
		await t.test(`request ${name}`, async () => {
			const { candidate, transport, receipts } = await cycle5ReadinessScenario();
			const broker = new FakeDecisionBroker();
			const request = broker.request.bind(broker);
			broker.request = (async (value: GitHubDecisionRequest, context?: TestCallContext) => {
				const record = structuredClone(await request(value, context)) as unknown as Record<string, unknown>;
				mutate(record);
				return record as unknown as HumanDecisionRecord;
			}) as never;
			await assertReadinessDoesNotMutate(transport, () => orchestratorFor(transport, broker)
				.reconcileParentReadiness(candidate, receipts, decisionPolicy));
		});
	}

	await t.test("poll record cannot forge immutable request fields", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		const broker = new FakeDecisionBroker();
		broker.poll = (async () => ({
			...createHumanDecisionRecord({
				requestId: "wrong-polled-request",
				gate: "parent_merge",
				binding: {
					repository: candidate.repository,
					target: { kind: "pull_request", number: 900 },
					generation: candidate.generation,
					headSha: "e".repeat(40),
				},
				allowedOptions: ["approve-merge", "reject"],
				actorAllowlist: ["maintainer"],
				expiresAt: decisionPolicy.expiresAt,
				question: decisionPolicy.question,
			}, new Date("2026-07-21T12:00:00.000Z")),
			status: "decided",
			decision: approvedDecision,
		})) as never;
		await assertReadinessDoesNotMutate(transport, () => orchestratorFor(transport, broker)
			.reconcileParentReadiness(candidate, receipts, decisionPolicy));
	});

	await t.test("poll and consume decision evidence requires actor, source, and timestamp", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		const broker = new FakeDecisionBroker();
		broker.pollResult = { status: "decided", decision: { option: "approve-merge" } as never, attempts: 1 };
		broker.consume = (async () => ({ option: "approve-merge" })) as never;
		await assertReadinessDoesNotMutate(transport, () => orchestratorFor(transport, broker)
			.reconcileParentReadiness(candidate, receipts, decisionPolicy));
	});
});

test("cycle 5 refreshes every plan-bound policy for receipt reuse and each readiness stage", async (t) => {
	await t.test("moved child-base policy blocks receipt reuse", async () => {
		const candidate = await cycle3Plan(true);
		const transport = new FakeTransport();
		const receipts = seedIntegrationRoster(candidate, transport);
		const { child, handoff } = handoffForReceipt(candidate, transport, receipts[0]);
		const queries: string[] = [];
		const source = {
			async findRequiredCheckPolicies(query: { repository: string; baseBranch: string }): Promise<unknown> {
				queries.push(query.baseBranch);
				const policy = cycle3CheckPolicy(query.baseBranch, query.baseBranch === candidate.parentBranch ? { revision: 8 } : {});
				return { items: [{
					schemaVersion: 1, authority: "controller", repository: query.repository, baseBranch: query.baseBranch,
					revision: policy.revision, digest: policy.digest, observedAt: "2026-07-21T12:06:00.000Z",
				}], complete: true };
			},
		};
		const result = await orchestratorFor(transport, undefined, source).integrateChild(candidate, child, handoff);
		assert.equal(result.kind, "blocked");
		assert.deepEqual(new Set(queries), new Set(candidate.requiredCheckPolicies.map((policy) => policy.baseBranch)));
	});

	await t.test("moved child-base policy blocks readiness before broker", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		const broker = new FakeDecisionBroker();
		const queries: string[] = [];
		const source = {
			async findRequiredCheckPolicies(query: { repository: string; baseBranch: string }): Promise<unknown> {
				queries.push(query.baseBranch);
				const policy = cycle3CheckPolicy(query.baseBranch, query.baseBranch === candidate.parentBranch ? { revision: 8 } : {});
				return { items: [{
					schemaVersion: 1, authority: "controller", repository: query.repository, baseBranch: query.baseBranch,
					revision: policy.revision, digest: policy.digest, observedAt: "2026-07-21T12:06:00.000Z",
				}], complete: true };
			},
		};
		await assertReadinessDoesNotMutate(transport, () => orchestratorFor(transport, broker, source)
			.reconcileParentReadiness(candidate, receipts, decisionPolicy));
		assert.equal(broker.requests.length, 0);
		assert.deepEqual(new Set(queries), new Set(candidate.requiredCheckPolicies.map((policy) => policy.baseBranch)));
	});

	await t.test("exact readiness queries the complete set before, after, and during recovery", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		const queries: string[] = [];
		const source = {
			async findRequiredCheckPolicies(query: { repository: string; baseBranch: string }): Promise<unknown> {
				queries.push(query.baseBranch);
				return defaultPolicySource(transport).findRequiredCheckPolicies(query);
			},
		};
		const result = await orchestratorFor(transport, new FakeDecisionBroker(), source)
			.reconcileParentReadiness(candidate, receipts, decisionPolicy);
		assert.equal(result.kind, "ready");
		for (const policy of candidate.requiredCheckPolicies) {
			assert.ok(queries.filter((branch) => branch === policy.baseBranch).length >= 3, policy.baseBranch);
		}
	});
});

test("cycle 5 exposes an authoritative full-policy async plan boundary for a port-only controller", async () => {
	const objective = JSON.parse(await readFile(objectivePath, "utf8")) as Record<string, unknown>;
	const transport = new FakeTransport();
	const queries: unknown[] = [];
	const policySource = {
		async findParentOrchestrationPolicyBundle(query: Record<string, unknown>, context?: TestCallContext): Promise<unknown> {
			queries.push(structuredClone(query));
			transport.trackContext("findParentOrchestrationPolicyBundle", context);
			return {
				items: [{
					schemaVersion: 1,
					authority: "controller",
					repository: objective.repository,
					parentIssue: objective.parentIssue,
					generation: objective.generation,
					parentBranch: objective.parentBranch,
					parentBaseBranch: objective.parentBaseBranch,
					revision: 7,
					observedAt: "2026-07-21T12:06:00.000Z",
					policyBundle: {
						schemaVersion: 1,
						requiredCheckPolicies: [
							cycle3CheckPolicy(String(objective.parentBranch)),
							cycle3CheckPolicy(String(objective.parentBaseBranch)),
						],
					},
				}],
				complete: true,
			};
		},
		...defaultPolicySource(transport),
	};
	const orchestrator = orchestratorFor(transport, undefined, policySource);
	const createPlan = (orchestrator as unknown as Record<string, unknown>).createPlan;
	assert.equal(typeof createPlan, "function");
	const created = await (createPlan as (this: GitHubParentOrchestrator, value: unknown, context?: unknown) => Promise<ParentOrchestrationPlan>)
		.call(orchestrator, objective, {});
	assert.equal(created.canonical.digest.length, 64);
	assert.deepEqual(created.requiredCheckPolicies.map((policy) => policy.baseBranch).sort(), [
		String(objective.parentBaseBranch), String(objective.parentBranch),
	].sort());
	assert.equal(queries.length, 1);
	assert.equal(transport.createIssueCalls + transport.createPullRequestCalls + transport.publishRosterCalls, 0);
});

test("cycle 5 independently reauthorizes current child evidence instead of trusting receipt claims", async (t) => {
	await t.test("missing current child review and attestation blocks readiness", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		for (let index = 0; index < transport.pullRequests.length; index += 1) {
			if (transport.pullRequests[index].number !== 900) {
				transport.pullRequests[index] = { ...transport.pullRequests[index], reviews: [] };
			}
		}
		await assertReadinessDoesNotMutate(transport, () => orchestratorFor(transport, new FakeDecisionBroker())
			.reconcileParentReadiness(candidate, receipts, decisionPolicy));
	});

	await t.test("forged controller revision/time plus recomputed transport digest still blocks", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		const forged = structuredClone(receipts);
		forged[0].controllerProvenance.evidenceRevision = 1;
		forged[0].controllerProvenance.observedAt = "2026-07-21T10:00:00.000Z";
		const child = candidate.children[0];
		const mutation = createDurableMutationIntent(
			"child_integration",
			[candidate.repository, child.markers.pullRequest],
			{
				repository: candidate.repository,
				childId: child.id,
				pullRequest: forged[0].pullRequest,
				generation: candidate.generation,
				marker: child.markers.pullRequest,
				baseSha: forged[0].baseSha,
				headSha: forged[0].headSha,
				parentBranch: candidate.parentBranch,
				pullRequestSnapshot: forged[0].pullRequestSnapshot,
				observation: forged[0].observation,
				controllerProvenance: forged[0].controllerProvenance,
			},
			null,
		);
		forged[0].transportProvenance = {
			...forged[0].transportProvenance,
			idempotencyKey: mutation.idempotencyKey,
			intentDigest: mutation.intentDigest,
		};
		transport.integrations.splice(0, transport.integrations.length, ...forged);
		await assertReadinessDoesNotMutate(transport, () => orchestratorFor(transport, new FakeDecisionBroker())
			.reconcileParentReadiness(candidate, forged, decisionPolicy));
	});
});

test("cycle 5 centralizes current child PR eligibility for reuse and readiness", async (t) => {
	for (const state of ["open", "merged"] as const) {
		await t.test(`draft ${state} child`, async () => {
			const { candidate, transport, receipts } = await cycle5ReadinessScenario();
			for (let index = 0; index < transport.pullRequests.length; index += 1) {
				if (transport.pullRequests[index].number !== 900) {
					transport.pullRequests[index] = { ...transport.pullRequests[index], state, draft: true };
				}
			}
			await assertReadinessDoesNotMutate(transport, () => orchestratorFor(transport, new FakeDecisionBroker())
				.reconcileParentReadiness(candidate, receipts, decisionPolicy));
		});
	}
});

test("cycle 5 binds CAS preconditions while keeping volatile child observations outside mutation identity", async () => {
	const first = createDurableMutationIntent("parent_ready", ["polymetrics-ai/cli", "ready"], { logical: "ready" }, 41);
	const second = createDurableMutationIntent("parent_ready", ["polymetrics-ai/cli", "ready"], { logical: "ready" }, 42);
	assert.notEqual(first.idempotencyKey, second.idempotencyKey);
	assert.notEqual(first.intentDigest, second.intentDigest);

	const mutations: Array<{ idempotencyKey: string; intentDigest: string }> = [];
	for (const [revision, observedAt] of [[42, "2026-07-21T12:05:00.000Z"], [43, "2026-07-21T12:06:00.000Z"]] as const) {
		const candidate = await cycle3Plan(true);
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
		const handoff = childHandoff(issue.number, child.branch, child.prBase);
		transport.pullRequests.push(cleanPullRequest(childPullRequestRequest(candidate, child, handoff), {
			revision,
			observedAt,
		}, 812));
		transport.integrateChild = (async (request: IntegrateChildRequest) => {
			mutations.push(request.mutation);
			throw new Error("synthetic integration boundary failure");
		}) as never;
		await assert.rejects(orchestratorFor(transport).integrateChild(candidate, child, handoff));
	}
	assert.equal(mutations.length, 2);
	assert.equal(mutations[0].idempotencyKey, mutations[1].idempotencyKey);
	assert.equal(mutations[0].intentDigest, mutations[1].intentDigest);
});

test("cycle 5 rejects cookie/session values across plan, outbound bodies, and decision questions", async (t) => {
	const source = JSON.parse(await readFile(objectivePath, "utf8")) as Record<string, unknown>;
	const child = (source.children as Array<Record<string, unknown>>)[0];
	const marker = "SYNTHETIC_CYCLE5_SESSION";
	const variants: Array<[string, Record<string, unknown>]> = [
		["parent title", { ...source, title: `Set-Cookie: session=${marker}; HttpOnly` }],
		["parent objective", { ...source, objective: `Cookie: sid=${marker}` }],
		["child title", { ...source, children: [{ ...child, title: `X-Session-Token: ${marker}` }] }],
		["child objective/issue body", { ...source, children: [{ ...child, objective: `session cookie=${marker}` }] }],
		["verification description", {
			...source,
			children: [{ ...child, verification: [{ id: "focused", kind: "test", description: `X-CSRF-Token: ${marker}` }] }],
		}],
	];
	for (const [name, value] of variants) {
		await t.test(name, () => assert.throws(() => createPlanFromSource(value), /credential|secret|sensitive/i));
	}

	await t.test("decision question", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		await assertReadinessDoesNotMutate(transport, () => orchestratorFor(transport, new FakeDecisionBroker())
			.reconcileParentReadiness(candidate, receipts, {
				...decisionPolicy,
				question: `Set-Cookie: session=${marker}; HttpOnly`,
			}));
	});
});

test("cycle 5 links caller lifecycle and retains keyed ownership until live port settlement", async () => {
	const candidate = await cycle3Plan(true);
	const preAbortedTransport = new FakeTransport();
	const preAborted = new AbortController();
	preAborted.abort();
	await assert.rejects(
		(orchestratorFor(preAbortedTransport).ensureChildIssue as any)(candidate, "evidence", { signal: preAborted.signal }),
		/cancel|abort|deadline|timeout|external/i,
	);
	assert.equal(preAbortedTransport.createIssueCalls, 0);
	assert.equal(preAbortedTransport.callContexts.length, 0);

	const deadlineTransport = new FakeTransport();
	let observedDeadline = "";
	deadlineTransport.findChildIssues = (async (_query: unknown, context?: TestCallContext & { deadlineAt?: string }) => {
		observedDeadline = context?.deadlineAt ?? "";
		return { items: [...deadlineTransport.issues], complete: true };
	}) as never;
	const callerDeadline = new Date(Date.now() + 5).toISOString();
	await (orchestratorFor(deadlineTransport, undefined, defaultPolicySource(deadlineTransport), 50).ensureChildIssue as any)(
		candidate,
		"evidence",
		{ deadlineAt: callerDeadline },
	);
	assert.ok(observedDeadline <= callerDeadline);

	const transport = new FakeTransport();
	let starts = 0;
	transport.findChildIssues = (async (_query: unknown, context?: TestCallContext & { acknowledgeAbort?: () => void }) => {
		starts += 1;
		context?.signal?.addEventListener("abort", () => context.acknowledgeAbort?.(), { once: true });
		return new Promise(() => {});
	}) as never;
	const orchestrator = orchestratorFor(transport, undefined, defaultPolicySource(transport), 10);
	const first = await settleWithin((orchestrator.ensureChildIssue as any)(candidate, "evidence")
		.then(() => "resolved", () => "rejected"), 100);
	assert.equal(first, "rejected");
	const secondDeadline = new Date(Date.now() + 25).toISOString();
	const second = await settleWithin((orchestrator.ensureChildIssue as any)(candidate, "evidence", { deadlineAt: secondDeadline })
		.then(() => "resolved", () => "rejected"), 100);
	assert.equal(second, "rejected");
	assert.equal(starts, 1, "same key must remain quarantined while the first port invocation is live");
	const stop = (orchestrator as unknown as Record<string, unknown>).stop;
	assert.equal(typeof stop, "function");
	const stopped = await (stop as (this: GitHubParentOrchestrator, context?: unknown) => Promise<Record<string, unknown>>)
		.call(orchestrator, { deadlineAt: new Date(Date.now() + 20).toISOString() });
	assert.equal(stopped.kind, "incomplete");
	assert.equal(stopped.active, 1);
});

test("cycle 5 durable run state names current review truth and exact available checkpoints", async () => {
	const raw = await readFile(".planning/phases/478-shepherd-github-parent-orchestration/RUN-STATE.json", "utf8");
	const state = JSON.parse(raw) as any;
	const historical = state.details.reviewHistory.find((entry: { cycle?: number }) => entry.cycle === 5);
	assert.equal(historical.priorReviewedCandidate, "63ac436fdac5fc46be7004f8109c4f068aa5749c");
	assert.equal(historical.review1, "blocked");
	assert.equal(historical.review2, "blocked");
	assert.equal(state.details.checkpoints.cycle5Plan, "7cf9c88ddadee395020444c19ee9f001b0807a53");
	assert.equal(state.details.checkpoints.cycle5Red, "6cb21902244e4bccf390c4e7556eb615e5e1697f");
	assert.equal(state.details.checkpoints.cycle5Evidence, "63ac436fdac5fc46be7004f8109c4f068aa5749c");
	assert.doesNotMatch(raw, /3f285722a505ea426d53a34f95716781d1aca7c2/u);
});

test("cycle 6 composes the actual broker through its owned canonical record boundary", async () => {
	const { candidate, transport, receipts } = await cycle5ReadinessScenario();
	const repository = new MemoryHumanDecisionRepository();
	const comments: GitHubComment[] = [];
	let now = new Date("2026-07-21T12:00:00.000Z");
	const githubTransport = {
		async getAuthenticatedActor(): Promise<string> {
			return "shepherd-host";
		},
		async listComments(): Promise<GitHubComment[]> {
			return structuredClone(comments);
		},
		async createDecisionRequestComment(record: HumanDecisionRecord): Promise<GitHubComment> {
			const created: GitHubComment = {
				id: 1,
				url: "https://github.com/polymetrics-ai/cli/pull/900#issuecomment-1",
				body: renderDecisionRequestComment(record),
				actor: { login: "shepherd-host", type: "User" },
				createdAt: "2026-07-21T12:00:00.000Z",
				updatedAt: "2026-07-21T12:00:00.000Z",
			};
			comments.push(created);
			return structuredClone(created);
		},
	};
	const actual = new GitHubDecisionBroker(repository, githubTransport, {
		now: () => now,
		sleep: async () => {},
		polling: { maxAttempts: 1, initialDelayMs: 1, maxDelayMs: 1 },
		transportRetry: { maxAttempts: 1, initialDelayMs: 1, maxDelayMs: 1 },
	});
	const adapt: Cycle6BrokerAdapterFactory = githubOrchestratorApi.adaptGitHubDecisionBroker;
	assert.equal(typeof adapt, "function");
	const broker = adapt(actual);
	const orchestrator = orchestratorFor(transport, broker);

	const pending = await orchestrator.reconcileParentReadiness(candidate, receipts, decisionPolicy);
	assert.deepEqual(pending, { kind: "awaiting_human", reason: "pending" });
	assert.equal(comments.length, 1);
	assert.equal((await repository.load(decisionPolicy.requestId))?.status, "pending");

	comments.push({
		id: 2,
		url: "https://github.com/polymetrics-ai/cli/pull/900#issuecomment-2",
		body: `/shepherd decide ${decisionPolicy.requestId} approve-merge`,
		actor: { login: "maintainer", type: "User" },
		createdAt: "2026-07-21T12:00:30.000Z",
		updatedAt: "2026-07-21T12:00:30.000Z",
	});
	now = new Date("2026-07-21T12:00:40.000Z");
	const ready = await orchestrator.reconcileParentReadiness(candidate, receipts, decisionPolicy);
	assert.equal(ready.kind, "ready");
	assert.equal(transport.markReadyCalls, 1);
	const consumed = await repository.load(decisionPolicy.requestId);
	assert.equal(consumed?.status, "consumed");
	assert.equal(consumed?.decision?.sourceUrl, comments[1].url);
	assert.equal(consumed?.requestComment?.id, comments[0].id);
});

test("cycle 6 fails closed on incomplete chronology and hostile foreign broker DTOs", async (t) => {
	await t.test("a consumed record without persisted request-comment provenance cannot authorize ready", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		const broker = new FakeDecisionBroker();
		const consume = broker.consume.bind(broker);
		broker.consume = (async (requestId: string, binding: HumanDecisionRecord["binding"], context?: TestCallContext) => {
			const record = await consume(requestId, binding, context);
			const { requestComment: _requestComment, ...withoutRequestComment } = record;
			return withoutRequestComment as HumanDecisionRecord;
		}) as never;
		await assertReadinessDoesNotMutate(transport, () => orchestratorFor(transport, broker)
			.reconcileParentReadiness(candidate, receipts, decisionPolicy));
	});

	await t.test("updatedAt cannot precede decision or consumption", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		const broker = new FakeDecisionBroker();
		const consume = broker.consume.bind(broker);
		broker.consume = (async (requestId: string, binding: HumanDecisionRecord["binding"], context?: TestCallContext) => ({
			...await consume(requestId, binding, context),
			updatedAt: "2026-07-21T12:00:20.000Z",
		})) as never;
		await assertReadinessDoesNotMutate(transport, () => orchestratorFor(transport, broker)
			.reconcileParentReadiness(candidate, receipts, decisionPolicy));
	});

	await t.test("wide records reject before an accessor is invoked and errors are normalized", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		const broker = new FakeDecisionBroker();
		let accessed = false;
		broker.request = (async () => {
			const wide: Record<string, unknown> = {};
			Object.defineProperty(wide, "schemaVersion", {
				enumerable: true,
				get() {
					accessed = true;
					throw new Error("SYNTHETIC_CYCLE6_BROKER_ACCESSOR");
				},
			});
			for (let index = 0; index < 300; index += 1) wide[`field${index}`] = index;
			return wide as unknown as HumanDecisionRecord;
		}) as never;
		let rejection: unknown;
		try {
			await orchestratorFor(transport, broker).reconcileParentReadiness(candidate, receipts, decisionPolicy);
		} catch (error) {
			rejection = error;
		}
		assert.ok(rejection instanceof Error);
		assert.equal((rejection as Error & { code?: string }).code, "external_port_failed");
		assert.equal(accessed, false);
		assert.doesNotMatch(String(rejection), /SYNTHETIC_CYCLE6_BROKER_ACCESSOR/u);
	});

	await t.test("normal and revoked broker proxies reject without traps or host error text", async () => {
		for (const mode of ["normal", "revoked"] as const) {
			const { candidate, transport, receipts } = await cycle5ReadinessScenario();
			const broker = new FakeDecisionBroker();
			let trapped = false;
			if (mode === "normal") {
				broker.request = (async () => new Proxy({}, {
					ownKeys() {
						trapped = true;
						throw new Error("SYNTHETIC_CYCLE6_BROKER_PROXY");
					},
				}) as HumanDecisionRecord) as never;
			} else {
				const revoked = Proxy.revocable({}, {});
				revoked.revoke();
				broker.request = (async () => revoked.proxy as HumanDecisionRecord) as never;
			}
			let rejection: unknown;
			try {
				await orchestratorFor(transport, broker).reconcileParentReadiness(candidate, receipts, decisionPolicy);
			} catch (error) {
				rejection = error;
			}
			assert.ok(rejection instanceof Error, mode);
			assert.equal((rejection as Error & { code?: string }).code, "external_port_failed", mode);
			assert.equal(trapped, false, mode);
			assert.doesNotMatch(String(rejection), /SYNTHETIC_CYCLE6_BROKER_PROXY|Cannot perform|revoked/u, mode);
		}
	});
});

test("cycle 6 makes parent-ready one conditional authorization effect with rollback", async (t) => {
	await t.test("request carries a closed current-authorization token", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		let captured: Record<string, unknown> | undefined;
		const mark = transport.markParentReady.bind(transport);
		transport.markParentReady = (async (request: MarkParentReadyRequest, context?: TestCallContext) => {
			captured = request as unknown as Record<string, unknown>;
			return mark(request, context);
		}) as never;
		await orchestratorFor(transport, new FakeDecisionBroker())
			.reconcileParentReadiness(candidate, receipts, decisionPolicy);
		const authorization = captured?.authorization as Record<string, unknown> | undefined;
		assert.equal(authorization?.schemaVersion, 1);
		assert.match(String(authorization?.digest), /^[0-9a-f]{64}$/u);
		assert.match(String(authorization?.decisionDigest), /^[0-9a-f]{64}$/u);
		assert.match(String(authorization?.childRosterDigest), /^[0-9a-f]{64}$/u);
		assert.match(String(authorization?.policySetDigest), /^[0-9a-f]{64}$/u);
	});

	await t.test("authority movement inside the conditional effect leaves the parent draft", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		const authority: ParentReadyDurableAuthorityBoundary = {
			async compareConsumeAndMarkParentReady() {
				return { schemaVersion: 1, kind: "conflict", coordinate: "policy" };
			},
			async quarantineAndRollbackParentReady() {
				throw new Error("rollback must not authorize a pre-effect conflict");
			},
		};
		let result: unknown;
		try {
			result = await orchestratorFor(transport, new FakeDecisionBroker(), defaultPolicySource(transport), 25, authority)
				.reconcileParentReadiness(candidate, receipts, decisionPolicy);
		} catch {
			// A typed conditional-conflict rejection is also fail closed.
		}
		assert.notEqual((result as { kind?: string } | undefined)?.kind, "ready");
		assert.equal(transport.markReadyCalls, 0);
		assert.equal(transport.pullRequests.find((pullRequest) => pullRequest.number === 900)?.draft, true);
	});

	await t.test("after-effect drift invokes one idempotent rollback and verifies draft state", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		let effectApplied = false;
		let rollbackCalls = 0;
		const mark = transport.markParentReady.bind(transport);
		transport.markParentReady = (async (request: MarkParentReadyRequest, context?: TestCallContext) => {
			const result = await mark(request, context);
			effectApplied = true;
			return result;
		}) as never;
		Object.assign(transport, {
			async rollbackParentReady(request: { pullRequest: number; mutation: { idempotencyKey: string; intentDigest: string } }) {
				rollbackCalls += 1;
				const index = transport.pullRequests.findIndex((pullRequest) => pullRequest.number === request.pullRequest);
				assert.ok(index >= 0);
				const updated = { ...transport.pullRequests[index], draft: true, revision: transport.pullRequests[index].revision + 1 };
				transport.pullRequests.splice(index, 1, updated);
				return {
					schemaVersion: 1,
					idempotencyKey: request.mutation.idempotencyKey,
					intentDigest: request.mutation.intentDigest,
					revision: 999,
					applied: true,
					value: updated,
				};
			},
		});
		const baseline = defaultPolicySource(transport);
		const source = {
			async findRequiredCheckPolicies(query: { repository: string; baseBranch: string }, context?: TestCallContext): Promise<unknown> {
				const result = await baseline.findRequiredCheckPolicies(query, context) as {
					items: Array<Record<string, unknown>>;
					complete: boolean;
				};
				if (!effectApplied) return result;
				return {
					...result,
					items: result.items.map((item) => ({ ...item, revision: Number(item.revision) + 1 })),
				};
			},
		};
		await assert.rejects(
			orchestratorFor(transport, new FakeDecisionBroker(), source)
				.reconcileParentReadiness(candidate, receipts, decisionPolicy),
		);
		assert.equal(rollbackCalls, 1);
		assert.equal(transport.pullRequests.find((pullRequest) => pullRequest.number === 900)?.draft, true);
	});
});

test("cycle 6 uses intrinsic signal ownership and truthful abort acknowledgement", async (t) => {
	await t.test("a pre-aborted genuine signal cannot be revived by own shadows", async () => {
		const candidate = await cycle3Plan(true);
		const transport = new FakeTransport();
		const controller = new AbortController();
		controller.abort();
		Object.defineProperties(controller.signal, {
			aborted: { configurable: true, value: false },
			addEventListener: { configurable: true, value: () => {} },
			removeEventListener: { configurable: true, value: () => {} },
		});
		await assert.rejects(
			(orchestratorFor(transport).ensureChildIssue as any)(candidate, "evidence", { signal: controller.signal }),
			/cancel|abort|external/i,
		);
		assert.equal(transport.callContexts.length, 0);
		assert.equal(transport.createIssueCalls, 0);
	});

	await t.test("AbortSignal proxies reject without invoking traps or leaking host text", async () => {
		const candidate = await cycle3Plan(true);
		const transport = new FakeTransport();
		let trapped = false;
		const signal = new Proxy(new AbortController().signal, {
			get() {
				trapped = true;
				throw new Error("SYNTHETIC_CYCLE6_SIGNAL_PROXY");
			},
		});
		let rejection: unknown;
		try {
			await (orchestratorFor(transport).ensureChildIssue as any)(candidate, "evidence", { signal });
		} catch (error) {
			rejection = error;
		}
		assert.ok(rejection instanceof Error);
		assert.equal(trapped, false);
		assert.doesNotMatch(String(rejection), /SYNTHETIC_CYCLE6_SIGNAL_PROXY|incompatible receiver/u);
	});

	await t.test("acknowledgement before local abort remains unacknowledged", async () => {
		const candidate = await cycle3Plan(true);
		const transport = new FakeTransport();
		transport.findChildIssues = (async (_query: unknown, context?: TestCallContext & { acknowledgeAbort?: () => void }) => {
			context?.acknowledgeAbort?.();
			return new Promise(() => {});
		}) as never;
		const orchestrator = orchestratorFor(transport, undefined, defaultPolicySource(transport), 5);
		await assert.rejects(orchestrator.ensureChildIssue(candidate, "evidence"), /timeout|external/i);
		const stopped = await (orchestrator.stop as any)({ deadlineAt: new Date(Date.now() + 10).toISOString() });
		assert.equal(stopped.kind, "incomplete");
		assert.equal(stopped.active, 1);
		assert.equal(stopped.unacknowledged, 1);
	});
});

test("cycle 6 keeps exact-head review authority ordered and integration identity semantic", async (t) => {
	await t.test("equivalent later clean attempts preserve one mutation intent", async () => {
		const mutations: IntegrateChildRequest["mutation"][] = [];
		for (const completedAt of ["2026-07-21T12:00:00.000Z", "2026-07-21T12:02:00.000Z"]) {
			const candidate = await cycle3Plan(true);
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
			const handoff = childHandoff(issue.number, child.branch, child.prBase);
			const pullRequest = cleanPullRequest(childPullRequestRequest(candidate, child, handoff));
			pullRequest.reviews = [{ ...pullRequest.reviews[0], completedAt }];
			transport.pullRequests.push(pullRequest);
			transport.integrateChild = (async (request: IntegrateChildRequest) => {
				mutations.push(request.mutation);
				throw new Error("synthetic capture-only integration failure");
			}) as never;
			await assert.rejects(orchestratorFor(transport).integrateChild(candidate, child, handoff));
		}
		assert.equal(mutations.length, 2);
		assert.equal(mutations[0].idempotencyKey, mutations[1].idempotencyKey);
		assert.equal(mutations[0].intentDigest, mutations[1].intentDigest);
	});

	await t.test("restart reuses a receipt after an equivalent later clean attempt", async () => {
		const candidate = await cycle3Plan(true);
		const transport = new FakeTransport();
		const receipts = seedIntegrationRoster(candidate, transport);
		const receipt = receipts[0];
		const pullRequest = transport.pullRequests.find((entry) => entry.number === receipt.pullRequest);
		assert.ok(pullRequest);
		pullRequest.reviews.push({ ...pullRequest.reviews[0], completedAt: "2026-07-21T12:02:00.000Z" });
		const { child, handoff } = handoffForReceipt(candidate, transport, receipt);
		const result = await orchestratorFor(transport).integrateChild(candidate, child, handoff);
		assert.equal(result.kind, "integrated");
		if (result.kind === "integrated") assert.equal(result.reused, true);
		assert.equal(transport.integrateCalls, 0);
	});

	const authorizationDigest: Cycle6ReviewAuthorizationDigest = "0".repeat(64);
	const reviewCompletedAt: Cycle6ReviewCompletedAt = "2026-07-21T12:00:00.000Z";
	assert.equal(authorizationDigest.length, 64);
	assert.match(reviewCompletedAt, /^2026-/u);
});

test("cycle 6 applies receipt chronology before reuse and parent readiness", async (t) => {
	await t.test("new integration is no earlier than every authority observation", async () => {
		const candidate = await cycle3Plan(true);
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
		const handoff = childHandoff(issue.number, child.branch, child.prBase);
		transport.pullRequests.push(cleanPullRequest(childPullRequestRequest(candidate, child, handoff)));
		const result = await orchestratorFor(transport).integrateChild(candidate, child, handoff);
		assert.equal(result.kind, "integrated");
		if (result.kind !== "integrated") return;
		const integratedAt = new Date(result.receipt.integratedAt).valueOf();
		assert.ok(integratedAt >= new Date(result.receipt.pullRequestSnapshot.observedAt).valueOf());
		assert.ok(integratedAt >= new Date(result.receipt.observation.observedAt).valueOf());
		assert.ok(integratedAt >= new Date(result.receipt.controllerProvenance.observedAt).valueOf());
		assert.ok(integratedAt >= new Date(result.receipt.controllerProvenance.policyObservedAt).valueOf());
	});

	for (const [name, integratedAt] of [
		["before path evidence", "2026-07-21T11:57:00.000Z"],
		["before review completion", "2026-07-21T11:59:00.000Z"],
		["before PR snapshot", "2026-07-21T12:04:00.000Z"],
		["before policy observation", "2026-07-21T12:05:00.000Z"],
		["impossibly future", "2999-01-01T00:00:00.000Z"],
	] as const) {
		await t.test(name, async () => {
			const { candidate, transport, receipts } = await cycle5ReadinessScenario();
			const forged = structuredClone(receipts);
			forged[0].integratedAt = integratedAt;
			transport.integrations.splice(0, transport.integrations.length, ...forged);
			await assertReadinessDoesNotMutate(transport, () => orchestratorFor(transport, new FakeDecisionBroker())
				.reconcileParentReadiness(candidate, forged, decisionPolicy));
		});
	}
});

test("cycle 6 normalizes revoked orchestration arrays and applies the shared grammar to plans", async (t) => {
	await t.test("revoked child arrays reject without host IsArray text", async () => {
		const source = JSON.parse(await readFile(objectivePath, "utf8")) as Record<string, unknown>;
		const revoked = Proxy.revocable([], {});
		revoked.revoke();
		let rejection: unknown;
		try {
			createPlanFromSource({ ...source, children: revoked.proxy });
		} catch (error) {
			rejection = error;
		}
		assert.ok(rejection instanceof Error);
		assert.match(String(rejection), /array|proxy|shape|invalid/i);
		assert.doesNotMatch(String(rejection), /Cannot perform|revoked/i);
	});

	const source = JSON.parse(await readFile(objectivePath, "utf8")) as Record<string, unknown>;
	const child = (source.children as Array<Record<string, unknown>>)[0];
	const variants: Array<[string, Record<string, unknown>]> = [
		["parent title npm", { ...source, title: "//registry.invalid/:_authToken=SYNTHETIC_NPM_MARKER" }],
		["parent objective netrc", { ...source, objective: "machine github.com login maintainer password SYNTHETIC_NETRC_MARKER" }],
		["child title lowercase cloud", { ...source, children: [{ ...child, title: "aws_secret_access_key = SYNTHETIC_AWS_MARKER" }] }],
		["child objective credential file", { ...source, children: [{ ...child, objective: "credentials_file = /tmp/SYNTHETIC_CREDENTIAL_FILE" }] }],
	];
	for (const [name, value] of variants) {
		await t.test(name, () => assert.throws(() => createPlanFromSource(value), /credential|secret|sensitive/i));
	}
});

test("cycle 6 durable run state is current, exact, and non-self-referential", async () => {
	const raw = await readFile(".planning/phases/478-shepherd-github-parent-orchestration/RUN-STATE.json", "utf8");
	const state = JSON.parse(raw) as any;
	assert.equal(state.details.candidateRef, "HEAD");
	const historical = state.details.reviewHistory.find((entry: { cycle?: number }) => entry.cycle === 6);
	assert.equal(historical.priorReviewedCandidate, "dbce5b7d0c698bc802594211072fed77eff23c1c");
	assert.equal(historical.findingsConsolidatedIntoCycle, 7);
	assert.equal(state.details.checkpoints.cycle5Evidence, "63ac436fdac5fc46be7004f8109c4f068aa5749c");
	assert.equal(state.details.checkpoints.cycle6Plan, "2832993b93d07ea20197bad52ec23700fe21fc1e");
	assert.match(state.details.checkpoints.cycle6Red, /^[0-9a-f]{40}$/u);
	assert.match(state.details.checkpoints.cycle6Green, /^[0-9a-f]{40}$/u);
	assert.equal(state.details.checkpoints.cycle6Evidence, "dbce5b7d0c698bc802594211072fed77eff23c1c");
	assert.doesNotMatch(raw, /"cycle5Evidence"\s*:\s*null|"cycle6Evidence"\s*:\s*null/u);
});

const cycle7CredentialSamples = [
	"client-key-data: SYNTHETIC_KUBERNETES_KEY_DATA",
	"token: SYNTHETIC_KUBERNETES_TOKEN",
	'{"auth":"SYNTHETIC_DOCKER_AUTH"}',
	'{"identitytoken":"SYNTHETIC_DOCKER_IDENTITY_TOKEN"}',
	"aws_access_key_id = SYNTHETIC_AWS_ACCESS_KEY_ID",
	"aws_secret_access_key = SYNTHETIC_AWS_SECRET_ACCESS_KEY",
	"aws_session_token = SYNTHETIC_AWS_SESSION_TOKEN",
	"ASIAABCDEFGHIJKLMNOP",
] as const;

test("cycle 7 rejects every finite schema credential at every plan and decision-question boundary", async (t) => {
	const source = JSON.parse(await readFile(objectivePath, "utf8")) as Record<string, unknown>;
	const child = (source.children as Array<Record<string, unknown>>)[0];
	for (const [index, sample] of cycle7CredentialSamples.entries()) {
		await t.test(`schema form ${index + 1}`, async () => {
			const variants: Record<string, unknown>[] = [
				{ ...source, title: sample },
				{ ...source, objective: sample },
				{ ...source, children: [{ ...child, title: sample }] },
				{ ...source, children: [{ ...child, objective: sample }] },
				{ ...source, children: [{ ...child, verification: [{ id: "focused", kind: "test", description: sample }] }] },
			];
			for (const value of variants) {
				let rejection: unknown;
				try { createPlanFromSource(value); } catch (error) { rejection = error; }
				assert.ok(rejection instanceof Error);
				assert.match(rejection.message, /credential|secret|sensitive/i);
				assert.doesNotMatch(rejection.message, /SYNTHETIC_/u);
			}
			const { candidate, transport, receipts } = await cycle5ReadinessScenario();
			await assertReadinessDoesNotMutate(transport, () => orchestratorFor(transport, new FakeDecisionBroker())
				.reconcileParentReadiness(candidate, receipts, { ...decisionPolicy, question: sample }));
		});
	}
});

const cycle8CredentialAssignmentSuffixes = [
	"AUTHORIZATION",
	"TOKEN",
	"ACCESS_TOKEN",
	"REFRESH_TOKEN",
	"API_KEY",
	"PASSWORD",
	"SECRET",
	"CLIENT_SECRET",
	"PRIVATE_KEY",
	"DATABASE_URL",
	"CREDENTIAL",
	"CREDENTIALS",
	"COOKIE",
	"COOKIES",
	"SET_COOKIE",
	"SESSION",
	"SESSION_ID",
	"SESSION_TOKEN",
	"SESSION_COOKIE",
	"CSRF_TOKEN",
] as const;

test("cycle 8 rejects every provider-neutral credential suffix through orchestration consumers", async (t) => {
	const source = JSON.parse(await readFile(objectivePath, "utf8")) as Record<string, unknown>;
	const child = (source.children as Array<Record<string, unknown>>)[0];
	for (const [index, suffix] of cycle8CredentialAssignmentSuffixes.entries()) {
		await t.test(suffix.toLowerCase(), async () => {
			const marker = `CYCLE8_ASSIGNMENT_${index + 1}`;
			const sample = `UNLISTED_VENDOR_${suffix}=${marker}`;
			const variants: Record<string, unknown>[] = [
				{ ...source, title: sample },
				{ ...source, objective: sample },
				{ ...source, children: [{ ...child, title: sample }] },
				{ ...source, children: [{ ...child, objective: sample }] },
				{ ...source, children: [{ ...child, verification: [{ id: "focused", kind: "test", description: sample }] }] },
			];
			for (const value of variants) {
				let rejection: unknown;
				try { createPlanFromSource(value); } catch (error) { rejection = error; }
				assert.ok(rejection instanceof Error);
				assert.match(rejection.message, /credential|secret|sensitive/i);
				assert.doesNotMatch(rejection.message, new RegExp(marker, "u"));
			}
			const { candidate, transport, receipts } = await cycle5ReadinessScenario();
			let rejection: unknown;
			try {
				await orchestratorFor(transport, new FakeDecisionBroker()).reconcileParentReadiness(
					candidate,
					receipts,
					{ ...decisionPolicy, question: sample },
				);
			} catch (error) {
				rejection = error;
			}
			assert.ok(rejection instanceof Error);
			assert.match(rejection.message, /credential|secret|sensitive/i);
			assert.doesNotMatch(rejection.message, new RegExp(marker, "u"));
		});
	}

	const safePlan = createPlanFromSource({ ...source, title: "FEATURE_TOKEN=enabled" });
	assert.equal(safePlan.title, "FEATURE_TOKEN=enabled");
	const { candidate, transport, receipts } = await cycle5ReadinessScenario();
	const safeDecision = await orchestratorFor(transport, new FakeDecisionBroker()).reconcileParentReadiness(
		candidate,
		receipts,
		{ ...decisionPolicy, question: "FEATURE_TOKEN=enabled" },
	);
	assert.equal(safeDecision.kind, "ready");
});

function cycle7FutureParentDecision(
	request: GitHubDecisionRequest,
	mode: "creation" | "request_comment" | "decision" | "consumption" | "update" | "all",
): HumanDecisionRecord {
	const binding = {
		repository: request.repository,
		target: { kind: "pull_request" as const, number: request.pullRequest },
		generation: request.generation,
		headSha: request.headSha,
	};
	const record = createHumanDecisionRecord({
		requestId: request.requestId,
		gate: "parent_merge",
		binding,
		allowedOptions: ["approve-merge", "reject"],
		actorAllowlist: request.actorAllowlist,
		expiresAt: request.expiresAt,
		question: request.question,
	}, new Date("2026-07-21T12:00:00.000Z"));
	const future = "2026-07-22T00:00:02.000Z";
	const requestComment = {
		id: 1,
		url: `https://github.com/${request.repository}/pull/${request.pullRequest}#issuecomment-1`,
		actor: "shepherd-host",
		createdAt: ["creation", "request_comment", "all"].includes(mode)
			? future
			: "2026-07-21T12:00:00.000Z",
	};
	const decision = {
		...approvedDecision,
		decidedAt: ["creation", "request_comment", "decision", "all"].includes(mode)
			? "2026-07-22T00:00:03.000Z"
			: approvedDecision.decidedAt,
	};
	const consumedAt = ["creation", "request_comment", "decision", "consumption", "all"].includes(mode)
		? "2026-07-22T00:00:04.000Z"
		: "2026-07-21T12:00:40.000Z";
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

test("cycle 7 real broker adapter rejects each future chronology before parent ready", async (t) => {
	for (const mode of ["creation", "request_comment", "decision", "consumption", "update", "all"] as const) {
		await t.test(mode, async () => {
			const { candidate, transport, receipts } = await cycle5ReadinessScenario();
			const repository = new MemoryHumanDecisionRepository();
			const request: GitHubDecisionRequest = {
				requestId: decisionPolicy.requestId,
				gate: "parent_merge",
				repository: candidate.repository,
				parentIssue: candidate.parentIssue,
				pullRequest: 900,
				generation: candidate.generation,
				headSha: "e".repeat(40),
				allowedOptions: ["approve-merge", "reject"],
				actorAllowlist: [...decisionPolicy.actorAllowlist],
				expiresAt: decisionPolicy.expiresAt,
				question: decisionPolicy.question,
			};
			repository.records.set(request.requestId, cycle7FutureParentDecision(request, mode));
			const broker = new GitHubDecisionBroker(repository, {
				async getAuthenticatedActor(): Promise<string> { throw new Error("unexpected transport call"); },
				async listComments(): Promise<GitHubComment[]> { throw new Error("unexpected transport call"); },
				async createDecisionRequestComment(): Promise<GitHubComment> { throw new Error("unexpected transport call"); },
			}, {
				now: () => new Date("2026-07-22T00:00:00.000Z"),
				sleep: async () => {},
				polling: { maxAttempts: 1, initialDelayMs: 1, maxDelayMs: 1 },
				transportRetry: { maxAttempts: 1, initialDelayMs: 1, maxDelayMs: 1 },
			});
			await assertReadinessDoesNotMutate(transport, () => orchestratorFor(
				transport,
				githubOrchestratorApi.adaptGitHubDecisionBroker(broker),
			).reconcileParentReadiness(candidate, receipts, decisionPolicy));
		});
	}
});

test("cycle 7 receipt audit fields must match an authoritative attested review attempt", async (t) => {
	for (const [name, mutate] of [
		["forged result digest", (receipt: ChildIntegrationReceipt) => { receipt.controllerProvenance.reviewResultDigest = "f".repeat(64); }],
		["forged completion time", (receipt: ChildIntegrationReceipt) => { receipt.controllerProvenance.reviewCompletedAt = "2026-07-21T11:59:30.000Z"; }],
	] as const) {
		await t.test(name, async () => {
			const { candidate, transport, receipts } = await cycle5ReadinessScenario();
			const forged = structuredClone(receipts);
			mutate(forged[0]);
			transport.integrations.splice(0, transport.integrations.length, ...forged);
			await assertReadinessDoesNotMutate(transport, () => orchestratorFor(transport, new FakeDecisionBroker())
				.reconcileParentReadiness(candidate, forged, decisionPolicy));
		});
	}

	await t.test("equivalent later clean preserves stable authority while retaining original attempt history", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		const childPullRequest = transport.pullRequests.find((entry) => entry.number === receipts[0].pullRequest);
		assert.ok(childPullRequest);
		childPullRequest.reviews.push({
			...childPullRequest.reviews[0],
			completedAt: "2026-07-21T12:02:00.000Z",
		});
		const result = await orchestratorFor(transport, new FakeDecisionBroker())
			.reconcileParentReadiness(candidate, receipts, decisionPolicy);
		assert.equal(result.kind, "ready");
	});
});

async function captureCycle7ReadyRequest(options: {
	policyObservedAt?: string;
	ancestryObservedAt?: string;
	ancestryRevision?: number;
	parentReviewCompletedAt?: string;
} = {}): Promise<MarkParentReadyRequest> {
	const { candidate, transport, receipts } = await cycle5ReadinessScenario();
	const parent = transport.pullRequests.find((entry) => entry.number === 900);
	assert.ok(parent);
	if (options.parentReviewCompletedAt !== undefined) {
		parent.reviews = [{ ...parent.reviews[0], completedAt: options.parentReviewCompletedAt }];
	}
	transport.ancestryProof = async (query) => ({
		schemaVersion: 1,
		authority: "transport",
		...query,
		result: true,
		revision: options.ancestryRevision ?? 1,
		observedAt: options.ancestryObservedAt ?? "2026-07-21T12:05:00.000Z",
	});
	const baseline = defaultPolicySource(transport);
	const policySource = {
		async findRequiredCheckPolicies(query: { repository: string; baseBranch: string }, context?: TestCallContext): Promise<unknown> {
			const result = await baseline.findRequiredCheckPolicies(query, context) as { items: Array<Record<string, unknown>>; complete: boolean };
			return {
				...result,
				items: result.items.map((item) => ({
					...item,
					observedAt: options.policyObservedAt ?? item.observedAt,
				})),
			};
		},
	};
	let captured: MarkParentReadyRequest | undefined;
	const authority: ParentReadyDurableAuthorityBoundary = {
		async compareConsumeAndMarkParentReady(request: MarkParentReadyRequest, _context: ExternalCallContext) {
			captured = request;
			return { schemaVersion: 1, kind: "conflict", coordinate: "authorization_state" };
		},
		async quarantineAndRollbackParentReady() {
			throw new Error("cycle 7 capture boundary should not recover a pre-effect failure");
		},
	};
	try {
		await orchestratorFor(transport, new FakeDecisionBroker(), policySource, 25, authority)
			.reconcileParentReadiness(candidate, receipts, decisionPolicy);
	} catch {
		// The capture boundary deliberately conflicts before the effect.
	}
	assert.ok(captured, "production authority boundary must receive the prepared request");
	return captured;
}

class DelayedParentReadyAuthority implements ParentReadyDurableAuthorityBoundary {
	compareCalls = 0;
	quarantineCalls = 0;
	quarantined = false;
	effectStarted = false;
	rollbackFailures = 0;
	readonly transport: FakeTransport;
	readonly mode: "before_effect" | "after_effect";
	readonly delayMs: number;

	constructor(
		transport: FakeTransport,
		mode: "before_effect" | "after_effect",
		delayMs = 80,
	) {
		this.transport = transport;
		this.mode = mode;
		this.delayMs = delayMs;
	}

	async compareConsumeAndMarkParentReady(
		request: MarkParentReadyRequest,
		context: ExternalCallContext,
	): Promise<any> {
		this.compareCalls += 1;
		if (this.mode === "before_effect") {
			await new Promise<void>((resolve) => setTimeout(resolve, this.delayMs));
			if (this.quarantined) throw new Error("durable authority was quarantined before effect");
			this.effectStarted = true;
			return {
				schemaVersion: 1,
				kind: "applied",
				mutation: await this.transport.markParentReady(request, context),
			};
		}
		this.effectStarted = true;
		const result = await this.transport.markParentReady(request, context);
		await new Promise<void>((resolve) => setTimeout(resolve, this.delayMs));
		return { schemaVersion: 1, kind: "applied", mutation: result };
	}

	async quarantineAndRollbackParentReady(
		request: RollbackParentReadyRequest,
		context: ExternalCallContext,
	): Promise<any> {
		this.quarantineCalls += 1;
		this.quarantined = true;
		if (this.rollbackFailures > 0) {
			this.rollbackFailures -= 1;
			throw new Error("transient durable quarantine rollback failure");
		}
		return this.transport.rollbackParentReady(request, context);
	}
}

test("cycle 7 joins and quarantines every uncertain late parent-ready effect", async (t) => {
	for (const mode of ["before_effect", "after_effect"] as const) {
		await t.test(`500 ms ${mode.replace("_", " ")} after 100 ms timeout`, async () => {
			const { candidate, transport, receipts } = await cycle5ReadinessScenario();
			const authority = new DelayedParentReadyAuthority(transport, mode, 500);
			const orchestrator = orchestratorFor(transport, new FakeDecisionBroker(), defaultPolicySource(transport), 100, authority);
			await assert.rejects(
				orchestrator.reconcileParentReadiness(candidate, receipts, decisionPolicy),
				/timeout|cancel|external|ready/i,
			);
			const second = await settleWithin(orchestrator.reconcileParentReadiness(candidate, receipts, decisionPolicy)
				.then(() => "resolved", () => "rejected"), 150);
			assert.notEqual(second, "resolved");
			assert.equal(authority.compareCalls, 1, "same key stays owned while the first effect is live");
			await new Promise<void>((resolve) => setTimeout(resolve, 650));
			assert.equal(transport.pullRequests.find((entry) => entry.number === 900)?.draft, true);
			assert.ok(authority.quarantineCalls >= 1);
			assert.equal((await orchestrator.stop()).kind, "joined");
		});
	}

	await t.test("caller cancellation before a 500 ms late effect is quarantined and joined", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		const authority = new DelayedParentReadyAuthority(transport, "before_effect", 500);
		const orchestrator = orchestratorFor(transport, new FakeDecisionBroker(), defaultPolicySource(transport), 1_000, authority);
		const controller = new AbortController();
		const cancelled = orchestrator.reconcileParentReadiness(
			candidate,
			receipts,
			decisionPolicy,
			{ signal: controller.signal },
		);
		setTimeout(() => controller.abort(), 30);
		await assert.rejects(cancelled, /cancel|abort|external|ready/i);
		await new Promise<void>((resolve) => setTimeout(resolve, 650));
		assert.equal(authority.compareCalls, 1);
		assert.ok(authority.quarantineCalls >= 1);
		assert.equal(transport.pullRequests.find((entry) => entry.number === 900)?.draft, true);
		assert.equal((await orchestrator.stop()).kind, "joined");
	});

	await t.test("restart before ready visibility observes durable quarantine", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		transport.parentReadyVisibilityLag = 10;
		const authority = new DelayedParentReadyAuthority(transport, "after_effect");
		const first = orchestratorFor(transport, new FakeDecisionBroker(), defaultPolicySource(transport), 20, authority);
		await assert.rejects(first.reconcileParentReadiness(candidate, receipts, decisionPolicy));
		const restarted = orchestratorFor(transport, new FakeDecisionBroker(), defaultPolicySource(transport), 20, authority);
		const restartResult = await restarted.reconcileParentReadiness(candidate, receipts, decisionPolicy);
		assert.notEqual(restartResult.kind, "ready");
		await new Promise<void>((resolve) => setTimeout(resolve, 140));
		assert.equal(transport.pullRequests.find((entry) => entry.number === 900)?.draft, true);
		assert.equal(authority.compareCalls, 1);
	});

	await t.test("read failure after the effect cannot prevent boundary-owned rollback", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		const authority = new DelayedParentReadyAuthority(transport, "after_effect");
		transport.onPullRequestRead = (_query, matches) => {
			if (authority.effectStarted) throw new Error("synthetic post-effect read failure");
			return matches;
		};
		const orchestrator = orchestratorFor(transport, new FakeDecisionBroker(), defaultPolicySource(transport), 20, authority);
		await assert.rejects(orchestrator.reconcileParentReadiness(candidate, receipts, decisionPolicy));
		await new Promise<void>((resolve) => setTimeout(resolve, 140));
		assert.equal(transport.pullRequests.find((entry) => entry.number === 900)?.draft, true);
		assert.ok(authority.quarantineCalls >= 1);
	});

	await t.test("transient rollback failure retries before key and join release", async () => {
		const { candidate, transport, receipts } = await cycle5ReadinessScenario();
		const authority = new DelayedParentReadyAuthority(transport, "after_effect");
		authority.rollbackFailures = 1;
		const orchestrator = orchestratorFor(transport, new FakeDecisionBroker(), defaultPolicySource(transport), 20, authority);
		await assert.rejects(orchestrator.reconcileParentReadiness(candidate, receipts, decisionPolicy));
		const earlyStop = await orchestrator.stop({ deadlineAt: new Date(Date.now() + 10).toISOString() });
		assert.equal(earlyStop.kind, "incomplete");
		await new Promise<void>((resolve) => setTimeout(resolve, 160));
		assert.ok(authority.quarantineCalls >= 2);
		assert.equal(transport.pullRequests.find((entry) => entry.number === 900)?.draft, true);
		assert.equal((await orchestrator.stop()).kind, "joined");
	});
});

class ImmediateApplyThenRejectAuthority implements ParentReadyDurableAuthorityBoundary {
	compareCalls = 0;
	rollbackCalls = 0;
	readonly rollbackRequests: RollbackParentReadyRequest[] = [];
	readonly #transport: FakeTransport;
	readonly #recoveryStarted: Promise<void>;
	#announceRecovery = (): void => {};
	readonly #recoveryGate: Promise<void>;
	#releaseRecovery = (): void => {};

	constructor(transport: FakeTransport) {
		this.#transport = transport;
		this.#recoveryStarted = new Promise<void>((resolve) => { this.#announceRecovery = resolve; });
		this.#recoveryGate = new Promise<void>((resolve) => { this.#releaseRecovery = resolve; });
	}

	get recoveryStarted(): Promise<void> {
		return this.#recoveryStarted;
	}

	releaseRecovery(): void {
		this.#releaseRecovery();
	}

	async compareConsumeAndMarkParentReady(
		request: MarkParentReadyRequest,
		context: ExternalCallContext,
	): Promise<ParentReadyCompareEffectResult> {
		this.compareCalls += 1;
		await this.#transport.markParentReady(request, context);
		throw new Error("synthetic response serialization failure after durable ready effect");
	}

	async quarantineAndRollbackParentReady(
		request: RollbackParentReadyRequest,
		_context: ExternalCallContext,
	): Promise<DurableMutationResult<GitHubPullRequestEvidence>> {
		this.rollbackCalls += 1;
		this.rollbackRequests.push(structuredClone(request));
		this.#announceRecovery();
		await this.#recoveryGate;
		const index = this.#transport.pullRequests.findIndex((entry) => entry.number === request.pullRequest);
		const current = this.#transport.pullRequests[index];
		if (index < 0 || current === undefined) throw new Error("parent pull request missing");
		const draft = { ...current, draft: true, revision: current.revision + 1 };
		this.#transport.pullRequests.splice(index, 1, draft);
		return {
			schemaVersion: 1,
			idempotencyKey: request.mutation.idempotencyKey,
			intentDigest: request.mutation.intentDigest,
			revision: this.rollbackCalls,
			applied: true,
			value: structuredClone(draft),
		};
	}
}

async function cycle8ImmediateUncertainScenario(): Promise<{
	candidate: ParentOrchestrationPlan;
	receipts: ChildIntegrationReceipt[];
	transport: FakeTransport;
	authority: ImmediateApplyThenRejectAuthority;
	orchestrator: GitHubParentOrchestrator;
	first: Promise<"rejected" | "resolved">;
}> {
	const { candidate, transport, receipts } = await cycle5ReadinessScenario();
	const authority = new ImmediateApplyThenRejectAuthority(transport);
	transport.onPullRequestRead = (_query, matches) => {
		if (authority.compareCalls > 0) throw new Error("synthetic all-read outage after ready effect");
		return matches;
	};
	const orchestrator = orchestratorFor(
		transport,
		new FakeDecisionBroker(),
		defaultPolicySource(transport),
		25,
		authority,
	);
	const first = orchestrator.reconcileParentReadiness(candidate, receipts, decisionPolicy)
		.then(() => "resolved" as const, () => "rejected" as const);
	return { candidate, receipts, transport, authority, orchestrator, first };
}

test("cycle 8 quarantines every immediate uncertain non-value outcome", async (t) => {
	await t.test("apply then reject starts recovery even when every visibility read fails", async () => {
		const scenario = await cycle8ImmediateUncertainScenario();
		assert.equal(await settleWithin(scenario.first, 100), "rejected");
		const started = await settleWithin(scenario.authority.recoveryStarted, 100);
		scenario.authority.releaseRecovery();
		assert.notEqual(started, "hung");
		await new Promise<void>((resolve) => setTimeout(resolve, 20));
	});

	await t.test("durable recovery restores the exact parent draft", async () => {
		const scenario = await cycle8ImmediateUncertainScenario();
		await scenario.first;
		const started = await settleWithin(scenario.authority.recoveryStarted, 100);
		scenario.authority.releaseRecovery();
		assert.notEqual(started, "hung");
		await new Promise<void>((resolve) => setTimeout(resolve, 30));
		const parent = scenario.transport.pullRequests.find((entry) => entry.number === 900);
		assert.ok(parent);
		assert.equal(parent.draft, true);
	});

	await t.test("same keyed operation cannot reenter while recovery owns the effect", async () => {
		const scenario = await cycle8ImmediateUncertainScenario();
		await scenario.first;
		const started = await settleWithin(scenario.authority.recoveryStarted, 100);
		assert.notEqual(started, "hung");
		scenario.transport.onPullRequestRead = undefined;
		const second = await settleWithin(
			scenario.orchestrator.reconcileParentReadiness(scenario.candidate, scenario.receipts, decisionPolicy)
				.then(() => "resolved" as const, () => "rejected" as const),
			80,
		);
		assert.notEqual(second, "resolved");
		assert.equal(scenario.authority.compareCalls, 1);
		scenario.authority.releaseRecovery();
		await new Promise<void>((resolve) => setTimeout(resolve, 30));
	});

	await t.test("stop remains incomplete until durable recovery settles and then joins", async () => {
		const scenario = await cycle8ImmediateUncertainScenario();
		await scenario.first;
		const started = await settleWithin(scenario.authority.recoveryStarted, 100);
		assert.notEqual(started, "hung");
		const early = await scenario.orchestrator.stop({ deadlineAt: new Date(Date.now() + 10).toISOString() });
		assert.equal(early.kind, "incomplete");
		scenario.authority.releaseRecovery();
		await new Promise<void>((resolve) => setTimeout(resolve, 30));
		assert.equal((await scenario.orchestrator.stop()).kind, "joined");
	});
});

class FencedRollbackAuthority implements ParentReadyDurableAuthorityBoundary {
	readonly rollbackRequests: RollbackParentReadyRequest[] = [];
	readonly rollbackContexts: ExternalCallContext[] = [];
	readyRequest?: MarkParentReadyRequest;
	firstAttemptAborted = false;
	readonly #transport: FakeTransport;
	readonly #secondAttemptStarted: Promise<void>;
	#announceSecondAttempt = (): void => {};
	readonly #secondAttemptGate: Promise<void>;
	#releaseSecondAttempt = (): void => {};
	#resolveFirstAttempt = (): void => {};
	#staleReady?: GitHubPullRequestEvidence;

	constructor(transport: FakeTransport) {
		this.#transport = transport;
		this.#secondAttemptStarted = new Promise<void>((resolve) => { this.#announceSecondAttempt = resolve; });
		this.#secondAttemptGate = new Promise<void>((resolve) => { this.#releaseSecondAttempt = resolve; });
	}

	get secondAttemptStarted(): Promise<void> {
		return this.#secondAttemptStarted;
	}

	releaseSecondAttempt(): void {
		this.#releaseSecondAttempt();
	}

	resolveSupersededFirstAttempt(): void {
		this.#resolveFirstAttempt();
	}

	async compareConsumeAndMarkParentReady(
		request: MarkParentReadyRequest,
		context: ExternalCallContext,
	): Promise<ParentReadyCompareEffectResult> {
		this.readyRequest = structuredClone(request);
		const mutation = await this.#transport.markParentReady(request, context);
		this.#staleReady = structuredClone(mutation.value);
		await new Promise<void>((resolve) => setTimeout(resolve, 120));
		return { schemaVersion: 1, kind: "applied", mutation };
	}

	async quarantineAndRollbackParentReady(
		request: RollbackParentReadyRequest,
		context: ExternalCallContext,
	): Promise<DurableMutationResult<GitHubPullRequestEvidence>> {
		this.rollbackRequests.push(structuredClone(request));
		this.rollbackContexts.push(context);
		const attempt = this.rollbackRequests.length;
		if (attempt === 1) {
			context.signal.addEventListener("abort", () => { this.firstAttemptAborted = true; }, { once: true });
			return new Promise<DurableMutationResult<GitHubPullRequestEvidence>>((resolve) => {
				this.#resolveFirstAttempt = () => {
					const staleReady = this.#staleReady;
					if (staleReady === undefined) throw new Error("missing stale ready response");
					resolve({
						schemaVersion: 1,
						idempotencyKey: request.mutation.idempotencyKey,
						intentDigest: request.mutation.intentDigest,
						revision: 1,
						applied: true,
						value: structuredClone(staleReady),
					});
				};
			});
		}
		this.#announceSecondAttempt();
		await this.#secondAttemptGate;
		const index = this.#transport.pullRequests.findIndex((entry) => entry.number === request.pullRequest);
		const current = this.#transport.pullRequests[index];
		if (index < 0 || current === undefined) throw new Error("parent pull request missing");
		const draft = current.draft ? current : { ...current, draft: true, revision: current.revision + 1 };
		this.#transport.pullRequests.splice(index, 1, draft);
		return {
			schemaVersion: 1,
			idempotencyKey: request.mutation.idempotencyKey,
			intentDigest: request.mutation.intentDigest,
			revision: attempt,
			applied: true,
			value: structuredClone(draft),
		};
	}
}

test("cycle 8 bounds and durably fences rollback attempts before releasing quarantine", async () => {
	const { candidate, transport, receipts } = await cycle5ReadinessScenario();
	const authority = new FencedRollbackAuthority(transport);
	const orchestrator = orchestratorFor(
		transport,
		new FakeDecisionBroker(),
		defaultPolicySource(transport),
		20,
		authority,
	);
	await assert.rejects(
		orchestrator.reconcileParentReadiness(candidate, receipts, decisionPolicy),
		/timeout|external|ready/i,
	);
	const retried = await settleWithin(authority.secondAttemptStarted, 250);
	assert.notEqual(retried, "hung", "a non-settling rollback response must not defeat the next fenced attempt");
	assert.equal(authority.firstAttemptAborted, true, "the first rollback response wait receives a real abort");
	assert.ok(authority.readyRequest);
	assert.ok(authority.rollbackRequests.length >= 2);
	const first = authority.rollbackRequests[0];
	const second = authority.rollbackRequests[1];
	const firstFence: ParentReadyRecoveryFence = first.recovery;
	const secondFence: ParentReadyRecoveryFence = second.recovery;
	assert.equal(firstFence.recoveryId, secondFence.recoveryId);
	assert.equal(firstFence.attempt, 1);
	assert.equal(secondFence.attempt, 2);
	assert.equal(firstFence.supersedesAttempt, null);
	assert.equal(secondFence.supersedesAttempt, 1);
	assert.deepEqual(firstFence.readyMutation, authority.readyRequest.mutation);
	assert.deepEqual(secondFence.readyMutation, authority.readyRequest.mutation);
	assert.equal(first.mutation.expectedResourceRevision, null);
	assert.equal(second.mutation.expectedResourceRevision, null);
	const blockedStop = await orchestrator.stop({ deadlineAt: new Date(Date.now() + 10).toISOString() });
	assert.equal(blockedStop.kind, "incomplete", "a claimed fence alone is not an authoritative draft observation");
	authority.releaseSecondAttempt();
	await new Promise<void>((resolve) => setTimeout(resolve, 150));
	const draft = transport.pullRequests.find((entry) => entry.number === 900);
	assert.ok(draft);
	assert.equal(draft.draft, true);
	assert.equal((await orchestrator.stop()).kind, "joined");
	authority.resolveSupersededFirstAttempt();
	await new Promise<void>((resolve) => setTimeout(resolve, 20));
	assert.equal(transport.pullRequests.find((entry) => entry.number === 900)?.draft, true);
	assert.equal((await orchestrator.stop()).kind, "joined", "a superseded late result cannot resettle controller state");
});

test("cycle 7 stable parent-ready identity excludes volatile freshness metadata", async (t) => {
	const baseline = await captureCycle7ReadyRequest();
	for (const [name, refreshed] of [
		["policy observation", await captureCycle7ReadyRequest({ policyObservedAt: "2026-07-21T12:07:00.000Z" })],
		["ancestry observation", await captureCycle7ReadyRequest({ ancestryObservedAt: "2026-07-21T12:07:00.000Z", ancestryRevision: 2 })],
		["equivalent later clean", await captureCycle7ReadyRequest({ parentReviewCompletedAt: "2026-07-21T12:02:00.000Z" })],
		["combined retry refresh", await captureCycle7ReadyRequest({
			policyObservedAt: "2026-07-21T12:08:00.000Z",
			ancestryObservedAt: "2026-07-21T12:08:00.000Z",
			ancestryRevision: 3,
			parentReviewCompletedAt: "2026-07-21T12:03:00.000Z",
		})],
	] as const) {
		await t.test(name, () => {
			assert.equal(refreshed.mutation.idempotencyKey, baseline.mutation.idempotencyKey);
			assert.equal(refreshed.mutation.intentDigest, baseline.mutation.intentDigest);
			assert.equal(refreshed.authorization.digest, baseline.authorization.digest);
			const baselineFreshness: ParentReadyFreshnessEnvelope = baseline.freshness;
			const refreshedFreshness: ParentReadyFreshnessEnvelope = refreshed.freshness;
			assert.notEqual(refreshedFreshness.digest, baselineFreshness.digest);
		});
	}
});

type Cycle8FreshnessRefresh = "policy" | "ancestry" | "review" | "combined";

async function captureCycle8PublicPreparedCommit(
	refresh: Cycle8FreshnessRefresh,
): Promise<{
	journaled: PreparedParentReadyOperation;
	journaledBeforeCommit: PreparedParentReadyOperation;
	refreshed: PreparedParentReadyOperation;
	request: MarkParentReadyRequest;
}> {
	const { candidate, transport, receipts } = await cycle5ReadinessScenario();
	let policyObservedAt = "2026-07-21T12:06:00.000Z";
	let ancestryObservedAt = "2026-07-21T12:05:00.000Z";
	let ancestryRevision = 1;
	transport.ancestryProof = async (query) => ({
		schemaVersion: 1,
		authority: "transport",
		...query,
		result: true,
		revision: ancestryRevision,
		observedAt: ancestryObservedAt,
	});
	const baselinePolicySource = defaultPolicySource(transport);
	const policySource: RequiredCheckPolicySource = {
		async findRequiredCheckPolicies(query, context) {
			const found = await baselinePolicySource.findRequiredCheckPolicies(query, context) as {
				items: RequiredGitHubCheckPolicyObservation[];
				complete: boolean;
			};
			return {
				...found,
				items: found.items.map((item) => ({ ...item, observedAt: policyObservedAt })),
			};
		},
	};
	let captured: MarkParentReadyRequest | undefined;
	const authority: ParentReadyDurableAuthorityBoundary = {
		async compareConsumeAndMarkParentReady(request): Promise<ParentReadyCompareEffectResult> {
			captured = structuredClone(request);
			return { schemaVersion: 1, kind: "conflict", coordinate: "authorization_state" };
		},
		async quarantineAndRollbackParentReady(): Promise<DurableMutationResult<GitHubPullRequestEvidence>> {
			throw new Error("freshness capture must not invoke rollback");
		},
	};
	const orchestrator = orchestratorFor(transport, new FakeDecisionBroker(), policySource, 25, authority);
	const prepared = await orchestrator.prepareParentReadiness(candidate, receipts, decisionPolicy);
	assert.equal(prepared.kind, "prepared");
	if (prepared.kind !== "prepared") throw new Error("parent readiness did not prepare");
	const journaled = structuredClone(prepared.operation);
	if (refresh === "policy" || refresh === "combined") {
		policyObservedAt = "2026-07-21T12:08:00.000Z";
	}
	if (refresh === "ancestry" || refresh === "combined") {
		ancestryObservedAt = "2026-07-21T12:08:00.000Z";
		ancestryRevision = 3;
	}
	if (refresh === "review" || refresh === "combined") {
		const parent = transport.pullRequests.find((entry) => entry.number === 900);
		assert.ok(parent);
		parent.reviews.push({ ...parent.reviews[0], completedAt: "2026-07-21T12:03:00.000Z" });
	}
	const current = await orchestrator.prepareParentReadiness(candidate, receipts, decisionPolicy);
	assert.equal(current.kind, "prepared");
	if (current.kind !== "prepared") throw new Error("refreshed parent readiness did not prepare");
	const journaledBeforeCommit = structuredClone(journaled);
	const result = await orchestrator.commitPreparedParentReadiness(candidate, receipts, journaled);
	assert.deepEqual(result, {
		kind: "blocked",
		blockers: ["parent_ready_authority_conflict:authorization_state"],
	});
	assert.ok(captured);
	return { journaled, journaledBeforeCommit, refreshed: current.operation, request: captured };
}

test("cycle 8 public prepared commit forwards refreshed evidence with stable journaled identity", async (t) => {
	for (const [name, refresh] of [
		["policy observation", "policy"],
		["ancestry observation", "ancestry"],
		["equivalent later clean review", "review"],
		["combined freshness envelope", "combined"],
	] as const) {
		await t.test(name, async () => {
			const captured = await captureCycle8PublicPreparedCommit(refresh);
			assert.notEqual(captured.refreshed.freshness.digest, captured.journaled.freshness.digest);
			assert.deepEqual(captured.request.freshness, captured.refreshed.freshness);
		});
	}

	await t.test("original authorization key and intent remain exact while the journal stays immutable", async () => {
		const captured = await captureCycle8PublicPreparedCommit("combined");
		assert.deepEqual(captured.request.authorization, captured.journaled.authorization);
		assert.deepEqual(captured.request.mutation, captured.journaled.mutation);
		assert.deepEqual(captured.journaled, captured.journaledBeforeCommit);
		assert.deepEqual(captured.request.freshness, captured.refreshed.freshness);
		assert.notEqual(captured.request.freshness.digest, captured.journaled.freshness.digest);
	});
});

test("cycle 8 actual broker adapter resumes one persisted consumed authorization after expiry", async () => {
	const { candidate, transport, receipts } = await cycle5ReadinessScenario();
	const repository = new MemoryHumanDecisionRepository();
	const request: GitHubDecisionRequest = {
		requestId: decisionPolicy.requestId,
		gate: "parent_merge",
		repository: candidate.repository,
		parentIssue: candidate.parentIssue,
		pullRequest: 900,
		generation: candidate.generation,
		headSha: "e".repeat(40),
		allowedOptions: ["approve-merge", "reject"],
		actorAllowlist: [...decisionPolicy.actorAllowlist],
		expiresAt: decisionPolicy.expiresAt,
		question: decisionPolicy.question,
	};
	const binding = {
		repository: request.repository,
		target: { kind: "pull_request" as const, number: request.pullRequest },
		generation: request.generation,
		headSha: request.headSha,
	};
	repository.records.set(request.requestId, createHumanDecisionRecord({
		requestId: request.requestId,
		gate: request.gate,
		binding,
		allowedOptions: request.allowedOptions,
		actorAllowlist: request.actorAllowlist,
		expiresAt: request.expiresAt,
		question: request.question,
	}, new Date("2026-07-21T12:00:00.000Z")));
	await recordHumanDecisionRequestComment(repository, request.requestId, binding, {
		id: 1,
		url: `https://github.com/${request.repository}/pull/${request.pullRequest}#issuecomment-1`,
		actor: "shepherd-host",
		createdAt: "2026-07-21T12:00:10.000Z",
	}, new Date("2026-07-21T12:00:10.000Z"));
	await recordHumanDecision(repository, request.requestId, binding, approvedDecision);
	await consumeHumanDecision(repository, request.requestId, binding, new Date("2026-07-21T12:00:40.000Z"));
	let unexpectedGitHubCalls = 0;
	const githubTransport: GitHubDecisionTransport = {
		async getAuthenticatedActor(): Promise<string> {
			unexpectedGitHubCalls += 1;
			throw new Error("existing consumed decision must not query GitHub actor");
		},
		async listComments(): Promise<GitHubComment[]> {
			unexpectedGitHubCalls += 1;
			throw new Error("existing consumed decision must not list GitHub comments");
		},
		async createDecisionRequestComment(): Promise<GitHubComment> {
			unexpectedGitHubCalls += 1;
			throw new Error("existing consumed decision must not publish another comment");
		},
	};
	const beforeExpiry = githubOrchestratorApi.adaptGitHubDecisionBroker(new GitHubDecisionBroker(
		repository,
		githubTransport,
		{ now: () => new Date("2026-07-21T12:01:00.000Z") },
	));
	const firstController = orchestratorFor(
		transport,
		beforeExpiry,
		defaultPolicySource(transport),
		25,
		transport,
		() => new Date("2026-07-21T12:01:00.000Z"),
	);
	const prepared = await firstController.prepareParentReadiness(candidate, receipts, decisionPolicy);
	assert.equal(prepared.kind, "prepared");
	if (prepared.kind !== "prepared") throw new Error("parent readiness did not prepare before expiry");
	const journaled = structuredClone(prepared.operation);
	const afterExpiry = githubOrchestratorApi.adaptGitHubDecisionBroker(new GitHubDecisionBroker(
		repository,
		githubTransport,
		{ now: () => new Date("2028-07-21T12:01:00.000Z") },
	));
	const restartedController = orchestratorFor(
		transport,
		afterExpiry,
		defaultPolicySource(transport),
		25,
		transport,
		() => new Date("2028-07-21T12:01:00.000Z"),
	);
	const committed = await restartedController.commitPreparedParentReadiness(candidate, receipts, journaled);
	assert.equal(committed.kind, "ready");
	assert.equal(transport.markReadyCalls, 1);
	const replayed = await restartedController.commitPreparedParentReadiness(candidate, receipts, journaled);
	assert.equal(replayed.kind, "ready");
	assert.equal(transport.markReadyCalls, 1, "the journaled authorization applies exactly once");
	assert.equal(unexpectedGitHubCalls, 0);
	assert.equal((await restartedController.stop()).kind, "joined");
});

test("cycle 7 timeout retry reuses one intent after harmless authority refresh", async () => {
	const { candidate, transport, receipts } = await cycle5ReadinessScenario();
	let policyObservedAt = "2026-07-21T12:06:00.000Z";
	let ancestryObservedAt = "2026-07-21T12:06:00.000Z";
	let ancestryRevision = 1;
	transport.ancestryProof = async (query) => ({
		schemaVersion: 1,
		authority: "transport",
		...query,
		result: true,
		revision: ancestryRevision,
		observedAt: ancestryObservedAt,
	});
	const baselinePolicySource = defaultPolicySource(transport);
	const policySource: RequiredCheckPolicySource = {
		async findRequiredCheckPolicies(query, context) {
			const result = await baselinePolicySource.findRequiredCheckPolicies(query, context) as {
				items: RequiredGitHubCheckPolicyObservation[];
				complete: boolean;
			};
			return {
				...result,
				items: result.items.map((item) => ({ ...item, observedAt: policyObservedAt })),
			};
		},
	};
	const requests: MarkParentReadyRequest[] = [];
	let first = true;
	const authority: ParentReadyDurableAuthorityBoundary = {
		async compareConsumeAndMarkParentReady(request, context): Promise<any> {
			requests.push(structuredClone(request));
			if (first) {
				first = false;
				await new Promise<void>((resolve) => setTimeout(resolve, 120));
				throw new Error("synthetic pre-effect settlement without mutation");
			}
			return {
				schemaVersion: 1,
				kind: "applied",
				mutation: await transport.markParentReady(request, context),
			};
		},
		async quarantineAndRollbackParentReady(request): Promise<any> {
			const parent = transport.pullRequests.find((entry) => entry.number === request.pullRequest);
			if (parent === undefined || !parent.draft) throw new Error("unexpected parent state");
			return {
				schemaVersion: 1,
				idempotencyKey: request.mutation.idempotencyKey,
				intentDigest: request.mutation.intentDigest,
				revision: 1,
				applied: true,
				value: structuredClone(parent),
			};
		},
	};
	const orchestrator = orchestratorFor(
		transport,
		new FakeDecisionBroker(),
		policySource,
		30,
		authority,
	);
	await assert.rejects(
		orchestrator.reconcileParentReadiness(candidate, receipts, decisionPolicy),
		/timeout|external|ready/i,
	);
	await new Promise<void>((resolve) => setTimeout(resolve, 180));
	policyObservedAt = "2026-07-21T12:08:00.000Z";
	ancestryObservedAt = "2026-07-21T12:08:00.000Z";
	ancestryRevision = 2;
	const parent = transport.pullRequests.find((entry) => entry.number === 900);
	assert.ok(parent);
	parent.reviews.push({ ...parent.reviews[0], completedAt: "2026-07-21T12:03:00.000Z" });
	const retried = await orchestrator.reconcileParentReadiness(candidate, receipts, decisionPolicy);
	assert.equal(retried.kind, "ready");
	assert.equal(requests.length, 2);
	assert.equal(requests[1].mutation.idempotencyKey, requests[0].mutation.idempotencyKey);
	assert.equal(requests[1].mutation.intentDigest, requests[0].mutation.intentDigest);
	assert.equal(requests[1].authorization.digest, requests[0].authorization.digest);
	assert.notEqual(requests[1].freshness.digest, requests[0].freshness.digest);
});

interface PortOnlyReadinessSeed {
	issues: readonly GitHubChildIssue[];
	pullRequests: readonly GitHubPullRequestEvidence[];
	rosters: readonly GitHubRosterSnapshot[];
	integrations: readonly ChildIntegrationReceipt[];
}

class PortOnlyReadinessBacking {
	readonly issues: GitHubChildIssue[];
	readonly pullRequests: GitHubPullRequestEvidence[];
	readonly rosters: GitHubRosterSnapshot[];
	readonly integrations: ChildIntegrationReceipt[];
	pullRequestReadCalls = 0;

	constructor(seed: PortOnlyReadinessSeed) {
		this.issues = structuredClone([...seed.issues]);
		this.pullRequests = structuredClone([...seed.pullRequests]);
		this.rosters = structuredClone([...seed.rosters]);
		this.integrations = structuredClone([...seed.integrations]);
	}
}

class PortOnlyReadinessTransport implements GitHubOrchestrationTransport {
	readonly #backing: PortOnlyReadinessBacking;

	constructor(backing: PortOnlyReadinessBacking) {
		this.#backing = backing;
	}

	async findChildIssues(
		query: ChildIssueMarkerQuery,
		_context: ExternalCallContext,
	): Promise<AuthoritativeLookup<GitHubChildIssue>> {
		return { items: this.#backing.issues.filter((entry) => entry.marker === query.marker), complete: true };
	}

	async createChildIssue(_request: CreateChildIssueRequest, _context: ExternalCallContext): Promise<never> {
		throw new Error("port-only readiness fixture does not create issues");
	}

	async findPullRequests(
		query: PullRequestMarkerQuery,
		_context: ExternalCallContext,
	): Promise<AuthoritativeLookup<GitHubPullRequestEvidence>> {
		this.#backing.pullRequestReadCalls += 1;
		return { items: this.#backing.pullRequests.filter((entry) => entry.marker === query.marker), complete: true };
	}

	async createPullRequest(_request: CreatePullRequestRequest, _context: ExternalCallContext): Promise<never> {
		throw new Error("port-only readiness fixture does not create pull requests");
	}

	async findParentRosters(
		query: GitHubRosterQuery,
		_context: ExternalCallContext,
	): Promise<AuthoritativeLookup<GitHubRosterSnapshot>> {
		return { items: this.#backing.rosters.filter((entry) => entry.marker === query.marker), complete: true };
	}

	async publishParentRoster(_request: PublishRosterRequest, _context: ExternalCallContext): Promise<never> {
		throw new Error("port-only readiness fixture does not publish rosters");
	}

	async findChildIntegration(
		query: ChildIntegrationQuery,
		_context: ExternalCallContext,
	): Promise<AuthoritativeLookup<ChildIntegrationReceipt>> {
		return {
			items: this.#backing.integrations.filter((entry) => entry.childId === query.childId && entry.marker === query.marker),
			complete: true,
		};
	}

	async integrateChild(_request: IntegrateChildRequest, _context: ExternalCallContext): Promise<never> {
		throw new Error("port-only readiness fixture does not integrate children");
	}

	async proveAncestry(query: GitAncestryQuery, _context: ExternalCallContext): Promise<GitAncestryProof> {
		return {
			schemaVersion: 1,
			authority: "transport",
			...query,
			result: true,
			revision: 1,
			observedAt: "2026-07-21T12:05:00.000Z",
		};
	}
}

class PortOnlyParentReadyJournal implements ParentReadyOperationJournal {
	readonly #prepared = new Map<string, PreparedParentReadyOperation>();
	readonly settlements: ParentReadySettlementRecord[] = [];

	constructor(snapshot?: {
		prepared: readonly PreparedParentReadyOperation[];
		settlements: readonly ParentReadySettlementRecord[];
	}) {
		for (const operation of snapshot?.prepared ?? []) {
			const query = {
				planDigest: operation.planDigest,
				authorizationDigest: operation.authorization.digest,
				mutationIdempotencyKey: operation.mutation.idempotencyKey,
			};
			this.#prepared.set(this.key(query), structuredClone(operation));
		}
		this.settlements.push(...structuredClone(snapshot?.settlements ?? []));
	}

	private key(query: ParentReadyJournalQuery): string {
		return `${query.planDigest}\u0000${query.authorizationDigest}\u0000${query.mutationIdempotencyKey}`;
	}

	async persistPrepared(operation: PreparedParentReadyOperation, _context: ExternalCallContext): Promise<void> {
		const query = {
			planDigest: operation.planDigest,
			authorizationDigest: operation.authorization.digest,
			mutationIdempotencyKey: operation.mutation.idempotencyKey,
		};
		this.#prepared.set(this.key(query), structuredClone(operation));
	}

	async readPrepared(
		query: ParentReadyJournalQuery,
		_context: ExternalCallContext,
	): Promise<PreparedParentReadyOperation | null> {
		return structuredClone(this.#prepared.get(this.key(query)) ?? null);
	}

	async persistSettlement(settlement: ParentReadySettlementRecord, _context: ExternalCallContext): Promise<void> {
		this.settlements.push(githubOrchestratorApi.validateParentReadySettlementRecord(settlement));
	}

	snapshot(): { prepared: PreparedParentReadyOperation[]; settlements: ParentReadySettlementRecord[] } {
		return {
			prepared: structuredClone([...this.#prepared.values()]),
			settlements: structuredClone(this.settlements),
		};
	}
}

class PortOnlyParentReadyAuthority implements ParentReadyDurableAuthorityBoundary {
	readonly #backing: PortOnlyReadinessBacking;
	readonly #journal: ParentReadyOperationJournal;
	#mutationRevision = 0;
	readonly #mutations = new Map<string, { digest: string; value: GitHubPullRequestEvidence; revision: number }>();

	constructor(backing: PortOnlyReadinessBacking, journal: ParentReadyOperationJournal) {
		this.#backing = backing;
		this.#journal = journal;
	}

	async compareConsumeAndMarkParentReady(
		requestValue: MarkParentReadyRequest,
		context: ExternalCallContext,
	): Promise<ParentReadyCompareEffectResult> {
		const request = githubOrchestratorApi.validateMarkParentReadyRequest(requestValue);
		const query = {
			planDigest: request.authorization.planDigest,
			authorizationDigest: request.authorization.digest,
			mutationIdempotencyKey: request.mutation.idempotencyKey,
		};
		const prepared = await this.#journal.readPrepared(query, context);
		if (prepared === null
			|| prepared.authorization.digest !== request.authorization.digest
			|| prepared.mutation.idempotencyKey !== request.mutation.idempotencyKey
			|| prepared.mutation.intentDigest !== request.mutation.intentDigest) {
			return { schemaVersion: 1, kind: "conflict", coordinate: "authorization_state" };
		}
		const existing = this.#mutations.get(request.mutation.idempotencyKey);
		if (existing !== undefined) {
			if (existing.digest !== request.mutation.intentDigest) {
				return { schemaVersion: 1, kind: "conflict", coordinate: "authorization_state" };
			}
			return {
				schemaVersion: 1,
				kind: "applied",
				mutation: {
					schemaVersion: 1,
					idempotencyKey: request.mutation.idempotencyKey,
					intentDigest: request.mutation.intentDigest,
					revision: existing.revision,
					applied: false,
					value: structuredClone(existing.value),
				},
			};
		}
		const index = this.#backing.pullRequests.findIndex((entry) => entry.number === request.pullRequest);
		const current = this.#backing.pullRequests[index];
		if (index < 0 || current === undefined || current.headSha !== request.headSha) {
			return { schemaVersion: 1, kind: "conflict", coordinate: "head" };
		}
		if (!current.draft || current.revision !== request.authorization.pullRequestRevision) {
			return { schemaVersion: 1, kind: "conflict", coordinate: "pull_request_revision" };
		}
		const updated = { ...current, draft: false, revision: current.revision + 1 };
		this.#backing.pullRequests.splice(index, 1, updated);
		const revision = ++this.#mutationRevision;
		this.#mutations.set(request.mutation.idempotencyKey, {
			digest: request.mutation.intentDigest,
			value: structuredClone(updated),
			revision,
		});
		return {
			schemaVersion: 1,
			kind: "applied",
			mutation: {
				schemaVersion: 1,
				idempotencyKey: request.mutation.idempotencyKey,
				intentDigest: request.mutation.intentDigest,
				revision,
				applied: true,
				value: structuredClone(updated),
			},
		};
	}

	async quarantineAndRollbackParentReady(
		request: RollbackParentReadyRequest,
		_context: ExternalCallContext,
	): Promise<DurableMutationResult<GitHubPullRequestEvidence>> {
		const index = this.#backing.pullRequests.findIndex((entry) => entry.number === request.pullRequest);
		if (index < 0) throw new Error("parent pull request missing");
		const current = this.#backing.pullRequests[index];
		const updated = { ...current, draft: true, revision: current.revision + 1 };
		this.#backing.pullRequests.splice(index, 1, updated);
		return {
			schemaVersion: 1,
			idempotencyKey: request.mutation.idempotencyKey,
			intentDigest: request.mutation.intentDigest,
			revision: ++this.#mutationRevision,
			applied: true,
			value: structuredClone(updated),
		};
	}
}

class PortOnlyDelayedParentReadyAuthority implements ParentReadyDurableAuthorityBoundary {
	readonly #delegate: ParentReadyDurableAuthorityBoundary;
	readonly #recoveryStarted: Promise<void>;
	#announceRecovery = (): void => {};
	readonly #recoveryGate: Promise<void>;
	#releaseRecovery = (): void => {};
	rollbackCalls = 0;

	constructor(delegate: ParentReadyDurableAuthorityBoundary) {
		this.#delegate = delegate;
		this.#recoveryStarted = new Promise<void>((resolve) => { this.#announceRecovery = resolve; });
		this.#recoveryGate = new Promise<void>((resolve) => { this.#releaseRecovery = resolve; });
	}

	get recoveryStarted(): Promise<void> {
		return this.#recoveryStarted;
	}

	releaseRecovery(): void {
		this.#releaseRecovery();
	}

	async compareConsumeAndMarkParentReady(
		request: MarkParentReadyRequest,
		context: ExternalCallContext,
	): Promise<ParentReadyCompareEffectResult> {
		const result = await this.#delegate.compareConsumeAndMarkParentReady(request, context);
		await new Promise<void>((resolve) => setTimeout(resolve, 80));
		return result;
	}

	async quarantineAndRollbackParentReady(
		request: RollbackParentReadyRequest,
		context: ExternalCallContext,
	): Promise<DurableMutationResult<GitHubPullRequestEvidence>> {
		this.rollbackCalls += 1;
		const result = await this.#delegate.quarantineAndRollbackParentReady(request, context);
		this.#announceRecovery();
		await this.#recoveryGate;
		return result;
	}
}

function portOnlyContext(): ExternalCallContext {
	return {
		signal: new AbortController().signal,
		deadlineAt: "2026-07-22T00:01:00.000Z",
		acknowledgeAbort: () => {},
	};
}

function portOnlyPolicySource(): RequiredCheckPolicySource {
	return {
		async findRequiredCheckPolicies(query, _context) {
			const policy = cycle3CheckPolicy(query.baseBranch);
			return {
				items: [{
					schemaVersion: 1,
					authority: "controller",
					repository: query.repository,
					baseBranch: query.baseBranch,
					revision: Number(policy.revision),
					digest: String(policy.digest),
					observedAt: "2026-07-21T12:06:00.000Z",
				}],
				complete: true,
			};
		},
	};
}

class PortOnlyDecisionBacking {
	record: HumanDecisionRecord | null = null;
}

class PortOnlyDecisionBroker implements ParentDecisionBroker {
	readonly #backing: PortOnlyDecisionBacking;

	constructor(backing: PortOnlyDecisionBacking) {
		this.#backing = backing;
	}

	async request(request: GitHubDecisionRequest, _context: ExternalCallContext): Promise<HumanDecisionRecord> {
		if (this.#backing.record !== null) return structuredClone(this.#backing.record);
		if (request.gate !== "parent_merge") throw new Error("port-only broker accepts only parent merge decisions");
		const binding = {
			repository: request.repository,
			target: { kind: "pull_request" as const, number: request.pullRequest },
			generation: request.generation,
			...(request.headSha === undefined ? {} : { headSha: request.headSha }),
		};
		const created = createHumanDecisionRecord({
			requestId: request.requestId,
			gate: request.gate,
			binding,
			allowedOptions: request.allowedOptions,
			actorAllowlist: request.actorAllowlist,
			expiresAt: request.expiresAt,
			question: request.question,
		}, new Date("2026-07-21T12:00:00.000Z"));
		const persisted: HumanDecisionRecord = {
			...created,
			requestComment: {
				id: 1,
				url: `https://github.com/${request.repository}/pull/${request.pullRequest}#issuecomment-1`,
				actor: "shepherd-host",
				createdAt: "2026-07-21T12:00:10.000Z",
			},
			updatedAt: "2026-07-21T12:00:10.000Z",
		};
		this.#backing.record = persisted;
		return structuredClone(persisted);
	}

	async poll(
		_requestId: string,
		_binding: HumanDecisionRecord["binding"],
		_options: GitHubDecisionPollOptions,
		_context: ExternalCallContext,
	): Promise<HumanDecisionRecord> {
		const current = this.#backing.record;
		if (current === null) throw new Error("decision request is absent");
		if (current.status === "pending") {
			const decided: HumanDecisionRecord = {
				...current,
				status: "decided",
				decision: approvedDecision,
				updatedAt: approvedDecision.decidedAt,
			};
			this.#backing.record = decided;
			return structuredClone(decided);
		}
		return structuredClone(current);
	}

	async consume(
		_requestId: string,
		_binding: HumanDecisionRecord["binding"],
		_context: ExternalCallContext,
	): Promise<HumanDecisionRecord> {
		const current = this.#backing.record;
		if (current === null || current.status !== "decided" || current.decision === undefined) {
			throw new Error("decision is not consumable");
		}
		this.#backing.record = {
			...current,
			status: "consumed",
			consumedAt: "2026-07-21T12:00:40.000Z",
			updatedAt: "2026-07-21T12:00:40.000Z",
		};
		return structuredClone(this.#backing.record);
	}
}

function portOnlyAttestations(backing: PortOnlyReadinessBacking): AgentSessionAttestationSource {
	return {
		async findAttestations(target, _context) {
			return {
				items: backing.pullRequests
					.filter((pullRequest) => pullRequest.number === target.pullRequest)
					.flatMap((pullRequest) => pullRequest.reviews)
					.filter((review) => review.workItemId === target.workItemId)
					.map(attestReview),
				complete: true,
			};
		},
		async findChangedPathEvidence(target, _context) {
			return {
				items: backing.pullRequests
					.filter((pullRequest) => pullRequest.number === target.pullRequest)
					.map((pullRequest): GitHubChangedPathEvidence => ({
						schemaVersion: 1,
						authority: "controller",
						repository: pullRequest.repository,
						workItemId: pullRequest.workItemId,
						pullRequest: pullRequest.number,
						generation: pullRequest.generation,
						baseSha: pullRequest.baseSha,
						headSha: pullRequest.headSha,
						paths: [...pullRequest.changedPaths],
						complete: true,
						revision: Math.max(1, pullRequest.revision - 1),
						observedAt: "2026-07-21T11:58:00.000Z",
					})),
				complete: true,
			};
		},
	};
}

async function cycle8PortOnlyPreparedScenario(
	wrapAuthority: (authority: ParentReadyDurableAuthorityBoundary) => ParentReadyDurableAuthorityBoundary
		= (authority) => authority,
): Promise<{
	backing: PortOnlyReadinessBacking;
	transport: GitHubOrchestrationTransport;
	journal: ParentReadyOperationJournal;
	journalImplementation: PortOnlyParentReadyJournal;
	authority: ParentReadyDurableAuthorityBoundary;
	orchestrator: GitHubParentOrchestrator;
	candidate: ParentOrchestrationPlan;
	receipts: ChildIntegrationReceipt[];
	operation: PreparedParentReadyOperation;
}> {
	const scenario = await cycle5ReadinessScenario();
	const backing = new PortOnlyReadinessBacking({
		issues: scenario.transport.issues,
		pullRequests: scenario.transport.pullRequests,
		rosters: scenario.transport.rosters,
		integrations: scenario.transport.integrations,
	});
	const transport: GitHubOrchestrationTransport = new PortOnlyReadinessTransport(backing);
	const journalImplementation = new PortOnlyParentReadyJournal();
	const journal: ParentReadyOperationJournal = journalImplementation;
	const authority = wrapAuthority(new PortOnlyParentReadyAuthority(backing, journal));
	const orchestrator = new GitHubParentOrchestrator(
		transport,
		new FakeDecisionBroker(),
		portOnlyAttestations(backing),
		portOnlyPolicySource(),
		{
			externalCallTimeoutMs: 20,
			parentReadyAuthority: authority,
			now: () => new Date("2026-07-22T00:00:00.000Z"),
		},
	);
	const prepared = await orchestrator.prepareParentReadiness(scenario.candidate, scenario.receipts, decisionPolicy);
	if (prepared.kind !== "prepared") throw new Error("port-only readiness did not prepare");
	return {
		backing,
		transport,
		journal,
		journalImplementation,
		authority,
		orchestrator,
		candidate: scenario.candidate,
		receipts: scenario.receipts,
		operation: prepared.operation,
	};
}

test("cycle 7 exposes a journalable prepare and commit boundary for issue 479", async () => {
	const { candidate, transport, receipts } = await cycle5ReadinessScenario();
	const orchestrator = orchestratorFor(transport, new FakeDecisionBroker());
	const prepared = await orchestrator.prepareParentReadiness(candidate, receipts, decisionPolicy);
	assert.equal(prepared.kind, "prepared");
	if (prepared.kind !== "prepared") return;
	const journal: PreparedParentReadyOperation[] = [];
	journal.push(structuredClone(prepared.operation));
	assert.equal(journal[0].decision.status, "consumed");
	assert.equal(journal[0].mutation.operation, "parent_ready");
	const settled = await orchestrator.commitPreparedParentReadiness(candidate, receipts, journal[0]);
	assert.equal(settled.kind, "ready");
	assert.equal(transport.markReadyCalls, 1);
	assert.equal((await orchestrator.stop()).kind, "joined");
});

test("cycle 7 production contract requires atomic authority and exports canonical validators", () => {
	const transport: GitHubOrchestrationTransport = new PortOnlyReadinessTransport(new PortOnlyReadinessBacking({
		issues: [],
		pullRequests: [],
		rosters: [],
		integrations: [],
	}));
	assert.equal("markParentReady" in transport, false);
	assert.equal("rollbackParentReady" in transport, false);
	assert.throws(
		() => Reflect.construct(GitHubParentOrchestrator, [transport, undefined, undefined, undefined, {}]),
		/parent.ready.*authority|authority.*required/i,
	);
	assert.equal(typeof githubOrchestratorApi.validateParentReadyAuthorization, "function");
	assert.equal(typeof githubOrchestratorApi.validateParentReadyCompareEffectResult, "function");
	assert.equal(typeof githubOrchestratorApi.validateRollbackParentReadyRequest, "function");
});

test("cycle 7 issue 479-shaped wiring uses only public production ports", async () => {
	const scenario = await cycle5ReadinessScenario();
	const { candidate, receipts } = scenario;
	const backing = new PortOnlyReadinessBacking({
		issues: scenario.transport.issues,
		pullRequests: scenario.transport.pullRequests,
		rosters: scenario.transport.rosters,
		integrations: scenario.transport.integrations,
	});
	const transport: GitHubOrchestrationTransport = new PortOnlyReadinessTransport(backing);
	const journalImplementation = new PortOnlyParentReadyJournal();
	const journal: ParentReadyOperationJournal = journalImplementation;
	const authority: ParentReadyDurableAuthorityBoundary = new PortOnlyParentReadyAuthority(backing, journal);
	const policySource: RequiredCheckPolicySource = {
		async findRequiredCheckPolicies(query, _context) {
			const policy = cycle3CheckPolicy(query.baseBranch);
			return {
				items: [{
					schemaVersion: 1,
					authority: "controller",
					repository: query.repository,
					baseBranch: query.baseBranch,
					revision: Number(policy.revision),
					digest: String(policy.digest),
					observedAt: "2026-07-21T12:06:00.000Z",
				}],
				complete: true,
			};
		},
	};
	const attestations: AgentSessionAttestationSource = portOnlyAttestations(backing);
	const repository = new MemoryHumanDecisionRepository();
	const decisionRequest: GitHubDecisionRequest = {
		requestId: decisionPolicy.requestId,
		gate: "parent_merge",
		repository: candidate.repository,
		parentIssue: candidate.parentIssue,
		pullRequest: 900,
		generation: candidate.generation,
		headSha: "e".repeat(40),
		allowedOptions: ["approve-merge", "reject"],
		actorAllowlist: [...decisionPolicy.actorAllowlist],
		expiresAt: decisionPolicy.expiresAt,
		question: decisionPolicy.question,
	};
	const decisionBinding = {
		repository: decisionRequest.repository,
		target: { kind: "pull_request" as const, number: decisionRequest.pullRequest },
		generation: decisionRequest.generation,
		headSha: decisionRequest.headSha,
	};
	repository.records.set(decisionRequest.requestId, createHumanDecisionRecord({
		requestId: decisionRequest.requestId,
		gate: "parent_merge",
		binding: decisionBinding,
		allowedOptions: ["approve-merge", "reject"],
		actorAllowlist: decisionRequest.actorAllowlist,
		expiresAt: decisionRequest.expiresAt,
		question: decisionRequest.question,
	}, new Date("2026-07-21T12:00:00.000Z")));
	await recordHumanDecisionRequestComment(repository, decisionRequest.requestId, decisionBinding, {
		id: 1,
		url: `https://github.com/${decisionRequest.repository}/pull/${decisionRequest.pullRequest}#issuecomment-1`,
		actor: "shepherd-host",
		createdAt: "2026-07-21T12:00:10.000Z",
	}, new Date("2026-07-21T12:00:10.000Z"));
	await recordHumanDecision(repository, decisionRequest.requestId, decisionBinding, approvedDecision);
	await consumeHumanDecision(
		repository,
		decisionRequest.requestId,
		decisionBinding,
		new Date("2026-07-21T12:00:40.000Z"),
	);
	const githubTransport: GitHubDecisionTransport = {
		async getAuthenticatedActor() { throw new Error("unexpected GitHub actor read"); },
		async listComments() { throw new Error("unexpected GitHub comment read"); },
		async createDecisionRequestComment() { throw new Error("unexpected GitHub comment write"); },
	};
	const broker = githubOrchestratorApi.adaptGitHubDecisionBroker(new GitHubDecisionBroker(
		repository,
		githubTransport,
		{ now: () => new Date("2026-07-22T00:00:00.000Z") },
	));
	const orchestrator = new GitHubParentOrchestrator(
		transport,
		broker,
		attestations,
		policySource,
		{
			externalCallTimeoutMs: 25,
			parentReadyAuthority: authority,
			now: () => new Date("2026-07-22T00:00:00.000Z"),
		},
	);
	const prepared = await orchestrator.prepareParentReadiness(candidate, receipts, decisionPolicy);
	assert.equal(prepared.kind, "prepared");
	if (prepared.kind !== "prepared") return;
	const context = portOnlyContext();
	await journal.persistPrepared(prepared.operation, context);
	const query = {
		planDigest: prepared.operation.planDigest,
		authorizationDigest: prepared.operation.authorization.digest,
		mutationIdempotencyKey: prepared.operation.mutation.idempotencyKey,
	};
	const journaled = await journal.readPrepared(query, context);
	assert.ok(journaled);
	assert.equal(journaled.decision.status, "consumed");
	const settled = await orchestrator.commitPreparedParentReadiness(candidate, receipts, journaled);
	assert.equal(settled.kind, "ready");
	await journal.persistSettlement({
		schemaVersion: 1,
		...query,
		outcome: "ready",
		settledAt: "2026-07-22T00:00:00.000Z",
	}, context);
	assert.equal(journalImplementation.settlements.length, 1);
	assert.equal((await orchestrator.stop()).kind, "joined");
});

test("cycle 8 issue 479 composition keeps every production role exact and public", async (t) => {
	await t.test("transport returns an exact GitAncestryProof", async () => {
		const backing = new PortOnlyReadinessBacking({ issues: [], pullRequests: [], rosters: [], integrations: [] });
		const transport: GitHubOrchestrationTransport = new PortOnlyReadinessTransport(backing);
		const proof: GitAncestryProof = await transport.proveAncestry({
			repository: "polymetrics-ai/cli",
			ancestorSha: "a".repeat(40),
			descendantSha: "b".repeat(40),
		}, portOnlyContext());
		assert.equal(proof.result, true);
		assert.equal(proof.authority, "transport");
	});

	await t.test("typed applied success crosses transport authority and journal roles", async () => {
		const scenario = await cycle8PortOnlyPreparedScenario();
		await scenario.journal.persistPrepared(scenario.operation, portOnlyContext());
		const result = await scenario.orchestrator.commitPreparedParentReadiness(
			scenario.candidate,
			scenario.receipts,
			scenario.operation,
		);
		assert.equal(result.kind, "ready");
		assert.equal(scenario.backing.pullRequests.find((entry) => entry.number === 900)?.draft, false);
	});

	await t.test("typed non-applied conflict remains explicit", async () => {
		const scenario = await cycle8PortOnlyPreparedScenario();
		const result = await scenario.orchestrator.commitPreparedParentReadiness(
			scenario.candidate,
			scenario.receipts,
			scenario.operation,
		);
		assert.deepEqual(result, {
			kind: "blocked",
			blockers: ["parent_ready_authority_conflict:authorization_state"],
		});
		assert.equal(scenario.backing.pullRequests.find((entry) => entry.number === 900)?.draft, true);
	});

	await t.test("uncertain applied result invokes the exact rollback result contract", async () => {
		let delayed: PortOnlyDelayedParentReadyAuthority | undefined;
		const scenario = await cycle8PortOnlyPreparedScenario((authority) => {
			delayed = new PortOnlyDelayedParentReadyAuthority(authority);
			return delayed;
		});
		await scenario.journal.persistPrepared(scenario.operation, portOnlyContext());
		await assert.rejects(
			scenario.orchestrator.commitPreparedParentReadiness(
				scenario.candidate,
				scenario.receipts,
				scenario.operation,
			),
			/timeout|external|ready/i,
		);
		assert.ok(delayed);
		assert.notEqual(await settleWithin(delayed.recoveryStarted, 100), "hung");
		delayed.releaseRecovery();
		await new Promise<void>((resolve) => setTimeout(resolve, 100));
		assert.equal(delayed.rollbackCalls, 1);
		assert.equal(scenario.backing.pullRequests.find((entry) => entry.number === 900)?.draft, true);
	});

	await t.test("live recovery reports stop incomplete and then joined", async () => {
		let delayed: PortOnlyDelayedParentReadyAuthority | undefined;
		const scenario = await cycle8PortOnlyPreparedScenario((authority) => {
			delayed = new PortOnlyDelayedParentReadyAuthority(authority);
			return delayed;
		});
		await scenario.journal.persistPrepared(scenario.operation, portOnlyContext());
		await assert.rejects(scenario.orchestrator.commitPreparedParentReadiness(
			scenario.candidate,
			scenario.receipts,
			scenario.operation,
		));
		assert.ok(delayed);
		assert.notEqual(await settleWithin(delayed.recoveryStarted, 100), "hung");
		assert.equal((await scenario.orchestrator.stop({ deadlineAt: new Date(Date.now() + 10).toISOString() })).kind, "incomplete");
		delayed.releaseRecovery();
		await new Promise<void>((resolve) => setTimeout(resolve, 100));
		assert.equal((await scenario.orchestrator.stop()).kind, "joined");
	});

	await t.test("durable settlement uses the public journal record", async () => {
		const scenario = await cycle8PortOnlyPreparedScenario();
		const query: ParentReadyJournalQuery = {
			planDigest: scenario.operation.planDigest,
			authorizationDigest: scenario.operation.authorization.digest,
			mutationIdempotencyKey: scenario.operation.mutation.idempotencyKey,
		};
		await scenario.journal.persistPrepared(scenario.operation, portOnlyContext());
		await scenario.journal.persistSettlement({
			schemaVersion: 1,
			...query,
			outcome: "blocked",
			settledAt: "2026-07-22T00:00:00.000Z",
		}, portOnlyContext());
		assert.equal(scenario.journalImplementation.settlements.length, 1);
		assert.deepEqual(scenario.journalImplementation.settlements[0], {
			schemaVersion: 1,
			...query,
			outcome: "blocked",
			settledAt: "2026-07-22T00:00:00.000Z",
		});
	});
});

test("cycle 8 reconstructs every readiness role from serialized durable state", async (t) => {
	const scenario = await cycle5ReadinessScenario();
	const originalBacking = new PortOnlyReadinessBacking({
		issues: scenario.transport.issues,
		pullRequests: scenario.transport.pullRequests,
		rosters: scenario.transport.rosters,
		integrations: scenario.transport.integrations,
	});
	const originalTransport: GitHubOrchestrationTransport = new PortOnlyReadinessTransport(originalBacking);
	const originalJournalImplementation = new PortOnlyParentReadyJournal();
	const originalJournal: ParentReadyOperationJournal = originalJournalImplementation;
	const originalAuthorityAdapter = new PortOnlyDelayedParentReadyAuthority(
		new PortOnlyParentReadyAuthority(originalBacking, originalJournal),
	);
	const decisionBacking = new PortOnlyDecisionBacking();
	const originalBroker: ParentDecisionBroker = new PortOnlyDecisionBroker(decisionBacking);
	const originalController = new GitHubParentOrchestrator(
		originalTransport,
		originalBroker,
		portOnlyAttestations(originalBacking),
		portOnlyPolicySource(),
		{
			externalCallTimeoutMs: 20,
			parentReadyAuthority: originalAuthorityAdapter,
			now: () => new Date("2026-07-22T00:00:00.000Z"),
		},
	);
	const prepared = await originalController.prepareParentReadiness(scenario.candidate, scenario.receipts, decisionPolicy);
	assert.equal(prepared.kind, "prepared");
	if (prepared.kind !== "prepared") throw new Error("restart fixture did not prepare");
	await originalJournal.persistPrepared(prepared.operation, portOnlyContext());
	await assert.rejects(originalController.commitPreparedParentReadiness(
		scenario.candidate,
		scenario.receipts,
		prepared.operation,
	));
	assert.notEqual(await settleWithin(originalAuthorityAdapter.recoveryStarted, 100), "hung");
	assert.equal((await originalController.stop({ deadlineAt: new Date(Date.now() + 10).toISOString() })).kind, "incomplete");

	const serialized = JSON.stringify({
		readiness: {
			issues: originalBacking.issues,
			pullRequests: originalBacking.pullRequests,
			rosters: originalBacking.rosters,
			integrations: originalBacking.integrations,
		},
		journal: originalJournalImplementation.snapshot(),
	});
	const decoded = JSON.parse(serialized);
	const restartedBacking = new PortOnlyReadinessBacking(decoded.readiness);
	const restartedJournalImplementation = new PortOnlyParentReadyJournal(decoded.journal);
	const restartedJournal: ParentReadyOperationJournal = restartedJournalImplementation;
	const restartedTransport: GitHubOrchestrationTransport = new PortOnlyReadinessTransport(restartedBacking);
	const restartedAuthority: ParentReadyDurableAuthorityBoundary = new PortOnlyParentReadyAuthority(
		restartedBacking,
		restartedJournal,
	);
	const restartedBroker: ParentDecisionBroker = new PortOnlyDecisionBroker(decisionBacking);
	const restartedController = new GitHubParentOrchestrator(
		restartedTransport,
		restartedBroker,
		portOnlyAttestations(restartedBacking),
		portOnlyPolicySource(),
		{
			externalCallTimeoutMs: 20,
			parentReadyAuthority: restartedAuthority,
			now: () => new Date("2026-07-22T00:00:00.000Z"),
		},
	);

	await t.test("prepared operation and journal survive value serialization", async () => {
		const restored = await restartedJournal.readPrepared({
			planDigest: prepared.operation.planDigest,
			authorizationDigest: prepared.operation.authorization.digest,
			mutationIdempotencyKey: prepared.operation.mutation.idempotencyKey,
		}, portOnlyContext());
		assert.deepEqual(restored, prepared.operation);
	});

	await t.test("controller broker journal transport and authority adapters are reconstructed", () => {
		assert.notEqual(restartedController, originalController);
		assert.notEqual(restartedBroker, originalBroker);
		assert.notEqual(restartedJournal, originalJournal);
		assert.notEqual(restartedTransport, originalTransport);
		assert.notEqual(restartedAuthority, originalAuthorityAdapter);
	});

	await t.test("reconstructed roles reconcile the uncertain draft and resume once", async () => {
		assert.equal(restartedBacking.pullRequests.find((entry) => entry.number === 900)?.draft, true);
		const resumed = await restartedController.prepareParentReadiness(scenario.candidate, scenario.receipts, decisionPolicy);
		assert.equal(resumed.kind, "prepared");
		if (resumed.kind !== "prepared") throw new Error("restart did not produce a resumable authorization");
		await restartedJournal.persistPrepared(resumed.operation, portOnlyContext());
		const committed = await restartedController.commitPreparedParentReadiness(
			scenario.candidate,
			scenario.receipts,
			resumed.operation,
		);
		assert.equal(committed.kind, "ready");
		const replayed = await restartedController.commitPreparedParentReadiness(
			scenario.candidate,
			scenario.receipts,
			resumed.operation,
		);
		assert.equal(replayed.kind, "ready");
		assert.equal(restartedBacking.pullRequests.find((entry) => entry.number === 900)?.draft, false);
	});

	await t.test("cross-instance truth has no module WeakMap or authority-object identity", async () => {
		const production = await readFile(".pi/extensions/shepherd/github-orchestrator.ts", "utf8");
		assert.doesNotMatch(production, /\bWeakMap\b/u);
		assert.equal((await restartedController.stop()).kind, "joined");
	});

	originalAuthorityAdapter.releaseRecovery();
	await new Promise<void>((resolve) => setTimeout(resolve, 100));
	assert.equal((await originalController.stop()).kind, "joined");
});

test("cycle 7 atomic authority boundary rejects every moved coordinate without recovery", async (t) => {
	const mutations: Array<[string, ParentReadyAuthorityCoordinate, (authorization: Record<string, any>) => void]> = [
		["policy", "policy", (value) => { value.policies[0].revision += 1; }],
		["review", "review", (value) => { value.reviewAuthorizationDigest = "f".repeat(64); }],
		["exact paths", "exact_paths", (value) => { value.exactPaths = [...value.exactPaths, "cycle-7/moved.ts"]; }],
		["child receipt", "child_receipt", (value) => { value.children[0].receipt.integratedAt = "2026-07-21T12:11:00.000Z"; }],
		["ancestry", "ancestry", (value) => { value.children[0].ancestry.descendantSha = "f".repeat(40); }],
		["decision", "decision", (value) => {
			value.decision.consumedAt = "2026-07-21T12:00:41.000Z";
			value.decision.updatedAt = "2026-07-21T12:00:41.000Z";
		}],
		["plan", "plan", (value) => { value.planDigest = "f".repeat(64); }],
		["head", "head", (value) => { value.headSha = "f".repeat(40); }],
		["PR revision", "pull_request_revision", (value) => { value.pullRequestRevision += 1; }],
		["durable authorization state", "authorization_state", () => {}],
	];
	for (const [name, coordinate, mutate] of mutations) {
		await t.test(name, async () => {
			const scenario = await cycle5ReadinessScenario();
			const { candidate, receipts } = scenario;
			const backing = new PortOnlyReadinessBacking({
				issues: scenario.transport.issues,
				pullRequests: scenario.transport.pullRequests,
				rosters: scenario.transport.rosters,
				integrations: scenario.transport.integrations,
			});
			const implementation = new PortOnlyReadinessTransport(backing);
			const transport: GitHubOrchestrationTransport = implementation;
			let current: ParentReadyAuthorization | undefined;
			let quarantined = false;
			let readsAtCompare = 0;
			let rollbackCalls = 0;
			const authority: ParentReadyDurableAuthorityBoundary = {
				async compareConsumeAndMarkParentReady(request: MarkParentReadyRequest) {
					readsAtCompare = backing.pullRequestReadCalls;
					if (quarantined || current?.digest !== request.authorization.digest) {
						return { schemaVersion: 1, kind: "conflict", coordinate };
					}
					throw new Error("atomic movement case unexpectedly retained current authority");
				},
				async quarantineAndRollbackParentReady() {
					rollbackCalls += 1;
					throw new Error("rollback is unavailable and must not be needed for an authority conflict");
				},
			};
			const orchestrator = new GitHubParentOrchestrator(
				transport,
				new FakeDecisionBroker(),
				portOnlyAttestations(backing),
				portOnlyPolicySource(),
				{
					externalCallTimeoutMs: 25,
					parentReadyAuthority: authority,
					now: () => new Date("2026-07-22T00:00:00.000Z"),
				},
			);
			const prepared = await orchestrator.prepareParentReadiness(candidate, receipts, decisionPolicy);
			assert.equal(prepared.kind, "prepared");
			if (prepared.kind !== "prepared") return;
			const moved = structuredClone(prepared.operation.authorization) as unknown as Record<string, any>;
			mutate(moved);
			if (name === "durable authorization state") {
				current = prepared.operation.authorization;
				quarantined = true;
			} else {
				const { digest: _digest, ...payload } = moved;
				current = githubOrchestratorApi.canonicalizeParentReadyAuthorization(payload);
			}
			const result = await orchestrator.commitPreparedParentReadiness(candidate, receipts, prepared.operation);
			assert.deepEqual(result, {
				kind: "blocked",
				blockers: [`parent_ready_authority_conflict:${coordinate}`],
			});
			assert.equal(rollbackCalls, 0);
			assert.equal(backing.pullRequestReadCalls, readsAtCompare, "conflict returns without recovery reads");
			assert.equal(backing.pullRequests.find((entry) => entry.number === 900)?.draft, true);
		});
	}
});

test("cycle 7 run-state schema has one current HEAD semantic and rejects historical SHA reuse", async () => {
	const raw = await readFile(".planning/phases/478-shepherd-github-parent-orchestration/RUN-STATE.json", "utf8");
	const state = JSON.parse(raw) as Record<string, unknown>;
	const details = state.details as Record<string, unknown>;
	assert.equal(details.candidateRef, "HEAD");
	assert.equal(Object.hasOwn(details, "frozenCandidate"), false);
	assert.equal((details.priorReviewState as Record<string, unknown>).priorReviewedCandidate,
		"b90037df1fff38c755ebc8025579120d17031330");
	assert.equal((details.priorReviewState as Record<string, unknown>).findingsConsolidatedIntoCycle, 8);
	const validate = githubOrchestratorApi.validateRunStateCandidateSemantics;
	assert.doesNotThrow(() => validate(state));
	const invalid = structuredClone(state);
	(invalid.details as Record<string, unknown>).candidateRef = "dbce5b7d0c698bc802594211072fed77eff23c1c";
	assert.throws(() => validate(invalid), /candidate|HEAD|historical|current/i);
});
