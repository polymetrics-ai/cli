import type { ShepherdResumeCommand, ShepherdStartCommand } from "./arguments.ts";
import {
	ProductionLifecycleError,
	type ProductionChildSpec,
	type ProductionParentPlanDocument,
	type ProductionStageCheckpoint,
} from "./autonomous-production-contract.ts";
import {
	advanceProductionGeneration,
	authorizeProductionChildRetry,
	assertProductionPlanBinding,
	evolveProductionState,
	refreshProductionChildOwnership,
	reconcileProductionChildHead,
	waitForProductionChildIntervention,
	type ProductionAutonomousState,
	type ProductionChildRuntimeState,
	type ProductionStateFence,
	type ProductionStateStore,
} from "./autonomous-production-state.ts";
import type { ProductionRecoveryFence, ProductionRecoveryResult } from "./autonomous-recovery.ts";
import type { ProductionPlanSnapshot } from "./production-intake.ts";
import { selectProductionChildren } from "./production-scheduler.ts";

export interface ProductionPlanIntakePort {
	load(issue: number, signal: AbortSignal): Promise<ProductionPlanSnapshot>;
}

export interface ProductionRecoveryBarrierPort {
	open(fence: ProductionRecoveryFence): Promise<ProductionRecoveryResult>;
}

export interface ProductionChildPipelineContext {
	plan: ProductionParentPlanDocument;
	state: ProductionAutonomousState;
	child: ProductionChildSpec;
	runtime: ProductionChildRuntimeState;
	runId: string;
	resourceGeneration: number;
	generation: number;
	timeoutMs: number;
	signal: AbortSignal;
}

export interface ProductionChildInterventionObservation {
	status: "pending" | "authorized" | "aborted";
	effectKey?: string;
}

export interface ProductionChildPipelinePort {
	workspace(context: ProductionChildPipelineContext): Promise<ProductionStageCheckpoint>;
	implement(context: ProductionChildPipelineContext): Promise<ProductionStageCheckpoint>;
	verify(context: ProductionChildPipelineContext): Promise<ProductionStageCheckpoint>;
	publish(context: ProductionChildPipelineContext): Promise<ProductionStageCheckpoint>;
	review(context: ProductionChildPipelineContext): Promise<ProductionStageCheckpoint>;
	correct(context: ProductionChildPipelineContext, findings: readonly string[]): Promise<ProductionStageCheckpoint>;
	reconcileChildHead(context: ProductionChildPipelineContext): Promise<{
		checkpoint: ProductionStageCheckpoint;
		previousHead: string;
		head: string;
		invalidated: { verification: true; review: true; integration: true };
	}>;
	refresh(context: ProductionChildPipelineContext): Promise<ProductionStageCheckpoint>;
	integrate(context: ProductionChildPipelineContext): Promise<ProductionStageCheckpoint>;
	requestIntervention(
		context: ProductionChildPipelineContext,
		reason: "retry_budget_exhausted" | "correction_budget_exhausted",
	): Promise<{ requestId: string; pullRequest?: number; head?: string; effectKey?: string }>;
	observeIntervention(state: ProductionAutonomousState, signal: AbortSignal): Promise<ProductionChildInterventionObservation>;
	/** Marks an observed effect applied only after the supplied controller state is durably CAS-persisted. */
	acknowledge(effectKey: string, state: ProductionAutonomousState): Promise<void>;
	abort(runId: string): Promise<void>;
	join(runId: string): Promise<void>;
	close(): Promise<void>;
}

export interface ProductionParentFinalization {
	pullRequest: number;
	head: string;
	summary: string;
}

export interface ProductionParentFinalizerPort {
	finalize(
		plan: ProductionParentPlanDocument,
		state: ProductionAutonomousState,
		signal: AbortSignal,
	): Promise<ProductionParentFinalization>;
	close(): Promise<void>;
}

export type ProductionParentGateObservation =
	| { status: "pending" | "approved_waiting_for_merge" | "rejected" }
	| {
		status: "invalidated";
		repository: string;
		pullRequest: number;
		previousHead: string;
		currentHead: string;
		revision: number;
		observedAt: string;
	}
	| {
		status: "merged";
		repository: string;
		pullRequest: number;
		head: string;
		mergedAt: string;
		mergeCommitSha: string;
		revision: number;
		observedAt: string;
	};

