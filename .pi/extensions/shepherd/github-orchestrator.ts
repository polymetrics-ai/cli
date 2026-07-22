import { createHash } from "node:crypto";
import { AsyncLocalStorage } from "node:async_hooks";
import { types as nodeTypes } from "node:util";

import {
	DependencyGraphError,
	selectReadyWork,
	validateDependencyGraph,
	type DependencyWorkItem,
	type ReadyQueueSelection,
	type WorkAccess,
	type WorkItemStatus,
} from "./dependency-graph.ts";
import {
	canonicalIssueBranch,
	type GitReviewedChildIntegrationEvidence,
} from "./git-adapter.ts";
import {
	canonicalGitRef,
	evaluateGitHubPullRequestEvidence,
	validateGitHubChangedPathEvidence,
	validateGitHubPullRequestEvidence,
	validateRequiredGitHubCheckPolicyObservation,
	type GitHubEvidenceBlocker,
	type GitHubChangedPathEvidence,
	type GitHubPullRequestEvidence,
	type RequiredGitHubCheckPolicy,
	type RequiredGitHubCheckPolicyObservation,
	validateRequiredGitHubCheckPolicy,
} from "./github-evidence.ts";
import {
	GitHubDecisionBroker,
	type GitHubDecisionPollOptions,
	type GitHubDecisionPollResult,
	type GitHubDecisionRequest,
} from "./github-decision-broker.ts";
import {
	assertHumanDecisionBinding,
	validateHumanDecisionRecord,
	type HumanDecisionBinding,
	type HumanDecisionGate,
	type HumanDecisionRecord,
} from "./human-decision.ts";
import { reconcileAutonomy } from "./reconciler.ts";
import {
	assertNoSensitiveText,
	independentReviewAuthorizationDigest,
	independentReviewResultDigest,
	readBoundedExactRecord,
	validateAgentSessionAttestation,
	type AgentSessionAttestation,
	type IndependentReviewRecord,
	type IndependentReviewTarget,
} from "./review-router.ts";
import type {
	ClaimedWorkspace,
	WorkspaceHandoffEvidence,
} from "./workspace-adapter.ts";

const MAX_CHILDREN = 64;
const MAX_LIST = 64;
const MAX_BODY_BYTES = 65_536;
const MAX_GITHUB_NUMBER = 2_147_483_647;
const MAX_CANONICAL_PLAN_BYTES = 1_000_000;
const SHA = /^[0-9a-f]{40}$/;
const IDENTITY = /^[0-9a-f]{64}$/;
const CHILD_ID = /^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$/;
const SLUG = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
const SKILL = /^[A-Za-z0-9][A-Za-z0-9:._-]{0,127}$/;
const REPOSITORY = /^[A-Za-z0-9][A-Za-z0-9._-]{0,99}\/[A-Za-z0-9][A-Za-z0-9._-]{0,99}$/;
const RFC3339_UTC = /^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(?:\.(\d{1,3}))?Z$/;
const UNSAFE_INLINE = /[\u0000-\u001f\u007f-\u009f\u00ad\u061c\u200b-\u200f\u2028-\u202e\u2060-\u206f\ufeff]/u;
const UNSAFE_BODY = /[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f-\u009f\u00ad\u061c\u200b-\u200f\u2028-\u202e\u2060-\u206f\ufeff]/u;
const abortSignalAbortedGetter = Object.getOwnPropertyDescriptor(AbortSignal.prototype, "aborted")?.get;
const eventTargetAddEventListener = EventTarget.prototype.addEventListener;
const eventTargetRemoveEventListener = EventTarget.prototype.removeEventListener;

function intrinsicSignalAborted(value: AbortSignal): boolean {
	if (abortSignalAbortedGetter === undefined) throw new Error("AbortSignal intrinsic is unavailable");
	try {
		return abortSignalAbortedGetter.call(value) as boolean;
	} catch {
		throw new Error("invalid orchestration caller AbortSignal");
	}
}

function canonicalAbortSignal(value: unknown): AbortSignal {
	if (typeof value !== "object" || value === null || nodeTypes.isProxy(value)) {
		throw new Error("invalid orchestration caller AbortSignal");
	}
	const signal = value as AbortSignal;
	intrinsicSignalAborted(signal);
	try {
		const probe = (): void => {};
		eventTargetAddEventListener.call(signal, "abort", probe, { once: true });
		eventTargetRemoveEventListener.call(signal, "abort", probe);
	} catch {
		throw new Error("invalid orchestration caller AbortSignal");
	}
	return signal;
}

function leaseAbortSignal(signal: AbortSignal, listener: () => void): () => void {
	eventTargetAddEventListener.call(signal, "abort", listener, { once: true });
	if (intrinsicSignalAborted(signal)) listener();
	return () => eventTargetRemoveEventListener.call(signal, "abort", listener);
}

export type VerificationRequirementKind = "test" | "typecheck" | "offline_rpc" | "diff_scope";

export interface VerificationRequirement {
	id: string;
	kind: VerificationRequirementKind;
	description: string;
}

export interface PlannedBranch {
	kind: "canonical_issue_branch";
	slug: string;
}

export interface ChildIdempotencyMarkers {
	issue: string;
	pullRequest: string;
}

export interface BoundedChildRecord extends DependencyWorkItem {
	title: string;
	objective: string;
	branch: PlannedBranch;
	prBase: string;
	requiredSkills: string[];
	verification: VerificationRequirement[];
	humanGates: HumanDecisionGate[];
	markers: ChildIdempotencyMarkers;
	issueBody: string;
}

export interface MaterializedChildRecord extends Omit<BoundedChildRecord, "branch"> {
	issue: number;
	branch: string;
}

export interface ParentOrchestrationPlan {
	schemaVersion: 1;
	repository: string;
	parentIssue: number;
	generation: number;
	title: string;
	objective: string;
	parentBranch: string;
	parentBaseBranch: string;
	markers: {
		parentPullRequest: string;
		roster: string;
	};
	requiredCheckPolicies: RequiredGitHubCheckPolicy[];
	children: BoundedChildRecord[];
	canonical: {
		schemaVersion: 1;
		serialized: string;
		digest: string;
	};
}

export interface ParentOrchestrationPolicyBundle {
	schemaVersion: 1;
	requiredCheckPolicies: readonly RequiredGitHubCheckPolicy[];
}

export interface ParentOrchestrationPolicyAuthority {
	schemaVersion: 1;
	authority: "controller";
	repository: string;
	parentIssue: number;
	generation: number;
	parentBranch: string;
	parentBaseBranch: string;
	revision: number;
	observedAt: string;
	policyBundle: ParentOrchestrationPolicyBundle;
}

export type DurableMutationOperation = "child_issue" | "pull_request" | "parent_roster" | "child_integration" | "parent_ready" | "parent_ready_rollback";

export interface DurableMutationIntent {
	schemaVersion: 1;
	operation: DurableMutationOperation;
	idempotencyKey: string;
	intentDigest: string;
	expectedResourceRevision: number | null;
}

export interface DurableMutationResult<T> {
	schemaVersion: 1;
	idempotencyKey: string;
	intentDigest: string;
	revision: number;
	applied: boolean;
	value: T;
}

export interface GitHubChildIssue {
	number: number;
	marker: string;
	title: string;
	body: string;
	state: "open" | "closed";
	parentIssue: number;
}

export interface ChildIssueMarkerQuery {
	repository: string;
	marker: string;
}

export interface CreateChildIssueRequest extends ChildIssueMarkerQuery {
	parentIssue: number;
	title: string;
	body: string;
	mutation: DurableMutationIntent;
}

export interface PullRequestMarkerQuery {
	repository: string;
	marker: string;
}

export interface CreatePullRequestRequest extends PullRequestMarkerQuery {
	workItemId: string;
	generation: number;
	title: string;
	body: string;
	draft: boolean;
	baseBranch: string;
	headBranch: string;
	baseSha: string;
	headSha: string;
	changedPaths: readonly string[];
	allowedScopes: readonly string[];
	policyDigest: string;
	mutation: DurableMutationIntent;
}

export interface GitHubRosterQuery {
	repository: string;
	marker: string;
}

export interface GitHubRosterSnapshot {
	id: number;
	marker: string;
	parentIssue: number;
	generation: number;
	body: string;
	statuses: Record<string, WorkItemStatus>;
	statusEpoch: number;
	revision: number;
	updatedAt: string;
}

export interface PublishRosterRequest extends GitHubRosterQuery {
	parentIssue: number;
	generation: number;
	body: string;
	statuses: Readonly<Record<string, WorkItemStatus>>;
	statusEpoch: number;
	mutation: DurableMutationIntent;
}

export interface GitHubPullRequestQuery {
	repository: string;
	pullRequest: number;
}

export interface AuthoritativeLookup<T> {
	items: T[];
	complete: boolean;
}

export interface ChildIntegrationQuery {
	repository: string;
	childId: string;
	marker: string;
}

export interface GitAncestryQuery {
	repository: string;
	ancestorSha: string;
	descendantSha: string;
}

export interface GitAncestryProof extends GitAncestryQuery {
	schemaVersion: 1;
	authority: "transport";
	result: boolean;
	revision: number;
	observedAt: string;
}

export interface CanonicalPullRequestSnapshot {
	repository: string;
	workItemId: string;
	number: number;
	generation: number;
	marker: string;
	baseBranch: string;
	headBranch: string;
	baseSha: string;
	headSha: string;
	changedPaths: string[];
	allowedScopes: string[];
	policyDigest: string;
	revision: number;
	observedAt: string;
	digest: string;
}

export interface PullRequestObservation {
	revision: number;
	observedAt: string;
	state: GitHubPullRequestEvidence["state"];
}

export interface ControllerIntegrationProvenance {
	authority: "controller";
	planDigest: string;
	policyDigest: string;
	policyRevision: number;
	policyObservedAt: string;
	changedPathDigest: string;
	reviewAuthorizationDigest: string;
	reviewResultDigest: string;
	reviewCompletedAt: string;
	evidenceRevision: number;
	observedAt: string;
}

export interface TransportMutationProvenance {
	authority: "transport";
	idempotencyKey: string;
	intentDigest: string;
	revision: number;
}

export interface ChildIntegrationReceipt {
	childId: string;
	pullRequest: number;
	generation: number;
	marker: string;
	baseSha: string;
	headSha: string;
	parentBranch: string;
	pullRequestSnapshot: CanonicalPullRequestSnapshot;
	observation: PullRequestObservation;
	controllerProvenance: ControllerIntegrationProvenance;
	transportProvenance: TransportMutationProvenance;
	integratedAt: string;
}

export interface IntegrateChildRequest extends GitHubPullRequestQuery {
	childId: string;
	generation: number;
	marker: string;
	baseSha: string;
	headSha: string;
	parentBranch: string;
	/** Plan-bound authoritative default/base branch that child integration must never target. */
	parentBaseBranch: string;
	/** Exact lease-bound Git CAS evidence; the transport may publish but cannot merge. */
	integration: GitReviewedChildIntegrationEvidence;
	pullRequestSnapshot: CanonicalPullRequestSnapshot;
	observation: PullRequestObservation;
	controllerProvenance: ControllerIntegrationProvenance;
	mutation: DurableMutationIntent;
}

export interface ChildIntegrationMutationAuthorization {
	repository: string;
	childId: string;
	pullRequest: number;
	generation: number;
	marker: string;
	parentBranch: string;
	parentBaseBranch: string;
	baseSha: string;
	headSha: string;
	idempotencyKey: string;
	intentDigest: string;
}

export type ChildIntegrationMutation = (
	authorization: ChildIntegrationMutationAuthorization,
) => Promise<GitReviewedChildIntegrationEvidence>;

export interface MarkParentReadyRequest extends GitHubPullRequestQuery {
	marker: string;
	headSha: string;
	generation: number;
	decisionRequestId: string;
	authorization: ParentReadyAuthorization;
	freshness: ParentReadyFreshnessEnvelope;
	mutation: DurableMutationIntent;
}

export interface ParentReadyPolicyAuthority {
	schemaVersion: 1;
	authority: "controller";
	repository: string;
	baseBranch: string;
	revision: number;
	digest: string;
}

export interface ParentReadyAncestryAuthority {
	schemaVersion: 1;
	authority: "transport";
	repository: string;
	ancestorSha: string;
	descendantSha: string;
	result: true;
}

export interface ParentReadyChildAuthority {
	receipt: ChildIntegrationReceipt;
	ancestry: ParentReadyAncestryAuthority;
}

export interface ParentReadyAuthorization {
	schemaVersion: 1;
	repository: string;
	parentIssue: number;
	generation: number;
	pullRequest: number;
	decisionRequestId: string;
	decisionDigest: string;
	childRosterDigest: string;
	policySetDigest: string;
	parentReviewDigest: string;
	parentPathDigest: string;
	policies: ParentReadyPolicyAuthority[];
	reviewAuthorizationDigest: string;
	exactPaths: string[];
	children: ParentReadyChildAuthority[];
	decision: HumanDecisionRecord;
	planDigest: string;
	headSha: string;
	pullRequestRevision: number;
	digest: string;
}

export interface ParentReadyFreshnessEnvelope {
	schemaVersion: 1;
	policyObservations: RequiredGitHubCheckPolicyObservation[];
	parentReviewResultDigest: string;
	parentReviewCompletedAt: string;
	parentPathRevision: number;
	parentPathObservedAt: string;
	childAncestryProofs: GitAncestryProof[];
	digest: string;
}

export interface PreparedParentReadyOperation {
	schemaVersion: 1;
	planDigest: string;
	policy: ParentDecisionPolicy;
	decision: HumanDecisionRecord;
	authorization: ParentReadyAuthorization;
	freshness: ParentReadyFreshnessEnvelope;
	mutation: DurableMutationIntent;
}

export const PARENT_READY_AUTHORITY_COORDINATES = [
	"policy",
	"review",
	"exact_paths",
	"child_receipt",
	"ancestry",
	"decision",
	"plan",
	"head",
	"pull_request_revision",
	"authorization_state",
] as const;

export type ParentReadyAuthorityCoordinate = typeof PARENT_READY_AUTHORITY_COORDINATES[number];

export interface ParentReadyConflictTombstone extends ParentReadyAuthorityQuery {
	schemaVersion: 1;
	kind: "tombstoned";
	invocationId: string;
	authorizationDigest: string;
	mutationIdempotencyKey: string;
	mutationIntentDigest: string;
}

export type ParentReadyCompareEffectResult =
	| {
		schemaVersion: 1;
		kind: "applied";
		mutation: DurableMutationResult<GitHubPullRequestEvidence>;
	}
	| {
		schemaVersion: 1;
		kind: "conflict";
		coordinate: ParentReadyAuthorityCoordinate;
		terminal: ParentReadyConflictTombstone;
	};

export interface ParentReadyJournalQuery {
	planDigest: string;
	authorizationDigest: string;
	mutationIdempotencyKey: string;
}

export interface ParentReadySettlementRecord extends ParentReadyJournalQuery {
	schemaVersion: 1;
	outcome: "ready" | "blocked";
	settledAt: string;
}

export interface ParentReadyRecoveryFence {
	schemaVersion: 1;
	invocationId: string;
	recoveryId: string;
	attempt: number;
	supersedesAttempt: number | null;
	readyMutation: DurableMutationIntent;
}

export interface ParentReadyAuthorityQuery extends GitHubPullRequestQuery {
	marker: string;
	headSha: string;
	generation: number;
}

export type ParentReadyAuthorityPhase =
	| "ready_invoking"
	| "ready_effect_applied"
	| "ready_settled"
	| "recovery_claimed"
	| "draft_restored";

export interface ParentReadyAuthorityState extends ParentReadyAuthorityQuery {
	schemaVersion: 1;
	invocationId: string;
	recoveryId: string;
	originalRevision: number;
	appliedRevision: number | null;
	authorization: ParentReadyAuthorization;
	readyMutation: DurableMutationIntent;
	rollbackMutation: DurableMutationIntent | null;
	phase: ParentReadyAuthorityPhase;
	status: "unsettled" | "settled";
	fence: number;
}

export interface SettleParentReadyAuthorityRequest extends ParentReadyAuthorityQuery {
	invocationId: string;
	authorizationDigest: string;
	readyMutation: DurableMutationIntent;
	expectedPhase: "ready_effect_applied";
	expectedFence: 0;
}

export interface RollbackParentReadyRequest extends GitHubPullRequestQuery {
	marker: string;
	headSha: string;
	generation: number;
	authorizationDigest: string;
	recovery: ParentReadyRecoveryFence;
	mutation: DurableMutationIntent;
}

export interface ExternalCallContext {
	signal: AbortSignal;
	deadlineAt: string;
	acknowledgeAbort(): void;
}

export interface OrchestrationCallContext {
	signal?: AbortSignal;
	deadlineAt?: string;
}

export interface GitHubOrchestrationTransport {
	findChildIssues(query: ChildIssueMarkerQuery, context: ExternalCallContext): Promise<AuthoritativeLookup<GitHubChildIssue>>;
	createChildIssue(request: CreateChildIssueRequest, context: ExternalCallContext): Promise<DurableMutationResult<GitHubChildIssue>>;
	findPullRequests(query: PullRequestMarkerQuery, context: ExternalCallContext): Promise<AuthoritativeLookup<GitHubPullRequestEvidence>>;
	createPullRequest(request: CreatePullRequestRequest, context: ExternalCallContext): Promise<DurableMutationResult<GitHubPullRequestEvidence>>;
	findParentRosters(query: GitHubRosterQuery, context: ExternalCallContext): Promise<AuthoritativeLookup<GitHubRosterSnapshot>>;
	publishParentRoster(request: PublishRosterRequest, context: ExternalCallContext): Promise<DurableMutationResult<GitHubRosterSnapshot>>;
	findChildIntegration(query: ChildIntegrationQuery, context: ExternalCallContext): Promise<AuthoritativeLookup<ChildIntegrationReceipt>>;
	integrateChild(request: IntegrateChildRequest, context: ExternalCallContext): Promise<DurableMutationResult<ChildIntegrationReceipt>>;
	proveAncestry(query: GitAncestryQuery, context: ExternalCallContext): Promise<GitAncestryProof>;
}

export interface ParentReadyDurableAuthorityBoundary {
	/** Read the canonical durable state for one exact parent-ready target. */
	readParentReadyState(
		query: ParentReadyAuthorityQuery,
		context: ExternalCallContext,
	): Promise<ParentReadyAuthorityState | null>;
	/** Durably establish the exact invocation before any uncertain pull-request effect is called. */
	beginParentReady(
		request: MarkParentReadyRequest,
		context: ExternalCallContext,
	): Promise<ParentReadyAuthorityState>;
	/**
	 * Atomically compare the established invocation and pull-request revision, consume the matching
	 * authority, clear draft, and persist the exact applied resource revision. A conflict is
	 * explicitly non-applied.
	 */
	compareConsumeAndMarkParentReady(
		request: MarkParentReadyRequest,
		context: ExternalCallContext,
	): Promise<ParentReadyCompareEffectResult>;
	/** Settle a validated applied response. Only this transition may authorize a ready result. */
	settleParentReady(
		request: SettleParentReadyAuthorityRequest,
		context: ExternalCallContext,
	): Promise<ParentReadyAuthorityState>;
	/**
	 * Durably claim the ordered recovery attempt before returning. Claiming a newer attempt fences
	 * every predecessor from applying or settling. The only permitted effect is an idempotent
	 * restoration of the exact draft owned by recovery.readyMutation.
	 */
	quarantineAndRollbackParentReady(
		request: RollbackParentReadyRequest,
		context: ExternalCallContext,
	): Promise<DurableMutationResult<GitHubPullRequestEvidence>>;
}

export interface ParentReadyOperationJournal {
	persistPrepared(
		operation: PreparedParentReadyOperation,
		context: ExternalCallContext,
	): Promise<void>;
	readPrepared(
		query: ParentReadyJournalQuery,
		context: ExternalCallContext,
	): Promise<PreparedParentReadyOperation | null>;
	persistSettlement(
		settlement: ParentReadySettlementRecord,
		context: ExternalCallContext,
	): Promise<void>;
}

export interface AgentSessionAttestationSource {
	findAttestations(target: IndependentReviewTarget, context: ExternalCallContext): Promise<AuthoritativeLookup<AgentSessionAttestation>>;
	findChangedPathEvidence(query: Omit<IndependentReviewTarget, "changedPaths" | "allowedScopes">, context: ExternalCallContext): Promise<AuthoritativeLookup<GitHubChangedPathEvidence>>;
}

export interface ParentDecisionBroker {
	request(request: GitHubDecisionRequest, context: ExternalCallContext): Promise<HumanDecisionRecord>;
	poll(
		requestId: string,
		binding: HumanDecisionBinding,
		options: GitHubDecisionPollOptions,
		context: ExternalCallContext,
	): Promise<HumanDecisionRecord>;
	consume(requestId: string, binding: HumanDecisionBinding, context: ExternalCallContext): Promise<HumanDecisionRecord>;
}

function brokerBinding(request: GitHubDecisionRequest): HumanDecisionBinding {
	return {
		repository: request.repository,
		target: { kind: "pull_request", number: request.pullRequest },
		generation: request.generation,
		...(request.headSha === undefined ? {} : { headSha: request.headSha }),
	};
}

function validateActualBrokerPollResult(
	value: unknown,
	record: HumanDecisionRecord,
): GitHubDecisionPollResult {
	const candidate = readBoundedExactRecord(value, ["status", "attempts"], ["decision"], "GitHub decision poll result");
	const attempts = generation(candidate.attempts);
	if (candidate.status === "pending") {
		if (candidate.decision !== undefined || record.status !== "pending") {
			throw new Error("GitHub decision poll pending result disagrees with canonical state");
		}
		return { status: "pending", attempts };
	}
	if (candidate.status === "expired") {
		if (candidate.decision !== undefined || record.status !== "expired") {
			throw new Error("GitHub decision poll expiry result disagrees with canonical state");
		}
		return { status: "expired", attempts };
	}
	if (candidate.status !== "decided" || candidate.decision === undefined
		|| (record.status !== "decided" && record.status !== "consumed") || record.decision === undefined) {
		throw new Error("GitHub decision poll result disagrees with canonical state");
	}
	const evidence = readBoundedExactRecord(
		candidate.decision,
		["option", "actor", "sourceUrl", "decidedAt"],
		[],
		"GitHub decision poll evidence",
	);
	if (evidence.option !== record.decision.option || evidence.actor !== record.decision.actor
		|| evidence.sourceUrl !== record.decision.sourceUrl || evidence.decidedAt !== record.decision.decidedAt) {
		throw new Error("GitHub decision poll evidence disagrees with canonical state");
	}
	return { status: "decided", decision: { ...record.decision }, attempts };
}

async function withActualBrokerContext<T>(context: ExternalCallContext, operation: () => Promise<T>): Promise<T> {
	if (intrinsicSignalAborted(context.signal)) {
		context.acknowledgeAbort();
		throw new Error("GitHub decision broker call was cancelled");
	}
	try {
		return await operation();
	} finally {
		if (intrinsicSignalAborted(context.signal)) context.acknowledgeAbort();
	}
}

export function adaptGitHubDecisionBroker(broker: GitHubDecisionBroker): ParentDecisionBroker {
	if (!(broker instanceof GitHubDecisionBroker)) throw new Error("invalid GitHub decision broker adapter input");
	return {
		async request(request, context) {
			return withActualBrokerContext(context, async () => {
				const binding = brokerBinding(request);
				const requested = validateHumanDecisionRecord(await broker.request(request, { signal: context.signal }));
				const canonical = await broker.readRecord(request.requestId, binding);
				if (!canonicalDataEqual(requested, canonical)) {
					throw new Error("GitHub decision request result disagrees with canonical state");
				}
				return canonical;
			});
		},
		async poll(requestId, binding, options, context) {
			return withActualBrokerContext(context, async () => {
				const result = await broker.poll(requestId, binding, { ...options, signal: context.signal });
				const canonical = await broker.readRecord(requestId, binding);
				validateActualBrokerPollResult(result, canonical);
				return canonical;
			});
		},
		async consume(requestId, binding, context) {
			return withActualBrokerContext(context, async () => {
				const evidence = readBoundedExactRecord(
					await broker.consume(requestId, binding),
					["option", "actor", "sourceUrl", "decidedAt"],
					[],
					"GitHub decision consume evidence",
				);
				const canonical = await broker.readRecord(requestId, binding);
				if (canonical.status !== "consumed" || canonical.decision === undefined
					|| evidence.option !== canonical.decision.option || evidence.actor !== canonical.decision.actor
					|| evidence.sourceUrl !== canonical.decision.sourceUrl || evidence.decidedAt !== canonical.decision.decidedAt) {
					throw new Error("GitHub decision consume evidence disagrees with canonical state");
				}
				return canonical;
			});
		},
	};
}

export interface RequiredCheckPolicySource {
	findRequiredCheckPolicies(
		query: { repository: string; baseBranch: string },
		context: ExternalCallContext,
	): Promise<AuthoritativeLookup<RequiredGitHubCheckPolicyObservation>>;
	findParentOrchestrationPolicyBundle?(
		query: {
			repository: string;
			parentIssue: number;
			generation: number;
			parentBranch: string;
			parentBaseBranch: string;
		},
		context: ExternalCallContext,
	): Promise<AuthoritativeLookup<ParentOrchestrationPolicyAuthority>>;
}

export interface ParentDecisionPolicy {
	requestId: string;
	actorAllowlist: readonly string[];
	expiresAt: string;
	question: string;
}

export interface WorkspaceHandoffSource {
	captureHandoff(
		workspace: ClaimedWorkspace,
		verificationState: "passed",
		context: ExternalCallContext,
	): Promise<WorkspaceHandoffEvidence>;
}

export type ChildIntegrationDecision =
	| { kind: "integrated"; receipt: ChildIntegrationReceipt; reused: boolean }
	| { kind: "blocked"; blockers: Array<GitHubEvidenceBlocker | "handoff_invalid" | "pull_request_missing" | "policy_moved"> };

export type ParentReadinessDecision =
	| { kind: "ready"; pullRequest: GitHubPullRequestEvidence; reused: boolean }
	| { kind: "blocked"; blockers: string[] }
	| { kind: "awaiting_human"; reason: "pending" | "expired" | "broker_unavailable" }
	| { kind: "rejected" };

export type ParentReadinessPreparation = ParentReadinessDecision
	| { kind: "prepared"; operation: PreparedParentReadyOperation };

type ExactRecord = Record<string, unknown>;

function exactRecord(value: unknown, required: readonly string[], optional: readonly string[] = []): ExactRecord {
	return readBoundedExactRecord(value, required, optional, "parent orchestration plan");
}

function inlineText(value: unknown, description: string, maximum = 1_024): string {
	if (typeof value !== "string" || value.length === 0 || value.length > maximum || Buffer.byteLength(value) > maximum
		|| value.trim() !== value || UNSAFE_INLINE.test(value)) {
		throw new Error(`invalid ${description}`);
	}
	assertNoSensitiveText(value, description);
	return value;
}

function bodyText(value: unknown, description: string): string {
	if (typeof value !== "string" || value.length === 0 || value.length > MAX_BODY_BYTES
		|| Buffer.byteLength(value) > MAX_BODY_BYTES || UNSAFE_BODY.test(value)) {
		throw new Error(`${description} must be bounded safe text`);
	}
	assertNoSensitiveText(value, description);
	return value.replace(/\r\n?/gu, "\n");
}

