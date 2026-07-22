import { createHash } from "node:crypto";

import type { HumanDecisionGate } from "./human-decision.ts";
import { readBoundedExactRecord } from "./review-router.ts";

const MAX_CHILDREN = 64;
const MAX_LIST = 64;
const MAX_TEXT_BYTES = 48 * 1024;
const SAFE_ID = /^[a-z0-9][a-z0-9_-]{0,63}$/;
const SAFE_SLUG = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
const SAFE_REPOSITORY = /^[A-Za-z0-9][A-Za-z0-9._-]{0,99}\/[A-Za-z0-9][A-Za-z0-9._-]{0,99}$/;
const SAFE_REF = /^(?!\/|.*(?:\.\.|\s|[~^:?*\\\[\]])|.*\/$)[A-Za-z0-9][A-Za-z0-9._\/-]{0,239}$/;
const SAFE_EXECUTABLE = /^[A-Za-z0-9][A-Za-z0-9._+-]{0,127}$/;
const UNSAFE_TEXT = /[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f-\u009f\u00ad\u061c\u200b-\u200f\u2028-\u202e\u2060-\u206f\ufeff]/u;

export type ProductionEffectKind =
	| "workspace_claim"
	| "agent_implementation"
	| "agent_correction"
	| "shell_verification"
	| "git_commit"
	| "git_push"
	| "child_pull_request"
	| "independent_review"
	| "child_integration"
	| "parent_refresh"
	| "human_request"
	| "human_consume"
	| "parent_merge_observation";

export type ProductionEffectPhase = "prepared" | "observed" | "applied";

export interface ProductionEffectRecord {
	schemaVersion: 1;
	key: string;
	kind: ProductionEffectKind;
	phase: ProductionEffectPhase;
	runId: string;
	generation: number;
	childId?: string;
	intentDigest: string;
	preparedAt: string;
	observedAt?: string;
	appliedAt?: string;
	resultDigest?: string;
}

export interface ProductionVerificationCommand {
	id: string;
	executable: string;
	args: string[];
	cwd: string;
	timeoutMs: number;
	maxOutputBytes: number;
}

export interface ProductionChildSpec {
	id: string;
	issue: number;
	title: string;
	task: string;
	slug: string;
	dependsOn: string[];
	access: "mutating";
	writeScopes: string[];
	requiredSkills: string[];
	verification: ProductionVerificationCommand[];
	humanGates: HumanDecisionGate[];
	maxAttempts: number;
	maxCorrections: number;
}

export interface ProductionParentPlanDocument {
	schemaVersion: 2;
	planId: string;
	parentIssue: number;
	repository: string;
	title: string;
	objective: string;
	parentBranch: string;
	parentBaseBranch: string;
	actorAllowlist: string[];
	decisionExpiresAt: string;
	children: ProductionChildSpec[];
}

export interface ProductionWorkspaceBinding {
	claimId: string;
	ownershipId: string;
	repositoryIdentity: string;
	worktreeIdentity: string;
	cwd: string;
	branch: string;
	baseBranch: string;
	baseHead: string;
	head: string;
	writeScopes: string[];
}

export interface ProductionReviewCheckpoint {
	status: "pending" | "blocked" | "clean";
	baseHead: string;
	head: string;
	resultDigest?: string;
	authorizationDigest?: string;
	completedAt?: string;
	findings: Array<{ id: string; summary: string; disposition?: string }>;
}

export interface ProductionStageCheckpoint {
	summary: string;
	effectKey?: string;
	workspace?: ProductionWorkspaceBinding;
	pullRequest?: number;
	review?: ProductionReviewCheckpoint;
	integrationReceiptDigest?: string;
	parentHead?: string;
}

export type ProductionLifecycleFailureKind =
	| "retryable"
	| "correction_required"
	| "stale_parent"
	| "human_required"
	| "terminal";

export class ProductionLifecycleError extends Error {
	readonly kind: ProductionLifecycleFailureKind;
	readonly blockers: string[];

	constructor(kind: ProductionLifecycleFailureKind, message: string, blockers: readonly string[] = []) {
		super(message);
		this.name = "ProductionLifecycleError";
		this.kind = kind;
		this.blockers = [...blockers];
	}
}

function exact(value: unknown, required: readonly string[], optional: readonly string[] = [], description = "production plan") {
	return readBoundedExactRecord(value, required, optional, description);
}

