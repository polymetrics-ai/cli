import { execFile } from "node:child_process";

import {
	assertHumanDecisionBinding,
	consumeHumanDecision,
	expireHumanDecision,
	persistHumanDecisionRequest,
	recordHumanDecision,
	routeHumanDecisionTarget,
	validateHumanDecisionBinding,
	validateHumanDecisionRecord,
	validateHumanDecisionRequestComment,
	type HumanDecisionBinding,
	type HumanDecisionEvidence,
	type HumanDecisionRecord,
	type HumanDecisionRepository,
	type HumanDecisionRequestSpec,
	type ParentMergeDecisionOption,
	type StandardHumanDecisionGate,
} from "./human-decision.ts";

const GH_TIMEOUT_MS = 15_000;
const GH_MAX_OUTPUT_BYTES = 2 * 1024 * 1024;
const GH_COMMENTS_PER_PAGE = 100;
const GH_MAX_COMMENT_PAGES = 10;
const MAX_COMMENT_BODY_BYTES = 64 * 1024;
const MAX_COMMENT_ID = Number.MAX_SAFE_INTEGER;
const LOGIN = /^[a-z\d](?:[a-z\d-]{0,37}[a-z\d])?$/;
const EXACT_COMMAND = /^\/shepherd decide ([A-Za-z0-9][A-Za-z0-9_-]{0,127}) ([a-z][a-z0-9_-]{0,63})$/;
const RFC3339_UTC = /^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(?:\.(\d{1,3}))?Z$/;

export interface GitHubComment {
	id: number;
	url: string;
	body: string;
	actor: {
		login: string;
		type: string;
	};
	createdAt: string;
	updatedAt: string;
}

export interface GitHubDecisionTransport {
	getAuthenticatedActor(): Promise<string>;
	listComments(binding: HumanDecisionBinding): Promise<GitHubComment[]>;
	createDecisionRequestComment(record: HumanDecisionRecord): Promise<GitHubComment>;
}

interface GitHubDecisionRequestBase {
	requestId: string;
	repository: string;
	parentIssue: number;
	pullRequest: number;
	generation: number;
	headSha?: string;
	actorAllowlist: string[];
	expiresAt: string;
	question: string;
}

export type GitHubDecisionRequest =
	| (GitHubDecisionRequestBase & {
		gate: StandardHumanDecisionGate;
		allowedOptions: string[];
	})
	| (GitHubDecisionRequestBase & {
		gate: "parent_merge";
		allowedOptions: ParentMergeDecisionOption[];
	});

export interface DecisionPollingPolicy {
	maxAttempts: number;
	initialDelayMs: number;
	maxDelayMs: number;
}

export interface GitHubDecisionBrokerOptions {
	now?: () => Date;
	sleep?: (delayMs: number, signal?: AbortSignal) => Promise<void>;
	polling?: Partial<DecisionPollingPolicy>;
	transportRetry?: Partial<DecisionPollingPolicy>;
}

export type GitHubDecisionPollResult =
	| { status: "decided"; decision: HumanDecisionEvidence; attempts: number }
	| { status: "pending"; attempts: number }
	| { status: "expired"; attempts: number };

export interface GitHubDecisionPollOptions {
	signal?: AbortSignal;
}

export type GhDecisionExecutor = (file: string, args: string[]) => Promise<string>;

interface ClassifiedTransportFailure extends Error {
	readonly retryable: boolean;
}

class GitHubTransportFailure extends Error implements ClassifiedTransportFailure {
	readonly retryable: boolean;

	constructor(retryable: boolean) {
		super(`GitHub transport ${retryable ? "transient" : "permanent"} failure`);
		this.name = "GitHubTransportFailure";
		this.retryable = retryable;
	}
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null && !Array.isArray(value);
}

function canonicalTimestamp(value: unknown, description: string): string {
	if (typeof value !== "string" || value.length > 128) throw new Error(`GitHub returned invalid ${description}`);
	const match = RFC3339_UTC.exec(value);
	if (!match) throw new Error(`GitHub returned invalid ${description}`);
	const canonical = `${match[1]}.${(match[2] ?? "").padEnd(3, "0")}Z`;
	const timestamp = new Date(canonical);
	if (!Number.isFinite(timestamp.valueOf()) || timestamp.toISOString() !== canonical) {
		throw new Error(`GitHub returned invalid ${description}`);
	}
	return canonical;
}

