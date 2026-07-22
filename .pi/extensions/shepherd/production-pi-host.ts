import { createHash, randomUUID } from "node:crypto";
import { lstat, mkdir, readFile, rename, writeFile } from "node:fs/promises";
import { isAbsolute, join, resolve } from "node:path";

import type { ShepherdResumeCommand, ShepherdStartCommand } from "./arguments.ts";
import type { ProductionAutonomousState } from "./autonomous-production-state.ts";
import type { ProductionEffectRecoveryPort } from "./autonomous-recovery.ts";
import {
	ShepherdAgentSessionRuntime,
	type AgentSessionRuntimeSdk,
} from "./agent-session-runtime.ts";
import type { AutonomousShepherdControllerPort } from "./extension.ts";
import {
	createProductionGitHubOrchestrationFacade,
	defaultGhOrchestrationExecutor,
	type GhCliOrchestrationTransportOptions,
	type GhOrchestrationExecutor,
} from "./gh-orchestration-transport.ts";
import {
	GitHubParentOrchestrator,
	type ExternalCallContext,
	type MarkParentReadyRequest,
	type ParentReadyAuthorityQuery,
	type ParentReadyAuthorityState,
	type ParentReadyCompareEffectResult,
	type ParentReadyDurableAuthorityBoundary,
	type RollbackParentReadyRequest,
	type SettleParentReadyAuthorityRequest,
} from "./github-orchestrator.ts";
import { GitAdapter, type GitBinding } from "./git-adapter.ts";
import { ProductionRepositoryPlanIntake } from "./production-intake.ts";
import { productionOrchestrationObjective } from "./production-orchestration-plan.ts";
import {
	EmbeddedAgentSessionProductionReviewSession,
	type ProductionReviewRoleRequestFactory,
	type ProductionReviewSession,
} from "./production-review-adapter.ts";
import type {
	ProductionParentReadyTransitionPort,
	ProductionParentReadyTransitionReceipt,
	ProductionParentReadyTransitionRequest,
} from "./production-parent-lifecycle.ts";
import {
	createExactHeadReviewRoleRequestFactory,
	createProductionShepherdController,
	type ProductionShepherdRuntimeOptions,
} from "./production-runtime.ts";
import type { ProductionAgentSessionPort } from "./production-workspace-lifecycle.ts";

const SAFE_ACTOR = /^[A-Za-z0-9](?:[A-Za-z0-9._-]*[A-Za-z0-9])?$/u;
const SHA = /^[0-9a-f]{40}$/u;
const REPOSITORY = /^[A-Za-z0-9][A-Za-z0-9._-]{0,99}\/[A-Za-z0-9][A-Za-z0-9._-]{0,99}$/u;
const DEFAULT_GITHUB_TIMEOUT_MS = 30_000;
const DEFAULT_GITHUB_OUTPUT_BYTES = 2 * 1024 * 1024;

type ClosableAgentRuntime = ProductionAgentSessionPort & Pick<ShepherdAgentSessionRuntime, "close">;
type ParentDraftEnsurer = (issue: number, signal: AbortSignal) => Promise<void>;
type ParentAuthorityValidator = (issue: number, signal: AbortSignal) => Promise<void>;

export interface ProductionPiEntrypointControllerOptions {
	issue: number;
	delegate: AutonomousShepherdControllerPort;
	validateAuthority: ParentAuthorityValidator;
	ensureParentDraft: ParentDraftEnsurer;
	resources: Array<{ close(): Promise<void> }>;
}

/**
 * Launch boundary for the production controller. A fresh run cannot create durable state until
 * the marker-bound non-default parent draft exists. Resume never republishes it: the durable run
 * already proves the start preflight completed.
 */
export class ProductionPiEntrypointController implements AutonomousShepherdControllerPort {
	readonly #issue: number;
	readonly #delegate: AutonomousShepherdControllerPort;
	readonly #validateAuthority: ParentAuthorityValidator;
	readonly #ensureParentDraft: ParentDraftEnsurer;
	readonly #resources: Array<{ close(): Promise<void> }>;
	#preflight: { controller: AbortController; promise: Promise<void> } | undefined;
	#closePromise: Promise<void> | undefined;
	#closed = false;