export interface ProductionParentGatePort {
	/** Purely prepares the deterministic external request identity for durable pre-effect persistence. */
	prepare(
		plan: ProductionParentPlanDocument,
		state: ProductionAutonomousState,
		finalization: ProductionParentFinalization,
	): { requestId: string };
	/** Idempotently applies or reconciles the exact prepared request. */
	request(
		plan: ProductionParentPlanDocument,
		state: ProductionAutonomousState,
		finalization: ProductionParentFinalization,
		signal: AbortSignal,
	): Promise<{ requestId: string }>;
	/** Replay-safe durable decision and authoritative merge reconciliation. */
	observe(
		plan: ProductionParentPlanDocument,
		state: ProductionAutonomousState,
		signal: AbortSignal,
	): Promise<ProductionParentGateObservation>;
	close(): Promise<void>;
}

export interface ProductionShepherdControllerOptions {
	stateStore: ProductionStateStore;
	intake: ProductionPlanIntakePort;
	recovery: ProductionRecoveryBarrierPort;
	pipeline: ProductionChildPipelinePort;
	finalizer: ProductionParentFinalizerPort;
	parentGate: ProductionParentGatePort;
	now?: () => Date;
	newRunId?: () => string;
}

interface ActiveRun {
	issue: number;
	runId: string;
	generation: number;
	controller: AbortController;
	promise: Promise<ProductionAutonomousState>;
}

interface ChildSettlement {
	id: string;
	kind: "succeeded" | "waiting" | "failed";
	error?: unknown;
}

type ChildStage = "workspace" | "implementation" | "verification" | "publication" | "review" | "correction" | "integration";

class StaleProductionGenerationError extends Error {
	constructor() {
		super("stale production generation result was fenced");
		this.name = "StaleProductionGenerationError";
	}
}

function fence(state: ProductionAutonomousState): ProductionStateFence {
	return { issue: state.parentIssue, revision: state.revision, generation: state.generation, runId: state.runId };
}

function safeFailureSummary(kind: ProductionLifecycleError["kind"], stage: ChildStage): string {
	return `${kind} production failure during ${stage}`;
}

function mergeCheckpoint(
	previous: ProductionStageCheckpoint | undefined,
	next: ProductionStageCheckpoint,
): ProductionStageCheckpoint {
	const effectKeys = [...new Set([
		...(previous?.effectKey === undefined ? [] : [previous.effectKey]),
		...(previous?.effectKeys ?? []),
		...(next.effectKey === undefined ? [] : [next.effectKey]),
		...(next.effectKeys ?? []),
	])];
	return {
		...(previous ?? { summary: next.summary }),
		...next,
		...(effectKeys.length === 0 ? {} : { effectKeys }),
		...(next.workspace === undefined && previous?.workspace !== undefined ? { workspace: previous.workspace } : {}),
		...(next.verification === undefined && previous?.verification !== undefined ? { verification: previous.verification } : {}),
		...(next.pullRequest === undefined && previous?.pullRequest !== undefined ? { pullRequest: previous.pullRequest } : {}),
		...(next.review === undefined && previous?.review !== undefined ? { review: previous.review } : {}),
		...(next.integrationReceiptDigest === undefined && previous?.integrationReceiptDigest !== undefined
			? { integrationReceiptDigest: previous.integrationReceiptDigest } : {}),
		...(next.parentHead === undefined && previous?.parentHead !== undefined ? { parentHead: previous.parentHead } : {}),
	};
}

function lifecycleError(value: unknown): ProductionLifecycleError {
	if (value instanceof ProductionLifecycleError) return value;
	return new ProductionLifecycleError("terminal", "production child adapter failed closed", ["adapter_failed"]);
}

function correctionFindings(runtime: ProductionChildRuntimeState): string[] {
	const review = runtime.checkpoint?.review?.findings.map((finding) => finding.summary) ?? [];
	if (review.length > 0) return review;
	return runtime.checkpoint?.verification?.commands
		.filter((command) => command.status === "failed")
		.map((command) => `Verification ${command.id} failed (${command.failureKind ?? "unknown"}).`) ?? [];
}

export class ProductionShepherdController {
	readonly #store: ProductionStateStore;
	readonly #intake: ProductionPlanIntakePort;
	readonly #recovery: ProductionRecoveryBarrierPort;
	readonly #pipeline: ProductionChildPipelinePort;
	readonly #finalizer: ProductionParentFinalizerPort;
	readonly #parentGate: ProductionParentGatePort;
	readonly #now: () => Date;
	readonly #newRunId: () => string;
	#active: ActiveRun | undefined;
	#current: ProductionAutonomousState | undefined;
	#mutationTail: Promise<void> = Promise.resolve();
	#closed = false;
	#closePromise: Promise<void> | undefined;

