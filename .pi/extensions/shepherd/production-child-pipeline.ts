import { createHash } from "node:crypto";

import {
	ProductionLifecycleError,
	type ProductionEffectKind,
	type ProductionEffectRecord,
	type ProductionParentPlanDocument,
	type ProductionStageCheckpoint,
	type ProductionWorkspaceBinding,
} from "./autonomous-production-contract.ts";
import {
	productionEffectKey,
	type ProductionEffectJournalPort,
	type ProductionEffectIntent,
} from "./autonomous-effect-journal.ts";
import type { ProductionEffectRecoveryPort } from "./autonomous-recovery.ts";
import {
	validateProductionAutonomousState,
	type ProductionAutonomousState,
} from "./autonomous-production-state.ts";
import type {
	ProductionChildInterventionObservation,
	ProductionChildPipelineContext,
	ProductionChildPipelinePort,
} from "./production-controller.ts";
import type { GitBinding } from "./git-adapter.ts";
import {
	validateGitHubPullRequestEvidence,
} from "./github-evidence.ts";
import {
	materializeChildRecord,
	validateChildIntegrationReceipt,
	type ChildIntegrationDecision,
	type ChildIntegrationReceipt,
	type GitHubChildIssue,
	type GitHubParentOrchestrator,
	type MaterializedChildRecord,
	type ExternalCallContext,
	type ParentDecisionBroker,
	type ParentOrchestrationPlan,
} from "./github-orchestrator.ts";
import {
	assertHumanDecisionBinding,
	validateHumanDecisionRecord,
	type HumanDecisionBinding,
	type HumanDecisionRecord,
} from "./human-decision.ts";
import {
	buildProductionChildInterventionDecisionRequest,
} from "./production-human-gate.ts";
import {
	productionOrchestrationObjective,
} from "./production-orchestration-plan.ts";
import type {
	AgentSessionProductionReviewAdapter,
	ProductionReviewArtifact,
	ProductionReviewRepository,
} from "./production-review-adapter.ts";
import {
	createIndependentReviewWork,
	independentReviewAuthorizationDigest,
	independentReviewResultDigest,
	validateIndependentReviewRecord,
	type IndependentReviewTarget,
} from "./review-router.ts";
import type {
	ProductionWorkspaceLifecycle,
	ProductionWorkspaceSession,
} from "./production-workspace-lifecycle.ts";

const SHA = /^[0-9a-f]{40}$/u;
const DIGEST = /^[0-9a-f]{64}$/u;
const SAFE_ID = /^[A-Za-z0-9](?:[A-Za-z0-9._:-]*[A-Za-z0-9])?$/u;

export type ProductionWorkspaceLifecyclePort = Pick<ProductionWorkspaceLifecycle, "claim" | "abort" | "close">;

export type ProductionChildGitHubPort = Pick<
	GitHubParentOrchestrator,
	"createPlan" | "ensureChildIssue" | "ensureChildPullRequest" | "integrateChild" | "stop"
>;

export type ProductionExactHeadReviewPort = Pick<AgentSessionProductionReviewAdapter, "review">;

export interface ProductionParentHeadObservation {
	repository: string;
	branch: string;
	head: string;
}

/** Reads the authoritative current head of the non-default parent branch. */
export interface ProductionParentHeadSource {
	observe(plan: ProductionParentPlanDocument, signal: AbortSignal): Promise<ProductionParentHeadObservation>;
}

export interface ProductionChildIntegrationReceiptQuery {
	repository: string;
	childId: string;
	generation: number;
	digest: string;
}

/** Durable receipt lookup used by parent finalization after child integration. */
export interface ProductionChildIntegrationReceiptAuthority {
	find(
		query: ProductionChildIntegrationReceiptQuery,
		signal: AbortSignal,
	): Promise<ChildIntegrationReceipt | undefined>;
}

export interface ProductionChildPipelineOptions {
	workspaceLifecycle: ProductionWorkspaceLifecyclePort;
	github: ProductionChildGitHubPort;
	reviewer: ProductionExactHeadReviewPort;
	reviewRepository: ProductionReviewRepository;
	effects: ProductionEffectJournalPort;
	decisionBroker: ParentDecisionBroker;
	parentHeads: ProductionParentHeadSource;
	coordinator: GitBinding;
	trustedWorktreeRoot: string;
	dispositionActor: string;
	/** Required by the recovery barrier after a crash before controller checkpoint CAS. */
	recovery?: ProductionEffectRecoveryPort;
	/** Optional durable fallback for parent finalization after process restart. */
	receiptAuthority?: ProductionChildIntegrationReceiptAuthority;
	now?: () => Date;
}

export interface ProductionChildHeadReconciliationResult {
	checkpoint: ProductionStageCheckpoint;
	previousHead: string;
	head: string;
	invalidated: {
		verification: true;
		review: true;
		integration: true;
	};
}

interface EffectExecution<T> {
	kind: ProductionEffectKind;
	key: string;
	resultDigest: string;
	value: T;
}

interface SessionEntry {
	session: ProductionWorkspaceSession;
	runId: string;
	generation: number;
	childId: string;
}

interface PendingAcknowledgement {
	runId: string;
	generation: number;
	childId: string;
	resultDigest: string;
	accepts(state: ProductionAutonomousState): boolean;
}

function sha256(value: string): string {
	return createHash("sha256").update(value).digest("hex");
}

function canonical(value: unknown): unknown {
	if (Array.isArray(value)) return value.map(canonical);
	if (typeof value !== "object" || value === null) return value;
	const output: Record<string, unknown> = {};
	for (const key of Object.keys(value as Record<string, unknown>).sort()) {
		const item = (value as Record<string, unknown>)[key];
		if (item !== undefined) output[key] = canonical(item);
	}
	return output;
}

function digest(value: unknown): string {
	return sha256(JSON.stringify(canonical(value)));
}

function throwIfAborted(signal: AbortSignal): void {
	if (!(signal instanceof AbortSignal)) throw new Error("production pipeline AbortSignal is invalid");
	if (signal.aborted) {
		throw new ProductionLifecycleError("retryable", "production child operation was cancelled", ["cancelled"]);
	}
}

function externalContext(signal: AbortSignal, timeoutMs: number): ExternalCallContext {
	throwIfAborted(signal);
	return {
		signal,
		deadlineAt: new Date(Date.now() + timeoutMs).toISOString(),
		acknowledgeAbort() { /* The direct caller owns this signal. */ },
	};
}