	constructor(options: ProductionPiEntrypointControllerOptions) {
		if (!Number.isSafeInteger(options.issue) || options.issue < 1
			|| typeof options.delegate?.status !== "function"
			|| typeof options.delegate.start !== "function"
			|| typeof options.delegate.resume !== "function"
			|| typeof options.delegate.stop !== "function"
			|| typeof options.delegate.shutdown !== "function"
			|| typeof options.validateAuthority !== "function"
			|| typeof options.ensureParentDraft !== "function"
			|| !Array.isArray(options.resources)
			|| options.resources.some((resource) => typeof resource?.close !== "function")) {
			throw new Error("production Pi entrypoint options are invalid");
		}
		this.#issue = options.issue;
		this.#delegate = options.delegate;
		this.#validateAuthority = options.validateAuthority;
		this.#ensureParentDraft = options.ensureParentDraft;
		this.#resources = [...options.resources];
	}

	#assertIssue(issue: number): void {
		if (issue !== this.#issue) throw new Error(`production controller is bound to parent issue #${this.#issue}`);
		if (this.#closed) throw new Error("production Pi entrypoint is closed");
	}

	status(issue: number): Promise<ProductionAutonomousState | undefined> {
		this.#assertIssue(issue);
		return this.#delegate.status(issue) as Promise<ProductionAutonomousState | undefined>;
	}

	async #runPreflight(issue: number, prepareDraft: boolean): Promise<void> {
		if (this.#preflight !== undefined) throw new Error("production authority preflight is already active");
		const controller = new AbortController();
		const promise = (async () => {
			await this.#validateAuthority(issue, controller.signal);
			if (prepareDraft) await this.#ensureParentDraft(issue, controller.signal);
		})();
		this.#preflight = { controller, promise };
		try {
			await promise;
		} finally {
			if (this.#preflight?.promise === promise) this.#preflight = undefined;
		}
	}

	async start(command: ShepherdStartCommand): Promise<ProductionAutonomousState> {
		this.#assertIssue(command.issue);
		if (await this.#delegate.status(command.issue)) {
			throw new Error(`production Shepherd state already exists for issue #${command.issue}; use resume`);
		}
		await this.#runPreflight(command.issue, true);
		return this.#delegate.start(command) as Promise<ProductionAutonomousState>;
	}

	async resume(command: ShepherdResumeCommand): Promise<ProductionAutonomousState> {
		this.#assertIssue(command.issue);
		await this.#runPreflight(command.issue, false);
		return this.#delegate.resume(command) as Promise<ProductionAutonomousState>;
	}

	async stop(issue: number): Promise<ProductionAutonomousState> {
		this.#assertIssue(issue);
		const preflight = this.#preflight;
		if (preflight !== undefined) {
			preflight.controller.abort(new Error("production authority preflight stop requested"));
			try { await preflight.promise; } catch { /* The authoritative state check below decides the result. */ }
		}
		const durable = await this.#delegate.status(issue);
		if (durable === undefined) {
			throw new Error(`production Shepherd initialization stopped before state for issue #${issue} was created`);
		}
		return this.#delegate.stop(issue) as Promise<ProductionAutonomousState>;
	}

	shutdown(): Promise<void> {
		if (this.#closePromise !== undefined) return this.#closePromise;
		this.#closed = true;
		this.#preflight?.controller.abort(new Error("production Pi entrypoint shutdown requested"));
		this.#closePromise = (async () => {
			const preflight = this.#preflight?.promise;
			if (preflight !== undefined) await preflight.catch(() => undefined);
			const results = await Promise.allSettled([
				this.#delegate.shutdown(),
				...this.#resources.map((resource) => resource.close()),
			]);
			const failures = results
				.filter((result): result is PromiseRejectedResult => result.status === "rejected")
				.map((result) => result.reason);
			if (failures.length > 0) throw new AggregateError(failures, "production Pi entrypoint shutdown failed");
		})();
		return this.#closePromise;
	}
}

/** The production finalizer owns parent ready transitions. This unused legacy authority fails closed. */
class FinalizerOwnedParentReadyAuthority implements ParentReadyDurableAuthorityBoundary {
	async readParentReadyState(_query: ParentReadyAuthorityQuery, _context: ExternalCallContext): Promise<ParentReadyAuthorityState | null> {
		return null;
	}
	async beginParentReady(_request: MarkParentReadyRequest): Promise<ParentReadyAuthorityState> { return this.#unsupported(); }
	async compareConsumeAndMarkParentReady(_request: MarkParentReadyRequest): Promise<ParentReadyCompareEffectResult> { return this.#unsupported(); }
	async settleParentReady(_request: SettleParentReadyAuthorityRequest): Promise<ParentReadyAuthorityState> { return this.#unsupported(); }
	async quarantineAndRollbackParentReady(_request: RollbackParentReadyRequest): Promise<never> { return this.#unsupported(); }
	#unsupported(): never {
		throw new Error("parent-ready mutation is owned only by the production parent finalizer transition port");
	}
}

interface DurableReadyDocument {
	schemaVersion: 1;
	request: ProductionParentReadyTransitionRequest;
	status: "prepared" | "applied";
	receipt?: ProductionParentReadyTransitionReceipt;
}

function canonicalTimestamp(value: unknown, description: string): string {
	if (typeof value !== "string" || value.length > 64) throw new Error(`invalid ${description}`);
	const date = new Date(value);
	if (!Number.isFinite(date.valueOf())) throw new Error(`invalid ${description}`);
	return date.toISOString();
}

function revisionAt(value: string): number {
	const result = Math.floor(new Date(value).valueOf() / 1_000);
	if (!Number.isSafeInteger(result) || result < 1) throw new Error("GitHub parent revision is invalid");
	return result;
}

function exactRecord(value: unknown, description: string): Record<string, unknown> {
	if (typeof value !== "object" || value === null || Array.isArray(value)) {
		throw new Error(`GitHub returned malformed ${description}`);
	}
	return value as Record<string, unknown>;
}

function parseGitHub(output: string, description: string): Record<string, unknown> {
	if (Buffer.byteLength(output) > DEFAULT_GITHUB_OUTPUT_BYTES) throw new Error(`GitHub ${description} output exceeded its bound`);
	try { return exactRecord(JSON.parse(output), description); }
	catch (error) { throw new Error(`GitHub returned malformed ${description} JSON`, { cause: error }); }
}

/** Durable, restart-reconciling draft-to-ready adapter. Its only mutation is GraphQL mark-ready. */
export class DurableGhParentReadiness implements ProductionParentReadyTransitionPort {
	readonly #root: string;
	readonly #execute: GhOrchestrationExecutor;
	readonly #now: () => Date;
	#tail: Promise<void> = Promise.resolve();

	constructor(stateRoot: string, options: GhCliOrchestrationTransportOptions = {}) {
		this.#root = join(resolve(stateRoot), "parent-ready-transitions");
		this.#execute = options.execute ?? defaultGhOrchestrationExecutor;
		this.#now = options.now ?? (() => new Date());
	}

	markExistingDraftReady(
		request: ProductionParentReadyTransitionRequest,
		context: ExternalCallContext,
	): Promise<ProductionParentReadyTransitionReceipt> {
		const run = this.#tail.then(() => this.#mark(request, context));
		this.#tail = run.then(() => undefined, () => undefined);
		return run;
	}

	#path(request: ProductionParentReadyTransitionRequest): string {
		const id = createHash("sha256").update(JSON.stringify(request)).digest("hex");
		return join(this.#root, `${id}.json`);
	}

	async #read(request: ProductionParentReadyTransitionRequest): Promise<DurableReadyDocument | undefined> {
		try {
			const value = JSON.parse(await readFile(this.#path(request), "utf8")) as DurableReadyDocument;
			if (value.schemaVersion !== 1 || JSON.stringify(value.request) !== JSON.stringify(request)
				|| (value.status !== "prepared" && value.status !== "applied")
				|| (value.status === "applied") !== (value.receipt !== undefined)) {
				throw new Error("durable parent-ready transition state is invalid or conflicts");
			}
			return value;
		} catch (error) {
			if (typeof error === "object" && error !== null && "code" in error && error.code === "ENOENT") return undefined;
			throw error;
		}
	}

	async #write(document: DurableReadyDocument): Promise<void> {
		await mkdir(this.#root, { recursive: true, mode: 0o700 });
		const path = this.#path(document.request);
		const temporary = `${path}.${process.pid}.${randomUUID()}.tmp`;
		await writeFile(temporary, `${JSON.stringify(document)}\n`, { encoding: "utf8", mode: 0o600, flag: "wx" });
		await rename(temporary, path);
		const metadata = await lstat(path);
		if (!metadata.isFile() || metadata.isSymbolicLink()) throw new Error("durable parent-ready transition path is unsafe");
	}

	async #call(args: readonly string[], context: ExternalCallContext): Promise<Record<string, unknown>> {
		const deadline = new Date(context.deadlineAt).valueOf();
		if (!Number.isFinite(deadline) || deadline <= Date.now()) throw new Error("parent-ready GitHub deadline expired");
		const output = await this.#execute("gh", args, {
			signal: context.signal,
			timeoutMs: Math.max(1, Math.min(DEFAULT_GITHUB_TIMEOUT_MS, deadline - Date.now())),
			maxOutputBytes: DEFAULT_GITHUB_OUTPUT_BYTES,
		});
		return parseGitHub(output, "parent pull request");
	}

	async #observe(request: ProductionParentReadyTransitionRequest, context: ExternalCallContext) {
		const row = await this.#call([
			"api", "--method", "GET", `/repos/${request.repository}/pulls/${request.pullRequest}`,
		], context);
		const head = exactRecord(row.head, "parent pull request head");
		const updatedAt = canonicalTimestamp(row.updated_at, "parent pull request update time");
		if (row.number !== request.pullRequest || row.state !== "open"
			|| head.ref !== request.branch || head.sha !== request.headSha
			|| typeof row.node_id !== "string" || row.node_id.length < 1 || row.node_id.length > 512) {
			throw new Error("parent pull request moved from the exact draft-to-ready authority");
		}
		return {
			draft: row.draft === true,
			revision: revisionAt(updatedAt),
			observedAt: this.#now().toISOString(),
			nodeId: row.node_id,
		};
	}

	async #mark(
		request: ProductionParentReadyTransitionRequest,
		context: ExternalCallContext,
	): Promise<ProductionParentReadyTransitionReceipt> {
		if (!REPOSITORY.test(request.repository) || !Number.isSafeInteger(request.parentIssue) || request.parentIssue < 1
			|| !Number.isSafeInteger(request.pullRequest) || request.pullRequest < 1
			|| !Number.isSafeInteger(request.generation) || request.generation < 1
			|| !SHA.test(request.headSha) || !Number.isSafeInteger(request.expectedRevision) || request.expectedRevision < 1) {
			throw new Error("parent-ready transition request is invalid");
		}
		try {
			const existing = await this.#read(request);
			if (existing?.receipt !== undefined) return structuredClone(existing.receipt);
			if (existing === undefined) await this.#write({ schemaVersion: 1, request, status: "prepared" });
			let observed = await this.#observe(request, context);
			let mutationFailure: unknown;
			if (observed.draft) {
				if (observed.revision !== request.expectedRevision) {
					throw new Error("parent pull request revision moved before draft-to-ready transition");
				}
				try {
					await this.#call([
						"api", "graphql",
						"-f", "query=mutation($pullRequestId:ID!){markPullRequestReadyForReview(input:{pullRequestId:$pullRequestId}){pullRequest{id isDraft}}}",
						"-f", `pullRequestId=${observed.nodeId}`,
					], context);
				} catch (error) { mutationFailure = error; }
				observed = await this.#observe(request, context);
			}
			if (observed.draft || observed.revision <= request.expectedRevision) {
				if (mutationFailure !== undefined) throw mutationFailure;
				throw new Error("parent pull request was not authoritatively marked ready");
			}
			const receipt: ProductionParentReadyTransitionReceipt = {
				schemaVersion: 1,
				authority: "transport",
				operation: "existing_draft_to_ready",
				...request,
				appliedRevision: observed.revision,
				observedAt: observed.observedAt,
			};
			await this.#write({ schemaVersion: 1, request, status: "applied", receipt });
			return receipt;
		} finally {
			if (context.signal.aborted) context.acknowledgeAbort();
		}
	}
}

