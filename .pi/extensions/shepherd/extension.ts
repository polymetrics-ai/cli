import { execFile } from "node:child_process";
import { realpath } from "node:fs/promises";
import { isAbsolute } from "node:path";
import { promisify } from "node:util";

import { parseShepherdCommand, ShepherdArgumentError, type ShepherdCommand } from "./arguments.ts";
import type { ShepherdRunState } from "./domain.ts";

const execFileAsync = promisify(execFile);

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

export interface CanonicalWorktree {
	cwd: string;
	identity: string;
}

interface ExtensionDependencies {
	resolveWorktree(context: ShepherdCommandContext): Promise<CanonicalWorktree>;
	createController(context: ShepherdCommandContext, worktree: CanonicalWorktree): ShepherdControllerPort;
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

export async function canonicalizeGitWorktree(cwd: string): Promise<CanonicalWorktree> {
	if (typeof cwd !== "string" || cwd.length === 0 || cwd.length > 4_096 || /[\u0000-\u001f\u007f]/.test(cwd)) {
		throw new Error("Shepherd cwd must be a bounded path without control characters");
	}
	const result = await execFileAsync(
		"git",
		["-C", cwd, "rev-parse", "--show-toplevel"],
		{ encoding: "utf8", maxBuffer: 64 * 1024 },
	);
	const topLevel = String(result.stdout).trim();
	if (!isAbsolute(topLevel) || topLevel.length > 4_096 || /[\u0000-\u001f\u007f]/.test(topLevel)) {
		throw new Error("Git returned an invalid Shepherd worktree root");
	}
	const canonicalCwd = await realpath(topLevel);
	if (!isAbsolute(canonicalCwd) || canonicalCwd.length > 4_096 || /[\u0000-\u001f\u007f]/.test(canonicalCwd)) {
		throw new Error("Git worktree resolves to an invalid canonical path");
	}
	return { cwd: canonicalCwd, identity: canonicalCwd };
}

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
	const controllers = new Map<string, ShepherdControllerPort>();
	let activeRun: { issue: number; promise: Promise<ShepherdRunState> } | undefined;
	let launchingIssue: number | undefined;
	let shuttingDown = false;

	const controllerFor = async (issue: number, context: ShepherdCommandContext): Promise<ShepherdControllerPort> => {
		const worktree = await dependencies.resolveWorktree(context);
		const key = `${worktree.identity}\u0000${issue}`;
		const existing = controllers.get(key);
		if (existing) return existing;
		const canonicalContext = { ...context, cwd: worktree.cwd };
		const controller = dependencies.createController(canonicalContext, worktree);
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
		if (activeRun || launchingIssue !== undefined) {
			context.ui.notify(`A Shepherd run is already active for issue #${activeRun?.issue ?? launchingIssue}.`, "warning");
			return;
		}
		if (!context.isIdle()) {
			context.ui.notify("The parent Pi agent is busy. Retry after it becomes idle.", "warning");
			return;
		}
		launchingIssue = issue;
		let controller: ShepherdControllerPort;
		try {
			controller = await controllerFor(issue, context);
		} finally {
			if (launchingIssue === issue) launchingIssue = undefined;
		}
		if (shuttingDown) {
			context.ui.notify("The Shepherd is shutting down; start a fresh Pi session before retrying.", "warning");
			return;
		}
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
				if (!shuttingDown && activeRun === ownedRun) activeRun = undefined;
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
				const controller = await controllerFor(issue, context);
				context.ui.notify(renderState(await controller.status(issue)), "info");
				return;
			}
			if (command.action === "stop") {
				const controller = await controllerFor(issue, context);
				context.ui.notify(renderState(await controller.stop(issue)), "info");
				return;
			}
			await launch(command, context);
		},
	});

	host.on("session_shutdown", async () => {
		shuttingDown = true;
		const pending = activeRun?.promise;
		await settleBeforeShutdown((async () => {
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
