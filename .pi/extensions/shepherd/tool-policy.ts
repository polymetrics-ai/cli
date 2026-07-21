import { posix } from "node:path";

const DEFAULT_MAX_TOOL_OUTPUT_BYTES = 64 * 1024;
const DEFAULT_MAX_READ_CHARACTERS = 64 * 1024;
const DEFAULT_MAX_WRITE_CHARACTERS = 256 * 1024;
const MAX_PATH_CHARACTERS = 512;
const MAX_CAPABILITIES = 32;
const MAX_REFERENCES = 32;
const MAX_REFERENCE_CHARACTERS = 512;
const MAX_CAPABILITY_SUMMARY_CHARACTERS = 4 * 1024;
const MAX_CAPABILITY_SCHEMA_BYTES = 32 * 1024;

const forbiddenCapabilityPatterns = [
	/(?:^|_)(?:bash|shell|exec|command)(?:_|$)/,
	/(?:^|_)(?:subagent|spawn_agent|delegate|orchestrat(?:e|ion)|agent_create)(?:_|$)/,
	/(?:^|_)http_(?:request|write|post|put|patch|delete)(?:_|$)/,
	/(?:^|_)sql_(?:request|query|write|insert|update|delete|execute)(?:_|$)/,
	/(?:^|_)(?:secret|credential|token|password|auth)_(?:read|get|list|dump|export)(?:_|$)/,
] as const;

const sensitivePathPatterns = [
	/(?:^|\/)\.git(?:\/|$)/i,
	/(?:^|\/)\.env(?:\.|$)/i,
	/(?:^|\/)(?:credentials?|secrets?|auth)(?:[._-]|$)/i,
	/(?:^|\/)(?:id_rsa|id_ed25519|known_hosts)(?:\.|$)/i,
	/\.(?:pem|p12|pfx|key)$/i,
] as const;

export interface WorkspaceMutationResult {
	changed: boolean;
	summary: string;
}

/** Opaque, already-owned workspace capability. This policy never creates or removes workspaces. */
export interface ScopedWorkspace {
	readonly id: string;
	readonly cwd: string;
	readText(path: string, options: { offset?: number; limit?: number; signal?: AbortSignal }): Promise<string>;
	editText(path: string, oldText: string, newText: string, signal?: AbortSignal): Promise<WorkspaceMutationResult>;
	writeText(path: string, content: string, signal?: AbortSignal): Promise<WorkspaceMutationResult>;
}

export interface CapabilityResult {
	status: "ok" | "blocked" | "failed";
	summary: string;
	references?: string[];
}

/** A narrow host action. Generic transports and orchestration capabilities are rejected by name. */
export interface HostCapability {
	readonly name: string;
	readonly description: string;
	readonly mutates: boolean;
	readonly parameters: Readonly<Record<string, unknown>>;
	execute(input: Readonly<Record<string, unknown>>, signal?: AbortSignal): Promise<CapabilityResult>;
}

export interface ToolAuthority {
	workspaceId: string;
	readPrefixes: string[];
	writePrefixes: string[];
	capabilityNames: string[];
}

export interface SessionToolResult {
	content: Array<{ type: "text"; text: string }>;
	details?: unknown;
}

/** Structural subset of Pi 0.80.6 ToolDefinition used by injected AgentSession factories. */
export interface SessionTool {
	name: string;
	label: string;
	description: string;
	promptSnippet?: string;
	promptGuidelines?: string[];
	parameters: Readonly<Record<string, unknown>>;
	executionMode?: "sequential" | "parallel";
	execute(
		toolCallId: string,
		params: Readonly<Record<string, unknown>>,
		signal: AbortSignal | undefined,
		onUpdate?: unknown,
		context?: unknown,
	): Promise<SessionToolResult>;
}

export interface ToolPolicy {
	readonly names: string[];
	readonly tools: SessionTool[];
}

export interface ToolPolicyInput {
	readOnly: boolean;
	workspace: ScopedWorkspace;
	authority: ToolAuthority;
	capabilities: HostCapability[];
}

export interface ToolPolicyOptions {
	maxToolOutputBytes?: number;
	maxReadCharacters?: number;
	maxWriteCharacters?: number;
}

