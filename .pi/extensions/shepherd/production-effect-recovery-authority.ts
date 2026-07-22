import { createHash, randomUUID } from "node:crypto";
import { constants } from "node:fs";
import { link, lstat, mkdir, open, rm } from "node:fs/promises";
import { join, resolve } from "node:path";

import {
	validateProductionAutonomousState,
	type ProductionAutonomousState,
	type ProductionStateStore,
} from "./autonomous-production-state.ts";
import {
	ProductionLifecycleError,
	type ProductionEffectKind,
	type ProductionEffectRecord,
} from "./autonomous-production-contract.ts";
import type {
	ProductionEffectRecoveryPort,
	ProductionRecoveryObservation,
} from "./autonomous-recovery.ts";
import { readBoundedExactRecord } from "./review-router.ts";

const DIGEST = /^[0-9a-f]{64}$/u;
const MAX_PROJECTION_BYTES = 2 * 1024 * 1024;

export interface ProductionRecoveryProbeRequest {
	record: ProductionEffectRecord;
	descriptor: unknown;
	currentState: ProductionAutonomousState;
	signal: AbortSignal;
}

export type ProductionRecoveryProbeEvidence =
	| { status: "absent" }
	| {
		status: "applied";
		/** Digest of the canonical external result, equal to the normal effect result digest. */
		resultDigest: string;
		/** Exact state transition justified by the external evidence. */
		projectedState: ProductionAutonomousState;
	};

export type ProductionRecoveryProbe = (
	request: ProductionRecoveryProbeRequest,
) => Promise<ProductionRecoveryProbeEvidence>;

/**
 * Every external-effect kind has an explicit evidence route. Construction cannot silently omit a
 * new kind: adding a ProductionEffectKind makes this table fail type-checking until recovery is
 * designed for it.
 */
export type ProductionRecoveryProbeTable = {
	[K in ProductionEffectKind]: ProductionRecoveryProbe;
};

interface DurableRecoveryProjection {
	schemaVersion: 1;
	key: string;
	kind: ProductionEffectKind;
	runId: string;
	generation: number;
	resultDigest: string;
	stateDigest: string;
	projectedState: ProductionAutonomousState;
}

function digest(value: unknown): string {
	return createHash("sha256").update(JSON.stringify(value)).digest("hex");
}

function exactDescriptor(record: ProductionEffectRecord): unknown {
	if (record.recoveryDescriptor === undefined) {
		throw new ProductionLifecycleError(
			"terminal",
			"prepared production effect lacks durable recovery coordinates",
			["recovery_descriptor_missing"],
		);
	}
	return structuredClone(record.recoveryDescriptor);
}

function descriptorRecord(value: unknown, description: string): Record<string, unknown> {
	return readBoundedExactRecord(
		value,
		[],
		Object.keys(value !== null && typeof value === "object" && !Array.isArray(value)
			? value as Record<string, unknown> : {}),
		description,
	);
}

function checkpointContains(state: ProductionAutonomousState, childId: string, key: string): boolean {
	const checkpoint = state.children.find((child) => child.id === childId)?.checkpoint;
	return checkpoint?.effectKey === key || checkpoint?.effectKeys?.includes(key) === true;
}

function gateMatches(
	record: ProductionEffectRecord,
	state: ProductionAutonomousState,
	descriptor: Record<string, unknown>,
): boolean {
	const nested = descriptor.request !== null && typeof descriptor.request === "object" && !Array.isArray(descriptor.request)
		? descriptor.request as Record<string, unknown> : undefined;
	const requestId = descriptor.requestId ?? nested?.requestId;
	if (typeof requestId !== "string" || requestId.length === 0) return false;
	if (record.childId !== undefined) {
		return state.childGate?.childId === record.childId && state.childGate.requestId === requestId;
	}
	return state.humanGate?.requestId === requestId;
}

