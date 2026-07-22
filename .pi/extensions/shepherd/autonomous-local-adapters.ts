import { createHash } from "node:crypto";
import { execFile } from "node:child_process";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { dirname, isAbsolute, relative, resolve } from "node:path";
import { promisify } from "node:util";

import type {
	AutonomousChildContext,
	AutonomousChildLifecyclePort,
	AutonomousHumanGatePort,
	AutonomousIntakePort,
	AutonomousParentPlan,
	AutonomousStageResult,
} from "./autonomous-controller.ts";
import type { AutonomousChildPlan, AutonomousShepherdRunState } from "./autonomous-state.ts";
import { ShepherdAgentSessionRuntime } from "./agent-session-runtime.ts";
import type { ShepherdAgentRole } from "./role-prompts.ts";
import type { ScopedWorkspace, WorkspaceMutationResult } from "./tool-policy.ts";

const execFileAsync = promisify(execFile);

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null && !Array.isArray(value);
}

function text(value: unknown, field: string): string {
	if (typeof value !== "string" || value.trim() === "" || value.length > 48 * 1024) {
		throw new Error(`autonomous plan ${field} must be bounded text`);
	}
	return value;
}

function positiveInteger(value: unknown, field: string): number {
	if (!Number.isSafeInteger(value) || (value as number) < 1) {
		throw new Error(`autonomous plan ${field} must be a positive integer`);
	}
	return value as number;
}

function textArray(value: unknown, field: string): string[] {
	if (!Array.isArray(value) || value.length > 64 || !value.every((item) => typeof item === "string")) {
		throw new Error(`autonomous plan ${field} must be a bounded text array`);
	}
	return [...value];
}

function childPlan(value: unknown): AutonomousChildPlan {
	if (!isRecord(value)) throw new Error("autonomous plan child must be an object");
	if (value.access !== "mutating") throw new Error("autonomous MVP children must be scoped mutating lanes");
	return {
		id: text(value.id, "child id"),
		issue: positiveInteger(value.issue, "child issue"),
		title: text(value.title, "child title"),
		task: text(value.task, "child task"),
		dependsOn: textArray(value.dependsOn, "child dependencies"),
		access: "mutating",
		writeScopes: textArray(value.writeScopes, "child write scopes"),
	};
}

/** Repository-local bounded plan intake used by the first testable MVP. */
export class RepositoryManifestIntake implements AutonomousIntakePort {
	readonly #cwd: string;

	constructor(cwd: string) {
		this.#cwd = cwd;
	}

	async load(issue: number, signal: AbortSignal): Promise<AutonomousParentPlan> {
		if (signal.aborted) throw new Error("autonomous plan intake was cancelled");
		const path = resolve(this.#cwd, ".planning", "shepherd", `issue-${issue}.json`);
		let parsed: unknown;
		try {
			parsed = JSON.parse(await readFile(path, "utf8"));
		} catch (error) {
			throw new Error(
				`autonomous MVP plan is unavailable; create .planning/shepherd/issue-${issue}.json`,
				{ cause: error },
			);
		}
		if (!isRecord(parsed) || parsed.schemaVersion !== 1 || parsed.parentIssue !== issue
			|| !Array.isArray(parsed.children) || parsed.children.length < 1 || parsed.children.length > 64) {
			throw new Error("autonomous MVP plan has an invalid root contract");
		}
		return {
			planId: text(parsed.planId, "planId"),
			children: parsed.children.map(childPlan),
		};
	}
}

class RepositoryScopedWorkspace implements ScopedWorkspace {
	readonly id: string;
	readonly cwd: string;

	constructor(id: string, cwd: string) {
		this.id = id;
		this.cwd = cwd;
	}

	#resolve(path: string): string {
		if (typeof path !== "string" || path === "" || isAbsolute(path) || path.includes("\\")) {
			throw new Error("workspace path must be relative");
		}
		const target = resolve(this.cwd, path);
		const back = relative(this.cwd, target);
		if (back === "" || back === ".." || back.startsWith(`..${process.platform === "win32" ? "\\" : "/"}`)) {
			throw new Error("workspace path escapes the repository");
		}
		return target;
	}