export interface ProductionPiHostDependencies {
	git: GitAdapter;
	inspectCoordinator(git: GitAdapter, repositoryRoot: string): Promise<GitBinding>;
	createAgentRuntime(role: "implementation" | "review", sdk: AgentSessionRuntimeSdk): ClosableAgentRuntime;
	createReviewSession(
		runtime: ClosableAgentRuntime,
		requestFactory: ProductionReviewRoleRequestFactory,
	): ProductionReviewSession;
	createController(options: ProductionShepherdRuntimeOptions): AutonomousShepherdControllerPort;
	createParentReadyAuthority(stateRoot: string): ParentReadyDurableAuthorityBoundary;
	createParentReadiness(
		stateRoot: string,
		github: GhCliOrchestrationTransportOptions,
	): ProductionParentReadyTransitionPort;
	createParentDraftEnsurer(options: {
		repositoryRoot: string;
		git: GitAdapter;
		coordinator: GitBinding;
		github: GhCliOrchestrationTransportOptions;
		parentReadyAuthority: ParentReadyDurableAuthorityBoundary;
	}): ParentDraftEnsurer;
}

export interface ProductionPiHostOptions {
	issue: number;
	repositoryRoot: string;
	stateRoot: string;
	trustedWorktreeRoot: string;
	runtimeSdk: AgentSessionRuntimeSdk;
	dispositionActor?: string;
	github?: GhCliOrchestrationTransportOptions;
	/** Optional test/advanced override; production composes the exhaustive recovery authority. */
	effectRecovery?: ProductionEffectRecoveryPort;
	dependencies?: Partial<ProductionPiHostDependencies> & Pick<ProductionPiHostDependencies, "git">;
}

