import { createHash } from "node:crypto";
import { TextDecoder, types as nodeTypes } from "node:util";

const MAX_ARRAY_ITEMS = 64;
const MAX_TEXT_BYTES = 512;
const MAX_PATH_BYTES = 4_096;
const MAX_RAW_JSON_BYTES = 1_048_576;
const SHA = /^[0-9a-f]{40}$/;
const REPOSITORY = /^(?:[a-z0-9.-]+(?::[0-9]{1,5})?\/)?[A-Za-z0-9][A-Za-z0-9._-]{0,99}\/[A-Za-z0-9][A-Za-z0-9._-]{0,99}$/;
const RFC3339_UTC = /^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(?:\.(\d{1,3}))?Z$/;
const UNSAFE_TEXT = /[\u0000-\u001f\u007f-\u009f\u00ad\u061c\u200b-\u200f\u2028-\u202e\u2060-\u206f\ufeff]/u;

export interface IndependentReviewTarget {
	repository: string;
	workItemId: string;
	pullRequest: number;
	generation: number;
	baseBranch: string;
	headBranch: string;
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
	baseBranch: string;
	headBranch: string;
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
	attestations?: readonly unknown[];
}

export interface AgentSessionAttestation {
	schemaVersion: 1;
	authority: "controller";
	sessionId: string;
	runId: string;
	provider: "openai-codex";
	model: "gpt-5.6-sol";
	reasoningEffort: "xhigh";
	readOnly: true;
	repository: string;
	workItemId: string;
	pullRequest: number;
	generation: number;
	baseBranch: string;
	headBranch: string;
	baseSha: string;
	headSha: string;
	changedPaths: string[];
	allowedScopes: string[];
	reviewMarker: string;
	resultDigest: string;
	completedAt: string;
}

export interface CreateAgentSessionAttestationInput {
	sessionId: string;
	runId: string;
	review: IndependentReviewRecord;
}

type ExactRecord = Record<string, unknown>;

const typedArrayByteLengthGetter = Object.getOwnPropertyDescriptor(
	Object.getPrototypeOf(Uint8Array.prototype) as object,
	"byteLength",
)?.get;

export function decodeBoundedJsonPayload(value: string | Uint8Array, maximumBytes = MAX_RAW_JSON_BYTES): unknown {
	if (!Number.isSafeInteger(maximumBytes) || maximumBytes < 1 || maximumBytes > MAX_RAW_JSON_BYTES) {
		throw new Error("invalid bounded JSON payload byte limit");
	}
	if (nodeTypes.isProxy(value)) throw new Error("invalid bounded JSON payload shape");
	let serialized: string;
	if (typeof value === "string") {
		if (Buffer.byteLength(value) > maximumBytes) throw new Error("bounded JSON payload is oversized");
		serialized = value;
	} else if (value instanceof Uint8Array && Object.getPrototypeOf(value) === Uint8Array.prototype) {
		if (typedArrayByteLengthGetter === undefined) throw new Error("bounded JSON payload byte length is unavailable");
		let byteLength: number;
		try {
			byteLength = typedArrayByteLengthGetter.call(value) as number;
		} catch {
			throw new Error("invalid bounded JSON payload shape");
		}
		if (byteLength > maximumBytes) throw new Error("bounded JSON payload is oversized");
		try {
			serialized = new TextDecoder("utf-8", { fatal: true }).decode(value);
		} catch {
			throw new Error("bounded JSON payload is not valid UTF-8");
		}
		if (Buffer.byteLength(serialized) > maximumBytes) throw new Error("bounded JSON payload is oversized");
	} else {
		throw new Error("bounded JSON payload must be serialized UTF-8 JSON");
	}
	try {
		return JSON.parse(serialized) as unknown;
	} catch {
		throw new Error("bounded JSON payload is invalid");
	}
}

