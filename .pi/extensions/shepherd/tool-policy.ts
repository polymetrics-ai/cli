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
const INTRINSIC_OBJECT_PROTOTYPE = Object.prototype;
const INTRINSIC_ARRAY_PROTOTYPE = Array.prototype;
const INTRINSIC_OBJECT_KEYS = Object.keys;
const INTRINSIC_GET_PROTOTYPE_OF = Object.getPrototypeOf;
const INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR = Object.getOwnPropertyDescriptor;
const INTRINSIC_DEFINE_PROPERTY = Object.defineProperty;
const INTRINSIC_OBJECT_CREATE = Object.create;
const INTRINSIC_OBJECT_FREEZE = Object.freeze;
const INTRINSIC_OBJECT_HAS_OWN = Object.hasOwn;
const INTRINSIC_OBJECT_VALUES = Object.values;
const INTRINSIC_IS_ARRAY = Array.isArray;
const INTRINSIC_IS_PROXY = nodeTypes.isProxy;
const INTRINSIC_IS_NATIVE_ERROR = nodeTypes.isNativeError;
const INTRINSIC_REFLECT_APPLY = Reflect.apply;
const INTRINSIC_NUMBER = Number;
const INTRINSIC_NUMBER_IS_FINITE = Number.isFinite;
const INTRINSIC_NUMBER_IS_INTEGER = Number.isInteger;
const INTRINSIC_NUMBER_IS_SAFE_INTEGER = Number.isSafeInteger;
const INTRINSIC_NUMBER_PARSE_INT = Number.parseInt;
const INTRINSIC_NUMBER_MIN_SAFE_INTEGER = Number.MIN_SAFE_INTEGER;
const INTRINSIC_NUMBER_MAX_SAFE_INTEGER = Number.MAX_SAFE_INTEGER;
const INTRINSIC_STRING = String;
const INTRINSIC_JSON_STRINGIFY = JSON.stringify;
const INTRINSIC_STRING_CHAR_CODE_AT = String.prototype.charCodeAt;
const INTRINSIC_STRING_CODE_POINT_AT = String.prototype.codePointAt;
const INTRINSIC_STRING_TRIM = String.prototype.trim;
const INTRINSIC_STRING_SLICE = String.prototype.slice;
const INTRINSIC_STRING_REPLACE = String.prototype.replace;
const INTRINSIC_STRING_STARTS_WITH = String.prototype.startsWith;
const INTRINSIC_STRING_ENDS_WITH = String.prototype.endsWith;
const INTRINSIC_STRING_INCLUDES = String.prototype.includes;
const INTRINSIC_STRING_SPLIT = String.prototype.split;
const INTRINSIC_STRING_TO_LOWER_CASE = String.prototype.toLowerCase;
const INTRINSIC_STRING_FROM_CHAR_CODE = String.fromCharCode;
const INTRINSIC_ARRAY_PUSH = Array.prototype.push;
const INTRINSIC_ARRAY_POP = Array.prototype.pop;
const INTRINSIC_ARRAY_AT = Array.prototype.at;
const INTRINSIC_ARRAY_SPLICE = Array.prototype.splice;
const INTRINSIC_ARRAY_JOIN = Array.prototype.join;
const INTRINSIC_ARRAY_INCLUDES = Array.prototype.includes;
const INTRINSIC_ARRAY_LAST_INDEX_OF = Array.prototype.lastIndexOf;
const INTRINSIC_MATH_MIN = Math.min;
const INTRINSIC_MATH_MAX = Math.max;
const INTRINSIC_ERROR = Error;
const INTRINSIC_WEAK_SET = WeakSet;
const INTRINSIC_WEAK_SET_HAS = WeakSet.prototype.has;
const INTRINSIC_WEAK_SET_ADD = WeakSet.prototype.add;
const INTRINSIC_WEAK_SET_DELETE = WeakSet.prototype.delete;
const NATIVE_ABORTED_GETTER = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(AbortSignal.prototype, "aborted")?.get;

function intrinsicString(value: unknown): string {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING, undefined, [value]) as string;
}

function intrinsicMin(left: number, right: number): number {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_MATH_MIN, undefined, [left, right]) as number;
}

function intrinsicMax(left: number, right: number): number {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_MATH_MAX, undefined, [left, right]) as number;
}

function arrayPush<T>(target: T[], value: T): number {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_ARRAY_PUSH, target, [value]) as number;
}

function stringSlice(value: string, start: number, end?: number): string {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_SLICE, value, end === undefined ? [start] : [start, end]) as string;
}

function stringReplace(value: string, pattern: string | RegExp, replacement: string): string {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_REPLACE, value, [pattern, replacement]) as string;
}

function stringStartsWith(value: string, search: string): boolean {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_STARTS_WITH, value, [search]) as boolean;
}

function stringEndsWith(value: string, search: string): boolean {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_ENDS_WITH, value, [search]) as boolean;
}

function stringIncludes(value: string, search: string): boolean {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_INCLUDES, value, [search]) as boolean;
}

function stringTrim(value: string): string {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_TRIM, value, []) as string;
}

function stringToLowerCase(value: string): string {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_TO_LOWER_CASE, value, []) as string;
}

function stringSplit(value: string, separator: string): string[] {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_SPLIT, value, [separator]) as string[];
}

function stringFromCharCode(value: number): string {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_FROM_CHAR_CODE, undefined, [value]) as string;
}

function arrayAt<T>(value: readonly T[], index: number): T | undefined {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_ARRAY_AT, value, [index]) as T | undefined;
}

function arrayPop<T>(value: T[]): T | undefined {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_ARRAY_POP, value, []) as T | undefined;
}

function arraySplice<T>(value: T[], start: number, deleteCount: number): T[] {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_ARRAY_SPLICE, value, [start, deleteCount]) as T[];
}

function arrayJoin(value: readonly string[], separator: string): string {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_ARRAY_JOIN, value, [separator]) as string;
}

function arrayIncludes<T>(value: readonly T[], candidate: T): boolean {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_ARRAY_INCLUDES, value, [candidate]) as boolean;
}

function arrayLastIndexOf<T>(value: readonly T[], candidate: T): number {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_ARRAY_LAST_INDEX_OF, value, [candidate]) as number;
}

/** The complete reviewed host-tool domain. Unknown strings have no extension path. */
export const HOST_CAPABILITY_REGISTRY = INTRINSIC_OBJECT_FREEZE({
	host_inspect: INTRINSIC_OBJECT_FREEZE({ mutates: false as const }),
	host_verify: INTRINSIC_OBJECT_FREEZE({ mutates: true as const }),
} as const);

export type HostCapabilityName = keyof typeof HOST_CAPABILITY_REGISTRY;

const WORKSPACE_TOOL_REGISTRY = INTRINSIC_OBJECT_FREEZE({
	workspace_read: INTRINSIC_OBJECT_FREEZE({ mutates: false as const }),
	workspace_edit: INTRINSIC_OBJECT_FREEZE({ mutates: true as const }),
	workspace_write: INTRINSIC_OBJECT_FREEZE({ mutates: true as const }),
} as const);

export type WorkspaceToolName = keyof typeof WORKSPACE_TOOL_REGISTRY;
export type SessionToolName = WorkspaceToolName | HostCapabilityName;

export function isHostCapabilityName(value: unknown): value is HostCapabilityName {
	return typeof value === "string" && INTRINSIC_OBJECT_HAS_OWN(HOST_CAPABILITY_REGISTRY, value);
}

export function isSessionToolName(value: unknown): value is SessionToolName {
	return typeof value === "string" &&
		(INTRINSIC_OBJECT_HAS_OWN(WORKSPACE_TOOL_REGISTRY, value) || INTRINSIC_OBJECT_HAS_OWN(HOST_CAPABILITY_REGISTRY, value));
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

function matchesSensitivePathPattern(value: string): boolean {
	for (const pattern of sensitivePathPatterns) {
		if (pattern.test(value)) return true;
	}
	return false;
}

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
	projectArguments(
		name: string,
		raw: unknown,
		maximumBytes?: number,
	): Readonly<Record<string, unknown>>;
}

export interface ToolPolicyInput {
	readOnly: boolean;
	/** False grants only explicitly declared host mutation capabilities, never workspace edit/write tools. */
	workspaceMutation?: boolean;
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

function captureFixedOwnDataFields(
	value: unknown,
	fields: readonly string[],
	description: string,
	optional: ReadonlySet<string> = new Set<string>(),
): Readonly<Record<string, unknown>> {
	if (!value || typeof value !== "object" || INTRINSIC_IS_PROXY(value)) {
		throw new ToolPolicyError(`${description} must be a non-proxy record`);
	}
	const prototype = INTRINSIC_GET_PROTOTYPE_OF(value);
	if (prototype !== INTRINSIC_OBJECT_PROTOTYPE && prototype !== null) {
		throw new ToolPolicyError(`${description} must use an exact approved prototype`);
	}
	const captured = INTRINSIC_OBJECT_CREATE(null) as Record<string, unknown>;
	for (const field of fields) {
		const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, field);
		if (!descriptor) {
			if (optional.has(field)) continue;
			throw new ToolPolicyError(`${description} requires own field ${field}`);
		}
		if (!descriptor.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
			throw new ToolPolicyError(`${description}.${field} must be an own enumerable data field`);
		}
		INTRINSIC_DEFINE_PROPERTY(captured, field, {
			value: descriptor.value,
			enumerable: true,
			writable: false,
			configurable: false,
		});
	}
	return INTRINSIC_OBJECT_FREEZE(captured);
}

function captureToolPolicyOptions(options: ToolPolicyOptions): ToolPolicyOptions {
	const fields = ["maxToolOutputBytes", "maxReadCharacters", "maxWriteCharacters"] as const;
	return captureFixedOwnDataFields(
		options,
		fields,
		"tool policy options",
		new Set(fields),
	) as ToolPolicyOptions;
}

function captureToolPolicyInput(input: ToolPolicyInput): ToolPolicyInput {
	const root = captureFixedOwnDataFields(
		input,
		["readOnly", "workspaceMutation", "workspace", "authority", "capabilities"],
		"tool policy input",
		new Set(["workspaceMutation"]),
	);
	const readOnly = root.readOnly;
	if (typeof readOnly !== "boolean") throw new ToolPolicyError("readOnly must be boolean");
	const workspaceMutation = root.workspaceMutation === undefined ? !readOnly : root.workspaceMutation;
	if (typeof workspaceMutation !== "boolean" || (readOnly && workspaceMutation)) {
		throw new ToolPolicyError("workspaceMutation conflicts with readOnly authority");
	}

	const authoritySource = captureFixedOwnDataFields(
		root.authority,
		["workspaceId", "readPrefixes", "writePrefixes", "capabilityNames"],
		"tool authority",
	);
	const authority = INTRINSIC_OBJECT_FREEZE({
		workspaceId: authoritySource.workspaceId,
		readPrefixes: capturePolicyArray<string>(authoritySource.readPrefixes, "read prefixes", 64, false),
		writePrefixes: capturePolicyArray<string>(authoritySource.writePrefixes, "write prefixes", 64, readOnly),
		capabilityNames: capturePolicyArray<HostCapabilityName>(
			authoritySource.capabilityNames,
			"capability authority",
			MAX_CAPABILITIES,
			true,
		),
	}) as ToolAuthority;

	const workspaceSource = root.workspace;
	const workspaceFields = captureFixedOwnDataFields(
		workspaceSource,
		["id", "cwd", "readText", "editText", "writeText"],
		"workspace capability",
	);
	const readText = workspaceFields.readText;
	const editText = workspaceFields.editText;
	const writeText = workspaceFields.writeText;
	if (typeof readText !== "function" || typeof editText !== "function" || typeof writeText !== "function") {
		throw new ToolPolicyError("workspace capability has an invalid method contract");
	}
	const workspace = {
		id: workspaceFields.id as string,
		cwd: workspaceFields.cwd as string,
		readText(path: string, methodOptions: { offset?: number; limit?: number; signal?: AbortSignal }) {
			return INTRINSIC_REFLECT_APPLY(readText, workspaceSource, [path, methodOptions]) as Promise<string>;
		},
		editText(path: string, oldText: string, newText: string, signal?: AbortSignal) {
			return INTRINSIC_REFLECT_APPLY(editText, workspaceSource, [path, oldText, newText, signal]) as
				Promise<WorkspaceMutationResult>;
		},
		writeText(path: string, content: string, signal?: AbortSignal) {
			return INTRINSIC_REFLECT_APPLY(writeText, workspaceSource, [path, content, signal]) as
				Promise<WorkspaceMutationResult>;
		},
	} satisfies ScopedWorkspace;
	INTRINSIC_OBJECT_FREEZE(workspace);

	const capabilitySources = capturePolicyArray<HostCapability>(
		root.capabilities,
		"typed host capabilities",
		MAX_CAPABILITIES,
		true,
	);
	const capabilities: HostCapability[] = [];
	for (let index = 0; index < capabilitySources.length; index += 1) {
		const capabilitySource = capabilitySources[index]!;
		const fields = captureFixedOwnDataFields(
			capabilitySource,
			["name", "description", "mutates", "parameters", "execute"],
			"typed host capability",
		);
		const execute = fields.execute;
		if (typeof execute !== "function") throw new ToolPolicyError("typed host capability has an invalid execute method");
		const capturedCapability = {
			name: fields.name,
			description: fields.description,
			mutates: fields.mutates,
			parameters: fields.parameters,
			execute(arguments_: Readonly<Record<string, unknown>>, signal?: AbortSignal) {
				return INTRINSIC_REFLECT_APPLY(execute, capabilitySource, [arguments_, signal]);
			},
		} as unknown as HostCapability;
		capabilities[index] = INTRINSIC_OBJECT_FREEZE(capturedCapability);
	}
	INTRINSIC_OBJECT_FREEZE(capabilities);

	return INTRINSIC_OBJECT_FREEZE({ readOnly, workspaceMutation, workspace, authority, capabilities });
}