export class ToolPolicyError extends Error {
	constructor(message: string, options?: ErrorOptions) {
		super(message, options);
		this.name = "ToolPolicyError";
	}
}

export function createToolPolicy(input: ToolPolicyInput, options: ToolPolicyOptions = {}): ToolPolicy {
	const limits = {
		maxToolOutputBytes: positiveInteger(options.maxToolOutputBytes ?? DEFAULT_MAX_TOOL_OUTPUT_BYTES, "maxToolOutputBytes"),
		maxReadCharacters: positiveInteger(options.maxReadCharacters ?? DEFAULT_MAX_READ_CHARACTERS, "maxReadCharacters"),
		maxWriteCharacters: positiveInteger(options.maxWriteCharacters ?? DEFAULT_MAX_WRITE_CHARACTERS, "maxWriteCharacters"),
	};
	validatePolicyInput(input);

	const readPrefixes = normalizePrefixes(input.authority.readPrefixes, "read");
	const writePrefixes = input.readOnly && input.authority.writePrefixes.length === 0
		? []
		: normalizePrefixes(input.authority.writePrefixes, "write");
	const tools: SessionTool[] = [workspaceReadTool(input.workspace, readPrefixes, limits)];
	if (!input.readOnly) {
		tools.push(
			workspaceEditTool(input.workspace, writePrefixes, limits),
			workspaceWriteTool(input.workspace, writePrefixes, limits),
		);
	}

	const declared = new Set(input.authority.capabilityNames);
	for (const capability of input.capabilities) {
		if (!declared.has(capability.name)) {
			throw new ToolPolicyError(`undeclared capability ${JSON.stringify(capability.name)} cannot expand authority`);
		}
		if (input.readOnly && capability.mutates) continue;
		tools.push(hostCapabilityTool(capability, limits));
	}

	return {
		names: tools.map((tool) => tool.name),
		tools,
	};
}

function validatePolicyInput(input: ToolPolicyInput): void {
	if (!input || typeof input !== "object") throw new ToolPolicyError("tool policy input is required");
	if (typeof input.readOnly !== "boolean") throw new ToolPolicyError("readOnly must be boolean");
	if (!input.workspace || typeof input.workspace !== "object") throw new ToolPolicyError("workspace capability is required");
	if (!validIdentifier(input.workspace.id) || input.workspace.id !== input.authority?.workspaceId) {
		throw new ToolPolicyError("workspace identity does not match the immutable authority envelope");
	}
	for (const method of ["readText", "editText", "writeText"] as const) {
		if (typeof input.workspace[method] !== "function") {
			throw new ToolPolicyError(`workspace capability is missing ${method}`);
		}
	}
	if (!Array.isArray(input.capabilities) || input.capabilities.length > MAX_CAPABILITIES) {
		throw new ToolPolicyError("typed host capabilities must be a bounded array");
	}
	if (!Array.isArray(input.authority.capabilityNames) || input.authority.capabilityNames.length > MAX_CAPABILITIES) {
		throw new ToolPolicyError("capability authority must be a bounded array");
	}

	const declared = new Set<string>();
	for (const name of input.authority.capabilityNames) {
		validateCapabilityName(name);
		if (declared.has(name)) throw new ToolPolicyError(`duplicate declared capability ${name}`);
		declared.add(name);
	}
	const supplied = new Set<string>();
	for (const capability of input.capabilities) {
		if (!capability || typeof capability !== "object") throw new ToolPolicyError("capability must be an object");
		validateCapabilityName(capability.name);
		if (supplied.has(capability.name)) throw new ToolPolicyError(`duplicate supplied capability ${capability.name}`);
		supplied.add(capability.name);
		if (!declared.has(capability.name)) {
			throw new ToolPolicyError(`undeclared capability ${capability.name} cannot expand authority`);
		}
		if (typeof capability.description !== "string" || capability.description.length < 1 || capability.description.length > 512 ||
			/[\u0000-\u001f\u007f]/.test(capability.description)) {
			throw new ToolPolicyError(`capability ${capability.name} has an invalid description`);
		}
		if (typeof capability.mutates !== "boolean" || typeof capability.execute !== "function") {
			throw new ToolPolicyError(`capability ${capability.name} has an invalid typed contract`);
		}
		if (!isRecord(capability.parameters)) {
			throw new ToolPolicyError(`capability ${capability.name} requires a bounded parameter schema`);
		}
		if (capability.parameters.type !== "object" || capability.parameters.additionalProperties !== false) {
			throw new ToolPolicyError(`capability ${capability.name} parameter schema must be a closed object`);
		}
		try {
			if (byteLength(JSON.stringify(capability.parameters)) > MAX_CAPABILITY_SCHEMA_BYTES) {
				throw new ToolPolicyError(`capability ${capability.name} parameter schema exceeded its bound`);
			}
		} catch (error) {
			if (error instanceof ToolPolicyError) throw error;
			throw new ToolPolicyError(`capability ${capability.name} parameter schema is not serializable`, { cause: error });
		}
	}
	for (const name of declared) {
		if (!supplied.has(name)) throw new ToolPolicyError(`declared capability ${name} was not supplied`);
	}
}

