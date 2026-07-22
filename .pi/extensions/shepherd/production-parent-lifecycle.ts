import { createHash } from "node:crypto";

import {
	validateProductionParentPlan,
	type ProductionParentPlanDocument,
} from "./autonomous-production-contract.ts";
import {
	assertProductionPlanBinding,
	validateProductionAutonomousState,
	type ProductionAutonomousState,
} from "./autonomous-production-state.ts";
import {
	validateChildIntegrationReceipt,
	type AuthoritativeLookup,
	type ChildIntegrationReceipt,
	type ExternalCallContext,
	type GitHubOrchestrationTransport,
	type ParentDecisionBroker,
} from "./github-orchestrator.ts";
import {
	evaluateGitHubPullRequestEvidence,
	validateGitHubChangedPathEvidence,
	validateGitHubPullRequestEvidence,
	validateRequiredGitHubCheckPolicy,
	type GitHubChangedPathEvidence,
	type GitHubPullRequestEvidence,
	type RequiredGitHubCheckPolicy,
} from "./github-evidence.ts";
import { createProductionOrchestrationPlan } from "./production-orchestration-plan.ts";
import type { ProductionReviewArtifact } from "./production-review-adapter.ts";
import {
	readBoundedExactRecord,
	reconcileIndependentReview,
	validateAgentSessionAttestation,
	validateIndependentReviewRecord,
	type AgentSessionAttestation,
	type IndependentReviewTarget,
} from "./review-router.ts";
import {
	ProductionHumanParentMergeGate,
	type AuthoritativeParentMergeState,
	type ParentPullRequestMergeLookup,
	type ProductionParentMergeRequest,
} from "./production-human-gate.ts";
import type {
	ProductionParentFinalization,
	ProductionParentFinalizerPort,
	ProductionParentGateObservation,
	ProductionParentGatePort,
} from "./production-controller.ts";

const SHA = /^[0-9a-f]{40}$/u;
const DIGEST = /^[0-9a-f]{64}$/u;
const MAX_TIMEOUT_MS = 120_000;

type AbortKind = "caller" | "timeout" | "close";

function duration(value: number | undefined, fallback: number, description: string): number {
	const candidate = value ?? fallback;
	if (!Number.isSafeInteger(candidate) || candidate < 1 || candidate > MAX_TIMEOUT_MS) {
		throw new Error(`${description} must be a bounded positive integer`);
	}
	return candidate;
}

function sameStrings(left: readonly string[], right: readonly string[]): boolean {
	return left.length === right.length && left.every((entry, index) => entry === right[index]);
}

function exactTimestamp(value: unknown, description: string): string {
	if (typeof value !== "string" || value.length > 64) throw new Error(`${description} must be a canonical timestamp`);
	const parsed = new Date(value);
	if (!Number.isFinite(parsed.valueOf()) || parsed.toISOString() !== value) {
		throw new Error(`${description} must be a canonical timestamp`);
	}
	return value;
}

function integrationReceiptDigest(value: ChildIntegrationReceipt): string {
	return createHash("sha256").update(JSON.stringify(validateChildIntegrationReceipt(value))).digest("hex");
}

class BoundedParentCalls {
	readonly #name: string;
	readonly #closeTimeoutMs: number;
	readonly #closing = new AbortController();
	readonly #active = new Map<symbol, { controller: AbortController; settlement: Promise<void> }>();
	#closed = false;
	#closePromise: Promise<void> | undefined;

	constructor(name: string, closeTimeoutMs: number) {
		this.#name = name;
		this.#closeTimeoutMs = closeTimeoutMs;
	}

