import { execFile } from "node:child_process";
import { createHash } from "node:crypto";

import {
	createRequiredGitHubCheckPolicy,
	createRequiredGitHubCheckPolicyObservation,
	type GitHubCheckEvidence,
	type GitHubFindingDisposition,
	type GitHubPullRequestEvidence,
	type GitHubRequestedChangeEvidence,
	type GitHubReviewThreadEvidence,
	type RequiredGitHubCheck,
} from "./github-evidence.ts";
import {
	validateIndependentReviewRecord,
	type IndependentReviewRecord,
} from "./review-router.ts";
import { validateReviewedChildIntegrationEvidence } from "./github-orchestrator.ts";
import type {
	AuthoritativeLookup,
	ChildIntegrationQuery,
	ChildIntegrationReceipt,
	ChildIssueMarkerQuery,
	CreateChildIssueRequest,
	CreatePullRequestRequest,
	DurableMutationIntent,
	DurableMutationResult,
	ExternalCallContext,
	GitAncestryProof,
	GitAncestryQuery,
	GitHubChildIssue,
	GitHubOrchestrationTransport,
	GitHubRosterQuery,
	GitHubRosterSnapshot,
	IntegrateChildRequest,
	ParentOrchestrationPolicyAuthority,
	PublishRosterRequest,
	PullRequestMarkerQuery,
	RequiredCheckPolicySource,
} from "./github-orchestrator.ts";

const DEFAULT_TIMEOUT_MS = 15_000;
const DEFAULT_MAX_OUTPUT_BYTES = 2 * 1024 * 1024;
const DEFAULT_MAX_PAGES = 10;
const PAGE_SIZE = 100;
const REPOSITORY = /^[A-Za-z0-9][A-Za-z0-9._-]{0,99}\/[A-Za-z0-9][A-Za-z0-9._-]{0,99}$/u;
const SHA = /^[0-9a-f]{40}$/u;
const PR_META_PREFIX = "<!-- shepherd-transport-pr-meta:v1:";
const INTEGRATION_PENDING_PREFIX = "<!-- shepherd-transport-integration-pending:v1:";
const INTEGRATION_PREFIX = "<!-- shepherd-transport-integration:v1:";
const REVIEW_PREFIX = "<!-- shepherd-production-review:v1:";
const ROSTER_META_PREFIX = "<!-- shepherd-transport-roster-meta:v1:";
const MAX_GITHUB_REVISION = 2_147_483_647;

export interface GhExecutionOptions {
	signal: AbortSignal;
	timeoutMs: number;
	maxOutputBytes: number;
}

export type GhOrchestrationExecutor = (
	file: "gh",
	args: readonly string[],
	options: GhExecutionOptions,
) => Promise<string>;

export interface GhCliOrchestrationTransportOptions {
	execute?: GhOrchestrationExecutor;
	now?: () => Date;
	timeoutMs?: number;
	maxOutputBytes?: number;
	maxPages?: number;
}

export interface ProductionGitHubOrchestrationFacade {
	transport: GitHubOrchestrationTransport;
	policySource: RequiredCheckPolicySource;
	parentReadyAuthority: null;
	parentReadyAuthorityDependency: "required-external-durable-authority";
}

interface PullRequestMetadata {
	workItemId: string;
	generation: number;
	marker: string;
	baseSha: string;
	headSha: string;
	changedPaths: string[];
	allowedScopes: string[];
	policyDigest: string;
}

interface ReviewCommentArtifact {
	review: IndependentReviewRecord;
	dispositions: GitHubFindingDisposition[];
}

function positiveInteger(value: unknown, description: string): number {
	if (!Number.isSafeInteger(value) || (value as number) < 1 || (value as number) > Number.MAX_SAFE_INTEGER) {
		throw new Error(`invalid ${description}`);
	}
	return value as number;
}

function repository(value: unknown): string {
	if (typeof value !== "string" || !REPOSITORY.test(value)) throw new Error("invalid GitHub repository");
	return value;
}

function sha(value: unknown, description: string): string {
	if (typeof value !== "string" || !SHA.test(value)) throw new Error(`invalid ${description}`);
	return value;
}

function record(value: unknown, description: string): Record<string, unknown> {
	if (typeof value !== "object" || value === null || Array.isArray(value)) throw new Error(`GitHub returned malformed ${description}`);
	return value as Record<string, unknown>;
}

function array(value: unknown, description: string, maximum = PAGE_SIZE): unknown[] {
	if (!Array.isArray(value) || value.length > maximum) throw new Error(`GitHub returned malformed or unbounded ${description}`);
	return value;
}

function text(value: unknown, description: string, maximum = 65_536): string {
	if (typeof value !== "string" || Buffer.byteLength(value) > maximum) throw new Error(`GitHub returned invalid ${description}`);
	return value;
}

function timestamp(value: unknown, description: string): string {
	const date = new Date(text(value, description, 128));
	if (!Number.isFinite(date.valueOf())) throw new Error(`GitHub returned invalid ${description}`);
	return date.toISOString();
}

function parseJson(output: string, description: string, maximum: number): unknown {
	if (Buffer.byteLength(output) > maximum) throw new Error(`GitHub ${description} output exceeded its bound`);
	try { return JSON.parse(output); } catch { throw new Error(`GitHub returned malformed ${description} JSON`); }
}

function encodedEnvelope(prefix: string, value: unknown): string {
	return `${prefix}${Buffer.from(JSON.stringify(value)).toString("base64url")} -->`;
}

