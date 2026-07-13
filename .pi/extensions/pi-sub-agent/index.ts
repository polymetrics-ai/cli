/**
 * Subagent Tool - Delegate tasks to specialized agents
 *
 * Spawns a separate `pi` process for each subagent invocation, giving it an
 * isolated context window.
 */

import { spawn } from "node:child_process";
import * as fs from "node:fs";
import { basename, dirname, isAbsolute, join, resolve, sep } from "node:path";
import * as os from "node:os";
import { fileURLToPath } from "node:url";
import { StringEnum } from "@earendil-works/pi-ai";
import type { ExtensionAPI, ThemeColor } from "@earendil-works/pi-coding-agent";
import {
	DEFAULT_MAX_BYTES,
	DEFAULT_MAX_LINES,
	formatSize,
	getAgentDir,
	getMarkdownTheme,
	getSettingsListTheme,
	keyText,
	truncateTail,
	withFileMutationQueue,
} from "@earendil-works/pi-coding-agent";
import { Container, Markdown, SelectList, Spacer, type SettingItem, SettingsList, Text } from "@earendil-works/pi-tui";
import { Type } from "typebox";
import {
	type AgentConfig,
	type AgentScope,
	type AgentThinkingLevel,
	THINKING_LEVELS,
	discoverAgents,
	resolveAgentModel,
	updateAgentSettingsContent,
} from "./agents.js";

const extensionDir = dirname(fileURLToPath(import.meta.url));
const MAX_PARALLEL_TASKS = 8;
const MAX_CHAIN_STEPS = 8;
const MAX_CONCURRENCY = 4;
const COLLAPSED_ITEM_COUNT = 10;
const SUBAGENT_TOOL_NAME = "subagent";
const BUNDLED_AGENT_NAMES = "scout, planner, worker, reviewer, debugger, verifier, security-auditor, docs-writer, refactorer";
const BUNDLED_AGENT_SELECTION_GUIDANCE = [
	"Use scout for codebase recon.",
	"Use planner for implementation plans.",
	"Use worker for general implementation.",
	"Use reviewer for code quality review.",
	"Use debugger for root-cause investigation.",
	"Use verifier for running checks.",
	"Use security-auditor for security review.",
	"Use docs-writer for documentation.",
	"Use refactorer for behavior-preserving cleanup.",
] as const;
const AGENT_NAME_DESCRIPTION = `Name of the agent to invoke. Bundled agents: ${BUNDLED_AGENT_NAMES}. Use these exact names for bundled agents; do not invent names such as default, general-purpose, security, or general.`;
const SUBAGENT_DEPTH_ENV = "PI_SUB_AGENT_DEPTH";
const MAX_SUBAGENT_DEPTH = 1;

interface UsageStats {
	input: number;
	output: number;
	cacheRead: number;
	cacheWrite: number;
	cost: number;
	contextTokens: number;
	turns: number;
}

interface RawUsage {
	input?: number;
	output?: number;
	cacheRead?: number;
	cacheWrite?: number;
	totalTokens?: number;
	cost?: {
		total?: number;
	};
}

interface RawMessage {
	role?: string;
	content?: unknown;
	usage?: RawUsage;
	model?: string;
	stopReason?: string;
	errorMessage?: string;
}

interface SingleResult {
	agent: string;
	agentSource: "user" | "project" | "extension" | "unknown";
	task: string;
	exitCode: number;
	messages: RawMessage[];
	stderr: string;
	stdout?: string;
	usage: UsageStats;
	model?: string;
	stopReason?: string;
	errorMessage?: string;
	step?: number;
}

interface SubagentDetails {
	mode: "single" | "parallel" | "chain";
	agentScope: AgentScope;
	projectAgentsDir: string | null;
	results: SingleResult[];
	error?: string;
}

type OnUpdateCallback = (partial: { content: Array<{ type: "text"; text: string }>; details: SubagentDetails }) => void;

type DisplayItem =
	| { type: "text"; text: string }
	| { type: "toolCall"; name: string; args: Record<string, unknown> };

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null;
}

function formatTokens(count: number): string {
	if (count < 1000) return count.toString();
	if (count < 10000) return `${(count / 1000).toFixed(1)}k`;
	if (count < 1000000) return `${Math.round(count / 1000)}k`;
	return `${(count / 1000000).toFixed(1)}M`;
}

function formatUsageStats(usage: UsageStats, model?: string): string {
	const parts: string[] = [];
	if (usage.turns) parts.push(`${usage.turns} turn${usage.turns > 1 ? "s" : ""}`);
	if (usage.input) parts.push(`↑${formatTokens(usage.input)}`);
	if (usage.output) parts.push(`↓${formatTokens(usage.output)}`);
	if (usage.cacheRead) parts.push(`R${formatTokens(usage.cacheRead)}`);
	if (usage.cacheWrite) parts.push(`W${formatTokens(usage.cacheWrite)}`);
	if (usage.cost) parts.push(`$${usage.cost.toFixed(4)}`);
	if (usage.contextTokens) parts.push(`ctx:${formatTokens(usage.contextTokens)}`);
	if (model) parts.push(model);
	return parts.join(" ");
}

function formatAggregateUsageStats(results: SingleResult[]): string {
	const usage: UsageStats = { input: 0, output: 0, cacheRead: 0, cacheWrite: 0, cost: 0, contextTokens: 0, turns: 0 };
	for (const result of results) {
		usage.input += result.usage.input;
		usage.output += result.usage.output;
		usage.cacheRead += result.usage.cacheRead;
		usage.cacheWrite += result.usage.cacheWrite;
		usage.cost += result.usage.cost;
		usage.turns += result.usage.turns;
		usage.contextTokens = Math.max(usage.contextTokens, result.usage.contextTokens);
	}
	return formatUsageStats(usage);
}

function shortenPath(filePath: string): string {
	const home = os.homedir();
	if (filePath === home) return "~";
	return filePath.startsWith(`${home}${sep}`) ? `~${filePath.slice(home.length)}` : filePath;
}

function formatToolCall(
	toolName: string,
	args: Record<string, unknown>,
	themeFg: (color: ThemeColor, text: string) => string,
): string {
	switch (toolName) {
		case "bash": {
			const command = typeof args.command === "string" ? args.command : "...";
			const preview = makePlaceholder(command);
			return themeFg("muted", "$ ") + themeFg("toolOutput", preview);
		}
		case "read": {
			const rawPath = typeof args.file_path === "string" ? args.file_path : typeof args.path === "string" ? args.path : "...";
			let text = themeFg("accent", shortenPath(rawPath));
			const offset = typeof args.offset === "number" ? args.offset : undefined;
			const limit = typeof args.limit === "number" ? args.limit : undefined;
			if (offset !== undefined || limit !== undefined) {
				const startLine = offset ?? 1;
				const endLine = limit !== undefined ? startLine + limit - 1 : "";
				text += themeFg("warning", `:${startLine}${endLine ? `-${endLine}` : ""}`);
			}
			return themeFg("muted", "read ") + text;
		}
		case "write": {
			const rawPath = typeof args.file_path === "string" ? args.file_path : typeof args.path === "string" ? args.path : "...";
			const content = typeof args.content === "string" ? args.content : "";
			const lines = content ? content.split("\n").length : 0;
			let text = themeFg("muted", "write ") + themeFg("accent", shortenPath(rawPath));
			if (lines > 1) text += themeFg("dim", ` (${lines} lines)`);
			return text;
		}
		case "edit": {
			const rawPath = typeof args.file_path === "string" ? args.file_path : typeof args.path === "string" ? args.path : "...";
			return themeFg("muted", "edit ") + themeFg("accent", shortenPath(rawPath));
		}
		case "ls": {
			const rawPath = typeof args.path === "string" ? args.path : ".";
			return themeFg("muted", "ls ") + themeFg("accent", shortenPath(rawPath));
		}
		case "find": {
			const pattern = typeof args.pattern === "string" ? args.pattern : "*";
			const rawPath = typeof args.path === "string" ? args.path : ".";
			return themeFg("muted", "find ") + themeFg("accent", pattern) + themeFg("dim", ` in ${shortenPath(rawPath)}`);
		}
		case "grep": {
			const pattern = typeof args.pattern === "string" ? args.pattern : "";
			const rawPath = typeof args.path === "string" ? args.path : ".";
			return themeFg("muted", "grep ") + themeFg("accent", `/${pattern}/`) + themeFg("dim", ` in ${shortenPath(rawPath)}`);
		}
		default: {
			const argsText = JSON.stringify(args);
			const preview = makePlaceholder(argsText);
			return themeFg("accent", toolName) + themeFg("dim", ` ${preview}`);
		}
	}
}