const FORBIDDEN_PARENT_INTEGRATION_TARGETS = new Set(["main", "master", "trunk"]);

function defaultParentAuthorityValidator(options: {
	repositoryRoot: string;
	git: GitAdapter;
	coordinator: GitBinding;
	inspectCoordinator(git: GitAdapter, repositoryRoot: string): Promise<GitBinding>;
}): ParentAuthorityValidator {
	return async (issue, signal) => {
		const snapshot = await new ProductionRepositoryPlanIntake(options.repositoryRoot).load(issue, signal);
		const live = await options.inspectCoordinator(options.git, options.repositoryRoot);
		if (signal.aborted) throw signal.reason ?? new Error("production authority preflight aborted");
		if (live.repositoryIdentity !== options.coordinator.repositoryIdentity
			|| live.worktreeIdentity !== options.coordinator.worktreeIdentity
			|| live.remoteIdentity !== options.coordinator.remoteIdentity
			|| live.fetchEndpointIdentity !== options.coordinator.fetchEndpointIdentity
			|| live.pushEndpointIdentity !== options.coordinator.pushEndpointIdentity) {
			throw new Error("production coordinator or remote identity changed since controller construction");
		}
		if (live.defaultBranch === undefined
			|| snapshot.plan.parentBaseBranch !== live.defaultBranch) {
			throw new Error("production parent plan base no longer matches the authoritative remote default branch");
		}
		if (snapshot.plan.parentBranch === live.defaultBranch
			|| FORBIDDEN_PARENT_INTEGRATION_TARGETS.has(snapshot.plan.parentBranch)) {
			throw new Error("production parent integration target is a protected default branch alias");
		}
	};
}

