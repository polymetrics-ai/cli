import { createHash, randomUUID } from "node:crypto";
import { constants } from "node:fs";
import { lstat, mkdir, open, readFile, readdir, rename, rm } from "node:fs/promises";
import { join } from "node:path";

import {
	productionPlanDigest,
	validateProductionParentPlan,
	type ProductionLifecycleFailureKind,
	type ProductionParentPlanDocument,
	type ProductionStageCheckpoint,
	type ProductionWorkspaceBinding,
} from "./autonomous-production-contract.ts";
import { readBoundedExactRecord } from "./review-router.ts";

const MAX_STATE_BYTES = 1024 * 1024;
const MAX_CHILDREN = 64;
const MAX_TEXT_BYTES = 48 * 1024;
const LOCK_RETRY_MS = 5;
const LOCK_ATTEMPTS = 2_000;
const SAFE_ID = /^[a-z0-9][a-z0-9_-]{0,127}$/;
const DIGEST = /^[0-9a-f]{64}$/;
const SHA = /^[0-9a-f]{40}$/;
const UNSAFE_TEXT = /[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f-\u009f\u00ad\u061c\u200b-\u200f\u2028-\u202e\u2060-\u206f\ufeff]/u;
const SECRET_TEXT = /(?:github_pat_|ghp_|sk-[A-Za-z0-9_-]{16,}|-----BEGIN [A-Z ]*PRIVATE KEY-----|(?:api[_-]?key|access[_-]?token|client[_-]?secret|password)\s*[:=])/iu;

export type ProductionRunStatus = "running" | "stopping" | "stopped" | "waiting_human" | "failed" | "completed";
export type ProductionRunStage =
	| "recovery"
	| "schedule"
	| "child_lifecycle"
	| "human_decision"
	| "blocked"
	| "completed";
export type ProductionChildStatus = "pending" | "running" | "blocked" | "succeeded" | "failed" | "cancelled";
export type ProductionChildStage =
	| "pending"
	| "workspace"
	| "implementation"
	| "verification"
	| "publication"
	| "review"
	| "correction"
	| "integration"
	| "succeeded"
	| "failed"
	| "cancelled";

export interface ProductionChildFailure {
	kind: ProductionLifecycleFailureKind;
	summary: string;
	at: string;
}

export interface ProductionOwnershipRefreshRecord {
	outcome: "rebased" | "reclaimed";
	previousClaimId: string;
	previousBaseHead: string;
	effectKey: string;
	refreshedAt: string;
}

export interface ProductionChildHeadReconciliationRecord {
	previousHead: string;
	head: string;
	effectKey: string;
	reconciledAt: string;
}

export interface ProductionOwnershipRefreshInput {
	childId: string;
	outcome: "rebased" | "reclaimed";
	previousClaimId: string;
	previousBaseHead: string;
	newBinding: ProductionWorkspaceBinding;
	effectKey: string;
	summary?: string;
	now?: Date;
}

export interface ProductionChildHeadReconciliationInput {
	childId: string;
	previousHead: string;
	head: string;
	checkpoint: ProductionStageCheckpoint;
	now?: Date;
}

/** A persisted child contains plan identity, not its task/prompt or model output. */
export interface ProductionChildRuntimeState {
	id: string;
	issue: number;
	slug: string;
	specDigest: string;
	dependsOn: string[];
	writeScopes: string[];
	maxAttempts: number;
	maxCorrections: number;
	attempts: number;
	authorizedAttempts: number;
	corrections: number;
	status: ProductionChildStatus;
	stage: ProductionChildStage;
	resumeStage?: ProductionChildStage;
	ownership?: ProductionWorkspaceBinding;
	checkpoint?: ProductionStageCheckpoint;
	ownershipRefresh?: ProductionOwnershipRefreshRecord;
	childHeadReconciliation?: ProductionChildHeadReconciliationRecord;
	retryAuthorization?: ProductionRetryAuthorizationRecord;
	lastFailure?: ProductionChildFailure;
}

export interface ProductionParentHumanGate {
	repository: string;
	pullRequest: number;
	generation: number;
	head: string;
	requestId: string;
	status: "prepared" | "pending" | "merged" | "rejected" | "invalidated";
	mergeEvidence?: {
		mergedAt: string;
		mergeCommitSha: string;
		revision: number;
		observedAt: string;
	};
	invalidationEvidence?: {
		currentHead: string;
		revision: number;
		observedAt: string;
	};
}

export interface ProductionChildHumanGate {
	childId: string;
	repository: string;
	issue: number;
	pullRequest?: number;
	generation: number;
	head?: string;
	requestId: string;
	reason: string;
	status: "pending" | "authorized" | "aborted";
}

export interface ProductionChildInterventionInput {
	childId: string;
	requestId: string;
	reason: string;
	pullRequest?: number;
	head?: string;
	now?: Date;
}

export interface ProductionRetryAuthorizationRecord {
	requestId: string;
	generation: number;
	authorizedAt: string;
}

export interface ProductionChildRetryAuthorizationInput {
	childId: string;
	requestId: string;
	now?: Date;
}

export interface ProductionAutonomousState {
	schemaVersion: 1;
	kind: "production_autonomous";
	parentIssue: number;
	repository: string;
	planId: string;
	planDigest: string;
	parentBranch: string;
	parentBaseBranch: string;
	runId: string;
	/** Immutable identity epoch for GitHub issues, PRs, reviews, and integration receipts. */
	resourceGeneration: number;
	generation: number;
	revision: number;
	maxConcurrency: number;
	timeoutMs: number;
	status: ProductionRunStatus;
	stage: ProductionRunStage;
	createdAt: string;
	updatedAt: string;
	idleReason?: string;
	terminalBlocker?: string;
	humanGate?: ProductionParentHumanGate;
	invalidatedParentGates?: ProductionParentHumanGate[];
	childGate?: ProductionChildHumanGate;
	children: ProductionChildRuntimeState[];
}

export interface ProductionStateFence {
	issue: number;
	revision: number;
	generation: number;
	runId: string;
}

export interface ProductionStateCreateOptions {
	runId: string;
	now?: Date;
	maxConcurrency?: number;
	timeoutMs?: number;
}

export interface ProductionStateStore {
	load(issue: number): Promise<ProductionAutonomousState | undefined>;
	create(state: ProductionAutonomousState): Promise<ProductionAutonomousState>;
	compareAndSwap(fence: ProductionStateFence, next: ProductionAutonomousState): Promise<ProductionAutonomousState>;
}

export class ProductionStateConflictError extends Error {
	constructor(message: string) {
		super(message);
		this.name = "ProductionStateConflictError";
	}
}

function exact(value: unknown, required: readonly string[], optional: readonly string[] = [], description = "production state") {
	return readBoundedExactRecord(value, required, optional, description);
}

function positive(value: unknown, description: string, allowZero = false): number {
	if (!Number.isSafeInteger(value) || (value as number) < (allowZero ? 0 : 1)) {
		throw new Error(`${description} must be a ${allowZero ? "non-negative" : "positive"} integer`);
	}
	return value as number;
}

function safeText(value: unknown, description: string, maximum = MAX_TEXT_BYTES): string {
	if (typeof value !== "string" || value.length === 0 || Buffer.byteLength(value) > maximum || UNSAFE_TEXT.test(value)) {
		throw new Error(`${description} must be bounded safe text`);
	}
	if (SECRET_TEXT.test(value)) throw new Error(`${description} contains sensitive credential material`);
	return value;
}

function timestamp(value: unknown, description: string): string {
	const candidate = safeText(value, description, 64);
	const parsed = new Date(candidate);
	if (!Number.isFinite(parsed.valueOf()) || parsed.toISOString() !== candidate) {
		throw new Error(`${description} must be a canonical timestamp`);
	}
	return candidate;
}

function digest(value: unknown, description: string): string {
	const candidate = safeText(value, description, 64);
	if (!DIGEST.test(candidate)) throw new Error(`${description} must be a SHA-256 digest`);
	return candidate;
}