function classifyTransportFailure(error: unknown, inspectExecutorDetails = false): ClassifiedTransportFailure {
	if (error instanceof GitHubTransportFailure) return error;
	if (isRecord(error) && typeof error.retryable === "boolean") {
		return new GitHubTransportFailure(error.retryable);
	}
	if (!inspectExecutorDetails) return new GitHubTransportFailure(false);
	const code = isRecord(error) && typeof error.code === "string" ? error.code.toUpperCase() : "";
	const killed = isRecord(error) && error.killed === true;
	const signal = isRecord(error) && typeof error.signal === "string" ? error.signal.toUpperCase() : "";
	const message = error instanceof Error ? error.message : "";
	const stderr = isRecord(error) && typeof error.stderr === "string" ? error.stderr.slice(0, 16_384) : "";
	const diagnostic = `${message}\n${stderr}`;
	const retryable = killed
		|| signal === "SIGTERM"
		|| ["EAI_AGAIN", "ECONNABORTED", "ECONNREFUSED", "ECONNRESET", "ENETDOWN", "ENETUNREACH", "ETIMEDOUT"].includes(code)
		|| /\b(?:408|425|429|500|502|503|504)\b/.test(diagnostic)
		|| /(?:timed?\s*out|timeout|temporar(?:y|ily) unavailable|connection reset|secondary rate limit|rate limit exceeded)/i.test(diagnostic);
	return new GitHubTransportFailure(retryable);
}

function normalizedActorLogin(value: unknown, description: string): string {
	if (typeof value !== "string") throw new Error(`GitHub returned invalid ${description}`);
	const login = value.toLowerCase();
	const base = login.endsWith("[bot]") ? login.slice(0, -5) : login;
	if (login.length > 64 || !LOGIN.test(base)) throw new Error(`GitHub returned invalid ${description}`);
	return login;
}

function validateComment(value: unknown): GitHubComment {
	if (!isRecord(value) || !Number.isSafeInteger(value.id) || (value.id as number) < 1 || (value.id as number) > MAX_COMMENT_ID) {
		throw new Error("GitHub returned an invalid comment ID");
	}
	if (typeof value.url !== "string" || value.url.length === 0 || value.url.length > 2_048) {
		throw new Error("GitHub returned an invalid comment URL");
	}
	if (typeof value.body !== "string" || Buffer.byteLength(value.body) > MAX_COMMENT_BODY_BYTES) {
		throw new Error("GitHub returned an invalid or oversized comment body");
	}
	if (!isRecord(value.actor) || typeof value.actor.type !== "string" || value.actor.type.length === 0 || value.actor.type.length > 64) {
		throw new Error("GitHub returned an invalid comment actor");
	}
	return {
		id: value.id as number,
		url: value.url,
		body: value.body,
		actor: { login: normalizedActorLogin(value.actor.login, "comment actor"), type: value.actor.type },
		createdAt: canonicalTimestamp(value.createdAt, "comment creation timestamp"),
		updatedAt: canonicalTimestamp(value.updatedAt, "comment update timestamp"),
	};
}

function defaultSleep(delayMs: number, signal?: AbortSignal): Promise<void> {
	if (signal?.aborted) return Promise.reject(signal.reason ?? new Error("human decision polling aborted"));
	return new Promise((resolve, reject) => {
		const finish = () => {
			signal?.removeEventListener("abort", abort);
			resolve();
		};
		const timer = setTimeout(finish, delayMs);
		const abort = () => {
			clearTimeout(timer);
			signal?.removeEventListener("abort", abort);
			reject(signal?.reason ?? new Error("human decision polling aborted"));
		};
		signal?.addEventListener("abort", abort, { once: true });
	});
}

function pollingValue(value: unknown, fallback: number, maximum: number, description: string): number {
	const resolved = value ?? fallback;
	if (!Number.isSafeInteger(resolved) || (resolved as number) < 1 || (resolved as number) > maximum) {
		throw new Error(`invalid human decision polling ${description}`);
	}
	return resolved as number;
}

