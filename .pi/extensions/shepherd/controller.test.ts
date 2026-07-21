import assert from "node:assert/strict";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import { ShepherdController } from "./controller.ts";
import { FileStateStore } from "./state-store.ts";

const head = "a".repeat(40);
const perfectDimensions = {
	correctStage: 1,
	artifactValid: 1,
	gatesRespected: 1,
	realProgress: 1,
	noHallucination: 1,
	noConflict: 1,
};

class MemoryStore {
	states = new Map();
	lease;
	async load(issue) {
		return structuredClone(this.states.get(issue));
	}
	async save(state) {
		this.states.set(state.issue, structuredClone(state));
	}
	async acquireLease(claim) {
		if (this.lease) throw new Error("another Pi process owns the Shepherd lease");
		const lease = {
			claim,
			async assertOwned() {
				if (this.owner.lease !== lease) throw new Error("Shepherd lease ownership was lost");
			},
			async release() {
				if (this.owner.lease === lease) this.owner.lease = undefined;
			},
			owner: this,
		};
		this.lease = lease;
		return lease;
	}
}

function command(action = "canary") {
	return {
		action,
		issue: 397,
		pr: 438,
		readOnly: true,
		backend: "sdk-inproc",
		experimental: true,
		maxConcurrency: 2,
		timeoutMs: 60_000,
	};
}

function makeHarness(overrides = {}) {
	const store = new MemoryStore();
	let active = 0;
	let maxActive = 0;
	const requests = [];
	const aborted = [];
	let closed = false;
	const runner = {
		async run(request) {
			requests.push(request);
			active += 1;
			maxActive = Math.max(maxActive, active);
			await new Promise((resolve) => setTimeout(resolve, 5));
			active -= 1;
			return {
				...request.binding,
				summary: `${request.laneId} completed`,
				dimensions: perfectDimensions,
				observedMutation: false,
			};
		},
		async abort(runId) {
			aborted.push(runId);
		},
		async close() {
			closed = true;
		},
	};
	const targetEvidence = {
		async capture() {
			return {
				cwd: "/tmp/read-only-pr",
				branch: "feat/cli-architecture-v2",
				candidateHead: head,
				clean: true,
			};
		},
	};
	const controller = new ShepherdController({
		store,
		runner,
		targetEvidence,
		clock: () => "2026-07-21T08:30:00.000Z",
		createRunId: () => "run-1",
		createNonce: () => "nonce-1234567890",
		...overrides,
	});
	return { controller, store, requests, aborted, get maxActive() { return maxActive; }, get closed() { return closed; } };
}

test("runs independent read-only lanes concurrently and persists a completed rating", async () => {
	const harness = makeHarness();
	const result = await harness.controller.start(command());
	assert.equal(harness.maxActive, 2);
	assert.deepEqual(harness.requests.map((request) => request.laneId).sort(), ["scout", "validator"]);
	assert.ok(harness.requests.every((request) => request.readOnly));
	assert.ok(harness.requests.every((request) => request.tools.length === 0));
	assert.deepEqual(harness.requests.map((request) => request.thinking), ["xhigh", "xhigh"]);
	assert.equal(result.status, "completed");
	assert.equal(result.score, 1);
	assert.equal((await harness.store.load(397)).status, "completed");
});

test("revalidates the exact target after the lanes and halts if the head changes", async () => {
	let captures = 0;
	const harness = makeHarness({
		targetEvidence: {
			async capture() {
				captures += 1;
				return {
					cwd: "/tmp/read-only-pr",
					branch: "feat/cli-architecture-v2",
					candidateHead: captures === 1 ? head : "b".repeat(40),
					clean: true,
				};
			},
		},
	});
	const result = await harness.controller.start(command());
	assert.equal(captures, 2);
	assert.equal(result.status, "halted");
	assert.ok(result.hardGates.includes("target_changed"));
});

