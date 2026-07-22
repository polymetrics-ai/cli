import type { ShepherdResumeCommand, ShepherdStartCommand } from "./arguments.ts";
import type {
	AutonomousChildState,
	AutonomousChildPlan,
	AutonomousShepherdRunState,
	AutonomousStateStore,
} from "./autonomous-state.ts";
import { selectReadyWork, validateDependencyGraph, type DependencyWorkItem } from "./dependency-graph.ts";

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
	timeoutMs: number;
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

interface ActiveRun {
	issue: number;
	runId: string;
	controller: AbortController;
	promise: Promise<AutonomousShepherdRunState>;
}

interface ChildSettlement {
	id: string;
	error?: unknown;
}

function dependencyItems(children: readonly AutonomousChildState[]): DependencyWorkItem[] {
	return children.map((child) => ({
		id: child.id,
		dependsOn: [...child.dependsOn],
		status: child.status,
		access: child.access,
		writeScopes: [...child.writeScopes],
	}));
}

function childStates(children: readonly AutonomousChildPlan[]): AutonomousChildState[] {
	validateDependencyGraph(children.map((child) => ({
		id: child.id,
		dependsOn: [...child.dependsOn],
		status: "pending" as const,
		access: child.access,
		writeScopes: [...child.writeScopes],
	})));
	return children.map((child) => ({
		...structuredClone(child),
		status: "pending",
		phase: "pending",
	}));
}

export class AutonomousShepherdController {
	readonly #store: AutonomousStateStore;
	readonly #intake: AutonomousIntakePort;
	readonly #lifecycle: AutonomousChildLifecyclePort;
	readonly #humanGate: AutonomousHumanGatePort;
	readonly #now: () => Date;
	readonly #newRunId: () => string;
	#active: ActiveRun | undefined;
	#saveTail: Promise<void> = Promise.resolve();
	#closed = false;
	#closePromise: Promise<void> | undefined;

	constructor(options: AutonomousShepherdControllerOptions) {
		this.#store = options.store;
		this.#intake = options.intake;
		this.#lifecycle = options.lifecycle;
		this.#humanGate = options.humanGate;
		this.#now = options.now ?? (() => new Date());
		this.#newRunId = options.newRunId ?? (() => crypto.randomUUID());
	}

	async status(issue: number): Promise<AutonomousShepherdRunState | undefined> {
		return this.#store.load(issue);
	}

	start(command: ShepherdStartCommand): Promise<AutonomousShepherdRunState> {
		return this.#launch(command.issue, (controller, runId) => this.#start(command, controller.signal, runId));
	}

	resume(command: ShepherdResumeCommand): Promise<AutonomousShepherdRunState> {
		return this.#launch(command.issue, (controller, runId) => this.#resume(command, controller.signal, runId));
	}

	async stop(issue: number): Promise<AutonomousShepherdRunState> {
		const active = this.#active;
		if (active?.issue === issue) {
			active.controller.abort(new Error("Shepherd stop requested"));
			let abortError: unknown;
			try {
				await this.#lifecycle.abort(active.runId);
			} catch (error) {
				abortError = error;
			}
			const state = await active.promise;
			if (abortError !== undefined) throw abortError;
			return state;
		}
		if (active) throw new Error(`a Shepherd run is active for issue #${active.issue}`);
		const state = await this.#store.load(issue);
		if (!state) throw new Error(`no autonomous Shepherd run exists for issue #${issue}`);
		if (state.status === "completed" || state.status === "failed") return state;
		state.status = "stopped";
		state.stage = "SCHEDULE";
		for (const child of state.children) {
			if (child.status !== "succeeded") {
				child.status = "pending";
				child.phase = "pending";
			}
		}
		state.updatedAt = this.#timestamp();
		await this.#persist(state);
		return structuredClone(state);
	}

	shutdown(): Promise<void> {
		if (!this.#closePromise) {
			this.#closed = true;
			this.#closePromise = (async () => {
				const active = this.#active;
				if (active) await this.stop(active.issue);
				await Promise.all([this.#lifecycle.close(), this.#humanGate.close()]);
			})();
		}
		return this.#closePromise;
	}