function normalizePolling(value: Partial<DecisionPollingPolicy> = {}): DecisionPollingPolicy {
	const policy = {
		maxAttempts: pollingValue(value.maxAttempts, 8, 100, "attempt count"),
		initialDelayMs: pollingValue(value.initialDelayMs, 1_000, 300_000, "initial delay"),
		maxDelayMs: pollingValue(value.maxDelayMs, 30_000, 300_000, "maximum delay"),
	};
	if (policy.maxDelayMs < policy.initialDelayMs) throw new Error("human decision polling maximum delay is below its initial delay");
	return policy;
}

function normalizeTransportRetry(value: Partial<DecisionPollingPolicy> = {}): DecisionPollingPolicy {
	const policy = {
		maxAttempts: pollingValue(value.maxAttempts, 3, 10, "transport retry attempt count"),
		initialDelayMs: pollingValue(value.initialDelayMs, 250, 30_000, "transport retry initial delay"),
		maxDelayMs: pollingValue(value.maxDelayMs, 2_000, 30_000, "transport retry maximum delay"),
	};
	if (policy.maxDelayMs < policy.initialDelayMs) {
		throw new Error("human decision transport retry maximum delay is below its initial delay");
	}
	return policy;
}

function requestSpec(request: GitHubDecisionRequest): HumanDecisionRequestSpec {
	const target = routeHumanDecisionTarget(request.gate, request.parentIssue, request.pullRequest);
	return {
		requestId: request.requestId,
		gate: request.gate,
		binding: {
			repository: request.repository,
			target,
			generation: request.generation,
			...(request.headSha !== undefined ? { headSha: request.headSha } : {}),
		},
		allowedOptions: request.allowedOptions,
		actorAllowlist: request.actorAllowlist,
		expiresAt: request.expiresAt,
		question: request.question,
	} as HumanDecisionRequestSpec;
}

function escapeMarkdownText(value: string): string {
	return value
		.replace(/[\\`*_{}\[\]()<>#+.!|>~-]/g, "\\$&")
		.replace(/@/g, "&#64;");
}

function renderQuotedQuestion(value: string): string {
	return value.split("\n").map((line) => `> ${escapeMarkdownText(line)}`).join("\n");
}

export function renderDecisionRequestComment(record: HumanDecisionRecord): string {
	const validated = validateHumanDecisionRecord(record);
	const head = validated.binding.headSha ?? "not-applicable (issue-scoped gate)";
	const gateWarning = validated.gate === "parent_merge"
		? "This is the distinct parent-merge approval for this exact head; no other signal authorizes that merge."
		: "No review text, reaction, status check, emoji, or silence authorizes this decision.";
	return [
		validated.idempotencyMarker,
		"### Shepherd human decision required",
		"",
		renderQuotedQuestion(validated.question),
		"",
		`- Request: \`${validated.requestId}\``,
		`- Gate: \`${validated.gate}\``,
		`- Repository: \`${validated.binding.repository}\``,
		`- Target: \`${validated.binding.target.kind} #${validated.binding.target.number}\``,
		`- Generation: \`${validated.binding.generation}\``,
		`- Exact head: \`${head}\``,
		`- Expires: \`${validated.expiresAt}\``,
		`- Allowed options: ${validated.allowedOptions.map((option) => `\`${option}\``).join(", ")}`,
		`- Allowed humans: ${validated.actorAllowlist.map((actor) => `@${actor}`).join(", ")}`,
		"",
		`Reply with exactly \`/shepherd decide ${validated.requestId} <option>\` from an allowlisted human account.`,
		gateWarning,
	].join("\n");
}

function requestCommentEvidence(record: HumanDecisionRecord, comment: GitHubComment): NonNullable<HumanDecisionRecord["requestComment"]> {
	return validateHumanDecisionRequestComment(record, {
		id: comment.id,
		url: comment.url,
		actor: comment.actor.login,
		createdAt: comment.createdAt,
	});
}