	async readText(path: string, options: { offset?: number; limit?: number; signal?: AbortSignal }): Promise<string> {
		if (options.signal?.aborted) throw new Error("workspace read was cancelled");
		const value = await readFile(this.#resolve(path), "utf8");
		const offset = options.offset ?? 0;
		return value.slice(offset, options.limit === undefined ? undefined : offset + options.limit);
	}

	async editText(path: string, oldText: string, newText: string, signal?: AbortSignal): Promise<WorkspaceMutationResult> {
		if (signal?.aborted) throw new Error("workspace edit was cancelled");
		const target = this.#resolve(path);
		const current = await readFile(target, "utf8");
		const first = current.indexOf(oldText);
		if (first < 0 || current.indexOf(oldText, first + oldText.length) >= 0) {
			throw new Error("workspace edit requires one exact oldText match");
		}
		const next = `${current.slice(0, first)}${newText}${current.slice(first + oldText.length)}`;
		await writeFile(target, next, "utf8");
		return { changed: next !== current, summary: next === current ? "unchanged" : "updated one file" };
	}

	async writeText(path: string, content: string, signal?: AbortSignal): Promise<WorkspaceMutationResult> {
		if (signal?.aborted) throw new Error("workspace write was cancelled");
		const target = this.#resolve(path);
		let current: string | undefined;
		try { current = await readFile(target, "utf8"); } catch { /* New file. */ }
		await mkdir(dirname(target), { recursive: true });
		await writeFile(target, content, "utf8");
		return { changed: current !== content, summary: current === content ? "unchanged" : "wrote one file" };
	}
}

async function gitValue(cwd: string, args: string[]): Promise<string> {
	const { stdout } = await execFileAsync("git", ["-C", cwd, ...args], { encoding: "utf8" });
	return stdout.trim();
}

/** Runs the actual implementation, verification, and review roles in embedded Pi AgentSessions. */
export class AgentSessionMvpLifecycle implements AutonomousChildLifecyclePort {
	readonly #runtime: ShepherdAgentSessionRuntime;
	readonly #cwd: string;

	constructor(runtime: ShepherdAgentSessionRuntime, cwd: string) {
		this.#runtime = runtime;
		this.#cwd = cwd;
	}

	async #runRole(context: AutonomousChildContext, role: ShepherdAgentRole, readOnly: boolean): Promise<AutonomousStageResult> {
		const [head, branch] = await Promise.all([
			gitValue(this.#cwd, ["rev-parse", "HEAD"]),
			gitValue(this.#cwd, ["branch", "--show-current"]),
		]);
		const laneId = `${context.child.id}-${role}`;
		const workspaceId = `issue-${context.child.issue}-${context.child.id}`;
		const binding = {
			runId: context.runId,
			generation: context.generation,
			laneId,
			candidateHead: head,
			validationNonce: createHash("sha256").update(`${context.runId}:${context.generation}:${laneId}:${head}`).digest("hex"),
		};
		const task = role === "implementation"
			? `${context.child.task}\nUse strict red-green-refactor and stay inside the declared write scopes.`
			: role === "verification"
				? `Verify the completed objective without modifying files: ${context.child.task}`
				: `Independently review the completed objective and report blocking findings only: ${context.child.task}`;
		const handoff = await this.#runtime.run({
			role,
			task,
			context: [
				`Parent issue #${context.parentIssue}; child issue #${context.child.issue}.`,
				`Declared write scopes: ${context.child.writeScopes.join(", ")}.`,
			],
			timeoutMs: context.timeoutMs,
			signal: context.signal,
			workspace: new RepositoryScopedWorkspace(workspaceId, this.#cwd),
			capabilities: [],
			authority: {
				issue: context.child.issue,
				branch,
				readOnly,
				workspaceId,
				readPrefixes: ["."],
				writePrefixes: readOnly ? [] : [...context.child.writeScopes],
				capabilityNames: [],
			},
			binding,
		});
		if (handoff.status !== "completed") throw new Error(`${role} AgentSession did not complete: ${handoff.summary}`);
		return { summary: handoff.summary };
	}

	execute(context: AutonomousChildContext): Promise<AutonomousStageResult> {
		return this.#runRole(context, "implementation", false);
	}

	verify(context: AutonomousChildContext): Promise<AutonomousStageResult> {
		return this.#runRole(context, "verification", true);
	}

	review(context: AutonomousChildContext): Promise<AutonomousStageResult> {
		return this.#runRole(context, "review", true);
	}

	async integrate(_context: AutonomousChildContext): Promise<AutonomousStageResult> {
		return { summary: "MVP lifecycle accepted after implementation, verification, and independent review" };
	}

	abort(runId: string): Promise<void> {
		return this.#runtime.abort(runId);
	}

	close(): Promise<void> {
		return this.#runtime.close();
	}
}

/** Local MVP gate: durable state waits for the human; it never exposes a merge effect. */
export class LocalParentMergeGate implements AutonomousHumanGatePort {
	async request(state: AutonomousShepherdRunState, _signal: AbortSignal): Promise<{ requestId: string }> {
		return { requestId: `local-parent-merge-${state.issue}-${state.generation}` };
	}

	async observe(_state: AutonomousShepherdRunState, _signal: AbortSignal): Promise<"pending"> {
		return "pending";
	}

	async close(): Promise<void> {}
}