function mergeObservationMatches(state: ProductionAutonomousState, descriptor: Record<string, unknown>): boolean {
	const pullRequest = descriptor.pullRequest;
	const head = descriptor.head;
	const requestId = descriptor.requestId;
	const stateRevision = descriptor.stateRevision;
	if (!Number.isSafeInteger(pullRequest) || (pullRequest as number) < 1
		|| typeof head !== "string" || typeof requestId !== "string" || requestId.length === 0
		|| !Number.isSafeInteger(stateRevision) || (stateRevision as number) < 1
		|| state.revision !== (stateRevision as number) + 1) return false;
	const coordinatesMatch = (candidate: ProductionAutonomousState["humanGate"]): boolean => {
		if (candidate === undefined) return false;
		return candidate.repository === descriptor.repository
			&& candidate.pullRequest === pullRequest
			&& candidate.generation === descriptor.generation
			&& candidate.requestId === requestId
			&& candidate.head === head;
	};
	const gate = state.humanGate;
	if (gate?.status === "pending") {
		return state.status === "waiting_human" && state.stage === "human_decision" && coordinatesMatch(gate);
	}
	if (gate?.status === "merged") {
		return state.status === "completed" && state.stage === "completed" && coordinatesMatch(gate);
	}
	if (gate?.status === "rejected") {
		return state.status === "failed" && state.stage === "blocked" && coordinatesMatch(gate);
	}
	const archived = state.invalidatedParentGates?.find((candidate) => candidate.requestId === requestId);
	return state.status === "running" && state.stage === "schedule" && gate === undefined
		&& archived?.status === "invalidated" && coordinatesMatch(archived);
}

function assertProjectionBinding(
	record: ProductionEffectRecord,
	current: ProductionAutonomousState,
	projectedValue: ProductionAutonomousState,
): ProductionAutonomousState {
	const projected = validateProductionAutonomousState(projectedValue);
	if (current.parentIssue !== projected.parentIssue || current.repository !== projected.repository
		|| current.planId !== projected.planId || current.planDigest !== projected.planDigest
		|| current.parentBranch !== projected.parentBranch || current.parentBaseBranch !== projected.parentBaseBranch
		|| record.runId !== current.runId || record.generation !== current.generation
		|| projected.runId !== record.runId || projected.generation !== record.generation) {
		throw new ProductionLifecycleError("terminal", "recovery projection moved from its durable run binding", ["recovery_projection_moved"]);
	}
	const alreadyProjected = JSON.stringify(current) === JSON.stringify(projected);
	if (!alreadyProjected && projected.revision !== current.revision + 1) {
		throw new ProductionLifecycleError("terminal", "recovery projection is not one exact CAS transition", ["recovery_projection_revision"]);
	}
	const descriptor = descriptorRecord(exactDescriptor(record), "production recovery descriptor");
	let bound: boolean;
	switch (record.kind) {
		case "human_request":
		case "human_consume":
			bound = gateMatches(record, projected, descriptor);
			break;
		case "parent_merge_observation":
			bound = mergeObservationMatches(projected, descriptor);
			break;
		case "workspace_claim":
		case "agent_implementation":
		case "agent_correction":
		case "shell_verification":
		case "git_commit":
		case "git_push":
		case "child_pull_request":
		case "independent_review":
		case "child_integration":
		case "parent_refresh":
		case "child_head_reconciliation":
			bound = record.childId !== undefined && checkpointContains(projected, record.childId, record.key);
			break;
		default: {
			const exhaustive: never = record.kind;
			throw new Error(`unrouted production effect kind ${String(exhaustive)}`);
		}
	}
	if (!bound) {
		throw new ProductionLifecycleError(
			"terminal",
			`recovery projection does not acknowledge exact ${record.kind} effect`,
			["recovery_projection_missing"],
		);
	}
	return projected;
}

function validateProjection(value: unknown): DurableRecoveryProjection {
	const candidate = readBoundedExactRecord(value, [
		"schemaVersion", "key", "kind", "runId", "generation", "resultDigest", "stateDigest", "projectedState",
	], [], "durable production recovery projection");
	if (candidate.schemaVersion !== 1 || typeof candidate.key !== "string" || !DIGEST.test(candidate.key)
		|| typeof candidate.resultDigest !== "string" || !DIGEST.test(candidate.resultDigest)
		|| typeof candidate.stateDigest !== "string" || !DIGEST.test(candidate.stateDigest)
		|| typeof candidate.runId !== "string" || candidate.runId.length === 0
		|| !Number.isSafeInteger(candidate.generation) || (candidate.generation as number) < 1) {
		throw new Error("durable production recovery projection is invalid");
	}
	const projectedState = validateProductionAutonomousState(candidate.projectedState);
	if (candidate.stateDigest !== digest(projectedState)) throw new Error("durable recovery state digest conflicts");
	return {
		schemaVersion: 1,
		key: candidate.key,
		kind: candidate.kind as ProductionEffectKind,
		runId: candidate.runId,
		generation: candidate.generation as number,
		resultDigest: candidate.resultDigest,
		stateDigest: candidate.stateDigest,
		projectedState,
	};
}