function getPiInvocation(args: string[]): { command: string; args: string[] } {
	const currentScript = process.argv[1];
	const isBunVirtualScript = currentScript?.startsWith("/$bunfs/root/");
	if (currentScript && !isBunVirtualScript && fs.existsSync(currentScript)) {
		return { command: process.execPath, args: [currentScript, ...args] };
	}

	const execName = basename(process.execPath).toLowerCase();
	if (/^(node|bun)(\.exe)?$/.test(execName)) {
		return { command: "pi", args };
	}
	return { command: process.execPath, args };
}

function addUsage(base: UsageStats, raw?: RawUsage): UsageStats {
	if (!raw) return base;
	return {
		...base,
		input: base.input + (raw.input ?? 0),
		output: base.output + (raw.output ?? 0),
		cacheRead: base.cacheRead + (raw.cacheRead ?? 0),
		cacheWrite: base.cacheWrite + (raw.cacheWrite ?? 0),
		cost: base.cost + (raw.cost?.total ?? 0),
		contextTokens: raw.totalTokens ?? base.contextTokens,
	};
}

function collectTextFromMessage(message: RawMessage): string | undefined {
	const content = message.content;
	if (typeof content === "string") return content;
	if (!Array.isArray(content)) return undefined;
	const textParts: string[] = [];
	for (const part of content) {
		if (isRecord(part) && part.type === "text" && typeof part.text === "string") {
			textParts.push(part.text);
		}
	}
	return textParts.length > 0 ? textParts.join("") : undefined;
}

function collectFinalOutput(messages: RawMessage[]): string {
	for (let i = messages.length - 1; i >= 0; i--) {
		const message = messages[i];
		if (!message) continue;
		if (message.role !== "assistant") continue;
		const text = collectTextFromMessage(message);
		if (text !== undefined) return text;
	}
	return "";
}

function collectDisplayItems(messages: RawMessage[]): DisplayItem[] {
	const items: DisplayItem[] = [];
	for (const message of messages) {
		if (message.role !== "assistant") continue;
		const content = message.content;
		if (typeof content === "string") {
			items.push({ type: "text", text: content });
			continue;
		}
		if (!Array.isArray(content)) continue;
		for (const part of content) {
			if (!isRecord(part)) continue;
			if (part.type === "text" && typeof part.text === "string") {
				items.push({ type: "text", text: part.text });
			} else if (part.type === "toolCall" && typeof part.name === "string") {
				items.push({
					type: "toolCall",
					name: part.name,
					args: isRecord(part.arguments) ? part.arguments : {},
				});
			}
		}
	}
	return items;
}

function getLastErrorMessage(messages: RawMessage[]): string {
	for (let i = messages.length - 1; i >= 0; i--) {
		const message = messages[i];
		if (!message) continue;
		if (message.role === "assistant" && typeof message.errorMessage === "string" && message.errorMessage) {
			return message.errorMessage;
		}
	}
	return "";
}

async function writePromptToTempFile(agentName: string, prompt: string): Promise<{ dir: string; filePath: string }> {
	const tmpDir = await fs.promises.mkdtemp(join(os.tmpdir(), "pi-subagent-"));
	const safeName = agentName.replace(/[^\w.-]+/g, "_");
	const filePath = join(tmpDir, `prompt-${safeName}.md`);
	await withFileMutationQueue(filePath, async () => {
		await fs.promises.writeFile(filePath, prompt, { encoding: "utf-8", mode: 0o600 });
	});
	return { dir: tmpDir, filePath };
}

function makePlaceholder(text = ""): string {
	return text.length > 120 ? `${text.slice(0, 120)}...` : text;
}

function resolveSubagentCwd(defaultCwd: string, cwd: string | undefined): string {
	if (!cwd) return defaultCwd;
	const normalized = cwd.startsWith("@") ? cwd.slice(1) : cwd;
	return isAbsolute(normalized) ? normalized : resolve(defaultCwd, normalized);
}

function getInvalidSubagentCwdReason(cwd: string): string | undefined {
	try {
		const stats = fs.statSync(cwd);
		return stats.isDirectory() ? undefined : "exists but is not a directory";
	} catch (error) {
		if (isRecord(error) && typeof error.code === "string") return error.code;
		return error instanceof Error ? error.message : String(error);
	}
}

function getSubagentDepth(value = process.env[SUBAGENT_DEPTH_ENV]): number {
	if (value === undefined) return 0;
	const parsed = Number.parseInt(value, 10);
	return Number.isFinite(parsed) && parsed > 0 ? parsed : 0;
}

function normalizeToolNames(toolNames: readonly string[] | undefined): string[] | undefined {
	if (toolNames === undefined) return undefined;
	return Array.from(new Set(toolNames.map((tool) => tool.trim()).filter((tool) => Boolean(tool) && tool !== SUBAGENT_TOOL_NAME)));
}

function resolveChildToolAllowlist(
	agentTools: readonly string[] | undefined,
	parentActiveTools: readonly string[] | undefined,
): string[] | undefined {
	const parentTools = normalizeToolNames(parentActiveTools);
	const requestedTools = normalizeToolNames(agentTools);
	if (parentTools === undefined) return requestedTools;
	if (requestedTools === undefined) return parentTools;
	const parentToolSet = new Set(parentTools);
	return requestedTools.filter((tool) => parentToolSet.has(tool));
}

function writeFullOutputTempFile(text: string): string {
	const tmpDir = fs.mkdtempSync(join(os.tmpdir(), "pi-subagent-output-"));
	const filePath = join(tmpDir, "output.txt");
	fs.writeFileSync(filePath, text, { encoding: "utf-8", mode: 0o600 });
	fs.chmodSync(filePath, 0o600);
	return filePath;
}

function truncateForToolContent(text: string): string {
	const truncation = truncateTail(text, {
		maxLines: DEFAULT_MAX_LINES,
		maxBytes: DEFAULT_MAX_BYTES,
	});
	if (!truncation.truncated) return truncation.content;
	const fullOutputPath = writeFullOutputTempFile(text);
	return [
		truncation.content,
		`[Subagent output truncated: showing ${truncation.outputLines} of ${truncation.totalLines} lines (${formatSize(truncation.outputBytes)} of ${formatSize(truncation.totalBytes)}). Full output saved to: ${fullOutputPath}\nExpand the tool result for full structured details.]`,
	].filter(Boolean).join("\n\n");
}

function expandToolOutputHint(theme: { fg: (color: ThemeColor, text: string) => string }): string {
	const expandKey = keyText("app.tools.expand") || "ctrl+o";
	return `(${theme.fg("dim", expandKey)}${theme.fg("muted", " to expand")})`;
}

