import { createHash } from "node:crypto";
import { execFile } from "node:child_process";
import { isAbsolute, join, resolve } from "node:path";

import {
	ProductionLifecycleError,
	type ProductionEffectKind,
	type ProductionEffectRecord,
	type ProductionStageCheckpoint,
	type ProductionVerificationCommand,
	type ProductionWorkspaceBinding,
} from "./autonomous-production-contract.ts";
import { ProductionEffectJournal, type ProductionEffectJournalPort } from "./autonomous-effect-journal.ts";
import { ProductionRecoveryBarrier, type ProductionEffectRecoveryPort } from "./autonomous-recovery.ts";
import {
	authorizeProductionChildRetry,
	evolveProductionState,
	ProductionFileStateStore,
	reconcileProductionChildHead,
	refreshProductionChildOwnership,
	waitForProductionChildIntervention,
	type ProductionAutonomousState,
	type ProductionStateStore,
} from "./autonomous-production-state.ts";
import {
	ProductionWorkspaceLifecycle,
	productionWorkspaceOwnershipId,
	type ProductionAgentSessionPort,
	type ProductionVerificationPort,
} from "./production-workspace-lifecycle.ts";
import { ProductionRepositoryPlanIntake } from "./production-intake.ts";
import {
	ProductionShepherdController,
	type ProductionParentFinalizerPort,
	type ProductionParentGatePort,
	type ProductionParentMergeObservationEffectPort,
	type ProductionPlanIntakePort,
} from "./production-controller.ts";
import {
	ProductionChildPipeline,
	productionChildIntegrationReceiptDigest,
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
	validateChildIntegrationReceipt,
	adaptGitHubDecisionBroker,
	GitHubParentOrchestrator,
	type ExternalCallContext,
	type GitHubOrchestrationTransport,
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
import { FileHumanDecisionRepository, type HumanDecisionRepository } from "./human-decision.ts";
import { assertHumanDecisionBinding, validateHumanDecisionRecord, type HumanDecisionBinding } from "./human-decision.ts";
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
import type { ParentPullRequestMergeLookup } from "./production-human-gate.ts";
import {
	validateGitHubPullRequestEvidence,
	validateGitHubChangedPathEvidence,
	validateRequiredGitHubCheckPolicy,
	type GitHubChangedPathEvidence,
	type RequiredGitHubCheckPolicy,
} from "./github-evidence.ts";
import type { RoleRunRequest } from "./agent-session-runtime.ts";
import {
	createIndependentReviewWork,
	independentReviewAuthorizationDigest,
	independentReviewResultDigest,
	validateAgentSessionAttestation,
	validateIndependentReviewRecord,
	type IndependentReviewTarget,
	type IndependentReviewWork,
} from "./review-router.ts";
import { validateScopedPath, type ScopedWorkspace } from "./tool-policy.ts";
import {
	ProductionEffectRecoveryAuthority,
	type ProductionRecoveryProbe,
	type ProductionRecoveryProbeTable,
} from "./production-effect-recovery-authority.ts";
import { ProductionAgentEffectReceiptRepository } from "./production-agent-effect-receipts.ts";
import { ProductionParentMergeEffectJournal } from "./production-parent-merge-effect.ts";

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
	parentMergeEffects: ProductionParentMergeObservationEffectPort;
	now?: () => Date;
	newRunId?: () => string;
}

export interface ProductionShepherdRuntimeOptions {
	/** One controller/authority is durably bound to one parent issue. */
	parentIssue: number;
	repositoryRoot: string;
	stateRoot: string;
	trustedWorktreeRoot: string;
	coordinator: GitBinding;
	/** The exact GitAdapter instance that inspected and issued the coordinator binding. */
	git: GitAdapter;
	agentSession: ProductionAgentSessionPort;
	reviewSession: ProductionReviewSession;
	/** Optional test/advanced override; production constructs the genuine exhaustive authority. */
	effectRecovery?: ProductionEffectRecoveryPort;
	/** Durable authority around the uncertain existing-parent-PR ready transition. */
	parentReadyAuthority: ParentReadyDurableAuthorityBoundary;
	/** Bounded adapter that changes only an existing parent draft PR to ready. It cannot merge. */
	parentReadiness: ProductionParentReadyTransitionPort;
	dispositionActor: string;
	github?: GhCliOrchestrationTransportOptions;
	decision?: GitHubDecisionBrokerOptions;
	now?: () => Date;
}