function denseArray(value: unknown, description: string, maximum = MAX_CHILDREN): unknown[] {
	if (!Array.isArray(value) || value.length > maximum) throw new Error(`${description} must be a bounded dense array`);
	const result: unknown[] = [];
	for (let index = 0; index < value.length; index += 1) {
		const descriptor = Object.getOwnPropertyDescriptor(value, index);
		if (!descriptor || !Object.hasOwn(descriptor, "value") || descriptor.enumerable !== true) {
			throw new Error(`${description} must contain dense data values`);
		}
		result.push(descriptor.value);
	}
	return result;
}

function stringArray(value: unknown, description: string): string[] {
	const result = denseArray(value, description).map((item) => safeText(item, description, 4_096));
	if (new Set(result).size !== result.length) throw new Error(`${description} contains duplicates`);
	return result;
}

function validateWorkspace(value: unknown): ProductionWorkspaceBinding {
	const candidate = exact(value, [
		"claimId", "ownershipId", "repositoryIdentity", "worktreeIdentity", "cwd", "branch", "baseBranch",
		"baseHead", "head", "writeScopes",
	], [], "production workspace binding");
	return {
		claimId: safeText(candidate.claimId, "workspace claim ID", 256),
		ownershipId: safeText(candidate.ownershipId, "workspace ownership ID", 256),
		repositoryIdentity: safeText(candidate.repositoryIdentity, "workspace repository identity", 256),
		worktreeIdentity: safeText(candidate.worktreeIdentity, "worktree identity", 4_096),
		cwd: safeText(candidate.cwd, "workspace cwd", 4_096),
		branch: safeText(candidate.branch, "workspace branch", 256),
		baseBranch: safeText(candidate.baseBranch, "workspace base branch", 256),
		baseHead: safeText(candidate.baseHead, "workspace base head", 256),
		head: safeText(candidate.head, "workspace head", 256),
		writeScopes: stringArray(candidate.writeScopes, "workspace write scopes"),
	};
}

function validateReview(value: unknown): NonNullable<ProductionStageCheckpoint["review"]> {
	const candidate = exact(
		value,
		["status", "baseHead", "head", "findings"],
		["resultDigest", "authorizationDigest", "completedAt"],
		"production review checkpoint",
	);
	if (candidate.status !== "pending" && candidate.status !== "blocked" && candidate.status !== "clean") {
		throw new Error("invalid production review status");
	}
	const findings = denseArray(candidate.findings, "production review findings").map((item) => {
		const finding = exact(item, ["id", "summary"], ["disposition"], "production review finding");
		return {
			id: safeText(finding.id, "review finding ID", 256),
			summary: safeText(finding.summary, "review finding summary", 4_096),
			...(finding.disposition === undefined ? {} : { disposition: safeText(finding.disposition, "review disposition", 4_096) }),
		};
	});
	if (new Set(findings.map((finding) => finding.id)).size !== findings.length) throw new Error("duplicate review finding ID");
	if (candidate.status === "clean" && findings.length !== 0) throw new Error("clean review cannot retain findings");
	if (candidate.status === "blocked" && findings.length === 0) throw new Error("blocked review requires findings");
	if (candidate.status === "clean" && (candidate.resultDigest === undefined || candidate.completedAt === undefined)) {
		throw new Error("clean review requires result digest and completion time");
	}
	return {
		status: candidate.status,
		baseHead: safeText(candidate.baseHead, "review base head", 256),
		head: safeText(candidate.head, "review head", 256),
		...(candidate.resultDigest === undefined ? {} : { resultDigest: digest(candidate.resultDigest, "review result digest") }),
		...(candidate.authorizationDigest === undefined ? {} : { authorizationDigest: digest(candidate.authorizationDigest, "review authorization digest") }),
		...(candidate.completedAt === undefined ? {} : { completedAt: timestamp(candidate.completedAt, "review completion time") }),
		findings,
	};
}

function validateVerification(value: unknown): NonNullable<ProductionStageCheckpoint["verification"]> {
	const candidate = exact(value, ["status", "resultDigest", "commands"], [], "production verification checkpoint");
	if (candidate.status !== "passed" && candidate.status !== "failed") {
		throw new Error("invalid production verification status");
	}
	const commands = denseArray(candidate.commands, "production verification commands").map((value) => {
		const command = exact(value, ["id", "status"], ["failureKind"], "production verification command");
		if (command.status !== "passed" && command.status !== "failed") throw new Error("invalid verification command status");
		const failureKinds = ["spawn", "exit", "timeout", "output_limit", "aborted"];
		if ((command.status === "passed" && command.failureKind !== undefined)
			|| (command.status === "failed" && !failureKinds.includes(command.failureKind as string))) {
			throw new Error("verification command failure evidence is invalid");
		}
		return {
			id: safeText(command.id, "verification command ID", 128),
			status: command.status as "passed" | "failed",
			...(command.failureKind === undefined ? {} : {
				failureKind: command.failureKind as "spawn" | "exit" | "timeout" | "output_limit" | "aborted",
			}),
		};
	});
	if ((candidate.status === "passed" && commands.some((command) => command.status !== "passed"))
		|| (candidate.status === "failed" && commands.every((command) => command.status === "passed"))) {
		throw new Error("verification checkpoint status conflicts with its command evidence");
	}
	return {
		status: candidate.status,
		resultDigest: digest(candidate.resultDigest, "verification result digest"),
		commands,
	};
}

function validateCheckpoint(value: unknown): ProductionStageCheckpoint {
	const candidate = exact(
		value,
		["summary"],
		["effectKey", "effectKeys", "workspace", "verification", "pullRequest", "review", "integrationReceiptDigest", "parentHead"],
		"production stage checkpoint",
	);
	return {
		summary: safeText(candidate.summary, "checkpoint summary", 4_096),
		...(candidate.effectKey === undefined ? {} : { effectKey: safeText(candidate.effectKey, "checkpoint effect key", 256) }),
		...(candidate.effectKeys === undefined ? {} : { effectKeys: stringArray(candidate.effectKeys, "checkpoint effect keys") }),
		...(candidate.workspace === undefined ? {} : { workspace: validateWorkspace(candidate.workspace) }),
		...(candidate.verification === undefined ? {} : { verification: validateVerification(candidate.verification) }),
		...(candidate.pullRequest === undefined ? {} : { pullRequest: positive(candidate.pullRequest, "checkpoint pull request") }),
		...(candidate.review === undefined ? {} : { review: validateReview(candidate.review) }),
		...(candidate.integrationReceiptDigest === undefined ? {} : {
			integrationReceiptDigest: digest(candidate.integrationReceiptDigest, "integration receipt digest"),
		}),
		...(candidate.parentHead === undefined ? {} : { parentHead: safeText(candidate.parentHead, "checkpoint parent head", 256) }),
	};
}

function validateFailure(value: unknown): ProductionChildFailure {
	const candidate = exact(value, ["kind", "summary", "at"], [], "production child failure");
	const kinds: ProductionLifecycleFailureKind[] = ["retryable", "correction_required", "stale_parent", "human_required", "terminal"];
	if (!kinds.includes(candidate.kind as ProductionLifecycleFailureKind)) throw new Error("invalid production failure kind");
	return {
		kind: candidate.kind as ProductionLifecycleFailureKind,
		summary: safeText(candidate.summary, "failure summary", 4_096),
		at: timestamp(candidate.at, "failure timestamp"),
	};
}

function validateOwnershipRefresh(value: unknown): ProductionOwnershipRefreshRecord {
	const candidate = exact(
		value,
		["outcome", "previousClaimId", "previousBaseHead", "effectKey", "refreshedAt"],
		[],
		"production ownership refresh",
	);
	if (candidate.outcome !== "rebased" && candidate.outcome !== "reclaimed") {
		throw new Error("invalid production ownership refresh outcome");
	}
	return {
		outcome: candidate.outcome,
		previousClaimId: safeText(candidate.previousClaimId, "previous workspace claim ID", 256),
		previousBaseHead: safeText(candidate.previousBaseHead, "previous workspace base head", 256),
		effectKey: safeText(candidate.effectKey, "ownership refresh effect key", 256),
		refreshedAt: timestamp(candidate.refreshedAt, "ownership refresh time"),
	};
}