function workspaceReadTool(
	workspace: ScopedWorkspace,
	prefixes: string[],
	limits: Required<ToolPolicyOptions>,
): SessionTool {
	return {
		name: "workspace_read",
		label: "Read scoped workspace file",
		description: "Read bounded text from an allowlisted path in the assigned workspace.",
		promptSnippet: "Read an allowlisted file from the assigned workspace",
		parameters: closedObject({
			path: { type: "string", minLength: 1, maxLength: MAX_PATH_CHARACTERS },
			offset: { type: "integer", minimum: 0 },
			limit: { type: "integer", minimum: 1, maximum: limits.maxReadCharacters },
		}, ["path"]),
		executionMode: "parallel",
		async execute(_callId, params, signal) {
			assertSignal(signal);
			assertOnlyKeys(params, ["path", "offset", "limit"], "workspace_read");
			const path = validateScopedPath(requiredString(params.path, "path"), prefixes);
			const offset = optionalBoundedInteger(params.offset, "offset", 0, Number.MAX_SAFE_INTEGER);
			const limit = optionalBoundedInteger(params.limit, "limit", 1, limits.maxReadCharacters);
			const value = await workspace.readText(path, { offset, limit, signal });
			if (typeof value !== "string") throw new ToolPolicyError("workspace read returned non-text data");
			if (value.length > limits.maxReadCharacters) throw new ToolPolicyError("workspace read exceeded the bounded output limit");
			return textResult(redactSensitiveText(value), limits.maxToolOutputBytes);
		},
	};
}

function workspaceEditTool(
	workspace: ScopedWorkspace,
	prefixes: string[],
	limits: Required<ToolPolicyOptions>,
): SessionTool {
	return {
		name: "workspace_edit",
		label: "Edit scoped workspace file",
		description: "Replace bounded exact text in an allowlisted file in the assigned workspace.",
		promptSnippet: "Edit an allowlisted file in the assigned workspace",
		parameters: closedObject({
			path: { type: "string", minLength: 1, maxLength: MAX_PATH_CHARACTERS },
			oldText: { type: "string", minLength: 1, maxLength: limits.maxWriteCharacters },
			newText: { type: "string", maxLength: limits.maxWriteCharacters },
		}, ["path", "oldText", "newText"]),
		executionMode: "sequential",
		async execute(_callId, params, signal) {
			assertSignal(signal);
			assertOnlyKeys(params, ["path", "oldText", "newText"], "workspace_edit");
			const path = validateScopedPath(requiredString(params.path, "path"), prefixes);
			const oldText = boundedString(params.oldText, "oldText", limits.maxWriteCharacters, false);
			const newText = boundedString(params.newText, "newText", limits.maxWriteCharacters, true);
			const result = await workspace.editText(path, oldText, newText, signal);
			return mutationResult(result, limits.maxToolOutputBytes);
		},
	};
}

