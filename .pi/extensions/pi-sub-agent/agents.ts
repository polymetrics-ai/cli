/**
 * Agent discovery and configuration
 */

import * as fs from "node:fs";
import * as path from "node:path";
import { getAgentDir, parseFrontmatter } from "@earendil-works/pi-coding-agent";

export type AgentScope = "user" | "project" | "both";

export const THINKING_LEVELS = ["off", "minimal", "low", "medium", "high", "xhigh"] as const;

export type AgentThinkingLevel = (typeof THINKING_LEVELS)[number];

export interface AgentConfig {
	name: string;
	description: string;
	tools?: string[];
	model?: string;
	thinking?: AgentThinkingLevel;
	systemPrompt: string;
	source: "user" | "project" | "extension";
	filePath: string;
}
export interface AgentDiscoveryResult {
	agents: AgentConfig[];
	projectAgentsDir: string | null;
}

function frontmatterString(value: unknown): string | undefined {
	if (typeof value !== "string") return undefined;
	const trimmed = value.trim();
	return trimmed ? trimmed : undefined;
}

function frontmatterThinking(value: unknown): AgentThinkingLevel | undefined {
	if (typeof value !== "string") return undefined;
	const trimmed = value.trim();
	return isThinkingLevel(trimmed) ? trimmed : undefined;
}

function isThinkingLevel(value: string): value is AgentThinkingLevel {
	return (THINKING_LEVELS as readonly string[]).includes(value);
}

export function splitModelThinking(model: string): { model: string; thinking?: AgentThinkingLevel } {
	const match = model.match(/:(off|minimal|low|medium|high|xhigh)$/);
	if (!match) return { model };
	const thinking = match[1];
	if (!thinking || !isThinkingLevel(thinking)) return { model };
	return { model: model.slice(0, -thinking.length - 1), thinking };
}

export function formatModelWithThinking(model: string | undefined, thinking: AgentThinkingLevel | undefined): string | undefined {
	if (!model) return undefined;
	const base = splitModelThinking(model).model;
	if (!thinking) return base;
	return `${base}:${thinking}`;
}

export function resolveAgentModel(agent: Pick<AgentConfig, "model" | "thinking">, fallbackModel: string | undefined): string | undefined {
	const fallbackThinking = fallbackModel ? splitModelThinking(fallbackModel).thinking : undefined;
	const thinking = agent.thinking ?? fallbackThinking;
	if (agent.model) return formatModelWithThinking(agent.model, thinking);
	if (agent.thinking) return formatModelWithThinking(fallbackModel, agent.thinking);
	return fallbackModel;
}

export interface AgentSettingsUpdate {
	model?: string | null;
	thinking?: AgentThinkingLevel | null;
}

export function updateAgentSettingsContent(content: string, update: AgentSettingsUpdate): string {
	const match = content.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n?/);
	if (!match) throw new Error("Agent file is missing YAML frontmatter");
	const frontmatter = match[1] ?? "";
	const body = content.slice(match[0].length);
	const lines = frontmatter.split(/\r?\n/).filter((line) => !/^\s*(model|thinking)\s*:/.test(line));

	if (update.model !== undefined && update.model !== null && update.model.trim()) {
		lines.push(`model: ${splitModelThinking(update.model.trim()).model}`);
	}
	if (update.thinking !== undefined && update.thinking !== null) {
		lines.push(`thinking: ${update.thinking}`);
	}

	return `---\n${lines.join("\n")}\n---\n${body}`;
}

function parseTools(value: unknown): string[] | undefined {
	const rawTools = Array.isArray(value) ? value : typeof value === "string" ? value.split(",") : [];
	const tools = rawTools
		.filter((tool): tool is string => typeof tool === "string")
		.map((tool) => tool.trim())
		.filter(Boolean);
	return tools.length > 0 ? tools : undefined;
}

function loadAgentsFromDir(dir: string, source: "user" | "project" | "extension"): AgentConfig[] {
	const agents: AgentConfig[] = [];

	if (!fs.existsSync(dir)) {
		return agents;
	}

	let entries: fs.Dirent[];
	try {
		entries = fs.readdirSync(dir, { withFileTypes: true });
	} catch {
		return agents;
	}

	for (const entry of entries) {
		if (!entry.name.endsWith(".md")) continue;
		if (!entry.isFile() && !entry.isSymbolicLink()) continue;

		const filePath = path.join(dir, entry.name);
		let content: string;
		try {
			content = fs.readFileSync(filePath, "utf-8");
		} catch {
			continue;
		}

		let parsed: { frontmatter: Record<string, unknown>; body: string };
		try {
			parsed = parseFrontmatter<Record<string, unknown>>(content);
		} catch {
			continue;
		}

		const { frontmatter, body } = parsed;
		const name = frontmatterString(frontmatter.name);
		const description = frontmatterString(frontmatter.description);

		if (!name || !description) {
			continue;
		}

		const tools = parseTools(frontmatter.tools);
		const rawModel = frontmatterString(frontmatter.model);
		const parsedModel = rawModel ? splitModelThinking(rawModel) : undefined;
		const model = parsedModel?.model;
		const thinking = frontmatterThinking(frontmatter.thinking) ?? parsedModel?.thinking;

		const agent: AgentConfig = {
			name,
			description,
			systemPrompt: body,
			source,
			filePath,
		};
		if (tools) {
			agent.tools = tools;
		}
		if (model) {
			agent.model = model;
		}
		if (thinking) {
			agent.thinking = thinking;
		}
		agents.push(agent);
	}

	return agents;
}

function isDirectory(p: string): boolean {
	try {
		return fs.statSync(p).isDirectory();
	} catch {
		return false;
	}
}

function findNearestProjectAgentsDir(cwd: string): string | null {
	let currentDir = cwd;
	while (true) {
		const candidate = path.join(currentDir, ".pi", "agents");
		if (isDirectory(candidate)) return candidate;

		const parentDir = path.dirname(currentDir);
		if (parentDir === currentDir) return null;
		currentDir = parentDir;
	}
}

export function discoverAgents(cwd: string, scope: AgentScope, extensionAgentsDir?: string): AgentDiscoveryResult {
	const userDir = path.join(getAgentDir(), "agents");
	const projectAgentsDir = findNearestProjectAgentsDir(cwd);
	const extensionAgents = extensionAgentsDir ? loadAgentsFromDir(extensionAgentsDir, "extension") : [];

	const userAgents = scope === "project" ? [] : loadAgentsFromDir(userDir, "user");
	const projectAgents = scope === "user" || !projectAgentsDir ? [] : loadAgentsFromDir(projectAgentsDir, "project");

	const agentMap = new Map<string, AgentConfig>();

	for (const agent of extensionAgents) agentMap.set(agent.name, agent);
	if (scope !== "project") {
		for (const agent of userAgents) agentMap.set(agent.name, agent);
	}
	if (scope !== "user") {
		for (const agent of projectAgents) agentMap.set(agent.name, agent);
	}

	return { agents: Array.from(agentMap.values()), projectAgentsDir };
}

export function formatAgentList(agents: AgentConfig[], maxItems: number): { text: string; remaining: number } {
	if (agents.length === 0) return { text: "none", remaining: 0 };
	const listed = agents.slice(0, maxItems);
	const remaining = agents.length - listed.length;
	return {
		text: listed.map((a) => `${a.name} (${a.source}): ${a.description}`).join("; "),
		remaining,
	};
}