export function readBoundedExactRecord(
	input: unknown,
	required: readonly string[],
	optional: readonly string[] = [],
	description = "record",
): ExactRecord {
	const serializedBytes = typeof input === "object" && input !== null
		&& !nodeTypes.isProxy(input) && input instanceof Uint8Array;
	const value = typeof input === "string" || serializedBytes
		? decodeBoundedJsonPayload(input as string | Uint8Array)
		: input;
	if (typeof value !== "object" || value === null || nodeTypes.isProxy(value) || Array.isArray(value)) {
		throw new Error(`invalid ${description} shape`);
	}
	const prototype = Object.getPrototypeOf(value);
	if (prototype !== Object.prototype && prototype !== null) throw new Error(`invalid ${description} shape`);
	const allowed = new Set([...required, ...optional]);
	if (allowed.size !== required.length + optional.length) throw new Error(`invalid ${description} schema`);
	const enumerable = new Set<string>();
	for (const key in value) {
		if (!Object.hasOwn(value, key)) continue;
		if (enumerable.size >= allowed.size) throw new Error(`${description} bounded envelope is oversized`);
		if (!allowed.has(key)) throw new Error(`unknown ${description} field`);
		enumerable.add(key);
	}
	const result: ExactRecord = {};
	for (const key of allowed) {
		const descriptor = Object.getOwnPropertyDescriptor(value, key);
		if (descriptor === undefined) continue;
		if (!Object.hasOwn(descriptor, "value") || descriptor.enumerable !== true || !enumerable.has(key)) {
			throw new Error(`invalid ${description} shape`);
		}
		result[key] = descriptor.value;
	}
	for (const key of required) {
		if (!Object.hasOwn(result, key)) throw new Error(`invalid ${description} shape`);
	}
	return result;
}

function exactRecord(value: unknown, required: readonly string[], optional: readonly string[] = []): ExactRecord {
	return readBoundedExactRecord(value, required, optional, "independent review record");
}

