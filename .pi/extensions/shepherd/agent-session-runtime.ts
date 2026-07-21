import { posix, win32 } from "node:path";

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
	redactSensitiveText,
	validateScopedPath,
	type HostCapability,
	type ScopedWorkspace,
	type SessionTool,
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
const MAX_HANDOFF_SUMMARY_CHARACTERS = 4 * 1024;
const MAX_HANDOFF_ARRAY_ITEMS = 32;
const MAX_HANDOFF_ITEM_CHARACTERS = 2 * 1024;

interface RuntimeResourceLoader {
	reload(): Promise<void>;
}

interface RuntimeModel {
	provider: string;
	id: string;
}

export interface RuntimeAgentSession {
	model: RuntimeModel | undefined;
	thinkingLevel: ShepherdAgentThinking | string;
	sessionFile: string | undefined;
	getActiveToolNames(): string[];
	subscribe(listener: (event: AgentSessionEvent) => void): () => void;
	prompt(prompt: string, options: { expandPromptTemplates: false; source: "extension" }): Promise<void>;
	abort(): Promise<void>;
	waitForIdle(): Promise<void>;
	dispose(): void;
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
	findModel(provider: string, model: string): unknown;
	hasConfiguredAuth(model: unknown): boolean;
	createSettingsManager(settings: Record<string, unknown>, options: Record<string, unknown>): unknown;
	createSessionManager(cwd: string): unknown;
	createResourceLoader(options: Record<string, unknown>): RuntimeResourceLoader;
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
		super(message, options);
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
	readonly session: RuntimeAgentSession;
	unsubscribe: (() => void) | undefined;
	#abortPromise: Promise<void> | undefined;
	#disposePromise: Promise<void> | undefined;
	#waitPromise: Promise<void> | undefined;

	constructor(session: RuntimeAgentSession) {
		this.session = session;
	}

	abortOnce(): Promise<void> {
		if (!this.#abortPromise) this.#abortPromise = Promise.resolve().then(() => this.session.abort());
		return this.#abortPromise;
	}

	waitOnce(): Promise<void> {
		if (!this.#waitPromise) {
			this.#waitPromise = Promise.resolve().then(() => this.session.waitForIdle());
			this.#waitPromise.catch(() => undefined);
		}
		return this.#waitPromise;
	}

	disposeOnce(): Promise<void> {
		if (!this.#disposePromise) {
			this.#disposePromise = Promise.resolve().then(() => {
				let firstError: unknown;
				try {
					this.unsubscribe?.();
				} catch (error) {
					firstError = error;
				} finally {
					this.unsubscribe = undefined;
				}
				try {
					this.session.dispose();
				} catch (error) {
					firstError ??= error;
				}
				if (firstError !== undefined) throw firstError;
			});
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
	let firstError: unknown;
	if (abort) {
		try {
			await bounded(owned.abortOnce(), timeoutMs, "session abort");
		} catch (error) {
			firstError = error;
		}
	}
	try {
		await bounded(owned.waitOnce(), timeoutMs, "session idle wait");
	} catch (error) {
		firstError ??= error;
	}
	// Disposal is synchronous at the Pi port and must remain reachable after either bounded phase.
	try {
		await owned.disposeOnce();
	} catch (error) {
		firstError ??= error;
	}
	if (firstError !== undefined) throw firstError;
}

class SessionCreationOwnership {
	readonly promise: Promise<RuntimeSessionResult>;
	readonly terminal: Promise<void>;
	readonly #cleanupTimeoutMs: number;
	readonly #onCleanupFailure: (error: unknown) => void;
	readonly #validateCreated: (created: RuntimeSessionResult) => void;
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
		validateCreated: (created: RuntimeSessionResult) => void,
		onCleanupFailure: (error: unknown) => void,
	) {
		this.promise = promise;
		this.#cleanupTimeoutMs = cleanupTimeoutMs;
		this.#validateCreated = validateCreated;
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

	claim(created: RuntimeSessionResult): OwnedSession {
		if (this.#state !== "pending") {
			throw new AgentSessionRuntimeError("AgentSession creation ownership was already settled");
		}
		let owned: OwnedSession;
		try {
			owned = ownedSessionFromResult(created);
		} catch (error) {
			this.#state = "failed";
			this.#reportFailure(error);
			this.#settleTerminal();
			throw error;
		}
		this.#state = "claimed";
		this.#settleTerminal();
		return owned;
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
		let owned: OwnedSession | undefined;
		let firstError: unknown;
		try {
			owned = ownedSessionFromResult(created);
		} catch (error) {
			firstError = error;
		}
		if (owned) {
			try {
				this.#validateCreated(created);
			} catch (error) {
				firstError ??= error;
			}
			try {
				await this.#cleanup(owned);
			} catch (error) {
				firstError ??= error;
			}
		}
		if (firstError !== undefined) this.#reportFailure(firstError);
		this.#settleTerminal();
	}

	async #cleanup(owned: OwnedSession): Promise<void> {
		const deadlineAt = Date.now() + this.#cleanupTimeoutMs;
		let firstError: unknown;
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
			if (settlement.status === "rejected") firstError ??= settlement.reason;
		}
		try {
			await owned.disposeOnce();
		} catch (error) {
			firstError ??= error;
		}
		if (firstError !== undefined) throw firstError;
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

interface ActiveRun {
	key: string;
	runId: string;
	readOnly: boolean;
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
	content: Array<{ type: string; text?: string }>;
}

export class ShepherdAgentSessionRuntime {
	readonly #sdk: AgentSessionRuntimeSdk;
	readonly #options: Required<Omit<AgentSessionRuntimeOptions, "parentSignal">>;
	readonly #active = new Map<string, ActiveRun>();
	readonly #creations = new Set<SessionCreationOwnership>();
	readonly #parentSignal: AbortSignal | undefined;
	readonly #parentAbortListener: (() => void) | undefined;
	#activeMutator = false;
	#closing = false;
	#closed = false;
	#closePromise: Promise<void> | undefined;
	#quarantineFailure: unknown;