async function mapWithConcurrencyLimit<TIn, TOut>(
	items: TIn[],
	concurrency: number,
	fn: (item: TIn, index: number) => Promise<TOut>,
): Promise<TOut[]> {
	if (items.length === 0) return [];
	const limit = Math.max(1, Math.min(concurrency, items.length));
	const results = new Array<TOut | undefined>(items.length);
	let nextIndex = 0;
	const workers = new Array(limit).fill(null).map(async () => {
		while (true) {
			const current = nextIndex++;
			if (current >= items.length) return;
			const item = items[current];
			if (item === undefined) continue;
			results[current] = await fn(item, current);
		}
	});
	await Promise.all(workers);
	return results.map((value, index) => {
		if (value === undefined) {
			throw new Error(`Parallel worker did not return a result for task index ${index}`);
		}
		return value;
	});
}

function isTaskError(result: SingleResult): boolean {
	return (
		result.exitCode !== 0 ||
		result.stopReason === "error" ||
		result.stopReason === "aborted" ||
		result.stopReason === "length"
	);
}

function isFailedResultLike(value: unknown): boolean {
	if (!isRecord(value)) return false;
	const exitCode = typeof value.exitCode === "number" ? value.exitCode : 0;
	const stopReason = typeof value.stopReason === "string" ? value.stopReason : undefined;
	return exitCode !== 0 || stopReason === "error" || stopReason === "aborted" || stopReason === "length";
}

function hasFailedSubagentResult(details: unknown): boolean {
	if (!isRecord(details)) return false;
	if (typeof details.error === "string" && details.error.trim()) return true;
	return Array.isArray(details.results) && details.results.some(isFailedResultLike);
}

function addDistinctSection(parts: string[], label: string, value: string | undefined): void {
	const trimmed = value?.trim();
	if (!trimmed) return;
	const section = `${label}:\n${trimmed}`;
	if (parts.includes(section)) return;
	parts.push(section);
}

function formatFailureOutput(result: SingleResult): string {
	const parts: string[] = [];
	// Keep diagnostics after assistant output because truncateForToolContent() preserves the tail.
	addDistinctSection(parts, "Output", collectFinalOutput(result.messages));
	addDistinctSection(parts, "stderr", result.stderr);
	addDistinctSection(parts, "stdout", result.stdout);

	const errorMessage = result.errorMessage?.trim();
	const errorAlreadyInProcessOutput = Boolean(
		errorMessage &&
		(result.stderr.includes(errorMessage) || result.stdout?.includes(errorMessage))
	);
	if (!errorAlreadyInProcessOutput) {
		addDistinctSection(parts, "Error", errorMessage);
	}

	addDistinctSection(parts, "stopReason", result.stopReason ? `${result.stopReason} (exit code ${result.exitCode})` : undefined);
	if (result.exitCode !== 0 && !result.stopReason) parts.push(`Exit code:\n${result.exitCode}`);
	if (parts.length > 0) return parts.join("\n\n");
	return `Subagent exited with code ${result.exitCode}.`;
}

function formatParallelToolContent(results: SingleResult[]): string {
	const succeeded = results.filter((result) => !isTaskError(result)).length;
	const sections = [`Parallel tasks: ${succeeded}/${results.length} succeeded`];

	for (const [index, result] of results.entries()) {
		const failed = isTaskError(result);
		const icon = failed ? "✗" : "✓";
		const usage = formatUsageStats(result.usage, result.model);
		const body = failed ? formatFailureOutput(result) : collectFinalOutput(result.messages) || "(no output)";
		const section = [
			`## ${index + 1}. ${icon} ${result.agent}`,
			`Task: ${result.task}`,
			...(usage ? [`Usage: ${usage}`] : []),
			"",
			failed ? "Failure:" : "Output:",
			truncateForToolContent(body),
		];
		sections.push(section.join("\n"));
	}

	return sections.join("\n\n");
}

function getFailureDiagnostic(result: SingleResult): { label: string; text: string } | undefined {
	const errorMessage = result.errorMessage?.trim();
	const stderr = result.stderr.trim();
	const stdout = result.stdout?.trim();
	if (errorMessage) {
		const parts = [errorMessage];
		if (stderr && !stderr.includes(errorMessage)) parts.push(`stderr:\n${stderr}`);
		if (stdout) parts.push(`stdout:\n${stdout}`);
		return { label: "Error", text: parts.join("\n\n") };
	}

	if (stderr) return { label: "stderr", text: stderr };

	if (stdout) return { label: "stdout", text: stdout };

	if (result.stopReason) {
		return { label: "stopReason", text: `${result.stopReason} (exit code ${result.exitCode})` };
	}

	if (result.exitCode !== 0) return { label: "Exit code", text: String(result.exitCode) };
	return undefined;
}

function formatFailureDiagnostic(result: SingleResult, compact = false): string {
	const diagnostic = getFailureDiagnostic(result);
	if (!diagnostic) return "";
	const text = compact ? diagnostic.text.replace(/\s+/g, " ") : diagnostic.text;
	return compact ? makePlaceholder(`${diagnostic.label}: ${text}`) : `${diagnostic.label}:\n${text}`;
}

function snapshotResult(result: SingleResult): SingleResult {
	return {
		...result,
		messages: [...result.messages],
		usage: { ...result.usage },
	};
}

