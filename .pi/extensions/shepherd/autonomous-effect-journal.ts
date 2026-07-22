import { createHash, randomUUID } from "node:crypto";
import { constants } from "node:fs";
import { lstat, mkdir, open, readFile, rename, rm } from "node:fs/promises";
import { join } from "node:path";

import type {
	ProductionEffectKind,
	ProductionEffectRecord,
} from "./autonomous-production-contract.ts";
import { readBoundedExactRecord } from "./review-router.ts";

const DIGEST = /^[0-9a-f]{64}$/;
const SAFE_ID = /^[a-z0-9][a-z0-9_.:-]{0,255}$/;
const MAX_JOURNAL_BYTES = 4 * 1024 * 1024;
const MAX_EFFECTS = 4_096;
const LOCK_ATTEMPTS = 2_000;
const LOCK_RETRY_MS = 5;

export interface ProductionEffectIntentCoordinates {
	kind: ProductionEffectKind;
	runId: string;
	generation: number;
	childId?: string;
	intentDigest: string;
}

export interface ProductionEffectIntent extends ProductionEffectIntentCoordinates {
	key: string;
}

export interface ProductionEffectFence {
	runId: string;
	generation: number;
}

export interface ProductionEffectJournalPort {
	prepare(intent: ProductionEffectIntent, now?: Date): Promise<ProductionEffectRecord>;
	load(key: string): Promise<ProductionEffectRecord | undefined>;
	listNonApplied(): Promise<ProductionEffectRecord[]>;
	observe(key: string, fence: ProductionEffectFence, resultDigest: string, now?: Date): Promise<ProductionEffectRecord>;
	apply(key: string, fence: ProductionEffectFence, now?: Date): Promise<ProductionEffectRecord>;
}

interface ProductionJournalDocument {
	schemaVersion: 1;
	records: ProductionEffectRecord[];
}

function exact(value: unknown, required: readonly string[], optional: readonly string[] = [], description = "production effect") {
	return readBoundedExactRecord(value, required, optional, description);
}

function safeId(value: unknown, description: string): string {
	if (typeof value !== "string" || !SAFE_ID.test(value)) throw new Error(`invalid ${description}`);
	return value;
}

function positive(value: unknown, description: string): number {
	if (!Number.isSafeInteger(value) || (value as number) < 1) throw new Error(`${description} must be a positive integer`);
	return value as number;
}

function digest(value: unknown, description: string): string {
	if (typeof value !== "string" || !DIGEST.test(value)) throw new Error(`${description} must be a SHA-256 digest`);
	return value;
}

function timestamp(value: unknown, description: string): string {
	if (typeof value !== "string" || value.length > 64) throw new Error(`${description} must be a canonical timestamp`);
	const parsed = new Date(value);
	if (!Number.isFinite(parsed.valueOf()) || parsed.toISOString() !== value) throw new Error(`${description} must be a canonical timestamp`);
	return value;
}

function kind(value: unknown): ProductionEffectKind {
	const kinds: ProductionEffectKind[] = [
		"workspace_claim", "agent_implementation", "agent_correction", "shell_verification", "git_commit", "git_push",
		"child_pull_request", "independent_review", "child_integration", "parent_refresh", "human_request", "human_consume",
		"parent_merge_observation",
	];
	if (!kinds.includes(value as ProductionEffectKind)) throw new Error("invalid production effect kind");
	return value as ProductionEffectKind;
}

function canonicalIntent(value: ProductionEffectIntentCoordinates): ProductionEffectIntentCoordinates {
	const candidate = exact(
		value,
		["kind", "runId", "generation", "intentDigest"],
		["childId"],
		"production effect intent coordinates",
	);
	return {
		kind: kind(candidate.kind),
		runId: safeId(candidate.runId, "effect run ID"),
		generation: positive(candidate.generation, "effect generation"),
		...(candidate.childId === undefined ? {} : { childId: safeId(candidate.childId, "effect child ID") }),
		intentDigest: digest(candidate.intentDigest, "effect intent digest"),
	};
}

export function productionEffectKey(value: ProductionEffectIntentCoordinates): string {
	return createHash("sha256").update(JSON.stringify(canonicalIntent(value))).digest("hex");
}

