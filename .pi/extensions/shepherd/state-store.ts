import { randomUUID } from "node:crypto";
import { constants } from "node:fs";
import { chmod, link, mkdir, open, readFile, rename, rm, stat } from "node:fs/promises";
import { isAbsolute, join } from "node:path";

import type { ShepherdLaneState, ShepherdRunState } from "./domain.ts";

const MAX_STATE_BYTES = 1_048_576;
const MAX_LEASE_BYTES = 4_096;
const DEFAULT_SUMMARY_LENGTH = 2_048;
const LEASE_SCHEMA_VERSION = 1;
const allowedRunStatuses = new Set(["pending", "running", "completed", "failed", "interrupted", "stopped", "halted"]);
const allowedLaneStatuses = new Set(["pending", "running", "succeeded", "failed", "interrupted", "stopped", "halted"]);
const allowedRunFields = new Set([
	"schemaVersion",
	"issue",
	"pr",
	"runId",
	"generation",
	"status",
	"candidateHead",
	"validationNonce",
	"createdAt",
	"updatedAt",
	"lanes",
	"score",
	"hardGates",
]);
const allowedLaneFields = new Set([
	"id",
	"mutating",
	"dependsOn",
	"role",
	"status",
	"summary",
	"score",
	"hardGates",
]);
const allowedLeaseFields = new Set(["schemaVersion", "issue", "runId", "pid", "token", "createdAt"]);

export interface RunLeaseClaim {
	issue: number;
	runId: string;
	mode: "start" | "resume";
}

export interface RunLease {
	assertOwned(): Promise<void>;
	release(): Promise<void>;
}

export interface FileStateStoreOptions {
	processId?: number;
	now?: () => Date;
	isProcessAlive?: (pid: number) => boolean | Promise<boolean>;
	tokenFactory?: () => string;
}

interface RunLeaseMetadata {
	schemaVersion: 1;
	issue: number;
	runId: string;
	pid: number;
	token: string;
	createdAt: string;
}

interface LeaseRecord {
	metadata: RunLeaseMetadata;
	device: number | bigint;
	inode: number | bigint;
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null && !Array.isArray(value);
}

function validIssue(issue: unknown): issue is number {
	return Number.isSafeInteger(issue) && (issue as number) > 0 && (issue as number) <= 2_147_483_647;
}

function validBoundedText(value: unknown, maximum: number): value is string {
	return typeof value === "string" && value.length > 0 && value.length <= maximum && !/[\u0000-\u001f\u007f-\u009f]/.test(value);
}

function assertOnlyAllowedFields(value: Record<string, unknown>, allowed: ReadonlySet<string>, description: string): void {
	for (const key of Object.keys(value)) {
		if (!allowed.has(key)) throw new Error(`invalid Shepherd state: unknown ${description} field ${JSON.stringify(key)}`);
	}
}

function validScore(value: unknown): value is number {
	return typeof value === "number" && Number.isFinite(value) && value >= 0 && value <= 1;
}

function validateHardGates(value: unknown, description: string): void {
	if (!Array.isArray(value) || value.length > 64 || value.some((gate) => !validBoundedText(gate, 128))) {
		throw new Error(`invalid Shepherd state: invalid ${description} hard gates`);
	}
}

