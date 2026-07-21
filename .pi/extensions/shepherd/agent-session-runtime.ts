import { posix, win32 } from "node:path";
import { types as nodeTypes } from "node:util";

import type {
	AgentSessionEvent,
	CreateAgentSessionOptions,
} from "@earendil-works/pi-coding-agent";

import {
	buildRolePrompts,
	routeForRole,
	type PromptBinding,
	type ShepherdAgentRole,
	type ShepherdAgentThinking,
} from "./role-prompts.ts";
import {
	createToolPolicy,
	normalizeScopedPrefixes,
	redactSensitiveText,
	validateScopedPath,
	type HostCapability,
	type ScopedWorkspace,
	type ToolPolicy,
	type ToolAuthority,
} from "./tool-policy.ts";

const REQUIRED_PI_VERSION = "0.80.6";
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
const MAX_EVENT_RECORD_KEYS = 256;
const MAX_EVENT_ARRAY_ITEMS = 4_096;
const MAX_HANDOFF_SUMMARY_CHARACTERS = 4 * 1024;
const MAX_HANDOFF_ARRAY_ITEMS = 32;
const MAX_HANDOFF_ITEM_CHARACTERS = 2 * 1024;
const NATIVE_ABORTED_GETTER = Object.getOwnPropertyDescriptor(AbortSignal.prototype, "aborted")?.get;

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

