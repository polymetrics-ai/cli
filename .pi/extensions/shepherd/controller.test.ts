import assert from "node:assert/strict";
import test from "node:test";

import { ShepherdController } from "./controller.ts";

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
	assert.deepEqual(harness.requests.map((request) => request.thinking), ["xhigh", "xhigh"]);
	assert.equal(result.status, "completed");
	assert.equal(result.score, 1);
	assert.equal((await harness.store.load(397)).status, "completed");
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