function validateState(value: unknown, expectedIssue: number, allowUnknownFields = false): asserts value is ShepherdRunState {
	if (!isRecord(value)) throw new Error("invalid Shepherd state: expected an object");
	if (!allowUnknownFields) assertOnlyAllowedFields(value, allowedRunFields, "state");
	if (value.issue !== expectedIssue) throw new Error("Shepherd state issue identity mismatch");
	if (value.schemaVersion !== 1) throw new Error("invalid Shepherd state: unsupported schema version");
	if (!validIssue(value.issue)) throw new Error("invalid Shepherd state: invalid issue");
	if (value.pr !== undefined && !validIssue(value.pr)) throw new Error("invalid Shepherd state: invalid pull request");
	if (!validBoundedText(value.runId, 256)) throw new Error("invalid Shepherd state: invalid run id");
	if (!Number.isSafeInteger(value.generation) || (value.generation as number) < 1) {
		throw new Error("invalid Shepherd state: invalid generation");
	}
	if (typeof value.status !== "string" || !allowedRunStatuses.has(value.status)) {
		throw new Error("invalid Shepherd state: invalid run status");
	}
	if (typeof value.candidateHead !== "string" || !/^[0-9a-f]{40}$/i.test(value.candidateHead)) {
		throw new Error("invalid Shepherd state: invalid candidate head");
	}
	if (!validBoundedText(value.validationNonce, 256)) throw new Error("invalid Shepherd state: invalid nonce");
	if (!validBoundedText(value.createdAt, 128) || !validBoundedText(value.updatedAt, 128)) {
		throw new Error("invalid Shepherd state: invalid timestamps");
	}
	if (value.score !== undefined && !validScore(value.score)) {
		throw new Error("invalid Shepherd state: invalid score");
	}
	if (value.hardGates !== undefined) validateHardGates(value.hardGates, "run");
	if (!Array.isArray(value.lanes) || value.lanes.length > 64) throw new Error("invalid Shepherd state: invalid lanes");

	const laneIds = new Set<string>();
	for (const lane of value.lanes) {
		if (!isRecord(lane)) throw new Error("invalid Shepherd state: invalid lane");
		if (!allowUnknownFields) assertOnlyAllowedFields(lane, allowedLaneFields, "lane");
		if (!validBoundedText(lane.id, 128) || laneIds.has(lane.id)) {
			throw new Error("invalid Shepherd state: invalid lane identity");
		}
		laneIds.add(lane.id);
		if (typeof lane.mutating !== "boolean") throw new Error("invalid Shepherd state: invalid lane mutation flag");
		if (!validBoundedText(lane.role, 256)) throw new Error("invalid Shepherd state: invalid lane role");
		if (typeof lane.status !== "string" || !allowedLaneStatuses.has(lane.status)) {
			throw new Error("invalid Shepherd state: invalid lane status");
		}
		if (!Array.isArray(lane.dependsOn) || lane.dependsOn.length > 64 || lane.dependsOn.some((item) => !validBoundedText(item, 128))) {
			throw new Error("invalid Shepherd state: invalid lane dependencies");
		}
		if (lane.summary !== undefined && (
			typeof lane.summary !== "string" ||
			lane.summary.length > (allowUnknownFields ? MAX_STATE_BYTES : DEFAULT_SUMMARY_LENGTH) ||
			(!allowUnknownFields && /[\u0000-\u001f\u007f-\u009f]/.test(lane.summary))
		)) {
			throw new Error("invalid Shepherd state: invalid lane summary");
		}
		if (lane.score !== undefined && !validScore(lane.score)) {
			throw new Error("invalid Shepherd state: invalid lane score");
		}
		if (lane.hardGates !== undefined) validateHardGates(lane.hardGates, "lane");
	}
}

function redactSecrets(input: string): string {
	return input
		.replace(/-----BEGIN [^-\r\n]+ PRIVATE KEY-----[\s\S]*?-----END [^-\r\n]+ PRIVATE KEY-----/gi, "[REDACTED]")
		.replace(/\bAuthorization\s*:\s*(?:Bearer|Basic|Token)\s+[^\s,;]+/gi, "Authorization: [REDACTED]")
		.replace(/\bBearer\s+[^\s,;]+/gi, "Bearer [REDACTED]")
		.replace(/\b(token|access[_-]?token|api[_-]?key|password|secret)\s*[:=]\s*[^\s,;]+/gi, "$1=[REDACTED]")
		.replace(/\b(?:gh[pousr]_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,}|sk-[A-Za-z0-9_-]{20,})\b/g, "[REDACTED]");
}

