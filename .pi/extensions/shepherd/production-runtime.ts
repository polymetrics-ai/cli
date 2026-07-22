import { createHash } from "node:crypto";
import { execFile } from "node:child_process";
import { isAbsolute, join, resolve } from "node:path";

import type { ProductionEffectRecord } from "./autonomous-production-contract.ts";
import { ProductionEffectJournal, type ProductionEffectJournalPort } from "./autonomous-effect-journal.ts";
import { ProductionRecoveryBarrier, type ProductionEffectRecoveryPort } from "./autonomous-recovery.ts";
import {
	ProductionFileStateStore,
	type ProductionStateStore,
} from "./autonomous-production-state.ts";
import {
	ProductionWorkspaceLifecycle,
	type ProductionAgentSessionPort,
} from "./production-workspace-lifecycle.ts";
import { ProductionRepositoryPlanIntake } from "./production-intake.ts";
import {
	ProductionShepherdController,
	type ProductionParentFinalizerPort,
	type ProductionParentGatePort,
	type ProductionPlanIntakePort,
} from "./production-controller.ts";
import {
	ProductionChildPipeline,
	type ProductionChildGitHubPort,
	type ProductionChildIntegrationReceiptAuthority,
	type ProductionExactHeadReviewPort,
	type ProductionParentHeadObservation,
	type ProductionParentHeadSource,
	type ProductionWorkspaceLifecyclePort,
} from "./production-child-pipeline.ts";
import { BoundedVerificationRunner } from "./bounded-verification.ts";
import { GitAdapter, type GitBinding } from "./git-adapter.ts";
import { WorkspaceAdapter } from "./workspace-adapter.ts";
import {
	createProductionGitHubOrchestrationFacade,
	defaultGhOrchestrationExecutor,
	type GhCliOrchestrationTransportOptions,
	type GhOrchestrationExecutor,
} from "./gh-orchestration-transport.ts";
import {
	adaptGitHubDecisionBroker,
	GitHubParentOrchestrator,
	type ExternalCallContext,
	type ParentDecisionBroker,
	type ParentOrchestrationPolicyAuthority,
	type ParentReadyDurableAuthorityBoundary,
	type RequiredCheckPolicySource,
} from "./github-orchestrator.ts";
import {
	GhCliDecisionTransport,
	GitHubDecisionBroker,
	type GitHubDecisionBrokerOptions,
} from "./github-decision-broker.ts";
import { FileHumanDecisionRepository } from "./human-decision.ts";
import {
	AgentSessionProductionReviewAdapter,
	GhProductionReviewRepository,
	type ProductionChangedPathEvidenceSource,
	type ProductionReviewRepository,
	type ProductionReviewRoleRequestFactory,
	type ProductionReviewSession,
} from "./production-review-adapter.ts";
import {
	ProductionParentFinalizer,
	ProductionParentGateAdapter,
	type ProductionParentCheckPolicyAuthority,
	type ProductionParentReadyTransitionPort,
} from "./production-parent-lifecycle.ts";
import { GhParentPullRequestMergeLookup } from "./production-human-gate.ts";
import {
	validateGitHubChangedPathEvidence,
	validateRequiredGitHubCheckPolicy,
	type GitHubChangedPathEvidence,
	type RequiredGitHubCheckPolicy,
} from "./github-evidence.ts";
import type { RoleRunRequest } from "./agent-session-runtime.ts";
import type { IndependentReviewTarget, IndependentReviewWork } from "./review-router.ts";
import { validateScopedPath, type ScopedWorkspace } from "./tool-policy.ts";

const SHA = /^[0-9a-f]{40}$/u;
const REPOSITORY = /^[A-Za-z0-9][A-Za-z0-9._-]{0,99}\/[A-Za-z0-9][A-Za-z0-9._-]{0,99}$/u;
const SAFE_BRANCH = /^(?!\/|.*(?:\.\.|\s|[~^:?*\\\[\]])|.*\/$)[A-Za-z0-9][A-Za-z0-9._\/-]{0,239}$/u;
const SAFE_ACTOR = /^[A-Za-z0-9](?:[A-Za-z0-9._-]*[A-Za-z0-9])?$/u;
const MAX_GITHUB_OUTPUT_BYTES = 2 * 1024 * 1024;
const MAX_GITHUB_PAGES = 10;
const MAX_REVIEW_OBJECT_BYTES = 256 * 1024;
const DEFAULT_EXTERNAL_TIMEOUT_MS = 15_000;
const NULL_DEVICE = process.platform === "win32" ? "NUL" : "/dev/null";

