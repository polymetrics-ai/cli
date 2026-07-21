import { createHash, randomUUID } from "node:crypto";
import { constants } from "node:fs";
import { link, lstat, mkdir, open, readdir, rename, rm, unlink } from "node:fs/promises";
import { isAbsolute, join } from "node:path";

const SCHEMA_VERSION = 1;
const MAX_GITHUB_NUMBER = Number.MAX_SAFE_INTEGER;
const MAX_STATE_BYTES = 256 * 1024;
const DEFAULT_LOCK_RETRY_MS = 10;
const DEFAULT_LOCK_MAX_ATTEMPTS = 500;
const DEFAULT_LOCK_STALE_MS = 600_000;
const GITHUB_SECOND_RESOLUTION_SKEW_MS = 999;
const REQUEST_ID = /^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$/;
const OPTION = /^[a-z][a-z0-9_-]{0,63}$/;
const LOGIN = /^[a-z\d](?:[a-z\d-]{0,37}[a-z\d])?$/;
const REPOSITORY_PART = /^[A-Za-z\d](?:[A-Za-z\d_.-]{0,98}[A-Za-z\d_.-])?$/;
const HEAD_SHA = /^[0-9a-f]{40}$/;
const RFC3339_UTC = /^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(?:\.(\d{1,3}))?Z$/;
const UNSAFE_UNICODE_FORMAT = /[\u00ad\u061c\u200b-\u200f\u2028-\u202e\u2060-\u206f\ufeff]/u;
const GITHUB_MENTION = /(^|[^A-Za-z0-9_])@[A-Za-z\d](?:[A-Za-z\d-]{0,37}[A-Za-z\d])?/u;

export type HumanDecisionGate = "requirements" | "scope" | "review" | "head" | "merge" | "parent_merge";
export type StandardHumanDecisionGate = Exclude<HumanDecisionGate, "parent_merge">;
export type ParentMergeDecisionOption = "approve-merge" | "reject";
export type HumanDecisionTarget =
	| { kind: "issue"; number: number }
	| { kind: "pull_request"; number: number };

export interface HumanDecisionBinding {
	repository: string;
	target: HumanDecisionTarget;
	generation: number;
	headSha?: string;
}

interface HumanDecisionRequestBase {
	requestId: string;
	binding: HumanDecisionBinding;
	actorAllowlist: string[];
	expiresAt: string;
	question: string;
}

export type HumanDecisionRequestSpec =
	| (HumanDecisionRequestBase & {
		gate: StandardHumanDecisionGate;
		allowedOptions: string[];
	})
	| (HumanDecisionRequestBase & {
		gate: "parent_merge";
		allowedOptions: ParentMergeDecisionOption[];
	});

export interface HumanDecisionRequestComment {
	id: number;
	url: string;
	actor: string;
	createdAt: string;
}

export interface HumanDecisionEvidence {
	option: string;
	actor: string;
	sourceUrl: string;
	decidedAt: string;
}

interface HumanDecisionRecordState {
	schemaVersion: 1;
	idempotencyMarker: string;
	status: "pending" | "decided" | "consumed" | "expired";
	createdAt: string;
	updatedAt: string;
	requestComment?: HumanDecisionRequestComment;
	decision?: HumanDecisionEvidence;
	consumedAt?: string;
}

export type HumanDecisionRecord = HumanDecisionRequestSpec & HumanDecisionRecordState;

export interface HumanDecisionTransaction<T> {
	state: HumanDecisionRecord | null;
	value: T;
}

export interface HumanDecisionRepository {
	load(requestId: string): Promise<HumanDecisionRecord | null>;
	transact<T>(
		requestId: string,
		operation: (state: HumanDecisionRecord | null) => Promise<HumanDecisionTransaction<T>> | HumanDecisionTransaction<T>,
	): Promise<T>;
}

export interface FileHumanDecisionRepositoryOptions {
	lockRetryMs?: number;
	lockMaxAttempts?: number;
	lockStaleMs?: number;
}

interface HumanDecisionLockOwner {
	pid: number;
	token: string;
	createdAt: string;
}

interface HumanDecisionLockLease extends HumanDecisionLockOwner {
	requestId: string;
	path: string;
	state: "candidate" | "active";
}

type HumanDecisionLockScan =
	| { status: "stable"; locks: HumanDecisionLockLease[] }
	| { status: "changed" };

type HumanDecisionLockOwnerInspection =
	| { status: "valid"; owner: HumanDecisionLockOwner }
	| { status: "missing" }
	| { status: "invalid" };

function compareLocks(left: HumanDecisionLockLease, right: HumanDecisionLockLease): number {
	return new Date(left.createdAt).valueOf() - new Date(right.createdAt).valueOf()
		|| left.token.localeCompare(right.token);
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null && !Array.isArray(value);
}

