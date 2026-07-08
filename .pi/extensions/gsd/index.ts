import { execFile } from "node:child_process";
import { readFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { promisify } from "node:util";
import { fileURLToPath } from "node:url";
import type { ExtensionAPI, ExtensionCommandContext } from "@earendil-works/pi-coding-agent";

const execFileAsync = promisify(execFile);
const baseDir = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(baseDir, "../../..");
const gsdScript = join(repoRoot, "scripts", "gsd");
const commandsPath = join(repoRoot, ".gsd", "commands.json");

interface GSDCommand {
	name: string;
	officialCommand: string;
	description?: string;
}

interface GSDRegistry {
	commands: GSDCommand[];
}

function loadCommands(): GSDCommand[] {
	try {
		const registry = JSON.parse(readFileSync(commandsPath, "utf8")) as GSDRegistry;
		return registry.commands ?? [];
	} catch {
		return [];
	}
}

async function runGSD(args: string[]): Promise<string> {
	const { stdout, stderr } = await execFileAsync(gsdScript, args, {
		cwd: repoRoot,
		maxBuffer: 10 * 1024 * 1024,
	});
	return [stdout, stderr].filter(Boolean).join("\n").trim();
}

function notifyOutput(ctx: ExtensionCommandContext, title: string, output: string): void {
	const text = output.length > 6000 ? `${output.slice(0, 6000)}\n… truncated …` : output;
	ctx.ui.notify(`${title}\n\n${text || "(no output)"}`, "info");
}

async function sendPrompt(command: string, args: string, ctx: ExtensionCommandContext, pi: ExtensionAPI): Promise<void> {
	if (!ctx.isIdle()) {
		ctx.ui.notify("Agent is busy. Queue the GSD command after the current turn finishes.", "warning");
		return;
	}
	const parts = args.trim() ? [command, ...args.trim().split(/\s+/)] : [command];
	const prompt = await runGSD(["prompt", ...parts]);
	pi.sendUserMessage(prompt);
}

export default function gsdPiExtension(pi: ExtensionAPI): void {
	const commands = loadCommands();
	const commandNames = commands.map((command) => command.name).sort();

	pi.registerCommand("gsd", {
		description: "Run a repo-local official GSD Core command through Pi",
		getArgumentCompletions: (prefix) => {
			const candidates = ["doctor", "list", "sources", "version", "verify-pi", ...commandNames];
			return candidates
				.filter((candidate) => candidate.startsWith(prefix))
				.slice(0, 30)
				.map((candidate) => ({ value: candidate, label: candidate }));
		},
		handler: async (args, ctx) => {
			const [subcommand = "help", ...rest] = args.trim().split(/\s+/).filter(Boolean);
			if (["doctor", "list", "version", "verify-pi"].includes(subcommand)) {
				const output = await runGSD([subcommand === "list" ? "list" : subcommand]);
				notifyOutput(ctx, `scripts/gsd ${subcommand}`, output);
				return;
			}
			if (subcommand === "sources") {
				if (!rest[0]) {
					ctx.ui.notify("Usage: /gsd sources <command>", "warning");
					return;
				}
				const output = await runGSD(["sources", rest[0]]);
				notifyOutput(ctx, `scripts/gsd sources ${rest[0]}`, output);
				return;
			}
			await sendPrompt(subcommand, rest.join(" "), ctx, pi);
		},
	});

	for (const command of commands) {
		const piCommand = command.officialCommand;
		pi.registerCommand(piCommand, {
			description: command.description || `Run /${piCommand}`,
			handler: async (args, ctx) => {
				await sendPrompt(command.name, args, ctx, pi);
			},
		});
	}

	pi.on("before_agent_start", (event) => {
		const prompt = event.prompt.toLowerCase();
		const shouldRemind = /\b(issue|implement|implementation|plan|planning|phase|milestone|roadmap|tdd|verify|connector parity)\b/.test(prompt);
		if (!shouldRemind) return;
		return {
			message: {
				customType: "gsd-core-pi-context",
				display: false,
				content: `This repository uses repo-local GSD Core through Pi. For planning or implementation work, prefer /gsd <command> or scripts/gsd prompt <command>. Before production edits, create/update the GSD plan, TDD ledger, and verification checklist. Respect AGENTS.md safety gates: no secrets, no new dependencies without approval, no credentialed connector checks unless requested, no raw generic write tools, and reverse ETL remains plan → preview → approval → execute.`,
			},
		};
	});
}