const CREDENTIAL_ASSIGNMENT_SUFFIXES = [
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

const PUBLIC_CREDENTIAL_SHAPED_ASSIGNMENT_NAMES: ReadonlySet<string> = new Set(["FEATURE_TOKEN"]);

function redactCredentialAssignment(match: string, prefix: string, name: string, separator: string): string {
	const classified = CREDENTIAL_ASSIGNMENT_SUFFIXES.some((suffix) => name === suffix || name.endsWith(`_${suffix}`));
	if (!classified || PUBLIC_CREDENTIAL_SHAPED_ASSIGNMENT_NAMES.has(name)) return match;
	return `${prefix}${name}${separator}[REDACTED]`;
}

export function redactSensitiveText(input: string): string {
	const secretName = "(?:authorization|token|access[_-]?token|refresh[_-]?token|api[_-]?key|password|secret|client[_-]?secret|private[_-]?key|database[_-]?url|credentials?|cookies?|set[_-]?cookie|session(?:[_ -]?(?:id|token|cookie))?|csrf[_-]?token)";
	const credentialEnvironmentName = "(?:OPENAI_API_KEY|ANTHROPIC_API_KEY|GOOGLE_API_KEY|GITHUB_TOKEN|GH_TOKEN|NPM_TOKEN|AWS_ACCESS_KEY_ID|AWS_SECRET_ACCESS_KEY|AWS_SESSION_TOKEN|AZURE_CLIENT_SECRET|GOOGLE_APPLICATION_CREDENTIALS|DATABASE_URL)";
	const credentialAssignment = /(^|[^A-Za-z0-9_])([A-Z][A-Z0-9_]{0,127})([ \t]*=[ \t]*)(?:"[^"\r\n]*"|'[^'\r\n]*'|[^\s,;]+)/gmu;
	return input
		.replace(/-----BEGIN [^-\r\n]*PRIVATE KEY-----[\s\S]*?(?:-----END [^-\r\n]*PRIVATE KEY-----|$)/giu, "[REDACTED]")
		.replace(/\b((?:Set-Cookie|Cookie|X-(?:Session|Auth|CSRF)(?:-Id|-Token)?))\s*:\s*[^\r\n]+/giu, "$1: [REDACTED]")
		.replace(/\bAuthorization\s*:\s*[^\r\n,;]+/giu, "Authorization: [REDACTED]")
		.replace(/\b(?:Bearer|Basic|Token)\s+[^\s,;]+/giu, "[REDACTED]")
		.replace(new RegExp(`(["']${secretName}["']\\s*:\\s*)(?:"[^"\\r\\n]*"|'[^'\\r\\n]*'|[^,}\\r\\n]+)`, "giu"), "$1\"[REDACTED]\"")
		.replace(/(?:^|[\r\n])([ \t]*(?:(?:machine|default)[ \t]+[^\r\n]+?[ \t]+)?password[ \t]+)(?:"[^"\r\n]*"|'[^'\r\n]*'|[^\s,;]+)/giu, "$1[REDACTED]")
		.replace(/((?:^|[^A-Za-z0-9])(?:\/\/[^\s:]+\/)?_(?:authToken|auth|password)\s*=\s*)(?:"[^"\r\n]*"|'[^'\r\n]*'|[^\s,;]+)/giu, "$1[REDACTED]")
		.replace(credentialAssignment, redactCredentialAssignment)
		.replace(new RegExp(`\\b${credentialEnvironmentName}\\s*=\\s*(?:"[^"]*"|'[^']*'|[^\\s,;]+)`, "giu"), "SECRET=[REDACTED]")
		.replace(/\b((?:credentials?|credential)[_-](?:file|path)|google_application_credentials)\s*=\s*(?:"[^"]*"|'[^']*'|[^\s,;]+)/giu, "$1=[REDACTED]")
		.replace(/\b((?:client-key-data|token)\s*:\s*)(?:"[^"\r\n]*"|'[^'\r\n]*'|[^\s,;}]+)/giu, "$1[REDACTED]")
		.replace(/(["'](?:auth|identitytoken)["']\s*:\s*)(?:"[^"\r\n]*"|'[^'\r\n]*'|[^,}\r\n]+)/giu, "$1\"[REDACTED]\"")
		.replace(/\b((?:aws_access_key_id|aws_secret_access_key|aws_session_token)\s*=\s*)(?:"[^"]*"|'[^']*'|[^\s,;]+)/giu, "$1[REDACTED]")
		.replace(new RegExp(`\\b(${secretName})\\b\\s*[:=]\\s*(?:"[^"]*"|'[^']*'|[^\\s,;]+)`, "giu"), "$1=[REDACTED]")
		.replace(/\b([a-z][a-z0-9+.-]*:\/\/)[^\s\/@:]+:[^\s\/@]+@/giu, "$1[REDACTED]@")
		.replace(/([?&](?:token|access[_-]?token|refresh[_-]?token|api[_-]?key|password|secret)=)[^&#\s]+/giu, "$1[REDACTED]")
		.replace(/\b(?:gh[pousr]_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,}|sk_(?:live|test)_[A-Za-z0-9_-]{10,}|sk-[A-Za-z0-9_-]{20,}|(?:AKIA|ASIA)[0-9A-Z]{16})\b/gu, "[REDACTED]");
}

export function assertNoSensitiveText(value: string, description = "text"): void {
	if (redactSensitiveText(value) !== value) throw new Error(`${description} appears to contain a credential or secret`);
}

function safeText(value: unknown, description: string, maximum = MAX_TEXT_BYTES): string {
	if (typeof value !== "string" || value.length === 0 || value.length > maximum || Buffer.byteLength(value) > maximum
		|| value.trim() !== value || UNSAFE_TEXT.test(value)) {
		throw new Error(`invalid ${description}`);
	}
	assertNoSensitiveText(value, description);
	return value;
}

function positiveNumber(value: unknown, description: string): number {
	if (!Number.isSafeInteger(value) || (value as number) < 1 || (value as number) > 2_147_483_647) {
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
	const match = RFC3339_UTC.exec(value);
	if (match === null) throw new Error(`invalid ${description}`);
	const canonical = `${match[1]}.${(match[2] ?? "").padEnd(3, "0")}Z`;
	const parsed = new Date(canonical);
	if (!Number.isFinite(parsed.valueOf()) || parsed.toISOString() !== canonical) throw new Error(`invalid ${description}`);
	return canonical;
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

const PSEUDO_REFS = new Set([
	"HEAD", "FETCH_HEAD", "ORIG_HEAD", "MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD",
	"REBASE_HEAD", "BISECT_HEAD", "AUTO_MERGE",
]);

export function canonicalGitRef(value: unknown, description = "Git ref"): string {
	const result = safeText(value, description, 240);
	if (result.startsWith("refs/") || result.startsWith("-") || result.startsWith("/")
		|| result.endsWith("/") || result.includes("\\") || result.includes("..") || result.includes("//")
		|| result.includes("@{") || result === "@" || /[ ~^:?*\[\]{}]/u.test(result)
		|| result.split("/").some((segment) => segment === "" || segment === "." || segment === ".."
			|| segment.startsWith(".") || segment.endsWith(".") || segment.endsWith(".lock") || PSEUDO_REFS.has(segment))) {
		throw new Error(`invalid ${description}`);
	}
	return result;
}

function exactArrayValues(value: unknown, description: string, allowEmpty: boolean, maximum: number): unknown[] {
	if (nodeTypes.isProxy(value) || !Array.isArray(value) || Object.getPrototypeOf(value) !== Array.prototype) {
		throw new Error(`${description} must be a canonical array`);
	}
	const lengthDescriptor = Object.getOwnPropertyDescriptor(value, "length");
	if (lengthDescriptor === undefined || !Object.hasOwn(lengthDescriptor, "value")
		|| !Number.isSafeInteger(lengthDescriptor.value) || lengthDescriptor.value < 0
		|| (!allowEmpty && lengthDescriptor.value === 0) || lengthDescriptor.value > maximum) {
		throw new Error(`${description} must be a bounded array of at most ${MAX_ARRAY_ITEMS} values`);
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

function stringArray(value: unknown, description: string, pathLike = false, allowEmpty = false): string[] {
	const values = exactArrayValues(value, description, allowEmpty, MAX_ARRAY_ITEMS)
		.map((entry) => pathLike ? pathValue(entry, description) : safeText(entry, description));
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
		"baseBranch",
		"headBranch",
		"baseSha",
		"headSha",
		"changedPaths",
		"allowedScopes",
	]);
	const repository = safeText(candidate.repository, "repository").toLowerCase();
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
		generation: positiveNumber(candidate.generation, "review generation"),
		baseBranch: canonicalGitRef(candidate.baseBranch, "review base branch"),
		headBranch: canonicalGitRef(candidate.headBranch, "review head branch"),
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
		"baseBranch",
		"headBranch",
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
		baseBranch: candidate.baseBranch as string,
		headBranch: candidate.headBranch as string,
		baseSha: candidate.baseSha as string,
		headSha: candidate.headSha as string,
		changedPaths: candidate.changedPaths as string[],
		allowedScopes: candidate.allowedScopes as string[],
	});
	if (candidate.idempotencyMarker !== work.idempotencyMarker) throw new Error("independent review marker mismatch");
	if (candidate.verdict !== "clean" && candidate.verdict !== "findings") {
		throw new Error("invalid independent review verdict");
	}
	const findings = exactArrayValues(candidate.findings, "independent review findings", true, MAX_ARRAY_ITEMS)
		.map(validateFinding);
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

export function independentReviewResultDigest(value: IndependentReviewRecord): string {
	const review = validateIndependentReviewRecord(value);
	return createHash("sha256").update(JSON.stringify({
		idempotencyMarker: review.idempotencyMarker,
		repository: review.repository,
		workItemId: review.workItemId,
		pullRequest: review.pullRequest,
		generation: review.generation,
		baseBranch: review.baseBranch,
		headBranch: review.headBranch,
		baseSha: review.baseSha,
		headSha: review.headSha,
		changedPaths: review.changedPaths,
		allowedScopes: review.allowedScopes,
		completedAt: review.completedAt,
		verdict: review.verdict,
		findings: review.findings,
	})).digest("hex");
}

export function independentReviewAuthorizationDigest(value: IndependentReviewRecord): string {
	const review = validateIndependentReviewRecord(value);
	if (review.verdict !== "clean") throw new Error("independent review authorization requires a clean verdict");
	return createHash("sha256").update(JSON.stringify({
		idempotencyMarker: review.idempotencyMarker,
		repository: review.repository,
		workItemId: review.workItemId,
		pullRequest: review.pullRequest,
		generation: review.generation,
		baseBranch: review.baseBranch,
		headBranch: review.headBranch,
		baseSha: review.baseSha,
		headSha: review.headSha,
		changedPaths: review.changedPaths,
		allowedScopes: review.allowedScopes,
		verdict: review.verdict,
	})).digest("hex");
}

export function validateAgentSessionAttestation(
	value: unknown,
	reviewValue?: IndependentReviewRecord,
): AgentSessionAttestation {
	const candidate = exactRecord(value, [
		"schemaVersion", "authority", "sessionId", "runId", "provider", "model", "reasoningEffort",
		"readOnly", "repository", "workItemId", "pullRequest", "generation", "baseBranch", "headBranch", "baseSha", "headSha",
		"changedPaths", "allowedScopes", "reviewMarker", "resultDigest", "completedAt",
	]);
	if (candidate.schemaVersion !== 1 || candidate.authority !== "controller"
		|| candidate.provider !== "openai-codex" || candidate.model !== "gpt-5.6-sol"
		|| candidate.reasoningEffort !== "xhigh" || candidate.readOnly !== true) {
		throw new Error("invalid controller AgentSession attestation route");
	}
	const target = normalizeTarget({
		repository: candidate.repository,
		workItemId: candidate.workItemId,
		pullRequest: candidate.pullRequest,
		generation: candidate.generation,
		baseBranch: candidate.baseBranch,
		headBranch: candidate.headBranch,
		baseSha: candidate.baseSha,
		headSha: candidate.headSha,
		changedPaths: candidate.changedPaths,
		allowedScopes: candidate.allowedScopes,
	});
	if (typeof candidate.resultDigest !== "string" || !/^[0-9a-f]{64}$/u.test(candidate.resultDigest)) {
		throw new Error("invalid AgentSession result digest");
	}
	const attestation: AgentSessionAttestation = {
		schemaVersion: 1,
		authority: "controller",
		sessionId: safeText(candidate.sessionId, "AgentSession ID", 256),
		runId: safeText(candidate.runId, "AgentSession run ID", 256),
		provider: "openai-codex",
		model: "gpt-5.6-sol",
		reasoningEffort: "xhigh",
		readOnly: true,
		...target,
		changedPaths: [...target.changedPaths],
		allowedScopes: [...target.allowedScopes],
		reviewMarker: safeText(candidate.reviewMarker, "review marker", 512),
		resultDigest: candidate.resultDigest,
		completedAt: canonicalTimestamp(candidate.completedAt, "AgentSession completion timestamp"),
	};
	if (reviewValue !== undefined && !attestsReview(attestation, validateIndependentReviewRecord(reviewValue))) {
		throw new Error("AgentSession attestation does not bind the review result digest and target");
	}
	return attestation;
}

export function createAgentSessionAttestation(value: CreateAgentSessionAttestationInput): AgentSessionAttestation {
	const candidate = exactRecord(value, ["sessionId", "runId", "review"]);
	const review = validateIndependentReviewRecord(candidate.review);
	return validateAgentSessionAttestation({
		schemaVersion: 1,
		authority: "controller",
		sessionId: safeText(candidate.sessionId, "AgentSession ID", 256),
		runId: safeText(candidate.runId, "AgentSession run ID", 256),
		provider: "openai-codex",
		model: "gpt-5.6-sol",
		reasoningEffort: "xhigh",
		readOnly: true,
		repository: review.repository,
		workItemId: review.workItemId,
		pullRequest: review.pullRequest,
		generation: review.generation,
		baseBranch: review.baseBranch,
		headBranch: review.headBranch,
		baseSha: review.baseSha,
		headSha: review.headSha,
		changedPaths: review.changedPaths,
		allowedScopes: review.allowedScopes,
		reviewMarker: review.idempotencyMarker,
		resultDigest: independentReviewResultDigest(review),
		completedAt: review.completedAt,
	}, review);
}

function sameStrings(left: readonly string[], right: readonly string[]): boolean {
	return left.length === right.length && left.every((entry, index) => entry === right[index]);
}

function attestsReview(attestation: AgentSessionAttestation, review: IndependentReviewRecord): boolean {
	return attestation.repository === review.repository
		&& attestation.workItemId === review.workItemId
		&& attestation.pullRequest === review.pullRequest
		&& attestation.generation === review.generation
		&& attestation.baseBranch === review.baseBranch
		&& attestation.headBranch === review.headBranch
		&& attestation.baseSha === review.baseSha
		&& attestation.headSha === review.headSha
		&& sameStrings(attestation.changedPaths, review.changedPaths)
		&& sameStrings(attestation.allowedScopes, review.allowedScopes)
		&& attestation.reviewMarker === review.idempotencyMarker
		&& attestation.resultDigest === independentReviewResultDigest(review)
		&& attestation.completedAt === review.completedAt;
}

export function reconcileIndependentReview(request: IndependentReviewReconcileRequest): IndependentReviewDecision {
	const candidate = exactRecord(request, ["target", "reviews"], ["attestations"]);
	const work = createIndependentReviewWork(candidate.target as IndependentReviewTarget);
	const reviews = exactArrayValues(candidate.reviews, "review records", true, MAX_ARRAY_ITEMS)
		.map(validateIndependentReviewRecord);
	const attestations = exactArrayValues(candidate.attestations ?? [], "AgentSession attestations", true, MAX_ARRAY_ITEMS)
		.map((entry) => validateAgentSessionAttestation(entry));
	if (new Set(attestations.map((entry) => JSON.stringify([entry.sessionId, entry.runId]))).size !== attestations.length) {
		throw new Error("duplicate AgentSession attestation");
	}
	for (const attestation of attestations) {
		const markerMatches = reviews.filter((review) => review.idempotencyMarker === attestation.reviewMarker);
		if (markerMatches.length > 0 && !markerMatches.some((review) => attestsReview(attestation, review))) {
			throw new Error("AgentSession attestation does not bind the review result digest and target");
		}
	}
	const attested = reviews.filter((review) => attestations.some((attestation) => attestsReview(attestation, review)));
	const attempts = new Map<string, Set<string>>();
	for (const review of attested) {
		const key = `${review.idempotencyMarker}\u0000${review.completedAt}`;
		const digests = attempts.get(key) ?? new Set<string>();
		digests.add(independentReviewResultDigest(review));
		attempts.set(key, digests);
	}
	if ([...attempts.values()].some((digests) => digests.size > 1)) {
		throw new Error("ambiguous same-marker review attempts at one completion boundary");
	}
	const exact = attested.filter((review) => review.baseSha === work.baseSha
		&& review.headSha === work.headSha
		&& review.pullRequest === work.pullRequest
		&& review.generation === work.generation
		&& review.repository === work.repository
		&& review.workItemId === work.workItemId
		&& review.baseBranch === work.baseBranch
		&& review.headBranch === work.headBranch
		&& review.idempotencyMarker === work.idempotencyMarker
		&& sameStrings(review.changedPaths, work.changedPaths)
		&& sameStrings(review.allowedScopes, work.allowedScopes))
		.sort((left, right) => right.completedAt.localeCompare(left.completedAt)
			|| independentReviewResultDigest(left).localeCompare(independentReviewResultDigest(right)));
	return exact.length > 0 && exact[0].verdict === "clean"
		? { kind: "satisfied", review: exact[0] }
		: { kind: "dispatch", work };
}
