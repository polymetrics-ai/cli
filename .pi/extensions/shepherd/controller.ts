import {
	assessLaneEvidence,
	reconcileInterruptedRun,
	type LaneBinding,
	type LaneEvidence,
	type ShepherdRunState,
} from "./domain.ts";
import { sanitizeSummary } from "./state-store.ts";
import type { AgentRunRequest, AgentRunner } from "./runner.ts";

export interface ShepherdCommandConfig {
	action: "start" | "resume" | "canary";
	issue: number;
	pr?: number;
	readOnly: true;
	backend: "sdk-inproc";
	experimental: true;
	maxConcurrency: number;
	timeoutMs: number;
}

export interface TargetEvidence {
	cwd: string;
	branch: string;
	candidateHead: string;
	clean: boolean;
	pr?: number;
	prUrl?: string;
	baseBranch?: string;
	draft?: boolean;
	prState?: string;
	mergeStateStatus?: string;
	reviewDecision?: string;
	statusChecks?: Array<{ name: string; status: string; conclusion?: string }>;
}

export interface StateStore {
	load(issue: number): Promise<ShepherdRunState | undefined>;
	save(state: ShepherdRunState): Promise<void>;
	acquireLease(claim: {
		issue: number;
		runId: string;
		mode: "start" | "resume";
	}): Promise<{ assertOwned(): Promise<void>; release(): Promise<void> }>;
}

export interface TargetEvidenceSource {
	capture(command: ShepherdCommandConfig): Promise<TargetEvidence>;
}

interface ControllerDependencies {
	store: StateStore;
	runner: AgentRunner;
	targetEvidence: TargetEvidenceSource;
	clock?: () => string;
	createRunId?: () => string;
	createNonce?: () => string;
}

interface LaneDefinition {
	id: "scout" | "validator";
	role: "scout" | "validator";
	systemPrompt: string;
}

interface LaneOutcome {
	laneId: string;
	evidence?: LaneEvidence;
	decision: "proceed" | "correct" | "halt";
	score: number;
	hardGates: string[];
	summary: string;
}

interface ActiveRunLifecycle {
	runId: string;
	phase: "active" | "terminal";
	cancelReason?: "stopped" | "interrupted";
	abortController: AbortController;
	done: Promise<void>;
	resolveDone(): void;
	lease?: { assertOwned(): Promise<void>; release(): Promise<void> };
}

const PROVIDER = "openai-codex";
const MODEL = "gpt-5.6-sol";
const READ_ONLY_LANES: LaneDefinition[] = [
	{
		id: "scout",
		role: "scout",
		systemPrompt: [
			"You are the read-only reconnaissance lane for the Polymetrics Pi Shepherd.",
			"Assess only the host-verified target snapshot supplied in the prompt.",
			"You have no tools. Do not infer facts that are absent from the snapshot.",
			"Return exactly one compact JSON object matching the requested schema.",
		].join("\n"),
	},
	{
		id: "validator",
		role: "validator",
		systemPrompt: [
			"You are an independent read-only validator for the Polymetrics Pi Shepherd.",
			"Independently assess only the host-verified target snapshot supplied in the prompt.",
			"You have no tools. Do not infer facts that are absent from the snapshot or trust another lane.",
			"Return exactly one compact JSON object matching the requested schema.",
		].join("\n"),
	},
];

function defaultId(prefix: string): string {
	return `${prefix}-${crypto.randomUUID()}`;
}

function buildPrompt(
	command: ShepherdCommandConfig,
	target: TargetEvidence,
	binding: LaneBinding,
	role: LaneDefinition["role"],
): string {
	const objective =
		role === "scout"
			? "Summarize the verified PR state, visible gate results, and concrete blockers at this exact head."
			: "Independently assess snapshot consistency, visible gate compliance, conflicts, and the next safe action.";
	const { cwd: _cwd, ...verifiedTarget } = target;
	return [
		`Target issue: #${command.issue}`,
		command.pr ? `Target pull request: #${command.pr}` : "Target pull request: none",
		`Branch: ${target.branch}`,
		`Exact candidate head: ${target.candidateHead}`,
		`Host-verified target snapshot: ${JSON.stringify(verifiedTarget)}`,
		`Objective: ${objective}`,
		"",
		"Return JSON only with these fields:",
		JSON.stringify({
			...binding,
			summary: "bounded evidence summary, maximum 2000 characters",
			dimensions: {
				correctStage: "number from 0 to 1",
				artifactValid: "number from 0 to 1",
				gatesRespected: "number from 0 to 1",
				realProgress: "number from 0 to 1",
				noHallucination: "number from 0 to 1",
				noConflict: "number from 0 to 1",
			},
			observedMutation: false,
		}),
		"",
		"Echo every binding field exactly. Scores are diagnostic only; deterministic code verifies bindings.",
	].join("\n");
}

