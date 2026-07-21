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
	extensionsResult: { extensions: unknown[]; errors: unknown[] };
	modelFallbackMessage?: string;
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

class OwnedSession {
	readonly #session: RuntimeAgentSession;
	readonly #abort: RuntimeAgentSession["abort"];
	readonly #dispose: RuntimeAgentSession["dispose"];
	readonly #prompt: RuntimeAgentSession["prompt"];
	readonly #subscribe: RuntimeAgentSession["subscribe"];
	readonly #waitForIdle: RuntimeAgentSession["waitForIdle"];
	readonly #getActiveToolNames: RuntimeAgentSession["getActiveToolNames"];
	#abortPromise: Promise<void> | undefined;
	#disposePromise: Promise<void> | undefined;
	#unsubscribe: (() => void | PromiseLike<void>) | undefined;
	#unsubscribePromise: Promise<void> | undefined;
	#waitPromise: Promise<void> | undefined;

	constructor(session: RuntimeAgentSession) {
		this.#session = session;
		// Capture the exact acquired operations. Host mutation after ownership transfer must
		// not redirect validation, prompting, cancellation, or cleanup to another surface.
		this.#abort = session.abort;
		this.#dispose = session.dispose;
		this.#prompt = session.prompt;
		this.#subscribe = session.subscribe;
		this.#waitForIdle = session.waitForIdle;
		this.#getActiveToolNames = session.getActiveToolNames;
		if (typeof this.#abort !== "function" || typeof this.#dispose !== "function" ||
			typeof this.#prompt !== "function" || typeof this.#subscribe !== "function" ||
			typeof this.#waitForIdle !== "function" || typeof this.#getActiveToolNames !== "function") {
			throw new AgentSessionRuntimeError("Pi returned an incomplete AgentSession surface");
		}
	}

