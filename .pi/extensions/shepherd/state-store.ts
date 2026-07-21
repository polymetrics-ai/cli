import { createHash, randomUUID } from "node:crypto";
import { execFile } from "node:child_process";
import { constants } from "node:fs";
import { link, lstat, mkdir, open, realpath, rename, rm } from "node:fs/promises";
import { basename, dirname, isAbsolute, join, parse, relative, resolve, sep } from "node:path";

import type { ShepherdLaneState, ShepherdRunState } from "./domain.ts";

const MAX_STATE_BYTES = 1_048_576;
const MAX_LEASE_BYTES = 4_096;
const MAX_LEASE_CHAIN = 512;
const DEFAULT_SUMMARY_LENGTH = 2_048;
const LEASE_SCHEMA_VERSION = 1;
const allowedRunStatuses = new Set(["pending", "running", "completed", "failed", "interrupted", "stopped", "halted"]);
const allowedLaneStatuses = new Set(["pending", "running", "succeeded", "failed", "interrupted", "stopped", "halted"]);
const terminalLaneStatuses = new Set(["succeeded", "failed", "interrupted", "stopped", "halted"]);
const allowedRunFields = new Set([
	"schemaVersion", "issue", "pr", "runId", "generation", "status", "candidateHead",
	"validationNonce", "createdAt", "updatedAt", "lanes", "score", "hardGates",
]);
const allowedLaneFields = new Set(["id", "mutating", "dependsOn", "role", "status", "summary", "score", "hardGates"]);
const requiredOwnerLeaseFields = new Set(["schemaVersion", "issue", "runId", "pid", "token", "createdAt"]);
const allowedOwnerLeaseFields = new Set([...requiredOwnerLeaseFields, "ownerIdentity"]);
const allowedReleaseLeaseFields = new Set(["schemaVersion", "recordType", "releasedLeaseToken", "token", "createdAt"]);

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
	processIdentity?: string;
	now?: () => Date;
	isProcessAlive?: (pid: number) => boolean | Promise<boolean>;
	getProcessIdentity?: (pid: number) => string | undefined | Promise<string | undefined>;
	tokenFactory?: () => string;
	trustedRoot?: string;
	testHooks?: {
		afterStateOpen?: () => void | Promise<void>;
	};
}

interface OwnerLeaseMetadata {
	schemaVersion: 1;
	issue: number;
	runId: string;
	pid: number;
	token: string;
	createdAt: string;
	ownerIdentity?: string;
}

interface ReleaseLeaseMetadata {
	schemaVersion: 1;
	recordType: "released";
	releasedLeaseToken: string;
	token: string;
	createdAt: string;
}

type LeaseMetadata = OwnerLeaseMetadata | ReleaseLeaseMetadata;

interface LeaseRecord {
	metadata: LeaseMetadata;
	device: number | bigint;
	inode: number | bigint;
}

interface CorruptLeaseRecord {
	kind: "crash-partial";
	device: number | bigint;
	inode: number | bigint;
}

interface RootGuard {
	path: string;
	handle: Awaited<ReturnType<typeof open>>;
	device: number | bigint;
	inode: number | bigint;
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null && !Array.isArray(value);
}

async function readBounded(handle: Awaited<ReturnType<typeof open>>, maximum: number, description: string): Promise<string> {
	const buffer: Uint8Array<ArrayBuffer> = new Uint8Array(maximum + 1);
	let offset = 0;
	while (offset <= maximum) {
		const { bytesRead } = await handle.read(buffer, offset, maximum + 1 - offset, null);
		if (bytesRead === 0) return new TextDecoder().decode(buffer.subarray(0, offset));
		offset += bytesRead;
	}
	throw new Error(`${description} exceeds its byte limit`);
}

function validIssue(issue: unknown): issue is number {
	return Number.isSafeInteger(issue) && (issue as number) > 0 && (issue as number) <= 2_147_483_647;
}

function validBoundedText(value: unknown, maximum: number): value is string {
	return typeof value === "string" && value.length > 0 && value.length <= maximum && !/[\u0000-\u001f\u007f-\u009f]/.test(value);
}

function validLeaseToken(value: unknown): value is string {
	return typeof value === "string" && /^[A-Za-z0-9_-]{1,128}$/.test(value);
}

function assertOnlyAllowedFields(value: Record<string, unknown>, allowed: ReadonlySet<string>, description: string): void {
	for (const key of Object.keys(value)) {
		if (!allowed.has(key)) throw new Error(`invalid Shepherd state: unknown ${description} field ${JSON.stringify(key)}`);
	}
}

function validScore(value: unknown): value is number {
	return typeof value === "number" && Number.isFinite(value) && value >= 0 && value <= 1;
}

function canonicalTimestamp(value: unknown, description: string): string {
	if (!validBoundedText(value, 128)) throw new Error(`invalid Shepherd state: invalid ${description}`);
	const timestamp = new Date(value);
	if (!Number.isFinite(timestamp.valueOf()) || timestamp.toISOString() !== value) {
		throw new Error(`invalid Shepherd state: ${description} must be canonical ISO-8601`);
	}
	return value;
}

