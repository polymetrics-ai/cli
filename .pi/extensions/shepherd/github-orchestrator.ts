import { createHash } from "node:crypto";
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
	evaluateGitHubPullRequestEvidence,
	validateGitHubPullRequestEvidence,
	type GitHubEvidenceBlocker,
	type GitHubPullRequestEvidence,
} from "./github-evidence.ts";
import type {
	GitHubDecisionBroker,
	GitHubDecisionRequest,
} from "./github-decision-broker.ts";
import {
	assertHumanDecisionBinding,
	type HumanDecisionBinding,
	type HumanDecisionGate,
} from "./human-decision.ts";
import { reconcileAutonomy } from "./reconciler.ts";
import type { IndependentReviewRecord } from "./review-router.ts";
import type {
	ClaimedWorkspace,
	WorkspaceHandoffEvidence,
} from "./workspace-adapter.ts";

const MAX_CHILDREN = 64;
const MAX_LIST = 64;
const MAX_BODY_BYTES = 65_536;
const MAX_GITHUB_NUMBER = 2_147_483_647;
const SHA = /^[0-9a-f]{40}$/;
const IDENTITY = /^[0-9a-f]{64}$/;
const CHILD_ID = /^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$/;
const SLUG = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
const SKILL = /^[A-Za-z0-9][A-Za-z0-9:._-]{0,127}$/;
const REPOSITORY = /^[A-Za-z0-9][A-Za-z0-9._-]{0,99}\/[A-Za-z0-9][A-Za-z0-9._-]{0,99}$/;
const RFC3339_UTC = /^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(?:\.(\d{1,3}))?Z$/;
const UNSAFE_INLINE = /[\u0000-\u001f\u007f-\u009f\u00ad\u061c\u200b-\u200f\u2028-\u202e\u2060-\u206f\ufeff]/u;
const UNSAFE_BODY = /[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f-\u009f\u00ad\u061c\u200b-\u200f\u2028-\u202e\u2060-\u206f\ufeff]/u;

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
	children: BoundedChildRecord[];
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
	updatedAt: string;
}

export interface PublishRosterRequest extends GitHubRosterQuery {
	parentIssue: number;
	generation: number;
	body: string;
}

export interface GitHubPullRequestQuery {
	repository: string;
	pullRequest: number;
}

export interface ChildIntegrationReceipt {
	childId: string;
	pullRequest: number;
	generation: number;
	marker: string;
	baseSha: string;
	headSha: string;
	parentBranch: string;
	integratedAt: string;
}

export interface IntegrateChildRequest extends GitHubPullRequestQuery {
	childId: string;
	generation: number;
	marker: string;
	baseSha: string;
	headSha: string;
	parentBranch: string;
}

export interface MarkParentReadyRequest extends GitHubPullRequestQuery {
	headSha: string;
	generation: number;
	decisionRequestId: string;
}

export interface GitHubOrchestrationTransport {
	findChildIssues(query: ChildIssueMarkerQuery): Promise<GitHubChildIssue[]>;
	createChildIssue(request: CreateChildIssueRequest): Promise<GitHubChildIssue>;
	findPullRequests(query: PullRequestMarkerQuery): Promise<GitHubPullRequestEvidence[]>;
	createPullRequest(request: CreatePullRequestRequest): Promise<GitHubPullRequestEvidence>;
	findParentRosters(query: GitHubRosterQuery): Promise<GitHubRosterSnapshot[]>;
	publishParentRoster(request: PublishRosterRequest): Promise<GitHubRosterSnapshot>;
	findChildIntegration(query: GitHubPullRequestQuery): Promise<ChildIntegrationReceipt | null>;
	integrateChild(request: IntegrateChildRequest): Promise<ChildIntegrationReceipt>;
	markParentReady(request: MarkParentReadyRequest): Promise<GitHubPullRequestEvidence>;
}

export type ParentDecisionBroker = Pick<GitHubDecisionBroker, "request" | "poll" | "consume">;

export interface ParentDecisionPolicy {
	requestId: string;
	actorAllowlist: readonly string[];
	expiresAt: string;
	question: string;
}