function validateChildHeadReconciliation(value: unknown): ProductionChildHeadReconciliationRecord {
	const candidate = exact(
		value,
		["previousHead", "head", "effectKey", "reconciledAt"],
		[],
		"production child head reconciliation",
	);
	const previousHead = safeText(candidate.previousHead, "previous child head", 40);
	const head = safeText(candidate.head, "reconciled child head", 40);
	if (!SHA.test(previousHead) || !SHA.test(head) || previousHead === head) {
		throw new Error("child head reconciliation requires two distinct exact heads");
	}
	return {
		previousHead,
		head,
		effectKey: safeText(candidate.effectKey, "child head reconciliation effect key", 256),
		reconciledAt: timestamp(candidate.reconciledAt, "child head reconciliation time"),
	};
}

function validateRetryAuthorization(value: unknown): ProductionRetryAuthorizationRecord {
	const candidate = exact(
		value,
		["requestId", "generation", "authorizedAt"],
		[],
		"production child retry authorization",
	);
	return {
		requestId: safeText(candidate.requestId, "retry authorization request ID", 256),
		generation: positive(candidate.generation, "retry authorization generation"),
		authorizedAt: timestamp(candidate.authorizedAt, "retry authorization time"),
	};
}

function validateChild(value: unknown): ProductionChildRuntimeState {
	const candidate = exact(value, [
		"id", "issue", "slug", "specDigest", "dependsOn", "writeScopes", "maxAttempts", "maxCorrections",
		"attempts", "authorizedAttempts", "corrections", "status", "stage",
	], ["resumeStage", "ownership", "checkpoint", "ownershipRefresh", "childHeadReconciliation", "retryAuthorization", "lastFailure"], "production child state");
	const id = safeText(candidate.id, "child ID", 128);
	if (!SAFE_ID.test(id)) throw new Error("invalid child ID");
	const statuses: ProductionChildStatus[] = ["pending", "running", "blocked", "succeeded", "failed", "cancelled"];
	const stages: ProductionChildStage[] = [
		"pending", "workspace", "implementation", "verification", "publication", "review", "correction",
		"integration", "succeeded", "failed", "cancelled",
	];
	if (!statuses.includes(candidate.status as ProductionChildStatus)) throw new Error("invalid child status");
	if (!stages.includes(candidate.stage as ProductionChildStage)) throw new Error("invalid child stage");
	if (candidate.resumeStage !== undefined && !stages.includes(candidate.resumeStage as ProductionChildStage)) {
		throw new Error("invalid child resume stage");
	}
	const resumeStage = candidate.resumeStage as ProductionChildStage | undefined;
	if (resumeStage === "pending" || resumeStage === "succeeded" || resumeStage === "failed" || resumeStage === "cancelled") {
		throw new Error("child resume stage must identify uncompleted lifecycle work");
	}
	const maxAttempts = positive(candidate.maxAttempts, "child attempt budget");
	const maxCorrections = positive(candidate.maxCorrections, "child correction budget", true);
	const attempts = positive(candidate.attempts, "child attempts", true);
	const authorizedAttempts = positive(candidate.authorizedAttempts, "authorized child attempts", true);
	const corrections = positive(candidate.corrections, "child corrections", true);
	if (authorizedAttempts > MAX_CHILDREN) throw new Error("authorized child attempts exceed the bounded intervention limit");
	if (attempts > maxAttempts + authorizedAttempts) throw new Error("child attempt budget exhausted");
	if (corrections > maxCorrections) throw new Error("child correction budget exhausted");
	if (candidate.status === "pending" && candidate.stage !== "pending") throw new Error("pending child must have a pending stage");
	if (candidate.status === "running" && attempts === 0) throw new Error("running child requires an accepted attempt");
	if (candidate.stage === "correction" && corrections === 0) throw new Error("correction stage requires a correction count");
	if (candidate.status === "succeeded" && candidate.stage !== "succeeded") throw new Error("succeeded child must have a succeeded stage");
	if (candidate.status === "failed" && candidate.stage !== "failed") throw new Error("failed child must have a failed stage");
	if (candidate.status === "cancelled" && candidate.stage !== "cancelled") throw new Error("cancelled child must have a cancelled stage");
	if (resumeStage !== undefined && candidate.status !== "pending" && candidate.status !== "cancelled" && candidate.status !== "blocked") {
		throw new Error("resume stage is only valid for interrupted or pending child work");
	}
	if (candidate.status === "cancelled" && resumeStage === undefined) {
		throw new Error("cancelled child must preserve its interrupted resume stage");
	}
	if (candidate.status === "pending" && attempts === 0 && resumeStage !== undefined) {
		throw new Error("fresh pending child cannot have an interrupted resume stage");
	}
	const ownership = candidate.ownership === undefined ? undefined : validateWorkspace(candidate.ownership);
	const checkpoint = candidate.checkpoint === undefined ? undefined : validateCheckpoint(candidate.checkpoint);
	const ownershipRefresh = candidate.ownershipRefresh === undefined ? undefined : validateOwnershipRefresh(candidate.ownershipRefresh);
	const childHeadReconciliation = candidate.childHeadReconciliation === undefined
		? undefined : validateChildHeadReconciliation(candidate.childHeadReconciliation);
	const retryAuthorization = candidate.retryAuthorization === undefined ? undefined : validateRetryAuthorization(candidate.retryAuthorization);
	if ((authorizedAttempts > 0) !== (retryAuthorization !== undefined)) {
		throw new Error("authorized child attempts require exact consumed gate evidence");
	}
	const writeScopes = stringArray(candidate.writeScopes, "child write scopes");
	if (ownership && JSON.stringify(ownership.writeScopes) !== JSON.stringify(writeScopes)) {
		throw new Error("workspace ownership scopes differ from the durable child binding");
	}
	if (ownership && checkpoint?.workspace && JSON.stringify(ownership) !== JSON.stringify(checkpoint.workspace)) {
		throw new Error("checkpoint workspace conflicts with immutable ownership binding");
	}
	if (candidate.status === "succeeded" && checkpoint?.integrationReceiptDigest === undefined) {
		throw new Error("succeeded child requires an integration receipt checkpoint");
	}
	return {
		id,
		issue: positive(candidate.issue, "child issue"),
		slug: safeText(candidate.slug, "child slug", 128),
		specDigest: digest(candidate.specDigest, "child spec digest"),
		dependsOn: stringArray(candidate.dependsOn, "child dependencies"),
		writeScopes,
		maxAttempts,
		maxCorrections,
		attempts,
		authorizedAttempts,
		corrections,
		status: candidate.status as ProductionChildStatus,
		stage: candidate.stage as ProductionChildStage,
		...(resumeStage === undefined ? {} : { resumeStage }),
		...(ownership === undefined ? {} : { ownership }),
		...(checkpoint === undefined ? {} : { checkpoint }),
		...(ownershipRefresh === undefined ? {} : { ownershipRefresh }),
		...(childHeadReconciliation === undefined ? {} : { childHeadReconciliation }),
		...(retryAuthorization === undefined ? {} : { retryAuthorization }),
		...(candidate.lastFailure === undefined ? {} : { lastFailure: validateFailure(candidate.lastFailure) }),
	};
}

