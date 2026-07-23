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
import {
	applyRegisteredProviderConfigs,
	assertEmbeddedModelAuth,
	ExtensionModelRuntimeOwner,
	type RegisteredProviderRuntime,
} from "./embedded-model-runtime.ts";
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
	modelRuntimeOwner: ExtensionModelRuntimeOwner<ModelRuntime>,
	modelRegistry: ShepherdModelRegistry,
	options: CreateAgentSessionOptions,
): ReturnType<typeof createAgentSession> {
	const agentDir = options.agentDir ?? getAgentDir();
	const modelRuntime = await modelRuntimeOwner.acquire(
		agentDir,
		() => ModelRuntime.create({ authPath: join(agentDir, "auth.json") }),
	);
	applyRegisteredProviderConfigs(
		modelRuntime as unknown as RegisteredProviderRuntime,
		modelRegistry,
	);
	const selected = options.model;
	if (selected) {
		assertEmbeddedModelAuth(
			selected.provider,
			modelRegistry.isUsingOAuth(selected as never),
			modelRegistry.getProviderAuthStatus(selected.provider),
			modelRuntime.getProviderAuthStatus(selected.provider),
		);
	}
	return createAgentSession({ ...options, modelRuntime });
}

function embeddedSdk(
	modelRuntimeOwner: ExtensionModelRuntimeOwner<ModelRuntime>,
	modelRegistry: ShepherdModelRegistry,
): ShepherdSdk {
	return {
		version: VERSION,
		requiredVersion: REQUIRED_PI_VERSION,
		getAgentDir,
		createSettingsManager: (settings, options) =>
			SettingsManager.inMemory(settings as never, options as never),
		createSessionManager: (cwd) => SessionManager.inMemory(cwd),
		createResourceLoader: (options) => new DefaultResourceLoader(options as never),
		createSession: (options) =>
			createEmbeddedAgentSession(
				modelRuntimeOwner,
				modelRegistry,
				options as CreateAgentSessionOptions,
			) as unknown as ReturnType<ShepherdSdk["createSession"]>,
	};
}

function embeddedRuntimeSdk(
	modelRuntimeOwner: ExtensionModelRuntimeOwner<ModelRuntime>,
	modelRegistry: ShepherdModelRegistry,
): AgentSessionRuntimeSdk {
	return {
		version: VERSION,
		requiredVersion: REQUIRED_PI_VERSION,
		getAgentDir,
		findModel: (provider, model) => modelRegistry.find(provider, model) as never,
		hasConfiguredAuth: (model) => modelRegistry.hasConfiguredAuth(model as never),
		createSettingsManager: (settings, options) => SettingsManager.inMemory(settings as never, options as never),
		createSessionManager: (cwd) => SessionManager.inMemory(cwd),
		createResourceLoader: (options) => new DefaultResourceLoader(options as never),
		createAgentSession: (options) => createEmbeddedAgentSession(
			modelRuntimeOwner,
			modelRegistry,
			options,
		) as unknown as ReturnType<AgentSessionRuntimeSdk["createAgentSession"]>,
	};
}

export default function shepherdExtension(pi: ExtensionAPI): void {
	assertShepherdPiCompatibility(VERSION, REQUIRED_PI_VERSION);
	const agentDir = getAgentDir();
	const modelRuntimeOwner = new ExtensionModelRuntimeOwner<ModelRuntime>(agentDir);
	registerShepherdExtension(pi as unknown as ShepherdExtensionHost, {
		resolveWorktree: (context, options) => canonicalizeGitWorktree(context.cwd, options),
		createController(context: ShepherdCommandContext, worktree) {
			const root = join(
				agentDir,
				"shepherd",
				stateFingerprint(worktree.worktreeIdentity),
			);
			const registry = context.modelRegistry as unknown as ShepherdModelRegistry;
			return new ShepherdController({
				store: new FileStateStore(root),
				runner: new SdkAgentRunner(
					embeddedSdk(modelRuntimeOwner, registry),
					registry,
					{ maxConcurrency: 2 },
				),
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
				agentDir,
				"shepherd",
				stateFingerprint(worktree.worktreeIdentity),
			);
			const registry = context.modelRegistry as ShepherdModelRegistry;
			const runtimeSdk = embeddedRuntimeSdk(modelRuntimeOwner, registry);
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