function assertOnlyFields(value: Record<string, unknown>, allowed: readonly string[], description: string): void {
	const allowedFields = new Set(allowed);
	for (const field of Object.keys(value)) {
		if (!allowedFields.has(field)) throw new Error(`invalid human decision ${description} field ${JSON.stringify(field)}`);
	}
}

function validGitHubNumber(value: unknown): value is number {
	return Number.isSafeInteger(value) && (value as number) > 0 && (value as number) <= MAX_GITHUB_NUMBER;
}

function canonicalTimestamp(value: unknown, description: string): string {
	if (typeof value !== "string" || value.length > 128) throw new Error(`invalid human decision ${description}`);
	const match = RFC3339_UTC.exec(value);
	if (!match) throw new Error(`invalid human decision ${description}`);
	const canonical = `${match[1]}.${(match[2] ?? "").padEnd(3, "0")}Z`;
	const parsed = new Date(canonical);
	if (!Number.isFinite(parsed.valueOf()) || parsed.toISOString() !== canonical) {
		throw new Error(`invalid human decision ${description}`);
	}
	return canonical;
}

function safeText(value: unknown, maximum: number, description: string, allowNewlines = false): string {
	if (typeof value !== "string" || value.length === 0 || value.length > maximum) {
		throw new Error(`invalid human decision ${description}`);
	}
	const forbidden = allowNewlines ? /[\u0000-\u0009\u000b\u000c\u000e-\u001f\u007f-\u009f]/ : /[\u0000-\u001f\u007f-\u009f]/;
	if (forbidden.test(value) || UNSAFE_UNICODE_FORMAT.test(value)) throw new Error(`invalid human decision ${description}`);
	return value;
}

