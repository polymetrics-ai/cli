import { execFile } from "node:child_process";
import { createHash } from "node:crypto";
import { isAbsolute, resolve } from "node:path";

import {
	validateProductionPlanningIssueFacts,
	type ProductionPlanningCallContext,
	type ProductionPlanningIssueFacts,
	type ProductionPlanningIssueSource,
} from "./production-plan-bootstrap.ts";

const DEFAULT_TIMEOUT_MS = 30_000;
const DEFAULT_MAX_OUTPUT_BYTES = 2 * 1024 * 1024;
const DEFAULT_MAX_PAGES = 10;
const PAGE_SIZE = 100;
const MAX_SUBISSUES = 64;
const API_VERSION_HEADER = "X-GitHub-Api-Version: 2022-11-28";

export interface ProductionPlanningIssueExecutionOptions {
	cwd: string;
	signal: AbortSignal;
	timeoutMs: number;
	maxOutputBytes: number;
}

export type ProductionPlanningIssueExecutor = (
	args: readonly string[],
	options: ProductionPlanningIssueExecutionOptions,
) => Promise<string>;

export interface GhProductionPlanningIssueSourceOptions {
	execute?: ProductionPlanningIssueExecutor;
	now?: () => Date;
	timeoutMs?: number;
	maxOutputBytes?: number;
	maxPages?: number;
}

function positive(value: unknown, fallback: number, maximum: number, description: string): number {
	const result = value ?? fallback;
	if (!Number.isSafeInteger(result) || (result as number) < 1 || (result as number) > maximum) {
		throw new Error(`${description} is invalid`);
	}
	return result as number;
}

function record(value: unknown, description: string): Record<string, unknown> {
	if (typeof value !== "object" || value === null || Array.isArray(value)) {
		throw new Error(`GitHub returned malformed ${description}`);
	}
	return value as Record<string, unknown>;
}

function text(value: unknown, description: string, maximum: number, allowEmpty = false): string {
	if (typeof value !== "string" || (!allowEmpty && value.length < 1) || Buffer.byteLength(value) > maximum
		|| /[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f-\u009f]/u.test(value)) {
		throw new Error(`GitHub returned invalid ${description}`);
	}
	return value;
}

function timestamp(value: unknown, description: string): string {
	const date = new Date(text(value, description, 128));
	if (!Number.isFinite(date.valueOf())) throw new Error(`GitHub returned invalid ${description}`);
	return date.toISOString();
}

function issueNumber(value: unknown, description: string): number {
	if (!Number.isSafeInteger(value) || (value as number) < 1) throw new Error(`GitHub returned invalid ${description}`);
	return value as number;
}

function parse(output: string, description: string, maximum: number): unknown {
	if (Buffer.byteLength(output) > maximum) throw new Error(`GitHub ${description} output exceeded its bound`);
	try { return JSON.parse(output); }
	catch { throw new Error(`GitHub returned malformed ${description} JSON`); }
}

function canonical(value: unknown): unknown {
	if (Array.isArray(value)) return value.map(canonical);
	if (value !== null && typeof value === "object") {
		return Object.fromEntries(Object.keys(value as Record<string, unknown>).sort()
			.map((key) => [key, canonical((value as Record<string, unknown>)[key])]));
	}
	return value;
}

export const defaultProductionPlanningIssueExecutor: ProductionPlanningIssueExecutor = (
	args,
	options,
) => new Promise((resolveCall, rejectCall) => {
	execFile("gh", [...args], {
		cwd: options.cwd,
		encoding: "utf8",
		env: process.env,
		maxBuffer: options.maxOutputBytes,
		timeout: options.timeoutMs,
		killSignal: "SIGTERM",
		signal: options.signal,
		windowsHide: true,
	}, (error, stdout) => {
		if (error !== null) {
			rejectCall(Object.assign(new Error("bounded GitHub planning read failed"), {
				code: (error as NodeJS.ErrnoException).code,
				killed: (error as { killed?: boolean }).killed,
			}));
			return;
		}
		resolveCall(stdout);
	});
});

export class GhProductionPlanningIssueSource implements ProductionPlanningIssueSource {
	readonly #execute: ProductionPlanningIssueExecutor;
	readonly #now: () => Date;
	readonly #timeoutMs: number;
	readonly #maxOutputBytes: number;
	readonly #maxPages: number;