export interface ProductionControllerCompositionOptions {
	stateStore: ProductionStateStore;
	intake: ProductionPlanIntakePort;
	effects: ProductionEffectJournalPort;
	effectRecovery: ProductionEffectRecoveryPort;
	workspaceLifecycle: ProductionWorkspaceLifecyclePort;
	github: ProductionChildGitHubPort;
	reviewer: ProductionExactHeadReviewPort;
	reviewRepository: ProductionReviewRepository;
	decisionBroker: ParentDecisionBroker;
	parentHeads: ProductionParentHeadSource;
	coordinator: GitBinding;
	trustedWorktreeRoot: string;
	dispositionActor: string;
	receiptAuthority?: ProductionChildIntegrationReceiptAuthority;
	finalizer: ProductionParentFinalizerPort;
	parentGate: ProductionParentGatePort;
	now?: () => Date;
	newRunId?: () => string;
}

export interface ProductionShepherdRuntimeOptions {
	repositoryRoot: string;
	stateRoot: string;
	trustedWorktreeRoot: string;
	coordinator: GitBinding;
	/** The exact GitAdapter instance that inspected and issued the coordinator binding. */
	git: GitAdapter;
	agentSession: ProductionAgentSessionPort;
	reviewSession: ProductionReviewSession;
	/** Mandatory authoritative replay/projector. A rejecting placeholder is not a production runtime. */
	effectRecovery: ProductionEffectRecoveryPort;
	/** Durable authority around the uncertain existing-parent-PR ready transition. */
	parentReadyAuthority: ParentReadyDurableAuthorityBoundary;
	/** Bounded adapter that changes only an existing parent draft PR to ready. It cannot merge. */
	parentReadiness: ProductionParentReadyTransitionPort;
	dispositionActor: string;
	github?: GhCliOrchestrationTransportOptions;
	decision?: GitHubDecisionBrokerOptions;
	now?: () => Date;
}

export interface ProductionGitObjectReadRequest {
	cwd: string;
	headSha: string;
	path: string;
	timeoutMs: number;
	maxOutputBytes: number;
	signal: AbortSignal;
}

export type ProductionGitObjectReader = (request: ProductionGitObjectReadRequest) => Promise<Buffer>;

export interface ExactHeadReviewRoleRequestFactoryOptions {
	git: Pick<GitAdapter, "assertBinding" | "resolveBranchHead">;
	coordinator: GitBinding;
	parentIssue: number;
	readObject?: ProductionGitObjectReader;
	timeoutMs?: number;
	maxObjectBytes?: number;
	now?: () => Date;
}

function safeAbsolutePath(value: unknown, description: string): string {
	if (typeof value !== "string" || !isAbsolute(value) || value.length > 4_096
		|| /[\u0000-\u001f\u007f]/u.test(value)) {
		throw new Error(`${description} must be an absolute bounded safe path`);
	}
	return resolve(value);
}

function boundedPositive(value: unknown, fallback: number, maximum: number, description: string): number {
	const candidate = value ?? fallback;
	if (!Number.isSafeInteger(candidate) || (candidate as number) < 1 || (candidate as number) > maximum) {
		throw new Error(`${description} must be a bounded positive integer`);
	}
	return candidate as number;
}

function canonicalTimestamp(value: unknown, description: string): string {
	if (typeof value !== "string" || value.length > 64) throw new Error(`invalid ${description}`);
	const parsed = new Date(value);
	if (!Number.isFinite(parsed.valueOf())) throw new Error(`invalid ${description}`);
	return parsed.toISOString();
}

function exactRecord(value: unknown, description: string): Record<string, unknown> {
	if (typeof value !== "object" || value === null || Array.isArray(value)) {
		throw new Error(`GitHub returned malformed ${description}`);
	}
	return value as Record<string, unknown>;
}

function parseGitHubJson(value: string, description: string, maximum = MAX_GITHUB_OUTPUT_BYTES): unknown {
	if (Buffer.byteLength(value) > maximum) throw new Error(`GitHub ${description} output exceeded its bound`);
	try {
		return JSON.parse(value);
	} catch {
		throw new Error(`GitHub returned malformed ${description} JSON`);
	}
}