function workspaceWriteTool(
	workspace: ScopedWorkspace,
	prefixes: string[],
	limits: Required<ToolPolicyOptions>,
): SessionTool {
	return {
		name: "workspace_write",
		label: "Write scoped workspace file",
		description: "Write bounded text to an allowlisted file in the assigned workspace.",
		promptSnippet: "Write an allowlisted file in the assigned workspace",
		parameters: closedObject({
			path: { type: "string", minLength: 1, maxLength: MAX_PATH_CHARACTERS },
			content: { type: "string", maxLength: limits.maxWriteCharacters },
		}, ["path", "content"]),
		executionMode: "sequential",
		async execute(_callId, params, signal) {
			assertSignal(signal);
			assertOnlyKeys(params, ["path", "content"], "workspace_write");
			const path = validateScopedPath(requiredString(params.path, "path"), prefixes);
			const content = boundedString(params.content, "content", limits.maxWriteCharacters, true);
			const result = await workspace.writeText(path, content, signal);
			return mutationResult(result, limits.maxToolOutputBytes);
		},
	};
}

function hostCapabilityTool(
	capability: HostCapability,
	limits: Required<ToolPolicyOptions>,
): SessionTool {
	return {
		name: capability.name,
		label: capability.description,
		description: capability.description,
		parameters: capability.parameters,
		executionMode: capability.mutates ? "sequential" : "parallel",
		async execute(_callId, params, signal) {
			assertSignal(signal);
			if (!isRecord(params)) throw new ToolPolicyError(`${capability.name} input must be an object`);
			const inputSize = byteLength(JSON.stringify(params));
			if (inputSize > limits.maxWriteCharacters) throw new ToolPolicyError(`${capability.name} input exceeded its bound`);
			const result = await capability.execute(params, signal);
			validateCapabilityResult(capability.name, result);
			return textResult(JSON.stringify({
				status: result.status,
				summary: redactSensitiveText(result.summary),
				references: result.references?.map(redactSensitiveText) ?? [],
			}), limits.maxToolOutputBytes);
		},
	};
}

export function validateScopedPath(path: string, prefixes: readonly string[]): string {
	if (typeof path !== "string" || path.length < 1 || path.length > MAX_PATH_CHARACTERS) {
		throw new ToolPolicyError("workspace path is empty or exceeds its bound");
	}
	if (/[\u0000-\u001f\u007f\\]/.test(path) || path.startsWith("/") || /^[A-Za-z]:/.test(path)) {
		throw new ToolPolicyError("workspace path must be a portable relative path without control characters");
	}
	const normalized = posix.normalize(path);
	if (normalized === ".." || normalized.startsWith("../") || normalized === ".") {
		throw new ToolPolicyError("workspace path traversal or workspace-root access is forbidden");
	}
	if (sensitivePathPatterns.some((pattern) => pattern.test(normalized))) {
		throw new ToolPolicyError("sensitive workspace paths are outside agent authority");
	}
	if (!prefixes.some((prefix) => prefix === "." || normalized === prefix || normalized.startsWith(`${prefix}/`))) {
		throw new ToolPolicyError(`workspace path ${JSON.stringify(normalized)} is outside the declared scope`);
	}
	return normalized;
}

