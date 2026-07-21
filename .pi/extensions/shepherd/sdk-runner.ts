import { posix, win32 } from "node:path";

import type { AgentSessionEvent } from "@earendil-works/pi-coding-agent";

import type {
	AgentRunner,
	AgentRunRequest,
	AgentRunResult,
	DimensionScores,
} from "./runner.ts";

const REQUIRED_PI_VERSION = "0.80.6";
const REQUIRED_PROVIDER = "openai-codex";
const REQUIRED_MODEL = "gpt-5.6-sol";
const DIMENSION_NAMES = [
	"correctStage",
	"artifactValid",
	"gatesRespected",
	"realProgress",
	"noHallucination",
	"noConflict",
] as const;

type SessionModel = { provider: string; id: string };
type MessageEndEvent = Extract<AgentSessionEvent, { type: "message_end" }>;
type AssistantTerminalMessage = Extract<MessageEndEvent["message"], { role: "assistant" }>;

interface EmbeddedSession {
	model: SessionModel;
	thinkingLevel: string;
	sessionFile: string | undefined;
	getActiveToolNames(): string[];
	subscribe(listener: (event: AgentSessionEvent) => void): () => void;
	prompt(prompt: string, options: { expandPromptTemplates: false; source: "extension" }): Promise<void>;
	waitForIdle(): Promise<void>;
	abort(): Promise<void>;
	dispose(): void;
}

interface ExtensionLoadResult {
	extensions: unknown[];
	errors: unknown[];
}

interface ResourceLoader {
	reload(): Promise<void>;
}

export interface ShepherdSdk {
	version: string;
	requiredVersion?: string;
	getAgentDir(): string;
	createSettingsManager(settings: Record<string, unknown>, options: Record<string, unknown>): unknown;
	createSessionManager(cwd: string): unknown;
	createResourceLoader(options: Record<string, unknown>): ResourceLoader;
	createSession(options: Record<string, unknown>): Promise<{
		session: EmbeddedSession;
		extensionsResult: ExtensionLoadResult;
	}>;
}

export interface ShepherdModelRegistry {
	authStorage: unknown;
	find(provider: string, model: string): SessionModel | undefined;
	hasConfiguredAuth(model: SessionModel): boolean;
}

export interface SdkAgentRunnerOptions {
	maxConcurrency?: number;
	maxAssistantBytes?: number;
	maxSummaryCharacters?: number;
	maxEvents?: number;
	maxEventBytes?: number;
	cleanupTimeoutMs?: number;
}

export class AgentRunnerError extends Error {
	constructor(message: string, options?: ErrorOptions) {
		super(message, options);
		this.name = "AgentRunnerError";
	}
}

class OwnedChild {
	readonly runId: string;
	readonly laneKey: string;
	readonly session: EmbeddedSession;
	unsubscribe: (() => void) | undefined;

	#abortPromise: Promise<void> | undefined;
	#cleanupPromise: Promise<void> | undefined;
	readonly #cleanupTimeoutMs: number;

	constructor(runId: string, laneKey: string, session: EmbeddedSession, cleanupTimeoutMs: number) {
		this.runId = runId;
		this.laneKey = laneKey;
		this.session = session;
		this.#cleanupTimeoutMs = cleanupTimeoutMs;
	}

	abortOnce(): Promise<void> {
		if (!this.#abortPromise) {
			this.#abortPromise = Promise.resolve().then(() => this.session.abort());
		}
		return this.#abortPromise;
	}

	cleanup(deadlineAt = Date.now() + this.#cleanupTimeoutMs): Promise<void> {
		if (!this.#cleanupPromise) {
			this.#cleanupPromise = this.#cleanup(deadlineAt);
		}
		return this.#cleanupPromise;
	}

