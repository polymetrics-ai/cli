import { randomUUID } from "node:crypto";
import { chmod, mkdir, open, readFile, rename, rm, stat } from "node:fs/promises";
import { isAbsolute, join } from "node:path";

import type { ShepherdLaneState, ShepherdRunState } from "./domain.ts";

const MAX_STATE_BYTES = 1_048_576;
const DEFAULT_SUMMARY_LENGTH = 2_048;
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

export class FileStateStore {
	readonly root: string;

	constructor(root: string) {
		if (typeof root !== "string" || !isAbsolute(root) || /[\u0000-\u001f\u007f]/.test(root)) {
			throw new TypeError("Shepherd state root must be an absolute path without control characters");
		}
		this.root = root;
	}

	private pathFor(issue: number): string {
		if (!validIssue(issue)) throw new RangeError("issue must be a positive bounded integer");
		return join(this.root, `issue-${issue}.json`);
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