function assertOwnedRequestComment(
	record: HumanDecisionRecord,
	comment: GitHubComment,
	authenticatedActor: string,
): GitHubComment {
	const normalized = validateComment(comment);
	if (normalized.actor.login !== authenticatedActor) throw new Error("human decision marker comment has a foreign owner");
	if (normalized.createdAt !== normalized.updatedAt) throw new Error("human decision marker comment was edited");
	if (normalized.body !== renderDecisionRequestComment(record)) throw new Error("human decision marker body collision");
	return normalized;
}

function markerComments(record: HumanDecisionRecord, comments: readonly GitHubComment[]): GitHubComment[] {
	return comments.map(validateComment).filter((comment) => comment.body.includes(record.idempotencyMarker));
}

function assertPersistedRequestComment(record: HumanDecisionRecord, comments: readonly GitHubComment[]): GitHubComment {
	if (!record.requestComment) throw new Error("human decision request comment is not persisted");
	const validated = comments.map(validateComment);
	const markers = markerComments(record, validated);
	if (markers.length !== 1 || markers[0].id !== record.requestComment.id) {
		throw new Error("persisted human decision marker is missing, duplicated, or no longer authoritative");
	}
	const matches = validated.filter((comment) => comment.id === record.requestComment?.id);
	if (matches.length !== 1) throw new Error("persisted human decision marker comment is missing or duplicated");
	const comment = matches[0];
	if (comment.body !== renderDecisionRequestComment(record)
		|| comment.actor.login !== record.requestComment.actor
		|| comment.url !== record.requestComment.url
		|| comment.createdAt !== record.requestComment.createdAt
		|| comment.updatedAt !== comment.createdAt) {
		throw new Error("persisted human decision marker comment changed or lost ownership");
	}
	return comment;
}

function parseValidDecision(record: HumanDecisionRecord, comment: GitHubComment, observedAt: Date): HumanDecisionEvidence | null {
	if (comment.id <= (record.requestComment?.id ?? MAX_COMMENT_ID)) return null;
	if (comment.actor.type !== "User" || comment.createdAt !== comment.updatedAt) return null;
	if (!record.actorAllowlist.includes(comment.actor.login)) return null;
	if (!Number.isFinite(observedAt.valueOf()) || new Date(comment.createdAt).valueOf() > observedAt.valueOf()) return null;
	const match = EXACT_COMMAND.exec(comment.body);
	if (!match
		|| match[1] !== record.requestId
		|| !(record.allowedOptions as readonly string[]).includes(match[2])) return null;
	return { option: match[2], actor: comment.actor.login, sourceUrl: comment.url, decidedAt: comment.createdAt };
}

export class GitHubDecisionBroker {
	private readonly repository: HumanDecisionRepository;
	private readonly transport: GitHubDecisionTransport;
	private readonly now: () => Date;
	private readonly sleep: (delayMs: number, signal?: AbortSignal) => Promise<void>;
	private readonly polling: DecisionPollingPolicy;
	private readonly transportRetry: DecisionPollingPolicy;

	constructor(
		repository: HumanDecisionRepository,
		transport: GitHubDecisionTransport,
		options: GitHubDecisionBrokerOptions = {},
	) {
		this.repository = repository;
		this.transport = transport;
		this.now = options.now ?? (() => new Date());
		this.sleep = options.sleep ?? defaultSleep;
		this.polling = normalizePolling(options.polling);
		this.transportRetry = normalizeTransportRetry(options.transportRetry);
	}

	private retryDelay(attempt: number): number {
		return Math.min(
			this.transportRetry.maxDelayMs,
			this.transportRetry.initialDelayMs * (2 ** (attempt - 1)),
		);
	}

	private async callTransport<T>(
		operation: () => Promise<T>,
		signal?: AbortSignal,
	): Promise<T> {
		for (let attempt = 1; attempt <= this.transportRetry.maxAttempts; attempt += 1) {
			try {
				return await operation();
			} catch (error) {
				const failure = classifyTransportFailure(error);
				if (!failure.retryable || attempt === this.transportRetry.maxAttempts) throw failure;
				await this.sleep(this.retryDelay(attempt), signal);
			}
		}
		throw new Error("unreachable GitHub transport retry state");
	}

