import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import type { ProductionEffectKind } from "./autonomous-production-contract.ts";
import { ProductionEffectJournal, productionEffectKey, type ProductionEffectIntent } from "./autonomous-effect-journal.ts";
import { ProductionRecoveryBarrier } from "./autonomous-recovery.ts";

function intent(kind: ProductionEffectKind, marker: string): ProductionEffectIntent {
	const recoveryDescriptor = { kind, marker };
	const intentDigest = createHash("sha256").update(JSON.stringify(recoveryDescriptor)).digest("hex");
	const coordinates = { kind, runId: "run-1", generation: 1, childId: "state", intentDigest };
	return { key: productionEffectKey(coordinates), ...coordinates, recoveryDescriptor };
}

test("reconciles every uncertain effect in key order before opening the scheduling barrier", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-recovery-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const journal = new ProductionEffectJournal(root);
	const first = await journal.prepare(intent("git_push", "a".repeat(64)));
	const second = await journal.prepare(intent("child_pull_request", "b".repeat(64)));
	await journal.observe(second.key, { runId: "run-1", generation: 1 }, "d".repeat(64));
	const calls: string[] = [];
	const barrier = new ProductionRecoveryBarrier(journal, {
		async observe(record) {
			calls.push(`observe:${record.key}`);
			return { resultDigest: "c".repeat(64) };
		},
		async apply(record) { calls.push(`apply:${record.key}`); },
	});

	const result = await barrier.open({ runId: "run-1", generation: 1 });
	assert.equal(result.reconciled, 2);
	const expected = [first, (await journal.load(second.key))!].sort((left, right) => left.key.localeCompare(right.key)).flatMap((record) => [
		...(record.phase === "prepared" ? [`observe:${record.key}`] : []),
		`apply:${record.key}`,
	]);
	assert.deepEqual(calls, expected);
	assert.deepEqual(await journal.listNonApplied(), []);
	assert.deepEqual(await barrier.open({ runId: "run-1", generation: 1 }), { reconciled: 0 });
});

test("fails closed on stale generations before invoking recovery adapters", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-recovery-stale-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const journal = new ProductionEffectJournal(root);
	await journal.prepare(intent("git_push", "a".repeat(64)));
	let called = false;
	const barrier = new ProductionRecoveryBarrier(journal, {
		async observe() { called = true; return { resultDigest: "b".repeat(64) }; },
		async apply() { called = true; },
	});
	await assert.rejects(barrier.open({ runId: "run-2", generation: 2 }), /stale|generation|fence/i);
	assert.equal(called, false);
});

test("fails closed before external reconciliation when a legacy prepared effect has no recovery coordinates", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-recovery-no-descriptor-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const journal = new ProductionEffectJournal(root);
	const legacyCoordinates = {
		kind: "git_push" as const,
		runId: "run-1",
		generation: 1,
		childId: "state",
		intentDigest: "a".repeat(64),
	};
	await journal.prepare({ key: productionEffectKey(legacyCoordinates), ...legacyCoordinates });
	let called = false;
	const barrier = new ProductionRecoveryBarrier(journal, {
		async observe() { called = true; return { resultDigest: "b".repeat(64) }; },
		async apply() { called = true; },
	});
	await assert.rejects(barrier.open({ runId: "run-1", generation: 1 }), /recovery coordinates|descriptor/i);
	assert.equal(called, false);
});

test("cancellation closes the barrier and never applies an observation completed after abort", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-recovery-cancel-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const journal = new ProductionEffectJournal(root);
	await journal.prepare(intent("git_push", "a".repeat(64)));
	const controller = new AbortController();
	let applied = false;
	const barrier = new ProductionRecoveryBarrier(journal, {
		async observe() {
			controller.abort(new Error("operator stop"));
			return { resultDigest: "b".repeat(64) };
		},
		async apply() { applied = true; },
	});
	await assert.rejects(barrier.open({ runId: "run-1", generation: 1, signal: controller.signal }), /cancel|abort|stop/i);
	assert.equal(applied, false);
	assert.equal((await journal.listNonApplied())[0].phase, "prepared");
});

test("fresh processes reconcile every external-effect kind at prepared and observed crash checkpoints without duplicate replay", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-recovery-matrix-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const kinds: ProductionEffectKind[] = [
		"workspace_claim", "agent_implementation", "agent_correction", "shell_verification", "git_commit", "git_push",
		"child_pull_request", "independent_review", "child_integration", "parent_refresh", "human_request", "human_consume",
		"parent_merge_observation",
	];
	const writer = new ProductionEffectJournal(root);
	const preparedKeys = new Set<string>();
	for (const [index, kind] of kinds.entries()) {
		const record = await writer.prepare(intent(kind, `effect-${index}`));
		if (index % 2 === 0) preparedKeys.add(record.key);
		else await writer.observe(record.key, { runId: "run-1", generation: 1 }, createHash("sha256").update(`result-${index}`).digest("hex"));
	}

	const observed: string[] = [];
	const applied: string[] = [];
	const restarted = new ProductionEffectJournal(root);
	const barrier = new ProductionRecoveryBarrier(restarted, {
		async observe(record) {
			assert.deepEqual(record.recoveryDescriptor, {
				kind: record.kind,
				marker: `effect-${kinds.indexOf(record.kind)}`,
			});
			observed.push(record.key);
			return { resultDigest: createHash("sha256").update(`reconciled-${record.key}`).digest("hex") };
		},
		async apply(record) { applied.push(record.key); },
	});
	assert.deepEqual(await barrier.open({ runId: "run-1", generation: 1 }), { reconciled: kinds.length });
	assert.deepEqual(new Set(observed), preparedKeys);
	assert.equal(applied.length, kinds.length);
	assert.deepEqual(await barrier.open({ runId: "run-1", generation: 1 }), { reconciled: 0 });
	assert.equal(observed.length, preparedKeys.size);
	assert.equal(applied.length, kinds.length);
});