function validateHumanGate(value: unknown): ProductionParentHumanGate {
	const candidate = exact(
		value,
		["repository", "pullRequest", "generation", "head", "requestId", "status"],
		["mergeEvidence", "invalidationEvidence"],
		"production parent human gate",
	);
	if (candidate.status !== "prepared" && candidate.status !== "pending" && candidate.status !== "merged" && candidate.status !== "rejected"
		&& candidate.status !== "invalidated") {
		throw new Error("invalid production parent human gate status");
	}
	let mergeEvidence: ProductionParentHumanGate["mergeEvidence"];
	if (candidate.mergeEvidence !== undefined) {
		const evidence = exact(
			candidate.mergeEvidence,
			["mergedAt", "mergeCommitSha", "revision", "observedAt"],
			[],
			"production parent merge evidence",
		);
		const mergeCommitSha = safeText(evidence.mergeCommitSha, "parent merge commit SHA", 40);
		if (!SHA.test(mergeCommitSha)) throw new Error("parent merge evidence requires an exact commit SHA");
		mergeEvidence = {
			mergedAt: timestamp(evidence.mergedAt, "parent merge time"),
			mergeCommitSha,
			revision: positive(evidence.revision, "parent merge revision"),
			observedAt: timestamp(evidence.observedAt, "parent merge observation time"),
		};
	}
	if ((candidate.status === "merged") !== (mergeEvidence !== undefined)) {
		throw new Error("only a merged parent gate may contain authoritative merge evidence");
	}
	let invalidationEvidence: ProductionParentHumanGate["invalidationEvidence"];
	if (candidate.invalidationEvidence !== undefined) {
		const evidence = exact(
			candidate.invalidationEvidence,
			["currentHead", "revision", "observedAt"],
			[],
			"production parent gate invalidation evidence",
		);
		const currentHead = safeText(evidence.currentHead, "invalidated parent head", 40);
		if (!SHA.test(currentHead) || currentHead === candidate.head) {
			throw new Error("parent gate invalidation requires a distinct exact current head");
		}
		invalidationEvidence = {
			currentHead,
			revision: positive(evidence.revision, "parent gate invalidation revision"),
			observedAt: timestamp(evidence.observedAt, "parent gate invalidation observation time"),
		};
	}
	if ((candidate.status === "invalidated") !== (invalidationEvidence !== undefined)) {
		throw new Error("only an invalidated parent gate may contain invalidation evidence");
	}
	return {
		repository: safeText(candidate.repository, "human gate repository", 256),
		pullRequest: positive(candidate.pullRequest, "human gate pull request"),
		generation: positive(candidate.generation, "human gate generation"),
		head: safeText(candidate.head, "human gate head", 256),
		requestId: safeText(candidate.requestId, "human gate request ID", 256),
		status: candidate.status,
		...(mergeEvidence === undefined ? {} : { mergeEvidence }),
		...(invalidationEvidence === undefined ? {} : { invalidationEvidence }),
	};
}

function validateChildGate(value: unknown): ProductionChildHumanGate {
	const candidate = exact(
		value,
		["childId", "repository", "issue", "generation", "requestId", "reason", "status"],
		["pullRequest", "head"],
		"production child human gate",
	);
	if (candidate.status !== "pending" && candidate.status !== "authorized" && candidate.status !== "aborted") {
		throw new Error("invalid production child human gate status");
	}
	if ((candidate.pullRequest === undefined) !== (candidate.head === undefined)) {
		throw new Error("child pull request intervention requires an exact head binding");
	}
	return {
		childId: safeText(candidate.childId, "child intervention ID", 128),
		repository: safeText(candidate.repository, "child intervention repository", 256),
		issue: positive(candidate.issue, "child intervention issue"),
		...(candidate.pullRequest === undefined ? {} : { pullRequest: positive(candidate.pullRequest, "child intervention pull request") }),
		generation: positive(candidate.generation, "child intervention generation"),
		...(candidate.head === undefined ? {} : { head: safeText(candidate.head, "child intervention head", 256) }),
		requestId: safeText(candidate.requestId, "child intervention request ID", 256),
		reason: safeText(candidate.reason, "child intervention reason", 4_096),
		status: candidate.status,
	};
}

export function validateProductionAutonomousState(value: unknown): ProductionAutonomousState {
	const candidate = exact(value, [
		"schemaVersion", "kind", "parentIssue", "repository", "planId", "planDigest", "parentBranch",
		"parentBaseBranch", "runId", "resourceGeneration", "generation", "revision", "maxConcurrency", "timeoutMs", "status", "stage",
		"createdAt", "updatedAt", "children",
	], ["idleReason", "terminalBlocker", "humanGate", "invalidatedParentGates", "childGate"]);
	if (candidate.schemaVersion !== 1 || candidate.kind !== "production_autonomous") throw new Error("unsupported production state schema");
	const statuses: ProductionRunStatus[] = ["running", "stopping", "stopped", "waiting_human", "failed", "completed"];
	const stages: ProductionRunStage[] = ["recovery", "schedule", "child_lifecycle", "human_decision", "blocked", "completed"];
	if (!statuses.includes(candidate.status as ProductionRunStatus)) throw new Error("invalid production run status");
	if (!stages.includes(candidate.stage as ProductionRunStage)) throw new Error("invalid production run stage");
	const createdAt = timestamp(candidate.createdAt, "state creation time");
	const updatedAt = timestamp(candidate.updatedAt, "state update time");
	if (updatedAt < createdAt) throw new Error("state update time precedes creation time");
	const children = denseArray(candidate.children, "production children").map(validateChild);
	if (children.length === 0) throw new Error("production state requires children");
	if (new Set(children.map((child) => child.id)).size !== children.length) throw new Error("duplicate child ID");
	const ids = new Set(children.map((child) => child.id));
	if (children.some((child) => child.dependsOn.some((dependency) => !ids.has(dependency) || dependency === child.id))) {
		throw new Error("invalid durable child dependency binding");
	}
	if (candidate.status === "completed" && (candidate.stage !== "completed" || children.some((child) => child.status !== "succeeded"))) {
		throw new Error("completed production run requires every child to be succeeded");
	}
	const humanGate = candidate.humanGate === undefined ? undefined : validateHumanGate(candidate.humanGate);
	const invalidatedParentGates = candidate.invalidatedParentGates === undefined
		? undefined
		: denseArray(candidate.invalidatedParentGates, "invalidated parent gates").map(validateHumanGate);
	if (invalidatedParentGates?.some((gate) => gate.status !== "invalidated"
		|| gate.repository !== candidate.repository || gate.generation > (candidate.generation as number))) {
		throw new Error("invalidated parent gate history is stale or malformed");
	}
	if (invalidatedParentGates !== undefined
		&& new Set(invalidatedParentGates.map((gate) => gate.requestId)).size !== invalidatedParentGates.length) {
		throw new Error("invalidated parent gate history contains duplicate requests");
	}
	const childGate = candidate.childGate === undefined ? undefined : validateChildGate(candidate.childGate);
	if (humanGate && (humanGate.repository !== candidate.repository || humanGate.generation !== candidate.generation)) {
		throw new Error("parent human gate is stale or bound to another repository");
	}
	if (childGate) {
		const child = children.find((entry) => entry.id === childGate.childId);
		if (!child || child.issue !== childGate.issue || childGate.repository !== candidate.repository
			|| childGate.generation !== candidate.generation) {
			throw new Error("child human gate is stale or bound to another child");
		}
	}
	if (humanGate && childGate) throw new Error("production state cannot wait on parent and child gates simultaneously");
	if (candidate.status === "waiting_human"
		&& (candidate.stage !== "human_decision"
			|| ((humanGate?.status === "pending" ? 1 : 0) + (childGate?.status === "pending" ? 1 : 0)) !== 1)) {
		throw new Error("human wait requires exactly one exact pending gate");
	}
	return {
		schemaVersion: 1,
		kind: "production_autonomous",
		parentIssue: positive(candidate.parentIssue, "parent issue"),
		repository: safeText(candidate.repository, "repository", 256),
		planId: safeText(candidate.planId, "plan ID", 256),
		planDigest: digest(candidate.planDigest, "plan digest"),
		parentBranch: safeText(candidate.parentBranch, "parent branch", 256),
		parentBaseBranch: safeText(candidate.parentBaseBranch, "parent base branch", 256),
		runId: safeText(candidate.runId, "run ID", 256),
		resourceGeneration: positive(candidate.resourceGeneration, "resource generation"),
		generation: positive(candidate.generation, "state generation"),
		revision: positive(candidate.revision, "state revision"),
		maxConcurrency: positive(candidate.maxConcurrency, "maximum concurrency"),
		timeoutMs: positive(candidate.timeoutMs, "production timeout"),
		status: candidate.status as ProductionRunStatus,
		stage: candidate.stage as ProductionRunStage,
		createdAt,
		updatedAt,
		...(candidate.idleReason === undefined ? {} : { idleReason: safeText(candidate.idleReason, "idle reason", 1_024) }),
		...(candidate.terminalBlocker === undefined ? {} : {
			terminalBlocker: safeText(candidate.terminalBlocker, "terminal blocker", 4_096),
		}),
		...(humanGate === undefined ? {} : { humanGate }),
		...(invalidatedParentGates === undefined ? {} : { invalidatedParentGates }),
		...(childGate === undefined ? {} : { childGate }),
		children,
	};
}

