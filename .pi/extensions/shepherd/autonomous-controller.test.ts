import assert from "node:assert/strict";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import { AutonomousShepherdController } from "./autonomous-controller.ts";
import { AutonomousFileStateStore, type AutonomousChildPlan } from "./autonomous-state.ts";

function deferred<T>() {
	let resolve!: (value: T) => void;
	const promise = new Promise<T>((accept) => { resolve = accept; });
	return { promise, resolve };
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
	assert.deepEqual(effects, [
		"verify:alpha", "review:alpha", "integrate:alpha",
		"verify:beta", "review:beta", "integrate:beta",
		"verify:gamma", "review:gamma", "integrate:gamma",
		"request_human_gate",
	]);
	assert.equal("mergeMain" in resumedController, false);
});
