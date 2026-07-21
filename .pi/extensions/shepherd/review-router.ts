import { createHash } from "node:crypto";
import { types as nodeTypes } from "node:util";

const MAX_ARRAY_ITEMS = 64;
const MAX_TEXT_BYTES = 512;
const MAX_PATH_BYTES = 4_096;
const SHA = /^[0-9a-f]{40}$/;
const REPOSITORY = /^(?:[a-z0-9.-]+(?::[0-9]{1,5})?\/)?[A-Za-z0-9][A-Za-z0-9._-]{0,99}\/[A-Za-z0-9][A-Za-z0-9._-]{0,99}$/;
const UNSAFE_TEXT = /[\u0000-\u001f\u007f-\u009f\u00ad\u061c\u200b-\u200f\u2028-\u202e\u2060-\u206f\ufeff]/u;

export interface IndependentReviewTarget {
	repository: string;
	workItemId: string;
	pullRequest: number;
	generation: number;
	baseSha: string;
	headSha: string;
	changedPaths: readonly string[];
	allowedScopes: readonly string[];
}

export interface IndependentReviewWork {
	schemaVersion: 1;
	idempotencyMarker: string;
	kind: "codex_independent";
	provider: "openai-codex";
	model: "gpt-5.6-sol";
	reasoningEffort: "xhigh";
	readOnly: true;
	repository: string;
	workItemId: string;
	pullRequest: number;
	generation: number;
	baseSha: string;
	headSha: string;
	changedPaths: string[];
	allowedScopes: string[];
}

export interface IndependentReviewFinding {
	id: string;
	severity: "blocking" | "warning";
	summary: string;
	threadId?: string;
}

export interface IndependentReviewRecord extends IndependentReviewWork {
	completedAt: string;
	verdict: "clean" | "findings";
	findings: IndependentReviewFinding[];
}

export type IndependentReviewDecision =
	| { kind: "satisfied"; review: IndependentReviewRecord }
	| { kind: "dispatch"; work: IndependentReviewWork };

export interface IndependentReviewReconcileRequest {
	target: IndependentReviewTarget;
	reviews: readonly unknown[];
}

type ExactRecord = Record<string, unknown>;

function exactRecord(value: unknown, required: readonly string[], optional: readonly string[] = []): ExactRecord {
	if (typeof value !== "object" || value === null || Array.isArray(value) || nodeTypes.isProxy(value)) {
		throw new Error("invalid independent review record shape");
	}
	const prototype = Object.getPrototypeOf(value);
	if (prototype !== Object.prototype && prototype !== null) throw new Error("invalid independent review record shape");
	const descriptors = Object.getOwnPropertyDescriptors(value);
	const allowed = new Set([...required, ...optional]);
	for (const key of required) {
		const descriptor = descriptors[key];
		if (descriptor === undefined || !Object.hasOwn(descriptor, "value") || descriptor.enumerable !== true) {
			throw new Error("invalid independent review record shape");
		}
	}
	for (const key of Reflect.ownKeys(descriptors)) {
		if (typeof key !== "string" || !allowed.has(key)) throw new Error("unknown independent review record field");
		const descriptor = descriptors[key];
		if (!Object.hasOwn(descriptor, "value") || descriptor.enumerable !== true) {
			throw new Error("invalid independent review record shape");
		}
	}
	return Object.fromEntries(Object.entries(descriptors).map(([key, descriptor]) => [key, descriptor.value]));
}

function safeText(value: unknown, description: string, maximum = MAX_TEXT_BYTES): string {
	if (typeof value !== "string" || value.length === 0 || Buffer.byteLength(value) > maximum
		|| value.trim() !== value || UNSAFE_TEXT.test(value)) {
		throw new Error(`invalid ${description}`);
	}
	return value;
}

function positiveNumber(value: unknown, description: string): number {
	if (!Number.isSafeInteger(value) || (value as number) < 1 || (value as number) > 2_147_483_647) {
		throw new Error(`invalid ${description}`);
	}
	return value as number;
}

