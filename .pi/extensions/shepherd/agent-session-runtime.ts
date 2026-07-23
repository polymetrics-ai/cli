import { posix, win32 } from "node:path";
import { types as nodeTypes } from "node:util";

import type {
	AgentSessionEvent,
	CreateAgentSessionOptions,
} from "@earendil-works/pi-coding-agent";

import { assertShepherdPiCompatibility } from "./pi-compatibility.ts";
import {
	buildRolePrompts,
	routeForRole,
	type PromptBinding,
	type ShepherdAgentRole,
	type ShepherdAgentThinking,
} from "./role-prompts.ts";
import {
	createToolPolicy,
	HOST_CAPABILITY_REGISTRY,
	isHostCapabilityName,
	normalizeScopedPrefixes,
	redactSensitiveText,
	validateScopedPath,
	type HostCapability,
	type HostCapabilityName,
	type ScopedWorkspace,
	type ToolPolicy,
	type ToolAuthority,
} from "./tool-policy.ts";

const REQUIRED_PROVIDER = "openai-codex";
const REQUIRED_MODEL = "gpt-5.6-sol";
const DEFAULT_MAX_CONCURRENCY = 4;
const DEFAULT_MAX_EVENTS = 4_096;
const DEFAULT_MAX_EVENT_BYTES = 4 * 1024 * 1024;
const DEFAULT_MAX_ASSISTANT_BYTES = 64 * 1024;
const DEFAULT_CLEANUP_TIMEOUT_MS = 5_000;
const MAX_TIMEOUT_MS = 24 * 60 * 60 * 1_000;
const MAX_CONCURRENCY = 32;
const MAX_EVENTS = 65_536;
const MAX_EVENT_BYTES = 16 * 1024 * 1024;
const MAX_ASSISTANT_BYTES = 1024 * 1024;
const MAX_CLEANUP_TIMEOUT_MS = MAX_TIMEOUT_MS;
const MAX_EVENT_DEPTH = 64;
const MAX_EVENT_NODES = 65_536;
const MAX_EVENT_ARRAY_ITEMS = 4_096;
const MAX_HANDOFF_SUMMARY_CHARACTERS = 4 * 1024;
const MAX_HANDOFF_ARRAY_ITEMS = 32;
const MAX_HANDOFF_ITEM_CHARACTERS = 2 * 1024;
const INTRINSIC_OBJECT_PROTOTYPE = Object.prototype;
const INTRINSIC_ARRAY_PROTOTYPE = Array.prototype;
const INTRINSIC_ARRAY_IS_ARRAY = Array.isArray;
const INTRINSIC_GET_PROTOTYPE_OF = Object.getPrototypeOf;
const INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR = Object.getOwnPropertyDescriptor;
const INTRINSIC_OBJECT_CREATE = Object.create;
const INTRINSIC_OBJECT_DEFINE_PROPERTY = Object.defineProperty;
const INTRINSIC_OBJECT_FREEZE = Object.freeze;
const INTRINSIC_OBJECT_HAS_OWN = Object.hasOwn;
const INTRINSIC_NUMBER = Number;
const INTRINSIC_NUMBER_IS_FINITE = Number.isFinite;
const INTRINSIC_NUMBER_IS_SAFE_INTEGER = Number.isSafeInteger;
const INTRINSIC_STRING = String;
const INTRINSIC_STRING_CHAR_CODE_AT = String.prototype.charCodeAt;
const INTRINSIC_STRING_SLICE = String.prototype.slice;
const INTRINSIC_STRING_REPLACE = String.prototype.replace;
const INTRINSIC_STRING_STARTS_WITH = String.prototype.startsWith;
const INTRINSIC_STRING_ENDS_WITH = String.prototype.endsWith;
const INTRINSIC_STRING_INCLUDES = String.prototype.includes;
const INTRINSIC_STRING_SPLIT = String.prototype.split;
const INTRINSIC_STRING_TRIM = String.prototype.trim;
const INTRINSIC_ARRAY_PUSH = Array.prototype.push;
const INTRINSIC_ARRAY_JOIN = Array.prototype.join;
const INTRINSIC_ARRAY_INCLUDES = Array.prototype.includes;
const INTRINSIC_MATH_MIN = Math.min;
const INTRINSIC_MATH_MAX = Math.max;
const INTRINSIC_ERROR = Error;
const INTRINSIC_ERROR_PROTOTYPE = Error.prototype;
const INTRINSIC_AGGREGATE_ERROR_PROTOTYPE = AggregateError.prototype;
const INTRINSIC_WEAK_SET = WeakSet;
const INTRINSIC_WEAK_SET_HAS = WeakSet.prototype.has;
const INTRINSIC_WEAK_SET_ADD = WeakSet.prototype.add;
const INTRINSIC_WEAK_SET_DELETE = WeakSet.prototype.delete;
const INTRINSIC_IS_PROXY = nodeTypes.isProxy;
const INTRINSIC_IS_PROMISE = nodeTypes.isPromise;
const INTRINSIC_IS_NATIVE_ERROR = nodeTypes.isNativeError;
const INTRINSIC_REFLECT_APPLY = Reflect.apply;
const INTRINSIC_JSON = JSON;
const INTRINSIC_JSON_PARSE = JSON.parse;
const INTRINSIC_JSON_STRINGIFY = JSON.stringify;
const INTRINSIC_PROMISE = Promise;
const INTRINSIC_PROMISE_PROTOTYPE = Promise.prototype;
const INTRINSIC_PROMISE_THEN = Promise.prototype.then;
const NATIVE_ABORTED_GETTER = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(AbortSignal.prototype, "aborted")?.get;

function intrinsicString(value: unknown): string {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING, undefined, [value]) as string;
}

function intrinsicNumber(value: unknown): number {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_NUMBER, undefined, [value]) as number;
}

function intrinsicMin(left: number, right: number): number {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_MATH_MIN, undefined, [left, right]) as number;
}

function intrinsicMax(left: number, right: number): number {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_MATH_MAX, undefined, [left, right]) as number;
}

function arrayPush<T>(value: T[], item: T): number {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_ARRAY_PUSH, value, [item]) as number;
}

function arrayJoin(value: readonly string[], separator: string): string {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_ARRAY_JOIN, value, [separator]) as string;
}

function arrayIncludes<T>(value: readonly T[], item: T): boolean {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_ARRAY_INCLUDES, value, [item]) as boolean;
}

function stringSlice(value: string, start: number, end?: number): string {
	return INTRINSIC_REFLECT_APPLY(
		INTRINSIC_STRING_SLICE,
		value,
		end === undefined ? [start] : [start, end],
	) as string;
}

function stringReplace(value: string, pattern: RegExp, replacement: string): string {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_REPLACE, value, [pattern, replacement]) as string;
}

function stringStartsWith(value: string, prefix: string): boolean {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_STARTS_WITH, value, [prefix]) as boolean;
}

function stringEndsWith(value: string, suffix: string): boolean {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_ENDS_WITH, value, [suffix]) as boolean;
}

function stringIncludes(value: string, search: string): boolean {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_INCLUDES, value, [search]) as boolean;
}

function stringSplit(value: string, separator: string | RegExp): string[] {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_SPLIT, value, [separator]) as string[];
}

function stringTrim(value: string): string {
	return INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_TRIM, value, []) as string;
}

interface RuntimeResourceLoader {
	reload(): Promise<void>;
}

interface RuntimeSessionModel {
	provider: string;
	id: string;
}

type RuntimeModel = NonNullable<CreateAgentSessionOptions["model"]>;
type RuntimeResourceLoaderOption = NonNullable<CreateAgentSessionOptions["resourceLoader"]>;
type RuntimeSessionManager = NonNullable<CreateAgentSessionOptions["sessionManager"]>;
type RuntimeSettingsManager = NonNullable<CreateAgentSessionOptions["settingsManager"]>;
type ToolArgumentProjector = ToolPolicy["projectArguments"];

export interface RuntimeAgentSession {
	model: RuntimeSessionModel | undefined;
	thinkingLevel: ShepherdAgentThinking | string;
	sessionFile: string | undefined;
	getActiveToolNames(): string[];
	getLastAssistantText(): string | undefined;
	subscribe(listener: (event: AgentSessionEvent) => void): () => void | PromiseLike<void>;
	prompt(prompt: string, options: { expandPromptTemplates: false; source: "extension" }): Promise<void>;
	abort(): Promise<void>;
	waitForIdle(): Promise<void>;
	dispose(): void | PromiseLike<void>;
}

interface RuntimeSessionResult {
	session: RuntimeAgentSession;
	extensionsResult: { extensions: unknown[]; errors: unknown[]; runtime: unknown };
	modelFallbackMessage: string | undefined;
}

/** Injected adapter over the bounded compatible public Pi createAgentSession API and in-memory services. */
export interface AgentSessionRuntimeSdk {
	version: string;
	requiredVersion?: string;
	getAgentDir(): string;
	findModel(provider: string, model: string): RuntimeModel | undefined;
	hasConfiguredAuth(model: RuntimeModel): boolean;
	createSettingsManager(settings: Record<string, unknown>, options: Record<string, unknown>): RuntimeSettingsManager;
	createSessionManager(cwd: string): RuntimeSessionManager;
	createResourceLoader(options: Record<string, unknown>): RuntimeResourceLoader & RuntimeResourceLoaderOption;
	createAgentSession(options: CreateAgentSessionOptions): Promise<RuntimeSessionResult>;
}

export interface RoleAuthority extends ToolAuthority {
	issue: number;
	branch: string;
	readOnly: boolean;
}

export interface RoleRunRequest {
	role: ShepherdAgentRole;
	task: string;
	context: string[];
	timeoutMs: number;
	deadlineAt?: number;
	signal?: AbortSignal;
	workspace: ScopedWorkspace;
	/** False permits only explicitly declared host mutation capabilities, not workspace edits. */
	workspaceMutation?: boolean;
	capabilities: HostCapability[];
	authority: RoleAuthority;
	binding: PromptBinding;
}

export interface HandoffVerification {
	name: string;
	status: "passed" | "failed" | "blocked" | "not_run";
	summary: string;
}

export interface AgentSessionHandoff extends PromptBinding {
	schemaVersion: 1;
	role: ShepherdAgentRole;
	status: "completed" | "blocked" | "failed";
	summary: string;
	observedMutation: boolean;
	changedPaths: string[];
	verification: HandoffVerification[];
	findings: string[];
}

export interface AgentSessionRuntimeOptions {
	maxConcurrency?: number;
	maxEvents?: number;
	maxEventBytes?: number;
	maxAssistantBytes?: number;
	cleanupTimeoutMs?: number;
	parentSignal?: AbortSignal;
}

export class AgentSessionRuntimeError extends INTRINSIC_ERROR {
	constructor(message: string, options?: ErrorOptions) {
		// Passing an options object even for a literal-undefined reason gives every public
		// failure a stable own `cause` field instead of erasing reasonless adapter failures.
		super(message, { cause: options?.cause });
		INTRINSIC_OBJECT_DEFINE_PROPERTY(this, "name", {
			value: "AgentSessionRuntimeError", enumerable: false, writable: true, configurable: true,
		});
		INTRINSIC_OBJECT_DEFINE_PROPERTY(this, "stack", {
			value: `AgentSessionRuntimeError: ${message}`,
			enumerable: false,
			writable: true,
			configurable: true,
		});
	}
}

class CancellationScope {
	readonly deadlineAt: number;
	readonly signal: AbortSignal;
	readonly #controller = new AbortController();
	readonly #termination: Promise<never>;
	#reject: ((error: AgentSessionRuntimeError) => void) | undefined;
	#failure: AgentSessionRuntimeError | undefined;
	#timer: ReturnType<typeof setTimeout> | undefined;
	#onCancel: (() => void) | undefined;
	#finished = false;

	constructor(deadlineAt: number, timeoutDescription: string) {
		this.deadlineAt = deadlineAt;
		this.signal = this.#controller.signal;
		this.#termination = new Promise<never>((_resolve, reject) => { this.#reject = reject; });
		this.#termination.catch(() => undefined);
		const remaining = intrinsicMax(0, deadlineAt - Date.now());
		this.#timer = setTimeout(() => this.cancel(new AgentSessionRuntimeError(timeoutDescription)), remaining);
	}

