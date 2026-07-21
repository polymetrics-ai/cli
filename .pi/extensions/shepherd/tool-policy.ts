import { posix } from "node:path";

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
	validatePolicyInput(input);

	const readPrefixes = normalizeScopedPrefixes(input.authority.readPrefixes, "read");
	const writePrefixes = input.readOnly && input.authority.writePrefixes.length === 0
		? input.authority.writePrefixes
		: normalizeScopedPrefixes(input.authority.writePrefixes, "write");
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

export interface RedactionScanMetrics {
	lineBoundaryVisits: number;
	keyStartVisits?: number;
	totalCharacterVisits?: number;
}

export function redactSensitiveText(value: string): string;
export function redactSensitiveText(value: string, metrics: RedactionScanMetrics): string;
export function redactSensitiveText(value: string, metrics?: RedactionScanMetrics | number): string {
	if (typeof value !== "string") return "[REDACTED]";
	const scanMetrics = typeof metrics === "object" && metrics !== null ? metrics : undefined;
	return redactStructuredAssignments(redactPrivateKeyBlocks(value), scanMetrics);
}

type SensitiveAssignmentKind = "authorization" | "secret";
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
}

interface SensitiveAssignment {
	kind: SensitiveAssignmentKind;
	context: SensitiveAssignmentContext;
	keyColumn: number;
	normalizedKey: string;
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

interface QuotedAssignmentKey {
	key: string;
	next: number;
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