function assertProductionRecovery(value: unknown): asserts value is ProductionEffectRecoveryPort {
	if (typeof value !== "object" || value === null
		|| typeof (value as ProductionEffectRecoveryPort).observe !== "function"
		|| typeof (value as ProductionEffectRecoveryPort).apply !== "function") {
		throw new Error("an authoritative production effect recovery authority is required");
	}
}

function validateComposition(options: ProductionControllerCompositionOptions): void {
	assertProductionRecovery(options.effectRecovery);
	safeAbsolutePath(options.trustedWorktreeRoot, "trusted worktree root");
	if (typeof options.dispositionActor !== "string" || !SAFE_ACTOR.test(options.dispositionActor)) {
		throw new Error("production review disposition actor is invalid");
	}
}

/**
 * Pure synchronous wiring seam. Tests inject deterministic ports here; production callers use
 * createProductionShepherdController, which supplies the concrete durable adapters below.
 */
export function composeProductionShepherdController(
	options: ProductionControllerCompositionOptions,
): ProductionShepherdController {
	validateComposition(options);
	const pipeline = new ProductionChildPipeline({
		workspaceLifecycle: options.workspaceLifecycle,
		github: options.github,
		reviewer: options.reviewer,
		reviewRepository: options.reviewRepository,
		effects: options.effects,
		decisionBroker: options.decisionBroker,
		parentHeads: options.parentHeads,
		coordinator: options.coordinator,
		trustedWorktreeRoot: options.trustedWorktreeRoot,
		dispositionActor: options.dispositionActor,
		recovery: options.effectRecovery,
		...(options.receiptAuthority === undefined ? {} : { receiptAuthority: options.receiptAuthority }),
		...(options.now === undefined ? {} : { now: options.now }),
	});
	return new ProductionShepherdController({
		stateStore: options.stateStore,
		intake: options.intake,
		recovery: new ProductionRecoveryBarrier(options.effects, pipeline),
		pipeline,
		finalizer: options.finalizer,
		parentGate: options.parentGate,
		...(options.now === undefined ? {} : { now: options.now }),
		...(options.newRunId === undefined ? {} : { newRunId: options.newRunId }),
	});
}

class ProductionParentPolicyAuthority implements ProductionParentCheckPolicyAuthority {
	readonly #source: RequiredCheckPolicySource;

	constructor(source: RequiredCheckPolicySource) {
		this.#source = source;
	}

	async findRequiredCheckPolicies(
		query: {
			repository: string;
			parentIssue: number;
			generation: number;
			parentBranch: string;
			parentBaseBranch: string;
		},
		context: ExternalCallContext,
	): Promise<{ items: RequiredGitHubCheckPolicy[]; complete: boolean }> {
		const findBundle = this.#source.findParentOrchestrationPolicyBundle;
		if (findBundle === undefined) throw new Error("production parent policy bundle authority is unavailable");
		const lookup = await findBundle.call(this.#source, query, context);
		if (lookup.complete !== true || lookup.items.length !== 1) {
			throw new Error("production parent policy authority is incomplete or ambiguous");
		}
		const authority: ParentOrchestrationPolicyAuthority = lookup.items[0];
		if (authority.repository !== query.repository || authority.parentIssue !== query.parentIssue
			|| authority.generation !== query.generation || authority.parentBranch !== query.parentBranch
			|| authority.parentBaseBranch !== query.parentBaseBranch) {
			throw new Error("production parent policy authority moved from the exact plan");
		}
		return {
			items: authority.policyBundle.requiredCheckPolicies.map(validateRequiredGitHubCheckPolicy),
			complete: true,
		};
	}
}

class GhProductionParentHeadSource implements ProductionParentHeadSource {
	readonly #execute: GhOrchestrationExecutor;
	readonly #timeoutMs: number;
	readonly #maxOutputBytes: number;

	constructor(options: GhCliOrchestrationTransportOptions) {
		this.#execute = options.execute ?? defaultGhOrchestrationExecutor;
		this.#timeoutMs = boundedPositive(options.timeoutMs, DEFAULT_EXTERNAL_TIMEOUT_MS, 120_000, "GitHub timeout");
		this.#maxOutputBytes = boundedPositive(options.maxOutputBytes, MAX_GITHUB_OUTPUT_BYTES, 8 * 1024 * 1024, "GitHub output limit");
	}