	async #cleanup(deadlineAt: number): Promise<void> {
		let firstError: unknown;
		let timedOut = false;
		try {
			await boundedCleanupUntil(this.abortOnce(), deadlineAt, "abort");
		} catch (error) {
			firstError = error;
			timedOut ||= isCleanupTimeout(error);
		}
		try {
			await boundedCleanupUntil(this.session.waitForIdle(), deadlineAt, "wait-for-idle");
		} catch (error) {
			firstError ??= error;
			timedOut ||= isCleanupTimeout(error);
		}
		try {
			this.unsubscribe?.();
		} catch (error) {
			firstError ??= error;
		} finally {
			this.unsubscribe = undefined;
		}
		try {
			this.session.dispose();
		} catch (error) {
			firstError ??= error;
		}
		if (firstError) {
			if (timedOut) throw new AgentRunnerError("AgentSession cleanup timed out", { cause: firstError });
			throw new AgentRunnerError("AgentSession cleanup failed", { cause: firstError });
		}
	}
}

class RunScope {
	readonly deadlineAt: number;
	readonly signal: AbortSignal;
	readonly #controller = new AbortController();
	readonly #termination: Promise<never>;
	#rejectTermination: ((error: AgentRunnerError) => void) | undefined;
	#timer: ReturnType<typeof setTimeout> | undefined;
	#finished = false;

