import { createHash } from "node:crypto";

import type { ProductionEffectJournalPort } from "./autonomous-effect-journal.ts";
import { productionEffectKey } from "./autonomous-effect-journal.ts";
import type { ProductionAutonomousState } from "./autonomous-production-state.ts";
import type { ProductionParentPlanDocument } from "./autonomous-production-contract.ts";
import type {
	ProductionParentGateObservation,
	ProductionParentGatePort,
	ProductionParentMergeObservationEffectPort,
} from "./production-controller.ts";
import { readBoundedExactRecord } from "./review-router.ts";

const SHA = /^[0-9a-f]{40}$/u;

export interface ProductionParentMergeRecoveryDescriptor {
	operation: "parent_merge_observation";
	parentIssue: number;
	repository: string;
	planId: string;
	planDigest: string;
	parentBranch: string;
	parentBaseBranch: string;
	runId: string;
	resourceGeneration: number;
	generation: number;
	stateRevision: number;
	pullRequest: number;
	requestId: string;
	head: string;
}

interface PendingAcknowledgment {
	descriptor: ProductionParentMergeRecoveryDescriptor;
	observation: ProductionParentGateObservation;
}

function canonical(value: unknown): unknown {
	if (Array.isArray(value)) return value.map(canonical);
	if (value !== null && typeof value === "object") {
		return Object.fromEntries(Object.entries(value as Record<string, unknown>)
			.sort(([left], [right]) => left.localeCompare(right))
			.map(([key, item]) => [key, canonical(item)]));
	}
	return value;
}

function digest(value: unknown): string {
	return createHash("sha256").update(JSON.stringify(canonical(value))).digest("hex");
}

function positive(value: unknown, description: string): number {
	if (!Number.isSafeInteger(value) || (value as number) < 1) throw new Error(`${description} must be a positive integer`);
	return value as number;
}

function nonEmpty(value: unknown, description: string): string {
	if (typeof value !== "string" || value.length === 0 || value.length > 4_096) throw new Error(`${description} is invalid`);
	return value;
}

function sha(value: unknown, description: string): string {
	if (typeof value !== "string" || !SHA.test(value)) throw new Error(`${description} must be an exact lowercase commit SHA`);
	return value;
}

function timestamp(value: unknown, description: string): string {
	const text = nonEmpty(value, description);
	const parsed = new Date(text);
	if (!Number.isFinite(parsed.valueOf()) || parsed.toISOString() !== text) throw new Error(`${description} must be canonical`);
	return text;
}

function exactObservation(value: unknown): ProductionParentGateObservation {
	const status = value !== null && typeof value === "object" ? (value as { status?: unknown }).status : undefined;
	if (status === "pending" || status === "approved_waiting_for_merge" || status === "rejected") {
		const candidate = readBoundedExactRecord(value, ["status"], [], "parent merge observation");
		return { status: candidate.status as "pending" | "approved_waiting_for_merge" | "rejected" };
	}
	if (status === "invalidated") {
		const candidate = readBoundedExactRecord(
			value,
			["status", "repository", "pullRequest", "previousHead", "currentHead", "revision", "observedAt"],
			[],
			"parent merge invalidation",
		);
		return {
			status,
			repository: nonEmpty(candidate.repository, "parent repository"),
			pullRequest: positive(candidate.pullRequest, "parent pull request"),
			previousHead: sha(candidate.previousHead, "previous parent head"),
			currentHead: sha(candidate.currentHead, "current parent head"),
			revision: positive(candidate.revision, "parent decision revision"),
			observedAt: timestamp(candidate.observedAt, "parent invalidation observation time"),
		};
	}
	if (status === "merged") {
		const candidate = readBoundedExactRecord(
			value,
			["status", "repository", "pullRequest", "head", "mergedAt", "mergeCommitSha", "revision", "observedAt"],
			[],
			"parent merge observation",
		);
		return {
			status,
			repository: nonEmpty(candidate.repository, "parent repository"),
			pullRequest: positive(candidate.pullRequest, "parent pull request"),
			head: sha(candidate.head, "observed parent head"),
			mergedAt: timestamp(candidate.mergedAt, "parent merge time"),
			mergeCommitSha: sha(candidate.mergeCommitSha, "parent merge commit"),
			revision: positive(candidate.revision, "parent decision revision"),
			observedAt: timestamp(candidate.observedAt, "parent merge observation time"),
		};
	}
	throw new Error("parent merge observation has an invalid status");
}

