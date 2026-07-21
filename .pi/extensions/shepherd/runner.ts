export type ShepherdThinkingLevel = "high" | "xhigh";

export interface AgentBinding {
	runId: string;
	generation: number;
	laneId: string;
	candidateHead: string;
	validationNonce: string;
	readOnly: boolean;
	provider: string;
	model: string;
	thinking: ShepherdThinkingLevel;
}

export interface DimensionScores {
	correctStage: number;
	artifactValid: number;
	gatesRespected: number;
	realProgress: number;
	noHallucination: number;
	noConflict: number;
}

export interface AgentRunRequest {
	runId: string;
	laneId: string;
	role: string;
	cwd: string;
	readOnly: boolean;
	provider: string;
	model: string;
	thinking: ShepherdThinkingLevel;
	tools: string[];
	systemPrompt: string;
	prompt: string;
	timeoutMs: number;
	signal?: AbortSignal;
	binding: AgentBinding;
}

export interface AgentRunResult extends AgentBinding {
	summary: string;
	dimensions: DimensionScores;
	observedMutation: boolean;
}

/** Port used by the deterministic supervisor. Implementations own their child lifecycle. */
export interface AgentRunner {
	run(request: AgentRunRequest): Promise<AgentRunResult>;
	abort(runId: string): Promise<void>;
	close(): Promise<void>;
}