export function createToolPolicy(input: ToolPolicyInput, options: ToolPolicyOptions = {}): ToolPolicy {
	const capturedOptions = captureToolPolicyOptions(options);
	const limits = {
		maxToolOutputBytes: boundedPositiveInteger(
			capturedOptions.maxToolOutputBytes ?? DEFAULT_MAX_TOOL_OUTPUT_BYTES,
			"maxToolOutputBytes",
			MAX_TOOL_OUTPUT_BYTES,
		),
		maxReadCharacters: boundedPositiveInteger(
			capturedOptions.maxReadCharacters ?? DEFAULT_MAX_READ_CHARACTERS,
			"maxReadCharacters",
			MAX_READ_CHARACTERS,
		),
		maxWriteCharacters: boundedPositiveInteger(
			capturedOptions.maxWriteCharacters ?? DEFAULT_MAX_WRITE_CHARACTERS,
			"maxWriteCharacters",
			MAX_WRITE_CHARACTERS,
		),
	};
	const capturedInput = captureToolPolicyInput(input);
	const capabilities = validatePolicyInput(capturedInput, limits.maxWriteCharacters);

	const readPrefixes = normalizeScopedPrefixes(capturedInput.authority.readPrefixes, "read");
	const writePrefixes = capturedInput.readOnly && capturedInput.authority.writePrefixes.length === 0
		? capturedInput.authority.writePrefixes
		: normalizeScopedPrefixes(capturedInput.authority.writePrefixes, "write");
	const tools: SessionTool[] = [workspaceReadTool(capturedInput.workspace, readPrefixes, limits)];
	if (!capturedInput.readOnly && capturedInput.workspaceMutation !== false) {
		arrayPush(tools, workspaceEditTool(capturedInput.workspace, writePrefixes, limits));
		arrayPush(tools, workspaceWriteTool(capturedInput.workspace, writePrefixes, limits));
	}

	const declared = new Set(capturedInput.authority.capabilityNames);
	const hostContracts = new Map<HostCapabilityName, CompiledToolArgumentContract>();
	for (const capability of capabilities) {
		if (!declared.has(capability.name)) {
			throw new ToolPolicyError(`undeclared capability ${INTRINSIC_JSON_STRINGIFY(capability.name)} cannot expand authority`);
		}
		if (capturedInput.readOnly && capability.mutates) continue;
		arrayPush(tools, hostCapabilityTool(capability, limits));
		hostContracts.set(capability.name, capability.argumentContract);
	}

	for (const tool of tools) INTRINSIC_OBJECT_FREEZE(tool);
	const names: SessionToolName[] = [];
	for (let index = 0; index < tools.length; index += 1) {
		const tool = tools[index]!;
		if (!isSessionToolName(tool.name)) throw new ToolPolicyError("tool policy constructed an unregistered tool identity");
		names[index] = tool.name;
	}
	INTRINSIC_OBJECT_FREEZE(names);
	INTRINSIC_OBJECT_FREEZE(tools);
	return INTRINSIC_OBJECT_FREEZE({
		names,
		tools,
		projectArguments(name: string, raw: unknown, maximumBytes = MAX_TOOL_INPUT_BYTES) {
			if (!isSessionToolName(name)) throw new ToolPolicyError("tool arguments use an unregistered schema");
			if (isHostCapabilityName(name)) {
				const contract = hostContracts.get(name);
				if (!contract) throw new ToolPolicyError(`tool arguments use unavailable capability ${name}`);
				return contract.project(raw, maximumBytes);
			}
			return projectWorkspaceArguments(name, raw, maximumBytes);
		},
	});
}

interface CompiledHostCapability extends HostCapabilityContract<HostCapabilityName> {
	readonly argumentContract: CompiledToolArgumentContract;
}

interface CompiledToolArgumentContract {
	readonly schema: PlainJsonSchema;
	project(raw: unknown, maximumBytes: number): Readonly<Record<string, unknown>>;
}

function validatePolicyInput(input: ToolPolicyInput, maximumStringLength: number): CompiledHostCapability[] {
	if (!input || typeof input !== "object") throw new ToolPolicyError("tool policy input is required");
	if (typeof input.readOnly !== "boolean") throw new ToolPolicyError("readOnly must be boolean");
	if (input.workspaceMutation !== undefined && typeof input.workspaceMutation !== "boolean") {
		throw new ToolPolicyError("workspaceMutation must be boolean");
	}
	if (input.readOnly && input.workspaceMutation === true) {
		throw new ToolPolicyError("read-only authority cannot mutate the workspace");
	}
	if (!input.workspace || typeof input.workspace !== "object") throw new ToolPolicyError("workspace capability is required");
	if (!validIdentifier(input.workspace.id) || input.workspace.id !== input.authority?.workspaceId) {
		throw new ToolPolicyError("workspace identity does not match the immutable authority envelope");
	}
	for (const method of ["readText", "editText", "writeText"] as const) {
		if (typeof input.workspace[method] !== "function") {
			throw new ToolPolicyError(`workspace capability is missing ${method}`);
		}
	}
	if (!INTRINSIC_IS_ARRAY(input.capabilities) || input.capabilities.length > MAX_CAPABILITIES) {
		throw new ToolPolicyError("typed host capabilities must be a bounded array");
	}
	if (!INTRINSIC_IS_ARRAY(input.authority.capabilityNames) || input.authority.capabilityNames.length > MAX_CAPABILITIES) {
		throw new ToolPolicyError("capability authority must be a bounded array");
	}

	const declared = new Set<HostCapabilityName>();
	for (const name of input.authority.capabilityNames) {
		if (!isHostCapabilityName(name)) {
			throw new ToolPolicyError(`capability ${INTRINSIC_JSON_STRINGIFY(name)} is not in the closed host registry`);
		}
		if (declared.has(name)) throw new ToolPolicyError(`duplicate declared capability ${name}`);
		declared.add(name);
	}
	const supplied = new Set<HostCapabilityName>();
	const capabilities: CompiledHostCapability[] = [];
	for (const capability of input.capabilities) {
		if (!capability || typeof capability !== "object") throw new ToolPolicyError("capability must be an object");
		const name = capability.name;
		const description = capability.description;
		const mutates = capability.mutates;
		const parameterSource = capability.parameters;
		const execute = capability.execute;
		if (!isHostCapabilityName(name)) {
			throw new ToolPolicyError(`capability ${INTRINSIC_JSON_STRINGIFY(name)} is not in the closed host registry`);
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
		const parameterSnapshot = snapshotCapabilitySchema(parameterSource, name);
		const argumentContract = compileToolArgumentContract(parameterSnapshot, name, maximumStringLength);
		arrayPush(capabilities, INTRINSIC_OBJECT_FREEZE({
			name,
			description,
			mutates,
			parameters: argumentContract.schema,
			argumentContract,
			execute(input: Readonly<Record<string, unknown>>, signal?: AbortSignal) {
				return INTRINSIC_REFLECT_APPLY(execute, capability, [input, signal]);
			},
		}) as CompiledHostCapability);
	}
	for (const name of declared) {
		if (!supplied.has(name)) throw new ToolPolicyError(`declared capability ${name} was not supplied`);
	}
	return capabilities;
}

const MAX_PROJECTED_NODES = 4_096;
const MAX_PROJECTED_KEYS = 4_096;
const MAX_PROJECTED_ITEMS = 4_096;
const MAX_PROJECTED_DEPTH = 64;
const MAX_PROJECTED_CONTAINER_ENCOUNTERS = 4_096;

interface ProjectionBudget {
	nodes: number;
	keys: number;
	items: number;
	containers: number;
	scalarUnits: number;
	bytes: number;
	maximumBytes: number;
	seen: WeakSet<object>;
}

type CompiledSchemaProjector = (
	value: unknown,
	description: string,
	depth: number,
	budget: ProjectionBudget,
) => JsonData;

interface CompiledSchemaNode {
	readonly schema: PlainJsonSchema;
	readonly project: CompiledSchemaProjector;
}

function compileToolArgumentContract(
	schema: PlainJsonSchema,
	name: string,
	maximumStringLength: number,
): CompiledToolArgumentContract {
	const compiled = compileSchemaNode(schema, `${name} parameters`, 0, maximumStringLength);
	return INTRINSIC_OBJECT_FREEZE({
		schema: compiled.schema,
		project(raw: unknown, maximumBytes: number): Readonly<Record<string, unknown>> {
			if (!INTRINSIC_NUMBER_IS_SAFE_INTEGER(maximumBytes) || maximumBytes < 1) {
				throw new ToolPolicyError(`${name} input has an invalid byte bound`);
			}
			const effectiveMaximum = intrinsicMin(maximumBytes, MAX_TOOL_INPUT_BYTES);
			const projected = compiled.project(raw, `${name} input`, 0, {
				nodes: 0,
				keys: 0,
				items: 0,
				containers: 0,
				scalarUnits: 0,
				bytes: 0,
				maximumBytes: effectiveMaximum,
				seen: new INTRINSIC_WEAK_SET<object>(),
			});
			if (!isRecord(projected)) throw new ToolPolicyError(`${name} input must be an object`);
			return projected;
		},
	});
}

function compileSchemaNode(
	schema: PlainJsonSchema,
	description: string,
	depth: number,
	maximumStringLength: number,
): CompiledSchemaNode {
	if (depth > MAX_CAPABILITY_SCHEMA_DEPTH) throw new ToolPolicyError(`${description} exceeded its depth bound`);
	const type = ownSchemaField(schema, "type", description);
	if (type === "string") {
		assertSchemaVocabulary(schema, ["type"], ["minLength", "maxLength", "enum"], description);
		const minimum = optionalSchemaInteger(schema, "minLength", 0, maximumStringLength, description) ?? 0;
		const authoredMaximum = optionalSchemaInteger(schema, "maxLength", 0, MAX_WRITE_CHARACTERS, description);
		const maximum = intrinsicMin(authoredMaximum ?? maximumStringLength, maximumStringLength);
		const enumSource = optionalOwnSchemaField(schema, "enum", description);
		const enumValues = enumSource === undefined ? undefined : captureCompiledEnum(enumSource, "string", description);
		if (minimum > maximum) {
			throw new ToolPolicyError(`${description} has inverted string bounds`);
		}
		const entries: Array<readonly [string, JsonData]> = [
			["type", type], ["minLength", minimum], ["maxLength", maximum],
		];
		if (enumValues !== undefined) arrayPush(entries, ["enum", enumValues as JsonData[]]);
		return INTRINSIC_OBJECT_FREEZE({ schema: canonicalSchemaRecord(entries), project: (
			value: unknown,
			valueDescription: string,
			valueDepth: number,
			budget: ProjectionBudget,
		) => {
			chargeProjectionNode(budget, valueDepth, valueDescription);
			const normalized = normalizeSchemaString(value, valueDescription);
			const normalizedLength = chargeProjectedStringAndMeasure(budget, normalized, valueDescription);
			if (normalizedLength < minimum || normalizedLength > maximum || !enumAccepts(enumValues, normalized)) {
				throw new ToolPolicyError(`${valueDescription} violates its string schema`);
			}
			return normalized;
		} });
	}
	if (type === "integer" || type === "number") {
		assertSchemaVocabulary(schema, ["type"], ["minimum", "maximum", "enum"], description);
		const authoredMinimum = optionalSchemaFiniteNumber(schema, "minimum", description);
		const authoredMaximum = optionalSchemaFiniteNumber(schema, "maximum", description);
		if (type === "integer" &&
			((authoredMinimum !== undefined && !INTRINSIC_NUMBER_IS_INTEGER(authoredMinimum)) ||
				(authoredMaximum !== undefined && !INTRINSIC_NUMBER_IS_INTEGER(authoredMaximum)))) {
			throw new ToolPolicyError(`${description} integer bounds must be integral`);
		}
		const minimum = type === "integer"
			? intrinsicMax(authoredMinimum ?? INTRINSIC_NUMBER_MIN_SAFE_INTEGER, INTRINSIC_NUMBER_MIN_SAFE_INTEGER)
			: authoredMinimum;
		const maximum = type === "integer"
			? intrinsicMin(authoredMaximum ?? INTRINSIC_NUMBER_MAX_SAFE_INTEGER, INTRINSIC_NUMBER_MAX_SAFE_INTEGER)
			: authoredMaximum;
		const enumSource = optionalOwnSchemaField(schema, "enum", description);
		const enumValues = enumSource === undefined ? undefined : captureCompiledEnum(enumSource, type, description);
		if (minimum !== undefined && maximum !== undefined && minimum > maximum) {
			throw new ToolPolicyError(`${description} has inverted numeric bounds`);
		}
		const entries: Array<readonly [string, JsonData]> = [["type", type]];
		if (minimum !== undefined) arrayPush(entries, ["minimum", minimum]);
		if (maximum !== undefined) arrayPush(entries, ["maximum", maximum]);
		if (enumValues !== undefined) arrayPush(entries, ["enum", enumValues as JsonData[]]);
		return INTRINSIC_OBJECT_FREEZE({ schema: canonicalSchemaRecord(entries), project: (
			value: unknown,
			valueDescription: string,
			valueDepth: number,
			budget: ProjectionBudget,
		) => {
			chargeProjectionNode(budget, valueDepth, valueDescription);
			const normalized = normalizeSchemaNumber(value, type, valueDescription, budget);
			if ((type === "integer" && !INTRINSIC_NUMBER_IS_SAFE_INTEGER(normalized)) ||
				(minimum !== undefined && normalized < minimum) || (maximum !== undefined && normalized > maximum) ||
				!enumAccepts(enumValues, normalized)) {
				throw new ToolPolicyError(`${valueDescription} violates its numeric schema`);
			}
			chargeProjectedScalarBytes(budget, normalized, valueDescription);
			return normalized;
		} });
	}
	if (type === "boolean") {
		assertSchemaVocabulary(schema, ["type"], ["enum"], description);
		const enumSource = optionalOwnSchemaField(schema, "enum", description);
		const enumValues = enumSource === undefined ? undefined : captureCompiledEnum(enumSource, "boolean", description);
		const entries: Array<readonly [string, JsonData]> = [["type", type]];
		if (enumValues !== undefined) arrayPush(entries, ["enum", enumValues as JsonData[]]);
		return INTRINSIC_OBJECT_FREEZE({ schema: canonicalSchemaRecord(entries), project: (
			value: unknown,
			valueDescription: string,
			valueDepth: number,
			budget: ProjectionBudget,
		) => {
			chargeProjectionNode(budget, valueDepth, valueDescription);
			const normalized = normalizeSchemaBoolean(value, valueDescription);
			if (!enumAccepts(enumValues, normalized)) {
				throw new ToolPolicyError(`${valueDescription} violates its boolean schema`);
			}
			chargeProjectedScalarBytes(budget, normalized, valueDescription);
			return normalized;
		} });
	}
	if (type === "array") {
		assertSchemaVocabulary(schema, ["type", "items"], ["minItems", "maxItems"], description);
		const items = ownSchemaField(schema, "items", description);
		if (!isRecord(items)) throw new ToolPolicyError(`${description} requires an array item schema`);
		const item = compileSchemaNode(items, `${description} items`, depth + 1, maximumStringLength);
		const minimum = optionalSchemaInteger(schema, "minItems", 0, MAX_CAPABILITY_SCHEMA_ARRAY_ITEMS, description) ?? 0;
		const maximum = optionalSchemaInteger(schema, "maxItems", 0, MAX_CAPABILITY_SCHEMA_ARRAY_ITEMS, description) ??
			MAX_CAPABILITY_SCHEMA_ARRAY_ITEMS;
		if (minimum > maximum) throw new ToolPolicyError(`${description} has inverted array bounds`);
		const entries: Array<readonly [string, JsonData]> = [
			["type", type], ["items", item.schema as JsonData], ["minItems", minimum], ["maxItems", maximum],
		];
		return INTRINSIC_OBJECT_FREEZE({ schema: canonicalSchemaRecord(entries), project: (
			value: unknown,
			valueDescription: string,
			valueDepth: number,
			budget: ProjectionBudget,
		) => {
			chargeProjectionNode(budget, valueDepth, valueDescription);
			if (!INTRINSIC_IS_ARRAY(value) || INTRINSIC_IS_PROXY(value) ||
				INTRINSIC_GET_PROTOTYPE_OF(value) !== INTRINSIC_ARRAY_PROTOTYPE) {
				throw new ToolPolicyError(`${valueDescription} must be an exact array`);
			}
			chargeProjectionContainer(budget, value, valueDescription);
			const length = ownArrayLength(value, valueDescription);
			if (length < minimum || length > maximum) throw new ToolPolicyError(`${valueDescription} violates its array bounds`);
			chargeProjectionItems(budget, length, valueDescription);
			chargeProjectionBytes(budget, 2 + intrinsicMax(0, length - 1), valueDescription);
			const projected: JsonData[] = [];
			for (let index = 0; index < length; index += 1) {
				const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, intrinsicString(index));
				if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
					throw new ToolPolicyError(`${valueDescription} contains a sparse or accessor item`);
				}
				arrayPush(projected, item.project(descriptor.value, `${valueDescription}[${index}]`, valueDepth + 1, budget));
			}
			return INTRINSIC_OBJECT_FREEZE(projected) as JsonData[];
		} });
	}
	if (type !== "object") {
		throw new ToolPolicyError(`${description} uses an unsupported schema type`);
	}
	assertSchemaVocabulary(schema, ["type", "additionalProperties", "properties", "required"], [], description);
	if (ownSchemaField(schema, "additionalProperties", description) !== false) {
		throw new ToolPolicyError(`${description} must be a closed object schema`);
	}
	const properties = ownSchemaField(schema, "properties", description);
	const required = ownSchemaField(schema, "required", description);
	if (!isRecord(properties)) throw new ToolPolicyError(`${description} requires a properties record`);
	const propertyPrototype = INTRINSIC_GET_PROTOTYPE_OF(properties);
	if (INTRINSIC_IS_PROXY(properties) ||
		(propertyPrototype !== INTRINSIC_OBJECT_PROTOTYPE && propertyPrototype !== null)) {
		throw new ToolPolicyError(`${description} properties must use an exact approved prototype`);
	}
	if (!INTRINSIC_IS_ARRAY(required) || INTRINSIC_IS_PROXY(required) ||
		INTRINSIC_GET_PROTOTYPE_OF(required) !== INTRINSIC_ARRAY_PROTOTYPE) {
		throw new ToolPolicyError(`${description} required fields must be an exact array`);
	}
	const requiredLength = ownArrayLength(required, `${description} required fields`);
	if (requiredLength > 64) throw new ToolPolicyError(`${description} has too many required fields`);
	const requiredNames: string[] = [];
	const unique = new Set<string>();
	for (let index = 0; index < requiredLength; index += 1) {
		const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(required, intrinsicString(index));
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor) ||
			typeof descriptor.value !== "string" || descriptor.value.length < 1 || descriptor.value.length > 128 ||
			unique.has(descriptor.value)) {
			throw new ToolPolicyError(`${description} has an invalid required field vector`);
		}
		const propertyName = descriptor.value;
		unique.add(propertyName);
		arrayPush(requiredNames, propertyName);
	}
	const names = INTRINSIC_OBJECT_KEYS(properties);
	let propertySetMatches = names.length === requiredNames.length;
	for (let index = 0; propertySetMatches && index < names.length; index += 1) {
		propertySetMatches = unique.has(names[index]!);
	}
	if (!propertySetMatches) {
		throw new ToolPolicyError(`${description} cannot contain optional or undeclared properties`);
	}
	const projectors: CompiledSchemaProjector[] = [];
	const canonicalProperties = INTRINSIC_OBJECT_CREATE(null) as Record<string, JsonData>;
	for (const propertyName of names) {
		const propertyDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(properties, propertyName);
		if (!propertyDescriptor?.enumerable || propertyDescriptor.get || propertyDescriptor.set ||
			!("value" in propertyDescriptor) || !isRecord(propertyDescriptor.value)) {
			throw new ToolPolicyError(`${description} property ${INTRINSIC_JSON_STRINGIFY(propertyName)} is invalid`);
		}
		const property = compileSchemaNode(
			propertyDescriptor.value,
			`${description}.${propertyName}`,
			depth + 1,
			maximumStringLength,
		);
		arrayPush(projectors, property.project);
		INTRINSIC_DEFINE_PROPERTY(canonicalProperties, propertyName, {
			value: property.schema,
			enumerable: true,
			writable: false,
			configurable: false,
		});
	}
	INTRINSIC_OBJECT_FREEZE(requiredNames);
	INTRINSIC_OBJECT_FREEZE(names);
	INTRINSIC_OBJECT_FREEZE(projectors);
	INTRINSIC_OBJECT_FREEZE(canonicalProperties);
	const canonicalRequired = INTRINSIC_OBJECT_FREEZE([...requiredNames]);
	const canonicalSchema = canonicalSchemaRecord([
		["type", type],
		["additionalProperties", false],
		["properties", canonicalProperties],
		["required", canonicalRequired as JsonData[]],
	]);
	return INTRINSIC_OBJECT_FREEZE({ schema: canonicalSchema, project: (
		value: unknown,
		valueDescription: string,
		valueDepth: number,
		budget: ProjectionBudget,
	) => {
		chargeProjectionNode(budget, valueDepth, valueDescription);
		if (!isRecord(value) || INTRINSIC_IS_PROXY(value)) throw new ToolPolicyError(`${valueDescription} must be an object`);
		const prototype = INTRINSIC_GET_PROTOTYPE_OF(value);
		if (prototype !== INTRINSIC_OBJECT_PROTOTYPE && prototype !== null) {
			throw new ToolPolicyError(`${valueDescription} must use an exact approved prototype`);
		}
		chargeProjectionContainer(budget, value, valueDescription);
		assertBoundedClosedRecordKeys(value, unique, names.length, budget, valueDescription);
		chargeProjectionBytes(budget, 2 + intrinsicMax(0, names.length - 1), valueDescription);
		for (const field of names) {
			chargeProjectedStringBytes(budget, field, valueDescription);
			chargeProjectionBytes(budget, 1, valueDescription);
		}
		const projected = INTRINSIC_OBJECT_CREATE(null) as Record<string, JsonData>;
		for (let index = 0; index < names.length; index += 1) {
			const field = names[index]!;
			const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, field);
			if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
				throw new ToolPolicyError(`${valueDescription}.${field} must be an own data field`);
			}
			INTRINSIC_DEFINE_PROPERTY(projected, field, {
				value: projectors[index]!(descriptor.value, `${valueDescription}.${field}`, valueDepth + 1, budget),
				enumerable: true,
				writable: false,
				configurable: false,
			});
		}
		return INTRINSIC_OBJECT_FREEZE(projected);
	} });
}

