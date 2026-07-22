import { parseShepherdCommand, ShepherdArgumentError, type ShepherdCommand } from "./arguments.ts";
import type { AutonomousShepherdRunState } from "./autonomous-state.ts";
import type { ShepherdRunState } from "./domain.ts";
import {
	resolveCanonicalGitWorktree,
	type CanonicalGitWorktree,
	type CanonicalGitWorktreeOptions,
} from "./target-evidence.ts";

interface CommandUi {
	notify(message: string, level: "info" | "warning" | "error"): void;
	setStatus(key: string, value: string | undefined): void;
}

export interface ShepherdCommandContext {
	cwd: string;
	modelRegistry: unknown;
	isIdle(): boolean;
	ui: CommandUi;
}

interface CommandDefinition {
	description: string;
	getArgumentCompletions?: (prefix: string) => Array<{ value: string; label: string }>;
	handler(args: string, context: ShepherdCommandContext): Promise<void>;
}

export interface ShepherdExtensionHost {
	registerCommand(name: string, definition: CommandDefinition): void;
	on(event: "session_shutdown", handler: () => Promise<void>): void;
}

export type CanonicalWorktree = CanonicalGitWorktree;

interface ExtensionDependencies {
	resolveWorktree(context: ShepherdCommandContext, options: CanonicalGitWorktreeOptions): Promise<CanonicalWorktree>;
	/** Hardened schema-v1 controller retained only for the explicit read-only canary. */
	createController(context: ShepherdCommandContext, worktree: CanonicalWorktree): ShepherdControllerPort;
	/** Autonomous schema-v2 controller. Optional only for compatibility with old extension harnesses. */
	createAutonomousController?(context: ShepherdCommandContext, worktree: CanonicalWorktree): AutonomousShepherdControllerPort;
}

export interface ShepherdControllerPort {
	status(issue: number): Promise<ShepherdRunState | undefined>;
	start(command: Extract<ShepherdCommand, { action: "canary" }>): Promise<ShepherdRunState>;
	stop(issue: number): Promise<ShepherdRunState>;
	shutdown(): Promise<void>;
}

export interface AutonomousShepherdControllerPort {
	status(issue: number): Promise<AutonomousShepherdRunState | undefined>;
	start(command: Extract<ShepherdCommand, { action: "start" }>): Promise<AutonomousShepherdRunState>;
	resume(command: Extract<ShepherdCommand, { action: "resume" }>): Promise<AutonomousShepherdRunState>;
	stop(issue: number): Promise<AutonomousShepherdRunState>;
	shutdown(): Promise<void>;
}

type DisplayState = ShepherdRunState | AutonomousShepherdRunState;
type ControllerEntry = ShepherdControllerPort | AutonomousShepherdControllerPort;

interface LaunchSetup {
	issue: number;
	abortController: AbortController;
	promise: Promise<void>;
}

interface ActiveExtensionRun {
	issue: number;
	mode: "autonomous" | "canary";
	controller: ControllerEntry;
	promise: Promise<DisplayState>;
}

const HELP = [
	"Pi AgentSession Shepherd (autonomous MVP plus explicit read-only canary)",
	"",
	"Commands:",
	"  /pm-shepherd status --issue N",
	"  /pm-shepherd canary --issue N --pr N --read-only --backend sdk-inproc --experimental",
	"  /pm-shepherd start --issue N [--pr N] [--backend sdk-inproc]",
	"  /pm-shepherd resume --issue N [--pr N] [--backend sdk-inproc]",
	"  /pm-shepherd stop --issue N",
	"",
	"Embedded sessions share this Pi process, memory, environment, and crash domain.",
	"They stop if Pi exits; reopen Pi and use resume. Parent-to-main merge remains human-owned.",
].join("\n");
const SHUTDOWN_TIMEOUT_MS = 45_000;

async function settleBeforeShutdown(promise: Promise<unknown>): Promise<void> {
	let timer: ReturnType<typeof setTimeout> | undefined;
	const timeout = new Promise<never>((_resolve, reject) => {
		timer = setTimeout(() => reject(new Error("Shepherd shutdown deadline exceeded")), SHUTDOWN_TIMEOUT_MS);
	});
	try {
		await Promise.race([promise, timeout]);
	} finally {
		if (timer) clearTimeout(timer);
	}
}

export function canonicalizeGitWorktree(
	cwd: string,
	options: CanonicalGitWorktreeOptions = {},
): Promise<CanonicalWorktree> {
	return resolveCanonicalGitWorktree(cwd, options);
}