function childSpecDigest(child: ProductionParentPlanDocument["children"][number]): string {
	return createHash("sha256").update(JSON.stringify(child)).digest("hex");
}

export function createProductionAutonomousState(
	value: ProductionParentPlanDocument,
	options: ProductionStateCreateOptions,
): ProductionAutonomousState {
	const plan = validateProductionParentPlan(value, value.parentIssue);
	const now = (options.now ?? new Date()).toISOString();
	return validateProductionAutonomousState({
		schemaVersion: 1,
		kind: "production_autonomous",
		parentIssue: plan.parentIssue,
		repository: plan.repository,
		planId: plan.planId,
		planDigest: productionPlanDigest(plan),
		parentBranch: plan.parentBranch,
		parentBaseBranch: plan.parentBaseBranch,
		runId: safeText(options.runId, "run ID", 256),
		resourceGeneration: 1,
		generation: 1,
		revision: 1,
		maxConcurrency: positive(options.maxConcurrency ?? 1, "maximum concurrency"),
		timeoutMs: positive(options.timeoutMs ?? 30_000, "production timeout"),
		status: "running",
		stage: "recovery",
		createdAt: now,
		updatedAt: now,
		children: plan.children.map((child) => ({
			id: child.id,
			issue: child.issue,
			slug: child.slug,
			specDigest: childSpecDigest(child),
			dependsOn: [...child.dependsOn],
			writeScopes: [...child.writeScopes],
			maxAttempts: child.maxAttempts,
			maxCorrections: child.maxCorrections,
			attempts: 0,
			authorizedAttempts: 0,
			corrections: 0,
			status: "pending",
			stage: "pending",
		})),
	});
}

export function assertProductionPlanBinding(stateValue: ProductionAutonomousState, planValue: ProductionParentPlanDocument): void {
	const state = validateProductionAutonomousState(stateValue);
	const plan = validateProductionParentPlan(planValue, state.parentIssue);
	if (state.planId !== plan.planId || state.planDigest !== productionPlanDigest(plan)
		|| state.repository !== plan.repository || state.parentBranch !== plan.parentBranch
		|| state.parentBaseBranch !== plan.parentBaseBranch || state.children.length !== plan.children.length) {
		throw new ProductionStateConflictError("durable production plan binding changed");
	}
	const specs = new Map(plan.children.map((child) => [child.id, child]));
	for (const child of state.children) {
		const spec = specs.get(child.id);
		if (!spec || child.specDigest !== childSpecDigest(spec)) {
			throw new ProductionStateConflictError("durable production child binding changed");
		}
	}
}

function assertFence(state: ProductionAutonomousState, fence: ProductionStateFence): void {
	if (state.parentIssue !== fence.issue || state.revision !== fence.revision
		|| state.generation !== fence.generation || state.runId !== fence.runId) {
		throw new ProductionStateConflictError("stale production state CAS fence");
	}
}

function immutableBinding(state: ProductionAutonomousState) {
	return {
		parentIssue: state.parentIssue,
		repository: state.repository,
		planId: state.planId,
		planDigest: state.planDigest,
		parentBranch: state.parentBranch,
		parentBaseBranch: state.parentBaseBranch,
		resourceGeneration: state.resourceGeneration,
		createdAt: state.createdAt,
		maxConcurrency: state.maxConcurrency,
		timeoutMs: state.timeoutMs,
		children: state.children.map((child) => ({
			id: child.id,
			issue: child.issue,
			slug: child.slug,
			specDigest: child.specDigest,
			dependsOn: child.dependsOn,
			writeScopes: child.writeScopes,
			maxAttempts: child.maxAttempts,
			maxCorrections: child.maxCorrections,
		})),
	};
}