test("halts if host-verified PR gate evidence changes during the run", async () => {
	let captures = 0;
	const harness = makeHarness({
		targetEvidence: {
			async capture() {
				captures += 1;
				return {
					cwd: "/tmp/read-only-pr",
					branch: "feat/cli-architecture-v2",
					candidateHead: head,
					clean: true,
					pr: 438,
					statusChecks: [{ name: "verify", status: "COMPLETED", conclusion: captures === 1 ? "SUCCESS" : "FAILURE" }],
				};
			},
		},
	});
	const result = await harness.controller.start(command());
	assert.equal(result.status, "halted");
	assert.ok(result.hardGates.includes("target_changed"));
});

test("a stop request cannot be overwritten by a finishing lane", async () => {
	let release;
	const gate = new Promise((resolve) => { release = resolve; });
	let started;
	const startedPromise = new Promise((resolve) => { started = resolve; });
	const harness = makeHarness({
		runner: {
			async run(request) {
				started();
				await gate;
				return {
					...request.binding,
					summary: "finished after stop",
					dimensions: perfectDimensions,
					observedMutation: false,
				};
			},
			async abort() {},
			async close() {},
		},
	});
	const pending = harness.controller.start(command());
	await startedPromise;
	const stopping = harness.controller.stop(397);
	release();
	await stopping;
	const result = await pending;
	assert.equal(result.status, "stopped");
	assert.equal((await harness.store.load(397)).status, "stopped");
});

test("stop during initial target capture prevents every lane from starting", async () => {
	let releaseCapture;
	const captureGate = new Promise((resolve) => { releaseCapture = resolve; });
	let captureStarted;
	const captureStartedGate = new Promise((resolve) => { captureStarted = resolve; });
	let runs = 0;
	const harness = makeHarness({
		targetEvidence: {
			async capture() {
				captureStarted();
				await captureGate;
				return {
					cwd: "/tmp/read-only-pr",
					branch: "feat/cli-architecture-v2",
					candidateHead: head,
					clean: true,
				};
			},
		},
		runner: {
			async run() { runs += 1; throw new Error("must not run"); },
			async abort() {},
			async close() {},
		},
	});
	const pending = harness.controller.start(command());
	await captureStartedGate;
	const stopping = harness.controller.stop(397);
	releaseCapture();
	const [result, stopped] = await Promise.all([pending, stopping]);
	assert.equal(result.status, "stopped");
	assert.equal(stopped.status, "stopped");
	assert.equal(runs, 0);
	assert.equal((await harness.store.load(397)).status, "stopped");
});

test("rejects duplicate active ownership before dispatch", async () => {
	const harness = makeHarness();
	await harness.store.save({
		schemaVersion: 1,
		issue: 397,
		runId: "already-running",
		generation: 1,
		status: "running",
		candidateHead: head,
		validationNonce: "nonce-old",
		createdAt: "2026-07-21T08:00:00Z",
		updatedAt: "2026-07-21T08:00:00Z",
		lanes: [],
	});
	await assert.rejects(harness.controller.start(command("start")), /already active/);
	assert.equal(harness.requests.length, 0);
});

test("halts the run when perfect-scored evidence is stale", async () => {
	const harness = makeHarness({
		runner: {
			async run(request) {
				return {
					...request.binding,
					candidateHead: "b".repeat(40),
					summary: "claimed success",
					dimensions: perfectDimensions,
					observedMutation: false,
				};
			},
			async abort() {},
			async close() {},
		},
	});
	const result = await harness.controller.start(command());
	assert.equal(result.status, "halted");
	assert.ok(result.hardGates.includes("stale_head"));
});

test("resume creates a fresh generation, head binding, and nonce", async () => {
	const harness = makeHarness({ createRunId: () => "run-2", createNonce: () => "nonce-fresh-12345" });
	await harness.store.save({
		schemaVersion: 1,
		issue: 397,
		pr: 438,
		runId: "run-1",
		generation: 1,
		status: "interrupted",
		candidateHead: "b".repeat(40),
		validationNonce: "nonce-old",
		createdAt: "2026-07-21T08:00:00Z",
		updatedAt: "2026-07-21T08:10:00Z",
		lanes: [],
	});
	const result = await harness.controller.resume(command("resume"));
	assert.equal(result.generation, 2);
	assert.equal(result.runId, "run-2");
	assert.equal(result.candidateHead, head);
	assert.equal(result.validationNonce, "nonce-fresh-12345");
});