function text(value: unknown, description: string, maximum = MAX_TEXT_BYTES): string {
	if (typeof value !== "string" || value.length === 0 || Buffer.byteLength(value) > maximum || UNSAFE_TEXT.test(value)) {
		throw new Error(`${description} must be bounded safe text`);
	}
	return value;
}

function positiveInteger(value: unknown, description: string, maximum = Number.MAX_SAFE_INTEGER): number {
	if (!Number.isSafeInteger(value) || (value as number) < 1 || (value as number) > maximum) {
		throw new Error(`${description} must be a bounded positive integer`);
	}
	return value as number;
}

function denseArray(value: unknown, description: string, maximum = MAX_LIST, allowEmpty = true): unknown[] {
	if (!Array.isArray(value) || value.length > maximum || (!allowEmpty && value.length === 0)) {
		throw new Error(`${description} must be a bounded dense array`);
	}
	const result: unknown[] = [];
	for (let index = 0; index < value.length; index += 1) {
		const descriptor = Object.getOwnPropertyDescriptor(value, index);
		if (descriptor === undefined || !Object.hasOwn(descriptor, "value") || descriptor.enumerable !== true) {
			throw new Error(`${description} must contain only dense data values`);
		}
		result.push(descriptor.value);
	}
	return result;
}

function uniqueStrings(value: unknown, description: string, pattern?: RegExp, allowEmpty = true): string[] {
	const values = denseArray(value, description, MAX_LIST, allowEmpty).map((entry) => text(entry, description, 512));
	if (pattern !== undefined && values.some((entry) => !pattern.test(entry))) throw new Error(`invalid ${description}`);
	if (new Set(values).size !== values.length) throw new Error(`duplicate ${description}`);
	return [...values];
}

function validateRelativePath(value: unknown, description: string, allowDot = false): string {
	const path = text(value, description, 4_096);
	if (path.includes("\\") || path.startsWith("/") || path.endsWith("/")
		|| (!allowDot && path === ".") || path.split("/").some((part) => part === "" || part === "..")) {
		throw new Error(`${description} must remain inside the worktree`);
	}
	return path;
}

function validateVerification(value: unknown): ProductionVerificationCommand {
	const candidate = exact(
		value,
		["id", "executable", "args", "cwd", "timeoutMs", "maxOutputBytes"],
		[],
		"verification command",
	);
	const executable = text(candidate.executable, "verification executable", 128);
	if (!SAFE_EXECUTABLE.test(executable) || executable.includes("/")) {
		throw new Error("verification executable must be an allowlistable program name");
	}
	const args = uniqueStrings(candidate.args, "verification argv", undefined, true);
	if (args.some((argument) => argument.length > 4_096 || UNSAFE_TEXT.test(argument))) {
		throw new Error("verification argv contains unsafe text");
	}
	return {
		id: text(candidate.id, "verification ID", 128),
		executable,
		args,
		cwd: validateRelativePath(candidate.cwd, "verification cwd", true),
		timeoutMs: positiveInteger(candidate.timeoutMs, "verification timeout", 120_000),
		maxOutputBytes: positiveInteger(candidate.maxOutputBytes, "verification output limit", 4 * 1024 * 1024),
	};
}

function validateChild(value: unknown): ProductionChildSpec {
	const candidate = exact(value, [
		"id", "issue", "title", "task", "slug", "dependsOn", "access", "writeScopes",
		"requiredSkills", "verification", "humanGates", "maxAttempts", "maxCorrections",
	], [], "production child");
	const id = text(candidate.id, "child ID", 64);
	if (!SAFE_ID.test(id)) throw new Error("invalid child ID");
	if (candidate.access !== "mutating") throw new Error("top-level production children must be mutating");
	const slug = text(candidate.slug, "child slug", 100);
	if (!SAFE_SLUG.test(slug)) throw new Error("invalid child slug");
	const scopes = uniqueStrings(candidate.writeScopes, "child write scopes", undefined, false)
		.map((scope) => validateRelativePath(scope, "child write scope"));
	const humanGates = uniqueStrings(candidate.humanGates, "child human gates") as HumanDecisionGate[];
	const allowedGates: HumanDecisionGate[] = ["requirements", "scope", "review", "head", "merge", "parent_merge"];
	if (humanGates.some((gate) => !allowedGates.includes(gate))) throw new Error("invalid child human gate");
	return {
		id,
		issue: positiveInteger(candidate.issue, "child issue"),
		title: text(candidate.title, "child title", 256),
		task: text(candidate.task, "child task", 4_096),
		slug,
		dependsOn: uniqueStrings(candidate.dependsOn, "child dependencies", SAFE_ID),
		access: "mutating",
		writeScopes: scopes,
		requiredSkills: uniqueStrings(candidate.requiredSkills, "required skills", /^[A-Za-z0-9][A-Za-z0-9:._-]{0,127}$/),
		verification: denseArray(candidate.verification, "verification commands", MAX_LIST, false).map(validateVerification),
		humanGates,
		maxAttempts: positiveInteger(candidate.maxAttempts, "maximum attempts", 10),
		maxCorrections: positiveInteger(candidate.maxCorrections, "maximum corrections", 5),
	};
}