function assertSchemaVocabulary(
	schema: PlainJsonSchema,
	required: readonly string[],
	optional: readonly string[],
	description: string,
): void {
	const allowed = new Set([...required, ...optional]);
	const keys = INTRINSIC_OBJECT_KEYS(schema);
	let supported = keys.length >= required.length;
	for (let index = 0; supported && index < keys.length; index += 1) {
		supported = allowed.has(keys[index]!);
	}
	if (!supported) {
		throw new ToolPolicyError(`${description} uses unsupported schema keywords`);
	}
	for (const field of required) ownSchemaField(schema, field, description);
}

function canonicalSchemaRecord(entries: readonly (readonly [string, JsonData])[]): PlainJsonSchema {
	const schema = INTRINSIC_OBJECT_CREATE(null) as Record<string, JsonData>;
	for (const [field, value] of entries) {
		INTRINSIC_DEFINE_PROPERTY(schema, field, {
			value,
			enumerable: true,
			writable: false,
			configurable: false,
		});
	}
	return INTRINSIC_OBJECT_FREEZE(schema);
}

function ownSchemaField(schema: PlainJsonSchema, field: string, description: string): unknown {
	const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(schema, field);
	if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
		throw new ToolPolicyError(`${description} requires own schema field ${field}`);
	}
	return descriptor.value;
}

function optionalOwnSchemaField(schema: PlainJsonSchema, field: string, description: string): unknown {
	const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(schema, field);
	if (!descriptor) return undefined;
	if (!descriptor.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
		throw new ToolPolicyError(`${description} has an invalid schema field ${field}`);
	}
	return descriptor.value;
}

function optionalSchemaInteger(
	schema: PlainJsonSchema,
	field: string,
	minimum: number,
	maximum: number,
	description: string,
): number | undefined {
	const value = optionalOwnSchemaField(schema, field, description);
	if (value === undefined) return undefined;
	if (!INTRINSIC_NUMBER_IS_SAFE_INTEGER(value) || (value as number) < minimum || (value as number) > maximum) {
		throw new ToolPolicyError(`${description} has an invalid ${field}`);
	}
	return value as number;
}

function optionalSchemaFiniteNumber(schema: PlainJsonSchema, field: string, description: string): number | undefined {
	const value = optionalOwnSchemaField(schema, field, description);
	if (value === undefined) return undefined;
	if (typeof value !== "number" || !INTRINSIC_NUMBER_IS_FINITE(value)) {
		throw new ToolPolicyError(`${description} has an invalid ${field}`);
	}
	return value === 0 ? 0 : value;
}

function captureCompiledEnum(
	source: unknown,
	type: "string" | "integer" | "number" | "boolean",
	description: string,
): readonly JsonData[] {
	if (!INTRINSIC_IS_ARRAY(source) || INTRINSIC_IS_PROXY(source) ||
		INTRINSIC_GET_PROTOTYPE_OF(source) !== INTRINSIC_ARRAY_PROTOTYPE) {
		throw new ToolPolicyError(`${description} enum must be an exact array`);
	}
	const length = ownArrayLength(source, `${description} enum`);
	if (length < 1 || length > MAX_CAPABILITY_SCHEMA_ARRAY_ITEMS) {
		throw new ToolPolicyError(`${description} enum exceeded its bound`);
	}
	const values: JsonData[] = [];
	for (let index = 0; index < length; index += 1) {
		const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(source, intrinsicString(index));
		const sourceValue = descriptor && "value" in descriptor ? descriptor.value : undefined;
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor) ||
			typeof sourceValue !== (type === "integer" || type === "number" ? "number" : type) ||
			(typeof sourceValue === "number" &&
				(!INTRINSIC_NUMBER_IS_FINITE(sourceValue) ||
					(type === "integer" && !INTRINSIC_NUMBER_IS_SAFE_INTEGER(sourceValue))))) {
			throw new ToolPolicyError(`${description} enum contains an unsupported value`);
		}
		const value = typeof sourceValue === "number" && sourceValue === 0 ? 0 : sourceValue as JsonData;
		for (const existing of values) {
			if (existing === value) throw new ToolPolicyError(`${description} enum contains a canonical duplicate`);
		}
		arrayPush(values, value);
	}
	return INTRINSIC_OBJECT_FREEZE(values);
}

function enumAccepts(values: readonly JsonData[] | undefined, value: JsonData): boolean {
	if (values === undefined) return true;
	for (const candidate of values) {
		if (candidate === value) return true;
	}
	return false;
}

function normalizeSchemaString(value: unknown, description: string): string {
	if (typeof value === "string") return value;
	if (value === null) return "";
	if (typeof value === "boolean") return intrinsicString(value);
	if (typeof value === "number" && INTRINSIC_NUMBER_IS_FINITE(value)) return intrinsicString(value);
	throw new ToolPolicyError(`${description} violates its string schema`);
}

function isTypeBoxCombiningMark(codePoint: number): boolean {
	return (codePoint >= 0x0300 && codePoint <= 0x036f) ||
		(codePoint >= 0x1ab0 && codePoint <= 0x1aff) ||
		(codePoint >= 0x1dc0 && codePoint <= 0x1dff) ||
		(codePoint >= 0xfe20 && codePoint <= 0xfe2f) ||
		(codePoint >= 0xfe00 && codePoint <= 0xfe0f);
}

function isHighSurrogate(code: number): boolean {
	return code >= 0xd800 && code <= 0xdbff;
}

function isLowSurrogate(code: number): boolean {
	return code >= 0xdc00 && code <= 0xdfff;
}

function codePointAt(value: string, index: number): number {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_CODE_POINT_AT, value, [index]) as number;
}