test("resume inherits a persisted PR when --pr is omitted", async () => {
	const captures = [];
	const harness = makeHarness({
		createRunId: () => "run-2",
		targetEvidence: {
			async capture(capturedCommand) {
				captures.push(structuredClone(capturedCommand));
				return {
					cwd: "/tmp/read-only-pr",
					branch: "feat/cli-architecture-v2",
					candidateHead: head,
					clean: true,
					pr: capturedCommand.pr,
				};
			},
		},
	});
	await harness.store.save({
		schemaVersion: 1,
		issue: 397,
		pr: 438,
		runId: "run-1",
		generation: 1,
		status: "interrupted",
		candidateHead: "b".repeat(40),
		validationNonce: "nonce-old",
		createdAt: "2026-07-21T08:00:00Z",
		updatedAt: "2026-07-21T08:10:00Z",
		lanes: [],
	});
	const resume = command("resume");
	delete resume.pr;
	const result = await harness.controller.resume(resume);
	assert.equal(result.pr, 438);
	assert.deepEqual(captures.map((capture) => capture.pr), [438, 438]);
	assert.equal((await harness.store.load(397)).pr, 438);
});

test("resume rejects a PR that differs from the persisted target", async () => {
	let captures = 0;
	const harness = makeHarness({
		targetEvidence: {
			async capture() {
				captures += 1;
				throw new Error("must not capture a mismatched target");
			},
		},
	});
	await harness.store.save({
		schemaVersion: 1,
		issue: 397,
		pr: 438,
		runId: "run-1",
		generation: 1,
		status: "interrupted",
		candidateHead: head,
		validationNonce: "nonce-old",
		createdAt: "2026-07-21T08:00:00Z",
		updatedAt: "2026-07-21T08:10:00Z",
		lanes: [],
	});
	await assert.rejects(
		harness.controller.resume({ ...command("resume"), pr: 999 }),
		/persisted PR #438.*requested PR #999/i,
	);
	assert.equal(captures, 0);
});

test("stop refuses to rewrite an unowned persisted run", async () => {
	const harness = makeHarness();
	await harness.store.save({
		schemaVersion: 1,
		issue: 397,
		runId: "run-owned",
		generation: 1,
		status: "completed",
		candidateHead: head,
		validationNonce: "nonce-old",
		createdAt: "2026-07-21T08:00:00Z",
		updatedAt: "2026-07-21T08:00:00Z",
		lanes: [],
	});
	await assert.rejects(harness.controller.stop(397), /not owned.*Pi session/i);
	assert.equal((await harness.store.load(397)).status, "completed");
	assert.deepEqual(harness.aborted, []);
	await harness.controller.shutdown();
	assert.equal(harness.closed, true);
});

test("parent shutdown cancels owned work and persists interrupted", async () => {
	let release;
	const gate = new Promise((resolve) => { release = resolve; });
	let started;
	const startedGate = new Promise((resolve) => { started = resolve; });
	const store = new MemoryStore();
	const aborted = [];
	const controller = new ShepherdController({
		store,
		targetEvidence: {
			async capture() {
				return { cwd: "/tmp/read-only-pr", branch: "feat/cli-architecture-v2", candidateHead: head, clean: true };
			},
		},
		runner: {
			async run(request) {
				started();
				await gate;
				return { ...request.binding, summary: "late", dimensions: perfectDimensions, observedMutation: false };
			},
			async abort(runId) { aborted.push(runId); release(); },
			async close() { release(); },
		},
		clock: () => "2026-07-21T08:30:00.000Z",
		createRunId: () => "run-shutdown",
		createNonce: () => "nonce-shutdown-123",
	});
	const pending = controller.start(command());
	await startedGate;
	await controller.shutdown();
	const result = await pending;
	assert.equal(result.status, "interrupted");
	assert.equal((await store.load(397)).status, "interrupted");
	assert.deepEqual(aborted, ["run-shutdown"]);
});