function assertTransition(previous: ProductionAutonomousState, next: ProductionAutonomousState): void {
	if (JSON.stringify(immutableBinding(previous)) !== JSON.stringify(immutableBinding(next))) {
		throw new ProductionStateConflictError("immutable production plan binding changed");
	}
	if (next.revision !== previous.revision + 1) throw new ProductionStateConflictError("production state revision must advance by one");
	if (next.generation !== previous.generation && next.generation !== previous.generation + 1) {
		throw new ProductionStateConflictError("production state generation must remain current or advance by one");
	}
	if (next.generation === previous.generation && next.runId !== previous.runId) {
		throw new ProductionStateConflictError("run identity changed without a generation fence");
	}
	if (next.generation === previous.generation + 1 && next.runId === previous.runId) {
		throw new ProductionStateConflictError("generation advance requires a fresh run identity");
	}
	for (let index = 0; index < previous.children.length; index += 1) {
		const before = previous.children[index];
		const after = next.children[index];
		let refreshedOwnership = false;
		if (after.attempts < before.attempts || after.corrections < before.corrections) {
			throw new ProductionStateConflictError("child retry or correction counter regressed");
		}
		if (after.authorizedAttempts < before.authorizedAttempts) {
			throw new ProductionStateConflictError("authorized child attempt count regressed");
		}
		if (after.authorizedAttempts > before.authorizedAttempts) {
			const expectedResumeStage = before.resumeStage ?? before.stage;
			if (after.authorizedAttempts !== before.authorizedAttempts + 1
				|| after.attempts !== before.attempts + 1
				|| after.resumeStage !== expectedResumeStage
				|| previous.childGate?.childId !== before.id
				|| previous.childGate.status !== "pending"
				|| next.childGate?.status !== "authorized"
				|| next.childGate.requestId !== previous.childGate.requestId
				|| after.retryAuthorization?.requestId !== previous.childGate.requestId
				|| after.retryAuthorization.generation !== next.generation
				|| next.status !== "running"
				|| next.stage !== "schedule") {
				throw new ProductionStateConflictError("authorized child attempt lacks exact consumed gate evidence");
			}
		} else if (JSON.stringify(before.retryAuthorization) !== JSON.stringify(after.retryAuthorization)) {
			throw new ProductionStateConflictError("child retry authorization changed without a consumed gate");
		}
		if (before.resumeStage !== undefined) {
			if (after.resumeStage === undefined) {
				if (after.status !== "running" || after.stage !== before.resumeStage) {
					throw new ProductionStateConflictError("child resume stage was discarded without exact-stage continuation");
				}
			} else if (after.resumeStage !== before.resumeStage) {
				throw new ProductionStateConflictError("child resume stage changed across a durable transition");
			}
		}
		if (!before.ownership && after.ownership
			&& JSON.stringify(after.ownership) !== JSON.stringify(after.checkpoint?.workspace)) {
			throw new ProductionStateConflictError("initial ownership requires an exact durable workspace checkpoint");
		}
		if (before.ownership && JSON.stringify(before.ownership) !== JSON.stringify(after.ownership)) {
			const exactPublishedHeadAdvance = after.ownership
				&& before.ownership.claimId === after.ownership.claimId
				&& before.ownership.ownershipId === after.ownership.ownershipId
				&& before.ownership.repositoryIdentity === after.ownership.repositoryIdentity
				&& before.ownership.worktreeIdentity === after.ownership.worktreeIdentity
				&& before.ownership.cwd === after.ownership.cwd
				&& before.ownership.branch === after.ownership.branch
				&& before.ownership.baseBranch === after.ownership.baseBranch
				&& before.ownership.baseHead === after.ownership.baseHead
				&& before.ownership.head !== after.ownership.head
				&& JSON.stringify(before.ownership.writeScopes) === JSON.stringify(after.ownership.writeScopes)
				&& JSON.stringify(after.ownership) === JSON.stringify(after.checkpoint?.workspace)
				&& (after.stage === "publication" || after.stage === "review"
					|| after.stage === "integration" || after.stage === "succeeded")
				&& typeof after.checkpoint?.effectKey === "string"
				&& after.checkpoint.effectKey !== before.checkpoint?.effectKey
				&& before.checkpoint?.integrationReceiptDigest === undefined
				&& (after.checkpoint.integrationReceiptDigest === undefined
					|| after.stage === "integration" || after.stage === "succeeded")
				&& JSON.stringify(before.ownershipRefresh) === JSON.stringify(after.ownershipRefresh)
				&& JSON.stringify(before.childHeadReconciliation) === JSON.stringify(after.childHeadReconciliation);
			const refresh = after.ownershipRefresh;
			const commonRefreshTruth = after.ownership && refresh
				&& refresh.previousClaimId === before.ownership.claimId
				&& refresh.previousBaseHead === before.ownership.baseHead
				&& refresh.effectKey === after.checkpoint?.effectKey
				&& JSON.stringify(after.ownership) === JSON.stringify(after.checkpoint?.workspace)
				&& JSON.stringify(after.ownership.writeScopes) === JSON.stringify(after.writeScopes)
				&& after.stage === "verification"
				&& after.checkpoint.review === undefined
				&& after.checkpoint.integrationReceiptDigest === undefined;
			const exactRebase = refresh?.outcome === "rebased" && after.ownership
				&& after.ownership.claimId === before.ownership.claimId
				&& after.ownership.ownershipId === before.ownership.ownershipId
				&& after.ownership.worktreeIdentity === before.ownership.worktreeIdentity
				&& after.ownership.cwd === before.ownership.cwd
				&& after.ownership.repositoryIdentity === before.ownership.repositoryIdentity
				&& after.ownership.branch === before.ownership.branch
				&& after.ownership.baseBranch === before.ownership.baseBranch
				&& after.ownership.baseHead !== before.ownership.baseHead
				&& after.ownership.head !== before.ownership.head;
			const exactReclaim = refresh?.outcome === "reclaimed" && after.ownership
				&& after.ownership.claimId !== before.ownership.claimId
				&& after.ownership.worktreeIdentity !== before.ownership.worktreeIdentity;
			const reconciliation = after.childHeadReconciliation;
			const exactChildHeadReconciliation = after.ownership && reconciliation
				&& reconciliation.previousHead === before.ownership.head
				&& reconciliation.head === after.ownership.head
				&& reconciliation.effectKey === after.checkpoint?.effectKey
				&& before.ownership.claimId === after.ownership.claimId
				&& before.ownership.ownershipId === after.ownership.ownershipId
				&& before.ownership.repositoryIdentity === after.ownership.repositoryIdentity
				&& before.ownership.worktreeIdentity === after.ownership.worktreeIdentity
				&& before.ownership.cwd === after.ownership.cwd
				&& before.ownership.branch === after.ownership.branch
				&& before.ownership.baseBranch === after.ownership.baseBranch
				&& before.ownership.baseHead === after.ownership.baseHead
				&& JSON.stringify(before.ownership.writeScopes) === JSON.stringify(after.ownership.writeScopes)
				&& JSON.stringify(after.ownership) === JSON.stringify(after.checkpoint?.workspace)
				&& after.stage === "verification"
				&& after.checkpoint?.verification === undefined
				&& after.checkpoint?.review === undefined
				&& after.checkpoint?.integrationReceiptDigest === undefined;
			if (!after.ownership || (!exactPublishedHeadAdvance
				&& (!refresh || !commonRefreshTruth || (!exactRebase && !exactReclaim))
				&& !exactChildHeadReconciliation)) {
				throw new ProductionStateConflictError("immutable child ownership binding changed without an exact refresh receipt");
			}
			refreshedOwnership = !exactPublishedHeadAdvance;
		}
		if (before.checkpoint?.integrationReceiptDigest
			&& before.checkpoint.integrationReceiptDigest !== after.checkpoint?.integrationReceiptDigest
			&& !refreshedOwnership) {
			throw new ProductionStateConflictError("integration checkpoint truth regressed");
		}
	}
	const previousInvalidated = previous.invalidatedParentGates ?? [];
	const nextInvalidated = next.invalidatedParentGates ?? [];
	if (nextInvalidated.length < previousInvalidated.length
		|| JSON.stringify(nextInvalidated.slice(0, previousInvalidated.length)) !== JSON.stringify(previousInvalidated)) {
		throw new ProductionStateConflictError("invalidated parent gate history changed");
	}
	const appendedInvalidations = nextInvalidated.slice(previousInvalidated.length);
	if (previous.humanGate && next.generation === previous.generation) {
		const {
			status: _beforeStatus,
			mergeEvidence: _beforeEvidence,
			invalidationEvidence: _beforeInvalidation,
			...beforeBinding
		} = previous.humanGate;
		if (next.humanGate) {
			const {
				status: _afterStatus,
				mergeEvidence: _afterEvidence,
				invalidationEvidence: _afterInvalidation,
				...afterBinding
			} = next.humanGate;
			if (JSON.stringify(beforeBinding) !== JSON.stringify(afterBinding)) {
				throw new ProductionStateConflictError("immutable parent human gate binding changed");
			}
			const validStatusAdvance = previous.humanGate.status === "prepared"
				? next.humanGate.status === "prepared" || next.humanGate.status === "pending"
				: previous.humanGate.status === "pending"
					? next.humanGate.status === "pending" || next.humanGate.status === "merged" || next.humanGate.status === "rejected"
					: next.humanGate.status === previous.humanGate.status;
			if (!validStatusAdvance) {
				throw new ProductionStateConflictError("terminal parent human gate status changed");
			}
			if (appendedInvalidations.length !== 0) {
				throw new ProductionStateConflictError("parent gate cannot remain active while archiving an invalidation");
			}
		} else {
			const archived = appendedInvalidations[0];
			const {
				status: _archivedStatus,
				mergeEvidence: _archivedMerge,
				invalidationEvidence: _archivedInvalidation,
				...archivedBinding
			} = archived ?? previous.humanGate;
			if (previous.humanGate.status !== "pending" || appendedInvalidations.length !== 1
				|| archived?.status !== "invalidated" || JSON.stringify(beforeBinding) !== JSON.stringify(archivedBinding)) {
				throw new ProductionStateConflictError("pending parent gate may be removed only by exact durable invalidation");
			}
		}
	} else if (appendedInvalidations.length !== 0) {
		throw new ProductionStateConflictError("invalidated parent gate history lacks its exact active predecessor");
	}
	if (previous.childGate && next.generation === previous.generation) {
		const { status: _beforeStatus, ...beforeBinding } = previous.childGate;
		const { status: _afterStatus, ...afterBinding } = next.childGate ?? previous.childGate;
		if (!next.childGate || JSON.stringify(beforeBinding) !== JSON.stringify(afterBinding)) {
			throw new ProductionStateConflictError("immutable child human gate binding changed");
		}
		if (previous.childGate.status !== "pending" && next.childGate.status !== previous.childGate.status) {
			throw new ProductionStateConflictError("terminal child human gate status changed");
		}
	}
}

export function waitForProductionChildIntervention(
	currentValue: ProductionAutonomousState,
	fence: ProductionStateFence,
	input: ProductionChildInterventionInput,
): ProductionAutonomousState {
	const current = validateProductionAutonomousState(currentValue);
	if (current.humanGate || current.childGate) throw new ProductionStateConflictError("another production human gate is already bound");
	const child = current.children.find((candidate) => candidate.id === input.childId);
	if (!child) throw new ProductionStateConflictError("child intervention target is not plan-bound");
	if ((input.pullRequest === undefined) !== (input.head === undefined)) {
		throw new ProductionStateConflictError("child pull request intervention requires an exact head binding");
	}
	return evolveProductionState(current, fence, (draft) => {
		draft.status = "waiting_human";
		draft.stage = "human_decision";
		draft.childGate = {
			childId: child.id,
			repository: current.repository,
			issue: child.issue,
			...(input.pullRequest === undefined ? {} : { pullRequest: input.pullRequest }),
			generation: current.generation,
			...(input.head === undefined ? {} : { head: input.head }),
			requestId: safeText(input.requestId, "child intervention request ID", 256),
			reason: safeText(input.reason, "child intervention reason", 4_096),
			status: "pending",
		};
		const target = draft.children.find((candidate) => candidate.id === child.id)!;
		target.status = "blocked";
	}, input.now ?? new Date());
}

