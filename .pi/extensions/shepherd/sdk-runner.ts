import type {
	AgentRunner,
	AgentRunRequest,
	AgentRunResult,
	DimensionScores,
} from "./runner.ts";

const REQUIRED_PI_VERSION = "0.80.6";
const REQUIRED_PROVIDER = "openai-codex";
const REQUIRED_MODEL = "gpt-5.6-sol";
const READ_ONLY_TOOLS = new Set(["read", "grep", "find", "ls"]);
const DIMENSION_NAMES = [
	"correctStage",
	"artifactValid",
	"gatesRespected",
	"realProgress",
	"noHallucination",
	"noConflict",
] as const;

type SessionModel = { provider: string; id: string };

interface EmbeddedSession {
	model: SessionModel;
	thinkingLevel: string;
	sessionFile: string | undefined;
	getActiveToolNames(): string[];
	subscribe(listener: (event: unknown) => void): () => void;
	prompt(prompt: string, options: { expandPromptTemplates: false; source: "extension" }): Promise<void>;
	waitForIdle(): Promise<void>;
	abort(): Promise<void>;
	dispose(): void;
	getLastAssistantText(): string | undefined;
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

	constructor(runId: string, laneKey: string, session: EmbeddedSession) {
		this.runId = runId;
		this.laneKey = laneKey;
		this.session = session;
	}

	abortOnce(): Promise<void> {
		if (!this.#abortPromise) {
			this.#abortPromise = Promise.resolve().then(() => this.session.abort());
		}
		return this.#abortPromise;
	}

	cleanup(): Promise<void> {
		if (!this.#cleanupPromise) {
			this.#cleanupPromise = this.#cleanup();
		}
		return this.#cleanupPromise;
	}

	async #cleanup(): Promise<void> {
		let firstError: unknown;
		try {
			await this.abortOnce();
		} catch (error) {
			firstError = error;
		}
		try {
			await this.session.waitForIdle();
		} catch (error) {
			firstError ??= error;
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
			throw new AgentRunnerError("AgentSession cleanup failed", { cause: firstError });
		}
	}
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
	#activeCount = 0;
	#activeMutator = false;
	#closed = false;

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
		let child: OwnedChild | undefined;
		let primaryFailed = false;

		try {
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
			await resourceLoader.reload();
			this.#assertOpen();

			const created = await this.#sdk.createSession({
				cwd: request.cwd,
				agentDir: this.#sdk.getAgentDir(),
				authStorage: this.#modelRegistry.authStorage,
				modelRegistry: this.#modelRegistry,
				model,
				thinkingLevel: request.thinking,
				tools: [...request.tools],
				resourceLoader,
				sessionManager,
				settingsManager,
			});
			child = new OwnedChild(request.runId, laneKey, created.session);
			this.#register(child);
			this.#assertOpen();
			validateCreatedSession(created, request);

			let eventCount = 0;
			let eventBytes = 0;
			let eventFailure: AgentRunnerError | undefined;
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
				if (eventFailure) void child?.abortOnce().catch(() => undefined);
			});

			await withTimeout(
				created.session.prompt(request.prompt, {
					expandPromptTemplates: false,
					source: "extension",
				}),
				request.timeoutMs,
				() => child?.abortOnce(),
			);
			if (eventFailure) throw eventFailure;

			const assistantText = created.session.getLastAssistantText();
			return parseEvidence(assistantText, this.#options.maxAssistantBytes, this.#options.maxSummaryCharacters);
		} catch (error) {
			primaryFailed = true;
			throw error;
		} finally {
			try {
				await child?.cleanup();
			} catch (error) {
				if (!primaryFailed) throw error;
			} finally {
				if (child) this.#unregister(child);
				this.#release(request, laneKey);
			}
		}
	}

	async abort(runId: string): Promise<void> {
		const children = [...(this.#children.get(runId) ?? [])];
		await Promise.all(children.map((child) => child.abortOnce()));
	}

	async close(): Promise<void> {
		if (this.#closed) return;
		this.#closed = true;
		const children = [...this.#children.values()].flatMap((entries) => [...entries]);
		const results = await Promise.allSettled(children.map((child) => child.cleanup()));
		const failure = results.find((result) => result.status === "rejected");
		if (failure?.status === "rejected") {
			throw new AgentRunnerError("one or more AgentSessions failed to close", { cause: failure.reason });
		}
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
		if (this.#closed) throw new AgentRunnerError("AgentSession runner is closed");
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
		this.#activeMutator ||= !request.readOnly;
		this.#activeLaneKeys.add(laneKey);
	}

	#release(request: AgentRunRequest, laneKey: string): void {
		if (!this.#activeLaneKeys.delete(laneKey)) return;
		this.#activeCount -= 1;
		if (!request.readOnly) this.#activeMutator = false;
	}

	#register(child: OwnedChild): void {
		const children = this.#children.get(child.runId) ?? new Set<OwnedChild>();
		children.add(child);
		this.#children.set(child.runId, children);
	}

	#unregister(child: OwnedChild): void {
		const children = this.#children.get(child.runId);
		if (!children) return;
		children.delete(child);
		if (children.size === 0) this.#children.delete(child.runId);
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
	if (!request.cwd.startsWith("/") || /[\u0000-\u001f\u007f]/.test(request.cwd) ||
		request.cwd.split("/").includes("..")) {
		throw new AgentRunnerError("cwd must be an absolute non-traversing path without control characters");
	}
	if (request.systemPrompt.length === 0 || request.systemPrompt.length > 32 * 1024 ||
		request.prompt.length === 0 || request.prompt.length > 64 * 1024) {
		throw new AgentRunnerError("AgentSession prompts must be non-empty and bounded");
	}
	if (request.tools.length === 0 || new Set(request.tools).size !== request.tools.length ||
		request.tools.some((tool) => !/^[a-z][a-z0-9_-]{0,63}$/.test(tool))) {
		throw new AgentRunnerError("tool allowlist is invalid");
	}
	if (request.readOnly && request.tools.some((tool) => !READ_ONLY_TOOLS.has(tool))) {
		throw new AgentRunnerError("read-only lane requested a mutating or unknown tool");
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
		typeof session.dispose !== "function" || typeof session.getLastAssistantText !== "function" ||
		typeof session.getActiveToolNames !== "function") {
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
	if (!sameStringSet(session.getActiveToolNames(), request.tools)) {
		throw new AgentRunnerError("embedded AgentSession tool allowlist mismatch");
	}
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

async function withTimeout<T>(
	operation: Promise<T>,
	timeoutMs: number,
	onTimeout: () => Promise<void> | undefined,
): Promise<T> {
	let timer: ReturnType<typeof setTimeout> | undefined;
	const timeout = new Promise<never>((_resolve, reject) => {
		timer = setTimeout(() => {
			void onTimeout()?.catch(() => undefined);
			reject(new AgentRunnerError(`AgentSession timed out after ${timeoutMs}ms`));
		}, timeoutMs);
	});
	try {
		return await Promise.race([operation, timeout]);
	} finally {
		if (timer) clearTimeout(timer);
	}
}

function sameStringSet(left: string[], right: string[]): boolean {
	if (!Array.isArray(left) || left.length !== right.length) return false;
	if (new Set(left).size !== left.length) return false;
	const expected = new Set(right);
	return left.every((value) => expected.has(value));
}

function byteLength(value: string): number {
	return new TextEncoder().encode(value).byteLength;
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null && !Array.isArray(value);
}

export { REQUIRED_PI_VERSION };