function sameStrings(left: readonly string[], right: readonly string[]): boolean {
	return left.length === right.length && left.every((entry, index) => entry === right[index]);
}

function laneKey(runId: string, generation: number, childId: string): string {
	return `${runId}\u0000${generation}\u0000${childId}`;
}

function validateContext(context: ProductionChildPipelineContext): void {
	if (typeof context !== "object" || context === null || !SAFE_ID.test(context.runId)
		|| !Number.isSafeInteger(context.resourceGeneration) || context.resourceGeneration < 1
		|| !Number.isSafeInteger(context.generation) || context.generation < 1
		|| !Number.isSafeInteger(context.timeoutMs) || context.timeoutMs < 1
		|| context.state.runId !== context.runId || context.state.generation !== context.generation
		|| context.state.resourceGeneration !== context.resourceGeneration
		|| context.runtime.id !== context.child.id || context.child.access !== "mutating") {
		throw new ProductionLifecycleError("terminal", "production child context is invalid", ["context_invalid"]);
	}
	throwIfAborted(context.signal);
}

function validateParentHead(
	value: ProductionParentHeadObservation,
	plan: ProductionParentPlanDocument,
): ProductionParentHeadObservation {
	if (value.repository !== plan.repository || value.branch !== plan.parentBranch || !SHA.test(value.head)) {
		throw new ProductionLifecycleError("stale_parent", "authoritative parent head observation is mismatched", ["parent_head_mismatch"]);
	}
	return { repository: value.repository, branch: value.branch, head: value.head };
}

function checkpointContainsKey(checkpoint: ProductionStageCheckpoint | undefined, key: string): boolean {
	return checkpoint?.effectKey === key || checkpoint?.effectKeys?.includes(key) === true;
}

function checkpointContains(
	actual: ProductionStageCheckpoint | undefined,
	expected: ProductionStageCheckpoint,
): boolean {
	if (actual === undefined) return false;
	if (expected.workspace !== undefined && JSON.stringify(actual.workspace) !== JSON.stringify(expected.workspace)) return false;
	if (expected.pullRequest !== undefined && actual.pullRequest !== expected.pullRequest) return false;
	if (expected.review !== undefined && JSON.stringify(actual.review) !== JSON.stringify(expected.review)) return false;
	if (expected.integrationReceiptDigest !== undefined
		&& actual.integrationReceiptDigest !== expected.integrationReceiptDigest) return false;
	if (expected.parentHead !== undefined && actual.parentHead !== expected.parentHead) return false;
	return true;
}

function receiptMatchesQuery(receipt: ChildIntegrationReceipt, query: ProductionChildIntegrationReceiptQuery): boolean {
	return receipt.pullRequestSnapshot.repository === query.repository
		&& receipt.childId === query.childId
		&& receipt.generation === query.generation
		&& productionChildIntegrationReceiptDigest(receipt) === query.digest;
}

export function productionChildIntegrationReceiptDigest(value: ChildIntegrationReceipt): string {
	return digest(validateChildIntegrationReceipt(value));
}

function reviewTargetMatches(artifact: ProductionReviewArtifact, target: IndependentReviewTarget): boolean {
	const expected = createIndependentReviewWork(target);
	const actual = validateIndependentReviewRecord(artifact.review);
	return actual.idempotencyMarker === expected.idempotencyMarker
		&& actual.repository === expected.repository && actual.workItemId === expected.workItemId
		&& actual.pullRequest === expected.pullRequest && actual.generation === expected.generation
		&& actual.baseBranch === expected.baseBranch && actual.headBranch === expected.headBranch
		&& actual.baseSha === expected.baseSha && actual.headSha === expected.headSha
		&& sameStrings(actual.changedPaths, expected.changedPaths)
		&& sameStrings(actual.allowedScopes, expected.allowedScopes);
}

function interventionBinding(state: ProductionAutonomousState): HumanDecisionBinding {
	const gate = state.childGate;
	if (gate === undefined) {
		throw new ProductionLifecycleError("terminal", "child intervention gate is absent", ["intervention_binding_missing"]);
	}
	if ((gate.pullRequest === undefined) !== (gate.head === undefined)
		|| (gate.head !== undefined && !SHA.test(gate.head))) {
		throw new ProductionLifecycleError("terminal", "child intervention binding is malformed", ["intervention_binding_missing"]);
	}
	return {
		repository: gate.repository,
		target: gate.pullRequest === undefined
			? { kind: "issue", number: gate.issue }
			: { kind: "pull_request", number: gate.pullRequest },
		generation: gate.generation,
		...(gate.head === undefined ? {} : { headSha: gate.head }),
	};
}

function validateInterventionRecord(
	value: HumanDecisionRecord,
	requestId: string,
	binding: HumanDecisionBinding,
): HumanDecisionRecord {
	const record = validateHumanDecisionRecord(value);
	assertHumanDecisionBinding(record, binding);
	if (record.requestId !== requestId || record.gate !== "scope"
		|| record.allowedOptions.length !== 2
		|| record.allowedOptions[0] !== "authorize-one-retry"
		|| record.allowedOptions[1] !== "abort-child") {
		throw new ProductionLifecycleError("terminal", "child intervention record is stale or mismatched", ["intervention_mismatch"]);
	}
	return record;
}

function classifyIntegration(blockers: readonly string[]): ProductionLifecycleError {
	const values = [...new Set(blockers)];
	if (values.includes("head_moved")) {
		return new ProductionLifecycleError(
			"correction_required",
			"authoritative child pull request head moved and must be reclaimed, reverified, and rereviewed",
			["child_head_moved", ...values.filter((value) => value !== "head_moved")],
		);
	}
	if (values.some((value) => value === "topology_mismatch"
		|| value === "policy_moved" || value === "policy_mismatch" || value === "stale_evidence")) {
		return new ProductionLifecycleError("stale_parent", "child integration evidence moved", values);
	}
	if (values.some((value) => value === "review_missing" || value === "undispositioned_finding"
		|| value === "changes_requested" || value === "unresolved_thread")) {
		return new ProductionLifecycleError("correction_required", "child integration review evidence is not clean", values);
	}
	if (values.some((value) => value === "ci_not_green" || value === "merge_blocked")) {
		return new ProductionLifecycleError("retryable", "child integration checks are not authoritatively green", values);
	}
	return new ProductionLifecycleError("terminal", "child integration evidence failed closed", values);
}

/**
 * Concrete controller bridge. Every external mutation is journaled through observed, then the
 * controller calls acknowledge only after its checkpoint CAS has persisted exact evidence.
 */
