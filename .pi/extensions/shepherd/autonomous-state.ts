import { mkdir, readFile, rename, writeFile } from "node:fs/promises";
import { join } from "node:path";

import type { ParentLifecycleStage } from "./autonomy-policy.ts";
import type { WorkAccess, WorkItemStatus } from "./dependency-graph.ts";

export type AutonomousChildPhase =
	| "pending"
	| "execute"
	| "verify"
	| "review"
	| "integrate"
	| "succeeded"
	| "failed";

export type AutonomousRunStatus = "running" | "stopped" | "failed" | "waiting_human" | "completed";

export interface AutonomousChildPlan {
	id: string;
	issue: number;
	title: string;
	task: string;
	dependsOn: string[];
	access: WorkAccess;
	writeScopes: string[];
}

export interface AutonomousChildState extends AutonomousChildPlan {
	status: WorkItemStatus;
	phase: AutonomousChildPhase;
	summary?: string;
}

export interface AutonomousHumanGate {
	kind: "parent_merge";
	requestId: string;
	status: "pending" | "merged" | "rejected";
}

export interface AutonomousShepherdRunState {
	schemaVersion: 2;
	kind: "autonomous";
	issue: number;
	pr?: number;
	planId: string;
	runId: string;
	generation: number;
	status: AutonomousRunStatus;
	stage: ParentLifecycleStage;
	maxConcurrency: number;
	createdAt: string;
	updatedAt: string;
	children: AutonomousChildState[];
	humanGate?: AutonomousHumanGate;
}

export interface AutonomousStateStore {
	load(issue: number): Promise<AutonomousShepherdRunState | undefined>;
	save(state: AutonomousShepherdRunState): Promise<void>;
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null && !Array.isArray(value);
}

function cloneState(state: AutonomousShepherdRunState): AutonomousShepherdRunState {
	return structuredClone(state);
}

export function validateAutonomousState(value: unknown): AutonomousShepherdRunState {
	if (!isRecord(value) || value.schemaVersion !== 2 || value.kind !== "autonomous") {
		throw new Error("invalid autonomous Shepherd state");
	}
	if (!Number.isSafeInteger(value.issue) || (value.issue as number) < 1
		|| !Number.isSafeInteger(value.generation) || (value.generation as number) < 1
		|| !Number.isSafeInteger(value.maxConcurrency) || (value.maxConcurrency as number) < 1
		|| typeof value.planId !== "string" || typeof value.runId !== "string"
		|| typeof value.createdAt !== "string" || typeof value.updatedAt !== "string"
		|| !Array.isArray(value.children)) {
		throw new Error("invalid autonomous Shepherd state fields");
	}
	return cloneState(value as unknown as AutonomousShepherdRunState);
}

/** Separate v2 file path keeps the hardened v1 canary state independently readable. */
export class AutonomousFileStateStore implements AutonomousStateStore {
	readonly #root: string;

	constructor(root: string) {
		this.#root = root;
	}

	#path(issue: number): string {
		if (!Number.isSafeInteger(issue) || issue < 1) throw new Error("issue must be a positive integer");
		return join(this.#root, `autonomous-issue-${issue}.json`);
	}

	async load(issue: number): Promise<AutonomousShepherdRunState | undefined> {
		try {
			return validateAutonomousState(JSON.parse(await readFile(this.#path(issue), "utf8")));
		} catch (error) {
			if (isRecord(error) && error.code === "ENOENT") return undefined;
			throw error;
		}
	}

	async save(state: AutonomousShepherdRunState): Promise<void> {
		const snapshot = validateAutonomousState(state);
		await mkdir(this.#root, { recursive: true, mode: 0o700 });
		const target = this.#path(snapshot.issue);
		const temporary = `${target}.${process.pid}.${crypto.randomUUID()}.tmp`;
		await writeFile(temporary, `${JSON.stringify(snapshot, null, 2)}\n`, { mode: 0o600, flag: "wx" });
		await rename(temporary, target);
	}
}