async function runSingleAgent(
	defaultCwd: string,
	agents: AgentConfig[],
	agentName: string,
	task: string,
	cwd: string | undefined,
	fallbackModel: string | undefined,
	parentActiveTools: readonly string[] | undefined,
	childSubagentDepth: number,
	step: number | undefined,
	signal: AbortSignal | undefined,
	onUpdate: OnUpdateCallback | undefined,
	makeDetails: (results: SingleResult[]) => SubagentDetails,
): Promise<SingleResult> {
	const agent = agents.find((candidate) => candidate.name === agentName);
	const emptyUsage: UsageStats = { input: 0, output: 0, cacheRead: 0, cacheWrite: 0, cost: 0, contextTokens: 0, turns: 0 };

	if (!agent) {
		const known = agents.map((entry) => entry.name).sort().join(", ") || "none";
		const result: SingleResult = {
			agent: agentName,
			agentSource: "unknown",
			task,
			exitCode: 1,
			messages: [],
			stderr: `Unknown agent: "${agentName}". Available: ${known}`,
			usage: { ...emptyUsage },
		};
		if (step !== undefined) {
			result.step = step;
		}
		return result;
	}

	// Polymetrics local modification (vendored from pi-sub-agent@0.1.5, MIT): child sessions are
	// recordable via PI_SUBAGENT_SESSION_DIR for loop observability (loop-trace digests). Upstream
	// hardcodes --no-session; we keep that default when the env var is unset.
	const subSessionDir = process.env.PI_SUBAGENT_SESSION_DIR;
	const args = subSessionDir
		? ["--mode", "json", "-p", "--session-dir", subSessionDir]
		: ["--mode", "json", "-p", "--no-session"];
	const stdinPrompt = `Task: ${task}`;
	const selectedModel = resolveAgentModel(agent, fallbackModel);
	if (selectedModel) {
		args.push("--model", selectedModel);
	}
	const childTools = resolveChildToolAllowlist(agent.tools, parentActiveTools);
	if (childTools !== undefined) {
		if (childTools.length > 0) {
			args.push("--tools", childTools.join(","));
		} else {
			args.push("--no-tools");
		}
	}

	let tmpDir: string | null = null;
	let tmpPromptPath: string | null = null;
	let aborted = false;
	let abortListener: (() => void) | null = null;
	const result: SingleResult = {
		agent: agent.name,
		agentSource: agent.source,
		task,
		exitCode: -1,
		messages: [],
		stderr: "",
		stdout: "",
		usage: { ...emptyUsage },
	};
	if (step !== undefined) {
		result.step = step;
	}
	if (selectedModel) {
		result.model = selectedModel;
	}

	const emitUpdate = () => {
		if (!onUpdate) return;
		const output = collectFinalOutput(result.messages);
		onUpdate({
			content: [{ type: "text", text: output ? truncateForToolContent(output) : "(running...)" }],
			details: makeDetails([snapshotResult(result)]),
		});
	};

	try {
		const childCwd = resolveSubagentCwd(defaultCwd, cwd);
		const invalidCwdReason = getInvalidSubagentCwdReason(childCwd);
		if (invalidCwdReason) {
			const diagnostic = `Subagent working directory does not exist or is not a directory: ${childCwd} (${invalidCwdReason})`;
			result.exitCode = 1;
			result.stderr = diagnostic;
			result.errorMessage = diagnostic;
			return result;
		}

		if (agent.systemPrompt.trim()) {
			const file = await writePromptToTempFile(agent.name, agent.systemPrompt);
			tmpDir = file.dir;
			tmpPromptPath = file.filePath;
			args.push("--append-system-prompt", file.filePath);
		}

		const exit = await new Promise<{ code: number; signal: NodeJS.Signals | null }>((resolve) => {
			const invocation = getPiInvocation(args);
			const childEnv = { ...process.env, [SUBAGENT_DEPTH_ENV]: String(childSubagentDepth) };
			const proc = spawn(invocation.command, invocation.args, {
				cwd: childCwd,
				env: childEnv,
				shell: false,
				stdio: ["pipe", "pipe", "pipe"],
			});

			let buffer = "";
			let closed = false;
			let killTimer: ReturnType<typeof setTimeout> | undefined;
			const parseLine = (line: string): void => {
				const trimmed = line.trim();
				if (!trimmed) return;
				let event: unknown;
				try {
					event = JSON.parse(trimmed);
				} catch {
					result.stdout = result.stdout ? `${result.stdout}\n${line}` : line;
					return;
				}
				if (!isRecord(event)) return;
				const eventType = event.type;
				if (eventType !== "message_end" && eventType !== "tool_result_end") return;
				const rawMessage = event.message;
				if (!isRecord(rawMessage)) return;
				const message = rawMessage as RawMessage;
				result.messages.push(message);
				if (message.role === "assistant") {
					result.usage = addUsage(result.usage, message.usage);
					result.usage.turns += 1;
					if (!result.model && typeof message.model === "string") {
						result.model = message.model;
					}
					if (typeof message.stopReason === "string") {
						result.stopReason = message.stopReason;
					}
					if (typeof message.errorMessage === "string" && message.errorMessage) {
						result.errorMessage = message.errorMessage;
					}
				}
				emitUpdate();
			};

			proc.stdout.on("data", (data) => {
				buffer += data.toString();
				const lines = buffer.split("\n");
				buffer = lines.pop() ?? "";
				for (const line of lines) {
					parseLine(line);
				}
			});
			proc.stderr.on("data", (data) => {
				result.stderr += data.toString();
			});
			proc.stdin?.on("error", () => {
				// Ignore stdin errors from subprocesses that exit before consuming the prompt.
			});
			proc.stdin?.end(stdinPrompt);

			const abort = (): void => {
				if (aborted) return;
				aborted = true;
				proc.kill("SIGTERM");
				killTimer = setTimeout(() => {
					if (!closed) proc.kill("SIGKILL");
				}, 5000);
			};

			if (signal) {
				if (signal.aborted) {
					abort();
				} else {
					abortListener = () => abort();
					signal.addEventListener("abort", abortListener, { once: true });
				}
			}

			proc.on("close", (code, signalName) => {
				closed = true;
				if (killTimer) {
					clearTimeout(killTimer);
					killTimer = undefined;
				}
				if (buffer.trim()) parseLine(buffer);
				if (signal && abortListener) {
					signal.removeEventListener("abort", abortListener);
					abortListener = null;
				}
				resolve({ code: code ?? (signalName ? 1 : 0), signal: signalName });
			});
			proc.on("error", (error) => {
				const detail = error instanceof Error ? error.message : String(error);
				const diagnostic = `Failed to start subagent process (${invocation.command}): ${detail}`;
				result.stderr = result.stderr ? `${result.stderr}\n${diagnostic}` : diagnostic;
				result.errorMessage = diagnostic;
				resolve({ code: 1, signal: null });
			});
		});

		result.exitCode = exit.code;
		if (exit.signal && !aborted) {
			result.errorMessage = `Subagent process terminated by signal ${exit.signal}.`;
		}
		if (result.exitCode === 0 && result.messages.length === 0 && result.stdout?.trim()) {
			result.exitCode = 1;
			result.errorMessage = "Subagent produced non-JSON stdout without any JSON messages.";
		}
		if (aborted) {
			result.exitCode = 1;
			result.stopReason = "aborted";
			if (!result.errorMessage) {
				result.errorMessage = "Subagent execution was aborted.";
			}
		}
		if (!result.errorMessage) {
			const last = getLastErrorMessage(result.messages);
			if (last) result.errorMessage = last;
		}
	} catch (error) {
		result.exitCode = 1;
		result.errorMessage = error instanceof Error ? error.message : String(error);
	} finally {
		if (tmpPromptPath) {
			try {
				fs.unlinkSync(tmpPromptPath);
			} catch {
				// Ignore cleanup errors.
			}
		}
		if (tmpDir) {
			try {
				fs.rmdirSync(tmpDir);
			} catch {
				// Ignore cleanup errors.
			}
		}
		if (!result.stderr && result.exitCode !== 0 && !collectFinalOutput(result.messages)) {
			result.stderr = `Subagent exited with code ${result.exitCode}.`;
		}
	}

	if (result.exitCode !== 0 && result.errorMessage && !result.stderr.includes(result.errorMessage)) {
		result.stderr = result.stderr ? `${result.stderr}\n${result.errorMessage}` : result.errorMessage;
	}

	return result;
}

const TaskItem = Type.Object({
	agent: Type.String({ description: AGENT_NAME_DESCRIPTION, minLength: 1 }),
	task: Type.String({ description: "Task to delegate", minLength: 1 }),
	cwd: Type.Optional(Type.String({ description: "Working directory override for this task", minLength: 1 })),
});

const ChainItem = Type.Object({
	agent: Type.String({ description: AGENT_NAME_DESCRIPTION, minLength: 1 }),
	task: Type.String({
		description: "Task with optional {previous} placeholder",
		minLength: 1,
	}),
	cwd: Type.Optional(Type.String({ description: "Working directory override for this task", minLength: 1 })),
});

const AgentScopeSchema = StringEnum(["user", "project", "both"] as const, {
	description: 'Which agent directories to use. Default: "user". Use "both" to include project-local agents.',
	default: "user",
});

const SubagentParams = Type.Object({
	agent: Type.Optional(Type.String({ description: `Single mode agent name. ${AGENT_NAME_DESCRIPTION}`, minLength: 1 })),
	task: Type.Optional(Type.String({ description: "Single mode task text", minLength: 1 })),
	tasks: Type.Optional(Type.Array(TaskItem, { description: "Parallel mode task list", minItems: 1, maxItems: MAX_PARALLEL_TASKS })),
	chain: Type.Optional(Type.Array(ChainItem, { description: "Chain mode task list", minItems: 1, maxItems: MAX_CHAIN_STEPS })),
	agentScope: Type.Optional(AgentScopeSchema),
	confirmProjectAgents: Type.Optional(Type.Boolean({ description: "Confirm before running project agents", default: true })),
	cwd: Type.Optional(Type.String({ description: "Default working directory for single, parallel, and chain modes. Per-task or per-step cwd overrides this default.", minLength: 1 })),
});