function overallScore(outcomes: LaneOutcome[]): number {
	if (outcomes.length === 0) return 0;
	const product = outcomes.reduce((value, outcome) => value * outcome.score, 1);
	return Math.pow(product, 1 / outcomes.length);
}

async function mapWithLimit<T, R>(
	items: T[],
	limit: number,
	worker: (item: T) => Promise<R>,
	onFailure: (error: unknown) => Promise<void>,
): Promise<R[]> {
	const results = new Array<R>(items.length);
	let next = 0;
	let stopped = false;
	const failures: unknown[] = [];
	let cancellation: Promise<void> | undefined;
	async function consume(): Promise<void> {
		while (!stopped) {
			const index = next;
			next += 1;
			if (index >= items.length) return;
			try {
				results[index] = await worker(items[index]);
			} catch (error) {
				failures.push(error);
				if (!stopped) {
					stopped = true;
					cancellation = onFailure(error);
				}
				return;
			}
		}
	}
	await Promise.allSettled(Array.from({ length: Math.min(limit, items.length) }, () => consume()));
	if (cancellation) {
		try {
			await cancellation;
		} catch (error) {
			failures.push(error);
		}
	}
	if (failures.length === 1) throw failures[0];
	if (failures.length > 1) throw new AggregateError(failures, "Shepherd lane task group failed");
	return results;
}

function laneBinding(run: ShepherdRunState, lane: LaneDefinition): LaneBinding {
	return {
		runId: run.runId,
		generation: run.generation,
		laneId: lane.id,
		candidateHead: run.candidateHead,
		validationNonce: run.validationNonce,
		readOnly: true,
		provider: PROVIDER,
		model: MODEL,
		thinking: "xhigh",
	};
}

export class ShepherdController {
	private readonly store: StateStore;
	private readonly runner: AgentRunner;
	private readonly targetEvidence: TargetEvidenceSource;
	private readonly clock: () => string;
	private readonly createRunId: () => string;
	private readonly createNonce: () => string;
	private persistChain: Promise<void> = Promise.resolve();
	private readonly activeRuns = new Map<number, ActiveRunLifecycle>();

	constructor(dependencies: ControllerDependencies) {
		this.store = dependencies.store;
		this.runner = dependencies.runner;
		this.targetEvidence = dependencies.targetEvidence;
		this.clock = dependencies.clock ?? (() => new Date().toISOString());
		this.createRunId = dependencies.createRunId ?? (() => defaultId("run"));
		this.createNonce = dependencies.createNonce ?? (() => defaultId("nonce"));
	}

	async status(issue: number): Promise<ShepherdRunState | undefined> {
		await this.persistChain;
		return this.store.load(issue);
	}

	async start(command: ShepherdCommandConfig): Promise<ShepherdRunState> {
		const lifecycle = this.reserve(command.issue);
		try {
			lifecycle.lease = await this.store.acquireLease({
				issue: command.issue,
				runId: lifecycle.runId,
				mode: "start",
			});
			const existing = await this.store.load(command.issue);
			if (existing?.status === "running") {
				throw new Error(`Shepherd run ${existing.runId} is already active for issue #${command.issue}`);
			}
			const generation = existing ? existing.generation + 1 : 1;
			return await this.execute(command, generation, lifecycle);
		} finally {
			lifecycle.phase = "terminal";
			try {
				await lifecycle.lease?.release();
			} finally {
				this.release(command.issue, lifecycle);
				lifecycle.resolveDone();
			}
		}
	}

	async resume(command: ShepherdCommandConfig): Promise<ShepherdRunState> {
		const lifecycle = this.reserve(command.issue);
		try {
			lifecycle.lease = await this.store.acquireLease({
				issue: command.issue,
				runId: lifecycle.runId,
				mode: "resume",
			});
			let existing = await this.store.load(command.issue);
			if (!existing) {
				throw new Error(`No persisted Shepherd run exists for issue #${command.issue}`);
			}
			if (command.pr !== undefined && command.pr !== existing.pr) {
				const persistedTarget = existing.pr === undefined ? "no PR" : `PR #${existing.pr}`;
				throw new Error(
					`Cannot resume issue #${command.issue}: persisted ${persistedTarget} differs from requested PR #${command.pr}`,
				);
			}
			const { pr: _requestedPr, ...commandWithoutPr } = command;
			const effectiveCommand: ShepherdCommandConfig = existing.pr === undefined
				? commandWithoutPr
				: { ...commandWithoutPr, pr: existing.pr };
			if (existing.status === "running") {
				existing = reconcileInterruptedRun(existing, this.clock());
				await this.persist(existing, lifecycle);
			}
			if (existing.status === "completed") {
				throw new Error(`Shepherd run for issue #${command.issue} is already completed; use start for a fresh run`);
			}
			return await this.execute(effectiveCommand, existing.generation + 1, lifecycle);
		} finally {
			lifecycle.phase = "terminal";
			try {
				await lifecycle.lease?.release();
			} finally {
				this.release(command.issue, lifecycle);
				lifecycle.resolveDone();
			}
		}
	}