	invoke<T>(operation: string, signal: AbortSignal, timeoutMs: number, call: (context: ExternalCallContext) => Promise<T>): Promise<T> {
		if (this.#closed) return Promise.reject(new Error(`${this.#name} is closed`));
		if (signal.aborted) return Promise.reject(new Error(`${operation} was cancelled by the caller`));
		const key = Symbol(operation);
		const controller = new AbortController();
		let abortKind: AbortKind | undefined;
		let abortReject!: (error: Error) => void;
		const abortPromise = new Promise<never>((_resolve, reject) => { abortReject = reject; });
		const abort = (kind: AbortKind) => {
			if (abortKind !== undefined) return;
			abortKind = kind;
			controller.abort(new Error(`${operation} ${kind}`));
			abortReject(new Error(kind === "timeout"
				? `${operation} timed out`
				: kind === "close" ? `${operation} was cancelled because ${this.#name} closed` : `${operation} was cancelled by the caller`));
		};
		const onCallerAbort = () => abort("caller");
		const onClose = () => abort("close");
		signal.addEventListener("abort", onCallerAbort, { once: true });
		this.#closing.signal.addEventListener("abort", onClose, { once: true });
		const timer = setTimeout(() => abort("timeout"), timeoutMs);
		const context: ExternalCallContext = {
			signal: controller.signal,
			deadlineAt: new Date(Date.now() + timeoutMs).toISOString(),
			acknowledgeAbort() {},
		};
		let operationPromise: Promise<T>;
		try {
			operationPromise = Promise.resolve(call(context));
		} catch (error) {
			operationPromise = Promise.reject(error);
		}
		const settlement = operationPromise.then(() => undefined, () => undefined).finally(() => {
			this.#active.delete(key);
		});
		this.#active.set(key, { controller, settlement });
		return (async () => {
			try {
				return await Promise.race([operationPromise, abortPromise]);
			} catch (error) {
				// A timeout, caller stop, or close revokes authority immediately, but the public
				// operation does not settle until the already-accepted adapter call has joined.
				// This prevents a late external effect from escaping after stop() resolves.
				if (abortKind !== undefined) await settlement;
				throw error;
			} finally {
				clearTimeout(timer);
				signal.removeEventListener("abort", onCallerAbort);
				this.#closing.signal.removeEventListener("abort", onClose);
			}
		})();
	}

	close(): Promise<void> {
		if (!this.#closePromise) {
			this.#closed = true;
			this.#closing.abort(new Error(`${this.#name} closed`));
			this.#closePromise = (async () => {
				const settlements = [...this.#active.values()].map((entry) => entry.settlement);
				if (settlements.length === 0) return;
				let timer: ReturnType<typeof setTimeout> | undefined;
				try {
					await Promise.race([
						Promise.all(settlements).then(() => undefined),
						new Promise<never>((_resolve, reject) => {
							timer = setTimeout(() => reject(new Error(`${this.#name} close timed out while joining active calls`)), this.#closeTimeoutMs);
						}),
					]);
				} finally {
					if (timer !== undefined) clearTimeout(timer);
				}
			})();
		}
		return this.#closePromise;
	}
}

export type ProductionParentFinalizationTransport = Pick<
	GitHubOrchestrationTransport,
	"findPullRequests" | "findChildIntegration" | "proveAncestry"
>;

export interface ProductionParentCheckPolicyAuthority {
	findRequiredCheckPolicies(
		query: {
			repository: string;
			parentIssue: number;
			generation: number;
			parentBranch: string;
			parentBaseBranch: string;
		},
		context: ExternalCallContext,
	): Promise<AuthoritativeLookup<RequiredGitHubCheckPolicy>>;
}

export interface ProductionParentReviewAuthority {
	findChangedPathEvidence(
		query: Omit<IndependentReviewTarget, "changedPaths" | "allowedScopes">,
		context: ExternalCallContext,
	): Promise<AuthoritativeLookup<GitHubChangedPathEvidence>>;
	review(target: IndependentReviewTarget, context: ExternalCallContext): Promise<ProductionReviewArtifact>;
}

export interface ProductionParentReadyTransitionRequest {
	repository: string;
	parentIssue: number;
	pullRequest: number;
	generation: number;
	branch: string;
	headSha: string;
	expectedRevision: number;
}

export interface ProductionParentReadyTransitionReceipt extends ProductionParentReadyTransitionRequest {
	schemaVersion: 1;
	authority: "transport";
	operation: "existing_draft_to_ready";
	appliedRevision: number;
	observedAt: string;
}

export interface ProductionParentReadyTransitionPort {
	markExistingDraftReady(
		request: ProductionParentReadyTransitionRequest,
		context: ExternalCallContext,
	): Promise<ProductionParentReadyTransitionReceipt>;
}

export interface ProductionParentFinalizerOptions {
	transport: ProductionParentFinalizationTransport;
	policies: ProductionParentCheckPolicyAuthority;
	reviews: ProductionParentReviewAuthority;
	readiness?: ProductionParentReadyTransitionPort;
	timeoutMs?: number;
	closeTimeoutMs?: number;
}

function canonicalPolicySet(
	lookup: AuthoritativeLookup<RequiredGitHubCheckPolicy>,
	plan: ProductionParentPlanDocument,
): RequiredGitHubCheckPolicy[] {
	if (lookup.complete !== true) throw new Error("authoritative required-check policy lookup is incomplete");
	if (!Array.isArray(lookup.items)) throw new Error("authoritative required-check policy lookup is malformed");
	const values = lookup.items.map(validateRequiredGitHubCheckPolicy)
		.sort((left, right) => left.baseBranch.localeCompare(right.baseBranch));
	const requiredBranches = [plan.parentBaseBranch, plan.parentBranch].sort();
	if (values.length !== requiredBranches.length
		|| !sameStrings(values.map((policy) => policy.baseBranch), requiredBranches)
		|| values.some((policy) => policy.repository !== plan.repository)) {
		throw new Error("authoritative required-check policy set does not exactly bind both parent branches");
	}
	return values;
}

function policySetIdentity(values: readonly RequiredGitHubCheckPolicy[]): string {
	return JSON.stringify(values);
}

function aggregateScopes(plan: ProductionParentPlanDocument): string[] {
	return [...new Set(plan.children.flatMap((child) => child.writeScopes))].sort();
}

function assertExactParentCoordinates(
	pullRequest: GitHubPullRequestEvidence,
	plan: ProductionParentPlanDocument,
	state: ProductionAutonomousState,
	marker: string,
	expected?: Pick<GitHubPullRequestEvidence, "number" | "baseSha" | "headSha">,
): void {
	if (pullRequest.repository !== plan.repository || pullRequest.workItemId !== `parent-${plan.parentIssue}`
		|| pullRequest.generation !== state.resourceGeneration || pullRequest.marker !== marker
		|| pullRequest.baseBranch !== plan.parentBaseBranch || pullRequest.headBranch !== plan.parentBranch
		|| pullRequest.state !== "open" || !sameStrings(pullRequest.allowedScopes, aggregateScopes(plan))) {
		throw new Error("authoritative parent pull request does not match the exact repository, generation, branch, or scope binding");
	}
	if (expected !== undefined && (pullRequest.number !== expected.number || pullRequest.baseSha !== expected.baseSha
		|| pullRequest.headSha !== expected.headSha)) {
		throw new Error("authoritative parent pull request head moved during finalization");
	}
}

function exactTransitionReceipt(
	value: ProductionParentReadyTransitionReceipt,
	request: ProductionParentReadyTransitionRequest,
): ProductionParentReadyTransitionReceipt {
	const candidate = readBoundedExactRecord(value, [
		"schemaVersion", "authority", "operation", "repository", "parentIssue", "pullRequest", "generation",
		"branch", "headSha", "expectedRevision", "appliedRevision", "observedAt",
	], [], "parent ready transition receipt");
	if (candidate.schemaVersion !== 1 || candidate.authority !== "transport"
		|| candidate.operation !== "existing_draft_to_ready") {
		throw new Error("invalid bounded existing-draft-to-ready transition receipt");
	}
	for (const key of ["parentIssue", "pullRequest", "generation", "expectedRevision", "appliedRevision"] as const) {
		if (!Number.isSafeInteger(candidate[key]) || (candidate[key] as number) < 1) {
			throw new Error("invalid bounded existing-draft-to-ready transition receipt");
		}
	}
	const receipt: ProductionParentReadyTransitionReceipt = {
		schemaVersion: 1,
		authority: "transport",
		operation: "existing_draft_to_ready",
		repository: String(candidate.repository),
		parentIssue: candidate.parentIssue as number,
		pullRequest: candidate.pullRequest as number,
		generation: candidate.generation as number,
		branch: String(candidate.branch),
		headSha: String(candidate.headSha),
		expectedRevision: candidate.expectedRevision as number,
		appliedRevision: candidate.appliedRevision as number,
		observedAt: exactTimestamp(candidate.observedAt, "parent ready transition observation"),
	};
	if (receipt.repository !== request.repository || receipt.parentIssue !== request.parentIssue
		|| receipt.pullRequest !== request.pullRequest || receipt.generation !== request.generation
		|| receipt.branch !== request.branch || receipt.headSha !== request.headSha
		|| receipt.expectedRevision !== request.expectedRevision || receipt.appliedRevision <= receipt.expectedRevision
		|| !SHA.test(receipt.headSha)) {
		throw new Error("existing-draft-to-ready transition receipt moved from the exact parent authority");
	}
	return receipt;
}

export class ProductionParentFinalizer implements ProductionParentFinalizerPort {
	readonly #transport: ProductionParentFinalizationTransport;
	readonly #policies: ProductionParentCheckPolicyAuthority;
	readonly #reviews: ProductionParentReviewAuthority;
	readonly #readiness?: ProductionParentReadyTransitionPort;
	readonly #timeoutMs: number;
	readonly #calls: BoundedParentCalls;

