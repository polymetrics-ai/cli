import {
	ProductionLifecycleError,
	type ProductionEffectRecord,
} from "./autonomous-production-contract.ts";
import type {
	ProductionEffectFence,
	ProductionEffectJournalPort,
} from "./autonomous-effect-journal.ts";
import { readBoundedExactRecord } from "./review-router.ts";

const DIGEST = /^[0-9a-f]{64}$/;
const MAX_RECOVERY_PASSES = 4_096;

export interface ProductionRecoveryFence extends ProductionEffectFence {
	signal?: AbortSignal;
}

export interface ProductionEffectRecoveryPort {
	/**
	 * Authoritatively reconcile a prepared intent from its durable recoveryDescriptor;
	 * never infer success from a timeout or replay a mutation merely because no receipt was found.
	 */
	observe(record: ProductionEffectRecord, signal: AbortSignal): Promise<{ resultDigest: string }>;
	/** Idempotently project observed truth into durable controller state under the record's fence. */
	apply(record: ProductionEffectRecord, signal: AbortSignal): Promise<void>;
}

export interface ProductionRecoveryResult {
	reconciled: number;
}

function recoverySignal(signal?: AbortSignal): AbortSignal {
	return signal ?? new AbortController().signal;
}

function throwIfCancelled(signal: AbortSignal): void {
	if (signal.aborted) {
		throw new ProductionLifecycleError("retryable", "production recovery was cancelled before scheduling", ["cancelled"]);
	}
}

function assertFence(record: ProductionEffectRecord, fence: ProductionRecoveryFence): void {
	if (record.runId !== fence.runId || record.generation !== fence.generation) {
		throw new ProductionLifecycleError(
			"stale_parent",
			"non-applied production effect belongs to a stale run generation",
			["stale_generation"],
		);
	}
}

function resultDigest(value: unknown): string {
	const candidate = readBoundedExactRecord(value, ["resultDigest"], [], "production recovery observation");
	if (typeof candidate.resultDigest !== "string" || !DIGEST.test(candidate.resultDigest)) {
		throw new Error("production recovery observation requires a SHA-256 result digest");
	}
	return candidate.resultDigest;
}

/**
 * A scheduling barrier over the journal. It first fences the complete pending set, then
 * authoritatively observes prepared effects and idempotently applies every observation.
 */
export class ProductionRecoveryBarrier {
	readonly #journal: ProductionEffectJournalPort;
	readonly #recovery: ProductionEffectRecoveryPort;
	#tail: Promise<void> = Promise.resolve();

	constructor(journal: ProductionEffectJournalPort, recovery: ProductionEffectRecoveryPort) {
		this.#journal = journal;
		this.#recovery = recovery;
	}

	open(fence: ProductionRecoveryFence): Promise<ProductionRecoveryResult> {
		const run = this.#tail.then(() => this.#open(fence));
		this.#tail = run.then(() => undefined, () => undefined);
		return run;
	}

	async #open(fence: ProductionRecoveryFence): Promise<ProductionRecoveryResult> {
		if (!Number.isSafeInteger(fence.generation) || fence.generation < 1
			|| typeof fence.runId !== "string" || fence.runId.length === 0) {
			throw new Error("invalid production recovery fence");
		}
		const signal = recoverySignal(fence.signal);
		throwIfCancelled(signal);
		let reconciled = 0;
		for (let pass = 0; pass < MAX_RECOVERY_PASSES; pass += 1) {
			const pending = await this.#journal.listNonApplied();
			throwIfCancelled(signal);
			if (pending.length === 0) return { reconciled };
			// Fence the whole snapshot before invoking any adapter. A foreign generation blocks all work.
			for (const record of pending) assertFence(record, fence);
			for (const snapshot of pending) {
				throwIfCancelled(signal);
				let record = await this.#journal.load(snapshot.key);
				if (!record || record.phase === "applied") continue;
				assertFence(record, fence);
				if (record.recoveryDescriptor === undefined) {
					throw new ProductionLifecycleError(
						"terminal",
						"non-applied production effect lacks durable recovery coordinates",
						["recovery_descriptor_missing"],
					);
				}
				if (record.phase === "prepared") {
					const observed = await this.#recovery.observe(structuredClone(record), signal);
					throwIfCancelled(signal);
					record = await this.#journal.observe(
						record.key,
						{ runId: fence.runId, generation: fence.generation },
						resultDigest(observed),
					);
				}
				throwIfCancelled(signal);
				await this.#recovery.apply(structuredClone(record), signal);
				// Once application returns, journal it even if cancellation raced with the return.
				await this.#journal.apply(record.key, { runId: fence.runId, generation: fence.generation });
				reconciled += 1;
				throwIfCancelled(signal);
			}
		}
		throw new ProductionLifecycleError(
			"terminal",
			"production recovery could not drain its bounded effect set",
			["recovery_not_quiescent"],
		);
	}
}