export interface WorkspaceHandoffSource {
	captureHandoff(workspace: ClaimedWorkspace, verificationState: "passed"): Promise<WorkspaceHandoffEvidence>;
}

export type ChildIntegrationDecision =
	| { kind: "integrated"; receipt: ChildIntegrationReceipt; reused: boolean }
	| { kind: "blocked"; blockers: Array<GitHubEvidenceBlocker | "handoff_invalid" | "pull_request_missing"> };

export type ParentReadinessDecision =
	| { kind: "ready"; pullRequest: GitHubPullRequestEvidence; reused: boolean }
	| { kind: "blocked"; blockers: string[] }
	| { kind: "awaiting_human"; reason: "pending" | "expired" | "broker_unavailable" }
	| { kind: "rejected" };

type ExactRecord = Record<string, unknown>;

function exactRecord(value: unknown, required: readonly string[], optional: readonly string[] = []): ExactRecord {
	if (typeof value !== "object" || value === null || Array.isArray(value) || nodeTypes.isProxy(value)) {
		throw new Error("invalid parent orchestration shape");
	}
	const prototype = Object.getPrototypeOf(value);
	if (prototype !== Object.prototype && prototype !== null) throw new Error("invalid parent orchestration shape");
	const descriptors = Object.getOwnPropertyDescriptors(value);
	const allowed = new Set([...required, ...optional]);
	for (const key of required) {
		const descriptor = descriptors[key];
		if (descriptor === undefined || !Object.hasOwn(descriptor, "value") || descriptor.enumerable !== true) {
			throw new Error("invalid parent orchestration shape");
		}
	}
	for (const key of Reflect.ownKeys(descriptors)) {
		if (typeof key !== "string" || !allowed.has(key)) throw new Error("unknown parent orchestration field");
		const descriptor = descriptors[key];
		if (!Object.hasOwn(descriptor, "value") || descriptor.enumerable !== true) {
			throw new Error("invalid parent orchestration shape");
		}
	}
	return Object.fromEntries(Object.entries(descriptors).map(([key, descriptor]) => [key, descriptor.value]));
}

function inlineText(value: unknown, description: string, maximum = 1_024): string {
	if (typeof value !== "string" || value.length === 0 || value.length > maximum || Buffer.byteLength(value) > maximum
		|| value.trim() !== value || UNSAFE_INLINE.test(value)) {
		throw new Error(`invalid ${description}`);
	}
	return value;
}

function bodyText(value: unknown, description: string): string {
	if (typeof value !== "string" || value.length === 0 || value.length > MAX_BODY_BYTES
		|| Buffer.byteLength(value) > MAX_BODY_BYTES || UNSAFE_BODY.test(value)) {
		throw new Error(`${description} must be bounded safe text`);
	}
	return value.replace(/\r\n?/gu, "\n");
}

function githubNumber(value: unknown, description: string): number {
	if (!Number.isSafeInteger(value) || (value as number) < 1 || (value as number) > MAX_GITHUB_NUMBER) {
		throw new Error(`invalid ${description}`);
	}
	return value as number;
}