// Pi 0.80.6's pinned TypeBox guard uses this intentionally small grapheme
// implementation. Preserve its fast-path quirks as part of the tool ABI while
// charging each consumed UTF-16 point's exact JSON/UTF-8 bytes in the same walk.
function chargeProjectedStringAndMeasure(
	budget: ProjectionBudget,
	value: string,
	description: string,
): number {
	chargeProjectionBytes(budget, 2, description);
	if (value.length > budget.maximumBytes - budget.bytes) {
		throw new ToolPolicyError(`${description} exceeded its incremental encoded-byte bound`);
	}
	let requiresSlowPath = false;
	const consumePoint = (index: number): number => {
		const code = INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_CHAR_CODE_AT, value, [index]) as number;
		if (isHighSurrogate(code) || (code >= 0x0300 && code <= 0x036f) || code === 0x200d) {
			requiresSlowPath = true;
		}
		if (code === 0x22 || code === 0x5c || code === 0x08 || code === 0x09 || code === 0x0a ||
			code === 0x0c || code === 0x0d) {
			chargeProjectionBytes(budget, 2, description);
			return index + 1;
		}
		if (code <= 0x1f) {
			chargeProjectionBytes(budget, 6, description);
			return index + 1;
		}
		if (isHighSurrogate(code)) {
			if (index + 1 < value.length) {
				const second = INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_CHAR_CODE_AT, value, [index + 1]) as number;
				if (isLowSurrogate(second)) {
					chargeProjectionBytes(budget, 4, description);
					return index + 2;
				}
			}
			chargeProjectionBytes(budget, 6, description);
			return index + 1;
		}
		if (isLowSurrogate(code)) chargeProjectionBytes(budget, 6, description);
		else if (code <= 0x7f) chargeProjectionBytes(budget, 1, description);
		else if (code <= 0x7ff) chargeProjectionBytes(budget, 2, description);
		else chargeProjectionBytes(budget, 3, description);
		return index + 1;
	};

	let graphemes = 0;
	let index = 0;
	while (index < value.length) {
		graphemes += 1;
		const firstPoint = codePointAt(value, index);
		index = consumePoint(index);

		// TypeBox pairs leading regional indicators, matching its pinned guard.
		if (firstPoint >= 0x1f1e6 && firstPoint <= 0x1f1ff && index < value.length) {
			const nextPoint = codePointAt(value, index);
			if (nextPoint >= 0x1f1e6 && nextPoint <= 0x1f1ff) index = consumePoint(index);
		}

		while (index < value.length && isTypeBoxCombiningMark(codePointAt(value, index))) {
			index = consumePoint(index);
		}
		while (index < value.length && codePointAt(value, index) === 0x200d) {
			index = consumePoint(index);
			if (index >= value.length) {
				graphemes += 1;
				break;
			}
			index = consumePoint(index);
			while (index < value.length && isTypeBoxCombiningMark(codePointAt(value, index))) {
				index = consumePoint(index);
			}
		}
	}
	return requiresSlowPath ? graphemes : value.length;
}

function chargeProjectedStringBytes(budget: ProjectionBudget, value: string, description: string): void {
	chargeProjectedStringAndMeasure(budget, value, description);
}

function normalizeSchemaNumber(
	value: unknown,
	type: "integer" | "number",
	description: string,
	budget: ProjectionBudget,
): number {
	let normalized: number;
	if (value === null) normalized = 0;
	else if (typeof value === "boolean") normalized = value ? 1 : 0;
	else if (typeof value === "number") normalized = value;
	else if (typeof value === "string") {
		chargeProjectionScalarWork(budget, value.length, description);
		if (stringTrim(value).length < 1) throw new ToolPolicyError(`${description} violates its numeric schema`);
		normalized = INTRINSIC_REFLECT_APPLY(INTRINSIC_NUMBER, undefined, [value]) as number;
	}
	else throw new ToolPolicyError(`${description} violates its numeric schema`);
	if (!INTRINSIC_NUMBER_IS_FINITE(normalized) || (type === "integer" && !INTRINSIC_NUMBER_IS_INTEGER(normalized))) {
		throw new ToolPolicyError(`${description} violates its numeric schema`);
	}
	return normalized === 0 ? 0 : normalized;
}

function normalizeSchemaBoolean(value: unknown, description: string): boolean {
	if (typeof value === "boolean") return value;
	if (value === null) return false;
	if (value === "true" || value === 1) return true;
	if (value === "false" || (typeof value === "number" && value === 0)) return false;
	throw new ToolPolicyError(`${description} violates its boolean schema`);
}

function chargeProjectionNode(budget: ProjectionBudget, depth: number, description: string): void {
	budget.nodes += 1;
	if (budget.nodes > MAX_PROJECTED_NODES || depth > MAX_PROJECTED_DEPTH) {
		throw new ToolPolicyError(`${description} exceeded its shared node or depth bound`);
	}
}

function chargeProjectionContainer(budget: ProjectionBudget, value: object, description: string): void {
	budget.containers += 1;
	if (budget.containers > MAX_PROJECTED_CONTAINER_ENCOUNTERS) {
		throw new ToolPolicyError(`${description} exceeded its shared container bound`);
	}
	if (INTRINSIC_REFLECT_APPLY(INTRINSIC_WEAK_SET_HAS, budget.seen, [value])) {
		throw new ToolPolicyError(`${description} repeats an object or array identity`);
	}
	INTRINSIC_REFLECT_APPLY(INTRINSIC_WEAK_SET_ADD, budget.seen, [value]);
}

function assertBoundedClosedRecordKeys(
	value: Readonly<Record<string, unknown>>,
	allowed: ReadonlySet<string>,
	expectedCount: number,
	budget: ProjectionBudget,
	description: string,
): void {
	let ownCount = 0;
	for (const field in value) {
		chargeProjectionKeys(budget, 1, description);
		if (!INTRINSIC_OBJECT_HAS_OWN(value, field)) continue;
		ownCount += 1;
		if (ownCount > expectedCount || !allowed.has(field)) {
			throw new ToolPolicyError(`${description} contains unknown or missing fields`);
		}
	}
	if (ownCount !== expectedCount) {
		throw new ToolPolicyError(`${description} contains unknown or missing fields`);
	}
}

function chargeProjectionKeys(budget: ProjectionBudget, count: number, description: string): void {
	if (count > MAX_PROJECTED_KEYS - budget.keys) {
		throw new ToolPolicyError(`${description} exceeded its shared object-key bound`);
	}
	budget.keys += count;
}

function chargeProjectionItems(budget: ProjectionBudget, count: number, description: string): void {
	if (count > MAX_PROJECTED_ITEMS - budget.items) {
		throw new ToolPolicyError(`${description} exceeded its shared array-item bound`);
	}
	budget.items += count;
}

function chargeProjectionBytes(budget: ProjectionBudget, count: number, description: string): void {
	if (count > budget.maximumBytes - budget.bytes) {
		throw new ToolPolicyError(`${description} exceeded its incremental encoded-byte bound`);
	}
	budget.bytes += count;
}

function chargeProjectionScalarWork(budget: ProjectionBudget, count: number, description: string): void {
	if (count > budget.maximumBytes - budget.scalarUnits) {
		throw new ToolPolicyError(`${description} exceeded its bounded scalar-work limit`);
	}
	budget.scalarUnits += count;
}

function chargeProjectedScalarBytes(budget: ProjectionBudget, value: JsonData, description: string): void {
	if (typeof value === "string") {
		chargeProjectedStringBytes(budget, value, description);
		return;
	}
	if (value === null) {
		chargeProjectionBytes(budget, 4, description);
		return;
	}
	if (typeof value === "boolean") {
		chargeProjectionBytes(budget, value ? 4 : 5, description);
		return;
	}
	if (typeof value === "number" && INTRINSIC_NUMBER_IS_FINITE(value)) {
		const encoded = INTRINSIC_REFLECT_APPLY(INTRINSIC_JSON_STRINGIFY, undefined, [value]);
		if (typeof encoded === "string") {
			chargeProjectionBytes(budget, encoded.length, description);
			return;
		}
	}
	throw new ToolPolicyError(`${description} contains a non-JSON scalar`);
}

function ownArrayLength(value: readonly unknown[], description: string): number {
	const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, "length");
	if (!descriptor || descriptor.enumerable || descriptor.get || descriptor.set || !("value" in descriptor) ||
		!INTRINSIC_NUMBER_IS_SAFE_INTEGER(descriptor.value) || descriptor.value < 0) {
		throw new ToolPolicyError(`${description} has an invalid authoritative length`);
	}
	return descriptor.value as number;
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
		new INTRINSIC_WEAK_SET<object>(),
		`${name} parameter schema`,
	);
	if (!isRecord(snapshot)) {
		throw new ToolPolicyError(`capability ${name} requires a bounded parameter schema`);
	}
	if (snapshot.type !== "object" || snapshot.additionalProperties !== false) {
		throw new ToolPolicyError(`capability ${name} parameter schema must be a closed object`);
	}
	const serialized = INTRINSIC_JSON_STRINGIFY(snapshot);
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
		if (!INTRINSIC_NUMBER_IS_FINITE(value)) throw new ToolPolicyError(`${description} contains a non-JSON number`);
		addSchemaBytes(budget, 24, description);
		return value;
	}
	if (typeof value !== "object") throw new ToolPolicyError(`${description} contains a non-JSON value`);
	if (INTRINSIC_IS_PROXY(value)) throw new ToolPolicyError(`${description} cannot be a Proxy`);
	if (INTRINSIC_REFLECT_APPLY(INTRINSIC_WEAK_SET_HAS, ancestors, [value])) {
		throw new ToolPolicyError(`${description} contains a cycle`);
	}
	INTRINSIC_REFLECT_APPLY(INTRINSIC_WEAK_SET_ADD, ancestors, [value]);
	addSchemaBytes(budget, 2, description);

	try {
		if (INTRINSIC_IS_ARRAY(value)) {
			if (INTRINSIC_GET_PROTOTYPE_OF(value) !== INTRINSIC_ARRAY_PROTOTYPE) {
				throw new ToolPolicyError(`${description} array must use the exact Array prototype`);
			}
			if (value.length > MAX_CAPABILITY_SCHEMA_ARRAY_ITEMS) {
				throw new ToolPolicyError(`${description} array exceeded its bound`);
			}
			const lengthDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, "length");
			if (!lengthDescriptor || !("value" in lengthDescriptor) || lengthDescriptor.value !== value.length ||
				lengthDescriptor.enumerable || lengthDescriptor.get || lengthDescriptor.set) {
				throw new ToolPolicyError(`${description} array must be dense, plain, and bounded`);
			}
			budget.keys += value.length;
			if (budget.keys > MAX_CAPABILITY_SCHEMA_KEYS) throw new ToolPolicyError(`${description} exceeded its key bound`);
			const result: JsonData[] = [];
			for (let index = 0; index < value.length; index += 1) {
				const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, intrinsicString(index));
				if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
					throw new ToolPolicyError(`${description} array contains an accessor or sparse item`);
				}
				arrayPush(result, snapshotJsonData(descriptor.value, depth + 1, budget, ancestors, description));
			}
			return INTRINSIC_OBJECT_FREEZE(result) as JsonData[];
		}

		const prototype = INTRINSIC_GET_PROTOTYPE_OF(value);
		if (prototype !== INTRINSIC_OBJECT_PROTOTYPE && prototype !== null) {
			throw new ToolPolicyError(`${description} must contain plain JSON objects only`);
		}
		// Prototype safety is established before bounded own-name discovery. Hidden peers stay
		// opaque; enumerable JSON fields are then captured descriptor by descriptor.
		const result = INTRINSIC_OBJECT_CREATE(null) as { [key: string]: JsonData };
		const keys = INTRINSIC_OBJECT_KEYS(value);
		if (keys.length > MAX_CAPABILITY_SCHEMA_KEYS - budget.keys) {
			throw new ToolPolicyError(`${description} exceeded its key bound`);
		}
		for (const key of keys) {
			budget.keys += 1;
			if (budget.keys > MAX_CAPABILITY_SCHEMA_KEYS) {
				throw new ToolPolicyError(`${description} exceeded its key bound`);
			}
			addSchemaBytes(budget, byteLength(key) + 3, description);
			const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, key);
			if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
				throw new ToolPolicyError(`${description} contains an accessor field`);
			}
			INTRINSIC_DEFINE_PROPERTY(result, key, {
				value: snapshotJsonData(descriptor.value, depth + 1, budget, ancestors, description),
				enumerable: true,
				writable: false,
				configurable: false,
			});
		}
		return INTRINSIC_OBJECT_FREEZE(result);
	} finally {
		INTRINSIC_REFLECT_APPLY(INTRINSIC_WEAK_SET_DELETE, ancestors, [value]);
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
				const offset = optionalBoundedInteger(params.offset, "offset", 0, INTRINSIC_NUMBER_MAX_SAFE_INTEGER);
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
	capability: CompiledHostCapability,
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
				const params = capability.argumentContract.project(rawParams, limits.maxWriteCharacters);
				const result = captureCapabilityResult(capability.name, await capability.execute(params, signal));
				const references: string[] = [];
				for (let index = 0; index < result.references.length; index += 1) {
					references[index] = redactSensitiveText(result.references[index]!);
				}
				return textResult(INTRINSIC_JSON_STRINGIFY({
					status: result.status,
					summary: redactSensitiveText(result.summary),
					references,
				}), limits.maxToolOutputBytes);
			});
		},
	};
}

function projectWorkspaceArguments(
	name: WorkspaceToolName,
	raw: unknown,
	maximumBytes: number,
): Readonly<Record<string, unknown>> {
	if (!INTRINSIC_NUMBER_IS_SAFE_INTEGER(maximumBytes) || maximumBytes < 1) {
		throw new ToolPolicyError(`${name} arguments have an invalid byte bound`);
	}
	const effectiveMaximum = intrinsicMin(maximumBytes, MAX_TOOL_INPUT_BYTES);
	const allowed = name === "workspace_read"
		? ["path", "offset", "limit"] as const
		: name === "workspace_edit"
			? ["path", "oldText", "newText"] as const
			: ["path", "content"] as const;
	const projected = recordParams(raw, name);
	assertOnlyKeys(projected, allowed, name);
	if (byteLength(INTRINSIC_JSON_STRINGIFY(projected)) > effectiveMaximum) {
		throw new ToolPolicyError(`${name} arguments exceeded their byte bound`);
	}
	return projected;
}

