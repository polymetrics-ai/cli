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
const INTRINSIC_OBJECT_KEYS = Object.keys;
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
	projectArguments(
		name: string,
		raw: unknown,
		maximumBytes?: number,
	): Readonly<Record<string, unknown>>;
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
	const hostContracts = new Map<HostCapabilityName, CompiledToolArgumentContract>();
	for (const capability of capabilities) {
		if (!declared.has(capability.name)) {
			throw new ToolPolicyError(`undeclared capability ${JSON.stringify(capability.name)} cannot expand authority`);
		}
		if (capturedInput.readOnly && capability.mutates) continue;
		tools.push(hostCapabilityTool(capability, limits));
		hostContracts.set(capability.name, capability.argumentContract);
	}

	for (const tool of tools) Object.freeze(tool);
	const names = tools.map((tool) => {
		if (!isSessionToolName(tool.name)) throw new ToolPolicyError("tool policy constructed an unregistered tool identity");
		return tool.name;
	});
	Object.freeze(names);
	Object.freeze(tools);
	return Object.freeze({
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

function validatePolicyInput(input: ToolPolicyInput): CompiledHostCapability[] {
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
	const capabilities: CompiledHostCapability[] = [];
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
		const argumentContract = compileToolArgumentContract(parameters, name);
		capabilities.push(Object.freeze({
			name,
			description,
			mutates,
			parameters,
			argumentContract,
			execute(input: Readonly<Record<string, unknown>>, signal?: AbortSignal) {
				return Reflect.apply(execute, capability, [input, signal]);
			},
		}) as CompiledHostCapability);
	}
	for (const name of declared) {
		if (!supplied.has(name)) throw new ToolPolicyError(`declared capability ${name} was not supplied`);
	}
	return capabilities;
}

type CompiledSchemaProjector = (value: unknown, description: string, depth: number) => JsonData;

function compileToolArgumentContract(schema: PlainJsonSchema, name: string): CompiledToolArgumentContract {
	const projectValue = compileSchemaProjector(schema, `${name} parameters`, 0);
	return Object.freeze({
		schema,
		project(raw: unknown, maximumBytes: number): Readonly<Record<string, unknown>> {
			if (!Number.isSafeInteger(maximumBytes) || maximumBytes < 1) {
				throw new ToolPolicyError(`${name} input has an invalid byte bound`);
			}
			const effectiveMaximum = Math.min(maximumBytes, MAX_TOOL_INPUT_BYTES);
			const projected = projectValue(raw, `${name} input`, 0);
			if (!isRecord(projected)) throw new ToolPolicyError(`${name} input must be an object`);
			if (byteLength(JSON.stringify(projected)) > effectiveMaximum) {
				throw new ToolPolicyError(`${name} input exceeded its bound`);
			}
			return projected;
		},
	});
}

function compileSchemaProjector(
	schema: PlainJsonSchema,
	description: string,
	depth: number,
): CompiledSchemaProjector {
	if (depth > MAX_CAPABILITY_SCHEMA_DEPTH) throw new ToolPolicyError(`${description} exceeded its depth bound`);
	const type = ownSchemaField(schema, "type", description);
	const enumSource = optionalOwnSchemaField(schema, "enum", description);
	const enumValues = enumSource === undefined ? undefined : captureCompiledEnum(enumSource, description);
	if (type === "string") {
		const minimum = optionalSchemaInteger(schema, "minLength", 0, MAX_WRITE_CHARACTERS, description);
		const maximum = optionalSchemaInteger(schema, "maxLength", 0, MAX_WRITE_CHARACTERS, description);
		if (minimum !== undefined && maximum !== undefined && minimum > maximum) {
			throw new ToolPolicyError(`${description} has inverted string bounds`);
		}
		return (value, valueDescription) => {
			if (typeof value !== "string" || (minimum !== undefined && value.length < minimum) ||
				(maximum !== undefined && value.length > maximum) || !enumAccepts(enumValues, value)) {
				throw new ToolPolicyError(`${valueDescription} violates its string schema`);
			}
			return value;
		};
	}
	if (type === "integer" || type === "number") {
		const minimum = optionalSchemaFiniteNumber(schema, "minimum", description);
		const maximum = optionalSchemaFiniteNumber(schema, "maximum", description);
		if (minimum !== undefined && maximum !== undefined && minimum > maximum) {
			throw new ToolPolicyError(`${description} has inverted numeric bounds`);
		}
		return (value, valueDescription) => {
			if (typeof value !== "number" || !Number.isFinite(value) || (type === "integer" && !Number.isSafeInteger(value)) ||
				(minimum !== undefined && value < minimum) || (maximum !== undefined && value > maximum) ||
				!enumAccepts(enumValues, value)) {
				throw new ToolPolicyError(`${valueDescription} violates its numeric schema`);
			}
			return value;
		};
	}
	if (type === "boolean") {
		return (value, valueDescription) => {
			if (typeof value !== "boolean" || !enumAccepts(enumValues, value)) {
				throw new ToolPolicyError(`${valueDescription} violates its boolean schema`);
			}
			return value;
		};
	}
	if (type === "array") {
		const items = ownSchemaField(schema, "items", description);
		if (!isRecord(items)) throw new ToolPolicyError(`${description} requires an array item schema`);
		const itemProjector = compileSchemaProjector(items, `${description} items`, depth + 1);
		const minimum = optionalSchemaInteger(schema, "minItems", 0, MAX_CAPABILITY_SCHEMA_ARRAY_ITEMS, description) ?? 0;
		const maximum = optionalSchemaInteger(schema, "maxItems", 0, MAX_CAPABILITY_SCHEMA_ARRAY_ITEMS, description) ??
			MAX_CAPABILITY_SCHEMA_ARRAY_ITEMS;
		if (minimum > maximum) throw new ToolPolicyError(`${description} has inverted array bounds`);
		return (value, valueDescription, valueDepth) => {
			if (!Array.isArray(value) || nodeTypes.isProxy(value) || Object.getPrototypeOf(value) !== Array.prototype) {
				throw new ToolPolicyError(`${valueDescription} must be an exact array`);
			}
			const length = ownArrayLength(value, valueDescription);
			if (length < minimum || length > maximum) throw new ToolPolicyError(`${valueDescription} violates its array bounds`);
			const projected: JsonData[] = [];
			for (let index = 0; index < length; index += 1) {
				const descriptor = Object.getOwnPropertyDescriptor(value, String(index));
				if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
					throw new ToolPolicyError(`${valueDescription} contains a sparse or accessor item`);
				}
				projected.push(itemProjector(descriptor.value, `${valueDescription}[${index}]`, valueDepth + 1));
			}
			return Object.freeze(projected) as JsonData[];
		};
	}
	if (type !== "object" || ownSchemaField(schema, "additionalProperties", description) !== false) {
		throw new ToolPolicyError(`${description} uses an unsupported schema type`);
	}
	const properties = ownSchemaField(schema, "properties", description);
	const requiredSource = optionalOwnSchemaField(schema, "required", description);
	const required = requiredSource === undefined ? Object.freeze([]) : requiredSource;
	if (!isRecord(properties)) throw new ToolPolicyError(`${description} requires a properties record`);
	const propertyPrototype = Object.getPrototypeOf(properties);
	if (nodeTypes.isProxy(properties) || (propertyPrototype !== Object.prototype && propertyPrototype !== null)) {
		throw new ToolPolicyError(`${description} properties must use an exact approved prototype`);
	}
	if (!Array.isArray(required) || nodeTypes.isProxy(required) || Object.getPrototypeOf(required) !== Array.prototype) {
		throw new ToolPolicyError(`${description} required fields must be an exact array`);
	}
	const requiredLength = ownArrayLength(required, `${description} required fields`);
	if (requiredLength > 64) throw new ToolPolicyError(`${description} has too many required fields`);
	const names: string[] = [];
	const projectors: CompiledSchemaProjector[] = [];
	const unique = new Set<string>();
	for (let index = 0; index < requiredLength; index += 1) {
		const descriptor = Object.getOwnPropertyDescriptor(required, String(index));
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor) ||
			typeof descriptor.value !== "string" || descriptor.value.length < 1 || descriptor.value.length > 128 ||
			unique.has(descriptor.value)) {
			throw new ToolPolicyError(`${description} has an invalid required field vector`);
		}
		const propertyName = descriptor.value;
		unique.add(propertyName);
		const propertyDescriptor = Object.getOwnPropertyDescriptor(properties, propertyName);
		if (!propertyDescriptor?.enumerable || propertyDescriptor.get || propertyDescriptor.set ||
			!("value" in propertyDescriptor) || !isRecord(propertyDescriptor.value)) {
			throw new ToolPolicyError(`${description} required property ${JSON.stringify(propertyName)} is invalid`);
		}
		names.push(propertyName);
		projectors.push(compileSchemaProjector(propertyDescriptor.value, `${description}.${propertyName}`, depth + 1));
	}
	Object.freeze(names);
	Object.freeze(projectors);
	return (value, valueDescription, valueDepth) => {
		if (!isRecord(value) || nodeTypes.isProxy(value)) throw new ToolPolicyError(`${valueDescription} must be an object`);
		const prototype = Object.getPrototypeOf(value);
		if (prototype !== Object.prototype && prototype !== null) {
			throw new ToolPolicyError(`${valueDescription} must use an exact approved prototype`);
		}
		const suppliedNames = INTRINSIC_OBJECT_KEYS(value);
		if (suppliedNames.length !== names.length || suppliedNames.some((field) => !unique.has(field))) {
			throw new ToolPolicyError(`${valueDescription} contains unknown or missing fields`);
		}
		const projected = Object.create(null) as Record<string, JsonData>;
		for (let index = 0; index < names.length; index += 1) {
			const field = names[index]!;
			const descriptor = Object.getOwnPropertyDescriptor(value, field);
			if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
				throw new ToolPolicyError(`${valueDescription}.${field} must be an own data field`);
			}
			Object.defineProperty(projected, field, {
				value: projectors[index]!(descriptor.value, `${valueDescription}.${field}`, valueDepth + 1),
				enumerable: true,
				writable: false,
				configurable: false,
			});
		}
		return Object.freeze(projected);
	};
}

