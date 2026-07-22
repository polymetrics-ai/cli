import { posix } from "node:path";
import { types as nodeTypes } from "node:util";

import type {
	AgentToolResult,
	ToolDefinition,
} from "@earendil-works/pi-coding-agent";

const DEFAULT_MAX_TOOL_OUTPUT_BYTES = 64 * 1024;
const DEFAULT_MAX_READ_CHARACTERS = 64 * 1024;
const DEFAULT_MAX_WRITE_CHARACTERS = 256 * 1024;
const MAX_TOOL_OUTPUT_BYTES = 256 * 1024;
const MAX_READ_CHARACTERS = 256 * 1024;
const MAX_WRITE_CHARACTERS = 1024 * 1024;
const MAX_PATH_CHARACTERS = 512;
const MAX_CAPABILITIES = 32;
const MAX_REFERENCES = 32;
const MAX_REFERENCE_CHARACTERS = 512;
const MAX_CAPABILITY_SUMMARY_CHARACTERS = 4 * 1024;
const MAX_CAPABILITY_SCHEMA_BYTES = 32 * 1024;
const MAX_CAPABILITY_SCHEMA_DEPTH = 64;
const MAX_CAPABILITY_SCHEMA_NODES = 4_096;
const MAX_CAPABILITY_SCHEMA_KEYS = 4_096;
const MAX_CAPABILITY_SCHEMA_ARRAY_ITEMS = 512;
const MAX_TOOL_INPUT_BYTES = MAX_WRITE_CHARACTERS + 64 * 1024;
const NATIVE_ABORTED_GETTER = Object.getOwnPropertyDescriptor(AbortSignal.prototype, "aborted")?.get;

/** The complete reviewed host-tool domain. Unknown strings have no extension path. */
export const HOST_CAPABILITY_REGISTRY = Object.freeze({
	host_inspect: Object.freeze({ mutates: false as const }),
	host_verify: Object.freeze({ mutates: true as const }),
} as const);

export type HostCapabilityName = keyof typeof HOST_CAPABILITY_REGISTRY;

const WORKSPACE_TOOL_REGISTRY = Object.freeze({
	workspace_read: Object.freeze({ mutates: false as const }),
	workspace_edit: Object.freeze({ mutates: true as const }),
	workspace_write: Object.freeze({ mutates: true as const }),
} as const);

export type WorkspaceToolName = keyof typeof WORKSPACE_TOOL_REGISTRY;
export type SessionToolName = WorkspaceToolName | HostCapabilityName;

export function isHostCapabilityName(value: unknown): value is HostCapabilityName {
	return typeof value === "string" && Object.hasOwn(HOST_CAPABILITY_REGISTRY, value);
}

export function isSessionToolName(value: unknown): value is SessionToolName {
	return typeof value === "string" &&
		(Object.hasOwn(WORKSPACE_TOOL_REGISTRY, value) || Object.hasOwn(HOST_CAPABILITY_REGISTRY, value));
}

export function sessionToolMutates(name: SessionToolName): boolean {
	return isHostCapabilityName(name)
		? HOST_CAPABILITY_REGISTRY[name].mutates
		: WORKSPACE_TOOL_REGISTRY[name].mutates;
}

const sensitivePathPatterns = [
	/(?:^|\/)\.git(?:\/|$)/i,
	/(?:^|\/)\.env(?:\.|$)/i,
	/(?:^|\/)\.envrc(?:$|\/)/i,
	/(?:^|\/)(?:credentials?|secrets?|auth)(?:[._-]|$)/i,
	/(?:^|\/)\.ssh\/(?:id_rsa|id_dsa|id_ecdsa|id_ed25519)(?:\.|$)/i,
	/(?:^|\/)(?:id_rsa|id_dsa|id_ecdsa|id_ed25519|known_hosts)(?:\.|$)/i,
	/\.(?:pem|p12|pfx|key)$/i,
	/(?:^|\/)\.(?:npmrc|yarnrc(?:\.yml)?|pnpmrc|pypirc|netrc)$/i,
	/(?:^|\/)_netrc$/i,
	/(?:^|\/)\.git-credentials$/i,
	/(?:^|\/)\.kube\/config$/i,
	/(?:^|\/)\.docker\/config\.json$/i,
	/(?:^|\/)\.config\/containers\/auth\.json$/i,
	/(?:^|\/)\.aws\/credentials$/i,
	/(?:^|\/)\.aws\/config$/i,
	/(?:^|\/)\.aws\/(?:sso|cli)\/cache(?:\/|$)/i,
	/(?:^|\/)\.azure\/accesstokens\.json$/i,
	/(?:^|\/)\.azure\/(?:msal[_-]?token[_-]?cache|token[_-]?cache)(?:[._/-]|$)/i,
	/(?:^|\/)\.config\/gcloud\/application_default_credentials\.json$/i,
	/(?:^|\/)\.config\/gcloud\/(?:legacy_credentials|access_tokens)(?:[._/-]|$)/i,
	/(?:^|\/)\.config\/gh\/hosts\.ya?ml$/i,
	/(?:^|\/)pip\/pip\.conf$/i,
	/(?:^|\/)nuget\.config$/i,
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

interface HostCapabilityContract<Name extends HostCapabilityName> {
	readonly name: Name;
	readonly description: string;
	readonly mutates: (typeof HOST_CAPABILITY_REGISTRY)[Name]["mutates"];
	readonly parameters: Readonly<Record<string, unknown>>;
	execute(input: Readonly<Record<string, unknown>>, signal?: AbortSignal): Promise<CapabilityResult>;
}

/** A narrow host action whose identity and mutability are closed by the runtime registry. */
export type HostCapability = {
	[Name in HostCapabilityName]: HostCapabilityContract<Name>;
}[HostCapabilityName];

export interface ToolAuthority {
	workspaceId: string;
	readPrefixes: string[];
	writePrefixes: string[];
	capabilityNames: HostCapabilityName[];
}

type PlainJsonSchema = Readonly<Record<string, unknown>>;
type PiSessionTool = ToolDefinition<PlainJsonSchema, unknown>;

export type SessionToolResult = Omit<AgentToolResult<unknown>, "content"> & {
	content: Array<{ type: "text"; text: string }>;
};

/** Pi 0.80.6 public custom-tool definition using its supported plain-JSON-schema path. */
export type SessionTool = Omit<PiSessionTool, "execute"> & {
	execute(
		toolCallId: string,
		params: Parameters<PiSessionTool["execute"]>[1],
		signal: Parameters<PiSessionTool["execute"]>[2],
		onUpdate?: Parameters<PiSessionTool["execute"]>[3],
		context?: Parameters<PiSessionTool["execute"]>[4],
	): Promise<SessionToolResult>;
};

export interface ToolPolicy {
	readonly names: SessionToolName[];
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
		maxToolOutputBytes: boundedPositiveInteger(
			options.maxToolOutputBytes ?? DEFAULT_MAX_TOOL_OUTPUT_BYTES,
			"maxToolOutputBytes",
			MAX_TOOL_OUTPUT_BYTES,
		),
		maxReadCharacters: boundedPositiveInteger(
			options.maxReadCharacters ?? DEFAULT_MAX_READ_CHARACTERS,
			"maxReadCharacters",
			MAX_READ_CHARACTERS,
		),
		maxWriteCharacters: boundedPositiveInteger(
			options.maxWriteCharacters ?? DEFAULT_MAX_WRITE_CHARACTERS,
			"maxWriteCharacters",
			MAX_WRITE_CHARACTERS,
		),
	};
	if (!input || typeof input !== "object" || !input.authority || typeof input.authority !== "object") {
		throw new ToolPolicyError("tool policy input and authority are required");
	}
	const capturedInput: ToolPolicyInput = Object.freeze({
		readOnly: input.readOnly,
		workspace: input.workspace,
		authority: Object.freeze({
			workspaceId: input.authority.workspaceId,
			readPrefixes: capturePolicyArray<string>(input.authority.readPrefixes, "read prefixes", 64, false),
			writePrefixes: capturePolicyArray<string>(input.authority.writePrefixes, "write prefixes", 64, input.readOnly),
			capabilityNames: capturePolicyArray<HostCapabilityName>(
				input.authority.capabilityNames,
				"capability authority",
				MAX_CAPABILITIES,
				true,
			),
		}),
		capabilities: capturePolicyArray<HostCapability>(
			input.capabilities,
			"typed host capabilities",
			MAX_CAPABILITIES,
			true,
		),
	});
	const capabilities = validatePolicyInput(capturedInput);

	const readPrefixes = normalizeScopedPrefixes(capturedInput.authority.readPrefixes, "read");
	const writePrefixes = capturedInput.readOnly && capturedInput.authority.writePrefixes.length === 0
		? capturedInput.authority.writePrefixes
		: normalizeScopedPrefixes(capturedInput.authority.writePrefixes, "write");
	const tools: SessionTool[] = [workspaceReadTool(capturedInput.workspace, readPrefixes, limits)];
	if (!capturedInput.readOnly) {
		tools.push(
			workspaceEditTool(capturedInput.workspace, writePrefixes, limits),
			workspaceWriteTool(capturedInput.workspace, writePrefixes, limits),
		);
	}

	const declared = new Set(capturedInput.authority.capabilityNames);
	for (const capability of capabilities) {
		if (!declared.has(capability.name)) {
			throw new ToolPolicyError(`undeclared capability ${JSON.stringify(capability.name)} cannot expand authority`);
		}
		if (capturedInput.readOnly && capability.mutates) continue;
		tools.push(hostCapabilityTool(capability, limits));
	}

	for (const tool of tools) Object.freeze(tool);
	const names = tools.map((tool) => {
		if (!isSessionToolName(tool.name)) throw new ToolPolicyError("tool policy constructed an unregistered tool identity");
		return tool.name;
	});
	Object.freeze(names);
	Object.freeze(tools);
	return Object.freeze({ names, tools });
}