test("hard-gate and stop states round-trip through the validating file store", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-controller-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const store = new FileStateStore(root);
	const staleRunner = {
		async run(request) {
			return {
				...request.binding,
				candidateHead: "b".repeat(40),
				summary: "claimed success",
				dimensions: perfectDimensions,
				observedMutation: false,
			};
		},
		async abort() {},
		async close() {},
	};
	const haltedHarness = makeHarness({ store, runner: staleRunner });
	const halted = await haltedHarness.controller.start(command());
	assert.equal(halted.status, "halted");
	assert.equal((await store.load(397)).lanes.every((lane) => lane.status === "halted"), true);

	let release;
	const gate = new Promise((resolve) => { release = resolve; });
	let started;
	const startedGate = new Promise((resolve) => { started = resolve; });
	const runningHarness = makeHarness({
		store,
		createRunId: () => "run-owned",
		runner: {
			async run(request) {
				started();
				await gate;
				return { ...request.binding, summary: "late", dimensions: perfectDimensions, observedMutation: false };
			},
			async abort() {},
			async close() {},
		},
	});
	const pending = runningHarness.controller.start(command());
	await startedGate;
	const stopping = runningHarness.controller.stop(397);
	release();
	const [result, stopped] = await Promise.all([pending, stopping]);
	assert.equal(result.status, "stopped");
	assert.equal(stopped.status, "stopped");
	assert.equal((await store.load(397)).lanes[0].status, "stopped");
});

test("resume treats a persisted running state as interrupted after process restart", async () => {
	const harness = makeHarness({ createRunId: () => "run-after-restart", createNonce: () => "nonce-after-restart" });
	await harness.store.save({
		schemaVersion: 1,
		issue: 397,
		pr: 438,
		runId: "orphaned-run",
		generation: 4,
		status: "running",
		candidateHead: "b".repeat(40),
		validationNonce: "nonce-old",
		createdAt: "2026-07-21T08:00:00Z",
		updatedAt: "2026-07-21T08:00:00Z",
		lanes: [
			{ id: "scout", role: "scout", mutating: false, dependsOn: [], status: "running" },
		],
	});
	const resumed = await harness.controller.resume(command("resume"));
	assert.equal(resumed.generation, 5);
	assert.equal(resumed.runId, "run-after-restart");
	assert.equal(resumed.status, "completed");
});

test("a global file lease prevents two controllers from dispatching the same repository", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-lease-controller-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	let release;
	const gate = new Promise((resolve) => { release = resolve; });
	let started;
	const startedGate = new Promise((resolve) => { started = resolve; });
	let firstRuns = 0;
	let secondRuns = 0;
	const targetEvidence = {
		async capture() {
			return { cwd: "/tmp/read-only-pr", branch: "feat/cli-architecture-v2", candidateHead: head, clean: true };
		},
	};
	const first = new ShepherdController({
		store: new FileStateStore(root),
		targetEvidence,
		runner: {
			async run(request) {
				firstRuns += 1;
				started();
				await gate;
				return { ...request.binding, summary: "first", dimensions: perfectDimensions, observedMutation: false };
			},
			async abort() {},
			async close() {},
		},
		createRunId: () => "run-first",
		createNonce: () => "nonce-first-12345",
	});
	const second = new ShepherdController({
		store: new FileStateStore(root),
		targetEvidence,
		runner: {
			async run(request) {
				secondRuns += 1;
				return { ...request.binding, summary: "second", dimensions: perfectDimensions, observedMutation: false };
			},
			async abort() {},
			async close() {},
		},
		createRunId: () => "run-second",
		createNonce: () => "nonce-second-1234",
	});
	const pending = first.start(command());
	await startedGate;
	await assert.rejects(second.start({ ...command(), issue: 471 }), /lease|active|another Pi/i);
	assert.equal(secondRuns, 0);
	release();
	const result = await pending;
	assert.equal(result.status, "completed");
	assert.equal(firstRuns, 2);
});

