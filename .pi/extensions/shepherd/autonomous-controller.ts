import type { ShepherdResumeCommand, ShepherdStartCommand } from "./arguments.ts";
import type {
	AutonomousChildPlan,
	AutonomousShepherdRunState,
	AutonomousStateStore,
} from "./autonomous-state.ts";

export interface AutonomousParentPlan {
	planId: string;
	children: AutonomousChildPlan[];
}

export interface AutonomousIntakePort {
	load(issue: number, signal: AbortSignal): Promise<AutonomousParentPlan>;
}

export interface AutonomousChildContext {
	parentIssue: number;
	runId: string;
	generation: number;
	child: AutonomousChildPlan;
	signal: AbortSignal;
}

export interface AutonomousStageResult {
	summary: string;
}

export interface AutonomousChildLifecyclePort {
	execute(context: AutonomousChildContext): Promise<AutonomousStageResult>;
	verify(context: AutonomousChildContext): Promise<AutonomousStageResult>;
	review(context: AutonomousChildContext): Promise<AutonomousStageResult>;
	integrate(context: AutonomousChildContext): Promise<AutonomousStageResult>;
	abort(runId: string): Promise<void>;
	close(): Promise<void>;
}

export interface AutonomousHumanGatePort {
	request(state: AutonomousShepherdRunState, signal: AbortSignal): Promise<{ requestId: string }>;
	observe(state: AutonomousShepherdRunState, signal: AbortSignal): Promise<"pending" | "merged" | "rejected">;
	close(): Promise<void>;
}

export interface AutonomousShepherdControllerOptions {
	store: AutonomousStateStore;
	intake: AutonomousIntakePort;
	lifecycle: AutonomousChildLifecyclePort;
	humanGate: AutonomousHumanGatePort;
	now?: () => Date;
	newRunId?: () => string;
}

/** Test-first contract scaffold; the RED trajectory drives the implementation. */
export class AutonomousShepherdController {
	constructor(_options: AutonomousShepherdControllerOptions) {}

	async status(_issue: number): Promise<AutonomousShepherdRunState | undefined> {
		throw new Error("autonomous Shepherd MVP is not implemented");
	}

	async start(_command: ShepherdStartCommand): Promise<AutonomousShepherdRunState> {
		throw new Error("autonomous Shepherd MVP is not implemented");
	}

	async resume(_command: ShepherdResumeCommand): Promise<AutonomousShepherdRunState> {
		throw new Error("autonomous Shepherd MVP is not implemented");
	}

	async stop(_issue: number): Promise<AutonomousShepherdRunState> {
		throw new Error("autonomous Shepherd MVP is not implemented");
	}

	async shutdown(): Promise<void> {}
}