	async stop(issue: number): Promise<ShepherdRunState> {
		const active = this.activeRuns.get(issue);
		if (!active || active.phase !== "active") {
			throw new Error(`Shepherd run for issue #${issue} is not owned by this Pi session`);
		}
		active.cancelReason = "stopped";
		active.abortController.abort("stopped");
		await this.runner.abort(active.runId).catch(() => undefined);
		await active.done;
		const state = await this.store.load(issue);
		if (!state || state.runId !== active.runId || state.status !== "stopped") {
			throw new Error(`Owned Shepherd run for issue #${issue} did not persist a stopped state`);
		}
		return state;
	}

	async shutdown(): Promise<void> {
		const owned = [...this.activeRuns.values()];
		const active = owned.filter((lifecycle) => lifecycle.phase === "active");
		for (const lifecycle of active) {
			lifecycle.cancelReason ??= "interrupted";
			lifecycle.abortController.abort("interrupted");
		}
		const cleanup = await Promise.allSettled([
			...active.map((lifecycle) => this.runner.abort(lifecycle.runId)),
			this.runner.close(),
		]);
		await Promise.all(owned.map((lifecycle) => lifecycle.done));
		const failures = cleanup
			.filter((result): result is PromiseRejectedResult => result.status === "rejected")
			.map((result) => result.reason);
		if (failures.length > 0) throw new AggregateError(failures, "Shepherd controller shutdown failed");
	}

	private async execute(
		command: ShepherdCommandConfig,
		generation: number,
		lifecycle: ActiveRunLifecycle,
	): Promise<ShepherdRunState> {
		const target = await this.targetEvidence.capture(command);
		if (!target.clean) throw new Error("Target worktree must be clean");
		const now = this.clock();
		const run: ShepherdRunState = {
			schemaVersion: 1,
			issue: command.issue,
			...(command.pr ? { pr: command.pr } : {}),
			runId: lifecycle.runId,
			generation,
			status: "running",
			candidateHead: target.candidateHead,
			validationNonce: this.createNonce(),
			createdAt: now,
			updatedAt: now,
			lanes: READ_ONLY_LANES.map((lane) => ({
				id: lane.id,
				role: lane.role,
				mutating: false,
				dependsOn: [],
				status: "pending",
			})),
		};
		await this.persist(run, lifecycle);
		if (lifecycle.cancelReason) return this.persistCancelled(run, lifecycle);

		const outcomes = await mapWithLimit(
			READ_ONLY_LANES,
			command.maxConcurrency,
			async (lane) => this.executeLane(run, command, target, lane, lifecycle),
			async (error) => {
				lifecycle.abortController.abort(error);
				await this.runner.abort(lifecycle.runId);
			},
		);
		if (lifecycle.cancelReason) return this.persistCancelled(run, lifecycle);

		for (const outcome of outcomes) {
			const lane = run.lanes.find((candidate) => candidate.id === outcome.laneId);
			if (!lane) continue;
			lane.status =
				outcome.decision === "proceed"
					? "succeeded"
					: outcome.decision === "correct"
						? "failed"
						: "halted";
			lane.summary = sanitizeSummary(outcome.summary, 2000);
			lane.score = outcome.score;
			lane.hardGates = [...outcome.hardGates];
		}

		run.score = overallScore(outcomes);
		run.hardGates = [...new Set(outcomes.flatMap((outcome) => outcome.hardGates))].sort();
		try {
			const finalTarget = await this.targetEvidence.capture(command);
			if (!sameTarget(target, finalTarget)) run.hardGates.push("target_changed");
		} catch {
			run.hardGates.push("target_revalidation_failed");
		}
		run.hardGates = [...new Set(run.hardGates)].sort();
		if (lifecycle.cancelReason) return this.persistCancelled(run, lifecycle);
		if (run.hardGates.length > 0 || outcomes.some((outcome) => outcome.decision === "halt")) {
			run.status = "halted";
		} else if (outcomes.some((outcome) => outcome.decision === "correct")) {
			run.status = "failed";
		} else {
			run.status = "completed";
		}
		run.updatedAt = this.clock();
		await this.persist(run, lifecycle);
		if (lifecycle.cancelReason) return this.persistCancelled(run, lifecycle);
		lifecycle.phase = "terminal";
		return run;
	}

