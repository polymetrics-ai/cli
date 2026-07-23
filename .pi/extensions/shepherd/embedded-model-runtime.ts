export interface ModelAuthStatus {
	configured: boolean;
	source?: string;
}

export interface RegisteredProviderRuntime {
	unregisterProvider(provider: string): void;
	registerProvider(provider: string, config: unknown): void;
}

export interface RegisteredProviderSource {
	getRegisteredProviderIds(): readonly string[];
	getRegisteredProviderConfig(provider: string): unknown;
}

/** Apply the extension host's public provider registrations without resolving credential values. */
export function applyRegisteredProviderConfigs(
	runtime: RegisteredProviderRuntime,
	registry: RegisteredProviderSource,
): void {
	for (const provider of registry.getRegisteredProviderIds()) {
		const config = registry.getRegisteredProviderConfig(provider);
		if (config === undefined) continue;
		runtime.unregisterProvider(provider);
		runtime.registerProvider(provider, config);
	}
}

/**
 * Owns one lazy ModelRuntime initialization for one extension host.
 * A rejected initialization is forgotten so an explicit later attempt can retry.
 */
export class ExtensionModelRuntimeOwner<Runtime> {
	readonly #authorityAgentDir: string;
	#runtime: Promise<Runtime> | undefined;

	constructor(authorityAgentDir: string) {
		this.#authorityAgentDir = authorityAgentDir;
	}

	acquire(requestedAgentDir: string, initialize: () => Promise<Runtime>): Promise<Runtime> {
		if (requestedAgentDir !== this.#authorityAgentDir) {
			throw new Error("embedded AgentSession agentDir does not match extension host authority");
		}
		if (!this.#runtime) {
			const runtime = Promise.resolve().then(initialize);
			this.#runtime = runtime;
			void runtime.catch(() => {
				if (this.#runtime === runtime) this.#runtime = undefined;
			});
		}
		return this.#runtime;
	}
}

/** Fail closed unless the embedded runtime can resolve auth from its own normal credential store. */
export function assertEmbeddedModelAuth(
	provider: string,
	usesOAuth: boolean,
	hostAuth: ModelAuthStatus,
	embeddedAuth: ModelAuthStatus,
): void {
	if (embeddedAuth.configured) return;
	if (hostAuth.configured && usesOAuth) {
		throw new Error(`embedded AgentSession cannot inherit host-only OAuth for ${provider}`);
	}
	if (hostAuth.configured) {
		throw new Error(`embedded AgentSession cannot inherit host-only auth for ${provider}`);
	}
	throw new Error(`embedded AgentSession has no configured auth for ${provider}`);
}