function validateHardGates(value: unknown, description: string): void {
	if (!Array.isArray(value) || value.length > 64 || value.some((gate) => !validBoundedText(gate, 128))) {
		throw new Error(`invalid Shepherd state: invalid ${description} hard gates`);
	}
}

function validateDependencyGraph(lanes: ShepherdLaneState[]): void {
	const ids = new Set(lanes.map((lane) => lane.id));
	for (const lane of lanes) {
		const dependencies = new Set<string>();
		for (const dependency of lane.dependsOn) {
			if (!ids.has(dependency)) throw new Error("invalid Shepherd state: lane dependency does not exist");
			if (dependency === lane.id) throw new Error("invalid Shepherd state: lane cannot depend on itself");
			if (dependencies.has(dependency)) throw new Error("invalid Shepherd state: duplicate lane dependency");
			dependencies.add(dependency);
		}
	}

	const visiting = new Set<string>();
	const visited = new Set<string>();
	const byId = new Map(lanes.map((lane) => [lane.id, lane]));
	const visit = (id: string): void => {
		if (visiting.has(id)) throw new Error("invalid Shepherd state: cyclic lane dependencies");
		if (visited.has(id)) return;
		visiting.add(id);
		for (const dependency of byId.get(id)?.dependsOn ?? []) visit(dependency);
		visiting.delete(id);
		visited.add(id);
	};
	for (const lane of lanes) visit(lane.id);
}