function validateIntent(value: unknown): ProductionEffectIntent {
	const candidate = exact(
		value,
		["key", "kind", "runId", "generation", "intentDigest"],
		["childId"],
		"production effect intent",
	);
	const coordinates = canonicalIntent({
		kind: candidate.kind as ProductionEffectKind,
		runId: candidate.runId as string,
		generation: candidate.generation as number,
		...(candidate.childId === undefined ? {} : { childId: candidate.childId as string }),
		intentDigest: candidate.intentDigest as string,
	});
	const key = digest(candidate.key, "effect key");
	if (key !== productionEffectKey(coordinates)) throw new Error("effect key conflicts with its exact intent");
	return { key, ...coordinates };
}

export function validateProductionEffectRecord(value: unknown): ProductionEffectRecord {
	const candidate = exact(
		value,
		["schemaVersion", "key", "kind", "phase", "runId", "generation", "intentDigest", "preparedAt"],
		["childId", "observedAt", "appliedAt", "resultDigest"],
		"production effect record",
	);
	if (candidate.schemaVersion !== 1) throw new Error("unsupported production effect schema");
	if (candidate.phase !== "prepared" && candidate.phase !== "observed" && candidate.phase !== "applied") {
		throw new Error("invalid production effect phase");
	}
	const intent = validateIntent({
		key: candidate.key,
		kind: candidate.kind,
		runId: candidate.runId,
		generation: candidate.generation,
		...(candidate.childId === undefined ? {} : { childId: candidate.childId }),
		intentDigest: candidate.intentDigest,
	});
	const preparedAt = timestamp(candidate.preparedAt, "effect preparation time");
	const observedAt = candidate.observedAt === undefined ? undefined : timestamp(candidate.observedAt, "effect observation time");
	const appliedAt = candidate.appliedAt === undefined ? undefined : timestamp(candidate.appliedAt, "effect application time");
	const resultDigest = candidate.resultDigest === undefined ? undefined : digest(candidate.resultDigest, "effect result digest");
	if (candidate.phase === "prepared" && (observedAt || appliedAt || resultDigest)) {
		throw new Error("prepared effect cannot contain observation or application truth");
	}
	if (candidate.phase === "observed" && (!observedAt || appliedAt || !resultDigest)) {
		throw new Error("observed effect requires only observation truth");
	}
	if (candidate.phase === "applied" && (!observedAt || !appliedAt || !resultDigest)) {
		throw new Error("applied effect requires observation and application truth");
	}
	if ((observedAt && observedAt < preparedAt) || (appliedAt && observedAt && appliedAt < observedAt)) {
		throw new Error("production effect chronology is invalid");
	}
	return {
		schemaVersion: 1,
		...intent,
		phase: candidate.phase,
		preparedAt,
		...(observedAt === undefined ? {} : { observedAt }),
		...(appliedAt === undefined ? {} : { appliedAt }),
		...(resultDigest === undefined ? {} : { resultDigest }),
	};
}

function validateDocument(value: unknown): ProductionJournalDocument {
	const candidate = exact(value, ["schemaVersion", "records"], [], "production effect journal");
	if (candidate.schemaVersion !== 1 || !Array.isArray(candidate.records) || candidate.records.length > MAX_EFFECTS) {
		throw new Error("invalid production effect journal");
	}
	const records: ProductionEffectRecord[] = [];
	for (let index = 0; index < candidate.records.length; index += 1) {
		const descriptor = Object.getOwnPropertyDescriptor(candidate.records, index);
		if (!descriptor || !Object.hasOwn(descriptor, "value") || descriptor.enumerable !== true) {
			throw new Error("production effect journal must contain dense data records");
		}
		records.push(validateProductionEffectRecord(descriptor.value));
	}
	if (new Set(records.map((record) => record.key)).size !== records.length) throw new Error("duplicate production effect key");
	return { schemaVersion: 1, records: records.sort((left, right) => left.key.localeCompare(right.key)) };
}

function isErrno(error: unknown, code: string): boolean {
	return typeof error === "object" && error !== null && "code" in error && (error as { code?: unknown }).code === code;
}

interface JournalLock { path: string; token: string }

export class ProductionEffectJournal implements ProductionEffectJournalPort {
	readonly #root: string;

	constructor(root: string) {
		if (typeof root !== "string" || root.length === 0) throw new Error("production effect journal root is required");
		this.#root = root;
	}

	#path(): string { return join(this.#root, "production-effects.json"); }
	#lockPath(): string { return join(this.#root, ".production-effects.lock"); }