function nonNegativeNumber(value: unknown, description: string): number {
	if (!Number.isSafeInteger(value) || (value as number) < 0 || (value as number) > 2_147_483_647) {
		throw new Error(`invalid ${description}`);
	}
	return value as number;
}

function sha(value: unknown, description: string): string {
	if (typeof value !== "string" || !SHA.test(value)) throw new Error(`invalid ${description}`);
	return value;
}

function canonicalTimestamp(value: unknown, description: string): string {
	if (typeof value !== "string" || value.length > 64) throw new Error(`invalid ${description}`);
	const parsed = new Date(value);
	if (!Number.isFinite(parsed.valueOf()) || parsed.toISOString() !== value) throw new Error(`invalid ${description}`);
	return value;
}

function pathValue(value: unknown, description: string): string {
	const path = safeText(value, description, MAX_PATH_BYTES).normalize("NFC");
	if (path === "." || path.startsWith("/") || path.endsWith("/") || path.includes("\\")
		|| /[*?\[\]{}]/u.test(path)
		|| path.split("/").some((segment) => segment === "" || segment === "." || segment === ".." || /^[A-Za-z]:$/u.test(segment))) {
		throw new Error(`invalid ${description} path or scope`);
	}
	return path;
}

function stringArray(value: unknown, description: string, pathLike = false, allowEmpty = false): string[] {
	if (!Array.isArray(value) || (!allowEmpty && value.length === 0) || value.length > MAX_ARRAY_ITEMS) {
		throw new Error(`${description} must be a bounded array of at most ${MAX_ARRAY_ITEMS} values`);
	}
	const values = value.map((entry) => pathLike ? pathValue(entry, description) : safeText(entry, description));
	if (new Set(values).size !== values.length) throw new Error(`duplicate ${description}`);
	return [...values].sort();
}

function pathWithinScope(path: string, scope: string): boolean {
	return path === scope || path.startsWith(`${scope}/`);
}

function normalizeTarget(value: unknown): IndependentReviewTarget {
	const candidate = exactRecord(value, [
		"repository",
		"workItemId",
		"pullRequest",
		"generation",
		"baseSha",
		"headSha",
		"changedPaths",
		"allowedScopes",
	]);
	const repository = safeText(candidate.repository, "repository");
	if (!REPOSITORY.test(repository)) throw new Error("invalid review repository");
	const changedPaths = stringArray(candidate.changedPaths, "changed paths", true, true);
	const allowedScopes = stringArray(candidate.allowedScopes, "allowed scopes", true);
	if (changedPaths.some((path) => !allowedScopes.some((scope) => pathWithinScope(path, scope)))) {
		throw new Error("review changed path is outside its allowed scope");
	}
	return {
		repository,
		workItemId: safeText(candidate.workItemId, "work item ID"),
		pullRequest: positiveNumber(candidate.pullRequest, "pull request"),
		generation: nonNegativeNumber(candidate.generation, "review generation"),
		baseSha: sha(candidate.baseSha, "review base SHA"),
		headSha: sha(candidate.headSha, "review head SHA"),
		changedPaths,
		allowedScopes,
	};
}

function markerFor(target: IndependentReviewTarget): string {
	const digest = createHash("sha256").update(JSON.stringify(target)).digest("hex").slice(0, 24);
	return `<!-- shepherd-review:v1:${target.pullRequest}:${target.generation}:${digest} -->`;
}

export function createIndependentReviewWork(value: IndependentReviewTarget): IndependentReviewWork {
	const target = normalizeTarget(value);
	return {
		schemaVersion: 1,
		idempotencyMarker: markerFor(target),
		kind: "codex_independent",
		provider: "openai-codex",
		model: "gpt-5.6-sol",
		reasoningEffort: "xhigh",
		readOnly: true,
		...target,
		changedPaths: [...target.changedPaths],
		allowedScopes: [...target.allowedScopes],
	};
}