function ownSchemaField(schema: PlainJsonSchema, field: string, description: string): unknown {
	const descriptor = Object.getOwnPropertyDescriptor(schema, field);
	if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
		throw new ToolPolicyError(`${description} requires own schema field ${field}`);
	}
	return descriptor.value;
}

function optionalOwnSchemaField(schema: PlainJsonSchema, field: string, description: string): unknown {
	const descriptor = Object.getOwnPropertyDescriptor(schema, field);
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
	if (!Number.isSafeInteger(value) || Number(value) < minimum || Number(value) > maximum) {
		throw new ToolPolicyError(`${description} has an invalid ${field}`);
	}
	return Number(value);
}

function optionalSchemaFiniteNumber(schema: PlainJsonSchema, field: string, description: string): number | undefined {
	const value = optionalOwnSchemaField(schema, field, description);
	if (value === undefined) return undefined;
	if (typeof value !== "number" || !Number.isFinite(value)) {
		throw new ToolPolicyError(`${description} has an invalid ${field}`);
	}
	return value;
}

function captureCompiledEnum(source: unknown, description: string): readonly JsonData[] {
	if (!Array.isArray(source) || nodeTypes.isProxy(source) || Object.getPrototypeOf(source) !== Array.prototype) {
		throw new ToolPolicyError(`${description} enum must be an exact array`);
	}
	const length = ownArrayLength(source, `${description} enum`);
	if (length < 1 || length > MAX_CAPABILITY_SCHEMA_ARRAY_ITEMS) {
		throw new ToolPolicyError(`${description} enum exceeded its bound`);
	}
	const values: JsonData[] = [];
	for (let index = 0; index < length; index += 1) {
		const descriptor = Object.getOwnPropertyDescriptor(source, String(index));
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor) ||
			!(["string", "number", "boolean"].includes(typeof descriptor.value)) ||
			(typeof descriptor.value === "number" && !Number.isFinite(descriptor.value))) {
			throw new ToolPolicyError(`${description} enum contains an unsupported value`);
		}
		values.push(descriptor.value as JsonData);
	}
	return Object.freeze(values);
}

