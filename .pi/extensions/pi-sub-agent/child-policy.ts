const SUBAGENT_TOOL_NAME = "subagent";
export const SUBAGENT_DEPTH_ENV = "PI_SUB_AGENT_DEPTH";
const MAX_SUBAGENT_DEPTH = 1;

export function getSubagentDepth(value = process.env[SUBAGENT_DEPTH_ENV]): number {
	if (value === undefined) return 0;
	const parsed = Number.parseInt(value, 10);
	return Number.isFinite(parsed) && parsed > 0 ? parsed : 0;
}

export function canStartSubagent(value = process.env[SUBAGENT_DEPTH_ENV]): boolean {
	return getSubagentDepth(value) < MAX_SUBAGENT_DEPTH;
}

function normalizeToolNames(toolNames: readonly string[] | undefined): string[] | undefined {
	if (toolNames === undefined) return undefined;
	return Array.from(new Set(toolNames.map((tool) => tool.trim()).filter((tool) => Boolean(tool) && tool !== SUBAGENT_TOOL_NAME)));
}

export function resolveChildToolAllowlist(
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

export function childInvocationArgs(subSessionDir = process.env.PI_SUBAGENT_SESSION_DIR): string[] {
	const args = subSessionDir
		? ["--mode", "json", "-p", "--session-dir", subSessionDir]
		: ["--mode", "json", "-p", "--no-session"];
	args.push("--no-extensions");
	return args;
}