function validateFinding(value: unknown): IndependentReviewFinding {
	const candidate = exactRecord(value, ["id", "severity", "summary"], ["threadId"]);
	if (candidate.severity !== "blocking" && candidate.severity !== "warning") {
		throw new Error("invalid independent review finding severity");
	}
	return {
		id: safeText(candidate.id, "review finding ID"),
		severity: candidate.severity,
		summary: safeText(candidate.summary, "review finding summary", 2_048),
		...(candidate.threadId !== undefined ? { threadId: safeText(candidate.threadId, "review thread ID") } : {}),
	};
}

export function validateIndependentReviewRecord(value: unknown): IndependentReviewRecord {
	const candidate = exactRecord(value, [
		"schemaVersion",
		"idempotencyMarker",
		"kind",
		"provider",
		"model",
		"reasoningEffort",
		"readOnly",
		"repository",
		"workItemId",
		"pullRequest",
		"generation",
		"baseSha",
		"headSha",
		"changedPaths",
		"allowedScopes",
		"completedAt",
		"verdict",
		"findings",
	]);
	if (candidate.schemaVersion !== 1
		|| candidate.kind !== "codex_independent"
		|| candidate.provider !== "openai-codex"
		|| candidate.model !== "gpt-5.6-sol"
		|| candidate.reasoningEffort !== "xhigh"
		|| candidate.readOnly !== true) {
		throw new Error("ineligible independent review route; exact model, xhigh effort, and read-only mode are required");
	}
	const work = createIndependentReviewWork({
		repository: candidate.repository as string,
		workItemId: candidate.workItemId as string,
		pullRequest: candidate.pullRequest as number,
		generation: candidate.generation as number,
		baseSha: candidate.baseSha as string,
		headSha: candidate.headSha as string,
		changedPaths: candidate.changedPaths as string[],
		allowedScopes: candidate.allowedScopes as string[],
	});
	if (candidate.idempotencyMarker !== work.idempotencyMarker) throw new Error("independent review marker mismatch");
	if (candidate.verdict !== "clean" && candidate.verdict !== "findings") {
		throw new Error("invalid independent review verdict");
	}
	if (!Array.isArray(candidate.findings) || candidate.findings.length > MAX_ARRAY_ITEMS) {
		throw new Error("independent review findings must be bounded");
	}
	const findings = candidate.findings.map(validateFinding);
	if (new Set(findings.map((finding) => finding.id)).size !== findings.length) {
		throw new Error("duplicate independent review finding ID");
	}
	if ((candidate.verdict === "clean") !== (findings.length === 0)) {
		throw new Error("clean review must have no findings and findings verdict must contain findings");
	}
	return {
		...work,
		completedAt: canonicalTimestamp(candidate.completedAt, "review completion timestamp"),
		verdict: candidate.verdict,
		findings,
	};
}

export function reviewCoversExactRange(review: IndependentReviewRecord, base: string, head: string): boolean {
	try {
		const validated = validateIndependentReviewRecord(review);
		return validated.verdict === "clean"
			&& validated.baseSha === sha(base, "expected review base SHA")
			&& validated.headSha === sha(head, "expected review head SHA");
	} catch {
		return false;
	}
}

export function reconcileIndependentReview(request: IndependentReviewReconcileRequest): IndependentReviewDecision {
	const work = createIndependentReviewWork(request.target);
	if (!Array.isArray(request.reviews) || request.reviews.length > MAX_ARRAY_ITEMS) {
		throw new Error("review records must be a bounded array");
	}
	const reviews = request.reviews.map(validateIndependentReviewRecord);
	const exact = reviews.filter((review) => reviewCoversExactRange(review, work.baseSha, work.headSha)
		&& review.pullRequest === work.pullRequest
		&& review.generation === work.generation
		&& review.repository === work.repository
		&& review.workItemId === work.workItemId);
	if (exact.length > 1) throw new Error("duplicate exact-head independent review records are ambiguous");
	return exact.length === 1 ? { kind: "satisfied", review: exact[0] } : { kind: "dispatch", work };
}