function assertExternalBinding(
	descriptor: ProductionParentMergeRecoveryDescriptor,
	observation: ProductionParentGateObservation,
): void {
	if (observation.status === "invalidated") {
		if (observation.repository !== descriptor.repository
			|| observation.pullRequest !== descriptor.pullRequest
			|| observation.previousHead !== descriptor.head) {
			throw new Error("authoritative parent invalidation moved from the durable exact-head gate");
		}
	}
	if (observation.status === "merged") {
		if (observation.repository !== descriptor.repository
			|| observation.pullRequest !== descriptor.pullRequest
			|| observation.head !== descriptor.head) {
			throw new Error("authoritative parent merge observation moved from the durable exact-head gate");
		}
	}
}

function descriptorFor(
	plan: ProductionParentPlanDocument,
	state: ProductionAutonomousState,
): ProductionParentMergeRecoveryDescriptor {
	const gate = state.humanGate;
	if (!gate || gate.status !== "pending" || state.status !== "waiting_human" || state.stage !== "human_decision") {
		throw new Error("parent merge observation requires a durable pending exact-head gate");
	}
	if (plan.parentIssue !== state.parentIssue || plan.repository !== state.repository || plan.planId !== state.planId
		|| plan.parentBranch !== state.parentBranch || plan.parentBaseBranch !== state.parentBaseBranch
		|| gate.repository !== state.repository || gate.generation !== state.generation) {
		throw new Error("parent merge observation moved from its durable plan binding");
	}
	return {
		operation: "parent_merge_observation",
		parentIssue: state.parentIssue,
		repository: state.repository,
		planId: state.planId,
		planDigest: state.planDigest,
		parentBranch: state.parentBranch,
		parentBaseBranch: state.parentBaseBranch,
		runId: state.runId,
		resourceGeneration: state.resourceGeneration,
		generation: state.generation,
		stateRevision: state.revision,
		pullRequest: gate.pullRequest,
		requestId: gate.requestId,
		head: gate.head,
	};
}

function commonStateMatches(
	descriptor: ProductionParentMergeRecoveryDescriptor,
	state: ProductionAutonomousState,
): boolean {
	return state.parentIssue === descriptor.parentIssue
		&& state.repository === descriptor.repository
		&& state.planId === descriptor.planId
		&& state.planDigest === descriptor.planDigest
		&& state.parentBranch === descriptor.parentBranch
		&& state.parentBaseBranch === descriptor.parentBaseBranch
		&& state.runId === descriptor.runId
		&& state.resourceGeneration === descriptor.resourceGeneration
		&& state.generation === descriptor.generation
		&& state.revision === descriptor.stateRevision + 1;
}

function gateCoordinatesMatch(
	descriptor: ProductionParentMergeRecoveryDescriptor,
	gate: ProductionAutonomousState["humanGate"],
): boolean {
	return gate?.repository === descriptor.repository
		&& gate.pullRequest === descriptor.pullRequest
		&& gate.generation === descriptor.generation
		&& gate.requestId === descriptor.requestId
		&& gate.head === descriptor.head;
}

