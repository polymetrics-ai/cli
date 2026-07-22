import { createHash } from "node:crypto";
import { lstat, readFile, realpath } from "node:fs/promises";
import { isAbsolute, relative, resolve, sep } from "node:path";

import {
	assertProductionVerificationRecipe,
	ProductionLifecycleError,
	type ProductionVerificationCommand,
} from "./autonomous-production-contract.ts";
import type { RoleRunRequest } from "./agent-session-runtime.ts";
import type { ProductionVerificationResult } from "./bounded-verification.ts";
import type { ProductionAgentSessionPort } from "./production-workspace-lifecycle.ts";
import {
	validateScopedPath,
	type CapabilityResult,
	type HostCapability,
	type ScopedWorkspace,
} from "./tool-policy.ts";

const SHA = /^[0-9a-f]{40}$/u;
const SAFE_ID = /^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$/u;
const SAFE_BRANCH = /^(?!\/|.*(?:\.\.|\s|[~^:?*\\\[\]])|.*\/$)[A-Za-z0-9][A-Za-z0-9._\/-]{0,239}$/u;
const MAX_READ_BYTES = 256 * 1024;

export interface ProductionVerificationCommandExecutor {
	run(
		worktreeRoot: string,
		command: ProductionVerificationCommand,
		signal?: AbortSignal,
	): Promise<ProductionVerificationResult>;
}

export interface ProductionVerificationSessionBinding {
	issue: number;
	branch: string;
	runId: string;
	generation: number;
	laneId: string;
	candidateHead: string;
}

export interface AgentSessionVerificationRunnerOptions {
	agentSession: ProductionAgentSessionPort;
	executor: ProductionVerificationCommandExecutor;
	timeoutMs?: number;
}

export class AgentSessionVerificationRunner {
	readonly #agentSession: ProductionAgentSessionPort;
	readonly #executor: ProductionVerificationCommandExecutor;
	readonly #timeoutMs: number;