function githubNumber(value: unknown, description: string): number {
	if (!Number.isSafeInteger(value) || (value as number) < 1 || (value as number) > MAX_GITHUB_NUMBER) {
		throw new Error(`invalid ${description}`);
	}
	return value as number;
}

function generation(value: unknown): number {
	if (!Number.isSafeInteger(value) || (value as number) < 1 || (value as number) > MAX_GITHUB_NUMBER) {
		throw new Error("invalid orchestration generation");
	}
	return value as number;
}

function sha(value: unknown, description: string): string {
	if (typeof value !== "string" || !SHA.test(value)) throw new Error(`invalid ${description}`);
	return value;
}

function timestamp(value: unknown, description: string): string {
	if (typeof value !== "string" || value.length > 64) throw new Error(`invalid ${description}`);
	const match = RFC3339_UTC.exec(value);
	if (match === null) throw new Error(`invalid ${description}`);
	const canonical = `${match[1]}.${(match[2] ?? "").padEnd(3, "0")}Z`;
	const parsed = new Date(canonical);
	if (!Number.isFinite(parsed.valueOf()) || parsed.toISOString() !== canonical) throw new Error(`invalid ${description}`);
	return canonical;
}

function repository(value: unknown): string {
	const result = inlineText(value, "repository", 512).toLowerCase();
	if (!REPOSITORY.test(result)) throw new Error("invalid GitHub repository");
	return result;
}

function boundedArray(value: unknown, description: string, maximum = MAX_LIST, allowEmpty = false): unknown[] {
	if (nodeTypes.isProxy(value) || !Array.isArray(value) || Object.getPrototypeOf(value) !== Array.prototype) {
		throw new Error(`${description} must be a canonical array`);
	}
	const lengthDescriptor = Object.getOwnPropertyDescriptor(value, "length");
	if (lengthDescriptor === undefined || !Object.hasOwn(lengthDescriptor, "value")
		|| !Number.isSafeInteger(lengthDescriptor.value) || lengthDescriptor.value < 0
		|| (!allowEmpty && lengthDescriptor.value === 0) || lengthDescriptor.value > maximum) {
		throw new Error(`${description} must be a bounded array of at most ${maximum} values`);
	}
	const length = lengthDescriptor.value as number;
	const values: unknown[] = [];
	let entries = 0;
	for (const key in value) {
		if (!Object.hasOwn(value, key)) continue;
		if (entries >= length) throw new Error(`${description} has an invalid array field`);
		if (typeof key !== "string" || !/^(?:0|[1-9]\d*)$/u.test(key)) throw new Error(`${description} has an invalid array field`);
		const index = Number(key);
		const descriptor = Object.getOwnPropertyDescriptor(value, key);
		if (index >= length || descriptor === undefined || !Object.hasOwn(descriptor, "value") || descriptor.enumerable !== true) {
			throw new Error(`${description} must contain only dense data values`);
		}
		values[index] = descriptor.value;
		entries += 1;
	}
	if (entries !== length) throw new Error(`${description} must be a dense canonical array`);
	for (let index = 0; index < length; index += 1) {
		if (!Object.hasOwn(values, index)) throw new Error(`${description} must be a dense canonical array`);
	}
	return values;
}

function authoritativeItems(value: unknown, description: string, maximum = MAX_LIST): unknown[] {
	const snapshot = exactRecord(value, ["items", "complete"]);
	if (snapshot.complete !== true) throw new Error(`${description} is incomplete; bounded pagination must be exhausted`);
	return boundedArray(snapshot.items, `${description} items`, maximum, true);
}

function unique<T>(values: readonly T[], key: (value: T) => string, description: string): void {
	const keys = values.map(key);
	if (new Set(keys).size !== keys.length) throw new Error(`duplicate ${description}`);
}

function stableMarker(kind: string, coordinates: readonly (string | number)[], intent: unknown): string {
	const digest = createHash("sha256").update(JSON.stringify(intent)).digest("hex").slice(0, 24);
	return `<!-- shepherd-${kind}:v1:${coordinates.join(":")}:${digest} -->`;
}

function deepFreeze<T>(value: T): T {
	if (typeof value !== "object" || value === null || Object.isFrozen(value)) return value;
	for (const child of Object.values(value)) deepFreeze(child);
	return Object.freeze(value);
}

function validateVerification(value: unknown): VerificationRequirement {
	const candidate = exactRecord(value, ["id", "kind", "description"]);
	if (!["test", "typecheck", "offline_rpc", "diff_scope"].includes(candidate.kind as string)) {
		throw new Error("invalid verification requirement kind");
	}
	return {
		id: inlineText(candidate.id, "verification ID", 128),
		kind: candidate.kind as VerificationRequirementKind,
		description: inlineText(candidate.description, "verification description", 2_048),
	};
}

function validateStringList(
	value: unknown,
	description: string,
	pattern?: RegExp,
	allowEmpty = false,
): string[] {
	const values = boundedArray(value, description, MAX_LIST, allowEmpty).map((entry) => inlineText(entry, description, 512));
	if (pattern !== undefined && values.some((entry) => !pattern.test(entry))) throw new Error(`invalid ${description}`);
	unique(values, (entry) => entry, description);
	return [...values].sort();
}

const HUMAN_GATES: readonly HumanDecisionGate[] = ["requirements", "scope", "review", "head", "merge", "parent_merge"];

interface ChildIntent {
	id: string;
	title: string;
	objective: string;
	slug: string;
	dependsOn: string[];
	access: WorkAccess;
	writeScopes: string[];
	requiredSkills: string[];
	verification: VerificationRequirement[];
	humanGates: HumanDecisionGate[];
}

function validateChildIntent(value: unknown): ChildIntent {
	const candidate = exactRecord(value, [
		"id",
		"title",
		"objective",
		"slug",
		"dependsOn",
		"access",
		"writeScopes",
		"requiredSkills",
		"verification",
		"humanGates",
	]);
	const id = inlineText(candidate.id, "child ID", 64);
	if (!CHILD_ID.test(id)) throw new Error("invalid child ID");
	const slug = inlineText(candidate.slug, "child branch slug", 100);
	if (!SLUG.test(slug)) throw new Error("invalid child branch slug");
	if (candidate.access !== "mutating") throw new Error("top-level child work must be mutating");
	const dependencies = validateStringList(candidate.dependsOn, "child dependencies", CHILD_ID, true);
	const scopes = validateStringList(candidate.writeScopes, "child write scopes");
	const skills = validateStringList(candidate.requiredSkills, "required skills", SKILL);
	const verification = boundedArray(candidate.verification, "verification requirements").map(validateVerification);
	unique(verification, (requirement) => requirement.id, "verification requirement ID");
	const humanGates = validateStringList(candidate.humanGates, "human gates") as HumanDecisionGate[];
	if (humanGates.some((gate) => !HUMAN_GATES.includes(gate))) throw new Error("invalid human gate");
	return {
		id,
		title: inlineText(candidate.title, "child title", 256),
		objective: inlineText(candidate.objective, "child objective", 4_096),
		slug,
		dependsOn: dependencies,
		access: candidate.access,
		writeScopes: scopes,
		requiredSkills: skills,
		verification,
		humanGates,
	};
}

function renderChildIssueBody(parentIssue: number, child: ChildIntent, marker: string): string {
	const dependencies = child.dependsOn.length > 0 ? child.dependsOn.map((id) => `- ${id}`).join("\n") : "- none";
	const scopes = child.writeScopes.map((scope) => `- ${scope}`).join("\n");
	return [
		child.objective,
		"",
		`Parent: #${parentIssue}`,
		"",
		"Dependencies:",
		dependencies,
		"",
		"Owned scopes:",
		scopes,
		"",
		marker,
	].join("\n");
}

type CanonicalPlanData = Omit<ParentOrchestrationPlan, "canonical">;

function buildParentOrchestrationPlan(value: unknown, policyValue: unknown): CanonicalPlanData {
	const candidate = exactRecord(value, [
		"repository",
		"parentIssue",
		"generation",
		"title",
		"objective",
		"parentBranch",
		"parentBaseBranch",
		"children",
	]);
	const bundle = exactRecord(policyValue, ["schemaVersion", "requiredCheckPolicies"]);
	if (bundle.schemaVersion !== 1) throw new Error("unsupported parent orchestration policy bundle");
	const canonicalRepository = repository(candidate.repository);
	const parentIssue = githubNumber(candidate.parentIssue, "parent issue number");
	const canonicalGeneration = generation(candidate.generation);
	const title = inlineText(candidate.title, "parent title", 256);
	const objective = inlineText(candidate.objective, "parent objective", 4_096);
	const parentBranch = canonicalGitRef(candidate.parentBranch, "parent branch");
	const parentBaseBranch = canonicalGitRef(candidate.parentBaseBranch, "parent base branch");
	if (parentBranch === parentBaseBranch) throw new Error("parent branch and base branch must differ");
	const requiredCheckPolicies = boundedArray(bundle.requiredCheckPolicies, "required-check policies")
		.map(validateRequiredGitHubCheckPolicy)
		.sort((left, right) => left.baseBranch.localeCompare(right.baseBranch));
	unique(requiredCheckPolicies, (policy) => `${policy.repository}\u0000${policy.baseBranch}`, "required-check policy coordinate");
	const requiredPolicyBranches = [parentBaseBranch, parentBranch].sort();
	if (requiredCheckPolicies.length !== requiredPolicyBranches.length
		|| requiredCheckPolicies.some((policy, index) => policy.repository !== canonicalRepository
			|| policy.baseBranch !== requiredPolicyBranches[index])) {
		throw new Error("required-check policies must exactly cover the repository parent base and parent branch");
	}
	const intents = boundedArray(candidate.children, "children", MAX_CHILDREN).map(validateChildIntent);
	unique(intents, (child) => child.id, "child ID");
	unique(intents, (child) => child.slug, "child branch slug");
	let graph;
	try {
		graph = validateDependencyGraph(intents.map((child): DependencyWorkItem => ({
			id: child.id,
			dependsOn: child.dependsOn,
			status: "pending",
			access: child.access,
			writeScopes: child.writeScopes,
		})));
	} catch (error) {
		if (error instanceof DependencyGraphError) {
			throw new Error(`invalid child dependency graph: ${error.code}: ${error.message}`);
		}
		throw error;
	}
	const byId = new Map(intents.map((child) => [child.id, child]));
	const children = graph.items.map((item): BoundedChildRecord => {
		const child = byId.get(item.id);
		if (child === undefined) throw new Error("validated child graph lost metadata");
		const markerIntent = {
			parentIssue,
			generation: canonicalGeneration,
			requiredCheckPolicies: requiredCheckPolicies.map((policy) => policy.digest),
			...child,
			prBase: parentBranch,
		};
		const issueMarker = stableMarker("child-issue", [parentIssue, child.id], markerIntent);
		const pullRequestMarker = stableMarker("child-pr", [parentIssue, child.id], markerIntent);
		return {
			id: child.id,
			dependsOn: [...item.dependsOn],
			status: "pending",
			access: child.access,
			writeScopes: [...item.writeScopes],
			title: child.title,
			objective: child.objective,
			branch: { kind: "canonical_issue_branch", slug: child.slug },
			prBase: parentBranch,
			requiredSkills: [...child.requiredSkills],
			verification: child.verification.map((requirement) => ({ ...requirement })),
			humanGates: [...child.humanGates],
			markers: { issue: issueMarker, pullRequest: pullRequestMarker },
			issueBody: renderChildIssueBody(parentIssue, child, issueMarker),
		};
	});
	const parentIntent = {
		repository: canonicalRepository,
		parentIssue,
		generation: canonicalGeneration,
		title,
		objective,
		parentBranch,
		parentBaseBranch,
		requiredCheckPolicies: requiredCheckPolicies.map((policy) => ({
			repository: policy.repository,
			baseBranch: policy.baseBranch,
			revision: policy.revision,
			digest: policy.digest,
			requiredChecks: policy.requiredChecks,
		})),
		children: children.map((child) => ({ id: child.id, markers: child.markers })),
	};
	return deepFreeze({
		schemaVersion: 1,
		repository: canonicalRepository,
		parentIssue,
		generation: canonicalGeneration,
		title,
		objective,
		parentBranch,
		parentBaseBranch,
		markers: {
			parentPullRequest: stableMarker("parent-pr", [parentIssue, canonicalGeneration], parentIntent),
			roster: stableMarker("parent-roster", [parentIssue, canonicalGeneration], parentIntent),
		},
		requiredCheckPolicies,
		children,
	});
}

function canonicalObjective(plan: CanonicalPlanData): Record<string, unknown> {
	return {
		repository: plan.repository,
		parentIssue: plan.parentIssue,
		generation: plan.generation,
		title: plan.title,
		objective: plan.objective,
		parentBranch: plan.parentBranch,
		parentBaseBranch: plan.parentBaseBranch,
		children: plan.children.map((child) => ({
			id: child.id,
			title: child.title,
			objective: child.objective,
			slug: child.branch.slug,
			dependsOn: child.dependsOn,
			access: child.access,
			writeScopes: child.writeScopes,
			requiredSkills: child.requiredSkills,
			verification: child.verification,
			humanGates: child.humanGates,
		})),
	};
}

function canonicalPolicyBundle(plan: CanonicalPlanData): ParentOrchestrationPolicyBundle {
	return {
		schemaVersion: 1,
		requiredCheckPolicies: plan.requiredCheckPolicies.map((policy) => ({
			...policy,
			requiredChecks: policy.requiredChecks.map((check) => ({ ...check })),
		})),
	};
}

function canonicalPlanDigest(serialized: string): string {
	return createHash("sha256").update(serialized).digest("hex");
}

export function createParentOrchestrationPlan(
	value: unknown,
	policyValue: ParentOrchestrationPolicyBundle,
): ParentOrchestrationPlan {
	const built = buildParentOrchestrationPlan(value, policyValue);
	const serialized = JSON.stringify({
		objective: canonicalObjective(built),
		policyBundle: canonicalPolicyBundle(built),
	});
	if (Buffer.byteLength(serialized) > MAX_CANONICAL_PLAN_BYTES) throw new Error("canonical plan serialization is oversized");
	return deepFreeze({
		...built,
		canonical: { schemaVersion: 1, serialized, digest: canonicalPlanDigest(serialized) },
	});
}

function objectiveAuthorityQuery(value: unknown): {
	repository: string;
	parentIssue: number;
	generation: number;
	parentBranch: string;
	parentBaseBranch: string;
} {
	const candidate = exactRecord(value, [
		"repository", "parentIssue", "generation", "title", "objective", "parentBranch", "parentBaseBranch", "children",
	]);
	const parentBranch = canonicalGitRef(candidate.parentBranch, "parent branch");
	const parentBaseBranch = canonicalGitRef(candidate.parentBaseBranch, "parent base branch");
	if (parentBranch === parentBaseBranch) throw new Error("parent branch and base branch must differ");
	return {
		repository: repository(candidate.repository),
		parentIssue: githubNumber(candidate.parentIssue, "parent issue number"),
		generation: generation(candidate.generation),
		parentBranch,
		parentBaseBranch,
	};
}

function validateParentPolicyAuthority(
	value: unknown,
	expected: ReturnType<typeof objectiveAuthorityQuery>,
): ParentOrchestrationPolicyAuthority {
	const candidate = exactRecord(value, [
		"schemaVersion", "authority", "repository", "parentIssue", "generation", "parentBranch", "parentBaseBranch",
		"revision", "observedAt", "policyBundle",
	]);
	const authority = {
		schemaVersion: candidate.schemaVersion,
		authority: candidate.authority,
		repository: repository(candidate.repository),
		parentIssue: githubNumber(candidate.parentIssue, "policy authority parent issue"),
		generation: generation(candidate.generation),
		parentBranch: canonicalGitRef(candidate.parentBranch, "policy authority parent branch"),
		parentBaseBranch: canonicalGitRef(candidate.parentBaseBranch, "policy authority parent base branch"),
		revision: generation(candidate.revision),
		observedAt: timestamp(candidate.observedAt, "policy authority observation timestamp"),
		policyBundle: candidate.policyBundle as ParentOrchestrationPolicyBundle,
	};
	if (authority.schemaVersion !== 1 || authority.authority !== "controller"
		|| authority.repository !== expected.repository || authority.parentIssue !== expected.parentIssue
		|| authority.generation !== expected.generation || authority.parentBranch !== expected.parentBranch
		|| authority.parentBaseBranch !== expected.parentBaseBranch) {
		throw new Error("parent orchestration policy authority does not match the exact objective coordinates");
	}
	return authority as ParentOrchestrationPolicyAuthority;
}

function childFor(plan: ParentOrchestrationPlan, childId: string): BoundedChildRecord {
	const child = plan.children.find((candidate) => candidate.id === childId);
	if (child === undefined) throw new Error(`unknown child record ${childId}`);
	return child;
}

export function selectReadyChildren(
	plan: ParentOrchestrationPlan,
	statuses: Readonly<Record<string, WorkItemStatus>>,
	maxConcurrency: number,
): ReadyQueueSelection {
	const canonicalPlan = validateParentOrchestrationPlan(plan);
	const canonicalStatuses = statusesForPlan(canonicalPlan, statuses);
	const items = canonicalPlan.children.map((child): DependencyWorkItem => ({
		id: child.id,
		dependsOn: [...child.dependsOn],
		status: canonicalStatuses[child.id],
		access: child.access,
		writeScopes: [...child.writeScopes],
	}));
	return selectReadyWork(items, { maxConcurrency });
}

export function validateGitHubChildIssue(value: unknown): GitHubChildIssue {
	const candidate = exactRecord(value, ["number", "marker", "title", "body", "state", "parentIssue"]);
	if (candidate.state !== "open" && candidate.state !== "closed") throw new Error("invalid child issue state");
	return {
		number: githubNumber(candidate.number, "child issue number"),
		marker: inlineText(candidate.marker, "child issue marker", 512),
		title: inlineText(candidate.title, "child issue title", 256),
		body: bodyText(candidate.body, "child issue body"),
		state: candidate.state,
		parentIssue: githubNumber(candidate.parentIssue, "child parent issue number"),
	};
}

function assertChildIssueMatches(issue: GitHubChildIssue, plan: ParentOrchestrationPlan, child: BoundedChildRecord): void {
	if (issue.marker !== child.markers.issue || issue.title !== child.title || issue.body !== child.issueBody
		|| issue.parentIssue !== plan.parentIssue || issue.state !== "open"
		|| issue.body.split(child.markers.issue).length !== 2) {
		throw new Error("child issue marker collision or canonical resource mismatch");
	}
}

export function materializeChildRecord(
	plan: ParentOrchestrationPlan,
	childId: string,
	issueValue: GitHubChildIssue,
): MaterializedChildRecord {
	const canonicalPlan = validateParentOrchestrationPlan(plan);
	const child = childFor(canonicalPlan, childId);
	const issue = validateGitHubChildIssue(issueValue);
	assertChildIssueMatches(issue, canonicalPlan, child);
	return {
		...child,
		issue: issue.number,
		branch: canonicalIssueBranch(issue.number, child.branch.slug),
	};
}

function boundedDataEntries(value: unknown, maximum: number, description: string): Array<[string, unknown]> {
	if (typeof value !== "object" || value === null || Array.isArray(value) || nodeTypes.isProxy(value)
		|| (Object.getPrototypeOf(value) !== Object.prototype && Object.getPrototypeOf(value) !== null)) {
		throw new Error(`invalid ${description}`);
	}
	const entries: Array<[string, unknown]> = [];
	for (const key in value) {
		if (!Object.hasOwn(value, key)) continue;
		if (entries.length >= maximum) throw new Error(`${description} is oversized`);
		const descriptor = Object.getOwnPropertyDescriptor(value, key);
		if (descriptor === undefined || !Object.hasOwn(descriptor, "value") || descriptor.enumerable !== true) {
			throw new Error(`invalid ${description}`);
		}
		entries.push([key, descriptor.value]);
	}
	return entries;
}

function validateStatusRecord(value: unknown): Record<string, WorkItemStatus> {
	const entries = boundedDataEntries(value, MAX_CHILDREN, "roster status snapshot");
	const allowed: readonly WorkItemStatus[] = ["pending", "running", "succeeded", "failed", "blocked"];
	const result: Record<string, WorkItemStatus> = {};
	for (const [key, entry] of entries) {
		if (!allowed.includes(entry as WorkItemStatus)) throw new Error("invalid roster status");
		result[key] = entry as WorkItemStatus;
	}
	return result;
}

export function validateGitHubRosterSnapshot(value: unknown): GitHubRosterSnapshot {
	const candidate = exactRecord(value, [
		"id", "marker", "parentIssue", "generation", "body", "statuses", "statusEpoch", "revision", "updatedAt",
	]);
	const statuses = validateStatusRecord(candidate.statuses);
	return {
		id: githubNumber(candidate.id, "roster resource ID"),
		marker: inlineText(candidate.marker, "roster marker", 512),
		parentIssue: githubNumber(candidate.parentIssue, "roster parent issue"),
		generation: generation(candidate.generation),
		body: bodyText(candidate.body, "roster body"),
		statuses,
		statusEpoch: generation(candidate.statusEpoch),
		revision: generation(candidate.revision),
		updatedAt: timestamp(candidate.updatedAt, "roster update timestamp"),
	};
}

function pullRequestSnapshotDigest(value: Omit<CanonicalPullRequestSnapshot, "digest">): string {
	const { revision: _revision, observedAt: _observedAt, ...stableIdentity } = value;
	return createHash("sha256").update(JSON.stringify(stableIdentity)).digest("hex");
}

export function createCanonicalPullRequestSnapshot(value: GitHubPullRequestEvidence): CanonicalPullRequestSnapshot {
	const canonical = validateGitHubPullRequestEvidence(value);
	const snapshot = {
		repository: canonical.repository,
		workItemId: canonical.workItemId,
		number: canonical.number,
		generation: canonical.generation,
		marker: canonical.marker,
		baseBranch: canonical.baseBranch,
		headBranch: canonical.headBranch,
		baseSha: canonical.baseSha,
		headSha: canonical.headSha,
		changedPaths: [...canonical.changedPaths],
		allowedScopes: [...canonical.allowedScopes],
		policyDigest: canonical.policyDigest,
		revision: canonical.revision,
		observedAt: canonical.observedAt,
	};
	return { ...snapshot, digest: pullRequestSnapshotDigest(snapshot) };
}

function validatePullRequestSnapshot(value: unknown): CanonicalPullRequestSnapshot {
	const candidate = exactRecord(value, [
		"repository", "workItemId", "number", "generation", "marker", "baseBranch", "headBranch", "baseSha",
		"headSha", "changedPaths", "allowedScopes", "policyDigest", "revision", "observedAt", "digest",
	]);
	const snapshot = {
		repository: repository(candidate.repository),
		workItemId: inlineText(candidate.workItemId, "snapshot work item ID"),
		number: githubNumber(candidate.number, "snapshot pull request"),
		generation: generation(candidate.generation),
		marker: inlineText(candidate.marker, "snapshot pull request marker", 512),
		baseBranch: canonicalGitRef(candidate.baseBranch, "snapshot base branch"),
		headBranch: canonicalGitRef(candidate.headBranch, "snapshot head branch"),
		baseSha: sha(candidate.baseSha, "snapshot base SHA"),
		headSha: sha(candidate.headSha, "snapshot head SHA"),
		changedPaths: validateStringList(candidate.changedPaths, "snapshot changed paths", undefined, true),
		allowedScopes: validateStringList(candidate.allowedScopes, "snapshot allowed scopes"),
		policyDigest: typeof candidate.policyDigest === "string" && IDENTITY.test(candidate.policyDigest)
			? candidate.policyDigest
			: (() => { throw new Error("invalid snapshot policy digest"); })(),
		revision: generation(candidate.revision),
		observedAt: timestamp(candidate.observedAt, "snapshot observation timestamp"),
	};
	if (typeof candidate.digest !== "string" || !IDENTITY.test(candidate.digest)
		|| candidate.digest !== pullRequestSnapshotDigest(snapshot)) throw new Error("pull request snapshot digest mismatch");
	return { ...snapshot, digest: candidate.digest };
}

function pullRequestObservation(value: GitHubPullRequestEvidence): PullRequestObservation {
	return { revision: value.revision, observedAt: value.observedAt, state: value.state };
}

function validatePullRequestObservation(value: unknown): PullRequestObservation {
	const candidate = exactRecord(value, ["revision", "observedAt", "state"]);
	if (candidate.state !== "open" && candidate.state !== "closed" && candidate.state !== "merged") {
		throw new Error("invalid pull request observation state");
	}
	return {
		revision: generation(candidate.revision),
		observedAt: timestamp(candidate.observedAt, "pull request observation timestamp"),
		state: candidate.state,
	};
}

function validateControllerProvenance(value: unknown): ControllerIntegrationProvenance {
	const candidate = exactRecord(value, [
		"authority", "planDigest", "policyDigest", "policyRevision", "policyObservedAt", "changedPathDigest",
		"reviewAuthorizationDigest", "reviewResultDigest", "reviewCompletedAt", "evidenceRevision", "observedAt",
	]);
	if (candidate.authority !== "controller" || typeof candidate.planDigest !== "string" || !IDENTITY.test(candidate.planDigest)
		|| typeof candidate.policyDigest !== "string" || !IDENTITY.test(candidate.policyDigest)
		|| typeof candidate.changedPathDigest !== "string" || !IDENTITY.test(candidate.changedPathDigest)
		|| typeof candidate.reviewAuthorizationDigest !== "string" || !IDENTITY.test(candidate.reviewAuthorizationDigest)
		|| typeof candidate.reviewResultDigest !== "string" || !IDENTITY.test(candidate.reviewResultDigest)) {
		throw new Error("invalid controller integration provenance");
	}
	return {
		authority: "controller",
		planDigest: candidate.planDigest,
		policyDigest: candidate.policyDigest,
		policyRevision: generation(candidate.policyRevision),
		policyObservedAt: timestamp(candidate.policyObservedAt, "controller policy observation timestamp"),
		changedPathDigest: candidate.changedPathDigest,
		reviewAuthorizationDigest: candidate.reviewAuthorizationDigest,
		reviewResultDigest: candidate.reviewResultDigest,
		reviewCompletedAt: timestamp(candidate.reviewCompletedAt, "controller review completion timestamp"),
		evidenceRevision: generation(candidate.evidenceRevision),
		observedAt: timestamp(candidate.observedAt, "controller evidence observation timestamp"),
	};
}

function changedPathEvidenceDigest(value: GitHubChangedPathEvidence): string {
	const { revision: _revision, observedAt: _observedAt, ...stableEvidence } = value;
	return createHash("sha256").update(JSON.stringify(stableEvidence)).digest("hex");
}

function canonicalDigest(value: unknown): string {
	return createHash("sha256").update(JSON.stringify(value)).digest("hex");
}

function parentReadyPolicyAuthority(value: RequiredGitHubCheckPolicyObservation): ParentReadyPolicyAuthority {
	return {
		schemaVersion: 1,
		authority: "controller",
		repository: value.repository,
		baseBranch: value.baseBranch,
		revision: value.revision,
		digest: value.digest,
	};
}