	constructor(sdk: AgentSessionRuntimeSdk, options: AgentSessionRuntimeOptions = {}) {
		this.#sdk = sdk;
		this.#options = {
			maxConcurrency: positiveInteger(options.maxConcurrency ?? DEFAULT_MAX_CONCURRENCY, "maxConcurrency"),
			maxEvents: positiveInteger(options.maxEvents ?? DEFAULT_MAX_EVENTS, "maxEvents"),
			maxEventBytes: positiveInteger(options.maxEventBytes ?? DEFAULT_MAX_EVENT_BYTES, "maxEventBytes"),
			maxAssistantBytes: positiveInteger(options.maxAssistantBytes ?? DEFAULT_MAX_ASSISTANT_BYTES, "maxAssistantBytes"),
			cleanupTimeoutMs: positiveInteger(options.cleanupTimeoutMs ?? DEFAULT_CLEANUP_TIMEOUT_MS, "cleanupTimeoutMs"),
		};
		if (this.#options.maxConcurrency > 32) throw new AgentSessionRuntimeError("maxConcurrency exceeds the embedded runtime bound");
		this.#parentSignal = options.parentSignal;
		if (options.parentSignal) {
			this.#parentAbortListener = () => { void this.#close("parent shutdown requested").catch(() => undefined); };
			options.parentSignal.addEventListener("abort", this.#parentAbortListener, { once: true });
			if (options.parentSignal.aborted) this.#parentAbortListener();
		}
	}

	async run(request: RoleRunRequest): Promise<AgentSessionHandoff> {
		validateRunRequest(request);
		this.#assertSdk();
		this.#assertOpen();
		const route = routeForRole(request.role);
		const model = this.#sdk.findModel(route.provider, route.model);
		if (!isRecord(model) || model.provider !== REQUIRED_PROVIDER || model.id !== REQUIRED_MODEL) {
			throw new AgentSessionRuntimeError(`required model ${REQUIRED_PROVIDER}/${REQUIRED_MODEL} is unavailable; fallback is forbidden`);
		}
		if (!this.#sdk.hasConfiguredAuth(model)) {
			throw new AgentSessionRuntimeError(`required model ${REQUIRED_PROVIDER}/${REQUIRED_MODEL} has no configured auth`);
		}

		const effectiveDeadline = computeDeadline(request.timeoutMs, request.deadlineAt);
		const active = this.#reserve(
			request,
			effectiveDeadline,
			request.deadlineAt !== undefined && effectiveDeadline === request.deadlineAt
				? "AgentSession deadline expired"
				: `AgentSession timed out after ${request.timeoutMs}ms`,
		);
		const scope = active.scope;
		const externalAbort = () => scope.cancel(new AgentSessionRuntimeError("AgentSession run was cancelled by its parent signal"));
		let listenerAttached = false;
		let listenerFailure: unknown;
		try {
			if (request.signal) {
				request.signal.addEventListener("abort", externalAbort, { once: true });
				listenerAttached = true;
				if (request.signal.aborted) externalAbort();
			}
			return await this.#execute(request, route.thinking, model, active);
		} finally {
			try {
				if (listenerAttached) request.signal?.removeEventListener("abort", externalAbort);
			} catch (error) {
				listenerFailure = error;
			} finally {
				try {
					scope.finish();
				} finally {
					this.#release(active);
				}
			}
			if (listenerFailure !== undefined) throw listenerFailure;
		}
	}