	constructor(options: AgentSessionVerificationRunnerOptions) {
		if (typeof options !== "object" || options === null || typeof options.agentSession?.run !== "function"
			|| typeof options.agentSession.abort !== "function" || typeof options.executor?.run !== "function") {
			throw new Error("AgentSession verification options are invalid");
		}
		this.#agentSession = options.agentSession;
		this.#executor = options.executor;
		this.#timeoutMs = options.timeoutMs ?? 15 * 60 * 1_000;
		if (!Number.isSafeInteger(this.#timeoutMs) || this.#timeoutMs < 1 || this.#timeoutMs > 60 * 60 * 1_000) {
			throw new Error("AgentSession verification timeout is invalid");
		}
	}

	async createAgentCapability(
		worktreeRoot: string,
		commands: readonly ProductionVerificationCommand[],
		signal?: AbortSignal,
	): Promise<HostCapability> {
		if (typeof worktreeRoot !== "string" || !isAbsolute(worktreeRoot)
			|| !Array.isArray(commands) || commands.length < 1 || commands.length > 64
			|| (signal !== undefined && !(signal instanceof AbortSignal))) {
			throw new Error("interactive AgentSession verification request is invalid");
		}
		const sessionSignal = signal ?? new AbortController().signal;
		if (sessionSignal.aborted) throw sessionSignal.reason ?? new Error("interactive verification cancelled");
		const canonicalRoot = await realpath(worktreeRoot);
		for (const command of commands) assertProductionVerificationRecipe(command);
		const ids = commands.map((command) => command.id);
		if (new Set(ids).size !== ids.length || ids.some((id) => !SAFE_ID.test(id))) {
			throw new Error("verification command IDs must be unique and safe");
		}
		return {
			name: "host_verify",
			description: "Run one immutable plan verification tuple by ID and return its bounded diagnostic.",
			mutates: true,
			parameters: {
				type: "object",
				additionalProperties: false,
				required: ["id"],
				properties: { id: { type: "string", enum: ids } },
			},
			execute: async (input, operationSignal): Promise<CapabilityResult> => {
				const keys = Object.keys(input);
				const id = input.id;
				if (keys.length !== 1 || keys[0] !== "id" || typeof id !== "string") {
					return { status: "blocked", summary: "host_verify accepts exactly one immutable command ID" };
				}
				const command = commands.find((candidate) => candidate.id === id);
				if (command === undefined) {
					return { status: "blocked", summary: "host_verify rejected an undeclared command ID" };
				}
				const activeSignal = operationSignal ?? sessionSignal;
				if (activeSignal.aborted) throw activeSignal.reason ?? new Error("interactive verification cancelled");
				const result = await this.#executor.run(canonicalRoot, command, activeSignal);
				assertExecutionResult(result, command);
				return {
					status: result.status === "passed" ? "ok" : "failed",
					summary: verificationResultSummary(result),
					references: [verificationResultReference(result)],
				};
			},
		};
	}

	async runAll(
		worktreeRoot: string,
		commands: readonly ProductionVerificationCommand[],
		signal?: AbortSignal,
		bindingValue?: ProductionVerificationSessionBinding,
	): Promise<ProductionVerificationResult[]> {
		if (typeof worktreeRoot !== "string" || !isAbsolute(worktreeRoot)
			|| !Array.isArray(commands) || commands.length < 1 || commands.length > 64
			|| (signal !== undefined && !(signal instanceof AbortSignal))) {
			throw new Error("AgentSession verification request is invalid");
		}
		const sessionSignal = signal ?? new AbortController().signal;
		if (sessionSignal.aborted) throw sessionSignal.reason ?? new Error("AgentSession verification cancelled");
		const binding = validateBinding(bindingValue);
		const canonicalRoot = await realpath(worktreeRoot);
		for (const command of commands) assertProductionVerificationRecipe(command);
		const ids = commands.map((command) => command.id);
		if (new Set(ids).size !== ids.length || ids.some((id) => !SAFE_ID.test(id))) {
			throw new Error("verification command IDs must be unique and safe");
		}
		const results: ProductionVerificationResult[] = [];
		let next = 0;
		let protocolViolation = false;
		const capability: HostCapability = {
			name: "host_verify",
			description: "Run the next immutable verification command selected only by its declared ID.",
			mutates: true,
			parameters: {
				type: "object",
				additionalProperties: false,
				required: ["id"],
				properties: { id: { type: "string", enum: ids } },
			},
			execute: async (input, operationSignal): Promise<CapabilityResult> => {
				const keys = Object.keys(input);
				const id = input.id;
				if (keys.length !== 1 || keys[0] !== "id" || typeof id !== "string") {
					protocolViolation = true;
					return { status: "blocked", summary: "host_verify accepts only one immutable command ID" };
				}
				const planned = commands[next];
				if (planned === undefined || id !== planned.id || results.some((result) => result.status !== "passed")) {
					protocolViolation = true;
					return { status: "blocked", summary: "verification command ID was omitted, repeated, or out of order" };
				}
				const activeSignal = operationSignal ?? sessionSignal;
				if (activeSignal.aborted) throw activeSignal.reason ?? new Error("host verification cancelled");
				const result = await this.#executor.run(canonicalRoot, planned, activeSignal);
				assertExecutionResult(result, planned);
				results.push(structuredClone(result));
				next += 1;
				return {
					status: result.status === "passed" ? "ok" : "failed",
					summary: verificationResultSummary(result),
					references: [verificationResultReference(result)],
				};
			},
		};
		const workspaceId = `verification-${binding.issue}-${createHash("sha256")
			.update(`${canonicalRoot}\0${binding.laneId}`)
			.digest("hex")
			.slice(0, 12)}`;
		const request: RoleRunRequest = {
			role: "verification",
			task: "Call host_verify once for every declared verification ID in the listed order. Stop after the first host failure, then return the typed handoff.",
			context: [
				`Immutable verification order: ${ids.join(", ")}.`,
				"The host owns executable, argv, cwd, environment, timeout, output bounds, and pass/fail truth.",
				"Repository tests execute with explicitly accepted trusted-local user authority; no sandbox is claimed.",
			],
			timeoutMs: this.#timeoutMs,
			signal: sessionSignal,
			workspace: verificationWorkspace(canonicalRoot, workspaceId),
			workspaceMutation: false,
			capabilities: [capability],
			authority: {
				issue: binding.issue,
				branch: binding.branch,
				readOnly: false,
				workspaceId,
				readPrefixes: ["."],
				writePrefixes: [`.shepherd-verification/${binding.laneId}`],
				capabilityNames: ["host_verify"],
			},
			binding: {
				runId: binding.runId,
				generation: binding.generation,
				laneId: binding.laneId,
				candidateHead: binding.candidateHead,
				validationNonce: createHash("sha256")
					.update(`${binding.runId}\0${binding.generation}\0${binding.laneId}\0${binding.candidateHead}\0${ids.join("\0")}`)
					.digest("hex"),
			},
		};
		const handoff = await this.#agentSession.run(request);
		if (sessionSignal.aborted) throw sessionSignal.reason ?? new Error("AgentSession verification cancelled");
		const expectedCount = results.some((result) => result.status === "failed")
			? results.length
			: commands.length;
		if (protocolViolation || results.length !== expectedCount
			|| (results.every((result) => result.status === "passed") && results.length !== commands.length)
			|| handoff.status !== "completed" || handoff.role !== "verification"
			|| handoff.observedMutation !== true || handoff.changedPaths.length !== 0
			|| handoff.runId !== request.binding.runId || handoff.generation !== request.binding.generation
			|| handoff.laneId !== request.binding.laneId || handoff.candidateHead !== request.binding.candidateHead
			|| handoff.validationNonce !== request.binding.validationNonce) {
			throw new ProductionLifecycleError(
				"correction_required",
				"verification AgentSession did not run every required command in immutable order",
				["verification_agent_protocol"],
			);
		}
		return results;
	}
}

function assertExecutionResult(
	result: ProductionVerificationResult,
	command: ProductionVerificationCommand,
): void {
	if (result.id !== command.id || (result.status !== "passed" && result.status !== "failed")
		|| (result.status === "failed" && result.failureKind === undefined)) {
		throw new Error("host verification returned malformed authoritative evidence");
	}
}

function verificationResultReference(result: ProductionVerificationResult): string {
	return createHash("sha256").update(JSON.stringify({
		id: result.id,
		status: result.status,
		exitCode: result.exitCode,
		signal: result.signal,
		failureKind: result.failureKind,
	})).digest("hex");
}

function verificationResultSummary(result: ProductionVerificationResult): string {
	if (result.status === "passed") return `verification ${result.id} passed`;
	const source = result.stderr.trim().length > 0 ? result.stderr : result.stdout;
	const excerpt = source
		.slice(0, 2_048)
		.replace(/[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f-\u009f]/gu, "�")
		.trim();
	return `verification ${result.id} failed (${result.failureKind})${excerpt.length === 0 ? "" : `\n${excerpt}`}`;
}

function validateBinding(value: unknown): ProductionVerificationSessionBinding {
	if (typeof value !== "object" || value === null || Array.isArray(value)) {
		throw new Error("verification AgentSession binding is required");
	}
	const candidate = value as Record<string, unknown>;
	if (JSON.stringify(Object.keys(candidate).sort()) !== JSON.stringify([
		"branch", "candidateHead", "generation", "issue", "laneId", "runId",
	])
		|| !Number.isSafeInteger(candidate.issue) || (candidate.issue as number) < 1
		|| !Number.isSafeInteger(candidate.generation) || (candidate.generation as number) < 1
		|| typeof candidate.branch !== "string" || !SAFE_BRANCH.test(candidate.branch)
		|| typeof candidate.runId !== "string" || !SAFE_ID.test(candidate.runId)
		|| typeof candidate.laneId !== "string" || !SAFE_ID.test(candidate.laneId)
		|| typeof candidate.candidateHead !== "string" || !SHA.test(candidate.candidateHead)) {
		throw new Error("verification AgentSession binding is invalid");
	}
	return candidate as unknown as ProductionVerificationSessionBinding;
}

function verificationWorkspace(root: string, id: string): ScopedWorkspace {
	return Object.freeze({
		id,
		cwd: root,
		async readText(
			path: string,
			options: { offset?: number; limit?: number; signal?: AbortSignal },
		) {
			if (options.signal?.aborted) throw options.signal.reason ?? new Error("verification read cancelled");
			const normalized = validateScopedPath(path, ["."]);
			const offset = options.offset ?? 0;
			const limit = options.limit ?? MAX_READ_BYTES;
			if (!Number.isSafeInteger(offset) || offset < 0 || !Number.isSafeInteger(limit)
				|| limit < 1 || limit > MAX_READ_BYTES) throw new Error("verification read range is invalid");
			const target = resolve(root, normalized);
			const back = relative(root, target);
			if (back === ".." || back.startsWith(`..${sep}`)) throw new Error("verification read escapes worktree");
			let current = root;
			for (const component of back.split(sep)) {
				if (component.length === 0) continue;
				current = resolve(current, component);
				const metadata = await lstat(current);
				if (metadata.isSymbolicLink()) throw new Error("verification read cannot traverse symlinks");
			}
			const metadata = await lstat(target);
			if (!metadata.isFile() || metadata.size > MAX_READ_BYTES) throw new Error("verification read requires a bounded file");
			const value = await readFile(target, "utf8");
			return value.slice(offset, offset + limit);
		},
		async editText() { throw new Error("verification workspace mutation is unavailable"); },
		async writeText() { throw new Error("verification workspace mutation is unavailable"); },
	});
}