	private async ensureRequestComment(
		record: HumanDecisionRecord,
		authenticatedActor: string,
	): Promise<GitHubComment> {
		for (let attempt = 1; attempt <= this.transportRetry.maxAttempts; attempt += 1) {
			const comments = await this.callTransport(() => this.transport.listComments(record.binding));
			const markers = markerComments(record, comments);
			if (markers.length > 1) throw new Error("duplicate human decision marker comments are ambiguous");
			if (markers.length === 1) return assertOwnedRequestComment(record, markers[0], authenticatedActor);
			let created: GitHubComment;
			try {
				created = await this.transport.createDecisionRequestComment(record);
			} catch (error) {
				const failure = classifyTransportFailure(error);
				if (!failure.retryable || attempt === this.transportRetry.maxAttempts) throw failure;
				await this.sleep(this.retryDelay(attempt));
				continue;
			}
			return assertOwnedRequestComment(record, created, authenticatedActor);
		}
		throw new Error("unreachable GitHub request-comment retry state");
	}

	async readRecord(requestId: string, binding: HumanDecisionBinding): Promise<HumanDecisionRecord> {
		const stored = await this.repository.load(requestId);
		if (stored === null) throw new Error("human decision request does not exist");
		const record = validateHumanDecisionRecord(stored, this.now());
		assertHumanDecisionBinding(record, binding);
		return record;
	}

	async request(request: GitHubDecisionRequest): Promise<HumanDecisionRecord> {
		const proposedAt = this.now();
		const persisted = validateHumanDecisionRecord(
			await persistHumanDecisionRequest(this.repository, requestSpec(request), proposedAt),
			proposedAt,
		);
		return this.repository.transact(persisted.requestId, async (existing) => {
			if (existing === null) throw new Error("human decision request disappeared during marker creation");
			const observedAt = this.now();
			const record = validateHumanDecisionRecord(existing, observedAt);
			assertHumanDecisionBinding(record, persisted.binding);
			if (record.requestComment) return { state: record, value: record };
			if (observedAt.valueOf() >= new Date(record.expiresAt).valueOf()) {
				const timestamp = this.now().toISOString();
				const expired = { ...record, status: "expired" as const, updatedAt: timestamp };
				return { state: expired, value: expired };
			}
			const authenticatedActor = normalizedActorLogin(
				await this.callTransport(() => this.transport.getAuthenticatedActor()),
				"authenticated actor",
			);
			const comment = await this.ensureRequestComment(record, authenticatedActor);
			const updated = {
				...record,
				requestComment: requestCommentEvidence(record, comment),
				updatedAt: this.now().toISOString(),
			};
			return { state: updated, value: updated };
		});
	}

	async poll(
		requestId: string,
		binding: HumanDecisionBinding,
		options: GitHubDecisionPollOptions = {},
	): Promise<GitHubDecisionPollResult> {
		for (let attempt = 1; attempt <= this.polling.maxAttempts; attempt += 1) {
			if (options.signal?.aborted) throw options.signal.reason ?? new Error("human decision polling aborted");
			const record = await this.readRecord(requestId, binding);
			if ((record.status === "decided" || record.status === "consumed") && record.decision) {
				return { status: "decided", decision: { ...record.decision }, attempts: attempt };
			}
			if (record.status === "expired" || this.now().valueOf() >= new Date(record.expiresAt).valueOf()) {
				if (record.status !== "expired") await expireHumanDecision(this.repository, requestId, binding, this.now());
				return { status: "expired", attempts: attempt };
			}
			if (record.status !== "pending") throw new Error(`human decision request cannot be polled while ${record.status}`);
			const comments = await this.callTransport(
				() => this.transport.listComments(record.binding),
				options.signal,
			);
			assertPersistedRequestComment(record, comments);
			const candidates = comments.map(validateComment)
				.map((comment) => parseValidDecision(record, comment, this.now()))
				.filter((decision): decision is HumanDecisionEvidence => decision !== null);
			if (candidates.length > 1) throw new Error("multiple valid human decision responses are ambiguous");
			if (candidates.length === 1) {
				const decided = await recordHumanDecision(this.repository, requestId, binding, candidates[0]);
				return { status: "decided", decision: { ...decided.decision! }, attempts: attempt };
			}
			if (attempt < this.polling.maxAttempts) {
				const delay = Math.min(this.polling.maxDelayMs, this.polling.initialDelayMs * (2 ** (attempt - 1)));
				await this.sleep(delay, options.signal);
			}
		}
		return { status: "pending", attempts: this.polling.maxAttempts };
	}

