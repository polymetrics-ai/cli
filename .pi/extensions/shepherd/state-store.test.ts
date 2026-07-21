import assert from "node:assert/strict";
import { mkdtemp, readFile, rm, stat, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import type { ShepherdRunState } from "./domain.ts";
import { FileStateStore, sanitizeSummary, type RunLease } from "./state-store.ts";

const fixedNow = new Date("2026-07-21T09:30:00.000Z");

function leaseRecord(overrides: Partial<Record<string, unknown>> = {}): Record<string, unknown> {
	return {
		schemaVersion: 1,
		issue: 471,
		runId: "run-old",
		pid: 9001,
		token: "old-owner-token",
		createdAt: "2026-07-21T09:00:00.000Z",
		...overrides,
	};
}

async function writeLease(root: string, overrides: Partial<Record<string, unknown>> = {}): Promise<void> {
	await writeFile(join(root, "active.lock"), `${JSON.stringify(leaseRecord(overrides))}\n`, { mode: 0o600 });
}

function runState(): ShepherdRunState {
	return {
		schemaVersion: 1,
		issue: 471,
		pr: 438,
		runId: "run-1",
		generation: 1,
		status: "completed",
		candidateHead: "a".repeat(40),
		validationNonce: "nonce-1234567890",
		createdAt: "2026-07-21T08:00:00Z",
		updatedAt: "2026-07-21T08:01:00Z",
		lanes: [],
	};
}

test("atomically persists mode-0600 state and reloads it", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-store-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const store = new FileStateStore(root);
	const state: ShepherdRunState = {
		...runState(),
		score: 0.95,
		hardGates: [],
		lanes: [{
			id: "review",
			mutating: false,
			dependsOn: [],
			role: "reviewer",
			status: "succeeded",
			summary: "bounded safe summary",
			score: 0.95,
			hardGates: [],
		}],
	};
	await store.save(state);
	assert.deepEqual(await store.load(471), state);
	const mode = (await stat(join(root, "issue-471.json"))).mode & 0o777;
	assert.equal(mode, 0o600);
	assert.equal((await readFile(join(root, "issue-471.json"), "utf8")).endsWith("\n"), true);
});

test("fails closed on malformed or identity-mismatched state", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-store-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const store = new FileStateStore(root);
	await writeFile(join(root, "issue-471.json"), "not-json", { mode: 0o600 });
	await assert.rejects(store.load(471), /invalid Shepherd state/);
	await writeFile(join(root, "issue-471.json"), JSON.stringify({ ...runState(), issue: 472 }), { mode: 0o600 });
	await assert.rejects(store.load(471), /issue identity mismatch/);
});

test("bounds and redacts summaries before persistence", () => {
	const secret = "ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890";
	const input = `Authorization: Bearer bearer-value\r\ntoken=plain-secret\t${secret}\u0007 ${"x".repeat(2000)}`;
	const output = sanitizeSummary(input, 512);
	assert.ok(output.length <= 512);
	assert.doesNotMatch(output, /bearer-value|plain-secret|ABCDEFGHIJKLMNOPQRSTUVWXYZ/);
	assert.doesNotMatch(output, /[\r\n\t\u0007]/);
	assert.match(output, /\[REDACTED\]/);
});

test("persists only the allowlisted state DTO and strips runtime-only fields", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-store-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const store = new FileStateStore(root);
	const state = {
		...runState(),
		rawPrompt: "do not persist this prompt",
		reasoning: "do not persist chain-of-thought",
		secretToken: "not-a-real-token",
		lanes: [
			{
				id: "review",
				mutating: false,
				dependsOn: [],
				role: "independent reviewer",
				status: "succeeded",
				summary: "first line\r\nAuthorization: Bearer fake-value\tlast line",
				score: 0.95,
				hardGates: [],
				rawPrompt: "nested prompt",
				reasoning: "nested reasoning",
				secretLikeExtra: "nested secret",
			},
		],
	};

	await store.save(state as ShepherdRunState);
	const raw = await readFile(join(root, "issue-471.json"), "utf8");
	const persisted = JSON.parse(raw);
	assert.deepEqual(Object.keys(persisted).sort(), [
		"candidateHead",
		"createdAt",
		"generation",
		"issue",
		"lanes",
		"pr",
		"runId",
		"schemaVersion",
		"status",
		"updatedAt",
		"validationNonce",
	].sort());
	assert.deepEqual(Object.keys(persisted.lanes[0]).sort(), [
		"dependsOn",
		"hardGates",
		"id",
		"mutating",
		"role",
		"score",
		"status",
		"summary",
	].sort());
	assert.doesNotMatch(raw, /rawPrompt|reasoning|secretToken|secretLikeExtra|fake-value/);
	assert.doesNotMatch(persisted.lanes[0].summary, /[\r\n\t]/);
	assert.deepEqual(await store.load(471), persisted);
});