function defaultParentDraftEnsurer(options: {
	repositoryRoot: string;
	git: GitAdapter;
	coordinator: GitBinding;
	github: GhCliOrchestrationTransportOptions;
	parentReadyAuthority: ParentReadyDurableAuthorityBoundary;
}): ParentDraftEnsurer {
	return async (issue, signal) => {
		const intake = new ProductionRepositoryPlanIntake(options.repositoryRoot);
		const snapshot = await intake.load(issue, signal);
		const plan = snapshot.plan;
		if (plan.parentBaseBranch !== options.coordinator.defaultBranch) {
			throw new Error("production parent plan base is not the authoritative default branch");
		}
		if (await options.git.currentBranch(options.coordinator) !== plan.parentBranch) {
			throw new Error("production coordinator is not on the exact planned parent branch");
		}
		const status = await options.git.status(options.coordinator);
		if (!status.clean) throw new Error("production parent draft preflight requires a clean coordinator worktree");
		const [baseHead, head] = await Promise.all([
			options.git.resolveBranchHead(options.coordinator, plan.parentBaseBranch),
			options.git.resolveBranchHead(options.coordinator, plan.parentBranch),
		]);
		const scopes = [...new Set(plan.children.flatMap((child) => child.writeScopes))].sort();
		const changed = await options.git.diff(options.coordinator, { baseHead, head, scopes });
		const facade = createProductionGitHubOrchestrationFacade(options.github);
		const orchestrator = new GitHubParentOrchestrator(
			facade.transport,
			undefined,
			undefined,
			facade.policySource,
			{ parentReadyAuthority: options.parentReadyAuthority },
		);
		try {
			const orchestration = await orchestrator.createPlan(
				productionOrchestrationObjective(plan, 1),
				{ signal, deadlineAt: new Date(Date.now() + DEFAULT_GITHUB_TIMEOUT_MS).toISOString() },
			);
			await orchestrator.ensureParentDraftPullRequest(orchestration, {
				issue: plan.parentIssue,
				branch: plan.parentBranch,
				prBase: plan.parentBaseBranch,
				baseHead,
				head,
				changedScope: changed.changedScope,
				verificationState: "passed",
				repositoryIdentity: options.coordinator.repositoryIdentity,
				worktreeIdentity: options.coordinator.worktreeIdentity,
				dirty: false,
			}, { signal, deadlineAt: new Date(Date.now() + DEFAULT_GITHUB_TIMEOUT_MS).toISOString() });
		} finally {
			await orchestrator.stop({ deadlineAt: new Date(Date.now() + 5_000).toISOString() });
		}
	};
}

