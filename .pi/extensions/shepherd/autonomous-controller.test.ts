import assert from "node:assert/strict";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import { AutonomousShepherdController } from "./autonomous-controller.ts";
import { AutonomousFileStateStore, type AutonomousChildPlan } from "./autonomous-state.ts";

function deferred<T>() {
	let resolve!: (value: T) => void;
	let reject!: (reason?: unknown) => void;
	const promise = new Promise<T>((accept, decline) => { resolve = accept; reject = decline; });
	return { promise, resolve, reject };
}

function child(id: string, issue: number, dependsOn: string[], writeScope: string): AutonomousChildPlan {
	return { id, issue, title: id, task: `implement ${id}`, dependsOn, access: "mutating", writeScopes: [writeScope] };
}

function command(action: "start" | "resume") {
	return {
		action,
		issue: 479,
		backend: "sdk-inproc" as const,
		maxConcurrency: 2,
		timeoutMs: 30_000,
	};
}

async function eventually(assertion: () => void): Promise<void> {
	for (let attempt = 0; attempt < 100; attempt += 1) {
		try { assertion(); return; } catch { await new Promise((resolve) => setImmediate(resolve)); }
	}
	assertion();
}

test("start overlaps independent children, stop joins, resume continues dependencies, and status persists the human wait", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-mvp-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const store = new AutonomousFileStateStore(root);
	const plans = [
		child("alpha", 501, [], "owned/alpha"),
		child("beta", 502, [], "owned/beta"),
		child("gamma", 503, ["alpha"], "owned/gamma"),
	];
	let intakeCalls = 0;
	const gates = new Map<string, ReturnType<typeof deferred<{ summary: string }>>>();
	const starts: string[] = [];
	const active = new Set<string>();
	let maxActive = 0;
	const abortCalls: string[] = [];
	const effects: string[] = [];
	const lifecycle = {
		async execute({ child: item }) {
			starts.push(item.id);
			active.add(item.id);
			maxActive = Math.max(maxActive, active.size);
			const gate = deferred<{ summary: string }>();
			gates.set(item.id, gate);
			const result = await gate.promise;
			active.delete(item.id);
			return result;
		},
		async verify({ child: item }) { effects.push(`verify:${item.id}`); return { summary: "verified" }; },
		async review({ child: item }) { effects.push(`review:${item.id}`); return { summary: "reviewed" }; },
		async integrate({ child: item }) { effects.push(`integrate:${item.id}`); return { summary: "integrated" }; },
		async abort(runId: string) { abortCalls.push(runId); },
		async close() {},
	};
	const options = {
		store,
		intake: {
			async load() { intakeCalls += 1; return { planId: "plan-479", children: plans }; },
		},
		lifecycle,
		humanGate: {
			async request() { effects.push("request_human_gate"); return { requestId: "merge-479" }; },
			async observe() { return "pending" as const; },
			async close() {},
		},
		newRunId: (() => { let id = 0; return () => `run-${++id}`; })(),
		now: (() => { let tick = 0; return () => new Date(Date.UTC(2026, 6, 22, 10, 0, tick++)); })(),
	};
	const controller = new AutonomousShepherdController(options);
	const firstRun = controller.start(command("start"));
	await eventually(() => assert.deepEqual(starts, ["alpha", "beta"]));
	assert.equal(intakeCalls, 1);
	assert.equal(maxActive, 2);
	assert.equal(starts.includes("gamma"), false);

	gates.get("alpha")?.resolve({ summary: "alpha implemented" });
	await eventually(() => assert.deepEqual(starts, ["alpha", "beta", "gamma"]));
	const observer = new AutonomousShepherdController(options);
	const running = await observer.status(479);
	assert.equal(running?.children.find((item) => item.id === "alpha")?.status, "succeeded");
	assert.equal(running?.children.find((item) => item.id === "beta")?.status, "running");
	assert.equal(running?.children.find((item) => item.id === "gamma")?.status, "running");

	let stopSettled = false;
	const stopping = controller.stop(479).then((state) => { stopSettled = true; return state; });
	gates.get("beta")?.resolve({ summary: "cancelled beta" });
	await new Promise((resolve) => setImmediate(resolve));
	assert.equal(stopSettled, false, "stop must join every accepted child");
	gates.get("gamma")?.resolve({ summary: "cancelled gamma" });
	const [stopped, firstResult] = await Promise.all([stopping, firstRun]);
	assert.equal(stopped.status, "stopped");
	assert.equal(firstResult.status, "stopped");
	assert.equal(abortCalls.length, 1);
	assert.equal(stopped.children.find((item) => item.id === "alpha")?.status, "succeeded");
	assert.deepEqual(
		stopped.children.filter((item) => item.id !== "alpha").map((item) => item.status),
		["pending", "pending"],
	);

	starts.length = 0;
	const resumedController = new AutonomousShepherdController(options);
	const resumed = resumedController.resume(command("resume"));
	await eventually(() => assert.deepEqual(starts, ["beta", "gamma"]));
	assert.equal(intakeCalls, 1, "resume must use the durable plan");
	gates.get("beta")?.resolve({ summary: "beta implemented" });
	gates.get("gamma")?.resolve({ summary: "gamma implemented" });
	const final = await resumed;
	assert.equal(final.status, "waiting_human");
	assert.equal(final.stage, "HUMAN_DECISION");
	assert.deepEqual(final.humanGate, { kind: "parent_merge", requestId: "merge-479", status: "pending" });
	assert.equal(final.generation, 2);
	assert.equal((await observer.status(479))?.humanGate?.requestId, "merge-479");
	for (const id of ["alpha", "beta", "gamma"]) {
		const childEffects = effects.filter((effect) => effect.endsWith(`:${id}`));
		assert.deepEqual(childEffects, [`verify:${id}`, `review:${id}`, `integrate:${id}`]);
	}
	assert.equal(effects.at(-1), "request_human_gate");
	assert.equal("mergeMain" in resumedController, false);
});

test("a failed child aborts and joins its live sibling before the run settles", { timeout: 2_000 }, async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-mvp-failure-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const sibling = deferred<{ summary: string }>();
	let abortCalls = 0;
	const controller = new AutonomousShepherdController({
		store: new AutonomousFileStateStore(root),
		intake: {
			async load() {
				return {
					planId: "failure-plan",
					children: [child("alpha", 501, [], "owned/alpha"), child("beta", 502, [], "owned/beta")],
				};
			},
		},
		lifecycle: {
			async execute({ child: item }) {
				if (item.id === "alpha") throw new Error("alpha failed");
				return sibling.promise;
			},
			async verify() { return { summary: "verified" }; },
			async review() { return { summary: "reviewed" }; },
			async integrate() { return { summary: "integrated" }; },
			async abort() {
				abortCalls += 1;
				sibling.reject(new Error("beta aborted"));
			},
			async close() {},
		},
		humanGate: {
			async request() { throw new Error("human gate must not be reached"); },
			async observe() { return "pending" as const; },
			async close() {},
		},
		newRunId: () => "failure-run",
	});

	const state = await controller.start(command("start"));
	assert.equal(abortCalls, 1);
	assert.equal(state.status, "failed");
	assert.equal(state.stage, "BLOCKED");
	assert.equal(state.children.find((item) => item.id === "alpha")?.status, "failed");
	assert.equal(state.children.find((item) => item.id === "beta")?.status, "failed");
});