export function validateScopedPath(path: string, prefixes: readonly string[]): string {
	if (typeof path !== "string" || path.length < 1 || path.length > MAX_PATH_CHARACTERS) {
		throw new ToolPolicyError("workspace path is empty or exceeds its bound");
	}
	if (/[\u0000-\u001f\u007f\\]/.test(path) || stringStartsWith(path, "/") || /^[A-Za-z]:/.test(path)) {
		throw new ToolPolicyError("workspace path must be a portable relative path without control characters");
	}
	const normalized = posix.normalize(path);
	if (normalized === ".." || stringStartsWith(normalized, "../") || normalized === ".") {
		throw new ToolPolicyError("workspace path traversal or workspace-root access is forbidden");
	}
	if (matchesSensitivePathPattern(normalized)) {
		throw new ToolPolicyError("sensitive workspace paths are outside agent authority");
	}
	let allowed = false;
	for (let index = 0; index < prefixes.length; index += 1) {
		const prefix = prefixes[index];
		if (prefix === "." || normalized === prefix || stringStartsWith(normalized, `${prefix}/`)) {
			allowed = true;
			break;
		}
	}
	if (!allowed) {
		throw new ToolPolicyError(`workspace path ${INTRINSIC_JSON_STRINGIFY(normalized)} is outside the declared scope`);
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
	recognizerCharacterVisits: number;
	lexicalTransitions: number;
	frameOperations: number;
	recoveryTransitions: number;
	rangeEmissions: number;
	rangeExaminations: number;
	rangeInsertions: number;
	rangeCoalescences: number;
	replacementEmissions: number;
	renderedSourceUnits: number;
	totalWork: number;
}

export function redactSensitiveText(value: string): string;
export function redactSensitiveText(value: string, metrics: RedactionScanMetrics): string;
export function redactSensitiveText(value: string, metrics?: RedactionScanMetrics | number): string {
	if (typeof value !== "string") return "[REDACTED]";
	const scanMetrics = typeof metrics === "object" && metrics !== null ? metrics : undefined;
	return applyRedactionPlan(value, scanRedactionPlan(value, scanMetrics), scanMetrics);
}

type PlannedRangePriority = "assignment" | "credential" | "private-key" | "recovery";

interface PlannedRedactionRange {
	start: number;
	end: number;
	replacement: string;
	priority: PlannedRangePriority;
	active: boolean;
}

interface StreamingFlowFrame {
	closer: FlowCloser;
	openedAt: number;
	publicOwner?: StreamingAssignment;
}

type StreamingAssignmentMode = "waiting" | "plain" | "quoted" | "quoted-closed" | "composite";

interface StreamingAssignment {
	kind: AssignmentKeyClassification;
	context: SensitiveAssignmentContext;
	delimiter: ":" | "=";
	keyColumn: number;
	normalizedKey: string;
	baseDepth: number;
	mode: StreamingAssignmentMode;
	quote?: LexicalQuote;
	escapeNext: boolean;
	skipNextSingleQuote: boolean;
	valueStart?: number;
	valueEnd: number;
	valuePrefix: string;
	nonWhitespace: boolean;
	range?: PlannedRedactionRange;
	compositeDepth?: number;
	pendingLineBoundary: boolean;
}

interface StreamingKeyCandidate {
	start: number;
	context: SensitiveAssignmentContext;
	keyColumn: number;
	quoted?: LexicalQuote;
	quoteClosed: boolean;
	escapeNext: boolean;
	unicodeRemaining: number;
	unicodeValue: string;
	skipNextSingleQuote: boolean;
	exact: boolean;
	decoded: string;
	decodedLength: number;
	pendingWhitespace: boolean;
	locator: boolean;
	innerDelimiter?: { delimiter: ":" | "="; index: number };
}

interface StreamingRecovery {
	range: PlannedRedactionRange;
}

interface PrivateKeyHeaderCandidate {
	start: number;
	text: string;
}

interface PrivateKeyBlockState {
	range: PlannedRedactionRange;
	endMarker: string;
	tail: string;
}

interface UrlCredentialState {
	inAuthority: boolean;
	pendingSlashes: number;
	schemeLength: number;
	pendingRange?: PlannedRedactionRange;
}

interface QueryCredentialState {
	collecting: boolean;
	key: string;
	activeRange?: PlannedRedactionRange;
}

interface PasswordCredentialState {
	word: string;
	waitingValue: boolean;
	activeRange?: PlannedRedactionRange;
}

interface CookieCredentialState {
	tail: string;
	waitingValue: boolean;
	activeRange?: PlannedRedactionRange;
}

interface StrongCredentialState {
	privatePrefixProgress: number;
	privateHeader?: PrivateKeyHeaderCandidate;
	privateBlock?: PrivateKeyBlockState;
	url: UrlCredentialState;
	query: QueryCredentialState;
	password: PasswordCredentialState;
	cookie: CookieCredentialState;
}

interface RedactionStreamingState {
	i: number;
	lineStart: number;
	lineHasContent: boolean;
	lineCandidateSlot: boolean;
	flowCandidateSlot: boolean;
	comment: boolean;
	candidate?: StreamingKeyCandidate;
	assignments: StreamingAssignment[];
	frames: StreamingFlowFrame[];
	recovery?: StreamingRecovery;
	ranges: PlannedRedactionRange[];
	strong: StrongCredentialState;
	visits?: Uint8Array;
}

const MAX_REDACTION_NESTING = 256;
const PRIVATE_KEY_BEGIN_PREFIX = "-----begin ";
const PRIVATE_KEY_HEADER_SUFFIX = " private key-----";
const QUERY_SECRET_KEYS = new Set([
	"accesstoken",
	"refreshtoken",
	"apikey",
	"password",
	"secret",
	"clientsecret",
]);

/**
 * Plans every replacement against the caller's original UTF-16 coordinate space. The loop below
 * is the only source cursor: assignment, quote, flow, recovery, and strong-credential recognizers
 * are fixed-state consumers fed by the same character.
 */
function scanRedactionPlan(source: string, metrics?: RedactionScanMetrics): PlannedRedactionRange[] {
	initializeStreamingMetrics(metrics, source.length);
	const state: RedactionStreamingState = {
		i: 0,
		lineStart: 0,
		lineHasContent: false,
		lineCandidateSlot: true,
		flowCandidateSlot: false,
		comment: false,
		assignments: [],
		frames: [],
		ranges: [],
		strong: {
			privatePrefixProgress: 0,
			url: { inAuthority: false, pendingSlashes: 0, schemeLength: 0 },
			query: { collecting: false, key: "" },
			password: { word: "", waitingValue: false },
			cookie: { tail: "", waitingValue: false },
		},
		visits: metrics ? new Uint8Array(source.length) : undefined,
	};

	while (state.i < source.length) {
		const position = state.i;
		const character = source[position]!;
		consumeOriginalCharacter(state, metrics, position);
		feedStrongCredentialRecognizers(state, character, position, metrics);

		if (state.recovery) {
			state.recovery.range.end = position + 1;
			if (isPhysicalLineEnding(character)) finishRecoveryAtLine(state, position, metrics);
			continue;
		}

		if (isPhysicalLineEnding(character)) {
			finishPhysicalLine(source, state, position, metrics);
			continue;
		}

		if (!state.lineHasContent && !isHorizontalWhitespace(character)) {
			resolvePendingLineAssignments(state, position - state.lineStart, position, metrics);
			state.lineHasContent = true;
		}

		if (state.comment) continue;

		if (state.candidate) {
			if (advanceStreamingCandidate(source, state, character, position, metrics)) continue;
		}

		const top = arrayAt(state.assignments, -1);
		if (top?.mode === "quoted" || top?.mode === "quoted-closed") {
			if (advanceStreamingQuotedValue(source, state, top, character, position, metrics)) continue;
		}

		if (startsStreamingComment(source, state, character, position)) {
			state.comment = true;
			finishAssignmentsAtBoundary(state, position, false, metrics);
			continue;
		}

		const candidateContext = streamingCandidateContext(state);
		if (candidateContext && isPotentialAssignmentCandidateStart(character)) {
			beginStreamingCandidate(state, character, position, candidateContext, metrics);
			continue;
		}
		if (candidateContext && isHorizontalWhitespace(character)) continue;

		const current = arrayAt(state.assignments, -1);
		if (current?.mode === "waiting") {
			if (isHorizontalWhitespace(character)) continue;
			beginStreamingValue(state, current, character, position, metrics);
			continue;
		}

		if ((character === "," || character === ";") &&
			(state.frames.length > 0 || current?.context === "flow")) {
			finishAssignmentsAtMemberBoundary(state, position, metrics);
			state.flowCandidateSlot = state.frames.length > 0;
			state.lineCandidateSlot = false;
			continue;
		}

		const closer = flowCloserForOpener(character);
		if (closer) {
			pushStreamingFrameOrRecover(state, closer, position, current, metrics);
			continue;
		}

		if (character === "}" || character === "]") {
			closeStreamingFrameOrRecover(state, character, position, metrics);
			continue;
		}

		if (current?.mode === "plain") {
			appendStreamingValueCharacter(current, character, position);
			state.lineCandidateSlot = false;
			state.flowCandidateSlot = false;
			continue;
		}

		if (!isHorizontalWhitespace(character)) {
			state.lineCandidateSlot = false;
			state.flowCandidateSlot = false;
		}
	}

	finishStreamingEof(state, source.length, metrics);
	return state.ranges;
}

function consumeOriginalCharacter(
	state: RedactionStreamingState,
	metrics: RedactionScanMetrics | undefined,
	position: number,
): void {
	state.i = position + 1;
	if (!metrics) return;
	metrics.cursorAdvances += 1;
	chargeWork(metrics);
	const visits = state.visits;
	if (visits) {
		visits[position] = intrinsicMin(255, visits[position]! + 1);
		if (visits[position]! > metrics.maxMainCursorVisits) {
			metrics.maxMainCursorVisits = visits[position]!;
		}
	}
	metrics.boundaryCharacterVisits += 1;
	chargeWork(metrics);
}

function initializeStreamingMetrics(metrics: RedactionScanMetrics | undefined, sourceLength: number): void {
	if (!metrics) return;
	metrics.sourceLength = sourceLength;
	metrics.cursorAdvances = 0;
	metrics.cursorRegressions = 0;
	metrics.maxMainCursorVisits = 0;
	metrics.keyCharacterVisits = 0;
	metrics.boundaryCharacterVisits = 0;
	metrics.recognizerCharacterVisits = 0;
	metrics.lexicalTransitions = 0;
	metrics.frameOperations = 0;
	metrics.recoveryTransitions = 0;
	metrics.rangeEmissions = 0;
	metrics.rangeExaminations = 0;
	metrics.rangeInsertions = 0;
	metrics.rangeCoalescences = 0;
	metrics.replacementEmissions = 0;
	metrics.renderedSourceUnits = 0;
	metrics.totalWork = 0;
}

function chargeKeyTransition(metrics?: RedactionScanMetrics): void {
	if (!metrics) return;
	metrics.keyCharacterVisits += 1;
	chargeWork(metrics);
}

function chargeWork(metrics?: RedactionScanMetrics): void {
	if (metrics) metrics.totalWork += 1;
}

type RedactionWorkCategory =
	| "recognizerCharacterVisits"
	| "lexicalTransitions"
	| "frameOperations"
	| "recoveryTransitions"
	| "rangeEmissions"
	| "rangeExaminations"
	| "rangeInsertions"
	| "rangeCoalescences"
	| "replacementEmissions"
	| "renderedSourceUnits";

function chargeCategorizedWork(metrics: RedactionScanMetrics | undefined, category: RedactionWorkCategory): void {
	if (!metrics) return;
	metrics[category] += 1;
	metrics.totalWork += 1;
}

function emitPlannedRange(
	state: RedactionStreamingState,
	start: number,
	end: number,
	replacement: string,
	priority: PlannedRangePriority,
	active: boolean,
	metrics?: RedactionScanMetrics,
): PlannedRedactionRange {
	const range = { start, end, replacement, priority, active } satisfies PlannedRedactionRange;
	arrayPush(state.ranges, range);
	chargeCategorizedWork(metrics, "rangeEmissions");
	return range;
}

function streamingCandidateContext(state: RedactionStreamingState): SensitiveAssignmentContext | undefined {
	if (state.lineCandidateSlot) return state.frames.length > 0 ? "flow" : "line";
	if (state.flowCandidateSlot) return "flow";
	return undefined;
}

function beginStreamingCandidate(
	state: RedactionStreamingState,
	character: string,
	position: number,
	context: SensitiveAssignmentContext,
	metrics?: RedactionScanMetrics,
): void {
	const quoted = character === "\"" || character === "'" ? character : undefined;
	state.candidate = {
		start: position,
		context,
		keyColumn: position - state.lineStart,
		quoted,
		quoteClosed: false,
		escapeNext: false,
		unicodeRemaining: 0,
		unicodeValue: "",
		skipNextSingleQuote: false,
		exact: true,
		decoded: quoted ? "" : character,
		decodedLength: quoted ? 0 : 1,
		pendingWhitespace: false,
		locator: false,
	};
	chargeKeyTransition(metrics);
	state.lineCandidateSlot = false;
	state.flowCandidateSlot = false;
}

function advanceStreamingCandidate(
	source: string,
	state: RedactionStreamingState,
	character: string,
	position: number,
	metrics?: RedactionScanMetrics,
): boolean {
	const candidate = state.candidate!;
	chargeKeyTransition(metrics);
	if (candidate.quoted) {
		if (!candidate.quoteClosed) {
			if (candidate.skipNextSingleQuote) {
				candidate.skipNextSingleQuote = false;
				return true;
			}
			if (candidate.unicodeRemaining > 0) {
				candidate.unicodeValue += character;
				candidate.unicodeRemaining -= 1;
				if (candidate.unicodeRemaining === 0) {
					if (/^[0-9a-fA-F]{4}$/.test(candidate.unicodeValue)) {
						appendDecodedCandidateCharacter(
							candidate,
							stringFromCharCode(INTRINSIC_NUMBER_PARSE_INT(candidate.unicodeValue, 16)),
						);
					} else candidate.exact = false;
					candidate.unicodeValue = "";
				}
				return true;
			}
			if (candidate.escapeNext) {
				candidate.escapeNext = false;
				if (character === "u") {
					candidate.unicodeRemaining = 4;
					return true;
				}
				const decoded = simpleQuotedEscape(character);
				if (decoded === undefined) candidate.exact = false;
				else appendDecodedCandidateCharacter(candidate, decoded);
				return true;
			}
			if (candidate.quoted === "\"" && character === "\\") {
				candidate.escapeNext = true;
				return true;
			}
			if (candidate.quoted === "'" && character === "'" && source[position + 1] === "'") {
				appendDecodedCandidateCharacter(candidate, "'");
				candidate.skipNextSingleQuote = true;
				return true;
			}
			if (character === candidate.quoted) {
				candidate.quoteClosed = true;
				return true;
			}
			if ((character === ":" || character === "=") && candidate.innerDelimiter === undefined) {
				candidate.innerDelimiter = { delimiter: character, index: position };
			}
			if ((character === "," && candidate.context === "flow") || character === "}" || character === "]") {
				finishMalformedQuotedCandidate(source, state, candidate, position, metrics);
				reprocessStreamingBoundary(state, character, position, metrics);
				return true;
			}
			appendDecodedCandidateCharacter(candidate, character);
			return true;
		}

		if (isHorizontalWhitespace(character)) return true;
		if (character === ":" || character === "=") {
			commitStreamingCandidate(state, candidate, character, position);
			return true;
		}
		if ((character === "," && candidate.context === "flow") || character === "}" || character === "]") {
			state.candidate = undefined;
			reprocessStreamingBoundary(state, character, position, metrics);
			return true;
		}
		candidate.exact = false;
		return true;
	}

	if (((character === "," || character === ";") && candidate.context === "flow") ||
		character === "}" || character === "]") {
		state.candidate = undefined;
		reprocessStreamingBoundary(state, character, position, metrics);
		return true;
	}
	if (candidate.locator) return true;
	if (character === "=" || (character === ":" && admittedStreamingColon(source, position, candidate.context))) {
		commitStreamingCandidate(state, candidate, character, position);
		return true;
	}
	if (character === ":" && source[position + 1] === "/" && source[position + 2] === "/") {
		candidate.locator = true;
		candidate.exact = false;
		return true;
	}
	if (isHorizontalWhitespace(character)) {
		candidate.pendingWhitespace = candidate.decodedLength > 0;
		candidate.exact = false;
		return true;
	}
	candidate.decodedLength += 1;
	if (candidate.decodedLength <= 64 && isAssignmentKeyCharacter(character) && !candidate.pendingWhitespace) {
		candidate.decoded += character;
	} else candidate.exact = false;
	candidate.pendingWhitespace = false;
	return true;
}

function simpleQuotedEscape(character: string): string | undefined {
	if (character === "\"") return "\"";
	if (character === "\\") return "\\";
	if (character === "/") return "/";
	if (character === "b") return "\b";
	if (character === "f") return "\f";
	if (character === "n") return "\n";
	if (character === "r") return "\r";
	if (character === "t") return "\t";
	return undefined;
}

function appendDecodedCandidateCharacter(candidate: StreamingKeyCandidate, character: string): void {
	candidate.decodedLength += 1;
	if (candidate.decodedLength <= 64 && isAssignmentKeyCharacter(character)) candidate.decoded += character;
	else candidate.exact = false;
}

function admittedStreamingColon(
	source: string,
	position: number,
	context: SensitiveAssignmentContext,
): boolean {
	const next = source[position + 1];
	if (next === undefined || isHorizontalWhitespace(next) || isPhysicalLineEnding(next) || next === "#") return true;
	if (context !== "flow") return false;
	return next !== "/" && !(next >= "0" && next <= "9");
}

function commitStreamingCandidate(
	state: RedactionStreamingState,
	candidate: StreamingKeyCandidate,
	delimiter: ":" | "=",
	position: number,
): void {
	const exactKey = candidate.exact && candidate.decodedLength > 0 && candidate.decodedLength <= 64
		? candidate.decoded
		: undefined;
	const kind = exactKey === undefined ? "unknown-sensitive" : assignmentKeyClassification(exactKey);
	const normalizedKeyParts = exactKey === undefined ? [] : stringSplit(stringToLowerCase(exactKey), ".");
	const normalizedKey = exactKey === undefined
		? ""
		: stringReplace(arrayAt(normalizedKeyParts, -1) ?? exactKey, /[-_]/g, "");
	arrayPush(state.assignments, {
		kind,
		context: candidate.context,
		delimiter,
		keyColumn: candidate.keyColumn,
		normalizedKey,
		baseDepth: state.frames.length,
		mode: "waiting",
		escapeNext: false,
		skipNextSingleQuote: false,
		valueEnd: position + 1,
		valuePrefix: "",
		nonWhitespace: false,
		pendingLineBoundary: false,
	});
	state.candidate = undefined;
	state.lineCandidateSlot = false;
	state.flowCandidateSlot = false;
}

function finishMalformedQuotedCandidate(
	source: string,
	state: RedactionStreamingState,
	candidate: StreamingKeyCandidate,
	end: number,
	metrics?: RedactionScanMetrics,
): void {
	if (candidate.innerDelimiter) {
		let start = candidate.innerDelimiter.index + 1;
		while (start < end && isHorizontalWhitespace(source[start])) start += 1;
		if (start < end) emitPlannedRange(state, start, end, REDACTED_TEXT, "assignment", true, metrics);
	}
	state.candidate = undefined;
}

function beginStreamingValue(
	state: RedactionStreamingState,
	assignment: StreamingAssignment,
	character: string,
	position: number,
	metrics?: RedactionScanMetrics,
): void {
	assignment.valueStart = character === "\"" || character === "'" ? position + 1 : position;
	assignment.valueEnd = assignment.valueStart;
	assignment.nonWhitespace = true;
	assignment.mode = character === "\"" || character === "'" ? "quoted" : "plain";
	if (assignment.mode === "quoted") assignment.quote = character as LexicalQuote;
	assignment.range = emitPlannedRange(
		state,
		assignment.valueStart,
		assignment.valueStart,
		REDACTED_TEXT,
		assignment.kind === "public" ? "recovery" : "assignment",
		assignment.kind !== "public",
		metrics,
	);
	if (assignment.mode === "quoted") return;
	appendStreamingValueCharacter(assignment, character, position);
	const closer = flowCloserForOpener(character);
	if (closer) pushStreamingFrameOrRecover(state, closer, position, assignment, metrics);
}

function appendStreamingValueCharacter(
	assignment: StreamingAssignment,
	character: string,
	position: number,
): void {
	assignment.valueEnd = position + 1;
	if (assignment.range) assignment.range.end = position + 1;
	if (!assignment.nonWhitespace && isHorizontalWhitespace(character)) return;
	assignment.nonWhitespace = true;
	if (assignment.valuePrefix.length < 160) assignment.valuePrefix += character;
}

function advanceStreamingQuotedValue(
	source: string,
	state: RedactionStreamingState,
	assignment: StreamingAssignment,
	character: string,
	position: number,
	metrics?: RedactionScanMetrics,
): boolean {
	if (assignment.mode === "quoted-closed") {
		if (isHorizontalWhitespace(character)) return true;
		if (character === "," || character === ";" || character === "}" || character === "]") {
			finishOneStreamingAssignment(state, assignment, position, false, metrics);
			reprocessStreamingBoundary(state, character, position, metrics);
			return true;
		}
		if (character === "#" && isHorizontalWhitespace(source[position - 1])) {
			finishOneStreamingAssignment(state, assignment, position, false, metrics);
			state.comment = true;
			return true;
		}
		assignment.mode = "plain";
		appendStreamingValueCharacter(assignment, character, position);
		return true;
	}
	if (assignment.skipNextSingleQuote) {
		assignment.skipNextSingleQuote = false;
		appendStreamingValueCharacter(assignment, character, position);
		return true;
	}
	if (assignment.escapeNext) {
		assignment.escapeNext = false;
		appendStreamingValueCharacter(assignment, character, position);
		return true;
	}
	if (assignment.quote === "\"" && character === "\\") {
		assignment.escapeNext = true;
		appendStreamingValueCharacter(assignment, character, position);
		return true;
	}
	if (assignment.quote === "'" && character === "'" && source[position + 1] === "'") {
		assignment.skipNextSingleQuote = true;
		appendStreamingValueCharacter(assignment, character, position);
		return true;
	}
	if (character === assignment.quote) {
		if (assignment.range) assignment.range.end = position;
		assignment.valueEnd = position;
		assignment.mode = "quoted-closed";
		return true;
	}
	appendStreamingValueCharacter(assignment, character, position);
	return true;
}

function startsStreamingComment(
	source: string,
	state: RedactionStreamingState,
	character: string,
	position: number,
): boolean {
	return character === "#" &&
		(position === state.lineStart || isHorizontalWhitespace(source[position - 1]));
}

function resolvePendingLineAssignments(
	state: RedactionStreamingState,
	column: number,
	position: number,
	metrics?: RedactionScanMetrics,
): void {
	while (true) {
		const assignment = state.assignments.at(-1);
		if (!assignment?.pendingLineBoundary) return;
		if (column > assignment.keyColumn) {
			assignment.pendingLineBoundary = false;
			if (assignment.mode === "waiting") beginContinuationRange(state, assignment, position, metrics);
			state.lineCandidateSlot = false;
			state.flowCandidateSlot = false;
			return;
		}
		finishOneStreamingAssignment(state, assignment, assignment.valueEnd, false, metrics);
	}
}

function beginContinuationRange(
	state: RedactionStreamingState,
	assignment: StreamingAssignment,
	position: number,
	metrics?: RedactionScanMetrics,
): void {
	assignment.valueStart = position;
	assignment.valueEnd = position;
	assignment.mode = "plain";
	assignment.nonWhitespace = true;
	assignment.range = emitPlannedRange(
		state,
		position,
		position,
		REDACTED_TEXT,
		assignment.kind === "public" ? "recovery" : "assignment",
		assignment.kind !== "public",
		metrics,
	);
}

function finishPhysicalLine(
	source: string,
	state: RedactionStreamingState,
	position: number,
	metrics?: RedactionScanMetrics,
): void {
	if (state.candidate) {
		const candidate = state.candidate;
		if (candidate.quoted && !candidate.quoteClosed) {
			finishMalformedQuotedCandidate(source, state, candidate, position, metrics);
		} else state.candidate = undefined;
	}
	const top = arrayAt(state.assignments, -1);
	if (top?.mode === "quoted") {
		finishOneStreamingAssignment(state, top, position, true, metrics);
	} else if (top && top.compositeDepth === undefined) {
		top.pendingLineBoundary = true;
		top.valueEnd = position;
		if (top.range) top.range.end = position;
	}
	state.comment = false;
	state.lineStart = position + 1;
	state.lineHasContent = false;
	state.lineCandidateSlot = true;
	state.flowCandidateSlot = false;
}

function finishAssignmentsAtBoundary(
	state: RedactionStreamingState,
	position: number,
	malformedQuote: boolean,
	metrics?: RedactionScanMetrics,
): void {
	const assignment = arrayAt(state.assignments, -1);
	if (assignment && assignment.compositeDepth === undefined) {
		finishOneStreamingAssignment(state, assignment, position, malformedQuote, metrics);
	}
}

function finishAssignmentsAtMemberBoundary(
	state: RedactionStreamingState,
	position: number,
	metrics?: RedactionScanMetrics,
): void {
	while (true) {
		const assignment = arrayAt(state.assignments, -1);
		if (!assignment || assignment.baseDepth !== state.frames.length || assignment.compositeDepth !== undefined) return;
		finishOneStreamingAssignment(state, assignment, position, false, metrics);
	}
}

function finishOneStreamingAssignment(
	state: RedactionStreamingState,
	assignment: StreamingAssignment,
	end: number,
	malformed: boolean,
	metrics?: RedactionScanMetrics,
): void {
	chargeCategorizedWork(metrics, "lexicalTransitions");
	assignment.valueEnd = intrinsicMax(assignment.valueStart ?? end, end);
	if (assignment.range) assignment.range.end = assignment.valueEnd;
	const prefix = stringTrim(assignment.valuePrefix);
	if (assignment.kind === "public") {
		if (assignment.range) {
			assignment.range.active = malformed && !isDocumentaryPublicMultilinePrefix(prefix);
		}
	} else if (assignment.kind === "authorization") {
		finalizeAuthorizationRange(assignment, prefix);
	} else if (assignment.range) {
		const documentary = assignment.delimiter === ":" && isBoundedDocumentaryProse(prefix);
		const publicLiteral = assignment.kind === "secret" && isReviewedPublicLiteral(prefix);
		if (documentary || publicLiteral || prefix.length === 0) assignment.range.active = false;
	}
	const index = arrayLastIndexOf(state.assignments, assignment);
	if (index >= 0) arraySplice(state.assignments, index, 1);
}

function finalizeAuthorizationRange(assignment: StreamingAssignment, prefix: string): void {
	const range = assignment.range;
	if (!range) return;
	let firstSpace = -1;
	for (let index = 0; index < prefix.length; index += 1) {
		if (isWhitespace(prefix[index])) { firstSpace = index; break; }
	}
	if (firstSpace < 0) return;
	let credential = firstSpace;
	while (credential < prefix.length && isWhitespace(prefix[credential])) credential += 1;
	if (credential >= prefix.length) {
		range.active = false;
		return;
	}
	const scheme = stringToLowerCase(stringSlice(prefix, 0, firstSpace));
	const credentialText = stringSlice(prefix, credential);
	const parameterized = scheme === "digest" || scheme === "signature" || stringStartsWith(scheme, "aws4-");
	if (!parameterized && containsWhitespace(credentialText)) {
		range.active = false;
		return;
	}
	range.start = (assignment.valueStart ?? range.start) + credential;
}

function isBoundedDocumentaryProse(value: string): boolean {
	const lower = stringToLowerCase(value);
	return stringStartsWith(lower, "number of ") || stringStartsWith(lower, "name of ") ||
		stringStartsWith(lower, "description of ") || stringStartsWith(lower, "meaning of ") ||
		stringStartsWith(lower, "definition of ") || stringStartsWith(lower, "count of ") ||
		lower === "a surprising detail in a story." ||
		stringStartsWith(lower, "describes ") || stringStartsWith(lower, "describe ") ||
		stringStartsWith(lower, "names ") || stringStartsWith(lower, "name ") ||
		stringStartsWith(lower, "explains ") || stringStartsWith(lower, "explain ") ||
		stringStartsWith(lower, "means ") || stringStartsWith(lower, "refers to ");
}

function isDocumentaryPublicMultilinePrefix(value: string): boolean {
	return stringStartsWith(stringToLowerCase(value), "the following lines document configuration vocabulary");
}

function isReviewedPublicLiteral(value: string): boolean {
	const lower = stringToLowerCase(value);
	return lower === "true" || lower === "false" || lower === "null" || lower === "~";
}

function pushStreamingFrameOrRecover(
	state: RedactionStreamingState,
	closer: FlowCloser,
	position: number,
	assignment: StreamingAssignment | undefined,
	metrics?: RedactionScanMetrics,
): void {
	if (state.frames.length >= MAX_REDACTION_NESTING) {
		beginStreamingRecovery(state, position, assignment?.range, "recovery", metrics);
		return;
	}
	if (assignment && assignment.valueStart === position && assignment.mode === "plain") {
		assignment.mode = "composite";
		assignment.compositeDepth = state.frames.length + 1;
	}
	arrayPush(state.frames, {
		closer,
		openedAt: position,
		publicOwner: assignment?.kind === "public" && assignment.compositeDepth === state.frames.length + 1
			? assignment
			: undefined,
	});
	chargeCategorizedWork(metrics, "frameOperations");
	state.flowCandidateSlot = true;
	state.lineCandidateSlot = false;
}

function closeStreamingFrameOrRecover(
	state: RedactionStreamingState,
	character: FlowCloser,
	position: number,
	metrics?: RedactionScanMetrics,
): void {
	finishAssignmentsAtMemberBoundary(state, position, metrics);
	const frame = arrayAt(state.frames, -1);
	if (!frame || frame.closer !== character) {
		const publicRange = latestPublicRecoveryRange(state, metrics);
		beginStreamingRecovery(state, publicRange?.start ?? position, publicRange, "recovery", metrics);
		return;
	}
	const closingDepth = state.frames.length;
	arrayPop(state.frames);
	chargeCategorizedWork(metrics, "frameOperations");
	while (true) {
		const assignment = arrayAt(state.assignments, -1);
		if (!assignment || assignment.compositeDepth !== closingDepth) break;
		finishOneStreamingAssignment(state, assignment, position + 1, false, metrics);
	}
	state.flowCandidateSlot = false;
}

function latestPublicRecoveryRange(
	state: RedactionStreamingState,
	metrics?: RedactionScanMetrics,
): PlannedRedactionRange | undefined {
	for (let index = state.assignments.length - 1; index >= 0; index -= 1) {
		chargeCategorizedWork(metrics, "recoveryTransitions");
		const assignment = state.assignments[index]!;
		if (assignment.kind === "public" && assignment.range) return assignment.range;
	}
	return undefined;
}

function beginStreamingRecovery(
	state: RedactionStreamingState,
	start: number,
	existing: PlannedRedactionRange | undefined,
	priority: PlannedRangePriority,
	metrics?: RedactionScanMetrics,
): void {
	const range = existing ?? emitPlannedRange(state, start, start, REDACTED_TEXT, priority, true, metrics);
	range.active = true;
	range.priority = priority;
	range.end = intrinsicMax(range.end, state.i);
	state.recovery = { range };
	chargeCategorizedWork(metrics, "recoveryTransitions");
	state.candidate = undefined;
	state.assignments.length = 0;
}

function finishRecoveryAtLine(
	state: RedactionStreamingState,
	position: number,
	metrics?: RedactionScanMetrics,
): void {
	if (state.recovery) state.recovery.range.end = position;
	chargeCategorizedWork(metrics, "recoveryTransitions");
	state.recovery = undefined;
	state.frames.length = 0;
	state.assignments.length = 0;
	state.candidate = undefined;
	state.comment = false;
	state.lineStart = position + 1;
	state.lineHasContent = false;
	state.lineCandidateSlot = true;
	state.flowCandidateSlot = false;
}

function reprocessStreamingBoundary(
	state: RedactionStreamingState,
	character: string,
	position: number,
	metrics?: RedactionScanMetrics,
): void {
	if (character === "," || character === ";") {
		finishAssignmentsAtMemberBoundary(state, position, metrics);
		state.flowCandidateSlot = state.frames.length > 0;
		return;
	}
	if (character === "}" || character === "]") {
		closeStreamingFrameOrRecover(state, character, position, metrics);
	}
}

function finishStreamingEof(
	state: RedactionStreamingState,
	end: number,
	metrics?: RedactionScanMetrics,
): void {
	if (state.recovery) {
		state.recovery.range.end = end;
		chargeCategorizedWork(metrics, "recoveryTransitions");
	}
	if (state.candidate?.quoted && !state.candidate.quoteClosed) {
		// The delimiter-adjacent span has already been consumed and is bounded by EOF.
		const candidate = state.candidate;
		if (candidate.innerDelimiter) {
			emitPlannedRange(
				state,
				candidate.innerDelimiter.index + 1,
				end,
				REDACTED_TEXT,
				"assignment",
				true,
				metrics,
			);
		}
	}
	while (state.assignments.length > 0) {
		const assignment = state.assignments.at(-1)!;
		const malformed = assignment.kind === "public" &&
			(assignment.mode === "quoted" || assignment.mode === "composite");
		finishOneStreamingAssignment(state, assignment, end, malformed, metrics);
	}
	for (const frame of state.frames) {
		chargeCategorizedWork(metrics, "frameOperations");
		if (frame.publicOwner?.range) {
			frame.publicOwner.range.active = true;
			frame.publicOwner.range.end = end;
		}
	}
	if (state.strong.privateBlock) state.strong.privateBlock.range.end = end;
	finishStrongScalarRange(state.strong.query.activeRange, end);
	finishStrongScalarRange(state.strong.password.activeRange, end);
	finishStrongScalarRange(state.strong.cookie.activeRange, end);
}

function isPhysicalLineEnding(character: string): boolean {
	return character === "\n" || character === "\r";
}

function feedStrongCredentialRecognizers(
	state: RedactionStreamingState,
	character: string,
	position: number,
	metrics?: RedactionScanMetrics,
): void {
	chargeCategorizedWork(metrics, "recognizerCharacterVisits");
	feedPrivateKeyRecognizer(state, character, position, metrics);
	chargeCategorizedWork(metrics, "recognizerCharacterVisits");
	feedUrlCredentialRecognizer(state, character, position, metrics);
	chargeCategorizedWork(metrics, "recognizerCharacterVisits");
	feedQueryCredentialRecognizer(state, character, position, metrics);
	chargeCategorizedWork(metrics, "recognizerCharacterVisits");
	feedPasswordCredentialRecognizer(state, character, position, metrics);
	chargeCategorizedWork(metrics, "recognizerCharacterVisits");
	feedCookieCredentialRecognizer(state, character, position, metrics);
	if (isPhysicalLineEnding(character)) {
		state.strong.password.word = "";
		state.strong.password.waitingValue = false;
	}
}

function feedCookieCredentialRecognizer(
	state: RedactionStreamingState,
	character: string,
	position: number,
	metrics?: RedactionScanMetrics,
): void {
	const cookie = state.strong.cookie;
	if (cookie.activeRange) {
		if (isPhysicalLineEnding(character)) {
			finishStrongScalarRange(cookie.activeRange, position);
			cookie.activeRange = undefined;
			cookie.tail = "";
			return;
		}
		cookie.activeRange.end = position + 1;
		return;
	}
	if (cookie.waitingValue) {
		if (isHorizontalWhitespace(character)) return;
		cookie.waitingValue = false;
		if (!isPhysicalLineEnding(character)) {
			cookie.activeRange = emitPlannedRange(
				state,
				position,
				position + 1,
				REDACTED_TEXT,
				"credential",
				true,
				metrics,
			);
		}
		return;
	}
	cookie.tail += character.toLowerCase();
	if (cookie.tail.length > 48) cookie.tail = stringSlice(cookie.tail, -48);
	if (character !== ":") return;
	if (stringEndsWith(cookie.tail, "request header cookie:") ||
		stringEndsWith(cookie.tail, "request headers cookie:") ||
		stringEndsWith(cookie.tail, "request header set-cookie:") ||
		stringEndsWith(cookie.tail, "request headers set-cookie:") ||
		stringEndsWith(cookie.tail, "response header cookie:") ||
		stringEndsWith(cookie.tail, "response headers cookie:") ||
		stringEndsWith(cookie.tail, "response header set-cookie:") ||
		stringEndsWith(cookie.tail, "response headers set-cookie:")) {
		cookie.waitingValue = true;
	}
}

function feedPrivateKeyRecognizer(
	state: RedactionStreamingState,
	character: string,
	position: number,
	metrics?: RedactionScanMetrics,
): void {
	const lower = character.toLowerCase();
	const block = state.strong.privateBlock;
	if (block) {
		block.tail += lower;
		if (block.tail.length > block.endMarker.length) block.tail = stringSlice(block.tail, -block.endMarker.length);
		if (block.tail === block.endMarker) {
			block.range.end = position + 1;
			state.strong.privateBlock = undefined;
		}
		return;
	}

	const header = state.strong.privateHeader;
	if (header) {
		if (isPhysicalLineEnding(character) || header.text.length >= 96) {
			state.strong.privateHeader = undefined;
		} else {
			header.text += lower;
			let label: string | undefined;
			if (header.text === "private key-----") label = "";
			else if (stringEndsWith(header.text, PRIVATE_KEY_HEADER_SUFFIX)) {
				const candidateLabel = stringSlice(header.text, 0, -PRIVATE_KEY_HEADER_SUFFIX.length);
				if (candidateLabel.length > 0 && candidateLabel.length <= 64 && isPrivateKeyLabel(candidateLabel)) {
					label = candidateLabel;
				}
			}
			if (label !== undefined) {
				const range = emitPlannedRange(
					state,
					header.start,
					position + 1,
					"[REDACTED PRIVATE KEY]",
					"private-key",
					true,
					metrics,
				);
				state.strong.privateBlock = {
					range,
					endMarker: `-----end ${label.length > 0 ? `${label} ` : ""}private key-----`,
					tail: "",
				};
				state.strong.privateHeader = undefined;
			}
		}
		return;
	}

	let progress = state.strong.privatePrefixProgress;
	if (lower === PRIVATE_KEY_BEGIN_PREFIX[progress]) progress += 1;
	else progress = lower === PRIVATE_KEY_BEGIN_PREFIX[0] ? 1 : 0;
	if (progress === PRIVATE_KEY_BEGIN_PREFIX.length) {
		state.strong.privateHeader = {
			start: position - PRIVATE_KEY_BEGIN_PREFIX.length + 1,
			text: "",
		};
		progress = 0;
	}
	state.strong.privatePrefixProgress = progress;
}

function isPrivateKeyLabel(value: string): boolean {
	for (const character of value) {
		const code = INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_CHAR_CODE_AT, character, [0]) as number;
		if (!((code >= 48 && code <= 57) || (code >= 97 && code <= 122) || character === " ")) return false;
	}
	return true;
}