function redactSensitiveText(input: string): string {
	const secretName = "(?:authorization|token|access[_-]?token|refresh[_-]?token|api[_-]?key|password|secret|client[_-]?secret|private[_-]?key|database[_-]?url|credentials?)";
	return input
		.replace(/-----BEGIN [^-\r\n]*PRIVATE KEY-----[\s\S]*?(?:-----END [^-\r\n]*PRIVATE KEY-----|$)/gi, "[REDACTED]")
		.replace(/\bAuthorization\s*:\s*[^\r\n,;]+/gi, "Authorization: [REDACTED]")
		.replace(/\b(?:Bearer|Basic|Token)\s+[^\s,;]+/gi, "[REDACTED]")
		.replace(new RegExp(`(["']${secretName}["']\\s*:\\s*)(?:"[^"\\r\\n]*"|'[^'\\r\\n]*'|[^,}\\r\\n]+)`, "gi"), "$1\"[REDACTED]\"")
		.replace(/\b[A-Z][A-Z0-9_]*(?:TOKEN|SECRET|PASSWORD|API_KEY|PRIVATE_KEY|DATABASE_URL|CREDENTIALS?|SECRET_ACCESS_KEY|ACCESS_KEY)\s*=\s*(?:"[^"]*"|'[^']*'|[^\s,;]+)/g, "SECRET=[REDACTED]")
		.replace(new RegExp(`\\b(${secretName})\\b\\s*[:=]\\s*(?:"[^"]*"|'[^']*'|[^\\s,;]+)`, "gi"), "$1=[REDACTED]")
		.replace(/\b([a-z][a-z0-9+.-]*:\/\/)[^\s\/@:]+:[^\s\/@]+@/gi, "$1[REDACTED]@")
		.replace(/([?&](?:token|access[_-]?token|refresh[_-]?token|api[_-]?key|password|secret)=)[^&#\s]+/gi, "$1[REDACTED]")
		.replace(/\b(?:gh[pousr]_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,}|sk_(?:live|test)_[A-Za-z0-9_-]{10,}|sk-[A-Za-z0-9_-]{20,}|AKIA[0-9A-Z]{16})\b/g, "[REDACTED]");
}

function assertNoSecret(value: string): void {
	if (redactSensitiveText(value) !== value) throw new Error("human decision text appears to contain a credential or secret");
}

function normalizeRepository(value: unknown): string {
	const repository = safeText(value, 201, "repository");
	const parts = repository.split("/");
	if (parts.length !== 2
		|| parts.some((part) => !REPOSITORY_PART.test(part) || part.startsWith(".") || part.endsWith(".") || part.includes(".."))) {
		throw new Error("invalid human decision repository");
	}
	return repository.toLowerCase();
}

function normalizeTarget(value: unknown): HumanDecisionTarget {
	if (!isRecord(value) || (value.kind !== "issue" && value.kind !== "pull_request") || !validGitHubNumber(value.number)) {
		throw new Error("invalid human decision target");
	}
	assertOnlyFields(value, ["kind", "number"], "target");
	return { kind: value.kind, number: value.number };
}

function normalizeBinding(value: HumanDecisionBinding): HumanDecisionBinding {
	if (!isRecord(value) || !Number.isSafeInteger(value.generation) || value.generation < 1) {
		throw new Error("invalid human decision generation binding");
	}
	assertOnlyFields(value, ["repository", "target", "generation", "headSha"], "binding");
	const target = normalizeTarget(value.target);
	const headSha = value.headSha;
	if (target.kind === "pull_request") {
		if (typeof headSha !== "string" || !HEAD_SHA.test(headSha)) {
			throw new Error("pull request human decision requires an exact lowercase head SHA");
		}
	} else if (headSha !== undefined) {
		throw new Error("issue human decision must not include a head SHA");
	}
	return {
		repository: normalizeRepository(value.repository),
		target,
		generation: value.generation,
		...(headSha ? { headSha } : {}),
	};
}

export function validateHumanDecisionBinding(value: HumanDecisionBinding): HumanDecisionBinding {
	return normalizeBinding(value);
}

function normalizeOptions(values: unknown): string[] {
	if (!Array.isArray(values) || values.length === 0 || values.length > 16) {
		throw new Error("human decision allowed options must be a bounded non-empty array");
	}
	const normalized = values.map((value) => {
		if (typeof value !== "string" || !OPTION.test(value)) throw new Error("invalid human decision option");
		return value;
	});
	if (new Set(normalized).size !== normalized.length) throw new Error("duplicate human decision option");
	return normalized;
}

function normalizeParentMergeOptions(values: unknown): ParentMergeDecisionOption[] {
	const normalized = normalizeOptions(values);
	if (normalized.length !== 2 || normalized[0] !== "approve-merge" || normalized[1] !== "reject") {
		throw new Error("parent_merge requires exact allowed options approve-merge and reject");
	}
	return ["approve-merge", "reject"];
}

function normalizeActors(values: unknown): string[] {
	if (!Array.isArray(values) || values.length === 0 || values.length > 64) {
		throw new Error("human decision actor allowlist must be a bounded non-empty array");
	}
	const normalized = values.map((value) => {
		if (typeof value !== "string") throw new Error("invalid human decision actor");
		const login = value.toLowerCase();
		if (!LOGIN.test(login)) throw new Error("invalid human decision actor");
		return login;
	});
	if (new Set(normalized).size !== normalized.length) throw new Error("duplicate human decision actor");
	return normalized;
}

function normalizeGitHubActor(value: unknown, description: string): string {
	const actor = safeText(value, 64, description).toLowerCase();
	const base = actor.endsWith("[bot]") ? actor.slice(0, -5) : actor;
	if (!LOGIN.test(base)) throw new Error(`invalid human decision ${description}`);
	return actor;
}

function normalizeQuestion(value: unknown): string {
	const question = safeText(value, 4_096, "question", true).replace(/\r\n?/g, "\n").trim();
	if (question.length === 0 || question.includes("<!-- shepherd-decision:")) {
		throw new Error("invalid human decision question");
	}
	if (GITHUB_MENTION.test(question)) throw new Error("invalid human decision question mention");
	assertNoSecret(question);
	return question;
}

function gateTargetKind(gate: HumanDecisionGate): HumanDecisionTarget["kind"] {
	return gate === "requirements" || gate === "scope" ? "issue" : "pull_request";
}

function normalizeSpec(spec: HumanDecisionRequestSpec, now: Date): HumanDecisionRequestSpec {
	if (!isRecord(spec) || typeof spec.requestId !== "string" || !REQUEST_ID.test(spec.requestId)) {
		throw new Error("invalid human decision request ID");
	}
	if (!["requirements", "scope", "review", "head", "merge", "parent_merge"].includes(spec.gate)) {
		throw new Error("invalid human decision gate");
	}
	const binding = normalizeBinding(spec.binding);
	if (binding.target.kind !== gateTargetKind(spec.gate)) {
		throw new Error(`human decision gate ${spec.gate} is bound to the wrong target kind`);
	}
	const expiresAt = canonicalTimestamp(spec.expiresAt, "expiry");
	if (!Number.isFinite(now.valueOf()) || new Date(expiresAt).valueOf() <= now.valueOf()) {
		throw new Error("human decision request is already expired");
	}
	const allowedOptions = spec.gate === "parent_merge"
		? normalizeParentMergeOptions(spec.allowedOptions)
		: normalizeOptions(spec.allowedOptions);
	return {
		requestId: spec.requestId,
		gate: spec.gate,
		binding,
		allowedOptions,
		actorAllowlist: normalizeActors(spec.actorAllowlist),
		expiresAt,
		question: normalizeQuestion(spec.question),
	} as HumanDecisionRequestSpec;
}

function markerFor(spec: HumanDecisionRequestSpec): string {
	const digest = createHash("sha256").update(JSON.stringify({
		requestId: spec.requestId,
		gate: spec.gate,
		binding: spec.binding,
		allowedOptions: spec.allowedOptions,
		actorAllowlist: spec.actorAllowlist,
		expiresAt: spec.expiresAt,
		question: spec.question,
	})).digest("hex").slice(0, 24);
	return `<!-- shepherd-decision:v1:${spec.requestId}:${digest} -->`;
}

function canonicalNow(now: Date): string {
	if (!Number.isFinite(now.valueOf())) throw new Error("invalid human decision clock");
	return now.toISOString();
}

function sameBinding(left: HumanDecisionBinding, right: HumanDecisionBinding): boolean {
	return left.repository === right.repository
		&& left.target.kind === right.target.kind
		&& left.target.number === right.target.number
		&& left.generation === right.generation
		&& left.headSha === right.headSha;
}

export function assertHumanDecisionBinding(record: HumanDecisionRecord, binding: HumanDecisionBinding): HumanDecisionBinding {
	const normalized = normalizeBinding(binding);
	if (!sameBinding(record.binding, normalized)) {
		throw new Error("human decision binding is stale or targets a different repository, generation, head, or issue/PR");
	}
	return normalized;
}

export function routeHumanDecisionTarget(
	gate: HumanDecisionGate,
	parentIssue: number,
	pullRequest: number,
): HumanDecisionTarget {
	if (!validGitHubNumber(parentIssue) || !validGitHubNumber(pullRequest)) {
		throw new Error("invalid human decision issue or pull request number");
	}
	if (!["requirements", "scope", "review", "head", "merge", "parent_merge"].includes(gate)) {
		throw new Error("invalid human decision gate");
	}
	return gateTargetKind(gate) === "issue"
		? { kind: "issue", number: parentIssue }
		: { kind: "pull_request", number: pullRequest };
}

export function createHumanDecisionRecord(
	spec: HumanDecisionRequestSpec,
	now = new Date(),
): HumanDecisionRecord {
	const normalized = normalizeSpec(spec, now);
	const timestamp = canonicalNow(now);
	return {
		schemaVersion: SCHEMA_VERSION,
		...normalized,
		idempotencyMarker: markerFor(normalized),
		status: "pending",
		createdAt: timestamp,
		updatedAt: timestamp,
	} as HumanDecisionRecord;
}

function immutableProjection(record: HumanDecisionRecord): unknown {
	return {
		requestId: record.requestId,
		gate: record.gate,
		binding: record.binding,
		allowedOptions: record.allowedOptions,
		actorAllowlist: record.actorAllowlist,
		expiresAt: record.expiresAt,
		question: record.question,
		idempotencyMarker: record.idempotencyMarker,
	};
}

export async function persistHumanDecisionRequest(
	repository: HumanDecisionRepository,
	spec: HumanDecisionRequestSpec,
	now = new Date(),
): Promise<HumanDecisionRecord> {
	const proposed = createHumanDecisionRecord(spec, now);
	return repository.transact(proposed.requestId, (existing) => {
		if (existing === null) return { state: proposed, value: proposed };
		if (JSON.stringify(immutableProjection(existing)) !== JSON.stringify(immutableProjection(proposed))) {
			throw new Error("persisted human decision request conflicts with retry specification");
		}
		return { state: existing, value: existing };
	});
}

export function validateHumanDecisionRequestComment(
	record: HumanDecisionRecord,
	evidence: HumanDecisionRequestComment,
): HumanDecisionRequestComment {
	if (!isRecord(evidence) || !validGitHubNumber(evidence.id)) throw new Error("invalid human decision request comment ID");
	assertOnlyFields(evidence, ["id", "url", "actor", "createdAt"], "request comment");
	const actor = normalizeGitHubActor(evidence.actor, "request comment actor");
	const url = validateSourceUrl(record.binding, evidence.url, "request comment URL");
	const createdAt = canonicalTimestamp(evidence.createdAt, "request comment timestamp");
	if (new Date(createdAt).valueOf() + GITHUB_SECOND_RESOLUTION_SKEW_MS < new Date(record.createdAt).valueOf()
		|| new Date(createdAt).valueOf() >= new Date(record.expiresAt).valueOf()) {
		throw new Error("human decision request comment timestamp is outside the request lifetime");
	}
	return { id: evidence.id, url, actor, createdAt };
}

export async function recordHumanDecisionRequestComment(
	repository: HumanDecisionRepository,
	requestId: string,
	binding: HumanDecisionBinding,
	evidence: HumanDecisionRequestComment,
	now = new Date(),
): Promise<HumanDecisionRecord> {
	return repository.transact(requestId, (existing) => {
		if (existing === null) throw new Error("human decision request does not exist");
		assertHumanDecisionBinding(existing, binding);
		const normalized = validateHumanDecisionRequestComment(existing, evidence);
		if (existing.requestComment) {
			if (JSON.stringify(existing.requestComment) !== JSON.stringify(normalized)) {
				throw new Error("human decision request comment conflicts with persisted marker owner");
			}
			return { state: existing, value: existing };
		}
		const updated = { ...existing, requestComment: normalized, updatedAt: canonicalNow(now) };
		return { state: updated, value: updated };
	});
}

function validateSourceUrl(binding: HumanDecisionBinding, value: unknown, description: string): string {
	const text = safeText(value, 2_048, description);
	let url: URL;
	try { url = new URL(text); } catch { throw new Error(`invalid human decision ${description}`); }
	if (url.protocol !== "https:" || url.username || url.password || url.search || url.hostname.toLowerCase() !== "github.com") {
		throw new Error(`invalid human decision ${description}`);
	}
	const [owner, name] = binding.repository.split("/");
	const kind = binding.target.kind === "issue" ? "issues" : "pull";
	const prefix = `/${owner}/${name}/${kind}/${binding.target.number}`.toLowerCase();
	if (url.pathname.toLowerCase() !== prefix || !/^#issuecomment-\d+$/.test(url.hash)) {
		throw new Error(`human decision ${description} is not bound to the exact target`);
	}
	return url.toString();
}

function normalizeDecision(record: HumanDecisionRecord, evidence: HumanDecisionEvidence): HumanDecisionEvidence {
	if (!isRecord(evidence)
		|| typeof evidence.option !== "string"
		|| !(record.allowedOptions as readonly string[]).includes(evidence.option)) {
		throw new Error("human decision option is not allowed");
	}
	assertOnlyFields(evidence, ["option", "actor", "sourceUrl", "decidedAt"], "decision evidence");
	const actor = safeText(evidence.actor, 39, "actor").toLowerCase();
	if (!LOGIN.test(actor) || !record.actorAllowlist.includes(actor)) {
		throw new Error("human decision actor is not allowlisted");
	}
	const decidedAt = canonicalTimestamp(evidence.decidedAt, "decision timestamp");
	const decidedTime = new Date(decidedAt).valueOf();
	if (decidedTime + GITHUB_SECOND_RESOLUTION_SKEW_MS < new Date(record.createdAt).valueOf()
		|| decidedTime >= new Date(record.expiresAt).valueOf()) {
		throw new Error("human decision timestamp is outside the request lifetime");
	}
	return {
		option: evidence.option,
		actor,
		sourceUrl: validateSourceUrl(record.binding, evidence.sourceUrl, "source URL"),
		decidedAt,
	};
}

export async function recordHumanDecision(
	repository: HumanDecisionRepository,
	requestId: string,
	binding: HumanDecisionBinding,
	evidence: HumanDecisionEvidence,
): Promise<HumanDecisionRecord> {
	return repository.transact(requestId, (existing) => {
		if (existing === null) throw new Error("human decision request does not exist");
		assertHumanDecisionBinding(existing, binding);
		if (existing.status !== "pending") throw new Error(`human decision request is already ${existing.status}`);
		const decision = normalizeDecision(existing, evidence);
		const updated = { ...existing, status: "decided" as const, decision, updatedAt: decision.decidedAt };
		return { state: updated, value: updated };
	});
}

export async function expireHumanDecision(
	repository: HumanDecisionRepository,
	requestId: string,
	binding: HumanDecisionBinding,
	now = new Date(),
): Promise<HumanDecisionRecord> {
	return repository.transact(requestId, (existing) => {
		if (existing === null) throw new Error("human decision request does not exist");
		assertHumanDecisionBinding(existing, binding);
		if (existing.status !== "pending") return { state: existing, value: existing };
		if (now.valueOf() < new Date(existing.expiresAt).valueOf()) throw new Error("human decision request is not expired");
		const updated = { ...existing, status: "expired" as const, updatedAt: canonicalNow(now) };
		return { state: updated, value: updated };
	});
}

export async function consumeHumanDecision(
	repository: HumanDecisionRepository,
	requestId: string,
	binding: HumanDecisionBinding,
	now = new Date(),
): Promise<HumanDecisionEvidence> {
	return repository.transact(requestId, (existing) => {
		if (existing === null) throw new Error("human decision request does not exist");
		assertHumanDecisionBinding(existing, binding);
		if (existing.status === "consumed") throw new Error("human decision was already consumed");
		if (existing.status !== "decided" || !existing.decision) throw new Error("human decision is not ready to consume");
		const decision = { ...existing.decision };
		const timestamp = canonicalNow(now);
		const updated = { ...existing, status: "consumed" as const, consumedAt: timestamp, updatedAt: timestamp };
		return { state: updated, value: decision };
	});
}

function validateRecord(value: unknown): HumanDecisionRecord {
	if (!isRecord(value) || value.schemaVersion !== SCHEMA_VERSION) throw new Error("invalid human decision state schema");
	assertOnlyFields(value, [
		"schemaVersion", "requestId", "gate", "binding", "allowedOptions", "actorAllowlist", "expiresAt",
		"question", "idempotencyMarker", "status", "createdAt", "updatedAt", "requestComment", "decision", "consumedAt",
	], "state");
	const createdAt = canonicalTimestamp(value.createdAt, "creation timestamp");
	const proposed = createHumanDecisionRecord({
		requestId: value.requestId as string,
		gate: value.gate as HumanDecisionGate,
		binding: value.binding as HumanDecisionBinding,
		allowedOptions: value.allowedOptions as string[],
		actorAllowlist: value.actorAllowlist as string[],
		expiresAt: value.expiresAt as string,
		question: value.question as string,
	} as HumanDecisionRequestSpec, new Date(new Date(createdAt).valueOf() - 1));
	if (value.idempotencyMarker !== proposed.idempotencyMarker) throw new Error("invalid human decision idempotency marker");
	if (!["pending", "decided", "consumed", "expired"].includes(value.status as string)) throw new Error("invalid human decision status");
	const updatedAt = canonicalTimestamp(value.updatedAt, "update timestamp");
	if (new Date(updatedAt).valueOf() < new Date(createdAt).valueOf()) throw new Error("invalid human decision update chronology");
	const record: HumanDecisionRecord = {
		...proposed,
		status: value.status as HumanDecisionRecord["status"],
		createdAt,
		updatedAt,
	};
	if (value.requestComment !== undefined) record.requestComment = validateHumanDecisionRequestComment(record, value.requestComment as HumanDecisionRequestComment);
	if (value.decision !== undefined) record.decision = normalizeDecision(record, value.decision as HumanDecisionEvidence);
	if (value.consumedAt !== undefined) record.consumedAt = canonicalTimestamp(value.consumedAt, "consumption timestamp");
	if ((record.status === "decided" || record.status === "consumed") !== Boolean(record.decision)) {
		throw new Error("invalid human decision state/decision coherence");
	}
	if ((record.status === "consumed") !== Boolean(record.consumedAt)) {
		throw new Error("invalid human decision consumption coherence");
	}
	if (record.consumedAt && record.decision
		&& new Date(record.consumedAt).valueOf() < new Date(record.decision.decidedAt).valueOf()) {
		throw new Error("invalid human decision consumption chronology");
	}
	if (record.status === "expired" && new Date(record.updatedAt).valueOf() < new Date(record.expiresAt).valueOf()) {
		throw new Error("invalid human decision expiry chronology");
	}
	return record;
}

export function validateHumanDecisionRecord(value: unknown): HumanDecisionRecord {
	return validateRecord(value);
}

function validateRepositoryOption(value: unknown, fallback: number, maximum: number, description: string): number {
	const option = value ?? fallback;
	if (!Number.isSafeInteger(option) || (option as number) < 1 || (option as number) > maximum) {
		throw new Error(`invalid human decision repository ${description}`);
	}
	return option as number;
}

export class FileHumanDecisionRepository implements HumanDecisionRepository {
	private readonly root: string;
	private readonly lockRetryMs: number;
	private readonly lockMaxAttempts: number;
	private readonly lockStaleMs: number;

	constructor(root: string, options: FileHumanDecisionRepositoryOptions = {}) {
		if (typeof root !== "string" || !isAbsolute(root) || root.length > 4_096 || /[\u0000-\u001f\u007f]/.test(root)) {
			throw new Error("human decision repository root must be an absolute safe path");
		}
		this.root = root;
		this.lockRetryMs = validateRepositoryOption(options.lockRetryMs, DEFAULT_LOCK_RETRY_MS, 10_000, "lock retry");
		this.lockMaxAttempts = validateRepositoryOption(options.lockMaxAttempts, DEFAULT_LOCK_MAX_ATTEMPTS, 10_000, "lock attempts");
		this.lockStaleMs = validateRepositoryOption(options.lockStaleMs, DEFAULT_LOCK_STALE_MS, 3_600_000, "stale lock bound");
	}

	private statePath(requestId: string): string {
		if (!REQUEST_ID.test(requestId)) throw new Error("invalid human decision request ID");
		return join(this.root, `${requestId}.json`);
	}

	private async ensureRoot(): Promise<void> {
		await mkdir(this.root, { recursive: true, mode: 0o700 });
		const metadata = await lstat(this.root);
		if (!metadata.isDirectory() || metadata.isSymbolicLink()) throw new Error("human decision repository root is not a trusted directory");
		if (process.platform !== "win32" && (metadata.mode & 0o077) !== 0) {
			throw new Error("human decision repository root must not be accessible by group or other users");
		}
	}

	private async readState(requestId: string): Promise<HumanDecisionRecord | null> {
		await this.ensureRoot();
		let handle: Awaited<ReturnType<typeof open>>;
		try {
			handle = await open(this.statePath(requestId), constants.O_RDONLY | (constants.O_NOFOLLOW ?? 0));
		} catch (error) {
			if (isRecord(error) && error.code === "ENOENT") return null;
			throw error;
		}
		let contents: string;
		try {
			const metadata = await handle.stat();
			if (!metadata.isFile() || metadata.size > MAX_STATE_BYTES) throw new Error("human decision state is not a bounded regular file");
			contents = await handle.readFile("utf8");
		} finally {
			await handle.close();
		}
		let parsed: unknown;
		try { parsed = JSON.parse(contents); } catch { throw new Error("invalid human decision state JSON"); }
		return validateRecord(parsed);
	}

	async load(requestId: string): Promise<HumanDecisionRecord | null> {
		return this.readState(requestId);
	}

	private async syncRoot(): Promise<void> {
		if (process.platform === "win32") return;
		const directory = await open(this.root, constants.O_RDONLY);
		try { await directory.sync(); } finally { await directory.close(); }
	}

	private lockPath(requestId: string, token: string, state: HumanDecisionLockLease["state"]): string {
		return join(this.root, `${requestId}.lock.${token}.${state}`);
	}

	private async publishLock(requestId: string): Promise<HumanDecisionLockLease> {
		const token = randomUUID();
		const createdAt = new Date().toISOString();
		const lockPath = this.lockPath(requestId, token, "candidate");
		const candidatePath = join(this.root, `.${requestId}.${token}.lock-owner.tmp`);
		const owner = await open(candidatePath, "wx", 0o600);
		try {
			await owner.writeFile(JSON.stringify({
				schemaVersion: 1,
				pid: process.pid,
				token,
				createdAt,
			}), "utf8");
			await owner.sync();
		} finally {
			await owner.close();
		}
		try {
			await link(candidatePath, lockPath);
			try {
				await this.syncRoot();
			} catch (error) {
				const published = await this.readLockOwner(lockPath);
				if (published?.token === token && published.pid === process.pid) await unlink(lockPath);
				throw error;
			}
			return { requestId, path: lockPath, pid: process.pid, token, createdAt, state: "candidate" };
		} finally {
			await rm(candidatePath, { force: true });
		}
	}

	private async acquireLock(requestId: string): Promise<HumanDecisionLockLease> {
		await this.ensureRoot();
		let lease = await this.publishLock(requestId);
		try {
			for (let attempt = 0; attempt <= this.lockMaxAttempts; attempt += 1) {
				await new Promise((resolve) => setTimeout(resolve, this.lockRetryMs));
				const scan = await this.liveLocks(requestId);
				if (scan.status === "changed") continue;
				const locks = scan.locks;
				const own = locks.find((lock) => lock.token === lease.token);
				if (!own) throw new Error("human decision lock candidate disappeared before acquisition");
				lease = own;
				const active = locks.filter((lock) => lock.state === "active").sort(compareLocks);
				if (lease.state === "active") {
					if (active[0]?.token !== lease.token) {
						lease = await this.moveLock(lease, "candidate");
						continue;
					}
					if (active.length === 1) return lease;
					continue;
				}
				if (active.length > 0) continue;
				const candidates = locks.filter((lock) => lock.state === "candidate").sort(compareLocks);
				if (candidates[0]?.token !== lease.token) continue;
				lease = await this.moveLock(lease, "active");
			}
			throw new Error("timed out acquiring human decision repository lock");
		} catch (error) {
			await this.removeOwnedLock(lease, true);
			throw error;
		}
	}

	private async lockMetadata(lockPath: string) {
		try {
			return await lstat(lockPath);
		} catch (error) {
			if (isRecord(error) && error.code === "ENOENT") return null;
			throw error;
		}
	}

	private async inspectLockOwner(lockPath: string): Promise<HumanDecisionLockOwnerInspection> {
		let handle: Awaited<ReturnType<typeof open>>;
		try {
			handle = await open(lockPath, constants.O_RDONLY | (constants.O_NOFOLLOW ?? 0));
		} catch (error) {
			if (isRecord(error) && error.code === "ENOENT") return { status: "missing" };
			throw error;
		}
		try {
			const metadata = await handle.stat();
			if (!metadata.isFile() || metadata.size > 4_096) return { status: "invalid" };
			let value: unknown;
			try { value = JSON.parse(await handle.readFile("utf8")); } catch { return { status: "invalid" }; }
			if (!isRecord(value)
				|| value.schemaVersion !== 1
				|| !Number.isSafeInteger(value.pid)
				|| (value.pid as number) < 1
				|| typeof value.token !== "string"
				|| !/^[0-9a-f-]{36}$/.test(value.token)
				|| typeof value.createdAt !== "string") return { status: "invalid" };
			let createdAt: string;
			try { createdAt = canonicalTimestamp(value.createdAt, "lock creation timestamp"); } catch {
				return { status: "invalid" };
			}
			return { status: "valid", owner: { pid: value.pid as number, token: value.token, createdAt } };
		} finally {
			await handle.close();
		}
	}

	private async readLockOwner(lockPath: string): Promise<HumanDecisionLockOwner | null> {
		const inspection = await this.inspectLockOwner(lockPath);
		return inspection.status === "valid" ? inspection.owner : null;
	}

	private processIsAlive(pid: number): boolean {
		try {
			process.kill(pid, 0);
			return true;
		} catch (error) {
			return !(isRecord(error) && error.code === "ESRCH");
		}
	}

	private async moveLock(
		lease: HumanDecisionLockLease,
		state: HumanDecisionLockLease["state"],
	): Promise<HumanDecisionLockLease> {
		const owner = await this.readLockOwner(lease.path);
		if (!owner || owner.token !== lease.token || owner.pid !== lease.pid) {
			throw new Error("human decision lock ownership changed before transition");
		}
		const path = this.lockPath(lease.requestId, lease.token, state);
		await rename(lease.path, path);
		await this.syncRoot();
		return { ...lease, path, state };
	}

	private async removeOwnedLock(lease: HumanDecisionLockLease, allowMissing = false): Promise<boolean> {
		const inspection = await this.inspectLockOwner(lease.path);
		if (inspection.status !== "valid") {
			if (inspection.status === "invalid") {
				throw new Error("human decision lock owner record is invalid before release");
			}
			if (allowMissing) return false;
			throw new Error("human decision lock ownership disappeared before release");
		}
		const owner = inspection.owner;
		if (owner.token !== lease.token || owner.pid !== lease.pid) {
			throw new Error("human decision lock ownership changed before release; replacement lock preserved");
		}
		try {
			await unlink(lease.path);
		} catch (error) {
			if (!(allowMissing && isRecord(error) && error.code === "ENOENT")) throw error;
			return false;
		}
		await this.syncRoot();
		return true;
	}

	private async liveLocks(requestId: string): Promise<HumanDecisionLockScan> {
		const expression = new RegExp(`^${requestId}\\.lock\\.([0-9a-f-]{36})\\.(candidate|active)$`);
		const locks: HumanDecisionLockLease[] = [];
		for (const entry of await readdir(this.root, { withFileTypes: true })) {
			const match = expression.exec(entry.name);
			if (!match) continue;
			const path = join(this.root, entry.name);
			if (!entry.isFile() || entry.isSymbolicLink()) throw new Error("human decision lock is not a trusted regular file");
			const metadata = await this.lockMetadata(path);
			if (!metadata) return { status: "changed" };
			const inspection = await this.inspectLockOwner(path);
			if (inspection.status === "missing") return { status: "changed" };
			const owner = inspection.status === "valid" ? inspection.owner : null;
			if (!owner || owner.token !== match[1]) {
				const current = await this.lockMetadata(path);
				if (!current || current.dev !== metadata.dev || current.ino !== metadata.ino) {
					return { status: "changed" };
				}
				if (Date.now() - metadata.mtimeMs <= this.lockStaleMs) {
					throw new Error("human decision lock owner record is invalid");
				}
				await rm(path, { force: true });
				continue;
			}
			const lock: HumanDecisionLockLease = {
				...owner,
				requestId,
				path,
				state: match[2] as HumanDecisionLockLease["state"],
			};
			if (!this.processIsAlive(owner.pid)) {
				if (!await this.removeOwnedLock(lock, true)) return { status: "changed" };
				continue;
			}
			locks.push(lock);
		}
		return { status: "stable", locks };
	}

	private async releaseLock(lease: HumanDecisionLockLease): Promise<void> {
		if (lease.state !== "active") throw new Error("human decision lock was not active at release");
		await this.removeOwnedLock(lease);
	}

	private async writeState(requestId: string, state: HumanDecisionRecord): Promise<void> {
		const validated = validateRecord(state);
		const serialized = `${JSON.stringify(validated, null, 2)}\n`;
		if (Buffer.byteLength(serialized) > MAX_STATE_BYTES) throw new Error("human decision state exceeds its byte limit");
		const temporary = join(this.root, `.${requestId}.${randomUUID()}.tmp`);
		const handle = await open(temporary, "wx", 0o600);
		try {
			await handle.writeFile(serialized, "utf8");
			await handle.sync();
		} finally {
			await handle.close();
		}
		try {
			await rename(temporary, this.statePath(requestId));
			await this.syncRoot();
		} catch (error) {
			await rm(temporary, { force: true });
			throw error;
		}
	}

	async transact<T>(
		requestId: string,
		operation: (state: HumanDecisionRecord | null) => Promise<HumanDecisionTransaction<T>> | HumanDecisionTransaction<T>,
	): Promise<T> {
		if (!REQUEST_ID.test(requestId)) throw new Error("invalid human decision request ID");
		const lock = await this.acquireLock(requestId);
		try {
			const existing = await this.readState(requestId);
			const transaction = await operation(existing === null ? null : structuredClone(existing));
			if (!isRecord(transaction) || !("state" in transaction) || !("value" in transaction)) {
				throw new Error("invalid human decision transaction result");
			}
			if (transaction.state === null) {
				await rm(this.statePath(requestId), { force: true });
			} else {
				await this.writeState(requestId, transaction.state);
			}
			return transaction.value;
		} finally {
			await this.releaseLock(lock);
		}
	}
}