	async #ensureRoot(): Promise<void> {
		await mkdir(this.#root, { recursive: true, mode: 0o700 });
		const metadata = await lstat(this.#root);
		if (!metadata.isDirectory() || metadata.isSymbolicLink()) throw new Error("effect journal root must be a trusted directory");
	}

	async #syncRoot(): Promise<void> {
		if (process.platform === "win32") return;
		const handle = await open(this.#root, constants.O_RDONLY);
		try { await handle.sync(); } finally { await handle.close(); }
	}

	#processIsAlive(pid: number): boolean {
		try {
			process.kill(pid, 0);
			return true;
		} catch (error) {
			return !isErrno(error, "ESRCH");
		}
	}

	async #reclaimDeadLock(path: string): Promise<boolean> {
		let value: unknown;
		try { value = JSON.parse(await readFile(join(path, "owner.json"), "utf8")); } catch (error) {
			if (isErrno(error, "ENOENT")) return false;
			return false;
		}
		let owner: ReturnType<typeof exact>;
		try { owner = exact(value, ["schemaVersion", "pid", "token"], [], "effect journal lock owner"); } catch { return false; }
		if (owner.schemaVersion !== 1 || !Number.isSafeInteger(owner.pid) || (owner.pid as number) < 1
			|| typeof owner.token !== "string" || !/^[0-9a-f-]{36}$/.test(owner.token)) return false;
		if (this.#processIsAlive(owner.pid as number)) return false;
		const quarantine = `${path}.stale.${randomUUID()}`;
		try { await rename(path, quarantine); } catch (error) {
			if (isErrno(error, "ENOENT")) return true;
			throw error;
		}
		try {
			const moved = exact(
				JSON.parse(await readFile(join(quarantine, "owner.json"), "utf8")),
				["schemaVersion", "pid", "token"],
				[],
				"effect journal lock owner",
			);
			if (moved.token !== owner.token || moved.pid !== owner.pid) {
				throw new Error("effect journal stale lock identity changed during reclamation");
			}
		} finally { await rm(quarantine, { recursive: true, force: true }); }
		return true;
	}

	async #acquire(): Promise<JournalLock> {
		await this.#ensureRoot();
		const path = this.#lockPath();
		for (let attempt = 0; attempt < LOCK_ATTEMPTS; attempt += 1) {
			const token = randomUUID();
			try {
				await mkdir(path, { mode: 0o700 });
				const owner = await open(join(path, "owner.json"), "wx", 0o600);
				try {
					await owner.writeFile(JSON.stringify({ schemaVersion: 1, pid: process.pid, token }), "utf8");
					await owner.sync();
				} finally { await owner.close(); }
				return { path, token };
			} catch (error) {
				if (!isErrno(error, "EEXIST")) {
					await rm(path, { recursive: true, force: true });
					throw error;
				}
				if (await this.#reclaimDeadLock(path)) continue;
				await new Promise((resolve) => setTimeout(resolve, LOCK_RETRY_MS));
			}
		}
		throw new Error("timed out acquiring production effect journal lock");
	}