export function sanitizeSummary(input: string, maximumLength = DEFAULT_SUMMARY_LENGTH): string {
	if (typeof input !== "string") throw new TypeError("summary must be text");
	if (!Number.isSafeInteger(maximumLength) || maximumLength < 1) {
		throw new RangeError("maximum summary length must be a positive integer");
	}
	const safe = redactSecrets(input)
		.replace(/\u001b\[[0-?]*[ -/]*[@-~]/g, "")
		.replace(/(?:\r\n|[\r\n\t\u2028\u2029])+/g, " ")
		.replace(/[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f-\u009f]/g, " ");
	return safe.slice(0, maximumLength);
}

function projectLane(lane: ShepherdLaneState): ShepherdLaneState {
	const persisted: ShepherdLaneState = {
		id: lane.id,
		mutating: lane.mutating,
		dependsOn: [...lane.dependsOn],
		role: lane.role,
		status: lane.status,
	};
	if (lane.summary !== undefined) persisted.summary = sanitizeSummary(lane.summary);
	if (lane.score !== undefined) persisted.score = lane.score;
	if (lane.hardGates !== undefined) persisted.hardGates = [...lane.hardGates];
	return persisted;
}

function projectState(state: ShepherdRunState): ShepherdRunState {
	const persisted: ShepherdRunState = {
		schemaVersion: state.schemaVersion,
		issue: state.issue,
		runId: state.runId,
		generation: state.generation,
		status: state.status,
		candidateHead: state.candidateHead,
		validationNonce: state.validationNonce,
		createdAt: state.createdAt,
		updatedAt: state.updatedAt,
		lanes: state.lanes.map(projectLane),
	};
	if (state.pr !== undefined) persisted.pr = state.pr;
	if (state.score !== undefined) persisted.score = state.score;
	if (state.hardGates !== undefined) persisted.hardGates = [...state.hardGates];
	return persisted;
}

function serializedState(state: ShepherdRunState): string {
	return `${JSON.stringify(state, null, 2)}\n`;
}

function validProcessId(value: unknown): value is number {
	return Number.isSafeInteger(value) && (value as number) > 0 && (value as number) <= 2_147_483_647;
}

function validateLeaseMetadata(value: unknown): asserts value is RunLeaseMetadata {
	if (!isRecord(value)) throw new Error("invalid Shepherd run lease: expected an object");
	for (const key of Object.keys(value)) {
		if (!allowedLeaseFields.has(key)) {
			throw new Error(`invalid Shepherd run lease: unknown metadata field ${JSON.stringify(key)}`);
		}
	}
	if (Object.keys(value).length !== allowedLeaseFields.size) {
		throw new Error("invalid Shepherd run lease: incomplete metadata");
	}
	if (value.schemaVersion !== LEASE_SCHEMA_VERSION) {
		throw new Error("invalid Shepherd run lease: unsupported schema version");
	}
	if (!validIssue(value.issue)) throw new Error("invalid Shepherd run lease: invalid issue");
	if (!validBoundedText(value.runId, 256)) throw new Error("invalid Shepherd run lease: invalid run id");
	if (!validProcessId(value.pid)) throw new Error("invalid Shepherd run lease: invalid process id");
	if (!validBoundedText(value.token, 256)) throw new Error("invalid Shepherd run lease: invalid owner token");
	if (!validBoundedText(value.createdAt, 128)) throw new Error("invalid Shepherd run lease: invalid timestamp");
	const timestamp = new Date(value.createdAt);
	if (!Number.isFinite(timestamp.valueOf()) || timestamp.toISOString() !== value.createdAt) {
		throw new Error("invalid Shepherd run lease: timestamp must be canonical ISO-8601");
	}
}

function validateLeaseClaim(claim: RunLeaseClaim): void {
	if (!isRecord(claim)) throw new TypeError("Shepherd run lease claim must be an object");
	if (!validIssue(claim.issue)) throw new RangeError("lease issue must be a positive bounded integer");
	if (!validBoundedText(claim.runId, 256)) throw new TypeError("lease run id must be bounded text without control characters");
	if (claim.mode !== "start" && claim.mode !== "resume") {
		throw new TypeError("lease mode must be start or resume");
	}
}

function defaultProcessLiveness(pid: number): boolean {
	try {
		process.kill(pid, 0);
		return true;
	} catch (error) {
		const code = (error as NodeJS.ErrnoException).code;
		if (code === "ESRCH") return false;
		if (code === "EPERM") return true;
		throw new Error(`unable to determine whether Shepherd lease process ${pid} is alive`, { cause: error });
	}
}

function sameLeaseOwner(left: RunLeaseMetadata, right: RunLeaseMetadata): boolean {
	return left.token === right.token
		&& left.issue === right.issue
		&& left.runId === right.runId
		&& left.pid === right.pid
		&& left.createdAt === right.createdAt;
}

async function readLeaseRecord(path: string): Promise<LeaseRecord | undefined> {
	let handle;
	try {
		handle = await open(path, constants.O_RDONLY | constants.O_NOFOLLOW);
	} catch (error) {
		if ((error as NodeJS.ErrnoException).code === "ENOENT") return undefined;
		throw new Error("unable to open Shepherd run lease safely", { cause: error });
	}

	try {
		const file = await handle.stat();
		if (!file.isFile() || file.size < 1 || file.size > MAX_LEASE_BYTES) {
			throw new Error("invalid Shepherd run lease: lock is not a bounded regular file");
		}
		if ((file.mode & 0o777) !== 0o600) {
			throw new Error("invalid Shepherd run lease: lock permissions must be 0600");
		}
		let parsed: unknown;
		try {
			parsed = JSON.parse(await handle.readFile("utf8"));
		} catch (error) {
			throw new Error("invalid Shepherd run lease: malformed JSON", { cause: error });
		}
		validateLeaseMetadata(parsed);
		return { metadata: parsed, device: file.dev, inode: file.ino };
	} finally {
		await handle.close();
	}
}

export class FileStateStore {
	readonly root: string;
	private readonly processId: number;
	private readonly now: () => Date;
	private readonly isProcessAlive: (pid: number) => boolean | Promise<boolean>;
	private readonly tokenFactory: () => string;

	constructor(root: string, options: FileStateStoreOptions = {}) {
		if (typeof root !== "string" || !isAbsolute(root) || /[\u0000-\u001f\u007f]/.test(root)) {
			throw new TypeError("Shepherd state root must be an absolute path without control characters");
		}
		const processId = options.processId ?? process.pid;
		if (!validProcessId(processId)) throw new RangeError("Shepherd lease process id must be a positive bounded integer");
		if (options.now !== undefined && typeof options.now !== "function") throw new TypeError("Shepherd lease clock must be a function");
		if (options.isProcessAlive !== undefined && typeof options.isProcessAlive !== "function") {
			throw new TypeError("Shepherd process liveness check must be a function");
		}
		if (options.tokenFactory !== undefined && typeof options.tokenFactory !== "function") {
			throw new TypeError("Shepherd lease token factory must be a function");
		}
		this.root = root;
		this.processId = processId;
		this.now = options.now ?? (() => new Date());
		this.isProcessAlive = options.isProcessAlive ?? defaultProcessLiveness;
		this.tokenFactory = options.tokenFactory ?? randomUUID;
	}

	private pathFor(issue: number): string {
		if (!validIssue(issue)) throw new RangeError("issue must be a positive bounded integer");
		return join(this.root, `issue-${issue}.json`);
	}

	private leasePath(): string {
		return join(this.root, "active.lock");
	}

	private async prepareRoot(): Promise<void> {
		await mkdir(this.root, { recursive: true, mode: 0o700 });
		await chmod(this.root, 0o700);
	}

	private createLeaseMetadata(claim: RunLeaseClaim): RunLeaseMetadata {
		const now = this.now();
		if (!(now instanceof Date) || !Number.isFinite(now.valueOf())) {
			throw new Error("Shepherd lease clock returned an invalid Date");
		}
		const token = this.tokenFactory();
		const metadata: RunLeaseMetadata = {
			schemaVersion: LEASE_SCHEMA_VERSION,
			issue: claim.issue,
			runId: claim.runId,
			pid: this.processId,
			token,
			createdAt: now.toISOString(),
		};
		validateLeaseMetadata(metadata);
		return metadata;
	}

	private async tryCreateLease(metadata: RunLeaseMetadata): Promise<LeaseRecord | undefined> {
		const path = this.leasePath();
		let handle;
		try {
			handle = await open(path, "wx", 0o600);
		} catch (error) {
			if ((error as NodeJS.ErrnoException).code === "EEXIST") return undefined;
			throw new Error("unable to acquire Shepherd run lease", { cause: error });
		}

		try {
			await handle.chmod(0o600);
			const payload = `${JSON.stringify(metadata)}\n`;
			if (Buffer.byteLength(payload, "utf8") > MAX_LEASE_BYTES) {
				throw new Error("invalid Shepherd run lease: serialized metadata is too large");
			}
			await handle.writeFile(payload, "utf8");
			await handle.sync();
		} catch (error) {
			await handle.close().catch(() => undefined);
			handle = undefined;
			await rm(path, { force: true }).catch(() => undefined);
			throw error;
		} finally {
			await handle?.close();
		}

		const acquired = await readLeaseRecord(path);
		if (!acquired || !sameLeaseOwner(acquired.metadata, metadata)) {
			throw new Error("Shepherd run lease ownership was lost during acquisition");
		}
		return acquired;
	}

	private async ownerIsAlive(metadata: RunLeaseMetadata): Promise<boolean> {
		const alive = await this.isProcessAlive(metadata.pid);
		if (typeof alive !== "boolean") {
			throw new Error("Shepherd process liveness check returned a non-boolean result");
		}
		return alive;
	}

	private sameLeaseRecord(left: LeaseRecord, right: LeaseRecord): boolean {
		return sameLeaseOwner(left.metadata, right.metadata)
			&& left.device === right.device
			&& left.inode === right.inode;
	}

	private async restoreDisplacedLease(path: string): Promise<void> {
		try {
			await link(path, this.leasePath());
		} catch (error) {
			if ((error as NodeJS.ErrnoException).code === "EEXIST") {
				throw new Error(`Shepherd lease changed concurrently; displaced lock retained at ${path}`);
			}
			throw new Error(`unable to restore concurrently displaced Shepherd lease retained at ${path}`, { cause: error });
		}
		await rm(path);
	}

	private async removeStaleLease(expected: LeaseRecord, claim: RunLeaseClaim): Promise<boolean> {
		const current = await readLeaseRecord(this.leasePath());
		if (!current) return false;
		if (!this.sameLeaseRecord(current, expected)) return false;
		if (current.metadata.issue !== claim.issue) {
			throw new Error(
				`Shepherd run lease is stale for issue #${current.metadata.issue}; resume that issue before running issue #${claim.issue}`,
			);
		}
		if (await this.ownerIsAlive(current.metadata)) {
			throw new Error(
				`Shepherd run lease is held by live process ${current.metadata.pid} for issue #${current.metadata.issue}`,
			);
		}

		const displacedPath = join(this.root, `.active.lock.stale.${randomUUID()}`);
		try {
			await rename(this.leasePath(), displacedPath);
		} catch (error) {
			if ((error as NodeJS.ErrnoException).code === "ENOENT") return false;
			throw new Error("unable to take over stale Shepherd run lease", { cause: error });
		}
		const displaced = await readLeaseRecord(displacedPath);
		if (!displaced || !this.sameLeaseRecord(displaced, current)) {
			await this.restoreDisplacedLease(displacedPath);
			throw new Error("Shepherd run lease changed concurrently during stale takeover");
		}
		await rm(displacedPath);
		return true;
	}

	private async assertLeaseOwned(expected: LeaseRecord): Promise<void> {
		const current = await readLeaseRecord(this.leasePath());
		if (!current || !this.sameLeaseRecord(current, expected)) {
			throw new Error("Shepherd run lease ownership was lost (owner token mismatch)");
		}
	}

	private async releaseLease(expected: LeaseRecord): Promise<void> {
		await this.assertLeaseOwned(expected);
		const displacedPath = join(this.root, `.active.lock.release.${randomUUID()}`);
		try {
			await rename(this.leasePath(), displacedPath);
		} catch (error) {
			throw new Error("Shepherd run lease ownership was lost before release", { cause: error });
		}
		const displaced = await readLeaseRecord(displacedPath);
		if (!displaced || !this.sameLeaseRecord(displaced, expected)) {
			await this.restoreDisplacedLease(displacedPath);
			throw new Error("Shepherd run lease ownership changed during release; another owner's lock was not deleted");
		}
		await rm(displacedPath);
	}

	async acquireLease(claim: RunLeaseClaim): Promise<RunLease> {
		validateLeaseClaim(claim);
		await this.prepareRoot();
		const metadata = this.createLeaseMetadata(claim);
		for (let attempt = 0; attempt < 16; attempt += 1) {
			const acquired = await this.tryCreateLease(metadata);
			if (acquired) {
				let released = false;
				return {
					assertOwned: async () => {
						if (released) throw new Error("Shepherd run lease was already released");
						await this.assertLeaseOwned(acquired);
					},
					release: async () => {
						if (released) throw new Error("Shepherd run lease was already released");
						await this.releaseLease(acquired);
						released = true;
					},
				};
			}

			const existing = await readLeaseRecord(this.leasePath());
			if (!existing) continue;
			if (await this.ownerIsAlive(existing.metadata)) {
				throw new Error(
					`Shepherd run lease is held by live process ${existing.metadata.pid} for issue #${existing.metadata.issue} (run ${existing.metadata.runId})`,
				);
			}
			if (claim.mode !== "resume") {
				throw new Error(
					`Shepherd run lease is stale for issue #${existing.metadata.issue} (dead process ${existing.metadata.pid}); use resume to recover it`,
				);
			}
			await this.removeStaleLease(existing, claim);
		}
		throw new Error("unable to acquire Shepherd run lease after repeated concurrent changes");
	}

	async load(issue: number): Promise<ShepherdRunState | undefined> {
		const path = this.pathFor(issue);
		let metadata;
		try {
			metadata = await stat(path);
		} catch (error) {
			if ((error as NodeJS.ErrnoException).code === "ENOENT") return undefined;
			throw error;
		}
		if (!metadata.isFile() || metadata.size > MAX_STATE_BYTES) {
			throw new Error("invalid Shepherd state: state file is not a bounded regular file");
		}

		let parsed: unknown;
		try {
			parsed = JSON.parse(await readFile(path, "utf8"));
		} catch (error) {
			throw new Error("invalid Shepherd state: malformed JSON", { cause: error });
		}
		validateState(parsed, issue);
		return parsed;
	}

	async save(state: ShepherdRunState): Promise<void> {
		if (!validIssue(state?.issue)) throw new Error("invalid Shepherd state: invalid issue");
		validateState(state, state.issue, true);
		const persisted = projectState(state);
		validateState(persisted, persisted.issue);
		await mkdir(this.root, { recursive: true, mode: 0o700 });
		await chmod(this.root, 0o700);

		const destination = this.pathFor(state.issue);
		const temporary = join(this.root, `.issue-${state.issue}.${process.pid}.${randomUUID()}.tmp`);
		const payload = serializedState(persisted);
		if (Buffer.byteLength(payload, "utf8") > MAX_STATE_BYTES) {
			throw new Error("invalid Shepherd state: serialized state is too large");
		}

		let handle;
		try {
			handle = await open(temporary, "wx", 0o600);
			await handle.writeFile(payload, "utf8");
			await handle.sync();
			await handle.close();
			handle = undefined;
			await rename(temporary, destination);
			await chmod(destination, 0o600);
		} finally {
			await handle?.close().catch(() => undefined);
			await rm(temporary, { force: true }).catch(() => undefined);
		}
	}
}