	constructor(options: ProductionParentFinalizerOptions) {
		if (typeof options?.transport?.findPullRequests !== "function"
			|| typeof options.transport.findChildIntegration !== "function"
			|| typeof options.transport.proveAncestry !== "function"
			|| typeof options?.policies?.findRequiredCheckPolicies !== "function"
			|| typeof options?.reviews?.findChangedPathEvidence !== "function"
			|| typeof options.reviews.review !== "function"
			|| (options.readiness !== undefined && typeof options.readiness.markExistingDraftReady !== "function")) {
			throw new Error("production parent finalizer requires bounded authoritative ports");
		}
		this.#transport = options.transport;
		this.#policies = options.policies;
		this.#reviews = options.reviews;
		this.#readiness = options.readiness;
		this.#timeoutMs = duration(options.timeoutMs, 30_000, "parent finalization timeout");
		this.#calls = new BoundedParentCalls(
			"production parent finalizer",
			duration(options.closeTimeoutMs, 5_000, "parent finalizer close timeout"),
		);
	}

	finalize(
		planValue: ProductionParentPlanDocument,
		stateValue: ProductionAutonomousState,
		signal: AbortSignal,
	): Promise<ProductionParentFinalization> {
		return this.#calls.invoke("parent finalization", signal, this.#timeoutMs, (context) =>
			this.#finalize(planValue, stateValue, context));
	}