function generation(value: unknown): number {
	if (!Number.isSafeInteger(value) || (value as number) < 0 || (value as number) > MAX_GITHUB_NUMBER) {
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

function branch(value: unknown, description: string): string {
	const result = inlineText(value, description, 240);
	if (result.startsWith("-") || result.startsWith("/") || result.endsWith("/") || result.includes("\\")
		|| result.includes("..") || result.includes("//") || /[~^:?*\[\]{}]/u.test(result)
		|| result.split("/").some((segment) => segment === "" || segment === "." || segment === ".." || segment.endsWith("."))) {
		throw new Error(`invalid ${description}`);
	}
	return result;
}

function boundedArray(value: unknown, description: string, maximum = MAX_LIST, allowEmpty = false): unknown[] {
	if (!Array.isArray(value) || nodeTypes.isProxy(value) || Object.getPrototypeOf(value) !== Array.prototype) {
		throw new Error(`${description} must be a canonical array`);
	}
	const descriptors = Object.getOwnPropertyDescriptors(value);
	const lengthDescriptor = Object.getOwnPropertyDescriptor(value, "length");
	if (lengthDescriptor === undefined || !Object.hasOwn(lengthDescriptor, "value")
		|| !Number.isSafeInteger(lengthDescriptor.value) || lengthDescriptor.value < 0
		|| (!allowEmpty && lengthDescriptor.value === 0) || lengthDescriptor.value > maximum) {
		throw new Error(`${description} must be a bounded array of at most ${maximum} values`);
	}
	const length = lengthDescriptor.value as number;
	const values: unknown[] = [];
	for (const key of Reflect.ownKeys(descriptors)) {
		if (key === "length") continue;
		if (typeof key !== "string" || !/^(?:0|[1-9]\d*)$/u.test(key)) throw new Error(`${description} has an invalid array field`);
		const index = Number(key);
		const descriptor = descriptors[key];
		if (index >= length || !Object.hasOwn(descriptor, "value") || descriptor.enumerable !== true) {
			throw new Error(`${description} must contain only dense data values`);
		}
		values[index] = descriptor.value;
	}
	for (let index = 0; index < length; index += 1) {
		if (!Object.hasOwn(values, index)) throw new Error(`${description} must be a dense canonical array`);
	}
	return values;
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
	if (candidate.access !== "read_only" && candidate.access !== "mutating") throw new Error("invalid child access");
	const dependencies = validateStringList(candidate.dependsOn, "child dependencies", CHILD_ID, true);
	const scopes = validateStringList(candidate.writeScopes, "child write scopes", undefined, true);
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

export function createParentOrchestrationPlan(value: unknown): ParentOrchestrationPlan {
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
	const canonicalRepository = repository(candidate.repository);
	const parentIssue = githubNumber(candidate.parentIssue, "parent issue number");
	const canonicalGeneration = generation(candidate.generation);
	const title = inlineText(candidate.title, "parent title", 256);
	const objective = inlineText(candidate.objective, "parent objective", 4_096);
	const parentBranch = branch(candidate.parentBranch, "parent branch");
	const parentBaseBranch = branch(candidate.parentBaseBranch, "parent base branch");
	if (parentBranch === parentBaseBranch) throw new Error("parent branch and base branch must differ");
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
		children,
	});
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
	const canonicalStatuses = statusesForPlan(plan, statuses);
	const items = plan.children.map((child): DependencyWorkItem => ({
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
	const child = childFor(plan, childId);
	const issue = validateChildIssue(issueValue);
	assertChildIssueMatches(issue, plan, child);
	return {
		...child,
		issue: issue.number,
		branch: canonicalIssueBranch(issue.number, child.branch.slug),
	};
}

function validateRoster(value: unknown): GitHubRosterSnapshot {
	const candidate = exactRecord(value, ["id", "marker", "parentIssue", "generation", "body", "updatedAt"]);
	return {
		id: githubNumber(candidate.id, "roster resource ID"),
		marker: inlineText(candidate.marker, "roster marker", 512),
		parentIssue: githubNumber(candidate.parentIssue, "roster parent issue"),
		generation: generation(candidate.generation),
		body: bodyText(candidate.body, "roster body"),
		updatedAt: timestamp(candidate.updatedAt, "roster update timestamp"),
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
		"integratedAt",
	]);
	return {
		childId: inlineText(candidate.childId, "integration child ID", 64),
		pullRequest: githubNumber(candidate.pullRequest, "integrated pull request"),
		generation: generation(candidate.generation),
		marker: inlineText(candidate.marker, "integrated pull request marker", 512),
		baseSha: sha(candidate.baseSha, "integrated base SHA"),
		headSha: sha(candidate.headSha, "integrated head SHA"),
		parentBranch: branch(candidate.parentBranch, "integration parent branch"),
		integratedAt: timestamp(candidate.integratedAt, "integration timestamp"),
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

function childPullRequestBody(plan: ParentOrchestrationPlan, child: MaterializedChildRecord): string {
	return `Refs #${child.issue}\nRefs #${plan.parentIssue}\n\n${child.markers.pullRequest}`;
}

function childPullRequestMatches(
	pullRequest: GitHubPullRequestEvidence,
	plan: ParentOrchestrationPlan,
	child: MaterializedChildRecord,
	handoff: WorkspaceHandoffEvidence,
): boolean {
	return pullRequest.marker === child.markers.pullRequest
		&& pullRequest.title === child.title
		&& pullRequest.body === childPullRequestBody(plan, child)
		&& pullRequest.baseBranch === child.prBase
		&& pullRequest.headBranch === child.branch
		&& pullRequest.baseSha === handoff.baseHead;
}

function reviewMatchesChild(
	review: IndependentReviewRecord,
	plan: ParentOrchestrationPlan,
	child: MaterializedChildRecord,
	pullRequest: number,
	handoff: WorkspaceHandoffEvidence,
): boolean {
	return review.repository === plan.repository
		&& review.workItemId === child.id
		&& review.pullRequest === pullRequest
		&& review.generation === plan.generation
		&& review.baseSha === handoff.baseHead
		&& review.headSha === handoff.head
		&& sameStrings(review.changedPaths, handoff.changedScope)
		&& sameStrings(review.allowedScopes, child.writeScopes);
}

function receiptMatchesChild(
	receipt: ChildIntegrationReceipt,
	plan: ParentOrchestrationPlan,
	child: MaterializedChildRecord,
	pullRequest: number,
	handoff: WorkspaceHandoffEvidence,
): boolean {
	return receipt.childId === child.id
		&& receipt.pullRequest === pullRequest
		&& receipt.generation === plan.generation
		&& receipt.marker === child.markers.pullRequest
		&& receipt.baseSha === handoff.baseHead
		&& receipt.headSha === handoff.head
		&& receipt.parentBranch === plan.parentBranch;
}

function parentPullRequestBody(plan: ParentOrchestrationPlan): string {
	return `Closes #${plan.parentIssue}\n\n${plan.markers.parentPullRequest}`;
}

function parentPullRequestMatches(
	pullRequest: GitHubPullRequestEvidence,
	plan: ParentOrchestrationPlan,
): boolean {
	return pullRequest.marker === plan.markers.parentPullRequest
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
		&& review.baseSha === pullRequest.baseSha
		&& review.headSha === pullRequest.headSha
		&& sameStrings(review.allowedScopes, scopes)
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

export class GitHubParentOrchestrator {
	readonly #transport: GitHubOrchestrationTransport;
	readonly #broker?: ParentDecisionBroker;

	constructor(transport: GitHubOrchestrationTransport, broker?: ParentDecisionBroker) {
		this.#transport = transport;
		this.#broker = broker;
	}

	private async matchingIssues(plan: ParentOrchestrationPlan, child: BoundedChildRecord): Promise<GitHubChildIssue[]> {
		const raw = await this.#transport.findChildIssues({ repository: plan.repository, marker: child.markers.issue });
		return boundedArray(raw, "child issue lookup", MAX_LIST, true).map(validateChildIssue);
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

	async ensureChildIssue(plan: ParentOrchestrationPlan, childId: string): Promise<GitHubChildIssue> {
		const child = childFor(plan, childId);
		const existing = this.resolveIssueMatches(await this.matchingIssues(plan, child), plan, child);
		if (existing !== null) return existing;
		const request: CreateChildIssueRequest = {
			repository: plan.repository,
			parentIssue: plan.parentIssue,
			marker: child.markers.issue,
			title: child.title,
			body: child.issueBody,
		};
		try {
			const created = validateChildIssue(await this.#transport.createChildIssue(request));
			assertChildIssueMatches(created, plan, child);
			return created;
		} catch (error) {
			const recovered = this.resolveIssueMatches(await this.matchingIssues(plan, child), plan, child);
			if (recovered !== null) return recovered;
			throw error;
		}
	}

	private async matchingPullRequests(query: PullRequestMarkerQuery): Promise<GitHubPullRequestEvidence[]> {
		const raw = await this.#transport.findPullRequests(query);
		return boundedArray(raw, "pull request lookup", MAX_LIST, true).map(validateGitHubPullRequestEvidence);
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
			|| value.baseBranch !== request.baseBranch || value.headBranch !== request.headBranch
			|| value.baseSha !== request.baseSha || value.headSha !== request.headSha) {
			throw new Error("pull request marker collision or canonical resource mismatch");
		}
		return value;
	}

	private async ensurePullRequest(request: CreatePullRequestRequest): Promise<GitHubPullRequestEvidence> {
		const query = { repository: request.repository, marker: request.marker };
		const existing = await this.singlePullRequest(query);
		if (existing !== null) return this.assertPublishedPullRequest(existing, request);
		try {
			return this.assertPublishedPullRequest(
				validateGitHubPullRequestEvidence(await this.#transport.createPullRequest(request)),
				request,
			);
		} catch (error) {
			const recovered = await this.singlePullRequest(query);
			if (recovered !== null) return this.assertPublishedPullRequest(recovered, request);
			throw error;
		}
	}

	async captureChildHandoff(
		plan: ParentOrchestrationPlan,
		child: MaterializedChildRecord,
		workspace: ClaimedWorkspace,
		source: WorkspaceHandoffSource,
	): Promise<WorkspaceHandoffEvidence> {
		childFor(plan, child.id);
		const handoff = await source.captureHandoff(workspace, "passed");
		return validateHandoff(handoff, child.issue, child.branch, child.prBase, child.writeScopes);
	}

	async captureParentHandoff(
		plan: ParentOrchestrationPlan,
		workspace: ClaimedWorkspace,
		source: WorkspaceHandoffSource,
	): Promise<WorkspaceHandoffEvidence> {
		const handoff = await source.captureHandoff(workspace, "passed");
		return validateHandoff(
			handoff,
			plan.parentIssue,
			plan.parentBranch,
			plan.parentBaseBranch,
			aggregateScopes(plan),
		);
	}

	async ensureChildPullRequest(
		plan: ParentOrchestrationPlan,
		child: MaterializedChildRecord,
		handoffValue: WorkspaceHandoffEvidence,
	): Promise<GitHubPullRequestEvidence> {
		childFor(plan, child.id);
		const handoff = validateHandoff(handoffValue, child.issue, child.branch, child.prBase, child.writeScopes);
		return this.ensurePullRequest({
			repository: plan.repository,
			workItemId: child.id,
			generation: plan.generation,
			marker: child.markers.pullRequest,
			title: child.title,
			body: childPullRequestBody(plan, child),
			draft: false,
			baseBranch: child.prBase,
			headBranch: child.branch,
			baseSha: handoff.baseHead,
			headSha: handoff.head,
			changedPaths: handoff.changedScope,
			allowedScopes: child.writeScopes,
		});
	}

	async ensureParentDraftPullRequest(
		plan: ParentOrchestrationPlan,
		handoffValue: WorkspaceHandoffEvidence,
	): Promise<GitHubPullRequestEvidence> {
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
		});
	}

	private async rosterMatches(plan: ParentOrchestrationPlan): Promise<GitHubRosterSnapshot[]> {
		const raw = await this.#transport.findParentRosters({ repository: plan.repository, marker: plan.markers.roster });
		return boundedArray(raw, "roster lookup", MAX_LIST, true).map(validateRoster);
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
	): Promise<GitHubRosterSnapshot> {
		const statuses = statusesForPlan(plan, statusValue);
		const body = renderRoster(plan, statuses);
		const existing = this.resolveRoster(await this.rosterMatches(plan), plan);
		if (existing?.body === body) return existing;
		const request: PublishRosterRequest = {
			repository: plan.repository,
			marker: plan.markers.roster,
			parentIssue: plan.parentIssue,
			generation: plan.generation,
			body,
		};
		try {
			const published = validateRoster(await this.#transport.publishParentRoster(request));
			if (published.marker !== request.marker || published.parentIssue !== request.parentIssue
				|| published.generation !== request.generation || published.body !== request.body) {
				throw new Error("published parent roster does not match requested state");
			}
			return published;
		} catch (error) {
			const recovered = this.resolveRoster(await this.rosterMatches(plan), plan);
			if (recovered?.body === body) return recovered;
			throw error;
		}
	}

	async integrateChild(
		plan: ParentOrchestrationPlan,
		child: MaterializedChildRecord,
		handoffValue: WorkspaceHandoffEvidence,
	): Promise<ChildIntegrationDecision> {
		const planned = childFor(plan, child.id);
		if (child.issue < 1 || !Number.isSafeInteger(child.issue)
			|| child.branch !== canonicalIssueBranch(child.issue, planned.branch.slug)
			|| child.prBase !== planned.prBase
			|| child.title !== planned.title
			|| child.markers.pullRequest !== planned.markers.pullRequest
			|| !sameStrings(child.writeScopes, planned.writeScopes)) {
			throw new Error("materialized child does not match its parent orchestration plan");
		}
		let handoff: WorkspaceHandoffEvidence;
		try {
			handoff = validateHandoff(handoffValue, child.issue, child.branch, child.prBase, child.writeScopes);
		} catch {
			return { kind: "blocked", blockers: ["handoff_invalid"] };
		}
		const query = { repository: plan.repository, marker: child.markers.pullRequest };
		const first = await this.singlePullRequest(query);
		if (first === null) return { kind: "blocked", blockers: ["pull_request_missing"] };
		if (!childPullRequestMatches(first, plan, child, handoff)) {
			return { kind: "blocked", blockers: ["resource_mismatch"] };
		}
		const expected = {
			number: first.number,
			marker: child.markers.pullRequest,
			baseBranch: child.prBase,
			headBranch: child.branch,
			baseSha: handoff.baseHead,
			headSha: handoff.head,
		};
		const existing = await this.#transport.findChildIntegration({ repository: plan.repository, pullRequest: first.number });
		if (existing !== null) {
			const receipt = validateReceipt(existing);
			if (!receiptMatchesChild(receipt, plan, child, first.number, handoff)) {
				throw new Error("existing child integration receipt is stale or mismatched");
			}
			if (first.headSha !== handoff.head) return { kind: "blocked", blockers: ["head_moved"] };
			if (first.state === "closed") return { kind: "blocked", blockers: ["pr_not_open"] };
			if (first.draft) return { kind: "blocked", blockers: ["draft"] };
			return { kind: "integrated", receipt, reused: true };
		}
		const assessed = evaluateGitHubPullRequestEvidence(first, expected);
		if (assessed.kind === "blocked") return assessed;
		if (!reviewMatchesChild(assessed.review, plan, child, first.number, handoff)) {
			return { kind: "blocked", blockers: ["review_missing"] };
		}
		const second = await this.singlePullRequest(query);
		if (second === null) return { kind: "blocked", blockers: ["pull_request_missing"] };
		if (!childPullRequestMatches(second, plan, child, handoff)) {
			return { kind: "blocked", blockers: ["resource_mismatch"] };
		}
		const revalidated = evaluateGitHubPullRequestEvidence(second, expected);
		if (revalidated.kind === "blocked") return revalidated;
		if (!reviewMatchesChild(revalidated.review, plan, child, second.number, handoff)) {
			return { kind: "blocked", blockers: ["review_missing"] };
		}
		const request: IntegrateChildRequest = {
			repository: plan.repository,
			childId: child.id,
			pullRequest: second.number,
			generation: plan.generation,
			marker: child.markers.pullRequest,
			baseSha: handoff.baseHead,
			headSha: handoff.head,
			parentBranch: plan.parentBranch,
		};
		try {
			const receipt = validateReceipt(await this.#transport.integrateChild(request));
			if (!receiptMatchesChild(receipt, plan, child, request.pullRequest, handoff)) {
				throw new Error("child integration receipt does not match requested exact head");
			}
			return { kind: "integrated", receipt, reused: false };
		} catch (error) {
			const recovered = await this.#transport.findChildIntegration({ repository: plan.repository, pullRequest: second.number });
			if (recovered !== null) {
				const receipt = validateReceipt(recovered);
				if (receiptMatchesChild(receipt, plan, child, request.pullRequest, handoff)) {
					return { kind: "integrated", receipt, reused: true };
				}
			}
			throw error;
		}
	}

	private completeIntegrationRoster(
		plan: ParentOrchestrationPlan,
		values: readonly ChildIntegrationReceipt[],
	): ChildIntegrationReceipt[] | null {
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
				|| receipt.parentBranch !== plan.parentBranch;
		})) return null;
		return receipts;
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

	async reconcileParentReadiness(
		plan: ParentOrchestrationPlan,
		integrationValues: readonly ChildIntegrationReceipt[],
		policyValue: ParentDecisionPolicy,
	): Promise<ParentReadinessDecision> {
		if (this.completeIntegrationRoster(plan, integrationValues) === null) {
			return { kind: "blocked", blockers: ["children_incomplete"] };
		}
		const query = { repository: plan.repository, marker: plan.markers.parentPullRequest };
		const first = await this.singlePullRequest(query);
		if (first === null) return { kind: "blocked", blockers: ["parent_pull_request_missing"] };
		if (!parentPullRequestMatches(first, plan)) {
			return { kind: "blocked", blockers: ["parent_pull_request_collision"] };
		}
		const expected = {
			number: first.number,
			marker: plan.markers.parentPullRequest,
			baseBranch: plan.parentBaseBranch,
			headBranch: plan.parentBranch,
			baseSha: first.baseSha,
			headSha: first.headSha,
		};
		const assessed = evaluateGitHubPullRequestEvidence(first, expected, { allowDraft: true });
		if (assessed.kind === "blocked") return assessed;
		if (!reviewMatchesParent(assessed.review, plan, first)) {
			return { kind: "blocked", blockers: ["review_missing"] };
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
		const record = await this.#broker.request(request);
		assertHumanDecisionBinding(record, binding);
		let decision = record.status === "decided" || record.status === "consumed" ? record.decision : undefined;
		if (decision === undefined) {
			const polled = await this.#broker.poll(policy.requestId, binding);
			if (polled.status === "pending") return { kind: "awaiting_human", reason: "pending" };
			if (polled.status === "expired") return { kind: "awaiting_human", reason: "expired" };
			decision = polled.decision;
		}
		if (record.status !== "consumed") decision = await this.#broker.consume(policy.requestId, binding);
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
		const revalidated = evaluateGitHubPullRequestEvidence(second, expected, { allowDraft: true });
		if (revalidated.kind === "blocked") return revalidated;
		if (!reviewMatchesParent(revalidated.review, plan, second)) {
			return { kind: "blocked", blockers: ["review_missing"] };
		}
		const lifecycle = this.humanDecisionLifecycle(plan, "approve_merge");
		if (lifecycle.kind !== "transition" || lifecycle.to !== "MERGE") {
			return { kind: "blocked", blockers: ["autonomy_policy_blocked_parent_ready"] };
		}
		const markRequest: MarkParentReadyRequest = {
			repository: plan.repository,
			pullRequest: second.number,
			headSha: second.headSha,
			generation: plan.generation,
			decisionRequestId: policy.requestId,
		};
		if (!second.draft) return { kind: "ready", pullRequest: second, reused: true };
		try {
			const ready = validateGitHubPullRequestEvidence(await this.#transport.markParentReady(markRequest));
			const readyDecision = evaluateGitHubPullRequestEvidence(ready, expected);
			if (!parentPullRequestMatches(ready, plan) || readyDecision.kind !== "eligible"
				|| !reviewMatchesParent(readyDecision.review, plan, ready)) {
				throw new Error("parent ready result does not match approved exact head");
			}
			return { kind: "ready", pullRequest: ready, reused: false };
		} catch (error) {
			const recovered = await this.singlePullRequest(query);
			if (recovered !== null && parentPullRequestMatches(recovered, plan)) {
				const recoveredDecision = evaluateGitHubPullRequestEvidence(recovered, expected);
				if (recoveredDecision.kind === "eligible" && reviewMatchesParent(recoveredDecision.review, plan, recovered)) {
					return { kind: "ready", pullRequest: recovered, reused: true };
				}
			}
			throw error;
		}
	}
}