	async observe(
		plan: { repository: string; parentBranch: string },
		signal: AbortSignal,
	): Promise<ProductionParentHeadObservation> {
		if (!REPOSITORY.test(plan.repository) || !SAFE_BRANCH.test(plan.parentBranch)) {
			throw new Error("production parent head coordinates are invalid");
		}
		const output = await this.#execute("gh", [
			"api", "--method", "GET",
			`/repos/${plan.repository}/git/ref/heads/${plan.parentBranch.split("/").map(encodeURIComponent).join("/")}`,
		], { signal, timeoutMs: this.#timeoutMs, maxOutputBytes: this.#maxOutputBytes });
		const record = exactRecord(parseGitHubJson(output, "parent branch", this.#maxOutputBytes), "parent branch");
		const object = exactRecord(record.object, "parent branch object");
		if (record.ref !== `refs/heads/${plan.parentBranch}` || typeof object.sha !== "string" || !SHA.test(object.sha)) {
			throw new Error("GitHub parent branch observation is stale or malformed");
		}
		return { repository: plan.repository, branch: plan.parentBranch, head: object.sha };
	}
}

class GhProductionChangedPathSource implements ProductionChangedPathEvidenceSource {
	readonly #execute: GhOrchestrationExecutor;
	readonly #now: () => Date;
	readonly #timeoutMs: number;
	readonly #maxOutputBytes: number;
	readonly #maxPages: number;

	constructor(options: GhCliOrchestrationTransportOptions) {
		this.#execute = options.execute ?? defaultGhOrchestrationExecutor;
		this.#now = options.now ?? (() => new Date());
		this.#timeoutMs = boundedPositive(options.timeoutMs, DEFAULT_EXTERNAL_TIMEOUT_MS, 120_000, "GitHub timeout");
		this.#maxOutputBytes = boundedPositive(options.maxOutputBytes, MAX_GITHUB_OUTPUT_BYTES, 8 * 1024 * 1024, "GitHub output limit");
		this.#maxPages = boundedPositive(options.maxPages, MAX_GITHUB_PAGES, 100, "GitHub page limit");
	}

	async #get(endpoint: string, context: ExternalCallContext): Promise<unknown> {
		const deadline = new Date(context.deadlineAt).valueOf();
		if (!Number.isFinite(deadline) || deadline <= Date.now()) throw new Error("GitHub changed-path deadline expired");
		try {
			const output = await this.#execute("gh", ["api", "--method", "GET", endpoint], {
				signal: context.signal,
				timeoutMs: Math.max(1, Math.min(this.#timeoutMs, deadline - Date.now())),
				maxOutputBytes: this.#maxOutputBytes,
			});
			return parseGitHubJson(output, "changed paths", this.#maxOutputBytes);
		} finally {
			if (context.signal.aborted) context.acknowledgeAbort();
		}
	}

	async findChangedPathEvidence(
		query: Omit<IndependentReviewTarget, "changedPaths" | "allowedScopes">,
		context: ExternalCallContext,
	): Promise<{ items: GitHubChangedPathEvidence[]; complete: boolean }> {
		if (!REPOSITORY.test(query.repository) || !Number.isSafeInteger(query.pullRequest) || query.pullRequest < 1
			|| !Number.isSafeInteger(query.generation) || query.generation < 1
			|| !SHA.test(query.baseSha) || !SHA.test(query.headSha)) {
			throw new Error("production changed-path query is invalid");
		}
		const pull = exactRecord(
			await this.#get(`/repos/${query.repository}/pulls/${query.pullRequest}`, context),
			"pull request",
		);
		const base = exactRecord(pull.base, "pull request base");
		const head = exactRecord(pull.head, "pull request head");
		if (pull.number !== query.pullRequest || base.sha !== query.baseSha || head.sha !== query.headSha) {
			throw new Error("authoritative pull request moved during changed-path observation");
		}
		const paths: string[] = [];
		let complete = false;
		for (let page = 1; page <= this.#maxPages; page += 1) {
			const value = await this.#get(
				`/repos/${query.repository}/pulls/${query.pullRequest}/files?per_page=100&page=${page}`,
				context,
			);
			if (!Array.isArray(value) || value.length > 100) throw new Error("GitHub returned malformed changed paths");
			for (const item of value) {
				const row = exactRecord(item, "changed path");
				paths.push(validateScopedPath(String(row.filename), ["."]));
			}
			if (value.length < 100) { complete = true; break; }
		}
		if (!complete || new Set(paths).size !== paths.length) {
			return { items: [], complete: false };
		}
		const updatedAt = canonicalTimestamp(pull.updated_at, "pull request update time");
		const revision = Math.max(1, Math.floor(new Date(updatedAt).valueOf() / 1_000));
		return {
			items: [validateGitHubChangedPathEvidence({
				schemaVersion: 1,
				authority: "controller",
				repository: query.repository,
				workItemId: query.workItemId,
				pullRequest: query.pullRequest,
				generation: query.generation,
				baseSha: query.baseSha,
				headSha: query.headSha,
				paths: paths.sort(),
				complete: true,
				revision,
				observedAt: this.#now().toISOString(),
			})],
			complete: true,
		};
	}
}

function decisionTransport(
	execute: GhOrchestrationExecutor | undefined,
): GhCliDecisionTransport {
	if (execute === undefined) return new GhCliDecisionTransport();
	return new GhCliDecisionTransport(async (_file, args) => execute("gh", args, {
		signal: new AbortController().signal,
		timeoutMs: DEFAULT_EXTERNAL_TIMEOUT_MS,
		maxOutputBytes: MAX_GITHUB_OUTPUT_BYTES,
	}));
}

function validateRuntimeOptions(options: ProductionShepherdRuntimeOptions): void {
	if (typeof options !== "object" || options === null) throw new Error("production runtime options are required");
	safeAbsolutePath(options.repositoryRoot, "repository root");
	safeAbsolutePath(options.stateRoot, "state root");
	safeAbsolutePath(options.trustedWorktreeRoot, "trusted worktree root");
	assertProductionRecovery(options.effectRecovery);
	if (!(options.git instanceof GitAdapter)) throw new Error("production runtime requires the genuine coordinator Git adapter");
	if (typeof options.agentSession?.run !== "function" || typeof options.agentSession.abort !== "function") {
		throw new Error("production runtime requires the embedded AgentSession adapter");
	}
	if (typeof options.reviewSession?.run !== "function") throw new Error("production runtime requires an exact-head review session");
	if (typeof options.parentReadyAuthority?.readParentReadyState !== "function"
		|| typeof options.parentReadyAuthority.beginParentReady !== "function"
		|| typeof options.parentReadyAuthority.compareConsumeAndMarkParentReady !== "function"
		|| typeof options.parentReadyAuthority.settleParentReady !== "function"
		|| typeof options.parentReadyAuthority.quarantineAndRollbackParentReady !== "function") {
		throw new Error("production runtime requires durable parent-ready authority");
	}
	if (typeof options.parentReadiness?.markExistingDraftReady !== "function") {
		throw new Error("production runtime requires a bounded existing-parent-draft ready transition");
	}
	if (typeof options.dispositionActor !== "string" || !SAFE_ACTOR.test(options.dispositionActor)) {
		throw new Error("production review disposition actor is invalid");
	}
}

/**
 * Builds the complete production controller synchronously. Constructors perform no network or
 * Git mutation; all later external calls remain typed, bounded, and fail closed. No adapter in
 * this graph exposes parent-to-default-branch merge authority.
 */
export function createProductionShepherdController(
	options: ProductionShepherdRuntimeOptions,
): ProductionShepherdController {
	validateRuntimeOptions(options);
	const repositoryRoot = safeAbsolutePath(options.repositoryRoot, "repository root");
	const stateRoot = safeAbsolutePath(options.stateRoot, "state root");
	const trustedWorktreeRoot = safeAbsolutePath(options.trustedWorktreeRoot, "trusted worktree root");
	const now = options.now ?? (() => new Date());
	const githubOptions: GhCliOrchestrationTransportOptions = { ...(options.github ?? {}), now };
	const githubFacade = createProductionGitHubOrchestrationFacade(githubOptions);
	const decisionRepository = new FileHumanDecisionRepository(join(stateRoot, "human-decisions"));
	const decision = adaptGitHubDecisionBroker(new GitHubDecisionBroker(
		decisionRepository,
		decisionTransport(githubOptions.execute),
		{ ...(options.decision ?? {}), now },
	));
	const reviewRepository = new GhProductionReviewRepository(githubOptions);
	const changedPaths = new GhProductionChangedPathSource(githubOptions);
	const reviewer = new AgentSessionProductionReviewAdapter(
		reviewRepository,
		options.reviewSession,
		changedPaths,
	);
	const github = new GitHubParentOrchestrator(
		githubFacade.transport,
		decision,
		reviewer,
		githubFacade.policySource,
		{ parentReadyAuthority: options.parentReadyAuthority, now },
	);
	const workspaceLifecycle = new ProductionWorkspaceLifecycle({
		workspaceAdapter: new WorkspaceAdapter(options.git),
		verification: new BoundedVerificationRunner({ executables: { node: process.execPath } }),
		agentSession: options.agentSession,
	});
	const finalizer = new ProductionParentFinalizer({
		transport: githubFacade.transport,
		policies: new ProductionParentPolicyAuthority(githubFacade.policySource),
		reviews: reviewer,
		readiness: options.parentReadiness,
	});
	const parentGate = new ProductionParentGateAdapter(
		decision,
		new GhParentPullRequestMergeLookup(githubOptions.execute ?? defaultGhOrchestrationExecutor, now),
	);
	return composeProductionShepherdController({
		stateStore: new ProductionFileStateStore(stateRoot),
		intake: new ProductionRepositoryPlanIntake(repositoryRoot),
		effects: new ProductionEffectJournal(stateRoot),
		effectRecovery: options.effectRecovery,
		workspaceLifecycle,
		github,
		reviewer,
		reviewRepository,
		decisionBroker: decision,
		parentHeads: new GhProductionParentHeadSource(githubOptions),
		coordinator: options.coordinator,
		trustedWorktreeRoot,
		dispositionActor: options.dispositionActor,
		finalizer,
		parentGate,
		now,
	});
}

function gitObjectEnvironment(): NodeJS.ProcessEnv {
	return {
		PATH: process.env.PATH,
		GIT_CONFIG_NOSYSTEM: "1",
		GIT_CONFIG_GLOBAL: NULL_DEVICE,
		GIT_TERMINAL_PROMPT: "0",
		GIT_OPTIONAL_LOCKS: "0",
		GIT_PAGER: "",
		LC_ALL: "C",
		LANG: "C",
		TZ: "UTC",
	};
}

const defaultProductionGitObjectReader: ProductionGitObjectReader = (request) => new Promise((resolveRead, reject) => {
	execFile("git", ["cat-file", "blob", `${request.headSha}:${request.path}`], {
		cwd: request.cwd,
		env: gitObjectEnvironment(),
		encoding: "buffer",
		maxBuffer: request.maxOutputBytes,
		timeout: request.timeoutMs,
		killSignal: "SIGTERM",
		signal: request.signal,
		windowsHide: true,
	}, (error, stdout) => {
		if (error !== null) {
			reject(new Error("bounded exact-head Git object read failed"));
			return;
		}
		resolveRead(Buffer.from(stdout));
	});
});

function reviewWorkspace(
	work: IndependentReviewWork,
	options: ExactHeadReviewRoleRequestFactoryOptions,
	readObject: ProductionGitObjectReader,
	timeoutMs: number,
	maxObjectBytes: number,
): ScopedWorkspace {
	const workspaceId = createHash("sha256")
		.update(`${work.repository}\0${work.headBranch}\0${work.headSha}`)
		.digest("hex")
		.slice(0, 32);
	const workspace: ScopedWorkspace = {
		id: `review-${workspaceId}`,
		cwd: options.coordinator.cwd,
		async readText(path: string, readOptions: { offset?: number; limit?: number; signal?: AbortSignal }) {
			if (!(readOptions.signal instanceof AbortSignal)) throw new Error("exact-head review read requires an AbortSignal");
			if (readOptions.signal.aborted) throw readOptions.signal.reason ?? new Error("exact-head review read was cancelled");
			const offset = readOptions.offset ?? 0;
			const limit = readOptions.limit ?? maxObjectBytes;
			if (!Number.isSafeInteger(offset) || offset < 0 || !Number.isSafeInteger(limit) || limit < 1
				|| limit > maxObjectBytes) throw new Error("exact-head review read range is invalid");
			const normalized = validateScopedPath(path, work.allowedScopes);
			await options.git.assertBinding(options.coordinator);
			const branchHead = await options.git.resolveBranchHead(options.coordinator, work.headBranch);
			if (branchHead !== work.headSha) throw new Error("review branch moved from the exact authorized head");
			const contents = await readObject({
				cwd: options.coordinator.cwd,
				headSha: work.headSha,
				path: normalized,
				timeoutMs,
				maxOutputBytes: maxObjectBytes,
				signal: readOptions.signal,
			});
			if (!Buffer.isBuffer(contents) || contents.byteLength > maxObjectBytes) {
				throw new Error("exact-head Git object exceeded its bound");
			}
			let text: string;
			try {
				text = new TextDecoder("utf-8", { fatal: true }).decode(contents);
			} catch {
				throw new Error("exact-head review object is not canonical UTF-8 text");
			}
			return text.slice(offset, offset + limit);
		},
		async editText() { throw new Error("exact-head review workspace is read-only"); },
		async writeText() { throw new Error("exact-head review workspace is read-only"); },
	};
	return Object.freeze(workspace);
}

/**
 * Creates read-only xhigh review requests over immutable Git objects. Reads use one fixed
 * `git cat-file blob <40sha>:<scoped-path>` argv shape; mutation methods always reject.
 */
export function createExactHeadReviewRoleRequestFactory(
	options: ExactHeadReviewRoleRequestFactoryOptions,
): ProductionReviewRoleRequestFactory {
	if (typeof options !== "object" || options === null || !Number.isSafeInteger(options.parentIssue)
		|| options.parentIssue < 1 || typeof options.git?.assertBinding !== "function"
		|| typeof options.git.resolveBranchHead !== "function") {
		throw new Error("exact-head review request factory options are invalid");
	}
	const timeoutMs = boundedPositive(options.timeoutMs, 30_000, 120_000, "review timeout");
	const maxObjectBytes = boundedPositive(
		options.maxObjectBytes,
		MAX_REVIEW_OBJECT_BYTES,
		MAX_REVIEW_OBJECT_BYTES,
		"review object limit",
	);
	const now = options.now ?? (() => new Date());
	const readObject = options.readObject ?? defaultProductionGitObjectReader;
	return (work, context): RoleRunRequest => {
		if (!SHA.test(work.baseSha) || !SHA.test(work.headSha) || !SAFE_BRANCH.test(work.headBranch)
			|| options.coordinator.defaultBranch === work.headBranch || work.headBranch === "main") {
			throw new Error("exact-head review target is invalid or targets the default branch");
		}
		const deadline = new Date(context.deadlineAt).valueOf();
		const remaining = Math.floor(deadline - now().valueOf());
		if (!Number.isSafeInteger(remaining) || remaining < 1) throw new Error("exact-head review deadline expired");
		const effectiveTimeout = Math.min(timeoutMs, remaining);
		const identity = createHash("sha256")
			.update(`${work.idempotencyMarker}\0${work.headSha}`)
			.digest("hex");
		return {
			role: "review",
			task: `Independently review ${work.workItemId} at exact range ${work.baseSha}..${work.headSha}.`,
			context: [
				`Repository ${work.repository}; pull request #${work.pullRequest}.`,
				`Base branch ${work.baseBranch}; head branch ${work.headBranch}.`,
				`Changed paths: ${work.changedPaths.join(", ")}.`,
				`Allowed scopes: ${work.allowedScopes.join(", ")}.`,
				"Return findings only; never mutate, approve, merge, or expand scope.",
			],
			timeoutMs: effectiveTimeout,
			signal: context.signal,
			workspace: reviewWorkspace(work, options, readObject, effectiveTimeout, maxObjectBytes),
			capabilities: [],
			authority: {
				issue: options.parentIssue,
				branch: work.headBranch,
				readOnly: true,
				workspaceId: `review-${identity.slice(0, 32)}`,
				readPrefixes: [...work.allowedScopes],
				writePrefixes: [],
				capabilityNames: [],
			},
			binding: {
				runId: `review-${identity.slice(0, 32)}`,
				generation: work.generation,
				laneId: `review-${identity.slice(32, 64)}`,
				candidateHead: work.headSha,
				validationNonce: identity,
			},
		};
	};
}

// Keep the recovery dependency visible to generated API/docs without pretending sparse journal
// records are self-sufficient. The parent recovery contract supplies the concrete projector.
export type ProductionRuntimeRecoveryAuthority = ProductionEffectRecoveryPort;
export type ProductionRuntimeRecoveryRecord = ProductionEffectRecord;
