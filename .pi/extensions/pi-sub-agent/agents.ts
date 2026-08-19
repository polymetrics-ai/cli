/**
 * Agent discovery and configuration
 */

import * as fs from "node:fs";
import * as path from "node:path";
import { execFileSync } from "node:child_process";
import { getAgentDir, parseFrontmatter } from "@earendil-works/pi-coding-agent";

export type AgentScope = "clean-project" | "user" | "project" | "both";

const CLEAN_PROJECT_SCOPE = "clean-project";
const AGENT_CONTRACT_CHECK_TIMEOUT_MS = 30_000;

export function shouldConfirmProjectAgents(scope: AgentScope): boolean {
	return scope === CLEAN_PROJECT_SCOPE || scope === "project" || scope === "both";
}

export const THINKING_LEVELS = ["off", "minimal", "low", "medium", "high", "xhigh"] as const;

export type AgentThinkingLevel = (typeof THINKING_LEVELS)[number];

export interface AgentConfig {
	name: string;
	description: string;
	tools?: string[];
	model?: string;
	thinking?: AgentThinkingLevel;
	systemPrompt: string;
	source: "user" | "project";
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

function loadAgentFile(filePath: string, source: "user" | "project"): AgentConfig | undefined {
	let content: string;
	try {
		content = fs.readFileSync(filePath, "utf-8");
	} catch {
		return undefined;
	}

	let parsed: { frontmatter: Record<string, unknown>; body: string };
	try {
		parsed = parseFrontmatter<Record<string, unknown>>(content);
	} catch {
		return undefined;
	}

	const { frontmatter, body } = parsed;
	const name = frontmatterString(frontmatter.name);
	const description = frontmatterString(frontmatter.description);
	if (!name || !description) return undefined;

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
	if (tools) agent.tools = tools;
	if (model) agent.model = model;
	if (thinking) agent.thinking = thinking;
	return agent;
}

function loadAgentsFromDir(dir: string, source: "user" | "project"): AgentConfig[] {
	if (!fs.existsSync(dir)) return [];

	let entries: fs.Dirent[];
	try {
		entries = fs.readdirSync(dir, { withFileTypes: true });
	} catch {
		return [];
	}

	const agents: AgentConfig[] = [];
	for (const entry of entries) {
		if (!entry.name.endsWith(".md") || (!entry.isFile() && !entry.isSymbolicLink())) continue;
		const agent = loadAgentFile(path.join(dir, entry.name), source);
		if (agent) agents.push(agent);
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

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null;
}

function isSafeAgentFileStem(value: string): boolean {
	return /^[a-z0-9][a-z0-9-]*$/.test(value);
}

function isRealDirectory(directory: string): boolean {
	try {
		const stats = fs.lstatSync(directory);
		return stats.isDirectory() && !stats.isSymbolicLink();
	} catch {
		return false;
	}
}

function isRegularFile(filePath: string): boolean {
	try {
		const stats = fs.lstatSync(filePath);
		return stats.isFile() && !stats.isSymbolicLink();
	} catch {
		return false;
	}
}

function generatedAgentContractIsCurrent(projectRoot: string): boolean {
	try {
		execFileSync("go", ["run", "./cmd/agentcontractgen", "check", "--root", projectRoot], {
			cwd: projectRoot,
			stdio: "ignore",
			timeout: AGENT_CONTRACT_CHECK_TIMEOUT_MS,
			windowsHide: true,
		});
		return true;
	} catch {
		return false;
	}
}

function canonicalCleanProjectRoles(projectRoot: string): string[] {
	const canonicalPath = path.join(projectRoot, ".agents", "agentic-delivery", "canonical", "delivery-contract.json");
	let value: unknown;
	try {
		value = JSON.parse(fs.readFileSync(canonicalPath, "utf-8"));
	} catch {
		return [];
	}
	if (!isRecord(value)) return [];
	const baseRole = isRecord(value.base_role) ? value.base_role : undefined;
	const connectorOverlay = isRecord(value.connector_overlay) ? value.connector_overlay : undefined;
	const piHarness = isRecord(value.pi_harness) ? value.pi_harness : undefined;
	const baseName = baseRole?.name;
	const connectorName = connectorOverlay?.name;
	const scope = piHarness?.clean_project_scope;
	const roles = piHarness?.roles;
	if (
		scope !== CLEAN_PROJECT_SCOPE ||
		typeof baseName !== "string" ||
		typeof connectorName !== "string" ||
		!isSafeAgentFileStem(baseName) ||
		!isSafeAgentFileStem(connectorName) ||
		baseName === connectorName ||
		!Array.isArray(roles) ||
		roles.length !== 2 ||
		roles[0] !== baseName ||
		roles[1] !== connectorName ||
		roles.some((role) => typeof role !== "string" || !role.trim())
	) {
		return [];
	}
	return [baseName, connectorName];
}

function loadCleanProjectAgents(projectAgentsDir: string): AgentConfig[] {
	const projectRoot = path.dirname(path.dirname(projectAgentsDir));
	const canonicalDir = path.join(projectRoot, ".agents", "agentic-delivery", "canonical");
	const canonicalPath = path.join(canonicalDir, "delivery-contract.json");
	for (const directory of [
		path.join(projectRoot, ".pi"),
		projectAgentsDir,
		path.join(projectRoot, ".agents"),
		path.join(projectRoot, ".agents", "agentic-delivery"),
		canonicalDir,
	]) {
		if (!isRealDirectory(directory)) return [];
	}
	if (!isRegularFile(canonicalPath) || !generatedAgentContractIsCurrent(projectRoot)) return [];

	const roles = canonicalCleanProjectRoles(projectRoot);
	if (roles.length !== 2) return [];

	const agents: AgentConfig[] = [];
	for (const role of roles) {
		const filePath = path.join(projectAgentsDir, `${role}.md`);
		if (!isRegularFile(filePath)) return [];
		const agent = loadAgentFile(filePath, "project");
		if (!agent || agent.name !== role) return [];
		agents.push(agent);
	}
	return agents;
}

export function discoverAgents(cwd: string, scope: AgentScope): AgentDiscoveryResult {
	const projectAgentsDir = findNearestProjectAgentsDir(cwd);
	const userAgents = scope === "project" || scope === CLEAN_PROJECT_SCOPE
		? []
		: loadAgentsFromDir(path.join(getAgentDir(), "agents"), "user");
	const projectAgents = scope === "user" || !projectAgentsDir
		? []
		: scope === CLEAN_PROJECT_SCOPE
			? loadCleanProjectAgents(projectAgentsDir)
			: loadAgentsFromDir(projectAgentsDir, "project");

	if (scope === CLEAN_PROJECT_SCOPE) {
		return { agents: projectAgents, projectAgentsDir };
	}

	const agentMap = new Map<string, AgentConfig>();
	for (const agent of userAgents) agentMap.set(agent.name, agent);
	for (const agent of projectAgents) agentMap.set(agent.name, agent);

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