function isErrno(error: unknown, code: string): boolean {
	return typeof error === "object" && error !== null && "code" in error
		&& (error as { code?: unknown }).code === code;
}

/**
 * Durable, exhaustive recovery authority. Probes only observe exact external truth; this class
 * owns projection persistence and the controller-state CAS. An authoritative absence is returned
 * to the barrier for transactional journal reset and is deliberately not cached across retries.
 */
export class ProductionEffectRecoveryAuthority implements ProductionEffectRecoveryPort {
	readonly #root: string;
	readonly #issue: number;
	readonly #store: ProductionStateStore;
	readonly #probes: ProductionRecoveryProbeTable;

	constructor(options: {
		stateRoot: string;
		issue: number;
		stateStore: ProductionStateStore;
		probes: ProductionRecoveryProbeTable;
	}) {
		if (typeof options?.stateRoot !== "string" || options.stateRoot.length === 0
			|| !Number.isSafeInteger(options.issue) || options.issue < 1
			|| typeof options.stateStore?.load !== "function" || typeof options.stateStore.compareAndSwap !== "function") {
			throw new Error("production effect recovery authority options are invalid");
		}
		for (const kind of Object.keys(options.probes ?? {}) as ProductionEffectKind[]) {
			if (typeof options.probes[kind] !== "function") throw new Error(`production recovery probe ${kind} is invalid`);
		}
		const kinds: ProductionEffectKind[] = [
			"workspace_claim", "agent_implementation", "agent_correction", "shell_verification", "git_commit", "git_push",
			"child_pull_request", "independent_review", "child_integration", "parent_refresh", "child_head_reconciliation",
			"human_request", "human_consume", "parent_merge_observation",
		];
		if (kinds.some((kind) => typeof options.probes?.[kind] !== "function")) {
			throw new Error("production recovery requires an exhaustive kind-specific evidence table");
		}
		this.#root = join(resolve(options.stateRoot), "effect-recovery-projections");
		this.#issue = options.issue;
		this.#store = options.stateStore;
		this.#probes = options.probes;
	}

	#path(key: string): string { return join(this.#root, `${key}.json`); }

	async #read(key: string): Promise<DurableRecoveryProjection | undefined> {
		let handle: Awaited<ReturnType<typeof open>>;
		try { handle = await open(this.#path(key), constants.O_RDONLY | (constants.O_NOFOLLOW ?? 0)); }
		catch (error) { if (isErrno(error, "ENOENT")) return undefined; throw error; }
		try {
			const metadata = await handle.stat();
			if (!metadata.isFile() || metadata.size > MAX_PROJECTION_BYTES) throw new Error("recovery projection is not a bounded regular file");
			let value: unknown;
			try { value = JSON.parse(await handle.readFile("utf8")); }
			catch { throw new Error("invalid durable production recovery projection JSON"); }
			return validateProjection(value);
		} finally { await handle.close(); }
	}

	async #write(documentValue: DurableRecoveryProjection): Promise<DurableRecoveryProjection> {
		const document = validateProjection(documentValue);
		await mkdir(this.#root, { recursive: true, mode: 0o700 });
		const metadata = await lstat(this.#root);
		if (!metadata.isDirectory() || metadata.isSymbolicLink()) throw new Error("recovery projection root is unsafe");
		const serialized = `${JSON.stringify(document)}\n`;
		if (Buffer.byteLength(serialized) > MAX_PROJECTION_BYTES) throw new Error("recovery projection exceeds its bound");
		const temporary = join(this.#root, `.${document.key}.${randomUUID()}.tmp`);
		const handle = await open(temporary, "wx", 0o600);
		try { await handle.writeFile(serialized, "utf8"); await handle.sync(); }
		finally { await handle.close(); }
		try {
			await link(temporary, this.#path(document.key));
		} catch (error) {
			if (!isErrno(error, "EEXIST")) throw error;
			const existing = await this.#read(document.key);
			if (existing === undefined || JSON.stringify(existing) !== JSON.stringify(document)) {
				throw new ProductionLifecycleError(
					"terminal",
					"competing recovery projections disagree",
					["recovery_projection_conflict"],
				);
			}
			return existing;
		} finally {
			await rm(temporary, { force: true });
		}
		return document;
	}

	#assertRecord(document: DurableRecoveryProjection, record: ProductionEffectRecord): void {
		if (document.key !== record.key || document.kind !== record.kind || document.runId !== record.runId
			|| document.generation !== record.generation
			|| (record.resultDigest !== undefined && document.resultDigest !== record.resultDigest)) {
			throw new ProductionLifecycleError("terminal", "durable recovery projection conflicts with its effect", ["recovery_projection_conflict"]);
		}
	}

	async #probe(record: ProductionEffectRecord, signal: AbortSignal): Promise<ProductionRecoveryProbeEvidence> {
		if (signal.aborted) throw signal.reason ?? new Error("production recovery cancelled");
		const current = await this.#store.load(this.#issue);
		if (current === undefined) throw new ProductionLifecycleError("terminal", "production recovery state is absent", ["recovery_state_missing"]);
		if (current.runId !== record.runId || current.generation !== record.generation) {
			throw new ProductionLifecycleError("stale_parent", "production recovery probe crossed a run generation", ["stale_generation"]);
		}
		const evidence = await this.#probes[record.kind]({
			record: structuredClone(record),
			descriptor: exactDescriptor(record),
			currentState: structuredClone(current),
			signal,
		});
		if (signal.aborted) throw signal.reason ?? new Error("production recovery cancelled");
		if (evidence.status === "absent") return { status: "absent" };
		if (evidence.status !== "applied" || !DIGEST.test(evidence.resultDigest)) {
			throw new Error(`production ${record.kind} recovery probe returned malformed evidence`);
		}
		const projectedState = assertProjectionBinding(record, current, evidence.projectedState);
		return { status: "applied", resultDigest: evidence.resultDigest, projectedState };
	}