function validatePolicyInput(input: ToolPolicyInput): HostCapability[] {
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

	const declared = new Set<HostCapabilityName>();
	for (const name of input.authority.capabilityNames) {
		if (!isHostCapabilityName(name)) {
			throw new ToolPolicyError(`capability ${JSON.stringify(name)} is not in the closed host registry`);
		}
		if (declared.has(name)) throw new ToolPolicyError(`duplicate declared capability ${name}`);
		declared.add(name);
	}
	const supplied = new Set<HostCapabilityName>();
	const capabilities: HostCapability[] = [];
	for (const capability of input.capabilities) {
		if (!capability || typeof capability !== "object") throw new ToolPolicyError("capability must be an object");
		const name = capability.name;
		const description = capability.description;
		const mutates = capability.mutates;
		const parameterSource = capability.parameters;
		const execute = capability.execute;
		if (!isHostCapabilityName(name)) {
			throw new ToolPolicyError(`capability ${JSON.stringify(name)} is not in the closed host registry`);
		}
		if (supplied.has(name)) throw new ToolPolicyError(`duplicate supplied capability ${name}`);
		supplied.add(name);
		if (!declared.has(name)) {
			throw new ToolPolicyError(`undeclared capability ${name} cannot expand authority`);
		}
		if (typeof description !== "string" || description.length < 1 || description.length > 512 ||
			/[\u0000-\u001f\u007f]/.test(description)) {
			throw new ToolPolicyError(`capability ${name} has an invalid description`);
		}
		if (mutates !== HOST_CAPABILITY_REGISTRY[name].mutates || typeof execute !== "function") {
			throw new ToolPolicyError(`capability ${name} has an invalid typed contract`);
		}
		const parameters = snapshotCapabilitySchema(parameterSource, name);
		capabilities.push(Object.freeze({
			name,
			description,
			mutates,
			parameters,
			execute(input: Readonly<Record<string, unknown>>, signal?: AbortSignal) {
				return Reflect.apply(execute, capability, [input, signal]);
			},
		}) as HostCapability);
	}
	for (const name of declared) {
		if (!supplied.has(name)) throw new ToolPolicyError(`declared capability ${name} was not supplied`);
	}
	return capabilities;
}

type JsonData = null | boolean | number | string | JsonData[] | { [key: string]: JsonData };

interface JsonSnapshotBudget {
	nodes: number;
	keys: number;
	bytes: number;
	maximumBytes?: number;
}

function snapshotCapabilitySchema(source: unknown, name: string): PlainJsonSchema {
	const snapshot = snapshotJsonData(
		source,
		0,
		{ nodes: 0, keys: 0, bytes: 0 },
		new WeakSet<object>(),
		`${name} parameter schema`,
	);
	if (!isRecord(snapshot)) {
		throw new ToolPolicyError(`capability ${name} requires a bounded parameter schema`);
	}
	if (snapshot.type !== "object" || snapshot.additionalProperties !== false) {
		throw new ToolPolicyError(`capability ${name} parameter schema must be a closed object`);
	}
	const serialized = JSON.stringify(snapshot);
	if (byteLength(serialized) > MAX_CAPABILITY_SCHEMA_BYTES) {
		throw new ToolPolicyError(`capability ${name} parameter schema exceeded its byte bound`);
	}
	return snapshot;
}

function snapshotJsonData(
	value: unknown,
	depth: number,
	budget: JsonSnapshotBudget,
	ancestors: WeakSet<object>,
	description: string,
): JsonData {
	budget.nodes += 1;
	if (budget.nodes > MAX_CAPABILITY_SCHEMA_NODES || depth > MAX_CAPABILITY_SCHEMA_DEPTH) {
		throw new ToolPolicyError(`${description} exceeded its depth or node bound`);
	}
	if (value === null) {
		addSchemaBytes(budget, 4, description);
		return value;
	}
	if (typeof value === "string") {
		addSchemaBytes(budget, byteLength(value) + 2, description);
		return value;
	}
	if (typeof value === "boolean") {
		addSchemaBytes(budget, 5, description);
		return value;
	}
	if (typeof value === "number") {
		if (!Number.isFinite(value)) throw new ToolPolicyError(`${description} contains a non-JSON number`);
		addSchemaBytes(budget, 24, description);
		return value;
	}
	if (typeof value !== "object") throw new ToolPolicyError(`${description} contains a non-JSON value`);
	if (nodeTypes.isProxy(value)) throw new ToolPolicyError(`${description} cannot be a Proxy`);
	if (ancestors.has(value)) throw new ToolPolicyError(`${description} contains a cycle`);
	ancestors.add(value);
	addSchemaBytes(budget, 2, description);

	try {
		if (Array.isArray(value)) {
			if (value.length > MAX_CAPABILITY_SCHEMA_ARRAY_ITEMS) {
				throw new ToolPolicyError(`${description} array exceeded its bound`);
			}
			const lengthDescriptor = Object.getOwnPropertyDescriptor(value, "length");
			if (!lengthDescriptor || !("value" in lengthDescriptor) || lengthDescriptor.value !== value.length ||
				lengthDescriptor.enumerable || lengthDescriptor.get || lengthDescriptor.set) {
				throw new ToolPolicyError(`${description} array must be dense, plain, and bounded`);
			}
			let enumerableItems = 0;
			for (const key in value) {
				if (!Object.hasOwn(value, key) || !/^(?:0|[1-9][0-9]*)$/.test(key) || Number(key) >= value.length) {
					throw new ToolPolicyError(`${description} array must be dense, plain, and bounded`);
				}
				enumerableItems += 1;
				if (enumerableItems > value.length) {
					throw new ToolPolicyError(`${description} array must be dense, plain, and bounded`);
				}
			}
			if (enumerableItems !== value.length) {
				throw new ToolPolicyError(`${description} array must be dense, plain, and bounded`);
			}
			budget.keys += value.length;
			if (budget.keys > MAX_CAPABILITY_SCHEMA_KEYS) throw new ToolPolicyError(`${description} exceeded its key bound`);
			const result: JsonData[] = [];
			for (let index = 0; index < value.length; index += 1) {
				const descriptor = Object.getOwnPropertyDescriptor(value, String(index));
				if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
					throw new ToolPolicyError(`${description} array contains an accessor or sparse item`);
				}
				result.push(snapshotJsonData(descriptor.value, depth + 1, budget, ancestors, description));
			}
			return Object.freeze(result) as JsonData[];
		}

		const prototype = Object.getPrototypeOf(value);
		if (prototype !== Object.prototype && prototype !== null) {
			throw new ToolPolicyError(`${description} must contain plain JSON objects only`);
		}
		// JavaScript has no incremental API for hidden or symbol keys. Capture only bounded
		// enumerable JSON data, descriptor by descriptor, and deliberately discard all other
		// source authority instead of materializing an attacker-sized own-key array.
		const result = Object.create(null) as { [key: string]: JsonData };
		for (const key in value) {
			if (!Object.hasOwn(value, key)) {
				throw new ToolPolicyError(`${description} must contain own JSON fields only`);
			}
			budget.keys += 1;
			if (budget.keys > MAX_CAPABILITY_SCHEMA_KEYS) {
				throw new ToolPolicyError(`${description} exceeded its key bound`);
			}
			addSchemaBytes(budget, byteLength(key) + 3, description);
			const descriptor = Object.getOwnPropertyDescriptor(value, key);
			if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
				throw new ToolPolicyError(`${description} contains an accessor field`);
			}
			Object.defineProperty(result, key, {
				value: snapshotJsonData(descriptor.value, depth + 1, budget, ancestors, description),
				enumerable: true,
				writable: false,
				configurable: false,
			});
		}
		return Object.freeze(result);
	} finally {
		ancestors.delete(value);
	}
}