function isAutonomousState(state: DisplayState): state is AutonomousShepherdRunState {
	return "kind" in state && state.kind === "autonomous";
}

function renderState(state: DisplayState | undefined): string {
	if (!state) return "No persisted Shepherd run exists for this issue.";
	if (isAutonomousState(state)) {
		const lines = [
			`Issue #${state.issue}${state.pr ? ` / PR #${state.pr}` : ""}`,
			`run=${state.runId} generation=${state.generation} status=${state.status} stage=${state.stage}`,
		];
		for (const child of state.children) {
			lines.push(`${child.id}: ${child.status}/${child.phase}${child.summary ? ` — ${child.summary}` : ""}`);
		}
		if (state.humanGate) lines.push(`human_gate=${state.humanGate.kind}:${state.humanGate.status}`);
		return lines.join("\n");
	}
	const lines = [
		`Issue #${state.issue}${state.pr ? ` / PR #${state.pr}` : ""}`,
		`run=${state.runId} generation=${state.generation} status=${state.status}`,
		`head=${state.candidateHead.slice(0, 12)} score=${state.score === undefined ? "pending" : state.score.toFixed(3)}`,
	];
	for (const lane of state.lanes) {
		lines.push(
			`${lane.id}: ${lane.status}${lane.score === undefined ? "" : ` score=${lane.score.toFixed(3)}`}${lane.summary ? ` — ${lane.summary}` : ""}`,
		);
	}
	if (state.hardGates?.length) lines.push(`hard_gates=${state.hardGates.join(",")}`);
	return lines.join("\n");
}

function commandIssue(command: ShepherdCommand): number {
	if (!("issue" in command) || !command.issue) {
		throw new ShepherdArgumentError("This action requires --issue N");
	}
	return command.issue;
}