		const context: SensitiveAssignmentContext = state.flowClosers.length > 0 ? "flow" : "line";
		const allowKey = state.index === state.structuredKeyStart ||
			(context === "flow" && isFlowMappingKeyStart(value, state.index, state));
		const assignment = sensitiveAssignmentAt(value, state.index, state.lineStart, allowKey, context, metrics);
		if (!assignment) {
			if (isCommentStart(value, state.index, state.lineStart)) {
				advanceScannerIndex(state, state.lineEnd, metrics);
				continue;
			}
			if ((character === '"' || character === "'") && isQuotedSegmentStart(value, state.index, state.lineStart)) {
				state.mode = {
					kind: "quoted",
					quote: character,
					multiline: state.flowClosers.length > 0 || isStructuredValueQuote(value, state.index, state, metrics),
				};
				advanceScannerIndex(state, state.index + 1, metrics);
				continue;
			}
			const closer = flowCloserForOpener(value, state.index, state, metrics);
			if (closer) state.flowClosers.push(closer);
			else if (character === state.flowClosers[state.flowClosers.length - 1]) state.flowClosers.pop();
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
	allowKey: boolean,
	context: SensitiveAssignmentContext,
	metrics?: RedactionScanMetrics,
): SensitiveAssignment | undefined {
	if (!allowKey) return undefined;
	let cursor = index;
	let key: string;
	const quote = value[cursor] === '"' || value[cursor] === "'" ? value[cursor] : undefined;
	if (quote) {
		const decoded = decodeQuotedAssignmentKey(value, cursor, quote as LexicalQuote, metrics);
		if (!decoded) return undefined;
		key = decoded.key;
		cursor = decoded.next;
	} else {
		const keyStart = cursor;
		while (cursor < value.length && isAssignmentKeyCharacter(value[cursor]) && cursor - keyStart <= 64) {
			cursor += 1;
			recordTotalVisits(metrics, 1);
		}
		if (cursor === keyStart) return undefined;
		key = value.slice(keyStart, cursor);
	}

	const kind = sensitiveAssignmentKind(key);
	if (!kind) return undefined;
	const normalizedKey = key.toLowerCase().replace(/[-_]/g, "");
	while (isHorizontalWhitespace(value[cursor])) {
		cursor += 1;
		recordTotalVisits(metrics, 1);
	}
	if (value[cursor] !== ":" && value[cursor] !== "=") return undefined;
	cursor += 1;
	recordTotalVisits(metrics, 1);
	while (isHorizontalWhitespace(value[cursor])) {
		cursor += 1;
		recordTotalVisits(metrics, 1);
	}
	return { kind, context, keyColumn: index - lineStart, normalizedKey, valueStart: cursor };
}

function decodeQuotedAssignmentKey(
	value: string,
	start: number,
	quote: LexicalQuote,
	metrics?: RedactionScanMetrics,
): QuotedAssignmentKey | undefined {
	const MAX_DECODED_KEY_CHARACTERS = 64;
	const MAX_ENCODED_KEY_CHARACTERS = 384;
	let cursor = start + 1;
	let decoded = "";
	let malformed = false;
	let sensitivePrefix = false;
	while (cursor < value.length && cursor - start <= MAX_ENCODED_KEY_CHARACTERS) {
		const character = value[cursor];
		recordTotalVisits(metrics, 1);
		if (character === quote) {
			if (quote === "'" && value[cursor + 1] === "'") {
				decoded += "'";
				cursor += 2;
				recordTotalVisits(metrics, 1);
				continue;
			}
			if (decoded.length === 0 || decoded.length > MAX_DECODED_KEY_CHARACTERS) return undefined;
			const valid = [...decoded].every(isAssignmentKeyCharacter);
			if (!valid && !(malformed && sensitivePrefix)) return undefined;
			return { key: malformed && sensitivePrefix ? sensitiveKeyPrefix(decoded) : decoded, next: cursor + 1 };
		}
		if (character === "\n" || character === "\r") return undefined;
		if (quote === '"' && character === "\\") {
			const escaped = value[cursor + 1];
			if (escaped === "u") {
				const hex = value.slice(cursor + 2, cursor + 6);
				if (/^[0-9a-fA-F]{4}$/.test(hex)) {
					decoded += String.fromCharCode(Number.parseInt(hex, 16));
					cursor += 6;
					recordTotalVisits(metrics, 5);
					continue;
				}
				malformed = true;
				sensitivePrefix ||= sensitiveAssignmentKind(decoded) !== undefined;
				cursor += 2;
				recordTotalVisits(metrics, 1);
				continue;
			}
			const simpleEscapes: Record<string, string> = {
				'"': '"', "\\": "\\", "/": "/", b: "\b", f: "\f", n: "\n", r: "\r", t: "\t",
			};
			if (escaped !== undefined && Object.hasOwn(simpleEscapes, escaped)) {
				decoded += simpleEscapes[escaped];
				cursor += 2;
				recordTotalVisits(metrics, 1);
				continue;
			}
			malformed = true;
			sensitivePrefix ||= sensitiveAssignmentKind(decoded) !== undefined;
			cursor += Math.min(2, value.length - cursor);
			continue;
		}
		decoded += character;
		if (!isAssignmentKeyCharacter(character)) {
			malformed = true;
			sensitivePrefix ||= sensitiveAssignmentKind(decoded.slice(0, -1)) !== undefined;
		}
		if (decoded.length > MAX_DECODED_KEY_CHARACTERS) {
			malformed = true;
			sensitivePrefix ||= sensitiveAssignmentKind(decoded.slice(0, MAX_DECODED_KEY_CHARACTERS)) !== undefined;
		}
		cursor += 1;
	}
	return undefined;
}

function sensitiveKeyPrefix(value: string): string {
	for (let end = Math.min(value.length, 64); end > 0; end -= 1) {
		const prefix = value.slice(0, end);
		if (sensitiveAssignmentKind(prefix)) return prefix;
	}
	return value;
}

function redactionForAssignment(
	value: string,
	assignment: SensitiveAssignment,
	lineEnd: number,
	metrics?: RedactionScanMetrics,
): RedactionDecision {
	const quote = value[assignment.valueStart];
	if (quote === '"' || quote === "'") {
		return quotedValueRedaction(value, assignment, quote, lineEnd, metrics);
	}
	if (assignment.kind === "secret" && isYamlBlockHeader(value.slice(assignment.valueStart, lineEnd))) {
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
	const quoteEnd = findQuotedValueEnd(value, contentStart, quote, metrics);
	const contentEnd = quoteEnd ?? lineEnd;
	const resumeAt = quoteEnd === undefined ? afterLineEnding(value, lineEnd) : quoteEnd + 1;
	if (assignment.kind === "secret") {
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
	const resumeAt = assignment.context === "flow" ? scalarEnd : continuation?.end ?? Math.max(lineEnd, scalarEnd);
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
	const ambiguousLineProse = assignment.context === "line" &&
		assignment.keyColumn === 0 && continuation === undefined &&
		["token", "password", "passwd", "secret"].includes(assignment.normalizedKey) &&
		containsWhitespace(scalar);
	if (ambiguousLineProse) return { resumeAt: unredactedResumeAt };
	if (isPublicScalar(scalar)) return { resumeAt };
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
	if (cursor >= end) return undefined;
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

function sensitiveAssignmentKind(key: string): SensitiveAssignmentKind | undefined {
	const normalized = key.toLowerCase().replace(/[-_]/g, "");
	if (normalized === "authorization") return "authorization";
	if (secretAssignmentKeys.has(normalized) || normalized === "awssecretaccesskey" ||
		normalized.endsWith("token") || normalized.endsWith("password") ||
		normalized.endsWith("secret") || normalized.endsWith("apikey") ||
		normalized.endsWith("privatekey") || normalized.endsWith("databaseurl")) {
		return "secret";
	}
	return undefined;
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
	return isPotentialAssignmentKeyStart(value[cursor]) ? cursor : undefined;
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
	if (state.flowClosers[state.flowClosers.length - 1] !== "}" || !isPotentialAssignmentKeyStart(value[index])) {
		return false;
	}
	let cursor = index - 1;
	while (cursor >= state.lineStart && isHorizontalWhitespace(value[cursor])) cursor -= 1;
	return value[cursor] === "{" || value[cursor] === ",";
}

function flowCloserForOpener(
	value: string,
	index: number,
	state: StructuredScannerState,
	metrics?: RedactionScanMetrics,
): FlowCloser | undefined {
	if (value[index] === "{") {
		return state.flowClosers.length > 0 || looksLikeFlowMapping(value, index, metrics)
			? "}"
			: undefined;
	}
	if (value[index] === "[") {
		return state.flowClosers.length > 0 || looksLikeFlowSequence(value, index, metrics)
			? "]"
			: undefined;
	}
	return undefined;
}

function looksLikeFlowMapping(value: string, index: number, metrics?: RedactionScanMetrics): boolean {
	let cursor = index + 1;
	while (cursor < value.length && isWhitespace(value[cursor])) {
		cursor += 1;
		recordTotalVisits(metrics, 1);
	}
	const quote = value[cursor] === '"' || value[cursor] === "'" ? value[cursor] as LexicalQuote : undefined;
	if (quote) {
		cursor += 1;
		recordTotalVisits(metrics, 1);
		let keyLength = 0;
		while (cursor < value.length && keyLength <= 64) {
			if (quote === '"' && value[cursor] === "\\") {
				cursor += 2;
				keyLength += 1;
				recordTotalVisits(metrics, 2);
				continue;
			}
			if (value[cursor] === quote) break;
			cursor += 1;
			keyLength += 1;
			recordTotalVisits(metrics, 1);
		}
		if (keyLength === 0 || keyLength > 64 || value[cursor] !== quote) return false;
		cursor += 1;
	} else {
		const keyStart = cursor;
		while (cursor < value.length && isAssignmentKeyCharacter(value[cursor]) && cursor - keyStart <= 64) {
			cursor += 1;
			recordTotalVisits(metrics, 1);
		}
		if (cursor === keyStart || cursor - keyStart > 64) return false;
	}
	while (cursor < value.length && isHorizontalWhitespace(value[cursor])) {
		cursor += 1;
		recordTotalVisits(metrics, 1);
	}
	return value[cursor] === ":";
}

function looksLikeFlowSequence(value: string, index: number, metrics?: RedactionScanMetrics): boolean {
	let cursor = index + 1;
	while (cursor < value.length && isWhitespace(value[cursor])) {
		cursor += 1;
		recordTotalVisits(metrics, 1);
	}
	return value[cursor] === "{" || value[cursor] === "[";
}

function isStructuredValueQuote(
	value: string,
	index: number,
	state: StructuredScannerState,
	metrics?: RedactionScanMetrics,
): boolean {
	const keyStart = state.structuredKeyStart;
	if (keyStart === undefined || keyStart >= index) return false;
	let cursor = keyStart;
	const keyQuote = value[cursor] === '"' || value[cursor] === "'" ? value[cursor] : undefined;
	if (keyQuote) {
		cursor += 1;
		recordTotalVisits(metrics, 1);
		while (cursor < index && value[cursor] !== keyQuote) {
			cursor += 1;
			recordTotalVisits(metrics, 1);
		}
		if (value[cursor] !== keyQuote) return false;
		cursor += 1;
		recordTotalVisits(metrics, 1);
	} else {
		const unquotedKeyStart = cursor;
		while (cursor < index && isAssignmentKeyCharacter(value[cursor])) {
			cursor += 1;
			recordTotalVisits(metrics, 1);
		}
		if (cursor === unquotedKeyStart) return false;
	}
	while (cursor < index && isHorizontalWhitespace(value[cursor])) {
		cursor += 1;
		recordTotalVisits(metrics, 1);
	}
	if (value[cursor] !== ":" && value[cursor] !== "=") return false;
	cursor += 1;
	recordTotalVisits(metrics, 1);
	while (cursor < index && isHorizontalWhitespace(value[cursor])) {
		cursor += 1;
		recordTotalVisits(metrics, 1);
	}
	return cursor === index;
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
		(code >= 97 && code <= 122) || character === "_" || character === "-";
}

function isPotentialAssignmentKeyStart(character: string | undefined): boolean {
	return isAssignmentKeyCharacter(character) || character === '"' || character === "'";
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

function findQuotedValueEnd(
	value: string,
	start: number,
	quote: string,
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
		if (value[index] === quote) return index;
	}
	return undefined;
}

function findLineEnd(value: string, start: number, metrics?: RedactionScanMetrics): number {
	let index = start;
	while (index < value.length && value[index] !== "\n" && value[index] !== "\r") {
		if (metrics) {
			metrics.lineBoundaryVisits += 1;
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
	recordTotalVisits(metrics, Math.max(0, nextIndex - state.index));
	state.index = nextIndex;
}

function recordKeyStartVisit(metrics: RedactionScanMetrics | undefined, count = 1): void {
	if (!metrics || count <= 0) return;
	metrics.keyStartVisits = (metrics.keyStartVisits ?? 0) + count;
	recordTotalVisits(metrics, count);
}

function recordTotalVisits(metrics: RedactionScanMetrics | undefined, count: number): void {
	if (!metrics || count <= 0) return;
	metrics.totalCharacterVisits = (metrics.totalCharacterVisits ?? 0) + count;
}

function validateCapabilityName(name: string): void {
	if (typeof name !== "string" || !/^host_[a-z][a-z0-9_]{1,63}$/.test(name)) {
		throw new ToolPolicyError(`capability name ${JSON.stringify(name)} must use the bounded host_ namespace`);
	}
	if (forbiddenCapabilityPatterns.some((pattern) => pattern.test(name))) {
		throw new ToolPolicyError(`capability ${name} requests forbidden generic, secret, or recursive authority`);
	}
}

export function normalizeScopedPrefixes(prefixes: unknown, description: string): string[] {
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
	if (Object.isFrozen(prefixes) && normalized.every((value, index) => value === prefixes[index])) {
		return prefixes as string[];
	}
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

function boundedPositiveInteger(value: number, name: string, maximum: number): number {
	if (!Number.isSafeInteger(value) || value <= 0 || value > maximum) {
		throw new ToolPolicyError(`${name} must be a positive safe integer within the embedded maximum ${maximum}`);
	}
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