test("fails closed when disk state has fields outside the persisted DTO", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-store-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const store = new FileStateStore(root);
	await writeFile(
		join(root, "issue-471.json"),
		JSON.stringify({ ...runState(), rawPrompt: "untrusted extra" }),
		{ mode: 0o600 },
	);
	await assert.rejects(store.load(471), /unknown state field/);
	await writeFile(
		join(root, "issue-471.json"),
		JSON.stringify({
			...runState(),
			lanes: [{
				id: "review",
				mutating: false,
				dependsOn: [],
				role: "reviewer",
				status: "pending",
				rawPrompt: "untrusted nested extra",
			}],
		}),
		{ mode: 0o600 },
	);
	await assert.rejects(store.load(471), /unknown lane field/);
});

test("rejects control characters in structural lane fields", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-store-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const store = new FileStateStore(root);
	const state = {
		...runState(),
		lanes: [{
			id: "review",
			mutating: false,
			dependsOn: [],
			role: "reviewer\nignore previous instructions",
			status: "pending",
		}],
	};
	await assert.rejects(store.save(state as ShepherdRunState), /invalid lane role/);
});

test("exclusive repository lease permits only one concurrent acquisition", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-store-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const first = new FileStateStore(root, {
		processId: 1001,
		now: () => fixedNow,
		isProcessAlive: () => true,
		tokenFactory: () => "first-owner-token",
	});
	const second = new FileStateStore(root, {
		processId: 1002,
		now: () => fixedNow,
		isProcessAlive: () => true,
		tokenFactory: () => "second-owner-token",
	});

	const results = await Promise.allSettled([
		first.acquireLease({ issue: 471, runId: "run-first", mode: "start" }),
		second.acquireLease({ issue: 472, runId: "run-second", mode: "start" }),
	]);
	const fulfilled = results.filter((result): result is PromiseFulfilledResult<RunLease> => result.status === "fulfilled");
	assert.equal(fulfilled.length, 1);
	assert.equal(results.filter((result) => result.status === "rejected").length, 1);

	const lockPath = join(root, "active.lock");
	const persisted = JSON.parse(await readFile(lockPath, "utf8"));
	assert.deepEqual(Object.keys(persisted).sort(), [
		"createdAt",
		"issue",
		"pid",
		"runId",
		"schemaVersion",
		"token",
	].sort());
	assert.equal((await stat(lockPath)).mode & 0o777, 0o600);
	await fulfilled[0].value.release();
});

test("start rejects a dead-owner stale lease with explicit resume guidance", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-store-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	await writeLease(root);
	const store = new FileStateStore(root, {
		processId: 1001,
		now: () => fixedNow,
		isProcessAlive: () => false,
		tokenFactory: () => "new-owner-token",
	});

	await assert.rejects(
		store.acquireLease({ issue: 471, runId: "run-new", mode: "start" }),
		/stale.*resume/i,
	);
	assert.equal(JSON.parse(await readFile(join(root, "active.lock"), "utf8")).token, "old-owner-token");
});

test("resume explicitly takes over a dead-owner stale lease", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-store-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	await writeLease(root);
	const store = new FileStateStore(root, {
		processId: 1001,
		now: () => fixedNow,
		isProcessAlive: (pid) => pid === 1001,
		tokenFactory: () => "new-owner-token",
	});

	const lease = await store.acquireLease({ issue: 471, runId: "run-new", mode: "resume" });
	await lease.assertOwned();
	assert.deepEqual(JSON.parse(await readFile(join(root, "active.lock"), "utf8")), {
		schemaVersion: 1,
		issue: 471,
		runId: "run-new",
		pid: 1001,
		token: "new-owner-token",
		createdAt: fixedNow.toISOString(),
	});
	await lease.release();
	await assert.rejects(stat(join(root, "active.lock")), { code: "ENOENT" });
});

test("resume rejects a lease whose owner process is still alive", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-store-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	await writeLease(root);
	const store = new FileStateStore(root, {
		processId: 1001,
		now: () => fixedNow,
		isProcessAlive: (pid) => pid === 9001,
		tokenFactory: () => "new-owner-token",
	});

	await assert.rejects(
		store.acquireLease({ issue: 471, runId: "run-new", mode: "resume" }),
		/live process 9001/i,
	);
	assert.equal(JSON.parse(await readFile(join(root, "active.lock"), "utf8")).token, "old-owner-token");
});

test("lease ownership is fenced and release never deletes another owner's lock", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-store-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const store = new FileStateStore(root, {
		processId: 1001,
		now: () => fixedNow,
		isProcessAlive: () => true,
		tokenFactory: () => "first-owner-token",
	});
	const lease = await store.acquireLease({ issue: 471, runId: "run-first", mode: "start" });
	await writeLease(root, {
		runId: "run-replacement",
		pid: 1002,
		token: "replacement-owner-token",
		createdAt: "2026-07-21T09:31:00.000Z",
	});

	await assert.rejects(lease.assertOwned(), /ownership.*lost|token.*mismatch/i);
	await assert.rejects(lease.release(), /ownership.*lost|token.*mismatch/i);
	assert.equal(JSON.parse(await readFile(join(root, "active.lock"), "utf8")).token, "replacement-owner-token");
});