export function registerShepherdExtension(
	host: ShepherdExtensionHost,
	dependencies: ExtensionDependencies,
): void {
	const controllers = new Map<string, ControllerEntry>();
	let activeRun: ActiveExtensionRun | undefined;
	let launchSetup: LaunchSetup | undefined;
	let shuttingDown = false;

	const controllerFor = async (
		mode: "autonomous" | "canary",
		issue: number,
		context: ShepherdCommandContext,
		options: CanonicalGitWorktreeOptions = {},
	): Promise<ControllerEntry> => {
		const worktree = await dependencies.resolveWorktree(context, options);
		const key = `${mode}\u0000${worktree.repositoryIdentity}\u0000${worktree.worktreeIdentity}\u0000${issue}`;
		const existing = controllers.get(key);
		if (existing) return existing;
		const canonicalContext = { ...context, cwd: worktree.cwd };
		const controller = mode === "canary" || dependencies.createAutonomousController === undefined
			? dependencies.createController(canonicalContext, worktree)
			: dependencies.createAutonomousController(canonicalContext, worktree);
		controllers.set(key, controller);
		return controller;
	};

	const launch = async (
		command: Extract<ShepherdCommand, { action: "start" | "resume" | "canary" }>,
		context: ShepherdCommandContext,
	): Promise<void> => {
		const issue = commandIssue(command);
		if (shuttingDown) {
			context.ui.notify("The Shepherd is shutting down; start a fresh Pi session before retrying.", "warning");
			return;
		}
		if (activeRun || launchSetup) {
			context.ui.notify(`A Shepherd run is already active for issue #${activeRun?.issue ?? launchSetup?.issue}.`, "warning");
			return;
		}
		if (!context.isIdle()) {
			context.ui.notify("The parent Pi agent is busy. Retry after it becomes idle.", "warning");
			return;
		}
		const abortController = new AbortController();
		let ownedSetup: LaunchSetup;
		const setupPromise = (async () => {
			let controller: ControllerEntry;
			try {
				controller = await controllerFor(
					command.action === "canary" ? "canary" : "autonomous",
					issue,
					context,
					{ signal: abortController.signal },
				);
			} catch (error) {
				if (shuttingDown && abortController.signal.aborted) return;
				throw error;
			}
			if (shuttingDown || abortController.signal.aborted) return;
			const statusKey = `pm-shepherd-${issue}`;
			context.ui.setStatus(statusKey, `issue #${issue}: starting`);
			const promise: Promise<DisplayState> = command.action === "canary"
				? (controller as ShepherdControllerPort).start(command)
				: command.action === "resume"
					? (controller as AutonomousShepherdControllerPort).resume(command)
					: (controller as AutonomousShepherdControllerPort).start(command);
			const ownedRun: ActiveExtensionRun = {
				issue,
				mode: command.action === "canary" ? "canary" : "autonomous",
				controller,
				promise,
			};
			activeRun = ownedRun;
			context.ui.notify(
				`${command.action === "canary" ? "Embedded read-only canary" : "Autonomous in-process Shepherd"} started for issue #${issue}. Use /pm-shepherd status --issue ${issue}.`,
				"info",
			);
			void promise
				.then((state) => {
					if (!shuttingDown) {
						context.ui.notify(renderState(state), state.status === "completed" ? "info" : "warning");
					}
				})
				.catch((error) => {
					if (!shuttingDown) context.ui.notify(error instanceof Error ? error.message : String(error), "error");
				})
				.finally(() => {
					if (!shuttingDown && activeRun === ownedRun) activeRun = undefined;
					if (!shuttingDown) context.ui.setStatus(statusKey, undefined);
				});
		})();
		ownedSetup = { issue, abortController, promise: setupPromise };
		launchSetup = ownedSetup;
		try {
			await setupPromise;
		} finally {
			if (launchSetup === ownedSetup) launchSetup = undefined;
		}
	};

	host.registerCommand("pm-shepherd", {
		description: "Run the in-process autonomous Shepherd or its read-only canary",
		getArgumentCompletions: (prefix) =>
			["help", "status", "canary", "start", "resume", "stop"]
				.filter((candidate) => candidate.startsWith(prefix))
				.map((candidate) => ({ value: candidate, label: candidate })),
		handler: async (args, context) => {
			let command: ShepherdCommand;
			try {
				command = parseShepherdCommand(args);
			} catch (error) {
				context.ui.notify(error instanceof Error ? error.message : String(error), "warning");
				return;
			}

			if (command.action === "help") {
				context.ui.notify(HELP, "info");
				return;
			}

			const issue = commandIssue(command);
			if (command.action === "status") {
				const running = activeRun?.issue === issue ? activeRun : undefined;
				if (running?.mode === "canary") {
					context.ui.notify(renderState(await (running.controller as ShepherdControllerPort).status(issue)), "info");
				} else if (running?.mode === "autonomous") {
					context.ui.notify(
						renderState(await (running.controller as AutonomousShepherdControllerPort).status(issue)),
						"info",
					);
				} else {
					const autonomous = await controllerFor("autonomous", issue, context) as AutonomousShepherdControllerPort;
					const autonomousState = await autonomous.status(issue);
					if (autonomousState || dependencies.createAutonomousController === undefined) {
						context.ui.notify(renderState(autonomousState), "info");
					} else {
						const canary = await controllerFor("canary", issue, context) as ShepherdControllerPort;
						context.ui.notify(renderState(await canary.status(issue)), "info");
					}
				}
				return;
			}
			if (command.action === "stop") {
				const running = activeRun?.issue === issue ? activeRun : undefined;
				if (running?.mode === "canary") {
					context.ui.notify(renderState(await (running.controller as ShepherdControllerPort).stop(issue)), "info");
				} else if (running?.mode === "autonomous") {
					context.ui.notify(renderState(await (running.controller as AutonomousShepherdControllerPort).stop(issue)), "info");
				} else {
					const autonomous = await controllerFor("autonomous", issue, context) as AutonomousShepherdControllerPort;
					const autonomousState = await autonomous.status(issue);
					if (autonomousState || dependencies.createAutonomousController === undefined) {
						context.ui.notify(renderState(await autonomous.stop(issue)), "info");
					} else {
						const canary = await controllerFor("canary", issue, context) as ShepherdControllerPort;
						context.ui.notify(renderState(await canary.stop(issue)), "info");
					}
				}
				return;
			}
			await launch(command, context);
		},
	});

	host.on("session_shutdown", async () => {
		shuttingDown = true;
		const setup = launchSetup;
		setup?.abortController.abort(new Error("Shepherd extension is shutting down"));
		await settleBeforeShutdown((async () => {
			if (setup) await setup.promise;
			const pending = activeRun?.promise;
			const results = await Promise.allSettled([
				...[...controllers.values()].map((controller) => controller.shutdown()),
				...(pending ? [pending] : []),
			]);
			const failures = results
				.filter((result): result is PromiseRejectedResult => result.status === "rejected")
				.map((result) => result.reason);
			if (failures.length > 0) {
				throw new AggregateError(failures, "Shepherd extension shutdown failed");
			}
		})());
		controllers.clear();
		activeRun = undefined;
	});
}
