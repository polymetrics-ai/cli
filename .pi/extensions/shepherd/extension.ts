import { parseShepherdCommand, ShepherdArgumentError, type ShepherdCommand } from "./arguments.ts";
import type { ShepherdRunState } from "./domain.ts";

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

interface ExtensionDependencies {
	createController(context: ShepherdCommandContext): ShepherdControllerPort;
}

export interface ShepherdControllerPort {
	status(issue: number): Promise<ShepherdRunState | undefined>;
	start(command: Extract<ShepherdCommand, { action: "start" | "canary" }>): Promise<ShepherdRunState>;
	resume(command: Extract<ShepherdCommand, { action: "resume" }>): Promise<ShepherdRunState>;
	stop(issue: number): Promise<ShepherdRunState>;
	shutdown(): Promise<void>;
}

const HELP = [
	"Pi AgentSession Shepherd (experimental, interactive, read-only canary)",
	"",
	"Commands:",
	"  /pm-shepherd status --issue N",
	"  /pm-shepherd canary --issue N --pr N --read-only --backend sdk-inproc --experimental",
	"  /pm-shepherd start --issue N [--pr N] --read-only --backend sdk-inproc --experimental",
	"  /pm-shepherd resume --issue N [--pr N] --read-only --backend sdk-inproc --experimental",
	"  /pm-shepherd stop --issue N",
	"",
	"Embedded sessions share this Pi process, memory, environment, and crash domain.",
	"They stop if Pi exits; reopen Pi and use resume. No main merge or GitHub mutation is authorized.",
].join("\n");

function renderState(state: ShepherdRunState | undefined): string {
	if (!state) return "No persisted Shepherd run exists for this issue.";
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
	const controllers = new Map<number, ShepherdControllerPort>();
	let activeRun: { issue: number; promise: Promise<ShepherdRunState> } | undefined;
	let shuttingDown = false;

	const controllerFor = (issue: number, context: ShepherdCommandContext): ShepherdControllerPort => {
		const existing = controllers.get(issue);
		if (existing) return existing;
		const controller = dependencies.createController(context);
		controllers.set(issue, controller);
		return controller;
	};

	const launch = (
		command: Extract<ShepherdCommand, { action: "start" | "resume" | "canary" }>,
		context: ShepherdCommandContext,
	): void => {
		const issue = commandIssue(command);
		if (shuttingDown) {
			context.ui.notify("The Shepherd is shutting down; start a fresh Pi session before retrying.", "warning");
			return;
		}
		if (activeRun) {
			context.ui.notify(`A Shepherd run is already active for issue #${activeRun.issue}.`, "warning");
			return;
		}
		if (!context.isIdle()) {
			context.ui.notify("The parent Pi agent is busy. Retry after it becomes idle.", "warning");
			return;
		}
		const controller = controllerFor(issue, context);
		const statusKey = `pm-shepherd-${issue}`;
		context.ui.setStatus(statusKey, `issue #${issue}: starting`);
		const promise =
			command.action === "resume"
				? controller.resume(command)
				: controller.start(command);
		const ownedRun = { issue, promise };
		activeRun = ownedRun;
		context.ui.notify(
			`Embedded read-only Shepherd started for issue #${issue}. Use /pm-shepherd status --issue ${issue}.`,
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
				if (activeRun === ownedRun) activeRun = undefined;
				if (!shuttingDown) context.ui.setStatus(statusKey, undefined);
			});
	};

	host.registerCommand("pm-shepherd", {
		description: "Run the experimental in-process Pi AgentSession Shepherd",
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
				const controller = controllerFor(issue, context);
				context.ui.notify(renderState(await controller.status(issue)), "info");
				return;
			}
			if (command.action === "stop") {
				const controller = controllerFor(issue, context);
				context.ui.notify(renderState(await controller.stop(issue)), "info");
				return;
			}
			launch(command, context);
		},
	});

	host.on("session_shutdown", async () => {
		shuttingDown = true;
		const pending = activeRun?.promise;
		await Promise.allSettled([...controllers.values()].map((controller) => controller.shutdown()));
		if (pending) await Promise.allSettled([pending]);
		controllers.clear();
		activeRun = undefined;
	});
}
