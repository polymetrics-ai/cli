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
	async load(issue) {
		return structuredClone(this.states.get(issue));
	}
	async save(state) {
		this.states.set(state.issue, structuredClone(state));
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
		clock: () => "2026-07-21T08:30:00Z",
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

test("stop and parent shutdown only address owned runner handles", async () => {
	const harness = makeHarness();
	await harness.store.save({
		schemaVersion: 1,
		issue: 397,
		runId: "run-owned",
		generation: 1,
		status: "running",
		candidateHead: head,
		validationNonce: "nonce-old",
		createdAt: "2026-07-21T08:00:00Z",
		updatedAt: "2026-07-21T08:00:00Z",
		lanes: [],
	});
	const stopped = await harness.controller.stop(397);
	assert.equal(stopped.status, "stopped");
	assert.deepEqual(harness.aborted, ["run-owned"]);
	await harness.controller.shutdown();
	assert.equal(harness.closed, true);
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

	await store.save({
		schemaVersion: 1,
		issue: 397,
		pr: 438,
		runId: "run-owned",
		generation: 2,
		status: "running",
		candidateHead: head,
		validationNonce: "nonce-old",
		createdAt: "2026-07-21T08:00:00Z",
		updatedAt: "2026-07-21T08:00:00Z",
		lanes: [
			{ id: "scout", role: "scout", mutating: false, dependsOn: [], status: "running" },
		],
	});
	const stopped = await haltedHarness.controller.stop(397);
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