	activeToolNames(): unknown {
		return Reflect.apply(this.#getActiveToolNames, this.#session, []);
	}

	prompt(value: string, options: { expandPromptTemplates: false; source: "extension" }): Promise<void> {
		return Promise.resolve(Reflect.apply(this.#prompt, this.#session, [value, options]));
	}

	subscribe(listener: (event: AgentSessionEvent) => void): void {
		if (this.#unsubscribe !== undefined || this.#unsubscribePromise !== undefined) {
			throw new AgentSessionRuntimeError("AgentSession subscription ownership was already acquired");
		}
		const unsubscribe = Reflect.apply(this.#subscribe, this.#session, [listener]);
		if (typeof unsubscribe !== "function") {
			throw new AgentSessionRuntimeError("AgentSession subscribe returned an invalid cleanup operation");
		}
		this.#unsubscribe = unsubscribe;
	}

	abortOnce(): Promise<void> {
		if (!this.#abortPromise) {
			this.#abortPromise = Promise.resolve().then(() => Reflect.apply(this.#abort, this.#session, []));
			this.#abortPromise.catch(() => undefined);
		}
		return this.#abortPromise;
	}

	waitOnce(): Promise<void> {
		if (!this.#waitPromise) {
			this.#waitPromise = Promise.resolve().then(() => Reflect.apply(this.#waitForIdle, this.#session, []));
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
			this.#disposePromise = Promise.resolve().then(() =>
				Promise.resolve(Reflect.apply(this.#dispose, this.#session, [])));
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

	abandon(): void {
		if (this.#state !== "pending") return;
		this.#state = "abandoned";
		if (this.#settlement?.status === "fulfilled") this.#startLateCleanup(this.#settlement.value);
		else if (this.#settlement?.status === "rejected") this.#settleTerminal();
	}

	#startLateCleanup(created: RuntimeSessionResult): void {
		if (this.#lateCleanupStarted) return;
		this.#lateCleanupStarted = true;
		void this.#finishLateCreation(created);
	}

	async #finishLateCreation(created: RuntimeSessionResult): Promise<void> {
		let claim: CreatedSessionClaim | undefined;
		const failures: unknown[] = [];
		try {
			claim = this.#captureCreated(created);
		} catch (error) {
			failures.push(error);
		}
		if (claim) {
			try {
				claim.validate();
			} catch (error) {
				failures.push(error);
			}
			try {
				await this.#cleanup(claim.owned);
			} catch (error) {
				failures.push(error);
			}
		}
		if (failures.length > 0) this.#reportFailure(combineFailures(failures, "late AgentSession ownership failed"));
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
			await bounded(owned.unsubscribeOnce(), this.#cleanupTimeoutMs, "abandoned session unsubscribe");
		} catch (error) {
			failures.push(error);
		}
		try {
			await bounded(owned.disposeOnce(), this.#cleanupTimeoutMs, "abandoned session dispose");
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
	readonly #add: AbortSignal["addEventListener"];
	readonly #remove: AbortSignal["removeEventListener"];
	readonly #removeWasOwn: boolean;
	#mayBeAttached = false;
	#released = false;

	constructor(signal: AbortSignal, listener: () => void) {
		this.#signal = signal;
		this.#listener = listener;
		this.#add = signal.addEventListener;
		this.#remove = signal.removeEventListener;
		this.#removeWasOwn = Object.hasOwn(signal, "removeEventListener");
		if (typeof this.#add !== "function" || typeof this.#remove !== "function") {
			throw new AgentSessionRuntimeError("AbortSignal listener operations are invalid");
		}
	}

	attach(): void {
		if (this.#mayBeAttached) return;
		// EventTarget may mutate before a hostile wrapper throws. Own the possible lease first.
		this.#mayBeAttached = true;
		Reflect.apply(this.#add, this.#signal, ["abort", this.#listener, { once: true }]);
	}

	release(): void {
		if (this.#released || !this.#mayBeAttached) return;
		this.#released = true;
		try {
			Reflect.apply(this.#remove, this.#signal, ["abort", this.#listener]);
			if (!this.#removeWasOwn && Object.hasOwn(this.#signal, "removeEventListener")) {
				throw new AgentSessionRuntimeError("AbortSignal listener operation mutated after acquisition");
			}
		} catch (error) {
			let fallbackFailurePresent = false;
			let fallbackFailure: unknown;
			try {
				EventTarget.prototype.removeEventListener.call(this.#signal, "abort", this.#listener);
			} catch (fallbackError) {
				fallbackFailurePresent = true;
				fallbackFailure = fallbackError;
			}
			throw fallbackFailurePresent
				? combineFailures([error, fallbackFailure], "AbortSignal listener detach failed")
				: error;
		}
	}
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
	done: Promise<void>;
	resolveDone(): void;
}

interface TerminalCapture {
	messageEnd?: AssistantTerminal;
	agentEnd?: AssistantTerminal;
	agentEndCount: number;
	agentEndWillRetry: boolean;
	failure?: AgentSessionRuntimeError;
	eventCount: number;
	eventBytes: number;
}

interface AssistantTerminal {
	role: "assistant";
	provider: string;
	model: string;
	stopReason: string;
	timestamp: number;
	content: ReadonlyArray<Readonly<{ type: string; text?: string }>>;
}

export class ShepherdAgentSessionRuntime {
	readonly #sdk: AgentSessionRuntimeSdk;
	readonly #options: Required<Omit<AgentSessionRuntimeOptions, "parentSignal">>;
	readonly #active = new Map<string, ActiveRun>();
	readonly #mutatorLeases = new Map<string, MutationLease>();
	readonly #creations = new Set<SessionCreationOwnership>();
	readonly #parentListenerLease: AbortListenerLease | undefined = undefined;
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
				lease.attach();
			} catch (error) {
				try { lease.release(); } catch { /* Preserve the attachment failure. */ }
				throw error;
			}
			if (parentSignal.aborted) parentAbortListener();
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
		const normalizedRequest = normalizeRunRequest(request);
		this.#assertSdk();
		this.#assertOpen();
		// Capability schemas become bounded immutable Pi tools before any SDK lookup or await.
		const toolPolicy = createToolPolicy({
			readOnly: normalizedRequest.authority.readOnly,
			workspace: normalizedRequest.workspace,
			authority: normalizedRequest.authority,
			capabilities: normalizedRequest.capabilities,
		});
		const route = routeForRole(normalizedRequest.role);
		const model = this.#sdk.findModel(route.provider, route.model);
		if (!isRecord(model) || model.provider !== REQUIRED_PROVIDER || model.id !== REQUIRED_MODEL) {
			throw new AgentSessionRuntimeError(`required model ${REQUIRED_PROVIDER}/${REQUIRED_MODEL} is unavailable; fallback is forbidden`);
		}
		if (!this.#sdk.hasConfiguredAuth(model)) {
			throw new AgentSessionRuntimeError(`required model ${REQUIRED_PROVIDER}/${REQUIRED_MODEL} has no configured auth`);
		}

		const effectiveDeadline = computeDeadline(normalizedRequest.timeoutMs, normalizedRequest.deadlineAt);
		const active = this.#reserve(
			normalizedRequest,
			effectiveDeadline,
			normalizedRequest.deadlineAt !== undefined && effectiveDeadline === normalizedRequest.deadlineAt
				? "AgentSession deadline expired"
				: `AgentSession timed out after ${normalizedRequest.timeoutMs}ms`,
		);
		const scope = active.scope;
		const externalAbort = () => scope.cancel(new AgentSessionRuntimeError("AgentSession run was cancelled by its parent signal"));
		let listenerLease: AbortListenerLease | undefined;
		let result: AgentSessionHandoff | undefined;
		let primaryFailurePresent = false;
		let primaryFailure: unknown;
		try {
			if (normalizedRequest.signal) {
				listenerLease = new AbortListenerLease(normalizedRequest.signal, externalAbort);
				listenerLease?.attach();
				if (normalizedRequest.signal.aborted) externalAbort();
			}
			result = await this.#execute(normalizedRequest, route.thinking, model, active, toolPolicy);
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
			const matches = [...this.#active.values()].filter((active) => active.runId === runId);
			for (const active of matches) {
				active.scope.cancel(new AgentSessionRuntimeError(`AgentSession run ${runId} was aborted`));
				void active.owned?.abortOnce().catch(() => undefined);
			}
			await Promise.all(matches.map((active) => active.done));
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
		for (const active of this.#active.values()) {
			active.scope.cancel(new AgentSessionRuntimeError(reason));
			void active.owned?.abortOnce().catch(() => undefined);
		}
		await Promise.all([...this.#active.values()].map((active) => active.done));
		const creationOwners = [...this.#creations];
		if (creationOwners.length > 0) {
			const terminal = Promise.all(creationOwners.map((creation) => creation.terminal)).then(() => undefined);
			const settlement = await settleWithin(terminal, this.#options.cleanupTimeoutMs);
			if (settlement.status === "pending") {
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
			this.#creations.add(owner);
			void owner.terminal.then(() => { this.#creations.delete(owner); });
			const created = await scope.race(createPromise);
			const claim = creation.claim(created);
			owned = claim.owned;
			active.owned = owned;
			scope.setOnCancel(() => { void owned?.abortOnce().catch(() => undefined); });
			scope.assertActive();
			try {
				claim.validate();
			} catch (error) {
				this.#setQuarantine(error);
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
				if (settlement.status === "pending" || settlement.status === "rejected") {
					creation.abandon();
				} else {
					const claim = creation.claim(settlement.value);
					owned = claim.owned;
					active.owned = owned;
					try {
						claim.validate();
					} catch (error) {
						this.#setQuarantine(error);
						if (!cleanupFailurePresent) {
							cleanupFailurePresent = true;
							cleanupFailure = error;
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
				await cleanupOwnedSession(owned, scope.failure !== undefined, this.#options.cleanupTimeoutMs);
			} catch (error) {
				if (!cleanupFailurePresent) {
					cleanupFailurePresent = true;
					cleanupFailure = error;
				}
			}
		}
		if (cleanupFailurePresent) {
			this.#setQuarantine(cleanupFailure);
		}

		if (cleanupFailurePresent) {
			throw new AgentSessionRuntimeError("AgentSession cleanup/join failed; runtime quarantined", {
				cause: primaryFailurePresent
					? combineFailures([primaryFailure, cleanupFailure], "AgentSession primary and cleanup failures")
					: cleanupFailure,
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

	#setQuarantine(error: unknown): void {
		if (this.#quarantineFailurePresent) return;
		this.#quarantineFailurePresent = true;
		this.#quarantineFailure = error;
	}
}

function normalizeRunRequest(request: RoleRunRequest): RoleRunRequest {
	if (!isRecord(request)) throw new AgentSessionRuntimeError("AgentSession request must be an object");
	assertOnlyKeys(request, [
		"role", "task", "context", "timeoutMs", "deadlineAt", "signal", "workspace", "capabilities", "authority", "binding",
	], "request");
	// Read every caller-owned top-level field exactly once, then use only the frozen snapshot.
	const role = request.role;
	const task = request.task;
	const contextSource = request.context;
	const timeoutMs = request.timeoutMs;
	const deadlineAt = request.deadlineAt;
	const signal = request.signal;
	const workspaceSource = request.workspace;
	const capabilitiesSource = request.capabilities;
	const authoritySource = request.authority;
	const bindingSource = request.binding;

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

	if (!isRecord(authoritySource)) throw new AgentSessionRuntimeError("request authority is required");
	assertOnlyKeys(authoritySource, [
		"issue", "branch", "workspaceId", "readOnly", "readPrefixes", "writePrefixes", "capabilityNames",
	], "authority");
	const issue = authoritySource.issue;
	const branch = authoritySource.branch;
	const authorityWorkspaceId = authoritySource.workspaceId;
	const readOnly = authoritySource.readOnly;
	const readPrefixesSource = authoritySource.readPrefixes;
	const writePrefixesSource = authoritySource.writePrefixes;
	const capabilityNamesSource = authoritySource.capabilityNames;
	if (!Number.isSafeInteger(issue) || issue < 1) throw new AgentSessionRuntimeError("authority issue is invalid");
	if (typeof branch !== "string" || branch.length < 1 || branch.length > 255 ||
		/[\u0000-\u001f\u007f]/.test(branch) || branch === "main") {
		throw new AgentSessionRuntimeError("authority branch is invalid or targets main");
	}
	if (!validIdentifier(authorityWorkspaceId)) throw new AgentSessionRuntimeError("authority workspace identity is invalid");
	if (typeof readOnly !== "boolean") throw new AgentSessionRuntimeError("authority readOnly is invalid");
	const readPrefixes = frozenArray(normalizeScopedPrefixes(readPrefixesSource, "read"));
	const writePrefixes = readOnly && Array.isArray(writePrefixesSource) && writePrefixesSource.length === 0
		? frozenArray<string>([])
		: frozenArray(normalizeScopedPrefixes(writePrefixesSource, "write"));
	if (!Array.isArray(capabilityNamesSource) || capabilityNamesSource.length > 32) {
		throw new AgentSessionRuntimeError("authority capability names must be a bounded array");
	}
	const capabilityNames = frozenArray(capabilityNamesSource.map((name) => {
		if (typeof name !== "string") throw new AgentSessionRuntimeError("authority capability name is invalid");
		return name;
	}));
	const authority = Object.freeze({
		issue: Number(issue),
		branch,
		workspaceId: authorityWorkspaceId,
		readOnly,
		readPrefixes,
		writePrefixes,
		capabilityNames,
	}) as RoleAuthority;

	if (!isRecord(bindingSource)) throw new AgentSessionRuntimeError("request binding is required");
	assertOnlyKeys(bindingSource, ["runId", "generation", "laneId", "candidateHead", "validationNonce"], "binding");
	const runId = bindingSource.runId;
	const generation = bindingSource.generation;
	const laneId = bindingSource.laneId;
	const candidateHead = bindingSource.candidateHead;
	const validationNonce = bindingSource.validationNonce;
	if (!validIdentifier(runId) || !validIdentifier(laneId) ||
		!Number.isSafeInteger(generation) || generation < 1 ||
		typeof candidateHead !== "string" || !/^[0-9a-f]{40}$/.test(candidateHead) ||
		!validIdentifier(validationNonce) || validationNonce.length < 12) {
		throw new AgentSessionRuntimeError("request binding is invalid");
	}
	const binding = Object.freeze({ runId, generation: Number(generation), laneId, candidateHead, validationNonce });

	if (!workspaceSource || typeof workspaceSource !== "object") {
		throw new AgentSessionRuntimeError("workspace capability is required");
	}
	const workspaceId = workspaceSource.id;
	const workspaceCwd = workspaceSource.cwd;
	const readText = workspaceSource.readText;
	const editText = workspaceSource.editText;
	const writeText = workspaceSource.writeText;
	if (workspaceId !== authorityWorkspaceId || !isAbsoluteNonTraversingPath(workspaceCwd) ||
		typeof readText !== "function" || typeof editText !== "function" || typeof writeText !== "function") {
		throw new AgentSessionRuntimeError("workspace identity, cwd, or capability does not match the immutable authority envelope");
	}
	const canonicalCwd = canonicalWorkspacePath(workspaceCwd);
	const workspace = Object.freeze({
		id: workspaceId,
		cwd: canonicalCwd,
		readText(path: string, options: { offset?: number; limit?: number; signal?: AbortSignal }) {
			return Reflect.apply(readText, workspaceSource, [path, options]);
		},
		editText(path: string, oldText: string, newText: string, operationSignal?: AbortSignal) {
			return Reflect.apply(editText, workspaceSource, [path, oldText, newText, operationSignal]);
		},
		writeText(path: string, content: string, operationSignal?: AbortSignal) {
			return Reflect.apply(writeText, workspaceSource, [path, content, operationSignal]);
		},
	}) satisfies ScopedWorkspace;

	if (!Array.isArray(capabilitiesSource) || capabilitiesSource.length > 32) {
		throw new AgentSessionRuntimeError("typed host capabilities must be a bounded array");
	}
	const capabilities = frozenArray(capabilitiesSource.map((capability) => normalizeCapability(capability)));
	if (!Array.isArray(contextSource)) throw new AgentSessionRuntimeError("role context must be a bounded array");
	const context = frozenArray(contextSource.map((item) => item));
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
	if (!capability || typeof capability !== "object") throw new AgentSessionRuntimeError("capability must be an object");
	const name = capability.name;
	const description = capability.description;
	const mutates = capability.mutates;
	const parameters = capability.parameters;
	const execute = capability.execute;
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
	if (!created || typeof created !== "object") {
		throw new AgentSessionRuntimeError("Pi returned an invalid AgentSession result");
	}
	// This is the sole read of the SDK-owned result's session field. Every later operation is
	// invoked through methods captured by OwnedSession from this exact object.
	const session = created.session;
	if (!session || typeof session !== "object") {
		throw new AgentSessionRuntimeError("Pi returned an invalid AgentSession result without a cleanable session");
	}
	const owned = new OwnedSession(session);
	let captureFailurePresent = false;
	let captureFailure: unknown;
	let extensionShapeValid = false;
	let extensionCount = 0;
	let extensionErrorCount = 0;
	let fallbackMessage: unknown;
	let modelProvider: unknown;
	let modelId: unknown;
	let thinkingLevel: unknown;
	let sessionFile: unknown;
	let activeTools: readonly string[] | undefined;
	try {
		const extensionsResult = created.extensionsResult;
		fallbackMessage = created.modelFallbackMessage;
		if (extensionsResult && typeof extensionsResult === "object") {
			const extensions = extensionsResult.extensions;
			const errors = extensionsResult.errors;
			extensionShapeValid = Array.isArray(extensions) && Array.isArray(errors);
			if (extensionShapeValid) {
				extensionCount = extensions.length;
				extensionErrorCount = errors.length;
			}
		}
		const model = session.model;
		modelProvider = model?.provider;
		modelId = model?.id;
		thinkingLevel = session.thinkingLevel;
		sessionFile = session.sessionFile;
		activeTools = captureToolNameArray(owned.activeToolNames());
	} catch (error) {
		captureFailurePresent = true;
		captureFailure = error;
	}

	return Object.freeze({
		owned,
		validate(): void {
			if (captureFailurePresent) throw captureFailure;
			if (!extensionShapeValid) throw new AgentSessionRuntimeError("Pi returned an invalid AgentSession result");
			if (fallbackMessage !== undefined) throw new AgentSessionRuntimeError("Pi attempted a forbidden model fallback");
			if (extensionCount > 0 || extensionErrorCount > 0) {
				throw new AgentSessionRuntimeError("embedded AgentSession unexpectedly loaded extensions or extension errors");
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

function captureToolNameArray(value: unknown): readonly string[] | undefined {
	if (!Array.isArray(value) || nodeTypes.isProxy(value) || value.length > 256) return undefined;
	if (Reflect.ownKeys(value).length !== value.length + 1) return undefined;
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
	return { agentEndCount: 0, agentEndWillRetry: false, eventCount: 0, eventBytes: 0 };
}

function captureEvent(
	capture: TerminalCapture,
	event: unknown,
	options: Required<Omit<AgentSessionRuntimeOptions, "parentSignal">>,
): void {
	if (capture.failure) return;
	capture.eventCount += 1;
	if (capture.eventCount > options.maxEvents) {
		capture.failure = new AgentSessionRuntimeError("AgentSession event stream exceeded its bound");
		return;
	}
	try {
		capture.eventBytes += boundedEventBytes(event, options.maxEventBytes - capture.eventBytes);
	} catch (error) {
		capture.failure = new AgentSessionRuntimeError("AgentSession emitted an event that exceeded bounded safe accounting", {
			cause: error,
		});
		return;
	}
	if (capture.eventBytes > options.maxEventBytes) {
		capture.failure = new AgentSessionRuntimeError("AgentSession event stream exceeded its bound");
		return;
	}
	try {
		const eventFields = captureClosedRecordFields(event, "AgentSession event", MAX_EVENT_RECORD_KEYS);
		const eventType = eventFields.get("type");
		if (typeof eventType !== "string") return;
		if (eventType === "message_end") {
			capture.messageEnd = captureAssistantTerminal(eventFields.get("message"));
		}
		if (eventType === "agent_end") {
			capture.agentEndCount += 1;
			capture.agentEndWillRetry ||= eventFields.get("willRetry") === true;
			const messages = captureDenseArray(eventFields.get("messages"), "AgentSession terminal messages");
			for (const message of messages) {
				const terminal = captureAssistantTerminal(message);
				if (terminal) capture.agentEnd = terminal;
			}
		}
	} catch (error) {
		capture.failure = new AgentSessionRuntimeError("AgentSession emitted an invalid terminal event", { cause: error });
	}
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
		const keys = Reflect.ownKeys(object);
		if (array) {
			if (value.length > MAX_EVENT_ARRAY_ITEMS || keys.length !== value.length + 1) {
				throw new AgentSessionRuntimeError("AgentSession event contains a sparse or oversized array");
			}
			for (let index = 0; index < value.length; index += 1) {
				if (!Object.hasOwn(value, String(index))) {
					throw new AgentSessionRuntimeError("AgentSession event contains a sparse array");
				}
			}
		} else if (keys.length > MAX_EVENT_RECORD_KEYS) {
			throw new AgentSessionRuntimeError("AgentSession event record is too wide");
		}
		add(2);
		for (const key of keys) {
			if (typeof key !== "string") throw new AgentSessionRuntimeError("AgentSession event contains a symbol key");
			const descriptor = Reflect.getOwnPropertyDescriptor(object, key);
			if (array && key === "length" && descriptor && !descriptor.enumerable && "value" in descriptor) continue;
			if (!descriptor?.enumerable) {
				throw new AgentSessionRuntimeError("AgentSession event contains a non-enumerable field");
			}
			if (descriptor.get || descriptor.set || !("value" in descriptor)) {
				throw new AgentSessionRuntimeError("AgentSession event contains an accessor");
			}
			add(byteLength(key) + 3);
			stack.push({ value: descriptor.value, depth: current.depth + 1 });
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
	const keys = Reflect.ownKeys(value);
	if (keys.length > maximumKeys) throw new AgentSessionRuntimeError(`${description} contains too many fields`);
	const fields = new Map<string, unknown>();
	for (const key of keys) {
		if (typeof key !== "string") throw new AgentSessionRuntimeError(`${description} contains a symbol field`);
		const descriptor = Reflect.getOwnPropertyDescriptor(value, key);
		if (!descriptor?.enumerable) throw new AgentSessionRuntimeError(`${description} contains a non-enumerable field`);
		if (descriptor.get || descriptor.set || !("value" in descriptor)) {
			throw new AgentSessionRuntimeError(`${description} contains an accessor field`);
		}
		fields.set(key, descriptor.value);
	}
	return fields;
}

function captureDenseArray(value: unknown, description: string): readonly unknown[] {
	if (!Array.isArray(value) || nodeTypes.isProxy(value) || value.length > MAX_EVENT_ARRAY_ITEMS) {
		throw new AgentSessionRuntimeError(`${description} must be a bounded dense non-proxy array`);
	}
	const keys = Reflect.ownKeys(value);
	if (keys.length !== value.length + 1) {
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

function captureAssistantTerminal(value: unknown): AssistantTerminal | undefined {
	const fields = captureClosedRecordFields(value, "AgentSession terminal message", 32);
	const role = fields.get("role");
	if (role !== "assistant") return undefined;
	const provider = fields.get("provider");
	const model = fields.get("model");
	const stopReason = fields.get("stopReason");
	const timestamp = fields.get("timestamp");
	if (typeof provider !== "string" || typeof model !== "string" || typeof stopReason !== "string" ||
		typeof timestamp !== "number" || !Number.isFinite(timestamp)) {
		throw new AgentSessionRuntimeError("AgentSession assistant terminal contains invalid routing fields");
	}
	const content = captureDenseArray(fields.get("content"), "AgentSession assistant terminal content").map((part) => {
		const partFields = captureClosedRecordFields(part, "AgentSession assistant content part", 16);
		const type = partFields.get("type");
		const text = partFields.get("text");
		if (typeof type !== "string" || (type === "text" && typeof text !== "string")) {
			throw new AgentSessionRuntimeError("AgentSession assistant content part is invalid");
		}
		return Object.freeze(type === "text" ? { type, text: text as string } : { type });
	});
	return Object.freeze({
		role: "assistant",
		provider,
		model,
		stopReason,
		timestamp,
		content: Object.freeze(content),
	});
}

function verifyTerminalCapture(capture: TerminalCapture): AssistantTerminal {
	if (capture.agentEndCount !== 1 || capture.agentEndWillRetry || !capture.messageEnd || !capture.agentEnd ||
		!sameTerminal(capture.messageEnd, capture.agentEnd)) {
		throw new AgentSessionRuntimeError("AgentSession returned an invalid terminal event sequence");
	}
	if (capture.agentEnd.stopReason !== "stop") {
		throw new AgentSessionRuntimeError(`AgentSession terminal stop reason ${capture.agentEnd.stopReason} is not accepted`);
	}
	return capture.agentEnd;
}

function sameTerminal(left: AssistantTerminal, right: AssistantTerminal): boolean {
	return left.provider === right.provider && left.model === right.model && left.stopReason === right.stopReason &&
		left.timestamp === right.timestamp && assistantText(left) === assistantText(right);
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
	const redacted = redactSensitiveText(value);
	const terminalControls = /[\u0000-\u001f\u007f-\u009f\u061c\u200e\u200f\u2028-\u202e\u2066-\u2069]/;
	if (!terminalControls.test(redacted)) return redacted;
	// Retained multiline structured diagnostics are allowed only when the redactor actually
	// consumed credential material; remove every remaining terminal control before delivery.
	// An unchanged string containing any such control remains a hard validation failure.
	if (redacted === value) {
		throw new AgentSessionRuntimeError(`${description} contains a terminal control character`);
	}
	return redacted.replace(/[\u0000-\u001f\u007f-\u009f\u061c\u200e\u200f\u2028-\u202e\u2066-\u2069]/g, " ");
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
	for (const key of Object.keys(value)) {
		if (!allowedSet.has(key)) throw new AgentSessionRuntimeError(`${description} contains unknown field ${JSON.stringify(key)}`);
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

async function bounded<T>(operation: Promise<T>, timeoutMs: number, description: string): Promise<T> {
	let timer: ReturnType<typeof setTimeout> | undefined;
	const timeout = new Promise<never>((_resolve, reject) => {
		timer = setTimeout(() => reject(new AgentSessionRuntimeError(`${description} timed out after ${timeoutMs}ms`)), timeoutMs);
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
	return error instanceof AgentSessionRuntimeError
		? error
		: new AgentSessionRuntimeError(
			`AgentSession run failed${error instanceof Error && error.message ? `: ${error.message}` : ""}`,
			{ cause: error },
		);
}

function combineFailures(failures: readonly unknown[], message: string): unknown {
	if (failures.length === 1) return failures[0];
	return new AggregateError([...failures], message);
}