function feedUrlCredentialRecognizer(
	state: RedactionStreamingState,
	character: string,
	position: number,
	metrics?: RedactionScanMetrics,
): void {
	const url = state.strong.url;
	if (url.pendingSlashes > 0) {
		if (character === "/") {
			url.pendingSlashes -= 1;
			if (url.pendingSlashes === 0) url.inAuthority = true;
		} else url.pendingSlashes = 0;
		return;
	}
	if (url.inAuthority) {
		if (character === "@") {
			if (url.pendingRange) {
				url.pendingRange.active = url.pendingRange.start < position;
				url.pendingRange.end = position;
				url.pendingRange = undefined;
			}
			return;
		}
		if (character === ":" && !url.pendingRange) {
			url.pendingRange = emitPlannedRange(
				state,
				position + 1,
				position + 1,
				REDACTED_TEXT,
				"credential",
				false,
				metrics,
			);
			return;
		}
		if (character === "/" || character === "?" || character === "#" || isWhitespace(character)) {
			if (url.pendingRange) url.pendingRange.active = false;
			url.pendingRange = undefined;
			url.inAuthority = false;
			url.schemeLength = 0;
			return;
		}
		if (url.pendingRange) url.pendingRange.end = position + 1;
		return;
	}

	if (character === ":" && url.schemeLength > 0) {
		url.pendingSlashes = 2;
		url.schemeLength = 0;
		return;
	}
	const code = INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_CHAR_CODE_AT, character, [0]) as number;
	if ((code >= 65 && code <= 90) || (code >= 97 && code <= 122) ||
		(url.schemeLength > 0 && ((code >= 48 && code <= 57) || character === "+" || character === "." || character === "-"))) {
		url.schemeLength += 1;
	} else url.schemeLength = 0;
}