	async #release(lock: JournalLock): Promise<void> {
		let owner: unknown;
		try { owner = JSON.parse(await readFile(join(lock.path, "owner.json"), "utf8")); } catch {
			throw new Error("effect journal lock ownership disappeared before release");
		}
		const candidate = exact(owner, ["schemaVersion", "pid", "token"], [], "effect journal lock owner");
		if (candidate.schemaVersion !== 1 || candidate.pid !== process.pid || candidate.token !== lock.token) {
			throw new Error("effect journal lock ownership changed before release");
		}
		await rm(lock.path, { recursive: true });
	}

	async #read(): Promise<ProductionJournalDocument> {
		let handle: Awaited<ReturnType<typeof open>>;
		try { handle = await open(this.#path(), constants.O_RDONLY | (constants.O_NOFOLLOW ?? 0)); } catch (error) {
			if (isErrno(error, "ENOENT")) return { schemaVersion: 1, records: [] };
			throw error;
		}
		try {
			const metadata = await handle.stat();
			if (!metadata.isFile() || metadata.size > MAX_JOURNAL_BYTES) throw new Error("effect journal is not a bounded regular file");
			let value: unknown;
			try { value = JSON.parse(await handle.readFile("utf8")); } catch { throw new Error("invalid production effect journal JSON"); }
			return validateDocument(value);
		} finally { await handle.close(); }
	}

	async #write(documentValue: ProductionJournalDocument): Promise<void> {
		const document = validateDocument(documentValue);
		const serialized = `${JSON.stringify(document, null, 2)}\n`;
		if (Buffer.byteLength(serialized) > MAX_JOURNAL_BYTES) throw new Error("production effect journal exceeds its byte limit");
		const temporary = join(this.#root, `.production-effects.${randomUUID()}.tmp`);
		const handle = await open(temporary, "wx", 0o600);
		try {
			await handle.writeFile(serialized, "utf8");
			await handle.sync();
		} finally { await handle.close(); }
		try {
			await rename(temporary, this.#path());
			await this.#syncRoot();
		} catch (error) {
			await rm(temporary, { force: true });
			throw error;
		}
	}

	async #transact(operation: (records: ProductionEffectRecord[]) => ProductionEffectRecord): Promise<ProductionEffectRecord> {
		const lock = await this.#acquire();
		try {
			const document = await this.#read();
			const result = operation(document.records);
			await this.#write({ schemaVersion: 1, records: document.records });
			return structuredClone(result);
		} finally { await this.#release(lock); }
	}

	async prepare(intentValue: ProductionEffectIntent, now = new Date()): Promise<ProductionEffectRecord> {
		return this.#transact((records) => {
			const existingByKey = records.find((record) => record.key === intentValue.key);
			let intent: ProductionEffectIntent;
			try { intent = validateIntent(intentValue); } catch (error) {
				if (existingByKey) throw new Error("production effect key conflicts with a different exact intent");
				throw error;
			}
			const existing = records.find((record) => record.key === intent.key);
			if (existing) {
				const identity = ({ key, kind: effectKind, runId, generation, childId, intentDigest }: ProductionEffectRecord) =>
					({ key, kind: effectKind, runId, generation, ...(childId === undefined ? {} : { childId }), intentDigest });
				if (JSON.stringify(identity(existing)) !== JSON.stringify(intent)) {
					throw new Error("production effect key conflicts with a different exact intent");
				}
				return existing;
			}
			if (records.length >= MAX_EFFECTS) throw new Error("production effect journal is full");
			const record = validateProductionEffectRecord({
				schemaVersion: 1,
				...intent,
				phase: "prepared",
				preparedAt: now.toISOString(),
			});
			records.push(record);
			return record;
		});
	}

	async load(keyValue: string): Promise<ProductionEffectRecord | undefined> {
		const key = digest(keyValue, "effect key");
		await this.#ensureRoot();
		const record = (await this.#read()).records.find((candidate) => candidate.key === key);
		return record === undefined ? undefined : structuredClone(record);
	}

	async listNonApplied(): Promise<ProductionEffectRecord[]> {
		await this.#ensureRoot();
		return (await this.#read()).records
			.filter((record) => record.phase !== "applied")
			.map((record) => structuredClone(record));
	}

	async observe(
		keyValue: string,
		fence: ProductionEffectFence,
		resultDigestValue: string,
		now = new Date(),
	): Promise<ProductionEffectRecord> {
		const key = digest(keyValue, "effect key");
		const resultDigest = digest(resultDigestValue, "effect result digest");
		return this.#transact((records) => {
			const index = records.findIndex((record) => record.key === key);
			if (index < 0) throw new Error("production effect was not prepared");
			const current = records[index];
			if (current.runId !== fence.runId || current.generation !== fence.generation) {
				throw new Error("stale production effect fence");
			}
			if (current.phase !== "prepared") {
				if (current.resultDigest !== resultDigest) throw new Error("production effect result conflicts with its observation");
				return current;
			}
			const next = validateProductionEffectRecord({
				...current,
				phase: "observed",
				observedAt: now.toISOString(),
				resultDigest,
			});
			records[index] = next;
			return next;
		});
	}

	async apply(keyValue: string, fence: ProductionEffectFence, now = new Date()): Promise<ProductionEffectRecord> {
		const key = digest(keyValue, "effect key");
		return this.#transact((records) => {
			const index = records.findIndex((record) => record.key === key);
			if (index < 0) throw new Error("production effect was not prepared");
			const current = records[index];
			if (current.runId !== fence.runId || current.generation !== fence.generation) {
				throw new Error("stale production effect fence");
			}
			if (current.phase === "prepared") throw new Error("production effect must be observed before it is applied");
			if (current.phase === "applied") return current;
			const next = validateProductionEffectRecord({
				...current,
				phase: "applied",
				appliedAt: now.toISOString(),
			});
			records[index] = next;
			return next;
		});
	}
}