const INHERIT = "inherit";

function formatAgentSettings(agent: Pick<AgentConfig, "model" | "thinking">): string {
	return `${agent.model ?? INHERIT} • ${agent.thinking ?? INHERIT}`;
}

function getAgentSettingsTarget(agent: AgentConfig): string {
	if (agent.source !== "extension") return agent.filePath;
	return join(getAgentDir(), "agents", `${agent.name}.md`);
}

function persistAgentSettings(agent: AgentConfig): string {
	const targetPath = getAgentSettingsTarget(agent);
	const content = fs.readFileSync(agent.filePath, "utf-8");
	const updated = updateAgentSettingsContent(content, {
		model: agent.model ?? null,
		thinking: agent.thinking ?? null,
	});
	fs.mkdirSync(dirname(targetPath), { recursive: true });
	fs.writeFileSync(targetPath, updated, "utf-8");
	return targetPath;
}

function isAgentThinkingLevel(value: string): value is AgentThinkingLevel {
	return (THINKING_LEVELS as readonly string[]).includes(value);
}

export default function (pi: ExtensionAPI): void {
	if (typeof pi.registerCommand === "function") pi.registerCommand("sub-agent-settings", {
		description: "Configure sub-agent models and thinking effort",
		handler: async (_args, ctx) => {
			if (!ctx.hasUI) {
				ctx.ui.notify("Sub-agent settings require an interactive UI.", "warning");
				return;
			}

			const discovery = discoverAgents(ctx.cwd, "user", join(extensionDir, "agents"));
			const agents = discovery.agents.sort((a, b) => a.name.localeCompare(b.name));
			if (agents.length === 0) {
				ctx.ui.notify("No sub-agents found", "warning");
				return;
			}

			ctx.modelRegistry.refresh();
			const modelLoadError = ctx.modelRegistry.getError();
			if (modelLoadError) {
				ctx.ui.notify(modelLoadError, "warning");
			}

			const models = ctx.modelRegistry
				.getAvailable()
				.map((model) => ({
					value: `${model.provider}/${model.id}`,
					label: model.id,
					description: `${model.provider}${model.reasoning ? " • reasoning" : ""}`,
				}))
				.sort((a, b) => a.value.localeCompare(b.value));

			await ctx.ui.custom((tui, theme, _keybindings, done) => {
				const mainItems: SettingItem[] = agents.map((agent) => ({
					id: agent.name,
					label: agent.name,
					currentValue: formatAgentSettings(agent),
					description: `${agent.source} • ${agent.description}`,
					submenu: (_currentValue, closeAgentMenu) => {
						const agentSettingsItems: SettingItem[] = [
							{
								id: "model",
								label: "Model",
								currentValue: agent.model ?? INHERIT,
								description: "Model used by this sub-agent. Inherit uses the parent Pi session model.",
								submenu: (_currentModel, closeModelMenu) => {
									const selectItems = [
										{ value: INHERIT, label: INHERIT, description: "Use the parent Pi session model" },
										...models,
									];
									const selectList = new SelectList(selectItems, Math.min(selectItems.length, 12), {
										selectedPrefix: (text: string) => theme.fg("accent", text),
										selectedText: (text: string) => theme.fg("accent", text),
										description: (text: string) => theme.fg("muted", text),
										scrollInfo: (text: string) => theme.fg("dim", text),
										noMatch: (text: string) => theme.fg("warning", text),
									});
									const selectedIndex = selectItems.findIndex((item) => item.value === (agent.model ?? INHERIT));
									selectList.setSelectedIndex(Math.max(0, selectedIndex));
									selectList.onSelect = (item) => closeModelMenu(item.value);
									selectList.onCancel = () => closeModelMenu(undefined);
									return selectList;
								},
							},
							{
								id: "thinking",
								label: "Thinking",
								currentValue: agent.thinking ?? INHERIT,
								description: "Thinking effort for this sub-agent. Inherit uses the parent Pi session thinking level.",
								values: [INHERIT, ...THINKING_LEVELS],
							},
						];

						const persistAndNotify = () => {
							const targetPath = persistAgentSettings(agent);
							ctx.ui.notify(`Saved ${agent.name} settings to ${shortenPath(targetPath)}`, "info");
						};
						const settingsList = new SettingsList(
							agentSettingsItems,
							agentSettingsItems.length + 2,
							getSettingsListTheme(),
							(id, newValue) => {
								if (id === "model") {
									if (newValue === INHERIT) {
										delete agent.model;
									} else {
										agent.model = newValue;
									}
								} else if (id === "thinking") {
									if (newValue === INHERIT) {
										delete agent.thinking;
									} else if (isAgentThinkingLevel(newValue)) {
										agent.thinking = newValue;
									}
								}
								persistAndNotify();
							},
							() => closeAgentMenu(formatAgentSettings(agent)),
						);

						return {
							render(width: number) {
								return [
									theme.fg("accent", theme.bold(`Sub-agent: ${agent.name}`)),
									"",
									...settingsList.render(width),
								];
							},
							invalidate() {
								settingsList.invalidate();
							},
							handleInput(data: string) {
								settingsList.handleInput(data);
								tui.requestRender();
							},
						};
					},
				}));

				const settingsList = new SettingsList(
					mainItems,
					Math.min(mainItems.length + 2, 15),
					getSettingsListTheme(),
					() => {},
					() => done(undefined),
					{ enableSearch: true },
				);

				return {
					render(width: number) {
						return [
							theme.fg("accent", theme.bold("Sub-agent Settings")),
							theme.fg("dim", "Configure models and thinking effort. Bundled agents save as user overrides."),
							"",
							...settingsList.render(width),
						];
					},
					invalidate() {
						settingsList.invalidate();
					},
					handleInput(data: string) {
						settingsList.handleInput(data);
						tui.requestRender();
					},
				};
			});
		},
	});

	pi.on("tool_result", (event) => {
		if (event.toolName !== "subagent") return;
		if (!hasFailedSubagentResult(event.details)) return;
		return { isError: true };
	});

	pi.registerTool({
		name: "subagent",
		label: "Subagent",
		description: [
			"Delegate tasks to specialized subagents with isolated context.",
			`Supports single, parallel, and chain flows; parallel mode is capped at ${MAX_PARALLEL_TASKS} tasks and chain mode at ${MAX_CHAIN_STEPS} steps.`,
			`LLM-facing output is truncated per included subagent output to ${DEFAULT_MAX_LINES} lines or ${formatSize(DEFAULT_MAX_BYTES)}; full structured details remain available for rendering.`,
			"Nested subagent calls are disabled to avoid runaway recursive delegation.",
			`Bundled agents: ${BUNDLED_AGENT_NAMES}. Use these exact names; do not invent names such as default, general-purpose, security, or general.`,
			'User agents are used by default from ~/.pi/agent/agents.',
			'Use agentScope "project" or "both" to include trusted project-local agents from .pi/agents.',
		].join(" "),
		promptSnippet: "Delegate work to specialized subagents in isolated Pi processes; supports single, parallel, and chain modes.",
		promptGuidelines: [
			"Use subagent when a task benefits from isolated context, parallel research, or specialized bundled/user/project agents.",
			`Bundled agents: ${BUNDLED_AGENT_NAMES}. Use these exact names; do not invent names such as default, general-purpose, security, or general.`,
			...BUNDLED_AGENT_SELECTION_GUIDANCE,
			`Keep subagent parallel task lists to ${MAX_PARALLEL_TASKS} tasks or fewer and chain step lists to ${MAX_CHAIN_STEPS} steps or fewer.`,
			"Use exactly one argument mode: single {agent, task, cwd?}, parallel {tasks, cwd?}, or chain {chain, cwd?}. Top-level cwd is a default for every subagent run in the call; per-task or per-step cwd overrides it.",
			'Use subagent with agentScope "project" or "both" only for trusted repositories because project agents are repo-controlled prompts.',
			"Use subagent chain tasks with {previous} only when each step should consume the previous agent output.",
			"Do not ask subagent-launched agents to call subagent again; recursive delegation is blocked.",
		],
		parameters: SubagentParams,
		async execute(_toolCallId, params, signal, onUpdate, ctx) {
			const agentScope: AgentScope = params.agentScope ?? "user";
			const parentActiveTools = typeof pi.getActiveTools === "function" ? pi.getActiveTools() : undefined;
			const parentThinkingLevel = typeof pi.getThinkingLevel === "function" ? pi.getThinkingLevel() : "off";
			const parentThinkingSuffix = parentThinkingLevel && parentThinkingLevel !== "off" ? `:${parentThinkingLevel}` : "";
			const parentModel = ctx.model ? `${ctx.model.provider}/${ctx.model.id}${parentThinkingSuffix}` : undefined;
			const discovery = discoverAgents(ctx.cwd, agentScope, join(extensionDir, "agents"));
			const agents = discovery.agents;
			const makeDetails = (mode: "single" | "parallel" | "chain") => (results: SingleResult[], error?: string): SubagentDetails => ({
				mode,
				agentScope,
				projectAgentsDir: discovery.projectAgentsDir,
				results,
				...(error ? { error } : {}),
			});

			const hasAgent = params.agent !== undefined;
			const hasTask = params.task !== undefined;
			const hasSingle = hasAgent && hasTask;
			const hasPartialSingle = hasAgent !== hasTask;
			const hasParallel = (params.tasks?.length ?? 0) > 0;
			const hasChain = (params.chain?.length ?? 0) > 0;
			const modeCount = Number(hasSingle) + Number(hasParallel) + Number(hasChain);

			if (modeCount !== 1 || hasPartialSingle) {
				const error = "Invalid subagent arguments. Use exactly one of: {agent,task,cwd?}, {tasks,cwd?}, or {chain,cwd?}.";
				return {
					content: [
						{
							type: "text",
							text: error,
						},
					],
					details: makeDetails("single")([], error),
				};
			}

			const selectedMode = hasChain ? "chain" : hasParallel ? "parallel" : "single";
			const currentDepth = getSubagentDepth();
			if (currentDepth >= MAX_SUBAGENT_DEPTH) {
				const error = "Nested subagent execution is disabled to avoid runaway recursive delegation.";
				return {
					content: [
						{
							type: "text",
							text: error,
						},
					],
					details: makeDetails(selectedMode)([], error),
				};
			}
			const childSubagentDepth = currentDepth + 1;

			if ((agentScope === "project" || agentScope === "both") && params.confirmProjectAgents !== false) {
				const names = new Set<string>();
				if (params.agent) names.add(params.agent);
				for (const task of params.tasks ?? []) names.add(task.agent);
				for (const step of params.chain ?? []) names.add(step.agent);
				const projectAgents = Array.from(names)
					.map((name) => agents.find((agent) => agent.name === name))
					.filter((agent): agent is AgentConfig => agent !== undefined && agent.source === "project");
				if (projectAgents.length > 0) {
					if (!ctx.hasUI) {
						const error = "Canceled: running project-local agents requires confirmation, but no UI is available. Set confirmProjectAgents: false only for trusted repositories.";
						return {
							content: [
								{
									type: "text",
									text: error,
								},
							],
							details: makeDetails(selectedMode)([], error),
						};
					}

					const approved = await ctx.ui.confirm(
						"Run project-local agents?",
						`Agents: ${projectAgents.map((entry) => entry.name).join(", ")}\nSource: ${discovery.projectAgentsDir ?? "(none)"}`,
					);
					if (!approved) {
						const error = "Canceled: project-local agents not approved.";
						return {
							content: [
								{
									type: "text",
									text: error,
								},
							],
							details: makeDetails(selectedMode)([], error),
						};
					}
				}
			}

			if (hasChain && params.chain) {
				if (params.chain.length > MAX_CHAIN_STEPS) {
					const error = `Too many chain steps; max is ${MAX_CHAIN_STEPS}.`;
					return {
						content: [
							{ type: "text", text: error },
						],
						details: makeDetails("chain")([], error),
					};
				}

				const results: SingleResult[] = [];
				let previous = "";
				for (let i = 0; i < params.chain.length; i++) {
					const step = params.chain[i];
					if (!step) continue;
					const task = step.task.replace(/\{previous\}/g, previous);
					const result = await runSingleAgent(
						ctx.cwd,
						agents,
						step.agent,
						task,
						step.cwd ?? params.cwd,
						parentModel,
						parentActiveTools,
						childSubagentDepth,
						i + 1,
						signal,
						onUpdate
							? (partial) => {
								const current = partial.details.results[0];
								if (!current) return;
								onUpdate({
									content: partial.content,
									details: {
										mode: "chain",
										agentScope,
										projectAgentsDir: discovery.projectAgentsDir,
										results: [...results, current],
									},
								});
							}
							: undefined,
						makeDetails("chain"),
					);
					results.push(result);
					if (isTaskError(result)) {
						return {
							content: [
								{
									type: "text",
									text: `Chain stopped at step ${i + 1} (${result.agent}).\n\n${truncateForToolContent(formatFailureOutput(result))}`,
								},
							],
							details: makeDetails("chain")(results),
						};
					}
					previous = collectFinalOutput(result.messages);
				}

				return {
					content: [
						{
							type: "text",
							text: truncateForToolContent(collectFinalOutput(results.at(-1)?.messages ?? []) || "(no output)"),
						},
					],
					details: makeDetails("chain")(results),
				};
			}

			if (hasParallel && params.tasks) {
				if (params.tasks.length > MAX_PARALLEL_TASKS) {
					const error = `Too many parallel tasks; max is ${MAX_PARALLEL_TASKS}.`;
					return {
						content: [
							{ type: "text", text: error },
						],
						details: makeDetails("parallel")([], error),
					};
				}

				const live: SingleResult[] = params.tasks.map((task) => ({
					agent: task.agent,
					agentSource: "unknown",
					task: task.task,
					exitCode: -1,
					messages: [],
					stderr: "",
					usage: { input: 0, output: 0, cacheRead: 0, cacheWrite: 0, cost: 0, contextTokens: 0, turns: 0 },
				}));

				const results = await mapWithConcurrencyLimit(params.tasks, MAX_CONCURRENCY, async (task, index) => {
					const result = await runSingleAgent(
						ctx.cwd,
						agents,
						task.agent,
						task.task,
						task.cwd ?? params.cwd,
						parentModel,
						parentActiveTools,
						childSubagentDepth,
						undefined,
						signal,
						onUpdate
							? (partial) => {
								const current = partial.details.results[0];
								if (!current) return;
								live[index] = current;
								onUpdate({
									content: partial.content,
									details: makeDetails("parallel")(live.map(snapshotResult)),
								});
							}
							: undefined,
						makeDetails("parallel"),
					);
					live[index] = result;
					return result;
				});

				return {
					content: [
						{
							type: "text",
							text: formatParallelToolContent(results),
						},
					],
					details: makeDetails("parallel")(results),
				};
			}

			if (params.agent && params.task) {
				const result = await runSingleAgent(
					ctx.cwd,
					agents,
					params.agent,
					params.task,
					params.cwd,
					parentModel,
					parentActiveTools,
					childSubagentDepth,
					undefined,
					signal,
					onUpdate,
					makeDetails("single"),
				);
				if (isTaskError(result)) {
					return {
						content: [
							{
								type: "text",
								text: truncateForToolContent(formatFailureOutput(result)),
							},
						],
						details: makeDetails("single")([result]),
					};
				}
				return {
					content: [{ type: "text", text: truncateForToolContent(collectFinalOutput(result.messages) || "(no output)") }],
					details: makeDetails("single")([result]),
				};
			}

			const available = agents.map((agent) => `${agent.name} (${agent.source})`).join(", ") || "none";
			const error = `Invalid request. Available agents: ${available}`;
			return {
				content: [
					{
						type: "text",
						text: error,
					},
				],
				details: makeDetails("single")([], error),
			};
		},
		renderCall(args, theme) {
			const scope: AgentScope = args.agentScope ?? "user";
			if (args.chain && args.chain.length > 0) {
				let text =
					theme.fg("toolTitle", theme.bold("subagent ")) +
					theme.fg("accent", `chain (${args.chain.length} steps)`) +
					theme.fg("muted", ` [${scope}]`);
				for (let i = 0; i < Math.min(args.chain.length, 3); i++) {
					const step = args.chain[i];
					if (!step) continue;
					const cleanTask = step.task.replace(/\{previous\}/g, "").trim();
					const preview = makePlaceholder(cleanTask);
					text += `\n  ${theme.fg("muted", `${i + 1}.`)} ${theme.fg("accent", step.agent)}${theme.fg("dim", ` ${preview}`)}`;
				}
				if (args.chain.length > 3) {
					text += `\n  ${theme.fg("muted", `... +${args.chain.length - 3} more`)}`;
				}
				return new Text(text, 0, 0);
			}

			if (args.tasks && args.tasks.length > 0) {
				let text =
					theme.fg("toolTitle", theme.bold("subagent ")) +
					theme.fg("accent", `parallel (${args.tasks.length} tasks)`) +
					theme.fg("muted", ` [${scope}]`);
				for (const task of args.tasks.slice(0, 3)) {
					const preview = makePlaceholder(task.task);
					text += `\n  ${theme.fg("accent", task.agent)}${theme.fg("dim", ` ${preview}`)}`;
				}
				if (args.tasks.length > 3) {
					text += `\n  ${theme.fg("muted", `... +${args.tasks.length - 3} more`)}`;
				}
				return new Text(text, 0, 0);
			}

			const agentName = args.agent || "...";
			const preview = args.task ? makePlaceholder(args.task) : "...";
			const text =
				theme.fg("toolTitle", theme.bold("subagent ")) +
				theme.fg("accent", agentName) +
				theme.fg("muted", ` [${scope}]`) +
				`\n  ${theme.fg("dim", preview)}`;
			return new Text(text, 0, 0);
		},
		renderResult(result, { expanded, isPartial }, theme) {
			const details = result.details as SubagentDetails | undefined;
			if (!details || details.results.length === 0) {
				const first = result.content[0];
				return new Text(first?.type === "text" ? first.text : "(no output)", 0, 0);
			}

			const mdTheme = getMarkdownTheme();
			const renderDisplayItems = (items: DisplayItem[], limit?: number): string => {
				const visibleItems = limit ? items.slice(-limit) : items;
				const skipped = limit && items.length > limit ? items.length - limit : 0;
				let text = skipped > 0 ? theme.fg("muted", `... ${skipped} earlier items\n`) : "";
				for (const item of visibleItems) {
					if (item.type === "text") {
						const preview = expanded ? item.text : item.text.split("\n").slice(0, 3).join("\n");
						text += `${theme.fg("toolOutput", preview)}\n`;
					} else {
						text += `${theme.fg("muted", "→ ")}${formatToolCall(item.name, item.args, theme.fg.bind(theme))}\n`;
					}
				}
				return text.trimEnd();
			};

			if (details.mode === "single" && details.results.length === 1) {
				const single = details.results[0];
				if (!single) return new Text("(no output)", 0, 0);
				const isRunning = single.exitCode === -1 || (isPartial && single.exitCode === 0);
				const isError = !isRunning && isTaskError(single);
				const icon = isRunning ? theme.fg("warning", "⏳") : isError ? theme.fg("error", "✗") : theme.fg("success", "✓");
				const displayItems = collectDisplayItems(single.messages);
				const finalOutput = collectFinalOutput(single.messages);

				if (expanded) {
					const container = new Container();
					let header = `${icon} ${theme.fg("toolTitle", theme.bold(single.agent))}${theme.fg("muted", ` (${single.agentSource})`)}`;
					if (isRunning) header += ` ${theme.fg("warning", "[running]")}`;
					if (isError && single.stopReason) header += ` ${theme.fg("error", `[${single.stopReason}]`)}`;
					container.addChild(new Text(header, 0, 0));
					const diagnostic = isError ? formatFailureDiagnostic(single) : "";
					if (diagnostic) {
						container.addChild(new Text(theme.fg("error", diagnostic), 0, 0));
					}
					container.addChild(new Spacer(1));
					container.addChild(new Text(theme.fg("muted", "─── Task ───"), 0, 0));
					container.addChild(new Text(theme.fg("dim", single.task), 0, 0));
					container.addChild(new Spacer(1));
					container.addChild(new Text(theme.fg("muted", "─── Output ───"), 0, 0));
					for (const item of displayItems) {
						if (item.type === "toolCall") {
							container.addChild(new Text(theme.fg("muted", "→ ") + formatToolCall(item.name, item.args, theme.fg.bind(theme)), 0, 0));
						}
					}
					if (finalOutput) {
						container.addChild(new Spacer(1));
						container.addChild(new Markdown(finalOutput.trim(), 0, 0, mdTheme));
					} else if (displayItems.length === 0) {
						container.addChild(new Text(theme.fg("muted", isRunning ? "(running...)" : "(no output)"), 0, 0));
					}
					const usage = formatUsageStats(single.usage, single.model);
					if (usage) {
						container.addChild(new Spacer(1));
						container.addChild(new Text(theme.fg("dim", usage), 0, 0));
					}
					return container;
				}

				let text = `${icon} ${theme.fg("toolTitle", theme.bold(single.agent))}${theme.fg("muted", ` (${single.agentSource})`)}`;
				if (isRunning) text += ` ${theme.fg("warning", "[running]")}`;
				if (isError && single.stopReason) text += ` ${theme.fg("error", `[${single.stopReason}]`)}`;
				const diagnostic = isError ? formatFailureDiagnostic(single, true) : "";
				if (diagnostic) {
					text += `\n${theme.fg("error", diagnostic)}`;
				} else if (displayItems.length === 0) {
					text += `\n${theme.fg("muted", isRunning ? "(running...)" : "(no output)")}`;
				} else {
					text += `\n${renderDisplayItems(displayItems, COLLAPSED_ITEM_COUNT)}`;
					if (displayItems.length > COLLAPSED_ITEM_COUNT) text += `\n${expandToolOutputHint(theme)}`;
				}
				const usage = formatUsageStats(single.usage, single.model);
				if (usage) text += `\n${theme.fg("dim", usage)}`;
				return new Text(text, 0, 0);
			}

			if (details.mode === "chain") {
				const successCount = details.results.filter((entry) => entry.exitCode !== -1 && !isTaskError(entry)).length;
				const runningCount = details.results.filter((entry) => entry.exitCode === -1).length;
				const icon = runningCount > 0 ? theme.fg("warning", "⏳") : successCount === details.results.length ? theme.fg("success", "✓") : theme.fg("error", "✗");

				if (expanded) {
					const container = new Container();
					container.addChild(new Text(`${icon} ${theme.fg("toolTitle", theme.bold("chain "))}${theme.fg("accent", `${successCount}/${details.results.length} steps`)}`, 0, 0));
					for (const entry of details.results) {
						const entryRunning = entry.exitCode === -1;
						const entryIcon = entryRunning ? theme.fg("warning", "⏳") : isTaskError(entry) ? theme.fg("error", "✗") : theme.fg("success", "✓");
						const displayItems = collectDisplayItems(entry.messages);
						const finalOutput = collectFinalOutput(entry.messages);
						const diagnostic = !entryRunning && isTaskError(entry) ? formatFailureDiagnostic(entry) : "";
						container.addChild(new Spacer(1));
						container.addChild(new Text(`${theme.fg("muted", `─── Step ${entry.step ?? "?"}: `)}${theme.fg("accent", entry.agent)} ${entryIcon}`, 0, 0));
						container.addChild(new Text(theme.fg("muted", "Task: ") + theme.fg("dim", entry.task), 0, 0));
						if (diagnostic) {
							container.addChild(new Text(theme.fg("error", diagnostic), 0, 0));
						}
						for (const item of displayItems) {
							if (item.type === "toolCall") {
								container.addChild(new Text(theme.fg("muted", "→ ") + formatToolCall(item.name, item.args, theme.fg.bind(theme)), 0, 0));
							}
						}
						if (finalOutput) {
							container.addChild(new Spacer(1));
							container.addChild(new Markdown(finalOutput.trim(), 0, 0, mdTheme));
						} else if (entryRunning) {
							container.addChild(new Text(theme.fg("muted", "(running...)"), 0, 0));
						}
						const usage = formatUsageStats(entry.usage, entry.model);
						if (usage) container.addChild(new Text(theme.fg("dim", usage), 0, 0));
					}
					const totalUsage = formatAggregateUsageStats(details.results);
					if (totalUsage) {
						container.addChild(new Spacer(1));
						container.addChild(new Text(theme.fg("dim", `Total: ${totalUsage}`), 0, 0));
					}
					return container;
				}

				let text = `${icon} ${theme.fg("toolTitle", theme.bold("chain "))}${theme.fg("accent", `${successCount}/${details.results.length} steps`)}`;
				for (const entry of details.results) {
					const entryIcon = entry.exitCode === -1 ? theme.fg("warning", "⏳") : isTaskError(entry) ? theme.fg("error", "✗") : theme.fg("success", "✓");
					const displayItems = collectDisplayItems(entry.messages);
					const diagnostic = entry.exitCode !== -1 && isTaskError(entry) ? formatFailureDiagnostic(entry, true) : "";
					text += `\n\n${theme.fg("muted", `─── Step ${entry.step ?? "?"}: `)}${theme.fg("accent", entry.agent)} ${entryIcon}`;
					if (diagnostic) {
						text += `\n${theme.fg("error", diagnostic)}`;
					} else if (displayItems.length === 0) {
						text += `\n${theme.fg("muted", entry.exitCode === -1 ? "(running...)" : "(no output)")}`;
					} else {
						text += `\n${renderDisplayItems(displayItems, 5)}`;
					}
				}
				const totalUsage = formatAggregateUsageStats(details.results);
				if (totalUsage) text += `\n\n${theme.fg("dim", `Total: ${totalUsage}`)}`;
				text += `\n${expandToolOutputHint(theme)}`;
				return new Text(text, 0, 0);
			}

			if (details.mode === "parallel") {
				const runningCount = details.results.filter((entry) => entry.exitCode === -1).length;
				const successCount = details.results.filter((entry) => entry.exitCode !== -1 && !isTaskError(entry)).length;
				const failCount = details.results.filter((entry) => entry.exitCode !== -1 && isTaskError(entry)).length;
				const isRunning = runningCount > 0;
				const icon = isRunning ? theme.fg("warning", "⏳") : failCount > 0 ? theme.fg("warning", "◐") : theme.fg("success", "✓");
				const status = isRunning ? `${successCount + failCount}/${details.results.length} done, ${runningCount} running` : `${successCount}/${details.results.length} tasks`;

				if (expanded && !isRunning) {
					const container = new Container();
					container.addChild(new Text(`${icon} ${theme.fg("toolTitle", theme.bold("parallel "))}${theme.fg("accent", status)}`, 0, 0));
					for (const entry of details.results) {
						const entryIcon = isTaskError(entry) ? theme.fg("error", "✗") : theme.fg("success", "✓");
						const displayItems = collectDisplayItems(entry.messages);
						const finalOutput = collectFinalOutput(entry.messages);
						const diagnostic = isTaskError(entry) ? formatFailureDiagnostic(entry) : "";
						container.addChild(new Spacer(1));
						container.addChild(new Text(`${theme.fg("muted", "─── ")}${theme.fg("accent", entry.agent)} ${entryIcon}`, 0, 0));
						container.addChild(new Text(theme.fg("muted", "Task: ") + theme.fg("dim", entry.task), 0, 0));
						if (diagnostic) {
							container.addChild(new Text(theme.fg("error", diagnostic), 0, 0));
						}
						for (const item of displayItems) {
							if (item.type === "toolCall") {
								container.addChild(new Text(theme.fg("muted", "→ ") + formatToolCall(item.name, item.args, theme.fg.bind(theme)), 0, 0));
							}
						}
						if (finalOutput) {
							container.addChild(new Spacer(1));
							container.addChild(new Markdown(finalOutput.trim(), 0, 0, mdTheme));
						}
						const usage = formatUsageStats(entry.usage, entry.model);
						if (usage) container.addChild(new Text(theme.fg("dim", usage), 0, 0));
					}
					const totalUsage = formatAggregateUsageStats(details.results);
					if (totalUsage) {
						container.addChild(new Spacer(1));
						container.addChild(new Text(theme.fg("dim", `Total: ${totalUsage}`), 0, 0));
					}
					return container;
				}

				let text = `${icon} ${theme.fg("toolTitle", theme.bold("parallel "))}${theme.fg("accent", status)}`;
				for (const entry of details.results) {
					const entryIcon = entry.exitCode === -1 ? theme.fg("warning", "⏳") : isTaskError(entry) ? theme.fg("error", "✗") : theme.fg("success", "✓");
					const displayItems = collectDisplayItems(entry.messages);
					const diagnostic = entry.exitCode !== -1 && isTaskError(entry) ? formatFailureDiagnostic(entry, true) : "";
					text += `\n\n${theme.fg("muted", "─── ")}${theme.fg("accent", entry.agent)} ${entryIcon}`;
					if (diagnostic) {
						text += `\n${theme.fg("error", diagnostic)}`;
					} else if (displayItems.length === 0) {
						text += `\n${theme.fg("muted", entry.exitCode === -1 ? "(running...)" : "(no output)")}`;
					} else {
						text += `\n${renderDisplayItems(displayItems, 5)}`;
					}
				}
				if (!isRunning) {
					const totalUsage = formatAggregateUsageStats(details.results);
					if (totalUsage) text += `\n\n${theme.fg("dim", `Total: ${totalUsage}`)}`;
				}
				if (!expanded) text += `\n${expandToolOutputHint(theme)}`;
				return new Text(text, 0, 0);
			}

			const first = result.content[0];
			return new Text(first?.type === "text" ? first.text : "(no output)", 0, 0);
		},
	});
}