function feedQueryCredentialRecognizer(
	state: RedactionStreamingState,
	character: string,
	position: number,
	metrics?: RedactionScanMetrics,
): void {
	const query = state.strong.query;
	if (query.activeRange) {
		if (character === "&" || character === "#" || isWhitespace(character)) {
			finishStrongScalarRange(query.activeRange, position);
			query.activeRange = undefined;
			query.collecting = character === "&" || character === "#";
			query.key = "";
			return;
		}
		query.activeRange.end = position + 1;
		return;
	}
	if (character === "?" || character === "&" || character === "#") {
		query.collecting = true;
		query.key = "";
		return;
	}
	if (!query.collecting) return;
	if (character === "=") {
		const normalized = stringToLowerCase(stringReplace(query.key, /[-_]/g, ""));
		if (QUERY_SECRET_KEYS.has(normalized)) {
			query.activeRange = emitPlannedRange(
				state,
				position + 1,
				position + 1,
				REDACTED_TEXT,
				"credential",
				true,
				metrics,
			);
		}
		query.collecting = false;
		return;
	}
	const code = INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_CHAR_CODE_AT, character, [0]) as number;
	if ((code >= 48 && code <= 57) || (code >= 65 && code <= 90) || (code >= 97 && code <= 122) ||
		character === "_" || character === "-") {
		if (query.key.length < 32) query.key += character;
		else query.collecting = false;
	} else query.collecting = false;
}

function feedPasswordCredentialRecognizer(
	state: RedactionStreamingState,
	character: string,
	position: number,
	metrics?: RedactionScanMetrics,
): void {
	const password = state.strong.password;
	if (password.activeRange) {
		if (isWhitespace(character) || character === "#") {
			finishStrongScalarRange(password.activeRange, position);
			password.activeRange = undefined;
			password.word = "";
			return;
		}
		password.activeRange.end = position + 1;
		return;
	}
	if (password.waitingValue) {
		if (isHorizontalWhitespace(character)) return;
		password.waitingValue = false;
		if (character !== "=" && character !== ":" && !isPhysicalLineEnding(character)) {
			password.activeRange = emitPlannedRange(
				state,
				position,
				position + 1,
				REDACTED_TEXT,
				"credential",
				true,
				metrics,
			);
		}
		return;
	}
	const code = INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_CHAR_CODE_AT, character, [0]) as number;
	const wordCharacter = (code >= 65 && code <= 90) || (code >= 97 && code <= 122);
	if (wordCharacter) {
		if (password.word.length < 16) password.word += character.toLowerCase();
		else password.word = "";
		return;
	}
	if (isHorizontalWhitespace(character) && password.word === "password") {
		password.waitingValue = true;
	}
	password.word = "";
}

function finishStrongScalarRange(range: PlannedRedactionRange | undefined, end: number): void {
	if (!range) return;
	range.end = end;
	if (range.start >= range.end) range.active = false;
}