function enumAccepts(values: readonly JsonData[] | undefined, value: JsonData): boolean {
	return values === undefined || values.some((candidate) => Object.is(candidate, value));
}

function ownArrayLength(value: readonly unknown[], description: string): number {
	const descriptor = Object.getOwnPropertyDescriptor(value, "length");
	if (!descriptor || descriptor.enumerable || descriptor.get || descriptor.set || !("value" in descriptor) ||
		!Number.isSafeInteger(descriptor.value) || descriptor.value < 0) {
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
			if (Object.getPrototypeOf(value) !== Array.prototype) {
				throw new ToolPolicyError(`${description} array must use the exact Array prototype`);
			}
			if (value.length > MAX_CAPABILITY_SCHEMA_ARRAY_ITEMS) {
				throw new ToolPolicyError(`${description} array exceeded its bound`);
			}
			const lengthDescriptor = Object.getOwnPropertyDescriptor(value, "length");
			if (!lengthDescriptor || !("value" in lengthDescriptor) || lengthDescriptor.value !== value.length ||
				lengthDescriptor.enumerable || lengthDescriptor.get || lengthDescriptor.set) {
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
		// Prototype safety is established before bounded own-name discovery. Hidden peers stay
		// opaque; enumerable JSON fields are then captured descriptor by descriptor.
		const result = Object.create(null) as { [key: string]: JsonData };
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
				return textResult(JSON.stringify({
					status: result.status,
					summary: redactSensitiveText(result.summary),
					references: result.references.map(redactSensitiveText),
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
	if (!Number.isSafeInteger(maximumBytes) || maximumBytes < 1) {
		throw new ToolPolicyError(`${name} arguments have an invalid byte bound`);
	}
	const effectiveMaximum = Math.min(maximumBytes, MAX_TOOL_INPUT_BYTES);
	const allowed = name === "workspace_read"
		? ["path", "offset", "limit"] as const
		: name === "workspace_edit"
			? ["path", "oldText", "newText"] as const
			: ["path", "content"] as const;
	const projected = recordParams(raw, name);
	assertOnlyKeys(projected, allowed, name);
	if (byteLength(JSON.stringify(projected)) > effectiveMaximum) {
		throw new ToolPolicyError(`${name} arguments exceeded their byte bound`);
	}
	return projected;
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
	valuePrefixTruncated: boolean;
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
		feedStrongCredentialRecognizers(source, state, character, position, metrics);

		if (state.recovery) {
			state.recovery.range.end = position + 1;
			if (isPhysicalLineEnding(character)) finishRecoveryAtLine(state, position);
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

		const top = state.assignments.at(-1);
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

		const current = state.assignments.at(-1);
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
	metrics.totalWork += 1;
	const visits = state.visits;
	if (visits) {
		visits[position] = Math.min(255, visits[position]! + 1);
		if (visits[position]! > metrics.maxMainCursorVisits) {
			metrics.maxMainCursorVisits = visits[position]!;
		}
	}
	metrics.boundaryCharacterVisits += 1;
	metrics.totalWork += 1;
}

function initializeStreamingMetrics(metrics: RedactionScanMetrics | undefined, sourceLength: number): void {
	if (!metrics) return;
	metrics.sourceLength = sourceLength;
	metrics.cursorAdvances = 0;
	metrics.cursorRegressions = 0;
	metrics.maxMainCursorVisits = 0;
	metrics.keyCharacterVisits = 0;
	metrics.boundaryCharacterVisits = 0;
	metrics.totalWork = 0;
}

function chargeKeyTransition(metrics?: RedactionScanMetrics): void {
	if (!metrics) return;
	metrics.keyCharacterVisits += 1;
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
	state.ranges.push(range);
	if (metrics) metrics.totalWork += 1;
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
							String.fromCharCode(Number.parseInt(candidate.unicodeValue, 16)),
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
	if ((character === "," && candidate.context === "flow") || character === "}" || character === "]") {
		state.candidate = undefined;
		reprocessStreamingBoundary(state, character, position, metrics);
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
	const normalizedKey = exactKey === undefined
		? ""
		: (exactKey.toLowerCase().split(".").at(-1) ?? exactKey).replace(/[-_]/g, "");
	state.assignments.push({
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
		valuePrefixTruncated: false,
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
	else assignment.valuePrefixTruncated = true;
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
			finishOneStreamingAssignment(state, assignment, position, false);
			reprocessStreamingBoundary(state, character, position, metrics);
			return true;
		}
		if (character === "#" && isHorizontalWhitespace(source[position - 1])) {
			finishOneStreamingAssignment(state, assignment, position, false);
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
	if (assignment.context === "flow" && (character === "," || character === ";")) {
		finishOneStreamingAssignment(state, assignment, position, true);
		reprocessStreamingBoundary(state, character, position, metrics);
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
		finishOneStreamingAssignment(state, assignment, assignment.valueEnd, false);
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
	const top = state.assignments.at(-1);
	if (top?.mode === "quoted") {
		finishOneStreamingAssignment(state, top, position, true);
	} else if (top && top.compositeDepth === undefined) {
		top.pendingLineBoundary = true;
		top.valueEnd = position;
		if (top.range) top.range.end = position;
	}
	state.comment = false;
	state.lineStart = position + 1;
	if (characterStartsCrLf(source, position)) state.lineStart = position + 1;
	state.lineHasContent = false;
	state.lineCandidateSlot = true;
	state.flowCandidateSlot = false;
}

function characterStartsCrLf(source: string, position: number): boolean {
	return source[position] === "\r" && source[position + 1] === "\n";
}

function finishAssignmentsAtBoundary(
	state: RedactionStreamingState,
	position: number,
	malformedQuote: boolean,
	_metrics?: RedactionScanMetrics,
): void {
	const assignment = state.assignments.at(-1);
	if (assignment && assignment.compositeDepth === undefined) {
		finishOneStreamingAssignment(state, assignment, position, malformedQuote);
	}
}

function finishAssignmentsAtMemberBoundary(
	state: RedactionStreamingState,
	position: number,
	_metrics?: RedactionScanMetrics,
): void {
	while (true) {
		const assignment = state.assignments.at(-1);
		if (!assignment || assignment.baseDepth !== state.frames.length || assignment.compositeDepth !== undefined) return;
		finishOneStreamingAssignment(state, assignment, position, false);
	}
}

function finishOneStreamingAssignment(
	state: RedactionStreamingState,
	assignment: StreamingAssignment,
	end: number,
	malformed: boolean,
): void {
	assignment.valueEnd = Math.max(assignment.valueStart ?? end, end);
	if (assignment.range) assignment.range.end = assignment.valueEnd;
	const prefix = assignment.valuePrefix.trim();
	if (assignment.kind === "public") {
		if (assignment.range) {
			assignment.range.active = malformed && !isDocumentaryPublicMultilinePrefix(prefix);
		}
	} else if (assignment.kind === "authorization") {
		finalizeAuthorizationRange(assignment, prefix);
	} else if (assignment.range) {
		const documentary = assignment.delimiter === ":" &&
			(isBoundedDocumentaryProse(prefix) ||
				(assignment.keyColumn === 0 &&
					["token", "password", "passwd", "secret"].includes(assignment.normalizedKey) &&
					containsWhitespace(prefix)));
		const publicLiteral = assignment.kind === "secret" && isReviewedPublicLiteral(prefix);
		if (documentary || publicLiteral || prefix.length === 0) assignment.range.active = false;
	}
	const index = state.assignments.lastIndexOf(assignment);
	if (index >= 0) state.assignments.splice(index, 1);
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
	const scheme = prefix.slice(0, firstSpace).toLowerCase();
	const credentialText = prefix.slice(credential);
	const parameterized = scheme === "digest" || scheme === "signature" || scheme.startsWith("aws4-");
	if (!parameterized && containsWhitespace(credentialText)) {
		range.active = false;
		return;
	}
	range.start = (assignment.valueStart ?? range.start) + credential;
}

function isBoundedDocumentaryProse(value: string): boolean {
	const lower = value.toLowerCase();
	return lower.startsWith("number of ") || lower.startsWith("name of ") ||
		lower.startsWith("description of ") || lower.startsWith("meaning of ") ||
		lower.startsWith("definition of ") || lower.startsWith("count of ") ||
		lower.startsWith("describes ") || lower.startsWith("describe ") ||
		lower.startsWith("names ") || lower.startsWith("name ") ||
		lower.startsWith("explains ") || lower.startsWith("explain ") ||
		lower.startsWith("means ") || lower.startsWith("refers to ");
}

function isDocumentaryPublicMultilinePrefix(value: string): boolean {
	return value.toLowerCase().startsWith("the following lines document configuration vocabulary");
}

function isReviewedPublicLiteral(value: string): boolean {
	const lower = value.toLowerCase();
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
	state.frames.push({
		closer,
		openedAt: position,
		publicOwner: assignment?.kind === "public" && assignment.compositeDepth === state.frames.length + 1
			? assignment
			: undefined,
	});
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
	const frame = state.frames.at(-1);
	if (!frame || frame.closer !== character) {
		const publicRange = latestPublicRecoveryRange(state);
		beginStreamingRecovery(state, publicRange?.start ?? position, publicRange, "recovery", metrics);
		return;
	}
	const closingDepth = state.frames.length;
	state.frames.pop();
	while (true) {
		const assignment = state.assignments.at(-1);
		if (!assignment || assignment.compositeDepth !== closingDepth) break;
		finishOneStreamingAssignment(state, assignment, position + 1, false);
	}
	state.flowCandidateSlot = false;
}

function latestPublicRecoveryRange(state: RedactionStreamingState): PlannedRedactionRange | undefined {
	for (let index = state.assignments.length - 1; index >= 0; index -= 1) {
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
	range.end = Math.max(range.end, state.i);
	state.recovery = { range };
	state.candidate = undefined;
	state.assignments.length = 0;
}

function finishRecoveryAtLine(state: RedactionStreamingState, position: number): void {
	if (state.recovery) state.recovery.range.end = position;
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
	if (state.recovery) state.recovery.range.end = end;
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
		finishOneStreamingAssignment(state, assignment, end, malformed);
	}
	for (const frame of state.frames) {
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
	source: string,
	state: RedactionStreamingState,
	character: string,
	position: number,
	metrics?: RedactionScanMetrics,
): void {
	feedPrivateKeyRecognizer(state, character, position, metrics);
	feedUrlCredentialRecognizer(state, character, position, metrics);
	feedQueryCredentialRecognizer(state, character, position, metrics);
	feedPasswordCredentialRecognizer(state, character, position, metrics);
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
	if (cookie.tail.length > 48) cookie.tail = cookie.tail.slice(-48);
	if (character !== ":") return;
	if (cookie.tail.endsWith("request header cookie:") ||
		cookie.tail.endsWith("request headers cookie:") ||
		cookie.tail.endsWith("request header set-cookie:") ||
		cookie.tail.endsWith("request headers set-cookie:") ||
		cookie.tail.endsWith("response header cookie:") ||
		cookie.tail.endsWith("response headers cookie:") ||
		cookie.tail.endsWith("response header set-cookie:") ||
		cookie.tail.endsWith("response headers set-cookie:")) {
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
		if (block.tail.length > block.endMarker.length) block.tail = block.tail.slice(-block.endMarker.length);
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
			else if (header.text.endsWith(PRIVATE_KEY_HEADER_SUFFIX)) {
				const candidateLabel = header.text.slice(0, -PRIVATE_KEY_HEADER_SUFFIX.length);
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
		const code = character.charCodeAt(0);
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
	const code = character.charCodeAt(0);
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
		const normalized = query.key.replace(/[-_]/g, "").toLowerCase();
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
	const code = character.charCodeAt(0);
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
	const code = character.charCodeAt(0);
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
		if (!candidate.active || candidate.end <= candidate.start) continue;
		const start = Math.max(0, Math.min(source.length, candidate.start));
		const end = Math.max(start, Math.min(source.length, candidate.end));
		if (end <= start) continue;
		const previous = ranges.at(-1);
		if (previous && start <= previous.end) {
			previous.end = Math.max(previous.end, end);
			if (rangePriority(candidate.priority) > rangePriority(previous.priority)) {
				previous.priority = candidate.priority;
				previous.replacement = candidate.replacement;
			}
		} else {
			ranges.push({ ...candidate, start, end });
		}
		if (metrics) metrics.totalWork += 1;
	}
	if (metrics) metrics.totalWork += source.length;
	if (ranges.length === 0) return source;
	const chunks: string[] = [];
	let cursor = 0;
	for (const range of ranges) {
		chunks.push(source.slice(cursor, range.start), range.replacement);
		cursor = range.end;
	}
	chunks.push(source.slice(cursor));
	return chunks.join("");
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
	const segments = key.toLowerCase().split(".").map((segment) => segment.replace(/[-_]/g, ""));
	return { segments, path: segments.join("."), terminal: segments.at(-1) ?? "" };
}

function flowCloserForOpener(character: string | undefined): FlowCloser | undefined {
	if (character === "{") return "}";
	if (character === "[") return "]";
	return undefined;
}

function isAssignmentKeyCharacter(character: string | undefined): boolean {
	if (character === undefined) return false;
	const code = character.charCodeAt(0);
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
	if (Object.getPrototypeOf(value) !== Array.prototype) {
		throw new ToolPolicyError(`${description} must use the exact Array prototype`);
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
	const resultPrototype = Object.getPrototypeOf(result);
	if (resultPrototype !== Object.prototype && resultPrototype !== null) {
		throw new ToolPolicyError(`${name} result must use an exact approved prototype`);
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
		if (Object.getPrototypeOf(referencesSource) !== Array.prototype) {
			throw new ToolPolicyError(`${name} references must use the exact Array prototype`);
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
	const resultPrototype = Object.getPrototypeOf(result);
	if (resultPrototype !== Object.prototype && resultPrototype !== null) {
		throw new ToolPolicyError("workspace mutation result must use an exact approved prototype");
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
	if (nodeTypes.isProxy(value)) throw new ToolPolicyError(`${description} input cannot be a Proxy`);
	const prototype = Object.getPrototypeOf(value);
	if (prototype !== Object.prototype && prototype !== null) {
		throw new ToolPolicyError(`${description} input must use an exact approved prototype`);
	}
	const allowedSet = new Set(allowed);
	const keys = INTRINSIC_OBJECT_KEYS(value);
	if (keys.length > allowed.length) throw new ToolPolicyError(`${description} input contains too many fields`);
	for (const key of keys) {
		if (!allowedSet.has(key)) {
			throw new ToolPolicyError(`${description} input contains unknown field ${JSON.stringify(key)}`);
		}
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