	private async executeLane(
		run: ShepherdRunState,
		command: ShepherdCommandConfig,
		target: TargetEvidence,
		lane: LaneDefinition,
		lifecycle: ActiveRunLifecycle,
	): Promise<LaneOutcome> {
		const laneState = run.lanes.find((candidate) => candidate.id === lane.id);
		if (laneState) laneState.status = "running";
		run.updatedAt = this.clock();
		await this.persist(run, lifecycle);
		if (lifecycle.cancelReason || lifecycle.abortController.signal.aborted) {
			return {
				laneId: lane.id,
				decision: "halt",
				score: 0,
				hardGates: ["lane_cancelled"],
				summary: "Lane cancelled before AgentSession dispatch",
			};
		}
		const binding = laneBinding(run, lane);
		const request: AgentRunRequest = {
			runId: run.runId,
			laneId: lane.id,
			role: lane.role,
			cwd: target.cwd,
			readOnly: true,
			provider: PROVIDER,
			model: MODEL,
			thinking: "xhigh",
			tools: [],
			systemPrompt: lane.systemPrompt,
			prompt: buildPrompt(command, target, binding, lane.role),
			timeoutMs: command.timeoutMs,
			signal: lifecycle.abortController.signal,
			binding,
		};
		try {
			const evidence = await this.runner.run(request);
			const assessment = assessLaneEvidence(binding, evidence);
			return {
				laneId: lane.id,
				evidence,
				...assessment,
				summary: evidence.summary,
			};
		} catch (error) {
			const summary = error instanceof Error ? error.message : String(error);
			return {
				laneId: lane.id,
				decision: "halt",
				score: 0,
				hardGates: ["lane_execution_failed"],
				summary,
			};
		}
	}

	private persist(state: ShepherdRunState, lifecycle?: ActiveRunLifecycle): Promise<void> {
		const snapshot = structuredClone(state);
		const task = this.persistChain.then(async () => {
			await lifecycle?.lease?.assertOwned();
			await this.store.save(snapshot);
		});
		this.persistChain = task.catch(() => undefined);
		return task;
	}

	private reserve(issue: number): ActiveRunLifecycle {
		if (this.activeRuns.has(issue)) {
			throw new Error(`A Shepherd run is already active for issue #${issue}`);
		}
		let resolveDone: () => void = () => {};
		const done = new Promise<void>((resolve) => { resolveDone = resolve; });
		const lifecycle: ActiveRunLifecycle = {
			runId: this.createRunId(),
			phase: "active",
			abortController: new AbortController(),
			done,
			resolveDone,
		};
		this.activeRuns.set(issue, lifecycle);
		return lifecycle;
	}

	private release(issue: number, lifecycle: ActiveRunLifecycle): void {
		if (this.activeRuns.get(issue) === lifecycle) this.activeRuns.delete(issue);
	}

	private async persistCancelled(
		run: ShepherdRunState,
		lifecycle: ActiveRunLifecycle,
	): Promise<ShepherdRunState> {
		const status = lifecycle.cancelReason === "interrupted" ? "interrupted" : "stopped";
		run.status = status;
		run.updatedAt = this.clock();
		run.lanes = run.lanes.map((lane) =>
			lane.status === "running" || lane.status === "pending"
				? { ...lane, status }
				: lane,
		);
		await this.persist(run, lifecycle);
		lifecycle.phase = "terminal";
		return run;
	}
}

function sameTarget(initial: TargetEvidence, final: TargetEvidence): boolean {
	return initial.clean === final.clean
		&& initial.cwd === final.cwd
		&& initial.branch === final.branch
		&& initial.candidateHead === final.candidateHead
		&& initial.pr === final.pr
		&& initial.prUrl === final.prUrl
		&& initial.baseBranch === final.baseBranch
		&& initial.draft === final.draft
		&& initial.prState === final.prState
		&& initial.mergeStateStatus === final.mergeStateStatus
		&& initial.reviewDecision === final.reviewDecision
		&& JSON.stringify(initial.statusChecks ?? []) === JSON.stringify(final.statusChecks ?? []);
}
