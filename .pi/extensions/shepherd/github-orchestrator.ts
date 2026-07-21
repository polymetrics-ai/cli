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
import { canonicalIssueBranch } from "./git-adapter.ts";
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
	pullRequestSnapshot: CanonicalPullRequestSnapshot;
	observation: PullRequestObservation;
	controllerProvenance: ControllerIntegrationProvenance;
	mutation: DurableMutationIntent;
}

export interface MarkParentReadyRequest extends GitHubPullRequestQuery {
	headSha: string;
	generation: number;
	decisionRequestId: string;
	authorization: ParentReadyAuthorization;
	mutation: DurableMutationIntent;
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
	planDigest: string;
	headSha: string;
	pullRequestRevision: number;
	digest: string;
}

export interface RollbackParentReadyRequest extends GitHubPullRequestQuery {
	headSha: string;
	generation: number;
	authorizationDigest: string;
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
	markParentReady(request: MarkParentReadyRequest, context: ExternalCallContext): Promise<DurableMutationResult<GitHubPullRequestEvidence>>;
	rollbackParentReady(request: RollbackParentReadyRequest, context: ExternalCallContext): Promise<DurableMutationResult<GitHubPullRequestEvidence>>;
	proveAncestry(query: GitAncestryQuery, context: ExternalCallContext): Promise<GitAncestryProof>;
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
				const requested = validateHumanDecisionRecord(await broker.request(request));
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

function validateChildIssue(value: unknown): GitHubChildIssue {
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
	const issue = validateChildIssue(issueValue);
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

function validateRoster(value: unknown): GitHubRosterSnapshot {
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

function createParentReadyAuthorization(
	plan: ParentOrchestrationPlan,
	pullRequest: GitHubPullRequestEvidence,
	decision: HumanDecisionRecord,
	review: IndependentReviewRecord,
	changedPaths: GitHubChangedPathEvidence,
	policies: Map<string, RequiredGitHubCheckPolicyObservation>,
	roster: { receipts: ChildIntegrationReceipt[]; ancestryProofs: GitAncestryProof[] },
): ParentReadyAuthorization {
	const policySet = [...policies.values()].sort((left, right) => left.baseBranch.localeCompare(right.baseBranch));
	const decisionDigest = canonicalDigest(validateHumanDecisionRecord(decision));
	const childRosterDigest = canonicalDigest({
		receipts: roster.receipts,
		ancestryProofs: roster.ancestryProofs,
	});
	const policySetDigest = canonicalDigest(policySet);
	const parentReviewDigest = independentReviewResultDigest(review);
	const parentPathDigest = changedPathEvidenceDigest(changedPaths);
	const payload = {
		schemaVersion: 1 as const,
		repository: plan.repository,
		parentIssue: plan.parentIssue,
		generation: plan.generation,
		pullRequest: pullRequest.number,
		decisionRequestId: decision.requestId,
		decisionDigest,
		childRosterDigest,
		policySetDigest,
		parentReviewDigest,
		parentPathDigest,
		planDigest: plan.canonical.digest,
		headSha: pullRequest.headSha,
		pullRequestRevision: pullRequest.revision,
	};
	return {
		...payload,
		digest: canonicalDigest(payload),
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

function validateReceipt(value: unknown): ChildIntegrationReceipt {
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

function validateDurableMutationResult<T>(
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
		&& pullRequest.baseSha === handoff.baseHead
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
	const expectedSnapshot = pullRequestEvidence === undefined ? undefined : createCanonicalPullRequestSnapshot(pullRequestEvidence);
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
): HumanDecisionRecord {
	const record = validateHumanDecisionRecord(value);
	assertHumanDecisionBinding(record, binding);
	if (!sameDecisionRequest(record, request)) throw new Error("human decision broker record does not match the exact request");
	if (prior !== undefined && (record.idempotencyMarker !== prior.idempotencyMarker
		|| record.createdAt !== prior.createdAt
		|| !canonicalDataEqual(record.requestComment, prior.requestComment))) {
		throw new Error("human decision broker record changed immutable request provenance");
	}
	return record;
}

export interface GitHubParentOrchestratorOptions {
	externalCallTimeoutMs?: number;
	shutdownTimeoutMs?: number;
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
	readonly #externalCallTimeoutMs: number;
	readonly #shutdownTimeoutMs: number;
	readonly #ensureLocks = new Map<string, Promise<void>>();
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
		broker?: ParentDecisionBroker,
		attestations?: AgentSessionAttestationSource,
		policySource?: RequiredCheckPolicySource,
		options: GitHubParentOrchestratorOptions = {},
	) {
		this.#transport = transport;
		this.#broker = broker;
		this.#attestations = attestations;
		this.#policySource = policySource;
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
		const active = {
			operation,
			controller,
			settlement: Promise.resolve(),
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
		const settlement = invocation.then(() => {}, () => {});
		active.settlement = settlement;
		this.#activeCalls.set(token, active);
		const ensureCalls = this.#ensureCallScope.getStore();
		ensureCalls?.add(settlement);
		void settlement.finally(() => {
			this.#activeCalls.delete(token);
			ensureCalls?.delete(settlement);
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
		if (outcome.kind === "value") return outcome.value;
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
		return authoritativeItems(raw, "child issue lookup").map(validateChildIssue);
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
					validateChildIssue,
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
		return authoritativeItems(raw, "roster lookup").map(validateRoster);
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
					validateRoster,
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
		);
		const decision = evaluateGitHubPullRequestEvidence(evidence, {
			...expected,
			changedPathEvidence: observation,
			minimumObservation: { revision: observation.revision, observedAt: observation.observedAt },
			requiredCheckPolicy,
			reviewTarget: canonicalTarget,
			attestations: attestations as AgentSessionAttestation[],
			}, { allowDraft, allowIntegrated });
		return { decision, observation, target: canonicalTarget };
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
	): Promise<ControllerIntegrationProvenance | null> {
		if (!currentChildPullRequestEligible(pullRequest, integratedState)
			|| !childPullRequestMatches(pullRequest, plan, child, handoff)
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
		const assessed = await this.evaluateEvidence(pullRequest, expected, reviewTarget, policy, false, integratedState !== undefined);
		if (assessed.decision.kind === "blocked"
			|| !sameStrings(assessed.observation.paths, handoff.changedScope)
			|| !reviewMatchesChild(assessed.decision.review, plan, child, pullRequest.number, handoff)) return null;
		return controllerIntegrationProvenance(plan, policy, policyObservation, assessed.observation, assessed.decision.review);
	}

	async integrateChild(
		plan: ParentOrchestrationPlan,
		child: MaterializedChildRecord,
		handoffValue: WorkspaceHandoffEvidence,
		context?: OrchestrationCallContext,
	): Promise<ChildIntegrationDecision> {
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
			if (!childPullRequestMatches(first, plan, canonicalChild, handoff)
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
				const receipt = validateReceipt(existingItems[0]);
				if (!receiptMatchesChild(receipt, plan, canonicalChild, first.number, handoff, first)) {
					throw new Error("existing child integration receipt is stale or mismatched");
				}
				if (first.headSha !== handoff.head) return { kind: "blocked", blockers: ["head_moved"] };
				if (!currentChildPullRequestEligible(first, receipt.observation.state)) {
					return { kind: "blocked", blockers: [first.draft ? "draft" : "pr_not_open"] };
				}
				const currentProvenance = await this.currentChildAuthorization(
					plan, canonicalChild, first, handoff, receipt.observation.state,
				);
				if (currentProvenance === null || !controllerAuthorizationMatches(
					currentProvenance, receipt.controllerProvenance, first, receipt.pullRequestSnapshot,
				)) {
					return { kind: "blocked", blockers: ["policy_moved"] };
				}
				return { kind: "integrated", receipt, reused: true };
			}
			if (!currentChildPullRequestEligible(first)) {
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
			const assessed = await this.evaluateEvidence(first, expected, reviewTarget, policy);
			if (assessed.decision.kind === "blocked") return assessed.decision;
			if (!sameStrings(assessed.observation.paths, handoff.changedScope)
				|| !reviewMatchesChild(assessed.decision.review, plan, canonicalChild, first.number, handoff)) {
				return { kind: "blocked", blockers: ["review_missing"] };
			}
			const second = await this.singlePullRequest(query);
			if (second === null) return { kind: "blocked", blockers: ["pull_request_missing"] };
			if (second.headSha !== handoff.head) return { kind: "blocked", blockers: ["head_moved"] };
			if (!childPullRequestMatches(second, plan, canonicalChild, handoff)
				|| !sameStrings(second.changedPaths, handoff.changedScope)) {
				return { kind: "blocked", blockers: ["resource_mismatch"] };
			}
			const revalidated = await this.evaluateEvidence(second, expected, reviewTarget, policy);
			if (revalidated.decision.kind === "blocked") return revalidated.decision;
			if (!sameStrings(revalidated.observation.paths, handoff.changedScope)
				|| !reviewMatchesChild(revalidated.decision.review, plan, canonicalChild, second.number, handoff)) {
				return { kind: "blocked", blockers: ["review_missing"] };
			}
			const policyObservation = await this.currentPolicyObservation(plan, canonicalChild.prBase, second.observedAt);
			if (policyObservation === null) {
				return { kind: "blocked", blockers: ["policy_moved"] };
			}
			const snapshot = createCanonicalPullRequestSnapshot(second);
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
			const request: IntegrateChildRequest = { ...requestIntent, mutation };
			let mutationError: unknown;
			let mutationApplied: boolean | undefined;
			try {
				const result = validateDurableMutationResult(
					await this.callExternal("integrateChild", (context) => this.#transport.integrateChild(request, context), true),
					mutation,
					validateReceipt,
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
				const receipt = validateReceipt(items[0]);
				const current = await this.singlePullRequest(query);
				if (current === null) return null;
				const currentProvenance = await this.currentChildAuthorization(
					plan, canonicalChild, current, handoff, receipt.observation.state,
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
			receipts = snapshot.map(validateReceipt);
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
				receipt = validateReceipt(items[0]);
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
					plan, materialized, pullRequest, handoff, receipt.observation.state,
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
		const intent = {
			repository: plan.repository,
			pullRequest: ready.number,
			headSha: ready.headSha,
			generation: plan.generation,
			authorizationDigest: authorization.digest,
		};
		const mutation = createDurableMutationIntent(
			"parent_ready_rollback",
			[plan.repository, plan.markers.parentPullRequest, ready.headSha, authorization.digest],
			intent,
			ready.revision,
		);
		const request: RollbackParentReadyRequest = { ...intent, mutation };
		let mutationError: unknown;
		let rollbackRevision: number | undefined;
		try {
			const result = validateDurableMutationResult(
				await this.callExternal(
					"rollbackParentReady",
					(context) => this.#transport.rollbackParentReady(request, context),
					true,
				),
				mutation,
				validateGitHubPullRequestEvidence,
			);
			rollbackRevision = result.value.revision;
			if (!result.value.draft || result.value.number !== ready.number || result.value.headSha !== ready.headSha
				|| result.value.revision <= ready.revision) {
				throw new Error("parent ready rollback did not restore a newer exact draft");
			}
		} catch (error) {
			mutationError = error;
		}
		const restored = await this.reconcileVisible(async () => {
			const current = await this.singlePullRequest({ repository: plan.repository, marker: plan.markers.parentPullRequest });
			return current !== null && current.draft && parentPullRequestMatches(current, plan)
				&& current.headSha === ready.headSha && current.revision > ready.revision
				&& (rollbackRevision === undefined || current.revision >= rollbackRevision) ? current : null;
		});
		if (restored !== null) return;
		if (mutationError !== undefined) throw mutationError;
		throw new Error("parent ready rollback was not durably visible");
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
		const query = { repository: plan.repository, marker: plan.markers.parentPullRequest };
		const first = await this.singlePullRequest(query);
		if (first === null) return { kind: "blocked", blockers: ["parent_pull_request_missing"] };
		if (!parentPullRequestMatches(first, plan)) {
			return { kind: "blocked", blockers: ["parent_pull_request_collision"] };
		}
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
			await this.#broker!.request(request, context), request, binding,
		), true);
		if (record.status === "expired") return { kind: "awaiting_human", reason: "expired" };
		if (record.status === "pending") {
			const polled = await this.callExternal("broker.poll", async (context) => validateBrokerRecord(
				await this.#broker!.poll(policy.requestId, binding, { signal: context.signal }, context),
				request,
				binding,
				record,
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
		const authorization = createParentReadyAuthorization(
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
		const markRequest: MarkParentReadyRequest = { ...markIntent, mutation };
		if (!second.draft) return { kind: "ready", pullRequest: second, reused: true };
		let mutationError: unknown;
		let mutationApplied: boolean | undefined;
		let readyRevision: number | undefined;
		try {
			const result = validateDurableMutationResult(
				await this.callExternal("markParentReady", (context) => this.#transport.markParentReady(markRequest, context), true),
				mutation,
				validateGitHubPullRequestEvidence,
			);
			mutationApplied = result.applied;
			readyRevision = result.value.revision;
			if (readyRevision <= second.revision) throw new Error("parent ready CAS revision did not advance");
		} catch (error) {
			mutationError = error;
		}
		const recovered = await this.reconcileVisible(async () => {
			try {
				const ready = await this.singlePullRequest(query);
				if (ready === null || ready.draft || !parentPullRequestMatches(ready, plan)
					|| ready.revision <= second.revision || (readyRevision !== undefined && ready.revision !== readyRevision)) return null;
				const readyRoster = await this.completeIntegrationRoster(plan, integrationValues, ready);
				const readyDecision = await this.evaluateEvidence(ready, expected, reviewTarget, requiredCheckPolicy);
				const readyPolicies = await this.currentPolicySet(plan, ready.observedAt);
				const readyRecord = await this.callExternal("broker.request", async (context) => validateBrokerRecord(
					await this.#broker!.request(request, context), request, binding,
				), true);
				return readyRoster !== null
					&& readyPolicies !== null
					&& readyRecord.status === "consumed"
					&& canonicalDataEqual(readyRecord, record)
					&& readyDecision.decision.kind === "eligible"
					&& sameStrings(readyDecision.observation.paths, ready.changedPaths)
					&& reviewMatchesParent(readyDecision.decision.review, plan, ready) ? ready : null;
			} catch {
				return null;
			}
		});
		if (recovered !== null) return {
			kind: "ready",
			pullRequest: recovered,
			reused: mutationError !== undefined || mutationApplied === false,
		};
		const visible = await this.singlePullRequest(query);
		if (visible !== null && !visible.draft && parentPullRequestMatches(visible, plan)
			&& visible.headSha === second.headSha && visible.revision > second.revision) {
			await this.rollbackParentReadyEffect(plan, visible, authorization);
			throw new Error("parent ready authorization moved after the effect; the parent was restored to draft");
		}
		if (mutationError instanceof Error && mutationError.message === "parent ready CAS revision did not advance") {
			return { kind: "blocked", blockers: ["parent_ready_revision_not_advanced"] };
		}
		if (mutationError !== undefined) throw mutationError;
		throw new Error("parent ready mutation was not durably visible");
		});
		});
	}
}