	async #projection(record: ProductionEffectRecord, signal: AbortSignal): Promise<DurableRecoveryProjection | undefined> {
		const existing = await this.#read(record.key);
		if (existing !== undefined) { this.#assertRecord(existing, record); return existing; }
		const evidence = await this.#probe(record, signal);
		if (evidence.status === "absent") return undefined;
		const document: DurableRecoveryProjection = {
			schemaVersion: 1,
			key: record.key,
			kind: record.kind,
			runId: record.runId,
			generation: record.generation,
			resultDigest: evidence.resultDigest,
			stateDigest: digest(evidence.projectedState),
			projectedState: evidence.projectedState,
		};
		return this.#write(document);
	}

	async observe(record: ProductionEffectRecord, signal: AbortSignal): Promise<ProductionRecoveryObservation> {
		if (record.phase !== "prepared") throw new Error("recovery observation requires a prepared effect");
		const projection = await this.#projection(record, signal);
		return projection === undefined
			? { status: "absent" }
			: { status: "applied", resultDigest: projection.resultDigest };
	}

	async apply(record: ProductionEffectRecord, signal: AbortSignal): Promise<void> {
		if (record.phase !== "observed" && record.phase !== "applied") {
			throw new Error("recovery projection requires an observed effect");
		}
		const projection = await this.#projection(record, signal);
		if (projection === undefined) {
			throw new ProductionLifecycleError("terminal", "observed effect is authoritatively absent", ["observed_effect_absent"]);
		}
		this.#assertRecord(projection, record);
		const current = await this.#store.load(this.#issue);
		if (current === undefined) throw new Error("production recovery state is absent");
		if (JSON.stringify(current) === JSON.stringify(projection.projectedState)) return;
		const currentDescriptor = descriptorRecord(exactDescriptor(record), "production recovery descriptor");
		if (checkpointContains(current, record.childId ?? "", record.key)
			|| record.kind === "human_request" && gateMatches(record, current, currentDescriptor)
			|| record.kind === "human_consume" && gateMatches(record, current, currentDescriptor)
				&& (record.childId === undefined
					? current.humanGate?.status !== "prepared" && current.humanGate?.status !== "pending"
					: current.childGate?.status === "authorized" || current.childGate?.status === "aborted")
			|| record.kind === "parent_merge_observation"
				&& mergeObservationMatches(current, currentDescriptor)) {
			return;
		}
		if (current.revision + 1 !== projection.projectedState.revision) {
			throw new ProductionLifecycleError("terminal", "recovery CAS was superseded without exact effect evidence", ["recovery_projection_conflict"]);
		}
		await this.#store.compareAndSwap({
			issue: current.parentIssue,
			revision: current.revision,
			generation: current.generation,
			runId: current.runId,
		}, projection.projectedState);
	}
}
