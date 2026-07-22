import assert from "node:assert/strict";
import { mkdir, mkdtemp, readdir, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import {
	ProductionEffectJournal,
	productionEffectKey,
	type ProductionEffectIntent,
} from "./autonomous-effect-journal.ts";

function intent(overrides: Partial<ProductionEffectIntent> = {}): ProductionEffectIntent {
	const coordinates = {
		kind: "git_push" as const,
		runId: "run-479-1",
		generation: 1,
		childId: "state",
		intentDigest: "a".repeat(64),
	};
	return { key: productionEffectKey(coordinates), ...coordinates, ...overrides };
}

test("durably advances prepared -> observed -> applied and replays each exact phase idempotently", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-effect-journal-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const first = new ProductionEffectJournal(root);
	const prepared = await first.prepare(intent(), new Date("2026-07-22T10:00:00Z"));
	assert.equal(prepared.phase, "prepared");
	assert.deepEqual(await first.prepare(intent(), new Date("2026-07-22T10:01:00Z")), prepared);

	const restarted = new ProductionEffectJournal(root);
	const observed = await restarted.observe(prepared.key, {
		runId: prepared.runId,
		generation: prepared.generation,
	}, "b".repeat(64), new Date("2026-07-22T10:02:00Z"));
	assert.equal(observed.phase, "observed");
	assert.equal(observed.resultDigest, "b".repeat(64));
	assert.deepEqual(
		await restarted.observe(prepared.key, { runId: prepared.runId, generation: 1 }, "b".repeat(64)),
		observed,
	);

	const applied = await restarted.apply(prepared.key, {
		runId: prepared.runId,
		generation: prepared.generation,
	}, new Date("2026-07-22T10:03:00Z"));
	assert.equal(applied.phase, "applied");
	assert.deepEqual(await restarted.apply(prepared.key, { runId: prepared.runId, generation: 1 }), applied);
	assert.deepEqual(await restarted.listNonApplied(), []);
});

test("rejects key conflicts, changed observations, skipped phases, stale fences, and malformed digests", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-effect-conflicts-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const journal = new ProductionEffectJournal(root);
	const exact = intent();
	await journal.prepare(exact);
	await assert.rejects(journal.prepare({ ...exact, intentDigest: "c".repeat(64) }), /conflict|intent/i);
	await assert.rejects(journal.apply(exact.key, { runId: exact.runId, generation: 1 }), /observed|phase/i);
	await assert.rejects(journal.observe(exact.key, { runId: exact.runId, generation: 2 }, "b".repeat(64)), /stale|fence|generation/i);
	await journal.observe(exact.key, { runId: exact.runId, generation: 1 }, "b".repeat(64));
	await assert.rejects(
		journal.observe(exact.key, { runId: exact.runId, generation: 1 }, "d".repeat(64)),
		/conflict|result/i,
	);
	await assert.rejects(journal.prepare({ ...intent(), key: "e".repeat(64), intentDigest: "not-a-digest" }), /digest/i);
});

test("serializes competing phase transitions across repository instances", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-effect-concurrency-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const one = new ProductionEffectJournal(root);
	const two = new ProductionEffectJournal(root);
	const exact = intent();
	await one.prepare(exact);
	const outcomes = await Promise.allSettled([
		one.observe(exact.key, { runId: exact.runId, generation: 1 }, "b".repeat(64)),
		two.observe(exact.key, { runId: exact.runId, generation: 1 }, "c".repeat(64)),
	]);
	assert.equal(outcomes.filter((result) => result.status === "fulfilled").length, 1);
	assert.equal(outcomes.filter((result) => result.status === "rejected").length, 1);
});

test("restart reclaims a complete orphan journal lock owned by a dead process", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-effect-orphan-lock-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const lock = join(root, ".production-effects.lock");
	await mkdir(lock);
	await writeFile(join(lock, "owner.json"), JSON.stringify({
		schemaVersion: 1,
		pid: 99_999_999,
		token: "00000000-0000-4000-8000-000000000002",
	}));
	const journal = new ProductionEffectJournal(root);
	assert.equal((await journal.prepare(intent())).phase, "prepared");
	assert.deepEqual(await readdir(root), ["production-effects.json"]);
});