export interface RuntimeAgentSession {
	model: RuntimeSessionModel | undefined;
	thinkingLevel: ShepherdAgentThinking | string;
	sessionFile: string | undefined;
	getActiveToolNames(): string[];
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

/** Injected adapter over the public Pi 0.80.6 createAgentSession API and in-memory services. */
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

export class AgentSessionRuntimeError extends Error {
	constructor(message: string, options?: ErrorOptions) {
		// Passing an options object even for a literal-undefined reason gives every public
		// failure a stable own `cause` field instead of erasing reasonless adapter failures.
		super(message, { cause: options?.cause });
		this.name = "AgentSessionRuntimeError";
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
		const remaining = Math.max(0, deadlineAt - Date.now());
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
			failure: new AgentSessionRuntimeError(`Pi AgentSession operation ${String(name)} is missing or invalid`),
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
	#abortPromise: Promise<void> | undefined;
	#disposePromise: Promise<void> | undefined;
	#unsubscribe: (() => void | PromiseLike<void>) | undefined;
	#unsubscribePromise: Promise<void> | undefined;
	#waitPromise: Promise<void> | undefined;

	constructor(session: RuntimeAgentSession) {
		this.#session = session;
		// Capture every operation independently. A hostile getter for an operational method
		// must never prevent us from acquiring a later dispose operation.
		this.#abort = captureSessionOperation(session, "abort");
		this.#waitForIdle = captureSessionOperation(session, "waitForIdle");
		this.#dispose = captureSessionOperation(session, "dispose");
		this.#prompt = captureSessionOperation(session, "prompt");
		this.#subscribe = captureSessionOperation(session, "subscribe");
		this.#getActiveToolNames = captureSessionOperation(session, "getActiveToolNames");
	}

	validationFailures(): readonly unknown[] {
		return [this.#abort, this.#waitForIdle, this.#dispose, this.#prompt, this.#subscribe, this.#getActiveToolNames]
			.filter((captured) => captured.failurePresent)
			.map((captured) => captured.failure);
	}

	activeToolNames(): unknown {
		if (!this.#getActiveToolNames.available) throw this.#getActiveToolNames.failure;
		return Reflect.apply(this.#getActiveToolNames.operation!, this.#session, []);
	}

	prompt(value: string, options: { expandPromptTemplates: false; source: "extension" }): Promise<void> {
		if (!this.#prompt.available) return Promise.reject(this.#prompt.failure);
		return Promise.resolve(Reflect.apply(this.#prompt.operation!, this.#session, [value, options])).then(() => undefined);
	}

	subscribe(listener: (event: AgentSessionEvent) => void): void {
		if (this.#unsubscribe !== undefined || this.#unsubscribePromise !== undefined) {
			throw new AgentSessionRuntimeError("AgentSession subscription ownership was already acquired");
		}
		if (!this.#subscribe.available) throw this.#subscribe.failure;
		const unsubscribe = Reflect.apply(this.#subscribe.operation!, this.#session, [listener]);
		if (typeof unsubscribe !== "function") {
			throw new AgentSessionRuntimeError("AgentSession subscribe returned an invalid cleanup operation");
		}
		this.#unsubscribe = unsubscribe as () => void | PromiseLike<void>;
	}

	abortOnce(): Promise<void> {
		if (!this.#abortPromise) {
			this.#abortPromise = this.#abort.available
				? Promise.resolve().then(() => Reflect.apply(this.#abort.operation!, this.#session, [])).then(() => undefined)
				: Promise.resolve();
			this.#abortPromise.catch(() => undefined);
		}
		return this.#abortPromise;
	}

	waitOnce(): Promise<void> {
		if (!this.#waitPromise) {
			this.#waitPromise = this.#waitForIdle.available
				? Promise.resolve().then(() => Reflect.apply(this.#waitForIdle.operation!, this.#session, [])).then(() => undefined)
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
					Promise.resolve(Reflect.apply(this.#dispose.operation!, this.#session, []))).then(() => undefined)
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
			failures.push(error);
		}
	}
	try {
		await bounded(owned.waitOnce(), timeoutMs, "session idle wait");
	} catch (error) {
		failures.push(error);
	}
	try {
		await bounded(owned.unsubscribeOnce(), timeoutMs, "session unsubscribe");
	} catch (error) {
		failures.push(error);
	}
	// Disposal owns a separate bound and remains reachable after every earlier phase.
	try {
		await bounded(owned.disposeOnce(), timeoutMs, "session dispose");
	} catch (error) {
		failures.push(error);
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
			if (settlement.status === "rejected") failures.push(settlement.reason);
		}
		try {
			await bounded(owned.unsubscribeOnce(), this.#cleanupTimeoutMs, "abandoned session unsubscribe", true);
		} catch (error) {
			failures.push(error);
		}
		try {
			await bounded(owned.disposeOnce(), this.#cleanupTimeoutMs, "abandoned session dispose", true);
		} catch (error) {
			failures.push(error);
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
	return Boolean(Reflect.apply(NATIVE_ABORTED_GETTER, signal, []));
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

interface TerminalCapture {
	messageEnd?: AssistantTerminal;
	agentEnd?: AssistantTerminal;
	messageEndCount: number;
	agentEndCount: number;
	agentEndWillRetry: boolean;
	piPhase: "initial" | "agent" | "turn" | "turn-ended" | "agent-ended" | "settled";
	piOpenMessageRole?: "user" | "assistant" | "toolResult";
	piTurnAssistant?: AssistantTerminal;
	piSettled: boolean;
	frozen: boolean;
	contentPhases: Map<number, { kind: "text" | "thinking" | "toolCall"; phase: "open" | "ended" }>;
	toolExecutions: Map<string, { toolName: string; argsIdentity: string; phase: "started" | "ended" }>;
	stream?: AssistantTerminal;
	failure?: AgentSessionRuntimeError;
	eventCount: number;
	eventBytes: number;
}

interface CapturedAssistantContent {
	type: "text" | "thinking" | "toolCall";
	text?: string;
	thinking?: string;
	id?: string;
	name?: string;
	arguments?: JsonEventValue;
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
			const matches = [...this.#active.values()].filter((active) => active.runId === runId);
			const owners = new Set(this.#creationsByRunId.get(runId) ?? []);
			for (const active of matches) {
				active.scope.cancel(new AgentSessionRuntimeError(`AgentSession run ${runId} was aborted`));
				void active.owned?.abortOnce().catch(() => undefined);
				if (active.creation) owners.add(active.creation);
			}
			await Promise.all([...admissions.map((admission) => admission.done), ...matches.map((active) => active.done)]);
			for (const owner of this.#creationsByRunId.get(runId) ?? []) owners.add(owner);
			if (owners.size > 0) {
				const joined = await Promise.all([...owners].map((owner) => owner.joinForAbort()));
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
		await Promise.all([...this.#active.values()].map((active) => active.done));
		const creationOwners = [...this.#creations];
		if (creationOwners.length > 0) {
			const joined = await Promise.all(creationOwners.map((creation) => creation.joinForClose()));
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
			failures.push(error);
		}
		if (this.#quarantineFailurePresent) {
			failures.push(this.#quarantineFailure);
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
		try {
			scope.assertActive();
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
			const sessionManager = this.#sdk.createSessionManager(request.workspace.cwd);
			const resourceLoader = this.#sdk.createResourceLoader({
				cwd: request.workspace.cwd,
				agentDir: this.#sdk.getAgentDir(),
				settingsManager,
				noExtensions: true,
				noSkills: true,
				noPromptTemplates: true,
				noThemes: true,
				noContextFiles: true,
				systemPrompt: prompts.systemPrompt,
			});
			reloadPromise = Promise.resolve().then(() => resourceLoader.reload());
			await scope.race(reloadPromise);
			scope.assertActive();

			const createOptions: CreateAgentSessionOptions = {
				cwd: request.workspace.cwd,
				agentDir: this.#sdk.getAgentDir(),
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
			const createPromise = Promise.resolve().then(() => this.#sdk.createAgentSession(createOptions));
			const owner = new SessionCreationOwnership(
				createPromise,
				this.#options.cleanupTimeoutMs,
				(created) => captureCreatedSession(created, thinking, expectedToolNames),
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
			const claim = creation.claim(created);
			owned = claim.owned;
			active.owned = owned;
			scope.setOnCancel(() => { void owned?.abortOnce().catch(() => undefined); });
			scope.assertActive();
			try {
				claim.validate();
			} catch (error) {
				throw error;
			}

			const capture = newTerminalCapture();
			owned.subscribe((event) => {
				captureEvent(capture, event, this.#options);
				if (capture.failure) void owned?.abortOnce().catch(() => undefined);
			});
			await scope.race(owned.prompt(prompts.userPrompt, {
				expandPromptTemplates: false,
				source: "extension",
			}));
			scope.assertActive();
			// Prompt settlement follows Pi's synchronous agent_settled dispatch. Freeze before
			// releasing the listener, then await release so idle/dispose callbacks cannot mutate
			// evidence that is about to be verified.
			capture.frozen = true;
			await scope.race(bounded(owned.unsubscribeOnce(), this.#options.cleanupTimeoutMs, "session unsubscribe"));
			scope.assertActive();
			if (capture.failure) throw capture.failure;
			const terminal = verifyTerminalCapture(capture);
			if (terminal.provider !== REQUIRED_PROVIDER || terminal.model !== REQUIRED_MODEL) {
				throw new AgentSessionRuntimeError("AgentSession terminal model routing mismatch; fallback is forbidden");
			}
			result = parseHandoff(assistantText(terminal), request, this.#options.maxAssistantBytes);
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
				: new Error("AgentSession cleanup failure was unavailable");
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
		if (this.#sdk.version !== REQUIRED_PI_VERSION ||
			(this.#sdk.requiredVersion !== undefined && this.#sdk.requiredVersion !== REQUIRED_PI_VERSION)) {
			throw new AgentSessionRuntimeError(`AgentSession Shepherd requires Pi ${REQUIRED_PI_VERSION}; found ${this.#sdk.version}`);
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
	if (!request || typeof request !== "object" || Array.isArray(request) || nodeTypes.isProxy(request)) {
		throw new AgentSessionRuntimeError("AgentSession request must be an object");
	}
	const bindingDescriptor = Reflect.getOwnPropertyDescriptor(request, "binding");
	if (!bindingDescriptor?.enumerable || bindingDescriptor.get || bindingDescriptor.set || !("value" in bindingDescriptor) ||
		!bindingDescriptor.value || typeof bindingDescriptor.value !== "object" ||
		Array.isArray(bindingDescriptor.value) || nodeTypes.isProxy(bindingDescriptor.value)) {
		throw new AgentSessionRuntimeError("request binding must be an own data field");
	}
	const runIdDescriptor = Reflect.getOwnPropertyDescriptor(bindingDescriptor.value, "runId");
	if (!runIdDescriptor?.enumerable || runIdDescriptor.get || runIdDescriptor.set || !("value" in runIdDescriptor) ||
		!validIdentifier(runIdDescriptor.value)) {
		throw new AgentSessionRuntimeError("request runId must be an own data field");
	}
	return runIdDescriptor.value;
}

function normalizeRunRequest(request: RoleRunRequest): RoleRunRequest {
	const requestFields = captureClosedRecordFields(request, "AgentSession request", 11);
	assertAllowedCapturedFields(requestFields, [
		"role", "task", "context", "timeoutMs", "deadlineAt", "signal", "workspace", "capabilities", "authority", "binding",
	], ["role", "task", "context", "timeoutMs", "workspace", "capabilities", "authority", "binding"], "request");
	const role = requestFields.get("role") as ShepherdAgentRole;
	const task = requestFields.get("task") as string;
	const contextSource = requestFields.get("context");
	const timeoutMs = requestFields.get("timeoutMs") as number;
	const deadlineAt = requestFields.get("deadlineAt") as number | undefined;
	const signal = requestFields.get("signal") as AbortSignal | undefined;
	const workspaceSource = requestFields.get("workspace");
	const capabilitiesSource = requestFields.get("capabilities");
	const authoritySource = requestFields.get("authority");
	const bindingSource = requestFields.get("binding");

	routeForRole(role);
	if (!Number.isSafeInteger(timeoutMs) || timeoutMs <= 0 || timeoutMs > MAX_TIMEOUT_MS) {
		throw new AgentSessionRuntimeError("timeoutMs must be a positive bounded safe integer");
	}
	if (deadlineAt !== undefined && (!Number.isSafeInteger(deadlineAt) || deadlineAt <= Date.now())) {
		throw new AgentSessionRuntimeError("deadlineAt must be a future epoch-millisecond safe integer");
	}
	if (signal !== undefined && !(signal instanceof AbortSignal)) {
		throw new AgentSessionRuntimeError("request signal is invalid");
	}

	const authorityFields = captureClosedRecordFields(authoritySource, "request authority", 7);
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
	if (typeof issue !== "number" || !Number.isSafeInteger(issue) || issue < 1) {
		throw new AgentSessionRuntimeError("authority issue is invalid");
	}
	if (typeof branch !== "string" || branch.length < 1 || branch.length > 255 ||
		/[\u0000-\u001f\u007f]/.test(branch) || branch === "main") {
		throw new AgentSessionRuntimeError("authority branch is invalid or targets main");
	}
	if (!validIdentifier(authorityWorkspaceId)) throw new AgentSessionRuntimeError("authority workspace identity is invalid");
	if (typeof readOnly !== "boolean") throw new AgentSessionRuntimeError("authority readOnly is invalid");
	const readPrefixes = frozenArray(normalizeScopedPrefixes(readPrefixesSource, "read"));
	const writePrefixes = readOnly && writePrefixesSource.length === 0
		? frozenArray<string>([])
		: frozenArray(normalizeScopedPrefixes(writePrefixesSource, "write"));
	const capabilityNames: string[] = [];
	for (const name of capabilityNamesSource) {
		if (typeof name !== "string") throw new AgentSessionRuntimeError("authority capability name is invalid");
		capabilityNames.push(name);
	}
	Object.freeze(capabilityNames);
	const authority = Object.freeze({
		issue: Number(issue),
		branch,
		workspaceId: authorityWorkspaceId,
		readOnly,
		readPrefixes,
		writePrefixes,
		capabilityNames,
	}) as RoleAuthority;

	const bindingFields = captureClosedRecordFields(bindingSource, "request binding", 5);
	assertExactCapturedFields(bindingFields, ["runId", "generation", "laneId", "candidateHead", "validationNonce"], "binding");
	const runId = bindingFields.get("runId");
	const generation = bindingFields.get("generation");
	const laneId = bindingFields.get("laneId");
	const candidateHead = bindingFields.get("candidateHead");
	const validationNonce = bindingFields.get("validationNonce");
	if (!validIdentifier(runId) || !validIdentifier(laneId) ||
		typeof generation !== "number" || !Number.isSafeInteger(generation) || generation < 1 ||
		typeof candidateHead !== "string" || !/^[0-9a-f]{40}$/.test(candidateHead) ||
		!validIdentifier(validationNonce) || validationNonce.length < 12) {
		throw new AgentSessionRuntimeError("request binding is invalid");
	}
	const binding = Object.freeze({ runId, generation: Number(generation), laneId, candidateHead, validationNonce });

	const workspaceFields = captureClosedRecordFields(workspaceSource, "workspace capability", 5);
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
	const workspace = Object.freeze({
		id: workspaceId,
		cwd: canonicalCwd,
		readText(path: string, options: { offset?: number; limit?: number; signal?: AbortSignal }) {
			return Reflect.apply(readText, workspaceSource as object, [path, options]);
		},
		editText(path: string, oldText: string, newText: string, operationSignal?: AbortSignal) {
			return Reflect.apply(editText, workspaceSource as object, [path, oldText, newText, operationSignal]);
		},
		writeText(path: string, content: string, operationSignal?: AbortSignal) {
			return Reflect.apply(writeText, workspaceSource as object, [path, content, operationSignal]);
		},
	}) satisfies ScopedWorkspace;

	const capabilityValues = captureFreshDenseArray(capabilitiesSource, "typed host capabilities", 32, true);
	const capabilities: HostCapability[] = [];
	for (const capability of capabilityValues) {
		capabilities.push(normalizeCapability(capability as HostCapability));
	}
	Object.freeze(capabilities);
	const contextValues = captureFreshDenseArray(contextSource, "role context", 64, true);
	const context: string[] = [];
	for (const entry of contextValues) {
		if (typeof entry !== "string") throw new AgentSessionRuntimeError("role context item is invalid");
		context.push(entry);
	}
	Object.freeze(context);
	const normalized = Object.freeze({
		role,
		task,
		context,
		timeoutMs,
		deadlineAt,
		signal,
		workspace,
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
	const fields = captureClosedRecordFields(capability, "capability", 5);
	assertExactCapturedFields(fields, ["name", "description", "mutates", "parameters", "execute"], "capability");
	const name = fields.get("name") as string;
	const description = fields.get("description") as string;
	const mutates = fields.get("mutates") as boolean;
	const parameters = fields.get("parameters") as HostCapability["parameters"];
	const execute = fields.get("execute");
	if (typeof execute !== "function") throw new AgentSessionRuntimeError("capability execute must be a function");
	return Object.freeze({
		name,
		description,
		mutates,
		parameters,
		execute(input: Readonly<Record<string, unknown>>, signal?: AbortSignal) {
			return Reflect.apply(execute, capability, [input, signal]);
		},
	});
}

function captureFreshDenseArray(
	value: unknown,
	description: string,
	maximum: number,
	allowEmpty: boolean,
): unknown[] {
	if (!Array.isArray(value) || nodeTypes.isProxy(value) || Object.getPrototypeOf(value) !== Array.prototype) {
		throw new AgentSessionRuntimeError(`${description} must be a plain non-proxy array`);
	}
	const lengthDescriptor = Reflect.getOwnPropertyDescriptor(value, "length");
	const lengthValue = lengthDescriptor && "value" in lengthDescriptor ? lengthDescriptor.value : undefined;
	if (!lengthDescriptor || lengthDescriptor.get || lengthDescriptor.set || !("value" in lengthDescriptor) ||
		typeof lengthValue !== "number" || !Number.isSafeInteger(lengthValue) || lengthValue < (allowEmpty ? 0 : 1) ||
		lengthValue > maximum) {
		throw new AgentSessionRuntimeError(`${description} has an invalid authoritative length`);
	}
	const length = lengthValue;
	const ownKeys = Reflect.ownKeys(value);
	if (ownKeys.length !== length + 1 || ownKeys.some((key) =>
		typeof key !== "string" || (key !== "length" && !/^(?:0|[1-9][0-9]*)$/.test(key)))) {
		throw new AgentSessionRuntimeError(`${description} contains hidden, symbol, sparse, or extra behavior`);
	}
	const captured: unknown[] = [];
	for (let index = 0; index < length; index += 1) {
		const descriptor = Reflect.getOwnPropertyDescriptor(value, String(index));
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
			throw new AgentSessionRuntimeError(`${description} contains a sparse or accessor element`);
		}
		captured.push(descriptor.value);
	}
	return captured;
}

function mutationLeaseFor(request: RoleRunRequest): MutationLease {
	return Object.freeze({
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
): CreatedSessionClaim {
	if (!created || typeof created !== "object" || Array.isArray(created) || nodeTypes.isProxy(created)) {
		throw new AgentSessionRuntimeError("Pi returned an invalid AgentSession result");
	}
	// Acquire the cleanup root before validating any peer field. The legacy public Pi result
	// permits a session getter, so it is invoked exactly once; every other result field must be
	// a data descriptor and is never read through ordinary property lookup.
	const sessionDescriptor = Object.getOwnPropertyDescriptor(created, "session");
	let session: unknown;
	if (sessionDescriptor && "value" in sessionDescriptor) session = sessionDescriptor.value;
	else if (sessionDescriptor?.get) session = Reflect.apply(sessionDescriptor.get, created, []);
	if (!session || typeof session !== "object") {
		throw new AgentSessionRuntimeError("Pi returned an invalid AgentSession result without a cleanable session");
	}
	const owned = new OwnedSession(session as RuntimeAgentSession);
	const captureFailures: unknown[] = [...owned.validationFailures()];
	let modelProvider: unknown;
	let modelId: unknown;
	let thinkingLevel: unknown;
	let sessionFile: unknown;
	let activeTools: readonly string[] | undefined;
	const recordCaptureFailure = (error: unknown): void => { captureFailures.push(error); };
	if (!sessionDescriptor?.enumerable || sessionDescriptor.set ||
		(!("value" in sessionDescriptor) && !sessionDescriptor.get)) {
		recordCaptureFailure(new AgentSessionRuntimeError("Pi session ownership descriptor is invalid"));
	}
	try {
		assertEnumerableFieldsAllowed(
			created,
			["session", "extensionsResult", "modelFallbackMessage"],
			"Pi AgentSession result",
		);
	} catch (error) {
		recordCaptureFailure(error);
	}
	const extensionsDescriptor = Object.getOwnPropertyDescriptor(created, "extensionsResult");
	if (!extensionsDescriptor?.enumerable || extensionsDescriptor.get || extensionsDescriptor.set ||
		!("value" in extensionsDescriptor)) {
		recordCaptureFailure(new AgentSessionRuntimeError("Pi returned an invalid extensions result descriptor"));
	} else {
		try {
			captureEmptyExtensionsResult(extensionsDescriptor.value);
		} catch (error) {
			recordCaptureFailure(error);
		}
	}
	const fallbackDescriptor = Object.getOwnPropertyDescriptor(created, "modelFallbackMessage");
	if (!fallbackDescriptor?.enumerable || fallbackDescriptor.get || fallbackDescriptor.set || !("value" in fallbackDescriptor)) {
		recordCaptureFailure(new AgentSessionRuntimeError("Pi returned an invalid fallback descriptor"));
	} else if (fallbackDescriptor.value !== undefined) {
		recordCaptureFailure(new AgentSessionRuntimeError("Pi attempted a forbidden model fallback"));
	}
	try {
		const model = (session as RuntimeAgentSession).model;
		modelProvider = model?.provider;
		modelId = model?.id;
	} catch (error) {
		recordCaptureFailure(error);
	}
	try {
		thinkingLevel = (session as RuntimeAgentSession).thinkingLevel;
	} catch (error) {
		recordCaptureFailure(error);
	}
	try {
		sessionFile = (session as RuntimeAgentSession).sessionFile;
	} catch (error) {
		recordCaptureFailure(error);
	}
	try {
		activeTools = captureToolNameArray(owned.activeToolNames());
	} catch (error) {
		recordCaptureFailure(error);
	}

	return Object.freeze({
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

function captureEmptyExtensionsResult(value: unknown): void {
	if (!value || typeof value !== "object" || Array.isArray(value) || nodeTypes.isProxy(value)) {
		throw new AgentSessionRuntimeError("Pi returned an invalid extensions result");
	}
	assertEnumerableFieldsAllowed(value, ["extensions", "errors", "runtime"], "Pi extensions result");
	for (const key of ["extensions", "errors"] as const) {
		const descriptor = Object.getOwnPropertyDescriptor(value, key);
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
			throw new AgentSessionRuntimeError(`Pi extensions result ${key} must be an own data field`);
		}
		captureExactEmptyArray(descriptor.value, `Pi extensions result ${key}`);
	}
	const runtimeDescriptor = Object.getOwnPropertyDescriptor(value, "runtime");
	if (!runtimeDescriptor?.enumerable || runtimeDescriptor.get || runtimeDescriptor.set ||
		!("value" in runtimeDescriptor)) {
		throw new AgentSessionRuntimeError("Pi extensions result runtime must be an own data field");
	}
}

function assertEnumerableFieldsAllowed(
	value: object,
	allowed: readonly string[],
	description: string,
): void {
	const allowedSet = new Set(allowed);
	let count = 0;
	for (const key in value) {
		if (!Object.hasOwn(value, key) || !allowedSet.has(key)) {
			throw new AgentSessionRuntimeError(`${description} contains an unknown enumerable field`);
		}
		count += 1;
		if (count > allowed.length) throw new AgentSessionRuntimeError(`${description} contains too many fields`);
	}
}

function captureExactEmptyArray(value: unknown, description: string): void {
	if (!Array.isArray(value) || nodeTypes.isProxy(value)) {
		throw new AgentSessionRuntimeError(`${description} must be an exact non-proxy empty array`);
	}
	const lengthDescriptor = Object.getOwnPropertyDescriptor(value, "length");
	if (!lengthDescriptor || lengthDescriptor.get || lengthDescriptor.set || !("value" in lengthDescriptor) ||
		lengthDescriptor.value !== 0) {
		throw new AgentSessionRuntimeError(`${description} must be empty`);
	}
	for (const key in value) {
		if (Object.hasOwn(value, key)) {
			throw new AgentSessionRuntimeError(`${description} contains an enumerable item or extra field`);
		}
		throw new AgentSessionRuntimeError(`${description} contains an inherited field`);
	}
}

function captureToolNameArray(value: unknown): readonly string[] | undefined {
	if (!Array.isArray(value) || nodeTypes.isProxy(value) || value.length > 256) return undefined;
	let enumerableItems = 0;
	for (const key in value) {
		if (!Object.hasOwn(value, key) || !/^(?:0|[1-9][0-9]*)$/.test(key) || Number(key) >= value.length) return undefined;
		enumerableItems += 1;
		if (enumerableItems > value.length) return undefined;
	}
	if (enumerableItems !== value.length) return undefined;
	const names: string[] = [];
	for (let index = 0; index < value.length; index += 1) {
		const descriptor = Reflect.getOwnPropertyDescriptor(value, String(index));
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor) ||
			typeof descriptor.value !== "string") {
			return undefined;
		}
		names.push(descriptor.value);
	}
	return Object.freeze(names);
}

function newTerminalCapture(): TerminalCapture {
	return {
		messageEndCount: 0,
		agentEndCount: 0,
		agentEndWillRetry: false,
		piPhase: "initial",
		piSettled: false,
		frozen: false,
		contentPhases: new Map(),
		toolExecutions: new Map(),
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
		const eventFields = captureClosedRecordFields(event, "AgentSession event", MAX_EVENT_RECORD_KEYS);
		const eventType = eventFields.get("type");
		if (typeof eventType !== "string") throw new AgentSessionRuntimeError("AgentSession event type is invalid");
		const directCharge = capturePiLifecycleEvent(capture, eventFields, eventType, options);
		capture.eventBytes += directCharge ?? boundedEventBytes(event, options.maxEventBytes - capture.eventBytes);
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
			if ([...capture.toolExecutions.values()].some((execution) => execution.phase !== "ended")) {
				throw new AgentSessionRuntimeError("turn_start preceded completion of tool execution");
			}
			capture.toolExecutions.clear();
			return undefined;
		case "message_start": {
			assertExactCapturedFields(fields, ["type", "message"], "message_start event");
			if (capture.piPhase !== "turn" || capture.piOpenMessageRole !== undefined) {
				throw new AgentSessionRuntimeError("message_start is out of order");
			}
			const role = capturePiMessageRole(fields.get("message"));
			capture.piOpenMessageRole = role;
			if (role === "assistant") {
				const initial = captureAssistantTerminal(fields.get("message"), true, options.maxEventBytes);
				if (!initial) throw new AgentSessionRuntimeError("assistant message_start is invalid");
				capture.stream = initial;
				capture.contentPhases.clear();
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
				const terminal = captureAssistantTerminal(fields.get("message"), false, options.maxEventBytes);
				if (!terminal) throw new AgentSessionRuntimeError("assistant message_end is invalid");
				assertCompletedContentPhases(capture);
				capture.messageEndCount += 1;
				capture.messageEnd = terminal;
				capture.piTurnAssistant = terminal;
			}
			capture.piOpenMessageRole = undefined;
			return undefined;
		}
		case "turn_end": {
			assertExactCapturedFields(fields, ["type", "message", "toolResults"], "turn_end event");
			if (capture.piPhase !== "turn" || capture.piOpenMessageRole !== undefined || !capture.piTurnAssistant) {
				throw new AgentSessionRuntimeError("turn_end is out of order");
			}
			const terminal = captureAssistantTerminal(fields.get("message"), false, options.maxEventBytes);
			if (!terminal || !sameTerminal(terminal, capture.piTurnAssistant)) {
				throw new AgentSessionRuntimeError("turn_end assistant does not match message_end");
			}
			captureDenseArray(fields.get("toolResults"), "turn_end tool results");
			if ([...capture.toolExecutions.values()].some((execution) => execution.phase !== "ended")) {
				throw new AgentSessionRuntimeError("turn_end preceded completion of tool execution");
			}
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
				const assistant = captureAssistantTerminal(message, false, options.maxEventBytes);
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
			if (capture.toolExecutions.has(toolCallId)) {
				throw new AgentSessionRuntimeError("tool execution ID is duplicated");
			}
			const args = snapshotEventJson(fields.get("args"), "tool execution arguments", options.maxEventBytes);
			capture.toolExecutions.set(toolCallId, { toolName, argsIdentity: canonicalJson(args), phase: "started" });
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
			const execution = capture.toolExecutions.get(toolCallId);
			const args = snapshotEventJson(fields.get("args"), "tool execution arguments", options.maxEventBytes);
			if (!execution || execution.phase !== "started" || execution.toolName !== toolName ||
				execution.argsIdentity !== canonicalJson(args)) {
				throw new AgentSessionRuntimeError("tool execution update does not match its start");
			}
			snapshotEventJson(fields.get("partialResult"), "tool execution partial result", options.maxEventBytes);
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
			const execution = capture.toolExecutions.get(toolCallId);
			if (!execution || execution.phase !== "started" || execution.toolName !== toolName ||
				typeof fields.get("isError") !== "boolean") {
				throw new AgentSessionRuntimeError("tool execution end does not match its start");
			}
			snapshotEventJson(fields.get("result"), "tool execution result", options.maxEventBytes);
			execution.phase = "ended";
			return undefined;
		}
		default:
			throw new AgentSessionRuntimeError(`unsupported Pi AgentSession event ${JSON.stringify(type)}`);
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
	const fields = captureClosedRecordFields(value, "Pi lifecycle message", MAX_EVENT_RECORD_KEYS);
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
	const fields = captureClosedRecordFields(value, "Pi assistant streaming event", 8);
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
	if (!shape) throw new AgentSessionRuntimeError(`unsupported Pi assistant streaming event ${JSON.stringify(type)}`);
	assertExactCapturedFields(fields, shape, `Pi ${type} streaming event`);
	const message = captureAssistantTerminal(messageValue, true, maximum);
	if (!message) throw new AgentSessionRuntimeError("message_update did not contain an assistant message");
	const innerValue = fields.has("partial") ? fields.get("partial") : fields.has("message")
		? fields.get("message") : fields.get("error");
	const inner = captureAssistantTerminal(innerValue, true, maximum);
	if (!inner || inner.identity !== message.identity) {
		throw new AgentSessionRuntimeError(`Pi ${type} message and cumulative snapshot disagree`);
	}
	if (fields.has("contentIndex") && (!Number.isSafeInteger(fields.get("contentIndex")) || Number(fields.get("contentIndex")) < 0)) {
		throw new AgentSessionRuntimeError(`Pi ${type} content index is invalid`);
	}
	const variable = fields.has("delta") ? fields.get("delta") : undefined;
	if (variable !== undefined && typeof variable !== "string") {
		throw new AgentSessionRuntimeError(`Pi ${type} delta is invalid`);
	}
	if (type === "done" && !["stop", "length", "toolUse"].includes(String(fields.get("reason")))) {
		throw new AgentSessionRuntimeError("Pi done reason is invalid");
	}
	if (type === "error" && !["aborted", "error"].includes(String(fields.get("reason")))) {
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
	const index = Number(fields.get("contentIndex"));
	const previousContent = previous?.content ?? [];
	if (index >= current.content.length || current.content.length < previousContent.length ||
		current.content.length > previousContent.length + 1 ||
		(current.content.length === previousContent.length + 1 && index !== previousContent.length)) {
		throw new AgentSessionRuntimeError(`Pi ${type} content index skipped or replaced stream state`);
	}
	for (let cursor = 0; cursor < Math.min(previousContent.length, current.content.length); cursor += 1) {
		if (cursor !== index && previousContent[cursor]?.identity !== current.content[cursor]?.identity) {
			throw new AgentSessionRuntimeError(`Pi ${type} replaced an unrelated content item`);
		}
	}
	const prior = previousContent[index];
	const next = current.content[index];
	if (!next) throw new AgentSessionRuntimeError(`Pi ${type} content item is missing`);
	const kind = type.startsWith("text_") ? "text" : type.startsWith("thinking_") ? "thinking" : "toolCall";
	const transition = type.endsWith("_start") ? "start" : type.endsWith("_delta") ? "delta" : "end";
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
	if (type.startsWith("text_")) {
		if (next.type !== "text") throw new AgentSessionRuntimeError(`Pi ${type} content type is invalid`);
		const before = prior?.type === "text" ? prior.text ?? "" : "";
		const after = next.text ?? "";
		if (type === "text_start" && after !== "") throw new AgentSessionRuntimeError("Pi text_start must begin empty");
		if (type === "text_delta" && after !== before + String(delta)) {
			throw new AgentSessionRuntimeError("Pi text_delta is not the actual novel suffix");
		}
		if (type === "text_end" && (fields.get("content") !== after || !after.startsWith(before))) {
			throw new AgentSessionRuntimeError("Pi text_end replaced or misreported content");
		}
		const novelBytes = byteLength(after.slice(before.length));
		return envelopeGrowth + cumulativeIdentityGrowth(prior?.identity, next.identity, novelBytes);
	}
	if (type.startsWith("thinking_")) {
		if (next.type !== "thinking") throw new AgentSessionRuntimeError(`Pi ${type} content type is invalid`);
		const before = prior?.type === "thinking" ? prior.thinking ?? "" : "";
		const after = next.thinking ?? "";
		if (type === "thinking_start" && after !== "") throw new AgentSessionRuntimeError("Pi thinking_start must begin empty");
		if (type === "thinking_delta" && after !== before + String(delta)) {
			throw new AgentSessionRuntimeError("Pi thinking_delta is not the actual novel suffix");
		}
		if (type === "thinking_end" && (fields.get("content") !== after || !after.startsWith(before))) {
			throw new AgentSessionRuntimeError("Pi thinking_end replaced or misreported content");
		}
		const novelBytes = byteLength(after.slice(before.length));
		return envelopeGrowth + cumulativeIdentityGrowth(prior?.identity, next.identity, novelBytes);
	}
	if (!type.startsWith("toolcall_") || next.type !== "toolCall") {
		throw new AgentSessionRuntimeError(`Pi ${type} content type is invalid`);
	}
	if (type === "toolcall_start") {
		if ((next.partialJson ?? "") !== "") throw new AgentSessionRuntimeError("Pi toolcall_start must begin empty");
		return envelopeGrowth + cumulativeIdentityGrowth(prior?.identity, next.identity, 0);
	}
	if (type === "toolcall_delta") {
		const before = prior?.type === "toolCall" ? prior.partialJson ?? "" : "";
		if (next.partialJson !== before + String(delta)) {
			throw new AgentSessionRuntimeError("Pi toolcall_delta is not the actual novel suffix");
		}
		return envelopeGrowth + cumulativeIdentityGrowth(prior?.identity, next.identity, byteLength(String(delta)));
	}
	const toolCall = snapshotEventJson(fields.get("toolCall"), "Pi toolcall_end value", MAX_EVENT_BYTES);
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
	if (previous === undefined) return Math.max(novelBytes, byteLength(current));
	if (previous === current) return novelBytes;
	const previousBytes = byteLength(previous);
	const currentBytes = byteLength(current);
	if (novelBytes > 0 && currentBytes > previousBytes) {
		return Math.max(novelBytes, currentBytes - previousBytes);
	}
	// A changed identity without provable append-only growth is replacement state. Charge the
	// complete replacement so equal-size or shrinking metadata cannot evade the aggregate bound.
	return Math.max(novelBytes, currentBytes);
}

function boundedEventBytes(root: unknown, maximum: number): number {
	if (maximum < 0) throw new AgentSessionRuntimeError("AgentSession event byte bound was exhausted");
	const stack: Array<{ value: unknown; depth: number }> = [{ value: root, depth: 0 }];
	const seen = new WeakSet<object>();
	let bytes = 0;
	let nodes = 0;
	const add = (count: number) => {
		bytes += count;
		if (bytes > maximum) throw new AgentSessionRuntimeError("AgentSession event exceeded its byte bound");
	};
	while (stack.length > 0) {
		const current = stack.pop();
		if (!current) break;
		nodes += 1;
		if (nodes > MAX_EVENT_NODES || current.depth > MAX_EVENT_DEPTH) {
			throw new AgentSessionRuntimeError("AgentSession event exceeded its depth or node bound");
		}
		const value = current.value;
		if (value === null) { add(4); continue; }
		switch (typeof value) {
			case "string": add(byteLength(value) + 2); continue;
			case "number": add(24); continue;
			case "boolean": add(5); continue;
			case "undefined": add(4); continue;
			case "object": break;
			default: throw new AgentSessionRuntimeError("AgentSession event contains an unsupported value");
		}
		const object = value as object;
		if (nodeTypes.isProxy(object)) throw new AgentSessionRuntimeError("AgentSession event contains a proxy");
		if (seen.has(object)) throw new AgentSessionRuntimeError("AgentSession event contains a cycle");
		seen.add(object);
		const array = Array.isArray(value);
		if (array) {
			if (value.length > MAX_EVENT_ARRAY_ITEMS) {
				throw new AgentSessionRuntimeError("AgentSession event contains a sparse or oversized array");
			}
		}
		add(2);
		let enumerableKeys = 0;
		for (const key in value) {
			if (!Object.hasOwn(value, key)) throw new AgentSessionRuntimeError("AgentSession event contains inherited fields");
			enumerableKeys += 1;
			if (enumerableKeys > (array ? value.length : MAX_EVENT_RECORD_KEYS)) {
				throw new AgentSessionRuntimeError("AgentSession event record is too wide or sparse");
			}
			if (array && (!/^(?:0|[1-9][0-9]*)$/.test(key) || Number(key) >= value.length)) {
				throw new AgentSessionRuntimeError("AgentSession event array contains an extra field");
			}
			const descriptor = Object.getOwnPropertyDescriptor(object, key);
			if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
				throw new AgentSessionRuntimeError("AgentSession event contains an accessor or invalid field");
			}
			add(byteLength(key) + 3);
			stack.push({ value: descriptor.value, depth: current.depth + 1 });
		}
		if (array && enumerableKeys !== value.length) {
			throw new AgentSessionRuntimeError("AgentSession event contains a sparse array");
		}
	}
	return bytes;
}

function captureClosedRecordFields(
	value: unknown,
	description: string,
	maximumKeys: number,
): ReadonlyMap<string, unknown> {
	if (!value || typeof value !== "object" || Array.isArray(value) || nodeTypes.isProxy(value)) {
		throw new AgentSessionRuntimeError(`${description} must be a plain non-proxy record`);
	}
	const prototype = Object.getPrototypeOf(value);
	if (prototype !== Object.prototype && prototype !== null) {
		throw new AgentSessionRuntimeError(`${description} must have a plain prototype`);
	}
	const fields = new Map<string, unknown>();
	for (const key in value) {
		if (!Object.hasOwn(value, key)) throw new AgentSessionRuntimeError(`${description} contains inherited fields`);
		if (fields.size >= maximumKeys) throw new AgentSessionRuntimeError(`${description} contains too many fields`);
		const descriptor = Reflect.getOwnPropertyDescriptor(value, key);
		if (!descriptor?.enumerable) throw new AgentSessionRuntimeError(`${description} contains an invalid field`);
		if (descriptor.get || descriptor.set || !("value" in descriptor)) {
			throw new AgentSessionRuntimeError(`${description} contains an accessor field`);
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
	if (fields.size !== expected.length || expected.some((name) => !fields.has(name))) {
		throw new AgentSessionRuntimeError(`${description} must be an exact closed record`);
	}
}

function captureDenseArray(value: unknown, description: string): readonly unknown[] {
	if (!Array.isArray(value) || nodeTypes.isProxy(value) || value.length > MAX_EVENT_ARRAY_ITEMS) {
		throw new AgentSessionRuntimeError(`${description} must be a bounded dense non-proxy array`);
	}
	let enumerableItems = 0;
	for (const key in value) {
		if (!Object.hasOwn(value, key) || !/^(?:0|[1-9][0-9]*)$/.test(key) || Number(key) >= value.length) {
			throw new AgentSessionRuntimeError(`${description} contains an inherited or extra field`);
		}
		enumerableItems += 1;
		if (enumerableItems > value.length) {
			throw new AgentSessionRuntimeError(`${description} must not be sparse or contain extra fields`);
		}
	}
	if (enumerableItems !== value.length) {
		throw new AgentSessionRuntimeError(`${description} must not be sparse or contain extra fields`);
	}
	const captured: unknown[] = [];
	for (let index = 0; index < value.length; index += 1) {
		const descriptor = Reflect.getOwnPropertyDescriptor(value, String(index));
		if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
			throw new AgentSessionRuntimeError(`${description} contains an invalid array element`);
		}
		captured.push(descriptor.value);
	}
	return Object.freeze(captured);
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

function snapshotEventJson(value: unknown, description: string, maximum: number): JsonEventValue {
	return snapshotEventJsonValue(
		value,
		0,
		{ nodes: 0, keys: 0, bytes: 0, maximum },
		new WeakSet<object>(),
		description,
	);
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
		if (!Number.isFinite(value)) throw new AgentSessionRuntimeError(`${description} contains a non-JSON number`);
		add(24);
		return value;
	}
	if (typeof value !== "object") throw new AgentSessionRuntimeError(`${description} contains a non-JSON value`);
	if (nodeTypes.isProxy(value)) throw new AgentSessionRuntimeError(`${description} contains a proxy`);
	if (ancestors.has(value)) throw new AgentSessionRuntimeError(`${description} contains a cycle`);
	ancestors.add(value);
	add(2);
	try {
		if (Array.isArray(value)) {
			if (value.length > MAX_EVENT_ARRAY_ITEMS) throw new AgentSessionRuntimeError(`${description} array is oversized`);
			let enumerableItems = 0;
			for (const key in value) {
				if (!Object.hasOwn(value, key) || !/^(?:0|[1-9][0-9]*)$/.test(key) || Number(key) >= value.length) {
					throw new AgentSessionRuntimeError(`${description} array contains an inherited or extra field`);
				}
				enumerableItems += 1;
				if (enumerableItems > value.length) throw new AgentSessionRuntimeError(`${description} array is invalid`);
			}
			if (enumerableItems !== value.length) throw new AgentSessionRuntimeError(`${description} array is sparse`);
			budget.keys += value.length;
			if (budget.keys > MAX_EVENT_NODES) throw new AgentSessionRuntimeError(`${description} exceeded its key bound`);
			const snapshot: JsonEventValue[] = [];
			for (let index = 0; index < value.length; index += 1) {
				const descriptor = Object.getOwnPropertyDescriptor(value, String(index));
				if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
					throw new AgentSessionRuntimeError(`${description} array contains an invalid item`);
				}
				snapshot.push(snapshotEventJsonValue(descriptor.value, depth + 1, budget, ancestors, description));
			}
			return Object.freeze(snapshot);
		}
		const prototype = Object.getPrototypeOf(value);
		if (prototype !== Object.prototype && prototype !== null) {
			throw new AgentSessionRuntimeError(`${description} contains a non-plain object`);
		}
		const keys: string[] = [];
		for (const key in value) {
			if (!Object.hasOwn(value, key)) throw new AgentSessionRuntimeError(`${description} contains inherited data`);
			keys.push(key);
			budget.keys += 1;
			if (keys.length > MAX_EVENT_RECORD_KEYS || budget.keys > MAX_EVENT_NODES) {
				throw new AgentSessionRuntimeError(`${description} exceeded its key bound`);
			}
		}
		keys.sort();
		const snapshot = Object.create(null) as Record<string, JsonEventValue>;
		for (const key of keys) {
			const descriptor = Object.getOwnPropertyDescriptor(value, key);
			if (!descriptor?.enumerable || descriptor.get || descriptor.set || !("value" in descriptor)) {
				throw new AgentSessionRuntimeError(`${description} contains an invalid field`);
			}
			add(byteLength(key) + 3);
			Object.defineProperty(snapshot, key, {
				value: snapshotEventJsonValue(descriptor.value, depth + 1, budget, ancestors, description),
				enumerable: true,
				writable: false,
				configurable: false,
			});
		}
		return Object.freeze(snapshot);
	} finally {
		ancestors.delete(value);
	}
}

function canonicalJson(value: JsonEventValue): string {
	return JSON.stringify(value);
}

function canonicalToolContent(value: CapturedAssistantContent, _terminal: boolean): string {
	return value.terminalIdentity ?? value.identity;
}

function captureAssistantTerminal(
	value: unknown,
	streaming = false,
	maximum = MAX_EVENT_BYTES,
): AssistantTerminal | undefined {
	const fields = captureClosedRecordFields(value, "AgentSession terminal message", 32);
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
		typeof timestamp !== "number" || !Number.isFinite(timestamp)) {
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
	const content = captureDenseArray(fields.get("content"), "AgentSession assistant terminal content").map((part) => {
		const partFields = captureClosedRecordFields(part, "AgentSession assistant content part", 16);
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
			return Object.freeze({ type, text, identity });
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
			return Object.freeze({ type, thinking, identity });
		}
		if (type === "toolCall") {
			assertAllowedCapturedFields(partFields, ["type", "id", "name", "arguments", "thoughtSignature", ...(streaming ? ["partialJson"] : [])],
				["type", "id", "name", "arguments"], "AgentSession assistant tool-call content");
			const id = partFields.get("id");
			const name = partFields.get("name");
			if (typeof id !== "string" || typeof name !== "string") {
				throw new AgentSessionRuntimeError("AgentSession assistant tool-call identity is invalid");
			}
			const args = snapshotEventJson(partFields.get("arguments"), "AgentSession assistant tool arguments", maximum);
			const thoughtSignature = optionalCapturedString(partFields, "thoughtSignature", "assistant tool thought signature");
			const partialJson = optionalCapturedString(partFields, "partialJson", "assistant tool partial JSON");
			const terminalValue = snapshotEventJson({
				type,
				id,
				name,
				arguments: args,
				...(thoughtSignature === undefined ? {} : { thoughtSignature }),
			}, "AgentSession assistant tool identity", maximum);
			const terminalIdentity = canonicalJson(terminalValue);
			const identity = partialJson === undefined
				? terminalIdentity
				: canonicalJson(snapshotEventJson({ terminal: terminalValue, partialJson },
					"AgentSession assistant streaming tool identity", maximum));
			return Object.freeze({ type, id, name, arguments: args, partialJson, identity, terminalIdentity });
		}
		throw new AgentSessionRuntimeError(`AgentSession assistant content type ${JSON.stringify(type)} is invalid`);
	});
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
	const identity = canonicalJson(snapshotEventJson({
		...envelope,
		content: content.map((part) => part.identity),
	}, "AgentSession assistant identity", maximum));
	return Object.freeze({
		role: "assistant",
		api,
		provider,
		model,
		stopReason,
		timestamp,
		content: Object.freeze(content),
		envelopeIdentity,
		identity,
	});
}

function captureAssistantDiagnostics(value: unknown, maximum: number): JsonEventValue {
	const diagnostics = captureDenseArray(value, "AgentSession assistant diagnostics");
	const projected: JsonEventValue[] = [];
	for (const diagnostic of diagnostics) {
		const fields = captureClosedRecordFields(diagnostic, "AgentSession assistant diagnostic", 4);
		assertAllowedCapturedFields(fields, ["type", "timestamp", "error", "details"], ["type", "timestamp"],
			"AgentSession assistant diagnostic");
		const type = fields.get("type");
		const timestamp = fields.get("timestamp");
		if (typeof type !== "string" || type.length === 0 || type.length > 256 ||
			typeof timestamp !== "number" || !Number.isFinite(timestamp)) {
			throw new AgentSessionRuntimeError("AgentSession assistant diagnostic identity is invalid");
		}
		const output: Record<string, JsonEventValue> = { type, timestamp };
		const errorValue = fields.get("error");
		if (errorValue !== undefined) {
			const errorFields = captureClosedRecordFields(errorValue, "AgentSession assistant diagnostic error", 4);
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
				if (typeof code !== "string" && (typeof code !== "number" || !Number.isFinite(code))) {
					throw new AgentSessionRuntimeError("AgentSession assistant diagnostic error code is invalid");
				}
				errorOutput.code = code;
			}
			output.error = errorOutput;
		}
		const detailsValue = fields.get("details");
		if (detailsValue !== undefined) {
			const details = projectDiagnosticJson(detailsValue, 0, { nodes: 0, keys: 0 }, new WeakSet<object>());
			if (details === undefined || !details || Array.isArray(details) || typeof details !== "object") {
				throw new AgentSessionRuntimeError("AgentSession assistant diagnostic details are invalid");
			}
			output.details = details;
		}
		projected.push(snapshotEventJson(output, "AgentSession assistant diagnostic snapshot", maximum));
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
		if (!Number.isFinite(value)) throw new AgentSessionRuntimeError("AgentSession assistant diagnostic contains a non-JSON number");
		return value;
	}
	if (typeof value !== "object" || nodeTypes.isProxy(value)) {
		throw new AgentSessionRuntimeError("AgentSession assistant diagnostic contains a non-JSON value");
	}
	if (ancestors.has(value)) throw new AgentSessionRuntimeError("AgentSession assistant diagnostic contains a cycle");
	ancestors.add(value);
	try {
		if (Array.isArray(value)) {
			const items = captureDenseArray(value, "AgentSession assistant diagnostic array");
			budget.keys += items.length;
			if (budget.keys > MAX_EVENT_NODES) throw new AgentSessionRuntimeError("AgentSession assistant diagnostic is too wide");
			const output: JsonEventValue[] = [];
			for (const item of items) {
				const projected = projectDiagnosticJson(item, depth + 1, budget, ancestors);
				if (projected === undefined) {
					throw new AgentSessionRuntimeError("AgentSession assistant diagnostic array contains undefined");
				}
				output.push(projected);
			}
			return Object.freeze(output);
		}
		const fields = captureClosedRecordFields(value, "AgentSession assistant diagnostic record", MAX_EVENT_RECORD_KEYS);
		budget.keys += fields.size;
		if (budget.keys > MAX_EVENT_NODES) throw new AgentSessionRuntimeError("AgentSession assistant diagnostic is too wide");
		const output = Object.create(null) as Record<string, JsonEventValue>;
		for (const [key, entry] of fields) {
			const projected = projectDiagnosticJson(entry, depth + 1, budget, ancestors);
			if (projected === undefined) continue;
			Object.defineProperty(output, key, {
				value: projected,
				enumerable: true,
				writable: false,
				configurable: false,
			});
		}
		return Object.freeze(output);
	} finally {
		ancestors.delete(value);
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
	const fields = captureClosedRecordFields(value, "AgentSession assistant usage", 16);
	assertAllowedCapturedFields(fields, [
		"input", "output", "cacheRead", "cacheWrite", "cacheWrite1h", "reasoning", "totalTokens", "cost",
	], ["input", "output", "cacheRead", "cacheWrite", "totalTokens", "cost"], "AgentSession assistant usage");
	for (const name of ["input", "output", "cacheRead", "cacheWrite", "cacheWrite1h", "reasoning", "totalTokens"] as const) {
		if (!fields.has(name)) continue;
		const entry = fields.get(name);
		if (typeof entry !== "number" || !Number.isFinite(entry) || entry < 0) {
			throw new AgentSessionRuntimeError(`AgentSession assistant usage ${name} is invalid`);
		}
	}
	const cost = captureClosedRecordFields(fields.get("cost"), "AgentSession assistant usage cost", 8);
	assertAllowedCapturedFields(cost, ["input", "output", "cacheRead", "cacheWrite", "total"],
		["input", "output", "cacheRead", "cacheWrite", "total"], "AgentSession assistant usage cost");
	for (const name of ["input", "output", "cacheRead", "cacheWrite", "total"] as const) {
		const entry = cost.get(name);
		if (typeof entry !== "number" || !Number.isFinite(entry) || entry < 0) {
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
		if (!allowedSet.has(key)) throw new AgentSessionRuntimeError(`${description} contains unknown field ${JSON.stringify(key)}`);
	}
	if (required.some((key) => !fields.has(key))) {
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
	return terminal.content.filter((part) => part.type === "text" && typeof part.text === "string")
		.map((part) => part.text).join("").trim();
}

function parseHandoff(text: string, request: RoleRunRequest, maxBytes: number): AgentSessionHandoff {
	if (!text || byteLength(text) > maxBytes) throw new AgentSessionRuntimeError("AgentSession assistant output is empty or exceeds its bound");
	let candidate: unknown;
	try {
		candidate = JSON.parse(text);
	} catch (error) {
		throw new AgentSessionRuntimeError("AgentSession handoff must be exactly one JSON object", { cause: error });
	}
	if (!isRecord(candidate)) throw new AgentSessionRuntimeError("AgentSession handoff must be an object");
	assertOnlyKeys(candidate, [
		"schemaVersion", "runId", "generation", "laneId", "candidateHead", "validationNonce", "role", "status",
		"summary", "observedMutation", "changedPaths", "verification", "findings",
	], "handoff");
	if (candidate.schemaVersion !== 1) throw new AgentSessionRuntimeError("handoff schemaVersion is invalid");
	for (const [name, actual, expected] of [
		["runId", candidate.runId, request.binding.runId],
		["generation", candidate.generation, request.binding.generation],
		["laneId", candidate.laneId, request.binding.laneId],
		["candidateHead", candidate.candidateHead, request.binding.candidateHead],
		["validationNonce", candidate.validationNonce, request.binding.validationNonce],
		["role", candidate.role, request.role],
	] as const) {
		if (actual !== expected) throw new AgentSessionRuntimeError(`handoff ${name} binding mismatch`);
	}
	if (!["completed", "blocked", "failed"].includes(String(candidate.status))) {
		throw new AgentSessionRuntimeError("handoff status is invalid");
	}
	const summary = redactedBoundedString(candidate.summary, "handoff summary", MAX_HANDOFF_SUMMARY_CHARACTERS, false);
	if (typeof candidate.observedMutation !== "boolean") throw new AgentSessionRuntimeError("handoff observedMutation is invalid");
	if (request.authority.readOnly && candidate.observedMutation) {
		throw new AgentSessionRuntimeError("read-only handoff reported a mutation");
	}
	const changedPaths = boundedArray(candidate.changedPaths, "handoff changedPaths").map((path) => {
		if (typeof path !== "string") throw new AgentSessionRuntimeError("handoff changed path is invalid");
		return validateScopedPath(path, request.authority.writePrefixes);
	});
	if (request.authority.readOnly && changedPaths.length > 0) throw new AgentSessionRuntimeError("read-only handoff contains changed paths");
	const verification = boundedArray(candidate.verification, "handoff verification").map((entry) => {
		if (!isRecord(entry)) throw new AgentSessionRuntimeError("handoff verification entry is invalid");
		assertOnlyKeys(entry, ["name", "status", "summary"], "handoff verification");
		const status = String(entry.status);
		if (!["passed", "failed", "blocked", "not_run"].includes(status)) {
			throw new AgentSessionRuntimeError("handoff verification status is invalid");
		}
		return {
			name: redactedBoundedString(entry.name, "handoff verification name", 128, false),
			status: status as HandoffVerification["status"],
			summary: redactedBoundedString(entry.summary, "handoff verification summary", MAX_HANDOFF_ITEM_CHARACTERS, false),
		};
	});
	const findings = boundedArray(candidate.findings, "handoff findings").map((finding) =>
		redactedBoundedString(finding, "handoff finding", MAX_HANDOFF_ITEM_CHARACTERS, false));
	return {
		schemaVersion: 1,
		runId: request.binding.runId,
		generation: request.binding.generation,
		laneId: request.binding.laneId,
		candidateHead: request.binding.candidateHead,
		validationNonce: request.binding.validationNonce,
		role: request.role,
		status: candidate.status as AgentSessionHandoff["status"],
		summary,
		observedMutation: candidate.observedMutation,
		changedPaths,
		verification,
		findings,
	};
}

function boundedArray(value: unknown, description: string): unknown[] {
	if (!Array.isArray(value) || value.length > MAX_HANDOFF_ARRAY_ITEMS) {
		throw new AgentSessionRuntimeError(`${description} must be a bounded array`);
	}
	return value;
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
	return deadlineAt === undefined ? timeoutDeadline : Math.min(timeoutDeadline, deadlineAt);
}

function isAbsoluteNonTraversingPath(value: unknown): value is string {
	if (typeof value !== "string" || value.length < 1 || value.length > 4_096 || /[\u0000-\u001f\u007f]/.test(value)) return false;
	const flavor = win32.isAbsolute(value) ? win32 : posix;
	if (!flavor.isAbsolute(value)) return false;
	const segments = flavor === win32 ? value.split(/[\\/]+/) : value.split("/");
	return !segments.includes("..");
}

function canonicalWorkspacePath(value: string): string {
	return /^(?:[A-Za-z]:[\\/]|\\\\)/.test(value) ? win32.normalize(value) : posix.normalize(value);
}

function assertOnlyKeys(value: Record<string, unknown>, allowed: readonly string[], description: string): void {
	const allowedSet = new Set(allowed);
	let count = 0;
	for (const key in value) {
		if (!Object.hasOwn(value, key) || !allowedSet.has(key)) {
			throw new AgentSessionRuntimeError(`${description} contains unknown field ${JSON.stringify(key)}`);
		}
		count += 1;
		if (count > allowed.length) throw new AgentSessionRuntimeError(`${description} contains too many fields`);
	}
}

function validIdentifier(value: unknown): value is string {
	return typeof value === "string" && /^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$/.test(value);
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null && !Array.isArray(value);
}

function boundedPositiveInteger(value: number, description: string, maximum: number): number {
	if (!Number.isSafeInteger(value) || value <= 0 || value > maximum) {
		throw new AgentSessionRuntimeError(
			`${description} must be a positive safe integer within the embedded maximum ${maximum}`,
		);
	}
	return value;
}

function frozenArray<T>(values: T[]): T[] {
	Object.freeze(values);
	return values;
}

function byteLength(value: string): number { return new TextEncoder().encode(value).byteLength; }

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
	const timeoutMs = Math.max(0, deadlineAt - Date.now());
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
			if (nodeTypes.isProxy(error)) {
				return new AgentSessionRuntimeError("AgentSession run failed", {
					cause: new Error("proxied failure was omitted"),
				});
			}
		}
		if (error instanceof AgentSessionRuntimeError) {
			const message = safeRuntimeMessage(readErrorMessage(error), "AgentSession operation failed");
			let cause: unknown;
			try {
				cause = Object.hasOwn(error, "cause") ? snapshotRuntimeFailure(error.cause) : undefined;
			} catch {
				cause = new Error("failure cause was unavailable");
			}
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
			cause: new Error("failure normalization was unavailable"),
		});
	}
}

function snapshotRuntimeFailure(
	error: unknown,
	depth = 0,
	seen: WeakSet<object> = new WeakSet<object>(),
): unknown {
	if (error === undefined || error === null || typeof error === "boolean" || typeof error === "number") return error;
	if (typeof error === "string") return safeRuntimeMessage(error, "failure");
	if (typeof error !== "object" && typeof error !== "function") return "unsupported failure";
	const object = error as object;
	try {
		if (nodeTypes.isProxy(object)) return new Error("proxied failure omitted");
	} catch {
		return new Error("failure shape unavailable");
	}
	if (seen.has(object)) return new Error("cyclic failure omitted");
	if (depth >= 4) return new Error("nested failure omitted");
	seen.add(object);
	try {
		if (error instanceof AggregateError) {
			const nested = snapshotAggregateMembers(error, depth, seen);
			let cause: unknown;
			try {
				cause = Object.hasOwn(error, "cause") ? snapshotRuntimeFailure(error.cause, depth + 1, seen) : undefined;
			} catch {
				cause = new Error("aggregate cause was unavailable");
			}
			return new AggregateError(nested, safeRuntimeMessage(readErrorMessage(error), "multiple failures"), { cause });
		}
		if (error instanceof Error) {
			let cause: unknown;
			try {
				cause = Object.hasOwn(error, "cause") ? snapshotRuntimeFailure(error.cause, depth + 1, seen) : undefined;
			} catch {
				cause = new Error("failure cause was unavailable");
			}
			return new Error(safeRuntimeMessage(readErrorMessage(error), "external failure"), { cause });
		}
		return new Error("non-Error failure object");
	} finally {
		seen.delete(object);
	}
}

function snapshotAggregateMembers(error: AggregateError, depth: number, seen: WeakSet<object>): unknown[] {
	const nested: unknown[] = [];
	let iterator: Iterator<unknown> | undefined;
	let exhausted = false;
	try {
		const source = error.errors as Iterable<unknown>;
		const iteratorFactory = source?.[Symbol.iterator];
		if (typeof iteratorFactory !== "function") throw new Error("aggregate members are not iterable");
		iterator = Reflect.apply(iteratorFactory, source, []) as Iterator<unknown>;
		const next = iterator?.next;
		if (typeof next !== "function") throw new Error("aggregate iterator is invalid");
		for (let index = 0; index < 16; index += 1) {
			const step = Reflect.apply(next, iterator, []) as IteratorResult<unknown>;
			if (!step || typeof step !== "object") throw new Error("aggregate iterator step is invalid");
			if (step.done) {
				exhausted = true;
				break;
			}
			nested.push(snapshotRuntimeFailure(step.value, depth + 1, seen));
		}
	} catch {
		nested.push(new Error("aggregate members were unavailable"));
	} finally {
		if (iterator && !exhausted) {
			try {
				const close = iterator.return;
				if (typeof close === "function") Reflect.apply(close, iterator, []);
			} catch {
				if (nested.length < 16) nested.push(new Error("aggregate iterator close failed"));
			}
		}
	}
	return nested;
}

function readErrorMessage(error: unknown): string {
	try {
		return error instanceof Error && typeof error.message === "string" ? error.message : "";
	} catch {
		return "";
	}
}

function safeRuntimeMessage(value: string, fallback: string): string {
	const source = value.length > 0 ? value.slice(0, 4_096) : fallback;
	return redactSensitiveText(source)
		.replace(/[\u0000-\u001f\u007f-\u009f\u061c\u200e\u200f\u2028-\u202e\u2066-\u2069]/g, " ")
		.slice(0, 2_048) || fallback;
}

function combineFailures(failures: readonly unknown[], message: string): unknown {
	if (failures.length === 1) return failures[0];
	return new AggregateError([...failures], message);
}