function validateRunLaneCoherence(value: ShepherdRunState): void {
	const statuses = value.lanes.map((lane) => lane.status);
	const every = (allowed: ReadonlySet<string>) => statuses.every((status) => allowed.has(status));
	if (value.status === "pending" && !every(new Set(["pending"]))) {
		throw new Error("invalid Shepherd state: pending run has non-pending lanes");
	}
	if (value.status === "running" && !every(new Set(["pending", "running"]))) {
		throw new Error("invalid Shepherd state: running run has terminal lanes");
	}
	if (value.status === "completed" && (statuses.length === 0 || !every(new Set(["succeeded"])))) {
		throw new Error("invalid Shepherd state: completed run requires successful lanes");
	}
	if (value.status === "failed" && (!every(terminalLaneStatuses) || !statuses.includes("failed"))) {
		throw new Error("invalid Shepherd state: failed run requires a failed terminal lane");
	}
	if (value.status === "halted" && (
		!every(terminalLaneStatuses) || (!statuses.includes("halted") && (value.hardGates?.length ?? 0) === 0)
	)) {
		throw new Error("invalid Shepherd state: halted run requires a halted lane or hard gate");
	}
	if (value.status === "stopped" && (!every(new Set(["succeeded", "stopped"])) || !statuses.includes("stopped"))) {
		throw new Error("invalid Shepherd state: stopped run has incompatible lanes");
	}
	if (value.status === "interrupted" && !every(new Set(["pending", "succeeded", "interrupted"]))) {
		throw new Error("invalid Shepherd state: interrupted run has incompatible lanes");
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
	if (!Number.isSafeInteger(value.generation) || (value.generation as number) < 1) throw new Error("invalid Shepherd state: invalid generation");
	if (typeof value.status !== "string" || !allowedRunStatuses.has(value.status)) throw new Error("invalid Shepherd state: invalid run status");
	if (typeof value.candidateHead !== "string" || !/^[0-9a-f]{40}$/i.test(value.candidateHead)) {
		throw new Error("invalid Shepherd state: invalid candidate head");
	}
	if (!validBoundedText(value.validationNonce, 256)) throw new Error("invalid Shepherd state: invalid nonce");
	const createdAt = canonicalTimestamp(value.createdAt, "created timestamp");
	const updatedAt = canonicalTimestamp(value.updatedAt, "updated timestamp");
	if (updatedAt < createdAt) throw new Error("invalid Shepherd state: updated timestamp precedes creation");
	if (value.score !== undefined && !validScore(value.score)) throw new Error("invalid Shepherd state: invalid score");
	if (value.hardGates !== undefined) validateHardGates(value.hardGates, "run");
	if (!Array.isArray(value.lanes) || value.lanes.length > 64) throw new Error("invalid Shepherd state: invalid lanes");

	const laneIds = new Set<string>();
	for (const lane of value.lanes) {
		if (!isRecord(lane)) throw new Error("invalid Shepherd state: invalid lane");
		if (!allowUnknownFields) assertOnlyAllowedFields(lane, allowedLaneFields, "lane");
		if (!validBoundedText(lane.id, 128) || laneIds.has(lane.id)) throw new Error("invalid Shepherd state: invalid lane identity");
		laneIds.add(lane.id);
		if (typeof lane.mutating !== "boolean") throw new Error("invalid Shepherd state: invalid lane mutation flag");
		if (!validBoundedText(lane.role, 256)) throw new Error("invalid Shepherd state: invalid lane role");
		if (typeof lane.status !== "string" || !allowedLaneStatuses.has(lane.status)) throw new Error("invalid Shepherd state: invalid lane status");
		if (!Array.isArray(lane.dependsOn) || lane.dependsOn.length > 64 || lane.dependsOn.some((item) => !validBoundedText(item, 128))) {
			throw new Error("invalid Shepherd state: invalid lane dependencies");
		}
		if (lane.summary !== undefined && (
			typeof lane.summary !== "string" || lane.summary.length > (allowUnknownFields ? MAX_STATE_BYTES : DEFAULT_SUMMARY_LENGTH)
			|| (!allowUnknownFields && /[\u0000-\u001f\u007f-\u009f]/.test(lane.summary))
		)) throw new Error("invalid Shepherd state: invalid lane summary");
		if (lane.score !== undefined && !validScore(lane.score)) throw new Error("invalid Shepherd state: invalid lane score");
		if (lane.hardGates !== undefined) validateHardGates(lane.hardGates, "lane");
	}
	validateDependencyGraph(value.lanes);
	validateRunLaneCoherence(value as unknown as ShepherdRunState);
}

function redactSecrets(input: string): string {
	const secretName = "(?:token|access[_-]?token|api[_-]?key|password|secret|client[_-]?secret|private[_-]?key|database[_-]?url)";
	return input
		.replace(/-----BEGIN [^-\r\n]+ PRIVATE KEY-----[\s\S]*?-----END [^-\r\n]+ PRIVATE KEY-----/gi, "[REDACTED]")
		.replace(/\bAuthorization\s*:\s*[^\r\n,;]+/gi, "Authorization: [REDACTED]")
		.replace(/\b(?:Bearer|Basic|Token)\s+[^\s,;]+/gi, "[REDACTED]")
		.replace(new RegExp(`(["']${secretName}["']\\s*:\\s*)["'][^"'\\r\\n]*["']`, "gi"), "$1\"[REDACTED]\"")
		.replace(/\b[A-Z][A-Z0-9_]*(?:TOKEN|SECRET|PASSWORD|API_KEY|PRIVATE_KEY|DATABASE_URL)\s*=\s*(?:"[^"]*"|'[^']*'|[^\s,;]+)/g, "SECRET=[REDACTED]")
		.replace(new RegExp(`\\b(${secretName})\\s*[:=]\\s*(?:"[^"]*"|'[^']*'|[^\\s,;]+)`, "gi"), "$1=[REDACTED]")
		.replace(/\b([a-z][a-z0-9+.-]*:\/\/)[^\s\/@:]+:[^\s\/@]+@/gi, "$1[REDACTED]@")
		.replace(/([?&](?:token|api[_-]?key|password|secret)=)[^&#\s]+/gi, "$1[REDACTED]")
		.replace(/\b(?:gh[pousr]_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,}|sk-[A-Za-z0-9_-]{20,})\b/g, "[REDACTED]");
}

export function sanitizeSummary(input: string, maximumLength = DEFAULT_SUMMARY_LENGTH): string {
	if (typeof input !== "string") throw new TypeError("summary must be text");
	if (!Number.isSafeInteger(maximumLength) || maximumLength < 1) throw new RangeError("maximum summary length must be a positive integer");
	const safe = redactSecrets(input)
		.replace(/\u001b\[[0-?]*[ -/]*[@-~]/g, "")
		.replace(/(?:\r\n|[\r\n\t\u2028\u2029])+/g, " ")
		.replace(/[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f-\u009f]/g, " ");
	return safe.slice(0, maximumLength);
}

function projectLane(lane: ShepherdLaneState): ShepherdLaneState {
	const persisted: ShepherdLaneState = {
		id: lane.id, mutating: lane.mutating, dependsOn: [...lane.dependsOn], role: lane.role, status: lane.status,
	};
	// Model/provider text is never a persistence boundary: keep only a host-derived fixed category.
	if (lane.summary !== undefined) persisted.summary = `lane_${lane.status}`;
	if (lane.score !== undefined) persisted.score = lane.score;
	if (lane.hardGates !== undefined) persisted.hardGates = [...lane.hardGates];
	return persisted;
}

function projectState(state: ShepherdRunState): ShepherdRunState {
	const persisted: ShepherdRunState = {
		schemaVersion: state.schemaVersion, issue: state.issue, runId: state.runId, generation: state.generation,
		status: state.status, candidateHead: state.candidateHead, validationNonce: state.validationNonce,
		createdAt: state.createdAt, updatedAt: state.updatedAt, lanes: state.lanes.map(projectLane),
	};
	if (state.pr !== undefined) persisted.pr = state.pr;
	if (state.score !== undefined) persisted.score = state.score;
	if (state.hardGates !== undefined) persisted.hardGates = [...state.hardGates];
	return persisted;
}

function validProcessId(value: unknown): value is number {
	return Number.isSafeInteger(value) && (value as number) > 0 && (value as number) <= 2_147_483_647;
}

function validateCanonicalLeaseTimestamp(value: unknown): asserts value is string {
	if (!validBoundedText(value, 128)) throw new Error("invalid Shepherd run lease: invalid timestamp");
	const timestamp = new Date(value);
	if (!Number.isFinite(timestamp.valueOf()) || timestamp.toISOString() !== value) {
		throw new Error("invalid Shepherd run lease: timestamp must be canonical ISO-8601");
	}
}

function validateOwnerLeaseMetadata(value: unknown): asserts value is OwnerLeaseMetadata {
	if (!isRecord(value)) throw new Error("invalid Shepherd run lease: expected owner metadata");
	for (const key of Object.keys(value)) {
		if (!allowedOwnerLeaseFields.has(key)) throw new Error(`invalid Shepherd run lease: unknown metadata field ${JSON.stringify(key)}`);
	}
	for (const key of requiredOwnerLeaseFields) {
		if (!(key in value)) throw new Error("invalid Shepherd run lease: incomplete metadata");
	}
	if (value.schemaVersion !== LEASE_SCHEMA_VERSION) throw new Error("invalid Shepherd run lease: unsupported schema version");
	if (!validIssue(value.issue)) throw new Error("invalid Shepherd run lease: invalid issue");
	if (!validBoundedText(value.runId, 256)) throw new Error("invalid Shepherd run lease: invalid run id");
	if (!validProcessId(value.pid)) throw new Error("invalid Shepherd run lease: invalid process id");
	if (!validLeaseToken(value.token)) throw new Error("invalid Shepherd run lease: invalid owner token");
	if (value.ownerIdentity !== undefined && !validBoundedText(value.ownerIdentity, 256)) {
		throw new Error("invalid Shepherd run lease: invalid process identity");
	}
	validateCanonicalLeaseTimestamp(value.createdAt);
}

function validateLeaseMetadata(value: unknown): asserts value is LeaseMetadata {
	if (!isRecord(value)) throw new Error("invalid Shepherd run lease: expected an object");
	if (value.recordType !== "released") {
		validateOwnerLeaseMetadata(value);
		return;
	}
	for (const key of Object.keys(value)) {
		if (!allowedReleaseLeaseFields.has(key)) throw new Error(`invalid Shepherd run lease: unknown release field ${JSON.stringify(key)}`);
	}
	if (Object.keys(value).length !== allowedReleaseLeaseFields.size || value.schemaVersion !== LEASE_SCHEMA_VERSION) {
		throw new Error("invalid Shepherd run lease: incomplete release metadata");
	}
	if (!validLeaseToken(value.releasedLeaseToken) || !validLeaseToken(value.token)) {
		throw new Error("invalid Shepherd run lease: invalid release token");
	}
	validateCanonicalLeaseTimestamp(value.createdAt);
}

function validateLeaseClaim(claim: RunLeaseClaim): void {
	if (!isRecord(claim)) throw new TypeError("Shepherd run lease claim must be an object");
	if (!validIssue(claim.issue)) throw new RangeError("lease issue must be a positive bounded integer");
	if (!validBoundedText(claim.runId, 256)) throw new TypeError("lease run id must be bounded text without control characters");
	if (claim.mode !== "start" && claim.mode !== "resume") throw new TypeError("lease mode must be start or resume");
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

async function readSmallFile(path: string, maximum: number): Promise<string | undefined> {
	let handle;
	try {
		handle = await open(path, constants.O_RDONLY | constants.O_NOFOLLOW);
	} catch (error) {
		if ((error as NodeJS.ErrnoException).code === "ENOENT") return undefined;
		return undefined;
	}
	try {
		const metadata = await handle.stat();
		if (!metadata.isFile() || metadata.size < 1 || metadata.size > maximum) return undefined;
		return await readBounded(handle, maximum, "process identity file");
	} finally {
		await handle.close();
	}
}

async function defaultProcessIdentity(pid: number): Promise<string | undefined> {
	if (process.platform === "linux") {
		const [bootId, processStat] = await Promise.all([
			readSmallFile("/proc/sys/kernel/random/boot_id", 128),
			readSmallFile(`/proc/${pid}/stat`, 8_192),
		]);
		if (!bootId || !processStat) return undefined;
		const closeParen = processStat.lastIndexOf(")");
		if (closeParen < 0) return undefined;
		const fields = processStat.slice(closeParen + 2).trim().split(/\s+/);
		const startTime = fields[19];
		if (!startTime || !/^\d+$/.test(startTime)) return undefined;
		return `${bootId.trim()}:${startTime}`;
	}
	if (process.platform !== "darwin" && process.platform !== "freebsd") return undefined;
	return new Promise((resolveIdentity) => {
		execFile(
			"/bin/ps",
			["-p", String(pid), "-o", "lstart="],
			{ encoding: "utf8", timeout: 1_000, maxBuffer: 1_024, windowsHide: true },
			(error, stdout) => {
				const startedAt = stdout.trim();
				resolveIdentity(error || !startedAt ? undefined : `ps:${startedAt}`);
			},
		);
	});
}

function leaseToken(metadata: LeaseMetadata): string {
	return metadata.token;
}

function isOwner(metadata: LeaseMetadata): metadata is OwnerLeaseMetadata {
	return !("recordType" in metadata);
}

function sameLeaseOwner(left: OwnerLeaseMetadata, right: OwnerLeaseMetadata): boolean {
	return left.token === right.token && left.issue === right.issue && left.runId === right.runId
		&& left.pid === right.pid && left.createdAt === right.createdAt && left.ownerIdentity === right.ownerIdentity;
}

function sameLeaseRecord(left: LeaseRecord, right: LeaseRecord): boolean {
	return leaseToken(left.metadata) === leaseToken(right.metadata) && left.device === right.device && left.inode === right.inode;
}

function successorName(token: string): string {
	return `.active.next.${createHash("sha256").update(token).digest("hex")}`;
}

export class FileStateStore {
	readonly root: string;
	private readonly processId: number;
	private readonly processIdentity?: string;
	private readonly now: () => Date;
	private readonly isProcessAlive: (pid: number) => boolean | Promise<boolean>;
	private readonly getProcessIdentity: (pid: number) => string | undefined | Promise<string | undefined>;
	private readonly tokenFactory: () => string;
	private readonly trustedRoot?: string;
	private readonly testHooks?: FileStateStoreOptions["testHooks"];
	private canonicalRoot?: string;

	constructor(root: string, options: FileStateStoreOptions = {}) {
		if (typeof root !== "string" || !isAbsolute(root) || /[\u0000-\u001f\u007f]/.test(root)) {
			throw new TypeError("Shepherd state root must be an absolute path without control characters");
		}
		this.root = resolve(root);
		if (this.root === parse(this.root).root) throw new TypeError("Shepherd state root cannot be a broad filesystem root");
		if (options.trustedRoot !== undefined) {
			if (!isAbsolute(options.trustedRoot) || /[\u0000-\u001f\u007f]/.test(options.trustedRoot)) {
				throw new TypeError("Shepherd trusted root must be an absolute path without control characters");
			}
			this.trustedRoot = resolve(options.trustedRoot);
			const relation = relative(this.trustedRoot, this.root);
			if (relation === "" || relation === ".." || relation.startsWith(`..${sep}`) || isAbsolute(relation)) {
				throw new TypeError("Shepherd state root must be beneath the trusted root");
			}
		} else if (basename(dirname(this.root)) === "shepherd") {
			// Production layout is <Pi agent dir>/shepherd/<repository fingerprint>.
			this.trustedRoot = dirname(dirname(this.root));
		}
		const processId = options.processId ?? process.pid;
		if (!validProcessId(processId)) throw new RangeError("Shepherd lease process id must be a positive bounded integer");
		if (options.now !== undefined && typeof options.now !== "function") throw new TypeError("Shepherd lease clock must be a function");
		if (options.isProcessAlive !== undefined && typeof options.isProcessAlive !== "function") {
			throw new TypeError("Shepherd process liveness check must be a function");
		}
		if (options.getProcessIdentity !== undefined && typeof options.getProcessIdentity !== "function") {
			throw new TypeError("Shepherd process identity lookup must be a function");
		}
		if (options.tokenFactory !== undefined && typeof options.tokenFactory !== "function") {
			throw new TypeError("Shepherd lease token factory must be a function");
		}
		if (options.processIdentity !== undefined && !validBoundedText(options.processIdentity, 256)) {
			throw new TypeError("Shepherd process identity must be bounded text without control characters");
		}
		this.processId = processId;
		this.processIdentity = options.processIdentity;
		this.now = options.now ?? (() => new Date());
		this.isProcessAlive = options.isProcessAlive ?? defaultProcessLiveness;
		this.getProcessIdentity = options.getProcessIdentity ?? defaultProcessIdentity;
		this.tokenFactory = options.tokenFactory ?? randomUUID;
		this.testHooks = options.testHooks;
	}

	private timestamp(): string {
		const now = this.now();
		if (!(now instanceof Date) || !Number.isFinite(now.valueOf())) throw new Error("Shepherd lease clock returned an invalid Date");
		return now.toISOString();
	}

	private async prepareRoot(): Promise<RootGuard> {
		await this.assertTrustedComponents(false);
		try {
			const before = await lstat(this.root);
			if (before.isSymbolicLink()) throw new Error("Shepherd trusted root must not be a symlink");
			if (!before.isDirectory()) throw new Error("Shepherd trusted root must be a directory");
		} catch (error) {
			if ((error as NodeJS.ErrnoException).code !== "ENOENT") throw error;
			await mkdir(this.root, { recursive: true, mode: 0o700 });
		}
		await this.assertTrustedComponents(true);
		const canonical = await realpath(this.root);
		if (this.trustedRoot) {
			const trustedCanonical = await realpath(this.trustedRoot);
			const relation = relative(trustedCanonical, canonical);
			if (relation === "" || relation === ".." || relation.startsWith(`..${sep}`) || isAbsolute(relation)) {
				throw new Error("Shepherd state root escaped its trusted root");
			}
		}
		if (this.canonicalRoot !== undefined && canonical !== this.canonicalRoot) {
			throw new Error("Shepherd trusted root identity changed");
		}
		this.canonicalRoot = canonical;
		const handle = await open(canonical, constants.O_RDONLY | constants.O_DIRECTORY | constants.O_NOFOLLOW);
		try {
			const metadata = await handle.stat();
			if (!metadata.isDirectory()) throw new Error("Shepherd trusted root must be a directory");
			await handle.chmod(0o700);
			const secured = await handle.stat();
			if ((secured.mode & 0o777) !== 0o700) throw new Error("Shepherd trusted root permissions must be 0700");
			return { path: canonical, handle, device: secured.dev, inode: secured.ino };
		} catch (error) {
			await handle.close();
			throw error;
		}
	}

	private async assertTrustedComponents(requireAll: boolean): Promise<void> {
		const anchor = this.trustedRoot ?? this.root;
		const targets = [anchor];
		if (this.trustedRoot) {
			let current = anchor;
			for (const component of relative(anchor, this.root).split(sep).filter(Boolean)) {
				current = join(current, component);
				targets.push(current);
			}
		}
		for (const target of targets) {
			try {
				const metadata = await lstat(target);
				if (metadata.isSymbolicLink()) throw new Error(`Shepherd trusted root component is a symlink: ${target}`);
				if (!metadata.isDirectory()) throw new Error(`Shepherd trusted root component is not a directory: ${target}`);
				if (this.trustedRoot && target === anchor && process.platform !== "win32") {
					const currentUid = process.getuid?.();
					if (currentUid !== undefined && metadata.uid !== currentUid) {
						throw new Error("Shepherd trusted root must be owned by the current user");
					}
					if ((metadata.mode & 0o077) !== 0) throw new Error("Shepherd trusted root must not be group/world accessible");
				}
			} catch (error) {
				if (!requireAll && (error as NodeJS.ErrnoException).code === "ENOENT") break;
				throw error;
			}
		}
	}

	private async assertRootGuard(root: RootGuard): Promise<void> {
		const metadata = await lstat(root.path);
		if (metadata.isSymbolicLink() || !metadata.isDirectory() || metadata.dev !== root.device || metadata.ino !== root.inode) {
			throw new Error("Shepherd trusted root identity changed");
		}
	}

	private async closeRoot(root: RootGuard): Promise<void> {
		await root.handle.close();
	}

	private async readLeaseRecord(
		root: RootGuard,
		name: string,
		allowCrashPartial = false,
	): Promise<LeaseRecord | CorruptLeaseRecord | undefined> {
		await this.assertRootGuard(root);
		let handle;
		try {
			handle = await open(join(root.path, name), constants.O_RDONLY | constants.O_NOFOLLOW);
		} catch (error) {
			if ((error as NodeJS.ErrnoException).code === "ENOENT") return undefined;
			throw new Error("unable to open Shepherd run lease safely", { cause: error });
		}
		try {
			const file = await handle.stat();
			if (!file.isFile() || file.size > MAX_LEASE_BYTES || (file.mode & 0o777) !== 0o600) {
				throw new Error("invalid Shepherd run lease: lock must be a bounded mode-0600 regular file");
			}
			if (file.size === 0 && allowCrashPartial) return { kind: "crash-partial", device: file.dev, inode: file.ino };
			let parsed: unknown;
			try {
				parsed = JSON.parse(await readBounded(handle, MAX_LEASE_BYTES, "Shepherd run lease"));
				validateLeaseMetadata(parsed);
			} catch (error) {
				if (allowCrashPartial) return { kind: "crash-partial", device: file.dev, inode: file.ino };
				throw new Error("invalid Shepherd run lease: malformed or invalid JSON", { cause: error });
			}
			return { metadata: parsed, device: file.dev, inode: file.ino };
		} finally {
			await handle.close();
		}
	}

	private async resolveLease(root: RootGuard): Promise<LeaseRecord | CorruptLeaseRecord | undefined> {
		let current = await this.readLeaseRecord(root, "active.lock", true);
		if (!current) return undefined;
		if ("kind" in current) {
			const recovered = await this.readLeaseRecord(root, ".active.recovery");
			if (!recovered) return current;
			if ("kind" in recovered) throw new Error("invalid Shepherd recovery lease");
			current = recovered;
		}
		const seen = new Set<string>();
		for (let depth = 0; depth < MAX_LEASE_CHAIN; depth += 1) {
			const token = leaseToken(current.metadata);
			if (seen.has(token)) throw new Error("invalid Shepherd run lease: cyclic successor chain");
			seen.add(token);
			const next = await this.readLeaseRecord(root, successorName(token));
			if (!next) return current;
			if ("kind" in next) throw new Error("invalid Shepherd successor lease");
			if (isOwner(current.metadata) && !isOwner(next.metadata) && next.metadata.releasedLeaseToken !== token) {
				throw new Error("invalid Shepherd run lease: release does not match its owner");
			}
			if (!isOwner(current.metadata) && !isOwner(next.metadata)) {
				throw new Error("invalid Shepherd run lease: consecutive release records");
			}
			current = next;
		}
		throw new Error("invalid Shepherd run lease: successor chain exceeds its bound");
	}

	private async publishRecord(root: RootGuard, destinationName: string, metadata: LeaseMetadata): Promise<LeaseRecord | undefined> {
		validateLeaseMetadata(metadata);
		const payload = `${JSON.stringify(metadata)}\n`;
		if (Buffer.byteLength(payload, "utf8") > MAX_LEASE_BYTES) throw new Error("invalid Shepherd run lease: metadata is too large");
		const temporaryName = `.lease-record.${process.pid}.${randomUUID()}.tmp`;
		const temporary = join(root.path, temporaryName);
		let handle;
		try {
			await this.assertRootGuard(root);
			handle = await open(temporary, constants.O_CREAT | constants.O_EXCL | constants.O_WRONLY | constants.O_NOFOLLOW, 0o600);
			await handle.chmod(0o600);
			await handle.writeFile(payload, "utf8");
			await handle.sync();
			await handle.close();
			handle = undefined;
			try {
				await link(temporary, join(root.path, destinationName));
			} catch (error) {
				if ((error as NodeJS.ErrnoException).code === "EEXIST") return undefined;
				throw new Error("unable to publish Shepherd run lease atomically", { cause: error });
			}
			await root.handle.sync();
			const published = await this.readLeaseRecord(root, destinationName);
			if (!published || "kind" in published) throw new Error("Shepherd run lease publication was lost");
			return published;
		} finally {
			await handle?.close().catch(() => undefined);
			await rm(temporary, { force: true }).catch(() => undefined);
		}
	}

	private async createOwnerMetadata(claim: RunLeaseClaim): Promise<OwnerLeaseMetadata> {
		const token = this.tokenFactory();
		if (!validLeaseToken(token)) throw new Error("invalid Shepherd run lease: token factory returned an unsafe token");
		const detectedIdentity = this.processIdentity ?? await this.getProcessIdentity(this.processId);
		const metadata: OwnerLeaseMetadata = {
			schemaVersion: 1, issue: claim.issue, runId: claim.runId, pid: this.processId, token,
			createdAt: this.timestamp(), ownerIdentity: detectedIdentity ?? `unverified:${this.processId}`,
		};
		validateOwnerLeaseMetadata(metadata);
		return metadata;
	}

	private async ownerIsAlive(metadata: OwnerLeaseMetadata): Promise<boolean> {
		const alive = await this.isProcessAlive(metadata.pid);
		if (typeof alive !== "boolean") throw new Error("Shepherd process liveness check returned a non-boolean result");
		if (!alive) return false;
		if (!metadata.ownerIdentity || metadata.ownerIdentity.startsWith("unverified:")) return true;
		const currentIdentity = await this.getProcessIdentity(metadata.pid);
		return currentIdentity === undefined || currentIdentity === metadata.ownerIdentity;
	}

	private leaseFor(expected: LeaseRecord): RunLease {
		let released = false;
		return {
			assertOwned: async () => {
				if (released) throw new Error("Shepherd run lease was already released");
				const root = await this.prepareRoot();
				try {
					const current = await this.resolveLease(root);
					if (!current || "kind" in current || !sameLeaseRecord(current, expected)) {
						throw new Error("Shepherd run lease ownership was lost (owner token mismatch)");
					}
				} finally {
					await this.closeRoot(root);
				}
			},
			release: async () => {
				if (released) throw new Error("Shepherd run lease was already released");
				const root = await this.prepareRoot();
				try {
					const current = await this.resolveLease(root);
					if (!current || "kind" in current || !sameLeaseRecord(current, expected) || !isOwner(current.metadata)) {
						throw new Error("Shepherd run lease ownership was lost before release");
					}
					const releaseMetadata: ReleaseLeaseMetadata = {
						schemaVersion: 1, recordType: "released", releasedLeaseToken: current.metadata.token,
						token: randomUUID(), createdAt: this.timestamp(),
					};
					const publication = await this.publishRecord(root, successorName(current.metadata.token), releaseMetadata);
					if (!publication) throw new Error("Shepherd run lease ownership was lost during release");
					released = true;
				} finally {
					await this.closeRoot(root);
				}
			},
		};
	}

	async acquireLease(claim: RunLeaseClaim): Promise<RunLease> {
		validateLeaseClaim(claim);
		const metadata = await this.createOwnerMetadata(claim);
		const root = await this.prepareRoot();
		try {
			for (let attempt = 0; attempt < 32; attempt += 1) {
				const current = await this.resolveLease(root);
				if (!current) {
					const acquired = await this.publishRecord(root, "active.lock", metadata);
					if (acquired) return this.leaseFor(acquired);
					continue;
				}
				if ("kind" in current) {
					if (claim.mode !== "resume") throw new Error("Shepherd run lease has an incomplete crash publication; use resume to recover it");
					const acquired = await this.publishRecord(root, ".active.recovery", metadata);
					if (acquired) return this.leaseFor(acquired);
					continue;
				}
				if (!isOwner(current.metadata)) {
					const acquired = await this.publishRecord(root, successorName(current.metadata.token), metadata);
					if (acquired) return this.leaseFor(acquired);
					continue;
				}
				if (await this.ownerIsAlive(current.metadata)) {
					throw new Error(
						`Shepherd run lease is held by live process ${current.metadata.pid} for issue #${current.metadata.issue} (run ${current.metadata.runId})`,
					);
				}
				if (claim.mode !== "resume") {
					throw new Error(
						`Shepherd run lease is stale for issue #${current.metadata.issue} (dead process ${current.metadata.pid}); use resume to recover it`,
					);
				}
				if (current.metadata.issue !== claim.issue) {
					throw new Error(`Shepherd run lease is stale for issue #${current.metadata.issue}; resume that issue before issue #${claim.issue}`);
				}
				const acquired = await this.publishRecord(root, successorName(current.metadata.token), metadata);
				if (acquired) return this.leaseFor(acquired);
			}
			throw new Error("unable to acquire Shepherd run lease after repeated concurrent changes");
		} finally {
			await this.closeRoot(root);
		}
	}

	async load(issue: number): Promise<ShepherdRunState | undefined> {
		if (!validIssue(issue)) throw new RangeError("issue must be a positive bounded integer");
		const root = await this.prepareRoot();
		let handle;
		try {
			await this.assertRootGuard(root);
			try {
				handle = await open(join(root.path, `issue-${issue}.json`), constants.O_RDONLY | constants.O_NOFOLLOW);
			} catch (error) {
				if ((error as NodeJS.ErrnoException).code === "ENOENT") return undefined;
				throw new Error("unable to open Shepherd state safely", { cause: error });
			}
			await this.testHooks?.afterStateOpen?.();
			const metadata = await handle.stat();
			if (!metadata.isFile() || metadata.size < 1 || metadata.size > MAX_STATE_BYTES || (metadata.mode & 0o777) !== 0o600) {
				throw new Error("invalid Shepherd state: state file must be a bounded mode-0600 regular file");
			}
			let parsed: unknown;
			try {
				parsed = JSON.parse(await readBounded(handle, MAX_STATE_BYTES, "Shepherd state"));
			} catch (error) {
				throw new Error("invalid Shepherd state: malformed JSON", { cause: error });
			}
			validateState(parsed, issue);
			return parsed;
		} finally {
			await handle?.close().catch(() => undefined);
			await this.closeRoot(root);
		}
	}

	async save(state: ShepherdRunState): Promise<void> {
		if (!validIssue(state?.issue)) throw new Error("invalid Shepherd state: invalid issue");
		validateState(state, state.issue, true);
		const persisted = projectState(state);
		validateState(persisted, persisted.issue);
		const payload = `${JSON.stringify(persisted, null, 2)}\n`;
		if (Buffer.byteLength(payload, "utf8") > MAX_STATE_BYTES) throw new Error("invalid Shepherd state: serialized state is too large");
		const root = await this.prepareRoot();
		const temporary = join(root.path, `.issue-${state.issue}.${process.pid}.${randomUUID()}.tmp`);
		let handle;
		try {
			await this.assertRootGuard(root);
			handle = await open(temporary, constants.O_CREAT | constants.O_EXCL | constants.O_WRONLY | constants.O_NOFOLLOW, 0o600);
			await handle.chmod(0o600);
			await handle.writeFile(payload, "utf8");
			await handle.sync();
			await handle.close();
			handle = undefined;
			await this.assertRootGuard(root);
			try {
				const existing = await lstat(join(root.path, `issue-${state.issue}.json`));
				if (existing.isSymbolicLink() || !existing.isFile()) {
					throw new Error("invalid Shepherd state: destination must not be a symlink or non-file");
				}
			} catch (error) {
				if ((error as NodeJS.ErrnoException).code !== "ENOENT") throw error;
			}
			await rename(temporary, join(root.path, `issue-${state.issue}.json`));
			await root.handle.sync();
			await this.assertRootGuard(root);
		} finally {
			await handle?.close().catch(() => undefined);
			await rm(temporary, { force: true }).catch(() => undefined);
			await this.closeRoot(root);
		}
	}
}
