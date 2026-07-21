export const SHEPHERD_SCHEMA_VERSION = 1;
export const SHEPHERD_PROCEED_THRESHOLD = 0.8;

export interface DimensionScores {
	correctStage: number;
	artifactValid: number;
	gatesRespected: number;
	realProgress: number;
	noHallucination: number;
	noConflict: number;
}

export interface LaneBinding {
	runId: string;
	generation: number;
	laneId: string;
	candidateHead: string;
	validationNonce: string;
	readOnly: boolean;
	provider: string;
	model: string;
	thinking: string;
}

export interface LaneEvidence extends LaneBinding {
	summary: string;
	dimensions: DimensionScores;
	observedMutation: boolean;
}

export type LaneStatus = "pending" | "running" | "succeeded" | "failed" | "interrupted" | "stopped" | "halted";
export type ShepherdRunStatus = "pending" | "running" | "completed" | "failed" | "interrupted" | "stopped" | "halted";

export interface LaneDefinition {
	id: string;
	mutating: boolean;
	dependsOn: string[];
}

export interface ShepherdLaneState extends LaneDefinition {
	role: string;
	status: LaneStatus;
	summary?: string;
	score?: number;
	hardGates?: string[];
}

export interface ShepherdRunState {
	schemaVersion: number;
	issue: number;
	pr?: number;
	runId: string;
	generation: number;
	status: ShepherdRunStatus;
	candidateHead: string;
	validationNonce: string;
	createdAt: string;
	updatedAt: string;
	lanes: ShepherdLaneState[];
	score?: number;
	hardGates?: string[];
}

export type LaneDecision = "proceed" | "correct" | "halt";

export interface LaneAssessment {
	decision: LaneDecision;
	score: number;
	hardGates: string[];
}

const dimensionNames = [
	"correctStage",
	"artifactValid",
	"gatesRespected",
	"realProgress",
	"noHallucination",
	"noConflict",
] as const;

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null;
}

export function rateDimensions(dimensions: DimensionScores): number {
	if (!isRecord(dimensions)) throw new TypeError("dimension scores must be an object");
	let product = 1;
	for (const name of dimensionNames) {
		const score = dimensions[name];
		if (typeof score !== "number" || !Number.isFinite(score) || score < 0 || score > 1) {
			throw new RangeError(`dimension ${name} must be between 0 and 1`);
		}
		product *= score;
	}
	return Math.pow(product, 1 / dimensionNames.length);
}

export function assessLaneEvidence(binding: LaneBinding, evidence: LaneEvidence): LaneAssessment {
	const hardGates: string[] = [];
	if (!isRecord(evidence)) {
		return { decision: "halt", score: 0, hardGates: ["artifact_invalid"] };
	}

	if (evidence.runId !== binding.runId) hardGates.push("run_identity_mismatch");
	if (evidence.generation !== binding.generation) hardGates.push("generation_mismatch");
	if (evidence.laneId !== binding.laneId) hardGates.push("lane_identity_mismatch");
	if (evidence.candidateHead !== binding.candidateHead) hardGates.push("stale_head");
	if (evidence.validationNonce !== binding.validationNonce) hardGates.push("stale_nonce");
	if (evidence.readOnly !== binding.readOnly) hardGates.push("read_only_binding_mismatch");
	if (evidence.provider !== binding.provider) hardGates.push("provider_mismatch");
	if (evidence.model !== binding.model) hardGates.push("model_mismatch");
	if (evidence.thinking !== binding.thinking) hardGates.push("thinking_mismatch");
	if (binding.readOnly && evidence.observedMutation !== false) hardGates.push("read_only_violation");
	if (typeof evidence.summary !== "string" || evidence.summary.trim() === "") hardGates.push("artifact_invalid");

	let score = 0;
	try {
		score = rateDimensions(evidence.dimensions);
	} catch {
		if (!hardGates.includes("artifact_invalid")) hardGates.push("artifact_invalid");
	}

	if (hardGates.length > 0) return { decision: "halt", score, hardGates };
	return {
		decision: score >= SHEPHERD_PROCEED_THRESHOLD ? "proceed" : "correct",
		score,
		hardGates,
	};
}

export function selectReadyLanes<T extends LaneDefinition>(
	lanes: readonly T[],
	statuses: ReadonlyMap<string, string>,
	maxConcurrency: number,
): T[] {
	if (!Number.isSafeInteger(maxConcurrency) || maxConcurrency < 1) {
		throw new RangeError("maxConcurrency must be a positive integer");
	}

	const laneById = new Map<string, T>();
	for (const lane of lanes) {
		if (!lane.id || laneById.has(lane.id)) throw new Error(`invalid or duplicate lane id ${JSON.stringify(lane.id)}`);
		laneById.set(lane.id, lane);
	}

	const running = lanes.filter((lane) => statuses.get(lane.id) === "running");
	let available = Math.max(0, maxConcurrency - running.length);
	if (available === 0) return [];
	let mutatorSelected = running.some((lane) => lane.mutating);
	const ready: T[] = [];

	for (const lane of lanes) {
		if (available === 0) break;
		if ((statuses.get(lane.id) ?? "pending") !== "pending") continue;
		if (!lane.dependsOn.every((dependency) => laneById.has(dependency) && statuses.get(dependency) === "succeeded")) {
			continue;
		}
		if (lane.mutating && mutatorSelected) continue;

		ready.push(lane);
		available -= 1;
		if (lane.mutating) mutatorSelected = true;
	}
	return ready;
}

export function reconcileInterruptedRun<T extends ShepherdRunState>(run: T, now: string): T {
	if (typeof now !== "string" || now.length === 0) throw new TypeError("reconciliation timestamp is required");
	const lanes = run.lanes.map((lane) => ({
		...lane,
		status: lane.status === "running" ? "interrupted" as const : lane.status,
	}));
	return {
		...run,
		status: run.status === "running" ? "interrupted" : run.status,
		updatedAt: now,
		lanes,
	};
}