	async #policiesFor(
		plan: ProductionParentPlanDocument,
		state: ProductionAutonomousState,
		context: ExternalCallContext,
	): Promise<RequiredGitHubCheckPolicy[]> {
		return canonicalPolicySet(await this.#policies.findRequiredCheckPolicies({
			repository: plan.repository,
			parentIssue: plan.parentIssue,
			generation: state.resourceGeneration,
			parentBranch: plan.parentBranch,
			parentBaseBranch: plan.parentBaseBranch,
		}, context), plan);
	}

	async #singleParent(
		repository: string,
		marker: string,
		context: ExternalCallContext,
	): Promise<GitHubPullRequestEvidence> {
		const lookup = await this.#transport.findPullRequests({ repository, marker }, context);
		if (lookup.complete !== true || !Array.isArray(lookup.items)) {
			throw new Error("authoritative parent pull request lookup is incomplete");
		}
		if (lookup.items.length !== 1) throw new Error("authoritative parent pull request is absent or ambiguous");
		return validateGitHubPullRequestEvidence(lookup.items[0]);
	}

	async #assertAncestry(
		repository: string,
		ancestorSha: string,
		descendantSha: string,
		context: ExternalCallContext,
	): Promise<void> {
		const proof = await this.#transport.proveAncestry({ repository, ancestorSha, descendantSha }, context);
		if (proof.schemaVersion !== 1 || proof.authority !== "transport" || proof.repository !== repository
			|| proof.ancestorSha !== ancestorSha || proof.descendantSha !== descendantSha || proof.result !== true
			|| !Number.isSafeInteger(proof.revision) || proof.revision < 1) {
			throw new Error("authoritative child integration ancestry proof is absent or mismatched");
		}
		exactTimestamp(proof.observedAt, "integration ancestry observation");
	}

	async #exactReceipts(
		plan: ProductionParentPlanDocument,
		state: ProductionAutonomousState,
		orchestration: ReturnType<typeof createProductionOrchestrationPlan>,
		parent: GitHubPullRequestEvidence,
		policies: readonly RequiredGitHubCheckPolicy[],
		context: ExternalCallContext,
	): Promise<ChildIntegrationReceipt[]> {
		const runtimeById = new Map(state.children.map((child) => [child.id, child]));
		const policy = policies.find((candidate) => candidate.baseBranch === plan.parentBranch)!;
		const receipts: ChildIntegrationReceipt[] = [];
		for (const child of orchestration.children) {
			const runtime = runtimeById.get(child.id)!;
			const lookup = await this.#transport.findChildIntegration({
				repository: plan.repository,
				childId: child.id,
				marker: child.markers.pullRequest,
			}, context);
			if (lookup.complete !== true || !Array.isArray(lookup.items)) {
				throw new Error(`authoritative integration receipt lookup is incomplete for ${child.id}`);
			}
			if (lookup.items.length !== 1) throw new Error(`authoritative integration receipt is absent or ambiguous for ${child.id}`);
			const receipt = validateChildIntegrationReceipt(lookup.items[0]);
			if (receipt.childId !== child.id || receipt.generation !== state.resourceGeneration
				|| receipt.marker !== child.markers.pullRequest || receipt.parentBranch !== plan.parentBranch
				|| receipt.pullRequest !== runtime.checkpoint!.pullRequest
				|| receipt.controllerProvenance.planDigest !== orchestration.canonical.digest
				|| receipt.controllerProvenance.policyDigest !== policy.digest
				|| receipt.pullRequestSnapshot.repository !== plan.repository
				|| receipt.pullRequestSnapshot.workItemId !== child.id
				|| receipt.pullRequestSnapshot.number !== receipt.pullRequest
				|| receipt.pullRequestSnapshot.generation !== state.resourceGeneration
				|| receipt.pullRequestSnapshot.marker !== child.markers.pullRequest
				|| receipt.pullRequestSnapshot.baseBranch !== plan.parentBranch
				|| receipt.pullRequestSnapshot.baseSha !== receipt.baseSha
				|| receipt.pullRequestSnapshot.headSha !== receipt.headSha
				|| receipt.pullRequestSnapshot.policyDigest !== policy.digest
				|| !sameStrings(receipt.pullRequestSnapshot.allowedScopes, child.writeScopes)
				|| receipt.observation.revision !== receipt.pullRequestSnapshot.revision
				|| receipt.observation.observedAt !== receipt.pullRequestSnapshot.observedAt) {
				throw new Error(`authoritative integration receipt does not exactly bind child ${child.id}`);
			}
			if (integrationReceiptDigest(receipt) !== runtime.checkpoint!.integrationReceiptDigest) {
				throw new Error(`integration receipt digest does not match the durable exact receipt for ${child.id}`);
			}
			await this.#assertAncestry(plan.repository, receipt.headSha, parent.headSha, context);
			await this.#assertAncestry(plan.repository, runtime.checkpoint!.parentHead!, parent.headSha, context);
			receipts.push(receipt);
		}
		if (new Set(receipts.map((receipt) => receipt.pullRequest)).size !== receipts.length) {
			throw new Error("authoritative integration receipts reuse a child pull request");
		}
		return receipts;
	}

	async #finalize(
		planValue: ProductionParentPlanDocument,
		stateValue: ProductionAutonomousState,
		context: ExternalCallContext,
	): Promise<ProductionParentFinalization> {
		const plan = validateProductionParentPlan(planValue, planValue.parentIssue);
		const state = validateProductionAutonomousState(stateValue);
		assertProductionPlanBinding(state, plan);
		for (const child of state.children) {
			if (child.status !== "succeeded" || child.stage !== "succeeded"
				|| child.checkpoint === undefined || child.checkpoint.pullRequest === undefined
				|| child.checkpoint.integrationReceiptDigest === undefined
				|| child.checkpoint.parentHead === undefined || !SHA.test(child.checkpoint.parentHead)
				|| !DIGEST.test(child.checkpoint.integrationReceiptDigest)) {
				throw new Error("parent finalization requires every child succeeded with an exact integration receipt and parent head");
			}
		}
		const initialPolicies = await this.#policiesFor(plan, state, context);
		const orchestration = createProductionOrchestrationPlan(plan, state.resourceGeneration, initialPolicies);
		const first = await this.#singleParent(plan.repository, orchestration.markers.parentPullRequest, context);
		assertExactParentCoordinates(first, plan, state, orchestration.markers.parentPullRequest);
		await this.#exactReceipts(plan, state, orchestration, first, initialPolicies, context);

		const targetWithoutPaths: Omit<IndependentReviewTarget, "changedPaths" | "allowedScopes"> = {
			repository: plan.repository,
			workItemId: `parent-${plan.parentIssue}`,
			pullRequest: first.number,
			generation: state.resourceGeneration,
			baseBranch: plan.parentBaseBranch,
			headBranch: plan.parentBranch,
			baseSha: first.baseSha,
			headSha: first.headSha,
		};
		const pathLookup = await this.#reviews.findChangedPathEvidence(targetWithoutPaths, context);
		if (pathLookup.complete !== true || !Array.isArray(pathLookup.items) || pathLookup.items.length !== 1) {
			throw new Error("authoritative parent changed-path evidence is incomplete or ambiguous");
		}
		const paths = validateGitHubChangedPathEvidence(pathLookup.items[0]);
		if (paths.repository !== plan.repository || paths.workItemId !== targetWithoutPaths.workItemId
			|| paths.pullRequest !== first.number || paths.generation !== state.resourceGeneration
			|| paths.baseSha !== first.baseSha || paths.headSha !== first.headSha
			|| !sameStrings(paths.paths, first.changedPaths)) {
			throw new Error("authoritative parent changed-path evidence moved from the exact parent head");
		}
		const reviewTarget: IndependentReviewTarget = {
			...targetWithoutPaths,
			changedPaths: paths.paths,
			allowedScopes: aggregateScopes(plan),
		};
		const artifact = await this.#reviews.review(reviewTarget, context);
		const review = validateIndependentReviewRecord(artifact.review);
		const attestation = validateAgentSessionAttestation(artifact.attestation, review);
		const reviewDecision = reconcileIndependentReview({
			target: reviewTarget,
			reviews: [review],
			attestations: [attestation],
		});
		if (reviewDecision.kind !== "satisfied" || reviewDecision.review.verdict !== "clean"
			|| reviewDecision.review.findings.length !== 0 || artifact.dispositions.length !== 0) {
			throw new Error("parent finalization requires a clean independent exact-head review");
		}
		if (!Number.isSafeInteger(artifact.revision) || artifact.revision < 1
			|| exactTimestamp(artifact.publishedAt, "parent review publication") < review.completedAt) {
			throw new Error("parent review artifact lacks durable publication evidence");
		}

		const second = await this.#singleParent(plan.repository, orchestration.markers.parentPullRequest, context);
		assertExactParentCoordinates(second, plan, state, orchestration.markers.parentPullRequest, first);
		const currentPolicies = await this.#policiesFor(plan, state, context);
		if (policySetIdentity(initialPolicies) !== policySetIdentity(currentPolicies)) {
			throw new Error("authoritative required-check policy moved during parent finalization");
		}
		const parentPolicy = currentPolicies.find((policy) => policy.baseBranch === plan.parentBaseBranch)!;
		const expected = {
			repository: plan.repository,
			workItemId: `parent-${plan.parentIssue}`,
			generation: state.resourceGeneration,
			number: first.number,
			marker: orchestration.markers.parentPullRequest,
			baseBranch: plan.parentBaseBranch,
			headBranch: plan.parentBranch,
			baseSha: first.baseSha,
			headSha: first.headSha,
			changedPathEvidence: paths,
			minimumObservation: { revision: first.revision, observedAt: first.observedAt },
			requiredCheckPolicy: parentPolicy,
			reviewTarget,
			attestations: [attestation] as AgentSessionAttestation[],
		};
		const assessed = evaluateGitHubPullRequestEvidence(second, expected, { allowDraft: second.draft });
		if (assessed.kind === "blocked") {
			throw new Error(`authoritative parent evidence blocked finalization: ${assessed.blockers.join(",")}`);
		}

		let transition: ProductionParentReadyTransitionReceipt | undefined;
		if (second.draft) {
			if (this.#readiness === undefined) {
				throw new Error("parent pull request is draft and no bounded existing-draft-to-ready transition port is configured");
			}
			const request: ProductionParentReadyTransitionRequest = {
				repository: plan.repository,
				parentIssue: plan.parentIssue,
				pullRequest: second.number,
				generation: state.resourceGeneration,
				branch: plan.parentBranch,
				headSha: second.headSha,
				expectedRevision: second.revision,
			};
			transition = exactTransitionReceipt(await this.#readiness.markExistingDraftReady(request, context), request);
		}

		const finalPolicies = await this.#policiesFor(plan, state, context);
		if (policySetIdentity(initialPolicies) !== policySetIdentity(finalPolicies)) {
			throw new Error("authoritative required-check policy moved before parent finalization settled");
		}
		const final = await this.#singleParent(plan.repository, orchestration.markers.parentPullRequest, context);
		assertExactParentCoordinates(final, plan, state, orchestration.markers.parentPullRequest, first);
		if (transition !== undefined && (final.revision !== transition.appliedRevision
			|| final.observedAt < transition.observedAt)) {
			throw new Error("authoritative parent pull request does not match the exact draft-to-ready transition receipt");
		}
		const finalAssessment = evaluateGitHubPullRequestEvidence(final, {
			...expected,
			requiredCheckPolicy: finalPolicies.find((policy) => policy.baseBranch === plan.parentBaseBranch)!,
		});
		if (finalAssessment.kind === "blocked") {
			throw new Error(`authoritative parent evidence blocked finalization: ${finalAssessment.blockers.join(",")}`);
		}
		const count = state.children.length;
		return {
			pullRequest: final.number,
			head: final.headSha,
			summary: `Parent PR #${final.number} at ${final.headSha} finalized with ${count} exact child ${count === 1 ? "receipt" : "receipts"}, trusted green CI, and a clean independent xhigh read-only review.`,
		};
	}

	close(): Promise<void> { return this.#calls.close(); }
}

export interface ProductionParentGateAdapterOptions {
	requestTimeoutMs?: number;
	pollTimeoutMs?: number;
	closeTimeoutMs?: number;
}

function parentMergeRequestId(
	plan: ProductionParentPlanDocument,
	state: ProductionAutonomousState,
	pullRequest: number,
	head: string,
): string {
	const digest = createHash("sha256").update(JSON.stringify({
		repository: plan.repository,
		parentIssue: plan.parentIssue,
		pullRequest,
		generation: state.generation,
		head,
		actorAllowlist: plan.actorAllowlist,
		expiresAt: plan.decisionExpiresAt,
	})).digest("hex").slice(0, 24);
	return `parent-merge-${plan.parentIssue}-${state.generation}-${digest}`;
}

function parentMergeQuestion(pullRequest: number, head: string): string {
	return `Approve the human-owned merge of parent PR #${pullRequest} at exact head ${head}? Shepherd cannot merge the default branch.`;
}

function mergeRequest(
	plan: ProductionParentPlanDocument,
	state: ProductionAutonomousState,
	pullRequest: number,
	head: string,
	requestId = parentMergeRequestId(plan, state, pullRequest, head),
): ProductionParentMergeRequest {
	if (!Number.isSafeInteger(pullRequest) || pullRequest < 1 || !SHA.test(head)) {
		throw new Error("parent merge gate requires an exact pull request and head");
	}
	return {
		requestId,
		repository: plan.repository,
		parentIssue: plan.parentIssue,
		pullRequest,
		generation: state.generation,
		headSha: head,
		actorAllowlist: [...plan.actorAllowlist],
		expiresAt: plan.decisionExpiresAt,
		question: parentMergeQuestion(pullRequest, head),
	};
}

function assertExactMergeState(
	value: AuthoritativeParentMergeState,
	request: ProductionParentMergeRequest,
	allowMerged: boolean,
): AuthoritativeParentMergeState {
	if (value.repository !== request.repository || value.pullRequest !== request.pullRequest) {
		throw new Error("authoritative parent merge coordinates moved from the exact durable gate");
	}
	if (!SHA.test(value.headSha)) throw new Error("authoritative parent merge observation has an invalid head");
	if (value.headSha !== request.headSha) {
		throw new Error("authoritative parent merge head moved from the exact durable gate");
	}
	if (!Number.isSafeInteger(value.revision) || value.revision < 1) {
		throw new Error("authoritative parent merge observation has an invalid revision");
	}
	exactTimestamp(value.observedAt, "parent merge observation");
	if (value.state === "open") {
		if (value.mergedAt !== null || value.mergeCommitSha !== null) {
			throw new Error("authoritative open parent pull request has ambiguous merge evidence");
		}
		return value;
	}
	if (value.state === "merged" && allowMerged) {
		if (value.mergedAt === null || value.mergeCommitSha === null || !SHA.test(value.mergeCommitSha)) {
			throw new Error("authoritative merged parent pull request lacks exact merge evidence");
		}
		exactTimestamp(value.mergedAt, "parent merge time");
		return value;
	}
	throw new Error("authoritative parent pull request is closed without the exact approved merge");
}

function parentHeadInvalidation(
	value: AuthoritativeParentMergeState,
	request: ProductionParentMergeRequest,
): Extract<ProductionParentGateObservation, { status: "invalidated" }> | undefined {
	if (value.repository !== request.repository || value.pullRequest !== request.pullRequest) {
		throw new Error("authoritative parent merge coordinates moved from the exact durable gate");
	}
	if (!SHA.test(value.headSha) || !Number.isSafeInteger(value.revision) || value.revision < 1) {
		throw new Error("authoritative parent merge observation is malformed");
	}
	const observedAt = exactTimestamp(value.observedAt, "parent merge observation");
	if (value.headSha === request.headSha) return undefined;
	return {
		status: "invalidated",
		repository: value.repository,
		pullRequest: value.pullRequest,
		previousHead: request.headSha,
		currentHead: value.headSha,
		revision: value.revision,
		observedAt,
	};
}

export class ProductionParentGateAdapter implements ProductionParentGatePort {
	readonly #gate: ProductionHumanParentMergeGate;
	readonly #lookup: ParentPullRequestMergeLookup;
	readonly #requestTimeoutMs: number;
	readonly #pollTimeoutMs: number;
	readonly #calls: BoundedParentCalls;

	constructor(
		broker: ParentDecisionBroker,
		lookup: ParentPullRequestMergeLookup,
		options: ProductionParentGateAdapterOptions = {},
	) {
		if (typeof broker?.request !== "function" || typeof broker.poll !== "function" || typeof broker.consume !== "function"
			|| typeof lookup?.observeExactPullRequest !== "function") {
			throw new Error("production parent gate requires a durable decision broker and authoritative merge lookup");
		}
		this.#gate = new ProductionHumanParentMergeGate(broker, lookup);
		this.#lookup = lookup;
		this.#requestTimeoutMs = duration(options.requestTimeoutMs, 15_000, "parent merge request timeout");
		this.#pollTimeoutMs = duration(options.pollTimeoutMs, 15_000, "parent merge poll timeout");
		this.#calls = new BoundedParentCalls(
			"production parent gate",
			duration(options.closeTimeoutMs, 5_000, "parent merge close timeout"),
		);
	}

	prepare(
		planValue: ProductionParentPlanDocument,
		stateValue: ProductionAutonomousState,
		finalization: ProductionParentFinalization,
	): { requestId: string } {
		const plan = validateProductionParentPlan(planValue, planValue.parentIssue);
		const state = validateProductionAutonomousState(stateValue);
		assertProductionPlanBinding(state, plan);
		if (state.humanGate !== undefined) throw new Error("parent merge gate is already durably prepared or requested");
		if (typeof finalization.summary !== "string" || finalization.summary.length === 0
			|| Buffer.byteLength(finalization.summary) > 4_096 || /[\u0000-\u001f\u007f-\u009f]/u.test(finalization.summary)) {
			throw new Error("parent finalization summary must be bounded safe text");
		}
		return { requestId: parentMergeRequestId(plan, state, finalization.pullRequest, finalization.head) };
	}

	request(
		planValue: ProductionParentPlanDocument,
		stateValue: ProductionAutonomousState,
		finalization: ProductionParentFinalization,
		signal: AbortSignal,
	): Promise<{ requestId: string }> {
		return this.#calls.invoke("parent merge request", signal, this.#requestTimeoutMs, async (context) => {
			const plan = validateProductionParentPlan(planValue, planValue.parentIssue);
			const state = validateProductionAutonomousState(stateValue);
			assertProductionPlanBinding(state, plan);
			const durable = state.humanGate;
			if (durable === undefined || durable.status !== "prepared") {
				throw new Error("parent merge gate requires a durable prepared request");
			}
			if (typeof finalization.summary !== "string" || finalization.summary.length === 0
				|| Buffer.byteLength(finalization.summary) > 4_096 || /[\u0000-\u001f\u007f-\u009f]/u.test(finalization.summary)) {
				throw new Error("parent finalization summary must be bounded safe text");
			}
			const request = mergeRequest(plan, state, finalization.pullRequest, finalization.head, durable.requestId);
			if (durable.repository !== request.repository || durable.pullRequest !== request.pullRequest
				|| durable.generation !== request.generation || durable.head !== request.headSha
				|| durable.requestId !== parentMergeRequestId(plan, state, request.pullRequest, request.headSha)) {
				throw new Error("durable prepared parent merge request changed its exact binding");
			}
			assertExactMergeState(await this.#lookup.observeExactPullRequest({
				repository: request.repository,
				pullRequest: request.pullRequest,
				headSha: request.headSha,
			}, context), request, false);
			const record = await this.#gate.request(request, context);
			if (record.requestId !== request.requestId) throw new Error("durable parent decision broker changed the exact request ID");
			return { requestId: request.requestId };
		});
	}

	observe(
		planValue: ProductionParentPlanDocument,
		stateValue: ProductionAutonomousState,
		signal: AbortSignal,
	): Promise<ProductionParentGateObservation> {
		return this.#calls.invoke("parent merge poll", signal, this.#pollTimeoutMs, async (context) => {
			const plan = validateProductionParentPlan(planValue, planValue.parentIssue);
			const state = validateProductionAutonomousState(stateValue);
			assertProductionPlanBinding(state, plan);
			const durable = state.humanGate;
			if (durable === undefined || durable.repository !== plan.repository || durable.generation !== state.generation) {
				throw new Error("durable parent merge gate is absent or changed generation");
			}
			const request = mergeRequest(plan, state, durable.pullRequest, durable.head, durable.requestId);
			if (durable.requestId !== parentMergeRequestId(plan, state, durable.pullRequest, durable.head)) {
				throw new Error("durable parent merge gate request ID does not match the exact binding");
			}
			const beforeValue = await this.#lookup.observeExactPullRequest({
				repository: request.repository,
				pullRequest: request.pullRequest,
				headSha: request.headSha,
			}, context);
			const invalidated = parentHeadInvalidation(beforeValue, request);
			if (invalidated !== undefined) return invalidated;
			const before = assertExactMergeState(beforeValue, request, true);
			const observation = await this.#gate.observe(request, context);
			if (observation.status === "invalidated") {
				return {
					status: "invalidated",
					repository: observation.repository,
					pullRequest: observation.pullRequest,
					previousHead: observation.previousHead,
					currentHead: observation.currentHead,
					revision: observation.revision,
					observedAt: observation.observedAt,
				};
			}
			if (observation.status !== "merged") return { status: observation.status };
			if (observation.repository !== request.repository || observation.pullRequest !== request.pullRequest
				|| observation.headSha !== request.headSha || observation.revision < before.revision
				|| observation.observedAt < before.observedAt) {
				throw new Error("authoritative merged parent evidence changed during gate observation");
			}
			return {
				status: "merged",
				repository: observation.repository,
				pullRequest: observation.pullRequest,
				head: observation.headSha,
				mergedAt: observation.mergedAt,
				mergeCommitSha: observation.mergeCommitSha,
				revision: observation.revision,
				observedAt: observation.observedAt,
			};
		});
	}

	close(): Promise<void> { return this.#calls.close(); }
}

export type { ChildIntegrationReceipt };