export interface ProductionRuntimeRecoveryProbeOptions {
	git: Pick<GitAdapter,
		"assertBinding" | "inspect" | "currentBranch" | "resolveBranchHead" | "resolveRemoteBranchHead"
		| "readCommitSubject" | "isAncestor" | "status" | "diff"
	>;
	workspace: Pick<WorkspaceAdapter, "findClaim" | "findParentRefreshReceipt" | "findChildHeadReceipt">;
	agentEffects: Pick<ProductionAgentEffectReceiptRepository, "find">;
	coordinator: GitBinding;
	trustedWorktreeRoot: string;
	verification: ProductionVerificationPort;
	github: Pick<GitHubOrchestrationTransport, "findPullRequests" | "findChildIntegration" | "proveAncestry">;
	reviews: Pick<ProductionReviewRepository, "find">;
	decisions: Pick<HumanDecisionRepository, "load">;
	parentMerges: ParentPullRequestMergeLookup;
	parentHeads: {
		observe(
			plan: { repository: string; parentBranch: string },
			signal: AbortSignal,
		): Promise<ProductionParentHeadObservation>;
	};
	dispositionActor: string;
	now?: () => Date;
	timeoutMs?: number;
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

function canonicalRecoveryValue(value: unknown): unknown {
	if (Array.isArray(value)) return value.map(canonicalRecoveryValue);
	if (value === null || typeof value !== "object") return value;
	return Object.fromEntries(Object.keys(value as Record<string, unknown>).sort().flatMap((key) => {
		const item = (value as Record<string, unknown>)[key];
		return item === undefined ? [] : [[key, canonicalRecoveryValue(item)]];
	}));
}

function recoveryDigest(value: unknown): string {
	return createHash("sha256").update(JSON.stringify(canonicalRecoveryValue(value))).digest("hex");
}

function sameRecoveryValue(left: unknown, right: unknown): boolean {
	return JSON.stringify(canonicalRecoveryValue(left)) === JSON.stringify(canonicalRecoveryValue(right));
}

function recoveryRecord(value: unknown, description: string): Record<string, unknown> {
	if (typeof value !== "object" || value === null || Array.isArray(value)) {
		throw new ProductionLifecycleError("terminal", `${description} is malformed`, ["recovery_descriptor_invalid"]);
	}
	return value as Record<string, unknown>;
}

function recoveryCommon(
	descriptorValue: unknown,
	kind: ProductionEffectKind,
	state: ProductionAutonomousState,
	childId: string | undefined,
): Record<string, unknown> {
	const descriptor = recoveryRecord(descriptorValue, `production ${kind} recovery descriptor`);
	const child = childId === undefined
		? undefined
		: state.children.find((candidate) => candidate.id === childId);
	if (descriptor.operation !== kind || descriptor.repository !== state.repository
		|| descriptor.parentIssue !== state.parentIssue || descriptor.parentBranch !== state.parentBranch
		|| descriptor.parentBaseBranch !== state.parentBaseBranch || descriptor.planDigest !== state.planDigest
		|| descriptor.generation !== state.resourceGeneration
		|| (childId !== undefined && (descriptor.childId !== childId || child === undefined
			|| descriptor.childIssue !== child.issue || descriptor.childSlug !== child.slug))) {
		throw new ProductionLifecycleError(
			"terminal",
			`production ${kind} recovery descriptor moved from durable plan coordinates`,
			["recovery_descriptor_moved"],
		);
	}
	return descriptor;
}

function recoveryWorkspace(value: unknown, description: string): ProductionWorkspaceBinding {
	const workspace = recoveryRecord(value, description) as unknown as ProductionWorkspaceBinding;
	if (typeof workspace.claimId !== "string" || typeof workspace.ownershipId !== "string"
		|| typeof workspace.repositoryIdentity !== "string" || typeof workspace.worktreeIdentity !== "string"
		|| typeof workspace.cwd !== "string" || !isAbsolute(workspace.cwd)
		|| typeof workspace.branch !== "string" || typeof workspace.baseBranch !== "string"
		|| !SHA.test(workspace.baseHead) || !SHA.test(workspace.head)
		|| !Array.isArray(workspace.writeScopes) || workspace.writeScopes.length === 0
		|| workspace.writeScopes.some((scope) => typeof scope !== "string")) {
		throw new ProductionLifecycleError("terminal", `${description} is incomplete`, ["recovery_descriptor_invalid"]);
	}
	return structuredClone(workspace);
}

function recoveryChild(
	state: ProductionAutonomousState,
	childId: string | undefined,
	description: string,
): ProductionAutonomousState["children"][number] {
	const child = childId === undefined ? undefined : state.children.find((candidate) => candidate.id === childId);
	if (child === undefined) {
		throw new ProductionLifecycleError("terminal", `${description} child is absent`, ["recovery_child_missing"]);
	}
	return child;
}

function recoveredWorkspaceCheckpoint(
	state: ProductionAutonomousState,
	childId: string | undefined,
	description: string,
): ProductionWorkspaceBinding {
	const child = recoveryChild(state, childId, description);
	const workspace = child.checkpoint?.workspace ?? child.ownership;
	if (workspace === undefined) {
		throw new ProductionLifecycleError("terminal", `${description} lacks durable workspace truth`, ["recovery_workspace_missing"]);
	}
	return structuredClone(workspace);
}

function assertScopedStatus(
	entries: readonly { path: string; originalPath?: string }[],
	scopes: readonly string[],
	description: string,
): void {
	for (const entry of entries) {
		try {
			validateScopedPath(entry.path, scopes);
			if (entry.originalPath !== undefined) validateScopedPath(entry.originalPath, scopes);
		} catch (error) {
			throw new ProductionLifecycleError(
				"terminal",
				`${description} escaped immutable child scopes`,
				["recovery_scope_escape"],
			);
		}
	}
}

function projectRecoveredCorrection(
	request: Parameters<ProductionRecoveryProbe>[0],
	checkpoint: ProductionStageCheckpoint,
	now: () => Date,
): ProductionAutonomousState {
	const childId = request.record.childId;
	return evolveProductionState(request.currentState, recoveryFence(request.currentState), (draft) => {
		const child = draft.children.find((candidate) => candidate.id === childId);
		if (child === undefined) throw new Error("recovered correction child is absent");
		child.checkpoint = mergeRecoveryCheckpoint(child.checkpoint, checkpoint);
		delete child.checkpoint.verification;
		delete child.checkpoint.review;
		delete child.checkpoint.integrationReceiptDigest;
		child.status = "running";
		child.stage = "verification";
	}, now());
}

function recoveryFence(state: ProductionAutonomousState) {
	return {
		issue: state.parentIssue,
		revision: state.revision,
		generation: state.generation,
		runId: state.runId,
	};
}

const RECOVERY_STAGE_ORDER: Record<string, number> = {
	pending: 0,
	workspace: 1,
	implementation: 2,
	verification: 3,
	publication: 4,
	review: 5,
	correction: 6,
	integration: 7,
	succeeded: 8,
	failed: 8,
	cancelled: 8,
};

function monotonicStage(current: string, requested: string): typeof current {
	return (RECOVERY_STAGE_ORDER[current] ?? -1) >= (RECOVERY_STAGE_ORDER[requested] ?? -1) ? current : requested;
}

function mergeRecoveryCheckpoint(
	current: ProductionStageCheckpoint | undefined,
	next: ProductionStageCheckpoint,
): ProductionStageCheckpoint {
	const effectKeys = [...new Set([
		...(current?.effectKeys ?? []),
		...(current?.effectKey === undefined ? [] : [current.effectKey]),
		...(next.effectKeys ?? []),
		...(next.effectKey === undefined ? [] : [next.effectKey]),
	])];
	return {
		...(current ?? {}),
		...next,
		...(effectKeys.length === 0 ? {} : { effectKeys }),
		...(next.workspace === undefined && current?.workspace !== undefined ? { workspace: current.workspace } : {}),
		...(next.verification === undefined && current?.verification !== undefined ? { verification: current.verification } : {}),
		...(next.pullRequest === undefined && current?.pullRequest !== undefined ? { pullRequest: current.pullRequest } : {}),
		...(next.review === undefined && current?.review !== undefined ? { review: current.review } : {}),
		...(next.integrationReceiptDigest === undefined && current?.integrationReceiptDigest !== undefined
			? { integrationReceiptDigest: current.integrationReceiptDigest } : {}),
		...(next.parentHead === undefined && current?.parentHead !== undefined ? { parentHead: current.parentHead } : {}),
	};
}

function projectRecoveredChild(
	request: Parameters<ProductionRecoveryProbe>[0],
	checkpoint: ProductionStageCheckpoint,
	nextStage: "implementation" | "verification" | "publication" | "review" | "correction" | "integration" | "succeeded",
	now: () => Date,
	options: { forceStage?: boolean; succeeded?: boolean } = {},
): ProductionAutonomousState {
	const childId = request.record.childId;
	if (childId === undefined) throw new Error(`production ${request.record.kind} recovery requires a child binding`);
	return evolveProductionState(request.currentState, recoveryFence(request.currentState), (draft) => {
		const child = draft.children.find((candidate) => candidate.id === childId);
		if (child === undefined) throw new Error(`production ${request.record.kind} recovery child is absent`);
		const laterWorkspace = (RECOVERY_STAGE_ORDER[child.stage] ?? -1) > (RECOVERY_STAGE_ORDER[nextStage] ?? -1)
			? child.checkpoint?.workspace ?? child.ownership
			: undefined;
		const effectiveCheckpoint = laterWorkspace === undefined ? checkpoint : { ...checkpoint, workspace: laterWorkspace };
		child.checkpoint = mergeRecoveryCheckpoint(child.checkpoint, effectiveCheckpoint);
		if (effectiveCheckpoint.workspace !== undefined) {
			if (child.ownership !== undefined) {
				const before = child.ownership;
				const after = effectiveCheckpoint.workspace;
				if (before.claimId !== after.claimId || before.ownershipId !== after.ownershipId
					|| before.repositoryIdentity !== after.repositoryIdentity || before.worktreeIdentity !== after.worktreeIdentity
					|| before.cwd !== after.cwd || before.branch !== after.branch || before.baseBranch !== after.baseBranch
					|| before.baseHead !== after.baseHead || !sameRecoveryValue(before.writeScopes, after.writeScopes)) {
					throw new Error(`production ${request.record.kind} recovery workspace changed immutable ownership coordinates`);
				}
			}
			child.ownership = structuredClone(effectiveCheckpoint.workspace);
		}
		child.stage = options.forceStage ? nextStage : monotonicStage(child.stage, nextStage) as typeof child.stage;
		if (options.succeeded) {
			child.status = "succeeded";
			child.stage = "succeeded";
		}
	}, now());
}

function probeContext(signal: AbortSignal, timeoutMs: number): ExternalCallContext {
	return {
		signal,
		deadlineAt: new Date(Date.now() + timeoutMs).toISOString(),
		acknowledgeAbort() {},
	};
}

function sameReviewTarget(reviewValue: unknown, targetValue: IndependentReviewTarget): boolean {
	const review = validateIndependentReviewRecord(reviewValue);
	const work = createIndependentReviewWork(targetValue);
	return review.repository === work.repository && review.workItemId === work.workItemId
		&& review.pullRequest === work.pullRequest && review.generation === work.generation
		&& review.baseBranch === work.baseBranch && review.headBranch === work.headBranch
		&& review.baseSha === work.baseSha && review.headSha === work.headSha
		&& review.idempotencyMarker === work.idempotencyMarker
		&& JSON.stringify(review.changedPaths) === JSON.stringify(work.changedPaths)
		&& JSON.stringify(review.allowedScopes) === JSON.stringify(work.allowedScopes);
}

/**
 * Builds an exhaustive production recovery table. Every implemented route re-observes exact
 * durable evidence; routes without a truthful reader stop terminally instead of guessing that a
 * timed-out effect happened. The remaining readers are filled as their owning adapters expose
 * durable receipts.
 */
export function createProductionRecoveryProbeTable(
	options: ProductionRuntimeRecoveryProbeOptions,
): ProductionRecoveryProbeTable {
	if (typeof options !== "object" || options === null || typeof options.verification?.runAll !== "function"
		|| typeof options.git?.inspect !== "function" || typeof options.git.resolveBranchHead !== "function"
		|| typeof options.git.resolveRemoteBranchHead !== "function" || typeof options.git.readCommitSubject !== "function"
		|| typeof options.workspace?.findClaim !== "function"
		|| typeof options.workspace.findParentRefreshReceipt !== "function"
		|| typeof options.workspace.findChildHeadReceipt !== "function"
		|| typeof options.agentEffects?.find !== "function"
		|| typeof options.github?.findPullRequests !== "function" || typeof options.github.findChildIntegration !== "function"
		|| typeof options.github.proveAncestry !== "function"
		|| typeof options.reviews?.find !== "function" || typeof options.decisions?.load !== "function"
		|| typeof options.parentMerges?.observeExactPullRequest !== "function"
		|| typeof options.parentHeads?.observe !== "function"
		|| typeof options.dispositionActor !== "string" || !SAFE_ACTOR.test(options.dispositionActor)) {
		throw new Error("production recovery probe options are invalid");
	}
	const now = options.now ?? (() => new Date());
	const timeoutMs = boundedPositive(options.timeoutMs, DEFAULT_EXTERNAL_TIMEOUT_MS, 120_000, "recovery probe timeout");
	const claim: ProductionRecoveryProbe = async (request) => {
		const descriptor = recoveryCommon(request.descriptor, "workspace_claim", request.currentState, request.record.childId);
		const child = recoveryChild(request.currentState, request.record.childId, "workspace claim recovery");
		const ownershipId = productionWorkspaceOwnershipId(
			request.currentState.parentIssue,
			child.issue,
			child.id,
		);
		if ((descriptor.mode !== "start" && descriptor.mode !== "resume")
			|| (descriptor.mode === "start" && descriptor.ownershipId !== undefined)
			|| (descriptor.mode === "resume" && descriptor.ownershipId !== ownershipId)
			|| descriptor.parentHead === undefined || typeof descriptor.parentHead !== "string" || !SHA.test(descriptor.parentHead)
			|| descriptor.trustedWorktreeRoot !== options.trustedWorktreeRoot
			|| !sameRecoveryValue(descriptor.coordinator, options.coordinator)
			|| !sameRecoveryValue(descriptor.writeScopes, child.writeScopes)) {
			throw new ProductionLifecycleError(
				"terminal",
				"workspace claim recovery descriptor moved from immutable ownership coordinates",
				["recovery_workspace_claim_moved"],
			);
		}
		const evidence = await options.workspace.findClaim({
			coordinator: options.coordinator,
			trustedWorktreeRoot: options.trustedWorktreeRoot,
			issue: child.issue,
			slug: child.slug,
			parentIssue: request.currentState.parentIssue,
			parentBranch: request.currentState.parentBranch,
			parentHead: descriptor.parentHead,
			ownershipId,
			allowedScopes: child.writeScopes,
			leaseMode: descriptor.mode,
		});
		if (evidence === undefined) return { status: "absent" };
		assertScopedStatus(
			evidence.changedScope.map((path) => ({ path })),
			child.writeScopes,
			"recovered workspace claim",
		);
		const binding: ProductionWorkspaceBinding = {
			claimId: evidence.claimId,
			ownershipId,
			repositoryIdentity: evidence.repositoryIdentity,
			worktreeIdentity: evidence.worktreeIdentity,
			cwd: evidence.cwd,
			branch: evidence.branch,
			baseBranch: evidence.baseBranch,
			baseHead: evidence.baseHead,
			head: evidence.head,
			writeScopes: [...evidence.writeScopes],
		};
		if (descriptor.mode === "resume" && child.ownership !== undefined
			&& !sameRecoveryValue(binding, child.ownership)) {
			throw new ProductionLifecycleError("terminal", "resumed workspace recovery ownership moved", ["recovery_workspace_moved"]);
		}
		return {
			status: "applied",
			resultDigest: recoveryDigest(binding),
			projectedState: projectRecoveredChild(request, {
				summary: "isolated production workspace claimed",
				effectKey: request.record.key,
				effectKeys: [request.record.key],
				workspace: binding,
			}, "implementation", now),
		};
	};
	const agent = (
		kind: "agent_implementation" | "agent_correction",
		role: "implementation" | "correction",
	): ProductionRecoveryProbe => async (request) => {
		const descriptor = recoveryCommon(request.descriptor, kind, request.currentState, request.record.childId);
		const workspace = recoveryWorkspace(descriptor.workspace, `${kind} recovery workspace`);
		const child = recoveryChild(request.currentState, request.record.childId, `${kind} recovery`);
		const durable = child.checkpoint?.workspace ?? child.ownership;
		if (durable === undefined || !sameRecoveryValue(durable, workspace)) {
			throw new ProductionLifecycleError("terminal", `${kind} recovery workspace moved`, ["recovery_workspace_moved"]);
		}
		const receipt = await options.agentEffects.find(request.record.key);
		if (receipt === undefined) return { status: "absent" };
		if (receipt.start.effectKey !== request.record.key || receipt.start.claimId !== workspace.claimId
			|| receipt.start.role !== role || !sameRecoveryValue(receipt.start.binding, workspace)) {
			throw new ProductionLifecycleError("terminal", `${kind} start receipt conflicts`, ["recovery_agent_receipt_moved"]);
		}
		if (receipt.completion === undefined) {
			throw new ProductionLifecycleError(
				"terminal",
				`${kind} began but lacks an exact completion receipt`,
				["recovery_agent_effect_ambiguous"],
			);
		}
		const completion = receipt.completion;
		if (completion.effectKey !== request.record.key || completion.claimId !== workspace.claimId
			|| completion.role !== role || !sameRecoveryValue(completion.binding, workspace)
			|| !sameRecoveryValue(completion.completedBinding, workspace)) {
			throw new ProductionLifecycleError("terminal", `${kind} completion receipt conflicts`, ["recovery_agent_receipt_moved"]);
		}
		const value = { workspace: structuredClone(completion.completedBinding) };
		if (completion.resultDigest !== recoveryDigest(value)) {
			throw new ProductionLifecycleError("terminal", `${kind} completion digest conflicts`, ["recovery_result_mismatch"]);
		}
		if (kind === "agent_correction" && descriptor.target !== undefined) {
			const target = descriptor.target as IndependentReviewTarget;
			const findings = descriptor.findings;
			if (!Array.isArray(findings) || findings.length === 0 || findings.some((finding) => typeof finding !== "string")) {
				throw new ProductionLifecycleError("terminal", "correction recovery findings are malformed", ["recovery_descriptor_invalid"]);
			}
			createIndependentReviewWork(target);
			const lookup = await options.reviews.find(target, probeContext(request.signal, timeoutMs));
			if (!lookup.complete) throw new Error("production correction recovery review lookup is incomplete");
			const exact = lookup.items.filter((artifact) => sameReviewTarget(artifact.review, target));
			if (exact.length !== 1) {
				throw new ProductionLifecycleError(
					"terminal",
					"correction recovery disposition evidence is absent or ambiguous",
					["recovery_correction_disposition_ambiguous"],
				);
			}
			const artifact = exact[0];
			const reviewRecord = validateIndependentReviewRecord(artifact.review);
			validateAgentSessionAttestation(artifact.attestation, reviewRecord);
			const expected = reviewRecord.findings.filter((finding) => findings.includes(finding.summary));
			if (expected.length !== findings.length || artifact.dispositions.length !== expected.length
				|| expected.some((finding) => !artifact.dispositions.some((disposition) =>
					disposition.findingId === finding.id && disposition.kind === "fixed"
					&& disposition.actor === options.dispositionActor && disposition.headSha === reviewRecord.headSha))) {
				throw new ProductionLifecycleError(
					"terminal",
					"correction recovery lacks exact fixed finding dispositions",
					["recovery_correction_disposition_missing"],
				);
			}
		}
		const checkpoint: ProductionStageCheckpoint = {
			summary: kind === "agent_implementation"
				? "production implementation AgentSession completed"
				: "bounded correction completed; verification and independent review invalidated",
			effectKey: request.record.key,
			effectKeys: [request.record.key],
			workspace,
			...(kind === "agent_correction" && (descriptor.target as IndependentReviewTarget | undefined)?.pullRequest !== undefined
				? { pullRequest: (descriptor.target as IndependentReviewTarget).pullRequest } : {}),
		};
		return {
			status: "applied",
			resultDigest: completion.resultDigest,
			projectedState: kind === "agent_implementation"
				? projectRecoveredChild(request, checkpoint, "verification", now)
				: projectRecoveredCorrection(request, checkpoint, now),
		};
	};
	const shell: ProductionRecoveryProbe = async (request) => {
		const descriptor = recoveryCommon(request.descriptor, "shell_verification", request.currentState, request.record.childId);
		const workspace = descriptor.workspace as ProductionWorkspaceBinding;
		const commands = descriptor.commands as ProductionVerificationCommand[];
		if (typeof workspace !== "object" || workspace === null || typeof workspace.cwd !== "string"
			|| !Array.isArray(commands) || commands.length === 0) {
			throw new ProductionLifecycleError("terminal", "shell recovery descriptor is incomplete", ["recovery_descriptor_invalid"]);
		}
		const results = await options.verification.runAll(workspace.cwd, commands, request.signal);
		if (results.length < 1 || results.length > commands.length
			|| results.some((result, index) => result.id !== commands[index]?.id)
			|| results.some((result) => result.status === "failed" && result.failureKind === undefined)) {
			throw new ProductionLifecycleError("terminal", "shell recovery evidence is malformed", ["verification_evidence_invalid"]);
		}
		const stable = results.map((result) => ({
			id: result.id,
			status: result.status,
			...(result.failureKind === undefined ? {} : { failureKind: result.failureKind }),
		}));
		const passed = stable.length === commands.length && stable.every((result) => result.status === "passed");
		const checkpoint: ProductionStageCheckpoint = {
			summary: passed
				? "all bounded production verification commands passed"
				: "bounded production verification failed and requires correction",
			effectKey: request.record.key,
			effectKeys: [request.record.key],
			workspace,
			verification: {
				status: passed ? "passed" : "failed",
				resultDigest: recoveryDigest(stable),
				commands: stable,
			},
		};
		return {
			status: "applied",
			resultDigest: recoveryDigest(stable),
			projectedState: projectRecoveredChild(
				request,
				checkpoint,
				passed ? "publication" : "correction",
				now,
			),
		};
	};
	const commit: ProductionRecoveryProbe = async (request) => {
		const descriptor = recoveryCommon(request.descriptor, "git_commit", request.currentState, request.record.childId);
		const workspace = recoveryWorkspace(descriptor.workspace, "commit recovery workspace");
		if (descriptor.issue !== recoveryChild(request.currentState, request.record.childId, "commit recovery").issue
			|| descriptor.slug !== recoveryChild(request.currentState, request.record.childId, "commit recovery").slug
			|| typeof descriptor.message !== "string" || descriptor.message.length === 0) {
			throw new ProductionLifecycleError("terminal", "commit recovery descriptor is invalid", ["recovery_descriptor_invalid"]);
		}
		const binding = await options.git.inspect(workspace.cwd);
		if (binding.cwd !== workspace.cwd || binding.repositoryIdentity !== workspace.repositoryIdentity
			|| binding.worktreeIdentity !== workspace.worktreeIdentity
			|| await options.git.currentBranch(binding) !== workspace.branch) {
			throw new ProductionLifecycleError("terminal", "commit recovery worktree moved", ["recovery_workspace_moved"]);
		}
		const head = await options.git.resolveBranchHead(binding, workspace.branch);
		if (head === workspace.head) {
			const value = { committed: false, previousHead: workspace.head, head: workspace.head };
			if (request.record.phase !== "observed") return { status: "absent" };
			if (request.record.resultDigest !== recoveryDigest(value)) {
				throw new ProductionLifecycleError("terminal", "observed no-op commit digest conflicts", ["recovery_result_mismatch"]);
			}
			return {
				status: "applied",
				resultDigest: request.record.resultDigest,
				projectedState: projectRecoveredChild(request, {
					summary: "verified no-op child commit recovered",
					effectKey: request.record.key,
					effectKeys: [request.record.key],
					workspace,
				}, "publication", now),
			};
		}
		if (!SHA.test(head) || !(await options.git.isAncestor(binding, workspace.head, head))) {
			throw new ProductionLifecycleError("terminal", "commit recovery history is ambiguous", ["recovery_commit_ambiguous"]);
		}
		const expectedSubject = `${descriptor.message} [shepherd-effect:${request.record.key}]`;
		if (await options.git.readCommitSubject(binding, head) !== expectedSubject) {
			throw new ProductionLifecycleError("terminal", "commit recovery head lacks the exact effect marker", ["recovery_commit_ambiguous"]);
		}
		const value = { committed: true, previousHead: workspace.head, head };
		const after = { ...workspace, head };
		return {
			status: "applied",
			resultDigest: recoveryDigest(value),
			projectedState: projectRecoveredChild(request, {
				summary: "verified child commit recovered from exact effect marker",
				effectKey: request.record.key,
				effectKeys: [request.record.key],
				workspace: after,
			}, "publication", now),
		};
	};
	const push: ProductionRecoveryProbe = async (request) => {
		const descriptor = recoveryCommon(request.descriptor, "git_push", request.currentState, request.record.childId);
		const workspace = recoveryWorkspace(descriptor.workspace, "push recovery workspace");
		if (descriptor.branch !== workspace.branch || descriptor.head !== workspace.head) {
			throw new ProductionLifecycleError("terminal", "push recovery descriptor moved", ["recovery_descriptor_moved"]);
		}
		const binding = await options.git.inspect(workspace.cwd);
		if (binding.cwd !== workspace.cwd || binding.repositoryIdentity !== workspace.repositoryIdentity
			|| binding.worktreeIdentity !== workspace.worktreeIdentity
			|| await options.git.currentBranch(binding) !== workspace.branch
			|| await options.git.resolveBranchHead(binding, workspace.branch) !== workspace.head) {
			throw new ProductionLifecycleError("terminal", "push recovery worktree moved", ["recovery_workspace_moved"]);
		}
		const remoteHead = await options.git.resolveRemoteBranchHead(binding, workspace.branch);
		if (remoteHead === undefined) return { status: "absent" };
		if (remoteHead !== workspace.head) {
			throw new ProductionLifecycleError("terminal", "push recovery remote branch is ambiguous", ["recovery_push_ambiguous"]);
		}
		const value = { branch: workspace.branch, head: workspace.head, remoteName: "origin" as const };
		return {
			status: "applied",
			resultDigest: recoveryDigest(value),
			projectedState: projectRecoveredChild(request, {
				summary: "exact child branch push recovered",
				effectKey: request.record.key,
				effectKeys: [request.record.key],
				workspace,
			}, "publication", now),
		};
	};
	const pullRequest: ProductionRecoveryProbe = async (request) => {
		const descriptor = recoveryCommon(request.descriptor, "child_pull_request", request.currentState, request.record.childId);
		const child = recoveryChild(request.currentState, request.record.childId, "pull request recovery");
		if (typeof descriptor.marker !== "string" || descriptor.marker.length === 0
			|| typeof descriptor.branch !== "string" || typeof descriptor.baseBranch !== "string"
			|| typeof descriptor.baseHead !== "string" || !SHA.test(descriptor.baseHead)
			|| typeof descriptor.head !== "string" || !SHA.test(descriptor.head)
			|| !Array.isArray(descriptor.changedScope)
			|| descriptor.changedScope.some((path) => typeof path !== "string")) {
			throw new ProductionLifecycleError("terminal", "pull request recovery descriptor is invalid", ["recovery_descriptor_invalid"]);
		}
		const lookup = await options.github.findPullRequests({
			repository: request.currentState.repository,
			marker: descriptor.marker,
		}, probeContext(request.signal, timeoutMs));
		if (!lookup.complete) throw new Error("production recovery pull request lookup is incomplete");
		if (lookup.items.length === 0) return { status: "absent" };
		if (lookup.items.length !== 1) {
			throw new ProductionLifecycleError("terminal", "pull request recovery evidence is ambiguous", ["recovery_pull_request_ambiguous"]);
		}
		const evidence = validateGitHubPullRequestEvidence(lookup.items[0]);
		if (evidence.repository !== request.currentState.repository || evidence.workItemId !== child.id
			|| evidence.generation !== request.currentState.resourceGeneration || evidence.marker !== descriptor.marker
			|| evidence.state !== "open" || evidence.draft || evidence.baseBranch !== descriptor.baseBranch
			|| evidence.headBranch !== descriptor.branch || evidence.baseSha !== descriptor.baseHead
			|| evidence.headSha !== descriptor.head
			|| !sameRecoveryValue([...evidence.changedPaths].sort(), [...descriptor.changedScope as string[]].sort())
			|| !sameRecoveryValue([...evidence.allowedScopes].sort(), [...child.writeScopes].sort())) {
			throw new ProductionLifecycleError("terminal", "pull request recovery evidence moved", ["recovery_pull_request_moved"]);
		}
		const currentWorkspace = recoveredWorkspaceCheckpoint(request.currentState, request.record.childId, "pull request recovery");
		const workspace = { ...currentWorkspace, head: descriptor.head };
		if (workspace.branch !== descriptor.branch || workspace.baseBranch !== descriptor.baseBranch
			|| workspace.baseHead !== descriptor.baseHead) {
			throw new ProductionLifecycleError("terminal", "pull request recovery workspace moved", ["recovery_workspace_moved"]);
		}
		const binding = await options.git.inspect(currentWorkspace.cwd);
		if (binding.cwd !== currentWorkspace.cwd || binding.repositoryIdentity !== currentWorkspace.repositoryIdentity
			|| binding.worktreeIdentity !== currentWorkspace.worktreeIdentity
			|| await options.git.currentBranch(binding) !== descriptor.branch
			|| await options.git.resolveBranchHead(binding, descriptor.branch as string) !== descriptor.head
			|| await options.git.resolveRemoteBranchHead(binding, descriptor.branch as string) !== descriptor.head) {
			throw new ProductionLifecycleError(
				"terminal",
				"pull request recovery local or remote branch moved from exact publication head",
				["recovery_pull_request_branch_moved"],
			);
		}
		return {
			status: "applied",
			resultDigest: recoveryDigest(evidence),
			projectedState: projectRecoveredChild(request, {
				summary: "verified child commit pushed and exact stacked pull request published",
				effectKey: request.record.key,
				effectKeys: [request.record.key],
				workspace,
				pullRequest: evidence.number,
			}, "review", now),
		};
	};
	const review: ProductionRecoveryProbe = async (request) => {
		const descriptor = recoveryCommon(request.descriptor, "independent_review", request.currentState, request.record.childId);
		const target = descriptor.target as IndependentReviewTarget;
		createIndependentReviewWork(target);
		const lookup = await options.reviews.find(target, probeContext(request.signal, timeoutMs));
		if (lookup.complete !== true) throw new Error("production recovery review lookup is incomplete");
		const exact = lookup.items.filter((artifact) => sameReviewTarget(artifact.review, target)).map((artifact) => {
			const reviewRecord = validateIndependentReviewRecord(artifact.review);
			validateAgentSessionAttestation(artifact.attestation, reviewRecord);
			return { artifact, review: reviewRecord };
		}).sort((left, right) => right.review.completedAt.localeCompare(left.review.completedAt)
			|| independentReviewResultDigest(left.review).localeCompare(independentReviewResultDigest(right.review)));
		if (exact.length === 0) return { status: "absent" };
		if (exact.length > 1 && exact[0].review.completedAt === exact[1].review.completedAt
			&& independentReviewResultDigest(exact[0].review) !== independentReviewResultDigest(exact[1].review)) {
			throw new Error("production recovery review evidence is ambiguous");
		}
		const artifact = exact[0].artifact;
		const reviewRecord = exact[0].review;
		const resultDigest = independentReviewResultDigest(reviewRecord);
		const currentWorkspace = recoveredWorkspaceCheckpoint(request.currentState, request.record.childId, "review recovery");
		if (currentWorkspace.branch !== target.headBranch || currentWorkspace.baseBranch !== target.baseBranch
			|| currentWorkspace.baseHead !== target.baseSha) {
			throw new ProductionLifecycleError("terminal", "review recovery workspace moved", ["recovery_workspace_moved"]);
		}
		const workspace = { ...currentWorkspace, head: target.headSha };
		const checkpoint: ProductionStageCheckpoint = {
			summary: reviewRecord.verdict === "clean"
				? "independent exact-head review is clean"
				: "independent review returned findings",
			effectKey: request.record.key,
			effectKeys: [request.record.key],
			workspace,
			pullRequest: target.pullRequest,
			review: {
				status: reviewRecord.verdict === "clean" ? "clean" : "blocked",
				baseHead: reviewRecord.baseSha,
				head: reviewRecord.headSha,
				resultDigest,
				...(reviewRecord.verdict === "clean"
					? { authorizationDigest: independentReviewAuthorizationDigest(reviewRecord) } : {}),
				completedAt: reviewRecord.completedAt,
				findings: reviewRecord.findings.map((finding) => ({ id: finding.id, summary: finding.summary })),
			},
		};
		return {
			status: "applied",
			resultDigest: recoveryDigest(artifact),
			projectedState: projectRecoveredChild(
				request,
				checkpoint,
				reviewRecord.verdict === "clean" ? "integration" : "correction",
				now,
			),
		};
	};
	const integration: ProductionRecoveryProbe = async (request) => {
		const descriptor = recoveryCommon(request.descriptor, "child_integration", request.currentState, request.record.childId);
		const child = recoveryChild(request.currentState, request.record.childId, "child integration recovery");
		if (typeof descriptor.marker !== "string" || descriptor.marker.length === 0
			|| !Number.isSafeInteger(descriptor.pullRequest) || (descriptor.pullRequest as number) < 1
			|| typeof descriptor.baseHead !== "string" || !SHA.test(descriptor.baseHead)
			|| typeof descriptor.head !== "string" || !SHA.test(descriptor.head)
			|| typeof descriptor.reviewResultDigest !== "string") {
			throw new ProductionLifecycleError("terminal", "child integration recovery descriptor is invalid", ["recovery_descriptor_invalid"]);
		}
		const lookup = await options.github.findChildIntegration({
			repository: request.currentState.repository,
			childId: child.id,
			marker: descriptor.marker,
		}, probeContext(request.signal, timeoutMs));
		if (!lookup.complete) throw new Error("production recovery child integration lookup is incomplete");
		if (lookup.items.length === 0) return { status: "absent" };
		if (lookup.items.length !== 1) {
			throw new ProductionLifecycleError("terminal", "child integration recovery evidence is ambiguous", ["recovery_integration_ambiguous"]);
		}
		const receipt = validateChildIntegrationReceipt(lookup.items[0]);
		if (receipt.childId !== child.id || receipt.generation !== request.currentState.resourceGeneration
			|| receipt.marker !== descriptor.marker || receipt.pullRequest !== descriptor.pullRequest
			|| receipt.baseSha !== descriptor.baseHead || receipt.headSha !== descriptor.head
			|| receipt.parentBranch !== request.currentState.parentBranch
			|| receipt.controllerProvenance.planDigest !== request.currentState.planDigest
			|| receipt.controllerProvenance.reviewResultDigest !== descriptor.reviewResultDigest) {
			throw new ProductionLifecycleError("terminal", "child integration recovery receipt moved", ["recovery_integration_moved"]);
		}
		const currentWorkspace = recoveredWorkspaceCheckpoint(request.currentState, request.record.childId, "child integration recovery");
		const workspace = { ...currentWorkspace, head: descriptor.head as string };
		const durableReview = child.checkpoint?.review;
		if (workspace.baseHead !== descriptor.baseHead || workspace.baseBranch !== request.currentState.parentBranch
			|| (child.checkpoint?.pullRequest !== undefined && child.checkpoint.pullRequest !== descriptor.pullRequest)
			|| (durableReview !== undefined && (durableReview.status !== "clean"
				|| durableReview.baseHead !== descriptor.baseHead || durableReview.head !== descriptor.head
				|| durableReview.resultDigest !== descriptor.reviewResultDigest))) {
			throw new ProductionLifecycleError("terminal", "child integration recovery workspace moved", ["recovery_workspace_moved"]);
		}
		const parentHead = await options.parentHeads.observe({
			repository: request.currentState.repository,
			parentBranch: request.currentState.parentBranch,
		}, request.signal);
		if (parentHead.repository !== request.currentState.repository || parentHead.branch !== request.currentState.parentBranch
			|| !SHA.test(parentHead.head) || parentHead.head === descriptor.baseHead) {
			throw new ProductionLifecycleError("retryable", "recovered integration lacks an advanced parent head", ["parent_head_not_advanced"]);
		}
		const ancestry = await options.github.proveAncestry({
			repository: request.currentState.repository,
			ancestorSha: descriptor.baseHead,
			descendantSha: parentHead.head,
		}, probeContext(request.signal, timeoutMs));
		if (ancestry.repository !== request.currentState.repository || ancestry.ancestorSha !== descriptor.baseHead
			|| ancestry.descendantSha !== parentHead.head || ancestry.result !== true) {
			throw new ProductionLifecycleError("terminal", "recovered integration parent head lacks exact ancestry", ["recovery_parent_head_moved"]);
		}
		const value = { kind: "integrated" as const, receipt };
		const recoveredReview: NonNullable<ProductionStageCheckpoint["review"]> = durableReview ?? {
			status: "clean",
			baseHead: receipt.baseSha,
			head: receipt.headSha,
			resultDigest: receipt.controllerProvenance.reviewResultDigest,
			authorizationDigest: receipt.controllerProvenance.reviewAuthorizationDigest,
			completedAt: receipt.controllerProvenance.reviewCompletedAt,
			findings: [],
		};
		return {
			status: "applied",
			resultDigest: recoveryDigest(value),
			projectedState: projectRecoveredChild(request, {
				summary: "exact reviewed child integration reconciled into the non-default parent branch",
				effectKey: request.record.key,
				effectKeys: [request.record.key],
				workspace,
				pullRequest: receipt.pullRequest,
				review: recoveredReview,
				integrationReceiptDigest: productionChildIntegrationReceiptDigest(receipt),
				parentHead: parentHead.head,
			}, "succeeded", now, { succeeded: true }),
		};
	};
	const refresh: ProductionRecoveryProbe = async (request) => {
		const descriptor = recoveryCommon(request.descriptor, "parent_refresh", request.currentState, request.record.childId);
		const child = recoveryChild(request.currentState, request.record.childId, "parent refresh recovery");
		const previous = recoveryWorkspace(descriptor.workspace, "parent refresh recovery workspace");
		if (descriptor.previousBaseHead !== previous.baseHead || descriptor.previousHead !== previous.head
			|| typeof descriptor.newParentHead !== "string" || !SHA.test(descriptor.newParentHead)
			|| child.ownership === undefined || !sameRecoveryValue(child.ownership, previous)) {
			throw new ProductionLifecycleError("terminal", "parent refresh recovery descriptor moved", ["recovery_descriptor_moved"]);
		}
		const receipt = await options.workspace.findParentRefreshReceipt({
			trustedWorktreeRoot: options.trustedWorktreeRoot,
			issue: child.issue,
			claimId: previous.claimId,
			effectKey: request.record.key,
		});
		if (receipt === undefined) return { status: "absent" };
		if (receipt.previousBaseHead !== previous.baseHead || receipt.baseHead !== descriptor.newParentHead
			|| receipt.previousHead !== previous.head || !SHA.test(receipt.head)
			|| receipt.verificationInvalidated !== true || receipt.reviewInvalidated !== true) {
			throw new ProductionLifecycleError("terminal", "parent refresh recovery receipt moved", ["recovery_refresh_moved"]);
		}
		const ownershipId = productionWorkspaceOwnershipId(request.currentState.parentIssue, child.issue, child.id);
		const claimEvidence = await options.workspace.findClaim({
			coordinator: options.coordinator,
			trustedWorktreeRoot: options.trustedWorktreeRoot,
			issue: child.issue,
			slug: child.slug,
			parentIssue: request.currentState.parentIssue,
			parentBranch: request.currentState.parentBranch,
			parentHead: descriptor.newParentHead,
			ownershipId,
			allowedScopes: child.writeScopes,
			leaseMode: "resume",
		});
		if (claimEvidence === undefined) {
			throw new ProductionLifecycleError("terminal", "parent refresh receipt lacks its resulting claim", ["recovery_refresh_claim_missing"]);
		}
		const binding: ProductionWorkspaceBinding = {
			claimId: claimEvidence.claimId,
			ownershipId,
			repositoryIdentity: claimEvidence.repositoryIdentity,
			worktreeIdentity: claimEvidence.worktreeIdentity,
			cwd: claimEvidence.cwd,
			branch: claimEvidence.branch,
			baseBranch: claimEvidence.baseBranch,
			baseHead: claimEvidence.baseHead,
			head: claimEvidence.head,
			writeScopes: [...claimEvidence.writeScopes],
		};
		if (binding.baseHead !== receipt.baseHead || binding.head !== receipt.head
			|| binding.ownershipId !== previous.ownershipId) {
			throw new ProductionLifecycleError("terminal", "parent refresh resulting workspace moved", ["recovery_refresh_claim_moved"]);
		}
		return {
			status: "applied",
			resultDigest: recoveryDigest(receipt),
			projectedState: refreshProductionChildOwnership(
				request.currentState,
				recoveryFence(request.currentState),
				{
					childId: child.id,
					outcome: binding.claimId === previous.claimId ? "rebased" : "reclaimed",
					previousClaimId: previous.claimId,
					previousBaseHead: previous.baseHead,
					newBinding: binding,
					effectKey: request.record.key,
					summary: "workspace refreshed to the authoritative parent head; verification and review required",
					now: now(),
				},
			),
		};
	};
	const reconcileHead: ProductionRecoveryProbe = async (request) => {
		const descriptor = recoveryCommon(request.descriptor, "child_head_reconciliation", request.currentState, request.record.childId);
		const child = recoveryChild(request.currentState, request.record.childId, "child-head recovery");
		const previous = recoveryWorkspace(descriptor.workspace, "child-head recovery workspace");
		if (descriptor.branch !== previous.branch || descriptor.baseHead !== previous.baseHead
			|| descriptor.previousHead !== previous.head || descriptor.pullRequest !== child.checkpoint?.pullRequest
			|| child.ownership === undefined) {
			throw new ProductionLifecycleError("terminal", "child-head recovery descriptor moved", ["recovery_descriptor_moved"]);
		}
		const receipt = await options.workspace.findChildHeadReceipt({
			trustedWorktreeRoot: options.trustedWorktreeRoot,
			issue: child.issue,
			claimId: previous.claimId,
			effectKey: request.record.key,
		});
		if (receipt === undefined) return { status: "absent" };
		if (receipt.branch !== previous.branch || receipt.baseHead !== previous.baseHead
			|| receipt.previousHead !== previous.head || !SHA.test(receipt.head) || receipt.head === previous.head
			|| receipt.verificationInvalidated !== true || receipt.reviewInvalidated !== true
			|| receipt.integrationInvalidated !== true) {
			throw new ProductionLifecycleError("terminal", "child-head recovery receipt moved", ["recovery_child_head_moved"]);
		}
		assertScopedStatus(receipt.changedScope.map((path) => ({ path })), child.writeScopes, "child-head recovery");
		const binding = { ...previous, head: receipt.head };
		const checkpoint: ProductionStageCheckpoint = {
			summary: "authoritative child head reclaimed; verification and exact-head review required",
			effectKey: request.record.key,
			effectKeys: [request.record.key],
			workspace: binding,
			pullRequest: descriptor.pullRequest as number,
		};
		return {
			status: "applied",
			resultDigest: recoveryDigest(receipt),
			projectedState: reconcileProductionChildHead(
				request.currentState,
				recoveryFence(request.currentState),
				{
					childId: child.id,
					previousHead: previous.head,
					head: receipt.head,
					checkpoint,
					now: now(),
				},
			),
		};
	};
	const humanRequest: ProductionRecoveryProbe = async (request) => {
		const descriptor = recoveryCommon(request.descriptor, "human_request", request.currentState, request.record.childId);
		const child = recoveryChild(request.currentState, request.record.childId, "human request recovery");
		const decisionRequest = recoveryRecord(descriptor.request, "human request recovery request");
		if (typeof decisionRequest.requestId !== "string" || decisionRequest.gate !== "scope"
			|| !sameRecoveryValue(decisionRequest.allowedOptions, ["authorize-one-retry", "abort-child"])
			|| typeof decisionRequest.question !== "string") {
			throw new ProductionLifecycleError("terminal", "human request recovery descriptor is invalid", ["recovery_descriptor_invalid"]);
		}
		const reasonMatch = /^\[(retry_budget_exhausted|correction_budget_exhausted)\] /u.exec(decisionRequest.question);
		if (reasonMatch === null) {
			throw new ProductionLifecycleError("terminal", "human request recovery reason is absent", ["recovery_descriptor_invalid"]);
		}
		const binding: HumanDecisionBinding = {
			repository: request.currentState.repository,
			target: { kind: "issue", number: child.issue },
			generation: request.currentState.resourceGeneration,
		};
		const stored = await options.decisions.load(decisionRequest.requestId);
		if (stored === null) return { status: "absent" };
		const record = validateHumanDecisionRecord(stored, now());
		assertHumanDecisionBinding(record, binding);
		if (record.requestId !== decisionRequest.requestId || record.gate !== "scope"
			|| !sameRecoveryValue(record.allowedOptions, decisionRequest.allowedOptions)
			|| !sameRecoveryValue(record.actorAllowlist, decisionRequest.actorAllowlist)
			|| record.expiresAt !== decisionRequest.expiresAt || record.question !== decisionRequest.question) {
			throw new ProductionLifecycleError("terminal", "human request recovery record moved", ["recovery_human_request_moved"]);
		}
		return {
			status: "applied",
			resultDigest: recoveryDigest(record),
			projectedState: waitForProductionChildIntervention(
				request.currentState,
				recoveryFence(request.currentState),
				{
					childId: child.id,
					requestId: record.requestId,
					reason: reasonMatch[1],
					now: now(),
				},
			),
		};
	};
	const humanConsume: ProductionRecoveryProbe = async (request) => {
		const descriptor = recoveryCommon(request.descriptor, "human_consume", request.currentState, request.record.childId);
		const child = recoveryChild(request.currentState, request.record.childId, "human consume recovery");
		if (typeof descriptor.requestId !== "string"
			|| (descriptor.option !== "authorize-one-retry" && descriptor.option !== "abort-child")) {
			throw new ProductionLifecycleError("terminal", "human consume recovery descriptor is invalid", ["recovery_descriptor_invalid"]);
		}
		const binding = recoveryRecord(descriptor.binding, "human consume recovery binding") as unknown as HumanDecisionBinding;
		const stored = await options.decisions.load(descriptor.requestId);
		if (stored === null) return { status: "absent" };
		const record = validateHumanDecisionRecord(stored, now());
		assertHumanDecisionBinding(record, binding);
		if (record.requestId !== descriptor.requestId || record.gate !== "scope"
			|| !sameRecoveryValue(record.allowedOptions, ["authorize-one-retry", "abort-child"])) {
			throw new ProductionLifecycleError("terminal", "human consume recovery record moved", ["recovery_human_consume_moved"]);
		}
		if (record.status !== "consumed") return { status: "absent" };
		if (record.decision?.option !== descriptor.option) {
			throw new ProductionLifecycleError("terminal", "human consume recovery option moved", ["recovery_human_consume_moved"]);
		}
		let projectedState: ProductionAutonomousState;
		if (descriptor.option === "authorize-one-retry") {
			projectedState = authorizeProductionChildRetry(
				request.currentState,
				recoveryFence(request.currentState),
				{ childId: child.id, requestId: descriptor.requestId, now: now() },
			);
		} else {
			projectedState = evolveProductionState(request.currentState, recoveryFence(request.currentState), (draft) => {
				if (draft.childGate?.childId !== child.id || draft.childGate.requestId !== descriptor.requestId) {
					throw new Error("human consume recovery gate moved");
				}
				draft.childGate.status = "aborted";
				draft.status = "failed";
				draft.stage = "blocked";
				draft.terminalBlocker = "human aborted the exhausted child";
				const target = draft.children.find((candidate) => candidate.id === child.id)!;
				target.status = "failed";
				target.stage = "failed";
			}, now());
		}
		return { status: "applied", resultDigest: recoveryDigest(record), projectedState };
	};
	const parentMerge: ProductionRecoveryProbe = async (request) => {
		const descriptor = recoveryRecord(request.descriptor, "parent merge recovery descriptor");
		const state = request.currentState;
		const gate = state.humanGate;
		if (descriptor.operation !== "parent_merge_observation" || descriptor.parentIssue !== state.parentIssue
			|| descriptor.repository !== state.repository || descriptor.planId !== state.planId
			|| descriptor.planDigest !== state.planDigest || descriptor.parentBranch !== state.parentBranch
			|| descriptor.parentBaseBranch !== state.parentBaseBranch || descriptor.runId !== state.runId
			|| descriptor.resourceGeneration !== state.resourceGeneration || descriptor.generation !== state.generation
			|| descriptor.stateRevision !== state.revision || gate?.status !== "pending"
			|| descriptor.pullRequest !== gate.pullRequest || descriptor.requestId !== gate.requestId
			|| descriptor.head !== gate.head) {
			throw new ProductionLifecycleError("terminal", "parent merge recovery descriptor moved", ["recovery_descriptor_moved"]);
		}
		const binding: HumanDecisionBinding = {
			repository: state.repository,
			target: { kind: "pull_request", number: gate.pullRequest },
			generation: state.generation,
			headSha: gate.head,
		};
		const stored = await options.decisions.load(gate.requestId);
		if (stored === null) return { status: "absent" };
		const record = validateHumanDecisionRecord(stored, now());
		assertHumanDecisionBinding(record, binding);
		if (record.requestId !== gate.requestId || record.gate !== "parent_merge"
			|| !sameRecoveryValue(record.allowedOptions, ["approve-merge", "reject"])) {
			throw new ProductionLifecycleError("terminal", "parent merge recovery decision moved", ["recovery_parent_decision_moved"]);
		}
		let observation: Record<string, unknown>;
		if (record.status === "pending" || record.status === "expired") {
			observation = { status: "pending" };
		} else if (record.status === "decided") {
			return { status: "absent" };
		} else if (record.status !== "consumed" || record.decision === undefined) {
			throw new ProductionLifecycleError("terminal", "parent merge recovery decision is malformed", ["recovery_parent_decision_moved"]);
		} else if (record.decision.option === "reject") {
			observation = { status: "rejected" };
		} else {
			if (record.decision.option !== "approve-merge") {
				throw new ProductionLifecycleError("terminal", "parent merge recovery option is unauthorized", ["recovery_parent_decision_moved"]);
			}
			const exact = await options.parentMerges.observeExactPullRequest({
				repository: state.repository,
				pullRequest: gate.pullRequest,
				headSha: gate.head,
			}, probeContext(request.signal, timeoutMs));
			if (exact.repository !== state.repository || exact.pullRequest !== gate.pullRequest
				|| typeof exact.headSha !== "string" || !SHA.test(exact.headSha)
				|| !Number.isSafeInteger(exact.revision) || exact.revision < 1) {
				throw new ProductionLifecycleError("terminal", "parent merge recovery observation is malformed", ["recovery_parent_merge_moved"]);
			}
			if (exact.headSha !== gate.head) {
				observation = {
					status: "invalidated",
					repository: exact.repository,
					pullRequest: exact.pullRequest,
					previousHead: gate.head,
					currentHead: exact.headSha,
					revision: exact.revision,
					observedAt: exact.observedAt,
				};
			} else if (exact.state !== "merged") {
				if (exact.mergedAt !== null || exact.mergeCommitSha !== null) {
					throw new ProductionLifecycleError("terminal", "parent merge recovery observation is ambiguous", ["recovery_parent_merge_ambiguous"]);
				}
				observation = { status: "approved_waiting_for_merge" };
			} else {
				if (exact.mergedAt === null || exact.mergeCommitSha === null || !SHA.test(exact.mergeCommitSha)) {
					throw new ProductionLifecycleError("terminal", "parent merge recovery receipt is incomplete", ["recovery_parent_merge_ambiguous"]);
				}
				observation = {
					status: "merged",
					repository: exact.repository,
					pullRequest: exact.pullRequest,
					head: gate.head,
					mergedAt: exact.mergedAt,
					mergeCommitSha: exact.mergeCommitSha,
					revision: exact.revision,
					observedAt: exact.observedAt,
				};
			}
		}
		const projectedState = evolveProductionState(state, recoveryFence(state), (draft) => {
			if (observation.status === "pending" || observation.status === "approved_waiting_for_merge") return;
			if (observation.status === "invalidated") {
				const active = draft.humanGate!;
				draft.invalidatedParentGates = [...(draft.invalidatedParentGates ?? []), {
					...active,
					status: "invalidated",
					invalidationEvidence: {
						currentHead: observation.currentHead as string,
						revision: observation.revision as number,
						observedAt: observation.observedAt as string,
					},
				}];
				delete draft.humanGate;
				draft.status = "running";
				draft.stage = "schedule";
				return;
			}
			if (observation.status === "merged") {
				draft.humanGate!.status = "merged";
				draft.humanGate!.mergeEvidence = {
					mergedAt: observation.mergedAt as string,
					mergeCommitSha: observation.mergeCommitSha as string,
					revision: observation.revision as number,
					observedAt: observation.observedAt as string,
				};
				draft.status = "completed";
				draft.stage = "completed";
				return;
			}
			draft.humanGate!.status = "rejected";
			draft.status = "failed";
			draft.stage = "blocked";
			draft.terminalBlocker = "human rejected the exact parent merge";
		}, now());
		return { status: "applied", resultDigest: recoveryDigest(observation), projectedState };
	};
	return {
		workspace_claim: claim,
		agent_implementation: agent("agent_implementation", "implementation"),
		agent_correction: agent("agent_correction", "correction"),
		shell_verification: shell,
		git_commit: commit,
		git_push: push,
		child_pull_request: pullRequest,
		independent_review: review,
		child_integration: integration,
		parent_refresh: refresh,
		child_head_reconciliation: reconcileHead,
		human_request: humanRequest,
		human_consume: humanConsume,
		parent_merge_observation: parentMerge,
	};
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
		parentMergeEffects: options.parentMergeEffects,
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
	if (!Number.isSafeInteger(options.parentIssue) || options.parentIssue < 1) {
		throw new Error("production runtime parent issue must be a positive integer");
	}
	safeAbsolutePath(options.repositoryRoot, "repository root");
	safeAbsolutePath(options.stateRoot, "state root");
	safeAbsolutePath(options.trustedWorktreeRoot, "trusted worktree root");
	if (options.effectRecovery !== undefined) assertProductionRecovery(options.effectRecovery);
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
	const stateStore = new ProductionFileStateStore(stateRoot);
	const effects = new ProductionEffectJournal(stateRoot);
	const decisionRepository = new FileHumanDecisionRepository(join(stateRoot, "human-decisions"));
	const githubDecisionBroker = new GitHubDecisionBroker(
		decisionRepository,
		decisionTransport(githubOptions.execute),
		{ ...(options.decision ?? {}), now },
	);
	const decision = adaptGitHubDecisionBroker(githubDecisionBroker);
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
	const verification = new BoundedVerificationRunner({ executables: { node: process.execPath } });
	const workspaceAdapter = new WorkspaceAdapter(options.git);
	const agentEffects = new ProductionAgentEffectReceiptRepository(trustedWorktreeRoot);
	const workspaceLifecycle = new ProductionWorkspaceLifecycle({
		workspaceAdapter,
		verification,
		agentSession: options.agentSession,
	});
	const finalizer = new ProductionParentFinalizer({
		transport: githubFacade.transport,
		policies: new ProductionParentPolicyAuthority(githubFacade.policySource),
		reviews: reviewer,
		readiness: options.parentReadiness,
	});
	const parentMergeLookup = new GhParentPullRequestMergeLookup(
		githubOptions.execute ?? defaultGhOrchestrationExecutor,
		now,
	);
	const parentGate = new ProductionParentGateAdapter(
		decision,
		parentMergeLookup,
	);
	const parentHeads = new GhProductionParentHeadSource(githubOptions);
	const parentMergeEffects = new ProductionParentMergeEffectJournal({ journal: effects, parentGate });
	const effectRecovery = options.effectRecovery ?? new ProductionEffectRecoveryAuthority({
		stateRoot,
		issue: options.parentIssue,
		stateStore,
		probes: createProductionRecoveryProbeTable({
			git: options.git,
			workspace: workspaceAdapter,
			agentEffects,
			coordinator: options.coordinator,
			trustedWorktreeRoot,
			verification,
			github: githubFacade.transport,
			reviews: reviewRepository,
			decisions: decisionRepository,
			parentMerges: parentMergeLookup,
			parentHeads,
			dispositionActor: options.dispositionActor,
			now,
		}),
	});
	return composeProductionShepherdController({
		stateStore,
		intake: new ProductionRepositoryPlanIntake(repositoryRoot),
		effects,
		effectRecovery,
		workspaceLifecycle,
		github,
		reviewer,
		reviewRepository,
		decisionBroker: decision,
		parentHeads,
		coordinator: options.coordinator,
		trustedWorktreeRoot,
		dispositionActor: options.dispositionActor,
		finalizer,
		parentGate,
		parentMergeEffects,
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

// Recovery records now carry bounded canonical coordinates. The required authority uses those
// coordinates to re-observe external truth and project it; it must never infer success from timeout.
export type ProductionRuntimeRecoveryAuthority = ProductionEffectRecoveryPort;
export type ProductionRuntimeRecoveryRecord = ProductionEffectRecord;