function productionDependencies(options: ProductionPiHostOptions): ProductionPiHostDependencies {
	const supplied = options.dependencies;
	const git = supplied?.git ?? new GitAdapter();
	return {
		git,
		inspectCoordinator: supplied?.inspectCoordinator ?? ((adapter, root) => adapter.inspect(root)),
		createAgentRuntime: supplied?.createAgentRuntime ?? ((role, sdk) => new ShepherdAgentSessionRuntime(sdk, {
			maxConcurrency: role === "implementation" ? 4 : 1,
		})),
		createReviewSession: supplied?.createReviewSession
			?? ((runtime, request) => new EmbeddedAgentSessionProductionReviewSession(runtime as ShepherdAgentSessionRuntime, request)),
		createController: supplied?.createController ?? createProductionShepherdController,
		createParentReadyAuthority: supplied?.createParentReadyAuthority
			?? (() => new FinalizerOwnedParentReadyAuthority()),
		createParentReadiness: supplied?.createParentReadiness
			?? ((stateRoot, github) => new DurableGhParentReadiness(stateRoot, github)),
		createParentDraftEnsurer: supplied?.createParentDraftEnsurer ?? defaultParentDraftEnsurer,
	};
}

/** Build the production controller without network calls; external ports run only on start/resume. */
export async function createProductionPiHostController(
	options: ProductionPiHostOptions,
): Promise<ProductionPiEntrypointController> {
	if (!Number.isSafeInteger(options.issue) || options.issue < 1
		|| !isAbsolute(options.repositoryRoot) || !isAbsolute(options.stateRoot) || !isAbsolute(options.trustedWorktreeRoot)) {
		throw new Error("production Pi host paths and issue are invalid");
	}
	const actor = options.dispositionActor ?? "shepherd-controller";
	if (!SAFE_ACTOR.test(actor)) throw new Error("production Pi host disposition actor is invalid");
	const dependencies = productionDependencies(options);
	await mkdir(options.stateRoot, { recursive: true, mode: 0o700 });
	await mkdir(options.trustedWorktreeRoot, { recursive: true, mode: 0o700 });
	const coordinator = await dependencies.inspectCoordinator(dependencies.git, options.repositoryRoot);
	if (coordinator.defaultBranch === undefined) throw new Error("production coordinator has no authoritative default branch");
	const implementation = dependencies.createAgentRuntime("implementation", options.runtimeSdk);
	const review = dependencies.createAgentRuntime("review", options.runtimeSdk);
	try {
		const reviewSession = dependencies.createReviewSession(
			review,
			createExactHeadReviewRoleRequestFactory({
				git: dependencies.git,
				coordinator,
				parentIssue: options.issue,
			}),
		);
		const readyAuthority = dependencies.createParentReadyAuthority(options.stateRoot);
		const github = options.github ?? {};
		const validateAuthority = defaultParentAuthorityValidator({
			repositoryRoot: options.repositoryRoot,
			git: dependencies.git,
			coordinator,
			inspectCoordinator: dependencies.inspectCoordinator,
		});
		const delegate = dependencies.createController({
			parentIssue: options.issue,
			repositoryRoot: options.repositoryRoot,
			stateRoot: options.stateRoot,
			trustedWorktreeRoot: options.trustedWorktreeRoot,
			coordinator,
			git: dependencies.git,
			agentSession: implementation,
			reviewSession,
			...(options.effectRecovery === undefined ? {} : { effectRecovery: options.effectRecovery }),
			parentReadyAuthority: readyAuthority,
			parentReadiness: dependencies.createParentReadiness(options.stateRoot, github),
			dispositionActor: actor,
			github,
		});
		return new ProductionPiEntrypointController({
			issue: options.issue,
			delegate,
			validateAuthority,
			ensureParentDraft: dependencies.createParentDraftEnsurer({
				repositoryRoot: options.repositoryRoot,
				git: dependencies.git,
				coordinator,
				github,
				parentReadyAuthority: readyAuthority,
			}),
			resources: [implementation, review],
		});
	} catch (error) {
		await Promise.allSettled([implementation.close(), review.close()]);
		throw error;
	}
}