	async abort(runId: string): Promise<void> {
		if (!validIdentifier(runId)) throw new AgentSessionRuntimeError("abort runId is invalid");
		const matches = [...this.#active.values()].filter((active) => active.runId === runId);
		for (const active of matches) {
			active.scope.cancel(new AgentSessionRuntimeError(`AgentSession run ${runId} was aborted`));
			void active.owned?.abortOnce().catch(() => undefined);
		}
		await Promise.all(matches.map((active) => active.done));
	}

	close(): Promise<void> { return this.#close("AgentSession runtime closed"); }
	shutdown(): Promise<void> { return this.#close("AgentSession parent shutdown requested"); }

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
			if (!settlement.settled) {
				this.#quarantineFailure ??= new AgentSessionRuntimeError(
					"AgentSession creation remained pending during bounded close",
				);
			}
		}
		this.#closed = true;
		if (this.#parentSignal && this.#parentAbortListener) {
			this.#parentSignal.removeEventListener("abort", this.#parentAbortListener);
		}
		if (this.#quarantineFailure !== undefined) {
			throw new AgentSessionRuntimeError("AgentSession runtime closed while quarantined after cleanup failure", {
				cause: this.#quarantineFailure,
			});
		}
	}

	async #execute(
		request: RoleRunRequest,
		thinking: ShepherdAgentThinking,
		model: Record<string, unknown>,
		active: ActiveRun,
	): Promise<AgentSessionHandoff> {
		const scope = active.scope;
		const toolPolicy = createToolPolicy({
			readOnly: request.authority.readOnly,
			workspace: request.workspace,
			authority: request.authority,
			capabilities: request.capabilities,
		});
		const prompts = buildRolePrompts({
			role: request.role,
			task: request.task,
			context: request.context,
			authority: {
				issue: request.authority.issue,
				branch: request.authority.branch,
				workspaceId: request.authority.workspaceId,
				readOnly: request.authority.readOnly,
				toolNames: toolPolicy.names,
				binding: request.binding,
			},
		});

		let creation: SessionCreationOwnership | undefined;
		let reloadPromise: Promise<void> | undefined;
		let owned: OwnedSession | undefined;
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

			const createPromise = Promise.resolve().then(() => this.#sdk.createAgentSession({
				cwd: request.workspace.cwd,
				agentDir: this.#sdk.getAgentDir(),
				model,
				thinkingLevel: thinking,
				scopedModels: [{ model, thinkingLevel: thinking }],
				noTools: "all",
				tools: toolPolicy.names,
				excludeTools: ["bash"],
				customTools: toolPolicy.tools as unknown as CreateAgentSessionOptions["customTools"],
				resourceLoader,
				sessionManager,
				settingsManager,
			} as unknown as CreateAgentSessionOptions));
			const owner = new SessionCreationOwnership(
				createPromise,
				this.#options.cleanupTimeoutMs,
				(created) => validateCreatedSession(created, thinking, toolPolicy.names),
				(error) => {
					this.#quarantineFailure ??= error;
				},
			);
			creation = owner;
			this.#creations.add(owner);
			void owner.terminal.then(() => { this.#creations.delete(owner); });
			const created = await scope.race(createPromise);
			owned = creation.claim(created);
			active.owned = owned;
			scope.setOnCancel(() => { void owned?.abortOnce().catch(() => undefined); });
			scope.assertActive();
			validateCreatedSession(created, thinking, toolPolicy.names);

			const capture = newTerminalCapture();
			owned.unsubscribe = created.session.subscribe((event) => {
				captureEvent(capture, event, this.#options);
				if (capture.failure) void owned?.abortOnce().catch(() => undefined);
			});
			await scope.race(created.session.prompt(prompts.userPrompt, {
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
			primaryFailure = error;
		}

		let cleanupFailure: unknown;
		if (reloadPromise) {
			try {
				await bounded(reloadPromise, this.#options.cleanupTimeoutMs, "resource loader settlement");
			} catch (error) {
				cleanupFailure ??= error;
			}
		}
		if (!owned && creation?.pending) {
			try {
				const settlement = await settleWithin(creation.promise, this.#options.cleanupTimeoutMs);
				if (!settlement.settled) {
					creation.abandon();
				} else {
					owned = creation.claim(settlement.value);
					active.owned = owned;
					try {
						validateCreatedSession(settlement.value, thinking, toolPolicy.names);
					} catch (error) {
						cleanupFailure ??= error;
					}
				}
			} catch (error) {
				creation.abandon();
				cleanupFailure ??= error;
			}
		}
		if (owned) {
			try {
				await cleanupOwnedSession(owned, scope.failure !== undefined, this.#options.cleanupTimeoutMs);
			} catch (error) {
				cleanupFailure ??= error;
			}
		}
		if (cleanupFailure !== undefined) {
			this.#quarantineFailure ??= cleanupFailure;
		}

		if (cleanupFailure !== undefined) {
			throw new AgentSessionRuntimeError("AgentSession cleanup/join failed; runtime quarantined", { cause: cleanupFailure });
		}
		if (primaryFailure !== undefined) throw normalizeRuntimeError(primaryFailure);
		// Cancellation may win after terminal evidence is parsed but before child settlement completes.
		// Never return otherwise-valid late evidence after close, shutdown, abort, or deadline.
		if (scope.failure) throw scope.failure;
		if (!result) throw new AgentSessionRuntimeError("AgentSession completed without a handoff");
		return result;
	}

	#reserve(request: RoleRunRequest, deadlineAt: number, timeoutDescription: string): ActiveRun {
		const key = `${request.binding.runId}:${request.binding.generation}:${request.binding.laneId}`;
		if (this.#active.has(key)) throw new AgentSessionRuntimeError("run/lane/generation is already active");
		if (this.#active.size >= this.#options.maxConcurrency) {
			throw new AgentSessionRuntimeError(`AgentSession concurrency limit ${this.#options.maxConcurrency} reached`);
		}
		if (!request.authority.readOnly && this.#activeMutator) {
			throw new AgentSessionRuntimeError("only one mutating AgentSession may run at a time");
		}
		const scope = new CancellationScope(deadlineAt, timeoutDescription);
		const completion = deferred();
		const active: ActiveRun = {
			key,
			runId: request.binding.runId,
			readOnly: request.authority.readOnly,
			scope,
			done: completion.promise,
			resolveDone: completion.resolve,
		};
		this.#active.set(key, active);
		this.#activeMutator ||= !request.authority.readOnly;
		return active;
	}

	#release(active: ActiveRun): void {
		if (!this.#active.delete(active.key)) return;
		if (!active.readOnly) this.#activeMutator = false;
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
		if (this.#quarantineFailure !== undefined) {
			throw new AgentSessionRuntimeError("AgentSession runtime is quarantined after failed cleanup", {
				cause: this.#quarantineFailure,
			});
		}
		if (this.#closing || this.#closed) throw new AgentSessionRuntimeError("AgentSession runtime is closed");
	}
}

function validateRunRequest(request: RoleRunRequest): void {
	if (!isRecord(request)) throw new AgentSessionRuntimeError("AgentSession request must be an object");
	assertOnlyKeys(request, [
		"role", "task", "context", "timeoutMs", "deadlineAt", "signal", "workspace", "capabilities", "authority", "binding",
	], "request");
	routeForRole(request.role);
	if (!Number.isSafeInteger(request.timeoutMs) || request.timeoutMs <= 0 || request.timeoutMs > MAX_TIMEOUT_MS) {
		throw new AgentSessionRuntimeError("timeoutMs must be a positive bounded safe integer");
	}
	if (request.deadlineAt !== undefined && (!Number.isSafeInteger(request.deadlineAt) || request.deadlineAt <= Date.now())) {
		throw new AgentSessionRuntimeError("deadlineAt must be a future epoch-millisecond safe integer");
	}
	if (request.signal !== undefined && !(request.signal instanceof AbortSignal)) {
		throw new AgentSessionRuntimeError("request signal is invalid");
	}
	if (!request.workspace || request.workspace.id !== request.authority?.workspaceId || !isAbsoluteNonTraversingPath(request.workspace.cwd)) {
		throw new AgentSessionRuntimeError("workspace identity or cwd does not match the immutable authority envelope");
	}
	if (!isRecord(request.authority)) throw new AgentSessionRuntimeError("request authority is required");
	assertOnlyKeys(request.authority, [
		"issue", "branch", "workspaceId", "readOnly", "readPrefixes", "writePrefixes", "capabilityNames",
	], "authority");
	if (!Number.isSafeInteger(request.authority.issue) || request.authority.issue < 1) {
		throw new AgentSessionRuntimeError("authority issue is invalid");
	}
	if (typeof request.authority.branch !== "string" || request.authority.branch.length < 1 ||
		request.authority.branch.length > 255 || /[\u0000-\u001f\u007f]/.test(request.authority.branch) ||
		request.authority.branch === "main") {
		throw new AgentSessionRuntimeError("authority branch is invalid or targets main");
	}
	if (typeof request.authority.readOnly !== "boolean") throw new AgentSessionRuntimeError("authority readOnly is invalid");
	if (!request.binding || !isRecord(request.binding)) throw new AgentSessionRuntimeError("request binding is required");
	assertOnlyKeys(request.binding, ["runId", "generation", "laneId", "candidateHead", "validationNonce"], "binding");
	if (!validIdentifier(request.binding.runId) || !validIdentifier(request.binding.laneId) ||
		!Number.isSafeInteger(request.binding.generation) || request.binding.generation < 1 ||
		!/^[0-9a-f]{40}$/.test(request.binding.candidateHead) ||
		!validIdentifier(request.binding.validationNonce) || request.binding.validationNonce.length < 12) {
		throw new AgentSessionRuntimeError("request binding is invalid");
	}
	// Prompt construction performs the byte-level task/context bounds before any SDK call.
	buildRolePrompts({
		role: request.role,
		task: request.task,
		context: request.context,
		authority: {
			issue: request.authority.issue,
			branch: request.authority.branch,
			workspaceId: request.authority.workspaceId,
			readOnly: request.authority.readOnly,
			toolNames: [],
			binding: request.binding,
		},
	});
}

function validateCreatedSession(
	created: RuntimeSessionResult,
	thinking: ShepherdAgentThinking,
	expectedTools: string[],
): void {
	if (!created || !created.extensionsResult || !Array.isArray(created.extensionsResult.extensions) ||
		!Array.isArray(created.extensionsResult.errors)) {
		throw new AgentSessionRuntimeError("Pi returned an invalid AgentSession result");
	}
	if (created.modelFallbackMessage) throw new AgentSessionRuntimeError("Pi attempted a forbidden model fallback");
	if (created.extensionsResult.extensions.length > 0 || created.extensionsResult.errors.length > 0) {
		throw new AgentSessionRuntimeError("embedded AgentSession unexpectedly loaded extensions or extension errors");
	}
	const session = created.session;
	if (!session || typeof session.prompt !== "function" || typeof session.abort !== "function" ||
		typeof session.waitForIdle !== "function" || typeof session.dispose !== "function" ||
		typeof session.subscribe !== "function" || typeof session.getActiveToolNames !== "function") {
		throw new AgentSessionRuntimeError("Pi returned an incomplete AgentSession surface");
	}
	if (session.model?.provider !== REQUIRED_PROVIDER || session.model?.id !== REQUIRED_MODEL) {
		throw new AgentSessionRuntimeError("embedded AgentSession model routing mismatch");
	}
	if (session.thinkingLevel !== thinking) throw new AgentSessionRuntimeError("embedded AgentSession thinking route mismatch");
	if (session.sessionFile !== undefined) throw new AgentSessionRuntimeError("embedded AgentSession persistence is forbidden");
	const activeTools = session.getActiveToolNames();
	if (!Array.isArray(activeTools) || activeTools.length !== expectedTools.length ||
		activeTools.some((name, index) => name !== expectedTools[index])) {
		throw new AgentSessionRuntimeError("embedded AgentSession active tool authority drifted");
	}
}

function ownedSessionFromResult(created: RuntimeSessionResult): OwnedSession {
	const session = created?.session;
	if (!session || typeof session.abort !== "function" || typeof session.waitForIdle !== "function" ||
		typeof session.dispose !== "function") {
		throw new AgentSessionRuntimeError("Pi returned an invalid AgentSession result without a cleanable session");
	}
	return new OwnedSession(session);
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
	try {
		capture.eventBytes += byteLength(JSON.stringify(event));
	} catch (error) {
		capture.failure = new AgentSessionRuntimeError("AgentSession emitted an unserializable event", { cause: error });
		return;
	}
	if (capture.eventCount > options.maxEvents || capture.eventBytes > options.maxEventBytes) {
		capture.failure = new AgentSessionRuntimeError("AgentSession event stream exceeded its bound");
		return;
	}
	if (!isRecord(event) || typeof event.type !== "string") return;
	if (event.type === "message_end" && isAssistantTerminal(event.message)) capture.messageEnd = event.message;
	if (event.type === "agent_end") {
		capture.agentEndCount += 1;
		capture.agentEndWillRetry ||= event.willRetry === true;
		if (Array.isArray(event.messages)) {
			capture.agentEnd = [...event.messages].reverse().find(isAssistantTerminal);
		}
	}
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
	if (typeof value !== "string" || (!allowEmpty && value.length < 1) || value.length > max || /[\u0000\u007f]/.test(value)) {
		throw new AgentSessionRuntimeError(`${description} must be ${allowEmpty ? "a" : "a non-empty"} bounded string`);
	}
	return redactSensitiveText(value);
}

function computeDeadline(timeoutMs: number, deadlineAt: number | undefined): number {
	const timeoutDeadline = Date.now() + timeoutMs;
	return deadlineAt === undefined ? timeoutDeadline : Math.min(timeoutDeadline, deadlineAt);
}

function isAssistantTerminal(value: unknown): value is AssistantTerminal {
	return isRecord(value) && value.role === "assistant" && typeof value.provider === "string" &&
		typeof value.model === "string" && typeof value.stopReason === "string" &&
		typeof value.timestamp === "number" && Array.isArray(value.content);
}

function isAbsoluteNonTraversingPath(value: unknown): value is string {
	if (typeof value !== "string" || value.length < 1 || value.length > 4_096 || /[\u0000-\u001f\u007f]/.test(value)) return false;
	const flavor = win32.isAbsolute(value) ? win32 : posix;
	if (!flavor.isAbsolute(value)) return false;
	const segments = flavor === win32 ? value.split(/[\\/]+/) : value.split("/");
	return !segments.includes("..");
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

function positiveInteger(value: number, description: string): number {
	if (!Number.isSafeInteger(value) || value <= 0) throw new AgentSessionRuntimeError(`${description} must be a positive safe integer`);
	return value;
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
	| { settled: true; value: T }
	| { settled: false };

async function settleWithin<T>(operation: Promise<T>, timeoutMs: number): Promise<BoundedSettlement<T>> {
	let timer: ReturnType<typeof setTimeout> | undefined;
	const timeout = new Promise<BoundedSettlement<T>>((resolve) => {
		timer = setTimeout(() => resolve({ settled: false }), timeoutMs);
	});
	try {
		return await Promise.race([
			operation.then((value): BoundedSettlement<T> => ({ settled: true, value })),
			timeout,
		]);
	} finally {
		if (timer) clearTimeout(timer);
	}
}

function normalizeRuntimeError(error: unknown): AgentSessionRuntimeError {
	return error instanceof AgentSessionRuntimeError
		? error
		: new AgentSessionRuntimeError("AgentSession run failed", { cause: error });
}