test("a lane persistence failure cancels and joins a running sibling before releasing the lease", async () => {
	let releaseSibling;
	const siblingGate = new Promise((resolve) => { releaseSibling = resolve; });
	let siblingStarted;
	const siblingStartedGate = new Promise((resolve) => { siblingStarted = resolve; });
	let persistenceFailed;
	const persistenceFailedGate = new Promise((resolve) => { persistenceFailed = resolve; });
	let runningSaves = 0;
	let leaseReleased = false;
	const store = new MemoryStore();
	const originalSave = store.save.bind(store);
	store.save = async (state) => {
		if (state.status === "running" && state.lanes.some((lane) => lane.status === "running")) {
			runningSaves += 1;
			if (runningSaves === 2) {
				persistenceFailed();
				throw new Error("injected lane persistence failure");
			}
		}
		await originalSave(state);
	};
	const originalAcquire = store.acquireLease.bind(store);
	store.acquireLease = async (claim) => {
		const lease = await originalAcquire(claim);
		const release = lease.release.bind(lease);
		lease.release = async () => {
			leaseReleased = true;
			await release();
		};
		return lease;
	};
	let aborts = 0;
	const controller = new ShepherdController({
		store,
		targetEvidence: {
			async capture() {
				return { cwd: "/tmp/read-only-pr", branch: "feat", candidateHead: head, clean: true };
			},
		},
		runner: {
			async run(request) {
				siblingStarted();
				await siblingGate;
				return { ...request.binding, summary: "joined", dimensions: perfectDimensions, observedMutation: false };
			},
			async abort() { aborts += 1; },
			async close() {},
		},
		createRunId: () => "run-structured",
		createNonce: () => "nonce-structured-1",
	});
	const pending = controller.start(command("start"));
	let settled = false;
	void pending.then(() => { settled = true; }, () => { settled = true; });
	await siblingStartedGate;
	await persistenceFailedGate;
	await new Promise((resolve) => setImmediate(resolve));
	try {
		assert.equal(settled, false);
		assert.equal(leaseReleased, false);
		assert.equal(aborts, 1);
	} finally {
		releaseSibling();
	}
	await assert.rejects(pending, /injected lane persistence failure/);
	assert.equal(leaseReleased, true);
});

test("terminal commit makes stop unowned before lease release completes", async () => {
	let releaseLease;
	const releaseGate = new Promise((resolve) => { releaseLease = resolve; });
	let releaseStarted;
	const releaseStartedGate = new Promise((resolve) => { releaseStarted = resolve; });
	const store = new MemoryStore();
	const originalAcquire = store.acquireLease.bind(store);
	store.acquireLease = async (claim) => {
		const lease = await originalAcquire(claim);
		const release = lease.release.bind(lease);
		lease.release = async () => {
			releaseStarted();
			await releaseGate;
			await release();
		};
		return lease;
	};
	const harness = makeHarness({ store });
	const starting = harness.controller.start(command("start"));
	await releaseStartedGate;
	const stopping = harness.controller.stop(397);
	let stopSettled = false;
	void stopping.then(() => { stopSettled = true; }, () => { stopSettled = true; });
	await new Promise((resolve) => setImmediate(resolve));
	try {
		assert.equal(stopSettled, true);
	} finally {
		releaseLease();
	}
	await assert.rejects(stopping, /not owned.*Pi session/i);
	assert.equal((await starting).status, "completed");
	assert.equal((await store.load(397)).status, "completed");
});

test("shutdown aggregates abort and close failures after owned work exits", async () => {
	let release;
	const gate = new Promise((resolve) => { release = resolve; });
	let started;
	const startedGate = new Promise((resolve) => { started = resolve; });
	const harness = makeHarness({
		runner: {
			async run(request) {
				started();
				await gate;
				return { ...request.binding, summary: "late", dimensions: perfectDimensions, observedMutation: false };
			},
			async abort() { throw new Error("abort cleanup failed"); },
			async close() {
				release();
				throw new Error("runner close failed");
			},
		},
	});
	const starting = harness.controller.start(command("start"));
	await startedGate;
	await assert.rejects(
		harness.controller.shutdown(),
		(error) => error instanceof AggregateError
			&& error.errors.some((entry) => /abort cleanup failed/.test(String(entry)))
			&& error.errors.some((entry) => /runner close failed/.test(String(entry))),
	);
	assert.equal((await starting).status, "interrupted");
});