	constructor(options: GhProductionPlanningIssueSourceOptions = {}) {
		if (typeof options !== "object" || options === null) throw new Error("GitHub planning source options are invalid");
		this.#execute = options.execute ?? defaultProductionPlanningIssueExecutor;
		this.#now = options.now ?? (() => new Date());
		this.#timeoutMs = positive(options.timeoutMs, DEFAULT_TIMEOUT_MS, 120_000, "GitHub planning timeout");
		this.#maxOutputBytes = positive(
			options.maxOutputBytes,
			DEFAULT_MAX_OUTPUT_BYTES,
			8 * 1024 * 1024,
			"GitHub planning output bound",
		);
		this.#maxPages = positive(options.maxPages, DEFAULT_MAX_PAGES, 20, "GitHub planning page bound");
	}

	async observe(
		query: { repositoryRoot: string; parentIssue: number },
		context: ProductionPlanningCallContext,
	): Promise<ProductionPlanningIssueFacts> {
		if (typeof query !== "object" || query === null || !isAbsolute(query.repositoryRoot)
			|| !Number.isSafeInteger(query.parentIssue) || query.parentIssue < 1) {
			throw new Error("GitHub planning query is invalid");
		}
		this.#assertActive(context);
		const repositoryRoot = resolve(query.repositoryRoot);
		const repository = record(parse(await this.#call([
			"repo", "view", "--json", "nameWithOwner,defaultBranchRef,viewerPermission",
		], repositoryRoot, context), "repository", this.#maxOutputBytes), "repository");
		const nameWithOwner = text(repository.nameWithOwner, "repository name", 201);
		const defaultBranch = record(repository.defaultBranchRef, "default branch");
		const permission = text(repository.viewerPermission, "viewer permission", 32).toLowerCase();
		if (permission !== "admin" && permission !== "maintain") {
			throw new Error("GitHub planning requires an authenticated admin or maintainer");
		}
		const viewer = record(parse(await this.#call([
			"api", "--method", "GET", "-H", API_VERSION_HEADER, "/user",
		], repositoryRoot, context), "viewer", this.#maxOutputBytes), "viewer");
		const parentRaw = record(parse(await this.#call([
			"api", "--method", "GET", "-H", API_VERSION_HEADER,
			`/repos/${nameWithOwner}/issues/${query.parentIssue}`,
		], repositoryRoot, context), "parent issue", this.#maxOutputBytes), "parent issue");
		if (Object.hasOwn(parentRaw, "pull_request")) throw new Error("GitHub planning parent is a pull request, not an issue");
		if (parentRaw.state !== "open") throw new Error("GitHub planning parent issue must be open");
		if (parentRaw.number !== query.parentIssue) throw new Error("GitHub planning parent issue number moved");
		const parent = {
			number: issueNumber(parentRaw.number, "parent issue number"),
			nodeId: text(parentRaw.node_id, "parent issue node ID", 512),
			title: text(parentRaw.title, "parent issue title", 256),
			body: parentRaw.body === null ? "" : text(parentRaw.body, "parent issue body", 64 * 1024, true),
			state: "open" as const,
			updatedAt: timestamp(parentRaw.updated_at, "parent issue update time"),
		};
		const subissues: ProductionPlanningIssueFacts["subissues"] = [];
		for (let page = 1; page <= this.#maxPages; page += 1) {
			const raw = parse(await this.#call([
				"api", "--method", "GET", "-H", API_VERSION_HEADER,
				`/repos/${nameWithOwner}/issues/${query.parentIssue}/sub_issues?per_page=${PAGE_SIZE}&page=${page}`,
			], repositoryRoot, context), "subissues", this.#maxOutputBytes);
			if (!Array.isArray(raw) || raw.length > PAGE_SIZE) throw new Error("GitHub returned malformed subissues page");
			for (const entry of raw) {
				const issue = record(entry, "subissue");
				if (Object.hasOwn(issue, "pull_request")) throw new Error("GitHub planning subissue cannot be a pull request");
				if (subissues.length >= MAX_SUBISSUES) throw new Error("GitHub planning has more than 64 subissues");
				subissues.push({
					number: issueNumber(issue.number, "subissue number"),
					title: text(issue.title, "subissue title", 256),
					body: issue.body === null ? "" : text(issue.body, "subissue body", 64 * 1024, true),
					state: issue.state === "open" || issue.state === "closed"
						? issue.state : (() => { throw new Error("GitHub returned invalid subissue state"); })(),
					updatedAt: timestamp(issue.updated_at, "subissue update time"),
				});
			}
			if (raw.length < PAGE_SIZE) break;
			if (page === this.#maxPages) throw new Error("GitHub subissue pagination is incomplete");
		}
		const source = {
			repository: nameWithOwner,
			defaultBranch: text(defaultBranch.name, "default branch name", 240),
			viewer: { login: text(viewer.login, "viewer login", 39).toLowerCase(), permission },
			parent,
			subissues,
		};
		const revisionDigest = createHash("sha256")
			.update(JSON.stringify(canonical(source)))
			.digest("hex");
		return validateProductionPlanningIssueFacts({
			schemaVersion: 1,
			...source,
			complete: true,
			revisionDigest,
			observedAt: this.#now().toISOString(),
		}, query.parentIssue);
	}

	#assertActive(context: ProductionPlanningCallContext): void {
		if (typeof context !== "object" || context === null || !(context.signal instanceof AbortSignal)) {
			throw new Error("GitHub planning context is invalid");
		}
		if (context.signal.aborted) throw context.signal.reason ?? new Error("GitHub planning read cancelled");
		const deadline = new Date(context.deadlineAt);
		if (!Number.isFinite(deadline.valueOf()) || deadline.valueOf() <= Date.now()) {
			throw new Error("GitHub planning deadline expired");
		}
	}

	async #call(
		args: readonly string[],
		cwd: string,
		context: ProductionPlanningCallContext,
	): Promise<string> {
		this.#assertActive(context);
		const deadline = new Date(context.deadlineAt).valueOf();
		return this.#execute(args, {
			cwd,
			signal: context.signal,
			timeoutMs: Math.max(1, Math.min(this.#timeoutMs, deadline - Date.now())),
			maxOutputBytes: this.#maxOutputBytes,
		});
	}
}
