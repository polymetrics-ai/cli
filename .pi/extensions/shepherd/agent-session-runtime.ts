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
	normalizeScopedPrefixes,
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
const MAX_CONCURRENCY = 32;
const MAX_EVENTS = 65_536;
const MAX_EVENT_BYTES = 16 * 1024 * 1024;
const MAX_ASSISTANT_BYTES = 1024 * 1024;
const MAX_CLEANUP_TIMEOUT_MS = MAX_TIMEOUT_MS;
const MAX_EVENT_DEPTH = 64;
const MAX_EVENT_NODES = 65_536;
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
	unsubscribe: (() => void | PromiseLike<void>) | undefined;
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
			this.#disposePromise = Promise.resolve().then(async () => {
				let failurePresent = false;
				let firstError: unknown;
				const unsubscribe = this.unsubscribe;
				this.unsubscribe = undefined;
				try {
					if (unsubscribe) await Promise.resolve(unsubscribe());
				} catch (error) {
					failurePresent = true;
					firstError = error;
				}
				try {
					await Promise.resolve(this.session.dispose());
				} catch (error) {
					if (!failurePresent) {
						failurePresent = true;
						firstError = error;
					}
				}
				if (failurePresent) throw firstError;
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
	let failurePresent = false;
	let firstError: unknown;
	if (abort) {
		try {
			await bounded(owned.abortOnce(), timeoutMs, "session abort");
		} catch (error) {
			failurePresent = true;
			firstError = error;
		}
	}
	try {
		await bounded(owned.waitOnce(), timeoutMs, "session idle wait");
	} catch (error) {
		if (!failurePresent) {
			failurePresent = true;
			firstError = error;
		}
	}
	// Teardown must remain reachable after either bounded phase, including for thenable adapters.
	try {
		await owned.disposeOnce();
	} catch (error) {
		if (!failurePresent) {
			failurePresent = true;
			firstError = error;
		}
	}
	if (failurePresent) throw firstError;
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
		let failurePresent = false;
		let firstError: unknown;
		try {
			owned = ownedSessionFromResult(created);
		} catch (error) {
			failurePresent = true;
			firstError = error;
		}
		if (owned) {
			try {
				this.#validateCreated(created);
			} catch (error) {
				if (!failurePresent) {
					failurePresent = true;
					firstError = error;
				}
			}
			try {
				await this.#cleanup(owned);
			} catch (error) {
				if (!failurePresent) {
					failurePresent = true;
					firstError = error;
				}
			}
		}
		if (failurePresent) this.#reportFailure(firstError);
		this.#settleTerminal();
	}

	async #cleanup(owned: OwnedSession): Promise<void> {
		const deadlineAt = Date.now() + this.#cleanupTimeoutMs;
		let failurePresent = false;
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
			if (settlement.status === "rejected" && !failurePresent) {
				failurePresent = true;
				firstError = settlement.reason;
			}
		}
		try {
			await owned.disposeOnce();
		} catch (error) {
			if (!failurePresent) {
				failurePresent = true;
				firstError = error;
			}
		}
		if (failurePresent) throw firstError;
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

	attach(): void {
		if (this.#mayBeAttached) return;
		// EventTarget may mutate before a hostile wrapper throws. Own the possible lease first.
		this.#mayBeAttached = true;
		this.#signal.addEventListener("abort", this.#listener, { once: true });
	}

	release(): void {
		if (this.#released || !this.#mayBeAttached) return;
		this.#released = true;
		this.#signal.removeEventListener("abort", this.#listener);
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
	content: Array<{ type: string; text?: string }>;
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
		const normalizedRequest = normalizeRunRequest(request);
		this.#assertSdk();
		this.#assertOpen();
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
		const listenerLease = normalizedRequest.signal
			? new AbortListenerLease(normalizedRequest.signal, externalAbort)
			: undefined;
		let listenerFailurePresent = false;
		let listenerFailure: unknown;
		try {
			if (normalizedRequest.signal) {
				listenerLease?.attach();
				if (normalizedRequest.signal.aborted) externalAbort();
			}
			return await this.#execute(normalizedRequest, route.thinking, model, active);
		} finally {
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
			if (listenerFailurePresent) throw listenerFailure;
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
				this.#setQuarantine(new AgentSessionRuntimeError(
					"AgentSession creation remained pending during bounded close",
				));
			}
		}
		this.#closed = true;
		this.#parentListenerLease?.release();
		if (this.#quarantineFailurePresent) {
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
					readPrefixes: request.authority.readPrefixes,
					writePrefixes: request.authority.writePrefixes,
					toolNames: toolPolicy.names,
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
					this.#setQuarantine(error);
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
			primaryFailurePresent = true;
			primaryFailure = error;
		}

		let cleanupFailurePresent = false;
		let cleanupFailure: unknown;
		if (reloadPromise) {
			try {
				await bounded(reloadPromise, this.#options.cleanupTimeoutMs, "resource loader settlement");
			} catch (error) {
				cleanupFailurePresent = true;
				cleanupFailure = error;
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
			throw new AgentSessionRuntimeError("AgentSession cleanup/join failed; runtime quarantined", { cause: cleanupFailure });
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
		parameters: isRecord(parameters) ? Object.freeze({ ...parameters }) : parameters,
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
		if (seen.has(object)) throw new AgentSessionRuntimeError("AgentSession event contains a cycle");
		seen.add(object);
		const descriptors = Object.getOwnPropertyDescriptors(object);
		add(Array.isArray(value) ? 2 : 2);
		for (const key of Reflect.ownKeys(descriptors)) {
			if (typeof key !== "string") throw new AgentSessionRuntimeError("AgentSession event contains a symbol key");
			const descriptor = descriptors[key];
			if (!descriptor?.enumerable) continue;
			if (descriptor.get || descriptor.set || !("value" in descriptor)) {
				throw new AgentSessionRuntimeError("AgentSession event contains an accessor");
			}
			add(byteLength(key) + 3);
			stack.push({ value: descriptor.value, depth: current.depth + 1 });
		}
	}
	return bytes;
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
	if (typeof value !== "string" || (!allowEmpty && value.length < 1) || value.length > max ||
		/[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f-\u009f]/.test(value)) {
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
	return Object.freeze(values) as unknown as T[];
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