export class ProductionChildPipeline implements
	ProductionChildPipelinePort,
	ProductionEffectRecoveryPort,
	ProductionChildIntegrationReceiptAuthority {
	readonly #workspaceLifecycle: ProductionWorkspaceLifecyclePort;
	readonly #github: ProductionChildGitHubPort;
	readonly #reviewer: ProductionExactHeadReviewPort;
	readonly #reviewRepository: ProductionReviewRepository;
	readonly #effects: ProductionEffectJournalPort;
	readonly #decisionBroker: ParentDecisionBroker;
	readonly #parentHeads: ProductionParentHeadSource;
	readonly #coordinator: GitBinding;
	readonly #trustedWorktreeRoot: string;
	readonly #dispositionActor: string;
	readonly #recovery?: ProductionEffectRecoveryPort;
	readonly #receiptAuthority?: ProductionChildIntegrationReceiptAuthority;
	readonly #now: () => Date;
	readonly #sessions = new Map<string, SessionEntry>();
	readonly #pendingEffects = new Map<string, Map<string, EffectExecution<unknown>>>();
	readonly #acknowledgements = new Map<string, PendingAcknowledgement>();
	readonly #plans = new Map<string, Promise<ParentOrchestrationPlan>>();
	readonly #children = new Map<string, MaterializedChildRecord>();
	readonly #receipts = new Map<string, ChildIntegrationReceipt>();
	readonly #abortPromises = new Map<string, Promise<void>>();
	readonly #joinPromises = new Map<string, Promise<void>>();
	#closed = false;
	#closePromise: Promise<void> | undefined;

	constructor(options: ProductionChildPipelineOptions) {
		if (typeof options !== "object" || options === null
			|| typeof options.workspaceLifecycle?.claim !== "function"
			|| typeof options.workspaceLifecycle?.abort !== "function"
			|| typeof options.workspaceLifecycle?.close !== "function"
			|| typeof options.github?.createPlan !== "function"
			|| typeof options.github?.ensureChildIssue !== "function"
			|| typeof options.github?.ensureChildPullRequest !== "function"
			|| typeof options.github?.integrateChild !== "function"
			|| typeof options.github?.stop !== "function"
			|| typeof options.reviewer?.review !== "function"
			|| typeof options.reviewRepository?.find !== "function"
			|| typeof options.reviewRepository?.recordDispositions !== "function"
			|| typeof options.effects?.prepare !== "function"
			|| typeof options.effects?.observe !== "function"
			|| typeof options.effects?.apply !== "function"
			|| typeof options.decisionBroker?.request !== "function"
			|| typeof options.decisionBroker?.poll !== "function"
			|| typeof options.decisionBroker?.consume !== "function"
			|| typeof options.parentHeads?.observe !== "function"
			|| typeof options.trustedWorktreeRoot !== "string" || !options.trustedWorktreeRoot.startsWith("/")
			|| typeof options.dispositionActor !== "string" || !SAFE_ID.test(options.dispositionActor)) {
			throw new Error("production child pipeline ports are invalid");
		}
		this.#workspaceLifecycle = options.workspaceLifecycle;
		this.#github = options.github;
		this.#reviewer = options.reviewer;
		this.#reviewRepository = options.reviewRepository;
		this.#effects = options.effects;
		this.#decisionBroker = options.decisionBroker;
		this.#parentHeads = options.parentHeads;
		this.#coordinator = structuredClone(options.coordinator);
		this.#trustedWorktreeRoot = options.trustedWorktreeRoot;
		this.#dispositionActor = options.dispositionActor;
		this.#recovery = options.recovery;
		this.#receiptAuthority = options.receiptAuthority;
		this.#now = options.now ?? (() => new Date());
		if (!Number.isFinite(this.#now().valueOf())) throw new Error("production child pipeline clock is invalid");
	}

	async #effect<T>(
		context: ProductionChildPipelineContext,
		kind: ProductionEffectKind,
		descriptor: unknown,
		operation: (effectKey: string) => Promise<T>,
	): Promise<EffectExecution<T>> {
		validateContext(context);
		return this.#effectFor({
			runId: context.runId,
			generation: context.generation,
			childId: context.child.id,
			signal: context.signal,
		}, kind, descriptor, operation);
	}

	async #effectFor<T>(
		fence: { runId: string; generation: number; childId: string; signal: AbortSignal },
		kind: ProductionEffectKind,
		descriptor: unknown,
		operation: (effectKey: string) => Promise<T>,
	): Promise<EffectExecution<T>> {
		throwIfAborted(fence.signal);
		const intentDigest = digest(descriptor);
		const coordinates = {
			kind,
			runId: fence.runId,
			generation: fence.generation,
			childId: fence.childId,
			intentDigest,
		};
		const intent: ProductionEffectIntent = {
			...coordinates,
			key: productionEffectKey(coordinates),
			recoveryDescriptor: descriptor,
		};
		await this.#effects.prepare(intent, this.#now());
		throwIfAborted(fence.signal);
		const value = await operation(intent.key);
		throwIfAborted(fence.signal);
		const resultDigest = digest(value);
		await this.#effects.observe(
			intent.key,
			{ runId: fence.runId, generation: fence.generation },
			resultDigest,
			this.#now(),
		);
		const result = { kind, key: intent.key, resultDigest, value };
		const key = laneKey(fence.runId, fence.generation, fence.childId);
		const pending = this.#pendingEffects.get(key) ?? new Map<string, EffectExecution<unknown>>();
		pending.set(result.key, result);
		this.#pendingEffects.set(key, pending);
		return result;
	}

	#pending(context: ProductionChildPipelineContext): EffectExecution<unknown>[] {
		return [...(this.#pendingEffects.get(laneKey(context.runId, context.generation, context.child.id))?.values() ?? [])];
	}

	#pendingKind<T>(context: ProductionChildPipelineContext, kind: ProductionEffectKind): EffectExecution<T> | undefined {
		return this.#pending(context).find((effect) => effect.kind === kind) as EffectExecution<T> | undefined;
	}

	#checkpoint(
		context: ProductionChildPipelineContext,
		primary: EffectExecution<unknown>,
		value: Omit<ProductionStageCheckpoint, "effectKey" | "effectKeys">,
	): ProductionStageCheckpoint {
		const checkpoint: ProductionStageCheckpoint = {
			...value,
			effectKey: primary.key,
			effectKeys: this.#pending(context).map((effect) => effect.key),
		};
		for (const effect of this.#pending(context)) {
			this.#acknowledgements.set(effect.key, {
				runId: context.runId,
				generation: context.generation,
				childId: context.child.id,
				resultDigest: effect.resultDigest,
				accepts: (state) => {
					const child = state.children.find((candidate) => candidate.id === context.child.id);
					return checkpointContainsKey(child?.checkpoint, effect.key)
						&& checkpointContains(child?.checkpoint, checkpoint);
				},
			});
		}
		return checkpoint;
	}

	async #session(context: ProductionChildPipelineContext): Promise<SessionEntry> {
		validateContext(context);
		if (this.#closed) throw new Error("production child pipeline is closed");
		const key = laneKey(context.runId, context.generation, context.child.id);
		const existing = this.#sessions.get(key);
		if (existing !== undefined) return existing;
		const prior = context.runtime.ownership;
		const parentHead = prior?.baseHead ?? validateParentHead(
			await this.#parentHeads.observe(context.plan, context.signal),
			context.plan,
		).head;
		let claimed: ProductionWorkspaceSession | undefined;
		const effect = await this.#effect(context, "workspace_claim", {
			parentIssue: context.plan.parentIssue,
			childIssue: context.child.issue,
			childId: context.child.id,
			parentHead,
			mode: prior === undefined ? "start" : "resume",
			ownershipId: prior?.ownershipId,
			attempt: context.runtime.attempts,
		}, async () => {
			claimed = await this.#workspaceLifecycle.claim({
				runId: context.runId,
				generation: context.generation,
				coordinator: this.#coordinator,
				trustedWorktreeRoot: this.#trustedWorktreeRoot,
				parentIssue: context.plan.parentIssue,
				parentBranch: context.plan.parentBranch,
				parentHead,
				child: context.child,
				mode: prior === undefined ? "start" : "resume",
				...(prior === undefined ? {} : { ownershipId: prior.ownershipId }),
			});
			return claimed.binding;
		});
		if (claimed === undefined || JSON.stringify(claimed.binding) !== JSON.stringify(effect.value)) {
			throw new ProductionLifecycleError("terminal", "workspace claim lacks exact binding evidence", ["workspace_binding_missing"]);
		}
		if (prior !== undefined && JSON.stringify(claimed.binding) !== JSON.stringify(prior)) {
			throw new ProductionLifecycleError("stale_parent", "resumed workspace ownership changed", ["ownership_moved"]);
		}
		const entry = { session: claimed, runId: context.runId, generation: context.generation, childId: context.child.id };
		this.#sessions.set(key, entry);
		return entry;
	}

	async #plan(context: ProductionChildPipelineContext): Promise<ParentOrchestrationPlan> {
		const key = `${context.state.planDigest}:${context.resourceGeneration}`;
		let pending = this.#plans.get(key);
		if (pending === undefined) {
			pending = this.#github.createPlan(
				productionOrchestrationObjective(context.plan, context.resourceGeneration),
				{ signal: context.signal, deadlineAt: new Date(Date.now() + context.timeoutMs).toISOString() },
			);
			this.#plans.set(key, pending);
			void pending.catch(() => { if (this.#plans.get(key) === pending) this.#plans.delete(key); });
		}
		return pending;
	}

	async #materialized(
		context: ProductionChildPipelineContext,
		parent: ParentOrchestrationPlan,
	): Promise<MaterializedChildRecord> {
		const key = laneKey(context.runId, context.generation, context.child.id);
		const existing = this.#children.get(key);
		if (existing !== undefined) return existing;
		const issue: GitHubChildIssue = await this.#github.ensureChildIssue(parent, context.child.id, {
			signal: context.signal,
			deadlineAt: new Date(Date.now() + context.timeoutMs).toISOString(),
		});
		if (issue.number !== context.child.issue) {
			throw new ProductionLifecycleError("terminal", "canonical child issue differs from the executable plan", ["child_issue_mismatch"]);
		}
		const child = materializeChildRecord(parent, context.child.id, issue);
		this.#children.set(key, child);
		return child;
	}

	async workspace(context: ProductionChildPipelineContext): Promise<ProductionStageCheckpoint> {
		const entry = await this.#session(context);
		const primary = this.#pending(context).at(-1);
		if (primary === undefined) {
			return { summary: "workspace already durably claimed", workspace: entry.session.binding };
		}
		return this.#checkpoint(context, primary, {
			summary: "isolated production workspace claimed",
			workspace: entry.session.binding,
		});
	}

	async implement(context: ProductionChildPipelineContext): Promise<ProductionStageCheckpoint> {
		const entry = await this.#session(context);
		const effect = await this.#effect(context, "agent_implementation", {
			workspace: entry.session.binding,
			attempt: context.runtime.attempts,
			corrections: context.runtime.corrections,
			taskDigest: digest(context.child.task),
		}, async () => entry.session.implement({ timeoutMs: context.timeoutMs, signal: context.signal }));
		return this.#checkpoint(context, effect, {
			summary: "production implementation AgentSession completed",
			workspace: entry.session.binding,
		});
	}

	async verify(context: ProductionChildPipelineContext): Promise<ProductionStageCheckpoint> {
		const entry = await this.#session(context);
		const effect = await this.#effect(context, "shell_verification", {
			workspace: entry.session.binding,
			attempt: context.runtime.attempts,
			corrections: context.runtime.corrections,
			commands: context.child.verification,
		}, async () => {
			const results = await entry.session.verify(context.signal);
			if (results.length < 1 || results.length > context.child.verification.length
				|| results.some((result, index) => result.id !== context.child.verification[index]?.id)
				|| results.some((result) => result.status === "failed" && result.failureKind === undefined)) {
				throw new ProductionLifecycleError("terminal", "bounded verification returned malformed command evidence", ["verification_evidence_invalid"]);
			}
			return results.map((result) => ({
				id: result.id,
				status: result.status,
				...(result.failureKind === undefined ? {} : { failureKind: result.failureKind }),
			}));
		});
		const commands = effect.value;
		const passed = commands.length === context.child.verification.length
			&& commands.every((command) => command.status === "passed");
		return this.#checkpoint(context, effect, {
			summary: passed
				? "all bounded production verification commands passed"
				: "bounded production verification failed and requires correction",
			workspace: entry.session.binding,
			verification: {
				status: passed ? "passed" : "failed",
				resultDigest: digest(commands),
				commands,
			},
		});
	}

	async publish(context: ProductionChildPipelineContext): Promise<ProductionStageCheckpoint> {
		const entry = await this.#session(context);
		const commit = this.#pendingKind<Awaited<ReturnType<ProductionWorkspaceSession["commit"]>>>(context, "git_commit")
			?? await this.#effect(context, "git_commit", {
			workspace: entry.session.binding,
			issue: context.child.issue,
			attempt: context.runtime.attempts,
			corrections: context.runtime.corrections,
		}, async () => entry.session.commit(`feat(shepherd): complete #${context.child.issue} ${context.child.slug}`, context.signal));
		if (commit.value.head !== entry.session.binding.head || !SHA.test(commit.value.head)) {
			throw new ProductionLifecycleError("terminal", "commit evidence does not bind the current child head", ["commit_head_mismatch"]);
		}
		const push = this.#pendingKind<Awaited<ReturnType<ProductionWorkspaceSession["push"]>>>(context, "git_push")
			?? await this.#effect(context, "git_push", {
			branch: entry.session.binding.branch,
			head: entry.session.binding.head,
		}, async () => entry.session.push(context.signal));
		if (push.value.branch !== entry.session.binding.branch || push.value.head !== entry.session.binding.head) {
			throw new ProductionLifecycleError("terminal", "push evidence does not bind the current child head", ["push_head_mismatch"]);
		}
		const handoff = await entry.session.captureHandoff(context.signal);
		const publication = await this.#effect(context, "child_pull_request", {
			repository: context.plan.repository,
			childId: context.child.id,
			generation: context.resourceGeneration,
			branch: handoff.branch,
			baseBranch: handoff.prBase,
			baseHead: handoff.baseHead,
			head: handoff.head,
			changedScope: handoff.changedScope,
		}, async () => {
			const parent = await this.#plan(context);
			const child = await this.#materialized(context, parent);
			if (child.branch !== handoff.branch || child.prBase !== handoff.prBase) {
				throw new ProductionLifecycleError("terminal", "workspace branch is not the canonical stacked child branch", ["branch_mismatch"]);
			}
			return this.#github.ensureChildPullRequest(parent, child, handoff, {
				signal: context.signal,
				deadlineAt: new Date(Date.now() + context.timeoutMs).toISOString(),
			});
		});
		const evidence = validateGitHubPullRequestEvidence(publication.value);
		if (evidence.repository !== context.plan.repository || evidence.workItemId !== context.child.id
			|| evidence.generation !== context.resourceGeneration || evidence.state !== "open" || evidence.draft
			|| evidence.baseBranch !== handoff.prBase || evidence.headBranch !== handoff.branch
			|| evidence.baseSha !== handoff.baseHead || evidence.headSha !== handoff.head
			|| !sameStrings(evidence.changedPaths, [...handoff.changedScope].sort())
			|| !sameStrings(evidence.allowedScopes, [...context.child.writeScopes].sort())) {
			throw new ProductionLifecycleError("terminal", "stacked pull request evidence is wrong, draft, or stale", ["pull_request_mismatch"]);
		}
		return this.#checkpoint(context, publication, {
			summary: "verified child commit pushed and exact stacked pull request published",
			workspace: entry.session.binding,
			pullRequest: evidence.number,
		});
	}

	#target(
		context: ProductionChildPipelineContext,
		handoff: Awaited<ReturnType<ProductionWorkspaceSession["captureHandoff"]>>,
	): IndependentReviewTarget {
		const pullRequest = context.runtime.checkpoint?.pullRequest;
		if (!Number.isSafeInteger(pullRequest) || (pullRequest as number) < 1) {
			throw new ProductionLifecycleError("terminal", "review requires an exact published pull request", ["pull_request_missing"]);
		}
		return {
			repository: context.plan.repository,
			workItemId: context.child.id,
			pullRequest: pullRequest as number,
			generation: context.resourceGeneration,
			baseBranch: handoff.prBase,
			headBranch: handoff.branch,
			baseSha: handoff.baseHead,
			headSha: handoff.head,
			changedPaths: [...handoff.changedScope],
			allowedScopes: [...context.child.writeScopes],
		};
	}

	async review(context: ProductionChildPipelineContext): Promise<ProductionStageCheckpoint> {
		const entry = await this.#session(context);
		const handoff = await entry.session.captureHandoff(context.signal);
		const target = this.#target(context, handoff);
		const effect = await this.#effect(context, "independent_review", target, async () =>
			this.#reviewer.review(target, externalContext(context.signal, context.timeoutMs)));
		if (!reviewTargetMatches(effect.value, target)) {
			throw new ProductionLifecycleError("terminal", "independent review is not bound to the exact PR head", ["review_mismatch"]);
		}
		const review = validateIndependentReviewRecord(effect.value.review);
		const authorizationDigest = review.verdict === "clean"
			? independentReviewAuthorizationDigest(review)
			: undefined;
		return this.#checkpoint(context, effect, {
			summary: review.verdict === "clean" ? "independent exact-head review is clean" : "independent review returned findings",
			workspace: entry.session.binding,
			pullRequest: target.pullRequest,
			review: {
				status: review.verdict === "clean" ? "clean" : "blocked",
				baseHead: review.baseSha,
				head: review.headSha,
				resultDigest: independentReviewResultDigest(review),
				...(authorizationDigest === undefined ? {} : { authorizationDigest }),
				completedAt: review.completedAt,
				findings: review.findings.map((finding) => ({ id: finding.id, summary: finding.summary })),
			},
		});
	}

	async correct(
		context: ProductionChildPipelineContext,
		findings: readonly string[],
	): Promise<ProductionStageCheckpoint> {
		const entry = await this.#session(context);
		if (!Array.isArray(findings) || findings.length === 0) {
			throw new ProductionLifecycleError("terminal", "correction requires exact durable findings", ["correction_mismatch"]);
		}
		const reviewCheckpoint = context.runtime.checkpoint?.review;
		const verificationCheckpoint = context.runtime.checkpoint?.verification;
		let target: IndependentReviewTarget | undefined;
		let artifact: ProductionReviewArtifact | undefined;
		let source: { kind: "review" | "verification"; resultDigest: string };
		if (reviewCheckpoint?.status === "blocked") {
			const handoff = await entry.session.captureHandoff(context.signal);
			target = this.#target(context, handoff);
			if (reviewCheckpoint.head !== target.headSha
				|| !sameStrings(findings, reviewCheckpoint.findings.map((finding) => finding.summary))) {
				throw new ProductionLifecycleError("terminal", "correction is not bound to authoritative review findings", ["correction_mismatch"]);
			}
			const lookup = await this.#reviewRepository.find(target, externalContext(context.signal, context.timeoutMs));
			if (!lookup.complete) throw new ProductionLifecycleError("terminal", "review repository lookup is incomplete", ["review_incomplete"]);
			const matches = lookup.items.filter((candidate) => reviewTargetMatches(candidate, target!)
				&& independentReviewResultDigest(candidate.review) === reviewCheckpoint.resultDigest);
			if (matches.length !== 1) {
				throw new ProductionLifecycleError("terminal", "exact review artifact is absent or ambiguous", ["review_ambiguous"]);
			}
			artifact = matches[0];
			source = { kind: "review", resultDigest: independentReviewResultDigest(artifact.review) };
		} else if (verificationCheckpoint?.status === "failed") {
			const expected = verificationCheckpoint.commands
				.filter((command) => command.status === "failed")
				.map((command) => `Verification ${command.id} failed (${command.failureKind ?? "unknown"}).`);
			if (!sameStrings(findings, expected)) {
				throw new ProductionLifecycleError("terminal", "correction is not bound to failed verification evidence", ["correction_mismatch"]);
			}
			source = { kind: "verification", resultDigest: verificationCheckpoint.resultDigest };
		} else {
			throw new ProductionLifecycleError("terminal", "correction has no exact review or verification authority", ["correction_authority_missing"]);
		}
		const effect = await this.#effect(context, "agent_correction", {
			source,
			...(target === undefined ? {} : { target }),
			findings,
			correction: context.runtime.corrections,
		}, async () => {
			const handoffResult = await entry.session.correct({
				timeoutMs: context.timeoutMs,
				signal: context.signal,
				findings: [...findings],
			});
			if (target === undefined || artifact === undefined) return { handoff: handoffResult };
			const recordedAt = this.#now().toISOString();
			const disposition = await this.#reviewRepository.recordDispositions(target, artifact.review.findings.map((finding) => ({
				findingId: finding.id,
				kind: "fixed" as const,
				rationale: "Corrected by the bounded production correction AgentSession; exact head requires re-verification and re-review.",
				actor: this.#dispositionActor,
				headSha: artifact.review.headSha,
				recordedAt,
			})), externalContext(context.signal, context.timeoutMs));
			return { handoff: handoffResult, dispositionRevision: disposition.revision };
		});
		return this.#checkpoint(context, effect, {
			summary: "bounded correction completed; verification and independent review invalidated",
			workspace: entry.session.binding,
			...(target === undefined ? {} : { pullRequest: target.pullRequest }),
		});
	}

	async refresh(context: ProductionChildPipelineContext): Promise<ProductionStageCheckpoint> {
		const entry = await this.#session(context);
		const previous = entry.session.binding;
		const current = validateParentHead(await this.#parentHeads.observe(context.plan, context.signal), context.plan);
		if (current.head === previous.baseHead) {
			throw new ProductionLifecycleError("stale_parent", "parent refresh did not observe a newer exact head", ["parent_head_unchanged"]);
		}
		const effect = await this.#effect(context, "parent_refresh", {
			previousBaseHead: previous.baseHead,
			newParentHead: current.head,
			previousHead: previous.head,
			attempt: context.runtime.attempts,
		}, async (effectKey) => entry.session.refreshParent({
			previousParentHead: previous.baseHead,
			newParentHead: current.head,
			effectKey,
		}, context.signal));
		if (effect.value.previousBaseHead !== previous.baseHead || effect.value.baseHead !== current.head
			|| effect.value.verificationInvalidated !== true || effect.value.reviewInvalidated !== true
			|| entry.session.binding.baseHead !== current.head || entry.session.binding.head !== effect.value.head) {
			throw new ProductionLifecycleError("stale_parent", "parent refresh evidence is incomplete or mismatched", ["refresh_mismatch"]);
		}
		return this.#checkpoint(context, effect, {
			summary: "workspace refreshed to the authoritative parent head; verification and review required",
			workspace: entry.session.binding,
		});
	}

	/**
	 * Reclaims an externally moved canonical child branch. The controller must persist the
	 * returned checkpoint while deleting the invalidated verification, review, and integration
	 * fields before acknowledging the effect, then rerun verification and exact-head review.
	 */
	async reconcileChildHead(
		context: ProductionChildPipelineContext,
	): Promise<ProductionChildHeadReconciliationResult> {
		const durable = context.runtime.checkpoint?.workspace ?? context.runtime.ownership;
		const pullRequest = context.runtime.checkpoint?.pullRequest;
		if (durable === undefined || !SHA.test(durable.head) || !Number.isSafeInteger(pullRequest)
			|| (pullRequest as number) < 1) {
			throw new ProductionLifecycleError(
				"terminal",
				"child-head reconciliation requires exact durable workspace and pull request coordinates",
				["child_head_reconciliation_binding_missing"],
			);
		}
		if (context.runtime.checkpoint?.integrationReceiptDigest !== undefined) {
			throw new ProductionLifecycleError(
				"terminal",
				"an already integrated child cannot reclaim a later pull request head",
				["integrated_child_head_moved"],
			);
		}
		const review = context.runtime.checkpoint?.review;
		if (review?.status !== "clean" || review.head !== durable.head || review.baseHead !== durable.baseHead) {
			throw new ProductionLifecycleError(
				"terminal",
				"child-head reconciliation is not bound to the displaced exact-head review",
				["child_head_reconciliation_review_missing"],
			);
		}
		const entry = await this.#session(context);
		const before = entry.session.binding;
		if (before.branch !== durable.branch || before.baseBranch !== durable.baseBranch
			|| before.baseHead !== durable.baseHead
			|| !sameStrings(before.writeScopes, durable.writeScopes)) {
			throw new ProductionLifecycleError(
				"terminal",
				"active child workspace moved outside the durable reconciliation coordinates",
				["child_head_reconciliation_workspace_mismatch"],
			);
		}
		const effect = await this.#effect(context, "child_head_reconciliation", {
			operation: "child_head_reconciliation",
			repository: context.plan.repository,
			childId: context.child.id,
			pullRequest: pullRequest as number,
			branch: durable.branch,
			baseHead: durable.baseHead,
			previousHead: durable.head,
			reviewResultDigest: review.resultDigest,
		}, async (effectKey) => entry.session.reconcileChildHead({
			previousHead: durable.head,
			effectKey,
		}, context.signal));
		const after = entry.session.binding;
		if (effect.value.previousHead !== durable.head || effect.value.branch !== durable.branch
			|| effect.value.baseHead !== durable.baseHead || !SHA.test(effect.value.head)
			|| effect.value.head === durable.head || after.head !== effect.value.head
			|| after.baseHead !== durable.baseHead || after.branch !== durable.branch
			|| effect.value.verificationInvalidated !== true || effect.value.reviewInvalidated !== true
			|| effect.value.integrationInvalidated !== true) {
			throw new ProductionLifecycleError(
				"terminal",
				"child-head reconciliation evidence is incomplete or mismatched",
				["child_head_reconciliation_mismatch"],
			);
		}
		const checkpoint = this.#checkpoint(context, effect, {
			summary: "authoritative child head reclaimed; verification and exact-head review required",
			workspace: after,
			pullRequest: pullRequest as number,
		});
		return {
			checkpoint,
			previousHead: durable.head,
			head: effect.value.head,
			invalidated: { verification: true, review: true, integration: true },
		};
	}

	async integrate(context: ProductionChildPipelineContext): Promise<ProductionStageCheckpoint> {
		const entry = await this.#session(context);
		const handoff = await entry.session.captureHandoff(context.signal);
		const currentParent = validateParentHead(await this.#parentHeads.observe(context.plan, context.signal), context.plan);
		if (currentParent.head !== handoff.baseHead) {
			throw new ProductionLifecycleError("stale_parent", "parent head moved before child integration", ["head_moved"]);
		}
		const review = context.runtime.checkpoint?.review;
		if (review?.status !== "clean" || review.baseHead !== handoff.baseHead || review.head !== handoff.head) {
			throw new ProductionLifecycleError("correction_required", "child integration requires a clean exact-head review", ["review_missing"]);
		}
		const effect = await this.#effect(context, "child_integration", {
			repository: context.plan.repository,
			childId: context.child.id,
			generation: context.resourceGeneration,
			pullRequest: context.runtime.checkpoint?.pullRequest,
			baseHead: handoff.baseHead,
			head: handoff.head,
			reviewResultDigest: review.resultDigest,
		}, async () => {
			const parent = await this.#plan(context);
			const child = await this.#materialized(context, parent);
			const decision = await this.#github.integrateChild(parent, child, handoff, {
				signal: context.signal,
				deadlineAt: new Date(Date.now() + context.timeoutMs).toISOString(),
			});
			if (decision.kind === "blocked") return { decision };
			const parentHead = validateParentHead(await this.#parentHeads.observe(context.plan, context.signal), context.plan);
			return { decision, parentHead };
		});
		const decision: ChildIntegrationDecision = effect.value.decision;
		if (decision.kind === "blocked") throw classifyIntegration(decision.blockers);
		const receipt = validateChildIntegrationReceipt(decision.receipt);
		if (receipt.childId !== context.child.id || receipt.generation !== context.resourceGeneration
			|| receipt.pullRequest !== context.runtime.checkpoint?.pullRequest
			|| receipt.baseSha !== handoff.baseHead || receipt.headSha !== handoff.head
			|| receipt.parentBranch !== context.plan.parentBranch) {
			throw new ProductionLifecycleError("terminal", "child integration receipt is stale or mismatched", ["integration_receipt_mismatch"]);
		}
		if (effect.value.parentHead === undefined || effect.value.parentHead.head === handoff.baseHead) {
			throw new ProductionLifecycleError("retryable", "integrated child lacks the resulting authoritative parent head", ["parent_head_not_advanced"]);
		}
		const receiptDigest = productionChildIntegrationReceiptDigest(receipt);
		this.#receipts.set(receiptDigest, receipt);
		return this.#checkpoint(context, effect, {
			summary: decision.reused ? "existing exact child integration reconciled" : "child integrated into the non-default parent branch",
			workspace: entry.session.binding,
			pullRequest: receipt.pullRequest,
			integrationReceiptDigest: receiptDigest,
			parentHead: effect.value.parentHead.head,
		});
	}

	async requestIntervention(
		context: ProductionChildPipelineContext,
		reason: "retry_budget_exhausted" | "correction_budget_exhausted",
	): Promise<{ requestId: string; pullRequest?: number; head?: string; effectKey?: string }> {
		validateContext(context);
		const pullRequest = context.runtime.checkpoint?.pullRequest;
		const head = context.runtime.checkpoint?.workspace?.head ?? context.runtime.ownership?.head;
		if (head === undefined || !SHA.test(head)) {
			throw new ProductionLifecycleError("human_required", "child intervention requires exact workspace ownership", ["intervention_binding_missing"]);
		}
		const requestId = `shepherd-${digest({
			runId: context.runId,
			generation: context.generation,
			childId: context.child.id,
			reason,
			pullRequest,
			head,
		}).slice(0, 48)}`;
		const request = buildProductionChildInterventionDecisionRequest({
			requestId,
			repository: context.plan.repository,
			childIssue: context.child.issue,
			generation: context.generation,
			reason,
			actorAllowlist: context.plan.actorAllowlist,
			expiresAt: context.plan.decisionExpiresAt,
			question: "Authorize exactly one additional bounded child attempt, or abort this child?",
		});
		const binding: HumanDecisionBinding = {
			repository: request.repository,
			target: { kind: "issue", number: context.child.issue },
			generation: request.generation,
		};
		const effect = await this.#effect(context, "human_request", request, async () =>
			validateInterventionRecord(
				await this.#decisionBroker.request(request, externalContext(context.signal, context.timeoutMs)),
				requestId,
				binding,
			));
		this.#acknowledgements.set(effect.key, {
			runId: context.runId,
			generation: context.generation,
			childId: context.child.id,
			resultDigest: effect.resultDigest,
			accepts: (state) => state.childGate?.childId === context.child.id
				&& state.childGate.requestId === requestId && state.childGate.reason === reason
				&& state.childGate.pullRequest === undefined && state.childGate.head === undefined
				&& state.childGate.status === "pending",
		});
		return { requestId, effectKey: effect.key };
	}

	async observeIntervention(
		stateValue: ProductionAutonomousState,
		signal: AbortSignal,
	): Promise<ProductionChildInterventionObservation> {
		const state = validateProductionAutonomousState(stateValue);
		throwIfAborted(signal);
		const gate = state.childGate;
		if (gate === undefined) throw new ProductionLifecycleError("terminal", "child intervention gate is absent", ["intervention_missing"]);
		const binding = interventionBinding(state);
		const polled = validateInterventionRecord(
			await this.#decisionBroker.poll(gate.requestId, binding, { signal }, externalContext(signal, state.timeoutMs)),
			gate.requestId,
			binding,
		);
		if (polled.status === "pending" || polled.status === "expired") return { status: "pending" };
		if ((polled.status !== "decided" && polled.status !== "consumed") || polled.decision === undefined) {
			throw new ProductionLifecycleError("terminal", "child intervention decision is invalid", ["intervention_invalid"]);
		}
		const option = polled.decision.option;
		if (option !== "authorize-one-retry" && option !== "abort-child") {
			throw new ProductionLifecycleError("terminal", "child intervention option is unauthorized", ["intervention_option_invalid"]);
		}
		const child = state.children.find((candidate) => candidate.id === gate.childId);
		if (child === undefined) throw new ProductionLifecycleError("terminal", "child intervention target is absent", ["child_missing"]);
		const effect = await this.#effectFor({
			runId: state.runId,
			generation: state.generation,
			childId: child.id,
			signal,
		}, "human_consume", {
			requestId: gate.requestId,
			binding,
			option,
		}, async () => validateInterventionRecord(
			await this.#decisionBroker.consume(gate.requestId, binding, externalContext(signal, state.timeoutMs)),
			gate.requestId,
			binding,
		));
		if (effect.value.status !== "consumed" || effect.value.decision?.option !== option) {
			throw new ProductionLifecycleError("terminal", "child intervention was not durably consumed", ["intervention_not_consumed"]);
		}
		const status = option === "authorize-one-retry" ? "authorized" as const : "aborted" as const;
		this.#acknowledgements.set(effect.key, {
			runId: state.runId,
			generation: state.generation,
			childId: gate.childId,
			resultDigest: effect.resultDigest,
			accepts: (persisted) => persisted.childGate?.requestId === gate.requestId
				&& persisted.childGate.childId === gate.childId && persisted.childGate.status === status,
		});
		return { status, effectKey: effect.key };
	}

	async acknowledge(effectKey: string, stateValue: ProductionAutonomousState): Promise<void> {
		if (!DIGEST.test(effectKey)) throw new Error("production effect acknowledgement key is invalid");
		const state = validateProductionAutonomousState(stateValue);
		const record = await this.#effects.load(effectKey);
		if (record === undefined) throw new Error("production effect acknowledgement is absent");
		const pending = this.#acknowledgements.get(effectKey);
		if (pending === undefined || pending.runId !== record.runId || pending.generation !== record.generation
			|| pending.childId !== record.childId || state.runId !== record.runId || state.generation !== record.generation
			|| !pending.accepts(state)) {
			throw new Error("persisted controller state does not acknowledge the exact production effect");
		}
		if (record.phase === "prepared" || record.resultDigest !== pending.resultDigest) {
			throw new Error("production effect acknowledgement lacks exact observed result evidence");
		}
		await this.#effects.apply(effectKey, { runId: record.runId, generation: record.generation }, this.#now());
		this.#pendingEffects.get(laneKey(record.runId, record.generation, record.childId!))?.delete(effectKey);
	}

	async observe(record: ProductionEffectRecord, signal: AbortSignal): Promise<{ resultDigest: string }> {
		if (this.#recovery === undefined) {
			throw new ProductionLifecycleError("terminal", "authoritative production effect recovery is not configured", ["recovery_authority_missing"]);
		}
		return this.#recovery.observe(structuredClone(record), signal);
	}

	async apply(record: ProductionEffectRecord, signal: AbortSignal): Promise<void> {
		if (this.#recovery === undefined) {
			throw new ProductionLifecycleError("terminal", "authoritative production effect recovery is not configured", ["recovery_authority_missing"]);
		}
		await this.#recovery.apply(structuredClone(record), signal);
	}

	async find(
		query: ProductionChildIntegrationReceiptQuery,
		signal: AbortSignal,
	): Promise<ChildIntegrationReceipt | undefined> {
		throwIfAborted(signal);
		if (!DIGEST.test(query.digest) || !Number.isSafeInteger(query.generation) || query.generation < 1) {
			throw new Error("production integration receipt query is invalid");
		}
		const local = this.#receipts.get(query.digest);
		if (local !== undefined) return receiptMatchesQuery(local, query) ? structuredClone(local) : undefined;
		const durable = await this.#receiptAuthority?.find(query, signal);
		if (durable === undefined) return undefined;
		const receipt = validateChildIntegrationReceipt(durable);
		return receiptMatchesQuery(receipt, query) ? receipt : undefined;
	}

	abort(runId: string): Promise<void> {
		if (!SAFE_ID.test(runId)) return Promise.reject(new Error("production run ID is invalid"));
		const existing = this.#abortPromises.get(runId);
		if (existing !== undefined) return existing;
		const operation = this.#workspaceLifecycle.abort(runId).finally(() => {
			for (const [key, entry] of this.#sessions) if (entry.runId === runId) this.#sessions.delete(key);
		});
		this.#abortPromises.set(runId, operation);
		return operation;
	}

	join(runId: string): Promise<void> {
		if (!SAFE_ID.test(runId)) return Promise.reject(new Error("production run ID is invalid"));
		const existing = this.#joinPromises.get(runId);
		if (existing !== undefined) return existing;
		const sessions = [...this.#sessions.entries()].filter(([, entry]) => entry.runId === runId);
		const operation = Promise.all(sessions.map(([, entry]) => entry.session.join())).then(() => {
			for (const [key] of sessions) this.#sessions.delete(key);
		});
		this.#joinPromises.set(runId, operation);
		return operation;
	}

	close(): Promise<void> {
		if (this.#closePromise !== undefined) return this.#closePromise;
		this.#closed = true;
		this.#closePromise = (async () => {
			const results = await Promise.allSettled([
				this.#workspaceLifecycle.close(),
				this.#github.stop({ deadlineAt: new Date(Date.now() + 30_000).toISOString() }),
			]);
			const failures = results.filter((result): result is PromiseRejectedResult => result.status === "rejected");
			if (failures.length === 1) throw failures[0].reason;
			if (failures.length > 1) throw new AggregateError(failures.map((failure) => failure.reason), "production child pipeline close failed");
			const github = results[1];
			if (github.status === "fulfilled" && github.value.kind !== "joined") {
				throw new Error("production GitHub orchestrator did not join every active external call");
			}
		})();
		return this.#closePromise;
	}
}