	#launch(
		issue: number,
		run: (controller: AbortController, runId: string) => Promise<AutonomousShepherdRunState>,
	): Promise<AutonomousShepherdRunState> {
		if (this.#closed) return Promise.reject(new Error("autonomous Shepherd controller is closed"));
		if (this.#active) return Promise.reject(new Error(`a Shepherd run is already active for issue #${this.#active.issue}`));
		const controller = new AbortController();
		const runId = this.#newRunId();
		const promise = run(controller, runId);
		const active = { issue, runId, controller, promise };
		this.#active = active;
		void promise.then(
			() => { if (this.#active === active) this.#active = undefined; },
			() => { if (this.#active === active) this.#active = undefined; },
		);
		return promise;
	}

	async #start(
		command: ShepherdStartCommand,
		signal: AbortSignal,
		runId: string,
	): Promise<AutonomousShepherdRunState> {
		if (await this.#store.load(command.issue)) {
			throw new Error(`autonomous Shepherd state already exists for issue #${command.issue}; use resume`);
		}
		const plan = await this.#intake.load(command.issue, signal);
		if (signal.aborted) throw new Error("autonomous Shepherd start was cancelled during intake");
		const timestamp = this.#timestamp();
		const state: AutonomousShepherdRunState = {
			schemaVersion: 2,
			kind: "autonomous",
			issue: command.issue,
			...(command.pr === undefined ? {} : { pr: command.pr }),
			planId: plan.planId,
			runId,
			generation: 1,
			status: "running",
			stage: "SCHEDULE",
			maxConcurrency: command.maxConcurrency,
			timeoutMs: command.timeoutMs,
			createdAt: timestamp,
			updatedAt: timestamp,
			children: childStates(plan.children),
		};
		await this.#persist(state);
		return this.#drive(state, signal);
	}

	async #resume(
		command: ShepherdResumeCommand,
		signal: AbortSignal,
		runId: string,
	): Promise<AutonomousShepherdRunState> {
		const state = await this.#store.load(command.issue);
		if (!state) throw new Error(`no autonomous Shepherd run exists for issue #${command.issue}`);
		state.runId = runId;
		state.generation += 1;
		state.maxConcurrency = command.maxConcurrency;
		state.timeoutMs = command.timeoutMs;
		if (command.pr !== undefined) state.pr = command.pr;
		state.updatedAt = this.#timestamp();
		if (state.status === "waiting_human" && state.humanGate) {
			state.status = "running";
			await this.#persist(state);
			const observation = await this.#humanGate.observe(structuredClone(state), signal);
			state.humanGate.status = observation;
			state.status = observation === "merged" ? "completed" : observation === "rejected" ? "failed" : "waiting_human";
			state.stage = observation === "merged" ? "COMPLETE" : observation === "rejected" ? "BLOCKED" : "HUMAN_DECISION";
			state.updatedAt = this.#timestamp();
			await this.#persist(state);
			return structuredClone(state);
		}
		delete state.humanGate;
		state.status = "running";
		state.stage = "SCHEDULE";
		for (const child of state.children) {
			if (child.status !== "succeeded") {
				child.status = "pending";
				child.phase = "pending";
				delete child.summary;
			}
		}
		validateDependencyGraph(dependencyItems(state.children));
		await this.#persist(state);
		return this.#drive(state, signal);
	}

	async #drive(state: AutonomousShepherdRunState, signal: AbortSignal): Promise<AutonomousShepherdRunState> {
		const active = new Map<string, Promise<ChildSettlement>>();
		let failure: unknown;
		while (!signal.aborted && failure === undefined) {
			const selection = selectReadyWork(dependencyItems(state.children), {
				maxConcurrency: state.maxConcurrency,
				allowMutating: true,
			});
			if (selection.kind === "selected") {
				for (const id of selection.itemIds) {
					const child = state.children.find((candidate) => candidate.id === id);
					if (!child) throw new Error(`selected unknown child ${id}`);
					child.status = "running";
					child.phase = "execute";
				}
				state.stage = "EXECUTE";
				state.updatedAt = this.#timestamp();
				await this.#persist(state);
				for (const id of selection.itemIds) {
					const child = state.children.find((candidate) => candidate.id === id)!;
					const task = this.#runChild(state, child, signal).then(
						() => ({ id }),
						(error) => ({ id, error }),
					);
					active.set(id, task);
				}
				continue;
			}
			if (selection.kind === "complete" && active.size === 0) break;
			if (active.size === 0) {
				failure = new Error(`autonomous scheduler blocked: ${selection.kind === "blocked" ? selection.blocker : selection.kind}`);
				break;
			}
			const settled = await Promise.race(active.values());
			active.delete(settled.id);
			if (settled.error !== undefined) failure = settled.error;
		}

		if (signal.aborted || failure !== undefined) {
			if (failure !== undefined && !signal.aborted) {
				for (const child of state.children) {
					if (child.status === "running") child.status = "blocked";
				}
			}
			const remaining = await Promise.all(active.values());
			failure ??= remaining.find((settlement) => settlement.error !== undefined)?.error;
			for (const child of state.children) {
				if (child.status !== "succeeded") {
					child.status = failure === undefined ? "pending" : child.status === "failed" ? "failed" : "blocked";
					child.phase = failure === undefined ? "pending" : child.phase;
				}
			}
			state.status = failure === undefined ? "stopped" : "failed";
			state.stage = failure === undefined ? "SCHEDULE" : "BLOCKED";
			state.updatedAt = this.#timestamp();
			await this.#persist(state);
			return structuredClone(state);
		}

		state.stage = "HUMAN_DECISION";
		state.updatedAt = this.#timestamp();
		const request = await this.#humanGate.request(structuredClone(state), signal);
		state.humanGate = { kind: "parent_merge", requestId: request.requestId, status: "pending" };
		state.status = "waiting_human";
		state.updatedAt = this.#timestamp();
		await this.#persist(state);
		return structuredClone(state);
	}

	async #runChild(
		state: AutonomousShepherdRunState,
		child: AutonomousChildState,
		signal: AbortSignal,
	): Promise<void> {
		const context: AutonomousChildContext = {
			parentIssue: state.issue,
			runId: state.runId,
			generation: state.generation,
			child: structuredClone(child),
			timeoutMs: state.timeoutMs,
			signal,
		};
		try {
			const execution = await this.#lifecycle.execute(context);
			if (signal.aborted) return;
			await this.#advance(state, child, "verify");
			await this.#lifecycle.verify(context);
			if (signal.aborted) return;
			await this.#advance(state, child, "review");
			await this.#lifecycle.review(context);
			if (signal.aborted) return;
			await this.#advance(state, child, "integrate");
			const integration = await this.#lifecycle.integrate(context);
			if (signal.aborted) return;
			child.status = "succeeded";
			child.phase = "succeeded";
			child.summary = integration.summary || execution.summary;
			state.stage = "SCHEDULE";
			state.updatedAt = this.#timestamp();
			await this.#persist(state);
		} catch (error) {
			if (signal.aborted) return;
			child.status = "failed";
			child.phase = "failed";
			child.summary = "child lifecycle failed";
			state.updatedAt = this.#timestamp();
			await this.#persist(state);
			throw error;
		}
	}

	async #advance(
		state: AutonomousShepherdRunState,
		child: AutonomousChildState,
		phase: "verify" | "review" | "integrate",
	): Promise<void> {
		child.phase = phase;
		state.stage = phase === "verify" ? "VERIFY" : phase === "review" ? "REVIEW" : "INTEGRATE";
		state.updatedAt = this.#timestamp();
		await this.#persist(state);
	}

	#timestamp(): string {
		const value = this.#now();
		if (!Number.isFinite(value.valueOf())) throw new Error("autonomous Shepherd clock is invalid");
		return value.toISOString();
	}

	async #persist(state: AutonomousShepherdRunState): Promise<void> {
		const snapshot = structuredClone(state);
		const operation = this.#saveTail.then(() => this.#store.save(snapshot));
		this.#saveTail = operation.catch(() => undefined);
		await operation;
	}
}
