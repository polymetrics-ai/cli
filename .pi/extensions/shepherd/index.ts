import { createHash } from "node:crypto";
import { join } from "node:path";

import {
	DefaultResourceLoader,
	SessionManager,
	SettingsManager,
	VERSION,
	createAgentSession,
	getAgentDir,
	type CreateAgentSessionOptions,
	type ExtensionAPI,
} from "@earendil-works/pi-coding-agent";

import { ShepherdAgentSessionRuntime, type AgentSessionRuntimeSdk } from "./agent-session-runtime.ts";
import { AutonomousShepherdController } from "./autonomous-controller.ts";
import {
	AgentSessionMvpLifecycle,
	LocalParentMergeGate,
	RepositoryManifestIntake,
} from "./autonomous-local-adapters.ts";
import { AutonomousFileStateStore } from "./autonomous-state.ts";
import { ShepherdController } from "./controller.ts";
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

function stateFingerprint(worktreeIdentity: string): string {
	return createHash("sha256").update(worktreeIdentity).digest("hex").slice(0, 24);
}

function embeddedSdk(): ShepherdSdk {
	return {
		version: VERSION,
		requiredVersion: REQUIRED_PI_VERSION,
		getAgentDir,
		createSettingsManager: (settings, options) =>
			SettingsManager.inMemory(settings as never, options as never),
		createSessionManager: (cwd) => SessionManager.inMemory(cwd),
		createResourceLoader: (options) => new DefaultResourceLoader(options as never),
		createSession: (options) =>
			createAgentSession(options as CreateAgentSessionOptions) as unknown as ReturnType<ShepherdSdk["createSession"]>,
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
		createAgentSession: (options) => createAgentSession({
			...options,
			authStorage: modelRegistry.authStorage as never,
			modelRegistry: modelRegistry as never,
		} as CreateAgentSessionOptions) as unknown as ReturnType<AgentSessionRuntimeSdk["createAgentSession"]>,
	};
}

export default function shepherdExtension(pi: ExtensionAPI): void {
	const sdk = embeddedSdk();
	registerShepherdExtension(pi as unknown as ShepherdExtensionHost, {
		resolveWorktree: (context, options) => canonicalizeGitWorktree(context.cwd, options),
		createController(context: ShepherdCommandContext, worktree) {
			const root = join(
				getAgentDir(),
				"shepherd",
				stateFingerprint(worktree.worktreeIdentity),
			);
			return new ShepherdController({
				store: new FileStateStore(root),
				runner: new SdkAgentRunner(sdk, context.modelRegistry as never, { maxConcurrency: 2 }),
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
		createAutonomousController(context: ShepherdCommandContext, worktree) {
			const root = join(
				getAgentDir(),
				"shepherd",
				stateFingerprint(worktree.worktreeIdentity),
			);
			const registry = context.modelRegistry as ShepherdModelRegistry;
			const runtimeSdk = embeddedRuntimeSdk(registry);
			return new AutonomousShepherdController({
				store: new AutonomousFileStateStore(root),
				intake: new RepositoryManifestIntake(context.cwd),
				lifecycle: new AgentSessionMvpLifecycle(
					() => new ShepherdAgentSessionRuntime(runtimeSdk, { maxConcurrency: 1 }),
					context.cwd,
				),
				humanGate: new LocalParentMergeGate(),
			});
		},
	});
}