function canonicalPlan(value: ProductionParentPlanDocument): ProductionParentPlanDocument {
	return {
		...value,
		actorAllowlist: [...value.actorAllowlist].sort(),
		children: value.children.map((child) => ({
			...child,
			dependsOn: [...child.dependsOn].sort(),
			writeScopes: [...child.writeScopes].sort(),
			requiredSkills: [...child.requiredSkills].sort(),
			humanGates: [...child.humanGates].sort(),
			verification: child.verification.map((command) => ({ ...command, args: [...command.args] })),
		})).sort((left, right) => left.id.localeCompare(right.id)),
	};
}

export function validateProductionParentPlan(value: unknown, expectedIssue?: number): ProductionParentPlanDocument {
	const candidate = exact(value, [
		"schemaVersion", "planId", "parentIssue", "repository", "title", "objective", "parentBranch",
		"parentBaseBranch", "actorAllowlist", "decisionExpiresAt", "children",
	]);
	if (candidate.schemaVersion !== 2) throw new Error("unsupported production plan schema");
	const parentIssue = positiveInteger(candidate.parentIssue, "parent issue");
	if (expectedIssue !== undefined && parentIssue !== expectedIssue) throw new Error("production plan parent issue mismatch");
	const repository = text(candidate.repository, "repository", 201);
	if (!SAFE_REPOSITORY.test(repository)) throw new Error("repository must be owner/name");
	const parentBranch = text(candidate.parentBranch, "parent branch", 240);
	const parentBaseBranch = text(candidate.parentBaseBranch, "parent base branch", 240);
	if (!SAFE_REF.test(parentBranch) || !SAFE_REF.test(parentBaseBranch) || parentBranch === parentBaseBranch) {
		throw new Error("invalid production parent branches");
	}
	const expiry = text(candidate.decisionExpiresAt, "decision expiry", 64);
	if (!Number.isFinite(new Date(expiry).valueOf())) throw new Error("decision expiry must be an RFC3339 timestamp");
	const children = denseArray(candidate.children, "production children", MAX_CHILDREN, false).map(validateChild);
	if (new Set(children.map((child) => child.id)).size !== children.length) throw new Error("duplicate child ID");
	if (new Set(children.map((child) => child.issue)).size !== children.length) throw new Error("duplicate child issue");
	const ids = new Set(children.map((child) => child.id));
	for (const child of children) {
		if (child.dependsOn.includes(child.id) || child.dependsOn.some((dependency) => !ids.has(dependency))) {
			throw new Error(`invalid dependencies for ${child.id}`);
		}
	}
	return canonicalPlan({
		schemaVersion: 2,
		planId: text(candidate.planId, "plan ID", 256),
		parentIssue,
		repository,
		title: text(candidate.title, "parent title", 256),
		objective: text(candidate.objective, "parent objective", 4_096),
		parentBranch,
		parentBaseBranch,
		actorAllowlist: uniqueStrings(candidate.actorAllowlist, "actor allowlist", /^[a-z\d](?:[a-z\d-]{0,37}[a-z\d])?$/, false),
		decisionExpiresAt: new Date(expiry).toISOString(),
		children,
	});
}

export function productionPlanDigest(value: ProductionParentPlanDocument): string {
	const canonical = validateProductionParentPlan(value, value.parentIssue);
	return createHash("sha256").update(JSON.stringify(canonical)).digest("hex");
}