function decodeEnvelope(body: unknown, prefix: string, description: string): unknown | null {
	if (typeof body !== "string" || !body.startsWith(prefix) || !body.endsWith(" -->")) return null;
	const encoded = body.slice(prefix.length, -4);
	if (!/^[A-Za-z0-9_-]+$/u.test(encoded) || encoded.length > 1_000_000) throw new Error(`invalid ${description} envelope`);
	let decoded: string;
	try { decoded = Buffer.from(encoded, "base64url").toString("utf8"); } catch { throw new Error(`invalid ${description} envelope`); }
	return parseJson(decoded, description, DEFAULT_MAX_OUTPUT_BYTES);
}

function parentIssueFromBody(body: string): number {
	const matches = [...body.matchAll(/^Parent: #(\d+)$/gmu)];
	if (matches.length !== 1) throw new Error("child issue body does not contain one canonical parent issue");
	return positiveInteger(Number(matches[0][1]), "child parent issue");
}

function durableResult<T>(intent: DurableMutationIntent, revision: number, applied: boolean, value: T): DurableMutationResult<T> {
	return {
		schemaVersion: 1,
		idempotencyKey: intent.idempotencyKey,
		intentDigest: intent.intentDigest,
		revision: positiveInteger(revision, "mutation revision"),
		applied,
		value,
	};
}

function revisionAt(value: string | Date): number {
	const milliseconds = value instanceof Date ? value.valueOf() : new Date(value).valueOf();
	const revision = Math.floor(milliseconds / 1_000);
	if (!Number.isSafeInteger(revision) || revision < 1 || revision > MAX_GITHUB_REVISION) {
		throw new Error("GitHub observation timestamp is outside the supported durable revision range");
	}
	return revision;
}

function stableResourceRevision(value: unknown): number {
	return Number.parseInt(createHash("sha256").update(String(value)).digest("hex").slice(0, 7), 16) + 1;
}

function validateBound(value: unknown, fallback: number, maximum: number, description: string): number {
	const resolved = value ?? fallback;
	if (!Number.isSafeInteger(resolved) || (resolved as number) < 1 || (resolved as number) > maximum) {
		throw new Error(`invalid ${description}`);
	}
	return resolved as number;
}

export const defaultGhOrchestrationExecutor: GhOrchestrationExecutor = (file, args, options) => new Promise((resolve, reject) => {
	execFile(file, [...args], {
		encoding: "utf8",
		env: process.env,
		maxBuffer: options.maxOutputBytes,
		timeout: options.timeoutMs,
		killSignal: "SIGTERM",
		signal: options.signal,
	}, (error, stdout) => {
		if (error !== null) {
			// Never propagate stderr, argv, environment, or credential-bearing diagnostics.
			reject(Object.assign(new Error("bounded GitHub CLI operation failed"), {
				code: (error as NodeJS.ErrnoException).code,
				killed: (error as { killed?: boolean }).killed,
			}));
			return;
		}
		resolve(stdout);
	});
});

export function createProductionGitHubOrchestrationFacade(
	options: GhCliOrchestrationTransportOptions = {},
): ProductionGitHubOrchestrationFacade {
	const policySource = new GhRequiredCheckPolicySource(options);
	return {
		transport: new GhCliOrchestrationTransport(options),
		policySource,
		parentReadyAuthority: null,
		parentReadyAuthorityDependency: "required-external-durable-authority",
	};
}

function stablePolicyRevision(value: unknown): number {
	return Number.parseInt(createHash("sha256").update(JSON.stringify(value)).digest("hex").slice(0, 7), 16) + 1;
}

/** Authoritative branch-protection policy adapter used for plan creation and later CI revalidation. */
export class GhRequiredCheckPolicySource implements RequiredCheckPolicySource {
	readonly #execute: GhOrchestrationExecutor;
	readonly #now: () => Date;
	readonly #timeoutMs: number;
	readonly #maxOutputBytes: number;

	constructor(options: GhCliOrchestrationTransportOptions = {}) {
		this.#execute = options.execute ?? defaultGhOrchestrationExecutor;
		this.#now = options.now ?? (() => new Date());
		this.#timeoutMs = validateBound(options.timeoutMs, DEFAULT_TIMEOUT_MS, 120_000, "GitHub timeout");
		this.#maxOutputBytes = validateBound(options.maxOutputBytes, DEFAULT_MAX_OUTPUT_BYTES, 8 * 1024 * 1024, "GitHub output limit");
	}

	async #policy(repositoryValue: string, branch: string, context: ExternalCallContext) {
		const repo = repository(repositoryValue);
		const deadline = new Date(context.deadlineAt).valueOf();
		let output: string;
		try {
			output = await this.#execute("gh", ["api", "--method", "GET",
				`/repos/${repo}/branches/${encodeURIComponent(branch)}/protection/required_status_checks`], {
				signal: context.signal,
				timeoutMs: Math.max(1, Math.min(this.#timeoutMs, deadline - Date.now())),
				maxOutputBytes: this.#maxOutputBytes,
			});
		} finally {
			if (context.signal.aborted) context.acknowledgeAbort();
		}
		const raw = record(parseJson(output, "required-check policy", this.#maxOutputBytes), "required-check policy");
		const checks: RequiredGitHubCheck[] = [];
		const modernContexts = new Set<string>();
		for (const item of array(raw.checks ?? [], "required-check app contexts")) {
			const check = record(item, "required-check app context");
			const name = text(check.context, "required-check context", 256);
			modernContexts.add(name);
			checks.push({ name, producerId: String(positiveInteger(check.app_id, "required-check app ID")) });
		}
		for (const item of array(raw.contexts ?? [], "required-check legacy contexts")) {
			const name = text(item, "required-check context", 256);
			if (!modernContexts.has(name)) checks.push({ name, producerId: "legacy" });
		}
		checks.sort((left, right) => left.name.localeCompare(right.name) || left.producerId.localeCompare(right.producerId));
		const policy = createRequiredGitHubCheckPolicy({
			schemaVersion: 1,
			repository: repo,
			baseBranch: branch,
			revision: stablePolicyRevision(checks),
			requiredChecks: checks,
		});
		return policy;
	}

	async findRequiredCheckPolicies(
		query: { repository: string; baseBranch: string },
		context: ExternalCallContext,
	) {
		const policy = await this.#policy(query.repository, query.baseBranch, context);
		return {
			items: [createRequiredGitHubCheckPolicyObservation({
				schemaVersion: 1,
				authority: "controller",
				repository: policy.repository,
				baseBranch: policy.baseBranch,
				revision: policy.revision,
				digest: policy.digest,
				observedAt: this.#now().toISOString(),
			})],
			complete: true,
		};
	}

	async findParentOrchestrationPolicyBundle(
		query: { repository: string; parentIssue: number; generation: number; parentBranch: string; parentBaseBranch: string },
		context: ExternalCallContext,
	): Promise<AuthoritativeLookup<ParentOrchestrationPolicyAuthority>> {
		const policies = await Promise.all([
			this.#policy(query.repository, query.parentBranch, context),
			this.#policy(query.repository, query.parentBaseBranch, context),
		]);
		const observedAt = this.#now().toISOString();
		return {
			items: [{
				schemaVersion: 1,
				authority: "controller",
				repository: query.repository,
				parentIssue: query.parentIssue,
				generation: query.generation,
				parentBranch: query.parentBranch,
				parentBaseBranch: query.parentBaseBranch,
				revision: stablePolicyRevision(policies.map((policy) => policy.digest).sort()),
				observedAt,
				policyBundle: { schemaVersion: 1, requiredCheckPolicies: policies },
			}],
			complete: true,
		};
	}
}

export class GhCliOrchestrationTransport implements GitHubOrchestrationTransport {
	readonly #execute: GhOrchestrationExecutor;
	readonly #now: () => Date;
	readonly #timeoutMs: number;
	readonly #maxOutputBytes: number;
	readonly #maxPages: number;

	constructor(options: GhCliOrchestrationTransportOptions = {}) {
		this.#execute = options.execute ?? defaultGhOrchestrationExecutor;
		this.#now = options.now ?? (() => new Date());
		this.#timeoutMs = validateBound(options.timeoutMs, DEFAULT_TIMEOUT_MS, 120_000, "GitHub timeout");
		this.#maxOutputBytes = validateBound(options.maxOutputBytes, DEFAULT_MAX_OUTPUT_BYTES, 8 * 1024 * 1024, "GitHub output limit");
		this.#maxPages = validateBound(options.maxPages, DEFAULT_MAX_PAGES, 100, "GitHub page limit");
	}

	async #gh(args: readonly string[], context: ExternalCallContext, mutation = false): Promise<string> {
		if (context.signal.aborted) {
			context.acknowledgeAbort();
			throw new Error("GitHub operation cancelled");
		}
		const deadline = new Date(context.deadlineAt).valueOf();
		if (!Number.isFinite(deadline) || deadline <= Date.now()) throw new Error("GitHub operation deadline expired");
		try {
			return await this.#execute("gh", args, {
				signal: context.signal,
				timeoutMs: Math.max(1, Math.min(this.#timeoutMs, deadline - Date.now())),
				maxOutputBytes: this.#maxOutputBytes,
			});
		} catch (error) {
			if (context.signal.aborted) {
				context.acknowledgeAbort();
				throw new Error("GitHub operation cancelled");
			}
			const code = typeof error === "object" && error !== null && "code" in error ? String(error.code) : "";
			throw Object.assign(new Error(mutation ? "GitHub mutation failed with uncertain publication" : "GitHub lookup failed"), {
				code,
				uncertain: mutation,
			});
		}
	}

	async #api(method: "GET" | "POST" | "PATCH" | "PUT", endpoint: string, context: ExternalCallContext, fields: readonly string[] = []): Promise<unknown> {
		const output = await this.#gh(["api", "--method", method, endpoint, ...fields.flatMap((field) => ["-f", field])], context, method !== "GET");
		return parseJson(output, "API", this.#maxOutputBytes);
	}

	async #paged(endpoint: (page: number) => string, context: ExternalCallContext): Promise<{ items: unknown[]; complete: boolean }> {
		const items: unknown[] = [];
		for (let page = 1; page <= this.#maxPages; page += 1) {
			const current = array(await this.#api("GET", endpoint(page), context), "page");
			items.push(...current);
			if (current.length < PAGE_SIZE) return { items, complete: true };
		}
		return { items, complete: false };
	}

	async #issueComments(repo: string, issue: number, context: ExternalCallContext): Promise<{ items: Record<string, unknown>[]; complete: boolean }> {
		const result = await this.#paged(
			(page) => `/repos/${repo}/issues/${issue}/comments?per_page=${PAGE_SIZE}&page=${page}`,
			context,
		);
		return { items: result.items.map((entry) => record(entry, "issue comment")), complete: result.complete };
	}

	async findChildIssues(query: ChildIssueMarkerQuery, context: ExternalCallContext): Promise<AuthoritativeLookup<GitHubChildIssue>> {
		const repo = repository(query.repository);
		const result = await this.#paged(
			(page) => `/repos/${repo}/issues?state=all&sort=created&direction=desc&per_page=${PAGE_SIZE}&page=${page}`,
			context,
		);
		const matches = result.items.map((entry) => record(entry, "issue"))
			.filter((entry) => entry.pull_request === undefined && typeof entry.body === "string"
				&& entry.body.split(query.marker).length === 2)
			.map((entry): GitHubChildIssue => ({
				number: positiveInteger(entry.number, "child issue number"),
				marker: query.marker,
				title: text(entry.title, "child issue title", 256),
				body: text(entry.body, "child issue body"),
				state: entry.state === "open" ? "open" : "closed",
				parentIssue: parentIssueFromBody(entry.body as string),
			}));
		return { items: matches, complete: result.complete };
	}

	async createChildIssue(request: CreateChildIssueRequest, context: ExternalCallContext): Promise<DurableMutationResult<GitHubChildIssue>> {
		const existing = await this.findChildIssues(request, context);
		if (existing.items.length > 1) throw new Error("duplicate child issue marker is ambiguous");
		if (existing.items.length === 1) return durableResult(request.mutation, existing.items[0].number, false, existing.items[0]);
		let created: Record<string, unknown> | undefined;
		try {
			created = record(await this.#api("POST", `/repos/${repository(request.repository)}/issues`, context, [
				`title=${request.title}`,
				`body=${request.body}`,
			]), "created issue");
		} catch (error) {
			const recovered = await this.findChildIssues(request, context);
			if (recovered.items.length === 1) return durableResult(request.mutation, recovered.items[0].number, false, recovered.items[0]);
			throw error;
		}
		const issue: GitHubChildIssue = {
			number: positiveInteger(created.number, "created issue number"),
			marker: request.marker,
			title: text(created.title, "created issue title", 256),
			body: text(created.body, "created issue body"),
			state: created.state === "open" ? "open" : "closed",
			parentIssue: request.parentIssue,
		};
		return durableResult(request.mutation, issue.number, true, issue);
	}

	async #findPullRows(query: PullRequestMarkerQuery, context: ExternalCallContext): Promise<{ rows: Record<string, unknown>[]; complete: boolean }> {
		const repo = repository(query.repository);
		const result = await this.#paged(
			(page) => `/repos/${repo}/pulls?state=all&sort=created&direction=desc&per_page=${PAGE_SIZE}&page=${page}`,
			context,
		);
		return {
			rows: result.items.map((entry) => record(entry, "pull request"))
				.filter((entry) => typeof entry.body === "string" && entry.body.split(query.marker).length === 2),
			complete: result.complete,
		};
	}

	async #metadata(repo: string, pullRequest: number, marker: string, context: ExternalCallContext): Promise<PullRequestMetadata> {
		const comments = await this.#issueComments(repo, pullRequest, context);
		if (!comments.complete) throw new Error("pull request metadata comments exceed the bounded window");
		const values = comments.items.map((entry) => decodeEnvelope(entry.body, PR_META_PREFIX, "pull request metadata"))
			.filter((entry): entry is Record<string, unknown> => entry !== null && typeof entry === "object" && !Array.isArray(entry))
			.filter((entry) => entry.marker === marker);
		if (values.length !== 1) throw new Error("pull request metadata is absent or ambiguous");
		const value = values[0];
		return {
			workItemId: text(value.workItemId, "pull request work item", 256),
			generation: positiveInteger(value.generation, "pull request generation"),
			marker: text(value.marker, "pull request marker", 512),
			baseSha: sha(value.baseSha, "pull request base SHA"),
			headSha: sha(value.headSha, "pull request head SHA"),
			changedPaths: array(value.changedPaths, "pull request changed paths").map((entry) => text(entry, "changed path", 4_096)).sort(),
			allowedScopes: array(value.allowedScopes, "pull request allowed scopes").map((entry) => text(entry, "allowed scope", 4_096)).sort(),
			policyDigest: text(value.policyDigest, "pull request policy digest", 64),
		};
	}

	async #reviews(repo: string, pullRequest: number, context: ExternalCallContext): Promise<ReviewCommentArtifact[]> {
		const comments = await this.#issueComments(repo, pullRequest, context);
		if (!comments.complete) throw new Error("review comments exceed the bounded window");
		return comments.items.map((entry) => decodeEnvelope(entry.body, REVIEW_PREFIX, "production review"))
			.filter((entry): entry is Record<string, unknown> => entry !== null && typeof entry === "object" && !Array.isArray(entry))
			.map((entry) => ({
				review: validateIndependentReviewRecord(entry.review),
				dispositions: array(entry.dispositions ?? [], "review dispositions").map((item) => record(item, "review disposition")) as unknown as GitHubFindingDisposition[],
			}));
	}

	async #checks(repo: string, headSha: string, context: ExternalCallContext): Promise<{ items: GitHubCheckEvidence[]; complete: boolean }> {
		const rawItems: Array<Record<string, unknown> & { source: "check" | "status" }> = [];
		let complete = true;
		for (let page = 1; page <= this.#maxPages; page += 1) {
			const payload = record(await this.#api("GET",
				`/repos/${repo}/commits/${headSha}/check-runs?per_page=${PAGE_SIZE}&page=${page}`, context), "check-runs");
			const pageItems = array(payload.check_runs, "check-runs");
			rawItems.push(...pageItems.map((item) => ({ ...record(item, "check-run"), source: "check" as const })));
			if (pageItems.length < PAGE_SIZE) break;
			if (page === this.#maxPages) complete = false;
		}
		for (let page = 1; page <= this.#maxPages; page += 1) {
			const pageItems = array(await this.#api("GET",
				`/repos/${repo}/commits/${headSha}/statuses?per_page=${PAGE_SIZE}&page=${page}`, context), "commit statuses");
			rawItems.push(...pageItems.map((item) => ({ ...record(item, "commit status"), source: "status" as const })));
			if (pageItems.length < PAGE_SIZE) break;
			if (page === this.#maxPages) complete = false;
		}
		const ordered = rawItems.sort((left, right) => {
			const leftTime = String(left.completed_at ?? left.updated_at ?? left.created_at ?? "");
			const rightTime = String(right.completed_at ?? right.updated_at ?? right.created_at ?? "");
			return leftTime.localeCompare(rightTime) || Number(left.id) - Number(right.id);
		});
		const sequences = new Map<string, number>();
		const items = ordered.map((item): GitHubCheckEvidence => {
			const checkRun = item.source === "check";
			const app = checkRun ? record(item.app, "check-run app") : undefined;
			const name = text(checkRun ? item.name : item.context, "check name", 256);
			const producerId = checkRun ? String(positiveInteger(app?.id, "check app ID")) : "legacy";
			const key = `${name}\u0000${producerId}\u0000${headSha}`;
			const sequence = (sequences.get(key) ?? 0) + 1;
			sequences.set(key, sequence);
			const rawStatus = checkRun ? item.status : item.state === "pending" ? "in_progress" : "completed";
			const status = rawStatus === "queued" ? "queued" : rawStatus === "in_progress" ? "in_progress" : "completed";
			let conclusion: GitHubCheckEvidence["conclusion"] = null;
			if (status === "completed") {
				const value = checkRun ? item.conclusion : item.state === "success" ? "success" : "failure";
				conclusion = ["success", "failure", "cancelled", "timed_out", "action_required", "neutral", "skipped"]
					.includes(String(value)) ? value as GitHubCheckEvidence["conclusion"] : "failure";
			}
			const updatedAt = timestamp(item.updated_at ?? item.created_at, "check update time");
			return {
				id: `${item.source}-${positiveInteger(item.id, "check ID")}`,
				name,
				producerId,
				sequence,
				status,
				conclusion,
				headSha,
				updatedAt,
				completedAt: status === "completed"
					? timestamp(item.completed_at ?? item.updated_at ?? item.created_at, "check completion time")
					: null,
			};
		});
		return { items, complete };
	}

	async #requestedChanges(repo: string, pullRequest: number, headSha: string, context: ExternalCallContext): Promise<{ items: GitHubRequestedChangeEvidence[]; complete: boolean }> {
		const result = await this.#paged(
			(page) => `/repos/${repo}/pulls/${pullRequest}/reviews?per_page=${PAGE_SIZE}&page=${page}`,
			context,
		);
		return {
			items: result.items.map((item) => record(item, "pull request review"))
				.filter((item) => item.state === "CHANGES_REQUESTED")
				.map((item) => ({
					id: String(positiveInteger(item.id, "requested-change review ID")),
					actor: text(record(item.user, "review actor").login, "review actor login", 64),
					commitSha: sha(item.commit_id ?? headSha, "requested-change commit SHA"),
					dismissed: false,
					submittedAt: timestamp(item.submitted_at, "requested-change time"),
				})),
			complete: result.complete,
		};
	}

	async #threads(repo: string, pullRequest: number, context: ExternalCallContext): Promise<{ items: GitHubReviewThreadEvidence[]; complete: boolean }> {
		const [owner, name] = repo.split("/");
		const query = `query($owner:String!,$name:String!,$number:Int!,$cursor:String){repository(owner:$owner,name:$name){pullRequest(number:$number){reviewThreads(first:100,after:$cursor){nodes{id isResolved comments(last:1){nodes{updatedAt originalCommit{oid}}}} pageInfo{hasNextPage endCursor}}}}}`;
		const items: GitHubReviewThreadEvidence[] = [];
		let cursor: string | null = null;
		for (let page = 1; page <= this.#maxPages; page += 1) {
			const fields = ["-f", `query=${query}`, "-F", `owner=${owner}`, "-F", `name=${name}`, "-F", `number=${pullRequest}`];
			if (cursor !== null) fields.push("-F", `cursor=${cursor}`);
			const payload = record(parseJson(await this.#gh(["api", "graphql", ...fields], context), "review threads", this.#maxOutputBytes), "GraphQL payload");
			const data = record(payload.data, "GraphQL data");
			const repositoryNode = record(data.repository, "GraphQL repository");
			const pullRequestNode = record(repositoryNode.pullRequest, "GraphQL pull request");
			const threads = record(pullRequestNode.reviewThreads, "GraphQL review threads");
			for (const item of array(threads.nodes, "review threads")) {
				const thread = record(item, "review thread");
				const comments = record(thread.comments, "review thread comments");
				const nodes = array(comments.nodes, "review thread comments");
				if (nodes.length !== 1) throw new Error("review thread lacks exact last-comment evidence");
				const comment = record(nodes[0], "review thread comment");
				const commit = record(comment.originalCommit, "review thread commit");
				items.push({
					id: text(thread.id, "review thread ID", 256),
					resolved: thread.isResolved === true,
					headSha: sha(commit.oid, "review thread head SHA"),
					updatedAt: timestamp(comment.updatedAt, "review thread update time"),
				});
			}
			const pageInfo = record(threads.pageInfo, "review thread page info");
			if (pageInfo.hasNextPage !== true) return { items, complete: true };
			cursor = text(pageInfo.endCursor, "review thread cursor", 512);
		}
		return { items, complete: false };
	}

	async #pullEvidence(rowValue: Record<string, unknown>, query: PullRequestMarkerQuery, context: ExternalCallContext): Promise<GitHubPullRequestEvidence> {
		const number = positiveInteger(rowValue.number, "pull request number");
		const row = record(await this.#api("GET", `/repos/${query.repository}/pulls/${number}`, context), "pull request");
		const metadata = await this.#metadata(query.repository, number, query.marker, context);
		const base = record(row.base, "pull request base");
		const head = record(row.head, "pull request head");
		const actualBaseSha = sha(base.sha, "pull request base SHA");
		const actualHeadSha = sha(head.sha, "pull request head SHA");
		const changedPathsResult = await this.#paged(
			(page) => `/repos/${query.repository}/pulls/${number}/files?per_page=${PAGE_SIZE}&page=${page}`,
			context,
		);
		const changedPaths = changedPathsResult.items.map((item) => text(record(item, "pull request file").filename, "changed path", 4_096)).sort();
		const reviews = await this.#reviews(query.repository, number, context);
		const checks = await this.#checks(query.repository, actualHeadSha, context);
		const requestedChanges = await this.#requestedChanges(query.repository, number, actualHeadSha, context);
		const threads = await this.#threads(query.repository, number, context);
		const observedAt = this.#now().toISOString();
		const updated = timestamp(row.updated_at, "pull request update time");
		const merged = row.merged === true || row.merged_at !== null;
		const mergeState = row.mergeable_state === "clean" ? "clean"
			: row.mergeable_state === "behind" ? "behind"
				: row.mergeable_state === "dirty" ? "conflicting"
					: row.mergeable_state === "blocked" ? "blocked" : "unknown";
		return {
			schemaVersion: 2,
			repository: query.repository,
			workItemId: metadata.workItemId,
			generation: metadata.generation,
			number,
			marker: metadata.marker,
			title: text(row.title, "pull request title", 256),
			body: text(row.body, "pull request body"),
			state: merged ? "merged" : row.state === "open" ? "open" : "closed",
			draft: row.draft === true,
			baseBranch: text(base.ref, "pull request base ref", 240),
			headBranch: text(head.ref, "pull request head ref", 240),
			baseSha: actualBaseSha,
			headSha: actualHeadSha,
			changedPathsComplete: changedPathsResult.complete,
			changedPaths,
			allowedScopes: metadata.allowedScopes,
			mergeState,
			policyDigest: metadata.policyDigest,
			checksComplete: checks.complete,
			checks: checks.items,
			requestedChangesComplete: requestedChanges.complete,
			requestedChanges: requestedChanges.items,
			threadsComplete: threads.complete,
			threads: threads.items,
			reviews: reviews.map((entry) => entry.review),
			reviewsComplete: true,
			dispositionsComplete: true,
			dispositions: reviews.flatMap((entry) => entry.dispositions),
			revision: revisionAt(updated),
			observedAt,
		};
	}

	async findPullRequests(query: PullRequestMarkerQuery, context: ExternalCallContext): Promise<AuthoritativeLookup<GitHubPullRequestEvidence>> {
		const canonical = { repository: repository(query.repository), marker: text(query.marker, "pull request marker", 512) };
		const result = await this.#findPullRows(canonical, context);
		return {
			items: await Promise.all(result.rows.map((row) => this.#pullEvidence(row, canonical, context))),
			complete: result.complete,
		};
	}

	async #postComment(repo: string, issue: number, body: string, context: ExternalCallContext): Promise<Record<string, unknown>> {
		return record(await this.#api("POST", `/repos/${repo}/issues/${issue}/comments`, context, [`body=${body}`]), "created comment");
	}

	async #ensurePrMetadata(request: CreatePullRequestRequest, pullRequest: number, context: ExternalCallContext): Promise<void> {
		const comments = await this.#issueComments(request.repository, pullRequest, context);
		const existing = comments.items.map((entry) => decodeEnvelope(entry.body, PR_META_PREFIX, "pull request metadata"))
			.filter((entry): entry is Record<string, unknown> => entry !== null && typeof entry === "object" && !Array.isArray(entry))
			.filter((entry) => entry.marker === request.marker);
		if (existing.length > 1) throw new Error("duplicate pull request metadata is ambiguous");
		if (existing.length === 1) return;
		const metadata: PullRequestMetadata = {
			workItemId: request.workItemId,
			generation: request.generation,
			marker: request.marker,
			baseSha: request.baseSha,
			headSha: request.headSha,
			changedPaths: [...request.changedPaths].sort(),
			allowedScopes: [...request.allowedScopes].sort(),
			policyDigest: request.policyDigest,
		};
		await this.#postComment(request.repository, pullRequest, encodedEnvelope(PR_META_PREFIX, metadata), context);
	}

	async createPullRequest(request: CreatePullRequestRequest, context: ExternalCallContext): Promise<DurableMutationResult<GitHubPullRequestEvidence>> {
		const existingRows = await this.#findPullRows(request, context);
		if (existingRows.rows.length > 1) throw new Error("duplicate pull request marker is ambiguous");
		if (existingRows.rows.length === 1) {
			const number = positiveInteger(existingRows.rows[0].number, "existing pull request number");
			await this.#ensurePrMetadata(request, number, context);
			const existing = await this.findPullRequests(request, context);
			if (existing.items.length !== 1) throw new Error("existing pull request is not authoritatively visible");
			return durableResult(request.mutation, existing.items[0].revision, false, existing.items[0]);
		}
		let number: number;
		try {
			const created = record(await this.#api("POST", `/repos/${repository(request.repository)}/pulls`, context, [
				`title=${request.title}`, `body=${request.body}`, `base=${request.baseBranch}`, `head=${request.headBranch}`,
				`draft=${request.draft ? "true" : "false"}`,
			]), "created pull request");
			number = positiveInteger(created.number, "created pull request number");
			await this.#ensurePrMetadata(request, number, context);
		} catch (error) {
			const rows = await this.#findPullRows(request, context);
			if (rows.rows.length !== 1) throw error;
			number = positiveInteger(rows.rows[0].number, "recovered pull request number");
			await this.#ensurePrMetadata(request, number, context);
			const recovered = await this.findPullRequests(request, context);
			if (recovered.items.length !== 1) throw error;
			return durableResult(request.mutation, recovered.items[0].revision, false, recovered.items[0]);
		}
		const found = await this.findPullRequests(request, context);
		const value = found.items.find((entry) => entry.number === number);
		if (value === undefined) throw new Error("created pull request is not authoritatively visible");
		return durableResult(request.mutation, value.revision, true, value);
	}

	async #searchMarker(repo: string, marker: string, context: ExternalCallContext): Promise<number[]> {
		const query = encodeURIComponent(`repo:${repo} in:comments \"${marker}\"`);
		const payload = record(await this.#api("GET", `/search/issues?q=${query}&per_page=${PAGE_SIZE}&page=1`, context), "search result");
		if (payload.incomplete_results === true) throw new Error("GitHub marker search is incomplete");
		return array(payload.items, "marker search items").map((entry) => positiveInteger(record(entry, "search item").number, "search issue number"));
	}

	async findParentRosters(query: GitHubRosterQuery, context: ExternalCallContext): Promise<AuthoritativeLookup<GitHubRosterSnapshot>> {
		const repo = repository(query.repository);
		const numbers = await this.#searchMarker(repo, query.marker, context);
		const results: GitHubRosterSnapshot[] = [];
		for (const number of numbers) {
			const comments = await this.#issueComments(repo, number, context);
			for (const comment of comments.items.filter((entry) => typeof entry.body === "string" && entry.body.split(query.marker).length === 2)) {
				const body = text(comment.body, "roster body");
				const commentId = positiveInteger(comment.id, "roster comment ID");
				const metadata = comments.items.map((entry) => decodeEnvelope(entry.body, ROSTER_META_PREFIX, "roster metadata"))
					.filter((entry): entry is Record<string, unknown> => entry !== null && typeof entry === "object" && !Array.isArray(entry))
					.filter((entry) => entry.marker === query.marker && entry.commentId === commentId);
				if (metadata.length > 1) throw new Error("parent roster metadata is ambiguous");
				if (metadata.length === 0) continue;
				const meta = metadata[0];
				const statuses = record(meta.statuses, "roster statuses") as Record<string, "pending" | "running" | "succeeded" | "failed" | "blocked">;
				results.push({
					id: commentId,
					marker: query.marker,
					parentIssue: positiveInteger(meta.parentIssue, "roster parent issue"),
					generation: positiveInteger(meta.generation, "roster generation"),
					body,
					statuses,
					statusEpoch: positiveInteger(meta.statusEpoch, "roster status epoch"),
					revision: positiveInteger(meta.revision, "roster revision"),
					updatedAt: timestamp(meta.updatedAt, "roster update time"),
				});
			}
		}
		return { items: results, complete: true };
	}

	async publishParentRoster(request: PublishRosterRequest, context: ExternalCallContext): Promise<DurableMutationResult<GitHubRosterSnapshot>> {
		const existing = await this.findParentRosters(request, context);
		if (existing.items.length > 1) throw new Error("duplicate parent roster marker is ambiguous");
		let comment: Record<string, unknown>;
		let metadataComment: Record<string, unknown> | undefined;
		if (existing.items.length === 0) {
			const rawComments = await this.#issueComments(request.repository, request.parentIssue, context);
			const orphan = rawComments.items.filter((entry) => entry.body === request.body);
			if (orphan.length > 1) throw new Error("duplicate parent roster publication is ambiguous");
			comment = orphan[0] ?? await this.#postComment(request.repository, request.parentIssue, request.body, context);
		} else {
			if (request.mutation.expectedResourceRevision !== existing.items[0].revision) throw new Error("parent roster CAS revision conflict");
			comment = record(await this.#api("PATCH", `/repos/${request.repository}/issues/comments/${existing.items[0].id}`, context, [`body=${request.body}`]), "updated roster comment");
			const comments = await this.#issueComments(request.repository, request.parentIssue, context);
			const metadataMatches = comments.items.filter((entry) => {
				const value = decodeEnvelope(entry.body, ROSTER_META_PREFIX, "roster metadata");
				return typeof value === "object" && value !== null && !Array.isArray(value)
					&& (value as Record<string, unknown>).marker === request.marker
					&& (value as Record<string, unknown>).commentId === existing.items[0].id;
			});
			if (metadataMatches.length !== 1) throw new Error("parent roster metadata is absent or ambiguous");
			metadataComment = metadataMatches[0];
		}
		const updatedAt = this.#now().toISOString();
		const revision = Math.max((existing.items[0]?.revision ?? 0) + 1, revisionAt(this.#now()));
		if (revision > MAX_GITHUB_REVISION) throw new Error("parent roster revision space is exhausted");
		const value: GitHubRosterSnapshot = {
			id: positiveInteger(comment.id, "roster comment ID"),
			marker: request.marker,
			parentIssue: request.parentIssue,
			generation: request.generation,
			body: request.body,
			statuses: { ...request.statuses },
			statusEpoch: request.statusEpoch,
			revision,
			updatedAt,
		};
		const metadataBody = encodedEnvelope(ROSTER_META_PREFIX, {
			marker: request.marker,
			commentId: value.id,
			parentIssue: request.parentIssue,
			generation: request.generation,
			statuses: value.statuses,
			statusEpoch: request.statusEpoch,
			revision,
			updatedAt,
		});
		if (metadataComment === undefined) {
			await this.#postComment(request.repository, request.parentIssue, metadataBody, context);
		} else {
			await this.#api("PATCH", `/repos/${request.repository}/issues/comments/${positiveInteger(metadataComment.id, "roster metadata comment ID")}`, context, [`body=${metadataBody}`]);
		}
		return durableResult(request.mutation, value.revision, true, value);
	}

	async findChildIntegration(query: ChildIntegrationQuery, context: ExternalCallContext): Promise<AuthoritativeLookup<ChildIntegrationReceipt>> {
		const rows = await this.#findPullRows({ repository: query.repository, marker: query.marker }, context);
		const receipts: ChildIntegrationReceipt[] = [];
		for (const row of rows.rows) {
			const comments = await this.#issueComments(query.repository, positiveInteger(row.number, "pull request number"), context);
			for (const comment of comments.items) {
				const value = decodeEnvelope(comment.body, INTEGRATION_PREFIX, "integration receipt");
				if (typeof value === "object" && value !== null && !Array.isArray(value)
					&& (value as Record<string, unknown>).childId === query.childId) receipts.push(value as ChildIntegrationReceipt);
			}
		}
		return { items: receipts, complete: rows.complete };
	}

	async integrateChild(request: IntegrateChildRequest, context: ExternalCallContext): Promise<DurableMutationResult<ChildIntegrationReceipt>> {
		if (typeof request.parentBaseBranch !== "string" || request.parentBaseBranch.length === 0
			|| typeof request.parentBranch !== "string" || request.parentBranch.length === 0) {
			throw new Error("child integration transport requires authoritative parent/default branch evidence");
		}
		if (request.parentBranch === request.parentBaseBranch) {
			throw new Error("child integration transport refuses default-branch and parent-to-main merges");
		}
		if (["main", "master", "trunk"].includes(request.parentBranch)) {
			throw new Error("child integration transport refuses conventional protected branch aliases");
		}
		const integration = validateReviewedChildIntegrationEvidence(request.integration);
		if (integration.parentBranch !== request.parentBranch
			|| integration.baseSha !== request.baseSha || integration.headSha !== request.headSha
			|| integration.mergeCommitSha === integration.baseSha
			|| integration.mergeCommitSha === integration.headSha) {
			throw new Error("child integration Git evidence does not match the exact request");
		}
		const existing = await this.findChildIntegration(request, context);
		if (existing.items.length > 1) throw new Error("duplicate child integration receipt is ambiguous");
		if (existing.items.length === 1) {
			return durableResult(request.mutation, existing.items[0].transportProvenance.revision, false, existing.items[0]);
		}
		const repo = repository(request.repository);
		const repositoryEvidence = record(await this.#api("GET", `/repos/${repo}`, context), "repository");
		const liveDefaultBranch = text(repositoryEvidence.default_branch, "repository default branch", 240);
		if (liveDefaultBranch !== request.parentBaseBranch || request.parentBranch === liveDefaultBranch) {
			throw new Error("child integration remote default branch moved before publication");
		}
		const pr = record(await this.#api("GET", `/repos/${repo}/pulls/${request.pullRequest}`, context), "child pull request");
		const base = record(pr.base, "child pull request base");
		const head = record(pr.head, "child pull request head");
		if (base.ref !== request.parentBranch || head.sha !== request.headSha) throw new Error("child integration exact branch/head moved");
		if (pr.merged !== true || sha(pr.merge_commit_sha, "child pull request merge commit SHA") !== integration.mergeCommitSha) {
			throw new Error("child integration is not authoritatively merged at the exact Git integration commit");
		}
		const parentRef = record(await this.#api(
			"GET",
			`/repos/${repo}/git/ref/heads/${encodeURIComponent(request.parentBranch)}`,
			context,
		), "parent branch ref");
		const parentObject = record(parentRef.object, "parent branch ref object");
		if (parentRef.ref !== `refs/heads/${request.parentBranch}`
			|| parentObject.type !== "commit" || parentObject.sha !== integration.parentHead) {
			throw new Error("child integration parent ref does not match lease-bound Git evidence");
		}
		const pendingBody = `${INTEGRATION_PENDING_PREFIX}${Buffer.from(request.mutation.idempotencyKey).toString("base64url")} -->`;
		let pending = (await this.#issueComments(request.repository, request.pullRequest, context)).items
			.find((entry) => entry.body === pendingBody);
		if (pending === undefined) pending = await this.#postComment(request.repository, request.pullRequest, pendingBody, context);
		const commentId = positiveInteger(pending.id, "integration comment ID");
		const revision = stableResourceRevision(commentId);
		const integratedAt = this.#now().toISOString();
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
			integratedAt,
		};
		await this.#api("PATCH", `/repos/${request.repository}/issues/comments/${commentId}`, context, [
			`body=${encodedEnvelope(INTEGRATION_PREFIX, receipt)}`,
		]);
		return durableResult(request.mutation, revision, true, receipt);
	}

	async proveAncestry(query: GitAncestryQuery, context: ExternalCallContext): Promise<GitAncestryProof> {
		const repo = repository(query.repository);
		const ancestor = sha(query.ancestorSha, "ancestry ancestor SHA");
		const descendant = sha(query.descendantSha, "ancestry descendant SHA");
		const comparison = record(await this.#api("GET", `/repos/${repo}/compare/${ancestor}...${descendant}`, context), "ancestry comparison");
		const result = comparison.status === "ahead" || comparison.status === "identical";
		return {
			schemaVersion: 1,
			authority: "transport",
			repository: repo,
			ancestorSha: ancestor,
			descendantSha: descendant,
			result,
			revision: revisionAt(this.#now()),
			observedAt: this.#now().toISOString(),
		};
	}
}