function stateAcknowledges(
	descriptor: ProductionParentMergeRecoveryDescriptor,
	observation: ProductionParentGateObservation,
	state: ProductionAutonomousState,
): boolean {
	if (!commonStateMatches(descriptor, state)) return false;
	if (observation.status === "pending" || observation.status === "approved_waiting_for_merge") {
		return state.status === "waiting_human" && state.stage === "human_decision"
			&& state.humanGate?.status === "pending" && gateCoordinatesMatch(descriptor, state.humanGate);
	}
	if (observation.status === "merged") {
		return state.status === "completed" && state.stage === "completed"
			&& state.humanGate?.status === "merged" && gateCoordinatesMatch(descriptor, state.humanGate)
			&& JSON.stringify(state.humanGate.mergeEvidence) === JSON.stringify({
				mergedAt: observation.mergedAt,
				mergeCommitSha: observation.mergeCommitSha,
				revision: observation.revision,
				observedAt: observation.observedAt,
			});
	}
	if (observation.status === "rejected") {
		return state.status === "failed" && state.stage === "blocked"
			&& state.terminalBlocker === "human rejected the exact parent merge"
			&& state.humanGate?.status === "rejected" && gateCoordinatesMatch(descriptor, state.humanGate);
	}
	if (observation.status !== "invalidated") return false;
	const archived = state.invalidatedParentGates?.find((gate) => gate.requestId === descriptor.requestId);
	return state.status === "running" && state.stage === "schedule" && state.humanGate === undefined
		&& archived?.status === "invalidated" && gateCoordinatesMatch(descriptor, archived)
		&& JSON.stringify(archived.invalidationEvidence) === JSON.stringify({
			currentHead: observation.currentHead,
			revision: observation.revision,
			observedAt: observation.observedAt,
		});
}

/**
 * Crash-safe adapter for the only externally effectful parent-gate poll. The controller must
 * persist the exact projected state before calling acknowledge; a fresh process delegates any
 * non-applied record to the production recovery barrier instead of replaying the consume call.
 */
export class ProductionParentMergeEffectJournal implements ProductionParentMergeObservationEffectPort {
	readonly #journal: ProductionEffectJournalPort;
	readonly #parentGate: ProductionParentGatePort;
	readonly #pending = new Map<string, PendingAcknowledgment>();

	constructor(options: { journal: ProductionEffectJournalPort; parentGate: ProductionParentGatePort }) {
		if (typeof options?.journal?.prepare !== "function" || typeof options.journal.observe !== "function"
			|| typeof options.journal.apply !== "function" || typeof options?.parentGate?.observe !== "function") {
			throw new Error("parent merge effect journal options are invalid");
		}
		this.#journal = options.journal;
		this.#parentGate = options.parentGate;
	}

	async observe(
		plan: ProductionParentPlanDocument,
		state: ProductionAutonomousState,
		signal: AbortSignal,
	): Promise<{ observation: ProductionParentGateObservation; effectKey: string }> {
		if (signal.aborted) throw signal.reason ?? new Error("parent merge observation cancelled");
		const descriptor = descriptorFor(plan, state);
		const intentDigest = digest(descriptor);
		const coordinates = {
			kind: "parent_merge_observation" as const,
			runId: descriptor.runId,
			generation: descriptor.generation,
			intentDigest,
		};
		const effectKey = productionEffectKey(coordinates);
		const prepared = await this.#journal.prepare({
			key: effectKey,
			...coordinates,
			recoveryDescriptor: descriptor,
		});
		if (prepared.phase !== "prepared") {
			throw new Error("non-applied parent merge observation must be reconciled by the recovery barrier");
		}
		const observation = exactObservation(await this.#parentGate.observe(plan, structuredClone(state), signal));
		assertExternalBinding(descriptor, observation);
		await this.#journal.observe(effectKey, {
			runId: descriptor.runId,
			generation: descriptor.generation,
		}, digest(observation));
		this.#pending.set(effectKey, { descriptor, observation });
		if (signal.aborted) throw signal.reason ?? new Error("parent merge observation cancelled");
		return { observation: structuredClone(observation), effectKey };
	}

	async acknowledge(effectKey: string, state: ProductionAutonomousState): Promise<void> {
		const pending = this.#pending.get(effectKey);
		if (!pending) throw new Error("parent merge effect acknowledgment lacks its exact observed result");
		if (!stateAcknowledges(pending.descriptor, pending.observation, state)) {
			throw new Error("parent merge effect can be applied only after its exact controller CAS");
		}
		await this.#journal.apply(effectKey, {
			runId: pending.descriptor.runId,
			generation: pending.descriptor.generation,
		});
		this.#pending.delete(effectKey);
	}
}