function parentReadyAncestryAuthority(value: GitAncestryProof): ParentReadyAncestryAuthority {
	return {
		schemaVersion: 1,
		authority: "transport",
		repository: value.repository,
		ancestorSha: value.ancestorSha,
		descendantSha: value.descendantSha,
		result: true,
	};
}

export function canonicalizeParentReadyAuthorization(value: unknown): ParentReadyAuthorization {
	const candidate = exactRecord(value, [
		"schemaVersion", "repository", "parentIssue", "generation", "pullRequest", "decisionRequestId",
		"decisionDigest", "childRosterDigest", "policySetDigest", "parentReviewDigest", "parentPathDigest",
		"policies", "reviewAuthorizationDigest", "exactPaths", "children", "decision", "planDigest", "headSha",
		"pullRequestRevision",
	]);
	if (candidate.schemaVersion !== 1) throw new Error("invalid parent-ready authorization schema");
	const policies = boundedArray(candidate.policies, "parent-ready policy authority", MAX_LIST, true).map((entry) => {
		const policy = exactRecord(entry, ["schemaVersion", "authority", "repository", "baseBranch", "revision", "digest"]);
		if (policy.schemaVersion !== 1 || policy.authority !== "controller"
			|| typeof policy.digest !== "string" || !IDENTITY.test(policy.digest)) {
			throw new Error("invalid parent-ready policy authority");
		}
		return {
			schemaVersion: 1 as const,
			authority: "controller" as const,
			repository: repository(policy.repository),
			baseBranch: canonicalGitRef(policy.baseBranch, "parent-ready policy base branch"),
			revision: generation(policy.revision),
			digest: policy.digest,
		};
	});
	const children = boundedArray(candidate.children, "parent-ready child authority", MAX_CHILDREN, true).map((entry) => {
		const child = exactRecord(entry, ["receipt", "ancestry"]);
		const ancestry = exactRecord(child.ancestry, [
			"schemaVersion", "authority", "repository", "ancestorSha", "descendantSha", "result",
		]);
		if (ancestry.schemaVersion !== 1 || ancestry.authority !== "transport" || ancestry.result !== true) {
			throw new Error("invalid parent-ready ancestry authority");
		}
		return {
			receipt: validateChildIntegrationReceipt(child.receipt),
			ancestry: {
				schemaVersion: 1 as const,
				authority: "transport" as const,
				repository: repository(ancestry.repository),
				ancestorSha: sha(ancestry.ancestorSha, "parent-ready ancestry ancestor SHA"),
				descendantSha: sha(ancestry.descendantSha, "parent-ready ancestry descendant SHA"),
				result: true as const,
			},
		};
	});
	const digestFields = [
		candidate.decisionDigest,
		candidate.childRosterDigest,
		candidate.policySetDigest,
		candidate.parentReviewDigest,
		candidate.parentPathDigest,
		candidate.reviewAuthorizationDigest,
		candidate.planDigest,
	];
	if (digestFields.some((entry) => typeof entry !== "string" || !IDENTITY.test(entry))) {
		throw new Error("invalid parent-ready authorization digest coordinate");
	}
	const payload = {
		schemaVersion: 1 as const,
		repository: repository(candidate.repository),
		parentIssue: generation(candidate.parentIssue),
		generation: generation(candidate.generation),
		pullRequest: generation(candidate.pullRequest),
		decisionRequestId: inlineText(candidate.decisionRequestId, "parent-ready decision request ID", 128),
		decisionDigest: candidate.decisionDigest as string,
		childRosterDigest: candidate.childRosterDigest as string,
		policySetDigest: candidate.policySetDigest as string,
		parentReviewDigest: candidate.parentReviewDigest as string,
		parentPathDigest: candidate.parentPathDigest as string,
		policies,
		reviewAuthorizationDigest: candidate.reviewAuthorizationDigest as string,
		exactPaths: validateStringList(candidate.exactPaths, "parent-ready exact paths", undefined, true),
		children,
		decision: validateHumanDecisionRecord(candidate.decision),
		planDigest: candidate.planDigest as string,
		headSha: sha(candidate.headSha, "parent-ready head SHA"),
		pullRequestRevision: generation(candidate.pullRequestRevision),
	};
	return { ...payload, digest: canonicalDigest(payload) };
}

export function validateParentReadyAuthorization(value: unknown): ParentReadyAuthorization {
	const candidate = exactRecord(value, [
		"schemaVersion", "repository", "parentIssue", "generation", "pullRequest", "decisionRequestId",
		"decisionDigest", "childRosterDigest", "policySetDigest", "parentReviewDigest", "parentPathDigest",
		"policies", "reviewAuthorizationDigest", "exactPaths", "children", "decision", "planDigest", "headSha",
		"pullRequestRevision", "digest",
	]);
	const { digest, ...payload } = candidate;
	const canonical = canonicalizeParentReadyAuthorization(payload);
	if (digest !== canonical.digest) throw new Error("parent-ready authorization digest mismatch");
	return canonical;
}

function createParentReadyAuthority(
	plan: ParentOrchestrationPlan,
	pullRequest: GitHubPullRequestEvidence,
	decision: HumanDecisionRecord,
	review: IndependentReviewRecord,
	changedPaths: GitHubChangedPathEvidence,
	policies: Map<string, RequiredGitHubCheckPolicyObservation>,
	roster: { receipts: ChildIntegrationReceipt[]; ancestryProofs: GitAncestryProof[] },
): { authorization: ParentReadyAuthorization; freshness: ParentReadyFreshnessEnvelope } {
	const policySet = [...policies.values()].sort((left, right) => left.baseBranch.localeCompare(right.baseBranch));
	const stablePolicies = policySet.map(parentReadyPolicyAuthority);
	const stableChildren = roster.receipts.map((receipt, index) => ({
		receipt,
		ancestry: parentReadyAncestryAuthority(roster.ancestryProofs[index]),
	}));
	const decisionDigest = canonicalDigest(validateHumanDecisionRecord(decision));
	const childRosterDigest = canonicalDigest({ children: stableChildren });
	const policySetDigest = canonicalDigest(stablePolicies);
	const reviewAuthorizationDigest = independentReviewAuthorizationDigest(review);
	const parentPathDigest = changedPathEvidenceDigest(changedPaths);
	const authorization = canonicalizeParentReadyAuthorization({
		schemaVersion: 1,
		repository: plan.repository,
		parentIssue: plan.parentIssue,
		generation: plan.generation,
		pullRequest: pullRequest.number,
		decisionRequestId: decision.requestId,
		decisionDigest,
		childRosterDigest,
		policySetDigest,
		parentReviewDigest: reviewAuthorizationDigest,
		parentPathDigest,
		policies: stablePolicies,
		reviewAuthorizationDigest,
		exactPaths: [...changedPaths.paths],
		children: stableChildren,
		decision: validateHumanDecisionRecord(decision),
		planDigest: plan.canonical.digest,
		headSha: pullRequest.headSha,
		pullRequestRevision: pullRequest.revision,
	});
	const freshnessPayload = {
		schemaVersion: 1 as const,
		policyObservations: policySet,
		parentReviewResultDigest: independentReviewResultDigest(review),
		parentReviewCompletedAt: review.completedAt,
		parentPathRevision: changedPaths.revision,
		parentPathObservedAt: changedPaths.observedAt,
		childAncestryProofs: roster.ancestryProofs,
	};
	return {
		authorization,
		freshness: { ...freshnessPayload, digest: canonicalDigest(freshnessPayload) },
	};
}

function controllerIntegrationProvenance(
	plan: ParentOrchestrationPlan,
	policy: RequiredGitHubCheckPolicy,
	policyObservation: RequiredGitHubCheckPolicyObservation,
	changedPaths: GitHubChangedPathEvidence,
	review: IndependentReviewRecord,
): ControllerIntegrationProvenance {
	return {
		authority: "controller",
		planDigest: plan.canonical.digest,
		policyDigest: policy.digest,
		policyRevision: policyObservation.revision,
		policyObservedAt: policyObservation.observedAt,
		changedPathDigest: changedPathEvidenceDigest(changedPaths),
		reviewAuthorizationDigest: independentReviewAuthorizationDigest(review),
		reviewResultDigest: independentReviewResultDigest(review),
		reviewCompletedAt: review.completedAt,
		evidenceRevision: changedPaths.revision,
		observedAt: changedPaths.observedAt,
	};
}

function controllerAuthorizationMatches(
	current: ControllerIntegrationProvenance,
	receipt: ControllerIntegrationProvenance,
	currentPullRequest: GitHubPullRequestEvidence,
	receiptSnapshot: CanonicalPullRequestSnapshot,
): boolean {
	const currentStable = {
		authority: current.authority,
		planDigest: current.planDigest,
		policyDigest: current.policyDigest,
		policyRevision: current.policyRevision,
		changedPathDigest: current.changedPathDigest,
		reviewAuthorizationDigest: current.reviewAuthorizationDigest,
	};
	const receiptStable = {
		authority: receipt.authority,
		planDigest: receipt.planDigest,
		policyDigest: receipt.policyDigest,
		policyRevision: receipt.policyRevision,
		changedPathDigest: receipt.changedPathDigest,
		reviewAuthorizationDigest: receipt.reviewAuthorizationDigest,
	};
	const samePullRequestObservation = currentPullRequest.revision === receiptSnapshot.revision
		&& currentPullRequest.observedAt === receiptSnapshot.observedAt;
	return currentPullRequest.revision >= receiptSnapshot.revision
		&& currentPullRequest.observedAt >= receiptSnapshot.observedAt
		&& (samePullRequestObservation
			? current.evidenceRevision === receipt.evidenceRevision && current.observedAt === receipt.observedAt
			: current.evidenceRevision >= receipt.evidenceRevision && current.observedAt >= receipt.observedAt)
		&& current.policyObservedAt >= receipt.policyObservedAt
		&& canonicalDataEqual(currentStable, receiptStable);
}

function validateTransportProvenance(value: unknown): TransportMutationProvenance {
	const candidate = exactRecord(value, ["authority", "idempotencyKey", "intentDigest", "revision"]);
	if (candidate.authority !== "transport" || typeof candidate.intentDigest !== "string" || !IDENTITY.test(candidate.intentDigest)) {
		throw new Error("invalid transport mutation provenance");
	}
	return {
		authority: "transport",
		idempotencyKey: inlineText(candidate.idempotencyKey, "transport mutation key", 512),
		intentDigest: candidate.intentDigest,
		revision: generation(candidate.revision),
	};
}

export function validateReviewedChildIntegrationEvidence(value: unknown): GitReviewedChildIntegrationEvidence {
	const candidate = exactRecord(value, [
		"schemaVersion", "authority", "parentBranch", "baseSha", "headSha", "mergeCommitSha", "parentHead", "reused",
	]);
	if (candidate.schemaVersion !== 1 || candidate.authority !== "git" || typeof candidate.reused !== "boolean") {
		throw new Error("invalid reviewed child Git integration evidence");
	}
	return {
		schemaVersion: 1,
		authority: "git",
		parentBranch: canonicalGitRef(candidate.parentBranch, "Git integration parent branch"),
		baseSha: sha(candidate.baseSha, "Git integration base SHA"),
		headSha: sha(candidate.headSha, "Git integration head SHA"),
		mergeCommitSha: sha(candidate.mergeCommitSha, "Git integration merge commit SHA"),
		parentHead: sha(candidate.parentHead, "Git integration parent head"),
		reused: candidate.reused,
	};
}

export function validateChildIntegrationReceipt(value: unknown): ChildIntegrationReceipt {
	const candidate = exactRecord(value, [
		"childId",
		"pullRequest",
		"generation",
		"marker",
		"baseSha",
		"headSha",
		"parentBranch",
		"pullRequestSnapshot",
		"observation",
		"controllerProvenance",
		"transportProvenance",
		"integratedAt",
	]);
	const receipt: ChildIntegrationReceipt = {
		childId: inlineText(candidate.childId, "integration child ID", 64),
		pullRequest: githubNumber(candidate.pullRequest, "integrated pull request"),
		generation: generation(candidate.generation),
		marker: inlineText(candidate.marker, "integrated pull request marker", 512),
		baseSha: sha(candidate.baseSha, "integrated base SHA"),
		headSha: sha(candidate.headSha, "integrated head SHA"),
		parentBranch: canonicalGitRef(candidate.parentBranch, "integration parent branch"),
		pullRequestSnapshot: validatePullRequestSnapshot(candidate.pullRequestSnapshot),
		observation: validatePullRequestObservation(candidate.observation),
		controllerProvenance: validateControllerProvenance(candidate.controllerProvenance),
		transportProvenance: validateTransportProvenance(candidate.transportProvenance),
		integratedAt: timestamp(candidate.integratedAt, "integration timestamp"),
	};
	const integratedTime = new Date(receipt.integratedAt).valueOf();
	const earliest = Math.max(
		new Date(receipt.pullRequestSnapshot.observedAt).valueOf(),
		new Date(receipt.observation.observedAt).valueOf(),
		new Date(receipt.controllerProvenance.observedAt).valueOf(),
		new Date(receipt.controllerProvenance.policyObservedAt).valueOf(),
		new Date(receipt.controllerProvenance.reviewCompletedAt).valueOf(),
	);
	if (integratedTime < earliest) throw new Error("invalid integration receipt authority chronology");
	if (integratedTime > Date.now()) throw new Error("invalid integration receipt future chronology");
	return receipt;
}

export function createDurableMutationIntent(
	operation: DurableMutationOperation,
	coordinates: readonly (string | number)[],
	intent: unknown,
	expectedResourceRevision: number | null,
): DurableMutationIntent {
	const conditionalIntent = { expectedResourceRevision, intent };
	const intentDigest = createHash("sha256").update(JSON.stringify(conditionalIntent)).digest("hex");
	const keyDigest = createHash("sha256").update(JSON.stringify({
		operation,
		coordinates,
		expectedResourceRevision,
	})).digest("hex").slice(0, 32);
	return {
		schemaVersion: 1,
		operation,
		idempotencyKey: `shepherd-mutation:v1:${operation}:${keyDigest}`,
		intentDigest,
		expectedResourceRevision,
	};
}

export function validateDurableMutationResult<T>(
	value: unknown,
	intent: DurableMutationIntent,
	validateValue: (entry: unknown) => T,
): DurableMutationResult<T> {
	const candidate = exactRecord(value, [
		"schemaVersion", "idempotencyKey", "intentDigest", "revision", "applied", "value",
	]);
	if (candidate.schemaVersion !== 1 || candidate.idempotencyKey !== intent.idempotencyKey
		|| candidate.intentDigest !== intent.intentDigest || typeof candidate.applied !== "boolean") {
		throw new Error("durable mutation result provenance does not match its intent");
	}
	return {
		schemaVersion: 1,
		idempotencyKey: intent.idempotencyKey,
		intentDigest: intent.intentDigest,
		revision: generation(candidate.revision),
		applied: candidate.applied,
		value: validateValue(candidate.value),
	};
}

function transportProvenance(result: DurableMutationResult<unknown>): TransportMutationProvenance {
	return {
		authority: "transport",
		idempotencyKey: result.idempotencyKey,
		intentDigest: result.intentDigest,
		revision: result.revision,
	};
}

function validateAncestryProof(
	value: unknown,
	query: GitAncestryQuery,
	minimumObservedAt: string,
): GitAncestryProof {
	const candidate = exactRecord(value, [
		"schemaVersion", "authority", "repository", "ancestorSha", "descendantSha", "result", "revision", "observedAt",
	]);
	if (candidate.schemaVersion !== 1 || candidate.authority !== "transport" || candidate.result !== true
		|| repository(candidate.repository) !== query.repository
		|| sha(candidate.ancestorSha, "ancestry proof ancestor SHA") !== query.ancestorSha
		|| sha(candidate.descendantSha, "ancestry proof descendant SHA") !== query.descendantSha) {
		throw new Error("ancestry proof does not bind exact coordinates with a literal true result");
	}
	const observedAt = timestamp(candidate.observedAt, "ancestry proof observation timestamp");
	if (observedAt < minimumObservedAt) throw new Error("ancestry proof is stale");
	return {
		schemaVersion: 1,
		authority: "transport",
		repository: query.repository,
		ancestorSha: query.ancestorSha,
		descendantSha: query.descendantSha,
		result: true,
		revision: generation(candidate.revision),
		observedAt,
	};
}

function pathWithinScope(path: string, scope: string): boolean {
	return path === scope || path.startsWith(`${scope}/`);
}

function validChangedPath(path: unknown): path is string {
	return typeof path === "string" && path.length > 0 && Buffer.byteLength(path) <= 4_096
		&& !UNSAFE_INLINE.test(path) && !path.startsWith("/") && !path.includes("\\")
		&& !path.split("/").some((segment) => segment === "" || segment === "." || segment === "..");
}

function validateHandoff(
	value: WorkspaceHandoffEvidence,
	issue: number,
	branchName: string,
	prBase: string,
	allowedScopes: readonly string[],
): WorkspaceHandoffEvidence {
	const handoff = exactRecord(value, [
		"issue",
		"branch",
		"prBase",
		"baseHead",
		"head",
		"changedScope",
		"verificationState",
		"repositoryIdentity",
		"worktreeIdentity",
		"dirty",
	]);
	if (handoff.issue !== issue || handoff.branch !== branchName || handoff.prBase !== prBase) {
		throw new Error("workspace handoff issue, branch, or PR base mismatch");
	}
	const baseHead = sha(handoff.baseHead, "workspace handoff base head");
	const head = sha(handoff.head, "workspace handoff head");
	if (typeof handoff.repositoryIdentity !== "string" || typeof handoff.worktreeIdentity !== "string"
		|| !IDENTITY.test(handoff.repositoryIdentity) || !IDENTITY.test(handoff.worktreeIdentity)) {
		throw new Error("workspace handoff has invalid repository identity");
	}
	if (handoff.verificationState !== "passed" || handoff.dirty !== false) {
		throw new Error("workspace handoff is dirty or verification has not passed");
	}
	const changedScope = boundedArray(handoff.changedScope, "workspace changed scope", MAX_LIST, true);
	if (!changedScope.every(validChangedPath)
		|| changedScope.some((path) => !allowedScopes.some((scope) => pathWithinScope(path as string, scope)))) {
		throw new Error("workspace handoff contains an out-of-scope change");
	}
	return {
		issue,
		branch: branchName,
		prBase,
		baseHead,
		head,
		changedScope: (changedScope as string[]).sort(),
		verificationState: "passed",
		repositoryIdentity: handoff.repositoryIdentity,
		worktreeIdentity: handoff.worktreeIdentity,
		dirty: false,
	};
}

function aggregateScopes(plan: ParentOrchestrationPlan): string[] {
	return [...new Set(plan.children.flatMap((child) => child.writeScopes))].sort();
}

function sameStrings(left: readonly string[], right: readonly string[]): boolean {
	return left.length === right.length && left.every((value, index) => value === right[index]);
}

function canonicalDataEqual(
	left: unknown,
	right: unknown,
	leftSeen = new WeakSet<object>(),
	rightSeen = new WeakSet<object>(),
): boolean {
	if (Object.is(left, right)) return true;
	if (typeof left !== "object" || left === null || typeof right !== "object" || right === null
		|| nodeTypes.isProxy(left) || nodeTypes.isProxy(right)
		|| Array.isArray(left) !== Array.isArray(right)
		|| Object.getPrototypeOf(left) !== Object.getPrototypeOf(right)) return false;
	if (leftSeen.has(left) || rightSeen.has(right)) return false;
	leftSeen.add(left);
	rightSeen.add(right);
	if (Array.isArray(left) && Array.isArray(right)) {
		let leftValues: unknown[];
		let rightValues: unknown[];
		try {
			leftValues = boundedArray(left, "canonical left array", MAX_LIST * 8, true);
			rightValues = boundedArray(right, "canonical right array", MAX_LIST * 8, true);
		} catch {
			return false;
		}
		if (leftValues.length !== rightValues.length) return false;
		for (let index = 0; index < leftValues.length; index += 1) {
			if (!canonicalDataEqual(leftValues[index], rightValues[index], leftSeen, rightSeen)) return false;
		}
		return true;
	}
	let leftEntries: Array<[string, unknown]>;
	let rightEntries: Array<[string, unknown]>;
	try {
		leftEntries = boundedDataEntries(left, MAX_LIST * 8, "canonical object");
		rightEntries = boundedDataEntries(right, MAX_LIST * 8, "canonical object");
	} catch {
		return false;
	}
	if (leftEntries.length !== rightEntries.length) return false;
	const rightMap = new Map(rightEntries);
	for (const [key, entry] of leftEntries) {
		if (!rightMap.has(key) || !canonicalDataEqual(entry, rightMap.get(key), leftSeen, rightSeen)) return false;
	}
	return true;
}

function validateParentOrchestrationPlan(value: unknown): ParentOrchestrationPlan {
	const candidate = exactRecord(value, [
		"schemaVersion", "repository", "parentIssue", "generation", "title", "objective", "parentBranch",
		"parentBaseBranch", "markers", "requiredCheckPolicies", "children", "canonical",
	]);
	if (candidate.schemaVersion !== 1) throw new Error("unsupported canonical parent orchestration plan schema");
	const canonical = exactRecord(candidate.canonical, ["schemaVersion", "serialized", "digest"]);
	if (canonical.schemaVersion !== 1 || typeof canonical.serialized !== "string"
		|| canonical.serialized.length === 0 || Buffer.byteLength(canonical.serialized) > MAX_CANONICAL_PLAN_BYTES
		|| typeof canonical.digest !== "string" || !IDENTITY.test(canonical.digest)
		|| canonical.digest !== canonicalPlanDigest(canonical.serialized)) {
		throw new Error("canonical parent orchestration plan digest or serialization is invalid");
	}
	let seed: unknown;
	try {
		seed = JSON.parse(canonical.serialized);
	} catch {
		throw new Error("canonical parent orchestration plan serialization is invalid");
	}
	const parsed = exactRecord(seed, ["objective", "policyBundle"]);
	const rebuilt = buildParentOrchestrationPlan(parsed.objective, parsed.policyBundle);
	const canonicalSerialized = JSON.stringify({
		objective: canonicalObjective(rebuilt),
		policyBundle: canonicalPolicyBundle(rebuilt),
	});
	if (canonicalSerialized !== canonical.serialized) {
		throw new Error("canonical parent orchestration plan serialization is not normalized");
	}
	const supplied: CanonicalPlanData = {
		schemaVersion: candidate.schemaVersion as 1,
		repository: candidate.repository as string,
		parentIssue: candidate.parentIssue as number,
		generation: candidate.generation as number,
		title: candidate.title as string,
		objective: candidate.objective as string,
		parentBranch: candidate.parentBranch as string,
		parentBaseBranch: candidate.parentBaseBranch as string,
		markers: candidate.markers as ParentOrchestrationPlan["markers"],
		requiredCheckPolicies: candidate.requiredCheckPolicies as RequiredGitHubCheckPolicy[],
		children: candidate.children as BoundedChildRecord[],
	};
	if (!canonicalDataEqual(supplied, rebuilt)) {
		throw new Error("canonical parent orchestration plan provenance or derived topology was tampered");
	}
	return deepFreeze({
		...rebuilt,
		canonical: {
			schemaVersion: 1,
			serialized: canonical.serialized,
			digest: canonical.digest,
		},
	});
}

function validateMaterializedChild(plan: ParentOrchestrationPlan, value: unknown): MaterializedChildRecord {
	const candidate = exactRecord(value, [
		"id", "dependsOn", "status", "access", "writeScopes", "title", "objective", "issue", "branch", "prBase",
		"requiredSkills", "verification", "humanGates", "markers", "issueBody",
	]);
	const planned = childFor(plan, inlineText(candidate.id, "materialized child ID", 64));
	const issue = githubNumber(candidate.issue, "materialized child issue");
	const expectedBranch = canonicalIssueBranch(issue, planned.branch.slug);
	const comparisons: Array<[unknown, unknown]> = [
		[candidate.dependsOn, planned.dependsOn], [candidate.status, planned.status], [candidate.access, planned.access],
		[candidate.writeScopes, planned.writeScopes], [candidate.title, planned.title], [candidate.objective, planned.objective],
		[candidate.branch, expectedBranch], [candidate.prBase, planned.prBase], [candidate.requiredSkills, planned.requiredSkills],
		[candidate.verification, planned.verification], [candidate.humanGates, planned.humanGates],
		[candidate.markers, planned.markers], [candidate.issueBody, planned.issueBody],
	];
	if (comparisons.some(([actual, expected]) => !canonicalDataEqual(actual, expected))) {
		throw new Error("materialized child immutable topology does not match its parent orchestration plan");
	}
	return {
		...planned,
		issue,
		branch: expectedBranch,
	};
}

function policyFor(plan: ParentOrchestrationPlan, baseBranch: string): RequiredGitHubCheckPolicy {
	const matches = plan.requiredCheckPolicies.filter((policy) => policy.repository === plan.repository
		&& policy.baseBranch === baseBranch);
	if (matches.length !== 1) throw new Error("canonical plan does not contain one exact required-check policy");
	return matches[0];
}

function childPullRequestBody(plan: ParentOrchestrationPlan, child: MaterializedChildRecord): string {
	return `Refs #${child.issue}\nRefs #${plan.parentIssue}\n\n${child.markers.pullRequest}`;
}

interface ChildPullRequestTopology {
	baseHead: string;
	head: string;
	changedScope: readonly string[];
}

function currentChildPullRequestEligible(
	pullRequest: GitHubPullRequestEvidence,
	integratedState?: PullRequestObservation["state"],
): boolean {
	if (pullRequest.draft || pullRequest.state === "closed") return false;
	if (integratedState === undefined) return pullRequest.state === "open";
	if (integratedState === "open") return pullRequest.state === "open" || pullRequest.state === "merged";
	return integratedState === "merged" && pullRequest.state === "merged";
}

function childPullRequestMatches(
	pullRequest: GitHubPullRequestEvidence,
	plan: ParentOrchestrationPlan,
	child: MaterializedChildRecord,
	handoff: ChildPullRequestTopology,
	allowIntegratedBaseMovement = false,
): boolean {
	return pullRequest.repository === plan.repository
		&& pullRequest.workItemId === child.id
		&& pullRequest.generation === plan.generation
		&& pullRequest.policyDigest === policyFor(plan, child.prBase).digest
		&& sameStrings(pullRequest.allowedScopes, child.writeScopes)
		&& pullRequest.marker === child.markers.pullRequest
		&& pullRequest.title === child.title
		&& pullRequest.body === childPullRequestBody(plan, child)
		&& pullRequest.baseBranch === child.prBase
		&& pullRequest.headBranch === child.branch
		&& (pullRequest.baseSha === handoff.baseHead
			|| (allowIntegratedBaseMovement && pullRequest.state === "merged"))
		&& pullRequest.headSha === handoff.head
		&& sameStrings(pullRequest.changedPaths, handoff.changedScope)
		&& pullRequest.changedPaths.every((path) => child.writeScopes.some((scope) => pathWithinScope(path, scope)));
}