export function authorizeProductionChildRetry(
	currentValue: ProductionAutonomousState,
	fence: ProductionStateFence,
	input: ProductionChildRetryAuthorizationInput,
): ProductionAutonomousState {
	const current = validateProductionAutonomousState(currentValue);
	const gate = current.childGate;
	if (current.status !== "waiting_human" || current.stage !== "human_decision"
		|| gate?.status !== "pending" || gate.childId !== input.childId || gate.requestId !== input.requestId) {
		throw new ProductionStateConflictError("child retry authorization does not match the pending intervention gate");
	}
	const child = current.children.find((candidate) => candidate.id === input.childId)!;
	if (child.attempts < child.maxAttempts + child.authorizedAttempts) {
		throw new ProductionStateConflictError("child retry budget is not exhausted");
	}
	const resumeStage = child.resumeStage ?? child.stage;
	if (resumeStage === "pending" || resumeStage === "succeeded" || resumeStage === "failed" || resumeStage === "cancelled") {
		throw new ProductionStateConflictError("child retry authorization requires an exact interrupted lifecycle stage");
	}
	const now = input.now ?? new Date();
	return evolveProductionState(current, fence, (draft) => {
		draft.status = "running";
		draft.stage = "schedule";
		draft.childGate!.status = "authorized";
		const target = draft.children.find((candidate) => candidate.id === input.childId)!;
		target.authorizedAttempts += 1;
		target.attempts += 1;
		target.resumeStage = resumeStage;
		target.retryAuthorization = {
			requestId: input.requestId,
			generation: draft.generation,
			authorizedAt: now.toISOString(),
		};
		target.status = "pending";
		target.stage = "pending";
	}, now);
}

export function refreshProductionChildOwnership(
	currentValue: ProductionAutonomousState,
	fence: ProductionStateFence,
	input: ProductionOwnershipRefreshInput,
): ProductionAutonomousState {
	const current = validateProductionAutonomousState(currentValue);
	const child = current.children.find((candidate) => candidate.id === input.childId);
	if (!child?.ownership) throw new ProductionStateConflictError("ownership refresh requires an existing child claim");
	if (child.ownership.claimId !== input.previousClaimId || child.ownership.baseHead !== input.previousBaseHead) {
		throw new ProductionStateConflictError("ownership refresh does not match the previous claim and base");
	}
	const newBinding = validateWorkspace(input.newBinding);
	if (JSON.stringify(newBinding.writeScopes) !== JSON.stringify(child.writeScopes)) {
		throw new ProductionStateConflictError("refreshed ownership scopes differ from the durable child binding");
	}
	if (input.outcome === "rebased") {
		if (newBinding.claimId !== child.ownership.claimId
			|| newBinding.ownershipId !== child.ownership.ownershipId
			|| newBinding.worktreeIdentity !== child.ownership.worktreeIdentity
			|| newBinding.cwd !== child.ownership.cwd
			|| newBinding.baseHead === child.ownership.baseHead
			|| newBinding.head === child.ownership.head) {
			throw new ProductionStateConflictError("rebased ownership must preserve its claim and advance exact base/head truth");
		}
	} else if (input.outcome === "reclaimed") {
		if (newBinding.claimId === child.ownership.claimId || newBinding.worktreeIdentity === child.ownership.worktreeIdentity) {
			throw new ProductionStateConflictError("reclaimed ownership requires a new claim and worktree identity");
		}
	} else {
		throw new ProductionStateConflictError("invalid ownership refresh outcome");
	}
	const effectKey = safeText(input.effectKey, "ownership refresh effect key", 256);
	const now = input.now ?? new Date();
	return evolveProductionState(current, fence, (draft) => {
		const target = draft.children.find((candidate) => candidate.id === input.childId)!;
		target.ownership = newBinding;
		target.ownershipRefresh = {
			outcome: input.outcome,
			previousClaimId: input.previousClaimId,
			previousBaseHead: input.previousBaseHead,
			effectKey,
			refreshedAt: now.toISOString(),
		};
		target.status = "running";
		target.stage = "verification";
		target.checkpoint = {
			summary: safeText(input.summary ?? "workspace ownership refreshed; verification required", "ownership refresh summary", 4_096),
			effectKey,
			workspace: newBinding,
		};
	}, now);
}

export function reconcileProductionChildHead(
	currentValue: ProductionAutonomousState,
	fence: ProductionStateFence,
	input: ProductionChildHeadReconciliationInput,
): ProductionAutonomousState {
	const current = validateProductionAutonomousState(currentValue);
	const child = current.children.find((candidate) => candidate.id === input.childId);
	if (!child?.ownership || !child.checkpoint?.pullRequest) {
		throw new ProductionStateConflictError("child-head reconciliation requires durable ownership and pull request truth");
	}
	const nextBinding = input.checkpoint.workspace === undefined ? undefined : validateWorkspace(input.checkpoint.workspace);
	const effectKey = input.checkpoint.effectKey === undefined
		? undefined : safeText(input.checkpoint.effectKey, "child head reconciliation effect key", 256);
	if (!nextBinding || !effectKey || child.ownership.head !== input.previousHead
		|| nextBinding.head !== input.head || input.previousHead === input.head
		|| child.ownership.claimId !== nextBinding.claimId
		|| child.ownership.ownershipId !== nextBinding.ownershipId
		|| child.ownership.repositoryIdentity !== nextBinding.repositoryIdentity
		|| child.ownership.worktreeIdentity !== nextBinding.worktreeIdentity
		|| child.ownership.cwd !== nextBinding.cwd
		|| child.ownership.branch !== nextBinding.branch
		|| child.ownership.baseBranch !== nextBinding.baseBranch
		|| child.ownership.baseHead !== nextBinding.baseHead
		|| JSON.stringify(child.ownership.writeScopes) !== JSON.stringify(nextBinding.writeScopes)) {
		throw new ProductionStateConflictError("child-head reconciliation moved outside its exact durable workspace binding");
	}
	const now = input.now ?? new Date();
	return evolveProductionState(current, fence, (draft) => {
		const target = draft.children.find((candidate) => candidate.id === input.childId)!;
		const prior = target.checkpoint!;
		const effectKeys = [...new Set([
			...(prior.effectKey === undefined ? [] : [prior.effectKey]),
			...(prior.effectKeys ?? []),
			...(input.checkpoint.effectKey === undefined ? [] : [input.checkpoint.effectKey]),
			...(input.checkpoint.effectKeys ?? []),
		])];
		target.ownership = nextBinding;
		target.status = "running";
		target.stage = "verification";
		target.checkpoint = {
			...prior,
			...input.checkpoint,
			workspace: nextBinding,
			...(effectKeys.length === 0 ? {} : { effectKeys }),
		};
		delete target.checkpoint.verification;
		delete target.checkpoint.review;
		delete target.checkpoint.integrationReceiptDigest;
		target.childHeadReconciliation = {
			previousHead: input.previousHead,
			head: input.head,
			effectKey,
			reconciledAt: now.toISOString(),
		};
	}, now);
}

export function evolveProductionState(
	currentValue: ProductionAutonomousState,
	fence: ProductionStateFence,
	mutate: (draft: ProductionAutonomousState) => void,
	now = new Date(),
): ProductionAutonomousState {
	const current = validateProductionAutonomousState(currentValue);
	assertFence(current, fence);
	const draft = structuredClone(current);
	mutate(draft);
	draft.revision = current.revision + 1;
	draft.updatedAt = now.toISOString();
	const next = validateProductionAutonomousState(draft);
	assertTransition(current, next);
	return next;
}