	setOnCancel(callback: () => void): void {
		this.#onCancel = callback;
		if (this.#failure) callback();
	}

	race<T>(operation: Promise<T>): Promise<T> {
		return Promise.race([operation, this.#termination]);
	}

	cancel(error: AgentSessionRuntimeError): void {
		if (this.#finished || this.#failure) return;
		this.#failure = error;
		this.#controller.abort();
		this.#reject?.(error);
		this.#onCancel?.();
	}

	assertActive(): void {
		if (this.#failure) throw this.#failure;
		if (!this.#finished && Date.now() >= this.deadlineAt) {
			this.cancel(new AgentSessionRuntimeError("AgentSession deadline expired"));
			throw this.#failure;
		}
	}

	get failure(): AgentSessionRuntimeError | undefined { return this.#failure; }

	finish(): void {
		this.#finished = true;
		if (this.#timer) clearTimeout(this.#timer);
		this.#timer = undefined;
		this.#reject = undefined;
		this.#onCancel = undefined;
	}
}

interface CapturedSessionOperation {
	readonly available: boolean;
	readonly operation?: (...args: unknown[]) => unknown;
	readonly failurePresent: boolean;
	readonly failure: unknown;
}

type PromptSettlement =
	| { readonly status: "fulfilled" }
	| { readonly status: "rejected"; readonly reason: unknown };

class OwnedSettlementCell {
	readonly promise: Promise<PromptSettlement>;
	readonly #resolve: (settlement: PromptSettlement) => void;
	#settled = false;

	constructor() {
		let resolveCell: ((settlement: PromptSettlement) => void) | undefined;
		this.promise = new INTRINSIC_PROMISE<PromptSettlement>((resolve) => { resolveCell = resolve; });
		this.#resolve = resolveCell!;
		observePromise(this.promise);
	}

	settle(settlement: PromptSettlement): void {
		if (this.#settled) return;
		this.#settled = true;
		try {
			this.#resolve(INTRINSIC_OBJECT_FREEZE(settlement));
		} catch {
			// The captured intrinsic resolver is non-throwing; preserve totality if the host mutates globals.
		}
	}
}

function observePromise(promise: Promise<unknown>): void {
	try {
		INTRINSIC_REFLECT_APPLY(INTRINSIC_PROMISE_THEN, promise, [undefined, () => undefined]);
	} catch {
		// Only exact internally-created promises reach this sink.
	}
}

function installPromptSettlementHandlers(promise: Promise<unknown>, cell: OwnedSettlementCell): void {
	const observer = INTRINSIC_REFLECT_APPLY(INTRINSIC_PROMISE_THEN, promise, [
		() => { cell.settle({ status: "fulfilled" }); },
		(reason: unknown) => { cell.settle({ status: "rejected", reason }); },
	]) as Promise<unknown>;
	observePromise(observer);
}

function adoptPromptReturn(returned: unknown, cell: OwnedSettlementCell): void {
	if (INTRINSIC_IS_PROMISE(returned)) {
		if (INTRINSIC_IS_PROXY(returned) || INTRINSIC_GET_PROTOTYPE_OF(returned) !== INTRINSIC_PROMISE_PROTOTYPE ||
			INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(returned, "constructor") !== undefined ||
			INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(returned, "then") !== undefined) {
			throw new AgentSessionRuntimeError("Pi AgentSession prompt returned an unsupported native Promise shape");
		}
		installPromptSettlementHandlers(returned, cell);
		return;
	}
	const adopted = new INTRINSIC_PROMISE<unknown>((resolve) => { resolve(returned); });
	installPromptSettlementHandlers(adopted, cell);
}

const UNCAPTURED_SESSION_OPERATION: CapturedSessionOperation = INTRINSIC_OBJECT_FREEZE({
	available: false,
	operation: undefined,
	failurePresent: false,
	failure: undefined,
});

function captureSessionOperation(session: object, name: keyof RuntimeAgentSession): CapturedSessionOperation {
	let candidate: unknown;
	try {
		candidate = (session as unknown as Record<PropertyKey, unknown>)[name];
	} catch (error) {
		return { available: false, failurePresent: true, failure: error };
	}
	if (typeof candidate !== "function") {
		return {
			available: false,
			failurePresent: true,
			failure: new AgentSessionRuntimeError(`Pi AgentSession operation ${intrinsicString(name)} is missing or invalid`),
		};
	}
	return { available: true, operation: candidate as (...args: unknown[]) => unknown, failurePresent: false, failure: undefined };
}

class OwnedSession {
	readonly #session: RuntimeAgentSession;
	readonly #abort: CapturedSessionOperation;
	readonly #dispose: CapturedSessionOperation;
	readonly #prompt: CapturedSessionOperation;
	readonly #subscribe: CapturedSessionOperation;
	readonly #waitForIdle: CapturedSessionOperation;
	readonly #getActiveToolNames: CapturedSessionOperation;
	readonly #getLastAssistantText: CapturedSessionOperation;
	#abortPromise: Promise<void> | undefined;
	#disposePromise: Promise<void> | undefined;
	#promptCell: OwnedSettlementCell | undefined;
	#unsubscribe: (() => void | PromiseLike<void>) | undefined;
	#unsubscribePromise: Promise<void> | undefined;
	#waitPromise: Promise<void> | undefined;

	constructor(
		session: RuntimeAgentSession,
		afterReentrantCallback: () => void = () => undefined,
		mayContinue: () => boolean = () => true,
	) {
		this.#session = session;
		// Abort, idle, and disposal are the mandatory cleanup root. They remain independently
		// capturable after cancellation, while prompt/subscription/validation operations stop at
		// the first lifecycle barrier.
		const captureMandatory = (name: keyof RuntimeAgentSession): CapturedSessionOperation => {
			const captured = captureSessionOperation(session, name);
			if (mayContinue()) afterReentrantCallback();
			return captured;
		};
		const captureOptional = (name: keyof RuntimeAgentSession): CapturedSessionOperation => {
			if (!mayContinue()) return UNCAPTURED_SESSION_OPERATION;
			const captured = captureSessionOperation(session, name);
			afterReentrantCallback();
			return captured;
		};
		this.#abort = captureMandatory("abort");
		this.#waitForIdle = captureMandatory("waitForIdle");
		this.#dispose = captureMandatory("dispose");
		this.#prompt = captureOptional("prompt");
		this.#subscribe = captureOptional("subscribe");
		this.#getActiveToolNames = captureOptional("getActiveToolNames");
		this.#getLastAssistantText = captureOptional("getLastAssistantText");
	}

	validationFailures(): readonly unknown[] {
		const captured = [
			this.#abort,
			this.#waitForIdle,
			this.#dispose,
			this.#prompt,
			this.#subscribe,
			this.#getActiveToolNames,
			this.#getLastAssistantText,
		];
		const failures: unknown[] = [];
		for (let index = 0; index < captured.length; index += 1) {
			if (captured[index]!.failurePresent) arrayPush(failures, captured[index]!.failure);
		}
		return failures;
	}

	activeToolNames(): unknown {
		if (!this.#getActiveToolNames.available) throw this.#getActiveToolNames.failure;
		return INTRINSIC_REFLECT_APPLY(this.#getActiveToolNames.operation!, this.#session, []);
	}

	lastAssistantText(): unknown {
		if (!this.#getLastAssistantText.available) throw this.#getLastAssistantText.failure;
		return INTRINSIC_REFLECT_APPLY(this.#getLastAssistantText.operation!, this.#session, []);
	}

	modelRoute(): RuntimeSessionModel | undefined {
		return this.#session.model;
	}

	startPrompt(value: string, options: { expandPromptTemplates: false; source: "extension" }): Promise<PromptSettlement> {
		if (this.#promptCell) {
			throw new AgentSessionRuntimeError("AgentSession prompt ownership was already acquired");
		}
		const cell = new OwnedSettlementCell();
		// Publish the owned, always-fulfilled settlement before invoking Pi. A prompt callback may
		// synchronously abort or close the runtime; those barriers must still be able to join the
		// prompt result even when the post-call activity check rejects the run.
		this.#promptCell = cell;
		if (!this.#prompt.available) {
			cell.settle({ status: "rejected", reason: this.#prompt.failure });
			return cell.promise;
		}
		try {
			const returned = INTRINSIC_REFLECT_APPLY(this.#prompt.operation!, this.#session, [value, options]);
			adoptPromptReturn(returned, cell);
		} catch (reason) {
			cell.settle({ status: "rejected", reason });
		}
		return cell.promise;
	}

	promptSettlementOnce(): Promise<PromptSettlement> {
		return this.#promptCell?.promise ?? new INTRINSIC_PROMISE((resolve) => resolve({ status: "fulfilled" }));
	}

	subscribe(listener: (event: AgentSessionEvent) => void): void {
		if (this.#unsubscribe !== undefined || this.#unsubscribePromise !== undefined) {
			throw new AgentSessionRuntimeError("AgentSession subscription ownership was already acquired");
		}
		if (!this.#subscribe.available) throw this.#subscribe.failure;
		const unsubscribe = INTRINSIC_REFLECT_APPLY(this.#subscribe.operation!, this.#session, [listener]);
		if (typeof unsubscribe !== "function") {
			throw new AgentSessionRuntimeError("AgentSession subscribe returned an invalid cleanup operation");
		}
		this.#unsubscribe = unsubscribe as () => void | PromiseLike<void>;
	}

	abortOnce(): Promise<void> {
		if (!this.#abortPromise) {
			this.#abortPromise = this.#abort.available
				? Promise.resolve().then(() => INTRINSIC_REFLECT_APPLY(this.#abort.operation!, this.#session, [])).then(() => undefined)
				: Promise.resolve();
			this.#abortPromise.catch(() => undefined);
		}
		return this.#abortPromise;
	}

	waitOnce(): Promise<void> {
		if (!this.#waitPromise) {
			this.#waitPromise = this.#waitForIdle.available
				? Promise.resolve().then(() => INTRINSIC_REFLECT_APPLY(this.#waitForIdle.operation!, this.#session, [])).then(() => undefined)
				: Promise.resolve();
			this.#waitPromise.catch(() => undefined);
		}
		return this.#waitPromise;
	}

	unsubscribeOnce(): Promise<void> {
		if (!this.#unsubscribePromise) {
			const unsubscribe = this.#unsubscribe;
			this.#unsubscribe = undefined;
			this.#unsubscribePromise = Promise.resolve().then(() => {
				if (unsubscribe) return Promise.resolve(unsubscribe());
			});
			this.#unsubscribePromise.catch(() => undefined);
		}
		return this.#unsubscribePromise;
	}

	disposeOnce(): Promise<void> {
		if (!this.#disposePromise) {
			this.#disposePromise = this.#dispose.available
				? Promise.resolve().then(() =>
					Promise.resolve(INTRINSIC_REFLECT_APPLY(this.#dispose.operation!, this.#session, []))).then(() => undefined)
				: Promise.reject(this.#dispose.failure);
			this.#disposePromise.catch(() => undefined);
		}
		return this.#disposePromise;
	}
}

async function cleanupOwnedSession(
	owned: OwnedSession,
	abort: boolean,
	timeoutMs: number,
): Promise<void> {
	const failures: unknown[] = [];
	if (abort) {
		try {
			await bounded(owned.abortOnce(), timeoutMs, "session abort");
		} catch (error) {
			arrayPush(failures, error);
		}
	}
	try {
		await bounded(owned.promptSettlementOnce(), timeoutMs, "session prompt settlement");
	} catch (error) {
		arrayPush(failures, error);
	}
	try {
		await bounded(owned.waitOnce(), timeoutMs, "session idle wait");
	} catch (error) {
		arrayPush(failures, error);
	}
	try {
		await bounded(owned.unsubscribeOnce(), timeoutMs, "session unsubscribe");
	} catch (error) {
		arrayPush(failures, error);
	}
	// Disposal owns a separate bound and remains reachable after every earlier phase.
	try {
		await bounded(owned.disposeOnce(), timeoutMs, "session dispose");
	} catch (error) {
		arrayPush(failures, error);
	}
	if (failures.length > 0) throw combineFailures(failures, "AgentSession cleanup phases failed");
}

interface CreatedSessionClaim {
	readonly owned: OwnedSession;
	validate(): void;
}

class SessionCreationOwnership {
	readonly promise: Promise<RuntimeSessionResult>;
	readonly terminal: Promise<void>;
	readonly #cleanupTimeoutMs: number;
	readonly #onCleanupFailure: (error: unknown) => void;
	readonly #captureCreated: (created: RuntimeSessionResult) => CreatedSessionClaim;
	readonly #resolveTerminal: () => void;
	#state: "pending" | "claimed" | "abandoned" | "failed" = "pending";
	#settlement:
		| { status: "fulfilled"; value: RuntimeSessionResult }
		| { status: "rejected" }
		| undefined;
	#lateCleanupStarted = false;
	#terminalSettled = false;
	#pendingAtJoinBound = false;

	constructor(
		promise: Promise<RuntimeSessionResult>,
		cleanupTimeoutMs: number,
		captureCreated: (created: RuntimeSessionResult) => CreatedSessionClaim,
		onCleanupFailure: (error: unknown) => void,
	) {
		this.promise = promise;
		this.#cleanupTimeoutMs = cleanupTimeoutMs;
		this.#captureCreated = captureCreated;
		this.#onCleanupFailure = onCleanupFailure;
		const completion = deferred();
		this.terminal = completion.promise;
		this.#resolveTerminal = completion.resolve;
		void promise.then(
			(created) => {
				this.#settlement = { status: "fulfilled", value: created };
				if (this.#state === "abandoned") this.#startLateCleanup(created);
			},
			() => {
				this.#settlement = { status: "rejected" };
				if (this.#state === "abandoned") this.#settleTerminal();
			},
		);
	}

	get pending(): boolean { return this.#state === "pending"; }

	claim(created: RuntimeSessionResult): CreatedSessionClaim {
		if (this.#state !== "pending") {
			throw new AgentSessionRuntimeError("AgentSession creation ownership was already settled");
		}
		let claim: CreatedSessionClaim;
		try {
			claim = this.#captureCreated(created);
		} catch (error) {
			this.#state = "failed";
			this.#reportFailure(error);
			this.#settleTerminal();
			throw error;
		}
		this.#state = "claimed";
		this.#settleTerminal();
		return claim;
	}

	abandon(pendingAtJoinBound = false): void {
		if (this.#state !== "pending") return;
		this.#state = "abandoned";
		this.#pendingAtJoinBound = pendingAtJoinBound;
		if (this.#settlement?.status === "fulfilled") this.#startLateCleanup(this.#settlement.value);
		else if (this.#settlement?.status === "rejected") this.#settleTerminal();
	}

	async joinForClose(): Promise<boolean> {
		if (this.#terminalSettled) return true;
		// Once every active run has joined, an unclaimed creation belongs to close.
		if (this.#state === "pending") this.abandon();
		if (this.#state === "abandoned" && this.#settlement === undefined) {
			const settlement = await settleWithin(this.promise, this.#cleanupTimeoutMs);
			if (settlement.status === "pending") return false;
		}
		// A fulfilled abandoned creation now owns internally bounded abort/idle,
		// unsubscribe, and dispose phases. Do not race those phases against a shorter
		// outer timeout: their terminal is the close join contract.
		await this.terminal;
		return true;
	}

	async joinForAbort(): Promise<boolean> {
		if (this.#pendingAtJoinBound && this.#settlement === undefined) return false;
		return this.joinForClose();
	}

	#startLateCleanup(created: RuntimeSessionResult): void {
		if (this.#lateCleanupStarted) return;
		this.#lateCleanupStarted = true;
		void this.#finishLateCreation(created);
	}

	async #finishLateCreation(created: RuntimeSessionResult): Promise<void> {
		let claim: CreatedSessionClaim | undefined;
		let acquisitionFailurePresent = false;
		let acquisitionFailure: unknown;
		let cleanupFailurePresent = false;
		let cleanupFailure: unknown;
		try {
			claim = this.#captureCreated(created);
		} catch (error) {
			acquisitionFailurePresent = true;
			acquisitionFailure = error;
		}
		if (claim) {
			try {
				claim.validate();
			} catch { /* Validation is primary and retryable after successful forced cleanup. */ }
			try {
				await this.#cleanup(claim.owned);
			} catch (error) {
				cleanupFailurePresent = true;
				cleanupFailure = error;
			}
		}
		if (acquisitionFailurePresent) this.#reportFailure(acquisitionFailure);
		if (cleanupFailurePresent) this.#reportFailure(cleanupFailure);
		this.#settleTerminal();
	}

	async #cleanup(owned: OwnedSession): Promise<void> {
		const deadlineAt = Date.now() + this.#cleanupTimeoutMs;
		const failures: unknown[] = [];
		const abort = boundedUntil(
			owned.abortOnce(),
			deadlineAt,
			"abandoned session abort",
			true,
		);
		// Let abort() start before waitForIdle(), while still bounding both by one deadline.
		await Promise.resolve();
		const idle = boundedUntil(
			owned.waitOnce(),
			deadlineAt,
			"abandoned session idle wait",
			true,
		);
		for (const settlement of await Promise.allSettled([abort, idle])) {
			if (settlement.status === "rejected") arrayPush(failures, settlement.reason);
		}
		try {
			await bounded(owned.unsubscribeOnce(), this.#cleanupTimeoutMs, "abandoned session unsubscribe", true);
		} catch (error) {
			arrayPush(failures, error);
		}
		try {
			await bounded(owned.disposeOnce(), this.#cleanupTimeoutMs, "abandoned session dispose", true);
		} catch (error) {
			arrayPush(failures, error);
		}
		if (failures.length > 0) throw combineFailures(failures, "abandoned AgentSession cleanup phases failed");
	}

	#reportFailure(error: unknown): void {
		try {
			this.#onCleanupFailure(error);
		} catch {
			// Ownership continuations are total: an observer cannot create a detached rejection.
		}
	}

	#settleTerminal(): void {
		if (this.#terminalSettled) return;
		this.#terminalSettled = true;
		this.#resolveTerminal();
	}
}

class AbortListenerLease {
	readonly #signal: AbortSignal;
	readonly #listener: () => void;
	#mayBeAttached = false;
	#released = false;

	constructor(signal: AbortSignal, listener: () => void) {
		this.#signal = signal;
		this.#listener = listener;
	}

	attach(): boolean {
		if (this.#mayBeAttached) return nativeSignalAborted(this.#signal);
		this.#mayBeAttached = true;
		try {
			EventTarget.prototype.addEventListener.call(this.#signal, "abort", this.#listener, { once: true });
			return nativeSignalAborted(this.#signal);
		} catch (error) {
			try { EventTarget.prototype.removeEventListener.call(this.#signal, "abort", this.#listener); } catch { /* Preserve primary. */ }
			this.#mayBeAttached = false;
			throw new AgentSessionRuntimeError("AbortSignal listener attach failed", { cause: error });
		}
	}

	release(): void {
		if (this.#released || !this.#mayBeAttached) return;
		this.#released = true;
		try {
			EventTarget.prototype.removeEventListener.call(this.#signal, "abort", this.#listener);
		} catch (error) {
			throw new AgentSessionRuntimeError("AbortSignal listener detach failed", { cause: error });
		}
	}
}

function nativeSignalAborted(signal: AbortSignal): boolean {
	if (typeof NATIVE_ABORTED_GETTER !== "function") {
		throw new AgentSessionRuntimeError("native AbortSignal state getter is unavailable");
	}
	return Boolean(INTRINSIC_REFLECT_APPLY(NATIVE_ABORTED_GETTER, signal, []));
}

interface MutationLease {
	issue: number;
	branch: string;
	workspaceId: string;
	workspaceCwd: string;
	writePrefixes: readonly string[];
}

interface ActiveRun {
	key: string;
	runId: string;
	readOnly: boolean;
	mutationLease?: MutationLease;
	scope: CancellationScope;
	owned?: OwnedSession;
	creation?: SessionCreationOwnership;
	done: Promise<void>;
	resolveDone(): void;
}

class RunAdmissionOwnership {
	readonly runId: string;
	readonly done: Promise<void>;
	readonly #resolveDone: () => void;
	#failure: AgentSessionRuntimeError | undefined;
	#finished = false;

	constructor(runId: string) {
		this.runId = runId;
		const completion = deferred();
		this.done = completion.promise;
		this.#resolveDone = completion.resolve;
	}

	cancel(): void {
		this.#failure ??= new AgentSessionRuntimeError(`AgentSession run ${this.runId} was aborted during admission`);
	}

	assertActive(): void {
		if (this.#failure) throw this.#failure;
	}

	finish(): void {
		if (this.#finished) return;
		this.#finished = true;
		this.#resolveDone();
	}
}

interface ProgressCapture {
	readonly authorizedToolNames: ReadonlySet<string>;
	readonly observedToolNames: Set<string>;
	eventCount: number;
	eventBytes: number;
	saturated: boolean;
	frozen: boolean;
}

interface TerminalCapture {
	messageEnd?: AssistantTerminal;
	agentEnd?: AssistantTerminal;
	messageEndCount: number;
	agentEndCount: number;
	agentEndWillRetry: boolean;
	piPhase: "initial" | "agent" | "turn" | "turn-ended" | "agent-ended" | "settled";
	piOpenMessageRole?: "user" | "assistant" | "toolResult";
	piOpenToolResult?: CapturedToolResult;
	piTurnAssistant?: AssistantTerminal;
	piSettled: boolean;
	frozen: boolean;
	contentPhases: Map<number, { kind: "text" | "thinking" | "toolCall"; phase: "open" | "ended" }>;
	authorizedToolNames: ReadonlySet<string>;
	projectArguments: ToolArgumentProjector;
	toolCalls: Map<string, CapturedToolCall>;
	stream?: AssistantTerminal;
	failure?: AgentSessionRuntimeError;
	eventCount: number;
	eventBytes: number;
}

interface CapturedToolCall {
	id: string;
	name: string;
	argsIdentity: string;
	phase: "announced" | "started" | "ended" | "result" | "closed";
	resultIdentity?: string;
	isError?: boolean;
	messageIdentity?: string;
}

interface CapturedToolResult {
	toolCallId: string;
	toolName: string;
	resultIdentity: string;
	isError: boolean;
	timestamp: number;
	identity: string;
}

interface CapturedAssistantContent {
	type: "text" | "thinking" | "toolCall";
	text?: string;
	thinking?: string;
	id?: string;
	name?: string;
	argumentsIdentity?: string;
	partialJson?: string;
	identity: string;
	terminalIdentity?: string;
}

interface AssistantTerminal {
	role: "assistant";
	api: string;
	provider: string;
	model: string;
	stopReason: string;
	timestamp: number;
	content: ReadonlyArray<Readonly<CapturedAssistantContent>>;
	envelopeIdentity: string;
	identity: string;
}

export class ShepherdAgentSessionRuntime {
	readonly #sdk: AgentSessionRuntimeSdk;
	readonly #options: Required<Omit<AgentSessionRuntimeOptions, "parentSignal">>;
	readonly #active = new Map<string, ActiveRun>();
	readonly #mutatorLeases = new Map<string, MutationLease>();
	readonly #creations = new Set<SessionCreationOwnership>();
	readonly #creationsByRunId = new Map<string, Set<SessionCreationOwnership>>();
	readonly #runAdmissions = new Map<string, Set<RunAdmissionOwnership>>();
	readonly #parentListenerLease: AbortListenerLease | undefined = undefined;
	#admissions = 0;
	#admissionsDrained: ReturnType<typeof deferred> | undefined;
	#closing = false;
	#closed = false;
	#closePromise: Promise<void> | undefined;
	#quarantineFailurePresent = false;
	#quarantineFailure: unknown;

	constructor(sdk: AgentSessionRuntimeSdk, options: AgentSessionRuntimeOptions = {}) {
		this.#sdk = sdk;
		this.#options = {
			maxConcurrency: boundedPositiveInteger(options.maxConcurrency ?? DEFAULT_MAX_CONCURRENCY, "maxConcurrency", MAX_CONCURRENCY),
			maxEvents: boundedPositiveInteger(options.maxEvents ?? DEFAULT_MAX_EVENTS, "maxEvents", MAX_EVENTS),
			maxEventBytes: boundedPositiveInteger(options.maxEventBytes ?? DEFAULT_MAX_EVENT_BYTES, "maxEventBytes", MAX_EVENT_BYTES),
			maxAssistantBytes: boundedPositiveInteger(
				options.maxAssistantBytes ?? DEFAULT_MAX_ASSISTANT_BYTES,
				"maxAssistantBytes",
				MAX_ASSISTANT_BYTES,
			),
			cleanupTimeoutMs: boundedPositiveInteger(
				options.cleanupTimeoutMs ?? DEFAULT_CLEANUP_TIMEOUT_MS,
				"cleanupTimeoutMs",
				MAX_CLEANUP_TIMEOUT_MS,
			),
		};
		const parentSignal = options.parentSignal;
		if (parentSignal !== undefined && !(parentSignal instanceof AbortSignal)) {
			throw new AgentSessionRuntimeError("parentSignal is invalid");
		}
		if (parentSignal) {
			const parentAbortListener = () => { void this.#close("parent shutdown requested").catch(() => undefined); };
			const lease = new AbortListenerLease(parentSignal, parentAbortListener);
			this.#parentListenerLease = lease;
			try {
				if (lease.attach()) parentAbortListener();
			} catch (error) {
				try { lease.release(); } catch { /* Preserve the attachment failure. */ }
				throw normalizeRuntimeError(error);
			}
		}
	}

	async run(request: RoleRunRequest): Promise<AgentSessionHandoff> {
		try {
			return await this.#run(request);
		} catch (error) {
			throw normalizeRuntimeError(error);
		}
	}

	async #run(request: RoleRunRequest): Promise<AgentSessionHandoff> {
		const releaseAdmission = this.#beginAdmission();
		let runAdmission: RunAdmissionOwnership | undefined;
		let normalizedRequest: RoleRunRequest;
		let toolPolicy: ToolPolicy;
		let thinking: ShepherdAgentThinking;
		let model: RuntimeModel;
		let active: ActiveRun;
		try {
			runAdmission = this.#registerRunAdmission(request);
			normalizedRequest = normalizeRunRequest(request);
			runAdmission.assertActive();
			this.#assertSdk();
			runAdmission.assertActive();
			// Capability schemas become bounded immutable Pi tools before any SDK lookup or await.
			toolPolicy = createToolPolicy({
				readOnly: normalizedRequest.authority.readOnly,
				workspaceMutation: normalizedRequest.workspaceMutation,
				workspace: normalizedRequest.workspace,
				authority: normalizedRequest.authority,
				capabilities: normalizedRequest.capabilities,
			});
			runAdmission.assertActive();
			const route = routeForRole(normalizedRequest.role);
			thinking = route.thinking;
			const foundModel = this.#sdk.findModel(route.provider, route.model);
			runAdmission.assertActive();
			if (!isRecord(foundModel) || foundModel.provider !== REQUIRED_PROVIDER || foundModel.id !== REQUIRED_MODEL) {
				throw new AgentSessionRuntimeError(`required model ${REQUIRED_PROVIDER}/${REQUIRED_MODEL} is unavailable; fallback is forbidden`);
			}
			model = foundModel;
			if (!this.#sdk.hasConfiguredAuth(model)) {
				throw new AgentSessionRuntimeError(`required model ${REQUIRED_PROVIDER}/${REQUIRED_MODEL} has no configured auth`);
			}
			runAdmission.assertActive();
			const effectiveDeadline = computeDeadline(normalizedRequest.timeoutMs, normalizedRequest.deadlineAt);
			active = this.#reserve(
				normalizedRequest,
				effectiveDeadline,
				normalizedRequest.deadlineAt !== undefined && effectiveDeadline === normalizedRequest.deadlineAt
					? "AgentSession deadline expired"
					: `AgentSession timed out after ${normalizedRequest.timeoutMs}ms`,
			);
		} finally {
			if (runAdmission) this.#finishRunAdmission(runAdmission);
			releaseAdmission();
		}
		const scope = active.scope;
		const externalAbort = () => scope.cancel(new AgentSessionRuntimeError("AgentSession run was cancelled by its parent signal"));
		let listenerLease: AbortListenerLease | undefined;
		let result: AgentSessionHandoff | undefined;
		let primaryFailurePresent = false;
		let primaryFailure: unknown;
		try {
			if (normalizedRequest.signal) {
				listenerLease = new AbortListenerLease(normalizedRequest.signal, externalAbort);
				if (listenerLease.attach()) externalAbort();
			}
			result = await this.#execute(normalizedRequest, thinking, model, active, toolPolicy);
		} catch (error) {
			primaryFailurePresent = true;
			primaryFailure = error;
		}
		let listenerFailurePresent = false;
		let listenerFailure: unknown;
		try {
			listenerLease?.release();
		} catch (error) {
			listenerFailurePresent = true;
			listenerFailure = error;
		} finally {
			try {
				scope.finish();
			} finally {
				this.#release(active);
			}
		}
		if (primaryFailurePresent && listenerFailurePresent) {
			throw new AgentSessionRuntimeError("AgentSession run and listener cleanup both failed", {
				cause: combineFailures([primaryFailure, listenerFailure], "AgentSession primary and listener failures"),
			});
		}
		if (listenerFailurePresent) throw listenerFailure;
		if (primaryFailurePresent) throw primaryFailure;
		if (!result) throw new AgentSessionRuntimeError("AgentSession completed without a handoff");
		return result;
	}

	async abort(runId: string): Promise<void> {
		try {
			if (!validIdentifier(runId)) throw new AgentSessionRuntimeError("abort runId is invalid");
			const admissions = [...(this.#runAdmissions.get(runId) ?? [])];
			for (const admission of admissions) admission.cancel();
			const matches: ActiveRun[] = [];
			for (const active of this.#active.values()) {
				if (active.runId === runId) arrayPush(matches, active);
			}
			const owners = new Set(this.#creationsByRunId.get(runId) ?? []);
			for (const active of matches) {
				active.scope.cancel(new AgentSessionRuntimeError(`AgentSession run ${runId} was aborted`));
				void active.owned?.abortOnce().catch(() => undefined);
				if (active.creation) owners.add(active.creation);
			}
			const pending: Promise<void>[] = [];
			for (let index = 0; index < admissions.length; index += 1) arrayPush(pending, admissions[index]!.done);
			for (let index = 0; index < matches.length; index += 1) arrayPush(pending, matches[index]!.done);
			await Promise.all(pending);
			for (const owner of this.#creationsByRunId.get(runId) ?? []) owners.add(owner);
			if (owners.size > 0) {
				const ownerJoins: Promise<boolean>[] = [];
				for (const owner of owners) arrayPush(ownerJoins, owner.joinForAbort());
				const joined = await Promise.all(ownerJoins);
				if (joined.some((complete) => !complete)) {
					const failure = new AgentSessionRuntimeError(
						`AgentSession run ${runId} creation ownership remained pending at its join bound`,
					);
					this.#setQuarantine(failure);
					throw failure;
				}
			}
		} catch (error) {
			throw normalizeRuntimeError(error);
		}
	}

	async close(): Promise<void> {
		try { await this.#close("AgentSession runtime closed"); } catch (error) { throw normalizeRuntimeError(error); }
	}

	async shutdown(): Promise<void> {
		try { await this.#close("AgentSession parent shutdown requested"); } catch (error) { throw normalizeRuntimeError(error); }
	}

	#close(reason: string): Promise<void> {
		if (!this.#closePromise) {
			this.#closing = true;
			this.#closePromise = this.#performClose(reason);
		}
		return this.#closePromise;
	}

	async #performClose(reason: string): Promise<void> {
		await this.#waitForAdmissions();
		for (const active of this.#active.values()) {
			active.scope.cancel(new AgentSessionRuntimeError(reason));
			void active.owned?.abortOnce().catch(() => undefined);
		}
		const activeCompletions: Promise<void>[] = [];
		for (const active of this.#active.values()) arrayPush(activeCompletions, active.done);
		await Promise.all(activeCompletions);
		const creationOwners = [...this.#creations];
		if (creationOwners.length > 0) {
			const creationJoins: Promise<boolean>[] = [];
			for (let index = 0; index < creationOwners.length; index += 1) {
				arrayPush(creationJoins, creationOwners[index]!.joinForClose());
			}
			const joined = await Promise.all(creationJoins);
			if (joined.some((complete) => !complete)) {
				this.#setQuarantine(new AgentSessionRuntimeError(
					"AgentSession creation remained pending during bounded close",
				));
			}
		}
		this.#closed = true;
		const failures: unknown[] = [];
		try {
			this.#parentListenerLease?.release();
		} catch (error) {
			arrayPush(failures, error);
		}
		if (this.#quarantineFailurePresent) {
			arrayPush(failures, this.#quarantineFailure);
		}
		if (failures.length > 0) {
			throw new AgentSessionRuntimeError("AgentSession runtime closed while cleanup failed", {
				cause: combineFailures(failures, "AgentSession close failures"),
			});
		}
	}

	async #execute(
		request: RoleRunRequest,
		thinking: ShepherdAgentThinking,
		model: RuntimeModel,
		active: ActiveRun,
		toolPolicy: ToolPolicy,
	): Promise<AgentSessionHandoff> {
		const scope = active.scope;
		const expectedToolNames = frozenArray([...toolPolicy.names]);
		const piToolNames = frozenArray([...expectedToolNames]);
		const piCustomTools = frozenArray([...toolPolicy.tools]);
		const prompts = buildRolePrompts({
			role: request.role,
			task: request.task,
			context: request.context,
				authority: {
					issue: request.authority.issue,
					branch: request.authority.branch,
					workspaceId: request.authority.workspaceId,
					readOnly: request.authority.readOnly,
					readPrefixes: request.authority.readPrefixes,
					writePrefixes: request.authority.writePrefixes,
					toolNames: expectedToolNames,
				binding: request.binding,
			},
		});

		let creation: SessionCreationOwnership | undefined;
		let reloadPromise: Promise<void> | undefined;
		let owned: OwnedSession | undefined;
		let primaryFailurePresent = false;
		let primaryFailure: unknown;
		let result: AgentSessionHandoff | undefined;
		const assertExecutionActive = (): void => {
			if (this.#closing || this.#closed) {
				scope.cancel(new AgentSessionRuntimeError("AgentSession runtime closed before child creation"));
			}
			scope.assertActive();
		};
		try {
			assertExecutionActive();
			const settingsManager = this.#sdk.createSettingsManager({
				defaultProvider: REQUIRED_PROVIDER,
				defaultModel: REQUIRED_MODEL,
				defaultThinkingLevel: thinking,
				compaction: { enabled: false },
				retry: { enabled: false },
				packages: [],
				extensions: [],
				skills: [],
				prompts: [],
				themes: [],
			}, { projectTrusted: false });
			assertExecutionActive();
			const sessionManager = this.#sdk.createSessionManager(request.workspace.cwd);
			assertExecutionActive();
			const resourceAgentDir = this.#sdk.getAgentDir();
			assertExecutionActive();
			const resourceLoader = this.#sdk.createResourceLoader({
				cwd: request.workspace.cwd,
				agentDir: resourceAgentDir,
				settingsManager,
				noExtensions: true,
				noSkills: true,
				noPromptTemplates: true,
				noThemes: true,
				noContextFiles: true,
				systemPrompt: prompts.systemPrompt,
			});
			assertExecutionActive();
			reloadPromise = Promise.resolve().then(() => resourceLoader.reload());
			await scope.race(reloadPromise);
			assertExecutionActive();

			const sessionAgentDir = this.#sdk.getAgentDir();
			assertExecutionActive();

			const createOptions: CreateAgentSessionOptions = {
				cwd: request.workspace.cwd,
				agentDir: sessionAgentDir,
				model,
				thinkingLevel: thinking,
				scopedModels: [{ model, thinkingLevel: thinking }],
				noTools: "all",
				tools: piToolNames,
				excludeTools: frozenArray(["bash"]),
				customTools: piCustomTools,
				resourceLoader,
				sessionManager,
				settingsManager,
			};
			assertExecutionActive();
			const createPromise = Promise.resolve().then(() => this.#sdk.createAgentSession(createOptions));
			const owner = new SessionCreationOwnership(
				createPromise,
				this.#options.cleanupTimeoutMs,
				(created) => captureCreatedSession(created, thinking, expectedToolNames, assertExecutionActive),
				(error) => {
					this.#setQuarantine(error);
				},
			);
			creation = owner;
			active.creation = owner;
			this.#creations.add(owner);
			const runCreations = this.#creationsByRunId.get(request.binding.runId) ?? new Set<SessionCreationOwnership>();
			runCreations.add(owner);
			this.#creationsByRunId.set(request.binding.runId, runCreations);
			void owner.terminal.then(() => {
				this.#creations.delete(owner);
				runCreations.delete(owner);
				if (runCreations.size === 0) this.#creationsByRunId.delete(request.binding.runId);
			});
			const created = await scope.race(createPromise);
			assertExecutionActive();
			const claim = creation.claim(created);
			owned = claim.owned;
			active.owned = owned;
			scope.setOnCancel(() => { void owned?.abortOnce().catch(() => undefined); });
			assertExecutionActive();
			claim.validate();
			assertExecutionActive();

			const progress = newProgressCapture(expectedToolNames);
			assertExecutionActive();
			owned.subscribe((event) => captureProgressEvent(progress, event, this.#options));
			assertExecutionActive();
			const promptSettlement = owned.startPrompt(prompts.userPrompt, {
				expandPromptTemplates: false,
				source: "extension",
			});
			assertExecutionActive();
			const promptResult = await scope.race(promptSettlement);
			if (promptResult.status === "rejected") throw promptResult.reason;
			assertExecutionActive();
			// Pi documents prompt() as settling only after the accepted run finishes. Raw
			// lifecycle events are progress telemetry, never success or failure authority.
			progress.frozen = true;
			await scope.race(bounded(owned.unsubscribeOnce(), this.#options.cleanupTimeoutMs, "session unsubscribe"));
			assertExecutionActive();
			const terminalText = owned.lastAssistantText();
			const terminalRoute = owned.modelRoute();
			if (terminalRoute?.provider !== REQUIRED_PROVIDER || terminalRoute.id !== REQUIRED_MODEL) {
				throw new AgentSessionRuntimeError("AgentSession terminal model routing mismatch; fallback is forbidden");
			}
			if (typeof terminalText !== "string") {
				throw new AgentSessionRuntimeError("AgentSession completed without a typed terminal handoff");
			}
			result = parseHandoff(terminalText, request, this.#options.maxAssistantBytes);
		} catch (error) {
			primaryFailurePresent = true;
			primaryFailure = error;
		}

		let cleanupFailurePresent = false;
		let cleanupFailure: unknown;
		if (reloadPromise) {
			const settlement = await settleWithin(reloadPromise, this.#options.cleanupTimeoutMs);
			if (settlement.status === "pending") {
				cleanupFailurePresent = true;
				cleanupFailure = new AgentSessionRuntimeError(
					`resource loader settlement timed out after ${this.#options.cleanupTimeoutMs}ms`,
				);
			}
		}
		if (!owned && creation?.pending) {
			try {
				const settlement = await settleWithin(creation.promise, this.#options.cleanupTimeoutMs);
				if (settlement.status === "pending") {
					creation.abandon(true);
				} else if (settlement.status === "rejected") {
					creation.abandon();
				} else {
					const claim = creation.claim(settlement.value);
					owned = claim.owned;
					active.owned = owned;
					try {
						claim.validate();
					} catch (error) {
						if (!primaryFailurePresent) {
							primaryFailurePresent = true;
							primaryFailure = error;
						}
					}
				}
			} catch (error) {
				creation.abandon();
				if (!cleanupFailurePresent) {
					cleanupFailurePresent = true;
					cleanupFailure = error;
				}
			}
		}
		if (owned) {
			try {
				await cleanupOwnedSession(
					owned,
					scope.failure !== undefined || primaryFailurePresent,
					this.#options.cleanupTimeoutMs,
				);
			} catch (error) {
				if (!cleanupFailurePresent) {
					cleanupFailurePresent = true;
					cleanupFailure = error;
				}
			}
		}
		if (cleanupFailurePresent) this.#setQuarantine(cleanupFailure);

		if (cleanupFailurePresent) {
			const cleanupSnapshot = this.#quarantineFailurePresent
				? this.#quarantineFailure
				: createErrorSnapshot("AgentSession cleanup failure was unavailable");
			throw new AgentSessionRuntimeError("AgentSession cleanup/join failed; runtime quarantined", {
				cause: primaryFailurePresent
					? combineFailures([primaryFailure, cleanupSnapshot], "AgentSession primary and cleanup failures")
					: cleanupSnapshot,
			});
		}
		if (primaryFailurePresent) throw normalizeRuntimeError(primaryFailure);
		// Cancellation may win after terminal evidence is parsed but before child settlement completes.
		// Never return otherwise-valid late evidence after close, shutdown, abort, or deadline.
		if (scope.failure) throw scope.failure;
		if (!result) throw new AgentSessionRuntimeError("AgentSession completed without a handoff");
		return result;
	}

	#reserve(request: RoleRunRequest, deadlineAt: number, timeoutDescription: string): ActiveRun {
		this.#assertOpen();
		const key = `${request.binding.runId}:${request.binding.generation}:${request.binding.laneId}`;
		if (this.#active.has(key)) throw new AgentSessionRuntimeError("run/lane/generation is already active");
		const mutationLease = request.authority.readOnly ? undefined : mutationLeaseFor(request);
		if (mutationLease && [...this.#mutatorLeases.values()].some((activeLease) => mutationLeasesCollide(activeLease, mutationLease))) {
			throw new AgentSessionRuntimeError("mutating AgentSession authority overlaps an active mutator lease");
		}
		if (this.#active.size >= this.#options.maxConcurrency) {
			throw new AgentSessionRuntimeError(`AgentSession concurrency limit ${this.#options.maxConcurrency} reached`);
		}
		const scope = new CancellationScope(deadlineAt, timeoutDescription);
		const completion = deferred();
		const active: ActiveRun = {
			key,
			runId: request.binding.runId,
			readOnly: request.authority.readOnly,
			mutationLease,
			scope,
			done: completion.promise,
			resolveDone: completion.resolve,
		};
		this.#active.set(key, active);
		if (mutationLease) this.#mutatorLeases.set(key, mutationLease);
		return active;
	}

	#release(active: ActiveRun): void {
		if (!this.#active.delete(active.key)) return;
		this.#mutatorLeases.delete(active.key);
		active.resolveDone();
	}

	#assertSdk(): void {
		try {
			assertShepherdPiCompatibility(this.#sdk.version, this.#sdk.requiredVersion);
		} catch (error) {
			throw new AgentSessionRuntimeError(
				error instanceof Error ? error.message : "AgentSession Shepherd Pi compatibility check failed",
				{ cause: error },
			);
		}
		for (const method of [
			"getAgentDir",
			"findModel",
			"hasConfiguredAuth",
			"createSettingsManager",
			"createSessionManager",
			"createResourceLoader",
			"createAgentSession",
		] as const) {
			if (typeof this.#sdk[method] !== "function") throw new AgentSessionRuntimeError(`Pi SDK is missing ${method}`);
		}
	}

	#assertOpen(): void {
		if (this.#quarantineFailurePresent) {
			throw new AgentSessionRuntimeError("AgentSession runtime is quarantined after failed cleanup", {
				cause: this.#quarantineFailure,
			});
		}
		if (this.#closing || this.#closed) throw new AgentSessionRuntimeError("AgentSession runtime is closed");
	}

	#beginAdmission(): () => void {
		this.#assertOpen();
		this.#admissions += 1;
		this.#admissionsDrained ??= deferred();
		let released = false;
		return () => {
			if (released) return;
			released = true;
			this.#admissions -= 1;
			if (this.#admissions === 0) {
				this.#admissionsDrained?.resolve();
				this.#admissionsDrained = undefined;
			}
		};
	}

	#registerRunAdmission(request: RoleRunRequest): RunAdmissionOwnership {
		const runId = captureAdmissionRunId(request);
		const admission = new RunAdmissionOwnership(runId);
		const peers = this.#runAdmissions.get(runId) ?? new Set<RunAdmissionOwnership>();
		peers.add(admission);
		this.#runAdmissions.set(runId, peers);
		return admission;
	}

	#finishRunAdmission(admission: RunAdmissionOwnership): void {
		const peers = this.#runAdmissions.get(admission.runId);
		peers?.delete(admission);
		if (peers?.size === 0) this.#runAdmissions.delete(admission.runId);
		admission.finish();
	}

	async #waitForAdmissions(): Promise<void> {
		if (this.#admissions === 0) return;
		await this.#admissionsDrained?.promise;
	}

	#setQuarantine(error: unknown): void {
		if (this.#quarantineFailurePresent) return;
		this.#quarantineFailurePresent = true;
		this.#quarantineFailure = snapshotRuntimeFailure(error);
	}
}

function captureAdmissionRunId(request: RoleRunRequest): string {
	if (!request || typeof request !== "object" || INTRINSIC_ARRAY_IS_ARRAY(request) || INTRINSIC_IS_PROXY(request)) {
		throw new AgentSessionRuntimeError("AgentSession request must be an object");
	}
	const bindingDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(request, "binding");
	if (!bindingDescriptor?.enumerable || bindingDescriptor.get || bindingDescriptor.set || !("value" in bindingDescriptor) ||
		!bindingDescriptor.value || typeof bindingDescriptor.value !== "object" ||
		INTRINSIC_ARRAY_IS_ARRAY(bindingDescriptor.value) || INTRINSIC_IS_PROXY(bindingDescriptor.value)) {
		throw new AgentSessionRuntimeError("request binding must be an own data field");
	}
	const runIdDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(bindingDescriptor.value, "runId");
	if (!runIdDescriptor?.enumerable || runIdDescriptor.get || runIdDescriptor.set || !("value" in runIdDescriptor) ||
		!validIdentifier(runIdDescriptor.value)) {
		throw new AgentSessionRuntimeError("request runId must be an own data field");
	}
	return runIdDescriptor.value;
}

function normalizeRunRequest(request: RoleRunRequest): RoleRunRequest {
	const requestFields = captureKnownRecordFields(request, "AgentSession request", [
		"role", "task", "context", "timeoutMs", "deadlineAt", "signal", "workspace", "workspaceMutation", "capabilities", "authority", "binding",
		// Recognized legacy authority aliases are explicitly denied without discovering arbitrary peers.
		"provider", "model", "thinking", "tools", "issue", "workspaceId",
	]);
	assertAllowedCapturedFields(requestFields, [
		"role", "task", "context", "timeoutMs", "deadlineAt", "signal", "workspace", "workspaceMutation", "capabilities", "authority", "binding",
	], ["role", "task", "context", "timeoutMs", "workspace", "capabilities", "authority", "binding"], "request");
	const role = requestFields.get("role") as ShepherdAgentRole;
	const task = requestFields.get("task") as string;
	const contextSource = requestFields.get("context");
	const timeoutMs = requestFields.get("timeoutMs") as number;
	const deadlineAt = requestFields.get("deadlineAt") as number | undefined;
	const signal = requestFields.get("signal") as AbortSignal | undefined;
	const workspaceSource = requestFields.get("workspace");
	const workspaceMutationSource = requestFields.get("workspaceMutation");
	const capabilitiesSource = requestFields.get("capabilities");
	const authoritySource = requestFields.get("authority");
	const bindingSource = requestFields.get("binding");

	routeForRole(role);
	if (!INTRINSIC_NUMBER_IS_SAFE_INTEGER(timeoutMs) || timeoutMs <= 0 || timeoutMs > MAX_TIMEOUT_MS) {
		throw new AgentSessionRuntimeError("timeoutMs must be a positive bounded safe integer");
	}
	if (deadlineAt !== undefined && (!INTRINSIC_NUMBER_IS_SAFE_INTEGER(deadlineAt) || deadlineAt <= Date.now())) {
		throw new AgentSessionRuntimeError("deadlineAt must be a future epoch-millisecond safe integer");
	}
	if (signal !== undefined && !(signal instanceof AbortSignal)) {
		throw new AgentSessionRuntimeError("request signal is invalid");
	}

	const authorityFields = captureKnownRecordFields(authoritySource, "request authority", [
		"issue", "branch", "workspaceId", "readOnly", "readPrefixes", "writePrefixes", "capabilityNames",
	]);
	assertExactCapturedFields(authorityFields, [
		"issue", "branch", "workspaceId", "readOnly", "readPrefixes", "writePrefixes", "capabilityNames",
	], "authority");
	const issue = authorityFields.get("issue");
	const branch = authorityFields.get("branch");
	const authorityWorkspaceId = authorityFields.get("workspaceId");
	const readOnly = authorityFields.get("readOnly");
	const readPrefixesSource = captureFreshDenseArray(authorityFields.get("readPrefixes"), "authority read prefixes", 64, false);
	const writePrefixesSource = captureFreshDenseArray(authorityFields.get("writePrefixes"), "authority write prefixes", 64, true);
	const capabilityNamesSource = captureFreshDenseArray(authorityFields.get("capabilityNames"), "authority capability names", 32, true);
	if (typeof issue !== "number" || !INTRINSIC_NUMBER_IS_SAFE_INTEGER(issue) || issue < 1) {
		throw new AgentSessionRuntimeError("authority issue is invalid");
	}
	if (typeof branch !== "string" || branch.length < 1 || branch.length > 255 ||
		/[\u0000-\u001f\u007f]/.test(branch) || branch === "main") {
		throw new AgentSessionRuntimeError("authority branch is invalid or targets main");
	}
	if (!validIdentifier(authorityWorkspaceId)) throw new AgentSessionRuntimeError("authority workspace identity is invalid");
	if (typeof readOnly !== "boolean") throw new AgentSessionRuntimeError("authority readOnly is invalid");
	const workspaceMutation = workspaceMutationSource === undefined ? !readOnly : workspaceMutationSource;
	if (typeof workspaceMutation !== "boolean" || (readOnly && workspaceMutation)) {
		throw new AgentSessionRuntimeError("workspaceMutation conflicts with authority readOnly");
	}
	const readPrefixes = frozenArray(normalizeScopedPrefixes(readPrefixesSource, "read"));
	const writePrefixes = readOnly && writePrefixesSource.length === 0
		? frozenArray<string>([])
		: frozenArray(normalizeScopedPrefixes(writePrefixesSource, "write"));
	const capabilityNames: HostCapabilityName[] = [];
	for (const name of capabilityNamesSource) {
		if (!isHostCapabilityName(name)) {
			throw new AgentSessionRuntimeError("authority capability name is outside the closed host registry");
		}
		arrayPush(capabilityNames, name);
	}
	INTRINSIC_OBJECT_FREEZE(capabilityNames);
	const authority = INTRINSIC_OBJECT_FREEZE({
		issue,
		branch,
		workspaceId: authorityWorkspaceId,
		readOnly,
		readPrefixes,
		writePrefixes,
		capabilityNames,
	}) as RoleAuthority;

	const bindingFields = captureKnownRecordFields(bindingSource, "request binding", [
		"runId", "generation", "laneId", "candidateHead", "validationNonce",
	]);
	assertExactCapturedFields(bindingFields, ["runId", "generation", "laneId", "candidateHead", "validationNonce"], "binding");
	const runId = bindingFields.get("runId");
	const generation = bindingFields.get("generation");
	const laneId = bindingFields.get("laneId");
	const candidateHead = bindingFields.get("candidateHead");
	const validationNonce = bindingFields.get("validationNonce");
	if (!validIdentifier(runId) || !validIdentifier(laneId) ||
		typeof generation !== "number" || !INTRINSIC_NUMBER_IS_SAFE_INTEGER(generation) || generation < 1 ||
		typeof candidateHead !== "string" || !/^[0-9a-f]{40}$/.test(candidateHead) ||
		!validIdentifier(validationNonce) || validationNonce.length < 12) {
		throw new AgentSessionRuntimeError("request binding is invalid");
	}
	const binding = INTRINSIC_OBJECT_FREEZE({ runId, generation, laneId, candidateHead, validationNonce });

	const workspaceFields = captureKnownRecordFields(workspaceSource, "workspace capability", [
		"id", "cwd", "readText", "editText", "writeText",
	]);
	assertExactCapturedFields(workspaceFields, ["id", "cwd", "readText", "editText", "writeText"], "workspace capability");
	const workspaceId = workspaceFields.get("id");
	const workspaceCwd = workspaceFields.get("cwd");
	const readText = workspaceFields.get("readText");
	const editText = workspaceFields.get("editText");
	const writeText = workspaceFields.get("writeText");
	if (workspaceId !== authorityWorkspaceId || !isAbsoluteNonTraversingPath(workspaceCwd) ||
		typeof readText !== "function" || typeof editText !== "function" || typeof writeText !== "function") {
		throw new AgentSessionRuntimeError("workspace identity, cwd, or capability does not match the immutable authority envelope");
	}
	const canonicalCwd = canonicalWorkspacePath(workspaceCwd);
	const workspace = INTRINSIC_OBJECT_FREEZE({
		id: workspaceId,
		cwd: canonicalCwd,
		readText(path: string, options: { offset?: number; limit?: number; signal?: AbortSignal }) {
			return INTRINSIC_REFLECT_APPLY(readText, workspaceSource as object, [path, options]);
		},
		editText(path: string, oldText: string, newText: string, operationSignal?: AbortSignal) {
			return INTRINSIC_REFLECT_APPLY(editText, workspaceSource as object, [path, oldText, newText, operationSignal]);
		},
		writeText(path: string, content: string, operationSignal?: AbortSignal) {
			return INTRINSIC_REFLECT_APPLY(writeText, workspaceSource as object, [path, content, operationSignal]);
		},
	}) satisfies ScopedWorkspace;

	const capabilityValues = captureFreshDenseArray(capabilitiesSource, "typed host capabilities", 32, true);
	const capabilities: HostCapability[] = [];
	for (const capability of capabilityValues) {
		arrayPush(capabilities, normalizeCapability(capability as HostCapability));
	}
	INTRINSIC_OBJECT_FREEZE(capabilities);
	const contextValues = captureFreshDenseArray(contextSource, "role context", 64, true);
	const context: string[] = [];
	for (const entry of contextValues) {
		if (typeof entry !== "string") throw new AgentSessionRuntimeError("role context item is invalid");
		arrayPush(context, entry);
	}
	INTRINSIC_OBJECT_FREEZE(context);
	const normalized = INTRINSIC_OBJECT_FREEZE({
		role,
		task,
		context,
		timeoutMs,
		deadlineAt,
		signal,
		workspace,
		workspaceMutation,
		capabilities,
		authority,
		binding,
	}) as RoleRunRequest;

	// Prompt construction performs byte-level task/context and authority bounds before any SDK call.
	buildRolePrompts({
		role,
		task,
		context,
		authority: {
			issue: authority.issue,
			branch: authority.branch,
			workspaceId: authority.workspaceId,
			readOnly: authority.readOnly,
			readPrefixes: authority.readPrefixes,
			writePrefixes: authority.writePrefixes,
			toolNames: [],
			binding,
		},
	});
	return normalized;
}

function normalizeCapability(capability: HostCapability): HostCapability {
	const fields = captureKnownRecordFields(capability, "capability", [
		"name", "description", "mutates", "parameters", "execute",
	]);
	assertExactCapturedFields(fields, ["name", "description", "mutates", "parameters", "execute"], "capability");
	const name = fields.get("name");
	const description = fields.get("description") as string;
	const mutates = fields.get("mutates");
	const parameters = fields.get("parameters") as HostCapability["parameters"];
	const execute = fields.get("execute");
	if (!isHostCapabilityName(name) || mutates !== HOST_CAPABILITY_REGISTRY[name].mutates ||
		typeof execute !== "function") {
		throw new AgentSessionRuntimeError("capability identity, mutability, or execute contract is invalid");
	}
	return INTRINSIC_OBJECT_FREEZE({
		name,
		description,
		mutates,
		parameters,
		execute(input: Readonly<Record<string, unknown>>, signal?: AbortSignal) {
			return INTRINSIC_REFLECT_APPLY(execute, capability, [input, signal]);
		},
	}) as HostCapability;
}

function captureFreshDenseArray(
	value: unknown,
	description: string,
	maximum: number,
	allowEmpty: boolean,
): unknown[] {
	if (!INTRINSIC_ARRAY_IS_ARRAY(value) || INTRINSIC_IS_PROXY(value)) {
		throw new AgentSessionRuntimeError(`${description} must be a non-proxy array`);
	}
	if (INTRINSIC_GET_PROTOTYPE_OF(value) !== INTRINSIC_ARRAY_PROTOTYPE) {
		throw new AgentSessionRuntimeError(`${description} must use the exact Array prototype`);
	}
	const lengthDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, "length");
	const lengthValue = lengthDescriptor && "value" in lengthDescriptor ? lengthDescriptor.value : undefined;
	if (!lengthDescriptor || lengthDescriptor.get || lengthDescriptor.set || !("value" in lengthDescriptor) ||
		typeof lengthValue !== "number" || !INTRINSIC_NUMBER_IS_SAFE_INTEGER(lengthValue) || lengthValue < (allowEmpty ? 0 : 1) ||
		lengthValue > maximum) {
		throw new AgentSessionRuntimeError(`${description} has an invalid authoritative length`);
	}
	const length = lengthValue;
	const captured: unknown[] = [];
	for (let index = 0; index < length; index += 1) {
		const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, intrinsicString(index));
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
			throw new AgentSessionRuntimeError(`${description} contains a sparse or accessor element`);
		}
		captured[index] = descriptor.value;
	}
	INTRINSIC_OBJECT_FREEZE(captured);
	return captured;
}

function mutationLeaseFor(request: RoleRunRequest): MutationLease {
	return INTRINSIC_OBJECT_FREEZE({
		issue: request.authority.issue,
		branch: request.authority.branch,
		workspaceId: request.authority.workspaceId,
		workspaceCwd: request.workspace.cwd,
		writePrefixes: request.authority.writePrefixes,
	});
}

function mutationLeasesCollide(left: MutationLease, right: MutationLease): boolean {
	return left.issue === right.issue || left.branch === right.branch ||
		left.workspaceId === right.workspaceId || left.workspaceCwd === right.workspaceCwd;
}

function captureCreatedSession(
	created: RuntimeSessionResult,
	thinking: ShepherdAgentThinking,
	expectedTools: readonly string[],
	assertActive?: () => void,
): CreatedSessionClaim {
	if (!created || typeof created !== "object" || INTRINSIC_ARRAY_IS_ARRAY(created) || INTRINSIC_IS_PROXY(created)) {
		throw new AgentSessionRuntimeError("Pi returned an invalid AgentSession result");
	}
	const captureFailures: unknown[] = [];
	let captureActive = true;
	const recordCaptureFailure = (error: unknown): void => { arrayPush(captureFailures, error); };
	const afterReentrantCallback = (): void => {
		if (!captureActive || !assertActive) return;
		try {
			assertActive();
		} catch (error) {
			captureActive = false;
			recordCaptureFailure(error);
		}
	};
	const captureOptionalStep = (operation: () => void): void => {
		if (!captureActive) return;
		try {
			operation();
		} catch (error) {
			recordCaptureFailure(error);
			captureActive = false;
		}
		afterReentrantCallback();
	};
	// Acquire the cleanup root before validating any peer field. The legacy public Pi result
	// permits a session getter, so it is invoked exactly once; every other result field must be
	// a data descriptor and is never read through ordinary property lookup.
	const sessionDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(created, "session");
	let session: unknown;
	if (sessionDescriptor && "value" in sessionDescriptor) session = sessionDescriptor.value;
	else if (sessionDescriptor?.get) session = INTRINSIC_REFLECT_APPLY(sessionDescriptor.get, created, []);
	afterReentrantCallback();
	if (!session || typeof session !== "object") {
		throw new AgentSessionRuntimeError("Pi returned an invalid AgentSession result without a cleanable session");
	}
	const owned = new OwnedSession(
		session as RuntimeAgentSession,
		afterReentrantCallback,
		() => captureActive,
	);
	const operationFailures = owned.validationFailures();
	for (let index = 0; index < operationFailures.length; index += 1) {
		arrayPush(captureFailures, operationFailures[index]);
	}
	if (operationFailures.length > 0) captureActive = false;
	let modelProvider: unknown;
	let modelId: unknown;
	let thinkingLevel: unknown;
	let sessionFile: unknown;
	let extensions: unknown;
	let extensionErrors: unknown;
	let rawActiveTools: unknown;
	let activeTools: readonly string[] | undefined;
	captureOptionalStep(() => {
		if (!sessionDescriptor?.enumerable || sessionDescriptor.set ||
			(!("value" in sessionDescriptor) && !sessionDescriptor.get)) {
			throw new AgentSessionRuntimeError("Pi session ownership descriptor is invalid");
		}
	});
	captureOptionalStep(() => {
		assertApprovedRecordPrototype(created, "Pi AgentSession result");
	});
	captureOptionalStep(() => {
		const extensionsDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(created, "extensionsResult");
		if (!extensionsDescriptor?.enumerable || extensionsDescriptor.get || extensionsDescriptor.set ||
			!("value" in extensionsDescriptor)) {
			throw new AgentSessionRuntimeError("Pi returned an invalid extensions result descriptor");
		}
		({ extensions, errors: extensionErrors } = captureExtensionsResult(extensionsDescriptor.value));
	});
	captureOptionalStep(() => { captureExactEmptyArray(extensions, "Pi extensions result extensions"); });
	captureOptionalStep(() => { captureExactEmptyArray(extensionErrors, "Pi extensions result errors"); });
	captureOptionalStep(() => {
		const fallbackDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(created, "modelFallbackMessage");
		if (!fallbackDescriptor?.enumerable || fallbackDescriptor.get || fallbackDescriptor.set ||
			!("value" in fallbackDescriptor)) {
			throw new AgentSessionRuntimeError("Pi returned an invalid fallback descriptor");
		}
		if (fallbackDescriptor.value !== undefined) {
			throw new AgentSessionRuntimeError("Pi attempted a forbidden model fallback");
		}
	});
	let model: RuntimeSessionModel | undefined;
	captureOptionalStep(() => { model = (session as RuntimeAgentSession).model; });
	captureOptionalStep(() => { modelProvider = model?.provider; });
	captureOptionalStep(() => { modelId = model?.id; });
	captureOptionalStep(() => {
		thinkingLevel = (session as RuntimeAgentSession).thinkingLevel;
	});
	captureOptionalStep(() => {
		sessionFile = (session as RuntimeAgentSession).sessionFile;
	});
	captureOptionalStep(() => {
		rawActiveTools = owned.activeToolNames();
	});
	captureOptionalStep(() => { activeTools = captureToolNameArray(rawActiveTools); });

	return INTRINSIC_OBJECT_FREEZE({
		owned,
		validate(): void {
			if (captureFailures.length > 0) {
				throw combineFailures(captureFailures, "Pi AgentSession capture or validation failed");
			}
			if (modelProvider !== REQUIRED_PROVIDER || modelId !== REQUIRED_MODEL) {
				throw new AgentSessionRuntimeError("embedded AgentSession model routing mismatch");
			}
			if (thinkingLevel !== thinking) throw new AgentSessionRuntimeError("embedded AgentSession thinking route mismatch");
			if (sessionFile !== undefined) throw new AgentSessionRuntimeError("embedded AgentSession persistence is forbidden");
			if (!activeTools || activeTools.length !== expectedTools.length ||
				activeTools.some((name, index) => name !== expectedTools[index])) {
				throw new AgentSessionRuntimeError("embedded AgentSession active tool authority drifted");
			}
		},
	});
}

function captureExtensionsResult(value: unknown): { extensions: unknown; errors: unknown } {
	if (!value || typeof value !== "object" || INTRINSIC_ARRAY_IS_ARRAY(value) || INTRINSIC_IS_PROXY(value)) {
		throw new AgentSessionRuntimeError("Pi returned an invalid extensions result");
	}
	assertApprovedRecordPrototype(value, "Pi extensions result");
	const captured: { extensions?: unknown; errors?: unknown } = {};
	for (const key of ["extensions", "errors"] as const) {
		const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, key);
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
			throw new AgentSessionRuntimeError(`Pi extensions result ${key} must be an own data field`);
		}
		captured[key] = descriptor.value;
	}
	const runtimeDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, "runtime");
	if (!runtimeDescriptor?.enumerable || runtimeDescriptor.get || runtimeDescriptor.set ||
		!("value" in runtimeDescriptor)) {
		throw new AgentSessionRuntimeError("Pi extensions result runtime must be an own data field");
	}
	return { extensions: captured.extensions, errors: captured.errors };
}

function captureExactEmptyArray(value: unknown, description: string): void {
	if (!INTRINSIC_ARRAY_IS_ARRAY(value) || INTRINSIC_IS_PROXY(value)) {
		throw new AgentSessionRuntimeError(`${description} must be an exact non-proxy empty array`);
	}
	assertApprovedArrayPrototype(value, description);
	const lengthDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, "length");
	if (!lengthDescriptor || lengthDescriptor.get || lengthDescriptor.set || !("value" in lengthDescriptor) ||
		lengthDescriptor.value !== 0) {
		throw new AgentSessionRuntimeError(`${description} must be empty`);
	}
}

function captureToolNameArray(value: unknown): readonly string[] | undefined {
	if (!INTRINSIC_ARRAY_IS_ARRAY(value) || INTRINSIC_IS_PROXY(value)) return undefined;
	assertApprovedArrayPrototype(value, "Pi AgentSession active tool names");
	const lengthDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, "length");
	if (!lengthDescriptor || lengthDescriptor.get || lengthDescriptor.set || !("value" in lengthDescriptor) ||
		typeof lengthDescriptor.value !== "number" || !INTRINSIC_NUMBER_IS_SAFE_INTEGER(lengthDescriptor.value) ||
		lengthDescriptor.value < 0 || lengthDescriptor.value > 256) {
		return undefined;
	}
	const length = lengthDescriptor.value;
	const names: string[] = [];
	for (let index = 0; index < length; index += 1) {
		const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, intrinsicString(index));
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor) ||
			typeof descriptor.value !== "string") {
			return undefined;
		}
		arrayPush(names, descriptor.value);
	}
	return INTRINSIC_OBJECT_FREEZE(names);
}

function assertApprovedRecordPrototype(value: object, description: string): void {
	const prototype = INTRINSIC_GET_PROTOTYPE_OF(value);
	if (prototype !== null && prototype !== INTRINSIC_OBJECT_PROTOTYPE) {
		throw new AgentSessionRuntimeError(`${description} has a non-approved direct prototype`);
	}
}

function assertApprovedArrayPrototype(value: unknown[], description: string): void {
	if (INTRINSIC_GET_PROTOTYPE_OF(value) !== INTRINSIC_ARRAY_PROTOTYPE) {
		throw new AgentSessionRuntimeError(`${description} has a non-approved direct prototype`);
	}
}

function newProgressCapture(expectedToolNames: readonly string[]): ProgressCapture {
	return {
		authorizedToolNames: new Set(expectedToolNames),
		observedToolNames: new Set(),
		eventCount: 0,
		eventBytes: 0,
		saturated: false,
		frozen: false,
	};
}

function captureProgressEvent(
	capture: ProgressCapture,
	event: unknown,
	options: Required<Omit<AgentSessionRuntimeOptions, "parentSignal">>,
): void {
	if (capture.frozen || capture.saturated) return;
	try {
		const fields = captureKnownRecordFields(event, "AgentSession progress event", ["type", "toolName"]);
		if (fields.get("type") !== "tool_execution_start") return;
		capture.eventCount += 1;
		const toolName = fields.get("toolName");
		const charge = typeof toolName === "string" ? byteLength(toolName) + 32 : 32;
		if (capture.eventCount > options.maxEvents || capture.eventBytes + charge > options.maxEventBytes) {
			capture.saturated = true;
			return;
		}
		capture.eventBytes += charge;
		if (typeof toolName === "string" && capture.authorizedToolNames.has(toolName)) {
			capture.observedToolNames.add(toolName);
		}
	} catch {
		// Unknown, malformed, or accessor-backed telemetry is non-authoritative and inert.
	}
}

function newTerminalCapture(
	expectedToolNames: readonly string[],
	projectArguments: ToolArgumentProjector,
): TerminalCapture {
	return {
		messageEndCount: 0,
		agentEndCount: 0,
		agentEndWillRetry: false,
		piPhase: "initial",
		piSettled: false,
		frozen: false,
		contentPhases: new Map(),
		authorizedToolNames: new Set(expectedToolNames),
		projectArguments,
		toolCalls: new Map(),
		eventCount: 0,
		eventBytes: 0,
	};
}

function captureEvent(
	capture: TerminalCapture,
	event: unknown,
	options: Required<Omit<AgentSessionRuntimeOptions, "parentSignal">>,
): void {
	if (capture.failure) return;
	if (capture.frozen) {
		capture.failure = new AgentSessionRuntimeError("AgentSession emitted an event after its settled boundary");
		return;
	}
	capture.eventCount += 1;
	if (capture.eventCount > options.maxEvents) {
		capture.failure = new AgentSessionRuntimeError("AgentSession event stream exceeded its bound");
		return;
	}
	try {
		const eventFields = captureKnownRecordFields(event, "AgentSession event", [
			"type", "message", "assistantMessageEvent", "toolResults", "messages", "willRetry",
			"toolCallId", "toolName", "args", "partialResult", "result", "isError",
		]);
		const eventType = eventFields.get("type");
		if (typeof eventType !== "string") throw new AgentSessionRuntimeError("AgentSession event type is invalid");
		const directCharge = capturePiLifecycleEvent(capture, eventFields, eventType, options);
		capture.eventBytes += directCharge ?? byteLength(eventType) + 16;
		if (capture.eventBytes > options.maxEventBytes) {
			throw new AgentSessionRuntimeError("AgentSession event stream exceeded its bound");
		}
	} catch (error) {
		capture.failure = new AgentSessionRuntimeError(
			"AgentSession emitted an invalid, unbounded, or terminal-sequence event",
			{ cause: error },
		);
	}
}

function capturePiLifecycleEvent(
	capture: TerminalCapture,
	fields: ReadonlyMap<string, unknown>,
	type: string,
	options: Required<Omit<AgentSessionRuntimeOptions, "parentSignal">>,
): number | undefined {
	switch (type) {
		case "agent_start":
			assertExactCapturedFields(fields, ["type"], "agent_start event");
			if (capture.piPhase !== "initial") throw new AgentSessionRuntimeError("agent_start is out of order");
			capture.piPhase = "agent";
			return undefined;
		case "turn_start":
			assertExactCapturedFields(fields, ["type"], "turn_start event");
			if (capture.piPhase !== "agent" && capture.piPhase !== "turn-ended") {
				throw new AgentSessionRuntimeError("turn_start is out of order");
			}
			capture.piPhase = "turn";
			capture.piTurnAssistant = undefined;
			capture.stream = undefined;
			for (const call of capture.toolCalls.values()) {
				if (call.phase !== "closed") {
					throw new AgentSessionRuntimeError("turn_start preceded completion of an authorized tool call");
				}
			}
			capture.toolCalls.clear();
			return undefined;
		case "message_start": {
			assertExactCapturedFields(fields, ["type", "message"], "message_start event");
			if (capture.piPhase !== "turn" || capture.piOpenMessageRole !== undefined) {
				throw new AgentSessionRuntimeError("message_start is out of order");
			}
			const role = capturePiMessageRole(fields.get("message"));
			capture.piOpenMessageRole = role;
			if (role === "assistant") {
				if (capture.piTurnAssistant) {
					throw new AgentSessionRuntimeError("multiple assistant messages cannot replace one tool-bearing turn");
				}
				const initial = captureAssistantTerminal(
					fields.get("message"), true, options.maxEventBytes, capture.projectArguments,
				);
				if (!initial) throw new AgentSessionRuntimeError("assistant message_start is invalid");
				capture.stream = initial;
				capture.contentPhases.clear();
			} else if (role === "toolResult") {
				const toolResult = captureToolResultMessage(fields.get("message"), options.maxEventBytes);
				const call = capture.toolCalls.get(toolResult.toolCallId);
				if (!call || call.phase !== "ended" || call.name !== toolResult.toolName ||
					call.resultIdentity !== toolResult.resultIdentity || call.isError !== toolResult.isError) {
					throw new AgentSessionRuntimeError("tool result message does not match an ended authorized call");
				}
				capture.piOpenToolResult = toolResult;
			}
			return undefined;
		}
		case "message_update":
			assertExactCapturedFields(fields, ["type", "message", "assistantMessageEvent"], "message_update event");
			if (capture.piPhase !== "turn" || capture.piOpenMessageRole !== "assistant") {
				throw new AgentSessionRuntimeError("message_update is out of order");
			}
			return captureStreamingUpdateCharge(
				capture,
				fields.get("message"),
				fields.get("assistantMessageEvent"),
				options.maxEventBytes,
			);
		case "message_end": {
			assertExactCapturedFields(fields, ["type", "message"], "message_end event");
			if (capture.piPhase !== "turn" || capture.piOpenMessageRole === undefined) {
				throw new AgentSessionRuntimeError("message_end is out of order");
			}
			const role = capturePiMessageRole(fields.get("message"));
			if (role !== capture.piOpenMessageRole) {
				throw new AgentSessionRuntimeError("message_start and message_end roles disagree");
			}
			if (role === "assistant") {
				const terminal = captureAssistantTerminal(
					fields.get("message"), false, options.maxEventBytes, capture.projectArguments,
				);
				if (!terminal) throw new AgentSessionRuntimeError("assistant message_end is invalid");
				assertCompletedContentPhases(capture);
				capture.messageEndCount += 1;
				capture.messageEnd = terminal;
				capture.piTurnAssistant = terminal;
				registerAssistantToolCalls(capture, terminal);
			} else if (role === "toolResult") {
				const toolResult = captureToolResultMessage(fields.get("message"), options.maxEventBytes);
				const open = capture.piOpenToolResult;
				const call = capture.toolCalls.get(toolResult.toolCallId);
				if (!open || open.identity !== toolResult.identity || !call || call.phase !== "ended" ||
					call.name !== toolResult.toolName || call.resultIdentity !== toolResult.resultIdentity ||
					call.isError !== toolResult.isError) {
					throw new AgentSessionRuntimeError("tool result message end does not match its authorized call");
				}
				call.phase = "result";
				call.messageIdentity = toolResult.identity;
			}
			capture.piOpenMessageRole = undefined;
			capture.piOpenToolResult = undefined;
			return undefined;
		}
		case "turn_end": {
			assertExactCapturedFields(fields, ["type", "message", "toolResults"], "turn_end event");
			if (capture.piPhase !== "turn" || capture.piOpenMessageRole !== undefined || !capture.piTurnAssistant) {
				throw new AgentSessionRuntimeError("turn_end is out of order");
			}
			const terminal = captureAssistantTerminal(
				fields.get("message"), false, options.maxEventBytes, capture.projectArguments,
			);
			if (!terminal || !sameTerminal(terminal, capture.piTurnAssistant)) {
				throw new AgentSessionRuntimeError("turn_end assistant does not match message_end");
			}
			captureTurnToolResults(capture, fields.get("toolResults"), options.maxEventBytes);
			capture.piPhase = "turn-ended";
			return undefined;
		}
		case "agent_end": {
			assertExactCapturedFields(fields, ["type", "messages", "willRetry"], "agent_end event");
			if (capture.piPhase !== "turn-ended" || capture.agentEndCount !== 0) {
				throw new AgentSessionRuntimeError("agent_end is out of order");
			}
			const willRetry = fields.get("willRetry");
			if (typeof willRetry !== "boolean") throw new AgentSessionRuntimeError("agent_end willRetry is invalid");
			capture.agentEndWillRetry = willRetry;
			if (willRetry) throw new AgentSessionRuntimeError("retrying agent_end is not terminal evidence");
			let lastAssistant: AssistantTerminal | undefined;
			for (const message of captureDenseArray(fields.get("messages"), "AgentSession terminal messages")) {
				const assistant = captureAssistantTerminal(
					message, false, options.maxEventBytes, capture.projectArguments,
				);
				if (assistant) lastAssistant = assistant;
			}
			if (!lastAssistant || !capture.messageEnd || !sameTerminal(lastAssistant, capture.messageEnd)) {
				throw new AgentSessionRuntimeError("agent_end final assistant does not match message_end");
			}
			capture.agentEndCount = 1;
			capture.agentEnd = lastAssistant;
			capture.piPhase = "agent-ended";
			return undefined;
		}
		case "agent_settled":
			assertExactCapturedFields(fields, ["type"], "agent_settled event");
			if (capture.piPhase !== "agent-ended" || capture.piSettled) {
				throw new AgentSessionRuntimeError("agent_settled is out of order");
			}
			capture.piSettled = true;
			capture.piPhase = "settled";
			capture.frozen = true;
			return undefined;
		case "tool_execution_start": {
			assertExactCapturedFields(fields, ["type", "toolCallId", "toolName", "args"], "tool_execution_start event");
			if (capture.piPhase !== "turn" || capture.piOpenMessageRole !== undefined) {
				throw new AgentSessionRuntimeError("tool_execution_start is out of order");
			}
			const toolCallId = requiredLifecycleString(fields, "toolCallId", "tool execution ID");
			const toolName = requiredLifecycleString(fields, "toolName", "tool execution name");
			const argsIdentity = projectToolArguments(
				capture.projectArguments, toolName, fields.get("args"), options.maxEventBytes,
			);
			const call = capture.toolCalls.get(toolCallId);
			if (!call || call.phase !== "announced" || call.name !== toolName ||
				call.argsIdentity !== argsIdentity) {
				throw new AgentSessionRuntimeError("tool execution start does not match one authorized assistant call");
			}
			call.phase = "started";
			return undefined;
		}
		case "tool_execution_update": {
			assertExactCapturedFields(
				fields,
				["type", "toolCallId", "toolName", "args", "partialResult"],
				"tool_execution_update event",
			);
			if (capture.piPhase !== "turn" || capture.piOpenMessageRole !== undefined) {
				throw new AgentSessionRuntimeError("tool_execution_update is out of order");
			}
			const toolCallId = requiredLifecycleString(fields, "toolCallId", "tool execution ID");
			const toolName = requiredLifecycleString(fields, "toolName", "tool execution name");
			const execution = capture.toolCalls.get(toolCallId);
			const argsIdentity = projectToolArguments(
				capture.projectArguments, toolName, fields.get("args"), options.maxEventBytes,
			);
			if (!execution || execution.phase !== "started" || execution.name !== toolName ||
				execution.argsIdentity !== argsIdentity) {
				throw new AgentSessionRuntimeError("tool execution update does not match its start");
			}
			snapshotOpaqueValue(fields.get("partialResult"), "tool execution partial result", options.maxEventBytes);
			return undefined;
		}
		case "tool_execution_end": {
			assertExactCapturedFields(
				fields,
				["type", "toolCallId", "toolName", "result", "isError"],
				"tool_execution_end event",
			);
			if (capture.piPhase !== "turn" || capture.piOpenMessageRole !== undefined) {
				throw new AgentSessionRuntimeError("tool_execution_end is out of order");
			}
			const toolCallId = requiredLifecycleString(fields, "toolCallId", "tool execution ID");
			const toolName = requiredLifecycleString(fields, "toolName", "tool execution name");
			const execution = capture.toolCalls.get(toolCallId);
			if (!execution || execution.phase !== "started" || execution.name !== toolName ||
				typeof fields.get("isError") !== "boolean") {
				throw new AgentSessionRuntimeError("tool execution end does not match its start");
			}
			const resultIdentity = captureToolExecutionResultIdentity(fields.get("result"), options.maxEventBytes);
			execution.phase = "ended";
			execution.resultIdentity = resultIdentity;
			execution.isError = fields.get("isError") as boolean;
			return undefined;
		}
		default:
			throw new AgentSessionRuntimeError(`unsupported Pi AgentSession event ${INTRINSIC_JSON_STRINGIFY(type)}`);
	}
}

function registerAssistantToolCalls(capture: TerminalCapture, terminal: AssistantTerminal): void {
	const calls: Readonly<CapturedAssistantContent>[] = [];
	for (let index = 0; index < terminal.content.length; index += 1) {
		const part = terminal.content[index]!;
		if (part.type === "toolCall") arrayPush(calls, part);
	}
	if (terminal.stopReason !== "toolUse") {
		if (calls.length > 0) {
			throw new AgentSessionRuntimeError("non-tool assistant terminal contains a tool call");
		}
		return;
	}
	if (calls.length === 0 || capture.toolCalls.size !== 0) {
		throw new AgentSessionRuntimeError("tool-use assistant must originate a fresh authorized call set");
	}
	for (const part of calls) {
		if (!part.id || !part.name || part.argumentsIdentity === undefined ||
			!capture.authorizedToolNames.has(part.name) || capture.toolCalls.has(part.id)) {
			throw new AgentSessionRuntimeError("assistant originated an invalid or unauthorized tool call");
		}
		capture.toolCalls.set(part.id, {
			id: part.id,
			name: part.name,
			argsIdentity: part.argumentsIdentity,
			phase: "announced",
		});
	}
}

function captureToolExecutionResultIdentity(value: unknown, maximum: number): string {
	const fields = captureKnownRecordFields(value, "tool execution result", ["content", "details", "terminate"]);
	assertAllowedCapturedFields(
		fields,
		["content", "details", "terminate"],
		["content", "details"],
		"tool execution result",
	);
	const terminate = fields.get("terminate");
	if (terminate !== undefined && typeof terminate !== "boolean") {
		throw new AgentSessionRuntimeError("tool execution terminate flag is invalid");
	}
	return captureToolResultPayloadIdentity(
		fields.get("content"),
		fields.has("details"),
		fields.get("details"),
		maximum,
	);
}

function captureToolResultMessage(value: unknown, maximum: number): CapturedToolResult {
	const fields = captureKnownRecordFields(value, "Pi tool result message", [
		"role", "toolCallId", "toolName", "content", "details", "isError", "timestamp",
	]);
	assertAllowedCapturedFields(
		fields,
		["role", "toolCallId", "toolName", "content", "details", "isError", "timestamp"],
		["role", "toolCallId", "toolName", "content", "isError", "timestamp"],
		"Pi tool result message",
	);
	if (fields.get("role") !== "toolResult") {
		throw new AgentSessionRuntimeError("Pi tool result role is invalid");
	}
	const toolCallId = requiredLifecycleString(fields, "toolCallId", "tool result ID");
	const toolName = requiredLifecycleString(fields, "toolName", "tool result name");
	const isError = fields.get("isError");
	const timestamp = fields.get("timestamp");
	if (typeof isError !== "boolean" || typeof timestamp !== "number" || !INTRINSIC_NUMBER_IS_FINITE(timestamp)) {
		throw new AgentSessionRuntimeError("Pi tool result status or timestamp is invalid");
	}
	const resultIdentity = captureToolResultPayloadIdentity(
		fields.get("content"),
		fields.has("details"),
		fields.get("details"),
		maximum,
	);
	const identity = canonicalJson(snapshotEventJson({
		toolCallId,
		toolName,
		resultIdentity,
		isError,
		timestamp,
	}, "Pi tool result message identity", maximum));
	return INTRINSIC_OBJECT_FREEZE({ toolCallId, toolName, resultIdentity, isError, timestamp, identity });
}

function captureToolResultPayloadIdentity(
	contentSource: unknown,
	hasDetails: boolean,
	detailsSource: unknown,
	maximum: number,
): string {
	const content = snapshotToolResultContent(contentSource, maximum);
	const details = hasDetails && detailsSource !== undefined
		? snapshotOpaqueValue(detailsSource, "tool result details", maximum)
		: undefined;
	return canonicalJson(snapshotEventJson({
		content,
		...(details === undefined ? {} : { details }),
	}, "tool result payload identity", maximum));
}

function captureTurnToolResults(capture: TerminalCapture, value: unknown, maximum: number): void {
	const results = captureDenseArray(value, "turn_end tool results");
	const seen = new Set<string>();
	for (const source of results) {
		const result = captureToolResultMessage(source, maximum);
		const call = capture.toolCalls.get(result.toolCallId);
		if (seen.has(result.toolCallId) || !call || call.phase !== "result" || call.name !== result.toolName ||
			call.messageIdentity !== result.identity || call.resultIdentity !== result.resultIdentity ||
			call.isError !== result.isError) {
			throw new AgentSessionRuntimeError("turn_end tool result does not match one completed authorized call");
		}
		seen.add(result.toolCallId);
		call.phase = "closed";
	}
	const toolTurn = capture.piTurnAssistant?.stopReason === "toolUse";
	if ((toolTurn && capture.toolCalls.size === 0) || (!toolTurn && capture.toolCalls.size !== 0) ||
		seen.size !== capture.toolCalls.size) {
		throw new AgentSessionRuntimeError("turn_end does not close the assistant tool-call set exactly once");
	}
	for (const call of capture.toolCalls.values()) {
		if (call.phase !== "closed") {
			throw new AgentSessionRuntimeError("turn_end preceded complete tool-call correlation");
		}
	}
}

function requiredLifecycleString(
	fields: ReadonlyMap<string, unknown>,
	name: string,
	description: string,
): string {
	const value = fields.get(name);
	if (typeof value !== "string" || value.length === 0) {
		throw new AgentSessionRuntimeError(`${description} is invalid`);
	}
	return value;
}

function assertCompletedContentPhases(capture: TerminalCapture): void {
	for (const phase of capture.contentPhases.values()) {
		if (phase.phase !== "ended") {
			throw new AgentSessionRuntimeError("assistant content stream ended with an open phase");
		}
	}
}

function capturePiMessageRole(value: unknown): "user" | "assistant" | "toolResult" {
	const fields = captureKnownRecordFields(value, "Pi lifecycle message", ["role"]);
	const role = fields.get("role");
	if (role !== "user" && role !== "assistant" && role !== "toolResult") {
		throw new AgentSessionRuntimeError("Pi lifecycle message role is invalid");
	}
	return role;
}

function captureStreamingUpdateCharge(
	capture: TerminalCapture,
	messageValue: unknown,
	value: unknown,
	maximum: number,
): number {
	const fields = captureKnownRecordFields(value, "Pi assistant streaming event", [
		"type", "contentIndex", "partial", "delta", "content", "toolCall", "reason", "message", "error",
	]);
	const type = fields.get("type");
	if (typeof type !== "string") throw new AgentSessionRuntimeError("Pi assistant streaming event type is invalid");
	const shapes: Readonly<Record<string, readonly string[]>> = {
		start: ["type", "partial"],
		text_start: ["type", "contentIndex", "partial"],
		text_delta: ["type", "contentIndex", "delta", "partial"],
		text_end: ["type", "contentIndex", "content", "partial"],
		thinking_start: ["type", "contentIndex", "partial"],
		thinking_delta: ["type", "contentIndex", "delta", "partial"],
		thinking_end: ["type", "contentIndex", "content", "partial"],
		toolcall_start: ["type", "contentIndex", "partial"],
		toolcall_delta: ["type", "contentIndex", "delta", "partial"],
		toolcall_end: ["type", "contentIndex", "toolCall", "partial"],
		done: ["type", "reason", "message"],
		error: ["type", "reason", "error"],
	};
	const shape = shapes[type];
	if (!shape) throw new AgentSessionRuntimeError(`unsupported Pi assistant streaming event ${INTRINSIC_JSON_STRINGIFY(type)}`);
	assertExactCapturedFields(fields, shape, `Pi ${type} streaming event`);
	const message = captureAssistantTerminal(messageValue, true, maximum, capture.projectArguments);
	if (!message) throw new AgentSessionRuntimeError("message_update did not contain an assistant message");
	const innerValue = fields.has("partial") ? fields.get("partial") : fields.has("message")
		? fields.get("message") : fields.get("error");
	const inner = captureAssistantTerminal(innerValue, true, maximum, capture.projectArguments);
	if (!inner || inner.identity !== message.identity) {
		throw new AgentSessionRuntimeError(`Pi ${type} message and cumulative snapshot disagree`);
	}
	const contentIndex = fields.get("contentIndex");
	if (fields.has("contentIndex") && (!INTRINSIC_NUMBER_IS_SAFE_INTEGER(contentIndex) ||
		typeof contentIndex !== "number" || contentIndex < 0)) {
		throw new AgentSessionRuntimeError(`Pi ${type} content index is invalid`);
	}
	const variable = fields.has("delta") ? fields.get("delta") : undefined;
	if (variable !== undefined && typeof variable !== "string") {
		throw new AgentSessionRuntimeError(`Pi ${type} delta is invalid`);
	}
	if (type === "done" && !arrayIncludes(["stop", "length", "toolUse"] as const, intrinsicString(fields.get("reason")))) {
		throw new AgentSessionRuntimeError("Pi done reason is invalid");
	}
	if (type === "error" && !arrayIncludes(["aborted", "error"] as const, intrinsicString(fields.get("reason")))) {
		throw new AgentSessionRuntimeError("Pi error reason is invalid");
	}
	const growth = fields.has("contentIndex")
		? validateIndexedStreamTransition(capture, capture.stream, message, type, fields)
		: streamSnapshotGrowth(capture.stream, message);
	capture.stream = message;
	const metadata = byteLength(type) + (typeof variable === "string" ? byteLength(variable) : 0) + 32;
	if (growth + metadata > maximum) throw new AgentSessionRuntimeError("Pi streaming update exceeded its byte bound");
	return growth + metadata;
}

function validateIndexedStreamTransition(
	capture: TerminalCapture,
	previous: AssistantTerminal | undefined,
	current: AssistantTerminal,
	type: string,
	fields: ReadonlyMap<string, unknown>,
): number {
	const envelopeGrowth = cumulativeIdentityGrowth(previous?.envelopeIdentity, current.envelopeIdentity, 0);
	const index = intrinsicNumber(fields.get("contentIndex"));
	const previousContent = previous?.content ?? [];
	if (index >= current.content.length || current.content.length < previousContent.length ||
		current.content.length > previousContent.length + 1 ||
		(current.content.length === previousContent.length + 1 && index !== previousContent.length)) {
		throw new AgentSessionRuntimeError(`Pi ${type} content index skipped or replaced stream state`);
	}
	for (let cursor = 0; cursor < intrinsicMin(previousContent.length, current.content.length); cursor += 1) {
		if (cursor !== index && previousContent[cursor]?.identity !== current.content[cursor]?.identity) {
			throw new AgentSessionRuntimeError(`Pi ${type} replaced an unrelated content item`);
		}
	}
	const prior = previousContent[index];
	const next = current.content[index];
	if (!next) throw new AgentSessionRuntimeError(`Pi ${type} content item is missing`);
	const kind = stringStartsWith(type, "text_") ? "text" : stringStartsWith(type, "thinking_") ? "thinking" : "toolCall";
	const transition = stringEndsWith(type, "_start") ? "start" : stringEndsWith(type, "_delta") ? "delta" : "end";
	const ownedPhase = capture.contentPhases.get(index);
	if (transition === "start") {
		if (ownedPhase) throw new AgentSessionRuntimeError(`Pi ${type} duplicated an owned content phase`);
		capture.contentPhases.set(index, { kind, phase: "open" });
	} else if (!ownedPhase || ownedPhase.kind !== kind || ownedPhase.phase !== "open") {
		throw new AgentSessionRuntimeError(`Pi ${type} has no matching open content phase`);
	} else if (transition === "end") {
		ownedPhase.phase = "ended";
	}
	const delta = fields.get("delta");
	if (stringStartsWith(type, "text_")) {
		if (next.type !== "text") throw new AgentSessionRuntimeError(`Pi ${type} content type is invalid`);
		const before = prior?.type === "text" ? prior.text ?? "" : "";
		const after = next.text ?? "";
		if (type === "text_start" && after !== "") throw new AgentSessionRuntimeError("Pi text_start must begin empty");
		if (type === "text_delta" && after !== before + intrinsicString(delta)) {
			throw new AgentSessionRuntimeError("Pi text_delta is not the actual novel suffix");
		}
		if (type === "text_end" && (fields.get("content") !== after || !stringStartsWith(after, before))) {
			throw new AgentSessionRuntimeError("Pi text_end replaced or misreported content");
		}
		const novelBytes = byteLength(stringSlice(after, before.length));
		return envelopeGrowth + cumulativeIdentityGrowth(prior?.identity, next.identity, novelBytes);
	}
	if (stringStartsWith(type, "thinking_")) {
		if (next.type !== "thinking") throw new AgentSessionRuntimeError(`Pi ${type} content type is invalid`);
		const before = prior?.type === "thinking" ? prior.thinking ?? "" : "";
		const after = next.thinking ?? "";
		if (type === "thinking_start" && after !== "") throw new AgentSessionRuntimeError("Pi thinking_start must begin empty");
		if (type === "thinking_delta" && after !== before + intrinsicString(delta)) {
			throw new AgentSessionRuntimeError("Pi thinking_delta is not the actual novel suffix");
		}
		if (type === "thinking_end" && (fields.get("content") !== after || !stringStartsWith(after, before))) {
			throw new AgentSessionRuntimeError("Pi thinking_end replaced or misreported content");
		}
		const novelBytes = byteLength(stringSlice(after, before.length));
		return envelopeGrowth + cumulativeIdentityGrowth(prior?.identity, next.identity, novelBytes);
	}
	if (!stringStartsWith(type, "toolcall_") || next.type !== "toolCall") {
		throw new AgentSessionRuntimeError(`Pi ${type} content type is invalid`);
	}
	if (type === "toolcall_start") {
		if ((next.partialJson ?? "") !== "") throw new AgentSessionRuntimeError("Pi toolcall_start must begin empty");
		return envelopeGrowth + cumulativeIdentityGrowth(prior?.identity, next.identity, 0);
	}
	if (type === "toolcall_delta") {
		const before = prior?.type === "toolCall" ? prior.partialJson ?? "" : "";
		if (next.partialJson !== before + intrinsicString(delta)) {
			throw new AgentSessionRuntimeError("Pi toolcall_delta is not the actual novel suffix");
		}
		return envelopeGrowth + cumulativeIdentityGrowth(prior?.identity, next.identity, byteLength(intrinsicString(delta)));
	}
	const toolCall = snapshotToolCall(fields.get("toolCall"), MAX_EVENT_BYTES, capture.projectArguments);
	if (canonicalJson(toolCall) !== canonicalToolContent(next, false)) {
		throw new AgentSessionRuntimeError("Pi toolcall_end disagrees with cumulative content");
	}
	return envelopeGrowth + cumulativeIdentityGrowth(prior?.identity, next.identity, 0);
}

function streamSnapshotGrowth(previous: AssistantTerminal | undefined, current: AssistantTerminal): number {
	if (!previous) return byteLength(current.identity);
	if (current.content.length < previous.content.length) {
		throw new AgentSessionRuntimeError("Pi stream cumulative content shrank");
	}
	let growth = cumulativeIdentityGrowth(previous.envelopeIdentity, current.envelopeIdentity, 0);
	for (let index = 0; index < current.content.length; index += 1) {
		const before = previous.content[index];
		const after = current.content[index]!;
		growth += cumulativeIdentityGrowth(before?.identity, after.identity, 0);
	}
	return growth;
}

function cumulativeIdentityGrowth(previous: string | undefined, current: string, novelBytes: number): number {
	if (previous === undefined) return intrinsicMax(novelBytes, byteLength(current));
	if (previous === current) return novelBytes;
	const previousBytes = byteLength(previous);
	const currentBytes = byteLength(current);
	if (novelBytes > 0 && currentBytes > previousBytes) {
		return intrinsicMax(novelBytes, currentBytes - previousBytes);
	}
	// A changed identity without provable append-only growth is replacement state. Charge the
	// complete replacement so equal-size or shrinking metadata cannot evade the aggregate bound.
	return intrinsicMax(novelBytes, currentBytes);
}

function captureKnownRecordFields(
	value: unknown,
	description: string,
	knownFields: readonly string[],
): ReadonlyMap<string, unknown> {
	if (!value || typeof value !== "object" || INTRINSIC_ARRAY_IS_ARRAY(value) || INTRINSIC_IS_PROXY(value)) {
		throw new AgentSessionRuntimeError(`${description} must be a plain non-proxy record`);
	}
	assertApprovedRecordPrototype(value, description);
	const fields = new Map<string, unknown>();
	for (const key of knownFields) {
		const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, key);
		if (!descriptor) continue;
		if (!descriptor.enumerable) continue;
		if (descriptor.get || descriptor.set || !("value" in descriptor)) {
			throw new AgentSessionRuntimeError(`${description} contains an accessor field ${INTRINSIC_JSON_STRINGIFY(key)}`);
		}
		fields.set(key, descriptor.value);
	}
	return fields;
}

function assertExactCapturedFields(
	fields: ReadonlyMap<string, unknown>,
	expected: readonly string[],
	description: string,
): void {
	let exact = fields.size === expected.length;
	for (let index = 0; exact && index < expected.length; index += 1) exact = fields.has(expected[index]!);
	if (!exact) {
		throw new AgentSessionRuntimeError(`${description} must be an exact closed record`);
	}
}

function captureDenseArray(value: unknown, description: string): readonly unknown[] {
	if (!INTRINSIC_ARRAY_IS_ARRAY(value) || INTRINSIC_IS_PROXY(value)) {
		throw new AgentSessionRuntimeError(`${description} must be a bounded dense non-proxy array`);
	}
	assertApprovedArrayPrototype(value, description);
	const lengthDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, "length");
	const length = lengthDescriptor && "value" in lengthDescriptor &&
		typeof lengthDescriptor.value === "number" ? lengthDescriptor.value : -1;
	if (!INTRINSIC_NUMBER_IS_SAFE_INTEGER(length) || length < 0 || length > MAX_EVENT_ARRAY_ITEMS) {
		throw new AgentSessionRuntimeError(`${description} must be a bounded dense non-proxy array`);
	}
	const captured: unknown[] = [];
	for (let index = 0; index < length; index += 1) {
		const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, intrinsicString(index));
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
			throw new AgentSessionRuntimeError(`${description} contains an invalid array element`);
		}
		arrayPush(captured, descriptor.value);
	}
	return INTRINSIC_OBJECT_FREEZE(captured);
}

type JsonEventValue = null | boolean | number | string | JsonEventArray | JsonEventObject;
interface JsonEventArray extends ReadonlyArray<JsonEventValue> {}
interface JsonEventObject { readonly [key: string]: JsonEventValue }

interface JsonEventBudget {
	nodes: number;
	keys: number;
	bytes: number;
	maximum: number;
}

// Canonical JSON is deliberately finite. These are the only record fields consumed from Pi
// message/tool DTOs or runtime-owned identity records; caller-owned peers are never discovered by
// scanning a record. Keep this vocabulary schema-oriented and project arbitrary details elsewhere.
const JSON_EVENT_RECORD_FIELDS = INTRINSIC_OBJECT_FREEZE([
	"api", "argumentsIdentity", "cacheRead", "cacheWrite", "cacheWrite1h", "changed", "code", "content", "cost",
	"details", "diagnostics", "error", "errorMessage", "id", "input", "isError", "limit", "message", "model",
	"name", "newText", "offset", "oldText", "output", "partialJson", "path", "provider", "reasoning", "redacted",
	"references", "responseId", "responseModel", "resultIdentity", "role", "stack", "status", "stopReason", "summary",
	"target", "terminal", "terminate", "text", "textSignature", "thinking", "thinkingSignature", "thoughtSignature",
	"timestamp", "toolCallId", "toolName", "total", "totalTokens", "type", "usage",
] as const);

function snapshotEventJson(value: unknown, description: string, maximum: number): JsonEventValue {
	return snapshotEventJsonValue(
		value,
		0,
		{ nodes: 0, keys: 0, bytes: 0, maximum },
		new INTRINSIC_WEAK_SET<object>(),
		description,
	);
}

function projectToolArguments(
	projectArguments: ToolArgumentProjector,
	toolName: string,
	value: unknown,
	maximum: number,
): string {
	const projected = projectArguments(toolName, value, maximum);
	const identity = INTRINSIC_JSON_STRINGIFY(projected);
	if (typeof identity !== "string" || byteLength(identity) > maximum) {
		throw new AgentSessionRuntimeError(`${toolName} projected arguments exceeded their identity bound`);
	}
	return identity;
}

function snapshotToolCall(
	value: unknown,
	maximum: number,
	projectArguments: ToolArgumentProjector,
): JsonEventValue {
	const fields = captureKnownRecordFields(value, "Pi toolcall_end value", [
		"type", "id", "name", "arguments", "thoughtSignature",
	]);
	assertAllowedCapturedFields(
		fields,
		["type", "id", "name", "arguments", "thoughtSignature"],
		["type", "id", "name", "arguments"],
		"Pi toolcall_end value",
	);
	if (fields.get("type") !== "toolCall" || typeof fields.get("id") !== "string" ||
		typeof fields.get("name") !== "string") {
		throw new AgentSessionRuntimeError("Pi toolcall_end identity is invalid");
	}
	const thoughtSignature = optionalCapturedString(fields, "thoughtSignature", "Pi toolcall_end thought signature");
	return snapshotEventJson({
		type: "toolCall",
		id: fields.get("id") as string,
		name: fields.get("name") as string,
		argumentsIdentity: projectToolArguments(
			projectArguments,
			fields.get("name") as string,
			fields.get("arguments"),
			maximum,
		),
		...(thoughtSignature === undefined ? {} : { thoughtSignature }),
	}, "Pi toolcall_end projection", maximum);
}

function snapshotOpaqueValue(value: unknown, description: string, maximum: number): JsonEventValue {
	if (value === null || typeof value === "string" || typeof value === "boolean" || typeof value === "number") {
		return snapshotEventJson(value, description, maximum);
	}
	if (typeof value !== "object" || INTRINSIC_IS_PROXY(value)) {
		throw new AgentSessionRuntimeError(`${description} contains an unsupported opaque value`);
	}
	if (INTRINSIC_ARRAY_IS_ARRAY(value)) {
		assertApprovedArrayPrototype(value, `${description} opaque array`);
		const lengthDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, "length");
		const length = lengthDescriptor && "value" in lengthDescriptor ? lengthDescriptor.value : undefined;
		if (typeof length !== "number" || !INTRINSIC_NUMBER_IS_SAFE_INTEGER(length) || length < 0 || length > MAX_EVENT_ARRAY_ITEMS) {
			throw new AgentSessionRuntimeError(`${description} opaque array length is invalid`);
		}
		return snapshotEventJson(`[redacted opaque array:${length}]`, description, maximum);
	}
	assertApprovedRecordPrototype(value, `${description} opaque record`);
	return snapshotEventJson("[redacted opaque record]", description, maximum);
}

function snapshotToolResultContent(value: unknown, maximum: number): JsonEventValue {
	const content = captureDenseArray(value, "tool result content");
	const projected: JsonEventValue[] = [];
	for (const item of content) {
		const fields = captureKnownRecordFields(item, "tool result content item", ["type", "text"]);
		assertExactCapturedFields(fields, ["type", "text"], "tool result content item");
		if (fields.get("type") !== "text" || typeof fields.get("text") !== "string") {
			throw new AgentSessionRuntimeError("tool result content item is invalid");
		}
		arrayPush(projected, snapshotEventJson(
			{ type: "text", text: fields.get("text") as string },
			"tool result content item projection",
			maximum,
		));
	}
	return snapshotEventJson(projected, "tool result content projection", maximum);
}

function snapshotEventJsonValue(
	value: unknown,
	depth: number,
	budget: JsonEventBudget,
	ancestors: WeakSet<object>,
	description: string,
): JsonEventValue {
	budget.nodes += 1;
	if (budget.nodes > MAX_EVENT_NODES || depth > MAX_EVENT_DEPTH) {
		throw new AgentSessionRuntimeError(`${description} exceeded its node or depth bound`);
	}
	const add = (bytes: number): void => {
		budget.bytes += bytes;
		if (budget.bytes > budget.maximum) throw new AgentSessionRuntimeError(`${description} exceeded its byte bound`);
	};
	if (value === null) { add(4); return null; }
	if (typeof value === "string") { add(byteLength(value) + 2); return value; }
	if (typeof value === "boolean") { add(5); return value; }
	if (typeof value === "number") {
		if (!INTRINSIC_NUMBER_IS_FINITE(value)) throw new AgentSessionRuntimeError(`${description} contains a non-JSON number`);
		add(24);
		return value;
	}
	if (typeof value !== "object") throw new AgentSessionRuntimeError(`${description} contains a non-JSON value`);
	if (INTRINSIC_IS_PROXY(value)) throw new AgentSessionRuntimeError(`${description} contains a proxy`);
	if (INTRINSIC_REFLECT_APPLY(INTRINSIC_WEAK_SET_HAS, ancestors, [value])) {
		throw new AgentSessionRuntimeError(`${description} contains a cycle`);
	}
	INTRINSIC_REFLECT_APPLY(INTRINSIC_WEAK_SET_ADD, ancestors, [value]);
	add(2);
	try {
		if (INTRINSIC_ARRAY_IS_ARRAY(value)) {
			assertApprovedArrayPrototype(value, `${description} array`);
			const lengthDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, "length");
			const length = lengthDescriptor && "value" in lengthDescriptor &&
				typeof lengthDescriptor.value === "number" ? lengthDescriptor.value : -1;
			if (!INTRINSIC_NUMBER_IS_SAFE_INTEGER(length) || length < 0 || length > MAX_EVENT_ARRAY_ITEMS) {
				throw new AgentSessionRuntimeError(`${description} array is oversized`);
			}
			budget.keys += length;
			if (budget.keys > MAX_EVENT_NODES) throw new AgentSessionRuntimeError(`${description} exceeded its key bound`);
			const snapshot: JsonEventValue[] = [];
			for (let index = 0; index < length; index += 1) {
				const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, intrinsicString(index));
				if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
					throw new AgentSessionRuntimeError(`${description} array contains an invalid item`);
				}
				arrayPush(snapshot, snapshotEventJsonValue(descriptor.value, depth + 1, budget, ancestors, description));
			}
			return INTRINSIC_OBJECT_FREEZE(snapshot);
		}
		assertApprovedRecordPrototype(value, description);
		const snapshot = INTRINSIC_OBJECT_CREATE(null) as Record<string, JsonEventValue>;
		for (const key of JSON_EVENT_RECORD_FIELDS) {
			const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, key);
			if (!descriptor?.enumerable) continue;
			budget.keys += 1;
			if (budget.keys > MAX_EVENT_NODES) {
				throw new AgentSessionRuntimeError(`${description} exceeded its key bound`);
			}
			if (descriptor.get || descriptor.set || !("value" in descriptor)) {
				throw new AgentSessionRuntimeError(`${description} contains an invalid field`);
			}
			add(byteLength(key) + 3);
			INTRINSIC_OBJECT_DEFINE_PROPERTY(snapshot, key, {
				value: snapshotEventJsonValue(descriptor.value, depth + 1, budget, ancestors, description),
				enumerable: true,
				writable: false,
				configurable: false,
			});
		}
		return INTRINSIC_OBJECT_FREEZE(snapshot);
	} finally {
		INTRINSIC_REFLECT_APPLY(INTRINSIC_WEAK_SET_DELETE, ancestors, [value]);
	}
}

function canonicalJson(value: JsonEventValue): string {
	return INTRINSIC_JSON_STRINGIFY(value);
}

function canonicalToolContent(value: CapturedAssistantContent, _terminal: boolean): string {
	return value.terminalIdentity ?? value.identity;
}

function captureAssistantTerminal(
	value: unknown,
	streaming: boolean,
	maximum: number,
	projectArguments: ToolArgumentProjector,
): AssistantTerminal | undefined {
	const fields = captureKnownRecordFields(value, "AgentSession terminal message", [
		"role", "content", "api", "provider", "model", "responseModel", "responseId", "diagnostics", "usage",
		"stopReason", "errorMessage", "timestamp",
	]);
	const role = fields.get("role");
	if (role !== "assistant") return undefined;
	assertAllowedCapturedFields(fields, [
		"role", "content", "api", "provider", "model", "responseModel", "responseId", "diagnostics", "usage",
		"stopReason", "errorMessage", "timestamp",
	], ["role", "content", "api", "provider", "model", "usage", "stopReason", "timestamp"],
	"AgentSession assistant message");
	const api = fields.get("api");
	const provider = fields.get("provider");
	const model = fields.get("model");
	const stopReason = fields.get("stopReason");
	const timestamp = fields.get("timestamp");
	if (typeof api !== "string" || typeof provider !== "string" || typeof model !== "string" || typeof stopReason !== "string" ||
		typeof timestamp !== "number" || !INTRINSIC_NUMBER_IS_FINITE(timestamp)) {
		throw new AgentSessionRuntimeError("AgentSession assistant terminal contains invalid routing fields");
	}
	for (const optionalString of ["responseModel", "responseId", "errorMessage"] as const) {
		if (fields.has(optionalString) && typeof fields.get(optionalString) !== "string") {
			throw new AgentSessionRuntimeError(`AgentSession assistant ${optionalString} is invalid`);
		}
	}
	const usage = captureAssistantUsage(fields.get("usage"), maximum);
	const diagnostics = fields.has("diagnostics")
		? captureAssistantDiagnostics(fields.get("diagnostics"), maximum)
		: undefined;
	const contentSource = captureDenseArray(fields.get("content"), "AgentSession assistant terminal content");
	const content: Readonly<CapturedAssistantContent>[] = [];
	for (let contentIndex = 0; contentIndex < contentSource.length; contentIndex += 1) {
		const part = contentSource[contentIndex];
		const partFields = captureKnownRecordFields(part, "AgentSession assistant content part", [
			"type", "text", "textSignature", "thinking", "thinkingSignature", "redacted", "id", "name", "arguments",
			"thoughtSignature", "partialJson",
		]);
		const type = partFields.get("type");
		if (typeof type !== "string") {
			throw new AgentSessionRuntimeError("AgentSession assistant content part is invalid");
		}
		if (type === "text") {
			assertAllowedCapturedFields(partFields, ["type", "text", "textSignature"], ["type", "text"],
				"AgentSession assistant text content");
			const text = partFields.get("text");
			if (typeof text !== "string") throw new AgentSessionRuntimeError("AgentSession assistant text content is invalid");
			const textSignature = optionalCapturedString(partFields, "textSignature", "assistant text signature");
			const identity = canonicalJson(snapshotEventJson({ type, text, ...(textSignature === undefined ? {} : { textSignature }) },
				"AgentSession assistant text identity", maximum));
			arrayPush(content, INTRINSIC_OBJECT_FREEZE({ type, text, identity }));
			continue;
		}
		if (type === "thinking") {
			assertAllowedCapturedFields(partFields, ["type", "thinking", "thinkingSignature", "redacted"], ["type", "thinking"],
				"AgentSession assistant thinking content");
			const thinking = partFields.get("thinking");
			if (typeof thinking !== "string") {
				throw new AgentSessionRuntimeError("AgentSession assistant thinking content is invalid");
			}
			const thinkingSignature = optionalCapturedString(partFields, "thinkingSignature", "assistant thinking signature");
			const redacted = partFields.get("redacted");
			if (redacted !== undefined && typeof redacted !== "boolean") {
				throw new AgentSessionRuntimeError("AgentSession assistant thinking redacted flag is invalid");
			}
			const identity = canonicalJson(snapshotEventJson({
				type,
				thinking,
				...(thinkingSignature === undefined ? {} : { thinkingSignature }),
				...(redacted === undefined ? {} : { redacted }),
			}, "AgentSession assistant thinking identity", maximum));
			arrayPush(content, INTRINSIC_OBJECT_FREEZE({ type, thinking, identity }));
			continue;
		}
		if (type === "toolCall") {
			assertAllowedCapturedFields(partFields, ["type", "id", "name", "arguments", "thoughtSignature", ...(streaming ? ["partialJson"] : [])],
				["type", "id", "name", "arguments"], "AgentSession assistant tool-call content");
			const id = partFields.get("id");
			const name = partFields.get("name");
			if (typeof id !== "string" || typeof name !== "string") {
				throw new AgentSessionRuntimeError("AgentSession assistant tool-call identity is invalid");
			}
			const thoughtSignature = optionalCapturedString(partFields, "thoughtSignature", "assistant tool thought signature");
			const partialJson = optionalCapturedString(partFields, "partialJson", "assistant tool partial JSON");
			// Pi publishes `{}` while a tool call is still streaming. That placeholder has no
			// execution authority: project only the complete terminal arguments after partialJson
			// disappears. Missing or malformed terminal arguments still fail closed below.
			const argumentsIdentity = partialJson === undefined
				? projectToolArguments(projectArguments, name, partFields.get("arguments"), maximum)
				: undefined;
			const terminalIdentity = argumentsIdentity === undefined
				? undefined
				: canonicalJson(snapshotEventJson({
					type,
					id,
					name,
					argumentsIdentity,
					...(thoughtSignature === undefined ? {} : { thoughtSignature }),
				}, "AgentSession assistant tool identity", maximum));
			const identity = terminalIdentity ?? canonicalJson(snapshotEventJson({
				type,
				id,
				name,
				partialJson,
				...(thoughtSignature === undefined ? {} : { thoughtSignature }),
			}, "AgentSession assistant streaming tool identity", maximum));
			arrayPush(content, INTRINSIC_OBJECT_FREEZE({
				type, id, name, argumentsIdentity, partialJson, identity, terminalIdentity,
			}));
			continue;
		}
		throw new AgentSessionRuntimeError(`AgentSession assistant content type ${INTRINSIC_JSON_STRINGIFY(type)} is invalid`);
	}
	const envelope = {
		role: "assistant",
		api,
		provider,
		model,
		...(fields.has("responseModel") ? { responseModel: fields.get("responseModel") } : {}),
		...(fields.has("responseId") ? { responseId: fields.get("responseId") } : {}),
		...(diagnostics === undefined ? {} : { diagnostics }),
		usage,
		stopReason,
		...(fields.has("errorMessage") ? { errorMessage: fields.get("errorMessage") } : {}),
		timestamp,
	};
	const envelopeIdentity = canonicalJson(snapshotEventJson(
		envelope,
		"AgentSession assistant envelope identity",
		maximum,
	));
	const contentIdentities: string[] = [];
	for (let index = 0; index < content.length; index += 1) arrayPush(contentIdentities, content[index]!.identity);
	const identity = canonicalJson(snapshotEventJson({
		...envelope,
		content: contentIdentities,
	}, "AgentSession assistant identity", maximum));
	return INTRINSIC_OBJECT_FREEZE({
		role: "assistant",
		api,
		provider,
		model,
		stopReason,
		timestamp,
		content: INTRINSIC_OBJECT_FREEZE(content),
		envelopeIdentity,
		identity,
	});
}

function captureAssistantDiagnostics(value: unknown, maximum: number): JsonEventValue {
	const diagnostics = captureDenseArray(value, "AgentSession assistant diagnostics");
	const projected: JsonEventValue[] = [];
	for (const diagnostic of diagnostics) {
		const fields = captureKnownRecordFields(diagnostic, "AgentSession assistant diagnostic", [
			"type", "timestamp", "error", "details",
		]);
		assertAllowedCapturedFields(fields, ["type", "timestamp", "error", "details"], ["type", "timestamp"],
			"AgentSession assistant diagnostic");
		const type = fields.get("type");
		const timestamp = fields.get("timestamp");
		if (typeof type !== "string" || type.length === 0 || type.length > 256 ||
			typeof timestamp !== "number" || !INTRINSIC_NUMBER_IS_FINITE(timestamp)) {
			throw new AgentSessionRuntimeError("AgentSession assistant diagnostic identity is invalid");
		}
		const output: Record<string, JsonEventValue> = { type, timestamp };
		const errorValue = fields.get("error");
		if (errorValue !== undefined) {
			const errorFields = captureKnownRecordFields(errorValue, "AgentSession assistant diagnostic error", [
				"name", "message", "stack", "code",
			]);
			assertAllowedCapturedFields(errorFields, ["name", "message", "stack", "code"], ["message"],
				"AgentSession assistant diagnostic error");
			const message = errorFields.get("message");
			if (typeof message !== "string") {
				throw new AgentSessionRuntimeError("AgentSession assistant diagnostic error message is invalid");
			}
			const errorOutput: Record<string, JsonEventValue> = { message };
			for (const name of ["name", "stack"] as const) {
				const entry = errorFields.get(name);
				if (entry === undefined) continue;
				if (typeof entry !== "string") {
					throw new AgentSessionRuntimeError(`AgentSession assistant diagnostic error ${name} is invalid`);
				}
				errorOutput[name] = entry;
			}
			const code = errorFields.get("code");
			if (code !== undefined) {
				if (typeof code !== "string" && (typeof code !== "number" || !INTRINSIC_NUMBER_IS_FINITE(code))) {
					throw new AgentSessionRuntimeError("AgentSession assistant diagnostic error code is invalid");
				}
				errorOutput.code = code;
			}
			output.error = errorOutput;
		}
		const detailsValue = fields.get("details");
		if (detailsValue !== undefined) {
			const details = projectDiagnosticJson(
				detailsValue, 0, { nodes: 0, keys: 0 }, new INTRINSIC_WEAK_SET<object>(),
			);
			if (details === undefined || !details || INTRINSIC_ARRAY_IS_ARRAY(details) || typeof details !== "object") {
				throw new AgentSessionRuntimeError("AgentSession assistant diagnostic details are invalid");
			}
			output.details = details;
		}
		arrayPush(projected, snapshotEventJson(output, "AgentSession assistant diagnostic snapshot", maximum));
	}
	return snapshotEventJson(projected, "AgentSession assistant diagnostics snapshot", maximum);
}

function projectDiagnosticJson(
	value: unknown,
	depth: number,
	budget: { nodes: number; keys: number },
	ancestors: WeakSet<object>,
): JsonEventValue | undefined {
	budget.nodes += 1;
	if (budget.nodes > MAX_EVENT_NODES || depth > MAX_EVENT_DEPTH) {
		throw new AgentSessionRuntimeError("AgentSession assistant diagnostic exceeded its node or depth bound");
	}
	if (value === undefined) return undefined;
	if (value === null || typeof value === "string" || typeof value === "boolean") return value;
	if (typeof value === "number") {
		if (!INTRINSIC_NUMBER_IS_FINITE(value)) throw new AgentSessionRuntimeError("AgentSession assistant diagnostic contains a non-JSON number");
		return value;
	}
	if (typeof value !== "object" || INTRINSIC_IS_PROXY(value)) {
		throw new AgentSessionRuntimeError("AgentSession assistant diagnostic contains a non-JSON value");
	}
	if (INTRINSIC_REFLECT_APPLY(INTRINSIC_WEAK_SET_HAS, ancestors, [value])) {
		throw new AgentSessionRuntimeError("AgentSession assistant diagnostic contains a cycle");
	}
	INTRINSIC_REFLECT_APPLY(INTRINSIC_WEAK_SET_ADD, ancestors, [value]);
	try {
		if (INTRINSIC_ARRAY_IS_ARRAY(value)) {
			const items = captureDenseArray(value, "AgentSession assistant diagnostic array");
			budget.keys += items.length;
			if (budget.keys > MAX_EVENT_NODES) throw new AgentSessionRuntimeError("AgentSession assistant diagnostic is too wide");
			const output: JsonEventValue[] = [];
			for (const item of items) {
				const projected = projectDiagnosticJson(item, depth + 1, budget, ancestors);
				if (projected === undefined) {
					throw new AgentSessionRuntimeError("AgentSession assistant diagnostic array contains undefined");
				}
				arrayPush(output, projected);
			}
			return INTRINSIC_OBJECT_FREEZE(output);
		}
		assertApprovedRecordPrototype(value, "AgentSession assistant diagnostic record");
		budget.keys += 1;
		if (budget.keys > MAX_EVENT_NODES) throw new AgentSessionRuntimeError("AgentSession assistant diagnostic is too wide");
		const output = INTRINSIC_OBJECT_CREATE(null) as Record<string, JsonEventValue>;
		INTRINSIC_OBJECT_DEFINE_PROPERTY(output, "summary", {
			value: "[redacted diagnostic details]",
			enumerable: true,
			writable: false,
			configurable: false,
		});
		return INTRINSIC_OBJECT_FREEZE(output);
	} finally {
		INTRINSIC_REFLECT_APPLY(INTRINSIC_WEAK_SET_DELETE, ancestors, [value]);
	}
}

function optionalCapturedString(
	fields: ReadonlyMap<string, unknown>,
	name: string,
	description: string,
): string | undefined {
	if (!fields.has(name)) return undefined;
	const value = fields.get(name);
	if (typeof value !== "string") throw new AgentSessionRuntimeError(`${description} is invalid`);
	return value;
}

function captureAssistantUsage(value: unknown, maximum: number): JsonEventValue {
	const fields = captureKnownRecordFields(value, "AgentSession assistant usage", [
		"input", "output", "cacheRead", "cacheWrite", "cacheWrite1h", "reasoning", "totalTokens", "cost",
	]);
	assertAllowedCapturedFields(fields, [
		"input", "output", "cacheRead", "cacheWrite", "cacheWrite1h", "reasoning", "totalTokens", "cost",
	], ["input", "output", "cacheRead", "cacheWrite", "totalTokens", "cost"], "AgentSession assistant usage");
	for (const name of ["input", "output", "cacheRead", "cacheWrite", "cacheWrite1h", "reasoning", "totalTokens"] as const) {
		if (!fields.has(name)) continue;
		const entry = fields.get(name);
		if (typeof entry !== "number" || !INTRINSIC_NUMBER_IS_FINITE(entry) || entry < 0) {
			throw new AgentSessionRuntimeError(`AgentSession assistant usage ${name} is invalid`);
		}
	}
	const cost = captureKnownRecordFields(fields.get("cost"), "AgentSession assistant usage cost", [
		"input", "output", "cacheRead", "cacheWrite", "total",
	]);
	assertAllowedCapturedFields(cost, ["input", "output", "cacheRead", "cacheWrite", "total"],
		["input", "output", "cacheRead", "cacheWrite", "total"], "AgentSession assistant usage cost");
	for (const name of ["input", "output", "cacheRead", "cacheWrite", "total"] as const) {
		const entry = cost.get(name);
		if (typeof entry !== "number" || !INTRINSIC_NUMBER_IS_FINITE(entry) || entry < 0) {
			throw new AgentSessionRuntimeError(`AgentSession assistant usage cost ${name} is invalid`);
		}
	}
	return snapshotEventJson(value, "AgentSession assistant usage snapshot", maximum);
}

function assertAllowedCapturedFields(
	fields: ReadonlyMap<string, unknown>,
	allowed: readonly string[],
	required: readonly string[],
	description: string,
): void {
	const allowedSet = new Set(allowed);
	for (const key of fields.keys()) {
		if (!allowedSet.has(key)) {
			throw new AgentSessionRuntimeError(`${description} contains unknown field ${INTRINSIC_JSON_STRINGIFY(key)}`);
		}
	}
	let missing = false;
	for (let index = 0; !missing && index < required.length; index += 1) missing = !fields.has(required[index]!);
	if (missing) {
		throw new AgentSessionRuntimeError(`${description} is missing a required field`);
	}
}

function verifyTerminalCapture(capture: TerminalCapture): AssistantTerminal {
	if (capture.piPhase !== "settled" || !capture.piSettled || !capture.frozen ||
		capture.piOpenMessageRole !== undefined || capture.agentEndCount !== 1 ||
		capture.agentEndWillRetry || !capture.messageEnd || !capture.agentEnd ||
		!sameTerminal(capture.messageEnd, capture.agentEnd)) {
		throw new AgentSessionRuntimeError("AgentSession returned an invalid settled lifecycle");
	}
	if (capture.agentEnd.stopReason !== "stop") {
		throw new AgentSessionRuntimeError(`AgentSession terminal stop reason ${capture.agentEnd.stopReason} is not accepted`);
	}
	return capture.agentEnd;
}

function sameTerminal(left: AssistantTerminal, right: AssistantTerminal): boolean {
	return left.identity === right.identity;
}

function assistantText(terminal: AssistantTerminal): string {
	const text: string[] = [];
	for (let index = 0; index < terminal.content.length; index += 1) {
		const part = terminal.content[index]!;
		if (part.type === "text" && typeof part.text === "string") arrayPush(text, part.text);
	}
	return stringTrim(arrayJoin(text, ""));
}

function parseHandoff(text: string, request: RoleRunRequest, maxBytes: number): AgentSessionHandoff {
	if (!text || byteLength(text) > maxBytes) throw new AgentSessionRuntimeError("AgentSession assistant output is empty or exceeds its bound");
	let candidate: unknown;
	try {
		candidate = INTRINSIC_REFLECT_APPLY(INTRINSIC_JSON_PARSE, INTRINSIC_JSON, [text]) as unknown;
	} catch (error) {
		throw new AgentSessionRuntimeError("AgentSession handoff must be exactly one JSON object", { cause: error });
	}
	if (!isRecord(candidate)) throw new AgentSessionRuntimeError("AgentSession handoff must be an object");
	const handoffFields = [
		"schemaVersion", "runId", "generation", "laneId", "candidateHead", "validationNonce", "role", "status",
		"summary", "observedMutation", "changedPaths", "verification", "findings",
	] as const;
	const fields = captureKnownRecordFields(candidate, "handoff", [
		...handoffFields,
		// Known authority-injection aliases are denied; arbitrary peers remain inert and unread.
		"authority", "unknownField",
	]);
	assertAllowedCapturedFields(fields, handoffFields, handoffFields, "handoff");
	if (fields.get("schemaVersion") !== 1) throw new AgentSessionRuntimeError("handoff schemaVersion is invalid");
	for (const [name, actual, expected] of [
		["runId", fields.get("runId"), request.binding.runId],
		["generation", fields.get("generation"), request.binding.generation],
		["laneId", fields.get("laneId"), request.binding.laneId],
		["candidateHead", fields.get("candidateHead"), request.binding.candidateHead],
		["validationNonce", fields.get("validationNonce"), request.binding.validationNonce],
		["role", fields.get("role"), request.role],
	] as const) {
		if (actual !== expected) throw new AgentSessionRuntimeError(`handoff ${name} binding mismatch`);
	}
	const handoffStatus = fields.get("status");
	if (!arrayIncludes(["completed", "blocked", "failed"] as const, intrinsicString(handoffStatus))) {
		throw new AgentSessionRuntimeError("handoff status is invalid");
	}
	const summary = redactedBoundedString(fields.get("summary"), "handoff summary", MAX_HANDOFF_SUMMARY_CHARACTERS, false);
	const observedMutation = fields.get("observedMutation");
	if (typeof observedMutation !== "boolean") throw new AgentSessionRuntimeError("handoff observedMutation is invalid");
	if (request.authority.readOnly && observedMutation) {
		throw new AgentSessionRuntimeError("read-only handoff reported a mutation");
	}
	const changedPathValues = captureHandoffArray(fields.get("changedPaths"), "handoff changedPaths");
	const changedPaths: string[] = [];
	for (let index = 0; index < changedPathValues.length; index += 1) {
		const path = changedPathValues[index];
		if (typeof path !== "string") throw new AgentSessionRuntimeError("handoff changed path is invalid");
		changedPaths[index] = validateScopedPath(path, request.authority.writePrefixes);
	}
	INTRINSIC_OBJECT_FREEZE(changedPaths);
	if (request.authority.readOnly && changedPaths.length > 0) throw new AgentSessionRuntimeError("read-only handoff contains changed paths");
	const verificationValues = captureHandoffArray(fields.get("verification"), "handoff verification");
	const verification: HandoffVerification[] = [];
	for (let index = 0; index < verificationValues.length; index += 1) {
		const entry = verificationValues[index];
		if (!isRecord(entry)) throw new AgentSessionRuntimeError("handoff verification entry is invalid");
		const verificationFields = captureKnownRecordFields(entry, "handoff verification", ["name", "status", "summary"]);
		const status = intrinsicString(verificationFields.get("status"));
		if (!arrayIncludes(["passed", "failed", "blocked", "not_run"] as const, status)) {
			throw new AgentSessionRuntimeError("handoff verification status is invalid");
		}
		verification[index] = {
			name: redactedBoundedString(verificationFields.get("name"), "handoff verification name", 128, false),
			status: status as HandoffVerification["status"],
			summary: redactedBoundedString(verificationFields.get("summary"), "handoff verification summary", MAX_HANDOFF_ITEM_CHARACTERS, false),
		};
	}
	INTRINSIC_OBJECT_FREEZE(verification);
	const findingValues = captureHandoffArray(fields.get("findings"), "handoff findings");
	const findings: string[] = [];
	for (let index = 0; index < findingValues.length; index += 1) {
		findings[index] = redactedBoundedString(
			findingValues[index], "handoff finding", MAX_HANDOFF_ITEM_CHARACTERS, false,
		);
	}
	INTRINSIC_OBJECT_FREEZE(findings);
	return {
		schemaVersion: 1,
		runId: request.binding.runId,
		generation: request.binding.generation,
		laneId: request.binding.laneId,
		candidateHead: request.binding.candidateHead,
		validationNonce: request.binding.validationNonce,
		role: request.role,
		status: handoffStatus as AgentSessionHandoff["status"],
		summary,
		observedMutation,
		changedPaths,
		verification,
		findings,
	};
}

function captureHandoffArray(value: unknown, description: string): readonly unknown[] {
	if (!INTRINSIC_ARRAY_IS_ARRAY(value) || INTRINSIC_IS_PROXY(value) ||
		INTRINSIC_GET_PROTOTYPE_OF(value) !== INTRINSIC_ARRAY_PROTOTYPE) {
		throw new AgentSessionRuntimeError(`${description} must be an exact bounded dense non-proxy array`);
	}
	const lengthDescriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, "length");
	const length = lengthDescriptor && "value" in lengthDescriptor ? lengthDescriptor.value : undefined;
	if (!lengthDescriptor || lengthDescriptor.get || lengthDescriptor.set ||
		typeof length !== "number" || !INTRINSIC_NUMBER_IS_SAFE_INTEGER(length) ||
		length < 0 || length > MAX_HANDOFF_ARRAY_ITEMS) {
		throw new AgentSessionRuntimeError(`${description} must be an exact bounded dense non-proxy array`);
	}
	const snapshot: unknown[] = [];
	for (let index = 0; index < length; index += 1) {
		const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, `${index}`);
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
			throw new AgentSessionRuntimeError(`${description} contains a sparse or accessor element`);
		}
		snapshot[index] = descriptor.value;
	}
	return INTRINSIC_OBJECT_FREEZE(snapshot);
}

function redactedBoundedString(value: unknown, description: string, max: number, allowEmpty: boolean): string {
	if (typeof value !== "string" || (!allowEmpty && value.length < 1) || value.length > max) {
		throw new AgentSessionRuntimeError(`${description} must be ${allowEmpty ? "a" : "a non-empty"} bounded string`);
	}
	const terminalControls = /[\u0000-\u001f\u007f-\u009f\u061c\u200e\u200f\u2028-\u202e\u2066-\u2069]/;
	if (terminalControls.test(value)) {
		throw new AgentSessionRuntimeError(`${description} contains a terminal control character`);
	}
	return redactSensitiveText(value);
}

function computeDeadline(timeoutMs: number, deadlineAt: number | undefined): number {
	const timeoutDeadline = Date.now() + timeoutMs;
	return deadlineAt === undefined ? timeoutDeadline : intrinsicMin(timeoutDeadline, deadlineAt);
}

function isAbsoluteNonTraversingPath(value: unknown): value is string {
	if (typeof value !== "string" || value.length < 1 || value.length > 4_096 || /[\u0000-\u001f\u007f]/.test(value)) return false;
	const flavor = win32.isAbsolute(value) ? win32 : posix;
	if (!flavor.isAbsolute(value)) return false;
	const segments = flavor === win32 ? stringSplit(value, /[\\/]+/) : stringSplit(value, "/");
	return !arrayIncludes(segments, "..");
}

function canonicalWorkspacePath(value: string): string {
	return /^(?:[A-Za-z]:[\\/]|\\\\)/.test(value) ? win32.normalize(value) : posix.normalize(value);
}

function validIdentifier(value: unknown): value is string {
	return typeof value === "string" && /^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$/.test(value);
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null && !INTRINSIC_ARRAY_IS_ARRAY(value);
}

function boundedPositiveInteger(value: number, description: string, maximum: number): number {
	if (!INTRINSIC_NUMBER_IS_SAFE_INTEGER(value) || value <= 0 || value > maximum) {
		throw new AgentSessionRuntimeError(
			`${description} must be a positive safe integer within the embedded maximum ${maximum}`,
		);
	}
	return value;
}

function frozenArray<T>(values: T[]): T[] {
	INTRINSIC_OBJECT_FREEZE(values);
	return values;
}

function byteLength(value: string): number {
	let bytes = 0;
	for (let index = 0; index < value.length; index += 1) {
		const code = INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_CHAR_CODE_AT, value, [index]) as number;
		if (code <= 0x7f) bytes += 1;
		else if (code <= 0x7ff) bytes += 2;
		else if (code >= 0xd800 && code <= 0xdbff && index + 1 < value.length) {
			const second = INTRINSIC_REFLECT_APPLY(INTRINSIC_STRING_CHAR_CODE_AT, value, [index + 1]) as number;
			if (second >= 0xdc00 && second <= 0xdfff) {
				bytes += 4;
				index += 1;
			} else bytes += 3;
		} else bytes += 3;
	}
	return bytes;
}

function deferred(): { promise: Promise<void>; resolve(): void } {
	let resolvePromise: (() => void) | undefined;
	const promise = new Promise<void>((resolve) => { resolvePromise = resolve; });
	return { promise, resolve: () => resolvePromise?.() };
}

async function bounded<T>(
	operation: Promise<T>,
	timeoutMs: number,
	description: string,
	unref = false,
): Promise<T> {
	let timer: ReturnType<typeof setTimeout> | undefined;
	const timeout = new Promise<never>((_resolve, reject) => {
		timer = setTimeout(() => reject(new AgentSessionRuntimeError(`${description} timed out after ${timeoutMs}ms`)), timeoutMs);
		if (unref) unrefTimeout(timer);
	});
	try {
		return await Promise.race([operation, timeout]);
	} finally {
		if (timer) clearTimeout(timer);
	}
}

async function boundedUntil<T>(
	operation: Promise<T>,
	deadlineAt: number,
	description: string,
	unref: boolean,
): Promise<T> {
	let timer: ReturnType<typeof setTimeout> | undefined;
	const timeoutMs = intrinsicMax(0, deadlineAt - Date.now());
	const timeout = new Promise<never>((_resolve, reject) => {
		timer = setTimeout(() => reject(new AgentSessionRuntimeError(`${description} timed out after ${timeoutMs}ms`)), timeoutMs);
		if (unref) unrefTimeout(timer);
	});
	try {
		return await Promise.race([operation, timeout]);
	} finally {
		if (timer) clearTimeout(timer);
	}
}

function unrefTimeout(timer: ReturnType<typeof setTimeout>): void {
	const candidate = timer as ReturnType<typeof setTimeout> & { unref?: () => void };
	candidate.unref?.();
}

type BoundedSettlement<T> =
	| { status: "fulfilled"; value: T }
	| { status: "rejected"; reason: unknown }
	| { status: "pending" };

async function settleWithin<T>(operation: Promise<T>, timeoutMs: number): Promise<BoundedSettlement<T>> {
	let timer: ReturnType<typeof setTimeout> | undefined;
	const timeout = new Promise<BoundedSettlement<T>>((resolve) => {
		timer = setTimeout(() => resolve({ status: "pending" }), timeoutMs);
	});
	try {
		return await Promise.race([
			operation.then<BoundedSettlement<T>, BoundedSettlement<T>>(
				(value) => ({ status: "fulfilled", value }),
				(reason) => ({ status: "rejected", reason }),
			),
			timeout,
		]);
	} finally {
		if (timer) clearTimeout(timer);
	}
}

function normalizeRuntimeError(error: unknown): AgentSessionRuntimeError {
	try {
		if ((typeof error === "object" && error !== null) || typeof error === "function") {
			if (INTRINSIC_IS_PROXY(error)) {
				return new AgentSessionRuntimeError("AgentSession run failed", {
					cause: createErrorSnapshot("proxied failure was omitted"),
				});
			}
		}
		if (hasRuntimeErrorPrototype(error, AgentSessionRuntimeError.prototype)) {
			// Runtime-authored typed messages are finite literals plus already-validated identifiers.
			// Preserve their operational category (extension/tool/model/cleanup) while still removing
			// terminal controls. Untrusted SDK errors take the redacting path below.
			const message = sanitizeRuntimeMessage(readErrorMessage(error), "AgentSession operation failed");
			const causeDescriptor = readOwnDataDescriptor(error as object, "cause");
			const cause = causeDescriptor.status === "data"
				? snapshotRuntimeFailure(causeDescriptor.value)
				: causeDescriptor.status === "invalid"
					? createErrorSnapshot("failure cause was unavailable")
					: undefined;
			return new AgentSessionRuntimeError(message, { cause });
		}
		const sourceMessage = readErrorMessage(error);
		const safeMessage = sourceMessage ? safeRuntimeMessage(sourceMessage, "") : "";
		return new AgentSessionRuntimeError(
			`AgentSession run failed${safeMessage ? `: ${safeMessage}` : ""}`,
			{ cause: snapshotRuntimeFailure(error) },
		);
	} catch {
		return new AgentSessionRuntimeError("AgentSession operation failed", {
			cause: createErrorSnapshot("failure normalization was unavailable"),
		});
	}
}

type OwnDataDescriptorResult =
	| { readonly status: "absent" }
	| { readonly status: "invalid" }
	| { readonly status: "data"; readonly value: unknown };

function readOwnDataDescriptor(value: object, field: PropertyKey): OwnDataDescriptorResult {
	const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(value, field);
	if (!descriptor) return { status: "absent" };
	if (descriptor.get || descriptor.set || !("value" in descriptor)) return { status: "invalid" };
	return { status: "data", value: descriptor.value };
}

function hasRuntimeErrorPrototype(value: unknown, target: object): boolean {
	if ((typeof value !== "object" || value === null) && typeof value !== "function") return false;
	try {
		if (INTRINSIC_IS_PROXY(value)) return false;
		let prototype = INTRINSIC_GET_PROTOTYPE_OF(value);
		for (let depth = 0; prototype !== null && depth < 8; depth += 1) {
			if (prototype === target) return true;
			prototype = INTRINSIC_GET_PROTOTYPE_OF(prototype);
		}
	} catch {
		return false;
	}
	return false;
}

function isNativeRuntimeError(value: unknown): value is object {
	if ((typeof value !== "object" || value === null) && typeof value !== "function") return false;
	try {
		return !INTRINSIC_IS_PROXY(value) && (
			Boolean(INTRINSIC_REFLECT_APPLY(INTRINSIC_IS_NATIVE_ERROR, undefined, [value])) ||
			hasRuntimeErrorPrototype(value, INTRINSIC_ERROR_PROTOTYPE)
		);
	} catch {
		return false;
	}
}

function defineErrorData(value: object, field: string, data: unknown): void {
	INTRINSIC_OBJECT_DEFINE_PROPERTY(value, field, {
		value: data,
		enumerable: false,
		writable: false,
		configurable: false,
	});
}

function createErrorSnapshot(
	message: string,
	name = "Error",
	causePresent = false,
	cause?: unknown,
): Error {
	const snapshot = INTRINSIC_OBJECT_CREATE(INTRINSIC_ERROR_PROTOTYPE) as Error;
	defineErrorData(snapshot, "message", message);
	defineErrorData(snapshot, "name", name);
	defineErrorData(snapshot, "stack", `${name}: ${message}`);
	if (causePresent) defineErrorData(snapshot, "cause", cause);
	return INTRINSIC_OBJECT_FREEZE(snapshot);
}

function createAggregateErrorSnapshot(
	errors: readonly unknown[],
	message: string,
	causePresent = false,
	cause?: unknown,
): AggregateError {
	const members: unknown[] = [];
	const length = intrinsicMin(errors.length, 16);
	for (let index = 0; index < length; index += 1) members[index] = errors[index];
	INTRINSIC_OBJECT_FREEZE(members);
	const snapshot = INTRINSIC_OBJECT_CREATE(INTRINSIC_AGGREGATE_ERROR_PROTOTYPE) as AggregateError;
	defineErrorData(snapshot, "message", message);
	defineErrorData(snapshot, "name", "AggregateError");
	defineErrorData(snapshot, "stack", `AggregateError: ${message}`);
	defineErrorData(snapshot, "errors", members);
	if (causePresent) defineErrorData(snapshot, "cause", cause);
	return INTRINSIC_OBJECT_FREEZE(snapshot);
}

function snapshotRuntimeFailure(
	error: unknown,
	depth = 0,
	seen: WeakSet<object> = new INTRINSIC_WEAK_SET<object>(),
): unknown {
	if (error === undefined || error === null || typeof error === "boolean" || typeof error === "number") return error;
	if (typeof error === "string") return safeRuntimeMessage(error, "failure");
	if (typeof error !== "object" && typeof error !== "function") return "unsupported failure";
	const object = error as object;
	try {
		if (INTRINSIC_IS_PROXY(object)) return createErrorSnapshot("proxied failure omitted");
	} catch {
		return createErrorSnapshot("failure shape unavailable");
	}
	if (INTRINSIC_REFLECT_APPLY(INTRINSIC_WEAK_SET_HAS, seen, [object])) {
		return createErrorSnapshot("cyclic failure omitted");
	}
	if (depth >= 4) return createErrorSnapshot("nested failure omitted");
	INTRINSIC_REFLECT_APPLY(INTRINSIC_WEAK_SET_ADD, seen, [object]);
	try {
		if (!isNativeRuntimeError(error)) return createErrorSnapshot("non-Error failure object");
		const causeDescriptor = readOwnDataDescriptor(object, "cause");
		const causePresent = causeDescriptor.status !== "absent";
		const cause = causeDescriptor.status === "data"
			? snapshotRuntimeFailure(causeDescriptor.value, depth + 1, seen)
			: causeDescriptor.status === "invalid"
				? createErrorSnapshot("failure cause was unavailable")
				: undefined;
		if (hasRuntimeErrorPrototype(error, INTRINSIC_AGGREGATE_ERROR_PROTOTYPE)) {
			const nested = snapshotAggregateMembers(object, depth, seen);
			return createAggregateErrorSnapshot(
				nested,
				safeRuntimeMessage(readErrorMessage(error), "multiple failures"),
				causePresent,
				cause,
			);
		}
		return createErrorSnapshot(
			safeRuntimeMessage(readErrorMessage(error), "external failure"),
			"Error",
			causePresent,
			cause,
		);
	} finally {
		INTRINSIC_REFLECT_APPLY(INTRINSIC_WEAK_SET_DELETE, seen, [object]);
	}
}

function snapshotAggregateMembers(error: object, depth: number, seen: WeakSet<object>): unknown[] {
	const unavailable = (): unknown[] => [createErrorSnapshot("aggregate members were unavailable")];
	const errorsDescriptor = readOwnDataDescriptor(error, "errors");
	if (errorsDescriptor.status !== "data" || !INTRINSIC_ARRAY_IS_ARRAY(errorsDescriptor.value)) return unavailable();
	const source = errorsDescriptor.value;
	if (INTRINSIC_IS_PROXY(source) || INTRINSIC_GET_PROTOTYPE_OF(source) !== INTRINSIC_ARRAY_PROTOTYPE) {
		return unavailable();
	}
	const lengthDescriptor = readOwnDataDescriptor(source, "length");
	if (lengthDescriptor.status !== "data" || typeof lengthDescriptor.value !== "number" ||
		!INTRINSIC_NUMBER_IS_SAFE_INTEGER(lengthDescriptor.value) || lengthDescriptor.value < 0 ||
		lengthDescriptor.value > 16) return unavailable();
	const nested: unknown[] = [];
	const length = lengthDescriptor.value;
	for (let index = 0; index < length; index += 1) {
		const descriptor = INTRINSIC_GET_OWN_PROPERTY_DESCRIPTOR(source, intrinsicString(index));
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) return unavailable();
		nested[index] = snapshotRuntimeFailure(descriptor.value, depth + 1, seen);
	}
	return nested;
}

function readErrorMessage(error: unknown): string {
	try {
		if (!isNativeRuntimeError(error)) return "";
		const descriptor = readOwnDataDescriptor(error, "message");
		return descriptor.status === "data" && typeof descriptor.value === "string" ? descriptor.value : "";
	} catch {
		return "";
	}
}

function safeRuntimeMessage(value: string, fallback: string): string {
	const source = value.length > 0 ? stringSlice(value, 0, 4_096) : fallback;
	return stringSlice(stringReplace(
		redactSensitiveText(source),
		/[\u0000-\u001f\u007f-\u009f\u061c\u200e\u200f\u2028-\u202e\u2066-\u2069]/g,
		" ",
	), 0, 2_048) || fallback;
}

function sanitizeRuntimeMessage(value: string, fallback: string): string {
	const source = value.length > 0 ? stringSlice(value, 0, 4_096) : fallback;
	return stringSlice(stringReplace(
		source,
		/[\u0000-\u001f\u007f-\u009f\u061c\u200e\u200f\u2028-\u202e\u2066-\u2069]/g,
		" ",
	), 0, 2_048) || fallback;
}

function combineFailures(failures: readonly unknown[], message: string): unknown {
	if (failures.length === 1) return snapshotRuntimeFailure(failures[0]);
	const nested: unknown[] = [];
	const length = intrinsicMin(failures.length, 16);
	for (let index = 0; index < length; index += 1) nested[index] = snapshotRuntimeFailure(failures[index]);
	return createAggregateErrorSnapshot(nested, message);
}