function addSchemaBytes(budget: JsonSnapshotBudget, bytes: number, description: string): void {
	budget.bytes += bytes;
	if (budget.bytes > (budget.maximumBytes ?? MAX_CAPABILITY_SCHEMA_BYTES)) {
		throw new ToolPolicyError(`${description} exceeded its incremental byte bound`);
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
		async execute(_callId, rawParams, signal) {
			return toolBoundary("workspace read", async () => {
				assertSignal(signal);
				const params = recordParams(rawParams, "workspace_read");
				assertOnlyKeys(params, ["path", "offset", "limit"], "workspace_read");
				const path = validateScopedPath(requiredString(params.path, "path"), prefixes);
				const offset = optionalBoundedInteger(params.offset, "offset", 0, Number.MAX_SAFE_INTEGER);
				const limit = optionalBoundedInteger(params.limit, "limit", 1, limits.maxReadCharacters);
				const value = await workspace.readText(path, { offset, limit, signal });
				if (typeof value !== "string") throw new ToolPolicyError("workspace read returned non-text data");
				if (value.length > limits.maxReadCharacters) throw new ToolPolicyError("workspace read exceeded the bounded output limit");
				return textResult(redactSensitiveText(value), limits.maxToolOutputBytes);
			});
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
		async execute(_callId, rawParams, signal) {
			return toolBoundary("workspace edit", async () => {
				assertSignal(signal);
				const params = recordParams(rawParams, "workspace_edit");
				assertOnlyKeys(params, ["path", "oldText", "newText"], "workspace_edit");
				const path = validateScopedPath(requiredString(params.path, "path"), prefixes);
				const oldText = boundedString(params.oldText, "oldText", limits.maxWriteCharacters, false);
				const newText = boundedString(params.newText, "newText", limits.maxWriteCharacters, true);
				const result = await workspace.editText(path, oldText, newText, signal);
				return mutationResult(result, limits.maxToolOutputBytes);
			});
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
		async execute(_callId, rawParams, signal) {
			return toolBoundary("workspace write", async () => {
				assertSignal(signal);
				const params = recordParams(rawParams, "workspace_write");
				assertOnlyKeys(params, ["path", "content"], "workspace_write");
				const path = validateScopedPath(requiredString(params.path, "path"), prefixes);
				const content = boundedString(params.content, "content", limits.maxWriteCharacters, true);
				const result = await workspace.writeText(path, content, signal);
				return mutationResult(result, limits.maxToolOutputBytes);
			});
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
		async execute(_callId, rawParams, signal) {
			return toolBoundary(`capability ${capability.name}`, async () => {
				assertSignal(signal);
				const params = recordParams(rawParams, capability.name);
				const inputSize = byteLength(JSON.stringify(params));
				if (inputSize > limits.maxWriteCharacters) throw new ToolPolicyError(`${capability.name} input exceeded its bound`);
				const result = captureCapabilityResult(capability.name, await capability.execute(params, signal));
				return textResult(JSON.stringify({
					status: result.status,
					summary: redactSensitiveText(result.summary),
					references: result.references.map(redactSensitiveText),
				}), limits.maxToolOutputBytes);
			});
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
	let allowed = false;
	for (let index = 0; index < prefixes.length; index += 1) {
		const prefix = prefixes[index];
		if (prefix === "." || normalized === prefix || normalized.startsWith(`${prefix}/`)) {
			allowed = true;
			break;
		}
	}
	if (!allowed) {
		throw new ToolPolicyError(`workspace path ${JSON.stringify(normalized)} is outside the declared scope`);
	}
	return normalized;
}

export interface RedactionScanMetrics {
	sourceLength: number;
	cursorAdvances: number;
	cursorRegressions: number;
	maxMainCursorVisits: number;
	keyCharacterVisits: number;
	boundaryCharacterVisits: number;
	totalWork: number;
	/** @deprecated Cycle 16 removes this compatibility counter after retained tests migrate. */
	lineBoundaryVisits: number;
	/** @deprecated Cycle 16 removes this compatibility counter after retained tests migrate. */
	keyStartVisits?: number;
	/** @deprecated Cycle 16 removes this compatibility counter after retained tests migrate. */
	totalCharacterVisits?: number;
}

export function redactSensitiveText(value: string): string;
export function redactSensitiveText(value: string, metrics: RedactionScanMetrics): string;
export function redactSensitiveText(value: string, metrics?: RedactionScanMetrics | number): string {
	if (typeof value !== "string") return "[REDACTED]";
	const scanMetrics = typeof metrics === "object" && metrics !== null ? metrics : undefined;
	if (scanMetrics) initializeRedactionMetrics(scanMetrics, value.length);
	return redactStructuredAssignments(redactStrongCredentialSyntax(redactPrivateKeyBlocks(value)), scanMetrics);
}

type SensitiveAssignmentKind = "authorization" | "secret" | "unknown-sensitive";
type AssignmentKeyClassification = SensitiveAssignmentKind | "public";
type SensitiveAssignmentContext = "flow" | "line";
type LexicalQuote = "\"" | "'";
type FlowCloser = "}" | "]";

type LexicalMode =
	| { kind: "plain" }
	| { kind: "quoted"; quote: LexicalQuote; multiline: boolean };

interface StructuredScannerState {
	index: number;
	lineStart: number;
	lineEnd: number;
	structuredKeyStart: number | undefined;
	mode: LexicalMode;
	flowClosers: FlowCloser[];
	flowOverflowDepth: number;
}

interface SensitiveAssignment {
	kind: AssignmentKeyClassification;
	context: SensitiveAssignmentContext;
	keyColumn: number;
	normalizedKey: string;
	delimiter: ":" | "=";
	valueStart: number;
}

interface RedactionRange {
	start: number;
	end: number;
	replacement: string;
}

interface RedactionDecision {
	range?: RedactionRange;
	resumeAt: number;
}

interface AssignmentKeyCandidate {
	exactKey?: string;
	delimiter: ":" | "=";
	delimiterIndex: number;
}

const REDACTED_TEXT = "[REDACTED]";
const secretAssignmentKeys = new Set([
	"apikey",
	"accesstoken",
	"refreshtoken",
	"token",
	"password",
	"passwd",
	"secret",
	"clientsecret",
	"cookie",
	"setcookie",
]);
const sensitiveAssignmentTerminals = new Set([
	...secretAssignmentKeys,
	"privatekey",
	"databaseurl",
	"awssecretaccesskey",
	"secretaccesskey",
	"githubtoken",
]);
const sensitiveAssignmentPaths = new Set([
	"api.key",
	"private.key",
	"database.url",
	"aws.secret.access.key",
]);
const publicAssignmentTerminals = new Set(["safe", "enabled", "retained", "flavor", "message"]);
const publicAssignmentPaths = new Set([
	"api.version",
	"api.key.version",
	"private.key.algorithm",
	"database.url.scheme",
]);

function redactStructuredAssignments(value: string, metrics?: RedactionScanMetrics): string {
	const ranges: RedactionRange[] = [];
	// One monotonic cursor owns line, quote, comment, and balanced flow transitions. Value parsers
	// consume their complete span before the cursor resumes, so skipped nested delimiters cannot
	// corrupt the outer flow state and no global regex repeatedly rescans the input.
	const state: StructuredScannerState = {
		index: 0,
		lineStart: 0,
		lineEnd: findLineEnd(value, 0, metrics),
		structuredKeyStart: findStructuredKeyStart(value, 0, metrics),
		mode: { kind: "plain" },
		flowClosers: [],
		flowOverflowDepth: 0,
	};
	while (state.index < value.length) {
		if (state.index >= state.lineEnd) {
			advanceScannerLine(value, state, metrics);
			continue;
		}

		const character = value[state.index];
		if (state.mode.kind === "quoted") {
			if (state.mode.quote === '"' && character === "\\") {
				advanceScannerIndex(state, Math.min(state.lineEnd, state.index + 2), metrics);
				continue;
			}
			if (state.mode.quote === "'" && character === "'" && value[state.index + 1] === "'") {
				advanceScannerIndex(state, Math.min(state.lineEnd, state.index + 2), metrics);
				continue;
			}
			if (character === state.mode.quote) state.mode = { kind: "plain" };
			advanceScannerIndex(state, state.index + 1, metrics);
			continue;
		}

		const context: SensitiveAssignmentContext =
			state.flowClosers.length > 0 || state.flowOverflowDepth > 0 ? "flow" : "line";
		const allowKey = state.index === state.structuredKeyStart ||
			(context === "flow" && isFlowMappingKeyStart(value, state.index, state));
		const assignment = sensitiveAssignmentAt(
			value,
			state.index,
			state.lineStart,
			state.lineEnd,
			allowKey,
			context,
			metrics,
		);
		if (!assignment) {
			if (isCommentStart(value, state.index, state.lineStart)) {
				advanceScannerIndex(state, state.lineEnd, metrics);
				continue;
			}
			if ((character === '"' || character === "'") && isQuotedSegmentStart(value, state.index, state.lineStart)) {
				state.mode = {
					kind: "quoted",
					quote: character,
					multiline: state.flowClosers.length > 0 || state.flowOverflowDepth > 0,
				};
				advanceScannerIndex(state, state.index + 1, metrics);
				continue;
			}
			const closer = flowCloserForOpener(character);
			if (closer) {
				if (state.flowClosers.length < 256) state.flowClosers.push(closer);
				else state.flowOverflowDepth += 1;
			} else if (character === "}" || character === "]") {
				if (state.flowOverflowDepth > 0) state.flowOverflowDepth -= 1;
				else if (character === state.flowClosers[state.flowClosers.length - 1]) state.flowClosers.pop();
			}
			advanceScannerIndex(state, state.index + 1, metrics);
			continue;
		}
		const decision = redactionForAssignment(value, assignment, state.lineEnd, metrics);
		if (decision.range) ranges.push(decision.range);
		advanceScannerTo(value, state, Math.max(state.index + 1, decision.resumeAt), metrics);
	}
	if (ranges.length === 0) return value;
	let output = "";
	let cursor = 0;
	for (const range of ranges) {
		output += value.slice(cursor, range.start) + range.replacement;
		cursor = range.end;
	}
	return output + value.slice(cursor);
}

function sensitiveAssignmentAt(
	value: string,
	index: number,
	lineStart: number,
	lineEnd: number,
	allowKey: boolean,
	context: SensitiveAssignmentContext,
	metrics?: RedactionScanMetrics,
): SensitiveAssignment | undefined {
	if (!allowKey) return undefined;
	if (context === "line" && isStandalonePublicScalar(value, index, lineEnd)) return undefined;
	const candidate = scanAssignmentKeyCandidate(value, index, lineEnd, context, metrics);
	if (!candidate) return undefined;
	const classification = candidate.exactKey === undefined
		? "unknown-sensitive"
		: assignmentKeyClassification(candidate.exactKey);
	const normalizedKey = candidate.exactKey === undefined
		? ""
		: (candidate.exactKey.toLowerCase().split(".").at(-1) ?? candidate.exactKey).replace(/[-_]/g, "");
	let cursor = candidate.delimiterIndex + 1;
	recordTotalVisits(metrics, 1);
	while (isHorizontalWhitespace(value[cursor])) {
		cursor += 1;
		recordTotalVisits(metrics, 1);
	}
	return {
		kind: classification,
		context,
		keyColumn: index - lineStart,
		normalizedKey,
		delimiter: candidate.delimiter,
		valueStart: cursor,
	};
}

function scanAssignmentKeyCandidate(
	value: string,
	start: number,
	lineEnd: number,
	context: SensitiveAssignmentContext,
	metrics?: RedactionScanMetrics,
): AssignmentKeyCandidate | undefined {
	const quote = value[start] === '"' || value[start] === "'" ? value[start] as LexicalQuote : undefined;
	if (quote) return scanQuotedAssignmentKeyCandidate(value, start, lineEnd, quote, metrics);
	let cursor = start;
	let exact = true;
	let keyLength = 0;
	let pendingWhitespace = false;
	let exactKey = "";
	while (cursor < lineEnd) {
		const character = value[cursor];
		if (character === ":" || character === "=") {
			if (keyLength === 0) return undefined;
			if (character === "=" || isAdmittedColonDelimiter(value, cursor, lineEnd, context)) {
				return {
					exactKey: exact && keyLength <= 64 ? exactKey : undefined,
					delimiter: character,
					delimiterIndex: cursor,
				};
			}
			exact = false;
		}
		if (isHorizontalWhitespace(character)) {
			pendingWhitespace = keyLength > 0;
		} else {
			if (pendingWhitespace) exact = false;
			pendingWhitespace = false;
			keyLength += 1;
			if (keyLength <= 64 && isAssignmentKeyCharacter(character)) exactKey += character;
			else exact = false;
		}
		cursor += 1;
		recordKeyCharacterVisit(metrics);
		recordTotalVisits(metrics, 1);
	}
	return undefined;
}

function scanQuotedAssignmentKeyCandidate(
	value: string,
	start: number,
	lineEnd: number,
	quote: LexicalQuote,
	metrics?: RedactionScanMetrics,
): AssignmentKeyCandidate | undefined {
	let cursor = start + 1;
	let decoded = "";
	let decodedLength = 0;
	let exact = true;
	let closedAt: number | undefined;
	let innerDelimiter: number | undefined;
	const append = (character: string): void => {
		decodedLength += 1;
		recordKeyCharacterVisit(metrics);
		if (decodedLength <= 64 && isAssignmentKeyCharacter(character)) decoded += character;
		else exact = false;
	};
	while (cursor < lineEnd) {
		const character = value[cursor];
		recordTotalVisits(metrics, 1);
		if ((character === ":" || character === "=") && innerDelimiter === undefined) innerDelimiter = cursor;
		if (character === quote) {
			if (quote === "'" && value[cursor + 1] === "'") {
				append("'");
				cursor += 2;
				recordTotalVisits(metrics, 1);
				continue;
			}
			closedAt = cursor;
			break;
		}
		if (quote === '"' && character === "\\") {
			const escaped = value[cursor + 1];
			if (escaped === "u") {
				const hex = value.slice(cursor + 2, cursor + 6);
				if (/^[0-9a-fA-F]{4}$/.test(hex)) append(String.fromCharCode(Number.parseInt(hex, 16)));
				else exact = false;
				const consumed = Math.min(6, lineEnd - cursor);
				cursor += consumed;
				recordTotalVisits(metrics, Math.max(0, consumed - 1));
				continue;
			}
			const simpleEscapes: Record<string, string> = {
				'"': '"', "\\": "\\", "/": "/", b: "\b", f: "\f", n: "\n", r: "\r", t: "\t",
			};
			if (escaped !== undefined && Object.hasOwn(simpleEscapes, escaped)) append(simpleEscapes[escaped]);
			else exact = false;
			const consumed = Math.min(2, lineEnd - cursor);
			cursor += consumed;
			recordTotalVisits(metrics, Math.max(0, consumed - 1));
			continue;
		}
		append(character);
		cursor += 1;
	}

	if (closedAt !== undefined) {
		cursor = closedAt + 1;
		while (cursor < lineEnd && isHorizontalWhitespace(value[cursor])) {
			cursor += 1;
			recordTotalVisits(metrics, 1);
		}
		if (value[cursor] === ":" || value[cursor] === "=") {
			return {
				exactKey: exact && decodedLength > 0 && decodedLength <= 64 ? decoded : undefined,
				delimiter: value[cursor] as ":" | "=",
				delimiterIndex: cursor,
			};
		}
		// A syntactically closed quoted scalar with no external delimiter is public text. An
		// interior colon/equals must never be reinterpreted as an assignment delimiter.
		if (cursor === lineEnd) return undefined;
	}
	if (innerDelimiter !== undefined) {
		return {
			delimiter: value[innerDelimiter] as ":" | "=",
			delimiterIndex: innerDelimiter,
		};
	}
	return undefined;
}

function redactionForAssignment(
	value: string,
	assignment: SensitiveAssignment,
	lineEnd: number,
	metrics?: RedactionScanMetrics,
): RedactionDecision {
	const quote = value[assignment.valueStart];
	if (assignment.kind === "public") {
		if (quote === '"' || quote === "'") {
			const quoteEnd = scanPublicQuotedScalarEnd(value, assignment.valueStart + 1, quote, metrics);
			if (quoteEnd !== undefined) return { resumeAt: quoteEnd + 1 };
			return quotedValueRedaction(
				value,
				{ ...assignment, kind: "unknown-sensitive" },
				quote,
				lineEnd,
				metrics,
			);
		}
		const scalarEnd = findUnquotedValueEnd(value, assignment.valueStart, lineEnd, assignment.context, metrics);
		return { resumeAt: assignment.context === "flow" ? scalarEnd : lineEnd };
	}
	if (quote === '"' || quote === "'") {
		return quotedValueRedaction(value, assignment, quote, lineEnd, metrics);
	}
	if (assignment.kind !== "authorization" && isYamlBlockHeader(value.slice(assignment.valueStart, lineEnd))) {
		return yamlBlockRedaction(value, assignment, lineEnd, metrics);
	}
	return unquotedValueRedaction(value, assignment, lineEnd, metrics);
}

function quotedValueRedaction(
	value: string,
	assignment: SensitiveAssignment,
	quote: string,
	lineEnd: number,
	metrics?: RedactionScanMetrics,
): RedactionDecision {
	const contentStart = assignment.valueStart + 1;
	const quoteEnd = scanQuotedValueEnd(value, contentStart, quote, lineEnd, assignment.context, metrics);
	const contentEnd = quoteEnd ?? lineEnd;
	const resumeAt = quoteEnd === undefined ? afterLineEnding(value, lineEnd) : quoteEnd + 1;
	if (assignment.kind === "secret" || assignment.kind === "unknown-sensitive") {
		if (value.slice(contentStart, contentEnd).trim().length === 0) return { resumeAt };
		return { range: { start: contentStart, end: contentEnd, replacement: REDACTED_TEXT }, resumeAt };
	}

	const credentialStart = authorizationCredentialStart(
		value,
		contentStart,
		contentEnd,
		quoteEnd !== undefined && quoteEnd > lineEnd,
	);
	return credentialStart === undefined
		? { resumeAt }
		: { range: { start: credentialStart, end: contentEnd, replacement: REDACTED_TEXT }, resumeAt };
}

function yamlBlockRedaction(
	value: string,
	assignment: SensitiveAssignment,
	headerEnd: number,
	metrics?: RedactionScanMetrics,
): RedactionDecision {
	const contentStart = afterLineEnding(value, headerEnd);
	if (contentStart === headerEnd) return { resumeAt: headerEnd };
	let cursor = contentStart;
	let blockEnd = contentStart;
	let contentIndent: string | undefined;
	while (cursor < value.length) {
		const lineEnd = findLineEnd(value, cursor, metrics);
		const line = value.slice(cursor, lineEnd);
		const indentLength = leadingIndentLength(line);
		const blank = line.slice(indentLength).length === 0;
		if (!blank && indentLength <= assignment.keyColumn) break;
		if (!blank && contentIndent === undefined) contentIndent = line.slice(0, indentLength);
		blockEnd = afterLineEnding(value, lineEnd);
		if (blockEnd === lineEnd) break;
		cursor = blockEnd;
	}
	const block = value.slice(contentStart, blockEnd);
	if (block.trim().length === 0) return { resumeAt: blockEnd };
	const indent = contentIndent ?? " ".repeat(assignment.keyColumn + 2);
	return {
		range: {
			start: contentStart,
			end: blockEnd,
			replacement: `${indent}${REDACTED_TEXT}${trailingLineEnding(block)}`,
		},
		resumeAt: blockEnd,
	};
}

function unquotedValueRedaction(
	value: string,
	assignment: SensitiveAssignment,
	lineEnd: number,
	metrics?: RedactionScanMetrics,
): RedactionDecision {
	let scalarEnd = findUnquotedValueEnd(value, assignment.valueStart, lineEnd, assignment.context, metrics);
	scalarEnd = trimHorizontalEnd(value, assignment.valueStart, scalarEnd);
	const continuation = assignment.context === "line"
		? yamlPlainContinuation(value, lineEnd, assignment.keyColumn, metrics)
		: undefined;
	const resumeAt = continuation?.end ?? scalarEnd;
	const unredactedResumeAt = assignment.context === "flow" ? scalarEnd : assignment.valueStart;
	if (scalarEnd <= assignment.valueStart) {
		if (!continuation) return { resumeAt };
		return {
			range: {
				start: continuation.start,
				end: continuation.end,
				replacement: `${continuation.indent}${REDACTED_TEXT}${trailingLineEnding(continuation.value)}`,
			},
			resumeAt,
		};
	}
	if (assignment.kind === "authorization") {
		const authorizationEnd = continuation?.end ?? scalarEnd;
		const credentialStart = authorizationCredentialStart(
			value,
			assignment.valueStart,
			authorizationEnd,
			continuation !== undefined,
		);
		return credentialStart === undefined
			? { resumeAt: unredactedResumeAt }
			: {
				range: {
					start: credentialStart,
					end: authorizationEnd,
					replacement: `${REDACTED_TEXT}${trailingLineEnding(value.slice(credentialStart, authorizationEnd))}`,
				},
				resumeAt,
			};
	}

	const scalar = value.slice(assignment.valueStart, scalarEnd);
	const documentaryAssignmentProse = assignment.delimiter === ":" &&
		assignment.keyColumn === 0 && continuation === undefined &&
		isDocumentaryAssignmentProse(scalar);
	const ambiguousLineProse = assignment.kind === "secret" && assignment.context === "line" &&
		assignment.delimiter === ":" &&
		assignment.keyColumn === 0 && continuation === undefined &&
		["token", "password", "passwd", "secret"].includes(assignment.normalizedKey) &&
		containsWhitespace(scalar);
	if (ambiguousLineProse || documentaryAssignmentProse) return { resumeAt: unredactedResumeAt };
	if (assignment.kind !== "unknown-sensitive" && isPublicScalar(scalar)) return { resumeAt };
	const redactionEnd = continuation?.end ?? scalarEnd;
	return {
		range: {
			start: assignment.valueStart,
			end: redactionEnd,
			replacement: `${REDACTED_TEXT}${trailingLineEnding(value.slice(assignment.valueStart, redactionEnd))}`,
		},
		resumeAt,
	};
}

function isDocumentaryAssignmentProse(value: string): boolean {
	const prose = value.trim();
	return /^(?:number|name|description|meaning|definition|count)\s+of\b/i.test(prose) ||
		/^(?:describes?|explains?|means?|refers?\s+to)\b/i.test(prose);
}

interface YamlPlainContinuation {
	start: number;
	end: number;
	indent: string;
	value: string;
}

function yamlPlainContinuation(
	value: string,
	lineEnd: number,
	keyColumn: number,
	metrics?: RedactionScanMetrics,
): YamlPlainContinuation | undefined {
	const start = afterLineEnding(value, lineEnd);
	if (start === lineEnd) return undefined;
	let cursor = start;
	let end = start;
	let indent: string | undefined;
	let hasContent = false;
	while (cursor < value.length) {
		const continuationLineEnd = findLineEnd(value, cursor, metrics);
		const line = value.slice(cursor, continuationLineEnd);
		const indentLength = leadingIndentLength(line);
		const blank = line.slice(indentLength).length === 0;
		if (!blank && indentLength <= keyColumn) break;
		if (!blank) {
			hasContent = true;
			indent ??= line.slice(0, indentLength);
		}
		end = afterLineEnding(value, continuationLineEnd);
		if (end === continuationLineEnd) break;
		cursor = end;
	}
	if (!hasContent) return undefined;
	return {
		start,
		end,
		indent: indent ?? " ".repeat(keyColumn + 2),
		value: value.slice(start, end),
	};
}

function authorizationCredentialStart(
	value: string,
	start: number,
	end: number,
	allowCredentialWhitespace: boolean,
): number | undefined {
	let cursor = start;
	while (cursor < end && isWhitespace(value[cursor])) cursor += 1;
	const schemeStart = cursor;
	while (cursor < end && !isWhitespace(value[cursor])) cursor += 1;
	const scheme = value.slice(schemeStart, cursor).toLowerCase();
	if (scheme.length === 0) return undefined;
	if (cursor >= end) return schemeStart;
	while (cursor < end && isWhitespace(value[cursor])) cursor += 1;
	if (cursor >= end) return undefined;
	const parameterized = scheme === "digest" || scheme === "signature" || scheme.startsWith("aws4-");
	if (!allowCredentialWhitespace && !parameterized && containsWhitespace(value.slice(cursor, end))) return undefined;
	return cursor;
}

function redactPrivateKeyBlocks(value: string): string {
	const lower = value.toLowerCase();
	const beginPrefix = "-----begin ";
	const headerSuffix = " private key-----";
	let cursor = 0;
	let searchAt = 0;
	let output = "";
	while (searchAt < value.length) {
		const begin = lower.indexOf(beginPrefix, searchAt);
		if (begin < 0) break;
		const labelStart = begin + beginPrefix.length;
		let label = "";
		let headerEnd: number;
		let headerLength: number;
		if (lower.startsWith("private key-----", labelStart)) {
			headerEnd = labelStart;
			headerLength = "private key-----".length;
		} else {
			const headerWindow = lower.slice(labelStart, labelStart + 64 + headerSuffix.length);
			const relativeHeaderEnd = headerWindow.indexOf(headerSuffix);
			if (relativeHeaderEnd <= 0 || relativeHeaderEnd > 64) {
				searchAt = labelStart;
				continue;
			}
			headerEnd = labelStart + relativeHeaderEnd;
			headerLength = headerSuffix.length;
			label = lower.slice(labelStart, headerEnd);
			if (!/^[a-z0-9 ]+$/.test(label)) {
				searchAt = labelStart;
				continue;
			}
		}
		const endMarker = label.length > 0
			? `-----end ${label} private key-----`
			: "-----end private key-----";
		const end = lower.indexOf(endMarker, headerEnd + headerLength);
		if (end < 0) {
			output += value.slice(cursor, begin) + "[REDACTED PRIVATE KEY]";
			return output;
		}
		output += value.slice(cursor, begin) + "[REDACTED PRIVATE KEY]";
		cursor = end + endMarker.length;
		searchAt = cursor;
	}
	return output + value.slice(cursor);
}

function redactStrongCredentialSyntax(value: string): string {
	return value
		.replace(/^(\s*(?:set-cookie|cookie)\s*:\s*)([^\r\n]*)/gim, `$1${REDACTED_TEXT}`)
		.replace(/(\b(?:request|response)\s+headers?\s+(?:set-cookie|cookie)\s*:\s*)([^\r\n]*)/gi,
			`$1${REDACTED_TEXT}`)
		.replace(/\b([a-z][a-z0-9+.-]*:\/\/[^/\s:@]+:)([^@\s/]+)(@)/gi, `$1${REDACTED_TEXT}$3`)
		.replace(/([?&#](?:access[_-]?token|refresh[_-]?token|api[_-]?key|password|secret|client[_-]?secret)=)([^&#\s]+)/gi,
			`$1${REDACTED_TEXT}`)
		.replace(/((?:^|[/:])_authToken\s*=\s*)([^\s#]+)/gim, `$1${REDACTED_TEXT}`)
		.replace(/(\bpassword[\t ]+)(?![=:])([^\s#]+)/gi, `$1${REDACTED_TEXT}`);
}

function assignmentKeyClassification(key: string): AssignmentKeyClassification {
	const { segments, path, terminal } = canonicalAssignmentKey(key);
	if ((segments.length === 1 && publicAssignmentTerminals.has(terminal)) || publicAssignmentPaths.has(path)) {
		return "public";
	}
	if (terminal === "authorization" || terminal === "proxyauthorization") return "authorization";
	if (sensitiveAssignmentTerminals.has(terminal) || sensitiveAssignmentPaths.has(path)) return "secret";
	// The structured grammar is deliberately closed: once assignment syntax is recognized, only
	// reviewed public metadata escapes redaction. Unknown aliases do not become an extension path.
	return "unknown-sensitive";
}

function canonicalAssignmentKey(key: string): {
	segments: string[];
	path: string;
	terminal: string;
} {
	const segments = key.toLowerCase().split(".").map((segment) => segment.replace(/[-_]/g, ""));
	return { segments, path: segments.join("."), terminal: segments.at(-1) ?? "" };
}

function findStructuredKeyStart(
	value: string,
	lineStart: number,
	metrics?: RedactionScanMetrics,
): number | undefined {
	let cursor = lineStart;
	while (isHorizontalWhitespace(value[cursor])) {
		cursor += 1;
		recordKeyStartVisit(metrics);
	}
	if (value[cursor] === "-") {
		cursor += 1;
		recordKeyStartVisit(metrics);
		if (!isHorizontalWhitespace(value[cursor])) return undefined;
		while (isHorizontalWhitespace(value[cursor])) {
			cursor += 1;
			recordKeyStartVisit(metrics);
		}
	} else if (value.slice(cursor, cursor + 6).toLowerCase() === "export" && isHorizontalWhitespace(value[cursor + 6])) {
		cursor += 6;
		recordKeyStartVisit(metrics, 6);
		while (isHorizontalWhitespace(value[cursor])) {
			cursor += 1;
			recordKeyStartVisit(metrics);
		}
	}
	return isPotentialAssignmentCandidateStart(value[cursor]) ? cursor : undefined;
}

function advanceScannerLine(
	value: string,
	state: StructuredScannerState,
	metrics?: RedactionScanMetrics,
): void {
	const nextLineStart = afterLineEnding(value, state.lineEnd);
	if (state.mode.kind !== "quoted" || !state.mode.multiline) state.mode = { kind: "plain" };
	if (nextLineStart <= state.lineEnd) {
		advanceScannerIndex(state, value.length, metrics);
		state.structuredKeyStart = undefined;
		return;
	}
	advanceScannerIndex(state, nextLineStart, metrics);
	state.lineStart = nextLineStart;
	state.lineEnd = findLineEnd(value, nextLineStart, metrics);
	state.structuredKeyStart = findStructuredKeyStart(value, nextLineStart, metrics);
}

function advanceScannerTo(
	value: string,
	state: StructuredScannerState,
	resumeAt: number,
	metrics?: RedactionScanMetrics,
): void {
	const target = Math.min(value.length, resumeAt);
	let movedLine = false;
	while (state.lineEnd < target) {
		const nextLineStart = afterLineEnding(value, state.lineEnd);
		if (nextLineStart <= state.lineEnd) break;
		state.lineStart = nextLineStart;
		state.lineEnd = findLineEnd(value, nextLineStart, metrics);
		movedLine = true;
	}
	advanceScannerIndex(state, target, metrics);
	state.mode = { kind: "plain" };
	if (movedLine) state.structuredKeyStart = findStructuredKeyStart(value, state.lineStart, metrics);
	if (state.structuredKeyStart !== undefined && state.structuredKeyStart < target) {
		state.structuredKeyStart = undefined;
	}
}

function isFlowMappingKeyStart(value: string, index: number, state: StructuredScannerState): boolean {
	if (!["}", "]"].includes(state.flowClosers[state.flowClosers.length - 1] ?? "") ||
		!isPotentialAssignmentCandidateStart(value[index])) {
		return false;
	}
	let cursor = index - 1;
	while (cursor >= state.lineStart && isHorizontalWhitespace(value[cursor])) cursor -= 1;
	return value[cursor] === "{" || value[cursor] === "[" || value[cursor] === ",";
}

function flowCloserForOpener(character: string | undefined): FlowCloser | undefined {
	if (character === "{") return "}";
	if (character === "[") return "]";
	return undefined;
}

function isCommentStart(value: string, index: number, lineStart: number): boolean {
	return value[index] === "#" && (index === lineStart || isHorizontalWhitespace(value[index - 1]));
}

function isQuotedSegmentStart(value: string, index: number, lineStart: number): boolean {
	if (value[index - 1] === "\\") return false;
	let cursor = index - 1;
	while (cursor >= lineStart && isHorizontalWhitespace(value[cursor])) cursor -= 1;
	if (cursor < lineStart) return true;
	if (value[cursor] === "{" || value[cursor] === "[" || value[cursor] === "," ||
		value[cursor] === ":" || value[cursor] === "=") return true;
	if (value[cursor] !== "-") return false;
	let beforeHyphen = cursor - 1;
	while (beforeHyphen >= lineStart && isHorizontalWhitespace(value[beforeHyphen])) beforeHyphen -= 1;
	return beforeHyphen < lineStart;
}

function isAssignmentKeyCharacter(character: string | undefined): boolean {
	if (character === undefined) return false;
	const code = character.charCodeAt(0);
	return (code >= 48 && code <= 57) || (code >= 65 && code <= 90) ||
		(code >= 97 && code <= 122) || character === "_" || character === "-" || character === ".";
}

function isAdmittedColonDelimiter(
	value: string,
	index: number,
	lineEnd: number,
	context: SensitiveAssignmentContext,
): boolean {
	const next = value[index + 1];
	if (index + 1 >= lineEnd || isHorizontalWhitespace(next) || next === "#") return true;
	return context === "flow" && (next === '"' || next === "'" || next === "{" || next === "[");
}

function isStandalonePublicScalar(value: string, start: number, lineEnd: number): boolean {
	const scalar = value.slice(start, lineEnd).trimEnd();
	if (/^[A-Za-z][A-Za-z0-9+.-]*:\/\/[^\s]+$/.test(scalar)) return true;
	if (/^(?:\d{4}-\d{2}-\d{2}T)?\d{1,2}:\d{2}(?::\d{2}(?:\.\d+)?)?(?:Z|[+-]\d{2}:\d{2})?$/.test(scalar)) {
		return true;
	}
	const quote = scalar[0];
	if (quote !== '"' && quote !== "'") return false;
	let cursor = 1;
	while (cursor < scalar.length) {
		if (quote === '"' && scalar[cursor] === "\\") {
			cursor += 2;
			continue;
		}
		if (quote === "'" && scalar[cursor] === "'" && scalar[cursor + 1] === "'") {
			cursor += 2;
			continue;
		}
		if (scalar[cursor] === quote) return cursor === scalar.length - 1;
		cursor += 1;
	}
	return false;
}

function isPotentialAssignmentCandidateStart(character: string | undefined): boolean {
	return character !== undefined && !isWhitespace(character) && character !== "#" &&
		character !== "{" && character !== "}" && character !== "[" && character !== "]" &&
		character !== "," && character !== ":" && character !== "=";
}

function isHorizontalWhitespace(character: string | undefined): boolean {
	return character === " " || character === "\t";
}

function isWhitespace(character: string | undefined): boolean {
	return isHorizontalWhitespace(character) || character === "\n" || character === "\r" || character === "\f";
}

function containsWhitespace(value: string): boolean {
	for (const character of value) if (isWhitespace(character)) return true;
	return false;
}

function scanQuotedValueEnd(
	value: string,
	start: number,
	quote: string,
	lineEnd: number,
	context: SensitiveAssignmentContext,
	metrics?: RedactionScanMetrics,
): number | undefined {
	for (let index = start; index < lineEnd; index += 1) {
		recordTotalVisits(metrics, 1);
		if (quote === '"' && value[index] === "\\") {
			index += 1;
			recordTotalVisits(metrics, 1);
			continue;
		}
		if (quote === "'" && value[index] === "'" && value[index + 1] === "'") {
			index += 1;
			recordTotalVisits(metrics, 1);
			continue;
		}
		if (value[index] !== quote) continue;
		let boundary = index + 1;
		while (boundary < lineEnd && isHorizontalWhitespace(value[boundary])) {
			boundary += 1;
			recordTotalVisits(metrics, 1);
		}
		if (boundary >= lineEnd || value[boundary] === "#" ||
			(context === "flow" && (value[boundary] === "," || value[boundary] === ";" ||
				value[boundary] === "}" || value[boundary] === "]"))) {
			return index;
		}
	}
	return undefined;
}

function scanPublicQuotedScalarEnd(
	value: string,
	start: number,
	quote: LexicalQuote,
	metrics?: RedactionScanMetrics,
): number | undefined {
	for (let index = start; index < value.length; index += 1) {
		recordTotalVisits(metrics, 1);
		if (quote === '"' && value[index] === "\\") {
			index += 1;
			recordTotalVisits(metrics, 1);
			continue;
		}
		if (quote === "'" && value[index] === "'" && value[index + 1] === "'") {
			index += 1;
			recordTotalVisits(metrics, 1);
			continue;
		}
		if (value[index] !== quote) continue;
		let boundary = index + 1;
		while (isHorizontalWhitespace(value[boundary])) {
			boundary += 1;
			recordTotalVisits(metrics, 1);
		}
		if (boundary >= value.length || value[boundary] === "\r" || value[boundary] === "\n" ||
			value[boundary] === "," || value[boundary] === ";" || value[boundary] === "}" ||
			value[boundary] === "]" || value[boundary] === "#") {
			return index;
		}
	}
	return undefined;
}

function findLineEnd(value: string, start: number, metrics?: RedactionScanMetrics): number {
	let index = start;
	while (index < value.length && value[index] !== "\n" && value[index] !== "\r") {
		if (metrics) {
			metrics.lineBoundaryVisits += 1;
			metrics.boundaryCharacterVisits += 1;
			recordTotalVisits(metrics, 1);
		}
		index += 1;
	}
	return index;
}

function afterLineEnding(value: string, lineEnd: number): number {
	if (value[lineEnd] === "\r" && value[lineEnd + 1] === "\n") return lineEnd + 2;
	if (value[lineEnd] === "\r" || value[lineEnd] === "\n") return lineEnd + 1;
	return lineEnd;
}

function leadingIndentLength(value: string): number {
	let index = 0;
	while (isHorizontalWhitespace(value[index])) index += 1;
	return index;
}

function trailingLineEnding(value: string): string {
	if (value.endsWith("\r\n")) return "\r\n";
	if (value.endsWith("\n")) return "\n";
	if (value.endsWith("\r")) return "\r";
	return "";
}

function isYamlBlockHeader(value: string): boolean {
	return /^[|>](?:(?:[1-9][+-]?)|(?:[+-][1-9]?))?[ \t]*(?:#.*)?$/.test(value);
}

function findUnquotedValueEnd(
	value: string,
	start: number,
	lineEnd: number,
	context: SensitiveAssignmentContext,
	metrics?: RedactionScanMetrics,
): number {
	const nestedClosers: FlowCloser[] = [];
	let quote: LexicalQuote | undefined;
	let index = start;
	let currentLineStart = start;
	let currentLineEnd = lineEnd;
	while (index < value.length) {
		if (index >= currentLineEnd) {
			if (nestedClosers.length === 0 && context === "line") return currentLineEnd;
			const nextLineStart = afterLineEnding(value, currentLineEnd);
			if (nextLineStart <= currentLineEnd) return currentLineEnd;
			index = nextLineStart;
			currentLineStart = nextLineStart;
			currentLineEnd = findLineEnd(value, nextLineStart, metrics);
			quote = undefined;
			continue;
		}
		const character = value[index];
		recordTotalVisits(metrics, 1);
		if (quote) {
			if (quote === '"' && character === "\\") {
				index += 2;
				recordTotalVisits(metrics, 1);
				continue;
			}
			if (quote === "'" && character === "'" && value[index + 1] === "'") {
				index += 2;
				recordTotalVisits(metrics, 1);
				continue;
			}
			if (character === quote) quote = undefined;
			index += 1;
			continue;
		}
		if (isCommentStart(value, index, currentLineStart)) {
			if (nestedClosers.length === 0) return index;
			index = currentLineEnd;
			continue;
		}
		if ((character === '"' || character === "'") && isQuotedSegmentStart(value, index, currentLineStart)) {
			quote = character;
			index += 1;
			continue;
		}
		if (character === "{") {
			nestedClosers.push("}");
			index += 1;
			continue;
		}
		if (character === "[") {
			nestedClosers.push("]");
			index += 1;
			continue;
		}
		if (character === "}" || character === "]") {
			if (nestedClosers.length === 0) return index;
			if (nestedClosers[nestedClosers.length - 1] !== character) return index;
			nestedClosers.pop();
			index += 1;
			continue;
		}
		if (context === "flow" && nestedClosers.length === 0 && (character === "," || character === ";")) return index;
		index += 1;
	}
	return currentLineEnd;
}

function trimHorizontalEnd(value: string, start: number, end: number): number {
	while (end > start && isHorizontalWhitespace(value[end - 1])) end -= 1;
	return end;
}

function isPublicScalar(value: string): boolean {
	return /^(?:true|false|null|~)$/i.test(value);
}

function advanceScannerIndex(
	state: StructuredScannerState,
	nextIndex: number,
	metrics?: RedactionScanMetrics,
): void {
	const difference = nextIndex - state.index;
	if (metrics) {
		if (difference < 0) metrics.cursorRegressions += 1;
		else metrics.cursorAdvances += difference;
	}
	recordTotalVisits(metrics, Math.max(0, difference));
	state.index = nextIndex;
}

function initializeRedactionMetrics(metrics: RedactionScanMetrics, sourceLength: number): void {
	metrics.sourceLength = sourceLength;
	metrics.cursorAdvances = 0;
	metrics.cursorRegressions = 0;
	metrics.maxMainCursorVisits = sourceLength > 0 ? 1 : 0;
	metrics.keyCharacterVisits = 0;
	metrics.boundaryCharacterVisits = 0;
	metrics.totalWork = 0;
	metrics.lineBoundaryVisits = 0;
	metrics.keyStartVisits = 0;
	metrics.totalCharacterVisits = 0;
}

function recordKeyCharacterVisit(metrics: RedactionScanMetrics | undefined, count = 1): void {
	if (!metrics || count <= 0) return;
	metrics.keyCharacterVisits += count;
}

function recordKeyStartVisit(metrics: RedactionScanMetrics | undefined, count = 1): void {
	if (!metrics || count <= 0) return;
	metrics.keyStartVisits = (metrics.keyStartVisits ?? 0) + count;
	recordTotalVisits(metrics, count);
}

function recordTotalVisits(metrics: RedactionScanMetrics | undefined, count: number): void {
	if (!metrics || count <= 0) return;
	metrics.totalWork += count;
	metrics.totalCharacterVisits = (metrics.totalCharacterVisits ?? 0) + count;
}

export function normalizeScopedPrefixes(prefixes: unknown, description: string): string[] {
	const capturedPrefixes = capturePolicyArray(prefixes, `${description} prefixes`, 64, false);
	const normalized: string[] = [];
	for (const prefix of capturedPrefixes) {
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
	Object.freeze(normalized);
	return normalized;
}

function capturePolicyArray<T>(
	value: unknown,
	description: string,
	maximum: number,
	allowEmpty: boolean,
): T[] {
	if (!Array.isArray(value) || nodeTypes.isProxy(value)) {
		throw new ToolPolicyError(`${description} must be a bounded non-proxy array`);
	}
	const lengthDescriptor = Reflect.getOwnPropertyDescriptor(value, "length");
	const length = lengthDescriptor && "value" in lengthDescriptor ? lengthDescriptor.value : undefined;
	if (!lengthDescriptor || lengthDescriptor.get || lengthDescriptor.set || !("value" in lengthDescriptor) ||
		typeof length !== "number" || !Number.isSafeInteger(length) || length < (allowEmpty ? 0 : 1) ||
		length > maximum) {
		throw new ToolPolicyError(`${description} has an invalid authoritative length`);
	}
	const captured: T[] = [];
	for (let index = 0; index < length; index += 1) {
		const descriptor = Reflect.getOwnPropertyDescriptor(value, String(index));
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
			throw new ToolPolicyError(`${description} contains a sparse or accessor element`);
		}
		captured[index] = descriptor.value as T;
	}
	Object.freeze(captured);
	return captured;
}

function captureCapabilityResult(name: string, result: CapabilityResult): Readonly<Required<CapabilityResult>> {
	if (!result || typeof result !== "object" || nodeTypes.isProxy(result)) {
		throw new ToolPolicyError(`${name} returned an invalid result`);
	}
	const fields = captureOwnResultFields(result, ["status", "summary", "references"], `${name} result`);
	const status = fields.get("status");
	const summarySource = fields.get("summary");
	const referencesSource = fields.get("references");
	if (typeof status !== "string" || !["ok", "blocked", "failed"].includes(status)) {
		throw new ToolPolicyError(`${name} returned an invalid status`);
	}
	const summary = boundedString(summarySource, `${name} summary`, MAX_CAPABILITY_SUMMARY_CHARACTERS, false);
	const references: string[] = [];
	if (referencesSource !== undefined) {
		if (!Array.isArray(referencesSource) || nodeTypes.isProxy(referencesSource) || referencesSource.length > MAX_REFERENCES) {
			throw new ToolPolicyError(`${name} returned too many references`);
		}
		for (let index = 0; index < referencesSource.length; index += 1) {
			const descriptor = Object.getOwnPropertyDescriptor(referencesSource, String(index));
			if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
				throw new ToolPolicyError(`${name} returned sparse or accessor references`);
			}
			const reference = descriptor.value;
			references.push(boundedString(reference, `${name} reference`, MAX_REFERENCE_CHARACTERS, false));
		}
	}
	Object.freeze(references);
	return Object.freeze({ status: status as CapabilityResult["status"], summary, references });
}

function mutationResult(result: WorkspaceMutationResult, maxBytes: number): SessionToolResult {
	if (!result || typeof result !== "object" || nodeTypes.isProxy(result)) {
		throw new ToolPolicyError("workspace mutation returned an invalid result");
	}
	const fields = captureOwnResultFields(result, ["changed", "summary"], "workspace mutation result");
	const changed = fields.get("changed");
	const summarySource = fields.get("summary");
	if (typeof changed !== "boolean") throw new ToolPolicyError("workspace mutation returned an invalid result");
	const summary = boundedString(summarySource, "workspace mutation summary", MAX_CAPABILITY_SUMMARY_CHARACTERS, false);
	return textResult(JSON.stringify({ changed, summary: redactSensitiveText(summary) }), maxBytes);
}

function captureOwnResultFields(
	value: object,
	allowed: readonly string[],
	description: string,
): ReadonlyMap<string, unknown> {
	const allowedSet = new Set(allowed);
	const fields = new Map<string, unknown>();
	let enumerableFields = 0;
	for (const key in value) {
		if (!Object.hasOwn(value, key) || !allowedSet.has(key)) {
			throw new ToolPolicyError(`${description} contains an unknown field`);
		}
		enumerableFields += 1;
		if (enumerableFields > allowed.length) throw new ToolPolicyError(`${description} contains unknown fields`);
	}
	// Fixed host envelopes are projected through allowlisted descriptors. Hidden and
	// symbol peers are deliberately discarded without materializing the source key set.
	for (const key of allowed) {
		const descriptor = Object.getOwnPropertyDescriptor(value, key);
		if (!descriptor) continue;
		if (!descriptor?.enumerable || descriptor.set) {
			throw new ToolPolicyError(`${description} contains an invalid field`);
		}
		if ("value" in descriptor) fields.set(key, descriptor.value);
		else if (descriptor.get) fields.set(key, Reflect.apply(descriptor.get, value, []));
		else throw new ToolPolicyError(`${description} contains an unreadable field`);
	}
	return fields;
}

function sanitizedToolBoundaryError(operation: string, error: unknown): ToolPolicyError {
	let source = "external operation failed";
	try {
		if (error instanceof Error && typeof error.message === "string" && error.message.length > 0) {
			source = error.message.slice(0, 4_096);
		}
	} catch {
		// Hostile error accessors are reduced to the stable fallback.
	}
	const safe = redactSensitiveText(source)
		.replace(/[\u0000-\u001f\u007f-\u009f\u061c\u200e\u200f\u2028-\u202e\u2066-\u2069]/g, " ")
		.slice(0, 2_048);
	const cause = new Error(safe || "external operation failed");
	return new ToolPolicyError(`${operation} failed: ${safe || "external operation failed"}`, { cause });
}

async function toolBoundary(operation: string, execute: () => Promise<SessionToolResult>): Promise<SessionToolResult> {
	try {
		return await execute();
	} catch (error) {
		throw sanitizedToolBoundaryError(operation, error);
	}
}

function textResult(value: string, maxBytes: number): SessionToolResult {
	if (byteLength(value) > maxBytes) throw new ToolPolicyError("tool output exceeded the bounded output limit");
	const content = [{ type: "text" as const, text: value }];
	Object.freeze(content[0]);
	Object.freeze(content);
	return Object.freeze({ content, details: null });
}

function closedObject(properties: Record<string, unknown>, required: string[]): Readonly<Record<string, unknown>> {
	for (const value of Object.values(properties)) {
		if (value && typeof value === "object") Object.freeze(value);
	}
	Object.freeze(properties);
	Object.freeze(required);
	return Object.freeze({ type: "object", additionalProperties: false, properties, required });
}

function assertOnlyKeys(value: Readonly<Record<string, unknown>>, allowed: readonly string[], description: string): void {
	if (!isRecord(value)) throw new ToolPolicyError(`${description} input must be an object`);
	const allowedSet = new Set(allowed);
	let count = 0;
	for (const key in value) {
		if (!Object.hasOwn(value, key) || !allowedSet.has(key)) {
			throw new ToolPolicyError(`${description} input contains unknown field ${JSON.stringify(key)}`);
		}
		count += 1;
		if (count > allowed.length) throw new ToolPolicyError(`${description} input contains too many fields`);
	}
}

function recordParams(value: unknown, description: string): Readonly<Record<string, unknown>> {
	const snapshot = snapshotJsonData(
		value,
		0,
		{ nodes: 0, keys: 0, bytes: 0, maximumBytes: MAX_TOOL_INPUT_BYTES },
		new WeakSet<object>(),
		`${description} input`,
	);
	if (!isRecord(snapshot)) throw new ToolPolicyError(`${description} input must be an object`);
	return snapshot;
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

function boundedPositiveInteger(value: number, name: string, maximum: number): number {
	if (!Number.isSafeInteger(value) || value <= 0 || value > maximum) {
		throw new ToolPolicyError(`${name} must be a positive safe integer within the embedded maximum ${maximum}`);
	}
	return value;
}

function assertSignal(signal: AbortSignal | undefined): void {
	if (signal === undefined) return;
	if (!(signal instanceof AbortSignal) || Object.hasOwn(signal, "aborted") || typeof NATIVE_ABORTED_GETTER !== "function") {
		throw new ToolPolicyError("tool execution signal is invalid");
	}
	if (Reflect.apply(NATIVE_ABORTED_GETTER, signal, [])) throw new ToolPolicyError("tool execution was cancelled");
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