export function advanceProductionGeneration(
	currentValue: ProductionAutonomousState,
	fence: ProductionStateFence,
	runId: string,
	now = new Date(),
): ProductionAutonomousState {
	return evolveProductionState(currentValue, fence, (draft) => {
		draft.generation += 1;
		draft.runId = safeText(runId, "run ID", 256);
		draft.status = "running";
		draft.stage = "recovery";
		delete draft.idleReason;
		delete draft.terminalBlocker;
		delete draft.humanGate;
		delete draft.childGate;
		for (const child of draft.children) {
			if (child.status === "running") {
				child.resumeStage = child.stage;
				child.status = "pending";
				child.stage = "pending";
			} else if (child.status === "cancelled") {
				if (child.resumeStage === undefined) {
					throw new ProductionStateConflictError("cancelled child lost its interrupted resume stage");
				}
				child.status = "pending";
				child.stage = "pending";
			}
		}
	}, now);
}

function isErrno(error: unknown, code: string): boolean {
	return typeof error === "object" && error !== null && "code" in error && (error as { code?: unknown }).code === code;
}

interface FileLock {
	path: string;
	token: string;
}

export class ProductionFileStateStore implements ProductionStateStore {
	readonly #root: string;

	constructor(root: string) {
		if (typeof root !== "string" || root.length === 0) throw new Error("production state root is required");
		this.#root = root;
	}

	#path(issue: number): string {
		return join(this.#root, `production-issue-${positive(issue, "parent issue")}.json`);
	}

	#lockPath(issue: number): string {
		return join(this.#root, `.production-issue-${positive(issue, "parent issue")}.lock`);
	}

	async #ensureRoot(): Promise<void> {
		await mkdir(this.#root, { recursive: true, mode: 0o700 });
		const metadata = await lstat(this.#root);
		if (!metadata.isDirectory() || metadata.isSymbolicLink()) throw new Error("production state root must be a trusted directory");
	}

	async #syncRoot(): Promise<void> {
		if (process.platform === "win32") return;
		const directory = await open(this.#root, constants.O_RDONLY);
		try { await directory.sync(); } finally { await directory.close(); }
	}

	#processIsAlive(pid: number): boolean {
		try {
			process.kill(pid, 0);
			return true;
		} catch (error) {
			return !isErrno(error, "ESRCH");
		}
	}

	async #reclaimDeadLock(path: string): Promise<boolean> {
		let value: unknown;
		try { value = JSON.parse(await readFile(join(path, "owner.json"), "utf8")); } catch (error) {
			if (isErrno(error, "ENOENT")) return false;
			return false;
		}
		let owner: ReturnType<typeof exact>;
		try { owner = exact(value, ["schemaVersion", "pid", "token"], [], "production lock owner"); } catch { return false; }
		if (owner.schemaVersion !== 1 || !Number.isSafeInteger(owner.pid) || (owner.pid as number) < 1
			|| typeof owner.token !== "string" || !/^[0-9a-f-]{36}$/.test(owner.token)) return false;
		if (this.#processIsAlive(owner.pid as number)) return false;
		const quarantine = `${path}.stale.${randomUUID()}`;
		try { await rename(path, quarantine); } catch (error) {
			if (isErrno(error, "ENOENT")) return true;
			throw error;
		}
		try {
			const moved = exact(
				JSON.parse(await readFile(join(quarantine, "owner.json"), "utf8")),
				["schemaVersion", "pid", "token"],
				[],
				"production lock owner",
			);
			if (moved.token !== owner.token || moved.pid !== owner.pid) {
				throw new Error("production state stale lock identity changed during reclamation");
			}
		} finally { await rm(quarantine, { recursive: true, force: true }); }
		return true;
	}

	async #acquire(issue: number): Promise<FileLock> {
		await this.#ensureRoot();
		const path = this.#lockPath(issue);
		for (let attempt = 0; attempt < LOCK_ATTEMPTS; attempt += 1) {
			const token = randomUUID();
			try {
				await mkdir(path, { mode: 0o700 });
				const owner = await open(join(path, "owner.json"), "wx", 0o600);
				try {
					await owner.writeFile(JSON.stringify({ schemaVersion: 1, pid: process.pid, token }), "utf8");
					await owner.sync();
				} finally { await owner.close(); }
				return { path, token };
			} catch (error) {
				if (!isErrno(error, "EEXIST")) {
					await rm(path, { recursive: true, force: true });
					throw error;
				}
				if (await this.#reclaimDeadLock(path)) continue;
				await new Promise((resolve) => setTimeout(resolve, LOCK_RETRY_MS));
			}
		}
		throw new Error("timed out acquiring production state lock");
	}

	async #release(lock: FileLock): Promise<void> {
		let owner: unknown;
		try { owner = JSON.parse(await readFile(join(lock.path, "owner.json"), "utf8")); } catch {
			throw new Error("production state lock ownership disappeared before release");
		}
		const candidate = exact(owner, ["schemaVersion", "pid", "token"], [], "production lock owner");
		if (candidate.schemaVersion !== 1 || candidate.pid !== process.pid || candidate.token !== lock.token) {
			throw new Error("production state lock ownership changed before release");
		}
		await rm(lock.path, { recursive: true });
	}

	async #read(issue: number): Promise<ProductionAutonomousState | undefined> {
		let handle: Awaited<ReturnType<typeof open>>;
		try { handle = await open(this.#path(issue), constants.O_RDONLY | (constants.O_NOFOLLOW ?? 0)); } catch (error) {
			if (isErrno(error, "ENOENT")) return undefined;
			throw error;
		}
		try {
			const metadata = await handle.stat();
			if (!metadata.isFile() || metadata.size > MAX_STATE_BYTES) throw new Error("production state is not a bounded regular file");
			let parsed: unknown;
			try { parsed = JSON.parse(await handle.readFile("utf8")); } catch { throw new Error("invalid production state JSON"); }
			return validateProductionAutonomousState(parsed);
		} finally { await handle.close(); }
	}

	async #write(state: ProductionAutonomousState): Promise<void> {
		const snapshot = validateProductionAutonomousState(state);
		const serialized = `${JSON.stringify(snapshot, null, 2)}\n`;
		if (Buffer.byteLength(serialized) > MAX_STATE_BYTES) throw new Error("production state exceeds its byte limit");
		const temporary = join(this.#root, `.production-issue-${snapshot.parentIssue}.${randomUUID()}.tmp`);
		const handle = await open(temporary, "wx", 0o600);
		try {
			await handle.writeFile(serialized, "utf8");
			await handle.sync();
		} finally { await handle.close(); }
		try {
			await rename(temporary, this.#path(snapshot.parentIssue));
			await this.#syncRoot();
		} catch (error) {
			await rm(temporary, { force: true });
			throw error;
		}
	}

	async load(issue: number): Promise<ProductionAutonomousState | undefined> {
		await this.#ensureRoot();
		return this.#read(issue);
	}

	async create(stateValue: ProductionAutonomousState): Promise<ProductionAutonomousState> {
		const state = validateProductionAutonomousState(stateValue);
		if (state.revision !== 1 || state.generation !== 1) throw new Error("new production state must begin at revision and generation one");
		const lock = await this.#acquire(state.parentIssue);
		try {
			if (await this.#read(state.parentIssue)) throw new ProductionStateConflictError("production state already exists");
			await this.#write(state);
			return structuredClone(state);
		} finally { await this.#release(lock); }
	}

	async compareAndSwap(fence: ProductionStateFence, nextValue: ProductionAutonomousState): Promise<ProductionAutonomousState> {
		const next = validateProductionAutonomousState(nextValue);
		if (next.parentIssue !== fence.issue) throw new ProductionStateConflictError("production state issue does not match its CAS fence");
		const lock = await this.#acquire(fence.issue);
		try {
			const current = await this.#read(fence.issue);
			if (!current) throw new ProductionStateConflictError("production state does not exist");
			assertFence(current, fence);
			assertTransition(current, next);
			await this.#write(next);
			return structuredClone(next);
		} finally { await this.#release(lock); }
	}
}