function applyRedactionPlan(
	source: string,
	planned: readonly PlannedRedactionRange[],
	metrics?: RedactionScanMetrics,
): string {
	const ranges: PlannedRedactionRange[] = [];
	for (const candidate of planned) {
		chargeCategorizedWork(metrics, "rangeExaminations");
		if (!candidate.active || candidate.end <= candidate.start) continue;
		const start = intrinsicMax(0, intrinsicMin(source.length, candidate.start));
		const end = intrinsicMax(start, intrinsicMin(source.length, candidate.end));
		if (end <= start) continue;
		const previous = ranges.at(-1);
		if (previous && start <= previous.end) {
			chargeCategorizedWork(metrics, "rangeCoalescences");
			previous.end = intrinsicMax(previous.end, end);
			if (rangePriority(candidate.priority) > rangePriority(previous.priority)) {
				previous.priority = candidate.priority;
				previous.replacement = candidate.replacement;
			}
		} else {
			chargeCategorizedWork(metrics, "rangeInsertions");
			arrayPush(ranges, { ...candidate, start, end });
		}
	}
	if (ranges.length === 0) return source;
	const chunks: string[] = [];
	let cursor = 0;
	for (const range of ranges) {
		appendRenderedSource(chunks, source, cursor, range.start, metrics);
		arrayPush(chunks, range.replacement);
		chargeCategorizedWork(metrics, "replacementEmissions");
		cursor = range.end;
	}
	appendRenderedSource(chunks, source, cursor, source.length, metrics);
	return arrayJoin(chunks, "");
}

function appendRenderedSource(
	chunks: string[],
	source: string,
	start: number,
	end: number,
	metrics?: RedactionScanMetrics,
): void {
	if (end <= start) return;
	for (let position = start; position < end; position += 1) {
		chargeCategorizedWork(metrics, "renderedSourceUnits");
	}
	arrayPush(chunks, stringSlice(source, start, end));
}

function rangePriority(priority: PlannedRangePriority): number {
	if (priority === "recovery") return 4;
	if (priority === "private-key") return 3;
	if (priority === "credential") return 2;
	return 1;
}

type SensitiveAssignmentKind = "authorization" | "secret" | "unknown-sensitive";
type AssignmentKeyClassification = SensitiveAssignmentKind | "public";
type SensitiveAssignmentContext = "flow" | "line";
type LexicalQuote = "\"" | "'";
type FlowCloser = "}" | "]";

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
	const sourceSegments = stringSplit(stringToLowerCase(key), ".");
	const segments: string[] = [];
	for (let index = 0; index < sourceSegments.length; index += 1) {
		segments[index] = stringReplace(sourceSegments[index]!, /[-_]/g, "");
	}
	return { segments, path: arrayJoin(segments, "."), terminal: arrayAt(segments, -1) ?? "" };
}

function flowCloserForOpener(character: string | undefined): FlowCloser | undefined {
	if (character === "{") return "}";
	if (character === "[") return "]";
	return undefined;
}

function isAssignmentKeyCharacter(character: string | undefined): boolean {
	if (character === undefined) return false;
	const code = INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_CHAR_CODE_AT, character, [0]) as number;
	return (code >= 48 && code <= 57) || (code >= 65 && code <= 90) ||
		(code >= 97 && code <= 122) || character === "_" || character === "-" || character === ".";
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

export function normalizeScopedPrefixes(prefixes: unknown, description: string): string[] {
	const capturedPrefixes = capturePolicyArray(prefixes, `${description} prefixes`, 64, false);
	const normalized: string[] = [];
	for (const prefix of capturedPrefixes) {
		if (prefix === ".") {
			arrayPush(normalized, prefix);
			continue;
		}
		if (typeof prefix !== "string" || prefix.length < 1 || prefix.length > MAX_PATH_CHARACTERS ||
			/[\u0000-\u001f\u007f\\]/.test(prefix) || stringStartsWith(prefix, "/") || stringIncludes(prefix, "..")) {
			throw new ToolPolicyError(`${description} prefix is invalid`);
		}
		const value = stringReplace(posix.normalize(prefix), /\/$/, "");
		if (value === "." || matchesSensitivePathPattern(value)) {
			throw new ToolPolicyError(`${description} prefix grants sensitive or workspace-root authority`);
		}
		arrayPush(normalized, value);
	}
	if (new Set(normalized).size !== normalized.length) throw new ToolPolicyError(`duplicate ${description} prefix`);
	INTRINSIC_OBJECT_FREEZE(normalized);
	return normalized;
}

function capturePolicyArray<T>(
	value: unknown,
	description: string,
	maximum: number,
	allowEmpty: boolean,
): T[] {
	if (!INTRINSIC_IS_ARRAY(value) || INTRINSIC_IS_PROXY(value)) {
		throw new ToolPolicyError(`${description} must be a bounded non-proxy array`);
	}
	if (INTRINSIC_GET_PROTOTYPE_OF(value) !== INTRINSIC_ARRAY_PROTOTYPE) {
		throw new ToolPolicyError(`${description} must use the exact Array prototype`);
	}
	const lengthDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, "length");
	const length = lengthDescriptor && "value" in lengthDescriptor ? lengthDescriptor.value : undefined;
	if (!lengthDescriptor || lengthDescriptor.get || lengthDescriptor.set || !("value" in lengthDescriptor) ||
		typeof length !== "number" || !INTRINSIC_NUMBER_IS_SAFE_INTEGER(length) || length < (allowEmpty ? 0 : 1) ||
		length > maximum) {
		throw new ToolPolicyError(`${description} has an invalid authoritative length`);
	}
	const captured: T[] = [];
	for (let index = 0; index < length; index += 1) {
		const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, intrinsicString(index));
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
			throw new ToolPolicyError(`${description} contains a sparse or accessor element`);
		}
		captured[index] = descriptor.value as T;
	}
	INTRINSIC_OBJECT_FREEZE(captured);
	return captured;
}

function captureCapabilityResult(name: string, result: CapabilityResult): Readonly<Required<CapabilityResult>> {
	if (!result || typeof result !== "object" || INTRINSIC_IS_PROXY(result)) {
		throw new ToolPolicyError(`${name} returned an invalid result`);
	}
	const resultPrototype = INTRINSIC_GET_PROTOTYPE_OF(result);
	if (resultPrototype !== INTRINSIC_OBJECT_PROTOTYPE && resultPrototype !== null) {
		throw new ToolPolicyError(`${name} result must use an exact approved prototype`);
	}
	const fields = captureOwnResultFields(result, ["status", "summary", "references"], `${name} result`);
	const status = fields.get("status");
	const summarySource = fields.get("summary");
	const referencesSource = fields.get("references");
	if (typeof status !== "string" || !arrayIncludes(["ok", "blocked", "failed"] as const, status)) {
		throw new ToolPolicyError(`${name} returned an invalid status`);
	}
	const summary = boundedString(summarySource, `${name} summary`, MAX_CAPABILITY_SUMMARY_CHARACTERS, false);
	const references: string[] = [];
	if (referencesSource !== undefined) {
		if (!INTRINSIC_IS_ARRAY(referencesSource) || INTRINSIC_IS_PROXY(referencesSource) || referencesSource.length > MAX_REFERENCES) {
			throw new ToolPolicyError(`${name} returned too many references`);
		}
		if (INTRINSIC_GET_PROTOTYPE_OF(referencesSource) !== INTRINSIC_ARRAY_PROTOTYPE) {
			throw new ToolPolicyError(`${name} references must use the exact Array prototype`);
		}
		for (let index = 0; index < referencesSource.length; index += 1) {
			const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(referencesSource, intrinsicString(index));
			if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
				throw new ToolPolicyError(`${name} returned sparse or accessor references`);
			}
			const reference = descriptor.value;
			arrayPush(references, boundedString(reference, `${name} reference`, MAX_REFERENCE_CHARACTERS, false));
		}
	}
	INTRINSIC_OBJECT_FREEZE(references);
	return INTRINSIC_OBJECT_FREEZE({ status: status as CapabilityResult["status"], summary, references });
}

function mutationResult(result: WorkspaceMutationResult, maxBytes: number): SessionToolResult {
	if (!result || typeof result !== "object" || INTRINSIC_IS_PROXY(result)) {
		throw new ToolPolicyError("workspace mutation returned an invalid result");
	}
	const resultPrototype = INTRINSIC_GET_PROTOTYPE_OF(result);
	if (resultPrototype !== INTRINSIC_OBJECT_PROTOTYPE && resultPrototype !== null) {
		throw new ToolPolicyError("workspace mutation result must use an exact approved prototype");
	}
	const fields = captureOwnResultFields(result, ["changed", "summary"], "workspace mutation result");
	const changed = fields.get("changed");
	const summarySource = fields.get("summary");
	if (typeof changed !== "boolean") throw new ToolPolicyError("workspace mutation returned an invalid result");
	const summary = boundedString(summarySource, "workspace mutation summary", MAX_CAPABILITY_SUMMARY_CHARACTERS, false);
	return textResult(INTRINSIC_JSON_STRINGIFY({ changed, summary: redactSensitiveText(summary) }), maxBytes);
}

function captureOwnResultFields(
	value: object,
	allowed: readonly string[],
	description: string,
): ReadonlyMap<string, unknown> {
	const allowedSet = new Set(allowed);
	const fields = new Map<string, unknown>();
	const enumerableFields = INTRINSIC_OBJECT_KEYS(value);
	if (enumerableFields.length > allowed.length) {
		throw new ToolPolicyError(`${description} contains unknown fields`);
	}
	for (const key of enumerableFields) {
		if (!allowedSet.has(key)) {
			throw new ToolPolicyError(`${description} contains an unknown field`);
		}
	}
	// Fixed host envelopes are projected through allowlisted descriptors. Hidden and
	// symbol peers are deliberately discarded without materializing the source key set.
	for (const key of allowed) {
		const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, key);
		if (!descriptor) continue;
		if (!descriptor.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
			throw new ToolPolicyError(`${description} contains an invalid field`);
		}
		fields.set(key, descriptor.value);
	}
	return fields;
}

function sanitizedToolBoundaryError(operation: string, error: unknown): ToolPolicyError {
	let source = "external operation failed";
	try {
		if (!INTRINSIC_IS_PROXY(error) &&
			INTRINSIC_REFLECT_APPLY(INTRINSIC_IS_NATIVE_ERROR, undefined, [error])) {
			const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(error, "message");
			const message = descriptor && "value" in descriptor ? descriptor.value : undefined;
			if (descriptor && !descriptor.get && !descriptor.set && typeof message === "string" && message.length > 0) {
				source = stringSlice(message, 0, 4_096);
			}
		}
	} catch {
		// Hostile error accessors are reduced to the stable fallback.
	}
	const safe = stringSlice(stringReplace(
		redactSensitiveText(source),
		/[\u0000-\u001f\u007f-\u009f\u061c\u200e\u200f\u2028-\u202e\u2066-\u2069]/g,
		" ",
	), 0, 2_048);
	const cause = new INTRINSIC_ERROR(safe || "external operation failed");
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
	INTRINSIC_OBJECT_FREEZE(content[0]);
	INTRINSIC_OBJECT_FREEZE(content);
	return INTRINSIC_OBJECT_FREEZE({ content, details: null });
}

function closedObject(properties: Record<string, unknown>, required: string[]): Readonly<Record<string, unknown>> {
	for (const value of INTRINSIC_OBJECT_VALUES(properties)) {
		if (value && typeof value === "object") INTRINSIC_OBJECT_FREEZE(value);
	}
	INTRINSIC_OBJECT_FREEZE(properties);
	INTRINSIC_OBJECT_FREEZE(required);
	return INTRINSIC_OBJECT_FREEZE({ type: "object", additionalProperties: false, properties, required });
}

function assertOnlyKeys(value: Readonly<Record<string, unknown>>, allowed: readonly string[], description: string): void {
	if (!isRecord(value)) throw new ToolPolicyError(`${description} input must be an object`);
	if (INTRINSIC_IS_PROXY(value)) throw new ToolPolicyError(`${description} input cannot be a Proxy`);
	const prototype = INTRINSIC_GET_PROTOTYPE_OF(value);
	if (prototype !== INTRINSIC_OBJECT_PROTOTYPE && prototype !== null) {
		throw new ToolPolicyError(`${description} input must use an exact approved prototype`);
	}
	const allowedSet = new Set(allowed);
	const keys = INTRINSIC_OBJECT_KEYS(value);
	if (keys.length > allowed.length) throw new ToolPolicyError(`${description} input contains too many fields`);
	for (const key of keys) {
		if (!allowedSet.has(key)) {
			throw new ToolPolicyError(`${description} input contains unknown field ${INTRINSIC_JSON_STRINGIFY(key)}`);
		}
	}
}

function recordParams(value: unknown, description: string): Readonly<Record<string, unknown>> {
	const snapshot = snapshotJsonData(
		value,
		0,
		{ nodes: 0, keys: 0, bytes: 0, maximumBytes: MAX_TOOL_INPUT_BYTES },
		new INTRINSIC_WEAK_SET<object>(),
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
	if (!INTRINSIC_NUMBER_IS_SAFE_INTEGER(value) || (value as number) < min || (value as number) > max) {
		throw new ToolPolicyError(`${name} must be a bounded integer`);
	}
	return value as number;
}

function boundedPositiveInteger(value: number, name: string, maximum: number): number {
	if (!INTRINSIC_NUMBER_IS_SAFE_INTEGER(value) || value <= 0 || value > maximum) {
		throw new ToolPolicyError(`${name} must be a positive safe integer within the embedded maximum ${maximum}`);
	}
	return value;
}

function assertSignal(signal: AbortSignal | undefined): void {
	if (signal === undefined) return;
	if (!(signal instanceof AbortSignal) || INTRINSIC_OBJECT_HAS_OWN(signal, "aborted") || typeof NATIVE_ABORTED_GETTER !== "function") {
		throw new ToolPolicyError("tool execution signal is invalid");
	}
	if (INTRINSIC_REFLECT_APPLY(NATIVE_ABORTED_GETTER, signal, [])) throw new ToolPolicyError("tool execution was cancelled");
}

function validIdentifier(value: unknown): value is string {
	return typeof value === "string" && /^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$/.test(value);
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null && !INTRINSIC_IS_ARRAY(value);
}

function byteLength(value: string): number {
	let bytes = 0;
	for (let index = 0; index < value.length; index += 1) {
		const code = INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_CHAR_CODE_AT, value, [index]) as number;
		if (code <= 0x7f) bytes += 1;
		else if (code <= 0x7ff) bytes += 2;
		else if (isHighSurrogate(code) && index + 1 < value.length) {
			const second = INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_CHAR_CODE_AT, value, [index + 1]) as number;
			if (isLowSurrogate(second)) {
				bytes += 4;
				index += 1;
			} else bytes += 3;
		} else bytes += 3;
	}
	return bytes;
}