function reviewMatchesChild(
	review: IndependentReviewRecord,
	plan: ParentOrchestrationPlan,
	child: MaterializedChildRecord,
	pullRequest: number,
	handoff: ChildPullRequestTopology,
): boolean {
		return review.repository === plan.repository
		&& review.workItemId === child.id
		&& review.pullRequest === pullRequest
		&& review.generation === plan.generation
		&& review.baseBranch === child.prBase
		&& review.headBranch === child.branch
		&& review.baseSha === handoff.baseHead
		&& review.headSha === handoff.head
		&& sameStrings(review.changedPaths, handoff.changedScope)
		&& sameStrings(review.allowedScopes, child.writeScopes);
}

function childIntegrationPullRequestSnapshot(
	pullRequest: GitHubPullRequestEvidence,
	handoff: ChildPullRequestTopology,
): CanonicalPullRequestSnapshot {
	return createCanonicalPullRequestSnapshot(
		pullRequest.state === "merged" && pullRequest.baseSha !== handoff.baseHead
			? { ...pullRequest, baseSha: handoff.baseHead }
			: pullRequest,
	);
}

function childIntegrationMutationProjection(value: {
	repository: string;
	childId: string;
	pullRequest: number;
	generation: number;
	marker: string;
	baseSha: string;
	headSha: string;
	parentBranch: string;
	pullRequestSnapshot: CanonicalPullRequestSnapshot;
	controllerProvenance: ControllerIntegrationProvenance;
}): Record<string, unknown> {
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

function receiptMatchesChild(
	receipt: ChildIntegrationReceipt,
	plan: ParentOrchestrationPlan,
	child: MaterializedChildRecord,
	pullRequestNumber: number,
	handoff: ChildPullRequestTopology,
	pullRequestEvidence?: GitHubPullRequestEvidence,
): boolean {
	const expectedSnapshot = pullRequestEvidence === undefined
		? undefined
		: childIntegrationPullRequestSnapshot(pullRequestEvidence, handoff);
	const expectedMutation = createDurableMutationIntent(
		"child_integration",
		[plan.repository, child.markers.pullRequest],
		childIntegrationMutationProjection({
			repository: plan.repository,
			childId: child.id,
			pullRequest: pullRequestNumber,
			generation: plan.generation,
			marker: child.markers.pullRequest,
			baseSha: handoff.baseHead,
			headSha: handoff.head,
			parentBranch: plan.parentBranch,
			pullRequestSnapshot: receipt.pullRequestSnapshot,
			controllerProvenance: receipt.controllerProvenance,
		}),
		null,
	);
	return receipt.childId === child.id
		&& receipt.pullRequest === pullRequestNumber
		&& receipt.generation === plan.generation
		&& receipt.marker === child.markers.pullRequest
		&& receipt.baseSha === handoff.baseHead
		&& receipt.headSha === handoff.head
		&& receipt.parentBranch === plan.parentBranch
		&& receipt.controllerProvenance.authority === "controller"
		&& receipt.controllerProvenance.planDigest === plan.canonical.digest
		&& receipt.controllerProvenance.policyDigest === policyFor(plan, child.prBase).digest
		&& receipt.transportProvenance.idempotencyKey === expectedMutation.idempotencyKey
		&& receipt.transportProvenance.intentDigest === expectedMutation.intentDigest
		&& receipt.pullRequestSnapshot.repository === plan.repository
		&& receipt.pullRequestSnapshot.workItemId === child.id
		&& receipt.pullRequestSnapshot.number === pullRequestNumber
		&& receipt.pullRequestSnapshot.generation === plan.generation
		&& receipt.pullRequestSnapshot.marker === child.markers.pullRequest
		&& receipt.pullRequestSnapshot.baseBranch === child.prBase
		&& receipt.pullRequestSnapshot.headBranch === child.branch
		&& receipt.pullRequestSnapshot.baseSha === handoff.baseHead
		&& receipt.pullRequestSnapshot.headSha === handoff.head
		&& receipt.pullRequestSnapshot.policyDigest === policyFor(plan, child.prBase).digest
		&& sameStrings(receipt.pullRequestSnapshot.changedPaths, handoff.changedScope)
		&& sameStrings(receipt.pullRequestSnapshot.allowedScopes, child.writeScopes)
		&& receipt.observation.revision === receipt.pullRequestSnapshot.revision
		&& receipt.observation.observedAt === receipt.pullRequestSnapshot.observedAt
		&& (expectedSnapshot === undefined || receipt.pullRequestSnapshot.digest === expectedSnapshot.digest)
		&& (pullRequestEvidence === undefined || (
			pullRequestEvidence.revision >= receipt.observation.revision
			&& pullRequestEvidence.observedAt >= receipt.observation.observedAt
			&& currentChildPullRequestEligible(pullRequestEvidence, receipt.observation.state)
		));
}

function parentPullRequestBody(plan: ParentOrchestrationPlan): string {
	return `Closes #${plan.parentIssue}\n\n${plan.markers.parentPullRequest}`;
}

function parentPullRequestMatches(
	pullRequest: GitHubPullRequestEvidence,
	plan: ParentOrchestrationPlan,
): boolean {
	return pullRequest.repository === plan.repository
		&& pullRequest.workItemId === `parent-${plan.parentIssue}`
		&& pullRequest.generation === plan.generation
		&& pullRequest.policyDigest === policyFor(plan, plan.parentBaseBranch).digest
		&& sameStrings(pullRequest.allowedScopes, aggregateScopes(plan))
		&& pullRequest.marker === plan.markers.parentPullRequest
		&& pullRequest.title === plan.title
		&& pullRequest.body === parentPullRequestBody(plan)
		&& pullRequest.baseBranch === plan.parentBaseBranch
		&& pullRequest.headBranch === plan.parentBranch;
}

function reviewMatchesParent(
	review: IndependentReviewRecord,
	plan: ParentOrchestrationPlan,
	pullRequest: GitHubPullRequestEvidence,
): boolean {
	const scopes = aggregateScopes(plan);
	return review.repository === plan.repository
		&& review.workItemId === `parent-${plan.parentIssue}`
		&& review.pullRequest === pullRequest.number
		&& review.generation === plan.generation
		&& review.baseBranch === plan.parentBaseBranch
		&& review.headBranch === plan.parentBranch
		&& review.baseSha === pullRequest.baseSha
			&& review.headSha === pullRequest.headSha
			&& sameStrings(review.allowedScopes, scopes)
			&& sameStrings(review.changedPaths, pullRequest.changedPaths)
			&& review.changedPaths.every((path) => scopes.some((scope) => pathWithinScope(path, scope)));
}

function statusesForPlan(plan: ParentOrchestrationPlan, value: Readonly<Record<string, WorkItemStatus>>): Record<string, WorkItemStatus> {
	const expected = plan.children.map((child) => child.id).sort();
	const snapshot = exactRecord(value, expected);
	const keys = Object.keys(snapshot).sort();
	const allowed: readonly WorkItemStatus[] = ["pending", "running", "succeeded", "failed", "blocked"];
	for (const key of keys) if (!allowed.includes(snapshot[key] as WorkItemStatus)) throw new Error("invalid child roster status");
	return Object.fromEntries(keys.map((key) => [key, snapshot[key] as WorkItemStatus]));
}

function renderRoster(plan: ParentOrchestrationPlan, statuses: Readonly<Record<string, WorkItemStatus>>): string {
	const lines = plan.children.map((child) => `- ${child.id}: ${statuses[child.id]}`);
	return [`Shepherd child roster generation ${plan.generation}`, "", ...lines, "", plan.markers.roster].join("\n");
}

function assertMonotonicStatuses(previous: Readonly<Record<string, WorkItemStatus>>, next: Readonly<Record<string, WorkItemStatus>>): void {
	const rank: Record<WorkItemStatus, number> = { pending: 0, running: 1, succeeded: 2, failed: 2, blocked: 2 };
	for (const [child, prior] of Object.entries(previous)) {
		const current = next[child];
		if (current === undefined || rank[current] < rank[prior]
			|| (rank[prior] === 2 && current !== prior)) throw new Error("roster status regression is not allowed");
	}
}

function validateDecisionPolicy(value: ParentDecisionPolicy): ParentDecisionPolicy {
	const candidate = exactRecord(value, ["requestId", "actorAllowlist", "expiresAt", "question"]);
	const actors = validateStringList(candidate.actorAllowlist, "decision actor allowlist", /^[a-z\d](?:[a-z\d-]{0,37}[a-z\d])?$/);
	return {
		requestId: inlineText(candidate.requestId, "decision request ID", 128),
		actorAllowlist: actors,
		expiresAt: timestamp(candidate.expiresAt, "decision expiry"),
		question: inlineText(candidate.question, "decision question", 2_048),
	};
}

function sameDecisionRequest(record: HumanDecisionRecord, request: GitHubDecisionRequest): boolean {
	return record.requestId === request.requestId
		&& record.gate === request.gate
		&& canonicalDataEqual(record.binding, {
			repository: request.repository,
			target: { kind: "pull_request", number: request.pullRequest },
			generation: request.generation,
			headSha: request.headSha,
		})
		&& sameStrings(record.allowedOptions, request.allowedOptions)
		&& sameStrings(record.actorAllowlist, request.actorAllowlist)
		&& record.expiresAt === request.expiresAt
		&& record.question === request.question;
}

function validateBrokerRecord(
	value: unknown,
	request: GitHubDecisionRequest,
	binding: HumanDecisionBinding,
	prior?: HumanDecisionRecord,
	observedAt?: Date,
): HumanDecisionRecord {
	const record = validateHumanDecisionRecord(value, observedAt);
	assertHumanDecisionBinding(record, binding);
	if (!sameDecisionRequest(record, request)) throw new Error("human decision broker record does not match the exact request");
	if (prior !== undefined && (record.idempotencyMarker !== prior.idempotencyMarker
		|| record.createdAt !== prior.createdAt
		|| !canonicalDataEqual(record.requestComment, prior.requestComment))) {
		throw new Error("human decision broker record changed immutable request provenance");
	}
	return record;
}

export function validateParentReadyFreshness(value: unknown): ParentReadyFreshnessEnvelope {
	const candidate = exactRecord(value, [
		"schemaVersion", "policyObservations", "parentReviewResultDigest", "parentReviewCompletedAt",
		"parentPathRevision", "parentPathObservedAt", "childAncestryProofs", "digest",
	]);
	if (candidate.schemaVersion !== 1
		|| typeof candidate.parentReviewResultDigest !== "string"
		|| !IDENTITY.test(candidate.parentReviewResultDigest)) {
		throw new Error("invalid parent-ready freshness envelope");
	}
	const policyObservations = boundedArray(
		candidate.policyObservations,
		"parent-ready freshness policies",
		MAX_LIST,
		true,
	).map(validateRequiredGitHubCheckPolicyObservation);
	const childAncestryProofs = boundedArray(
		candidate.childAncestryProofs,
		"parent-ready freshness ancestry",
		MAX_CHILDREN,
		true,
	).map((entry): GitAncestryProof => {
		const proof = exactRecord(entry, [
			"schemaVersion", "authority", "repository", "ancestorSha", "descendantSha", "result", "revision", "observedAt",
		]);
		if (proof.schemaVersion !== 1 || proof.authority !== "transport" || proof.result !== true) {
			throw new Error("invalid parent-ready freshness ancestry proof");
		}
		return {
			schemaVersion: 1,
			authority: "transport",
			repository: repository(proof.repository),
			ancestorSha: sha(proof.ancestorSha, "parent-ready freshness ancestor SHA"),
			descendantSha: sha(proof.descendantSha, "parent-ready freshness descendant SHA"),
			result: true,
			revision: generation(proof.revision),
			observedAt: timestamp(proof.observedAt, "parent-ready freshness ancestry observation"),
		};
	});
	const payload = {
		schemaVersion: 1 as const,
		policyObservations,
		parentReviewResultDigest: candidate.parentReviewResultDigest,
		parentReviewCompletedAt: timestamp(candidate.parentReviewCompletedAt, "parent-ready review completion"),
		parentPathRevision: generation(candidate.parentPathRevision),
		parentPathObservedAt: timestamp(candidate.parentPathObservedAt, "parent-ready path observation"),
		childAncestryProofs,
	};
	const digest = canonicalDigest(payload);
	if (candidate.digest !== digest) throw new Error("parent-ready freshness digest mismatch");
	return { ...payload, parentReviewResultDigest: candidate.parentReviewResultDigest, digest };
}

export function validateMarkParentReadyRequest(value: unknown): MarkParentReadyRequest {
	const candidate = exactRecord(value, [
		"repository", "pullRequest", "marker", "headSha", "generation", "decisionRequestId", "authorization",
		"freshness", "mutation",
	]);
	const authorization = validateParentReadyAuthorization(candidate.authorization);
	const freshness = validateParentReadyFreshness(candidate.freshness);
	const mutation = validateDurableMutationIntent(candidate.mutation);
	const request = {
		repository: repository(candidate.repository),
		pullRequest: githubNumber(candidate.pullRequest, "parent-ready pull request"),
		marker: inlineText(candidate.marker, "parent-ready pull request marker", 512),
		headSha: sha(candidate.headSha, "parent-ready request head SHA"),
		generation: generation(candidate.generation),
		decisionRequestId: inlineText(candidate.decisionRequestId, "parent-ready decision request ID", 128),
		authorization,
		freshness,
		mutation,
	};
	if (authorization.repository !== request.repository
		|| authorization.pullRequest !== request.pullRequest
		|| authorization.headSha !== request.headSha
		|| authorization.generation !== request.generation
		|| authorization.decisionRequestId !== request.decisionRequestId
		|| mutation.operation !== "parent_ready"
		|| mutation.expectedResourceRevision !== authorization.pullRequestRevision) {
		throw new Error("parent-ready request authority coordinates do not match");
	}
	const stableFreshnessPolicies = freshness.policyObservations.map(parentReadyPolicyAuthority)
		.sort((left, right) => left.baseBranch.localeCompare(right.baseBranch));
	const stableFreshnessAncestry = freshness.childAncestryProofs.map(parentReadyAncestryAuthority);
	if (!canonicalDataEqual(stableFreshnessPolicies, authorization.policies)
		|| stableFreshnessAncestry.length !== authorization.children.length
		|| stableFreshnessAncestry.some((proof, index) => !canonicalDataEqual(proof, authorization.children[index].ancestry))) {
		throw new Error("parent-ready freshness does not match stable authority");
	}
	const expectedMutation = createDurableMutationIntent(
		"parent_ready",
		[authorization.repository, request.marker, authorization.headSha],
		{
			repository: authorization.repository,
			pullRequest: authorization.pullRequest,
			headSha: authorization.headSha,
			generation: authorization.generation,
			decisionRequestId: authorization.decisionRequestId,
			authorization,
		},
		authorization.pullRequestRevision,
	);
	if (!canonicalDataEqual(mutation, expectedMutation)) {
		throw new Error("parent-ready request mutation does not match its authorization");
	}
	return request;
}

export function validateParentReadyAuthorityQuery(value: unknown): ParentReadyAuthorityQuery {
	const candidate = exactRecord(value, ["repository", "pullRequest", "marker", "headSha", "generation"]);
	return {
		repository: repository(candidate.repository),
		pullRequest: githubNumber(candidate.pullRequest, "parent-ready authority pull request"),
		marker: inlineText(candidate.marker, "parent-ready authority marker", 512),
		headSha: sha(candidate.headSha, "parent-ready authority head SHA"),
		generation: generation(candidate.generation),
	};
}

function expectedParentReadyMutation(
	authorization: ParentReadyAuthorization,
	marker: string,
): DurableMutationIntent {
	return createDurableMutationIntent(
		"parent_ready",
		[authorization.repository, marker, authorization.headSha],
		{
			repository: authorization.repository,
			pullRequest: authorization.pullRequest,
			headSha: authorization.headSha,
			generation: authorization.generation,
			decisionRequestId: authorization.decisionRequestId,
			authorization,
		},
		authorization.pullRequestRevision,
	);
}

function parentReadyInvocationId(
	query: ParentReadyAuthorityQuery,
	authorizationDigest: string,
	readyMutation: DurableMutationIntent,
): string {
	return canonicalDigest({ schemaVersion: 1, ...query, authorizationDigest, readyMutation });
}

function parentReadyRecoveryId(
	query: ParentReadyAuthorityQuery,
	invocationId: string,
	authorizationDigest: string,
	readyMutation: DurableMutationIntent,
): string {
	return canonicalDigest({
		schemaVersion: 1,
		...query,
		invocationId,
		authorizationDigest,
		readyMutation,
	});
}

export function validateParentReadyRecoveryFence(value: unknown): ParentReadyRecoveryFence {
	const candidate = exactRecord(value, [
		"schemaVersion", "invocationId", "recoveryId", "attempt", "supersedesAttempt", "readyMutation",
	]);
	const readyMutation = validateDurableMutationIntent(candidate.readyMutation);
	const attempt = generation(candidate.attempt);
	const supersedesAttempt = candidate.supersedesAttempt === null
		? null
		: generation(candidate.supersedesAttempt);
	if (candidate.schemaVersion !== 1
		|| typeof candidate.invocationId !== "string" || !IDENTITY.test(candidate.invocationId)
		|| typeof candidate.recoveryId !== "string" || !IDENTITY.test(candidate.recoveryId)
		|| readyMutation.operation !== "parent_ready"
		|| (attempt === 1 ? supersedesAttempt !== null : supersedesAttempt !== attempt - 1)) {
		throw new Error("invalid parent-ready recovery fence");
	}
	return {
		schemaVersion: 1,
		invocationId: candidate.invocationId,
		recoveryId: candidate.recoveryId,
		attempt,
		supersedesAttempt,
		readyMutation,
	};
}

export function validateParentReadyAuthorityState(value: unknown): ParentReadyAuthorityState {
	const candidate = exactRecord(value, [
		"schemaVersion", "invocationId", "recoveryId", "repository", "pullRequest", "marker", "generation",
		"headSha", "originalRevision", "appliedRevision", "authorization", "readyMutation", "rollbackMutation", "phase",
		"status", "fence",
	]);
	const query = validateParentReadyAuthorityQuery({
		repository: candidate.repository,
		pullRequest: candidate.pullRequest,
		marker: candidate.marker,
		headSha: candidate.headSha,
		generation: candidate.generation,
	});
	const authorization = validateParentReadyAuthorization(candidate.authorization);
	const readyMutation = validateDurableMutationIntent(candidate.readyMutation);
	const rollbackMutation = candidate.rollbackMutation === null
		? null
		: validateDurableMutationIntent(candidate.rollbackMutation);
	const originalRevision = generation(candidate.originalRevision);
	const appliedRevision = candidate.appliedRevision === null ? null : generation(candidate.appliedRevision);
	const fence = candidate.fence === 0 ? 0 : generation(candidate.fence);
	if (candidate.schemaVersion !== 1
		|| typeof candidate.invocationId !== "string" || !IDENTITY.test(candidate.invocationId)
		|| typeof candidate.recoveryId !== "string" || !IDENTITY.test(candidate.recoveryId)
		|| !["ready_invoking", "ready_effect_applied", "ready_settled", "recovery_claimed", "draft_restored"]
			.includes(String(candidate.phase))
		|| !["unsettled", "settled"].includes(String(candidate.status))
		|| authorization.repository !== query.repository
		|| authorization.pullRequest !== query.pullRequest
		|| authorization.generation !== query.generation
		|| authorization.headSha !== query.headSha
		|| authorization.pullRequestRevision !== originalRevision
		|| !canonicalDataEqual(readyMutation, expectedParentReadyMutation(authorization, query.marker))) {
		throw new Error("invalid parent-ready durable authority state");
	}
	const invocationId = parentReadyInvocationId(query, authorization.digest, readyMutation);
	const recoveryId = parentReadyRecoveryId(query, invocationId, authorization.digest, readyMutation);
	const phase = candidate.phase as ParentReadyAuthorityPhase;
	const status = candidate.status as ParentReadyAuthorityState["status"];
	const recovering = phase === "recovery_claimed" || phase === "draft_restored";
	const settled = phase === "ready_settled" || phase === "draft_restored";
	const applied = phase === "ready_effect_applied" || phase === "ready_settled";
	if (candidate.invocationId !== invocationId
		|| candidate.recoveryId !== recoveryId
		|| status !== (settled ? "settled" : "unsettled")
		|| (phase === "ready_invoking" && appliedRevision !== null)
		|| (applied && (appliedRevision === null || appliedRevision <= originalRevision))
		|| (recovering && appliedRevision !== null && appliedRevision <= originalRevision)
		|| (recovering ? fence < 1 || rollbackMutation === null : fence !== 0 || rollbackMutation !== null)) {
		throw new Error("incoherent parent-ready durable authority phase");
	}
	if (rollbackMutation !== null) {
		validateRollbackParentReadyRequest({
			...query,
			authorizationDigest: authorization.digest,
			recovery: {
				schemaVersion: 1,
				invocationId,
				recoveryId,
				attempt: fence,
				supersedesAttempt: fence === 1 ? null : fence - 1,
				readyMutation,
			},
			mutation: rollbackMutation,
		});
	}
	return {
		schemaVersion: 1,
		...query,
		invocationId,
		recoveryId,
		originalRevision,
		appliedRevision,
		authorization,
		readyMutation,
		rollbackMutation,
		phase,
		status,
		fence,
	};
}

export function createParentReadyInvokingAuthorityState(
	requestValue: MarkParentReadyRequest,
): ParentReadyAuthorityState {
	const request = validateMarkParentReadyRequest(requestValue);
	const query = validateParentReadyAuthorityQuery({
		repository: request.repository,
		pullRequest: request.pullRequest,
		marker: request.marker,
		headSha: request.headSha,
		generation: request.generation,
	});
	const invocationId = parentReadyInvocationId(query, request.authorization.digest, request.mutation);
	return validateParentReadyAuthorityState({
		schemaVersion: 1,
		...query,
		invocationId,
		recoveryId: parentReadyRecoveryId(query, invocationId, request.authorization.digest, request.mutation),
		originalRevision: request.authorization.pullRequestRevision,
		appliedRevision: null,
		authorization: request.authorization,
		readyMutation: request.mutation,
		rollbackMutation: null,
		phase: "ready_invoking",
		status: "unsettled",
		fence: 0,
	});
}

export function createParentReadyConflictTombstone(
	requestValue: MarkParentReadyRequest,
): ParentReadyConflictTombstone {
	const request = validateMarkParentReadyRequest(requestValue);
	const query = validateParentReadyAuthorityQuery({
		repository: request.repository,
		pullRequest: request.pullRequest,
		marker: request.marker,
		headSha: request.headSha,
		generation: request.generation,
	});
	return {
		schemaVersion: 1,
		kind: "tombstoned",
		...query,
		invocationId: parentReadyInvocationId(query, request.authorization.digest, request.mutation),
		authorizationDigest: request.authorization.digest,
		mutationIdempotencyKey: request.mutation.idempotencyKey,
		mutationIntentDigest: request.mutation.intentDigest,
	};
}

export function validateParentReadyConflictTombstone(
	value: unknown,
	requestValue: MarkParentReadyRequest,
): ParentReadyConflictTombstone {
	const request = validateMarkParentReadyRequest(requestValue);
	let candidate: Record<string, unknown>;
	try {
		candidate = exactRecord(value, [
			"schemaVersion", "kind", "repository", "pullRequest", "marker", "headSha", "generation",
			"invocationId", "authorizationDigest", "mutationIdempotencyKey", "mutationIntentDigest",
		]);
	} catch {
		throw new Error("parent-ready conflict lacks exact atomic invocation tombstone proof");
	}
	const expected = createParentReadyConflictTombstone(request);
	if (candidate.schemaVersion !== 1 || candidate.kind !== "tombstoned"
		|| !canonicalDataEqual(candidate, expected)) {
		throw new Error("parent-ready conflict lacks exact atomic invocation tombstone proof");
	}
	return expected;
}

export function validateSettleParentReadyAuthorityRequest(value: unknown): SettleParentReadyAuthorityRequest {
	const candidate = exactRecord(value, [
		"repository", "pullRequest", "marker", "headSha", "generation", "invocationId", "authorizationDigest",
		"readyMutation", "expectedPhase", "expectedFence",
	]);
	const query = validateParentReadyAuthorityQuery({
		repository: candidate.repository,
		pullRequest: candidate.pullRequest,
		marker: candidate.marker,
		headSha: candidate.headSha,
		generation: candidate.generation,
	});
	const readyMutation = validateDurableMutationIntent(candidate.readyMutation);
	if (typeof candidate.authorizationDigest !== "string" || !IDENTITY.test(candidate.authorizationDigest)
		|| typeof candidate.invocationId !== "string" || !IDENTITY.test(candidate.invocationId)
		|| candidate.invocationId !== parentReadyInvocationId(query, candidate.authorizationDigest, readyMutation)
		|| candidate.expectedPhase !== "ready_effect_applied" || candidate.expectedFence !== 0) {
		throw new Error("invalid parent-ready authority settlement request");
	}
	return {
		...query,
		invocationId: candidate.invocationId,
		authorizationDigest: candidate.authorizationDigest,
		readyMutation,
		expectedPhase: "ready_effect_applied",
		expectedFence: 0,
	};
}

export function validateRollbackParentReadyRequest(value: unknown): RollbackParentReadyRequest {
	const candidate = exactRecord(value, [
		"repository", "pullRequest", "marker", "headSha", "generation", "authorizationDigest", "recovery", "mutation",
	]);
	const request = {
		repository: repository(candidate.repository),
		pullRequest: githubNumber(candidate.pullRequest, "parent-ready rollback pull request"),
		marker: inlineText(candidate.marker, "parent-ready rollback marker", 512),
		headSha: sha(candidate.headSha, "parent-ready rollback head SHA"),
		generation: generation(candidate.generation),
		authorizationDigest: typeof candidate.authorizationDigest === "string" && IDENTITY.test(candidate.authorizationDigest)
			? candidate.authorizationDigest
			: "",
	};
	if (request.authorizationDigest.length === 0) throw new Error("invalid parent-ready rollback authorization digest");
	const recovery = validateParentReadyRecoveryFence(candidate.recovery);
	const query = validateParentReadyAuthorityQuery({
		repository: request.repository,
		pullRequest: request.pullRequest,
		marker: request.marker,
		headSha: request.headSha,
		generation: request.generation,
	});
	const invocationId = parentReadyInvocationId(query, request.authorizationDigest, recovery.readyMutation);
	if (recovery.invocationId !== invocationId
		|| recovery.recoveryId !== parentReadyRecoveryId(
			query, invocationId, request.authorizationDigest, recovery.readyMutation,
		)) {
		throw new Error("parent-ready recovery identity does not match its original ready mutation");
	}
	const mutation = validateDurableMutationIntent(candidate.mutation);
	const expectedMutation = createDurableMutationIntent(
		"parent_ready_rollback",
		[recovery.recoveryId, recovery.attempt],
		{ ...request, recovery },
		null,
	);
	if (mutation.operation !== "parent_ready_rollback"
		|| mutation.expectedResourceRevision !== null
		|| !canonicalDataEqual(mutation, expectedMutation)) {
		throw new Error("parent-ready rollback mutation does not match its recovery fence");
	}
	return { ...request, recovery, mutation };
}

export function validateParentReadyCompareEffectResult(
	value: unknown,
	requestValue: MarkParentReadyRequest,
): ParentReadyCompareEffectResult {
	const request = validateMarkParentReadyRequest(requestValue);
	const candidate = exactRecord(value, ["schemaVersion", "kind"], ["mutation", "coordinate", "terminal"]);
	if (candidate.schemaVersion !== 1) throw new Error("invalid parent-ready compare/effect result schema");
	if (candidate.kind === "conflict") {
		if (candidate.mutation !== undefined
			|| typeof candidate.coordinate !== "string"
			|| !PARENT_READY_AUTHORITY_COORDINATES.includes(candidate.coordinate as ParentReadyAuthorityCoordinate)) {
			throw new Error("invalid parent-ready non-applied conflict result");
		}
		const terminal = validateParentReadyConflictTombstone(candidate.terminal, request);
		return {
			schemaVersion: 1,
			kind: "conflict",
			coordinate: candidate.coordinate as ParentReadyAuthorityCoordinate,
			terminal,
		};
	}
	if (candidate.kind !== "applied" || candidate.coordinate !== undefined || candidate.terminal !== undefined) {
		throw new Error("invalid parent-ready compare/effect result");
	}
	const mutation = validateDurableMutationResult(
		candidate.mutation,
		request.mutation,
		validateGitHubPullRequestEvidence,
	);
	if (mutation.value.draft
		|| mutation.value.repository !== request.repository
		|| mutation.value.number !== request.pullRequest
		|| mutation.value.headSha !== request.headSha) {
		throw new Error("parent-ready compare/effect result did not apply the exact conditional transition");
	}
	if (mutation.value.revision <= request.authorization.pullRequestRevision) {
		throw new Error("parent ready CAS revision did not advance");
	}
	return { schemaVersion: 1, kind: "applied", mutation };
}

export function validateParentReadySettlementRecord(value: unknown): ParentReadySettlementRecord {
	const candidate = exactRecord(value, [
		"schemaVersion", "planDigest", "authorizationDigest", "mutationIdempotencyKey", "outcome", "settledAt",
	]);
	if (candidate.schemaVersion !== 1
		|| typeof candidate.planDigest !== "string" || !IDENTITY.test(candidate.planDigest)
		|| typeof candidate.authorizationDigest !== "string" || !IDENTITY.test(candidate.authorizationDigest)
		|| !["ready", "blocked"].includes(String(candidate.outcome))) {
		throw new Error("invalid parent-ready settlement record");
	}
	return {
		schemaVersion: 1,
		planDigest: candidate.planDigest,
		authorizationDigest: candidate.authorizationDigest,
		mutationIdempotencyKey: inlineText(candidate.mutationIdempotencyKey, "parent-ready settlement mutation key", 512),
		outcome: candidate.outcome as ParentReadySettlementRecord["outcome"],
		settledAt: timestamp(candidate.settledAt, "parent-ready settlement time"),
	};
}

export function validateDurableMutationIntent(value: unknown): DurableMutationIntent {
	const candidate = exactRecord(value, [
		"schemaVersion", "operation", "idempotencyKey", "intentDigest", "expectedResourceRevision",
	]);
	if (candidate.schemaVersion !== 1
		|| !["child_issue", "pull_request", "parent_roster", "child_integration", "parent_ready", "parent_ready_rollback"].includes(String(candidate.operation))
		|| typeof candidate.intentDigest !== "string" || !IDENTITY.test(candidate.intentDigest)) {
		throw new Error("invalid durable mutation intent");
	}
	return {
		schemaVersion: 1,
		operation: candidate.operation as DurableMutationOperation,
		idempotencyKey: inlineText(candidate.idempotencyKey, "durable mutation key", 512),
		intentDigest: candidate.intentDigest,
		expectedResourceRevision: candidate.expectedResourceRevision === null
			? null
			: generation(candidate.expectedResourceRevision),
	};
}

export function validatePreparedParentReadyOperation(
	value: unknown,
	plan: ParentOrchestrationPlan,
): PreparedParentReadyOperation {
	const candidate = exactRecord(value, [
		"schemaVersion", "planDigest", "policy", "decision", "authorization", "freshness", "mutation",
	]);
	if (candidate.schemaVersion !== 1 || candidate.planDigest !== plan.canonical.digest) {
		throw new Error("prepared parent-ready operation does not match the canonical plan");
	}
	const policy = validateDecisionPolicy(candidate.policy as ParentDecisionPolicy);
	const decision = validateHumanDecisionRecord(candidate.decision);
	const authorization = validateParentReadyAuthorization(candidate.authorization);
	const freshness = validateParentReadyFreshness(candidate.freshness);
	const mutation = validateDurableMutationIntent(candidate.mutation);
	if (authorization.repository !== plan.repository
		|| authorization.parentIssue !== plan.parentIssue
		|| authorization.generation !== plan.generation
		|| authorization.planDigest !== plan.canonical.digest
		|| authorization.decisionRequestId !== policy.requestId
		|| !canonicalDataEqual(authorization.decision, decision)
		|| decision.status !== "consumed"
		|| decision.gate !== "parent_merge"
		|| decision.decision?.option !== "approve-merge"
		|| decision.binding.repository !== authorization.repository
		|| decision.binding.target.kind !== "pull_request"
		|| decision.binding.target.number !== authorization.pullRequest
		|| decision.binding.generation !== authorization.generation
		|| decision.binding.headSha !== authorization.headSha
		|| authorization.decisionDigest !== canonicalDigest(decision)
		|| mutation.operation !== "parent_ready"
		|| mutation.expectedResourceRevision !== authorization.pullRequestRevision) {
		throw new Error("prepared parent-ready operation decision/authorization coordinates do not match");
	}
	const expected = createDurableMutationIntent(
		"parent_ready",
		[plan.repository, plan.markers.parentPullRequest, authorization.headSha],
		{
			repository: authorization.repository,
			pullRequest: authorization.pullRequest,
			headSha: authorization.headSha,
			generation: authorization.generation,
			decisionRequestId: authorization.decisionRequestId,
			authorization,
		},
		authorization.pullRequestRevision,
	);
	if (!canonicalDataEqual(mutation, expected)) throw new Error("prepared parent-ready mutation identity mismatch");
	return {
		schemaVersion: 1,
		planDigest: plan.canonical.digest,
		policy,
		decision,
		authorization,
		freshness,
		mutation,
	};
}

interface ParentReadyRestartMutationRecord {
	key: string;
	digest: string;
	value: GitHubPullRequestEvidence;
	revision: number;
}

function parentReadyRestartMutationRecords(value: unknown, description: string): ParentReadyRestartMutationRecord[] {
	const keys = new Set<string>();
	return boundedArray(value, `${description} collection`, 32, true).map((entry) => {
		if (!Array.isArray(entry) || entry.length !== 2) throw new Error(`${description} tuple is invalid`);
		const key = inlineText(entry[0], `${description} key`, 512);
		if (keys.has(key)) throw new Error(`${description} key is duplicated`);
		keys.add(key);
		const record = exactRecord(entry[1], ["digest", "value", "revision"]);
		if (typeof record.digest !== "string" || !IDENTITY.test(record.digest)) {
			throw new Error(`${description} digest is invalid`);
		}
		return {
			key,
			digest: record.digest,
			value: validateGitHubPullRequestEvidence(record.value),
			revision: generation(record.revision),
		};
	});
}

interface ParentReadyRestartRecoveryAttempt {
	recoveryId: string;
	attempt: number;
}

function parentReadyRestartRecoveryAttempts(value: unknown): ParentReadyRestartRecoveryAttempt[] {
	const recoveryIds = new Set<string>();
	return boundedArray(value, "restart recovery attempts", 32, true).map((entry) => {
		if (!Array.isArray(entry) || entry.length !== 2) throw new Error("recovery attempt tuple is invalid");
		if (typeof entry[0] !== "string" || !IDENTITY.test(entry[0])) {
			throw new Error("recovery attempt ID is invalid");
		}
		if (recoveryIds.has(entry[0])) throw new Error("recovery attempt ID is duplicated");
		recoveryIds.add(entry[0]);
		return { recoveryId: entry[0], attempt: generation(entry[1]) };
	});
}

interface ParentReadyRestartAuthorityRecord {
	key: string;
	state: ParentReadyAuthorityState;
}

interface ParentReadyRestartGraphNode {
	operation: PreparedParentReadyOperation;
	settlement?: ParentReadySettlementRecord;
	current?: GitHubPullRequestEvidence;
	state?: ParentReadyAuthorityState;
	ready?: ParentReadyRestartMutationRecord;
	rollback?: ParentReadyRestartMutationRecord;
	recovery?: ParentReadyRestartRecoveryAttempt;
}

export interface ParentReadyRestartValidationResult {
	settlementRepairs: ParentReadySettlementRecord[];
}

function assertParentReadyRestartHistory(
	value: unknown,
	planValue: ParentOrchestrationPlan,
): ParentReadyRestartValidationResult {
	const plan = validateParentOrchestrationPlan(planValue);
	const record = exactRecord(value, [
		"schemaVersion", "pullRequests", "prepared", "settlements", "readyMutations", "rollbackMutations",
		"recoveryAttempts", "mutationRevision", "states", "decision",
	]);
	if (record.schemaVersion !== 1) throw new Error("restart history schema is invalid");
	const pullRequests = boundedArray(record.pullRequests, "restart pull requests", MAX_LIST, true)
		.map(validateGitHubPullRequestEvidence);
	const canonicalParentOwners = pullRequests.filter((pullRequest) =>
		pullRequest.repository === plan.repository
		&& pullRequest.marker === plan.markers.parentPullRequest
		&& pullRequest.generation === plan.generation);
	if (canonicalParentOwners.length !== 1) {
		throw new Error("restart history requires one unique current canonical parent marker owner");
	}
	const prepared = boundedArray(record.prepared, "restart prepared operations", 16, true)
		.map((entry) => validatePreparedParentReadyOperation(entry, plan));
	const settlements = boundedArray(record.settlements, "restart settlements", 16, true)
		.map(validateParentReadySettlementRecord);
	const readyMutations = parentReadyRestartMutationRecords(record.readyMutations, "ready mutation");
	const rollbackMutations = parentReadyRestartMutationRecords(record.rollbackMutations, "rollback mutation");
	const recoveryAttempts = parentReadyRestartRecoveryAttempts(record.recoveryAttempts);
	const mutationRevision = record.mutationRevision;
	if (!Number.isSafeInteger(mutationRevision) || (mutationRevision as number) < 0
		|| (mutationRevision as number) > MAX_GITHUB_NUMBER) {
		throw new Error("restart mutation high-water revision is invalid");
	}
	const stateKeys = new Set<string>();
	const states: ParentReadyRestartAuthorityRecord[] = boundedArray(
		record.states,
		"restart authority states",
		16,
		true,
	).map((entry) => {
		if (!Array.isArray(entry) || entry.length !== 2) throw new Error("authority state tuple is invalid");
		const state = validateParentReadyAuthorityState(entry[1]);
		const key = `${state.repository}\u0000${state.pullRequest}\u0000${state.marker}\u0000${state.generation}\u0000${state.headSha}`;
		if (entry[0] !== key) throw new Error("authority state key does not match its durable coordinates");
		if (stateKeys.has(key)) throw new Error("authority state key is duplicated");
		stateKeys.add(key);
		return { key, state };
	});
	const decision = validateHumanDecisionRecord(record.decision);
	const settlementKey = (entry: ParentReadyJournalQuery): string =>
		`${entry.planDigest}\u0000${entry.authorizationDigest}\u0000${entry.mutationIdempotencyKey}`;
	const nodes = prepared.map((operation): ParentReadyRestartGraphNode => ({ operation }));
	const nodeBySettlementKey = new Map<string, ParentReadyRestartGraphNode>();
	const nodeByInvocation = new Map<string, ParentReadyRestartGraphNode>();
	const nodeByReadyKey = new Map<string, ParentReadyRestartGraphNode>();
	for (const node of nodes) {
		const operation = node.operation;
		if (!canonicalDataEqual(decision, operation.decision)) {
			throw new Error("top-level decision diverges from its prepared operation");
		}
		const query: ParentReadyAuthorityQuery = {
			repository: operation.authorization.repository,
			pullRequest: operation.authorization.pullRequest,
			marker: plan.markers.parentPullRequest,
			headSha: operation.authorization.headSha,
			generation: operation.authorization.generation,
		};
		const invocationId = parentReadyInvocationId(query, operation.authorization.digest, operation.mutation);
		const key = settlementKey({
			planDigest: operation.planDigest,
			authorizationDigest: operation.authorization.digest,
			mutationIdempotencyKey: operation.mutation.idempotencyKey,
		});
		if (nodeBySettlementKey.has(key) || nodeByInvocation.has(invocationId)
			|| nodeByReadyKey.has(operation.mutation.idempotencyKey)) {
			throw new Error("prepared restart history is duplicated or ambiguously owned");
		}
		nodeBySettlementKey.set(key, node);
		nodeByInvocation.set(invocationId, node);
		nodeByReadyKey.set(operation.mutation.idempotencyKey, node);
	}

	for (const settlement of settlements) {
		const node = nodeBySettlementKey.get(settlementKey(settlement));
		if (node === undefined || node.settlement !== undefined) {
			throw new Error("settlement role is orphaned from exactly one prepared history");
		}
		node.settlement = settlement;
	}
	for (const authority of states) {
		const node = nodeByInvocation.get(authority.state.invocationId);
		if (node === undefined || node.state !== undefined) {
			throw new Error("authority role is orphaned from exactly one prepared history");
		}
		node.state = authority.state;
	}
	for (const mutation of readyMutations) {
		const node = nodeByReadyKey.get(mutation.key);
		if (node === undefined || node.state === undefined || node.ready !== undefined) {
			throw new Error("ready receipt is orphaned from exactly one prepared history");
		}
		node.ready = mutation;
	}
	const nodeByRollbackKey = new Map<string, ParentReadyRestartGraphNode>();
	const nodeByRecoveryId = new Map<string, ParentReadyRestartGraphNode>();
	for (const node of nodes) {
		const state = node.state;
		if (state?.rollbackMutation !== null && state?.rollbackMutation !== undefined) {
			if (nodeByRollbackKey.has(state.rollbackMutation.idempotencyKey)) {
				throw new Error("rollback intent is ambiguously owned by prepared histories");
			}
			nodeByRollbackKey.set(state.rollbackMutation.idempotencyKey, node);
		}
		if (state !== undefined && ["recovery_claimed", "draft_restored"].includes(state.phase)) {
			if (nodeByRecoveryId.has(state.recoveryId)) {
				throw new Error("recovery identity is ambiguously owned by prepared histories");
			}
			nodeByRecoveryId.set(state.recoveryId, node);
		}
	}
	for (const mutation of rollbackMutations) {
		const node = nodeByRollbackKey.get(mutation.key);
		if (node === undefined || node.rollback !== undefined) {
			throw new Error("rollback receipt is orphaned from exactly one prepared history");
		}
		node.rollback = mutation;
	}
	for (const recovery of recoveryAttempts) {
		const node = nodeByRecoveryId.get(recovery.recoveryId);
		if (node === undefined || node.recovery !== undefined) {
			throw new Error("recovery attempt is orphaned from exactly one prepared history");
		}
		node.recovery = recovery;
	}

	const pullRequestOwners = new Array<number>(pullRequests.length).fill(0);
	for (const node of nodes) {
		const operation = node.operation;
		const currentPullRequests = pullRequests
			.map((pullRequest, index) => ({ pullRequest, index }))
			.filter(({ pullRequest }) =>
			pullRequest.repository === operation.authorization.repository
			&& pullRequest.number === operation.authorization.pullRequest);
		if (currentPullRequests.length !== 1) {
			throw new Error("restart history requires exactly one current parent pull request");
		}
		const { pullRequest: current, index: currentIndex } = currentPullRequests[0];
		pullRequestOwners[currentIndex] += 1;
		node.current = current;
		if (current.marker !== plan.markers.parentPullRequest
			|| current.generation !== operation.authorization.generation
			|| current.headSha !== operation.authorization.headSha) {
			throw new Error("current parent pull request visibility diverges from authority coordinates");
		}
	}
	for (let index = 0; index < pullRequests.length; index += 1) {
		const current = pullRequests[index];
		if (current.repository === plan.repository && current.marker === plan.markers.parentPullRequest
			&& current.generation === plan.generation && pullRequestOwners[index] !== 1) {
			throw new Error("current pull request role is orphaned from exactly one prepared history");
		}
	}

	const sequenceRevisions = new Set<number>();
	for (const mutation of [...readyMutations, ...rollbackMutations]) {
		if (sequenceRevisions.has(mutation.revision)) {
			throw new Error("global causal mutation revisions must be globally unique");
		}
		sequenceRevisions.add(mutation.revision);
	}
	const greatestStoredRevision = [...sequenceRevisions].reduce((greatest, revision) =>
		Math.max(greatest, revision), 0);
	if ((mutationRevision as number) < greatestStoredRevision) {
		throw new Error("mutation high-water regresses behind the causal mutation sequence");
	}

	const settlementRepairs: ParentReadySettlementRecord[] = [];
	for (const node of nodes) {
		const operation = node.operation;
		const current = node.current!;
		const settlement = node.settlement;
		const state = node.state;
		if (settlement !== undefined
			&& (operation.decision.status !== "consumed" || operation.decision.consumedAt === undefined
				|| new Date(settlement.settledAt).valueOf() < new Date(operation.decision.consumedAt).valueOf())) {
			throw new Error("settlement is not causally after its consumed decision");
		}
		if (state === undefined) {
			if (node.ready !== undefined || node.rollback !== undefined || node.recovery !== undefined) {
				throw new Error("authority-free history retains an orphan mutation or recovery role");
			}
			if (settlement?.outcome === "ready" || !current.draft
				|| current.revision !== operation.authorization.pullRequestRevision) {
				throw new Error("authority-free settlement has incoherent parent visibility or revision");
			}
			continue;
		}
		if (state.authorization.digest !== operation.authorization.digest
			|| !canonicalDataEqual(state.readyMutation, operation.mutation)) {
			throw new Error("restart authority does not own its prepared operation");
		}
		const readyMutation = node.ready;
		const rollbackMutation = node.rollback;
		const recoveryAttempt = node.recovery;
		if (readyMutation !== undefined) {
			if (state.appliedRevision === null || readyMutation.digest !== state.readyMutation.intentDigest
				|| readyMutation.value.draft || readyMutation.value.repository !== state.repository
				|| readyMutation.value.number !== state.pullRequest || readyMutation.value.marker !== state.marker
				|| readyMutation.value.generation !== state.generation || readyMutation.value.headSha !== state.headSha
				|| readyMutation.value.revision !== state.appliedRevision) {
				throw new Error("ready receipt and authority visibility diverge");
			}
		} else if (state.appliedRevision !== null) {
			throw new Error("applied authority is missing its exact ready receipt");
		}
		if (rollbackMutation !== undefined) {
			const expectedRestoredRevision = state.appliedRevision === null
				? state.originalRevision
				: state.appliedRevision + 1;
			if (state.rollbackMutation === null
				|| rollbackMutation.digest !== state.rollbackMutation.intentDigest
				|| !rollbackMutation.value.draft || rollbackMutation.value.repository !== state.repository
				|| rollbackMutation.value.number !== state.pullRequest || rollbackMutation.value.marker !== state.marker
				|| rollbackMutation.value.generation !== state.generation || rollbackMutation.value.headSha !== state.headSha
				|| rollbackMutation.value.revision !== expectedRestoredRevision
				|| !canonicalDataEqual(current, rollbackMutation.value)) {
				throw new Error("rollback receipt and restored visibility diverge");
			}
			if (readyMutation !== undefined && readyMutation.revision >= rollbackMutation.revision) {
				throw new Error("causal mutation sequence places rollback before its ready receipt");
			}
		}
		if (state.phase !== "ready_settled" && settlement?.outcome === "ready") {
			throw new Error("unsettled authority cannot have a ready journal settlement");
		}
		if (state.phase === "ready_invoking") {
			if (!current.draft || current.revision !== state.originalRevision
				|| readyMutation !== undefined || rollbackMutation !== undefined || recoveryAttempt !== undefined) {
				throw new Error("invoking authority has incoherent receipt or draft visibility");
			}
			continue;
		}
		if (state.phase === "ready_effect_applied" || state.phase === "ready_settled") {
			if (readyMutation === undefined || rollbackMutation !== undefined || recoveryAttempt !== undefined
				|| !canonicalDataEqual(current, readyMutation.value)
				|| (state.phase === "ready_settled" && settlement !== undefined && settlement.outcome !== "ready")) {
				throw new Error("applied unsettled authority has incoherent mutation visibility");
			}
			if (state.phase === "ready_settled" && settlement === undefined) {
				settlementRepairs.push({
					schemaVersion: 1,
					planDigest: operation.planDigest,
					authorizationDigest: operation.authorization.digest,
					mutationIdempotencyKey: operation.mutation.idempotencyKey,
					outcome: "ready",
					settledAt: operation.decision.consumedAt!,
				});
			}
			continue;
		}
		if (recoveryAttempt === undefined || recoveryAttempt.attempt !== state.fence) {
			throw new Error("recovery authority is missing its exact recovery fence attempt");
		}
		if (rollbackMutation === undefined) {
			if (state.phase === "draft_restored") {
				throw new Error("restored recovery authority is missing its rollback receipt");
			}
			if (state.appliedRevision === null) {
				if (readyMutation !== undefined || !current.draft || current.revision !== state.originalRevision) {
					throw new Error("recovery claim-before-effect has incoherent draft visibility or receipt");
				}
			} else if (readyMutation === undefined || !canonicalDataEqual(current, readyMutation.value)) {
				throw new Error("applied recovery claim-before-effect has incoherent visibility or receipt");
			}
			continue;
		}
		if (state.phase === "draft_restored" && settlement !== undefined && settlement.outcome !== "blocked") {
			throw new Error("restored recovery authority is missing its blocked settlement");
		}
		if (state.phase === "draft_restored" && settlement === undefined) {
			settlementRepairs.push({
				schemaVersion: 1,
				planDigest: operation.planDigest,
				authorizationDigest: operation.authorization.digest,
				mutationIdempotencyKey: operation.mutation.idempotencyKey,
				outcome: "blocked",
				settledAt: operation.decision.consumedAt!,
			});
		}
	}
	settlementRepairs.sort((left, right) =>
		`${left.planDigest}\u0000${left.authorizationDigest}\u0000${left.mutationIdempotencyKey}`.localeCompare(
			`${right.planDigest}\u0000${right.authorizationDigest}\u0000${right.mutationIdempotencyKey}`,
		));
	return { settlementRepairs };
}

/** Validate causal consistency across decoded restart-store roles before any role is reconstructed. */
export function validateParentReadyRestartHistory(
	value: unknown,
	planValue: ParentOrchestrationPlan,
): ParentReadyRestartValidationResult {
	try {
		return assertParentReadyRestartHistory(value, planValue);
	} catch (error) {
		const detail = error instanceof Error ? error.message : "unknown inconsistency";
		throw new Error(`invalid parent-ready restart history: ${detail}`);
	}
}

export function validateRunStateCandidateSemantics(value: unknown): void {
	const state = exactRecord(value, [
		"schemaVersion", "phase", "issue", "parentIssue", "parentPr", "status", "reason", "details", "updatedAt",
	]);
	const details = exactRecord(state.details, [
		"branch", "baseBranch", "exactBase", "candidateRef", "priorCycleImplementationHead", "subPr", "subPrUrl",
		"adapter", "adapterHealth", "modelPolicy", "verificationPassed", "reviewCoveragePassed", "priorReviewState",
		"reviewHistory", "checkpoints", "verification", "orchestrationDecisions", "humanGates",
	]);
	if (details.candidateRef !== "HEAD") {
		throw new Error("current run-state candidate must use the single HEAD semantic; historical SHAs belong only in review history");
	}
}

export interface GitHubParentOrchestratorOptions {
	externalCallTimeoutMs?: number;
	shutdownTimeoutMs?: number;
	parentReadyAuthority: ParentReadyDurableAuthorityBoundary;
	now?: () => Date;
}

export class ExternalPortError extends Error {
	readonly code: "external_timeout" | "external_cancelled" | "external_port_failed";
	readonly operation: string;
	readonly uncertain: boolean;

	constructor(code: ExternalPortError["code"], operation: string, uncertain: boolean) {
		super(code === "external_timeout"
			? `external operation timed out: ${operation}`
			: code === "external_cancelled"
				? `external operation cancelled: ${operation}`
				: `external operation failed: ${operation}`);
		this.name = "ExternalPortError";
		this.code = code;
		this.operation = operation;
		this.uncertain = uncertain;
	}
}

export class GitHubParentOrchestrator {
	readonly #transport: GitHubOrchestrationTransport;
	readonly #broker?: ParentDecisionBroker;
	readonly #attestations?: AgentSessionAttestationSource;
	readonly #policySource?: RequiredCheckPolicySource;
	readonly #parentReadyAuthority: ParentReadyDurableAuthorityBoundary;
	readonly #now: () => Date;
	readonly #externalCallTimeoutMs: number;
	readonly #shutdownTimeoutMs: number;
	readonly #ensureLocks = new Map<string, Promise<void>>();
	readonly #parentReadyRecoveries = new Map<string, Promise<void>>();
	readonly #lifecycleScope = new AsyncLocalStorage<OrchestrationCallContext>();
	readonly #ensureCallScope = new AsyncLocalStorage<Set<Promise<void>>>();
	readonly #activeCalls = new Map<symbol, {
		operation: string;
		controller: AbortController;
		settlement: Promise<void>;
		abortAcknowledged: boolean;
	}>();
	readonly #stopController = new AbortController();
	#stopping = false;

	constructor(
		transport: GitHubOrchestrationTransport,
		broker: ParentDecisionBroker | undefined,
		attestations: AgentSessionAttestationSource | undefined,
		policySource: RequiredCheckPolicySource | undefined,
		options: GitHubParentOrchestratorOptions,
	) {
		if (options?.parentReadyAuthority === undefined
			|| typeof options.parentReadyAuthority.readParentReadyState !== "function"
			|| typeof options.parentReadyAuthority.compareConsumeAndMarkParentReady !== "function"
			|| typeof options.parentReadyAuthority.settleParentReady !== "function"
			|| typeof options.parentReadyAuthority.quarantineAndRollbackParentReady !== "function") {
			throw new Error("parent-ready durable authority boundary is required");
		}
		this.#transport = transport;
		this.#broker = broker;
		this.#attestations = attestations;
		this.#policySource = policySource;
		this.#parentReadyAuthority = options.parentReadyAuthority;
		this.#now = options.now ?? (() => new Date());
		if (!Number.isFinite(this.#now().valueOf())) throw new Error("parent orchestrator clock is invalid");
		const timeout = options.externalCallTimeoutMs ?? 15_000;
		if (!Number.isSafeInteger(timeout) || timeout < 1 || timeout > 120_000) {
			throw new Error("external call timeout must be a bounded positive integer");
		}
		this.#externalCallTimeoutMs = timeout;
		const shutdownTimeout = options.shutdownTimeoutMs ?? timeout;
		if (!Number.isSafeInteger(shutdownTimeout) || shutdownTimeout < 1 || shutdownTimeout > 120_000) {
			throw new Error("shutdown timeout must be a bounded positive integer");
		}
		this.#shutdownTimeoutMs = shutdownTimeout;
	}

	private lifecycleContext(value: OrchestrationCallContext | undefined): OrchestrationCallContext {
		const candidate = exactRecord(value ?? {}, [], ["signal", "deadlineAt"]);
		const signal = candidate.signal === undefined ? undefined : canonicalAbortSignal(candidate.signal);
		return {
			...(signal === undefined ? {} : { signal }),
			...(candidate.deadlineAt === undefined
				? {}
				: { deadlineAt: timestamp(candidate.deadlineAt, "orchestration caller deadline") }),
		};
	}

	private withLifecycle<T>(context: OrchestrationCallContext | undefined, operation: () => Promise<T>): Promise<T> {
		if (this.#stopping) return Promise.reject(new ExternalPortError("external_cancelled", "orchestrator", false));
		const canonical = this.lifecycleContext(context);
		if (canonical.signal !== undefined && intrinsicSignalAborted(canonical.signal)) {
			return Promise.reject(new ExternalPortError("external_cancelled", "orchestrator", false));
		}
		if (canonical.deadlineAt !== undefined && new Date(canonical.deadlineAt).valueOf() <= Date.now()) {
			return Promise.reject(new ExternalPortError("external_timeout", "orchestrator", false));
		}
		return this.#lifecycleScope.run(canonical, operation);
	}

	private async waitForLifecycle(value: Promise<void>): Promise<void> {
		const caller = this.#lifecycleScope.getStore();
		if ((caller?.signal !== undefined && intrinsicSignalAborted(caller.signal))
			|| intrinsicSignalAborted(this.#stopController.signal)) {
			throw new ExternalPortError("external_cancelled", "keyed orchestration wait", false);
		}
		const deadline = caller?.deadlineAt === undefined
			? Date.now() + this.#externalCallTimeoutMs
			: new Date(caller.deadlineAt).valueOf();
		if (deadline <= Date.now()) throw new ExternalPortError("external_timeout", "keyed orchestration wait", false);
		let timer: ReturnType<typeof setTimeout> | undefined;
		let dispose = (): void => {};
		const interrupted = new Promise<never>((_resolve, reject) => {
			let settled = false;
			const finish = (error: ExternalPortError): void => {
				if (settled) return;
				settled = true;
				reject(error);
			};
			const callerAbort = (): void => finish(new ExternalPortError("external_cancelled", "keyed orchestration wait", false));
			const stopAbort = (): void => finish(new ExternalPortError("external_cancelled", "keyed orchestration wait", false));
			const disposeCaller = caller?.signal === undefined ? (): void => {} : leaseAbortSignal(caller.signal, callerAbort);
			const disposeStop = leaseAbortSignal(this.#stopController.signal, stopAbort);
			timer = setTimeout(() => finish(new ExternalPortError("external_timeout", "keyed orchestration wait", false)), Math.max(0, deadline - Date.now()));
			dispose = () => {
				if (timer !== undefined) clearTimeout(timer);
				disposeCaller();
				disposeStop();
			};
		});
		try {
			await Promise.race([value, interrupted]);
		} finally {
			dispose();
		}
	}

	private async callExternal<T>(
		operation: string,
		invoke: (context: ExternalCallContext) => Promise<T>,
		uncertain = false,
		onUncertainInterruption?: () => Promise<void>,
		recoverAfterInvocationSettlement = false,
	): Promise<T> {
		const caller = this.#lifecycleScope.getStore();
		if (intrinsicSignalAborted(this.#stopController.signal)
			|| (caller?.signal !== undefined && intrinsicSignalAborted(caller.signal))) {
			throw new ExternalPortError("external_cancelled", operation, uncertain);
		}
		const callerDeadline = caller?.deadlineAt === undefined
			? Number.POSITIVE_INFINITY
			: new Date(caller.deadlineAt).valueOf();
		const deadline = Math.min(Date.now() + this.#externalCallTimeoutMs, callerDeadline);
		if (deadline <= Date.now()) throw new ExternalPortError("external_timeout", operation, uncertain);
		const controller = new AbortController();
		const token = Symbol(operation);
		let resolveFinalSettlement = (): void => {};
		const finalSettlement = new Promise<void>((resolve) => { resolveFinalSettlement = resolve; });
		const active = {
			operation,
			controller,
			settlement: finalSettlement,
			abortAcknowledged: false,
		};
		const context: ExternalCallContext = {
			signal: controller.signal,
			deadlineAt: new Date(deadline).toISOString(),
			acknowledgeAbort: () => {
				if (intrinsicSignalAborted(controller.signal)) active.abortAcknowledged = true;
			},
		};
		const invocation = Promise.resolve().then(() => invoke(context));
		const invocationSettlement = invocation.then(() => {}, () => {});
		let recovery: Promise<void> | undefined;
		const startRecovery = (): Promise<void> => {
			if (recovery === undefined) {
				const recover = onUncertainInterruption === undefined
					? (): Promise<void> => Promise.resolve()
					: onUncertainInterruption;
				recovery = (recoverAfterInvocationSettlement ? invocationSettlement : Promise.resolve())
					.then(recover)
					.catch(() => new Promise<void>(() => {}));
			}
			return recovery;
		};
		let finalizationStarted = false;
		const finalize = (includeRecovery: boolean): void => {
			if (finalizationStarted) return;
			finalizationStarted = true;
			const waits = includeRecovery ? [invocationSettlement, startRecovery()] : [invocationSettlement];
			void Promise.all(waits).then(resolveFinalSettlement);
		};
		this.#activeCalls.set(token, active);
		const ensureCalls = this.#ensureCallScope.getStore();
		ensureCalls?.add(finalSettlement);
		void finalSettlement.finally(() => {
			this.#activeCalls.delete(token);
			ensureCalls?.delete(finalSettlement);
		});
		const guarded = invocation.then(
			(value) => ({ kind: "value" as const, value }),
			() => ({ kind: "failed" as const }),
		);
		let timer: ReturnType<typeof setTimeout> | undefined;
		let dispose = (): void => {};
		const interrupted = new Promise<{ kind: "timeout" | "cancelled" }>((resolve) => {
			let settled = false;
			const finish = (kind: "timeout" | "cancelled"): void => {
				if (settled) return;
				settled = true;
				controller.abort();
				if (uncertain) startRecovery();
				resolve({ kind });
			};
			const callerAbort = (): void => finish("cancelled");
			const stopAbort = (): void => finish("cancelled");
			const disposeCaller = caller?.signal === undefined ? (): void => {} : leaseAbortSignal(caller.signal, callerAbort);
			const disposeStop = leaseAbortSignal(this.#stopController.signal, stopAbort);
			timer = setTimeout(() => finish("timeout"), Math.max(0, deadline - Date.now()));
			dispose = () => {
				if (timer !== undefined) clearTimeout(timer);
				disposeCaller();
				disposeStop();
			};
		});
		const outcome = await Promise.race([guarded, interrupted]);
		dispose();
		if (outcome.kind === "value") {
			finalize(false);
			return outcome.value;
		}
		if (outcome.kind === "failed") {
			finalize(uncertain);
			throw new ExternalPortError("external_port_failed", operation, uncertain);
		}
		finalize(uncertain);
		if (outcome.kind === "timeout") throw new ExternalPortError("external_timeout", operation, true);
		if (outcome.kind === "cancelled") throw new ExternalPortError("external_cancelled", operation, uncertain);
		throw new ExternalPortError("external_port_failed", operation, uncertain);
	}

	private async serializeEnsure<T>(key: string, operation: () => Promise<T>): Promise<T> {
		const previous = this.#ensureLocks.get(key) ?? Promise.resolve();
		let release = (): void => {};
		const gate = new Promise<void>((resolve) => { release = resolve; });
		const tail = previous.catch(() => {}).then(() => gate);
		this.#ensureLocks.set(key, tail);
		void tail.finally(() => {
			if (this.#ensureLocks.get(key) === tail) this.#ensureLocks.delete(key);
		});
		try {
			await this.waitForLifecycle(previous.catch(() => {}));
		} catch (error) {
			release();
			throw error;
		}
		const externalSettlements = new Set<Promise<void>>();
		try {
			return await this.#ensureCallScope.run(externalSettlements, operation);
		} finally {
			if (externalSettlements.size === 0) release();
			else void Promise.all([...externalSettlements]).then(release, release);
		}
	}

	private trackBackgroundSettlement(operation: string, settlement: Promise<void>): void {
		const token = Symbol(operation);
		const controller = new AbortController();
		const finalSettlement = settlement.then(() => {}, () => new Promise<void>(() => {}));
		this.#activeCalls.set(token, {
			operation,
			controller,
			settlement: finalSettlement,
			abortAcknowledged: false,
		});
		const ensureCalls = this.#ensureCallScope.getStore();
		ensureCalls?.add(finalSettlement);
		void finalSettlement.finally(() => {
			this.#activeCalls.delete(token);
			ensureCalls?.delete(finalSettlement);
		});
	}

	async stop(context?: OrchestrationCallContext): Promise<{ kind: "joined" | "incomplete"; active: number; unacknowledged: number }> {
		const caller = this.lifecycleContext(context);
		this.#stopping = true;
		this.#stopController.abort();
		const deadline = Math.min(
			Date.now() + this.#shutdownTimeoutMs,
			caller.deadlineAt === undefined ? Number.POSITIVE_INFINITY : new Date(caller.deadlineAt).valueOf(),
		);
		const settlements = [...this.#activeCalls.values()].map((call) => call.settlement);
		if (settlements.length > 0 && deadline > Date.now()
			&& !(caller.signal !== undefined && intrinsicSignalAborted(caller.signal))) {
			let timer: ReturnType<typeof setTimeout> | undefined;
			let disposeCaller = (): void => {};
			const interrupted = new Promise<void>((resolve) => {
				timer = setTimeout(resolve, Math.max(0, deadline - Date.now()));
				if (caller.signal !== undefined) disposeCaller = leaseAbortSignal(caller.signal, resolve);
			});
			await Promise.race([Promise.all(settlements).then(() => {}), interrupted]);
			if (timer !== undefined) clearTimeout(timer);
			disposeCaller();
		}
		const active = this.#activeCalls.size;
		const unacknowledged = [...this.#activeCalls.values()].filter((call) => !call.abortAcknowledged).length;
		return { kind: active === 0 ? "joined" : "incomplete", active, unacknowledged };
	}

	async createPlan(value: unknown, context?: OrchestrationCallContext): Promise<ParentOrchestrationPlan> {
		return this.withLifecycle(context, async () => {
			if (this.#policySource?.findParentOrchestrationPolicyBundle === undefined) {
				throw new Error("authoritative parent orchestration policy source is unavailable");
			}
			const query = objectiveAuthorityQuery(value);
			const records = authoritativeItems(
				await this.callExternal("findParentOrchestrationPolicyBundle", (callContext) =>
					this.#policySource!.findParentOrchestrationPolicyBundle!(query, callContext)),
				"parent orchestration policy authority lookup",
			);
			if (records.length !== 1) throw new Error("parent orchestration policy authority is absent or ambiguous");
			const authority = validateParentPolicyAuthority(records[0], query);
			return createParentOrchestrationPlan(value, authority.policyBundle);
		});
	}

	private async reconcileVisible<T>(read: () => Promise<T | null>): Promise<T | null> {
		for (let attempt = 0; attempt < 3; attempt += 1) {
			const value = await read();
			if (value !== null) return value;
		}
		return null;
	}

	private async matchingIssues(plan: ParentOrchestrationPlan, child: BoundedChildRecord): Promise<GitHubChildIssue[]> {
		const raw = await this.callExternal("findChildIssues", (context) => this.#transport.findChildIssues(
			{ repository: plan.repository, marker: child.markers.issue }, context,
		));
		return authoritativeItems(raw, "child issue lookup").map(validateGitHubChildIssue);
	}

	private resolveIssueMatches(
		matches: GitHubChildIssue[],
		plan: ParentOrchestrationPlan,
		child: BoundedChildRecord,
	): GitHubChildIssue | null {
		if (matches.length > 1) throw new Error("duplicate child issue marker is ambiguous");
		if (matches.length === 0) return null;
		assertChildIssueMatches(matches[0], plan, child);
		return matches[0];
	}

	async ensureChildIssue(
		plan: ParentOrchestrationPlan,
		childId: string,
		context?: OrchestrationCallContext,
	): Promise<GitHubChildIssue> {
		return this.withLifecycle(context, async () => {
		plan = validateParentOrchestrationPlan(plan);
		const child = childFor(plan, childId);
		return this.serializeEnsure(`${plan.repository}:issue:${child.markers.issue}`, async () => {
			const existing = this.resolveIssueMatches(await this.matchingIssues(plan, child), plan, child);
			if (existing !== null) return existing;
			const requestIntent = {
				repository: plan.repository,
				parentIssue: plan.parentIssue,
				marker: child.markers.issue,
				title: child.title,
				body: child.issueBody,
			};
			const mutation = createDurableMutationIntent("child_issue", [plan.repository, child.markers.issue], requestIntent, null);
			const request: CreateChildIssueRequest = { ...requestIntent, mutation };
			let mutationError: unknown;
			try {
				validateDurableMutationResult(
					await this.callExternal("createChildIssue", (context) => this.#transport.createChildIssue(request, context), true),
					mutation,
					validateGitHubChildIssue,
				);
			} catch (error) {
				mutationError = error;
			}
			const recovered = await this.reconcileVisible(async () => this.resolveIssueMatches(
				await this.matchingIssues(plan, child), plan, child,
			));
			if (recovered !== null) return recovered;
			if (mutationError !== undefined) throw mutationError;
			throw new Error("child issue create did not produce one authoritative canonical resource");
		});
		});
	}

	private async matchingPullRequests(query: PullRequestMarkerQuery): Promise<GitHubPullRequestEvidence[]> {
		const raw = await this.callExternal("findPullRequests", (context) => this.#transport.findPullRequests(query, context));
		return authoritativeItems(raw, "pull request lookup").map(validateGitHubPullRequestEvidence);
	}

	private async singlePullRequest(query: PullRequestMarkerQuery): Promise<GitHubPullRequestEvidence | null> {
		const matches = await this.matchingPullRequests(query);
		if (matches.length > 1) throw new Error("duplicate pull request marker is ambiguous");
		return matches[0] ?? null;
	}

	private assertPublishedPullRequest(
		value: GitHubPullRequestEvidence,
		request: CreatePullRequestRequest,
	): GitHubPullRequestEvidence {
		if (value.marker !== request.marker || value.title !== request.title || value.body !== request.body
			|| value.state !== "open" || value.draft !== request.draft
			|| value.repository !== request.repository || value.workItemId !== request.workItemId
			|| value.generation !== request.generation || value.policyDigest !== request.policyDigest
			|| value.baseBranch !== request.baseBranch || value.headBranch !== request.headBranch
			|| value.baseSha !== request.baseSha || value.headSha !== request.headSha
			|| !sameStrings(value.changedPaths, [...request.changedPaths].sort())
			|| !sameStrings(value.allowedScopes, [...request.allowedScopes].sort())) {
			throw new Error("pull request marker collision or canonical resource mismatch");
		}
		return value;
	}

	private async ensurePullRequest(requestValue: Omit<CreatePullRequestRequest, "mutation">): Promise<GitHubPullRequestEvidence> {
		const mutation = createDurableMutationIntent(
			"pull_request",
			[requestValue.repository, requestValue.marker],
			requestValue,
			null,
		);
		const request: CreatePullRequestRequest = { ...requestValue, mutation };
		return this.serializeEnsure(`${request.repository}:pr:${request.marker}`, async () => {
			const query = { repository: request.repository, marker: request.marker };
			const existing = await this.singlePullRequest(query);
			if (existing !== null) return this.assertPublishedPullRequest(existing, request);
			let mutationError: unknown;
			try {
				validateDurableMutationResult(
					await this.callExternal("createPullRequest", (context) => this.#transport.createPullRequest(request, context), true),
					request.mutation,
					validateGitHubPullRequestEvidence,
				);
			} catch (error) {
				mutationError = error;
			}
			const recovered = await this.reconcileVisible(() => this.singlePullRequest(query));
			if (recovered !== null) return this.assertPublishedPullRequest(recovered, request);
			if (mutationError !== undefined) throw mutationError;
			throw new Error("pull request create did not produce one authoritative canonical resource");
		});
	}

	private async canonicalMaterializedChild(
		plan: ParentOrchestrationPlan,
		value: MaterializedChildRecord,
	): Promise<MaterializedChildRecord> {
		const child = validateMaterializedChild(plan, value);
		const planned = childFor(plan, child.id);
		const issue = this.resolveIssueMatches(await this.matchingIssues(plan, planned), plan, planned);
		if (issue === null || issue.number !== child.issue) {
			throw new Error("materialized child issue is not the authoritative canonical issue");
		}
		return child;
	}

	async captureChildHandoff(
		plan: ParentOrchestrationPlan,
		child: MaterializedChildRecord,
		workspace: ClaimedWorkspace,
		source: WorkspaceHandoffSource,
		context?: OrchestrationCallContext,
	): Promise<WorkspaceHandoffEvidence> {
		return this.withLifecycle(context, async () => {
		plan = validateParentOrchestrationPlan(plan);
		const canonicalChild = await this.canonicalMaterializedChild(plan, child);
		const handoff = await this.callExternal("captureChildHandoff", (context) => source.captureHandoff(workspace, "passed", context));
		return validateHandoff(handoff, canonicalChild.issue, canonicalChild.branch, canonicalChild.prBase, canonicalChild.writeScopes);
		});
	}

	async captureParentHandoff(
		plan: ParentOrchestrationPlan,
		workspace: ClaimedWorkspace,
		source: WorkspaceHandoffSource,
		context?: OrchestrationCallContext,
	): Promise<WorkspaceHandoffEvidence> {
		return this.withLifecycle(context, async () => {
		plan = validateParentOrchestrationPlan(plan);
		const handoff = await this.callExternal("captureParentHandoff", (context) => source.captureHandoff(workspace, "passed", context));
		return validateHandoff(
			handoff,
			plan.parentIssue,
			plan.parentBranch,
			plan.parentBaseBranch,
			aggregateScopes(plan),
		);
		});
	}

	async ensureChildPullRequest(
		plan: ParentOrchestrationPlan,
		child: MaterializedChildRecord,
		handoffValue: WorkspaceHandoffEvidence,
		context?: OrchestrationCallContext,
	): Promise<GitHubPullRequestEvidence> {
		return this.withLifecycle(context, async () => {
		plan = validateParentOrchestrationPlan(plan);
		const canonicalChild = await this.canonicalMaterializedChild(plan, child);
		const handoff = validateHandoff(handoffValue, canonicalChild.issue, canonicalChild.branch, canonicalChild.prBase, canonicalChild.writeScopes);
		return this.ensurePullRequest({
			repository: plan.repository,
			workItemId: canonicalChild.id,
			generation: plan.generation,
			marker: canonicalChild.markers.pullRequest,
			title: canonicalChild.title,
			body: childPullRequestBody(plan, canonicalChild),
			draft: false,
			baseBranch: canonicalChild.prBase,
			headBranch: canonicalChild.branch,
			baseSha: handoff.baseHead,
			headSha: handoff.head,
			changedPaths: handoff.changedScope,
			allowedScopes: canonicalChild.writeScopes,
			policyDigest: policyFor(plan, canonicalChild.prBase).digest,
		});
		});
	}

	async ensureParentDraftPullRequest(
		plan: ParentOrchestrationPlan,
		handoffValue: WorkspaceHandoffEvidence,
		context?: OrchestrationCallContext,
	): Promise<GitHubPullRequestEvidence> {
		return this.withLifecycle(context, async () => {
		plan = validateParentOrchestrationPlan(plan);
		const handoff = validateHandoff(
			handoffValue,
			plan.parentIssue,
			plan.parentBranch,
			plan.parentBaseBranch,
			aggregateScopes(plan),
		);
		return this.ensurePullRequest({
			repository: plan.repository,
			workItemId: `parent-${plan.parentIssue}`,
			generation: plan.generation,
			marker: plan.markers.parentPullRequest,
			title: plan.title,
			body: parentPullRequestBody(plan),
			draft: true,
			baseBranch: plan.parentBaseBranch,
			headBranch: plan.parentBranch,
			baseSha: handoff.baseHead,
			headSha: handoff.head,
			changedPaths: handoff.changedScope,
			allowedScopes: aggregateScopes(plan),
			policyDigest: policyFor(plan, plan.parentBaseBranch).digest,
		});
		});
	}

	private async rosterMatches(plan: ParentOrchestrationPlan): Promise<GitHubRosterSnapshot[]> {
		const raw = await this.callExternal("findParentRosters", (context) => this.#transport.findParentRosters(
			{ repository: plan.repository, marker: plan.markers.roster }, context,
		));
		return authoritativeItems(raw, "roster lookup").map(validateGitHubRosterSnapshot);
	}

	private resolveRoster(matches: GitHubRosterSnapshot[], plan: ParentOrchestrationPlan): GitHubRosterSnapshot | null {
		if (matches.length > 1) throw new Error("duplicate parent roster marker is ambiguous");
		const existing = matches[0];
		if (existing === undefined) return null;
		if (existing.marker !== plan.markers.roster || existing.parentIssue !== plan.parentIssue
			|| existing.generation !== plan.generation || existing.body.split(plan.markers.roster).length !== 2) {
			throw new Error("parent roster marker collision or canonical resource mismatch");
		}
		return existing;
	}

	async reconcileParentRoster(
		plan: ParentOrchestrationPlan,
		statusValue: Readonly<Record<string, WorkItemStatus>>,
		statusEpoch: number,
		context?: OrchestrationCallContext,
	): Promise<GitHubRosterSnapshot> {
		return this.withLifecycle(context, async () => {
		plan = validateParentOrchestrationPlan(plan);
		const statuses = statusesForPlan(plan, statusValue);
		const epoch = generation(statusEpoch);
		return this.serializeEnsure(`${plan.repository}:roster:${plan.markers.roster}`, async () => {
			const body = renderRoster(plan, statuses);
			const existing = this.resolveRoster(await this.rosterMatches(plan), plan);
			if (existing !== null) {
				const canonicalExistingStatuses = statusesForPlan(plan, existing.statuses);
				if (existing.statusEpoch > epoch) throw new Error("stale roster status epoch");
				if (existing.statusEpoch === epoch) {
					if (existing.body !== body || !canonicalDataEqual(canonicalExistingStatuses, statuses)) {
						throw new Error("roster status epoch conflicts with existing state");
					}
					return existing;
				}
				assertMonotonicStatuses(canonicalExistingStatuses, statuses);
			}
			const requestIntent = {
				repository: plan.repository,
				marker: plan.markers.roster,
				parentIssue: plan.parentIssue,
				generation: plan.generation,
				body,
				statuses,
				statusEpoch: epoch,
			};
			const mutation = createDurableMutationIntent(
				"parent_roster", [plan.repository, plan.markers.roster, epoch], requestIntent, existing?.revision ?? null,
			);
			const request: PublishRosterRequest = { ...requestIntent, mutation };
			let mutationError: unknown;
			let publishedRevision: number | undefined;
			try {
				const result = validateDurableMutationResult(
					await this.callExternal("publishParentRoster", (context) => this.#transport.publishParentRoster(request, context), true),
					mutation,
					validateGitHubRosterSnapshot,
				);
				publishedRevision = result.value.revision;
				if (existing !== null && publishedRevision <= existing.revision) {
					throw new Error("parent roster CAS revision did not advance");
				}
			} catch (error) {
				mutationError = error;
			}
			const recovered = await this.reconcileVisible(async () => {
				const roster = this.resolveRoster(await this.rosterMatches(plan), plan);
				return roster?.body === body && roster.statusEpoch === epoch
					&& (existing === null || roster.revision > existing.revision)
					&& (publishedRevision === undefined || roster.revision === publishedRevision)
					&& canonicalDataEqual(statusesForPlan(plan, roster.statuses), statuses) ? roster : null;
			});
			if (recovered !== null) return recovered;
			if (mutationError !== undefined) throw mutationError;
			throw new Error("parent roster mutation was not durably visible");
		});
		});
	}

	private async evaluateEvidence(
		evidence: GitHubPullRequestEvidence,
		expected: {
			repository: string;
			workItemId: string;
			generation: number;
			number: number;
			marker: string;
			baseBranch: string;
			headBranch: string;
			baseSha: string;
			headSha: string;
		},
		target: Omit<IndependentReviewTarget, "changedPaths">,
		requiredCheckPolicy: RequiredGitHubCheckPolicy,
		allowDraft = false,
		allowIntegrated = false,
		allowIntegratedBaseMovement = false,
	) {
		if (this.#attestations === undefined) throw new Error("controller evidence source is unavailable");
		const query = {
			repository: target.repository,
			workItemId: target.workItemId,
			pullRequest: target.pullRequest,
			generation: target.generation,
			baseBranch: target.baseBranch,
			headBranch: target.headBranch,
			baseSha: target.baseSha,
			headSha: target.headSha,
		};
		const observations = authoritativeItems(
			await this.callExternal("findChangedPathEvidence", (context) => this.#attestations!.findChangedPathEvidence(query, context)),
			"controller changed-path evidence lookup",
		).map(validateGitHubChangedPathEvidence);
		if (observations.length !== 1) throw new Error("controller changed-path evidence is absent or ambiguous");
		const observation = observations[0];
		const canonicalTarget: IndependentReviewTarget = { ...target, changedPaths: observation.paths };
		const attestations = authoritativeItems(
			await this.callExternal("findAttestations", (context) => this.#attestations!.findAttestations(canonicalTarget, context)),
			"AgentSession attestation lookup",
		).map((entry) => validateAgentSessionAttestation(entry));
		const decision = evaluateGitHubPullRequestEvidence(evidence, {
			...expected,
			changedPathEvidence: observation,
			minimumObservation: { revision: observation.revision, observedAt: observation.observedAt },
			requiredCheckPolicy,
			reviewTarget: canonicalTarget,
			attestations,
			}, { allowDraft, allowIntegrated, allowIntegratedBaseMovement });
		return { decision, observation, target: canonicalTarget, attestations };
	}

	private async currentPolicySet(
		plan: ParentOrchestrationPlan,
		minimumObservedAt: string,
	): Promise<Map<string, RequiredGitHubCheckPolicyObservation> | null> {
		if (this.#policySource === undefined) return null;
		try {
			const entries = await Promise.all(plan.requiredCheckPolicies.map(async (expected) => {
				const raw = await this.callExternal("findRequiredCheckPolicies", (context) => this.#policySource!.findRequiredCheckPolicies(
					{ repository: expected.repository, baseBranch: expected.baseBranch }, context,
				));
				const observations = authoritativeItems(raw, "required-check policy observation lookup")
					.map(validateRequiredGitHubCheckPolicyObservation);
				if (observations.length !== 1) throw new Error("required-check policy observation is absent or ambiguous");
				const observation = observations[0];
				if (observation.repository !== expected.repository
					|| observation.baseBranch !== expected.baseBranch
					|| observation.revision !== expected.revision
					|| observation.digest !== expected.digest
					|| observation.observedAt < minimumObservedAt) {
					throw new Error("required-check policy moved or is stale");
				}
				return [expected.baseBranch, observation] as const;
			}));
			return new Map(entries);
		} catch {
			return null;
		}
	}

	private async currentPolicyObservation(
		plan: ParentOrchestrationPlan,
		baseBranch: string,
		minimumObservedAt: string,
	): Promise<RequiredGitHubCheckPolicyObservation | null> {
		return (await this.currentPolicySet(plan, minimumObservedAt))?.get(baseBranch) ?? null;
	}

	private async currentChildAuthorization(
		plan: ParentOrchestrationPlan,
		child: MaterializedChildRecord,
		pullRequest: GitHubPullRequestEvidence,
		handoff: ChildPullRequestTopology,
		integratedState?: PullRequestObservation["state"],
		receiptProvenance?: ControllerIntegrationProvenance,
	): Promise<ControllerIntegrationProvenance | null> {
		if (!currentChildPullRequestEligible(pullRequest, integratedState)
			|| !childPullRequestMatches(pullRequest, plan, child, handoff, integratedState === "merged")
			|| !sameStrings(pullRequest.changedPaths, handoff.changedScope)) return null;
		const policies = await this.currentPolicySet(plan, pullRequest.observedAt);
		const policy = policyFor(plan, child.prBase);
		const policyObservation = policies?.get(child.prBase);
		if (policyObservation === undefined) return null;
		const expected = {
			repository: plan.repository,
			workItemId: child.id,
			generation: plan.generation,
			number: pullRequest.number,
			marker: child.markers.pullRequest,
			baseBranch: child.prBase,
			headBranch: child.branch,
			baseSha: handoff.baseHead,
			headSha: handoff.head,
		};
		const reviewTarget: Omit<IndependentReviewTarget, "changedPaths"> = {
			repository: plan.repository,
			workItemId: child.id,
			pullRequest: pullRequest.number,
			generation: plan.generation,
			baseBranch: child.prBase,
			headBranch: child.branch,
			baseSha: handoff.baseHead,
			headSha: handoff.head,
			allowedScopes: child.writeScopes,
		};
		const assessed = await this.evaluateEvidence(
			pullRequest,
			expected,
			reviewTarget,
			policy,
			false,
			integratedState !== undefined,
			integratedState === "merged",
		);
		if (assessed.decision.kind === "blocked"
			|| !sameStrings(assessed.observation.paths, handoff.changedScope)
			|| !reviewMatchesChild(assessed.decision.review, plan, child, pullRequest.number, handoff)) return null;
		const review = assessed.decision.review;
		if (receiptProvenance !== undefined && !assessed.attestations.some((attestation) =>
			attestation.resultDigest === receiptProvenance.reviewResultDigest
			&& attestation.completedAt === receiptProvenance.reviewCompletedAt
			&& attestation.reviewMarker === review.idempotencyMarker
		)) return null;
		return controllerIntegrationProvenance(plan, policy, policyObservation, assessed.observation, review);
	}

	async integrateChild(
		plan: ParentOrchestrationPlan,
		child: MaterializedChildRecord,
		handoffValue: WorkspaceHandoffEvidence,
		mutate: ChildIntegrationMutation,
		context?: OrchestrationCallContext,
	): Promise<ChildIntegrationDecision> {
		if (typeof mutate !== "function") throw new Error("child integration requires lease-bound Git mutation authority");
		return this.withLifecycle(context, async () => {
		plan = validateParentOrchestrationPlan(plan);
		const canonicalChild = await this.canonicalMaterializedChild(plan, child);
		return this.serializeEnsure(`${plan.repository}:integration:${canonicalChild.markers.pullRequest}`, async () => {
			let handoff: WorkspaceHandoffEvidence;
			try {
				handoff = validateHandoff(handoffValue, canonicalChild.issue, canonicalChild.branch, canonicalChild.prBase, canonicalChild.writeScopes);
			} catch {
				return { kind: "blocked", blockers: ["handoff_invalid"] };
			}
			const query = { repository: plan.repository, marker: canonicalChild.markers.pullRequest };
			const first = await this.singlePullRequest(query);
			if (first === null) return { kind: "blocked", blockers: ["pull_request_missing"] };
			const recoveringIntegrated = first.state === "merged";
			if (!childPullRequestMatches(first, plan, canonicalChild, handoff, recoveringIntegrated)
				|| !sameStrings(first.changedPaths, handoff.changedScope)) {
				return { kind: "blocked", blockers: ["resource_mismatch"] };
			}
			const expected = {
				repository: plan.repository,
				workItemId: canonicalChild.id,
				generation: plan.generation,
				number: first.number,
				marker: canonicalChild.markers.pullRequest,
				baseBranch: canonicalChild.prBase,
				headBranch: canonicalChild.branch,
				baseSha: handoff.baseHead,
				headSha: handoff.head,
			};
			const integrationQuery = { repository: plan.repository, childId: canonicalChild.id, marker: canonicalChild.markers.pullRequest };
			const existingItems = authoritativeItems(
				await this.callExternal("findChildIntegration", (context) => this.#transport.findChildIntegration(integrationQuery, context)),
				"child integration lookup",
			);
			if (existingItems.length > 1) throw new Error("duplicate child integration receipt is ambiguous");
			if (existingItems.length === 1) {
				const receipt = validateChildIntegrationReceipt(existingItems[0]);
				if (!receiptMatchesChild(receipt, plan, canonicalChild, first.number, handoff, first)) {
					throw new Error("existing child integration receipt is stale or mismatched");
				}
				if (first.headSha !== handoff.head) return { kind: "blocked", blockers: ["head_moved"] };
				if (!currentChildPullRequestEligible(first, receipt.observation.state)) {
					return { kind: "blocked", blockers: [first.draft ? "draft" : "pr_not_open"] };
				}
				const currentProvenance = await this.currentChildAuthorization(
					plan, canonicalChild, first, handoff, receipt.observation.state, receipt.controllerProvenance,
				);
				if (currentProvenance === null || !controllerAuthorizationMatches(
					currentProvenance, receipt.controllerProvenance, first, receipt.pullRequestSnapshot,
				)) {
					return { kind: "blocked", blockers: ["policy_moved"] };
				}
				return { kind: "integrated", receipt, reused: true };
			}
			if (!currentChildPullRequestEligible(first, recoveringIntegrated ? "merged" : undefined)) {
				return { kind: "blocked", blockers: [first.draft ? "draft" : "pr_not_open"] };
			}
			const reviewTarget: Omit<IndependentReviewTarget, "changedPaths"> = {
				repository: plan.repository,
				workItemId: canonicalChild.id,
				pullRequest: first.number,
				generation: plan.generation,
				baseBranch: canonicalChild.prBase,
				headBranch: canonicalChild.branch,
				baseSha: handoff.baseHead,
				headSha: handoff.head,
				allowedScopes: canonicalChild.writeScopes,
			};
			const policy = policyFor(plan, canonicalChild.prBase);
			const assessed = await this.evaluateEvidence(
				first, expected, reviewTarget, policy, false, recoveringIntegrated, recoveringIntegrated,
			);
			if (assessed.decision.kind === "blocked") return assessed.decision;
			if (!sameStrings(assessed.observation.paths, handoff.changedScope)
				|| !reviewMatchesChild(assessed.decision.review, plan, canonicalChild, first.number, handoff)) {
				return { kind: "blocked", blockers: ["review_missing"] };
			}
			const second = await this.singlePullRequest(query);
			if (second === null) return { kind: "blocked", blockers: ["pull_request_missing"] };
			if (second.headSha !== handoff.head) return { kind: "blocked", blockers: ["head_moved"] };
			const recoveringSecond = recoveringIntegrated || second.state === "merged";
			if (!childPullRequestMatches(second, plan, canonicalChild, handoff, recoveringSecond)
				|| !sameStrings(second.changedPaths, handoff.changedScope)) {
				return { kind: "blocked", blockers: ["resource_mismatch"] };
			}
			const revalidated = await this.evaluateEvidence(
				second,
				expected,
				reviewTarget,
				policy,
				false,
				recoveringSecond,
				recoveringSecond,
			);
			if (revalidated.decision.kind === "blocked") return revalidated.decision;
			if (!sameStrings(revalidated.observation.paths, handoff.changedScope)
				|| !reviewMatchesChild(revalidated.decision.review, plan, canonicalChild, second.number, handoff)) {
				return { kind: "blocked", blockers: ["review_missing"] };
			}
			const policyObservation = await this.currentPolicyObservation(plan, canonicalChild.prBase, second.observedAt);
			if (policyObservation === null) {
				return { kind: "blocked", blockers: ["policy_moved"] };
			}
			const snapshot = childIntegrationPullRequestSnapshot(second, handoff);
			const observation = pullRequestObservation(second);
			const controllerProvenance = controllerIntegrationProvenance(
				plan, policy, policyObservation, revalidated.observation, revalidated.decision.review,
			);
			const requestIntent = {
				repository: plan.repository,
				childId: canonicalChild.id,
				pullRequest: second.number,
				generation: plan.generation,
				marker: canonicalChild.markers.pullRequest,
				baseSha: handoff.baseHead,
				headSha: handoff.head,
				parentBranch: plan.parentBranch,
				pullRequestSnapshot: snapshot,
				observation,
				controllerProvenance,
			};
			const mutation = createDurableMutationIntent(
				"child_integration",
				[plan.repository, canonicalChild.markers.pullRequest],
				childIntegrationMutationProjection(requestIntent),
				null,
			);
			const integration = validateReviewedChildIntegrationEvidence(await mutate({
				repository: plan.repository,
				childId: canonicalChild.id,
				pullRequest: second.number,
				generation: plan.generation,
				marker: canonicalChild.markers.pullRequest,
				parentBranch: plan.parentBranch,
				parentBaseBranch: plan.parentBaseBranch,
				baseSha: handoff.baseHead,
				headSha: handoff.head,
				idempotencyKey: mutation.idempotencyKey,
				intentDigest: mutation.intentDigest,
			}));
			if (integration.parentBranch !== plan.parentBranch
				|| integration.baseSha !== handoff.baseHead || integration.headSha !== handoff.head
				|| integration.mergeCommitSha === integration.baseSha
				|| integration.mergeCommitSha === integration.headSha) {
				throw new Error("lease-bound Git integration evidence does not match the exact reviewed child");
			}
			const request: IntegrateChildRequest = {
				...requestIntent,
				parentBaseBranch: plan.parentBaseBranch,
				integration,
				mutation,
			};
			let mutationError: unknown;
			let mutationApplied: boolean | undefined;
			try {
				const result = validateDurableMutationResult(
					await this.callExternal("integrateChild", (context) => this.#transport.integrateChild(request, context), true),
					mutation,
					validateChildIntegrationReceipt,
				);
				mutationApplied = result.applied;
				if (!canonicalDataEqual(result.value.transportProvenance, transportProvenance(result))) {
					throw new Error("integration receipt transport provenance mismatch");
				}
			} catch (error) {
				mutationError = error;
			}
			const recovered = await this.reconcileVisible(async () => {
				const items = authoritativeItems(
					await this.callExternal("findChildIntegration", (context) => this.#transport.findChildIntegration(integrationQuery, context)),
					"child integration recovery lookup",
				);
				if (items.length > 1) throw new Error("duplicate child integration receipt is ambiguous");
				if (items.length === 0) return null;
				const receipt = validateChildIntegrationReceipt(items[0]);
				const current = await this.singlePullRequest(query);
				if (current === null) return null;
				const currentProvenance = await this.currentChildAuthorization(
					plan, canonicalChild, current, handoff, receipt.observation.state, receipt.controllerProvenance,
				);
				return currentProvenance !== null
					&& controllerAuthorizationMatches(
						currentProvenance, receipt.controllerProvenance, current, receipt.pullRequestSnapshot,
					)
					&& receiptMatchesChild(receipt, plan, canonicalChild, request.pullRequest, handoff, current)
					? receipt : null;
			});
			if (recovered !== null) return {
				kind: "integrated",
				receipt: recovered,
				reused: mutationError !== undefined || mutationApplied === false,
			};
			if (mutationError !== undefined) throw mutationError;
			throw new Error("child integration mutation was not durably visible");
		});
		});
	}

	private async completeIntegrationRoster(
		plan: ParentOrchestrationPlan,
		values: readonly ChildIntegrationReceipt[],
		parentPullRequest: GitHubPullRequestEvidence,
	): Promise<{ receipts: ChildIntegrationReceipt[]; ancestryProofs: GitAncestryProof[] } | null> {
		let receipts: ChildIntegrationReceipt[];
		try {
			const snapshot = boundedArray(values, "child integration roster", MAX_CHILDREN, true);
			if (snapshot.length !== plan.children.length) return null;
			receipts = snapshot.map(validateChildIntegrationReceipt);
		} catch {
			return null;
		}
		unique(receipts, (receipt) => receipt.childId, "integrated child ID");
		unique(receipts, (receipt) => String(receipt.pullRequest), "integrated pull request");
		const planned = new Map(plan.children.map((child) => [child.id, child]));
		if (receipts.some((receipt) => {
			const child = planned.get(receipt.childId);
			return child === undefined
				|| receipt.generation !== plan.generation
				|| receipt.marker !== child.markers.pullRequest
				|| receipt.parentBranch !== plan.parentBranch
				|| receipt.controllerProvenance.planDigest !== plan.canonical.digest
				|| receipt.controllerProvenance.policyDigest !== policyFor(plan, child.prBase).digest
				|| receipt.pullRequestSnapshot.repository !== plan.repository
				|| receipt.pullRequestSnapshot.workItemId !== child.id
				|| receipt.pullRequestSnapshot.number !== receipt.pullRequest
				|| receipt.pullRequestSnapshot.generation !== plan.generation
				|| receipt.pullRequestSnapshot.marker !== child.markers.pullRequest
				|| receipt.pullRequestSnapshot.baseBranch !== plan.parentBranch
				|| receipt.pullRequestSnapshot.baseSha !== receipt.baseSha
				|| receipt.pullRequestSnapshot.headSha !== receipt.headSha
				|| receipt.pullRequestSnapshot.policyDigest !== policyFor(plan, child.prBase).digest
				|| !sameStrings(receipt.pullRequestSnapshot.allowedScopes, child.writeScopes)
				|| receipt.observation.revision !== receipt.pullRequestSnapshot.revision
				|| receipt.observation.observedAt !== receipt.pullRequestSnapshot.observedAt;
		})) return null;
		const callerByChild = new Map(receipts.map((receipt) => [receipt.childId, receipt]));
		const authoritative: ChildIntegrationReceipt[] = [];
		const ancestryProofs: GitAncestryProof[] = [];
		for (const child of plan.children) {
			let materialized: MaterializedChildRecord;
			try {
				const issues = await this.matchingIssues(plan, child);
				if (issues.length !== 1) return null;
				materialized = materializeChildRecord(plan, child.id, issues[0]);
			} catch {
				return null;
			}
			let items: unknown[];
			try {
				const integrationQuery = {
					repository: plan.repository,
					childId: child.id,
					marker: child.markers.pullRequest,
				};
				items = authoritativeItems(
					await this.callExternal("findChildIntegration", (context) => this.#transport.findChildIntegration(integrationQuery, context)),
					"child integration roster lookup",
				);
			} catch {
				return null;
			}
			if (items.length !== 1) return null;
			let receipt: ChildIntegrationReceipt;
			try {
				receipt = validateChildIntegrationReceipt(items[0]);
			} catch {
				return null;
			}
			const supplied = callerByChild.get(child.id);
			if (supplied === undefined || !canonicalDataEqual(receipt, supplied)
				|| receipt.childId !== child.id || receipt.marker !== child.markers.pullRequest
				|| receipt.generation !== plan.generation || receipt.parentBranch !== plan.parentBranch) return null;
			let pullRequests: GitHubPullRequestEvidence[];
			try {
				pullRequests = await this.matchingPullRequests({ repository: plan.repository, marker: child.markers.pullRequest });
			} catch {
				return null;
			}
			if (pullRequests.length !== 1) return null;
			const pullRequest = pullRequests[0];
			const handoff: WorkspaceHandoffEvidence = {
				issue: materialized.issue,
				branch: materialized.branch,
				prBase: materialized.prBase,
				baseHead: receipt.baseSha,
				head: receipt.headSha,
				changedScope: [...receipt.pullRequestSnapshot.changedPaths],
				verificationState: "passed",
				repositoryIdentity: "0".repeat(64),
				worktreeIdentity: "0".repeat(64),
				dirty: false,
			};
			if (!currentChildPullRequestEligible(pullRequest, receipt.observation.state)
				|| !childPullRequestMatches(pullRequest, plan, materialized, handoff)
				|| !receiptMatchesChild(receipt, plan, materialized, pullRequest.number, handoff, pullRequest)) return null;
			let currentProvenance: ControllerIntegrationProvenance | null;
			try {
				currentProvenance = await this.currentChildAuthorization(
					plan, materialized, pullRequest, handoff, receipt.observation.state, receipt.controllerProvenance,
				);
			} catch {
				return null;
			}
			if (currentProvenance === null || !controllerAuthorizationMatches(
				currentProvenance, receipt.controllerProvenance, pullRequest, receipt.pullRequestSnapshot,
			)) return null;
			const expectedMutation = createDurableMutationIntent(
				"child_integration",
				[plan.repository, child.markers.pullRequest],
				childIntegrationMutationProjection({
					repository: plan.repository,
					childId: child.id,
					pullRequest: pullRequest.number,
					generation: plan.generation,
					marker: child.markers.pullRequest,
					baseSha: receipt.pullRequestSnapshot.baseSha,
					headSha: receipt.pullRequestSnapshot.headSha,
					parentBranch: plan.parentBranch,
					pullRequestSnapshot: receipt.pullRequestSnapshot,
					controllerProvenance: receipt.controllerProvenance,
				}),
				null,
			);
			if (receipt.transportProvenance.idempotencyKey !== expectedMutation.idempotencyKey
				|| receipt.transportProvenance.intentDigest !== expectedMutation.intentDigest) return null;
			const ancestryQuery = {
				repository: plan.repository,
				ancestorSha: receipt.headSha,
				descendantSha: parentPullRequest.headSha,
			};
			try {
				const ancestryProof = validateAncestryProof(
					await this.callExternal("proveAncestry", (context) => this.#transport.proveAncestry(ancestryQuery, context)),
					ancestryQuery,
					parentPullRequest.observedAt,
				);
				ancestryProofs.push(ancestryProof);
			} catch {
				return null;
			}
			authoritative.push(receipt);
		}
		return { receipts: authoritative, ancestryProofs };
	}

	private humanDecisionLifecycle(plan: ParentOrchestrationPlan, decision: "pending" | "approve_merge" | "reject") {
		return reconcileAutonomy({
			persisted: { stage: "HUMAN_DECISION", retryAttempts: 0, correctionRounds: 0 },
			canonical: {
				observedStage: "HUMAN_DECISION",
				proposedStage: decision === "reject" ? "ABORTED" : "MERGE",
				transitionFacts: {
					humanDecision: decision,
					humanDecisionAuthenticated: decision !== "pending",
					exactHeadRevalidated: decision === "approve_merge",
				},
				workItems: plan.children.map((child) => ({
					id: child.id,
					dependsOn: [...child.dependsOn],
					status: "succeeded" as const,
					access: child.access,
					writeScopes: [...child.writeScopes],
				})),
				maxConcurrency: 1,
				constraints: {
					runtimeCapabilityAvailable: true,
					isolationAvailable: true,
					hardHumanGate: decision === "pending",
					verificationBlocked: false,
					reviewBlocked: false,
				},
			},
			budget: { maxRetries: 0, maxCorrectionRounds: 0 },
		});
	}

	private async rollbackParentReadyEffect(
		plan: ParentOrchestrationPlan,
		ready: GitHubPullRequestEvidence,
		authorization: ParentReadyAuthorization,
	): Promise<void> {
		if (ready.repository !== plan.repository
			|| ready.number !== authorization.pullRequest
			|| ready.marker !== plan.markers.parentPullRequest
			|| ready.headSha !== authorization.headSha) {
			throw new Error("parent ready rollback target does not match the original authorization");
		}
		const mutation = createDurableMutationIntent(
			"parent_ready",
			[authorization.repository, plan.markers.parentPullRequest, authorization.headSha],
			{
				repository: authorization.repository,
				pullRequest: authorization.pullRequest,
				headSha: authorization.headSha,
				generation: authorization.generation,
				decisionRequestId: authorization.decisionRequestId,
				authorization,
			},
			authorization.pullRequestRevision,
		);
		await this.quarantineUncertainParentReady(
			{ authorization, mutation },
			this.parentReadyAuthorityQuery(plan, authorization),
		);
	}

	private async prepareParentReadinessUnlocked(
		plan: ParentOrchestrationPlan,
		integrationValues: readonly ChildIntegrationReceipt[],
		policyValue: ParentDecisionPolicy,
	): Promise<ParentReadinessPreparation> {
		plan = validateParentOrchestrationPlan(plan);
		const query = { repository: plan.repository, marker: plan.markers.parentPullRequest };
		const first = await this.singlePullRequest(query);
		if (first === null) return { kind: "blocked", blockers: ["parent_pull_request_missing"] };
		if (!parentPullRequestMatches(first, plan)) {
			return { kind: "blocked", blockers: ["parent_pull_request_collision"] };
		}
		const firstAuthorityState = await this.readParentReadyAuthorityState(plan, {
			pullRequest: first.number,
			headSha: first.headSha,
		});
		const firstAuthorityBlock = this.parentReadyRestartBlock(plan, first, firstAuthorityState);
		if (firstAuthorityBlock !== null) return firstAuthorityBlock;
		if (await this.completeIntegrationRoster(plan, integrationValues, first) === null) {
			return { kind: "blocked", blockers: ["children_incomplete"] };
		}
		const expected = {
			repository: plan.repository,
			workItemId: `parent-${plan.parentIssue}`,
			generation: plan.generation,
			number: first.number,
			marker: plan.markers.parentPullRequest,
			baseBranch: plan.parentBaseBranch,
			headBranch: plan.parentBranch,
			baseSha: first.baseSha,
			headSha: first.headSha,
		};
		const reviewTarget: Omit<IndependentReviewTarget, "changedPaths"> = {
			repository: plan.repository,
			workItemId: `parent-${plan.parentIssue}`,
			pullRequest: first.number,
			generation: plan.generation,
			baseBranch: plan.parentBaseBranch,
			headBranch: plan.parentBranch,
			baseSha: first.baseSha,
			headSha: first.headSha,
			allowedScopes: aggregateScopes(plan),
		};
		const requiredCheckPolicy = policyFor(plan, plan.parentBaseBranch);
		const assessed = await this.evaluateEvidence(first, expected, reviewTarget, requiredCheckPolicy, true);
		if (assessed.decision.kind === "blocked") return assessed.decision;
		if (!sameStrings(assessed.observation.paths, first.changedPaths)
			|| !reviewMatchesParent(assessed.decision.review, plan, first)) {
			return { kind: "blocked", blockers: ["review_missing"] };
		}
		if (await this.currentPolicyObservation(plan, plan.parentBaseBranch, first.observedAt) === null) {
			return { kind: "blocked", blockers: ["policy_moved"] };
		}
		if (this.#broker === undefined) return { kind: "awaiting_human", reason: "broker_unavailable" };
		const policy = validateDecisionPolicy(policyValue);
		const request: GitHubDecisionRequest = {
			requestId: policy.requestId,
			gate: "parent_merge",
			repository: plan.repository,
			parentIssue: plan.parentIssue,
			pullRequest: first.number,
			generation: plan.generation,
			headSha: first.headSha,
			allowedOptions: ["approve-merge", "reject"],
			actorAllowlist: [...policy.actorAllowlist],
			expiresAt: policy.expiresAt,
			question: policy.question,
		};
		const binding: HumanDecisionBinding = {
			repository: plan.repository,
			target: { kind: "pull_request", number: first.number },
			generation: plan.generation,
			headSha: first.headSha,
		};
		let record = await this.callExternal("broker.request", async (context) => validateBrokerRecord(
			await this.#broker!.request(request, context), request, binding, undefined, this.#now(),
		), true);
		if (record.status === "expired") return { kind: "awaiting_human", reason: "expired" };
		if (record.status === "pending") {
			const polled = await this.callExternal("broker.poll", async (context) => validateBrokerRecord(
				await this.#broker!.poll(policy.requestId, binding, { signal: context.signal }, context),
				request,
				binding,
				record,
				this.#now(),
			));
			if (polled.status === "pending") return { kind: "awaiting_human", reason: "pending" };
			if (polled.status === "expired") return { kind: "awaiting_human", reason: "expired" };
			record = polled;
		}
		if (record.status !== "consumed") {
			const consumed = await this.callExternal("broker.consume", async (context) => validateBrokerRecord(
				await this.#broker!.consume(policy.requestId, binding, context),
				request,
				binding,
				record,
				this.#now(),
			), true);
			if (consumed.status !== "consumed" || !canonicalDataEqual(consumed.decision, record.decision)) {
				throw new Error("human decision broker consume did not preserve the exact decided evidence");
			}
			record = consumed;
		}
		const decision = record.decision;
		if (decision === undefined) throw new Error("human decision broker record has no authenticated decision evidence");
		if (decision.option === "reject") {
			const lifecycle = this.humanDecisionLifecycle(plan, "reject");
			if (lifecycle.kind !== "transition" || lifecycle.to !== "ABORTED") {
				return { kind: "blocked", blockers: ["autonomy_policy_rejected_decision_state"] };
			}
			return { kind: "rejected" };
		}
		if (decision.option !== "approve-merge") {
			return { kind: "blocked", blockers: ["invalid_parent_decision"] };
		}
		const second = await this.singlePullRequest(query);
		if (second === null) return { kind: "blocked", blockers: ["parent_pull_request_missing"] };
		if (!parentPullRequestMatches(second, plan)) {
			return { kind: "blocked", blockers: ["parent_pull_request_collision"] };
		}
		const authorityState = await this.readParentReadyAuthorityState(plan, {
			pullRequest: second.number,
			headSha: second.headSha,
		});
		const secondAuthorityBlock = this.parentReadyRestartBlock(plan, second, authorityState);
		if (secondAuthorityBlock !== null) return secondAuthorityBlock;
		const currentRoster = await this.completeIntegrationRoster(plan, integrationValues, second);
		if (currentRoster === null) {
			return { kind: "blocked", blockers: ["children_incomplete"] };
		}
		const revalidated = await this.evaluateEvidence(second, expected, reviewTarget, requiredCheckPolicy, true);
		if (revalidated.decision.kind === "blocked") return revalidated.decision;
		if (!sameStrings(revalidated.observation.paths, second.changedPaths)
			|| !reviewMatchesParent(revalidated.decision.review, plan, second)) {
			return { kind: "blocked", blockers: ["review_missing"] };
		}
		const currentPolicies = await this.currentPolicySet(plan, second.observedAt);
		if (currentPolicies === null || currentPolicies.get(plan.parentBaseBranch) === undefined) {
			return { kind: "blocked", blockers: ["policy_moved"] };
		}
		const lifecycle = this.humanDecisionLifecycle(plan, "approve_merge");
		if (lifecycle.kind !== "transition" || lifecycle.to !== "MERGE") {
			return { kind: "blocked", blockers: ["autonomy_policy_blocked_parent_ready"] };
		}
		if (authorityState?.phase === "ready_settled") {
			const semanticPullRequest = { ...second, revision: authorityState.originalRevision };
			const currentAuthority = createParentReadyAuthority(
				plan,
				semanticPullRequest,
				record,
				revalidated.decision.review,
				revalidated.observation,
				currentPolicies,
				currentRoster,
			).authorization;
			if (currentAuthority.digest !== authorityState.authorization.digest
				|| !canonicalDataEqual(
					expectedParentReadyMutation(currentAuthority, plan.markers.parentPullRequest),
					authorityState.readyMutation,
				)) return { kind: "blocked", blockers: ["parent_ready_authority_moved"] };
			return { kind: "ready", pullRequest: second, reused: true };
		}
		const { authorization, freshness } = createParentReadyAuthority(
			plan,
			second,
			record,
			revalidated.decision.review,
			revalidated.observation,
			currentPolicies,
			currentRoster,
		);
		const markIntent = {
			repository: plan.repository,
			pullRequest: second.number,
			headSha: second.headSha,
			generation: plan.generation,
			decisionRequestId: policy.requestId,
			authorization,
		};
		const mutation = createDurableMutationIntent(
			"parent_ready", [plan.repository, plan.markers.parentPullRequest, second.headSha], markIntent, second.revision,
		);
		return {
			kind: "prepared",
			operation: {
				schemaVersion: 1,
				planDigest: plan.canonical.digest,
				policy,
				decision: record,
				authorization,
				freshness,
				mutation,
			},
		};
	}

	private parentReadyAuthorityQuery(
		plan: ParentOrchestrationPlan,
		target: Pick<ParentReadyAuthorization, "pullRequest" | "headSha">,
	): ParentReadyAuthorityQuery {
		return validateParentReadyAuthorityQuery({
			repository: plan.repository,
			pullRequest: target.pullRequest,
			marker: plan.markers.parentPullRequest,
			headSha: target.headSha,
			generation: plan.generation,
		});
	}

	private async readParentReadyAuthorityState(
		plan: ParentOrchestrationPlan,
		target: Pick<ParentReadyAuthorization, "pullRequest" | "headSha">,
	): Promise<ParentReadyAuthorityState | null> {
		const query = this.parentReadyAuthorityQuery(plan, target);
		return this.callExternal("readParentReadyState", async (context) => {
			const state = await this.#parentReadyAuthority.readParentReadyState(query, context);
			if (state === null) return null;
			const validated = validateParentReadyAuthorityState(state);
			if (!canonicalDataEqual(query, {
				repository: validated.repository,
				pullRequest: validated.pullRequest,
				marker: validated.marker,
				headSha: validated.headSha,
				generation: validated.generation,
			})) throw new Error("parent-ready authority state does not match its query");
			return validated;
		});
	}

	private parentReadyOperationFromState(
		state: ParentReadyAuthorityState,
	): Pick<PreparedParentReadyOperation, "authorization" | "mutation"> {
		return { authorization: state.authorization, mutation: state.readyMutation };
	}

	private parentReadyAuthorityQueryFromState(state: ParentReadyAuthorityState): ParentReadyAuthorityQuery {
		return validateParentReadyAuthorityQuery({
			repository: state.repository,
			pullRequest: state.pullRequest,
			marker: state.marker,
			headSha: state.headSha,
			generation: state.generation,
		});
	}

	private parentReadyRestartBlock(
		plan: ParentOrchestrationPlan,
		pullRequest: GitHubPullRequestEvidence,
		state: ParentReadyAuthorityState | null,
	): Extract<ParentReadinessDecision, { kind: "blocked" }> | null {
		if (state !== null && ["ready_invoking", "ready_effect_applied", "recovery_claimed"].includes(state.phase)) {
			this.startTrackedParentReadyRecovery(
				plan,
				this.parentReadyOperationFromState(state),
				this.parentReadyAuthorityQueryFromState(state),
			);
			return { kind: "blocked", blockers: ["parent_ready_quarantined"] };
		}
		if (state?.phase === "ready_settled"
			&& (pullRequest.draft || state.appliedRevision !== pullRequest.revision)) {
			return { kind: "blocked", blockers: ["parent_ready_authority_inconsistent"] };
		}
		if (state?.phase === "draft_restored" && !pullRequest.draft) {
			return { kind: "blocked", blockers: ["parent_ready_authority_inconsistent"] };
		}
		return state === null && !pullRequest.draft
			? { kind: "blocked", blockers: ["parent_ready_authority_missing"] }
			: null;
	}

	private parentReadyRecovery(
		plan: ParentOrchestrationPlan,
		operation: Pick<PreparedParentReadyOperation, "authorization" | "mutation">,
		queryValue?: ParentReadyAuthorityQuery,
	): Promise<void> {
		const query = queryValue ?? this.parentReadyAuthorityQuery(plan, operation.authorization);
		const invocationId = parentReadyInvocationId(query, operation.authorization.digest, operation.mutation);
		const existing = this.#parentReadyRecoveries.get(invocationId);
		if (existing !== undefined) return existing;
		const settlement = this.quarantineUncertainParentReady(operation, query).finally(() => {
			if (this.#parentReadyRecoveries.get(invocationId) === settlement) {
				this.#parentReadyRecoveries.delete(invocationId);
			}
		});
		this.#parentReadyRecoveries.set(invocationId, settlement);
		return settlement;
	}

	private startTrackedParentReadyRecovery(
		plan: ParentOrchestrationPlan,
		operation: Pick<PreparedParentReadyOperation, "authorization" | "mutation">,
		query?: ParentReadyAuthorityQuery,
	): void {
		this.trackBackgroundSettlement("parent-ready recovery", this.parentReadyRecovery(plan, operation, query));
	}

	private trackParentReadyProof<T>(description: string, proof: Promise<T>): Promise<T> {
		this.trackBackgroundSettlement(description, proof.then(() => {}));
		return proof;
	}

	private async readExactParentReadyProof(query: ParentReadyAuthorityQuery): Promise<ParentReadyAuthorityState | null> {
		for (let attempt = 1; ; attempt += 1) {
			const outcome = await this.boundedParentReadyRecoveryCall((context) =>
				this.#parentReadyAuthority.readParentReadyState(query, context));
			if (outcome.kind === "value") {
				try {
					if (outcome.value === null) return null;
					const state = validateParentReadyAuthorityState(outcome.value);
					if (!canonicalDataEqual(query, this.parentReadyAuthorityQueryFromState(state))) {
						throw new Error("returned-coordinate authority reread moved to another coordinate");
					}
					return state;
				} catch {
					// A malformed or cross-coordinate reread is not terminal proof; retain ownership and retry.
				}
			}
			await new Promise<void>((resolve) => setTimeout(resolve, Math.min(10 * attempt, 25)));
		}
	}

	private async reconcileMismatchedParentReadyCoordinate(
		plan: ParentOrchestrationPlan,
		returned: ParentReadyAuthorityState,
	): Promise<"moved" | "quarantined"> {
		const query = this.parentReadyAuthorityQueryFromState(returned);
		const observed = await this.readExactParentReadyProof(query);
		if (observed === null || observed.phase === "ready_settled" || observed.phase === "draft_restored") {
			return "moved";
		}
		await this.parentReadyRecovery(plan, this.parentReadyOperationFromState(observed), query);
		return "quarantined";
	}

	private parentReadyRollbackRequest(
		query: ParentReadyAuthorityQuery,
		operation: Pick<PreparedParentReadyOperation, "authorization" | "mutation">,
		attempt: number,
	): RollbackParentReadyRequest {
		const intent: Omit<RollbackParentReadyRequest, "recovery" | "mutation"> = {
			repository: query.repository,
			pullRequest: operation.authorization.pullRequest,
			marker: query.marker,
			headSha: operation.authorization.headSha,
			generation: query.generation,
			authorizationDigest: operation.authorization.digest,
		};
		const invocationId = parentReadyInvocationId(query, operation.authorization.digest, operation.mutation);
		const recovery: ParentReadyRecoveryFence = {
			schemaVersion: 1,
			invocationId,
			recoveryId: parentReadyRecoveryId(query, invocationId, operation.authorization.digest, operation.mutation),
			attempt,
			supersedesAttempt: attempt === 1 ? null : attempt - 1,
			readyMutation: operation.mutation,
		};
		const mutation = createDurableMutationIntent(
			"parent_ready_rollback",
			[recovery.recoveryId, recovery.attempt],
			{ ...intent, recovery },
			null,
		);
		return validateRollbackParentReadyRequest({ ...intent, recovery, mutation });
	}

	private async boundedParentReadyRecoveryCall<T>(
		invoke: (context: ExternalCallContext) => Promise<T>,
	): Promise<
		| { kind: "value"; value: T }
		| { kind: "failed" }
		| { kind: "timeout"; settlement: Promise<void> }
	> {
		const controller = new AbortController();
		const deadline = Date.now() + this.#externalCallTimeoutMs;
		const context: ExternalCallContext = {
			signal: controller.signal,
			deadlineAt: new Date(deadline).toISOString(),
			acknowledgeAbort: () => {},
		};
		const invocation = Promise.resolve().then(() => invoke(context));
		const settlement = invocation.then(() => {}, () => {});
		const guarded = invocation.then(
			(value) => ({ kind: "value" as const, value }),
			() => ({ kind: "failed" as const }),
		);
		let timer: ReturnType<typeof setTimeout> | undefined;
		const timedOut = new Promise<{ kind: "timeout"; settlement: Promise<void> }>((resolve) => {
			timer = setTimeout(() => {
				controller.abort();
				resolve({ kind: "timeout", settlement });
			}, Math.max(0, deadline - Date.now()));
		});
		const outcome = await Promise.race([guarded, timedOut]);
		if (timer !== undefined) clearTimeout(timer);
		return outcome;
	}

	private async quarantineUncertainParentReady(
		operation: Pick<PreparedParentReadyOperation, "authorization" | "mutation">,
		query: ParentReadyAuthorityQuery,
	): Promise<void> {
		const abandonedSettlements = new Set<Promise<void>>();
		const invocationId = parentReadyInvocationId(query, operation.authorization.digest, operation.mutation);
		const recoveryId = parentReadyRecoveryId(query, invocationId, operation.authorization.digest, operation.mutation);
		let minimumReleaseFence = 0;
		for (let attempt = 1; ; attempt += 1) {
			const read = await this.boundedParentReadyRecoveryCall((context) =>
				this.#parentReadyAuthority.readParentReadyState(query, context));
			if (read.kind === "timeout") abandonedSettlements.add(read.settlement);
			if (read.kind !== "value") {
				await new Promise<void>((resolve) => setTimeout(resolve, Math.min(10 * attempt, 25)));
				continue;
			}
			let authorityState: ParentReadyAuthorityState | null;
			try {
				authorityState = read.value === null ? null : validateParentReadyAuthorityState(read.value);
			} catch {
				await new Promise<void>((resolve) => setTimeout(resolve, Math.min(10 * attempt, 25)));
				continue;
			}
			if (authorityState === null) {
				abandonedSettlements.clear();
				return;
			}
			if (authorityState.invocationId !== invocationId
				|| authorityState.recoveryId !== recoveryId
				|| authorityState.authorization.digest !== operation.authorization.digest
				|| !canonicalDataEqual(authorityState.readyMutation, operation.mutation)) {
				abandonedSettlements.clear();
				return;
			}
			if (authorityState.phase === "ready_settled") {
				abandonedSettlements.clear();
				return;
			}
			if (authorityState.phase === "draft_restored" && authorityState.fence >= minimumReleaseFence) {
				abandonedSettlements.clear();
				return;
			}
			if (!["ready_invoking", "ready_effect_applied", "recovery_claimed", "draft_restored"]
				.includes(authorityState.phase)) {
				abandonedSettlements.clear();
				return;
			}
			attempt = Math.max(attempt, authorityState.fence + 1, minimumReleaseFence);
			const request = this.parentReadyRollbackRequest(query, operation, attempt);
			const outcome = await this.boundedParentReadyRecoveryCall((context) =>
				this.#parentReadyAuthority.quarantineAndRollbackParentReady(request, context));
			if (outcome.kind === "timeout") abandonedSettlements.add(outcome.settlement);
			if (outcome.kind === "value") {
				try {
					const result = validateDurableMutationResult(
						outcome.value,
						request.mutation,
						validateGitHubPullRequestEvidence,
					);
					const minimumDraftRevision = authorityState.appliedRevision ?? operation.authorization.pullRequestRevision;
					if (!result.value.draft
						|| result.value.repository !== request.repository
						|| result.value.number !== request.pullRequest
						|| result.value.marker !== request.marker
						|| result.value.generation !== request.generation
						|| result.value.headSha !== request.headSha
						|| result.value.revision < minimumDraftRevision) {
						throw new Error("parent-ready quarantine did not restore the exact draft");
					}
					const confirmation = await this.boundedParentReadyRecoveryCall((context) =>
						this.#parentReadyAuthority.readParentReadyState(query, context));
					if (confirmation.kind === "timeout") abandonedSettlements.add(confirmation.settlement);
					if (confirmation.kind !== "value") throw new Error("parent-ready quarantine confirmation failed");
					const state = validateParentReadyAuthorityState(confirmation.value);
					if (state.phase !== "draft_restored" || state.status !== "settled"
						|| state.fence !== request.recovery.attempt
						|| state.invocationId !== request.recovery.invocationId
						|| state.recoveryId !== request.recovery.recoveryId
						|| state.authorization.digest !== operation.authorization.digest
						|| !canonicalDataEqual(state.readyMutation, operation.mutation)
						|| !canonicalDataEqual(state.rollbackMutation, request.mutation)) {
						throw new Error("parent-ready quarantine state is not durably settled");
					}
					abandonedSettlements.clear();
					return;
				} catch {
					// A malformed response cannot release the fence; a newer ordered attempt must supersede it.
					minimumReleaseFence = Math.max(minimumReleaseFence, request.recovery.attempt + 1);
				}
			}
			await new Promise<void>((resolve) => setTimeout(resolve, Math.min(10 * attempt, 25)));
		}
	}

	private async authorizedVisibleReady(
		plan: ParentOrchestrationPlan,
		integrationValues: readonly ChildIntegrationReceipt[],
		operation: PreparedParentReadyOperation,
		readyRevision?: number,
	): Promise<GitHubPullRequestEvidence | null> {
		try {
			const ready = await this.singlePullRequest({ repository: plan.repository, marker: plan.markers.parentPullRequest });
			if (ready === null || ready.draft || !parentPullRequestMatches(ready, plan)
				|| ready.headSha !== operation.authorization.headSha
				|| ready.revision <= operation.authorization.pullRequestRevision
				|| (readyRevision !== undefined && ready.revision !== readyRevision)) return null;
			const roster = await this.completeIntegrationRoster(plan, integrationValues, ready);
			if (roster === null) return null;
			const expected = {
				repository: plan.repository,
				workItemId: `parent-${plan.parentIssue}`,
				generation: plan.generation,
				number: ready.number,
				marker: plan.markers.parentPullRequest,
				baseBranch: plan.parentBaseBranch,
				headBranch: plan.parentBranch,
				baseSha: ready.baseSha,
				headSha: ready.headSha,
			};
			const reviewTarget: Omit<IndependentReviewTarget, "changedPaths"> = {
				repository: plan.repository,
				workItemId: `parent-${plan.parentIssue}`,
				pullRequest: ready.number,
				generation: plan.generation,
				baseBranch: plan.parentBaseBranch,
				headBranch: plan.parentBranch,
				baseSha: ready.baseSha,
				headSha: ready.headSha,
				allowedScopes: aggregateScopes(plan),
			};
			const requiredCheckPolicy = policyFor(plan, plan.parentBaseBranch);
			const assessed = await this.evaluateEvidence(ready, expected, reviewTarget, requiredCheckPolicy);
			if (assessed.decision.kind === "blocked"
				|| !sameStrings(assessed.observation.paths, ready.changedPaths)
				|| !reviewMatchesParent(assessed.decision.review, plan, ready)) return null;
			const policies = await this.currentPolicySet(plan, ready.observedAt);
			if (policies === null) return null;
			if (this.#broker === undefined) return null;
			const request: GitHubDecisionRequest = {
				requestId: operation.policy.requestId,
				gate: "parent_merge",
				repository: plan.repository,
				parentIssue: plan.parentIssue,
				pullRequest: ready.number,
				generation: plan.generation,
				headSha: ready.headSha,
				allowedOptions: ["approve-merge", "reject"],
				actorAllowlist: [...operation.policy.actorAllowlist],
				expiresAt: operation.policy.expiresAt,
				question: operation.policy.question,
			};
			const binding: HumanDecisionBinding = {
				repository: plan.repository,
				target: { kind: "pull_request", number: ready.number },
				generation: plan.generation,
				headSha: ready.headSha,
			};
			const record = await this.callExternal("broker.request", async (context) => validateBrokerRecord(
				await this.#broker!.request(request, context),
				request,
				binding,
				undefined,
				this.#now(),
			), true);
			if (record.status !== "consumed" || !canonicalDataEqual(record, operation.decision)) return null;
			const semanticPullRequest = { ...ready, revision: operation.authorization.pullRequestRevision };
			const current = createParentReadyAuthority(
				plan,
				semanticPullRequest,
				record,
				assessed.decision.review,
				assessed.observation,
				policies,
				roster,
			).authorization;
			return current.digest === operation.authorization.digest ? ready : null;
		} catch {
			return null;
		}
	}

	private async commitPreparedParentReadinessUnlocked(
		plan: ParentOrchestrationPlan,
		integrationValues: readonly ChildIntegrationReceipt[],
		operationValue: PreparedParentReadyOperation,
		revalidate: boolean,
	): Promise<ParentReadinessDecision> {
		const operation = validatePreparedParentReadyOperation(operationValue, plan);
		let freshness = operation.freshness;
		const sameOperation = (state: ParentReadyAuthorityState): boolean =>
			state.authorization.digest === operation.authorization.digest
			&& canonicalDataEqual(state.readyMutation, operation.mutation);
		const initialState = await this.readParentReadyAuthorityState(plan, operation.authorization);
		if (initialState !== null && ["ready_invoking", "ready_effect_applied", "recovery_claimed"].includes(initialState.phase)) {
			this.startTrackedParentReadyRecovery(
				plan,
				this.parentReadyOperationFromState(initialState),
				this.parentReadyAuthorityQueryFromState(initialState),
			);
			return { kind: "blocked", blockers: ["parent_ready_quarantined"] };
		}
		if (initialState?.phase === "ready_settled" && sameOperation(initialState)) {
			const reused = await this.reconcileVisible(() => this.authorizedVisibleReady(
				plan,
				integrationValues,
				operation,
				initialState.appliedRevision ?? undefined,
			));
			return reused === null
				? { kind: "blocked", blockers: ["parent_ready_authority_moved"] }
				: { kind: "ready", pullRequest: reused, reused: true };
		}
		if (revalidate) {
			const current = await this.prepareParentReadinessUnlocked(plan, integrationValues, operation.policy);
			if (current.kind !== "prepared") return current;
			if (current.operation.authorization.digest !== operation.authorization.digest
				|| current.operation.mutation.idempotencyKey !== operation.mutation.idempotencyKey
				|| current.operation.mutation.intentDigest !== operation.mutation.intentDigest) {
				return { kind: "blocked", blockers: ["parent_ready_authority_moved"] };
			}
			freshness = current.operation.freshness;
		}
		const request = validateMarkParentReadyRequest({
			repository: operation.authorization.repository,
			pullRequest: operation.authorization.pullRequest,
			marker: plan.markers.parentPullRequest,
			headSha: operation.authorization.headSha,
			generation: operation.authorization.generation,
			decisionRequestId: operation.authorization.decisionRequestId,
			authorization: operation.authorization,
			freshness,
			mutation: operation.mutation,
		});
		let begun: ParentReadyAuthorityState;
		try {
			begun = await this.callExternal(
				"beginParentReady",
				async (context) => validateParentReadyAuthorityState(
					await this.#parentReadyAuthority.beginParentReady(request, context),
				),
				true,
				() => this.parentReadyRecovery(plan, operation),
				true,
			);
		} catch (error) {
			if (error instanceof ExternalPortError) {
				return { kind: "blocked", blockers: ["parent_ready_quarantined"] };
			}
			throw error;
		}
		if (!sameOperation(begun)) {
			const requestedProof = this.trackParentReadyProof(
				"requested parent-ready reconciliation",
				this.parentReadyRecovery(plan, operation),
			);
			const observedProof = this.trackParentReadyProof(
				"returned-coordinate parent-ready reconciliation",
				this.reconcileMismatchedParentReadyCoordinate(plan, begun),
			);
			const [, observedOutcome] = await Promise.all([requestedProof, observedProof]);
			return observedOutcome === "quarantined"
				? { kind: "blocked", blockers: ["parent_ready_quarantined"] }
				: { kind: "blocked", blockers: ["parent_ready_authority_moved"] };
		}
		if (["ready_effect_applied", "recovery_claimed"].includes(begun.phase)) {
			this.startTrackedParentReadyRecovery(
				plan,
				this.parentReadyOperationFromState(begun),
				this.parentReadyAuthorityQueryFromState(begun),
			);
			return { kind: "blocked", blockers: ["parent_ready_quarantined"] };
		}
		if (begun.phase === "ready_settled") {
			const reused = await this.reconcileVisible(() => this.authorizedVisibleReady(
				plan,
				integrationValues,
				operation,
				begun.appliedRevision ?? undefined,
			));
			return reused === null
				? { kind: "blocked", blockers: ["parent_ready_authority_moved"] }
				: { kind: "ready", pullRequest: reused, reused: true };
		}
		if (begun.phase !== "ready_invoking" || begun.fence !== 0) {
			return { kind: "blocked", blockers: ["parent_ready_authority_moved"] };
		}
		let comparison: ParentReadyCompareEffectResult;
		try {
			comparison = await this.callExternal(
					"compareConsumeAndMarkParentReady",
					async (context) => validateParentReadyCompareEffectResult(
						await this.#parentReadyAuthority.compareConsumeAndMarkParentReady(request, context),
						request,
					),
					true,
					() => this.parentReadyRecovery(plan, operation),
				);
		} catch (error) {
			if (error instanceof ExternalPortError && error.uncertain) {
				return { kind: "blocked", blockers: ["parent_ready_quarantined"] };
			}
			throw error;
		}
		if (comparison.kind === "conflict") {
			return { kind: "blocked", blockers: [`parent_ready_authority_conflict:${comparison.coordinate}`] };
		}
		const readyRevision = comparison.mutation.value.revision;
		let authorityState: ParentReadyAuthorityState | null;
		try {
			authorityState = await this.readParentReadyAuthorityState(plan, operation.authorization);
		} catch {
			this.startTrackedParentReadyRecovery(plan, operation);
			return { kind: "blocked", blockers: ["parent_ready_quarantined"] };
		}
		if (authorityState === null || !sameOperation(authorityState)
			|| !["ready_effect_applied", "ready_settled"].includes(authorityState.phase)
			|| authorityState.appliedRevision !== readyRevision) {
			this.startTrackedParentReadyRecovery(plan, operation);
			return { kind: "blocked", blockers: ["parent_ready_quarantined"] };
		}
		const recovered = await this.reconcileVisible(() => this.authorizedVisibleReady(
			plan,
			integrationValues,
			operation,
			readyRevision,
		));
		if (recovered === null) {
			this.startTrackedParentReadyRecovery(plan, operation);
			return { kind: "blocked", blockers: ["parent_ready_quarantined"] };
		}
		if (authorityState.phase !== "ready_settled") {
			const query = this.parentReadyAuthorityQuery(plan, operation.authorization);
			const settlementRequest = validateSettleParentReadyAuthorityRequest({
				...query,
				invocationId: authorityState.invocationId,
				authorizationDigest: operation.authorization.digest,
				readyMutation: operation.mutation,
				expectedPhase: "ready_effect_applied",
				expectedFence: 0,
			});
			try {
				authorityState = await this.callExternal(
					"settleParentReady",
					async (context) => validateParentReadyAuthorityState(
						await this.#parentReadyAuthority.settleParentReady(settlementRequest, context),
					),
					true,
					() => this.parentReadyRecovery(plan, operation),
				);
			} catch (error) {
				if (!(error instanceof ExternalPortError) || !error.uncertain) throw error;
				try {
					const query = this.parentReadyAuthorityQuery(plan, operation.authorization);
					const reread = await this.boundedParentReadyRecoveryCall((context) =>
						this.#parentReadyAuthority.readParentReadyState(query, context));
					if (reread.kind !== "value" || reread.value === null) throw new Error("settlement reread failed");
					authorityState = validateParentReadyAuthorityState(reread.value);
				} catch {
					authorityState = null;
				}
			}
		}
		if (authorityState === null || authorityState.phase !== "ready_settled"
			|| authorityState.status !== "settled" || authorityState.fence !== 0
			|| authorityState.appliedRevision !== readyRevision || !sameOperation(authorityState)) {
			this.startTrackedParentReadyRecovery(plan, operation);
			return { kind: "blocked", blockers: ["parent_ready_quarantined"] };
		}
		return {
			kind: "ready",
			pullRequest: recovered,
			reused: comparison.mutation.applied === false,
		};
	}

	async prepareParentReadiness(
		plan: ParentOrchestrationPlan,
		integrationValues: readonly ChildIntegrationReceipt[],
		policyValue: ParentDecisionPolicy,
		context?: OrchestrationCallContext,
	): Promise<ParentReadinessPreparation> {
		return this.withLifecycle(context, async () => {
			plan = validateParentOrchestrationPlan(plan);
			return this.serializeEnsure(
				`${plan.repository}:ready:${plan.markers.parentPullRequest}`,
				() => this.prepareParentReadinessUnlocked(plan, integrationValues, policyValue),
			);
		});
	}

	async commitPreparedParentReadiness(
		plan: ParentOrchestrationPlan,
		integrationValues: readonly ChildIntegrationReceipt[],
		operation: PreparedParentReadyOperation,
		context?: OrchestrationCallContext,
	): Promise<ParentReadinessDecision> {
		return this.withLifecycle(context, async () => {
			plan = validateParentOrchestrationPlan(plan);
			return this.serializeEnsure(
				`${plan.repository}:ready:${plan.markers.parentPullRequest}`,
				() => this.commitPreparedParentReadinessUnlocked(plan, integrationValues, operation, true),
			);
		});
	}

	async reconcileParentReadiness(
		plan: ParentOrchestrationPlan,
		integrationValues: readonly ChildIntegrationReceipt[],
		policyValue: ParentDecisionPolicy,
		context?: OrchestrationCallContext,
	): Promise<ParentReadinessDecision> {
		return this.withLifecycle(context, async () => {
			plan = validateParentOrchestrationPlan(plan);
			return this.serializeEnsure(`${plan.repository}:ready:${plan.markers.parentPullRequest}`, async () => {
				const prepared = await this.prepareParentReadinessUnlocked(plan, integrationValues, policyValue);
				return prepared.kind === "prepared"
					? this.commitPreparedParentReadinessUnlocked(plan, integrationValues, prepared.operation, false)
					: prepared;
			});
		});
	}
}
