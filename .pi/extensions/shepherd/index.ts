import { createHash } from "node:crypto";
import { join } from "node:path";

import {
	DefaultResourceLoader,
	ModelRuntime,
	SessionManager,
	SettingsManager,
	VERSION,
	createAgentSession,
	getAgentDir,
	type CreateAgentSessionOptions,
	type ExtensionAPI,
} from "@earendil-works/pi-coding-agent";

import type { AgentSessionRuntimeSdk } from "./agent-session-runtime.ts";
import { ShepherdController } from "./controller.ts";
import { assertShepherdPiCompatibility } from "./pi-compatibility.ts";
import {
	canonicalizeGitWorktree,
	registerShepherdExtension,
	type ShepherdCommandContext,
	type ShepherdExtensionHost,
} from "./extension.ts";
import {
	SdkAgentRunner,
	REQUIRED_PI_VERSION,
	type ShepherdModelRegistry,
	type ShepherdSdk,
} from "./sdk-runner.ts";
import { FileStateStore } from "./state-store.ts";
import { captureTargetEvidence } from "./target-evidence.ts";
import { createProductionPiHostController } from "./production-pi-host.ts";

function stateFingerprint(worktreeIdentity: string): string {
	return createHash("sha256").update(worktreeIdentity).digest("hex").slice(0, 24);
}

async function createEmbeddedAgentSession(
	modelRegistry: ShepherdModelRegistry,
	options: CreateAgentSessionOptions,
): ReturnType<typeof createAgentSession> {
	const agentDir = options.agentDir ?? getAgentDir();
	const modelRuntime = await ModelRuntime.create({ authPath: join(agentDir, "auth.json") });
	for (const provider of modelRegistry.getRegisteredProviderIds()) {
		const config = modelRegistry.getRegisteredProviderConfig(provider);
		if (config === undefined) continue;
		modelRuntime.unregisterProvider(provider);
		modelRuntime.registerProvider(provider, config as never);
	}
	const selected = options.model;
	if (selected) {
		const hostAuth = modelRegistry.getProviderAuthStatus(selected.provider);
		const childAuth = modelRuntime.getProviderAuthStatus(selected.provider);
		if (hostAuth.configured && !childAuth.configured && modelRegistry.isUsingOAuth(selected as never)) {
			throw new Error(`embedded AgentSession cannot inherit host-only OAuth for ${selected.provider}`);
		}
		if (hostAuth.source === "runtime" || (hostAuth.configured && !childAuth.configured)) {
			const apiKey = await modelRegistry.getApiKeyForProvider(selected.provider);
			if (apiKey !== undefined) await modelRuntime.setRuntimeApiKey(selected.provider, apiKey);
		}
	}
	return createAgentSession({ ...options, modelRuntime });
}

function embeddedSdk(modelRegistry: ShepherdModelRegistry): ShepherdSdk {
	return {
		version: VERSION,
		requiredVersion: REQUIRED_PI_VERSION,
		getAgentDir,
		createSettingsManager: (settings, options) =>
			SettingsManager.inMemory(settings as never, options as never),
		createSessionManager: (cwd) => SessionManager.inMemory(cwd),
		createResourceLoader: (options) => new DefaultResourceLoader(options as never),
		createSession: (options) =>
			createEmbeddedAgentSession(modelRegistry, options as CreateAgentSessionOptions) as unknown as
				ReturnType<ShepherdSdk["createSession"]>,
	};
}

function embeddedRuntimeSdk(modelRegistry: ShepherdModelRegistry): AgentSessionRuntimeSdk {
	return {
		version: VERSION,
		requiredVersion: REQUIRED_PI_VERSION,
		getAgentDir,
		findModel: (provider, model) => modelRegistry.find(provider, model) as never,
		hasConfiguredAuth: (model) => modelRegistry.hasConfiguredAuth(model as never),
		createSettingsManager: (settings, options) => SettingsManager.inMemory(settings as never, options as never),
		createSessionManager: (cwd) => SessionManager.inMemory(cwd),
		createResourceLoader: (options) => new DefaultResourceLoader(options as never),
		createAgentSession: (options) => createEmbeddedAgentSession(modelRegistry, options) as unknown as
			ReturnType<AgentSessionRuntimeSdk["createAgentSession"]>,
	};
}

export default function shepherdExtension(pi: ExtensionAPI): void {
	assertShepherdPiCompatibility(VERSION, REQUIRED_PI_VERSION);
	registerShepherdExtension(pi as unknown as ShepherdExtensionHost, {
		resolveWorktree: (context, options) => canonicalizeGitWorktree(context.cwd, options),
		createController(context: ShepherdCommandContext, worktree) {
			const root = join(
				getAgentDir(),
				"shepherd",
				stateFingerprint(worktree.worktreeIdentity),
			);
			const registry = context.modelRegistry as unknown as ShepherdModelRegistry;
			return new ShepherdController({
				store: new FileStateStore(root),
				runner: new SdkAgentRunner(embeddedSdk(registry), registry, { maxConcurrency: 2 }),
				targetEvidence: {
					capture: async (command) => {
						const current = await canonicalizeGitWorktree(context.cwd);
						if (current.repositoryIdentity !== worktree.repositoryIdentity
							|| current.worktreeIdentity !== worktree.worktreeIdentity) {
							throw new Error("Shepherd target repository/worktree identity changed");
						}
						return captureTargetEvidence({
							cwd: current.cwd,
							repositoryIdentity: current.repositoryIdentity,
							worktreeIdentity: current.worktreeIdentity,
							issue: command.issue,
							...(command.pr === undefined ? {} : { pr: command.pr }),
						});
					},
				},
			});
		},
		createAutonomousController(context: ShepherdCommandContext, worktree, issue) {
			const root = join(
				getAgentDir(),
				"shepherd",
				stateFingerprint(worktree.worktreeIdentity),
			);
			const registry = context.modelRegistry as ShepherdModelRegistry;
			const runtimeSdk = embeddedRuntimeSdk(registry);
			return createProductionPiHostController({
				issue,
				repositoryRoot: context.cwd,
				stateRoot: root,
				trustedWorktreeRoot: join(root, "worktrees"),
				runtimeSdk,
			});
		},
	});
}