export function redactSensitiveText(value: string): string {
	if (typeof value !== "string") return "[REDACTED]";
	return value
		.replace(/\b(authorization\s*:\s*bearer\s+)[^\s"']+/gi, "$1[REDACTED]")
		.replace(/\b(api[_-]?key|access[_-]?token|refresh[_-]?token|token|password|passwd|secret)\b(\s*[:=]\s*)[^\s,;"']+/gi, "$1$2[REDACTED]")
		.replace(/-----BEGIN [^-]+ PRIVATE KEY-----[\s\S]*?-----END [^-]+ PRIVATE KEY-----/gi, "[REDACTED PRIVATE KEY]");
}

function validateCapabilityName(name: string): void {
	if (typeof name !== "string" || !/^host_[a-z][a-z0-9_]{1,63}$/.test(name)) {
		throw new ToolPolicyError(`capability name ${JSON.stringify(name)} must use the bounded host_ namespace`);
	}
	if (forbiddenCapabilityPatterns.some((pattern) => pattern.test(name))) {
		throw new ToolPolicyError(`capability ${name} requests forbidden generic, secret, or recursive authority`);
	}
}

function normalizePrefixes(prefixes: unknown, description: string): string[] {
	if (!Array.isArray(prefixes) || prefixes.length < 1 || prefixes.length > 64) {
		throw new ToolPolicyError(`${description} prefixes must be a non-empty bounded array`);
	}
	const normalized: string[] = [];
	for (const prefix of prefixes) {
		if (prefix === ".") {
			normalized.push(prefix);
			continue;
		}
		if (typeof prefix !== "string" || prefix.length < 1 || prefix.length > MAX_PATH_CHARACTERS ||
			/[\u0000-\u001f\u007f\\]/.test(prefix) || prefix.startsWith("/") || prefix.includes("..")) {
			throw new ToolPolicyError(`${description} prefix is invalid`);
		}
		const value = posix.normalize(prefix).replace(/\/$/, "");
		if (value === "." || sensitivePathPatterns.some((pattern) => pattern.test(value))) {
			throw new ToolPolicyError(`${description} prefix grants sensitive or workspace-root authority`);
		}
		normalized.push(value);
	}
	if (new Set(normalized).size !== normalized.length) throw new ToolPolicyError(`duplicate ${description} prefix`);
	return normalized;
}

function validateCapabilityResult(name: string, result: CapabilityResult): void {
	if (!result || typeof result !== "object" || !["ok", "blocked", "failed"].includes(result.status)) {
		throw new ToolPolicyError(`${name} returned an invalid status`);
	}
	boundedString(result.summary, `${name} summary`, MAX_CAPABILITY_SUMMARY_CHARACTERS, false);
	if (result.references !== undefined) {
		if (!Array.isArray(result.references) || result.references.length > MAX_REFERENCES) {
			throw new ToolPolicyError(`${name} returned too many references`);
		}
		for (const reference of result.references) {
			boundedString(reference, `${name} reference`, MAX_REFERENCE_CHARACTERS, false);
		}
	}
}

function mutationResult(result: WorkspaceMutationResult, maxBytes: number): SessionToolResult {
	if (!result || typeof result.changed !== "boolean") throw new ToolPolicyError("workspace mutation returned an invalid result");
	const summary = boundedString(result.summary, "workspace mutation summary", MAX_CAPABILITY_SUMMARY_CHARACTERS, false);
	return textResult(JSON.stringify({ changed: result.changed, summary: redactSensitiveText(summary) }), maxBytes);
}

function textResult(value: string, maxBytes: number): SessionToolResult {
	if (byteLength(value) > maxBytes) throw new ToolPolicyError("tool output exceeded the bounded output limit");
	return { content: [{ type: "text", text: value }] };
}

function closedObject(properties: Record<string, unknown>, required: string[]): Readonly<Record<string, unknown>> {
	return { type: "object", additionalProperties: false, properties, required };
}

function assertOnlyKeys(value: Readonly<Record<string, unknown>>, allowed: readonly string[], description: string): void {
	if (!isRecord(value)) throw new ToolPolicyError(`${description} input must be an object`);
	const allowedSet = new Set(allowed);
	for (const key of Object.keys(value)) {
		if (!allowedSet.has(key)) throw new ToolPolicyError(`${description} input contains unknown field ${JSON.stringify(key)}`);
	}
}

function requiredString(value: unknown, name: string): string {
	if (typeof value !== "string" || value.length < 1) throw new ToolPolicyError(`${name} must be a non-empty string`);
	return value;
}

function boundedString(value: unknown, name: string, max: number, allowEmpty: boolean): string {
	if (typeof value !== "string" || (!allowEmpty && value.length < 1) || value.length > max) {
		throw new ToolPolicyError(`${name} must be ${allowEmpty ? "a" : "a non-empty"} bounded string`);
	}
	return value;
}

function optionalBoundedInteger(value: unknown, name: string, min: number, max: number): number | undefined {
	if (value === undefined) return undefined;
	if (!Number.isSafeInteger(value) || Number(value) < min || Number(value) > max) {
		throw new ToolPolicyError(`${name} must be a bounded integer`);
	}
	return Number(value);
}

function positiveInteger(value: number, name: string): number {
	if (!Number.isSafeInteger(value) || value <= 0) throw new ToolPolicyError(`${name} must be a positive safe integer`);
	return value;
}

function assertSignal(signal: AbortSignal | undefined): void {
	if (signal?.aborted) throw new ToolPolicyError("tool execution was cancelled");
}

function validIdentifier(value: unknown): value is string {
	return typeof value === "string" && /^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$/.test(value);
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null && !Array.isArray(value);
}

function byteLength(value: string): number {
	return new TextEncoder().encode(value).byteLength;
}