	async consume(requestId: string, binding: HumanDecisionBinding): Promise<HumanDecisionEvidence> {
		return consumeHumanDecision(this.repository, requestId, binding, this.now());
	}
}

function parseGitHubApiComment(value: unknown): GitHubComment {
	if (!isRecord(value) || !isRecord(value.user)) throw new Error("GitHub returned a malformed issue comment");
	return validateComment({
		id: value.id,
		url: value.html_url,
		body: value.body,
		actor: { login: value.user.login, type: value.user.type },
		createdAt: value.created_at,
		updatedAt: value.updated_at,
	});
}

function parseJson(output: string, description: string): unknown {
	if (Buffer.byteLength(output) > GH_MAX_OUTPUT_BYTES) throw new Error(`GitHub ${description} output exceeds its byte limit`);
	try { return JSON.parse(output); } catch { throw new Error(`GitHub returned malformed ${description} JSON`); }
}

function defaultGhExecutor(file: string, args: string[]): Promise<string> {
	return new Promise((resolve, reject) => {
		execFile(file, args, {
			encoding: "utf8",
			env: process.env,
			maxBuffer: GH_MAX_OUTPUT_BYTES,
			timeout: GH_TIMEOUT_MS,
			killSignal: "SIGTERM",
		}, (error, stdout) => {
			if (error) {
				reject(error);
				return;
			}
			resolve(stdout);
		});
	});
}

function commentsEndpoint(binding: HumanDecisionBinding, page?: number): string {
	const validated = validateHumanDecisionBinding(binding);
	const endpoint = `repos/${validated.repository}/issues/${validated.target.number}/comments`;
	return page === undefined ? endpoint : `${endpoint}?per_page=${GH_COMMENTS_PER_PAGE}&page=${page}`;
}

export class GhCliDecisionTransport implements GitHubDecisionTransport {
	private readonly execute: GhDecisionExecutor;

	constructor(execute: GhDecisionExecutor = defaultGhExecutor) {
		this.execute = execute;
	}

	private async executeGh(args: string[]): Promise<string> {
		try {
			return await this.execute("gh", args);
		} catch (error) {
			throw classifyTransportFailure(error, true);
		}
	}

	async getAuthenticatedActor(): Promise<string> {
		const payload = parseJson(await this.executeGh(["api", "--method", "GET", "/user"]), "authenticated user");
		if (!isRecord(payload)) throw new Error("GitHub returned malformed authenticated user evidence");
		return normalizedActorLogin(payload.login, "authenticated user login");
	}

	async listComments(binding: HumanDecisionBinding): Promise<GitHubComment[]> {
		const comments: GitHubComment[] = [];
		for (let page = 1; page <= GH_MAX_COMMENT_PAGES; page += 1) {
			const payload = parseJson(
				await this.executeGh(["api", "--method", "GET", commentsEndpoint(binding, page)]),
				"issue comments",
			);
			if (!Array.isArray(payload) || payload.length > GH_COMMENTS_PER_PAGE) {
				throw new Error("GitHub returned malformed or unbounded issue comments");
			}
			comments.push(...payload.map(parseGitHubApiComment));
			if (payload.length < GH_COMMENTS_PER_PAGE) return comments;
		}
		throw new Error("GitHub issue comments exceed the bounded pagination window");
	}

	async createDecisionRequestComment(record: HumanDecisionRecord): Promise<GitHubComment> {
		const validated = validateHumanDecisionRecord(record);
		const body = renderDecisionRequestComment(validated);
		if (Buffer.byteLength(body) > MAX_COMMENT_BODY_BYTES) {
			throw new Error("human decision request comment is empty or oversized");
		}
		const payload = parseJson(await this.executeGh([
			"api", "--method", "POST", commentsEndpoint(validated.binding), "-f", `body=${body}`,
		]), "created issue comment");
		return parseGitHubApiComment(payload);
	}
}