	constructor(options: ProductionShepherdControllerOptions) {
		this.#store = options.stateStore;
		this.#intake = options.intake;
		this.#recovery = options.recovery;
		this.#pipeline = options.pipeline;
		this.#finalizer = options.finalizer;
		this.#parentGate = options.parentGate;
		this.#now = options.now ?? (() => new Date());
		this.#newRunId = options.newRunId ?? (() => crypto.randomUUID());
	}

	async status(issue: number): Promise<ProductionAutonomousState | undefined> {
		return this.#store.load(issue);
	}

	start(command: ShepherdStartCommand): Promise<ProductionAutonomousState> {
		return this.#launch(command.issue, (controller, runId) => this.#start(command, controller.signal, runId));
	}

	resume(command: ShepherdResumeCommand): Promise<ProductionAutonomousState> {
		return this.#launch(command.issue, (controller, runId) => this.#resume(command, controller.signal, runId));
	}

	async stop(issue: number): Promise<ProductionAutonomousState> {
		const active = this.#active;
		if (active?.issue === issue) {
			active.controller.abort(new Error("production Shepherd stop requested"));
			let abortFailure: unknown;
			try { await this.#pipeline.abort(active.runId); } catch (error) { abortFailure = error; }
			let runState: ProductionAutonomousState | undefined;
			let runFailure: unknown;
			try { runState = await active.promise; } catch (error) { runFailure = error; }
			let joinFailure: unknown;
			try { await this.#pipeline.join(active.runId); } catch (error) { joinFailure = error; }
			const durable = await this.#store.load(issue);
			if (durable !== undefined && durable.status !== "completed" && durable.status !== "failed") {
				this.#current = durable;
				const stopped = durable.status === "stopped" ? durable : await this.#evolve(durable.runId, durable.generation, (draft) => {
					draft.status = "stopped";
					draft.stage = draft.humanGate || draft.childGate ? "human_decision" : "schedule";
					for (const child of draft.children) {
						if (child.status === "running") {
							child.resumeStage = child.stage;
							child.status = "cancelled";
							child.stage = "cancelled";
						}
					}
				});
				if (abortFailure !== undefined) throw abortFailure;
				if (joinFailure !== undefined) throw joinFailure;
				return stopped;
			}
			if (abortFailure !== undefined) throw abortFailure;
			if (joinFailure !== undefined) throw joinFailure;
			if (runState !== undefined) return runState;
			throw new Error("production Shepherd stopped before durable initialization completed", { cause: runFailure });
		}
		if (active) throw new Error(`a production Shepherd run is active for issue #${active.issue}`);
		const state = await this.#store.load(issue);
		if (!state) throw new Error(`no production Shepherd run exists for issue #${issue}`);
		if (state.status === "completed" || state.status === "failed" || state.status === "waiting_human") return state;
		this.#current = state;
		return this.#evolve(state.runId, state.generation, (draft) => {
			draft.status = "stopped";
			draft.stage = "schedule";
			for (const child of draft.children) {
				if (child.status === "running") {
					child.resumeStage = child.stage;
					child.status = "cancelled";
					child.stage = "cancelled";
				}
			}
		});
	}

	shutdown(): Promise<void> {
		if (!this.#closePromise) {
			this.#closed = true;
			this.#closePromise = (async () => {
				const active = this.#active;
				if (active) await this.stop(active.issue);
				await Promise.all([this.#pipeline.close(), this.#finalizer.close(), this.#parentGate.close()]);
			})();
		}
		return this.#closePromise;
	}

	#launch(
		issue: number,
		run: (controller: AbortController, runId: string) => Promise<ProductionAutonomousState>,
	): Promise<ProductionAutonomousState> {
		if (this.#closed) return Promise.reject(new Error("production Shepherd controller is closed"));
		if (this.#active) return Promise.reject(new Error(`a production Shepherd run is already active for issue #${this.#active.issue}`));
		const controller = new AbortController();
		const runId = this.#newRunId();
		const promise = run(controller, runId);
		const active: ActiveRun = { issue, runId, generation: 0, controller, promise };
		this.#active = active;
		void promise.then(
			(state) => { active.generation = state.generation; if (this.#active === active) this.#active = undefined; },
			() => { if (this.#active === active) this.#active = undefined; },
		);
		return promise;
	}

	async #start(command: ShepherdStartCommand, signal: AbortSignal, runId: string): Promise<ProductionAutonomousState> {
		if (await this.#store.load(command.issue)) throw new Error(`production Shepherd state already exists for issue #${command.issue}; use resume`);
		const snapshot = await this.#intake.load(command.issue, signal);
		if (signal.aborted) throw signal.reason ?? new Error("production intake cancelled");
		const { createProductionAutonomousState } = await import("./autonomous-production-state.ts");
		this.#current = await this.#store.create(createProductionAutonomousState(snapshot.plan, {
			runId,
			now: this.#now(),
			maxConcurrency: command.maxConcurrency,
			timeoutMs: command.timeoutMs,
		}));
		await this.#recovery.open({ runId, generation: 1, signal });
		this.#current = await this.#reloadRecovered(command.issue, runId, 1, snapshot.plan);
		await this.#evolve(runId, 1, (draft) => { draft.stage = "schedule"; });
		return this.#drive(snapshot.plan, runId, 1, signal);
	}

	async #resume(command: ShepherdResumeCommand, signal: AbortSignal, runId: string): Promise<ProductionAutonomousState> {
		let state = await this.#store.load(command.issue);
		if (!state) throw new Error(`no production Shepherd run exists for issue #${command.issue}`);
		const snapshot = await this.#intake.load(command.issue, signal);
		assertProductionPlanBinding(state, snapshot.plan);
		this.#current = state;
		if (state.humanGate?.status === "prepared") {
			this.#current = state;
			return this.#ensurePreparedParentGate(snapshot.plan, state, signal);
		}
		if (state.humanGate) {
			const observed = await this.#observeParentGate(snapshot.plan, state, signal);
			if (!observed.invalidated) return observed.state;
			state = observed.state;
		}
		if (state.childGate) {
			const observation = await this.#pipeline.observeIntervention(state, signal);
			if (observation.status === "pending") return state;
			if (observation.status === "aborted") {
				const persisted = await this.#evolve(state.runId, state.generation, (draft) => {
					draft.childGate!.status = "aborted";
					draft.status = "failed";
					draft.stage = "blocked";
					draft.terminalBlocker = "human aborted the exhausted child";
					const child = draft.children.find((candidate) => candidate.id === draft.childGate!.childId)!;
					child.status = "failed";
					child.stage = "failed";
				});
				await this.#acknowledge(observation.effectKey, persisted);
				return persisted;
			}
			const childId = state.childGate.childId;
			const requestId = state.childGate.requestId;
			state = await this.#custom(state.runId, state.generation, (current, stateFence) =>
				authorizeProductionChildRetry(current, stateFence, {
					childId,
					requestId,
					now: this.#now(),
				}));
			await this.#acknowledge(observation.effectKey, state);
		}
		await this.#recovery.open({ runId: state.runId, generation: state.generation, signal });
		state = await this.#reloadRecovered(command.issue, state.runId, state.generation, snapshot.plan);
		this.#current = state;
		if (command.maxConcurrency !== state.maxConcurrency || command.timeoutMs !== state.timeoutMs) {
			throw new Error("resume concurrency/timeout differs from the durable production run policy");
		}
		const previous = state;
		state = advanceProductionGeneration(previous, fence(previous), runId, this.#now());
		this.#current = await this.#store.compareAndSwap(fence(previous), state);
		await this.#evolve(runId, state.generation, (draft) => { draft.stage = "schedule"; });
		return this.#drive(snapshot.plan, runId, state.generation, signal);
	}

	async #observeParentGate(
		plan: ProductionParentPlanDocument,
		state: ProductionAutonomousState,
		signal: AbortSignal,
	): Promise<{ state: ProductionAutonomousState; invalidated: boolean }> {
		const observation = await this.#parentGate.observe(plan, state, signal);
		if (observation.status === "pending" || observation.status === "approved_waiting_for_merge") {
			return { state, invalidated: false };
		}
		if (observation.status === "invalidated") {
			if (observation.repository !== state.humanGate?.repository
				|| observation.pullRequest !== state.humanGate.pullRequest
				|| observation.previousHead !== state.humanGate.head) {
				throw new Error("authoritative parent invalidation moved from the durable exact-head gate");
			}
			const invalidated = await this.#evolve(state.runId, state.generation, (draft) => {
				if (!draft.humanGate) throw new Error("parent gate disappeared during invalidation");
				draft.invalidatedParentGates = [
					...(draft.invalidatedParentGates ?? []),
					{
						...draft.humanGate,
						status: "invalidated",
						invalidationEvidence: {
							currentHead: observation.currentHead,
							revision: observation.revision,
							observedAt: observation.observedAt,
						},
					},
				];
				delete draft.humanGate;
				draft.status = "running";
				draft.stage = "schedule";
			});
			return { state: invalidated, invalidated: true };
		}
		const settled = await this.#evolve(state.runId, state.generation, (draft) => {
			if (!draft.humanGate) throw new Error("parent gate disappeared during observation");
			if (observation.status === "merged") {
				if (observation.repository !== draft.humanGate.repository
					|| observation.pullRequest !== draft.humanGate.pullRequest
					|| observation.head !== draft.humanGate.head) {
					throw new Error("authoritative parent merge observation moved from the durable exact-head gate");
				}
				draft.humanGate.status = "merged";
				draft.humanGate.mergeEvidence = {
					mergedAt: observation.mergedAt,
					mergeCommitSha: observation.mergeCommitSha,
					revision: observation.revision,
					observedAt: observation.observedAt,
				};
				draft.status = "completed";
				draft.stage = "completed";
			} else {
				draft.humanGate.status = "rejected";
				draft.status = "failed";
				draft.stage = "blocked";
				draft.terminalBlocker = "human rejected the exact parent merge";
			}
		});
		return { state: settled, invalidated: false };
	}

	async #ensurePreparedParentGate(
		plan: ProductionParentPlanDocument,
		state: ProductionAutonomousState,
		signal: AbortSignal,
	): Promise<ProductionAutonomousState> {
		const prepared = state.humanGate;
		if (!prepared || prepared.status !== "prepared") {
			throw new Error("durable parent gate request is not prepared");
		}
		const request = await this.#parentGate.request(plan, structuredClone(state), {
			pullRequest: prepared.pullRequest,
			head: prepared.head,
			summary: "Reconcile the durable exact-head parent merge request.",
		}, signal);
		if (signal.aborted) throw signal.reason ?? new Error("parent gate request cancelled");
		if (request.requestId !== prepared.requestId) {
			throw new Error("reconciled parent gate request changed its durable exact identity");
		}
		return this.#evolve(state.runId, state.generation, (draft) => {
			if (!draft.humanGate || draft.humanGate.status !== "prepared"
				|| draft.humanGate.requestId !== request.requestId) {
				throw new Error("durable prepared parent gate disappeared during reconciliation");
			}
			draft.humanGate.status = "pending";
			draft.status = "waiting_human";
			draft.stage = "human_decision";
		});
	}

	async #reloadRecovered(
		issue: number,
		runId: string,
		generation: number,
		plan: ProductionParentPlanDocument,
	): Promise<ProductionAutonomousState> {
		const recovered = await this.#store.load(issue);
		if (!recovered || recovered.runId !== runId || recovered.generation !== generation) {
			throw new StaleProductionGenerationError();
		}
		assertProductionPlanBinding(recovered, plan);
		return recovered;
	}

	#context(
		plan: ProductionParentPlanDocument,
		childId: string,
		runId: string,
		generation: number,
		signal: AbortSignal,
	): ProductionChildPipelineContext {
		const state = this.#current;
		if (!state || state.runId !== runId || state.generation !== generation) throw new StaleProductionGenerationError();
		const child = plan.children.find((candidate) => candidate.id === childId);
		const runtime = state.children.find((candidate) => candidate.id === childId);
		if (!child || !runtime) throw new Error(`unknown production child ${childId}`);
		return {
			plan,
			state: structuredClone(state),
			child: structuredClone(child),
			runtime: structuredClone(runtime),
			runId,
			resourceGeneration: state.resourceGeneration,
			generation,
			timeoutMs: state.timeoutMs,
			signal,
		};
	}

	async #drive(
		plan: ProductionParentPlanDocument,
		runId: string,
		generation: number,
		signal: AbortSignal,
	): Promise<ProductionAutonomousState> {
		const active = new Map<string, Promise<ChildSettlement>>();
		let failure: unknown;
		let waiting = false;
		while (!signal.aborted && failure === undefined && !waiting) {
			const current = this.#current;
			if (!current || current.runId !== runId || current.generation !== generation) throw new StaleProductionGenerationError();
			const decision = selectProductionChildren(current);
			if (decision.kind === "dispatch") {
				await this.#evolve(runId, generation, (draft) => {
					draft.stage = "child_lifecycle";
					delete draft.idleReason;
					for (const id of decision.childIds) {
						const child = draft.children.find((candidate) => candidate.id === id)!;
						child.status = "running";
						if (child.resumeStage !== undefined) {
							child.stage = child.resumeStage;
							delete child.resumeStage;
						} else {
							child.stage = "workspace";
							child.attempts += 1;
						}
					}
				});
				for (const id of decision.childIds) {
					const task = this.#runChild(plan, id, runId, generation, signal).then(
						(kind): ChildSettlement => ({ id, kind }),
						(error): ChildSettlement => ({ id, kind: "failed", error }),
					);
					active.set(id, task);
				}
				continue;
			}
			if (decision.kind === "complete" && active.size === 0) break;
			if (active.size === 0) {
				failure = new ProductionLifecycleError("terminal", `scheduler cannot progress: ${decision.kind === "idle" ? decision.reason : "unknown"}`);
				break;
			}
			if (decision.kind === "idle" && current.idleReason !== decision.reason) {
				await this.#evolve(runId, generation, (draft) => { draft.idleReason = decision.reason; });
			}
			const settled = await Promise.race(active.values());
			active.delete(settled.id);
			if (settled.kind === "waiting") waiting = true;
			if (settled.error !== undefined) failure = settled.error;
		}

		if (signal.aborted || failure !== undefined || waiting) {
			if (!signal.aborted) {
				try { await this.#pipeline.abort(runId); } catch (error) { failure ??= error; }
			}
			const remaining = await Promise.all(active.values());
			failure ??= remaining.find((result) => result.error !== undefined)?.error;
			await this.#pipeline.join(runId);
			if (waiting && failure === undefined) return structuredClone(this.#current!);
			return this.#evolve(runId, generation, (draft) => {
				draft.status = signal.aborted && failure === undefined ? "stopped" : "failed";
				draft.stage = signal.aborted && failure === undefined ? "schedule" : "blocked";
				if (failure !== undefined) draft.terminalBlocker = "production child lifecycle failed closed";
				for (const child of draft.children) {
					if (child.status === "running") {
						child.status = signal.aborted && failure === undefined ? "cancelled" : "blocked";
						if (signal.aborted && failure === undefined) {
							child.resumeStage = child.stage;
							child.stage = "cancelled";
						}
					}
				}
			});
		}

		await this.#pipeline.join(runId);
		const finalization = await this.#finalizer.finalize(plan, structuredClone(this.#current!), signal);
		if (signal.aborted) return this.#evolve(runId, generation, (draft) => { draft.status = "stopped"; draft.stage = "schedule"; });
		const prepared = this.#parentGate.prepare(plan, structuredClone(this.#current!), finalization);
		const preparedState = await this.#evolve(runId, generation, (draft) => {
			draft.status = "running";
			draft.stage = "human_decision";
			draft.humanGate = {
				repository: plan.repository,
				pullRequest: finalization.pullRequest,
				generation,
				head: finalization.head,
				requestId: prepared.requestId,
				status: "prepared",
			};
		});
		return this.#ensurePreparedParentGate(plan, preparedState, signal);
	}

	async #runChild(
		plan: ProductionParentPlanDocument,
		childId: string,
		runId: string,
		generation: number,
		signal: AbortSignal,
	): Promise<"succeeded" | "waiting"> {
		const resumed = this.#context(plan, childId, runId, generation, signal).runtime.stage;
		let stage: ChildStage = resumed === "workspace" || resumed === "implementation" || resumed === "verification"
			|| resumed === "publication" || resumed === "review" || resumed === "correction" || resumed === "integration"
			? resumed : "workspace";
		while (!signal.aborted) {
			try {
				if (stage === "workspace") {
					await this.#stage(plan, childId, runId, generation, signal, "workspace", (context) => this.#pipeline.workspace(context));
					stage = "implementation";
				}
				if (stage === "implementation") {
					await this.#stage(plan, childId, runId, generation, signal, "implementation", (context) => this.#pipeline.implement(context));
					stage = "verification";
				}
				if (stage === "verification") {
					const verified = await this.#stage(
						plan, childId, runId, generation, signal, "verification", (context) => this.#pipeline.verify(context),
					);
					if (verified.verification === undefined) {
						throw new ProductionLifecycleError("terminal", "verification stage lacks durable result evidence", ["verification_evidence_missing"]);
					}
					if (verified.verification.status === "failed") {
						throw new ProductionLifecycleError("correction_required", "bounded verification requires correction", ["verification_failed"]);
					}
					stage = this.#context(plan, childId, runId, generation, signal).runtime.checkpoint?.pullRequest === undefined
						? "publication" : "review";
				}
				if (stage === "publication") {
					await this.#stage(plan, childId, runId, generation, signal, "publication", (context) => this.#pipeline.publish(context));
					stage = "review";
				}
				if (stage === "review") {
					const reviewed = await this.#stage(plan, childId, runId, generation, signal, "review", (context) => this.#pipeline.review(context));
					if (reviewed.review?.status !== "clean") {
						throw new ProductionLifecycleError("correction_required", "independent review requires correction", ["review_findings"]);
					}
					stage = "integration";
				}
				if (stage === "correction") {
					const current = this.#context(plan, childId, runId, generation, signal).runtime;
					const findings = correctionFindings(current);
					const corrected = await this.#pipeline.correct(this.#context(plan, childId, runId, generation, signal), findings);
					if (signal.aborted) return "succeeded";
					const persisted = await this.#evolve(runId, generation, (draft) => {
						const child = draft.children.find((candidate) => candidate.id === childId)!;
						child.stage = "verification";
						child.checkpoint = mergeCheckpoint(child.checkpoint, corrected);
						delete child.checkpoint.verification;
						delete child.checkpoint.review;
						delete child.checkpoint.integrationReceiptDigest;
					});
					await this.#acknowledgeCheckpoint(corrected, persisted);
					stage = "verification";
				}
				if (stage === "integration") {
					const integrated = await this.#stage(plan, childId, runId, generation, signal, "integration", (context) => this.#pipeline.integrate(context));
					if (!integrated.integrationReceiptDigest) throw new ProductionLifecycleError("terminal", "integration lacks a durable receipt");
					await this.#evolve(runId, generation, (draft) => {
						const child = draft.children.find((candidate) => candidate.id === childId)!;
						child.status = "succeeded";
						child.stage = "succeeded";
					});
					return "succeeded";
				}
			} catch (error) {
				if (signal.aborted || error instanceof StaleProductionGenerationError) break;
				const failure = lifecycleError(error);
				await this.#evolve(runId, generation, (draft) => {
					const child = draft.children.find((candidate) => candidate.id === childId)!;
					child.lastFailure = { kind: failure.kind, summary: safeFailureSummary(failure.kind, stage), at: this.#now().toISOString() };
				});
				const current = this.#context(plan, childId, runId, generation, signal).runtime;
				if (failure.kind === "correction_required" && failure.blockers.includes("child_head_moved")) {
					const reconciled = await this.#pipeline.reconcileChildHead(
						this.#context(plan, childId, runId, generation, signal),
					);
					if (signal.aborted) return "succeeded";
					if (reconciled.invalidated.verification !== true || reconciled.invalidated.review !== true
						|| reconciled.invalidated.integration !== true) {
						throw new ProductionLifecycleError("terminal", "child-head reconciliation did not invalidate downstream evidence");
					}
					const persisted = await this.#custom(runId, generation, (state, stateFence) =>
						reconcileProductionChildHead(state, stateFence, {
							childId,
							previousHead: reconciled.previousHead,
							head: reconciled.head,
							checkpoint: reconciled.checkpoint,
							now: this.#now(),
						}));
					await this.#acknowledgeCheckpoint(reconciled.checkpoint, persisted);
					stage = "verification";
					continue;
				}
				if (failure.kind === "stale_parent") {
					if (current.attempts >= current.maxAttempts) return this.#waitForChild(plan, childId, runId, generation, signal, "retry_budget_exhausted");
					await this.#evolve(runId, generation, (draft) => { draft.children.find((candidate) => candidate.id === childId)!.attempts += 1; });
					const before = this.#context(plan, childId, runId, generation, signal).runtime.ownership;
					if (!before) throw new ProductionLifecycleError("terminal", "stale child has no durable ownership");
					const refreshed = await this.#pipeline.refresh(this.#context(plan, childId, runId, generation, signal));
					if (!refreshed.workspace || !refreshed.effectKey) throw new ProductionLifecycleError("terminal", "parent refresh lacks an exact receipt");
					const persisted = await this.#custom(runId, generation, (state, stateFence) => refreshProductionChildOwnership(state, stateFence, {
						childId,
						outcome: refreshed.workspace!.claimId === before.claimId ? "rebased" : "reclaimed",
						previousClaimId: before.claimId,
						previousBaseHead: before.baseHead,
						newBinding: refreshed.workspace!,
						effectKey: refreshed.effectKey!,
						summary: refreshed.summary,
						now: this.#now(),
					}));
					await this.#acknowledgeCheckpoint(refreshed, persisted);
					stage = "verification";
					continue;
				}
				if (failure.kind === "correction_required") {
					if (current.corrections >= current.maxCorrections) return this.#waitForChild(plan, childId, runId, generation, signal, "correction_budget_exhausted");
					await this.#evolve(runId, generation, (draft) => {
						const child = draft.children.find((candidate) => candidate.id === childId)!;
						child.corrections += 1;
						child.stage = "correction";
					});
					const findings = correctionFindings(current);
					const corrected = await this.#pipeline.correct(this.#context(plan, childId, runId, generation, signal), findings);
					if (signal.aborted) return "succeeded";
					const persisted = await this.#evolve(runId, generation, (draft) => {
						const child = draft.children.find((candidate) => candidate.id === childId)!;
						child.stage = "verification";
						child.checkpoint = mergeCheckpoint(child.checkpoint, corrected);
						delete child.checkpoint.verification;
						delete child.checkpoint.review;
						delete child.checkpoint.integrationReceiptDigest;
					});
					await this.#acknowledgeCheckpoint(corrected, persisted);
					stage = "verification";
					continue;
				}
				if (failure.kind === "retryable") {
					if (current.attempts >= current.maxAttempts) return this.#waitForChild(plan, childId, runId, generation, signal, "retry_budget_exhausted");
					await this.#evolve(runId, generation, (draft) => { draft.children.find((candidate) => candidate.id === childId)!.attempts += 1; });
					continue;
				}
				if (failure.kind === "human_required") return this.#waitForChild(plan, childId, runId, generation, signal, "retry_budget_exhausted");
				await this.#evolve(runId, generation, (draft) => {
					const child = draft.children.find((candidate) => candidate.id === childId)!;
					child.status = "failed";
					child.stage = "failed";
				});
				throw failure;
			}
		}
		return "succeeded";
	}

	async #stage(
		plan: ProductionParentPlanDocument,
		childId: string,
		runId: string,
		generation: number,
		signal: AbortSignal,
		stage: ChildStage,
		invoke: (context: ProductionChildPipelineContext) => Promise<ProductionStageCheckpoint>,
	): Promise<ProductionStageCheckpoint> {
		await this.#evolve(runId, generation, (draft) => {
			const child = draft.children.find((candidate) => candidate.id === childId)!;
			child.status = "running";
			child.stage = stage;
		});
		const result = await invoke(this.#context(plan, childId, runId, generation, signal));
		if (signal.aborted) throw signal.reason ?? new Error("production stage cancelled");
		const persisted = await this.#evolve(runId, generation, (draft) => {
			const child = draft.children.find((candidate) => candidate.id === childId)!;
			child.checkpoint = mergeCheckpoint(child.checkpoint, result);
			if (!child.ownership && result.workspace) child.ownership = result.workspace;
		});
		await this.#acknowledgeCheckpoint(result, persisted);
		return result;
	}

	async #waitForChild(
		plan: ProductionParentPlanDocument,
		childId: string,
		runId: string,
		generation: number,
		signal: AbortSignal,
		reason: "retry_budget_exhausted" | "correction_budget_exhausted",
	): Promise<"waiting"> {
		const request = await this.#pipeline.requestIntervention(this.#context(plan, childId, runId, generation, signal), reason);
		const persisted = await this.#custom(runId, generation, (state, stateFence) => waitForProductionChildIntervention(state, stateFence, {
			childId,
			requestId: request.requestId,
			reason,
			...(request.pullRequest === undefined ? {} : { pullRequest: request.pullRequest }),
			...(request.head === undefined ? {} : { head: request.head }),
			now: this.#now(),
		}));
		await this.#acknowledge(request.effectKey, persisted);
		return "waiting";
	}

	async #acknowledge(effectKey: string | undefined, state: ProductionAutonomousState): Promise<void> {
		if (effectKey !== undefined) await this.#pipeline.acknowledge(effectKey, structuredClone(state));
	}

	async #acknowledgeCheckpoint(checkpoint: ProductionStageCheckpoint, state: ProductionAutonomousState): Promise<void> {
		const keys = [...new Set([
			...(checkpoint.effectKeys ?? []),
			...(checkpoint.effectKey === undefined ? [] : [checkpoint.effectKey]),
		])];
		for (const key of keys) await this.#pipeline.acknowledge(key, structuredClone(state));
	}

	#evolve(
		runId: string,
		generation: number,
		mutate: (draft: ProductionAutonomousState) => void,
	): Promise<ProductionAutonomousState> {
		return this.#custom(runId, generation, (state, stateFence) => evolveProductionState(state, stateFence, mutate, this.#now()));
	}

	#custom(
		runId: string,
		generation: number,
		build: (state: ProductionAutonomousState, stateFence: ProductionStateFence) => ProductionAutonomousState,
	): Promise<ProductionAutonomousState> {
		let result!: ProductionAutonomousState;
		const operation = this.#mutationTail.then(async () => {
			const current = this.#current;
			if (!current || current.runId !== runId || current.generation !== generation) throw new StaleProductionGenerationError();
			const stateFence = fence(current);
			const next = build(structuredClone(current), stateFence);
			result = await this.#store.compareAndSwap(stateFence, next);
			this.#current = result;
		});
		this.#mutationTail = operation.catch(() => undefined);
		return operation.then(() => structuredClone(result));
	}
}