	constructor(timeoutMs: number) {
		this.deadlineAt = Date.now() + timeoutMs;
		this.signal = this.#controller.signal;
		this.#termination = new Promise<never>((_resolve, reject) => {
			this.#rejectTermination = reject;
		});
		this.#termination.catch(() => undefined);
		this.#timer = setTimeout(() => {
			if (this.#finished || this.signal.aborted) return;
			this.#controller.abort();
			this.#rejectTermination?.(new AgentRunnerError(`AgentSession timed out after ${timeoutMs}ms`));
		}, timeoutMs);
	}

	race<T>(operation: Promise<T>): Promise<T> {
		return Promise.race([operation, this.#termination]);
	}

	abort(): void {
		if (this.#finished || this.signal.aborted) return;
		this.#controller.abort();
		this.#rejectTermination?.(new AgentRunnerError("AgentSession run was cancelled"));
	}

	finish(): void {
		this.#finished = true;
		if (this.#timer) clearTimeout(this.#timer);
		this.#timer = undefined;
		this.#rejectTermination = undefined;
	}
}

function deferred<T>(): { promise: Promise<T>; resolve(value: T): void } {
	let resolvePromise: ((value: T) => void) | undefined;
	const promise = new Promise<T>((resolve) => { resolvePromise = resolve; });
	return { promise, resolve: (value) => resolvePromise?.(value) };
}

/**
 * Experimental Pi AgentSession adapter.
 *
 * The runtime SDK is injected by index.ts so this module remains independently testable and does
 * not make globally-installed Pi packages a Node test dependency.
 */
export class SdkAgentRunner implements AgentRunner {
	#sdk: ShepherdSdk;
	#modelRegistry: ShepherdModelRegistry;
	#options: Required<SdkAgentRunnerOptions>;
	#children = new Map<string, Set<OwnedChild>>();
	#activeLaneKeys = new Set<string>();
	#activeRunCounts = new Map<string, number>();
	#cancelledRuns = new Set<string>();
	#runScopes = new Map<string, Set<RunScope>>();
	#setupTasks = new Set<Promise<void>>();
	#idleWaiters = new Set<() => void>();
	#activeCount = 0;
	#activeMutator = false;
	#closing = false;
	#closed = false;
	#closePromise: Promise<void> | undefined;

	constructor(
		sdk: ShepherdSdk,
		modelRegistry: ShepherdModelRegistry,
		options: SdkAgentRunnerOptions = {},
	) {
		this.#sdk = sdk;
		this.#modelRegistry = modelRegistry;
		this.#options = {
			maxConcurrency: options.maxConcurrency ?? 2,
			maxAssistantBytes: options.maxAssistantBytes ?? 64 * 1024,
			maxSummaryCharacters: options.maxSummaryCharacters ?? 4 * 1024,
			maxEvents: options.maxEvents ?? 4_096,
			maxEventBytes: options.maxEventBytes ?? 4 * 1024 * 1024,
			cleanupTimeoutMs: options.cleanupTimeoutMs ?? 5_000,
		};
		if (!Number.isInteger(this.#options.maxConcurrency) || this.#options.maxConcurrency < 1 || this.#options.maxConcurrency > 2) {
			throw new AgentRunnerError("embedded AgentSession concurrency must be between 1 and 2");
		}
		for (const [name, value] of Object.entries(this.#options)) {
			if (!Number.isSafeInteger(value) || value <= 0) {
				throw new AgentRunnerError(`${name} must be a positive safe integer`);
			}
		}
	}

	async run(request: AgentRunRequest): Promise<AgentRunResult> {
		validateRequest(request);
		this.#assertSdkContract();
		this.#assertOpen();

		const model = this.#modelRegistry.find(request.provider, request.model);
		if (!model || model.provider !== request.provider || model.id !== request.model) {
			throw new AgentRunnerError(`required model ${request.provider}/${request.model} is unavailable`);
		}
		if (!this.#modelRegistry.hasConfiguredAuth(model)) {
			throw new AgentRunnerError(`required model ${request.provider}/${request.model} has no configured auth`);
		}

		const laneKey = `${request.runId}:${request.binding.generation}:${request.laneId}`;
		this.#reserve(request, laneKey);
		const scope = new RunScope(request.timeoutMs);
		this.#registerScope(request.runId, scope);
		const onAbort = () => { void this.abort(request.runId).catch(() => undefined); };
		request.signal?.addEventListener("abort", onAbort, { once: true });
		let child: OwnedChild | undefined;
		let onScopeAbort: (() => void) | undefined;
		let primaryFailed = false;

		try {
			this.#assertRunActive(request);
			const settingsManager = this.#sdk.createSettingsManager(
				{
					defaultProvider: request.provider,
					defaultModel: request.model,
					defaultThinkingLevel: request.thinking,
					compaction: { enabled: false },
					retry: { enabled: false },
					packages: [],
					extensions: [],
					skills: [],
					prompts: [],
					themes: [],
				},
				{ projectTrusted: false },
			);
			const sessionManager = this.#sdk.createSessionManager(request.cwd);
			const resourceLoader = this.#sdk.createResourceLoader({
				cwd: request.cwd,
				agentDir: this.#sdk.getAgentDir(),
				settingsManager,
				noExtensions: true,
				noSkills: true,
				noPromptTemplates: true,
				noThemes: true,
				noContextFiles: true,
				systemPrompt: request.systemPrompt,
			});
			const reloadPromise = Promise.resolve().then(() => resourceLoader.reload());
			this.#trackSetup(reloadPromise);
			await scope.race(reloadPromise);
			this.#assertRunActive(request);

			const creationPromise = Promise.resolve().then(() => this.#sdk.createSession({
				cwd: request.cwd,
				agentDir: this.#sdk.getAgentDir(),
				authStorage: this.#modelRegistry.authStorage,
				modelRegistry: this.#modelRegistry,
				model,
				thinkingLevel: request.thinking,
				noTools: "all",
				tools: [],
				customTools: [],
				resourceLoader,
				sessionManager,
				settingsManager,
			}));
			const creationDecision = deferred<"claimed" | "abandoned">();
			const creationLifecycle = creationPromise.then(async (created) => {
				if (await creationDecision.promise === "abandoned") {
					const lateChild = new OwnedChild(
						request.runId,
						laneKey,
						created.session,
						this.#options.cleanupTimeoutMs,
					);
					await lateChild.cleanup();
				}
			}, () => undefined);
			this.#trackSetup(creationLifecycle);

			let created: Awaited<typeof creationPromise>;
			try {
				created = await scope.race(creationPromise);
				creationDecision.resolve("claimed");
			} catch (error) {
				creationDecision.resolve("abandoned");
				throw error;
			}
			child = new OwnedChild(request.runId, laneKey, created.session, this.#options.cleanupTimeoutMs);
			this.#register(child);
			onScopeAbort = () => { void child?.abortOnce().catch(() => undefined); };
			scope.signal.addEventListener("abort", onScopeAbort, { once: true });
			this.#assertRunActive(request);
			validateCreatedSession(created, request);

			let eventCount = 0;
			let eventBytes = 0;
			let eventFailure: AgentRunnerError | undefined;
			let messageEndTerminal: AssistantTerminalMessage | undefined;
			let agentEndTerminal: AssistantTerminalMessage | undefined;
			let agentEndCount = 0;
			let agentEndWillRetry = false;
			child.unsubscribe = created.session.subscribe((event) => {
				if (eventFailure) return;
				eventCount += 1;
				try {
					eventBytes += byteLength(JSON.stringify(event));
				} catch (error) {
					eventFailure = new AgentRunnerError("AgentSession emitted an unserializable event", { cause: error });
				}
				if (eventCount > this.#options.maxEvents || eventBytes > this.#options.maxEventBytes) {
					eventFailure = new AgentRunnerError("AgentSession event limit exceeded");
				}
				if (!eventFailure && event.type === "message_end" && event.message.role === "assistant") {
					messageEndTerminal = event.message;
				}
				if (!eventFailure && event.type === "agent_end") {
					agentEndCount += 1;
					agentEndWillRetry ||= event.willRetry;
					agentEndTerminal = [...event.messages].reverse().find(
						(message): message is AssistantTerminalMessage => message.role === "assistant",
					);
				}
				if (eventFailure) void child?.abortOnce().catch(() => undefined);
			});

			this.#assertRunActive(request);
			await scope.race(
				created.session.prompt(request.prompt, {
					expandPromptTemplates: false,
					source: "extension",
				}),
			);
			this.#assertRunActive(request);
			if (eventFailure) throw eventFailure;
			const terminal = verifyTerminalEvents({
				messageEndTerminal,
				agentEndTerminal,
				agentEndCount,
				agentEndWillRetry,
			});
			const assistantText = assistantTextFromTerminal(terminal);
			return parseEvidence(assistantText, this.#options.maxAssistantBytes, this.#options.maxSummaryCharacters);
		} catch (error) {
			primaryFailed = true;
			throw error;
		} finally {
			try {
				await child?.cleanup(Math.min(scope.deadlineAt, Date.now() + this.#options.cleanupTimeoutMs));
			} catch (error) {
				if (!primaryFailed) throw error;
			} finally {
				if (onScopeAbort) scope.signal.removeEventListener("abort", onScopeAbort);
				request.signal?.removeEventListener("abort", onAbort);
				if (child) this.#unregister(child);
				scope.finish();
				this.#unregisterScope(request.runId, scope);
				this.#release(request, laneKey);
			}
		}
	}

	async abort(runId: string): Promise<void> {
		if ((this.#activeRunCounts.get(runId) ?? 0) > 0) this.#cancelledRuns.add(runId);
		for (const scope of this.#runScopes.get(runId) ?? []) scope.abort();
		const children = [...(this.#children.get(runId) ?? [])];
		await Promise.all(children.map((child) =>
			boundedCleanup(child.abortOnce(), this.#options.cleanupTimeoutMs, "abort"),
		));
	}

	close(): Promise<void> {
		if (!this.#closePromise) {
			this.#closing = true;
			this.#closePromise = this.#close();
		}
		return this.#closePromise;
	}

	async #close(): Promise<void> {
		const deadlineAt = Date.now() + this.#options.cleanupTimeoutMs;
		for (const runId of this.#activeRunCounts.keys()) this.#cancelledRuns.add(runId);
		for (const scopes of this.#runScopes.values()) {
			for (const scope of scopes) scope.abort();
		}
		const children = [...this.#children.values()].flatMap((entries) => [...entries]);
		const childResultsPromise = Promise.allSettled(
			children.map((child) => child.cleanup(deadlineAt)),
		);
		let setupResults: PromiseSettledResult<void>[];
		try {
			setupResults = await boundedCleanupUntil(
				Promise.allSettled([...this.#setupTasks]),
				deadlineAt,
				"runner-close",
			);
		} catch (error) {
			await childResultsPromise;
			throw new AgentRunnerError("AgentSession runner close timed out", { cause: error });
		}
		const childResults = await childResultsPromise;
		await this.#waitForIdle();
		const failure = [...setupResults, ...childResults].find((result) => result.status === "rejected");
		if (failure?.status === "rejected") {
			throw new AgentRunnerError("one or more AgentSessions failed to close", { cause: failure.reason });
		}
		this.#closed = true;
	}

	#assertSdkContract(): void {
		if (this.#sdk.version !== REQUIRED_PI_VERSION ||
			(this.#sdk.requiredVersion !== undefined && this.#sdk.requiredVersion !== REQUIRED_PI_VERSION)) {
			throw new AgentRunnerError(`AgentSession Shepherd requires Pi ${REQUIRED_PI_VERSION}; found ${this.#sdk.version}`);
		}
		for (const name of [
			"getAgentDir",
			"createSettingsManager",
			"createSessionManager",
			"createResourceLoader",
			"createSession",
		] as const) {
			if (typeof this.#sdk[name] !== "function") {
				throw new AgentRunnerError(`Pi ${REQUIRED_PI_VERSION} SDK surface is missing ${name}`);
			}
		}
		if (typeof this.#modelRegistry.find !== "function" || typeof this.#modelRegistry.hasConfiguredAuth !== "function") {
			throw new AgentRunnerError("Pi model registry surface is incomplete");
		}
	}

	#assertOpen(): void {
		if (this.#closing || this.#closed) throw new AgentRunnerError("AgentSession runner is closed");
	}

	#assertRunActive(request: AgentRunRequest): void {
		this.#assertOpen();
		if (request.signal?.aborted || this.#cancelledRuns.has(request.runId)) {
			throw new AgentRunnerError(`AgentSession run ${request.runId} was cancelled`);
		}
	}

	#reserve(request: AgentRunRequest, laneKey: string): void {
		if (this.#activeLaneKeys.has(laneKey)) {
			throw new AgentRunnerError(`lane ${request.laneId} is already active for this run generation`);
		}
		if (this.#activeCount >= this.#options.maxConcurrency) {
			throw new AgentRunnerError(`embedded AgentSession concurrency limit ${this.#options.maxConcurrency} reached`);
		}
		if (!request.readOnly && this.#activeMutator) {
			throw new AgentRunnerError("only one mutating AgentSession may run at a time");
		}
		this.#activeCount += 1;
		this.#activeRunCounts.set(request.runId, (this.#activeRunCounts.get(request.runId) ?? 0) + 1);
		this.#activeMutator ||= !request.readOnly;
		this.#activeLaneKeys.add(laneKey);
	}

	#release(request: AgentRunRequest, laneKey: string): void {
		if (!this.#activeLaneKeys.delete(laneKey)) return;
		this.#activeCount -= 1;
		const remaining = (this.#activeRunCounts.get(request.runId) ?? 1) - 1;
		if (remaining <= 0) {
			this.#activeRunCounts.delete(request.runId);
			this.#cancelledRuns.delete(request.runId);
		} else {
			this.#activeRunCounts.set(request.runId, remaining);
		}
		if (!request.readOnly) this.#activeMutator = false;
		if (this.#activeCount === 0) {
			for (const resolve of this.#idleWaiters) resolve();
			this.#idleWaiters.clear();
		}
	}

	#register(child: OwnedChild): void {
		const children = this.#children.get(child.runId) ?? new Set<OwnedChild>();
		children.add(child);
		this.#children.set(child.runId, children);
	}

	#registerScope(runId: string, scope: RunScope): void {
		const scopes = this.#runScopes.get(runId) ?? new Set<RunScope>();
		scopes.add(scope);
		this.#runScopes.set(runId, scopes);
	}

	#unregisterScope(runId: string, scope: RunScope): void {
		const scopes = this.#runScopes.get(runId);
		if (!scopes) return;
		scopes.delete(scope);
		if (scopes.size === 0) this.#runScopes.delete(runId);
	}

	#trackSetup(operation: Promise<unknown>): void {
		let tracked: Promise<void>;
		tracked = operation.then(() => undefined).finally(() => this.#setupTasks.delete(tracked));
		tracked.catch(() => undefined);
		this.#setupTasks.add(tracked);
	}

	#unregister(child: OwnedChild): void {
		const children = this.#children.get(child.runId);
		if (!children) return;
		children.delete(child);
		if (children.size === 0) this.#children.delete(child.runId);
	}

	#waitForIdle(): Promise<void> {
		if (this.#activeCount === 0) return Promise.resolve();
		return new Promise((resolve) => this.#idleWaiters.add(resolve));
	}
}

function validateRequest(request: AgentRunRequest): void {
	if (request.provider !== REQUIRED_PROVIDER || request.model !== REQUIRED_MODEL) {
		throw new AgentRunnerError(`sdk-inproc requires ${REQUIRED_PROVIDER}/${REQUIRED_MODEL}`);
	}
	const expectedThinking = request.readOnly ? "xhigh" : "high";
	if (request.thinking !== expectedThinking) {
		throw new AgentRunnerError(`${request.readOnly ? "read-only" : "mutating"} lanes require ${expectedThinking} thinking`);
	}
	if (!Number.isSafeInteger(request.timeoutMs) || request.timeoutMs <= 0) {
		throw new AgentRunnerError("timeoutMs must be a positive safe integer");
	}
	for (const [name, value] of [
		["runId", request.runId],
		["laneId", request.laneId],
		["role", request.role],
	] as const) {
		if (!/^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$/.test(value)) {
			throw new AgentRunnerError(`${name} is invalid`);
		}
	}
	if (!isAbsoluteNonTraversingPath(request.cwd)) {
		throw new AgentRunnerError("cwd must be an absolute non-traversing path without control characters");
	}
	if (request.systemPrompt.length === 0 || request.systemPrompt.length > 32 * 1024 ||
		request.prompt.length === 0 || request.prompt.length > 64 * 1024) {
		throw new AgentRunnerError("AgentSession prompts must be non-empty and bounded");
	}
	if (!Array.isArray(request.tools) || request.tools.length !== 0) {
		throw new AgentRunnerError("embedded AgentSession child tools are disabled; tools must be []");
	}
	const untrustedRequest = request as AgentRunRequest & { customTools?: unknown };
	if (untrustedRequest.customTools !== undefined &&
		(!Array.isArray(untrustedRequest.customTools) || untrustedRequest.customTools.length !== 0)) {
		throw new AgentRunnerError("embedded AgentSession custom child tools are disabled");
	}

	const bindingPairs: Array<[string, unknown, unknown]> = [
		["runId", request.runId, request.binding.runId],
		["laneId", request.laneId, request.binding.laneId],
		["readOnly", request.readOnly, request.binding.readOnly],
		["provider", request.provider, request.binding.provider],
		["model", request.model, request.binding.model],
		["thinking", request.thinking, request.binding.thinking],
	];
	for (const [name, expected, actual] of bindingPairs) {
		if (actual !== expected) throw new AgentRunnerError(`request ${name} does not match its binding`);
	}
	if (!Number.isSafeInteger(request.binding.generation) || request.binding.generation < 1) {
		throw new AgentRunnerError("binding generation is invalid");
	}
	if (!/^[0-9a-f]{40}$/.test(request.binding.candidateHead)) {
		throw new AgentRunnerError("binding candidate head is invalid");
	}
	if (!/^[A-Za-z0-9._-]{12,128}$/.test(request.binding.validationNonce)) {
		throw new AgentRunnerError("binding validation nonce is invalid");
	}
}

function validateCreatedSession(
	created: { session: EmbeddedSession; extensionsResult: ExtensionLoadResult },
	request: AgentRunRequest,
): void {
	if (!created.extensionsResult || !Array.isArray(created.extensionsResult.extensions) ||
		!Array.isArray(created.extensionsResult.errors)) {
		throw new AgentRunnerError("Pi returned an invalid extension load result");
	}
	if (created.extensionsResult.extensions.length > 0) {
		throw new AgentRunnerError("embedded AgentSession unexpectedly loaded extensions");
	}
	if (created.extensionsResult.errors.length > 0) {
		throw new AgentRunnerError("embedded AgentSession reported extension loading errors");
	}
	const session = created.session;
	if (!session || typeof session.prompt !== "function" || typeof session.abort !== "function" ||
		typeof session.waitForIdle !== "function" || typeof session.subscribe !== "function" ||
		typeof session.dispose !== "function" || typeof session.getActiveToolNames !== "function") {
		throw new AgentRunnerError("Pi returned an incomplete AgentSession surface");
	}
	if (session.model?.provider !== request.provider || session.model?.id !== request.model) {
		throw new AgentRunnerError("embedded AgentSession model routing mismatch");
	}
	if (session.thinkingLevel !== request.thinking) {
		throw new AgentRunnerError("embedded AgentSession thinking level mismatch");
	}
	if (session.sessionFile !== undefined) {
		throw new AgentRunnerError("embedded AgentSession unexpectedly enabled persistence");
	}
	const activeToolNames = session.getActiveToolNames();
	if (!Array.isArray(activeToolNames) || activeToolNames.length !== 0) {
		throw new AgentRunnerError("embedded AgentSession must expose zero active tools");
	}
}

function verifyTerminalEvents(events: {
	messageEndTerminal: AssistantTerminalMessage | undefined;
	agentEndTerminal: AssistantTerminalMessage | undefined;
	agentEndCount: number;
	agentEndWillRetry: boolean;
}): AssistantTerminalMessage {
	const { messageEndTerminal, agentEndTerminal, agentEndCount, agentEndWillRetry } = events;
	if (agentEndCount !== 1 || agentEndWillRetry || !messageEndTerminal || !agentEndTerminal ||
		!sameTerminalMessage(messageEndTerminal, agentEndTerminal)) {
		throw new AgentRunnerError("AgentSession returned an invalid terminal event sequence");
	}
	if (agentEndTerminal.stopReason !== "stop") {
		throw new AgentRunnerError(`AgentSession terminal stop reason ${agentEndTerminal.stopReason} is not accepted`);
	}
	return agentEndTerminal;
}

function sameTerminalMessage(left: AssistantTerminalMessage, right: AssistantTerminalMessage): boolean {
	return left.stopReason === right.stopReason &&
		left.timestamp === right.timestamp &&
		left.provider === right.provider &&
		left.model === right.model &&
		assistantTextFromTerminal(left) === assistantTextFromTerminal(right);
}

function assistantTextFromTerminal(message: AssistantTerminalMessage): string | undefined {
	let text = "";
	for (const content of message.content) {
		if (content.type === "text") text += content.text;
	}
	return text.trim() || undefined;
}

function parseEvidence(
	text: string | undefined,
	maxAssistantBytes: number,
	maxSummaryCharacters: number,
): AgentRunResult {
	if (!text) throw new AgentRunnerError("AgentSession returned no assistant evidence");
	if (byteLength(text) > maxAssistantBytes) {
		throw new AgentRunnerError("AgentSession assistant evidence exceeded the output limit");
	}

	let candidate: unknown;
	try {
		candidate = JSON.parse(text);
	} catch (error) {
		throw new AgentRunnerError("AgentSession evidence must be one JSON object", { cause: error });
	}
	if (!isRecord(candidate)) throw new AgentRunnerError("AgentSession evidence must be a JSON object");

	const stringFields = ["runId", "laneId", "candidateHead", "validationNonce", "provider", "model", "thinking"] as const;
	for (const field of stringFields) {
		if (typeof candidate[field] !== "string") throw new AgentRunnerError(`AgentSession evidence field ${field} is invalid`);
	}
	if (!Number.isSafeInteger(candidate.generation) || Number(candidate.generation) < 1) {
		throw new AgentRunnerError("AgentSession evidence generation is invalid");
	}
	if (typeof candidate.readOnly !== "boolean" || typeof candidate.observedMutation !== "boolean") {
		throw new AgentRunnerError("AgentSession evidence mutation fields are invalid");
	}
	if (typeof candidate.summary !== "string" || candidate.summary.length === 0 ||
		candidate.summary.length > maxSummaryCharacters) {
		throw new AgentRunnerError("AgentSession evidence summary is empty or exceeds its limit");
	}
	if (!isRecord(candidate.dimensions)) throw new AgentRunnerError("AgentSession evidence dimensions are invalid");

	const dimensions = {} as DimensionScores;
	for (const name of DIMENSION_NAMES) {
		const value = candidate.dimensions[name];
		if (typeof value !== "number" || !Number.isFinite(value) || value < 0 || value > 1) {
			throw new AgentRunnerError(`AgentSession evidence dimension ${name} must be between 0 and 1`);
		}
		dimensions[name] = value;
	}

	return {
		runId: candidate.runId as string,
		generation: candidate.generation as number,
		laneId: candidate.laneId as string,
		candidateHead: candidate.candidateHead as string,
		validationNonce: candidate.validationNonce as string,
		readOnly: candidate.readOnly as boolean,
		provider: candidate.provider as string,
		model: candidate.model as string,
		thinking: candidate.thinking as AgentRunResult["thinking"],
		summary: candidate.summary,
		dimensions,
		observedMutation: candidate.observedMutation,
	};
}

function isAbsoluteNonTraversingPath(value: string): boolean {
	if (value.length === 0 || /[\u0000-\u001f\u007f]/.test(value)) return false;
	const pathFlavor = win32.isAbsolute(value) ? win32 : posix;
	if (!pathFlavor.isAbsolute(value)) return false;
	const segments = pathFlavor === win32 ? value.split(/[\\/]+/) : value.split("/");
	return !segments.includes("..");
}

class CleanupTimeoutError extends Error {
	constructor(step: string, timeoutMs: number) {
		super(`AgentSession cleanup ${step} timed out after ${timeoutMs}ms`);
		this.name = "CleanupTimeoutError";
	}
}

async function boundedCleanup<T>(operation: Promise<T>, timeoutMs: number, step: string): Promise<T> {
	return boundedCleanupUntil(operation, Date.now() + timeoutMs, step);
}

async function boundedCleanupUntil<T>(operation: Promise<T>, deadlineAt: number, step: string): Promise<T> {
	const timeoutMs = Math.max(0, deadlineAt - Date.now());
	let timer: ReturnType<typeof setTimeout> | undefined;
	const timeout = new Promise<never>((_resolve, reject) => {
		timer = setTimeout(() => reject(new CleanupTimeoutError(step, timeoutMs)), timeoutMs);
	});
	try {
		return await Promise.race([operation, timeout]);
	} finally {
		if (timer) clearTimeout(timer);
	}
}

function isCleanupTimeout(error: unknown): boolean {
	return error instanceof CleanupTimeoutError;
}

function byteLength(value: string): number {
	return new TextEncoder().encode(value).byteLength;
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null && !Array.isArray(value);
}

export { REQUIRED_PI_VERSION };
